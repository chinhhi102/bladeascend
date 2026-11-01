import { AnimatedSprite, ColorMatrixFilter, Container, Graphics, Rectangle, Sprite, Texture, TilingSprite } from 'pixi.js';

import type { ClientWorld } from '../net/state';
import type { GameMap, MapObstacle, Shape } from '../net/protocol';
import { assets } from './assets';
import { BIOME_TILES, OBSTACLE_ALIAS } from './catalog';
import { buildDecals, type DecalLayer, obstacleSkirt } from './decals';
import { GATE_COLORS } from './palette';
import { emergency } from './quality';

/**
 * Mỗi biome 2–3 texture nền, trộn bằng mặt nạ nhiễu (spec §8.6 mục 1). Lớp đầu phủ kín, các
 * lớp sau nằm trong mặt nạ loang nên ranh giới là mảng màu tự nhiên chứ không phải ô vuông.
 * Thiếu texture nào thì bỏ qua lớp đó, không vỡ gì.
 *
 * Bảng nằm ở `catalog.ts` — một nguồn duy nhất, dùng chung với `assets:report` /
 * `check:assets`. Khai lại ở đây là cách hai bảng đã trôi khỏi nhau lần trước.
 */
/**
 * BIOME_COLOR — màu nền khi KHÔNG có ảnh nền nào (chỉ dùng lúc art chưa về).
 *
 * Khoá là giá trị `biome` trong `.tmj`. Chỗ gọi trước đây tra bằng **tên tệp tile**
 * (`BIOME_TILES[biome][0]` → `'grass'`, `'dirt'`, `'concrete'`… tuỳ bảng), nên khớp hay
 * trượt là chuyện tình cờ: `grass`/`dirt`/`concrete` trùng tên tệp nên khớp, còn Băng Nguyệt
 * Hồ thì trượt và ra xanh cỏ. Giờ tra thẳng bằng `map.biome`, và có thêm `snow`.
 */
const BIOME_COLOR: Record<string, number> = { grass: 0x7cc46a, concrete: 0xa9a49b, dirt: 0xc9a46a, snow: 0xdfe9f2 };
// Màu placeholder theo LOẠI. Mỗi bậc một sắc rõ ràng để lúc chưa có art vẫn đọc được bản
// đồ: tre vàng-xanh, gạch đỏ, đá xám, băng xanh nhạt, ngọc lục, thạch bi tím.
const OBSTACLE_COLOR: Record<string, number> = {
  fence: 0xc8a165, stall: 0xe76f51, cart: 0x9c6644, bamboo_thin: 0xa8c66c,
  brick: 0xb5523b, bamboo_wall: 0x7d9b4e, boulder: 0x8d8d8d, bamboo_ancient: 0x5c7a3a,
  stone_wall: 0x5b6b5b, sect_wall: 0x7a6a52, cliff: 0x6b6256, boulder_huge: 0x6e6e6e,
  ice_eternal: 0xa8d8e8, jade_pillar: 0x4fae8b, seal_stele: 0x7a5aa8, border: 0x3a4038,
};
const TILE_WORLD = 400; // one 512px ground tile covers 400 world units
/**
 * GROUND_CHUNK: cạnh một ô nền trộn, tính bằng đơn vị thế giới.
 *
 * 1 024 chứ không phải 2 048: mặt nạ sprite cấp phát render texture bằng kích thước vật bị
 * che, nên ô càng nhỏ thì tấm texture càng nhỏ — 1 024² ≈ 5,6 MB so với 2 048² ≈ 22 MB.
 * Nhỏ hơn nữa thì số ô (và số lệnh vẽ) tăng nhanh hơn phần bộ nhớ tiết kiệm được.
 */
const GROUND_CHUNK = 1024;
/** GROUND_PAD: giữ thêm một vành ô quanh khung nhìn để lúc camera chạy không thấy ô hiện ra. */
const GROUND_PAD = 1;
const ART_FILL = 0.9; // object art fills ~90% of its canvas

function bounds(s: Shape): { x: number; y: number; w: number; h: number } {
  if (s.c) {
    const r = s.r ?? 0;
    return { x: s.x - r, y: s.y - r, w: r * 2, h: r * 2 };
  }
  return { x: s.x, y: s.y, w: s.w ?? 0, h: s.h ?? 0 };
}

/**
 * drawShape vẽ đúng hình mà SERVER dùng để xét va chạm — đa giác hữu cơ nếu có (spec §8.6),
 * không thì hình tròn / chữ nhật.
 */
/**
 * noiseMask dựng mặt nạ loang để trộn hai texture nền (spec §8.6 mục 1).
 *
 * Vẽ bằng canvas ở độ phân giải thấp (128×128 cho cả bản đồ 12 000) rồi để GPU kéo giãn —
 * mặt nạ mờ nhoè chính là thứ ta cần, và 128² thì dựng xong trong một nhịp mắt. Dùng hai
 * tầng tần số: tầng thô tạo mảng lớn, tầng mịn làm mép lởm chởm.
 */
function noiseMask(size: number, seed: number, cover: number): Sprite {
  const n = 128;
  const cv = document.createElement('canvas');
  cv.width = n;
  cv.height = n;
  const ctx = cv.getContext('2d') as CanvasRenderingContext2D;
  const img = ctx.createImageData(n, n);
  const wave = (x: number, y: number, f: number, p: number): number =>
    Math.sin(x * f + p) * Math.cos(y * f * 1.3 + p * 1.7) + Math.sin((x + y) * f * 0.7 + p * 2.1);
  const s = (seed % 1000) * 0.013;
  for (let y = 0; y < n; y++) {
    for (let x = 0; x < n; x++) {
      const v = wave(x, y, 0.055, s) * 0.65 + wave(x, y, 0.17, s * 1.9) * 0.35;
      // v nằm quanh [-2, 2]; đẩy qua ngưỡng để ra mảng chứ không ra gradient đều.
      const a = v > (0.5 - cover) * 2 ? 255 : 0;
      const i = (y * n + x) * 4;
      img.data[i] = img.data[i + 1] = img.data[i + 2] = 255;
      img.data[i + 3] = a;
    }
  }
  ctx.putImageData(img, 0, 0);
  // Làm nhoè mép để hai texture hoà vào nhau thay vì cắt ngọt.
  ctx.filter = 'blur(2px)';
  ctx.drawImage(cv, 0, 0);
  const sp = new Sprite(Texture.from(cv));
  sp.width = size;
  sp.height = size;
  return sp;
}

