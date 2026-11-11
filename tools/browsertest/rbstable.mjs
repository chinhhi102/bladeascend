// Bảng xếp hạng KHÔNG được co giãn — GAME_SPEC §11.5 (v6.47) + dòng nghiệm thu §17:
//
//   "Ghi lại chiều cao + số dòng bảng xếp hạng mỗi khung trong 2 phút (phòng 20 người, có
//    người vào/ra/chết): chiều cao KHÔNG đổi lần nào, số dòng luôn 10 (+ dòng của mình);
//    số phần tử DOM của bảng không đổi."
//
//   node tools/browsertest/rbstable.mjs http://127.0.0.1:9999
//   SECS=30 node tools/browsertest/rbstable.mjs      # chạy nhanh khi đang sửa
//
// Bài này ĐO MỖI KHUNG, không chụp vài lần: lỗi 07/11 là bảng tụt còn 3 dòng rồi bung lại
// 10 trong khoảng một giây — chụp ảnh thưa thì xác suất trượt rất cao.
//
// Ba con số mỗi khung, vì chúng tố giác ba thủ phạm KHÁC NHAU:
//   · `rows`  — số dòng đang hiện. Tụt = hoặc server gửi thiếu, hoặc client ẩn bớt.
//   · `pkt`   — số dòng trong gói Leaderboard vừa nhận (`window.__rb()`). Tụt = SERVER.
//   · `nodes` — số phần tử DOM trong bảng. Đổi = CLIENT xoá/dựng lại, dù dữ liệu đủ.
import { launch } from './browser.mjs';
import { cpus, loadavg } from 'node:os';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9999';
const SECS = Number(process.env.SECS ?? 120);
/** Cần ≥ ngần này người trong phòng mới kết luận được "số dòng luôn 10". */
const MIN_PLAYERS = Number(process.env.MIN_PLAYERS ?? 11);
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

console.log(`tải máy: ${loadavg()[0].toFixed(1)} trên ${cpus().length} nhân`);

// Cờ GPU như `weakperf.mjs`: thiếu nó thì Chrome headless vẽ bằng SwiftShader và cả trò
// chơi chạy ~1 khung/giây — mẫu đo thưa tới mức có thể bỏ sót một lần bảng co trong 1 giây.
const b = await launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'],
});

async function join(name) {
  const ctx = await b.newContext({ viewport: { width: 1280, height: 720 }, locale: 'vi-VN' });
  const p = await ctx.newPage();
  p.on('pageerror', (e) => console.log(`  [${name}] pageerror`, String(e).slice(0, 160)));
  await p.goto(`${TARGET}/?debug=true`, { waitUntil: 'domcontentloaded' });
  await p.waitForSelector('input', { timeout: 30000 });
  await p.fill('input', name);
  await p.locator('.char-card').first().click();
  await p.locator('.home .btn.primary').click();
  await p.waitForSelector('.room-board', { timeout: 60000 });
  return { page: p, ctx };
}

const me = await join('KiemTraBang');
await sleep(4000);

// Đếm người từ ĐẦU BẢNG, không từ `window.kx`: `kx` bị khoá sau quyền debug của server,
// và lần chạy đầu nó `undefined` → bài báo "phòng 0 người" trong khi phòng đang đông.
const n0 = await me.page.evaluate(() => {
  const txt = document.querySelector('.rb-count')?.textContent ?? '';
  return Number(txt.replace(/\D+/g, '')) || 0;
});
if (n0 < MIN_PLAYERS) {
  console.error(`\nKHÔNG ĐO ĐƯỢC: phòng chỉ có ${n0} người (cần ≥ ${MIN_PLAYERS} để bảng đủ 10 dòng).`);
  console.error('Chạy server có bot (APP_ENV=dev/staging) rồi chạy lại.');
  await b.close();
  // Mã 2 = KHÔNG ĐO ĐƯỢC, khác hẳn mã 1 = ĐO ĐƯỢC VÀ HỎNG. Cổng deploy chỉ chặn ở mã 1;
  // chặn deploy vì phòng vắng là cách chắc chắn để người ta tắt cổng kiểm đi.
  process.exit(2);
}
console.log(`phòng ${n0} người, ghi mỗi khung trong ${SECS}s…`);

// Đồng hồ ghi nằm TRONG trang: ra/vào Node mỗi lần đo thì chính phép đo làm tụt khung.
//
// Lấy mẫu bằng `setInterval` 20 Hz, KHÔNG bằng `requestAnimationFrame`: nhịp rAF chính là
// nhịp khung của trò chơi, và trong máy đo nó có thể tụt xuống 1–2 khung/giây — lúc ấy "ghi
// mỗi khung" chỉ còn vài chục mẫu cho 2 phút, thừa sức bỏ sót một lần bảng co. Số khung
// thật vẫn được đếm riêng để báo cáo.
await me.page.evaluate(() => {
  const rec = { h: [], nodes: [], rows: [], pkt: [], frames: 0, samples: 0 };
  window.__rbRec = rec;
  const frame = () => {
    rec.frames++;
    requestAnimationFrame(frame);
  };
  requestAnimationFrame(frame);
  const tick = () => {
    const el = document.querySelector('.room-board');
    if (el) {
      rec.samples++;
      rec.h.push(Math.round(el.getBoundingClientRect().height));
      rec.nodes.push(el.querySelectorAll('*').length);
      // Dòng ĐANG HIỆN: thẻ top + dòng 4–10 không bị `display:none`.
      const vis = (e) => getComputedStyle(e).display !== 'none';
      const cards = [...el.querySelectorAll('.rb-top1, .rb-pod')].filter(vis).length;
      const list = [...el.querySelectorAll('.rb-list > .rb-row')].filter(vis).length;
      rec.rows.push(cards + list);
      rec.pkt.push(window.__rb ? window.__rb().rows : -1);
    }
  };
  setInterval(tick, 50);
});

