// Xem bản đồ — ADMIN_SPEC §5.
//
// Canvas 2D, không PixiJS, không art: §5 nói "nhẹ, không cần art". Trang quản trị phải mở
// được từ điện thoại qua 4G trong lúc server đang có vấn đề — đó là lúc người ta cần nó.
//
// Hai chế độ:
//   §5.1 trực tiếp — người, quái, viên, vật cản đang bị chém; 5 lần/giây; bám theo một người.
//   §5.2 tĩnh      — bố cục + kết quả `mapgen --check` + thanh trượt cảnh giới r = 0..13.
//
// **Chỉ xem.** Không có thao tác nào lên trận ở đây.

/** TIER_COLORS: màu theo bậc vật cản 0–13 (§5.1 "bảng chú thích bậc"). */
const TIER_COLORS = [
  '#6b7a52', '#7a8a58', '#8a9a5e', '#9aa864', '#a8b06a', '#9a8f6a', '#8a7f72',
  '#7a6f7a', '#6a6382', '#5a5a8a', '#6a5a92', '#7a5a9a', '#8a5aa2', '#9a5aaa',
];
const BORDER_TIER = 99;

/**
 * ITEM_NAMES: tên tiếng Việt và màu cho từng loại vật phẩm trên đất.
 *
 * `orb` là **Linh khí** (viên EXP thường, hàng nghìn viên mỗi khu) và `tien_khi` là
 * **Tiên khí** (hiếm, chỉ rơi quanh Đài). Hai thứ khác hẳn nhau về ý nghĩa lẫn số lượng,
 * nên chúng phải tắt/bật được RIÊNG: bật Linh khí lên là 6 000 chấm phủ kín bản đồ và
 * không còn thấy vật cản đâu nữa.
 */
const ITEM_NAMES = {
  orb: ['Linh khí', 'rgba(140,220,255,.5)'],
  tien_khi: ['Tiên khí', 'rgba(255,210,79,.9)'],
  sword: ['Kiếm', 'rgba(220,220,230,.75)'],
  bundle: ['Bó kiếm', 'rgba(220,200,255,.8)'],
  spin: ['Xoay nhanh', 'rgba(150,255,180,.8)'],
  haste: ['Chạy nhanh', 'rgba(150,255,180,.8)'],
  magnet: ['Nam châm', 'rgba(255,170,120,.8)'],
  hard: ['Cương Kiếm', 'rgba(255,150,150,.8)'],
  shield: ['Khiên', 'rgba(150,200,255,.8)'],
  exp2: ['EXP ×2', 'rgba(255,230,120,.8)'],
  stealth: ['Ẩn thân', 'rgba(190,190,210,.7)'],
  skill_scroll: ['Bí kíp', 'rgba(255,140,220,.85)'],
};
const itemName = (k) => (ITEM_NAMES[k] || [k, 'rgba(255,210,79,.55)'])[0];
const itemColor = (k) => (ITEM_NAMES[k] || [k, 'rgba(255,210,79,.55)'])[1];

/** hiddenItems: loại vật phẩm đang bị ẩn. Linh khí ẩn SẴN — xem ITEM_NAMES. */
const hiddenItems = new Set(['orb']);

const mp = {
  mode: 'live', room: '', zone: 0, map: null, check: null, live: null,
  realm: 0, follow: 0, timer: 0,
  // Khung nhìn: `k` là tỉ lệ, `ox/oy` là góc trên-trái đang hiện (toạ độ thế giới).
  k: 1, ox: 0, oy: 0, fitted: false,
  layers: { tier: true, zones: true, foot: false, reach: false, grid: false },
};

const cv = () => document.getElementById('mpcv');

/** fit đặt khung nhìn vừa khít bản đồ. */
function mpFit() {
  if (!mp.map) return;
  const c = cv();
  mp.k = Math.min(c.width, c.height) / mp.map.size;
  mp.ox = 0; mp.oy = 0;
  mp.fitted = true;
}
const sx = (x) => (x - mp.ox) * mp.k;
const sy = (y) => (y - mp.oy) * mp.k;

