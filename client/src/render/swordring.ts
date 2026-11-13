import { Container, Graphics, Particle, ParticleContainer, Texture } from 'pixi.js';

import { type GameConfig, realmOf } from '../config';
import { statusOn } from '../net/protocol';
import { omega, type PlayerFrame } from '../net/state';
import { assets } from './assets';
import { REALM_COLORS } from './palette';
import { bladeLen, hiltR, ringCap, type RingGeom, ringGeom, ringOf, ringStart, slotAngle, tipDist } from './ringgeom';
import { declareTexture } from './gpumem';

// Drawing proportions (visual only, not balance numbers).
const HILT_ANCHOR = 0.95; // blade art: hilt sits at 95% of the canvas height, tip at the top
const ART_FILL = 0.9; // blade art fills ~90% of its canvas height
const GHOSTS = 2; // trailing copies per sword that fake motion blur (spec §4.1b)
const GHOST_STEP = (7 * Math.PI) / 180;
const MAIN_ALPHA = 0.95;
// Lớp trân châu trên vòng kiếm người bị bịt mắt (v6.27 §6.3): chỉ để nhận ra, không phải số cân bằng.
const PEARLS = 18;
const PEARL_COLOR = 0x3b2416;

let bladeTex: Texture | null = null;

/**
 * ringWrites: số lần ĐẶT LẠI một thanh kiếm (x, y, rotation) trong khung vừa rồi — bộ đếm
 * CÔNG VIỆC, không phải thời gian.
 *
 * Vì sao đếm thay vì bấm giờ: máy dev dùng chung với việc khác, và con số mili-giây nở ra
 * co lại theo tải máy (đo 11/11: cùng một commit cho ra 0,66% rồi 18,8% khung nhảy cóc chỉ
 * vì load 39 trên 12 nhân). Số LẦN GHI thì không đổi theo tải — nó là thứ chứng minh được
 * "trước 1 350 lượt mỗi khung, sau 2 lượt" dù máy đang bận.
 */
export const ringWork = { writes: 0, frame: 0, would: 0, wouldFrame: 0 };

/** ringWorkTick: gọi đầu mỗi khung vẽ — chốt số của khung trước rồi đếm lại từ 0. */
export function ringWorkTick(): void {
  ringWork.frame = ringWork.writes;
  ringWork.wouldFrame = ringWork.would;
  ringWork.writes = 0;
  ringWork.would = 0;
}

/** placeholderBlade: a white/grey blade so every ring can be tinted (spec v3.1 item 6). */
function placeholderBlade(): Texture {
  if (bladeTex) {
    return bladeTex;
  }
  const w = 24;
  const h = 128;
  const cv = document.createElement('canvas');
  cv.width = w;
  cv.height = h;
  const ctx = cv.getContext('2d') as CanvasRenderingContext2D;
  ctx.beginPath();
  ctx.moveTo(w / 2, 2);
  ctx.lineTo(w - 4, 20);
  ctx.lineTo(w - 5, h - 26);
  ctx.lineTo(5, h - 26);
  ctx.lineTo(4, 20);
  ctx.closePath();
  const grad = ctx.createLinearGradient(0, 0, w, 0);
  grad.addColorStop(0, '#d9d9d9');
  grad.addColorStop(0.5, '#ffffff');
  grad.addColorStop(1, '#bdbdbd');
  ctx.fillStyle = grad;
  ctx.fill();
  ctx.lineWidth = 2;
  ctx.strokeStyle = '#3a3a3a';
  ctx.stroke();
  ctx.fillStyle = '#8a8a8a';
  ctx.fillRect(1, h - 26, w - 2, 6); // guard
  ctx.fillRect(w / 2 - 3, h - 20, 6, 16); // grip
  bladeTex = Texture.from(cv);
  declareTexture(bladeTex, 'fx'); // §13.10: lưỡi kiếm dự phòng
  return bladeTex;
}

function parseTint(hex: string): number {
  return Number.parseInt(hex.replace('#', ''), 16) || 0xffffff;
}

