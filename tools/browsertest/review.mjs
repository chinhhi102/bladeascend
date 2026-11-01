// Chụp ảnh duyệt cho thứ người chơi nhìn thấy: PC 1366×768 (Chromium) + iPhone (WebKit).
//
//   node tools/browsertest/review.mjs <url> <thư mục ra> [tên] [--full]
//
// Ghi `<tên>_pc.png` và `<tên>_iphone.png` (thêm `_full` khi chụp cả trang). Dùng cho
// docs/review/<mục>/ — luật "thứ người chơi nhìn thấy phải có ảnh PC + iPhone".
import { chromium, webkit, devices } from 'playwright';
import { mkdirSync } from 'node:fs';
import { join } from 'node:path';

const [url, out, name = 'home', ...flags] = process.argv.slice(2);
if (!url || !out) {
  console.error('dùng: node review.mjs <url> <thư mục ra> [tên] [--full]');
  process.exit(2);
}
const full = flags.includes('--full');
mkdirSync(out, { recursive: true });

async function shoot(browserType, ctxOpts, suffix) {
  const b = await browserType.launch();
  const ctx = await b.newContext(ctxOpts);
  const page = await ctx.newPage();
  const errs = [];
  page.on('pageerror', (e) => errs.push(String(e)));
  await page.goto(url, { waitUntil: 'networkidle', timeout: 60000 }).catch(() => {});
  // WebKit nạp chunk ngôn ngữ + font Thái chậm hơn Chromium vài giây: chờ màn hình thật hiện.
  await page.waitForSelector('.screen', { timeout: 20000 }).catch(() => {});
  await page.waitForTimeout(1500);
  if (full) {
    // Ảnh `loading="lazy"` chỉ tải khi cuộn tới: cuộn hết trang trước rồi mới chụp.
    await page.evaluate(async () => {
      for (let y = 0; y < document.body.scrollHeight; y += 400) {
        window.scrollTo(0, y);
        await new Promise((r) => setTimeout(r, 60));
      }
      window.scrollTo(0, 0);
    });
    await page.waitForTimeout(800);
  }
  const file = join(out, `${name}_${suffix}${full ? '_full' : ''}.png`);
  await page.screenshot({ path: file, fullPage: full });
  await b.close();
  return { file, errs };
}

const pc = await shoot(chromium, { viewport: { width: 1366, height: 768 } }, 'pc');
const ip = await shoot(webkit, { ...devices['iPhone 13'] }, 'iphone');
console.log(JSON.stringify({ pc: pc.file, iphone: ip.file, errors: [...pc.errs, ...ip.errs] }));
