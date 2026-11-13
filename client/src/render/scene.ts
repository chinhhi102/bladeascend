// Bộ đọc KTX2 KHÔNG tự đăng ký khi `import 'pixi.js'` — phải nạp riêng đường phụ này.
//
// Thiếu nó thì mọi tệp `.ktx2` tải về xong rồi bị bỏ, với đúng một dòng cảnh báo trong
// console: "could not be loaded as we don't know how to parse it". Hậu quả nhìn thấy là
// nhân vật về placeholder và mặt đất thành một mảng xanh trơn — tức là ART HỎNG HẾT, nhưng
// không có lỗi nào được ném và bài đo VRAM lại báo con số ĐẸP HƠN vì chẳng còn texture nào.
import 'pixi.js/ktx2';
import { Application, Container, setKTXTranscoderPath, Texture, type Ticker, UPDATE_PRIORITY } from 'pixi.js';
import { AnimViewer } from './animviewer';
import { guardLines, killFilters, noteDrawError, noteSysError, sysOffList, sysRun } from './guard';
import { census, type Census, censusFlat, censusLine, ledgerSet } from './ledger';
import { setEmoteCharCount } from './emote';
import { paceTicker } from './framepace';
import { countDrawCalls, countNodes, drawCalls, type LayerStat, measureLayers } from './drawstats';
import { probeEnabled, probeFrame, probeHeap, probeLines, probeMark, probeStats, probeSubFlush } from './probe';
import { assets } from './assets';
import { statusWork, statusWorkTick } from './statusview';
import { ringWork, ringWorkTick } from './swordring';
import { SkillPresenter, type StatusChip } from './skillpresent';
import { type BudgetCat, gpuBiggest, gpuBudget, gpuMemMB, gpuPeakMB, gpuSites, gpuListCat, gpuTextureCount, gpuUnclassified, vramLine, trackGpuMemory, trackSourceNames, trackTextureFrom } from './gpumem';
import { gpuBudgetMB, isMobile, rendererResolution } from './device';
import { FADE_OUT_SECS, ViewPool } from './viewpool';
import { BobaOverlay } from './boba';

import { sound } from '../audio/sound';
import { absorbable, banksTienKhi, cameraViewHeight, type GameConfig, realmName, realmOf, skillRange, tienIndex } from '../config';
import type { ClientWorld } from '../net/state';
import type { ItemState } from '../net/protocol';
import { FLAG, type GameEvent, moodName, statusOn } from '../net/protocol';
import { pick, STR } from '../ui/strings';
import { Beast, type BeastMode, TIEN_BEASTS } from './beast';
import { ItemSprite, MonsterSprite } from './entities';
import { MapView, type Viewer } from './mapview';
import { OrbLayer } from './orblayer';
import { Ambient } from './ambient';
import { Breakthrough, HITSTOP_SECS } from './breakthrough';
import { autoQuality, autoQualityByMemory, emergency, gpuPressure, quality } from './quality';
import { AURA_FAR_PX, TienAura } from './tienaura';
import { FloatTextLayer } from './floattext';
import { StallView } from './stallview';
import { VfxLayer } from './vfxlayer';
import { PlayerSprite } from './player';
import { ringGeom } from './ringgeom';
import { Vfx } from './vfx';

const SUCK_SECS = 0.22; // vật phẩm nhặt được bay hút vào người trong 0.22s
// Ngân sách số thanh kiếm vẽ mỗi khung hình. Vượt thì các VÒNG DÀY bị giãn bớt — vòng dày
// vốn là một vành sát thương liền (§4.1b) nên bỏ bớt thanh trong vành nhìn vẫn y hệt, còn
// vòng thưa giữ nguyên vì ở đó từng thanh là một lưỡi kiếm riêng người chơi phải né.
// Không có trần này thì 5 người Đại Thừa 500 kiếm gặp nhau là 2 500 thanh × 60 khung/giây.
// Spec §3.5 v4.7: thần thú LUÔN bay quanh người Tiên giới, không hẹn giờ, không tắt được.
// Giới hạn duy nhất là vệt hạt — chỉ bật cho vài con gần mình nhất.
const BEAST_TRAIL_NEAR = 6;
// Chủ nhỏ hơn ngần này pixel trên màn hình thì thần thú rút gọn còn đầu + 4 đốt.
const BEAST_FAR_PX = 40;

const SWORD_BUDGET = 1600;

/** Watchdog §13.8: không có khung nào trong ngần này (tab đang hiện) → khởi động lại ticker. */
const WATCHDOG_MS = 3000;
/** Nhịp kiểm của watchdog. 1 giây là đủ: ngưỡng là 3 giây, chậm nhất phát hiện sau 4. */
const WATCHDOG_CHECK_MS = 1000;

// `drawCalls` / `countDrawCalls` chuyển sang `drawstats.ts` (cảnh thử dùng chung); xuất
// lại ở đây để các chỗ đang `import { drawCalls } from './scene'` không phải sửa theo.
export { countDrawCalls, drawCalls };

const MOOD_EMOTE_CHANCE = 0.12; // chance per mood change that a character shouts (run/chase/flee)

/** GameScene renders the ClientWorld every frame with a camera that zooms out as you grow (spec §4.1). */
export class GameScene {
  /**
   * app KHÔNG còn `readonly`: §13.8 cho phép dựng lại renderer sau lỗi vẽ lặp lại, và dựng
   * lại nghĩa là thay hẳn `Application`. Mọi nơi đọc `scene.app` phải đọc qua trường này
   * chứ không được giữ bản sao — giữ bản sao là trỏ vào renderer đã huỷ.
   */
  app = new Application();
  /** parent / maxFps: nhớ lại để `buildApp` dựng lại được renderer mà không cần `init`. */
  private parent: HTMLElement | null = null;
  private maxFps = 60;
  /** pace: nhịp khung hiện tại (framepace.ts); dựng lại renderer phải đặt lại đúng nhịp. */
  private pace = 1;
  /**
   * onFrame: phần CẬP NHẬT của một khung, do `main.ts` cắm vào.
   *
   * Trước đây `main.ts` tự gọi `app.ticker.add(...)`. Không dùng được nữa: dựng lại
   * renderer là thay `Application`, và lời đăng ký kia nằm trên ticker cũ — sau khi dựng
   * lại thì game vẽ cảnh nhưng không còn cập nhật gì nữa.
   */
  onFrame: ((dt: number) => void) | null = null;
  /** onGuardEvent: báo cho telemetry mỗi lần lớp chống sập phải ra tay (§13.8). */
  onGuardEvent: ((kind: string, err: unknown) => void) | null = null;
  /**
   * onBreadcrumb: ghi một VIỆC vào vòng 30 của telemetry — không gửi gói riêng.
   *
   * Khác `onGuardEvent` ở chỗ đó: sự kiện telemetry bị chặn 20 cái/phút, còn breadcrumb chỉ
   * nằm trong bộ nhớ và đi kèm khi CÓ lỗi. Nên chỗ nào đổi trạng thái cũng ghi được mà không
   * tốn gì.
   */
  onBreadcrumb: ((code: string, detail?: string | number) => void) | null = null;
  /** onRebuildFailed: dựng lại renderer thất bại → `main.ts` tải lại trang + vào lại phiên (§7.1). */
  onRebuildFailed: (() => void) | null = null;
  /**
   * onCanvasReplaced: dựng lại renderer xong thì canvas là một thẻ MỚI.
   *
   * `main.ts` phải gắn lại bộ canh mất ngữ cảnh WebGL lên thẻ mới — bỏ bước này thì sau
   * lần dựng lại đầu tiên, một cú mất ngữ cảnh THẬT không còn ai bắt.
   */
  onCanvasReplaced: ((canvas: HTMLCanvasElement) => void) | null = null;
  private rebuilding = false;