interface Blade {
  main: Particle;
  ghosts: Particle[];
  ring: number;
  /** Góc cố định của khe này so với `ringAngle`, cùng cos/sin đã tính sẵn. */
  off: number;
  cosOff: number;
  sinOff: number;
  /** Chiều cao gốc kiếm và tỉ lệ co, chỉ phụ thuộc vòng nên tính một lần lúc dựng. */
  hilt: number;
  len: number;
}

/**
 * SwordRing draws a player's nested sword rings (spec §4.1a/b): outer rings first, every ring
 * tinted with ring_tints, blades leaning by the server's tilt, a spin-blur annulus per ring and
 * fading ghost copies so a 300 ms revolution never looks frozen or reversed.
 */
export class SwordRing {
  readonly root = new Container();
  private readonly blur = new Graphics();
  /**
   * Hai bể kiếm khai TĨNH hết — không một thuộc tính nào `dynamic`.
   *
   * Pixi gói lại buffer thuộc tính `dynamic` cho TỪNG hạt, MỖI khung, bất kể hạt có đổi hay
   * không (`ParticleBuffer.update`: buffer tĩnh chỉ gói lại khi `_childrenDirty`). Bản cũ
   * khai cả bốn là dynamic, nên một Đại Thừa 1 050 kiếm vẫn tốn 1 050 lượt gói × 4 thuộc
   * tính × mỗi khung × mỗi người — ngay cả sau khi v6.52 đã làm JS không ghi gì nữa. Việc
   * đó không nằm trong cột `p.ring` của bảng F3 (nó xảy ra trong `app.render()`), nên suốt
   * hai lượt tối ưu nó vô hình.
   *
   * Góc quay của cả vòng là `rotation` của CHÍNH container này — một phép ghi, không phải
   * thuộc tính hạt — nên vòng vẫn quay mà buffer không cần gói lại. Đổi chỗ kiếm thật (đổi
   * số kiếm, lên tầng, nghiêng sau cú va) thì gọi `markDirty()`.
   */
  private readonly tinted = new ParticleContainer({ dynamicProperties: { position: false, rotation: false, vertex: false, color: false } });
  private readonly skin = new ParticleContainer({ dynamicProperties: { position: false, rotation: false, vertex: false, color: false } });
  private readonly cracks = new Graphics();
  private readonly hitbox = new Graphics();
  private readonly pearl = new Graphics();
  /** Viền màu quanh vòng kiếm theo kỹ năng (§6.5): Gồng cam, Trộm Kiếm tím, đàn em tím sẫm. */
  private readonly halo = new Graphics();
  private haloKey = '';
  /** haloColor: PlayerSprite đặt mỗi khung; 0 = không viền. */
  haloColor = 0;
  /** reach: bán kính ngoài của vòng kiếm khung vừa vẽ (cho hiệu ứng kỹ năng bám theo). */
  reach = 0;
  /** Bán kính lớp trân châu đang vẽ (0 = tắt) — chỉ vẽ lại khi bật/tắt hay vòng đổi cỡ. */
  private pearlR = 0;
  /**
   * layoutSeq: tăng mỗi lần `rebuild` xếp lại các thanh kiếm.
   *
   * Chữ ký của vành mờ và vết nứt phải đổi mỗi khi BỐ CỤC đổi. Dùng `blades.length` là
   * chưa đủ: đổi số kiếm cùng lúc đổi mức giãn chi tiết có thể ra đúng số thanh cũ với góc
   * khe MỚI, và khi đó hình vẽ phải đổi mà chữ ký thì không. Một bộ đếm thì không có kẽ hở
   * nào kiểu đó — và chỗ này là loại lỗi chỉ hiện ra bằng mắt, rất lâu sau, trên máy người
   * chơi.
   */
  private layoutSeq = 0;
  /** Chữ ký vành mờ / vết nứt / hộp va chạm đang vẽ — '' nghĩa là Graphics đang rỗng. */
  private blurKey = '';
  private crackKey = '';
  private hitboxKey = '';
  private blades: Blade[] = [];
  /** localKey: chữ ký bố cục cục bộ đang áp (xem `placeAll`). '' = phải đặt lại. */
  private localKey = '';
  private realm = -1;
  /** count: số kiếm đang vẽ (đồng hồ debug đọc qua PlayerSprite.swordCount). */
  count = 0;
  private prevSlots: number[] = [];
  private tweenStart = 0;
  /** Chữ ký hình học đang áp: lên tầng (scale) hay đổi số vòng thì phải tính lại. */
  private geomKey = '';
  /** Mức giãn: 1 = vẽ mọi thanh, 2 = vẽ cách một thanh… chỉ áp cho vòng dày. */
  private detail = 1;
  private readonly tints: number[];

