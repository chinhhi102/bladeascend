package room

import (
	"math"

	"bladeascend/server/internal/game"
	"bladeascend/server/internal/wire"
)

// lbZoneHidden là `zone_id` của người đang Ẩn Thân (proto/schema.txt, LbRow).
const lbZoneHidden = 255

// lbRows mã hoá top 10 phòng thành dãy [TypeLbRow][LbRow]… — MỘT lần mỗi giây cho cả phòng.
//
// Phần này giống hệt nhau với mọi người nhận, chỉ phần đầu (hạng của mình, Uy Danh của mình)
// là riêng. Mã hoá một lần rồi ghép sau từng đầu gói thay vì mã hoá lại cho mỗi client: một
// phòng 5 khu đầy người là vài trăm lần mã hoá thừa mỗi giây.
func (r *Room) lbRows(top []*game.Player) []byte {
	zoneOf := make(map[uint32]int, len(top))
	stealthy := make(map[uint32]bool, len(top))
	for _, z := range r.zones {
		for _, p := range top {
			if _, ok := z.G.Players[p.ID]; ok {
				zoneOf[p.ID] = z.Index
				stealthy[p.ID] = p.HasBuff(game.BuffStealth, z.G.Time)
			}
		}
	}
	prog := r.zones[0].G.Prog
	buf := make([]byte, 0, 32*len(top))
	for i, p := range top {
		row := wire.LbRow{
			ID: p.ID, Name: p.Name, Character: charIndex(r.cfg.Characters, p.Char),
			TitleID: clampU8(p.Title), Rank: uint8(i + 1), ZoneID: clampU8(zoneOf[p.ID]),
			Realm: clampU8(prog.Realm(p.Level)), Tier: clampU8(prog.Tier(p.Level) + 1),
			// Điểm từ bản chụp đã giữ, CÙNG nhịp với thứ tự (heldboard.go). Uy Danh của chính
			// mình ở HUD đi `LbHead.MyScore` — vẫn tức thời.
			Score: roundU32(r.board.score(p.ID)),
		}
		if stealthy[p.ID] {
			row.ZoneID, row.Stealth = lbZoneHidden, 1
		}
		buf = append(buf, wire.TypeLbRow)
		buf = row.Encode(buf)
	}
	return buf
}

// heldTop trả những người trong thứ tự đã giữ (heldboard.go), đúng thứ tự đó.
func (r *Room) heldTop(ranking []*game.Player) []*game.Player {
	byID := make(map[uint32]*game.Player, len(ranking))
	for _, p := range ranking {
		byID[p.ID] = p
	}
	top := make([]*game.Player, 0, len(r.board.held))
	for _, id := range r.board.held {
		if p, ok := byID[id]; ok {
			top = append(top, p)
		}
	}
	return top
}

// shownRank: hạng người nhận thấy ở dòng ghim. Trong bảng → hạng trên bảng. Ngoài bảng →
// hạng thật, nhưng không nhỏ hơn "ngay sau bảng": người vừa vượt lên hạng 7 mà bảng chưa
// kịp nhận (đang giữ 3 s) mà ghim "▶ 7" dưới một bảng đã có hạng 7 thì trông như lỗi.
func (r *Room) shownRank(id uint32, shown int) int {
	if k := r.board.rank(id); k > 0 {
		return k
	}
	live := r.rankAt[id]
	if live == 0 {
		return 0
	}
	return max(live, shown+1)
}

// lbFrame ghép đầu gói riêng của một người nhận với phần dòng chung.
func lbFrame(head wire.LbHead, rows []byte) []byte {
	buf := make([]byte, 0, 16+len(rows))
	buf = append(buf, wire.TypeLbHead)
	buf = head.Encode(buf)
	return append(buf, rows...)
}

// roundU32 làm tròn Uy Danh về số nguyên không âm, chặn trần u32.
func roundU32(v float64) uint32 {
	if v <= 0 {
		return 0
	}
	return uint32(math.Min(math.Round(v), math.MaxUint32))
}
