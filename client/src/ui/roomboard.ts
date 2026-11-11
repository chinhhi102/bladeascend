import type { GameConfig } from '../config';
import type { RoomBoardMsg, RoomBoardRow } from '../net/protocol';
import { assets } from '../render/assets';
import { quality } from '../render/quality';
import { h } from './dom';
import { narrowScreen } from './layout';
import { ClipCell, crown, FaceCell, fmtScore, medal, PillCell, titleText, ZoneCell } from './roomboard-parts';
import { STR } from './strings';

/**
 * Bảng xếp hạng phòng góc trên phải — GAME_SPEC §11.5 (v6.47).
 *
 * Top 1 là thẻ to viền vàng, Top 2/3 là hai thẻ nửa rộng bạc/đồng, hạng 4–10 là dòng gọn,
 * dòng của mình ghim cuối. Điện thoại: mặc định thu gọn thành một hàng (3 chân dung +
 * "Bạn #N"), chạm để mở, chạm ngoài để đóng.
 *
 * **KÍCH THƯỚC CỐ ĐỊNH (v6.47 — sửa lỗi 07/11 "bảng tụt còn 3 dòng rồi bung lại 10").**
 * Mọi ô được dựng MỘT LẦN trong `constructor`: 3 thẻ Top + 7 ô dòng + 1 ô của mình. Cập
 * nhật chỉ sửa chữ/màu bên trong ô; không `replaceChildren`, không `remove()`, không
 * `display:none` một dòng nào. Phòng ít hơn 10 người → ô thừa mang class `empty`, chữ là
 * khoảng trắng cứng nên vẫn chiếm đúng chiều cao.
 *
 * Ba thứ đo được ở bản cũ, cả ba đều từ việc DỰNG LẠI DOM theo dữ liệu (đo 07/11, 45 s,
 * phòng 105 người): số phần tử đổi 15 lần; `trimToScreen()` ẩn dòng cho vừa `max-height`
 * nên chỉ còn 5 dòng và số dòng phụ thuộc độ dài tên; `setList` xoá/chèn phần tử mỗi khi
 * top 10 đổi người. Giờ bảng vừa màn hình bằng cách THU NHỎ cả bảng (`fit`), không bằng
 * cách bớt dòng.
 *
 * Hiệu năng (§11.5): dữ liệu tới 1 lần/giây, mỗi ô có CHỮ KÝ — chữ ký không đổi thì không
 * đụng vào DOM. Cả gói được áp trong MỘT `requestAnimationFrame`, không vẽ nửa chừng.
 */

/** Số ô dòng hạng 4–10 — luôn có mặt, kể cả phòng 4 người (§11.5 v6.47). */
const LIST_SLOTS = 7;
/** Tổng số người hiện trên bảng: 3 thẻ Top + 7 dòng. */
const TOP_N = 3 + LIST_SLOTS;
/** Thời gian trượt dòng khi đổi hạng (§11.5: 0.25 s). */
const SLIDE_MS = 250;
/** Thẻ Top 1 rung trước khi đổi người (§11.5). */
const SHAKE_MS = 320;
/** Nhấn giữ bao lâu thì hiện đủ tên/danh hiệu bị cắt "…" (§11.5). */
const HOLD_MS = 450;
/** Điện thoại: bảng đang mở mà không ai chạm vào ngần này thì tự đóng (chủ dự án chốt 1/11). */
const IDLE_CLOSE_MS = 8000;
/** Bảng không bao giờ cao quá ngần này màn hình (chủ dự án chốt 31/10). */
const MAX_SCREEN_FRAC = 0.5;
/** Thu nhỏ tới đâu là hết (v6.49): dưới mức này chữ bé quá, thà bớt dòng. */
const MIN_SCALE = 0.85;
/**
 * Số ô dòng theo CHIỀU CAO MÀN HÌNH (Q49 / §11.5 v6.49) — tính 1 lần khi vào trận và khi
 * đổi cỡ cửa sổ, KHÔNG theo dữ liệu.
 *
 * Hai luật "luôn đủ 10 dòng" và "không quá nửa màn hình" không cùng đúng được ở màn thấp:
 * bảng 10 dòng cao ~539 px, nửa màn 1366×768 chỉ có 384 px. Chủ dự án chốt (Q49, phương án
 * c): màn thấp thì hiện ít dòng hơn — nhưng ít dòng ấy cũng là một con số CỐ ĐỊNH, không
 * phụ thuộc có bao nhiêu người trong phòng hay tên họ dài bao nhiêu.
 */
