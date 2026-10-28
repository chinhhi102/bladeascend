import { Container, Graphics, Sprite, Texture } from 'pixi.js';

import { assets } from './assets';
import { hslToHex } from './palette';

/**
 * Thần thú Tiên giới (spec §3.5): mỗi cảnh giới Tiên một con, dựng bằng **chuỗi đốt** chứ
 * không vẽ từng frame — 1 đầu + 12–20 đốt thân + 1 đuôi, đầu chạy theo quỹ đạo, các đốt sau
 * đi theo vị trí đầu ở những thời điểm trước đó (như con rắn). Đổi màu bằng tint.
 *
 * Chiều sâu: đốt nào đang ở nửa SAU vòng xoắn (`sin θ < 0`) vẽ dưới nhân vật, nửa trước vẽ
 * trên → trông như rồng quấn quanh người thật. Mỗi đốt có sẵn hai bản (một ở lớp sau, một ở
 * lớp trước) và chỉ bật/tắt `visible`, rẻ hơn là chuyển cha mỗi khung hình.
 */

export type BeastMode = 'breakthrough' | 'ambient' | 'shatter';

export interface BeastSpec {
  /** thư mục art trong assets/beasts (ART_PROMPTS §10.1) */
  art: 'dragon' | 'crane' | 'phoenix';
  /**
   * Cách ghép đốt. `serpent` (rồng): 1 đầu + nhiều đốt tròn nối được + 1 đuôi.
   * `bird` (hạc, phượng): chỉ có **một** thân, phần còn lại là đuôi thưa dần — ảnh thân của
   * loài chim là nguyên con chim nên lặp 14 lần sẽ thành 14 con chim xếp hàng.
   */
  chain: 'serpent' | 'bird';
  color: number;
  /** hạt vệt theo sau */
  trail: 'cloud' | 'spark' | 'glow' | 'feather';
  /** Đạo Tổ: song long âm dương, một trắng một đen, viền cầu vồng */
  twin?: boolean;
}

/** Bảng thần thú theo spec §3.5, thứ tự khớp cultivation.tien_realms. */
export const TIEN_BEASTS: BeastSpec[] = [
  { art: 'crane', chain: 'bird', color: 0xe8f1ff, trail: 'cloud' }, // Chân Tiên — Tiên hạc
  { art: 'dragon', chain: 'serpent', color: 0xffd60a, trail: 'spark' }, // Kim Tiên — Kim long
  { art: 'dragon', chain: 'serpent', color: 0x4ce0b3, trail: 'glow' }, // Thái Ất Ngọc Tiên — Ngọc long
  { art: 'phoenix', chain: 'bird', color: 0xb15cff, trail: 'feather' }, // Đại La Kim Tiên — Tử phượng
  { art: 'dragon', chain: 'serpent', color: 0xffffff, trail: 'cloud', twin: true }, // Đạo Tổ — Song long âm dương
];

// Thời lượng từng kiểu xuất hiện (spec §3.5).
export const BEAST_SECS: Record<BeastMode, number> = { breakthrough: 2.5, ambient: 1.5, shatter: 0.8 };

const SEGMENTS = 16; // 1 đầu + 14 đốt thân + 1 đuôi, nằm trong khoảng 12–20 của spec
const SEG_GAP = 0.035; // đốt sau trễ hơn đốt trước bao nhiêu trên quỹ đạo
const FADE_SECS = 0.4; // spec: mờ dần ở 0.4s cuối
const TRAIL_SECS = 0.5;
const MAX_TRAIL = 36;
const DEPTH_SQUASH = 0.35; // ép dẹt quỹ đạo theo góc nhìn 3/4 (spec §3.5)

let placeholders: { head: Texture; body: Texture; tail: Texture } | null = null;

/** placeholderParts: đầu = tam giác, đốt = hình tròn (spec §3.5), trắng để tint theo màu. */
function placeholderParts(): { head: Texture; body: Texture; tail: Texture } {
  if (placeholders) {
    return placeholders;
  }
  const draw = (paint: (ctx: CanvasRenderingContext2D) => void): Texture => {
    const cv = document.createElement('canvas');
    cv.width = 64;
    cv.height = 64;
    const ctx = cv.getContext('2d') as CanvasRenderingContext2D;
    ctx.fillStyle = '#ffffff';
    ctx.strokeStyle = '#111111';
    ctx.lineWidth = 4;
    paint(ctx);
    ctx.fill();
    ctx.stroke();
    return Texture.from(cv);
  };
  const tri = (pts: number[]) => (ctx: CanvasRenderingContext2D): void => {
    ctx.beginPath();
    ctx.moveTo(pts[0], pts[1]);
    ctx.lineTo(pts[2], pts[3]);
    ctx.lineTo(pts[4], pts[5]);
    ctx.closePath();
  };
  placeholders = {
    // Đầu nhọn hướng sang PHẢI (0 rad) để khớp với rotation tính từ hướng bay.
    head: draw(tri([60, 32, 8, 6, 8, 58])),
    body: draw((ctx) => {
      ctx.beginPath();
      ctx.arc(32, 32, 26, 0, Math.PI * 2);
    }),
    tail: draw(tri([8, 32, 58, 10, 58, 54])),
  };
  return placeholders;
}

