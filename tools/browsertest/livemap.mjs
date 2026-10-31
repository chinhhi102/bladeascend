// Kiểm A4 — xem bản đồ (ADMIN_SPEC §5), gồm cả lời hứa §11 về hiệu năng.
//
//   node tools/browsertest/livemap.mjs [cổng] [mật khẩu]
//
// Mục đáng giá nhất ở đây là mục cuối: §11 nói "admin xem KHÔNG làm tăng tick ms người
// chơi quá 0,3 ms". Cách duy nhất biết được là ĐO — đo tick khi không ai xem, rồi đo lại
// khi 2 admin đang kéo 5 lần/giây, rồi trừ.
import { webkit } from 'playwright';
import { decode, encode } from '../../client/node_modules/@msgpack/msgpack/dist.esm/index.mjs';

const PORT = Number(process.argv[2] ?? 9997);
const PASSWORD = process.argv[3] ?? 'mat-khau-thu-nghiem-123';
const BASE = `http://127.0.0.1:${PORT}`;

let failures = 0;
const check = (ok, label) => {
  console.log(`  ${ok ? '✓' : '✗'} ${label}`);
  if (!ok) failures++;
};
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

// Một người chơi thật, để khu có gì mà vẽ.
function join(name = 'ThuBanDo') {
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(`ws://127.0.0.1:${PORT}/ws`);
    ws.binaryType = 'arraybuffer';
    let done = false;
    ws.onopen = () => ws.send(encode({ t: 'join', name, char: 'a_hon', guest_id: 'livemap-test' }));
    ws.onmessage = (ev) => {
      const m = decode(new Uint8Array(ev.data));
      if (!done && m.t === 'welcome') { done = true; resolve({ ws, wel: m }); }
    };
    ws.onerror = reject;
    setTimeout(() => !done && reject(new Error('hết giờ chờ vào trận')), 12000);
  });
}

const login = await fetch(`${BASE}/api/admin/login`, {
  method: 'POST', headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ password: PASSWORD }),
});
const cookie = /kx_admin=[^;]+/.exec(login.headers.get('set-cookie') || '')?.[0];
if (!cookie) { console.log('  ✗ không đăng nhập được admin'); process.exit(1); }
const get = (p) => fetch(`${BASE}${p}`, { headers: { Cookie: cookie } });

const player = await join();
await sleep(800);
// Server xếp người vào khu ÍT NGƯỜI NHẤT (§8.0), nên không đoán được zone. Lấy từ gói
// `welcome`; đoán `zone=0` thì có lần trúng khu rỗng và bài kiểm báo "0 chấm người".
// Tên trường msgpack là `zn`/`room`, không phải `zone` — xem `netproto.Welcome`.
const ZONE = player.wel.zn ?? 0;
const ROOM = player.wel.room ?? '';
console.log(`  (người thử vào ${ROOM} khu ${ZONE} — ${player.wel.znm || player.wel.map})`);

// ── 1. Chưa đăng nhập thì không xem được ─────────────────────────────────────────────────
const anon = await fetch(`${BASE}/api/admin/live`);
check(anon.status === 401, `người lạ không xem được bản đồ trực tiếp (HTTP ${anon.status})`);
const anonJs = await fetch(`${BASE}/admin/map.js`, { redirect: 'manual' });
check(anonJs.status === 302, `mã trang bản đồ cũng sau cửa đăng nhập (HTTP ${anonJs.status})`);

// ── 2. Gói xem trực tiếp có đủ thứ §5.1 đòi ──────────────────────────────────────────────
const v = await (await get(`/api/admin/live?room=${ROOM}&zone=${ZONE}`)).json();
check(!!v.map && v.size > 0, `gói có bản đồ và kích thước (${v.map}, ${v.size})`);
check(Array.isArray(v.players) && v.players.length > 0, `có chấm người (${v.players?.length})`);
const human = v.players.find((p) => !p.bot);
check(!!human, 'phân biệt được người thật với bot');
check(!!human?.name, `người thật có tên ("${human?.name}")`);
check(human?.reach > 0, `có tầm kiếm để vẽ vòng mờ (${Math.round(human?.reach)})`);
check(human?.sc > 0, `có cỡ nhân vật (scale ${human?.sc})`);
check(Array.isArray(v.broken), 'có danh sách vật cản đang bị chém');
check(v.items && v.monsters && v.n_items > 0, `có quái và viên Tiên khí (${v.n_items} viên, ${v.n_monsters} quái)`);