const HEIGHT_TIERS: { minH: number; rows: number; cls: string }[] = [
  { minH: 900, rows: 7, cls: '' },      // đủ top 10
  { minH: 700, rows: 4, cls: 'h-mid' }, // top 7
  { minH: 0, rows: 2, cls: 'h-low' },   // top 5
];

/** heightTier: tầng chiều cao đang dùng cho cửa sổ hiện tại. */
function heightTier(): { minH: number; rows: number; cls: string } {
  const h = window.innerHeight;
  return HEIGHT_TIERS.find((t) => h >= t.minH) ?? HEIGHT_TIERS[HEIGHT_TIERS.length - 1];
}

/**
 * rbStats: số liệu cho bài nghiệm thu §17 (`tools/browsertest/rbstable.mjs`) — bảng có bị
 * co giãn không, và nếu có thì lỗi ở server gửi thiếu hay ở client dựng lại DOM.
 *
 * Không khoá sau quyền debug, cùng lý do như `__stats` (main.ts): đây là số đếm, không phải
 * cửa vào trạng thái trò chơi, và bài đo phải chạy được trên bản build thật.
 */
export const rbStats = { packets: 0, applied: 0, short: 0, rows: 0, total: 0 };

/** Phím điều khiển: bấm là đóng bảng đầy đủ đang mở (§11.5 v6.49). */
const MOVE_KEYS = new Set(['KeyW', 'KeyA', 'KeyS', 'KeyD',
  'ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight', 'Space']);

/** SelfInfo: cái gói bảng không chở về chính mình nhưng dòng ghim cần (§11.5). */
interface SelfInfo {
  name: string;
  realm: number;
  tier: number;
}

interface Ctx {
  myId: number;
  myZone: number;
  self?: SelfInfo;
}

/** rowSig: mọi thứ hiển thị của một người — đổi bất cứ gì thì ô đó mới vẽ lại. */
function rowSig(row: RoomBoardRow | null, ctx: Ctx): string {
  if (!row) {
    return '-';
  }
  return [row.id, row.name, row.character, row.titleId, row.rank, row.zoneId, row.realm, row.tier,
    row.score, row.id === ctx.myId, row.zoneId === ctx.myZone].join('|');
}

/** arrow: ▲ xanh khi lên hạng, ▼ đỏ khi xuống, mờ dần 2 s (§11.5) — phần tử CỐ ĐỊNH. */
class ArrowCell {
  readonly el = h('span', 'rb-arrow');
  private at = 0;

  set(from: number | undefined, to: number): void {
    if (from === undefined || from === to) {
      return;
    }
    const up = to < from;
    this.el.textContent = up ? '▲' : '▼';
    this.el.classList.toggle('up', up);
    this.el.classList.toggle('down', !up);
    // Chạy lại hoạt ảnh mờ dần dù class vừa mới chạy xong.
    this.el.classList.remove('show');
    void this.el.offsetWidth;
    this.el.classList.add('show');
    this.at = performance.now();
  }

  tick(now: number): void {
    if (this.at > 0 && now - this.at > 2000) {
      this.at = 0;
      this.el.classList.remove('show');
      this.el.textContent = '';
    }
  }
}

/** Top1Slot: thẻ hạng 1 — vương miện, chân dung 40 px, vệt sáng, đốm lấp lánh. */
class Top1Slot {
  readonly el = h('div', 'rb-top1');
  readonly arrow = new ArrowCell();
  private readonly face: FaceCell;
  private readonly pill: PillCell;
  private readonly zone = new ZoneCell();
  private readonly name = new ClipCell('rb-name');
  private readonly title = new ClipCell('rb-title');
  private readonly score = h('span', 'rb-score');
  private sig = '';

