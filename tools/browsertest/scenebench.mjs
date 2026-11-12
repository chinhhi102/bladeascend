// ĐO HIỆU NĂNG TRÊN CẢNH THỬ CỐ ĐỊNH — bàn cân A/B của mọi lượt tối ưu đồ hoạ.
//
//   cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true
//   node tools/browsertest/scenebench.mjs http://127.0.0.1:9998
//
// ── Vì sao không dùng weakperf để so hai bản build ──────────────────────────────────────
//
// `weakperf.mjs` đo cảnh THẬT, và cảnh thật đổi theo bot: ngày 12/11, cùng một build trên
// cùng một máy rảnh, hai lượt liên tiếp cho **185 lô vẽ / 47 FPS** rồi **447 lô vẽ / 23 FPS**.
// `ringcost.mjs` cũng vậy: ba lượt trên CÙNG bản build cho giá vòng kiếm 19,79 · 0,55 ·
// −0,40 ms. Hai bài đó vẫn hữu ích để biết người chơi thật gặp gì, nhưng KHÔNG dùng để nói
// "bản mới nhanh hơn bản cũ".
//
// Cảnh thử (`?scene=testbed`) không có bot, không có mạng, không có thời gian — nên số lô vẽ
// và số nút là con số CỐ ĐỊNH, và hiệu `pixiMs` giữa hai build là hiệu của CODE.
//
// Vẫn giữ cổng tải máy của luật 16: cảnh cố định thì số LÔ VẼ không đổi theo tải, nhưng
// mili-giây thì có.
import { cpus, loadavg } from 'node:os';
import { launch } from './browser.mjs';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
const CPUS = (process.env.CPU_RATES ?? '1,4,6').split(',').map(Number);
const SAMPLES = Number(process.env.SAMPLES ?? 14);
const MAX_LOAD_PER_CPU = Number(process.env.MAX_LOAD ?? 1.5);
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const median = (a) => (a.length ? [...a].sort((x, y) => x - y)[Math.floor(a.length / 2)] : 0);

const perCpu = loadavg()[0] / cpus().length;
console.log(`tải máy: ${loadavg()[0].toFixed(1)} trên ${cpus().length} nhân = ${perCpu.toFixed(2)}/nhân`);
if (perCpu > MAX_LOAD_PER_CPU && process.env.IGNORE_LOAD !== '1') {
  console.error(`\nKHÔNG DÙNG ĐƯỢC: máy đang chạy ${perCpu.toFixed(2)} lần số nhân (ngưỡng ${MAX_LOAD_PER_CPU}).`);
  process.exit(2);
}

const b = await launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'],
});
const ctx = await b.newContext({ viewport: { width: 1400, height: 820 }, locale: 'vi-VN' });
const p = await ctx.newPage();
p.on('pageerror', (e) => console.log('  pageerror', String(e).slice(0, 160)));
await p.goto(`${TARGET}/?scene=testbed&debug=true`, { waitUntil: 'domcontentloaded' });
await p.waitForFunction(() => window.kxTestbed?.ready === true, { timeout: 90000 });
await sleep(2500);

const cdp = await p.context().newCDPSession(p);
const rows = [];
for (const rate of CPUS) {
  await cdp.send('Emulation.setCPUThrottlingRate', { rate });
  await sleep(2500);
  const pixi = [];
  const fps = [];
  let calls = 0;
  let nodes = 0;
  for (let i = 0; i < SAMPLES; i++) {
    await sleep(400);
    const s = await p.evaluate(() => window.kxTestbed.stats());
    pixi.push(s.pixiMs);
    fps.push(s.fps);
    calls = s.drawCalls;
    nodes = s.nodes;
  }
  rows.push({ rate, pixi: median(pixi), fps: median(fps), calls, nodes });
}
await cdp.send('Emulation.setCPUThrottlingRate', { rate: 1 });

console.log('');
console.log('| CPU | Pixi vẽ ms | FPS | lô vẽ | nút cây cảnh |');
console.log('|---|---:|---:|---:|---:|');
for (const r of rows) {
  console.log(`| ${r.rate}× | ${r.pixi.toFixed(2)} | ${r.fps.toFixed(0)} | ${r.calls} | ${r.nodes} |`);
}
console.log('');
console.log(`cảnh thử cố định · ${SAMPLES} mẫu mỗi mức · tải: đầu ${perCpu.toFixed(2)} · cuối ${(loadavg()[0] / cpus().length).toFixed(2)}`);
await b.close();
