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
		{"same realm each lose u=1", ClashSide{2, 30, 0}, ClashSide{2, 27, 0}, 1, 1, 0, 0},
		// Khác cảnh giới: bên thấp luôn mất đúng u = 1, bên cao chỉ nứt lost/k (v4.1).
		{"Trúc Cơ vs Luyện Khí: thấp -1, cao nứt 1/1.2", ClashSide{1, 60, 0}, ClashSide{0, 50, 0}, 0, 1, 0.8333, 0},
		{"Kim Đan vs Luyện Khí: cao nứt 1/1.44", ClashSide{2, 60, 0}, ClashSide{0, 50, 0}, 0, 1, 0.6944, 0},
		{"Luyện Khí vs Nguyên Anh: thấp -1, cao nứt 1/1.728", ClashSide{0, 50, 0}, ClashSide{3, 100, 0}, 1, 0, 0, 0.5787},
		{"Luyện Khí(1 kiếm) vs Nguyên Anh: mất nốt cây cuối", ClashSide{0, 1, 0}, ClashSide{3, 100, 0.5}, 1, 1, 0, 0.0787},
		{"nứt đủ 1 thì bên cao mới mất kiếm", ClashSide{3, 100, 0.9}, ClashSide{0, 3, 0}, 1, 1, 0.4787, 0},
		{"Đại Thừa vs Luyện Khí: cao nứt 1/4.3", ClashSide{8, 400, 0}, ClashSide{0, 60, 0}, 0, 1, 0.2326, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := ResolveClash(c, tc.a, tc.b)
			if out.LostA != tc.lostA || out.LostB != tc.lostB {
				t.Fatalf("lost = %d,%d want %d,%d", out.LostA, out.LostB, tc.lostA, tc.lostB)
			}
			// crack giờ là 1/1.2^d nên là số lẻ vô hạn: so tới 4 chữ số thập phân.
			if math.Abs(out.CrackA-tc.crackA) > 1e-4 || math.Abs(out.CrackB-tc.crackB) > 1e-4 {
				t.Fatalf("crack = %v,%v want %v,%v", out.CrackA, out.CrackB, tc.crackA, tc.crackB)
			}
			if tc.a.Count-out.LostA < c.MinCount || tc.b.Count-out.LostB < c.MinCount {
				t.Fatalf("a ring dropped below min_count = %d", c.MinCount)
			}
		})
	}
}

// Spec §5.2: min_count = 0 nên người chơi có thể bị tước sạch kiếm, và khi đó
// vòng kiếm không chạm được ai nữa.
func TestRingCanBeStrippedToZero(t *testing.T) {
	g := quietGame(t, nil)
	if g.Cfg.Sword.MinCount != 0 {
		t.Fatalf("min_count = %d, luật hiện tại phải là 0", g.Cfg.Sword.MinCount)
	}
	a := g.AddPlayer(SpawnOptions{Name: "A", Level: 20, Swords: 20})
	victim := g.AddPlayer(SpawnOptions{Name: "V", Level: 20, Swords: 1})
	// Nạn nhân đứng sát thân A, nằm gọn trong tầm vòng kiếm.
	place(g, a, Vec{X: 6000, Y: 6000}, 0)
	place(g, victim, Vec{X: 6000 + g.Geom(a).Inner, Y: 6000}, 0)
	a.Ring.SweepLo, a.Ring.Sweep = 0, 2*math.Pi // quét trọn vòng như một tick bình thường

	// Đối chứng: còn kiếm thì chạm được.
	if _, hit := g.swordTouchesCircle(a, victim.Pos, g.BodyR(victim)); !hit {
		t.Fatal("thiết lập sai: A còn 20 kiếm mà không chạm được nạn nhân")
	}
	// Tước sạch kiếm — min_count = 0 nên xuống tới 0.
	if lost := g.LoseSwords(a, 99); lost != 20 || a.Ring.Count != 0 {
		t.Fatalf("mất %d kiếm, còn lại %d; muốn mất 20 còn 0", lost, a.Ring.Count)
	}
	if _, hit := g.swordTouchesCircle(a, victim.Pos, g.BodyR(victim)); hit {
		t.Fatal("người 0 kiếm vẫn chạm được nạn nhân")
	}
}

