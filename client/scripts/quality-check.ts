// Luật chọn mức chất lượng — chủ dự án chốt 1/11.
//
//   npm run check:quality
//
// Kiểm ba điều bằng đồng hồ GIẢ (truyền `now` tay), nên chạy tức thì và lặp lại được:
//   1. Mặc định: điện thoại / GPU tích hợp → Thấp, PC còn lại → Vừa. Không ai bắt đầu ở Cao.
//   2. Vừa → Cao chỉ khi ≥ 55 FPS suốt 20 giây; 19 giây là CHƯA.
//   3. Tụt dưới 45 FPS suốt 3 giây thì hạ một bậc; Thấp → Vừa vẫn 10 giây như spec §13.6.
//
// Đã thử bài có răng: đặt `UP_TO_HIGH_SECS` về 10 thì điều 2 trượt ngay.
import { autoQuality, defaultQuality, quality, setQuality } from '../src/render/quality.ts';

let fails = 0;
function expect(name: string, got: unknown, want: unknown): void {
  const ok = got === want;
  if (!ok) {
    fails++;
  }
  console.log(`${ok ? 'đạt ' : 'TRƯỢT'}  ${name}: ${String(got)}${ok ? '' : ` (mong ${String(want)})`}`);
}

/** chay cho FPS cố định chạy từ giây `from` tới `to`, mỗi khung 1/60 giây. */
function chay(fps: number, from: number, to: number): void {
  for (let t = from; t <= to; t += 1 / 60) {
    autoQuality(fps, t);
  }
}

// 1. Mặc định theo thiết bị
expect('điện thoại', defaultQuality(true, 'Apple GPU'), 'low');
expect('PC Intel UHD', defaultQuality(false, 'ANGLE (Intel, Intel(R) UHD Graphics 620 Direct3D11)'), 'low');
expect('PC Iris Xe', defaultQuality(false, 'ANGLE (Intel, Intel(R) Iris(R) Xe Graphics)'), 'low');
expect('PC Radeon Vega tích hợp', defaultQuality(false, 'AMD Radeon(TM) Vega 8 Graphics'), 'low');
expect('PC SwiftShader', defaultQuality(false, 'Google SwiftShader'), 'low');
expect('PC RTX 3060', defaultQuality(false, 'ANGLE (NVIDIA, NVIDIA GeForce RTX 3060)'), 'medium');
expect('PC không rõ GPU', defaultQuality(false, ''), 'medium');
expect('mức lúc nạp (Node, không có thiết bị)', quality(), 'medium');

// 2. Vừa → Cao cần 20 giây ổn định
setQuality('medium');
chay(60, 0, 19);
expect('Vừa, 60 FPS suốt 19 s', quality(), 'medium');
chay(60, 19, 20.1);
expect('Vừa, 60 FPS suốt 20 s', quality(), 'high');

// Ngắt quãng làm lại từ đầu: 15 s tốt, 1 s ở 50 FPS, 15 s tốt → vẫn chưa lên.
setQuality('medium');
chay(60, 100, 115);
chay(50, 115, 116);
chay(60, 116, 131);
expect('Vừa, 15 s + nghẽn 1 s + 15 s', quality(), 'medium');

// 3. Hạ khi tụt, và Thấp → Vừa vẫn 10 giây
setQuality('high');
chay(40, 200, 203.1);
expect('Cao, 40 FPS suốt 3 s', quality(), 'medium');
setQuality('low');
chay(60, 300, 310.1);
expect('Thấp, 60 FPS suốt 10 s', quality(), 'medium');

if (fails > 0) {
  console.error(`\n${fails} điều TRƯỢT`);
  process.exit(1);
}
console.log('\nLuật chất lượng: đạt hết');
