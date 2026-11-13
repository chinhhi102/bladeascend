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
  /** Hạng mục ngân sách (§13.10). Xem `budgetCategory`. */
  cat: BudgetCat;
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
 * ── SỔ VRAM THEO HẠNG MỤC NGÂN SÁCH — GAME_SPEC §13.10 ─────────────────────────────────
 *
 * Spec đòi "mọi texture mới phải khai hạng mục ngân sách khi tạo". Khai bằng cách sửa từng
 * chỗ `new Texture(...)` thì sẽ có chỗ quên, và chỗ quên im lặng không tính vào đâu cả —
 * đúng kiểu sổ sách nói ít hơn sự thật. Nên ở đây xếp loại từ NHÃN NGUỒN mà Pixi đã có sẵn
 * (đường dẫn tệp, `text:…`, `renderTexture`), cộng một thùng `khác` mà sổ KÊU LÊN khi có
 * gì rơi vào. Thêm một loại texture mới mà quên khai thì nó hiện ngay ở dòng `khác`, không
 * biến mất.
 */
export type BudgetCat = 'chars' | 'ground' | 'world' | 'fx' | 'dynamic' | 'framebuffer' | 'other';

/** Nhãn tiếng Việt cho bảng F3/admin. */
export const BUDGET_LABELS: Record<BudgetCat, string> = {
  chars: 'Atlas nhân vật', // i18n-ok: lớp số liệu debug
  ground: 'Nền', // i18n-ok: lớp số liệu debug
  world: 'Vật cản · decal · vật phẩm', // i18n-ok: lớp số liệu debug
  fx: 'VFX · hào quang · UI', // i18n-ok: lớp số liệu debug
  dynamic: 'Chữ/số bay (atlas động)', // i18n-ok: lớp số liệu debug
  framebuffer: 'Bộ đệm khung · bộ lọc', // i18n-ok: lớp số liệu debug
  other: 'KHÁC (chưa khai hạng mục)', // i18n-ok: lớp số liệu debug
};

/**
 * Trần ngân sách MB theo §13.10. `phone` là cột điện thoại.
 *
 * `other` có trần 0: mọi byte rơi vào đây là một texture chưa ai khai hạng mục, và spec đòi
 * sổ phải báo vượt. Trần 0 làm nó đỏ ngay byte đầu tiên.
 */
export const BUDGET_MB: Record<BudgetCat, { pc: number; phone: number }> = {
  chars: { pc: 24, phone: 6 },
  ground: { pc: 4, phone: 2 },
  world: { pc: 6, phone: 3 },
  fx: { pc: 4, phone: 2 },
  dynamic: { pc: 2, phone: 2 },
  framebuffer: { pc: 12, phone: 8 },
  other: { pc: 0, phone: 0 },
};

/**
 * declared: hạng mục do MÃ GAME khai tường minh, theo `TextureSource`.
 *
 * Phân loại bằng nhãn nguồn chỉ đúng với texture tải từ tệp — chúng mang đường dẫn. Texture
 * dựng tại chỗ (canvas của mặt nạ nền, bóng đổ, lưỡi kiếm dự phòng, render texture) thì
 * nhãn của Pixi là tên lớp, mà **bản build thật đã minify tên lớp**: đo lần đầu ra
 * `fe/none`, `it/CANVAS`, `b/CANVAS` — 231 trên 290 MB rơi vào "chưa khai hạng mục" và sổ
 * không nói được gì. Nên mọi chỗ dựng texture tại chỗ phải gọi `declareTexture`.
 */
const declared = new WeakMap<object, BudgetCat>();

/**
 * declareTexture khai hạng mục ngân sách cho một texture (§13.10).
 *
 * Gọi NGAY sau khi tạo, trước khi nó được vẽ lần đầu. Nhận cả `Texture` lẫn `TextureSource`
 * cho gọn chỗ gọi.
 */
