// Kiểm form góp ý có ảnh (ADMIN_SPEC §7.1) trên PC, iPhone (WebKit) và Android (Chromium).
//
// Chạy:  cd client && npm run build   (cần client/dist + assets/dist)
//        cd tools/browsertest && node feedback-images.mjs [thư-mục-ảnh-chụp]
//
// Tự dựng một server RIÊNG (cổng trống, DB tạm), vào trận, mở form góp ý, chọn ảnh qua
// <input type=file> (một JPEG có khối EXIF GPS + một PNG rất lớn), gửi đi, rồi đọc thẳng
// thư mục ảnh của server để chắc: đủ số ảnh, đúng định dạng, ≤ 300 KB, cạnh dài ≤ 1 600 px,
// và KHÔNG còn EXIF/GPS. Trên Chromium còn thử một tệp HEIC giả để thấy câu báo lỗi.
import { chromium, webkit, devices } from 'playwright';
import { spawn, execFileSync } from 'node:child_process';
import { mkdtempSync, mkdirSync, readdirSync, readFileSync, writeFileSync, existsSync } from 'node:fs';
import { createServer } from 'node:net';
import { tmpdir } from 'node:os';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const REPO = join(dirname(fileURLToPath(import.meta.url)), '..', '..');
const SHOTS = process.argv[2] ?? join(REPO, 'docs', 'review', '10-gop-y-anh');
const GPS = 'GPS-21.0285N-105.8542E';
const MAX_BYTES = 300 * 1024;
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

function freePort() {
  return new Promise((resolve) => {
    const s = createServer();
    s.listen(0, '127.0.0.1', () => {
      const { port } = s.address();
      s.close(() => resolve(port));
    });
  });
}

// ── 1. Server riêng ──────────────────────────────────────────────────────────────────────
const dir = mkdtempSync(join(tmpdir(), 'fbimg-'));
const bin = join(dir, 'server');
execFileSync('go', ['build', '-o', bin, './cmd/server'], { cwd: join(REPO, 'server'), stdio: 'inherit' });
const cfg = JSON.parse(readFileSync(join(REPO, 'config', 'game.json'), 'utf8'));
cfg.room.bot_fill_target = 0;
writeFileSync(join(dir, 'game.json'), JSON.stringify(cfg));
const PORT = await freePort();
const srv = spawn(bin, ['-addr', `127.0.0.1:${PORT}`, '-config', join(dir, 'game.json'),
  '-db', join(dir, 'lb.db'), '-telemetry=false'], { stdio: ['ignore', 'pipe', 'pipe'] });
srv.stdout.on('data', () => {});
srv.stderr.on('data', () => {});
process.on('exit', () => srv.kill());
const FB_DIR = join(dir, 'feedback');
for (let i = 0; i < 60; i++) {
  try {
    if ((await fetch(`http://127.0.0.1:${PORT}/api/config`)).ok) {
      break;
    }
  } catch (_) { /* chưa lên */ }
  await sleep(250);
}

// ── 2. Ảnh mẫu ───────────────────────────────────────────────────────────────────────────
function exifGPS() {
  // TIFF big-endian: IFD0 → GPS IFD có chuỗi GPS để dò.
  const t = Buffer.alloc(56 + GPS.length);
  t.write('MM\x00\x2A', 0, 'latin1');
  t.writeUInt32BE(8, 4);
  t.writeUInt16BE(1, 8);
  t.writeUInt16BE(0x8825, 10); t.writeUInt16BE(4, 12); t.writeUInt32BE(1, 14); t.writeUInt32BE(26, 18);
  t.writeUInt32BE(0, 22);
  t.writeUInt16BE(1, 26);
  t.writeUInt16BE(0x001B, 28); t.writeUInt16BE(7, 30); t.writeUInt32BE(GPS.length, 32); t.writeUInt32BE(56, 36);
  t.writeUInt32BE(0, 40);
  t.write(GPS, 56, 'latin1');
  const payload = Buffer.concat([Buffer.from('Exif\x00\x00', 'latin1'), t]);
  const seg = Buffer.alloc(4);
  seg.writeUInt16BE(0xFFE1, 0);
  seg.writeUInt16BE(payload.length + 2, 2);
  return Buffer.concat([seg, payload]);
}

