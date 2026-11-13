// ĐƯỜNG LÙI WEBP — máy KHÔNG hỗ trợ texture nén khối vẫn phải chơi được (§13.10).
//
//   node tools/browsertest/nocompress.mjs http://127.0.0.1:9998
//
// `assets.ts` chọn KTX2 khi `ktxSupported()` thấy một trong BC7/S3TC/ASTC/ETC. Máy không có
// cái nào — GPU cũ, trình duyệt lạ, máy ảo phần mềm — phải rơi về WebP. Đường đó CHƯA AI THỬ
// kể từ khi bật KTX2 ngày 13/11, và nó là một trong những nghi phạm của vụ đứng hình hôm đó
// (hoá ra không phải, nhưng "hoá ra không phải" chỉ nói được sau khi thử).
//
// Bài này chặn đúng chỗ `ktxSupported()` hỏi: `getExtension` trả null cho mọi phần mở rộng
// nén. Rồi canh: vào trận được, 0 placeholder, 0 lỗi vẽ, và KHÔNG tệp .ktx2 nào được tải.
import { launch } from './browser.mjs';
import { requireDebug } from './needdebug.mjs';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

const b = await launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'],
});
const ctx = await b.newContext({ viewport: { width: 1280, height: 720 }, locale: 'vi-VN' });
const p = await ctx.newPage();
p.on('pageerror', (e) => console.log('  pageerror', String(e).slice(0, 200)));

// Giấu mọi phần mở rộng nén khối, trên CẢ hai lớp ngữ cảnh WebGL.
await p.addInitScript(() => {
  const BLOCK = /compressed_texture|texture_compression/i;
  for (const C of [WebGLRenderingContext, WebGL2RenderingContext]) {
    const orig = C.prototype.getExtension;
    C.prototype.getExtension = function (name) {
      return BLOCK.test(String(name)) ? null : orig.call(this, name);
    };
    const origList = C.prototype.getSupportedExtensions;
    C.prototype.getSupportedExtensions = function () {
      return (origList.call(this) ?? []).filter((n) => !BLOCK.test(n));
    };
  }
});

const ktx = [];
p.on('request', (r) => {
  if (r.url().endsWith('.ktx2')) {
    ktx.push(r.url().split('/').pop());
  }
});

await p.goto(`${TARGET}/?debug=true`, { waitUntil: 'domcontentloaded' });
await p.waitForSelector('input', { timeout: 30000 });
await p.fill('input', 'KiemTraKhongNen');
await p.locator('.char-card').nth(0).click();
await p.locator('.home .btn.primary').click();
await p.waitForSelector('.room-board', { timeout: 60000 });
await requireDebug(p, 'nocompress.mjs');
await p.evaluate(() => {
  window.__kxDraw = 0;
  const prev = window.kx.scene.onGuardEvent;
  window.kx.scene.onGuardEvent = (k, e) => {
    if (k === 'draw_error') {
      window.__kxDraw++;
    }
    prev?.(k, e);
  };
});
await sleep(9000);
// Đổi một khu: đây là đường đã làm hỏng build 13/11, và nó cũng ép tải thêm asset.
await p.evaluate(() => window.kx.conn.send({ t: 'dbg', cmd: 'zone', n: 2 }));
await sleep(6000);

const got = await p.evaluate(() => ({
  draw: window.__kxDraw,
  fmt: window.kx.scene.loadReport(),
  stats: window.kx.scene.drawStats(),
}));
await b.close();

let fails = 0;
const check = (name, ok, detail) => {
  console.log(`${ok ? 'đạt ' : 'TRƯỢT'}  ${name}${detail ? `: ${detail}` : ''}`);
  if (!ok) {
    fails++;
  }
};
console.log('');
check('nhận ra máy KHÔNG hỗ trợ nén khối', got.fmt.ktxOn === false, `ktxOn=${got.fmt.ktxOn}`);
check('không tải tệp .ktx2 nào', ktx.length === 0, ktx.length ? `${ktx.length} tệp: ${ktx.slice(0, 3).join(', ')}` : 'đúng, chỉ webp');
check('0 placeholder', got.stats.placeholder === false, `${got.stats.charsInView} nhân vật`);
check('0 lỗi vẽ (kể cả lúc đổi khu)', got.draw === 0, String(got.draw));
console.log('');
if (fails > 0) {
  console.error(`✗ đường lùi WebP: ${fails} điều TRƯỢT — máy không hỗ trợ nén khối sẽ không chơi được.`);
  process.exit(1);
}
console.log('✓ đường lùi WebP chạy đúng.');
