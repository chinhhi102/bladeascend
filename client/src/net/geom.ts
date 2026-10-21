import type { Shape } from './protocol';

export interface Vec {
  x: number;
  y: number;
}

export function contains(s: Shape, x: number, y: number): boolean {
  if (s.c) {
    const r = s.r ?? 0;
    return (x - s.x) ** 2 + (y - s.y) ** 2 <= r * r;
  }
  return x >= s.x && x <= s.x + (s.w ?? 0) && y >= s.y && y <= s.y + (s.h ?? 0);
}

/** Mirrors world.Shape.PushOut on the server so prediction agrees near walls. */
export function pushOut(s: Shape, p: Vec, r: number): Vec | null {
  if (s.c) {
    const sr = s.r ?? 0;
    const dx = p.x - s.x;
    const dy = p.y - s.y;
    const d = Math.hypot(dx, dy);
    if (d >= sr + r) {
      return null;
    }
    const k = d > 1e-6 ? (sr + r) / d : 0;
    return d > 1e-6 ? { x: s.x + dx * k, y: s.y + dy * k } : { x: s.x + sr + r, y: s.y };
  }
  const w = s.w ?? 0;
  const h = s.h ?? 0;
  const cx = Math.max(s.x, Math.min(p.x, s.x + w));
  const cy = Math.max(s.y, Math.min(p.y, s.y + h));
  const dx = p.x - cx;
  const dy = p.y - cy;
  const d = Math.hypot(dx, dy);
  if (d >= r) {
    return null;
  }
  if (d > 1e-6) {
    return { x: cx + (dx * r) / d, y: cy + (dy * r) / d };
  }
  const left = p.x - s.x;
  const right = s.x + w - p.x;
  const top = p.y - s.y;
  const bottom = s.y + h - p.y;
  const m = Math.min(left, right, top, bottom);
  if (m === left) {
    return { x: s.x - r, y: p.y };
  }
  if (m === right) {
    return { x: s.x + w + r, y: p.y };
  }
  if (m === top) {
    return { x: p.x, y: s.y - r };
  }
  return { x: p.x, y: s.y + h + r };
}

export function lerp(a: number, b: number, t: number): number {
  return a + (b - a) * t;
}