  constructor(cfg: GameConfig) {
    this.face = new FaceCell(cfg, 40);
    this.pill = new PillCell(cfg, false);
    const shine = h('i', 'rb-shine');
    const streak = assets.url('ui/shine_streak.png');
    if (streak) {
      shine.style.backgroundImage = `url(${streak})`;
    }
    this.el.append(
      shine, ...[1, 2, 3, 4].map((i) => h('i', `rb-spark s${i}`)),
      h('div', 'rb-portrait', crown(), this.face.el),
      h('div', 'rb-info',
        h('div', 'rb-line', this.name.el, this.score, this.arrow.el),
        this.title.el,
        h('div', 'rb-tags', this.pill.el, this.zone.el),
      ),
    );
  }

  set(row: RoomBoardRow | null, ctx: Ctx): void {
    const sig = rowSig(row, ctx);
    if (sig === this.sig) {
      return;
    }
    this.sig = sig;
    this.el.classList.toggle('empty', !row);
    this.el.classList.toggle('me', !!row && row.id === ctx.myId);
    this.face.set(row);
    this.pill.set(row);
    this.zone.set(row, ctx.myZone);
    this.name.set(row?.name ?? '');
    this.title.set(row ? titleText(row) : '');
    this.score.textContent = row ? fmtScore(row.score) : '';
  }
}

/** PodSlot: thẻ hạng 2 hoặc 3 — huy chương, chân dung 28 px, tên trên dòng riêng. */
class PodSlot {
  readonly el: HTMLElement;
  readonly arrow = new ArrowCell();
  private readonly face: FaceCell;
  private readonly pill: PillCell;
  private readonly zone = new ZoneCell();
  private readonly name = new ClipCell('rb-name');
  private readonly title = new ClipCell('rb-title');
  private readonly score = h('span', 'rb-score');
  private sig = '';

  constructor(cfg: GameConfig, rank: 2 | 3) {
    this.el = h('div', `rb-pod p${rank}`);
    this.face = new FaceCell(cfg, 28);
    this.pill = new PillCell(cfg, true);
    // Tên lên dòng RIÊNG, đủ cả bề rộng thẻ (duyệt 1/11: chung dòng với chân dung thì tên
    // chỉ còn 4 ký tự). Dòng trên: huy chương + chân dung bên trái, Uy Danh góc phải.
    this.el.append(
      h('i', 'rb-shine'),
      h('div', 'rb-line', medal(rank), this.face.el, this.score, this.arrow.el),
      this.name.el, this.title.el,
      h('div', 'rb-tags', this.pill.el, this.zone.el),
    );
  }

  set(row: RoomBoardRow | null, ctx: Ctx): void {
    const sig = rowSig(row, ctx);
    if (sig === this.sig) {
      return;
    }
    this.sig = sig;
    this.el.classList.toggle('empty', !row);
    this.el.classList.toggle('me', !!row && row.id === ctx.myId);
    this.face.set(row);
    this.pill.set(row);
    this.zone.set(row, ctx.myZone);
    this.name.set(row?.name ?? '');
    this.title.set(row ? titleText(row) : '');
    this.score.textContent = row ? fmtScore(row.score) : '';
  }
}

/** RowSlot: một dòng gọn hạng 4–10, và cả dòng ghim của mình. */
class RowSlot {
  readonly el: HTMLElement;
  readonly arrow = new ArrowCell();
  private readonly pill: PillCell;
  private readonly zone = new ZoneCell();
  private readonly name = new ClipCell('rb-name');
  private readonly sub = new ClipCell('rb-sub');
  private readonly rank = h('span', 'rb-rank');
  private readonly score = h('span', 'rb-score');
  private sig = '';

  constructor(cfg: GameConfig, private readonly mine = false) {
    this.el = h('div', `rb-row${mine ? ' rb-me' : ''}`);
    this.pill = new PillCell(cfg, true);
    this.el.append(
      h('div', 'rb-line', this.rank, this.name.el, this.pill.el, this.zone.el, this.score, this.arrow.el),
      this.sub.el,
    );
  }

