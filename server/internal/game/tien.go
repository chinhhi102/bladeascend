package game

import "math"

// Tiên giới (spec §3.4 v4): the five realms above Đại Thừa. They ignore Linh khí (EXP)
// entirely — the only way up is **Tiên khí**, a special orb that only the Đài Tiên Khí on
// the map spray out (§8.5). The "sát mạng" rule of the previous draft is gone.

// TienRealmOf is the player's Tiên realm index (0-based), or -1 while still mortal.
//
// Tính thẳng từ **tổng Tiên khí** (spec v5.7 §3.4), không đọc `Level`. `Level` vẫn được giữ
// đồng bộ để phần còn lại của game (scale, số kiếm tối đa, hào quang) dùng, nhưng nguồn sự
// thật chỉ có một — nếu không thì hai con số sẽ lệch nhau đúng lúc bị chém.
func (g *Game) TienRealmOf(p *Player) int {
	if !g.BanksTienKhi(p) {
		return -1
	}
	return g.Prog.TienRealmForKhi(p.TienKhi)
}

// BanksTienKhi: người này đã sang đường tu Tiên khí chưa.
//
// Mốc là **Đại Thừa tầng 10 ĐẦY THANH** (spec v5.7 §3.4), không phải chỉ cần đứng ở Đại
// Thừa tầng 10. Khác biệt thật: trước đây vừa chạm tầng 10 là Linh khí nhặt được nữa cũng
// bị bỏ phí trong khi Tiên khí thì hiếm, nên người chơi kẹt cứng ở đó. Giờ họ nạp nốt thanh
// Đại Thừa bằng Linh khí như thường, đầy rồi mới chuyển sang săn Tiên khí.
//
// Đây cũng là ranh giới của HAI ĐƯỜNG TU TÁCH HẲN (anh chốt 2026-10-28, xem Q18): ai đã
// sang Tiên khí thì không hấp thụ Linh khí nữa, và ngược lại.
func (g *Game) BanksTienKhi(p *Player) bool {
	if !g.Cfg.Cultivation.TienEntryRequiresFullExp {
		return p.Level >= g.Prog.TopMortalLevel()
	}
	return p.TotalExp >= g.Prog.MaxExp()-1e-9
}

// CanAbsorbTienKhi / CanAbsorbExp: mỗi người chỉ hấp thụ đúng một loại tu vi.
func (g *Game) CanAbsorbTienKhi(p *Player) bool { return g.BanksTienKhi(p) }
func (g *Game) CanAbsorbExp(p *Player) bool     { return !g.BanksTienKhi(p) }

// GainTienKhi banks Tiên khí and promotes when a threshold is crossed. Returns how much
// actually went into the Tiên khí bar — 0 for anyone still on the Linh khí path.
func (g *Game) GainTienKhi(p *Player, points float64) float64 {
	if points <= 0 || p.Dead || !g.CanAbsorbTienKhi(p) {
		return 0
	}
	before := g.TienRealmOf(p)
	p.TienKhi += points
	g.overflowToScore(p)
	// Cảnh giới suy ra từ tổng Tiên khí, nên chỉ cần tính lại một lần — không còn vòng lặp
	// "cộng dồn rồi thăng từng bậc". Nhặt một cục 400 điểm thì nhảy thẳng mấy cảnh giới.
	// v6.46: cảnh giới VÀ tầng đều suy từ tổng Tiên khí. Lên tầng thì chỉ đồng bộ `Level`
	// (glow nhỏ do client tự vẽ khi `L` đổi); lên CẢNH GIỚI mới có Thiên Kiếp.
	if after := g.TienRealmOf(p); after > before {
		g.promoteTien(p, after)
	} else if after >= 0 {
		g.syncTienLevel(p)
	}
	return points
}

// overflowToScore xử lý phần Tiên khí vượt trần Đạo Tổ (1 500).
//
// Luật v5.7: Đạo Tổ **vẫn nạp thanh bình thường từ 1 001 lên 1 500**; chỉ phần vượt QUÁ
// 1 500 mới đổi Uy Danh (10:1). Bản v5.6 ghim thanh ở đúng mốc vào cảnh giới nên Đạo Tổ
// không bao giờ có phần dư để Thiêu Đốt — đó là Q26.
//
// Vẫn chặn thanh ở 1 500 chứ không cho phình mãi: `robOnKill` tính phần chia theo `tien_khi − đáy`,
// không chặn thì một người Đạo Tổ ngồi lâu thành cục mồi khổng lồ.
func (g *Game) overflowToScore(p *Player) {
	cap := g.Prog.TienMaxKhi()
	over := p.TienKhi - cap
	if over <= 0 {
		return
	}
	p.TienKhi = cap
	if rate := g.Cfg.Cultivation.TienOverflowScore; rate > 0 {
		g.addScore(p, "tien_overflow", over*rate)
	}
}

// syncTienLevel đồng bộ `Level` theo tổng Tiên khí (cảnh giới + TẦNG, v6.46) mà KHÔNG báo
// Thiên Kiếp — dùng cho lên tầng trong cùng một cảnh giới, và cho Thiêu Đốt rớt tầng.
func (g *Game) syncTienLevel(p *Player) {
	idx := g.TienRealmOf(p)
	if idx < 0 {
		return
	}
	want := g.Prog.TienLevel(idx, g.Prog.TienTierForKhi(p.TienKhi))
	if want == p.Level {
		return
	}
	p.Level = want
	p.PeakLevel = max(p.PeakLevel, p.Level)
	g.refreshDerived(p)
}

// promoteTien moves a player up to Tiên realm idx with the Thiên Kiếp announcement.
func (g *Game) promoteTien(p *Player, idx int) {
	p.Level = g.Prog.TienLevel(idx, g.Prog.TienTierForKhi(p.TienKhi))
	p.PeakLevel = max(p.PeakLevel, p.Level)
	g.refreshDerived(p)
	realm := g.Prog.Realm(p.Level)
	g.awardBreakthrough(p, realm)
	g.setMood(p, MoodBreakthrough, g.Cfg.Mood.BreakthroughSecs)
	g.emit(Event{Kind: "Breakthrough", A: p.ID, N: p.Level, S: "tien",
		X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
	g.emit(Event{Kind: "KillFeed", Global: true, S: "tien_ascend", A: p.ID, N: realm})
	g.shockNearby(p)
}

// TienProgress is what the HUD bar shows: Tiên khí banked inside the current realm and how
// wide that realm is (spec §3.4 "Tiên khí 150/300").
//
// Đạo Tổ KHÔNG còn trả (0, 0) nữa: từ v5.7 họ vẫn nạp thanh 1 001 → 1 500, và chính thanh
// đó quyết định lúc nào Tiên khí dư mới đổi sang Uy Danh — giấu thanh đi thì người chơi
// không biết còn bao xa mới tới mốc đổi.
func (g *Game) TienProgress(p *Player) (have, at float64) {
	idx := g.TienRealmOf(p)
	if idx < 0 {
		return 0, 0
	}
	// v6.46: thanh là của TẦNG hiện tại, không phải của cả cảnh giới — giống phàm giới.
	// Thanh cả cảnh giới ở Đạo Tổ rộng 5 000 điểm, nhích một viên không thấy gì động đậy.
	s := g.Prog.TienTierForKhi(p.TienKhi)
	top := g.Prog.TienTierTop(idx, s)
	from := g.Prog.TienEntry(idx)
	if s > 0 {
		from = g.Prog.TienTierTop(idx, s-1) + 1
	}
	return math.Max(0, p.TienKhi-from+1), math.Max(1, top-from+1)
}
