import { sound } from '../audio/sound';
import type { GameConfig } from '../config';
import { maxSwords, realmOf, tierOf } from '../config';
import type { FeedMsg, GameEvent, LeaderboardMsg, PlayerState } from '../net/protocol';
import { BUFFS } from '../net/protocol';
import { assets } from '../render/assets';
import { BUFF_COLORS } from '../render/palette';
import { button, h } from './dom';
import { STR } from './strings';

const FEED_KEEP = 5;
const FEED_SECS = 6;

/** Hud is the in-match DOM overlay: EXP bar, swords, buffs, room top 10, kill feed (spec §6, §11). */
export class Hud {
  readonly root = h('div', 'hud');
  private readonly badge = h('div', 'badge');
  private readonly realm = h('div', 'realm');
  private readonly expFill = h('div', 'exp-fill');
  private readonly swords = h('div', 'swords');
  private readonly buffs = h('div', 'buffs');
  private readonly board = h('div', 'room-board');
  private readonly feed = h('div', 'feed');
  private readonly banner = h('div', 'banner');
  private readonly ping = h('div', 'ping');
  private readonly debugHelp = h('div', 'debug-help', STR.debugHelp);
  private buffEnds = new Map<string, { end: number; total: number }>();
  private bannerUntil = 0;
  private lastLevel = -1;

  constructor(private readonly cfg: GameConfig) {
    const mute = button('🔊', 'mute', () => {
      mute.textContent = sound.toggleMute() ? '🔇' : '🔊';
    });
    this.root.append(
      h('div', 'status', this.badge, h('div', 'status-text', this.realm, h('div', 'exp-bar', this.expFill), this.swords)),
      this.buffs,
      this.board,
      this.feed,
      this.banner,
      h('div', 'corner', this.ping, mute),
      this.debugHelp,
    );
    this.debugHelp.style.display = 'none';
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
      if (e.k === 'BuffStart' && e.a === myId && e.s) {
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

  setLeaderboard(lb: LeaderboardMsg, myId: number): void {
    const rows = lb.top.map((e, i) =>
      h('div', `lb-line${e.id === myId ? ' me' : ''}`, h('span', '', `${i + 1}. ${e.nm}`), h('span', '', Math.round(e.xp).toLocaleString('vi-VN'))),
    );
    this.board.replaceChildren(h('div', 'board-title', STR.roomTop), ...rows, h('div', 'my-rank', STR.yourRank(lb.rank, lb.total)));
  }

  addFeed(f: FeedMsg): void {
    const realm = this.cfg.cultivation.realms[f.r] ?? '';
    const text =
      f.k === 'kill' ? STR.feed.kill(f.a, f.b ?? '')
      : f.k === 'realm_drop' ? STR.feed.realm_drop(f.a, f.b ?? '', realm)
      : f.k === 'breakthrough' ? STR.feed.breakthrough(f.a, realm)
      : f.k === 'gate_kill' ? STR.feed.gate_kill(f.a, f.b ?? '')
      : '';
    if (!text) {
      return;
    }
    const line = h('div', 'feed-line', text);
    this.feed.prepend(line);
    while (this.feed.children.length > FEED_KEEP) {
      this.feed.lastElementChild?.remove();
    }
    window.setTimeout(() => line.remove(), FEED_SECS * 1000);
  }
}