/** hashSeed biến id bản đồ thành số, để mọi máy rải decal giống hệt nhau. */
function hashSeed(id: string): number {
  let h = 2166136261;
  for (let i = 0; i < id.length; i++) {
    h = Math.imul(h ^ id.charCodeAt(i), 16777619);
  }
  return h >>> 0;
}

function drawShape(g: Graphics, s: Shape): Graphics {
  if (s.p && s.p.length >= 6) {
    return g.poly(s.p);
  }
  return s.c ? g.circle(s.x, s.y, s.r ?? 0) : g.rect(s.x, s.y, s.w ?? 0, s.h ?? 0);
}

/** fitSprite stretches a canvas-sized texture so its art covers the shape bounds. */
function fitSprite(tex: Texture, s: Shape): Sprite {
  const b = bounds(s);
  const sp = new Sprite(tex);
  sp.anchor.set(0.5);
  sp.position.set(b.x + b.w / 2, b.y + b.h / 2);
  sp.width = b.w / ART_FILL;
  sp.height = b.h / ART_FILL;
  return sp;
}

const TILE_MIN = 120; // smallest drawn obstacle piece, world units
const TILE_PER_THICKNESS = 2.4; // a 3/4-view piece is drawn taller than its footprint
// Khoảng cách giữa hai mảnh = ĐÚNG chiều dài mảnh, không nhân hệ số: nhân vào là dãy tường
// bị hở. Vẽ thừa nửa mảnh ở hai đầu nên hai đoạn kề nhau luôn chạm nhau.
const TILE_STEP = 0.55; // bước lát theo chiều dài mảnh; < phần vẽ thật nên luôn chồng mép
const VOID_MARGIN = 300; // phải khớp mapMargin của cmd/mapgen

/**
 * tiledObstacle keeps the art's aspect ratio and repeats it along long fences/walls,
 * standing each piece on the hitbox like a 3/4 view, instead of stretching one image.
 */
function tiledObstacle(tex: Texture, s: Shape, vertical: boolean): Container {
  const b = bounds(s);
  const long = vertical ? b.h : b.w;
  const thick = vertical ? b.w : b.h;
  const size = Math.min(Math.max(thick * TILE_PER_THICKNESS, TILE_MIN), long / ART_FILL);
  // Bước lát: phần vẽ thật trong mỗi ảnh hẹp hơn canvas (khối đá chỉ lấp ~60% bề ngang),
  // nên lấy bước nhỏ hơn để hai mảnh kề nhau chồng mép — dãy tường liền, không hở.
  const step = size * TILE_STEP;
  const n = Math.max(1, Math.ceil(long / step));
  const c = new Container();
  for (let i = 0; i < n; i++) {
    const sp = new Sprite(tex);
    // Mảnh dọc: neo giữa để xếp chồng theo trục Y; mảnh ngang: neo chân như cũ.
    sp.anchor.set(0.5, vertical ? 0.5 : 0.95);
    sp.width = sp.height = size;
    const t = Math.min(long, (i + 0.5) * step);
    if (vertical) {
      sp.position.set(b.x + b.w / 2, b.y + t);
    } else {
      sp.position.set(b.x + t, b.y + b.h);
    }
    c.addChild(sp);
  }
  return c;
}

// Từ v6.7 vật cản chỉ có HAI vẻ: còn đó, hoặc đã biến mất. Không nứt, không đống đổ nát —
// spec §8.2: "biến mất ngay lập tức, không hoạt ảnh vỡ, không đống đổ nát".
type ObstacleLook = 'intact' | 'gone';

interface ObstacleNode {
  o: MapObstacle;
  node: Container;
  look: ObstacleLook | '';
  /** tierDrawn: bậc có chém được hay không ở lần vẽ gần nhất, để biết khi nào phải vẽ lại viền. */
  tierDrawn: boolean;
  warnUntil: number;
  box: { x: number; y: number; w: number; h: number }; // drawn bounds, for see-through
  /** drawnWithArt: lần vẽ gần nhất dùng ảnh thật hay placeholder (§13.7 mục 3). */
  drawnWithArt: boolean;
  /** fadeFrom: mốc bắt đầu mờ dần khi ảnh thật vừa thay placeholder. 0 = không mờ. */
  fadeFrom: number;
  /** edge: viền tối của vật cản QUÁ BẬC; chỉ hiện khi nằm trong tầm kiếm (spec §8.2). */
  edge: Container | null;
}

/** MyReach: tâm và tầm kiếm của mình — vật cản quá bậc chỉ hiện viền khi lọt vào đây. */
export interface MyReach {
  x: number;
  y: number;
  r: number;
}

/** ART_FADE_SECS: ảnh thật thay placeholder thì mờ dần vào, không nhấp nháy (§13.7 mục 3). */
const ART_FADE_SECS = 0.2;

/** Viewer is the local character's drawn box, used to fade obstacles standing in front of it. */
export interface Viewer {
  x: number;
  top: number;
  feet: number;
  halfW: number;
}

// ART_PROMPTS §9.0: calm, low-contrast ground so gameplay objects stand out.
function backgroundFilter(): ColorMatrixFilter {
  const f = new ColorMatrixFilter();
  f.saturate(-0.5);
  f.brightness(0.9, true);
  f.contrast(-0.2, true);
  return f;
}

