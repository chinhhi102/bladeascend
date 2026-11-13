// Bản đồ toàn cảnh — ADMIN_SPEC §5. Vẽ bằng CHÍNH bộ vẽ của game (`render/mapview.ts` +
// decal + ambient + vật cản), không phải một canvas vẽ lại riêng: thứ admin nhìn thấy phải
// là thứ người chơi nhìn thấy.
//
// Không có luồng vị trí người chơi (§5 bỏ luồng 5 Hz của v1.1). Bản đồ tĩnh lấy từ
// `/api/map/{id}`, art lấy theo manifest như game.
//
// Vẽ THEO YÊU CẦU (kéo, phóng, đổi lớp phủ), không chạy 60 khung/giây: toàn cảnh 12 000²
// có hàng trăm ô nền trộn dùng mặt nạ, vẽ liên tục là đốt GPU của máy admin vô ích. Bật
// "Chuyển động" thì mới chạy vòng khung (nước gợn, mây, đom đóm).
import { Application, Container, Graphics, Rectangle } from 'pixi.js';

import { Ambient } from '../../../src/render/ambient';
import { assets } from '../../../src/render/assets';
import { MapView, type MapWorld } from '../../../src/render/mapview';
import { TIER_BORDER, type GameMap, type Shape } from '../../../src/net/protocol';

/** TIER_COLORS: màu theo bậc vật cản 0–13 (lớp phủ "bậc vật cản"). */
const TIER_COLORS = [
  0x6b7a52, 0x7a8a58, 0x8a9a5e, 0x9aa864, 0xa8b06a, 0x9a8f6a, 0x8a7f72,
  0x7a6f7a, 0x6a6382, 0x5a5a8a, 0x6a5a92, 0x7a5a9a, 0x8a5aa2, 0x9a5aaa,
];

const ZONE_COLORS: Record<string, number> = { bush: 0x2fa84f, mud: 0x8a6430, ice: 0x78bee6, water: 0x2f6fd0, road: 0xb09a70, wind: 0xbfe9ff };

/** MAX_ZOOM: cận cảnh như trong game (1 đơn vị thế giới ≈ 1 điểm ảnh CSS). */
const MAX_ZOOM = 1.5;

export interface Overlays {
  tiers: boolean;
  grid: boolean;
  zones: boolean;
  /** realm: ≥ 0 thì giả lập cảnh giới r — vật cản bậc ≤ r coi như đã chém (`--check`). */
  realm: number;
  motion: boolean;
}

export const NO_OVERLAYS: Overlays = { tiers: false, grid: false, zones: false, realm: -1, motion: false };

/** shapePath vẽ đúng hình server dùng (đa giác hữu cơ nếu có), như `drawShape` của mapview. */
function shapePath(g: Graphics, s: Shape): Graphics {
  if (s.p && s.p.length >= 6) {
    return g.poly(s.p);
  }
  return s.c ? g.circle(s.x, s.y, s.r ?? 0) : g.rect(s.x, s.y, s.w ?? 0, s.h ?? 0);
}

/** AdminWorld: trạng thái giả cho bộ vẽ — không có ai trong trận, chém được mọi thứ (không viền "quá bậc"). */
class AdminWorld implements MapWorld {
  readonly brokenObstacles = new Map<number, boolean>();

  canCutObstacle(): boolean {
    return true;
  }
}

export class MapStage {
  private readonly app = new Application();
  private readonly world = new Container();
  /** mapRoot: phần do MapView dựng — huỷ và dựng lại khi đổi khu. Lớp phủ và ambient sống lâu hơn. */
  private readonly mapRoot = new Container();
  private readonly overlay = new Graphics();
  private readonly ambient = new Ambient();
  private readonly fake = new AdminWorld();
  private view: MapView | null = null;
  private map: GameMap | null = null;
  private k = 1;
  private ov: Overlays = NO_OVERLAYS;
  private frame = 0;
  private ready = false;
  private destroyed = false;
  private readonly ro: ResizeObserver;
  private lastGen = -1;
  private genTimer = 0;
  private readonly pointers = new Map<number, { x: number; y: number }>();

  constructor(private readonly host: HTMLElement) {
    this.ro = new ResizeObserver(() => this.onResize());
  }

