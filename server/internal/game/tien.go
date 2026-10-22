package game

import "math"

// Tiên giới (spec §3.4 v4): the five realms above Đại Thừa. They ignore Linh khí (EXP)
// entirely — the only way up is **Tiên khí**, a special orb that only the Đài Tiên Khí on
// the map spray out (§8.5). The "sát mạng" rule of the previous draft is gone.

// TienRealmOf is the player's Tiên realm index (0-based), or -1 while still mortal.
func (g *Game) TienRealmOf(p *Player) int {
	r := g.Prog.Realm(p.Level)
	if !g.Prog.IsTien(r) {
		return -1
	}
	return g.Prog.TienIndex(r)
}

// BanksTienKhi is true from Đại Thừa tầng 10 upwards: those players put Tiên khí straight
// into the bar. Everyone below converts it to Linh khí instead.
func (g *Game) BanksTienKhi(p *Player) bool { return p.Level >= g.Prog.TopMortalLevel() }

// TienKhiAsExp is what one Tiên khí point is worth as Linh khí for a player who cannot bank
// it yet: base × 2^r (spec §3.4). That keeps the altars worth fighting over at every realm.
func (g *Game) TienKhiAsExp(p *Player, points float64) float64 {
	c := g.Cfg.Cultivation
	return points * c.TienKhiToExpBase * math.Pow(2, float64(g.Prog.Realm(p.Level)))
}

// GainTienKhi banks Tiên khí (or converts it to Linh khí) and promotes when a threshold is
// crossed. Returns how much actually went into the Tiên khí bar.
func (g *Game) GainTienKhi(p *Player, points float64) float64 {
	if points <= 0 || p.Dead {
		return 0
	}
	if !g.BanksTienKhi(p) {
		g.GainExp(p, g.TienKhiAsExp(p, points), false)
		return 0
	}
	p.TienKhi += points
	for {
		idx := g.TienRealmOf(p)
		next := g.Prog.TienNextAt(idx)
		if next <= 0 || p.TienKhi < next {
			return points
		}
		g.promoteTien(p, idx+1)
	}
}

// promoteTien moves a player up to Tiên realm idx with the Thiên Kiếp announcement.
func (g *Game) promoteTien(p *Player, idx int) {
	p.Level = g.Prog.TienLevel(idx)
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

// robTienKhi applies spec §3.4 when a Tiên giới player is cut: they drop one realm, their
// Tiên khí falls back to the new realm's entry threshold, the attacker robs tien_rob_share
// of the loss and tien_drop_share scatters as orbs around the victim.
func (g *Game) robTienKhi(attacker, victim *Player) {
	idx := g.TienRealmOf(victim)
	if idx < 0 {
		return
	}
	c := g.Cfg.Cultivation
	var entry float64
	if idx == 0 {
		// Chân Tiên → về Đại Thừa tầng 10, hết Tiên giới.
		victim.Level = g.Prog.TopMortalLevel()
		victim.TotalExp = g.Prog.MaxExp()
	} else {
		victim.Level = g.Prog.TienLevel(idx - 1)
		entry = g.Prog.TienEntry(idx - 1)
	}
	loss := math.Max(0, victim.TienKhi-entry)
	victim.TienKhi = entry
	g.refreshDerived(victim)
	g.setMood(victim, MoodRealmDrop, g.Cfg.Mood.RealmDropSecs)
	g.emit(Event{Kind: "RealmDrop", A: victim.ID, N: victim.Level,
		X: float32(victim.Pos.X), Y: float32(victim.Pos.Y)})
	g.dropExcessSwords(victim)
	g.emit(Event{Kind: "KillFeed", Global: true, S: "realm_drop", A: attacker.ID, B: victim.ID,
		N: g.Prog.Realm(victim.Level)})
	if loss <= 0 {
		return
	}
	g.GainTienKhi(attacker, loss*c.TienRobShare)
	g.scatterTienKhi(victim.Pos, loss*c.TienDropShare)
}

// TienProgress is what the HUD bar shows: Tiên khí banked towards the next realm and the
// amount that realm costs (spec §3.4 "Tiên khí 150/300"). at == 0 means already at Đạo Tổ.
func (g *Game) TienProgress(p *Player) (have, at float64) {
	idx := g.TienRealmOf(p)
	if idx < 0 {
		// Đại Thừa tầng 10 đã bắt đầu tích Tiên khí cho chặng Chân Tiên.
		if !g.BanksTienKhi(p) {
			return 0, 0
		}
		return p.TienKhi, g.Prog.TienNextAt(-1)
	}
	from := g.Prog.TienEntry(idx)
	next := g.Prog.TienNextAt(idx)
	if next <= 0 {
		return 0, 0
	}
	return math.Max(0, p.TienKhi-from), next - from
}
