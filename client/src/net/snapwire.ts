// Đuôi `.ts` tường minh: `scripts/decode-fuzz.ts` nạp cây mô-đun này thẳng bằng Node
// (không qua Vite), mà Node ESM bắt buộc có đuôi. Ba tệp dưới đây đều không import gì
// nữa nên chuỗi dừng ở đây. `allowImportingTsExtensions` bật sẵn trong tsconfig.
import { EVENT_KINDS, MONSTER_KINDS } from './eventkinds.ts';
import type { GameEvent, MonsterState, PlayerState, Snapshot, WireItemState } from './protocol.ts';
import {
  decodeDesc, decodeDespawn, decodeEvent, decodeItemAdd, decodeItemDel, decodeMonster,
  decodeSelfExtra, decodeSelfState, decodeSnapHead, decodeSpawn, decodeUpdate,
  MsgType, Reader,
} from './wire.gen.ts';

/**
 * Giải mã snapshot nhị phân — Protocol v2 (NETCODE_SPEC §4.1).
 *
 * Trả về ĐÚNG hình dạng `Snapshot` mà renderer đã dùng từ bản msgpack. Đó là chủ ý: đổi
 * cách mã hoá là một việc, đổi cách vẽ là việc khác, và làm cả hai cùng lúc thì lúc hình
 * ảnh sai sẽ không biết sai vì phần nào.
 *
 * Client giữ trạng thái ĐÃ NHẬN của từng entity, vì server chỉ gửi trường ĐÃ ĐỔI (§4.3).
 * Cứ 2 giây server gửi một gói keyframe chở lại toàn bộ — cần vì `SendLossy` bên server cố
 * ý vứt snapshot cũ khi mạng yếu, nên "client nhận đủ mọi gói" là điều không ai hứa.
 */

/** SNAP_MAGIC: byte đầu của gói nhị phân.
 *
 * Phân biệt được với msgpack mà không cần thêm byte nào: map msgpack luôn bắt đầu từ 0x80
 * trở lên (fixmap/map16/map32), còn mọi `MsgType` của ta đều < 0x80.
 */
export const SNAP_MAGIC = MsgType.SnapHead;

/** isBinarySnapshot: gói này là snapshot v2 hay là msgpack? */
export function isBinarySnapshot(b: Uint8Array): boolean {
  return b.length > 0 && b[0] === SNAP_MAGIC;
}

/** Bit của `changed` trong Update — phải khớp thứ tự dòng bitmask trong `proto/schema.txt`. */
const CH = {
  x: 1 << 0, y: 1 << 1, facing: 1 << 2, level: 1 << 3, scale: 1 << 4,
  swords: 1 << 5, ringPhase: 1 << 6, ringDir: 1 << 7, mood: 1 << 8,
  buffs: 1 << 9, flags: 1 << 10, tilt: 1 << 11, status: 1 << 12, crack: 1 << 13,
} as const;

const TAU = Math.PI * 2;
const deAngle8 = (v: number): number => (v / 256) * TAU;
const deAngle16 = (v: number): number => (v / 65536) * TAU;

/**
 * SnapshotDecoder giữ trạng thái giữa các gói.
 *
 * Phải là một đối tượng chứ không phải một hàm thuần: delta chỉ có nghĩa khi nhớ được bản
 * trước. Mỗi kết nối một bộ; nối lại là dựng bộ mới (xem `Conn.open`).
 */
export class SnapshotDecoder {
  /** known: trạng thái mới nhất của từng entity, dựng dần từ Spawn + các Update. */
  private readonly known = new Map<number, PlayerState>();
  /** desc: tên và nhân vật, giữ RIÊNG khỏi state vì chúng sống lâu hơn (§5.5). */
  private readonly desc = new Map<number, { name: string; char: number; title: number }>();
  /** myId: để biết `SelfState`/`SelfExtra` thuộc về ai. Gán từ gói `welcome`. */
  myId = 0;
  /** chars: `config.characters`, để đổi chỉ số nhân vật về tên. */
  chars: string[] = [];

  reset(): void {
    this.known.clear();
    this.desc.clear();
  }

