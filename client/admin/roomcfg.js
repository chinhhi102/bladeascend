// Cấu hình phòng — ADMIN_SPEC v1.4 §4, GAME_SPEC v6.28 §9.5. Bản JS thuần cho tới khi trang
// quản trị chuyển sang React + antd. Dùng chung `el` / `api` của admin.js; không innerHTML.

const ROOMCFG_FIELDS = [
  ['max_humans', 'Trần người thật / phòng', 'Bot không tính. Hạ trần không đuổi ai — phòng đang vượt chỉ ngừng nhận người mới.'],
  ['zone_min_chars', 'Sàn nhân vật mỗi khu', 'Bot lấp mỗi khu cho đủ số này khi được phép sinh.'],
  ['bot_auto_off_at', 'Tự tắt sinh bot khi người thật ≥', 'Bật lại khi dưới ngưỡng − 5. 0 = không tự tắt.'],
];

let roomcfgLoaded = false;

function roomcfgWarn(v, lt) {
  return lt > 0 && v > lt ? `⚠ Cao hơn mức đã qua loadtest của máy này (${lt} người) — tick có thể vượt 4 ms khi đông.` : '';
}

function renderRoomcfg(box, s) {
  const card = el('div', 'roomcfg');
  card.append(el('h3', '', 'Cấu hình phòng'),
    el('div', 'muted', `Môi trường: ${s.env} · mặc định ${s.defaults.max_humans} / ${s.defaults.zone_min_chars} / ${s.defaults.bot_auto_off_at} · đã qua loadtest: ${s.loadtested_humans} người/phòng`));
  const inputs = {};
  const warn = el('div', 'roomcfg-warn', roomcfgWarn(s.settings.max_humans, s.loadtested_humans));
  for (const [key, label, hint] of ROOMCFG_FIELDS) {
    const inp = el('input');
    inp.type = 'number';
    inp.min = '0';
    inp.value = String(s.settings[key]);
    if (key === 'max_humans') {
      inp.addEventListener('input', () => { warn.textContent = roomcfgWarn(Number(inp.value), s.loadtested_humans); });
    }
    inputs[key] = inp;
    const row = el('label', 'roomcfg-row');
    row.append(el('span', '', label), inp, el('small', 'muted', hint));
    card.append(row);
  }
  const save = el('button', 'primary', 'Lưu');
  const msg = el('span', 'muted', '');
  save.addEventListener('click', async () => {
    const body = {};
    for (const [key] of ROOMCFG_FIELDS) body[key] = Math.round(Number(inputs[key].value));
    const reason = prompt(`Áp dụng ngay: trần ${body.max_humans} · sàn ${body.zone_min_chars}/khu · tắt bot ở ${body.bot_auto_off_at}?\n\nLý do (ghi vào nhật ký quản trị):`, '');
    if (reason === null) return;
    try {
      const next = await api('/api/admin/room-settings', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ ...body, reason: reason.trim() }) });
      renderRoomcfg(box, next);
    } catch (e) {
      msg.textContent = `Không lưu được: ${e.message || e}`;
    }
  });
  const actions = el('div', 'row');
  actions.append(save, msg);
  card.append(warn, actions);
  box.replaceChildren(card);
}

async function loadRoomcfg() {
  const box = document.getElementById('roomcfg');
  if (!box) return;
  renderRoomcfg(box, await api('/api/admin/room-settings'));
  roomcfgLoaded = true;
}

// Nạp MỘT lần khi mở tab Phòng (không làm mới theo nhịp 2 s như bảng bot — đang gõ số thì
// bị ghi đè mất).
setInterval(() => {
  if (!roomcfgLoaded && !document.getElementById('rooms').hidden) loadRoomcfg().catch((e) => console.warn('[admin] roomcfg', e));
}, 500);
