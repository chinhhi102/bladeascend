import { BitmapText, Container } from 'pixi.js';

import { FONT_REALM } from './fonts';

/**
 * FloatTextLayer vẽ chữ nổi ngắn trong THẾ GIỚI: bay lên, mờ dần, rồi tự biến mất.
 *
 * Dùng cho "Cần Kim Đan" khi kiếm chạm vật cản quá bậc (spec §8.2 "Nhận biết bậc"). Chữ
 * phải nằm trong thế giới chứ không phải HUD vì nó nói về MỘT vật cản cụ thể — một dòng
 * chữ ở góc màn hình không cho biết cái tường nào.
 *
 * Dùng `BitmapText` chứ không `Text`: `Text` dựng một texture mới cho mỗi chuỗi, và mỗi
 * nhịp kiếm quét qua 5 bức tường là 5 texture 256 px sinh ra rồi bỏ. Trên iPhone, nơi ngân
 * sách GPU chỉ 256 MB (§13.8), đó đúng là loại rò rỉ đã làm tab bị kill.
 */

const RISE = 70; // bay lên bao nhiêu đơn vị trong cả vòng đời
const LIFE = 1.4; // giây

interface Item {
  node: BitmapText;
  x: number;
  y: number;
  from: number;
}

export class FloatTextLayer {
  readonly layer = new Container();
  private readonly live: Item[] = [];
  /** Pool: chữ nổi xuất hiện theo chùm rồi hết, dựng lại mỗi lần là rác cho GC. */
  private readonly pool: BitmapText[] = [];

  constructor() {
    this.layer.eventMode = 'none';
  }

  show(text: string, x: number, y: number, now: number, color = 0xffd9d9): void {
    const node = this.pool.pop() ?? new BitmapText({ text, style: { fontFamily: FONT_REALM, fontSize: 30 } });
    node.text = text;
    node.tint = color;
    node.anchor.set(0.5, 1);
    node.alpha = 1;
    node.visible = true;
    node.position.set(x, y);
    this.layer.addChild(node);
    this.live.push({ node, x, y, from: now });
  }

  update(now: number): void {
    for (let i = this.live.length - 1; i >= 0; i--) {
      const it = this.live[i];
      const k = (now - it.from) / LIFE;
      if (k >= 1) {
        it.node.visible = false;
        this.layer.removeChild(it.node);
        this.pool.push(it.node);
        this.live.splice(i, 1);
        continue;
      }
      it.node.position.set(it.x, it.y - RISE * k);
      // Hiện nhanh rồi mờ chậm: nửa đầu đọc được rõ, nửa sau nhường lại cho trận đánh.
      it.node.alpha = k < 0.2 ? k / 0.2 : 1 - (k - 0.2) / 0.8;
    }
  }

  clear(): void {
    for (const it of this.live) {
      it.node.destroy();
    }
    this.live.length = 0;
    for (const n of this.pool) {
      n.destroy();
    }
    this.pool.length = 0;
    this.layer.removeChildren();
  }
}
