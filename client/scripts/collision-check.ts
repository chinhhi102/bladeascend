/**
 * Đối chiếu va chạm client ↔ server (spec §8.2a).
 *
 * `cd server && go run ./cmd/collisionfixture -o /tmp/fixture.json` sinh 1 000 bước đi
 * ngẫu nhiên sát vật cản kèm vị trí **server** tính được. Script này phát lại đúng từng
 * bước qua mã **client** và báo lỗi nếu có bước nào lệch quá 0,5 đơn vị.
 *
 * Vì sao cần: client dự đoán vị trí của chính mình để nhân vật nhúc nhích ngay khi bấm
 * phím. Lệch một chút là server kéo về 20 lần mỗi giây và người chơi thấy rung. Hai bản
 * cài đặt ở hai ngôn ngữ khác nhau thì không cách nào bảo đảm bằng mắt được.
 *
 *   cd client && npm run check:collision -- /tmp/fixture.json
 */
import { readFileSync } from 'node:fs';

import { loadConfigFromJSON } from './load-config.ts';
import { moveStep } from '../src/net/movestep.ts';
import type { GameMap } from '../src/net/protocol.ts';

interface Case {
  x: number;
  y: number;
  vx: number;
  vy: number;
  a: number;
  sc: number;
  dt: number;
  mv: boolean;
  ex: number;
  ey: number;
  evx: number;
  evy: number;
}

const TOLERANCE = 0.5;

const path = process.argv[2];
if (!path) {
  console.error('dùng: npm run check:collision -- <fixture.json>');
  process.exit(2);
}

const fx = JSON.parse(readFileSync(path, 'utf8')) as { map: GameMap; cases: Case[] };
const cfg = loadConfigFromJSON();
const never = (): boolean => false;
const run = (c: Case) =>
  moveStep(cfg, fx.map, { x: c.x, y: c.y }, c.a, c.sc, c.dt, never, 1, { x: c.vx, y: c.vy }, c.mv);

let worst = 0;
let worstCase: Case | null = null;
let bad = 0;
for (const c of fx.cases) {
  const got = run(c);
  // So cả VỊ TRÍ lẫn VẬN TỐC: trên mặt băng vận tốc là trạng thái, lệch vận tốc thì vài
  // bước sau vị trí mới lệch — kiểm riêng vị trí sẽ bỏ sót đúng cái khó nhất.
  const off = Math.max(
    Math.hypot(got.pos.x - c.ex, got.pos.y - c.ey),
    Math.hypot(got.vel.x - c.evx, got.vel.y - c.evy) * c.dt,
  );
  if (off > worst) {
    worst = off;
    worstCase = c;
  }
  if (off > TOLERANCE) {
    bad++;
  }
}

console.log(`${fx.cases.length} bước đi · lệch lớn nhất ${worst.toFixed(4)} đơn vị · quá ngưỡng: ${bad}`);
if (bad > 0 && worstCase) {
  const got = run(worstCase);
  console.error(
    `ca lệch nhất: từ (${worstCase.x.toFixed(1)}, ${worstCase.y.toFixed(1)}) v=(${worstCase.vx.toFixed(1)}, ${worstCase.vy.toFixed(1)}) ` +
      `góc=${worstCase.a.toFixed(3)} cỡ=${worstCase.sc.toFixed(2)} đi=${worstCase.mv}\n` +
      `  server → vị trí (${worstCase.ex.toFixed(2)}, ${worstCase.ey.toFixed(2)}) vận tốc (${worstCase.evx.toFixed(2)}, ${worstCase.evy.toFixed(2)})\n` +
      `  client → vị trí (${got.pos.x.toFixed(2)}, ${got.pos.y.toFixed(2)}) vận tốc (${got.vel.x.toFixed(2)}, ${got.vel.y.toFixed(2)})`,
  );
  process.exit(1);
}
console.log(`ĐẠT: client khớp server trong ${TOLERANCE} đơn vị`);
