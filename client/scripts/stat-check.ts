// `npm run check:stat`: hàm stat() phía client phải ra đúng bộ số mẫu mà server cũng chạy
// (server/internal/game/testdata/stat_vectors.json, spec v6.32 §4.0). Lệch là client dự đoán
// tốc chạy / góc vòng kiếm khác server → bị kéo về mỗi tick.
import { readFileSync } from 'node:fs';
import { stat } from '../src/net/stat.ts';

interface Vector { name: string; base: number; floor: number; cap: number; terms: number[]; want: number }

const file = new URL('../../server/internal/game/testdata/stat_vectors.json', import.meta.url);
const { vectors } = JSON.parse(readFileSync(file, 'utf8')) as { vectors: Vector[] };
let bad = 0;
for (const v of vectors) {
  const got = stat(v.base, { floor: v.floor, cap: v.cap }, ...v.terms);
  if (Math.abs(got - v.want) > 1e-6) {
    bad += 1;
    console.error(`✗ ${v.name}: ${got}, muốn ${v.want}`);
  }
}
if (bad > 0) {
  process.exit(1);
}
console.log(`✓ stat(): ${vectors.length}/${vectors.length} số mẫu khớp server`);
