import { Application, Container } from 'pixi.js';

import { sound } from '../audio/sound';
import { absorbable, banksTienKhi, cameraViewHeight, type GameConfig, realmOf, tienIndex } from '../config';
import type { ClientWorld } from '../net/state';
import type { ItemState } from '../net/protocol';
import { BUFFS, FLAG, type GameEvent } from '../net/protocol';
import { pick, STR } from '../ui/strings';
import { Beast, type BeastMode, TIEN_BEASTS } from './beast';
import { ItemSprite, MonsterSprite } from './entities';
import { MapView, type Viewer } from './mapview';
import { OrbLayer } from './orblayer';
import { PlayerSprite } from './player';
import { ringGeom } from './ringgeom';
import { Vfx } from './vfx';

const SUCK_SECS = 0.22; // vật phẩm nhặt được bay hút vào người trong 0.22s
const BEAST_MAX = 6; // spec §3.5: tối đa 6 thần thú cùng lúc trên màn hình
const BEAST_EVERY = 20; // spec §3.5: bản thường trực mỗi 20s
const AMBIENT_KEY = 'kx.beastAmbient'; // tắt/bật bản thường trực trong cài đặt đồ họa
const MOOD_EMOTE_CHANCE = 0.12; // chance per mood change that a character shouts (run/chase/flee)

/** GameScene renders the ClientWorld every frame with a camera that zooms out as you grow (spec §4.1). */
export class GameScene {
  readonly app = new Application();
  private readonly camera = new Container();
  private readonly entities = new Container();
  /** Viên Linh khí vẽ riêng bằng particle: vài trăm viên cùng lúc, xem orblayer.ts. */
  private readonly orbs = new OrbLayer();
  private readonly rings = new Container();
  private readonly vfx = new Vfx();
  private mapView: MapView | null = null;
  private readonly players = new Map<number, PlayerSprite>();
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
  private world: ClientWorld | null = null;
  private lastFrames: ReturnType<ClientWorld['playerFrames']> = [];
  private lastSpark = { t: 0, x: 0, y: 0 };
  private showHitbox = false;
  // Thần thú Tiên giới (spec §3.5): mỗi con có 2 lớp, một dưới nhân vật một trên.
  private beasts: { b: Beast; owner: number }[] = [];
  private nextAmbient = new Map<number, number>();
  /** Cảnh giới Tiên của từng người ở khung hình trước — RealmDrop cần biết họ VỪA là gì. */
  private lastTien = new Map<number, number>();

  constructor(private readonly cfg: GameConfig) {}

  /** viewSize is the world-space rectangle currently on screen (used by the minimap, §12.1). */
  viewSize(): { w: number; h: number } {
    return { w: this.viewH * (this.app.screen.width / Math.max(this.app.screen.height, 1)), h: this.viewH };
  }

  async init(parent: HTMLElement, maxFps: number): Promise<void> {
    await this.app.init({
      resizeTo: window,
      background: 0x2b2b2b,
      antialias: true,
      resolution: Math.min(window.devicePixelRatio || 1, 2),
      autoDensity: true,
    });
    this.app.ticker.maxFPS = maxFps;
    parent.appendChild(this.app.canvas);
    this.entities.sortableChildren = true;
    this.rings.sortableChildren = true;
    this.app.stage.addChild(this.camera);
  }

  setWorld(w: ClientWorld): void {
    this.world = w;
    this.clear();
    if (!w.map) {
      return;
    }
    this.mapView = new MapView(w.map, this.cfg.gates.radius);
    for (const node of this.mapView.obstacleLayer()) {
      this.entities.addChild(node);
    }
    // Rings sit under every character so a blade never covers a body.
    // Viên EXP luôn nằm dưới mọi thứ khác nên xếp ngay trên mặt đất, không cần y-sort.
    this.camera.addChild(this.mapView.ground, this.mapView.decor, this.mapView.under, this.orbs.root, this.rings, this.entities, this.mapView.over, this.vfx.layer);
  }

  clear(): void {
    this.camera.removeChildren();
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
    this.nextAmbient.clear();
    this.lastTien.clear();
    this.flying.length = 0;
    this.pickedBy.clear();
    this.orbs.clear();
    this.lastOrb.clear();
    this.players.clear();
    this.monsters.clear();
    this.items.clear();
    this.entities.removeChildren();
    this.rings.removeChildren();
  }

