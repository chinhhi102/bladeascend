import { Container, Graphics, Sprite, Texture } from 'pixi.js';

import type { GameConfig } from '../config';
import { hslToHex } from './palette';
import { quality } from './quality';

/**
 * Hào quang Tiên giới (spec §3.6).
 *
 * Bản cũ là **một đĩa phẳng + một vòng viền**, blend thường, không phát sáng, không chuyển
 * động nhiều lớp — nhìn như hình dán. Giờ dựng đúng 4 lớp của spec, mỗi lớp chuyển động với
 * tốc độ riêng và đều blend **cộng** để chồng lên nhau thì sáng lên chứ không che nhau:
 *
 *   1. quầng sáng mềm (gradient tròn), thở ±8% chu kỳ 2s
 *   2. vòng phù văn quay chậm, 1 vòng / 12s
 *   3. tia sáng tỏa ra quay NGƯỢC chiều, 1 vòng / 20s, độ trong 40%
 *   4. hạt sáng bay lên từ chân
 *
 * Đạo Tổ thêm lớp âm dương quay + viền đổi màu liên tục (hue shift).
 *
 * Cả bốn lớp nằm trong `vfxLayer` để ăn chung một lượt bloom (§3.6) — không gắn filter riêng
 * cho từng nhân vật.
 */

const BREATH_SECS = 2;
const RUNE_SECS = 12;
const RAY_SECS = 20;
const MOTES_BY_QUALITY = { high: 10, medium: 5, low: 3 } as const;
/** Nhân vật nhỏ hơn ngần này pixel trên màn hình thì chỉ còn lớp quầng sáng (§3.6). */
export const AURA_FAR_PX = 40;

let glowTex: Texture | null = null;

/** glowTexture: một đốm sáng gradient tròn, trắng để tint theo màu cảnh giới. */
function glowTexture(): Texture {
  if (glowTex) {
    return glowTex;
  }
  const n = 256;
  const cv = document.createElement('canvas');
  cv.width = n;
  cv.height = n;
  const ctx = cv.getContext('2d') as CanvasRenderingContext2D;
  const g = ctx.createRadialGradient(n / 2, n / 2, 0, n / 2, n / 2, n / 2);
  g.addColorStop(0, 'rgba(255,255,255,0.95)');
  g.addColorStop(0.35, 'rgba(255,255,255,0.45)');
  g.addColorStop(0.7, 'rgba(255,255,255,0.12)');
  g.addColorStop(1, 'rgba(255,255,255,0)');
  ctx.fillStyle = g;
  ctx.fillRect(0, 0, n, n);
  glowTex = Texture.from(cv);
  return glowTex;
}

interface Mote {
  x: number;
  phase: number;
  speed: number;
}

export class TienAura {
  readonly root = new Container();
  private readonly glow = new Sprite(glowTexture());
  private readonly runes = new Graphics();
  private readonly rays = new Graphics();
  private readonly motes = new Graphics();
  private readonly yinYang = new Graphics();
  private readonly dust: Mote[] = [];
  private idx = -1;
  private color = 0xffffff;
  private rainbow = false;

  constructor(private readonly cfg: GameConfig) {
    this.glow.anchor.set(0.5);
    for (const n of [this.glow, this.yinYang, this.runes, this.rays, this.motes]) {
      n.blendMode = 'add'; // spec §3.6 nguyên tắc 1: chồng lên nhau thì SÁNG lên
      this.root.addChild(n);
    }
    for (let i = 0; i < MOTES_BY_QUALITY.high; i++) {
      this.dust.push({ x: Math.random() * 2 - 1, phase: Math.random(), speed: 0.25 + Math.random() * 0.35 });
    }
  }

  /** setRealm đổi màu theo cảnh giới Tiên (idx 0–4); -1 = không phải Tiên giới. */
  setRealm(idx: number): void {
    if (idx === this.idx) {
      return;
    }
    this.idx = idx;
    const spec = this.cfg.cultivation.tien_aura_colors[idx] ?? '#ffffff';
    this.rainbow = spec === 'yin_yang';
    this.color = this.rainbow ? 0xffffff : Number(`0x${spec.replace('#', '')}`);
    this.yinYang.visible = this.rainbow;
  }

