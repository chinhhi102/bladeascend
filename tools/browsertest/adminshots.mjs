// Chụp các trang admin (React + antd) trên PC 1366×768 và iPhone WebKit sau khi đăng nhập.
//   node tools/browsertest/adminshots.mjs <base-url> <outdir> <tệp-mật-khẩu>
// Mật khẩu ĐỌC TỪ TỆP, không nhận qua dòng lệnh, không in ra (CLAUDE.md luật 14).
import { readFileSync, mkdirSync } from 'node:fs';
import { chromium, webkit, devices } from 'playwright';

const [base, out, pwFile] = process.argv.slice(2);
if (!base || !out || !pwFile) {
  console.error('dùng: node adminshots.mjs <base-url> <outdir> <tệp-mật-khẩu>');
  process.exit(2);
}
// Trang có thể hiện tên người chơi thật, guest_id, IP hay nội dung góp ý → ảnh KHÔNG được
// vào repo công khai (CLAUDE.md quy tắc 15). Ghi vào `docs/private/` (đã gitignore).
const PRIVATE_PAGES = new Set(['board', 'audit', 'players', 'feedback', 'bans']);
const privateOut = process.env.PRIVATE_OUT || 'docs/private/review/11-admin';
mkdirSync(out, { recursive: true });
mkdirSync(privateOut, { recursive: true });
const pages = process.env.PAGES ? process.env.PAGES.split(',') : ['dashboard', 'rooms', 'feedback', 'map', 'players', 'errors', 'bans', 'chars', 'board', 'audit'];

async function shoot(browserType, ctxOpts, suffix) {
  const b = await browserType.launch(browserType === chromium ? { args: ['--use-gl=angle', '--use-angle=gl', '--ignore-gpu-blocklist'] } : {});
  const ctx = await b.newContext(ctxOpts);
  const pw = readFileSync(pwFile, 'utf8').trim();
  const res = await ctx.request.post(`${base}/api/admin/login`, { data: { password: pw } });
  if (!res.ok()) {
    throw new Error(`đăng nhập admin lỗi HTTP ${res.status()}`);
  }
  const p = await ctx.newPage();
  const errs = [];
  p.on('pageerror', (e) => errs.push(String(e)));
  const shots = [];
  for (const name of pages) {
    await p.goto(`${base}/admin/#${name}`, { waitUntil: 'networkidle' }).catch(() => {});
    if (name === 'map') {
      // Bộ vẽ thật của game nạp texture cả khu (vài giây trên WebKit): chờ nút "Xuất PNG" hết quay.
      await p.waitForFunction(() => !document.querySelector('.ant-btn-loading'), null, { timeout: 45000 }).catch(() => {});
    }
    await p.waitForTimeout(name === 'map' ? 3000 : 1500);
    const f = `${PRIVATE_PAGES.has(name) ? privateOut : out}/${name}_${suffix}.png`;
    await p.screenshot({ path: f });
    shots.push(f);
  }
  await b.close();
  return { shots: shots.length, errs };
}

const pc = await shoot(chromium, { viewport: { width: 1366, height: 768 } }, 'pc');
const ip = await shoot(webkit, { ...devices['iPhone 13'] }, 'iphone');
console.log(JSON.stringify({ pc, iphone: ip }));
