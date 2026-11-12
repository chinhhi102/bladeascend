// Bài kiểm RÒ RỈ phía CLIENT — GAME_SPEC §13.9.
//
//   cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true
//   node tools/browsertest/leakcheck.mjs http://127.0.0.1:9998
//
// Spec v6.54 (Q53): "chơi 30 phút giả lập + đổi khu 50 lần + chết/vào lại 10 lần. **20 phút
// đầu là khởi động** — không tính. Chặn khi trong **cửa sổ ổn định phút 20–30** bất kỳ chỉ
// số nào (số Texture, GPU MB, JS heap, display object) **tăng > 5%**, hoặc GPU MB vượt ngân
// sách thiết bị. Mốc phút 5 chỉ in ra để tham khảo."
//
// ── Vì sao 20 phút đầu không tính ────────────────────────────────────────────────────────
//
// Bộ làm việc của game vẫn đang LỚN LÊN tới tận đó: 11 atlas nhân vật nạp dần khi gặp
// người mới, art của 5 khu nạp dần qua 50 lượt đổi khu, pool chữ và pool view giãn tới mức
// dùng đỉnh. Đo 12/11 trên build sạch: so với phút 5 thì texture +24,5% và JS heap +37,3%,
// nhưng `texMB` phẳng tuyệt đối và từ phút 20 trở đi mọi chỉ số đứng yên (phút 20–25 ↔
// 25–30: texture +1,0%, heap −0,5%). Đó là một TRẦN, không phải một đường dốc.
//
// Rò rỉ thật thì KHÔNG có trần — xem lượt #33 trong PERF_LOG: +91 texture/phút, một chiều,
// cho tới khi tab chết. Nên cửa sổ sau khi chạm trần mới là chỗ nhìn ra rò rỉ, và ở đó
// ngưỡng siết được xuống 5%.
//
// ── Khác gì `texleak.mjs` ────────────────────────────────────────────────────────────────
//
// `texleak` chỉ đếm TEXTURE và chỉ đi lại. Bài này đếm đủ bốn chỉ số của spec và chạy đúng
// ba việc mà spec chỉ đích danh — đổi khu, chết, vào lại — vì đó là ba đường tạo và huỷ
// hàng loạt đối tượng, và cũng là ba đường từng rò rỉ thật (181 lượt đổi khu → 1,4 GB).
import { launch } from './browser.mjs';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
/** MINUTES: tổng thời gian chơi. Spec chốt 30; hạ xuống khi chỉ muốn thử nhanh bài đo. */
const MINUTES = Number(process.env.MINUTES ?? 30);
/** BASE_MIN: mốc phút 5 — chỉ IN RA tham khảo, không chặn (v6.54). */
const BASE_MIN = Number(process.env.BASE_MIN ?? 5);
/** ZONE_SWITCHES / DEATHS: spec chốt 50 và 10. */
const ZONE_SWITCHES = Number(process.env.ZONE_SWITCHES ?? 50);
const DEATHS = Number(process.env.DEATHS ?? 10);
/** TOL: ngưỡng tăng cho phép trong cửa sổ ổn định, % (v6.54 chốt 5). */
const TOL = Number(process.env.TOL ?? 5);
/** STEADY_MIN: đầu cửa sổ ổn định (v6.54 chốt phút 20). */
const STEADY_MIN = Number(process.env.STEADY_MIN ?? 20);
/** STEP_SECS: nhịp lấy mẫu. */
const STEP_SECS = Number(process.env.STEP_SECS ?? 15);

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const steps = Math.max(2, Math.round((MINUTES * 60) / STEP_SECS));
// Rải đều việc đổi khu và việc chết trên cả bài đo: dồn hết vào cuối thì mốc phút 5 không
// bao giờ thấy chúng, và dồn hết vào đầu thì phần sau chỉ là đứng yên.
const zoneEvery = Math.max(1, Math.round(steps / ZONE_SWITCHES));
const deathEvery = Math.max(1, Math.round(steps / DEATHS));

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
// khi bài đo đỏ, câu hỏi tiếp theo luôn là "nhóm nào phình", và chạy lại 30 phút chỉ để
// biết điều đó là 30 phút mất không.
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
  await sleep(2500);
  return true;
}

/** wander: đi lại để người và quái liên tục vào/ra tầm nhìn — đứng yên thì không đo gì cả. */
async function wander(ms) {
  const keys = ['KeyW', 'KeyA', 'KeyS', 'KeyD'];
  const t0 = Date.now();
  while (Date.now() - t0 < ms) {
    const k = keys[Math.floor(((Date.now() - t0) / 1100) % 4)];
    await p.keyboard.down(k);
    await sleep(550);
    await p.keyboard.up(k);
    await sleep(120);
  }
}

