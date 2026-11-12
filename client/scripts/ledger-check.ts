// Sổ tài nguyên + trần mảng lịch sử — GAME_SPEC §13.9.
//
//   npm run check:ledger
//
// Kiểm bốn điều:
//   1. `census` đếm đúng theo loại, đi hết cây (kể cả nút lồng nhiều tầng).
//   2. Phân loại theo `renderPipeId` — chuỗi Pixi tự gắn — chứ KHÔNG theo tên lớp. Tên lớp
//      bị minify ở bản phát hành, và một bảng số liệu chỉ đúng ở bản dev là bảng vô dụng.
//   3. `filters` đếm cả filter đơn lẫn mảng filter.
//   4. `capArray` cắt đúng phần CŨ NHẤT và giữ đúng trần.
import { capArray, census, censusFlat, censusLine, ledgerSet } from '../src/render/ledger.ts';

let fails = 0;
function expect(name: string, got: unknown, want: unknown): void {
  const ok = got === want;
  if (!ok) {
    fails++;
  }
  console.log(`${ok ? 'đạt ' : 'TRƯỢT'}  ${name}: ${String(got)}${ok ? '' : ` (mong ${String(want)})`}`);
}

/** node dựng một nút giả giống nút Pixi ở đúng những trường sổ đọc tới. */
const node = (id?: string, kids: unknown[] = [], filters?: unknown): unknown =>
  ({ renderPipeId: id, children: kids, filters });

// 1 + 2. Đếm theo loại, đi hết cây.
const tree = node(undefined, [
  node('sprite'),
  node('sprite'),
  node(undefined, [
    node('graphics'),
    node('text'),
    node(undefined, [node('bitmapText'), node('particleContainer')]),
  ]),
  // Lớp tự đặt tên khác (`renderPipeId` lạ) không được tính nhầm vào loại nào.
  node('cucGachLa'),
]);
const c = census(tree as Parameters<typeof census>[0]);
expect('tổng số nút', c.nodes, 10);
expect('sprite', c.byKind.sprite, 2);
expect('graphics', c.byKind.graphics, 1);
expect('text', c.byKind.text, 1);
expect('bitmapText', c.byKind.bitmapText, 1);
expect('particleContainer', c.byKind.particleContainer, 1);
expect('loại lạ không lọt vào bảng', Object.keys(c.byKind).includes('cucGachLa'), false);

// 3. Filter: mảng và đơn đều đếm.
const withFilters = node(undefined, [
  node('sprite', [], [{}, {}]),
  node('graphics', [], {}),
]);
expect('đếm filter', census(withFilters as Parameters<typeof census>[0]).filters, 3);

// Những thứ NGOÀI cây cảnh tự khai qua `ledgerSet`.
ledgerSet('pool', 12);
ledgerSet('poolText', 7);
const c2 = census(tree as Parameters<typeof census>[0]);
expect('pool khai ngoài cây', c2.extra.pool, 12);
expect('bảng phẳng có n_pool', censusFlat(c2).n_pool, 12);
expect('bảng phẳng có n_sprite', censusFlat(c2).n_sprite, 2);
expect('dòng F3 nhắc pool', censusLine(c2).includes('poolText 7'), true);
// Loại bằng 0 không được làm dòng F3 dài ra.
expect('dòng F3 bỏ loại bằng 0', censusLine(c2).includes('mesh'), false);

// 4. capArray: cắt phần CŨ NHẤT, giữ đúng trần.
const a = [1, 2, 3, 4, 5];
capArray(a, 3);
expect('trần 3 giữ 3 phần tử', a.length, 3);
expect('giữ phần MỚI nhất', a.join(','), '3,4,5');
capArray(a, 10);
expect('chưa chạm trần thì không đụng', a.join(','), '3,4,5');
capArray(a, 0);
expect('trần 0 xoá sạch', a.length, 0);

if (fails > 0) {
  console.error(`\n${fails} điều TRƯỢT`);
  process.exit(1);
}
console.log('\nSổ tài nguyên (§13.9): đạt hết');
