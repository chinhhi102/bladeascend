/**
 * Nội suy vị trí entity NGƯỜI KHÁC và QUÁI (Q48). Chính mình đi đường khác — `State.myFrame`
 * có dự đoán riêng, đừng gọi hàm ở đây cho mình.
 *
 * Tách khỏi `state.ts` để **đo được bằng bài đơn vị** (`client/scripts/interp-check.ts`):
 * bơm một chuỗi gói giả có dao động, mất gói và đồng hồ trôi, rồi canh quỹ đạo vẽ ra. Hai
 * lần trước tôi sửa thẳng trong `state.ts` rồi thử bằng trình duyệt — mỗi vòng mất ~10 phút
 * và không nhìn thấy trạng thái bên trong, nên cả hai lần đều đẩy ra bản tệ hơn bản cũ.
 *
 * Hàm này hiện là **bản nội suy thuần** (đúng hành vi đang chạy): tìm hai gói quanh thời
 * điểm vẽ rồi nội suy thẳng. Kho cạn thì giữ nguyên gói cuối — đó chính là "khung đứng
 * hình" mà Q48 bước 2–5 phải chữa. Bài đơn vị đang ĐỎ ở đúng chỗ đó, cố ý.
 */

/** Một gói đã nhận, `t` là DẤU THỜI GIAN SERVER trong gói (giây), không phải lúc gói tới. */
export interface Timed<T> {
  t: number;
  s: T;
}

export interface XY {
  x: number;
  y: number;
}

/** Track: trạng thái giữa các khung của MỘT entity. Rỗng lúc đầu là hợp lệ. */
export interface Track {
  shown?: XY;
}

export interface InterpOpts {
  /** Lệch hơn ngần này giữa HAI GÓI thì coi là dịch chuyển/hồi sinh — nhảy thẳng, không nội suy. */
  jumpUnits: number;
}

/** bracket tìm hai gói quanh thời điểm `t` và tỉ lệ pha giữa chúng. */
export function bracket<T>(buf: Timed<T>[], t: number): { a: Timed<T>; b: Timed<T>; k: number } {
  for (let i = buf.length - 1; i > 0; i--) {
    if (buf[i - 1].t <= t) {
      const a = buf[i - 1];
      const b = buf[i];
      const span = b.t - a.t;
      return { a, b, k: span > 0 ? Math.min(1, (t - a.t) / span) : 1 };
    }
  }
  return { a: buf[0], b: buf[0], k: 1 };
}

const lerp = (a: number, b: number, t: number): number => a + (b - a) * t;

/**
 * followSnaps: vị trí VẼ RA của một entity ở thời điểm vẽ `rt` (giờ server − độ trễ nội suy).
 *
 * `tr` giữ trạng thái giữa các khung; hàm có thể ghi vào nó.
 */
export function followSnaps<T extends XY>(
  buf: Timed<T>[], rt: number, tr: Track, opts: InterpOpts,
): XY {
  if (buf.length === 0) {
    return tr.shown ?? { x: 0, y: 0 };
  }
  const { a, b, k } = bracket(buf, rt);
  const jump = Math.hypot(b.s.x - a.s.x, b.s.y - a.s.y) > opts.jumpUnits;
  const t = jump ? 1 : k;
  const at = { x: lerp(a.s.x, b.s.x, t), y: lerp(a.s.y, b.s.y, t) };
  tr.shown = at;
  return at;
}
