// Tái hiện "entity respawn trong 2 giây" như người chơi THẬT gặp — bổ sung cho netcode-bench.
//
//   node tools/browsertest/respawn-wander.mjs [client-dir] [giây] [số-tab]
//
// Vì sao có bài này: telemetry người thật 1/11 (build 20261101_1149) ghi 100 lần respawn/17
// phút, trong khi `netcode-bench` cùng ngày chạy 3 lượt ra `rời=0 · quay lại=0` — bài đó
// không ép được entity ra/vào khung nhìn (tab đi không lách vật cản tới được tab kia, và với
// `bot_fill_target = 0` thì `trimOverCap` xoá ngay bot thả bằng phím B).
//
// Bài này làm như người chơi: server RIÊNG (cổng 9993, DB tạm, telemetry tắt) với đủ 60 bot,
// N tab đi lang thang theo nhịp cố định, mỗi tab đọc thẳng bộ đếm của client
// (`world.respawnBy`: do `Despawn` của server · do luật hết hạn 2 giây ở client).
// `client-dir` cho phép chạy CÙNG bài trên một bản client cũ trong `server/data/deploy-archive/`
// để so trước/sau theo build.
import { chromium } from 'playwright';
import { spawn } from 'node:child_process';
import { mkdtempSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const REPO = join(dirname(fileURLToPath(import.meta.url)), '..', '..');
const CLIENT = process.argv[2] ?? join(REPO, 'server', 'data', 'deploy', 'client');
const SECS = Number(process.argv[3] ?? 120);
const TABS = Number(process.argv[4] ?? 4);
const PORT = 9993;
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

const dir = mkdtempSync(join(tmpdir(), 'kx-wander-'));
const srv = spawn(join(REPO, 'server', 'bladeascend'), [
  '-addr', `127.0.0.1:${PORT}`, '-db', join(dir, 'lb.db'), '-telemetry=false',
  '-max-conns-per-ip', '40', '-client', CLIENT, '-assets', join(CLIENT, '..', 'assets'),
], { stdio: 'ignore', env: { ...process.env, APP_ENV: 'dev' } });
process.on('exit', () => srv.kill());
await sleep(3000);

const browser = await chromium.launch({ headless: true, args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'] });
async function tab(i) {
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 720 } });
  const page = await ctx.newPage();
  const errs = [];
  page.on('pageerror', (e) => errs.push(String(e.message)));
  await page.goto(`http://127.0.0.1:${PORT}/?debug=true`);
  await page.waitForSelector('input', { timeout: 60000 });
  await page.fill('input', `Lang${i}`);
  await page.locator('.home .btn.primary').click();
  await page.waitForFunction(() => window.__playing?.() === true, null, { timeout: 60000 });
  // Bất tử + vài cảnh giới để không chết giữa bài (chết = hồi sinh id mới, làm nhiễu số).
  for (let k = 0; k < 3; k++) { await page.keyboard.press('r'); await sleep(200); }
  const base = await page.evaluate(() => ({ ds: window.kx.world.respawnBy.ds, het: window.kx.world.respawnBy.hethan, quai: window.kx.world.respawnBy.quai ?? 0 }));
  const keys = [['KeyD'], ['KeyD', 'KeyS'], ['KeyS'], ['KeyA', 'KeyS'], ['KeyA'], ['KeyA', 'KeyW'], ['KeyW'], ['KeyD', 'KeyW']];
  const t0 = Date.now();
  let step = i * 3;
  while (Date.now() - t0 < SECS * 1000) {
    const k = keys[step++ % keys.length];
    for (const x of k) { await page.keyboard.down(x); }
    await sleep(2500);
    for (const x of k) { await page.keyboard.up(x); }
    if (!(await page.evaluate(() => window.__playing?.()))) {
      await page.locator('.summary .btn.primary').click().catch(() => {});
      await sleep(2000);
    }
  }
  const end = await page.evaluate(() => ({ ds: window.kx.world.respawnBy.ds, het: window.kx.world.respawnBy.hethan, quai: window.kx.world.respawnBy.quai ?? 0 }));
  await ctx.close();
  return { ds: end.ds - base.ds, het: end.het - base.het, quai: end.quai - base.quai, errs };
}
const rows = await Promise.all(Array.from({ length: TABS }, (_, i) => tab(i)));
await browser.close();
const ds = rows.reduce((t, r) => t + r.ds, 0);
const het = rows.reduce((t, r) => t + r.het, 0);
const quai = rows.reduce((t, r) => t + r.quai, 0);
const errs = rows.flatMap((r) => r.errs);
const mins = (SECS * TABS) / 60;
console.log(`client: ${CLIENT}`);
console.log(`${TABS} tab × ${SECS}s · respawn ${ds + het} (do Despawn ${ds} · do hết hạn 2s ${het}, trong đó quái ${quai}) · ${((ds + het) / mins).toFixed(2)}/phút-tab · lỗi JS ${errs.length}`);
for (const e of [...new Set(errs)].slice(0, 3)) { console.log(`  lỗi: ${e.slice(0, 120)}`); }
process.exit(0);
