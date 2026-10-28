import { Container, Graphics, Sprite, Texture } from 'pixi.js';

import { assets } from './assets';
import { hslToHex } from './palette';
import { particle } from './particles';
import { BEAST_SEGMENTS, quality, TRAIL_RATE } from './quality';

/**
 * Thần thú Tiên giới (spec §3.5): mỗi cảnh giới Tiên một con, dựng bằng **chuỗi đốt** chứ
 * không vẽ từng frame — 1 đầu + 12–20 đốt thân + 1 đuôi, đầu chạy theo quỹ đạo, các đốt sau
 * đi theo vị trí đầu ở những thời điểm trước đó (như con rắn). Đổi màu bằng tint.
 *
 * Chiều sâu: đốt nào đang ở nửa SAU vòng xoắn (`sin θ < 0`) vẽ dưới nhân vật, nửa trước vẽ
 * trên → trông như rồng quấn quanh người thật. Mỗi đốt có sẵn hai bản (một ở lớp sau, một ở
 * lớp trước) và chỉ bật/tắt `visible`, rẻ hơn là chuyển cha mỗi khung hình.
 */

export type BeastMode = 'breakthrough' | 'companion' | 'shatter';

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

// Thời lượng từng kiểu (spec §3.5). `companion` không hết hạn — spec v4.7 bỏ kiểu "mỗi 20s
// bay một vòng": sau màn đột phá thần thú ở lại bay quanh chủ suốt thời gian còn ở Tiên giới.
export const BEAST_SECS: Record<BeastMode, number> = { breakthrough: 2.5, companion: Infinity, shatter: 0.8 };

const SEG_FRAMES = 3; // đốt sau trễ bao nhiêu KHUNG HÌNH so với đốt trước
const FADE_SECS = 0.4; // spec: mờ dần ở 0.4s cuối
const TRAIL_SECS = 0.5;
const MAX_TRAIL = 24;
const DEPTH_SQUASH = 0.35; // ép dẹt quỹ đạo theo góc nhìn 3/4 (spec §3.5)
const ORBIT_SECS = 4; // quỹ đạo nền: 1 vòng / 4s
const BOB_SECS = 3; // độ cao nhấp nhô chu kỳ 3s
const CENTER_LERP = 0.1; // tâm quỹ đạo bám chủ có độ trễ → chạy thì thần thú bị kéo dài ra sau
const FAR_SEGMENTS = 5; // chủ nhỏ hơn 40 px trên màn hình: chỉ vẽ đầu + 4 đốt

