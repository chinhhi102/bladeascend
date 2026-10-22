package game

import (
	"math"
	"strconv"
)

// Score sources shown in the Tổng kết breakdown (spec §11.1).
const (
	ScoreHit          = "hit"
	ScoreKill         = "kill"
	ScoreBreakthrough = "breakthrough"
	ScoreTop1         = "top1"
)

// victimKey identifies a victim for the anti-farm cooldown: humans by guest_id so
// reconnecting does not reset it, bots by player id (spec §11.1).
func victimKey(b *Player) string {
	if b.GuestID != "" {
		return b.GuestID
	}
	return "#" + strconv.FormatUint(uint64(b.ID), 10)
}

// UpsetMult is lật_kèo = 1 + upset_per_realm × (r_victim − r_attacker), at least 1.
func (g *Game) UpsetMult(attackerRealm, victimRealm int) float64 {
	return math.Max(1, 1+g.Cfg.Score.UpsetPerRealm*float64(victimRealm-attackerRealm))
}

// creditScore adds points to the running total and the per-source breakdown, silently.
func (g *Game) creditScore(p *Player, source string, points float64) {
	if points <= 0 {
		return
	}
	p.Score += points
	if p.ScoreBy == nil {
		p.ScoreBy = map[string]float64{}
	}
	p.ScoreBy[source] += points
}

// addScore credits points and tells the player's own client to pop the number up.
func (g *Game) addScore(p *Player, source string, points float64) {
	if points <= 0 {
		return
	}
	g.creditScore(p, source, points)
	g.emit(Event{Kind: "Score", A: p.ID, To: p.ID, N: int(math.Round(points)), S: source,
		X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}

// awardHit gives the attacker Uy Danh for landing a body hit (spec §11.1). victimRealm is the
// victim's realm *before* the hit knocked them down. Hitting the same victim again within
// same_victim_cooldown_secs scores nothing, so camping one player cannot farm points.
func (g *Game) awardHit(a, b *Player, victimRealm int, killed bool) {
	a.Hits++
	sc := g.Cfg.Score
	if a.HitCooldown == nil {
		a.HitCooldown = map[string]float64{}
	}
	key := victimKey(b)
	onCooldown := a.HitCooldown[key] > g.Time
	a.HitCooldown[key] = g.Time + sc.SameVictimCooldown
	// Sát mạng dùng chung bộ chống cày với Uy Danh (spec §3.4).
	g.addTienKill(a, b, onCooldown)
	if onCooldown {
		return
	}
	points := sc.HitBase * float64(victimRealm+1) * g.UpsetMult(g.Prog.Realm(a.Level), victimRealm)
	source := ScoreHit
	if killed {
		points += sc.KillBonus
		source = ScoreKill
	}
	if b.IsBot {
		points *= sc.BotVictimMult
	}
	g.addScore(a, source, points)
}

// awardBreakthrough gives points the first time a player reaches each realm in one life.
func (g *Game) awardBreakthrough(p *Player, realm int) {
	if p.ScoredRealms == nil {
		p.ScoredRealms = map[int]bool{}
	}
	if p.ScoredRealms[realm] {
		return
	}
	p.ScoredRealms[realm] = true
	g.addScore(p, ScoreBreakthrough, g.Cfg.Score.BreakthroughPerRealm*float64(realm))
}

// updateTop1 credits top1_per_sec to whoever currently leads the room and tracks how long
// they have held the lead (spec §11.1, stored as top1_secs).
func (g *Game) updateTop1(dt float64) {
	ranking := g.Ranking()
	if len(ranking) == 0 {
		return
	}
	// Credited silently: at tick_hz this would otherwise emit 30 pop-ups per second.
	lead := ranking[0]
	lead.Top1Secs += dt
	g.creditScore(lead, ScoreTop1, g.Cfg.Score.Top1PerSec*dt)
}

// ChestMarker returns where the minimap should draw the Rương Báu: either the live chest
// item, or the spot announced during the 10s warning (spec §12.1). warning is true while it
// has not spawned yet.
func (g *Game) ChestMarker() (pos Vec, warning, ok bool) {
	for _, it := range g.Items {
		if it.Kind == "chest" {
			return it.Pos, false, true
		}
	}
	if g.chestAt != nil {
		return *g.chestAt, true, true
	}
	return Vec{}, false, false
}

// updateTitles hands out the Top 1–3 room titles with an anti-flicker hold: a player must
// keep the new rank for title_hold_secs before the title actually changes (spec §11.3).
func (g *Game) updateTitles() {
	rank := map[uint32]int{}
	for i, p := range g.Ranking() {
		if i < len(g.Cfg.Titles) {
			rank[p.ID] = i + 1
		}
	}
	for _, p := range g.PlayerList() {
		want := rank[p.ID]
		if p.Dead {
			want = 0
		}
		if want == p.Title {
			p.pendingTitle, p.titleSince = want, 0
			continue
		}
		if p.pendingTitle != want || p.titleSince == 0 {
			p.pendingTitle, p.titleSince = want, g.Time
		}
		if g.Time-p.titleSince < g.Cfg.TitleHold {
			continue
		}
		old := p.Title
		p.Title, p.titleSince = want, 0
		switch {
		case want == 1:
			g.emit(Event{Kind: "KillFeed", Global: true, S: "title_gain", A: p.ID, N: want})
		case old == 1:
			g.emit(Event{Kind: "KillFeed", Global: true, S: "title_lose", A: p.ID, N: want})
		}
	}
}
