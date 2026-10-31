// Kiểm A2 — người chơi, cờ nghi hack, kick và cấm (ADMIN_SPEC §6).
//
//   node tools/browsertest/kickban.mjs [cổng] [mật khẩu]
//
// Điều quan trọng nhất kiểm ở đây KHÔNG phải "nút kick chạy được" mà là:
//   1. server KHÔNG BAO GIỜ tự kick — gắp cờ đầy người mà kết nối vẫn sống;
//   2. kick/cấm không nêu lý do thì bị từ chối;
//   3. người bị kick ĐỌC ĐƯỢC lý do (gói `kick` chở `reason`);
//   4. người bị cấm không vào lại được, và biết tới bao giờ;
//   5. gỡ cấm thì vào lại được ngay.
import { webkit } from 'playwright';
// Cùng đường dẫn như `dbggate.mjs`: dùng msgpack CỦA CLIENT để chắc chắn hai bên mã hoá
// giống hệt nhau. `WebSocket` là bản dựng sẵn của Node 24, không cần gói ngoài.
import { decode, encode } from '../../client/node_modules/@msgpack/msgpack/dist.esm/index.mjs';

const PORT = Number(process.argv[2] ?? 9997);
const PASSWORD = process.argv[3] ?? 'mat-khau-thu-nghiem-123';
const BASE = `http://127.0.0.1:${PORT}`;
const GUEST = 'kickban-test-' + PORT;

let failures = 0;
const check = (ok, label) => {
  console.log(`  ${ok ? '✓' : '✗'} ${label}`);
  if (!ok) failures++;
};

/** join mở một kết nối và vào trận; trả về { ws, msgs, kicked }. */
function join(name = 'ThuKick', guest = GUEST) {
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(`ws://127.0.0.1:${PORT}/ws`);
    const msgs = [];
    let settled = false;
    ws.binaryType = 'arraybuffer';
    ws.onopen = () => ws.send(encode({ t: 'join', name, char: 'a_hon', guest_id: guest }));
    ws.onmessage = (ev) => {
      const m = decode(new Uint8Array(ev.data));
      msgs.push(m);
      if (!settled && (m.t === 'welcome' || m.t === 'kick')) {
        settled = true;
        resolve({ ws, msgs, first: m });
      }
    };
    ws.onerror = reject;
    setTimeout(() => !settled && reject(new Error('hết giờ chờ vào trận')), 12000);
  });
}
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const lastKick = (msgs) => [...msgs].reverse().find((m) => m.t === 'kick');

const login = await fetch(`${BASE}/api/admin/login`, {
  method: 'POST', headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ password: PASSWORD }),
});
const cookie = /kx_admin=[^;]+/.exec(login.headers.get('set-cookie') || '')?.[0];
if (!cookie) { console.log('  ✗ không đăng nhập được admin'); process.exit(1); }
const me = await (await fetch(`${BASE}/api/admin/me`, { headers: { Cookie: cookie } })).json();
const hdr = { Cookie: cookie, 'X-Admin-Token': me.csrf, 'Content-Type': 'application/json' };
const players = async () =>
  (await (await fetch(`${BASE}/api/admin/players`, { headers: { Cookie: cookie } })).json());

// ── 1. Vào trận và hiện ở bảng người chơi ────────────────────────────────────────────────
let conn = await join();
check(conn.first.t === 'welcome', 'vào trận được');
await sleep(500);
let me1 = (await players()).find((p) => p.guest_id === GUEST);
check(!!me1, 'hiện ở bảng người chơi của trang admin');
check(!!me1 && me1.device.includes('·'), `cột thiết bị có dữ liệu ("${me1?.device}")`);
check(!!me1 && /x\.x$|^$/.test(me1.ip_mask), `IP được che (${me1?.ip_mask})`);

// ── 2. Gắp cờ: gửi một đống gói rác, rồi xem kết nối CÓ SỐNG KHÔNG ───────────────────────
//
// Đây là mục quan trọng nhất của cả bài kiểm. §6.1: "chỉ gắn cờ, không tự kick".
for (let i = 0; i < 60; i++) conn.ws.send(Buffer.from([0xc1, 0xff, 0x00, i & 0xff]));
await sleep(1200);
me1 = (await players()).find((p) => p.guest_id === GUEST);
check(!!me1?.flags.length, `gói rác bị gắn cờ (${me1?.flags.map((f) => f.name).join(', ') || 'không có cờ nào'})`);
check(!!me1?.flags[0]?.evidence, `cờ có kèm bằng chứng ("${me1?.flags[0]?.evidence}")`);
check(conn.ws.readyState === WebSocket.OPEN && !lastKick(conn.msgs),
  'KẾT NỐI VẪN SỐNG — server không tự kick ai (§6.1)');

// ── 3. Bỏ cờ ─────────────────────────────────────────────────────────────────────────────
await fetch(`${BASE}/api/admin/clearflag`, {
  method: 'POST', headers: hdr,
  body: JSON.stringify({ guest_id: GUEST, kind: me1.flags[0].kind }),
});
await sleep(300);
const after = (await players()).find((p) => p.guest_id === GUEST);
check(!after?.flags.some((f) => f.kind === me1.flags[0].kind), 'bỏ được cờ đánh dấu nhầm');

