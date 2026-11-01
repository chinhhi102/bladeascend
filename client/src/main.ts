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

import { FLAG, PROTOCOL_VERSION } from './net/protocol';
import { SCHEMA_HASH } from './net/wire.gen';
import { pendingRejoin, watchContextLoss } from './render/glrecover';
import type { DeadMsg, GameMap, KickedMsg, ServerMsg, Snapshot, Welcome } from './net/protocol';
import { ClientWorld } from './net/state';
import { assets } from './render/assets';
import { isMobile } from './render/device';
import { loadFonts } from './render/fonts';
import { GameScene } from './render/scene';
import { loadJitter } from './render/jitter';
import { cycleQuality, quality, QUALITY_NAME } from './render/quality';
import { ringGeom } from './render/ringgeom';
import type { Telemetry, TelemetryContext } from './telemetry';
import { Screens, button, h } from './ui/dom';
import { LoadingScreen, startBgLoadIndicator } from './ui/loading';
import { kickedScreen } from './ui/kicked';
import { guestId, homeScreen, type Profile, savedProfile } from './ui/home';
import { Hud } from './ui/hud';
import { leaderboardScreen } from './ui/leaderboard';
import * as feedback from './ui/feedback';
import { STR } from './ui/strings';
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
  KeyG: { cmd: 'skill', n: 1 }, // +1 lượt kỹ năng, khỏi phải đi nhặt Bí Kíp
};

