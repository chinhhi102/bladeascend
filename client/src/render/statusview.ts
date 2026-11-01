import { Container, Graphics, Sprite } from 'pixi.js';

import { type GameConfig, skillU } from '../config';
import { STATUS, statusOn, type StatusEv } from '../net/protocol';
import { quality } from './quality';
import { type StatusIconKey, statusIconTexture } from './statusicons';
import { statusTracker } from './statustrack';

/**
 * StatusView: lớp ③ và ④ của spec v6.36 §6.5 trên MỘT nhân vật — dấu hiệu suốt thời gian
 * hiệu lực (zzz, sao choáng, hào quang gồng, tia chớp, mai rùa, võng…), icon trạng thái có
 * vòng đếm ngược trên đầu, và vòng đếm dưới chân cho người đang có buff kỹ năng.
 *
 * Bật/tắt theo bit `st` trong snapshot (nguồn sự thật luôn có); số giây lấy từ event
 * `StatusStart` qua `statusTracker`. Người lọt vào khung hình giữa chừng thì không có đồng
 * hồ — vẫn thấy icon, chỉ thiếu vòng đếm.
 *
 * Toạ độ trong `root` của PlayerSprite là ĐƠN VỊ THẾ GIỚI (gốc = tâm thân); chữ/icon chia
 * cho `zoom` để giữ nguyên cỡ trên màn hình.
 */

// Tỉ lệ vẽ (không phải số cân bằng).
const FEET_Y = 0.75;
const ICON_PX = 24;
const ICON_GAP_PX = 28;
const ICON_ABOVE_HEAD_PX = 54;
const BAR_W_PX = 46;
const BAR_H_PX = 6;
const TRAIL_POINTS = 8;
const TRAIL_STEP_SECS = 0.03;

const COLORS = {
  brace: 0xff8c2a,
  vein: 0xd9342b,
  ghost: 0xffd23f,
  shell: 0x7ad7b0,
  steal: 0x9b59b6,
  minion: 0x4a235a,
  sleep: 0xbcd4ff,
  star: 0xffe14d,
  dust: 0xb8a58a,
  rope: 0x7a5230,
} as const;

/** Thứ tự icon cố định: người xem quen vị trí, icon không nhảy chỗ khi một cái hết. */
const ICON_ORDER: readonly StatusIconKey[] = ['sleep', 'stun', 'blind', 'push', 'brace', 'ghost', 'shell', 'steal'];
const SELF_BUFFS: readonly StatusEv[] = ['brace', 'ghost', 'shell', 'steal'];

export interface StatusInput {
  id: number;
  st: number;
  /** r: bán kính thân (đơn vị thế giới); sc: scale (cho vòng nhặt). */
  r: number;
  sc: number;
  headY: number;
  zoom: number;
  now: number;
  /** vx, vy: vận tốc vẽ (đơn vị/giây) — hướng vệt bụi / vệt trượt / tia chớp. */
  vx: number;
  vy: number;
  /** wx, wy: vị trí thế giới (cho vệt tia chớp — điểm cũ tính tương đối với chỗ đang đứng). */
  wx: number;
  wy: number;
}

interface IconNode {
  key: StatusIconKey;
  node: Container;
  sprite: Sprite;
  ring: Graphics;
}

export class StatusView {
  /** under: dưới thân (hào quang, võng, vệt). over: trên thân (zzz, sao, mai, icon). */
  readonly under = new Graphics();
  readonly over = new Container();
  private readonly fx = new Graphics();
  private readonly iconRow = new Container();
  private readonly bar = new Graphics();
  private icons: IconNode[] = [];
  private iconKey = '';
  private readonly trail: { x: number; y: number }[] = [];
  private trailAt = 0;
  /** halo: màu viền vòng kiếm khung này (0 = không). bodyScale: Gồng phồng 10%. */
  halo = 0;
  bodyScale = 1;
  /** drawn: khung trước có vẽ gì không (để biết có cần xoá khi hết trạng thái). */
  private drawn = true;

  constructor(private readonly cfg: GameConfig) {
    this.over.addChild(this.fx, this.iconRow, this.bar);
  }

  update(s: StatusInput): void {
    // Đường nóng: 60 nhân vật × 60 khung/giây, gần như ai cũng KHÔNG có trạng thái nào. Không
    // có gì để vẽ và khung trước cũng không vẽ gì thì khỏi `clear()` — xoá một Graphics là
    // đánh dấu dựng lại hình học, kể cả khi nó vốn rỗng.
    const idle = s.st === 0 && !statusTracker.has(s.id);
    if (idle && !this.drawn) {
      return;
    }
    this.drawn = !idle;
    const on = (k: StatusEv): boolean => this.isOn(s, k);
    const low = quality() === 'low';
    this.under.clear();
    this.fx.clear();
    this.drawUnder(s, on, low);
    this.drawOver(s, on, low);
    this.drawIcons(s, on);
    this.drawMinionBar(s, on('minion'));
    this.halo = on('brace') ? COLORS.brace : on('steal') ? COLORS.steal : on('minion') ? COLORS.minion : 0;
    this.bodyScale = on('brace') ? 1.1 : 1;
  }

