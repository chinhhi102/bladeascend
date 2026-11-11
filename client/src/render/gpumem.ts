/**
 * Ước tính bộ nhớ texture trên GPU (NETCODE_SPEC §5 L1, TELEMETRY_SPEC §2.2 cột `tex_mb`).
 *
 * Vì sao phải tự đo: **không trình duyệt nào cho biết một trang đang giữ bao nhiêu MB
 * texture.** Cột `tex_mb` trong telemetry vì thế luôn bằng 0 từ ngày mở, và con số
 * "1 076 → 102 MB" trong PERF_LOG lượt #6 là đo tay một lần rồi thôi — không ai theo dõi
 * được nó trên máy người chơi.
 *
 * Mà đó lại đúng là con số cần nhất lúc này: iPhone có ngân sách bộ nhớ GPU rất chặt, và
 * khi vượt thì Safari **giết hẳn tab** chứ không báo lỗi gì. Không đo thì chỉ còn cách đoán.
 *
 * Cách đo: bọc đúng bốn hàm WebGL cấp phát/giải phóng texture và cộng trừ theo texture đang
 * được gắn. Con số là **ước tính**: giả định RGBA8 (4 byte/điểm ảnh) và mipmap +33%. Nó
 * không chính xác tới từng byte, nhưng đúng về ĐỘ LỚN và đúng về XU HƯỚNG — đủ để biết
 * 80 MB hay 800 MB, và đủ để hạ chất lượng trước khi tab bị giết.
 */

interface Alloc {
  bytes: number;
  w: number;
  h: number;
  site?: string;
}

const alloc = new Map<WebGLTexture, Alloc>();
let totalBytes = 0;
let peakBytes = 0;

/**
 * TRACE: ghi lại NƠI cấp phát từng texture (`?textrace=1`).
 *
 * Vì sao cần: telemetry người thật cho thấy số texture leo từ 150 lên hơn 5 000 trong một
 * phiên 25 phút. Biết "có rò rỉ" là chưa đủ để sửa — phải biết dòng lệnh nào sinh ra chúng.
 * Tắt mặc định vì `new Error().stack` trong đường cấp phát texture là rất đắt.
 */
const TRACE = typeof location !== 'undefined' && location.search.includes('textrace');

function siteOf(): string | undefined {
  if (!TRACE) {
    return undefined;
  }
  return curSource || '(outside onSourceUpdate)';
}

/**
 * trackSourceNames gắn TÊN NGUỒN ẢNH vào mỗi texture — thứ duy nhất chỉ đúng thủ phạm.
 *
 * Vết ngăn xếp ở `texImage2D` chỉ thấy ruột PixiJS (`uploadImageWebGL2`…): lúc tải lên thì
 * đang ở giữa vòng vẽ, không còn dấu gì của chỗ tạo. Nhưng Pixi tải lên THEO `TextureSource`
 * và mỗi nguồn có `label` (đường dẫn tệp, `text:…`, `renderTexture`…). Bọc đúng một hàm
 * `GlTextureSystem.onSourceUpdate` là biết texture vừa tải lên thuộc nguồn nào.
 */
let curSource = '';

/** fromStacks: nguồn ảnh dựng bằng `Texture.from(canvas)` → ngăn xếp của chỗ gọi. */
const fromStacks = new WeakMap<object, string>();

/** trackTextureFrom bọc `Texture.from` để biết MÃ NÀO của game dựng ra canvas đó. */
export function trackTextureFrom(TextureClass: object): void {
  if (!TRACE) {
    return;
  }
  const holder = TextureClass as { from: (...a: unknown[]) => unknown };
  const orig = holder.from.bind(TextureClass);
  holder.from = (...a: unknown[]): unknown => {
    const tex = orig(...a) as { source?: object } | undefined;
    const src = tex?.source;
    if (src && !fromStacks.has(src)) {
      const lines = (new Error().stack ?? '').split('\n').slice(2, 6);
      fromStacks.set(
        src,
        lines
          .map((l) => l.trim().replace(/^at /, '').replace(/\?[^ )]*/g, '').split('/').slice(-1)[0])
          .join(' ← '),
      );
    }
    return tex;
  };
}

export function trackSourceNames(texSystem: Record<string, unknown>): void {
  if (!TRACE || typeof texSystem.onSourceUpdate !== 'function') {
    return;
  }
  const orig = (texSystem.onSourceUpdate as (s: unknown) => unknown).bind(texSystem);
  texSystem.onSourceUpdate = (src: unknown): unknown => {
    const s = src as
      | { label?: string; uid?: number; resource?: { src?: string; nodeName?: string } }
      | null;
    const kind = (s as { constructor?: { name?: string } } | null)?.constructor?.name ?? '?';
    const res = s?.resource;
    const resKind = res ? ((res as { nodeName?: string }).nodeName ?? res.constructor?.name ?? '?') : 'none';
    curSource =
      (s ? fromStacks.get(s) : undefined) || s?.label || res?.src || `${kind}/${resKind}`;
    try {
      return orig(src);
    } finally {
      curSource = '';
    }
  };
}

/** BYTES_PER_PIXEL: giả định RGBA8. Phần lớn texture của game là PNG/WebP có alpha. */
const BYTES_PER_PIXEL = 4;
/** MIP_FACTOR: có mipmap thì tốn thêm ~1/3. Pixi bật mipmap cho phần lớn texture. */
const MIP_FACTOR = 4 / 3;

