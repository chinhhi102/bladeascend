// BÃO HOÀ hay RÒ RỈ — phân định cho Q57 (GAME_SPEC §13.9 v6.59).
//
//   node tools/browsertest/leaksaturation.mjs <tệp .json của leakcheck>
//
// ── Vì sao phép phân định nằm trong CODE, không nằm trong đầu người đọc ─────────────────
//
// Chủ dự án yêu cầu ghi rõ "cái gì sẽ phân định" TRƯỚC khi chạy, "để lượt đo không kết thúc
// bằng *có vẻ như*". Một bảng số 480 dòng thì ai đọc cũng thấy được thứ mình muốn thấy; luật
// viết thành code thì không.
//
// ── Luật ────────────────────────────────────────────────────────────────────────────────
//
//   BÃO HOÀ        = `glBatchKB` ĐỨNG YÊN trong ≥ 40 phút cuối, dù vẫn đổi khu và vẫn chết.
//                    → không phải rò rỉ, mà là một khoản đỉnh MỘT LẦN: vào bảng ngân sách
//                    §13.10 thành một dòng riêng, và sửa bằng cách chặn trần kích thước lô.
//   KHÔNG BÃO HOÀ  = còn dốc dương sau 2 giờ → rò rỉ thật, lúc đó mới bàn tới cách vẽ.
//
// "Đứng yên" = một trong hai điều, vì hai điều này hỏng theo hai cách khác nhau:
//   (a) KTC 95% của dốc 40 phút cuối ôm qua 0 — không chứng minh được là còn tăng; HOẶC
//   (b) |dốc × 40 phút| < 2% giá trị trung bình — có thể còn tăng nhưng quá nhỏ để đáng gọi
//       là rò rỉ (2% của ~39 MB là 0,8 MB trong 40 phút).
//
// Và phải kèm điều kiện kịch bản VẪN CHẠY trong 40 phút đó: một đồ thị phẳng vì nhân vật
// đứng im ở màn Tổng kết thì phẳng vì lý do khác hẳn.
import { readFileSync } from 'node:fs';

const path = process.argv[2];
if (!path) {
  console.error('dùng: node tools/browsertest/leaksaturation.mjs <tệp .json của leakcheck>');
  process.exit(2);
}
const j = JSON.parse(readFileSync(path, 'utf8'));
const samples = j.samples ?? [];
if (samples.length === 0) {
  console.error('tệp không có mẫu thô (`samples`) — cần bản leakcheck từ 23/11 trở đi.');
  process.exit(2);
}

/** TAIL_MIN: độ dài cửa sổ cuối dùng để phán quyết (spec v6.59: ≥ 40 phút). */
const TAIL_MIN = Number(process.env.TAIL_MIN ?? 40);
/** FLAT_PCT: ngưỡng "đủ nhỏ để không gọi là rò rỉ", % của giá trị trung bình. */
const FLAT_PCT = Number(process.env.FLAT_PCT ?? 2);

const T95 = { 1: 12.706, 2: 4.303, 3: 3.182, 4: 2.776, 5: 2.571, 6: 2.447, 7: 2.365, 8: 2.306, 9: 2.262, 10: 2.228, 15: 2.131, 20: 2.086, 30: 2.042, 60: 2.0 };
const tcrit = (df) => {
  if (df <= 0) {
    return Infinity;
  }
  const keys = Object.keys(T95).map(Number).sort((a, b) => a - b);
  for (const k of keys) {
    if (df <= k) {
      return T95[k];
    }
  }
  return 1.96;
};

/** trend: dốc OLS + KTC 95% từ sai số dư. */
function trend(pts) {
  const n = pts.length;
  const mx = pts.reduce((a, [x]) => a + x, 0) / n;
  const my = pts.reduce((a, [, y]) => a + y, 0) / n;
  const sxx = pts.reduce((a, [x]) => a + (x - mx) ** 2, 0);
  const slope = sxx > 0 ? pts.reduce((a, [x, y]) => a + (x - mx) * (y - my), 0) / sxx : 0;
  const intercept = my - slope * mx;
  const sse = pts.reduce((a, [x, y]) => a + (y - (intercept + slope * x)) ** 2, 0);
  const se = n > 2 && sxx > 0 ? Math.sqrt(sse / (n - 2) / sxx) : Infinity;
  const half = tcrit(n - 2) * se;
  return { slope, lo: slope - half, hi: slope + half, mean: my, n };
}

const last = samples[samples.length - 1].min;
const tail = samples.filter((s) => s.min >= last - TAIL_MIN);
const get = (k) => (s) => s.kinds?.[k] ?? 0;

console.log(`${path}`);
console.log(`build ${j.build} · ${samples.length} mẫu · tới phút ${last} · cửa sổ phán quyết: ${TAIL_MIN} phút cuối (${tail.length} mẫu)\n`);