/** obstacleTexture: art THẬT của một loại vật cản, null nếu chưa có. */
function obstacleTexture(kind: string, vertical: boolean): Texture | null {
  return assets.texture(`obstacles/${kind}/intact${vertical ? '_vertical' : ''}.png`)
    ?? (vertical ? null : assets.texture(`obstacles/${kind}/intact.png`));
}

/** grownShape: nở hình ra quanh TÂM theo hệ số `k`. */
function grownShape(s: Shape, k: number): Shape {
  if (s.c) {
    return { ...s, r: (s.r ?? 0) * k };
  }
  const w = (s.w ?? 0) * k;
  const h = (s.h ?? 0) * k;
  return { ...s, x: s.x + ((s.w ?? 0) - w) / 2, y: s.y + ((s.h ?? 0) - h) / 2, w, h };
}

/** tintAll nhuộm mọi sprite con — `tiledObstacle` trả về Container nhiều mảnh. */
function tintAll(node: Container, tint: number): void {
  if (node instanceof Sprite || node instanceof TilingSprite) {
    node.tint = tint;
  }
  for (const c of node.children) {
    tintAll(c as Container, tint);
  }
}

/**
 * tooHardOutline vẽ viền tối mờ quanh vật cản mà cảnh giới hiện tại chưa chém được
 * (spec §8.2 "Nhận biết bậc").
 *
 * Viền thay cho bảng máu của bản cũ: không còn máu nên không còn gì để đếm, nhưng người
 * chơi vẫn cần biết TRƯỚC khi mài kiếm rằng cái này không chém được. Vẽ bằng một nét dày
 * mờ phía trong, không phải một lớp phủ đen — phủ lên cả vật cản thì mất hết chi tiết art.
 */
/**
 * silhouette nhuộm đen mọi sprite trong một bản sao của art vật cản.
 *
 * Viền "quá bậc" (spec §8.2) phải ôm theo ĐƯỜNG NÉT CỦA ẢNH. Bản trước vẽ hình chữ nhật
 * theo hình va chạm trong Tiled — mà hình đó thấp hơn phần ảnh (kiểu vẽ 3/4) — nên trên
 * màn hình là một khối tối lệch hẳn khỏi tảng băng (chủ dự án chụp 1/11 16:00). Đặt bản
 * sao đen này SAU ảnh và to hơn một chút thì chỉ phần rìa lộ ra: một viền tối đúng dáng.
 */
function silhouette(art: Container): Container {
  // Pixi 8: `tint` của Container truyền xuống mọi con (Sprite, TilingSprite, Graphics).
  art.tint = 0x000000;
  art.alpha = 0.6;
  return art;
}

/** inReach: hình vật cản có chạm vòng tròn tầm kiếm không (điểm gần nhất của khung bao). */
function inReach(o: MapObstacle, me: MyReach): boolean {
  const b = bounds(o.s);
  const dx = Math.max(b.x - me.x, 0, me.x - (b.x + b.w));
  const dy = Math.max(b.y - me.y, 0, me.y - (b.y + b.h));
  return dx * dx + dy * dy <= me.r * me.r;
}

/**
 * MapWorld: phần trạng thái trận mà bộ vẽ bản đồ thật sự đọc. Tách hẹp ra để trang quản trị
 * (ADMIN_SPEC §5 — xem toàn cảnh bằng chính bộ vẽ này) đưa vào một trạng thái giả, không cần
 * dựng cả `ClientWorld` với kết nối mạng.
 */
export type MapWorld = Pick<ClientWorld, 'brokenObstacles' | 'canCutObstacle'>;

/** EDGE_GROW: viền to hơn ảnh ngần này (tỉ lệ) — đủ thấy, không thành bóng đổ. */
const EDGE_GROW = 1.07;

/**
 * MapView draws the static map. Obstacle nodes are handed to the scene's entity layer so they
 * y-sort with characters; chúng chỉ có hai vẻ — còn đó hoặc biến mất (spec §8.2).
 */
/** Vật cản Shipper đang đi xuyên qua mờ còn ngần này (spec v6.36 §6.5). */
const GHOST_PASS_ALPHA = 0.5;

export class MapView {
  readonly ground = new Container();
  /** blendRoot chứa các ô nền trộn (lớp 2–3), dựng và huỷ theo khung nhìn. */
  private readonly blendRoot = new Container();
  /** blendNames: tên tile của lớp trộn; rỗng nghĩa là biome này chỉ có một lớp nền. */
  private blendNames: string[] = [];
  /** blendChunks: ô đang dựng, khoá `"cx,cy"`. Ra ngoài khung nhìn là **huỷ**, không ẩn. */
  private readonly blendChunks = new Map<string, Container>();
  readonly decor = new Container();
  readonly under = new Container(); // zones under entities (mud, wind)
  /**
   * gateLayer: cổng dịch chuyển, nằm TRÊN vòng kiếm (chủ dự án yêu cầu 1/11).
   *
   * Trước đây cổng nằm chung lớp `under` với bùn/gió, tức là dưới vòng kiếm. Từ Hoá Thần
   * trở lên vòng kiếm phủ kín cả khoảng quanh người, nên cổng bị che mất đúng lúc người
   * chơi đang cần nhắm tới nó.
   */
  readonly gateLayer = new Container();
  readonly over = new Container(); // bush canopy above entities
  /** Ao/suối: mặt nước gợn sóng, cập nhật mỗi khung (spec §8.6 mục 4). */
  private readonly water: { root: Container; ripple: Graphics; s: Shape }[] = [];
  private decals!: DecalLayer;
  private readonly obstacleNodes = new Map<number, ObstacleNode>();
  private readonly bushes: { id: number; node: Container }[] = [];
  private readonly gates = new Map<number, { node: Container; warnUntil: number }>();

