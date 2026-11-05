import { Container, Graphics, Sprite } from 'pixi.js';

import { assets } from './assets';
import { quality } from './quality';

/**
 * Bong bóng mặt — GAME_SPEC §12.5 (v6.38, quyết định Q44).
 *
 * Hướng A "một thân nhiều mặt" đã bỏ: art thử cho thấy AI luôn tự vẽ mặt vào thân, thân và
 * mặt khác nét khác góc, không có neo đầu. Thay vào đó dùng lại **4 chân dung biểu cảm sẵn
 * có** của §12.4 làm emote nổi trên đầu khi tâm trạng đổi. Thân vẫn chạy bằng bộ frame cũ,
 * không đổi dáng người, và KHÔNG cần art mới.
 *
 * Khung ảnh cắt y như ô đại diện trên HUD: `background: center top / 180%` trong
 * `style.css`. Cùng một phép cắt ở hai chỗ thì cùng một khuôn mặt — chân dung là ảnh toàn
 * thân ở `neutral`/`cry` và bán thân ở `evil`/`panic`, nên không cắt thì bong bóng ra cái
 * người tí xíu thay vì cái mặt.
 */

/** Hai biểu cảm dùng cho emote. `neutral`/`cry` để cho HUD và màn tổng kết. */
export type EmoteFace = 'evil' | 'panic';

const LIFE = 1.5;
/** Cooldown mỗi người (spec §12.5). Không có nó thì lúc hỗn chiến mặt nhấp nháy liên tục. */
const COOLDOWN = 3;
const POP = 0.18;
const FADE = 0.25;
/** Phóng ảnh trong khung tròn — khớp `background: ... / 180%` của `.badge .portrait`. */
const ZOOM = 1.8;
/** Bán kính bong bóng, tính bằng PIXEL MÀN HÌNH (xem `update`). */
const R = 28;

/**
 * Trần số người trên màn hình để còn vẽ emote ở chất lượng Thấp (spec §12.5).
 *
 * Máy chạy chất lượng Thấp là điện thoại và GPU tích hợp; mỗi bong bóng là một texture
 * chân dung 512×512 riêng, nên đông người là đông texture. Trên 15 người thì tắt hẳn —
 * emote là thứ trang trí, không phải thứ để chơi được.
 */
const LOW_MAX_CHARS = 15;

/** Số người đang vẽ, do scene cập nhật mỗi khung. Dùng cho luật trần ở trên. */
let charsOnScreen = 0;
export function setEmoteCharCount(n: number): void {
  charsOnScreen = n;
}

/**
 * Báo ra console khi bong bóng hiện — CHỈ ở `?debug=true`.
 *
 * Bong bóng sống 1,5 giây và chỉ bật khi tâm trạng đổi, nên chụp ảnh duyệt bằng cách hẹn
 * giờ là trò may rủi: lượt chụp đầu ngày 05/11 lấy 24 ảnh mà dính 0 cái. Có dòng này thì
 * `tools/browsertest/emoteshots.mjs` nghe console rồi bấm máy ĐÚNG lúc.
 *
 * Đọc `location.search` một lần lúc nạp mô-đun, không đọc mỗi lần hiện.
 */
const DEBUG = typeof location !== 'undefined' && location.search.includes('debug=true');

function allowed(): boolean {
  return quality() !== 'low' || charsOnScreen <= LOW_MAX_CHARS;
}

export class Emote {
  readonly root = new Container();
  private readonly bubble = new Graphics();
  private readonly face = new Sprite();
  private readonly mask = new Graphics();
  private shownAt = -999;
  private lastAt = -999;
  private radius = 0;

  constructor() {
    this.root.visible = false;
    this.root.addChild(this.bubble, this.face, this.mask);
    this.face.mask = this.mask;
  }

  /**
   * show bật bong bóng. Trả về false nếu bị cooldown hoặc chất lượng chặn — người gọi không
   * cần biết luật, chỉ cần gọi mỗi lần tâm trạng đổi.
   */
  show(char: string, face: EmoteFace, now: number): boolean {
    if (now - this.lastAt < COOLDOWN || !allowed()) {
      return false;
    }
    const tex = assets.texture(`characters/${char}/portrait_${face}.png`);
    if (!tex) {
      return false; // chưa tải xong chân dung thì thôi, không vẽ ô trống
    }
    this.face.texture = tex;
    this.fitFace(R);
    this.shownAt = now;
    this.lastAt = now;
    this.root.visible = true;
    if (DEBUG) {
      console.log(`[EMOTE] ${char} ${face}`);
    }
    return true;
  }

