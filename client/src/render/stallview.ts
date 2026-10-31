import { Container, Graphics, Sprite } from 'pixi.js';

import type { ClientWorld } from '../net/state';
import { assets } from './assets';

/**
 * StallView vẽ dãy sạp tạm của Cô Ba (spec v6.10 §6.3) và vỏ chuối dưới chân chúng.
 *
 * Tách khỏi `MapView` vì hai thứ có chu kỳ sống khác nhau: vật cản của bản đồ dựng một lần
 * lúc vào khu rồi nằm đó, còn sạp xuất hiện và biến mất cứ 6 giây một lượt. Trộn vào
 * `MapView.obstacleNodes` là thêm vòng đời thứ hai vào một chỗ đã lo đủ thứ (chunk nền,
 * nâng cấp art, y-sort).
 *
 * Nằm trong lớp `entities` của scene để y-sort đúng theo chân, như mọi vật cản khác.
 */
export class StallView {
  readonly layer = new Container();
  private readonly nodes = new Map<number, { node: Container; cutDrawn: boolean }>();

  constructor() {
    this.layer.sortableChildren = true;
  }

  /** sync dựng node cho sạp mới, xoá node của sạp đã hết hạn. Gọi mỗi khung. */
  sync(w: ClientWorld): void {
    for (const [id, st] of w.stalls) {
      // Vẽ lại khi quyền chém ĐỔI: lên cảnh giới giữa lúc dãy sạp còn đó thì viền tối phải
      // tắt ngay, nếu không người chơi vẫn tưởng mình bị chặn và đi đường vòng vô ích.
      const canCut = w.canCutObstacle(st.tier);
      const have = this.nodes.get(id);
      if (have && have.cutDrawn === canCut) {
        continue;
      }
      have?.node.destroy({ children: true });
      const node = this.build(st.shape, canCut);
      node.zIndex = st.shape.y + (st.shape.h ?? 0);
      this.layer.addChild(node);
      this.nodes.set(id, { node, cutDrawn: canCut });
    }
    for (const [id, n] of this.nodes) {
      if (!w.stalls.has(id)) {
        n.node.destroy({ children: true });
        this.nodes.delete(id);
      }
    }
  }

  private build(sh: { x: number; y: number; w?: number; h?: number }, canCut: boolean): Container {
    const node = new Container();
    const w = sh.w ?? 0;
    const h = sh.h ?? 0;
    // Sạp tạm dùng LẠI art `obstacles/stall` của sạp hàng trong Phàm Trấn: cùng là cái sạp
    // bày hàng, nên không cần thêm một bộ art thứ hai chỉ vì nó sống 6 giây.
    const tex = assets.texture('obstacles/stall/intact.png');
    if (tex) {
      const img = new Sprite(tex);
      img.width = w;
      img.height = h;
      img.position.set(sh.x, sh.y);
      node.addChild(img);
    } else {
      const g = new Graphics();
      g.rect(sh.x, sh.y, w, h).fill(0xe08f3c).stroke({ color: 0x5c3a1a, width: 5 });
      // Mái sạp: một nét chéo cho ra dáng cái sạp bày hàng, không phải một cái thùng.
      g.moveTo(sh.x, sh.y + h * 0.3).lineTo(sh.x + w, sh.y + h * 0.3).stroke({ color: 0xb5523b, width: 6 });
      node.addChild(g);
    }
    if (!canCut) {
      const o = new Graphics();
      o.rect(sh.x, sh.y, w, h).stroke({ color: 0x000000, width: 8, alpha: 0.5 });
      node.addChild(o);
    }
    return node;
  }

  clear(): void {
    for (const n of this.nodes.values()) {
      n.node.destroy({ children: true });
    }
    this.nodes.clear();
    this.layer.removeChildren();
  }
}
