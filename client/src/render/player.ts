import { AnimatedSprite, Container, Graphics, Sprite, Text, type Texture } from 'pixi.js';

import { type GameConfig, realmName, realmOf, tienIndex, tierOf } from '../config';
import type { PlayerFrame } from '../net/state';
import { BUFFS, FLAG } from '../net/protocol';
import { assets, type AnimSet } from './assets';
import { animShort, initials } from './catalog';
import { BUFF_COLORS, charColor, hslToHex, REALM_COLORS, TITLE_COLORS } from './palette';
import { STR } from '../ui/strings';
import { SwordRing } from './swordring';

// Visual proportions relative to the body hitbox radius (not balance numbers).
const CHAR_HEIGHT_PER_R = 2.6; // drawn character height = 2.6 × body radius
const FEET_BELOW_CENTER = 0.75; // feet sit 0.75 × radius below the hitbox centre
const ART_FILL = 0.72; // characters fill 72% of their 512 canvas (ART_PROMPTS §1)
// Largest text box inside the oval of ui/speech_bubble.png, as fractions of the canvas.
const BUBBLE_INNER = { x0: 0.21, y0: 0.15, x1: 0.65, y1: 0.59 };
const BUBBLE_MAX_TEXT_W = 220;

// Animations missing from an atlas fall back along this chain.
const FALLBACK: Record<string, string[]> = {
  chase: ['run'],
  flee: ['run'],
  hit: ['idle'],
  levelup: ['idle'],
  breakthrough: ['levelup', 'idle'],
  realm_drop: ['hit', 'idle'],
  teleport_out: ['idle'],
  teleport_in: ['idle'],
  death: ['hit', 'idle'],
  respawn: ['idle'],
};

/** PlayerSprite draws one character: body, sword ring, aura, name tag, emote bubble. */
export class PlayerSprite {
  readonly root = new Container();
  private readonly swordRing: SwordRing;
  private readonly shadow = new Graphics();
  /** tienGround is the turning rune circle at a Tiên realm character's feet (spec §3.4). */
  private readonly tienGround = new Graphics();
  /** tienAura is the halo behind Tiên realm characters (spec §3.4). */
  private readonly tienAura = new Graphics();
  /** tienFront is the part of the Tiên effect drawn in front of the body, so it reads even
   * when the character is buried under hundreds of blades. */
  private readonly tienFront = new Graphics();
  private readonly aura = new Graphics();
  private readonly bodyBox = new Container();
  private body: AnimatedSprite | Container | null = null;
  private readonly label: Text;
  private readonly title: Text; // danh hiệu Top 1–10 phòng (spec §11.3)
  private titleRank = 0;
  private readonly realmTag: Text; // cảnh giới + tầng, ngay trên tên (spec §11.2)
  private realmLevel = -1;
  private readonly atlas: AnimSet | null;
  private mood = '';
  private bubble: Container | null = null;
  private readonly animLabel = new Text({ text: '', style: { fontSize: 40, fontWeight: '800', fill: 0xffffff, stroke: { color: 0x000000, width: 6 } } });
  private bubbleUntil = 0;

