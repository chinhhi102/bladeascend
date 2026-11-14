import './ui/style.css';

import { sound } from './audio/sound';
import { type GameConfig, levelThresholds, loadConfig, realmOf, tienIndex } from './config';
import { Input } from './input';
import { Conn } from './net/conn';
/** Chữ tên khu hiện bao lâu khi vào khu mới (spec §8.0). */
const ZONE_BANNER_SECS = 1.5;

/**
 * ZONE_ART_WAIT_MS: qua cổng sang khu chưa tải xong thì giữ màn hình dịch chuyển tối đa
 * ngần này rồi vào với placeholder (GAME_SPEC §13.7 mục 3, dòng cuối).
 */
const ZONE_ART_WAIT_MS = 2000;

import { FLAG, KICK_OTHER_TAB, KICK_SERVER_UPDATE, PROTOCOL_VERSION } from './net/protocol';
import { SCHEMA_HASH } from './net/wire.gen';
import { watchContextLoss } from './render/glrecover';
import { drawGuardStats, faultSystem, setSysErrorHook, type SysKey, sysResetRun, sysRun, sysSummary, isFake } from './render/guard';
import { clearSession, loadSession, saveSession } from './net/session';
import type { DeadMsg, GameMap, KickedMsg, ServerMsg, Snapshot, Welcome } from './net/protocol';
import { ClientWorld } from './net/state';
import { assets, setAssetTroubleHook } from './render/assets';
import { isMobile } from './render/device';
import { loadFonts } from './render/fonts';
import { frameDivider, measureRefreshHz } from './render/framepace';
import { GameScene } from './render/scene';
import { loadJitter } from './render/jitter';
import { cycleQuality, quality, setQuality } from './render/quality';
import { ringGeom } from './render/ringgeom';
import type { Telemetry, TelemetryContext } from './telemetry';
import { Screens, button, h, setUiErrorHook } from './ui/dom';
import { LoadingScreen, startBgLoadIndicator } from './ui/loading';
import { kickedScreen } from './ui/kicked';
import { guestId, homeScreen, type Profile, savedProfile } from './ui/home';
import { Hud } from './ui/hud';
import { leaderboardScreen } from './ui/leaderboard';
import { rbStats } from './ui/roomboard';
import * as feedback from './ui/feedback';
import { LANG } from './i18n';
import { STR } from './ui/strings';
import { clearEndedRun, endedRun, saveEndedRun } from './ui/runend';
import { summaryScreen } from './ui/summary';

/**
 * Phím dùng kỹ năng (spec §6.3): **Space**.
 *
 * Từng thử **Ctrl** nhưng bỏ, và lý do đáng ghi lại để đừng ai đổi lại: phím đi là **WASD**,
 * nên vừa giữ Ctrl vừa bấm W là **Ctrl+W — đóng tab**. Chrome xếp Ctrl+W vào nhóm phím tắt
 * dành riêng cho trình duyệt, `preventDefault()` KHÔNG chặn được; cách duy nhất chặn thật là
 * Keyboard Lock API, mà nó chỉ chạy ở chế độ toàn màn hình. Space không đụng phím tắt nào.
 */
const SKILL_KEYS = ['Space'];

const DEBUG_KEYS: Record<string, Record<string, unknown>> = {
  KeyE: { cmd: 'exp', n: 100 },
  KeyR: { cmd: 'realm', n: 0 },
  KeyK: { cmd: 'sword', n: 10 },
  KeyX: { cmd: 'die', n: 0 },
  KeyB: { cmd: 'bot', n: 0 }, // thả 1 bot cùng cảnh giới ngay cạnh mình
  KeyG: { cmd: 'skill', n: 1 }, // hồi đầy lượt kỹ năng, khỏi chờ hồi chiêu
  KeyU: { cmd: 'summon', n: 0 }, // kéo người thật khác trong khu về cạnh mình (chụp ảnh kỹ năng)
};

/** fpsCap: 60 on PC, 45 on phones to save battery; override with ?fps=30 (M7). */
function fpsCap(): number {
  const q = Number(new URLSearchParams(location.search).get('fps'));
  if (q > 0) {
    return q;
  }
  return Input.isTouchDevice() ? 45 : 60;
}

/**
 * paceFrames: máy tính thì đổi trần 60 thô thành nhịp chia đều theo màn hình (framepace.ts).
 * Điện thoại giữ trần 45 như cũ (pin, nhiệt) — chưa đo lại trên máy thật nên chưa đụng.
 * `?fps=` trên URL là lệnh tay của bài đo, giữ nguyên.
 */
async function paceFrames(scene: GameScene): Promise<void> {
  if (Input.isTouchDevice() || Number(new URLSearchParams(location.search).get('fps')) > 0) {
    return;
  }
  scene.paceFrames(frameDivider(await measureRefreshHz(), fpsCap()));
}

class App {
  private readonly screens: Screens;
  private readonly scene: GameScene;
  private readonly world: ClientWorld;
  private readonly hud: Hud;
  private readonly input: Input;
  private readonly conn: Conn;
  private readonly thresholds: number[];
  private profile: Profile | null = null;
  /** lastFps: FPS gần nhất, để đính kèm vào góp ý (ADMIN_SPEC §7.1). */
  private lastFps = 0;
  /**
   * playing — đang trong trận, hay đang ở trang chủ.
   *
   * Dùng cặp get/set thay vì một trường trơn vì đoạn giới thiệu HTML tĩnh (§20) phải TẮT
   * khi đang chơi: để nó lại thì trang cuộn được, và một cú quét ngón tay giữa trận kéo
   * màn hình game đi mất. Cờ này đổi ở năm chỗ trong tệp này — nối vào setter thì không
   * chỗ nào quên được.
   */
  private playingFlag = false;
  private get playing(): boolean {
    return this.playingFlag;
  }
  private set playing(v: boolean) {
    this.playingFlag = v;
    document.body.classList.toggle('in-game', v);
  }
  // Snapshots that arrive while the map is still loading: the first ones carry names and
  // characters (sent only once per player), so they must not be dropped.
  private joining = false;
  private early: Snapshot[] = [];
  private debug = false;
  /** telemetry: null khi build không bật cờ TELEMETRY (bản release). */
  private telemetry: Telemetry | null = null;
  /** zoneIndex: khu đang đứng, lấy từ Welcome (spec §8.0). Telemetry chia nhóm theo nó. */
  private zoneIndex = 0;
  /** zoneAt: mốc (performance.now) lần đổi khu gần nhất — mọi báo cáo lỗi cần nó (§13.8). */
  private zoneAt = 0;
  /** animInfo: phím J đang bật hay tắt (GAME_SPEC §12.5 C). */
  private animInfo = false;
  /** detachGl: gỡ bộ canh mất ngữ cảnh WebGL khỏi canvas CŨ khi dựng lại renderer. */
  private detachGl: (() => void) | null = null;