  /** isRebuilding: đang thay renderer — bộ canh WebGL phải bỏ qua cú mất ngữ cảnh lúc này. */
  isRebuilding(): boolean {
    return this.rebuilding;
  }
  private rebuilds = 0;
  private lastTickAt = 0;
  private watchdogTried = false;
  private watchdogRestarts = 0;
  /** resizes: số lần canvas đổi cỡ — đi kèm mọi báo cáo lỗi (§13.8 "số lần resize"). */
  private resizes = 0;
  /** drawFaults: số khung còn phải ném lỗi giả trong khâu vẽ (`kx.faultDraw`, §17). */
  private drawFaults = 0;
  /** lastCensus / censusAt: sổ tài nguyên §13.9, chốt 1 lần/giây. */
  private lastCensus: Census = census({ children: [] });
  private censusAt = 0;
  private readonly camera = new Container();
  private readonly entities = new Container();
  /** Viên Linh khí vẽ riêng bằng particle: vài trăm viên cùng lúc, xem orblayer.ts. */
  private readonly orbs = new OrbLayer();
  /** Lớp phát sáng dùng chung, ăn một lượt bloom cho cả màn hình (spec §3.6). */
  private readonly glow = new VfxLayer();
  /** Sự sống của bản đồ: mây trôi, lá rơi, đom đóm… chỉ ở client (spec §8.6 mục 5). */
  private readonly ambient = new Ambient();
  /** Màn đột phá Tiên giới: sóng xung kích + khựng hình + chớp trắng + chữ (spec §3.6). */
  /** Lớp trân châu dính màn hình khi trúng kỹ năng Thằng Mập (spec §6.3). */
  readonly boba = new BobaOverlay();
  private readonly breakthrough = new Breakthrough();
  private hitstopUntil = 0;
  private readonly auras = new Map<number, TienAura>();
  private readonly rings = new Container();
  private readonly vfx = new Vfx();
  /** skill: mọi thứ NHÌN THẤY của kỹ năng — lớp ①②⑤, rung, mũi tên (spec v6.36 §6.5). */
  private readonly skill: SkillPresenter;
  private mapView: MapView | null = null;
  private readonly players = new Map<number, PlayerSprite>();
  /**
   * shadows: MỌI cái bóng đổ nằm chung một lớp, ngay trên mặt đất.
   *
   * Gom lại vì tất cả dùng CHUNG một texture: nằm chung một lớp thì Pixi gom hết vào một lô
   * vẽ, còn nằm rải trong `root` của từng nhân vật thì mỗi cái bóng kẹp giữa hai atlas và
   * tự ngắt một lô. Bóng vốn nằm dưới mọi nhân vật nên gom lại không đổi thứ tự nhìn thấy.
   */
  private readonly shadows = new Container();
  /** ringsMode: chế độ vòng kiếm của bài đo (xem `setRings`). Luôn 'all' khi chơi thật. */
  private ringsMode: 'all' | 'mine' | 'none' = 'all';
  /** Chữ nổi trong thế giới: "Cần {cảnh giới}" khi chém vật cản quá bậc (§8.2). */
  private readonly floatText = new FloatTextLayer();
  /** Dãy sạp tạm của Cô Ba (§6.3 v6.10) — vòng đời 6 giây, tách khỏi MapView. */
  private readonly stallView = new StallView();
  private readonly monsters = new Map<number, MonsterSprite>();
  private readonly items = new Map<number, ItemSprite>();
  // Items picked up this snapshot: itemId → who grabbed it, so we can fly them in.
  private readonly pickedBy = new Map<number, number>();
  private readonly flying: { sp: ItemSprite; by: number; from: { x: number; y: number }; at: number }[] = [];
  /** Bản sao trạng thái viên EXP, chỉ để dựng hiệu ứng bay hút lúc nó vừa bị nhặt. */
  private readonly lastOrb = new Map<number, ItemState>();
  private readonly lastMood = new Map<number, string>();
  private zoom = 1;
  private viewH = 900;
  /** Tâm máy quay (toạ độ thế giới), đuổi theo nhân vật — null = chưa có, đặt thẳng lần đầu. */
  private camAt: { x: number; y: number } | null = null;
  private world: ClientWorld | null = null;
  private lastFrames: ReturnType<ClientWorld['playerFrames']> = [];
  private fps = 0;
  private fpsLast = 0;
  private fpsAcc = 0;
  private fpsFrames = 0;
  private lastRenderMs = 0;
  /**
   * pixiMs: thời gian của chính `app.render()` — trung bình trượt.
   *
   * `lastRenderMs` nghe như "thời gian vẽ" nhưng nó đo vòng CẬP NHẬT (mọi hệ thống trong
   * `onFrame`). Lượt vẽ của Pixi — duyệt cây cảnh, gom lô, gói lại buffer particle, nạp lên
   * GPU — trước nay không được đo ở đâu cả.
   *
   * Đó không phải chi tiết nhỏ: lượt đo 12/11 (máy rảnh, CPU 6×) cho khung 33 ms trong khi
   * TỔNG mọi cột đo được chỉ ~5 ms. 28 ms còn lại nằm đúng ở chỗ không ai bấm giờ, và nó
   * khớp với hồ sơ CPU của B15 ("không có hàm nào của mã game trong top 22, toàn bộ là
   * PixiJS duyệt cây cảnh và gom lô"). Tối ưu phần vẽ mà không có cột này thì mọi phép đo
   * trước/sau đều đọc ra "không đổi gì".
   */
  private pixiMs = 0;
  /** charsInUse: dùng lại mỗi khung thay vì cấp phát Set mới — §5 "không tạo object lúc chơi". */
  private readonly charsInUse = new Set<string>();
  /** myReach: tâm + tầm kiếm của mình, tính một lần mỗi khung (camera và viền vật cản dùng chung). */
  private readonly myReach = { x: 0, y: 0, r: 0 };
  /** myChar: nhân vật của chính mình, không bao giờ nhả. */
  myChar = '';
  private detail = 1;
  private lastSpark = { t: 0, x: 0, y: 0 };
  private showHitbox = false;
  // Thần thú Tiên giới (spec §3.5): mỗi con có 2 lớp, một dưới nhân vật một trên.
  private beasts: { b: Beast; owner: number }[] = [];
  /** Cảnh giới Tiên của từng người ở khung hình trước — RealmDrop cần biết họ VỪA là gì. */
  private lastTien = new Map<number, number>();
  /** memCheckAt: lần cuối kiểm ngân sách bộ nhớ GPU (giây). */
  private memCheckAt = 0;
  /** onMemoryPressure: báo ra ngoài khi phải tự hạ chất lượng vì bộ nhớ — telemetry ghi lại. */
  onMemoryPressure: ((mb: number) => void) | null = null;

  constructor(private readonly cfg: GameConfig) {
    this.skill = new SkillPresenter(cfg, {
      myId: () => this.world?.myId ?? 0,
      casterAt: (id) => this.casterAt(id),
      sprite: (id) => this.players.get(id),
      toScreen: (x, y) => ({ x: this.camera.x + x * this.zoom, y: this.camera.y + y * this.zoom }),
      screen: () => ({ w: this.app.screen.width, h: this.app.screen.height }),
      floatText: this.floatText,
      vfx: this.vfx,
    });
  }

  /** viewSize is the world-space rectangle currently on screen (used by the minimap, §12.1). */
  viewSize(): { w: number; h: number } {
    return { w: this.viewH * (this.app.screen.width / Math.max(this.app.screen.height, 1)), h: this.viewH };
  }

  async init(parent: HTMLElement, maxFps: number): Promise<void> {
    this.parent = parent;
    this.maxFps = maxFps;
    await this.buildApp();
    this.startWatchdog();
  }

  /**
   * buildApp dựng `Application` và gắn nó vào trang. Chạy lúc khởi động VÀ lúc phải **dựng
   * lại renderer** sau lỗi vẽ lặp lại (§13.8) — nên mọi thứ gắn với renderer phải nằm ở
   * đây, không nằm trong `init`.
   */
  private async buildApp(): Promise<void> {
    const parent = this.parent;
    if (!parent) {
      throw new Error('buildApp trước init');
    }
    trackTextureFrom(Texture);
    // Bộ giải mã KTX2 phục vụ TẠI CHỖ, không lấy từ CDN.
    //
    // Mặc định Pixi trỏ `cdn.jsdelivr.net`. Game này chạy trên máy chủ LAN của đội test và
    // sẽ chạy trên k8s sau — một phụ thuộc mạng ngoài ngay trong đường vào trận là thứ hỏng
    // đúng lúc không ai gỡ được. Tệp nằm ở `client/public/transcoders/ktx`.
    setKTXTranscoderPath({ jsUrl: '/transcoders/ktx/libktx.js', wasmUrl: '/transcoders/ktx/libktx.wasm' });
    await this.app.init({
      // Ép WebGL. Pixi 8 vốn xếp WebGL trước, nhưng ghi rõ để một lần nâng Pixi đổi thứ tự
      // mặc định không lặng lẽ chuyển người chơi sang WebGPU — đường đó chưa từng được đo
      // (`countDrawCalls`, `trackGpuMemory` bên dưới chỉ đọc được WebGL).
      preference: 'webgl',
      resizeTo: window,
      background: 0x2b2b2b,
      antialias: true,
      // `resolution` quyết định mọi bộ đệm khung hình và render texture (xem device.ts).
      // iPhone DPR 3 để nguyên là tốn gấp 9 lần DPR 1.
      resolution: rendererResolution(quality() === 'low'),
      autoDensity: true,
    });
    this.app.ticker.maxFPS = this.maxFps;
    const gl = (this.app.renderer as unknown as { gl?: WebGL2RenderingContext }).gl;
    if (gl) {
      // Đo bộ nhớ texture LUÔN BẬT, kể cả bản chơi thật: nó là chỉ số quyết định tab có bị
      // iOS giết hay không, và chi phí là vài phép cộng mỗi lần nạp ảnh — không phải mỗi
      // khung hình. Khác hẳn `countDrawCalls` (bọc đúng hàm vẽ, chạy hàng nghìn lần/giây)
      // nên cái đó vẫn chỉ bật khi đang đo.
      trackGpuMemory(gl);
      trackSourceNames(
        (this.app.renderer as unknown as { texture: Record<string, unknown> }).texture,
      );
      if (new URLSearchParams(location.search).get('debug') === 'true') {
        countDrawCalls(gl);
      }
    }
    parent.appendChild(this.app.canvas);
    this.entities.sortableChildren = true;
    this.rings.sortableChildren = true;
    // Lớp trân châu nằm TRÊN cảnh nhưng vẫn DƯỚI HUD (HUD là DOM, không phải Pixi), đúng
    // yêu cầu "không che HUD" của spec §6.3.
    // Bộ xem animation nằm TRÊN CÙNG và ngoài camera: nó là khung xem cố định giữa màn
    // hình, không phải thứ nằm trong thế giới (ANSWERS mục 1).
    this.app.stage.addChild(this.camera, this.skill.arrows.root, this.breakthrough.overlay, this.boba.root, this.animViewer.root);
    this.boba.resize(this.app.screen.width, this.app.screen.height);
    this.app.renderer.on('resize', (w: number, h: number) => {
      this.resizes++;
      this.onBreadcrumb?.('doi-co', `${Math.round(w)}x${Math.round(h)}`);
      this.boba.resize(w, h);
    });
    this.takeOverTicker();
  }

  /**
   * takeOverTicker giành lại quyền gọi `render()` từ Pixi.
   *
   * `TickerPlugin` tự gắn `app.render` vào ticker ở mức ưu tiên LOW. Để nguyên thì lỗi vẽ
   * ném ra từ bên trong ticker của Pixi và **không có chỗ nào bắt được** — nó đi thẳng ra
   * `window.onerror`, và vì ticker vẫn chạy nên nó lặp lại 60 lần mỗi giây. Gỡ ra rồi tự
   * gọi là cách duy nhất đặt được try/catch quanh khâu vẽ (§13.8).
   *
   * Thứ tự trong một khung vẫn y như cũ: cập nhật (onFrame) → vẽ.
   */
  private takeOverTicker(): void {
    this.app.ticker.remove(this.app.render, this.app);
    this.app.ticker.add(this.tickSafe, this, UPDATE_PRIORITY.LOW);
  }

  /** tickSafe là THÂN TICKER có bắt lỗi (§13.8 "thân ticker bọc try/catch"). */
  private tickSafe(t: Ticker): void {
    this.lastTickAt = performance.now();
    // Phần cập nhật: `main.ts` cắm vào đây. Bên trong nó mỗi hệ thống đã có ranh giới
    // riêng (`sysRun`), nên khối này chỉ còn đỡ những gì rơi ra ngoài mọi ranh giới.
    try {
      this.onFrame?.(t.deltaMS / 1000);
    } catch (err) {
      noteSysError('hud', err);
    }
    // Phần vẽ: luật riêng của §13.8 — bỏ khung + tắt bộ lọc, lặp 3 lần/10 s thì dựng lại.
    try {
      if (this.drawFaults > 0) {
        this.drawFaults--;
        throw new Error('lỗi giả trong khâu vẽ (kx.faultDraw)');
      }
      const t = performance.now();
      this.app.render();
      this.pixiMs = this.pixiMs * 0.9 + (performance.now() - t) * 0.1;
    } catch (err) {
      this.onDrawError(err);
    }
  }

  /** faultDraw hẹn khâu vẽ ném lỗi giả `n` khung tiếp theo (checklist §17). */
  faultDraw(n = 1): void {
    this.drawFaults = Math.max(0, n);
  }

