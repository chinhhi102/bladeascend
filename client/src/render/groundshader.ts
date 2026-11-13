// Shader trộn nền — GAME_SPEC §13.10, thay cho "bake RenderTexture + mask".
//
// ── Ba đời của lớp nền này, và vì sao lại đổi lần nữa ───────────────────────────────────
//
//  1. **Mặt nạ sprite** (tới 12/11): mỗi lớp trộn là một `TilingSprite` bị một `Sprite` mặt
//     nạ che. Mặt nạ sprite trong Pixi v8 là một lượt vẽ ra texture riêng cho MỖI vật bị
//     che, MỖI khung — đo được **100 lô vẽ với 48 mask**, 41% toàn bộ lô vẽ của cảnh.
//  2. **Nướng sẵn** (13/11, lượt #39): vẽ mấy lớp đó một lần ra `RenderTexture` rồi dùng một
//     sprite phẳng. Lô vẽ 100 → 5. Nhưng mỗi ô giữ một texture 1 MB, và sổ VRAM §13.10 đo ra
//     **69 MB cho riêng lớp nền** (trần 4 MB) — đổi một vấn đề lấy một vấn đề khác.
//  3. **Shader trộn** (đây): một mesh mỗi ô, một shader lấy mẫu 2–3 tile theo toạ độ THẾ GIỚI
//     và trộn theo một mặt nạ nhỏ. Không RenderTexture, không mask, không lượt vẽ phụ.
//
// Bộ nhớ: tile dùng chung cho cả bản đồ (3 × 512²), mặt nạ **một texture 128² cho cả hai
// lớp** — kênh R là lớp 2, kênh G là lớp 3. Hai mặt nạ riêng tốn gấp đôi số texture cho
// cùng một lượng tin; gói hai kênh vào một là thứ GPU vốn làm sẵn.
//
// Tile phải lặp được (`addressMode: 'repeat'`), nên chúng phải là TỆP RIÊNG chứ không phải
// ô trong một atlas. Hiện `assets/dist` để `tiles/*.png` riêng từng tệp — nếu sau này gom
// tile vào atlas thì shader này hỏng im lặng (hoa văn kéo giãn thay vì lặp), nên có
// `check:tiles` canh.
import { BufferImageSource, ColorMatrixFilter, Mesh, MeshGeometry, Shader, Texture } from 'pixi.js';

import { declareTexture } from './gpumem';

const VERT = `
in vec2 aPosition;
in vec2 aLocal;
in vec2 aOffset;
in vec2 aMaskUV;
in float aMaskSel;

out vec2 vWorld;
out vec2 vLocal;
out vec2 vOffset;
out vec2 vMaskUV;
out float vMaskSel;

uniform mat3 uProjectionMatrix;
uniform mat3 uWorldTransformMatrix;
uniform mat3 uTransformMatrix;

void main() {
  mat3 mvp = uProjectionMatrix * uWorldTransformMatrix * uTransformMatrix;
  gl_Position = vec4((mvp * vec3(aPosition, 1.0)).xy, 0.0, 1.0);
  vWorld = aPosition;
  vLocal = aLocal;
  vOffset = aOffset;
  vMaskUV = aMaskUV;
  vMaskSel = aMaskSel;
}
`;

