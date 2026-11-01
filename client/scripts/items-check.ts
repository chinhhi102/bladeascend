// npm run check:items — mọi loại vật phẩm phải vẽ được bằng ảnh thật, không rơi placeholder.
//
// Vì sao cần bài này: ngày 30/10 Bí Kíp hiện ra là một vòng tròn trắng trơn trong game. Ảnh
// `items/skill_scroll.png` đã có từ lâu — client tra nhầm sang `buffs/skill_scroll.png`.
// Sai kiểu này KHÔNG LÀM GÌ ỒN ÀO: không tìm thấy ảnh thì vẽ placeholder, nhìn y hệt lúc art
// chưa xong, nên nó sống sót qua mọi lần build xanh cho tới khi chủ dự án nhìn thấy bằng mắt.
//
// Bài kiểm đi theo ĐÚNG luật tra đường dẫn của `render/entities.ts`. Đổi luật bên đó mà quên
// đổi ở đây thì bài kiểm sai — nên hai chỗ đều ghi chú trỏ về nhau.
import { existsSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

import { ITEM_KINDS } from '../src/net/protocol.ts';

const ASSETS = join(dirname(fileURLToPath(import.meta.url)), '..', '..', 'assets', 'src');

/** artFor trả về các đường dẫn ảnh một loại vật phẩm cần; rỗng = cố ý không dùng ảnh. */
function artFor(kind: string): string[] {
  switch (kind) {
    case 'orb':
      return ['items/orb_small.png', 'items/orb_medium.png', 'items/orb_large.png'];
    case 'chest':
      return ['items/chest/idle_00.png'];
    case 'skill_scroll':
      // Bí Kíp đã bỏ (v6.25 §6.3): chỉ số còn trong ITEM_KINDS vì thứ tự là giao thức, nhưng
      // server không sinh nữa và art đã sang assets/library.
      return [];
    case 'sword':
      // Kiếm rơi dùng ảnh thật, nhưng qua một nhánh RIÊNG trong `drawItem` chứ không qua
      // `itemTexture` (entities.ts ~dòng 175). Đừng tin chú thích cũ trong `itemTexture`
      // nói "sword art is drawn inside the placeholder" — placeholder của nó chỉ vẽ một
      // vòng tròn trắng và một thanh nâu, không dán ảnh nào.
      return ['items/sword.png'];
    case 'bundle':
      // Chủ dự án duyệt 30/10, prompt ở ART_PROMPTS §8. Vẽ to hơn bong bóng kiếm đơn 20%.
      return ['items/bundle.png'];
    case 'tien_khi':
      // Hai mệnh giá LỚN (100, 500) có ảnh riêng từ spec v6.4 §3.4 mục 10 — chúng phải nhìn
      // ra ngay là món đáng vòng lại nhặt. Mệnh giá 1 / 2 / 10 **cố ý** vẽ bằng Graphics,
      // không phải thiếu art: trên màn hình chúng bé bằng hạt đậu.
      //
      // Trả [] ở đây (không chặn) vì thiếu ảnh viên lớn thì nó rơi về Graphics như cũ, vẫn
      // chơi được; `assets:report` mới là chỗ liệt kê chúng.
      return [];
    default:
      return [`buffs/${kind}.png`];
  }
}

const missing: { kind: string; path: string }[] = [];
const noArt: string[] = [];
for (const kind of ITEM_KINDS) {
  const paths = artFor(kind);
  if (paths.length === 0) {
    noArt.push(kind);
    continue;
  }
  for (const p of paths) {
    if (!existsSync(join(ASSETS, p))) {
      missing.push({ kind, path: p });
    }
  }
}

console.log(`${ITEM_KINDS.length} loại vật phẩm · ${noArt.length} cố ý vẽ bằng hình (${noArt.join(', ')})`);
if (missing.length > 0) {
  console.error('\nLỖI: những loại sau sẽ hiện ra placeholder trong game:');
  for (const m of missing) {
    console.error(`  ${m.kind.padEnd(14)} thiếu ${m.path}`);
  }
  console.error('\nHoặc sinh ảnh (`artgen run`), hoặc sửa luật tra đường dẫn trong render/entities.ts.');
  process.exit(1);
}
console.log('OK: mọi loại vật phẩm còn lại đều có ảnh thật.');
