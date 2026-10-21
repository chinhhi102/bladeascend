import { AnimatedSprite, Container, Graphics, Sprite, Text, type Texture } from 'pixi.js';

import type { ItemState, MonsterState } from '../net/protocol';
import { assets } from './assets';
import { BUFF_COLORS, MONSTER_COLORS, MONSTER_VARIANTS } from './palette';

const MONSTER_HEIGHT_PER_R = 2.4;
const ITEM_SIZE_PER_R = 2.6;
const ORB_COLORS = { small: 0x5dff7a, medium: 0x3a86ff, large: 0xffd23f };

function variantOf(kind: string, id: number): string {
  const list = MONSTER_VARIANTS[kind] ?? ['banh_bao_cau'];
  return list[id % list.length];
}

/** MonsterSprite: animated monster (or placeholder blob) with its HP printed on the belly. */
export class MonsterSprite {
  readonly root = new Container();
  private readonly body: Container;
  private readonly hpText: Text;
  private hurtUntil = 0;
  private lastHp = -1;

  constructor(s: MonsterState) {
    const anims = assets.atlas(`monsters/${variantOf(s.k, s.id)}`);
    const idle = anims?.anims.get('idle');
    if (anims && idle) {
      const sp = new AnimatedSprite(idle.textures);
      sp.anchor.set(anims.pivotX, anims.pivotY);
      sp.animationSpeed = idle.fps / 60;
      sp.play();
      const k = (s.r * MONSTER_HEIGHT_PER_R) / 0.72 / sp.texture.height;
      sp.scale.set(k);
      sp.position.set(0, s.r * 0.8);
      this.body = sp;
    } else {
      this.body = new Graphics().circle(0, 0, s.r).fill(MONSTER_COLORS[s.k] ?? 0xffffff).stroke({ color: 0, width: s.r * 0.12 });
    }
    this.hpText = new Text({ text: '', style: { fontFamily: 'system-ui', fontSize: 40, fontWeight: '900', fill: 0x222222, stroke: { color: 0xffffff, width: 6 } } });
    this.hpText.anchor.set(0.5);
    this.hpText.scale.set((s.r * 0.9) / 40);
    this.root.addChild(new Graphics().ellipse(0, s.r * 0.8, s.r, s.r * 0.35).fill({ color: 0, alpha: 0.2 }), this.body, this.hpText);
  }

  update(x: number, y: number, s: MonsterState, now: number): void {
    this.root.position.set(x, y);
    this.root.zIndex = y;
    if (s.hp !== this.lastHp) {
      if (this.lastHp > s.hp) {
        this.hurtUntil = now + 0.2;
      }
      this.lastHp = s.hp;
      this.hpText.text = String(s.hp);
    }
    const hurt = now < this.hurtUntil;
    this.body.tint = hurt ? 0xff8080 : 0xffffff;
    this.body.scale.y = Math.abs(this.body.scale.y) * (hurt ? 0.97 : 1);
  }

  destroy(): void {
    this.root.destroy({ children: true });
  }
}

function itemTexture(it: ItemState): Texture | null {
  if (it.k === 'orb') {
    const size = (it.v ?? 0) >= 60 ? 'large' : (it.v ?? 0) >= 20 ? 'medium' : 'small';
    return assets.texture(`items/orb_${size}.png`);
  }
  if (it.k === 'sword') {
    return assets.texture('items/sword.png');
  }
  if (it.k === 'chest') {
    return assets.frames('items/chest', 'idle')[0] ?? null;
  }
  return assets.texture(`buffs/${it.k}.png`);
}

function orbRadius(v: number): number {
  return v >= 60 ? 22 : v >= 20 ? 17 : 12;
}

/** ItemSprite: bobbing bubble/orb/chest. */
export class ItemSprite {
  readonly root = new Container();
  private readonly phase = Math.random() * Math.PI * 2;

  constructor(readonly it: ItemState, itemR: number) {
    const r = it.k === 'orb' ? orbRadius(it.v ?? 0) : it.k === 'chest' ? itemR * 2.2 : itemR;
    const chestFrames = it.k === 'chest' ? assets.frames('items/chest', 'idle') : [];
    const tex = itemTexture(it);
    let node: Container;
    if (chestFrames.length > 1) {
      const sp = new AnimatedSprite(chestFrames);
      sp.anchor.set(0.5);
      sp.animationSpeed = 0.1;
      sp.play();
      sp.width = sp.height = r * ITEM_SIZE_PER_R;
      node = sp;
    } else if (tex) {
      const sp = new Sprite(tex);
      sp.anchor.set(0.5);
      sp.width = sp.height = r * ITEM_SIZE_PER_R;
      node = sp;
    } else {
      node = this.placeholder(it, r);
    }
    this.root.addChild(node);
    this.root.position.set(it.x, it.y);
    this.root.zIndex = it.y - 1000;
  }

  private placeholder(it: ItemState, r: number): Graphics {
    const g = new Graphics();
    if (it.k === 'orb') {
      const size = (it.v ?? 0) >= 60 ? 'large' : (it.v ?? 0) >= 20 ? 'medium' : 'small';
      return g.circle(0, 0, r).fill(ORB_COLORS[size]).stroke({ color: 0xffffff, width: 3 });
    }
    if (it.k === 'chest') {
      return g.roundRect(-r, -r * 0.7, r * 2, r * 1.4, 8).fill(0x9c6644).stroke({ color: 0xffd23f, width: 6 });
    }
    const color = it.k === 'sword' ? 0xffffff : BUFF_COLORS[it.k] ?? 0xcccccc;
    g.circle(0, 0, r).fill({ color, alpha: 0.35 }).stroke({ color, width: 4 });
    if (it.k === 'sword') {
      g.rect(-3, -r * 0.7, 6, r * 1.4).fill(0xc8a165);
    }
    return g;
  }

  update(now: number): void {
    this.root.pivot.y = Math.sin(now * 3 + this.phase) * 4;
  }

  destroy(): void {
    this.root.destroy({ children: true });
  }
}