  constructor(private readonly cfg: GameConfig) {
    this.tints = cfg.sword.ring_tints.map(parseTint);
    this.root.addChild(this.halo, this.blur, this.tinted, this.skin, this.cracks, this.pearl, this.hitbox);
  }

  update(f: PlayerFrame, now: number, showHitbox: boolean, detail = 1): void {
    const s = f.s;
    const realm = realmOf(this.cfg, s.L);
    const g = ringGeom(this.cfg, s.sc, s.n);
    if (s.n !== this.count || realm !== this.realm || detail !== this.detail) {
      this.detail = detail;
      this.rebuild(s.n, realm, g, now);
    }
    // Hình học chỉ đổi khi lên tầng (scale) hoặc mở vòng mới, không đổi mỗi khung —
    // nên bán kính, chiều dài và tỉ lệ của từng thanh kiếm tính sẵn ở đây, không tính lại
    // trong vòng lặp nóng.
    const key = `${g.rings}:${s.sc}:${g.width}`;
    if (key !== this.geomKey) {
      this.geomKey = key;
      this.refreshGeom(g);
    }
    this.root.position.set(f.x, f.y);
    this.root.zIndex = f.y + 1;
    const t = Math.min(1, (now - this.tweenStart) / this.cfg.sword.tween_secs);
    // Vệt mờ theo đúng tốc quay thật (nội tại + Tốc Kiếm, luật cộng §4.0) so với base.
    const spin = omega(this.cfg, s, f.view.char) / ((2 * Math.PI) / this.cfg.sword.rev_secs);
    this.drawBlur(g, s.tl, realm, spin);
    // `would`: đường CŨ sẽ ghi bấy nhiêu ở đúng khung này — mỗi thanh kiếm một thân + hai
    // vệt mờ, mỗi khung, mỗi người. Đếm nó ngay cạnh số ghi THẬT để bảng so sánh là hai con
    // số đo trong CÙNG một cảnh, không phải một con số đo và một con số suy ra.
    ringWork.would += (1 + GHOSTS) * this.blades.length;
    this.placeAll(f, t, s.tl, s.d);
    this.drawCracks(f, g);
    this.drawPearl(f, g);
    this.drawHalo(g, now);
    this.drawHitbox(f, g, showHitbox);
  }

  /** drawHalo: vẽ lại chỉ khi màu/cỡ đổi; nhịp sáng bằng alpha, không vẽ lại mỗi khung. */
  private drawHalo(g: RingGeom, now: number): void {
    this.reach = g.reach;
    const want = this.count > 0 ? this.haloColor : 0;
    const key = `${want}:${g.reach}`;
    if (key !== this.haloKey) {
      this.haloKey = key;
      this.halo.clear();
      if (want) {
        this.halo.circle(0, 0, g.reach * 1.04).stroke({ color: want, width: Math.max(4, g.reach * 0.05), alpha: 0.9 });
        this.halo.circle(0, 0, g.reach * 1.04).fill({ color: want, alpha: 0.08 });
      }
    }
    this.halo.visible = want !== 0;
    this.halo.alpha = 0.7 + 0.3 * Math.sin(now * 8);
  }