/**
 * mpDraw vẽ lại toàn bộ. Gọi mỗi lần có dữ liệu mới hoặc người dùng kéo/phóng.
 *
 * Vẽ lại TẤT CẢ mỗi khung thay vì vẽ chênh lệch: ~900 hình chữ nhật và vài trăm chấm, 5
 * lần/giây. Canvas 2D làm việc đó mà không kịp thở; còn giữ trạng thái "đã vẽ gì" là mở
 * cửa cho một lớp lỗi hiển thị không ai tìm ra.
 */
function mpDraw() {
  const c = cv(), g = c.getContext('2d');
  g.clearRect(0, 0, c.width, c.height);
  if (!mp.map) {
    g.fillStyle = '#9aa0a6'; g.font = '14px system-ui';
    g.fillText('Đang tải bản đồ...', 16, 28);
    return;
  }
  if (!mp.fitted) mpFit();
  const S = mp.map.size;

  // Nền + vùng chơi được.
  g.fillStyle = '#10131a';
  g.fillRect(sx(0), sy(0), S * mp.k, S * mp.k);
  const pm = mp.map.play_margin || 0;
  g.fillStyle = '#161a22';
  g.fillRect(sx(pm), sy(pm), (S - 2 * pm) * mp.k, (S - 2 * pm) * mp.k);

  if (mp.layers.grid) {
    g.strokeStyle = '#242a34'; g.lineWidth = 1;
    for (let v = 0; v <= S; v += 1000) {
      g.beginPath();
      g.moveTo(sx(v), sy(0)); g.lineTo(sx(v), sy(S));
      g.moveTo(sx(0), sy(v)); g.lineTo(sx(S), sy(v));
      g.stroke();
    }
  }

  // Vùng (bụi, bùn, băng, nước, đường). Hình chữ nhật nằm trong `s`, toạ độ GÓC TRÊN-TRÁI
  // — xem `/api/map/{id}`; `z.x` trần không tồn tại.
  if (mp.layers.zones) {
    for (const z of mp.map.zones || []) {
      g.fillStyle = {
        bush: 'rgba(90,160,90,.22)', mud: 'rgba(130,100,60,.22)',
        ice: 'rgba(120,190,230,.20)', water: 'rgba(70,120,200,.24)',
        road: 'rgba(170,150,110,.14)',
      }[z.k] || 'rgba(150,150,150,.12)';
      g.fillRect(sx(z.s.x), sy(z.s.y), z.s.w * mp.k, z.s.h * mp.k);
    }
  }

  // Vật cản. Ở chế độ tĩnh, vật cản bậc ≤ r coi như KHÔNG CÒN (người cảnh giới r chém
  // được) — đó chính là thứ thanh trượt dùng để xem: Đại Thừa còn lại bao nhiêu địa hình.
  // Mỗi vật cản có HAI hình chữ nhật, và phân biệt chúng là quan trọng:
  //   `s` = hình vẽ (thứ người chơi nhìn thấy, thứ kiếm chém trúng — §8.2);
  //   `f` = chân đế (thứ chặn đường đi).
  // Vẽ `s` tô đặc và `f` viền sáng hơn: hai con số mật độ của `--check` nói về hai thứ
  // khác nhau, nên trang này phải cho thấy được cả hai.
  const broken = new Set(mp.live?.broken || []);
  (mp.map.obstacles || []).forEach((o, i) => {
    const tier = o.t ?? 0;
    const cut = mp.mode === 'seed' && tier !== BORDER_TIER && tier <= mp.realm;
    if (cut) return;
    const r = o.s || o.f;
    if (!r) return;
    const x = sx(r.x), y = sy(r.y), w = r.w * mp.k, h = r.h * mp.k;
    g.fillStyle = mp.layers.tier
      ? (tier === BORDER_TIER ? '#434a55' : TIER_COLORS[Math.min(tier, 13)])
      : '#55606e';
    if (broken.has(i)) {
      // Đang bị chém: nét đứt, không tô đặc (§5.1).
      g.save();
      g.setLineDash([4, 3]); g.strokeStyle = '#ffcc66'; g.lineWidth = 1.5;
      g.strokeRect(x, y, w, h);
      g.restore();
    } else {
      g.fillRect(x, y, w, h);
      if (o.f && mp.layers.foot) {
        g.strokeStyle = 'rgba(255,255,255,.45)'; g.lineWidth = 1;
        g.strokeRect(sx(o.f.x), sy(o.f.y), o.f.w * mp.k, o.f.h * mp.k);
      }
    }
  });

  // Cổng (Thiên Môn ở tâm, bến ở mép) và Đài Tiên Khí. Cả hai chỉ có tâm + loại, không
  // có bề rộng — vẽ vòng tròn cỡ cố định trên màn hình để luôn thấy được khi thu nhỏ.
  for (const gate of mp.map.gates || []) {
    g.strokeStyle = gate.n === 'world' ? '#ffd24f' : '#7ad4ff';
    g.lineWidth = 2;
    g.beginPath(); g.arc(sx(gate.x), sy(gate.y), 9, 0, 7); g.stroke();
  }
  g.fillStyle = '#9ad7ff';
  for (const a of mp.map.altars || []) {
    g.beginPath(); g.arc(sx(a.x), sy(a.y), a.k === 'big' ? 7 : 5, 0, 7); g.fill();
  }

  if (mp.mode !== 'live' || !mp.live) { mpLegend(); return; }

  // Viên Tiên khí và quái đến dưới dạng mảng số phẳng theo loại, toạ độ đã chia cho `q`
  // (xem `liveGroup` bên server). Nhân lại khi vẽ.
  const Q = mp.live.q || 1;
  // Chấm viên vẽ NHỎ và MỜ, cỡ theo mức phóng to.
  //
  // Bản đầu vẽ 4×4 px đặc: 6 300 viên phủ kín bản đồ và không còn nhìn thấy vật cản đâu
  // nữa — đúng thứ trang này sinh ra để xem. Viên là NỀN, không phải nội dung chính.
  for (const [kind, arr] of Object.entries(mp.live.items || {})) {
    if (hiddenItems.has(kind)) continue;
    // Linh khí vẽ nhỏ nhất (nó là nền); thứ hiếm vẽ to hơn để nhìn ra giữa đám đông.
    const d = kind === 'orb'
      ? Math.max(1, Math.min(3, 30 * mp.k))
      : Math.max(3, Math.min(7, 90 * mp.k));
    g.fillStyle = itemColor(kind);
    for (let i = 0; i < arr.length; i += 2) {
      g.fillRect(sx(arr[i] * Q) - d / 2, sy(arr[i + 1] * Q) - d / 2, d, d);
    }
  }
  g.fillStyle = '#e06a6a';
  const mr = Math.max(2, Math.min(7, 120 * mp.k));
  for (const arr of Object.values(mp.live.monsters || {})) {
    for (let i = 0; i < arr.length; i += 2) {
      g.beginPath(); g.arc(sx(arr[i] * Q), sy(arr[i + 1] * Q), mr, 0, 7); g.fill();
    }
  }
  // Người: người thật sáng và có tên, bot xám (§5.1).
  for (const p of mp.live.players) {
    if (p.dead) continue;
    const r = Math.max(3, 24 * (p.sc || 1) * mp.k);
    if (mp.layers.reach && !p.bot) {
      g.strokeStyle = 'rgba(120,200,255,.3)'; g.lineWidth = 1;
      g.beginPath(); g.arc(sx(p.x), sy(p.y), p.reach * mp.k, 0, 7); g.stroke();
    }
    g.fillStyle = p.bot ? '#6b7280' : (p.id === mp.follow ? '#ffd24f' : '#7ad4ff');
    g.beginPath(); g.arc(sx(p.x), sy(p.y), r, 0, 7); g.fill();
    if (!p.bot) {
      g.fillStyle = '#e8eaed'; g.font = '11px system-ui'; g.textAlign = 'center';
      // Tên người chơi: `fillText` vẽ CHỮ, không dựng DOM — không có đường nào thành mã.
      g.fillText(p.name || '?', sx(p.x), sy(p.y) - r - 3);
      g.textAlign = 'left';
    }
  }
  mpLegend();
}

