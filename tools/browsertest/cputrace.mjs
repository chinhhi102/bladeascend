// 41 ms mỗi khung đi đâu? — hồ sơ CPU của luồng chính.
//
//   node tools/browsertest/cputrace.mjs [url] [giây] [--cpu=N] [--daoto]
//
// `pingsource.mjs` đã tách được: mạng ~1 ms, server ~0,3 ms, giải mã gói 0,00 ms, Pixi vẽ
// 0,7–3,6 ms — mà một khung vẫn tốn 41 ms. Bài này trả lời "41 ms còn lại nằm ở hàm nào".
//
// Dùng `Profiler` của CDP (hồ sơ lấy mẫu), không dùng `tracing`: hồ sơ lấy mẫu cho ra
// **thời gian tự thân từng hàm**, đúng thứ cần để biết sửa chỗ nào. Tracing cho ra dòng
// thời gian sự kiện — đẹp để nhìn, nhưng không cộng được thành bảng.
import { chromium } from 'playwright';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
const SECS = Number(process.argv[3] ?? 30);
const CPU = Number((process.argv.find((a) => a.startsWith('--cpu=')) ?? '--cpu=1').slice(6));
const DAOTO = process.argv.includes('--daoto');
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

const br = await chromium.launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'],
});
const ctx = await br.newContext({ viewport: { width: 1280, height: 720 } });
const page = await ctx.newPage();
const cdp = await ctx.newCDPSession(page);
if (CPU > 1) {
  await cdp.send('Emulation.setCPUThrottlingRate', { rate: CPU });
}
await page.goto(TARGET, { waitUntil: 'domcontentloaded' });
await page.waitForSelector('input', { timeout: 60000 });
await page.fill('input', 'HoSo');
await page.locator('.home .btn.primary').click();
await page.waitForFunction(() => window.__playing?.() === true, null, { timeout: 60000 });
await sleep(4000);

if (DAOTO) {
  const ok = await page.evaluate(() => typeof window.kx === 'object' && window.kx !== null);
  if (!ok) {
    console.error('--daoto cần APP_ENV=dev trên server (ADMIN_SPEC §1.1).');
    process.exit(2);
  }
  for (let i = 0; i < 14; i++) { await page.keyboard.press('KeyR'); await sleep(100); }
  for (let i = 0; i < 110; i++) { await page.keyboard.press('KeyK'); await sleep(30); }
  await sleep(3000);
}

console.log(`Lấy hồ sơ ${SECS}s · CPU ${CPU}× · ${TARGET}`);
await cdp.send('Profiler.enable');
// 1000 µs = 1000 mẫu/giây. Đủ dày để một hàm chạy 1 ms mỗi khung vẫn hiện ra, mà không
// làm chính việc lấy mẫu méo kết quả.
await cdp.send('Profiler.setSamplingInterval', { interval: 1000 });
await cdp.send('Profiler.start');

const t0 = Date.now();
while (Date.now() - t0 < SECS * 1000) {
  await page.keyboard.down('KeyW'); await sleep(400); await page.keyboard.up('KeyW');
  await page.keyboard.down('KeyD'); await sleep(400); await page.keyboard.up('KeyD');
}
const { profile } = await cdp.send('Profiler.stop');
const stats = await page.evaluate(() => (window.__stats ? window.__stats() : null));
await br.close();

// ── Cộng thời gian TỰ THÂN theo từng hàm ────────────────────────────────────────────────
//
// `profile.samples[i]` là id nút đang chạy ở lần lấy mẫu thứ i; `timeDeltas[i]` là micro-giây
// từ mẫu trước. Cộng delta vào đúng nút đang chạy → thời gian TỰ THÂN (không gồm hàm con).
// Đó là con số nói "sửa hàm này thì tiết kiệm được bao nhiêu".
const byId = new Map(profile.nodes.map((n) => [n.id, n]));
const self = new Map();
let total = 0;
for (let i = 0; i < profile.samples.length; i++) {
  const dt = profile.timeDeltas[i] ?? 0;
  if (dt <= 0) continue;
  total += dt;
  const n = byId.get(profile.samples[i]);
  if (!n) continue;
  const f = n.callFrame;
  // Gộp theo tên hàm + tệp: cùng một hàm gọi từ nhiều chỗ vẫn là một hàm cần sửa.
  const file = (f.url || '').split('/').pop() || '(trong)';
  const name = f.functionName || '(ẩn danh)';
  const key = `${name} · ${file}:${f.lineNumber + 1}`;
  self.set(key, (self.get(key) ?? 0) + dt);
}

const rows = [...self.entries()].sort((a, b) => b[1] - a[1]);
const ms = (us) => (us / 1000).toFixed(0);
console.log(`\nTổng ${ms(total)} ms CPU luồng chính trong ${SECS}s.`);
if (stats) {
  console.log(`FPS lúc kết thúc ${stats.fps.toFixed(1)} · sprite ${stats.sprites} · draw call ${stats.drawCalls} · AOI ${stats.aoi}`);
}
console.log('\n  %CPU    ms  hàm');
for (const [key, us] of rows.slice(0, 22)) {
  const pct = (us / total) * 100;
  if (pct < 0.4) break;
  console.log(`  ${pct.toFixed(1).padStart(5)}  ${ms(us).padStart(5)}  ${key}`);
}
// `(program)` và `(idle)` là thời gian NGOÀI JavaScript — vẽ, giải nén ảnh, GC của trình
// duyệt. Tách riêng để không nhầm với mã của mình.
const outside = rows.filter(([k]) => k.startsWith('(program)') || k.startsWith('(idle)') || k.startsWith('(garbage'));
if (outside.length) {
  console.log('\n  Ngoài JavaScript:');
  for (const [k, us] of outside) {
    console.log(`  ${((us / total) * 100).toFixed(1).padStart(5)}  ${ms(us).padStart(5)}  ${k}`);
  }
}
process.exit(0);
