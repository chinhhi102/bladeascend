// Kịch bản đo netcode CÓ KIỂM SOÁT — nút thắt B3b trong PERF_LOG.
//
//   node tools/browsertest/netcode-bench.mjs [số lượt]
//
// ── Vì sao cần ──────────────────────────────────────────────────────────────────────────
// `flicker.mjs` đo trên server thật, nơi 60 bot tự do đi lại. Số entity đi ngang khung nhìn
// mỗi lượt dao động 1–22, và chỉ số respawn theo đó nhảy 0–9 TRÊN CÙNG MỘT BẢN BUILD. Ngày
// 30/10 tôi đã lấy "2/13 lượt" so với "8/15 lượt" rồi kết luận một thay đổi làm tốt lên —
// kết luận đó sai, vì hai con số ấy nằm trong cùng một khoảng dao động.
//
// Bài này bỏ hết yếu tố ngẫu nhiên:
//   - Server RIÊNG, cổng riêng, DB tạm, **đúng MỘT bot** thay vì 60 (`bot_fill_target = 1`).
//   - Một khu duy nhất: bỏ luôn yếu tố chuyển khu.
//   - Một tab quan sát, đi theo đường CỐ ĐỊNH và nhịp CỐ ĐỊNH để liên tục kéo con bot kia
//     qua lại ranh giới AOI.
//
// Bản đầu tôi dùng hai tab người chơi và tắt sạch bot. Nó chạy ra toàn số 0 đẹp đẽ — nhưng
// `thấy=1` lộ ra rằng hai tab KHÔNG BAO GIỜ nhìn thấy nhau (chúng sinh ra ở hai chỗ ngẫu
// nhiên cách nhau cả nghìn đơn vị). Bài đo không chạm tới thứ nó định đo, mà vẫn báo "đạt".
//
// Nhờ vậy hai lần chạy khác nhau chỉ còn khác vì thứ ta vừa sửa, không phải vì hôm nay bot
// đi đâu. Muốn so hai cấu hình thì chạy bài này cho từng cấu hình, cùng số lượt.
import { chromium } from 'playwright';
import { spawn } from 'node:child_process';
import { mkdtempSync, readFileSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const REPO = join(dirname(fileURLToPath(import.meta.url)), '..', '..');
const PORT = 9996; // KHÔNG phải 9999: bài đo không bao giờ được đụng tới server người chơi
const ROUNDS = Number(process.argv[2] ?? 5);
/**
 * Mỗi lượt đo bao lâu.
 *
 * Phải chứa trọn MỘT VÒNG đi-đứng-về, nếu không bài đo chỉ ghi được nửa đầu. Trần mỗi chiều
 * là `LEG_LIMIT_MS` = 40 s, cộng `DWELL_MS` = 3 s, nên một vòng xấu nhất mất 83 s. Để 60 s
 * thì người đi ra tới nơi, đứng chờ, rồi HẾT GIỜ ngay trước lúc quay lại: bài đo in
 * "rời=1 · quay lại=0" và cảnh báo "chưa chạm tới thứ nó định đo" — đúng chuyện đã xảy ra
 * khi bản đồ v6.12 dày lên và mỗi chiều đi lâu hơn hẳn.
 */
const WATCH_MS = 150000;
/**
 * Người đi: mỗi chiều bao lâu rồi quay đầu, và đứng ngoài tầm bao lâu.
 *
 * **Phải đủ XA và đủ LÂU**, nếu không bài đo không chạm tới thứ nó định đo. Bản trước đi
 * 4,5 giây mỗi chiều ≈ 880 đơn vị, trong khi bán kính RA của AOI là nửa khung nhìn (~800)
 * cộng lề 900 ≈ **1 700**. Người đi chưa bao giờ ra khỏi tầm, nên không có `Despawn` nào,
 * không có lần quay lại nào — và bài đo in ra toàn số 0 đẹp đẽ mà chẳng chứng minh gì.
 * Chính cảnh báo "không entity nào QUAY LẠI" ở cuối bài là thứ lộ ra điều đó.
 *
 * 11 giây ≈ 2 145 đơn vị ở `move_base` 195 → vượt hẳn bán kính ra. Đứng ngoài 3 giây để
 * qua hạn trễ `aoiLingerSecs` = 2 giây, rồi mới quay lại.
 */
const DWELL_MS = 3000;
/** OUT_DIST: đi xa hơn ngần này thì chắc chắn vượt bán kính ra (830 + 900 = 1 730). */
const OUT_DIST = 2400;
/** BACK_DIST: về gần mốc hơn ngần này thì chắc chắn mọi entity cũ đã vào lại khung nhìn. */
const BACK_DIST = 700;
/** LEG_LIMIT_MS: trần thời gian mỗi chiều, phòng khi vướng vật cản không đi nổi. */
const LEG_LIMIT_MS = 40000;

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

// ── 1. Config riêng: tắt sạch bot ────────────────────────────────────────────────────────
const dir = mkdtempSync(join(tmpdir(), 'kx-bench-'));
const cfg = JSON.parse(readFileSync(join(REPO, 'config', 'game.json'), 'utf8'));
// ĐÚNG MỘT bot. Zero thì không có gì để đo; 60 thì phương sai nuốt mất tín hiệu.
// KHÔNG bot. Thứ bài đo theo dõi là tab `walker` — nó đứng YÊN một chỗ suốt lượt, nên khi
// người quan sát đi xa rồi quay về thì nó chắc chắn rời khung rồi vào lại khung.
//
// Bản trước để 1 bot và đếm được "rời 1 lần, quay lại 0 lần": bot đi lang thang, và tệ hơn,
// nó **chém chết tab đứng yên** — người chết hồi sinh với id mới, nên với người quan sát đó
// là một entity khác, không phải "quay lại". Bài đo vì thế báo `respawn=0` mà kèm cảnh báo
// "chưa chạm tới thứ nó định đo": con số 0 đẹp đẽ nhưng không chứng minh gì.
cfg.room.bot_fill_target = 0;
// KHÔNG quái. Quái giết chết mấy con bot mà bài đo đang theo dõi, và người chết **hồi sinh
// với id mới** — với người quan sát đó là một entity khác, nên lần quay lại không bao giờ
// được đếm. Triệu chứng: "rời khung 2 lần · quay lại 0 lần" kèm cảnh báo cuối bài.
//
// Bài này đo VÒNG ĐỜI ENTITY theo AOI (NETCODE_SPEC §4.3), không đo chiến đấu; quái chỉ
// thêm nhiễu.
cfg.room.monster_count = 0;
cfg.room.zone_min_chars = 0;
cfg.room.zones = 1; // một khu: bỏ luôn yếu tố chuyển khu
const cfgPath = join(dir, 'game.json');
writeFileSync(cfgPath, JSON.stringify(cfg));

const bin = join(REPO, 'server', 'kxserver');
const srv = spawn(bin, [
  '-addr', `127.0.0.1:${PORT}`,
  '-config', cfgPath,
  '-db', join(dir, 'lb.db'),
  '-telemetry=false',
  '-client', join(REPO, 'server', 'data', 'deploy', 'client'),
  '-assets', join(REPO, 'server', 'data', 'deploy', 'assets'),
], { stdio: ['ignore', 'pipe', 'pipe'] });
srv.stdout.on('data', () => {});
srv.stderr.on('data', () => {});
process.on('exit', () => srv.kill());

await sleep(3000);

// ── 2. Một lượt đo ───────────────────────────────────────────────────────────────────────
async function round(browser) {
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 720 } });
  const watcher = await ctx.newPage();
  const walker = await ctx.newPage();

  const join = async (page, name) => {
    await page.goto(`http://127.0.0.1:${PORT}/?debug=true`, { waitUntil: 'domcontentloaded' });
    await page.waitForSelector('input', { timeout: 30000 });
    await page.fill('input', name);
    await page.getByText('Vào chơi', { exact: true }).click();
    await page.waitForFunction(() => window.kx?.playing?.() === true, null, { timeout: 30000 });
  };
  await join(watcher, 'QuanSat');
  await join(walker, 'NguoiDi');
  await sleep(1500);

  // KHÔNG lên cảnh giới, KHÔNG thêm kiếm: khung nhìn phải NHỎ để một quãng đi cố định đủ
  // cắt qua ranh giới AOI cả hai chiều, và người quan sát không được giết người đi.
  //
  // Bản trước thả 3 bot cạnh mình. Nó cho `rời = 3` đẹp đẽ nhưng `quay lại = 0`: bot đi mất
  // hoặc chết, không bao giờ trở lại. Mà `respawn = 0` khi KHÔNG AI quay lại là con số hiển
  // nhiên, không phải kết quả đo.

  // Người đi TỰ TÌM TỚI người quan sát. Hai tab sinh ra ở hai chỗ ngẫu nhiên cách nhau cả
  // nghìn đơn vị; không có bước này thì họ không bao giờ thấy nhau.
  const posOf = (page) => page.evaluate(() => {
    const me = window.kx.world.me();
    return me ? { x: me.x, y: me.y } : null;
  });
  const target = await posOf(watcher);
  for (let i = 0; i < 120; i++) {
    const me = await posOf(walker);
    if (!me || !target) { break; }
    const dx = target.x - me.x;
    const dy = target.y - me.y;
    if (Math.hypot(dx, dy) < 220) { break; }
    const keys = [];
    if (dx > 60) { keys.push('KeyD'); } else if (dx < -60) { keys.push('KeyA'); }
    if (dy > 60) { keys.push('KeyS'); } else if (dy < -60) { keys.push('KeyW'); }
    for (const k of keys) { await walker.keyboard.down(k); }
    await sleep(250);
    for (const k of keys) { await walker.keyboard.up(k); }
  }
  // Thả thêm vài bot NGAY CẠNH mình: bot do phòng tự sinh nằm ở chỗ ngẫu nhiên, có thể cả
  // lượt không đi ngang qua. Thả cạnh mình thì chắc chắn chúng ở trong khung nhìn lúc bắt
  // đầu, và đường đi cố định bên dưới sẽ kéo chúng qua ranh giới.
  for (let i = 0; i < 3; i++) { await watcher.keyboard.press('b'); await sleep(200); }
  await sleep(1200);

  // Bắt đầu đếm từ tab quan sát.
  const watch = watcher.evaluate(async (ms) => {
    const w = window.kx.world;
    const base = { ds: w.respawnBy.ds, hethan: w.respawnBy.hethan };
    const goneAt = new Map();
    let gone = 0, respawn2s = 0, placeholder = 0, frames = 0, maxSeen = 0;
    // `quayLai` đếm MỌI lần entity trở lại, không chỉ trong 2 giây.
    //
    // Thiếu con số này thì bài đo tự lừa mình: nếu entity đi rồi không bao giờ quay lại,
    // `respawn2s = 0` là hiển nhiên và chẳng chứng minh được gì. Có quay lại thì số 0 kia
    // mới có nghĩa là "quay lại mà KHÔNG nhấp nháy".
    let quayLai = 0;
    let prev = new Set();
    const t0 = performance.now();
    while (performance.now() - t0 < ms) {
      const now = performance.now();
      const cur = new Set(w.players.keys());
      maxSeen = Math.max(maxSeen, cur.size);
      for (const id of prev) { if (!cur.has(id)) { goneAt.set(id, now); gone++; } }
      for (const id of cur) {
        const at = goneAt.get(id);
        if (at !== undefined) { quayLai++; if (now - at < 2000) { respawn2s++; } goneAt.delete(id); }
      }
      prev = cur;
      frames++;
      for (const sp of window.kx.scene.players.values()) {
        if (!sp.hasAtlas()) { placeholder++; break; }
      }
      await new Promise((r) => setTimeout(r, 50));
    }
    return {
      gone, quayLai, respawn2s, placeholder, frames, maxSeen,
      doDespawn: w.respawnBy.ds - base.ds,
      doHetHan: w.respawnBy.hethan - base.hethan,
    };
  }, WATCH_MS);

  // **NGƯỜI QUAN SÁT là người đi**, và đi theo KHOẢNG CÁCH chứ không theo đồng hồ.
  //
  // Hai lần sửa trước đều hỏng vì đoán quãng đường bằng thời gian:
  //  - Bản đầu đi 4,5 giây mỗi chiều. Đo ra chỉ **798 đơn vị trong 11 giây** (≈72/giây,
  //    không phải 195 như `move_base` — vật cản và va chạm ăn mất phần lớn), trong khi bán
  //    kính RA là nửa khung nhìn 830 + lề 900 = **1 730**. Chưa bao giờ ra khỏi tầm.
  //  - Bản sau cho `walker` đi: nó ra được, nhưng ngoài kia **bị bot chém chết và hồi sinh
  //    với id mới**, nên với tab quan sát đó là người khác, không phải "quay lại".
  //
  // Đi theo khoảng cách thì không phải đoán gì: cứ đi tới khi vượt hẳn bán kính ra, đứng
  // chờ qua hạn trễ, rồi quay lại tới khi gần. Chậm hay nhanh, vướng hay không, bài đo vẫn
  // chạm đúng thứ nó định đo.
  const posOfPage = async (page) =>
    page.evaluate(() => {
      const m = window.kx.world.me();
      return m ? { x: m.x, y: m.y } : null;
    });
  // Mốc đo là **chỗ người quan sát đứng lúc bắt đầu**, không phải khoảng cách tới tab kia.
  //
  // Đo theo tab kia thì bài đo phụ thuộc chỗ hai tab ngẫu nhiên sinh ra: một lượt chúng
  // cách nhau 3 900 đơn vị và không bao giờ nhìn thấy nhau, lượt khác lại sát nhau. Chạy 4
  // lượt ra 0 lần quay lại, chạy 2 lượt khác ra 2 — cùng một bản build. Đó là đúng loại
  // dao động mà cả bài đo này sinh ra để xoá.
  //
  // Đo theo chỗ mình đứng thì **mọi** entity đang trong khung nhìn lúc bắt đầu (3 con bot
  // thả sát bên + tab kia nếu có) đều bị kéo ra rồi kéo vào — không phụ thuộc may rủi.
  const home = await posOfPage(watcher);
  if (!home) {
    throw new Error('không đọc được vị trí người quan sát');
  }
  /**
   * walkUntil đi theo hướng `key` cho tới khi `ok(khoảng cách tới mốc)` đúng, hoặc hết hạn.
   *
   * **Biết lách vật cản.** Bản trước chỉ giữ đúng một phím: trên bản đồ cũ (mật độ chân đế
   * 1–6%) đi thẳng 2 400 đơn vị gần như lúc nào cũng lọt. Trên bản đồ v6.12 (5–7%, vật cản
   * xếp thành đảo) thì người đi đâm vào một cái đảo rồi **đứng ì tại đó cho hết giờ** — bài
   * đo in ra `quay lại=0` kèm cảnh báo "chưa chạm tới thứ nó định đo", tức là không đo gì cả.
   *
   * Cách lách: cứ 400 ms xem đã nhích được chưa. Kẹt hai nhịp liền thì giữ thêm phím dọc
   * (W hoặc S, đổi bên mỗi lần kẹt) trong một nhịp để trượt dọc mặt đảo rồi đi tiếp. Không
   * cần tìm đường gì cả — đảo rộng nhiều nhất 800, trượt vài nhịp là qua.
   */
  const walkUntil = async (key, ok, limitMs) => {
    const start = Date.now();
    await watcher.keyboard.down(key);
    let last = await posOfPage(watcher);
    let stuck = 0;
    let side = 'KeyW';
    try {
      while (Date.now() - start < limitMs) {
        await sleep(400);
        const a2 = await posOfPage(watcher);
        if (!a2) {
          return false;
        }
        const dnow = Math.hypot(a2.x - home.x, a2.y - home.y);
        if (process.env.KX_BENCH_DEBUG) {
          console.log(`    [] d=${dnow.toFixed(0)} pos=(${a2.x.toFixed(0)},${a2.y.toFixed(0)})`);
        }
        if (ok(dnow)) {
          return true;
        }
        // Nhích dưới 30 đơn vị trong 400 ms ở tốc độ ~195/s nghĩa là đang bị chặn.
        if (last && Math.hypot(a2.x - last.x, a2.y - last.y) < 30) {
          stuck++;
        } else {
          stuck = 0;
        }
        last = a2;
        if (stuck >= 2) {
          stuck = 0;
          await watcher.keyboard.down(side);
          await sleep(700);
          await watcher.keyboard.up(side);
          side = side === 'KeyW' ? 'KeyS' : 'KeyW'; // lần sau lách bên kia
        }
      }
    } finally {
      await watcher.keyboard.up(key);
    }
    return false;
  };

  const t0 = Date.now();
  while (Date.now() - t0 < WATCH_MS) {
    // Ra xa hẳn bán kính ra (nửa khung nhìn + lề 900), cộng biên an toàn.
    if (!(await walkUntil('KeyD', (d) => d > OUT_DIST, LEG_LIMIT_MS))) {
      break;
    }
    await sleep(DWELL_MS); // qua hạn trễ `aoiLingerSecs` = 2 giây → server bắn `Despawn`
    if (Date.now() - t0 >= WATCH_MS) {
      break;
    }
    // Quay lại cho tới khi vào hẳn trong khung nhìn.
    if (!(await walkUntil('KeyA', (d) => d < BACK_DIST, LEG_LIMIT_MS))) {
      break;
    }
    await sleep(1200);
  }
  const out = await watch;
  await ctx.close();
  return out;
}

