package room

import (
	"kiemxoay/server/internal/game"
	"kiemxoay/server/internal/netproto"
)

// Debug applies a test shortcut (M1 requirement); only reachable when the server runs with -debug.
func (r *Room) Debug(c *Client, d netproto.Debug) {
	r.Do(func() {
		p, ok := r.G.Players[c.PlayerID]
		if !ok || p.Dead {
			return
		}
		switch d.Cmd {
		case "exp":
			r.G.GainExp(p, float64(d.N), false)
		case "realm":
			// Jump to the start of the next realm.
			next := (r.G.Prog.Realm(p.Level) + 1) * r.G.Prog.Tiers()
			r.G.GainExp(p, r.G.Prog.Threshold(next)-p.TotalExp, false)
		case "sword":
			r.G.AddSwords(p, max(1, d.N))
		case "die":
			r.G.DebugKill(p)
		case "buff":
			if d.N >= 0 && d.N < game.BuffCount {
				r.G.ApplyBuff(p, game.BuffNames[d.N])
			}
		default:
		}
	})
}