  /**
   * drawPearl: người bị Thằng Mập bịt mắt thì kiếm không chém chết ai (`blind_no_body_hit`),
   * nên người khác phải NHÌN ra điều đó — một lớp trân châu mờ phủ lên vòng kiếm, quay theo vòng.
   */
  private drawPearl(f: PlayerFrame, g: RingGeom): void {
    const want = statusOn(f.s.st, 'blind') && f.s.n > 0 ? g.reach : 0;
    if (want !== this.pearlR) {
      this.pearlR = want;
      const gr = this.pearl.clear();
      if (want > 0) {
        gr.circle(0, 0, want).fill({ color: PEARL_COLOR, alpha: 0.18 });
        const dot = Math.max(4, want * 0.06);
        for (let i = 0; i < PEARLS; i++) {
          const a = (i * Math.PI * 2) / PEARLS;
          const r = want * (i % 2 === 0 ? 0.62 : 0.85);
          gr.circle(Math.cos(a) * r, Math.sin(a) * r, dot);
        }
        gr.fill({ color: PEARL_COLOR, alpha: 0.55 });
      }
    }
    this.pearl.visible = want > 0;
    this.pearl.rotation = f.ringAngle;
  }

  /**
   * placeAll — v6.52: KHÔNG đặt lại từng thanh kiếm mỗi khung nữa.
   *
   * Góc của mỗi khe là CỐ ĐỊNH so với `ringAngle`. Nên đặt kiếm một lần theo toạ độ CỤC BỘ
   * (coi như vòng đang ở góc 0) rồi mỗi khung chỉ **xoay cái container**: hai phép ghi cho
   * cả vòng, thay cho 3 × số kiếm phép ghi.
   *
   * Đo được (bộ đếm `ringWork`, không phụ thuộc tải máy): một Đại Thừa 450 kiếm × (1 thân +
   * 2 vệt mờ) = **1 350 lượt ghi mỗi khung mỗi người**; nhân chục người trong tầm nhìn là
   * hàng chục nghìn lượt mỗi khung. Sau khi đổi: **0 lượt** ở khung thường, và chỉ đặt lại
   * khi có lý do thật — đổi số kiếm, lên tầng, đổi chiều quay, hoặc trong lúc nghiêng kiếm
   * đang tween (`tilt`) sau một cú va.
   *
   * Phần tween sau khi số kiếm đổi vẫn đi đường chậm: lúc đó mỗi khe thật sự trượt về chỗ
   * mới nên không có cách nào gộp thành một phép xoay.
   */
  private placeAll(f: PlayerFrame, t: number, tilt: number, dir: number): void {
    // Container mang góc quay; mọi thứ bên trong nằm ở hệ toạ độ "vòng đứng yên".
    this.tinted.rotation = f.ringAngle;
    this.skin.rotation = f.ringAngle;
    if (t < 1) {
      this.localKey = ''; // đang tween: đặt lại mỗi khung, và nhớ đặt lại lần cuối khi xong
      for (let i = 0; i < this.blades.length; i++) {
        this.placeLocal(this.blades[i], this.slot(i, this.blades[i].off, t), tilt, dir);
      }
      this.markDirty();
      return;
    }
    // Chữ ký của BỐ CỤC CỤC BỘ: chỉ ba thứ này đổi thì mới phải đặt lại.
    //
    // `tilt` làm tròn 3 số lẻ: nó tween mượt sau mỗi cú va (`tilt_tween_secs`), nên trong
    // khoảng đó vẫn đặt lại mỗi khung — đúng như cũ — rồi đứng yên hẳn và thôi.
    const key = `${this.blades.length}:${dir}:${tilt.toFixed(3)}`;
    if (key === this.localKey) {
      return;
    }
    this.localKey = key;
    for (const b of this.blades) {
      this.placeLocal(b, b.off, tilt, dir);
    }
    this.markDirty();
  }

  /**
   * markDirty báo Pixi gói lại buffer TĨNH của hai bể kiếm ở khung này.
   *
   * Quên gọi nó sau khi dời kiếm thì kiếm đứng im ở chỗ cũ — mà "đứng im" là hỏng lặng lẽ,
   * không ném lỗi, không hiện trong bảng số. Nên mọi đường ghi vào hạt đều phải đi qua đây:
   * `rebuild` (thêm/bớt hạt), `refreshGeom` (đổi tỉ lệ), `placeAll` (dời chỗ).
   */
  private markDirty(): void {
    this.tinted.update();
    this.skin.update();
  }

