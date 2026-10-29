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
  // Hộp bao lọc trước, y như `world.Shape.Contains` bên server.
  if (x < s.x || x > s.x + (s.w ?? 0) || y < s.y || y > s.y + (s.h ?? 0)) {
    return false;
  }
  return isPoly(s) ? containsPoly(s, x, y) : true;
}

/** Vùng địa hình hữu cơ (spec §8.6) là ĐA GIÁC, không phải hình chữ nhật. */
function isPoly(s: Shape): boolean {
  return (s.p?.length ?? 0) >= 6;
}

/**
 * containsPoly: ray casting kinh điển — bắn một tia sang phải rồi đếm số cạnh nó cắt, lẻ
 * là trong, chẵn là ngoài. Bản port đúng từng dòng của `Shape.containsPoly` bên server.
 *
 * Thiếu hàm này, client coi mọi vùng đa giác là **cả hộp bao** của nó: đi ngang một vũng
 * bùn hình méo thì client tưởng mình đang lội bùn (hoặc ngược lại) trong khi server nghĩ
 * khác, nên tốc độ hai bên lệch nhau và server kéo nhân vật về mỗi tick. Đo thật trước khi
 * sửa: đi thẳng giữa đồng trống vẫn ra **20,2 lần hiệu chỉnh/giây**.
 */
function containsPoly(s: Shape, x: number, y: number): boolean {
  const pts = s.p as number[];
  let inside = false;
  const n = pts.length / 2;
  for (let i = 0, j = n - 1; i < n; j = i, i++) {
    const xi = pts[i * 2];
    const yi = pts[i * 2 + 1];
    const xj = pts[j * 2];
    const yj = pts[j * 2 + 1];
    if (yi > y !== yj > y && x < ((xj - xi) * (y - yi)) / (yj - yi) + xi) {
      inside = !inside;
    }
  }
  return inside;
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
