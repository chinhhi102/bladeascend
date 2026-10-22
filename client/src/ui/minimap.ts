import type { GameConfig } from '../config';
import type { GameMap, LeaderboardMsg, Shape } from '../net/protocol';
import { GATE_COLORS } from '../render/palette';
import { h } from './dom';

// Colours of the static layer. Kept muted so the live markers stand out (spec §12.1).
const BG = '#13170f';
const TERRAIN: Record<string, string> = {
  stone_wall: '#6a7466', // không phá được
  boulder: '#7b7b7b',
  bush: '#2f6b3a',
  mud: '#6b5433',
  wind: '#3c5a73',
};
const ME = '#ffe14d';
const VIEW = 'rgba(255,255,255,0.65)';
const CROWN = '#ffd700';
const CHEST = '#ff9a3c';

function hex(n: number): string {
  return `#${n.toString(16).padStart(6, '0')}`;
}

function shapeRect(s: Shape): { x: number; y: number; w: number; h: number } {
  if (s.c) {
    const r = s.r ?? 0;
    return { x: s.x - r, y: s.y - r, w: r * 2, h: r * 2 };
  }
  return { x: s.x, y: s.y, w: s.w ?? 0, h: s.h ?? 0 };
}

/**
 * Minimap is the top-left overview (spec §12.1). It deliberately shows almost nothing live:
 * only you, your camera box, the gates, the room's Top 1 and the Rương Báu. Other players,
 * bots, monsters and items stay hidden so ambushes still work.
 */
export class Minimap {
  readonly root = h('div', 'minimap');
  private readonly canvas = document.createElement('canvas');
  private readonly ctx: CanvasRenderingContext2D | null;
  private terrain: HTMLCanvasElement | null = null; // static layer, drawn once per map
  private map: GameMap | null = null;
  private size: number;
  private big = false;
  private lb: LeaderboardMsg | null = null;

  constructor(cfg: GameConfig, isTouch: boolean) {
    this.size = isTouch ? cfg.minimap.size_mobile : cfg.minimap.size_pc;
    this.ctx = this.canvas.getContext('2d');
    this.root.append(this.canvas);
    this.root.addEventListener('click', () => this.toggle());
    // Cửa sổ đổi kích thước → tính lại, nếu không bản phóng to sẽ tràn ra ngoài.
    window.addEventListener('resize', () => this.resize());
    this.resize();
  }

  /** toggle switches between the corner map and the big centred one (tap or key M). */
  toggle(): void {
    this.big = !this.big;
    this.root.classList.toggle('big', this.big);
    this.resize();
  }

  private resize(): void {
    // Không bao giờ rộng hơn chỗ trống thực tế: chừa 12px mép và phần trên đã có HUD.
    const roomW = window.innerWidth - 24;
    const roomH = window.innerHeight - 148;
    const px = this.big
      ? Math.min(window.innerWidth, window.innerHeight) * 0.7
      : Math.max(90, Math.min(this.size, roomW, roomH));
    const dpr = window.devicePixelRatio || 1;
    this.root.style.width = this.root.style.height = `${Math.round(px)}px`;
    this.canvas.width = this.canvas.height = Math.round(px * dpr);
    this.canvas.style.width = this.canvas.style.height = '100%';
  }

  setMap(map: GameMap): void {
    this.map = map;
    this.terrain = this.drawTerrain(map);
  }

  setLeaderboard(lb: LeaderboardMsg): void {
    this.lb = lb;
  }

  /** drawTerrain renders the static layer once: it never changes during a match. */
  private drawTerrain(map: GameMap): HTMLCanvasElement {
    const px = 512; // fixed resolution, scaled down when blitted
    const c = document.createElement('canvas');
    c.width = c.height = px;
    const g = c.getContext('2d');
    if (!g) {
      return c;
    }
    const k = px / map.size;
    g.fillStyle = BG;
    g.fillRect(0, 0, px, px);
    // Zones first (ground tint), then the obstacles that matter for navigation.
    for (const z of map.zones) {
      const color = TERRAIN[z.k];
      if (!color) {
        continue;
      }
      const r = shapeRect(z.s);
      g.fillStyle = color;
      g.globalAlpha = 0.55;
      g.fillRect(r.x * k, r.y * k, Math.max(1, r.w * k), Math.max(1, r.h * k));
      g.globalAlpha = 1;
    }
    for (const o of map.obstacles) {
      // Chỉ vẽ thứ không đổi trạng thái: tường không phá được và đá tảng (spec §12.1).
      const color = o.hp === 0 ? TERRAIN.stone_wall : o.k === 'boulder' ? TERRAIN.boulder : '';
      if (!color) {
        continue;
      }
      const r = shapeRect(o.s);
      g.fillStyle = color;
      g.fillRect(r.x * k, r.y * k, Math.max(1, r.w * k), Math.max(1, r.h * k));
    }
    return c;
  }

  /** render draws one frame. me is the camera centre in world units. */
  render(meX: number, meY: number, facing: number, viewW: number, viewH: number, now: number): void {
    const g = this.ctx;
    if (!g || !this.map || !this.terrain) {
      return;
    }
    const px = this.canvas.width;
    const k = px / this.map.size;
    g.clearRect(0, 0, px, px);
    g.drawImage(this.terrain, 0, 0, px, px);
    for (const gate of this.map.gates) {
      this.dot(g, gate.x * k, gate.y * k, px * 0.012, hex(GATE_COLORS[gate.c] ?? 0xa55bff));
    }
    // Camera box: tells you how much of the map you can actually see right now.
    g.strokeStyle = VIEW;
    g.lineWidth = Math.max(1, px * 0.004);
    g.strokeRect((meX - viewW / 2) * k, (meY - viewH / 2) * k, viewW * k, viewH * k);
    const lb = this.lb;
    if (lb?.t1) {
      this.marker(g, (lb.t1x ?? 0) * k, (lb.t1y ?? 0) * k, px * 0.022, CROWN, '👑');
    }
    if (lb?.ch) {
      // Nhấp nháy cả khi đang báo trước 10s lẫn khi rương đã hiện.
      if (!lb.chw || Math.floor(now * 2) % 2 === 0) {
        this.marker(g, (lb.chx ?? 0) * k, (lb.chy ?? 0) * k, px * 0.02, CHEST, '🎁');
      }
    }
    this.arrow(g, meX * k, meY * k, facing, px * 0.022);
  }

  private dot(g: CanvasRenderingContext2D, x: number, y: number, r: number, color: string): void {
    g.fillStyle = color;
    g.beginPath();
    g.arc(x, y, r, 0, Math.PI * 2);
    g.fill();
  }

  private marker(g: CanvasRenderingContext2D, x: number, y: number, r: number, color: string, glyph: string): void {
    this.dot(g, x, y, r, color);
    g.font = `${Math.round(r * 1.8)}px system-ui, sans-serif`;
    g.textAlign = 'center';
    g.textBaseline = 'middle';
    g.fillText(glyph, x, y);
  }

  /** arrow is me, pointing the way I am facing. */
  private arrow(g: CanvasRenderingContext2D, x: number, y: number, angle: number, r: number): void {
    g.save();
    g.translate(x, y);
    g.rotate(angle);
    g.fillStyle = ME;
    g.strokeStyle = '#000';
    g.lineWidth = Math.max(1, r * 0.18);
    g.beginPath();
    g.moveTo(r, 0);
    g.lineTo(-r * 0.7, r * 0.7);
    g.lineTo(-r * 0.3, 0);
    g.lineTo(-r * 0.7, -r * 0.7);
    g.closePath();
    g.fill();
    g.stroke();
    g.restore();
  }
}
