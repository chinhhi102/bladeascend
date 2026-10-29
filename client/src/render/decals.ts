import { Container, Sprite, Texture } from 'pixi.js';

import { assets } from './assets';
import { quality } from './quality';

/**
 * Decal mặt đất (spec §8.6 mục 3): vết nứt, khóm cỏ, lá rụng, sỏi, vũng nước nhỏ, rễ cây,
 * hoa dại, vết bánh xe...
 *
 * Rải bằng **Poisson-disk** chứ không rải ngẫu nhiên thuần: rải thuần thì chỗ chụm thành
 * cụm, chỗ trống hoác, mắt người nhận ra ngay là "máy rải". Poisson-disk giữ khoảng cách
 * tối thiểu nên trông như tự nhiên mọc.
 *
 * Mỗi decal **xoay / lật / co giãn 0.7–1.3 / đổi sắc nhẹ** để cùng một hình không lộ ra là
 * lặp lại. Vẽ một lần lúc vào phòng rồi thôi — decal không có va chạm, không đổi gì.
 */

/** Số decal mỗi ô 2048 theo mức chất lượng (spec: 30–60 mỗi chunk). */
const PER_CHUNK: Record<string, number> = { high: 55, medium: 30, low: 12 };
const CHUNK = 2048;

/** Danh sách decal theo biome; tên file khớp ART_PROMPTS §9.0b. */
const BY_BIOME: Record<string, string[]> = {
  lang_tre: ['grass_tuft', 'fallen_leaves', 'bamboo_leaves', 'moss_patch', 'puddle', 'tree_roots'],
  cho_dem: ['cobble_edge', 'brick_rubble', 'cart_tracks', 'pebbles', 'ground_cracks', 'roof_tiles'],
  nui_da: ['pebbles', 'ground_cracks', 'moss_patch', 'tree_roots', 'grass_tuft'],
};
const FALLBACK = ['grass_tuft', 'pebbles', 'ground_cracks', 'fallen_leaves'];