  /** onDrawError: bỏ khung này, tắt mọi bộ lọc; lỗi lặp ≥ 3 lần/10 s thì dựng lại renderer. */
  private onDrawError(err: unknown): void {
    const act = noteDrawError(err);
    this.onGuardEvent?.('draw_error', err);
    killFilters();
    this.stripFilters();
    if (act === 'rebuild') {
      // HOÃN sang nhịp sau, KHÔNG dựng lại ngay tại đây.
      //
      // `onDrawError` chạy từ bên trong `ticker.update()`. `rebuild()` huỷ cả
      // `Application`, mà huỷ Application là huỷ chính cái ticker đang chạy dòng này —
      // Pixi ném lỗi giữa vòng lặp của nó, `rebuild` bắt được rồi rơi xuống đường lùi
      // "tải lại trang". Nhìn từ ngoài: lỗi vẽ lần thứ ba làm trang tải lại, đúng thứ
      // §13.8 nói là KHÔNG được làm. Bắt được bằng `tools/browsertest/crashtest.mjs`.
      window.setTimeout(() => void this.rebuild(), 0);
    }
  }

  /** stripFilters gỡ filter khỏi mọi container đang đeo (bloom, sóng xung kích, nền). */
  private stripFilters(): void {
    for (const c of [this.camera, this.glow.root, this.mapView?.ground, this.mapView?.decor]) {
      if (c) {
        c.filters = [];
      }
    }
    this.glow.apply(); // đặt lại cờ bên trong VfxLayer, nếu không nó tưởng bloom vẫn đang gắn
  }

  /**
   * rebuild dựng lại renderer + cây cảnh từ trạng thái hiện có, KHÔNG tải lại trang (§13.8).
   *
   * Cách làm: gỡ các container gốc khỏi stage cũ (không huỷ chúng — chúng là trạng thái),
   * huỷ `Application` cũ **không đụng texture** (texture thuộc về `assets`, huỷ là phải tải
   * lại cả bộ art), dựng Application mới rồi gắn lại cảnh từ `world`.
   *
   * Thất bại thì báo ra ngoài để `main.ts` tải lại trang và vào lại phiên (§7.1) — đúng
   * đường lùi mà spec chỉ định.
   */
  async rebuild(): Promise<void> {
    if (this.rebuilding) {
      return;
    }
    this.rebuilding = true;
    const old = this.app;
    try {
      old.ticker.remove(this.tickSafe, this);
      old.stage.removeChildren();
      // `children: false` — mấy container trên là trạng thái sống của cảnh, huỷ chúng là
      // mất luôn chỗ để dựng lại. `texture: false` vì texture do `assets` sở hữu.
      old.destroy({ removeView: true }, { children: false, texture: false, textureSource: false, context: false });
      this.app = new Application();
      // Renderer mới bắt đầu ở độ phân giải mặc định, nên cờ "đang ở bước khẩn" phải về 0 —
      // không thì `applyEmergencyResolution` tưởng đã áp rồi và không áp lại.
      this.resEmergency = false;
      await this.buildApp();
      // Dựng lại cây cảnh: `setWorld` tự `clear()` rồi tạo MapView mới và xếp lại các lớp.
      const w = this.world;
      if (w) {
        this.setWorld(w);
      }
      this.rebuilds++;
      this.onCanvasReplaced?.(this.app.canvas as HTMLCanvasElement);
      this.onGuardEvent?.('renderer_rebuilt', null);
      console.warn(`[chống sập] đã dựng lại renderer (lần ${this.rebuilds})`);
    } catch (err) {
      console.error('[chống sập] dựng lại renderer THẤT BẠI', err);
      this.onGuardEvent?.('renderer_rebuild_failed', err);
      this.onRebuildFailed?.();
    } finally {
      this.rebuilding = false;
    }
  }

  /**
   * startWatchdog: không có khung nào trong 3 giây **khi tab đang hiện** → khởi động lại
   * ticker; vẫn chết → dựng lại renderer (§13.8).
   *
   * Điều kiện "tab đang hiện" là bắt buộc: trình duyệt hạ rAF xuống ~1 Hz hoặc dừng hẳn khi
   * tab ẩn, nên không kiểm thì mỗi lần người chơi chuyển sang app khác là một lần dựng lại
   * renderer — tự tạo ra đúng sự cố đang đi chữa.
   */
  private startWatchdog(): void {
    this.lastTickAt = performance.now();
    window.setInterval(() => {
      if (document.visibilityState !== 'visible' || this.rebuilding) {
        return;
      }
      const stuck = performance.now() - this.lastTickAt;
      if (stuck < WATCHDOG_MS) {
        this.watchdogTried = false;
        return;
      }
      if (!this.watchdogTried) {
        this.watchdogTried = true;
        this.watchdogRestarts++;
        console.warn(`[chống sập] ${(stuck / 1000).toFixed(1)}s không có khung — khởi động lại ticker`);
        this.onGuardEvent?.('watchdog_restart', null);
        this.app.ticker.start();
        paceTicker(this.app.ticker, this.pace);
        this.lastTickAt = performance.now();
        return;
      }
      console.warn('[chống sập] khởi động lại ticker không ăn thua — dựng lại renderer');
      this.watchdogTried = false;
      this.lastTickAt = performance.now();
      void this.rebuild();
    }, WATCHDOG_CHECK_MS);
  }

  /** resEmergency: bước khẩn đã áp resolution 1.0 chưa — chỉ đổi khi trạng thái đổi. */
  private resEmergency = false;

  /**
   * applyEmergencyResolution: bước khẩn (spec §13.6 mục 8) hạ resolution về 1.0; thoát bước
   * khẩn thì trả về mức của chất lượng hiện tại. Chỉ gọi `resize` khi trạng thái ĐỔI — đổi
   * resolution là dựng lại bộ đệm khung hình, làm mỗi khung là tự giết FPS.
   */
  private applyEmergencyResolution(): void {
    const on = emergency();
    if (on === this.resEmergency) {
      return;
    }
    this.resEmergency = on;
    const res = on ? 1 : rendererResolution(quality() === 'low');
    this.app.renderer.resize(this.app.screen.width, this.app.screen.height, res);
  }

  setWorld(w: ClientWorld): void {
    this.world = w;
    this.clear();
    // Bản đồ khu CŨ phải được huỷ hẳn, không chỉ gỡ khỏi cảnh: `camera.removeChildren()`
    // bỏ node nhưng texture nền vẫn nằm trên GPU (PixiJS không thu hồi theo bộ dọn rác JS).
    // Telemetry người thật: 181 lượt đổi khu một phiên — bỏ quên mỗi lượt vài chục MB.
    this.mapView?.destroy();
    this.mapView = null;
    this.camAt = null; // khu mới: đặt máy quay thẳng vào chỗ đứng, không lướt từ khu cũ sang
    if (!w.map) {
      return;
    }
    this.mapView = new MapView(w.map, this.cfg.gates.radius);
    for (const node of this.mapView.obstacleLayer()) {
      this.entities.addChild(node);
    }
    // Rings sit under every character so a blade never covers a body.
    // Viên EXP luôn nằm dưới mọi thứ khác nên xếp ngay trên mặt đất, không cần y-sort.
    // Lớp phát sáng nằm dưới nhân vật (hào quang ở sau lưng) nhưng trên vòng kiếm, để
    // hàng trăm thanh kiếm không che mất hào quang.
    // Cổng nằm NGAY TRÊN vòng kiếm, dưới nhân vật: vòng kiếm không che được cổng.
    this.entities.addChild(this.stallView.layer);
    this.camera.addChild(this.mapView.ground, this.mapView.decor, this.mapView.under, this.shadows, this.orbs.root, this.rings, this.mapView.gateLayer, this.glow.root, this.entities, this.mapView.over, this.ambient.root, ...this.skill.layers, this.vfx.layer, this.floatText.layer);
    this.ambient.setBiome(w.map.id);
  }

  clear(): void {
    this.skill.clear();
    this.floatText.clear();
    this.stallView.clear();
    this.camera.removeChildren();
    // Pool phải dọn theo: view trong pool thuộc về khu CŨ, giữ lại là giữ texture của một
    // bản đồ không còn dùng — đúng kiểu rò rỉ bộ nhớ mang tên "tối ưu".
    this.pool.clear();
    this.leaving.clear();
    for (const p of this.players.values()) {
      p.destroy();
    }
    for (const m of this.monsters.values()) {
      m.destroy();
    }
    for (const it of this.items.values()) {
      it.destroy();
    }
    for (const f of this.flying) {
      f.sp.destroy();
    }
    for (const e of this.beasts) {
      e.b.destroy();
    }
    this.beasts.length = 0;
    this.lastTien.clear();
    this.flying.length = 0;
    this.pickedBy.clear();
    this.orbs.clear();
    this.lastOrb.clear();
    this.ambient.clear();
    for (const a of this.auras.values()) {
      a.destroy();
    }
    this.auras.clear();
    this.players.clear();
    this.monsters.clear();
    this.items.clear();
    this.entities.removeChildren();
    this.rings.removeChildren();
  }

