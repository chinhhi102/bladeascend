package game

import "math"

// itemGrid buckets items by position so pickups and AOI only look at nearby cells; the map
// holds thousands of ambient EXP orbs (spec §6.1).
type itemGrid struct {
	cell  float64
	n     int
	cells []map[uint32]*Item
}

func newItemGrid(size, cell float64) *itemGrid {
	n := max(1, int(math.Ceil(size/cell)))
	return &itemGrid{cell: cell, n: n, cells: make([]map[uint32]*Item, n*n)}
}

func (ig *itemGrid) index(p Vec) int {
	x := min(max(int(p.X/ig.cell), 0), ig.n-1)
	y := min(max(int(p.Y/ig.cell), 0), ig.n-1)
	return y*ig.n + x
}

func (ig *itemGrid) add(it *Item) {
	i := ig.index(it.Pos)
	if ig.cells[i] == nil {
		ig.cells[i] = map[uint32]*Item{}
	}
	ig.cells[i][it.ID] = it
}

func (ig *itemGrid) remove(it *Item) { delete(ig.cells[ig.index(it.Pos)], it.ID) }

// each calls fn for every item in cells overlapping the rectangle [lo, hi].
func (ig *itemGrid) each(lo, hi Vec, fn func(*Item)) {
	a, b := ig.index(lo), ig.index(hi)
	x0, y0, x1, y1 := a%ig.n, a/ig.n, b%ig.n, b/ig.n
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			for _, it := range ig.cells[y*ig.n+x] {
				fn(it)
			}
		}
	}
}

// ItemsNear calls fn for items in cells overlapping the square of half-size r around c.
func (g *Game) ItemsNear(c Vec, rx, ry float64, fn func(*Item)) {
	g.itemGrid.each(Vec{X: c.X - rx, Y: c.Y - ry}, Vec{X: c.X + rx, Y: c.Y + ry}, fn)
}