  set(row: RoomBoardRow | null, ctx: Ctx): void {
    const sig = rowSig(row, ctx);
    if (sig === this.sig) {
      return;
    }
    this.sig = sig;
    this.el.classList.toggle('empty', !row);
    this.el.classList.toggle('me', !this.mine && !!row && row.id === ctx.myId);
    this.rank.textContent = row ? (this.mine ? `▶ ${row.rank}` : String(row.rank)) : '';
    this.name.set(row?.name ?? '');
    // Dòng của mình: nền nổi bật đã nói "đây là bạn", không cần danh hiệu dòng phụ.
    this.sub.set(row && !this.mine ? titleText(row) : '');
    this.pill.set(row);
    this.zone.set(row, this.mine ? -1 : ctx.myZone);
    this.score.textContent = row ? fmtScore(row.score) : '';
  }
}

export class RoomBoard {
  readonly el = h('div', 'room-board rb');
  private readonly head = h('div', 'rb-head');
  private readonly count = h('span', 'rb-count');
  private readonly top1: Top1Slot;
  private readonly pods: PodSlot[];
  private readonly rows: RowSlot[];
  private readonly meRow: RowSlot;
  private readonly mini = h('div', 'rb-mini');
  private readonly miniFaces: FaceCell[];
  private readonly miniRank = h('span', 'rb-mrank');
  private prevRank = new Map<number, number>();
  /** slotOf: ô mà mỗi người đứng lần trước — để dịch ô bằng `transform` khi đổi hạng. */
  private slotOf = new Map<number, number>();
  private top1Id = 0;
  private open = false;
  private holdTimer = 0;
  private idleTimer = 0;
  private raf = 0;
  private pending: { msg: RoomBoardMsg; ctx: Ctx } | null = null;
  private natural = 0;
  private tier = HEIGHT_TIERS[0];
  private readonly closeBtn = h('button', 'rb-close', '✕');

  /** onOpenChange: báo HUD để ẩn nút kỹ năng/🔥, hoãn thông báo khi bảng mở (§11.5, duyệt 1/11). */
  constructor(cfg: GameConfig, private readonly onOpenChange: (open: boolean) => void = () => {}) {
    this.top1 = new Top1Slot(cfg);
    this.pods = [new PodSlot(cfg, 2), new PodSlot(cfg, 3)];
    this.rows = Array.from({ length: LIST_SLOTS }, () => new RowSlot(cfg));
    this.meRow = new RowSlot(cfg, true);
    this.miniFaces = [0, 1, 2].map(() => new FaceCell(cfg, 26));

    this.head.append(h('span', 'rb-htitle', '★ ', STR.board.title), this.count);
    this.mini.append(
      ...this.miniFaces.map((f, i) => h('span', `rb-mface r${i + 1}`, ...(i === 0 ? [crown()] : []), f.el)),
      this.miniRank,
    );
    this.mini.title = STR.board.expand;
    const list = h('div', 'rb-list', ...this.rows.map((r) => r.el));
    const pair = h('div', 'rb-pair', ...this.pods.map((p) => p.el));
    this.el.append(this.closeBtn, this.mini, this.head, this.top1.el, pair, list, this.meRow.el);

    this.closeBtn.title = STR.board.close;
    this.closeBtn.addEventListener('click', (e) => {
      e.stopPropagation();
      this.setOpen(false);
    });
    this.el.addEventListener('click', () => this.openIfCompact());
    // Pha BẮT (capture): chạm joystick trên canvas cũng phải đóng bảng, kể cả khi bộ điều
    // khiển chặn sự kiện lan tiếp.
    document.addEventListener('pointerdown', (e) => this.closeIfOutside(e), true);
    this.el.addEventListener('pointerdown', (e) => {
      this.armIdle();
      this.startHold(e);
    });
    this.el.addEventListener('pointerup', () => window.clearTimeout(this.holdTimer));
    window.addEventListener('resize', () => {
      this.natural = 0;
      this.setTier();
    });
    // Bấm TIÊU ĐỀ (không phải cả bảng — cả bảng thì rê chuột xem tên cũng mở) → bảng đầy đủ
    // 10 dòng dạng lớp phủ, kể cả trên PC (§11.5 v6.49).
    this.head.addEventListener('click', (e) => {
      e.stopPropagation();
      this.setOpen(!this.open);
    });
    // "Tự đóng khi di chuyển": phím điều khiển là tín hiệu sớm nhất và không cần ai gọi hộ.
    window.addEventListener('keydown', (e) => {
      if (this.open && MOVE_KEYS.has(e.code)) {
        this.setOpen(false);
      }
    });
    this.setTier();
  }

