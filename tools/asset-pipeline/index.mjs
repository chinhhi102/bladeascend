// tools/asset-pipeline — `assets/src` → `assets/dist` (GAME_SPEC §13.7 mục 2).
//
//   node tools/asset-pipeline/index.mjs            # dựng lại phần đã đổi
//   node tools/asset-pipeline/index.mjs --force    # dựng lại tất cả
//   node tools/asset-pipeline/index.mjs --quiet    # chỉ in bảng tổng kết
//
// Làm gì: thu nhỏ về kích thước hiển thị → WebP (có alpha) → tên file kèm hash nội dung →
// `manifest.json` ghi dung lượng byte và NHÓM TẢI của từng tệp. Âm thanh nén lại bằng ffmpeg.
//
// Có **bộ đệm**: mỗi tệp đích ghi kèm vân tay (hash nội dung nguồn + mọi tham số xử lý).
// Vân tay trùng thì bỏ qua. Chạy lần đầu ~90 giây, chạy lại khi không đổi gì ~1 giây — quan
// trọng vì `npm run build` gọi nó mỗi lần.
import { createHash } from 'node:crypto';
import { execFileSync } from 'node:child_process';
import { existsSync, mkdirSync, readFileSync, readdirSync, rmSync, statSync, writeFileSync } from 'node:fs';
import { dirname, join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

import { encodeToKTX2 } from 'ktx2-encoder';

import { sharpGrade } from '../colorgrade.mjs';
import { charIds, ktxFor, neverShip, plan, zoneObstacleKinds, ZONE_MAP_IDS } from './groups.mjs';

const HERE = dirname(fileURLToPath(import.meta.url));
const REPO = join(HERE, '..', '..');
const SRC = join(REPO, 'assets', 'src');
const DIST = join(REPO, 'assets', 'dist');
const MAPS = join(REPO, 'maps');
const CACHE = join(DIST, '.pipeline-cache.json');

/**
 * PIPELINE_VERSION — tăng MỖI KHI cách dựng ảnh đổi.
 *
 * Bộ đệm khoá theo `sha(kế hoạch + byte nguồn)`. Nó bắt được "art đổi" và "kế hoạch đổi",
 * nhưng KHÔNG bắt được "pipeline đổi" — và đó đúng là chỗ vừa sập: tôi thêm phép chỉnh màu
 * cho decal/trang trí, chạy build, thấy "ĐẠT" và commit — trong khi mọi tệp decal vẫn là
 * bản dựng từ 30/10. Ảnh không đổi, nên hash không đổi, nên bộ đệm trả lại bản cũ, và cả
 * `scenediff` lẫn bảng ngân sách đều không có gì để kêu.
 *
 * Số này nằm trong fingerprint, nên tăng nó là dựng lại toàn bộ.
 *
 *   1 — ban đầu
 *   2 — chỉnh màu nền cho decal/trang trí (§13.10), xuất thêm KTX2
 *   3 — ngưỡng KTX2 nâng 256 → 512 px (ảnh nhỏ nén khối tốn tải về hơn WebP)
 *   4 — `ui` chuyển UASTC → ETC1S (nền trang chủ 244 KB → 2,4 MB là không chấp nhận được)
 *   5 — ngưỡng cỡ TÁCH theo chế độ: UASTC ≥ 512 px, ETC1S ≥ 128 px (71 decal 256² đang
 *       chiếm 23,7 MB VRAM vì bị ngưỡng 512 loại ra)
 *   6 — dưới 512 px thì LUÔN ETC1S, kể cả nhóm UASTC (40 texture kiếm/VFX 256² = 13,3 MB)
 *   7 — vật cản maxPx 512 → 384 (hạng mục ngốn VRAM thứ hai)
 *   8 — KTX2 nhân sẵn alpha + gom atlas vật cản (lô vẽ, xem PERF_LOG #52)
 *   9 — dọn bộ đệm hỏng của lượt build 8 (mọi mục mất KTX2)
 */
const PIPELINE_VERSION = 9;

/**
 * Thư mục được áp ma trận chỉnh màu nền lúc build (§13.10).
 *
 * CHỈ decal và trang trí. Tile nền KHÔNG nằm đây: shader trộn đã tự áp ma trận, áp thêm lần
 * nữa vào ảnh là chỉnh màu hai lần.
 */
const GRADED_DIRS = ['decals/', 'decor/'];

const FORCE = process.argv.includes('--force');
const QUIET = process.argv.includes('--quiet');

let sharp;
try {
  sharp = (await import('sharp')).default;
} catch {
  console.error(
    'Thiếu `sharp`. Chạy:  cd tools/asset-pipeline && npm install\n' +
      '(GAME_SPEC §13.7 mục 2 đã duyệt thư viện này.)',
  );
  process.exit(1);
}

function log(...a) {
  if (!QUIET) {
    console.log(...a);
  }
}

function walk(dir, base = dir, out = []) {
  for (const e of readdirSync(dir, { withFileTypes: true })) {
    const p = join(dir, e.name);
    if (e.isDirectory()) {
      walk(p, base, out);
    } else if (e.isFile()) {
      out.push(relative(base, p).split('\\').join('/'));
    }
  }
  return out;
}

function sha(buf) {
  return createHash('sha256').update(buf).digest('hex');
}

const ctx = { zoneKinds: zoneObstacleKinds(MAPS), chars: charIds(SRC) };

const all = walk(SRC);
const shipped = all.filter((rel) => !neverShip(rel));

/** Vân tay: đổi nguồn HOẶC đổi tham số xử lý là dựng lại. Thiếu phần thứ hai thì sửa
 *  chất lượng WebP trong `groups.mjs` sẽ không có tác dụng gì — một cái bẫy im lặng. */
const cache = !FORCE && existsSync(CACHE) ? JSON.parse(readFileSync(CACHE, 'utf8')) : { entries: {} };
const nextCache = { entries: {} };

mkdirSync(DIST, { recursive: true });

/** Kết quả từng tệp, khoá là đường dẫn LOGIC (y như trong `assets/src`). */
const out = new Map();
const stats = { reused: 0, built: 0, srcBytes: 0, outBytes: 0 };
const atlasJsons = [];

function outName(rel, ext, hash) {
  const dot = rel.lastIndexOf('.');
  const stem = dot < 0 ? rel : rel.slice(0, dot);
  return `${stem}.${hash.slice(0, 8)}.${ext}`;
}

function write(relOut, buf) {
  const p = join(DIST, relOut);
  mkdirSync(dirname(p), { recursive: true });
  writeFileSync(p, buf);
}

async function buildImage(rel, p, srcBuf) {
  let img = sharp(srcBuf, { limitInputPixels: false });
  const meta = await img.metadata();
  let scale = 1;
  if (p.maxPx > 0 && Math.max(meta.width, meta.height) > p.maxPx) {
    scale = p.maxPx / Math.max(meta.width, meta.height);
    img = img.resize({
      width: Math.round(meta.width * scale),
      height: Math.round(meta.height * scale),
      fit: 'fill',
      kernel: 'lanczos3',
    });
  }
  // Chỉnh màu nền áp NGAY LÚC BUILD cho decal và trang trí (GAME_SPEC §13.10).
  //
  // Hai thứ đó nằm trên mặt đất và vốn được `ColorMatrixFilter` trên container làm dịu lại
  // cùng với nền. Bỏ filter đi (vì nó là một đích vẽ toàn màn hình mỗi khung) thì phép màu
  // phải đi đâu đó: nền tính trong shader, còn decal/trang trí là ảnh nên áp thẳng vào ảnh.
  //
  // `recomb` rồi `linear`: `sharp.recomb` chỉ làm phần 3×3, không có phần cộng, mà ma trận
  // của ta có phần cộng (do `contrast` sinh ra). Offset của Pixi tính trên thang 0–1 nên
  // phải nhân 255.
  if (GRADED_DIRS.some((d) => rel.startsWith(d))) {
    const g = sharpGrade();
    img = img.recomb(g.recomb).linear([1, 1, 1], g.offset255);
  }
  const buf = await img.webp({ quality: p.quality, alphaQuality: 100, effort: 5 }).toBuffer();
  const w = Math.round(meta.width * scale);
  const h = Math.round(meta.height * scale);
  // Bản KTX2 (§13.10) đi KÈM bản WebP, không thay nó: thiết bị không hỗ trợ nén khối vẫn
  // phải chơi được, và `assets.ts` chọn bản nào lúc chạy.
  const ktx = await buildKtx(rel, img, w, h);
  return { buf, ext: 'webp', scale, w, h, ktx };
}

/**
 * buildKtx xuất bản nén khối KTX2 — GAME_SPEC §13.10.
 *
 * UASTC cho thứ nhìn gần (nhân vật, vật phẩm, VFX), ETC1S cho thứ trải rộng (nền, vật cản,
 * decal). Khác nhau ở chỗ: UASTC giữ nét gần như WebP và tốn 1 byte/điểm ảnh trên GPU;
 * ETC1S mờ hơn một chút nhưng chỉ 0,5 byte, và tệp tải về nhẹ hơn WebP nhiều lần.
 *
 * Đo thật (lượt 13/11): atlas nhân vật 1152² — WebP 250 KB tải về, 5,3 MB trên GPU; UASTC
 * 457 KB tải về, **1,3 MB** trên GPU. Tile nền 512² — ETC1S **11 KB** tải về, 128 KB GPU.
 *
 * Trả `null` khi không áp dụng; mọi lỗi mã hoá cũng trả `null` kèm cảnh báo — một tệp không
 * nén được không được phép làm sập cả bản build (cùng luật với `buildAudio`).
 */
/** ktxFails: tệp nào không nén được — đếm để CHẶN khi hỏng hàng loạt (xem cuối tệp). */
const ktxFails = [];

async function buildKtx(rel, img, w, h) {
  const mode = ktxFor(rel, w, h);
  if (!mode) {
    return null;
  }
  try {
    const png = await img.clone().png().toBuffer();
    const opt = mode === 'uastc'
      ? { isUASTC: true, uastcLDRQualityLevel: 2, needSupercompression: true }
      : { isUASTC: false, qualityLevel: 160 };
    const out = await encodeToKTX2(png, {
      ...opt,
      isKTX: true,
      isSetKTX: true,
      generateMipmap: true,
      imageDecoder: rawDecoder,
    });
    return { buf: Buffer.from(out), mode };
  } catch (err) {
    console.error(`  ⚠ ${rel}: không nén KTX2 được (${String(err).split('\n')[0]}) → chỉ có WebP`);
    ktxFails.push(rel);
    return null;
  }
}

/** rawDecoder: bộ mã hoá KTX2 cần điểm ảnh thô RGBA; `sharp` đã có sẵn trong pipeline. */
async function rawDecoder(buf) {
  // `.premultiply()`: RGB nhân sẵn với alpha TRƯỚC khi nén.
  //
  // Vì sao — đo ngày 13/11 trên cảnh thử đông: lớp `entities` tốn 128 lô vẽ cho 64 vật cản,
  // tức ĐÚNG hai lô mỗi cái, dù cả cảnh chỉ còn hai nguồn ảnh. Nguyên nhân không phải số
  // texture: ép mọi sprite dùng chung một texture thì lô vẽ sụp từ 184 xuống 72.
  //
  // Nguyên nhân là CHẾ ĐỘ ALPHA. Pixi nạp canvas với `premultiply-alpha-on-upload`, còn bộ
  // đọc KTX2 khai `no-premultiply-alpha`. Hai chế độ cần hai phép trộn khác nhau, nên mỗi
  // lần xen kẽ là một lần đổi trạng thái trộn — và đổi trạng thái trộn thì cắt lô vẽ, dù
  // còn thừa chỗ texture (máy đo có 32 đơn vị texture).
  //
  // Tức là KTX2 đang ĐÁNH ĐỔI VRAM LẤY LÔ VẼ. Nhân sẵn alpha ở đây rồi khai
  // `premultiplied-alpha` ở client (`assets.ts`) thì hết đổi phép trộn, và giữ nguyên cả
  // phần VRAM đã hạ được.
  const { data, info } = await sharp(buf).ensureAlpha().raw().toBuffer({ resolveWithObject: true });
  // Nhân tay, KHÔNG dùng `sharp.premultiply()`: bản sharp trong repo không có hàm đó, và
  // `buildKtx` nuốt mọi lỗi nên gọi nhầm thì **mất sạch 401 tệp KTX2 trong im lặng** — lô vẽ
  // tụt đẹp đúng như mong đợi (184 → 61) vì cả cảnh rơi về WebP. Đã dính đúng một lần.
  for (let i = 0; i < data.length; i += 4) {
    const a = data[i + 3];
    if (a !== 255) {
      data[i] = (data[i] * a + 127) / 255 | 0;
      data[i + 1] = (data[i + 1] * a + 127) / 255 | 0;
      data[i + 2] = (data[i + 2] * a + 127) / 255 | 0;
    }
  }
  return { data, width: info.width, height: info.height };
}

/**
 * buildAudio nén lại **chỉ khi nguồn nặng hơn mức §13.7 yêu cầu**.
 *
 * Bộ âm thanh hiện tại (`tools/make_sfx.py`, `make_bgm.py`) đã xuất 22–44 kHz mono ở
 * 31–80 kb/s — tức đã DƯỚI ngưỡng 96/128 kb/s của spec. Ép qua libvorbis ở 128 kb/s với
 * nguồn 22 050 Hz mono thì nó báo "encoder setup failed" (ngoài dải hợp lệ của tần số đó),
 * và nếu có chạy được thì cũng chỉ làm tệp TO HƠN. Nên luật là: đo trước, nén sau.
 */
function buildAudio(rel, p, srcBuf) {
  const targetKbps = Number.parseInt(p.bitrate, 10);
  // ffmpeg không đọc được từ Buffer qua execFileSync một cách gọn gàng → dùng tệp tạm.
  // Tệp tạm PHẢI giữ đuôi gốc: ffmpeg chọn bộ giải mã theo đuôi, và `.tmp-in` không đuôi
  // làm nó báo "encoder setup failed" — một thông báo không nói gì về nguyên nhân thật.
  const ext0 = rel.slice(rel.lastIndexOf('.'));
  const tmpIn = join(DIST, `.tmp-in${ext0}`);
  const tmpOut = join(DIST, '.tmp-out.ogg');
  writeFileSync(tmpIn, srcBuf);
  try {
    const probe = execFileSync('ffprobe', ['-v', 'error', '-show_entries', 'stream=bit_rate,channels', '-of', 'json', tmpIn], {
      encoding: 'utf8',
    });
    const st = JSON.parse(probe).streams?.[0] ?? {};
    const srcKbps = Number(st.bit_rate ?? 0) / 1000 || (srcBuf.length * 8) / 1000; // thiếu số thì đoán cao để vẫn nén
    if (srcKbps <= targetKbps && (!p.mono || Number(st.channels ?? 1) <= 1)) {
      return { buf: srcBuf, ext: ext0.slice(1), note: `giữ nguyên (${Math.round(srcKbps)} kb/s ≤ ${targetKbps})` };
    }
    execFileSync(
      'ffmpeg',
      ['-y', '-loglevel', 'error', '-i', tmpIn, '-c:a', 'libvorbis', '-b:a', p.bitrate, ...(p.mono ? ['-ac', '1'] : []), '-f', 'ogg', tmpOut],
      { stdio: ['ignore', 'ignore', 'pipe'] },
    );
    const buf = readFileSync(tmpOut);
    return buf.length < srcBuf.length ? { buf, ext: 'ogg' } : { buf: srcBuf, ext: ext0.slice(1) };
  } catch (err) {
    // Một tệp âm thanh không nén được KHÔNG được phép làm sập cả bản build.
    console.error(`  ⚠ ${rel}: không nén lại được (${String(err).split('\n')[0]}) → phát hành nguyên trạng`);
    return { buf: srcBuf, ext: ext0.slice(1) };
  } finally {
    rmSync(tmpIn, { force: true });
    rmSync(tmpOut, { force: true });
  }
}

// ——— GOM ATLAS VẬT CẢN (GAME_SPEC §13.10 · nút thắt B30) ———
//
// Đo ngày 13/11 trên cảnh thật: lớp `entities` có **97 nguồn ảnh khác nhau** và 615 lô vẽ,
// trong đó phần đông nhất là vật cản — MỖI loại × MỖI hướng là một tệp 384² riêng
// (`border/intact`, `border/intact_vertical`, `cliff/intact`, `bamboo_wall/intact`…).
// Pixi gom được nhiều sprite vào một lô vẽ, nhưng mỗi lô chỉ buộc được 16 texture; 25 texture
// vật cản nằm xen kẽ nhau theo thứ tự y là 25 lần cắt lô.
//
// Gom chúng vào MỘT trang 2048² cho mỗi nhóm tải thì 20 texture thành 1.
//
// Gom THEO NHÓM TẢI, không gom tất: `boot` tải trước khi vào trận, còn `zone_*` chỉ tải khi
// vào khu đó. Trộn chung là bắt người chơi tải vật cản của bốn khu chưa tới — gói chặn đang
// còn đúng 0,98 MB dư (B8).
const PACK_PX = 2048;
/** Lề 4 điểm ảnh: ETC1S/UASTC nén theo khối 4×4, nên mọi khung phải nằm đúng mốc 4. */
const PACK_PAD = 4;

/**
 * planPacks xếp chỗ TRƯỚC khi dựng, để vòng lặp chính bỏ qua tệp đã được gom.
 *
 * Xếp kiểu "kệ": sắp theo chiều cao giảm dần rồi đặt từng hàng. Với vật cản (gần như đều
 * 384²) cách này không phí chỗ đáng kể, và nó ngắn hơn một bộ xếp hình thật sự vài chục lần.
 */
async function planPacks(rels, ctx) {
  const byGroup = new Map();
  for (const rel of rels) {
    if (!rel.startsWith('obstacles/') || !/\.png$/.test(rel)) {
      continue;
    }
    const p = plan(rel, ctx);
    if (p.kind !== 'image') {
      continue;
    }
    const key = p.groups.join('+');
    if (!byGroup.has(key)) {
      byGroup.set(key, { groups: p.groups, items: [] });
    }
    const meta = await sharp(readFileSync(join(SRC, rel)), { limitInputPixels: false }).metadata();
    const sc = p.maxPx > 0 && Math.max(meta.width, meta.height) > p.maxPx
      ? p.maxPx / Math.max(meta.width, meta.height) : 1;
    const w = Math.ceil(Math.round(meta.width * sc) / 4) * 4;
    const h = Math.ceil(Math.round(meta.height * sc) / 4) * 4;
    byGroup.get(key).items.push({ rel, p, w, h, sw: Math.round(meta.width * sc), sh: Math.round(meta.height * sc) });
  }

  const byRel = new Map();
  const pages = [];
  for (const [key, g] of byGroup) {
    if (g.items.length < 2) {
      continue; // một mình thì gom chẳng để làm gì
    }
    g.items.sort((a, b) => b.h - a.h || b.w - a.w);
    let page = null;
    let x = 0, y = 0, rowH = 0;
    const flush = () => {
      if (page) {
        page.h = Math.ceil((y + rowH) / 4) * 4;
        pages.push(page);
      }
    };
    const open = () => {
      page = { rel: `obstacles/_pack_${pages.length}_${key.replace(/[^a-z0-9_]+/gi, '-')}.png`,
        groups: g.groups, items: [], w: PACK_PX, h: PACK_PX };
      x = 0; y = 0; rowH = 0;
    };
    open();
    for (const it of g.items) {
      if (x + it.w + PACK_PAD > PACK_PX) {
        x = 0;
        y += rowH + PACK_PAD;
        rowH = 0;
      }
      if (y + it.h > PACK_PX) {
        flush();
        open();
      }
      page.items.push({ ...it, x, y });
      byRel.set(it.rel, { page, x, y });
      x += it.w + PACK_PAD;
      rowH = Math.max(rowH, it.h);
    }
    flush();
  }
  return { byRel, pages };
}

/** buildPack dựng một trang: ghép ảnh nguồn (đã thu nhỏ đúng plan) rồi nén như ảnh thường. */
async function buildPack(page) {
  const layers = [];
  for (const it of page.items) {
    let img = sharp(readFileSync(join(SRC, it.rel)), { limitInputPixels: false });
    if (it.sw !== undefined) {
      img = img.resize({ width: it.sw, height: it.sh, fit: 'fill', kernel: 'lanczos3' });
    }
    layers.push({ input: await img.png().toBuffer(), left: it.x, top: it.y });
  }
  const sheet = sharp({
    create: { width: page.w, height: page.h, channels: 4, background: { r: 0, g: 0, b: 0, alpha: 0 } },
  }).composite(layers);
  const png = await sheet.png().toBuffer();
  const img = sharp(png, { limitInputPixels: false });
  const buf = await img.clone().webp({ quality: 80, alphaQuality: 100, effort: 5 }).toBuffer();
  const ktx = await buildKtx(page.rel, img, page.w, page.h);
  return { buf, ktx };
}

// `KX_NO_PACK=1`: dựng KHÔNG gom, để đo được trước/sau cùng một bản mã.
const packs = process.env.KX_NO_PACK === '1' ? { byRel: new Map(), pages: [] } : await planPacks(shipped, ctx);
if (packs.pages.length > 0) {
  log(`   gom atlas vật cản: ${packs.byRel.size} tệp → ${packs.pages.length} trang`);
}

for (const rel of shipped) {
  const p = plan(rel, ctx);
  if (packs.byRel.has(rel)) {
    continue; // đã nằm trong một trang gom — xuất riêng nữa là tải về hai lần
  }
  const srcPath = join(SRC, rel);
  const srcBuf = readFileSync(srcPath);
  stats.srcBytes += srcBuf.length;
  const fingerprint = sha(Buffer.concat([Buffer.from(`v${PIPELINE_VERSION}:${JSON.stringify(p)}`), srcBuf]));

  if (p.kind === 'atlas-json') {
    // Phải chờ biết atlas.png thu nhỏ bao nhiêu mới rescale được toạ độ → xử lý sau.
    // Vì vậy 17 tệp atlas.json KHÔNG dùng bộ đệm và lần nào cũng dựng lại. Chúng là JSON
    // vài KB nên chỉ tốn vài milligiây, và nội dung không đổi thì hash không đổi, tên tệp
    // trong `assets/dist` vẫn y nguyên — không làm mất cache của trình duyệt.
    atlasJsons.push({ rel, p, srcBuf, fingerprint });
    continue;
  }

  const hit = cache.entries[rel];
  // Bộ đệm phải TỰ CHỮA khi thiếu bản nén.
  //
  // `PIPELINE_VERSION` chỉ bắt được "pipeline đã đổi", không bắt được "lượt build trước
  // chạy hỏng". Ngày 13/11 một lượt build lỗi ghi vào bộ đệm 401 mục version-8 KHÔNG có
  // KTX2; lượt sau sửa đúng mã nhưng dùng lại y nguyên bộ đệm đó, và VRAM nhảy 46 → 90 MB
  // trong khi bản build vẫn xanh. Một mục nói "không có bản nén" trong khi `ktxFor` nói
  // "phải có" là mục HỎNG, không phải mục dùng lại được.
  const wantKtx = hit && hit.w ? ktxFor(rel, hit.w, hit.h) !== null : false;
  const stale = wantKtx && !hit.ktx;
  if (hit && !stale && hit.fingerprint === fingerprint && existsSync(join(DIST, hit.url))) {
    out.set(rel, { ...hit, groups: p.groups });
    nextCache.entries[rel] = hit;
    stats.reused++;
    stats.outBytes += hit.bytes;
    continue;
  }

  let built;
  if (p.kind === 'image') {
    built = await buildImage(rel, p, srcBuf);
  } else if (p.kind === 'audio') {
    built = buildAudio(rel, p, srcBuf);
  } else {
    // font / data: phát hành nguyên trạng. Font đã là WOFF-class nhị phân, nén thêm là
    // việc của gzip/brotli ở tầng HTTP (§13.7 mục 2).
    const ext = rel.slice(rel.lastIndexOf('.') + 1);
    built = { buf: srcBuf, ext };
  }

  const url = outName(rel, built.ext, sha(built.buf));
  write(url, built.buf);
  const entry = { url, bytes: built.buf.length, fingerprint, scale: built.scale ?? 1, w: built.w, h: built.h };
  if (built.ktx) {
    const ktxUrl = outName(rel, 'ktx2', sha(built.ktx.buf));
    write(ktxUrl, built.ktx.buf);
    entry.ktx = { url: ktxUrl, bytes: built.ktx.buf.length, mode: built.ktx.mode };
  }
  out.set(rel, { ...entry, groups: p.groups });
  nextCache.entries[rel] = entry;
  stats.built++;
  stats.outBytes += built.buf.length;
  if (stats.built % 100 === 0) {
    log(`   … ${stats.built} tệp đã dựng`);
  }
}

// ——— dựng các trang gom vật cản ———
//
// Dựng SAU vòng chính để dùng lại `buildKtx`, và vì trang gom không có tệp nguồn nên nó
// không đi qua bộ đệm theo `fingerprint` được; mỗi lượt build lại vài trang 2048² tốn
// khoảng một giây.
for (const page of packs.pages) {
  const built = await buildPack(page);
  const url = outName(page.rel, 'webp', sha(built.buf));
  write(url, built.buf);
  const entry = { url, bytes: built.buf.length, fingerprint: 'pack', scale: 1, w: page.w, h: page.h };
  if (built.ktx) {
    const ktxUrl = outName(page.rel, 'ktx2', sha(built.ktx.buf));
    write(ktxUrl, built.ktx.buf);
    entry.ktx = { url: ktxUrl, bytes: built.ktx.buf.length, mode: built.ktx.mode };
  }
  out.set(page.rel, { ...entry, groups: page.groups });
  stats.built++;
  stats.outBytes += built.buf.length;
  // Mỗi tệp thành viên vẫn có một dòng manifest — mã game gọi `assets.texture('obstacles/
  // border/intact.png')` và không được biết gì về chuyện gom. `bytes: 0` để bài đo ngân sách
  // không đếm cùng một điểm ảnh hai lần; `pack` cho bộ nạp biết lấy khung nào trên trang nào.
  for (const it of page.items) {
    out.set(it.rel, {
      url, bytes: 0, fingerprint: 'pack', scale: 1, w: it.sw, h: it.sh,
      groups: page.groups,
      ...(built.ktx ? { ktx: { url: entry.ktx.url, bytes: 0, mode: built.ktx.mode } } : {}),
      pack: { page: page.rel, x: it.x, y: it.y, w: it.sw, h: it.sh },
    });
  }
}

// ——— atlas.json: rescale toạ độ theo đúng tỉ lệ atlas.png đã thu nhỏ ———
for (const { rel, p, srcBuf, fingerprint } of atlasJsons) {
  const pngRel = rel.replace(/atlas\.json$/, 'atlas.png');
  const png = out.get(pngRel);
  if (!png) {
    console.error(`atlas thiếu ảnh: ${pngRel} — bỏ ${rel}`);
    continue;
  }
  const scale = png.scale ?? 1;
  const atlas = JSON.parse(srcBuf.toString('utf8'));
  if (scale !== 1) {
    atlas.cell = Math.round(atlas.cell * scale);
    for (const f of Object.values(atlas.frames)) {
      f.x = Math.round(f.x * scale);
      f.y = Math.round(f.y * scale);
      f.w = Math.round(f.w * scale);
      f.h = Math.round(f.h * scale);
    }
    // `pivot_x`/`pivot_y` là tỉ lệ 0–1 nên KHÔNG nhân scale. Nhân vào là nhân vật tụt
    // xuống đất một phần ba chiều cao — và chỉ thấy khi đã ghép xong cả bộ art.
  }
  // Trỏ thẳng sang tệp ảnh đã có hash: bộ nạp không phải suy ra tên.
  atlas.image = png.url.split('/').at(-1);
  const buf = Buffer.from(JSON.stringify(atlas));
  const url = outName(rel, 'json', sha(buf));
  write(url, buf);
  const entry = { url, bytes: buf.length, fingerprint: `${fingerprint}:${png.url}` };
  out.set(rel, { ...entry, groups: p.groups });
  nextCache.entries[rel] = entry;
  stats.built++;
  stats.outBytes += buf.length;
}

// ——— manifest ———
// Một tệp có thể nằm trong NHIỀU nhóm (vật cản dùng ở 2–4 khu). `groups[g].bytes` vì vậy
// là "tải nhóm này một mình thì bao nhiêu byte", KHÔNG cộng lại thành tổng — tổng thật là
// `files`. Bộ nạp lọc trùng theo đường dẫn logic nên không tải lại.
const groups = {};
const files = {};
for (const [rel, e] of [...out.entries()].sort()) {
  files[rel] = { url: e.url, bytes: e.bytes, groups: e.groups, ...(e.ktx ? { ktx: e.ktx } : {}), ...(e.pack ? { pack: e.pack } : {}) };
  for (const g of e.groups) {
    (groups[g] ??= { bytes: 0, files: [] }).files.push(rel);
    groups[g].bytes += e.bytes;
  }
}

const manifest = {
  // `v` để bộ nạp cũ (manifest v1 chỉ có `files: string[]`) báo lỗi rõ thay vì đọc rác.
  v: 2,
  builtFrom: 'assets/src',
  zones: ZONE_MAP_IDS,
  chars: ctx.chars,
  groups: Object.fromEntries(Object.entries(groups).map(([g, v]) => [g, { bytes: v.bytes, files: v.files.sort() }])),
  files,
};
writeFileSync(join(DIST, 'manifest.json'), JSON.stringify(manifest));
writeFileSync(CACHE, JSON.stringify(nextCache));

// ——— Dọn tệp mồ côi: hash đổi thì bản cũ còn nằm đó và `assets:budget` sẽ đếm nó ———
// Giữ CẢ bản KTX2, không chỉ `url`.
//
// Bước dọn này xoá mọi tệp không nằm trong danh sách giữ — và vì danh sách chỉ lấy `e.url`,
// nó lặng lẽ xoá sạch mọi `.ktx2` vừa ghi ra. Triệu chứng: pipeline báo "ghi xong", log có
// đủ số byte, mà `find assets/dist -name '*.ktx2'` trả về 0. Mất gần một tiếng mới tìm ra,
// vì mọi bước TRƯỚC bước này đều báo thành công.
const keep = new Set(
  [...out.values()].flatMap((e) => (e.ktx ? [e.url, e.ktx.url] : [e.url]))
    .concat(['manifest.json', '.pipeline-cache.json']),
);
let removed = 0;
for (const rel of walk(DIST)) {
  if (!keep.has(rel)) {
    rmSync(join(DIST, rel), { force: true });
    removed++;
  }
}

// ——— Bảng tổng kết ———
const mb = (b) => (b / 1e6).toFixed(2);
const order = ['home', 'boot', ...ZONE_MAP_IDS.map((z) => `zone_${z}`), ...ctx.chars.map((c) => `char_${c}`), 'fx_high', 'audio', 'debug'];
const sortedGroups = [...Object.keys(groups)].sort((a, b) => {
  const ia = order.indexOf(a);
  const ib = order.indexOf(b);
  return (ia < 0 ? 99 : ia) - (ib < 0 ? 99 : ib) || a.localeCompare(b);
});

console.log(`\nasset-pipeline: ${stats.built} dựng mới, ${stats.reused} dùng lại bộ đệm${removed ? `, ${removed} tệp cũ xoá đi` : ''}`);
console.log(`  nguồn  ${mb(stats.srcBytes)} MB  →  xuất  ${mb(stats.outBytes)} MB   (−${(100 - (stats.outBytes / stats.srcBytes) * 100).toFixed(1)}%)\n`);
console.log('Nhóm'.padEnd(22), 'Dung lượng'.padStart(11), 'File'.padStart(6));
console.log('-'.repeat(22), '-'.repeat(11), '-'.repeat(6));
for (const g of sortedGroups) {
  console.log(g.padEnd(22), `${mb(groups[g].bytes)} MB`.padStart(11), String(groups[g].files.length).padStart(6));
}

// ——— Con số thật sự phải so với ngân sách 8 MB ———
//
// Không phải `boot`, mà là `home + boot + khu được xếp vào + nhân vật đã chọn` (§13.7 bảng
// nhóm): đó mới là tập tệp chặn người chơi vào trận. Lấy trường hợp XẤU NHẤT, vì người chơi
// không chọn được khu.
const worstZone = ZONE_MAP_IDS.map((z) => ({ z, b: groups[`zone_${z}`]?.bytes ?? 0 })).sort((a, b) => b.b - a.b)[0];
const worstChar = ctx.chars.map((c) => ({ c, b: groups[`char_${c}`]?.bytes ?? 0 })).sort((a, b) => b.b - a.b)[0];
const home = groups.home?.bytes ?? 0;
const blocking = home + (groups.boot?.bytes ?? 0) + worstZone.b + worstChar.b;
console.log(
  `\nChặn TRANG CHỦ            = home = ${mb(home)} MB` +
    `\nChặn VÀO TRẬN (xấu nhất)  = home + boot + zone_${worstZone.z} + char_${worstChar.c}` +
    `\n  ${mb(home)} + ${mb(groups.boot?.bytes ?? 0)} + ${mb(worstZone.b)} + ${mb(worstChar.b)} = ${mb(blocking)} MB` +
    `   (ngân sách §13.7: 8 MB — ${blocking <= 8e6 ? 'ĐẠT' : `VƯỢT ${(blocking / 8e6).toFixed(2)}×`})`,
);
console.log(`Toàn bộ game = ${mb(stats.outBytes)} MB   (ngân sách 80 MB — ${stats.outBytes <= 80e6 ? 'ĐẠT' : 'VƯỢT'})`);

// ——— KTX2 hỏng hàng loạt thì DỪNG, không in một dòng cảnh báo rồi đi tiếp ———
//
// `buildKtx` nuốt lỗi từng tệp để một ảnh hỏng không làm sập cả bản build. Đúng với MỘT tệp,
// sai hoàn toàn với 401 tệp: ngày 13/11 một hàm sharp gọi nhầm làm mọi tệp KTX2 biến mất,
// bản build vẫn xanh, và bài đo lô vẽ cho một con số đẹp bất thường (184 → 61) vì cả cảnh
// rơi về WebP. VRAM thì tăng gấp ba, nhưng bài đo lô vẽ không nhìn VRAM.
//
// Ngưỡng là 5: một vài ảnh cạnh lẻ không nén được là chuyện thường, hàng chục thì là lỗi mã.
const ktxWant = [...out.values()].filter((e) => e.w && ktxFor('x.png', e.w, e.h) !== null).length;
if (ktxFails.length > 5) {
  console.error(`\n✗ ${ktxFails.length} tệp không nén KTX2 được (${ktxFails.slice(0, 3).join(', ')}…).`);
  console.error('  Một vài tệp hỏng là chuyện thường; hàng chục tệp là lỗi mã, không phải lỗi ảnh.');
  process.exit(1);
}
const ktxHave = [...out.values()].filter((e) => e.ktx).length;
if (ktxWant > 20 && ktxHave === 0) {
  console.error('\n✗ KHÔNG có tệp KTX2 nào trong khi lẽ ra phải có. Dừng build.');
  process.exit(1);
}
console.log(`KTX2: ${ktxHave} tệp nén khối${ktxFails.length ? ` · ${ktxFails.length} tệp trượt` : ''}`);
