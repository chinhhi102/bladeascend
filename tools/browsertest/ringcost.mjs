// GIÁ CỦA VÒNG KIẾM — đo bằng PHÉP TRỪ trong cùng một lượt.
//
//   cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true
//   node tools/browsertest/ringcost.mjs http://127.0.0.1:9998
//
// ── Vì sao không dùng thẳng weakperf ────────────────────────────────────────────────────
//
// `weakperf.mjs` trả FPS và "Pixi vẽ ms" của CẢ cảnh. So hai lần chạy khác nhau là so hai
// cảnh khác nhau: ngày 12/11, cùng một bài trên cùng một máy rảnh, số sprite trên màn hình
// đi từ 2 974 tới 4 813 chỉ vì lúc đó có mấy con bot đứng gần. Biên độ ấy LỚN HƠN thứ đang
// muốn đo, nên bảng trước/sau đọc ra gì cũng được.
//
// Bài này bật rồi tắt vòng kiếm trong CÙNG một lượt, cách nhau vài giây, trên đúng đám đông
// đang đứng đó. Hiệu số là giá của riêng vòng kiếm — và CHÍNH con số đó mới đem so
// trước/sau được.
//
// Đổi qua lại nhiều vòng (BẬT, TẮT, BẬT, TẮT…) để nếu cảnh có đổi giữa chừng thì nó rơi đều
// vào cả hai bên chứ không dồn vào một bên.
import { launch } from './browser.mjs';
import { cpus, loadavg } from 'node:os';
import { requireDebug } from './needdebug.mjs';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
const CYCLES = Number(process.env.CYCLES ?? 3);
const PHASE_MS = Number(process.env.PHASE_MS ?? 6000);
const CPU_RATE = Number(process.env.CPU_RATE ?? 4);
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
const ctx = await b.newContext({ viewport: { width: 1280, height: 720 }, locale: 'vi-VN' });
const p = await ctx.newPage();
p.on('pageerror', (e) => console.log('  pageerror', String(e).slice(0, 160)));

await p.goto(`${TARGET}/?debug=true`, { waitUntil: 'domcontentloaded' });
await p.waitForSelector('input', { timeout: 30000 });
await p.fill('input', 'KiemTraVongKiem');
await p.locator('.char-card').nth(0).click();
await p.locator('.home .btn.primary').click();
await p.waitForSelector('.room-board', { timeout: 60000 });
await requireDebug(p, 'ringcost.mjs');
await sleep(5000);

// Đạo Tổ đủ kiếm VÀ bất tử: phải giữ nguyên cảnh suốt bài, chết giữa chừng là hai pha đo
// hai thứ khác nhau.
await p.evaluate(() => window.kx.conn.send({ t: 'dbg', cmd: 'max', n: 1 }));
await sleep(3000);

const cdp = await p.context().newCDPSession(p);
await cdp.send('Emulation.setCPUThrottlingRate', { rate: CPU_RATE });
await sleep(2000);

/** đo một pha: trả trung vị của `pixiMs`, FPS và số kiếm đang vẽ trên màn hình. */
async function phase(ringsOn) {
  await p.evaluate((on) => window.kx.scene.setRings(on ? 'all' : 'none'), ringsOn);
  await sleep(1500); // cho trung bình trượt kịp quên pha trước
  const pixi = [];
  const fps = [];
  const t0 = Date.now();
  while (Date.now() - t0 < PHASE_MS) {
    await sleep(400);
    const s = await p.evaluate(() => {
      const st = window.kx.scene.perfStats();
      return { pixi: st.pixiMs ?? 0, fps: st.fps ?? 0, swords: st.swordsDrawn ?? 0 };
    });
    pixi.push(s.pixi);
    fps.push(s.fps);
  }
  return { pixi: median(pixi), fps: median(fps) };
}

const on = [];
const off = [];
for (let i = 0; i < CYCLES; i++) {
  on.push(await phase(true));
  off.push(await phase(false));
  console.log(`vòng ${i + 1}: bật ${on[i].pixi.toFixed(2)} ms / ${on[i].fps} FPS · tắt ${off[i].pixi.toFixed(2)} ms / ${off[i].fps} FPS`);
}
await p.evaluate(() => window.kx.scene.setRings('all'));

const onPixi = median(on.map((r) => r.pixi));
const offPixi = median(off.map((r) => r.pixi));
const onFps = median(on.map((r) => r.fps));
const offFps = median(off.map((r) => r.fps));
const sw = await p.evaluate(() => {
  const st = window.kx.scene.perfStats();
  return { players: window.kx.world.players.size, mySwords: st.mySwords ?? 0 };
});

console.log('');
console.log(`| | Pixi vẽ ms | FPS |`);
console.log(`|---|---:|---:|`);
console.log(`| vòng kiếm BẬT | ${onPixi.toFixed(2)} | ${onFps} |`);
console.log(`| vòng kiếm TẮT | ${offPixi.toFixed(2)} | ${offFps} |`);
console.log(`| **giá vòng kiếm** | **${(onPixi - offPixi).toFixed(2)}** | ${onFps - offFps} |`);
console.log('');
console.log(`CPU ${CPU_RATE}× · ${CYCLES} vòng × ${PHASE_MS / 1000}s mỗi pha · ${sw.players} người trong tầm · mình ${sw.mySwords} kiếm`);
console.log(`tải: đầu ${perCpu.toFixed(2)} · cuối ${(loadavg()[0] / cpus().length).toFixed(2)} (mỗi nhân)`);
await b.close();