async function makeFixtures() {
  const b = await chromium.launch();
  const p = await b.newPage();
  const urls = await p.evaluate(() => {
    const draw = (w, h, noise) => {
      const c = document.createElement('canvas');
      c.width = w; c.height = h;
      const g = c.getContext('2d');
      const gr = g.createLinearGradient(0, 0, w, h);
      gr.addColorStop(0, '#3a7bd5'); gr.addColorStop(1, '#f5a623');
      g.fillStyle = gr; g.fillRect(0, 0, w, h);
      if (noise) {
        const d = g.getImageData(0, 0, w, h);
        for (let i = 0; i < d.data.length; i += 4) {
          d.data[i] ^= (Math.random() * 90) | 0;
          d.data[i + 1] ^= (Math.random() * 90) | 0;
        }
        g.putImageData(d, 0, 0);
      }
      g.fillStyle = '#fff'; g.font = `bold ${h / 6}px sans-serif`;
      g.fillText(noise ? 'PNG lớn' : 'Ảnh GPS', w / 10, h / 2);
      return c;
    };
    return {
      jpg: draw(2400, 1800, false).toDataURL('image/jpeg', 0.92),
      png: draw(3000, 2000, true).toDataURL('image/png'),
    };
  });
  await b.close();
  const jpg = Buffer.from(urls.jpg.split(',')[1], 'base64');
  const withGPS = Buffer.concat([jpg.subarray(0, 2), exifGPS(), jpg.subarray(2)]);
  const f = {
    gps: join(dir, 'dien-thoai-gps.jpg'),
    big: join(dir, 'anh-lon.png'),
    heic: join(dir, 'iphone.heic'),
  };
  writeFileSync(f.gps, withGPS);
  writeFileSync(f.big, Buffer.from(urls.png.split(',')[1], 'base64'));
  // HEIC giả: hộp `ftyp heic` rồi rác — Chromium không giải mã được, phải ra câu báo lỗi.
  writeFileSync(f.heic, Buffer.concat([Buffer.from('\x00\x00\x00\x18ftypheic\x00\x00\x00\x00mif1heic', 'latin1'), Buffer.alloc(512, 7)]));
  return f;
}

// ── 3. Đọc ảnh server đã lưu ─────────────────────────────────────────────────────────────
function dims(b) {
  if (b[0] === 0xFF && b[1] === 0xD8) {
    for (let i = 2; i < b.length - 9;) {
      if (b[i] !== 0xFF) { return null; }
      const m = b[i + 1];
      if (m >= 0xC0 && m <= 0xC2) {
        return { type: 'jpeg', h: b.readUInt16BE(i + 5), w: b.readUInt16BE(i + 7) };
      }
      i += 2 + b.readUInt16BE(i + 2);
    }
    return null;
  }
  if (b.toString('latin1', 0, 4) === 'RIFF') {
    const fourcc = b.toString('latin1', 12, 16);
    if (fourcc === 'VP8X') {
      return { type: 'webp', w: 1 + b.readUIntLE(24, 3), h: 1 + b.readUIntLE(27, 3) };
    }
    if (fourcc === 'VP8 ') {
      return { type: 'webp', w: b.readUInt16LE(26) & 0x3fff, h: b.readUInt16LE(28) & 0x3fff };
    }
    if (fourcc === 'VP8L') {
      const v = b.readUInt32LE(21);
      return { type: 'webp', w: 1 + (v & 0x3fff), h: 1 + ((v >> 14) & 0x3fff) };
    }
  }
  if (b[0] === 0x89 && b.toString('latin1', 1, 4) === 'PNG') {
    return { type: 'png', w: b.readUInt32BE(16), h: b.readUInt32BE(20) };
  }
  return null;
}

function storedSince(before) {
  if (!existsSync(FB_DIR)) { return []; }
  return readdirSync(FB_DIR).filter((n) => !before.has(n)).sort();
}

// ── 4. Một lượt trên một thiết bị ────────────────────────────────────────────────────────
const failures = [];
function check(ok, what) {
  console.log(`  ${ok ? 'OK ' : 'LỖI'} ${what}`);
  if (!ok) { failures.push(what); }
}

// SAFARI_NO_WEBP: Safari thật KHÔNG mã hoá được WebP — `toBlob/toDataURL('image/webp')`
// lặng lẽ trả PNG. WebKit của Playwright (bản GTK) lại mã hoá được, nên giả lập đúng hành vi
// Safari bằng cách đổi kiểu yêu cầu, để chắc đường lùi sang JPEG chạy.
const SAFARI_NO_WEBP = () => {
  const swap = (t) => (t === 'image/webp' ? 'image/png' : t);
  const blob = HTMLCanvasElement.prototype.toBlob;
  HTMLCanvasElement.prototype.toBlob = function (cb, t, q) { return blob.call(this, cb, swap(t), q); };
  const url = HTMLCanvasElement.prototype.toDataURL;
  HTMLCanvasElement.prototype.toDataURL = function (t, q) { return url.call(this, swap(t), q); };
};

