import { Container, Graphics } from 'pixi.js';

import { type GameConfig, skillRange } from '../config';

/**
 * Chỉ báo phạm vi kỹ năng (GAME_SPEC §6.3, CHANGES_v6_5 mục 3).
 *
 * Vẽ **đúng vùng server xét trúng**: nón của Loa Phường / Phun Trân Châu, vòng tròn của Ru
 * Ngủ / Lưới Mì / sóng Rụt Cổ. Hệ số lấy từ `skills.ranges_u` — cùng bảng server đọc — nhân
 * với U của **người dùng** (`pickup_r × scale`).
 *
 * Vì sao phải vẽ: từ v6.5 phạm vi co giãn theo cảnh giới, nên cùng một kỹ năng ở Luyện Khí
 * với 540 đơn vị còn ở Đạo Tổ với 3 240. Không vẽ ra thì người chơi không có cách nào biết
 * mình đang đứng trong hay ngoài tầm, và cũng không học được tầm của người khác.
 *
 * Vẽ bằng `Graphics` dựng một lần cho mỗi lần cast rồi mờ dần — không phải mỗi khung hình.
 */

/**
 * FADE_SECS: chỉ báo loé lên rồi tan trong ngần này — spec v6.36 §6.5 lớp ②: "loé lên 0.3s,
 * nét đứt sáng, màu riêng kỹ năng".
 */
const FADE_SECS = 0.3;
/** Số nét đứt trên một vòng tròn — cố định theo góc nên vòng to hay nhỏ đều đọc ra "nét đứt". */
const DASHES = 36;

interface Mark {
  g: Graphics;
  born: number;
}

export class SkillFx {
  readonly layer = new Container();
  private readonly marks: Mark[] = [];

  constructor(private readonly cfg: GameConfig) {
    this.layer.zIndex = 50;
  }

  /** cone vẽ hình quạt: dùng cho Loa Phường (60°) và Phun Trân Châu (45°). */
  cone(x: number, y: number, facing: number, scale: number, key: string, deg: number, color: number, now: number): void {
    const r = skillRange(this.cfg, scale, key);
    if (r <= 0) {
      return;
    }
    const half = (deg * Math.PI) / 360;
    const g = new Graphics();
    g.moveTo(0, 0).arc(0, 0, r, facing - half, facing + half).lineTo(0, 0).fill({ color, alpha: 0.16 });
    const w = Math.max(3, r * 0.014);
    dashedLine(g, 0, 0, Math.cos(facing - half) * r, Math.sin(facing - half) * r, r / 14);
    dashedLine(g, 0, 0, Math.cos(facing + half) * r, Math.sin(facing + half) * r, r / 14);
    dashedArc(g, r, facing - half, facing + half, Math.max(4, Math.round((DASHES * deg) / 360)));
    g.stroke({ color: brighten(color), width: w, alpha: 0.95 });
    g.position.set(x, y);
    this.add(g, now);
  }

  /** circle vẽ vòng tròn: Ru Ngủ Trưa, vòng hút Gom Đơn, cung sạp, sóng Rụt Cổ. */
  circle(x: number, y: number, scale: number, key: string, color: number, now: number): void {
    const r = skillRange(this.cfg, scale, key);
    if (r <= 0) {
      return;
    }
    const g = new Graphics();
    g.circle(0, 0, r).fill({ color, alpha: 0.14 });
    dashedArc(g, r, 0, Math.PI * 2, DASHES);
    g.stroke({ color: brighten(color), width: Math.max(3, r * 0.022), alpha: 0.95 });
    g.position.set(x, y);
    this.add(g, now);
  }

  /**
   * beam vẽ một đoạn thẳng giữa hai điểm: sợi mì của Sư Phụ.
   *
   * Khác `cone`/`circle` ở chỗ nó KHÔNG đọc `ranges_u`: đây không phải chỉ báo phạm vi mà
   * là đường đi của một thứ đã trúng rồi — hai đầu là hai người thật, server gửi kèm.
   */
  beam(x0: number, y0: number, x1: number, y1: number, color: number, now: number): void {
    const g = new Graphics();
    g.moveTo(x0, y0).lineTo(x1, y1).stroke({ color, width: 10, alpha: 0.85 });
    this.add(g, now);
  }

  private add(g: Graphics, now: number): void {
    this.layer.addChild(g);
    this.marks.push({ g, born: now });
  }

  /** update mờ dần rồi huỷ. Gọi mỗi khung từ `GameScene`. */
  update(now: number): void {
    for (let i = this.marks.length - 1; i >= 0; i--) {
      const m = this.marks[i];
      const k = (now - m.born) / FADE_SECS;
      if (k >= 1) {
        m.g.destroy();
        this.marks.splice(i, 1);
        continue;
      }
      m.g.alpha = 1 - k;
      // Phình nhẹ ra khi tan: đọc ra là "sóng lan", không phải "hình bị xoá".
      m.g.scale.set(1 + k * 0.08);
    }
  }

  clear(): void {
    for (const m of this.marks) {
      m.g.destroy();
    }
    this.marks.length = 0;
  }
}

/** dashedArc: cung nét đứt bán kính r từ a0 tới a1, `n` nét (nửa nét, nửa khoảng trống). */
function dashedArc(g: Graphics, r: number, a0: number, a1: number, n: number): void {
  const step = (a1 - a0) / n;
  for (let i = 0; i < n; i++) {
    const s = a0 + i * step;
    g.moveTo(Math.cos(s) * r, Math.sin(s) * r).arc(0, 0, r, s, s + step * 0.55);
  }
}

/** dashedLine: đoạn thẳng nét đứt, mỗi nét dài `dash`. */
function dashedLine(g: Graphics, x0: number, y0: number, x1: number, y1: number, dash: number): void {
  const len = Math.hypot(x1 - x0, y1 - y0);
  const n = Math.max(1, Math.floor(len / (dash * 2)));
  for (let i = 0; i < n; i++) {
    const a = (i * 2 * dash) / len;
    const b = Math.min(1, ((i * 2 + 1) * dash) / len);
    g.moveTo(x0 + (x1 - x0) * a, y0 + (y1 - y0) * a).lineTo(x0 + (x1 - x0) * b, y0 + (y1 - y0) * b);
  }
}

/** brighten: nét viền sáng hơn màu nền một bậc — spec "nét đứt sáng". */
function brighten(c: number): number {
  const ch = (sh: number): number => Math.min(255, (((c >> sh) & 0xff) * 1.35 + 40) | 0);
  return (ch(16) << 16) | (ch(8) << 8) | ch(0);
}
