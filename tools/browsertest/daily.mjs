// MỘT BUỔI CHƠI BÌNH THƯỜNG — 0 lỗi vẽ, 0 hình tạm, không rơi khỏi trận.
//
//   cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true
//   node tools/browsertest/daily.mjs http://127.0.0.1:9998
//
// ── Vì sao có bài này ───────────────────────────────────────────────────────────────────
//
// Ngày 13/11 một lỗi vẽ làm ba người chơi đứng hình rồi văng khỏi trận, và MỌI cổng kiểm
// đều xanh. Lý do: không bài nào ĐỔI KHU — tất cả đứng yên một chỗ trong một khu. Rà lại
// thì còn một loạt việc người chơi làm hằng ngày mà không bài nào đi qua: chết rồi bấm
// "Chơi tiếp", đổi mức chất lượng, đổi cỡ cửa sổ, chuyển tab, mạng chập chờn.
//
// Bài này đi hết các việc đó trong MỘT phiên, và sau MỖI bước hỏi đúng ba câu:
//
//   1. Có lỗi vẽ nào không (lỗi THẬT — lỗi giả của `crashtest` mang tên khác)?
//   2. Có ai đang vẽ bằng hình tạm không?
//   3. Còn ở trong trận không, hay đã bị ném về trang chủ?
//
// Nó KHÔNG thay `crashtest` (thử lớp bắt lỗi bằng lỗi giả) hay `drawguard` (đổi đủ 5 khu).
// Ba bài hỏi ba câu khác nhau.
import { launch } from './browser.mjs';
import { requireDebug } from './needdebug.mjs';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

let fails = 0;
function fail(step, why) {
  console.log(`TRƯỢT  ${step}: ${why}`);
  fails++;
}

const b = await launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'],
});
const ctx = await b.newContext({ viewport: { width: 1280, height: 720 }, locale: 'vi-VN' });
const p = await ctx.newPage();
p.on('pageerror', (e) => console.log('  pageerror', String(e).slice(0, 180)));

/** hook: cắm bộ đếm lỗi vẽ vào Scene. Phải cắm LẠI sau mỗi lần tải lại trang. */
async function hook() {
  await p.evaluate(() => {
    window.__kxErr = [];
    const prev = window.kx.scene.onGuardEvent;
    window.kx.scene.onGuardEvent = (kind, err) => {
      // `*_gia` là lỗi do bài kiểm bắn ra (xem `guard.ts: markFake`) — không tính.
      if (!String(kind).endsWith('_gia')) {
        window.__kxErr.push(`${kind}: ${String(err?.message ?? err ?? '').slice(0, 160)}`);
      }
      prev?.(kind, err);
    };
  });
}

/** check: ba câu hỏi sau mỗi bước. */
async function check(step) {
  await sleep(1200);
  const got = await p.evaluate(() => ({
    err: window.__kxErr ?? [],
    ph: window.kx?.scene?.drawStats?.().placeholder ?? null,
    chars: window.kx?.scene?.drawStats?.().charsInView ?? 0,
    inMatch: !!document.querySelector('.room-board'),
    home: !!document.querySelector('.home'),
  }));
  if (got.err.length) {
    fail(step, `${got.err.length} lỗi vẽ — ${got.err[0]}`);
    await p.evaluate(() => { window.__kxErr = []; });
  } else if (got.ph) {
    fail(step, `có nhân vật đang vẽ bằng hình tạm (${got.chars} nhân vật trên màn)`);
  } else if (!got.inMatch) {
    fail(step, got.home ? 'bị ném về TRANG CHỦ' : 'không còn ở trong trận');
  } else {
    console.log(`đạt   ${step}`);
  }
}

async function enterMatch(name) {
  await p.waitForSelector('input', { timeout: 30000 });
  await p.fill('input', name);
  await p.locator('.char-card').nth(0).click();
  await p.locator('.home .btn.primary').click();
  await p.waitForSelector('.room-board', { timeout: 60000 });
}

// ── 1. Vào trận ─────────────────────────────────────────────────────────────────────────
await p.goto(`${TARGET}/?debug=true`, { waitUntil: 'domcontentloaded' });
await enterMatch('KiemTraNgayThuong');
await requireDebug(p, 'daily.mjs');
await hook();
await sleep(4000);

// ── 0. ĐỐI CHỨNG ÂM: bài đo phải BẮT ĐƯỢC một lỗi vẽ thật ───────────────────────────────
//
// Một bài kiểm chưa bao giờ đỏ thì chưa chứng minh được gì — nó có thể đang canh một cái
// móc đã rụng từ lâu. Nên mỗi lượt chạy, trước khi đo gì cả, ném MỘT lỗi vẽ thật vào vòng
// vẽ và đòi thấy nó. Lỗi này KHÔNG mang dấu `markFake`, nên nó đi đúng đường mà một lỗi
// thật sẽ đi.
await p.evaluate(() => {
  const r = window.kx.scene.app.renderer;
  const orig = r.render.bind(r);
  let once = true;
  r.render = (...a) => {
    if (once) {
      once = false;
      throw new Error('đối chứng âm của daily.mjs');
    }
    return orig(...a);
  };
});
await sleep(1500);
const caught = await p.evaluate(() => (window.__kxErr ?? []).length);
if (caught === 0) {
  fail('đối chứng âm', 'ném một lỗi vẽ thật mà bài đo KHÔNG thấy — mọi dòng "đạt" bên dưới vô nghĩa');
} else {
  console.log(`đạt   đối chứng âm (bắt được ${caught} lỗi vẽ thật)`);
}
await p.evaluate(() => { window.__kxErr = []; });
await sleep(1500);