  constructor(
    private readonly map: GameMap,
    private readonly gateR: number,
  ) {
    this.buildGround();
    // Decal rải khắp mặt đất (spec §8.6 mục 3): nằm trên nền, dưới mọi vùng và vật cản.
    this.decals = buildDecals(map.id, map.size, hashSeed(map.id));
    this.ground.addChild(this.decals.root);
    this.buildZones();
    this.buildGates();
    this.buildDecor();
    for (const o of map.obstacles) {
      const n: ObstacleNode = { o, node: new Container(), look: '', tierDrawn: true, warnUntil: 0, box: bounds(o.s), drawnWithArt: false, fadeFrom: 0, edge: null };
      this.obstacleNodes.set(o.id, n);
    }
    const filter = backgroundFilter();
    this.ground.filters = [filter];
    this.decor.filters = [filter];
    // `filterArea` BẮT BUỘC phải đặt — xem clampFilterArea. Đặt tạm một ô nhỏ ở đây để
    // khung hình đầu tiên (trước khi `update` chạy lần nào) cũng không cấp phát cả bản đồ.
    this.ground.filterArea = new Rectangle(0, 0, 1, 1);
    this.decor.filterArea = new Rectangle(0, 0, 1, 1);
  }

  /**
   * updateGroundChunks dựng ô nền trộn quanh khung nhìn và **huỷ** ô đi xa.
   *
   * Huỷ chứ không chỉ ẩn: `renderable = false` vẫn giữ nguyên texture trên GPU, mà texture
   * mới đúng là thứ đang phải cắt. Một ô giữ 1–2 mặt nạ 1 024² (~5,6 MB mỗi cái), nên 30 ô
   * mồ côi là ~300 MB nằm im trên card.
   */
  private updateGroundChunks(view: { x: number; y: number; w: number; h: number }): void {
    if (this.blendNames.length === 0) {
      return;
    }
    const last = Math.floor(this.map.size / GROUND_CHUNK);
    const cx0 = Math.max(0, Math.floor((view.x - view.w / 2) / GROUND_CHUNK) - GROUND_PAD);
    const cx1 = Math.min(last, Math.floor((view.x + view.w / 2) / GROUND_CHUNK) + GROUND_PAD);
    const cy0 = Math.max(0, Math.floor((view.y - view.h / 2) / GROUND_CHUNK) - GROUND_PAD);
    const cy1 = Math.min(last, Math.floor((view.y + view.h / 2) / GROUND_CHUNK) + GROUND_PAD);

    for (const [key, node] of this.blendChunks) {
      const [kx, ky] = key.split(',').map(Number);
      if (kx < cx0 || kx > cx1 || ky < cy0 || ky > cy1) {
        node.destroy({ children: true });
        this.blendChunks.delete(key);
      }
    }
    for (let cy = cy0; cy <= cy1; cy++) {
      for (let cx = cx0; cx <= cx1; cx++) {
        const key = `${cx},${cy}`;
        if (!this.blendChunks.has(key)) {
          const node = this.buildBlendChunk(cx, cy);
          this.blendChunks.set(key, node);
          this.blendRoot.addChild(node);
        }
      }
    }
  }

  /** buildBlendChunk dựng một ô: mỗi lớp trộn là một TilingSprite che bởi mặt nạ cỡ ô. */
  private buildBlendChunk(cx: number, cy: number): Container {
    const node = new Container();
    const x = cx * GROUND_CHUNK;
    const y = cy * GROUND_CHUNK;
    const w = Math.min(GROUND_CHUNK, this.map.size - x);
    const h = Math.min(GROUND_CHUNK, this.map.size - y);
    this.blendNames.forEach((name, i) => {
      const t2 = assets.texture(`tiles/${name}.png`);
      if (!t2 || w <= 0 || h <= 0) {
        return;
      }
      const layer = new TilingSprite({ texture: t2, width: w, height: h });
      layer.tileScale.set(TILE_WORLD / t2.width);
      // Gốc lát bám TOẠ ĐỘ THẾ GIỚI, không bám góc ô: không có dòng này thì hoa văn nhảy
      // lại từ đầu ở mỗi mép ô và cả bản đồ hiện ra một lưới ô vuông.
      layer.tilePosition.set(-x * layer.tileScale.x, -y * layer.tileScale.y);
      layer.position.set(x, y);
      // Mặt nạ gieo theo Ô: cùng một ô luôn ra cùng mảng loang, dù dựng lại bao nhiêu lần
      // khi người chơi đi qua đi lại. Gieo theo thời gian thì nền "đổi kiểu" sau lưng họ.
      const mask = noiseMask(Math.max(w, h), hashSeed(`${this.map.id}:${cx}:${cy}:${i}`), 0.42 - i * 0.12);
      mask.position.set(x, y);
      mask.width = w;
      mask.height = h;
      layer.mask = mask;
      node.addChild(layer, mask);
    });
    return node;
  }

  /**
   * clampFilterArea giới hạn vùng bộ lọc màu nền vào đúng phần đang nhìn thấy.
   *
   * Không có nó, PixiJS cấp phát render texture bằng KÍCH THƯỚC CẢ CONTAINER để chạy bộ lọc.
   * `ground` trải khắp bản đồ 12 000×12 000, làm tròn lên lũy thừa 2 là **16 384×16 384 ×4
   * byte = 1 024 MB** cho một tấm texture duy nhất — đo được trên máy thật ngày 30/10, lớn
   * gấp ba lần toàn bộ art của game cộng lại.
   *
   * Hậu quả không phải là "hơi chậm": laptop dùng GPU tích hợp thường có `MAX_TEXTURE_SIZE`
   * 8 192 hoặc ít VRAM hơn 1 GB, nên lệnh cấp phát HỎNG — và mặt đất hiện ra trắng trơn chỉ
   * còn nét viền, đúng ảnh chụp của đội test lúc 12:18. Máy bàn có GPU rời nuốt được 1 GB
   * nên lỗi này không bao giờ xuất hiện ở máy dev.
   *
   * Giới hạn vào khung nhìn thì tấm texture chỉ còn cỡ màn hình (~1280×720), và hình hiện ra
   * y hệt: bộ lọc này chỉ đổi màu từng điểm ảnh, không lấy mẫu điểm bên cạnh, nên cắt vùng
   * không làm sai kết quả.
   */
  private clampFilterArea(view: { x: number; y: number; w: number; h: number }): void {
    // Nới mỗi bên một chút để lúc camera chạy không thấy mép vùng lọc quét theo.
    const pad = 64;
    const x = view.x - view.w / 2 - pad;
    const y = view.y - view.h / 2 - pad;
    const w = view.w + pad * 2;
    const h = view.h + pad * 2;
    for (const c of [this.ground, this.decor]) {
      const r = c.filterArea as Rectangle;
      r.x = x;
      r.y = y;
      r.width = w;
      r.height = h;
    }
  }

