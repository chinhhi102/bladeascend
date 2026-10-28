import { AnimatedSprite, Container, Graphics, Text } from 'pixi.js';

import { assets } from './assets';

interface Effect {
  node: Container;
  until: number;
  update?: (t: number) => void;
}

// Placeholder colours per effect id when the frames are missing.
const FALLBACK_COLOR: Record<string, number> = {
  clash_spark: 0xfff06a,
  sword_shatter: 0xd0d8e0,
  death_smoke: 0xeeeeee,
  levelup_aura: 0xffd23f,
  lightning: 0xfff36b,
  shield_break: 0xffd23f,
  stun_stars: 0xffe14d,
  run_dust: 0xc9a46a,
};

// Trần số hiệu ứng chạy cùng lúc. Một sự kiện dồn dập từ server (ví dụ bể kiếm hàng loạt)
// từng dựng ~1 800 AnimatedSprite mỗi giây và làm tụt khung hình; quá trần thì bỏ bớt hiệu
// ứng cũ nhất, thà mất một tia lửa còn hơn giật cả màn hình.
const MAX_EFFECTS = 60;

/** Vfx plays one-shot effects (ART_PROMPTS §10) and floating texts; also owns camera shake. */
export class Vfx {
  readonly layer = new Container();
  private effects: Effect[] = [];
  shakeUntil = 0;
  shakeAmp = 0;

  /** play shows effect `id` at (x, y) with a world-space size. */
  play(id: string, x: number, y: number, size: number, now: number): void {
    const frames = assets.frames(`vfx/${id}`, id);
    if (frames.length) {
      const sp = new AnimatedSprite(frames);
      sp.anchor.set(0.5);
      sp.width = sp.height = size;
      sp.loop = false;
      sp.animationSpeed = 0.35;
      sp.position.set(x, y);
      sp.play();
      this.add({ node: sp, until: now + frames.length / 21 });
      return;
    }
    this.burst(x, y, size, FALLBACK_COLOR[id] ?? 0xffffff, now);
  }

  /** burst: simple expanding ring + sparks placeholder. */
  private burst(x: number, y: number, size: number, color: number, now: number): void {
    const g = new Graphics();
    g.position.set(x, y);
    const start = now;
    const dur = 0.35;
    this.add({
      node: g,
      until: now + dur,
      update: (t) => {
        const k = Math.min(1, (t - start) / dur);
        g.clear();
        g.circle(0, 0, (size / 2) * (0.3 + 0.7 * k)).stroke({ color, width: size * 0.06, alpha: 1 - k });
        for (let i = 0; i < 8; i++) {
          const a = (i / 8) * Math.PI * 2;
          const r = (size / 2) * (0.4 + 0.8 * k);
          g.circle(Math.cos(a) * r, Math.sin(a) * r, size * 0.04 * (1 - k)).fill(color);
        }
      },
    });
  }

  /** floatText shows "+20" style numbers rising and fading; size is screen px via 1/zoom. */
  floatText(text: string, x: number, y: number, color: number, zoom: number, now: number): void {
    const t = new Text({ text, style: { fontFamily: 'system-ui', fontSize: 20, fontWeight: '900', fill: color, stroke: { color: 0, width: 4 } } });
    t.anchor.set(0.5);
    t.scale.set(1 / zoom);
    t.position.set(x, y);
    const start = now;
    this.add({
      node: t,
      until: now + 1,
      update: (tt) => {
        const k = tt - start;
        t.y = y - k * 60 / zoom;
        t.alpha = 1 - k;
      },
    });
  }

  shake(amp: number, secs: number, now: number): void {
    this.shakeAmp = Math.max(this.shakeAmp, amp);
    this.shakeUntil = Math.max(this.shakeUntil, now + secs);
  }

  shakeOffset(now: number): { x: number; y: number } {
    if (now > this.shakeUntil) {
      this.shakeAmp = 0;
      return { x: 0, y: 0 };
    }
    return { x: (Math.random() - 0.5) * this.shakeAmp, y: (Math.random() - 0.5) * this.shakeAmp };
  }

  private add(e: Effect): void {
    while (this.effects.length >= MAX_EFFECTS) {
      const old = this.effects.shift();
      old?.node.destroy({ children: true });
    }
    this.layer.addChild(e.node);
    this.effects.push(e);
  }

  update(now: number): void {
    this.effects = this.effects.filter((e) => {
      if (now >= e.until) {
        e.node.destroy({ children: true });
        return false;
      }
      e.update?.(now);
      return true;
    });
  }
}
