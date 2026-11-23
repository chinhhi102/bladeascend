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

// ── Phần 2: destroyDeep phải phân biệt context RIÊNG với context DÙNG CHUNG ─────────────
//
// Đã thử bài có răng: đổi `destroyDeep` thành luôn truyền `context: true` thì điều thứ ba
// đỏ ngay; bỏ vòng lặp con thì điều thứ nhất đỏ.
//
// Phần trên là lint (không ai gọi `destroy({children:true})` nữa). Phần này canh HÀNH VI
// của bản thay thế, vì chính chỗ này đã sai một lần: bản đầu truyền `context: true` cho mọi
// nút và làm vỡ hình trong 15 giây (texture về 0, FPS 16) — `entities.ts` cố ý dùng context
// dùng chung cho hàng trăm viên Linh khí cùng hình.
//
// Dùng nút GIẢ, không dùng Pixi: luật cần canh là "truyền cờ nào cho ai", không phải Pixi
// huỷ ra sao — và một bài test không cần WebGL thì chạy được trong mọi cổng.
const { destroyDeep } = await import('../src/render/destroydeep.ts');

interface Call { who: string; opts: { children: boolean; context?: boolean } }
const calls: Call[] = [];
const node = (who: string, owned: boolean, children: unknown[] = []): unknown => ({
  children,
  _ownedContext: owned ? {} : undefined,
  destroy: (opts: { children: boolean; context?: boolean }) => calls.push({ who, opts }),
});

const own = node('riêng', true);
const shared = node('dùng chung', false);
const parent = node('cha', false, [own, shared]);
destroyDeep(parent as Parameters<typeof destroyDeep>[0]);

let fails = 0;
const expect = (name: string, got: unknown, want: unknown): void => {
  const ok = JSON.stringify(got) === JSON.stringify(want);
  if (!ok) {
    fails++;
  }
  console.log(`${ok ? 'đạt ' : 'TRƯỢT'}  ${name}: ${JSON.stringify(got)}${ok ? '' : ` (mong ${JSON.stringify(want)})`}`);
};

expect('huỷ con trước, cha sau', calls.map((c) => c.who), ['riêng', 'dùng chung', 'cha']);
expect('nút có context RIÊNG được huỷ context', calls[0]?.opts, { children: false, context: true });
expect('nút dùng context CHUNG thì không', calls[1]?.opts, { children: false });
expect('không ai được nhận children: true (tự huỷ con, không nhờ Pixi)',
  calls.every((c) => c.opts.children === false), true);

// Không nổ khi nút rỗng / không có children / không có destroy.
destroyDeep(null);
destroyDeep(undefined);
destroyDeep({} as Parameters<typeof destroyDeep>[0]);
console.log('đạt   nút rỗng không làm nổ');

if (fails > 0) {
  console.error(`\ncheck:destroy TRƯỢT — ${fails} điều về hành vi destroyDeep`);
  process.exit(1);
}
