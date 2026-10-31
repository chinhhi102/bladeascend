// Trang quản trị — ADMIN_SPEC §2, §3, §4, §7.
//
// Viết bằng JS thuần, không qua Vite: trang này phục vụ từ `client/admin/` và CHỈ sau khi
// đăng nhập (xem `adminPageGuard` bên server). Gói nó vào bundle của game là đưa mã quản
// trị ra công khai, đúng thứ §1 cấm.
//
// LUẬT TUYỆT ĐỐI CỦA FILE NÀY: không bao giờ `innerHTML` với dữ liệu động.
// Tên người chơi và nội dung góp ý là thứ người chơi tự nhập (§1 "XSS"). Mọi chỗ đổ dữ
// liệu đều đi qua `textContent`, và `el()` bên dưới là cách duy nhất dựng DOM ở đây.

/** el dựng một phần tử; `text` luôn đi vào textContent, không bao giờ innerHTML. */
function el(tag, cls, text) {
  const n = document.createElement(tag);
  if (cls) n.className = cls;
  if (text !== undefined) n.textContent = String(text);
  return n;
}

/** api gọi API quản trị; 401 thì về trang đăng nhập. */
async function api(path, opts = {}) {
  const o = { credentials: 'same-origin', ...opts, headers: { ...(opts.headers || {}) } };
  if (o.method && o.method !== 'GET') o.headers['X-Admin-Token'] = csrf;
  const r = await fetch(path, o);
  if (r.status === 401) { location.href = '/admin/login'; throw new Error('401'); }
  if (!r.ok) throw new Error('HTTP ' + r.status);
  return r.json();
}

let csrf = '';

/** fmt gọn vài con số hay dùng. */
const f1 = (n) => (Math.round(n * 10) / 10).toFixed(1);

// ── Dashboard ────────────────────────────────────────────────────────────────────────────

/** card dựng một ô chỉ số; `level` quyết màu (xanh/vàng/đỏ) theo ngưỡng §3. */
function card(key, value, level) {
  const c = el('div', 'card' + (level ? ' ' + level : ''));
  c.append(el('div', 'k', key), el('div', 'v', value));
  return c;
}

/** tickLevel: ngưỡng màu cho tick ms — §3 nói "tô đỏ nếu > 4 ms". */
function tickLevel(ms) {
  if (ms > 4) return 'bad';
  if (ms > 2.5) return 'warn';
  return 'ok';
}

function renderDashboard(s) {
  const cards = document.getElementById('cards');
  cards.replaceChildren(
    card('Người thật online', s.players.humans),
    card('Bot', s.players.bots),
    card('Phòng', s.players.rooms),
    card('Tick TB cao nhất (ms)', f1(s.server.tick_avg_max), tickLevel(s.server.tick_avg_max)),
    card('Tick p95 cao nhất (ms)', f1(s.server.tick_p95_max), tickLevel(s.server.tick_p95_max)),
    card('Goroutine', s.server.goroutines),
    card('Heap (MB)', f1(s.server.heap_mb)),
    card('Người Tiên giới', s.tien_count),
    card('Uptime', uptime(s.uptime_secs)),
  );
  const realms = document.getElementById('realms');
  const keys = Object.keys(s.realms || {}).map(Number).sort((a, b) => a - b);
  realms.replaceChildren(...(keys.length
    ? keys.map((r) => el('span', '', `${realmName(r)}: ${s.realms[r]}`))
    : [el('span', '', 'chưa có người thật nào')]));
  document.getElementById('tick').textContent =
    'cập nhật ' + new Date(s.at * 1000).toLocaleTimeString('vi-VN');
  document.getElementById('build').textContent = s.build;
}

function uptime(secs) {
  const h = Math.floor(secs / 3600), m = Math.floor((secs % 3600) / 60);
  return h > 0 ? `${h}h${String(m).padStart(2, '0')}` : `${m}m`;
}

/**
 * REALMS phải khớp `cultivation.realms` + `tien_realms` trong config.
 *
 * Chép ra đây thay vì gọi `/api/config`: dashboard là trang nội bộ, không nên phụ thuộc một
 * request nữa chỉ để đổi số thành chữ. Lệch bảng thì chỉ sai NHÃN trên một trang quản trị,
 * không ảnh hưởng luật chơi — nên đánh đổi này an toàn.
 */
