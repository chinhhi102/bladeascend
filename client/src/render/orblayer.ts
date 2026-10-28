import { Particle, ParticleContainer, Texture } from 'pixi.js';

import type { ItemState } from '../net/protocol';
import { assets } from './assets';
import { ORB_COLORS, orbSize } from './entities';

/**
 * OrbLayer vẽ toàn bộ viên Linh khí bằng một ParticleContainer.
 *
 * EXP rải khắp bản đồ (spec §6.1) nên lúc camera lùi xa có **700–800 viên** cùng trên màn
 * hình. Mỗi viên trước đây là 3 node (Container + Graphics quầng sáng + Sprite) → hơn 2 000
 * node chỉ để vẽ EXP, kéo khung hình xuống dưới 40 fps. Gộp lại thành particle thì còn đúng
 * một lượt vẽ, và viên EXP vốn luôn nằm dưới mọi thứ khác nên cũng không cần y-sort.
 */

const BOB_PX = 4; // biên độ nhấp nhô, giống ItemSprite
const DIM_ALPHA = 0.3; // viên không hấp thụ được (spec §3.4) thì mờ đi

const baked = new Map<string, Texture>();

/** orbTexture: ảnh thật nếu đã vẽ, không thì tự vẽ một viên tròn có quầng sáng. */
function orbTexture(size: 'small' | 'medium' | 'large'): Texture {
  const real = assets.texture(`items/orb_${size}.png`);
  if (real) {
    return real;
  }
  const cached = baked.get(size);
  if (cached) {
    return cached;
  }
  const n = 64;
  const cv = document.createElement('canvas');
  cv.width = n;
  cv.height = n;
  const ctx = cv.getContext('2d') as CanvasRenderingContext2D;
  const hex = `#${ORB_COLORS[size].toString(16).padStart(6, '0')}`;
  const grad = ctx.createRadialGradient(n / 2, n / 2, n * 0.12, n / 2, n / 2, n / 2);
  grad.addColorStop(0, hex);
  grad.addColorStop(0.45, hex);
  grad.addColorStop(1, 'rgba(0,0,0,0)');
  ctx.fillStyle = grad;
  ctx.beginPath();
  ctx.arc(n / 2, n / 2, n / 2, 0, Math.PI * 2);
  ctx.fill();
  ctx.beginPath();
  ctx.arc(n / 2, n / 2, n * 0.28, 0, Math.PI * 2);
  ctx.fillStyle = '#ffffff';
  ctx.globalAlpha = 0.85;
  ctx.fill();
  const tex = Texture.from(cv);
  baked.set(size, tex);
  return tex;
}

/** orbScreenR: bán kính vẽ (gồm cả quầng), to hơn bán kính nhặt một chút cho dễ thấy. */
function orbScreenR(v: number): number {
  return v >= 60 ? 34 : v >= 20 ? 26 : 19;
}

interface Orb {
  p: Particle;
  phase: number;
  y: number;
}

export class OrbLayer {
  readonly root = new ParticleContainer({ dynamicProperties: { position: true, alpha: true } });
  private readonly orbs = new Map<number, Orb>();

  /** ids: danh sách viên đang vẽ (scene duyệt để dọn viên đã biến mất). */
  ids(): number[] {
    return [...this.orbs.keys()];
  }

  add(it: ItemState): void {
    if (this.orbs.has(it.id)) {
      return;
    }
    const size = orbSize(it.v ?? 0);
    const r = orbScreenR(it.v ?? 0);
    const tex = orbTexture(size);
    const p = new Particle({ texture: tex, anchorX: 0.5, anchorY: 0.5 });
    p.scaleX = (r * 2) / tex.width;
    p.scaleY = (r * 2) / tex.height;
    p.x = it.x;
    p.y = it.y;
    this.root.addParticle(p);
    this.orbs.set(it.id, { p, phase: Math.random() * Math.PI * 2, y: it.y });
  }

  /** at: vị trí hiện tại của viên, để làm hiệu ứng bay hút khi bị nhặt. */
  at(id: number): { x: number; y: number } | null {
    const o = this.orbs.get(id);
    return o ? { x: o.p.x, y: o.p.y } : null;
  }

  remove(id: number): void {
    const o = this.orbs.get(id);
    if (!o) {
      return;
    }
    this.root.removeParticle(o.p);
    this.orbs.delete(id);
  }

  /** update nhấp nhô tất cả viên; `absorbable = false` thì làm mờ (hai đường tu, §3.4). */
  update(now: number, absorbable: boolean): void {
    const alpha = absorbable ? 1 : DIM_ALPHA;
    for (const o of this.orbs.values()) {
      o.p.y = o.y - Math.sin(now * 3 + o.phase) * BOB_PX;
      o.p.alpha = alpha;
    }
  }

  clear(): void {
    this.root.removeParticles();
    this.orbs.clear();
  }
}
