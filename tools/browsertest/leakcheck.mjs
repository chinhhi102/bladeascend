// Bài kiểm RÒ RỈ phía CLIENT — GAME_SPEC §13.9 (v6.57).
//
//   cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true -seed 20261123
//   node tools/browsertest/leakcheck.mjs http://127.0.0.1:9998
//
// Spec v6.57 chốt bốn điều, và cả bốn đều là PHÉP ĐO chứ không phải luật chơi:
//
//   1. KỊCH BẢN CỐ ĐỊNH — cùng seed, cùng chuỗi hành động, cùng số lượt đổi khu (50) và
//      chết/vào lại (10), kết thúc ở cùng một khu.
//   2. 60 PHÚT — phút 0–10 là khởi động (không tính), cửa sổ đo phút 10–60, mẫu mỗi 15 s.
//   3. XÉT BẰNG DỐC CÓ KHOẢNG TIN CẬY — chia cửa sổ thành 5 khối 10 phút, mỗi khối fit một
//      dốc OLS, rồi lấy KTC 95% của dốc trung bình (t, df = 4). Bỏ hẳn kiểu so hai mẫu và
//      ngưỡng "tăng > 5%" của v6.54.
//   4. MỐC KỶ LỤC cho mức nền — build mới vượt mức nền thấp nhất từng đo > 10% là ĐỎ.
//
// ── Vì sao v6.54 phải bỏ ─────────────────────────────────────────────────────────────────
//
// Sáu lượt đo cùng kịch bản, cùng build, máy rảnh cho dốc heap 0,19 · 0,25 · 0,61 · 0,92 ·
// 1,17 · 2,00 MB/phút — chênh nhau HƠN MƯỜI LẦN. Nhiễu máy ±13–15 MB không giải thích nổi
// khoảng đó. Thứ giải thích được: mỗi lượt chơi một kiểu. Số lượt đổi khu khác nhau, khu
// kết thúc khác nhau, mật độ người/quái khác nhau — mà chính những thứ đó quyết định số
// đối tượng đang sống, tức quyết định con số đang đo.
//
// Nên bài này không còn "đi lang thang 30 phút rồi so đầu với cuối". Nó chạy một CHU KỲ
// lặp lại chính xác tới từng mốc 15 giây:
//
//   • một chu kỳ = 60 giây = 4 mẫu. Đầu chu kỳ đổi sang khu kế tiếp (0→1→2→3→4→0…), phần
//     còn lại đi lại theo một chuỗi phím sinh từ hạt giống cố định.
//   • cứ 5 chu kỳ (5 phút) thì có một lượt chết + vào lại, luôn rơi vào cùng một khu.
//   • 60 chu kỳ = 60 phút. Cửa sổ đo (phút 10–60) chứa ĐÚNG 50 lượt đổi khu và 10 lượt
//     chết, đúng hai con số của spec; 10 chu kỳ khởi động nằm ngoài cửa sổ nên không tính.
//   • 5 khối × 10 phút = 10 chu kỳ/khối = 2 chu kỳ-chết/khối, nên mọi khối có cùng thành
//     phần hành động. Không khối nào "nặng" hơn khối nào vì rơi vào pha khác của kịch bản.
//
// Và thế giới cũng phải cố định: server nhận `-seed`, nếu không thì vị trí viên, quái và
// bot đổi theo từng lần chạy.
//
// ── Khác gì `texleak.mjs` ────────────────────────────────────────────────────────────────
//
// `texleak` chỉ đếm TEXTURE và chỉ đi lại. Bài này đếm đủ các chỉ số của spec và chạy đúng
// ba việc mà spec chỉ đích danh — đổi khu, chết, vào lại — vì đó là ba đường tạo và huỷ
// hàng loạt đối tượng, và cũng là ba đường từng rò rỉ thật (181 lượt đổi khu → 1,4 GB).
import { readFileSync, writeFileSync, renameSync, mkdirSync } from 'node:fs';
import { cpus, loadavg } from 'node:os';
import { dirname, join } from 'node:path';
import { launch } from './browser.mjs';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
/** MINUTES: tổng thời gian chơi. Spec v6.57 chốt 60; hạ xuống khi chỉ muốn thử bài đo còn chạy. */
const MINUTES = Number(process.env.MINUTES ?? 60);
/** WARMUP_MIN: phút 0–10 là khởi động, KHÔNG tính (v6.57). */
const WARMUP_MIN = Number(process.env.WARMUP_MIN ?? 10);
/** STEP_SECS: nhịp lấy mẫu, spec chốt 15 s. */
const STEP_SECS = Number(process.env.STEP_SECS ?? 15);
/** BLOCKS: số khối chia cửa sổ, spec chốt 5 (mỗi khối 10 phút). */
const BLOCKS = Number(process.env.BLOCKS ?? 5);
/** RISE_PCT: dốc × 60 phút vượt ngần này % mức nền thì ĐỎ (spec chốt 10). */
const RISE_PCT = Number(process.env.RISE_PCT ?? 10);
/** RATCHET_PCT: mức nền vượt mốc kỷ lục ngần này % thì ĐỎ (spec chốt 10). */
const RATCHET_PCT = Number(process.env.RATCHET_PCT ?? 10);
/**
 * STATE_DIR: nơi giữ MỐC KỶ LỤC và SỔ ĐÊM. Rỗng = không đọc, không ghi.
 *
 * Chỉ bài hằng đêm (`tools/leaknightly.sh`) truyền vào. Một lượt chạy tay — thường là bài
 * 6 phút để xem bài đo còn sống — không được phép hạ mốc kỷ lục hay ghi thêm một đêm vào
 * chuỗi dấu, vì cả hai thứ đó là trí nhớ dài hạn của cổng gác.
 */
const STATE_DIR = process.env.STATE_DIR ?? '';
/** JSON_OUT: ghi toàn bộ số liệu lượt này ra tệp (trang quản trị đọc). */
const JSON_OUT = process.env.JSON_OUT ?? (STATE_DIR ? join(STATE_DIR, 'leak-client.json') : '');
/** BUILD: mã build đang đo, chỉ để ghi vào sổ. */
const BUILD = process.env.BUILD ?? '';
/**
 * MAX_LOAD / IGNORE_LOAD: luật 16 của CLAUDE.md — máy bận thì KHÔNG kết luận.
 *
 * Bài này đo bộ nhớ chứ không đo thời gian, nên nghe như ít nhạy với tải. Không đúng: máy
 * quá tải kéo FPS trình duyệt xuống dưới ngưỡng tự hạ chất lượng, mà hạ chất lượng ĐỔI
 * chính những thứ đang đếm (bớt bộ lọc, bớt texture, bớt sprite). Ngày 23/11 một lượt đo
 * rơi từ 55 FPS xuống 20 FPS giữa chừng vì việc của dự án khác chiếm 690% CPU.
 *
 * Đo bằng `loadavg` chứ không bằng thời gian khung của chính tab đang đo — tab headless vẽ
 * nhẹ nên khung vẫn đều trong khi máy đã chìm (bài học của `gomdon-smooth`, 10/11).
 */
