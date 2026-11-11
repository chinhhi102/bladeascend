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
import { launch } from './browser.mjs';
import { requireDebug } from './needdebug.mjs';
import { cpus, loadavg } from 'node:os';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
const WATCH_MS = Number(process.env.WATCH_MS ?? 20000);
// Ngưỡng ĐẠT nay là TUYỆT ĐỐI: 0 nhảy cóc, 0 đứng hình — tính trên VẾT trong vòng vẽ
// (xem phần phán quyết ở cuối). Ngưỡng tạm "≤ 3% khung" của Q48 đã bỏ: nó đo theo nhịp của
// chính bài đo, và nhịp ấy lệch với nhịp vẽ nên con số dao động 0–3,5% giữa các lượt trên
// CÙNG một bản (đo 11/11). Hạ về 0 là bước cuối của Q48, chốt 11/11 sau khi đồng hồ chỉnh
// tốc + trễ nội suy tự điều chỉnh vào chạy thật.
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

const b = await launch({
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
  // Vết (thời_điểm_vẽ, x, y) ghi TRONG vòng vẽ của game — xem `ClientWorld.traceEntity`.
  window.kx.world.traceEntity?.(id);
  const fast = [];            // tốc độ vẽ (đv/s) của những khung có di chuyển
  let frames = 0, missing = 0, gone = 0, respawn2s = 0, goneAt = -1;
  let nhayCoc = 0, dungHinh = 0, nhanhNhat = 0;
  let toNhat = {};
  let nhayCocTheoVe = 0, nhanhNhatTheoVe = 0, prevRt = 0;
  let nhayCocNoiSuy = 0, nhanhNhatNoiSuy = 0, prevRaw = null, capNoiSuy = {};
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
    const off = w.clockOffsetSecs ? w.clockOffsetSecs() : (w.offset ?? 0);
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
    // Tốc độ đo theo THỜI ĐIỂM VẼ, không theo giờ máy.
    //
    // Hai con số này chỉ khác nhau khi vòng lấy mẫu (rAF của bài đo) lệch nhịp với vòng vẽ
    // của game: lúc ấy một khung lấy mẫu "thấy" hai khung vẽ gộp lại, và tốc độ tính theo
    // giờ máy vọt lên gấp đôi dù hình vẽ ra hoàn toàn đều. Đây chính là kiểu lỗi PHÉP ĐO
    // từng làm lượt đo 10/11 đỏ oan.
    // Vị trí NỘI SUY THUẦN (đầu ra của `followSnaps`, trước khi lớp vẽ đụng vào) — so với
    // vị trí SPRITE để biết cú nhảy sinh ra ở netcode hay ở lớp vẽ.
    const raw = v?.track?.shown;
    if (raw && prevRaw && prevRt > 0) {
      const rtPrev = prevRt;
      const dRt = w.renderTime() - rtPrev;
      if (dRt > 0) {
        const vRaw = Math.hypot(raw.x - prevRaw.x, raw.y - prevRaw.y) / dRt;
        nhanhNhatNoiSuy = Math.max(nhanhNhatNoiSuy, vRaw);
        if (vRaw > cap * jumpMult) {
        nhayCocNoiSuy++;
        // Cặp gói ĐANG ĐƯỢC NỘI SUY ở đúng khung này (không phải hai gói mới nhất): nếu
        // chính cặp ấy đã đi 1 700 đv/s thì dữ liệu server vốn có cú dịch nhanh (đòn đẩy
        // lùi của kỹ năng đi 6 U trong 0,3 s = gấp đôi tốc chạy), và "nhảy cóc" là THẬT
        // nhưng đúng luật — không phải lỗi nội suy.
        if (vRaw > (capNoiSuy.v ?? 0)) {
          const rtv = w.renderTime();
          let ia = -1;
          for (let k = snaps.length - 1; k > 0; k--) {
            if (snaps[k - 1].t <= rtv) { ia = k - 1; break; }
          }
          capNoiSuy = ia >= 0 ? {
            v: +vRaw.toFixed(0),
            goiCachMs: +((snaps[ia + 1].t - snaps[ia].t) * 1000).toFixed(0),
            goiDoiCho: +Math.hypot(snaps[ia + 1].s.x - snaps[ia].s.x, snaps[ia + 1].s.y - snaps[ia].s.y).toFixed(0),
            tocCapGoi: +(Math.hypot(snaps[ia + 1].s.x - snaps[ia].s.x, snaps[ia + 1].s.y - snaps[ia].s.y) / Math.max(snaps[ia + 1].t - snaps[ia].t, 1e-6)).toFixed(0),
          } : { v: +vRaw.toFixed(0), goiCachMs: -1 };
        }
      }
      }
    }
    prevRaw = raw ? { x: raw.x, y: raw.y } : null;
    const rtNow = w.renderTime();
    if (prev && prevRt > 0 && rtNow > prevRt) {
      const vRt = Math.hypot(at.x - prev.x, at.y - prev.y) / (rtNow - prevRt);
      nhanhNhatTheoVe = Math.max(nhanhNhatTheoVe, vRt);
      if (vRt > cap * jumpMult) nhayCocTheoVe++;
    }
    prevRt = rtNow;
    if (prev && dt > 0 && dt < 0.5) {
      const v = Math.hypot(at.x - prev.x, at.y - prev.y) / dt;
      nhanhNhat = Math.max(nhanhNhat, v);
      if (v > cap * jumpMult) {
        nhayCoc++;
        if (du <= 0) nhayCocKhoCan++;
        // Ghi lại BỐI CẢNH của cú nhảy to nhất: hai gói gần nhất và thời điểm vẽ. Không có
        // nó thì chỉ biết "có nhảy" mà không biết nhảy vì kho cạn, vì đồng hồ, hay vì chính
        // dữ liệu server gửi đã nhảy.
        if (v > (toNhat.v ?? 0)) {
          const n = snaps.length;
          toNhat = {
            v: +v.toFixed(0), du: +du.toFixed(1), interpMs: w.interpNowMs ? +w.interpNowMs().toFixed(0) : 0,
            doiKhung: +Math.hypot(at.x - prev.x, at.y - prev.y).toFixed(0),
            goiCach: n >= 2 ? +((snaps[n - 1].t - snaps[n - 2].t) * 1000).toFixed(0) : 0,
            goiDoiCho: n >= 2 ? +Math.hypot(snaps[n - 1].s.x - snaps[n - 2].s.x, snaps[n - 1].s.y - snaps[n - 2].s.y).toFixed(0) : 0,
            co: n ? (snaps[n - 1].s.fl ?? 0) : 0,
          };
        }
      }
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
    cuNhayToNhat: toNhat,
    ...(() => {
      // Tốc độ tính từ VẾT: cặp (thời điểm vẽ, vị trí) lấy cùng một khung, nên không còn
      // lệch nhịp giữa vòng đo và vòng vẽ.
      const tr = window.kx.world.takeTrace?.() ?? [];
      // Vết là bộ 4: (thời điểm vẽ, x, y, tốc độ của DỮ LIỆU ở khung đó).
      //
      // Có cột cuối mới tách được ba thứ nhìn ra đều giống "giật":
      //   · dữ liệu nhảy (qua cổng, hồi sinh) → hình phải nhảy theo, ĐÚNG;
      //   · dữ liệu đứng (người ta dừng thật) → hình đứng, ĐÚNG;
      //   · dữ liệu đều mà hình nhảy hoặc đứng → LỖI nội suy, chính là thứ Q48 đi chữa.
      let n = 0, fastest = 0, stalls = 0, teleports = 0;
      for (let i = 4; i < tr.length; i += 4) {
        const dRt = tr[i] - tr[i - 4];
        if (dRt <= 0) continue;
        const vv = Math.hypot(tr[i + 1] - tr[i - 3], tr[i + 2] - tr[i - 2]) / dRt;
        const dataV = tr[i + 3];
        if (dataV > cap * 1.3) {
          teleports++; // dịch chuyển/hồi sinh: dữ liệu vốn đã nhảy
          continue;
        }
        fastest = Math.max(fastest, vv);
        if (vv > cap * jumpMult) n++;
        if (vv < 1 && dataV > cap * 0.5) stalls++;
      }
      return {
        nhayCocTheoVet: n, nhanhNhatTheoVet: +fastest.toFixed(0),
        dungHinhTheoVet: stalls, soMauVet: tr.length / 4, dichChuyenThat: teleports,
      };
    })(),
    nhayCocNoiSuyThuan: nhayCocNoiSuy,
    capGoiLucNhay: capNoiSuy,
    nhanhNhatNoiSuyThuan: +nhanhNhatNoiSuy.toFixed(0),
    nhayCocTheoThoiDiemVe: nhayCocTheoVe,
    nhanhNhatTheoThoiDiemVe: +nhanhNhatTheoVe.toFixed(0),
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
// Thấy được bao nhiêu khung? Lượt nào Shipper chết sớm (hoặc chưa kịp vào tầm nhìn) thì
// bài này đo trên vài chục khung rồi báo "nhảy cóc 0%" — một màu XANH GIẢ, nguy hiểm hơn
// hẳn một lần đỏ. Đo được 11/11: hai lượt liền chỉ thấy ~150/1 200 khung mà vẫn ĐẠT.
const thayTyLe = res.frames > 0 ? res.khungThay / res.frames : 0;
if (thayTyLe < 0.5) {
  console.error(`\nKHÔNG ĐO ĐƯỢC: chỉ thấy ${res.khungThay}/${res.frames} khung ` +
    `(${Math.round(thayTyLe * 100)}%) — Shipper chết sớm hoặc chưa vào tầm nhìn. Chạy lại.`);
  await b.close();
  process.exit(2);
}
const perCpu = loadavg()[0] / cpus().length;
console.log(`tải máy: ${loadavg()[0].toFixed(1)} trên ${cpus().length} nhân = ${perCpu.toFixed(2)}/nhân`);
if (perCpu > MAX_LOAD_PER_CPU && process.env.IGNORE_LOAD !== '1') {
  console.error(`\nKHÔNG DÙNG ĐƯỢC: máy đang chạy ${perCpu.toFixed(2)} lần số nhân ` +
    `(ngưỡng ${MAX_LOAD_PER_CPU}). Số đo là nhiễu máy, không phải hành vi netcode.`);
  console.error('Chạy lại lúc máy rảnh, hoặc IGNORE_LOAD=1 nếu cố tình muốn xem số.');
  process.exit(2);
}

// PHÁN QUYẾT đọc từ VẾT, không từ vòng lấy mẫu.
//
// Hai cột `nhayCoc*` cũ đo bằng cách: lấy vị trí sprite ở nhịp `rAF` của bài đo, chia cho
// giờ máy. Vòng đo và vòng vẽ của game không cùng nhịp, nên một khung đo có lúc "thấy" hai
// khung vẽ gộp lại và tốc độ vọt gấp đôi dù hình chạy hoàn toàn đều. Đo 11/11 cho thấy rõ:
// cùng một lượt, đo theo nhịp bài là 2 807 đv/s, đo theo vết (vị trí + thời điểm vẽ lấy
// trong CÙNG một khung vẽ) là 816 đv/s — trong khi cặp gói đang nội suy chỉ đi 939 đv/s,
// tức dữ liệu vốn đã đúng. Giữ hai cột cũ để đọc, nhưng không cho chúng quyết đỏ/xanh nữa.
const ok = res.hienLai2s === 0 && res.nhayCocTheoVet === 0 && res.dungHinhTheoVet === 0;
console.log(ok
  ? `ĐẠT: 0 nhảy cóc, 0 đứng hình trên ${res.soMauVet} khung vẽ — nhanh nhất ${res.nhanhNhatTheoVet} đv/s (trần ${res.tranToc})`
  : `TRƯỢT: ${res.nhayCocTheoVet} nhảy cóc + ${res.dungHinhTheoVet} đứng hình trên ${res.soMauVet} khung vẽ ` +
    `(nhanh nhất ${res.nhanhNhatTheoVet} đv/s, trần ${res.tranToc}), dựng lại ${res.hienLai2s} lần`);
await b.close();
process.exit(ok ? 0 : 1);
