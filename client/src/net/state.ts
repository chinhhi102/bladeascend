import { realmOf, type GameConfig } from '../config';
import { lerp, type Vec } from './geom';
import { bracket, Clock, followSnaps, type Track } from './interp.ts';
import { statusTracker } from '../render/statustrack';
import { moveStep } from './movestep.ts';
import { bonus, innate, stat } from './stat';
import {
  BUFFS,
  TIER_BORDER,
  type GameEvent,
  type GameMap,
  type ItemState,
  itemKind,
  type MonsterState,
  type PlayerState,
  type Shape,
  type Snapshot,
  statusOn,
  type Welcome,
} from './protocol';

interface Timed<T> {
  t: number;
  s: T;
}

export interface PlayerView {
  id: number;
  name: string;
  char: string;
  snaps: Timed<PlayerState>[];
  /** Trạng thái nội suy giữa các khung (net/interp.ts). Chính mình không dùng. */
  track?: Track;
}

interface PendingInput {
  seq: number;
  angle: number;
  moving: boolean;
  burn: boolean;
}

/** What the renderer needs for one player this frame. */
export interface PlayerFrame {
  view: PlayerView;
  s: PlayerState;
  x: number;
  y: number;
  ringAngle: number;
}

/** Lệch dưới ngần này là sai số dấu phẩy động, không phải hiệu chỉnh thật. */
const CORRECTION_EPS = 0.5;

/**
 * DESC_KEEP_SECS: giữ mô tả entity bao lâu sau khi `Despawn` (Entity Store, §4.3).
 *
 * 30 giây: đủ dài để bao trọn mọi lần đi vòng quanh mép khung nhìn, đủ ngắn để một trận dài
 * không tích thành chỗ rò bộ nhớ. Người rời phòng hẳn thì id không bao giờ quay lại.
 */
const DESC_KEEP_SECS = 30;

const SNAP_KEEP = 30;
/**
 * Giữ entity thêm ngần này giây sau khi vắng mặt khỏi snapshot (NETCODE_SPEC §4.3).
 *
 * 2 giây là con số spec chọn, và nó phải LỚN HƠN HẲN nhịp 10 Hz của người ở xa (0,1 giây)
 * — nếu không thì vẫn nhấp nháy, chỉ chậm hơn.
 */
const KEEP_GONE_SECS = 2;

/** ClientWorld holds the replicated world: interpolation for others, prediction for me. */
export class ClientWorld {
  myId = 0;
  map: GameMap | null = null;
  readonly players = new Map<number, PlayerView>();
  /** Trạng thái nội suy của quái — quái không có object riêng như người. */
  private readonly mtracks = new Map<number, Track>();
  readonly monsters = new Map<number, Timed<MonsterState>[]>();
  readonly items = new Map<number, ItemState>();
  readonly brokenObstacles = new Map<number, boolean>();
  /**
   * stalls: dãy sạp tạm của Cô Ba (spec v6.10 §6.3), theo id sạp.
   *
   * Không nằm trong `map.obstacles` vì chúng chỉ sống 6 giây và là của một phòng cụ thể —
   * xem `server/internal/game/stall.go` để biết vì sao server cũng tách chúng ra.
   */
  readonly stalls = new Map<number, { owner: number; tier: number; shape: Shape }>();
  /** goneAt: lần đầu một entity vắng mặt khỏi snapshot — xem `expire`. */
  private readonly goneAt = new Map<number, number>();
  /** deletedAt: lúc một entity bị xoá hẳn, để nhận ra nó quay lại ngay sau đó (nhấp nháy). */
  private readonly deletedAt = new Map<number, number>();
  /** respawn2s: số lần nhấp nháy chưa báo cáo — xem takeRespawn2s. */
  private respawn2s = 0;
  /**
   * whyGone[id] = lý do entity bị xoá, để biết nhấp nháy đến từ đâu.
   *
   * 'ds' = server bảo đi (Despawn tường minh), 'hethan' = luật 2 giây im lặng cũ. Hai
   * nguyên nhân khác hẳn nhau và cần hai cách chữa khác hẳn nhau — không tách ra thì chỉ
   * còn cách đoán, và tôi đã đoán sai hai lần liền.
   */
  readonly whyGone = new Map<number, string>();
  /** respawnBy: đếm nhấp nháy theo nguyên nhân, cho bài đo đọc. */
  readonly respawnBy = { ds: 0, hethan: 0, quai: 0 };
  /** despawnTold: số lần server báo `Despawn` (§4.3). */
  despawnTold = 0;
  /**
   * descCache — **Entity Store** của Protocol v2 (NETCODE_SPEC §4.3, §5 L0).
   *
   * Giữ mô tả (tên + nhân vật) của entity **sau khi đã `Despawn`**, theo `(id, desc_ver)`.
   * Quay lại trong `DESC_KEEP_SECS` thì dựng lại được NGAY, không phải chờ server gửi lại
   * tên và không rơi về placeholder.
   *
   * Vì sao đây mới là chỗ chữa thật, chứ không phải "client xoá theo Despawn":
   * lượt trước bật client xoá ngay theo `ds` làm nhấp nháy **tệ đi** — 8 lần/15 lượt so với
   * 2 lần/13 lượt — và đo ra **toàn bộ** số nhấp nháy đến từ chính `Despawn`. Server xoá
   * đúng luật (ra ngoài lề 900 quá 2 giây), nhưng người đó quay lại đủ nhanh để mắt thấy
   * một lần biến mất rồi hiện lại. Xoá đúng luật vẫn là nhấp nháy.
   *
   * Có Entity Store thì ranh giới despawn→spawn không còn là "dựng lại từ đầu": mô tả còn
   * nguyên, `ViewPool` trả lại đúng sprite cũ, và người chơi không thấy gì cả.
   */
  private readonly descCache = new Map<number, { name: string; char: string; at: number }>();
  /** missingDesc: entity chưa biết vẽ nhân vật nào — xem takeMissingDesc. */
  private readonly missingDesc = new Set<number>();
  /**
   * clock: đồng hồ vẽ theo giờ server, kéo về đích có giới hạn tốc (`net/interp.ts`).
   * Thay cho `offset` thô trước đây — xem ghi chú trong `applySnapshot`.
   */
  private readonly clock: Clock;
  /** traceId/trace: vết của một entity cho bài đo (xem `traceEntity`); 0 = tắt. */
  private traceId = 0;
  private readonly trace: number[] = [];
  /**
   * TRACE_MAX: trần mảng vết (§13.9 "vết netcode 2 s").
   *
   * 4 số mỗi khung × 144 khung/giây × 2 giây = 1 152; làm tròn lên 1 200. Trần cũ là
   * 40 000 số, tức là ~70 giây trên màn 144 Hz — và bài đo chỉ đọc 2 giây cuối, nên 68
   * giây còn lại là bộ nhớ giữ cho vui.
   */
  private static readonly TRACE_MAX = 1200;
  /** gaps: khoảng cách giữa các gói snapshot gần đây (ms) — nguồn của `interpTarget`. */
  private readonly gaps: number[] = [];
  private interpMs: number;
  private interpTarget: number;
  private pending: PendingInput[] = [];
  private seq = 0;
  private predicted: Vec = { x: 0, y: 0 };
  private correction: Vec = { x: 0, y: 0 };
  /** Vận tốc dự đoán — chỉ khác 0 khi đang trượt trên băng (spec §8.0). */
  private predictedVel: Vec = { x: 0, y: 0 };
  /** Đo "giật khi chạm vật cản" (spec §8.2a): số lần và độ lệch khi server sửa vị trí. */
  fixPerSec = 0;
  fixMax = 0;
  private fixCount = 0;
  private fixWorst = 0;
  private fixAt = 0;
  /** Số cú nhảy thẳng (dịch chuyển, hồi sinh) — tách khỏi hiệu chỉnh để không làm sai số đo. */
  teleports = 0;
  private lastSendAt = 0;
  /**
   * Vị trí dự đoán ở nhịp gửi KẾ TIẾP (giữ nguyên phím đang bấm). Mỗi khung vẽ chỉ nội suy
   * thẳng từ `predicted` tới đây — không chạy lại phép trượt va chạm với bước lẻ mỗi khung
   * (bước lẻ sát vật cản cho ra điểm không nằm trên đường đi của bước đủ → nhân vật và máy
   * quay rung qua lại; chủ dự án 05/11 "giật nhất là lúc đi qua vật cản").
   */
  private next: Vec = { x: 0, y: 0 };
  /** Khoảng cách thật giữa hai lần gửi (trung bình trượt, ms) — `setInterval` hay trễ hơn 1/tick_hz. */
  private sendGapMs = 0;
  private lastInput: PendingInput = { seq: 0, angle: 0, moving: false, burn: false };
  lastSnapAt = 0;

