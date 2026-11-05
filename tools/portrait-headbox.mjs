#!/usr/bin/env node
/**
 * portrait-headbox — tìm KHUNG ĐẦU của từng ảnh chân dung, ghi ra
 * `client/src/render/headbox.ts`.
 *
 * Vì sao cần: ô đại diện HUD (§12.4) và bong bóng mặt (§12.5) đều là khung TRÒN, phải
 * "cắt từ đầu tới vai" như spec §12.4 ghi. Trước đây cả hai cắt bằng một luật CHUNG
 * `center top / 180%` trong CSS — luật đó chỉ đúng nếu mặt nằm giữa và chiếm gần hết mép
 * trên ảnh. Bộ chân dung thật là ảnh bán thân/toàn thân, đầu nhỏ và lệch, nên khung tròn
 * chỉ hứng được cái nón với hai con mắt (ảnh chụp 05/11). Mỗi ảnh cần một khung riêng.
 *
 * Cách tìm (tự động): quét alpha để lấy bề rộng từng hàng, chỗ hẹp nhất phía trên thân là
 * CỔ; phần trên cổ là đầu. Bề ngang đầu lấy theo phân vị 8–92% để vài hàng bất thường
 * (đòn gánh của Cô Ba Gánh, cánh tay giơ lên) không kéo khung rộng ra.
 *
 * Máy dò sai ở ảnh toàn thân và ảnh có nón to → bảng `OVERRIDE` bên dưới là số ĐO BẰNG MẮT,
 * luôn thắng máy. Thêm nhân vật mới thì chạy lại tool, xem tờ kiểm `--sheet`, ảnh nào lệch
 * thì thêm vào `OVERRIDE`.
 *
 *   node tools/portrait-headbox.mjs            # ghi client/src/render/headbox.ts
 *   node tools/portrait-headbox.mjs --sheet /tmp/box.png   # thêm tờ ảnh kiểm bằng mắt
 *
 * Cần `sharp` — đã có sẵn trong tools/asset-pipeline (chạy `npm install` ở đó trước).
 */
