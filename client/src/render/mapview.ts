import { AnimatedSprite, ColorMatrixFilter, Container, Graphics, Sprite, TilingSprite, type Texture } from 'pixi.js';

import type { ClientWorld } from '../net/state';
import type { GameMap, MapObstacle, Shape } from '../net/protocol';
import { assets } from './assets';
import { GATE_COLORS } from './palette';

const BIOME_TILE: Record<string, string> = { grass: 'grass', market: 'concrete', concrete: 'concrete', dirt: 'dirt', rock: 'dirt' };
const BIOME_COLOR: Record<string, number> = { grass: 0x7cc46a, concrete: 0xa9a49b, dirt: 0xc9a46a };
const OBSTACLE_COLOR: Record<string, number> = { fence: 0xc8a165, stall: 0xe76f51, cart: 0x9c6644, brick: 0xb5523b, boulder: 0x8d8d8d, stone_wall: 0x5b6b5b };
const TILE_WORLD = 400; // one 512px ground tile covers 400 world units
const ART_FILL = 0.9; // object art fills ~90% of its canvas

function bounds(s: Shape): { x: number; y: number; w: number; h: number } {
  if (s.c) {
    const r = s.r ?? 0;
    return { x: s.x - r, y: s.y - r, w: r * 2, h: r * 2 };
  }
  return { x: s.x, y: s.y, w: s.w ?? 0, h: s.h ?? 0 };
}

