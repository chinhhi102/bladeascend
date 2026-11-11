// Chụp ảnh BONG BÓNG MẶT (GAME_SPEC §12.5 v6.38) — PC + iPhone.
//
//   cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true
//   node tools/browsertest/emoteshots.mjs http://127.0.0.1:9998
//
// Bong bóng chỉ hiện 1,5 giây khi tâm trạng ĐỔI (đuổi → mặt cười nham hiểm, chạy trốn →
// mặt hoảng, bị ngủ/mù → mặt hoảng, hạ được người → mặt cười). Đứng im thì không bao giờ
// chụp được, nên bài này thả bot cạnh mình bằng phím debug `B`: bot lao vào đuổi, và cả
// hai bên đều đổi tâm trạng.
//
// Chụp một LOẠT ảnh rồi chọn, không chụp một phát: bong bóng sống 1,5 s trong khi lượt
// chụp mất ~0,3 s, nên canh đúng khoảnh khắc bằng cách hẹn giờ là trò may rủi.
import { devices } from 'playwright';
import { launch } from './browser.mjs';
import { mkdirSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
const ROOT = fileURLToPath(new URL('../..', import.meta.url));
const OUT = join(ROOT, 'docs/review/emote');
mkdirSync(OUT, { recursive: true });
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

const browser = await launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'],
});

async function shoot(label, ctxOpts, want) {
  const ctx = await browser.newContext(ctxOpts);
  const p = await ctx.newPage();
  const loi = [];
  p.on('pageerror', (e) => loi.push(String(e).slice(0, 120)));

  // Bấm máy ĐÚNG LÚC bong bóng hiện.
  //
  // Bong bóng sống 1,5 s và chỉ bật khi tâm trạng đổi; lượt chụp đầu ngày 05/11 lấy 24 ảnh
  // theo nhịp cố định và dính 0 cái. Nay `emote.ts` báo một dòng `[EMOTE]` ra console khi
  // `?debug=true`, và ta chụp ngay khi nghe thấy — một lượt chụp mất ~0,3 s nên luôn nằm
  // gọn trong cửa sổ 1,5 s.
  let shot = 0;
  let busy = false;
  const faces = [];
  p.on('console', (m) => {
    const t = m.text();
    if (!t.startsWith('[EMOTE]') || shot >= want || busy) return;
    busy = true;
    const n = ++shot;
    faces.push(t.replace('[EMOTE] ', ''));
    p.screenshot({ path: join(OUT, `${label}_${String(n).padStart(2, '0')}.png`) })
      .catch(() => undefined)
      .finally(() => { busy = false; });
  });

  await p.goto(`${TARGET}/?lang=vi&debug=true`, { waitUntil: 'domcontentloaded' });
  await p.waitForSelector('.home .btn.primary', { timeout: 60000 });
  await p.fill('input', 'Emote');
  await p.locator('.home .btn.primary').click();
  await p.waitForSelector('.room-board', { timeout: 60000 });
  await sleep(9000); // chờ art nạp xong, nếu không bong bóng ra ô trống

  // Lên vài cảnh giới cho vòng kiếm dày — ảnh duyệt nên giống lúc chơi thật.
  for (let i = 0; i < 3; i++) {
    await p.keyboard.press('KeyR');
    await sleep(250);
  }
  await p.keyboard.press('KeyK');

  // Thả bot liên tục và đi loanh quanh: bot lao vào đuổi (→ mặt cười nham hiểm), mình bỏ
  // chạy (→ mặt hoảng). Đứng im thì tâm trạng không ai đổi và không bao giờ có ảnh.
  const DIRS = ['KeyD', 'KeyW', 'KeyA', 'KeyS'];
  for (let round = 0; round < 40 && shot < want; round++) {
    if (round % 3 === 0) {
      await p.keyboard.press('KeyB');
    }
    if (await p.$('.summary')) {
      await p.locator('.summary .btn.primary').click().catch(() => undefined);
      await sleep(2500);
      continue;
    }
    const k = DIRS[round % DIRS.length];
    await p.keyboard.down(k);
    await sleep(900);
    await p.keyboard.up(k);
    await sleep(300);
  }
  await sleep(800); // để lượt chụp cuối kịp ghi xong
  console.log(`${label}: ${shot}/${want} ảnh · ${faces.join(', ') || 'không bắt được emote nào'}${loi.length ? ' · LỖI: ' + loi[0] : ''}`);
  await ctx.close();
}

await shoot('pc', { viewport: { width: 1366, height: 768 } }, 3);
await shoot('iphone', { ...devices['iPhone 13'] }, 3);
await browser.close();
console.log(`\nẢnh trong ${OUT.replace(ROOT, '')}`);
