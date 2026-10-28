/**
 * Mức chất lượng đồ họa (spec §3.6): Cao / Vừa / Thấp. Người chơi đổi bằng phím V, lựa chọn
 * lưu trong localStorage.
 *
 * Thần thú giờ luôn bật (spec v4.7 §3.5) nên số đốt phải co theo máy, không thể để cố định.
 */

export type Quality = 'high' | 'medium' | 'low';

const KEY = 'kx.quality';
const ORDER: Quality[] = ['high', 'medium', 'low'];

export const QUALITY_NAME: Record<Quality, string> = { high: 'Cao', medium: 'Vừa', low: 'Thấp' };

/** Số đốt thần thú theo chất lượng (spec §3.5: Cao 16 · Vừa 10 · Thấp 6). */
export const BEAST_SEGMENTS: Record<Quality, number> = { high: 16, medium: 10, low: 6 };

/** Tỉ lệ hạt vệt theo sau (Thấp = tắt hẳn). */
export const TRAIL_RATE: Record<Quality, number> = { high: 1, medium: 0.5, low: 0 };

let current: Quality = read();

function read(): Quality {
  const v = localStorage.getItem(KEY);
  return ORDER.includes(v as Quality) ? (v as Quality) : 'high';
}

export function quality(): Quality {
  return current;
}

export function setQuality(q: Quality): void {
  current = q;
  localStorage.setItem(KEY, q);
}

/** cycleQuality đổi sang mức tiếp theo và trả về mức mới. */
export function cycleQuality(): Quality {
  setQuality(ORDER[(ORDER.indexOf(current) + 1) % ORDER.length]);
  return current;
}