function drawShape(g: Graphics, s: Shape): Graphics {
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
  // Số mảnh đủ phủ kín: bước = đúng chiều dài hiển thị của mảnh, làm tròn LÊN.
  const step = size * ART_FILL;
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

type ObstacleLook = 'intact' | 'cracked' | 'nearly_broken' | 'rubble';

interface ObstacleNode {
  o: MapObstacle;
  node: Container;
  look: ObstacleLook | '';
  hpDrawn: number;
  warnUntil: number;
  box: { x: number; y: number; w: number; h: number }; // drawn bounds, for see-through
}

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

const HP_FULL = 0x4caf50;
const HP_MID = 0xffd23f;
const HP_LOW = 0xe63946;

function mixColor(a: number, b: number, t: number): number {
  const ch = (shift: number): number => Math.round(((a >> shift) & 255) + (((b >> shift) & 255) - ((a >> shift) & 255)) * t);
  return (ch(16) << 16) | (ch(8) << 8) | ch(0);
}

/** hpColor: green at full HP → yellow at half → red when nearly broken (placeholder obstacles). */
function hpColor(frac: number): number {
  return frac > 0.5 ? mixColor(HP_MID, HP_FULL, (frac - 0.5) * 2) : mixColor(HP_LOW, HP_MID, frac * 2);
}

/**
 * MapView draws the static map. Obstacle nodes are handed to the scene's entity layer so they
 * y-sort with characters; they change look with their HP (spec §8.2).
 */
export class MapView {
  readonly ground = new Container();
  readonly decor = new Container();
  readonly under = new Container(); // zones under entities (mud, wind) + gates
  readonly over = new Container(); // bush canopy above entities
  private readonly obstacleNodes = new Map<number, ObstacleNode>();
  private readonly bushes: { id: number; node: Container }[] = [];
  private readonly gates = new Map<number, { node: Container; warnUntil: number }>();

  constructor(
    private readonly map: GameMap,
    private readonly gateR: number,
  ) {
    this.buildGround();
    this.buildZones();
    this.buildGates();
    this.buildDecor();
    for (const o of map.obstacles) {
      const n: ObstacleNode = { o, node: new Container(), look: '', hpDrawn: -1, warnUntil: 0, box: bounds(o.s) };
      this.obstacleNodes.set(o.id, n);
    }
    const filter = backgroundFilter();
    this.ground.filters = [filter];
    this.decor.filters = [filter];
  }

  /** obstacleLayer returns every obstacle node, to be added to the y-sorted entity layer. */
  obstacleLayer(): Container[] {
    return [...this.obstacleNodes.values()].map((n) => n.node);
  }

  private buildGround(): void {
    const size = this.map.size;
    const tex = assets.texture(`tiles/${BIOME_TILE[this.map.biome] ?? 'grass'}.png`);
    if (tex) {
      const t = new TilingSprite({ texture: tex, width: size, height: size });
      t.tileScale.set(TILE_WORLD / tex.width);
      this.ground.addChild(t);
    } else {
      const g = new Graphics().rect(0, 0, size, size).fill(BIOME_COLOR[BIOME_TILE[this.map.biome] ?? 'grass'] ?? 0x7cc46a);
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
      this.under.addChild(node);
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

  update(w: ClientWorld, now: number, myBush: number, me: Viewer | null, occlusionAlpha: number): void {
    for (const n of this.obstacleNodes.values()) {
      this.updateObstacle(n, w, now);
      if (me && n.look !== 'rubble' && this.covers(n, me)) {
        n.node.alpha *= occlusionAlpha;
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

  // covers: the obstacle stands in front of me (its base is lower on screen) and overlaps my sprite.
  private covers(n: ObstacleNode, me: Viewer): boolean {
    const b = n.box;
    return b.y + b.h > me.feet && b.x < me.x + me.halfW && b.x + b.w > me.x - me.halfW && b.y < me.feet && b.y + b.h > me.top;
  }

  private updateObstacle(n: ObstacleNode, w: ClientWorld, now: number): void {
    const broken = w.brokenObstacles.get(n.o.id) ?? false;
    const hp = w.obstacleHp.get(n.o.id) ?? n.o.hp;
    const frac = n.o.hp > 0 ? hp / n.o.hp : 1;
    let look: ObstacleLook = 'intact';
    if (broken) {
      look = 'rubble';
    } else if (frac < 0.33) {
      look = 'nearly_broken';
    } else if (frac <= 0.66) {
      look = 'cracked';
    }
    const hasArt = assets.texture(`obstacles/${n.o.k}/${look}.png`) !== null;
    if (look !== n.look || (!hasArt && hp !== n.hpDrawn)) {
      n.look = look;
      n.hpDrawn = hp;
      n.node.removeChildren().forEach((c) => c.destroy());
      n.node.addChild(this.obstacleArt(n.o, look, frac));
      const lb = n.node.getLocalBounds();
      n.box = { x: lb.x, y: lb.y, w: lb.width, h: lb.height };
      n.node.zIndex = broken ? -1e6 : bounds(n.o.s).y + bounds(n.o.s).h;
    }
    // Blink 3s before regrowing (spec §8.2).
    n.node.alpha = broken && now < n.warnUntil ? 0.5 + 0.5 * Math.sin(now * 20) : 1;
  }

  private obstacleArt(o: MapObstacle, look: ObstacleLook, frac: number): Container {
    const b = bounds(o.s);
    // Đoạn dọc cần art riêng: xoay sprite ngang 90° trông sai hẳn (bóng đổ, mặt đứng).
    // Chưa có ảnh thì dùng placeholder dọc, và `assets:report` liệt kê để sinh sau.
    const vertical = !o.s.c && b.h > b.w;
    const tex = assets.texture(`obstacles/${o.k}/${look}${vertical ? '_vertical' : ''}.png`)
      ?? (vertical ? null : assets.texture(`obstacles/${o.k}/${look}.png`));
    if (tex) {
      return o.s.c ? fitSprite(tex, o.s) : tiledObstacle(tex, o.s, vertical);
    }
    // Placeholder: indestructible walls keep their colour; breakable ones show HP as colour.
    const color = o.hp > 0 ? hpColor(frac) : OBSTACLE_COLOR[o.k] ?? 0x5b6b5b;
    const g = new Graphics();
    if (look === 'rubble') {
      return drawShape(g, o.s).fill({ color: 0x777777, alpha: 0.3 }).stroke({ color: 0x555555, width: 4, alpha: 0.5 });
    }
    drawShape(g, o.s).fill(color).stroke({ color: 0x000000, width: 6 });
    if (look !== 'intact') {
      const b = bounds(o.s);
      g.moveTo(b.x + b.w * 0.2, b.y + b.h * 0.1).lineTo(b.x + b.w * 0.45, b.y + b.h * 0.6).lineTo(b.x + b.w * 0.35, b.y + b.h * 0.9);
      if (look === 'nearly_broken') {
        g.moveTo(b.x + b.w * 0.8, b.y + b.h * 0.15).lineTo(b.x + b.w * 0.6, b.y + b.h * 0.55).lineTo(b.x + b.w * 0.75, b.y + b.h * 0.85);
      }
      g.stroke({ color: 0x000000, width: 5 });
    }
    return g;
  }
}