  constructor(
    private readonly cfg: GameConfig,
    readonly id: number,
    name: string,
    readonly char: string,
    private readonly isMe: boolean,
  ) {
    this.atlas = assets.atlas(`characters/${char}`);
    this.swordRing = new SwordRing(cfg);
    this.label = new Text({
      text: name,
      style: { fontFamily: 'system-ui, sans-serif', fontSize: 15, fontWeight: '700', fill: isMe ? 0xfff3a0 : 0xffffff, stroke: { color: 0x000000, width: 4 } },
    });
    this.label.anchor.set(0.5, 1);
    this.title = new Text({
      text: '',
      style: { fontFamily: 'system-ui, sans-serif', fontSize: 13, fontWeight: '800', fill: 0xffd700, stroke: { color: 0x000000, width: 4 } },
    });
    this.title.anchor.set(0.5, 1);
    this.title.visible = false;
    this.realmTag = new Text({
      text: '',
      style: { fontFamily: 'system-ui, sans-serif', fontSize: 13, fontWeight: '700', fill: 0xbfe9ff, stroke: { color: 0x000000, width: 4 } },
    });
    this.realmTag.anchor.set(0.5, 1);
    // Hào quang nằm sau lưng: dưới sprite nhân vật, trên mặt đất (spec §3.4).
    // Vòng ký tự Tiên nằm dưới cùng (sát đất), hạt sáng Tiên nằm trên cùng (trước mặt).
    this.root.addChild(this.tienGround, this.shadow, this.tienAura, this.aura, this.bodyBox, this.tienFront, this.label, this.realmTag, this.title);
    if (!this.atlas) {
      this.body = this.placeholderBody();
      this.bodyBox.addChild(this.body);
    }
  }

  // Placeholder (no art yet): coloured circle + character initials + current animation name.
  private placeholderBody(): Container {
    const [main, accent] = charColor(this.char);
    const c = new Container();
    const g = new Graphics().circle(0, 0, 1).fill(main).stroke({ color: 0x000000, width: 0.12 });
    g.circle(0, -0.55, 0.32).fill(accent);
    const t = new Text({ text: initials(this.char || '?'), style: { fontSize: 44, fontWeight: '900', fill: 0xffffff, stroke: { color: 0x000000, width: 6 } } });
    t.anchor.set(0.5);
    t.scale.set(0.02);
    t.y = 0.1;
    this.animLabel.anchor.set(0.5, 0);
    this.animLabel.scale.set(0.016);
    this.animLabel.y = 1.25;
    c.addChild(g, t, this.animLabel);
    return c;
  }

  /** swordCount: số kiếm đang vẽ, cho đồng hồ đo hiệu năng ở chế độ debug. */
  get swordCount(): number {
    return this.swordRing.count;
  }

  /** ring is the sword ring layer; the scene draws all rings above characters and obstacles. */
  get ring(): Container {
    return this.swordRing.root;
  }

  update(f: PlayerFrame, zoom: number, now: number, showHitbox: boolean, swordDetail = 1): void {
    const s = f.s;
    const r = this.cfg.player.body_r * s.sc;
    this.root.position.set(f.x, f.y);
    this.root.zIndex = f.y + r * FEET_BELOW_CENTER; // y-sort by the feet (spec §8.2)
    this.drawShadow(r);
    this.updateBody(s.md, r, Math.cos(s.f) < 0);
    this.drawTienAura(s.L, r, now);
    this.drawAura(s.bf, r, now);
    if (s.fl & FLAG.burning) {
      this.drawBurn(r, now); // vẽ đè lên aura buff
    }
    this.applyAlpha(s.fl, now);
    // Xếp từ dưới lên: tên → cảnh giới → danh hiệu (spec §11.2, §11.3).
    const headY = -r * (CHAR_HEIGHT_PER_R - FEET_BELOW_CENTER) - 4 / zoom;
    this.label.position.set(0, headY);
    this.label.scale.set(1 / zoom);
    const realmY = this.updateRealmTag(s.L, headY, zoom);
    this.updateTitle(s.ti ?? 0, realmY, zoom);
    this.swordRing.update(f, now, showHitbox, swordDetail);
    this.updateBubble(r, zoom, now);
  }

  /** updateRealmTag draws "Trúc Cơ tầng 3" above the name; returns the y to stack onto. */
  private updateRealmTag(level: number, headY: number, zoom: number): number {
    if (level !== this.realmLevel) {
      this.realmLevel = level;
      const r = realmOf(this.cfg, level);
      // Tiên giới không chia tầng nên chỉ ghi tên cảnh giới (spec §3.4).
      this.realmTag.text = tienIndex(this.cfg, r) >= 0 ? realmName(this.cfg, r) : STR.realmTier(realmName(this.cfg, r), tierOf(this.cfg, level));
      this.realmTag.style.fill = REALM_COLORS[r] ?? 0xffd700;
    }
    const y = headY - this.label.height - 1 / zoom;
    this.realmTag.position.set(0, y);
    this.realmTag.scale.set(1 / zoom);
    return y;
  }