async function run({ label, engine, ctxOpts, shotName, tryHeic = false, noWebp = false }) {
  console.log(`\n== ${label}`);
  const browser = await engine.launch();
  const ctx = await browser.newContext(ctxOpts);
  if (noWebp) {
    await ctx.addInitScript(SAFARI_NO_WEBP);
  }
  const page = await ctx.newPage();
  page.on('pageerror', (e) => console.log('  [pageerror]', e.message));
  await page.goto(`http://127.0.0.1:${PORT}/`, { waitUntil: 'domcontentloaded' });
  await page.waitForSelector('input', { timeout: 30000 });
  await page.fill('input', label.slice(0, 10).replace(/\W/g, '') || 'Test');
  await page.locator('.home .btn.primary').click();
  await page.waitForFunction(() => window.__playing?.() === true, null, { timeout: 30000 });
  await sleep(1500);
  await page.locator('.fb-btn.fb-match').click();
  await page.waitForSelector('.fb-box');
  // Ảnh chụp màn hình tự kèm khi mở (loại mặc định = Báo lỗi).
  await page.waitForFunction(() => !document.querySelector('.fb-send')?.disabled, null, { timeout: 15000 });
  const hasShot = (await page.locator('.fb-thumb-tag').count()) === 1;
  if (process.env.FBDEBUG) {
    console.log('  [debug]', await page.evaluate(() => ({ msg: document.querySelector('.fb-msg')?.textContent,
      thumbs: document.querySelectorAll('.fb-thumb').length, disabled: document.querySelector('.fb-send')?.disabled,
      shot: document.querySelector('.fb-shot')?.checked })));
    page.on('console', (m) => console.log('  [console]', m.text()));
  }
  check(hasShot, 'ảnh chụp màn hình tự kèm thành ô xem trước');

  if (tryHeic) {
    await page.setInputFiles('.fb-file', FILES.heic);
    await page.waitForFunction(() => document.querySelector('.fb-msg')?.textContent?.includes('Chưa hỗ trợ'), null, { timeout: 15000 })
      .then(() => check(true, 'HEIC không đọc được → "Chưa hỗ trợ định dạng này"'))
      .catch(() => check(false, 'HEIC không đọc được → "Chưa hỗ trợ định dạng này"'));
  }

  await page.setInputFiles('.fb-file', [FILES.gps, FILES.big]);
  await page.waitForFunction((n) => document.querySelectorAll('.fb-thumb').length === n
    && !document.querySelector('.fb-send')?.disabled, hasShot ? 3 : 2, { timeout: 30000 });
  const n = await page.locator('.fb-thumb').count();
  check(n === (hasShot ? 3 : 2), `có ${n} ô xem trước`);

  // Thêm ảnh thứ 4 phải bị chặn.
  await page.setInputFiles('.fb-file', FILES.gps);
  await sleep(500);
  check((await page.locator('.fb-thumb').count()) === n, 'ảnh thứ 4 bị chặn (tối đa 3)');

  await page.fill('.fb-body', 'Kiểm thử tải ảnh góp ý từ ' + label);
  if (shotName) {
    await page.locator('.fb-box').screenshot({ path: join(SHOTS, shotName) });
    console.log(`  ảnh chụp: ${join(SHOTS, shotName)}`);
  }

  const before = new Set(existsSync(FB_DIR) ? readdirSync(FB_DIR) : []);
  await page.locator('.fb-send').click();
  await page.waitForFunction(() => document.querySelector('.fb-msg')?.textContent?.includes('Cảm ơn'), null, { timeout: 20000 })
    .then(() => check(true, 'server nhận: "Cảm ơn! Đã nhận góp ý."'))
    .catch(async () => check(false, 'gửi thất bại: ' + (await page.locator('.fb-msg').textContent())));

  const stored = storedSince(before);
  check(stored.length === n, `server lưu ${stored.length} ảnh`);
  for (const name of stored) {
    const b = readFileSync(join(FB_DIR, name));
    const d = dims(b);
    const clean = !b.includes(Buffer.from(GPS)) && !b.includes(Buffer.from('Exif'));
    const wantType = noWebp ? 'jpeg' : 'webp';
    check(d !== null && d.type === wantType && b.length <= MAX_BYTES && Math.max(d.w, d.h) <= 1600 && clean,
      `${name}: ${d?.type} ${d?.w}×${d?.h}, ${(b.length / 1024).toFixed(0)} KB, ${clean ? 'không EXIF/GPS' : 'CÒN EXIF/GPS'}`);
  }
  await browser.close();
}

mkdirSync(SHOTS, { recursive: true });
const FILES = await makeFixtures();
console.log(`server tạm: 127.0.0.1:${PORT}, DB ${join(dir, 'lb.db')}`);
await run({ label: 'PC Chromium 1366x768', engine: chromium, ctxOpts: { viewport: { width: 1366, height: 768 } },
  shotName: 'pc-1366x768-3-anh.png', tryHeic: true });
await run({ label: 'iPhone WebKit', engine: webkit, ctxOpts: { ...devices['iPhone 13'] },
  shotName: 'iphone-webkit-3-anh.png' });
await run({ label: 'iPhone WebKit, Safari không mã hoá WebP', engine: webkit, ctxOpts: { ...devices['iPhone 13'] },
  noWebp: true });
await run({ label: 'Android Chromium', engine: chromium, ctxOpts: { ...devices['Pixel 7'] },
  shotName: 'android-chromium-3-anh.png', tryHeic: true });

console.log(failures.length ? `\nTHẤT BẠI ${failures.length} mục` : '\nTẤT CẢ ĐẠT');
srv.kill();
process.exit(failures.length ? 1 : 0);
