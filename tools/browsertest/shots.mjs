// Chụp 3 ảnh TRONG GAME THẬT cho phần dưới trang chủ (GAME_SPEC §19, §20).
//
//   node tools/browsertest/shots.mjs [http://127.0.0.1:9999]
//
// Ghi ra `client/public/shots/shot_{1,2,3}.webp` (960×540). `public/` chứ không
// `assets/src/`: ba ảnh này nằm trong HTML TĨNH mà máy tìm kiếm đọc được, nên đường dẫn
// phải ổn định — asset-pipeline gắn hash vào tên tệp thì thẻ `<img>` trong `index.html`
// thành 404 câm sau mỗi lần build. Cùng lý do với `brand/` (xem `tools/make_brandart.py`).
//
// Vì sao chụp từ game đang chạy chứ không vẽ tay: ảnh quảng cáo đẹp hơn thực tế là lời
// hứa mà game không giữ được. Chụp thật thì ảnh trên trang chủ luôn đúng bằng thứ người
// chơi sẽ thấy, và nó tự cũ đi cùng game chứ không đứng lại ở một bản đã lâu.
//
// Cần `Pillow` để nén WebP: `pip install Pillow` (hoặc `apt install python3-pil`).
//
// Server phải chạy ở chế độ cho dùng lệnh dựng cảnh (ADMIN_SPEC §1.1):
//
//   cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true
//
// `?debug=true` trên URL KHÔNG còn cấp quyền gì — quyền đến từ phiên admin, hoặc từ cửa
// `APP_ENV=dev` + gọi từ localhost dành riêng cho công cụ đo. Thiếu `APP_ENV=dev` thì
// phím dựng cảnh im lặng không làm gì và ảnh ra toàn người mới tay không; tệ hơn, không
// có dòng lỗi nào để biết. Vì vậy tool TỰ KIỂM ở dưới, xem `ensureStaged`.
import { chromium } from 'playwright';
import { execFileSync } from 'node:child_process';
import { mkdirSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9999';
const ROOT = fileURLToPath(new URL('../..', import.meta.url));
const OUT = join(ROOT, 'client/public/shots');
const TMP = join(tmpdir(), 'bladeascend-shots');
mkdirSync(OUT, { recursive: true });
mkdirSync(TMP, { recursive: true });

// Chụp ở 1920×1080 rồi thu về 960×540: thu nhỏ từ ảnh to cho cạnh chữ và lưỡi kiếm mịn
// hơn hẳn chụp thẳng ở 960 (cùng lý do `icon()` trong make_brandart.py vẽ to gấp 4).
const W = 1920, H = 1080;

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

const b = await chromium.launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'],
});
const ctx = await b.newContext({ viewport: { width: W, height: H }, deviceScaleFactor: 1 });
const p = await ctx.newPage();
p.on('console', (m) => { if (m.type() === 'error') console.log('  [console]', m.text().slice(0, 120)); });

// `?debug=true` để dùng được phím dựng cảnh. Bảng đo F3 mặc định TẮT nên nó không lọt vào
// ảnh; nếu sau này bảng đó bật sẵn thì phải bấm F3 tắt đi trước khi chụp.
await p.goto(`${TARGET}/?debug=true`, { waitUntil: 'domcontentloaded' });
await p.waitForSelector('input', { timeout: 60000 });
await p.fill('input', 'Đạo Hữu');
await p.getByText('Vào chơi', { exact: true }).click();
await p.waitForSelector('.room-board', { timeout: 60000 });

// Chờ art nạp xong. Bắt đầu chụp quá sớm thì ảnh đầy hình placeholder (vòng tròn + chữ) —
// đúng thứ không được xuất hiện trên trang chủ.
await sleep(9000);

/**
 * PHẢI CHƠI THẬT, không đứng im.
 *
 * Bản đầu của tệp này chỉ chờ rồi chụp ba lần. Nhân vật đứng nguyên chỗ hồi sinh nên sau
 * 30 giây vẫn là "Luyện Khí tầng 1 · Kiếm 1/10" — ba ảnh ra gần như trùng nhau, và không
 * ảnh nào có VÒNG KIẾM, tức là thiếu đúng hình ảnh đặc trưng của game. Ảnh chụp từ game
 * thật vẫn sai nếu cái "thật" đó là một nhân vật không làm gì.
 *
 * `roam` đi theo hình ngôi sao: mỗi chặng một hướng khác, đủ dài để băng qua chỗ có linh
 * khí. Đi vòng tròn đều thì nhân vật quay lại chỗ vừa nhặt sạch.
 */