  constructor(
    private readonly cfg: GameConfig,
    private readonly overlay: HTMLElement,
    stage: HTMLElement,
    private readonly loading: LoadingScreen,
    /** fonts: BitmapFont đang tải nền. `join()` chờ nó trước khi vào trận (spec §12.3). */
    private readonly fonts: Promise<void>,
    /** artReady: nhóm `home` đang tải nền; trang chủ vá ảnh tại chỗ khi xong. */
    private readonly artReady: Promise<void>,
  ) {
    this.screens = new Screens(overlay);
    this.scene = new GameScene(cfg);
    this.world = new ClientWorld(cfg);
    this.hud = new Hud(cfg, Input.isTouchDevice());
    this.input = new Input(stage);
    this.bindBurnButton();
    this.bindSkillButton();
    this.conn = new Conn((m) => this.onMessage(m));
    // Gói hỏng phải ĐI VỀ ĐƯỢC, không chỉ nằm trong console của máy người chơi.
    // Kèm `errorContext()` (khu, cảnh giới, ping, FPS) vì một lỗi giải mã gần như chắc
    // chắn phụ thuộc hoàn cảnh — biết nó xảy ra lúc đang đổi khu hay lúc mạng yếu là nửa
    // phần chẩn đoán.
    this.conn.onDecodeFail = (info) => {
      this.telemetry?.event('decode_fail', JSON.stringify({ info, ...(this.telemetry?.errorContext() ?? {}) }));
    };
    // Gói hợp lệ nhưng xử lý nó ném lỗi (§13.8): gói đó bị bỏ, phần còn lại chạy tiếp.
    this.conn.onHandlerError = (err, kind) => {
      this.telemetry?.errorEvent(`xử lý gói "${kind}": ${err instanceof Error ? err.message : String(err)}`); // i18n-ok: chữ gửi về telemetry, không hiện cho người chơi
    };
    this.thresholds = levelThresholds(cfg);
    this.setupFeedback();
  }

  async start(stage: HTMLElement): Promise<void> {
    await this.scene.init(stage, fpsCap());
    void paceFrames(this.scene);
    // Cắm vào `scene.onFrame` chứ không `app.ticker.add`: dựng lại renderer (§13.8) thay
    // hẳn `Application`, và một lời đăng ký trên ticker cũ sẽ im lặng ngừng chạy — game vẽ
    // tiếp nhưng không còn cập nhật gì nữa, đúng kiểu hỏng khó tìm nhất.
    this.scene.onFrame = (dt) => {
      // FPS trượt mượt để đính kèm vào góp ý — số tức thời nhảy quá, đọc không ra gì.
      this.lastFps = this.lastFps * 0.9 + (1 / Math.max(dt, 0.001)) * 0.1;
      this.frame(dt);
    };
    this.scene.onBreadcrumb = (code, detail) => this.telemetry?.breadcrumb(code, detail);
    setAssetTroubleHook((code, detail) => this.telemetry?.breadcrumb(code, detail));
    this.scene.onGuardEvent = (kind, err) => {
      this.telemetry?.breadcrumb(kind === 'draw_error' ? 'loi-ve' : kind);
      this.telemetry?.event(kind, JSON.stringify({ err: err ? String(err).slice(0, 200) : '', ...(this.telemetry?.errorContext() ?? {}) }));
    };
    // Dựng lại renderer thất bại → tải lại trang và VÀO LẠI PHIÊN (§13.8 → §7.1). Đây là
    // đường lùi cuối cùng; tới đây thì cảnh đã hỏng tới mức không vẽ được nữa.
    this.scene.onRebuildFailed = () => {
      // Mã phiên đã nằm trong sessionStorage từ lúc vào trận, nên chỉ cần tải lại: luồng
      // khởi động ở `start()` tự gửi nó và vào lại đúng nhân vật (§7.1).
      window.setTimeout(() => location.reload(), 600);
    };
    window.setInterval(() => this.sendInput(), 1000 / this.cfg.net.tick_hz);
    window.addEventListener('keydown', (e) => {
      if (this.playing && SKILL_KEYS.includes(e.code) && !e.repeat) {
        this.skillPressed = true; // kỹ năng nhân vật (spec §6.3)
        e.preventDefault();
        return;
      }
      if (this.playing && e.code === 'KeyM' && !e.repeat) {
        this.hud.minimap.toggle(); // spec §12.1
        return;
      }
      if (this.playing && e.code === 'KeyJ' && !e.repeat && this.debug) {
        // GAME_SPEC §12.5 C: hiện tên anim, frame, pha và điểm độ giật trên từng nhân vật.
        this.animInfo = !this.animInfo;
        void loadJitter().then(() => this.scene.showAnimInfo(this.animInfo));
        this.hud.showBanner(this.animInfo ? STR.animInfoOn : STR.animInfoOff, 1.5);
        return;
      }
      if (this.playing && this.animInfo && (e.code === 'BracketLeft' || e.code === 'BracketRight') && !e.repeat) {
        // ANSWERS mục 1: đi lần lượt qua từng animation theo thứ tự bảng điểm để chủ dự án
        // chốt ngưỡng đỏ bằng mắt.
        this.scene.stepAnimViewer(e.code === 'BracketRight' ? 1 : -1);
        e.preventDefault();
        return;
      }
      if (this.playing && e.code === 'F3' && !e.repeat) {
        this.hud.togglePerf(); // spec §13.6: bảng đo bật/tắt bằng F3
        e.preventDefault();
        return;
      }
      if (this.playing && e.code === 'KeyV' && !e.repeat) {
        // Spec §13.6 mục 8 (v6.20): Tự động → Cao → Vừa → Thấp → Tự động.
        const pick = cycleQuality();
        const name = STR.quality[pick.q];
        this.hud.showBanner(STR.qualitySet(pick.auto ? STR.qualityAuto(name) : name), 2.5);
        return;
      }
      this.debugKey(e);
    });
    this.conn.connect(
      () => undefined,
      () => {
        if (!this.playing) {
          return;
        }
        this.playing = false;
        this.input.enabled = false;
        this.hud.root.remove();
        this.screens.show(h('div', 'screen center', h('div', 'panel', STR.disconnected)));
        // §7.1: mạng rớt KHÔNG còn là mất ván. Nhân vật ở lại thế giới 20 giây, nên khi
        // `Conn` tự nối lại (1,5 s sau) thì gửi mã phiên và vào lại đúng nhân vật đó.
        // Chỉ khi quá hạn — hoặc không có mã — mới đành về trang chủ.
        const s = loadSession();
        if (s && !endedRun()) {
          window.setTimeout(() => {
            if (!this.playing) {
              this.join({ name: s.name, char: s.char }, s.key);
            }
          }, 400);
          return;
        }
        window.setTimeout(() => this.showHome(), 2500);
      },
    );
    // Mất ngữ cảnh WebGL (trình điều khiển card reset, máy ngủ dậy): canvas sẽ đen vĩnh
    // viễn nếu không làm gì — tải lại trang rồi vào lại ĐÚNG NHÂN VẬT bằng mã phiên §7.1.
    //
    // Không còn tham số "nhớ hồ sơ": mã phiên đã nằm sẵn trong `sessionStorage` từ lúc vào
    // trận, nên mọi đường tải lại (F5, mất WebGL, dựng lại renderer hỏng) đều đi chung một
    // cơ chế. `pendingRejoin` cũ chỉ nhớ tên + nhân vật và vào một trận MỚI.
    const watchGl = (canvas: HTMLCanvasElement): void => {
      this.detachGl?.();
      this.detachGl = watchContextLoss(
        canvas,
        (canRecover) => {
        // Gửi về server KÈM BỐI CẢNH: mất ngữ cảnh WebGL trên iPhone gần như luôn là hệ quả
        // của hết bộ nhớ GPU, nên con số `texMB` đi cùng mới nói được nguyên nhân.
        this.telemetry?.breadcrumb('mat-gl', canRecover ? 'cuu-duoc' : 'mat-han');
      this.telemetry?.event('webgl_context_lost', JSON.stringify({ canRecover, ...(this.telemetry?.errorContext() ?? {}) }));
        this.playing = false;
        this.input.enabled = false;
        this.hud.root.remove();
        this.screens.show(h('div', 'screen center', h('div', 'panel', canRecover ? STR.glLost : STR.glLostManual)));
      },
      // Trình duyệt cấp lại ngữ cảnh (chuyện thường trên điện thoại khi quay lại app): về
      // trang chủ thay vì nằm mãi ở màn hình "đang khôi phục". Không tự vào lại phòng —
      // phòng cũ đã đá mình ra từ lâu, và tự nhảy vào một trận mới là đúng cái cảm giác
      // "game tự chơi lại từ đầu" mà ta đang sửa.
        () => {
          this.telemetry?.event('webgl_context_restored', JSON.stringify(this.telemetry?.errorContext() ?? {}));
          this.resumeOrHome();
        },
        // Đang DỰNG LẠI renderer thì chính ta vứt ngữ cảnh — không phải sự cố (§13.8).
        () => this.scene.isRebuilding(),
      );
    };
    watchGl(this.scene.app.canvas as HTMLCanvasElement);
    // Dựng lại renderer xong thì canvas là thẻ MỚI: gắn lại bộ canh lên thẻ đó.
    this.scene.onCanvasReplaced = watchGl;
    // Bắt đầu thu số liệu NGAY khi game khởi động xong, không đợi vào trận.
    //
    // Trước đây telemetry chỉ bật trong `onWelcome`, nên **máy nào không vào nổi trận thì
    // không để lại dấu vết nào cả** — đúng tình huống đang phải tìm: nhật ký server ngày
    // 31/10 cho thấy một người vào lại phòng 8 lần trong 40 giây, mà bảng `sessions` không
    // có lấy một phiên điện thoại nào. Thiết bị hỏng nặng nhất lại là thiết bị im lặng nhất.
    void this.startTelemetry();

    // Vừa tải lại trang giữa trận (F5, mất đồ hoạ, dựng lại renderer hỏng): vào lại ĐÚNG
    // nhân vật bằng mã phiên (§7.1), không bắt bấm lại "Vào chơi" và không vào trận mới.
    //
    // `endedRun()` đi trước: ván đã kết thúc vì chết thì luôn quay về màn Tổng kết, kể cả
    // khi mã phiên còn hạn (v6.29 §7).
    const again = endedRun() ? null : loadSession();
    if (again) {
      this.join({ name: again.name, char: again.char }, again.key);
    } else {
      this.resumeOrHome();
    }
  }

