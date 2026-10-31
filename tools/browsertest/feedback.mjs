// Kiểm A3 (ADMIN_SPEC §7, §11) trên server thật.
//
//   node tools/browsertest/feedback.mjs [cổng] [mật khẩu]
//
// Ba thứ §11 đòi:
//   1. gửi được từ WebKit giả lập iPhone (đó là trình duyệt người chơi dùng, và là trình
//      duyệt duy nhất đã có báo lỗi thật);
//   2. `<script>` trong nội dung VÀ trong tên → hiển thị nguyên văn, không chạy;
//   3. ảnh chỉ xem được khi đã đăng nhập.
import { webkit } from 'playwright';

const PORT = Number(process.argv[2] ?? 9997);
const PASSWORD = process.argv[3] ?? 'mat-khau-thu-nghiem-123';
const BASE = `http://127.0.0.1:${PORT}`;
const XSS = '<script>window.__pwned=1</script><img src=x onerror="window.__pwned=1">';

let failures = 0;
const check = (ok, label) => {
  console.log(`  ${ok ? '✓' : '✗'} ${label}`);
  if (!ok) failures++;
};

// ── 1. Gửi góp ý từ WebKit giả lập iPhone ────────────────────────────────────────────────
const br = await webkit.launch({ headless: true });
const iphone = await br.newContext({
  viewport: { width: 390, height: 844 },
  deviceScaleFactor: 3,
  isMobile: true,
  hasTouch: true,
  userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 '
    + '(KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1',
});
const page = await iphone.newPage();
await page.goto(`${BASE}/`, { waitUntil: 'domcontentloaded' });
await page.waitForSelector('input', { timeout: 30000 });

// Nút 💬 phải có mặt NGAY ở trang chủ, trước khi vào trận.
const btn = page.locator('.fb-btn').first();
check(await btn.isVisible(), 'nút 💬 hiện ở trang chủ (WebKit/iPhone)');

await btn.click();
await page.waitForSelector('.fb-box', { timeout: 10000 });
await page.selectOption('.fb-kind', 'bug');
await page.fill('.fb-body', `Thử từ iPhone. ${XSS}`);
await page.click('.fb-send');
await page.waitForFunction(
  () => document.querySelector('.fb-msg')?.textContent?.includes('Cảm ơn'),
  null, { timeout: 15000 },
).catch(() => {});
const msg = await page.locator('.fb-msg').textContent();
check(/Cảm ơn/.test(msg ?? ''), `gửi được từ WebKit/iPhone (server trả: "${(msg ?? '').trim()}")`);

// Rate limit: 3 lá / 10 phút / guest_id (§7.1). Lá thứ 4 phải bị chặn.
let limited = false;
for (let i = 0; i < 4; i++) {
  const r = await page.evaluate(async (body) => {
    const res = await fetch('/api/feedback', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ guest_id: 'rate-test', name: 'A', kind: 'other', body }),
    });
    return res.status;
  }, 'Thử giới hạn gửi nhiều lần liên tiếp.');
  if (r === 429) limited = true;
}
check(limited, 'gửi quá 3 lá trong 10 phút thì bị chặn (429)');
await br.close();

// ── 2 + 3. Trang admin: XSS và ảnh ───────────────────────────────────────────────────────
const login = await fetch(`${BASE}/api/admin/login`, {
  method: 'POST', headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ password: PASSWORD }),
});
const cookie = /kx_admin=[^;]+/.exec(login.headers.get('set-cookie') || '')?.[0];
if (!cookie) {
  console.log('  ✗ không đăng nhập được admin');
  process.exit(1);
}
const rows = await (await fetch(`${BASE}/api/admin/feedback`, { headers: { Cookie: cookie } })).json();
const mine = rows.find((f) => f.body.includes('Thử từ iPhone'));
check(!!mine, 'lá thư từ iPhone hiện ở trang admin');
check(!!mine && mine.body.includes('<script>'),
  'nội dung lưu NGUYÊN VĂN (server không escape, chỗ hiển thị mới lo)');

// Mở trang admin thật và xem `<script>` có chạy không.
const br2 = await webkit.launch({ headless: true });
const ctx2 = await br2.newContext();
await ctx2.addCookies([{
  name: 'kx_admin', value: cookie.split('=')[1], domain: '127.0.0.1', path: '/',
}]);
const ap = await ctx2.newPage();
await ap.goto(`${BASE}/admin/#feedback`, { waitUntil: 'domcontentloaded' });
await ap.waitForSelector('.fb .body', { timeout: 15000 });
const pwned = await ap.evaluate(() => !!window.__pwned);
check(!pwned, '<script> trong góp ý KHÔNG chạy trên trang admin');
const shown = await ap.locator('.fb .body').first().textContent();
check((shown ?? '').includes('<script>'), '<script> hiển thị nguyên văn cho admin đọc');
await br2.close();

// Ảnh: chưa đăng nhập thì không xem được (§8).
const anon = await fetch(`${BASE}/api/admin/feedback/1/image`);
check(anon.status === 401 || anon.status === 404,
  `ảnh góp ý chặn khi chưa đăng nhập (HTTP ${anon.status})`);

console.log(failures === 0 ? '\nĐẠT: A3 qua hết checklist §11.' : `\nTRƯỢT: ${failures} mục.`);
process.exit(failures === 0 ? 0 : 1);
