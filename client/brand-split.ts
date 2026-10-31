/**
 * splitName chia tên game thành các cụm để NGẮT DÒNG CÓ KIỂM SOÁT — GAME_SPEC §20.
 *
 * Vì sao là một tệp riêng ở gốc `client/`: tên game hiện ở HAI runtime khác nhau.
 *
 * - `vite.config.ts` (lúc build) thay token trong `index.html` — màn hình tải là HTML tĩnh,
 *   không chạy qua TypeScript của game.
 * - `src/ui/strings.ts` (lúc chạy) dựng tên trên bảng trang chủ bằng DOM.
 *
 * Viết cùng một phép chia ở cả hai chỗ là tạo ra hai bản sự thật, và khi chúng lệch nhau
 * thì màn hình tải ngắt một kiểu, trang chủ ngắt kiểu khác — người chơi thấy tên game
 * "nhảy" lúc màn hình tải mờ đi. Một hàm, hai nơi gọi.
 *
 * Phép chia: cắt theo SỐ CHỮ, không ghi cứng hai cụm. "Đạo Hữu Xin Dừng Bước" (5 chữ) ra
 * `["Đạo Hữu", "Xin Dừng Bước"]` — đúng chỗ ngắt chủ dự án chốt 1/11. Tên dưới 4 chữ thì
 * để nguyên: ngắt một cái tên ngắn chỉ làm nó khó đọc hơn.
 */
export function splitName(name: string): string[] {
  const w = name.trim().split(/\s+/);
  if (w.length < 4) return [name];
  const cut = Math.floor(w.length / 2);
  return [w.slice(0, cut).join(' '), w.slice(cut).join(' ')];
}
