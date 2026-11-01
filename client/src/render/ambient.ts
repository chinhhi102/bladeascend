import { Container, Sprite, Texture } from 'pixi.js';

import { particle } from './particles';
import { emergency, qualityFor } from './quality';

/**
 * Sự sống của bản đồ (spec §8.6 mục 5): bóng mây trôi, lá tre rơi, đom đóm, sương mù, bướm,
 * chim bay ngang. **Chỉ ở client, không ảnh hưởng gameplay** — không va chạm, không gửi lên
 * server, không có trong snapshot.
 *
 * Mục đích là làm bản đồ bớt "chết": một màn hình toàn mặt đất tĩnh thì người chơi cảm thấy
 * đang đứng trong file ảnh, chứ không phải trong một thế giới.
 *
 * Tất cả hạt sống trong một hệ toạ độ **theo camera**: chúng bám quanh người chơi chứ không
 * rải khắp bản đồ 12 000×12 000, nên số hạt không đổi dù bản đồ to cỡ nào.
 */

/** Trần số hạt theo chất lượng (spec §8.6: Cao 150 · Vừa 60 · Thấp 0). */
const BUDGET: Record<string, number> = { high: 150, medium: 60, low: 0 };

type Kind = 'cloud' | 'leaf' | 'firefly' | 'mist' | 'butterfly' | 'bird';

/** Hạt nào hợp biome nào. Bóng mây và chim thì bản đồ nào cũng có. */
const BY_BIOME: Record<string, Kind[]> = {
  lang_tre: ['cloud', 'leaf', 'firefly', 'bird'],
  cho_dem: ['cloud', 'butterfly', 'bird'],
  nui_da: ['cloud', 'mist', 'bird'],
};
const FALLBACK: Kind[] = ['cloud', 'bird'];

interface Mote {
  node: Sprite;
  /** Độ đậm gốc của loại này, để hiệu ứng mờ dần nhân vào chứ không ghi đè. */
  baseAlpha: number;
  kind: Kind;
  vx: number;
  vy: number;
  spin: number;
  life: number;
  born: number;
}

/**
 * Mỗi loại hạt nướng sẵn MỘT texture rồi dùng chung cho cả trăm hạt.
 *
 * Không dùng `Graphics` riêng từng hạt: mỗi Graphics có hình riêng là một lượt vẽ riêng — đo
 * thật, 150 hạt kiểu đó đẩy draw call từ 237 lên 549 và khung hình tụt từ 58 xuống 36.
 * Sprite chung texture thì Pixi gộp hết vào một lô.
 */
const tex = new Map<Kind, Texture>();

function textureFor(kind: Kind): Texture {
  const hit = tex.get(kind);
  if (hit) {
    return hit;
  }
  // Mây, sương và đom đóm dùng được luôn bộ hạt CC0 (spec §3.6); lá, bướm, chim thì phải
  // đúng hình nên vẫn tự vẽ.
  const shared: Partial<Record<Kind, 'smoke' | 'light'>> = { cloud: 'smoke', mist: 'smoke', firefly: 'light' };
  const want = shared[kind];
  if (want) {
    const t = particle(want);
    tex.set(kind, t);
    return t;
  }
  const n = 128;
  const cv = document.createElement('canvas');
  cv.width = n;
  cv.height = n;
  const c = cv.getContext('2d') as CanvasRenderingContext2D;
  const soft = (color: string, inner: number): void => {
    const g = c.createRadialGradient(n / 2, n / 2, n * inner, n / 2, n / 2, n / 2);
    g.addColorStop(0, color);
    g.addColorStop(1, 'rgba(0,0,0,0)');
    c.fillStyle = g;
    c.fillRect(0, 0, n, n);
  };
  switch (kind) {
    case 'cloud':
      soft('rgba(0,0,0,0.5)', 0.1);
      break;
    case 'mist':
      soft('rgba(223,232,238,0.55)', 0.05);
      break;
    case 'firefly':
      soft('rgba(255,243,160,1)', 0.12);
      break;
    case 'leaf':
      c.fillStyle = '#9ab06a';
      c.beginPath();
      c.ellipse(n / 2, n / 2, n * 0.42, n * 0.16, 0, 0, Math.PI * 2);
      c.fill();
      break;
    case 'butterfly':
      c.fillStyle = '#f4a7c0';
      for (const dx of [-0.16, 0.16]) {
        c.beginPath();
        c.ellipse(n / 2 + n * dx, n / 2, n * 0.17, n * 0.26, 0, 0, Math.PI * 2);
        c.fill();
      }
      break;
    default: // 'bird' — chỉ là bóng chim dưới đất
      c.fillStyle = 'rgba(0,0,0,0.75)';
      c.beginPath();
      c.ellipse(n / 2, n / 2, n * 0.45, n * 0.16, 0, 0, Math.PI * 2);
      c.fill();
      break;
  }
  const t = Texture.from(cv);
  tex.set(kind, t);
  return t;
}

/** Cỡ vẽ và độ đậm của từng loại, tính bằng đơn vị thế giới. */
const LOOK: Record<Kind, { w: number; h: number; alpha: number; add?: boolean }> = {
  cloud: { w: 900, h: 520, alpha: 0.1 },
  mist: { w: 640, h: 200, alpha: 0.12 },
  firefly: { w: 26, h: 26, alpha: 0.95, add: true },
  leaf: { w: 22, h: 10, alpha: 0.85 },
  butterfly: { w: 20, h: 18, alpha: 0.9 },
  bird: { w: 34, h: 13, alpha: 0.18 },
};

