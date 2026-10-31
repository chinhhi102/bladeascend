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

/** FADE_SECS: chỉ báo mờ dần trong ngần này. Đủ để kịp nhìn, chưa đủ để che màn hình. */
const FADE_SECS = 0.45;

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
    g.moveTo(0, 0).arc(0, 0, r, facing - half, facing + half).lineTo(0, 0);
    g.fill({ color, alpha: 0.16 }).stroke({ color, width: Math.max(3, r * 0.012), alpha: 0.7 });
    g.position.set(x, y);
    this.add(g, now);
  }

  /** circle vẽ vòng tròn: Ru Ngủ, Lưới Mì, sóng Rụt Cổ. */
  circle(x: number, y: number, scale: number, key: string, color: number, now: number): void {
    const r = skillRange(this.cfg, scale, key);
    if (r <= 0) {
      return;
    }
    const g = new Graphics();
    g.circle(0, 0, r).fill({ color, alpha: 0.14 }).stroke({ color, width: Math.max(3, r * 0.02), alpha: 0.75 });
    g.position.set(x, y);
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