const REALMS = ['Luyện Khí', 'Trúc Cơ', 'Kim Đan', 'Nguyên Anh', 'Hóa Thần', 'Luyện Hư',
  'Hợp Thể', 'Độ Kiếp', 'Đại Thừa', 'Chân Tiên', 'Kim Tiên', 'Thái Ất Ngọc Tiên',
  'Đại La Kim Tiên', 'Đạo Tổ'];
const realmName = (r) => REALMS[r] ?? `cảnh giới ${r}`;

// ── Phòng & Khu ──────────────────────────────────────────────────────────────────────────

function renderRooms(s) {
  const box = document.getElementById('roomlist');
  if (!s.rooms || s.rooms.length === 0) {
    box.replaceChildren(el('p', '', 'Chưa có phòng nào.'));
    return;
  }
  box.replaceChildren(...s.rooms.map((r) => {
    const wrap = el('div', 'room');
    wrap.append(el('h4', '', `${r.id} — ${r.humans} người, ${r.bots} bot`));
    const t = el('table');
    const head = el('tr');
    for (const h of ['Khu', 'Người', 'Bot', 'Tick TB', 'p95', 'Cảnh giới cao nhất',
      'Quái', 'Vật phẩm', 'Vật cản đang chém']) head.append(el('th', '', h));
    t.append(el('thead').appendChild(head).parentNode);
    const body = el('tbody');
    for (const z of r.zones || []) {
      const tr = el('tr');
      tr.append(el('td', '', z.name));
      for (const [v, cls] of [[z.humans, 'num'], [z.bots, 'num'],
        [f1(z.tick_avg), 'num'], [f1(z.tick_p95), 'num']]) tr.append(el('td', cls, v));
      tr.append(el('td', '', z.top_realm || '—'));
      for (const v of [z.monsters, z.items, z.broken_obstacles]) tr.append(el('td', 'num', v));
      body.append(tr);
    }
    t.append(body);
    wrap.append(t);
    return wrap;
  }));
}

// ── Nhật ký ──────────────────────────────────────────────────────────────────────────────

async function loadAudit() {
  const rows = await api('/api/admin/audit');
  const body = document.querySelector('#audittbl tbody');
  body.replaceChildren(...rows.map((e) => {
    const tr = el('tr');
    tr.append(
      el('td', '', new Date(e.at * 1000).toLocaleString('vi-VN')),
      el('td', '', e.action), el('td', '', e.target),
      el('td', '', e.reason || '—'), el('td', '', e.ip));
    return tr;
  }));
}

// ── Hòm thư góp ý (§7.2) ─────────────────────────────────────────────────────────────────

const FB_KINDS = { bug: 'Báo lỗi', balance: 'Cân bằng', idea: 'Ý tưởng', other: 'Khác' };
const FB_STATUS = { moi: 'Mới', da_xem: 'Đã xem', dang_lam: 'Đang làm', xong: 'Xong', bo_qua: 'Bỏ qua' };

/** fillSelect đổ options vào một <select>, giữ lại mục "mọi ..." đầu tiên. */
function fillSelect(el, map) {
  if (el.options.length > 1) return;
  for (const [v, label] of Object.entries(map)) {
    const o = document.createElement('option');
    o.value = v; o.textContent = label;
    el.appendChild(o);
  }
}

/**
 * fbCard dựng một lá thư.
 *
 * MỌI chỗ đổ dữ liệu đều `textContent` (qua `el()`): nội dung và tên là thứ NGƯỜI CHƠI tự
 * nhập. §1 nói rõ, và §11 có một mục kiểm riêng cho nó — gửi `<script>alert(1)</script>`
 * thì phải thấy nguyên văn mấy chữ đó.
 */