const MAX_LOAD = Number(process.env.MAX_LOAD ?? 1.5);
const IGNORE_LOAD = process.env.IGNORE_LOAD === '1';
/** BUSY_TOL: tỉ lệ mẫu được phép bận trong cửa sổ trước khi bài đo từ chối kết luận. */
const BUSY_TOL = Number(process.env.BUSY_TOL ?? 0.1);

const sleep = (ms) => new Promise((r) => setTimeout(r, Math.max(0, ms)));

// ── Kịch bản cố định ────────────────────────────────────────────────────────────────────
const ZONES = 5;
/** Số mẫu trong một chu kỳ 60 giây. */
const PER_CYCLE = Math.max(1, Math.round(60 / STEP_SECS));
/** DEATH_EVERY: cứ ngần này chu kỳ thì một lượt chết → 10 lượt trong cửa sổ 50 phút. */
const DEATH_EVERY = 5;
/**
 * ZONE_MODE: `cycle` = đổi khu mỗi chu kỳ (kịch bản chuẩn, 50 lượt trong cửa sổ) ·
 * `none` = KHÔNG đổi khu lần nào sau lượt neo đầu tiên.
 *
 * `none` là bên thứ hai của phép so "heap có tăng theo SỐ LƯỢT ĐỔI KHU không" (B38). Spec
 * v6.57 đòi ≥ 3 lượt mỗi bên, và hai bên phải khác nhau ĐÚNG MỘT THỨ — nên nó là một công
 * tắc của chính kịch bản này, không phải một bài đo riêng.
 */
const ZONE_MODE = process.env.ZONE_MODE === 'none' ? 'none' : 'cycle';
/**
 * SCENARIO_VERSION: tên phiên bản kịch bản, ghi kèm mốc kỷ lục và sổ đêm (v6.58).
 *
 * Mốc kỷ lục chỉ có nghĩa trong phạm vi MỘT kịch bản. Lỗi `dbg('zone')` thiếu tham số cho
 * thấy rõ vì sao: sáu lượt đo cũ chạy trên một kịch bản gần như không đổi khu, mức nền của
 * nó thấp hơn hẳn — giữ mốc đó lại thì mọi lượt của kịch bản thật đỏ ngay, mà đỏ vì lý do
 * sai. Đổi kịch bản → mốc cũ hết giá trị, chuỗi dấu bắt đầu lại.
 */
const SCENARIO_VERSION = `v6.58|win${WARMUP_MIN}-${MINUTES}|step${STEP_SECS}s|zone-${ZONE_MODE}|die1of${DEATH_EVERY}`;
const totalSamples = Math.round((MINUTES * 60) / STEP_SECS);
const totalCycles = Math.floor(totalSamples / PER_CYCLE);

/**
 * mulberry32: nguồn ngẫu nhiên GIEO ĐƯỢC, để chuỗi phím đi lại giống hệt nhau giữa các lượt.
 *
 * `Math.random()` thì mỗi lượt một đường đi khác, mà đường đi quyết định gặp bao nhiêu
 * người và bao nhiêu quái — tức quyết định chính con số đang đo.
 */
