import { sound } from '../audio/sound';
import type { GameConfig } from '../config';
import { maxSwords, realmOf, tierOf } from '../config';
import type { FeedMsg, GameEvent, LeaderboardMsg, PlayerState } from '../net/protocol';
import { BUFFS } from '../net/protocol';
import { assets } from '../render/assets';
import { BUFF_COLORS } from '../render/palette';
import { button, h } from './dom';
import { Minimap } from './minimap';
import { STR } from './strings';

const FEED_KEEP = 5;
const FEED_SECS = 6;
// Tin loan toàn server: ít dòng hơn nhưng ở lại lâu hơn vì nó là chuyện lớn (spec §11.4).
const SERVER_FEED_KEEP = 3;
const SERVER_FEED_SECS = 9;

/**
 * colorizeFeed splits a kill-feed sentence into coloured pieces so it can be read at a glance:
 * who did it (vàng), what happened (trắng), and to whom (đỏ). The sentences come from the joke
 * pools in strings.ts, so the split is done by finding the two names inside the finished text
 * rather than by rebuilding every variant.
 */
function colorizeFeed(text: string, actor: string, target: string): Node[] {
  const marks: { at: number; len: number; cls: string }[] = [];
  const findAll = (name: string, cls: string): void => {
    if (!name) {
      return;
    }
    for (let i = text.indexOf(name); i >= 0; i = text.indexOf(name, i + name.length)) {
      marks.push({ at: i, len: name.length, cls });
    }
  };
  // Tên dài tìm trước: tên này có thể nằm lọt trong tên kia.
  const [first, second] = actor.length >= target.length ? [actor, target] : [target, actor];
  findAll(first, first === actor ? 'f-actor' : 'f-target');
  findAll(second, second === actor ? 'f-actor' : 'f-target');
  marks.sort((a, b) => a.at - b.at);
  const out: Node[] = [];
  let i = 0;
  for (const m of marks) {
    if (m.at < i) {
      continue; // chồng lên đoạn đã tô
    }
    if (m.at > i) {
      out.push(h('span', 'f-act', text.slice(i, m.at)));
    }
    out.push(h('span', m.cls, text.slice(m.at, m.at + m.len)));
    i = m.at + m.len;
  }
  if (i < text.length) {
    out.push(h('span', 'f-act', text.slice(i)));
  }
  return out;
}

/** Hud is the in-match DOM overlay: EXP bar, swords, buffs, room top 10, kill feed (spec §6, §11). */
export class Hud {
  readonly root = h('div', 'hud');
  private readonly badge = h('div', 'badge');
  private readonly realm = h('div', 'realm');
  private readonly expFill = h('div', 'exp-fill');
  private readonly swords = h('div', 'swords');
  private readonly scoreBox = h('div', 'score');
  private readonly pops = h('div', 'score-pops');
  private readonly buffs = h('div', 'buffs');
  private readonly board = h('div', 'room-board');
  private readonly feed = h('div', 'feed');
  /** serverFeed is the top-centre strip for server-wide announcements (spec §11.4). */
  private readonly serverFeed = h('div', 'server-feed');
  private readonly banner = h('div', 'banner');
  private readonly ping = h('div', 'ping');
  private readonly debugHelp = h('div', 'debug-help', STR.debugHelp);
  private buffEnds = new Map<string, { end: number; total: number }>();
  private bannerUntil = 0;
  private lastLevel = -1;

  readonly minimap: Minimap;
  /** burnBtn is the touch-only hold button for Thiêu Đốt Tinh Huyết (spec §4.1c). */
  readonly burnBtn = h('button', 'burn-btn', '🔥');
  private readonly guide = h('div', 'guide');
  private readonly isTouch: boolean;

