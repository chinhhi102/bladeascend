import { Container, Graphics, Particle, ParticleContainer, Texture } from 'pixi.js';

import { type GameConfig, realmOf } from '../config';
import { BUFFS } from '../net/protocol';
import type { PlayerFrame } from '../net/state';
import { assets } from './assets';
import { REALM_COLORS } from './palette';
import { bladeLen, hiltR, ringCap, type RingGeom, ringGeom, ringOf, ringStart, slotAngle, tipDist } from './ringgeom';

// Drawing proportions (visual only, not balance numbers).
const HILT_ANCHOR = 0.95; // blade art: hilt sits at 95% of the canvas height, tip at the top
const ART_FILL = 0.9; // blade art fills ~90% of its canvas height
const GHOSTS = 2; // trailing copies per sword that fake motion blur (spec §4.1b)
const GHOST_STEP = (7 * Math.PI) / 180;
const MAIN_ALPHA = 0.95;
const LONG_BIT = 1 << BUFFS.indexOf('long');
const SPIN_BIT = 1 << BUFFS.indexOf('spin');

let bladeTex: Texture | null = null;

/** placeholderBlade: a white/grey blade so every ring can be tinted (spec v3.1 item 6). */
function placeholderBlade(): Texture {
  if (bladeTex) {
    return bladeTex;
  }
  const w = 24;
  const h = 128;
  const cv = document.createElement('canvas');
  cv.width = w;
  cv.height = h;
  const ctx = cv.getContext('2d') as CanvasRenderingContext2D;
  ctx.beginPath();
  ctx.moveTo(w / 2, 2);
  ctx.lineTo(w - 4, 20);
  ctx.lineTo(w - 5, h - 26);
  ctx.lineTo(5, h - 26);
  ctx.lineTo(4, 20);
  ctx.closePath();
  const grad = ctx.createLinearGradient(0, 0, w, 0);
  grad.addColorStop(0, '#d9d9d9');
  grad.addColorStop(0.5, '#ffffff');
  grad.addColorStop(1, '#bdbdbd');
  ctx.fillStyle = grad;
  ctx.fill();
  ctx.lineWidth = 2;
  ctx.strokeStyle = '#3a3a3a';
  ctx.stroke();
  ctx.fillStyle = '#8a8a8a';
  ctx.fillRect(1, h - 26, w - 2, 6); // guard
  ctx.fillRect(w / 2 - 3, h - 20, 6, 16); // grip
  bladeTex = Texture.from(cv);
  return bladeTex;
}

function parseTint(hex: string): number {
  return Number.parseInt(hex.replace('#', ''), 16) || 0xffffff;
}

interface Blade {
  main: Particle;
  ghosts: Particle[];
  ring: number;
}

/**
 * SwordRing draws a player's nested sword rings (spec §4.1a/b): outer rings first, every ring
 * tinted with ring_tints, blades leaning by the server's tilt, a spin-blur annulus per ring and
 * fading ghost copies so a 300 ms revolution never looks frozen or reversed.
 */
export class SwordRing {
  readonly root = new Container();
  private readonly blur = new Graphics();
  private readonly tinted = new ParticleContainer({ dynamicProperties: { position: true, rotation: true, vertex: true, color: true } });
  private readonly skin = new ParticleContainer({ dynamicProperties: { position: true, rotation: true, vertex: true, color: true } });
  private readonly cracks = new Graphics();
  private readonly hitbox = new Graphics();
  private blades: Blade[] = [];
  private angles: number[] = [];
  private realm = -1;
  private count = 0;
  private prevSlots: number[] = [];
  private tweenStart = 0;
  private readonly tints: number[];

  constructor(private readonly cfg: GameConfig) {
    this.tints = cfg.sword.ring_tints.map(parseTint);
    this.root.addChild(this.blur, this.tinted, this.skin, this.cracks, this.hitbox);
  }