function fbCard(f) {
  const box = el('div', 'fb');
  const head = el('div', 'head');
  head.append(
    el('span', 'tag', FB_KINDS[f.kind] ?? f.kind),
    el('span', 'tag', FB_STATUS[f.status] ?? f.status),
    el('span', '', new Date(f.at * 1000).toLocaleString('vi-VN')),
    el('span', '', f.name || '(không tên)'),
    el('span', '', f.guest_id.slice(0, 8)),
  );
  box.append(head, el('p', 'body', f.body));

  if (f.context && f.context !== '{}') {
    const c = el('div', 'ctx');
    try {
      const o = JSON.parse(f.context);
      // Lỗi console in riêng từng dòng cho dễ đọc; phần còn lại gộp một dòng.
      const errs = Array.isArray(o.console) ? o.console : null;
      delete o.console;
      c.textContent = Object.entries(o).map(([k, v]) => `${k}=${v}`).join(' · ');
      if (errs && errs.length) {
        c.append(el('div', '', `— ${errs.length} dòng lỗi —`));
        for (const line of errs) c.append(el('div', '', line));
      }
    } catch (_) { c.textContent = f.context; }
    box.append(c);
  }
  if (f.has_image) {
    const img = document.createElement('img');
    img.src = `/api/admin/feedback/${f.id}/image`;
    img.alt = 'ảnh chụp màn hình';
    img.loading = 'lazy';
    box.append(img);
  }

  const tools = el('div', 'tools');
  const st = document.createElement('select');
  for (const [v, label] of Object.entries(FB_STATUS)) {
    const o = document.createElement('option');
    o.value = v; o.textContent = label; o.selected = v === f.status;
    st.appendChild(o);
  }
  const note = document.createElement('input');
  note.placeholder = 'Ghi chú riêng của admin...';
  note.value = f.note || '';
  const save = el('button', '', 'Lưu');
  save.addEventListener('click', async () => {
    save.disabled = true;
    await api(`/api/admin/feedback/${f.id}`, {
      method: 'PATCH', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ status: st.value, note: note.value }),
    }).catch(() => {});
    save.disabled = false;
    loadFeedback();
  });
  tools.append(st, note, save);
  box.append(tools);
  return box;
}

async function loadFeedback() {
  const kind = document.getElementById('fbkind');
  const status = document.getElementById('fbstatus');
  fillSelect(kind, FB_KINDS);
  fillSelect(status, FB_STATUS);
  const q = new URLSearchParams();
  if (kind.value) q.set('kind', kind.value);
  if (status.value) q.set('status', status.value);
  const text = document.getElementById('fbq').value.trim();
  if (text) q.set('q', text);
  const rows = await api('/api/admin/feedback?' + q);
  const list = document.getElementById('fblist');
  list.replaceChildren(...(rows.length ? rows.map(fbCard) : [el('p', '', 'Chưa có góp ý nào.')]));
}

for (const id of ['fbkind', 'fbstatus']) {
  document.getElementById(id).addEventListener('change', () => loadFeedback().catch(() => {}));
}
document.getElementById('fbq').addEventListener('input', () => {
  clearTimeout(window.__fbT);
  window.__fbT = setTimeout(() => loadFeedback().catch(() => {}), 300);
});

// ── Người chơi, cờ nghi hack, kick và cấm (§6) ───────────────────────────────────────────
//
// Mọi nút ở đây đều hỏi lý do và hỏi xác nhận. Không có nút nào chạy mà không có người bấm.

/** BAN_SPANS: thời hạn cấm §6.2 cho phép. 0 = chỉ kick, -1 = vĩnh viễn. */
const BAN_SPANS = [
  { secs: 0, label: 'Chỉ mời ra (vào lại được ngay)' },
  { secs: 3600, label: 'Cấm 1 giờ' },
  { secs: 86400, label: 'Cấm 24 giờ' },
  { secs: 604800, label: 'Cấm 7 ngày' },
  { secs: -1, label: 'Cấm vĩnh viễn' },
];

let plRows = [];

/** secs rút gọn số giây thành "2 phút" / "1 giờ 3 phút". */
function secs(n) {
  const m = Math.floor(n / 60);
  if (m < 60) return m + ' phút';
  return Math.floor(m / 60) + ' giờ ' + (m % 60) + ' phút';
}

