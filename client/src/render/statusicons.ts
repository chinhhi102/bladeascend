import { Texture } from 'pixi.js';

import type { StatusEv } from '../net/protocol';
import { assets } from './assets';

/**
 * Icon trạng thái dùng chung 64 px (spec v6.36 §6.5): 💤 ngủ · 🧋 mù · 💫 choáng · 📢 bị đẩy ·
 * 💪 gồng · ⚡ gom đơn · 🐢 rụt cổ · 🟣 trộm kiếm.
 *
 * Spec cấm emoji hệ thống (mỗi máy vẽ một kiểu, có máy ra ô vuông), nên vẽ bằng SVG — một
 * nguồn cho cả hai nơi: HUD (DOM, `<img>`) và trên đầu nhân vật (Pixi, texture). Khi job
 * artgen `icon_status_<tên>` sinh ra `ui/status/<tên>.png` thì ảnh đó thay chỗ vector.
 */

export const STATUS_ICON_KEYS = ['sleep', 'blind', 'stun', 'push', 'brace', 'ghost', 'shell', 'steal'] as const;
export type StatusIconKey = (typeof STATUS_ICON_KEYS)[number];

const INK = '#1b1410';
const RING = `stroke="${INK}" stroke-width="3" stroke-linejoin="round" stroke-linecap="round"`;

function badge(bg: string, glyph: string): string {
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">`
    + `<circle cx="32" cy="32" r="29" fill="${bg}" ${RING}/>${glyph}</svg>`;
}

const SVG: Record<StatusIconKey, string> = {
  sleep: badge('#5b7cfa',
    `<path d="M16 34h10l-10 12h10M30 16h16L30 34h16" fill="none" stroke="#fff" stroke-width="5" stroke-linejoin="round" stroke-linecap="round"/>`),
  blind: badge('#c89b6d',
    `<path d="M41 21l6-11" stroke="#e84a5f" stroke-width="5" stroke-linecap="round"/>`
    + `<path d="M19 21h26l-4 29H23z" fill="#f3e2c7" ${RING}/>`
    + `<g fill="#2b1a10"><circle cx="27" cy="44" r="3.5"/><circle cx="35" cy="45" r="3.5"/><circle cx="31" cy="38" r="3.5"/></g>`),
  stun: badge('#ffcf3f',
    `<path d="M14 36c6-10 30-10 36 0" fill="none" stroke="${INK}" stroke-width="3" stroke-dasharray="4 4"/>`
    + `<g fill="#fff" ${RING}>`
    + `<path d="M20 22l2.5 5 5.5.8-4 3.9 1 5.5-5-2.6-5 2.6 1-5.5-4-3.9 5.5-.8z"/>`
    + `<path d="M44 18l2.5 5 5.5.8-4 3.9 1 5.5-5-2.6-5 2.6 1-5.5-4-3.9 5.5-.8z"/>`
    + `<path d="M32 36l2.5 5 5.5.8-4 3.9 1 5.5-5-2.6-5 2.6 1-5.5-4-3.9 5.5-.8z"/></g>`),
  push: badge('#e84a5f',
    `<path d="M14 27v11h8l16 10V17L22 27z" fill="#fff" ${RING}/>`
    + `<path d="M43 25a9 9 0 0 1 0 14M48 20a16 16 0 0 1 0 24" fill="none" stroke="#fff" stroke-width="4" stroke-linecap="round"/>`),
  brace: badge('#ff8c2a',
    `<path d="M16 46c-1-9 3-15 9-17l3-9c1-4 8-4 9 0l1 7c7 0 12 6 10 13-2 7-10 9-19 9H19c-2 0-3-1-3-3z" fill="#ffe0c2" ${RING}/>`
    + `<path d="M30 36c3 2 7 2 10-1" fill="none" stroke="#d9534f" stroke-width="3" stroke-linecap="round"/>`),
  ghost: badge('#2b2b44',
    `<path d="M37 9L19 36h11l-4 19 19-29H34z" fill="#ffd23f" ${RING}/>`),
  shell: badge('#2fbf8f',
    `<path d="M12 40c0-13 9-22 20-22s20 9 20 22z" fill="#7ad7b0" ${RING}/>`
    + `<path d="M32 23l7 4v7l-7 4-7-4v-7zM25 34l-8 5M39 34l8 5M32 38v4" fill="none" stroke="${INK}" stroke-width="2.5" stroke-linejoin="round"/>`
    + `<path d="M10 42h44" stroke="${INK}" stroke-width="3" stroke-linecap="round"/>`),
  steal: badge('#8e44ad',
    `<circle cx="32" cy="32" r="16" fill="#c39bd3" ${RING}/>`
    + `<path d="M23 41l16-16" stroke="#fff" stroke-width="5" stroke-linecap="round"/>`
    + `<path d="M24 32l8 8M20 44l4-4" stroke="${INK}" stroke-width="3" stroke-linecap="round"/>`),
};

/**
 * iconOf: trạng thái nào hiện icon nào. Trượt đi kèm choáng (vỏ chuối) nên dùng chung 💫;
 * `sing` (hát ru 0,5 s), `slow`, `root`, `dash` không có icon trong bảng spec; `minion` vẽ
 * thanh đếm ngược thay cho icon.
 */
export function iconOf(code: StatusEv): StatusIconKey | null {
  if (code === 'slip') {
    return 'stun';
  }
  return (STATUS_ICON_KEYS as readonly string[]).includes(code) ? (code as StatusIconKey) : null;
}

const urls = new Map<StatusIconKey, string>();

/** statusIconUrl: địa chỉ ảnh cho HUD — ảnh artgen nếu đã có, không thì SVG nội tuyến. */
export function statusIconUrl(k: StatusIconKey): string {
  const art = `ui/status/${k}.png`;
  if (assets.has(art)) {
    return assets.url(art);
  }
  let u = urls.get(k);
  if (!u) {
    u = `data:image/svg+xml;charset=utf-8,${encodeURIComponent(SVG[k])}`;
    urls.set(k, u);
  }
  return u;
}

const TEX_PX = 128; // vẽ gấp đôi 64 px để màn hình DPR 2 vẫn nét
const textures = new Map<StatusIconKey, Texture | 'loading'>();

/**
 * statusIconTexture: texture cho Pixi; `null` trong lúc SVG đang dựng (một hai khung đầu).
 * Dựng MỘT lần mỗi loại rồi dùng chung cho mọi nhân vật — 8 texture cho cả trận.
 */
export function statusIconTexture(k: StatusIconKey): Texture | null {
  const art = assets.texture(`ui/status/${k}.png`);
  if (art) {
    return art;
  }
  const have = textures.get(k);
  if (have && have !== 'loading') {
    return have;
  }
  if (!have && typeof Image !== 'undefined') {
    textures.set(k, 'loading');
    const img = new Image();
    img.onload = () => {
      const cv = document.createElement('canvas');
      cv.width = TEX_PX;
      cv.height = TEX_PX;
      cv.getContext('2d')?.drawImage(img, 0, 0, TEX_PX, TEX_PX);
      textures.set(k, Texture.from(cv));
    };
    img.onerror = () => textures.delete(k);
    img.src = statusIconUrl(k);
  }
  return null;
}
