import type { GameConfig } from '../config';
import { mountButton } from './feedback';
import { assets } from '../render/assets';
import { charColor } from '../render/palette';
import { button, h, portrait } from './dom';
import { LANG } from '../i18n';
import { langPicker } from './langpick';
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

const STAT_KEYS = ['speed', 'pickup', 'spin', 'durability'] as const;

/**
 * statBars: 4 thanh nội tại (Tốc · Nhặt · Quay · Bền) của nhân vật — spec v6.33 §6.4. Thanh
 * mọc từ giữa: dương sang phải (xanh), âm sang trái (đỏ). Nửa thanh = mức lệch lớn nhất có
 * trong config, để thanh dài nhất luôn chạm mép mà không phải ghi cứng "15%".
 */
function statBars(cfg: GameConfig, char: string): HTMLElement {
  const all = Object.values(cfg.innate ?? {}).flatMap((n) => STAT_KEYS.map((k) => Math.abs(n[k])));
  const span = Math.max(0.01, ...all);
  const inn = cfg.innate?.[char];
  const row = h('div', 'preview-stats');
  for (const k of STAT_KEYS) {
    const v = inn?.[k] ?? 0;
    const pct = Math.round(v * 100);
    const fill = h('span', `stat-fill ${v > 0 ? 'pos' : v < 0 ? 'neg' : 'zero'}`);
    fill.style.width = `${(Math.abs(v) / span) * 50}%`;
    row.append(h('span', 'stat',
      h('span', 'stat-label', STR.statLabels[k]),
      h('span', 'stat-bar', fill),
      h('span', `stat-val ${v > 0 ? 'pos' : v < 0 ? 'neg' : ''}`, pct > 0 ? `+${pct}%` : `${pct}%`)));
  }
  return row;
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
/** Ngưỡng "điện thoại" của logo — trùng `@media (max-width: 600px)` trong style.css. */
const LOGO_MOBILE_MQ = '(max-width: 600px)';

/**
 * Logo tên game (chủ dự án chốt 02/11): PC → `logo_vi_1dong`, điện thoại → `logo_vi_2dong`,
 * mọi ngôn ngữ khác → `logo_en`. `<picture>` để trình duyệt chỉ tải ĐÚNG MỘT bản.
 * `alt` = tên game theo ngôn ngữ, nên `<h1>` vẫn có chữ cho máy tìm kiếm và trình đọc màn hình.
 * Manifest chưa có logo (máy dev chưa dựng asset) → chữ như cũ.
 */
function gameLogo(): HTMLElement {
  const logo = h('h1', 'logo');
  const isVi = LANG === 'vi';
  const main = assets.url(isVi ? 'brand/logo_vi_1dong.png' : 'brand/logo_en.png');
  if (!main) {
    return textLogo(logo);
  }
  const pic = document.createElement('picture');
  const mobile = isVi ? assets.url('brand/logo_vi_2dong.png') : '';
  if (mobile) {
    const src = document.createElement('source');
    src.media = LOGO_MOBILE_MQ;
    src.srcset = mobile;
    src.width = 1288;
    src.height = 539;
    pic.append(src);
  }
  const img = document.createElement('img');
  img.className = 'logo-img';
  img.src = main;
  img.alt = STR.title;
  img.width = 1288;
  img.height = 372;
  img.decoding = 'async';
  pic.append(img);
  logo.classList.add('has-img');
  logo.append(pic);
  return logo;
}

/**
 * Tên game bằng chữ: mỗi cụm một `<span>` không-ngắt (xem `STR.titleParts`). Khoảng trắng
 * giữa hai span là NÚT VĂN BẢN THẬT, để màn rộng nối liền thành một dòng.
 */
function textLogo(logo: HTMLElement): HTMLElement {
  STR.titleParts.forEach((part, i) => {
    if (i > 0) {
      logo.append(document.createTextNode(' '));
    }
    logo.append(h('span', 'logo-part', part));
  });
  return logo;
}

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
  // Mười nhân vật khác nhau chủ yếu ở kỹ năng (+ nội tại nhẹ ±15%, spec §6.4), nên
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
      h('div', 'preview-name', STR.charNick[c] ? `${STR.characters[c] ?? c} — ${STR.charNick[c]}` : (STR.characters[c] ?? c)),
      h('div', 'preview-skill', `${STR.skillLabel}: ${STR.skillNames[c] ?? '—'}`),
      h('div', 'preview-cd', cooldownText(cfg, c)),
      statBars(cfg, c),
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
  const logo = gameLogo();
  // Lưới nhân vật + khung chi tiết nằm trong một vùng CUỘN RIÊNG.
  //
  // Trước 2/11 cả panel là một khối `overflow-y: auto`. Khi nội dung vượt chiều cao cửa
  // sổ, thứ bị đẩy xuống dưới mép là thứ nằm CUỐI — tức hai cái nút. Đo 2/11 trên bản vừa
  // pull: nút "Vào chơi" KHÔNG nhìn thấy ở 4 trong 7 cỡ màn hình, gồm cả 1366×768. Người
  // mới mở game trên laptop phổ thông nhất không thấy nút để bắt đầu.
  //
  // Bốn tầng `max-height` dựng hôm 1/11 đã bóp hết mức rồi vẫn không đủ, vì trang chủ từ
  // đó dày thêm bốn thứ: ảnh logo, ô chọn ngôn ngữ, dòng hồi chiêu trên mỗi thẻ, và bốn
  // thanh chỉ số trong khung chi tiết. Thêm tầng thứ năm chỉ là hoãn lần vỡ sau.
  //
  // Nên đổi cách chia chỗ: phần GIỮA co giãn và cuộn, phần ĐẦU (tên, ô nhập) và phần CUỐI
  // (hai nút, dòng hướng dẫn) luôn nằm ngoài vùng cuộn. Thêm bao nhiêu nội dung vào lưới
  // nữa thì nút "Vào chơi" vẫn ở nguyên chỗ nhìn thấy được.
  const scroller = h('div', 'home-scroll', grid, preview);
  panel.append(
    langPicker(),
    logo,
    h('div', 'subtitle', STR.subtitle),
    input,
    err,
    h('div', 'label', STR.pickCharacter),
    scroller,
    h('div', 'row', button(STR.play, 'btn primary', play), button(STR.leaderboard, 'btn', onLeaderboard)),
    h('div', 'hint', 'ontouchstart' in window ? STR.controlsMobile : STR.controlsPc),
  );
  root.append(panel);
  return root;
}