  /** placeLocal đặt một thanh kiếm (và vệt mờ của nó) theo góc khe CỤC BỘ `off`. */
  private placeLocal(b: Blade, off: number, tilt: number, dir: number): void {
    const turn = tilt + Math.PI / 2;
    const set = (p: Particle, ang: number): void => {
      p.x = Math.cos(ang) * b.hilt;
      p.y = Math.sin(ang) * b.hilt;
      p.rotation = ang + turn;
      ringWork.writes++;
    };
    set(b.main, off);
    for (let j = 0; j < b.ghosts.length; j++) {
      set(b.ghosts[j], off - dir * GHOST_STEP * (j + 1));
    }
  }

  /** slot is sword i's angle offset, tweened for tween_secs after the count changes. */
  private slot(i: number, to: number, t: number): number {
    const from = this.prevSlots[i];
    if (from === undefined || t >= 1) {
      return to;
    }
    let d = to - from;
    d = Math.atan2(Math.sin(d), Math.cos(d));
    return from + d * t;
  }



  /**
   * drawBlur paints the motion-blur annulus of each ring — but ONLY for rings that really do
   * sweep their whole annulus every tick, using the same test the server uses to decide a
   * ring is a solid damage band (spec §4.1b: `số kiếm × góc quét ≥ 360°`).
   *
   * Without that test a ring holding 1 sword out of 80 still painted a full faint circle, so
   * the drawn range looked far wider than what the blades can actually hit.
   */
  private drawBlur(g: RingGeom, tilt: number, realm: number, spin: number): void {
    // Vành mờ KHÔNG đổi mỗi khung: nó chỉ phụ thuộc hình học vòng, số kiếm, độ nghiêng, cảnh
    // giới và tốc quay. `clear()` một Graphics là đánh dấu dựng lại toàn bộ hình học của nó,
    // kể cả khi vẽ ra đúng hình cũ — và chỗ này chạy cho MỌI người trong tầm nhìn, mỗi khung.
    //
    // `spin` làm tròn 1/50: nó chỉ vào alpha (≤ 0,4), nên một bậc là 0,008 độ mờ — mắt không
    // thấy, mà vòng chạy đều thì chữ ký đứng yên hẳn thay vì nhảy theo số thực.
    const key = `${this.layoutSeq}:${g.rings}:${g.width}:${g.reach}:${this.count}:${tilt.toFixed(3)}:${realm}:${Math.round(spin * 50)}`;
    if (key === this.blurKey) {
      return;
    }
    this.blurKey = key;
    const gr = this.blur.clear();
    const sweepPerTick = (2 * Math.PI * spin) / (this.cfg.sword.rev_secs * this.cfg.net.tick_hz);
    for (let k = g.rings - 1; k >= 0; k--) {
      const cap = ringCap(g, k);
      const inRing = Math.min(cap, this.count - ringStart(g, k));
      if (inRing * sweepPerTick < 2 * Math.PI) {
        continue; // vòng thưa: từng lưỡi kiếm nhìn rõ, không có vành mờ liền
      }
      const h = hiltR(g, k);
      const tip = tipDist(g, k, tilt);
      const alpha = Math.min(0.4, (0.08 + 0.22 * (inRing / cap)) * spin);
      gr.circle(0, 0, (h + tip) / 2).stroke({ color: this.ringTint(k, realm, true), width: tip - h, alpha });
    }
  }

  private ringTint(k: number, realm: number, forBlur: boolean): number {
    if (k === 0) {
      // Ring 1 keeps the realm skin's own colour so players can read the realm (spec §4.1a).
      return this.hasSkin(realm) && !forBlur ? 0xffffff : (REALM_COLORS[realm] ?? 0xffffff);
    }
    return this.tints[k % this.tints.length] ?? 0xffffff;
  }

  private hasSkin(realm: number): boolean {
    return assets.texture(`swords/realm_${realm}.png`) !== null;
  }