  /** resumeOrHome: ván vừa kết thúc vì chết thì hiện lại Tổng kết, không thì về trang chủ. */
  private resumeOrHome(): void {
    const run = endedRun();
    if (!run) {
      this.showHome();
      return;
    }
    this.profile = run.profile;
    this.onDead(run.dead);
  }

  /** onKicked: bị mời ra hoặc bị cấm (ADMIN_SPEC §6.2). */
  private onKicked(m: KickedMsg): void {
    // `stop()` chứ không phải đóng thường: nó tắt luôn việc tự nối lại. Đóng thường thì
    // client nối lại sau vài giây, server cấm tiếp, và người vừa bị cấm quay vòng giữa
    // "bị cấm" và "đang kết nối" mà không đọc kịp lý do.
    this.conn.stop();
    this.playing = false;
    this.input.enabled = false;
    this.hud.root.remove();
    this.telemetry?.setPhase('home');
    // Nhân vật vừa bị TAB KHÁC của chính người này chiếm (§7.1): tab cũ phải quên mã phiên
    // ngay, nếu không nó nối lại và giành ngược nhân vật về — hai tab đá qua đá lại mãi.
    // Server đang cập nhật: khác hẳn bị admin mời ra — không ai làm gì sai, và vào lại
    // được sau vài giây. Màn Tổng kết đã hiện trước đó (server gửi `dead` ngay trước gói
    // này), nên ở đây chỉ cần nói vì sao và mời vào lại.
    if (m.reason === KICK_SERVER_UPDATE) {
      clearSession();
      clearEndedRun();
      this.screens.show(h('div', 'screen center', h('div', 'panel', STR.serverUpdating)));
      return;
    }
    if (m.reason === KICK_OTHER_TAB) {
      clearSession();
      clearEndedRun();
      this.screens.show(h('div', 'screen center', h('div', 'panel', STR.otherTab)));
      return;
    }
    clearSession();
    this.screens.show(kickedScreen(m.reason, m.until, () => this.showHome()));
  }

  private showHome(): void {
    this.telemetry?.setPhase('home');
    this.screens.show(homeScreen(this.cfg, (p) => this.join(p), () => this.showLeaderboard(() => this.showHome()), this.artReady));
  }

  private showLeaderboard(back: () => void): void {
    this.screens.show(leaderboardScreen(this.cfg, back));
  }