// Mỗi ô nền là một quad trong CÙNG một geometry. Thuộc tính theo đỉnh thay cho uniform theo
// ô — đó là cả lý do tồn tại của lượt này: uniform đổi theo ô nghĩa là mỗi ô một lô vẽ, và
// cảnh đông có 72 ô.
//
//   aPosition  toạ độ THẾ GIỚI (lớp nền lát theo thế giới)
//   aLocal     toạ độ trong ô    (lớp trộn lát theo ô — xem `aOffset`)
//   aOffset    gốc lát của lớp trộn: `x/texWidth`, KHÔNG phải `x/tileWorld`
//   aMaskUV    ô của mặt nạ này trong atlas
//
// `aOffset` giữ nguyên nếp của bản `TilingSprite` cũ: tile 512 px phủ 400 đơn vị, nên
// `tilePosition = -x·tileScale` cho ra `uv = local/400 + x/512`. Lệch một nhịp ở mỗi mép ô,
// nhưng đó là hình mà chủ dự án đã duyệt; sửa nó là một thay đổi đồ hoạ, không phải một lượt
// tối ưu.
//
// Ma trận màu (`uCM*`) áp cuối cùng, đúng phép mà `ColorMatrixFilter` của lớp nền vẫn làm.
const FRAG = `
in vec2 vWorld;
in vec2 vLocal;
in vec2 vOffset;
in vec2 vMaskUV;
in float vMaskSel;
out vec4 finalColor;

uniform sampler2D uTile0;
uniform sampler2D uTile1;
uniform sampler2D uTile2;
uniform sampler2D uMask;
uniform vec4 uGround;
uniform vec2 uLayers;
uniform vec4 uCM0;
uniform vec4 uCM1;
uniform vec4 uCM2;
uniform vec4 uCMOff;

void main() {
  vec2 tuvBase = vWorld / uGround.z;
  vec2 tuvBlend = vLocal / uGround.z + vOffset;
  vec4 c = texture(uTile0, tuvBase);
  vec4 mt = texture(uMask, vMaskUV);
  vec2 m = mix(mt.rg, mt.ba, vMaskSel);
  vec4 t1 = texture(uTile1, tuvBlend);
  vec4 t2 = texture(uTile2, tuvBlend);
  c.rgb = mix(c.rgb, t1.rgb, m.r * t1.a * uLayers.x);
  c.rgb = mix(c.rgb, t2.rgb, m.g * t2.a * uLayers.y);
  vec3 g = c.rgb;
  c.r = dot(g, uCM0.rgb) + uCM0.a * c.a + uCMOff.r;
  c.g = dot(g, uCM1.rgb) + uCM1.a * c.a + uCMOff.g;
  c.b = dot(g, uCM2.rgb) + uCM2.a * c.a + uCMOff.b;
  finalColor = c;
}
`;


/**
 * groundColorMatrix: ĐÚNG phép chỉnh màu mà `ColorMatrixFilter` của lớp nền vẫn làm
 * (ART_PROMPTS §9.0: nền dịu, tương phản thấp để vật thể gameplay nổi lên).
 *
 * Lấy ma trận từ CHÍNH `ColorMatrixFilter` rồi truyền vào shader, thay vì tự nhân ba phép
 * `saturate`/`brightness`/`contrast` bằng tay: nhân tay là một cơ hội sai số mà `scenediff`
 * sẽ bắt được nhưng không nói được sai ở đâu, và nếu một ngày ai đó đổi tham số ở
 * `backgroundFilter()` thì bản chép tay im lặng trôi khỏi bản thật.
 */
let cmCache: number[] | null = null;
export function groundColorMatrix(): number[] {
  if (cmCache) {
    return cmCache;
  }
  const f = new ColorMatrixFilter();
  f.saturate(-0.5);
  f.brightness(0.9, true);
  f.contrast(-0.2, true);
  cmCache = [...f.matrix];
  return cmCache;
}

/**
 * maskPlanes dựng mặt nạ trộn cho MỘT ô: hai mặt phẳng 128², mỗi mặt một lớp trộn.
 *
 * Trả dữ liệu thô chứ không trả texture: atlas gói **hai ô vào một ô atlas** (ô thứ nhất ở
 * kênh R,G — ô thứ hai ở B,A), nên chỗ ghi phải chọn được kênh.
 *
 * `seeds[i]` phải là ĐÚNG hạt giống bản mặt nạ cũ dùng cho lớp đó (`<map>:<cx>:<cy>:<i>`).
 * Lần đầu tôi gieo một hạt chung rồi lệch pha cho lớp thứ hai — hoa văn ra khác hẳn, và
 * `scenediff` báo 21/255 với 47% pixel đổi. Lượt này chỉ đổi CÁCH VẼ, không đổi hình.
 */
/**
 * MASK_PX_GEN: độ phân giải mặt nạ trộn của MỘT ô — 64, không phải 128.
 *
 * Atlas mặt nạ là một texture cố định; ở ô 128² nó là 1024² = 4 MB, chiếm hơn nửa hạng mục
 * "Nền" (trần 4 MB). Ở 64² thì atlas còn 512² = 1 MB.
 *
 * Tần số sóng nhân đôi để BÙ: nhiễu tính theo toạ độ texel, nên hạ nửa độ phân giải mà giữ
 * nguyên tần số là mảng loang to gấp đôi trong toạ độ thế giới — đổi hình cả bản đồ. Nhân
 * đôi tần số thì hoa văn giữ nguyên cỡ, chỉ mép mềm hơn một chút.
 */