  constructor(private readonly cfg: GameConfig) {
    this.clock = new Clock({
      rate: cfg.net.clock_rate ?? 0.08,
      windowSecs: cfg.net.clock_window_secs ?? 2,
    });
    this.interpMs = cfg.net.interp_ms;
    this.interpTarget = cfg.net.interp_ms;
  }

  reset(w: Welcome, map: GameMap): void {
    this.myId = w.id;
    this.map = map;
    this.players.clear();
    this.monsters.clear();
    this.goneAt.clear();
    this.deletedAt.clear();
    this.items.clear();
    this.brokenObstacles.clear();
    for (const o of w.obs ?? []) {
      this.brokenObstacles.set(o.id, o.b);
    }
    this.pending = [];
    this.clock.reset();
    this.gaps.length = 0;
    this.interpMs = this.interpTarget = this.cfg.net.interp_ms;
  }

  /** Estimated current server time in seconds. */
  serverNow(): number {
    return this.clock.now(performance.now() / 1000);
  }

  /**
   * renderTime: thời điểm được VẼ — luôn trễ hơn giờ server đúng `interpMs`.
   *
   * `interpMs` TỰ ĐIỀU CHỈNH theo nhịp gói thật (Q48 bước c): mạng đều thì bám sàn
   * `net.interp_ms`, mạng giật thì nâng lên tới `interp_ms_max`. Đo 11/11 trên bài
   * `gomdon-smooth`: khoảng cách hai gói trung vị 66 ms nhưng có lúc 100 ms, và kho nội
   * suy có lúc chỉ còn **11,9 ms** dự trữ — một gói về muộn hơn thế là kho cạn, và lúc kho
   * cạn thì hình nhảy (đo được 2 732 đv/s trên trần 975).
   */
  renderTime(): number {
    return this.serverNow() - this.interpMs / 1000;
  }

  /**
   * traceEntity / takeTrace: ghi vết (thời_điểm_vẽ, x, y) của MỘT entity, cho bài đo Q48.
   *
   * Vì sao phải nằm TRONG client chứ không đo từ ngoài: bài đo chạy vòng `rAF` riêng, còn
   * vị trí nội suy chỉ đổi theo vòng vẽ của game. Đọc vị trí ở nhịp này mà đọc thời điểm vẽ
   * ở nhịp kia là ghép sai cặp — ra "tốc vẽ 2 807 đv/s" trong khi hai gói đang nội suy chỉ
   * đi 939 đv/s. Ghi cả cặp trong CÙNG một khung vẽ thì con số mới có nghĩa.
   *
   * Tắt mặc định (`traceId = 0`): không ai gọi thì chi phí là một phép so số nguyên/khung.
   */
  traceEntity(id: number): void {
    this.traceId = id;
    this.trace.length = 0;
  }

