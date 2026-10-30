// npm run assets:size — đo dung lượng asset (GAME_SPEC §13.7 mục 1).
//
// In ba khối:
//   1. NGUỒN  `assets/src`  — tổng theo thư mục, số tệp; tách riêng phần không phát hành.
//   2. PHÁT HÀNH `assets/dist` — theo NHÓM TẢI, và gói nào chặn người chơi bao nhiêu byte.
//   3. Top 30 tệp nặng nhất trong bản phát hành.
//
// Vì sao phải tách nguồn và phát hành: `du -sh assets` cho 1,3 GB, nhưng gần hết là ảnh thô
// (`_raw/`), ảnh từng khung (`frames/`, `frames_norm/`) và tờ tham chiếu của bên vẽ —
// **người chơi không tải thứ nào trong đó**. Con số duy nhất đo được tiến bộ là "bao nhiêu
// byte chặn người chơi vào trận".
//
// `--json` cho máy đọc.
import { existsSync, readFileSync, readdirSync, statSync } from 'node:fs';
import { dirname, join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

const REPO = join(dirname(fileURLToPath(import.meta.url)), '..', '..');
const SRC = join(REPO, 'assets', 'src');
const DIST = join(REPO, 'assets', 'dist');

/** Thư mục/tệp nguồn KHÔNG phát hành — khớp `NEVER_SHIP` của tools/asset-pipeline/groups.mjs. */
const INTERMEDIATE = [/(^|\/)_raw\//, /(^|\/)frames\//, /(^|\/)frames_norm\//, /(^|\/)ref_sheet\.png$/, /^characters\/[^/]+\/(body|faces)\//];

interface FileRow {
  rel: string;
  bytes: number;
}

function walk(dir: string, base = dir, out: FileRow[] = []): FileRow[] {
  for (const e of readdirSync(dir, { withFileTypes: true })) {
    const p = join(dir, e.name);
    if (e.isDirectory()) {
      walk(p, base, out);
    } else if (e.isFile()) {
      out.push({ rel: relative(base, p).split('\\').join('/'), bytes: statSync(p).size });
    }
  }
  return out;
}

const mb = (b: number): string => (b / 1e6).toFixed(2);
const kb = (b: number): string => (b / 1e3).toFixed(1);

// ——— 1. Nguồn ———
const srcAll = walk(SRC);
const srcShipped = srcAll.filter((f) => !INTERMEDIATE.some((re) => re.test(f.rel)));
const srcInter = srcAll.filter((f) => INTERMEDIATE.some((re) => re.test(f.rel)));
const srcByDir = new Map<string, FileRow & { files: number }>();
for (const f of srcShipped) {
  const i = f.rel.indexOf('/');
  const d = i < 0 ? '(gốc)' : f.rel.slice(0, i);
  const row = srcByDir.get(d) ?? { rel: d, bytes: 0, files: 0 };
  row.bytes += f.bytes;
  row.files++;
  srcByDir.set(d, row);
}
const srcShippedBytes = srcShipped.reduce((s, f) => s + f.bytes, 0);
const srcInterBytes = srcInter.reduce((s, f) => s + f.bytes, 0);

// ——— 2. Phát hành ———
interface Manifest {
  v: number;
  zones: string[];
  chars: string[];
  groups: Record<string, { bytes: number; files: string[] }>;
  files: Record<string, { url: string; bytes: number; groups: string[] }>;
}
const manPath = join(DIST, 'manifest.json');
const man: Manifest | null = existsSync(manPath) ? (JSON.parse(readFileSync(manPath, 'utf8')) as Manifest) : null;

let blocking = 0;
let blockingGroups: string[] = [];
let distTotal = 0;
let distTop: FileRow[] = [];
if (man?.v === 2) {
  distTotal = Object.values(man.files).reduce((s, f) => s + f.bytes, 0);
  const worstZone = man.zones.map((z) => ({ n: `zone_${z}`, b: man.groups[`zone_${z}`]?.bytes ?? 0 })).sort((a, b) => b.b - a.b)[0];
  const worstChar = man.chars.map((c) => ({ n: `char_${c}`, b: man.groups[`char_${c}`]?.bytes ?? 0 })).sort((a, b) => b.b - a.b)[0];
  blockingGroups = ['home', 'boot', worstZone?.n ?? '', worstChar?.n ?? ''].filter(Boolean);
  const files = [...new Set(blockingGroups.flatMap((g) => man.groups[g]?.files ?? []))];
  blocking = files.reduce((s, f) => s + (man.files[f]?.bytes ?? 0), 0);
  distTop = Object.entries(man.files)
    .map(([rel, f]) => ({ rel, bytes: f.bytes }))
    .sort((a, b) => b.bytes - a.bytes)
    .slice(0, 30);
}

if (process.argv.includes('--json')) {
  console.log(
    JSON.stringify(
      {
        src: { bytes: srcShippedBytes, files: srcShipped.length, intermediate: { bytes: srcInterBytes, files: srcInter.length } },
        dist: man ? { bytes: distTotal, files: Object.keys(man.files).length, blocking, blockingGroups, groups: man.groups } : null,
      },
      null,
      2,
    ),
  );
  process.exit(0);
}

console.log(`## 1. NGUỒN — ${SRC}\n`);
console.log('Thư mục'.padEnd(16), 'Dung lượng'.padStart(11), 'Tệp'.padStart(6));
console.log('-'.repeat(16), '-'.repeat(11), '-'.repeat(6));
for (const r of [...srcByDir.values()].sort((a, b) => b.bytes - a.bytes)) {
  console.log(r.rel.padEnd(16), `${mb(r.bytes)} MB`.padStart(11), String(r.files).padStart(6));
}
console.log('TỔNG'.padEnd(16), `${mb(srcShippedBytes)} MB`.padStart(11), String(srcShipped.length).padStart(6));
console.log(`\n  Không phát hành (_raw, frames, frames_norm, ref_sheet, body, faces):`);
console.log(`  ${mb(srcInterBytes)} MB trong ${srcInter.length} tệp — người chơi KHÔNG tải.`);

if (!man) {
  console.log('\n## 2. PHÁT HÀNH — chưa có assets/dist');
  console.log('   Chạy:  node tools/asset-pipeline/index.mjs');
  process.exit(0);
}
if (man.v !== 2) {
  console.log(`\n## 2. PHÁT HÀNH — manifest v${man.v}, script này đọc v2.`);
  console.log('   Dựng lại:  node tools/asset-pipeline/index.mjs --force');
  process.exit(0);
}

console.log(`\n## 2. PHÁT HÀNH — ${DIST}  (theo nhóm tải §13.7)\n`);
console.log('Nhóm'.padEnd(22), 'Dung lượng'.padStart(11), 'Tệp'.padStart(6), '  Khi nào tải');
console.log('-'.repeat(22), '-'.repeat(11), '-'.repeat(6), ' ', '-'.repeat(28));
const WHEN: Record<string, string> = {
  home: 'trước khi hiện trang chủ',
  boot: 'khi bấm "Vào chơi"',
  fx_high: 'nền, trong trận',
  audio: 'nền, trong trận',
  debug: 'chỉ khi bấm phím J',
  legal: 'không bao giờ (giấy phép)',
};
const order = ['home', 'boot', ...man.zones.map((z) => `zone_${z}`), ...man.chars.map((c) => `char_${c}`), 'fx_high', 'audio', 'debug', 'legal'];
for (const g of order.filter((g) => man.groups[g])) {
  const when = WHEN[g] ?? (g.startsWith('zone_') ? 'khi vào khu đó' : 'nền (nhân vật mình: ngay)');
  console.log(g.padEnd(22), `${mb(man.groups[g].bytes)} MB`.padStart(11), String(man.groups[g].files.length).padStart(6), ' ', when);
}

console.log(`\n## Ngân sách §13.7\n`);
console.log(`  Chặn vào trận  ≤ 8 MB   → ${mb(blocking)} MB   ${blocking <= 8e6 ? 'ĐẠT' : `VƯỢT ${(blocking / 8e6).toFixed(2)}×`}`);
console.log(`    = ${blockingGroups.join(' + ')}  (trường hợp xấu nhất)`);
console.log(`  Toàn bộ game   ≤ 80 MB  → ${mb(distTotal)} MB   ${distTotal <= 80e6 ? 'ĐẠT' : `VƯỢT ${(distTotal / 80e6).toFixed(2)}×`}`);
console.log(`  Nguồn → phát hành: ${mb(srcShippedBytes)} MB → ${mb(distTotal)} MB  (−${(100 - (distTotal / srcShippedBytes) * 100).toFixed(1)}%)`);

console.log(`\n## 3. Top 30 tệp nặng nhất (bản phát hành)\n`);
distTop.forEach((f, i) => {
  const groups = man.files[f.rel]?.groups.join(',') ?? '';
  console.log(`${String(i + 1).padStart(3)}. ${kb(f.bytes).padStart(8)} KB  ${groups.padEnd(20)} ${f.rel}`);
});