const MASK_PX_GEN = 64;
/** Hệ số bù tần số cho `MASK_PX_GEN` (128 là mốc gốc của hoa văn đã duyệt). */
const WAVE_SCALE = 128 / MASK_PX_GEN;
export function maskPlanes(seeds: number[], covers: number[]): Uint8ClampedArray {
  const n = MASK_PX_GEN;
  const wave = (x: number, y: number, f: number, p: number): number =>
    Math.sin(x * f + p) * Math.cos(y * f * 1.3 + p * 1.7) + Math.sin((x + y) * f * 0.7 + p * 2.1);

  // Dựng từng lớp Y HỆT bản mặt nạ cũ — kể cả bước làm nhoè.
  //
  // Bước nhoè của bản cũ KHÔNG phải một phép blur thuần: `ctx.filter='blur(2px)'` rồi
  // `drawImage(cv, 0, 0)` là vẽ bản nhoè **chồng lên** bản gốc theo `source-over`, và vì nền
  // trong suốt (alpha 0) nên phần gốc vẫn hiện qua — mép chỉ MỀM ĐI chứ không nhoè hẳn. Một
  // bản gói hai lớp vào một canvas alpha 255 sẽ bị `source-over` THAY SẠCH bằng bản nhoè:
  // mảng loang to ra, và `scenediff` báo 9,3/255 lệch rải đều khắp khung.
  const out = new Uint8ClampedArray(n * n * 2);
  for (let i = 0; i < 2; i++) {
    const seed = seeds[i];
    if (seed === undefined) {
      continue;
    }
    const cover = covers[i] ?? 0;
    const cv = document.createElement('canvas');
    cv.width = n;
    cv.height = n;
    const ctx = cv.getContext('2d') as CanvasRenderingContext2D;
    const img = ctx.createImageData(n, n);
    const sd = (seed % 1000) * 0.013;
    for (let y = 0; y < n; y++) {
      for (let x = 0; x < n; x++) {
        const v = wave(x, y, 0.055 * WAVE_SCALE, sd) * 0.65 + wave(x, y, 0.17 * WAVE_SCALE, sd * 1.9) * 0.35;
        const k = (y * n + x) * 4;
        img.data[k] = img.data[k + 1] = img.data[k + 2] = 255;
        img.data[k + 3] = v > (0.5 - cover) * 2 ? 255 : 0;
      }
    }
    ctx.putImageData(img, 0, 0);
    ctx.filter = 'blur(2px)';
    ctx.drawImage(cv, 0, 0);
    const d = ctx.getImageData(0, 0, n, n).data;
    for (let k = 0; k < n * n; k++) {
      out[k * 2 + i] = d[k * 4 + 3];
    }
  }
  return out;
}

const ATLAS_PX = 512;
const MASK_PX = 64;
/** Số ô atlas (8×8). Mỗi ô atlas chứa HAI ô nền, nên sức chứa thật là gấp đôi. */
const ATLAS_TILES = (ATLAS_PX / MASK_PX) ** 2;
const ATLAS_SLOTS = ATLAS_TILES * 2;

/**
 * GroundField: TOÀN BỘ lớp nền trong **một** mesh và **một** atlas mặt nạ.
 *
 * Bản trước mỗi ô là một mesh riêng — đúng về hình, nhưng cảnh đông (camera Đạo Tổ lùi xa
 * nhất) có 72 ô trong tầm và lớp nền ngốn 74 lô vẽ, kéo tổng cảnh lên 424. Mỗi ô một mesh
 * là vì mặt nạ của nó là một uniform riêng; đưa mặt nạ vào một ATLAS và chỉ số ô vào một
 * thuộc tính đỉnh thì cả lưới ô gom được vào một lô.
 *
 * Bộ nhớ mặt nạ nay **cố định 4 MB** (atlas 1024² RGBA, 64 ô 128²) thay vì tăng theo số ô —
 * quan trọng không kém số lô vẽ, vì trần hạng mục "Nền" của §13.10 là 4 MB.
 *
 * Ô thứ 65 ghi đè ô LÂU KHÔNG DÙNG NHẤT. 64 ô 1 024 đơn vị phủ một vùng 8 192² — rộng hơn
 * khung nhìn xa nhất của game, nên trong thực tế vòng đè chỉ xảy ra khi người chơi đi liên
 * tục một quãng dài, và khi đó ô bị đè đã ở xa sau lưng.
 */

export interface ChunkSpec {
  key: string;
  x: number;
  y: number;
  w: number;
  h: number;
  seeds: number[];
  covers: number[];
}