  /** render one frame; dt in seconds. */
  render(dt: number): void {
    ringWorkTick();
    statusWorkTick();
    const w = this.world;
    if (!w || !this.mapView) {
      return;
    }
    const t0 = performance.now();
    const now = t0 / 1000;
    this.breakthrough.update(now, this.app.screen.width, this.app.screen.height);
    this.boba.update(now, dt);
    if (now < this.hitstopUntil) {
      return; // khựng hình: đóng băng cảnh một nhịp rất ngắn cho cú đột phá "nặng" hơn
    }
    probeFrame();
    // Từ đây xuống, MỖI hệ thống là một ranh giới bắt lỗi riêng (§13.8, render/guard.ts).
    // Một hệ thống ném lỗi thì nó bị tắt 2 giây, các hệ thống còn lại vẫn vẽ khung này.
    //
    // Nội suy hỏng thì giữ nguyên bộ khung hình CŨ thay vì bỏ trống: nhân vật đứng im một
    // nhịp còn hơn biến mất khỏi màn hình rồi hiện lại.
    let frames = this.lastFrames;
    sysRun('interp', () => {
      frames = w.playerFrames(dt);
    });
    probeMark('interp');
    this.lastFrames = frames;
    const me = frames.find((f) => f.view.id === w.myId);
    if (me) {
      this.myReach.x = me.x;
      this.myReach.y = me.y;
      this.myReach.r = ringGeom(this.cfg, me.s.sc, me.s.n).reach;
      this.updateCamera(me.x, me.y, me.s.sc, this.myReach.r, now, dt);
    }
    autoQuality(this.fps, now); // máy yếu thì tự hạ chất lượng (spec §13.6 mục 8)
    this.applyEmergencyResolution();
    // Ngân sách bộ nhớ GPU (NETCODE_SPEC §5 L1): kiểm mỗi giây, không mỗi khung — con số
    // này đổi khi NẠP ẢNH, không đổi khi vẽ. Hạ chất lượng vì bộ nhớ là nhánh riêng, vì
    // trên iOS hết bộ nhớ đồ hoạ không làm tụt FPS trước: máy chạy 60 fps tới đúng lúc
    // Safari giết tab.
    if (now - this.memCheckAt >= 1) {
      this.memCheckAt = now;
      const mb = gpuMemMB();
      // Chạm trần thì NHẢ TEXTURE TRƯỚC, hạ chất lượng sau.
      //
      // Hạ chất lượng một mình KHÔNG trả lại bộ nhớ đã nạp — đo được trên iPhone thật ngày
      // 31/10: `texMB 214` ở mức "Vừa" rồi `216` ở mức "Thấp", tức là TĂNG. Nó chỉ làm ảnh
      // nạp SAU này nhỏ hơn, mà lúc đó tab đã sắp bị thu hồi.
      if (gpuPressure(mb) > 0) {
        const freed = assets.releaseIdleCharsNow();
        if (freed > 0) {
          console.warn(`chạm trần bộ nhớ đồ hoạ (${mb.toFixed(0)} MB) — nhả ${freed} atlas nhân vật`);
        }
      }
      if (autoQualityByMemory(mb)) {
        this.onMemoryPressure?.(mb);
      }
    }
    this.glow.apply(); // mức chất lượng vừa đổi thì gắn lại bloom (§3.6)
    sysRun('players', () => {
      this.syncPlayers(frames, now);
      probeSubFlush();
    });
    probeMark('players');
    sysRun('monsters', () => this.syncMonsters(now));
    probeMark('monsters');
    sysRun('items', () => {
      this.syncItems(now, me ? banksTienKhi(me.s.fl ?? 0) : false);
      this.updateFlyingItems(now);
    });
    probeMark('items');
    sysRun('beasts', () => this.updateBeasts(frames, now));
    probeMark('beasts');
    const myBush = me && me.s.fl & 2 ? this.bushAt(me.x, me.y) : -1;
    if (me) {
      sysRun('ambient', () => {
        const v = this.viewSize();
        this.ambient.update(me.x, me.y, v.w / 2, v.h / 2, now);
      });
    }
    probeMark('ambient');
    // Bản đồ KHÔNG nằm trong `sysRun`: tắt nó đi thì màn hình chỉ còn nền xám và game không
    // chơi được nữa — bỏ khung còn tử tế hơn. Lỗi ở đây rơi xuống lớp bắt lỗi của ticker.
    const vs = this.viewSize();
    this.mapView.ghostPass = this.ghostRunners(frames);
    this.mapView.update(w, now, myBush, me ? this.viewer(me.x, me.y, me.s.sc) : null, this.cfg.obstacles.occlusion_alpha,
      me ? { x: me.x, y: me.y, w: vs.w, h: vs.h } : undefined,
      me ? this.myReach : undefined);
    probeMark('map');
    sysRun('vfx', () => {
      this.vfx.update(now);
      // Tháo bloom khi lớp phát sáng rỗng (§13.10 mục 6) — xem `VfxLayer.tick`.
      this.glow.tick();
      this.floatText.update(now);
    });
    probeMark('vfx');
    sysRun('skills', () => {
      this.stallView.sync(w, {
        now,
        secs: this.cfg.skills.ranges_u.bay_sap_secs ?? 0,
        owner: (id) => {
          const c = this.casterAt(id);
          return c ? { x: c.x, y: c.y, peelR: skillRange(this.cfg, c.sc, 'vo_chuoi_w') } : null;
        },
      });
      this.skill.update(now);
    });
    probeMark('skills');
    probeHeap(now);
    this.takeCensus(now);
    // Trung bình trượt để số không nhảy loạn; gồm cả lượt vẽ của Pixi ở khung trước.
    this.lastRenderMs = this.lastRenderMs * 0.9 + (performance.now() - t0) * 0.1;
    this.sampleFps(dt);
    // Nhả atlas nhân vật không còn ai trong khu dùng (QUESTIONS Q30).
    //
    // Báo DANH SÁCH "ai đang trên màn hình" chứ không đếm tăng/giảm: đếm bằng tay thì chỉ
    // cần một đường thoát quên giảm (chết, qua cổng, despawn, đổi khu, mất kết nối) là con
    // số kẹt mãi trên 0 và không bao giờ nhả được.
    this.charsInUse.clear();
    for (const sp of this.players.values()) {
      if (sp.char) {
        this.charsInUse.add(sp.char);
      }
      // Cả atlas ĐANG VẼ: sprite dùng nhân vật dự phòng vẽ bằng atlas của người khác.
      if (sp.atlasChar) {
        this.charsInUse.add(sp.atlasChar);
      }
    }
    assets.keepChars(this.charsInUse, this.myChar, performance.now());
    drawCalls.frames += 1;
    if (drawCalls.frames >= 30) {
      drawCalls.perFrame = drawCalls.acc / drawCalls.frames;
      drawCalls.ready = drawCalls.on;
      drawCalls.acc = 0;
      drawCalls.frames = 0;
    }
  }

  /**
   * setResolution: đổi độ phân giải bộ đệm khung — CHỈ cho bài đo (tối ưu máy yếu, bước 0).
   *
   * Phép thử nghẽn CPU hay GPU: hạ còn 0,5 rồi xem FPS. Tăng mạnh = GPU đang là nút thắt
   * (ít pixel hơn thì nhanh hơn). Gần như không đổi = CPU, và lúc đó hạ độ phân giải không
   * cứu được gì — biết điều đó TRƯỚC khi tối ưu thì khỏi làm nhầm cả tuần.
   *
   * Không dùng trong luật chơi: `applyEmergencyResolution` mới là đường tự động.
   */
  setResolution(res: number): void {
    this.app.renderer.resize(this.app.screen.width, this.app.screen.height, res);
  }

  /**
   * setRings bật/tắt TOÀN BỘ vòng kiếm — CHỈ cho bài đo (`tools/browsertest/ringcost.mjs`).
   *
   * Vì sao cần: so FPS giữa hai lần chạy khác nhau là so hai CẢNH khác nhau. Lượt đo 12/11
   * cho thấy cùng một bài, cùng máy rảnh, số sprite trên màn hình đi từ 2 974 tới 4 813 tuỳ
   * lúc đó có mấy con bot đứng gần — biên độ ấy lớn hơn thứ đang muốn đo.
   *
   * Bật/tắt trong CÙNG một lượt thì phép trừ tự khử cảnh: `Pixi vẽ ms (bật) − (tắt)` là giá
   * của riêng vòng kiếm, đo trên đúng đám đông đang đứng đó. Đem con số ĐÓ so trước/sau mới
   * là so cùng một thứ.
   */
  /**
   * `mode`:
   *   - `'all'`  — như lúc chơi thật (mặc định khi trả lại).
   *   - `'mine'` — CHỈ vòng của mình. Cần cho bài so ảnh: vòng của người khác ra vào khung
   *     tuỳ bot đứng đâu, và chính nó là nguồn nhiễu lớn nhất — hai phiên đo 12/11 có 6 và
   *     13 người trong tầm, nên ảnh hiệu lệch 28,7/255 dù cùng một bản build.
   *   - `'none'` — tắt sạch, để lấy nền trừ đi.
   */
  setRings(mode: 'all' | 'mine' | 'none'): void {
    const myId = this.world?.myId;
    for (const [id, p] of this.players) {
      p.hideRings = mode === 'none' || (mode === 'mine' && id !== myId);
    }
    this.ringsMode = mode;
  }

  /**
   * layerStats: lô vẽ theo TỪNG LỚP của cảnh đang chơi — mục 5 lượt tối ưu (B30).
   *
   * Danh sách lớp lấy ĐÚNG thứ tự `camera.addChild` ở `rebuildLayers`. Thêm một lớp mới mà
   * quên thêm vào đây thì bảng vẫn ra, chỉ thiếu một dòng — nên khi tổng các dòng không bằng
   * `drawCalls.perFrame` thì nghĩa là danh sách đã lạc hậu, không phải phép đo sai.
   */
  layerStats(): { layers: LayerStat[]; total: number } {
    const mv = this.mapView;
    const layers: { name: string; c: { visible?: boolean } }[] = [
      // Tên lớp bằng ASCII, không tiếng Việt: đây là nhãn số liệu cho công cụ đo, không
      // phải chữ cho người chơi — cùng lý do `probe.ts` đặt khoá ASCII (quy tắc 5 chặn đúng
      // khi tôi viết tên tiếng Việt ở đây, và nó chặn đúng).
      ...(mv ? [{ name: 'ground', c: mv.ground }, { name: 'decor', c: mv.decor }, { name: 'zone-under', c: mv.under }] : []),
      { name: 'shadows', c: this.shadows },
      { name: 'orbs', c: this.orbs.root },
      { name: 'rings', c: this.rings },
      ...(mv ? [{ name: 'gates', c: mv.gateLayer }] : []),
      { name: 'glow', c: this.glow.root },
      { name: 'entities', c: this.entities },
      ...(mv ? [{ name: 'canopy', c: mv.over }] : []),
      { name: 'ambient', c: this.ambient.root },
      ...this.skill.layers.map((c, i) => ({ name: `skill-${i + 1}`, c })),
      { name: 'vfx', c: this.vfx.layer },
      { name: 'floattext', c: this.floatText.layer },
    ];
    const got = measureLayers(() => this.app.render(), layers);
    // Vẽ lại một khung ĐẦY ĐỦ rồi mới trả về: khung cuối của vòng đo chỉ có một lớp bật, để
    // nguyên là người chơi (hoặc ảnh chụp kế tiếp) thấy một cảnh khuyết.
    this.app.render();
    return { layers: got, total: got.reduce((a, c) => a + c.calls, 0) };
  }

  /** qualityName: mức chất lượng đang chạy — bài đo VRAM phải ghi nó vào báo cáo (§13.10). */
  qualityName(): string {
    return quality();
  }

  /** setProbe bật/tắt đồng hồ theo hệ thống; `main` gọi theo trạng thái bảng F3. */
  setProbe(on: boolean): void {
    probeEnabled(on);
  }

  /** fpsNow: FPS đang đo (trung bình trượt) — cho bài đo tự động. */
  fpsNow(): number {
    return this.fps;
  }

  /** probe trả số liệu thô cho bài đo tự động (`window.kx.scene.probe()`). */
  probe(): Record<string, number> {
    return probeStats();
  }

  /** paceFrames: vẽ đúng mỗi `n` khung quét màn hình (1 = mọi khung) — xem framepace.ts. */
  paceFrames(n: number): void {
    // Nhớ lại nhịp: watchdog và `rebuild` phải dựng lại đúng nhịp này, không thì máy 144 Hz
    // vừa dựng lại renderer xong là chạy 144 khung/giây.
    this.pace = n;
    paceTicker(this.app.ticker, n);
  }

