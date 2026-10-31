// Tab CŨ gặp server MỚI phải thấy "Đã có bản mới — bấm để tải lại", không lỗi đỏ, không kẹt.
//
//   node tools/browsertest/oldclient.mjs [url] [thư-mục-client-cũ]
//
// Vì sao cần bài này: Protocol v2 đổi hẳn hình dạng gói tin. Một người mở tab từ hôm qua,
// sau khi server deploy lại, sẽ nhận những byte mà client của họ không hiểu. Nếu không có
// cửa kiểm phiên bản thì họ **không thấy lỗi gì cả** — chỉ là nhân vật đứng im hoặc màn
// hình trắng, và họ sẽ nghĩ game hỏng.
//
// ── Cách giả lập "tab cũ" ───────────────────────────────────────────────────────────────
//
// Phục vụ BUNDLE CŨ THẬT từ kho `deploy-archive`, chặn ở tầng HTTP của trình duyệt, trong
// khi WebSocket vẫn nối tới server MỚI. Đó đúng là tình huống thật: mã trong tab là mã của
// hôm qua, server thì đã đổi.
//
// Bản đầu tôi sửa BYTE của gói `welcome` để giả lệch phiên bản. Nó hỏng: sửa byte giữa một
// gói msgpack làm cả gói không giải mã được, nên client ném `RangeError` từ bộ giải msgpack
// và không bao giờ chạy tới cửa kiểm phiên bản. Bài đo khi đó kiểm nhầm thứ khác.
import { existsSync, readdirSync, readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { webkit } from 'playwright';

const REPO = join(dirname(fileURLToPath(import.meta.url)), '..', '..');
// Tên khác `URL`: hằng số đó che mất lớp `URL` dựng sẵn, và `new URL(...)` trong bộ
// chặn request ném ngay lập tức.
const TARGET = process.argv[2] ?? 'http://localhost:9999';
const ARCHIVE = join(REPO, 'server', 'data', 'deploy-archive');

/** oldClientDir: bản deploy cũ NHẤT còn trong kho — chắc chắn là bundle trước Protocol v2. */
function oldClientDir() {
  if (process.argv[3]) return process.argv[3];
  if (!existsSync(ARCHIVE)) return null;
  const builds = readdirSync(ARCHIVE).sort();
  // Bỏ bản mới nhất (đang chạy); lấy bản cũ nhất còn lại.
  for (const b of builds.slice(0, -1)) {
    const d = join(ARCHIVE, b, 'client');
    if (existsSync(join(d, 'index.html'))) return d;
  }
  return null;
}

const dir = oldClientDir();
if (!dir) {
  console.log('Kho chưa có bản cũ nào để giả lập tab cũ — bỏ qua bài kiểm này.');
  console.log('(Chạy tools/deploy.sh vài lần là kho có bản để lùi và để kiểm.)');
  process.exit(0);
}
console.log(`Tab cũ = ${dir.replace(REPO + '/', '')}`);

let failures = 0;
const check = (ok, label) => {
  console.log(`  ${ok ? '✓' : '✗'} ${label}`);
  if (!ok) failures++;
};

const br = await webkit.launch({ headless: true });
const ctx = await br.newContext({ viewport: { width: 1100, height: 760 } });
const page = await ctx.newPage();
const errors = [];
page.on('pageerror', (e) => errors.push(String(e)));

// Chặn HTML và bundle JS/CSS → lấy từ kho. Mọi thứ khác (API, asset, WebSocket) vẫn đi
// thẳng tới server mới.
const types = { '.html': 'text/html', '.js': 'text/javascript', '.css': 'text/css' };
await page.route('**/*', async (route) => {
  const u = new URL(route.request().url());
  const rel = u.pathname === '/' ? 'index.html' : u.pathname.replace(/^\//, '');
  const file = join(dir, rel);
  const ext = rel.slice(rel.lastIndexOf('.'));
  if (types[ext] && existsSync(file)) {
    await route.fulfill({ status: 200, contentType: types[ext], body: readFileSync(file) });
    return;
  }
  await route.continue();
});

await page.goto(TARGET, { waitUntil: 'domcontentloaded' });
await page.waitForSelector('input', { timeout: 60000 });
await page.fill('input', 'TabCu');
await page.getByText('Vào chơi', { exact: true }).click();
await page.waitForTimeout(10000);

const body = (await page.locator('body').textContent()) ?? '';
check(/bản mới|cập nhật/i.test(body), `hiện thông báo có bản mới ("${body.replace(/\s+/g, ' ').slice(0, 90).trim()}…")`);
// Không được có lỗi đỏ: một thông báo tử tế khác hẳn một ngoại lệ chưa bắt.
const real = errors.filter((e) => !/favicon|ResizeObserver/.test(e));
check(real.length === 0, `không ném lỗi nào${real.length ? ': ' + real[0].slice(0, 110) : ''}`);
// Thông báo phải HIỆN LÊN TRÊN, không nằm khuất dưới màn hình chờ.
//
// Kiểm bằng phần tử đang hiển thị, không bằng cách dò chữ trong `body`: trang luôn còn
// chữ của lớp tải nền ("⬇ 87%") nằm dưới, và dò chữ thì báo đỏ nhầm vì chuyện đó.
const panel = page.locator('.screen.center .panel');
check(await panel.count() > 0 && await panel.first().isVisible(),
  'thông báo hiện lên trên, không khuất dưới màn hình chờ');
// Và client ĐÃ NGẮT kết nối, không ngồi thử lại trong im lặng.
check(!/Đang tìm phòng/i.test((await panel.first().textContent()) ?? ''),
  'không còn cố vào phòng nữa');

await br.close();
console.log(failures === 0
  ? '\nĐẠT: tab cũ gặp server mới thì được báo rõ, không kẹt và không lỗi.'
  : `\nTRƯỢT: ${failures} mục.`);
process.exit(failures === 0 ? 0 : 1);
