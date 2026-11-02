// Đo "hiệu chỉnh/giây" khi ĐI SÁT TƯỜNG (GAME_SPEC §8.2a) — con số chủ dự án hỏi.
//
//   node wallhug.mjs [url]
//
// Vì sao cần script riêng: con số trong PROGRESS (20,2 → 4,8) đo ở **rìa bản đồ**, không
// phải ở tường. Hai chỗ lệch vì hai phép khác nhau (rìa là phép kẹp `play_margin`, tường là
// phép trượt dọc `resolveBlocking`), nên không suy ra nhau được.
//
// Lượt đo đầu tiên của script này ra 0 vì nó chỉ "đi thẳng 7 giây rồi mong đụng tường" —
// bản đồ 12 000 đơn vị thì 7 giây đi được 1 365, không đụng gì cả. Nên bây giờ nó **đọc
// bản đồ**: chặn `fetch` để lấy JSON `/api/map/…` client vừa tải, chọn một dãy TƯỜNG KHÔNG
// PHÁ ĐƯỢC (`hp: 0`) dài nhất ở gần, đi tới đó, rồi ép vào tường và trượt dọc.
import { chromium } from 'playwright';
import { requireDebug } from './needdebug.mjs';

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const URL = process.argv[2] ?? 'http://127.0.0.1:9999';

const SAMPLE_MS = 250;
const SLIDE_MS = 4000;
const WALK_TIMEOUT_MS = 45000;
const WALLS_TO_TRY = 3;

const b = await chromium.launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'],
});
const p = await (await b.newContext({ viewport: { width: 1280, height: 720 } })).newPage();

// Bắt lấy bản đồ client tải về. Không sửa code game chỉ để đo được — chặn `fetch` ở đây là đủ.
await p.addInitScript(() => {
  const orig = window.fetch;
  window.__map = null;
  window.fetch = async (...a) => {
    const res = await orig(...a);
    const url = typeof a[0] === 'string' ? a[0] : (a[0]?.url ?? '');
    if (url.includes('/api/map/')) {
      res
        .clone()
        .json()
        .then((m) => {
          window.__map = m;
        })
        .catch(() => {});
    }
    return res;
  };
});

await p.goto(`${URL}/?debug=true`, { waitUntil: 'domcontentloaded' });
await p.waitForSelector('input', { timeout: 60000 });
await p.fill('input', 'WallHug');
await p.locator('.home .btn.primary').click();
await p.waitForFunction(() => window.__playing?.() === true && window.__map, null, { timeout: 60000 });
await requireDebug(p, 'wallhug.mjs');
await sleep(2500);

const probe = () =>
  p.evaluate(() => {
    const me = window.kx.world.me();
    const s = window.__stats();
    return { x: me?.x ?? 0, y: me?.y ?? 0, fix: s.fixPerSec, worst: s.fixMax };
  });

/**
 * walls trả về các dãy tường KHÔNG PHÁ ĐƯỢC dài nhất, kèm điểm đứng và trục trượt.
 * Dùng `f` (chân đế) chứ không phải `s` (hình vẽ) vì chân đế mới là thứ chặn thân (§8.2a).
 */
