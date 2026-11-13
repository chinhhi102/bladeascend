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
import { ColorMatrixFilter, Mesh, MeshGeometry, Shader, Texture } from 'pixi.js';

import { declareTexture } from './gpumem';

const VERT = `
in vec2 aPosition;
out vec2 vLocal;

uniform mat3 uProjectionMatrix;
uniform mat3 uWorldTransformMatrix;
uniform mat3 uTransformMatrix;

void main() {
  mat3 mvp = uProjectionMatrix * uWorldTransformMatrix * uTransformMatrix;
  gl_Position = vec4((mvp * vec3(aPosition, 1.0)).xy, 0.0, 1.0);
  vLocal = aPosition;
}
`;

// `uGround`: x,y = góc ô trong toạ độ thế giới · z = số đơn vị thế giới cho một lần lặp
// tile · w = bề rộng ô. `uChunk` = (w, h) của ô; `uLayers` bật/tắt từng lớp trộn.
//
// Toạ độ lặp `tuv` tính theo THẾ GIỚI chứ không theo ô: tính theo ô thì hoa văn nhảy lại từ
// đầu ở mỗi mép và cả bản đồ hiện ra một lưới ô vuông — lỗi đã mắc một lần ở bản
// `TilingSprite`, và nó chỉ lộ ra khi camera lùi đủ xa.
//
// Hệ số trộn phải nhân với ALPHA của chính tile đó. Tile lớp trộn không phải một mảng màu
// đặc — `bamboo_litter` là vụn lá rải trên nền TRONG SUỐT. Trộn thẳng RGB của nó thì phần
// trong suốt (màu đen/xám) tràn ra khắp ô: đo được `scenediff` 9,2/255 với 17% pixel đổi,
// và nhìn bằng mắt thì cả bản đồ xám đi. Nhân với `t.a` là tái hiện đúng phép chồng alpha
// mà bản `TilingSprite` + mặt nạ vẫn làm.
//
// Chú thích để NGOÀI chuỗi GLSL: chữ trong chuỗi vừa đi thẳng vào gói tải của người chơi,
// vừa làm `check:i18n` tưởng là chữ hiện cho người chơi (quy tắc 5) — nó chặn đúng.
const FRAG = `
in vec2 vLocal;
out vec4 finalColor;

uniform sampler2D uTile0;
uniform sampler2D uTile1;
uniform sampler2D uTile2;
uniform sampler2D uMask;
uniform vec4 uGround;
uniform vec2 uChunk;
uniform vec2 uLayers;
uniform vec2 uOffset;
uniform vec4 uCM0;
uniform vec4 uCM1;
uniform vec4 uCM2;
uniform vec4 uCMOff;

void main() {
  vec2 tuvBase = (uGround.xy + vLocal) / uGround.z;
  vec2 tuvBlend = vLocal / uGround.z + uOffset;
  vec4 c = texture(uTile0, tuvBase);
  vec2 m = texture(uMask, vLocal / uChunk).rg;
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
 * groundMask dựng mặt nạ trộn cho MỘT ô: kênh R cho lớp 2, kênh G cho lớp 3.
 *
 * `seeds[i]` phải là ĐÚNG hạt giống bản mặt nạ cũ dùng cho lớp đó (`<map>:<cx>:<cy>:<i>`).
 * Lần đầu tôi gieo một hạt chung rồi lệch pha cho lớp thứ hai — hoa văn ra khác hẳn, và
 * `scenediff` báo 21/255 với 47% pixel đổi. Lượt này chỉ đổi CÁCH VẼ, không đổi hình.
 */
export function groundMask(seeds: number[], covers: number[]): Texture {
  const n = 128;
  const wave = (x: number, y: number, f: number, p: number): number =>
    Math.sin(x * f + p) * Math.cos(y * f * 1.3 + p * 1.7) + Math.sin((x + y) * f * 0.7 + p * 2.1);

  // Dựng từng lớp Y HỆT bản mặt nạ cũ — kể cả bước làm nhoè.
  //
  // Bước nhoè của bản cũ KHÔNG phải một phép blur thuần: `ctx.filter='blur(2px)'` rồi
  // `drawImage(cv, 0, 0)` là vẽ bản nhoè **chồng lên bản gốc** theo `source-over`, và vì
  // nền trong suốt (alpha 0) nên phần gốc vẫn hiện qua — mép chỉ MỀM ĐI chứ không nhoè hẳn.
  // Bản đầu của tôi gói hai lớp vào một canvas với alpha 255 khắp nơi, nên `source-over`
  // THAY SẠCH bằng bản nhoè: mảng loang to ra, nhiều vùng nửa vời hơn, và `scenediff` báo
  // 9,2/255 lệch rải đều khắp khung. Nên vẫn phải dựng từng lớp trên nền trong suốt, rồi
  // mới đọc kênh alpha ra gói vào R/G.
  const layer = (seed: number, cover: number): Uint8ClampedArray => {
    const cv = document.createElement('canvas');
    cv.width = n;
    cv.height = n;
    const ctx = cv.getContext('2d') as CanvasRenderingContext2D;
    const img = ctx.createImageData(n, n);
    const s = (seed % 1000) * 0.013;
    for (let y = 0; y < n; y++) {
      for (let x = 0; x < n; x++) {
        const v = wave(x, y, 0.055, s) * 0.65 + wave(x, y, 0.17, s * 1.9) * 0.35;
        const a = v > (0.5 - cover) * 2 ? 255 : 0;
        const k = (y * n + x) * 4;
        img.data[k] = img.data[k + 1] = img.data[k + 2] = 255;
        img.data[k + 3] = a;
      }
    }
    ctx.putImageData(img, 0, 0);
    ctx.filter = 'blur(2px)';
    ctx.drawImage(cv, 0, 0);
    return ctx.getImageData(0, 0, n, n).data;
  };

  const cv = document.createElement('canvas');
  cv.width = n;
  cv.height = n;
  const ctx = cv.getContext('2d') as CanvasRenderingContext2D;
  const out = ctx.createImageData(n, n);
  for (let i = 0; i < 2; i++) {
    const seed = seeds[i];
    if (seed === undefined) {
      continue;
    }
    const d = layer(seed, covers[i] ?? 0);
    for (let k = 0; k < n * n; k++) {
      out.data[k * 4 + i] = d[k * 4 + 3]; // kênh ALPHA của lớp → kênh R (lớp 2) / G (lớp 3)
    }
  }
  for (let k = 0; k < n * n; k++) {
    out.data[k * 4 + 3] = 255; // mặt nạ nay là DỮ LIỆU, không phải hình để nhìn
  }
  ctx.putImageData(out, 0, 0);
  const t = Texture.from(cv);
  declareTexture(t, 'ground'); // §13.10
  return t;
}

/**
 * groundChunkMesh dựng một ô nền.
 *
 * `tiles[0]` là lớp nền, `tiles[1..2]` là lớp trộn (có thể thiếu). `covers` là độ phủ của
 * từng lớp trộn, đúng con số bản cũ dùng (0,42 và 0,30).
 */
export function groundChunkMesh(
  tiles: Texture[], x: number, y: number, w: number, h: number,
  tileWorld: number, seeds: number[], covers: number[],
): { mesh: Mesh<MeshGeometry, Shader>; mask: Texture } {
  // `MeshGeometry` chứ không `Geometry` trần: `Mesh` của Pixi v8 khai kiểu chặt theo
  // `MeshGeometry`, và nó cấp sẵn cả `aUV` (shader này không dùng tới, không tốn gì).
  const geometry = new MeshGeometry({
    positions: new Float32Array([0, 0, w, 0, w, h, 0, h]),
    uvs: new Float32Array([0, 0, 1, 0, 1, 1, 0, 1]),
    indices: new Uint32Array([0, 1, 2, 0, 2, 3]),
  });
  const mask = groundMask(seeds, covers);
  const cm = groundColorMatrix();
  // Tile phải LẶP. Không đặt thì mép ô kéo giãn một hàng điểm ảnh ra khắp ô — nhìn ra ngay,
  // nhưng chỉ ở rìa bản đồ nơi ít ai đi qua, nên dễ lọt.
  for (const t of tiles) {
    t.source.style.addressMode = 'repeat';
  }
  const t0 = tiles[0];
  const t1 = tiles[1] ?? t0;
  const t2 = tiles[2] ?? t0;
  const shader = Shader.from({
    gl: { vertex: VERT, fragment: FRAG },
    resources: {
      uTile0: t0.source,
      uTile0Style: t0.source.style,
      uTile1: t1.source,
      uTile1Style: t1.source.style,
      uTile2: t2.source,
      uTile2Style: t2.source.style,
      uMask: mask.source,
      uMaskStyle: mask.source.style,
      groundUniforms: {
        uGround: { value: new Float32Array([x, y, tileWorld, w]), type: 'vec4<f32>' },
        uChunk: { value: new Float32Array([w, h]), type: 'vec2<f32>' },
        // Lớp trộn thiếu thì tắt hẳn bằng hệ số 0, thay vì trộn tile nền vào chính nó —
        // trộn vào chính nó thì vô hại, nhưng nó giấu mất việc art đang thiếu.
        uLayers: { value: new Float32Array([tiles[1] ? 1 : 0, tiles[2] ? 1 : 0]), type: 'vec2<f32>' },
        // Gốc lát của RIÊNG hai lớp trộn, tính đúng như bản `TilingSprite` cũ: `tilePosition = -x*tileScale`
        // với `tileScale = tileWorld/texWidth` cho ra uv = local/tileWorld + x/texWidth.
        // `x/texWidth` chứ không `x/tileWorld` — tile 512 px phủ 400 đơn vị, nên hai số đó
        // KHÁC nhau và hoa văn lệch một nhịp ở mỗi mép ô. Đó là nếp của bản cũ; giữ nguyên
        // để lượt này chỉ đổi CÁCH VẼ (xem ghi chú ở `mapview.buildBlendChunk`).
        //
        // Lớp NỀN thì không: nó vốn là một `TilingSprite` phủ cả bản đồ, căn theo toạ độ
        // thế giới, nên trong shader nó cũng phải căn theo thế giới. Cho cả ba lớp dùng
        // chung một gốc lát làm `scenediff` dừng ở 1,48/255 thay vì 0,06.
        uOffset: { value: new Float32Array([x / t0.source.width, y / t0.source.height]), type: 'vec2<f32>' },
        // Ma trận màu 4×5 của `ColorMatrixFilter`, đưa thẳng vào shader. Lớp nền không còn
        // filter nào nữa (§13.10 mục 6): một container có filter là một đích vẽ toàn màn
        // hình, mà phép này chỉ đổi màu TỪNG ĐIỂM ẢNH — không cần biết điểm bên cạnh, nên
        // nó thuộc về shader chứ không thuộc về một lượt vẽ riêng.
        uCM0: { value: new Float32Array([cm[0], cm[1], cm[2], cm[3]]), type: 'vec4<f32>' },
        uCM1: { value: new Float32Array([cm[5], cm[6], cm[7], cm[8]]), type: 'vec4<f32>' },
        uCM2: { value: new Float32Array([cm[10], cm[11], cm[12], cm[13]]), type: 'vec4<f32>' },
        uCMOff: { value: new Float32Array([cm[4], cm[9], cm[14], 0]), type: 'vec4<f32>' },
      },
    },
  });
  const mesh = new Mesh({ geometry, shader });
  mesh.position.set(x, y);
  return { mesh, mask };
}
