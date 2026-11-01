// Kiểm "Ẩn khỏi bảng xếp hạng" trên server thật — ADMIN_SPEC §6.2.
//
//   node tools/browsertest/hidelife.mjs [cổng] [mật khẩu]
//
// Ba điều phải đúng, và chỉ chạy thật mới thấy được:
//   1. ẩn xong thì `/api/leaderboard` **bớt ngay** tên đó (giữa chúng có một lớp cache);
//   2. dữ liệu CÒN NGUYÊN — `/api/admin/lives` vẫn thấy nó, mang cờ `hidden`;
//   3. người chưa đăng nhập KHÔNG ẩn được của ai (và CSRF chặn request không có token).
import { webkit } from 'playwright';

const PORT = Number(process.argv[2] ?? 9997);
const PASSWORD = process.argv[3] ?? 'mat-khau-thu-nghiem-123';
const BASE = `http://127.0.0.1:${PORT}`;

let failures = 0;
const check = (ok, label) => {
  console.log(`  ${ok ? '✓' : '✗'} ${label}`);
  if (!ok) failures++;
};

const login = await fetch(`${BASE}/api/admin/login`, {
  method: 'POST', headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ password: PASSWORD }),
});
const cookie = /kx_admin=[^;]+/.exec(login.headers.get('set-cookie') || '')?.[0];
if (!cookie) {
  console.log('  ✗ không đăng nhập được admin');
  process.exit(1);
}
const me = await (await fetch(`${BASE}/api/admin/me`, { headers: { Cookie: cookie } })).json();
const hdr = { Cookie: cookie, 'X-Admin-Token': me.csrf, 'Content-Type': 'application/json' };

// Chọn mục tiêu TỪ CHÍNH BẢNG XẾP HẠNG, không phải từ danh sách mạng gần nhất.
//
// Bảng chỉ lấy 50 dòng cao điểm nhất; một mạng chơi mới toanh với 3 điểm thì không có trên
// đó, và ẩn nó đi thì bảng chẳng đổi gì — test sẽ "đạt" mà chẳng chứng minh được điều cần
// chứng minh. Nên: lấy dòng ĐỨNG ĐẦU bảng, rồi tra ngược guest_id của nó.
const board = async () => (await (await fetch(`${BASE}/api/leaderboard?period=all&board=score`)).json()).rows ?? [];
const top = (await board())[0];
const lives = await (await fetch(`${BASE}/api/admin/lives?limit=1000`, { headers: { Cookie: cookie } })).json();
const target = top && lives.find((l) => l.name === top.name && l.ended_at === top.ended_at && !l.hidden);
if (!target) {
  console.log('  ✗ không tra được guest_id của dòng đứng đầu bảng (DB trống? chơi một trận rồi chạy lại)');
  process.exit(1);
}
const onBoard = async () =>
  (await board()).some((x) => x.name === target.name && x.ended_at === target.ended_at);

// ── 1. Lý do là bắt buộc ─────────────────────────────────────────────────────────────────
const noReason = await fetch(`${BASE}/api/admin/hide`, {
  method: 'POST', headers: hdr,
  body: JSON.stringify({ guest_id: target.guest_id, ended_at: target.ended_at, reason: '  ' }),
});
check(noReason.status === 400, `ẩn mà không nêu lý do thì bị từ chối (HTTP ${noReason.status})`);

// ── 2. Chưa đăng nhập / thiếu CSRF thì không ẩn được ─────────────────────────────────────
const anon = await fetch(`${BASE}/api/admin/hide`, {
  method: 'POST', headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ guest_id: target.guest_id, ended_at: target.ended_at, reason: 'x' }),
});
check(anon.status === 401 || anon.status === 403, `người lạ không ẩn được (HTTP ${anon.status})`);
const noCsrf = await fetch(`${BASE}/api/admin/hide`, {
  method: 'POST', headers: { Cookie: cookie, 'Content-Type': 'application/json' },
  body: JSON.stringify({ guest_id: target.guest_id, ended_at: target.ended_at, reason: 'x' }),
});
check(noCsrf.status === 403, `có cookie nhưng thiếu X-Admin-Token thì bị chặn (HTTP ${noCsrf.status})`);

