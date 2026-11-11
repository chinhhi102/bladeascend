// CI §6.6 (v6.48): "mọi số xuất hiện trong mô tả kỹ năng (mọi ngôn ngữ) phải lấy từ config;
// viết số cứng trong chuỗi i18n → fail".
//
// Vì sao cần một cổng riêng thay vì "nhớ sửa": ngày 11/11 spec đổi gần như cả mười kỹ năng,
// và mười dòng mô tả trong sáu tệp i18n vẫn kể về bản cũ — "hồi ngay 20% số kiếm", "bịt mắt
// 5 giây", "chạy nhanh gấp 3". Không có gì báo, vì chuỗi vẫn đúng ngữ pháp và vẫn dịch đủ.
// Cổng này bắt đúng kiểu hỏng đó: số nằm trong câu văn thì không ai sửa theo config được.
//
// Kiểm ba việc:
//   1. không chữ số nào trong `skillDesc` của bất kỳ ngôn ngữ nào;
//   2. mọi chỗ trống `{tên}` đều có giá trị trong `ui/skilldesc.ts`;
//   3. mọi ngôn ngữ dùng CÙNG bộ chỗ trống như tiếng Việt — thiếu một cái là bản dịch đó
//      im lặng bỏ mất một vế của luật chơi.
import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';

const I18N = new URL('../src/i18n/', import.meta.url).pathname;
const SRC = new URL('../src/ui/skilldesc.ts', import.meta.url).pathname;

/** Chữ số Ả Rập VÀ chữ số Thái (๐–๙) — bản dịch tiếng Thái có thể viết số theo kiểu của họ. */
const DIGIT = /[0-9๐-๙]/;

/** vars: tên chỗ trống của từng nhân vật, đọc thẳng từ bảng `VARS` trong `skilldesc.ts`. */
function declaredVars(): Map<string, Set<string>> {
  const src = readFileSync(SRC, 'utf8');
  const body = src.slice(src.indexOf('const VARS'), src.indexOf('/** skillDescVars'));
  const out = new Map<string, Set<string>>();
  for (const m of body.matchAll(/(\w+):\s*\(c\)\s*=>\s*\(\{([\s\S]*?)\}\),/g)) {
    const names = new Set<string>();
    for (const v of m[2].matchAll(/(?:^|[,{\s])(\w+)\s*:/g)) {
      names.add(v[1]);
    }
    out.set(m[1], names);
  }
  return out;
}

const vars = declaredVars();
const bad: string[] = [];
const viPlaceholders = new Map<string, Set<string>>();

for (const f of readdirSync(I18N).filter((n) => n.endsWith('.json'))) {
  const d = JSON.parse(readFileSync(join(I18N, f), 'utf8')) as { skillDesc?: Record<string, string> };
  for (const [char, text] of Object.entries(d.skillDesc ?? {})) {
    if (DIGIT.test(text)) {
      bad.push(`${f} · ${char}: còn chữ số viết cứng — "${text.slice(0, 70)}…"`);
    }
    const used = new Set([...text.matchAll(/\{(\w+)\}/g)].map((m) => m[1]));
    for (const name of used) {
      if (!vars.get(char)?.has(name)) {
        bad.push(`${f} · ${char}: chỗ trống {${name}} không có trong VARS của ui/skilldesc.ts`);
      }
    }
    if (f === 'vi.json') {
      viPlaceholders.set(char, used);
    } else {
      for (const name of viPlaceholders.get(char) ?? []) {
        if (!used.has(name)) {
          bad.push(`${f} · ${char}: thiếu chỗ trống {${name}} mà bản tiếng Việt có`);
        }
      }
    }
  }
}

if (bad.length > 0) {
  console.error(`\n✗ ${bad.length} lỗi mô tả kỹ năng (GAME_SPEC §6.6 — số phải lấy từ config):`);
  for (const b of bad) {
    console.error('  ' + b);
  }
  process.exit(1);
}
console.log(`✓ mô tả kỹ năng: ${vars.size} nhân vật, không có số viết cứng ở ngôn ngữ nào`);
