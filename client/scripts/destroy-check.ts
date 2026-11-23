// Huỷ sâu phải huỷ cả GraphicsContext — GAME_SPEC §13.9 ("không gì được tăng mãi").
//
//   npm run check:destroy
//
// ── Vì sao có cổng này ──────────────────────────────────────────────────────────────────
//
// `Container.destroy(options)` của Pixi truyền NGUYÊN options xuống từng con. Và
// `Graphics.destroy(options)` chỉ huỷ context của chính nó khi:
//
//     if (this._ownedContext && !options)        → huỷ
//     else if (options === true || options?.context === true) → huỷ
//
// Nên `destroy({ children: true })` — nghe như "huỷ sâu, huỷ hết" — lại là đúng cái nhánh
// KHÔNG huỷ context: `options` tồn tại nhưng `options.context` không phải `true`. Mỗi
// Graphics nằm trong cây bị huỷ kiểu đó để lại một `GraphicsContext` còn sống, và context
// đó giữ `GraphicsContextRenderData` → `Batcher` → geometry → bộ đệm đỉnh **~143 KB**.
//
// Đo thật ngày 23/11 (ảnh heap, kịch bản cố định §13.9): **192 context mới trong 20 phút**,
// kéo theo 209 Buffer và 30 MB. Không một chỉ số nào trong F3 nhúc nhích — chúng không nằm
// trong cây cảnh, không phải texture.
//
// Hai chữ `children: true` đọc như đã đủ, nên lỗi này sẽ quay lại nếu chỉ sửa 24 chỗ rồi đi
// tiếp. Cổng này là chỗ nó không quay lại được.
import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join } from 'node:path';

const SRC = new URL('../src/', import.meta.url).pathname;

function walk(dir: string, out: string[] = []): string[] {
  for (const name of readdirSync(dir)) {
    const p = join(dir, name);
    if (statSync(p).isDirectory()) {
      walk(p, out);
    } else if (p.endsWith('.ts')) {
      out.push(p);
    }
  }
  return out;
}

const bad: { file: string; line: number; text: string }[] = [];
for (const file of walk(SRC)) {
  const lines = readFileSync(file, 'utf8').split('\n');
  for (const [i, line] of lines.entries()) {
    // Chỉ xét lời gọi có `children: true`. `destroy(true)` và `destroy()` đều huỷ context
    // sẵn, còn `context: false` là lựa chọn CÓ Ý (xem `Scene.rebuild`: giữ lại cây cảnh để
    // dựng lại renderer) nên phải viết ra chữ `context` mới được bỏ qua.
    if (!/\.destroy\(\s*\{[^}]*children:\s*true/.test(line)) {
      continue;
    }
    if (/context:\s*(true|false)/.test(line)) {
      continue;
    }
    // `destroydeep.ts` nói VỀ lời gọi đó trong phần giải thích của chính nó.
    if (file.endsWith('destroydeep.ts')) {
      continue;
    }
    bad.push({ file: file.slice(SRC.length), line: i + 1, text: line.trim().slice(0, 100) });
  }
}

if (bad.length > 0) {
  console.error(`check:destroy TRƯỢT — ${bad.length} chỗ huỷ sâu mà không nói gì về \`context\`:\n`);
  for (const b of bad) {
    console.error(`  src/${b.file}:${b.line}  ${b.text}`);
  }
  console.error('\nDùng `destroyDeep(x)` (render/destroydeep.ts): nó huỷ context CỦA RIÊNG nút đó và');
  console.error('để nguyên context DÙNG CHUNG (entities.ts cố ý chia sẻ context cho hàng trăm viên).');
  console.error('Nếu thật sự cần tự quyết thì viết rõ `context: true` hoặc `context: false` kèm lý do.');
  process.exit(1);
}
console.log('check:destroy: mọi lời gọi huỷ sâu đều đã nói rõ về `context`.');