  /** isOpen: bảng đang mở to trên điện thoại. */
  get isOpen(): boolean {
    return this.open;
  }

  /** close đóng bảng (vd. vừa bị chém trúng — chủ dự án chốt 1/11). */
  close(): void {
    this.setOpen(false);
  }

  /**
   * update nhận gói 1 Hz. Gói THIẾU dòng (lỗi mạng/giải mã, hoặc server cũ gửi danh sách
   * tạm) thì GIỮ BẢNG CŨ — vẽ bảng thiếu đúng là cái lỗi 07/11 người chơi nhìn thấy.
   *
   * Gói đủ thì áp TRỌN trong một `requestAnimationFrame`: tất cả ô đổi trong cùng một khung
   * vẽ, không bao giờ thấy nửa bảng mới nửa bảng cũ.
   */
  update(msg: RoomBoardMsg, ctx: Ctx): void {
    rbStats.packets++;
    rbStats.rows = msg.rows.length;
    rbStats.total = msg.total;
    const want = Math.min(TOP_N, Math.max(msg.total, 0));
    if (msg.rows.length < want) {
      rbStats.short++;
      return;
    }
    this.pending = { msg, ctx };
    if (this.raf !== 0) {
      return;
    }
    this.raf = requestAnimationFrame(() => {
      this.raf = 0;
      const p = this.pending;
      this.pending = null;
      if (p) {
        this.apply(p.msg, p.ctx);
      }
    });
  }

  private apply(msg: RoomBoardMsg, ctx: Ctx): void {
    const compact = narrowScreen();
    this.el.classList.toggle('compact', compact);
    this.el.classList.toggle('open', this.open);
    this.el.classList.toggle('lowfx', quality() === 'low');
    this.count.textContent = STR.board.players(msg.total);

    const rows = msg.rows;
    const at = (i: number): RoomBoardRow | null => rows[i] ?? null;
    this.setTop1(at(0), ctx);
    this.pods[0].set(at(1), ctx);
    this.pods[1].set(at(2), ctx);
    for (let i = 0; i < LIST_SLOTS; i++) {
      this.rows[i].set(at(3 + i), ctx);
    }
    this.setMe(msg, ctx);
    this.setMini(msg);

    const now = performance.now();
    for (const [i, slot] of [this.top1, ...this.pods, ...this.rows].entries()) {
      const row = at(i);
      if (row) {
        slot.arrow.set(this.prevRank.get(row.id), row.rank);
      }
      slot.arrow.tick(now);
    }
    this.slide(rows);
    this.prevRank = new Map(rows.map((r) => [r.id, r.rank]));
    this.slotOf = new Map(rows.map((r, i) => [r.id, i]));
    rbStats.applied++;
    this.fit();
  }

  private setTop1(row: RoomBoardRow | null, ctx: Ctx): void {
    const changedHands = this.top1Id !== 0 && !!row && row.id !== this.top1Id;
    this.top1Id = row?.id ?? 0;
    if (!changedHands || this.lowFx()) {
      this.top1.set(row, ctx);
      return;
    }
    // Top 1 bị hạ: thẻ rung nhẹ, RỒI lật sang người mới kèm chớp vàng (§11.5). Chỉ NỘI DUNG
    // thẻ đổi muộn 0,32 s — không phần tử nào bị xoá nên bảng vẫn đứng yên.
    this.replay(this.top1.el, 'shake');
    window.setTimeout(() => {
      this.top1.set(row, ctx);
      this.replay(this.top1.el, 'flip');
    }, SHAKE_MS);
  }

