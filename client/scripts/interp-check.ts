/**
 * Bài đơn vị cho `followSnaps` — nội suy entity người khác / quái (Q48).
 *
 *   cd client && npm run check:interp
 *
 * ── Vì sao có bài này ────────────────────────────────────────────────────────────────────
 *
 * Q48 đo được bằng trình duyệt: nhìn người chạy 975 đv/s từ tab khác thì thấy đứng hình rồi
 * nhảy cóc. Hai lần tôi sửa thẳng trong `state.ts` rồi thử bằng trình duyệt, cả hai lần đều
 * ra bản TỆ HƠN bản cũ (16,5% khung nhảy cóc, rồi entity đứng im hẳn) — vì mỗi vòng thử mất
 * ~10 phút và không nhìn thấy trạng thái bên trong.
 *
 * Nên: bơm một chuỗi gói GIẢ có đúng những thứ mạng thật có, rồi canh quỹ đạo vẽ ra.
 *
 * Bốn thứ bơm vào, theo đúng cái đo được ở `gomdon-smooth.mjs` ngày 10/11:
 *   - **dao động ±30 ms** quanh nhịp gửi (đo thật: khoảng cách gói 33 và 67 ms xen kẽ, vì
 *     chu kỳ gửi 50 ms không chia hết cho nhịp tick 33,3 ms);
 *   - **mất 5% gói**;
 *   - **đồng hồ trôi 2%** (client và server không chạy cùng tốc độ);
 *   - **đổi tốc ×1 ↔ ×5** giữa chừng (Gom Đơn bật/tắt).
 *
 * Hai điều kiện ĐẠT:
 *   1. không khung nào vẽ nhanh hơn trần tốc (nhảy cóc);
 *   2. không khung nào đứng im trong lúc entity đang chạy (đứng hình).
 *
 * ĐANG ĐỎ Ở ĐIỀU KIỆN 2 — cố ý. `followSnaps` hiện là bản nội suy thuần, kho cạn thì giữ
 * nguyên gói cuối. Q48 bước 2–5 phải làm cho bài này xanh. **Chưa gắn vào `npm run build`**
 * vì nó đỏ; gắn vào là chặn mọi lần deploy.
 */
import { Clock, followSnaps, type Timed, type Track } from '../src/net/interp.ts';

const BASE = 195;          // player.move_base
const FAST = BASE * 5;     // Gom Đơn ×5 = 975 đv/s
const SEND_MS = 50;        // net.snapshot_hz 20
const JITTER_MS = 30;
const LOSS = 0.05;
const CLOCK_DRIFT = 0.02;
const INTERP_MS = 150;     // net.interp_ms
const FRAME_MS = 1000 / 60;
const SECONDS = 20;

/** Sinh ngẫu nhiên lặp lại được — bài kiểm phải ra cùng kết quả mỗi lần chạy. */
function rng(seed: number): () => number {
  let s = seed >>> 0;
  return () => {
    s = (s * 1664525 + 1013904223) >>> 0;
    return s / 4294967296;
  };
}

interface P { x: number; y: number }

/** Đường đi THẬT của entity trên server: chạy ngang, đổi tốc ×1 ↔ ×5 mỗi 2 giây. */
function truth(t: number): { at: P; v: number } {
  let x = 0;
  let v = BASE;
  const step = 0.001;
  for (let u = 0; u < t; u += step) {
    v = Math.floor(u / 2) % 2 === 0 ? BASE : FAST;
    x += v * step;
  }
  return { at: { x, y: 0 }, v };
}

const rand = rng(12345);

// --- bơm gói -----------------------------------------------------------------------------
//
// Mô hình: đồng hồ MÁY chạy lệch `CLOCK_DRIFT` so với đồng hồ SERVER. Gói mang dấu thời
// gian server, tới máy sau `LATENCY` kèm dao động. Client chỉ biết giờ máy và dấu thời gian
// trong gói — đúng như thật.
const LATENCY = 0.04;
interface Pkt { tm: number; at: number; s: P }   // tm = giờ server, at = giờ máy lúc tới
const pkts: Pkt[] = [];
for (let ts = 0; ts <= SECONDS; ts += SEND_MS / 1000) {
  if (rand() < LOSS) continue;
  const jitter = ((rand() * 2 - 1) * JITTER_MS) / 1000;
  const arriveServer = ts + LATENCY + jitter;
  pkts.push({ tm: ts, at: arriveServer / (1 + CLOCK_DRIFT), s: truth(ts).at });
}
pkts.sort((a, b) => a.at - b.at);

// --- vẽ từng khung -------------------------------------------------------------------------
// `OLD_CLOCK=1` dựng lại đồng hồ CŨ (lấy ước tính tốt nhất, nhảy thẳng, lùi 0,5 ms mỗi
// gói) để chứng minh chính `Clock` làm bài này xanh, chứ không phải tôi vừa đổi mô hình gói.
const OLD = process.env.OLD_CLOCK === '1';
const clock = new Clock({ rate: 0.08, windowSecs: 2 });
let oldOff: number | null = null;
const oldSample = (est: number): void => {
  oldOff = oldOff === null || est > oldOff ? est : oldOff - 0.0005;
};
const buf: Timed<P>[] = [];
const tr: Track = {};
let prev: P | null = null;
let jumps = 0;
let stalls = 0;
let worst = 0;
let movingFrames = 0;
let next = 0;
for (let i = 0; ; i++) {
  const local = (i * FRAME_MS) / 1000;               // giờ MÁY
  if (local * (1 + CLOCK_DRIFT) > SECONDS) break;
  while (next < pkts.length && pkts[next].at <= local) {
    const p = pkts[next++];
    clock.sample(p.tm - p.at, p.at);
    oldSample(p.tm - p.at);
    buf.push({ t: p.tm, s: p.s });
  }
  clock.step(FRAME_MS / 1000, local);
  if (buf.length < 2 || (OLD ? oldOff === null : !clock.has())) continue;
  const rt = (OLD ? local + (oldOff ?? 0) : clock.now(local)) - INTERP_MS / 1000;
  const at = followSnaps(buf, rt, tr, { jumpUnits: 600 });
  if (prev) {
    const v = Math.hypot(at.x - prev.x, at.y - prev.y) / (FRAME_MS / 1000);
    worst = Math.max(worst, v);
    if (v > FAST * 1.5) jumps++;
    if (truth(Math.max(0, rt)).v > BASE * 1.5) {
      movingFrames++;
      if (v < 1) stalls++;
    }
  }
  prev = at;
}

const rows = [
  ['đồng hồ', OLD ? 'CŨ (nhảy thẳng)' : 'MỚI (chỉnh tốc độ)'],
  ['gói gửi', String(pkts.length)],
  ['khung vẽ có mẫu', String(movingFrames)],
  ['nhảy cóc (vẽ > 1,5× trần)', String(jumps)],
  ['đứng hình lúc đang chạy nhanh', String(stalls)],
  ['tốc vẽ lớn nhất (đv/s)', worst.toFixed(0)],
  ['trần tốc (đv/s)', String(FAST)],
];
for (const [k, v] of rows) {
  console.log(`${k.padEnd(32)} ${v}`);
}

const ok = jumps === 0 && stalls === 0;
console.log(ok
  ? 'ĐẠT: quỹ đạo vẽ không vượt trần tốc và không đứng hình'
  : `TRƯỢT: ${jumps} khung nhảy cóc, ${stalls} khung đứng hình — Q48 bước 2–5 phải chữa chỗ này`);
process.exit(ok ? 0 : 1);
