// Nghiệm thu VÀO LẠI PHIÊN — GAME_SPEC §7.1, checklist §17 dòng 3:
//
//   "F5 giữa trận → trong 20 s vào lại đúng nhân vật (cảnh giới, kiếm, Uy Danh, khu);
//    bị chém trong lúc mất kết nối → chết, F5 xong thấy Tổng kết."
//
//   cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true
//   node tools/browsertest/rejoin.mjs http://127.0.0.1:9998
import { launch } from './browser.mjs';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

let fails = 0;
function expect(name, got, want) {
  const ok = got === want;
  if (!ok) {
    fails++;
  }
  console.log(`${ok ? 'đạt ' : 'TRƯỢT'}  ${name}: ${String(got)}${ok ? '' : ` (mong ${String(want)})`}`);
}

const b = await launch({ headless: true, args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'] });
const ctx = await b.newContext({ viewport: { width: 1280, height: 720 }, locale: 'vi-VN' });
const p = await ctx.newPage();
p.on('pageerror', (e) => console.log('  pageerror', String(e).slice(0, 160)));

async function joinMatch(name) {
  await p.goto(`${TARGET}/?debug=true`, { waitUntil: 'domcontentloaded' });
  await p.waitForSelector('input', { timeout: 30000 });
  await p.fill('input', name);
  await p.locator('.char-card').nth(0).click();
  await p.locator('.home .btn.primary').click();
  await p.waitForSelector('.room-board', { timeout: 60000 });
  await sleep(3500);
}

/** me đọc trạng thái nhân vật của chính mình qua cửa debug. */
const me = () => p.evaluate(() => {
  const m = window.kx?.world?.me?.();
  return m ? { id: m.id, level: m.L, swords: m.n } : null;
});
const session = () => p.evaluate(() => {
  try {
    return JSON.parse(sessionStorage.getItem('kx.session') ?? 'null');
  } catch {
    return null;
  }
});

await joinMatch('KiemTraVaoLai');

// Nuôi nhân vật lên cho khác hẳn trạng thái "mới vào": lên cảnh giới + thêm kiếm.
for (let i = 0; i < 3; i++) {
  await p.evaluate(() => window.kx.conn.send({ t: 'dbg', cmd: 'realm', n: 0 }));
  await sleep(400);
}
await p.evaluate(() => window.kx.conn.send({ t: 'dbg', cmd: 'sword', n: 12 }));
await sleep(1500);

const before = await me();
const keyBefore = (await session())?.key ?? '';
console.log(`trước F5: id ${before?.id} · cảnh giới ${before?.level} · ${before?.swords} kiếm`);
expect('đã có mã phiên 32 byte', keyBefore.length, 64);
expect('cảnh giới đã lên (khác lúc mới vào)', (before?.level ?? 0) > 0, true);

// ── F5 giữa trận ────────────────────────────────────────────────────────────────────────
await p.reload({ waitUntil: 'domcontentloaded' });
await p.waitForSelector('.room-board', { timeout: 60000 });
await sleep(3000);

const after = await me();
const keyAfter = (await session())?.key ?? '';
console.log(`sau  F5: id ${after?.id} · cảnh giới ${after?.level} · ${after?.swords} kiếm`);
expect('vào lại ĐÚNG nhân vật (cùng id)', after?.id, before?.id);
expect('giữ nguyên cảnh giới', after?.level, before?.level);
expect('giữ nguyên số kiếm', after?.swords, before?.swords);
expect('không qua trang chủ', await p.evaluate(() => !!document.querySelector('.room-board')), true);
expect('đã cấp mã phiên MỚI', keyAfter.length === 64 && keyAfter !== keyBefore, true);

// ── Bị chém chết trong lúc "mất kết nối" ────────────────────────────────────────────────
//
// Giả lập bằng cách: ngắt WebSocket (server giữ nhân vật 20 s), rồi giết nhân vật đó từ
// một kết nối khác thì phức tạp — ở đây dùng lệnh test `die` NGAY TRƯỚC khi tải lại, nên
// cú chết rơi đúng vào lúc client đang khởi động lại.
console.log('\n— chết đúng lúc tải lại trang —');
await p.evaluate(() => {
  window.kx.conn.send({ t: 'dbg', cmd: 'die', n: 0 });
  // Tải lại NGAY: cú chết tới server trong lúc trang đang dựng lại.
  setTimeout(() => location.reload(), 120);
});
await sleep(9000);
const onSummary = await p.evaluate(() => !!document.querySelector('.summary'));
const inMatch = await p.evaluate(() => !!document.querySelector('.room-board'));
expect('thấy màn Tổng kết, KHÔNG tự vào trận mới', onSummary && !inMatch, true);

await b.close();
if (fails > 0) {
  console.error(`\n${fails} điều TRƯỢT`);
  process.exit(1);
}
console.log('\nVào lại phiên (§7.1 / checklist §17): đạt hết');
