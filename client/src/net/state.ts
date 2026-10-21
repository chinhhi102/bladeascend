import { type GameConfig } from '../config';
import { contains, lerp, pushOut, type Vec } from './geom';
import {
  BUFFS,
  type GameEvent,
  type GameMap,
  type ItemState,
  type MonsterState,
  type PlayerState,
  type Snapshot,
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
}

interface PendingInput {
  seq: number;
  angle: number;
  moving: boolean;
}

/** What the renderer needs for one player this frame. */
export interface PlayerFrame {
  view: PlayerView;
  s: PlayerState;
  x: number;
  y: number;
  ringAngle: number;
}

const SNAP_KEEP = 30;
const OFFSET_DECAY = 0.0005; // lets the clock estimate adapt if latency grows

/** ClientWorld holds the replicated world: interpolation for others, prediction for me. */
export class ClientWorld {
  myId = 0;
  map: GameMap | null = null;
  readonly players = new Map<number, PlayerView>();
  readonly monsters = new Map<number, Timed<MonsterState>[]>();
  readonly items = new Map<number, ItemState>();
  readonly brokenObstacles = new Map<number, boolean>();
  readonly obstacleHp = new Map<number, number>();
  private offset: number | null = null;
  private pending: PendingInput[] = [];
  private seq = 0;
  private predicted: Vec = { x: 0, y: 0 };
  private correction: Vec = { x: 0, y: 0 };
  private lastSendAt = 0;
  private lastInput: PendingInput = { seq: 0, angle: 0, moving: false };
  lastSnapAt = 0;

  constructor(private readonly cfg: GameConfig) {}

  reset(w: Welcome, map: GameMap): void {
    this.myId = w.id;
    this.map = map;
    this.players.clear();
    this.monsters.clear();
    this.items.clear();
    this.brokenObstacles.clear();
    this.obstacleHp.clear();
    for (const o of w.obs ?? []) {
      this.obstacleHp.set(o.id, o.hp);
      this.brokenObstacles.set(o.id, o.b);
    }
    this.pending = [];
    this.offset = null;
  }

  /** Estimated current server time in seconds. */
  serverNow(): number {
    return performance.now() / 1000 + (this.offset ?? 0);
  }

  renderTime(): number {
    return this.serverNow() - this.cfg.net.interp_ms / 1000;
  }

  applySnapshot(s: Snapshot): GameEvent[] {
    const est = s.tm - performance.now() / 1000;
    this.offset = this.offset === null || est > this.offset ? est : this.offset - OFFSET_DECAY;
    this.lastSnapAt = performance.now();
    this.updatePlayers(s);
    this.updateMonsters(s);
    for (const it of s.ia ?? []) {
      this.items.set(it.id, it);
    }
    for (const id of s.ir ?? []) {
      this.items.delete(id);
    }
    const events = s.ev ?? [];
    this.applyObstacleEvents(events);
    this.reconcile(s.ack);
    return events;
  }

  private updatePlayers(s: Snapshot): void {
    const seen = new Set<number>();
    for (const p of s.p ?? []) {
      seen.add(p.id);
      let v = this.players.get(p.id);
      if (!v) {
        v = { id: p.id, name: p.nm ?? '?', char: p.ch ?? '', snaps: [] };
        this.players.set(p.id, v);
      }
      if (p.nm) {
        v.name = p.nm;
        v.char = p.ch ?? v.char;
      }
      v.snaps.push({ t: s.tm, s: p });
      if (v.snaps.length > SNAP_KEEP) {
        v.snaps.shift();
      }
    }
    for (const id of this.players.keys()) {
      if (!seen.has(id)) {
        this.players.delete(id);
      }
    }
  }

  private updateMonsters(s: Snapshot): void {
    const seen = new Set<number>();
    for (const m of s.m ?? []) {
      seen.add(m.id);
      const buf = this.monsters.get(m.id) ?? [];
      buf.push({ t: s.tm, s: m });
      if (buf.length > SNAP_KEEP) {
        buf.shift();
      }
      this.monsters.set(m.id, buf);
    }
    for (const id of this.monsters.keys()) {
      if (!seen.has(id)) {
        this.monsters.delete(id);
      }
    }
  }

  private applyObstacleEvents(events: GameEvent[]): void {
    for (const e of events) {
      const id = e.a ?? -1;
      if (e.k === 'ObstacleDamaged') {
        this.obstacleHp.set(id, e.n ?? 0);
      } else if (e.k === 'ObstacleBroken') {
        this.brokenObstacles.set(id, true);
        this.obstacleHp.set(id, 0);
      } else if (e.k === 'ObstacleRegrow' && e.s !== 'warn') {
        this.brokenObstacles.set(id, false);
        this.obstacleHp.delete(id);
      }
    }
  }

  me(): PlayerState | null {
    const v = this.players.get(this.myId);
    return v ? v.snaps[v.snaps.length - 1].s : null;
  }

  // ---- input & prediction ------------------------------------------------

  /** Records and returns the input message to send this network tick. */
  makeInput(angle: number, moving: boolean): Record<string, unknown> {
    const inp = { seq: ++this.seq, angle: quantize(angle), moving };
    this.pending.push(inp);
    const me = this.me();
    if (me && !(me.fl & 16)) {
      this.predicted = this.step(this.predicted, inp, me, 1 / this.cfg.net.tick_hz);
    }
    this.lastInput = inp;
    this.lastSendAt = performance.now();
    return { t: 'in', s: inp.seq, a: Math.round((inp.angle / (2 * Math.PI)) * 256) & 255, m: moving };
  }

