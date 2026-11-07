// Nhìn Shipper đang Gom Đơn TỪ MỘT TAB KHÁC — đo bằng số, không bằng mắt.
//
//   cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true
//   node tools/browsertest/gomdon-smooth.mjs http://127.0.0.1:9998
//
// Vì sao cần: v6.45 nâng Gom Đơn lên ×5 (975 đv/s ở `move_base` 195). Đó là tốc nhanh nhất
// game, gấp 2,5 lần mức netcode từng chạy hằng ngày. Hai thứ hỏng trước tiên ở tốc đó đều
// KHÔNG thấy được bằng mắt ở phía người chạy — chỉ người ĐỨNG XEM mới thấy:
//
//   1. **Nhảy cóc**: gói snapshot tới thưa hơn khung hình, client nội suy không kịp thì
//      nhân vật dịch từng nấc to thay vì trôi đều.
//   2. **Biến mất rồi hiện lại**: chạy 975 đv/s thì ra khỏi vùng AOI của người xem trong
//      chưa tới một giây; vào lại mà client dựng entity mới là nhấp nháy.
//
// Cách đo: tab A đứng yên xem, tab B là Shipper chạy qua lại và bật Gom Đơn liên tục. Tab A
// lấy mẫu vị trí VẼ RA của Shipper mỗi khung hình (`window.kx.scene.players`) KÈM thời gian
// khung, rồi quy ra **tốc độ quan sát** `đv/s` — không so quãng đường thô.
//
// Vì sao phải chuẩn hoá theo thời gian khung: lượt đo đầu tiên tôi so bước đi với trung vị
// chung và ra "260/1200 khung nhảy cóc". Sai cả hai đường. Một, Gom Đơn chỉ bật 4 giây
// trong 20 giây nên trung vị chung là tốc ĐI BỘ — mọi khung lúc chạy nhanh đều bị chấm là
// nhảy cóc. Hai, một khung hình kẹt 100 ms thì đi 97 đơn vị là ĐÚNG, không phải nhảy.
//
// Ngưỡng theo LUẬT, không theo mẫu: trần tốc là `move_base × speed.cap_gom_don`
// (195 × 5 = 975 đv/s, §4.1). Client vẽ nhanh hơn trần đó quá `JUMP_MULT` lần nghĩa là nó
// GIẬT nhân vật tới trước, không phải nội suy — đó mới là nhảy cóc thật.
//
//   - `nhayCoc`: số khung tốc độ vẽ > JUMP_MULT × trần
//   - `nhanhNhat`: tốc độ vẽ lớn nhất quan sát được (đv/s) — đọc kèm để biết còn cách trần bao xa
//   - `dungHinh`: số khung đứng im trong lúc ĐANG chạy nhanh (gói chưa tới, nội suy hết đà)
//   - `bienMat` / `hienLai2s`: entity bị huỷ rồi dựng lại trong khung nhìn (ngân sách §1: 0)
import { chromium } from 'playwright';
import { requireDebug } from './needdebug.mjs';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
const WATCH_MS = Number(process.env.WATCH_MS ?? 20000);
const NO_SKILL = process.env.NO_SKILL === '1'; // mốc so sánh: cùng kịch bản nhưng KHÔNG bật Gom Đơn
const JUMP_MULT = 1.5; // vẽ nhanh hơn trần tốc 1,5 lần = client giật tới trước, không phải nội suy
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

const cfg = await (await fetch(`${TARGET}/api/config`)).json();
const CAP = cfg.player.move_base * cfg.speed.cap_gom_don; // trần tốc tuyệt đối, §4.1
const shipperIdx = cfg.characters.indexOf('shipper');
if (shipperIdx < 0) {
  throw new Error('config không có nhân vật shipper');
}

const b = await chromium.launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'],
});

async function join(name, charIdx) {
  const ctx = await b.newContext({ viewport: { width: 1280, height: 720 }, locale: 'vi-VN' });
  const p = await ctx.newPage();
  p.on('pageerror', (e) => console.log(`  [${name}] pageerror`, String(e).slice(0, 160)));
  await p.goto(`${TARGET}/?debug=true`, { waitUntil: 'domcontentloaded' });
  await p.waitForSelector('input', { timeout: 30000 });
  await p.fill('input', name);
  await p.locator('.char-card').nth(charIdx).click();
  await p.locator('.home .btn.primary').click();
  await p.waitForSelector('.room-board', { timeout: 60000 });
  return p;
}