  reset(): void {
    this.drawn = true;
    this.trail.length = 0;
    this.iconKey = '';
    this.iconRow.removeChildren().forEach((c) => c.destroy({ children: true }));
    this.icons = [];
  }

  private isOn(s: StatusInput, k: StatusEv): boolean {
    if ((STATUS as readonly string[]).includes(k)) {
      return statusOn(s.st, k as (typeof STATUS)[number]);
    }
    return statusTracker.active(s.id, k, s.now);
  }

  // ── dưới thân ──────────────────────────────────────────────────────────────────────

  private drawUnder(s: StatusInput, on: (k: StatusEv) => boolean, low: boolean): void {
    const g = this.under;
    const { r, now } = s;
    if (on('steal')) {
      this.drawHammock(r, now);
    }
    if (on('brace')) {
      const p = 1 + 0.05 * Math.sin(now * 9);
      g.circle(0, r * 0.1, r * 1.35 * p).fill({ color: COLORS.brace, alpha: 0.2 }).stroke({ color: COLORS.brace, width: r * 0.1, alpha: 0.85 });
    }
    if (on('ghost')) {
      this.drawGhost(s, low);
    } else {
      this.trail.length = 0;
    }
    if (!low && (on('push') || on('slip'))) {
      this.drawDust(s, on('slip'));
    }
    this.drawFootTimers(s, on);
  }

  private drawHammock(r: number, now: number): void {
    const g = this.under;
    const sway = Math.sin(now * 3) * r * 0.18;
    const y0 = -r * 0.1;
    g.moveTo(-r * 1.5, y0 - r * 0.6).lineTo(-r * 1.3, y0).stroke({ color: COLORS.rope, width: r * 0.06 });
    g.moveTo(r * 1.5, y0 - r * 0.6).lineTo(r * 1.3, y0).stroke({ color: COLORS.rope, width: r * 0.06 });
    g.moveTo(-r * 1.3, y0).quadraticCurveTo(sway, r * 1.1, r * 1.3, y0)
      .stroke({ color: COLORS.rope, width: r * 0.22, alpha: 0.95 });
    g.moveTo(-r * 1.2, y0 + r * 0.05).quadraticCurveTo(sway, r * 0.95, r * 1.2, y0 + r * 0.05)
      .stroke({ color: COLORS.steal, width: r * 0.08, alpha: 0.8 });
  }

  /** Vệt tia chớp vàng sau lưng Shipper + vòng nhặt mờ (+bonus Gom Đơn). */
  private drawGhost(s: StatusInput, low: boolean): void {
    const g = this.under;
    const u = skillU(this.cfg, s.sc) * (1 + (this.cfg.skills.gom_don_pickup ?? 0));
    g.circle(0, 0, u).stroke({ color: COLORS.ghost, width: Math.max(2, s.r * 0.05), alpha: 0.35 });
    if (s.now - this.trailAt >= TRAIL_STEP_SECS) {
      this.trailAt = s.now;
      this.trail.unshift({ x: s.wx, y: s.wy });
      this.trail.length = Math.min(this.trail.length, TRAIL_POINTS);
    }
    if (low || this.trail.length < 2) {
      return;
    }
    // Gấp khúc: lệch trái/phải xen kẽ vuông góc với đoạn — ra dáng tia chớp, không phải vệt mờ.
    g.moveTo(0, 0);
    for (let i = 1; i < this.trail.length; i++) {
      const p = this.trail[i];
      const q = this.trail[i - 1];
      const dx = p.x - q.x;
      const dy = p.y - q.y;
      const len = Math.hypot(dx, dy) || 1;
      const k = (i % 2 === 0 ? 1 : -1) * s.r * 0.35;
      g.lineTo(p.x - s.wx - (dy / len) * k, p.y - s.wy + (dx / len) * k);
    }
    g.stroke({ color: COLORS.ghost, width: s.r * 0.22, alpha: 0.85, join: 'miter' });
  }