  /** obstacleLayer returns every obstacle node, to be added to the y-sorted entity layer. */
  obstacleLayer(): Container[] {
    return [...this.obstacleNodes.values()].map((n) => n.node);
  }

  /**
   * buildGround trải nền: lớp 1 phủ kín cả bản đồ, lớp 2–3 trộn bằng mặt nạ loang theo
   * **từng ô 1 024 đơn vị** quanh khung nhìn (spec §8.6 mục 1).
   *
   * Vì sao phải chia ô — đây là chỗ ngốn bộ nhớ GPU lớn nhất của cả game, đo ngày 31/10:
   * mặt nạ sprite của PixiJS cấp phát một render texture **bằng kích thước vật bị che**.
   * Lớp trộn trước đây trải khắp bản đồ 12 000, làm tròn lên luỹ thừa 2 là
   * **16 384 × 16 384 × 4 byte = 1 365 MB** cho MỘT tấm. Tổng bộ nhớ texture đo được là
   * **1,4 GB lúc vừa vào trận, 1,9 GB khi lên Đạo Tổ**.
   *
   * iPhone không báo lỗi khi hết bộ nhớ GPU — Safari **giết hẳn tab**. Đó khớp với báo cáo
   * "trên iPhone luôn bị lỗi" và với nhật ký server thấy một người vào lại phòng 8 lần
   * trong 40 giây.
   *
   * Chia ô 1 024 thì mỗi mặt nạ chỉ còn 1 024² (~5,6 MB), và ô ngoài khung nhìn bị **huỷ
   * hẳn** chứ không chỉ ẩn đi — ẩn thì texture vẫn nằm trên GPU, đúng thứ ta đang cắt.
   */
  private buildGround(): void {
    const size = this.map.size;
    const names = (BIOME_TILES[this.map.biome] ?? ['grass']).filter((n) => assets.texture(`tiles/${n}.png`));
    const tex = names.length ? assets.texture(`tiles/${names[0]}.png`) : null;
    if (tex) {
      // Lớp nền: MỘT TilingSprite phủ cả bản đồ. Không có mặt nạ nên không cấp phát gì —
      // `TilingSprite` chỉ lặp lại một texture 512², không phụ thuộc kích thước phủ.
      const t = new TilingSprite({ texture: tex, width: size, height: size });
      t.tileScale.set(TILE_WORLD / tex.width);
      this.ground.addChild(t);
      this.blendNames = names.slice(1).filter((n) => assets.texture(`tiles/${n}.png`));
      this.ground.addChild(this.blendRoot);
    } else {
      const g = new Graphics().rect(0, 0, size, size).fill(BIOME_COLOR[this.map.biome] ?? 0x7cc46a);
      for (let v = 0; v <= size; v += TILE_WORLD) {
        g.moveTo(v, 0).lineTo(v, size).moveTo(0, v).lineTo(size, v);
      }
      g.stroke({ color: 0x000000, alpha: 0.06, width: 4 });
      this.ground.addChild(g);
    }
    this.drawVoid(size);
  }

  /**
   * drawVoid fills everything outside the playable area with a dark abyss instead of flat
   * grey: layered bands that fade to black, plus drifting mist blobs and a few cloud wisps
   * so the edge reads as "vực sâu" rather than an unfinished tile (spec §8.1).
   */
  private drawVoid(size: number): void {
    const m = VOID_MARGIN;
    const far = size * 1.5; // vẽ rộng ra ngoài để camera lùi hết cỡ vẫn không thấy mép
    const g = new Graphics();
    // Nền vực: phủ kín phần ngoài khung chơi.
    g.rect(-far, -far, size + 2 * far, far + m)
      .rect(-far, size - m, size + 2 * far, far + m)
      .rect(-far, m, far + m, size - 2 * m)
      .rect(size - m, m, far + m, size - 2 * m)
      .fill(0x0b0d12);
    // Dải mờ dần vào trong cho đỡ gắt.
    for (let i = 0; i < 4; i++) {
      const t = i / 4;
      const w = m * (1 - t);
      g.rect(m - w, m - w, size - 2 * (m - w), size - 2 * (m - w))
        .stroke({ color: 0x1b2230, width: Math.max(6, m * 0.22), alpha: 0.5 - 0.1 * i });
    }
    // Mây / sương lững lờ: đốm elip mờ rải quanh mép, vị trí tất định theo kích thước bản đồ.
    let seed = 1337;
    const rnd = (): number => ((seed = (seed * 1103515245 + 12345) & 0x7fffffff) / 0x7fffffff);
    for (let i = 0; i < 90; i++) {
      const side = i % 4;
      const along = rnd() * size;
      const depth = m * (0.1 + rnd() * 1.4);
      const pos: [number, number] =
        side === 0 ? [along, -depth + m] :
        side === 1 ? [along, size - m + depth] :
        side === 2 ? [-depth + m, along] : [size - m + depth, along];
      g.ellipse(pos[0], pos[1], 120 + rnd() * 260, 50 + rnd() * 110)
        .fill({ color: rnd() > 0.5 ? 0x38425c : 0x222a3a, alpha: 0.18 + rnd() * 0.16 });
    }
    this.ground.addChild(g);
    // Viền trong cùng cho rõ ranh giới đi được.
    this.ground.addChild(new Graphics().rect(m, m, size - 2 * m, size - 2 * m)
      .stroke({ color: 0x3b2f2f, width: 50, alpha: 0.85 }));
  }

