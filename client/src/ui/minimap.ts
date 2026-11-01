import { TIER_BORDER } from '../net/protocol';
import type { GameConfig } from '../config';
import type { GameMap, LeaderboardMsg, Shape } from '../net/protocol';
import { GATE_COLORS } from '../render/palette';
import { h } from './dom';
import { narrowScreen } from './layout';
import { STR } from './strings';

/** Màu vàng kim riêng của Thiên Môn trên minimap (spec §8.4a). */
const HEAVEN = '#ffd60a';

// Colours of the static layer. Kept muted so the live markers stand out (spec §12.1).
const BG = '#13170f';
const TERRAIN: Record<string, string> = {
  border: '#6a7466', // biên khu — không ai chém được
  boulder: '#7b7b7b',
  bush: '#2f6b3a',
  mud: '#6b5433',
  wind: '#3c5a73',
};
const ME = '#ffe14d';
const VIEW = 'rgba(255,255,255,0.65)';
const CROWN = '#ffd700';
const DAO_TO = '#c084fc'; // tím Đạo Tổ — khác vàng Top 1 để không nhầm hai mục tiêu
const CHEST = '#ff9a3c';
const ALTAR_SMALL = '#9fe8ff';
const ALTAR_CENTER = '#ffd60a';

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
  /**
   * collapsed: minimap đang thu thành nút nhỏ, bấm mới mở (chủ dự án chốt 30/10 —
   * *"mini map đang che mất phần trung tâm, làm nhỏ lại, hoặc cho phép bấm mới mở ra"*).
   *
   * Chỉ thu trên màn hẹp. Trên PC màn rộng, bản đồ góc không vướng gì và việc phải bấm
   * một lần nữa mới xem được chỉ làm chậm tay.
   *
   * Ba trạng thái quay vòng khi bấm: **nút nhỏ → bản đồ góc → bản đồ to → nút nhỏ**.
   */
  private collapsed = false;
  private lb: LeaderboardMsg | null = null;
  /**
   * Cột liệt kê 5 khu + số người mỗi khu, chỉ hiện khi phóng to (spec §8.0).
   *
   * Cố ý **không** hiện vị trí ai ở đâu — chỉ số lượng. Biết khu nào đang đông là đủ để
   * chọn đi hay tránh; biết chính xác ai đứng chỗ nào thì cổng dịch chuyển hết bất ngờ.
   */
  private readonly zoneList = h('div', 'zone-list');
  /** Icon hiện thay cho bản đồ khi đang thu nhỏ. */
  private readonly pin = h('div', 'minimap-pin', '🗺');
  private myZone = 0;

  constructor(cfg: GameConfig, isTouch: boolean) {
    this.size = isTouch ? cfg.minimap.size_mobile : cfg.minimap.size_pc;
    this.ctx = this.canvas.getContext('2d');
    this.root.append(this.canvas, this.zoneList, this.pin);
    this.root.addEventListener('click', () => this.toggle());
    // Cửa sổ đổi kích thước → tính lại, nếu không bản phóng to sẽ tràn ra ngoài.
    window.addEventListener('resize', () => this.applyState());
    this.collapsed = narrowScreen();
    this.applyState();
  }

  /**
   * toggle quay vòng trạng thái. Trên màn hẹp có ba nấc (nút nhỏ → góc → to → nút nhỏ);
   * trên màn rộng chỉ hai nấc như cũ (góc ↔ to).
   */
  toggle(): void {
    if (!narrowScreen()) {
      this.collapsed = false;
      this.big = !this.big;
    } else if (this.collapsed) {
      this.collapsed = false;
    } else if (!this.big) {
      this.big = true;
    } else {
      this.big = false;
      this.collapsed = true;
    }
    this.applyState();
  }

  private applyState(): void {
    // Màn rộng thì không bao giờ ở trạng thái thu nhỏ, kể cả khi người chơi thu lại rồi
    // mới phóng to cửa sổ — nếu không họ sẽ thấy một cái nút lạ giữa HUD mà không hiểu.
    if (!narrowScreen()) {
      this.collapsed = false;
    }
    this.root.classList.toggle('big', this.big);
    this.root.classList.toggle('collapsed', this.collapsed);
    this.drawZoneList();
    this.resize();
  }

  /**
   * heavenArrow vẽ mũi tên nhỏ ở GÓC minimap chỉ hướng tới Thiên Môn (spec §8.4a).
   *
   * Chỉ vẽ khi Thiên Môn nằm ngoài khung nhìn hiện tại — thấy cổng rồi mà vẫn còn mũi tên
   * thì chỉ thêm rối.
   */
  private heavenArrow(g: CanvasRenderingContext2D, px: number, meX: number, meY: number): void {
    const gate = this.map?.gates.find((q) => q.hv);
    if (!gate) {
      return;
    }
    const dx = gate.x - meX;
    const dy = gate.y - meY;
    const d = Math.hypot(dx, dy);
    if (d < 1200) {
      return; // đã tới gần, khỏi chỉ đường nữa
    }
    const cx = px * 0.88;
    const cy = px * 0.12;
    const r = px * 0.055;
    const a = Math.atan2(dy, dx);
    g.save();
    g.translate(cx, cy);
    g.rotate(a);
    g.fillStyle = HEAVEN;
    g.beginPath();
    g.moveTo(r, 0);
    g.lineTo(-r * 0.6, r * 0.55);
    g.lineTo(-r * 0.25, 0);
    g.lineTo(-r * 0.6, -r * 0.55);
    g.closePath();
    g.fill();
    g.restore();
  }

  /** setZone nhớ khu đang đứng để tô đậm dòng tương ứng trong bảng. */
  setZone(index: number): void {
    this.myZone = index;
    this.drawZoneList();
  }

  private drawZoneList(): void {
    const counts = this.lb?.zc;
    if (!this.big || !counts || counts.length < 2) {
      this.zoneList.style.display = 'none';
      return;
    }
    this.zoneList.style.display = 'block';
    this.zoneList.replaceChildren(
      h('div', 'zone-list-title', STR.zoneListTitle),
      ...counts.map((n, i) =>
        h('div', `zone-row${i === this.myZone ? ' me' : ''}`,
          h('span', 'zone-row-name', STR.zoneNames[i] ?? `Khu ${i + 1}`),
          h('span', 'zone-row-n', String(n))),
      ),
    );
  }

  private resize(): void {
    if (this.collapsed) {
      // Nút nhỏ: cỡ do CSS quyết, canvas không cần vẽ gì.
      this.root.style.width = this.root.style.height = '';
      return;
    }
    // Không bao giờ rộng hơn chỗ trống thực tế: chừa 12px mép và phần trên đã có HUD.
    const roomW = window.innerWidth - 24;
    const roomH = window.innerHeight - 148;
    // Và không bao giờ quá 1/4 cạnh ngắn của màn hình. `size_mobile` là một con số px cố
    // định, mà "120px" trên máy 360px rộng là một phần ba bề ngang — to gấp đôi cảm giác
    // của chính con số đó trên máy 768px. Kẹp theo tỉ lệ thì màn nào cũng vừa mắt.
    const share = Math.min(window.innerWidth, window.innerHeight) * 0.24;
    const px = this.big
      ? Math.min(window.innerWidth, window.innerHeight) * 0.7
      : Math.max(70, Math.min(this.size, share, roomW, roomH));
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
    this.drawZoneList();
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
      // Chỉ vẽ thứ không đổi trạng thái: biên khu và đá tảng (spec §12.1).
      //
      // Trước v6.7 điều kiện là `hp === 0` ("không phá được"). Giờ mọi vật cản trong khu đều
      // phá được bởi một ai đó, nên thứ duy nhất vẽ lên minimap mà không bao giờ sai là
      // biên khu — vẽ tường bậc 11 rồi để Đạo Tổ chém mất thì minimap nói dối.
      const color = o.t >= TIER_BORDER ? TERRAIN.border : o.k === 'boulder' ? TERRAIN.boulder : '';
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
    if (!g || !this.map || !this.terrain || this.collapsed) {
      return; // đang thu thành nút: vẽ vào canvas ẩn chỉ tốn CPU mỗi khung
    }
    const px = this.canvas.width;
    const k = px / this.map.size;
    g.clearRect(0, 0, px, px);
    g.drawImage(this.terrain, 0, 0, px, px);
    for (const gate of this.map.gates) {
      if (gate.hv) {
        // Thiên Môn: icon vàng riêng, to hơn, LUÔN hiện (spec §8.4a). Nó là lối đi chắc
        // chắn tới Thiên Đình nên người chơi phải thấy nó ở mọi lúc, khác cổng thường.
        this.marker(g, gate.x * k, gate.y * k, px * 0.03, HEAVEN, '⛩');
        continue;
      }
      if (gate.dk) {
        continue; // bến chỉ là chỗ đặt chân, vẽ lên chỉ làm rối minimap
      }
      this.dot(g, gate.x * k, gate.y * k, px * 0.012, hex(GATE_COLORS[gate.c] ?? 0xa55bff));
    }
    this.heavenArrow(g, px, meX, meY);
    // Camera box: tells you how much of the map you can actually see right now.
    g.strokeStyle = VIEW;
    g.lineWidth = Math.max(1, px * 0.004);
    g.strokeRect((meX - viewW / 2) * k, (meY - viewH / 2) * k, viewW * k, viewH * k);
    const lb = this.lb;
    if (lb?.t1) {
      this.marker(g, (lb.t1x ?? 0) * k, (lb.t1y ?? 0) * k, px * 0.022, CROWN, '👑');
    }
    // Đạo Tổ (spec §3.4): cả phòng thấy để hợp sức săn. Vẽ sau vương miện, cùng chỗ thì
    // chấm tím đè lên — Đạo Tổ là mục tiêu hiếm và đáng chú ý hơn.
    for (const d of lb?.dt ?? []) {
      this.marker(g, d.x * k, d.y * k, px * 0.024, DAO_TO, '☯');
    }
    // Đài Tiên Khí: 5 cái cố định, nhấp nháy khi sắp phun lớn (spec §8.5).
    for (const a of lb?.al ?? []) {
      const center = a.k === 'center';
      if (a.w && Math.floor(now * 4) % 2 === 0) {
        continue;
      }
      this.marker(g, a.x * k, a.y * k, px * (center ? 0.024 : 0.016), center ? ALTAR_CENTER : ALTAR_SMALL, center ? '☯' : '◈');
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
