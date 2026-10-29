import { AdvancedBloomFilter } from 'pixi-filters';
import { Container } from 'pixi.js';

import { quality } from './quality';

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

  constructor() {
    this.apply();
  }

  /** apply gắn lại filter theo mức chất lượng hiện tại; gọi mỗi khi người chơi đổi mức. */
  apply(): void {
    const q = quality();
    if (q === this.applied) {
      return;
    }
    this.applied = q;
    if (q === 'low') {
      this.root.filters = [];
      return;
    }
    const high = q === 'high';
    this.root.filters = [
      new AdvancedBloomFilter({
        threshold: 0.45,
        bloomScale: high ? 1.1 : 0.8,
        brightness: 1,
        blur: high ? 6 : 3,
        quality: high ? 4 : 2,
      }),
    ];
  }
}
