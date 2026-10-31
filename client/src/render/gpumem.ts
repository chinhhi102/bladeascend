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
}

const alloc = new Map<WebGLTexture, Alloc>();
let totalBytes = 0;
let peakBytes = 0;

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
  alloc.set(tex, { bytes, w, h });
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

/** gpuBiggest liệt kê vài texture to nhất — để biết cắt chỗ nào trước. */
export function gpuBiggest(n = 8): { w: number; h: number; mb: number }[] {
  return [...alloc.values()]
    .sort((a, b) => b.bytes - a.bytes)
    .slice(0, n)
    .map((a) => ({ w: a.w, h: a.h, mb: +(a.bytes / 1024 / 1024).toFixed(2) }));
}
