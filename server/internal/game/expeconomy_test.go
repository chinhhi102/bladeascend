package game

import (
	"math"
	"testing"
)

// §6.0 luật 1: viên EXP môi trường = giá trị gốc × orb_value_mult^r của NGƯỜI NHẶT.
func TestAmbientOrbScalesWithPickerRealm(t *testing.T) {
	for _, realm := range []int{0, 5} {
		g := quietGame(t, nil)
		p := g.AddPlayer(SpawnOptions{Name: "A", Char: g.Cfg.Characters[0], Level: realm * g.Prog.Tiers(), Swords: 1})
		before := p.TotalExp
		g.pickup(p, g.addItem("orb", 20, p.Pos)) // Src mặc định = viên môi trường
		want := 20 * math.Pow(g.Cfg.ExpEconomy.OrbValueMult, float64(realm))
		if got := p.TotalExp - before; math.Abs(got-want) > 1e-6 {
			t.Errorf("cảnh giới %d: viên 20 cho %.2f EXP, muốn %.2f", realm, got, want)
		}
	}
}

// §6.0 luật 4: quái rơi tổng EXP gốc × hệ số^r của người chém, viên KHÔNG tự nhân lần nữa.
func TestMonsterExpScalesWithKillerRealmOnce(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "A", Char: g.Cfg.Characters[0], Level: 3 * g.Prog.Tiers(), Swords: 1})
	m := &Monster{ID: g.newID(), Kind: "small", Exp: 10, Pos: p.Pos.Add(Vec{X: 50})}
	g.Monsters[m.ID] = m
	before := p.TotalExp
	g.killMonster(m, p)
	for _, it := range g.Items {
		if it.Kind == "orb" {
			g.pickup(p, it)
		}
	}
	want := 10 * math.Pow(g.Cfg.ExpEconomy.OrbValueMult, 3)
	if got := p.TotalExp - before; math.Abs(got-math.Round(want)) > 1 {
		t.Fatalf("quái 10 EXP ở Nguyên Anh cho %.1f, muốn ≈ %.1f", got, want)
	}
}

// §6.0 luật 2 + 3, §4.0: bán kính hút = 90 × (1 + nội tại + (√scale − 1) + Nam Châm) — cộng.
func TestPickupRadiusSqrtScaleAndMagnet(t *testing.T) {
	g := quietGame(t, nil)
	ch := g.Cfg.Characters[0]
	p := g.AddPlayer(SpawnOptions{Name: "A", Char: ch, Level: 80, Swords: 1})
	inn := g.Cfg.Innate[ch].Pickup
	scaleTerm := math.Pow(p.Scale, g.Cfg.ExpEconomy.PickupScaleExp) - 1
	base := g.Cfg.Player.PickupR * (1 + inn + scaleTerm)
	if got := g.PickupRadius(p); math.Abs(got-base) > 1e-6 {
		t.Fatalf("bán kính hút %.1f, muốn %.1f (scale %.2f)", got, base, p.Scale)
	}
	g.ApplyBuff(p, "magnet")
	want := g.Cfg.Player.PickupR * (1 + inn + scaleTerm + (g.Cfg.Buffs["magnet"].Mult - 1))
	if got := g.PickupRadius(p); math.Abs(got-want) > 1e-6 {
		t.Fatalf("Nam Châm: %.1f, muốn %.1f (cộng trên base, không nhân dồn)", got, want)
	}
}

// §6.0 luật 5: chém người thật nhận attacker_share (50 %), chém bot nhận bot_steal_share (25 %).
func TestStealShareHumanVsBot(t *testing.T) {
	for _, isBot := range []bool{false, true} {
		g := quietGame(t, nil)
		a := g.AddPlayer(SpawnOptions{Name: "A", Char: g.Cfg.Characters[0], Swords: 1})
		b := g.AddPlayer(SpawnOptions{Name: "B", Char: g.Cfg.Characters[1], IsBot: isBot, Level: 2 * g.Prog.Tiers(), Swords: 1})
		b.InvulnUntil = 0
		victimExp := b.TotalExp
		before := a.TotalExp
		g.applyBodyHit(a, b)
		want := victimExp*g.Cfg.Player.AttackerShare + g.Cfg.Player.KillBonusExp
		if isBot {
			want = victimExp*g.Cfg.ExpEconomy.BotStealShare + g.Cfg.Player.KillBonusExp
		}
		if got := a.TotalExp - before; math.Abs(got-want) > 1e-6 {
			t.Errorf("bot=%v: người chém nhận %.0f, muốn %.0f", isBot, got, want)
		}
	}
}
