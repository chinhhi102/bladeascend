// ĐỔI KHU MÀ KHÔNG ĐƯỢC CÓ LỖI VẼ — §13.8 + §8.6.
//
//   cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true
//   node tools/browsertest/drawguard.mjs http://127.0.0.1:9998
//
// ── Vì sao có bài này ───────────────────────────────────────────────────────────────────
//
// Ngày 13/11, ba người chơi thật trên build `b6c1c72` bị ĐỨNG HÌNH rồi VĂNG RA. Telemetry
// cho đúng một chữ ký, 95 lần, trên mọi build từ 14:45 trở đi:
//
//   Error: shader and geometry incompatible, geometry missing the "aLocal" attribute
//
// `aLocal` là thuộc tính của shader trộn nền. `GroundField` dựng mesh với một `MeshGeometry`
// TRỐNG (chỉ có aPosition/aUV); bốn thuộc tính riêng chỉ xuất hiện sau `setChunks` đầu tiên.
// Giữa hai thời điểm đó mà có một khung được vẽ là ném lỗi — và lúc đổi khu thì đúng là có:
// `MapView` mới dựng xong, art ô nền của khu mới chưa tải, `baseName` còn rỗng nên
// `updateGroundChunks` thoát sớm và KHÔNG BAO GIỜ gọi `setChunks`.
//
// Lỗi lặp mỗi khung → lớp chống sập §13.8 đếm 3 lỗi/10 giây → dựng lại renderer → vẫn lỗi →
// đường lùi cuối cùng là TẢI LẠI TRANG. Nhìn từ ghế người chơi: đứng hình vài giây rồi văng
// ra khỏi trận. Một nguyên nhân, hai triệu chứng.
//
// Bài này đổi khu đủ 5 khu và canh **0 lỗi vẽ**. Nó bắt được lỗi trên vì nó đi đúng đường
// mà `scenediff`/`drawbreak` không đi: hai bài kia đứng yên một chỗ trong một khu.
import { launch } from './browser.mjs';
import { requireDebug } from './needdebug.mjs';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
const ROUNDS = Number(process.env.ROUNDS ?? 6);
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

const b = await launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'],
});
const ctx = await b.newContext({ viewport: { width: 1280, height: 720 }, locale: 'vi-VN' });
const p = await ctx.newPage();
p.on('pageerror', (e) => console.log('  pageerror', String(e).slice(0, 200)));

await p.goto(`${TARGET}/?debug=true`, { waitUntil: 'domcontentloaded' });
await p.waitForSelector('input', { timeout: 30000 });
await p.fill('input', 'KiemTraDoiKhu');
await p.locator('.char-card').nth(0).click();
await p.locator('.home .btn.primary').click();
await p.waitForSelector('.room-board', { timeout: 60000 });
await requireDebug(p, 'drawguard.mjs');
await sleep(4000);

// Bắt lỗi vẽ ngay tại nguồn (`Scene.onGuardEvent`), không đợi vòng telemetry 30 giây.
await p.evaluate(() => {
  window.__kxDraw = [];
  const prev = window.kx.scene.onGuardEvent;
  window.kx.scene.onGuardEvent = (kind, err) => {
    window.__kxDraw.push({ kind, err: String(err?.message ?? err ?? '').slice(0, 200) });
    prev?.(kind, err);
  };
});

for (let i = 0; i < ROUNDS; i++) {
  const zone = i % 5;
  await p.evaluate((z) => window.kx.conn.send({ t: 'dbg', cmd: 'zone', n: z }), zone);
  // 2,5 giây: đủ cho khu mới dựng xong, tải nhóm asset của nó, và vẽ hàng trăm khung —
  // nếu mesh nền đang mang geometry sai thì mỗi khung một lỗi, không thể lọt.
  await sleep(2500);
}
await sleep(2000);

const got = await p.evaluate(() => window.__kxDraw);
const zones = await p.evaluate(() => window.kx.world?.zone ?? -1);
await b.close();

const byKind = {};
for (const g of got) {
  byKind[g.kind] = (byKind[g.kind] ?? 0) + 1;
}
console.log(`\nđổi khu ${ROUNDS} lượt · khu cuối ${zones}`);
console.log(`sự kiện lớp chống sập: ${got.length ? JSON.stringify(byKind) : 'không có'}`);
for (const g of got.slice(0, 5)) {
  console.log(`  ${g.kind}: ${g.err}`);
}
if (got.length > 0) {
  console.error(`\n✗ ĐỔI KHU SINH LỖI VẼ: ${got.length} sự kiện. Người chơi sẽ thấy đứng hình rồi văng ra.`);
  process.exit(1);
}
console.log('\n✓ đổi khu sạch: 0 lỗi vẽ.');
