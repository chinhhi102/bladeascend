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
import { cpus, loadavg } from 'node:os';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
const WATCH_MS = Number(process.env.WATCH_MS ?? 20000);
// Ngưỡng ĐẠT theo TỈ LỆ khung nhảy cóc. Q48 chốt tạm ≤ 3% để bắt thoái lui; hạ về 0 khi
// xong bước 2–4. Đặt `MAX_PCT=0` để canh tuyệt đối.
const MAX_PCT = Number(process.env.MAX_PCT ?? 3);
// Máy chạy quá ngần này lần số nhân thì mọi con số bên dưới là nhiễu máy, không phải hành
// vi netcode. Đặt `IGNORE_LOAD=1` để vẫn đọc số.
//
// Đo tải bằng `loadavg` chứ không bằng thời gian khung của tab xem: 10/11 máy chạy load 39
// trên 12 nhân mà khung của tab vẫn ~16,7 ms (headless vẽ nhẹ), trong khi cùng một commit
// cho ra 0,66% rồi 18,8% khung nhảy cóc. Thời gian khung không nhìn thấy cái đang phá.
const MAX_LOAD_PER_CPU = Number(process.env.MAX_LOAD ?? 1.5);
const NO_SKILL = process.env.NO_SKILL === '1'; // mốc so sánh: cùng kịch bản nhưng KHÔNG bật Gom Đơn
const JUMP_MULT = 1.5;
// Mỗi lượt chạy một hướng. Ngắn, vì 975 đv/s giữ 2 giây là đi 1 950 đơn vị — vượt khỏi
// khung nhìn của người đứng xem và 84% số khung không thấy ai (lượt đo hỏng đầu tiên).
const LEG_MS = Number(process.env.LEG_MS ?? 700); // vẽ nhanh hơn trần tốc 1,5 lần = client giật tới trước, không phải nội suy
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

// Shipper: giữ Gom Đơn bật GẦN NHƯ LIÊN TỤC rồi chạy qua lại.
//
// Kỹ năng 4 s, hồi 20 s, vào trận 1 lượt — để tự nhiên thì trong 20 giây đo chỉ có 4–8 giây
// chạy nhanh, và tỉ lệ đó đổi từng lượt chạy. Lượt đo đầu tôi dính đúng cái đó: cùng một
// cấu hình mà trung vị tốc nhảy từ 641 xuống 250 đv/s, số khung hỏng thành ra không so
// được với nhau. Phím debug `G` nạp đầy lượt nên cứ 4 giây bật lại một lần: gần như cả cửa
// sổ đo là chạy nhanh, hai lần chạy mới so được. Mỗi lượt chạy chỉ `LEG_MS` để anh ta
// không rời khung nhìn — xem ghi chú ở `LEG_MS`.
const drive = (async () => {
  const t0 = Date.now();
  let dir = 0;
  while (Date.now() - t0 < WATCH_MS) {
    if (!NO_SKILL) {
      await shipper.keyboard.press('KeyG'); // nạp đầy lượt kỹ năng
      await shipper.keyboard.press('Space');
    }
    const key = ['KeyD', 'KeyA'][dir++ % 2];
    await shipper.keyboard.down(key);
    await sleep(LEG_MS);
    await shipper.keyboard.up(key);
  }
})();

