package game

import "testing"

// Các luật kỹ năng mới của spec v5.4 / v5.5 (§6.3).

// Đàn em mang 50% số kiếm HIỆN TẠI của đại ca, tối thiểu 1, và đại ca không bị trừ kiếm.
func TestMinionSwordsHalfOfBoss(t *testing.T) {
	g := skillGame(t)
	boss := put(g, "dai_ca_dep", Vec{X: 3000, Y: 3000}, 30, 40)
	bossSwords := boss.Ring.Count

	giveCharge(boss)
	g.CastSkill(boss)

	want := int(float64(bossSwords) * g.Cfg.Skills.DanEmSwordShare)
	n := 0
	for _, p := range g.PlayerList() {
		if !p.IsMinion || p.MinionOf != boss.ID {
			continue
		}
		n++
		if p.Ring.Count != want {
			t.Errorf("đàn em có %d kiếm, muốn %d (50%% của %d)", p.Ring.Count, want, bossSwords)
		}
		if p.Level != boss.Level {
			t.Errorf("đàn em cảnh giới %d, đại ca %d — spec nói cùng cảnh giới", p.Level, boss.Level)
		}
	}
	if n != g.Cfg.Skills.DanEmCount {
		t.Errorf("gọi được %d đàn em, config nói %d", n, g.Cfg.Skills.DanEmCount)
	}
	if boss.Ring.Count != bossSwords {
		t.Errorf("đại ca còn %d kiếm, trước khi gọi có %d — spec nói không bị trừ",
			boss.Ring.Count, bossSwords)
	}
}

// Đại ca chỉ còn 1 kiếm thì đàn em vẫn có 1, không phải 0 (làm tròn xuống ra 0).
func TestMinionSwordsAtLeastOne(t *testing.T) {
	g := skillGame(t)
	boss := put(g, "dai_ca_dep", Vec{X: 3000, Y: 3000}, 30, 1)
	giveCharge(boss)
	g.CastSkill(boss)

	for _, p := range g.PlayerList() {
		if p.IsMinion && p.Ring.Count < 1 {
			t.Fatalf("đàn em có %d kiếm — tối thiểu phải là 1, nếu không nó vô hại hoàn toàn",
				p.Ring.Count)
		}
	}
}

// Đàn em không nhặt đồ và không qua cổng (spec v5.4).
func TestMinionCannotPickUp(t *testing.T) {
	g := skillGame(t)
	boss := put(g, "dai_ca_dep", Vec{X: 3000, Y: 3000}, 30, 20)
	giveCharge(boss)
	g.CastSkill(boss)

	found := false
	for _, p := range g.PlayerList() {
		if !p.IsMinion {
			continue
		}
		found = true
		if g.CanPickUp(p) {
			t.Error("đàn em nhặt được đồ — spec cấm")
		}
	}
	if !found {
		t.Fatal("không gọi được đàn em nào")
	}
}

// Trúng Phun Trân Châu thì KHÔNG nhặt được bất cứ thứ gì trong suốt thời gian bị dính.
func TestBlindBlocksEveryPickup(t *testing.T) {
	g := skillGame(t)
	p := put(g, "cu_rua", Vec{X: 3000, Y: 3000}, 10, 20)
	if !g.CanPickUp(p) {
		t.Fatal("chưa bị gì mà đã không nhặt được")
	}
	g.ApplyStatus(p, StBlind, g.Cfg.Skills.TranChauSecs)
	if g.CanPickUp(p) {
		t.Error("đang bị bịt mắt mà vẫn nhặt được — spec v5.5 nói mọi thứ đi xuyên qua")
	}
	g.Time += g.Cfg.Skills.TranChauSecs + 0.01
	if !g.CanPickUp(p) {
		t.Error("hết bịt mắt rồi mà vẫn không nhặt được")
	}
}

// Vệt trái cây của Cô Ba gây CHOÁNG: không đi, không dùng kỹ năng, không Thiêu Đốt.
func TestSlickStuns(t *testing.T) {
	g := skillGame(t)
	co := put(g, "co_ba_ganh", Vec{X: 3000, Y: 3000}, 10, 20)
	co.Facing = 0
	victim := put(g, "cu_rua", Vec{X: 3000 + 200, Y: 3000}, 10, 20)
	giveCharge(co)
	g.CastSkill(co)
	g.updateSkillState(1.0 / 30)

	if !victim.HasStatus(StStun, g.Time) {
		t.Fatal("đi vào vệt trái cây mà không bị choáng (spec v5.5)")
	}
	if g.MoveSpeed(victim) != 0 {
		t.Errorf("đang choáng mà tốc độ %.1f — phải là 0", g.MoveSpeed(victim))
	}
	giveCharge(victim)
	if g.CanCastSkill(victim) {
		t.Error("đang choáng mà vẫn dùng được kỹ năng")
	}
	victim.Moving, victim.Burning = true, true
	g.GainExp(victim, g.Prog.ExpNeeded(victim.Level)*0.9, false)
	if g.Burning(victim) {
		t.Error("đang choáng mà vẫn Thiêu Đốt được")
	}
}

// Gồng Cơ miễn choáng — spec nói "không bị đẩy lùi/chậm/trói", và choáng cùng nhóm đó.
func TestBraceImmuneToStun(t *testing.T) {
	g := skillGame(t)
	p := put(g, "be_ti", Vec{X: 3000, Y: 3000}, 10, 20)
	giveCharge(p)
	g.CastSkill(p)
	g.ApplyStatus(p, StStun, 1)
	if p.HasStatus(StStun, g.Time) {
		t.Fatal("đang Gồng Cơ mà vẫn bị choáng")
	}
}

// Các con số thời gian của spec v5.4/v5.5 phải đúng trong config — đây là chỗ hay quên
// nhất khi spec đổi, và sai thì không có gì báo.
func TestSkillDurationsMatchSpec(t *testing.T) {
	g := skillGame(t)
	sc := g.Cfg.Skills
	for _, c := range []struct {
		name string
		got  float64
		want float64
	}{
		{"Gồng Cơ", sc.BraceSecs, 4},
		{"Rụt Cổ", sc.ShellSecs, 5},
		{"Gọi Đàn Em", sc.DanEmSecs, 8},
		{"Phun Trân Châu bịt mắt", sc.TranChauSecs, 5},
		{"Gom Đơn Hỏa Tốc", g.skillNum("gom_don_secs"), 4},
		{"Ru Ngủ Trưa ngủ", g.skillNum("ru_ngu_secs"), 2.5},
		{"Ru Ngủ Trưa hát", g.skillNum("ru_ngu_channel"), 0.5},
		{"Gom Đơn trần tốc", g.Cfg.Speed.CapGomDon, 3},
	} {
		if c.got != c.want {
			t.Errorf("%s = %g, spec nói %g", c.name, c.got, c.want)
		}
	}
	if sc.DanEmCount != 2 {
		t.Errorf("số đàn em = %d, spec nói 2", sc.DanEmCount)
	}
	// Phạm vi KHÔNG còn là số tuyệt đối từ v6.5 — chúng là hệ số × U của người dùng, và
	// `TestSkillRangesMatchSpec` khoá từng hệ số một.
}
