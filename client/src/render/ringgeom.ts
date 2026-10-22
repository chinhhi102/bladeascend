// Sword ring geometry of GAME_SPEC §4.1a — the same formulas as server/internal/game/sword.go,
// fed with numbers from /api/config so the drawing matches the server hitboxes.
import type { GameConfig } from '../config';

export interface RingGeom {
  base: number; // SWORD_LEN * scale * buff_long
  inner: number; // hilt radius of ring 0
  width: number;
  hiltStep: number;
  growth: number;
  compound: boolean;
  /** step: vòng thứ k chứa step × k kiếm (vòng 1 = 10, vòng 2 = 20… spec §4.1a). */
  step: number;
  rings: number;
  reach: number; // farthest tip at rest tilt + half width
}

/** ringCap: số kiếm vòng k chứa được (server RingGeom.RingCap). */
export function ringCap(g: RingGeom, k: number): number {
  return g.step * (k + 1);
}

/** ringStart: chỉ số kiếm đầu tiên của vòng k (server RingGeom.RingStart). */
export function ringStart(g: RingGeom, k: number): number {
  return (g.step * k * (k + 1)) / 2;
}

/** ringOf: sword i thuộc vòng nào (server RingGeom.RingOf). */
export function ringOf(g: RingGeom, i: number): number {
  let k = Math.floor((Math.sqrt(1 + (8 * i) / g.step) - 1) / 2);
  while (k > 0 && ringStart(g, k) > i) {
    k--;
  }
  while (ringStart(g, k + 1) <= i) {
    k++;
  }
  return k;
}

function ringsFor(step: number, n: number): number {
  let r = 1;
  while ((step * r * (r + 1)) / 2 < n) {
    r++;
  }
  return r;
}

export function ringGeom(cfg: GameConfig, scale: number, n: number, longMult: number): RingGeom {
  const c = cfg.sword;
  const step = Math.max(1, c.ring_step);
  const g: RingGeom = {
    base: c.length * scale * longMult,
    inner: (cfg.player.body_r + c.hilt_gap) * scale,
    width: c.width * scale,
    hiltStep: c.hilt_step,
    growth: c.ring_growth,
    compound: c.ring_growth_mode === 'compound',
    step,
    rings: ringsFor(step, n),
    reach: 0,
  };
  const tilt = (c.tilt_deg * Math.PI) / 180;
  for (let k = 0; k < g.rings; k++) {
    g.reach = Math.max(g.reach, tipDist(g, k, tilt));
  }
  g.reach += g.width / 2;
  return g;
}

export function hiltR(g: RingGeom, k: number): number {
  return g.inner + g.hiltStep * k * g.base;
}

export function bladeLen(g: RingGeom, k: number): number {
  if (g.compound) {
    return g.base * Math.pow(1 + g.growth, k) - g.hiltStep * k * g.base;
  }
  return g.base * (1 + (g.growth - g.hiltStep) * k);
}

export function tipDist(g: RingGeom, k: number, tilt: number): number {
  const h = hiltR(g, k);
  const l = bladeLen(g, k);
  return Math.sqrt(h * h + l * l + 2 * h * l * Math.cos(tilt));
}

/** slotAngle: angle of sword i relative to the ring phase (spec §4.1a, server SlotAngle). */
export function slotAngle(g: RingGeom, n: number, i: number): number {
  const k = ringOf(g, i);
  const cap = ringCap(g, k);
  const start = ringStart(g, k);
  const m = Math.min(cap, n - start);
  if (m <= 0) {
    return 0;
  }
  const slot = Math.round(((i - start) * cap) / m);
  const step = (2 * Math.PI) / cap;
  return slot * step + (k * step) / Math.max(g.rings, 1);
}

/** restTilt: blades trail behind the spin (server RestTilt). */
export function restTilt(cfg: GameConfig, dir: number): number {
  return (-dir * cfg.sword.tilt_deg * Math.PI) / 180;
}