/** partAt: đốt thứ i dùng bộ phận nào — rồng nhiều đốt thân, chim chỉ một thân rồi tới đuôi. */
function partAt(spec: BeastSpec, i: number): 'head' | 'body' | 'tail' {
  if (i === 0) {
    return 'head';
  }
  if (spec.chain === 'bird') {
    return i === 1 ? 'body' : 'tail';
  }
  return i === SEGMENTS - 1 ? 'tail' : 'body';
}

/** partTexture: dùng art thật nếu đã có, không thì placeholder. */
function partTexture(spec: BeastSpec, part: 'head' | 'body' | 'tail'): Texture {
  return assets.texture(`beasts/${spec.art}_${part}.png`) ?? placeholderParts()[part];
}

interface Seg {
  back: Sprite;
  front: Sprite;
  /** 0 = đầu, 1 = đuôi */
  k: number;
}

interface Mote {
  node: Container;
  born: number;
  vx: number;
  vy: number;
}

export class Beast {
  readonly back = new Container(); // nửa sau vòng xoắn: vẽ DƯỚI nhân vật
  readonly front = new Container(); // nửa trước: vẽ TRÊN nhân vật
  readonly mode: BeastMode;
  private readonly spec: BeastSpec;
  private readonly segs: Seg[] = [];
  private readonly motes: Mote[] = [];
  private readonly born: number;
  private readonly dur: number;
  private readonly bodyR: number;
  private readonly charH: number;
  private readonly size: number;
  /** điểm vỡ cố định cho kiểu shatter */
  private readonly shards: { a: number; vy: number }[] = [];

  constructor(spec: BeastSpec, mode: BeastMode, bodyR: number, charH: number, now: number) {
    this.spec = spec;
    this.mode = mode;
    this.born = now;
    this.dur = BEAST_SECS[mode];
    this.bodyR = bodyR;
    this.charH = charH;
    this.size = (mode === 'ambient' ? 0.6 : 1) * bodyR * 0.85; // spec: bản thường trực 60% cỡ
    for (let i = 0; i < SEGMENTS; i++) {
      const tex = partTexture(spec, partAt(spec, i));
      const mk = (): Sprite => {
        const sp = new Sprite(tex);
        sp.anchor.set(0.5);
        return sp;
      };
      const seg: Seg = { back: mk(), front: mk(), k: i / (SEGMENTS - 1) };
      this.back.addChild(seg.back);
      this.front.addChild(seg.front);
      this.segs.push(seg);
      this.shards.push({ a: (i / SEGMENTS) * Math.PI * 2, vy: 60 + (i % 5) * 40 });
    }
  }

  /** done: hiệu ứng đã hết hạn, gọi destroy được. */
  done(now: number): boolean {
    return now - this.born >= this.dur;
  }

  /**
   * update đặt lại toàn bộ đốt quanh nhân vật ở (cx, cy).
   * `feetY` là chân nhân vật trong toạ độ thế giới.
   */
  update(cx: number, feetY: number, now: number): void {
    const t = Math.min(1, (now - this.born) / this.dur);
    const fade = Math.min(1, (this.dur - (now - this.born)) / FADE_SECS);
    const alpha = Math.max(0, Math.min(1, fade));
    this.back.position.set(cx, 0);
    this.front.position.set(cx, 0);
    if (this.mode === 'shatter') {
      this.updateShatter(feetY, t, alpha);
      return;
    }
    for (let i = 0; i < this.segs.length; i++) {
      const seg = this.segs[i];
      const u = Math.max(0, t - i * SEG_GAP); // đốt sau đi theo vị trí đầu lúc trước
      const p = this.pathAt(u, feetY);
      const prev = this.pathAt(Math.max(0, u - 0.02), feetY);
      // Đốt nhỏ dần về đuôi (spec §3.5). `s` là bề ngang mong muốn tính bằng đơn vị thế
      // giới; place() mới chia cho bề ngang thật của texture — placeholder 64px và art thật
      // 512px phải ra cùng một cỡ trên màn hình.
      const s = this.size * (1 - 0.45 * seg.k) * 2;
      const rot = Math.atan2(p.y - prev.y, p.x - prev.x);
      const front = p.depth >= 0;
      this.place(seg.back, p.x, p.y, s, rot, alpha, !front);
      this.place(seg.front, p.x, p.y, s, rot, alpha, front);
      if (i === 0 && this.motes.length < MAX_TRAIL && now - this.born > 0.05) {
        this.emitMote(p.x, p.y, now);
      }
    }
    this.updateMotes(now);
  }

