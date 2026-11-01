package game

import (
	"math"
	"testing"
)

// Spec v6.26 §5.1 / §3.4 và v6.27 §6.3 — các dòng mới ở checklist §17.

// orbTotal: tổng Linh khí và số viên đang nằm trên đất.
func orbTotal(g *Game) (sum float64, n int) {
	for _, it := range g.Items {
		if it.Kind == "orb" {
			sum += it.Value
			n++
		}
	}
	return sum, n
}

func bundleSwords(g *Game) int {
	n := 0
	for _, it := range g.Items {
		if it.Kind == "bundle" {
			n += int(it.Value)
		}
	}
	return n
}

// §17: chém trúng thân người Kim Đan (còn 1 vòng) → chết ngay; người chém nhận 30 % Linh khí
// nạn nhân nhưng không lên quá 2 đại cảnh giới; xác rơi 30 % Linh khí + 50 % kiếm.
func TestBodyHitKillsAnyRealm(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Char: "a_hon", Level: 3 * g.Prog.Tiers(), Swords: 10}) // Nguyên Anh
	b := g.AddPlayer(SpawnOptions{Name: "B", Char: "be_ti", Level: 2*g.Prog.Tiers() + 5, Swords: 10})
	b.InvulnUntil = 0
	victimExp, before := b.TotalExp, a.TotalExp
	g.applyBodyHit(a, b)
	if !b.Dead {
		t.Fatal("Kim Đan bị chém trúng thân mà không chết")
	}
	want := victimExp*g.Cfg.Player.AttackerShare + g.Cfg.Player.KillBonusExp
	if got := a.TotalExp - before; math.Abs(got-want) > 1e-6 {
		t.Errorf("người chém nhận %.0f, muốn %.0f (30%% + thưởng hạ)", got, want)
	}
	sum, n := orbTotal(g)
	if math.Abs(sum-math.Round(victimExp*g.Cfg.Player.DropShare)) > 1 || n > g.Cfg.Player.DeathOrbsMax {
		t.Errorf("xác rơi %.0f Linh khí / %d viên, muốn %.0f / ≤ %d", sum, n, victimExp*0.3, g.Cfg.Player.DeathOrbsMax)
	}
	if got := bundleSwords(g); got != 5 {
		t.Errorf("xác rơi %d kiếm, muốn 5 (50%% của 10)", got)
	}
}

// Trần +2 đại cảnh giới: Luyện Khí chém chết Đại Thừa không nhảy thẳng lên Đại Thừa.
func TestKillRewardCappedAtTwoRealms(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Char: "a_hon", Swords: 10})
	b := g.AddPlayer(SpawnOptions{Name: "B", Char: "be_ti", Level: 8 * g.Prog.Tiers(), Swords: 10})
	b.InvulnUntil = 0
	g.applyBodyHit(a, b)
	floor := g.Prog.Threshold(2 * g.Prog.Tiers()) // sàn Kim Đan
	if a.TotalExp > floor+g.Cfg.Player.KillBonusExp+1e-6 {
		t.Fatalf("Luyện Khí nhận tới %.0f Linh khí, trần là sàn Kim Đan %.0f (+ thưởng hạ)", a.TotalExp, floor)
	}
	if r := g.Prog.Realm(a.Level); r > 2 {
		t.Fatalf("lên tới cảnh giới %d, trần +2", r)
	}
	if _, n := orbTotal(g); n > g.Cfg.Player.DeathOrbsMax {
		t.Fatalf("xác Đại Thừa rơi %d viên, trần %d", n, g.Cfg.Player.DeathOrbsMax)
	}
}