import { readdirSync, existsSync, writeFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const HERE = dirname(fileURLToPath(import.meta.url));
const ROOT = resolve(HERE, '..');
const CHARS = resolve(ROOT, 'assets/src/characters');
const OUT = resolve(ROOT, 'client/src/render/headbox.ts');
const MOODS = ['neutral', 'evil', 'panic', 'cry'];

const require = createRequire(resolve(ROOT, 'tools/asset-pipeline/package.json'));
let sharp;
try {
  sharp = require('sharp');
} catch {
  console.error('Thiếu sharp. Chạy: cd tools/asset-pipeline && npm install');
  process.exit(1);
}

/**
 * OVERRIDE — khung đo bằng mắt, [trái, trên, cạnh] theo TỈ LỆ cạnh ảnh (0–1).
 *
 * Chỉ ghi ở đây những ảnh máy dò sai. Lý do sai gom thành hai kiểu:
 *  - ảnh TOÀN THÂN (`*_evil` của Ông Bà Vòng, Shipper, Sư Phụ Mì, Cô Ba Gánh neutral):
 *    chỗ hẹp nhất rơi xuống tận eo → khung ôm cả người;
 *  - mặt có RÂU hoặc hàm bạnh (Cu Rùa, A Hôn evil): chỗ hẹp nhất là quai hàm, nằm TRÊN
 *    cằm → khung cắt mất mồm.
 */
const OVERRIDE = {
  'a_hon/evil': [0.20, 0.08, 0.64],
  'co_ba_ganh/neutral': [0.30, 0.03, 0.42],
  'cu_rua/neutral': [0.22, 0.07, 0.58],
  'cu_rua/panic': [0.22, 0.06, 0.62],
  'cu_rua/cry': [0.25, 0.08, 0.58],
  'ong_ba_vong/evil': [0.33, 0.09, 0.36],
  'ong_ba_vong/cry': [0.30, 0.06, 0.54],
  'shipper/evil': [0.45, 0.17, 0.40],
  'shipper/panic': [0.45, 0.18, 0.40],
  'shipper/cry': [0.39, 0.14, 0.44],
  'su_phu_mi/neutral': [0.31, 0.11, 0.50],
  'su_phu_mi/evil': [0.32, 0.10, 0.38],
};

const pct = (arr, p) => arr[Math.min(arr.length - 1, Math.max(0, Math.round((arr.length - 1) * p)))];

/** detect trả về khung đầu [trái, trên, cạnh] theo tỉ lệ, dò bằng alpha. */
async function detect(file) {
  const { data, info } = await sharp(file).ensureAlpha().raw().toBuffer({ resolveWithObject: true });
  const W = info.width, H = info.height, C = info.channels;
  const rowMin = new Int32Array(H).fill(-1), rowMax = new Int32Array(H).fill(-1);
  let y0 = H, y1 = -1;
  for (let y = 0; y < H; y++) {
    for (let x = 0; x < W; x++) {
      if (data[(y * W + x) * C + 3] > 40) {
        if (rowMin[y] < 0) rowMin[y] = x;
        rowMax[y] = x;
      }
    }
    if (rowMax[y] >= 0) { if (y < y0) y0 = y; y1 = y; }
  }
  if (y1 < 0) return null;
  const rawW = (y) => (rowMax[y] < 0 ? 0 : rowMax[y] - rowMin[y] + 1);
  // Làm trơn bề rộng ±3 hàng: viền nét vẽ làm bề rộng răng cưa, không trơn thì "chỗ hẹp
  // nhất" rơi vào một hàng ngẫu nhiên.
  const w = new Float64Array(H);
  for (let y = 0; y < H; y++) {
    let s = 0, n = 0;
    for (let k = -3; k <= 3; k++) { const j = y + k; if (j >= 0 && j < H) { s += rawW(j); n++; } }
    w[y] = s / n;
  }
  const subjH = y1 - y0 + 1;
  const span = (top, bot) => {
    const mi = [], ma = [];
    for (let y = top; y <= bot; y++) if (rowMin[y] >= 0) { mi.push(rowMin[y]); ma.push(rowMax[y]); }
    mi.sort((a, b) => a - b); ma.sort((a, b) => a - b);
    return [pct(mi, 0.08), pct(ma, 0.92)];
  };
  const neckIn = (hi) => {
    const lo = y0 + Math.round(subjH * 0.15);
    let found = -1, best = Infinity;
    for (let y = lo; y <= hi; y++) {
      let isMin = true;
      for (let k = -10; k <= 10; k++) { const j = y + k; if (j > lo && j < hi && w[j] < w[y] - 0.5) { isMin = false; break; } }
      if (isMin && w[y] < best) { best = w[y]; found = y; }
    }
    return found < 0 ? Math.round((lo + hi) / 2) : found;
  };
  let hi = y0 + Math.round(subjH * 0.75);
  let neck = neckIn(hi);
  let [hx0, hx1] = span(y0, neck);
  // Đầu không cao quá 1,35 lần bề ngang. Quá thì "cổ" là chỗ khác (eo, khe hai chân) →
  // thu trần tìm kiếm rồi dò lại.
  for (let i = 0; i < 4; i++) {
    const cap = y0 + Math.round((hx1 - hx0 + 1) * 1.35);
    if (neck <= cap) break;
    hi = Math.min(cap, hi - 1);
    neck = neckIn(hi);
    [hx0, hx1] = span(y0, neck);
  }
  const hw = hx1 - hx0 + 1;
  const bottom = neck + hw * 0.20; // cổ dò được nằm ở quai hàm, nới xuống cho hết cằm
  let size = Math.round(Math.max(hw, bottom - y0) * 1.12);
  const cx = (hx0 + hx1) / 2, cy = (y0 + bottom) / 2;
  let left = Math.max(0, Math.min(Math.round(cx - size / 2), W - size));
  let top = Math.max(0, Math.min(Math.round(cy - size / 2), H - size));
  size = Math.min(size, W - left, H - top);
  return [left / W, top / H, size / W];
}

const r3 = (n) => Number(n.toFixed(3));

const rows = [];
for (const char of readdirSync(CHARS).sort()) {
  for (const mood of MOODS) {
    const file = `${CHARS}/${char}/portrait_${mood}.png`;
    if (!existsSync(file)) continue;
    const key = `${char}/${mood}`;
    const manual = OVERRIDE[key];
    const box = manual ?? (await detect(file));
    if (!box) { console.warn(`bỏ qua ${key}: ảnh rỗng`); continue; }
    rows.push({ key, box: box.map(r3), manual: Boolean(manual), file });
  }
}

const body = rows.map((r) => `  '${r.key}': [${r.box.join(', ')}],${r.manual ? ' // đo bằng mắt' : ''}`).join('\n');
writeFileSync(OUT, `// SINH TỰ ĐỘNG bằng \`node tools/portrait-headbox.mjs\` — đừng sửa tay.
// Sửa số thì sửa bảng OVERRIDE trong tool rồi chạy lại, nếu không lần chạy sau mất hết.

/**
 * Khung đầu của từng chân dung: \`[trái, trên, cạnh]\` theo TỈ LỆ cạnh ảnh (0–1), luôn vuông.
 *
 * Dùng cho hai chỗ cắt ảnh vào khung TRÒN: ô đại diện HUD (§12.4) và bong bóng mặt (§12.5).
 * Chân dung là ảnh bán thân/toàn thân nên cắt bằng một luật chung ("lấy nửa trên, phóng
 * 180%") là ra cái nón với hai con mắt — mỗi ảnh phải có khung riêng.
 */
export const HEAD_BOX: Record<string, readonly [number, number, number]> = {
${body}
};

/**
 * Khung mặc định cho ảnh chưa có trong bảng (nhân vật mới, art vừa thả vào). Lấy nửa trên
 * và hơi hẹp hai bên — không chuẩn bằng số đo riêng nhưng còn thấy được mặt.
 */
export const DEFAULT_HEAD_BOX: readonly [number, number, number] = [0.2, 0.02, 0.6];

/** headBox tra khung đầu theo nhân vật + biểu cảm. */
export function headBox(char: string, mood: string): readonly [number, number, number] {
  return HEAD_BOX[\`\${char}/\${mood}\`] ?? HEAD_BOX[\`\${char}/neutral\`] ?? DEFAULT_HEAD_BOX;
}
`);
console.log(`ghi ${OUT} — ${rows.length} ảnh (${rows.filter((r) => r.manual).length} đo bằng mắt)`);

const sheetAt = process.argv.indexOf('--sheet');
if (sheetAt >= 0) {
  const out = process.argv[sheetAt + 1];
  if (!out) { console.error('--sheet cần đường dẫn ảnh'); process.exit(1); }
  const T = 120, GAP = 4, COLS = MOODS.length;
  const tiles = [];
  let x = 0, y = 0, col = 0, lastChar = '';
  for (const r of rows) {
    const char = r.key.split('/')[0];
    if (lastChar && char !== lastChar) { x = 0; y += T + GAP; col = 0; }
    lastChar = char;
    const meta = await sharp(r.file).metadata();
    const size = Math.max(1, Math.round(r.box[2] * meta.width));
    const left = Math.min(Math.max(0, Math.round(r.box[0] * meta.width)), meta.width - size);
    const top = Math.min(Math.max(0, Math.round(r.box[1] * meta.height)), meta.height - size);
    tiles.push({
      input: await sharp(r.file).extract({ left, top, width: size, height: size })
        .resize(T, T).flatten({ background: '#ffffff' }).png().toBuffer(),
      left: x, top: y,
    });
    col++; x += T + GAP;
  }
  const h = y + T + GAP;
  await sharp({ create: { width: COLS * (T + GAP), height: h, channels: 3, background: '#222222' } })
    .composite(tiles).png().toFile(out);
  console.log(`tờ kiểm: ${out}`);
}
