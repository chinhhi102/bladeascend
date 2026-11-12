// LÔ VẼ THEO TỪNG LỚP — mục 5 lượt tối ưu (B30).
//
//   cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true
//   node tools/browsertest/drawbreak.mjs http://127.0.0.1:9998            # cảnh THẬT
//   node tools/browsertest/drawbreak.mjs http://127.0.0.1:9998 --testbed  # cảnh THỬ
//
// Tổng số lô vẽ nói "đang tốn", không nói "tốn ở đâu". Bảng này nói từng lớp tốn mấy lô, và
// kèm những thứ GÂY ngắt lô để biết chữa kiểu gì:
//
//   - `texture` — số NGUỒN texture khác nhau trong lớp. Nhiều nguồn ⇒ gom vào ít atlas hơn.
//   - `blend`   — số blend mode khác nhau. Đổi blend là ngắt lô, không gom được; cách chữa
//                 là xếp cùng blend cạnh nhau hoặc tách thành lớp riêng.
//   - `graphics`/`text`/`mesh` — vật không gom lô được. Graphics TĨNH thì nướng thành sprite.
//   - `filter`/`mask` — mỗi cái là một lượt vẽ ra texture riêng, đắt nhất bảng.
//
// Cảnh THẬT cho biết người chơi đang gặp gì; cảnh THỬ là chỗ duy nhất so được hai bản build
// (xem `scenebench.mjs`).
import { launch } from './browser.mjs';
import { requireDebug } from './needdebug.mjs';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
const TESTBED = process.argv.includes('--testbed');
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

const b = await launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'],
});
const ctx = await b.newContext({ viewport: { width: 1400, height: 820 }, locale: 'vi-VN' });
const p = await ctx.newPage();
p.on('pageerror', (e) => console.log('  pageerror', String(e).slice(0, 160)));

let rows;
let label;
if (TESTBED) {
  await p.goto(`${TARGET}/?scene=testbed&debug=true`, { waitUntil: 'domcontentloaded' });
  await p.waitForFunction(() => window.kxTestbed?.ready === true, { timeout: 90000 });
  await sleep(2500);
  rows = await p.evaluate(() => window.kxTestbed.layers());
  label = 'cảnh THỬ cố định';
} else {
  await p.goto(`${TARGET}/?debug=true`, { waitUntil: 'domcontentloaded' });
  await p.waitForSelector('input', { timeout: 30000 });
  await p.fill('input', 'KiemTraLoVe');
  await p.locator('.char-card').nth(0).click();
  await p.locator('.home .btn.primary').click();
  await p.waitForSelector('.room-board', { timeout: 60000 });
  await requireDebug(p, 'drawbreak.mjs');
  await sleep(5000);
  // Đạo Tổ đủ kiếm + bất tử: cảnh nặng nhất, và không chết giữa lúc đo.
  await p.evaluate(() => window.kx.conn.send({ t: 'dbg', cmd: 'max', n: 1 }));
  await sleep(4000);
  const got = await p.evaluate(() => window.kx.scene.layerStats());
  rows = got.layers;
  const n = await p.evaluate(() => window.kx.world.players.size);
  label = `cảnh THẬT · ${n} người trong tầm`;
}

const k = (r, name) => r.kinds?.[name] ?? 0;
console.log('');
console.log(`| lớp | lô vẽ | nút | texture | blend | graphics | text | mesh | filter | mask |`);
console.log(`|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|`);
let total = 0;
for (const r of [...rows].sort((a, c) => c.calls - a.calls)) {
  total += r.calls;
  console.log(`| ${r.name} | ${r.calls} | ${r.nodes} | ${r.textures} | ${r.blends} | ${k(r, 'graphics')} | ${k(r, 'text') + k(r, 'bitmapText')} | ${k(r, 'mesh')} | ${r.filters} | ${r.masks} |`);
}
console.log(`| **tổng** | **${total}** | | | | | | | | |`);
console.log('');
console.log(`${label} · loại ống vẽ gặp trong cảnh: ${[...new Set(rows.flatMap((r) => Object.keys(r.kinds ?? {})))].join(', ')}`);
await b.close();
