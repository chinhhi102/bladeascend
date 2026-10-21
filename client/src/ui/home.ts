import type { GameConfig } from '../config';
import { assets } from '../render/assets';
import { charColor } from '../render/palette';
import { button, h, portrait } from './dom';
import { STR } from './strings';

const LS_NAME = 'kx_name';
const LS_CHAR = 'kx_char';
const LS_GUEST = 'kx_guest_id';

function lsGet(k: string): string | null {
  try {
    return localStorage.getItem(k);
  } catch {
    return null;
  }
}

function lsSet(k: string, v: string): void {
  try {
    localStorage.setItem(k, v);
  } catch {
    // private mode: remembering is only a convenience
  }
}

/** guestId is the anonymous leaderboard identity kept in localStorage (spec §11). */
export function guestId(): string {
  let id = lsGet(LS_GUEST);
  if (!id) {
    id = crypto.randomUUID ? crypto.randomUUID() : `${Date.now()}-${Math.random().toString(16).slice(2)}`;
    lsSet(LS_GUEST, id);
  }
  return id;
}

export interface Profile {
  name: string;
  char: string;
}

export function savedProfile(cfg: GameConfig): Profile {
  const char = lsGet(LS_CHAR) ?? '';
  return { name: lsGet(LS_NAME) ?? '', char: cfg.characters.includes(char) ? char : cfg.characters[0] };
}

function toHex(n: number): string {
  return `#${n.toString(16).padStart(6, '0')}`;
}

/** homeScreen: name input + character picker + play / leaderboard buttons (spec §2). */
export function homeScreen(cfg: GameConfig, onPlay: (p: Profile) => void, onLeaderboard: () => void): HTMLElement {
  const profile = savedProfile(cfg);
  const root = h('div', 'screen home');
  if (assets.has('ui/home_bg.png')) {
    root.style.backgroundImage = 'url(/assets/ui/home_bg.png)';
  }
  const panel = h('div', 'panel');
  const input = h('input', 'name-input');
  input.placeholder = STR.namePlaceholder;
  input.maxLength = cfg.player.name_max;
  input.value = profile.name;
  const err = h('div', 'error');
  const grid = h('div', 'char-grid');
  let selected = profile.char;
  const cards = new Map<string, HTMLElement>();
  for (const c of cfg.characters) {
    const card = h('button', 'char-card', portrait(c, (p) => assets.has(p), toHex(charColor(c)[0])), h('span', '', STR.characters[c] ?? c));
    card.addEventListener('click', () => {
      selected = c;
      cards.forEach((el, id) => el.classList.toggle('selected', id === c));
    });
    card.classList.toggle('selected', c === selected);
    cards.set(c, card);
    grid.append(card);
  }
  const play = (): void => {
    const name = input.value.trim();
    if ([...name].length < cfg.player.name_min || [...name].length > cfg.player.name_max) {
      err.textContent = STR.badName;
      return;
    }
    lsSet(LS_NAME, name);
    lsSet(LS_CHAR, selected);
    onPlay({ name, char: selected });
  };
  input.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') {
      play();
    }
  });
  panel.append(
    h('h1', 'logo', STR.title),
    h('div', 'subtitle', STR.subtitle),
    input,
    err,
    h('div', 'label', STR.pickCharacter),
    grid,
    h('div', 'row', button(STR.play, 'btn primary', play), button(STR.leaderboard, 'btn', onLeaderboard)),
    h('div', 'hint', 'ontouchstart' in window ? STR.controlsMobile : STR.controlsPc),
  );
  root.append(panel);
  return root;
}
