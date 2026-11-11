import { isTien, realmName, type GameConfig } from '../config';
import type { RoomBoardRow } from '../net/protocol';
import { assets } from '../render/assets';
import { headCrop } from '../render/headbox';
import { REALM_BADGE } from '../render/palette';
import { LB_ZONE_HIDDEN } from '../net/lbwire';
import { h } from './dom';
import { STR } from './strings';

/**
 * Những mảnh nhỏ của bảng xếp hạng phòng (spec §11.5): viên thuốc cảnh giới, nhãn khu,
 * chân dung, vương miện. Không giữ trạng thái — `roomboard.ts` lo phần đó.
 */

/** Màu icon từng khu (§11.5 "màu icon theo khu"), theo thứ tự `zoneNames`. */
const ZONE_COLORS = ['#ffd166', '#7bd88f', '#a3d977', '#f4a261', '#9fd8ff'];

/**
 * Icon khu vẽ bằng SVG (12×12): mái đình · núi · trúc · nhà · bông tuyết.
 *
 * Không dùng emoji như bản phác trong spec: emoji mỗi hệ điều hành một kiểu, Windows cũ còn
 * hiện ô vuông. Năm hình đơn giản cùng nét thì đọc được ở 12px trên mọi máy.
 */
const ZONE_ICONS = [
  'M1 4h10M2 2.5h8M3 4v7M9 4v7M3 7h6',
  'M1 11 4.5 4l2 3L8 5l3 6Z',
  'M4 1v10M8 2v9M4 4l-2-1M8 6l2-1M4 8l2-1',
  'M2 6 6 2.5 10 6M3 5.5V11h6V5.5M5 11V8h2v3',
  'M6 1v10M1.7 3.5l8.6 5M1.7 8.5l8.6-5',
];

const SVG_NS = 'http://www.w3.org/2000/svg';

function svgIcon(path: string, color: string): SVGSVGElement {
  const svg = document.createElementNS(SVG_NS, 'svg');
  svg.setAttribute('viewBox', '0 0 12 12');
  svg.setAttribute('class', 'rb-zicon');
  const p = document.createElementNS(SVG_NS, 'path');
  p.setAttribute('d', path);
  p.setAttribute('stroke', color);
  p.setAttribute('fill', 'none');
  p.setAttribute('stroke-width', '1.4');
  p.setAttribute('stroke-linecap', 'round');
  p.setAttribute('stroke-linejoin', 'round');
  svg.append(p);
  return svg;
}

/**
 * ZoneCell: icon + tên ngắn khu; "???" khi Ẩn Thân; chấm nhỏ khi cùng khu với mình.
 *
 * Là MỘT phần tử dựng sẵn rồi sửa nội dung, không phải hàm trả phần tử mới (v6.47 §11.5:
 * "dựng một lần, cập nhật chỉ sửa nội dung trong ô, không xoá/tạo lại phần tử"). Cả ba
 * trạng thái — có khu, Ẩn Thân, ô trống — đều dùng đúng số phần tử ấy; bản cũ trả về cây
 * DOM khác nhau cho từng trạng thái nên số phần tử của bảng nhảy liên tục (đo 07/11: 15
 * lần đổi trong 45 giây).
 */
export class ZoneCell {
  readonly el = h('span', 'rb-zone');
  private readonly icon = svgIcon('', '#888');
  private readonly name = h('span', 'rb-zname');
  private readonly same = h('i', 'rb-same');

  constructor() {
    this.same.title = STR.board.sameZone;
    this.el.append(this.icon, this.name, this.same);
  }

  set(row: RoomBoardRow | null, myZone: number): void {
    const hidden = !row || row.zoneId === LB_ZONE_HIDDEN || row.zoneId >= ZONE_ICONS.length;
    this.el.classList.toggle('hidden', hidden);
    this.el.classList.toggle('same', !!row && row.zoneId === myZone);
    this.name.textContent = !row ? '' : hidden ? STR.board.zoneHidden : (STR.board.zoneShort[row.zoneId] ?? '');
    if (!hidden && row) {
      const p = this.icon.firstElementChild;
      p?.setAttribute('d', ZONE_ICONS[row.zoneId]);
      p?.setAttribute('stroke', ZONE_COLORS[row.zoneId]);
    }
  }
}

/** realmShort: chữ cái đầu mỗi từ — "Kim Đan" → "KĐ" (§11.5 "hẹp quá thì viết tắt"). */
export function realmShort(name: string): string {
  return name.split(/\s+/).filter(Boolean).map((w) => w[0].toLocaleUpperCase('vi')).join('');
}

/** PillCell: viên thuốc màu huy hiệu cảnh giới §12.3, chữ "KIM ĐAN ⁵" hoặc "KĐ⁵". */
export class PillCell {
  readonly el = h('span', 'rb-pill');
  private readonly name = h('span', 'rb-pname');
  private readonly tier = h('sup');

  constructor(private readonly cfg: GameConfig, private readonly short: boolean) {
    this.el.append(this.name, this.tier);
  }

