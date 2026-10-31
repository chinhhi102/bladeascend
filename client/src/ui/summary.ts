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
  // Khung giấy cuộn 9-slice. Địa chỉ ảnh đặt bằng JS chứ không nằm trong `style.css`: tệp
  // trong `assets/dist` mang hash nội dung, nên CSS tĩnh không thể biết tên. Trước đây CSS
  // trỏ `/assets/ui/scroll.png` và vite còn cảnh báo "didn't resolve at build time" —
  // chạy được là vì server phục vụ thẳng thư mục nguồn, và sẽ 404 ngay khi hết làm vậy.
  const scroll = assets.url('ui/scroll.png');
  if (scroll) {
    panel.classList.add('has-scroll');
    // Cắt theo PHẦN TRĂM, không theo pixel — và đây là bài học đắt của ngày 31/10.
    //
    // Khung giấy cuộn là ảnh 9 ô: `border-image-slice` nói cắt ở đâu. Bản cũ ghi
    // `205 199 213 160`, tức **pixel của ảnh gốc 1024×768**. Rồi `tools/asset-pipeline`
    // thu `ui/scroll.png` về 256×192 theo luật chung của nhóm `ui` — và 160 + 199 = 359 đã
    // rộng hơn cả tấm ảnh 256. Trình duyệt bỏ phép cắt, mỗi góc khung lĩnh trọn cả tấm
    // cuộn giấy tí hon, giữa khung trống trơn (ảnh chụp màn Tổng kết).
    //
    // Phần trăm thì không phụ thuộc ảnh to hay nhỏ: 205/768 = 26.7% trên, 199/1024 = 19.4%
    // phải, 213/768 = 27.7% dưới, 160/1024 = 15.6% trái. Đổi kích thước ảnh bao nhiêu lần
    // nữa cũng không hỏng lại.
    panel.style.borderImageSource = `url(${scroll})`;
    panel.style.borderImageSlice = '26.7% 19.4% 27.7% 15.6% fill';
    panel.style.borderImageRepeat = 'stretch';
  }
  // Hồn ma dùng CHÂN DUNG, không dùng `frames/death_07.png`.
  //
  // `frames/` là ảnh NGUỒN để ghép atlas — nó không có trong `manifest.json` và không được
  // deploy (hơn 100 MB mà game không đọc tới). Trước đây đường dẫn này chạy được chỉ vì
  // server phục vụ thẳng thư mục làm việc; từ lúc có `tools/deploy.sh` thì nó 404 và màn
  // Tổng kết mất hình. Chân dung khóc hợp cảnh hơn và luôn có mặt.
  const ghostFile = assets.has(`characters/${myChar}/portrait_cry.png`) ? 'portrait_cry.png'
    : assets.has(`characters/${myChar}/portrait_neutral.png`) ? 'portrait_neutral.png'
    : '';
  const ghost = ghostFile ? h('img', 'ghost') : null;
  if (ghost) {
    ghost.src = assets.url(`characters/${myChar}/${ghostFile}`);
    ghost.onerror = () => ghost.remove();
  }
  const realm = cfg.cultivation.realms[realmOf(cfg, d.peak_level)] ?? '';
  const stat = (label: string, value: string): HTMLElement => h('div', 'stat', h('span', '', label), h('b', '', value));
  const killer = d.killer ? `${STR.killedBy(d.killer)} (${STR.characters[d.killer_char] ?? d.killer_char})` : STR.killedByUnknown;
  const num = (n: number): string => Math.round(n).toLocaleString('vi-VN');
  panel.append(
    h('h2', '', STR.summaryTitle),
    ...(ghost ? [ghost] : []),
    h('div', 'killer', killer),
    h('div', 'joke', pick(STR.deathJokes)),
    h('div', 'score-total', h('span', '', STR.score), h('b', '', num(d.score ?? 0))),
    ...scoreBreakdown(d, num),
    stat(STR.peakRealm, STR.realmTier(realm, tierOf(cfg, d.peak_level))),
    stat(STR.peakExp, num(d.peak_exp)),
    stat(STR.hits, String(d.hits ?? 0)),
    ...(d.burned_exp > 0 ? [stat(STR.burnSpent, num(d.burned_exp))] : []),
    stat(STR.kills, String(d.kills)),
    stat(STR.toDaiThua, d.dai_thua_secs > 0 ? STR.secs(d.dai_thua_secs) : STR.notReached),
    stat(STR.top1Held, STR.secs(d.top1_secs ?? 0)),
    stat(STR.survive, STR.secs(d.survive)),
    stat(STR.roomRank, `#${d.rank}`),
    h('div', 'row', button(STR.playAgain, 'btn primary', act.playAgain), button(STR.leaderboard, 'btn', act.leaderboard), button(STR.changeCharacter, 'btn', act.changeCharacter)),
  );
  root.append(panel);
  return root;
}

/** scoreBreakdown lists Uy Danh per source, biggest first (spec §11.1). */
function scoreBreakdown(d: DeadMsg, num: (n: number) => string): HTMLElement[] {
  const by = Object.entries(d.score_by ?? {}).filter(([, v]) => v >= 1);
  if (!by.length) {
    return [];
  }
  by.sort((a, b) => b[1] - a[1]);
  return [h('div', 'score-by', ...by.map(([k, v]) => h('div', 'score-line', h('span', '', STR.scoreSources[k] ?? k), h('b', '', `+${num(v)}`))))];
}