  /**
   * update đặt vị trí và vẽ lại. `topY` là mép TRÊN của cụm bảng tên (toạ độ cục bộ của
   * `PlayerView`). `canShow` là luật ẩn của người chơi (Ẩn Thân, trong bụi) —
   * truyền từ `PlayerView` để emote dùng CHUNG một luật với bảng tên, không phải luật thứ hai
   * tự nghĩ ra rồi lệch nhau.
   */
  update(topY: number, zoom: number, now: number, canShow: boolean): void {
    if (!this.root.visible) {
      return;
    }
    const age = now - this.shownAt;
    if (age > LIFE || !canShow) {
      this.root.visible = false;
      return;
    }
    // Vẽ ở cỡ PIXEL CỐ ĐỊNH rồi `scale = 1/zoom` để huỷ phóng của camera — đúng quy ước
    // `Nameplate.update`. Bản đầu nhân bán kính với `zoom` (ngược chiều) nên bong bóng teo
    // mất ở mức phóng thường gặp: ảnh chụp 05/11 không thấy bong bóng nào dù log báo đã bật.
    if (this.radius !== R) {
      this.radius = R;
      this.drawBubble(R);
    }
    // Nảy nhẹ khi hiện: phình quá 1 rồi về 1 (spec §12.5).
    let pop = 1;
    if (age < POP) {
      const t = age / POP;
      pop = 0.6 + 0.5 * Math.sin(t * Math.PI * 0.5) + 0.1 * Math.sin(t * Math.PI);
    }
    this.root.scale.set(pop / zoom);
    this.root.alpha = age > LIFE - FADE ? Math.max(0, (LIFE - age) / FADE) : 1;
    // Đặt TRÊN HẲN bảng tên, đuôi chỉ thẳng xuống (spec §12.5).
    //
    // Bản trước lệch sang phải và vẫn đè lên bảng tên — che mất tên người chơi, thấy rõ
    // trong ảnh chụp 05/11. `topY` là mép trên của cụm bảng tên, do `PlayerView` đo và
    // truyền vào: chiều cao cụm đó đổi theo danh hiệu và cảnh giới nên không đoán được
    // bằng một hằng số.
    this.root.x = 0;
    this.root.y = topY - (R * 1.2) / zoom;
  }

  /** drawBubble vẽ khung tròn + đuôi + mặt nạ, ở cỡ pixel cố định. Chỉ chạy một lần. */
  private drawBubble(r: number): void {
    this.bubble.clear()
      // Đuôi vẽ TRƯỚC thân để chỗ nối nằm dưới hình tròn, không thành một vết gấp.
      .poly([-r * 0.62, r * 0.52, -r * 0.18, r * 0.80, -r * 0.95, r * 1.25])
      .fill(0xfff7e6)
      .circle(0, 0, r)
      .fill(0xfff7e6)
      .circle(0, 0, r)
      .stroke({ width: Math.max(2, r * 0.14), color: 0xffffff, alignment: 0.5 });
    this.mask.clear().circle(0, 0, r * 0.92).fill(0xffffff);
    this.fitFace(r);
  }

  /**
   * fitFace đặt lại cỡ ảnh mặt trong khung.
   *
   * Gọi cả lúc ĐỔI TEXTURE, không chỉ lúc đổi bán kính: bốn chân dung có tỉ lệ khác nhau
   * (`evil`/`panic` là bán thân, `neutral`/`cry` toàn thân), nên dùng lại số đo của ảnh
   * trước là méo.
   */
  private fitFace(r: number): void {
    const tex = this.face.texture;
    if (!tex || tex.width <= 0) {
      return;
    }
    // `center top / 180%` — y như ô đại diện HUD trong `style.css`.
    const w = r * 2 * ZOOM;
    this.face.width = w;
    this.face.height = w * (tex.height / tex.width);
    this.face.x = -w / 2;
    this.face.y = -r * 0.92;
  }

  destroy(): void {
    this.root.destroy({ children: true });
  }
}