function plCard(p) {
  const box = el('div', 'pl' + (p.flags.length ? ' flagged' : ''));
  const head = el('div', 'head');
  head.append(
    el('strong', '', p.name || '(chưa vào trận)'),
    el('span', 'tag', p.character || '?'),
    el('span', '', p.guest_id.slice(0, 8)),
    el('span', '', p.device),
    el('span', '', p.ip_mask),
  );
  // Kết nối CÓ quyền lệnh test thì nói rõ — đó là thông tin quan trọng nhất về một dòng
  // đang có điểm cao bất thường, và nó giải thích luôn cờ "Hành động bị từ chối".
  if (p.debug_source) {
    head.append(el('span', 'tag dbg', 'lệnh test: ' + p.debug_source + (p.debug_used ? ' (đã dùng)' : '')));
  }
  box.append(head);

  const stats = el('div', 'stats');
  for (const [k, v] of [
    ['Khu', p.zone || '—'], ['Cảnh giới', `${p.realm || '—'} (tầng ${p.level})`],
    ['Kiếm', p.swords], ['Uy Danh', Math.round(p.score)], ['Chém trúng', p.hits],
    ['Sống', secs(p.alive_secs)], ['Ping', Math.round(p.ping_ms) + ' ms'],
    ['FPS', Math.round(p.fps)], ['Hiệu chỉnh/phút', p.corrections],
  ]) {
    const s = el('span');
    s.append(el('i', '', k), document.createTextNode(' ' + v));
    stats.append(s);
  }
  box.append(stats);

  for (const f of p.flags) {
    const fl = el('div', 'flag');
    fl.append(
      el('span', 'ftag', '⚠ ' + f.name + (f.count > 1 ? ` ×${f.count}` : '')),
      el('span', 'fev', f.evidence),
    );
    const x = el('button', 'small', 'Bỏ cờ');
    x.title = 'Đánh dấu nhầm — bỏ cờ này đi';
    x.addEventListener('click', async () => {
      x.disabled = true;
      await api('/api/admin/clearflag', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ guest_id: p.guest_id, kind: f.kind }),
      }).catch((e) => alert('Không xong: ' + e.message));
      loadPlayers().catch(() => {});
    });
    fl.append(x);
    box.append(fl);
  }

  const tools = el('div', 'tools');
  const span = document.createElement('select');
  for (const s of BAN_SPANS) {
    const o = document.createElement('option');
    o.value = String(s.secs); o.textContent = s.label;
    span.appendChild(o);
  }
  const go = el('button', 'danger', 'Mời ra');
  span.addEventListener('change', () => {
    go.textContent = span.value === '0' ? 'Mời ra' : 'Mời ra + cấm';
  });
  go.addEventListener('click', async () => {
    const banSecs = Number(span.value);
    const what = BAN_SPANS.find((s) => s.secs === banSecs).label;
    // Lý do BẮT BUỘC, và người chơi sẽ ĐỌC được nó trên màn hình của họ — nên nhắc rõ
    // ở đây, để không ai gõ một câu chỉ mình hiểu.
    const reason = prompt(`${what}\nNgười chơi: ${p.name} (${p.guest_id.slice(0, 8)})\n\n`
      + 'Lý do (bắt buộc — NGƯỜI CHƠI SẼ ĐỌC ĐƯỢC câu này):', '');
    if (reason === null) return;
    if (!reason.trim()) { alert('Phải nêu lý do.'); return; }
    if (!confirm(`${what}\n\n"${p.name}"\nLý do: ${reason}\n\nXác nhận?`)) return;
    go.disabled = true;
    try {
      const r = await api('/api/admin/kick', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ guest_id: p.guest_id, reason, ban_secs: banSecs }),
      });
      alert(`Đã ngắt ${r.kicked} kết nối.`);
    } catch (e) { alert('Không xong: ' + e.message); }
    go.disabled = false;
    loadPlayers().catch(() => {});
  });
  tools.append(span, go);
  box.append(tools);
  return box;
}

function renderPlayers() {
  const onlyFlag = document.getElementById('plflag').checked;
  const q = document.getElementById('plq').value.trim().toLowerCase();
  const rows = plRows.filter((p) => {
    if (onlyFlag && !p.flags.length) return false;
    if (q && !((p.name || '').toLowerCase().includes(q) || p.guest_id.toLowerCase().includes(q))) return false;
    return true;
  });
  const list = document.getElementById('pllist');
  list.replaceChildren(...(rows.length
    ? rows.map(plCard)
    : [el('p', '', plRows.length ? 'Không ai khớp bộ lọc.' : 'Chưa có người thật nào online.')]));
  const nf = plRows.filter((p) => p.flags.length).length;
  document.getElementById('plcount').textContent =
    `${plRows.length} người online · ${nf} có cờ · hiện ${rows.length}`;
}

async function loadPlayers() {
  plRows = await api('/api/admin/players');
  renderPlayers();
}

document.getElementById('plflag').addEventListener('change', renderPlayers);
document.getElementById('plq').addEventListener('input', () => {
  clearTimeout(window.__plT);
  window.__plT = setTimeout(renderPlayers, 200);
});