const picks = await p.evaluate(() => {
  const me = window.kx.world.me();
  const size = window.__map.size;
  // Mọi vật cản chặn thân đều tính, phá được hay không: người chơi đi sát hàng rào cũng
  // đúng là "đi sát tường". Chỉ cần chân đế đủ dài để trượt được vài giây.
  const obs = window.__map.obstacles.filter((o) => o.f);
  const out = [];
  for (const o of obs) {
    const { x, y, w, h } = o.f;
    const long = Math.max(w, h);
    if (long < 300) {
      continue;
    }
    // Bỏ tường biên: ở đó phép kẹp `play_margin` cũng chạy, nên đo ra là trộn hai nguyên
    // nhân. Muốn biết riêng phép trượt dọc thì phải là tường GIỮA bản đồ.
    if (x < 600 || y < 600 || x + w > size - 600 || y + h > size - 600) {
      continue;
    }
    const horiz = w >= h;
    const cx = x + w / 2;
    const cy = y + h / 2;
    // Đứng ở mặt gần nhất của tường, cách 60 đơn vị.
    const stand = horiz
      ? { x: cx, y: me.y < cy ? y - 60 : y + h + 60 }
      : { x: me.x < cx ? x - 60 : x + w + 60, y: cy };
    // Ép vào tường theo trục ngắn, trượt theo trục dài.
    const into = horiz ? (me.y < cy ? 'KeyS' : 'KeyW') : me.x < cx ? 'KeyD' : 'KeyA';
    const along = horiz ? ['KeyD', 'KeyA'] : ['KeyS', 'KeyW'];
    out.push({ stand, into, along, long, dist: Math.hypot(stand.x - me.x, stand.y - me.y), kind: o.k });
  }
  // Gần trước, dài sau: tường dài cho đoạn trượt dài, tường gần cho đỡ mất thời gian đi.
  out.sort((a, b) => a.dist / Math.max(a.long, 1) - b.dist / Math.max(b.long, 1));
  return out.slice(0, 3);
});

const mapName = await p.evaluate(() => window.__map.name);

if (!picks.length) {
  console.error(`Khu "${mapName}" không có dãy vật cản nào dài ≥ 300 đơn vị ở giữa bản đồ.`);
  await b.close();
  process.exit(2);
}

const held = new Set();
async function hold(keys) {
  for (const k of [...held]) {
    if (!keys.includes(k)) {
      await p.keyboard.up(k);
      held.delete(k);
    }
  }
  for (const k of keys) {
    if (!held.has(k)) {
      await p.keyboard.down(k);
      held.add(k);
    }
  }
}

/** goTo đi tới điểm `t`, bấm phím theo hướng còn lại mỗi nhịp lấy mẫu. */
async function goTo(t) {
  const t0 = Date.now();
  while (Date.now() - t0 < WALK_TIMEOUT_MS) {
    const c = await probe();
    const dx = t.x - c.x;
    const dy = t.y - c.y;
    if (Math.hypot(dx, dy) < 90) {
      await hold([]);
      return true;
    }
    const keys = [];
    if (Math.abs(dx) > 50) {
      keys.push(dx > 0 ? 'KeyD' : 'KeyA');
    }
    if (Math.abs(dy) > 50) {
      keys.push(dy > 0 ? 'KeyS' : 'KeyW');
    }
    await hold(keys);
    await sleep(SAMPLE_MS);
  }
  await hold([]);
  return false;
}

async function sample(ms, sink) {
  const t0 = Date.now();
  while (Date.now() - t0 < ms) {
    await sleep(SAMPLE_MS);
    sink.push((await probe()).fix);
  }
}

const wall = [];
const free = [];

// ——— Mốc so sánh: đi giữa đồng trống ———
await hold(['KeyD']);
await sample(4000, free);
await hold([]);
await sleep(400);

let used = 0;
for (const pick of picks.slice(0, WALLS_TO_TRY)) {
  if (!(await goTo(pick.stand))) {
    continue;
  }
  await sleep(500);
  // Ép vào tường trước, cho chắc là đang CHẠM.
  await hold([pick.into]);
  await sleep(700);
  used++;
  for (const side of pick.along) {
    await hold([pick.into, side]);
    await sample(SLIDE_MS, wall);
  }
  await hold([]);
  await sleep(400);
}

const avg = (a) => (a.length ? a.reduce((s, v) => s + v, 0) / a.length : 0);
const max = (a) => (a.length ? Math.max(...a) : 0);
const fin = await probe();

console.log(
  JSON.stringify(
    {
      khu: mapName,
      tuongThu: used,
      tuong: picks.slice(0, WALLS_TO_TRY).map((x) => `${x.kind} dài ${Math.round(x.long)}`),
      satTuong: { mau: wall.length, tb: +avg(wall).toFixed(2), dinh: +max(wall).toFixed(2) },
      dongTrong: { mau: free.length, tb: +avg(free).toFixed(2), dinh: +max(free).toFixed(2) },
      lechNhatDonVi: +fin.worst.toFixed(2),
    },
    null,
    1,
  ),
);
await b.close();