// ── 3. Chạy nhiều lượt rồi tổng hợp ─────────────────────────────────────────────────────
const browser = await chromium.launch({ headless: true, args: ['--use-gl=angle', '--use-angle=gl'] });
const rows = [];
for (let i = 0; i < ROUNDS; i++) {
  rows.push(await round(browser));
  const r = rows[rows.length - 1];
  console.log(`lượt ${i + 1}: respawn=${r.respawn2s} · quay lại=${r.quayLai} · rời=${r.gone} · placeholder=${r.placeholder} · thấy=${r.maxSeen}`);
}
await browser.close();
srv.kill();

const sum = (f) => rows.reduce((t, r) => t + f(r), 0);
const vals = rows.map((r) => r.respawn2s);
console.log('\n── tổng hợp ──');
console.log(`respawn: tổng ${sum((r) => r.respawn2s)} · mỗi lượt ${vals.join(', ')}`);
console.log(`placeholder: tổng ${sum((r) => r.placeholder)} khung / ${sum((r) => r.frames)}`);
console.log(`rời khung ${sum((r) => r.gone)} lần · quay lại ${sum((r) => r.quayLai)} lần`);
console.log(`nhấp nháy do Despawn ${sum((r) => r.doDespawn)} · do luật hết hạn 2s ${sum((r) => r.doHetHan)}`);
if (sum((r) => r.quayLai) === 0) {
  // Không có lần quay lại nào thì `respawn = 0` là hiển nhiên, không phải kết quả.
  console.error('\nCẢNH BÁO: không entity nào QUAY LẠI khung nhìn — bài đo chưa chạm tới thứ nó định đo.');
  process.exit(2);
}
// Thoát khác 0 khi vượt ngân sách §1 — chạy được trong CI.
process.exit(sum((r) => r.respawn2s) === 0 && sum((r) => r.placeholder) === 0 ? 0 : 1);
