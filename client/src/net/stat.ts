// Luật chung spec v6.32 §4.0: mọi thưởng/phạt CỘNG trên base, không nhân dồn.
//
//   giá_trị = base × clamp(1 + Σ khoản, sàn, trần)
//
// Bản sao 1–1 của `Stat` bên server (server/internal/game/stat.go). Hai bên cùng chạy bộ số
// mẫu `server/internal/game/testdata/stat_vectors.json` (`npm run check:stat`) — client dự
// đoán tốc chạy và góc vòng kiếm bằng hàm này, lệch là bị server kéo về mỗi tick.
import type { GameConfig, InnateCfg } from '../config';

export interface StatOpts {
  floor?: number;
  cap?: number;
}

/** base × clamp(1 + Σ terms, floor, cap). 0/thiếu = không kẹp phía đó. */
export function stat(base: number, opts: StatOpts, ...terms: number[]): number {
  let f = 1;
  for (const t of terms) {
    f += t;
  }
  if (opts.floor && opts.floor > 0 && f < opts.floor) {
    f = opts.floor;
  }
  if (opts.cap && opts.cap > 0 && f > opts.cap) {
    f = opts.cap;
  }
  return base * Math.max(0, f);
}

/** Hệ số nhân cũ `× k` → khoản cộng `k − 1` (spec §4.0). */
export function bonus(mult: number | undefined): number {
  return mult && mult > 0 ? mult - 1 : 0;
}

const ZERO: InnateCfg = { speed: 0, pickup: 0, spin: 0, durability: 0 };

/** Nội tại của nhân vật (spec v6.33 §6.4); nhân vật lạ / đàn em → 0. */
export function innate(cfg: GameConfig, char: string): InnateCfg {
  return cfg.innate?.[char] ?? ZERO;
}
