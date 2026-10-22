package game

import (
	"math"

	"kiemxoay/server/internal/geom"
)

// MoveSpeed implements spec §4.1 without the zone factor.
// Tốc độ KHÔNG phụ thuộc tu vi; chỉ buff Thần Hành (nhặt được) mới tăng tốc.
func (g *Game) MoveSpeed(p *Player) float64 {
	s := g.Cfg.Player.MoveBase
	if p.HasBuff(BuffHaste, g.Time) {
		s *= g.Cfg.Buffs["haste"].Mult
	}
	return s
}

// ZoneMult returns the terrain speed factor for moving along dir at pos (spec §8.3).
func (g *Game) ZoneMult(pos Vec, dir Vec) float64 {
	mult := 1.0
	g.QueryZones(pos, 1, func(id int) {
		z := g.Map.Zones[id]
		if !z.Shape.Contains(pos) {
			return
		}
		switch z.Kind {
		case "mud":
			mult *= g.Cfg.Zones.MudMult
		case "wind":
			mult *= g.windMult(dir, z.WindDir)
		default:
		}
	})
	return mult
}

// windMult: within wind_with_max_deg of the wind → with; beyond wind_against_min_deg → against;
// otherwise (crosswind) unchanged (spec §8.3).
func (g *Game) windMult(dir Vec, windDir float64) float64 {
	cos := geom.Clamp(dir.Dot(geom.FromAngle(windDir)), -1, 1)
	deg := math.Acos(cos) * 180 / math.Pi
	switch {
	case deg < g.Cfg.Zones.WindWithMaxDeg:
		return g.Cfg.Zones.WindWith
	case deg > g.Cfg.Zones.WindAgainstMinDeg:
		return g.Cfg.Zones.WindAgainst
	default:
		return 1
	}
}

// BushAt returns the id of the bush zone containing pos, or -1.
func (g *Game) BushAt(pos Vec) int {
	found := -1
	g.QueryZones(pos, 1, func(id int) {
		z := g.Map.Zones[id]
		if z.Kind == "bush" && z.Shape.Contains(pos) {
			found = id
		}
	})
	return found
}

func (g *Game) movePlayers(dt float64) {
	for _, p := range g.PlayerList() {
		if p.Dead || p.InTransit(g.Time) && !p.Transit.Arrived {
			continue
		}
		g.movePlayer(p, dt)
		p.Bush = g.BushAt(p.Pos)
		g.checkGate(p)
	}
}

func (g *Game) movePlayer(p *Player, dt float64) {
	if p.Moving {
		dir := geom.FromAngle(p.InputAngle)
		p.Pos = p.Pos.Add(dir.Mul(g.MoveSpeed(p) * g.ZoneMult(p.Pos, dir) * dt))
		p.Facing = p.InputAngle
	}
	if p.KnockUntil > g.Time {
		p.Pos = p.Pos.Add(p.KnockVel.Mul(dt))
	}
	p.Pos = g.resolveBlocking(p.Pos, g.BodyR(p))
}

// resolveBlocking pushes a circle out of solid obstacles and keeps it inside the map.
func (g *Game) resolveBlocking(pos Vec, r float64) Vec {
	for iter := 0; iter < 3; iter++ {
		moved := false
		g.QueryObstacles(pos, r, func(id int) {
			if !g.ObstacleSolid(id) {
				return
			}
			if np, hit := g.Map.Obstacles[id].Block().PushOut(pos, r); hit {
				pos, moved = np, true
			}
		})
		if !moved {
			break
		}
	}
	size := g.Map.Size
	return Vec{X: geom.Clamp(pos.X, r, size-r), Y: geom.Clamp(pos.Y, r, size-r)}
}

// circleFree reports whether a circle overlaps no solid obstacle.
func (g *Game) circleFree(pos Vec, r float64) bool {
	free := true
	g.QueryObstacles(pos, r, func(id int) {
		if g.ObstacleSolid(id) && g.Map.Obstacles[id].Block().OverlapsCircle(pos, r) {
			free = false
		}
	})
	return free
}

// randomFreePoint returns a random point not inside obstacles, preferring far from players.
func (g *Game) randomFreePoint(r, minPlayerDist float64, tries int) Vec {
	best, bestD := Vec{X: g.Map.Size / 2, Y: g.Map.Size / 2}, -1.0
	margin := r + 50
	for i := 0; i < tries; i++ {
		p := Vec{X: margin + g.Rng.Float64()*(g.Map.Size-2*margin), Y: margin + g.Rng.Float64()*(g.Map.Size-2*margin)}
		if !g.circleFree(p, r) {
			continue
		}
		d := g.nearestPlayerDist(p)
		if d >= minPlayerDist {
			return p
		}
		if d > bestD {
			best, bestD = p, d
		}
	}
	return best
}

func (g *Game) nearestPlayerDist(p Vec) float64 {
	best := math.Inf(1)
	for _, o := range g.Players {
		if !o.Dead {
			best = math.Min(best, o.Pos.Dist(p))
		}
	}
	return best
}

func (g *Game) safeSpawn() Vec {
	return g.randomFreePoint(g.Cfg.Player.BodyR*2, g.Cfg.Room.SafeSpawnR, 60)
}

func (g *Game) moveMonsters(dt float64) {
	mc := g.Cfg.Monsters
	for _, m := range g.Monsters {
		speed := mc.WanderSpeed
		if m.FleeUntil > g.Time {
			m.Heading = m.Pos.Sub(m.FleeFrom).Angle()
			speed = mc.FleeSpeed
		} else if g.Time >= m.TurnAt {
			m.Heading = g.Rng.Float64() * 2 * math.Pi
			m.TurnAt = g.Time + 2 + g.Rng.Float64()*3
		}
		next := g.resolveBlocking(m.Pos.Add(geom.FromAngle(m.Heading).Mul(speed*dt)), m.R)
		if next.Dist2(m.Pos) < 1e-6 {
			m.TurnAt = g.Time // stuck against a wall: pick a new heading next tick
		}
		m.Pos = next
	}
}
