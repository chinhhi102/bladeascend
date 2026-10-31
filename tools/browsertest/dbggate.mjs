// Kiểm ADMIN_SPEC §1.1 trên server THẬT, ở mức GIAO THỨC.
//
//   node tools/browsertest/dbggate.mjs [cổng] [mật khẩu]
//
// Vì sao không dùng trình duyệt: `window.kx` chỉ được gắn khi server đã cấp quyền debug,
// nên ở ca "không có quyền" thì chẳng có gì để đọc — bài đo treo chờ một đối tượng không
// bao giờ xuất hiện. Mà thứ cần kiểm lại đúng là ca đó, và kiểm ở mức gói tin còn CHẶT hơn:
// nó gửi lệnh debug thẳng qua WebSocket, bỏ qua mọi kiểm tra phía client — đúng kiểu một
// client bị sửa sẽ làm.
import { encode, decode } from '../../client/node_modules/@msgpack/msgpack/dist.esm/index.mjs';

const PORT = Number(process.argv[2] ?? 9997);
const PASSWORD = process.argv[3] ?? 'mat-khau-thu-nghiem-123';
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

/** joinAndTryDebug: vào trận rồi gửi 5 lệnh "lên cảnh giới" (dbg realm). Trả về quyền + cảnh giới. */
async function joinAndTryDebug(label, cookie) {
  const ws = new WebSocket(`ws://127.0.0.1:${PORT}/ws`, {
    headers: cookie ? { Cookie: cookie } : {},
  });
  ws.binaryType = 'arraybuffer';
  let welcomeDebug = null, level = -1, lastLevel = -1, myId = 0;

  await new Promise((res, rej) => { ws.onopen = res; ws.onerror = rej; });
  ws.onmessage = (ev) => {
    const m = decode(new Uint8Array(ev.data));
    if (m.t === 'welcome') { welcomeDebug = !!m.debug; myId = m.id; }
    if (m.t === 'snap') {
      for (const p of m.p || []) if (p.id === myId) { lastLevel = p.L; if (level < 0) level = p.L; }
    }
  };
  ws.send(encode({ t: 'join', name: label.slice(0, 12), char: 'a_hon', guest_id: 'dbgtest-' + label, debug: true }));
  await sleep(1500);
  // Gửi lệnh debug THỦ CÔNG — client thật chỉ gửi sau khi server cấp quyền, nên gói này
  // đến từ "một client bị sửa".
  for (let i = 0; i < 5; i++) ws.send(encode({ t: 'dbg', cmd: 'realm', n: 1 }));
  await sleep(2000);
  ws.close();
  return { welcomeDebug, before: level, after: lastLevel };
}

/** adminCookie đăng nhập và trả về chuỗi cookie phiên. */
async function adminCookie() {
  const r = await fetch(`http://127.0.0.1:${PORT}/api/admin/login`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ password: PASSWORD }),
  });
  if (!r.ok) throw new Error('đăng nhập admin thất bại: HTTP ' + r.status);
  const m = /kx_admin=[^;]+/.exec(r.headers.get('set-cookie') || '');
  if (!m) throw new Error('không nhận được cookie phiên');
  return m[0];
}

const show = (label, r) =>
  console.log(`${label.padEnd(18)} server cấp debug = ${String(r.welcomeDebug).padEnd(5)} · ` +
    `cảnh giới ${r.before} → ${r.after}  ${r.after > r.before ? '⚠ LỆNH ĐÃ CHẠY' : '✓ lệnh bị từ chối'}`);

const anon = await joinAndTryDebug('KhongDangNhap', null);
show('KHÔNG đăng nhập', anon);

const admin = await joinAndTryDebug('CoPhienAdmin', await adminCookie());
show('CÓ phiên admin', admin);

const ok = anon.welcomeDebug === false && anon.after <= anon.before
  && admin.welcomeDebug === true && admin.after > admin.before;
console.log(ok
  ? '\nĐẠT: ?debug=true một mình không đủ; lệnh debug thủ công bị từ chối; phiên admin thì được.'
  : '\nTRƯỢT: §1.1 chưa đúng.');
process.exit(ok ? 0 : 1);
