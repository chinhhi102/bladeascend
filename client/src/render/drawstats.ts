/**
 * Bộ đếm LÔ VẼ và số nút cây cảnh — dùng chung cho `Scene`, cảnh thử và bài đo.
 *
 * Tách ra khỏi `scene.ts` ngày 13/11: cảnh thử (`testbed.ts`) cần đúng bộ đếm này, mà nạp
 * cả `scene.ts` chỉ để lấy một biến đếm là kéo theo toàn bộ `GameScene` vào gói tải chậm.
 * Mục 5 của lượt tối ưu (giảm lô vẽ) cũng đọc từ đây.
 */

/**
 * drawCalls đếm số lệnh vẽ WebGL mỗi khung. Pixi không công bố con số này, mà nó lại là chỉ
 * số quan trọng nhất để biết nghẽn ở GPU hay ở CPU, nên bọc thẳng hàm vẽ của ngữ cảnh WebGL.
 * Chỉ bật ở chế độ debug (spec §13.6).
 */
export const drawCalls = { perFrame: 0, acc: 0, frames: 0, on: false, ready: false };

/**
 * countDrawCalls bọc hàm vẽ — gọi một lần. Trước 1/11 chỉ gắn khi URL có `?debug=true`, còn
 * phím F3 thì mở được overlay ở MỌI trang: overlay đọc một bộ đếm chưa bao giờ được gắn và
 * ghi "draw call 0" giữa lúc đang vẽ 5 000 sprite (chủ dự án chụp được). Giờ mở overlay là
 * gắn (`perfLine`), và chưa đo xong cửa sổ đầu thì ghi "—" chứ không ghi số 0 giả.
 */
export function countDrawCalls(gl: WebGL2RenderingContext): void {
  if (drawCalls.on) {
    return;
  }
  drawCalls.on = true;
  for (const name of ['drawElements', 'drawArrays', 'drawElementsInstanced', 'drawArraysInstanced'] as const) {
    const fn = gl[name] as (...a: unknown[]) => void;
    (gl as unknown as Record<string, unknown>)[name] = (...a: unknown[]): void => {
      drawCalls.acc += 1;
      fn.apply(gl, a);
    };
  }
}

/** countNodes đếm số đối tượng hiển thị đang nằm trong cây vẽ. */
export function countNodes(c: { children?: unknown[] }): number {
  let n = 1;
  for (const k of (c.children ?? []) as { children?: unknown[] }[]) {
    n += countNodes(k);
  }
  return n;
}