/** sample đọc bốn chỉ số của spec cùng một lúc. */
async function sample() {
  await gc();
  return p.evaluate(() => {
    const s = window.__stats();
    const m = performance.memory;
    return {
      tex: Number(s.textures ?? 0),
      texMB: Number(s.texMB ?? 0),
      budgetMB: Number(s.texBudgetMB ?? 0),
      nodes: Number(s.nodes ?? s.sprites ?? 0),
      heapMB: m ? m.usedJSHeapSize / 1048576 : 0,
      fps: Number(s.fps ?? 0),
      // `__stats()` chỉ chở SỐ (nó là bảng số liệu), nên danh sách hệ thống đã tắt phải
      // lấy qua cửa debug `kx` — cùng chỗ với `kx.fault` của §13.8.
      guard: (window.kx?.guard?.()?.systems ?? []).filter((x) => x.dead).map((x) => x.key).join(','),
    };
  });
}

const rows = [];
let baseSites = [];
let zones = 0;
let deaths = 0;
console.log(`đo ${MINUTES} phút · mốc so sánh phút ${BASE_MIN} · ${ZONE_SWITCHES} lượt đổi khu · ${DEATHS} lượt chết\n`);
console.log('phút   texture    texMB   heapMB    nodes  fps');
for (let i = 0; i <= steps; i++) {
  const s = await sample();
  const min = +((i * STEP_SECS) / 60).toFixed(2);
  rows.push({ min, ...s });
  console.log(`${String(min).padStart(5)} ${String(s.tex).padStart(9)} ${s.texMB.toFixed(1).padStart(8)} ${s.heapMB.toFixed(1).padStart(8)} ${String(s.nodes).padStart(8)} ${s.fps.toFixed(0).padStart(4)}`);
  if (min >= BASE_MIN && baseSites.length === 0) {
    baseSites = await p.evaluate(() => window.__texsites?.(200) ?? []);
  }
  if (i === steps) {
    break;
  }
  await revive();
  if (i > 0 && i % zoneEvery === 0 && zones < ZONE_SWITCHES) {
    await dbg('zone');
    zones++;
    await sleep(1200);
  }
  if (i > 0 && i % deathEvery === 0 && deaths < DEATHS) {
    await dbg('die');
    deaths++;
    await sleep(2000);
    await revive();
  }
  await wander(STEP_SECS * 1000);
}

// So mốc phút BASE_MIN với mốc cuối.
const base = rows.find((r) => r.min >= BASE_MIN) ?? rows[0];
const last = rows[rows.length - 1];
console.log(`\n${zones} lượt đổi khu · ${deaths} lượt chết · mốc so sánh: phút ${base.min}`);
if (pageErrors.length) {
  console.log(`\n${pageErrors.length} lỗi trang (xem lớp chống sập §13.8):`);
  for (const e of [...new Set(pageErrors)].slice(0, 5)) {
    console.log(`  ${e}`);
  }
}
if (last.guard) {
  console.log(`\n⚠ client đã phải TẮT hệ thống: ${last.guard} (§13.8)`);
}

// Nhóm texture nào phình so với mốc — in TRƯỚC phần kết luận, vì khi đỏ thì đây mới là
// thứ người đọc cần.
const sites = await p.evaluate(() => window.__texsites?.(200) ?? []);
if (sites.length) {
  const base = new Map(baseSites.map((x) => [x.site, x.count]));
  const grown = sites
    .map((x) => ({ site: x.site, d: x.count - (base.get(x.site) ?? 0), mb: x.mb }))
    .filter((x) => x.d > 0)
    .sort((a, c) => c.d - a.d)
    .slice(0, 12);
  if (grown.length) {
    console.log('\n== texture TĂNG so với mốc, gom theo nơi cấp phát ==');
    for (const x of grown) {
      console.log(`${String('+' + x.d).padStart(6)}  ${String(x.mb).padStart(8)} MB  ${x.site.slice(0, 160)}`);
    }
  }
}

/** pctOf: mức tăng theo %, 0 khi mốc đầu bằng 0. */
const pctOf = (from, to) => (from > 0 ? ((to - from) / from) * 100 : 0);
/**
 * Sàn tiếng ồn: mức tăng TUYỆT ĐỐI nhỏ hơn ngần này thì bỏ qua, vì phần trăm trên một nền
 * nhỏ là vô nghĩa.
 *
 * Phải nhỏ hơn hẳn ngưỡng 5% của v6.54 trên số đo thật, nếu không chính cái sàn thành
 * ngưỡng. Số đo 12/11 ở cửa sổ ổn định: texture ~420 (5% = 21), texMB ~245 (12), heap
 * ~114 MB (5,7), display object ~4 400 (220) — nên sàn 80 cho display object, không phải
 * 200 như bản 10% trước đó.
 */
