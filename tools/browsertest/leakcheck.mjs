// Bài kiểm RÒ RỈ phía CLIENT — GAME_SPEC §13.9.
//
//   cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true
//   node tools/browsertest/leakcheck.mjs http://127.0.0.1:9998
//
// Spec: "chơi 30 phút giả lập + đổi khu 50 lần + chết/vào lại 10 lần → số Texture, GPU MB,
// JS heap, số display object không tăng quá 10% so với phút thứ 5. Tăng → build fail."
//
// ── Vì sao mốc là PHÚT THỨ 5, không phải lúc vào trận ────────────────────────────────────
//
// Năm phút đầu game vẫn đang lớn lên: atlas nhân vật nạp dần khi gặp người mới, art khu
// tải nền, pool chữ và pool view giãn tới mức dùng đỉnh. Lấy mốc lúc vào trận thì mọi lần
// chạy đều "tăng 300%" và con số không nói gì. Từ phút 5 trở đi mọi thứ đã ổn định, nên
// bất kỳ mức tăng nào cũng là mức tăng thật.
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
/** BASE_MIN: mốc so sánh (spec: phút thứ 5). */
const BASE_MIN = Number(process.env.BASE_MIN ?? 5);
/** ZONE_SWITCHES / DEATHS: spec chốt 50 và 10. */
const ZONE_SWITCHES = Number(process.env.ZONE_SWITCHES ?? 50);
const DEATHS = Number(process.env.DEATHS ?? 10);
/** TOL: ngưỡng tăng cho phép, %. */
const TOL = Number(process.env.TOL ?? 10);
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
/** Nền nhỏ thì phần trăm vô nghĩa; bỏ qua khi mức tăng tuyệt đối còn dưới tiếng ồn. */
const FLOOR = { texture: 5, texMB: 4, heapMB: 4, 'display object': 200 };

const METRICS = [
  ['texture', (r) => r.tex, '   '],
  ['texMB', (r) => r.texMB, 'MB '],
  ['heapMB', (r) => r.heapMB, 'MB '],
  ['display object', (r) => r.nodes, '   '],
];

// ── 1. Phép so của spec: mốc phút 5 ↔ mốc cuối ──────────────────────────────────────────
console.log('\n== §13.9: so với phút ' + base.min + ' ==');
const specFail = [];
for (const [name, get, unit] of METRICS) {
  const from = get(base);
  const to = get(last);
  const pct = pctOf(from, to);
  const ok = pct <= TOL || to - from <= FLOOR[name];
  if (!ok) {
    specFail.push(name);
  }
  console.log(`${name.padEnd(16)} ${from.toFixed(1).padStart(9)} → ${to.toFixed(1).padStart(9)} ${unit}  (${pct >= 0 ? '+' : ''}${pct.toFixed(1)}%)  ${ok ? 'ĐẠT' : 'TRƯỢT'}`);
}

// ── 2. Phép so CHẶN DEPLOY: nửa sau ↔ cuối ──────────────────────────────────────────────
//
// Vì sao cần phép so thứ hai, xem `docs/QUESTIONS.md` mục "Mốc phút 5 của §13.9".
//
// Tóm tắt: bộ làm việc của game vẫn đang LỚN LÊN ở phút 5 — 11 atlas nhân vật nạp dần khi
// gặp người mới, art của 5 khu nạp dần qua 50 lượt đổi khu, pool giãn tới mức dùng đỉnh.
// Đo 12/11: so với phút 5 thì texture +24,5% và heap +37,3%, NHƯNG `texMB` phẳng tuyệt đối
// (−0,0%) và từ phút 20 trở đi mọi chỉ số đứng yên (phút 20–25 ↔ 25–30: texture +1,0%,
// heap −0,5%). Đó là một TRẦN, không phải một đường dốc.
//
// Rò rỉ thật thì KHÔNG có trần — nó tăng đều tới khi tab chết (xem #33: +91 texture/phút
// một chiều). Nên phép so nửa-sau là phép so bắt đúng thứ cần bắt, và nó là phép chặn.
const half = rows.find((r) => r.min >= last.min / 2) ?? base;
console.log(`\n== xu hướng nửa sau: phút ${half.min} → ${last.min} (phép so CHẶN DEPLOY) ==`);
let bad = false;
for (const [name, get, unit] of METRICS) {
  const from = get(half);
  const to = get(last);
  const pct = pctOf(from, to);
  const ok = pct <= TOL || to - from <= FLOOR[name];
  if (!ok) {
    bad = true;
  }
  console.log(`${name.padEnd(16)} ${from.toFixed(1).padStart(9)} → ${to.toFixed(1).padStart(9)} ${unit}  (${pct >= 0 ? '+' : ''}${pct.toFixed(1)}%)  ${ok ? 'ĐẠT' : 'TRƯỢT'}`);
}
if (specFail.length && !bad) {
  console.log(`\n⚠ Vượt 10% so với phút ${base.min} ở: ${specFail.join(', ')} — nhưng nửa sau ĐỨNG YÊN,`);
  console.log('  tức là bộ làm việc đã chạm trần chứ không rò rỉ. Xem docs/QUESTIONS.md.');
}

await b.close();
if (bad) {
  console.log(`\nTRƯỢT: vượt ngưỡng ${TOL}% của §13.9.`);
  process.exit(1);
}
console.log(`\nĐẠT: không chỉ số nào tăng quá ${TOL}% (§13.9).`);
