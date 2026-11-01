// Mức Cao tốn ở ĐÂU? — tắt từng thứ mà đo (chủ dự án chốt 1/11, PERF_LOG §6).
//
//   node tools/browsertest/qualitycost.mjs [giây-mỗi-lượt] [--zone=2] [--cpu=4] [--runs=2]
//
// Telemetry 31/10: Cao vẽ p95 64,9 ms, Vừa 18,2 ms. Con số gộp đó không nói phần nào nằm ở
// đâu, và nó lại lẫn người chơi ở các khu, các máy khác nhau. Bài này giữ mọi thứ CỐ ĐỊNH:
//   - Server RIÊNG cổng 9995, DB tạm, bản client vừa build (`client/dist`), không đụng :9999.
//   - Vào đúng MỘT khu (mặc định khu 2 — khu tệ nhất trong telemetry 31/10). Server chọn khu
//     ít người nhất, hoà thì ngẫu nhiên, nên bài này vào lại tới khi trúng khu cần đo.
//   - Một tab, CPU bóp 4× qua CDP (Emulation.setCPUThrottlingRate).
//   - Mức chất lượng ép bằng khoá lựa chọn tay (`kx.quality.manual`) để `autoQuality` không
//     tự đổi giữa lượt đo; từng thứ hạ riêng bằng `?qf=tên:medium` (render/quality.ts).
//
// Mỗi cấu hình đo `--runs` lượt, mỗi lượt vào trận MỚI (chỗ hồi sinh ngẫu nhiên), in trung
// vị. Số đo là khoảng cách giữa hai lần requestAnimationFrame — tức là khung hình người
// chơi thật thấy, gồm cả phần GPU chặn luồng chính.
import { chromium } from 'playwright';
import { spawn } from 'node:child_process';
import { mkdtempSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const REPO = join(dirname(fileURLToPath(import.meta.url)), '..', '..');
const PORT = 9995;
const arg = (name, def) => {
  const a = process.argv.find((x) => x.startsWith(`--${name}=`));
  return a ? Number(a.split('=')[1]) : def;
};
const SECS = Number(process.argv[2] && !process.argv[2].startsWith('--') ? process.argv[2] : 20);
const ZONE = arg('zone', 2);
const CPU = arg('cpu', 4);
const RUNS = arg('runs', 2);
const ONLY = (process.argv.find((x) => x.startsWith('--only=')) ?? '').slice(7).split(',').filter(Boolean);
const WARMUP_MS = 5000;
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

const PC = { viewport: { width: 1366, height: 768 } };
const PHONE = {
  viewport: { width: 390, height: 844 }, deviceScaleFactor: 3, isMobile: true, hasTouch: true,
  userAgent: 'Mozilla/5.0 (Linux; Android 14; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Mobile Safari/537.36',
};

/** CONFIGS: tên · thiết bị · mức · ghi đè từng thứ. "Cao −X" = Cao nhưng X hạ xuống Vừa. */
const CONFIGS = [
  { name: 'Vừa', dev: PC, q: 'medium', qf: '' },
  { name: 'Cao', dev: PC, q: 'high', qf: '' },
  { name: 'Cao −bloom', dev: PC, q: 'high', qf: 'bloom:medium' },
  { name: 'Cao −ambient', dev: PC, q: 'high', qf: 'ambient:medium' },
  { name: 'Cao −decals', dev: PC, q: 'high', qf: 'decals:medium' },
  { name: 'Cao −beast', dev: PC, q: 'high', qf: 'beast:medium' },
  { name: 'Cao −motes', dev: PC, q: 'high', qf: 'motes:medium' },
  { name: 'Thấp', dev: PC, q: 'low', qf: '' },
  { name: 'ĐT Thấp', dev: PHONE, q: 'low', qf: '' },
  { name: 'ĐT Vừa', dev: PHONE, q: 'medium', qf: '' },
].filter((c) => ONLY.length === 0 || ONLY.includes(c.name));

// ── Server riêng ─────────────────────────────────────────────────────────────────────────
const dir = mkdtempSync(join(tmpdir(), 'kx-qcost-'));
const srv = spawn(join(REPO, 'server', 'bladeascend'), [
  '-addr', `127.0.0.1:${PORT}`,
  '-config', join(REPO, 'config', 'game.json'),
  '-db', join(dir, 'lb.db'),
  '-telemetry=false',
  '-client', join(REPO, 'client', 'dist'),
  '-assets', join(REPO, 'assets', 'dist'),
], { stdio: ['ignore', 'pipe', 'pipe'], env: { ...process.env, APP_ENV: 'dev' } });
srv.stdout.on('data', () => {});
srv.stderr.on('data', () => {});
process.on('exit', () => srv.kill());
await sleep(3000);

const browser = await chromium.launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'],
});