  private reconcile(ack: number): void {
    const me = this.me();
    if (!me) {
      return;
    }
    this.pending = this.pending.filter((p) => p.seq > ack);
    let pos: Vec = { x: me.x, y: me.y };
    for (const inp of this.pending) {
      pos = this.step(pos, inp, me, 1 / this.cfg.net.tick_hz);
    }
    // Keep the displayed position continuous; the correction fades out in renderFrame.
    this.correction.x += this.predicted.x - pos.x;
    this.correction.y += this.predicted.y - pos.y;
    if (Math.hypot(this.correction.x, this.correction.y) > 400) {
      this.correction = { x: 0, y: 0 }; // teleport or respawn: snap
    }
    this.predicted = pos;
  }

  /** One movement step, mirroring game.movePlayer on the server. */
  private step(from: Vec, inp: PendingInput, me: PlayerState, dt: number): Vec {
    if (!inp.moving || me.fl & 8) {
      return from;
    }
    const dir = { x: Math.cos(inp.angle), y: Math.sin(inp.angle) };
    let speed = this.cfg.player.move_base * Math.sqrt(me.sc);
    if (me.bf & (1 << BUFFS.indexOf('haste'))) {
      speed *= this.cfg.buffs.haste.mult;
    }
    speed *= this.zoneMult(from, dir);
    return this.resolveBlocking({ x: from.x + dir.x * speed * dt, y: from.y + dir.y * speed * dt }, this.cfg.player.body_r * me.sc);
  }

  private zoneMult(p: Vec, dir: Vec): number {
    let mult = 1;
    for (const z of this.map?.zones ?? []) {
      if (!contains(z.s, p.x, p.y)) {
        continue;
      }
      if (z.k === 'mud') {
        mult *= this.cfg.zones.mud_mult;
      } else if (z.k === 'wind') {
        const wd = z.wd ?? 0;
        mult *= dir.x * Math.cos(wd) + dir.y * Math.sin(wd) >= 0 ? this.cfg.zones.wind_with : this.cfg.zones.wind_against;
      }
    }
    return mult;
  }

  private resolveBlocking(p: Vec, r: number): Vec {
    const map = this.map;
    if (!map) {
      return p;
    }
    for (let iter = 0; iter < 3; iter++) {
      let moved = false;
      for (const o of map.obstacles) {
        if (this.brokenObstacles.get(o.id)) {
          continue;
        }
        const np = pushOut(o.s, p, r);
        if (np) {
          p = np;
          moved = true;
        }
      }
      if (!moved) {
        break;
      }
    }
    return { x: Math.max(r, Math.min(map.size - r, p.x)), y: Math.max(r, Math.min(map.size - r, p.y)) };
  }

  // ---- per-frame queries -------------------------------------------------

  /** Interpolated (others) or predicted (me) state for rendering. */
  playerFrames(dt: number): PlayerFrame[] {
    const decay = Math.exp(-dt * 10);
    this.correction.x *= decay;
    this.correction.y *= decay;
    const rt = this.renderTime();
    const out: PlayerFrame[] = [];
    for (const v of this.players.values()) {
      out.push(v.id === this.myId ? this.myFrame(v) : interpPlayer(v, rt));
    }
    return out;
  }

  private myFrame(v: PlayerView): PlayerFrame {
    const last = v.snaps[v.snaps.length - 1];
    const s = last.s;
    // Advance within the current input step so motion is smooth at high frame rates.
    const since = Math.min((performance.now() - this.lastSendAt) / 1000, 1 / this.cfg.net.tick_hz);
    const ahead = this.step(this.predicted, this.lastInput, s, since);
    const ringAngle = s.ph + s.d * s.w * (this.serverNow() - last.t);
    return { view: v, s, x: ahead.x + this.correction.x, y: ahead.y + this.correction.y, ringAngle };
  }

  monsterFrames(): { s: MonsterState; x: number; y: number }[] {
    const rt = this.renderTime();
    const out: { s: MonsterState; x: number; y: number }[] = [];
    for (const buf of this.monsters.values()) {
      const { a, b, t } = bracket(buf, rt);
      out.push({ s: b.s, x: lerp(a.s.x, b.s.x, t), y: lerp(a.s.y, b.s.y, t) });
    }
    return out;
  }
}

function quantize(angle: number): number {
  const step = (2 * Math.PI) / 256;
  const a = ((angle % (2 * Math.PI)) + 2 * Math.PI) % (2 * Math.PI);
  return Math.round(a / step) * step;
}

/** bracket finds the two snapshots around time t and the blend factor between them. */
function bracket<T>(buf: Timed<T>[], t: number): { a: Timed<T>; b: Timed<T>; t: number } {
  for (let i = buf.length - 1; i > 0; i--) {
    if (buf[i - 1].t <= t) {
      const a = buf[i - 1];
      const b = buf[i];
      const span = b.t - a.t;
      return { a, b, t: span > 0 ? Math.min(1, (t - a.t) / span) : 1 };
    }
  }
  return { a: buf[0], b: buf[0], t: 1 };
}

function interpPlayer(v: PlayerView, rt: number): PlayerFrame {
  const { a, b, t } = bracket(v.snaps, rt);
  // Large jumps (teleport, respawn) are not interpolated.
  const jump = Math.hypot(b.s.x - a.s.x, b.s.y - a.s.y) > 600;
  const k = jump ? 1 : t;
  const base = t >= 1 ? b : a;
  const ringAngle = base.s.ph + base.s.d * base.s.w * (rt - base.t);
  return { view: v, s: b.s, x: lerp(a.s.x, b.s.x, k), y: lerp(a.s.y, b.s.y, k), ringAngle };
}