  decode(buf: Uint8Array): Snapshot {
    const r = new Reader(buf);
    if (r.u8v() !== MsgType.SnapHead) {
      throw new Error('snapwire: gói không bắt đầu bằng SnapHead');
    }
    const head = decodeSnapHead(r);
    if (head.flags & 1) {
      // Keyframe: gói này chở lại TOÀN BỘ, nên mọi thứ nhớ từ trước đã hết giá trị.
      // Không xoá thì một entity đã rời đi trong lúc mất gói sẽ còn lại mãi.
      this.known.clear();
    }

    const snap: Snapshot = {
      t: 'snap', tk: head.tick, tm: head.timeMs / 1000, ack: head.ack,
      p: [], m: [], sv: head.srvUs / 1000,
    };
    const items: WireItemState[] = [];
    const removed: number[] = [];
    const despawn: number[] = [];
    const events: GameEvent[] = [];
    const monsters: MonsterState[] = [];

    while (r.left > 0) {
      const type = r.u8v();
      switch (type) {
        case MsgType.Spawn: {
          const s = decodeSpawn(r);
          this.desc.set(s.id, { name: s.name, char: s.char, title: s.title });
          this.known.set(s.id, {
            id: s.id, x: s.x, y: s.y, f: deAngle8(s.facing), L: s.level,
            n: s.swords, ph: deAngle16(s.ringPhase), d: s.ringDir,
            tl: s.tilt / 128, sc: s.scale * 0.02, md: s.mood,
            bf: s.buffs, fl: s.flags, st: s.status, ck: s.crack / 255,
            xp: 0,
          });
          break;
        }
        case MsgType.Desc: {
          const d = decodeDesc(r);
          this.desc.set(d.id, { name: d.name, char: d.char, title: d.title });
          break;
        }
        case MsgType.Update: {
          const u = decodeUpdate(r);
          const cur = this.known.get(u.id);
          if (!cur) {
            // Update cho entity chưa từng Spawn: bỏ qua trong IM LẶNG, không ném lỗi.
            //
            // Chuyện này xảy ra bình thường khi `SendLossy` vứt mất đúng gói Spawn. Ném lỗi
            // ở đây là làm đứng cả vòng vẽ vì một gói lẻ; bỏ qua thì keyframe 2 giây sau tự
            // chữa lành.
            break;
          }
          if (u.changed & CH.x) cur.x = u.x ?? 0;
          if (u.changed & CH.y) cur.y = u.y ?? 0;
          if (u.changed & CH.facing) cur.f = deAngle8(u.facing ?? 0);
          if (u.changed & CH.level) cur.L = u.level ?? 0;
          if (u.changed & CH.scale) cur.sc = (u.scale ?? 0) * 0.02;
          if (u.changed & CH.swords) cur.n = u.swords ?? 0;
          if (u.changed & CH.ringPhase) cur.ph = deAngle16(u.ringPhase ?? 0);
          if (u.changed & CH.ringDir) cur.d = u.ringDir ?? 0;
          if (u.changed & CH.mood) cur.md = u.mood ?? 0;
          if (u.changed & CH.buffs) cur.bf = u.buffs ?? 0;
          if (u.changed & CH.flags) cur.fl = u.flags ?? 0;
          if (u.changed & CH.tilt) cur.tl = (u.tilt ?? 0) / 128;
          if (u.changed & CH.status) cur.st = u.status ?? 0;
          if (u.changed & CH.crack) cur.ck = (u.crack ?? 0) / 255;
          break;
        }
        case MsgType.SelfState: {
          const s = decodeSelfState(r);
          const me = this.known.get(this.myId);
          if (me) {
            // Vị trí của CHÍNH MÌNH giữ 1/256 đơn vị, không lượng tử về u16 như người khác:
            // nó dùng để hiệu chỉnh dự đoán (GAME_SPEC §8.2a) và sai số ở đây hoá thành giật.
            me.x = s.x / 256;
            me.y = s.y / 256;
            me.vx = s.vx / 256;
            me.vy = s.vy / 256;
          }
          snap.ack = s.ackSeq;
          break;
        }
        case MsgType.SelfExtra: {
          const e = decodeSelfExtra(r);
          const me = this.known.get(this.myId);
          if (me) {
            if (e.has & 1) me.xp = e.exp ?? 0;
            if (e.has & 2) me.skc = e.charges ?? 0;
            if (e.has & 4) me.ti = e.title ?? 0;
            if (e.has & 8) me.th = e.tienHave ?? 0;
            if (e.has & 16) me.tn = e.tienNeed ?? 0;
          }
          break;
        }
        case MsgType.Monster: {
          const m = decodeMonster(r);
          // `k`/`mx`/`r` chỉ đến ở gói ĐẦU TIÊN thấy con quái này; những gói sau client giữ
          // lại bản cũ. Điền chuỗi rỗng / 0 cho đúng kiểu — `state.ts` đã biết bỏ qua chúng.
          const ms: MonsterState = { id: m.id, x: m.x, y: m.y, k: '', mx: 0, r: 0 };
          if (m.has & 1) ms.hp = m.hp ?? 0;
          if (m.has & 2) ms.k = MONSTER_KINDS[m.kind ?? 0] ?? '';
          if (m.has & 4) ms.mx = m.maxHp ?? 0;
          if (m.has & 8) ms.r = m.r ?? 0;
          monsters.push(ms);
          break;
        }
        case MsgType.MonsterDel:
          r.u32(); // id — hiện chưa dùng: quái biến mất khi không còn trong danh sách
          break;
        case MsgType.ItemAdd: {
          const it = decodeItemAdd(r);
          items.push({ id: it.id, x: it.x, y: it.y, k: it.kind, v: it.has & 1 ? (it.value ?? 0) : 0 });
          break;
        }
        case MsgType.ItemDel:
          removed.push(decodeItemDel(r).id);
          break;
        case MsgType.Despawn: {
          const d = decodeDespawn(r);
          despawn.push(d.id);
          this.known.delete(d.id);
          this.desc.delete(d.id);
          break;
        }
        case MsgType.Event: {
          const e = decodeEvent(r);
          const ev: GameEvent = {
            k: EVENT_KINDS[e.kind] ?? '', x: e.x, y: e.y,
          };
          if (e.has & 1) ev.a = e.a ?? 0;
          if (e.has & 2) ev.b = e.b ?? 0;
          if (e.has & 4) ev.n = e.n ?? 0;
          if (e.has & 8) ev.s = e.s ?? '';
          if (e.has & 16) ev.f = e.f ?? 0;
          events.push(ev);
          break;
        }
        default:
          // Loại lạ = server mới hơn client. Không đọc tiếp được vì không biết message này
          // dài bao nhiêu — dừng hẳn gói này thay vì đọc rác từ giữa chừng.
          r.p = buf.length;
          break;
      }
    }

    // Trả về MỌI entity đang biết, không chỉ những cái gói này chạm tới.
    //
    // Bản đầu tôi chỉ trả về entity "đã chạm" và nó SAI theo hai đường, cả hai đều im lặng:
    //
    //  1. `state.expire()` xoá entity vắng mặt liên tục quá `KEEP_GONE_SECS`. Một người
    //     đứng yên hoàn toàn thì không có Update nào — và sau vài giây họ BIẾN MẤT khỏi
    //     màn hình dù vẫn đứng ngay đó.
    //  2. Bộ nội suy cần một mẫu cho MỖI entity ở MỖI snapshot để nội suy giữa hai mốc.
    //     Thiếu mẫu thì nó ngoại suy, và nhân vật trôi.
    //
    // Delta tiết kiệm ở ĐƯỜNG TRUYỀN, không phải ở đường vẽ: dựng lại danh sách đầy đủ từ
    // trạng thái đã nhớ không tốn một byte mạng nào.
    for (const [id, st] of this.known) {
      const d = this.desc.get(id);
      snap.p!.push(d ? { ...st, nm: d.name, ch: this.chars[d.char] ?? '', ti: st.ti ?? d.title } : { ...st });
    }
    snap.m = monsters;
    if (items.length) snap.ia = items;
    if (removed.length) snap.ir = removed;
    if (despawn.length) snap.ds = despawn;
    if (events.length) snap.ev = events;
    return snap;
  }
}