  /**
   * join: tải phần còn thiếu của gói chặn rồi mới vào phòng (GAME_SPEC §13.7 mục 3).
   *
   * Thứ tự ĐÚNG như bảng nhóm của spec: `boot` (nền, vật cản chung, cổng, quái, vật phẩm)
   * + `char_<nhân vật mình>`. Nhóm `zone_*` chờ tới `onWelcome`, vì chỉ lúc đó server mới
   * nói mình được xếp vào khu nào — xếp khu là việc của server (khu ít người nhất, §8.0),
   * client không đoán được, và đoán sai là tải 3,2 MB vật cản của khu không dùng tới.
   */
  private join(p: Profile, resume = ''): void {
    clearEndedRun();
    // Mạng mới = sổ chống sập mới (§13.8): "tắt hẳn tới hết mạng" là hết VÁN, không phải
    // hết đời tab. Rất nhiều lỗi vẽ chỉ gắn với một khu hoặc một đối thủ cụ thể.
    // VÀO LẠI (§7.1) thì KHÔNG xoá: vẫn là ván cũ, và một hệ thống vừa hỏng ở ván này
    // nhiều khả năng hỏng tiếp ngay sau khi tải lại.
    if (!resume) {
      sysResetRun();
    }
    this.profile = p;
    sound.startMusic();
    this.telemetry?.setPhase('loading');
    this.screens.show(null);
    // Gói vào trận đã tải xong trong lúc người chơi gõ tên thì **không hiện thanh tiến độ**
    // — hiện một thanh nhảy từ 0 lên 100% trong một khung hình chỉ làm người ta giật mình.
    const need = ['home', 'boot', `char_${p.char}`];
    const pending = assets.pending(need).length > 0;
    if (pending) {
      this.loading.show();
      this.loading.setRunner(p.char);
      this.loading.step(STR.loadChar(STR.characters[p.char] ?? p.char));
    }
    void (async () => {
      try {
        await assets.load(need, (pr) => {
          if (pending) {
            this.loading.bytes(pr);
          }
        });
        // Bảng tên trong trận vẽ bằng BitmapText, nên font PHẢI xong trước khi vào trận
        // (spec §12.3). Nó tải nền từ lúc mở trang nên hầu như luôn xong sẵn.
        await this.fonts;
      } catch (err) {
        // Mạng ngắt giữa lúc tải: hiện "Thử lại" và tải TIẾP phần còn thiếu — `assets.load`
        // bỏ qua tệp đã xong, nên bấm lại không tải lại từ đầu (§13.7 mục 4).
        console.warn('tải gói vào trận lỗi', err);
        this.loading.show();
        this.loading.setRunner(p.char);
        this.loading.fail(() => this.join(p, resume));
        return;
      }
      // Nhóm tải nền: VFX Tiên giới, âm thanh — và trên MÁY BÀN thì cả 10 nhân vật còn lại.
      //
      // Điện thoại KHÔNG tải sẵn 10 nhân vật kia: mỗi atlas là một texture 1280×1152 ≈
      // **7,5 MB trên GPU**, mười cái là 75 MB — một phần ba ngân sách 256 MB của iPhone,
      // tiêu cho những nhân vật người chơi có thể không gặp lần nào. `assets.charAtlas` đã
      // tự yêu cầu tải đúng nhân vật nào vừa lọt vào khung hình, nên chờ tới lúc đó.
      const bg = ['fx_high', 'audio'];
      if (!isMobile()) {
        bg.unshift(...assets.chars.map((c) => `char_${c}`));
      }
      assets.loadBackground(bg);
      if (pending || resume) {
        // "Đang trở lại trận…" thay cho "Đang tìm phòng…" (§7.1): người chơi vừa F5 giữa
        // trận cần biết mình đang QUAY LẠI chứ không phải bắt đầu lại.
        this.loading.step(resume ? STR.rejoining : STR.loadRoom);
      }
      const send = (): void => {
        if (!this.conn.isOpen) {
          window.setTimeout(send, 300);
          return;
        }
        // `?debug=true` chỉ là lời XIN. Server mới quyết (ADMIN_SPEC §1.1): nó xét phiên
        // đăng nhập admin trong cookie rồi trả `debug` trong gói Welcome, và client chỉ bật
        // phím test khi server nói có (xem `setWorld` → `this.debug = w.debug`).
        //
        // Không có phiên admin thì game chạy BÌNH THƯỜNG, không báo lỗi gì — người lạ gõ
        // `?debug=true` không được biết là có chế độ debug tồn tại.
        const wantDebug = new URLSearchParams(location.search).get('debug') === 'true';
        this.conn.send({ t: 'join', name: p.name, char: p.char, guest_id: guestId(), lang: LANG, debug: wantDebug, rs: resume });
      };
      send();
    })();
  }

  private onMessage(m: ServerMsg): void {
    switch (m.t) {
      case 'welcome':
        void this.onWelcome(m);
        break;
      case 'snap':
        if (this.joining) {
          this.early.push(m);
        } else if (this.playing) {
          this.applySnapshot(m);
        }
        break;
      case 'lb':
        this.hud.setLeaderboard(m);
        break;
      case 'rb':
        this.hud.setRoomBoard(m, this.world.myId, this.world.myName());
        break;
      case 'feed':
        this.hud.addFeed(m);
        break;
      case 'dead':
        this.onDead(m);
        break;
      case 'error':
        if (m.msg === 'bad_name') {
          this.showHome();
        }
        break;
      case 'kick':
        this.onKicked(m);
        break;
      default:
    }
  }

  private srvMs = 0;

  private applySnapshot(m: Snapshot): void {
    this.srvMs = m.sv ?? this.srvMs;
    const events = this.world.applySnapshot(m);
    this.scene.handleEvents(events);
    this.hud.onEvents(events, this.world.myId);
    // Nút kỹ năng đọc số lượt từ chính snapshot của mình (spec §6.3).
    const me = this.world.me();
    if (me) {
      // Hồi chiêu + trần lượt theo nhân vật, Tiên giới +1 (spec v6.25 §6.3). Tính ở client
      // chỉ để VẼ — server mới là bên đếm lượt.
      const sk = this.cfg.skills;
      const char = this.profile?.char ?? '';
      const [cdSecs, baseMax] = sk.cooldowns[char] ?? [0, 0];
      const maxC = baseMax + (tienIndex(this.cfg, realmOf(this.cfg, me.L)) >= 0 ? sk.tien_extra_charges : 0);
      const remain = me.ska ? Math.max(0, me.ska - this.world.serverNow()) : 0;
      this.hud.setSkill({ char, charges: me.skc ?? 0, maxCharges: maxC, remainSecs: remain, cooldownSecs: cdSecs });
    }
  }