  private rebuild(n: number, realm: number, g: RingGeom, now: number): void {
    // Giữ góc cũ của chính các thanh đang vẽ: khi giãn chi tiết thì chỉ số mảng không còn
    // khớp chỉ số kiếm nữa, nên không thể tính lại bằng slotAngle(count cũ).
    this.prevSlots = this.blades.map((b) => b.off);
    this.layoutSeq++;
    this.tweenStart = now;
    this.localKey = ''; // bố cục cũ không còn đúng — buộc đặt lại ở khung kế
    this.count = n;
    this.realm = realm;
    this.tinted.removeParticles();
    this.skin.removeParticles();
    this.blades = [];
    const skinTex = assets.texture(`swords/realm_${realm}.png`);
    const plain = placeholderBlade();
    const byRing: Blade[][] = [];
    // Vòng nào quét kín cả vành mỗi tick thì coi là dày (cùng phép thử với drawBlur), và
    // `minDense` là số kiếm tối thiểu để còn kín. Giãn chi tiết chỉ được phép tới mức vòng
    // VẪN còn kín — bỏ quá tay thì vành sát thương liền biến thành mấy thanh kiếm rời rạc,
    // nhìn ra ngay và còn nói dối người chơi về tầm chém.
    const minDense = Math.ceil(this.cfg.sword.rev_secs * this.cfg.net.tick_hz);
    const denseRings: boolean[] = [];
    const ringDetail: number[] = [];
    for (let k = 0; k < g.rings; k++) {
      const inRing = Math.min(ringCap(g, k), n - ringStart(g, k));
      denseRings[k] = inRing >= minDense;
      ringDetail[k] = denseRings[k] ? Math.max(1, Math.min(this.detail, Math.floor(inRing / minDense))) : 1;
    }
    for (let i = 0; i < n; i++) {
      const k = ringOf(g, i);
      // Vòng dày vốn là một vành sát thương liền (§4.1b) nên bỏ bớt thanh trong vành vẫn
      // nhìn y hệt; vòng thưa thì từng thanh là một lưỡi kiếm riêng, phải vẽ đủ.
      if (ringDetail[k] > 1 && (i - ringStart(g, k)) % ringDetail[k] !== 0) {
        continue;
      }
      const tex = k === 0 && skinTex ? skinTex : plain;
      const tint = this.ringTint(k, realm, false);
      const mk = (alpha: number): Particle => new Particle({ texture: tex, anchorX: 0.5, anchorY: HILT_ANCHOR, tint, alpha });
      const off = slotAngle(g, n, i);
      // Vòng dày thì hai thanh cạnh nhau chỉ cách vài độ, vệt mờ 7°/bước nằm trùng lên thanh
      // bên cạnh nên vẽ cũng không ai thấy — bỏ đi, tiết kiệm 2/3 số particle của vòng to.
      const ghosts = denseRings[k] ? 0 : GHOSTS;
      const b: Blade = {
        main: mk(MAIN_ALPHA),
        ghosts: Array.from({ length: ghosts }, (_, j) => mk(0.35 / (j + 1))),
        ring: k,
        off,
        cosOff: Math.cos(off),
        sinOff: Math.sin(off),
        hilt: hiltR(g, k),
        len: bladeLen(g, k),
      };
      this.blades.push(b);
      (byRing[k] ??= []).push(b);
    }
    // Outer rings first (drawn underneath), ring 1 last (on top); ghosts under their blades.
    for (let k = byRing.length - 1; k >= 0; k--) {
      const target = k === 0 && skinTex ? this.skin : this.tinted;
      for (const b of byRing[k] ?? []) {
        b.ghosts.forEach((p) => target.addParticle(p));
      }
      for (const b of byRing[k] ?? []) {
        target.addParticle(b.main);
      }
    }
    this.geomKey = ''; // particle mới chưa có tỉ lệ: buộc refreshGeom chạy ngay khung này
  }

  /** refreshGeom tính lại bán kính gốc kiếm, chiều dài và tỉ lệ vẽ của mọi thanh. */
  private refreshGeom(g: RingGeom): void {
    this.localKey = ''; // bán kính chuôi đổi → chỗ đứng của mọi thanh kiếm đổi theo
    for (const b of this.blades) {
      b.hilt = hiltR(g, b.ring);
      b.len = bladeLen(g, b.ring);
      for (const p of [b.main, ...b.ghosts]) {
        const ky = b.len / (p.texture.height * ART_FILL);
        p.scaleY = ky;
        p.scaleX = p.texture === bladeTex ? g.width / p.texture.width : ky;
      }
    }
    this.markDirty();
  }