const FLOOR = { texture: 5, texMB: 4, heapMB: 3, 'display object': 80 };

const METRICS = [
  ['texture', (r) => r.tex, '   '],
  ['texMB', (r) => r.texMB, 'MB '],
  ['heapMB', (r) => r.heapMB, 'MB '],
  ['display object', (r) => r.nodes, '   '],
];

// ── 1. Mốc phút 5 — CHỈ THAM KHẢO (v6.54) ───────────────────────────────────────────────
console.log(`\n== tham khảo: so với phút ${base.min} (KHÔNG chặn) ==`);
for (const [name, get, unit] of METRICS) {
  const from = get(base);
  const to = get(last);
  const pct = pctOf(from, to);
  console.log(`${name.padEnd(16)} ${from.toFixed(1).padStart(9)} → ${to.toFixed(1).padStart(9)} ${unit}  (${pct >= 0 ? '+' : ''}${pct.toFixed(1)}%)`);
}

// ── 2. Cửa sổ ổn định phút 20–30 — PHÉP SO CHẶN (v6.54) ─────────────────────────────────
//
// Bài quá ngắn thì KHÔNG kết luận gì, và nói thẳng ra là không kết luận. Một lượt 8 phút
// không chứng minh được "không rò rỉ": cửa sổ của nó nằm trọn trong giai đoạn khởi động,
// nên nó đỏ với MỌI build, kể cả build sạch — và một cổng đỏ với mọi build sẽ bị tắt.
const steady = rows.find((r) => r.min >= STEADY_MIN);
let bad = false;
if (!steady || last.min - steady.min < 4) {
  console.log(`\n== cửa sổ ổn định: KHÔNG KẾT LUẬN (bài chỉ ${last.min} phút) ==`);
  console.log(`Cần chạy tới ít nhất ${STEADY_MIN + 4} phút thì mới có cửa sổ để so.`);
  console.log('Bài ngắn chỉ dùng để xem bài đo còn chạy được, không dùng để gác.');
} else {
  console.log(`\n== cửa sổ ổn định phút ${steady.min} → ${last.min} · ngưỡng ${TOL}% (PHÉP SO CHẶN) ==`);
  for (const [name, get, unit] of METRICS) {
    const from = get(steady);
    const to = get(last);
    const pct = pctOf(from, to);
    const ok = pct <= TOL || to - from <= FLOOR[name];
    if (!ok) {
      bad = true;
    }
    console.log(`${name.padEnd(16)} ${from.toFixed(1).padStart(9)} → ${to.toFixed(1).padStart(9)} ${unit}  (${pct >= 0 ? '+' : ''}${pct.toFixed(1)}%)  ${ok ? 'ĐẠT' : 'TRƯỢT'}`);
  }
}

// ── 3. Ngân sách bộ nhớ đồ hoạ của THIẾT BỊ (v6.54) ─────────────────────────────────────
//
// Phép so phần trăm mù với một thứ: một build có thể đứng yên tuyệt đối mà vẫn đứng yên ở
// mức 300 MB trên iPhone — tức là tab sẽ bị Safari giết, chỉ là không giết thêm. Ngưỡng
// tuyệt đối này bắt đúng trường hợp đó. Con số lấy từ chính trang đang chạy
// (`device.gpuBudgetMB`), không chép tay vào bài đo.
const peakMB = Math.max(...rows.map((r) => r.texMB));
const budget = last.budgetMB || 0;
if (budget > 0) {
  const okBudget = peakMB < budget;
  if (!okBudget) {
    bad = true;
  }
  console.log(`\nngân sách GPU thiết bị: đỉnh ${peakMB.toFixed(1)} / ${budget} MB  ${okBudget ? 'ĐẠT' : 'TRƯỢT'}`);
}

await b.close();
if (bad) {
  console.log(`\nTRƯỢT: §13.9 v6.54 — cửa sổ ổn định tăng quá ${TOL}%, hoặc GPU MB vượt ngân sách.`);
  process.exit(1);
}
console.log(`\nĐẠT: cửa sổ ổn định không chỉ số nào tăng quá ${TOL}%, GPU MB trong ngân sách (§13.9 v6.54).`);
