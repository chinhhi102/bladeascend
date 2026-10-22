package game

import "math"

// emoteCooldownSecs keeps emote bubbles from spamming; per emote type.
const emoteCooldownSecs = 3.0

// setMood forces an animation for at least secs (spec §12).
func (g *Game) setMood(p *Player, mood string, secs float64) {
	if p.Mood != mood {
		p.MoodSince = g.Time
	}
	p.Mood = mood
	p.MoodUntil = g.Time + math.Max(secs, g.Cfg.Mood.MinSecs)
}

// emote shows a speech bubble over the player, rate limited per kind.
func (g *Game) emote(p *Player, kind string) {
	if p.EmoteReady[kind] > g.Time {
		return
	}
	p.EmoteReady[kind] = g.Time + emoteCooldownSecs
	g.emit(Event{Kind: "Emote", A: p.ID, S: kind, X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}

// Power is the strength estimate shared by moods and bot AI (spec §10).
func (g *Game) Power(p *Player) float64 {
	return float64(p.Ring.Count) * float64(g.Prog.Realm(p.Level)+1) * p.Geom.Reach
}

// updateMoods derives locomotion moods once forced moods expire.
func (g *Game) updateMoods() {
	for _, p := range g.PlayerList() {
		if p.Dead || g.Time < p.MoodUntil {
			continue
		}
		mood := g.locomotionMood(p)
		// Hold every mood at least min_secs to avoid flicker.
		if mood != p.Mood && g.Time-p.MoodSince < g.Cfg.Mood.MinSecs {
			continue
		}
		if mood != p.Mood {
			p.Mood, p.MoodSince = mood, g.Time
		}
	}
}

func (g *Game) locomotionMood(p *Player) string {
	if !p.Moving {
		return MoodIdle
	}
	dir := Vec{X: math.Cos(p.InputAngle), Y: math.Sin(p.InputAngle)}
	nearR := g.Cfg.Mood.NearR * p.Scale
	myPow := g.Power(p)
	mood := MoodRun
	for _, o := range g.Players {
		if o == p || o.Dead || o.Pos.Dist(p.Pos) > nearR {
			continue
		}
		toward := o.Pos.Sub(p.Pos).Norm().Dot(dir)
		op := g.Power(o)
		if op > myPow*g.Cfg.Bots.FleeRatio && toward < 0 {
			return MoodFlee
		}
		if op < myPow*g.Cfg.Bots.HuntRatio && toward > 0.5 {
			mood = MoodChase
		}
	}
	return mood
}
