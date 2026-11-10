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

/**
 * Clock: đồng hồ server ước tính của client (Q48 bước 2).
 *
 * ── Vì sao không lấy thẳng ước tính tốt nhất ─────────────────────────────────────────────
 *
 * Bản cũ: `offset = max(mọi ước tính)`, và lùi 0,5 ms mỗi gói. Gói nào tới sớm hơn thường
 * lệ — mạng lúc nào chẳng có gói may mắn — là kéo đồng hồ NHẢY lên ngay. Đo 10/11: nhảy
 * tới **6,8 ms một nhát**. Đồng hồ nhảy lên bao nhiêu thì kho nội suy mất bấy nhiêu, và kho
 * cạn là khung đứng hình (80% số khung đứng hình trùng đúng lúc kho cạn).
 *
 * Nay: giữ cùng cái đích đó (ước tính tốt nhất trong `windowSecs` giây gần đây, để đích
 * còn biết đi XUỐNG khi gói may mắn hết hạn), nhưng **đi tới đích bằng tốc độ có trần** —
 * đồng hồ chạy nhanh/chậm `rate` (5–10%) cho tới khi bắt kịp. Không bao giờ nhảy.
 */
export interface ClockOpts {
  /** Trần chỉnh: 0.08 = đồng hồ chạy nhanh hoặc chậm 8% so với thật cho tới khi bắt kịp. */
  rate: number;
  /** Giữ ước tính tốt nhất trong ngần này giây — hết hạn thì đích tự đi xuống. */
  windowSecs: number;
}

export class Clock {
  private offset: number | null = null;
  private readonly seen: { at: number; est: number }[] = [];
  /** Lúc nhận gói đầu tiên — trong `windowSecs` đầu thì bám thẳng đích, xem `step`. */
  private startedAt = 0;
  // Khai báo rồi gán trong thân hàm, KHÔNG dùng `constructor(private opts: …)`:
  // `node --experimental-strip-types` không dựng được thuộc tính khai trong tham số.
  private readonly opts: ClockOpts;

  constructor(opts: ClockOpts) {
    this.opts = opts;
  }

  /** Gọi mỗi khi nhận gói. `est = dấu_thời_gian_server − giờ_máy`, `at` = giờ máy (giây). */
  sample(est: number, at: number): void {
    this.seen.push({ at, est });
    const cut = at - this.opts.windowSecs;
    while (this.seen.length > 1 && this.seen[0].at < cut) {
      this.seen.shift();
    }
    if (this.offset === null) {
      this.offset = est; // gói đầu tiên: phải vào ngay, không có gì để bắt kịp dần
      this.startedAt = at;
    }
  }

  /** Đích hiện tại: ước tính tốt nhất còn trong cửa sổ. */
  target(): number | null {
    if (this.seen.length === 0) {
      return this.offset;
    }
    let best = this.seen[0].est;
    for (const s of this.seen) {
      if (s.est > best) {
        best = s.est;
      }
    }
    return best;
  }

  /**
   * Gọi mỗi khung với `dt` giây thật đã trôi. Kéo offset về đích, trần `rate × dt`.
   *
   * `windowSecs` giây ĐẦU thì bám thẳng đích, không giới hạn tốc độ — lúc đó chưa vẽ ai và
   * chưa có gì để giật. Không có giai đoạn này thì đồng hồ khởi tạo từ ƯỚC TÍNH CỦA ĐÚNG
   * MỘT GÓI đầu tiên; đo 10/11 gói đó lệch 1,07 giây, và bò về đích ở 8%/giây mất 13 giây —
   * suốt 13 giây ấy thời điểm vẽ chạy trước cả gói mới nhất, kho nội suy âm, người xem thấy
   * đứng hình liên tục.
   */
  step(dt: number, at = 0): void {
    const want = this.target();
    if (this.offset === null || want === null || dt <= 0) {
      return;
    }
    if (at > 0 && at - this.startedAt < this.opts.windowSecs) {
      this.offset = want;
      return;
    }
    const d = want - this.offset;
    const limit = this.opts.rate * dt;
    this.offset += Math.max(-limit, Math.min(limit, d));
  }

  /**
   * resync: bắt sóng lại NGAY (bỏ trần tốc độ) khi đồng hồ sai quá nhiều.
   *
   * Chốt chặn, không phải đường chạy thường. Đo 10/11: 3/5 lượt đồng hồ kẹt ở chỗ lệch
   * 1,7–16 giây; ở mức đó thời điểm vẽ chạy trước cả gói mới nhất suốt hàng chục giây và
   * người xem không thấy ai nhúc nhích. Bò về đích ở 8%/giây thì mất hàng phút. Lệch tới
   * mức này là đã hỏng rồi, giật một nhát để đúng còn hơn đứng hình mãi.
   */
  resync(): void {
    const want = this.target();
    if (want !== null) {
      this.offset = want;
    }
  }

  /** Giờ server ước tính, tính từ giờ máy `at` (giây). */
  now(at: number): number {
    return at + (this.offset ?? 0);
  }

  has(): boolean {
    return this.offset !== null;
  }

  reset(): void {
    this.offset = null;
    this.seen.length = 0;
  }
}
