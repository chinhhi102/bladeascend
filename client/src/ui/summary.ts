import type { GameConfig } from '../config';
import { realmOf, tierOf } from '../config';
import type { DeadMsg } from '../net/protocol';
import { assets } from '../render/assets';
import { button, h } from './dom';
import { pick, STR } from './strings';

export interface SummaryActions {
  playAgain: () => void;
  leaderboard: () => void;
  changeCharacter: () => void;
}

/** summaryScreen is the Tổng kết screen after death (spec §7). */
export function summaryScreen(cfg: GameConfig, d: DeadMsg, myChar: string, act: SummaryActions): HTMLElement {
  const root = h('div', 'screen summary');
  const panel = h('div', 'panel scroll-panel');
  if (assets.has('ui/scroll.png')) {
    panel.classList.add('has-scroll');
  }
  const ghost = assets.has(`characters/${myChar}/atlas.png`) ? h('img', 'ghost') : null;
  if (ghost) {
    ghost.src = `/assets/characters/${myChar}/frames/death_07.png`;
    ghost.onerror = () => ghost.remove();
  }
  const realm = cfg.cultivation.realms[realmOf(cfg, d.peak_level)] ?? '';
  const stat = (label: string, value: string): HTMLElement => h('div', 'stat', h('span', '', label), h('b', '', value));
  const killer = d.killer ? `${STR.killedBy(d.killer)} (${STR.characters[d.killer_char] ?? d.killer_char})` : STR.killedByUnknown;
  panel.append(
    h('h2', '', STR.summaryTitle),
    ...(ghost ? [ghost] : []),
    h('div', 'killer', killer),
    h('div', 'joke', pick(STR.deathJokes)),
    stat(STR.peakRealm, STR.realmTier(realm, tierOf(cfg, d.peak_level))),
    stat(STR.peakExp, Math.round(d.peak_exp).toLocaleString('vi-VN')),
    stat(STR.kills, String(d.kills)),
    stat(STR.survive, STR.secs(d.survive)),
    stat(STR.roomRank, `#${d.rank}`),
    h('div', 'row', button(STR.playAgain, 'btn primary', act.playAgain), button(STR.leaderboard, 'btn', act.leaderboard), button(STR.changeCharacter, 'btn', act.changeCharacter)),
  );
  root.append(panel);
  return root;
}
