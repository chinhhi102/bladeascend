import { AdvancedBloomFilter } from 'pixi-filters';
import { Container, TexturePool } from 'pixi.js';

import { filtersOff } from './guard';
import { qualityFor } from './quality';

/**
 * vfxLayer là container riêng cho MỌI thứ phát sáng: hào quang Tiên giới, Thiên Kiếp, Đài
 * Tiên Khí, hạt sáng (spec §3.6).
 *
 * Bloom gắn **một lần cho cả container này**, không gắn cho từng nhân vật và cũng không gắn
 * cho cả màn hình: gắn từng nhân vật thì mỗi người là một lượt render-to-texture riêng, còn
 * gắn cả màn hình thì mặt đất và nhân vật cũng bị loè theo.
 *
 * Mức chất lượng (§3.6): Cao = bloom đầy đủ · Vừa = bloom rẻ (ít lần mờ, nửa độ phân giải)
 * · Thấp = tắt filter, chỉ còn blend cộng.
 */
export class VfxLayer {
  readonly root = new Container();
  private applied = '';
  /** bloom: giữ lại để GẮN/THÁO theo nhu cầu, thay vì dựng mới mỗi lần. */
  private bloom: AdvancedBloomFilter | null = null;
  /** on: bloom đang gắn trên container hay không. */
  private on = false;
  /** emptyFrames: số khung liên tiếp lớp này rỗng — xem `tick`. */
  private emptyFrames = 0;

  constructor() {
    this.apply();
  }

  /**
   * tick: tháo bloom khi lớp VFX RỖNG, gắn lại khi có gì để phát sáng.
   *
   * Một container có filter là một lượt vẽ ra texture riêng MỖI KHUNG, kể cả khi nó không
   * có gì bên trong — và Pixi giữ lại những texture đó trong `TexturePool` để dùng lại.
   * Đo được (§13.10): hạng mục "Bộ đệm khung · bộ lọc" **79,5 MB**, trong đó **4 bản mỗi
   * cỡ** (4× 1920×1080, 4× 1920×1024, 4× 1024², 4× 512²) — trần là 12 MB.
   *
   * Phần lớn thời gian chơi KHÔNG có hiệu ứng phát sáng nào. Tháo filter ở những khung đó
   * là thứ rẻ nhất có thể làm.
   *
   * Chờ vài khung rồi mới tháo: hiệu ứng thường có một hai khung trống xen giữa (hạt vừa
   * tắt, hạt sau chưa sinh), và gắn/tháo filter mỗi khung thì Pixi phải dựng lại đường ống
   * — đắt hơn hẳn việc giữ nó thêm nửa giây.
   */
  tick(): void {
    const want = this.root.children.length > 0;
    if (want) {
      this.emptyFrames = 0;
    } else {
      this.emptyFrames++;
    }
    const should = want || this.emptyFrames < 30;
    if (should === this.on || !this.bloom) {
      return;
    }
    this.on = should;
    this.root.filters = should ? [this.bloom] : [];
    if (!should) {
      // Trả lại đích vẽ cho hệ thống. Không gọi thì Pixi giữ nguyên cả bốn bản mỗi cỡ cho
      // tới khi có ai đó cần đúng cỡ ấy — tức là giữ mãi.
      TexturePool.clear();
    }
  }

  /** apply gắn lại filter theo mức chất lượng hiện tại; gọi mỗi khi người chơi đổi mức. */
  apply(): void {
    // Vừa có lỗi vẽ → không gắn filter nữa cho tới hết mạng (§13.8, render/guard.ts).
    // Kiểm TRƯỚC khi so với `applied`, nếu không thì mức chất lượng không đổi là nhánh này
    // `return` sớm và bloom cũ vẫn nằm nguyên trên container.
    const q = filtersOff() ? 'low' : qualityFor('bloom');
    if (q === this.applied) {
      return;
    }
    this.applied = q;
    if (q === 'low') {
      this.root.filters = [];
      this.bloom = null;
      this.on = false;
      TexturePool.clear();
      return;
    }
    const high = q === 'high';
    this.bloom = new AdvancedBloomFilter({
      threshold: 0.45,
      bloomScale: high ? 1.1 : 0.8,
      brightness: 1,
      blur: high ? 6 : 3,
      quality: high ? 4 : 2,
    });
    // NỬA ĐỘ PHÂN GIẢI (§13.10 mục 6). Bloom là một vệt sáng nhoè — thứ duy nhất nó làm là
    // lấy trung bình vùng lân cận, nên nửa độ phân giải gần như không thấy bằng mắt mà đích
    // vẽ chỉ còn một phần tư byte.
    this.bloom.resolution = 0.5;
    this.on = true;
    this.root.filters = [this.bloom];
  }
}