// Spec v4.1 §5.2: k = 1.2^d ("compound") hoặc 1 + 0.2×d ("linear").
func TestDurabilityRatio(t *testing.T) {
	c := config.MustLoadDefault().Sword
	for d, want := range map[int]float64{0: 1, 1: 1.2, 2: 1.44, 3: 1.728, 8: 4.2998} {
		if got := Durability(c, d, 0); math.Abs(got-want) > 1e-3 {
			t.Errorf("compound d=%d: k = %v, muốn ≈%v", d, got, want)
		}
	}
	if got := Durability(c, 0, 3); math.Abs(got-1.728) > 1e-3 {
		t.Errorf("k phải đối xứng theo |d|: %v", got)
	}
	c.DurabilityMode = "linear"
	if got := Durability(c, 3, 0); math.Abs(got-1.6) > 1e-9 {
		t.Errorf("linear d=3: k = %v, muốn 1.6", got)
	}
}

// Spec v4.1 ví dụ: Trúc Cơ va Luyện Khí 12 lần → Luyện Khí mất 12, Trúc Cơ mất 10.
func TestDurabilityTwelveClashesTradeTwelveForTen(t *testing.T) {
	c := config.MustLoadDefault().Sword
	low := ClashSide{Realm: 0, Count: 100}
	high := ClashSide{Realm: 1, Count: 100}
	lostLow, lostHigh := 0, 0
	for i := 0; i < 12; i++ {
		out := ResolveClash(c, low, high)
		lostLow += out.LostA
		lostHigh += out.LostB
		low.Count -= out.LostA
		high.Count -= out.LostB
		low.Crack, high.Crack = out.CrackA, out.CrackB
	}
	if lostLow != 12 {
		t.Errorf("Luyện Khí mất %d kiếm sau 12 lần va, muốn 12", lostLow)
	}
	if lostHigh != 10 {
		t.Errorf("Trúc Cơ mất %d kiếm sau 12 lần va, muốn 10", lostHigh)
	}
}

func swordCfg(t *testing.T) config.Sword {
	t.Helper()
	return config.MustLoadDefault().Sword
}

// geomFor builds the ring geometry for n swords at scale 1 with the real config.
func geomFor(t *testing.T, n int) RingGeom {
	t.Helper()
	c := config.MustLoadDefault()
	return ComputeRingGeom(c.Sword, c.Player.BodyR, 1, n, 1)
}

// Spec §4.1a: vòng thứ k chứa step × k kiếm — vòng 1 là 10, vòng 2 là 20, vòng 10 là 100.
func TestRingCapacityGrowsByTen(t *testing.T) {
	g := geomFor(t, 1)
	if g.Step != 10 {
		t.Fatalf("ring_step = %d, muốn 10", g.Step)
	}
	for k := 0; k < 10; k++ {
		if got, want := g.RingCap(k), 10*(k+1); got != want {
			t.Errorf("vòng %d chứa %d kiếm, muốn %d", k+1, got, want)
		}
	}
	// Cộng dồn: 10, 30, 60, 100, 150, 210, 280, 360, 450, 550.
	for k, want := range []int{0, 10, 30, 60, 100, 150, 210, 280, 360, 450, 550} {
		if got := g.RingStart(k); got != want {
			t.Errorf("kiếm đầu của vòng %d là #%d, muốn #%d", k+1, got, want)
		}
	}
	// Mỗi kiếm phải rơi đúng vòng của nó.
	for i := 0; i < 550; i++ {
		k := g.RingOf(i)
		if i < g.RingStart(k) || i >= g.RingStart(k+1) {
			t.Fatalf("kiếm #%d bị xếp vào vòng %d (khoảng %d..%d)", i, k+1, g.RingStart(k), g.RingStart(k+1))
		}
	}
}

// Số vòng vừa đủ chứa hết số kiếm, không thừa vòng rỗng.
func TestRingsForSwordCount(t *testing.T) {
	cases := map[int]int{1: 1, 10: 1, 11: 2, 30: 2, 31: 3, 60: 3, 120: 5, 160: 6, 280: 7, 400: 9}
	for n, want := range cases {
		if got := geomFor(t, n).Rings; got != want {
			t.Errorf("%d kiếm → %d vòng, muốn %d", n, got, want)
		}
	}
}

