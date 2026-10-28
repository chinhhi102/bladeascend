package geom

import "math"

type Vec struct{ X, Y float64 }

func (a Vec) Add(b Vec) Vec       { return Vec{a.X + b.X, a.Y + b.Y} }
func (a Vec) Sub(b Vec) Vec       { return Vec{a.X - b.X, a.Y - b.Y} }
func (a Vec) Mul(k float64) Vec   { return Vec{a.X * k, a.Y * k} }
func (a Vec) Dot(b Vec) float64   { return a.X*b.X + a.Y*b.Y }
func (a Vec) Len() float64        { return math.Hypot(a.X, a.Y) }
func (a Vec) Dist(b Vec) float64  { return math.Hypot(a.X-b.X, a.Y-b.Y) }
func (a Vec) Dist2(b Vec) float64 { dx, dy := a.X-b.X, a.Y-b.Y; return dx*dx + dy*dy }
func FromAngle(a float64) Vec     { return Vec{math.Cos(a), math.Sin(a)} }
func (a Vec) Angle() float64      { return math.Atan2(a.Y, a.X) }
func (a Vec) Norm() Vec {
	l := a.Len()
	if l < 1e-9 {
		return Vec{1, 0}
	}
	return Vec{a.X / l, a.Y / l}
}

// pointSegDist2 is the squared distance from p to segment ab.
func PointSegDist2(p, a, b Vec) float64 {
	ab := b.Sub(a)
	l2 := ab.Dot(ab)
	if l2 < 1e-12 {
		return p.Dist2(a)
	}
	t := math.Max(0, math.Min(1, p.Sub(a).Dot(ab)/l2))
	return p.Dist2(a.Add(ab.Mul(t)))
}

func cross(a, b Vec) float64 { return a.X*b.Y - a.Y*b.X }

// segSegDist2 is the squared distance between segments p1q1 and p2q2.
func SegSegDist2(p1, q1, p2, q2 Vec) float64 {
	d1, d2 := q1.Sub(p1), q2.Sub(p2)
	r := p1.Sub(p2)
	den := cross(d1, d2)
	if math.Abs(den) > 1e-12 {
		// p1 + t·d1 = p2 + u·d2, với r = p1 − p2 thì t·d1 − u·d2 = −r.
		// Nhân chéo với d2 ra t = cross(d2, r)/den, với d1 ra u = cross(d1, r)/den.
		//
		// Bản cũ viết `cross(d2, r.Mul(-1))`, tức là ĐẢO DẤU cả t lẫn u, nên nó báo "hai
		// đoạn cắt nhau" đúng vào lúc phần KÉO DÀI NGƯỢC của chúng cắt nhau — hai thanh
		// kiếm cách nhau 207 đơn vị vẫn bị tính là va vào nhau (bắt được khi tối ưu mục 26,
		// bản nhanh lọc đúng theo hình học lại "lệch" so với bản cũ).
		t := cross(d2, r) / den
		u := cross(d1, r) / den
		if t >= 0 && t <= 1 && u >= 0 && u <= 1 {
			return 0
		}
	}
	return math.Min(
		math.Min(PointSegDist2(p1, p2, q2), PointSegDist2(q1, p2, q2)),
		math.Min(PointSegDist2(p2, p1, q1), PointSegDist2(q2, p1, q1)),
	)
}

// normAngle wraps an angle to (-π, π].
func NormAngle(a float64) float64 {
	a = math.Mod(a+math.Pi, 2*math.Pi)
	if a < 0 {
		a += 2 * math.Pi
	}
	return a - math.Pi
}

func Clamp(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }
