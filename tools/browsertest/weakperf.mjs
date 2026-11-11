// Bảng đo MÁY YẾU — bước 0 của lượt tối ưu: đo trước, sửa sau.
//
//   cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true
//   node tools/browsertest/weakperf.mjs http://127.0.0.1:9998
//
// Ba câu hỏi, ba cột:
//
//  1. **Khung 41 ms đi đâu?** — `window.kx.scene.probe()` trả thời gian từng hệ thống mỗi
//     khung (nội suy, người, quái, vật phẩm, thần thú, bản đồ, ambient, vfx, kỹ năng).
//     Bật bằng F3; tắt thì mỗi mốc chỉ là một phép so boolean.
//  2. **Rác có nhiều không?** — heap dùng và BIÊN ĐỘ răng cưa trong cửa sổ đo. Phẳng =
//     không cấp phát trong vòng khung; răng cưa lớn = GC sẽ khựng.
//  3. **Nghẽn CPU hay GPU?** — hạ độ phân giải còn 0,5 rồi đo lại FPS. Tăng mạnh = GPU;
//     gần như không đổi = CPU, và khi đó hạ độ phân giải KHÔNG cứu được gì.
//
// Chạy ở ba mức bóp CPU (1× / 4× / 6×) để thấy máy yếu gãy ở đâu trước.
//
// Bài này TỪ CHỐI kết luận khi máy đang tải cao — xem ghi chú ở `gomdon-smooth.mjs`: ngày
// 10/11 cùng một commit cho ra 0,66% rồi 18,8% chỉ vì máy chạy load 39 trên 12 nhân.
import { launch } from './browser.mjs';
import { cpus, loadavg } from 'node:os';
import { requireDebug } from './needdebug.mjs';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
const SAMPLE_MS = Number(process.env.SAMPLE_MS ?? 8000);
const MAX_LOAD_PER_CPU = Number(process.env.MAX_LOAD ?? 1.5);
const CPUS = (process.env.CPU_RATES ?? '1,4,6').split(',').map(Number);
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

const perCpu = loadavg()[0] / cpus().length;
console.log(`tải máy: ${loadavg()[0].toFixed(1)} trên ${cpus().length} nhân = ${perCpu.toFixed(2)}/nhân`);
if (perCpu > MAX_LOAD_PER_CPU && process.env.IGNORE_LOAD !== '1') {
  console.error(`\nKHÔNG DÙNG ĐƯỢC: máy đang chạy ${perCpu.toFixed(2)} lần số nhân (ngưỡng ${MAX_LOAD_PER_CPU}).`);
  console.error('Bài này ĐO chính cái đang bị nhiễu. Chạy lại lúc máy rảnh, hoặc IGNORE_LOAD=1.');
  process.exit(2);
}

const b = await launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'],
});

const cfg = await (await fetch(`${TARGET}/api/config`)).json();

async function join(name, charIdx) {
  const ctx = await b.newContext({ viewport: { width: 1280, height: 720 }, locale: 'vi-VN' });
  const p = await ctx.newPage();
  p.on('pageerror', (e) => console.log(`  [${name}] pageerror`, String(e).slice(0, 140)));
  await p.goto(`${TARGET}/?debug=true`, { waitUntil: 'domcontentloaded' });
  await p.waitForSelector('input', { timeout: 30000 });
  await p.fill('input', name);
  await p.locator('.char-card').nth(charIdx).click();
  await p.locator('.home .btn.primary').click();
  await p.waitForSelector('.room-board', { timeout: 60000 });
  return p;
}

// Phòng đông + 3 Đạo Tổ trên màn hình: thả bot cạnh mình rồi kéo cấp lên.
const me = await join('KiemTraYeu', 0);
await requireDebug(me, 'weakperf.mjs');
await sleep(6000);
// KHÔNG bấm F3: `?debug=true` đã mở sẵn bảng đo, bấm nữa là TẮT nó — và tắt bảng thì
// `main` gọi `setProbe(false)` mỗi khung, đè lên lệnh bật của bài đo. Lượt đo đầu ra một
// bảng toàn số 0 vì đúng chỗ này.
for (let i = 0; i < 9; i++) {
  await me.keyboard.press('KeyR'); // lên cảnh giới: vòng kiếm dày lên, thân to ra
  await sleep(400);                // thưa tay: lên dồn dập thì client kẹt giữa lúc đổi khu
}
await sleep(3000);

// KHÔNG thả thêm bot bằng `B`: phòng dev đã sẵn hơn chục người, và bot thả cạnh mình ở
// cảnh giới cao thì chém chết người đo trong vài giây (lượt trước ra `playing=false`).
// Chết thì bấm "Chơi tiếp" rồi lên cảnh giới lại.
async function reviveIfDead() {
  if (!(await me.$('.summary'))) {
    return;
  }
  await me.locator('.summary .btn.primary').click();
  await me.waitForSelector('.room-board', { timeout: 30000 });
  await sleep(2500);
  for (let i = 0; i < 9; i++) {
    await me.keyboard.press('KeyR');
    await sleep(400);
  }
  await sleep(2000);
}
await reviveIfDead();

