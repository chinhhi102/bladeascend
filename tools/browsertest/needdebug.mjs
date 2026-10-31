/**
 * Bài đo nào cần `window.kx` thì gọi hàm này NGAY sau khi vào trận.
 *
 * ── Vì sao có file này ──────────────────────────────────────────────────────────────────
 *
 * Từ ADMIN_SPEC §1.1, `window.kx` chỉ tồn tại khi **server CẤP** quyền lệnh test —
 * `?debug=true` trên URL chỉ còn là một lời xin, server không nghe. Server thật ở `:9999`
 * không cấp cho ai cả.
 *
 * Hậu quả không phải là một lỗi đỏ. Bài đo chạy hết, rồi in ra một bảng số SAI KỊCH BẢN:
 *   - `netcode-bench` treo 30 giây rồi chết;
 *   - bài 10 tab in "0/5 tab lên được Đạo Tổ" rồi "BỎ LƯỢT" — đúng 13 phút để biết điều đó;
 *   - `flicker`/`wallhug` thì đọc `window.kx.world` của `undefined` và ném một lỗi chẳng
 *     liên quan gì tới thứ chúng đang đo.
 *
 * Nên: dừng NGAY, nói rõ vì sao, và nói luôn cách chạy đúng.
 */
export async function requireDebug(page, toolName) {
  const ok = await page.evaluate(() => typeof window.kx === 'object' && window.kx !== null);
  if (ok) {
    return;
  }
  console.error(`\n${toolName}: server KHÔNG cấp quyền lệnh test, nên không đo được gì.\n`);
  console.error('`window.kx` chỉ có khi server cấp quyền debug (ADMIN_SPEC §1.1);');
  console.error('`?debug=true` một mình không còn đủ từ ngày 31/10.\n');
  console.error('Cách chạy đúng — dựng một server RIÊNG ở chế độ dev:\n');
  console.error('  cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 \\');
  console.error('      -client ../client/dist -assets ../assets/dist -telemetry=false &');
  console.error(`  node tools/browsertest/${toolName} http://127.0.0.1:9998\n`);
  console.error('(APP_ENV=dev chỉ mở cửa cho kết nối từ localhost, không chạm tới :9999.)');
  process.exit(2);
}
