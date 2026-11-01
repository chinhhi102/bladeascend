// Bảng theo nhân vật — GAME_SPEC v6.30 §6.3 (Q36). JS thuần cho tới khi admin chuyển React + antd.

let charsShownAt = 0;

async function loadChars() {
  const days = document.getElementById('chdays').value;
  const s = await api(`/api/admin/characters?days=${encodeURIComponent(days)}`);
  const body = document.querySelector('#chtbl tbody');
  const rows = (s.rows || []).map((r) => {
    const tr = el('tr');
    tr.append(el('td', '', r.character), el('td', '', r.lives), el('td', '', `${f1(r.pick_share)}%`),
      el('td', '', f1(r.score_per_min)), el('td', '', `${Math.round(r.avg_survive)} s`),
      el('td', '', f1(r.avg_peak_realm)), el('td', '', f1(r.casts_per_min)));
    return tr;
  });
  if (rows.length === 0) {
    const tr = el('tr');
    const td = el('td', 'hint', 'Chưa có mạng nào đủ 30 giây trong khoảng này.');
    td.colSpan = 7;
    tr.append(td);
    rows.push(tr);
  }
  body.replaceChildren(...rows);
}

document.getElementById('chdays').addEventListener('change', () => loadChars().catch((e) => console.warn('[admin] chars', e)));
// Nạp khi mở tab, làm mới mỗi phút (số liệu theo ngày, không cần nhanh hơn).
setInterval(() => {
  if (document.getElementById('chars').hidden || Date.now() - charsShownAt < 60000) return;
  charsShownAt = Date.now();
  loadChars().catch((e) => console.warn('[admin] chars', e));
}, 500);
