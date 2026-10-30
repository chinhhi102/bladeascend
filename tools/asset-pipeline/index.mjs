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

import { charIds, neverShip, plan, zoneObstacleKinds, ZONE_MAP_IDS } from './groups.mjs';

const HERE = dirname(fileURLToPath(import.meta.url));
const REPO = join(HERE, '..', '..');
const SRC = join(REPO, 'assets', 'src');
const DIST = join(REPO, 'assets', 'dist');
const MAPS = join(REPO, 'maps');
const CACHE = join(DIST, '.pipeline-cache.json');

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
  const buf = await img.webp({ quality: p.quality, alphaQuality: 100, effort: 5 }).toBuffer();
  return { buf, ext: 'webp', scale, w: Math.round(meta.width * scale), h: Math.round(meta.height * scale) };
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

for (const rel of shipped) {
  const p = plan(rel, ctx);
  const srcPath = join(SRC, rel);
  const srcBuf = readFileSync(srcPath);
  stats.srcBytes += srcBuf.length;
  const fingerprint = sha(Buffer.concat([Buffer.from(JSON.stringify(p)), srcBuf]));

  if (p.kind === 'atlas-json') {
    // Phải chờ biết atlas.png thu nhỏ bao nhiêu mới rescale được toạ độ → xử lý sau.
    // Vì vậy 17 tệp atlas.json KHÔNG dùng bộ đệm và lần nào cũng dựng lại. Chúng là JSON
    // vài KB nên chỉ tốn vài milligiây, và nội dung không đổi thì hash không đổi, tên tệp
    // trong `assets/dist` vẫn y nguyên — không làm mất cache của trình duyệt.
    atlasJsons.push({ rel, p, srcBuf, fingerprint });
    continue;
  }

  const hit = cache.entries[rel];
  if (hit && hit.fingerprint === fingerprint && existsSync(join(DIST, hit.url))) {
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
  out.set(rel, { ...entry, groups: p.groups });
  nextCache.entries[rel] = entry;
  stats.built++;
  stats.outBytes += built.buf.length;
  if (stats.built % 100 === 0) {
    log(`   … ${stats.built} tệp đã dựng`);
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
  files[rel] = { url: e.url, bytes: e.bytes, groups: e.groups };
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
const keep = new Set([...out.values()].map((e) => e.url).concat(['manifest.json', '.pipeline-cache.json']));
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
