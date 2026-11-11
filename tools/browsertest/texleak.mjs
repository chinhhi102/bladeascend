// Bài đo RÒ RỈ TEXTURE — telemetry người thật cho thấy số texture leo từ ~150 lên hơn
// 5 000 (1,4 GB) trong một phiên 25 phút, trong khi mọi bài đo cũ chỉ chạy vài chục giây
// nên không bao giờ thấy.
//
//   cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true
//   node tools/browsertest/texleak.mjs http://127.0.0.1:9998
//
// Biến môi trường: MINUTES (mặc định 6), STEP_MS (mặc định 15000).
//
// In ra mỗi mốc: số texture đang giữ, MB, và — nhờ `?textrace=1` — bảng gom theo NƠI cấp
// phát. Nhóm nào tăng đều trong khi nhóm khác đứng yên chính là chỗ quên nhả.
import { launch } from './browser.mjs';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
const MINUTES = Number(process.env.MINUTES ?? 6);
const STEP_MS = Number(process.env.STEP_MS ?? 15000);
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

const b = await launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'],
});

const ctx = await b.newContext({ viewport: { width: 1280, height: 720 }, locale: 'vi-VN' });
const p = await ctx.newPage();
p.on('pageerror', (e) => console.log('  pageerror', String(e).slice(0, 160)));
await p.goto(`${TARGET}/?debug=true&textrace=1`, { waitUntil: 'domcontentloaded' });
await p.waitForSelector('input', { timeout: 30000 });
await p.fill('input', 'KiemTraTexture');
await p.locator('.char-card').nth(0).click();
await p.locator('.home .btn.primary').click();
await p.waitForSelector('.room-board', { timeout: 60000 });
await sleep(4000);

// Chết thì vào lại — bài chạy nhiều phút nên chắc chắn có lúc chết.
async function reviveIfDead() {
  if (!(await p.$('.summary'))) {
    return;
  }
  await p.locator('.summary .btn.primary').click();
  await p.waitForSelector('.room-board', { timeout: 30000 });
  await sleep(2000);
}

// Đi lại liên tục: đứng yên thì không ai vào/ra tầm nhìn, không có đòn, không có số bay
// lên — tức là không chạm vào đúng những đường nghi ngờ nhất.
async function wander(ms) {
  const keys = ['KeyW', 'KeyA', 'KeyS', 'KeyD'];
  const t0 = Date.now();
  while (Date.now() - t0 < ms) {
    const k = keys[Math.floor(((Date.now() - t0) / 1200) % 4)];
    await p.keyboard.down(k);
    await sleep(600);
    await p.keyboard.up(k);
    await sleep(150);
  }
}

const rows = [];
let firstSites = [];
const steps = Math.max(1, Math.round((MINUTES * 60000) / STEP_MS));
console.log(`đo ${MINUTES} phút, mỗi ${STEP_MS / 1000}s một mốc\n`);
console.log('phút   texture      MB   sprite  fps');
for (let i = 0; i <= steps; i++) {
  const s = await p.evaluate(() => window.__stats());
  rows.push({ min: +((i * STEP_MS) / 60000).toFixed(2), tex: s.textures, mb: s.texMB, sp: s.sprites, fps: s.fps });
  const r = rows[rows.length - 1];
  console.log(
    `${String(r.min).padStart(5)} ${String(r.tex).padStart(9)} ${r.mb.toFixed(1).padStart(7)} ${String(r.sp).padStart(8)} ${r.fps.toFixed(0).padStart(4)}`,
  );
  if (i === 0) {
    firstSites = await p.evaluate(() => window.__texsites(80));
  }
  if (i === steps) {
    break;
  }
  await reviveIfDead();
  await wander(STEP_MS);
}

const first = rows[0];
const last = rows[rows.length - 1];
const perMin = (last.tex - first.tex) / Math.max(last.min - first.min, 0.01);
console.log(`\ntăng ${last.tex - first.tex} texture / ${(last.mb - first.mb).toFixed(0)} MB trong ${last.min} phút`);
console.log(`→ ${perMin.toFixed(0)} texture/phút · ${(((last.mb - first.mb) / last.min) || 0).toFixed(1)} MB/phút`);

const sites = await p.evaluate(() => window.__texsites(80));
console.log('\n== texture đang giữ, gom theo nơi cấp phát ==');
for (const s of sites) {
  console.log(`${String(s.count).padStart(6)} ${String(s.mb).padStart(8)} MB  ${s.site.slice(0, 200)}`);
}
console.log('\n== TĂNG so với mốc đầu (nhóm nào còn rò rỉ) ==');
const base = new Map(firstSites.map((x) => [x.site, x.count]));
const grown = sites
  .map((x) => ({ site: x.site, d: x.count - (base.get(x.site) ?? 0) }))
  .filter((x) => x.d !== 0)
  .sort((a, c) => c.d - a.d);
for (const x of grown.slice(0, 20)) {
  console.log(`${String(x.d > 0 ? '+' + x.d : x.d).padStart(6)}  ${x.site.slice(0, 200)}`);
}

await b.close();
