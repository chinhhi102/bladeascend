package game

import "math"

// GainExp adds EXP (applying Song Bội Đan when boosted) and handles level-ups (spec §3).
func (g *Game) GainExp(p *Player, amount float64, boosted bool) {
	// Hai đường tu tách hẳn (§3.4, anh chốt 2026-10-28): từ Đại Thừa tầng 10 trở lên chỉ
	// hấp thụ Tiên khí, Linh khí không vào nữa.
	if amount <= 0 || p.Dead || !g.CanAbsorbExp(p) {
		return
	}
	if boosted && p.HasBuff(BuffExp2, g.Time) {
		amount *= g.Cfg.Buffs["exp2"].Mult
	}
	g.setTotalExp(p, math.Min(p.TotalExp+amount, g.Prog.MaxExp()))
}

// setTotalExp updates total_exp and emits LevelUp / Breakthrough / RealmDrop as needed.
func (g *Game) setTotalExp(p *Player, exp float64) {
	oldL := p.Level
	p.TotalExp = math.Max(0, exp)
	// Tiên giới không dùng EXP (spec §3.4): cấp bậc ở đó do sát mạng quyết định, nên EXP
	// nhặt thêm chỉ cộng vào con số chứ không được kéo người chơi tụt về Đại Thừa.
	if g.Prog.IsTien(g.Prog.Realm(p.Level)) {
		if p.TotalExp > p.PeakExp {
			p.PeakExp = p.TotalExp
		}
		return
	}
	p.Level = g.Prog.LevelFromExp(p.TotalExp)
	if p.TotalExp > p.PeakExp {
		p.PeakExp = p.TotalExp
	}
	p.PeakLevel = max(p.PeakLevel, p.Level)
	g.refreshDerived(p)
	if p.Level == oldL {
		return
	}
	oldR, newR := g.Prog.Realm(oldL), g.Prog.Realm(p.Level)
	pos := p.Pos
	switch {
	case newR > oldR:
		g.awardBreakthrough(p, newR)
		if newR == len(g.Cfg.Cultivation.Realms)-1 && p.DaiThuaSecs == 0 {
			p.DaiThuaSecs = g.Time - p.SpawnedAt // bảng Thần Tốc (spec §11.2)
		}
		g.setMood(p, MoodBreakthrough, g.Cfg.Mood.BreakthroughSecs)
		g.emit(Event{Kind: "Breakthrough", A: p.ID, N: p.Level, X: float32(pos.X), Y: float32(pos.Y)})
		g.emit(Event{Kind: "KillFeed", Global: true, S: "breakthrough", A: p.ID, N: newR})
		g.shockNearby(p)
	case p.Level > oldL:
		g.setMood(p, MoodLevelup, g.Cfg.Mood.LevelupSecs)
		g.emit(Event{Kind: "LevelUp", A: p.ID, N: p.Level, X: float32(pos.X), Y: float32(pos.Y)})
	case newR < oldR:
		g.setMood(p, MoodRealmDrop, g.Cfg.Mood.RealmDropSecs)
		g.emit(Event{Kind: "RealmDrop", A: p.ID, N: p.Level, X: float32(pos.X), Y: float32(pos.Y)})
		g.dropExcessSwords(p)
	default:
	}
}

// dropExcessSwords turns swords above the new realm's maximum into bundles (spec §5.1).
func (g *Game) dropExcessSwords(p *Player) {
	excess := p.Ring.Count - g.Prog.MaxSwords(g.Prog.Realm(p.Level))
	if excess <= 0 {
		return
	}
	p.Ring.Count -= excess
	g.refreshDerived(p)
	g.dropBundles(p.Pos, excess, p.Geom.Reach)
}

// Geom returns the player's current ring geometry.
func (g *Game) Geom(p *Player) RingGeom { return p.Geom }

// AddSwords handles picking up a sword item or bundle: +n swords up to the realm maximum;
// already at the maximum → overcap EXP instead (spec §6.2).
func (g *Game) AddSwords(p *Player, n int) {
	limit := g.Prog.MaxSwords(g.Prog.Realm(p.Level))
	if p.Ring.Count >= limit {
		g.GainExp(p, g.Cfg.Items.OvercapExp, true)
		return
	}
	p.Ring.Count = min(limit, p.Ring.Count+max(n, 1))
	g.refreshDerived(p)
}

// LoseSwords removes up to n swords, never going below the minimum. Returns how many were lost.
func (g *Game) LoseSwords(p *Player, n int) int {
	lost := loseUpTo(p.Ring.Count, n, g.Cfg.Sword.MinCount)
	p.Ring.Count -= lost
	g.refreshDerived(p)
	return lost
}

// shockNearby makes characters around a breakthrough show a surprised emote (spec §3.2).
func (g *Game) shockNearby(p *Player) {
	r := g.Cfg.Mood.NearR * p.Scale
	for _, o := range g.PlayerList() {
		if o != p && !o.Dead && o.Pos.Dist(p.Pos) <= r {
			g.emote(o, "shocked")
		}
	}
}
