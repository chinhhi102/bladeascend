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
mkdirSync(out, { recursive: true });
const pages = ['dashboard', 'rooms', 'feedback', 'map', 'players', 'errors', 'bans', 'chars', 'board', 'audit'];

async function shoot(browserType, ctxOpts, suffix) {
  const b = await browserType.launch();
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
    await p.waitForTimeout(name === 'map' ? 4000 : 1500);
    const f = `${out}/${name}_${suffix}.png`;
    await p.screenshot({ path: f });
    shots.push(f);
  }
  await b.close();
  return { shots: shots.length, errs };
}

const pc = await shoot(chromium, { viewport: { width: 1366, height: 768 } }, 'pc');
const ip = await shoot(webkit, { ...devices['iPhone 13'] }, 'iphone');
console.log(JSON.stringify({ pc, iphone: ip }));