  // Danh hiệu nằm trên cùng. Server đã bỏ `ti` với người đang Ẩn Thân nên danh hiệu tự
  // ẩn theo, không làm lộ vị trí (spec §11.3).
  private updateTitle(rank: number, headY: number, zoom: number): void {
    if (rank !== this.titleRank) {
      this.titleRank = rank;
      const name = STR.titles[rank - 1] ?? '';
      this.title.text = rank === 1 ? `👑 ${name}` : name;
      this.title.style.fill = TITLE_COLORS[rank - 1] ?? 0xffffff;
      this.title.visible = rank > 0;
    }
    if (!this.title.visible) {
      return;
    }
    this.title.position.set(0, headY - this.label.height - 2 / zoom);
    this.title.scale.set(1 / zoom);
  }

  private drawShadow(r: number): void {
    this.shadow.clear().ellipse(0, r * FEET_BELOW_CENTER, r * 0.9, r * 0.35).fill({ color: 0x000000, alpha: 0.25 });
  }

  private updateBody(mood: string, r: number, flip: boolean): void {
    if (this.atlas) {
      this.setMood(mood);
      const h = (r * CHAR_HEIGHT_PER_R) / ART_FILL; // world height of the whole 512 canvas
      const sp = this.body as AnimatedSprite;
      const k = h / sp.texture.height;
      sp.scale.set(flip ? -k : k, k);
      sp.position.set(0, r * FEET_BELOW_CENTER);
      return;
    }
    const body = this.body as Container;
    if (this.mood !== mood) {
      this.mood = mood;
      this.animLabel.text = animShort(mood);
    }
    // The placeholder circle is symmetric: never mirror it, so the initials stay readable.
    body.scale.set(r, r);
    const squash = mood === 'hit' || mood === 'death' ? 0.7 : 1;
    body.scale.y = r * squash;
  }

  private setMood(mood: string): void {
    if (mood === this.mood || !this.atlas) {
      return;
    }
    this.mood = mood;
    const anim = this.findAnim(mood);
    if (!anim) {
      return;
    }
    if (!this.body) {
      const sp = new AnimatedSprite(anim.textures);
      sp.anchor.set(this.atlas.pivotX, this.atlas.pivotY);
      this.body = sp;
      this.bodyBox.addChild(sp);
    }
    const sp = this.body as AnimatedSprite;
    sp.textures = anim.textures;
    sp.animationSpeed = anim.fps / 60;
    sp.loop = anim.loop;
    sp.gotoAndPlay(0);
  }

  private findAnim(mood: string): { textures: Texture[]; fps: number; loop: boolean } | null {
    const set = this.atlas?.anims;
    if (!set) {
      return null;
    }
    for (const name of [mood, ...(FALLBACK[mood] ?? []), 'idle']) {
      const a = set.get(name);
      if (a) {
        return a;
      }
    }
    return set.values().next().value ?? null;
  }

