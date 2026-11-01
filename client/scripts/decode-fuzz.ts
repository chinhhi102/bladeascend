// Gói HỎNG không được làm hỏng client — NETCODE_SPEC §4.9, chủ dự án chốt 1/11.
//
//   npm run check:decode
//
// Vì sao bài này tồn tại. Telemetry 31/10 có một lỗi chưa bắt trên Safari:
//
//     RangeError: Extra 18 of 121 byte(s) found at buffer[103]
//
// Đó là `@msgpack/msgpack` gặp byte thừa. Đường snapshot nhị phân đã có `try/catch` từ
// đầu, nhưng đường msgpack thì KHÔNG — và không ai thấy, vì đoạn giải mã nằm thẳng trong
// `ws.onmessage` nên muốn thử "gói hỏng thì sao" là phải dựng WebSocket giả. Nay nó là
// `decodeFrame()` gọi được từ Node, và bài này gọi nó 12 000 lần với gói rác.
//
// Ba điều phải đúng, và điều thứ ba mới là điều khó:
//
//   1. KHÔNG BAO GIỜ ném — mọi gói rác đều trả về `fail`.
//   2. `fail` phải nói đủ để dựng lại: loại, byte đầu, độ dài, 8 byte hex đầu.
//   3. KHÔNG LÀM HỎNG TRẠNG THÁI. Bộ giải mã snapshot nhớ trạng thái giữa các gói (server
//      chỉ gửi trường đã đổi). Nuốt một gói rác rồi để bộ nhớ trong lệch đi thì tệ hơn hẳn
//      ném lỗi: màn hình sai mà không ai biết. Nên sau mỗi gói rác, bài này gửi tiếp một
//      keyframe HỢP LỆ và đòi nó giải mã ra ĐÚNG như khi chưa gặp rác.
import { encode } from '@msgpack/msgpack';

import { decodeFrame } from '../src/net/frame.ts';
import { SnapshotDecoder, SNAP_MAGIC } from '../src/net/snapwire.ts';
import { encodeSnapHead, MsgType, Writer } from '../src/net/wire.gen.ts';