  /**
   * update đặt hào quang quanh nhân vật ở (cx, feetY) với bán kính thân r.
   * `near` = nhân vật đủ to trên màn hình; nhỏ quá thì chỉ giữ lớp quầng sáng (§3.6).
   */
  update(cx: number, feetY: number, r: number, now: number, alpha: number, near: boolean): void {
    this.root.position.set(cx, feetY);
    this.root.alpha = alpha;
    const charH = r * 2.6;
    const cy = -charH * 0.45; // ngang ngực
    const breathe = 1 + 0.08 * Math.sin((now * 2 * Math.PI) / BREATH_SECS);
    const color = this.rainbow ? hslToHex((now * 50) % 360, 0.85, 0.6) : this.color;

    // Lớp 1 — quầng sáng mềm
    const glowR = charH * 1.6 * breathe;
    this.glow.position.set(0, cy);
    this.glow.width = glowR * 2;
    this.glow.height = glowR * 2;
    this.glow.tint = color;

    this.runes.visible = near;
    this.rays.visible = near;
    this.motes.visible = near;
    if (!near) {
      this.yinYang.clear();
      return;
    }

    // Lớp 2 — vòng phù văn quay chậm (1 vòng / 12s)
    const a2 = (now * 2 * Math.PI) / RUNE_SECS;
    const rr = charH * 0.95;
    const gr = this.runes.clear();
    gr.circle(0, cy, rr).stroke({ color, width: r * 0.07, alpha: 0.6 });
    gr.circle(0, cy, rr * 0.78).stroke({ color, width: r * 0.04, alpha: 0.4 });
    const marks = 8 + this.idx * 2;
    for (let i = 0; i < marks; i++) {
      const a = a2 + (i / marks) * Math.PI * 2;
      const c = Math.cos(a);
      const s = Math.sin(a);
      gr.moveTo(c * rr * 0.8, cy + s * rr * 0.8).lineTo(c * rr * 1.04, cy + s * rr * 1.04);
    }
    gr.stroke({ color, width: r * 0.08, alpha: 0.75 });

    // Lớp 3 — tia sáng tỏa ra, quay NGƯỢC chiều (1 vòng / 20s), độ trong 40%
    const a3 = -(now * 2 * Math.PI) / RAY_SECS;
    const ray = this.rays.clear();
    for (let i = 0; i < 14; i++) {
      const a = a3 + (i / 14) * Math.PI * 2;
      const c = Math.cos(a);
      const s = Math.sin(a);
      const long = i % 2 === 0 ? 1.75 : 1.4;
      ray.moveTo(c * charH * 1.0, cy + s * charH * 1.0).lineTo(c * charH * long, cy + s * charH * long);
    }
    ray.stroke({ color, width: r * 0.06, alpha: 0.4 });

    // Lớp 4 — hạt sáng bay lên từ chân
    const n = MOTES_BY_QUALITY[quality()];
    const mg = this.motes.clear();
    for (let i = 0; i < n; i++) {
      const d = this.dust[i];
      const t = (now * d.speed + d.phase) % 1;
      const y = -charH * 2.1 * t;
      mg.circle(d.x * r * 1.6, y, r * 0.11 * (1 - t)).fill({ color, alpha: 0.9 * (1 - t) });
    }

    // Đạo Tổ — âm dương quay chậm, viền đổi màu liên tục
    const yy = this.yinYang.clear();
    if (this.rainbow) {
      const a = now * 0.6;
      const rad = charH * 0.8;
      yy.circle(0, cy, rad).fill({ color: 0x0a0a0a, alpha: 0.55 });
      yy.arc(0, cy, rad, a, a + Math.PI).fill({ color: 0xf2f2f2, alpha: 0.55 });
      yy.circle(Math.cos(a) * rad * 0.5, cy + Math.sin(a) * rad * 0.5, rad * 0.5).fill({ color: 0xf2f2f2, alpha: 0.55 });
      yy.circle(-Math.cos(a) * rad * 0.5, cy - Math.sin(a) * rad * 0.5, rad * 0.5).fill({ color: 0x0a0a0a, alpha: 0.55 });
      yy.circle(0, cy, rad).stroke({ color, width: r * 0.16, alpha: 0.95 });
    }
  }

  destroy(): void {
    this.root.destroy({ children: true });
  }
}