const sample = watcher.evaluate(async ({ ms, id, cap, jumpMult }) => {
  const fast = [];            // tốc độ vẽ (đv/s) của những khung có di chuyển
  let frames = 0, missing = 0, gone = 0, respawn2s = 0, goneAt = -1;
  let nhayCoc = 0, dungHinh = 0, nhanhNhat = 0;
  let duMin = 1e9, duSum = 0, duN = 0, khoCan = 0;
  let lastOff = null, offUp = 0, offDown = 0, offJumpMax = 0;
  const gaps = [];
  const frameGaps = [];
  let seenT = 0;
  let dungHinhKhoCan = 0, nhayCocKhoCan = 0;
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
    // --- Q48 bước 0: ba con số chẩn đoán, lấy cùng khung với mẫu vị trí ---
    const w = window.kx.world;
    const v = w.players.get(id);
    const snaps = v?.snaps ?? [];
    const lastT = snaps.length ? snaps[snaps.length - 1].t : 0;
    // (a) bộ đệm còn dư bao nhiêu ms so với thời điểm vẽ
    const du = (lastT - w.renderTime()) * 1000;
    duMin = Math.min(duMin, du);
    duSum += du; duN++;
    if (du <= 0) khoCan++;
    // (b) độ lệch đồng hồ ước tính (giây) — chỉ cần biết nó NHẢY bao nhiêu
    const off = w.offset ?? 0;
    if (lastOff !== null && off !== lastOff) {
      const jump = (off - lastOff) * 1000;
      offJumpMax = Math.max(offJumpMax, jump);
      if (jump > 0) offUp++; else offDown++;
    }
    lastOff = off;
    // (c) tần số gửi thực tế của entity đang xem — khoảng cách giữa hai dấu thời gian gói
    if (lastT !== seenT) {
      if (seenT > 0) gaps.push((lastT - seenT) * 1000);
      seenT = lastT;
    }

    const at = { x: sp.root.x, y: sp.root.y };
    if (prevT > 0) frameGaps.push(now - prevT);
    const dt = (now - prevT) / 1000;
    if (prev && dt > 0 && dt < 0.5) {
      const v = Math.hypot(at.x - prev.x, at.y - prev.y) / dt;
      nhanhNhat = Math.max(nhanhNhat, v);
      if (v > cap * jumpMult) { nhayCoc++; if (du <= 0) nhayCocKhoCan++; }
      if (v > 1) fast.push(v);
      // Đứng hình: khung gần như không nhúc nhích ngay giữa lúc đang chạy nhanh.
      else if (fast.length && fast[fast.length - 1] > cap * 0.5) { dungHinh++; if (du <= 0) dungHinhKhoCan++; }
    }
    prev = at;
    prevT = now;
  }
  fast.sort((a, c) => a - c);
  const pct = (q) => (fast.length ? +fast[Math.floor((fast.length - 1) * q)].toFixed(0) : 0);
  const thay = frames - missing;
  return {
    frames,
    khungThay: thay,
    khungTrungViMs: (() => { const f = [...frameGaps].sort((a, c) => a - c); return f.length ? +f[Math.floor(f.length / 2)].toFixed(1) : 0; })(),
    tyLeNhayCoc: thay ? +(100 * nhayCoc / thay).toFixed(2) : 0,
    tocTrungVi: pct(0.5),
    tocP99: pct(0.99),
    nhanhNhat: +nhanhNhat.toFixed(0),
    tranToc: +cap.toFixed(0),
    nhayCoc,
    dungHinh,
    // --- chẩn đoán bước 0 ---
    duTrungBinhMs: duN ? +(duSum / duN).toFixed(1) : 0,
    duThapNhatMs: duMin < 1e8 ? +duMin.toFixed(1) : 0,
    khungKhoCan: khoCan,
    dungHinhLucKhoCan: dungHinhKhoCan,
    nhayCocLucKhoCan: nhayCocKhoCan,
    goiTrungViMs: (() => { const g = [...gaps].sort((a, c) => a - c); return g.length ? +g[Math.floor(g.length / 2)].toFixed(0) : 0; })(),
    goiLonNhatMs: gaps.length ? +Math.max(...gaps).toFixed(0) : 0,
    goiTren75ms: gaps.filter((x) => x > 75).length,
    soGoi: gaps.length,
    dongHoNhichLen: offUp,
    dongHoLui: offDown,
    dongHoNhayLonNhatMs: +offJumpMax.toFixed(1),
    bienMat: gone,
    hienLai2s: respawn2s,
    khungKhongThay: missing,
  };
}, { ms: WATCH_MS, id: shipperID, cap: CAP, jumpMult: JUMP_MULT });

await drive;
const res = await sample;
console.log(JSON.stringify({ ...res, giay: WATCH_MS / 1000 }, null, 1));

// ĐẠT khi: không entity nào bị dựng lại trong khung nhìn, và không có nấc nhảy cóc nào.
// `dungHinh` chỉ để đọc — Shipper có lúc đứng yên thật (giữa hai lần đổi hướng).
// Máy quá tải thì KHÔNG kết luận gì.
const perCpu = loadavg()[0] / cpus().length;
console.log(`tải máy: ${loadavg()[0].toFixed(1)} trên ${cpus().length} nhân = ${perCpu.toFixed(2)}/nhân`);
if (perCpu > MAX_LOAD_PER_CPU && process.env.IGNORE_LOAD !== '1') {
  console.error(`\nKHÔNG DÙNG ĐƯỢC: máy đang chạy ${perCpu.toFixed(2)} lần số nhân ` +
    `(ngưỡng ${MAX_LOAD_PER_CPU}). Số đo là nhiễu máy, không phải hành vi netcode.`);
  console.error('Chạy lại lúc máy rảnh, hoặc IGNORE_LOAD=1 nếu cố tình muốn xem số.');
  process.exit(2);
}

const ok = res.hienLai2s === 0 && res.tyLeNhayCoc <= MAX_PCT;
console.log(ok
  ? `ĐẠT (ngưỡng ≤ ${MAX_PCT}%): nhảy cóc ${res.tyLeNhayCoc}% — nhanh nhất ${res.nhanhNhat} đv/s (trần ${res.tranToc}), 0 dựng lại`
  : `TRƯỢT (ngưỡng ≤ ${MAX_PCT}%): nhảy cóc ${res.nhayCoc}/${res.khungThay} khung thấy được (${res.tyLeNhayCoc}%, nhanh nhất ${res.nhanhNhat} đv/s, trần ${res.tranToc}), dựng lại ${res.hienLai2s} lần`);
process.exit(ok ? 0 : 1);
await b.close();
