package world

import (
	"container/heap"
	"math"

	"kiemxoay/server/internal/geom"
)

// Grid is the A* path grid (spec §10). Cells overlapping an intact obstacle are blocked.
type Grid struct {
	Cell    float64
	N       int
	owners  [][]int // obstacle ids covering each cell
	blocked []bool
}

func NewGrid(m *Map, cell float64) *Grid {
	n := int(math.Ceil(m.Size / cell))
	g := &Grid{Cell: cell, N: n, owners: make([][]int, n*n), blocked: make([]bool, n*n)}
	for _, o := range m.Obstacles {
		g.indexObstacle(o)
	}
	return g
}

func (g *Grid) indexObstacle(o Obstacle) {
	lo, hi := o.Block().Bounds()
	x0, y0 := g.cellOf(lo)
	x1, y1 := g.cellOf(hi)
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			c := geom.Vec{X: (float64(x) + 0.5) * g.Cell, Y: (float64(y) + 0.5) * g.Cell}
			if o.Block().OverlapsCircle(c, g.Cell*0.5) {
				g.owners[y*g.N+x] = append(g.owners[y*g.N+x], o.ID)
			}
		}
	}
}

// Rebuild recomputes blocked cells; isSolid reports whether an obstacle currently blocks.
func (g *Grid) Rebuild(isSolid func(id int) bool) {
	for i, ids := range g.owners {
		g.blocked[i] = false
		for _, id := range ids {
			if isSolid(id) {
				g.blocked[i] = true
				break
			}
		}
	}
}

func (g *Grid) cellOf(p geom.Vec) (int, int) {
	x := int(geom.Clamp(p.X/g.Cell, 0, float64(g.N-1)))
	y := int(geom.Clamp(p.Y/g.Cell, 0, float64(g.N-1)))
	return x, y
}

func (g *Grid) Blocked(p geom.Vec) bool {
	x, y := g.cellOf(p)
	return g.blocked[y*g.N+x]
}

func (g *Grid) center(i int) geom.Vec {
	return geom.Vec{X: (float64(i%g.N) + 0.5) * g.Cell, Y: (float64(i/g.N) + 0.5) * g.Cell}
}

type pqItem struct {
	idx int
	f   float64
}
type pq []pqItem

func (q pq) Len() int           { return len(q) }
func (q pq) Less(i, j int) bool { return q[i].f < q[j].f }
func (q pq) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q *pq) Push(x any)        { *q = append(*q, x.(pqItem)) }
func (q *pq) Pop() any          { o := *q; it := o[len(o)-1]; *q = o[:len(o)-1]; return it }

var dirs8 = [8][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}

// FindPath runs A* and returns waypoints (cell centres) from→to, or nil when unreachable
// within maxExpand node expansions (keeps bot cost bounded).
func (g *Grid) FindPath(from, to geom.Vec, maxExpand int) []geom.Vec {
	fx, fy := g.cellOf(from)
	tx, ty := g.cellOf(to)
	start, goal := fy*g.N+fx, ty*g.N+tx
	if g.blocked[goal] {
		return nil
	}
	cost := map[int]float64{start: 0}
	parent := map[int]int{}
	open := &pq{{start, 0}}
	for expanded := 0; open.Len() > 0 && expanded < maxExpand; expanded++ {
		cur := heap.Pop(open).(pqItem).idx
		if cur == goal {
			return g.trace(parent, start, goal)
		}
		g.expand(cur, goal, cost, parent, open)
	}
	return nil
}

func (g *Grid) expand(cur, goal int, cost map[int]float64, parent map[int]int, open *pq) {
	cx, cy := cur%g.N, cur/g.N
	gx, gy := goal%g.N, goal/g.N
	for _, d := range dirs8 {
		nx, ny := cx+d[0], cy+d[1]
		if nx < 0 || ny < 0 || nx >= g.N || ny >= g.N {
			continue
		}
		ni := ny*g.N + nx
		// Diagonal moves must not cut blocked corners.
		if g.blocked[ni] || g.blocked[cy*g.N+nx] || g.blocked[ny*g.N+cx] {
			continue
		}
		nc := cost[cur] + math.Hypot(float64(d[0]), float64(d[1]))
		if old, ok := cost[ni]; ok && old <= nc {
			continue
		}
		cost[ni] = nc
		parent[ni] = cur
		heap.Push(open, pqItem{ni, nc + math.Hypot(float64(gx-nx), float64(gy-ny))})
	}
}

func (g *Grid) trace(parent map[int]int, start, goal int) []geom.Vec {
	var rev []geom.Vec
	for c := goal; c != start; c = parent[c] {
		rev = append(rev, g.center(c))
	}
	out := make([]geom.Vec, len(rev))
	for i := range rev {
		out[i] = rev[len(rev)-1-i]
	}
	return out
}
