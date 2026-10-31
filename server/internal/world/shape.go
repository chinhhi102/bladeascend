package world

import (
	"math"

	"bladeascend/server/internal/geom"
)

// Shape là hình chữ nhật thẳng trục (X,Y góc trên trái, W,H), hình tròn (tâm X,Y bán kính
// R), hoặc **đa giác** (Poly, toạ độ tuyệt đối, xen kẽ x,y).
//
// Đa giác là thứ spec §8.6 bắt buộc cho các vùng địa hình: bản đồ hiển thị không được có
// hình chữ nhật thẳng cạnh hay hình tròn hoàn hảo nào lộ ra. Quan trọng là **server dùng
// đúng đa giác mà client vẽ** — nếu server xét hình chữ nhật còn client vẽ hình méo thì
// người chơi đứng chỗ nhìn như bùn mà không bị chậm, hoặc ngược lại.
//
// X,Y,W,H vẫn được giữ cho đa giác: đó là hộp bao, dùng để lọc nhanh qua lưới không gian
// trước khi làm phép điểm-trong-đa-giác.
type Shape struct {
	Circle bool      `json:"c,omitempty"`
	X      float64   `json:"x"`
	Y      float64   `json:"y"`
	W      float64   `json:"w,omitempty"`
	H      float64   `json:"h,omitempty"`
	R      float64   `json:"r,omitempty"`
	Poly   []float64 `json:"p,omitempty"`
}

// SetPoly gán đa giác và tính luôn hộp bao.
func (s *Shape) SetPoly(pts []float64) {
	s.Poly = pts
	s.Circle = false
	if len(pts) < 6 {
		return
	}
	minX, minY := pts[0], pts[1]
	maxX, maxY := pts[0], pts[1]
	for i := 2; i < len(pts); i += 2 {
		minX = math.Min(minX, pts[i])
		maxX = math.Max(maxX, pts[i])
		minY = math.Min(minY, pts[i+1])
		maxY = math.Max(maxY, pts[i+1])
	}
	s.X, s.Y, s.W, s.H = minX, minY, maxX-minX, maxY-minY
}

// containsPoly là phép ray casting kinh điển: bắn một tia sang phải từ điểm cần xét rồi đếm
// số cạnh nó cắt — lẻ là trong, chẵn là ngoài.
func (s Shape) containsPoly(p geom.Vec) bool {
	in := false
	n := len(s.Poly) / 2
	for i, j := 0, n-1; i < n; j, i = i, i+1 {
		xi, yi := s.Poly[i*2], s.Poly[i*2+1]
		xj, yj := s.Poly[j*2], s.Poly[j*2+1]
		if (yi > p.Y) != (yj > p.Y) && p.X < (xj-xi)*(p.Y-yi)/(yj-yi)+xi {
			in = !in
		}
	}
	return in
}

func (s Shape) Center() geom.Vec {
	if s.Circle {
		return geom.Vec{X: s.X, Y: s.Y}
	}
	return geom.Vec{X: s.X + s.W/2, Y: s.Y + s.H/2}
}

// IsPoly reports whether this shape is an organic polygon (spec §8.6).
func (s Shape) IsPoly() bool { return len(s.Poly) >= 6 }

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
	// Hộp bao lọc trước: rẻ hơn ray casting rất nhiều và loại được hầu hết trường hợp.
	if p.X < s.X || p.X > s.X+s.W || p.Y < s.Y || p.Y > s.Y+s.H {
		return false
	}
	if s.IsPoly() {
		return s.containsPoly(p)
	}
	return true
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