await check('vào trận');

// ── 2. Lên Đạo Tổ rồi dùng kỹ năng nhiều lượt ───────────────────────────────────────────
//
// Kỹ năng là nơi VFX nặng nhất chạy. `dbg skill` hồi đầy lượt nên không phải chờ hồi chiêu.
await p.evaluate(() => window.kx.conn.send({ t: 'dbg', cmd: 'max', n: 1 }));
await sleep(2500);
for (let i = 0; i < 10; i++) {
  await p.evaluate(() => window.kx.conn.send({ t: 'dbg', cmd: 'skill' }));
  await p.keyboard.press('Space');
  await sleep(900);
}
await check('dùng kỹ năng 10 lượt');

// ── 3. Đổi mức chất lượng ───────────────────────────────────────────────────────────────
for (const q of ['low', 'high', 'medium']) {
  await p.evaluate((qq) => window.kx.quality(qq), q);
  await sleep(1500);
  await check(`đổi chất lượng → ${q}`);
}

// ── 4. Đổi cỡ cửa sổ (xoay máy, kéo cửa sổ, thanh địa chỉ co lại) ───────────────────────
for (const [w, h] of [[900, 600], [1600, 900], [430, 860], [1280, 720]]) {
  await p.setViewportSize({ width: w, height: h });
  await sleep(1200);
  await check(`đổi cỡ cửa sổ ${w}×${h}`);
}

// ── 5. Chuyển tab đi rồi quay lại ───────────────────────────────────────────────────────
//
// Tab ẩn thì trình duyệt dừng `requestAnimationFrame`; lúc hiện lại, khung đầu tiên có `dt`
// rất lớn. Đó là chỗ các bộ đếm theo thời gian dễ nhảy số âm hoặc chia cho 0.
await p.evaluate(() => {
  Object.defineProperty(document, 'visibilityState', { value: 'hidden', configurable: true });
  document.dispatchEvent(new Event('visibilitychange'));
});
await sleep(3000);
await p.evaluate(() => {
  Object.defineProperty(document, 'visibilityState', { value: 'visible', configurable: true });
  document.dispatchEvent(new Event('visibilitychange'));
});
await check('chuyển tab đi rồi quay lại');

// ── 6. Mạng chập chờn ───────────────────────────────────────────────────────────────────
//
// Rớt mạng vài giây rồi có lại: client phải nối lại và vào đúng nhân vật cũ (§7.1), không
// phải vẽ lại cả thế giới bằng hình tạm.
await ctx.setOffline(true);
await sleep(4000);
await ctx.setOffline(false);
await sleep(6000);
await check('mạng rớt 4 giây rồi có lại');

// ── 7. Qua cổng sang khu khác ───────────────────────────────────────────────────────────
await p.evaluate(() => window.kx.conn.send({ t: 'dbg', cmd: 'zone', n: 4 }));
await check('qua khu Thiên Đình');

// ── 8. Chết rồi bấm "Chơi tiếp" ─────────────────────────────────────────────────────────
await p.evaluate(() => window.kx.conn.send({ t: 'dbg', cmd: 'die' }));
await p.waitForSelector('.summary', { timeout: 30000 }).catch(() => {});
if (!(await p.evaluate(() => !!document.querySelector('.summary')))) {
  fail('chết → màn Tổng kết', 'không thấy màn Tổng kết');
} else {
  console.log('đạt   chết → màn Tổng kết');
  await p.locator('.summary .btn.primary').first().click();
  await p.waitForSelector('.room-board', { timeout: 60000 });
  await hook(); // Scene có thể đã dựng lại
  await check('bấm "Chơi tiếp" → vào ván mới');
}

// ── 9. F5 giữa trận ─────────────────────────────────────────────────────────────────────
await p.reload({ waitUntil: 'domcontentloaded' });
await p.waitForSelector('.room-board, .home', { timeout: 60000 });
if (await p.evaluate(() => !!document.querySelector('.home'))) {
  await enterMatch('KiemTraNgayThuong');
}
await requireDebug(p, 'daily.mjs');
await hook();
await check('F5 giữa trận');

await b.close();
console.log('');
if (fails > 0) {
  console.error(`✗ buổi chơi bình thường: ${fails} bước TRƯỢT.`);
  process.exit(1);
}
console.log('✓ buổi chơi bình thường: mọi bước sạch (0 lỗi vẽ, 0 hình tạm, không rơi khỏi trận).');
