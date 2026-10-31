package game

import (
	"math"

	"bladeascend/server/internal/geom"
	"bladeascend/server/internal/world"
)

// staticIndex buckets obstacles and zones into a coarse grid (spec §13.4).
type staticIndex struct {
	cell  float64
	n     int
	obs   [][]int
	zones [][]int
	seen  []int
	stamp int
}

func newStaticIndex(m *world.Map, cell float64) *staticIndex {
	n := int(math.Ceil(m.Size/cell)) + 1
	s := &staticIndex{cell: cell, n: n, obs: make([][]int, n*n), zones: make([][]int, n*n)}
	for _, o := range m.Obstacles {
		s.insert(s.obs, o.ID, o.Shape)
	}
	for _, z := range m.Zones {
		s.insert(s.zones, z.ID, z.Shape)
	}
	s.seen = make([]int, max(len(m.Obstacles), len(m.Zones))+1)
	return s
}

func (s *staticIndex) insert(buckets [][]int, id int, sh world.Shape) {
	lo, hi := sh.Bounds()
	x0, y0, x1, y1 := s.rangeOf(lo, hi)
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			buckets[y*s.n+x] = append(buckets[y*s.n+x], id)
		}
	}
}

func (s *staticIndex) rangeOf(lo, hi geom.Vec) (int, int, int, int) {
	c := func(v float64) int { return int(geom.Clamp(v/s.cell, 0, float64(s.n-1))) }
	return c(lo.X), c(lo.Y), c(hi.X), c(hi.Y)
}

// query calls fn once per id whose bucket overlaps the circle (p, r).
func (s *staticIndex) query(buckets [][]int, p geom.Vec, r float64, fn func(id int)) {
	s.stamp++
	x0, y0, x1, y1 := s.rangeOf(geom.Vec{X: p.X - r, Y: p.Y - r}, geom.Vec{X: p.X + r, Y: p.Y + r})
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			for _, id := range buckets[y*s.n+x] {
				if s.seen[id] == s.stamp {
					continue
				}
				s.seen[id] = s.stamp
				fn(id)
			}
		}
	}
}

// QueryObstacles calls fn for every obstacle near the circle (p, r).
func (g *Game) QueryObstacles(p Vec, r float64, fn func(id int)) {
	g.obsIndex.query(g.obsIndex.obs, p, r, fn)
}

// QueryZones calls fn for every zone near p.
func (g *Game) QueryZones(p Vec, r float64, fn func(id int)) {
	g.obsIndex.query(g.obsIndex.zones, p, r, fn)
}

// ObstacleSolid reports whether obstacle id currently blocks movement.
func (g *Game) ObstacleSolid(id int) bool { return !g.Obstacles[id].Broken }