  /**
   * slide: đổi hạng thì DỊCH Ô bằng `transform` (§11.5 v6.47) — ô nằm nguyên chỗ trong luồng
   * bố cục, chỉ ảnh của nó trượt từ vị trí cũ về. Bản cũ xếp lại phần tử rồi đo
   * `getBoundingClientRect` (FLIP); xếp lại thì số phần tử và chiều cao đổi theo dữ liệu.
   */
  private slide(rows: RoomBoardRow[]): void {
    if (this.lowFx() || this.slotOf.size === 0) {
      return;
    }
    const step = this.rows[0].el.offsetHeight;
    if (step <= 0) {
      return;
    }
    for (let i = 0; i < LIST_SLOTS; i++) {
      const row = rows[3 + i];
      const was = row ? this.slotOf.get(row.id) : undefined;
      // Chỉ trượt trong vùng dòng gọn: thẻ Top cao khác hẳn, trượt qua lại trông như giật.
      if (was === undefined || was === 3 + i || was < 3) {
        continue;
      }
      const dy = (was - (3 + i)) * step;
      this.rows[i].el.animate([{ transform: `translateY(${dy}px)` }, { transform: 'none' }],
        { duration: SLIDE_MS, easing: 'ease-out' });
    }
  }

  /**
   * setMe: dòng của mình, LUÔN hiện — "▶ 14 Cô Ba [LK²] ❄ 120".
   *
   * v6.47 bỏ việc ẩn dòng này khi mình đang trong top 10: ẩn/hiện nó làm chiều cao bảng đổi
   * đúng lúc người chơi trèo vào top — chính kiểu co giãn §11.5 cấm. Dòng ghim trùng với một
   * dòng trên bảng thì cũng không sai gì: nó vẫn là "hạng của bạn".
   */
  private setMe(msg: RoomBoardMsg, ctx: Ctx): void {
    const self = ctx.self;
    const mine = msg.rows.find((r) => r.id === ctx.myId);
    const row: RoomBoardRow = {
      id: ctx.myId, name: self?.name || STR.board.you, character: mine?.character ?? 255,
      titleId: 0, rank: msg.rank, zoneId: ctx.myZone, realm: self?.realm ?? 0,
      tier: self?.tier ?? 1, score: msg.ms, stealth: false,
    };
    this.meRow.set(msg.rank > 0 ? row : null, ctx);
  }

  /** setMini: hàng thu gọn trên điện thoại — 3 chân dung tròn + "Bạn #N" (§11.5). */
  private setMini(msg: RoomBoardMsg): void {
    for (const [i, f] of this.miniFaces.entries()) {
      f.set(msg.rows[i] ?? null);
    }
    this.miniRank.textContent = STR.board.youRank(msg.rank);
  }

  /**
   * fit: bảng không cao quá nửa màn hình (chủ dự án chốt 31/10) bằng cách THU NHỎ cả bảng,
   * không bằng cách bớt dòng.
   *
   * Bản cũ (`trimToScreen`) ẩn dần dòng cuối tới khi vừa `max-height: 50vh`. Hệ quả đo được
   * 07/11: phòng 105 người mà bảng chỉ hiện 5 dòng, và số dòng đổi theo độ dài tên/danh hiệu
   * — hai luật "luôn 10 dòng" và "không quá nửa màn hình" không thể cùng đúng nếu chỉ có một
   * cỡ chữ. Hệ số thu nhỏ tính từ CHIỀU CAO MÀN HÌNH chứ không từ nội dung, nên trong lúc
   * chơi nó là một con số không đổi.
   *
   * (Ghi vào docs/QUESTIONS.md — Q49: chủ dự án có thể muốn chữ nhỏ hơn thay vì thu cả bảng.)
   */
  /** setTier: áp tầng chiều cao — đổi class trên gốc, CSS lo phần ẩn ô thừa. */
  private setTier(): void {
    const t = heightTier();
    if (t === this.tier && this.el.classList.contains(t.cls || 'h-tall')) {
      return;
    }
    this.tier = t;
    for (const x of HEIGHT_TIERS) {
      this.el.classList.toggle(x.cls || 'h-tall', x === t);
    }
  }

