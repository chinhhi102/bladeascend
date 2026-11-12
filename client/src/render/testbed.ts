// i18n-ok-file: cảnh thử chỉ mở với quyền debug, không phải màn hình của người chơi
//
// CẢNH THỬ CỐ ĐỊNH — `?scene=testbed` (chỉ khi server cấp quyền debug).
//
//   node tools/browsertest/review.mjs ... (xem tools/browsertest/scenediff.mjs)
//
// ── Vì sao phải có ──────────────────────────────────────────────────────────────────────
//
// Ngày 12/11 tôi thử ba cách so ảnh trước/sau một lượt tối ưu đồ hoạ, trên cảnh THẬT, và cả
// ba đều vô dụng:
//
//   1. Đọc pixel từ canvas WebGL → ảnh rỗng (`preserveDrawingBuffer` tắt, và phải tắt).
//   2. So ảnh thô giữa hai phiên → 37% pixel khác, toàn bộ là NỀN: mỗi phiên nhân vật sinh
//      ra ở khu khác, chỗ khác.
//   3. So ảnh hiệu |bật − tắt| trong cùng phiên → vẫn 32,5/255 trên CÙNG một bản build, vì
//      hai pha cách nhau vài giây và bot đã đi chỗ khác.
//
// Cùng ngày, bảng `weakperf` cho 185 lô vẽ / 47 FPS ở lượt này và 447 lô vẽ / 23 FPS ở lượt
// sau — CÙNG build, CÙNG máy rảnh. Một cảnh sống thì vừa không so được ảnh, vừa không so
// được số.
//
// Nên: một cảnh KHÔNG có gì ngẫu nhiên. Không mạng, không bot, không thời gian.
//
// ── Những gì bị ghim, và vì sao từng cái ────────────────────────────────────────────────
//
//   - **Không nối WebSocket.** Mọi thứ trong cảnh là dữ liệu bịa tại client. Không có snapshot
//     thì không có nội suy, không có người lạ đi vào khung, không có bot.
//   - **Giờ đóng băng** ở `NOW`. `PlayerSprite.update` tính `dt` từ lần gọi trước, nên giờ
//     không đổi ⇒ `dt = 0` ⇒ hoạt ảnh, vệt mờ, nhịp hào quang đứng yên. Vòng kiếm cũng đứng
//     ở `RING_ANGLE`.
//   - **Cỡ khung và `resolution` cố định** (1280×720 @ 1), KHÔNG theo thiết bị. Để nguyên
//     `rendererResolution()` thì cùng một commit chụp trên máy DPR 1 và DPR 2 ra hai ảnh
//     khác kích thước, và bài so ảnh sẽ báo "hình đã đổi" cho một lần đổi màn hình.
//   - **Camera ghim**: container thế giới đặt cứng, không đuổi theo ai.
//   - **Hạt giống bản đồ cố định**: `map.id` quyết định decal (`buildDecals`), nên đặt tên
//     cố định là có đúng một cảnh.
//
// Cảnh gồm bốn hình nộm phủ đủ các đường vẽ đắt: Luyện Khí (1 kiếm), Kim Đan (60), Chân Tiên
// (300 kiếm + hào quang Tiên giới) và Đạo Tổ (1 050 kiếm — vòng dày nhất game), mỗi người đủ
// thân, bảng tên, huy hiệu cảnh giới và vòng kiếm. Nền là `MapView` thật: mặt đất, decal rải
// theo hạt giống cố định, vùng và cổng.
//
// CÒN THIẾU — vật cản. Bản đồ thử có khai ba vật cản (`boulder`, `bamboo_thin`, `cart`) với
// art có thật trong nhóm `boot`, nodes được `obstacleLayer()` trả về và đã add vào lớp
// entity, nhưng chúng KHÔNG vẽ ra và tôi chưa tìm ra vì sao (13/11). Ghi ra đây thay vì im
// lặng: ai đọc cảnh này phải biết nó chưa gác được phần vẽ vật cản.
import { Application, Container } from 'pixi.js';

