package game

import "math"

// Tiên giới (spec §3.4): the five realms above Đại Thừa. They ignore EXP entirely — the only
// way up is sát mạng, one per successful body hit on another player. Because only Luyện Khí
// players actually die, counting kills alone would leave Tiên players with nothing to farm,
// so every landed hit counts.

// TienRealmOf is the player's Tiên realm index (0-based), or -1 while still mortal.
func (g *Game) TienRealmOf(p *Player) int {
	r := g.Prog.Realm(p.Level)
	if !g.Prog.IsTien(r) {
		return -1
	}
	return g.Prog.TienIndex(r)
}

// CanEarnTienKills is true from Đại Thừa tầng 10 upwards; hits landed before that are not
// banked towards Chân Tiên.
func (g *Game) CanEarnTienKills(p *Player) bool { return p.Level >= g.Prog.TopMortalLevel() }

// addTienKill banks one landed hit towards the next Tiên realm and promotes when the
// threshold is crossed. onCooldown is the same anti-farm check Uy Danh uses: hitting the
// same victim again inside the cooldown banks nothing.
func (g *Game) addTienKill(a, b *Player, onCooldown bool) {
	if !g.CanEarnTienKills(a) || onCooldown || g.Prog.TienCount() == 0 {
		return
	}
	v := 1.0
	if b.IsBot {
		v = g.Cfg.Cultivation.TienBotKillValue
	}
	a.TienKills += v
	for {
		idx := g.TienRealmOf(a)
		next := g.Prog.TienNextAt(idx)
		if next <= 0 || a.TienKills < next {
			return
		}
		g.promoteTien(a, idx+1)
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

// dropTienRealm applies spec §5.1 to someone standing in Tiên giới: one realm down, and the
// sát mạng counter falls back to the entry threshold of the realm they land on, so the
// progress they had banked towards the next one is lost.
func (g *Game) dropTienRealm(p *Player) {
	idx := g.TienRealmOf(p)
	if idx < 0 {
		return
	}
	if idx == 0 {
		// Chân Tiên → back to Đại Thừa tầng 10, mortal again.
		p.Level = g.Prog.TopMortalLevel()
		p.TotalExp = g.Prog.MaxExp()
		p.TienKills = 0
	} else {
		p.Level = g.Prog.TienLevel(idx - 1)
		p.TienKills = g.Prog.TienEntry(idx - 1)
	}
	g.refreshDerived(p)
	g.setMood(p, MoodRealmDrop, g.Cfg.Mood.RealmDropSecs)
	g.emit(Event{Kind: "RealmDrop", A: p.ID, N: p.Level, X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
	g.dropExcessSwords(p)
	g.emit(Event{Kind: "KillFeed", Global: true, S: "realm_drop", B: p.ID, N: g.Prog.Realm(p.Level)})
}

// TienProgress is what the HUD bar shows for a Tiên player: kills banked and the threshold
// for the next realm (spec §3.4 "Tiên khí 15/30"). at == 0 means already at Đạo Tổ.
func (g *Game) TienProgress(p *Player) (have, at float64) {
	idx := g.TienRealmOf(p)
	if idx < 0 {
		return 0, 0
	}
	from := g.Prog.TienEntry(idx)
	next := g.Prog.TienNextAt(idx)
	if next <= 0 {
		return 0, 0
	}
	return math.Max(0, p.TienKills-from), next - from
}
