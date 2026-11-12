package room

import (
	"log"
	"time"
)

// Quét mục mồ côi mỗi 60 giây — GAME_SPEC §13.9 ("vệ sinh bộ nhớ: không gì được tăng mãi").
//
// ── Vì sao cần, khi mọi map đã tự dọn ────────────────────────────────────────────────────
//
// Đúng là phần lớn map theo người chơi đã tự dọn ở đúng chỗ: `aoiTracker.Step` xoá entity
// đã rời, `knownPlayers`/`knownMonsters` xoá trong vòng dựng snapshot, `lastWire` dựng lại
// mỗi keyframe (2 s), `resetKnown` xoá sạch khi đổi khu. Bản quét này KHÔNG thay chúng.
//
// Nó tồn tại vì một lý do khác: mỗi chỗ dọn kia nằm trên một đường chạy CÓ ĐIỀU KIỆN —
// "khi gửi snapshot", "khi entity còn trong danh sách", "khi đổi khu". Thêm một nhánh
// thoát sớm (client nghẽn, khu suy giảm §13.8 bỏ tick, phòng đóng giữa chừng) là một mục
// nằm lại mãi, và không có gì báo. Quét định kỳ là lưới an toàn: nó chạy vô điều kiện, và
// nó ĐẾM những gì nó dọn được — nên nếu lưới bắt được cái gì thì ta biết có chỗ đang rò.
//
// Nguyên tắc: quét không bao giờ được sửa trạng thái TRẬN. Nó chỉ xoá những thứ đã mất
// chủ. Một bản quét "sửa" được thì nó là một luật chơi thứ hai chạy song song.

// sweepEverySecs: nhịp quét (§13.9 chốt 60 giây).
const sweepEverySecs = 60.0

// sweepStats đếm những gì lượt quét vừa dọn. Khác 0 là có chỗ đang rò.
type sweepStats struct {
	Names    int // tên bot hết hạn nhớ
	Holds    int // chỗ giữ §7.1 mất nhân vật
	Desc     int // yêu cầu gửi lại mô tả cho entity không còn
	Items    int // ô `knownItems` trỏ vào viên đã biến mất
	Withdraw int
}

func (s sweepStats) empty() bool {
	return s.Names == 0 && s.Holds == 0 && s.Desc == 0 && s.Items == 0 && s.Withdraw == 0
}

// sweepTick đếm giờ rồi gọi `sweepOrphans` mỗi 60 giây. Gọi từ `Step` (ranh giới tick).
func (r *Room) sweepTick(dt float64) {
	r.sweepAcc += dt
	if r.sweepAcc < sweepEverySecs {
		return
	}
	r.sweepAcc = 0
	if st := r.sweepOrphans(); !st.empty() {
		// Lưới bắt được cái gì nghĩa là một đường dọn chính đã hụt. Ghi ra để lần sau có
		// người đi tìm chỗ hụt đó, thay vì để lưới âm thầm che mãi.
		log.Printf("quét mồ côi phòng %s: tên %d · chỗ giữ %d · mô tả %d · viên %d · rút bot %d",
			r.ID, st.Names, st.Holds, st.Desc, st.Items, st.Withdraw)
	}
}

// sweepOrphans dọn mọi mục đã mất chủ. Chỉ gọi TRONG goroutine phòng.
func (r *Room) sweepOrphans() sweepStats {
	var st sweepStats
	now := r.now()

	// 1. Tên bot vừa rời: `rememberName` chỉ GHI, không bao giờ xoá. Trần tự nhiên là số
	//    tên trong config (~1 000), nhưng "có trần tự nhiên" không phải là "có trần".
	for name, until := range r.recentNames {
		if until <= now {
			delete(r.recentNames, name)
			st.Names++
		}
	}

	// 2. Chỗ giữ §7.1 mà nhân vật đã biến mất khỏi thế giới (và không phải vì chết — chết
	//    thì `dead` khác nil và chỗ giữ còn dùng để trả màn Tổng kết).
	for _, z := range r.zones {
		for id, h := range z.holds {
			if h.dead != nil {
				continue
			}
			if _, ok := z.G.Players[id]; !ok {
				r.dropHold(z, id, h)
				st.Holds++
			}
		}
	}

	// 3. Theo từng kết nối: yêu cầu gửi lại mô tả cho entity không còn trong khu, và ô
	//    `knownItems` trỏ vào viên đã biến mất.
	for _, z := range r.zones {
		for _, c := range z.clients {
			st.Desc += c.sweepDescAsks(z)
			st.Items += c.sweepKnownItems(z)
		}
	}

	// 4. Danh sách bot đang rút: bot có thể chết trước khi rút xong.
	for id := range r.withdraw {
		if r.zoneOf(id) == nil {
			delete(r.withdraw, id)
			st.Withdraw++
		}
	}
	return st
}