export class GroundField {
  readonly mesh: Mesh<MeshGeometry, Shader>;
  /**
   * Atlas là một BUFFER thô, không phải canvas.
   *
   * Canvas 2D lưu màu đã nhân alpha, và `getImageData` ở chỗ alpha = 0 trả về RGB = 0 —
   * tức là ghi ô chẵn (kênh R,G) rồi quay lại ghi ô lẻ (kênh B,A) sẽ XOÁ SẠCH ô chẵn. Ở đây
   * alpha không phải độ trong suốt mà là dữ liệu, nên phải tránh hẳn đường canvas.
   */
  private readonly atlasBuf: Uint8Array;
  private readonly atlasTex: Texture;
  /** khoá ô → chỉ số ô trong atlas. */
  private readonly slotOf = new Map<string, number>();
  /** chỉ số ô → khoá ô đang chiếm (rỗng = trống) và lần dùng gần nhất. */
  private readonly slotKey: string[] = [];
  private readonly slotUsed: number[] = [];
  private seq = 0;
  /** chữ ký lưới ô đang vẽ — geometry chỉ dựng lại khi tập ô đổi. */
  private gridKey = '';

  constructor(tiles: Texture[], tileWorld: number) {
    this.atlasBuf = new Uint8Array(ATLAS_PX * ATLAS_PX * 4);
    this.atlasTex = new Texture({
      source: new BufferImageSource({
        resource: this.atlasBuf,
        width: ATLAS_PX,
        height: ATLAS_PX,
        format: 'rgba8unorm',
        // Không nhân alpha: kênh A ở đây là DỮ LIỆU mặt nạ của ô lẻ, không phải độ trong suốt.
        alphaMode: 'no-premultiply-alpha',
        // KHÔNG mipmap: lấy mẫu mipmap sẽ trộn các ô cạnh nhau ở mức thấp, tức là mặt nạ
        // của ô này loang sang ô kia. Và nó cũng tiết kiệm 1/3 bộ nhớ.
        autoGenerateMipmaps: false,
      }),
    });
    declareTexture(this.atlasTex, 'ground'); // §13.10
    for (let i = 0; i < ATLAS_SLOTS; i++) {
      this.slotKey.push('');
      this.slotUsed.push(0);
    }
    for (const t of tiles) {
      t.source.style.addressMode = 'repeat';
    }
    const t0 = tiles[0];
    const t1 = tiles[1] ?? t0;
    const t2 = tiles[2] ?? t0;
    const cm = groundColorMatrix();
    const shader = Shader.from({
      gl: { vertex: VERT, fragment: FRAG },
      resources: {
        uTile0: t0.source,
        uTile0Style: t0.source.style,
        uTile1: t1.source,
        uTile1Style: t1.source.style,
        uTile2: t2.source,
        uTile2Style: t2.source.style,
        uMask: this.atlasTex.source,
        uMaskStyle: this.atlasTex.source.style,
        groundUniforms: {
          uGround: { value: new Float32Array([0, 0, tileWorld, 0]), type: 'vec4<f32>' },
          uLayers: { value: new Float32Array([tiles[1] ? 1 : 0, tiles[2] ? 1 : 0]), type: 'vec2<f32>' },
          uCM0: { value: new Float32Array([cm[0], cm[1], cm[2], cm[3]]), type: 'vec4<f32>' },
          uCM1: { value: new Float32Array([cm[5], cm[6], cm[7], cm[8]]), type: 'vec4<f32>' },
          uCM2: { value: new Float32Array([cm[10], cm[11], cm[12], cm[13]]), type: 'vec4<f32>' },
          uCMOff: { value: new Float32Array([cm[4], cm[9], cm[14], 0]), type: 'vec4<f32>' },
        },
      },
    });
    this.texW = t0.source.width;
    this.texH = t0.source.height;
    this.mesh = new Mesh({ geometry: new MeshGeometry({ positions: new Float32Array(0), uvs: new Float32Array(0), indices: new Uint32Array(0) }), shader });
  }

  private texW = 512;
  private texH = 512;

