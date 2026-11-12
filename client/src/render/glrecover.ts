/**
 * Khôi phục khi mất ngữ cảnh WebGL.
 *
 * Trình duyệt thu hồi ngữ cảnh WebGL khi trình điều khiển card màn hình bị reset, khi máy
 * ngủ dậy, hoặc khi hệ điều hành thấy một tab ngốn quá nhiều bộ nhớ GPU. Đo thật trên máy
 * dev (AMD RX 570, Chrome headless): cứ vài lượt chạy lại dính một lần.
 *
 * Trước đây client KHÔNG bắt sự kiện này: canvas đen vĩnh viễn, người chơi tưởng game treo
 * và chẳng có cách nào ngoài tự tải lại trang.
 *
 * Vì sao TẢI LẠI TRANG chứ không dựng lại renderer: mọi texture, atlas, bộ đệm hình học và
 * shader đều nằm trên GPU và mất sạch theo ngữ cảnh. (Lỗi VẼ thường thì §13.8 có dựng lại
 * renderer — xem render/scene.ts; nhưng mất hẳn ngữ cảnh là chuyện khác, không cứu tại chỗ
 * được.) Tải lại trang thì chắc chắn đúng, và mã phiên §7.1 đưa người chơi về ĐÚNG NHÂN
 * VẬT cũ — không chỉ "vào một trận mới cùng tên" như bản `pendingRejoin` trước đây.
 */

/**
 * memLast là chốt chống lặp DỰ PHÒNG khi `sessionStorage` không ghi được.
 *
 * Nó không sống qua lần tải lại — nhưng đúng chỗ cần: nếu ngữ cảnh mất liên tiếp trong CÙNG
 * một lần tải trang (điện thoại hay bị), biến này đã đủ chặn. Trước đây `sessionStorage`
 * hỏng là mất sạch chốt, và trang tải lại vòng tròn vô tận.
 */
let memLast = 0;
/** Mốc thời gian lần mất ngữ cảnh gần nhất, sống qua lần tải lại trang. */
const LAST_KEY = 'kx.glrecover.last';
/** Hai lần mất ngữ cảnh gần nhau hơn ngần này = máy đang hỏng thật, đừng tải lại vòng lặp. */
const LOOP_GUARD_SECS = 30;

/** watchContextLoss bắt sự kiện mất ngữ cảnh trên canvas. */
export function watchContextLoss(
  canvas: HTMLCanvasElement,
  onLost: (canRecover: boolean) => void,
  onRestored?: () => void,
): void {
  let reloadTimer = 0;
  const cancelReload = (): void => {
    if (reloadTimer) {
      window.clearTimeout(reloadTimer);
      reloadTimer = 0;
    }
  };

  /**
   * Trình duyệt cấp lại ngữ cảnh — đây là chuyện BÌNH THƯỜNG trên điện thoại, không phải
   * sự cố. Huỷ lệnh tải lại đang chờ.
   */
  canvas.addEventListener('webglcontextrestored', () => {
    cancelReload();
    onRestored?.();
  });

  canvas.addEventListener('webglcontextlost', (e) => {
    // Bắt buộc: không chặn sự kiện mặc định thì trình duyệt không bao giờ cấp lại ngữ cảnh,
    // và cũng không cho ta kịp làm gì.
    e.preventDefault();

    const now = Date.now() / 1000;
    let last = memLast;
    try {
      last = Math.max(last, Number(sessionStorage.getItem(LAST_KEY) ?? 0));
      sessionStorage.setItem(LAST_KEY, String(now));
    } catch {
      /* chế độ ẩn danh chặn sessionStorage — `memLast` bên dưới vẫn giữ được chốt */
    }
    memLast = now;
    const looping = last > 0 && now - last < LOOP_GUARD_SECS;
    onLost(!looping);
    if (looping) {
      return; // vừa tải lại xong lại mất tiếp: máy hỏng thật, đừng quay vòng
    }

    // Không cần nhớ gì ở đây nữa: mã phiên §7.1 (`net/session.ts`) đã nằm trong
    // sessionStorage từ lúc vào trận, và nó mang được cả ván chứ không chỉ tên + nhân vật.

    /**
     * **KHÔNG BAO GIỜ tải lại trang khi tab đang ẩn** — đây là lỗi làm điện thoại "lúc nào
     * cũng chơi lại từ đầu".
     *
     * Chuyển sang app khác, khoá màn hình, hay chỉ một thông báo kéo xuống là đủ để trình
     * duyệt di động thu hồi ngữ cảnh WebGL. Bản trước lập tức hẹn `location.reload()` sau
     * 1,2 giây — và nó chạy TRONG LÚC người chơi đang ở app khác. Lúc quay lại thì trang đã
     * tải lại và (hồi đó) `pendingRejoin` nhảy vào một phòng MỚI, cả ván vừa chơi biến mất.
     * Nhìn từ phía người chơi: "bật game lên là nó chơi lại từ đầu".
     *
     * Đúng ra phải CHỜ tới lúc họ quay lại rồi mới quyết: ngữ cảnh đã được cấp lại thì chơi
     * tiếp (nhánh `webglcontextrestored` ở trên đã huỷ lệnh tải lại), còn vẫn mất thì mới
     * tải lại — lúc đó họ đang nhìn màn hình và hiểu chuyện gì xảy ra.
     */
    const reloadNow = (): void => {
      reloadTimer = window.setTimeout(() => location.reload(), 1200);
    };
    if (document.visibilityState === 'hidden') {
      document.addEventListener(
        'visibilitychange',
        () => {
          if (document.visibilityState === 'visible' && isContextLost(canvas)) {
            reloadNow();
          }
        },
        { once: true },
      );
      return;
    }
    reloadNow();
  });
}

/** isContextLost: ngữ cảnh WebGL còn mất không. Không lấy được ngữ cảnh thì coi như mất. */
function isContextLost(canvas: HTMLCanvasElement): boolean {
  const gl = (canvas.getContext('webgl2') ?? canvas.getContext('webgl')) as WebGLRenderingContext | null;
  return gl === null || gl.isContextLost();
}
