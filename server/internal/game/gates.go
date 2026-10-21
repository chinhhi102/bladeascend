package game

import (
	"math"

	"kiemxoay/server/internal/geom"
)

// checkGate starts a teleport when a player's centre enters a gate (spec §8.4).
func (g *Game) checkGate(p *Player) {
	if p.Transit != nil || g.Time < p.GateReadyAt || len(g.Map.Gates) == 0 {
		return
	}
	gc := g.Cfg.Gates
	for _, gate := range g.Map.Gates {
		if p.Pos.Dist(Vec{X: gate.X, Y: gate.Y}) > gc.Radius {
			continue
		}
		dests := g.Map.GatesInNetwork(gate)
		if len(dests) == 0 {
			return
		}
		d := dests[g.Rng.Intn(len(dests))]
		half := gc.TransitSecs / 2
		p.Transit = &Transit{Dest: Vec{X: d.X, Y: d.Y}, DestGate: d.ID, OutAt: g.Time + half, DoneAt: g.Time + gc.TransitSecs}
		g.setMood(p, MoodTeleportOut, half)
		g.emit(Event{Kind: "TeleportOut", A: p.ID, B: uint32(d.ID), N: gate.ID, X: float32(gate.X), Y: float32(gate.Y)})
		return
	}
}

// updateTransit moves players through gates: warn flash, reappear, end immunity.
func (g *Game) updateTransit(p *Player) {
	t := p.Transit
	if t == nil {
		return
	}
	gc := g.Cfg.Gates
	if !t.WarnSent && g.Time >= t.OutAt-gc.ExitWarnSecs {
		t.WarnSent = true
		g.emit(Event{Kind: "GateWarn", N: t.DestGate, X: float32(t.Dest.X), Y: float32(t.Dest.Y)})
	}
	if !t.Arrived && g.Time >= t.OutAt {
		t.Arrived = true
		// Exit just outside the gate radius so the player does not re-enter immediately.
		off := geom.FromAngle(g.Rng.Float64() * 2 * math.Pi).Mul(gc.Radius + g.BodyR(p) + 10)
		p.Pos = g.resolveBlocking(t.Dest.Add(off), g.BodyR(p))
		p.lastTeleportIn = g.Time
		g.setMood(p, MoodTeleportIn, gc.TransitSecs/2)
		g.emit(Event{Kind: "TeleportIn", A: p.ID, N: t.DestGate, X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
	}
	if g.Time >= t.DoneAt {
		p.Transit = nil
		p.GateReadyAt = g.Time + gc.CooldownSecs
	}
}
