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
 * Cách tìm (tự động), HAI TẦNG:
 *  1. **Alpha** cho biết CỘT của đầu: quét bề rộng từng hàng, chỗ hẹp nhất phía trên thân
 *     là cổ, phần trên cổ là đầu. Bề ngang lấy theo phân vị 8–92% để đòn gánh của Cô Ba
 *     hay cánh tay giơ lên không kéo khung rộng ra.
 *  2. **Màu da** cho biết MẶT nằm đâu trong cột đó: gộp mọi mảng màu da trong cột ấy (tóc,
 *     kính râm, râu cắt mặt thành nhiều mảnh nên phải gộp, lấy mảng to nhất là mất trán
 *     hoặc mất cằm). Nới lên 42% chiều cao mặt cho tóc/nón, xuống 10% cho râu và cằm.
 *
 * Vì sao cần tầng 2: bản đầu chỉ dùng alpha, đoán cằm theo vị trí cổ — nhưng với mặt có
 * râu, hàm bạnh hay cổ rụt thì chỗ hẹp nhất là QUAI HÀM, nằm trên cằm, nên khung cắt ngang
 * mồm. Đoán bù bằng "đầu cao 1,35 lần bề ngang" thì ảnh nào có nón to lại phình ra ôm cả
 * người. Vùng da thì chỉ đúng một chỗ: nó chính là cái mặt.
 *
 * Máy vẫn dò sai ở ảnh TOÀN THÂN (cột đầu theo alpha ôm cả người nên lọt bắp tay, bắp
 * chân) → bảng `OVERRIDE` bên dưới là số ĐO BẰNG MẮT, luôn thắng máy. Thêm nhân vật mới
 * thì chạy lại tool, xem tờ kiểm `--sheet`, ảnh nào lệch thì thêm vào `OVERRIDE`.
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
 * Sáu ảnh, đều cùng một kiểu sai: **ảnh toàn thân**. Cột đầu dò theo alpha ôm trọn người
 * (chỗ hẹp nhất rơi xuống eo hoặc khe hai chân), nên tầng 2 thừa hưởng luôn cái sai —
 * bắp tay bắp chân lọt vào vùng da và khung phình ra, mặt bé tí. A Hôn dính cả bốn biểu
 * cảm vì nhân vật này mặc áo ba lỗ với quần đùi: hở da gần hết người, lại đội nón bánh bao
 * màu kem cũng lọt vào mặt nạ da.
 */
const OVERRIDE = {
  'a_hon/neutral': [0.22, 0.05, 0.52],
  'a_hon/evil': [0.20, 0.07, 0.66],
  'a_hon/panic': [0.16, 0.01, 0.66],
  'a_hon/cry': [0.26, 0.05, 0.48],
  'ong_ba_vong/evil': [0.29, 0.06, 0.47],
  'shipper/evil': [0.55, 0.24, 0.45],
};

const pct = (arr, p) => arr[Math.min(arr.length - 1, Math.max(0, Math.round((arr.length - 1) * p)))];

/**
 * skinMask — pixel nào là DA MẶT.
 *
 * Bộ art này tô da bằng màu cam nhạt sáng. Ba điều kiện lọc gần hết thứ khác: phải SÁNG
 * (áo nâu của Shipper và Cô Ba tối hơn hẳn), đỏ phải trội và xanh dương thấp nhất, và sắc
 * màu dưới 45° (áo vàng nhạt của A Hôn ra ~52° nên rụng). Nón bánh bao màu kem thì lọt
 * vào, nhưng nó nằm ngay trên đầu nên không hại gì.
 */
function skinMask(data, W, H, C) {
  const m = new Uint8Array(W * H);
  for (let i = 0; i < W * H; i++) {
    if (data[i * C + 3] < 140) continue;
    const r = data[i * C], g = data[i * C + 1], b = data[i * C + 2];
    const mx = Math.max(r, g, b), mn = Math.min(r, g, b), d = mx - mn;
    if (mx < 185 || r !== mx || b > g) continue;
    const sat = d / mx;
    if (sat < 0.08 || sat > 0.60) continue;
    if (d !== 0 && 60 * ((g - b) / d) > 45) continue;
    m[i] = 1;
  }
  return m;
}

/** blobs — các mảng pixel LIỀN NHAU của mặt nạ, kèm hộp bao. */
function blobs(mask, W, H) {
  const seen = new Uint8Array(W * H), st = new Int32Array(W * H), out = [];
  for (let i = 0; i < W * H; i++) {
    if (!mask[i] || seen[i]) continue;
    let sp = 0; st[sp++] = i; seen[i] = 1;
    let a = 0, x0 = W, x1 = -1, y0 = H, y1 = -1;
    while (sp) {
      const p = st[--sp], x = p % W, y = (p / W) | 0;
      a++;
      if (x < x0) x0 = x; if (x > x1) x1 = x; if (y < y0) y0 = y; if (y > y1) y1 = y;
      if (x > 0 && mask[p - 1] && !seen[p - 1]) { seen[p - 1] = 1; st[sp++] = p - 1; }
      if (x < W - 1 && mask[p + 1] && !seen[p + 1]) { seen[p + 1] = 1; st[sp++] = p + 1; }
      if (y > 0 && mask[p - W] && !seen[p - W]) { seen[p - W] = 1; st[sp++] = p - W; }
      if (y < H - 1 && mask[p + W] && !seen[p + W]) { seen[p + W] = 1; st[sp++] = p + W; }
    }
    out.push({ a, x0, x1, y0, y1, cx: (x0 + x1) / 2 });
  }
  return out;
}

