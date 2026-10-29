import { Texture } from 'pixi.js';

import { assets } from './assets';

/**
 * Texture hạt dùng chung cho hào quang, thần thú và sự sống bản đồ (spec §3.6).
 *
 * Spec muốn dùng bộ hạt CC0 có sẵn (Kenney Particle Pack, Brackeys VFX Bundle) vì ảnh AI
 * sinh cho hạt nhỏ thường kém ổn định. Chưa có file thì tự vẽ bằng canvas — game vẫn chạy
 * đủ, chỉ là hạt đơn giản hơn.
 *
 * **Thả file vào `assets/vfx/particles/<tên>.png` là tự động dùng**, không phải sửa code.
 */

const baked = new Map<string, Texture>();

export type ParticleKind = 'spark' | 'smoke' | 'light' | 'star' | 'dot' | 'magic' | 'flare' | 'trace';

/** Vẽ tạm bằng canvas khi chưa có art. */
function bake(kind: ParticleKind): Texture {
  const n = 128;
  const cv = document.createElement('canvas');
  cv.width = n;
  cv.height = n;
  const c = cv.getContext('2d') as CanvasRenderingContext2D;
  const radial = (stop: number): CanvasGradient => {
    const g = c.createRadialGradient(n / 2, n / 2, n * stop, n / 2, n / 2, n / 2);
    g.addColorStop(0, 'rgba(255,255,255,1)');
    g.addColorStop(1, 'rgba(255,255,255,0)');
    return g;
  };
  switch (kind) {
    case 'spark':
      c.fillStyle = radial(0.02);
      c.fillRect(0, 0, n, n);
      c.globalCompositeOperation = 'lighter';
      c.fillStyle = '#ffffff';
      c.fillRect(n * 0.47, n * 0.1, n * 0.06, n * 0.8); // tia dọc
      c.fillRect(n * 0.1, n * 0.47, n * 0.8, n * 0.06); // tia ngang
      break;
    case 'smoke':
      c.fillStyle = radial(0.25);
      c.globalAlpha = 0.6;
      c.fillRect(0, 0, n, n);
      break;
    case 'star':
      c.fillStyle = '#ffffff';
      c.beginPath();
      for (let i = 0; i < 10; i++) {
        const a = (i / 10) * Math.PI * 2 - Math.PI / 2;
        const r = i % 2 === 0 ? n * 0.46 : n * 0.18;
        const x = n / 2 + Math.cos(a) * r;
        const y = n / 2 + Math.sin(a) * r;
        if (i === 0) {
          c.moveTo(x, y);
        } else {
          c.lineTo(x, y);
        }
      }
      c.closePath();
      c.fill();
      break;
    case 'dot':
      c.fillStyle = '#ffffff';
      c.beginPath();
      c.arc(n / 2, n / 2, n * 0.4, 0, Math.PI * 2);
      c.fill();
      break;
    case 'magic':
    case 'flare':
    case 'trace':
    default: // 'light'
      c.fillStyle = radial(0.1);
      c.fillRect(0, 0, n, n);
      break;
  }
  return Texture.from(cv);
}

/** particle trả texture hạt: ảnh thật nếu đã thả vào assets/vfx/particles, không thì tự vẽ. */
export function particle(kind: ParticleKind): Texture {
  const real = assets.texture(`vfx/particles/${kind}.png`);
  if (real) {
    return real;
  }
  const hit = baked.get(kind);
  if (hit) {
    return hit;
  }
  const tex = bake(kind);
  baked.set(kind, tex);
  return tex;
}
