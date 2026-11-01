import { Container, Graphics } from 'pixi.js';

/**
 * EdgeArrows: bị trúng kỹ năng từ NGOÀI khung nhìn → mũi tên ở mép màn hình chỉ về phía
 * người dùng kỹ năng (spec v6.36 §6.5). Vẽ trong toạ độ MÀN HÌNH (ngoài camera), mỗi mũi
 * tên sống 1,5 s, mờ dần. Toạ độ người gây lấy từ X/Y của event `StatusStart`.
 */

const LIFE = 1.5;
const MARGIN = 44; // px cách mép màn hình
const SIZE = 26;
const MAX_ARROWS = 4;

interface Arrow {
  g: Graphics;
  born: number;
}

export class EdgeArrows {
  readonly root = new Container();
  private readonly live: Arrow[] = [];

  constructor() {
    this.root.eventMode = 'none';
  }

  /**
   * show: (sx, sy) là vị trí người gây trên MÀN HÌNH (có thể nằm ngoài khung); w, h là cỡ
   * màn hình. Người gây đang trong khung nhìn thì không vẽ — họ đã thấy được rồi.
   */
  show(sx: number, sy: number, w: number, h: number, now: number): boolean {
    if (sx >= 0 && sx <= w && sy >= 0 && sy <= h) {
      return false;
    }
    const cx = w / 2;
    const cy = h / 2;
    const a = Math.atan2(sy - cy, sx - cx);
    // Chiếu tia từ tâm tới mép hình chữ nhật đã thu vào MARGIN.
    const hw = cx - MARGIN;
    const hh = cy - MARGIN;
    const t = Math.min(Math.abs(hw / (Math.cos(a) || 1e-6)), Math.abs(hh / (Math.sin(a) || 1e-6)));
    const g = new Graphics();
    g.poly([SIZE, 0, -SIZE * 0.6, SIZE * 0.7, -SIZE * 0.25, 0, -SIZE * 0.6, -SIZE * 0.7])
      .fill(0xff4d4d).stroke({ color: 0xffffff, width: 3 });
    g.position.set(cx + Math.cos(a) * t, cy + Math.sin(a) * t);
    g.rotation = a;
    if (this.live.length >= MAX_ARROWS) {
      this.live.shift()?.g.destroy();
    }
    this.root.addChild(g);
    this.live.push({ g, born: now });
    return true;
  }

  update(now: number): void {
    for (let i = this.live.length - 1; i >= 0; i--) {
      const a = this.live[i];
      const k = (now - a.born) / LIFE;
      if (k >= 1) {
        a.g.destroy();
        this.live.splice(i, 1);
        continue;
      }
      // Nhấp nháy 3 lần trong nửa đầu cho bắt mắt, rồi mờ dần.
      a.g.alpha = k < 0.5 ? 0.6 + 0.4 * Math.abs(Math.sin(k * Math.PI * 6)) : 1 - (k - 0.5) / 0.5;
      a.g.scale.set(1 + 0.15 * Math.sin(k * Math.PI * 6));
    }
  }

  clear(): void {
    for (const a of this.live) {
      a.g.destroy();
    }
    this.live.length = 0;
  }
}