  takeTrace(): number[] {
    const out = this.trace.slice();
    this.trace.length = 0;
    return out;
  }

  /**
   * clockOffsetSecs: lệch giờ client↔server đang dùng (bài đo đọc để đếm "đồng hồ nhảy").
   *
   * Phải có getter vì `offset` thô đã bị thay bằng `Clock`; bài đo cũ đọc `world.offset`
   * và từ lúc đó nó im lặng trả về 0 — tức là cột chẩn đoán quan trọng nhất của Q48 báo
   * "đồng hồ không nhảy lần nào" bất kể thực tế.
   */
  clockOffsetSecs(): number {
    return this.clock.now(0);
  }

  /** interpNowMs: trễ nội suy đang dùng (bài đo `gomdon-smooth` đọc qua `window.kx`). */
  interpNowMs(): number {
    return this.interpMs;
  }

  /**
   * adaptInterp: nâng/hạ trễ nội suy theo khoảng cách gói THẬT.
   *
   * Đích = gói thưa nhất trong 40 gói gần nhất + một nhịp tick dự phòng. Lấy mốc cao nhất
   * chứ không trung bình: kho chỉ cần cạn MỘT lần là người xem thấy giật, còn trễ thêm vài
   * chục ms thì không ai nhận ra.
   */
  private adaptInterp(gapMs: number): void {
    if (gapMs > 0 && gapMs < 2000) {
      this.gaps.push(gapMs);
      if (this.gaps.length > 40) {
        this.gaps.shift();
      }
    }
    const base = this.cfg.net.interp_ms;
    const max = this.cfg.net.interp_ms_max ?? base;
    const worst = this.gaps.length > 0 ? Math.max(...this.gaps) : 0;
    this.interpTarget = Math.min(max, Math.max(base, worst + 1000 / this.cfg.net.tick_hz));
  }

  applySnapshot(s: Snapshot): GameEvent[] {
    // Đồng hồ vẽ CHỈNH TỐC, không nhảy thẳng (Q48 bước b, `net/interp.ts`).
    //
    // Bản cũ: `offset = max(ước tính)` rồi rỉ xuống 0,5 ms mỗi gói. Mỗi lần một gói về sớm
    // hơn thường lệ, offset nhảy thẳng tới mốc mới — thời điểm vẽ nhảy theo, và mọi người
    // trên màn hình dịch một phát. Bài đo đơn vị `npm run check:interp` đếm được 15 khung
    // nhảy cóc với đồng hồ cũ, tốc vẽ đỉnh 2 194 đv/s trên trần 975; với đồng hồ này là 0
    // khung và 1 053 đv/s (đúng bằng trần + 8% cho phép của chính nó).
    const now = performance.now() / 1000;
    const est = s.tm - now;
    this.clock.sample(est, now);
    // Lệch quá nhiều thì bắt sóng lại ngay: bò 8%/giây từ chỗ lệch vài giây mất hàng chục
    // giây, và suốt từng ấy thời gian thời điểm vẽ chạy trước cả gói mới nhất — người xem
    // thấy cả phòng đứng hình (đo 10/11: 3/5 lượt kẹt ở mức lệch 1,7–16 giây).
    if (Math.abs(est - (this.clock.now(now) - now)) > (this.cfg.net.clock_resync_secs ?? 1)) {
      this.clock.resync();
    }
    this.adaptInterp(performance.now() - this.lastSnapAt);
    this.lastSnapAt = performance.now();
    this.updatePlayers(s);
    this.updateMonsters(s);
    // Đổi chỉ số loại → tên ngay tại đây, một lần duy nhất. Phần còn lại của client vẫn
    // so `k === 'orb'` như cũ (xem ghi chú ở `ItemState`).
    for (const it of s.ia ?? []) {
      this.items.set(it.id, { id: it.id, x: it.x, y: it.y, k: itemKind(it.k), v: it.v });
    }
    for (const id of s.ir ?? []) {
      this.items.delete(id);
    }
    // `Despawn` tường minh (NETCODE_SPEC §4.3): giờ client HÀNH ĐỘNG theo — xem `applyDespawn`.
    this.applyDespawn(s.ds ?? [], s.tm);
    const events = s.ev ?? [];
    this.applyObstacleEvents(events);
    this.reconcile(s.ack);
    return events;
  }

  /**
   * applyDespawn xử lý `Despawn` tường minh từ server (NETCODE_SPEC §4.3).
   *
   * Xoá entity khỏi thế giới, nhưng **giữ lại mô tả** trong Entity Store. Đó là cả điểm
   * khác biệt: xoá mà vứt luôn mô tả thì lần quay lại phải chờ server gửi lại tên + nhân
   * vật, và trong lúc chờ người chơi thấy placeholder — tức là đổi một kiểu nhấp nháy lấy
   * một kiểu khác, đúng cái lượt đo trước đã bắt được.
   */
  private applyDespawn(ids: number[], now: number): void {
    this.despawnTold += ids.length;
    for (const id of ids) {
      const v = this.players.get(id);
      if (v) {
        if (v.char) {
          this.descCache.set(id, { name: v.name, char: v.char, at: now });
        }
        this.players.delete(id);
        this.deletedAt.set(id, now);
        this.whyGone.set(id, 'ds');
      }
      this.goneAt.delete(id);
      this.missingDesc.delete(id);
    }
    // Dọn mô tả quá hạn: giữ mãi thì bản đồ đông người biến Entity Store thành chỗ rò bộ nhớ.
    for (const [id, d] of this.descCache) {
      if (now - d.at > DESC_KEEP_SECS) {
        this.descCache.delete(id);
      }
    }
  }