export function declareTexture(tex: unknown, cat: BudgetCat): void {
  const t = tex as { source?: object } | null;
  const src = (t?.source ?? t) as (object & { label?: string }) | null;
  if (!src) {
    return;
  }
  declared.set(src, cat);
  // Ghi LUÔN vào `label` chứ không chỉ WeakMap.
  //
  // WeakMap khoá theo đúng một object; Pixi có những đường dựng lại `TextureSource` (đổi
  // kích thước, lấy từ pool) và khi đó khoá cũ không còn khớp — texture lặng lẽ rơi về
  // "chưa khai hạng mục". Nhãn thì đi theo nguồn qua mọi đường đó.
  if (!src.label || !src.label.startsWith('kx:')) {
    src.label = `kx:${cat}${src.label ? ' ' + src.label : ''}`;
  }
}

/** declaredCat: hạng mục đã khai cho nguồn đang tải lên, nếu có. */
function declaredCat(): BudgetCat | undefined {
  return curSourceObj ? declared.get(curSourceObj) : undefined;
}

/**
 * budgetCategory xếp một texture vào hạng mục ngân sách.
 *
 * Thứ tự: khai tường minh → đường dẫn tệp → cỡ texture. Rơi hết thì vào `other`, và `other`
 * có trần 0 nên sổ kêu ngay.
 */
export function budgetCategory(site: string | undefined, w: number, h: number): BudgetCat {
  const explicit = declaredCat();
  if (explicit) {
    return explicit;
  }
  const s = (site ?? '').toLowerCase();
  if (s.startsWith('kx:')) {
    return s.slice(3).split(/[ /]/)[0] as BudgetCat;
  }
  // Nguồn KHÔNG có tài nguyên phía CPU (`.../none`) là bộ đệm chỉ sống trên GPU: bộ đệm
  // khung, đích vẽ của bộ lọc, pool canvas của Pixi. Đó đúng là hạng mục "bộ đệm khung".
  //
  // Nhận ra bằng chữ `none` chứ không bằng tên lớp: bản build thật minify tên lớp thành
  // `fe`, `it`, `b` — lượt đo đầu 42 MB bộ đệm 1920×1080 rơi vào "chưa khai hạng mục" vì
  // tôi cố đoán theo tên lớp.
  if (s.endsWith('/none') || s.includes('rendertarget')) {
    return 'framebuffer';
  }
  // `CanvasPool` của Pixi KHÔNG phải bộ đệm khung — đó là canvas VẼ CHỮ.
  //
  // Tôi xếp nó vào "bộ đệm khung" suốt bốn lượt tối ưu vì nó không có tệp nguồn, và nó biến
  // 8 MB canvas chữ thành một dòng không ai soi. Ngày 13/11 móc ngăn xếp vào
  // `document.createElement` mới ra mặt: 6 tấm 512×512 đều do `CanvasPool` cấp cho `Text`.
  // Pixi làm tròn cỡ canvas chữ lên luỹ thừa 2, nên một dòng tên 300×40 ăn trọn 512×512.
  //
  // Nhận cả hai đường: có `?textrace=1` thì ngăn xếp ghi `CanvasPool`; không có thì nhãn chỉ
  // còn `<lớp đã minify>/CANVAS` — và trong game này, texture canvas nào không khai hạng mục
  // cũng là chữ.
  if (s.includes('canvaspool') || s.endsWith('/canvas')) {
    return 'dynamic';
  }
  // Tên thư mục lấy từ `assets/src` THẬT (characters, tiles, obstacles, decals, decor,
  // items, monsters, vfx, ui, swords, beasts, buffs, gates, ambient, zones, brand, fonts).
  // Lượt đo đầu tôi gõ `/chars/` và cột "Atlas nhân vật" ra 0,00 MB trong khi atlas nhân
  // vật nằm chình ình ở dòng "chưa khai hạng mục".
  if (s.includes('/characters/') || s.includes('characters/')) {
    return 'chars';
  }
  if (s.includes('/tiles/') || s.includes('tiles/') || s.includes('/zones/') || s.includes('zones/')) {
    return 'ground';
  }
  if (s.includes('/obstacles/') || s.includes('obstacles/') || s.includes('/decals/') || s.includes('decals/')
    || s.includes('/decor/') || s.includes('decor/') || s.includes('/items/') || s.includes('items/')
    || s.includes('/monsters/') || s.includes('monsters/') || s.includes('/gates/') || s.includes('gates/')) {
    return 'world';
  }
  if (s.includes('/vfx/') || s.includes('vfx/') || s.includes('/ui/') || s.includes('ui/')
    || s.includes('/beasts/') || s.includes('beasts/') || s.includes('/swords/') || s.includes('swords/')
    || s.includes('/buffs/') || s.includes('buffs/') || s.includes('/ambient/') || s.includes('ambient/')
    || s.includes('/brand/') || s.includes('brand/')) {
    return 'fx';
  }
  if (s.startsWith('text:') || s.includes('/fonts/') || s.includes('fonts/') || s.includes('bitmapfont')) {
    return 'dynamic';
  }
  // Texture 1×1/2×2 là của Pixi (trắng, trống) — không đáng một dòng cảnh báo.
  if (w <= 2 && h <= 2) {
    return 'fx';
  }
  return 'other';
}

