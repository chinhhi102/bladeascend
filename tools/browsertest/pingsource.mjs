// Trễ đến TỪ ĐÂU: mạng, server, hay luồng chính của client đang bận?
//
//   node tools/browsertest/pingsource.mjs [url] [giây]
//
// ── Vì sao cần ──────────────────────────────────────────────────────────────────────────
//
// Bài 10 tab đo được **ping p95 307 ms** ở mạng thường, trên LAN nội bộ. Con số đó vô lý
// nếu nó là trễ mạng thật: hai đầu cách nhau một sợi cáp. Nên nó phải đến từ chỗ khác —
// và "chỗ khác" có ba khả năng, mỗi khả năng đòi một cách sửa hoàn toàn khác:
//
//   1. **mạng**   → không sửa được bằng code, phải đổi hạ tầng;
//   2. **server** → tick chậm, hàng đợi gửi đầy;
//   3. **luồng chính client bận** → `setInterval` gửi ping bị trễ, VÀ dấu thời gian lúc
//      nhận pong cũng bị trễ, nên ping đo được = trễ thật + hai lần thời gian chờ luồng
//      chính. Đây là khả năng cao nhất với 10 tab Chrome trên một máy.
//
// Cách tách: chạy ĐỒNG THỜI hai phép đo ping trên CÙNG MỘT trang, tới CÙNG MỘT server:
//
//   - **ping luồng chính** — chính là ping của game (`window.__stats().ping`);
//   - **ping Web Worker**  — một WebSocket riêng, trong một luồng riêng, không bao giờ bị
//     vòng vẽ chặn.
//
// Hai số cùng lúc, cùng đường truyền, cùng server. Chênh lệch giữa chúng **chính là** phần
// trễ do luồng chính bận — không còn gì khác để nó là.
import { chromium } from 'playwright';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
const SECS = Number(process.argv[3] ?? 60);
/** --daoto: đưa tab lên Đạo Tổ 1050 kiếm — đúng tải vẽ của tab đã cho ra p95 307 ms. */
const DAOTO = process.argv.includes('--daoto');
/**
 * --tabs N: mở thêm N tab CÙNG CHƠI để tái hiện tranh CPU của bài 10 tab.
 *
 * Đây là biến số duy nhất khác nhau giữa "p95 307 ms" (bài 10 tab) và "p95 12 ms" (1 tab
 * Đạo Tổ, máy rảnh). Không tái hiện được nó thì không kết luận được 307 ms đến từ đâu.
 */
const EXTRA = Number((process.argv.find((a) => a.startsWith('--tabs=')) ?? '--tabs=0').slice(7));
/**
 * --cpu=N: bóp CPU N lần qua CDP, giả lập MÁY YẾU.
 *
 * Khác hẳn `--tabs`: `--tabs` làm nhiều tab tranh nhau một CPU nhanh (tình huống của bài
 * đo, không phải của người chơi), còn `--cpu` làm MỘT tab chạy trên một CPU chậm — đúng
 * tình huống của một người chơi máy cũ. Đây mới là con số nói được điều gì về họ.
 */
const CPU = Number((process.argv.find((a) => a.startsWith('--cpu=')) ?? '--cpu=1').slice(6));
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

/**
 * workerSource: mã của Web Worker, chạy trong luồng riêng.
 *
 * Nó mở WebSocket của riêng mình và **không gửi `join`** — nên server không xếp nó vào
 * phòng nào và không gửi snapshot cho nó. Mọi gói nó nhận đều là `pong` của chính nó.
 *
 * msgpack mã hoá bằng tay: gói `ping` chỉ có hai trường, và kéo cả thư viện msgpack vào
 * worker chỉ để mã hoá 15 byte là đổi một phép đo lấy một phụ thuộc.
 */
