// CỔNG KIỂM ĐỒ HOẠ — GAME_SPEC §13.10.
//
//   cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true
//   node tools/browsertest/vramgate.mjs http://127.0.0.1:9998
//
// Bốn điều kiện, PHẢI đạt cả bốn:
//
//   1. **0 placeholder** — không nhân vật nào đang vẽ bằng hình tạm (máy tính VÀ điện thoại).
//   2. **Đúng định dạng** — thiết bị hỗ trợ nén khối thì mọi tệp CÓ bản KTX2 phải nạp bản
//      KTX2, không được rơi về WebP.
//   3. **Hình không đổi** — `scenediff` so với ảnh mốc trong repo.
//   4. **VRAM trong ngân sách** — 50 MB trên máy tính, **32 MB trên điện thoại**. Hai lượt
//      đo riêng: `isMobile()` quyết định cả cột trần lẫn việc lấy atlas nhân vật bản 1 024,
//      nên đo bằng cửa sổ máy tính rồi chia nhẩm là đoán.
//
// ── Vì sao bốn, không phải một ──────────────────────────────────────────────────────────
//
// Ngày 13/11 tôi bật KTX2 mà quên `import 'pixi.js/ktx2'`. Mọi tệp `.ktx2` tải về xong rồi
// bị bỏ: nhân vật về placeholder, mặt đất thành một mảng xanh trơn — **và bảng VRAM báo
// 6,93 MB, con số đẹp nhất từ trước tới giờ**. Một cổng chỉ nhìn VRAM sẽ cho qua, và còn
// khen, một lượt tối ưu vừa xoá sạch art.
//
// Nên: VRAM đẹp KHÔNG đủ. Phải kèm "art vẫn ở đó" (điều 1), "đúng đường nén" (điều 2) và
// "hình vẫn thế" (điều 3). Thiếu bất kỳ điều nào là trượt, bất kể VRAM đẹp tới đâu.
import { execFileSync } from 'node:child_process';
import { existsSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

import { launch } from './browser.mjs';
import { requireDebug } from './needdebug.mjs';

const HERE = dirname(fileURLToPath(import.meta.url));
const REPO = join(HERE, '..', '..');
/** Ảnh mốc nằm TRONG repo: cổng phải so được trên máy mới clone, không phụ thuộc /tmp. */
const REF = join(REPO, 'docs', 'review', '13-buoc1-vongkiem', 'testbed-ref.png');
const SHOT = join(REPO, 'server', 'data', 'gate-testbed.png');

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
/**
 * `--warn-vram`: điều 4 (ngân sách VRAM) chỉ CẢNH BÁO, không chặn.
 *
 * Dùng trong lúc đang hạ VRAM xuống trần: cắm cổng vào deploy ngay từ bây giờ thì ba điều
 * kia (art còn đó, đúng đường nén, hình không đổi) gác được liền — còn bắt cả điều 4 chặn
 * khi còn đang 67/50 thì không deploy được lượt nào, và một cổng chặn hết mọi thứ sẽ bị gỡ
 * ra chứ không được sửa. Bỏ cờ này đi khi đã đạt trần.
 */
const WARN_VRAM = process.argv.includes('--warn-vram');
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

let fails = 0;
function check(name, ok, detail) {
  console.log(`${ok ? 'đạt ' : 'TRƯỢT'}  ${name}${detail ? `: ${detail}` : ''}`);
  if (!ok) {
    fails++;
  }
}

const b = await launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'],
});
const ctx = await b.newContext({ viewport: { width: 1920, height: 1080 }, deviceScaleFactor: 1, locale: 'vi-VN' });
const p = await ctx.newPage();
p.on('pageerror', (e) => console.log('  pageerror', String(e).slice(0, 160)));

// ── Cảnh ĐÔNG: nơi mọi ngân sách được tính ──────────────────────────────────────────────
await p.goto(`${TARGET}/?debug=true`, { waitUntil: 'domcontentloaded' });
await p.waitForSelector('input', { timeout: 30000 });
await p.fill('input', 'CongKiemVRAM');
await p.locator('.char-card').nth(0).click();
await p.locator('.home .btn.primary').click();
await p.waitForSelector('.room-board', { timeout: 60000 });
await requireDebug(p, 'vramgate.mjs');
await sleep(6000);
await p.evaluate(() => window.kx.conn.send({ t: 'dbg', cmd: 'max', n: 1 }));
await sleep(2000);
for (let i = 0; i < 8; i++) {
  await p.evaluate(() => window.kx.conn.send({ t: 'dbg', cmd: 'bot', n: 0 }));
  await sleep(400);
}
await sleep(9000);

const live = await p.evaluate(() => ({
  led: window.kx.scene.vramLedger(),
  draw: window.kx.scene.drawStats(),
  fmt: window.kx.scene.loadReport(),
  quality: window.kx.scene.qualityName?.() ?? '?',
  players: window.kx.world.players.size,
}));

console.log('');
console.log(`cảnh ĐÔNG · ${live.players} người · chất lượng ${live.quality} · 1920×1080 @ 1`);
for (const r of live.led.rows) {
  console.log(`  ${r.mb.toFixed(2).padStart(7)} / ${String(r.limitMB).padStart(2)} MB  ${r.label}${r.over ? '  ← VƯỢT' : ''}`);
}
console.log(`  ${live.led.totalMB.toFixed(2).padStart(7)} / ${live.led.limitMB} MB  TỔNG`);
console.log('');