// Kịch bản có còn chạy trong 40 phút cuối không — nếu không thì mọi đồ thị phẳng đều vô nghĩa.
let zoneChanges = 0;
for (let i = 1; i < tail.length; i++) {
  if (tail[i].zone !== tail[i - 1].zone) {
    zoneChanges++;
  }
}
const alive = tail.filter((s) => s.fps > 5).length / tail.length;
console.log(`kịch bản trong cửa sổ phán quyết: ${zoneChanges} lần đổi khu · ${(alive * 100).toFixed(0)}% mẫu còn vẽ (fps > 5)`);
const scenarioOK = zoneChanges >= TAIL_MIN / 2 && alive > 0.95;
if (!scenarioOK) {
  console.log('→ KHÔNG KẾT LUẬN: kịch bản không chạy đủ trong cửa sổ phán quyết (đồ thị phẳng vì lý do khác).');
  process.exit(3);
}

// Bộ đếm KHÔNG CÓ phải là "không kết luận", không phải "đứng yên".
//
// Chạy thử trên một tệp cũ (build chưa có `glBatchKB`) thì mọi mẫu bằng 0, dốc bằng 0, và
// bài này tuyên "BÃO HOÀ" — tức là đúng cái lỗi "thành công giả" mà §13.9 v6.58 bắt phải
// bịt, lần này nằm ngay trong luật phân định. Một bộ đếm vắng mặt không chứng minh được gì.
const batchVals = tail.map((s) => s.kinds?.n_glBatchKB ?? null);
if (batchVals.every((v) => v === null || v === 0)) {
  console.log('→ KHÔNG KẾT LUẬN: bản build của lượt đo này chưa có bộ đếm `glBatchKB`');
  console.log('   (mọi mẫu bằng 0 — một bộ đếm vắng mặt không phải là "đứng yên").');
  process.exit(3);
}

const ROWS = [
  ['glBatchKB', 'buffer gom lô (KB)'],
  ['glBufferKB', 'tất cả buffer (KB)'],
  ['poolTotal', 'bể Pixi đã cấp phát'],
  ['gfxContexts', 'GraphicsContext sống'],
];
console.log('\nchỉ số                    TB      đầu → cuối       dốc/phút   KTC 95%            kết luận');
let verdictFlat = true;
for (const [key, label] of ROWS) {
  const pts = tail.map((s) => [s.min, get(`n_${key}`)(s)]);
  const t = trend(pts);
  const rise = t.slope * TAIL_MIN;
  const risePct = t.mean > 0 ? (rise / t.mean) * 100 : 0;
  const flat = t.lo <= 0 || Math.abs(risePct) < FLAT_PCT;
  if (key === 'glBatchKB') {
    verdictFlat = flat;
  }
  console.log(`${label.padEnd(24)} ${t.mean.toFixed(0).padStart(7)} ${pts[0][1].toFixed(0).padStart(7)} →${pts[pts.length - 1][1].toFixed(0).padStart(7)}`
    + `   ${(t.slope >= 0 ? '+' : '') + t.slope.toFixed(2).padStart(7)}   [${t.lo.toFixed(2)}; ${t.hi.toFixed(2)}]`.padEnd(32)
    + `${flat ? 'đứng yên' : `CÒN TĂNG (+${risePct.toFixed(1)}% trong ${TAIL_MIN} phút)`}`);
}

// Heap tổng, để đọc kèm: nếu buffer gom lô bão hoà mà heap vẫn dốc thì còn nguyên nhân khác.
const h = trend(tail.map((s) => [s.min, s.heapMB]));
console.log(`\nheapMB (đọc kèm)         ${h.mean.toFixed(1).padStart(7)} ${tail[0].heapMB.toFixed(1).padStart(7)} →${tail[tail.length - 1].heapMB.toFixed(1).padStart(7)}`
  + `   ${(h.slope >= 0 ? '+' : '') + h.slope.toFixed(3)}   [${h.lo.toFixed(3)}; ${h.hi.toFixed(3)}]`);

console.log('');
if (verdictFlat) {
  console.log(`BÃO HOÀ (§13.9 v6.59): buffer gom lô đứng yên trong ${TAIL_MIN} phút cuối, kịch bản vẫn chạy.`);
  console.log('→ ĐÂY KHÔNG PHẢI RÒ RỈ. Nó là một khoản đỉnh MỘT LẦN:');
  console.log('   · vào bảng ngân sách §13.10 thành một dòng riêng;');
  console.log('   · cách sửa là CHẶN TRẦN kích thước lô, KHÔNG đụng VFX/decor.');
  if (h.lo > 0) {
    console.log(`\n⚠ nhưng heap vẫn dốc lên trong cửa sổ này (+${h.slope.toFixed(3)} MB/phút, KTC [${h.lo.toFixed(3)}; ${h.hi.toFixed(3)}])`);
    console.log('   → buffer gom lô không còn là nguyên nhân chính; còn một nguồn khác phải tìm.');
  }
  process.exit(0);
}
console.log(`KHÔNG BÃO HOÀ: buffer gom lô CÒN TĂNG sau ${last} phút.`);
console.log('→ Đây là rò rỉ thật. Lúc này mới bàn tới cách vẽ (gom decor tĩnh / batchMode) — và cả hai cần chủ dự án duyệt vì chúng đổi hình.');
process.exit(1);