  private async onWelcome(w: Welcome): Promise<void> {
    if ((w.pv ?? 0) !== PROTOCOL_VERSION || (w.sh ?? '') !== SCHEMA_HASH) {
      // Server và client nói hai thứ tiếng khác nhau. Dừng hẳn ở đây và nói rõ, thay vì
      // đọc rác rồi treo máy — chính là chuyện đã xảy ra một lần.
      //
      // Kiểm CẢ `sh` (băm schema nhị phân): số phiên bản là thứ người ta quên tăng, còn
      // băm thì đổi tự động mỗi khi `proto/schema.txt` đổi một dòng. Hai lớp vì hậu quả
      // của việc lọt là client đọc rác — không lỗi, không thông báo, chỉ nhân vật nhảy lung
      // tung hoặc màn hình đứng.
      //
      // Một NÚT, không phải một câu chữ: người đang mở tab từ hôm qua không nên phải tự
      // nghĩ ra Ctrl+F5.
      this.conn.stop();
      const panel = h('div', 'panel', h('p', '', STR.versionMismatch));
      const row = h('div', 'row');
      row.append(button(STR.versionReload, 'btn primary', () => location.reload()));
      panel.append(row);
      this.screens.show(h('div', 'screen center', panel));
      return;
    }
    // Bộ giải mã snapshot nhị phân cần biết "mình là ai" (để nhận SelfState) và bảng nhân
    // vật (để đổi chỉ số về tên).
    this.conn.snap.myId = w.id;
    this.conn.snap.chars = this.cfg.characters;
    // Mã phiên MỚI cho lần mất kết nối tiếp theo (§7.1). Lưu ngay, trước cả khi tải art
    // của khu: F5 đúng lúc màn hình tải cũng phải vào lại được.
    if (w.sk && this.profile) {
      saveSession(w.sk, this.profile);
    }
    // Nhân vật của chính mình không bao giờ bị nhả khỏi bộ nhớ đồ hoạ (Q30).
    this.scene.myChar = this.profile?.char ?? '';
    // Quên entity của khu CŨ.
    //
    // `welcome` đến giữa trận nghĩa là vừa qua cổng (§8.0). Bộ giải mã nhị phân giữ trạng
    // thái entity giữa các gói để tính delta; không xoá thì nó tiếp tục vẽ ra người và quái
    // của khu vừa rời, cho tới gói keyframe kế tiếp.
    this.conn.snap.reset();
    // Welcome đến GIỮA TRẬN nghĩa là vừa qua cổng sang khu khác (spec §8.0, §8.4): sang
    // khu là đổi hẳn bản đồ nên phải tải lại từ đầu, nhưng KHÔNG phải vào trận lại — giữ
    // nguyên hồ sơ, không phát lại âm thanh hồi sinh, không hiện lại hướng dẫn.
    const changedZone = this.playing;
    this.joining = true;
    this.early = [];
    const res = await fetch(`/api/map/${w.map}`);
    const map = (await res.json()) as GameMap;
    // Vật cản của khu này (§13.7 mục 3). Lần đầu vào trận thì CHỜ cho xong — nó nằm trong
    // ngân sách 8 MB và vào trận mà bản đồ toàn ô vuông placeholder thì không chơi được.
    // Qua cổng giữa trận thì chỉ chờ **tối đa 2 giây** rồi vào với placeholder, đúng câu
    // chữ của spec: người chơi đang bị dí, giữ họ ở màn hình dịch chuyển lâu hơn thế là
    // tệ hơn hẳn so với vài vật cản hiện muộn.
    await this.loadZoneArt(`zone_${w.map}`, w.znm ? STR.zoneName(w.zn, w.znm) : '', changedZone ? ZONE_ART_WAIT_MS : 0);
    this.world.reset(w, map);
    this.scene.setWorld(this.world);
    this.hud.minimap.setMap(map);
    this.debug = w.debug;
    this.hud.setDebug(w.debug);
    this.hud.setChar(this.profile?.char ?? '');
    {
      const char = this.profile?.char ?? '';
      const [cd, maxC] = this.cfg.skills.cooldowns[char] ?? [0, 0];
      this.hud.setSkill({ char, charges: 0, maxCharges: maxC, remainSecs: 0, cooldownSecs: cd });
    }
    const g = window as unknown as {
      kx?: unknown;
      __stats?: unknown;
      __playing?: unknown;
      __rb?: unknown;
      __texsites?: unknown;
    };
    // `__stats` KHÔNG khoá sau quyền debug: nó chỉ là số đo (FPS, bộ nhớ texture, số
    // sprite, ping) — không có một quyền nào lên luật chơi. Khoá nó lại thì bộ đo hiệu
    // năng không chạy được trên server THẬT, mà server thật là nơi duy nhất đo đúng: đó
    // là bản build người chơi nhận, với bản đồ thật và bot thật.
    //
    // `kx` thì vẫn khoá: nó cầm thẳng `scene`, `world`, `conn` — từ đó sửa được trạng
    // thái, và ADMIN_SPEC §1.1 nói rõ quyền ấy do phiên admin quyết.
    g.__stats = (): Record<string, number> => ({
      ...this.scene.perfStats(),
      ping: this.conn.rttMs,
      kbPerSec: this.conn.kbPerSec,
      snapKB: this.conn.lastSnapBytes / 1024,
      srvMs: this.srvMs,
      // Khu đang đứng: công cụ đo cần vào ĐÚNG một khu (khu tệ nhất) mới so được giữa các lần.
      zone: this.zoneIndex,
    });
    g.__playing = (): boolean => this.playing;
    // `__texsites`: texture đang giữ gom theo nơi cấp phát. Cùng loại với `__stats` — chỉ
    // là số đo, không cầm trạng thái game. Chỉ có dữ liệu khi mở trang với `?textrace=1`.
    g.__texsites = (n = 12): { site: string; count: number; mb: number }[] => this.scene.gpuSites(n);
    // `__rb`: số đếm của bảng xếp hạng cho bài §17 `tools/browsertest/rbstable.mjs` — gói
    // nhận được mấy dòng, có gói nào thiếu bị bỏ không. Cùng loại với `__stats`: chỉ số đo.
    g.__rb = (): Record<string, number> => ({ ...rbStats });
    if (w.debug) {
      // ?debug=true: mở cửa sau cho console của trình duyệt (xem VFX, đếm thần thú...).
      //
      // `playing` phải hỏi qua hàm: `world.me()` vẫn trả về dữ liệu CŨ sau khi chết và rời
      // phòng, nên công cụ đo tưởng nhân vật còn sống và không bấm "Chơi tiếp" (mất mấy
      // lượt đo vì chuyện này).
      g.kx = {
        scene: this.scene, world: this.world, conn: this.conn, playing: () => this.playing,
        // LỖI GIẢ để nghiệm thu §13.8 (checklist §17). Không có cửa này thì lớp chống sập
        // chỉ được thử lần đầu bằng một lỗi thật, giữa trận, trên máy người chơi.
        //   kx.fault('vfx')      → hệ thống VFX ném lỗi ở khung sau
        //   kx.fault('vfx', 3)   → ném 3 khung liên tiếp (đủ ngưỡng "tắt hẳn")
        //   kx.faultDraw(3)      → ném trong khâu VẼ 3 lần (đủ ngưỡng "dựng lại renderer")
        fault: (key: SysKey, n = 1): void => faultSystem(key, n),
        faultDraw: (n = 1): void => this.scene.faultDraw(n),
        guard: (): unknown => ({ systems: sysSummary(), draw: drawGuardStats() }),
        // Đổi mức chất lượng NGAY trong phiên, cho bài đo `daily.mjs`. Người chơi đổi bằng
        // bảng Cài đặt; bài đo cần một cửa không phụ thuộc bố cục giao diện, vì một lần đổi
        // tên nút không được phép làm mất cả một nhánh nghiệm thu.
        quality: (q: 'high' | 'medium' | 'low'): void => setQuality(q),
      };
    }
    this.screens.show(null);
    this.loading.done();
    this.telemetry?.setPhase('match');
    this.overlay.appendChild(this.hud.root);
    this.playing = true;
    this.joining = false;
    for (const snap of this.early) {
      this.applySnapshot(snap);
    }
    this.early = [];
    this.input.enabled = true;
    this.zoneIndex = w.zn ?? 0;
    this.zoneAt = performance.now();
    this.hud.setZone(w.znm ? STR.zoneName(w.zn, w.znm) : '', w.zn ?? 0);
    // Bắt đầu thu số liệu từ lần vào trận đầu tiên; sang khu khác cũng gửi `welcome` nhưng
    // `startTelemetry` tự bỏ qua lần hai, vì một PHIÊN là một lần chơi chứ không phải một khu.
    void this.startTelemetry();
    if (this.telemetry && changedZone) {
      this.telemetry.event('zone_switch', String(w.zn ?? 0));
      this.telemetry.breadcrumb('khu', w.zn ?? 0);
    }
    if (changedZone) {
      // Chữ lớn mờ dần giữa màn hình khi vào khu mới (spec §8.0).
      this.hud.showBanner(STR.zoneEntered(STR.zoneName(w.zn, w.znm ?? '')), ZONE_BANNER_SECS);
      return;
    }
    this.hud.showGuide(); // hướng dẫn mờ trong lúc còn bất tử (spec §18.1)
    sound.play('respawn');
  }