  /**
   * drawTienAura is the whole Tiên realm look (spec §3.4): a turning rune circle on the
   * ground, a breathing halo behind the body and ascending motes in front of it, all in the
   * realm's own colour. Đạo Tổ gets a yin-yang with a rainbow rim instead of a flat colour.
   *
   * Ba lớp chứ không chỉ một vành mờ sau lưng: ở cảnh giới Tiên nhân vật ngồi giữa hàng trăm
   * thanh kiếm, một vòng sáng nhỏ sau lưng gần như không thấy. Vòng ký tự dưới chân và hạt
   * sáng trước mặt là hai thứ luôn đọc được ngay.
   */
  private drawTienAura(level: number, r: number, now: number): void {
    const back = this.tienAura.clear();
    const ground = this.tienGround.clear();
    const front = this.tienFront.clear();
    const idx = tienIndex(this.cfg, realmOf(this.cfg, level));
    if (idx < 0) {
      return;
    }
    const spec = this.cfg.cultivation.tien_aura_colors[idx] ?? '#ffffff';
    const rainbow = spec === 'yin_yang';
    const color = rainbow ? hslToHex((now * 60) % 360, 0.9, 0.62) : Number(`0x${spec.replace('#', '')}`);
    this.drawTienGround(ground, idx, color, r, now);
    this.drawTienHalo(back, idx, color, rainbow, r, now);
    this.drawTienMotes(front, idx, color, r, now);
  }

  /** Vòng ký tự Tiên dưới chân: hai vành dẹt xoay chậm cùng các vạch chia. */
  private drawTienGround(g: Graphics, idx: number, color: number, r: number, now: number): void {
    const cy = r * FEET_BELOW_CENTER;
    const rad = r * 2.5;
    const squash = 0.4; // nhìn nghiêng từ trên xuống
    const spin = now * (0.3 + idx * 0.1);
    g.ellipse(0, cy, rad, rad * squash).fill({ color, alpha: 0.1 });
    g.ellipse(0, cy, rad, rad * squash).stroke({ color, width: r * 0.1, alpha: 0.6 });
    g.ellipse(0, cy, rad * 0.72, rad * 0.72 * squash).stroke({ color, width: r * 0.06, alpha: 0.45 });
    const ticks = 6 + idx * 2; // cảnh giới càng cao càng nhiều vạch
    for (let i = 0; i < ticks; i++) {
      const a = spin + (i / ticks) * Math.PI * 2;
      const cx = Math.cos(a);
      const sy = Math.sin(a) * squash;
      g.moveTo(cx * rad * 0.76, cy + sy * rad * 0.76).lineTo(cx * rad * 1.06, cy + sy * rad * 1.06);
    }
    g.stroke({ color, width: r * 0.1, alpha: 0.75 });
  }