/**
 * mpLegend vẽ bảng chú thích bậc, ở ĐÁY canvas trên một dải nền.
 *
 * Trên nền trong suốt ở góc trên thì chữ chú thích nằm đè lên bản đồ và không đọc được —
 * nhất là khi bản đồ sáng màu ở chỗ đó.
 */
function mpLegend() {
  const c = cv(), g = c.getContext('2d');
  const H = 34, y0 = c.height - H;
  g.fillStyle = 'rgba(10,13,18,.88)';
  g.fillRect(0, y0, c.width, H);
  g.font = '10px system-ui';
  if (mp.layers.tier) {
    for (let t = 0; t <= 13; t++) {
      const x = 8 + t * 20;
      g.fillStyle = TIER_COLORS[t];
      g.fillRect(x, y0 + 6, 16, 9);
      g.fillStyle = '#c8ccd2';
      g.fillText(String(t), x + 4, y0 + 26);
    }
    g.fillStyle = '#9aa0a6';
    g.fillText('bậc vật cản — chém được khi cảnh giới > bậc', 300, y0 + 13);
  }
  if (mp.mode === 'live') {
    const marks = [['#7ad4ff', 'người'], ['#6b7280', 'bot'],
      ['#e06a6a', 'quái'], ['#ffcc66', 'vật cản đang bị chém (nét đứt)']];
    let x = 300;
    for (const [col, label] of marks) {
      g.fillStyle = col; g.fillRect(x, y0 + 21, 8, 8);
      g.fillStyle = '#9aa0a6'; g.fillText(label, x + 12, y0 + 28);
      x += 22 + label.length * 5.2;
    }
  }
}