  /**
   * loadZoneArt tải nhóm art của một khu. `capMs > 0` thì chờ tối đa ngần đó rồi đi tiếp
   * (phần còn lại vẫn tải nền và tự hiện ra khi xong).
   */
  private async loadZoneArt(group: string, zoneName: string, capMs: number): Promise<void> {
    if (!assets.pending([group]).length) {
      return;
    }
    this.loading.show();
    this.loading.step(STR.loadZone(zoneName || group));
    const job = assets.load([group], (pr) => this.loading.bytes(pr));
    if (capMs > 0) {
      // Hết hạn thì thả ra chứ KHÔNG huỷ: `job` cứ chạy tiếp ở nền, và `assets.gen` tăng
      // khi xong nên bộ vẽ tự đổi placeholder sang ảnh thật.
      await Promise.race([job, new Promise((r) => window.setTimeout(r, capMs))]);
    } else {
      await job.catch((err: unknown) => console.warn('tải art khu lỗi', err));
    }
    this.loading.done();
  }

  private onDead(d: DeadMsg): void {
    this.playing = false;
    this.input.enabled = false;
    this.hud.root.remove();
    // Ván đã kết thúc: mã phiên hết nghĩa. Giữ lại thì một lần F5 sau đó gửi một mã chết
    // và người chơi thấy màn hình tải chớp một nhịp trước khi về Tổng kết (§7.1).
    clearSession();
    const p = this.profile;
    if (p) {
      saveEndedRun(d, p);
    }
    this.screens.show(
      summaryScreen(this.cfg, d, p?.char ?? '', {
        playAgain: () => p && this.join(p),
        leaderboard: () => this.showLeaderboard(() => this.onDead(d)),
        changeCharacter: () => {
          clearEndedRun();
          this.showHome();
        },
      }),
    );
  }

  /** skillPressed: nút kỹ năng vừa được bấm, chờ gói input kế tiếp mang đi (spec §6.3). */
  private skillPressed = false;

  /** bindSkillButton: nút kỹ năng trên điện thoại — bấm một phát, không phải giữ. */
  private bindSkillButton(): void {
    const b = this.hud.skillBtn;
    const press = (e: Event): void => {
      this.skillPressed = true;
      e.preventDefault();
    };
    b.addEventListener('touchstart', press, { passive: false });
    b.addEventListener('mousedown', press);
  }

  /** bindBurnButton makes the on-screen 🔥 a hold button (spec §4.1c). */
  /**
   * setupFeedback gắn nút 💬 và nối nó với ngữ cảnh của trận (ADMIN_SPEC §7.1).
   *
   * Ngữ cảnh lấy LÚC BẤM GỬI chứ không lúc mở form: người chơi có thể mở form rồi mô tả
   * một lúc, và cái cần biết là tình trạng lúc họ gặp vấn đề — tức lúc gần nhất có thể.
   */
  private setupFeedback(): void {
    feedback.installErrorCapture();
    feedback.configure(
      () => ({
        guest_id: guestId(),
        name: this.profile?.name ?? '',
        build: BUILD_ID,
        zone: this.world.map?.id ?? '',
        level: this.world.me()?.L ?? -1,
        swords: this.world.me()?.n ?? -1,
        ping: Math.round(this.conn.rttMs),
        fps: Math.round(this.lastFps),
        playing: this.playing,
      }),
      () => this.screenshot(),
    );
    feedback.mountButton(this.hud.root, 'match');
  }

  /**
   * screenshot chụp canvas ra data URL (§7.1).
   *
   * Không tự ép ngân sách ở đây nữa: form góp ý đưa ảnh này qua CÙNG đường nén với ảnh tải
   * lên (`ui/imagecompress.ts` — thu nhỏ ≤ 1 600 px, WebP/JPEG ≤ 300 KB). Ép ở đây thì
   * Safari (chụp ra PNG vì không mã hoá được WebP) hay trả null với màn hình lớn.
   * Trả null nếu không chụp được — form vẫn gửi được phần chữ.
   */
  private screenshot(): string | null {
    try {
      const cv = this.scene.app.canvas as HTMLCanvasElement | null;
      if (!cv) {
        return null;
      }
      // Vẽ lại một khung NGAY trước khi đọc: WebGL không giữ bộ đệm khung hình
      // (`preserveDrawingBuffer` tắt — bật thì tốn mọi khung hình), nên đọc ngoài nhịp vẽ
      // ra ảnh đen. Vẽ rồi đọc trong cùng một lượt chạy JS thì bộ đệm còn nguyên.
      this.scene.app.render();
      return cv.toDataURL('image/webp', 0.9);
    } catch (_) {
      return null; // canvas "bẩn" hoặc WebGL mất ngữ cảnh
    }
  }

  private bindBurnButton(): void {
    const b = this.hud.burnBtn;
    const set = (on: boolean) => (e: Event): void => {
      this.input.burnButton = on;
      e.preventDefault();
    };
    b.addEventListener('touchstart', set(true), { passive: false });
    b.addEventListener('touchend', set(false));
    b.addEventListener('touchcancel', set(false));
    b.addEventListener('mousedown', set(true));
    window.addEventListener('mouseup', () => (this.input.burnButton = false));
  }

  private sendInput(): void {
    if (!this.playing) {
      return;
    }
    const { angle, moving, burn } = this.input.read();
    // Nút kỹ năng gom lại từ lúc gửi gói trước: bấm giữa hai nhịp mạng cũng không mất.
    const skill = this.skillPressed;
    if (skill) {
      this.telemetry?.breadcrumb('ky-nang');
    }
    this.skillPressed = false;
    this.conn.send(this.world.makeInput(angle, moving, burn, skill));
    this.askMissingDesc();
  }

  /** descAskAt: lần cuối gửi `DescRequest`, để không hỏi dồn dập. */
  private descAskAt = 0;

  /**
   * askMissingDesc hỏi server mô tả của những entity client chưa biết vẽ gì
   * (NETCODE_SPEC §5 Hiển thị mục 2).
   *
   * Giới hạn 1 lần / 0.5 giây: entity vừa vào khung nhìn thường có mô tả ngay ở gói kế
   * tiếp, nên hỏi mỗi nhịp input (30 lần/giây) là bắn 30 gói thừa cho một thứ sắp tự đến.
   */
  private askMissingDesc(): void {
    const now = performance.now();
    if (now - this.descAskAt < 500) {
      return;
    }
    const ids = this.world.takeMissingDesc();
    if (ids.length === 0) {
      return;
    }
    this.descAskAt = now;
    this.conn.send({ t: 'dreq', ids: ids.slice(0, 32) });
  }