  /** toggleHitbox shows the server's sword capsules over the drawn swords (debug key H). */
  toggleHitbox(): void {
    this.showHitbox = !this.showHitbox;
  }

  private viewer(x: number, y: number, scale: number): Viewer {
    const r = this.cfg.player.body_r * scale;
    return { x, top: y - r * 1.9, feet: y + r * 0.75, halfW: r * 0.8 };
  }

  private bushAt(x: number, y: number): number {
    for (const z of this.world?.map?.zones ?? []) {
      if (z.k !== 'bush') {
        continue;
      }
      const s = z.s;
      const inside = s.c ? (x - s.x) ** 2 + (y - s.y) ** 2 <= (s.r ?? 0) ** 2 : x >= s.x && x <= s.x + (s.w ?? 0) && y >= s.y && y <= s.y + (s.h ?? 0);
      if (inside) {
        return z.id;
      }
    }
    return -1;
  }

  private updateCamera(x: number, y: number, scale: number, reach: number, now: number, dt: number): void {
    // Smoothly follow the camera height so growing does not pop.
    const targetH = cameraViewHeight(this.cfg, scale, reach);
    this.viewH += (targetH - this.viewH) * 0.08;
    this.zoom = this.app.screen.height / this.viewH;
    const c = this.followCamera(x, y, dt);
    const shake = this.vfx.shakeOffset(now);
    this.camera.scale.set(this.zoom);
    this.camera.position.set(this.app.screen.width / 2 - (c.x + shake.x) * this.zoom, this.app.screen.height / 2 - (c.y + shake.y) * this.zoom);
  }

  /**
   * followCamera: máy quay ĐUỔI theo nhân vật (giảm dần theo hàm mũ, `camera_follow_ms`)
   * thay vì dính cứng. Dính cứng thì mọi cú kéo vị trí nhỏ (server sửa khi trượt dọc vật
   * cản, nhịp gửi phím lệch vài ms) làm CẢ màn hình giật theo — chủ dự án 05/11. Hằng số
   * nhỏ (~70 ms) nên đi thẳng chỉ lệch tâm vài đơn vị, mắt không thấy trễ.
   * Nhảy xa (cổng, hồi sinh, đổi khu) thì đặt thẳng: lướt nửa bản đồ là chóng mặt.
   */
  private followCamera(x: number, y: number, dt: number): { x: number; y: number } {
    const tau = (this.cfg.player.camera_follow_ms ?? 0) / 1000;
    const c = this.camAt;
    if (!c || tau <= 0 || Math.hypot(x - c.x, y - c.y) > this.viewH * 0.5) {
      this.camAt = { x, y };
      return this.camAt;
    }
    const k = 1 - Math.exp(-Math.max(dt, 0) / tau);
    c.x += (x - c.x) * k;
    c.y += (y - c.y) * k;
    return c;
  }

  /**
   * assetGen: `assets.gen` lần kiểm gần nhất.
   *
   * Nhóm `chars` / `zone_*` tải NỀN (§13.7 mục 3), nên atlas có thể về giữa trận. Mốc này
   * là cách biết "vừa có ảnh mới" mà không phải thử `setCharacter` cho từng nhân vật ở mỗi
   * khung — 60 lần/giây × 60 nhân vật tra bảng cho một việc gần như không bao giờ xảy ra.
   */
  private assetGen = -1;

  private syncPlayers(frames: ReturnType<ClientWorld['playerFrames']>, now: number): void {
    const seen = new Set<number>();
    const newArt = this.assetGen !== assets.gen;
    this.assetGen = assets.gen;
    const detail = this.swordDetail(frames);
    const me = frames.find((f) => f.view.id === this.world?.myId);
    const myRealm = me ? realmOf(this.cfg, me.s.L) : 0;
    for (const f of frames) {
      seen.add(f.view.id);
      let sp = this.players.get(f.view.id);
      // Đổi nhân vật thì đổi TẠI CHỖ; chỉ dựng view mới khi thật sự chưa có (xem
      // `PlayerSprite.setCharacter`). Bản cũ huỷ rồi dựng lại, thêm một lần nhấp nháy.
      if (sp && f.view.char && (sp.char !== f.view.char || (newArt && sp.needsArtUpgrade()))) {
        sp.setCharacter(f.view.char);
      }
      if (!sp) {
        // Lấy từ pool trước khi dựng mới: dựng một `PlayerSprite` là tạo cả chục đối tượng
        // PixiJS (thân, vòng kiếm, bảng tên, bóng, hào quang). Làm việc đó giữa trận, mỗi
        // lần có người vào khung hình, là nguồn giật rác bộ nhớ — §5 nói rõ "không tạo/hủy
        // object trong lúc chơi".
        sp = this.pool.take(this.cfg, f.view.id, f.view.name, f.view.char, f.view.id === this.world?.myId);
        // Người vào khung hình GIỮA pha đo phải theo đúng chế độ đang đo, không thì phép
        // trừ của `ringcost.mjs` lẫn một người có vòng kiếm vào giữa pha "tắt".
        sp.hideRings = this.ringsMode === 'none'
          || (this.ringsMode === 'mine' && f.view.id !== this.world?.myId);
        this.players.set(f.view.id, sp);
        this.entities.addChild(sp.root);
        this.shadows.addChild(sp.shadow);
        this.rings.addChild(sp.ring);
        // Nhân vật vừa xuất hiện cũng phải bật sẵn lớp debug nếu phím J đang bật — không thì
        // đúng những con mới vào khung (chỗ hay giật nhất) lại là chỗ không có số.
        if (this.animInfoOn) {
          sp.showAnimInfo(true);
        }
      }
      sp.update(f, this.zoom, now, this.showHitbox, detail, myRealm, this.plateFade(f, frames));
      this.syncAura(f, now);
      this.moodChatter(f.view.id, moodName(f.s.md), now);
    }
    // Số người đang vẽ — luật tắt bong bóng mặt ở chất lượng Thấp (spec §12.5) đọc từ đây.
    // Đặt sau vòng vẽ để nó là con số của ĐÚNG khung vừa vẽ, không phải khung trước.
    setEmoteCharCount(this.players.size);
    // Rời khung hình: MỜ DẦN RA 150 ms rồi mới về pool, không biến mất đột ngột
    // (NETCODE_SPEC §5 Hiển thị mục 4). Huỷ hẳn sprite cũng bị bỏ: view vào pool, giữ
    // texture, lần sau lấy ra dùng lại (§5 "View Pool").
    for (const [id, sp] of this.players) {
      if (seen.has(id)) {
        this.leaving.delete(id);
        continue;
      }
      const since = this.leaving.get(id);
      if (since === undefined) {
        this.leaving.set(id, now);
        continue;
      }
      const t = (now - since) / FADE_OUT_SECS;
      if (t < 1) {
        sp.setLeaveFade(1 - t);
        continue;
      }
      this.leaving.delete(id);
      this.players.delete(id);
      this.shadows.removeChild(sp.shadow);
      this.pool.put(sp);
    }
    for (const [id, a] of this.auras) {
      if (!seen.has(id)) {
        a.destroy();
        this.auras.delete(id);
      }
    }
  }

  /**
   * plateFade làm mờ bảng tên của người Ở XA khi hai bảng chồng lên nhau, để cụm gần mình
   * luôn đọc được (spec §12.3 declutter).
   */
  private plateFade(f: ReturnType<ClientWorld['playerFrames']>[number], all: ReturnType<ClientWorld['playerFrames']>): number {
    const me = all.find((p) => p.view.id === this.world?.myId);
    if (!me || f.view.id === me.view.id) {
      return 1;
    }
    const mine = Math.hypot(f.x - me.x, f.y - me.y);
    for (const o of all) {
      if (o.view.id === f.view.id || o.view.id === me.view.id) {
        continue;
      }
      // Hai bảng coi là chồng nhau khi gần nhau theo chiều ngang và sàn sàn chiều cao.
      if (Math.abs(o.x - f.x) < 90 && Math.abs(o.y - f.y) < 50 && Math.hypot(o.x - me.x, o.y - me.y) < mine) {
        return 0.5;
      }
    }
    return 1;
  }

  /** syncAura dựng / cập nhật hào quang Tiên giới trong lớp phát sáng dùng chung (§3.6). */
  private syncAura(f: ReturnType<ClientWorld['playerFrames']>[number], now: number): void {
    const idx = tienIndex(this.cfg, realmOf(this.cfg, f.s.L));
    let a = this.auras.get(f.view.id);
    if (idx < 0) {
      if (a) {
        a.destroy();
        this.auras.delete(f.view.id);
      }
      return;
    }
    if (!a) {
      a = new TienAura(this.cfg);
      this.auras.set(f.view.id, a);
      this.glow.root.addChild(a.root);
    }
    a.setRealm(idx);
    const r = this.cfg.player.body_r * f.s.sc;
    // Ẩn Thân / trong bụi: mờ theo chủ (spec §3.4).
    const hidden = (f.s.fl & FLAG.stealth) !== 0 || (f.s.fl & FLAG.bush) !== 0;
    a.update(f.x, f.y + r * 0.75, r, now, hidden ? 0.2 : 1, r * 2.6 * this.zoom >= AURA_FAR_PX);
  }

  private moodChatter(id: number, mood: string, now: number): void {
    if (this.lastMood.get(id) === mood) {
      return;
    }
    this.lastMood.set(id, mood);
    const lines = STR.emotes[mood];
    if (lines && Math.random() < MOOD_EMOTE_CHANCE) {
      this.players.get(id)?.showBubble(pick(lines), now);
    }
  }

  private syncMonsters(now: number): void {
    const seen = new Set<number>();
    for (const m of this.world?.monsterFrames() ?? []) {
      seen.add(m.s.id);
      let sp = this.monsters.get(m.s.id);
      if (!sp) {
        sp = new MonsterSprite(m.s);
        this.monsters.set(m.s.id, sp);
        this.entities.addChild(sp.root);
      }
      sp.update(m.x, m.y, m.s, now);
    }
    for (const [id, sp] of this.monsters) {
      if (!seen.has(id)) {
        sp.destroy();
        this.monsters.delete(id);
      }
    }
  }

  private syncItems(now: number, meBanksTienKhi: boolean): void {
    const items = this.world?.items;
    if (!items) {
      return;
    }
    for (const [id, it] of items) {
      if (it.k === 'orb') {
        this.orbs.add(it);
        continue;
      }
      if (!this.items.has(id)) {
        const sp = new ItemSprite(it, this.cfg.items.item_r);
        this.items.set(id, sp);
        this.entities.addChild(sp.root);
      }
    }
    for (const [id, sp] of this.items) {
      if (!items.has(id)) {
        this.items.delete(id);
        const by = this.pickedBy.get(id);
        this.pickedBy.delete(id);
        if (by === undefined) {
          sp.destroy(); // hết hạn / ra khỏi tầm nhìn: biến mất luôn
        } else {
          // Bị nhặt → bay hút về phía người nhặt thay vì biến mất đột ngột.
          this.flying.push({ sp, by, from: { x: sp.root.x, y: sp.root.y }, at: now });
        }
        continue;
      }
      sp.update(now, absorbable(sp.it.k, meBanksTienKhi), this.zoom);
    }
    this.syncOrbs(items, now, meBanksTienKhi);
  }

