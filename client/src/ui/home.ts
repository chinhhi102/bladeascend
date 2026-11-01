import type { GameConfig } from '../config';
import { mountButton } from './feedback';
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
/** cooldownText: "Hồi 18s · tích 2 lượt" từ config (spec v6.25 §6.3). */
function cooldownText(cfg: GameConfig, char: string): string {
  const cd = cfg.skills.cooldowns[char];
  return cd ? STR.skillCooldown(cd[0], cd[1]) : '';
}

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

/**
 * homeScreen: name input + character picker + play / leaderboard buttons (spec §2).
 *
 * Hiện NGAY, không chờ asset nào (GAME_SPEC §13.7 mục 4 + chủ dự án chốt 31/10). Ảnh chân
 * dung và nền là thẻ `<img>` / `background-image` của DOM, nên **trình duyệt tự tải lấy** —
 * chỉ cần manifest để biết địa chỉ, không cần đi qua bộ nạp Pixi. Vì vậy trang chủ chỉ phụ
 * thuộc hai lời gọi nhỏ: `/api/config` và `manifest.json`.
 *
 * `artReady` (không bắt buộc) là lời hứa của nhóm art nền; xong thì vá lại ảnh **tại chỗ**.
 * Vá tại chỗ chứ không dựng lại cả màn hình: người chơi đang gõ tên, dựng lại là xoá mất
 * chữ họ vừa gõ.
 */
export function homeScreen(cfg: GameConfig, onPlay: (p: Profile) => void, onLeaderboard: () => void, artReady?: Promise<void>): HTMLElement {
  const profile = savedProfile(cfg);
  const root = h('div', 'screen home');
  const applyBg = (): void => {
    const bg = assets.url('ui/home_bg.png');
    if (bg) {
      root.style.backgroundImage = `url(${bg})`;
      // Cùng ảnh làm nền CẢ TRANG (lớp cố định sau `#below`) — cuộn xuống phần giới thiệu vẫn
      // thấy phố chợ hai bên, không hụt sang màu giấy trơn.
      document.documentElement.style.setProperty('--home-bg', `url(${bg})`);
    }
  };
  applyBg();
  // Nút 💬 ở trang chủ (ADMIN_SPEC §7.1). Đặt ở đây chứ không ở `main.ts` để nó sống và
  // chết cùng màn hình chủ — gắn vào body thì nó còn lại trên màn hình trong trận.
  mountButton(root, 'home');
  const panel = h('div', 'panel');
  const input = h('input', 'name-input');
  input.placeholder = STR.namePlaceholder;
  input.maxLength = cfg.player.name_max;
  input.value = profile.name;
  const err = h('div', 'error');
  const grid = h('div', 'char-grid');
  let selected = profile.char;
  const cards = new Map<string, HTMLElement>();

  // Khung xem trước: ảnh to + tên + KỸ NĂNG của nhân vật đang chọn.
  //
  // Mười nhân vật chỉ khác nhau ở kỹ năng (tốc độ, máu, kiếm đều như nhau — spec §6.3), nên
  // lưới chỉ có ảnh nhỏ + tên thì người chơi chọn mù hoàn toàn. Đưa kỹ năng ra đây là chỗ
  // DUY NHẤT họ đọc được trước khi vào trận.
  const previewArt = h('div', 'preview-art');
  // Chữ của CẢ MƯỜI nhân vật nằm chồng lên nhau trong cùng một ô lưới, chỉ hiện cái đang
  // chọn. Ô lưới cao bằng khối DÀI NHẤT, nên khung xem trước có MỘT chiều cao cố định —
  // rê chuột qua các thẻ không làm nút "Vào chơi" bên dưới nhảy lên nhảy xuống (chủ dự án
  // báo 1/11: mô tả Sư Phụ 4 dòng, Bé Tí 2 dòng). Đo theo chữ thật nên đúng ở mọi cỡ màn hình.
  const blocks = new Map<string, HTMLElement>();
  for (const c of cfg.characters) {
    blocks.set(c, h('div', 'preview-block',
      h('div', 'preview-name', STR.characters[c] ?? c),
      h('div', 'preview-skill', `${STR.skillLabel}: ${STR.skillNames[c] ?? '—'}`),
      h('div', 'preview-cd', cooldownText(cfg, c)),
      h('div', 'preview-desc', STR.skillDesc[c] ?? '')));
  }
  const preview = h('div', 'char-preview',
    previewArt,
    h('div', 'preview-text', ...blocks.values()));

  const showPreview = (c: string): void => {
    previewArt.replaceChildren(portrait(c, (p) => assets.url(p), toHex(charColor(c)[0])));
    for (const [id, el] of blocks) {
      el.classList.toggle('on', id === c);
    }
  };

  for (const c of cfg.characters) {
    // Thẻ dùng tên NGẮN (một dòng); tên đầy đủ nằm ở khung chi tiết bên dưới. `title` giữ
    // tên đầy đủ để rê chuột trên PC vẫn đọc được ngay tại thẻ.
    const card = h('button', 'char-card', portrait(c, (p) => assets.url(p), toHex(charColor(c)[0])),
      h('span', 'card-name', STR.charactersShort[c] ?? STR.characters[c] ?? c),
      h('span', 'card-cd', cooldownText(cfg, c)));
    card.title = STR.characters[c] ?? c;
    const pick = (): void => {
      selected = c;
      cards.forEach((el, id) => el.classList.toggle('selected', id === c));
      showPreview(c);
    };
    card.addEventListener('click', pick);
    // Rê chuột cũng xem được: trên PC người ta muốn so kỹ năng của mấy nhân vật rồi mới
    // chốt, bắt bấm từng cái thì mỗi lần so là một lần đổi lựa chọn đang giữ.
    card.addEventListener('mouseenter', () => showPreview(c));
    card.addEventListener('mouseleave', () => showPreview(selected));
    card.classList.toggle('selected', c === selected);
    cards.set(c, card);
    grid.append(card);
  }
  showPreview(selected);
  // Manifest về muộn (hoặc chưa về lúc dựng màn hình) thì vá ảnh tại chỗ khi có.
  const applyArt = (): void => {
    applyBg();
    for (const [c, card] of cards) {
      const old = card.querySelector('.portrait');
      if (old) {
        old.replaceWith(portrait(c, (p) => assets.url(p), toHex(charColor(c)[0])));
      }
    }
    showPreview(selected);
  };
  if (artReady) {
    void artReady.then(applyArt, () => undefined);
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
  // Tên game: mỗi cụm một `<span>` không-ngắt, nên chỗ ngắt duy nhất là khoảng giữa hai
  // cụm — xem `STR.titleParts`. Khoảng trắng giữa hai span phải là NÚT VĂN BẢN THẬT, không
  // phải lề CSS: có thế thì màn rộng mới nối liền thành một dòng đọc đúng như tên gốc.
  const logo = h('h1', 'logo');
  STR.titleParts.forEach((part, i) => {
    if (i > 0) {
      logo.append(document.createTextNode(' '));
    }
    logo.append(h('span', 'logo-part', part));
  });
  panel.append(
    logo,
    h('div', 'subtitle', STR.subtitle),
    input,
    err,
    h('div', 'label', STR.pickCharacter),
    grid,
    preview,
    h('div', 'row', button(STR.play, 'btn primary', play), button(STR.leaderboard, 'btn', onLeaderboard)),
    h('div', 'hint', 'ontouchstart' in window ? STR.controlsMobile : STR.controlsPc),
  );
  root.append(panel);
  return root;
}