/** mulberry32 — sinh số giả ngẫu nhiên CÓ SEED, để lần chạy nào cũng ra đúng bộ gói đó. */
function rng(seed: number): () => number {
  let a = seed >>> 0;
  return () => {
    a = (a + 0x6d2b79f5) >>> 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

const rand = rng(20261101);
const randInt = (n: number): number => Math.floor(rand() * n);

let kiemTra = 0;
let loi = 0;
const fail = (msg: string): void => {
  loi++;
  console.error(`  TRƯỢT: ${msg}`);
};

/**
 * Một gói keyframe HỢP LỆ, dựng bằng chính bộ mã hoá của protocol.
 *
 * Không lấy từ `cmd/protofixture`: fixture đó chứa payload của TỪNG LOẠI message rời
 * (Spawn, Update, Desc…), không phải khung snapshot hoàn chỉnh — ghép lại thành khung là
 * tự viết lại bộ ghép của server trong một bài test.
 *
 * Khung chỉ-có-header là hợp lệ và đủ dùng ở đây: `SnapshotDecoder.decode` đọc header rồi
 * lặp `while (r.left > 0)`, nên không section nào thì vòng lặp không chạy. Cờ `flags = 1`
 * đánh dấu keyframe — đúng loại gói mà client dùng để dựng lại trạng thái sau khi mất gói,
 * tức đúng thứ phải còn chạy sau khi gặp rác.
 */
function goiKeyframe(): Uint8Array {
  const w = new Writer();
  w.putU8(MsgType.SnapHead);
  encodeSnapHead(w, { tick: 4242, ack: 7, timeMs: 1234567, srvUs: 900, flags: 1 });
  return w.bytes();
}

// ── 1. Golden: những gói hỏng ĐÃ TỪNG xảy ra hoặc nhắm thẳng vào biên ────────────────
//
// "Extra N byte" là chính cái lỗi 31/10: msgpack hợp lệ rồi có rác đằng sau.
const golden: { ten: string; goi: Uint8Array }[] = [
  { ten: 'rỗng', goi: new Uint8Array(0) },
  { ten: 'một byte 0x00', goi: new Uint8Array([0]) },
  { ten: 'chỉ magic snapshot', goi: new Uint8Array([SNAP_MAGIC]) },
  { ten: 'snapshot cụt giữa chừng', goi: new Uint8Array([SNAP_MAGIC, 1, 2, 3]) },
  {
    // ĐÚNG hình dạng lỗi 31/10: msgpack hợp lệ + byte thừa phía sau.
    ten: 'msgpack hợp lệ kèm byte thừa (lỗi 31/10)',
    goi: (() => {
      const ok = encode({ t: 'pong', c: 123 });
      const out = new Uint8Array(ok.length + 18);
      out.set(ok, 0);
      out.fill(0x5a, ok.length);
      return out;
    })(),
  },
  { ten: 'msgpack báo map dài mà không có dữ liệu', goi: new Uint8Array([0xde, 0xff, 0xff]) },
  { ten: 'msgpack báo chuỗi 4GB', goi: new Uint8Array([0xdb, 0xff, 0xff, 0xff, 0xff]) },
  { ten: 'toàn 0xff', goi: new Uint8Array(64).fill(0xff) },
];

console.log('Golden — gói hỏng đã biết:');
for (const g of golden) {
  kiemTra++;
  const snap = new SnapshotDecoder();
  let res;
  try {
    res = decodeFrame(g.goi, snap);
  } catch (err) {
    fail(`"${g.ten}" NÉM lỗi ra ngoài: ${String(err).slice(0, 100)}`);
    continue;
  }
  if (!res.fail) {
    // Giải mã ĐƯỢC cũng không sao — miễn là không ném và không trả về rác im lặng.
    // Nhưng gói rỗng / cụt thì phải bị từ chối.
    if (g.goi.length <= 4) {
      fail(`"${g.ten}" được nhận là gói hợp lệ: ${JSON.stringify(res.msg).slice(0, 80)}`);
    }
    continue;
  }
  if (!res.fail.err || res.fail.len !== g.goi.length) {
    fail(`"${g.ten}" báo lỗi thiếu thông tin: ${JSON.stringify(res.fail)}`);
  }
}
console.log(`  ${golden.length} gói golden, ${loi} trượt`);

// ── 2. Fuzz: 12 000 gói rác có seed cố định ──────────────────────────────────────────
const SO_GOI = 12000;
let soBiTuChoi = 0;
let soGiaiMaDuoc = 0;
for (let i = 0; i < SO_GOI; i++) {
  const n = 1 + randInt(200);
  const goi = new Uint8Array(n);
  for (let j = 0; j < n; j++) {
    goi[j] = randInt(256);
  }
  // Một phần ba số gói ép byte đầu thành magic snapshot, để đường nhị phân cũng bị nện
  // chứ không chỉ đường msgpack.
  if (i % 3 === 0) {
    goi[0] = SNAP_MAGIC;
  }
  const snap = new SnapshotDecoder();
  kiemTra++;
  try {
    const res = decodeFrame(goi, snap);
    if (res.fail) {
      soBiTuChoi++;
    } else {
      soGiaiMaDuoc++;
    }
  } catch (err) {
    fail(`gói fuzz #${i} (${n} byte, byte0=${goi[0]}) NÉM lỗi: ${String(err).slice(0, 100)}`);
    break;
  }
}
console.log(`Fuzz — ${SO_GOI} gói rác: ${soBiTuChoi} bị từ chối, ${soGiaiMaDuoc} tình cờ hợp lệ, 0 lần ném`);

// ── 3. Gói rác KHÔNG được làm lệch trạng thái bộ giải mã ─────────────────────────────
const keyframe = goiKeyframe();
const sach = new SnapshotDecoder();
const mongDoi = JSON.stringify(decodeFrame(keyframe, sach).msg);
kiemTra++;
if (!mongDoi || mongDoi === 'undefined') {
  fail('keyframe tự dựng không giải mã được — bài kiểm trạng thái vô nghĩa, sửa bài trước');
} else {
  const banTay = new SnapshotDecoder();
  for (let i = 0; i < 200; i++) {
    const n = 1 + randInt(150);
    const rac = new Uint8Array(n);
    for (let j = 0; j < n; j++) {
      rac[j] = randInt(256);
    }
    rac[0] = SNAP_MAGIC; // ép vào đúng đường CÓ NHỚ trạng thái
    decodeFrame(rac, banTay);
  }
  kiemTra++;
  const sauRac = JSON.stringify(decodeFrame(keyframe, banTay).msg);
  if (sauRac !== mongDoi) {
    fail('sau 200 gói rác, keyframe HỢP LỆ giải mã ra khác ban đầu — trạng thái đã lệch');
    console.error(`    sạch: ${mongDoi.slice(0, 200)}`);
    console.error(`    sau : ${sauRac.slice(0, 200)}`);
  } else {
    console.log('Trạng thái — sau 200 gói rác, keyframe hợp lệ vẫn giải mã ĐÚNG như cũ');
  }
}

if (loi > 0) {
  console.error(`\nTRƯỢT: ${loi}/${kiemTra} phép kiểm.`);
  process.exit(1);
}
console.log(`\nĐẠT: ${kiemTra} phép kiểm, không gói hỏng nào thoát ra ngoài.`);