  /** syncOrbs dọn viên EXP đã biến mất và cho viên bị nhặt bay hút vào người nhặt. */
  private syncOrbs(items: Map<number, ItemState>, now: number, meBanksTienKhi: boolean): void {
    for (const id of this.orbs.ids()) {
      if (items.has(id)) {
        continue;
      }
      const by = this.pickedBy.get(id);
      const at = this.orbs.at(id);
      this.pickedBy.delete(id);
      if (by !== undefined && at) {
        const ghost = new ItemSprite({ ...(this.lastOrb.get(id) as ItemState), x: at.x, y: at.y }, this.cfg.items.item_r);
        this.entities.addChild(ghost.root);
        this.flying.push({ sp: ghost, by, from: at, at: now });
      }
      this.lastOrb.delete(id);
      this.orbs.remove(id);
    }
    for (const [id, it] of items) {
      if (it.k === 'orb' && !this.lastOrb.has(id)) {
        this.lastOrb.set(id, it);
      }
    }
    this.orbs.update(now, absorbable('orb', meBanksTienKhi));
  }

  /** updateFlyingItems tweens picked-up items into their owner, then drops them. */
  private updateFlyingItems(now: number): void {
    for (let i = this.flying.length - 1; i >= 0; i--) {
      const f = this.flying[i];
      const t = Math.min(1, (now - f.at) / SUCK_SECS);
      const owner = this.players.get(f.by)?.root;
      const tx = owner?.x ?? f.from.x;
      const ty = owner?.y ?? f.from.y;
      // Ease-in: lề mề lúc đầu rồi lao vút vào người, giống bị hút.
      const k = t * t;
      f.sp.root.position.set(f.from.x + (tx - f.from.x) * k, f.from.y + (ty - f.from.y) * k);
      f.sp.root.scale.set(1 - 0.7 * t);
      f.sp.root.alpha = 1 - t * t;
      if (t >= 1) {
        f.sp.destroy();
        this.flying.splice(i, 1);
      }
    }
  }

  // ---- events → VFX, sound, bubbles ---------------------------------------

  handleEvents(events: GameEvent[]): void {
    const now = performance.now() / 1000;
    for (const e of events) {
      this.handleEvent(e, now);
    }
  }

  private playerScale(id: number | undefined): number {
    const v = id !== undefined ? this.world?.players.get(id) : undefined;
    return v ? v.snaps[v.snaps.length - 1].s.sc : 1;
  }

  /**
   * swordDetail chọn mức giãn chung cho cả màn hình: 1 = vẽ đủ, 2 = vẽ cách một thanh…
   * Dùng bậc thang 1/2/3/4 kèm mép trễ để số kiếm dao động quanh ngưỡng không làm dựng lại
   * vòng kiếm liên tục.
   */
  private swordDetail(frames: ReturnType<ClientWorld['playerFrames']>): number {
    let total = 0;
    for (const f of frames) {
      total += f.s.n;
    }
    const want = Math.ceil(total / SWORD_BUDGET);
    if (want > this.detail) {
      this.detail = want; // quá tải thì giãn ngay
    } else if (want <= this.detail - 1 && total < SWORD_BUDGET * (this.detail - 1) * 0.85) {
      this.detail -= 1; // nhẹ hẳn rồi mới dày lại, tránh dựng lại liên tục
    }
    return Math.max(1, Math.min(4, this.detail));
  }

  /**
   * perfLine là dòng số liệu cho đồng hồ debug: khung hình, thời gian vẽ, và số thứ đang vẽ.
   * Có nó thì lúc ai đó kêu lag là biết ngay nghẽn ở đâu — khung hình thấp mà `vẽ` nhỏ là
   * nghẽn mạng chứ không phải máy yếu, còn `vẽ` cao là do quá nhiều thứ trên màn hình.
   */
  /**
   * drawStats là những con số chỉ đếm được bằng cách duyệt các nhân vật ĐANG VẼ
   * (TELEMETRY_SPEC §2.2). Để ở đây chứ không để `players` lộ ra ngoài: vòng lặp nằm cạnh
   * dữ liệu, và không ai ngoài scene chạm được vào bảng sprite.
   */
  /** showAnimInfo bật/tắt lớp chữ debug của phím `J` trên MỌI nhân vật đang vẽ. */
  showAnimInfo(on: boolean): void {
    this.animInfoOn = on;
    if (on) {
      this.animViewer.show(this.app.screen.width, this.app.screen.height);
    } else {
      this.animViewer.hide();
    }
    for (const p of this.players.values()) {
      p.showAnimInfo(on);
    }
  }

  /** animInfoOn: giữ trạng thái để nhân vật MỚI xuất hiện cũng bật sẵn, không phải bấm lại. */
  private animInfoOn = false;
  /** leaving[id] = lúc entity bắt đầu mờ ra. Xem vòng dọn ở cuối `drawPlayers`. */
  private readonly leaving = new Map<number, number>();
  /** pool: view đã rời khung hình, giữ nguyên texture để dùng lại (§5 "View Pool"). */
  private readonly pool = new ViewPool();
  /** animViewer: khung xem từng animation của chế độ `J` (phím `[` / `]`). */
  readonly animViewer = new AnimViewer();

  /** stepAnimViewer đi tới / lùi một animation trong bảng điểm. */
  stepAnimViewer(d: number): void {
    this.animViewer.step(d);
  }

  drawStats(): { charsInView: number; swordsDrawn: number; placeholder: boolean } {
    let swordsDrawn = 0;
    let placeholder = false;
    for (const p of this.players.values()) {
      swordsDrawn += p.swordCount;
      if (!p.hasAtlas()) {
        placeholder = true;
      }
    }
    return { charsInView: this.players.size, swordsDrawn, placeholder };
  }

  /** loadReport: định dạng asset đã nạp — bằng chứng cho cổng kiểm §13.10. */
  loadReport(): { ktxOn: boolean; ktx2: number; webp: number; webpWhenKtx: string[] } {
    return assets.loadReport();
  }

  /** gpuBiggest: vài texture to nhất đang giữ — để biết cắt chỗ nào trước (debug). */
  gpuBiggest(n = 10): { w: number; h: number; mb: number }[] {
    return gpuBiggest(n);
  }

  /** gpuSites: texture đang giữ gom theo nơi cấp phát — cần mở trang với `?textrace=1`. */
  gpuSites(n = 12): { site: string; count: number; mb: number }[] {
    return gpuSites(n);
  }

  /**
   * vramLedger: sổ VRAM theo hạng mục ngân sách (GAME_SPEC §13.10).
   *
   * Trả cả dòng `other` kể cả khi nó bằng 0: một bảng chỉ hiện hạng mục có số thì ngày nào
   * đó `other` xuất hiện rồi biến mất mà không ai kịp thấy.
   */
  vramLedger(): { rows: { cat: BudgetCat; label: string; mb: number; count: number; limitMB: number; over: boolean }[]; totalMB: number; limitMB: number; unclassified: { site: string; mb: number }[]; framebuffers: { size: string; count: number; mb: number; site: string }[]; world: { size: string; count: number; mb: number; site: string }[]; fx: { size: string; count: number; mb: number; site: string }[] } {
    const phone = isMobile();
    const rows = gpuBudget(phone);
    return {
      rows,
      totalMB: +rows.reduce((a, r) => a + r.mb, 0).toFixed(2),
      limitMB: phone ? 32 : 50,
      unclassified: gpuUnclassified(),
      framebuffers: gpuListCat('framebuffer'),
      world: gpuListCat('world', 14),
      fx: gpuListCat('fx', 14),
    };
  }

  /** perfStats là cùng bộ số của overlay F3 nhưng ở dạng object (tools/browsertest đọc). */
  perfStats(): Record<string, number> {
    let swords = 0;
    for (const p of this.players.values()) {
      swords += p.swordCount;
    }
    const me = this.world?.me();
    return {
      myLevel: me?.L ?? -1,
      mySwords: me?.n ?? 0,
      fps: this.fps,
      frameMs: 1000 / Math.max(this.fps, 1),
      drawMs: this.lastRenderMs,
      // pixiMs: lượt vẽ của Pixi, tách khỏi vòng cập nhật (xem khai báo `pixiMs`).
      pixiMs: this.pixiMs,
      sprites: countNodes(this.app.stage),
      drawCalls: drawCalls.perFrame,
      texMB: gpuMemMB(),
      // texBudgetMB: ngân sách bộ nhớ đồ hoạ của THIẾT BỊ NÀY (device.ts: iOS 256 · điện
      // thoại 384 · máy bàn 1024). Bài kiểm rò rỉ §13.9 chặn khi vượt ngân sách, và nó
      // phải đọc con số của chính máy đang chạy — chép tay một con số vào bài đo là sai
      // ngay lần đầu chạy trên máy khác.
      texBudgetMB: gpuBudgetMB(),
      texPeakMB: gpuPeakMB(),
      textures: gpuTextureCount(),
      fixPerSec: this.world?.fixPerSec ?? 0,
      fixMax: this.world?.fixMax ?? 0,
      players: this.players.size,
      swords,
      // ringWrites: số lần đặt lại một thanh kiếm trong khung vừa rồi (swordring.ts). Đây
      // là bộ đếm CÔNG VIỆC — không nở ra co lại theo tải máy như mili-giây.
      ringWrites: ringWork.frame,
      ringWouldWrite: ringWork.wouldFrame,
      statusDraws: statusWork.frameDraws,
      statusCalls: statusWork.frameCalls,
      monsters: this.monsters.size,
      orbs: this.orbs.ids().length,
      items: this.items.size,
      aoi: this.players.size + this.monsters.size + this.items.size + this.orbs.ids().length,
      detail: this.detail,
      // Sổ tài nguyên §13.9: số đối tượng đang sống theo loại. Chốt 1 lần/giây chứ không
      // đếm lại ở đây — `perfStats()` được gọi 4 lần/giây từ telemetry.
      ...censusFlat(this.lastCensus),
    };
  }