/** fpsCap: 60 on PC, 45 on phones to save battery; override with ?fps=30 (M7). */
function fpsCap(): number {
  const q = Number(new URLSearchParams(location.search).get('fps'));
  if (q > 0) {
    return q;
  }
  return Input.isTouchDevice() ? 45 : 60;
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
  /** animInfo: phím J đang bật hay tắt (GAME_SPEC §12.5 C). */
  private animInfo = false;

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
    this.thresholds = levelThresholds(cfg);
    this.setupFeedback();
  }

  async start(stage: HTMLElement): Promise<void> {
    await this.scene.init(stage, fpsCap());
    this.scene.app.ticker.add((t) => {
      // FPS trượt mượt để đính kèm vào góp ý — số tức thời nhảy quá, đọc không ra gì.
      this.lastFps = this.lastFps * 0.9 + (1000 / Math.max(t.deltaMS, 1)) * 0.1;
      this.frame(t.deltaMS / 1000);
    });
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
        const name = QUALITY_NAME[pick.q];
        this.hud.showBanner(STR.qualitySet(pick.auto ? STR.qualityAuto(name) : name), 2.5);
        return;
      }
      this.debugKey(e);
    });
    this.conn.connect(
      () => undefined,
      () => {
        if (this.playing) {
          this.playing = false;
          this.input.enabled = false;
          this.hud.root.remove();
          this.screens.show(h('div', 'screen center', h('div', 'panel', STR.disconnected)));
          window.setTimeout(() => this.showHome(), 2500);
        }
      },
    );
    // Mất ngữ cảnh WebGL (trình điều khiển card reset, máy ngủ dậy): canvas sẽ đen vĩnh
    // viễn nếu không làm gì. Nhớ lại tên + nhân vật rồi tải lại trang và tự vào lại phòng.
    watchContextLoss(
      this.scene.app.canvas as HTMLCanvasElement,
      () => this.profile,
      (canRecover) => {
        // Gửi về server KÈM BỐI CẢNH: mất ngữ cảnh WebGL trên iPhone gần như luôn là hệ quả
        // của hết bộ nhớ GPU, nên con số `texMB` đi cùng mới nói được nguyên nhân.
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
        this.showHome();
      },
    );
    // Bắt đầu thu số liệu NGAY khi game khởi động xong, không đợi vào trận.
    //
    // Trước đây telemetry chỉ bật trong `onWelcome`, nên **máy nào không vào nổi trận thì
    // không để lại dấu vết nào cả** — đúng tình huống đang phải tìm: nhật ký server ngày
    // 31/10 cho thấy một người vào lại phòng 8 lần trong 40 giây, mà bảng `sessions` không
    // có lấy một phiên điện thoại nào. Thiết bị hỏng nặng nhất lại là thiết bị im lặng nhất.
    void this.startTelemetry();

    // Vừa tải lại vì mất đồ hoạ thì vào thẳng phòng, đừng bắt bấm lại "Vào chơi".
    const again = pendingRejoin();
    if (again) {
      this.join(again);
    } else {
      this.showHome();
    }
  }

  /** onKicked: bị mời ra hoặc bị cấm (ADMIN_SPEC §6.2). */
  private onKicked(m: KickedMsg): void {
    // `stop()` chứ không phải đóng thường: nó tắt luôn việc tự nối lại. Đóng thường thì
    // client nối lại sau vài giây, server cấm tiếp, và người vừa bị cấm quay vòng giữa
    // "bị cấm" và "đang kết nối" mà không đọc kịp lý do.
    this.conn.stop();
    this.telemetry?.setPhase('home');
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
  private join(p: Profile): void {
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
        this.loading.fail(() => this.join(p));
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
      if (pending) {
        this.loading.step(STR.loadRoom);
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
        this.conn.send({ t: 'join', name: p.name, char: p.char, guest_id: guestId(), debug: wantDebug });
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
      // Trần lượt theo cảnh giới: Tiên giới 3, phàm giới 2 (spec v6.5 §6.3). Tính ở client
      // chỉ để VẼ đủ số chấm — server vẫn là bên chặn khi nhặt thêm.
      const sk = this.cfg.skills;
      const maxC = tienIndex(this.cfg, realmOf(this.cfg, me.L)) >= 0 ? sk.max_charges_tien : sk.max_charges;
      this.hud.setSkill(this.profile?.char ?? '', me.skc ?? 0, maxC);
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
    await this.loadZoneArt(`zone_${w.map}`, w.znm ?? '', changedZone ? ZONE_ART_WAIT_MS : 0);
    this.world.reset(w, map);
    this.scene.setWorld(this.world);
    this.hud.minimap.setMap(map);
    this.debug = w.debug;
    this.hud.setDebug(w.debug);
    this.hud.setChar(this.profile?.char ?? '');
    this.hud.setSkill(this.profile?.char ?? '', 0);
    const g = window as unknown as { kx?: unknown; __stats?: unknown; __playing?: unknown };
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
    if (w.debug) {
      // ?debug=true: mở cửa sau cho console của trình duyệt (xem VFX, đếm thần thú...).
      //
      // `playing` phải hỏi qua hàm: `world.me()` vẫn trả về dữ liệu CŨ sau khi chết và rời
      // phòng, nên công cụ đo tưởng nhân vật còn sống và không bấm "Chơi tiếp" (mất mấy
      // lượt đo vì chuyện này).
      g.kx = { scene: this.scene, world: this.world, conn: this.conn, playing: () => this.playing };
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
    this.hud.setZone(w.znm ?? '', w.zn ?? 0);
    // Bắt đầu thu số liệu từ lần vào trận đầu tiên; sang khu khác cũng gửi `welcome` nhưng
    // `startTelemetry` tự bỏ qua lần hai, vì một PHIÊN là một lần chơi chứ không phải một khu.
    void this.startTelemetry();
    if (this.telemetry && changedZone) {
      this.telemetry.event('zone_switch', String(w.zn ?? 0));
    }
    if (changedZone) {
      // Chữ lớn mờ dần giữa màn hình khi vào khu mới (spec §8.0).
      this.hud.showBanner(STR.zoneEntered(w.znm ?? ''), ZONE_BANNER_SECS);
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
    const p = this.profile;
    this.screens.show(
      summaryScreen(this.cfg, d, p?.char ?? '', {
        playAgain: () => p && this.join(p),
        leaderboard: () => this.showLeaderboard(() => this.onDead(d)),
        changeCharacter: () => this.showHome(),
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
   * screenshot chụp canvas thành WebP ≤ 300 KB (§7.1).
   *
   * Hạ dần chất lượng cho tới khi lọt ngân sách thay vì chụp một lần rồi chịu: ảnh 4K ở
   * chất lượng 0.9 dễ vượt 1 MB, mà server từ chối thì người chơi mất cả bài góp ý vừa gõ.
   * Trả null nếu không chụp được — form vẫn gửi được phần chữ.
   */
  private screenshot(): string | null {
    try {
      const cv = this.scene.app.canvas as HTMLCanvasElement | null;
      if (!cv) {
        return null;
      }
      for (const q of [0.7, 0.5, 0.35, 0.2]) {
        const url = cv.toDataURL('image/webp', q);
        // 300 KB sau base64 ≈ 400 000 ký tự. Đo trên chuỗi là đủ chính xác ở đây.
        if (url.length < 400_000) {
          return url;
        }
      }
      return null;
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
    this.hud.update(me, this.conn.rttMs, (L) => ({ from: this.thresholds[L], to: this.thresholds[L + 1] }));
    if (this.debug) {
      this.hud.setPerf(this.scene.perfLine(dt, {
        rttMs: this.conn.rttMs,
        kbPerSec: this.conn.kbPerSec,
        snapBytes: this.conn.lastSnapBytes,
        srvMs: this.srvMs,
      }));
    }
    if (me) {
      const view = this.scene.viewSize();
      this.hud.minimap.render(me.x, me.y, me.f, view.w, view.h, performance.now() / 1000);
    }
    // Telemetry ĐỨNG CUỐI vòng vẽ và chỉ ghi một con số: mọi thứ đắt hơn nó tự dồn về 4 Hz.
    this.telemetry?.frame();
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
      },
      BUILD_ID,
      PROTOCOL_VERSION,
    );
    t.start();
    this.telemetry = t;
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
      quality: QUALITY_NAME[quality()] ?? '',
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
