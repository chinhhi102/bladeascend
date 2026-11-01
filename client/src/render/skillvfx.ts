import { Container, Graphics } from 'pixi.js';

import { quality } from './quality';
import { note } from './statusview';

/**
 * SkillVfx: hiệu ứng NGẮN của kỹ năng (spec v6.36 §6.5) — sóng lan, kiếm bay, trân châu,
 * sợi mì, nốt nhạc, khung sạp nhấp nháy. Khác `SkillFx` (chỉ báo phạm vi đúng công thức U)
 * và `StatusView` (dấu hiệu kéo dài theo trạng thái).
 *
 * Mỗi hiệu ứng là MỘT `Graphics` vẽ lại theo tiến độ k = 0..1 rồi huỷ. Chất lượng Thấp: bỏ
 * hạt trang trí (nốt nhạc, trân châu bay/rơi, hơi nóng), giữ sóng, sợi mì, kiếm bay (ít hơn).
 */

interface Fx {
  g: Graphics;
  born: number;
  dur: number;
  draw: (g: Graphics, k: number) => void;
}

/** Trần hiệu ứng cùng lúc — 10 người cùng xả kỹ năng giữa chợ không được làm tụt khung hình. */
const MAX_FX = 48;

export interface Path {
  x0: number;
  y0: number;
  x1: number;
  y1: number;
}

export class SkillVfx {
  readonly layer = new Container();
  private readonly live: Fx[] = [];

  constructor() {
    this.layer.eventMode = 'none';
  }

  /** wave: vòng sóng lan từ tâm ra bán kính r (Loa Phường, sóng Rụt Cổ). */
  wave(x: number, y: number, r: number, color: number, now: number, dur = 0.4): void {
    this.add(x, y, now, dur, (g, k) => {
      const e = 1 - (1 - k) * (1 - k);
      g.circle(0, 0, r * e).stroke({ color, width: Math.max(4, r * 0.05 * (1 - k)), alpha: 1 - k });
      g.circle(0, 0, r * e * 0.8).stroke({ color: 0xffffff, width: Math.max(2, r * 0.02), alpha: (1 - k) * 0.6 });
    });
  }

  /**
   * swords: `n` thanh kiếm bay theo đường `p` lần lượt (Bánh Bao vào vòng, Hút Mì dọc sợi
   * mì, Trộm Kiếm vệt tím về võng). `arc` > 0 cho đường cong vồng lên.
   */
  swords(p: Path, n: number, color: number, now: number, arc = 0): void {
    const count = Math.min(n, quality() === 'low' ? 3 : 8);
    const len = Math.hypot(p.x1 - p.x0, p.y1 - p.y0);
    const size = Math.max(14, Math.min(60, len * 0.08));
    for (let i = 0; i < count; i++) {
      this.add(0, 0, now + i * 0.06, 0.45, (g, k) => {
        const at = along(p, k, arc * len);
        const back = along(p, Math.max(0, k - 0.12), arc * len);
        const ang = Math.atan2(at.y - back.y, at.x - back.x);
        g.moveTo(back.x, back.y).lineTo(at.x, at.y).stroke({ color, width: size * 0.35, alpha: 0.35 * (1 - k * 0.5) });
        blade(g, at.x, at.y, ang, size, color);
      });
    }
  }

  /** streak: vệt bụi lướt dọc đường p (đàn em của Đại Ca chạy từ ngoài vào). */
  streak(p: Path, color: number, now: number): void {
    const len = Math.hypot(p.x1 - p.x0, p.y1 - p.y0);
    this.add(0, 0, now, 0.4, (g, k) => {
      const head = along(p, Math.min(1, k * 1.6), 0);
      const tail = along(p, Math.max(0, k * 1.6 - 0.5), 0);
      g.moveTo(tail.x, tail.y).lineTo(head.x, head.y).stroke({ color, width: Math.max(8, len * 0.05), alpha: 0.6 * (1 - k), cap: 'round' });
    });
  }

  /** notes: nốt nhạc bay ra quanh vòng hát ru (Ông Ba). */
  notes(x: number, y: number, r: number, color: number, now: number): void {
    if (quality() === 'low') {
      return;
    }
    for (let i = 0; i < 6; i++) {
      const a = (i * Math.PI * 2) / 6 + 0.4;
      this.add(x, y, now + i * 0.04, 0.6, (g, k) => {
        const d = r * (0.3 + 0.7 * k);
        note(g, Math.cos(a) * d, Math.sin(a) * d - r * 0.15 * k, Math.max(8, r * 0.05), color, 1 - k);
      });
    }
  }

