// SO ẢNH cảnh thử cố định — trước/sau mỗi thay đổi đồ hoạ.
//
//   cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true
//   node tools/browsertest/scenediff.mjs http://127.0.0.1:9998 /tmp/truoc.png
//   # … sửa code, npm run build …
//   node tools/browsertest/scenediff.mjs http://127.0.0.1:9998 /tmp/sau.png
//   node tools/browsertest/scenediff.mjs --diff /tmp/truoc.png /tmp/sau.png
//
// Cảnh là `?scene=testbed` (`client/src/render/testbed.ts`): không mạng, không bot, không
// thời gian, camera ghim, cỡ khung và `resolution` cố định. Đọc ghi chú đầu tệp đó để biết
// vì sao từng thứ bị ghim — ba bài so ảnh trước nó đều vô dụng vì thiếu đúng mấy thứ này.
//
// Ảnh lưu thành PNG thật (xem được bằng mắt, gửi được cho chủ dự án). Phép so giải mã PNG
// bằng chính trình duyệt (data URL + `Image`), khỏi thêm thư viện vào repo.
//
// ── Ngưỡng ──────────────────────────────────────────────────────────────────────────────
//
// Đặt từ ĐỘ TỰ LẶP đo được, không đặt bằng cảm giác: chạy `--diff` giữa hai lần chụp CÙNG
// một bản build để biết biên nhiễu của chính bài đo, rồi mới tin con số trước/sau. Mặc định
// khít (0,3/255 · 0,1% pixel) vì cảnh đã xác định hoàn toàn; nới bằng biến môi trường
// `TOL_MEAN` / `TOL_PCT` nếu một máy khác có driver vẽ khác.
import { readFileSync, writeFileSync } from 'node:fs';

const TOL_MEAN = Number(process.env.TOL_MEAN ?? 0.3);
const TOL_PCT = Number(process.env.TOL_PCT ?? 0.1);
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

/** px đọc một PNG thành mảng độ xám, nhờ trình duyệt giải mã. */
async function px(page, file) {
  const b64 = readFileSync(file).toString('base64');
  return page.evaluate(async (data) => {
    const img = new Image();
    img.src = `data:image/png;base64,${data}`;
    await img.decode();
    const cv = document.createElement('canvas');
    cv.width = img.naturalWidth;
    cv.height = img.naturalHeight;
    const g = cv.getContext('2d');
    g.drawImage(img, 0, 0);
    const d = g.getImageData(0, 0, cv.width, cv.height).data;
    const out = new Array(cv.width * cv.height);
    for (let k = 0; k < out.length; k++) {
      out[k] = Math.round(d[k * 4] * 0.299 + d[k * 4 + 1] * 0.587 + d[k * 4 + 2] * 0.114);
    }
    return { w: cv.width, h: cv.height, g: out };
  }, b64);
}

const { launch } = await import('./browser.mjs');

if (process.argv[2] === '--diff') {
  const b = await launch({ headless: true });
  const page = await (await b.newContext()).newPage();
  const a = await px(page, process.argv[3]);
  const c = await px(page, process.argv[4]);
  await b.close();
  if (a.w !== c.w || a.h !== c.h) {
    console.error(`KHÔNG SO ĐƯỢC: ${a.w}×${a.h} với ${c.w}×${c.h}. Cảnh thử phải cố định cỡ khung.`);
    process.exit(1);
  }
  let sum = 0;
  let worst = 0;
  let over = 0;
  for (let i = 0; i < a.g.length; i++) {
    const d = Math.abs(a.g[i] - c.g[i]);
    sum += d;
    worst = Math.max(worst, d);
    if (d > 16) {
      over++;
    }
  }
  const mean = sum / a.g.length;
  const pct = (over / a.g.length) * 100;
  console.log(`cảnh thử ${a.w}×${a.h} · ${a.g.length} pixel`);
  console.log(`lệch trung bình : ${mean.toFixed(3)}/255  (ngưỡng ${TOL_MEAN})`);
  console.log(`lệch lớn nhất   : ${worst}/255`);
  console.log(`pixel lệch > 16 : ${over} (${pct.toFixed(3)}%, ngưỡng ${TOL_PCT}%)`);
  const ok = mean <= TOL_MEAN && pct <= TOL_PCT;
  console.log(ok ? '\nĐẠT: cảnh thử vẽ ra y như cũ.' : '\nTRƯỢT: cảnh thử ĐÃ ĐỔI hình — xem lại thay đổi đồ hoạ.');
  process.exit(ok ? 0 : 1);
}

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
const OUT = process.argv[3] ?? '/tmp/kx-scene.png';
const b = await launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'],
});
const ctx = await b.newContext({ viewport: { width: 1400, height: 820 }, locale: 'vi-VN' });
const page = await ctx.newPage();
page.on('pageerror', (e) => console.log('  pageerror', String(e).slice(0, 200)));
await page.goto(`${TARGET}/?scene=testbed&debug=true`, { waitUntil: 'domcontentloaded' });
// Chờ CỜ của cảnh, không chờ bằng `sleep`: cảnh phải tải xong atlas và font, mà thời gian
// đó phụ thuộc máy — `sleep` đủ trên máy này là `sleep` thiếu trên máy khác, và ảnh chụp
// thiếu font thì bài so sẽ báo "hình đã đổi".
await page.waitForFunction(() => window.kxTestbed?.ready === true, { timeout: 90000 });
await page.waitForSelector('canvas', { timeout: 30000 });
// Vài khung cho Pixi nạp xong texture và gói buffer lần đầu.
await sleep(2500);
const canvas = await page.locator('canvas').first();
const shot = await canvas.screenshot();
writeFileSync(OUT, shot);
const size = await page.evaluate(() => ({ w: window.kxTestbed.w, h: window.kxTestbed.h }));
console.log(`đã chụp cảnh thử ${size.w}×${size.h} → ${OUT}`);
await b.close();