  constructor(private readonly cfg: GameConfig, isTouch = false) {
    this.isTouch = isTouch;
    this.minimap = new Minimap(cfg, isTouch);
    this.burnBtn.title = STR.burnName;
    this.burnBtn.style.display = isTouch ? 'flex' : 'none';
    const mute = button('🔊', 'mute', () => {
      mute.textContent = sound.toggleMute() ? '🔇' : '🔊';
    });
    this.root.append(
      this.minimap.root,
      h('div', 'status', this.badge, h('div', 'status-text', this.realm, h('div', 'exp-bar', this.expFill), this.swords, this.scoreBox)),
      this.pops,
      this.buffs,
      this.board,
      this.feed,
      this.serverFeed,
      this.banner,
      this.burnBtn,
      this.guide,
      h('div', 'corner', this.ping, mute),
      this.debugHelp,
      h('div', 'rotate-hint', STR.rotateHint),
    );
    this.debugHelp.style.display = 'none';
  }

  /**
   * showGuide fades a short "how to play" card over the first seconds of a match so a new
   * player learns the rules while they are still invulnerable (spec §18.1).
   */
  showGuide(): void {
    this.guide.replaceChildren(
      h('div', 'guide-title', STR.guideTitle),
      ...STR.guideLines(this.isTouch).map((t) => h('div', 'guide-line', t)),
    );
    this.guide.classList.remove('fade');
    // Khởi động lại animation: trình duyệt bỏ qua nếu chỉ gỡ rồi gắn class ngay.
    void this.guide.offsetWidth;
    this.guide.classList.add('fade');
  }

  setDebug(on: boolean): void {
    this.debugHelp.style.display = on ? 'block' : 'none';
  }

  update(me: PlayerState | null, rttMs: number, prog: (L: number) => { from: number; to: number }): void {
    const now = performance.now() / 1000;
    this.ping.textContent = STR.ping(rttMs);
    if (now > this.bannerUntil) {
      this.banner.style.display = 'none';
    }
    this.updateBuffs(me, now);
    if (!me) {
      return;
    }
    const r = realmOf(this.cfg, me.L);
    if (me.L !== this.lastLevel) {
      this.lastLevel = me.L;
      this.setBadge(r);
    }
    const name = this.cfg.cultivation.realms[r] ?? '';
    this.realm.textContent = STR.realmTier(name, tierOf(this.cfg, me.L));
    const { from, to } = prog(me.L);
    const frac = to > from ? Math.min(1, (me.xp - from) / (to - from)) : 1;
    this.expFill.style.width = `${(frac * 100).toFixed(1)}%`;
    this.swords.textContent = STR.swords(me.n, maxSwords(this.cfg, r));
  }

  private setBadge(realm: number): void {
    const path = `ui/badge_${realm}.png`;
    this.badge.style.backgroundImage = assets.has(path) ? `url(/assets/${path})` : '';
    this.badge.textContent = assets.has(path) ? '' : String(realm + 1);
  }

  private updateBuffs(me: PlayerState | null, now: number): void {
    const active = BUFFS.filter((_, i) => me && me.bf & (1 << i));
    const nodes = active.map((b) => {
      const t = this.buffEnds.get(b);
      const left = t ? Math.max(0, t.end - now) : 0;
      const frac = t ? left / t.total : 1;
      const icon = h('div', 'buff');
      icon.title = STR.buffs[b] ?? b;
      const color = `#${(BUFF_COLORS[b] ?? 0xaaaaaa).toString(16).padStart(6, '0')}`;
      icon.style.background = `conic-gradient(${color} ${frac * 360}deg, rgba(0,0,0,.45) 0)`;
      const inner = h('div', 'buff-inner');
      if (assets.has(`buffs/${b}.png`)) {
        inner.style.backgroundImage = `url(/assets/buffs/${b}.png)`;
      } else {
        inner.textContent = (STR.buffs[b] ?? b).slice(0, 2);
      }
      icon.append(inner, h('span', 'buff-left', left > 0 ? String(Math.ceil(left)) : ''));
      return icon;
    });
    this.buffs.replaceChildren(...nodes);
  }

  /** onEvents tracks my buff timers and the chest warning. */
  onEvents(events: GameEvent[], myId: number): void {
    const now = performance.now() / 1000;
    for (const e of events) {
      if (e.k === 'Score' && e.a === myId) {
        // Điểm vượt mức cơ bản của một lần chém = có nhân lật kèo (spec §11.1).
        const base = this.cfg.score?.hit_base ?? 100;
        this.popScore(e.n ?? 0, e.s ?? '', e.s === 'hit' && (e.n ?? 0) > base);
      } else if (e.k === 'BuffStart' && e.a === myId && e.s) {
        const total = (e.n ?? 0) / 1000;
        this.buffEnds.set(e.s, { end: now + total, total });
      } else if (e.k === 'ChestIncoming') {
        this.showBanner(STR.chestIncoming(e.n ?? this.cfg.items.chest_warn_secs), 4);
      }
    }
  }