  private drawDust(s: StatusInput, slip: boolean): void {
    const v = Math.hypot(s.vx, s.vy);
    if (v < 1) {
      return;
    }
    const bx = -s.vx / v;
    const by = -s.vy / v;
    const g = this.under;
    for (let i = 1; i <= 4; i++) {
      const d = s.r * 0.55 * i;
      const x = bx * d;
      const y = s.r * FEET_Y + by * d;
      if (slip) {
        g.moveTo(x - by * s.r * 0.3, y + bx * s.r * 0.3).lineTo(x + by * s.r * 0.3, y - bx * s.r * 0.3);
        continue;
      }
      g.circle(x, y, s.r * (0.18 + 0.07 * i)).fill({ color: COLORS.dust, alpha: 0.5 - i * 0.1 });
    }
    if (slip) {
      g.stroke({ color: 0xffffff, width: s.r * 0.06, alpha: 0.6 });
    }
  }

  /** ④ cho người dùng có buff: vòng elip dưới chân, rút ngắn theo thời gian còn lại. */
  private drawFootTimers(s: StatusInput, on: (k: StatusEv) => boolean): void {
    let ring = 0;
    for (const k of SELF_BUFFS) {
      const left = on(k) ? statusTracker.left(s.id, k, s.now) : null;
      if (!left) {
        continue;
      }
      const R = s.r * (1.25 + ring * 0.2);
      const color = COLORS[k as keyof typeof COLORS] ?? 0xffffff;
      const segs = Math.max(2, Math.ceil(40 * left.frac));
      const a0 = -Math.PI / 2;
      const span = Math.PI * 2 * left.frac;
      for (let i = 0; i <= segs; i++) {
        const a = a0 + (span * i) / segs;
        const x = Math.cos(a) * R;
        const y = s.r * FEET_Y + Math.sin(a) * R * 0.38;
        if (i === 0) {
          this.under.moveTo(x, y);
        } else {
          this.under.lineTo(x, y);
        }
      }
      this.under.stroke({ color, width: s.r * 0.09, alpha: 0.9, cap: 'round' });
      ring++;
    }
  }

  // ── trên thân ──────────────────────────────────────────────────────────────────────

  private drawOver(s: StatusInput, on: (k: StatusEv) => boolean, low: boolean): void {
    const { r, now } = s;
    const g = this.fx;
    if (on('shell')) {
      const glow = 0.25 + 0.12 * Math.sin(now * 5);
      g.ellipse(0, -r * 0.55, r * 1.25, r * 1.35).fill({ color: COLORS.shell, alpha: glow })
        .stroke({ color: COLORS.shell, width: r * 0.08, alpha: 0.9 });
      g.poly(hexagon(0, -r * 0.7, r * 0.45)).stroke({ color: 0x1e7a5a, width: r * 0.05, alpha: 0.7 });
    }
    if (on('sleep')) {
      this.drawSleep(r, now, low);
    }
    if (on('stun') || on('slip')) {
      for (let i = 0; i < 3; i++) {
        const a = now * 5 + (i * Math.PI * 2) / 3;
        g.poly(star(Math.cos(a) * r * 0.65, -r * 1.95 + Math.sin(a) * r * 0.2, r * 0.18)).fill(COLORS.star).stroke({ color: 0x5a3b00, width: r * 0.03 });
      }
    }
    if (on('brace') && !low) {
      // Gân đỏ nổi hai bên tay — nhấp nháy theo nhịp gồng.
      const a = 0.5 + 0.5 * Math.sin(now * 12);
      for (const sx of [-1, 1]) {
        g.moveTo(sx * r * 0.55, -r * 0.9).lineTo(sx * r * 0.75, -r * 0.7).lineTo(sx * r * 0.6, -r * 0.5).lineTo(sx * r * 0.8, -r * 0.3);
      }
      g.stroke({ color: COLORS.vein, width: r * 0.07, alpha: a });
    }
    if (on('sing') && !low) {
      for (let i = 0; i < 3; i++) {
        const t = (now * 0.9 + i / 3) % 1;
        note(g, (i - 1) * r * 0.8 + Math.sin(t * 6) * r * 0.2, -r * (1.6 + t * 1.2), r * 0.18, COLORS.steal, 1 - t);
      }
    }
  }

  private drawSleep(r: number, now: number, low: boolean): void {
    const g = this.fx;
    // Bong bóng mũi phập phồng.
    const b = r * (0.14 + 0.12 * (0.5 + 0.5 * Math.sin(now * 3)));
    g.circle(r * 0.35, -r * 1.25, b).fill({ color: COLORS.sleep, alpha: 0.55 }).stroke({ color: 0xffffff, width: r * 0.03, alpha: 0.9 });
    if (low) {
      return;
    }
    for (let i = 0; i < 3; i++) {
      const t = (now * 0.6 + i / 3) % 1;
      const z = r * (0.14 + t * 0.12);
      const x = r * (0.6 + t * 0.6);
      const y = -r * (1.5 + t * 1.1);
      g.moveTo(x - z, y - z).lineTo(x + z, y - z).lineTo(x - z, y + z).lineTo(x + z, y + z)
        .stroke({ color: 0xffffff, width: r * 0.06, alpha: 1 - t });
    }
  }