  /** Hào quang sau lưng: ba lớp sáng nhòe dần + chữ ký riêng của từng cảnh giới. */
  private drawTienHalo(g: Graphics, idx: number, color: number, rainbow: boolean, r: number, now: number): void {
    const breathe = 1 + 0.06 * Math.sin(now * Math.PI); // chu kỳ 2s
    const rad = r * CHAR_HEIGHT_PER_R * 1.45 * breathe;
    const cy = -r * 0.9; // ngang ngực
    if (rainbow) {
      // Đạo Tổ: vòng âm dương xoay chậm, viền đổi màu liên tục.
      const a = now * 0.6;
      g.circle(0, cy, rad).fill({ color: 0x101010, alpha: 0.5 });
      g.arc(0, cy, rad, a, a + Math.PI).fill({ color: 0xf2f2f2, alpha: 0.5 });
      g.circle(Math.cos(a) * rad * 0.5, cy + Math.sin(a) * rad * 0.5, rad * 0.5).fill({ color: 0xf2f2f2, alpha: 0.5 });
      g.circle(-Math.cos(a) * rad * 0.5, cy - Math.sin(a) * rad * 0.5, rad * 0.5).fill({ color: 0x101010, alpha: 0.5 });
      g.circle(0, cy, rad).stroke({ color, width: r * 0.16, alpha: 0.95 });
      return;
    }
    for (const [k, a] of [[1.35, 0.08], [1.1, 0.13], [1, 0.2]] as const) {
      g.circle(0, cy, rad * k).fill({ color, alpha: a });
    }
    g.circle(0, cy, rad).stroke({ color, width: r * 0.12, alpha: 0.8 });
    switch (idx) {
      case 0: // Chân Tiên: mây trắng trôi quanh người
        for (let i = 0; i < 5; i++) {
          const a = now * 0.5 + (i / 5) * Math.PI * 2;
          g.ellipse(Math.cos(a) * rad * 0.85, cy + Math.sin(a) * rad * 0.85, rad * 0.26, rad * 0.12).fill({ color, alpha: 0.55 });
        }
        break;
      case 1: // Kim Tiên: tia sáng vàng tỏa ra, dài ngắn xen kẽ
        for (let i = 0; i < 16; i++) {
          const a = (i / 16) * Math.PI * 2 + now * 0.3;
          const long = i % 2 === 0 ? 1.3 : 1.12;
          g.moveTo(Math.cos(a) * rad * 0.8, cy + Math.sin(a) * rad * 0.8).lineTo(Math.cos(a) * rad * long, cy + Math.sin(a) * rad * long);
        }
        g.stroke({ color, width: r * 0.09, alpha: 0.85 });
        break;
      case 2: // Thái Ất Ngọc Tiên: hai cung ngọc chạy ngược chiều
        for (const dir of [1, -1]) {
          const a = now * 0.7 * dir;
          g.arc(0, cy, rad * (dir > 0 ? 1.18 : 1.04), a, a + Math.PI * 0.9).stroke({ color, width: r * 0.1, alpha: 0.9 });
        }
        break;
      case 3: // Đại La Kim Tiên: ba quả cầu sáng bay quanh theo quỹ đạo nghiêng
        for (let i = 0; i < 3; i++) {
          const a = now * 1.1 + (i / 3) * Math.PI * 2;
          g.circle(Math.cos(a) * rad * 1.15, cy + Math.sin(a) * rad * 0.5, r * 0.3).fill({ color, alpha: 0.9 });
        }
        g.circle(0, cy, rad * 1.15).stroke({ color, width: r * 0.05, alpha: 0.35 });
        break;
      default:
    }
  }

  /** Hạt sáng bay lên trước mặt — lớp duy nhất không bao giờ bị kiếm che. */
  private drawTienMotes(g: Graphics, idx: number, color: number, r: number, now: number): void {
    const n = 6 + idx * 2;
    const feet = r * FEET_BELOW_CENTER;
    const rise = r * (CHAR_HEIGHT_PER_R + 1.2); // bay từ chân lên quá đầu
    for (let i = 0; i < n; i++) {
      const t = (now * 0.4 + i / n) % 1;
      const a = i * 2.399 + now * 0.5; // góc vàng: rải đều, không thành hàng
      const x = Math.cos(a) * r * 1.5 * (1 - t * 0.55);
      const y = feet - rise * t;
      g.circle(x, y, r * 0.16 * (1 - t * 0.7)).fill({ color, alpha: 0.95 * (1 - t) });
    }
  }

  /** drawBurn is the blood-red flame ring of Thiêu Đốt Tinh Huyết (spec §4.1c). */
  private drawBurn(r: number, now: number): void {
    const flicker = 1 + 0.12 * Math.sin(now * 22);
    this.aura
      .circle(0, r * 0.2, r * 1.3 * flicker)
      .fill({ color: 0xff3b1f, alpha: 0.16 })
      .stroke({ color: 0xff6b2c, width: r * 0.14, alpha: 0.9 });
    // Lưỡi lửa liếm lên quanh chân, lệch pha nhau cho đỡ đều đặn.
    for (let i = 0; i < 6; i++) {
      const a = (i / 6) * Math.PI * 2 + now * 3;
      const h = r * (0.45 + 0.3 * Math.sin(now * 14 + i * 1.7));
      this.aura
        .circle(Math.cos(a) * r * 1.1, r * 0.2 + Math.sin(a) * r * 0.45 - h * 0.4, h * 0.35)
        .fill({ color: i % 2 ? 0xffb020 : 0xff4a1a, alpha: 0.5 });
    }
  }

