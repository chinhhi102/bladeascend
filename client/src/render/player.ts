import { AnimatedSprite, Container, Graphics, Sprite, Text, type Texture } from 'pixi.js';

import type { GameConfig } from '../config';
import type { PlayerFrame } from '../net/state';
import { BUFFS, FLAG } from '../net/protocol';
import { assets, type AnimSet } from './assets';
import { animShort, initials } from './catalog';
import { BUFF_COLORS, charColor } from './palette';
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
  private readonly aura = new Graphics();
  private readonly bodyBox = new Container();
  private body: AnimatedSprite | Container | null = null;
  private readonly label: Text;
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
    isMe: boolean,
  ) {
    this.atlas = assets.atlas(`characters/${char}`);
    this.swordRing = new SwordRing(cfg);
    this.label = new Text({
      text: name,
      style: { fontFamily: 'system-ui, sans-serif', fontSize: 15, fontWeight: '700', fill: isMe ? 0xfff3a0 : 0xffffff, stroke: { color: 0x000000, width: 4 } },
    });
    this.label.anchor.set(0.5, 1);
    this.root.addChild(this.shadow, this.aura, this.bodyBox, this.label);
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

  /** ring is the sword ring layer; the scene draws all rings above characters and obstacles. */
  get ring(): Container {
    return this.swordRing.root;
  }

  update(f: PlayerFrame, zoom: number, now: number, showHitbox: boolean): void {
    const s = f.s;
    const r = this.cfg.player.body_r * s.sc;
    this.root.position.set(f.x, f.y);
    this.root.zIndex = f.y + r * FEET_BELOW_CENTER; // y-sort by the feet (spec §8.2)
    this.drawShadow(r);
    this.updateBody(s.md, r, Math.cos(s.f) < 0);
    this.drawAura(s.bf, r, now);
    this.applyAlpha(s.fl, now);
    this.label.position.set(0, -r * (CHAR_HEIGHT_PER_R - FEET_BELOW_CENTER) - 4 / zoom);
    this.label.scale.set(1 / zoom);
    this.swordRing.update(f, now, showHitbox);
    this.updateBubble(r, zoom, now);
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
    if (flags & FLAG.stealth || flags & FLAG.bush) {
      a = 0.45; // only visible because we're allowed to see it (self or close by)
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
