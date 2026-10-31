// Thử game bằng giả lập iPhone — ưu tiên **WebKit** (cùng bộ máy với Safari).
//
//   node iphone.mjs [url]
//
// Kịch bản: trang chủ → vào trận → **đứng chỗ đông 2 phút**, ghi mọi lỗi và bộ nhớ GPU.
//
// Vì sao phải WebKit chứ không phải Chromium giả dạng iPhone: thứ đang hỏng trên iPhone là
// **bộ nhớ GPU và cách Safari thu hồi tab**, mà hai thứ đó thuộc về bộ máy trình duyệt, không
// thuộc về kích thước màn hình hay chuỗi User-Agent. Chromium đặt `deviceScaleFactor: 3` chỉ
// bắt chước được phần tốn bộ nhớ do DPR, không bắt chước được giới hạn của Safari.
//
// WebKit cần gói hệ thống `libavif16`. Thiếu thì script tự rơi về Chromium và **nói rõ** là
// lượt đo này không thay được Safari.
import { chromium, devices, webkit } from 'playwright';

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const URL = process.argv[2] ?? 'http://127.0.0.1:9999';
/** WATCH_MS: đứng chỗ đông bao lâu. Chủ dự án chốt 2 phút. */
const WATCH_MS = Number(process.env.WATCH_MS ?? 120000);

async function pickBrowser() {
  try {
    const b = await webkit.launch({ headless: true });
    return { b, engine: 'WebKit (như Safari)', real: true };
  } catch (err) {
    console.log(`⚠ Không chạy được WebKit: ${String(err.message).split('\n')[0]}`);
    console.log('  Cài bằng:  sudo apt-get install -y libavif16   (rồi npx playwright install webkit)');
    console.log('  Tạm dùng Chromium giả lập iPhone — KHÔNG thay được Safari cho phần bộ nhớ GPU.\n');
    const b = await chromium.launch({ headless: true, args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'] });
    return { b, engine: 'Chromium giả lập iPhone', real: false };
  }
}

const { b, engine, real } = await pickBrowser();
const ctx = await b.newContext({ ...devices['iPhone 13'] });
const p = await ctx.newPage();

const errs = [];
const warns = [];
p.on('console', (m) => {
  if (m.type() === 'error') {
    errs.push(m.text().slice(0, 200));
  } else if (m.type() === 'warning') {
    warns.push(m.text().slice(0, 120));
  }
});
p.on('pageerror', (e) => errs.push(`PAGEERROR ${e.message}`.slice(0, 200)));
p.on('crash', () => errs.push('TAB BỊ GIẾT (crash) — đúng triệu chứng iPhone'));

const mark = async (label) => {
  try {
    const s = await p.evaluate(() => (window.__stats ? window.__stats() : null));
    if (!s) {
      return { label, note: 'chưa vào trận' };
    }
    return {
      label,
      texMB: Math.round(s.texMB ?? 0),
      peakMB: Math.round(s.texPeakMB ?? 0),
      tex: s.textures ?? 0,
      fps: +Number(s.fps).toFixed(1),
      drawCalls: Math.round(s.drawCalls ?? 0),
      players: s.players ?? 0,
    };
  } catch (err) {
    return { label, err: String(err).slice(0, 80) };
  }
};

const rows = [];
const t0 = Date.now();
await p.goto(`${URL}/?debug=true`, { waitUntil: 'commit' });
await p.waitForSelector('input.name-input', { timeout: 120000 });
console.log(`${engine} · trang chủ hiện sau ${((Date.now() - t0) / 1000).toFixed(1)}s`);

await p.fill('input.name-input', 'iPhoneTest');
await p.getByText('Vào chơi', { exact: true }).click();
await p.waitForFunction(() => window.kx?.playing?.() === true, null, { timeout: 150000 });
console.log(`vào trận sau ${((Date.now() - t0) / 1000).toFixed(1)}s`);
await sleep(3000);
rows.push(await mark('vừa vào trận'));

// Ép chỗ đông: lên cảnh giới cao (khung nhìn rộng, nhiều người trong tầm) + thả bot + nhiều kiếm.
for (let i = 0; i < 13; i++) {
  await p.keyboard.press('r');
  await sleep(120);
}
for (let i = 0; i < 12; i++) {
  await p.keyboard.press('b');
  await sleep(120);
}
for (let i = 0; i < 25; i++) {
  await p.keyboard.press('k');
  await sleep(60);
}
rows.push(await mark('Đạo Tổ + 12 bot'));

// Đứng chỗ đông, đi qua lại để ép dựng/huỷ ô nền.
const until = Date.now() + WATCH_MS;
let step = 0;
while (Date.now() < until) {
  const key = ['KeyD', 'KeyS', 'KeyA', 'KeyW'][step++ % 4];
  await p.keyboard.down(key);
  await sleep(1800);
  await p.keyboard.up(key);
  await sleep(200);
  if (step % 6 === 0) {
    rows.push(await mark(`phút ${((WATCH_MS - (until - Date.now())) / 60000).toFixed(1)}`));
  }
}
rows.push(await mark('kết thúc'));

console.log(`\n| Mốc | Texture MB | Đỉnh MB | Số texture | FPS | Draw call | Người |`);
console.log('|---|---|---|---|---|---|---|');
for (const r of rows) {
  if (r.texMB === undefined) {
    console.log(`| ${r.label} | — | — | — | — | — | ${r.note ?? r.err ?? ''} |`);
    continue;
  }
  console.log(`| ${r.label} | ${r.texMB} | ${r.peakMB} | ${r.tex} | ${r.fps} | ${r.drawCalls} | ${r.players} |`);
}
console.log(`\nNgân sách GPU iOS: 256 MB (render/device.ts).`);
console.log(`Lỗi: ${errs.length ? errs.slice(0, 8).join('\n  ') : 'không'}`);
if (warns.length) {
  console.log(`Cảnh báo: ${warns.slice(0, 3).join(' · ')}`);
}
if (!real) {
  console.log('\n⚠ Lượt này chạy bằng Chromium, KHÔNG phải WebKit — số bộ nhớ đúng, nhưng');
  console.log('  cách Safari thu hồi tab thì không mô phỏng được. Cần cài libavif16.');
}
await b.close();
process.exit(errs.length ? 1 : 0);