  private debugKey(e: KeyboardEvent): void {
    if (!this.playing || !this.debug || e.repeat) {
      return;
    }
    if (e.code === 'KeyH') {
      this.scene.toggleHitbox();
      return;
    }
    const cmd = DEBUG_KEYS[e.code];
    if (cmd) {
      this.conn.send({ t: 'dbg', ...cmd });
      return;
    }
    const n = Number(e.key);
    if (n >= 1 && n <= 7) {
      this.conn.send({ t: 'dbg', cmd: 'buff', n: n - 1 });
    }
  }

  private frame(dt: number): void {
    if (!this.playing) {
      return;
    }
    this.scene.render(dt);
    const me = this.world.me();
    // HUD và minimap là hai ranh giới riêng (§13.8): HUD hỏng thì game vẫn chơi được, chỉ
    // mất thanh Linh khí — dừng cả vòng vẽ vì nó mới là hỏng thật.
    sysRun('hud', () => {
      this.hud.update(me, this.conn.rttMs, (L) => ({ from: this.thresholds[L], to: this.thresholds[L + 1] }));
      this.hud.setStatuses(this.scene.myStatusChips(performance.now() / 1000));
    });
    // F3 mở được bảng đo ở MỌI trang (spec §13.6), không chỉ khi `?debug=true` — trước đây
    // mở F3 trên trang thường thì bảng hiện ra nhưng số không bao giờ cập nhật.
    // Đồng hồ theo hệ thống chỉ chạy khi có người XEM bảng — tắt thì mỗi mốc chỉ là một
    // phép so boolean (render/probe.ts).
    this.scene.setProbe(this.debug || this.hud.perfVisible);
    if (this.debug || this.hud.perfVisible) {
      this.hud.setPerf(this.scene.perfLine(dt, {
        rttMs: this.conn.rttMs,
        kbPerSec: this.conn.kbPerSec,
        snapBytes: this.conn.lastSnapBytes,
        srvMs: this.srvMs,
      }));
    }
    if (me) {
      sysRun('minimap', () => {
        const view = this.scene.viewSize();
        this.hud.minimap.render(me.x, me.y, me.f, view.w, view.h, performance.now() / 1000);
      });
    }
    // Telemetry ĐỨNG CUỐI vòng vẽ và chỉ ghi một con số: mọi thứ đắt hơn nó tự dồn về 4 Hz.
    sysRun('telemetry', () => this.telemetry?.frame());
  }

  /**
   * startTelemetry nạp module thu số liệu — chỉ khi build bật cờ `TELEMETRY=on`.
   *
   * `import()` động nằm sau một hằng số biên dịch: bản release có `__TELEMETRY__ === false`
   * nên Rollup cắt hẳn nhánh này và KHÔNG đóng gói module (yêu cầu §1.6 — "module không
   * được đóng gói vào"). Kiểm bằng `npm run check:telemetry-off`.
   */
  private async startTelemetry(): Promise<void> {
    if (!__TELEMETRY__ || this.telemetry) {
      return;
    }
    const { Telemetry } = await import('./telemetry');
    const { setErrorDisplay } = await import('./telemetry/early');
    // `?debug=true`: mọi lỗi hiện thẳng lên màn hình. iPhone không mở được console nên đây
    // là cách duy nhất người cầm máy đọc được lỗi rồi chụp màn hình gửi về.
    if (new URLSearchParams(location.search).get('debug') === 'true') {
      const box = document.getElementById('kxerr');
      setErrorDisplay((msg) => {
        if (box) {
          box.classList.add('on');
          box.textContent = `${new Date().toLocaleTimeString()} ${msg}\n${box.textContent ?? ''}`.slice(0, 4000);
        }
      });
    }
    const t = new Telemetry(
      {
        perfStats: () => this.scene.perfStats(),
        net: () => ({ rttMs: this.conn.rttMs, kbPerSec: this.conn.kbPerSec, snapBytes: this.conn.lastSnapBytes }),
        context: () => this.telemetryContext(),
        send: (msg) => {
          // `Conn.send` tự bỏ gói khi chưa nối; telemetry cần BIẾT điều đó để còn giữ mẫu
          // lại gửi bù (§3), nên hỏi trạng thái trước rồi mới gửi.
          if (!this.conn.isOpen) {
            return false;
          }
          this.conn.send(msg);
          return true;
        },
        canvasSize: () => {
          const c = this.scene.app.canvas as HTMLCanvasElement;
          return { w: c.width, h: c.height };
        },
        guard: () => this.scene.guardStats(),
        sinceZone: () => (this.zoneAt > 0 ? (performance.now() - this.zoneAt) / 1000 : 0),
      },
      BUILD_ID,
      PROTOCOL_VERSION,
    );
    t.start();
    this.telemetry = t;
    // Mỗi lần lớp chống sập phải tắt một hệ thống thì server phải BIẾT (§13.8). Người chơi
    // sẽ không báo: với họ game vẫn chạy, chỉ là "hình như thiếu hiệu ứng gì đó".
    setSysErrorHook(({ key, msg, err, dead, errors }) => {
      // Lỗi do bài kiểm bắn ra đi bằng tên khác (`..._gia`): §13.8 vẫn được nghiệm thu đủ,
      // mà bảng lỗi và cảnh báo thì chỉ còn chuyện THẬT.
      t.event(isFake(err) ? 'client_sys_off_gia' : 'client_sys_off',
        `${key}: ${msg}\n${JSON.stringify({ dead, errors, ...t.errorContext() })}`);
    });
    setUiErrorHook((err, label) => t.errorEvent(`nút UI "${label}": ${err instanceof Error ? err.message : String(err)}`)); // i18n-ok: chữ gửi về telemetry
    // Tự hạ chất lượng vì chạm ngân sách bộ nhớ GPU: ghi lại, vì đó là dấu hiệu sớm nhất
    // của việc tab sắp bị iOS giết — và là thứ duy nhất báo trước.
    this.scene.onMemoryPressure = (mb) => {
      t.event('gpu_pressure', JSON.stringify({ mb: Math.round(mb), ...t.errorContext() }));
    };
    // Mất ngữ cảnh WebGL là ngưỡng cảnh báo §4.3 — chính là lỗi đội test gặp ngày 30/10.
    const canvas = this.scene.app.canvas as HTMLCanvasElement;
    canvas.addEventListener('webglcontextlost', () => t.event('webgl_context_lost'));
    canvas.addEventListener('webglcontextrestored', () => t.event('webgl_context_restored'));
  }