  async init(): Promise<void> {
    await this.app.init({ preference: 'webgl', background: 0x0b0d12, antialias: true, autoStart: false, resolution: Math.min(2, window.devicePixelRatio || 1), autoDensity: true, width: this.host.clientWidth || 600, height: this.host.clientHeight || 400 });
    if (this.destroyed) {
      this.app.destroy(true);
      return;
    }
    this.app.ticker.stop();
    this.host.appendChild(this.app.canvas);
    this.app.canvas.style.touchAction = 'none';
    this.app.canvas.style.display = 'block';
    this.app.stage.addChild(this.world);
    this.world.addChild(this.mapRoot, this.ambient.root, this.overlay);
    this.bindInput(this.app.canvas);
    this.ro.observe(this.host);
    if (!assets.ready) {
      await assets.init();
    }
    // Art về muộn (nhóm tải nền) thì bộ vẽ tự đổi placeholder sang ảnh thật — chỉ cần vẽ lại.
    this.genTimer = window.setInterval(() => {
      if (assets.gen !== this.lastGen) {
        this.lastGen = assets.gen;
        this.requestRender();
      }
    }, 1000);
    this.ready = true;
  }

  /** show nạp art của khu rồi dựng bản đồ. Dựng SAU khi art về: nền chọn ảnh lúc dựng. */
  async show(map: GameMap, gateR: number): Promise<void> {
    await assets.load(['boot', `zone_${map.id}`]).catch((e: unknown) => console.warn('[admin] tải art khu lỗi', e));
    if (this.destroyed || !this.ready) {
      return;
    }
    this.mapRoot.removeChildren().forEach((c) => c.destroy({ children: true }));
    this.map = map;
    this.view = new MapView(map, gateR);
    const entities = new Container();
    entities.sortableChildren = true;
    for (const node of this.view.obstacleLayer()) {
      entities.addChild(node);
    }
    this.ambient.setBiome(map.id);
    this.mapRoot.addChild(this.view.ground, this.view.decor, this.view.under, this.view.gateLayer, entities, this.view.over);
    this.applyRealm();
    this.drawOverlay();
    this.fit();
  }

  setOverlays(o: Overlays): void {
    const realmChanged = o.realm !== this.ov.realm;
    this.ov = o;
    if (realmChanged) {
      this.applyRealm();
    }
    this.drawOverlay();
    this.requestRender();
  }

  /** fit: thu nhỏ vừa khung (chế độ "toàn cảnh"). */
  fit(): void {
    if (!this.map) {
      return;
    }
    const { width, height } = this.app.screen;
    this.k = Math.min(width, height) / this.map.size;
    this.world.scale.set(this.k);
    this.world.position.set((width - this.map.size * this.k) / 2, (height - this.map.size * this.k) / 2);
    this.drawOverlay();
    this.requestRender();
  }

  /**
   * exportPng: ảnh chụp TOÀN KHU (§5 "Ảnh chụp toàn khu"), cạnh `px` điểm ảnh.
   *
   * Đặt tạm camera sao cho cả bản đồ nằm trong [0, px] của sân khấu rồi trích đúng khung đó
   * — không phụ thuộc kích thước cửa sổ. Xong thì trả camera về chỗ cũ.
   */
  async exportPng(px: number): Promise<Blob | null> {
    if (!this.map || !this.view) {
      return null;
    }
    const keep = { k: this.k, x: this.world.position.x, y: this.world.position.y };
    const size = this.map.size;
    this.world.scale.set(px / size);
    this.world.position.set(0, 0);
    this.updateView({ x: size / 2, y: size / 2, w: size, h: size }, performance.now() / 1000);
    const canvas = this.app.renderer.extract.canvas({ target: this.app.stage, frame: new Rectangle(0, 0, px, px), resolution: 1 }) as HTMLCanvasElement;
    this.world.scale.set(keep.k);
    this.world.position.set(keep.x, keep.y);
    this.requestRender();
    return new Promise((resolve) => canvas.toBlob((b) => resolve(b), 'image/png'));
  }

  maxTextureSize(): number {
    const gl = (this.app.renderer as unknown as { gl?: WebGLRenderingContext }).gl;
    return gl ? (gl.getParameter(gl.MAX_TEXTURE_SIZE) as number) : 4096;
  }

  destroy(): void {
    this.destroyed = true;
    this.ro.disconnect();
    window.clearInterval(this.genTimer);
    cancelAnimationFrame(this.frame);
    if (this.ready) {
      this.app.destroy(true, { children: true });
    }
  }

  requestRender(): void {
    if (this.frame || !this.ready) {
      return;
    }
    this.frame = requestAnimationFrame(() => {
      this.frame = 0;
      this.renderNow();
      if (this.ov.motion) {
        this.requestRender();
      }
    });
  }

  /** visibleRect: khung đang nhìn, toạ độ thế giới (tâm + kích thước, như camera của game). */
  private visibleRect(): { x: number; y: number; w: number; h: number } {
    const { width, height } = this.app.screen;
    const w = width / this.k;
    const h = height / this.k;
    return { x: (width / 2 - this.world.position.x) / this.k, y: (height / 2 - this.world.position.y) / this.k, w, h };
  }