  /** slotFor trả ô atlas của một ô nền, vẽ mặt nạ vào atlas nếu chưa có. */
  private slotFor(c: ChunkSpec): number {
    const hit = this.slotOf.get(c.key);
    if (hit !== undefined) {
      this.slotUsed[hit] = ++this.seq;
      return hit;
    }
    let slot = this.slotKey.indexOf('');
    if (slot < 0) {
      slot = 0;
      for (let i = 1; i < ATLAS_SLOTS; i++) {
        if (this.slotUsed[i] < this.slotUsed[slot]) {
          slot = i;
        }
      }
      this.slotOf.delete(this.slotKey[slot]);
    }
    this.slotKey[slot] = c.key;
    this.slotUsed[slot] = ++this.seq;
    this.slotOf.set(c.key, slot);
    // Mỗi ô atlas chứa HAI ô nền: ô chẵn ở kênh R,G — ô lẻ ở B,A.
    //
    // Cần thế vì cảnh đông (camera Đạo Tổ lùi xa nhất) có **72 ô trong tầm**, mà atlas
    // 1024² chỉ có 64 ô 128². Thiếu chỗ không chỉ là chậm: `slotFor` sẽ đuổi chính những ô
    // vừa cấp trong cùng một lượt, và khi đó geometry trỏ vào mặt nạ của ô KHÁC — mặt đất
    // hiện ra hoa văn sai mà không có lỗi nào được ném. Gói đôi cho sức chứa 128 ô.
    const tile = slot >> 1;
    const half = slot & 1;
    const per = ATLAS_PX / MASK_PX;
    const sx = (tile % per) * MASK_PX;
    const sy = Math.floor(tile / per) * MASK_PX;
    const planes = maskPlanes(c.seeds, c.covers);
    for (let y = 0; y < MASK_PX; y++) {
      let d = ((sy + y) * ATLAS_PX + sx) * 4 + half * 2;
      let srcI = y * MASK_PX * 2;
      for (let x = 0; x < MASK_PX; x++) {
        this.atlasBuf[d] = planes[srcI];
        this.atlasBuf[d + 1] = planes[srcI + 1];
        d += 4;
        srcI += 2;
      }
    }
    this.atlasTex.source.update();
    return slot;
  }

  /**
   * setChunks đặt tập ô đang hiển thị. Dựng lại geometry CHỈ khi tập ô đổi.
   *
   * Đổi tập ô xảy ra mỗi lần camera vượt một mép ô — vài giây một lần, không phải mỗi khung.
   */
  setChunks(list: ChunkSpec[]): void {
    const key = list.map((c) => c.key).join('|');
    if (key === this.gridKey) {
      return;
    }
    this.gridKey = key;
    const n = list.length;
    const pos = new Float32Array(n * 8);
    const local = new Float32Array(n * 8);
    const off = new Float32Array(n * 8);
    const muv = new Float32Array(n * 8);
    const sel = new Float32Array(n * 4);
    const idx = new Uint32Array(n * 6);
    const per = ATLAS_PX / MASK_PX;
    for (let i = 0; i < n; i++) {
      const c = list[i];
      const slot = this.slotFor(c);
      // Dùng TRỌN ô atlas, không lùi nửa texel.
      //
      // Lùi nửa texel tránh được việc lấy mẫu tuyến tính hút sang ô bên cạnh, nhưng nó co
      // mặt nạ lại 127/128 và dịch đi nửa texel — đo được `scenediff` 0,47/255 thay vì
      // 0,06. Mặt nạ là một trường mờ nhoè, nên phần hút sang ô bên chỉ ảnh hưởng nửa texel
      // sát mép và nhỏ hơn hẳn sai số do co lại.
      const tile = slot >> 1;
      const u0 = ((tile % per) * MASK_PX) / ATLAS_PX;
      const v0 = (Math.floor(tile / per) * MASK_PX) / ATLAS_PX;
      const u1 = ((tile % per) * MASK_PX + MASK_PX) / ATLAS_PX;
      const v1 = (Math.floor(tile / per) * MASK_PX + MASK_PX) / ATLAS_PX;
      const q = i * 8;
      pos.set([c.x, c.y, c.x + c.w, c.y, c.x + c.w, c.y + c.h, c.x, c.y + c.h], q);
      local.set([0, 0, c.w, 0, c.w, c.h, 0, c.h], q);
      const ox = c.x / this.texW;
      const oy = c.y / this.texH;
      off.set([ox, oy, ox, oy, ox, oy, ox, oy], q);
      muv.set([u0, v0, u1, v0, u1, v1, u0, v1], q);
      const hf = slot & 1;
      sel.set([hf, hf, hf, hf], i * 4);
      const b = i * 4;
      idx.set([b, b + 1, b + 2, b, b + 2, b + 3], i * 6);
    }
    const g = new MeshGeometry({ positions: pos, uvs: muv, indices: idx });
    g.addAttribute('aLocal', { buffer: local, format: 'float32x2' });
    g.addAttribute('aOffset', { buffer: off, format: 'float32x2' });
    g.addAttribute('aMaskUV', { buffer: muv, format: 'float32x2' });
    g.addAttribute('aMaskSel', { buffer: sel, format: 'float32' });
    const old = this.mesh.geometry;
    this.mesh.geometry = g;
    old.destroy();
  }

  destroy(): void {
    this.mesh.destroy({ children: true });
    this.atlasTex.destroy(true);
  }
}