  private updatePlayers(s: Snapshot): void {
    const seen = new Set<number>();
    for (const p of s.p ?? []) {
      seen.add(p.id);
      let v = this.players.get(p.id);
      if (!v) {
        // Quay lại sau `Despawn`: lấy mô tả từ Entity Store (§4.3). Server chỉ gửi tên +
        // nhân vật LẦN ĐẦU, nên không có bộ nhớ này thì mỗi lần quay lại là một lần
        // placeholder cho tới nhịp gửi lại tên.
        const cached = this.descCache.get(p.id);
        v = { id: p.id, name: p.nm ?? cached?.name ?? '?', char: p.ch ?? cached?.char ?? '', snaps: [] };
        this.players.set(p.id, v);
      }
      if (p.nm) {
        v.name = p.nm;
        v.char = p.ch ?? v.char;
      }
      // Chưa biết nhân vật nào = chưa vẽ được. Ghi lại để HỎI server thay vì vẽ placeholder
      // (NETCODE_SPEC §5 Hiển thị mục 2). Server chỉ gửi tên+nhân vật LẦN ĐẦU và nhớ là đã
      // gửi; nếu client mất mô tả đó thì không hỏi là kẹt vĩnh viễn.
      if (v.char === '') {
        this.missingDesc.add(p.id);
      } else {
        this.missingDesc.delete(p.id);
      }
      v.snaps.push({ t: s.tm, s: p });
      if (v.snaps.length > SNAP_KEEP) {
        v.snaps.shift();
      }
    }
    this.expire(this.players, seen, s.tm);
  }

  /**
   * expire xoá entity **sau khi vắng mặt liên tục `KEEP_GONE_SECS`**, không xoá ngay khi
   * vắng một gói (NETCODE_SPEC §4.3, §4.4).
   *
   * Đây là gốc của lỗi nhấp nháy. Server CỐ Ý bỏ qua người/quái ở xa trên các tick lẻ để
   * họ chỉ tốn 10 Hz thay vì 20 Hz (tối ưu mục 27) — nhưng client lại hiểu "vắng mặt một
   * gói" là "đã rời đi", xoá khỏi bảng, và `scene.ts` huỷ luôn sprite. Kết quả: **mọi
   * entity xa hơn nửa khung nhìn bị huỷ và dựng lại 10 lần mỗi giây.**
   *
   * Nó còn kéo theo lỗi placeholder: tên và nhân vật (`nm`, `ch`) chỉ gửi lần đầu client
   * thấy một người, server nhớ bằng `knownPlayers` và không hề biết client đã tự vứt view.
   * View dựng lại không có tên nhân vật → không tra được atlas → vẽ vòng tròn placeholder
   * cho tới nhịp gửi lại tên (2 giây).
   *
   * Đo được trước khi sửa: **8/8 lần biến mất rồi quay lại trong 2 giây**, 1/3 nhân vật
   * đang vẽ placeholder (`tools/browsertest/flicker.mjs`). Ngân sách NETCODE §1: cả hai = 0.
   */
  private expire<T>(m: Map<number, T>, seen: Set<number>, now: number): void {
    for (const id of m.keys()) {
      if (seen.has(id)) {
        // Quay lại sau khi đã bị xoá hẳn = một lần nhấp nháy người chơi NHÌN THẤY.
        // Đếm ở đây để telemetry báo được con số đó từ máy thật (TELEMETRY_SPEC §2.2
        // nhóm "Đồng bộ", ngân sách §4.3 là 0) — trước giờ chỉ đo được bằng flicker.mjs
        // trên máy dev, nơi lỗi này gần như không bao giờ xảy ra.
        const left = this.deletedAt.get(id);
        if (left !== undefined) {
          if (now - left < KEEP_GONE_SECS) {
            this.respawn2s++;
            const why = this.whyGone.get(id) ?? 'hethan';
            if (why === 'ds') { this.respawnBy.ds++; } else { this.respawnBy.hethan++; }
            if (m === (this.monsters as unknown)) { this.respawnBy.quai++; }
          }
          this.deletedAt.delete(id);
        }
        this.whyGone.delete(id);
        this.goneAt.delete(id);
        continue;
      }
      const since = this.goneAt.get(id);
      if (since === undefined) {
        this.goneAt.set(id, now);
      } else if (now - since >= KEEP_GONE_SECS) {
        m.delete(id);
        this.goneAt.delete(id);
        this.deletedAt.set(id, now);
        this.whyGone.set(id, 'hethan');
      }
    }
    // Dọn `deletedAt`: entity rời đi quá `KEEP_GONE_SECS` mà chưa quay lại thì có quay lại
    // cũng không còn tính là nhấp nháy nữa. Không dọn thì map này phình mãi — một trận dài
    // đi qua hàng nghìn entity, và đó là rò rỉ bộ nhớ thật chứ không phải lý thuyết.
    for (const [id, at] of this.deletedAt) {
      if (now - at >= KEEP_GONE_SECS) {
        this.deletedAt.delete(id);
        // Dọn cả `whyGone` theo: entity đi hẳn không bao giờ quay lại thì mục của nó nằm
        // đó mãi, và một trận dài đi qua hàng nghìn entity.
        this.whyGone.delete(id);
      }
    }
  }

