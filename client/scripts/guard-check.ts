// Lớp chống sập phía client — GAME_SPEC §13.8.
//
//   npm run check:guard
//
// Kiểm bằng ĐỒNG HỒ GIẢ (`setSysClock`), nên chạy tức thì và lặp lại được. Mọi luật của
// §13.8 đều là luật về thời gian — "tắt 2 giây", "3 lần trong 10 giây" — nên không có đồng
// hồ giả thì bài này hoặc phải chờ thật, hoặc chỉ kiểm được một nửa.
//
//   1. Lỗi một hệ thống → tắt riêng nó 2 s; các hệ thống khác không bị ảnh hưởng.
//   2. Hết 2 s thì nó tự chạy lại.
//   3. Lỗi 3 lần trong 10 s → tắt hẳn tới hết mạng, không bao giờ thử lại nữa.
//   4. Lỗi thưa (cách nhau > 10 s) KHÔNG cộng dồn — cửa sổ phải trượt.
//   5. Lỗi vẽ: 1–2 lần → "bỏ khung"; lần thứ 3 trong 10 s → "dựng lại renderer".
//   6. Lỗi vẽ bật công tắc tắt filter, và `sysResetRun` (mạng mới) mở lại.
//   7. `kx.fault` ném được lỗi giả đúng số lần đã hẹn (checklist §17).
//
// Đã thử bài có răng: đổi SYS_STRIKES thành 4 thì điều 3 trượt ngay.
import {
  drawGuardStats, faultSystem, filtersOff, killFilters, noteDrawError, setSysClock, setSysErrorHook,
  sysDead, sysLine, sysOff, sysResetRun, sysRun, sysSummary,
} from '../src/render/guard.ts';

let fails = 0;
function expect(name: string, got: unknown, want: unknown): void {
  const ok = got === want;
  if (!ok) {
    fails++;
  }
  console.log(`${ok ? 'đạt ' : 'TRƯỢT'}  ${name}: ${String(got)}${ok ? '' : ` (mong ${String(want)})`}`);
}

// Console sạch: bài này cố tình gây hàng chục lỗi, in hết thì không đọc được kết quả.
console.warn = (): void => undefined;

let t = 1000;
setSysClock(() => t);

// ── 1. Lỗi một hệ thống chỉ tắt chính nó ────────────────────────────────────────────────
let vfxRan = 0;
let hudRan = 0;
const vfxOk = (): void => {
  vfxRan++;
};
const vfxBad = (): never => {
  throw new Error('vfx vỡ');
};

expect('vfx chạy bình thường', sysRun('vfx', vfxOk), true);
expect('vfx lỗi → sysRun trả false', sysRun('vfx', vfxBad), false);
expect('vfx đang bị tắt', sysOff('vfx'), true);
expect('vfx CHƯA tắt hẳn', sysDead('vfx'), false);
sysRun('hud', () => {
  hudRan++;
});
expect('hud vẫn chạy khi vfx hỏng', hudRan, 1);

// ── 2. Hết 2 giây thì tự thử lại ────────────────────────────────────────────────────────
t += 1999;
expect('1,999 s: vfx vẫn tắt', sysRun('vfx', vfxOk), false);
t += 2;
expect('2,001 s: vfx chạy lại', sysRun('vfx', vfxOk), true);
expect('vfx đã chạy 2 lần', vfxRan, 2);

// ── 3. Ba lần trong 10 giây → tắt hẳn ───────────────────────────────────────────────────
sysResetRun();
t = 10_000;
sysRun('beasts', vfxBad); // lần 1
t += 2500;
sysRun('beasts', vfxBad); // lần 2 (đã qua 2 s tắt tạm)
t += 2500;
sysRun('beasts', vfxBad); // lần 3 — trong cùng 10 s
expect('3 lần/10 s → tắt hẳn', sysDead('beasts'), true);
t += 3_600_000; // một tiếng sau
expect('tắt hẳn là tới HẾT MẠNG, không hồi', sysRun('beasts', vfxOk), false);
expect('dòng F3 nói rõ hệ thống nào', sysLine().includes('beasts (tắt hẳn)'), true);

// ── 4. Cửa sổ 10 giây phải TRƯỢT, không cộng dồn vô hạn ─────────────────────────────────
sysResetRun();
t = 100_000;
sysRun('items', vfxBad);
t += 11_000;
sysRun('items', vfxBad);
t += 11_000;
sysRun('items', vfxBad);
expect('3 lỗi cách nhau 11 s → KHÔNG tắt hẳn', sysDead('items'), false);
expect('nhưng vẫn ghi đủ 3 lần', sysSummary().find((s) => s.key === 'items')?.errors, 3);

// ── 5. Lỗi vẽ: bỏ khung, lần thứ 3 thì dựng lại renderer ────────────────────────────────
sysResetRun();
t = 200_000;
expect('lỗi vẽ lần 1 → bỏ khung', noteDrawError(new Error('uProjectionMatrix')), 'skip');
expect('lỗi vẽ lần 2 → bỏ khung', noteDrawError(new Error('uProjectionMatrix')), 'skip');
expect('lỗi vẽ lần 3 → dựng lại', noteDrawError(new Error('uProjectionMatrix')), 'rebuild');
// Sau khi đã dựng lại, bộ đếm về 0 — nếu không thì mọi khung sau đó đều đòi dựng lại.
expect('ngay sau khi dựng lại → lại chỉ bỏ khung', noteDrawError(new Error('x')), 'skip');
expect('đếm đúng số lần dựng lại', drawGuardStats().rebuilds, 1);

// ── 6. Công tắc tắt filter ──────────────────────────────────────────────────────────────
expect('mạng mới: filter đang bật', filtersOff(), false);
killFilters();
expect('sau lỗi vẽ: filter tắt', filtersOff(), true);
sysResetRun();
expect('mạng mới mở lại filter', filtersOff(), false);

// ── 7. Lỗi giả của checklist §17 ────────────────────────────────────────────────────────
sysResetRun();
t = 300_000;
let ran = 0;
const count = (): void => {
  ran++;
};
faultSystem('minimap', 2);
expect('fault lần 1 → ném', sysRun('minimap', count), false);
t += 2100;
expect('fault lần 2 → ném', sysRun('minimap', count), false);
t += 2100;
expect('hết hẹn → chạy lại bình thường', sysRun('minimap', count), true);
expect('hàm thật chỉ chạy đúng 1 lần', ran, 1);

// ── Móc báo cáo ra ngoài (telemetry) ────────────────────────────────────────────────────
sysResetRun();
t = 400_000;
const seen: string[] = [];
setSysErrorHook(({ key, dead }) => seen.push(`${key}/${dead}`));
sysRun('ambient', vfxBad);
t += 2100;
sysRun('ambient', vfxBad);
t += 2100;
sysRun('ambient', vfxBad);
setSysErrorHook(null);
expect('mọi lần tắt đều báo ra ngoài', seen.join(' '), 'ambient/false ambient/false ambient/true');

setSysClock(null);
if (fails > 0) {
  console.error(`\n${fails} điều TRƯỢT`);
  process.exit(1);
}
console.log('\nLớp chống sập client (§13.8): đạt hết');