// Bố cục tĩnh KHÔNG gửi lại 5 lần/giây — nó phải đến từ `/api/map/{id}`.
check(v.obstacles === undefined,
  'gói trực tiếp KHÔNG chở lại bố cục vật cản (tiết kiệm băng thông §5.1)');
const sizeKB = JSON.stringify(v).length / 1024;
check(sizeKB < 80, `gói gọn: ${sizeKB.toFixed(1)} KB mỗi lần (5 lần/giây)`);

// ── 3. Trần 2 người xem cùng lúc (§5.1) ──────────────────────────────────────────────────
//
// Gọi 6 request CÙNG LÚC: trần đếm request đang chạy, nên phải bắn song song mới chạm.
const burst = await Promise.all(Array.from({ length: 6 }, () => get(`/api/admin/live?room=${ROOM}&zone=${ZONE}`)));
const limited = burst.filter((r) => r.status === 429).length;
check(limited > 0, `quá 2 người xem cùng lúc thì bị chặn (${limited}/6 nhận 429)`);
check(burst.some((r) => r.ok), 'nhưng không chặn nhầm tất cả — vẫn có request đi qua');

// ── 4. Báo cáo --check cho trang xem bố cục tĩnh (§5.2) ──────────────────────────────────
const rep = await get('/api/admin/check');
if (rep.ok) {
  const j = await rep.json();
  const ids = Object.keys(j.maps || {});
  check(ids.length === 5, `báo cáo --check có đủ 5 khu (${ids.length})`);
  check((j.maps[ids[0]] || []).length === 14, `mỗi khu đủ 14 cảnh giới (${j.maps[ids[0]]?.length})`);
  check(j.targets && Object.keys(j.targets).length === 5, 'có mục tiêu mật độ từng khu');
} else {
  check(false, 'có báo cáo --check (chạy: cd server && go run ./cmd/mapgen --check)');
}

// ── 5. §11: admin xem KHÔNG làm tick người chơi tăng quá 0,3 ms ──────────────────────────
//
// Đo thật, không đoán. Nghỉ cho tick ổn định, đọc mốc nền, rồi bật 2 luồng xem ở 5 lần/giây
// đúng như §5.1 cho phép, đọc lại, trừ.
const tickNow = async () => {
  const st = await (await get('/api/admin/stats')).json();
  return st.server.tick_avg_max;
};
await sleep(2500);
let base = 0;
for (let i = 0; i < 6; i++) { base = Math.max(base, await tickNow()); await sleep(400); }

let stop = false;
const viewer = async () => {
  while (!stop) { await get(`/api/admin/live?room=${ROOM}&zone=${ZONE}`).then((r) => r.text()).catch(() => {}); await sleep(200); }
};
const viewers = [viewer(), viewer()];
await sleep(2500);
let load = 0;
for (let i = 0; i < 6; i++) { load = Math.max(load, await tickNow()); await sleep(400); }
stop = true;
await Promise.all(viewers);

const delta = load - base;
check(delta <= 0.3,
  `2 admin xem 5 lần/giây làm tick tăng ${delta.toFixed(3)} ms (trần §11: 0,3) — nền ${base.toFixed(3)} → ${load.toFixed(3)}`);