const watcher = await join('KiemTraXem', 0);
await requireDebug(watcher, 'gomdon-smooth.mjs');
const shipper = await join('KiemTraShipper', shipperIdx);
await sleep(6000); // chờ atlas nhân vật về, nếu không mẫu đầu là placeholder

// Kéo Shipper về cạnh người xem (phím debug `U`), rồi lấy id của anh ta ở phía người xem.
await watcher.keyboard.press('KeyU');
await sleep(1500);
const shipperID = await shipper.evaluate(() => window.kx.world.myId);
if (!shipperID) {
  throw new Error('không lấy được id của Shipper');
}

// Shipper: bật Gom Đơn rồi chạy qua lại. Kỹ năng 4 s, hồi 20 s, vào trận có 1 lượt — nên
// cứ bấm đều, lượt nào sẵn thì nổ, lượt nào chưa thì vẫn chạy để có mẫu nền so sánh.
const drive = (async () => {
  const t0 = Date.now();
  while (Date.now() - t0 < WATCH_MS) {
    if (!NO_SKILL) await shipper.keyboard.press('Space');
    for (const key of ['KeyD', 'KeyA']) {
      await shipper.keyboard.down(key);
      await sleep(2000);
      await shipper.keyboard.up(key);
      if (Date.now() - t0 >= WATCH_MS) break;
    }
  }
})();

const sample = watcher.evaluate(async ({ ms, id, cap, jumpMult }) => {
  const fast = [];            // tốc độ vẽ (đv/s) của những khung có di chuyển
  let frames = 0, missing = 0, gone = 0, respawn2s = 0, goneAt = -1;
  let nhayCoc = 0, dungHinh = 0, nhanhNhat = 0;
  let prev = null, prevT = 0;
  const t0 = performance.now();
  const tick = () => new Promise((r) => requestAnimationFrame(r));
  while (performance.now() - t0 < ms) {
    await tick();
    frames++;
    const now = performance.now();
    const sp = window.kx.scene.players.get(id);
    if (!sp) {
      missing++;
      if (goneAt < 0) { goneAt = now; gone++; }
      prev = null;
      continue;
    }
    if (goneAt >= 0) {
      if (now - goneAt < 2000) respawn2s++;
      goneAt = -1;
    }
    const at = { x: sp.root.x, y: sp.root.y };
    const dt = (now - prevT) / 1000;
    if (prev && dt > 0 && dt < 0.5) {
      const v = Math.hypot(at.x - prev.x, at.y - prev.y) / dt;
      nhanhNhat = Math.max(nhanhNhat, v);
      if (v > cap * jumpMult) nhayCoc++;
      if (v > 1) fast.push(v);
      // Đứng hình: khung gần như không nhúc nhích ngay giữa lúc đang chạy nhanh.
      else if (fast.length && fast[fast.length - 1] > cap * 0.5) dungHinh++;
    }
    prev = at;
    prevT = now;
  }
  fast.sort((a, c) => a - c);
  const pct = (q) => (fast.length ? +fast[Math.floor((fast.length - 1) * q)].toFixed(0) : 0);
  return {
    frames,
    tocTrungVi: pct(0.5),
    tocP99: pct(0.99),
    nhanhNhat: +nhanhNhat.toFixed(0),
    tranToc: +cap.toFixed(0),
    nhayCoc,
    dungHinh,
    bienMat: gone,
    hienLai2s: respawn2s,
    khungKhongThay: missing,
  };
}, { ms: WATCH_MS, id: shipperID, cap: CAP, jumpMult: JUMP_MULT });

await drive;
const res = await sample;
console.log(JSON.stringify({ ...res, giay: WATCH_MS / 1000 }, null, 1));
await b.close();

// ĐẠT khi: không entity nào bị dựng lại trong khung nhìn, và không có nấc nhảy cóc nào.
// `dungHinh` chỉ để đọc — Shipper có lúc đứng yên thật (giữa hai lần đổi hướng).
const ok = res.hienLai2s === 0 && res.nhayCoc === 0;
console.log(ok
  ? `ĐẠT: nhìn từ tab khác, Shipper Gom Đơn trôi đều — nhanh nhất ${res.nhanhNhat} đv/s (trần ${res.tranToc}), 0 nhảy cóc, 0 dựng lại`
  : `TRƯỢT: nhảy cóc ${res.nhayCoc} khung (nhanh nhất ${res.nhanhNhat} đv/s, trần ${res.tranToc}), dựng lại ${res.hienLai2s} lần`);
process.exit(ok ? 0 : 1);