  /**
   * takeMissingDesc trả về danh sách entity chưa có mô tả, để gửi `DescRequest`.
   *
   * KHÔNG xoá khỏi tập: chỉ khi mô tả thật sự về tới (nhánh trên) mới xoá. Xoá ngay lúc hỏi
   * thì gói `dreq` rơi mất là không bao giờ hỏi lại nữa — đúng cái bẫy mà tính năng này sinh
   * ra để tránh. Người gọi tự giới hạn nhịp hỏi.
   */
  takeMissingDesc(): number[] {
    return [...this.missingDesc];
  }

  /**
   * takeRespawn2s trả về số lần nhấp nháy kể từ lần gọi trước, rồi đặt lại bộ đếm.
   *
   * "Lấy rồi xoá" chứ không phải "đọc": telemetry gộp theo cửa sổ 10 giây, nên mỗi lần
   * nhấp nháy phải được đếm đúng MỘT lần vào đúng một cửa sổ.
   */
  takeRespawn2s(): number {
    const n = this.respawn2s;
    this.respawn2s = 0;
    return n;
  }

  /**
   * takeRespawnBy trả bảng chia theo NGUYÊN NHÂN rồi đặt lại — cùng luật "lấy rồi xoá".
   *
   * Vì sao telemetry cần nó, chứ không chỉ cần tổng: ngày 2/11 đi tìm lại con số "1.357
   * lần respawn" của 31/10, dữ liệu người thật chỉ có TỔNG. Không trả lời được câu duy
   * nhất đáng hỏi — do `Despawn` của server, do luật hết hạn 2 giây ở client, hay do quái
   * — nên phải dựng lại cả chuỗi build mới lần ra. Ba nguyên nhân cần ba cách chữa khác
   * hẳn nhau.
   *
   * Lúc đó bảng này đã tồn tại nhưng chỉ `tools/browsertest/respawn-wander.mjs` đọc, tức
   * chỉ thấy được trên máy dev — đúng chỗ lỗi gần như không bao giờ xảy ra.
   */
  takeRespawnBy(): { ds: number; hethan: number; quai: number } {
    const out = { ...this.respawnBy };
    this.respawnBy.ds = 0;
    this.respawnBy.hethan = 0;
    this.respawnBy.quai = 0;
    return out;
  }

  private updateMonsters(s: Snapshot): void {
    const seen = new Set<number>();
    for (const m of s.m ?? []) {
      seen.add(m.id);
      const buf = this.monsters.get(m.id) ?? [];
      // Loại, máu tối đa và bán kính chỉ được gửi LẦN ĐẦU (spec mục 27) — chúng không bao
      // giờ đổi. Các gói sau thiếu ba trường đó, phải lấy lại từ gói trước.
      const prev = buf[buf.length - 1]?.s;
      // `hp` cũng chỉ gửi khi ĐỔI (delta snapshot, spec mục 27) — quái đứng yên gặm cỏ thì
      // gói nào cũng thiếu trường đó, phải giữ lại máu cũ chứ không được coi là 0.
      const full: MonsterState = prev
        ? { ...m, k: m.k || prev.k, mx: m.mx || prev.mx, r: m.r || prev.r, hp: m.hp ?? prev.hp }
        : m;
      buf.push({ t: s.tm, s: full });
      if (buf.length > SNAP_KEEP) {
        buf.shift();
      }
      this.monsters.set(m.id, buf);
    }
    this.expire(this.monsters, seen, s.tm);
  }

  private applyObstacleEvents(events: GameEvent[]): void {
    for (const e of events) {
      const id = e.a ?? -1;
      if (e.k === 'ObstacleBroken') {
        this.brokenObstacles.set(id, true);
      } else if (e.k === 'ObstacleRegrow' && e.s !== 'warn') {
        this.brokenObstacles.set(id, false);
      } else if (e.k === 'StallUp') {
        // `f` là bề rộng một cái sạp; `x`/`y` là TÂM sạp, nên lùi nửa bề rộng ra góc.
        const w = e.f ?? 0;
        this.stalls.set(e.b ?? -1, {
          owner: e.a ?? 0,
          tier: e.n ?? 0,
          shape: { x: (e.x ?? 0) - w / 2, y: (e.y ?? 0) - w / 2, w, h: w },
        });
      } else if (e.k === 'StallDown') {
        this.stalls.delete(e.b ?? -1);
      }
    }
  }

  /**
   * blockingStalls: những cái sạp của Cô Ba đang chặn `me` — bản port của `Game.stallBlocking`.
   *
   * Ba trường hợp đi xuyên, phải khớp y nguyên server, nếu không client dự đoán bị chặn
   * trong khi server cho qua (hoặc ngược lại) và nhân vật giật ở đúng miệng dãy sạp:
   *   1. chủ sạp (`bay_sap_owner_passes`);
   *   2. người chém được bậc đó;
   *   3. Shipper đang Gom Đơn (`ghost`) — xuyên hết, không cần xét gì.
   */
  private blockingStalls(me: PlayerState): Shape[] {
    if (this.stalls.size === 0 || statusOn(me.st, 'ghost')) {
      return [];
    }
    const ownerPasses = (this.cfg.skills.ranges_u?.bay_sap_owner_passes ?? 0) !== 0;
    const out: Shape[] = [];
    for (const st of this.stalls.values()) {
      if (ownerPasses && st.owner === this.myId) {
        continue;
      }
      if (this.canCutObstacle(st.tier)) {
        continue;
      }
      out.push(st.shape);
    }
    return out;
  }