import { type GameConfig, realmOf, tienIndex } from '../config';
import type { GameMap, PlayerState } from '../net/protocol';
import type { PlayerFrame, PlayerView } from '../net/state';
import { assets } from './assets';
import { countDrawCalls, countNodes, drawCalls } from './drawstats';
import { loadFonts } from './fonts';
import { MapView } from './mapview';
import { PlayerSprite } from './player';
import { ringGeom } from './ringgeom';
import { AURA_FAR_PX, TienAura } from './tienaura';

/** Cỡ khung và độ phân giải CỐ ĐỊNH — xem ghi chú đầu tệp. */
const W = 1280;
const H = 720;
/** Giờ đóng băng (giây). Số nào cũng được, miễn không đổi giữa hai lần chụp. */
const NOW = 1000;
/** Góc vòng kiếm đóng băng (rad). */
const RING_ANGLE = 0.6;
/** Tâm cảnh trong toạ độ thế giới — đặt giữa bản đồ thử cho gọn. */
const CX = 3000;
const CY = 2400;
/** Chừa mép quanh cụm hình nộm (đơn vị thế giới) để không ai bị cắt rìa khung. */
const PAD = 240;

/** mkMap: bản đồ thử bé, hạt giống cố định. */
function mkMap(): GameMap {
  const size = 6000;
  return {
    id: 'testbed',
    name: 'Cảnh thử',
    size,
    biome: 'forest',
    play_margin: 100,
    // Vật cản đặt tay quanh tâm cảnh: một cụm đá, một bụi, một cây. `s` là hình THẤY được
    // (kiếm chém vào), `f` là chân đế (chặn thân) — đúng hai hình của §8.2.
    obstacles: [
      // Tên loại phải có thật trong `assets/src/obstacles/` — `MapView` tìm
      // `obstacles/<k>/intact.png`. Gõ một cái tên không tồn tại thì vật cản im lặng không
      // hiện, và cảnh thử mất một phần hình mà không ai báo.
      // Đặt vào khoảng TRỐNG giữa bốn ô hình nộm (ô vẽ đè lên nền), nếu không vật cản nằm
      // ngay dưới bảng tên hoặc dưới vòng kiếm và ảnh chụp trông như thiếu nó.
      { id: 1, k: 'boulder', s: { x: 2760, y: 2040, r: 70 }, f: { x: 2760, y: 2070, r: 46 }, t: 0 },
      { id: 2, k: 'bamboo_thin', s: { x: 3000, y: 2280, r: 60 }, f: { x: 3000, y: 2310, r: 38 }, t: 0 },
      { id: 3, k: 'cart', s: { x: 3240, y: 2040, r: 80 }, f: { x: 3240, y: 2070, r: 52 }, t: 2 },
    ],
    zones: [{ id: 0, k: 'forest', s: { c: true, x: CX, y: CY, r: 1400 } }],
    gates: [{ id: 0, x: CX + 560, y: CY - 380, n: 'Cổng thử', c: 'forest' }],
    decor: [
      { sp: 'decor/grass_1.png', x: CX - 120, y: CY + 150 },
      { sp: 'decor/grass_1.png', x: CX + 180, y: CY - 220 },
    ],
  };
}

/**
 * mkFrame: một hình nộm. `realm` là chỉ số cảnh giới (0 = Luyện Khí; ≥ 9 = Tiên giới).
 *
 * `sc` truyền thẳng chứ không suy từ `level`: trong game thật `sc` do server tính, và cảnh
 * thử chỉ cần ba cỡ người RÕ RÀNG khác nhau để nhìn ra thay đổi đồ hoạ — không cần đúng
 * từng chữ số công thức tu luyện. Đặt cứng thì cảnh không đổi theo một lần chỉnh cân bằng.
 */
