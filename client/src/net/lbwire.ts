// Đuôi `.ts` tường minh: `scripts/decode-fuzz.ts` nạp tệp này bằng Node (xem frame.ts).
import type { RoomBoardMsg, RoomBoardRow } from './protocol.ts';
import { decodeLbHead, decodeLbRow, MsgType, Reader } from './wire.gen.ts';

/** LB_ZONE_HIDDEN: `zone_id` của người đang Ẩn Thân (proto/schema.txt, LbRow). */
export const LB_ZONE_HIDDEN = 255;

/** isBinaryLeaderboard: gói này là bảng xếp hạng phòng nhị phân (spec §11.5 v6.19)? */
export function isBinaryLeaderboard(b: Uint8Array): boolean {
  return b.length > 0 && b[0] === MsgType.LbHead;
}

/**
 * decodeLeaderboard đọc [LbHead][LbRow]… thành `RoomBoardMsg`. Gói hỏng thì NÉM — người gọi
 * (`decodeFrame`) bắt và bỏ gói, như với snapshot.
 */
export function decodeLeaderboard(u8: Uint8Array): RoomBoardMsg {
  const r = new Reader(u8);
  r.u8v();
  const head = decodeLbHead(r);
  const rows: RoomBoardRow[] = [];
  for (let i = 0; i < head.count; i++) {
    const t = r.u8v();
    if (t !== MsgType.LbRow) {
      throw new RangeError(`bảng xếp hạng: dòng ${i} mở đầu bằng loại ${t}`);
    }
    const w = decodeLbRow(r);
    rows.push({
      id: w.id, name: w.name, character: w.character, titleId: w.titleId, rank: w.rank,
      zoneId: w.zoneId, realm: w.realm, tier: w.tier, score: w.score, stealth: w.stealth === 1,
    });
  }
  return { t: 'rb', rank: head.rank, total: head.total, ms: head.myScore, rows };
}