const DIRS = ['KeyD', 'KeyS', 'KeyA', 'KeyW', 'KeyD', 'KeyW', 'KeyA', 'KeyS'];
let dirIdx = 0;

/**
 * CHẾT LÀ CHUYỆN BÌNH THƯỜNG — phải sống lại rồi đi tiếp.
 *
 * Lần chạy trước, nhân vật bị một bot chém ở giây thứ 44 và `shot_2` chụp trúng màn hình
 * TỔNG KẾT: một tấm ảnh "trong game" không có game nào trong đó. Phòng có bot đuổi theo
 * nên đi lang thang đủ lâu thì chắc chắn chết; chuyện cần xử là sống lại, không phải né.
 */
async function reviveIfDead() {
  const over = await p.$('.summary');
  if (!over) return false;
  await p.getByText('Chơi tiếp', { exact: true }).click();
  await p.waitForSelector('.room-board', { timeout: 30000 });
  await sleep(2500);
  return true;
}

async function roam(ms) {
  const until = Date.now() + ms;
  while (Date.now() < until) {
    if (await reviveIfDead()) continue;
    const key = DIRS[dirIdx++ % DIRS.length];
    await p.keyboard.down(key);
    await sleep(1400);
    await p.keyboard.up(key);
    await sleep(120);
  }
}

/**
 * Dựng sẵn cảnh giới và số kiếm bằng phím debug (`DEBUG_KEYS` trong `main.ts`).
 *
 * Vì sao cần: một kịch bản đi lang thang KHÔNG bao giờ mạnh lên. Lần chạy trước nó loanh
 * quanh 84 giây, chết hai lần, rớt bậc, và kết thúc ở "Luyện Khí tầng 2 · Kiếm 3/10 ·
 * hạng 59/59" — trong khi đầu bảng đã 1.009 điểm. Ba ảnh ra đều là người mới tay không,
 * tức trang chủ thiếu đúng VÒNG KIẾM, thứ mà cả cái tên game nói về.
 *
 * Vì sao vẫn trung thực: phím debug chỉ cấp tu vi và kiếm — những thứ người chơi thật đạt
 * được bằng cách chơi. Không có chế độ render riêng, không có hiệu ứng nào chỉ bật cho ảnh
 * quảng cáo. Ảnh vẫn là game thật vẽ một trạng thái thật, chỉ tới đó nhanh hơn.
 *
 * Cần server chạy với `-debug=true` và trang mở bằng `?debug=true`.
 */
/** swordCount đọc "Kiếm 3/10" trên HUD → 3. -1 nếu không đọc được. */
async function swordCount() {
  const txt = await p.evaluate(() => document.querySelector('.status')?.textContent ?? '');
  const m = txt.match(/Kiếm\s*(\d+)\s*\/\s*(\d+)/);
  return m ? Number(m[1]) : -1;
}

async function pressKeys(keys) {
  for (const [key, times] of keys) {
    for (let i = 0; i < times; i++) {
      await p.keyboard.press(key);
      await sleep(260);
    }
  }
  await sleep(900);
}

/**
 * stage dựng cảnh và TỰ KIỂM là lệnh có tới nơi thật.
 *
 * Cửa debug bị đóng thì lệnh đi vào hư vô, KHÔNG có lỗi nào nổi lên, và ảnh vẫn chụp được
 * — chỉ là chụp một nhân vật tay không. Đã mất hai lượt chạy vì đúng chuyện này, nên đếm
 * kiếm trước/sau là bằng chứng rẻ nhất.
 *
 * Nhưng "sau ít hơn trước" chưa đủ để kết luận cửa đóng: nhân vật CHẾT giữa chừng cũng
 * cho ra đúng dấu hiệu đó (lượt chạy 1/11 báo "kiếm 11 → 0" trong khi cửa vẫn mở). Nên
 * thử lại một lượt sau khi hồi sinh, rồi mới kết luận.
 */
async function stage(keys) {
  for (let attempt = 1; attempt <= 2; attempt++) {
    await reviveIfDead();
    const before = await swordCount();
    await pressKeys(keys);
    const after = await swordCount();
    if (after > before) return;
    if (attempt === 1) {
      console.log(`  dung canh khong len (kiem ${before} → ${after}) — co the vua chet, thu lai`);
    } else {
      throw new Error(
        `phim dung canh KHONG an sau 2 lan (kiem ${before} → ${after}). ` +
        'Server phai chay voi APP_ENV=dev va -debug=true, goi tu localhost. Xem ADMIN_SPEC §1.1.',
      );
    }
  }
}