// ── 4. Lý do bắt buộc ────────────────────────────────────────────────────────────────────
const noReason = await fetch(`${BASE}/api/admin/kick`, {
  method: 'POST', headers: hdr, body: JSON.stringify({ guest_id: GUEST, reason: '  ' }),
});
check(noReason.status === 400, `kick không nêu lý do thì bị từ chối (HTTP ${noReason.status})`);
const anon = await fetch(`${BASE}/api/admin/kick`, {
  method: 'POST', headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ guest_id: GUEST, reason: 'x' }),
});
check(anon.status === 401, `người lạ không kick được ai (HTTP ${anon.status})`);

// ── 5. Kick thật: người chơi phải ĐỌC ĐƯỢC lý do ─────────────────────────────────────────
const REASON = 'Thử nghiệm tự động — xin lỗi vì làm phiền';
const kicked = await (await fetch(`${BASE}/api/admin/kick`, {
  method: 'POST', headers: hdr, body: JSON.stringify({ guest_id: GUEST, reason: REASON }),
})).json();
check(kicked.kicked === 1, `kick ngắt đúng 1 kết nối (${kicked.kicked})`);
await sleep(600);
const kmsg = lastKick(conn.msgs);
check(!!kmsg, 'người chơi nhận được gói báo bị mời ra');
check(kmsg?.reason === REASON, `lý do tới được người chơi ("${kmsg?.reason}")`);
check(!kmsg?.until, 'kick thường không kèm hạn cấm — vào lại được ngay');

// Kick xong vào lại được ngay (§6.2).
conn = await join();
check(conn.first.t === 'welcome', 'kick xong vào lại được ngay');

// ── 6. Cấm 1 giờ: không vào lại được, và biết tới bao giờ ────────────────────────────────
const BAN_REASON = 'Thử nghiệm cấm tự động — sẽ gỡ ngay';
await fetch(`${BASE}/api/admin/kick`, {
  method: 'POST', headers: hdr,
  body: JSON.stringify({ guest_id: GUEST, reason: BAN_REASON, ban_secs: 3600 }),
});
await sleep(600);
const retry = await join();
check(retry.first.t === 'kick', 'người bị cấm KHÔNG vào trận được');
check(retry.first.reason === BAN_REASON, `biết vì sao bị cấm ("${retry.first.reason}")`);
check(retry.first.until > Math.floor(Date.now() / 1000), 'biết cấm tới bao giờ');
retry.ws.close();

const bans = await (await fetch(`${BASE}/api/admin/bans`, { headers: { Cookie: cookie } })).json();
const row = bans.find((b) => b.guest_id === GUEST);
check(!!row, 'lệnh cấm hiện ở danh sách cấm');
check(!!row && !/\d+\.\d+\.\d+\.\d+/.test(row.ip || ''), `IP trong danh sách cấm được che (${row?.ip})`);

const audit = await (await fetch(`${BASE}/api/admin/audit`, { headers: { Cookie: cookie } })).json();
check(audit.some((a) => a.action === 'kick_ban' && a.target === GUEST && a.reason === BAN_REASON),
  'kick + cấm để lại dấu kèm lý do trong nhật ký (§9)');

// ── 7. Gỡ cấm: vào lại được ──────────────────────────────────────────────────────────────
await fetch(`${BASE}/api/admin/unban`, {
  method: 'POST', headers: hdr, body: JSON.stringify({ guest_id: GUEST, reason: 'dọn sau khi kiểm' }),
});
await sleep(400);
const back = await join();
check(back.first.t === 'welcome', 'gỡ cấm xong vào lại được (đã dọn)');
back.ws.close();

// ── 8. Trang admin: tên người chơi không chạy được mã ────────────────────────────────────
//
// Tên tối đa 16 ký tự (`name_max`), nên không nhét vừa `onerror=`. Nhưng vừa đủ cho một
// thẻ HTML — và thẻ HTML là đủ: nếu trang admin dựng DOM bằng innerHTML thì `<img src=x>`
// sẽ thành một thẻ img thật, và chỗ nào nhận được `<img>` thì chỗ đó cũng nhận `<script>`
// từ một đường khác. Kiểm bằng thứ nhét vừa, chứ không bỏ hẳn phép kiểm.
const xss = await join('<img src=x>abc', GUEST + '-xss');
await sleep(400);
const br = await webkit.launch({ headless: true });
const ctx = await br.newContext();
await ctx.addCookies([{ name: 'kx_admin', value: cookie.split('=')[1], domain: '127.0.0.1', path: '/' }]);
const page = await ctx.newPage();
await page.goto(`${BASE}/admin/#players`, { waitUntil: 'domcontentloaded' });
await page.waitForSelector('#pllist .pl', { timeout: 15000 });
check((await page.locator('#pllist').evaluate((n) => n.querySelectorAll('img').length)) === 0,
  'tên có thẻ HTML KHÔNG biến thành thẻ thật trên trang admin');
check((await page.locator('#pllist').textContent()).includes('<img src=x>'),
  'tên hiển thị nguyên văn cho admin đọc');
const labels = await page.locator('#pllist button').allTextContents();
check(labels.some((t) => /Mời ra/.test(t)), `có nút mời ra (${[...new Set(labels)].join(', ')})`);
check(!labels.some((t) => /xo[áa]/i.test(t)), 'không có nút xoá dữ liệu người chơi nào');
await br.close();
xss.ws.close();
conn.ws.close();

console.log(failures === 0 ? '\nĐẠT: A2 đúng §6 — chỉ gắn cờ, mọi thao tác thủ công và có lý do.'
  : `\nTRƯỢT: ${failures} mục.`);
process.exit(failures === 0 ? 0 : 1);