  showBanner(text: string, secs: number): void {
    this.banner.textContent = text;
    this.banner.style.display = 'block';
    this.bannerUntil = performance.now() / 1000 + secs;
  }

  /** setLeaderboard draws the room top 10: realm+tier order, Uy Danh, Top 1–3 badges (§11.2). */
  setLeaderboard(lb: LeaderboardMsg, myId: number): void {
    this.minimap.setLeaderboard(lb);
    this.scoreBox.textContent = STR.scoreHud(lb.ms ?? 0);
    // Xếp theo Uy Danh; mỗi dòng ghi danh hiệu thay cho cảnh giới (cảnh giới đã hiện
    // trên đầu nhân vật trong trận — spec §11.2).
    const rows = lb.top.map((e, i) => {
      const line = h('div', `lb-line${e.id === myId ? ' me' : ''}`);
      const who = h('span', 'lb-who', `${i + 1}. ${e.nm}`);
      if (e.ti === 1) {
        who.prepend(h('span', 'lb-crown', '👑'));
      }
      const titleText = STR.titles[(e.ti || i + 1) - 1] ?? '';
      line.append(
        h('div', 'lb-main', who, h('span', 'lb-score', Math.round(e.sc ?? 0).toLocaleString('vi-VN'))),
        h('div', `lb-title t${Math.min(e.ti || i + 1, 3)}`, titleText),
      );
      return line;
    });
    this.board.replaceChildren(h('div', 'board-title', STR.roomTop), ...rows, h('div', 'my-rank', STR.yourRank(lb.rank, lb.total)));
  }

  /** popScore floats a "+800 LẬT KÈO!" number up when Uy Danh is earned (spec §11.1). */
  private popScore(points: number, source: string, upset: boolean): void {
    const el = h('div', `score-pop ${source}`, upset ? `${STR.scorePop(points)} ${STR.scoreUpset}` : STR.scorePop(points));
    this.pops.prepend(el);
    while (this.pops.children.length > 4) {
      this.pops.lastElementChild?.remove();
    }
    window.setTimeout(() => el.remove(), 1600);
  }

  addFeed(f: FeedMsg): void {
    const realm = this.cfg.cultivation.realms[f.r] ?? '';
    let text =
      f.k === 'kill' ? STR.feed.kill(f.a, f.b ?? '')
      : f.k === 'realm_drop' ? STR.feed.realm_drop(f.a, f.b ?? '', realm)
      : f.k === 'breakthrough' ? STR.feed.breakthrough(f.a, realm)
      : f.k === 'gate_kill' ? STR.feed.gate_kill(f.a, f.b ?? '')
      : f.k === 'title_gain' ? STR.feed.title_gain(f.a)
      : f.k === 'title_lose' ? STR.feed.title_lose(f.a)
      : '';
    if (!text) {
      return;
    }
    // Tin loan toàn server đi lên dải giữa-trên màn hình, không lẫn vào feed góc trái.
    if ((f.tp ?? 0) > 0) {
      const big = h('div', 'server-line');
      big.append(h('span', 'feed-tag', STR.feed.serverTag(f.tp ?? 0)), ...colorizeFeed(text, f.a, f.b ?? ''));
      this.serverFeed.prepend(big);
      while (this.serverFeed.children.length > SERVER_FEED_KEEP) {
        this.serverFeed.lastElementChild?.remove();
      }
      window.setTimeout(() => big.remove(), SERVER_FEED_SECS * 1000);
      return;
    }
    const line = h('div', 'feed-line', ...colorizeFeed(text, f.a, f.b ?? ''));
    this.feed.prepend(line);
    while (this.feed.children.length > FEED_KEEP) {
      this.feed.lastElementChild?.remove();
    }
    window.setTimeout(() => line.remove(), FEED_SECS * 1000);
  }
}