  private buildZones(): void {
    const mudTex = assets.texture('tiles/mud.png');
    for (const z of this.map.zones) {
      if (z.k === 'mud') {
        this.under.addChild(this.mudNode(z.s, mudTex));
      } else if (z.k === 'wind') {
        this.under.addChild(this.windNode(z.s, z.wd ?? 0));
      } else if (z.k === 'road') {
        // Đường lát đá sáng màu (spec §8.1 bước 2): chỉ là nền, không chặn ai.
        this.under.addChild(drawShape(new Graphics(), z.s).fill({ color: 0x8c8378, alpha: 0.55 }));
      } else if (z.k === 'water') {
        const node = this.waterNode(z.s);
        this.water.push(node);
        this.under.addChild(node.root);
      } else if (z.k === 'bush') {
        const node = this.bushNode(z.s);
        this.bushes.push({ id: z.id, node });
        this.over.addChild(node);
      }
    }
  }

  private mudNode(s: Shape, tex: Texture | null): Container {
    const mask = drawShape(new Graphics(), s).fill(0xffffff);
    if (!tex) {
      return drawShape(new Graphics(), s).fill({ color: 0x6b4a2b, alpha: 0.75 });
    }
    const b = bounds(s);
    const t = new TilingSprite({ texture: tex, width: b.w, height: b.h });
    t.position.set(b.x, b.y);
    t.tileScale.set(TILE_WORLD / 2 / tex.width);
    const c = new Container();
    c.addChild(t, mask);
    t.mask = mask;
    return c;
  }

  private windNode(s: Shape, dir: number): Container {
    const c = new Container();
    c.addChild(drawShape(new Graphics(), s).fill({ color: 0xbfe9ff, alpha: 0.22 }));
    const frames = assets.frames('zones', 'wind');
    const b = bounds(s);
    const step = 260;
    for (let x = b.x + step / 2; x < b.x + b.w; x += step) {
      for (let y = b.y + step / 2; y < b.y + b.h; y += step) {
        const node = frames.length ? new AnimatedSprite(frames) : new Graphics().poly([-40, -14, 30, -14, 30, -30, 60, 0, 30, 30, 30, 14, -40, 14]).fill({ color: 0xffffff, alpha: 0.55 });
        if (node instanceof AnimatedSprite) {
          node.anchor.set(0.5);
          node.width = 140;
          node.height = 140;
          node.animationSpeed = 0.12;
          node.play();
        }
        node.position.set(x, y);
        node.rotation = dir;
        c.addChild(node);
      }
    }
    return c;
  }

  /**
   * waterNode vẽ ao/suối: nền nước xanh mờ bên trong đa giác, cộng vài gợn sóng chạy chậm
   * và đốm lấp lánh. Gợn sóng vẽ lại mỗi khung nhưng chỉ vài nét nên rất rẻ; cả mặt nước
   * nằm trong mặt nạ hình đa giác nên không tràn ra ngoài bờ.
   */
  private waterNode(s: Shape): { root: Container; ripple: Graphics; s: Shape } {
    const root = new Container();
    const mask = drawShape(new Graphics(), s).fill(0xffffff);
    const base = drawShape(new Graphics(), s).fill({ color: 0x3f7f8c, alpha: 0.55 });
    const ripple = new Graphics();
    ripple.mask = mask;
    // Bờ nước: viền đậm hơn một chút cho mặt nước "lõm" xuống thay vì dán lên mặt đất.
    const edge = drawShape(new Graphics(), s).stroke({ color: 0x2b5c66, width: 10, alpha: 0.6 });
    root.addChild(base, ripple, edge, mask);
    return { root, ripple, s };
  }

  /** updateWater chạy gợn sóng; gọi mỗi khung từ MapView.update. */
  private updateWater(now: number): void {
    for (const w of this.water) {
      const b = bounds(w.s);
      const g = w.ripple.clear();
      const lines = 5;
      for (let i = 0; i < lines; i++) {
        const t = (now * 0.12 + i / lines) % 1;
        const y = b.y + b.h * t;
        g.moveTo(b.x, y);
        // Sóng là một đường sin chạy ngang, lệch pha theo thời gian.
        for (let x = b.x; x <= b.x + b.w; x += 24) {
          g.lineTo(x, y + Math.sin(x * 0.02 + now * 1.6 + i) * 4);
        }
      }
      g.stroke({ color: 0xcfeff5, width: 2.5, alpha: 0.28 });
      // Đốm lấp lánh thưa thớt.
      for (let i = 0; i < 4; i++) {
        const a = now * 0.5 + i * 1.9;
        const x = b.x + b.w * (0.2 + 0.6 * ((Math.sin(a) + 1) / 2));
        const y = b.y + b.h * (0.2 + 0.6 * ((Math.cos(a * 1.3) + 1) / 2));
        g.circle(x, y, 3 * Math.abs(Math.sin(now * 2 + i))).fill({ color: 0xffffff, alpha: 0.5 });
      }
    }
  }

  private bushNode(s: Shape): Container {
    const tex = assets.texture('zones/bush.png');
    if (tex) {
      return fitSprite(tex, s);
    }
    const g = drawShape(new Graphics(), s).fill({ color: 0x2f7d32, alpha: 0.92 });
    const b = bounds(s);
    for (let i = 0; i < 6; i++) {
      g.circle(b.x + b.w * (0.2 + 0.6 * ((i * 37) % 10) / 10), b.y + b.h * (0.2 + 0.6 * ((i * 53) % 10) / 10), Math.min(b.w, b.h) * 0.18).fill({ color: 0x43a047, alpha: 0.9 });
    }
    return g;
  }