// ── Lấy dữ liệu ──────────────────────────────────────────────────────────────────────────

async function mpLoadMap(id) {
  mp.map = await (await fetch(`/api/map/${encodeURIComponent(id)}`)).json();
  mp.fitted = false;
  mpDraw();
}

async function mpTickLive() {
  if (mp.mode !== 'live' || document.getElementById('map').hidden) return;
  try {
    const v = await api(`/api/admin/live?room=${encodeURIComponent(mp.room)}&zone=${mp.zone}`);
    mp.live = v;
    if (!mp.map || mp.map.id !== v.map) await mpLoadMap(v.map);
    // Bám theo: giữ người đó ở giữa khung (§5.1).
    if (mp.follow) {
      const p = v.players.find((x) => x.id === mp.follow);
      if (p) {
        const c = cv();
        mp.ox = p.x - c.width / 2 / mp.k;
        mp.oy = p.y - c.height / 2 / mp.k;
      }
    }
    mpItemFilters(v);
    document.getElementById('mapinfo').textContent =
      `${v.room} · ${v.zone} · ${v.players.filter((p) => !p.bot).length} người, `
      + `${v.players.filter((p) => p.bot).length} bot · ${v.n_items} viên · ${v.n_monsters} quái`
      + ` · ${v.broken.length} vật cản đang bị chém`;
    mpDraw();
  } catch (e) {
    mpItemFilters(v);
    document.getElementById('mapinfo').textContent = 'không đọc được: ' + e.message;
  }
}

/**
 * mpItemFilters dựng ô tích cho từng loại vật phẩm CÓ THẬT trong gói vừa nhận.
 *
 * Dựng theo dữ liệu chứ không theo một danh sách chép tay: thêm loại vật phẩm mới trong
 * `config/game.json` thì bộ lọc tự có, không phải nhớ sửa thêm chỗ này.
 */
