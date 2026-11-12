// Nghiệm thu lớp CHỐNG SẬP bằng LỖI GIẢ — GAME_SPEC §13.8, checklist §17 dòng 1:
//
//   "Ném lỗi giả vào từng hệ thống vẽ (VFX, bảng tên, minimap…) → game chạy tiếp, chỉ hệ
//    thống đó tắt; lỗi giả trong renderer 3 lần → tự dựng lại renderer, không tải trang."
//
//   cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true
//   node tools/browsertest/crashtest.mjs http://127.0.0.1:9998
//
// Vì sao phải là một bài đo chạy được, không phải một lần thử tay: lớp chống sập chỉ chạy
// đúng vào ngày có lỗi thật, tức là ngày không ai ngồi xem. Thứ duy nhất giữ cho nó còn
// hoạt động sau sáu tháng là một bài bấm nút được.
import { launch } from './browser.mjs';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

/** Hệ thống đem ra thử. Đủ các loại: nội suy, entity, hiệu ứng, HUD, âm thanh. */
const SYSTEMS = ['vfx', 'beasts', 'minimap', 'items', 'monsters', 'ambient', 'skills', 'emote', 'audio'];

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
const errors = [];
p.on('pageerror', (e) => errors.push(String(e).slice(0, 160)));

await p.goto(`${TARGET}/?debug=true`, { waitUntil: 'domcontentloaded' });
await p.waitForSelector('input', { timeout: 30000 });
await p.fill('input', 'KiemTraChongSap');
await p.locator('.char-card').nth(0).click();
await p.locator('.home .btn.primary').click();
await p.waitForSelector('.room-board', { timeout: 60000 });
await sleep(4000);

const stats = () => p.evaluate(() => window.__stats());
const guard = () => p.evaluate(() => window.kx.guard());
/** frames đếm khung bằng chính FPS của game: game còn vẽ thì con số này còn nhúc nhích. */
const alive = async (secs) => {
  const a = await stats();
  await sleep(secs * 1000);
  const c = await stats();
  return c.fps > 5 && a.nodes > 0 && c.nodes > 0;
};

console.log('— 1. Lỗi giả một lần vào từng hệ thống: game phải chạy tiếp —');
for (const key of SYSTEMS) {
  await p.evaluate((k) => window.kx.fault(k, 1), key);
  await sleep(700);
  const g = await guard();
  const hit = (g.systems ?? []).find((s) => s.key === key);
  expect(`${key}: đã bắt được lỗi`, !!hit && hit.errors >= 1, true);
  expect(`${key}: CHƯA tắt hẳn (mới 1 lần)`, hit?.dead ?? false, false);
}
expect('game vẫn vẽ sau 9 lỗi giả', await alive(2), true);

console.log('\n— 2. Lỗi 3 lần/10 s ở cùng hệ thống: tắt HẲN hệ thống đó, các hệ thống khác sống —');
await p.evaluate(() => window.kx.fault('vfx', 3));
// Ba lỗi cách nhau 2 giây (sau mỗi lỗi hệ thống nghỉ 2 s rồi mới thử lại).
await sleep(7000);
const g2 = await guard();
expect('vfx đã tắt hẳn', (g2.systems ?? []).find((s) => s.key === 'vfx')?.dead ?? false, true);
expect('minimap KHÔNG bị tắt lây', (g2.systems ?? []).find((s) => s.key === 'minimap')?.dead ?? false, false);
expect('game vẫn vẽ', await alive(2), true);

console.log('\n— 3. Lỗi giả trong khâu VẼ 3 lần: tự dựng lại renderer, KHÔNG tải lại trang —');
// Đánh dấu trang: nếu nó tải lại thì dấu này mất.
await p.evaluate(() => {
  window.__khongTaiLai = true;
});
const before = await guard();
await p.evaluate(() => window.kx.faultDraw(3));
await sleep(6000);
const after = await guard();
expect('đã ghi đủ 3 lỗi vẽ', (after.draw?.errors ?? 0) - (before.draw?.errors ?? 0) >= 3, true);
expect('đã dựng lại renderer', (after.draw?.rebuilds ?? 0) > (before.draw?.rebuilds ?? 0), true);
expect('KHÔNG tải lại trang', await p.evaluate(() => window.__khongTaiLai === true), true);
expect('vẫn còn trong trận', await p.evaluate(() => !!document.querySelector('.room-board')), true);
expect('game vẫn vẽ sau khi dựng lại', await alive(3), true);

console.log('\n— 4. Panic giả ở SERVER: 1 khu bỏ 1 tick, người khác không bị ngắt —');
await p.evaluate(() => window.kx.conn.send({ t: 'dbg', cmd: 'panic', n: 0 }));
await sleep(2500);
expect('vẫn trong trận sau panic tick', await p.evaluate(() => !!document.querySelector('.room-board')), true);
expect('game vẫn vẽ', await alive(2), true);

console.log('\n— 5. Panic giả khi xử lý GÓI của mình: chỉ mình bị ngắt —');
await p.evaluate(() => window.kx.conn.send({ t: 'dbg', cmd: 'panic', n: 1 }));
await sleep(4000);
// Server ngắt kết nối người đó; client tự nối lại và vào lại phiên (§7.1) hoặc về trang chủ.
const back = await p.evaluate(() => !!document.querySelector('.room-board') || !!document.querySelector('.home') || !!document.querySelector('.screen'));
expect('client xử lý được việc bị ngắt (không treo)', back, true);

if (errors.length) {
  console.log(`\n${errors.length} lỗi lọt ra window.onerror (lớp chống sập nên đã nuốt hết):`);
  for (const e of [...new Set(errors)].slice(0, 6)) {
    console.log(`  ${e}`);
  }
}

await b.close();
if (fails > 0) {
  console.error(`\n${fails} điều TRƯỢT`);
  process.exit(1);
}
console.log('\nLớp chống sập (§13.8 / checklist §17): đạt hết');
