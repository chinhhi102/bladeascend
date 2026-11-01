// CI đa ngôn ngữ (GAME_SPEC §19): "Không chuỗi cứng nào trong code".
//
// 1. CHẶN BUILD: chuỗi có chữ tiếng Việt có dấu trong `src/**/*.ts` mà không nằm trong
//    `src/i18n/`. Đọc bằng TypeScript API nên comment (viết tiếng Việt rất nhiều) không
//    bị tính — chỉ chuỗi thật sự tới tay người chơi.
// 2. CẢNH BÁO (không chặn): khoá có trong `vi.json` mà ngôn ngữ khác thiếu → game rơi về
//    English rồi Tiếng Việt cho khoá đó.
//
// Ngoại lệ có chủ đích (chỉ dev đọc, người chơi không bao giờ thấy):
//  - đối số của `console.*` và `new Error(...)` — tự động;
//  - dòng có comment `i18n-ok`, hoặc cả file có `i18n-ok-file: <lý do>` (công cụ dev,
//    lớp debug sau `?debug=true`).
import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join, relative } from 'node:path';
import ts from 'typescript';

const SRC = new URL('../src/', import.meta.url).pathname;
const I18N = join(SRC, 'i18n');
const VIET = /[àáạảãâầấậẩẫăằắặẳẵèéẹẻẽêềếệểễìíịỉĩòóọỏõôồốộổỗơờớợởỡùúụủũưừứựửữỳýỵỷỹđ]/i;

function walk(dir: string): string[] {
  return readdirSync(dir).flatMap((f) => {
    const p = join(dir, f);
    if (statSync(p).isDirectory()) {
      return p === I18N ? [] : walk(p);
    }
    return p.endsWith('.ts') && !p.endsWith('.d.ts') ? [p] : [];
  });
}

/** Chuỗi nằm trong `console.x(...)` hoặc `new Error(...)`: chỉ dev đọc trong DevTools. */
function devOnly(n: ts.Node): boolean {
  for (let p: ts.Node | undefined = n.parent; p; p = p.parent) {
    if (ts.isNewExpression(p) && /Error$/.test(p.expression.getText())) {
      return true;
    }
    if (ts.isCallExpression(p) && /^console\./.test(p.expression.getText())) {
      return true;
    }
    if (ts.isBlock(p) || ts.isSourceFile(p)) {
      return false;
    }
  }
  return false;
}

const bad: string[] = [];
for (const file of walk(SRC)) {
  const text = readFileSync(file, 'utf8');
  const sf = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
  const lines = text.split('\n');
  if (text.includes('i18n-ok-file:')) {
    continue;
  }
  const visit = (n: ts.Node): void => {
    if (ts.isStringLiteralLike(n) || ts.isTemplateHead(n) || ts.isTemplateMiddle(n) || ts.isTemplateTail(n)) {
      const s = (n as ts.LiteralLikeNode).text;
      if (VIET.test(s)) {
        const line = sf.getLineAndCharacterOfPosition(n.getStart()).line;
        if (!lines[line].includes('i18n-ok') && !devOnly(n)) {
          bad.push(`${relative(SRC, file)}:${line + 1}  ${JSON.stringify(s.slice(0, 60))}`);
        }
      }
    }
    ts.forEachChild(n, visit);
  };
  visit(sf);
}

// ── Khoá thiếu ở các ngôn ngữ khác ───────────────────────────────────────────────────
type J = Record<string, unknown>;
function keys(o: unknown, pre = ''): string[] {
  if (typeof o !== 'object' || o === null || Array.isArray(o)) {
    return [pre];
  }
  return Object.entries(o as J).flatMap(([k, v]) => keys(v, pre ? `${pre}.${k}` : k));
}
function get(o: unknown, path: string): unknown {
  return path.split('.').reduce<unknown>((a, k) => (a && typeof a === 'object' ? (a as J)[k] : undefined), o);
}
const vi = JSON.parse(readFileSync(join(I18N, 'vi.json'), 'utf8'));
const viKeys = keys(vi).filter((k) => !k.startsWith('_meta'));
for (const f of readdirSync(I18N).filter((f) => f.endsWith('.json') && f !== 'vi.json')) {
  const d = JSON.parse(readFileSync(join(I18N, f), 'utf8'));
  const miss = viKeys.filter((k) => {
    const v = get(d, k);
    const empty = v === undefined || v === '' || (Array.isArray(v) && v.length === 0);
    // Khoá mà chính tiếng Việt cũng để trống (dùng config) thì không bắt dịch.
    const viV = get(vi, k);
    return empty && !(viV === '' || (Array.isArray(viV) && viV.length === 0));
  });
  const draft = d._draft ? ' (bản nháp — cần người bản ngữ duyệt, xem i18n/NOTES.md)' : '';
  console.log(`i18n ${f}: ${viKeys.length - miss.length}/${viKeys.length} khoá${draft}`);
  if (miss.length) {
    console.log(`  ⚠ thiếu ${miss.length} khoá (rơi về English → Tiếng Việt): ${miss.slice(0, 8).join(', ')}${miss.length > 8 ? '…' : ''}`);
  }
}

if (bad.length) {
  console.error(`\n✗ ${bad.length} chuỗi tiếng Việt viết cứng trong code — chuyển vào src/i18n/vi.json (GAME_SPEC §19):`);
  for (const b of bad) {
    console.error('  ' + b);
  }
  process.exit(1);
}
console.log('✓ không còn chuỗi cứng tiếng Việt ngoài src/i18n/');
