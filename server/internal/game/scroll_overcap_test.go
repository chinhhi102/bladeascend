package game

import "testing"

// Q33 / v6.24 §6.0: nhặt Bí Kíp khi đã đủ lượt chỉ cho +20 Linh khí — ở MỌI cảnh giới.
// Bản lỗi cộng `TotalExp + 20`, tức nhân đôi tu vi mỗi cuốn (80–97 % EXP, EXP_AUDIT).
func TestScrollOvercapGivesFlatExp(t *testing.T) {
	for _, realm := range []int{0, 8} { // Luyện Khí, Đại Thừa
		g := quietGame(t, nil)
		L := realm*g.Prog.Tiers() + 3
		p := g.AddPlayer(SpawnOptions{Name: "A", Char: g.Cfg.Characters[0], Level: L, Swords: 1})
		p.SkillCharges = g.MaxSkillCharges(p)
		before := p.TotalExp
		it := g.addItem(ItemSkillScroll, 0, p.Pos)
		g.pickup(p, it)
		if got := p.TotalExp - before; got != g.Cfg.Skills.OvercapExp {
			t.Errorf("cảnh giới %d: Bí Kíp đủ lượt cho +%.0f EXP, muốn +%.0f", realm, got, g.Cfg.Skills.OvercapExp)
		}
	}
}

// Tiên giới: đủ lượt → +2 Tiên khí, không đụng Linh khí.
func TestScrollOvercapTienGivesTienKhi(t *testing.T) {
	g := quietGame(t, nil)
	top := g.Prog.TienLevel(0)
	p := g.AddPlayer(SpawnOptions{Name: "T", Char: g.Cfg.Characters[0], Swords: 1})
	p.TotalExp = g.Prog.MaxExp() // Đại Thừa 10 đầy thanh = đã sang đường Tiên khí
	g.ForceState(p, top, 1, p.Pos)
	p.SkillCharges = g.MaxSkillCharges(p)
	khi, exp := p.TienKhi, p.TotalExp
	g.pickup(p, g.addItem(ItemSkillScroll, 0, p.Pos))
	if p.TienKhi-khi != g.Cfg.Skills.OvercapTien || p.TotalExp != exp {
		t.Fatalf("Tiên giới đủ lượt: Tiên khí +%.1f, EXP +%.0f; muốn +%.0f Tiên khí, 0 EXP",
			p.TienKhi-khi, p.TotalExp-exp, g.Cfg.Skills.OvercapTien)
	}
}