const shots = [
  { n: 1, roamMs: 14000, keys: [], what: 'moi vao: mot kiem, nhat linh khi' },
  { n: 2, roamMs: 8000, keys: [['KeyR', 3], ['KeyK', 2]], what: 'vong kiem quanh nguoi' },
  { n: 3, roamMs: 8000, keys: [['KeyR', 3], ['KeyK', 4]], what: 'canh gioi cao + nhieu vong kiem' },
];

for (const s of shots) {
  if (s.keys.length) {
    await reviveIfDead();
    await stage(s.keys);
  }
  await roam(s.roamMs);
  // Dừng hẳn rồi mới chụp: đang giữ phím thì nhân vật nghiêng theo hướng chạy và vòng kiếm
  // bị kéo lệch về sau, trông như lỗi vẽ chứ không như một vòng kiếm.
  await sleep(600);
  // Chết đúng lúc sắp chụp: sống lại rồi đi thêm một đoạn ngắn cho có kiếm, chứ chụp ngay
  // sau hồi sinh thì lại ra một nhân vật tay không.
  if (await reviveIfDead()) {
    await roam(9000);
    await sleep(600);
  }
  if (await p.$('.summary')) {
    throw new Error(`shot_${s.n}: van dang o man tong ket sau khi hoi sinh — khong chup`);
  }
  // Giấu hai khối CHỈ CÓ Ở CHẾ ĐỘ TEST trước khi bấm máy.
  //
  // `?debug=true` bật dòng nhắc phím test ở mép trên và bảng đo F3 ở góc dưới-phải. Tôi
  // đã đoán "bảng đo mặc định tắt" và đoán sai — nó nằm chình ình trong ảnh đầu tiên.
  // Người chơi thật không bao giờ thấy hai khối này, nên chúng không được có trong ảnh
  // giới thiệu. Giấu bằng `style.display` thay vì bấm F3: bấm F3 chỉ tắt bảng đo, còn
  // dòng nhắc phím thì không có phím nào tắt.
  const hideDebugUI = () => p.evaluate(() => {
    for (const sel of ['.debug-help', '.perf']) {
      const e = document.querySelector(sel);
      if (e) e.style.display = 'none';
    }
  });
  // Chụp, rồi KIỂM BỐ CỤC: đường đi lang thang hay kết thúc sát biên bản đồ, và vùng
  // ngoài biên là một dải gần đen chiếm tới 40% chiều cao ảnh. Ảnh vẫn "đúng" nhưng nửa
  // dưới không có gì để nhìn. Đo tỉ lệ điểm ảnh tối ở dải đáy; quá ngưỡng thì đi tiếp
  // một đoạn rồi chụp lại, tối đa 3 lần.
  const png = join(TMP, `shot_${s.n}.png`);
  let dark = 100;
  for (let tryN = 1; tryN <= 3; tryN++) {
    // Giấu lại mỗi vòng: hồi sinh dựng lại HUD nên hai khối debug quay về.
    await hideDebugUI();
    await sleep(200);
    await p.screenshot({ path: png });
    dark = Number(execFileSync('python3', ['-c', `
import sys
from PIL import Image
im = Image.open(sys.argv[1]).convert('L')
w, h = im.size
band = im.crop((0, int(h * 0.75), w, h))
px = list(band.getdata())
print(round(100 * sum(1 for v in px if v < 40) / len(px)))
`, png]).toString().trim());
    if (dark <= 55) break;
    console.log(`  shot_${s.n}: day anh toi ${dark}% — di cho khac roi chup lai (lan ${tryN})`);
    await roam(6000);
    await sleep(600);
    await reviveIfDead();
  }
  if (dark > 55) {
    console.log(`  shot_${s.n}: VAN toi ${dark}% o day anh — giu nguyen, xem lai bang mat`);
  }
  const webp = join(OUT, `shot_${s.n}.webp`);
  // Thu nhỏ + nén bằng Pillow. `quality=82` là chỗ ảnh còn sạch mà tệp còn nhỏ; ba ảnh này
  // `loading="lazy"` nên chúng không chặn vào game, nhưng vẫn tính vào dung lượng trang.
  execFileSync('python3', ['-c', `
import sys
from PIL import Image
src, dst = sys.argv[1], sys.argv[2]
im = Image.open(src).convert('RGB').resize((960, 540), Image.LANCZOS)
im.save(dst, 'WEBP', quality=82, method=6)
`, png, webp], { stdio: 'inherit' });
  console.log(`  shot_${s.n}.webp  ${s.what}`);
}

rmSync(TMP, { recursive: true, force: true });
await ctx.close();
await b.close();
console.log(`\n3 anh trong ${OUT.replace(ROOT, '')} — Vite chep nguyen xi sang dist/.`);