  private drawAura(bits: number, r: number, now: number): void {
    this.aura.clear();
    let ring = 0;
    BUFFS.forEach((b, i) => {
      if (!(bits & (1 << i)) || !BUFF_COLORS[b]) {
        return;
      }
      const pulse = 1 + 0.06 * Math.sin(now * 6 + i);
      if (b === 'shield') {
        this.aura.circle(0, -r * 0.4, r * 1.5 * pulse).fill({ color: BUFF_COLORS[b], alpha: 0.18 }).stroke({ color: BUFF_COLORS[b], width: r * 0.08, alpha: 0.8 });
        return;
      }
      this.aura.circle(0, r * 0.2, r * (1.15 + ring * 0.18) * pulse).stroke({ color: BUFF_COLORS[b], width: r * 0.1, alpha: 0.7 });
      ring++;
    });
  }

  private applyAlpha(flags: number, now: number): void {
    let a = 1;
    const stealth = (flags & FLAG.stealth) !== 0;
    if (stealth && !this.isMe) {
      // Spec §6.2: Ẩn Thân = tàng hình 80%, nhìn thấy lờ mờ chứ không biến mất hẳn.
      a = this.cfg.buffs.stealth?.visible_alpha ?? 0.2;
    } else if (flags & FLAG.bush || stealth) {
      a = 0.45; // in a bush, or my own stealth so I can tell the buff is up
    }
    // Ẩn Thân giấu luôn tên và danh hiệu của người khác.
    this.label.visible = !stealth || this.isMe;
    this.realmTag.visible = this.label.visible;
    if (stealth && !this.isMe) {
      this.title.visible = false;
    }
    if (flags & FLAG.invuln && Math.floor(now * 10) % 2 === 0) {
      a *= 0.4; // blink while invulnerable (spec §5.1)
    }
    if (flags & FLAG.transit) {
      a *= 0.5;
    }
    this.root.alpha = a;
    this.ring.alpha = flags & FLAG.stealth ? 0.45 : 1;
    this.swordRing.visible = !(flags & FLAG.dead);
  }

  // ---- emote bubble --------------------------------------------------------

  showBubble(text: string, now: number): void {
    this.bubble?.destroy({ children: true });
    const c = new Container();
    const t = new Text({ text, style: { fontFamily: 'system-ui, sans-serif', fontSize: 16, fontWeight: '800', fill: 0x111111, wordWrap: true, wordWrapWidth: BUBBLE_MAX_TEXT_W, align: 'center' } });
    t.anchor.set(0.5);
    const tex = assets.texture('ui/speech_bubble.png');
    const pad = 10;
    if (tex) {
      // Size the art so its white inner box (not the whole canvas) wraps the text.
      const bg = new Sprite(tex);
      const box = BUBBLE_INNER;
      bg.anchor.set((box.x0 + box.x1) / 2, (box.y0 + box.y1) / 2);
      bg.width = (t.width + pad * 2) / (box.x1 - box.x0);
      bg.height = (t.height + pad * 2) / (box.y1 - box.y0);
      c.addChild(bg);
    } else {
      c.addChild(new Graphics().roundRect(-t.width / 2 - pad, -t.height / 2 - pad / 2, t.width + pad * 2, t.height + pad, 12).fill(0xffffff).stroke({ color: 0, width: 3 }));
    }
    c.addChild(t);
    this.bubble = c;
    this.bubbleUntil = now + 1.8;
    this.root.addChild(c);
  }

  private updateBubble(r: number, zoom: number, now: number): void {
    if (!this.bubble) {
      return;
    }
    if (now > this.bubbleUntil) {
      this.bubble.destroy({ children: true });
      this.bubble = null;
      return;
    }
    this.bubble.scale.set(1 / zoom);
    this.bubble.position.set(r * 0.8, -r * (CHAR_HEIGHT_PER_R - FEET_BELOW_CENTER) - 34 / zoom);
  }

  destroy(): void {
    this.root.destroy({ children: true });
    this.swordRing.destroy();
  }
}