  /**
   * canCutObstacle: cảnh giới của mình có chém được vật cản bậc `tier` hay không (spec §8.2).
   *
   * Phải khớp **từng chữ** với `Game.CanCutObstacle` bên server, vì client dùng nó để quyết
   * định vẽ viền tối mờ. Lệch một bậc là người chơi thấy viền ở vật cản họ chém được (hoặc
   * tệ hơn: không thấy viền ở vật cản họ không chém được rồi mài kiếm 5 giây vô ích).
   */
  canCutObstacle(tier: number): boolean {
    const me = this.me();
    if (!me || tier >= TIER_BORDER) {
      return false;
    }
    let reach = realmOf(this.cfg, me.L);
    if (me.bf & (1 << BUFFS.indexOf('hard'))) {
      reach += this.cfg.obstacles.hard_buff_extra_tier;
    }
    return reach >= tier;
  }

  /** myName: tên của mình như server đã duyệt (có thể khác tên gõ vào nếu bị lọc). */
  myName(): string {
    return this.players.get(this.myId)?.name ?? '';
  }

  me(): PlayerState | null {
    const v = this.players.get(this.myId);
    return v ? v.snaps[v.snaps.length - 1].s : null;
  }

  // ---- input & prediction ------------------------------------------------

  /** Records and returns the input message to send this network tick. */
  makeInput(angle: number, moving: boolean, burn = false, skill = false): Record<string, unknown> {
    const inp = { seq: ++this.seq, angle: quantize(angle), moving, burn };
    this.pending.push(inp);
    const me = this.me();
    const nowMs = performance.now();
    // Điểm đang vẽ ngay trước nhịp này: gửi sớm/muộn vài ms hoặc vừa đổi hướng thì điểm mới
    // lệch chút ít — dồn phần lệch vào `correction` (tan dần) thay vì để hình nhảy một khung.
    const before = this.shown(nowMs);
    if (me && !(me.fl & 16)) {
      const r = this.step(this.predicted, this.predictedVel, inp, me, 1 / this.cfg.net.tick_hz);
      this.predicted = r.pos;
      this.predictedVel = r.vel;
    }
    this.lastInput = inp;
    this.noteSendGap(nowMs - this.lastSendAt);
    this.lastSendAt = nowMs;
    this.advance(me);
    const after = this.shown(nowMs);
    if (Math.hypot(before.x - after.x, before.y - after.y) < this.cfg.net.correction_snap) {
      this.correction.x += before.x - after.x;
      this.correction.y += before.y - after.y;
    }
    const msg: Record<string, unknown> = { t: 'in', s: inp.seq, a: Math.round((inp.angle / (2 * Math.PI)) * 256) & 255, m: moving, b: burn };
    if (skill) {
      // Chỉ gửi khi thật sự bấm: đây là sự kiện một lần, không phải trạng thái giữ phím.
      msg.sk = true;
    }
    return msg;
  }

  private reconcile(ack: number): void {
    const me = this.me();
    if (!me) {
      return;
    }
    this.pending = this.pending.filter((p) => p.seq > ack);
    let pos: Vec = { x: me.x, y: me.y };
    // Phát lại từ đúng VẬN TỐC server đang giữ, không phải từ 0: trên mặt băng vận tốc là
    // trạng thái, bỏ nó đi là mỗi snapshot lại dự đoán sai và nhân vật giật cả mặt hồ.
    let vel: Vec = { x: me.vx ?? 0, y: me.vy ?? 0 };
    for (const inp of this.pending) {
      const r = this.step(pos, vel, inp, me, 1 / this.cfg.net.tick_hz);
      pos = r.pos;
      vel = r.vel;
    }
    // Đếm số lần server kéo nhân vật về chỗ khác (spec §8.2a): đây là thước đo trực tiếp
    // của cảm giác "giật khi đi sát vật cản". Dự đoán đúng thì con số này ~0; lệch thì nó
    // chạy lên bằng đúng tần số snapshot (20/giây) và người chơi thấy rung liên tục.
    const off = Math.hypot(this.predicted.x - pos.x, this.predicted.y - pos.y);
    // Bỏ qua cú nhảy lớn: đó là dịch chuyển qua cổng hoặc hồi sinh, không phải lỗi dự đoán.
    // Không lọc thì một lần qua cổng làm "lệch nhất" nhảy lên 12 000 và che hết số thật.
    if (off > CORRECTION_EPS && off <= this.cfg.net.correction_snap) {
      this.fixCount++;
      this.fixWorst = Math.max(this.fixWorst, off);
    }
    const nowMs = performance.now();
    if (nowMs - this.fixAt >= 1000) {
      this.fixPerSec = (this.fixCount * 1000) / (nowMs - this.fixAt);
      this.fixMax = this.fixWorst;
      this.fixCount = 0;
      this.fixWorst = 0;
      this.fixAt = nowMs;
    }
    // Keep the displayed position continuous; the correction fades out in renderFrame.
    // So ĐÚNG điểm đang vẽ trước/sau (đang nằm giữa predicted và next), không chỉ predicted.
    const before = this.shown(nowMs);
    this.predicted = pos;
    this.predictedVel = vel;
    this.advance(me);
    const after = this.shown(nowMs);
    this.correction.x += before.x - after.x;
    this.correction.y += before.y - after.y;
    // Spec §8.2a: lệch nhỏ thì kéo dần về trong `correction_tau_ms`; lệch lớn (dịch chuyển
    // qua cổng, hồi sinh) thì nhảy thẳng — kéo dần một cú nhảy nửa bản đồ chỉ làm nhân vật
    // trôi dài qua màn hình. Ngưỡng cũ 400 đơn vị quá rộng: một cú lệch 200 đơn vị vẫn bị
    // kéo dần và người chơi thấy nhân vật "bơi" về chỗ cũ.
    if (Math.hypot(this.correction.x, this.correction.y) > this.cfg.net.correction_snap) {
      this.correction = { x: 0, y: 0 };
      this.teleports++;
    }
  }

