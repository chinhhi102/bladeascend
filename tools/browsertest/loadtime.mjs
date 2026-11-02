// Đo THỜI GIAN VÀO TRẬN với mạng bị bóp (GAME_SPEC §13.7 mục 7 / CHANGES_v6_4 mục A.7).
//
//   node loadtime.mjs [url]
//
// Ba lượt, mỗi lượt một con số khác nhau và cả ba đều đáng biết:
//   1. "Fast 4G"  (4 Mbps / 20 ms) — preset của Chrome DevTools, CHANGES_v6_4 chỉ định.
//   2. "4G 10 Mbps"               — đúng dòng ngân sách §13.7 ("4G ~10 Mbps, lần đầu ≤ 10 s").
//   3. Lần 2 (cache)              — cùng profile trình duyệt, không xoá cache.
//
// Hai mốc được đo riêng, vì người chơi cảm nhận chúng khác nhau:
//   - `trangChu`: từ lúc mở trang tới lúc THẤY được ô nhập tên (nhóm `home`).
//   - `vaoTran` : từ lúc mở trang tới lúc nhân vật đã ở trong trận (`home+boot+char+zone`).
import { chromium } from 'playwright';
import { requireDebug } from './needdebug.mjs';

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const URL = process.argv[2] ?? 'http://127.0.0.1:9997';
/** TYPE_MS: người chơi gõ tên + ngắm nhân vật bao lâu. Art nền tải trong khoảng này. */
const TYPE_MS = Number(process.env.TYPE_MS ?? 6000);

/** Bóp mạng qua CDP. `down`/`up` tính bằng byte/giây, `rtt` milligiây. */
const PROFILES = [
  { name: 'Fast 4G (4 Mbps, 20 ms)', down: (4 * 1e6) / 8, up: (3 * 1e6) / 8, rtt: 20, fresh: true },
  { name: '4G 10 Mbps (20 ms)', down: (10 * 1e6) / 8, up: (3 * 1e6) / 8, rtt: 20, fresh: true },
];

const b = await chromium.launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'],
});

const rows = [];

async function measure({ name, down, up, rtt, ctx: reuse }) {
  const ctx = reuse ?? (await b.newContext({ viewport: { width: 1280, height: 720 } }));
  const p = await ctx.newPage();
  const cdp = await ctx.newCDPSession(p);
  await cdp.send('Network.enable');
  await cdp.send('Network.emulateNetworkConditions', { offline: false, latency: rtt, downloadThroughput: down, uploadThroughput: up });
  /**
   * wire: byte đi qua dây **trên luồng chính**, KHÔNG phải tổng.
   *
   * PixiJS 8 tải và giải mã ảnh trong một web worker (`createImageBitmap`), và phiên CDP
   * của trang không thấy request của worker. Thử bám thêm phiên CDP cho từng worker cũng
   * không ăn — nên con số này chỉ gồm HTML/JS/CSS/manifest/atlas.json/font.
   *
   * Ghi lại rõ vì nó là một cái bẫy đọc số: lượt đo đầu ra "0,93 MB" cho một gói 7 MB và
   * nhìn như một thành tích ngoạn mục. Dung lượng thật lấy từ `npm run assets:size` (đọc
   * manifest, chính xác tuyệt đối); cột này chỉ để thấy lần 2 **gần như không tải gì**.
   */
  let wire = 0;
  let fromCache = 0;
  cdp.on('Network.loadingFinished', (e) => {
    wire += e.encodedDataLength ?? 0;
  });
  cdp.on('Network.requestServedFromCache', () => {
    fromCache++;
  });

  const t0 = Date.now();
  await p.goto(`${URL}/?debug=true`, { waitUntil: 'commit' });
  // Trang chủ hiện = ô nhập tên có trong DOM (màn hình tải đã ẩn).
  await p.waitForSelector('input.name-input', { timeout: 180000 });
  const home = Date.now() - t0;
  // Người chơi gõ tên và ngắm mấy nhân vật — art nền tải trong khoảng này. `TYPE_MS` là
  // thời gian đó: 6 giây là ước lượng DÈ DẶT cho việc gõ tên + xem 10 thẻ kỹ năng. Đây là
  // con số giả định duy nhất trong bài đo, nên để lộ ra chứ không giấu trong code.
  await p.fill('input.name-input', 'LoadTime');
  await sleep(TYPE_MS);
  // ——— "Thời gian chờ THỰC" của người chơi: bấm Vào chơi → đang ở trong trận ———
  const tClick = Date.now();
  await p.locator('.home .btn.primary').click();
  await p.waitForFunction(() => window.__playing?.() === true, null, { timeout: 180000 });
  await requireDebug(p, 'loadtime.mjs');
  const waited = Date.now() - tClick;
  const play = Date.now() - t0;
  await p.close();
  return { name, trangChuMs: home, choThucMs: waited, vaoTranMs: play, wireMB: +(wire / 1e6).toFixed(2), tuCache: fromCache, ctx };
}

for (const prof of PROFILES) {
  const r = await measure(prof);
  rows.push(r);
  await r.ctx.close();
}

// Lượt 3: dùng LẠI context của một lượt mới chạy → cache trình duyệt còn nguyên.
const warm = await b.newContext({ viewport: { width: 1280, height: 720 } });
const first = await measure({ name: 'Fast 4G — lần 1', down: (4 * 1e6) / 8, up: (3 * 1e6) / 8, rtt: 20, ctx: warm });
const second = await measure({ name: 'Fast 4G — lần 2 (cache)', down: (4 * 1e6) / 8, up: (3 * 1e6) / 8, rtt: 20, ctx: warm });
rows.push(first, second);
await warm.close();

console.log(`\n(gõ tên ${TYPE_MS / 1000} s trước khi bấm Vào chơi)`);
console.log('\n| Kịch bản | Trang chủ hiện | **Chờ thực** (bấm → trong trận) | Mở trang → trong trận | Byte luồng chính | Tệp từ cache |');
console.log('|---|---|---|---|---|---|');
for (const r of rows) {
  console.log(
    `| ${r.name} | ${(r.trangChuMs / 1000).toFixed(1)} s | **${(r.choThucMs / 1000).toFixed(1)} s** | ${(r.vaoTranMs / 1000).toFixed(1)} s | ${r.wireMB} MB | ${r.tuCache} |`,
  );
}
console.log(
  '\nNgân sách §13.7: vào trận lần đầu ở 4G ~10 Mbps ≤ 10 s.' +
    '\n"Byte luồng chính" KHÔNG phải tổng dung lượng tải (worker của PixiJS không đi qua' +
    '\nphiên CDP này) — dung lượng thật xem `npm run assets:size`. Cột này chỉ dùng để thấy' +
    '\nlần vào thứ 2 gần như không tải gì.',
);
await b.close();