const workerSource = `
let ws, sent = 0;
const rtts = [];
function encodePing(ms) {
  // {"t":"ping","c":<f64>}  →  fixmap(2) · "t" · "ping" · "c" · float64
  const b = new Uint8Array(1 + 2 + 5 + 2 + 9);
  let p = 0;
  b[p++] = 0x82;                                  // map 2 phần tử
  b[p++] = 0xa1; b[p++] = 0x74;                   // "t"
  b[p++] = 0xa4; b[p++] = 0x70; b[p++] = 0x69; b[p++] = 0x6e; b[p++] = 0x67; // "ping"
  b[p++] = 0xa1; b[p++] = 0x63;                   // "c"
  b[p++] = 0xcb;                                  // float64
  new DataView(b.buffer).setFloat64(p, ms, false); // msgpack là big-endian
  return b;
}
self.onmessage = (e) => {
  if (e.data.cmd === 'start') {
    ws = new WebSocket(e.data.url);
    ws.binaryType = 'arraybuffer';
    ws.onopen = () => {
      setInterval(() => {
        sent = performance.now();
        try { ws.send(encodePing(sent)); } catch (_) {}
      }, 400);
    };
    ws.onmessage = () => {
      // Kết nối này không join nên chỉ nhận pong của chính nó.
      if (sent > 0) { rtts.push(performance.now() - sent); sent = 0; }
    };
  }
  if (e.data.cmd === 'read') {
    self.postMessage(rtts.slice());
  }
};
`;

const pct = (a, p) => {
  if (!a.length) return 0;
  const s = [...a].sort((x, y) => x - y);
  return s[Math.min(s.length - 1, Math.floor((s.length * p) / 100))];
};
const fmt = (a) => (a.length
  ? `p50 ${pct(a, 50).toFixed(0)} · p95 ${pct(a, 95).toFixed(0)} · tệ nhất ${Math.max(...a).toFixed(0)} ms (${a.length} mẫu)`
  : '(không có mẫu)');

// GPU thật, MỘT tab duy nhất: đây là điểm khác bài 10 tab — không có 9 tab kia tranh CPU.
const br = await chromium.launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'],
});
const ctx = await br.newContext({ viewport: { width: 1280, height: 720 } });
const page = await ctx.newPage();
if (CPU > 1) {
  const cdp = await ctx.newCDPSession(page);
  await cdp.send('Emulation.setCPUThrottlingRate', { rate: CPU });
  console.log(`  (bóp CPU ${CPU}× — giả lập máy yếu)`);
}
await page.goto(TARGET, { waitUntil: 'domcontentloaded' });
await page.waitForSelector('input', { timeout: 60000 });
await page.fill('input', 'DoTre');
await page.getByText('Vào chơi', { exact: true }).click();
await page.waitForFunction(() => window.__playing?.() === true, null, { timeout: 60000 });
await sleep(4000);

if (DAOTO) {
  // Đạo Tổ + 1050 kiếm: đây là tải VẼ đã cho ra p95 307 ms ở bài 10 tab. Không có bước
  // này thì tab chạy 58 fps và phép đo chỉ nói về một máy đang rảnh.
  const ok = await page.evaluate(() => typeof window.kx === 'object' && window.kx !== null);
  if (!ok) {
    console.error('--daoto cần quyền lệnh test: chạy server với APP_ENV=dev (ADMIN_SPEC §1.1).');
    process.exit(2);
  }
  for (let i = 0; i < 14; i++) {
    await page.keyboard.press('KeyR'); // lên một cảnh giới
    await sleep(120);
  }
  for (let i = 0; i < 110; i++) {
    await page.keyboard.press('KeyK'); // +10 kiếm
    await sleep(40);
  }
  await sleep(3000);
  const lv = await page.evaluate(() => window.kx.world.me()?.L ?? -1);
  const sw = await page.evaluate(() => window.kx.world.me()?.n ?? 0);
  console.log(`  (đã lên cảnh giới ${lv}, ${sw} kiếm)`);
}

// Mở thêm tab để tranh CPU, TRƯỚC khi đo.
const extraPages = [];
for (let i = 0; i < EXTRA; i++) {
  const p2 = await ctx.newPage();
  await p2.goto(TARGET, { waitUntil: 'domcontentloaded' });
  await p2.waitForSelector('input', { timeout: 60000 }).catch(() => {});
  await p2.fill('input', `Chen${i + 1}`).catch(() => {});
  await p2.getByText('Vào chơi', { exact: true }).click().catch(() => {});
  extraPages.push(p2);
}
if (EXTRA) {
  await sleep(5000);
  // Đưa tab chen lên Đạo Tổ luôn: tải vẽ của CHÚNG mới là thứ ăn CPU.
  for (const p2 of extraPages) {
    for (let i = 0; i < 14; i++) { await p2.keyboard.press('KeyR').catch(() => {}); }
    for (let i = 0; i < 110; i++) { await p2.keyboard.press('KeyK').catch(() => {}); }
  }
  await sleep(4000);
  console.log(`  (đã mở thêm ${EXTRA} tab Đạo Tổ để tranh CPU)`);
}

