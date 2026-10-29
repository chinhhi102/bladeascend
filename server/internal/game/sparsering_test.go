package game

import (
	"math"
	"math/rand"
	"testing"

	"kiemxoay/server/internal/config"
	"kiemxoay/server/internal/geom"
	"kiemxoay/server/internal/world"
)

// sparseRingsTouchNaive là bản gốc, chậm nhưng hiển nhiên đúng: không tính sẵn kiếm của b,
// không lọc thô. Bản nhanh trong combat.go phải cho ra ĐÚNG cùng kết quả — nó nằm trên
// đường va kiếm, sai một chút là luật chơi sai.
func sparseRingsTouchNaive(a, b *Player) (Vec, bool) {
	n := max(sweepSteps(a.Ring), sweepSteps(b.Ring))
	reach := (a.Geom.Width + b.Geom.Width) / 2
	for s := 0; s <= n; s++ {
		pa, pb := sweepPhase(a.Ring, s, n), sweepPhase(b.Ring, s, n)
		for i := 0; i < a.Ring.Count; i++ {
			p1, q1 := swordAt(a.Pos, pa, a.Ring, a.Geom, i)
			for j := 0; j < b.Ring.Count; j++ {
				p2, q2 := swordAt(b.Pos, pb, b.Ring, b.Geom, j)
				if geom.SegSegDist2(p1, q1, p2, q2) <= reach*reach {
					return p1.Add(q1).Add(p2).Add(q2).Mul(0.25), true
				}
			}
		}
	}
	return Vec{}, false
}

func TestSparseRingsTouchMatchesNaive(t *testing.T) {
	_ = t
	cfg := config.MustLoadDefault()
	cfg.Room.MonsterCount, cfg.Items.CountPerRoom = 0, 0
	g := New(cfg, world.EmptyMap(cfg.Room.MapSize), 1)
	rng := rand.New(rand.NewSource(7))
	for trial := 0; trial < 300; trial++ {
		a, b := &Player{}, &Player{}
		for _, p := range []*Player{a, b} {
			p.Level = rng.Intn(40)
			p.Ring = Ring{Count: 1 + rng.Intn(14), Dir: 1, Angle: rng.Float64() * 2 * math.Pi}
			p.Ring.Tilt = RestTilt(g.Cfg.Sword, p.Ring.Dir)
			g.refreshDerived(p)
		}
		// Đặt b quanh a ở nhiều khoảng cách: chồng lên nhau, vừa chạm, và quá xa.
		d := (a.Geom.Outer() + b.Geom.Outer()) * rng.Float64() * 1.2
		ang := rng.Float64() * 2 * math.Pi
		b.Pos = a.Pos.Add(Vec{X: math.Cos(ang) * d, Y: math.Sin(ang) * d})

		wantAt, want := sparseRingsTouchNaive(a, b)
		gotAt, got := g.sparseRingsTouch(a, b)
		if got != want {
			t.Fatalf("lần %d: bản nhanh nói chạm=%v, bản gốc nói %v (khoảng cách %.1f, kiếm %d/%d)",
				trial, got, want, d, a.Ring.Count, b.Ring.Count)
		}
		if want && (math.Abs(gotAt.X-wantAt.X) > 1e-9 || math.Abs(gotAt.Y-wantAt.Y) > 1e-9) {
			t.Fatalf("lần %d: điểm chạm lệch: %v vs %v", trial, gotAt, wantAt)
		}
	}
}