  // Light crack lines across ring 1 while the player carries crack damage (spec §5.2).
  private drawCracks(f: PlayerFrame, g: RingGeom): void {
    const ck = f.s.ck ?? 0;
    // Vẽ bằng toạ độ CỤC BỘ rồi xoay cả Graphics, y như bể kiếm làm từ v6.52. Bản cũ cộng
    // `f.ringAngle` vào từng đường nứt, nên góc đổi mỗi khung và phải vẽ lại mỗi khung —
    // trong khi hình vẽ ra thì xoay đúng một góc. Giờ nó chỉ vẽ lại khi vết nứt thật sự đổi.
    //
    // `ck` làm tròn 1/50: nó chỉ vào alpha (0,35 + 0,5·ck).
    this.cracks.rotation = f.ringAngle;
    const key = ck <= 0 ? '' : `${this.layoutSeq}:${Math.round(ck * 50)}:${g.width}:${f.s.tl.toFixed(3)}:${hiltR(g, 0)}:${bladeLen(g, 0)}`;
    if (key === this.crackKey) {
      return;
    }
    this.crackKey = key;
    const gr = this.cracks.clear();
    if (ck <= 0) {
      return;
    }
    const h = hiltR(g, 0);
    const len = bladeLen(g, 0);
    const z = g.width * 0.4;
    for (let i = 0; i < Math.min(this.blades.length, ringCap(g, 0)); i++) {
      // Góc THẬT của thanh kiếm = góc khe cục bộ + góc quay của vòng (container đang mang
      // góc quay, nhưng `cracks` là Graphics riêng nên phải cộng tay).
      const a = this.blades[i]?.off ?? 0;
      const hx = Math.cos(a) * h;
      const hy = Math.sin(a) * h;
      const dx = Math.cos(a + f.s.tl);
      const dy = Math.sin(a + f.s.tl);
      const at = (t: number, side: number): [number, number] => [hx + dx * len * t - dy * z * side, hy + dy * len * t + dx * z * side];
      gr.moveTo(...at(0.3, 1))
        .lineTo(...at(0.5, -1))
        .lineTo(...at(0.7, 1));
    }
    gr.stroke({ color: 0xffffff, width: Math.max(1.5, g.width * 0.18), alpha: 0.35 + 0.5 * Math.min(1, ck) });
  }

  /** drawHitbox overlays the server's sword capsules (debug key H, spec §17). */
  private drawHitbox(f: PlayerFrame, g: RingGeom, show: boolean): void {
    // Cũng vẽ cục bộ rồi xoay: hộp va chạm bật bằng phím H nên gần như luôn TẮT, mà bản cũ
    // vẫn `clear()` nó mỗi khung cho mọi người chỉ để không vẽ gì.
    this.hitbox.rotation = f.ringAngle;
    const key = show ? `${f.s.n}:${g.width}:${g.reach}:${f.s.tl.toFixed(3)}:${f.s.sc}` : '';
    if (key === this.hitboxKey) {
      return;
    }
    this.hitboxKey = key;
    const gr = this.hitbox.clear();
    if (!show) {
      return;
    }
    for (let i = 0; i < f.s.n; i++) {
      const k = ringOf(g, i);
      const a = slotAngle(g, f.s.n, i);
      const h = hiltR(g, k);
      const hx = Math.cos(a) * h;
      const hy = Math.sin(a) * h;
      const len = bladeLen(g, k);
      gr.moveTo(hx, hy).lineTo(hx + Math.cos(a + f.s.tl) * len, hy + Math.sin(a + f.s.tl) * len);
    }
    gr.stroke({ color: 0xff2d55, width: g.width, alpha: 0.35, cap: 'round' });
    gr.circle(0, 0, g.reach).stroke({ color: 0xff2d55, width: 2 / Math.max(f.s.sc, 1), alpha: 0.6 });
  }

  set visible(v: boolean) {
    this.root.visible = v;
  }

  destroy(): void {
    this.root.destroy({ children: true });
  }
}
