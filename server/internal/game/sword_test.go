package game

import (
	"math"
	"testing"

	"kiemxoay/server/internal/config"
)

// Every row of spec §5.2 and the acceptance checklist §17.
func TestResolveClashTable(t *testing.T) {
	c := config.MustLoadDefault().Sword
	cases := []struct {
		name           string
		a, b           ClashSide
		lostA, lostB   int
		crackA, crackB float64
	}{
		{"same realm each lose 1", ClashSide{2, 10, 0}, ClashSide{2, 7, 0}, 1, 1, 0, 0},
		{"same realm, one at 1 sword", ClashSide{0, 1, 0}, ClashSide{0, 4, 0}, 0, 1, 0, 0},
		{"Trúc Cơ vs Luyện Khí 1:2", ClashSide{1, 6, 0}, ClashSide{0, 5, 0}, 1, 2, 0, 0},
		{"Kim Đan vs Luyện Khí 1:3", ClashSide{2, 6, 0}, ClashSide{0, 5, 0}, 1, 3, 0, 0},
		{"Luyện Khí(5) vs Nguyên Anh: 4 and 1", ClashSide{0, 5, 0}, ClashSide{3, 20, 0}, 4, 1, 0, 0},
		{"Luyện Khí(3) vs Nguyên Anh: crack 2/4", ClashSide{0, 3, 0}, ClashSide{3, 20, 0}, 2, 0, 0, 0.5},
		{"Đại Thừa vs Luyện Khí 1:9", ClashSide{8, 45, 0}, ClashSide{0, 5, 0}, 0, 4, 4.0 / 9, 0},
		{"low side already at 1", ClashSide{0, 1, 0}, ClashSide{3, 20, 0.5}, 0, 0, 0, 0.5},
		{"crack accumulates to a break", ClashSide{3, 20, 0.5}, ClashSide{0, 3, 0}, 1, 2, 0, 0},
		{"high side at 1 sword keeps 1", ClashSide{3, 1, 0}, ClashSide{0, 5, 0}, 0, 4, 0, 0},
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
