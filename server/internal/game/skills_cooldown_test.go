package game

import "testing"

// Spec v6.25 §6.3: kỹ năng tự hồi + tích lượt theo nhân vật.

// stepFor chạy rechargeSkills đúng `secs` giây game (không chạy cả Tick cho nhanh và gọn).
func stepFor(g *Game, secs float64) {
	const dt = 1.0 / 30
	for t := 0.0; t < secs-1e-9; t += dt {
		g.Time += dt
		g.rechargeSkills(dt)
	}
}

// Vào trận có sẵn 1 lượt; hồi lần lượt theo bảng; đầy thì dừng.
func TestSkillStartsWithOneAndRecharges(t *testing.T) {
	g := skillGame(t)
	p := put(g, "ba_tam_loa", Vec{X: 3000, Y: 3000}, 10, 20) // [18s, 2]
	if p.SkillCharges != 1 {
		t.Fatalf("vào trận có %d lượt, spec: 1", p.SkillCharges)
	}
	stepFor(g, 17.9)
	if p.SkillCharges != 1 {
		t.Fatalf("chưa đủ 18s đã hồi: %d lượt", p.SkillCharges)
	}
	stepFor(g, 0.2)
	if p.SkillCharges != 2 || p.SkillRecharge != 0 {
		t.Fatalf("sau 18s: %d lượt, đồng hồ %.2f — muốn 2 lượt, đồng hồ dừng", p.SkillCharges, p.SkillRecharge)
	}
	stepFor(g, 60)
	if p.SkillCharges != 2 {
		t.Fatalf("đầy rồi vẫn tích thêm: %d", p.SkillCharges)
	}
	// Dùng 1 lượt khi đang đầy → đồng hồ chạy lại từ đầu.
	if !g.CastSkill(p) || p.SkillRecharge <= 17.9 {
		t.Fatalf("dùng khi đầy: đồng hồ %.2f, muốn ~18", p.SkillRecharge)
	}
}

// Mỗi nhân vật đúng hồi chiêu + lượt tối đa trong bảng spec; Tiên giới +1, rớt về cắt bớt.
func TestSkillCooldownTableAndTien(t *testing.T) {
	g := skillGame(t)
	want := map[string][2]float64{"a_hon": {30, 2}, "ba_tam_loa": {18, 2}, "ong_ba_vong": {25, 2},
		"dai_ca_dep": {40, 1}, "map_tra_sua": {20, 2}, "co_ba_ganh": {22, 2}, "su_phu_mi": {30, 2},
		"be_ti": {35, 1}, "shipper": {20, 2}, "cu_rua": {30, 2}}
	for ch, w := range want {
		p := put(g, ch, Vec{X: 3000, Y: 3000}, 10, 1)
		if g.SkillCooldownSecs(p) != w[0] || g.MaxSkillCharges(p) != int(w[1]) {
			t.Errorf("%s: hồi %.0fs · %d lượt, spec %v", ch, g.SkillCooldownSecs(p), g.MaxSkillCharges(p), w)
		}
	}
	p := put(g, "cu_rua", Vec{X: 3000, Y: 3000}, 10, 1)
	g.ForceState(p, g.Prog.TopMortalLevel(), 10, p.Pos)
	p.TotalExp = g.Prog.MaxExp()
	g.GainTienKhi(p, 1)
	if got := g.MaxSkillCharges(p); got != 3 {
		t.Fatalf("Tiên giới: %d lượt tối đa, muốn 2 + 1", got)
	}
	g.FillSkillCharges(p)
	g.ForceState(p, g.Prog.TopMortalLevel(), 10, p.Pos) // rớt về Đại Thừa
	p.TienKhi = 0
	g.refreshDerived(p)
	if p.SkillCharges != 2 {
		t.Fatalf("rớt về Đại Thừa còn %d lượt, muốn cắt về 2", p.SkillCharges)
	}
}

// Khoá 3 s giữa hai lần dùng, dù còn lượt.
func TestSkillLockout(t *testing.T) {
	g := skillGame(t)
	p := put(g, "cu_rua", Vec{X: 3000, Y: 3000}, 10, 20)
	g.FillSkillCharges(p)
	if !g.CastSkill(p) {
		t.Fatal("có lượt mà không dùng được")
	}
	if g.CastSkill(p) {
		t.Fatal("bấm liên tiếp được — thiếu khoá 3 giây")
	}
	g.Time += g.Cfg.Skills.LockSecs + 0.01
	if !g.CastSkill(p) {
		t.Fatal("hết khoá mà vẫn không dùng được")
	}
	if g.CastSkill(p) {
		t.Fatal("hết lượt mà vẫn dùng được")
	}
}

// Chết mất hết lượt, đồng hồ dừng; hồi sinh lại có 1 lượt.
func TestSkillChargesOnDeathAndRespawn(t *testing.T) {
	g := skillGame(t)
	p := put(g, "cu_rua", Vec{X: 3000, Y: 3000}, 10, 20)
	g.FillSkillCharges(p)
	g.DebugKill(p)
	if p.SkillCharges != 0 || p.SkillRecharge != 0 {
		t.Fatalf("chết còn %d lượt, đồng hồ %.1f", p.SkillCharges, p.SkillRecharge)
	}
	g.Respawn(p, 0, 1)
	if p.SkillCharges != 1 || p.SkillRecharge <= 0 {
		t.Fatalf("hồi sinh: %d lượt, đồng hồ %.1f — muốn 1 lượt, đồng hồ chạy", p.SkillCharges, p.SkillRecharge)
	}
}
