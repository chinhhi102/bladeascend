package game

import (
	"math"
	"testing"

	"kiemxoay/server/internal/world"
)

// altarGame builds a game whose map has one small altar and one centre altar.
func altarGame(t *testing.T) *Game {
	t.Helper()
	m := world.EmptyMap(12000)
	m.PlayMargin = 420
	m.Altars = []world.Altar{
		{ID: 0, Kind: "small", X: 3000, Y: 3000, R: 150},
		{ID: 1, Kind: "center", X: 6000, Y: 6000, R: 220},
	}
	g := quietGame(t, m)
	return g
}

func countKind(g *Game, kind string) (n int, total float64) {
	for _, it := range g.Items {
		if it.Kind == kind {
			n++
			total += it.Value
		}
	}
	return n, total
}

// Spec §8.5: đài nhỏ phun 2–3 viên 2 điểm mỗi 5s.
func TestSmallAltarSpraysOnSchedule(t *testing.T) {
	g := altarGame(t)
	ac := g.Cfg.Altars
	for g.Time < ac.Small.EverySecs-dt {
		g.Tick(dt)
	}
	if n, _ := countKind(g, ItemTienKhi); n != 0 {
		t.Fatalf("phun sớm: %d viên trước mốc %vs", n, ac.Small.EverySecs)
	}
	g.Tick(dt)
	g.Tick(dt)
	n, total := countKind(g, ItemTienKhi)
	if n < ac.Small.Orbs[0] || n > ac.Small.Orbs[1] {
		t.Fatalf("phun %d viên, muốn %d–%d", n, ac.Small.Orbs[0], ac.Small.Orbs[1])
	}
	if want := float64(n) * ac.Small.OrbValue; total != want {
		t.Fatalf("tổng điểm %v, muốn %v", total, want)
	}
	// Viên phải nằm trong bán kính spray_r quanh đài.
	for _, it := range g.Items {
		if it.Kind != ItemTienKhi {
			continue
		}
		d := it.Pos.Dist(Vec{X: 3000, Y: 3000})
		if d < ac.SprayR[0]-1 || d > ac.SprayR[1]+1 {
			t.Errorf("viên cách đài %.0f, ngoài khoảng %v", d, ac.SprayR)
		}
	}
}

// Spec §8.5: Thiên Tiên Đài mỗi 150s thả 1 viên 150 điểm, báo trước 10s toàn phòng.
func TestCenterAltarGiantOrbAndWarning(t *testing.T) {
	g := altarGame(t)
	ac := g.Cfg.Altars
	warned := false
	for g.Time < ac.Center.EverySecs+2*dt {
		g.Tick(dt)
		for _, e := range g.DrainEvents() {
			if e.Kind == "AltarWarn" && e.S == "center" {
				if !e.Global {
					t.Error("báo trước của Thiên Tiên Đài phải gửi toàn phòng")
				}
				if left := ac.Center.EverySecs - g.Time; math.Abs(left-ac.Center.WarnSecs) > 0.5 {
					t.Errorf("báo trước lúc còn %.1fs, muốn %vs", left, ac.Center.WarnSecs)
				}
				warned = true
			}
		}
	}
	if !warned {
		t.Fatal("không có thông báo trước của Thiên Tiên Đài")
	}
	giant := 0
	for _, it := range g.Items {
		if it.Kind == ItemTienKhi && it.Value == ac.Center.OrbValue {
			giant++
		}
	}
	if giant != 1 {
		t.Fatalf("%d viên khổng lồ, muốn 1", giant)
	}
}

// Spec §8.5: viên Tiên khí tan sau orb_lifetime_secs, không mọc lại chỗ khác.
func TestTienKhiOrbExpires(t *testing.T) {
	g := altarGame(t)
	g.addTienKhiOrb(Vec{X: 5000, Y: 5000}, 2)
	if n, _ := countKind(g, ItemTienKhi); n != 1 {
		t.Fatal("chưa tạo được viên")
	}
	for g.Time < g.Cfg.Altars.OrbLifeSec+1 {
		g.Tick(dt)
	}
	for _, it := range g.Items {
		if it.Kind == ItemTienKhi && it.ExpireAt > 0 && g.Time >= it.ExpireAt {
			t.Fatal("viên quá hạn vẫn còn")
		}
	}
}

// Hai đường tu tách hẳn (§3.4, anh chốt 2026-10-28): Đại Thừa 10 trở lên chỉ hấp thụ Tiên
// khí, dưới đó chỉ hấp thụ Linh khí — không đổi qua lại nữa.
func TestTienKhiAndLinhKhiDoNotMix(t *testing.T) {
	g := altarGame(t)
	top := g.AddPlayer(SpawnOptions{Name: "Top", Level: g.Prog.TopMortalLevel(), Swords: 1})
	if got := g.GainTienKhi(top, 50); got != 50 || top.TienKhi != 50 {
		t.Fatalf("Đại Thừa 10 phải tích thẳng: nhận %v, thanh %v", got, top.TienKhi)
	}
	topExp := top.TotalExp
	g.GainExp(top, 10000, false)
	if math.Abs(top.TotalExp-topExp) > 1e-9 {
		t.Fatalf("người tu Tiên khí vẫn hấp thụ Linh khí: %v → %v", topExp, top.TotalExp)
	}

	low := g.AddPlayer(SpawnOptions{Name: "Low", Level: 0, Swords: 1})
	before := low.TotalExp
	if got := g.GainTienKhi(low, 2); got != 0 {
		t.Fatalf("Luyện Khí không được tích Tiên khí, nhận %v", got)
	}
	if math.Abs(low.TotalExp-before) > 1e-9 {
		t.Fatalf("Tiên khí vẫn đổi thành %v Linh khí cho người phàm", low.TotalExp-before)
	}
	g.GainExp(low, 100, false)
	if low.TotalExp <= before {
		t.Fatal("người phàm phải hấp thụ được Linh khí")
	}
}

