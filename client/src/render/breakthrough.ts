import { ShockwaveFilter } from 'pixi-filters';
import { Container, Graphics, Text } from 'pixi.js';

import { quality } from './quality';

/**
 * Màn đột phá Tiên giới (spec §3.6 mục 5 — "juice"): sóng xung kích, khựng hình 80ms, chớp
 * trắng một khung, và tên cảnh giới bật ra (scale 0 → 1.3 → 1.0, easing back-out).
 *
 * Khựng hình (hit-stop) là mẹo cũ của game hành động: dừng thời gian một nhịp rất ngắn ngay
 * lúc cú đánh chạm, não người đọc thành "cú này nặng". 80ms đủ cảm thấy mà chưa thành giật.
 */

const SHOCK_SECS = 0.7;
const FLASH_SECS = 0.12;
const TEXT_SECS = 1.6;
export const HITSTOP_SECS = 0.08;

/** backOut: easing vọt quá rồi lùi về, cho chữ bật ra như lò xo. */
function backOut(t: number): number {
  const c = 1.9;
  const p = t - 1;
  return 1 + (c + 1) * p * p * p + c * p * p;
}

export class Breakthrough {
  /** layer phủ toàn màn hình: chớp trắng + chữ cảnh giới. */
  readonly overlay = new Container();
  private readonly flash = new Graphics();
  private readonly label: Text;
  private flashUntil = 0;
  private textFrom = -1;
  private shockFrom = -1;
  private shock: ShockwaveFilter | null = null;
  private target: Container | null = null;

  constructor() {
    this.flash.blendMode = 'add';
    this.label = new Text({
      text: '',
      style: { fontFamily: 'system-ui, sans-serif', fontSize: 64, fontWeight: '900', fill: 0xfff3a0, stroke: { color: 0x3a2000, width: 10 } },
    });
    this.label.anchor.set(0.5);
    this.label.visible = false;
    this.overlay.addChild(this.flash, this.label);
    this.overlay.eventMode = 'none';
  }

  /**
   * play bắt đầu màn đột phá. `at` là vị trí trên MÀN HÌNH (0–1 theo chiều rộng/cao) để
   * đặt tâm sóng xung kích; `camera` là container sẽ bị gắn filter sóng.
   */
  play(realmName: string, at: { x: number; y: number }, camera: Container, now: number): void {
    this.flashUntil = now + FLASH_SECS;
    this.textFrom = now;
    this.label.text = realmName;
    this.label.visible = true;
    if (quality() === 'low') {
      return; // mức Thấp: giữ chớp + chữ, bỏ filter (spec §3.6)
    }
    this.shock = new ShockwaveFilter({ center: { x: at.x, y: at.y }, amplitude: 28, wavelength: 180, speed: 900, radius: -1 });
    camera.filters = [...((camera.filters as unknown[]) ?? []), this.shock] as never;
    this.target = camera;
    this.shockFrom = now;
  }

  /** update chạy mỗi khung; trả về true khi đang khựng hình để vòng lặp bỏ qua một nhịp. */
  update(now: number, screenW: number, screenH: number): void {
    // Chớp trắng
    const f = this.flash.clear();
    if (now < this.flashUntil) {
      const k = 1 - (this.flashUntil - now) / FLASH_SECS;
      f.rect(0, 0, screenW, screenH).fill({ color: 0xffffff, alpha: 0.75 * (1 - k) });
    }

    // Chữ cảnh giới bật ra rồi mờ đi
    if (this.textFrom >= 0) {
      const t = (now - this.textFrom) / TEXT_SECS;
      if (t >= 1) {
        this.label.visible = false;
        this.textFrom = -1;
      } else {
        const pop = t < 0.35 ? backOut(t / 0.35) : 1;
        this.label.position.set(screenW / 2, screenH * 0.28);
        this.label.scale.set(pop * (t > 0.8 ? 1 + (t - 0.8) * 0.8 : 1));
        this.label.alpha = t > 0.7 ? 1 - (t - 0.7) / 0.3 : 1;
      }
    }

    // Sóng xung kích lan ra rồi gỡ filter
    if (this.shock && this.target) {
      const t = now - this.shockFrom;
      this.shock.time = t;
      if (t >= SHOCK_SECS) {
        this.target.filters = ((this.target.filters as unknown[]) ?? []).filter((x) => x !== this.shock) as never;
        this.shock = null;
        this.target = null;
      }
    }
  }
}
