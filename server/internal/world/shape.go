package world

import (
	"math"

	"kiemxoay/server/internal/geom"
)

// Shape is an axis-aligned rectangle (X,Y top-left, W,H) or a circle (centre X,Y, radius R).
type Shape struct {
	Circle bool    `json:"c,omitempty"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	W      float64 `json:"w,omitempty"`
	H      float64 `json:"h,omitempty"`
	R      float64 `json:"r,omitempty"`
}

func (s Shape) Center() geom.Vec {
	if s.Circle {
		return geom.Vec{X: s.X, Y: s.Y}
	}
	return geom.Vec{X: s.X + s.W/2, Y: s.Y + s.H/2}
}

// Bounds returns min/max corners.
func (s Shape) Bounds() (geom.Vec, geom.Vec) {
	if s.Circle {
		return geom.Vec{X: s.X - s.R, Y: s.Y - s.R}, geom.Vec{X: s.X + s.R, Y: s.Y + s.R}
	}
	return geom.Vec{X: s.X, Y: s.Y}, geom.Vec{X: s.X + s.W, Y: s.Y + s.H}
}

func (s Shape) Contains(p geom.Vec) bool {
	if s.Circle {
		return p.Dist2(geom.Vec{X: s.X, Y: s.Y}) <= s.R*s.R
	}
	return p.X >= s.X && p.X <= s.X+s.W && p.Y >= s.Y && p.Y <= s.Y+s.H
}

// OverlapsCircle reports whether a circle (c, r) touches the shape.
func (s Shape) OverlapsCircle(c geom.Vec, r float64) bool {
	_, ok := s.PushOut(c, r)
	return ok
}

// PushOut returns the corrected centre that moves circle (c, r) out of the shape.
func (s Shape) PushOut(c geom.Vec, r float64) (geom.Vec, bool) {
	if s.Circle {
		center := geom.Vec{X: s.X, Y: s.Y}
		d := c.Dist(center)
		if d >= s.R+r {
			return c, false
		}
		return center.Add(c.Sub(center).Norm().Mul(s.R + r)), true
	}
	return s.pushOutRect(c, r)
}

func (s Shape) pushOutRect(c geom.Vec, r float64) (geom.Vec, bool) {
	closest := geom.Vec{X: geom.Clamp(c.X, s.X, s.X+s.W), Y: geom.Clamp(c.Y, s.Y, s.Y+s.H)}
	d := c.Sub(closest)
	dist := d.Len()
	if dist >= r {
		return c, false
	}
	if dist > 1e-6 {
		return closest.Add(d.Mul(r / dist)), true
	}
	// Centre is inside the rectangle: exit through the nearest edge.
	left, right := c.X-s.X, s.X+s.W-c.X
	top, bottom := c.Y-s.Y, s.Y+s.H-c.Y
	m := math.Min(math.Min(left, right), math.Min(top, bottom))
	switch m {
	case left:
		return geom.Vec{X: s.X - r, Y: c.Y}, true
	case right:
		return geom.Vec{X: s.X + s.W + r, Y: c.Y}, true
	case top:
		return geom.Vec{X: c.X, Y: s.Y - r}, true
	default:
		return geom.Vec{X: c.X, Y: s.Y + s.H + r}, true
	}
}

// SegmentHits reports whether a capsule (segment ab with half-width hw) touches the shape.
func (s Shape) SegmentHits(a, b geom.Vec, hw float64) bool {
	if s.Circle {
		return geom.PointSegDist2(geom.Vec{X: s.X, Y: s.Y}, a, b) <= (s.R+hw)*(s.R+hw)
	}
	// Sample along the segment; swords are short relative to obstacles.
	for i := 0; i <= 4; i++ {
		p := a.Add(b.Sub(a).Mul(float64(i) / 4))
		if s.OverlapsCircle(p, hw) {
			return true
		}
	}
	return false
}

// Vec2 is a small helper for tests and tools.
func Vec2(x, y float64) geom.Vec { return geom.Vec{X: x, Y: y} }
