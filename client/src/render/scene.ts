import { Application, Container } from 'pixi.js';

import { sound } from '../audio/sound';
import { cameraViewHeight, type GameConfig, realmOf } from '../config';
import type { ClientWorld } from '../net/state';
import { BUFFS, type GameEvent } from '../net/protocol';
import { pick, STR } from '../ui/strings';
import { ItemSprite, MonsterSprite } from './entities';
import { MapView, type Viewer } from './mapview';
import { PlayerSprite } from './player';
import { ringGeom } from './ringgeom';
import { Vfx } from './vfx';

const MOOD_EMOTE_CHANCE = 0.12; // chance per mood change that a character shouts (run/chase/flee)

/** GameScene renders the ClientWorld every frame with a camera that zooms out as you grow (spec §4.1). */
export class GameScene {
  readonly app = new Application();
  private readonly camera = new Container();
  private readonly entities = new Container();
  private readonly rings = new Container();
  private readonly vfx = new Vfx();
  private mapView: MapView | null = null;
  private readonly players = new Map<number, PlayerSprite>();
  private readonly monsters = new Map<number, MonsterSprite>();
  private readonly items = new Map<number, ItemSprite>();
  private readonly lastMood = new Map<number, string>();
  private zoom = 1;
  private viewH = 900;
  private world: ClientWorld | null = null;
  private lastSpark = { t: 0, x: 0, y: 0 };
  private showHitbox = false;

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
    this.camera.addChild(this.mapView.ground, this.mapView.decor, this.mapView.under, this.rings, this.entities, this.mapView.over, this.vfx.layer);
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
    const me = frames.find((f) => f.view.id === w.myId);
    if (me) {
      const long = me.s.bf & (1 << BUFFS.indexOf('long')) ? (this.cfg.buffs.long?.mult ?? 1) : 1;
      this.updateCamera(me.x, me.y, me.s.sc, ringGeom(this.cfg, me.s.sc, me.s.n, long).reach, now);
    }
    this.syncPlayers(frames, now);
    this.syncMonsters(now);
    this.syncItems(now);
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

  private syncItems(now: number): void {
    const items = this.world?.items;
    if (!items) {
      return;
    }
    for (const [id, it] of items) {
      if (!this.items.has(id)) {
        const sp = new ItemSprite(it, this.cfg.items.item_r);
        this.items.set(id, sp);
        this.entities.addChild(sp.root);
      }
    }
    for (const [id, sp] of this.items) {
      if (!items.has(id)) {
        sp.destroy();
        this.items.delete(id);
        continue;
      }
      sp.update(now);
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
      case 'SwordBreak':
        this.vfx.play('sword_shatter', e.x, e.y, size * 1.5, now);
        sound.play('sword_break', mine ? 1 : 0.5);
        break;
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
        if (mine) {
          sound.play('realm_drop');
        }
        break;
      default:
        this.handleWorldEvent(e, now, mine, size);
    }
  }

  private onBreakthrough(e: GameEvent, size: number, mine: boolean, now: number): void {
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