  /**
   * takeCensus chốt sổ tài nguyên (§13.9). Gọi 1 lần/giây từ vòng vẽ.
   *
   * Duyệt cả cây cảnh nên KHÔNG chạy mỗi khung — ở chỗ đông cây có vài nghìn nút. Một
   * giây là đủ dày cho một chỉ số đo bằng phút.
   */
  private takeCensus(now: number): void {
    if (now - this.censusAt < 1) {
      return;
    }
    this.censusAt = now;
    // Pool là chỗ dễ biến thành rò rỉ nhất: nó giữ tham chiếu CÓ CHỦ Ý nên bộ dọn rác
    // không bao giờ đụng tới. Khai ra để nó nằm cạnh các con số khác.
    ledgerSet('pool', this.pool.size);
    ledgerSet('poolText', this.floatText.poolSize);
    ledgerSet('beasts', this.beasts.length);
    ledgerSet('auras', this.auras.size);
    ledgerSet('flying', this.flying.length);
    this.lastCensus = census(this.app.stage as unknown as Parameters<typeof census>[0]);
  }

  /**
   * sampleFps đếm khung hình — gọi TỪ VÒNG VẼ, mỗi khung, với mọi người chơi.
   *
   * Trước đây nó nằm trong `perfLine`, mà `perfLine` chỉ chạy khi mở overlay F3. Hậu quả:
   * `perfStats().fps` LUÔN bằng 0 với người chơi thường — và `perfStats()` chính là thứ
   * đính kèm vào sự kiện `gpu_pressure` gửi về server. Nên bằng chứng từ chiếc iPhone thật
   * ngày 31/10 ghi `"fps":0` trong lúc máy đang tụt khung: đúng con số cần nhất thì trống.
   *
   * Đo bằng ĐỒNG HỒ THẬT, không dùng `dt` của ticker: Pixi kẹp `deltaMS` ở minFPS=10, nên
   * một máy chạy 3 khung/giây vẫn báo đúng 10.0. Bài test 10 tab headless từng cho ra cột
   * FPS toàn 10.0 / 10.0 / 10.0 vì đúng lỗi này.
   */
  private sampleFps(dt: number): void {
    const t = performance.now() / 1000;
    this.fpsAcc += this.fpsLast > 0 ? t - this.fpsLast : dt;
    this.fpsLast = t;
    this.fpsFrames += 1;
    if (this.fpsAcc >= 0.5) {
      this.fps = this.fpsFrames / this.fpsAcc;
      this.fpsAcc = 0;
      this.fpsFrames = 0;
    }
  }

  /** ensureDrawCount gắn bộ đếm draw call lần đầu overlay F3 cần tới nó. */
  private ensureDrawCount(): void {
    const gl = (this.app.renderer as unknown as { gl?: WebGL2RenderingContext }).gl;
    if (gl) {
      countDrawCalls(gl);
    }
  }

  perfLine(dt: number, net: { rttMs: number; kbPerSec: number; snapBytes: number; srvMs: number }): string {
    void dt; // FPS nay đếm trong vòng vẽ (xem sampleFps), không đếm ở đây nữa
    this.ensureDrawCount();
    let swords = 0;
    for (const p of this.players.values()) {
      swords += p.swordCount;
    }
    const aoi = this.players.size + this.monsters.size + this.items.size + this.orbs.ids().length;
    return [
      `${this.fps.toFixed(0)} fps · ${(1000 / Math.max(this.fps, 1)).toFixed(1)} ms/khung · vẽ ${this.lastRenderMs.toFixed(1)} ms`, // i18n-ok: lớp số liệu debug
      `server ${net.srvMs.toFixed(2)} ms/tick · ping ${net.rttMs.toFixed(0)} ms`,
      `mạng ${net.kbPerSec.toFixed(1)} KB/s · snapshot ${(net.snapBytes / 1024).toFixed(1)} KB`, // i18n-ok: lớp số liệu debug
      `sprite ${countNodes(this.app.stage)} · draw call ${drawCalls.ready ? drawCalls.perFrame.toFixed(0) : '—'}`,
      `AOI ${aoi} · người ${this.players.size} · kiếm ${swords} · quái ${this.monsters.size}`, // i18n-ok: lớp số liệu debug
      `EXP ${this.orbs.ids().length} · vật phẩm ${this.items.size} · giãn kiếm 1/${this.detail}`, // i18n-ok: lớp số liệu debug
      // Spec §8.2a: số lần server kéo nhân vật về chỗ khác. Dự đoán đúng thì ~0; lệch thì
      // nó chạy lên bằng tần số snapshot và người chơi thấy rung khi đi sát vật cản.
      `hiệu chỉnh ${(this.world?.fixPerSec ?? 0).toFixed(1)}/giây · lệch nhất ${(this.world?.fixMax ?? 0).toFixed(1)}`, // i18n-ok: lớp số liệu debug
      ...probeLines(), // i18n-ok: lớp số liệu debug
      // Lớp chống sập §13.8: hệ thống nào đã phải tắt, đã dựng lại renderer mấy lần. Dòng
      // này RỖNG khi mọi thứ bình thường — nó chỉ hiện khi có chuyện.
      ...guardLines(this.rebuilds, this.watchdogRestarts),
      // Sổ tài nguyên §13.9 — con số phải ĐỨNG YÊN qua nhiều phút chơi.
      censusLine(this.lastCensus),
      // Sổ VRAM theo hạng mục ngân sách §13.10 — chỉ in hạng mục ĐANG VƯỢT, kèm tổng.
      // In đủ bảy dòng mọi lúc thì bảng F3 dài thêm một màn hình và không ai đọc nữa;
      // in khi vượt thì dòng này im lặng lúc mọi thứ ổn và kêu đúng lúc cần.
      vramLine(this.vramLedger()),
    ].join('\n');
  }

  /** guardStats: số liệu lớp chống sập, gắn kèm mọi báo cáo lỗi (§13.8). */
  guardStats(): { resizes: number; rebuilds: number; watchdog: number; sysOff: string } {
    return { resizes: this.resizes, rebuilds: this.rebuilds, watchdog: this.watchdogRestarts, sysOff: sysOffList() };
  }

  /** beastCount: số thần thú đang chạy, theo kiểu — dùng để kiểm tra bằng console debug. */
  beastCount(): Record<string, number> {
    const n: Record<string, number> = { breakthrough: 0, ambient: 0, shatter: 0 };
    for (const e of this.beasts) {
      n[e.b.mode] += 1;
    }
    return n;
  }

  /**
   * updateBeasts chạy toàn bộ thần thú Tiên giới mỗi khung hình: hẹn giờ bản thường trực,
   * đặt lại vị trí theo chủ, và xếp hai lớp ôm lấy nhân vật đúng thứ tự chiều sâu.
   */
  private updateBeasts(frames: ReturnType<ClientWorld['playerFrames']>, now: number): void {
    const byId = new Map(frames.map((f) => [f.view.id, f]));
    this.lastTien.clear();
    const me = frames.find((f) => f.view.id === this.world?.myId);
    for (const f of frames) {
      const idx = tienIndex(this.cfg, realmOf(this.cfg, f.s.L));
      this.lastTien.set(f.view.id, idx);
      if (idx < 0) {
        continue;
      }
      // Mỗi người Tiên giới luôn có thần thú đi kèm; Đạo Tổ có hai con lệch pha 180°.
      const want = TIEN_BEASTS[idx]?.twin ? 2 : 1;
      const have = this.beasts.filter((e) => e.owner === f.view.id && e.b.mode === 'companion').length;
      for (let k = have; k < want; k++) {
        this.spawnBeast(f.view.id, 'companion', now, idx, k * Math.PI);
      }
    }
    // Vệt hạt chỉ cho vài con gần mình nhất (spec §3.5): xếp theo khoảng cách rồi bật dần.
    const dist = (id: number): number => {
      const f = byId.get(id);
      return f && me ? Math.hypot(f.x - me.x, f.y - me.y) : Infinity;
    };
    const order = [...this.beasts].sort((a, b) => dist(a.owner) - dist(b.owner));
    order.forEach((e, i) => {
      e.b.trailOn = i < BEAST_TRAIL_NEAR;
    });

    this.beasts = this.beasts.filter((e) => {
      const f = byId.get(e.owner);
      const stillTien = f ? tienIndex(this.cfg, realmOf(this.cfg, f.s.L)) >= 0 : false;
      // Bản companion sống tới khi chủ rời Tiên giới hoặc ra khỏi tầm nhìn.
      if (!f || e.b.done(now) || (e.b.mode === 'companion' && !stillTien)) {
        this.removeBeast(e.b);
        return false;
      }
      const r = this.cfg.player.body_r * f.s.sc;
      const feet = f.y + r * 0.75; // chân nhân vật, khớp FEET_BELOW_CENTER của PlayerSprite
      e.b.resize(r, r * 2.6);
      // Ẩn Thân / trong bụi: mờ theo chủ (spec §3.5), không ẩn hẳn.
      const hidden = (f.s.fl & FLAG.stealth) !== 0 || (f.s.fl & FLAG.bush) !== 0;
      e.b.ownerAlpha = hidden ? 0.2 : 1;
      e.b.burning = (f.s.fl & FLAG.burning) !== 0;
      e.b.update(f.x, feet, now, r * 2.6 * this.zoom >= BEAST_FAR_PX);
      // Lớp sau nằm ngay dưới nhân vật, lớp trước ngay trên — nhân vật zIndex = y + r×0.75.
      e.b.back.zIndex = feet - 0.5;
      e.b.front.zIndex = feet + 0.5;
      return true;
    });
  }

  /** flashBeasts cho thần thú của một người lóe sáng + lượn nhanh (spec §3.5 v4.7). */
  private flashBeasts(owner: number, now: number): void {
    for (const e of this.beasts) {
      if (e.owner === owner && e.b.mode === 'companion') {
        e.b.flash(now);
      }
    }
  }

  /** spawnBeast tạo thần thú cho một người, tôn trọng trần 6 con của spec §3.5. */
  private spawnBeast(owner: number, mode: BeastMode, now: number, realmIdx?: number, phase = 0): void {
    const f = this.lastFrames.find((p) => p.view.id === owner);
    if (!f) {
      return;
    }
    const idx = realmIdx ?? tienIndex(this.cfg, realmOf(this.cfg, f.s.L));
    const spec = TIEN_BEASTS[idx];
    if (!spec) {
      return;
    }
    const r = this.cfg.player.body_r * f.s.sc;
    const b = new Beast(spec, mode, r, r * 2.6, now, phase);
    this.entities.addChild(b.back, b.front);
    this.beasts.push({ b, owner });
  }

  private removeBeast(b: Beast): void {
    this.entities.removeChild(b.back, b.front);
    b.destroy();
  }