  /** pearls: hạt trân châu bắn ra trong nón (Thằng Mập). */
  pearls(x: number, y: number, facing: number, r: number, deg: number, now: number): void {
    if (quality() === 'low') {
      return;
    }
    const half = (deg * Math.PI) / 360;
    this.add(x, y, now, 0.35, (g, k) => {
      for (let i = 0; i < 14; i++) {
        const a = facing - half + ((i * 7919) % 14 / 13) * half * 2;
        const d = r * k * (0.55 + 0.45 * (((i * 31) % 10) / 10));
        g.circle(Math.cos(a) * d, Math.sin(a) * d, Math.max(5, r * 0.025)).fill({ color: 0x2b1a10, alpha: 1 - k * 0.4 });
      }
    });
  }

  /** fall: hạt rơi lả tả rồi mờ (trân châu rụng khỏi vòng kiếm khi hết mù). */
  fall(x: number, y: number, r: number, color: number, now: number): void {
    if (quality() === 'low') {
      return;
    }
    this.add(x, y, now, 0.4, (g, k) => {
      for (let i = 0; i < 10; i++) {
        const a = (i * Math.PI * 2) / 10;
        const px = Math.cos(a) * r * (0.6 + 0.1 * k);
        const py = Math.sin(a) * r * 0.6 + r * 0.9 * k * k;
        g.circle(px, py, Math.max(4, r * 0.05)).fill({ color, alpha: 1 - k });
      }
    });
  }

  /** puff: cụm khói nhỏ tan (xì hơi, phanh gấp, sạp sụp). */
  puff(x: number, y: number, r: number, color: number, now: number): void {
    this.add(x, y, now, 0.35, (g, k) => {
      const n = quality() === 'low' ? 3 : 6;
      for (let i = 0; i < n; i++) {
        const a = (i * Math.PI * 2) / n;
        g.circle(Math.cos(a) * r * 0.6 * k, Math.sin(a) * r * 0.3 * k - r * 0.2 * k, r * (0.25 + 0.2 * k)).fill({ color, alpha: 0.55 * (1 - k) });
      }
    });
  }

  /** steam: hơi nóng bốc lên (A Hớn ăn bánh bao) — kèm cái bánh trên tay. */
  bun(x: number, y: number, r: number, now: number): void {
    this.add(x, y, now, 0.9, (g, k) => {
      const a = k < 0.8 ? 1 : (1 - k) / 0.2;
      g.ellipse(0, 0, r * 0.45, r * 0.32).fill({ color: 0xfff4dc, alpha: a }).stroke({ color: 0x8a6a3a, width: r * 0.05, alpha: a });
      g.moveTo(-r * 0.15, -r * 0.12).quadraticCurveTo(0, -r * 0.3, r * 0.15, -r * 0.12).stroke({ color: 0x8a6a3a, width: r * 0.04, alpha: a });
      if (quality() === 'low') {
        return;
      }
      for (let i = -1; i <= 1; i++) {
        const t = (k * 1.6 + (i + 1) / 3) % 1;
        const sx = i * r * 0.22 + Math.sin(t * 9) * r * 0.06;
        g.moveTo(sx, -r * 0.4 - t * r * 0.8).lineTo(sx + r * 0.05, -r * 0.55 - t * r * 0.8)
          .stroke({ color: 0xffffff, width: r * 0.06, alpha: (1 - t) * 0.8 * a });
      }
    });
  }

