import type { GameConfig } from '../config';
import type { RoomBoardMsg, RoomBoardRow } from '../net/protocol';
import { assets } from '../render/assets';
import { quality } from '../render/quality';
import { h } from './dom';
import { narrowScreen } from './layout';
import { clip, crown, face, fmtScore, medal, realmPill, titleText, zoneChip } from './roomboard-parts';
import { STR } from './strings';

/**
 * Bảng xếp hạng phòng góc trên phải — GAME_SPEC §11.5 (v6.19).
 *
 * Top 1 là thẻ to viền vàng, Top 2/3 là hai thẻ nửa rộng bạc/đồng, hạng 4–10 là dòng gọn,
 * dòng của mình ghim cuối khi ngoài top 10. Điện thoại: mặc định thu gọn thành một hàng
 * (3 chân dung + "Bạn #N"), chạm để mở, chạm ngoài để đóng.
 *
 * Hiệu năng (§11.5): dữ liệu tới 1 lần/giây, mỗi thẻ/dòng có CHỮ KÝ — chữ ký không đổi thì
 * không đụng vào DOM. Hiệu ứng động (vệt sáng, lấp lánh, lật thẻ) là CSS chồng lên, chất
 * lượng Thấp tắt hẳn bằng một class trên gốc.
 */

/** Thời gian trượt dòng khi đổi hạng (§11.5: 0.25 s). */
const SLIDE_MS = 250;
/** Thẻ Top 1 rung trước khi đổi người (§11.5). */
const SHAKE_MS = 320;
/** Nhấn giữ bao lâu thì hiện đủ tên/danh hiệu bị cắt "…" (§11.5). */
const HOLD_MS = 450;
/** Điện thoại: bảng đang mở mà không ai chạm vào ngần này thì tự đóng (chủ dự án chốt 1/11). */
const IDLE_CLOSE_MS = 8000;

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

/** rowSig: mọi thứ hiển thị của một người — đổi bất cứ gì thì dòng đó mới vẽ lại. */
function rowSig(r: RoomBoardRow, ctx: Ctx): string {
  return [r.id, r.name, r.character, r.titleId, r.rank, r.zoneId, r.realm, r.tier, r.score, r.id === ctx.myId, r.zoneId === ctx.myZone].join('|');
}

export class RoomBoard {
  readonly el = h('div', 'room-board rb');
  private readonly head = h('div', 'rb-head');
  private readonly top1 = h('div', 'rb-top1');
  private readonly pair = h('div', 'rb-pair');
  private readonly list = h('div', 'rb-list');
  private readonly me = h('div', 'rb-row rb-me');
  private readonly mini = h('div', 'rb-mini');
  private readonly sigs = new Map<HTMLElement, string>();
  private readonly rowEls = new Map<number, HTMLElement>();
  private prevRank = new Map<number, number>();
  private top1Id = 0;
  private open = false;
  private holdTimer = 0;

  private idleTimer = 0;
  private readonly closeBtn = h('button', 'rb-close', '✕');

  /** onOpenChange: báo HUD để ẩn nút kỹ năng/🔥, hoãn thông báo khi bảng mở (§11.5, duyệt 1/11). */
  constructor(private readonly cfg: GameConfig, private readonly onOpenChange: (open: boolean) => void = () => {}) {
    this.closeBtn.title = STR.board.close;
    this.closeBtn.addEventListener('click', (e) => {
      e.stopPropagation();
      this.setOpen(false);
    });
    this.el.append(this.closeBtn, this.mini, this.head, this.top1, this.pair, this.list, this.me);
    this.el.addEventListener('click', () => this.openIfCompact());
    // Pha BẮT (capture): chạm joystick trên canvas cũng phải đóng bảng, kể cả khi bộ điều
    // khiển chặn sự kiện lan tiếp.
    document.addEventListener('pointerdown', (e) => this.closeIfOutside(e), true);
    this.el.addEventListener('pointerdown', (e) => {
      this.armIdle();
      this.startHold(e);
    });
    this.el.addEventListener('pointerup', () => window.clearTimeout(this.holdTimer));
  }

  /** isOpen: bảng đang mở to trên điện thoại. */
  get isOpen(): boolean {
    return this.open;
  }

  /** close đóng bảng (vd. vừa bị chém trúng — chủ dự án chốt 1/11). */
  close(): void {
    this.setOpen(false);
  }

  update(msg: RoomBoardMsg, ctx: Ctx): void {
    const compact = narrowScreen();
    if (!compact && this.open) {
      this.setOpen(false);
    }
    this.el.classList.toggle('compact', compact);
    this.el.classList.toggle('open', compact && this.open);
    this.el.classList.toggle('lowfx', quality() === 'low');
    this.setHead(msg);
    const rows = msg.rows;
    this.setTop1(rows[0], ctx);
    this.setPair(rows[1], rows[2], ctx);
    this.setList(rows.slice(3), ctx);
    this.setMe(msg, ctx);
    this.setMini(msg, ctx);
    this.prevRank = new Map(rows.map((r) => [r.id, r.rank]));
    this.trimToScreen();
  }