// ── Danh sách cấm (§6.2) ─────────────────────────────────────────────────────────────────

async function loadBans() {
  const rows = await api('/api/admin/bans');
  const tb = document.querySelector('#bantbl tbody');
  if (!rows.length) {
    const tr = el('tr'); const td = el('td', '', 'Chưa cấm ai.'); td.colSpan = 6;
    tr.append(td); tb.replaceChildren(tr);
    return;
  }
  tb.replaceChildren(...rows.map((b) => {
    const tr = el('tr');
    tr.append(
      el('td', '', new Date(b.from * 1000).toLocaleString('vi-VN')),
      el('td', '', b.guest_id.slice(0, 12) || '—'),
      el('td', '', b.ip || '—'),
      el('td', '', b.until ? new Date(b.until * 1000).toLocaleString('vi-VN') : 'vĩnh viễn'),
      el('td', '', b.reason),
    );
    const cell = el('td');
    const btn = el('button', '', 'Gỡ cấm');
    btn.addEventListener('click', async () => {
      if (!confirm(`Gỡ cấm ${b.guest_id.slice(0, 12)}?\nLý do cấm: ${b.reason}`)) return;
      btn.disabled = true;
      await api('/api/admin/unban', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ guest_id: b.guest_id, reason: 'gỡ thủ công' }),
      }).catch((e) => alert('Không xong: ' + e.message));
      loadBans().catch(() => {});
    });
    cell.append(btn);
    tr.append(cell);
    return tr;
  }));
}

// ── Bảng xếp hạng: rà soát và ẩn (§6.2) ──────────────────────────────────────────────────
//
// Nút duy nhất ở đây là "Ẩn". Không có nút xoá, và sẽ không có — xem `internal/store/hidden.go`.

/**
 * SIM_SECS: thời gian chơi THẬT để đạt mốc, đo bằng `cmd/sim` (40 bot, 20 phút mô phỏng).
 * Đây là mốc duy nhất đáng tin: nó đo trên chính luật chơi đang chạy, nên khi cân bằng đổi
 * thì chỉ cần chạy lại sim rồi sửa hai con số này.
 */
const SIM_SECS = { 80: 429, 90: 479 };

/** suspectWhy trả về lý do ngờ một mạng chơi, hoặc '' nếu không có gì bất thường. */
function suspectWhy(r) {
  const s = Math.max(r.survive_secs, 0.1);
  if (r.peak_level >= 90 && s < SIM_SECS[90] / 2) {
    return `Tiên giới (tầng ${r.peak_level}) trong ${Math.round(s)}s — chơi thật cần ~${SIM_SECS[90]}s`;
  }
  if (r.peak_level >= 80 && s < SIM_SECS[80] / 2) {
    return `Đại Thừa (tầng ${r.peak_level}) trong ${Math.round(s)}s — chơi thật cần ~${SIM_SECS[80]}s`;
  }
  return '';
}

let bdRows = [];

function bdRow(r) {
  const why = suspectWhy(r);
  const tr = el('tr', [r.hidden ? 'hid' : '', why ? 'sus' : ''].filter(Boolean).join(' '));
  const name = el('td');
  name.append(el('span', '', r.name || '(không tên)'), el('small', '', ' ' + r.guest_id.slice(0, 8)));
  tr.append(
    el('td', '', new Date(r.ended_at * 1000).toLocaleString('vi-VN')),
    name,
    el('td', '', Math.round(r.score)),
    el('td', '', r.peak_level),
    el('td', '', f1(r.survive_secs) + 's'),
    el('td', '', r.hidden ? 'ĐANG ẨN — ' + r.hide_reason : why),
  );

  const cell = el('td');
  const b = el('button', '', r.hidden ? 'Bỏ ẩn' : 'Ẩn khỏi bảng');
  b.addEventListener('click', async () => {
    let reason = '';
    if (!r.hidden) {
      // Lý do là BẮT BUỘC, và gợi ý sẵn bằng chứng đã tính được — ba tháng sau đọc lại
      // "nghi hack" thì chẳng còn nghĩa gì, còn "Đạo Tổ trong 63 giây" thì vẫn kiểm được.
      reason = prompt('Lý do ẩn (bắt buộc, sẽ lưu vào nhật ký):', why || '');
      if (reason === null) return;
      if (!reason.trim()) { alert('Phải nêu lý do.'); return; }
    } else if (!confirm(`Bỏ ẩn mạng chơi của "${r.name}"? Nó sẽ hiện lại trên bảng.`)) {
      return;
    }
    b.disabled = true;
    try {
      await api('/api/admin/hide', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          guest_id: r.guest_id, ended_at: r.ended_at, reason, unhide: !!r.hidden,
        }),
      });
    } catch (e) { alert('Không xong: ' + e.message); }
    b.disabled = false;
    loadBoard().catch(() => {});
  });
  cell.append(b);
  tr.append(cell);
  return tr;
}