  update(f: PlayerFrame, now: number, showHitbox: boolean): void {
    const s = f.s;
    const realm = realmOf(this.cfg, s.L);
    const g = ringGeom(this.cfg, s.sc, s.n, s.bf & LONG_BIT ? (this.cfg.buffs.long?.mult ?? 1) : 1);
    if (s.n !== this.count || realm !== this.realm) {
      this.rebuild(s.n, realm, g, now);
    }
    this.root.position.set(f.x, f.y);
    this.root.zIndex = f.y + 1;
    const t = Math.min(1, (now - this.tweenStart) / this.cfg.sword.tween_secs);
    const spin = s.bf & SPIN_BIT ? (this.cfg.buffs.spin?.mult ?? 1) : 1;
    this.drawBlur(g, s.tl, realm, spin);
    for (let i = 0; i < this.blades.length; i++) {
      const a = f.ringAngle + this.slot(i, s.n, g, t);
      this.angles[i] = a;
      this.placeBlade(this.blades[i], g, a, s.tl, s.d);
    }
    this.drawCracks(f, g);
    this.drawHitbox(f, g, showHitbox);
  }

  /** slot is sword i's angle offset, tweened for tween_secs after the count changes. */
  private slot(i: number, n: number, g: RingGeom, t: number): number {
    const to = slotAngle(g, n, i);
    const from = this.prevSlots[i];
    if (from === undefined || t >= 1) {
      return to;
    }
    let d = to - from;
    d = Math.atan2(Math.sin(d), Math.cos(d));
    return from + d * t;
  }

  private placeBlade(b: Blade, g: RingGeom, a: number, tilt: number, dir: number): void {
    const h = hiltR(g, b.ring);
    const len = bladeLen(g, b.ring);
    const set = (p: Particle, ang: number): void => {
      p.x = Math.cos(ang) * h;
      p.y = Math.sin(ang) * h;
      p.rotation = ang + tilt + Math.PI / 2;
      const ky = len / (p.texture.height * ART_FILL);
      p.scaleY = ky;
      p.scaleX = p.texture === bladeTex ? g.width / p.texture.width : ky;
    };
    set(b.main, a);
    b.ghosts.forEach((p, j) => set(p, a - dir * GHOST_STEP * (j + 1)));
  }

  /**
   * drawBlur paints the motion-blur annulus of each ring — but ONLY for rings that really do
   * sweep their whole annulus every tick, using the same test the server uses to decide a
   * ring is a solid damage band (spec §4.1b: `số kiếm × góc quét ≥ 360°`).
   *
   * Without that test a ring holding 1 sword out of 80 still painted a full faint circle, so
   * the drawn range looked far wider than what the blades can actually hit.
   */
  private drawBlur(g: RingGeom, tilt: number, realm: number, spin: number): void {
    const gr = this.blur.clear();
    const sweepPerTick = (2 * Math.PI * spin) / (this.cfg.sword.rev_secs * this.cfg.net.tick_hz);
    for (let k = g.rings - 1; k >= 0; k--) {
      const cap = ringCap(g, k);
      const inRing = Math.min(cap, this.count - ringStart(g, k));
      if (inRing * sweepPerTick < 2 * Math.PI) {
        continue; // vòng thưa: từng lưỡi kiếm nhìn rõ, không có vành mờ liền
      }
      const h = hiltR(g, k);
      const tip = tipDist(g, k, tilt);
      const alpha = Math.min(0.4, (0.08 + 0.22 * (inRing / cap)) * spin);
      gr.circle(0, 0, (h + tip) / 2).stroke({ color: this.ringTint(k, realm, true), width: tip - h, alpha });
    }
  }

  private ringTint(k: number, realm: number, forBlur: boolean): number {
    if (k === 0) {
      // Ring 1 keeps the realm skin's own colour so players can read the realm (spec §4.1a).
      return this.hasSkin(realm) && !forBlur ? 0xffffff : (REALM_COLORS[realm] ?? 0xffffff);
    }
    return this.tints[k % this.tints.length] ?? 0xffffff;
  }