  private fit(): void {
    // Thu gọn (điện thoại) hoặc đang mở lớp phủ: để nguyên cỡ thật. Lớp phủ do người chơi
    // CHỦ ĐỘNG mở nên được phép cao hơn nửa màn hình (§11.5).
    if (this.open || this.el.classList.contains('compact')) {
      this.el.style.transform = '';
      this.natural = 0;
      return;
    }
    // `offsetHeight` là hộp BỐ CỤC — `transform` không ảnh hưởng, nên đo được mà không phải
    // bỏ hệ số cũ ra trước.
    //
    // Đo lại mỗi gói nhưng chỉ tính lại khi con số ĐỔI. Chiều cao tự nhiên của bảng không
    // phụ thuộc dữ liệu (ô cố định), nên nó chỉ đổi khi cửa sổ đổi cỡ, khi phông chữ về
    // muộn, hoặc khi stylesheet chưa kịp áp. Lần chạy đầu bắt được đúng ca cuối: bảng chưa
    // có CSS thì cao 22 500 px, hệ số ra 0,024 và nếu chỉ tính MỘT LẦN thì bảng bé bằng hạt
    // gạo suốt cả ván.
    const natural = this.el.offsetHeight;
    if (natural === this.natural || natural <= 0) {
      return;
    }
    this.natural = natural;
    // Sàn 0,85 (v6.49): số dòng đã giảm theo tầng chiều cao rồi, nên hệ số chỉ còn lo phần
    // lẻ. Thu nhỏ sâu hơn nữa là chữ bảng bé hơn chữ HUD, đọc không ra.
    const k = Math.max(MIN_SCALE, Math.min(1, (window.innerHeight * MAX_SCREEN_FRAC) / natural));
    this.el.style.transform = k < 1 ? `scale(${k.toFixed(3)})` : '';
  }

  private lowFx(): boolean {
    return quality() === 'low';
  }

  /** replay chạy lại một hoạt ảnh CSS dù class đó vừa mới chạy xong. */
  private replay(el: HTMLElement, cls: string): void {
    el.classList.remove(cls);
    void el.offsetWidth;
    el.classList.add(cls);
  }

  private openIfCompact(): void {
    if (this.open || !this.el.classList.contains('compact')) {
      return;
    }
    this.setOpen(true);
  }

  private setOpen(open: boolean): void {
    if (open === this.open) {
      return;
    }
    this.open = open;
    this.el.classList.toggle('open', open);
    // Lớp phủ trên PC: cũng là `open`, nhưng CSS của điện thoại gắn với `.compact`. Lúc mở
    // thì bỏ hệ số thu nhỏ đi — lớp phủ được phép cao hơn nửa màn hình.
    this.el.style.transform = '';
    this.natural = 0;
    window.clearTimeout(this.idleTimer);
    if (open) {
      this.armIdle();
    }
    this.onOpenChange(open);
  }

  /** armIdle: đếm lại 8 s không tương tác thì tự đóng. */
  private armIdle(): void {
    window.clearTimeout(this.idleTimer);
    if (this.open) {
      this.idleTimer = window.setTimeout(() => this.setOpen(false), IDLE_CLOSE_MS);
    }
  }

  private closeIfOutside(e: PointerEvent): void {
    if (this.open && !this.el.contains(e.target as Node)) {
      this.setOpen(false);
    }
  }

  /** startHold: nhấn giữ một chữ bị cắt "…" thì hiện đủ (điện thoại không có rê chuột). */
  private startHold(e: PointerEvent): void {
    const t = (e.target as HTMLElement).closest('.rb-clip');
    if (!t) {
      return;
    }
    this.holdTimer = window.setTimeout(() => t.classList.add('full'), HOLD_MS);
    window.setTimeout(() => t.classList.remove('full'), HOLD_MS + 2500);
  }
}