  /**
   * noodle: sợi mì vàng nối Sư Phụ ↔ nạn nhân 0,5 s rồi rút về (spec §6.5). Hụt thì sợi mì
   * vươn tới mũi nón rồi RƠI xuống đất.
   */
  noodle(p: Path, now: number, miss: boolean): void {
    this.add(0, 0, now, miss ? 0.6 : 0.65, (g, k) => {
      const out = Math.min(1, k / 0.2);
      const back = miss ? 1 : k < 0.7 ? 1 : 1 - (k - 0.7) / 0.3;
      const reach = Math.min(out, back);
      const drop = miss && k > 0.3 ? (k - 0.3) / 0.7 : 0;
      const len = Math.hypot(p.x1 - p.x0, p.y1 - p.y0) || 1;
      const nx = -(p.y1 - p.y0) / len;
      const ny = (p.x1 - p.x0) / len;
      const steps = 18;
      g.moveTo(p.x0, p.y0);
      for (let i = 1; i <= steps; i++) {
        const t = (i / steps) * reach;
        const wob = Math.sin(t * 22 + k * 30) * len * 0.02 * (1 - drop);
        g.lineTo(p.x0 + (p.x1 - p.x0) * t + nx * wob, p.y0 + (p.y1 - p.y0) * t + ny * wob + drop * len * 0.25 * t);
      }
      g.stroke({ color: 0xf2c14e, width: Math.max(6, len * 0.025), alpha: 1 - drop * 0.8, cap: 'round' });
    });
  }

  /** stallFlash: khung sạp nhấp nháy 0,3 s trước khi bật lên (Cô Ba). */
  stallFlash(x: number, y: number, w: number, now: number): void {
    this.add(x, y, now, 0.3, (g, k) => {
      const on = Math.floor(k * 6) % 2 === 0;
      g.rect(-w / 2, -w / 2, w, w).fill({ color: 0xe08f3c, alpha: on ? 0.3 : 0.08 })
        .stroke({ color: 0xffd08a, width: Math.max(4, w * 0.06), alpha: on ? 1 : 0.4 });
    });
  }

  update(now: number): void {
    for (let i = this.live.length - 1; i >= 0; i--) {
      const f = this.live[i];
      const k = (now - f.born) / f.dur;
      if (k >= 1) {
        f.g.destroy();
        this.live.splice(i, 1);
        continue;
      }
      f.g.visible = k >= 0; // hiệu ứng xếp lượt (kiếm bay lần lượt) chưa tới giờ thì ẩn
      if (k < 0) {
        continue;
      }
      f.g.clear();
      f.draw(f.g, k);
    }
  }

  clear(): void {
    for (const f of this.live) {
      f.g.destroy();
    }
    this.live.length = 0;
  }

  private add(x: number, y: number, born: number, dur: number, draw: Fx['draw']): void {
    if (this.live.length >= MAX_FX) {
      this.live.shift()?.g.destroy();
    }
    const g = new Graphics();
    g.position.set(x, y);
    g.visible = false;
    this.layer.addChild(g);
    this.live.push({ g, born, dur, draw });
  }
}

/** along: điểm thứ k trên đường p; `bulge` > 0 vồng sang trái hướng đi (đường parabol). */
function along(p: Path, k: number, bulge: number): { x: number; y: number } {
  const e = k * k * (3 - 2 * k);
  const len = Math.hypot(p.x1 - p.x0, p.y1 - p.y0) || 1;
  const off = bulge * 4 * e * (1 - e);
  return {
    x: p.x0 + (p.x1 - p.x0) * e - ((p.y1 - p.y0) / len) * off,
    y: p.y0 + (p.y1 - p.y0) * e + ((p.x1 - p.x0) / len) * off,
  };
}

/** blade: một thanh kiếm nhỏ (lưỡi + chuôi) hướng theo góc `a`. */
function blade(g: Graphics, x: number, y: number, a: number, s: number, color: number): void {
  const c = Math.cos(a);
  const sn = Math.sin(a);
  const pt = (u: number, v: number): [number, number] => [x + c * u - sn * v, y + sn * u + c * v];
  g.poly([...pt(s * 0.5, 0), ...pt(-s * 0.2, s * 0.09), ...pt(-s * 0.2, -s * 0.09)]).fill(0xf4f4f4).stroke({ color, width: s * 0.06 });
  const [hx, hy] = pt(-s * 0.2, s * 0.18);
  const [kx, ky] = pt(-s * 0.2, -s * 0.18);
  g.moveTo(hx, hy).lineTo(kx, ky).stroke({ color, width: s * 0.08 });
  const [ex, ey] = pt(-s * 0.45, 0);
  const [bx, by] = pt(-s * 0.2, 0);
  g.moveTo(bx, by).lineTo(ex, ey).stroke({ color: 0x5a3b1a, width: s * 0.1 });
}
