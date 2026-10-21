// npm run assets:report — lists which assets from docs/ART_PROMPTS.md exist in assets/ and which are
// still drawn as placeholders. Missing images show the artgen job / docs/ART_JOBS.md entry to make.
import { existsSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

import { catalog, type CatalogEntry } from '../src/render/catalog.ts';

const ASSETS = join(dirname(fileURLToPath(import.meta.url)), '..', '..', 'assets');
const AUDIO_EXTS = ['.ogg', '.mp3', '.wav'];

function exists(rel: string): boolean {
  if (rel.startsWith('audio/')) {
    const base = rel.replace(/\.[a-z0-9]+$/, '');
    return AUDIO_EXTS.some((e) => existsSync(join(ASSETS, base + e)));
  }
  return existsSync(join(ASSETS, rel));
}

interface Row {
  e: CatalogEntry;
  have: number;
}

const rows: Row[] = catalog().map((e) => ({ e, have: e.files.filter(exists).length }));
const groups = [...new Set(rows.map((r) => r.e.group))];
const showAll = process.argv.includes('--all');

console.log(`Thư mục asset: ${ASSETS}\n`);
console.log('Nhóm'.padEnd(12), 'Đủ'.padStart(5), 'Thiếu'.padStart(7));
for (const g of groups) {
  const rs = rows.filter((r) => r.e.group === g);
  const full = rs.filter((r) => r.have === r.e.files.length).length;
  console.log(g.padEnd(12), String(full).padStart(5), String(rs.length - full).padStart(7));
}

const jobs = new Set<string>();
for (const g of groups) {
  const missing = rows.filter((r) => r.e.group === g && (showAll || r.have < r.e.files.length));
  if (!missing.length) {
    continue;
  }
  console.log(`\n## ${g}`);
  for (const r of missing) {
    const status = r.have === 0 ? 'THIẾU   ' : r.have < r.e.files.length ? `${r.have}/${r.e.files.length} frame` : 'đủ      ';
    const job = r.e.job ? `→ job ${r.e.job}` : '(âm thanh: tools/make_sfx.py hoặc tự thả file)';
    console.log(`  ${status.padEnd(10)} ${r.e.label.padEnd(30)} ${job}`);
    if (r.have < r.e.files.length && r.e.job) {
      jobs.add(r.e.job);
    }
  }
}

const total = rows.length;
const done = rows.filter((r) => r.have === r.e.files.length).length;
console.log(`\nTổng: ${done}/${total} asset đủ. Thiếu thì game tự vẽ placeholder.`);
if (jobs.size) {
  console.log(`Cần sinh ${jobs.size} ảnh — prompt trong docs/ART_JOBS.md (tạo lại: cd server && go run ./cmd/artgen prompts).`);
  console.log('Lưu ý: anim thiếu vài frame (ví dụ 5/8) vẫn chạy — model vẽ ít frame hơn yêu cầu.');
}