function renderBoard() {
  const onlySus = document.getElementById('bdsus').checked;
  const onlyHid = document.getElementById('bdhid').checked;
  const q = document.getElementById('bdq').value.trim().toLowerCase();
  const rows = bdRows.filter((r) => {
    if (onlySus && !suspectWhy(r) && !r.hidden) return false;
    if (onlyHid && !r.hidden) return false;
    if (q && !(r.name.toLowerCase().includes(q) || r.guest_id.toLowerCase().includes(q))) return false;
    return true;
  });
  const tb = document.querySelector('#bdtbl tbody');
  tb.replaceChildren(...(rows.length
    ? rows.map(bdRow)
    : [(() => { const tr = el('tr'); const td = el('td', '', 'Không có dòng nào khớp.'); td.colSpan = 7; tr.append(td); return tr; })()]));
  const nSus = bdRows.filter((r) => suspectWhy(r)).length;
  const nHid = bdRows.filter((r) => r.hidden).length;
  document.getElementById('bdcount').textContent =
    `${bdRows.length} mạng gần nhất · ${nSus} đáng ngờ · ${nHid} đang ẩn · hiện ${rows.length}`;
}

async function loadBoard() {
  bdRows = await api('/api/admin/lives?limit=500');
  renderBoard();
}

for (const id of ['bdsus', 'bdhid']) {
  document.getElementById(id).addEventListener('change', renderBoard);
}
document.getElementById('bdq').addEventListener('input', () => {
  clearTimeout(window.__bdT);
  window.__bdT = setTimeout(renderBoard, 200);
});

// ── Khung trang ──────────────────────────────────────────────────────────────────────────

function show(id) {
  for (const sec of document.querySelectorAll('main > section')) sec.hidden = sec.id !== id;
  for (const a of document.querySelectorAll('#side a')) a.classList.toggle('on', a.hash === '#' + id);
  if (id === 'audit') loadAudit().catch(() => {});
  if (id === 'feedback') loadFeedback().catch(() => {});
  if (id === 'board') loadBoard().catch(() => {});
  if (id === 'players') loadPlayers().catch(() => {});
  if (id === 'bans') loadBans().catch(() => {});
}

window.addEventListener('hashchange', () => show(location.hash.slice(1) || 'dashboard'));

document.getElementById('logout').addEventListener('click', async () => {
  await api('/api/admin/logout', { method: 'POST' }).catch(() => {});
  location.href = '/admin/login';
});

async function tick() {
  try {
    const s = await api('/api/admin/stats');
    renderDashboard(s);
    renderRooms(s);
    // Số góp ý "Mới" ở thanh bên (§7.2).
    const pill = document.getElementById('fbcount');
    pill.hidden = !s.feedback_new;
    pill.textContent = s.feedback_new ?? 0;
    // Số người đang có cờ, ở thanh bên. Nếu đang mở đúng trang Người chơi thì làm mới
    // luôn danh sách — nhưng CHỈ khi đang mở: kéo 500 dòng mỗi 2 giây cho một trang
    // không ai nhìn là tốn băng thông và làm goroutine phòng bận không lý do.
    const flagPill = document.getElementById('flagcount');
    flagPill.hidden = !s.flagged;
    flagPill.textContent = s.flagged ?? 0;
    if (!document.getElementById('players').hidden) loadPlayers().catch(() => {});
  } catch (_) { /* 401 đã tự chuyển trang; lỗi mạng thì bỏ qua một nhịp */ }
}

(async () => {
  const me = await api('/api/admin/me');
  csrf = me.csrf;
  show(location.hash.slice(1) || 'dashboard');
  await tick();
  // §3: dashboard cập nhật mỗi 2 giây.
  setInterval(tick, 2000);
})();