// ── 3. Ẩn thật ───────────────────────────────────────────────────────────────────────────
check(await onBoard(), `trước khi ẩn, "${target.name}" có trên bảng xếp hạng`);
const REASON = 'kiểm thử tự động — sẽ bỏ ẩn ngay';
const hid = await fetch(`${BASE}/api/admin/hide`, {
  method: 'POST', headers: hdr,
  body: JSON.stringify({ guest_id: target.guest_id, ended_at: target.ended_at, reason: REASON }),
});
check(hid.ok, `ẩn được (HTTP ${hid.status})`);
check(!(await onBoard()), 'ẩn xong thì mất khỏi bảng NGAY (cache đã xoá)');

const after = await (await fetch(`${BASE}/api/admin/lives?limit=1000`, { headers: { Cookie: cookie } })).json();
const row = after.find((l) => l.guest_id === target.guest_id && l.ended_at === target.ended_at);
check(!!row, 'dữ liệu CÒN NGUYÊN — ẩn không phải xoá');
check(!!row && row.hidden === true, 'dòng mang cờ hidden để bỏ ẩn được');
check(!!row && row.hide_reason === REASON, 'lý do được lưu lại');
check(after.length === lives.length, `tổng số mạng chơi không đổi (${lives.length} → ${after.length})`);

const audit = await (await fetch(`${BASE}/api/admin/audit`, { headers: { Cookie: cookie } })).json();
check(audit.some((a) => a.action === 'leaderboard_hide' && a.target === target.guest_id),
  'thao tác ẩn có để lại dấu trong nhật ký (§9)');

// ── 4. Trang admin hiện đúng, và tên người chơi không chạy được mã ────────────────────────
const br = await webkit.launch({ headless: true });
const ctx = await br.newContext();
await ctx.addCookies([{ name: 'kx_admin', value: cookie.split('=')[1], domain: '127.0.0.1', path: '/' }]);
const page = await ctx.newPage();
await page.goto(`${BASE}/admin/#board`, { waitUntil: 'domcontentloaded' });
await page.waitForSelector('#bdtbl tbody tr.ant-table-row', { timeout: 15000 });
check(!(await page.evaluate(() => window.__pwned)), 'tên người chơi không chạy được mã trên trang admin');
const hidRows = await page.locator('#bdtbl tr.hid').count();
check(hidRows >= 1, `dòng đang ẩn hiện rõ bằng lớp .hid (${hidRows} dòng)`);
check((await page.locator('#bdtbl tr.hid button').first().textContent()) === 'Bỏ ẩn',
  'dòng đang ẩn có nút "Bỏ ẩn"');
// Không được có nút nào xoá. Đây là lời hứa chính của cả tính năng, nên kiểm bằng chữ.
const labels = await page.locator('#bdtbl button').allTextContents();
check(!labels.some((t) => /xo[áa]/i.test(t)), `không có nút xoá nào (${[...new Set(labels)].join(', ')})`);
await br.close();

// ── 5. Dọn: bỏ ẩn lại để không làm bẩn DB thật ───────────────────────────────────────────
const un = await fetch(`${BASE}/api/admin/hide`, {
  method: 'POST', headers: hdr,
  body: JSON.stringify({ guest_id: target.guest_id, ended_at: target.ended_at, unhide: true }),
});
check(un.ok && (await onBoard()), 'bỏ ẩn trả lại nguyên trạng (đã dọn sau khi kiểm)');

console.log(failures === 0 ? '\nĐẠT: ẩn/bỏ ẩn bảng xếp hạng đúng §6.2.' : `\nTRƯỢT: ${failures} mục.`);
process.exit(failures === 0 ? 0 : 1);
