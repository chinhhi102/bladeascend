// Mở trình duyệt cho bài đo — và BẢO ĐẢM nó được đóng, kể cả khi bài đo chết giữa chừng.
//
// ── Vì sao có tệp này ────────────────────────────────────────────────────────────────────
//
// 18/22 bài đo gọi `chromium.launch()` rồi `await b.close()` ở CUỐI. Đường thành công thì
// đúng; ném lỗi giữa chừng thì trình duyệt ở lại chạy mãi. Đã xảy ra thật trong phiên
// 10/11: `weakperf.mjs` ném `TypeError` vì gọi hàm chưa có trong bundle, và Chromium của nó
// sống tiếp. Mấy con như thế cộng dồn là một phần của chuyện máy chạy load 31–46 rồi mọi
// phép đo sau đó thành vô nghĩa — bài đo tự phá chính thứ nó đang đo.
//
// Vá từng tệp bằng `try/finally` thì phải nhớ vá bài mới viết sau này. Gói vào một chỗ thì
// không có chỗ nào để quên.
//
// Bắt đủ bốn đường thoát:
//   - chạy xong bình thường   → `exit`
//   - ném lỗi không ai bắt    → `uncaughtException`, `unhandledRejection`
//   - Ctrl-C / bị kill        → `SIGINT`, `SIGTERM`
//   - `process.exit()` giữa chừng → `exit` (chỉ kịp giết tiến trình con, không kịp `await`)
import { chromium, webkit } from 'playwright';

const open = new Set();
let hooked = false;

function hook() {
  if (hooked) {
    return;
  }
  hooked = true;
  // `exit` chạy ĐỒNG BỘ: không `await` được, nên giết thẳng tiến trình con.
  process.on('exit', () => {
    for (const b of open) {
      try {
        b.process()?.kill('SIGKILL');
      } catch {
        // đang tắt máy rồi, không còn gì để làm
      }
    }
  });
  for (const sig of ['SIGINT', 'SIGTERM']) {
    process.on(sig, () => {
      void closeAll().finally(() => process.exit(130));
    });
  }
  for (const ev of ['uncaughtException', 'unhandledRejection']) {
    process.on(ev, (err) => {
      console.error(`\n[browsertest] ${ev}:`, err?.stack ?? err);
      void closeAll().finally(() => process.exit(1));
    });
  }
}

async function closeAll() {
  for (const b of [...open]) {
    open.delete(b);
    await b.close().catch(() => b.process()?.kill('SIGKILL'));
  }
}

/** launch mở Chromium và ghi sổ để đóng chắc chắn. Dùng thay `chromium.launch`. */
export async function launch(opts = {}) {
  hook();
  const b = await chromium.launch(opts);
  open.add(b);
  b.on('disconnected', () => open.delete(b));
  return b;
}

/** launchWebkit: như trên, cho bài đo giả lập iPhone. */
export async function launchWebkit(opts = {}) {
  hook();
  const b = await webkit.launch(opts);
  open.add(b);
  b.on('disconnected', () => open.delete(b));
  return b;
}

/**
 * launchAny: khi bài đo nhận `browserType` (chromium hay webkit) làm tham số.
 *
 * `review.mjs` và `adminshots.mjs` chụp cả PC lẫn iPhone trong một lần chạy nên chúng
 * truyền loại trình duyệt vào hàm; vẫn phải vào chung sổ như hai hàm trên.
 */
export async function launchAny(browserType, opts = {}) {
  hook();
  const b = await browserType.launch(opts);
  open.add(b);
  b.on('disconnected', () => open.delete(b));
  return b;
}