// Spec §4.1a: vòng đầy thì chia đều 360°, và vòng k lệch đi k × (khe/R).
func TestSlotAngleStagger(t *testing.T) {
	deg := math.Pi / 180
	g := geomFor(t, 30) // 2 vòng: 10 + 20 kiếm
	if g.Rings != 2 {
		t.Fatalf("30 kiếm phải thành 2 vòng, đang %d", g.Rings)
	}
	// Vòng 1: 10 khe, mỗi khe 36°, không lệch.
	if a := g.SlotAngle(30, 0); math.Abs(a) > 1e-9 {
		t.Errorf("kiếm đầu vòng 1 ở %v°, muốn 0°", a/deg)
	}
	if d := g.SlotAngle(30, 1) - g.SlotAngle(30, 0); math.Abs(d-36*deg) > 1e-9 {
		t.Errorf("khe vòng 1 = %v°, muốn 36°", d/deg)
	}
	// Vòng 2: 20 khe, mỗi khe 18°, lệch 18°/2 = 9°.
	stagger := 18 * deg / 2
	if a := g.SlotAngle(30, 10); math.Abs(a-stagger) > 1e-9 {
		t.Errorf("kiếm đầu vòng 2 ở %v°, muốn %v°", a/deg, stagger/deg)
	}
	if d := g.SlotAngle(30, 11) - g.SlotAngle(30, 10); math.Abs(d-18*deg) > 1e-9 {
		t.Errorf("khe vòng 2 = %v°, muốn 18°", d/deg)
	}
	// Vòng 1 chưa đầy thì chia đều số kiếm đang có trên 10 khe.
	g5 := geomFor(t, 5)
	if d := g5.SlotAngle(5, 1); math.Abs(d-72*deg) > 1e-9 {
		t.Errorf("5 kiếm: khe %v°, muốn 72°", d/deg)
	}
}

// Vòng chưa đủ kiếm thì rải đều trên các khe của chính vòng đó.
func TestPartialRingSpreadsOverSlots(t *testing.T) {
	deg := math.Pi / 180
	g := geomFor(t, 15) // vòng 1 đủ 10, vòng 2 mới có 5/20 kiếm
	stagger := 18 * deg / 2
	for j := 0; j < 5; j++ {
		got := g.SlotAngle(15, 10+j)
		want := float64(j*4)*18*deg + stagger // 5 kiếm trên 20 khe → cách 4 khe
		if math.Abs(got-want) > 1e-9 {
			t.Errorf("kiếm %d ở %v°, muốn %v°", j, got/deg, want/deg)
		}
	}
}

func TestReachGrowsWithRingsAndOuterTipsAreFarther(t *testing.T) {
	c := swordCfg(t)
	// 10 kiếm = vừa đủ vòng 1; 60 kiếm = đủ 3 vòng (10 + 20 + 30).
	g10 := ComputeRingGeom(c, 30, 1, 10, 1)
	g60 := ComputeRingGeom(c, 30, 1, 60, 1)
	if g10.Rings != 1 || g60.Rings != 3 || g60.Reach <= g10.Reach {
		t.Fatalf("rings %d/%d reach %v/%v", g10.Rings, g60.Rings, g10.Reach, g60.Reach)
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
	long := ComputeRingGeom(c, 30, 1, 10, 1.3)
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
	spin := g.Cfg.Buffs["spin"].Mult
	g.ApplyBuff(p, "spin")
	if w := g.Omega(p); math.Abs(w-2*math.Pi/g.Cfg.Sword.RevSecs*spin) > 1e-9 {
		t.Fatalf("Tốc Kiếm omega %v, want ×%v", w, spin)
	}
}

// Spec v3.2: ring radius and length depend only on the ring index and scale, never on n.
func TestRingRadiusIndependentOfSwordCount(t *testing.T) {
	c := swordCfg(t)
	one := ComputeRingGeom(c, 30, 1.4, 1, 1)
	for _, n := range []int{5, 10} { // vòng 1 chứa 10 kiếm
		g := ComputeRingGeom(c, 30, 1.4, n, 1)
		if g.HiltR(0) != one.HiltR(0) || g.Len(0) != one.Len(0) || g.Reach != one.Reach {
			t.Fatalf("n=%d changed ring 1: %+v vs %+v", n, g, one)
		}
	}
	if one.HiltR(0) != (30+c.HiltGap)*1.4 {
		t.Fatalf("ring 1 hilt %v must sit next to the body", one.HiltR(0))
	}
	g11, g30 := ComputeRingGeom(c, 30, 1.4, 11, 1), ComputeRingGeom(c, 30, 1.4, 30, 1)
	if g11.Rings != 2 || g11.HiltR(1) <= g11.HiltR(0) || g11.Reach <= one.Reach {
		t.Fatalf("kiếm thứ 11 phải mở vòng 2 rộng hơn: %+v", g11)
	}
	if g30.Rings != 2 || g30.Reach != g11.Reach || g30.HiltR(1) != g11.HiltR(1) {
		t.Fatalf("30 kiếm vẫn 2 vòng, bán kính phải y như 11 kiếm: %+v vs %+v", g30, g11)
	}
	full := ComputeRingGeom(c, 30, 1.4, 60, 1) // Luyện Khí 10, full swords: spec says ≈ 200
	if full.Reach < 180 || full.Reach > 230 {
		t.Fatalf("Luyện Khí 10 reach %v, spec reference ≈ 200", full.Reach)
	}
}

// Spec v4 §6.2: Tốc Kiếm CHỈ nhân vào ω. Nó không được làm tầm chém dài ra —
// đó là việc của Trường Kiếm.
func TestSpinBuffDoesNotChangeReach(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "P", Level: 20, Swords: 30})
	before := g.Geom(p).Reach
	g.ApplyBuff(p, "spin")
	if after := g.Geom(p).Reach; after != before {
		t.Fatalf("Tốc Kiếm làm tầm chém đổi %v → %v, phải giữ nguyên", before, after)
	}
	// Và phải quay đúng 360° × mult trong 1 giây.
	want := 2 * math.Pi / g.Cfg.Sword.RevSecs * g.Cfg.Buffs["spin"].Mult
	if got := g.Omega(p); math.Abs(got-want) > 1e-9 {
		t.Fatalf("ω khi có Tốc Kiếm = %v, muốn %v", got, want)
	}
	// Trường Kiếm thì ngược lại: dài ra nhưng không đổi tốc.
	w := g.Omega(p)
	g.ApplyBuff(p, "long")
	if after := g.Geom(p).Reach; after <= before {
		t.Fatalf("Trường Kiếm phải làm tầm chém dài ra: %v → %v", before, after)
	}
	if got := g.Omega(p); math.Abs(got-w) > 1e-9 {
		t.Fatalf("Trường Kiếm làm đổi ω %v → %v", w, got)
	}
}