/**
 * trackSourceNames gắn TÊN NGUỒN ẢNH vào mỗi texture — thứ duy nhất chỉ đúng thủ phạm.
 *
 * Vết ngăn xếp ở `texImage2D` chỉ thấy ruột PixiJS (`uploadImageWebGL2`…): lúc tải lên thì
 * đang ở giữa vòng vẽ, không còn dấu gì của chỗ tạo. Nhưng Pixi tải lên THEO `TextureSource`
 * và mỗi nguồn có `label` (đường dẫn tệp, `text:…`, `renderTexture`…).
 */
let curSource = '';
/** curSourceObj: chính `TextureSource` đang tải lên — để tra hạng mục đã khai. */
let curSourceObj: object | null = null;

/** fromStacks: nguồn ảnh dựng bằng `Texture.from(canvas)` → ngăn xếp của chỗ gọi. */
const fromStacks = new WeakMap<object, string>();

/** trackTextureFrom bọc `Texture.from` để biết MÃ NÀO của game dựng ra canvas đó. */
let fromWrapped = false;

export function trackTextureFrom(TextureClass: object): void {
  if (!TRACE || fromWrapped) {
    return;
  }
  trackCanvasCreation();
  // Bọc đúng MỘT LẦN: `buildApp` chạy lại khi dựng lại renderer (§13.8), và bọc chồng lên
  // nhau thì mỗi lần tạo texture đi qua n lớp wrapper — chậm dần và vết gọi sai dần.
  fromWrapped = true;
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

/**
 * trackCanvasCreation ghi ngăn xếp lên CHÍNH thẻ `<canvas>` lúc nó ra đời.
 *
 * `trackTextureFrom` chỉ bắt được canvas đi qua `Texture.from`. Thứ KHÔNG đi qua đó —
 * `new CanvasSource({ resource: cv })` — về tới sổ VRAM chỉ còn cái nhãn `CanvasSource/CANVAS`
 * (bị rút gọn thành `b/CANVAS`), và 8 MB canvas 512² nằm trong hạng mục KHÁC suốt bốn lượt
 * tối ưu mà không ai chỉ được chỗ tạo. Móc vào `document.createElement` thì mọi đường đều bị
 * bắt, vì canvas nào cũng phải qua đây.
 *
 * Chỉ chạy sau `?textrace=1`: `new Error().stack` đắt, mà game tạo canvas cả trong vòng vẽ.
 */
function trackCanvasCreation(): void {
  const doc = globalThis.document as
    | (Document & { __kxCanvasTraced?: boolean })
    | undefined;
  if (!doc || doc.__kxCanvasTraced) {
    return;
  }
  doc.__kxCanvasTraced = true;
  const orig = doc.createElement.bind(doc);
  doc.createElement = ((tag: string, ...rest: unknown[]): HTMLElement => {
    const el = orig(tag as 'canvas', ...(rest as []));
    if (tag === 'canvas') {
      (el as HTMLElement & { __kxStack?: string }).__kxStack = shortStack(3, 7);
    }
    return el;
  }) as typeof doc.createElement;
}

/** shortStack: vài khung gọi đầu, bỏ phần đường dẫn, đủ để chỉ ra tệp:dòng. */
function shortStack(from: number, to: number): string {
  return (new Error().stack ?? '')
    .split('\n')
    .slice(from, to)
    .map((l) => l.trim().replace(/^at /, '').replace(/\?[^ )]*/g, '').split('/').slice(-1)[0])
    .join(' ← ');
}

