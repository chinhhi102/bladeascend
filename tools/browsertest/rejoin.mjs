// Nghiệm thu VÀO LẠI PHIÊN — GAME_SPEC §7.1, checklist §17 dòng 3:
//
//   "F5 giữa trận → trong 20 s vào lại đúng nhân vật (cảnh giới, kiếm, Uy Danh, khu);
//    bị chém trong lúc mất kết nối → chết, F5 xong thấy Tổng kết."
//
//   cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true
//   node tools/browsertest/rejoin.mjs http://127.0.0.1:9998
import { launch } from './browser.mjs';

const TARGET = process.argv[2] ?? 'http://127.0.0.1:9998';
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

let fails = 0;
function expect(name, got, want, detail) {
  const ok = got === want;
  if (!ok) {
    fails++;
  }
  const shown = detail ?? String(got);
  console.log(`${ok ? 'đạt ' : 'TRƯỢT'}  ${name}: ${shown}${ok ? '' : ` (mong ${String(want)})`}`);
}

const b = await launch({ headless: true, args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist'] });
const ctx = await b.newContext({ viewport: { width: 1280, height: 720 }, locale: 'vi-VN' });
const p = await ctx.newPage();
p.on('pageerror', (e) => console.log('  pageerror', String(e).slice(0, 160)));

async function joinMatch(name) {
  await p.goto(`${TARGET}/?debug=true`, { waitUntil: 'domcontentloaded' });
  await p.waitForSelector('input', { timeout: 30000 });
  await p.fill('input', name);
  await p.locator('.char-card').nth(0).click();
  await p.locator('.home .btn.primary').click();
  await p.waitForSelector('.room-board', { timeout: 60000 });
  await sleep(3500);
}

/** me đọc trạng thái nhân vật của chính mình qua cửa debug. */
const me = () => p.evaluate(() => {
  const m = window.kx?.world?.me?.();
  return m ? { id: m.id, level: m.L, swords: m.n } : null;
});
const session = () => p.evaluate(() => {
  try {
    return JSON.parse(sessionStorage.getItem('kx.session') ?? 'null');
  } catch {
    return null;
  }
});

await joinMatch('KiemTraVaoLai');

// Nuôi nhân vật lên cho khác hẳn trạng thái "mới vào", VÀ cho bất tử trong lúc đo.
//
// `max n=1` = lên Đạo Tổ đủ kiếm và GIỮ bất tử (`debug.go`: n=0 mới tắt bất tử). Bất tử là
// phần bắt buộc, không phải tiện tay: §7.1 nói rõ nhân vật ở lại thế giới lúc mất kết nối và
// "vẫn chém chết được". Bài đo cũ nuôi cảnh giới rồi so "giữ nguyên số kiếm" trước/sau F5 —
// nhưng trong 3 giây tải lại trang, một con bot đi ngang chém vài nhát là số kiếm đổi thật.
// Ngày 12/11 nó trượt đúng kiểu đó (13 → 7 kiếm) và chặn một lượt deploy, trong khi luật
// đang chạy ĐÚNG. Bài đo phải tách "nhân vật có được giữ lại không" khỏi "có ai chém nó
// không"; cái thứ hai đã có bài khác lo.
await p.evaluate(() => window.kx.conn.send({ t: 'dbg', cmd: 'max', n: 1 }));
await sleep(2500);

const before = await me();
const keyBefore = (await session())?.key ?? '';
console.log(`trước F5: id ${before?.id} · cảnh giới ${before?.level} · ${before?.swords} kiếm`);
expect('đã có mã phiên 32 byte', keyBefore.length, 64);
expect('cảnh giới đã lên (khác lúc mới vào)', (before?.level ?? 0) > 0, true);

// ── F5 giữa trận ────────────────────────────────────────────────────────────────────────
await p.reload({ waitUntil: 'domcontentloaded' });
await p.waitForSelector('.room-board', { timeout: 60000 });
await sleep(3000);

const after = await me();
const keyAfter = (await session())?.key ?? '';
console.log(`sau  F5: id ${after?.id} · cảnh giới ${after?.level} · ${after?.swords} kiếm`);
expect('vào lại ĐÚNG nhân vật (cùng id)', after?.id, before?.id);
expect('giữ nguyên cảnh giới', after?.level, before?.level);
// Số kiếm so theo BIÊN 2%, không so bằng nhau tuyệt đối.
//
// Bất tử chặn được "bị chém", nhưng KHÔNG chặn được luật bể kiếm: kiếm vỡ khi quệt vật cản
// rồi mọc lại dần (§4). Ở Đạo Tổ 1 050 kiếm, ba giây tải lại trang đủ để con số nhúc nhích
// vài đơn vị — 13/11 một lượt deploy trượt vì 1 040 → 1 038 trong khi §7.1 chạy đúng.
// Thứ bài này phải bắt là "F5 xong tiến độ bị dựng lại từ đầu", và cái đó rơi về vài chục
// kiếm chứ không phải hai.
const swordTol = Math.max(2, Math.ceil((before?.swords ?? 0) * 0.02));
expect(`giữ nguyên số kiếm (biên ±${swordTol})`,
  Math.abs((after?.swords ?? 0) - (before?.swords ?? 0)) <= swordTol, true,
  `${after?.swords} so với ${before?.swords}`);
expect('không qua trang chủ', await p.evaluate(() => !!document.querySelector('.room-board')), true);
expect('đã cấp mã phiên MỚI', keyAfter.length === 64 && keyAfter !== keyBefore, true);

// ── Bị chém chết trong lúc "mất kết nối" ────────────────────────────────────────────────
//
// Đúng câu của checklist §17: "bị chém trong lúc mất kết nối → chết, F5 xong thấy Tổng kết".
// Nên phải có NGƯỜI KHÁC chém: mở tab thứ hai, tab 1 tải lại trang (nhân vật vào diện giữ
// 20 giây), rồi tab 2 gửi `killid` vào đúng id đó trong lúc tab 1 còn đang dựng lại.
//
// Bản cũ làm bằng cách tự gửi `die` rồi hẹn `location.reload()` sau 120 ms. Đó là một cuộc
// đua giữa việc đẩy gói đi và việc trình duyệt đóng socket: máy rảnh thì gói kịp đi (chết,
// thấy Tổng kết), máy bận thì không kịp (nhân vật còn sống, vào thẳng trận). Ngày 12/11 nó
// chặn một lượt deploy của đúng tính năng đang chạy đúng. Một bài đo mà kết quả phụ thuộc
// tải máy thì nó đo tải máy, không đo tính năng.
console.log('\n— bị người khác chém chết trong lúc mất kết nối —');
const p2 = await ctx.newPage();
await p2.goto(`${TARGET}/?debug=true`, { waitUntil: 'domcontentloaded' });
await p2.waitForSelector('input', { timeout: 30000 });
await p2.fill('input', 'KiemTraKeChem');
await p2.locator('.char-card').nth(0).click();
await p2.locator('.home .btn.primary').click();
await p2.waitForSelector('.room-board', { timeout: 60000 });
await sleep(2000);

const victimID = after?.id;
await p.evaluate(() => location.reload());
// Chờ cho tab 1 rụng hẳn kết nối rồi mới chém: chém sớm quá thì nó chết lúc CÒN kết nối,
// tức là lại kiểm một đường khác.
// Chém HAI NHỊP, 1,5 s và 3,0 s sau khi bấm tải lại.
//
// Cửa sổ "đang mất kết nối" co giãn theo tải máy: lúc `tools/deploy.sh` chạy, máy vừa build
// vừa chạy server nên tab 1 có khi đã vào lại trước nhịp đầu, có khi socket còn chưa rụng.
// Một nhịp duy nhất đặt ở 1,5 s là đoán mò giữa hai đầu đó, và 13/11 nó trượt một lượt
// deploy trong khi chạy riêng thì xanh 3/3. `killid` bỏ qua người đã chết (`!t.Dead` trong
// `debug.go`) nên nhịp thứ hai không bao giờ chém nhầm hai lần.
const strike = () =>
  p2.evaluate((id) => window.kx.conn.send({ t: 'dbg', cmd: 'killid', n: id }), victimID);
await sleep(1500);
await strike();
await sleep(1500);
await strike();
// Chờ tới khi trang dựng xong HẲN một trong ba màn, đừng chờ cứng theo đồng hồ: lúc
// `tools/deploy.sh` chạy bài này thì máy đang vừa build vừa chạy server, và 12/11 một lượt
// deploy đã TRƯỢT chỉ vì sau 9 giây trang còn đang nạp — không màn nào tồn tại, nên bài đọc
// ra "không thấy Tổng kết" rồi kết luận sai về một tính năng đang chạy đúng.
await p.waitForSelector('.summary, .room-board, .home', { timeout: 45000 }).catch(() => {});
// Dựng xong rồi mới nghỉ một nhịp: nếu nó ĐỊNH tự nhảy vào trận mới thì nhịp này đủ để lộ ra.
await sleep(2000);
const onSummary = await p.evaluate(() => !!document.querySelector('.summary'));
const inMatch = await p.evaluate(() => !!document.querySelector('.room-board'));
if (!(onSummary && !inMatch)) {
  // In trạng thái thật khi trượt. Một dòng "false (mong true)" không phân biệt được
  // "tự vào trận mới" (lỗi thật) với "trang chưa nạp xong" (lỗi của phép đo).
  console.log('  trạng thái lúc trượt:', JSON.stringify(await p.evaluate(() => ({
    summary: !!document.querySelector('.summary'),
    board: !!document.querySelector('.room-board'),
    home: !!document.querySelector('.home'),
    body: document.body.className,
    ready: document.readyState,
  }))));
}
expect('thấy màn Tổng kết, KHÔNG tự vào trận mới', onSummary && !inMatch, true);

await b.close();
if (fails > 0) {
  console.error(`\n${fails} điều TRƯỢT`);
  process.exit(1);
}
console.log('\nVào lại phiên (§7.1 / checklist §17): đạt hết');
