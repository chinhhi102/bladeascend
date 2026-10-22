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
  perRing: number;
  rings: number;
  reach: number; // farthest tip at rest tilt + half width
}

export function ringGeom(cfg: GameConfig, scale: number, n: number, longMult: number): RingGeom {
  const c = cfg.sword;
  const perRing = Math.max(1, c.swords_per_ring);
  const g: RingGeom = {
    base: c.length * scale * longMult,
    inner: (cfg.player.body_r + c.hilt_gap) * scale,
    width: c.width * scale,
    hiltStep: c.hilt_step,
    growth: c.ring_growth,
    compound: c.ring_growth_mode === 'compound',
    perRing,
    rings: Math.max(1, Math.ceil(n / perRing)),
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
export function slotAngle(per: number, n: number, i: number): number {
  if (n <= per) {
    return (2 * Math.PI * i) / n;
  }
  const k = Math.floor(i / per);
  const j = i % per;
  const m = Math.min(per, n - k * per);
  const slot = Math.round((j * per) / m);
  const step = (2 * Math.PI) / per;
  const rings = Math.ceil(n / per);
  return slot * step + (k * step) / rings;
}

/** restTilt: blades trail behind the spin (server RestTilt). */
export function restTilt(cfg: GameConfig, dir: number): number {
  return (-dir * cfg.sword.tilt_deg * Math.PI) / 180;
}
