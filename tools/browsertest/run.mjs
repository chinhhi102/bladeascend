// 10 tab Chrome headless tự chơi, đo hiệu năng client (GAME_SPEC §13.6).
//
//   cd tools/browsertest && npm install && npm run install-browser
//   node run.mjs
//   node run.mjs --tabs 10 --warmup 30 --fight 60
//
// Bài test **tự dựng một server riêng** ở cổng 8099 với database tạm, rồi tắt khi xong —
// không đụng tới server 8080 đang có người chơi thật. 10 bot giả sẽ không chen vào phòng
// của ai, và bảng xếp hạng thật cũng không bị bẩn.
// Muốn dùng server có sẵn thì: `node run.mjs --port 8080 --no-spawn` (nhớ server đó phải
// chạy với `-max-conns-per-ip 40`, vì 10 tab cùng đến từ 127.0.0.1).
//
// Mỗi tab là một **browser context riêng** → localStorage riêng, coi như một người chơi
// khác máy. Tab tự đi lại ngẫu nhiên, sau `warmup` giây thì 5 tab lên Đạo Tổ đủ kiếm rồi
// cả 10 tab kéo về chụm một chỗ để ép va chạm kiếm và dồn vật phẩm vào tầm nhìn.
//
// ĐỌC KẾT QUẢ: headless Chrome vẽ WebGL bằng CPU (SwiftShader) và 10 tab tranh nhau CPU
// trên cùng một máy, nên **FPS ở đây thấp hơn thực tế rất nhiều — đừng dùng để kết luận về
// FPS**. Số đáng tin là: ping, KB/s, kích thước snapshot, tick server, số sprite, draw call.