export function trackSourceNames(texSystem: Record<string, unknown>): void {
  // Bọc CẢ `initSource`, không chỉ `onSourceUpdate`.
  //
  // Texture nén khối (KTX2) đi lên GPU qua `initSource` — `onSourceUpdate` không chạy cho
  // chúng. Thiếu chỗ này thì mọi atlas nén rơi vào hạng mục "chưa khai", và bảng ngân sách
  // §13.10 mất đúng hai dòng quan trọng nhất (atlas nhân vật, nền). Đã sập một lần: sau khi
  // bật KTX2, hai dòng đó về 0,00 MB còn dòng KHÁC phình lên.
  wrapSourceFn(texSystem, 'initSource');
  wrapSourceFn(texSystem, 'onSourceUpdate');
}

/** wrapSourceFn bọc một hàm nhận `TextureSource` để ghi nhãn nguồn đang tải lên. */
function wrapSourceFn(texSystem: Record<string, unknown>, name: string): void {
  // KHÔNG khoá sau `?textrace=1`: nhãn nguồn là thứ sổ VRAM đọc để xếp loại nên phải luôn
  // có. Phần ĐẮT là `new Error().stack` trong `trackTextureFrom`, và nó vẫn nằm sau cờ.
  if (typeof texSystem[name] !== 'function') {
    return;
  }
  const orig = (texSystem[name] as (s: unknown) => unknown).bind(texSystem);
  texSystem[name] = (src: unknown): unknown => {
    const s = src as
      | { label?: string; uid?: number; resource?: { src?: string; nodeName?: string } }
      | null;
    const kind = (s as { constructor?: { name?: string } } | null)?.constructor?.name ?? '?';
    const res = s?.resource;
    const resKind = res ? ((res as { nodeName?: string }).nodeName ?? res.constructor?.name ?? '?') : 'none';
    const canvasStack = (res as { __kxStack?: string } | undefined)?.__kxStack;
    curSource =
      (s ? fromStacks.get(s) : undefined) || s?.label || res?.src || canvasStack || `${kind}/${resKind}`;
    curSourceObj = (s as object | null) ?? null;
    try {
      return orig(src);
    } finally {
      curSource = '';
      curSourceObj = null;
    }
  };
}

/** BYTES_PER_PIXEL: giả định RGBA8. Phần lớn texture của game là PNG/WebP có alpha. */
const BYTES_PER_PIXEL = 4;
/**
 * Mã định dạng nén khối hay gặp (BC1–BC7, ETC2, ASTC 4×4…).
 *
 * Dùng để biết `texStorage2D` đang cấp phát một texture NÉN — lúc đó cỡ thật không phải
 * `w·h·4`, và con số thật sẽ do `compressedTexSubImage2D` ghi.
 */
const COMPRESSED_FORMATS = new Set([
  0x83f0, 0x83f1, 0x83f2, 0x83f3, // S3TC / BC1-BC3
  0x8e8c, 0x8e8d, 0x8e8e, 0x8e8f, // BPTC / BC7
  0x9274, 0x9275, 0x9276, 0x9277, 0x9278, 0x9279, // ETC2
  0x93b0, 0x93b1, 0x93b2, 0x93b3, 0x93b4, 0x93b5, 0x93b6, 0x93b7, // ASTC
  0x93b8, 0x93b9, 0x93ba, 0x93bb, 0x93bc, 0x93bd,
]);
/** MIP_FACTOR: có mipmap thì tốn thêm ~1/3. Pixi bật mipmap cho phần lớn texture. */
const MIP_FACTOR = 4 / 3;