  private setHead(msg: RoomBoardMsg): void {
    this.paint(this.head, String(msg.total), () => [
      h('span', 'rb-htitle', '★ ', STR.board.title),
      h('span', 'rb-count', STR.board.players(msg.total)),
    ]);
  }

  /** paint thay nội dung `el` chỉ khi chữ ký đổi. */
  private paint(el: HTMLElement, sig: string, build: () => (Node | string)[]): boolean {
    if (this.sigs.get(el) === sig) {
      return false;
    }
    this.sigs.set(el, sig);
    el.replaceChildren(...build());
    return true;
  }

  private setTop1(row: RoomBoardRow | undefined, ctx: Ctx): void {
    this.top1.style.display = row ? '' : 'none';
    if (!row) {
      return;
    }
    const changedHands = this.top1Id !== 0 && row.id !== this.top1Id;
    this.top1Id = row.id;
    const draw = (): void => {
      this.paint(this.top1, rowSig(row, ctx), () => this.top1Body(row, ctx));
      this.top1.classList.toggle('me', row.id === ctx.myId);
    };
    if (!changedHands || this.lowFx()) {
      draw();
      return;
    }
    // Top 1 bị hạ: thẻ rung nhẹ, RỒI lật sang người mới kèm chớp vàng (§11.5).
    this.replay(this.top1, 'shake');
    window.setTimeout(() => {
      draw();
      this.replay(this.top1, 'flip');
    }, SHAKE_MS);
  }

  private top1Body(row: RoomBoardRow, ctx: Ctx): (Node | string)[] {
    const sparks = [1, 2, 3, 4].map((i) => h('i', `rb-spark s${i}`));
    const shine = h('i', 'rb-shine');
    const streak = assets.url('ui/shine_streak.png');
    if (streak) {
      shine.style.backgroundImage = `url(${streak})`;
    }
    return [
      shine, ...sparks,
      h('div', 'rb-portrait', crown(), face(this.cfg, row, 40)),
      h('div', 'rb-info',
        h('div', 'rb-line', clip('rb-name', row.name), h('span', 'rb-score', fmtScore(row.score))),
        clip('rb-title', titleText(row)),
        h('div', 'rb-tags', realmPill(this.cfg, row, false), zoneChip(row, ctx.myZone)),
      ),
    ];
  }

  private setPair(r2: RoomBoardRow | undefined, r3: RoomBoardRow | undefined, ctx: Ctx): void {
    const sig = [r2, r3].map((r) => (r ? rowSig(r, ctx) : '-')).join('#');
    this.pair.style.display = r2 ? '' : 'none';
    this.paint(this.pair, sig, () => [r2, r3].filter((r): r is RoomBoardRow => !!r).map((r) => this.podiumCard(r, ctx)));
  }

  private podiumCard(row: RoomBoardRow, ctx: Ctx): HTMLElement {
    const rank = row.rank === 2 ? 2 : 3;
    const card = h('div', `rb-pod p${rank}${row.id === ctx.myId ? ' me' : ''}`);
    // Tên lên dòng RIÊNG, đủ cả bề rộng thẻ (duyệt 1/11: chung dòng với chân dung thì tên
    // chỉ còn 4 ký tự). Dòng trên: huy chương + chân dung bên trái, Uy Danh góc phải.
    card.append(
      h('i', 'rb-shine'),
      h('div', 'rb-line', medal(rank), face(this.cfg, row, 28), h('span', 'rb-score', fmtScore(row.score))),
      clip('rb-name', row.name),
      clip('rb-title', titleText(row)),
      h('div', 'rb-tags', realmPill(this.cfg, row, true), zoneChip(row, ctx.myZone)),
    );
    this.addArrow(card, row);
    return card;
  }

  /**
   * setList vẽ hạng 4–10, GIỮ phần tử của từng người qua các lần cập nhật để làm hoạt ảnh
   * trượt kiểu FLIP: đo vị trí cũ, xếp lại, rồi cho từng dòng trượt từ chỗ cũ về chỗ mới.
   */
  private setList(rows: RoomBoardRow[], ctx: Ctx): void {
    const before = new Map<number, number>();
    for (const [id, el] of this.rowEls) {
      before.set(id, el.getBoundingClientRect().top);
    }
    const keep = new Set(rows.map((r) => r.id));
    for (const [id, el] of this.rowEls) {
      if (!keep.has(id)) {
        el.remove();
        this.rowEls.delete(id);
      }
    }
    const els = rows.map((r) => this.listRow(r, ctx));
    this.list.replaceChildren(...els);
    this.slide(before);
  }