  /** telemetryContext gom bối cảnh trận (§2.2 nhóm cuối). Chỉ ĐỌC trạng thái game. */
  private telemetryContext(): TelemetryContext {
    const me = this.world.me();
    const draw = this.scene.drawStats();
    // Số người Tiên giới trong khung nhìn: spec §4.1 mục 4 tách riêng nhóm này vì họ mang
    // vòng kiếm nhiều lớp và hào quang — vài người Tiên giới cùng khung là tình huống nặng
    // nhất mà game tạo ra, và cũng là chỗ đáng ngờ nhất khi ai đó báo tụt FPS.
    let tien = 0;
    const tienFrom = this.cfg.cultivation.realms.length; // cảnh giới >= số cảnh giới thường
    for (const p of this.world.players.values()) {
      const last = p.snaps[p.snaps.length - 1];
      if (last && last.s.L >= tienFrom) {
        tien++;
      }
    }
    return {
      zone: this.zoneIndex,
      realm: me?.L ?? 0,
      swords: me?.n ?? 0,
      rings: me ? ringGeom(this.cfg, me.sc, me.n).rings : 0,
      charsInView: draw.charsInView,
      tienInView: tien,
      swordsDrawn: draw.swordsDrawn,
      // vfx: chưa đo được — lớp VFX không giữ bộ đếm số hiệu ứng đang chạy. Để 0 chứ không
      // đoán: một con số bịa trong báo cáo hiệu năng còn tệ hơn một ô trống.
      vfx: 0,
      burning: ((me?.fl ?? 0) & FLAG.burning) !== 0,
      zoning: this.joining,
      placeholderNow: draw.placeholder,
      respawn2s: this.world.takeRespawn2s(),
      // Lấy TỔNG trước rồi mới lấy bảng chia: cả hai đều "lấy rồi xoá", đảo thứ tự không
      // đổi kết quả, nhưng để cạnh nhau cho rõ chúng đếm cùng một sự kiện.
      respawnBy: this.world.takeRespawnBy(),
      quality: STR.quality[quality()] ?? '',
      qualityAuto: false,
    };
  }
}

/**
 * boot: mở trang chủ NGAY, rồi tải art nền trong lúc người chơi gõ tên (chủ dự án chốt 31/10).
 *
 * Ba bản, để đừng ai quay lại bản nào trước:
 * 1. Bản gốc gọi `assets.preload()` — tải **toàn bộ** manifest, 547 tệp / 77,4 MB, trước khi
 *    người chơi thấy bất cứ thứ gì. Đó là nguyên nhân "chờ tải rất lâu".
 * 2. Bản §13.7 chặn trang chủ ở nhóm `home` (0,86 MB) — 4,0 giây ngồi nhìn thanh tiến độ.
 * 3. Bản này **không chặn gì cả**. Trang chủ chỉ cần `/api/config` + `manifest.json`; ảnh
 *    chân dung và nền là `<img>` / `background-image` nên trình duyệt tự tải song song. Mọi
 *    thứ còn lại (`home`, `boot`, font, nhân vật đã lưu) tải NỀN trong lúc người chơi gõ
 *    tên và chọn nhân vật — khoảng thời gian trước đây bỏ không.
 *
 * Kết quả: thời gian chờ thật của người chơi không còn là "mở trang → vào trận" mà là
 * "bấm Vào chơi → vào trận", và phần lớn gói đã về trước lúc họ bấm.
 */
/**
 * Nút "Góp ý / báo lỗi" ở CHÂN TRANG (chủ dự án chốt 1/11).
 *
 * Nút nằm trong `index.html` dạng HTML tĩnh, nên nó có mặt trước cả khi bundle tải xong.
 * Gắn trình xử lý ở đây — không gắn trong `feedback.ts` cùng với nút 💬 nổi: nút kia do
 * `mountButton` dựng và sống/chết theo màn hình, còn nút này nằm sẵn trong trang.
 *
 * Bấm trước khi bundle tải xong thì không xảy ra gì và đó là hành vi đúng: hộp góp ý cần
 * ngữ cảnh (phiên bản, thiết bị, khu, FPS) mà chỉ bundle mới biết.
 */
function wireFooterFeedback(): void {
  document.getElementById('foot-feedback')?.addEventListener('click', () => feedback.open());
}

async function boot(): Promise<void> {
  const overlay = document.getElementById('overlay') as HTMLElement;
  const stage = document.getElementById('stage') as HTMLElement;
  const screen = new LoadingScreen();
  wireFooterFeedback();
  const run = async (): Promise<void> => {
    const cfg = await loadConfig();
    await assets.init();
    // `?scene=testbed`: cảnh thử cố định để so ảnh trước/sau một thay đổi đồ hoạ
    // (`render/testbed.ts`). Quyền do SERVER quyết — `/api/debug-allowed` dùng đúng
    // `admin.DebugGate` mà `/ws` dùng (ADMIN_SPEC §1.1). Tự xét ở client thì bất kỳ ai gõ
    // thêm một tham số vào URL cũng vào được.
    // Cả phần hỏi quyền lẫn câu từ chối nằm TRONG `testbed.ts` (tệp đã `i18n-ok-file`):
    // thêm một chuỗi tiếng Việt vào `strings.ts` cho một màn hình chỉ công cụ đo nhìn thấy
    // là thêm một khoá chưa dịch vào cả 5 bản dịch khác.
    if (new URLSearchParams(location.search).get('scene') === 'testbed') {
      const { runTestbed } = await import('./render/testbed');
      await runTestbed(stage, cfg);
      screen.done();
      return;
    }
    // Thứ tự nhân vật dự phòng = thứ tự trong config (NETCODE_SPEC §5 Hiển thị mục 1).
    // Thiếu atlas của một nhân vật thì vẽ bằng nhân vật đầu tiên CÓ atlas, không vẽ
    // placeholder — build thật cấm placeholder.
    assets.setFallbackChars(cfg.characters);
    sound.init();
    startBgLoadIndicator();
    // Font bắt đầu tải ngay nhưng KHÔNG chặn trang chủ: trang chủ là DOM, font hệ thống vẽ
    // được. Chỉ bảng tên trong trận cần BitmapFont, nên `join()` mới chờ nó.
    const fonts = loadFonts();
    // Tải nền ngay: `home` (texture giao diện cho Pixi) rồi `boot` (nền, vật cản, cổng,
    // quái) rồi atlas của nhân vật đang lưu — đúng thứ tự người chơi sẽ cần tới.
    const saved = savedProfile(cfg);
    const artReady = assets.load(['home']);
    void artReady.catch(() => undefined);
    void (async () => {
      try {
        await artReady;
        await assets.load(['boot', `char_${saved.char}`]);
      } catch (err) {
        // Tải nền thất bại thì KHÔNG hỏng gì ở đây: `join()` tải lại đúng phần còn thiếu
        // và lúc đó mới có thanh tiến độ + nút "Thử lại".
        console.warn('tải nền trước khi vào trận lỗi', err);
      }
    })();
    screen.done();
    await new App(cfg, overlay, stage, screen, fonts, artReady).start(stage);
  };
  try {
    await run();
  } catch (err) {
    console.error(err);
    // Lỗi mạng ở bước này: cho bấm "Thử lại" và tải TIẾP, thay vì tự tải lại trang sau 3
    // giây như bản cũ — tải lại trang là mất sạch phần đã tải xong.
    screen.fail(() => {
      void run().catch((e: unknown) => {
        console.error(e);
        screen.fail(() => location.reload());
      });
    });
  }
}

void boot();
