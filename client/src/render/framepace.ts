import type { Ticker } from 'pixi.js';

/**
 * Nhịp khung hình đều (chủ dự án 05/11: "di chuyển góc quay hơi giật").
 *
 * Trước đây máy tính đặt thẳng `ticker.maxFPS = 60`. Pixi kiểm trần bằng
 * `(now - last) | 0 < 1000/60`: khoảng rAF của màn 60 Hz dao động 16,5–16,8 ms, làm tròn
 * xuống thành 16 < 16,67 → BỎ khung. Đo bằng `tools/browsertest/wallhug.mjs`: 2–3 % khung
 * màn hình đứng im, độ giật máy quay 0,20; bỏ trần thì 0 % và 0,10. Trên màn 120/144 Hz trần
 * của Pixi giữ đúng TRUNG BÌNH 60 khung/giây nhưng bằng cách vẽ cách 1 rồi cách 2 khung quét
 * xen kẽ (nó mang phần dư sang khung sau) — chính là kiểu giật "lúc nhanh lúc chậm".
 *
 * Cách làm: đo tần số quét thật, chọn `n` = số khung quét giữa hai lần vẽ sao cho nhịp gần
 * trần nhất, rồi TỰ đếm khung quét: đúng mỗi n khung thì vẽ một lần. n = 1 thì để Pixi vẽ
 * theo màn hình, không trần.
 */

const SAMPLE_FRAMES = 30;

/** frameDivider: vẽ một lần mỗi bao nhiêu khung quét (≥ 1). refreshHz không đo được → 1. */
export function frameDivider(refreshHz: number, cap: number): number {
  if (!(refreshHz > 0) || !(cap > 0)) {
    return 1;
  }
  return Math.max(1, Math.round(refreshHz / cap));
}

/** measureRefreshHz đo tần số quét bằng trung vị khoảng cách giữa các rAF. */
export function measureRefreshHz(): Promise<number> {
  return new Promise((resolve) => {
    const gaps: number[] = [];
    let last = 0;
    const tick = (t: number): void => {
      if (last > 0) {
        gaps.push(t - last);
      }
      last = t;
      if (gaps.length < SAMPLE_FRAMES) {
        requestAnimationFrame(tick);
        return;
      }
      gaps.sort((a, b) => a - b);
      const mid = gaps[gaps.length >> 1];
      resolve(mid > 0 ? 1000 / mid : 0);
    };
    requestAnimationFrame(tick);
  });
}

/**
 * paceTicker: n = 1 → bỏ trần, Pixi tự chạy theo rAF. n > 1 → dừng vòng rAF của Pixi và
 * gọi `ticker.update` đúng mỗi n khung quét (deltaMS vẫn là thời gian thật, nên chuyển động
 * không đổi tốc độ — chỉ đổi số lần vẽ).
 */
export function paceTicker(ticker: Ticker, n: number): void {
  ticker.maxFPS = 0;
  if (n <= 1) {
    return;
  }
  ticker.stop();
  let count = 0;
  const loop = (t: number): void => {
    count++;
    if (count >= n) {
      count = 0;
      ticker.update(t);
    }
    requestAnimationFrame(loop);
  };
  requestAnimationFrame(loop);
}