function mulberry32(a) {
  return () => {
    a = (a + 0x6d2b79f5) | 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

/** zoneAtCycle: khu mà chu kỳ thứ c nhảy tới. Chu kỳ cuối (59) → khu 0, luôn luôn. */
const zoneAtCycle = (c) => (c + 1) % ZONES;

// ── Thống kê ────────────────────────────────────────────────────────────────────────────

/** ols: dốc bình phương tối thiểu của (x, y), kèm hệ số chặn. */
function ols(pts) {
  const n = pts.length;
  if (n < 2) {
    return { slope: 0, intercept: n ? pts[0][1] : 0, n };
  }
  const mx = pts.reduce((a, [x]) => a + x, 0) / n;
  const my = pts.reduce((a, [, y]) => a + y, 0) / n;
  const den = pts.reduce((a, [x]) => a + (x - mx) ** 2, 0);
  const slope = den > 0 ? pts.reduce((a, [x, y]) => a + (x - mx) * (y - my), 0) / den : 0;
  return { slope, intercept: my - slope * mx, n };
}

/**
 * T95: điểm tới hạn t hai phía, mức 95%, theo bậc tự do.
 *
 * df = 4 (năm khối) là ca chính của spec → 2,776. Bảng này viết thẳng ra thay vì gọi một
 * thư viện thống kê: thêm một phụ thuộc cho một hằng số là cái giá không đáng.
 */
const T95 = { 1: 12.706, 2: 4.303, 3: 3.182, 4: 2.776, 5: 2.571, 6: 2.447, 7: 2.365, 8: 2.306, 9: 2.262, 10: 2.228, 11: 2.201, 12: 2.179, 13: 2.160, 14: 2.145, 15: 2.131, 16: 2.120, 17: 2.110, 18: 2.101, 19: 2.093, 20: 2.086 };
const tcrit = (df) => (df <= 0 ? Infinity : (T95[df] ?? (df >= 60 ? 2.0 : 2.042)));

/**
 * blockTrend: dốc theo 5 khối + KTC 95% của dốc trung bình — phép xét CHẶN của v6.57.
 *
 * Vì sao không fit một dốc duy nhất trên cả 200 mẫu: sai số chuẩn của một OLS giả định các
 * mẫu độc lập, mà heap là tín hiệu RĂNG CƯA — hai mẫu cách nhau 15 giây tương quan rất
 * mạnh, nên KTC tính kiểu đó hẹp giả tạo và cổng sẽ đỏ với mọi build. Năm dốc khối là năm
 * quan sát gần như độc lập (mỗi khối 10 phút, chứa 2 chu kỳ-chết trọn vẹn), và độ phân tán
 * GIỮA các khối mới là thước đo thật của việc "lượt này có đang dốc lên không".
 */
function blockTrend(rows, get, winFrom, winTo) {
  const span = (winTo - winFrom) / BLOCKS;
  const slopes = [];
  for (let b = 0; b < BLOCKS; b++) {
    const lo = winFrom + b * span;
    const hi = lo + span;
    // Mép phải của khối cuối phải lấy cả mẫu đúng phút 60, nên dùng `<=` ở khối cuối.
    const pts = rows
      .filter((r) => r.min >= lo && (b === BLOCKS - 1 ? r.min <= hi : r.min < hi))
      .map((r) => [r.min, get(r)]);
    slopes.push(ols(pts).slope);
  }
  const n = slopes.length;
  const mean = slopes.reduce((a, v) => a + v, 0) / n;
  const sd = n > 1 ? Math.sqrt(slopes.reduce((a, v) => a + (v - mean) ** 2, 0) / (n - 1)) : 0;
  const half = tcrit(n - 1) * (sd / Math.sqrt(n));
  return { slopes, mean, sd, lo: mean - half, hi: mean + half };
}

/**
 * baselineAt: mức nền tại phút `at`, đọc từ ĐƯỜNG FIT của khối đầu, không từ một mẫu.
 *
 * Spec nói "giá trị tại phút 10". Lấy đúng mẫu ở phút 10 thì mốc kỷ lục phụ thuộc vào việc
 * bộ dọn rác vừa chạy hay chưa — ba đêm liên tiếp từng cho mẫu phút 20 lệch nhau 22 MB
 * trong khi mức thật không đổi. Fit qua cả khối đầu (40 mẫu) rồi đọc giá trị ở phút 10 là
 * cùng một con số nhưng không phụ thuộc chỗ rơi của một mẫu.
 */
function baselineAt(rows, get, at, span) {
  const pts = rows.filter((r) => r.min >= at && r.min <= at + span).map((r) => [r.min, get(r)]);
  const { slope, intercept } = ols(pts);
  return intercept + slope * at;
}

/** singleTrend: một dốc OLS + KTC 95% từ sai số dư — dùng cho sổ tài nguyên (ít điểm). */
function singleTrend(pts) {
  const n = pts.length;
  const { slope, intercept } = ols(pts);
  if (n < 3) {
    return { slope, lo: -Infinity, hi: Infinity, n };
  }
  const mx = pts.reduce((a, [x]) => a + x, 0) / n;
  const sxx = pts.reduce((a, [x]) => a + (x - mx) ** 2, 0);
  const sse = pts.reduce((a, [x, y]) => a + (y - (intercept + slope * x)) ** 2, 0);
  const se = sxx > 0 ? Math.sqrt(sse / (n - 2) / sxx) : Infinity;
  const half = tcrit(n - 2) * se;
  return { slope, lo: slope - half, hi: slope + half, n };
}

// ── Trí nhớ dài hạn: mốc kỷ lục + sổ đêm ────────────────────────────────────────────────

const readJSON = (p, dflt) => {
  try {
    return JSON.parse(readFileSync(p, 'utf8'));
  } catch {
    return dflt;
  }
};

/** writeJSONAtomic: ghi tệp tạm rồi đổi tên — server và deploy.sh có thể đọc đúng lúc này. */
function writeJSONAtomic(p, obj) {
  mkdirSync(dirname(p), { recursive: true });
  writeFileSync(`${p}.tmp`, `${JSON.stringify(obj, null, 2)}\n`);
  renameSync(`${p}.tmp`, p);
}

/**
 * ratchetStep: mốc kỷ lục cho mức nền (§13.9 v6.57).
 *
 * Mốc CHỈ đi xuống, và hạ mốc cần 2 đêm xác nhận. Một đêm thấp bất thường (máy rảnh khác
 * thường, một khu vắng người) không được phép hạ mốc — hạ rồi thì đêm sau mọi build đều
 * đỏ, và cổng sẽ bị tắt thay vì được sửa. Giữ lại `pending` chính là chỗ chờ xác nhận đó.
 *
 * Khi hạ, lấy mức CAO HƠN trong hai đêm xác nhận: mốc là thứ build phải sống dưới nó, nên
 * chọn bên thận trọng.
 */
function ratchetStep(state, value) {
  const prev = state.record ?? null;
  if (prev === null) {
    return { record: value, pending: null, first: true, exceeded: 0 };
  }
  const exceeded = ((value - prev) / prev) * 100;
  if (value >= prev) {
    // Không xuống thì không có gì để xác nhận; xoá lần chờ dở dang.
    return { record: prev, pending: null, first: false, exceeded };
  }
  const pend = state.pending;
  if (!pend) {
    return { record: prev, pending: { value, nights: 1 }, first: false, exceeded };
  }
  const merged = Math.max(pend.value, value);
  if (pend.nights + 1 >= 2) {
    return { record: merged, pending: null, first: false, exceeded, lowered: true };
  }
  return { record: prev, pending: { value: merged, nights: pend.nights + 1 }, first: false, exceeded };
}

// ── Chạy ────────────────────────────────────────────────────────────────────────────────

const b = await launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist',
    // Không có cờ này thì `performance.memory` làm tròn thô (bậc 5 MB) và một mức rò rỉ
    // vài trăm KB/phút biến mất hoàn toàn trong bước làm tròn.
    '--enable-precise-memory-info'],
});
const ctx = await b.newContext({ viewport: { width: 1280, height: 720 }, locale: 'vi-VN' });
const p = await ctx.newPage();
const pageErrors = [];
p.on('pageerror', (e) => pageErrors.push(String(e).slice(0, 200)));
// Ép dọn rác trước mỗi lần đọc heap.
//
// `performance.memory.usedJSHeapSize` là lượng heap ĐANG CẤP PHÁT, gồm cả rác chưa dọn.
// Đọc thẳng thì con số nhảy 40 MB tuỳ lúc bộ dọn rác vừa chạy hay chưa, và một mức rò rỉ
// thật 200 KB/phút chìm nghỉm trong biên độ đó. `HeapProfiler.collectGarbage` của CDP là
// cách duy nhất ép dọn thật từ ngoài trang.
const cdp = await ctx.newCDPSession(p);
await cdp.send('HeapProfiler.enable').catch(() => undefined);
const gc = async () => {
  await cdp.send('HeapProfiler.collectGarbage').catch(() => undefined);
  await sleep(250);
};

// `?textrace=1`: ghi NƠI CẤP PHÁT của từng texture (render/gpumem.ts). Bật mặc định —
// khi bài đo đỏ, câu hỏi tiếp theo luôn là "nhóm nào phình", và chạy lại 60 phút chỉ để
// biết điều đó là một giờ mất không.
await p.goto(`${TARGET}/?debug=true&textrace=1`, { waitUntil: 'domcontentloaded' });
await p.waitForSelector('input', { timeout: 30000 });
await p.fill('input', 'KiemTraRoRi');
await p.locator('.char-card').nth(0).click();
await p.locator('.home .btn.primary').click();
await p.waitForSelector('.room-board', { timeout: 60000 });
await sleep(4000);