import { chromium } from 'playwright';
import { spawn } from 'node:child_process';
import { existsSync } from 'node:fs';
import { mkdir, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const HERE = dirname(fileURLToPath(import.meta.url));
const REPO = resolve(HERE, '../..');

const args = Object.fromEntries(
  process.argv.slice(2).flatMap((a, i, all) => (a.startsWith('--') ? [[a.slice(2), all[i + 1] ?? '1']] : [])),
);
const PORT = Number(args.port ?? 8099);
const SPAWN = args['no-spawn'] === undefined;
const TABS = Number(args.tabs ?? 10);
const TIEN_TABS = Number(args.tien ?? 5);
const WARMUP_S = Number(args.warmup ?? 30);
const FIGHT_S = Number(args.fight ?? 60);
const BASE = `http://127.0.0.1:${PORT}/?debug=true`;

// 4G yếu (spec yêu cầu): trễ 100 ms, 1.6 Mbps xuống, 750 Kbps lên.
const SLOW_4G = { offline: false, latency: 100, downloadThroughput: (1.6 * 1024 * 1024) / 8, uploadThroughput: (750 * 1024) / 8 };

const DIRS = { up: 'KeyW', down: 'KeyS', left: 'KeyA', right: 'KeyD' };
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

/** goBin tìm lệnh `go`: trong PATH, hoặc bản nằm trong cache toolchain của máy này. */
function goBin() {
  const cached = join(
    process.env.HOME ?? '',
    'go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.11.linux-amd64/bin/go',
  );
  return existsSync(cached) ? cached : 'go';
}

/**
 * startServer dựng một server RIÊNG cho bài test: cổng khác, database tạm, nới trần kết nối
 * mỗi IP. Người đang chơi ở server thật không bị 10 bot giả chen vào phòng.
 */
async function startServer() {
  const db = join(tmpdir(), `kx-browsertest-${process.pid}.db`);
  const proc = spawn(
    goBin(),
    ['run', './cmd/server', '-addr', `:${PORT}`, '-db', db, '-max-conns-per-ip', '40'],
    {
      cwd: resolve(REPO, 'server'),
      env: { ...process.env, GOMODCACHE: process.env.GOMODCACHE ?? join(process.env.HOME ?? '', '.cache/kx-gomod') },
      stdio: ['ignore', 'pipe', 'pipe'],
    },
  );
  let log = '';
  proc.stdout.on('data', (b) => (log += b));
  proc.stderr.on('data', (b) => (log += b));
  const deadline = Date.now() + 180000;
  while (Date.now() < deadline) {
    try {
      const r = await fetch(`http://127.0.0.1:${PORT}/api/config`);
      if (r.ok) {
        console.log(`Server riêng cho bài test: cổng ${PORT}, database tạm ${db}`);
        return { proc, db };
      }
    } catch {
      /* chưa lên, thử lại */
    }
    await sleep(500);
  }
  proc.kill();
  throw new Error(`Không dựng được server test ở cổng ${PORT}:\n${log}`);
}

/** chơi một tab: vào game với tên cho trước rồi trả về các hàm điều khiển. */
async function openTab(browser, name, slow, full) {
  // SwiftShader tô TỪNG ĐIỂM ẢNH bằng CPU. 10 tab ở 960×540 ăn sạch CPU máy đo, và khi CPU
  // hết thì Chrome bỏ cả việc xử lý gói WebSocket — lượt đo đầu tiên ra ping 400–2000 ms và
  // băng thông 0.0 KB/s, tức là ta đang đo cái máy đo chứ không đo game.
  //
  // Vùng nhìn của server (AOI) tính từ config, KHÔNG phụ thuộc kích thước cửa sổ trình
  // duyệt, nên thu nhỏ cửa sổ không làm sai một byte nào của số liệu mạng. Vì vậy: chỉ MỘT
  // tab chạy khổ thật để lấy FPS / draw call, chín tab còn lại thu nhỏ và hạ chất lượng,
  // chúng chỉ có nhiệm vụ tạo đúng lượng người và lượng kiếm trên server.
  const ctx = await browser.newContext({ viewport: full ? { width: 960, height: 540 } : { width: 320, height: 180 } });
  if (!full) {
    await ctx.addInitScript(() => localStorage.setItem('kx.quality', 'low'));
  }
  const page = await ctx.newPage();
  page.on('pageerror', (e) => console.log(`  [${name}] lỗi trang: ${String(e).slice(0, 160)}`));
  page.on('console', (m) => {
    if (m.type() === 'error') console.log(`  [${name}] console: ${m.text().slice(0, 160)}`);
  });
  await page.goto(BASE, { waitUntil: 'domcontentloaded' });
  await page.waitForSelector('input', { timeout: 180000 });
  await page.fill('input', name);
  await page.getByText('Vào chơi', { exact: true }).click();
  await page.waitForFunction(() => typeof window.__stats === 'function', null, { timeout: 180000 }).catch((e) => {
    throw new Error(`${name} không vào được game. Nếu console báo HTTP 429 thì server đang chặn số kết nối mỗi IP — chạy server với -max-conns-per-ip 40. (${e.message})`);
  });
  if (slow) {
    // Bóp băng thông SAU khi đã vào game: ta đo lưu lượng lúc CHƠI, không đo lúc tải trang.
    // Bóp từ đầu thì 10 tab phải tải ~500 KB JS + hàng chục MB ảnh qua 1.6 Mbps, mất hàng
    // phút và chẳng nói lên điều gì về netcode.
    const cdp = await ctx.newCDPSession(page);
    await cdp.send('Network.enable');
    await cdp.send('Network.emulateNetworkConditions', SLOW_4G);
  }
  return { ctx, page, name, held: new Set() };
}

/** hold giữ đúng tập phím mong muốn, nhả những phím không còn cần. */
async function hold(tab, want) {
  for (const k of [...tab.held]) {
    if (!want.has(k)) {
      await tab.page.keyboard.up(k).catch(() => {});
      tab.held.delete(k);
    }
  }
  for (const k of want) {
    if (!tab.held.has(k)) {
      await tab.page.keyboard.down(k).catch(() => {});
      tab.held.add(k);
    }
  }
}

/** wander: đổi hướng ngẫu nhiên, thỉnh thoảng giữ Shift (Thiêu Đốt Tinh Huyết). */
async function wander(tab) {
  const want = new Set();
  const h = ['left', 'right', null][Math.floor(Math.random() * 3)];
  const v = ['up', 'down', null][Math.floor(Math.random() * 3)];
  if (h) want.add(DIRS[h]);
  if (v) want.add(DIRS[v]);
  if (want.size === 0) want.add(DIRS.right);
  if (Math.random() < 0.3) want.add('ShiftLeft');
  await hold(tab, want);
}

/** steer: đi về phía (tx, ty) bằng chính phím di chuyển, như người thật. */
async function steer(tab, tx, ty) {
  const me = await tab.page.evaluate(() => {
    const m = window.kx?.world?.me?.();
    return m ? { x: m.x, y: m.y } : null;
  }).catch(() => null);
  if (!me) {
    return null;
  }
  const dx = tx - me.x;
  const dy = ty - me.y;
  const want = new Set();
  if (Math.abs(dx) > 60) want.add(dx > 0 ? DIRS.right : DIRS.left);
  if (Math.abs(dy) > 60) want.add(dy > 0 ? DIRS.down : DIRS.up);
  await hold(tab, want);
  return Math.hypot(dx, dy);
}

const sample = (tab) => tab.page.evaluate(() => window.__stats()).catch(() => null);

function stats(values) {
  if (!values.length) {
    return { avg: 0, p95: 0, worst: 0 };
  }
  const v = [...values].sort((a, b) => a - b);
  return {
    avg: v.reduce((s, x) => s + x, 0) / v.length,
    p95: v[Math.min(v.length - 1, Math.floor(v.length * 0.95))],
    worst: v[v.length - 1],
  };
}

/**
 * summarise: với hầu hết chỉ số thì "xấu" = lớn (p95 lấy đuôi trên, tệ nhất = max). Riêng
 * FPS thì ngược lại — xấu = nhỏ, nên lấy p5 và min.
 */
function summarise(rows, key, higherIsBetter = false) {
  const v = rows.map((r) => r[key] ?? 0).sort((a, b) => a - b);
  if (!v.length) {
    return { avg: 0, p95: 0, worst: 0 };
  }
  const avg = v.reduce((s, x) => s + x, 0) / v.length;
  if (higherIsBetter) {
    return { avg, p95: v[Math.floor(v.length * 0.05)], worst: v[0] };
  }
  return { avg, p95: v[Math.min(v.length - 1, Math.floor(v.length * 0.95))], worst: v[v.length - 1] };
}

async function runOnce(label, slow) {
  console.log(`\n=== Lượt ${label} ===`);
  const browser = await chromium.launch({
    headless: true,
    // Chrome mặc định chỉ cho ~16 ngữ cảnh WebGL cùng lúc rồi giết bớt cái cũ; 10 tab +
    // ngữ cảnh phụ là chạm trần, nên phải nới ra.
    args: [
      '--use-gl=swiftshader',
      '--enable-unsafe-swiftshader',
      '--disable-dev-shm-usage',
      '--max-active-webgl-contexts=32',
      '--disable-frame-rate-limit',
    ],
  });
  const tabs = [];
  for (let i = 0; i < TABS; i++) {
    // Tab 1 là tab "mắt người": khổ thật, chất lượng thật, chỉ nó mới cho số FPS / draw
    // call đáng đọc. Nó cũng nằm trong nhóm Đạo Tổ nên thấy đúng cảnh nặng nhất.
    const full = i === 0;
    const t = await openTab(browser, `Test${String(i + 1).padStart(2, '0')}`, slow, full);
    t.full = full;
    tabs.push(t);
    console.log(`  ${t.name} đã vào`);
  }

  const samples = new Map(tabs.map((t) => [t.name, []]));
  let phase = 'wander';
  const collector = setInterval(async () => {
    for (const t of tabs) {
      const s = await sample(t);
      if (s) {
        samples.get(t.name).push({ ...s, phase });
      }
    }
  }, 1000);

  // Giai đoạn 1: đi lang thang
  console.log(`  đi lang thang ${WARMUP_S}s...`);
  const wanderUntil = Date.now() + WARMUP_S * 1000;
  while (Date.now() < wanderUntil) {
    await Promise.all(tabs.map(wander));
    await sleep(1000 + Math.random() * 1000);
  }

  // Giai đoạn 2: 5 tab lên Đạo Tổ đủ kiếm, rồi tất cả chụm về giữa bản đồ
  const tien = tabs.slice(0, TIEN_TABS);
  // Cấp Đạo Tổ lấy từ config chứ không viết cứng: đổi số cảnh giới trong config/game.json
  // là bài test tự theo, không âm thầm kiểm sai ngưỡng.
  const cult = (await (await fetch(`http://127.0.0.1:${PORT}/api/config`)).json()).cultivation;
  const TOP_LEVEL = (cult.realms.length + cult.tien_realms.length - 1) * cult.tiers_per_realm;
  // n: 1 = giữ bất tử. Không có nó thì 5 người Đạo Tổ chụm vào mài nhau chết sạch trong
  // chục giây, cả phòng tụt về Luyện Khí 1 kiếm và số đo thành vô nghĩa (đã dính một lần).
  const boost = () => Promise.all(tien.map((t) =>
    t.page.evaluate(() => window.kx.conn?.send?.({ t: 'dbg', cmd: 'max', n: 1 })).catch(() => {})));
  // Gửi LẠI mỗi vòng chứ không gửi một lần: lúc 10 tab đang tranh CPU, một lệnh đi lạc là
  // cả bảng số sai kịch bản (đã dính: cột cảnh giới ra 0.0 trong khi lẽ ra phải là 130).
  for (let i = 0; i < 20; i++) {
    await boost();
    await sleep(500);
    const levels = await Promise.all(tien.map((t) =>
      t.page.evaluate(() => window.kx?.world?.me?.()?.L ?? -1).catch(() => -1)));
    if (levels.every((L) => L >= TOP_LEVEL)) {
      console.log(`  xác nhận ${tien.length} tab đã lên Đạo Tổ`);
      break;
    }
    if (i === 19) {
      console.log(`  CẢNH BÁO: chỉ ${levels.filter((L) => L >= TOP_LEVEL).length}/${tien.length} tab lên được Đạo Tổ — bảng số sẽ không đúng kịch bản (cấp đọc được: ${levels.join(', ')})`);
    }
  }
  console.log(`  ${tien.length} tab lên Đạo Tổ, chụm lại đánh nhau ${FIGHT_S}s...`);
  phase = 'fight';
  const target = await tabs[0].page.evaluate(() => {
    const m = window.kx.world.me();
    return { x: m.x, y: m.y };
  });
  const fightUntil = Date.now() + FIGHT_S * 1000;
  let shot = false;
  let beat = 0;
  while (Date.now() < fightUntil) {
    await Promise.all(tabs.map((t) => steer(t, target.x, target.y)));
    // Nạp lại kiếm mỗi 2s: dù bất tử, va kiếm vẫn mài rụng ~120 kiếm/giây khi 5 người chụm
    // vào nhau, nên không nạp lại thì chỉ sau mươi giây là hết cảnh nặng cần đo.
    if (++beat % 4 === 0) {
      await boost();
    }
    await sleep(500);
    if (!shot && Date.now() > fightUntil - (FIGHT_S * 1000) / 3) { // ~2/3 pha, lúc đông nhất
      await mkdir(resolve(REPO, 'docs/perf'), { recursive: true });
      await tien[0].page.screenshot({ path: resolve(REPO, `docs/perf/browsertest-${label}.png`) });
      shot = true;
    }
  }

  clearInterval(collector);
  await browser.close();

  const rows = [];
  for (const [name, list] of samples) {
    for (const s of list) {
      rows.push({ ...s, name, tien: tien.some((t) => t.name === name), full: tabs.find((t) => t.name === name)?.full === true });
    }
  }
  return { label, rows };
}

function table(rows, title) {
  const fight = rows.filter((r) => r.phase === 'fight');
  const group = (want) => fight.filter((r) => r.tien === want);
  const out = [`**${title}** — ${fight.length} mẫu trong giai đoạn đánh nhau\n`];
  out.push('| Chỉ số | 5 tab Đạo Tổ (TB / p95 / tệ nhất) | 5 tab thường (TB / p95 / tệ nhất) |');
  out.push('|---|---|---|');
  // Chỉ tab 1 chạy khổ thật; chín tab kia thu nhỏ để không ăn hết CPU máy đo, nên số vẽ
  // của chúng vô nghĩa. Các chỉ số VẼ lấy riêng từ tab 1, các chỉ số MẠNG lấy cả nhóm.
  const eye = fight.filter((r) => r.full);
  const lineEye = (name, key, unit, higherIsBetter = false) => {
    if (!eye.length) {
      out.push(`| ${name} | — | — |`);
      return;
    }
    const s = summarise(eye, key, higherIsBetter);
    out.push(`| ${name} | ${s.avg.toFixed(1)} / ${s.p95.toFixed(1)} / ${s.worst.toFixed(1)}${unit ? ' ' + unit : ''} | *(chỉ đo ở tab 1)* |`);
  };
  const line = (name, key, unit, higherIsBetter = false) => {
    const cell = (g) => {
      if (!g.length) return '—';
      const s = summarise(g, key, higherIsBetter);
      return `${s.avg.toFixed(1)} / ${s.p95.toFixed(1)} / ${s.worst.toFixed(1)}${unit ? ' ' + unit : ''}`;
    };
    out.push(`| ${name} | ${cell(group(true))} | ${cell(group(false))} |`);
  };
  lineEye('FPS *(TB / p5 / thấp nhất — xem ghi chú)*', 'fps', '', true);
  line('Cảnh giới (cấp)', 'myLevel', '');
  line('Kiếm của chính tab', 'mySwords', '');
  lineEye('Thời gian vẽ', 'drawMs', 'ms');
  line('Ping', 'ping', 'ms');
  line('Băng thông xuống', 'kbPerSec', 'KB/s');
  line('Kích thước snapshot', 'snapKB', 'KB');
  line('Tick server', 'srvMs', 'ms');
  lineEye('Số sprite', 'sprites', '');
  lineEye('Draw call', 'drawCalls', '');
  line('Entity trong AOI', 'aoi', '');
  line('Tổng kiếm thấy trên màn hình', 'swords', '');
  return out.join('\n');
}

const server = SPAWN ? await startServer() : null;
const runs = [];
try {
  runs.push(await runOnce('mang-binh-thuong', false));
  runs.push(await runOnce('4g-yeu', true));
} finally {
  if (server) {
    server.proc.kill('SIGTERM');
    await rm(server.db, { force: true }).catch(() => {});
    console.log('Đã tắt server test và xoá database tạm.');
  }
}


const doc = [
  '# Kết quả đo 10 tab Chrome headless',
  '',
  `Chạy lại: \`cd tools/browsertest && node run.mjs\` — bài test **tự dựng server riêng ở`,
  `cổng ${PORT} với database tạm** rồi tắt khi xong, không đụng tới server đang có người chơi.`,
  `${TABS} browser context riêng, mỗi context một người chơi; ${WARMUP_S}s đi lang thang rồi`,
  `${TIEN_TABS} tab lên Đạo Tổ đủ kiếm và cả ${TABS} tab chụm một chỗ ${FIGHT_S}s.`,
  '',
  '> **FPS trong bảng này không dùng để kết luận được.** Chrome headless vẽ WebGL bằng CPU',
  '> (SwiftShader, không đụng tới card đồ họa) và 10 tab tranh nhau CPU trên cùng một máy.',
  '> Số đáng tin: **ping, KB/s, kích thước snapshot, tick server, số sprite, draw call**.',
  '>',
  '> Lượt B bóp băng thông **sau khi đã vào game**: ta đo lưu lượng lúc chơi, không đo lúc',
  '> tải trang.',
  '',
  table(runs[0].rows, 'Lượt A — mạng bình thường (LAN nội bộ)'),
  '',
  table(runs[1].rows, 'Lượt B — giả lập 4G yếu (trễ 100 ms, 1.6 Mbps xuống, 750 Kbps lên)'),
  '',
  `Ảnh chụp lúc đông nhất: \`docs/perf/browsertest-mang-binh-thuong.png\` và`,
  '`docs/perf/browsertest-4g-yeu.png`.',
  '',
].join('\n');

await mkdir(resolve(REPO, 'docs/perf'), { recursive: true });
await writeFile(resolve(REPO, 'docs/perf/BROWSERTEST.md'), doc, 'utf8');
console.log(`\n${doc}`);
console.log('\n→ docs/perf/BROWSERTEST.md');
