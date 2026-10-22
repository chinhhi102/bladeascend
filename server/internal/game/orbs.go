package game

import "math"

// Ambient EXP orbs cover the whole map (spec §6.1): one orb per exp_orbs.cell square. A picked
// orb grows back respawn_secs later at a random spot inside the same respawn_region square,
// so a farmed-out area fills up again instead of the orbs drifting elsewhere.

type pendingOrb struct {
	at     float64
	region int
}

func (g *Game) orbRegions() int {
	return max(1, int(math.Ceil(g.Map.Size/math.Max(g.Cfg.ExpOrbs.RespawnRegion, 1))))
}

// seedOrbs places the initial ambient orbs.
func (g *Game) seedOrbs() {
	oc := g.Cfg.ExpOrbs
	if oc.Cell <= 0 || len(oc.Values) == 0 {
		return
	}
	n := int(g.Map.Size / oc.Cell)
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			lo := Vec{X: float64(x) * oc.Cell, Y: float64(y) * oc.Cell}
			if at, ok := g.freePointIn(lo, oc.Cell); ok {
				g.addAmbientOrb(at)
			}
		}
	}
}

// respawnOrbs regrows orbs whose timer is due.
func (g *Game) respawnOrbs() {
	due := 0
	for due < len(g.orbRespawns) && g.orbRespawns[due].at <= g.Time {
		due++
	}
	size := g.Cfg.ExpOrbs.RespawnRegion
	n := g.orbRegions()
	for _, o := range g.orbRespawns[:due] {
		lo := Vec{X: float64(o.region%n) * size, Y: float64(o.region/n) * size}
		if at, ok := g.freePointIn(lo, size); ok {
			g.addAmbientOrb(at)
		} else {
			g.orbRespawns = append(g.orbRespawns, pendingOrb{at: g.Time + g.Cfg.ExpOrbs.RespawnSecs, region: o.region})
		}
	}
	g.orbRespawns = g.orbRespawns[due:]
}

func (g *Game) addAmbientOrb(at Vec) {
	oc := g.Cfg.ExpOrbs
	v := oc.Values[g.pickIndex(oc.Weights, len(oc.Values))]
	it := &Item{ID: g.newID(), Kind: "orb", Pos: at, Value: v}
	size := math.Max(oc.RespawnRegion, 1)
	n := g.orbRegions()
	it.Region = min(int(at.Y/size), n-1)*n + min(int(at.X/size), n-1)
	g.Items[it.ID] = it
	g.itemGrid.add(it)
}

// pickIndex chooses an index by weights (uniform when weights are missing).
func (g *Game) pickIndex(weights []int, n int) int {
	total := 0
	for i := 0; i < min(len(weights), n); i++ {
		total += weights[i]
	}
	if total <= 0 {
		return g.Rng.Intn(n)
	}
	roll := g.Rng.Intn(total)
	for i := 0; i < n; i++ {
		roll -= weights[i]
		if roll < 0 {
			return i
		}
	}
	return n - 1
}

// freePointIn returns a random point inside the square [lo, lo+size] that is clear of obstacles.
func (g *Game) freePointIn(lo Vec, size float64) (Vec, bool) {
	r := g.Cfg.Items.OrbR
	for try := 0; try < 4; try++ {
		p := Vec{X: lo.X + g.Rng.Float64()*size, Y: lo.Y + g.Rng.Float64()*size}
		if p.X < r || p.Y < r || p.X > g.Map.Size-r || p.Y > g.Map.Size-r {
			continue
		}
		if g.circleFree(p, r) {
			return p, true
		}
	}
	return Vec{}, false
}
