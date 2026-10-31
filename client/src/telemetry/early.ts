/**
 * Bộ đệm lỗi SỚM + móc hiển thị lỗi lên màn hình.
 *
 * Hai việc nhỏ nhưng là hai việc duy nhất giúp gỡ lỗi trên **iPhone**, nơi không mở được
 * console của trình duyệt:
 *
 * 1. `index.html` gắn `window.onerror` ngay ở thẻ `<script>` đầu tiên và đẩy lỗi vào
 *    `window.__kxEarlyErrors`. Module telemetry nạp xong thì vét bộ đệm đó gửi đi. Không có
 *    nó thì **mọi lỗi xảy ra trước khi bundle chạy xong đều biến mất** — mà đó đúng là loại
 *    lỗi làm game không vào nổi trận.
 * 2. `onErrorShown` để `main.ts` cắm lớp hiện lỗi ngay trên màn hình khi `?debug=true`.
 */

/** earlyErrors vét bộ đệm lỗi sớm do `index.html` ghi, rồi dọn sạch. */
export function earlyErrors(): string[] {
  const w = window as unknown as { __kxEarlyErrors?: string[] };
  const out = w.__kxEarlyErrors ?? [];
  w.__kxEarlyErrors = [];
  return out;
}

/** onErrorShown: nơi `main.ts` cắm lớp hiện lỗi trên màn hình. Mặc định không làm gì. */
export let onErrorShown: ((msg: string) => void) | null = null;

export function setErrorDisplay(fn: ((msg: string) => void) | null): void {
  onErrorShown = fn;
}
