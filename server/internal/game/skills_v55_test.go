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

// Đàn em NHẶT ĐỒ HỘ ĐẠI CA (v6.48 §6.3) — đồ vào thẳng túi đại ca, không phải túi nó.
//
// Trước v6.48 đàn em bị cấm nhặt (spec v5.4). Đổi vì Đại Ca là nhân vật "nhiều tay": hai
// đàn em quét đồ quanh đại ca trong 8 giây là lý do để gọi chúng ra, chứ không chỉ để chắn
// kiếm.
func TestMinionPicksUpForBoss(t *testing.T) {
	g := skillGame(t)
	boss := put(g, "dai_ca_dep", Vec{X: 3000, Y: 3000}, 30, 20)
	giveCharge(boss)
	g.CastSkill(boss)

	var m *Player
	for _, p := range g.PlayerList() {
		if p.IsMinion {
			m = p
			break
		}
	}
	if m == nil {
		t.Fatal("không gọi được đàn em nào")
	}
	if !g.CanPickUp(m) {
		t.Fatal("đàn em không nhặt được đồ — v6.48 cho phép nhặt hộ đại ca")
	}
	// Bó kiếm ngay cạnh đàn em, cách đại ca xa hơn vòng nhặt của chính ông.
	before := boss.Ring.Count
	g.spawnItemNear("bundle", 3, m.Pos, 0)
	g.resolvePickups()

	if n, _ := countItems(g, "bundle"); n != 0 {
		t.Fatal("đàn em đứng cạnh bó kiếm mà không nhặt")
	}
	if boss.Ring.Count <= before {
		t.Errorf("đại ca có %d kiếm, trước %d — đồ đàn em nhặt phải vào túi ĐẠI CA", boss.Ring.Count, before)
	}
	if m.Ring.Count != g.minionSwords(boss) && m.Ring.Count > g.minionSwords(boss)+3 {
		t.Errorf("đàn em tự giữ đồ: %d kiếm", m.Ring.Count)
	}
}

// Mỗi đàn em mang 15% kiếm của đại ca; gọi lần nữa thì LÀM MỚI, không đẻ thêm.
func TestDanEmShareAndRecall(t *testing.T) {
	g := skillGame(t)
	boss := put(g, "dai_ca_dep", Vec{X: 3000, Y: 3000}, 30, 40)
	giveCharge(boss)
	g.CastSkill(boss)

	want := int(40 * g.Cfg.Skills.DanEmSwordShare)
	n := 0
	for _, p := range g.PlayerList() {
		if !p.IsMinion {
			continue
		}
		n++
		if p.Ring.Count != want {
			t.Errorf("đàn em có %d kiếm, muốn %d (15%% của 40)", p.Ring.Count, want)
		}
	}
	if n != g.Cfg.Skills.DanEmCount {
		t.Fatalf("gọi ra %d đàn em, muốn %d", n, g.Cfg.Skills.DanEmCount)
	}

	step(g, g.Cfg.Skills.DanEmSecs/2)
	giveCharge(boss)
	boss.SkillReadyAt = 0
	g.CastSkill(boss)
	again := 0
	for _, p := range g.PlayerList() {
		if p.IsMinion {
			again++
			if p.MinionUntil < g.Time+g.Cfg.Skills.DanEmSecs-0.01 {
				t.Error("gọi lần hai mà không làm mới thời gian của đàn em")
			}
		}
	}
	if again != g.Cfg.Skills.DanEmCount {
		t.Fatalf("gọi lần hai ra %d đàn em, trần là %d", again, g.Cfg.Skills.DanEmCount)
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
		{"Rụt Cổ", sc.ShellSecs, 2}, // v6.51: 4s → 2s, trùng thời gian khiêu khích
		{"Gọi Đàn Em", sc.DanEmSecs, 8},
		{"Phun Trân Châu bịt mắt", sc.TranChauSecs, 3},           // v6.48
		{"Phun Trân Châu đảo điều khiển", sc.TranChauInvertSecs, 3}, // v6.48
		{"Hoá Khổng Lồ", sc.GiantSecs, 2},                        // v6.48
		{"Gom Đơn Hỏa Tốc", g.skillNum("gom_don_secs"), 4},
		{"Ru Ngủ Trưa ngủ", g.skillNum("ru_ngu_secs"), 2.5},
		{"Ru Ngủ Trưa hát", g.skillNum("ru_ngu_channel"), 0.5},
		{"Gom Đơn trần tốc", g.Cfg.Speed.CapGomDon, 5}, // v6.45
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
