// npm run assets:budget — chặn build nếu vượt ngân sách dung lượng (GAME_SPEC §13.7 mục 6).
//
//   Gói chặn vào trận  ≤ 8 MB
//   Toàn bộ game       ≤ 80 MB
//
// "Gói chặn vào trận" KHÔNG phải nhóm `boot`, mà là `home + boot + khu xấu nhất + nhân vật
// xấu nhất` — đó mới là tập tệp người chơi phải chờ xong mới vào được trận. Lấy trường hợp
// xấu nhất vì server xếp khu theo "khu ít người nhất", người chơi không chọn được.
//
// Vượt thì in 20 tệp nặng nhất của đúng cái gói bị vượt, rồi thoát khác 0.
import { existsSync, readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const REPO = join(dirname(fileURLToPath(import.meta.url)), '..', '..');
const DIST = join(REPO, 'assets', 'dist');
const MANIFEST = join(DIST, 'manifest.json');

const BLOCKING_BUDGET = 8e6;
const TOTAL_BUDGET = 80e6;
const SHOW = 20;

interface Manifest {
  v: number;
  zones: string[];
  chars: string[];
  groups: Record<string, { bytes: number; files: string[] }>;
  files: Record<string, { url: string; bytes: number; groups: string[] }>;
}

if (!existsSync(MANIFEST)) {
  console.error(`Không có ${MANIFEST}.\nChạy trước:  node tools/asset-pipeline/index.mjs`);
  process.exit(1);
}
const man = JSON.parse(readFileSync(MANIFEST, 'utf8')) as Manifest;
if (man.v !== 2) {
  console.error(`manifest v${man.v} — script này đọc v2. Dựng lại: node tools/asset-pipeline/index.mjs --force`);
  process.exit(1);
}

const mb = (b: number): string => (b / 1e6).toFixed(2);
const bytesOfGroup = (g: string): number => man.groups[g]?.bytes ?? 0;

const worstZone = man.zones.map((z) => ({ name: `zone_${z}`, b: bytesOfGroup(`zone_${z}`) })).sort((a, b) => b.b - a.b)[0] ?? { name: '—', b: 0 };
const worstChar = man.chars.map((c) => ({ name: `char_${c}`, b: bytesOfGroup(`char_${c}`) })).sort((a, b) => b.b - a.b)[0] ?? { name: '—', b: 0 };
const blockingGroups = ['home', 'boot', worstZone.name, worstChar.name];
/** Hợp các nhóm, bỏ trùng: một tệp vật cản nằm trong nhiều nhóm khu, đếm hai lần là sai. */
const blockingFiles = [...new Set(blockingGroups.flatMap((g) => man.groups[g]?.files ?? []))];
const blocking = blockingFiles.reduce((s, f) => s + (man.files[f]?.bytes ?? 0), 0);
const total = Object.values(man.files).reduce((s, f) => s + f.bytes, 0);

function top(files: string[]): string[] {
  return files
    .map((f) => ({ f, b: man.files[f]?.bytes ?? 0 }))
    .sort((a, b) => b.b - a.b)
    .slice(0, SHOW)
    .map((r, i) => `${String(i + 1).padStart(3)}. ${(r.b / 1e3).toFixed(1).padStart(8)} KB  ${r.f}`);
}

console.log('Ngân sách dung lượng §13.7\n');
console.log(`  Chặn vào trận (${blockingGroups.join(' + ')})`);
console.log(`    ${mb(blocking)} MB / 8 MB     ${blocking <= BLOCKING_BUDGET ? 'ĐẠT' : `VƯỢT ${(blocking / BLOCKING_BUDGET).toFixed(2)}×`}`);
console.log(`  Toàn bộ game`);
console.log(`    ${mb(total)} MB / 80 MB    ${total <= TOTAL_BUDGET ? 'ĐẠT' : `VƯỢT ${(total / TOTAL_BUDGET).toFixed(2)}×`}`);

let bad = false;
if (blocking > BLOCKING_BUDGET) {
  bad = true;
  console.error(`\nVƯỢT ngân sách gói chặn vào trận. ${SHOW} tệp nặng nhất TRONG GÓI ĐÓ:`);
  console.error(top(blockingFiles).join('\n'));
}
if (total > TOTAL_BUDGET) {
  bad = true;
  console.error(`\nVƯỢT ngân sách toàn bộ game. ${SHOW} tệp nặng nhất:`);
  console.error(top(Object.keys(man.files)).join('\n'));
}
if (bad) {
  console.error(
    '\nCách hạ: đổi `maxPx` / `quality` của nhóm tương ứng trong tools/asset-pipeline/groups.mjs,\n' +
      'hoặc chuyển asset sang nhóm tải nền (`fx_high` / `audio`) nếu giây đầu trận không cần nó.',
  );
  process.exit(1);
}
console.log('\nCả hai ngân sách ĐẠT.');
