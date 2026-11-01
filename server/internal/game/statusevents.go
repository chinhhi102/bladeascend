package game

// Event trạng thái cho lớp hiển thị kỹ năng (spec v6.36 §6.5).
//
// Snapshot đã chở bit `st` của từng người, nhưng bit chỉ nói "đang bị" — không nói CÒN BAO
// LÂU, cũng không nói AI gây ra. Spec đòi vòng đếm ngược trên icon, giây còn lại ở HUD và
// mũi tên chỉ về người dùng kỹ năng, mà "client không tự đoán trạng thái" — nên server gửi:
//
//	StatusStart {A = người bị, B = người gây (0 = tự thân), N = mã, F = giây, X/Y = vị trí người gây}
//	StatusEnd   {A = người bị, N = mã, X/Y = vị trí người bị}
//
// Người xem nhận event khi thấy NGƯỜI BỊ (room.eventVisible); người bị và người gây luôn nhận.

// Mã chỉ để hiển thị, nối sau `StatusNames`: không có bit trên dây, không đổi luật nào.
const (
	// StPush: đang bị đẩy lùi (Loa Phường, sóng Rụt Cổ) — theo `KnockUntil`.
	StPush = StatusCount + iota
	// StSteal: Ông Ba đang nằm võng Trộm Kiếm — theo `StealUntil`.
	StSteal
	// StMinion: đàn em của Đại Ca còn bao lâu — theo `MinionUntil`.
	StMinion
	EventStatusCount
)

// EventStatusNames: tên theo mã, khớp `STATUS_EV` ở client (`client/src/net/protocol.ts`).
var EventStatusNames = func() [EventStatusCount]string {
	var out [EventStatusCount]string
	copy(out[:], StatusNames[:])
	out[StPush], out[StSteal], out[StMinion] = "push", "steal", "minion"
	return out
}()

// statusStart phát `StatusStart` và ghi nhớ là đã báo, để sau này báo `StatusEnd` đúng một lần.
// Gia hạn (trúng lại khi còn hiệu lực) phát lại với số giây mới — client đặt lại đồng hồ.
func (g *Game) statusStart(src, p *Player, code int, secs float64) {
	at := p.Pos
	var from uint32
	if src != nil && src.ID != p.ID {
		at, from = src.Pos, src.ID
	}
	p.statusShown |= 1 << code
	g.emit(Event{Kind: "StatusStart", A: p.ID, B: from, N: code, F: float32(secs),
		X: float32(at.X), Y: float32(at.Y)})
}

// statusActive: mã này còn hiệu lực không — đọc đúng nguồn sự thật của từng mã.
func (g *Game) statusActive(p *Player, code int) bool {
	switch code {
	case StPush:
		return p.KnockUntil > g.Time
	case StSteal:
		return p.StealUntil > g.Time
	case StMinion:
		return p.IsMinion && p.MinionUntil > g.Time
	default:
		return code < StatusCount && p.HasStatus(code, g.Time)
	}
}

// endShownStatuses báo `StatusEnd` cho mọi mã đã báo bắt đầu mà nay hết (hết giờ, bị xoá
// bởi Gồng/Gom Đơn, bị chém tỉnh ngủ…). `all` = báo hết bất kể còn giờ — dùng khi chết.
//
// Quét mỗi tick thay vì cài vào từng chỗ xoá trạng thái: trạng thái bị xoá ở cả chục chỗ
// (`p.Status[s] = 0`), quên một chỗ là icon treo mãi trên đầu người chơi.
func (g *Game) endShownStatuses(p *Player, all bool) {
	for code := 0; p.statusShown != 0 && code < EventStatusCount; code++ {
		bit := uint16(1) << code
		if p.statusShown&bit == 0 || (!all && g.statusActive(p, code)) {
			continue
		}
		p.statusShown &^= bit
		g.emit(Event{Kind: "StatusEnd", A: p.ID, N: code, X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
	}
}