// now là đồng hồ game của khu đầu tiên — cùng thang với `recentNames`.
func (r *Room) now() float64 {
	if len(r.zones) == 0 {
		return 0
	}
	return r.zones[0].G.Time
}

// sweepDescAsks bỏ yêu cầu gửi lại mô tả của entity không còn trong khu.
//
// `takeDescAsks` vét sạch mỗi lần dựng snapshot, nên bình thường bảng này rỗng. Nhưng
// client nghẽn thì snapshot bị bỏ (SendLossy), và một client cố tình bắn `dreq` liên tục
// cho những id không tồn tại thì bảng nằm lại cho tới lần dựng snapshot kế.
func (c *Client) sweepDescAsks(z *Zone) int {
	c.inMu.Lock()
	defer c.inMu.Unlock()
	n := 0
	for id := range c.descAsk {
		if _, ok := z.G.Players[id]; ok {
			continue
		}
		if _, ok := z.G.Monsters[id]; ok {
			continue
		}
		delete(c.descAsk, id)
		n++
	}
	return n
}

// sweepKnownItems xoá những ô nhớ "đã gửi viên này" trỏ vào viên không còn trên đất.
//
// `knownItems` là slice theo CHỈ SỐ Ô, và ô được dùng lại — nên nó không phình. Thứ phình
// được là `knownSlots`, danh sách gọn các ô đang khác 0: mỗi ô vào đó một lần và chỉ ra
// khi vòng `diffItems` chạy. Vòng đó không chạy cho client đang nghẽn.
func (c *Client) sweepKnownItems(z *Zone) int {
	n := 0
	keep := c.knownSlots[:0]
	for _, slot := range c.knownSlots {
		if int(slot) < len(c.knownItems) && c.knownItems[slot] != 0 {
			if it := z.G.ItemBySlot(int(slot)); it != nil && it.ID == c.knownItems[slot] {
				keep = append(keep, slot)
				continue
			}
		}
		if int(slot) < len(c.knownItems) {
			c.knownItems[slot] = 0
		}
		n++
	}
	c.knownSlots = keep
	return n
}

// SweepNow chạy một lượt quét ngay — chỉ dùng cho test và cho bài kiểm rò rỉ.
func (r *Room) SweepNow() {
	r.sweepOrphans()
}

// SweepEvery là nhịp quét, cho test đọc mà không chép tay con số.
var SweepEvery = time.Duration(sweepEverySecs) * time.Second

// Reserved là số suất người phòng đang giữ (người đang chơi + người đang chờ nối lại
// §7.1). Chỉ để quan sát: `cmd/leakcheck` và trang quản trị đọc.
func (r *Room) Reserved() int { return r.mgr.humansOf(r) }

// RoomSlotAudit: số ô còn khác nil trong mảng nền của danh sách phòng.
//
// Bài test §13.9 đọc: `append(rooms[:i], rooms[i+1:]...)` hạ `len` nhưng để lại con trỏ ở
// ô cuối, và một `*Room` bỏ quên là cả một thế giới ≈ 13 MB không bao giờ được thu hồi.
func (m *Manager) RoomSlotAudit() (live, stale int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	live = len(m.rooms)
	full := m.rooms[:cap(m.rooms)]
	for i := live; i < len(full); i++ {
		if full[i] != nil {
			stale++
		}
	}
	return live, stale
}