/** detect trả về khung đầu [trái, trên, cạnh] theo tỉ lệ: alpha tìm cột đầu, da tìm mặt. */
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

  // Tầng 2 — vùng DA trong cột đầu chính là cái mặt. Bỏ mảng nào lệch khỏi cột (bắp tay
  // giơ ngang) hoặc tụt sâu dưới cổ (bụng, bàn tay cầm đồ). Gộp hết phần còn lại: tóc,
  // kính râm, râu cắt mặt thành nhiều mảnh rời nên lấy mảng to nhất là mất trán hoặc cằm.
  const skin = blobs(skinMask(data, W, H, C), W, H)
    .filter((b) => b.a > W * H * 0.003)
    .filter((b) => b.cx > hx0 - hw * 0.12 && b.cx < hx1 + hw * 0.12 && b.y0 < neck + hw * 0.45);
  if (!skin.length) return null;
  const fx0 = Math.min(...skin.map((b) => b.x0)), fx1 = Math.max(...skin.map((b) => b.x1));
  const fy0 = Math.min(...skin.map((b) => b.y0)), fy1 = Math.max(...skin.map((b) => b.y1));
  const fw = fx1 - fx0 + 1, fh = fy1 - fy0 + 1;
  // Da bắt đầu từ chân tóc nên nới LÊN cho tóc/nón; nới XUỐNG ít cho râu và cằm.
  const by0 = Math.max(y0, fy0 - fh * 0.42);
  const by1 = fy1 + fh * 0.10;
  const bx0 = fx0 - fw * 0.10, bx1 = fx1 + fw * 0.10;
  let size = Math.round(Math.max(bx1 - bx0, by1 - by0) * 1.04);
  const cx = (bx0 + bx1) / 2, cy = (by0 + by1) / 2;
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

// Khung chiếm gần hết ảnh = máy dò hỏng (tưởng cả người là cái đầu). Báo đích danh, đừng
// để nó lọt vào `headbox.ts` trong im lặng như lần 05/11.
const ngo = rows.filter((r) => !r.manual && r.box[2] > 0.8);
if (ngo.length) {
  console.warn(`NGỜ: ${ngo.length} khung chiếm > 80% ảnh, xem lại rồi đo tay — ${ngo.map((r) => r.key).join(', ')}`);
}

const body = rows.map((r) => `  '${r.key}': [${r.box.join(', ')}],${r.manual ? ' // đo bằng mắt' : ''}`).join('\n');
writeFileSync(OUT, `// SINH TỰ ĐỘNG bằng \`node tools/portrait-headbox.mjs\` — đừng sửa tay.
// Sửa số thì sửa bảng OVERRIDE trong tool rồi chạy lại, nếu không lần chạy sau mất hết.

/**
 * Khung đầu của từng chân dung: \`[trái, trên, cạnh]\` theo TỈ LỆ cạnh ảnh (0–1), luôn vuông.
 *
 * Dùng cho mọi chỗ hiện MẶT nhân vật trong ô nhỏ: ô đại diện HUD (§12.4), bong bóng mặt
 * (§12.5), chân dung tròn ở bảng xếp hạng (§11.5), ảnh nhân vật ở màn chọn. Chân dung là
 * ảnh bán thân/toàn thân nên cắt bằng một luật chung ("lấy nửa trên, phóng 180%") là ra
 * cái nón với hai con mắt, còn thu cả tấm ảnh vào ô 28–40 px thì mặt bé bằng hạt gạo —
 * mỗi ảnh phải có khung riêng.
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

/**
 * headCrop gắn ảnh chân dung vào một thẻ HTML sao cho khung đầu vừa đúng cả ô.
 *
 * Mọi chỗ hiện mặt trong ô nhỏ đều gọi hàm này, không ai tự chép công thức: ô phải VUÔNG
 * (khung đầu vuông) và thẻ phải có \`overflow: hidden\` nếu ô bo tròn.
 *
 * Công thức phần trăm của \`background-position\`: 0% là mép trái ảnh trùng mép trái ô,
 * 100% là mép phải trùng mép phải — nên toạ độ khung đầu phải chia cho phần ảnh THỪA ra
 * ngoài ô (1 - cạnh), không phải nhân thẳng với 100.
 */
export function headCrop(el: HTMLElement, url: string, char: string, mood = 'neutral'): void {
  const [bx, by, bs] = headBox(char, mood);
  el.style.backgroundRepeat = 'no-repeat';
  el.style.backgroundSize = \`\${(100 / bs).toFixed(2)}%\`;
  el.style.backgroundPosition = bs >= 1 ? 'center'
    : \`\${((bx / (1 - bs)) * 100).toFixed(2)}% \${((by / (1 - bs)) * 100).toFixed(2)}%\`;
  el.style.backgroundImage = \`url(\${url})\`;
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