/** dbg gửi một lệnh test qua chính kết nối của game (chỉ chạy khi server cấp quyền debug). */
async function dbg(cmd, n = 0) {
  await p.evaluate(([c, k]) => window.kx?.conn?.send({ t: 'dbg', cmd: c, n: k }), [cmd, n]);
}

/** revive: đang ở màn Tổng kết thì bấm "Chơi tiếp". */
async function revive() {
  if (!(await p.$('.summary'))) {
    return false;
  }
  await p.locator('.summary .btn.primary').click();
  await p.waitForSelector('.room-board', { timeout: 30000 });
  await sleep(2000);
  return true;
}

/**
 * wanderUntil: đi lại tới đúng mốc thời gian `deadline`, theo chuỗi phím GIEO TỪ `seed`.
 *
 * Đứng yên thì không đo được gì (không ai vào/ra tầm nhìn), nhưng đi lại ngẫu nhiên thì
 * mỗi lượt một kịch bản. Gieo theo chỉ số mẫu nên lượt nào cũng đi đúng đường đó.
 */
const KEYS = ['KeyW', 'KeyA', 'KeyS', 'KeyD'];
async function wanderUntil(deadline, seed) {
  const rnd = mulberry32(seed);
  while (Date.now() < deadline) {
    const k = KEYS[Math.floor(rnd() * KEYS.length) % KEYS.length];
    const hold = Math.min(520, deadline - Date.now());
    if (hold <= 0) {
      break;
    }
    await p.keyboard.down(k);
    await sleep(hold);
    await p.keyboard.up(k);
    await sleep(Math.min(80, Math.max(0, deadline - Date.now())));
  }
}

/** loadPerCPU: tải máy trên mỗi nhân, ngay lúc lấy mẫu (luật 16). */
const loadPerCPU = () => loadavg()[0] / Math.max(1, cpus().length);

/** sample đọc các chỉ số của spec cùng một lúc. */
async function sample() {
  await gc();
  const load = loadPerCPU();
  return Object.assign({ load }, await p.evaluate(() => {
    const s = window.__stats();
    const m = performance.memory;
    return {
      tex: Number(s.textures ?? 0),
      texMB: Number(s.texMB ?? 0),
      budgetMB: Number(s.texBudgetMB ?? 0),
      nodes: Number(s.nodes ?? s.sprites ?? 0),
      heapMB: m ? m.usedJSHeapSize / 1048576 : 0,
      fps: Number(s.fps ?? 0),
      zone: Number(s.zone ?? -1),
      // `__stats()` chỉ chở SỐ (nó là bảng số liệu), nên danh sách hệ thống đã tắt phải
      // lấy qua cửa debug `kx` — cùng chỗ với `kx.fault` của §13.8.
      guard: (window.kx?.guard?.()?.systems ?? []).filter((x) => x.dead).map((x) => x.key).join(','),
      // Sổ tài nguyên theo LOẠI (§13.9 "một sổ tài nguyên đếm số đang sống theo loại").
      kinds: Object.fromEntries(Object.entries(s).filter(([k]) =>
        k.startsWith('n_') || ['pool', 'poolText', 'beasts', 'auras', 'flying', 'filters'].includes(k))),
    };
  }));
}

const rows = [];
let zones = 0;
let deaths = 0;
let zonesInWindow = 0;
let deathsInWindow = 0;
/** revivesUnplanned: số lần thấy màn Tổng kết mà kịch bản KHÔNG hẹn trước (bị bot hạ). */
let revivesUnplanned = 0;
/** revivesPlanned: số lượt chết theo kịch bản mà bấm được "Chơi tiếp" thật. */
let revivesPlanned = 0;
console.log(`KỊCH BẢN CỐ ĐỊNH (§13.9 v6.58): ${SCENARIO_VERSION}`);
console.log(`${MINUTES} phút · chu kỳ 60 s · mẫu mỗi ${STEP_SECS} s · khởi động phút 0–${WARMUP_MIN} (không tính) · cửa sổ phút ${WARMUP_MIN}–${MINUTES} · ${BLOCKS} khối`);
console.log(ZONE_MODE === 'none'
  ? 'KHÔNG đổi khu sau lượt neo đầu tiên (bên "0 lượt đổi khu" của phép so B38); cứ 5 chu kỳ một lượt chết\n'
  : 'mỗi chu kỳ đổi sang khu kế tiếp; cứ 5 chu kỳ một lượt chết + vào lại\n');
console.log('phút   khu  texture    texMB   heapMB    nodes  fps');

