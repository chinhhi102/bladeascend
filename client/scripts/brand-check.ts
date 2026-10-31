// Chặn tên game cũ quay lại — GAME_SPEC §20, CHANGES_v6_16 mục 14.
//
//   npm run check:brand
//
// Tên game đã đổi sang BladeAscend / "Đạo Hữu Xin Dừng Bước". Lần đổi vừa rồi tìm thấy
// tên cũ ở **chín chỗ nằm trong sáu tệp** — tiêu đề tab, màn hình tải, hai trang quản trị,
// trang số liệu, dòng log khởi động. Không ai cố tình để sót; tên game chỉ đơn giản là thứ
// gõ vào theo phản xạ.
//
// Bài kiểm này chạy trong `npm run build`, nên một chuỗi "Kiếm Xoay" mới gõ vào sẽ chặn
// build ngay tại máy người viết, chứ không đợi ai đó nhìn thấy trên bản đã deploy.
//
// **Lịch sử được miễn.** `docs/`, `CHANGELOG`, thông điệp commit và chính tệp này cần nhắc
// tên cũ để câu chuyện còn đọc được. Cấm là cấm ở chỗ NGƯỜI CHƠI NHÌN THẤY.
import { readFileSync, readdirSync, statSync } from 'node:fs';
import { join, relative } from 'node:path';

const ROOT = new URL('../..', import.meta.url).pathname;

// Quét những cây thư mục ra tới mắt người chơi.
// `client/public` có mặt từ 1/11: `privacy.html` viết thẳng tên game (Vite chép nguyên xi
// các tệp trong `public/`, không thay token `%BRAND_NAME_VI%` như với `index.html`), nên
// đó đúng là loại chỗ mà lần đổi tên sau sẽ bỏ sót.
const SCAN = ['client/src', 'client/admin', 'client/index.html', 'client/public', 'server', 'config', 'maps', 'deploy', 'tools'];

const SKIP = new Set(['node_modules', 'dist', '.git', 'data', 'frames', '_raw', 'testdata']);

// Tên cũ dưới mọi dạng viết đã từng dùng.
const BANNED = [/Kiếm Xoay/u, /kiem[-_ ]?xoay/iu, /kiemxoaytutien/iu];

// Miễn trừ, kèm LÝ DO. Mỗi dòng ở đây là một món nợ — không phải một chỗ trốn.
//
// Món nợ cũ đã trả: đường dẫn module Go (`kiemxoay/server`, 82 tệp import) — đổi ngay
// trong lượt đổi tên này.
const ALLOW: Array<{ path: string; why: string }> = [
  {
    path: 'deploy/staging.sh',
    why: 'trỏ tới thư mục ~/kiemxoaytutien trên MÁY CHỦ staging. Đổi ở đây mà không đổi trên máy đó thì deploy ra một thư mục rỗng mới. Chủ dự án đổi tên thư mục trên máy chủ trước, rồi sửa dòng này.',
  },
];

const files: string[] = [];
const walk = (p: string): void => {
  const st = statSync(p, { throwIfNoEntry: false });
  if (!st) return;
  if (st.isFile()) {
    files.push(p);
    return;
  }
  for (const e of readdirSync(p, { withFileTypes: true })) {
    if (SKIP.has(e.name)) continue;
    walk(join(p, e.name));
  }
};
for (const s of SCAN) walk(join(ROOT, s));

// Chỉ đọc tệp chữ. Ảnh và font không chứa tên game dạng chuỗi tìm được, mà đọc hết 374 tệp
// art thì bài kiểm chậm tới mức người ta sẽ bỏ nó ra khỏi build.
const TEXT = /\.(ts|tsx|js|mjs|cjs|go|html|css|json|tmj|md|sh|mod|yaml|yml)$/;

type Hit = { file: string; line: number; text: string };
const hits: Hit[] = [];
let allowed = 0;

for (const f of files) {
  if (!TEXT.test(f)) continue;
  const rel = relative(ROOT, f);
  const exempt = ALLOW.find((a) => a.path === rel);
  const src = readFileSync(f, 'utf8');
  src.split('\n').forEach((text, i) => {
    if (!BANNED.some((re) => re.test(text))) return;
    if (exempt) {
      allowed++;
      return;
    }
    hits.push({ file: rel, line: i + 1, text: text.trim().slice(0, 110) });
  });
}

if (hits.length === 0) {
  const note = allowed > 0 ? ` (${allowed} dòng được miễn: ${ALLOW.map((a) => a.path).join(', ')})` : '';
  console.log(`OK: không còn tên game cũ ở chỗ người chơi nhìn thấy${note}.`);
  process.exit(0);
}

console.error(`Còn ${hits.length} chỗ dùng tên game CŨ. Tên đã chốt: BladeAscend / "Đạo Hữu Xin Dừng Bước".\n`);
for (const h of hits) console.error(`  ${h.file}:${h.line}\n    ${h.text}`);
console.error(
  '\nLấy tên từ `config.brand` — đừng gõ thẳng:\n' +
    '  TypeScript: `import { BRAND } from "./ui/strings"` → `BRAND.name_vi`\n' +
    '  HTML (client): token `%BRAND_NAME_VI%`, plugin `kx-brand` thay lúc build\n' +
    '  HTML (server): token `%BRAND_NAME_VI%` + `s.brandHTML(...)`\n' +
    '  Go:         `cfg.Brand.Name`\n',
);
process.exit(1);
