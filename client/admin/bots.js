// Điều khiển bot — ADMIN_SPEC §4, GAME_SPEC §9.5. Bản tối giản (JS thuần) cho tới khi trang
// quản trị chuyển sang React + antd. Dùng chung `el` / `api` của admin.js; không innerHTML.

const BOT_MODES = [['auto', 'Tự động'], ['on', 'Bật'], ['off', 'Tắt']];
const BOT_LABEL = { auto: 'Tự động', on: 'Bật', off: 'Tắt', '': 'Theo toàn server' };

/** botAsk: xác nhận + lý do (ghi vào admin_audit). null = huỷ. */
function botAsk(what) {
  const reason = prompt(`${what}\n\nLý do (ghi vào nhật ký quản trị, có thể để trống):`, '');
  return reason === null ? null : reason.trim();
}

async function botPost(path, body) {
  try {
    await api(path, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
  } catch (e) {
    console.warn('[admin] bot', path, e);
  }
  loadBots().catch(() => {});
}

function botModeButtons(cur, modes, onPick) {
  const wrap = el('span', 'botmodes');
  for (const [m, label] of modes) {
    const b = el('button', m === cur ? 'on' : '', label);
    b.disabled = m === cur;
    b.addEventListener('click', () => onPick(m, label));
    wrap.append(b);
  }
  return wrap;
}

/** botEffective: "Tự động — đang tắt vì 43 người" (ADMIN_SPEC §4). */
function botEffective(r) {
  const base = BOT_LABEL[r.effective] || r.effective;
  if (r.effective !== 'auto') return base;
  return r.auto_off ? `${base} — đang tắt vì ${r.humans} người` : `${base} — đang bật`;
}

function botRoomRow(r) {
  const row = el('tr');
  const modes = [['', 'Theo toàn server'], ...BOT_MODES];
  const pick = (m, label) => {
    const reason = botAsk(`Đổi chế độ bot phòng ${r.id} sang "${label}"?`);
    if (reason !== null) botPost('/api/admin/bots/mode', { room: r.id, mode: m, reason });
  };
  const wd = el('button', '', 'Rút bot ngay');
  wd.addEventListener('click', () => {
    const reason = botAsk(`Rút bot phòng ${r.id}? Bot không giao tranh trong 5 giây sẽ rời ngay, bot đang đánh rút sau.`);
    if (reason !== null) botPost('/api/admin/bots/withdraw', { room: r.id, reason });
  });
  row.append(el('td', '', r.id), el('td', '', r.ok ? botEffective(r) : '(phòng không trả lời)'),
    el('td', '', `${r.humans} / ${r.bots} / ${r.humans + r.bots}`), el('td', '', `giới hạn ${r.max_humans}`),
    el('td', '', r.withdrawing || ''));
  const ctl = el('td');
  ctl.append(botModeButtons(r.override, modes, pick), wd);
  row.append(ctl);
  return row;
}

async function loadBots() {
  const box = document.getElementById('botpanel');
  if (!box) return;
  const s = await api('/api/admin/bots');
  const head = el('div', 'botglobal');
  head.append(el('b', '', 'Bot toàn server: '), botModeButtons(s.mode, BOT_MODES, (m, label) => {
    const reason = botAsk(`Đổi chế độ bot TOÀN SERVER sang "${label}"?`);
    if (reason !== null) botPost('/api/admin/bots/mode', { room: '', mode: m, reason });
  }));
  const t = el('table', 'tbl');
  const hr = el('tr');
  for (const h of ['Phòng', 'Đang áp dụng', 'Người / bot / tổng', 'Trần', 'Chờ rút', 'Đổi']) hr.append(el('th', '', h));
  t.append(hr, ...s.rooms.map(botRoomRow));
  box.replaceChildren(head, t);
}

setInterval(() => {
  if (!document.getElementById('rooms').hidden) loadBots().catch(() => {});
}, 2000);