// Viên nào không hấp thụ được thì đi xuyên qua, vẫn nằm đó cho người dùng được nhặt.
func TestOrbsOnlyGoToTheRightCultivator(t *testing.T) {
	g := altarGame(t)
	low := g.AddPlayer(SpawnOptions{Name: "Low", Level: 0, Swords: 1})
	place(g, low, Vec{X: 6000, Y: 6000}, 0)
	khi := g.addItem(ItemTienKhi, 10, low.Pos) // ngay dưới chân
	g.Tick(dt)
	if g.Items[khi.ID] == nil {
		t.Fatal("người phàm nhặt mất viên Tiên khí")
	}

	top := g.AddPlayer(SpawnOptions{Name: "Top", Level: g.Prog.TopMortalLevel(), Swords: 1})
	place(g, top, Vec{X: 7000, Y: 7000}, 0)
	orb := g.addItem("orb", 20, top.Pos)
	g.Tick(dt)
	if g.Items[orb.ID] == nil {
		t.Fatal("người tu Tiên khí nhặt mất viên Linh khí")
	}
	// Đổi chủ: đúng người thì nhặt được. Dời Low đi để hai người không chém nhau rớt cảnh giới.
	place(g, low, Vec{X: 9000, Y: 9000}, 0)
	place(g, top, g.Items[khi.ID].Pos, 0)
	top.InvulnUntil = 1e9
	g.Tick(dt)
	if g.Items[khi.ID] != nil {
		t.Fatal("Đại Thừa 10 phải nhặt được viên Tiên khí")
	}
}

// Spec §3.4: đủ mốc thì phi thăng; mốc cộng dồn là 100, 300, 600, 1000, 1500.
func TestTienKhiPromotesAtThresholds(t *testing.T) {
	g := altarGame(t)
	p := g.AddPlayer(SpawnOptions{Name: "P", Level: g.Prog.TopMortalLevel(), Swords: 1})
	for i, want := range []float64{100, 300, 600, 1000, 1500} {
		if got := g.Prog.TienEntry(i); got != want {
			t.Fatalf("mốc Tiên giới %d = %v, muốn %v", i, got, want)
		}
	}
	g.GainTienKhi(p, 99)
	if g.TienRealmOf(p) != -1 {
		t.Fatal("99 Tiên khí đã lên Chân Tiên, phải cần 100")
	}
	g.GainTienKhi(p, 1)
	if g.TienRealmOf(p) != 0 {
		t.Fatalf("đủ 100 phải lên Chân Tiên, đang %d", g.TienRealmOf(p))
	}
	g.GainTienKhi(p, 1400) // tổng 1500 → thẳng lên Đạo Tổ
	if got := g.TienRealmOf(p); got != 4 {
		t.Fatalf("1500 Tiên khí phải tới Đạo Tổ (4), đang %d", got)
	}
}

// Spec §3.4 ví dụ: Kim Tiên đang 450 bị chém → Chân Tiên, Tiên khí về 100, loss = 350;
// người chém nhận 50%, 30% rơi thành viên.
func TestRobTienKhiOnHit(t *testing.T) {
	g := altarGame(t)
	victim := g.AddPlayer(SpawnOptions{Name: "V", Swords: 1})
	victim.Level = g.Prog.TienLevel(1) // Kim Tiên
	victim.TienKhi = 450
	g.refreshDerived(victim)
	robber := g.AddPlayer(SpawnOptions{Name: "R", Level: g.Prog.TopMortalLevel(), Swords: 1})
	g.robTienKhi(robber, victim)

	if got := g.TienRealmOf(victim); got != 0 {
		t.Fatalf("nạn nhân về Tiên giới %d, muốn Chân Tiên (0)", got)
	}
	if victim.TienKhi != 100 {
		t.Fatalf("Tiên khí nạn nhân = %v, muốn về mốc 100", victim.TienKhi)
	}
	loss := 350.0
	if want := loss * g.Cfg.Cultivation.TienRobShare; robber.TienKhi != want {
		t.Fatalf("người chém nhận %v, muốn %v", robber.TienKhi, want)
	}
	_, dropped := countKind(g, ItemTienKhi)
	if want := loss * g.Cfg.Cultivation.TienDropShare; math.Abs(dropped-want) > 0.5 {
		t.Fatalf("rơi %v điểm, muốn %v", dropped, want)
	}
}