function note(gl: WebGLRenderingContext | WebGL2RenderingContext, w: number, h: number, mip: boolean): void {
  const tex = gl.getParameter(gl.TEXTURE_BINDING_2D) as WebGLTexture | null;
  if (!tex || !(w > 0) || !(h > 0)) {
    return;
  }
  const bytes = Math.round(w * h * BYTES_PER_PIXEL * (mip ? MIP_FACTOR : 1));
  const old = alloc.get(tex);
  if (old) {
    totalBytes -= old.bytes;
  }
  alloc.set(tex, { bytes, w, h, site: old?.site ?? siteOf() });
  totalBytes += bytes;
  peakBytes = Math.max(peakBytes, totalBytes);
}

/**
 * trackGpuMemory bọc ngữ cảnh WebGL. Gọi MỘT lần ngay sau khi renderer dựng xong.
 *
 * Bọc cả `texImage2D` lẫn `texStorage2D` vì Pixi dùng cả hai tuỳ WebGL1/2 và tuỳ định dạng.
 * Thiếu một cái là con số hụt mà không có dấu hiệu nào.
 */
export function trackGpuMemory(gl: WebGLRenderingContext | WebGL2RenderingContext): void {
  const g = gl as unknown as Record<string, (...a: unknown[]) => unknown>;

  const texImage2D = g.texImage2D.bind(gl);
  g.texImage2D = (...a: unknown[]): unknown => {
    // Hai dạng chữ ký: (target, level, internalformat, width, height, border, format, type,
    // pixels) — 9 tham số; và (target, level, internalformat, format, type, source) — 6
    // tham số, kích thước lấy từ chính nguồn ảnh.
    if (a.length >= 9) {
      note(gl, Number(a[3]), Number(a[4]), Number(a[1]) === 0);
    } else if (a.length >= 6) {
      const src = a[5] as { width?: number; height?: number; videoWidth?: number; videoHeight?: number } | null;
      note(gl, src?.width ?? src?.videoWidth ?? 0, src?.height ?? src?.videoHeight ?? 0, Number(a[1]) === 0);
    }
    return texImage2D(...a);
  };

  const texStorage2D = g.texStorage2D?.bind(gl);
  if (texStorage2D) {
    g.texStorage2D = (...a: unknown[]): unknown => {
      note(gl, Number(a[3]), Number(a[4]), Number(a[1]) > 1);
      return texStorage2D(...a);
    };
  }

  const compressed = g.compressedTexImage2D?.bind(gl);
  if (compressed) {
    g.compressedTexImage2D = (...a: unknown[]): unknown => {
      note(gl, Number(a[3]), Number(a[4]), Number(a[1]) === 0);
      return compressed(...a);
    };
  }

  const deleteTexture = g.deleteTexture.bind(gl);
  g.deleteTexture = (...a: unknown[]): unknown => {
    const tex = a[0] as WebGLTexture | null;
    const old = tex ? alloc.get(tex) : undefined;
    if (tex && old) {
      totalBytes -= old.bytes;
      alloc.delete(tex);
    }
    return deleteTexture(...a);
  };
}

/** gpuMemMB là ước tính bộ nhớ texture đang giữ, tính bằng MB. */
export function gpuMemMB(): number {
  return totalBytes / 1024 / 1024;
}

/** gpuPeakMB là mức cao nhất từ lúc mở trang — thứ quyết định tab có bị giết hay không. */
export function gpuPeakMB(): number {
  return peakBytes / 1024 / 1024;
}

/** gpuTextureCount là số texture đang giữ trên GPU. */
export function gpuTextureCount(): number {
  return alloc.size;
}

/**
 * gpuSites gom texture ĐANG GIỮ theo nơi cấp phát — chỉ có số khi mở `?textrace=1`.
 *
 * Đây là thứ trả lời "rò rỉ từ đâu": một nhóm có số lượng tăng đều theo thời gian trong khi
 * các nhóm khác đứng yên thì chính nó là chỗ quên nhả.
 */
export function gpuSites(n = 12): { site: string; count: number; mb: number }[] {
  const by = new Map<string, { count: number; bytes: number }>();
  for (const a of alloc.values()) {
    // Gộp tên: `uid:123` và tệp có hash trong tên là MỖI CÁI MỘT DÒNG nếu để nguyên, và
    // một bảng 500 dòng count=1 thì không thấy được nhóm nào đang phình.
    const raw = a.site ?? '(no ?textrace=1)';
    const key = `${a.w}x${a.h} ${raw.replace(/\d+/g, '#').replace(/^.*\/assets\//, 'assets/')}`;
    const cur = by.get(key) ?? { count: 0, bytes: 0 };
    cur.count++;
    cur.bytes += a.bytes;
    by.set(key, cur);
  }
  return [...by.entries()]
    .sort((x, y) => y[1].count - x[1].count)
    .slice(0, n)
    .map(([site, v]) => ({ site, count: v.count, mb: +(v.bytes / 1024 / 1024).toFixed(2) }));
}

/** gpuBiggest liệt kê vài texture to nhất — để biết cắt chỗ nào trước. */
export function gpuBiggest(n = 8): { w: number; h: number; mb: number }[] {
  return [...alloc.values()]
    .sort((a, b) => b.bytes - a.bytes)
    .slice(0, n)
    .map((a) => ({ w: a.w, h: a.h, mb: +(a.bytes / 1024 / 1024).toFixed(2) }));
}