  private handleEvent(e: GameEvent, now: number): void {
    if (this.skill.handle(e, now)) {
      return; // kỹ năng + trạng thái: SkillPresenter lo hết
    }
    const mine = e.a === this.world?.myId;
    const size = 90 * this.playerScale(e.a);
    // Hạ được người → bong bóng mặt cười (spec §12.5 v6.38).
    //
    // `KillFeed` còn dùng cho đột phá / qua Thiên Môn / lên danh hiệu — những cái đó KHÔNG
    // có nạn nhân, nên `e.b` là thứ phân biệt. Thiếu phép thử này thì ai đột phá cũng cười
    // nham hiểm một cái.
    if (e.k === 'KillFeed' && e.b) {
      this.players.get(e.a ?? 0)?.onKill(now);
    }
    switch (e.k) {
      case 'RingFlip':
        // Both rings emit RingFlip at the same contact point; show one spark.
        if (now - this.lastSpark.t > 0.05 || Math.hypot(e.x - this.lastSpark.x, e.y - this.lastSpark.y) > 1) {
          this.lastSpark = { t: now, x: e.x, y: e.y };
          this.vfx.play('clash_spark', e.x, e.y, size * 1.2, now);
          sound.play('sword_clash', mine ? 1 : 0.5);
        }
        break;
      case 'SwordBreak': {
        // Server gộp số kiếm gãy lại rồi mới báo (xem flushSwordBreaks), nên một sự kiện có
        // thể là cả chục cây — vẽ một vụ nổ to theo số lượng thay vì chục vụ nổ nhỏ.
        const many = Math.min(3, 1 + Math.log2(Math.max(1, e.n ?? 1)) / 2);
        this.vfx.play('sword_shatter', e.x, e.y, size * 1.5 * many, now);
        sound.play('sword_break', mine ? 1 : 0.5);
        break;
      }
      case 'Hit':
        if (e.s === 'shield') {
          this.flashBeasts(e.b ?? 0, now); // khiên đỡ được một đòn → thần thú lóe sáng
        }
        this.vfx.play(e.s === 'shield' ? 'shield_break' : 'stun_stars', e.x, e.y - size * 0.6, size * 1.6, now);
        sound.play(e.s === 'shield' ? 'shield_break' : 'sword_hit_body', e.b === this.world?.myId || mine ? 1 : 0.5);
        if (e.b === this.world?.myId) {
          this.vfx.shake(18 * this.playerScale(e.b), 0.25, now);
        }
        break;
      case 'Death':
        this.vfx.play('death_smoke', e.x, e.y, size * 2, now);
        sound.play('death');
        break;
      case 'LevelUp':
        this.vfx.play('levelup_aura', e.x, e.y, size * 2, now);
        if (mine) {
          sound.play('levelup');
        }
        break;
      case 'Breakthrough':
        this.onBreakthrough(e, size, mine, now);
        break;
      case 'Blinded':
        // Server chỉ gửi event này cho ĐÚNG người bị trúng (spec §6.3), nên không cần lọc
        // thêm; thời gian lấy thẳng từ server để client không tự đoán.
        // Thời gian lấy THẲNG từ server; chỉ khi gói thiếu mới rơi về số trong config.
        // Tự đoán ở đây là client hết trân châu trong khi server vẫn tính là đang bị bịt
        // mắt — người chơi nhìn rõ nhưng vẫn không nhặt được gì, không hiểu vì sao.
        this.boba.hit(e.f ?? this.cfg.skills.tran_chau_secs, now, this.app.screen.width, this.app.screen.height);
        sound.play('sword_hit_body');
        break;
      case 'RealmDrop':
        {
          const was = this.lastTien.get(e.a ?? 0) ?? -1;
          if (was >= 0) {
            this.spawnBeast(e.a ?? 0, 'shatter', now, was); // thần thú gãy vụn (spec §3.5)
          }
        }
        if (mine) {
          sound.play('realm_drop');
        }
        break;
      default:
        this.handleWorldEvent(e, now, mine, size);
    }
  }

  /**
   * casterAt lấy vị trí + hướng mặt + scale của người dùng kỹ năng, từ SNAPSHOT.
   *
   * Lấy từ snapshot chứ không hỏi `PlayerSprite`: sprite chỉ giữ trạng thái vẽ (lật
   * trái/phải), không giữ góc mặt thật — mà nón kỹ năng cần đúng góc server đã xét.
   */
  private casterAt(id: number): { x: number; y: number; f: number; sc: number } | null {
    for (const fr of this.lastFrames) {
      if (fr.view.id === id) {
        return { x: fr.x, y: fr.y, f: fr.s.f, sc: fr.s.sc };
      }
    }
    return null;
  }

  private onBreakthrough(e: GameEvent, size: number, mine: boolean, now: number): void {
    if (e.s === 'tien') {
      // Phi thăng Tiên giới: thần thú to chui lên từ chân, ai trong tầm nhìn đều thấy (§3.5).
      // Lấy cảnh giới từ e.n (cấp MỚI) chứ không từ khung hình đã lưu — khung hình còn là
      // cảnh giới cũ nên sẽ gọi nhầm thần thú của cảnh giới vừa rời đi.
      const realm = realmOf(this.cfg, e.n ?? 0);
      this.spawnBeast(e.a ?? 0, 'breakthrough', now, tienIndex(this.cfg, realm));
      this.vfx.shake(22 * this.playerScale(e.a), 0.8, now);
      if (mine) {
        // Khựng hình 80ms + sóng xung kích + chớp trắng + chữ cảnh giới (spec §3.6 mục 5).
        this.hitstopUntil = now + HITSTOP_SECS;
        this.breakthrough.play(realmName(this.cfg, realm), { x: 0.5, y: 0.5 }, this.camera, now);
      }
    }
    this.vfx.play('lightning', e.x, e.y - size, size * 3, now);
    this.vfx.play('levelup_aura', e.x, e.y, size * 3, now);
    sound.play('breakthrough', mine ? 1 : 0.6);
    if (mine) {
      this.vfx.shake(30 * this.playerScale(e.a), 0.6, now);
    }
    this.players.get(e.a ?? 0)?.showBubble(pick(STR.emotes.breakthrough), now);
  }

  private handleWorldEvent(e: GameEvent, now: number, mine: boolean, size: number): void {
    switch (e.k) {
      case 'ItemPicked':
        // Viên Linh khí / Tiên khí được server GỘP: một sự kiện mang `n` viên, `b` là viên
        // cuối. Quét qua bãi viên dày có thể nhặt mấy trăm viên trong một gói — nếu vẫn một
        // sự kiện một viên thì gói snapshot phồng lên 21 KB và client kêu 400 tiếng chồng
        // lên nhau. Giờ vẽ một vệt bay hút và kêu đúng một tiếng.
        if (e.b !== undefined) {
          this.pickedBy.set(e.b, e.a ?? 0); // syncItems biến nó thành hiệu ứng bay hút
        }
        if (mine) {
          const s = e.s ?? '';
          sound.play(s === 'orb' ? 'pickup_exp' : s === 'sword' ? 'pickup_sword' : s === 'chest' ? 'chest_open' : 'pickup_buff');
          if (s === 'chest') {
            this.vfx.play('levelup_aura', e.x, e.y, size * 2, now);
          }
        }
        break;
      case 'TienKhiPicked':
        // Nhặt Tiên khí: thần thú mừng, lóe sáng và lượn nhanh một vòng (spec §3.5 v4.7).
        this.flashBeasts(e.a ?? 0, now);
        if (mine) {
          sound.play('pickup_exp');
        }
        break;
      case 'BuffEnd':
        if (mine) {
          sound.play('buff_end');
        }
        break;
      case 'MonsterHurt':
        sound.play('monster_hurt', 0.4);
        break;
      case 'BotLeave':
        // Admin "Rút bot ngay" (spec §9.5): bot rời phòng trong một làn khói, không biến
        // mất trơn — người đang nhìn thấy một nhân vật "tắt máy" chứ không thấy lỗi hiển thị.
        this.vfx.play('death_smoke', e.x, e.y, 180, now);
        break;
      case 'MonsterDie':
        this.vfx.play('death_smoke', e.x, e.y, 120, now);
        sound.play('monster_pop', 0.6);
        break;
      case 'ObstacleTooHard':
        // Tia lửa nhỏ + chữ nổi "Cần {cảnh giới}" (spec §8.2). Server đã giới hạn 1 lần /
        // 3 giây / vật cản, nên ở đây không phải tự đếm lại.
        this.vfx.play('clash_spark', e.x, e.y, 70, now);
        sound.play('obstacle_hit', 0.4);
        this.floatText.show(STR.needRealm(realmName(this.cfg, e.n ?? 0)), e.x, e.y, now);
        break;
      case 'ObstacleBroken':
        // Spec §8.2 (v6.7): vật cản "biến mất ngay lập tức, không hoạt ảnh vỡ, không đống
        // đổ nát (chỉ 1 tiếng bụp ngắn)" — nên KHÔNG phát khói ở đây.
        sound.play('obstacle_break', 0.7);
        break;
      case 'ObstacleRegrow':
        if (e.s === 'warn') {
          this.mapView?.warnObstacle(e.a ?? -1, now);
        } else {
          sound.play('obstacle_regrow', 0.6);
        }
        break;
      default:
        this.handleGateEvent(e, now, mine);
    }
  }

  private handleGateEvent(e: GameEvent, now: number, mine: boolean): void {
    switch (e.k) {
      case 'TeleportOut':
        this.vfx.play('levelup_aura', e.x, e.y, 300, now);
        sound.play('gate_in', mine ? 1 : 0.5);
        break;
      case 'GateWarn':
        this.mapView?.warnGate(e.n ?? -1, now);
        sound.play('gate_warn', 0.5);
        break;
      case 'TeleportIn':
        sound.play('gate_out', mine ? 1 : 0.5);
        if (mine) {
          this.players.get(e.a ?? 0)?.showBubble(pick(STR.emotes.teleport_in), now);
        }
        break;
      case 'Emote':
        this.players.get(e.a ?? 0)?.showBubble(pick(STR.emotes[e.s ?? ''] ?? ['!']), now);
        break;
      case 'ChestIncoming':
        sound.play('chest_incoming');
        break;
      default:
    }
  }

  /** ghostRunners: người đang Gom Đơn (bit `ghost`) — vật cản họ đi qua sẽ mờ (§6.5). */
  private ghostRunners(frames: ReturnType<ClientWorld['playerFrames']>): { x: number; y: number; r: number }[] {
    const out: { x: number; y: number; r: number }[] = [];
    for (const f of frames) {
      if (statusOn(f.s.st, 'ghost')) {
        out.push({ x: f.x, y: f.y, r: this.cfg.player.body_r * f.s.sc * 1.5 });
      }
    }
    return out;
  }

  /** myStatusChips: icon trạng thái + giây còn lại của chính mình cho HUD (§6.5). */
  myStatusChips(now: number): StatusChip[] {
    const me = this.world?.me();
    return me ? this.skill.chips(this.world?.myId ?? 0, me.st ?? 0, now) : [];
  }

  /** realm of the local player for HUD. */
  myRealm(): number {
    const me = this.world?.me();
    return me ? realmOf(this.cfg, me.L) : 0;
  }

  get currentZoom(): number {
    return this.zoom;
  }
}
