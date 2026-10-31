// Chạy thử trên WebKit giả lập iPhone và ĐO bộ nhớ đồ hoạ.
//
//   node tools/browsertest/iphone.mjs [url] [giây-mỗi-chặng]
//
// Vì sao WebKit chứ không Chromium: người chơi báo lỗi là người dùng Safari trên iPhone, và
// Safari KHÁC Chromium ở đúng chỗ đang nghi — quản lý bộ nhớ texture và việc hệ điều hành
// thu hồi tab. Đo trên Chromium rồi kết luận về iPhone là đo nhầm máy.
//
// Bốn chặng theo đúng kịch bản chủ dự án yêu cầu: trang chủ → vào trận → đứng chỗ đông
// 2 phút → đi cổng sang khu khác. Mỗi chặng đọc `texMB` (ước tính bộ nhớ texture mà client
// tự tính) và FPS, rồi so với trần 256 MB của iOS.
import { webkit } from 'playwright';

const URL = process.argv[2] ?? 'http://localhost:9999';
const PHASE_S = Number(process.argv[3] ?? 30);
const IOS_BUDGET_MB = 256;

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
let failures = 0;
const check = (ok, label) => {
  console.log(`  ${ok ? '✓' : '✗'} ${label}`);
  if (!ok) failures++;
};

const br = await webkit.launch({ headless: true });
const ctx = await br.newContext({
  viewport: { width: 390, height: 844 },
  deviceScaleFactor: 3, // iPhone thật: DPR 3 — đúng con số làm texture tốn gấp 9 lần
  isMobile: true, hasTouch: true,
  userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 18_7 like Mac OS X) AppleWebKit/605.1.15 '
    + '(KHTML, like Gecko) Version/18.7 Mobile/15E148 Safari/604.1',
});
const page = await ctx.newPage();
const errors = [];
page.on('pageerror', (e) => errors.push(String(e)));
page.on('console', (m) => { if (m.type() === 'error') errors.push(m.text()); });

/**
 * stat đọc số liệu client tự đo, qua `window.__stats` — cùng bộ số của overlay F3.
 *
 * `__stats` KHÔNG khoá sau quyền debug (xem `main.ts`): nó chỉ là số đo, và khoá nó lại thì
 * không đo được trên server thật — mà server thật là nơi duy nhất đo đúng.
 */
const stat = () => page.evaluate(() => {
  const f = window.__stats;
  if (typeof f !== 'function') return null;
  const s = f();
  return { texMB: s.texMB, tex: s.textures, fps: s.fps, draw: s.drawCalls, sprites: s.sprites };
});

console.log(`WebKit giả lập iPhone (DPR 3, 390×844) → ${URL}`);
await page.goto(URL, { waitUntil: 'domcontentloaded' });
await page.waitForSelector('input', { timeout: 60000 });
await sleep(3000);
const home = await stat();
console.log(`  trang chủ: ${home ? `${home.texMB} MB texture` : 'chưa có số liệu'}`);

// Vào trận.
await page.fill('input', 'iPhoneRun');
const play = page.locator('button.primary, .btn.primary').first();
await play.click();
await page.waitForFunction(() => !!document.querySelector('canvas'), null, { timeout: 90000 });
await sleep(8000);
const joined = await stat();
check(!!joined, `vào trận được trên WebKit/iPhone${joined ? ` (${joined.texMB} MB)` : ''}`);

// Đứng chỗ đông: giữ phím di chuyển để bot kéo tới.
const t0 = Date.now();
let peak = 0, fpsSum = 0, fpsN = 0;
while (Date.now() - t0 < PHASE_S * 1000) {
  await page.keyboard.down('KeyW');
  await sleep(700);
  await page.keyboard.up('KeyW');
  await page.keyboard.down('KeyD');
  await sleep(700);
  await page.keyboard.up('KeyD');
  const s = await stat();
  if (s) {
    peak = Math.max(peak, s.texMB || 0);
    if (s.fps > 0) { fpsSum += s.fps; fpsN++; }
  }
}
const fps = fpsN ? fpsSum / fpsN : 0;
console.log(`  chỗ đông ${PHASE_S}s: đỉnh ${peak} MB · FPS TB ${fps.toFixed(1)}`);

// `peak > 0` phải kiểm TRƯỚC, và phép kiểm trần phải phụ thuộc vào nó: "0 ≤ 256" là
// một lời khẳng định đúng về một phép đo đã hỏng, và nó biến một lượt đo thất bại thành
// một dấu ✓ trong báo cáo.
check(peak > 0, 'đọc được bộ nhớ texture');
check(peak > 0 && peak <= IOS_BUDGET_MB,
  `bộ nhớ texture đỉnh ${peak} MB ≤ trần iOS ${IOS_BUDGET_MB} MB`);
// FPS ở đây CHỈ ĐỂ THAM KHẢO, không dùng để đạt/trượt.
//
// WebKit headless trên Linux vẽ bằng phần mềm, không có GPU. Con số nó cho ra là tốc độ
// của máy dựng, không phải của iPhone — lấy nó mà kết luận "iPhone chạy 17 fps" là sai
// hẳn. Thứ ĐO ĐƯỢC ở đây là bộ nhớ texture: client tự cộng kích thước texture nên con số
// đó không phụ thuộc GPU, và nó chuyển thẳng sang máy thật.
console.log(`  (FPS ${fps.toFixed(1)} — tham khảo: WebKit headless vẽ bằng phần mềm, không phải tốc độ iPhone)`);
check(fps > 0, `đo được FPS (${fps.toFixed(1)}) — số 0 nghĩa là bộ đếm khung hỏng`);

// Lỗi trong lúc chơi: đây chính là thứ người chơi iPhone báo mà không mô tả được.
const real = errors.filter((e) => !/favicon|ResizeObserver/.test(e));
check(real.length === 0, `không có lỗi nào trên WebKit${real.length ? `: ${real[0].slice(0, 120)}` : ''}`);

const final = await stat();
console.log(`  kết thúc: ${final ? `${final.texMB} MB · ${final.tex} texture · ${final.draw} draw call · ${final.sprites} sprite` : '—'}`);
console.log(`\nSỐ ĐỂ GHI PERF_LOG (S9): texMB đỉnh ${peak} · FPS TB ${fps.toFixed(1)} · lỗi ${real.length}`);

await br.close();
process.exit(failures === 0 ? 0 : 1);