  // ── icon + thanh đếm ───────────────────────────────────────────────────────────────

  private drawIcons(s: StatusInput, on: (k: StatusEv) => boolean): void {
    const keys = ICON_ORDER.filter((k) => on(k) || (k === 'stun' && on('slip')));
    const key = keys.join(',');
    if (key !== this.iconKey) {
      this.iconKey = key;
      this.rebuildIcons(keys);
    }
    this.iconRow.visible = keys.length > 0;
    if (!this.iconRow.visible) {
      return;
    }
    const size = ICON_PX / s.zoom;
    const gap = ICON_GAP_PX / s.zoom;
    const y = s.headY - ICON_ABOVE_HEAD_PX / s.zoom;
    this.icons.forEach((ic, i) => {
      ic.node.position.set((i - (this.icons.length - 1) / 2) * gap, y);
      const tex = statusIconTexture(ic.key);
      if (tex && ic.sprite.texture !== tex) {
        ic.sprite.texture = tex;
      }
      ic.sprite.visible = tex !== null;
      ic.sprite.width = size;
      ic.sprite.height = size;
      const left = this.leftFor(s, ic.key);
      ic.ring.clear();
      if (left !== null) {
        const R = size * 0.62;
        ic.ring.arc(0, 0, R, -Math.PI / 2, -Math.PI / 2 + Math.PI * 2 * left).stroke({ color: 0xffffff, width: size * 0.12, cap: 'round' });
      }
    });
  }

  /** leftFor: tỉ lệ còn lại của icon (choáng/trượt lấy cái còn lâu hơn); null = không có đồng hồ. */
  private leftFor(s: StatusInput, key: StatusIconKey): number | null {
    const codes: StatusEv[] = key === 'stun' ? ['stun', 'slip'] : [key];
    let best: number | null = null;
    for (const c of codes) {
      const l = statusTracker.left(s.id, c, s.now);
      if (l && (best === null || l.frac > best)) {
        best = l.frac;
      }
    }
    return best;
  }

  private rebuildIcons(keys: StatusIconKey[]): void {
    this.iconRow.removeChildren().forEach((c) => c.destroy({ children: true }));
    this.icons = keys.map((key) => {
      const node = new Container();
      const sprite = new Sprite();
      sprite.anchor.set(0.5);
      const ring = new Graphics();
      node.addChild(sprite, ring);
      this.iconRow.addChild(node);
      return { key, node, sprite, ring };
    });
  }

  /** Đàn em: thanh đếm ngược 8 s trên đầu thay cho icon (spec §6.5). */
  private drawMinionBar(s: StatusInput, on: boolean): void {
    this.bar.clear();
    const left = on ? statusTracker.left(s.id, 'minion', s.now) : null;
    if (!left) {
      return;
    }
    const w = BAR_W_PX / s.zoom;
    const h = BAR_H_PX / s.zoom;
    const y = s.headY - (ICON_ABOVE_HEAD_PX - 14) / s.zoom;
    this.bar.roundRect(-w / 2, y, w, h, h / 2).fill({ color: 0x000000, alpha: 0.55 });
    this.bar.roundRect(-w / 2, y, w * left.frac, h, h / 2).fill(0xb06fd8);
  }
}

function star(cx: number, cy: number, r: number): number[] {
  const out: number[] = [];
  for (let i = 0; i < 10; i++) {
    const a = -Math.PI / 2 + (i * Math.PI) / 5;
    const k = i % 2 === 0 ? r : r * 0.45;
    out.push(cx + Math.cos(a) * k, cy + Math.sin(a) * k);
  }
  return out;
}

function hexagon(cx: number, cy: number, r: number): number[] {
  const out: number[] = [];
  for (let i = 0; i < 6; i++) {
    const a = (i * Math.PI) / 3 + Math.PI / 6;
    out.push(cx + Math.cos(a) * r, cy + Math.sin(a) * r);
  }
  return out;
}

/** note vẽ một nốt nhạc (đầu tròn + đuôi) — hát ru của Ông Ba. */
export function note(g: Graphics, x: number, y: number, s: number, color: number, alpha: number): void {
  g.ellipse(x, y, s, s * 0.75).fill({ color, alpha });
  g.moveTo(x + s * 0.9, y).lineTo(x + s * 0.9, y - s * 3).lineTo(x + s * 2, y - s * 2.4).stroke({ color, width: s * 0.4, alpha });
}
