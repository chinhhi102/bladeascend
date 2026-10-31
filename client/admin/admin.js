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

// ── Khung trang ──────────────────────────────────────────────────────────────────────────

function show(id) {
  for (const sec of document.querySelectorAll('main > section')) sec.hidden = sec.id !== id;
  for (const a of document.querySelectorAll('#side a')) a.classList.toggle('on', a.hash === '#' + id);
  if (id === 'audit') loadAudit().catch(() => {});
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