const t0 = Date.now();
for (let i = 0; i <= totalSamples; i++) {
  // Mốc lấy mẫu tính từ t0 chứ không cộng dồn: một lần chết mất 3 giây không được phép
  // đẩy lệch cả lưới mẫu còn lại, nếu không thì "khối 10 phút" của lượt này không còn
  // cùng nội dung với lượt khác.
  await sleep(t0 + i * STEP_SECS * 1000 - Date.now());
  const s = await sample();
  const min = +((i * STEP_SECS) / 60).toFixed(2);
  const cycle = Math.floor(i / PER_CYCLE);
  const phase = i % PER_CYCLE;
  rows.push({ min, cycle, phase, ...s });
  console.log(`${String(min).padStart(5)} ${String(s.zone).padStart(5)} ${String(s.tex).padStart(8)} ${s.texMB.toFixed(1).padStart(8)} ${s.heapMB.toFixed(1).padStart(8)} ${String(s.nodes).padStart(8)} ${s.fps.toFixed(0).padStart(4)}`);
  if (i === totalSamples) {
    break;
  }

  // Việc của chu kỳ, luôn theo đúng thứ tự này.
  //
  // `revive()` ở đầu bước là lưới an toàn: nếu nhân vật chết ngoài kế hoạch thì không được
  // nằm ở màn Tổng kết suốt phần còn lại. Nhưng mỗi lần nó bắt được một cái chết không hẹn
  // trước là kịch bản đã lệch, nên phải ĐẾM, không phải lặng lẽ chữa (v6.58 mục 4).
  if (await revive()) {
    revivesUnplanned++;
    console.log(`  ⚠ phút ${min}: chết ngoài kế hoạch (đã bấm Chơi tiếp) — kịch bản lệch`);
  }
  // Giữ hai lá chắn SỐNG mỗi chu kỳ, không gắn vào việc đổi khu.
  //
  // - `nogate`: từ v6.56 cổng bắn đi NGẪU NHIÊN khắp mạng, nên đi lại 45 giây mà đạp trúng
  //   một cái là lượt đo có thêm một lần đổi khu không ai đếm, và kết thúc ở khu khác.
  // - `invuln`: kịch bản đặt ra đúng 10 lượt chết; bị bot hạ giữa chừng là thêm một lần
  //   hồi sinh sang khu khác mà lượt đo kia không có.
  //
  // Cả hai hết hạn sau 120 giây. Trước đây chúng nằm trong nhánh đổi khu — đúng ở chế độ
  // `cycle` (đổi khu mỗi chu kỳ) nhưng ở chế độ `none` thì chúng chỉ được đặt một lần rồi
  // hết hạn, và bên "0 lượt đổi khu" của phép so B38 sẽ đầy cái chết ngoài kế hoạch.
  if (phase === 0) {
    await dbg('nogate', 120);
    await dbg('invuln', 120);
  }
  // Chu kỳ 0 luôn nhảy khu một lần để chỗ đứng ban đầu không do xúc xắc quyết (khu vào
  // trận là khu ít người nhất, §8.0). Ở chế độ `none` thì đó là lượt duy nhất.
  if (phase === 0 && cycle < totalCycles && (ZONE_MODE === 'cycle' || cycle === 0)) {
    await dbg('zone', ZONE_MODE === 'none' ? 1 : zoneAtCycle(cycle));
    zones++;
    if (min >= WARMUP_MIN) {
      zonesInWindow++;
    }
    await sleep(1200);
  }
  if (phase === 2 && cycle % DEATH_EVERY === DEATH_EVERY - 1) {
    await dbg('die');
    deaths++;
    if (min >= WARMUP_MIN) {
      deathsInWindow++;
    }
    await sleep(1800);
    // Lượt chết theo kịch bản PHẢI bấm được "Chơi tiếp". Không bấm được nghĩa là lệnh
    // `die` rơi vào hư vô — đúng loại "thành công giả" mà v6.58 mục 4 bắt phải phát hiện.
    if (await revive()) {
      revivesPlanned++;
    } else {
      console.log(`  ⚠ phút ${min}: gửi lệnh chết mà KHÔNG thấy màn Tổng kết`);
    }
    // Hồi sinh thả người vào khu ÍT NGƯỜI NHẤT (§8.0), không phải khu đang đứng — nên
    // kéo về đúng khu của chu kỳ này rồi mới đi tiếp. Lệnh này là no-op khi đã đúng khu,
    // nên nó không được tính vào số lượt đổi khu của kịch bản.
    await dbg('zone', ZONE_MODE === 'none' ? 1 : zoneAtCycle(cycle));
    await dbg('nogate', 120);
    // Và không để bot hạ giữa chừng: kịch bản đặt ra đúng 10 lượt chết, mỗi lượt chết
    // ngoài kế hoạch là một lần hồi sinh sang khu khác mà không lượt đo nào khác có.
    await dbg('invuln', 120);
    await sleep(1200);
  }
  await wanderUntil(t0 + (i + 1) * STEP_SECS * 1000 - 300, 0x9e3779b9 ^ i);
}

const last = rows[rows.length - 1];
const win = rows.filter((r) => r.min >= WARMUP_MIN);
const blockSpan = (MINUTES - WARMUP_MIN) / BLOCKS;
console.log(`\nkịch bản: ${zones} lượt đổi khu (${zonesInWindow} trong cửa sổ) · ${deaths} lượt chết (${deathsInWindow} trong cửa sổ) · khu kết thúc ${last.zone}`);

// Luật 16: máy bận thì KHÔNG kết luận — và "bận" phải đo trong SUỐT cửa sổ, không phải
// lúc bấm nút chạy. Một bài 60 phút bắt đầu lúc máy rảnh vẫn có thể chạy xuyên qua một
// đợt build của việc khác.
const busyRows = win.filter((r) => r.load > MAX_LOAD);
const busyFrac = win.length ? busyRows.length / win.length : 0;
const loadMax = win.length ? Math.max(...win.map((r) => r.load)) : 0;
const fpsMin = win.length ? Math.min(...win.map((r) => r.fps)) : 0;
const busyMachine = !IGNORE_LOAD && busyFrac > BUSY_TOL;
console.log(`tải máy trong cửa sổ: ${(busyFrac * 100).toFixed(0)}% số mẫu vượt ${MAX_LOAD}/nhân · đỉnh ${loadMax.toFixed(2)}/nhân · FPS thấp nhất ${fpsMin.toFixed(0)}`);

// ── Hành động đã XẢY RA chưa? (v6.58 mục 4) ─────────────────────────────────────────────
//
// Đếm số lệnh đã GỬI là không đủ — đó đúng là cách sáu lượt đo cũ tự nhận "50 lượt đổi khu"
// trong khi lệnh thiếu tham số nên gần như không đổi khu lần nào, và không ai biết. Nay đếm
// số lần khu THẬT SỰ đổi trong bảng mẫu, rồi so với số mong đợi; lệch là KHÔNG KẾT LUẬN.
let zoneSeen = 0;
let zoneSeenWindow = 0;
for (let i = 1; i < rows.length; i++) {
  if (rows[i].zone !== rows[i - 1].zone) {
    zoneSeen++;
    if (rows[i].min >= WARMUP_MIN) {
      zoneSeenWindow++;
    }
  }
}
const scenarioProblems = [];
if (zoneSeenWindow !== zonesInWindow) {
  scenarioProblems.push(`đổi khu trong cửa sổ: gửi ${zonesInWindow} lệnh nhưng bảng mẫu thấy ${zoneSeenWindow} lần đổi thật`);
}
if (revivesPlanned !== deaths) {
  scenarioProblems.push(`chết theo kịch bản: gửi ${deaths} lệnh nhưng chỉ ${revivesPlanned} lần thấy màn Tổng kết`);
}
if (revivesUnplanned > 0) {
  scenarioProblems.push(`${revivesUnplanned} lần chết NGOÀI kế hoạch (bot hạ được dù đã bật miễn đòn)`);
}
console.log(`kiểm hành động: ${zoneSeenWindow}/${zonesInWindow} lượt đổi khu thật trong cửa sổ (cả bài ${zoneSeen}/${zones})`
  + ` · ${revivesPlanned}/${deaths} lượt chết theo kịch bản · ${revivesUnplanned} lượt chết ngoài kế hoạch`);
if (scenarioProblems.length) {
  // Không chặn deploy: đây là bài đo tự khai mình chạy sai, không phải build sai.
  console.log('\n== KỊCH BẢN LỆCH — KHÔNG KẾT LUẬN ==');
  for (const s of scenarioProblems) {
    console.log(`  · ${s}`);
  }
}
/**
 * invalid: lượt đo này KHÔNG được phép kết luận gì — máy bận, hoặc kịch bản lệch.
 *
 * Hai lý do khác nhau nhưng hệ quả giống nhau, và hệ quả ấy là điều quan trọng: không chặn
 * deploy (build không có lỗi gì ở đây), và **không ghi vào mốc kỷ lục lẫn sổ đêm** (v6.58).
 */