  private buildGates(): void {
    for (const gt of this.map.gates) {
      const frames = assets.frames(`gates/${gt.c}`, 'idle');
      let node: Container;
      if (frames.length) {
        const sp = new AnimatedSprite(frames);
        sp.anchor.set(0.5);
        sp.animationSpeed = 0.15;
        sp.play();
        sp.width = sp.height = (2 * this.gateR) / ART_FILL;
        node = sp;
      } else {
        const color = GATE_COLORS[gt.c] ?? 0xa55bff;
        node = new Graphics().circle(0, 0, this.gateR).fill({ color, alpha: 0.35 }).stroke({ color, width: 14 }).circle(0, 0, this.gateR * 0.55).stroke({ color: 0xffffff, width: 6, alpha: 0.6 });
      }
      node.position.set(gt.x, gt.y);
      this.gateLayer.addChild(node);
      this.gates.set(gt.id, { node, warnUntil: 0 });
    }
  }

  private buildDecor(): void {
    for (const d of this.map.decor) {
      const tex = assets.texture(`decor/${d.sp}.png`);
      if (!tex) {
        continue;
      }
      const sp = new Sprite(tex);
      sp.anchor.set(0.5);
      sp.scale.set(160 / tex.width);
      sp.position.set(d.x, d.y);
      this.decor.addChild(sp);
    }
  }

  warnGate(id: number, now: number): void {
    const g = this.gates.get(id);
    if (g) {
      g.warnUntil = now + 0.5;
    }
  }

  warnObstacle(id: number, now: number): void {
    const n = this.obstacleNodes.get(id);
    if (n) {
      n.warnUntil = now + 3;
    }
  }

  update(w: MapWorld, now: number, myBush: number, me: Viewer | null, occlusionAlpha: number, view?: { x: number; y: number; w: number; h: number }, reach?: MyReach): void {
    this.updateWater(now);
    if (view) {
      this.decals.setThin(emergency());
      this.decals.update(view.x, view.y, view.w / 2, view.h / 2);
      this.clampFilterArea(view);
      this.updateGroundChunks(view);
    }
    for (const n of this.obstacleNodes.values()) {
      this.updateObstacle(n, w, now);
      if (n.edge) {
        n.edge.visible = reach !== undefined && inReach(n.o, reach);
      }
      if (me && n.look === 'intact' && this.covers(n, me)) {
        n.node.alpha *= occlusionAlpha;
      }
      if (this.ghostPass.length > 0 && this.ghostOver(n)) {
        n.node.alpha *= GHOST_PASS_ALPHA;
      }
    }
    for (const b of this.bushes) {
      // See-through canopy only for the bush I'm hiding in (ART_PROMPTS §9.2).
      b.node.alpha = b.id === myBush ? 0.45 : 1;
    }
    for (const g of this.gates.values()) {
      const warn = now < g.warnUntil;
      g.node.alpha = warn ? 0.6 + 0.4 * Math.sin(now * 40) : 1;
    }
  }

  /**
   * ghostPass: Shipper đang Gom Đơn (đi xuyên vật cản) — vật cản nào họ đang đi qua thì mờ
   * 50% cho thấy người bên trong (spec v6.36 §6.5). Scene đặt mỗi khung, rỗng là tắt.
   */
  ghostPass: { x: number; y: number; r: number }[] = [];

  private ghostOver(n: ObstacleNode): boolean {
    const b = n.box;
    return this.ghostPass.some((g) => g.x + g.r > b.x && g.x - g.r < b.x + b.w && g.y + g.r > b.y && g.y - g.r < b.y + b.h);
  }

  // covers: the obstacle stands in front of me (its base is lower on screen) and overlaps my sprite.
  private covers(n: ObstacleNode, me: Viewer): boolean {
    const b = n.box;
    return b.y + b.h > me.feet && b.x < me.x + me.halfW && b.x + b.w > me.x - me.halfW && b.y < me.feet && b.y + b.h > me.top;
  }