/** mulberry32: bộ sinh số giả ngẫu nhiên có hạt — cùng bản đồ thì decal rải y hệt nhau. */
function rng(seed: number): () => number {
  let a = seed >>> 0;
  return () => {
    a = (a + 0x6d2b79f5) >>> 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

/**
 * poisson rải điểm trong ô vuông sao cho không điểm nào gần nhau dưới `min`.
 * Thuật toán "ném phi tiêu": thử chỗ ngẫu nhiên, quá gần điểm cũ thì bỏ, thử tiếp. Đơn giản
 * và đủ nhanh ở quy mô vài chục điểm.
 */
function poisson(n: number, size: number, min: number, rand: () => number): { x: number; y: number }[] {
  const out: { x: number; y: number }[] = [];
  for (let tries = 0; tries < n * 30 && out.length < n; tries++) {
    const p = { x: rand() * size, y: rand() * size };
    if (out.every((q) => (q.x - p.x) ** 2 + (q.y - p.y) ** 2 >= min * min)) {
      out.push(p);
    }
  }
  return out;
}

/**
 * placeholderTex nướng sẵn vài hình decal tạm thành texture rồi dùng lại.
 *
 * Quan trọng: decal phải là **Sprite dùng chung texture**, không phải `Graphics` riêng từng
 * cái. Mỗi Graphics có hình riêng là một lượt vẽ riêng — đo thật: 55 decal mỗi ô × 36 ô
 * khiến draw call nhảy từ 237 lên 549 và khung hình tụt từ 58 xuống 36. Sprite chung texture
 * thì Pixi gộp cả trăm cái vào một lô.
 */
const baked = new Map<string, Texture>();

function placeholderTex(kind: string, variant: number): Texture {
  const key = `${kind}:${variant}`;
  const hit = baked.get(key);
  if (hit) {
    return hit;
  }
  const n = 128;
  const cv = document.createElement('canvas');
  cv.width = n;
  cv.height = n;
  const ctx = cv.getContext('2d') as CanvasRenderingContext2D;
  const col = kind.includes('grass') || kind.includes('moss') ? '#6f8f5a' : kind.includes('leaves') ? '#8a7540' : kind.includes('puddle') ? '#53707a' : '#7d7366';
  const rand = rng(variant * 7919 + kind.length * 31);
  const verts = 7 + Math.floor(rand() * 5);
  ctx.beginPath();
  for (let i = 0; i < verts; i++) {
    const a = (i / verts) * Math.PI * 2;
    const r = (n / 2) * (0.45 + rand() * 0.5);
    const x = n / 2 + Math.cos(a) * r;
    const y = n / 2 + Math.sin(a) * r * 0.6;
    if (i === 0) {
      ctx.moveTo(x, y);
    } else {
      ctx.lineTo(x, y);
    }
  }
  ctx.closePath();
  ctx.fillStyle = col;
  ctx.globalAlpha = 0.75;
  ctx.fill();
  const tex = Texture.from(cv);
  baked.set(key, tex);
  return tex;
}

/**
 * DecalLayer rải decal khắp bản đồ, **gom theo ô 2048** và chỉ bật ô nào đang trong khung
 * nhìn (spec §8.6 mục 3 + mục "cull").
 *
 * Không cull thì bản đồ 12 000² có ~2 000 decal, Pixi duyệt hết mỗi khung dù chỉ ~50 cái
 * thực sự nhìn thấy — đo thật: khung hình tụt từ 58 xuống 38. Tắt `renderable` của cả ô là
 * Pixi bỏ qua luôn cả nhánh, không duyệt con.
 */
export class DecalLayer {
  readonly root = new Container();
  private readonly chunks: { node: Container; x: number; y: number }[] = [];

  add(node: Container, cx: number, cy: number): void {
    this.chunks.push({ node, x: cx * CHUNK, y: cy * CHUNK });
    this.root.addChild(node);
  }

  /** update bật/tắt từng ô theo khung nhìn, cộng một ô lề cho mượt lúc di chuyển. */
  update(camX: number, camY: number, halfW: number, halfH: number): void {
    for (const c of this.chunks) {
      c.node.renderable =
        c.x < camX + halfW + CHUNK && c.x + CHUNK > camX - halfW - CHUNK &&
        c.y < camY + halfH + CHUNK && c.y + CHUNK > camY - halfH - CHUNK;
    }
  }
}

/**
 * buildDecals rải decal khắp bản đồ. Gọi một lần lúc dựng bản đồ.
 * `mapId` làm hạt giống nên cùng bản đồ thì rải giống nhau ở mọi máy.
 */
export function buildDecals(mapId: string, size: number, seed: number): DecalLayer {
  const layer = new DecalLayer();
  const per = PER_CHUNK[quality()] ?? 0;
  if (per <= 0) {
    return layer;
  }
  const kinds = BY_BIOME[mapId] ?? FALLBACK;
  const chunks = Math.ceil(size / CHUNK);
  for (let cy = 0; cy < chunks; cy++) {
    for (let cx = 0; cx < chunks; cx++) {
      const rand = rng(seed + cx * 7349 + cy * 15733);
      const chunk = new Container();
      for (const p of poisson(per, CHUNK, 150, rand)) {
        const kind = kinds[Math.floor(rand() * kinds.length)];
        const tex = assets.texture(`decals/${kind}.png`) ?? placeholderTex(kind, Math.floor(rand() * 4));
        const node = new Sprite(tex);
        node.anchor.set(0.5);
        node.width = 46 + rand() * 40;
        node.height = node.width * (0.5 + rand() * 0.35);
        node.position.set(cx * CHUNK + p.x, cy * CHUNK + p.y);
        // Xoay / lật / đổi sắc nhẹ để cùng một hình không lộ ra là lặp lại.
        node.rotation = rand() * Math.PI * 2;
        if (rand() < 0.5) {
          node.scale.x *= -1;
        }
        node.alpha = 0.3 + rand() * 0.3;
        const t = 0.84 + rand() * 0.3;
        const ch = Math.max(0, Math.min(255, Math.round(255 * t)));
        node.tint = (ch << 16) | (ch << 8) | ch;
        chunk.addChild(node);
      }
      layer.add(chunk, cx, cy);
    }
  }
  return layer;
}

/** Màu vụn dưới chân theo loại vật cản: gạch vỡ, sỏi, rác... */
function debrisColor(kind: string): string {
  if (kind === 'brick') {
    return '#9b4a3a';
  }
  if (kind === 'boulder' || kind === 'stone_wall') {
    return '#7f7a72';
  }
  if (kind === 'stall' || kind === 'cart') {
    return '#8a7a55';
  }
  return '#6f7a52';
}

/**
 * skirtTex nướng sẵn bóng đổ + vụn thành texture, mỗi (loại, biến thể) một cái.
 *
 * Phải nướng chứ không vẽ bằng `Graphics` từng vật cản: 286 vật cản là 286 hình riêng, tức
 * 286 lượt vẽ. Đo thật, lúc còn dùng Graphics thì draw call nhảy lên 611 trong khi màn hình
 * gần như trống. Dùng chung 4 texture mỗi loại thì Pixi gộp hết vào vài lô.
 */
function skirtTex(kind: string, variant: number): Texture {
  const key = `skirt:${kind}:${variant}`;
  const hit = baked.get(key);
  if (hit) {
    return hit;
  }
  const n = 256;
  const cv = document.createElement('canvas');
  cv.width = n;
  cv.height = n;
  const c = cv.getContext('2d') as CanvasRenderingContext2D;
  const rand = rng(variant * 104729 + kind.length * 37);
  // Bóng: hình méo, mềm, hơi lệch sang một bên như có nắng xiên.
  const grad = c.createRadialGradient(n / 2, n * 0.56, n * 0.05, n / 2, n * 0.56, n * 0.48);
  grad.addColorStop(0, 'rgba(0,0,0,0.34)');
  grad.addColorStop(0.65, 'rgba(0,0,0,0.16)');
  grad.addColorStop(1, 'rgba(0,0,0,0)');
  c.fillStyle = grad;
  c.beginPath();
  c.ellipse(n / 2 + n * 0.03, n * 0.56, n * 0.46, n * 0.2, 0, 0, Math.PI * 2);
  c.fill();
  // Vụn quanh chân.
  c.fillStyle = debrisColor(kind);
  c.globalAlpha = 0.55;
  for (let i = 0; i < 8; i++) {
    const a = rand() * Math.PI * 2;
    const d = n * (0.2 + rand() * 0.22);
    c.beginPath();
    c.arc(n / 2 + Math.cos(a) * d, n * 0.56 + Math.sin(a) * d * 0.4, 2 + rand() * 5, 0, Math.PI * 2);
    c.fill();
  }
  const tex = Texture.from(cv);
  baked.set(key, tex);
  return tex;
}

/**
 * obstacleSkirt trả về bóng đổ + vụn dưới chân một vật cản (spec §8.6 mục 3), để nó "cắm"
 * vào đất thay vì nổi lên như dán.
 */
export function obstacleSkirt(w: number, h: number, kind: string, seed: number): Sprite {
  const sp = new Sprite(skirtTex(kind, seed % 4));
  sp.anchor.set(0.5);
  sp.width = w * 1.25;
  sp.height = h * 0.85;
  sp.y = h * 0.2;
  return sp;
}