// Bật worker ping song song với ping của game.
await page.evaluate((src) => {
  const url = URL.createObjectURL(new Blob([src], { type: 'text/javascript' }));
  window.__pw = new Worker(url);
  window.__pwRtts = [];
  window.__pw.onmessage = (e) => { window.__pwRtts = e.data; };
  const proto = location.protocol === 'https:' ? 'wss' : 'ws';
  window.__pw.postMessage({ cmd: 'start', url: `${proto}://${location.host}/ws` });
}, workerSource);

console.log(`Đo ${SECS}s · ${1 + EXTRA} tab${DAOTO ? ' · Đạo Tổ 1050 kiếm' : ''} · ${TARGET}`);
const main = [];
const frames = [];
// Tách thời gian một khung ra ba phần để biết chuyển việc nào sang Worker mới đáng.
const decode = [];
const draw = [];
const frameMs = [];
const t0 = Date.now();
while (Date.now() - t0 < SECS * 1000) {
  // Đi lại để bot kéo tới, cho giống lúc đông người.
  await page.keyboard.down('KeyW'); await sleep(500); await page.keyboard.up('KeyW');
  await page.keyboard.down('KeyD'); await sleep(500); await page.keyboard.up('KeyD');
  const s = await page.evaluate(() => (window.__stats ? window.__stats() : null));
  if (s) {
    if (s.ping > 0) main.push(s.ping);
    if (s.fps > 0) frames.push(s.fps);
    if (s.decodeMs >= 0) decode.push(s.decodeMs);
    if (s.drawMs >= 0) draw.push(s.drawMs);
    if (s.fps > 0) frameMs.push(1000 / s.fps);
  }
}
const worker = await page.evaluate(() => new Promise((res) => {
  window.__pw.onmessage = (e) => res(e.data);
  window.__pw.postMessage({ cmd: 'read' });
}));
const srv = await page.evaluate(() => (window.__stats ? window.__stats().srvMs : 0));
await br.close();

console.log(`\n  ping LUỒNG CHÍNH (ping của game)  ${fmt(main)}`);
console.log(`  ping WEB WORKER (luồng riêng)     ${fmt(worker)}`);
// FPS p10 là con số quyết định: trung bình che mất những khung tệ nhất, mà chính những
// khung đó mới là lúc người chơi thấy giật.
const fp10 = frames.length ? pct(frames, 10) : 0;
const fp50 = frames.length ? pct(frames, 50) : 0;
console.log(`  FPS p10 ${fp10.toFixed(1)} · p50 ${fp50.toFixed(1)} · TB ${frames.length ? (frames.reduce((a, b) => a + b, 0) / frames.length).toFixed(1) : '?'} · server ${srv.toFixed(2)} ms/tick`);
console.log(`\n  MỘT KHUNG mổ ra: tổng ${pct(frameMs, 50).toFixed(1)} ms · giải mã gói ${pct(decode, 95).toFixed(2)} ms · Pixi vẽ ${pct(draw, 95).toFixed(2)} ms`);
console.log(`  → phần còn lại (dựng cảnh, nội suy, vật lý client, GC…): ${(pct(frameMs, 50) - pct(decode, 95) - pct(draw, 95)).toFixed(1)} ms`);

const mp95 = pct(main, 95);
const wp95 = pct(worker, 95);
const chenh = mp95 - wp95;
console.log('\n── Kết luận ──');
console.log(`  p95 luồng chính ${mp95.toFixed(0)} ms − p95 worker ${wp95.toFixed(0)} ms = ${chenh.toFixed(0)} ms`);
if (worker.length === 0) {
  console.log('  Worker không nhận được pong nào — phép đo hỏng, chưa kết luận được gì.');
  process.exit(2);
}
if (wp95 > 50) {
  console.log(`  → Trễ nằm ở MẠNG hoặc SERVER: worker (không qua luồng chính) vẫn ${wp95.toFixed(0)} ms.`);
} else if (chenh > wp95) {
  console.log(`  → Trễ nằm ở LUỒNG CHÍNH CLIENT: worker chỉ ${wp95.toFixed(0)} ms trên cùng đường truyền,`);
  console.log('    phần còn lại là thời gian ping nằm chờ vòng vẽ. Đây là lúc Web Worker giải mã gói');
  console.log('    (CHANGES_v4 mục 25) đáng làm — xem PERF_LOG.');
} else {
  console.log('  → Hai số sát nhau và đều thấp: không có nguồn trễ nào đáng kể.');
}
process.exit(0);
