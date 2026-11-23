// So nhiều lượt chạy `leakcheck.mjs` với nhau — GAME_SPEC §13.9 (v6.57).
//
//   node tools/browsertest/leakcompare.mjs server/data/leak-repeat/*.json
//
// Hai câu hỏi mà một lượt đo không trả lời được, nhưng ba lượt thì có:
//
//   1. "Kịch bản đã cố định thật chưa?" — ba lượt cùng commit mà dốc còn chênh > 3 lần
//      thì chưa, và spec bắt báo lại trước khi đi tiếp. Mốc so sánh: sáu lượt đo kiểu cũ
//      chênh nhau hơn 10 lần (0,19 → 2,00 MB/phút), đó là lý do v6.57 viết lại cách đo.
//   2. "Hai kịch bản có khác nhau thật không?" (vd 50 lượt đổi khu so với 100) — spec đòi
//      ≥ 3 lượt mỗi bên; truyền hai nhóm tệp vào rồi đọc hai dòng.
import { readFileSync } from 'node:fs';

const files = process.argv.slice(2);
if (files.length === 0) {
  console.error('dùng: node tools/browsertest/leakcompare.mjs <các tệp .json của leakcheck>');
  process.exit(2);
}
const runs = files.map((f) => ({ f, j: JSON.parse(readFileSync(f, 'utf8')) }));
const NAMES = ['heapMB', 'texMB', 'texture', 'display object'];

let shaky = false;
for (const name of NAMES) {
  const got = runs.map((r) => r.j.metrics?.[name]).filter(Boolean);
  if (got.length === 0) {
    continue;
  }
  console.log(`\n${name}`);
  runs.forEach((r, i) => {
    const m = r.j.metrics?.[name];
    console.log(`  lượt ${i + 1}: dốc ${(m ? (m.mean >= 0 ? '+' : '') + m.mean.toFixed(3) : '—').padStart(8)}/phút`
      + (m ? ` · KTC [${m.ci_lo.toFixed(2)}; ${m.ci_hi.toFixed(2)}] · nền ${m.baseline.toFixed(1)} · 60 phút ${m.rise_pct >= 0 ? '+' : ''}${m.rise_pct.toFixed(1)}% nền` : ''));
  });
  // Chỉ so tỉ lệ khi CÓ TÍN HIỆU.
  //
  // Phép "chênh ≤ 3 lần" là để biết kịch bản có cố định hay không, và nó chỉ có nghĩa với
  // một chỉ số đang dốc thật. Nếu KTC của mọi lượt đều ôm qua 0 thì cả ba con số là nhiễu
  // quanh 0, và tỉ lệ giữa hai con nhiễu thì muốn bao nhiêu cũng được — đo 23/11: texture
  // 0,107 / 0,261 / 0,393 (cả ba KTC ôm 0) ra "chênh 3,7 lần", rồi báo oan là kịch bản chưa
  // cố định, trong khi heap — chỉ số DUY NHẤT có tín hiệu — chỉ chênh 2,2 lần.
  const hasSignal = got.some((m) => m.ci_lo > 0 || m.ci_hi < 0);
  if (!hasSignal) {
    console.log('  mọi lượt đều KHÔNG kết luận được là dốc lên (KTC ôm qua 0) — tỉ lệ chênh vô nghĩa, bỏ qua');
    continue;
  }
  // Dốc gần 0 thì tỉ lệ vẫn vô nghĩa (chia cho số rất nhỏ), nên bỏ qua những lượt dưới sàn.
  const abs = got.map((m) => Math.abs(m.mean)).filter((v) => v > 0.05);
  if (abs.length >= 2) {
    const ratio = Math.max(...abs) / Math.min(...abs);
    const bad = ratio > 3;
    shaky = shaky || bad;
    console.log(`  chênh lớn nhất/nhỏ nhất = ${ratio.toFixed(1)} lần`
      + (bad ? '  ⚠ > 3 lần — kịch bản CHƯA cố định thật (§13.9 v6.57: báo chủ dự án trước khi đi tiếp)' : '  ✓ ≤ 3 lần'));
  } else {
    console.log('  mọi lượt đều phẳng (|dốc| ≤ 0,05) — không có gì để so tỉ lệ');
  }
}

console.log('\nkịch bản từng lượt (phải giống hệt nhau):');
for (const [i, r] of runs.entries()) {
  const s = r.j.scenario ?? {};
  console.log(`  lượt ${i + 1} ${r.f}: ${r.j.verdict} · ${s.zone_switches_window} lượt đổi khu trong cửa sổ · ${s.deaths_window} lượt chết · khu kết thúc ${s.end_zone} · build ${r.j.build}`);
}
if (shaky) {
  process.exit(1);
}