// Tốc độ mài: hai vòng ngoài cùng ngang cảnh giới, 1 vòng/giây → sau đúng 1 giây cả hai
// rụng trọn vòng ngoài. Bể kiếm tính mỗi tick (clash_cooldown = 0), không phải mỗi 0.3s.
func TestClashGrindsOneOuterRingPerSecond(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Level: 20, Swords: 60}) // 10+20+30 = 3 vòng
	b := g.AddPlayer(SpawnOptions{Name: "B", Level: 20, Swords: 60})
	at, bt := Vec{X: 6000, Y: 6000}, Vec{X: 6200, Y: 6000}
	place(g, a, at, 0)
	place(g, b, bt, 0)
	if g.Geom(a).Rings != 3 {
		t.Fatalf("60 kiếm phải là 3 vòng, đang %d", g.Geom(a).Rings)
	}
	a.InvulnUntil, b.InvulnUntil = 1e9, 1e9 // chỉ đo bể kiếm, không để rớt cảnh giới xen vào
	for i := 0; i < 30; i++ {               // đúng 1 giây
		g.Tick(dt)
		a.Pos, b.Pos = at, bt // giữ nguyên chỗ để đo
	}
	// Vòng ngoài 30 kiếm: sau 1 giây phải sạch, không hơn không kém một vòng.
	for _, p := range []*Player{a, b} {
		if p.Ring.Count != 30 {
			t.Fatalf("%s còn %d kiếm sau 1 giây, muốn 30 (rụng trọn vòng ngoài 30 kiếm)", p.Name, p.Ring.Count)
		}
	}
}

// Đổi chiều quay vẫn phải có cooldown riêng, nếu không vòng kiếm giật 30 lần/giây.
func TestRingFlipKeepsItsOwnCooldown(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Level: 20, Swords: 60})
	b := g.AddPlayer(SpawnOptions{Name: "B", Level: 20, Swords: 60})
	at, bt := Vec{X: 6000, Y: 6000}, Vec{X: 6200, Y: 6000}
	place(g, a, at, 0)
	place(g, b, bt, 0)
	a.InvulnUntil, b.InvulnUntil = 1e9, 1e9
	flips, prev := 0, a.Ring.Dir
	for i := 0; i < 30; i++ { // 1 giây
		g.Tick(dt)
		a.Pos, b.Pos = at, bt
		if a.Ring.Dir != prev {
			flips++
			prev = a.Ring.Dir
		}
	}
	// flip_cooldown = 0.3s → nhiều nhất 4 lần đổi chiều trong 1 giây.
	if flips > 4 {
		t.Fatalf("đổi chiều %d lần trong 1 giây — cooldown đổi chiều không còn tác dụng", flips)
	}
	if flips == 0 {
		t.Fatal("cùng cảnh giới mà không đổi chiều lần nào")
	}
}