  private updateObstacle(n: ObstacleNode, w: MapWorld, now: number): void {
    const broken = w.brokenObstacles.get(n.o.id) ?? false;
    const look: ObstacleLook = broken ? 'gone' : 'intact';
    // Bậc quá cao thì vẽ VIỀN TỐI MỜ (spec §8.2): người chơi phải nhìn là biết chém được
    // hay không, trước khi mất 3 giây mài vào một bức tường không nhúc nhích.
    const canCut = w.canCutObstacle(n.o.t);
    const hasArt = assets.texture(`obstacles/${n.o.k}/intact.png`) !== null;
    // Ảnh thật vừa tải xong (nhóm `zone_*` về sau khi vào trận, §13.7 mục 3): dựng lại
    // bằng ảnh thật và MỜ DẦN vào. Điều kiện phải là "lần trước vẽ placeholder" chứ không
    // phải "có ảnh": thiếu nó thì mỗi lần `assets.gen` tăng là dựng lại cả 300 vật cản.
    const upgraded = hasArt && !n.drawnWithArt && n.look !== '';
    if (upgraded) {
      n.fadeFrom = now;
    }
    if (look !== n.look || upgraded || canCut !== n.tierDrawn) {
      n.look = look;
      n.tierDrawn = canCut;
      n.drawnWithArt = hasArt;
      n.node.removeChildren().forEach((c) => c.destroy());
      if (look === 'intact') {
        // Bóng đổ mềm + vụn dưới chân để vật cản "cắm" vào đất thay vì nổi lên như dán
        // (spec §8.6 mục 3). Vẽ TRƯỚC phần art nên nằm dưới.
        const b = bounds(n.o.s);
        const skirt = obstacleSkirt(b.w, b.h, n.o.k, n.o.id * 2654435761);
        skirt.position.set(b.x + b.w / 2, b.y + b.h / 2);
        n.node.addChild(skirt);
        n.edge = null;
        if (!canCut) {
          // Viền nằm SAU ảnh, phóng quanh tâm vật cản: chỉ phần rìa lộ ra.
          const edge = new Container();
          edge.position.set(b.x + b.w / 2, b.y + b.h / 2);
          edge.scale.set(EDGE_GROW);
          const sil = silhouette(this.varied(this.obstacleArt(n.o), n.o));
          sil.position.set(sil.position.x - b.x - b.w / 2, sil.position.y - b.y - b.h / 2);
          edge.addChild(sil);
          edge.visible = false;
          n.node.addChild(edge);
          n.edge = edge;
        }
        n.node.addChild(this.varied(this.obstacleArt(n.o), n.o));
      }
      const shown = n.edge?.visible ?? false;
      if (n.edge) {
        n.edge.visible = true; // tính khung vẽ cả viền, để vùng nhìn-xuyên không đổi theo tầm kiếm
      }
      const lb = n.node.getLocalBounds();
      if (n.edge) {
        n.edge.visible = shown;
      }
      n.box = { x: lb.x, y: lb.y, w: lb.width, h: lb.height };
      n.node.zIndex = broken ? -1e6 : bounds(n.o.s).y + bounds(n.o.s).h;
    }
    // Blink 3s before regrowing (spec §8.2).
    n.node.alpha = broken && now < n.warnUntil ? 0.5 + 0.5 * Math.sin(now * 20) : 1;
    if (n.fadeFrom > 0) {
      const k = (now - n.fadeFrom) / ART_FADE_SECS;
      if (k >= 1) {
        n.fadeFrom = 0;
      } else {
        n.node.alpha *= Math.max(0, k);
      }
    }
  }

  /**
   * varied xoay nhẹ ±6° và thỉnh thoảng lật ngang một vật cản, để hàng vật cản cùng loại
   * không lộ ra là copy-paste (spec §8.6 mục 3).
   *
   * Phải xoay **quanh tâm chính nó**: `obstacleArt` dựng hình ở toạ độ THẾ GIỚI, nên đặt
   * `rotation` thẳng lên nó là xoay quanh gốc bản đồ (0,0) — vật cản ở (3000, 5000) sẽ văng
   * đi hàng trăm đơn vị, còn lật ngang thì ném nó sang hẳn nửa kia bản đồ. Đó đúng là lỗi
   * "vật cản nằm ngoài biên" nhìn thấy trong game.
   *
   * Tường dãy và tường biên thì KHÔNG đụng vào: chúng phải nối liền nhau, xoay một đoạn là
   * hở ngay.
   */
  private varied(art: Container, o: MapObstacle): Container {
    const b = bounds(o.s);
    const tiled = !o.s.c && (b.w > b.h * 1.6 || b.h > b.w * 1.6);
    if (tiled || o.k === 'stone_wall' || o.k === 'sect_wall' || o.k === 'cliff' || o.k === 'fence' || o.k === 'brick' || o.k === 'border') {
      return art;
    }
    const cx = b.x + b.w / 2;
    const cy = b.y + b.h / 2;
    const pivot = new Container();
    pivot.position.set(cx, cy);
    art.position.set(art.position.x - cx, art.position.y - cy);
    const r = ((o.id * 9301 + 49297) % 233280) / 233280;
    pivot.rotation = (r - 0.5) * 0.21;
    if ((o.id & 1) === 1) {
      pivot.scale.x = -1;
    }
    pivot.addChild(art);
    return pivot;
  }

  private obstacleArt(o: MapObstacle): Container {
    const b = bounds(o.s);
    // Đoạn dọc cần art riêng: xoay sprite ngang 90° trông sai hẳn (bóng đổ, mặt đứng).
    const vertical = !o.s.c && b.h > b.w;
    const real = obstacleTexture(o.k, vertical);
    if (real) {
      return o.s.c ? fitSprite(real, o.s) : tiledObstacle(real, o.s, vertical);
    }
    // Chưa có art THẬT → MƯỢN art loại gần giống (`OBSTACLE_ALIAS`), có phóng cỡ và nhuộm
    // màu. Thứ tự này quan trọng: art thật luôn thắng, nên artgen sinh xong là ảnh mới tự
    // thay vào mà không phải sửa gì ở đây.
    const al = OBSTACLE_ALIAS[o.k];
    const lent = al ? obstacleTexture(al.from, vertical) : null;
    if (al && lent) {
      // Phóng cỡ bằng cách nới HÌNH VẼ trước khi xếp gạch, không phải `scale` của sprite:
      // xếp gạch tính theo kích thước hình, scale sprite sau đó làm đường nối lệch ra.
      const shape = al.scale && al.scale !== 1 ? grownShape(o.s, al.scale) : o.s;
      const art = o.s.c ? fitSprite(lent, shape) : tiledObstacle(lent, shape, vertical);
      if (al.tint !== undefined) {
        tintAll(art, al.tint);
      }
      if (al.glow === undefined) {
        return art;
      }
      // Quầng sáng vẽ DƯỚI vật, nở ra 18%: thạch bi "phù văn phát sáng" (§8.2).
      const wrap = new Container();
      const halo = new Graphics();
      drawShape(halo, grownShape(o.s, 1.18)).fill({ color: al.glow, alpha: 0.3 });
      wrap.addChild(halo, art);
      return wrap;
    }
    // Không art, không alias: ô màu theo loại. Nhìn là biết chưa xong — đúng ý placeholder.
    const g = new Graphics();
    drawShape(g, o.s).fill(OBSTACLE_COLOR[o.k] ?? 0x5b6b5b).stroke({ color: 0x000000, width: 6 });
    return g;
  }
}