  private hasSkin(realm: number): boolean {
    return assets.texture(`swords/realm_${realm}.png`) !== null;
  }

  private rebuild(n: number, realm: number, g: RingGeom, now: number): void {
    this.prevSlots = this.blades.map((_, i) => slotAngle(g, this.count, i));
    this.tweenStart = now;
    this.count = n;
    this.realm = realm;
    this.tinted.removeParticles();
    this.skin.removeParticles();
    this.blades = [];
    const skinTex = assets.texture(`swords/realm_${realm}.png`);
    const plain = placeholderBlade();
    const byRing: Blade[][] = [];
    for (let i = 0; i < n; i++) {
      const k = ringOf(g, i);
      const tex = k === 0 && skinTex ? skinTex : plain;
      const tint = this.ringTint(k, realm, false);
      const mk = (alpha: number): Particle => new Particle({ texture: tex, anchorX: 0.5, anchorY: HILT_ANCHOR, tint, alpha });
      const b: Blade = { main: mk(MAIN_ALPHA), ghosts: Array.from({ length: GHOSTS }, (_, j) => mk(0.35 / (j + 1))), ring: k };
      this.blades.push(b);
      (byRing[k] ??= []).push(b);
    }
    // Outer rings first (drawn underneath), ring 1 last (on top); ghosts under their blades.
    for (let k = byRing.length - 1; k >= 0; k--) {
      const target = k === 0 && skinTex ? this.skin : this.tinted;
      for (const b of byRing[k] ?? []) {
        b.ghosts.forEach((p) => target.addParticle(p));
      }
      for (const b of byRing[k] ?? []) {
        target.addParticle(b.main);
      }
    }
  }

  // Light crack lines across ring 1 while the player carries crack damage (spec §5.2).
  private drawCracks(f: PlayerFrame, g: RingGeom): void {
    const gr = this.cracks.clear();
    const ck = f.s.ck ?? 0;
    if (ck <= 0) {
      return;
    }
    const h = hiltR(g, 0);
    const len = bladeLen(g, 0);
    const z = g.width * 0.4;
    for (let i = 0; i < Math.min(this.blades.length, ringCap(g, 0)); i++) {
      const a = this.angles[i] ?? 0;
      const hx = Math.cos(a) * h;
      const hy = Math.sin(a) * h;
      const dx = Math.cos(a + f.s.tl);
      const dy = Math.sin(a + f.s.tl);
      const at = (t: number, side: number): [number, number] => [hx + dx * len * t - dy * z * side, hy + dy * len * t + dx * z * side];
      gr.moveTo(...at(0.3, 1))
        .lineTo(...at(0.5, -1))
        .lineTo(...at(0.7, 1));
    }
    gr.stroke({ color: 0xffffff, width: Math.max(1.5, g.width * 0.18), alpha: 0.35 + 0.5 * Math.min(1, ck) });
  }

  /** drawHitbox overlays the server's sword capsules (debug key H, spec §17). */
  private drawHitbox(f: PlayerFrame, g: RingGeom, show: boolean): void {
    const gr = this.hitbox.clear();
    if (!show) {
      return;
    }
    for (let i = 0; i < f.s.n; i++) {
      const k = ringOf(g, i);
      const a = f.ringAngle + slotAngle(g, f.s.n, i);
      const h = hiltR(g, k);
      const hx = Math.cos(a) * h;
      const hy = Math.sin(a) * h;
      const len = bladeLen(g, k);
      gr.moveTo(hx, hy).lineTo(hx + Math.cos(a + f.s.tl) * len, hy + Math.sin(a + f.s.tl) * len);
    }
    gr.stroke({ color: 0xff2d55, width: g.width, alpha: 0.35, cap: 'round' });
    gr.circle(0, 0, g.reach).stroke({ color: 0xff2d55, width: 2 / Math.max(f.s.sc, 1), alpha: 0.6 });
  }

  set visible(v: boolean) {
    this.root.visible = v;
  }

  destroy(): void {
    this.root.destroy({ children: true });
  }
}
