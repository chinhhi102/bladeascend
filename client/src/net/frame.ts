import { decode } from '@msgpack/msgpack';

// Đuôi `.ts` tường minh: `scripts/decode-fuzz.ts` nạp tệp này thẳng bằng Node (không qua
// Vite), mà Node ESM bắt buộc có đuôi. `allowImportingTsExtensions` trong tsconfig bật sẵn
// cho đúng kiểu dùng này — xem ghi chú ở đó.
import type { ServerMsg } from './protocol.ts';
import { decodeLeaderboard, isBinaryLeaderboard } from './lbwire.ts';
import { isBinarySnapshot, type SnapshotDecoder } from './snapwire.ts';

/**
 * Giải mã MỘT khung từ WebSocket — tách khỏi `Conn` để KIỂM ĐƯỢC.
 *
 * Trước 1/11 đoạn này nằm thẳng trong `ws.onmessage`, nên muốn thử "gói hỏng thì sao" là
 * phải dựng một WebSocket giả. Không ai dựng, và hệ quả là đường msgpack đi suốt nhiều
 * tháng mà không có `try/catch`: một gói hỏng ném RangeError thẳng ra khỏi `onmessage`
 * (thấy trong telemetry 31/10, Safari).
 *
 * Hàm này KHÔNG BAO GIỜ ném. Gói hỏng trả về `fail`; người gọi bỏ qua gói đó và đi tiếp.
 */
export interface FrameFail {
  kind: 'snapshot_v2' | 'leaderboard_v2' | 'msgpack';
  /** Byte đầu — chính là `MsgType` với gói nhị phân, hoặc byte mở đầu msgpack. */
  byte0: number;
  len: number;
  /** 8 byte đầu dạng hex, đủ để dựng lại gói lúc gỡ lỗi. */
  head: string;
  err: string;
}

export interface FrameResult {
  msg?: ServerMsg;
  fail?: FrameFail;
}

function describe(kind: FrameFail['kind'], u8: Uint8Array, err: unknown): FrameFail {
  return {
    kind,
    byte0: u8.length > 0 ? u8[0] : -1,
    len: u8.length,
    head: [...u8.subarray(0, 8)].map((b) => b.toString(16).padStart(2, '0')).join(' '),
    err: String(err).slice(0, 160),
  };
}

export function decodeFrame(u8: Uint8Array, snap: SnapshotDecoder): FrameResult {
  if (isBinarySnapshot(u8)) {
    try {
      return { msg: snap.decode(u8) };
    } catch (err) {
      return { fail: describe('snapshot_v2', u8, err) };
    }
  }
  if (isBinaryLeaderboard(u8)) {
    try {
      return { msg: decodeLeaderboard(u8) };
    } catch (err) {
      return { fail: describe('leaderboard_v2', u8, err) };
    }
  }
  // Mọi message KHÁC snapshot (welcome, bảng xếp hạng, kill feed, kick…) đi msgpack.
  //
  // Bỏ qua một gói ở đây AN TOÀN: chúng là tin rời, mất một cái thì thiếu một dòng chữ
  // chứ không lệch trạng thái trận. Snapshot — thứ mà mất gói sẽ lệch — đi đường nhị phân
  // và có keyframe 2 giây một lần dựng lại toàn bộ.
  let v: unknown;
  try {
    v = decode(u8);
  } catch (err) {
    return { fail: describe('msgpack', u8, err) };
  }
  // Giải mã ĐƯỢC chưa có nghĩa là một message.
  //
  // msgpack nhận rất nhiều thứ: một byte `0x00` là số nguyên 0 hợp lệ, `0x90` là mảng
  // rỗng hợp lệ. Bài fuzz 1/11 cho thấy 3.754 trong 12.000 gói RÁC ngẫu nhiên vẫn "giải
  // mã thành công" như vậy. Ép thẳng chúng thành `ServerMsg` rồi đưa xuống `receive()` là
  // để một con số đi lang thang trong máy trạng thái — không nổ ngay, mà sai về sau ở chỗ
  // không ai lần ra.
  //
  // Mọi `ServerMsg` đều có trường `t` dạng chuỗi (snap, welcome, pong, kicked…). Đó là
  // cửa rẻ nhất và chặt nhất.
  if (typeof v !== 'object' || v === null || typeof (v as { t?: unknown }).t !== 'string') {
    return { fail: describe('msgpack', u8, `giải mã ra ${Array.isArray(v) ? 'mảng' : typeof v}, không phải message có trường "t"`) }; // i18n-ok: báo lỗi giải mã cho dev
  }
  return { msg: v as ServerMsg };
}
