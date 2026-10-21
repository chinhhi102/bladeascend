import type { GameConfig } from '../config';
import { realmOf, tierOf } from '../config';
import { assets } from '../render/assets';
import { button, h } from './dom';
import { STR } from './strings';

interface Row {
  name: string;
  character: string;
  peak_exp: number;
  peak_level: number;
  kills: number;
}

const MEDALS = ['gold', 'silver', 'bronze'];

/** leaderboardScreen: tabs Hôm nay / Tuần này / Mọi thời đại × Tu Vi Đỉnh Cao / Sát Thần (spec §11). */
export function leaderboardScreen(cfg: GameConfig, onBack: () => void): HTMLElement {
  const root = h('div', 'screen board');
  const panel = h('div', 'panel wide');
  let period = 'day';
  let metric = 'exp';
  const table = h('div', 'lb-table');
  const periods = h('div', 'tabs');
  const metrics = h('div', 'tabs');
  const tab = (group: HTMLElement, label: string, value: string, set: (v: string) => void, active: boolean): void => {
    const b = button(label, `tab${active ? ' active' : ''}`, () => {
      group.querySelectorAll('.tab').forEach((t) => t.classList.remove('active'));
      b.classList.add('active');
      set(value);
      void load();
    });
    group.append(b);
  };
  const load = async (): Promise<void> => {
    table.replaceChildren(h('div', 'muted', STR.lbLoading));
    try {
      const res = await fetch(`/api/leaderboard?period=${period}&metric=${metric}`);
      const data = (await res.json()) as { rows: Row[] };
      renderRows(cfg, table, data.rows, metric);
    } catch {
      table.replaceChildren(h('div', 'error', STR.lbError));
    }
  };
  tab(periods, STR.lbToday, 'day', (v) => (period = v), true);
  tab(periods, STR.lbWeek, 'week', (v) => (period = v), false);
  tab(periods, STR.lbAll, 'all', (v) => (period = v), false);
  tab(metrics, STR.lbExp, 'exp', (v) => (metric = v), true);
  tab(metrics, STR.lbKills, 'kills', (v) => (metric = v), false);
  panel.append(h('h2', '', STR.lbTitle), periods, metrics, table, button(STR.back, 'btn', onBack));
  root.append(panel);
  void load();
  return root;
}

function renderRows(cfg: GameConfig, table: HTMLElement, rows: Row[], metric: string): void {
  if (!rows.length) {
    table.replaceChildren(h('div', 'muted', STR.lbEmpty));
    return;
  }
  const list = rows.map((r, i) => {
    const rank = h('span', 'rank', String(i + 1));
    if (i < 3 && assets.has(`ui/medal_${MEDALS[i]}.png`)) {
      const img = h('img', 'medal');
      img.src = `/assets/ui/medal_${MEDALS[i]}.png`;
      rank.replaceChildren(img);
    }
    const realm = cfg.cultivation.realms[realmOf(cfg, r.peak_level)] ?? '';
    const value = metric === 'kills' ? String(r.kills) : Math.round(r.peak_exp).toLocaleString('vi-VN');
    return h('div', 'lb-row', rank, h('span', 'nm', r.name), h('span', 'realm', STR.realmTier(realm, tierOf(cfg, r.peak_level))), h('span', 'val', value));
  });
  table.replaceChildren(...list);
}
