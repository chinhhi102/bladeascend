package game

import (
	"math"
	"testing"

	"kiemxoay/server/internal/config"
)

func testProg(t *testing.T) *Progression {
	t.Helper()
	return NewProgression(config.MustLoadDefault().Cultivation)
}

func TestExpNeededMatchesSpecTable(t *testing.T) {
	p := testProg(t)
	cases := []struct{ L int; want float64 }{
		{0, 100}, {9, 172}, {10, 120}, {19, 206}, {20, 144}, {29, 248},
		{30, 173}, {39, 297}, {80, 430}, {89, 740},
	}
	for _, c := range cases {
		if got := p.ExpNeeded(c.L); got != c.want {
			t.Errorf("ExpNeeded(%d) = %v, want %v", c.L, got, c.want)
		}
	}
}

func TestRealmTotalsMatchSpec(t *testing.T) {
	p := testProg(t)
	want := []float64{1360, 1632, 1958, 2350, 2820, 3384, 4061, 4873, 5848}
	for r, w := range want {
		sum := 0.0
		for s := 0; s < 10; s++ {
			sum += p.ExpNeeded(r*10 + s)
		}
		// Spec table rounds each tier; allow ±2 for cumulative rounding.
		if math.Abs(sum-w) > 2 {
			t.Errorf("realm %d total = %v, want ≈%v", r, sum, w)
		}
	}
}

func TestLevelFromExp(t *testing.T) {
	p := testProg(t)
	if L := p.LevelFromExp(0); L != 0 {
		t.Fatalf("0 exp → level %d", L)
	}
	if L := p.LevelFromExp(99.9); L != 0 {
		t.Fatalf("99.9 exp → level %d", L)
	}
	if L := p.LevelFromExp(100); L != 1 {
		t.Fatalf("100 exp → level %d", L)
	}
	if L := p.LevelFromExp(p.Threshold(10)); L != 10 {
		t.Fatalf("threshold(10) → level %d", L)
	}
	if L := p.LevelFromExp(1e9); L != 89 {
		t.Fatalf("huge exp → level %d, want 89", L)
	}
}

func TestScaleRange(t *testing.T) {
	p := testProg(t)
	if s := p.Scale(0); s != 1 {
		t.Errorf("scale(0) = %v", s)
	}
	if s := p.Scale(89); math.Abs(s-5) > 1e-9 {
		t.Errorf("scale(89) = %v, want 5", s)
	}
	if p.Scale(9) <= p.Scale(0) {
		t.Error("Luyện Khí 10 must be bigger than Luyện Khí 1")
	}
	if s := p.Scale(10); math.Abs(s-1.449) > 0.01 {
		t.Errorf("Trúc Cơ 1 scale = %v, want ≈1.45", s)
	}
}

func TestMaxSwords(t *testing.T) {
	p := testProg(t)
	for r, want := range []int{5, 10, 15, 20, 25, 30, 35, 40, 45} {
		if got := p.MaxSwords(r); got != want {
			t.Errorf("MaxSwords(%d) = %d, want %d", r, got, want)
		}
	}
}

func TestRingGeomGrowsWithScale(t *testing.T) {
	c := config.MustLoadDefault().Sword
	small := ComputeRingGeom(c, 1, 5, 1)
	big := ComputeRingGeom(c, 5, 5, 1)
	if big.Radius <= small.Radius || big.Length <= small.Length {
		t.Fatalf("ring must grow with scale: %+v vs %+v", small, big)
	}
	if small.Radius != 70 || small.Length != 40 || small.Width != 10 {
		t.Fatalf("scale 1, 5 swords: %+v", small)
	}
	long := ComputeRingGeom(c, 1, 5, 1.3)
	if math.Abs(long.Radius-91) > 1e-9 || math.Abs(long.Length-52) > 1e-9 {
		t.Fatalf("Trường Kiếm geometry: %+v", long)
	}
	many := ComputeRingGeom(c, 1, 45, 1)
	if math.Abs(many.Radius-45*22/(2*math.Pi)) > 1e-9 {
		t.Fatalf("45 swords radius = %v", many.Radius)
	}
}