  /** render one frame; dt in seconds. */
  render(dt: number): void {
    const w = this.world;
    if (!w || !this.mapView) {
      return;
    }
    const now = performance.now() / 1000;
    const frames = w.playerFrames(dt);
    this.lastFrames = frames;
    const me = frames.find((f) => f.view.id === w.myId);
    if (me) {
      const long = me.s.bf & (1 << BUFFS.indexOf('long')) ? (this.cfg.buffs.long?.mult ?? 1) : 1;
      this.updateCamera(me.x, me.y, me.s.sc, ringGeom(this.cfg, me.s.sc, me.s.n, long).reach, now);
    }
    this.syncPlayers(frames, now);
    this.syncMonsters(now);
    this.syncItems(now, me ? banksTienKhi(this.cfg, me.s.L) : false);
    this.updateFlyingItems(now);
    this.updateBeasts(frames, now);
    const myBush = me && me.s.fl & 2 ? this.bushAt(me.x, me.y) : -1;
    this.mapView.update(w, now, myBush, me ? this.viewer(me.x, me.y, me.s.sc) : null, this.cfg.obstacles.occlusion_alpha);
    this.vfx.update(now);
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

  private updateCamera(x: number, y: number, scale: number, reach: number, now: number): void {
    // Smoothly follow the camera height so growing does not pop.
    const targetH = cameraViewHeight(this.cfg, scale, reach);
    this.viewH += (targetH - this.viewH) * 0.08;
    this.zoom = this.app.screen.height / this.viewH;
    const shake = this.vfx.shakeOffset(now);
    this.camera.scale.set(this.zoom);
    this.camera.position.set(this.app.screen.width / 2 - (x + shake.x) * this.zoom, this.app.screen.height / 2 - (y + shake.y) * this.zoom);
  }

  private syncPlayers(frames: ReturnType<ClientWorld['playerFrames']>, now: number): void {
    const seen = new Set<number>();
    for (const f of frames) {
      seen.add(f.view.id);
      let sp = this.players.get(f.view.id);
      if (!sp || sp.char !== f.view.char) {
        sp?.destroy();
        sp = new PlayerSprite(this.cfg, f.view.id, f.view.name, f.view.char, f.view.id === this.world?.myId);
        this.players.set(f.view.id, sp);
        this.entities.addChild(sp.root);
        this.rings.addChild(sp.ring);
      }
      sp.update(f, this.zoom, now, this.showHitbox);
      this.moodChatter(f.view.id, f.s.md, now);
    }
    for (const [id, sp] of this.players) {
      if (!seen.has(id)) {
        sp.destroy();
        this.players.delete(id);
      }
    }
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
      sp.update(now, absorbable(sp.it.k, meBanksTienKhi));
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

  /** beastCount: số thần thú đang chạy, theo kiểu — dùng để kiểm tra bằng console debug. */
  beastCount(): Record<string, number> {
    const n: Record<string, number> = { breakthrough: 0, ambient: 0, shatter: 0 };
    for (const e of this.beasts) {
      n[e.b.mode] += 1;
    }
    return n;
  }

  /** ambientBeasts: bản thường trực bật/tắt được trong cài đặt đồ họa (spec §3.5). */
  static ambientBeasts(): boolean {
    return localStorage.getItem(AMBIENT_KEY) !== 'off';
  }

  static setAmbientBeasts(on: boolean): void {
    localStorage.setItem(AMBIENT_KEY, on ? 'on' : 'off');
  }

  /**
   * updateBeasts chạy toàn bộ thần thú Tiên giới mỗi khung hình: hẹn giờ bản thường trực,
   * đặt lại vị trí theo chủ, và xếp hai lớp ôm lấy nhân vật đúng thứ tự chiều sâu.
   */
  private updateBeasts(frames: ReturnType<ClientWorld['playerFrames']>, now: number): void {
    const byId = new Map(frames.map((f) => [f.view.id, f]));
    this.lastTien.clear();
    for (const f of frames) {
      const idx = tienIndex(this.cfg, realmOf(this.cfg, f.s.L));
      this.lastTien.set(f.view.id, idx);
      // Ẩn Thân / trong bụi thì không phát (spec §3.5).
      const hidden = (f.s.fl & FLAG.stealth) !== 0 || (f.s.fl & FLAG.bush) !== 0;
      if (idx < 0 || hidden || !GameScene.ambientBeasts()) {
        this.nextAmbient.delete(f.view.id);
        continue;
      }
      const due = this.nextAmbient.get(f.view.id);
      if (due === undefined) {
        this.nextAmbient.set(f.view.id, now + BEAST_EVERY * Math.random()); // rải đều, không đồng loạt
      } else if (now >= due) {
        this.nextAmbient.set(f.view.id, now + BEAST_EVERY);
        this.spawnBeast(f.view.id, 'ambient', now);
      }
    }
    this.beasts = this.beasts.filter((e) => {
      const f = byId.get(e.owner);
      if (!f || e.b.done(now)) {
        this.removeBeast(e.b);
        return false;
      }
      const r = this.cfg.player.body_r * f.s.sc;
      const feet = f.y + r * 0.75; // chân nhân vật, khớp FEET_BELOW_CENTER của PlayerSprite
      e.b.update(f.x, feet, now);
      // Lớp sau nằm ngay dưới nhân vật, lớp trước ngay trên — nhân vật zIndex = y + r×0.75.
      e.b.back.zIndex = feet - 0.5;
      e.b.front.zIndex = feet + 0.5;
      return true;
    });
  }

  /** spawnBeast tạo thần thú cho một người, tôn trọng trần 6 con của spec §3.5. */
  private spawnBeast(owner: number, mode: BeastMode, now: number, realmIdx?: number): void {
    const f = this.lastFrames.find((p) => p.view.id === owner);
    if (!f) {
      return;
    }
    const idx = realmIdx ?? tienIndex(this.cfg, realmOf(this.cfg, f.s.L));
    const spec = TIEN_BEASTS[idx];
    if (!spec) {
      return;
    }
    if (this.beasts.length >= BEAST_MAX && !this.dropOneBeast(mode)) {
      return; // đầy mà không có con nào ưu tiên thấp hơn: bỏ qua
    }
    const r = this.cfg.player.body_r * f.s.sc;
    const b = new Beast(spec, mode, r, r * 2.6, now);
    this.entities.addChild(b.back, b.front);
    this.beasts.push({ b, owner });
  }

  /** dropOneBeast nhường chỗ: bản đột phá được ưu tiên, bản thường trực bị bỏ trước. */
  private dropOneBeast(incoming: BeastMode): boolean {
    const i = this.beasts.findIndex((e) => e.b.mode === 'ambient');
    if (i < 0 || incoming === 'ambient') {
      return false;
    }
    this.removeBeast(this.beasts[i].b);
    this.beasts.splice(i, 1);
    return true;
  }

  private removeBeast(b: Beast): void {
    this.entities.removeChild(b.back, b.front);
    b.destroy();
  }

  private handleEvent(e: GameEvent, now: number): void {
    const mine = e.a === this.world?.myId;
    const size = 90 * this.playerScale(e.a);
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

  private onBreakthrough(e: GameEvent, size: number, mine: boolean, now: number): void {
    if (e.s === 'tien') {
      // Phi thăng Tiên giới: thần thú to chui lên từ chân, ai trong tầm nhìn đều thấy (§3.5).
      // Lấy cảnh giới từ e.n (cấp MỚI) chứ không từ khung hình đã lưu — khung hình còn là
      // cảnh giới cũ nên sẽ gọi nhầm thần thú của cảnh giới vừa rời đi.
      this.spawnBeast(e.a ?? 0, 'breakthrough', now, tienIndex(this.cfg, realmOf(this.cfg, e.n ?? 0)));
      this.vfx.shake(22 * this.playerScale(e.a), 0.8, now);
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
      case 'BuffEnd':
        if (mine) {
          sound.play('buff_end');
        }
        break;
      case 'MonsterHurt':
        sound.play('monster_hurt', 0.4);
        break;
      case 'MonsterDie':
        this.vfx.play('death_smoke', e.x, e.y, 120, now);
        sound.play('monster_pop', 0.6);
        break;
      case 'ObstacleDamaged':
        sound.play('obstacle_hit', 0.4);
        break;
      case 'ObstacleBroken':
        this.vfx.play('death_smoke', e.x, e.y, 260, now);
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

  /** realm of the local player for HUD. */
  myRealm(): number {
    const me = this.world?.me();
    return me ? realmOf(this.cfg, me.L) : 0;
  }

  get currentZoom(): number {
    return this.zoom;
  }
}
