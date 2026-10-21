import './ui/style.css';

import { sound } from './audio/sound';
import { type GameConfig, levelThresholds, loadConfig } from './config';
import { Input } from './input';
import { Conn } from './net/conn';
import type { DeadMsg, GameMap, ServerMsg, Welcome } from './net/protocol';
import { ClientWorld } from './net/state';
import { assets } from './render/assets';
import { GameScene } from './render/scene';
import { h, Screens } from './ui/dom';
import { guestId, homeScreen, type Profile } from './ui/home';
import { Hud } from './ui/hud';
import { leaderboardScreen } from './ui/leaderboard';
import { STR } from './ui/strings';
import { summaryScreen } from './ui/summary';

const DEBUG_KEYS: Record<string, Record<string, unknown>> = {
  KeyE: { cmd: 'exp', n: 100 },
  KeyR: { cmd: 'realm', n: 0 },
  KeyK: { cmd: 'sword', n: 1 },
  KeyX: { cmd: 'die', n: 0 },
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
  private playing = false;
  private debug = false;

  constructor(private readonly cfg: GameConfig, private readonly overlay: HTMLElement, stage: HTMLElement) {
    this.screens = new Screens(overlay);
    this.scene = new GameScene(cfg);
    this.world = new ClientWorld(cfg);
    this.hud = new Hud(cfg);
    this.input = new Input(stage);
    this.conn = new Conn((m) => this.onMessage(m));
    this.thresholds = levelThresholds(cfg);
  }

  async start(stage: HTMLElement): Promise<void> {
    await this.scene.init(stage, fpsCap());
    this.scene.app.ticker.add((t) => this.frame(t.deltaMS / 1000));
    window.setInterval(() => this.sendInput(), 1000 / this.cfg.net.tick_hz);
    window.addEventListener('keydown', (e) => this.debugKey(e));
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
    this.showHome();
  }

  private showHome(): void {
    this.screens.show(homeScreen(this.cfg, (p) => this.join(p), () => this.showLeaderboard(() => this.showHome())));
  }

  private showLeaderboard(back: () => void): void {
    this.screens.show(leaderboardScreen(this.cfg, back));
  }

  private join(p: Profile): void {
    this.profile = p;
    sound.startMusic();
    this.screens.show(h('div', 'screen center', h('div', 'panel', this.conn.isOpen ? STR.findingRoom : STR.connecting)));
    const send = (): void => {
      if (!this.conn.isOpen) {
        window.setTimeout(send, 300);
        return;
      }
      this.conn.send({ t: 'join', name: p.name, char: p.char, guest_id: guestId() });
    };
    send();
  }

  private onMessage(m: ServerMsg): void {
    switch (m.t) {
      case 'welcome':
        void this.onWelcome(m);
        break;
      case 'snap': {
        if (!this.playing) {
          return;
        }
        const events = this.world.applySnapshot(m);
        this.scene.handleEvents(events);
        this.hud.onEvents(events, this.world.myId);
        break;
      }
      case 'lb':
        this.hud.setLeaderboard(m, this.world.myId);
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
      default:
    }
  }

  private async onWelcome(w: Welcome): Promise<void> {
    const res = await fetch(`/api/map/${w.map}`);
    const map = (await res.json()) as GameMap;
    this.world.reset(w, map);
    this.scene.setWorld(this.world);
    this.debug = w.debug;
    this.hud.setDebug(w.debug);
    this.screens.show(null);
    this.overlay.appendChild(this.hud.root);
    this.playing = true;
    this.input.enabled = true;
    sound.play('respawn');
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

  private sendInput(): void {
    if (!this.playing) {
      return;
    }
    const { angle, moving } = this.input.read();
    this.conn.send(this.world.makeInput(angle, moving));
  }

  private debugKey(e: KeyboardEvent): void {
    if (!this.playing || !this.debug || e.repeat) {
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
    this.hud.update(this.world.me(), this.conn.rttMs, (L) => ({ from: this.thresholds[L], to: this.thresholds[L + 1] }));
  }
}

async function boot(): Promise<void> {
  const overlay = document.getElementById('overlay') as HTMLElement;
  const stage = document.getElementById('stage') as HTMLElement;
  const loading = h('div', 'screen center', h('div', 'panel', h('div', 'logo', STR.title), h('div', 'muted', STR.connecting)));
  overlay.appendChild(loading);
  try {
    const cfg = await loadConfig();
    await assets.init();
    await assets.preload();
    sound.init();
    loading.remove();
    await new App(cfg, overlay, stage).start(stage);
  } catch (err) {
    console.error(err);
    loading.replaceChildren(h('div', 'panel', `${STR.disconnected} (${String(err)})`));
    window.setTimeout(() => location.reload(), 3000);
  }
}

void boot();