function draw(kind: Kind): Sprite {
  const look = LOOK[kind];
  const sp = new Sprite(textureFor(kind));
  sp.anchor.set(0.5);
  sp.width = look.w;
  sp.height = look.h;
  sp.alpha = look.alpha;
  if (look.add) {
    sp.blendMode = 'add';
  }
  return sp;
}

/** Thời gian sống và tốc độ của từng loại. */
function spawnSpec(kind: Kind): { life: number; vx: number; vy: number; spin: number } {
  const r = (a: number, b: number): number => a + Math.random() * (b - a);
  switch (kind) {
    case 'cloud':
      return { life: r(26, 42), vx: r(14, 30), vy: r(-5, 5), spin: 0 };
    case 'leaf':
      return { life: r(7, 12), vx: r(-40, -12), vy: r(26, 58), spin: r(-1.6, 1.6) };
    case 'firefly':
      return { life: r(5, 10), vx: r(-18, 18), vy: r(-18, 18), spin: 0 };
    case 'mist':
      return { life: r(20, 34), vx: r(10, 26), vy: r(-4, 4), spin: 0 };
    case 'butterfly':
      return { life: r(6, 11), vx: r(-45, 45), vy: r(-30, 30), spin: r(-2, 2) };
    default:
      return { life: r(4, 7), vx: r(180, 320), vy: r(-30, 30), spin: 0 };
  }
}

export class Ambient {
  readonly root = new Container();
  private readonly motes: Mote[] = [];
  private kinds: Kind[] = FALLBACK;
  private last = 0;

  setBiome(mapId: string): void {
    this.kinds = BY_BIOME[mapId] ?? FALLBACK;
    this.clear();
  }

  /**
   * update giữ số hạt quanh người chơi. `view` là nửa chiều rộng/cao khung nhìn, để sinh
   * hạt ngay ngoài mép màn hình rồi cho nó trôi ngang qua.
   */
  update(cx: number, cy: number, viewW: number, viewH: number, now: number): void {
    // Bước khẩn (spec §13.6 mục 8): tắt hẳn ambient.
    const budget = emergency() ? 0 : BUDGET[qualityFor('ambient')] ?? 0;
    if (budget === 0) {
      this.clear();
      return;
    }
    const dt = this.last > 0 ? Math.min(0.05, now - this.last) : 0;
    this.last = now;

    for (let i = this.motes.length - 1; i >= 0; i--) {
      const m = this.motes[i];
      const age = now - m.born;
      const off = Math.abs(m.node.x - cx) > viewW * 1.4 || Math.abs(m.node.y - cy) > viewH * 1.4;
      if (age >= m.life || off) {
        m.node.destroy();
        this.motes.splice(i, 1);
        continue;
      }
      m.node.x += m.vx * dt;
      m.node.y += m.vy * dt;
      if (m.spin) {
        m.node.rotation += m.spin * dt;
      }
      if (m.kind === 'firefly') {
        // Đom đóm chớp tắt, bay lượn chứ không đi thẳng.
        m.node.alpha = m.baseAlpha * (0.35 + 0.65 * Math.abs(Math.sin(now * 2.4 + m.born)));
        m.vx += Math.sin(now * 1.7 + m.born) * 14 * dt;
        m.vy += Math.cos(now * 1.3 + m.born) * 14 * dt;
      } else if (m.kind === 'butterfly') {
        m.node.y += Math.sin(now * 9 + m.born) * 24 * dt; // vỗ cánh nhấp nhô
      } else {
        // Vào và ra đều mờ dần để không "bật" ra giữa màn hình.
        const t = age / m.life;
        m.node.alpha = m.baseAlpha * Math.min(1, Math.min(t, 1 - t) * 6);
      }
    }

    while (this.motes.length < budget) {
      this.spawn(cx, cy, viewW, viewH, now);
    }
  }

  private spawn(cx: number, cy: number, viewW: number, viewH: number, now: number): void {
    const kind = this.kinds[Math.floor(Math.random() * this.kinds.length)];
    // Mây và sương thì ít mà to; hạt nhỏ thì nhiều. Bỏ qua nếu đã đủ loại lớn.
    if ((kind === 'cloud' || kind === 'mist') && this.motes.filter((m) => m.kind === kind).length >= 4) {
      return this.spawnSmall(cx, cy, viewW, viewH, now);
    }
    this.add(kind, cx, cy, viewW, viewH, now);
  }

  private spawnSmall(cx: number, cy: number, viewW: number, viewH: number, now: number): void {
    const small = this.kinds.filter((k) => k !== 'cloud' && k !== 'mist');
    this.add(small.length ? small[Math.floor(Math.random() * small.length)] : 'bird', cx, cy, viewW, viewH, now);
  }

  private add(kind: Kind, cx: number, cy: number, viewW: number, viewH: number, now: number): void {
    const spec = spawnSpec(kind);
    const node = draw(kind);
    // Sinh ở rìa khung nhìn, phía ngược hướng bay, để nó trôi ngang qua màn hình.
    node.x = cx + (spec.vx >= 0 ? -viewW * 1.2 : viewW * 1.2) + (Math.random() - 0.5) * viewW;
    node.y = cy + (Math.random() - 0.5) * viewH * 2.2;
    node.alpha = 0;
    this.root.addChild(node);
    this.motes.push({ node, baseAlpha: LOOK[kind].alpha, kind, vx: spec.vx, vy: spec.vy, spin: spec.spin, life: spec.life, born: now });
  }

  clear(): void {
    for (const m of this.motes) {
      m.node.destroy();
    }
    this.motes.length = 0;
  }
}
