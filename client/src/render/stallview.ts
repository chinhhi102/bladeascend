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
 * Lớp hiển thị v6.36 §6.5: tấm bạt cam + thanh đếm 6 s trên mỗi sạp, vỏ chuối vàng dọc chân
 * dãy (phía Cô Ba — đúng chỗ server rải vũng trơn), và lúc hết giờ thì sạp SỤP + vỏ chuối
 * mờ trong 0,3 s thay vì biến mất trơn.
 *
 * Nằm trong lớp `entities` của scene để y-sort đúng theo chân, như mọi vật cản khác.
 */

const COLLAPSE_SECS = 0.3;
const TARP = 0xf07b1d;
const PEEL = 0xffd84a;

/** StallCtx: những gì cần biết thêm mỗi khung — giờ, thời hạn sạp, chỗ chủ sạp đang đứng. */
export interface StallCtx {
  now: number;
  secs: number;
  /** peelR: bán kính vũng vỏ chuối theo U của chủ sạp; null nếu chủ sạp ngoài khung nhìn. */
  owner: (id: number) => { x: number; y: number; peelR: number } | null;
}

interface StallNode {
  node: Container;
  timer: Graphics;
  cutDrawn: boolean;
  born: number;
  w: number;
  x: number;
  y: number;
}

export class StallView {
  readonly layer = new Container();
  private readonly nodes = new Map<number, StallNode>();
  private readonly dying: { n: StallNode; at: number }[] = [];

  constructor() {
    this.layer.sortableChildren = true;
  }

  /** sync dựng node cho sạp mới, xoá node của sạp đã hết hạn. Gọi mỗi khung. */
  sync(w: ClientWorld, ctx: StallCtx): void {
    for (const [id, st] of w.stalls) {
      // Vẽ lại khi quyền chém ĐỔI: lên cảnh giới giữa lúc dãy sạp còn đó thì viền tối phải
      // tắt ngay, nếu không người chơi vẫn tưởng mình bị chặn và đi đường vòng vô ích.
      const canCut = w.canCutObstacle(st.tier);
      const have = this.nodes.get(id);
      if (have && have.cutDrawn === canCut) {
        continue;
      }
      have?.node.destroy({ children: true });
      const n = this.build(st.shape, canCut, have?.born ?? ctx.now);
      this.addPeel(n, ctx.owner(st.owner));
      this.layer.addChild(n.node);
      this.nodes.set(id, n);
    }
    for (const [id, n] of this.nodes) {
      if (!w.stalls.has(id)) {
        this.nodes.delete(id);
        this.dying.push({ n, at: ctx.now });
      }
    }
    for (const n of this.nodes.values()) {
      this.drawTimer(n, ctx);
    }
    this.collapse(ctx.now);
  }

  private build(sh: { x: number; y: number; w?: number; h?: number }, canCut: boolean, born: number): StallNode {
    const node = new Container();
    const w = sh.w ?? 0;
    const h = sh.h ?? 0;
    node.zIndex = sh.y + h;
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
      node.addChild(g);
    }
    // Tấm bạt cam (spec §6.5): mái răng cưa phủ nửa trên — nhìn là biết sạp TẠM của Cô Ba,
    // không phải sạp cố định của bản đồ (cùng art).
    const tarp = new Graphics();
    const teeth = 5;
    tarp.moveTo(sh.x - w * 0.04, sh.y + h * 0.05).lineTo(sh.x + w * 1.04, sh.y + h * 0.05).lineTo(sh.x + w * 1.04, sh.y + h * 0.3);
    for (let i = teeth; i > 0; i--) {
      const x = sh.x - w * 0.04 + ((i - 0.5) * w * 1.08) / teeth;
      tarp.lineTo(x, sh.y + h * 0.42).lineTo(sh.x - w * 0.04 + ((i - 1) * w * 1.08) / teeth, sh.y + h * 0.3);
    }
    tarp.closePath().fill({ color: TARP, alpha: 0.92 }).stroke({ color: 0x7a3a0a, width: Math.max(3, w * 0.03) });
    node.addChild(tarp);
    if (!canCut) {
      const o = new Graphics();
      o.rect(sh.x, sh.y, w, h).stroke({ color: 0x000000, width: 8, alpha: 0.5 });
      node.addChild(o);
    }
    const timer = new Graphics();
    node.addChild(timer);
    return { node, timer, cutDrawn: canCut, born, w, x: sh.x, y: sh.y };
  }

  /** Vỏ chuối dọc chân dãy, phía chủ sạp: tâm sạp lùi về phía Cô Ba đúng `vo_chuoi_w` U (như server). */
  private addPeel(n: StallNode, owner: ReturnType<StallCtx['owner']>): void {
    if (!owner) {
      return;
    }
    const cx = n.x + n.w / 2;
    const cy = n.y + n.w / 2;
    const d = Math.hypot(owner.x - cx, owner.y - cy) || 1;
    const px = cx + ((owner.x - cx) / d) * owner.peelR;
    const py = cy + ((owner.y - cy) / d) * owner.peelR;
    const g = new Graphics();
    const s = Math.max(10, owner.peelR * 0.3);
    for (let i = 0; i < 3; i++) {
      const a = i * 2.1;
      const x = px + Math.cos(a) * s * 1.2;
      const y = py + Math.sin(a) * s * 0.7;
      g.moveTo(x - s, y).quadraticCurveTo(x, y - s * 0.9, x + s, y).quadraticCurveTo(x, y - s * 0.35, x - s, y);
    }
    g.fill({ color: PEEL, alpha: 0.95 }).stroke({ color: 0x7a5a00, width: Math.max(2, s * 0.15) });
    g.zIndex = -1;
    n.node.addChildAt(g, 0);
  }

  /** Thanh đếm ngược nhỏ trên nóc sạp (spec §6.5). */
  private drawTimer(n: StallNode, ctx: StallCtx): void {
    const left = ctx.secs > 0 ? Math.max(0, 1 - (ctx.now - n.born) / ctx.secs) : 0;
    const bw = n.w * 0.8;
    const bh = Math.max(6, n.w * 0.07);
    const x = n.x + (n.w - bw) / 2;
    const y = n.y - bh * 2;
    n.timer.clear().roundRect(x, y, bw, bh, bh / 2).fill({ color: 0x000000, alpha: 0.55 });
    n.timer.roundRect(x, y, bw * left, bh, bh / 2).fill(TARP);
  }

  /** Sạp hết giờ: lún xuống + mờ trong 0,3 s rồi huỷ (lớp ⑤). */
  private collapse(now: number): void {
    for (let i = this.dying.length - 1; i >= 0; i--) {
      const { n, at } = this.dying[i];
      const k = (now - at) / COLLAPSE_SECS;
      if (k >= 1) {
        n.node.destroy({ children: true });
        this.dying.splice(i, 1);
        continue;
      }
      n.timer.visible = false;
      n.node.alpha = 1 - k;
      n.node.y = n.w * 0.35 * k * k;
    }
  }

  clear(): void {
    for (const n of this.nodes.values()) {
      n.node.destroy({ children: true });
    }
    for (const d of this.dying) {
      d.n.node.destroy({ children: true });
    }
    this.nodes.clear();
    this.dying.length = 0;
    this.layer.removeChildren();
  }
}