  private listRow(row: RoomBoardRow, ctx: Ctx): HTMLElement {
    let el = this.rowEls.get(row.id);
    if (!el) {
      el = h('div', 'rb-row');
      this.rowEls.set(row.id, el);
    }
    const target = el;
    target.classList.toggle('me', row.id === ctx.myId);
    if (this.paint(target, rowSig(row, ctx), () => this.rowBody(row, ctx))) {
      this.addArrow(target, row);
    }
    return target;
  }

  private rowBody(row: RoomBoardRow, ctx: Ctx): (Node | string)[] {
    return [
      h('div', 'rb-line',
        h('span', 'rb-rank', String(row.rank)), clip('rb-name', row.name),
        realmPill(this.cfg, row, true), zoneChip(row, ctx.myZone), h('span', 'rb-score', fmtScore(row.score))),
      clip('rb-sub', titleText(row)),
    ];
  }

  /** slide: hoạt ảnh trượt 0.25 s từ vị trí cũ (FLIP). Chất lượng Thấp: nhảy thẳng. */
  private slide(before: Map<number, number>): void {
    if (this.lowFx()) {
      return;
    }
    for (const [id, el] of this.rowEls) {
      const old = before.get(id);
      const dy = old === undefined ? 0 : old - el.getBoundingClientRect().top;
      if (Math.abs(dy) < 1) {
        continue;
      }
      el.animate([{ transform: `translateY(${dy}px)` }, { transform: 'none' }], { duration: SLIDE_MS, easing: 'ease-out' });
    }
  }

  /** addArrow: ▲ xanh khi lên hạng, ▼ đỏ khi xuống, mờ dần trong 2 s (§11.5). */
  private addArrow(el: HTMLElement, row: RoomBoardRow): void {
    const old = this.prevRank.get(row.id);
    if (old === undefined || old === row.rank) {
      return;
    }
    const up = row.rank < old;
    el.querySelector('.rb-line')?.append(h('span', `rb-arrow ${up ? 'up' : 'down'}`, up ? '▲' : '▼'));
  }

  /** setMe: dòng của mình ghim cuối khi ngoài top 10 — "▶ 14 Cô Ba [LK²] ❄ 120". */
  private setMe(msg: RoomBoardMsg, ctx: Ctx): void {
    const inTop = msg.rows.some((r) => r.id === ctx.myId);
    this.me.style.display = inTop || msg.rank <= 0 ? 'none' : '';
    const self = ctx.self;
    const row: RoomBoardRow = {
      id: ctx.myId, name: self?.name ?? '', character: 255, titleId: 0, rank: msg.rank,
      zoneId: ctx.myZone, realm: self?.realm ?? 0, tier: self?.tier ?? 1, score: msg.ms, stealth: false,
    };
    // Chỉ ▶ + hạng + TÊN, nền nổi bật đã nói "đây là bạn" (duyệt 1/11: bỏ "Bạn (…)").
    const who = self?.name || STR.board.you;
    this.paint(this.me, `${rowSig(row, ctx)}|${who}`, () => [
      h('div', 'rb-line', h('span', 'rb-rank', `▶ ${msg.rank}`), clip('rb-name', who),
        realmPill(this.cfg, row, true), zoneChip(row, -1),
        h('span', 'rb-score', fmtScore(msg.ms))),
    ]);
  }

  /** setMini: hàng thu gọn trên điện thoại — 3 chân dung tròn + "Bạn #N" (§11.5). */
  private setMini(msg: RoomBoardMsg, ctx: Ctx): void {
    const top3 = msg.rows.slice(0, 3);
    const sig = `${top3.map((r) => `${r.id}:${r.character}:${r.realm}`).join(',')}|${msg.rank}|${ctx.myId}`;
    this.paint(this.mini, sig, () => [
      ...top3.map((r) => h('span', `rb-mface r${r.rank}`, ...(r.rank === 1 ? [crown()] : []), face(this.cfg, r, 26))),
      h('span', 'rb-mrank', STR.board.youRank(msg.rank)),
    ]);
    this.mini.title = STR.board.expand;
  }

  /**
   * trimToScreen: bảng không bao giờ quá nửa màn hình (chủ dự án chốt 31/10). Ẩn bớt dòng
   * cuối của danh sách 4–10 tới khi vừa — ẩn chứ không xoá, để lần sau còn trượt được.
   */
  private trimToScreen(): void {
    // Thứ tự HIỂN THỊ (con của `list`), không phải thứ tự chèn vào `rowEls` — dùng Map thì
    // ẩn nhầm hạng 4–6 vì đó là những dòng được tạo trước (ảnh chụp 1/11).
    const rows = [...this.list.children] as HTMLElement[];
    rows.forEach((el) => (el.style.display = ''));
    for (let i = rows.length - 1; i >= 0 && this.el.scrollHeight > this.el.clientHeight; i--) {
      rows[i].style.display = 'none';
    }
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
    window.clearTimeout(this.idleTimer);
    if (open) {
      this.armIdle();
    }
    this.trimToScreen();
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
