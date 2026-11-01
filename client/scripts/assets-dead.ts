// npm run check:dead — chặn build nếu `assets/dist` có tệp KHÔNG NƠI NÀO trong client nạp.
//
// Hai bài kiểm ngược chiều nhau, và cần cả hai:
//   `check:assets` — code ĐÒI một tệp mà tệp không có  → người chơi thấy placeholder.
//   `check:dead`   — tệp CÓ mà không ai đòi            → người chơi tải về rồi vứt đi.
//
// Vì sao đáng chặn build: ngày 31/10 thư mục `assets/src` còn 78 tệp art v6.7 đã bỏ
// (`break/`, `cracked`, `nearly_broken`, `rubble`; nay nằm trong `assets/library/`) — **12,9 MB chưa từng được vẽ một lần
// nào**. Chúng đẩy gói tải vượt ngân sách 8 MB và chặn deploy, mà không ai hiểu vì sao
// nặng. Không có bài kiểm này thì chuyện đó lặp lại mỗi lần đổi hướng art.
import { existsSync, readFileSync, readdirSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

import { BIOME_TILES, catalog, OBSTACLE_ALIAS } from '../src/render/catalog.ts';

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..', '..');
const DIST = join(ROOT, 'assets', 'dist');
const SRC = join(ROOT, 'client', 'src');
const MAPS = join(ROOT, 'maps');

/**
 * ALLOW: những tệp CỐ Ý có mặt mà không nơi nào nạp.
 *
 * Danh sách phải ngắn và mỗi dòng phải có lý do. Một danh sách miễn trừ dài là cách êm ái
 * để bài kiểm này ngừng kiểm được gì.
 */
const ALLOW: { pattern: RegExp; why: string }[] = [
  { pattern: /^favicon/, why: 'trình duyệt tự xin, không qua assets.url()' },
  { pattern: /^og\//, why: 'ảnh xem trước khi chia sẻ link; mạng xã hội lấy, client không' },
];

const manifestPath = join(DIST, 'manifest.json');
if (!existsSync(manifestPath)) {
  console.log('Chưa có assets/dist/manifest.json — chạy `node tools/asset-pipeline/index.mjs` trước. Bỏ qua.');
  process.exit(0);
}
interface Manifest {
  files: Record<string, { url: string; bytes: number; groups?: string[] }>;
}
const manifest = JSON.parse(readFileSync(manifestPath, 'utf8')) as Manifest;

// ── Gom MỌI đường dẫn logic mà client có thể xin ────────────────────────────────────────

const wanted = new Set<string>();
/**
 * dirsWanted: THƯ MỤC được dùng, chứ không phải tệp.
 *
 * Cần vì `tools/asset-pipeline` GỘP các khung hình rời thành một `atlas.png` + `atlas.json`
 * (§13.7). Danh mục art liệt kê `characters/a_hon/run_00.png`, còn trong `assets/dist` chỉ
 * còn `characters/a_hon/atlas.png` — so theo tên tệp thì mọi atlas đều bị gắn nhãn "chết",
 * tức là 11 atlas nhân vật và 6 atlas quái, 3,2 MB báo nhầm. So theo THƯ MỤC thì đúng.
 */
const dirsWanted = new Set<string>();
const add = (p: string): void => {
  if (p) {
    wanted.add(p.replace(/^\/+/, ''));
  }
};

const dirOf = (p: string): string => p.slice(0, p.lastIndexOf('/'));

// 1. Danh mục art (nhân vật, quái, vật phẩm, kiếm, vật cản, giao diện…).
for (const e of catalog()) {
  for (const f of e.files) {
    add(f);
    if (f.includes('/')) {
      dirsWanted.add(dirOf(f));
      // Khung hình nguồn nằm ở `.../frames/`, nhưng atlas gộp ra lại nằm ở thư mục CHA
      // (`monsters/x/frames/walk_00.png` → `monsters/x/atlas.png`). Đánh dấu cả cha.
      if (dirOf(f).endsWith('/frames')) {
        dirsWanted.add(dirOf(dirOf(f)));
      }
    }
  }
}
// Art mượn: alias trỏ từ loại chưa có art sang loại đã có, nên NGUỒN được dùng thật.
for (const [kind, al] of Object.entries(OBSTACLE_ALIAS)) {
  void kind;
  for (const e of catalog()) {
    for (const f of e.files) {
      if (f.startsWith(`obstacles/${al.from}/`)) {
        add(f);
      }
    }
  }
}
// 2. Ô nền theo biome.
for (const tiles of Object.values(BIOME_TILES)) {
  for (const t of tiles) {
    add(`tiles/${t}.png`);
  }
}
// 2b. Ô nền có thật nhưng chưa khai trong BIOME_TILES thì KHÔNG coi là dùng — đó đúng là
//     thứ bài kiểm này sinh ra để bắt.

// 3. Trang trí VÀ VẬT CẢN do bản đồ đặt.
interface TmjObj { name?: string; type?: string; properties?: { name: string; value: unknown }[] }
interface TmjLayer { name?: string; objects?: TmjObj[]; layers?: TmjLayer[] }
function walk(l: TmjLayer, out: TmjObj[]): void {
  for (const o of l.objects ?? []) {
    out.push(o);
  }
  for (const c of l.layers ?? []) {
    walk(c, out);
  }
}
if (existsSync(MAPS)) {
  for (const f of readdirSync(MAPS).filter((x) => x.endsWith('.tmj'))) {
    const doc = JSON.parse(readFileSync(join(MAPS, f), 'utf8')) as { layers?: TmjLayer[] };
    const objs: TmjObj[] = [];
    for (const l of doc.layers ?? []) {
      walk(l, objs);
    }
    for (const o of objs) {
      const kind = String(o.type || o.name || '');
      if (!kind) {
        continue;
      }
      add(`decor/${kind}.png`);
      // Vật cản: art có thể là MỘT tệp (`obstacles/rock.png`) hoặc một THƯ MỤC nhiều
      // trạng thái (`obstacles/stall/states_00.png`). Đánh dấu cả hai dạng.
      add(`obstacles/${kind}.png`);
      dirsWanted.add(`obstacles/${kind}`);
      const al = OBSTACLE_ALIAS[kind];
      if (al) {
        add(`obstacles/${al.from}.png`);
        dirsWanted.add(`obstacles/${al.from}`);
      }
    }
  }
}
// 4. Chuỗi hằng trong mã client.
//
// Quét CHUỖI chứ không phân tích cú pháp: một đường dẫn ghép bằng biến thì bài kiểm này
// không thấy, nên nó có thể báo nhầm "chết". Đó chính là lý do kết quả ra CẢNH BÁO kèm tên
// tệp, và chỉ chặn build khi vượt ngưỡng dung lượng — xem cuối tệp.
const SRC_EXT = ['.ts', '.tsx', '.css', '.html'];
function scan(dir: string, out: string[]): void {
  for (const e of readdirSync(dir, { withFileTypes: true })) {
    const p = join(dir, e.name);
    if (e.isDirectory()) {
      scan(p, out);
    } else if (SRC_EXT.some((x) => e.name.endsWith(x))) {
      out.push(readFileSync(p, 'utf8'));
    }
  }
}
const sources: string[] = [];
scan(SRC, sources);
for (const dir of ['client/index.html', 'client/admin']) {
  const p = join(ROOT, dir);
  if (!existsSync(p)) {
    continue;
  }
  if (p.endsWith('.html')) {
    sources.push(readFileSync(p, 'utf8'));
  } else {
    scan(p, sources);
  }
}
const all = sources.join('\n');
// Đường dẫn asset luôn có đuôi tệp; bắt cả chuỗi nháy đơn, nháy kép và backtick.
for (const m of all.matchAll(/["'`]([\w\-./]+\.(?:png|jpg|webp|ogg|mp3|wav|json|woff2?|ttf))["'`]/g)) {
  add(m[1]);
}
// Tên gốc (không đuôi) hay dùng cho âm thanh: `sound.play('hit')`.
const bare = new Set<string>();
for (const m of all.matchAll(/["'`]([\w\-/]{2,60})["'`]/g)) {
  bare.add(m[1]);
}

// ── So với manifest ─────────────────────────────────────────────────────────────────────

/** ATLAS: tệp do pipeline gộp ra, tên cố định trong mỗi thư mục. */
const ATLAS = new Set(['atlas.png', 'atlas.json', 'atlas.webp']);

function isWanted(logical: string): boolean {
  if (wanted.has(logical)) {
    return true;
  }
  // Atlas gộp: dùng nếu THƯ MỤC của nó có ai dùng.
  const base = logical.slice(logical.lastIndexOf('/') + 1);
  if (logical.includes('/') && ATLAS.has(base) && dirsWanted.has(dirOf(logical))) {
    return true;
  }
  if (ALLOW.some((a) => a.pattern.test(logical))) {
    return true;
  }
  // Âm thanh gọi theo TÊN TRẦN (`sound.play('hit')`), không theo đường dẫn.
  const noExt = logical.replace(/\.[a-z0-9]+$/, '');
  if (bare.has(noExt) || bare.has(noExt.split('/').pop() ?? '')) {
    return true;
  }
  return false;
}

const dead: { file: string; bytes: number }[] = [];
let totalBytes = 0;
for (const [logical, info] of Object.entries(manifest.files)) {
  totalBytes += info.bytes;
  if (!isWanted(logical)) {
    dead.push({ file: logical, bytes: info.bytes });
  }
}
dead.sort((a, b) => b.bytes - a.bytes);
const deadBytes = dead.reduce((a, d) => a + d.bytes, 0);
const mb = (b: number): string => (b / 1024 / 1024).toFixed(2);

console.log(`Đã soi ${Object.keys(manifest.files).length} tệp trong assets/dist (${mb(totalBytes)} MB).`);
if (dead.length === 0) {
  console.log('OK: không có tệp art chết nào.');
  process.exit(0);
}
console.log(`\n${dead.length} tệp KHÔNG tìm thấy nơi nào nạp — ${mb(deadBytes)} MB:`);
for (const d of dead.slice(0, 30)) {
  console.log(`  ${(d.bytes / 1024).toFixed(0).padStart(6)} KB  ${d.file}`);
}
if (dead.length > 30) {
  console.log(`  … và ${dead.length - 30} tệp nữa.`);
}

/**
 * DEAD_MB_MAX: ngưỡng chặn build.
 *
 * Không chặn ngay từ tệp đầu tiên, vì phép quét chuỗi ở trên KHÔNG thấy đường dẫn ghép
 * bằng biến — nó có thể báo nhầm. Một bài kiểm lúc nào cũng đỏ sẽ bị bỏ qua, rồi tới lúc
 * đỏ vì lý do thật thì không ai nhìn nữa.
 *
 * 1 MB: đủ rộng cho vài tệp báo nhầm lẻ tẻ, đủ chặt để bắt được đúng thứ đã xảy ra —
 * 12,9 MB art đã bỏ nằm lại trong gói tải.
 */
const DEAD_MB_MAX = 1;
// Luật 13 CLAUDE.md / spec §13.7-3b: art đã sinh là tài nguyên trả tiền — KHÔNG BAO GIỜ đề xuất
// xoá. Không dùng nữa thì chuyển sang `assets/library/`, nơi pipeline không bao giờ đọc tới.
const MOVE_HINT = [
  'Không dùng nữa → CHUYỂN vào assets/library/<nhóm>/ (git mv, giữ nguyên tên) và ghi',
  'assets/library/INDEX.md. KHÔNG xoá art đã sinh. Định dùng → nối vào code.',
].join('\n');
if (deadBytes > DEAD_MB_MAX * 1024 * 1024) {
  console.error(`\nLỖI: ${mb(deadBytes)} MB art chết > ngưỡng ${DEAD_MB_MAX} MB.`);
  console.error(MOVE_HINT);
  console.error('Báo nhầm (đường dẫn ghép bằng biến) thì thêm vào ALLOW trong tệp này, KÈM LÝ DO.');
  process.exit(1);
}
console.log(`\nCẢNH BÁO: ${mb(deadBytes)} MB art chết (ngưỡng chặn build: ${DEAD_MB_MAX} MB).`);
console.log(MOVE_HINT);
