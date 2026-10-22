package game

// Lá chắn vòng kế ngoài cùng (spec §5.2 v4). The outermost ring is the grinding zone: enemy
// blades may enter it. Everything deeper is sealed — while that ring still has swords, nobody
// can push past `shield_r`. Break the outer ring and the shield falls back one layer; only
// when a single ring is left can a blade finally reach the body (§5.1).

// ShieldR is the radius nothing may cross: the blade tip of the SECOND-outermost ring.
//
// Returns 0 — no shield at all — once the player is down to a single ring: spec §5.2 says
// "chỉ khi người đó còn 1 vòng thì kiếm đối thủ mới chạm được tới thân", so the last ring must
// not keep attackers at arm's length or §5.1 could never fire.
func (g *Game) ShieldR(p *Player) float64 {
	geo := p.Geom
	if p.Ring.Count <= 0 || g.Cfg.Sword.ShieldRing == "" || geo.Rings < 2 {
		return 0
	}
	return geo.TipDist(geo.Rings-2, p.Ring.Tilt) + geo.Width/2
}

// minSeparation is how close `mover` may get to `other`: mover's blade tips may not cross
// other's shield, so dist ≥ reach(mover) + shield(other).
//
// Spec §5.2 writes this symmetrically as max(reachA + shield_B, reachB + shield_A), but that
// version makes the game unwinnable: a player with 2+ rings has a shield wider than their own
// reach, so it would hold every victim further away than the attacker can strike and NOBODY
// could ever land the §5.1 body hit. Applying it per-mover keeps the intent ("chỉ người đang
// đi bị chặn") and still seals the inner rings.
func (g *Game) minSeparation(mover, other *Player) float64 {
	if s := g.ShieldR(other); s > 0 {
		return mover.Geom.Reach + s
	}
	return 0
}

// blockedByShields pushes pos out of every other player's sealed zone. It only ever moves the
// player who is walking in — like hitting a wall, so they slide along it instead of bouncing
// back (spec §5.2).
func (g *Game) blockedByShields(p *Player, pos Vec) Vec {
	for iter := 0; iter < 3; iter++ {
		moved := false
		for _, o := range g.PlayerList() {
			if o == p || o.Dead || o.InTransit(g.Time) || p.InTransit(g.Time) {
				continue
			}
			need := g.minSeparation(p, o)
			d := pos.Sub(o.Pos)
			dist := d.Len()
			if dist >= need || need <= 0 {
				continue
			}
			away := d
			if dist < 1e-6 {
				away = Vec{X: 1} // trùng điểm: đẩy ra hướng bất kỳ
			} else {
				away = d.Mul(1 / dist)
			}
			pos = o.Pos.Add(away.Mul(need))
			moved = true
		}
		if !moved {
			break
		}
	}
	return pos
}