  /** updateShatter: thần thú hiện ra rồi gãy vụn thành mảnh sáng rơi xuống (spec §3.5). */
  private updateShatter(feetY: number, t: number, alpha: number): void {
    const cy = feetY - this.charH * 0.7;
    for (let i = 0; i < this.segs.length; i++) {
      const seg = this.segs[i];
      const sh = this.shards[i];
      const r = this.bodyR * 1.1;
      const x = Math.cos(sh.a) * r * (1 + t * 0.8);
      const y = cy + Math.sin(sh.a) * r * DEPTH_SQUASH + sh.vy * t * t * 3;
      const s = this.size * (1 - 0.45 * seg.k) * (1 - 0.4 * t) * 2;
      const front = Math.sin(sh.a) >= 0;
      this.place(seg.back, x, y, s, sh.a + t * 6, alpha, !front);
      this.place(seg.front, x, y, s, sh.a + t * 6, alpha, front);
    }
  }

  /** place đặt một đốt; `width` là bề ngang mong muốn tính bằng đơn vị thế giới. */
  private place(sp: Sprite, x: number, y: number, width: number, rot: number, alpha: number, visible: boolean): void {
    sp.visible = visible;
    if (!visible) {
      return;
    }
    sp.position.set(x, y);
    sp.scale.set(width / Math.max(sp.texture.width, 1));
    sp.rotation = rot;
    sp.alpha = alpha;
    sp.tint = this.tintOf(x);
  }

  /** Đạo Tổ: song long một trắng một đen + viền cầu vồng → đổi tint theo vị trí. */
  private tintOf(x: number): number {
    if (!this.spec.twin) {
      return this.spec.color;
    }
    return x >= 0 ? 0xffffff : 0x262626;
  }

  /**
   * pathAt là quỹ đạo xoắn ốc của spec §3.5:
   * `x = R(u)·cos θ(u)`, `y = feet − H(u) + R(u)·sin θ(u) × 0.35`.
   * Trả thêm `depth = sin θ` để biết đốt đang ở nửa trước hay nửa sau.
   */
  private pathAt(u: number, feetY: number): { x: number; y: number; depth: number } {
    const turns = this.mode === 'breakthrough' ? 1.5 : 1; // đột phá 1.5 vòng, thường trực 1 vòng
    const theta = u * turns * Math.PI * 2 - Math.PI / 2; // bắt đầu từ phía sau lưng
    let h: number;
    let r: number;
    if (this.mode === 'breakthrough') {
      // Chui lên từ chân, quấn quanh người rồi bay vút lên: H tăng gia tốc, R bung ra ở cuối.
      const flare = Math.max(0, (u - 0.8) / 0.2);
      h = 3 * this.charH * u * u;
      r = this.bodyR * (0.6 + 0.4 * u + 2.5 * flare * flare);
    } else {
      // Thường trực: bay 1 vòng ngang ngực rồi chui vào hào quang.
      const tuck = Math.max(0, (u - 0.75) / 0.25);
      h = this.charH * (0.55 + 0.25 * Math.sin(u * Math.PI));
      r = this.bodyR * 1.25 * (1 - 0.8 * tuck);
    }
    return {
      x: Math.cos(theta) * r,
      y: feetY - h + Math.sin(theta) * r * DEPTH_SQUASH,
      depth: Math.sin(theta),
    };
  }

  private emitMote(x: number, y: number, now: number): void {
    const c = this.spec.twin ? hslToHex((now * 120) % 360, 0.9, 0.65) : this.spec.color;
    const r = this.size * 0.22;
    const node = this.moteNode(c, r);
    node.position.set(x, y);
    this.front.addChild(node);
    this.motes.push({ node, born: now, vx: (Math.random() - 0.5) * 60, vy: -20 - Math.random() * 60 });
  }

  /** moteNode: ảnh hạt thật nếu đã vẽ (ART_PROMPTS §10.1), không thì hình đơn giản. */
  private moteNode(color: number, r: number): Container {
    const tex = assets.texture(`beasts/mote_${this.spec.trail}.png`);
    if (tex) {
      const sp = new Sprite(tex);
      sp.anchor.set(0.5);
      sp.tint = color;
      sp.scale.set((r * 2.4) / Math.max(tex.width, 1));
      return sp;
    }
    const g = new Graphics();
    if (this.spec.trail === 'cloud') {
      g.ellipse(0, 0, r * 1.4, r * 0.7).fill({ color, alpha: 0.8 });
    } else if (this.spec.trail === 'feather') {
      g.ellipse(0, 0, r * 0.4, r * 1.5).fill({ color, alpha: 0.85 });
    } else {
      g.circle(0, 0, r).fill({ color, alpha: 0.9 });
    }
    return g;
  }

  private updateMotes(now: number): void {
    for (let i = this.motes.length - 1; i >= 0; i--) {
      const m = this.motes[i];
      const age = now - m.born;
      if (age >= TRAIL_SECS) {
        m.node.destroy();
        this.motes.splice(i, 1);
        continue;
      }
      m.node.x += m.vx * 0.016;
      m.node.y += m.vy * 0.016;
      m.node.alpha = 1 - age / TRAIL_SECS;
    }
  }

  destroy(): void {
    this.back.destroy({ children: true });
    this.front.destroy({ children: true });
  }
}