/** joinZone vào trận, lặp lại tới khi trúng khu `ZONE`. Trả về trang đã vào, hoặc null. */
async function joinZone(ctx, cfg) {
  const url = `http://127.0.0.1:${PORT}/?debug=true${cfg.qf ? `&qf=${cfg.qf}` : ''}`;
  for (let tries = 0; tries < 15; tries++) {
    const page = await ctx.newPage();
    await page.goto(url, { waitUntil: 'domcontentloaded' });
    await page.waitForSelector('input', { timeout: 60000 });
    await page.fill('input', 'DoChatLuong');
    await page.getByText('Vào chơi', { exact: true }).click();
    await page.waitForFunction(() => window.__playing?.() === true, null, { timeout: 60000 });
    const zone = await page.evaluate(() => window.__stats().zone);
    if (zone === ZONE) {
      return page;
    }
    await page.close();
  }
  return null;
}

async function measure(cfg) {
  const ctx = await browser.newContext(cfg.dev);
  await ctx.addInitScript((q) => localStorage.setItem('kx.quality.manual', q), cfg.q);
  const page = await joinZone(ctx, cfg);
  if (!page) {
    await ctx.close();
    return null;
  }
  const cdp = await ctx.newCDPSession(page);
  await cdp.send('Emulation.setCPUThrottlingRate', { rate: CPU });
  await sleep(WARMUP_MS);
  // Đi vòng chậm để cảnh thay đổi như lúc chơi thật (decal, mây, nhân vật vào/ra khung).
  const walk = (async () => {
    const keys = ['KeyD', 'KeyS', 'KeyA', 'KeyW'];
    for (let i = 0; i < SECS / 2; i++) {
      const k = keys[i % 4];
      await page.keyboard.down(k).catch(() => {});
      await sleep(2000);
      await page.keyboard.up(k).catch(() => {});
    }
  })();
  const r = await page.evaluate(async (ms) => {
    const dts = [];
    let last = performance.now();
    const t0 = last;
    const draws = [], sprites = [], drawMs = [];
    let nextStat = t0;
    await new Promise((done) => {
      const tick = (now) => {
        dts.push(now - last);
        last = now;
        if (now >= nextStat) {
          const s = window.__stats();
          draws.push(s.drawCalls); sprites.push(s.sprites); drawMs.push(s.drawMs);
          nextStat = now + 500;
        }
        if (now - t0 < ms) { requestAnimationFrame(tick); } else { done(); }
      };
      requestAnimationFrame(tick);
    });
    const s = window.__stats();
    return { dts, draws, sprites, drawMs, zone: s.zone, players: s.players, quality: s.quality };
  }, SECS * 1000);
  await walk;
  const gpu = await page.evaluate(() => {
    const gl = document.createElement('canvas').getContext('webgl2');
    const e = gl?.getExtension('WEBGL_debug_renderer_info');
    return e ? String(gl.getParameter(e.UNMASKED_RENDERER_WEBGL)) : '?';
  });
  await ctx.close();
  return { ...r, gpu };
}

const pct = (v, p) => {
  const s = [...v].sort((a, b) => a - b);
  return s.length ? s[Math.min(s.length - 1, Math.floor(p * s.length))] : 0;
};
const med = (v) => pct(v, 0.5);

const rows = [];
let gpuSeen = '';
for (const cfg of CONFIGS) {
  const runs = [];
  for (let i = 0; i < RUNS; i++) {
    const r = await measure(cfg);
    if (!r) {
      console.error(`${cfg.name}: 15 lần vào không trúng khu ${ZONE} — bỏ lượt`);
      continue;
    }
    gpuSeen = r.gpu;
    // FPS p50 = 1000 / thời gian khung trung vị; FPS p1 = 1000 / thời gian khung ở phân vị 99.
    runs.push({
      p50: 1000 / med(r.dts), p1: 1000 / pct(r.dts, 0.99),
      drawMs: med(r.drawMs), draws: med(r.draws), sprites: med(r.sprites),
    });
  }
  if (runs.length === 0) {
    continue;
  }
  const row = {
    name: cfg.name,
    p50: med(runs.map((x) => x.p50)), p1: med(runs.map((x) => x.p1)),
    drawMs: med(runs.map((x) => x.drawMs)), draws: med(runs.map((x) => x.draws)),
    sprites: med(runs.map((x) => x.sprites)), n: runs.length,
  };
  rows.push(row);
  console.error(`… ${row.name}: p50 ${row.p50.toFixed(1)} · p1 ${row.p1.toFixed(1)}`);
}
await browser.close();

console.log(`\nKhu ${ZONE} · CPU ${CPU}× · ${SECS}s/lượt · ${RUNS} lượt/cấu hình · GPU: ${gpuSeen}\n`);
console.log('| Cấu hình | FPS p50 | FPS p1 | Cập nhật cảnh (ms) | Draw call | Sprite | Lượt |');
console.log('|---|---|---|---|---|---|---|');
for (const r of rows) {
  console.log(`| ${r.name} | ${r.p50.toFixed(1)} | ${r.p1.toFixed(1)} | ${r.drawMs.toFixed(2)} | ${r.draws.toFixed(0)} | ${r.sprites.toFixed(0)} | ${r.n} |`);
}
process.exit(0);
