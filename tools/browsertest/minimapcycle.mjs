import { chromium } from 'playwright';
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const b = await chromium.launch({ headless: true, args: ['--use-gl=angle','--use-angle=gl','--enable-gpu','--ignore-gpu-blocklist'] });

async function run(w, h, mobile) {
  const ctx = await b.newContext({ viewport: { width: w, height: h }, deviceScaleFactor: 2, isMobile: mobile, hasTouch: mobile });
  const p = await ctx.newPage();
  await p.goto('http://127.0.0.1:9998/', { waitUntil: 'domcontentloaded' });
  await p.waitForSelector('input', { timeout: 60000 });
  await p.fill('input', 'MapTest');
  await p.locator('.home .btn.primary').click();
  await p.waitForSelector('.room-board', { timeout: 60000 });
  await sleep(2000);
  const state = () => p.evaluate(() => {
    const e = document.querySelector('.minimap');
    const r = e.getBoundingClientRect();
    return { lop: e.className.replace('minimap', '').trim() || 'goc', w: Math.round(r.width), h: Math.round(r.height) };
  });
  const out = [await state()];
  for (let i = 0; i < 3; i++) {
    await p.click('.minimap', { force: true });
    await sleep(250);
    out.push(await state());
  }
  console.log(`${w}x${h}: ` + out.map((s) => `${s.lop}(${s.w}px)`).join(' → '));
  await ctx.close();
}
await run(360, 800, true);
await run(800, 360, true);
await run(1280, 900, false);
await b.close(); process.exit(0);
