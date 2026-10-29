import { Container, Graphics, Sprite, Texture } from 'pixi.js';

import { quality } from './quality';

/**
 * Trân châu dính màn hình — kỹ năng Phun Trân Châu của Thằng Mập (spec §6.3).
 *
 * Chỉ hiện trên màn hình **người bị trúng**: server gửi riêng event `Blinded {secs}` cho
 * đúng người đó, không ai khác thấy lớp này.
 *
 * Ba điều spec đòi và dễ làm sai:
 *   - **Không che HUD**: lớp này nằm dưới HUD trong DOM, và vẽ trên canvas game chứ không
 *     phải lên toàn trang — thanh tiến độ, nút bấm và minimap phải luôn đọc được, nếu không
 *     người chơi mất phương hướng hoàn toàn chứ không chỉ bị cản tầm nhìn.
 *   - **Dày hơn ở giữa**: rải đều thì nhìn như hạt nhiễu; dồn vào giữa mới ra cảm giác bị
 *     phun thẳng vào mặt.
 *   - **Trượt xuống để lại vệt** rồi mới tan, chứ không đứng im rồi mờ dần.
 *
 * Trúng lại khi đang dính thì làm mới thời gian và thêm 10 viên (spec §6.3).
 */

/** Số viên theo chất lượng; spec nói 25–40 viên ở mức bình thường. */
const COUNT = { high: 40, medium: 30, low: 25 } as const;
/** Mỗi lần trúng thêm khi đang dính thì cộng ngần này viên. */
const EXTRA_ON_REFRESH = 10;
/** Tỉ lệ diện tích màn hình bị phủ (spec: 55–65%). */
const COVER = 0.6;

interface Blob {
  node: Container;
  vy: number;
  spin: number;
  born: number;
  /** Chiều dài vệt trà sữa kéo theo. */
  trail: Graphics;
}

function bobaTexture(): Texture {
  const n = 64;
  const cv = document.createElement('canvas');
  cv.width = n;
  cv.height = n;
  const c = cv.getContext('2d') as CanvasRenderingContext2D;
  // Viên trân châu: đen bóng, có một chấm sáng lệch lên trên-trái cho ra chất "bóng".
  const g = c.createRadialGradient(n * 0.38, n * 0.34, n * 0.05, n / 2, n / 2, n / 2);
  g.addColorStop(0, '#6b5a4e');
  g.addColorStop(0.35, '#2a211c');
  g.addColorStop(1, '#120d0a');
  c.fillStyle = g;
  c.beginPath();
  c.arc(n / 2, n / 2, n / 2 - 1, 0, Math.PI * 2);
  c.fill();
  c.fillStyle = 'rgba(255,255,255,.55)';
  c.beginPath();
  c.ellipse(n * 0.36, n * 0.3, n * 0.1, n * 0.07, -0.6, 0, Math.PI * 2);
  c.fill();
  return Texture.from(cv);
}

let tex: Texture | null = null;

export class BobaOverlay {
  readonly root = new Container();
  private readonly blobs: Blob[] = [];
  private until = 0;
  private w = 0;
  private h = 0;

  constructor() {
    this.root.visible = false;
    this.root.eventMode = 'none';
  }

  /** active: còn đang bị bịt mắt không — dùng cho overlay F3 và test. */
  get active(): boolean {
    return this.root.visible;
  }

  /**
   * hit bắt đầu (hoặc làm mới) hiệu ứng. `secs` lấy thẳng từ server để client không tự
   * đoán thời gian — lệch một chút là người chơi thấy hết trân châu trong khi server vẫn
   * tính là đang bị bịt mắt.
   */
  hit(secs: number, now: number, w: number, h: number): void {
    this.w = w;
    this.h = h;
    const fresh = !this.root.visible;
    this.until = now + secs;
    this.root.visible = true;
    const n = fresh ? COUNT[quality()] : EXTRA_ON_REFRESH;
    for (let i = 0; i < n; i++) {
      this.addBlob(now);
    }
  }

  private addBlob(now: number): void {
    if (!tex) {
      tex = bobaTexture();
    }
    const node = new Container();
    const trail = new Graphics();
    const sp = new Sprite(tex);
    sp.anchor.set(0.5);
    // Cỡ viên tính theo diện tích cần phủ, chia cho số viên — nhờ vậy ở màn hình nào cũng
    // phủ đúng tỉ lệ, không phải màn to thì thưa thớt.
    const area = (this.w * this.h * COVER) / COUNT[quality()];
    const r = Math.sqrt(area / Math.PI) * (0.7 + Math.random() * 0.7);
    sp.width = r * 2;
    sp.height = r * 2 * (0.85 + Math.random() * 0.3);

    // Dồn về giữa: lấy trung bình hai số ngẫu nhiên thì phân bố co về tâm.
    const cx = ((Math.random() + Math.random()) / 2) * this.w;
    const cy = ((Math.random() + Math.random()) / 2) * this.h;
    node.position.set(cx, cy);
    node.addChild(trail, sp);
    node.rotation = Math.random() * Math.PI * 2;
    this.root.addChild(node);
    this.blobs.push({ node, vy: 8 + Math.random() * 26, spin: (Math.random() - 0.5) * 0.4, born: now, trail });
  }

  /** resize cập nhật khung màn hình; gọi khi cửa sổ đổi cỡ. */
  resize(w: number, h: number): void {
    this.w = w;
    this.h = h;
  }

  update(now: number, dt: number): void {
    if (!this.root.visible) {
      return;
    }
    if (now >= this.until) {
      this.clear();
      return;
    }
    const left = this.until - now;
    for (const b of this.blobs) {
      b.node.y += b.vy * dt;
      b.node.rotation += b.spin * dt;
      // Vệt trà sữa kéo lên trên theo quãng đã trượt.
      const slid = (now - b.born) * b.vy;
      b.trail.clear();
      if (slid > 4) {
        b.trail.rect(-6, -slid, 12, slid).fill({ color: 0x4a3326, alpha: 0.35 });
      }
      // Giây cuối thì tan dần, trước đó giữ nguyên độ đậm để che thật sự.
      b.node.alpha = Math.min(1, left);
    }
  }

  clear(): void {
    for (const b of this.blobs) {
      b.node.destroy({ children: true });
    }
    this.blobs.length = 0;
    this.root.visible = false;
    this.until = 0;
  }
}