  private updateView(v: { x: number; y: number; w: number; h: number }, now: number): void {
    this.view?.update(this.fake, now, -1, null, 1, v);
    if (this.ov.motion) {
      this.ambient.update(v.x, v.y, v.w / 2, v.h / 2, now);
    } else {
      this.ambient.clear();
    }
  }

  private renderNow(): void {
    if (!this.view) {
      this.app.render();
      return;
    }
    this.updateView(this.visibleRect(), performance.now() / 1000);
    this.app.render();
  }

  private onResize(): void {
    if (!this.ready) {
      return;
    }
    this.app.renderer.resize(this.host.clientWidth, this.host.clientHeight);
    this.requestRender();
  }

  /** applyRealm: giả lập cảnh giới r — vật cản bậc ≤ r biến mất như đã bị chém. */
  private applyRealm(): void {
    this.fake.brokenObstacles.clear();
    if (!this.map || this.ov.realm < 0) {
      return;
    }
    for (const o of this.map.obstacles) {
      if (o.t < TIER_BORDER && o.t <= this.ov.realm) {
        this.fake.brokenObstacles.set(o.id, true);
      }
    }
  }

  private drawOverlay(): void {
    const g = this.overlay.clear();
    const m = this.map;
    if (!m) {
      return;
    }
    if (this.ov.zones) {
      for (const z of m.zones) {
        shapePath(g, z.s).fill({ color: ZONE_COLORS[z.k] ?? 0x999999, alpha: 0.35 }).stroke({ color: ZONE_COLORS[z.k] ?? 0x999999, width: 3 / this.k, alpha: 0.9 });
      }
    }
    if (this.ov.tiers) {
      this.drawTiers(g, m);
    }
    if (this.ov.grid) {
      for (let v = 0; v <= m.size; v += 1000) {
        g.moveTo(v, 0).lineTo(v, m.size).moveTo(0, v).lineTo(m.size, v);
      }
      g.stroke({ color: 0xffffff, width: 1.5 / this.k, alpha: 0.45 });
    }
  }

  private drawTiers(g: Graphics, m: GameMap): void {
    for (const o of m.obstacles) {
      if (this.fake.brokenObstacles.get(o.id)) {
        continue;
      }
      const color = o.t >= TIER_BORDER ? 0x434a55 : TIER_COLORS[Math.min(o.t, 13)];
      shapePath(g, o.s).fill({ color, alpha: 0.6 });
    }
  }

  private zoomAt(sx: number, sy: number, factor: number): void {
    if (!this.map) {
      return;
    }
    const minK = (Math.min(this.app.screen.width, this.app.screen.height) / this.map.size) * 0.6;
    const k = Math.min(MAX_ZOOM, Math.max(minK, this.k * factor));
    const wx = (sx - this.world.position.x) / this.k;
    const wy = (sy - this.world.position.y) / this.k;
    this.k = k;
    this.world.scale.set(k);
    this.world.position.set(sx - wx * k, sy - wy * k);
    this.drawOverlay();
    this.requestRender();
  }

  private bindInput(c: HTMLCanvasElement): void {
    c.addEventListener('wheel', (e) => {
      e.preventDefault();
      this.zoomAt(e.offsetX, e.offsetY, Math.exp(-e.deltaY * 0.0015));
    }, { passive: false });
    c.addEventListener('pointerdown', (e) => {
      c.setPointerCapture(e.pointerId);
      this.pointers.set(e.pointerId, { x: e.offsetX, y: e.offsetY });
    });
    c.addEventListener('pointermove', (e) => this.onPointerMove(e));
    const up = (e: PointerEvent): void => {
      this.pointers.delete(e.pointerId);
    };
    c.addEventListener('pointerup', up);
    c.addEventListener('pointercancel', up);
  }

  /** onPointerMove: một ngón = kéo; hai ngón = kéo + phóng (điện thoại). */
  private onPointerMove(e: PointerEvent): void {
    const prev = this.pointers.get(e.pointerId);
    if (!prev) {
      return;
    }
    const pts = [...this.pointers.values()];
    const cur = { x: e.offsetX, y: e.offsetY };
    if (pts.length === 2) {
      const other = pts[0] === prev ? pts[1] : pts[0];
      const before = Math.hypot(prev.x - other.x, prev.y - other.y);
      const after = Math.hypot(cur.x - other.x, cur.y - other.y);
      if (before > 0) {
        this.zoomAt((cur.x + other.x) / 2, (cur.y + other.y) / 2, after / before);
      }
    }
    this.world.position.x += (cur.x - prev.x) / Math.max(1, pts.length);
    this.world.position.y += (cur.y - prev.y) / Math.max(1, pts.length);
    this.pointers.set(e.pointerId, cur);
    this.requestRender();
  }
}