function mpItemFilters(v) {
  const box = document.getElementById('mpitems');
  const kinds = Object.keys(v.items || {}).sort();
  if (box.dataset.kinds === kinds.join(',')) {
    // Chỉ cập nhật số đếm, giữ nguyên DOM — dựng lại mỗi 200 ms thì không bấm vào được.
    for (const k of kinds) {
      const n = box.querySelector(`[data-n="${k}"]`);
      if (n) n.textContent = ' ' + (v.items[k].length / 2);
    }
    return;
  }
  box.dataset.kinds = kinds.join(',');
  box.replaceChildren(el('span', 'flabel', 'Vật phẩm:'), ...kinds.map((k) => {
    const l = document.createElement('label');
    const i = document.createElement('input');
    i.type = 'checkbox';
    i.checked = !hiddenItems.has(k);
    i.addEventListener('change', () => {
      if (i.checked) hiddenItems.delete(k); else hiddenItems.add(k);
      mpDraw();
    });
    const sw = el('b', 'sw');
    sw.style.background = itemColor(k);
    const n = el('small', '', ' ' + v.items[k].length / 2);
    n.dataset.n = k;
    l.append(i, sw, document.createTextNode(' ' + itemName(k)), n);
    return l;
  }));
}

async function mpLoadCheck() {
  const box = document.getElementById('mpcheck');
  let rep;
  try {
    rep = await api('/api/admin/check');
  } catch (_) {
    box.replaceChildren(el('p', '', 'Chưa có báo cáo --check. Chạy: cd server && go run ./cmd/mapgen --check'));
    return;
  }
  mp.check = rep;
  mpShowCheck();
}

function mpShowCheck() {
  const box = document.getElementById('mpcheck');
  const id = mp.map?.id;
  const rows = mp.check?.maps?.[id];
  if (!rows) { box.replaceChildren(); return; }
  const r = rows[mp.realm];
  if (!r) { box.replaceChildren(); return; }
  const want = mp.check.targets?.[id];
  const pct = (v) => (v * 100).toFixed(1) + '%';
  const line = el('div', 'checkline');
  const add = (k, v, bad) => {
    const s = el('span', bad ? 'bad' : '');
    s.append(el('i', '', k), document.createTextNode(' ' + v));
    line.append(s);
  };
  add('Cảnh giới', `${r.Name} (r=${r.Realm}), thân ${Math.round(r.BodyR)}`);
  add('Mật độ hình vẽ', pct(r.Density) + (want ? ` (mục tiêu ${pct(want[0])}–${pct(want[1])})` : ''),
    want && (r.Density > want[1] || r.Density < want[0]));
  add('Khoảng thở', pct(r.Breathing), r.Breathing < 0.2);
  add('Đi được', pct(r.OpenFrac), r.OpenFrac < 0.6);
  add('Túi kín to nhất', r.BiggestTrap + ' ô', r.BiggestTrap >= 4);
  if (r.CutOff?.length) add('KHÔNG TỚI ĐƯỢC', r.CutOff.join(', '), true);
  box.replaceChildren(line);
}

// ── Điều khiển ───────────────────────────────────────────────────────────────────────────

function mpSetMode(mode) {
  mp.mode = mode;
  const live = mode === 'live';
  document.getElementById('mprlab').hidden = live;
  document.getElementById('mpcheck').hidden = live;
  if (live) { mp.live = null; mpTickLive(); } else { mp.live = null; mpLoadCheck(); mpDraw(); }
}

