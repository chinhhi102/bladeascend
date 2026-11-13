// SỔ VRAM THEO HẠNG MỤC NGÂN SÁCH — GAME_SPEC §13.10.
//
//   cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true
//   node tools/browsertest/vram.mjs http://127.0.0.1:9998            # cảnh ĐÔNG (mặc định)
//   node tools/browsertest/vram.mjs http://127.0.0.1:9998 --testbed  # cảnh thử cố định
//
// Cảnh đo theo đúng §13.10: phòng đầy, nhiều loại nhân vật trên màn hình, có người Tiên
// giới, chất lượng **Vừa**, **1080p tỉ lệ điểm ảnh 1**.
//
// ── Vì sao 1080p và devicePixelRatio 1 ──────────────────────────────────────────────────
//
// Bộ đệm khung là một hạng mục ngân sách (≤ 12 MB). Nó tỉ lệ với SỐ ĐIỂM ẢNH THẬT, nên đo
// ở cỡ cửa sổ khác là đo một ngân sách khác: 1920×1080 ở tỉ lệ 1 tốn 8,3 MB mỗi bộ đệm,
// còn ở tỉ lệ 2 thì tốn 33 MB — vượt trần chỉ vì cái màn hình.
//
// ── Con số này KHÔNG phải "GPU Memory" của Chrome ───────────────────────────────────────
//
// Sổ này đếm những gì GAME tải lên (texture + render texture). Task Manager của Chrome đếm
// cả những thứ game không thấy: bộ đệm của trình duyệt, lớp hợp ảnh, tài nguyên của chính
// trình điều khiển. Spec đòi hai số khớp nhau ±10%, và phép đối chiếu đó phải làm bằng tay
// (Shift+Esc) — bài này in ra con số của sổ để so.
import { launch } from './browser.mjs';
import { requireDebug } from './needdebug.mjs';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
const TESTBED = process.argv.includes('--testbed');
// `--trace` bật `?textrace=1`: ghi cả VẾT GỌI của chỗ dựng texture. Đắt (mỗi lần tạo
// texture là một `new Error().stack`), nên chỉ bật khi đang đi tìm "cái 8 MB kia ở đâu".
const TRACE = process.argv.includes('--trace');
const q = TRACE ? '&textrace=1' : '';
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

const b = await launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'],
});
// 1080p, tỉ lệ điểm ảnh 1 — xem ghi chú đầu tệp.
const ctx = await b.newContext({ viewport: { width: 1920, height: 1080 }, deviceScaleFactor: 1, locale: 'vi-VN' });
const p = await ctx.newPage();
p.on('pageerror', (e) => console.log('  pageerror', String(e).slice(0, 160)));

let led;
let label;
if (TESTBED) {
  await p.goto(`${TARGET}/?scene=testbed&debug=true${q}`, { waitUntil: 'domcontentloaded' });
  await p.waitForFunction(() => window.kxTestbed?.ready === true, { timeout: 90000 });
  await sleep(3000);
  led = await p.evaluate(() => window.kxTestbed.vram?.() ?? null);
  label = 'cảnh THỬ cố định';
} else {
  await p.goto(`${TARGET}/?debug=true${q}`, { waitUntil: 'domcontentloaded' });
  await p.waitForSelector('input', { timeout: 30000 });
  await p.fill('input', 'KiemTraVRAM');
  await p.locator('.char-card').nth(0).click();
  await p.locator('.home .btn.primary').click();
  await p.waitForSelector('.room-board', { timeout: 60000 });
  await requireDebug(p, 'vram.mjs');
  await sleep(6000);
  // Cảnh ĐÔNG theo §13.10: lên Đạo Tổ (có hào quang Tiên giới) và bất tử để cảnh đứng yên
  // suốt lượt đo, rồi thả thêm bot cho đủ loại nhân vật trên màn hình.
  await p.evaluate(() => window.kx.conn.send({ t: 'dbg', cmd: 'max', n: 1 }));
  await sleep(2000);
  for (let i = 0; i < 8; i++) {
    await p.evaluate(() => window.kx.conn.send({ t: 'dbg', cmd: 'bot', n: 0 }));
    await sleep(400);
  }
  // Chất lượng VỪA: `?q=` không có, mặc định là Vừa trên máy bàn — xác nhận lại cho chắc.
  await sleep(8000);
  const info = await p.evaluate(() => {
    const s = window.kx.scene;
    return { led: s.vramLedger(), quality: s.qualityName?.() ?? '?', players: window.kx.world.players.size };
  });
  led = info.led;
  label = `cảnh ĐÔNG · ${info.players} người trong tầm · chất lượng ${info.quality}`;
}

if (!led) {
  console.error('KHÔNG ĐỌC ĐƯỢC SỔ: thiếu `vramLedger()`.');
  await b.close();
  process.exit(1);
}

console.log('');
console.log('| Hạng mục | MB | texture | trần | |');
console.log('|---|---:|---:|---:|---|');
for (const r of led.rows) {
  console.log(`| ${r.label} | ${r.mb.toFixed(2)} | ${r.count} | ${r.limitMB} | ${r.over ? '**VƯỢT**' : 'đạt'} |`);
}
console.log(`| **TỔNG** | **${led.totalMB.toFixed(2)}** | | **${led.limitMB}** | ${led.totalMB > led.limitMB ? '**VƯỢT**' : 'đạt'} |`);
if (led.unclassified.length) {
  console.log('');
  console.log('Chưa khai hạng mục (hạng mục `KHÁC`):');
  for (const u of led.unclassified) {
    console.log(`  ${u.mb.toFixed(2)} MB  ${u.site}`);
  }
}
console.log('');
console.log(`${label} · 1920×1080 @ tỉ lệ điểm ảnh 1`);
console.log('Đối chiếu tay: Chrome Shift+Esc → cột "GPU Memory" của tab game (spec đòi khớp ±10%).');
await b.close();