/** Động tác ngẫu nhiên, mỗi 8–15s chọn một (spec §3.5 v4.7). */
const MOVES = ['spiral', 'figure8', 'flip', 'dash', 'coil'] as const;
type Move = (typeof MOVES)[number];

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
function partAt(spec: BeastSpec, i: number, n: number): 'head' | 'body' | 'tail' {
  if (i === 0) {
    return 'head';
  }
  if (spec.chain === 'bird') {
    return i === 1 ? 'body' : 'tail';
  }
  return i === n - 1 ? 'tail' : 'body';
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
  private bodyR: number;
  private charH: number;
  private size: number;
  /** Lệch pha quỹ đạo — Đạo Tổ có 2 con lệch nhau 180° (spec §3.5 v4.7). */
  readonly phase: number;
  /** Vòng đệm vị trí đầu các khung trước; đốt sau đọc lại. Cấp phát một lần, không mỗi frame. */
  private readonly histX: Float64Array;
  private readonly histY: Float64Array;
  private histAt = 0;
  private histFilled = false;
  /** Tâm quỹ đạo, bám chủ có độ trễ. */
  private cx = 0;
  private cy = 0;
  private centred = false;
  private move: Move | null = null;
  private moveFrom = 0;
  private moveSecs = 0;
  private nextMoveAt = 0;
  private dir = 1;
  /** Vệt hạt chỉ bật cho vài con gần nhất; mờ theo chủ khi Ẩn Thân / trong bụi. */
  trailOn = true;
  ownerAlpha = 1;
  /** Chủ đang Thiêu Đốt Tinh Huyết: bay sát người, vệt đỏ (spec §3.5 v4.7 mục phản ứng). */
  burning = false;
  /** Lóe sáng tới thời điểm này (nhặt Tiên khí, khiên đỡ đòn). */
  private flashUntil = 0;
  /** điểm vỡ cố định cho kiểu shatter */
  private flashing = false;
  private readonly shards: { a: number; vy: number }[] = [];

  constructor(spec: BeastSpec, mode: BeastMode, bodyR: number, charH: number, now: number, phase = 0) {
    this.spec = spec;
    this.mode = mode;
    this.phase = phase;
    this.born = now;
    this.dur = BEAST_SECS[mode];
    this.bodyR = bodyR;
    this.charH = charH;
    this.size = (mode === 'companion' ? 0.6 : 1) * bodyR * 0.85; // spec: bản thường trực 60% cỡ
    this.nextMoveAt = now + 8 + Math.random() * 7;
    const n = Math.max(3, BEAST_SEGMENTS[quality()]);
    this.histX = new Float64Array(n * SEG_FRAMES + 2);
    this.histY = new Float64Array(n * SEG_FRAMES + 2);
    for (let i = 0; i < n; i++) {
      const tex = partTexture(spec, partAt(spec, i, n));
      const mk = (): Sprite => {
        const sp = new Sprite(tex);
        sp.anchor.set(0.5);
        sp.blendMode = 'add'; // spec §3.6 mục 31: đốt thân blend cộng cho sáng rực
        return sp;
      };
      const seg: Seg = { back: mk(), front: mk(), k: i / (n - 1) };
      this.back.addChild(seg.back);
      this.front.addChild(seg.front);
      this.segs.push(seg);
      this.shards.push({ a: (i / n) * Math.PI * 2, vy: 60 + (i % 5) * 40 });
    }
  }

  /** done: hết hạn. Bản `companion` không bao giờ hết — scene gỡ khi chủ rời Tiên giới. */
  done(now: number): boolean {
    return now - this.born >= this.dur;
  }

  /**
   * flash cho thần thú lóe sáng và lượn nhanh một vòng — gọi khi chủ nhặt Tiên khí hoặc
   * khiên đỡ được một đòn (spec §3.5 v4.7).
   */
  flash(now: number): void {
    this.flashUntil = now + 0.9;
    // Chen luôn một động tác để nó "mừng" thấy rõ chứ không chỉ sáng lên.
    if (!this.move) {
      this.move = 'spiral';
      this.moveFrom = now;
      this.moveSecs = 0.9;
    }
  }

  /** resize cập nhật cỡ khi chủ lên tầng (scale đổi). */
  resize(bodyR: number, charH: number): void {
    this.bodyR = bodyR;
    this.charH = charH;
    this.size = (this.mode === 'companion' ? 0.6 : 1) * bodyR * 0.85;
  }

  /**
   * update đặt lại toàn bộ đốt. `cx`/`feetY` là chân nhân vật trong toạ độ thế giới;
   * `near` = chủ có đủ to trên màn hình không (spec §3.5: nhỏ hơn 40 px thì rút gọn đốt).
   */
  update(cx: number, feetY: number, now: number, near = true): void {
    const alpha = this.ownerAlpha * this.fadeAt(now);
    this.flashing = now < this.flashUntil;
    if (this.mode === 'shatter') {
      this.updateShatter(cx, feetY, Math.min(1, (now - this.born) / this.dur), alpha);
      return;
    }
    const head = this.mode === 'breakthrough' ? this.spiralHead(cx, feetY, now) : this.companionHead(cx, feetY, now);
    this.pushHistory(head.x, head.y);
    const shown = near ? this.segs.length : Math.min(FAR_SEGMENTS, this.segs.length);
    const waist = feetY - this.charH * 0.5;
    for (let i = 0; i < this.segs.length; i++) {
      const seg = this.segs[i];
      if (i >= shown) {
        seg.back.visible = false;
        seg.front.visible = false;
        continue;
      }
      const px = this.histAtX(i * SEG_FRAMES);
      const py = this.histAtY(i * SEG_FRAMES);
      const qx = this.histAtX(i * SEG_FRAMES + 1);
      const qy = this.histAtY(i * SEG_FRAMES + 1);
      // Đốt nhỏ dần về đuôi; `w` là bề ngang mong muốn tính bằng đơn vị thế giới, place()
      // mới chia cho bề ngang thật của texture (placeholder 64px vs art thật 512px).
      const w = this.size * (1 - 0.45 * seg.k) * 2;
      const rot = Math.atan2(py - qy, px - qx);
      // Đốt nằm thấp hơn ngang hông thì đang ở nửa TRƯỚC quỹ đạo → vẽ trên nhân vật.
      const front = i === 0 ? head.depth >= 0 : py >= waist;
      this.place(seg.back, px, py, w, rot, alpha, !front);
      this.place(seg.front, px, py, w, rot, alpha, front);
    }
    const rate = TRAIL_RATE[quality()];
    if (this.trailOn && rate > 0 && this.motes.length < MAX_TRAIL * rate) {
      this.emitMote(head.x, head.y, now);
    }
    this.updateMotes(now);
  }

  private fadeAt(now: number): number {
    if (this.dur === Infinity) {
      return Math.min(1, (now - this.born) / 0.3); // hiện ra mượt rồi ở lại
    }
    return Math.max(0, Math.min(1, (this.dur - (now - this.born)) / FADE_SECS));
  }

  /**
   * companionHead là quỹ đạo nền của bản thường trực (spec §3.5 v4.7): vòng elip quanh thân
   * bán kính 1.3–1.8 × chiều cao nhân vật, ép dẹt theo góc nhìn, 1 vòng / 4s, độ cao nhấp
   * nhô chu kỳ 3s, tâm bám chủ có độ trễ nên lúc chạy thần thú bị kéo dài ra sau như dải lụa.
   */
  private companionHead(cx: number, feetY: number, now: number): { x: number; y: number; depth: number } {
    if (!this.centred) {
      this.cx = cx;
      this.cy = feetY;
      this.centred = true;
    }
    this.cx += (cx - this.cx) * CENTER_LERP;
    this.cy += (feetY - this.cy) * CENTER_LERP;
    this.pickMove(now);

    // Bao hình sin để động tác vào và ra đều mượt, không giật về quỹ đạo nền.
    const env = this.move ? Math.sin(Math.min(1, (now - this.moveFrom) / this.moveSecs) * Math.PI) : 0;
    let theta = this.dir * now * ((2 * Math.PI) / ORBIT_SECS) + this.phase;
    // Chủ đang Thiêu Đốt thì bay sát người hẳn, như bị kéo theo luồng khí.
    let r = this.charH * (this.burning ? 0.75 : 1.3 + 0.25 * Math.sin(now * 0.37));
    let h = this.charH * (0.9 + 0.12 * Math.sin((now * 2 * Math.PI) / BOB_SECS));
    switch (this.move) {
      case 'spiral': // xoắn ốc vút lên rồi bổ nhào xuống
        h += this.charH * 1.8 * env;
        r *= 1 - 0.4 * env;
        theta += env * 4;
        break;
      case 'figure8': // lượn hình số 8 quanh người
        theta += Math.sin(theta * 2) * 0.9 * env;
        break;
      case 'dash': // lao ra xa một nhịp rồi vòng về
        r *= 1 + 1.6 * env;
        break;
      case 'coil': // cuộn tròn quanh đầu chủ
        h += this.charH * 0.8 * env;
        r *= 1 - 0.65 * env;
        theta += env * 6;
        break;
      default: // 'flip' chỉ đổi this.dir, quỹ đạo giữ nguyên
    }
    return { x: this.cx + Math.cos(theta) * r, y: this.cy - h + Math.sin(theta) * r * DEPTH_SQUASH, depth: Math.sin(theta) };
  }

  /** pickMove chọn động tác mới mỗi 8–15s, mỗi động tác 1.5–3s (spec §3.5 v4.7). */
  private pickMove(now: number): void {
    if (this.move && now - this.moveFrom >= this.moveSecs) {
      this.move = null;
      this.nextMoveAt = now + 8 + Math.random() * 7;
      return;
    }
    if (this.move || now < this.nextMoveAt) {
      return;
    }
    this.move = MOVES[Math.floor(Math.random() * MOVES.length)];
    this.moveFrom = now;
    this.moveSecs = 1.5 + Math.random() * 1.5;
    if (this.move === 'flip') {
      this.dir = -this.dir;
    }
  }

  /** spiralHead là màn đột phá: chui lên từ chân, quấn 1.5 vòng rồi bay vút lên (spec §3.5). */
  private spiralHead(cx: number, feetY: number, now: number): { x: number; y: number; depth: number } {
    const u = Math.min(1, (now - this.born) / this.dur);
    const theta = u * 1.5 * Math.PI * 2 - Math.PI / 2;
    const flare = Math.max(0, (u - 0.8) / 0.2);
    const h = 3 * this.charH * u * u;
    const r = this.bodyR * (0.6 + 0.4 * u + 2.5 * flare * flare);
    return { x: cx + Math.cos(theta) * r, y: feetY - h + Math.sin(theta) * r * DEPTH_SQUASH, depth: Math.sin(theta) };
  }

  private pushHistory(x: number, y: number): void {
    if (!this.histFilled) {
      this.histX.fill(x);
      this.histY.fill(y);
      this.histFilled = true;
    }
    this.histAt = (this.histAt + 1) % this.histX.length;
    this.histX[this.histAt] = x;
    this.histY[this.histAt] = y;
  }

  private histIndex(back: number): number {
    const n = this.histX.length;
    return (this.histAt - Math.min(back, n - 1) + n) % n;
  }

  private histAtX(back: number): number {
    return this.histX[this.histIndex(back)];
  }

  private histAtY(back: number): number {
    return this.histY[this.histIndex(back)];
  }

  /** updateShatter: thần thú hiện ra rồi gãy vụn thành mảnh sáng rơi xuống (spec §3.5). */
  private updateShatter(cx: number, feetY: number, t: number, alpha: number): void {
    const cy = feetY - this.charH * 0.7;
    for (let i = 0; i < this.segs.length; i++) {
      const seg = this.segs[i];
      const sh = this.shards[i];
      const r = this.bodyR * 1.1;
      const x = cx + Math.cos(sh.a) * r * (1 + t * 0.8);
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
    sp.tint = this.tintOf();
  }

  /** Đạo Tổ: song long một trắng một đen + viền cầu vồng → đổi tint theo vị trí. */
  private tintOf(): number {
    if (this.burning) {
      return 0xff5a3c; // Thiêu Đốt Tinh Huyết: vệt đỏ
    }
    if (this.flashing) {
      return 0xffffff; // lóe trắng khi chủ nhặt Tiên khí / khiên đỡ đòn
    }
    if (!this.spec.twin) {
      return this.spec.color;
    }
    // Đạo Tổ: hai con lệch pha 180°, con pha 0 màu trắng, con kia màu đen.
    return this.phase === 0 ? 0xffffff : 0x303030;
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
    // Ưu tiên ảnh riêng của thần thú; không có thì dùng bộ hạt CC0 dùng chung (spec §3.6).
    const shared: Record<string, 'smoke' | 'spark' | 'light' | 'star'> =
      { cloud: 'smoke', spark: 'spark', glow: 'light', feather: 'star' };
    const tex = assets.texture(`beasts/mote_${this.spec.trail}.png`) ?? particle(shared[this.spec.trail] ?? 'light');
    if (tex) {
      const sp = new Sprite(tex);
      sp.anchor.set(0.5);
      sp.blendMode = 'add';
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
