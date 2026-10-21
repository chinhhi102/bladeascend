import { AnimatedSprite, Container, Graphics, Sprite, Text, type Texture } from 'pixi.js';

import { type GameConfig, realmOf } from '../config';
import type { PlayerFrame } from '../net/state';
import { BUFFS, FLAG } from '../net/protocol';
import { assets, type AnimSet } from './assets';
import { BUFF_COLORS, charColor, REALM_COLORS } from './palette';

// Visual proportions relative to the body hitbox radius (not balance numbers).
const CHAR_HEIGHT_PER_R = 2.6; // drawn character height = 2.6 × body radius
const FEET_BELOW_CENTER = 0.75; // feet sit 0.75 × radius below the hitbox centre
const ART_FILL = 0.72; // characters fill 72% of their 512 canvas (ART_PROMPTS §1)

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
  readonly ring = new Container();
  private readonly shadow = new Graphics();
  private readonly aura = new Graphics();
  private readonly bodyBox = new Container();
  private body: AnimatedSprite | Container | null = null;
  private readonly label: Text;
  private readonly atlas: AnimSet | null;
  private mood = '';
  private swords: (Sprite | Graphics)[] = [];
  private swordRealm = -1;
  private tweenFrom = 0;
  private tweenTo = 0;
  private tweenStart = 0;
  private bubble: Container | null = null;
  private bubbleUntil = 0;

  constructor(
    private readonly cfg: GameConfig,
    readonly id: number,
    name: string,
    readonly char: string,
    isMe: boolean,
  ) {
    this.atlas = assets.atlas(`characters/${char}`);
    this.label = new Text({
      text: name,
      style: { fontFamily: 'system-ui, sans-serif', fontSize: 15, fontWeight: '700', fill: isMe ? 0xfff3a0 : 0xffffff, stroke: { color: 0x000000, width: 4 } },
    });
    this.label.anchor.set(0.5, 1);
    this.root.addChild(this.shadow, this.aura, this.bodyBox, this.label);
    if (!this.atlas) {
      this.body = this.placeholderBody(name);
      this.bodyBox.addChild(this.body);
    }
  }

  private placeholderBody(name: string): Container {
    const [main, accent] = charColor(this.char);
    const c = new Container();
    const g = new Graphics().circle(0, 0, 1).fill(main).stroke({ color: 0x000000, width: 0.12 });
    g.circle(0, -0.25, 0.35).fill(accent);
    const t = new Text({ text: (name[0] ?? '?').toUpperCase(), style: { fontSize: 48, fontWeight: '900', fill: 0xffffff } });
    t.anchor.set(0.5);
    t.scale.set(0.02);
    c.addChild(g, t);
    return c;
  }

  update(f: PlayerFrame, zoom: number, now: number): void {
    const s = f.s;
    const r = this.cfg.player.body_r * s.sc;
    this.root.position.set(f.x, f.y);
    this.ring.position.set(f.x, f.y);
    this.root.zIndex = f.y;
    this.drawShadow(r);
    this.updateBody(s.md, r, Math.cos(s.f) < 0);
    this.drawAura(s.bf, r, now);
    this.applyAlpha(s.fl, now);
    this.label.position.set(0, -r * (CHAR_HEIGHT_PER_R - FEET_BELOW_CENTER) - 4 / zoom);
    this.label.scale.set(1 / zoom);
    this.updateRing(f, now);
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
    body.scale.set(flip ? -r : r, r);
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
  }

  // ---- sword ring ----------------------------------------------------------

  private updateRing(f: PlayerFrame, now: number): void {
    const s = f.s;
    const realm = realmOf(this.cfg, s.L);
    if (s.n !== this.swords.length || realm !== this.swordRealm) {
      this.rebuildSwords(s.n, realm, now);
    }
    const t = Math.min(1, (now - this.tweenStart) / this.cfg.sword.tween_secs);
    const fromStep = this.tweenFrom > 0 ? (2 * Math.PI) / this.tweenFrom : 0;
    const toStep = (2 * Math.PI) / Math.max(1, this.tweenTo);
    this.swords.forEach((sw, i) => {
      // Tween the spacing between swords when the count changes (spec §4).
      const step = fromStep > 0 ? fromStep + (toStep - fromStep) * t : toStep;
      const a = f.ringAngle + step * i;
      sw.position.set(Math.cos(a) * s.rr, Math.sin(a) * s.rr);
      sw.rotation = a + Math.PI / 2;
      this.sizeSword(sw, s.sl, s.sw);
    });
    this.ring.zIndex = f.y + 1;
  }

  private rebuildSwords(n: number, realm: number, now: number): void {
    this.tweenFrom = this.swords.length;
    this.tweenTo = n;
    this.tweenStart = now;
    if (realm !== this.swordRealm) {
      this.ring.removeChildren().forEach((c) => c.destroy());
      this.swords = [];
      this.swordRealm = realm;
    }
    while (this.swords.length > n) {
      this.swords.pop()?.destroy();
    }
    const tex = assets.texture(`swords/realm_${realm}.png`);
    while (this.swords.length < n) {
      const sw = tex ? new Sprite(tex) : new Graphics().roundRect(-0.5, -0.5, 1, 1, 0.2).fill(REALM_COLORS[realm] ?? 0xffffff).stroke({ color: 0, width: 0.08 });
      if (sw instanceof Sprite) {
        sw.anchor.set(0.5);
      }
      this.ring.addChild(sw);
      this.swords.push(sw);
    }
  }

  private sizeSword(sw: Sprite | Graphics, length: number, width: number): void {
    if (sw instanceof Sprite) {
      // Sword art fills ~90% of its square canvas vertically.
      const k = length / (sw.texture.height * 0.9);
      sw.scale.set(k);
      return;
    }
    sw.scale.set(width, length);
  }

  // ---- emote bubble --------------------------------------------------------

  showBubble(text: string, now: number): void {
    this.bubble?.destroy({ children: true });
    const c = new Container();
    const t = new Text({ text, style: { fontFamily: 'system-ui, sans-serif', fontSize: 16, fontWeight: '800', fill: 0x111111 } });
    t.anchor.set(0.5);
    const tex = assets.texture('ui/speech_bubble.png');
    const pad = 14;
    if (tex) {
      const bg = new Sprite(tex);
      bg.anchor.set(0.5);
      bg.width = t.width + pad * 3;
      bg.height = t.height + pad * 2.4;
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
    this.ring.destroy({ children: true });
  }
}