  /** advance tính `next`: một bước đủ từ `predicted` với phím đang giữ. */
  private advance(me: PlayerState | null): void {
    this.next = me && !(me.fl & 16)
      ? this.step(this.predicted, this.predictedVel, this.lastInput, me, 1 / this.cfg.net.tick_hz).pos
      : { ...this.predicted };
  }

  private noteSendGap(gap: number): void {
    const ideal = 1000 / this.cfg.net.tick_hz;
    if (this.sendGapMs === 0) {
      this.sendGapMs = ideal;
    }
    // Bỏ khoảng hở dài (tab vừa quay lại, máy khựng): đó không phải nhịp gửi bình thường.
    if (gap > 0 && gap < ideal * 4) {
      this.sendGapMs += (gap - this.sendGapMs) * 0.1;
    }
  }

  /**
   * shown: điểm vẽ nhân vật của mình lúc `nowMs` — nội suy thẳng predicted → next theo phần
   * nhịp gửi đã trôi. Chia cho nhịp gửi THẬT (không phải 1/tick_hz): `setInterval` trễ đều
   * vài ms thì chia cho số lý thuyết sẽ chạm trần 1 sớm và nhân vật đứng hình một khung.
   */
  private shown(nowMs: number): Vec {
    const gap = this.sendGapMs || 1000 / this.cfg.net.tick_hz;
    const a = Math.min(Math.max((nowMs - this.lastSendAt) / gap, 0), 1);
    return { x: lerp(this.predicted.x, this.next.x, a), y: lerp(this.predicted.y, this.next.y, a) };
  }

  /** One movement step, mirroring game.movePlayer on the server. */
  private step(from: Vec, vel: Vec, inp: PendingInput, me: PlayerState, dt: number): { pos: Vec; vel: Vec } {
    const map = this.map;
    if (!map) {
      return { pos: from, vel };
    }
    // fl & 8 = đang dịch chuyển qua cổng: không điều khiển được, cũng không trôi.
    if (me.fl & 8) {
      return { pos: from, vel: { x: 0, y: 0 } };
    }
    // Trói / rụt mai / choáng: server đặt tốc độ 0, nên client cũng phải đứng yên — dự đoán
    // là mình đang chạy trong lúc server giữ tại chỗ thì mỗi tick bị kéo về một lần.
    // Phải khớp DANH SÁCH của `Game.MoveSpeed` bên server: trói, rụt mai, choáng, ngủ gục,
    // hát ru. Thiếu một cái là client dự đoán mình đang chạy trong lúc server giữ tại chỗ,
    // và mỗi tick bị kéo về một lần — đúng cái giật §8.2a.
    const frozen = statusOn(me.st, 'root') || statusOn(me.st, 'shell') || statusOn(me.st, 'stun')
      || statusOn(me.st, 'sleep') || statusOn(me.st, 'sing');
    // Tốc CỘNG trên base rồi kẹp (spec v6.25 §4.1) — cùng danh sách nguồn với
    // `Game.speedBonus`; vùng địa hình do `moveStep` cộng thêm.
    const sp = this.cfg.speed;
    // Nội tại nhân vật là một khoản cộng như mọi buff (spec v6.33 §6.4), khớp `Game.speedBonus`.
    let bonus = innate(this.cfg, this.players.get(this.myId)?.char ?? '').speed;
    let cap = sp.cap;
    if (me.bf & (1 << BUFFS.indexOf('haste'))) {
      bonus += sp.haste;
    }
    // Thiêu Đốt Tinh Huyết (spec §4.1c). Server dừng cộng tốc khi EXP chạm đáy cảnh giới
    // và sẽ sửa lại vị trí; ở đây chỉ cần dự đoán đúng trường hợp thường gặp.
    if (inp.burn && !frozen) {
      bonus += sp.burn;
    }
    if (statusOn(me.st, 'slow')) {
      bonus += sp.slow_status;
    }
    if (statusOn(me.st, 'ghost')) {
      bonus += sp.gom_don;
      cap = sp.cap_gom_don;
    }
    // Đảo điều khiển (Phun Trân Châu, v6.48 §6.3): server quay hướng đi 180° ngay ở cửa
    // vào `SetInput`, nên client phải đảo Y HỆT. Không đảo thì suốt 3 giây client vẽ mình
    // đi một đằng còn server kéo về một nẻo — 20 lần mỗi giây.
    let angle = statusOn(me.st, 'invert') ? inp.angle + Math.PI : inp.angle;
    let moving = inp.moving;
    // Khiêu khích của Cụ Rùa (v6.50): không tự chọn hướng, bị kéo bước về phía cụ. Điểm neo
    // là chỗ cụ đứng lúc giáng trạng thái — cụ ở trong mai nên đứng yên đúng 2 giây đó.
    const taunt = statusOn(me.st, 'taunt') ? statusTracker.get(this.myId, 'taunt') : undefined;
    if (taunt) {
      const dx = taunt.x - from.x;
      const dy = taunt.y - from.y;
      if (Math.hypot(dx, dy) > this.cfg.player.body_r * me.sc * 2) {
        angle = Math.atan2(dy, dx);
        moving = true;
      } else {
        moving = false;
      }
    }
    return moveStep(this.cfg, map, from, angle, me.sc, dt,
      (id: number) => this.brokenObstacles.get(id) === true, { bonus, cap, frozen }, vel, moving,
      statusOn(me.st, 'ghost'), this.blockingStalls(me));
  }