const invalid = busyMachine || scenarioProblems.length > 0;
if (pageErrors.length) {
  console.log(`\n${pageErrors.length} lỗi trang (xem lớp chống sập §13.8):`);
  for (const e of [...new Set(pageErrors)].slice(0, 5)) {
    console.log(`  ${e}`);
  }
}
if (last.guard) {
  console.log(`\n⚠ client đã phải TẮT hệ thống: ${last.guard} (§13.8)`);
}

const METRICS = [
  ['heapMB', (r) => r.heapMB, 'MB'],
  ['texMB', (r) => r.texMB, 'MB'],
  ['texture', (r) => r.tex, ''],
  ['display object', (r) => r.nodes, ''],
];

// Bài quá ngắn thì KHÔNG kết luận gì, và nói thẳng ra là không kết luận. Một lượt 8 phút
// không chứng minh được "không rò rỉ": cửa sổ của nó nằm trọn trong giai đoạn khởi động,
// nên nó sẽ đỏ với MỌI build, kể cả build sạch — và một cổng đỏ với mọi build sẽ bị tắt.
// Ít nhất 2 phút mỗi khối thì dốc khối mới có nghĩa. `MIN_WINDOW` hạ được xuống để thử
// đường đi của bài đo trong vài phút — nhưng mặc định phải là con số gác thật, vì một
// ngưỡng dễ dãi đặt sẵn trong tệp là ngưỡng sẽ được dùng.
const MIN_WINDOW = Number(process.env.MIN_WINDOW ?? BLOCKS * 2);
if (MINUTES - WARMUP_MIN < MIN_WINDOW || win.length < BLOCKS * 4) {
  console.log(`\n== KHÔNG KẾT LUẬN — cửa sổ chỉ ${(MINUTES - WARMUP_MIN).toFixed(1)} phút / ${win.length} mẫu ==`);
  console.log(`Cần ít nhất ${WARMUP_MIN + MIN_WINDOW} phút (spec chốt ${WARMUP_MIN + 50}) thì mới có ${BLOCKS} khối để fit dốc.`);
  console.log('Bài ngắn chỉ dùng để xem bài đo còn chạy được, không dùng để gác.');
  if (JSON_OUT) {
    writeJSONAtomic(JSON_OUT, {
      at: Math.floor(Date.now() / 1000), build: BUILD, verdict: 'chua_ket_luan',
      minutes: MINUTES, warmup_min: WARMUP_MIN, step_secs: STEP_SECS,
      note: `cửa sổ chỉ ${(MINUTES - WARMUP_MIN).toFixed(1)} phút`,
    });
  }
  await b.close();
  process.exit(3);
}

// ── 1. Dốc theo khối + KTC 95% ──────────────────────────────────────────────────────────
console.log(`\n== dốc theo ${BLOCKS} khối × ${blockSpan} phút · KTC 95% (t, df = ${BLOCKS - 1}) ==`);
const metrics = {};
const reasons = [];
let red = false;
for (const [name, get, unit] of METRICS) {
  const t = blockTrend(rows, get, WARMUP_MIN, MINUTES);
  const base = baselineAt(rows, get, WARMUP_MIN, blockSpan);
  const rise = t.mean * 60;
  const risePct = base > 0 ? (rise / base) * 100 : 0;
  // ĐỎ khi cận dưới KTC > 0 (chắc chắn đang dốc lên) VÀ mức tăng một giờ vượt 10% mức nền
  // (dốc lên đủ nhiều để đáng gọi là rò rỉ). Thiếu một trong hai thì chưa đủ để chặn.
  const isRed = t.lo > 0 && risePct > RISE_PCT;
  if (isRed) {
    red = true;
    reasons.push(`${name}: dốc ${t.mean.toFixed(3)}/phút, KTC [${t.lo.toFixed(3)}; ${t.hi.toFixed(3)}] > 0, một giờ +${risePct.toFixed(1)}% mức nền`);
  }
  metrics[name] = {
    blocks: t.slopes.map((v) => +v.toFixed(4)),
    mean: +t.mean.toFixed(4), ci_lo: +t.lo.toFixed(4), ci_hi: +t.hi.toFixed(4),
    baseline: +base.toFixed(2), rise60: +rise.toFixed(2), rise_pct: +risePct.toFixed(2),
    red: isRed,
  };
  console.log(`${name.padEnd(15)} nền ${base.toFixed(1).padStart(8)} ${unit.padEnd(2)} · dốc ${t.slopes.map((v) => v.toFixed(2).padStart(6)).join(' ')}`);
  console.log(`${''.padEnd(15)} TB ${(t.mean >= 0 ? '+' : '') + t.mean.toFixed(3)}/phút · KTC 95% [${t.lo.toFixed(3)}; ${t.hi.toFixed(3)}] · 60 phút ${(rise >= 0 ? '+' : '') + rise.toFixed(1)} ${unit} (${risePct >= 0 ? '+' : ''}${risePct.toFixed(1)}% nền)  ${isRed ? 'ĐỎ' : t.lo > 0 ? 'dốc lên nhưng < ngưỡng' : 'không kết luận được là dốc lên'}`);
}