function note(gl: WebGLRenderingContext | WebGL2RenderingContext, w: number, h: number, mip: boolean, exactBytes?: number): void {
  const tex = gl.getParameter(gl.TEXTURE_BINDING_2D) as WebGLTexture | null;
  if (!tex || !(w > 0) || !(h > 0)) {
    return;
  }
  const bytes = exactBytes ?? Math.round(w * h * BYTES_PER_PIXEL * (mip ? MIP_FACTOR : 1));
  const old = alloc.get(tex);
  if (old) {
    totalBytes -= old.bytes;
  }
  const site = old?.site ?? curSource ?? siteOf();
  alloc.set(tex, { bytes, w, h, site, cat: budgetCategory(site, w, h) });
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
      // Bỏ qua khi định dạng là NÉN: cỡ thật sẽ do `compressedTexSubImage2D` ghi. Tính
      // `w·h·4` ở đây là báo gấp 4–8 lần sự thật cho đúng thứ vừa được nén lại.
      if (!COMPRESSED_FORMATS.has(Number(a[2]))) {
        note(gl, Number(a[3]), Number(a[4]), Number(a[1]) > 1);
      }
      return texStorage2D(...a);
    };
  }

  const compressed = g.compressedTexImage2D?.bind(gl);
  if (compressed) {
    g.compressedTexImage2D = (...a: unknown[]): unknown => {
      // Texture NÉN KHỐI: kích thước thật là số byte đã truyền vào, KHÔNG phải `w·h·4`.
      //
      // Đây là cả điểm của §13.10: BC7 tốn 1 byte/điểm ảnh, BC1 tốn 0,5 — tính theo RGBA là
      // báo gấp 4–8 lần sự thật, và khi đó bảng ngân sách sẽ nói lượt tối ưu KTX2 chẳng được
      // gì. Đúng cái bẫy đã sập một lần: đo xong thấy số không đổi, suýt kết luận sai.
      //
      // Chỉ ghi mức 0 rồi nhân 4/3 cho cả chuỗi mipmap: mỗi mức là một lời gọi riêng, mà
      // `note` khoá theo texture nên mức sau sẽ ĐÈ mức trước — kết quả là chỉ còn mức nhỏ
      // nhất, tức gần bằng 0.
      if (Number(a[1]) === 0) {
        const data = a[6] as { byteLength?: number } | number | undefined;
        const n = typeof data === 'number' ? data : data?.byteLength;
        note(gl, Number(a[3]), Number(a[4]), true, n ? Math.round(n * MIP_FACTOR) : undefined);
      }
      return compressed(...a);
    };
  }

  // WebGL2 thường cấp phát bằng `texStorage2D` rồi nạp từng mức bằng `compressedTexSubImage2D`
  // — `compressedTexImage2D` không được gọi lần nào. Thiếu chỗ này thì toàn bộ texture nén
  // biến mất khỏi sổ: hai dòng "Atlas nhân vật" và "Nền" về 0,00 MB với 0 texture, trong khi
  // chúng đang chiếm chỗ thật trên GPU. Một cái sổ báo THIẾU còn nguy hơn báo thừa.
  const compressedSub = g.compressedTexSubImage2D?.bind(gl);
  if (compressedSub) {
    g.compressedTexSubImage2D = (...a: unknown[]): unknown => {
      if (Number(a[1]) === 0) {
        const data = a[7] as { byteLength?: number } | number | undefined;
        const n = typeof data === 'number' ? data : data?.byteLength;
        note(gl, Number(a[4]), Number(a[5]), true, n ? Math.round(n * MIP_FACTOR) : undefined);
      }
      return compressedSub(...a);
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

/**
 * gpuBudget: sổ VRAM theo hạng mục — bảng chính của §13.10.
 *
 * `phone` quyết định đọc cột trần nào. Trả cả `over` để F3 tô đỏ mà không phải tính lại.
 */
export function gpuBudget(phone: boolean): { cat: BudgetCat; label: string; mb: number; count: number; limitMB: number; over: boolean }[] {
  const by = new Map<BudgetCat, { bytes: number; count: number }>();
  for (const a of alloc.values()) {
    const cur = by.get(a.cat) ?? { bytes: 0, count: 0 };
    cur.bytes += a.bytes;
    cur.count++;
    by.set(a.cat, cur);
  }
  const cats: BudgetCat[] = ['chars', 'ground', 'world', 'fx', 'dynamic', 'framebuffer', 'other'];
  return cats.map((cat) => {
    const v = by.get(cat) ?? { bytes: 0, count: 0 };
    const mb = +(v.bytes / 1024 / 1024).toFixed(2);
    const limitMB = phone ? BUDGET_MB[cat].phone : BUDGET_MB[cat].pc;
    return { cat, label: BUDGET_LABELS[cat], mb, count: v.count, limitMB, over: mb > limitMB };
  });
}

/**
 * gpuUnclassified: vài nhãn nguồn đang rơi vào hạng mục `other`.
 *
 * Để câu cảnh báo nói được PHẢI SỬA CHỖ NÀO, chứ không chỉ "có 3 MB chưa khai".
 */
export function gpuUnclassified(n = 6): { site: string; mb: number }[] {
  const by = new Map<string, number>();
  for (const a of alloc.values()) {
    if (a.cat !== 'other') {
      continue;
    }
    const key = `${a.w}x${a.h} ${(a.site ?? '?').replace(/\d+/g, '#')}`;
    by.set(key, (by.get(key) ?? 0) + a.bytes);
  }
  return [...by.entries()]
    .sort((x, y) => y[1] - x[1])
    .slice(0, n)
    .map(([site, bytes]) => ({ site, mb: +(bytes / 1024 / 1024).toFixed(2) }));
}

/**
 * gpuListCat liệt kê texture của MỘT hạng mục, gom theo cỡ.
 *
 * Cần cho mục 6 của §13.10: "Bộ đệm khung · bộ lọc" đo ra 68,9 MB trong khi một khung 1080p
 * chỉ 8,3 MB — tức là có khoảng tám đích vẽ toàn màn hình đang sống cùng lúc, và câu hỏi
 * duy nhất đáng hỏi là CÁI NÀO.
 */
export function gpuListCat(cat: BudgetCat, n = 10): { size: string; count: number; mb: number; site: string }[] {
  const by = new Map<string, { count: number; bytes: number; site: string }>();
  for (const a of alloc.values()) {
    if (a.cat !== cat) {
      continue;
    }
    const key = `${a.w}x${a.h}`;
    const cur = by.get(key) ?? { count: 0, bytes: 0, site: (a.site ?? '?').slice(0, 60) };
    cur.count++;
    cur.bytes += a.bytes;
    by.set(key, cur);
  }
  return [...by.entries()]
    .sort((x, y) => y[1].bytes - x[1].bytes)
    .slice(0, n)
    .map(([size, v]) => ({ size, count: v.count, mb: +(v.bytes / 1024 / 1024).toFixed(2), site: v.site }));
}

/** gpuBiggest liệt kê vài texture to nhất — để biết cắt chỗ nào trước. */
export function gpuBiggest(n = 8): { w: number; h: number; mb: number }[] {
  return [...alloc.values()]
    .sort((a, b) => b.bytes - a.bytes)
    .slice(0, n)
    .map((a) => ({ w: a.w, h: a.h, mb: +(a.bytes / 1024 / 1024).toFixed(2) }));
}

/**
 * vramLine: một dòng F3 cho sổ ngân sách §13.10.
 *
 * Rỗng khi mọi hạng mục trong trần — bảng F3 đã dài, và một dòng luôn hiện là một dòng
 * không ai còn đọc. Vượt thì nó nêu ĐÍCH DANH hạng mục nào, vì "VRAM 290 MB" không nói
 * được phải sửa ở đâu.
 */
export function vramLine(led: { rows: { label: string; mb: number; limitMB: number; over: boolean }[]; totalMB: number; limitMB: number }): string {
  const bad = led.rows.filter((r) => r.over);
  if (!bad.length && led.totalMB <= led.limitMB) {
    return '';
  }
  const parts = bad.map((r) => `${r.label} ${r.mb.toFixed(1)}/${r.limitMB}`);
  return `⚠ VRAM ${led.totalMB.toFixed(1)}/${led.limitMB} MB — ${parts.join(' · ')}`; // i18n-ok: lớp số liệu debug
}