function mkFrame(cfg: GameConfig, id: number, name: string, char: string,
  realm: number, tier: number, swords: number, sc: number, x: number, y: number): PlayerFrame {
  const level = realm * cfg.cultivation.tiers_per_realm + tier;
  const s: PlayerState = {
    id, x, y, f: 0, L: level, n: swords, ph: 0, d: 1, tl: 0.2, sc,
    md: 0, bf: 0, fl: 0, xp: 0,
  };
  const view: PlayerView = { id, name, char, snaps: [] };
  return { view, s, x, y, ringAngle: RING_ANGLE };
}

/**
 * runTestbed dựng cảnh và vẽ mãi cùng một khung.
 *
 * Vẽ lại mỗi khung (thay vì vẽ một lần) là có chủ ý: ảnh chụp lúc nào cũng giống nhau, và
 * nếu một lượt tối ưu làm hỏng đường vẽ lặp (ví dụ buffer tĩnh không được gói lại) thì cảnh
 * này vẫn hiện ra đúng cái hỏng đó.
 */
export async function runTestbed(stage: HTMLElement, cfg: GameConfig): Promise<void> {
  // Quyền do SERVER quyết: `/api/debug-allowed` dùng đúng `admin.DebugGate` mà `/ws` dùng
  // (ADMIN_SPEC §1.1). Tự xét ở client thì bất kỳ ai gõ thêm một tham số vào URL cũng vào
  // được — và cảnh này bày ra cả hình nộm Đạo Tổ lẫn cửa debug.
  const allowed = await fetch('/api/debug-allowed')
    .then((r) => r.json() as Promise<{ ok?: boolean }>)
    .then((j) => j.ok === true)
    .catch(() => false);
  if (!allowed) {
    stage.textContent = 'Cảnh thử chỉ mở cho phiên có quyền debug (ADMIN_SPEC §1.1).';
    return;
  }

  const app = new Application();
  await app.init({
    preference: 'webgl',
    width: W,
    height: H,
    background: 0x2b2b2b,
    antialias: true,
    resolution: 1,
    autoDensity: false,
  });
  stage.appendChild(app.canvas);
  // Bộ đếm lô vẽ: cảnh thử là chỗ DUY NHẤT con số này so được giữa hai bản build, vì cảnh
  // không đổi. Trên cảnh thật nó đi từ 156 tới 460 lô tuỳ bot đứng đâu (đo 12/11).
  const gl = (app.renderer as unknown as { gl?: WebGL2RenderingContext }).gl;
  if (gl) {
    countDrawCalls(gl);
  }

  const chars = cfg.characters.map((c) => (typeof c === 'string' ? c : (c as { id: string }).id));
  const char = chars[0] ?? 'runner';
  assets.setFallbackChars(cfg.characters);
  // Chờ XONG mọi thứ cần vẽ: thiếu một atlas là ảnh chụp có một hình nộm vẽ bằng nhân vật
  // dự phòng, và lần chụp sau lại đủ — tức là một "thay đổi đồ hoạ" không ai gây ra.
  await assets.load(['home', 'boot', `char_${char}`]);
  await loadFonts();

  const map = mkMap();
  const world = new Container();
  const mapView = new MapView(map, cfg.gates.radius);
  const entities = new Container();
  const rings = new Container();
  const glow = new Container();
  entities.sortableChildren = true;
  // Thứ tự lớp COPY nguyên của `Scene`: vật cản nằm cùng lớp entity để y-sort với nhân vật,
  // vòng kiếm dưới mọi thân người, hào quang trên vòng kiếm, tán bụi trên cùng. Xếp khác
  // `Scene` thì cảnh thử đẹp hay xấu cũng không nói gì về cảnh thật.
  for (const node of mapView.obstacleLayer()) {
    entities.addChild(node);
  }
  world.addChild(mapView.ground, mapView.decor, mapView.under, rings, mapView.gateLayer,
    glow, entities, mapView.over);
  app.stage.addChild(world);

  // Nền: bản đồ ở một zoom vừa phải, camera ghim vào tâm cảnh.
  const BG_ZOOM = 0.5;
  world.scale.set(BG_ZOOM);
  world.position.set(W / 2 - CX * BG_ZOOM, H / 2 - CY * BG_ZOOM);

  const top = cfg.cultivation.realms.length + cfg.cultivation.tien_realms.length - 1;
  const maxSw = cfg.sword.max_swords;
  // Bốn hình nộm: cảnh giới, số kiếm, cỡ người. Vòng Đạo Tổ (1 050 kiếm, scale 5) có bán
  // kính hàng nghìn đơn vị thế giới, nên CHỖ ĐỨNG và ZOOM phải tính TỪ bán kính thật chứ
  // không đặt tay — đặt tay thì vòng to tràn hết khung và cảnh thử không nhìn ra gì.
  const spec = [
    { id: 101, name: 'LuyenKhi', realm: 0, tier: 0, n: 1, sc: 1 },
    { id: 102, name: 'KimDan', realm: 2, tier: 4, n: 60, sc: 2 },
    // Chân Tiên: cảnh giới Tiên đầu tiên → có hào quang Tiên giới (§3.4).
    { id: 104, name: 'ChanTien', realm: cfg.cultivation.realms.length, tier: 2, n: 300, sc: 3 },
    { id: 103, name: 'DaoTo', realm: top, tier: cfg.cultivation.tiers_per_realm - 1,
      n: maxSw[maxSw.length - 1] ?? 1050, sc: 5 },
  ];
  // BỐN Ô, MỖI Ô MỘT ZOOM RIÊNG.
  //
  // Vòng Luyện Khí (1 kiếm, scale 1) và vòng Đạo Tổ (1 050 kiếm, scale 5) chênh nhau khoảng
  // một trăm lần bán kính. Nhét cả bốn vào một zoom thì hoặc Đạo Tổ tràn hết khung, hoặc
  // Luyện Khí còn vài pixel — thử rồi, ra đúng vậy. Cảnh thử không phải một ván game, nó là
  // một BẢNG MẪU: mỗi hình nộm một ô, mỗi ô co sao cho vòng kiếm vừa ô.
  //
  // Giá phải trả: hình nộm không còn y-sort chung với vật cản như trong trận. Chấp nhận —
  // thứ cảnh này cần chứng minh là "từng thứ vẽ ra có đổi hình không", và nó làm được việc
  // đó rõ hơn hẳn một ván game thu nhỏ.
  const panels = new Container();
  app.stage.addChild(panels);
  const COLS = 2;
  const CELL_W = W / COLS;
  const CELL_H = H / 2;
  const sprites: { sp: PlayerSprite; zoom: number }[] = [];
  const auras: (TienAura | null)[] = [];
  const dummies: PlayerFrame[] = [];
  spec.forEach((d, i) => {
    const f = mkFrame(cfg, d.id, d.name, char, d.realm, d.tier, d.n, d.sc, 0, 0);
    dummies.push(f);
    const cell = new Container();
    const z = Math.min(CELL_W, CELL_H) / (ringGeom(cfg, d.sc, d.n).reach * 2 + PAD);
    cell.scale.set(z);
    cell.position.set((i % COLS) * CELL_W + CELL_W / 2, Math.floor(i / COLS) * CELL_H + CELL_H / 2);
    const ringLayer = new Container();
    const glowLayer = new Container();
    const bodyLayer = new Container();
    bodyLayer.sortableChildren = true;
    cell.addChild(ringLayer, glowLayer, bodyLayer);
    panels.addChild(cell);

    const sp = new PlayerSprite(cfg, f.view.id, f.view.name, f.view.char, false);
    bodyLayer.addChild(sp.root);
    ringLayer.addChild(sp.ring);
    sprites.push({ sp, zoom: z });

    const idx = tienIndex(cfg, realmOf(cfg, f.s.L));
    if (idx < 0) {
      auras.push(null);
    } else {
      const a = new TienAura(cfg);
      a.setRealm(idx);
      glowLayer.addChild(a.root);
      auras.push(a);
    }
  });

  // MapWorld hẹp đúng hai trường bộ vẽ bản đồ cần (xem `MapWorld`) — chính cách trang quản
  // trị đưa trạng thái giả vào bộ vẽ này, ADMIN_SPEC §5.
  const fakeWorld = {
    brokenObstacles: new Map<number, boolean>(),
    canCutObstacle: (tier: number): boolean => tier <= 0,
  } as unknown as Parameters<MapView['update']>[0];

  // Số đo của cảnh thử. Pixi vẽ bao lâu, mấy lô vẽ, bao nhiêu nút — trên một cảnh KHÔNG đổi,
  // nên hiệu giữa hai bản build là hiệu của CODE, không phải của cảnh.
  let pixiMs = 0;
  let frames = 0;
  let fpsAcc = 0;
  let fps = 0;
  let callsPerFrame = 0;

  // KHUNG ĐẦU lùi giờ 1 giây, các khung sau dùng `NOW`.
  //
  // `PlayerSprite.applyAlpha` đặt `bornAt = now` ở lần gọi đầu rồi mờ dần vào theo
  // `(now - bornAt) / 0.12`. Với đồng hồ đứng im thì hiệu đó mãi mãi bằng 0 — nhân vật alpha
  // 0 vĩnh viễn, và ảnh chụp đầu tiên của cảnh thử đúng là chỉ có vòng kiếm với hào quang,
  // không có người. Cho khung đầu một mốc sớm hơn thì fade-in xong ngay ở khung thứ hai, và
  // từ đó mọi khung giống hệt nhau.
  let warm = false;
  app.ticker.remove(app.render, app);
  app.ticker.add((t) => {
    const now = warm ? NOW : NOW - 1;
    warm = true;
    mapView.update(fakeWorld, now, 0, null, cfg.obstacles.occlusion_alpha);
    for (let i = 0; i < sprites.length; i++) {
      const f = dummies[i];
      const { sp, zoom } = sprites[i];
      sp.update(f, zoom, now, false, 1, 0, 1);
      const a = auras[i];
      if (a) {
        const r = cfg.player.body_r * f.s.sc;
        // Tham số cuối là "to trên màn hình" — cùng phép thử `Scene` dùng (AURA_FAR_PX),
        // không đặt cứng `true`, để hào quang vẽ đúng mức chi tiết của cỡ nó đang hiện.
        a.update(f.x, f.y + r * 0.75, r, now, 1, r * 2.6 * zoom >= AURA_FAR_PX);
      }
    }
    drawCalls.acc = 0;
    const t0 = performance.now();
    app.render();
    // Trung bình trượt, giống `Scene`: một khung lẻ nói rất ít, mà bài đo chỉ đọc một lần.
    pixiMs = pixiMs === 0 ? performance.now() - t0 : pixiMs * 0.9 + (performance.now() - t0) * 0.1;
    callsPerFrame = drawCalls.acc;
    frames++;
    fpsAcc += t.deltaMS;
    if (fpsAcc >= 500) {
      fps = (frames * 1000) / fpsAcc;
      frames = 0;
      fpsAcc = 0;
    }
  });

  // Cờ cho bài chụp: chờ cờ này rồi mới chụp, khỏi đoán bằng `sleep`.
  (window as unknown as { kxTestbed?: unknown }).kxTestbed = {
    ready: true,
    w: W,
    h: H,
    stats: (): Record<string, number> => ({
      pixiMs: +pixiMs.toFixed(3),
      fps: +fps.toFixed(1),
      drawCalls: callsPerFrame,
      nodes: countNodes(app.stage),
    }),
  };
}