// ── 2. Mốc kỷ lục cho mức nền + chuỗi dấu nhiều đêm ─────────────────────────────────────
//
// Hai luật này cần TRÍ NHỚ giữa các đêm, nên chỉ chạy khi có STATE_DIR — và KHÔNG chạy
// khi máy bận: một đêm nhiễu mà hạ được mốc kỷ lục thì mọi đêm sau đều đỏ oan, còn một
// dấu dốc nhiễu thì làm hỏng chuỗi 3 đêm. Trí nhớ dài hạn chỉ được nhận số sạch.
const ratchetOut = {};
const signsOut = {};
let yellow = false;
if (STATE_DIR && !invalid) {
  const ratchetPath = join(STATE_DIR, 'leak-ratchet.json');
  const nightsPath = join(STATE_DIR, 'leak-nights.json');
  const ratchetFile = readJSON(ratchetPath, {});
  const nightsFile = readJSON(nightsPath, {});
  // Đổi kịch bản → MỐC CŨ HẾT GIÁ TRỊ, chuỗi dấu bắt đầu lại (v6.58).
  //
  // Số của hai kịch bản khác nhau không so với nhau được. Lỗi `dbg('zone')` là ví dụ đắt
  // nhất: sáu lượt cũ chạy trên một kịch bản gần như không đổi khu nên mức nền thấp hơn
  // hẳn — giữ mốc đó lại thì mọi lượt của kịch bản thật đỏ ngay, mà đỏ vì lý do sai.
  const sameScenario = ratchetFile.scenario_version === SCENARIO_VERSION;
  const ratchet = sameScenario ? ratchetFile : {};
  const nights = nightsFile.scenario_version === SCENARIO_VERSION ? (nightsFile.nights ?? []) : [];
  if (!sameScenario && ratchetFile.scenario_version) {
    console.log(`\nkịch bản đổi: "${ratchetFile.scenario_version}" → "${SCENARIO_VERSION}"`);
    console.log('→ mốc kỷ lục cũ hết giá trị, chuỗi dấu bắt đầu lại từ lượt này (v6.58).');
  }

  // Mốc kỷ lục: gác MỨC NỀN của heap.
  //
  // Vì sao chỉ heap: §13.10 đã có một cổng tuyệt đối cho GPU MB (ngân sách thiết bị), còn
  // mức nền đang là vấn đề mở thì là JS heap — 137–152 MB trên mục tiêu 80 (B39). Mốc kỷ
  // lục của texMB vẫn được ghi lại để có số liệu, nhưng không chặn: thêm một cổng chặn mà
  // spec không nêu là tự sửa luật (CLAUDE.md luật 4).
  for (const key of ['heapMB', 'texMB']) {
    const value = metrics[key].baseline;
    const step = ratchetStep(ratchet[key] ?? {}, value);
    const gated = key === 'heapMB';
    const over = step.exceeded > RATCHET_PCT && !step.first;
    if (gated && over) {
      red = true;
      reasons.push(`${key}: mức nền ${value.toFixed(1)} vượt mốc kỷ lục ${(ratchet[key]?.record ?? 0).toFixed(1)} tới ${step.exceeded.toFixed(1)}% (> ${RATCHET_PCT}%)`);
    }
    ratchetOut[key] = {
      record: +step.record.toFixed(2), value, exceeded_pct: +step.exceeded.toFixed(2),
      pending: step.pending, lowered: !!step.lowered, first: !!step.first, gate: gated,
    };
    ratchet[key] = { record: step.record, pending: step.pending, at: Math.floor(Date.now() / 1000), build: BUILD };
    console.log(`\nmốc kỷ lục ${key}: nền lượt này ${value.toFixed(1)} / mốc ${step.record.toFixed(1)} `
      + `(${step.exceeded >= 0 ? '+' : ''}${step.exceeded.toFixed(1)}%)`
      + `${step.first ? ' — lượt đầu, đặt mốc' : ''}${step.lowered ? ' — HẠ MỐC (đã đủ 2 đêm xác nhận)' : ''}`
      + `${step.pending ? ` — chờ xác nhận ${step.pending.nights}/2 đêm ở ${step.pending.value.toFixed(1)}` : ''}`
      + `${gated ? (over ? '  ĐỎ' : '  ĐẠT') : '  (ghi sổ, không chặn)'}`);
  }

  // Chuỗi dấu nhiều đêm: lưới bắt rò rỉ CHẬM mà một đêm không phân giải nổi.
  //
  // Một dốc +0,2 MB/phút có KTC ôm qua 0 ở mọi đêm riêng lẻ, nhưng nếu nó dương ở cả ba
  // đêm liên tiếp thì xác suất ngẫu nhiên chỉ còn 1/8 — đủ để mở việc điều tra, chưa đủ
  // để chặn deploy. Đêm thứ tư vẫn dương: 1/16, lúc đó chặn.
  const night = {
    at: Math.floor(Date.now() / 1000), build: BUILD,
    mean: Object.fromEntries(Object.entries(metrics).map(([k, v]) => [k, v.mean])),
    baseline: Object.fromEntries(Object.entries(metrics).map(([k, v]) => [k, v.baseline])),
  };
  nights.push(night);
  const kept = nights.slice(-30);
  for (const [name] of METRICS) {
    const signs = kept.map((n) => ((n.mean?.[name] ?? 0) > 0 ? '+' : '−'));
    let streak = 0;
    for (let i = signs.length - 1; i >= 0 && signs[i] === '+'; i--) {
      streak++;
    }
    signsOut[name] = { signs: signs.slice(-7), streak };
    if (streak >= 4) {
      red = true;
      reasons.push(`${name}: dốc dương ${streak} đêm liên tiếp (đêm thứ 4 trở đi → ĐỎ)`);
    } else if (streak === 3) {
      yellow = true;
    }
  }
  console.log('\nchuỗi dấu dốc 7 đêm gần nhất (dấu của đêm cũ → đêm nay):');
  for (const [name] of METRICS) {
    const s = signsOut[name];
    console.log(`  ${name.padEnd(15)} ${s.signs.join(' ')}   chuỗi dương hiện tại: ${s.streak}${s.streak >= 4 ? ' → ĐỎ' : s.streak === 3 ? ' → VÀNG' : ''}`);
  }

  ratchet.scenario_version = SCENARIO_VERSION;
  writeJSONAtomic(ratchetPath, ratchet);
  writeJSONAtomic(nightsPath, { scenario_version: SCENARIO_VERSION, nights: kept });
}

// ── 3. Sổ tài nguyên: dốc tại CÙNG MỘT ĐIỂM trong chu kỳ ────────────────────────────────
//
// §13.9 v6.57: "lấy mẫu tại cùng một điểm trong chu kỳ kịch bản (vd luôn ngay sau khi vào
// khu thứ k). So hai mẫu tức thời chỉ đo lượt này kết thúc ở khu đông hay khu vắng."
//
// Điểm neo ở đây: mẫu thứ hai của chu kỳ (15 giây sau khi vào khu) trong những chu kỳ vừa
// nhảy sang khu 1. Mỗi 5 phút một điểm → 10 điểm trong cửa sổ. Ít điểm, nên KTC tính từ
// sai số dư của chính đường fit (t, df = n−2) chứ không chia khối.
const ledger = [];
const anchors = rows.filter((r) => r.min >= WARMUP_MIN && r.phase === 1 && r.cycle % DEATH_EVERY === 0 && r.kinds);
if (anchors.length >= 3) {
  const keys = [...new Set(anchors.flatMap((r) => Object.keys(r.kinds)))].sort();
  for (const k of keys) {
    const pts = anchors.map((r) => [r.min, r.kinds[k] ?? 0]);
    const t = singleTrend(pts);
    const mean = pts.reduce((a, [, v]) => a + v, 0) / pts.length;
    ledger.push({ key: k, slope: +t.slope.toFixed(3), ci_lo: +t.lo.toFixed(3), ci_hi: +t.hi.toFixed(3), mean: Math.round(mean), n: t.n, rising: t.lo > 0 });
  }
  ledger.sort((a, c) => c.slope - a.slope);
  console.log(`\n== sổ tài nguyên: dốc tại cùng điểm chu kỳ (${anchors.length} điểm, ngay sau khi vào khu 1) ==`);
  console.log('loại              TB    dốc/phút   KTC 95%              kết luận');
  for (const d of ledger.slice(0, 14)) {
    console.log(`  ${d.key.padEnd(14)} ${String(d.mean).padStart(6)} ${(d.slope >= 0 ? '+' : '') + d.slope.toFixed(2).padStart(8)}   [${d.ci_lo.toFixed(2)}; ${d.ci_hi.toFixed(2)}]`.padEnd(66)
      + (d.rising ? 'đang tăng' : 'chưa kết luận'));
  }
} else {
  console.log('\n== sổ tài nguyên: chưa đủ điểm neo để fit dốc ==');
}