/** mpZones đổ danh sách khu từ dashboard (đã có sẵn, không gọi thêm API). */
function mpZones(stats) {
  const sel = document.getElementById('mpzone');
  const want = [];
  for (const r of stats.rooms || []) {
    for (const z of r.zones || []) want.push({ v: `${r.id}|${z.index}`, label: `${r.id} · ${z.name}` });
  }
  const cur = sel.value;
  if (sel.options.length === want.length && cur) return;
  sel.replaceChildren(...want.map((w) => {
    const o = document.createElement('option');
    o.value = w.v; o.textContent = w.label;
    return o;
  }));
  if (want.length && !cur) {
    sel.value = want[0].v;
    [mp.room, mp.zone] = [want[0].v.split('|')[0], Number(want[0].v.split('|')[1])];
  }
}

function mpInit() {
  const c = cv();
  document.getElementById('mpmode').addEventListener('change', (e) => mpSetMode(e.target.value));
  document.getElementById('mpzone').addEventListener('change', (e) => {
    const [room, zone] = e.target.value.split('|');
    mp.room = room; mp.zone = Number(zone); mp.map = null; mp.follow = 0;
    if (mp.mode === 'live') mpTickLive(); else mpLoadCheck();
  });
  const slider = document.getElementById('mpr');
  slider.addEventListener('input', () => {
    mp.realm = Number(slider.value);
    document.getElementById('mprv').textContent = slider.value;
    mpDraw(); mpShowCheck();
  });

  // Lớp bật/tắt (§5.1).
  const box = document.getElementById('mplayers');
  for (const [key, label] of [['tier', 'bậc vật cản'], ['zones', 'vùng'],
    ['foot', 'chân đế'], ['reach', 'tầm kiếm'], ['grid', 'lưới 1 000']]) {
    const l = document.createElement('label');
    const i = document.createElement('input');
    i.type = 'checkbox'; i.checked = mp.layers[key];
    i.addEventListener('change', () => { mp.layers[key] = i.checked; mpDraw(); });
    l.append(i, document.createTextNode(' ' + label));
    box.append(l);
  }

  // Kéo và phóng to.
  let drag = null;
  c.addEventListener('pointerdown', (e) => {
    drag = { x: e.offsetX, y: e.offsetY, ox: mp.ox, oy: mp.oy, moved: false };
    c.setPointerCapture(e.pointerId);
  });
  c.addEventListener('pointermove', (e) => {
    if (!drag) return;
    const dx = e.offsetX - drag.x, dy = e.offsetY - drag.y;
    if (Math.abs(dx) + Math.abs(dy) > 3) drag.moved = true;
    mp.ox = drag.ox - dx / mp.k;
    mp.oy = drag.oy - dy / mp.k;
    mp.follow = 0; // kéo tay là thôi bám theo — nếu không thì khung giật về ngay
    mpDraw();
  });
  c.addEventListener('pointerup', (e) => {
    const wasDrag = drag?.moved;
    drag = null;
    if (wasDrag || mp.mode !== 'live' || !mp.live) return;
    // Bấm vào một người → bám theo (§5.1).
    const wx = mp.ox + e.offsetX / mp.k, wy = mp.oy + e.offsetY / mp.k;
    let best = null, bestD = 1e9;
    for (const p of mp.live.players) {
      const d = Math.hypot(p.x - wx, p.y - wy);
      if (d < bestD) { bestD = d; best = p; }
    }
    if (best && bestD < 200 / mp.k * 2) {
      mp.follow = mp.follow === best.id ? 0 : best.id;
      mpDraw();
    }
  });
  c.addEventListener('wheel', (e) => {
    e.preventDefault();
    const wx = mp.ox + e.offsetX / mp.k, wy = mp.oy + e.offsetY / mp.k;
    mp.k *= e.deltaY < 0 ? 1.15 : 1 / 1.15;
    mp.k = Math.min(Math.max(mp.k, 0.01), 2);
    mp.ox = wx - e.offsetX / mp.k;
    mp.oy = wy - e.offsetY / mp.k;
    mp.follow = 0;
    mpDraw();
  }, { passive: false });

  // 5 lần/giây — §5.1. Hàm tự thoát khi trang không mở, nên để chạy luôn là an toàn.
  mp.timer = setInterval(mpTickLive, 200);
}