// Vào trận thật chưa? Bảng số 0 trông y như "game nhẹ tênh", nguy hiểm hơn là báo lỗi.
const ready = await me.evaluate(() => ({
  playing: window.kx.playing?.() ?? false,
  players: window.kx.world.players.size,
  probeOn: Object.values(window.kx.scene.probe()).some((v) => v > 0),
}));
if (!ready.playing || ready.players === 0 || !ready.probeOn) {
  console.error(`\nKHÔNG ĐO ĐƯỢC: playing=${ready.playing} người=${ready.players} probe=${ready.probeOn}.`);
  console.error('Bảng toàn số 0 là dấu hiệu client chưa vào trận hoặc bảng đo đang tắt.');
  await b.close();
  process.exit(1);
}

const cdp = await me.context().newCDPSession(me);

async function sample(label) {
  await reviveIfDead();
  // Xoá biên độ heap cũ: tắt/bật đồng hồ là reset.
  await me.evaluate(() => { window.kx.scene.setProbe(false); window.kx.scene.setProbe(true); });
  const t0 = Date.now();
  const fps = [];
  while (Date.now() - t0 < SAMPLE_MS) {
    await sleep(500);
    fps.push(await me.evaluate(() => window.kx.scene.fpsNow?.() ?? 0));
  }
  const probe = await me.evaluate(() => window.kx.scene.probe());
  if (!Object.values(probe).some((v) => v > 0)) {
    throw new Error(`đồng hồ theo hệ thống trả toàn 0 ở "${label}" — bảng đo bị tắt giữa chừng`);
  }
  const good = fps.filter((f) => f > 0).sort((a, c) => a - c);
  return {
    label,
    fps: good.length ? +good[Math.floor(good.length / 2)].toFixed(0) : 0,
    ...probe,
  };
}

const rows = [];
for (const rate of CPUS) {
  await cdp.send('Emulation.setCPUThrottlingRate', { rate });
  await sleep(2500);
  rows.push(await sample(`CPU ${rate}×`));
}
// Nghẽn CPU hay GPU: giữ CPU 4×, hạ độ phân giải còn 0,5.
await cdp.send('Emulation.setCPUThrottlingRate', { rate: 4 });
await me.evaluate(() => window.kx.scene.setResolution(0.5));
await sleep(2500);
rows.push(await sample('CPU 4× · res 0.5'));
await cdp.send('Emulation.setCPUThrottlingRate', { rate: 1 });

const KEYS = ['interp', 'players', 'monsters', 'items', 'beasts', 'map', 'ambient', 'vfx', 'skills'];
console.log('');
console.log(`| Kịch bản | FPS | ${KEYS.join(' | ')} | heap MB | răng cưa |`);
console.log(`|---|---:|${KEYS.map(() => '---:').join('|')}|---:|---:|`);
for (const r of rows) {
  const cols = KEYS.map((k) => (r[k] ?? 0).toFixed(2));
  console.log(`| ${r.label} | ${r.fps} | ${cols.join(' | ')} | ${(r.heapMB ?? 0).toFixed(0)} | ${(r.heapRangeMB ?? 0).toFixed(0)} |`);
}

const full = rows.find((r) => r.label === 'CPU 4×');
const half = rows.find((r) => r.label === 'CPU 4× · res 0.5');
if (full && half) {
  const gain = full.fps > 0 ? (half.fps / full.fps - 1) * 100 : 0;
  console.log('');
  console.log(`Hạ độ phân giải còn 0,5 ở CPU 4×: FPS ${full.fps} → ${half.fps} (${gain >= 0 ? '+' : ''}${gain.toFixed(0)}%)`);
  console.log(gain > 25
    ? '→ NGHẼN GPU: ít pixel hơn thì nhanh hơn rõ. Dynamic resolution sẽ ăn tiền.'
    : '→ NGHẼN CPU: ít pixel hơn gần như không đổi. Phải cắt việc trong JS, hạ độ phân giải không cứu.');
}
const after = loadavg()[0] / cpus().length;
console.log(`\nbản đồ: ${cfg.brand?.name ?? '?'} · mẫu ${SAMPLE_MS / 1000}s mỗi kịch bản`);
console.log(`tải: đầu ${perCpu.toFixed(2)} · cuối ${after.toFixed(2)} (mỗi nhân)`);
if (Math.max(perCpu, after) > 1.5) {
  console.log('LƯU Ý: FPS tuyệt đối trong bảng này là NHIỄU MÁY. Phép so res 1.0 ↔ 0.5 trong' +
    ' cùng một lượt vẫn đọc được, vì hai mẫu cách nhau vài giây dưới cùng điều kiện.');
}
await b.close();
