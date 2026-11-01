// Chụp ảnh duyệt 10 kỹ năng (spec v6.36 §6.5, CHANGES_v6_36 mục 9) → docs/review/skills/.
//
//   cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true
//   node tools/browsertest/skillshots.mjs [http://127.0.0.1:9998] [a_hon,be_ti,...]
//
// Ba trình duyệt cùng phòng: NGƯỜI XEM (đứng phía trên, thấy cả cảnh), NGƯỜI DÙNG (đổi nhân
// vật lần lượt) và NGƯỜI BỊ TRÚNG (đứng bên phải người dùng 3,5 U). Mỗi kỹ năng ra 2 ảnh:
//   <nhân vật>_xem.jpg       góc nhìn người đứng xem
//   <nhân vật>_trung.jpg     màn hình người bị trúng (HUD icon + giây, mũi tên nếu có)
//                            — kỹ năng tự thân thì là màn hình NGƯỜI DÙNG (<nv>_dung.jpg)
// Tên thử là tên đội làm game (CLAUDE.md quy tắc 15), không phải người chơi thật.
import { chromium } from 'playwright';
import { mkdirSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
const ONLY = process.argv[3]?.split(',').filter(Boolean);
const ROOT = fileURLToPath(new URL('../..', import.meta.url));
const OUT = join(ROOT, 'docs/review/skills');
mkdirSync(OUT, { recursive: true });
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

// Kỹ năng tự thân: ảnh thứ hai là màn hình người dùng. `wait`: chờ bao lâu sau khi bấm mới
// chụp (Ru Ngủ hát 0,5 s rồi mới ngủ; Rụt Cổ chụp lúc đang rụt).
const PLAN = {
  a_hon: { self: true, wait: 350 },
  ba_tam_loa: { wait: 180 },
  ong_ba_vong: { wait: 900 },
  dai_ca_dep: { self: true, wait: 700 },
  map_tra_sua: { wait: 200 },
  co_ba_ganh: { wait: 150 },
  su_phu_mi: { wait: 220 },
  be_ti: { self: true, wait: 600 },
  shipper: { self: true, wait: 500 },
  cu_rua: { self: true, wait: 900 },
};

const cfg = await (await fetch(`${TARGET}/api/config`)).json();
const chars = cfg.characters.filter((c) => !ONLY || ONLY.includes(c));

const b = await chromium.launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'],
});

async function open(name, char) {
  const ctx = await b.newContext({ viewport: { width: 1280, height: 720 }, deviceScaleFactor: 1, locale: 'vi-VN' });
  const p = await ctx.newPage();
  p.on('pageerror', (e) => console.log(`  [${name}] pageerror`, String(e).slice(0, 160)));
  await join_(p, name, char);
  return p;
}

async function join_(p, name, char) {
  await p.goto(`${TARGET}/?debug=true`, { waitUntil: 'domcontentloaded' });
  // Vừa chết thì trang có thể vẫn hiện Tổng kết — về màn chọn nhân vật bằng nút của nó.
  await p.waitForSelector('input, .summary', { timeout: 30000 }).catch(() => {});
  if (await p.$('.summary')) {
    await p.getByText('Đổi nhân vật', { exact: true }).click();
  }
  await p.waitForSelector('input', { timeout: 30000 }).catch(async (e) => {
    await p.screenshot({ path: join(tmpdir(), 'skillshots-fail.png') });
    throw e;
  });
  await p.fill('input', name);
  await p.locator('.char-card').nth(cfg.characters.indexOf(char)).click();
  await p.getByText('Vào chơi', { exact: true }).click();
  await p.waitForSelector('.room-board', { timeout: 60000 });
}

const hideDebug = (p) => p.evaluate(() => {
  for (const sel of ['.debug-help', '.perf', '.guide']) {
    const e = document.querySelector(sel);
    if (e) e.style.display = 'none';
  }
});

async function revive(p) {
  if (await p.$('.summary')) {
    await p.getByText('Chơi tiếp', { exact: true }).click();
    await p.waitForSelector('.room-board', { timeout: 30000 });
    await sleep(2000);
  }
}

// Thứ tự vào phòng quyết định chỗ đứng khi `summon` (người vào trước đứng bên phải).
const victim = await open('KiemTraTrung', 'a_hon');
const watcher = await open('KiemTraXem', 'a_hon');
let caster = null;

for (const ch of chars) {
  const plan = PLAN[ch] ?? { wait: 300 };
  if (!caster) {
    caster = await open('KiemTraDung', ch);
  } else {
    await join_(caster, 'KiemTraDung', ch);
  }
  await sleep(6000); // chờ atlas nhân vật về
  for (const p of [watcher, victim, caster]) {
    await revive(p);
  }
  // Thêm vài thanh kiếm cho có vòng kiếm (thấy được viền / trân châu trên vòng).
  await victim.keyboard.press('KeyK'); // nạn nhân có thể vừa bị bot chém mất kiếm
  // Cùng cảnh giới với nạn nhân: người dùng mà cao hơn thì vòng kiếm chém chết nạn nhân
  // ngay trong ảnh (lượt chụp đầu Bà Tám ra màn hình Tổng kết).
  await caster.keyboard.press('KeyK');
  await sleep(200);
  await caster.keyboard.press('KeyG');
  await caster.keyboard.press('KeyU');
  await sleep(800); // chuyển khu xảy ra ở ranh giới tick, người được kéo phải nhận Welcome mới
  // Quay mặt sang phải (nạn nhân đứng bên phải) rồi dùng kỹ năng.
  await caster.keyboard.down('KeyD');
  await sleep(120);
  await caster.keyboard.up('KeyD');
  await caster.keyboard.press('KeyU');
  await sleep(250);
  await caster.keyboard.press('Space');
  await sleep(plan.wait);
  for (const p of [watcher, victim, caster]) {
    await hideDebug(p);
  }
  const second = plan.self ? caster : victim;
  // JPEG 85: 20 ảnh PNG là 20 MB trong repo, JPEG ~4 MB mà vẫn đủ nét để duyệt.
  const shot = { type: 'jpeg', quality: 85 };
  await Promise.all([
    watcher.screenshot({ ...shot, path: join(OUT, `${ch}_xem.jpg`) }),
    second.screenshot({ ...shot, path: join(OUT, `${ch}_${plan.self ? 'dung' : 'trung'}.jpg`) }),
  ]);
  console.log(`  ${ch}: xong`);
  await sleep(1500);
}

await b.close();
console.log(`ảnh trong ${OUT.replace(ROOT, '')}`);