  set(row: RoomBoardRow | null): void {
    if (!row) {
      this.name.textContent = this.tier.textContent = '';
      this.el.title = '';
      return;
    }
    const name = realmName(this.cfg, row.realm);
    this.el.classList.toggle('tien', isTien(this.cfg, row.realm));
    const bg = REALM_BADGE[Math.min(row.realm, REALM_BADGE.length - 1)] ?? 0x555555;
    this.el.style.setProperty('--pill', `#${bg.toString(16).padStart(6, '0')}`);
    this.name.textContent = this.short ? realmShort(name) : name.toLocaleUpperCase('vi');
    this.tier.textContent = String(row.tier);
    this.el.title = `${name} ${row.tier}`;
  }
}

/** realmBorder: màu viền chân dung theo cảnh giới (§11.5 "viền màu cảnh giới"). */
export function realmBorder(realm: number): string {
  const c = REALM_BADGE[Math.min(realm, REALM_BADGE.length - 1)] ?? 0x888888;
  return `#${c.toString(16).padStart(6, '0')}`;
}

/**
 * face: chân dung tròn; thiếu art thì chấm tròn màu cảnh giới với chữ cái đầu tên.
 *
 * Cắt lấy phần đầu (`headCrop`) chứ không thu CẢ TẤM ảnh vào ô: chân dung là ảnh bán
 * thân/toàn thân, nhét vào ô 28–40 px thì mặt bé bằng hạt gạo, nhìn ra cái người chứ không
 * ra ai.
 */
export class FaceCell {
  readonly el = h('span', 'rb-face');

  constructor(private readonly cfg: GameConfig, size: number) {
    this.el.style.width = this.el.style.height = `${size}px`;
  }

  set(row: RoomBoardRow | null): void {
    if (!row) {
      this.el.textContent = '';
      this.el.style.backgroundImage = 'none';
      return;
    }
    const char = this.cfg.characters[row.character] ?? '';
    const url = char ? assets.url(`characters/${char}/portrait_neutral.png`) : '';
    this.el.style.borderColor = realmBorder(row.realm);
    if (url) {
      this.el.textContent = '';
      headCrop(this.el, url, char);
    } else {
      this.el.style.backgroundImage = 'none';
      this.el.textContent = row.name.slice(0, 1).toLocaleUpperCase('vi');
    }
  }
}

/**
 * crown: vương miện vẽ (§11.5 "không dùng emoji hệ thống").
 *
 * Ảnh `ui/crown_small.png` đã được artgen vẽ lại đúng 1/11 (bản cũ là cả một nhân vật đội
 * vương miện; giờ `artgen iconcheck` chặn kiểu ảnh đó). Thiếu ảnh thì vẽ bằng SVG.
 */
export function crown(): HTMLElement {
  const wrap = h('span', 'rb-crown');
  const url = assets.url('ui/crown_small.png');
  if (url) {
    const img = h('img');
    img.src = url;
    img.alt = '';
    wrap.append(img);
    return wrap;
  }
  const svg = document.createElementNS(SVG_NS, 'svg');
  svg.setAttribute('viewBox', '0 0 24 18');
  const band = document.createElementNS(SVG_NS, 'path');
  band.setAttribute('d', 'M3 16 1.5 5l6 4.5L12 2l4.5 7.5 6-4.5L21 16Z');
  band.setAttribute('fill', '#ffd23f');
  band.setAttribute('stroke', '#8a5200');
  band.setAttribute('stroke-width', '1.3');
  band.setAttribute('stroke-linejoin', 'round');
  const gem = document.createElementNS(SVG_NS, 'circle');
  gem.setAttribute('cx', '12');
  gem.setAttribute('cy', '12');
  gem.setAttribute('r', '2');
  gem.setAttribute('fill', '#e63946');
  svg.append(band, gem);
  wrap.append(svg);
  return wrap;
}

/** medal: huy chương bạc/đồng (art `ui/medal_*.png`), thiếu art thì chấm tròn đúng màu. */
export function medal(rank: 2 | 3): HTMLElement {
  const url = assets.url(rank === 2 ? 'ui/medal_silver.png' : 'ui/medal_bronze.png');
  const el = h('span', `rb-medal m${rank}`);
  if (url) {
    const img = h('img');
    img.src = url;
    img.alt = '';
    el.append(img);
  } else {
    el.textContent = String(rank);
  }
  return el;
}

/** fmtScore: "4 931" — nhóm nghìn bằng khoảng hẹp như bản phác §11.5, không dấu chấm. */
export function fmtScore(n: number): string {
  return String(Math.round(n)).replace(/\B(?=(\d{3})+(?!\d))/g, ' ');
}

/** titleText: chữ danh hiệu đang đeo; 0 thì theo hạng (chưa qua 3 s chống nhấp nháy). */
export function titleText(row: RoomBoardRow): string {
  return STR.titles[(row.titleId || row.rank) - 1] ?? '';
}

/**
 * ClipCell: chữ một dòng cắt "…", rê chuột thấy đủ (§11.5).
 *
 * Ô TRỐNG nhận một khoảng trắng cứng chứ không phải chuỗi rỗng: chuỗi rỗng làm hộp dòng xẹp
 * xuống 0 và bảng co lại đúng cái lỗi v6.47 phải sửa. Khoảng trắng cứng giữ nguyên chiều cao
 * dòng mà không hiện gì.
 */
export class ClipCell {
  readonly el: HTMLElement;

  constructor(cls: string) {
    this.el = h('span', `${cls} rb-clip`);
  }

  set(text: string): void {
    this.el.textContent = text || '\u00a0';
    this.el.title = text;
  }
}
