// Đi tìm lỗi ĐỨNG HÌNH của người chơi ngày 12/11 — xem `docs/LOI_DUNG_HINH.md`.
//
//   node tools/browsertest/filtercrash.mjs [số khung]
//
// Bài này dựng lại CÂY CẢNH của game bằng PixiJS trần (không cần server, không cần client):
// camera chứa ground/decor **dùng CHUNG một** `ColorMatrixFilter` như `mapview.ts:336-337`,
// lớp vfx có `AdvancedBloomFilter`, thỉnh thoảng gắn thêm `ShockwaveFilter` lên camera như
// lúc đột phá (`breakthrough.ts:62`) — tức bộ lọc LỒNG NHAU — rồi vừa chạy vừa đổi cỡ cửa
// sổ, đổi resolution, quét `filterArea` ra cả ngoài màn hình.
//
// TRẠNG THÁI 12/11: chạy 2 000 khung **KHÔNG ra lỗi**. Giữ lại làm điểm xuất phát, việc kế
// tiếp là nới cho giống thật hơn (client thật + ép đột phá liên tục qua lệnh `dbg realm`,
// xem `dbggate.mjs`). Đừng xoá vì nó "không tìm thấy gì": nó đã loại được một giả thuyết.
import http from 'node:http';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { launch } from './browser.mjs';

const HERE = path.dirname(fileURLToPath(import.meta.url));
const NM = path.join(HERE, '../../client/node_modules');
const FRAMES = Number(process.argv[2] ?? 3000);

const PAGE = `<!doctype html><meta charset=utf-8>
<script type="importmap">{"imports":{"pixi.js":"/pixi.mjs"}}</script>
<body><script type=module>
import * as PIXI from 'pixi.js';
import { AdvancedBloomFilter, ShockwaveFilter } from '/pixi-filters.mjs';
const log = [];
window.__log = log;
window.addEventListener('error', (e) => log.push('ERROR: ' + (e.error?.stack ?? e.message)));
window.__run = async (frames) => {
  log.push('pixi ' + PIXI.VERSION);
  const app = new PIXI.Application();
  await app.init({ width: 1717, height: 820, preference: 'webgl', autoStart: false, resolution: 1 });
  document.body.appendChild(app.canvas);

  const camera = new PIXI.Container();
  const ground = new PIXI.Container();
  const decor = new PIXI.Container();
  const entities = new PIXI.Container();
  const vfx = new PIXI.Container();
  camera.addChild(ground, decor, entities, vfx);
  app.stage.addChild(camera);

  for (let i = 0; i < 40; i++) {
    ground.addChild(new PIXI.Graphics().rect(i * 300, 0, 280, 280).fill(0x224422));
    decor.addChild(new PIXI.Graphics().circle(i * 300, 400, 60).fill(0x886644));
    entities.addChild(new PIXI.Graphics().circle(i * 120, 300, 20).fill(0xff8844));
    vfx.addChild(new PIXI.Graphics().circle(i * 120, 300, 28).fill(0xffee99));
  }
  // Giống mapview.ts: MỘT instance filter dùng cho cả hai container.
  const cm = new PIXI.ColorMatrixFilter();
  cm.saturate(0.2, true);
  ground.filters = [cm];
  decor.filters = [cm];
  ground.filterArea = new PIXI.Rectangle(0, 0, 1, 1);
  decor.filterArea = new PIXI.Rectangle(0, 0, 1, 1);
  vfx.filters = [new AdvancedBloomFilter({ threshold: 0.45, bloomScale: 1.1, blur: 6, quality: 4 })];

  let shock = null;
  const sizes = [[1717, 820], [1717, 820], [1280, 700], [1920, 993], [900, 500]];
  for (let f = 0; f < frames; f++) {
    // máy quay chạy → filterArea quét theo (có lúc ra NGOÀI cảnh → Pixi "skip" bộ lọc)
    const x = Math.sin(f / 37) * 6000;
    const y = Math.cos(f / 53) * 3000;
    camera.position.set(-x + 860, -y + 410);
    for (const c of [ground, decor]) {
      c.filterArea.x = x - 900;
      c.filterArea.y = y - 500;
      c.filterArea.width = 1800;
      c.filterArea.height = 1000;
    }
    // đột phá: gắn Shockwave lên camera 0,7 s rồi gỡ → filter LỒNG filter
    if (f % 97 === 0) {
      shock = new ShockwaveFilter({ center: { x: 400, y: 300 }, amplitude: 28, wavelength: 180, speed: 900, radius: -1 });
      camera.filters = [...(camera.filters ?? []), shock];
    }
    if (shock && f % 97 === 42) {
      camera.filters = (camera.filters ?? []).filter((v) => v !== shock);
      shock = null;
    }
    if (shock) shock.time = (f % 97) / 60;
    // đổi cỡ cửa sổ / độ phân giải giữa chừng
    if (f % 211 === 0) {
      const [w, h] = sizes[(f / 211 | 0) % sizes.length];
      app.renderer.resize(w, h);
    }
    if (f % 509 === 0) {
      app.renderer.resize(app.screen.width, app.screen.height, f % 1018 === 0 ? 1 : 0.75);
    }
    // vfx đôi khi rỗng hoàn toàn (không có gì phát sáng) — bounds rỗng
    vfx.visible = (f % 31) > 3;
    try {
      app.renderer.render(app.stage);
    } catch (e) {
      log.push('CRASH khung ' + f + ': ' + (e?.stack ?? e));
      return log;
    }
    if (f % 500 === 0) await new Promise((r) => requestAnimationFrame(r));
  }
  log.push('chạy hết ' + frames + ' khung, không lỗi');
  return log;
};
</script>`;

const srv = http.createServer((req, res) => {
  const files = { '/pixi.mjs': 'pixi.js/dist/pixi.mjs', '/pixi-filters.mjs': 'pixi-filters/dist/pixi-filters.mjs' };
  if (files[req.url]) {
    res.writeHead(200, { 'content-type': 'text/javascript' });
    fs.createReadStream(path.join(NM, files[req.url])).pipe(res);
    return;
  }
  res.writeHead(200, { 'content-type': 'text/html' });
  res.end(PAGE);
});
await new Promise((r) => srv.listen(0, r));
const port = srv.address().port;

const browser = await launch({ args: ['--use-gl=angle', '--use-angle=swiftshader', '--enable-unsafe-swiftshader'] });
const page = await browser.newPage();
page.on('pageerror', (e) => console.log('  [pageerror]', String(e).slice(0, 500)));
await page.goto(`http://127.0.0.1:${port}/`);
const out = await page.evaluate((n) => window.__run(n), FRAMES);
for (const l of out) console.log(l.slice(0, 900));
await browser.close();
srv.close();
