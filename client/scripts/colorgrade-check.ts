// Ma trận chỉnh màu nền của `tools/colorgrade.mjs` phải KHỚP PixiJS — GAME_SPEC §13.10.
//
//   npm run check:colorgrade
//
// Vì sao cần: phép chỉnh màu nền nay nằm ở HAI chỗ — shader trộn (client, dùng Pixi) và
// asset-pipeline (Node, không có Pixi, nên chép tay phép toán của Pixi). Hai bên lệch nhau
// thì mặt đất và mấy cụm cỏ trên nó lệch màu — loại lỗi không ai gọi tên được, chỉ thấy
// "nhìn sai sai", và nó sẽ sống rất lâu.
//
// Bài này so từng số. Nó cũng là bài canh cho ngày PixiJS đổi công thức `saturate`/
// `contrast`: lúc đó build đỏ ngay, thay vì màu nền âm thầm trôi đi sau một lần nâng cấp.
// PixiJS chạm DOM ngay lúc NẠP MODULE (`getTestContext` dựng một canvas để dò độ chính xác
// shader), nên chạy nó trong Node cần một cái canvas giả. Trả `null` cho `getContext` là đủ:
// `ColorMatrixFilter` chỉ là phép toán 4×5, không cần ngữ cảnh đồ hoạ nào.
(globalThis as unknown as { document: unknown }).document = {
  createElement: (): unknown => ({ getContext: (): null => null, style: {} }),
};

const { ColorMatrixFilter } = await import('pixi.js');

import { GROUND_GRADE, groundMatrix } from '../../tools/colorgrade.mjs';

const f = new ColorMatrixFilter();
f.saturate(GROUND_GRADE.saturate);
f.brightness(GROUND_GRADE.brightness, true);
f.contrast(GROUND_GRADE.contrast, true);
const pixi = [...f.matrix];
const ours = groundMatrix();

let bad = 0;
for (let i = 0; i < 20; i++) {
  // 1e-6: cả hai bên đều là số thực 64 bit làm cùng một chuỗi phép nhân, nên sai khác chỉ
  // có thể là thứ tự cộng dồn. Lệch hơn ngần này là công thức đã khác nhau thật.
  if (Math.abs(pixi[i] - ours[i]) > 1e-6) {
    console.error(`  [${i}] Pixi ${pixi[i]} ≠ colorgrade.mjs ${ours[i]}`);
    bad++;
  }
}
if (bad > 0) {
  console.error(`\n✗ ma trận chỉnh màu nền LỆCH ${bad}/20 số giữa PixiJS và tools/colorgrade.mjs.`);
  console.error('  Sửa `tools/colorgrade.mjs` cho khớp, rồi dựng lại asset (ảnh decal/decor đã áp ma trận cũ).');
  process.exit(1);
}
console.log(`ma trận chỉnh màu nền: khớp PixiJS 20/20 số (saturate ${GROUND_GRADE.saturate}, brightness ${GROUND_GRADE.brightness}, contrast ${GROUND_GRADE.contrast})`);
