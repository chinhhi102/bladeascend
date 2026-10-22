package game

import (
	"math"
	"testing"

	"kiemxoay/server/internal/config"
)

// Every row of spec §5.2 (loss unit clash_unit = 5).
func TestResolveClashTable(t *testing.T) {
	c := config.MustLoadDefault().Sword
	cases := []struct {
		name           string
		a, b           ClashSide
		lostA, lostB   int
		crackA, crackB float64
	}{
		{"same realm each lose u=5", ClashSide{2, 30, 0}, ClashSide{2, 27, 0}, 5, 5, 0, 0},
		{"same realm, one at 3 swords keeps 1", ClashSide{0, 3, 0}, ClashSide{0, 40, 0}, 2, 5, 0, 0},
		{"Trúc Cơ vs Luyện Khí: -5 / -10", ClashSide{1, 60, 0}, ClashSide{0, 50, 0}, 5, 10, 0, 0},
		{"Kim Đan vs Luyện Khí: -5 / -15", ClashSide{2, 60, 0}, ClashSide{0, 50, 0}, 5, 15, 0, 0},
		{"Luyện Khí vs Nguyên Anh: -20 / -5", ClashSide{0, 50, 0}, ClashSide{3, 100, 0}, 20, 5, 0, 0},
		{"Luyện Khí(3) vs Nguyên Anh: crack 2/4", ClashSide{0, 3, 0}, ClashSide{3, 100, 0}, 2, 0, 0, 0.5},
		{"Đại Thừa vs Luyện Khí: -5 / -45", ClashSide{8, 400, 0}, ClashSide{0, 60, 0}, 5, 45, 0, 0},
		{"low side already at 1", ClashSide{0, 1, 0}, ClashSide{3, 100, 0.5}, 0, 0, 0, 0.5},
		{"crack accumulates to a break", ClashSide{3, 100, 0.5}, ClashSide{0, 3, 0}, 1, 2, 0, 0},
		{"high side at 1 sword keeps 1", ClashSide{3, 1, 0}, ClashSide{0, 50, 0}, 0, 20, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := ResolveClash(c, tc.a, tc.b)
			if out.LostA != tc.lostA || out.LostB != tc.lostB {
				t.Fatalf("lost = %d,%d want %d,%d", out.LostA, out.LostB, tc.lostA, tc.lostB)
			}
			if math.Abs(out.CrackA-tc.crackA) > 1e-9 || math.Abs(out.CrackB-tc.crackB) > 1e-9 {
				t.Fatalf("crack = %v,%v want %v,%v", out.CrackA, out.CrackB, tc.crackA, tc.crackB)
			}
			if tc.a.Count-out.LostA < 1 || tc.b.Count-out.LostB < 1 {
				t.Fatal("a ring dropped below 1 sword")
			}
		})
	}
}

func TestClashRatioConfigurable(t *testing.T) {
	c := config.MustLoadDefault().Sword
	if k := ClashRatio(c, 3, 0); k != 4 {
		t.Fatalf("k = %d, want 4", k)
	}
	c.ClashRatioPerRealm = 2
	if k := ClashRatio(c, 0, 3); k != 7 {
		t.Fatalf("k with per_realm=2 = %d, want 7", k)
	}
}

func swordCfg(t *testing.T) config.Sword {
	t.Helper()
	return config.MustLoadDefault().Sword
}

// Spec §4.1a v3.1: ring k is staggered by k * 18° / R.
func TestSlotAngleStagger(t *testing.T) {
	deg := math.Pi / 180
	for _, rings := range []int{1, 2, 3} {
		n := 20 * rings
		for k := 0; k < rings; k++ {
			got := SlotAngle(20, n, k*20)
			want := float64(k) * 18 * deg / float64(rings)
			if math.Abs(got-want) > 1e-9 {
				t.Errorf("R=%d ring %d first sword at %v°, want %v°", rings, k, got/deg, want/deg)
			}
			if d := SlotAngle(20, n, k*20+1) - got; math.Abs(d-18*deg) > 1e-9 {
				t.Errorf("R=%d ring %d slot step %v°, want 18°", rings, k, d/deg)
			}
		}
	}
	// n ≤ 20: evenly spread over 360°.
	if d := SlotAngle(20, 5, 1); math.Abs(d-72*deg) > 1e-9 {
		t.Errorf("5 swords: step %v°, want 72°", d/deg)
	}
}

// A ring with fewer than 20 swords still spreads them evenly over the 20 slots.
func TestPartialRingSpreadsOverSlots(t *testing.T) {
	deg := math.Pi / 180
	n := 25 // ring 1 holds 5 swords → every 4th slot
	stagger := 18 * deg / 2
	for j := 0; j < 5; j++ {
		got := SlotAngle(20, n, 20+j)
		want := float64(j*4)*18*deg + stagger
		if math.Abs(got-want) > 1e-9 {
			t.Errorf("sword %d at %v°, want %v°", j, got/deg, want/deg)
		}
	}
}