  // ---- per-frame queries -------------------------------------------------

  /** Interpolated (others) or predicted (me) state for rendering. */
  playerFrames(dt: number): PlayerFrame[] {
    // Nhích đồng hồ MỘT LẦN mỗi khung vẽ, ngay trước khi dùng nó: `playerFrames` là cửa
    // đầu tiên của vòng vẽ, nên mọi thứ vẽ sau đó đọc cùng một thời điểm.
    this.clock.step(dt, performance.now() / 1000);
    // Kéo trễ nội suy về đích CHẬM: đổi nhanh thì chính việc đổi làm thời điểm vẽ nhảy, tức
    // là tự gây ra đúng cái giật đang đi chữa.
    const stepMs = (this.cfg.net.interp_adapt_ms_per_sec ?? 20) * dt;
    const d = this.interpTarget - this.interpMs;
    this.interpMs += Math.max(-stepMs, Math.min(stepMs, d));
    const decay = Math.exp((-dt * 1000) / Math.max(this.cfg.net.correction_tau_ms, 1));
    this.correction.x *= decay;
    this.correction.y *= decay;
    const rt = this.renderTime();
    const out: PlayerFrame[] = [];
    for (const v of this.players.values()) {
      out.push(v.id === this.myId ? this.myFrame(v) : interpPlayer(v, this.cfg, rt));
    }
    if (this.traceId !== 0 && this.trace.length < ClientWorld.TRACE_MAX) {
      const f = out.find((p) => p.view.id === this.traceId);
      if (f) {
        // Kèm TỐC ĐỘ CỦA DỮ LIỆU ở đúng khung này: cặp gói đang được nội suy đi nhanh bao
        // nhiêu. Có nó mới tách được ba thứ khác hẳn nhau mà nhìn ra đều là "giật":
        // dịch chuyển qua cổng (dữ liệu nhảy — đúng), entity dừng thật (dữ liệu đứng —
        // đúng), và lỗi nội suy (dữ liệu đều mà hình nhảy — sai).
        const { a, b } = bracket(f.view.snaps, rt);
        const span = b.t - a.t;
        const dataV = span > 0 ? Math.hypot(b.s.x - a.s.x, b.s.y - a.s.y) / span : 0;
        this.trace.push(rt, f.x, f.y, dataV);
      }
    }
    return out;
  }

  private myFrame(v: PlayerView): PlayerFrame {
    const last = v.snaps[v.snaps.length - 1];
    const s = last.s;
    const ahead = this.shown(performance.now());
    const ringAngle = s.ph + s.d * omega(this.cfg, s, v.char) * (this.serverNow() - last.t);
    return { view: v, s, x: ahead.x + this.correction.x, y: ahead.y + this.correction.y, ringAngle };
  }

  monsterFrames(): { s: MonsterState; x: number; y: number }[] {
    const rt = this.renderTime();
    const out: { s: MonsterState; x: number; y: number }[] = [];
    for (const [id, buf] of this.monsters) {
      let tr = this.mtracks.get(id);
      if (!tr) {
        tr = {};
        this.mtracks.set(id, tr);
      }
      const at = followSnaps(buf, rt, tr, { jumpUnits: 600 });
      out.push({ s: buf[buf.length - 1].s, x: at.x, y: at.y });
    }
    for (const id of this.mtracks.keys()) {
      if (!this.monsters.has(id)) {
        this.mtracks.delete(id);
      }
    }
    return out;
  }
}

function quantize(angle: number): number {
  const step = (2 * Math.PI) / 256;
  const a = ((angle % (2 * Math.PI)) + 2 * Math.PI) % (2 * Math.PI);
  return Math.round(a / step) * step;
}

/**
 * omega: tốc độ quay vòng kiếm, rad/giây. Server KHÔNG gửi con số này nữa (spec mục 27) —
 * nó suy ra được từ `rev_secs` trong config và buff Tốc Kiếm mà client đã có sẵn, nên gửi
 * lại là lãng phí 7 byte mỗi người mỗi gói, 20 lần mỗi giây.
 */
export function omega(cfg: GameConfig, s: PlayerState, char: string): number {
  // v6.27: khống chế (kể cả ngủ) KHÔNG dừng vòng kiếm — phải khớp `Game.Omega` bên server,
  // lệch là kiếm trên màn hình quét trúng mà server không tính (cùng lớp lỗi §8.2a).
  // Luật cộng §4.0: nội tại + Tốc Kiếm, trần ×2.
  const spin = s.bf & (1 << BUFFS.indexOf('spin')) ? bonus(cfg.buffs.spin?.mult) : 0;
  return stat((2 * Math.PI) / cfg.sword.rev_secs, { cap: cfg.stats?.spin_cap }, innate(cfg, char).spin, spin);
}

function interpPlayer(v: PlayerView, cfg: GameConfig, rt: number): PlayerFrame {
  // `bracket` ở đây chỉ để lấy gói nền cho GÓC VÒNG KIẾM; vị trí do `followSnaps` lo.
  const { a, b, k } = bracket(v.snaps, rt);
  const base = k >= 1 ? b : a;
  const ringAngle = base.s.ph + base.s.d * omega(cfg, base.s, v.char) * (rt - base.t);
  // Vị trí đi qua `net/interp.ts` để bài đơn vị `check:interp` phủ đúng đường đang chạy.
  v.track ??= {};
  const at = followSnaps(v.snaps, rt, v.track, { jumpUnits: 600 });
  return { view: v, s: b.s, x: at.x, y: at.y, ringAngle };
}