// ── 6. Trang vẽ được, và tên người chơi không thành mã ───────────────────────────────────
const br = await webkit.launch({ headless: true });
const ctx = await br.newContext();
await ctx.addCookies([{ name: 'kx_admin', value: cookie.split('=')[1], domain: '127.0.0.1', path: '/' }]);
const page = await ctx.newPage();
const errs = [];
page.on('pageerror', (e) => errs.push(String(e)));
await page.goto(`${BASE}/admin/#map`, { waitUntil: 'domcontentloaded' });
await page.waitForSelector('#mpcv', { timeout: 15000 });
await page.waitForTimeout(2500);
check(errs.length === 0, `trang bản đồ chạy không lỗi${errs.length ? ': ' + errs[0] : ''}`);
// Canvas có vẽ gì chưa: đếm điểm ảnh khác nền.
const painted = await page.evaluate(() => {
  const c = document.getElementById('mpcv');
  const d = c.getContext('2d').getImageData(0, 0, c.width, c.height).data;
  let n = 0;
  for (let i = 3; i < d.length; i += 4 * 97) if (d[i] > 0) n++;
  return n;
});
check(painted > 100, `bản đồ đã vẽ ra canvas (${painted} mẫu có màu)`);
// Bộ lọc loại vật phẩm: Linh khí phải TẮT SẴN, vì 6 000 chấm phủ kín bản đồ.
const filters = await page.locator('#mpitems label').count();
check(filters >= 2, `có bộ lọc theo loại vật phẩm (${filters} loại)`);
const orbOn = await page.evaluate(() => {
  for (const l of document.querySelectorAll('#mpitems label')) {
    if (l.textContent.includes('Linh khí')) return l.querySelector('input').checked;
  }
  return null;
});
check(orbOn === false, 'Linh khí tắt sẵn (bật lên là 6 000 chấm che hết vật cản)');
const tienOn = await page.evaluate(() => {
  for (const l of document.querySelectorAll('#mpitems label')) {
    if (l.textContent.includes('Tiên khí')) return l.querySelector('input').checked;
  }
  return null;
});
check(tienOn === true, 'Tiên khí bật sẵn — hiếm, và là thứ đáng nhìn');

// Bật Linh khí lên phải vẽ THÊM: chứng minh bộ lọc có tác dụng thật, không chỉ là ô tích.
const paintedOf = () => page.evaluate(() => {
  const c = document.getElementById('mpcv');
  const d = c.getContext('2d').getImageData(0, 0, c.width, c.height).data;
  let n = 0;
  for (let i = 0; i < d.length; i += 4 * 13) if (d[i] > 60 && d[i + 2] > 60) n++;
  return n;
});
const before = await paintedOf();
await page.evaluate(() => {
  for (const l of document.querySelectorAll('#mpitems label')) {
    if (l.textContent.includes('Linh khí')) { l.querySelector('input').click(); return; }
  }
});
await page.waitForTimeout(600);
const after = await paintedOf();
// Ngưỡng 1,05: Linh khí vẽ 1–3 px nên nó thêm ~14% mẫu, không phải gấp đôi. Đặt 1,2 là
// đòi hỏi vô lý với một phép đếm điểm ảnh thưa — nó trượt dù bộ lọc chạy đúng.
check(after > before * 1.05, `bật Linh khí lên thì bản đồ dày thêm thật (${before} → ${after} mẫu, +${Math.round((after / before - 1) * 100)}%)`);

// §5.2: thanh trượt cảnh giới.
await page.selectOption('#mpmode', 'seed');
await page.waitForTimeout(800);
check(!(await page.locator('#mprlab').isHidden()), 'chế độ tĩnh hiện thanh trượt cảnh giới r=0..13');
await page.locator('#mpr').fill('13');
await page.waitForTimeout(500);
const txt = await page.locator('#mpcheck').textContent();
check(/Đạo Tổ/.test(txt || ''), `kéo tới r=13 thì hiện số liệu --check của Đạo Tổ ("${(txt || '').slice(0, 60)}…")`);
check((await page.evaluate(() => document.querySelectorAll('#map script, #map img').length)) === 0,
  'trang bản đồ không dựng thẻ nào từ dữ liệu — tên người chơi vẽ bằng fillText');
await br.close();
player.ws.close();

console.log(failures === 0 ? '\nĐẠT: A4 đúng §5, và admin xem không làm chậm người chơi.'
  : `\nTRƯỢT: ${failures} mục.`);
process.exit(failures === 0 ? 0 : 1);