// Nhóm texture nào phình, gom theo nơi cấp phát — chỉ để chỉ đường khi đỏ, không phải phép
// xét (nó là hai mẫu tức thời, đúng thứ v6.57 cấm dùng để kết luận).
const sites = await p.evaluate(() => window.__texsites?.(12) ?? []);
if (sites.length) {
  console.log('\n== texture đang giữ, gom theo nơi cấp phát (chỉ đường, KHÔNG phải phép xét) ==');
  for (const x of sites.slice(0, 8)) {
    console.log(`${String(x.count).padStart(6)}  ${String(x.mb).padStart(8)} MB  ${x.site.slice(0, 140)}`);
  }
}

// ── 4. Ngân sách bộ nhớ đồ hoạ của THIẾT BỊ ─────────────────────────────────────────────
//
// Phép xét theo dốc mù với một thứ: một build có thể đứng yên tuyệt đối mà vẫn đứng yên ở
// mức 300 MB trên iPhone — tức là tab sẽ bị Safari giết, chỉ là không giết thêm. Ngưỡng
// tuyệt đối này bắt đúng trường hợp đó. Con số lấy từ chính trang đang chạy, không chép tay.
const peakMB = Math.max(...rows.map((r) => r.texMB));
const budget = last.budgetMB || 0;
if (budget > 0 && peakMB >= budget) {
  red = true;
  reasons.push(`GPU MB đỉnh ${peakMB.toFixed(1)} vượt ngân sách thiết bị ${budget}`);
}
if (budget > 0) {
  console.log(`\nngân sách GPU thiết bị: đỉnh ${peakMB.toFixed(1)} / ${budget} MB  ${peakMB < budget ? 'ĐẠT' : 'TRƯỢT'}`);
}

// JS heap so với MỤC TIÊU §13.10 — in ra, KHÔNG chặn (v6.57 nói rõ 80 MB là đích).
const HEAP_TARGET = Number(process.env.HEAP_TARGET ?? 80);
console.log(`\nJS heap mức nền ${metrics.heapMB.baseline.toFixed(1)} / mục tiêu ${HEAP_TARGET} MB (§13.10) `
  + `${metrics.heapMB.baseline <= HEAP_TARGET ? 'ĐẠT' : '— trên đích, nhưng cổng chặn là MỐC KỶ LỤC, không phải 80 MB (v6.57)'}`);

const verdict = invalid ? 'chua_ket_luan' : red ? 'do' : yellow ? 'vang' : 'xanh';
if (JSON_OUT) {
  writeJSONAtomic(JSON_OUT, {
    at: Math.floor(Date.now() / 1000), build: BUILD, verdict, reasons,
    load: { busy_frac: +busyFrac.toFixed(3), max: +loadMax.toFixed(2), fps_min: +fpsMin.toFixed(0), limit: MAX_LOAD },
    scenario_version: SCENARIO_VERSION,
    note: invalid ? [busyMachine ? `máy bận ${(busyFrac * 100).toFixed(0)}% số mẫu trong cửa sổ` : '', ...scenarioProblems].filter(Boolean).join(' · ') : undefined,
    minutes: MINUTES, warmup_min: WARMUP_MIN, step_secs: STEP_SECS, blocks: BLOCKS,
    scenario: {
      zone_switches: zones, zone_switches_window: zonesInWindow,
      deaths, deaths_window: deathsInWindow, end_zone: last.zone, cycles: totalCycles,
    },
    metrics, ratchet: ratchetOut, signs: signsOut, ledger: ledger.slice(0, 20),
    gpu: { peak_mb: +peakMB.toFixed(1), budget_mb: budget },
    heap_target: HEAP_TARGET,
    page_errors: [...new Set(pageErrors)].slice(0, 5),
  });
}

await b.close();
if (invalid) {
  // Đặt TRƯỚC phán quyết đỏ: một lượt đo trên máy bận — hoặc chạy sai kịch bản — không được
  // phép chặn deploy, và cũng không được phép "chứng minh" là build sạch. Nó chỉ được phép
  // nói "đo lại đi".
  console.log('\nKHÔNG KẾT LUẬN — số ở trên chỉ để xem, không dùng để gác:');
  if (busyMachine) {
    console.log(`  · máy bận ${(busyFrac * 100).toFixed(0)}% số mẫu trong cửa sổ (luật 16, ngưỡng ${MAX_LOAD}/nhân); FPS tụt xuống ${fpsMin.toFixed(0)} thì client tự hạ chất lượng, mà hạ chất lượng đổi chính những thứ bài này đếm`);
  }
  for (const s of scenarioProblems) {
    console.log(`  · ${s}`);
  }
  console.log('Chạy lại lúc máy rảnh (tools/leakrepeat.sh tự chờ), hoặc IGNORE_LOAD=1 nếu chỉ muốn xem bài đo còn chạy.');
  process.exit(3);
}
if (red) {
  console.log('\nĐỎ (§13.9 v6.57) — chặn deploy kế tiếp:');
  for (const r of reasons) {
    console.log(`  · ${r}`);
  }
  process.exit(1);
}
if (yellow) {
  console.log('\nVÀNG (§13.9 v6.57): dốc dương 3 đêm liên tiếp nhưng KTC vẫn ôm qua 0 — mở việc điều tra, KHÔNG chặn deploy.');
  console.log('Đêm thứ 4 vẫn dương thì chuyển ĐỎ.');
  process.exit(0);
}
console.log('\nĐẠT (§13.9 v6.57): không chỉ số nào có dốc chắc chắn dương vượt ngưỡng, mức nền trong mốc kỷ lục, GPU MB trong ngân sách.');