// Vào/ra/chết trong lúc đo — đúng cảnh làm lộ lỗi: hai người phụ vào rồi rời, và người đo
// tự chết một lần (bấm X) rồi chơi tiếp.
const extras = [];
let joined = 0;
(async () => {
  for (let i = 0; i < 2; i++) {
    await sleep(8000 + i * 6000);
    try {
      extras.push(await join(`KiemTraVao${i}`));
      joined++;
    } catch {
      /* phòng đầy thì bỏ qua, bot chết/respawn đã đủ churn */
    }
  }
  await sleep(14000);
  for (const e of extras.splice(0)) {
    await e.ctx.close().catch(() => {});
  }
})();

const t0 = Date.now();
let revived = 0;
while (Date.now() - t0 < SECS * 1000) {
  await sleep(2000);
  if (await me.page.$('.summary')) {
    await me.page.locator('.summary .btn.primary').click().catch(() => {});
    await me.page.waitForSelector('.room-board', { timeout: 30000 }).catch(() => {});
    revived++;
    // Vòng ghi chết theo trang? Không — cùng một trang, `requestAnimationFrame` vẫn chạy.
  } else if (Date.now() - t0 > SECS * 400 && revived === 0) {
    await me.page.keyboard.press('KeyX'); // tự chết một lần cho có "người chết"
    revived = -1;
  }
}

const rec = await me.page.evaluate(() => window.__rbRec);
for (const e of extras) {
  await e.ctx.close().catch(() => {});
}
await b.close();

/** runs: các GIÁ TRỊ liên tiếp khác nhau — "10,10,5,5,10" → [10,5,10], 2 lần đổi. */
function runs(a) {
  const out = [];
  for (const v of a) {
    if (out.length === 0 || out[out.length - 1] !== v) {
      out.push(v);
    }
  }
  return out;
}

const hRuns = runs(rec.h);
const nRuns = runs(rec.nodes);
const rRuns = runs(rec.rows);
const pkt = rec.pkt.filter((v) => v >= 0);
const minRows = Math.min(...rec.rows);
const maxRows = Math.max(...rec.rows);

// Có thật là "có người vào/ra/chết" không? Bài nào không gây được biến động thì nó chỉ
// chứng minh bảng đứng yên khi KHÔNG CÓ GÌ XẢY RA — vô nghĩa.
console.log(`\n${rec.samples} mẫu / ${rec.frames} khung vẽ · ${joined} người vào rồi ra · người đo chết ${Math.max(revived, 0)} lần`);
console.log(`chiều cao : ${hRuns.length - 1} lần đổi — ${hRuns.slice(0, 12).join(' → ')}${hRuns.length > 12 ? ' …' : ''}`);
console.log(`phần tử   : ${nRuns.length - 1} lần đổi — ${nRuns.slice(0, 12).join(' → ')}${nRuns.length > 12 ? ' …' : ''}`);
console.log(`số dòng   : ${minRows}–${maxRows}, ${rRuns.length - 1} lần đổi — ${rRuns.slice(0, 16).join(' → ')}${rRuns.length > 16 ? ' …' : ''}`);
if (pkt.length > 0) {
  console.log(`gói server: ${Math.min(...pkt)}–${Math.max(...pkt)} dòng, ${runs(pkt).length - 1} lần đổi`);
} else {
  console.log('gói server: không đo được (`window.__rb` chưa có — bản client cũ)');
}

const fail = [];
if (hRuns.length > 1) {
  fail.push(`chiều cao bảng đổi ${hRuns.length - 1} lần (spec: không đổi lần nào)`);
}
if (nRuns.length > 1) {
  fail.push(`số phần tử DOM đổi ${nRuns.length - 1} lần (spec: dựng 1 lần, chỉ sửa nội dung)`);
}
if (minRows !== maxRows || maxRows !== 10) {
  fail.push(`số dòng ${minRows}–${maxRows} (spec: luôn 10, ô thừa trống giữ chiều cao)`);
}
if (pkt.length > 0 && Math.min(...pkt) < 10) {
  fail.push(`gói server có lúc chỉ ${Math.min(...pkt)} dòng (spec: luôn gửi đủ top 10)`);
}

if (fail.length > 0) {
  console.error(`\nKHÔNG ĐẠT §11.5:\n - ${fail.join('\n - ')}`);
  process.exit(1);
}
console.log('\nĐẠT: bảng đứng yên — chiều cao, số phần tử và số dòng không đổi lần nào.');