// §17: người có Kim Cang bị chém → không chết, khiên vỡ.
func TestShieldSavesFromDeath(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Char: "a_hon", Swords: 10})
	b := g.AddPlayer(SpawnOptions{Name: "B", Char: "be_ti", Level: 30, Swords: 10})
	b.InvulnUntil = 0
	g.ApplyBuff(b, "shield")
	g.applyBodyHit(a, b)
	if b.Dead || b.HasBuff(BuffShield, g.Time) || !b.Invulnerable(g.Time) {
		t.Fatalf("khiên: chết=%v, còn khiên=%v, bất tử=%v — muốn sống, vỡ khiên, bất tử 1 s",
			b.Dead, b.HasBuff(BuffShield, g.Time), b.Invulnerable(g.Time))
	}
}

// §17: Đạo Tổ bị chém → chết, rơi 30 % Tiên khí; người chém Tiên giới nhận 30 % nhưng không
// lên quá 1 cảnh giới Tiên.
func TestDaoToKilledDropsAndCapsTien(t *testing.T) {
	g := rulesGame(t)
	victim := tienPlayer(t, g, 15000)
	att := tienPlayer(t, g, 1) // Chân Tiên, trần +1 = Kim Tiên (3 000)
	g.applyBodyHit(att, victim)
	if !victim.Dead {
		t.Fatal("Đạo Tổ bị chém mà không chết")
	}
	if got := totalTienOrbs(g); got != 4500 {
		t.Errorf("rơi %d Tiên khí, muốn 4 500 (30%% của 15 000)", got)
	}
	if countTienOrbs(g) > 40 {
		t.Errorf("rơi %d viên Tiên khí, tối đa 40", countTienOrbs(g))
	}
	if att.TienKhi != g.Prog.TienCeil(1) {
		t.Errorf("người chém có %.0f Tiên khí, muốn kẹp ở trần Kim Tiên %.0f (30%% = 4 500 vượt trần)",
			att.TienKhi, g.Prog.TienCeil(1))
	}
}

// Người chém là phàm: không nhận Tiên khí, nhưng xác vẫn rơi đủ 30 %.
func TestMortalKillerStillDropsTien(t *testing.T) {
	g := rulesGame(t)
	victim := tienPlayer(t, g, 800)
	att := g.AddPlayer(SpawnOptions{Name: "A", Char: g.Cfg.Characters[1], Swords: 10})
	g.applyBodyHit(att, victim)
	if att.TienKhi != 0 || totalTienOrbs(g) != 240 {
		t.Fatalf("phàm nhận %.0f Tiên khí, rơi %d — muốn 0 / 240", att.TienKhi, totalTienOrbs(g))
	}
}

// v6.27 §17: người bị Ru Ngủ — không đi được nhưng vòng kiếm vẫn quay, va kiếm hai chiều.
func TestSleepDoesNotDisableRing(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "P", Char: "a_hon", Level: 20, Swords: 30})
	awake := g.Omega(p)
	g.ApplyStatus(p, StSleep, 2.5)
	if g.Omega(p) != awake || g.MoveSpeed(p) != 0 {
		t.Fatalf("ngủ: omega %.2f (muốn %.2f), tốc %.0f (muốn 0)", g.Omega(p), awake, g.MoveSpeed(p))
	}
}

// v6.27 §17: người bị bịt mắt — kiếm chạm thân người khác không gây chết.
func TestBlindedSwordCannotKill(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Char: "a_hon", Level: 20, Swords: 30})
	b := g.AddPlayer(SpawnOptions{Name: "B", Char: "be_ti", Swords: 1})
	a.InvulnUntil, b.InvulnUntil = 0, 0
	b.Pos = a.Pos.Add(Vec{X: a.Geom.Reach * 0.7})
	g.ApplyStatus(a, StBlind, 5)
	for i := 0; i < 30; i++ {
		g.Tick(1.0 / 30)
	}
	if b.Dead {
		t.Fatal("kiếm người mù chém chết người khác")
	}
	a.Status[StBlind] = 0
	for i := 0; i < 30 && !b.Dead; i++ {
		g.Tick(1.0 / 30)
	}
	if !b.Dead {
		t.Fatal("hết mù mà kiếm vẫn không chém được — bài kiểm đặt sai chỗ")
	}
}