func TestReachGrowsWithRingsAndOuterTipsAreFarther(t *testing.T) {
	c := swordCfg(t)
	g20 := ComputeRingGeom(c, 30, 1, 20, 1)
	g60 := ComputeRingGeom(c, 30, 1, 60, 1)
	if g20.Rings != 1 || g60.Rings != 3 || g60.Reach <= g20.Reach {
		t.Fatalf("rings %d/%d reach %v/%v", g20.Rings, g60.Rings, g20.Reach, g60.Reach)
	}
	tilt := c.TiltDeg * math.Pi / 180
	if g60.TipDist(2, tilt) <= g60.TipDist(0, tilt) {
		t.Fatal("ring 3 tip must be farther from the centre than ring 1")
	}
	// Linear growth: each ring's radial tip extends ring_growth * base farther.
	for k := 1; k < 3; k++ {
		d := g60.HiltR(k) + g60.Len(k) - (g60.HiltR(k-1) + g60.Len(k-1))
		if math.Abs(d-c.RingGrowth*g60.Base) > 1e-9 {
			t.Errorf("ring %d grows %v, want %v", k, d, c.RingGrowth*g60.Base)
		}
	}
	big := ComputeRingGeom(c, 30, 5, 60, 1)
	if big.Reach <= g60.Reach*4.9 {
		t.Fatalf("reach must scale with the character: %v vs %v", big.Reach, g60.Reach)
	}
	long := ComputeRingGeom(c, 30, 1, 20, 1.3)
	if math.Abs(long.Len(0)-c.Length*1.3) > 1e-9 {
		t.Fatalf("Trường Kiếm length %v", long.Len(0))
	}
}

func TestTiltFlipsWithDirection(t *testing.T) {
	c := swordCfg(t)
	r := Ring{Dir: 1, Tilt: RestTilt(c, 1)}
	if r.Tilt >= 0 {
		t.Fatalf("blades must trail the spin: tilt %v for dir +1", r.Tilt)
	}
	r.Dir = -1
	for i := 0; i < 3; i++ {
		stepTilt(c, &r, 1.0/30)
	}
	if math.Abs(r.Tilt-RestTilt(c, -1)) > 1e-9 || r.Tilt <= 0 {
		t.Fatalf("after a flip and %vs the tilt should be %v, got %v", 3.0/30, RestTilt(c, -1), r.Tilt)
	}
}

// Spec §4.1b: a target the blade swept past between ticks is hit; one just outside the arc is not.
func TestSweptHitCatchesFastBlades(t *testing.T) {
	c := swordCfg(t)
	g := ComputeRingGeom(c, 30, 1, 1, 1)
	ring := Ring{Count: 1, Dir: 1, Tilt: RestTilt(c, 1), SweepLo: 0, Sweep: 40 * math.Pi / 180}
	h, tip := swordAt(Vec{}, 0, ring, g, 0)
	mid := h.Add(tip).Mul(0.5)
	d, base := mid.Len(), mid.Angle()
	hit := func(a float64) bool {
		found := false
		sweptHits(Vec{}, ring, g, Vec{X: d * math.Cos(a), Y: d * math.Sin(a)}, 5, func(int) bool {
			found = true
			return false
		})
		return found
	}
	if !hit(base + 20*math.Pi/180) {
		t.Fatal("target in the middle of the sweep must be hit")
	}
	if hit(base + 90*math.Pi/180) {
		t.Fatal("target far outside the sweep must not be hit")
	}
}

func TestOneRevolutionPerRevSecs(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "P", Swords: 1})
	if w := g.Omega(p); math.Abs(w*g.Cfg.Sword.RevSecs-2*math.Pi) > 1e-9 {
		t.Fatalf("omega %v does not make one turn per %vs", w, g.Cfg.Sword.RevSecs)
	}
	g.ApplyBuff(p, "spin")
	if w := g.Omega(p); math.Abs(w-2*math.Pi/g.Cfg.Sword.RevSecs*1.3) > 1e-9 {
		t.Fatalf("Tốc Kiếm omega %v, want ×1.3", w)
	}
}

// Spec v3.2: ring radius and length depend only on the ring index and scale, never on n.
func TestRingRadiusIndependentOfSwordCount(t *testing.T) {
	c := swordCfg(t)
	one := ComputeRingGeom(c, 30, 1.4, 1, 1)
	for _, n := range []int{10, 20} {
		g := ComputeRingGeom(c, 30, 1.4, n, 1)
		if g.HiltR(0) != one.HiltR(0) || g.Len(0) != one.Len(0) || g.Reach != one.Reach {
			t.Fatalf("n=%d changed ring 1: %+v vs %+v", n, g, one)
		}
	}
	if one.HiltR(0) != (30+c.HiltGap)*1.4 {
		t.Fatalf("ring 1 hilt %v must sit next to the body", one.HiltR(0))
	}
	g21, g40 := ComputeRingGeom(c, 30, 1.4, 21, 1), ComputeRingGeom(c, 30, 1.4, 40, 1)
	if g21.Rings != 2 || g21.HiltR(1) <= g21.HiltR(0) || g21.Reach <= one.Reach {
		t.Fatalf("21 swords must open a wider ring 2: %+v", g21)
	}
	if g40.Rings != 2 || g40.Reach != g21.Reach || g40.HiltR(1) != g21.HiltR(1) {
		t.Fatalf("40 swords must keep the 2-ring radii of 21: %+v vs %+v", g40, g21)
	}
	full := ComputeRingGeom(c, 30, 1.4, 60, 1) // Luyện Khí 10, full swords: spec says ≈ 200
	if full.Reach < 180 || full.Reach > 230 {
		t.Fatalf("Luyện Khí 10 reach %v, spec reference ≈ 200", full.Reach)
	}
}