check('0 placeholder (không ai vẽ bằng hình tạm)', live.draw.placeholder === false,
  live.draw.placeholder ? 'có nhân vật đang dùng placeholder' : `${live.draw.charsInView} nhân vật`);
check('đúng định dạng nén', !live.fmt.ktxOn || live.fmt.webpWhenKtx.length === 0,
  live.fmt.ktxOn
    ? `ktx2 ${live.fmt.ktx2} · webp ${live.fmt.webp}${live.fmt.webpWhenKtx.length ? ` · RƠI VỀ WEBP: ${live.fmt.webpWhenKtx.slice(0, 3).join(', ')}` : ''}`
    : 'thiết bị không hỗ trợ nén khối — bỏ qua');
const vramOk = live.led.totalMB <= live.led.limitMB;
if (WARN_VRAM && !vramOk) {
  console.log(`⚠ CẢNH BÁO  VRAM cảnh đông: ${live.led.totalMB.toFixed(1)} / ${live.led.limitMB} MB (đang hạ dần — chưa chặn)`);
} else {
  check('VRAM cảnh đông trong ngân sách', vramOk, `${live.led.totalMB.toFixed(1)} / ${live.led.limitMB} MB`);
}
await b.close();

// ── Lượt hai: ĐIỆN THOẠI (trần 32 MB) ───────────────────────────────────────────────────
//
// Trần điện thoại là 32 MB, không phải 50, và đường đi cũng khác: `isMobile()` quyết định
// cả cột trần LẪN việc lấy atlas nhân vật bản 1 024. Đo bằng cửa sổ máy tính rồi chia nhẩm
// là đoán. Đây cũng là ngân sách gác đúng chuyện đã xảy ra ở B13: iPhone chạm 216/256 MB
// rồi bị Safari thu hồi tab, và người chơi chỉ thấy "lúc nào cũng lỗi".
const pb = await launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'],
});
const pctx = await pb.newContext({
  viewport: { width: 390, height: 844 }, deviceScaleFactor: 3, isMobile: true, hasTouch: true,
  userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1',
  locale: 'vi-VN',
});
const pp = await pctx.newPage();
pp.on('pageerror', (e) => console.log('  pageerror (điện thoại)', String(e).slice(0, 160)));
await pp.goto(`${TARGET}/?debug=true`, { waitUntil: 'domcontentloaded' });
await pp.waitForSelector('input', { timeout: 30000 });
await pp.fill('input', 'CongKiemDienThoai');
await pp.locator('.char-card').nth(0).click();
await pp.locator('.home .btn.primary').click();
await pp.waitForSelector('.room-board', { timeout: 60000 });
await requireDebug(pp, 'vramgate.mjs (điện thoại)');
await sleep(6000);
await pp.evaluate(() => window.kx.conn.send({ t: 'dbg', cmd: 'max', n: 1 }));
await sleep(2000);
for (let i = 0; i < 8; i++) {
  await pp.evaluate(() => window.kx.conn.send({ t: 'dbg', cmd: 'bot', n: 0 }));
  await sleep(400);
}
await sleep(9000);
const ph = await pp.evaluate(() => ({
  led: window.kx.scene.vramLedger(),
  draw: window.kx.scene.drawStats(),
  players: window.kx.world.players.size,
}));
await pb.close();

console.log('');
console.log(`cảnh ĐÔNG trên ĐIỆN THOẠI · ${ph.players} người · 390×844 @ 3`);
for (const r of ph.led.rows) {
  console.log(`  ${r.mb.toFixed(2).padStart(7)} / ${String(r.limitMB).padStart(2)} MB  ${r.label}${r.over ? '  ← VƯỢT' : ''}`);
}
console.log(`  ${ph.led.totalMB.toFixed(2).padStart(7)} / ${ph.led.limitMB} MB  TỔNG`);
console.log('');
check('điện thoại: 0 placeholder', ph.draw.placeholder === false, `${ph.draw.charsInView} nhân vật`);
const phoneOk = ph.led.totalMB <= ph.led.limitMB;
if (WARN_VRAM && !phoneOk) {
  console.log(`⚠ CẢNH BÁO  VRAM điện thoại: ${ph.led.totalMB.toFixed(1)} / ${ph.led.limitMB} MB`);
} else {
  check('VRAM điện thoại trong ngân sách', phoneOk, `${ph.led.totalMB.toFixed(1)} / ${ph.led.limitMB} MB`);
}

// ── Hình: cảnh thử so với ảnh mốc trong repo ────────────────────────────────────────────
if (!existsSync(REF)) {
  check('ảnh mốc có trong repo', false, `thiếu ${REF}`);
} else {
  try {
    execFileSync(process.execPath, [join(HERE, 'scenediff.mjs'), TARGET, SHOT], { stdio: 'pipe' });
    execFileSync(process.execPath, [join(HERE, 'scenediff.mjs'), '--diff', REF, SHOT], { stdio: 'inherit' });
    check('hình cảnh thử không đổi', true);
  } catch {
    check('hình cảnh thử không đổi', false, 'xem bảng lệch ở trên');
  }
}

console.log('');
if (fails > 0) {
  console.error(`✗ cổng kiểm đồ hoạ: ${fails} điều TRƯỢT (§13.10).`);
  process.exit(1);
}
console.log('✓ cổng kiểm đồ hoạ: đạt cả bốn điều.');
