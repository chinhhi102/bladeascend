package game

import (
	"math"
	"testing"

	"kiemxoay/server/internal/config"
	"kiemxoay/server/internal/world"
)

func skillGame(t *testing.T) *Game {
	t.Helper()
	cfg := config.MustLoadDefault()
	cfg.Room.MonsterCount, cfg.Items.CountPerRoom = 0, 0
	return New(cfg, world.EmptyMap(cfg.Room.MapSize), 1)
}

// put đặt một nhân vật ở đúng chỗ, bất tử tắt, để test điều khiển hoàn toàn.
func put(g *Game, char string, at Vec, level, swords int) *Player {
	p := g.AddPlayer(SpawnOptions{Name: char, Char: char, Level: level, Swords: swords})
	p.Pos = at
	p.InvulnUntil = 0
	g.refreshDerived(p)
	return p
}

// Mỗi nhân vật trong config phải có đúng một kỹ năng (spec §6.3). Thiếu một cái là nhân vật
// đó cầm Bí Kíp mà bấm không ra gì.
func TestEveryCharacterHasASkill(t *testing.T) {
	g := skillGame(t)
	for _, ch := range g.Cfg.Characters {
		sk := SkillOf(ch)
		if sk == nil {
			t.Errorf("nhân vật %q chưa có kỹ năng", ch)
			continue
		}
		if sk.Name == "" || sk.Line == "" || sk.cast == nil {
			t.Errorf("kỹ năng của %q thiếu tên, thoại hoặc phần thi hành", ch)
		}
	}
	for ch := range Skills {
		found := false
		for _, c := range g.Cfg.Characters {
			found = found || c == ch
		}
		if !found {
			t.Errorf("kỹ năng gắn cho nhân vật %q nhưng config không có nhân vật đó", ch)
		}
	}
}

// Không có Bí Kíp thì không bấm được; nhặt Bí Kíp mới có lượt, tối đa 2 (spec §6.3).
func TestSkillNeedsScrollAndCapsAtTwo(t *testing.T) {
	g := skillGame(t)
	p := put(g, "cu_rua", Vec{X: 3000, Y: 3000}, 10, 20)
	if g.CastSkill(p) {
		t.Fatal("chưa có Bí Kíp mà vẫn dùng được kỹ năng")
	}
	for i := 0; i < 5; i++ {
		g.AddSkillCharge(p)
	}
	if p.SkillCharges != g.Cfg.Skills.MaxCharges {
		t.Fatalf("giữ %d lượt, trần là %d", p.SkillCharges, g.Cfg.Skills.MaxCharges)
	}
	if !g.CastSkill(p) {
		t.Fatal("có lượt mà vẫn không dùng được")
	}
	// Khoá 3 giây giữa hai lần dùng, dù còn lượt.
	if g.CastSkill(p) {
		t.Fatal("bấm liên tiếp được — thiếu khoá 3 giây giữa hai lần dùng")
	}
	g.Time += g.Cfg.Skills.LockSecs + 0.01
	if !g.CastSkill(p) {
		t.Fatal("hết khoá rồi mà vẫn không dùng được")
	}
}

// Chết là mất hết lượt (spec §6.3).
func TestDeathClearsCharges(t *testing.T) {
	g := skillGame(t)
	p := put(g, "be_ti", Vec{X: 3000, Y: 3000}, 10, 20)
	g.AddSkillCharge(p)
	g.ApplyStatus(p, StSlow, 5)
	g.DebugKill(p)
	if p.SkillCharges != 0 {
		t.Fatalf("chết rồi mà còn %d lượt", p.SkillCharges)
	}
	if p.HasStatus(StSlow, g.Time) {
		t.Fatal("chết rồi mà còn dính trạng thái — hồi sinh xong vẫn bị chậm")
	}
}

// Loa Phường đẩy lùi và làm vòng kiếm địch đổi chiều, CHỈ với người đứng trong nón.
func TestLoaPhuongPushesOnlyInsideCone(t *testing.T) {
	g := skillGame(t)
	a := put(g, "ba_tam_loa", Vec{X: 3000, Y: 3000}, 10, 20)
	a.Facing = 0 // nhìn sang phải
	front := put(g, "cu_rua", Vec{X: 3000 + 300, Y: 3000}, 10, 20)
	back := put(g, "cu_rua", Vec{X: 3000 - 300, Y: 3000}, 10, 20)
	dirBefore := back.Ring.Dir

	g.AddSkillCharge(a)
	if !g.CastSkill(a) {
		t.Fatal("không dùng được Loa Phường")
	}
	if front.KnockUntil <= g.Time {
		t.Fatal("người đứng trước mặt không bị đẩy")
	}
	if back.KnockUntil > g.Time || back.Ring.Dir != dirBefore {
		t.Fatal("người đứng SAU LƯNG cũng trúng — nón 60° đang sai")
	}
}

// Ru Ngủ làm chậm cả chân lẫn vòng kiếm (spec §6.3).
func TestRuNguSlowsFeetAndRing(t *testing.T) {
	g := skillGame(t)
	a := put(g, "ong_ba_vong", Vec{X: 3000, Y: 3000}, 10, 20)
	b := put(g, "cu_rua", Vec{X: 3100, Y: 3000}, 10, 20)
	fastMove, fastSpin := g.MoveSpeed(b), g.Omega(b)

	g.AddSkillCharge(a)
	g.CastSkill(a)

	if g.MoveSpeed(b) >= fastMove {
		t.Fatal("trúng Ru Ngủ mà chân không chậm")
	}
	if g.Omega(b) >= fastSpin {
		t.Fatal("trúng Ru Ngủ mà vòng kiếm không chậm")
	}
	g.Time += g.Cfg.Skills.RuNguSecs + 0.01
	if g.MoveSpeed(b) != fastMove || g.Omega(b) != fastSpin {
		t.Fatal("hết giờ rồi mà vẫn còn chậm")
	}
}

// Lưới Mì trói chết chân nhưng kiếm vẫn quay (spec §6.3).
func TestLuoiMiRootsButRingKeepsSpinning(t *testing.T) {
	g := skillGame(t)
	a := put(g, "su_phu_mi", Vec{X: 3000, Y: 3000}, 10, 20)
	a.Facing = 0
	b := put(g, "cu_rua", Vec{X: 3000 + g.Cfg.Skills.LuoiMiRange, Y: 3000}, 10, 20)
	spin := g.Omega(b)

	g.AddSkillCharge(a)
	g.CastSkill(a)

	if g.MoveSpeed(b) != 0 {
		t.Fatalf("bị trói mà vẫn đi được %.1f", g.MoveSpeed(b))
	}
	if g.Omega(b) != spin {
		t.Fatal("bị trói mà vòng kiếm cũng dừng — spec nói kiếm vẫn quay")
	}
}

// Gồng Cơ miễn khống chế và gỡ luôn thứ đang dính (spec §6.3).
func TestGongCoBlocksControl(t *testing.T) {
	g := skillGame(t)
	p := put(g, "be_ti", Vec{X: 3000, Y: 3000}, 10, 20)
	g.ApplyStatus(p, StRoot, 5)
	g.AddSkillCharge(p)
	g.CastSkill(p)
	if p.HasStatus(StRoot, g.Time) {
		t.Fatal("gồng rồi mà vẫn bị trói")
	}
	g.ApplyStatus(p, StSlow, 5)
	if p.HasStatus(StSlow, g.Time) {
		t.Fatal("đang gồng mà vẫn dính khống chế mới")
	}
	// Độ bền ×2, cộng dồn với Cương Kiếm.
	if m := g.hardMult(p, "durability"); math.Abs(m-g.Cfg.Skills.BraceDurability) > 1e-9 {
		t.Fatalf("độ bền khi gồng ×%.2f, muốn ×%.2f", m, g.Cfg.Skills.BraceDurability)
	}
	g.ApplyBuff(p, "hard")
	want := g.Cfg.Skills.BraceDurability * g.Cfg.Buffs["hard"].DurabilityMult
	if m := g.hardMult(p, "durability"); math.Abs(m-want) > 1e-9 {
		t.Fatalf("gồng + Cương Kiếm ×%.2f, muốn ×%.2f (phải cộng dồn)", m, want)
	}
}

// Rụt Cổ: không bị bể kiếm, không bị chém thân, và ra khỏi mai thì bật sóng đẩy.
func TestRutCoProtectsThenBursts(t *testing.T) {
	g := skillGame(t)
	p := put(g, "cu_rua", Vec{X: 3000, Y: 3000}, 10, 30)
	other := put(g, "be_ti", Vec{X: 3000 + 120, Y: 3000}, 10, 20)
	g.AddSkillCharge(p)
	g.CastSkill(p)

	before := p.Ring.Count
	g.breakSwords(p, 5)
	if p.Ring.Count != before {
		t.Fatalf("đang rụt mai mà vẫn bể %d kiếm", before-p.Ring.Count)
	}
	// Ra khỏi mai → sóng đẩy người đứng cạnh.
	g.Time += g.Cfg.Skills.ShellSecs + 0.01
	g.updateSkillState(1.0 / 30)
	if other.KnockUntil <= g.Time {
		t.Fatal("ra khỏi mai mà không có sóng đẩy")
	}
}

// Bánh Bao Thần chỉ hồi phần kiếm bể TRONG CỬA SỔ 10 giây, không hồi cả trận (spec §6.3).
func TestBanhBaoOnlyRestoresRecentLosses(t *testing.T) {
	g := skillGame(t)
	// Để số kiếm thấp hơn trần của cảnh giới, nếu không AddSwords bị chặn và phép đo thành
	// vô nghĩa — không phải lỗi kỹ năng mà là test tự đặt sai.
	p := put(g, "a_hon", Vec{X: 3000, Y: 3000}, 30, 10)

	g.noteSwordLoss(p, 100)                                 // bể từ lâu
	g.Time += g.Cfg.Skills.BanhBaoWindowSecs + 1            // ... quá cửa sổ
	g.noteSwordLoss(p, 20)                                  // bể gần đây
	before := p.Ring.Count
	g.AddSkillCharge(p)
	g.CastSkill(p)

	want := int(20 * g.Cfg.Skills.BanhBaoShare)
	if got := p.Ring.Count - before; got != want {
		t.Fatalf("hồi %d kiếm, muốn %d — đang tính cả phần bể quá lâu", got, want)
	}
}

// Đàn em của Đại Ca không chém thân được (spec §6.3) — chúng chỉ mài kiếm.
func TestMinionsCannotBodyHit(t *testing.T) {
	g := skillGame(t)
	boss := put(g, "dai_ca_dep", Vec{X: 3000, Y: 3000}, 20, 30)
	victim := put(g, "cu_rua", Vec{X: 3000 + 60, Y: 3000}, 0, 1)
	g.AddSkillCharge(boss)
	g.CastSkill(boss)

	minions := 0
	for _, p := range g.PlayerList() {
		if p.IsMinion {
			minions++
			p.Pos = victim.Pos // ép đứng chồng lên nạn nhân
		}
	}
	if minions != g.Cfg.Skills.DanEmCount {
		t.Fatalf("gọi ra %d đàn em, muốn %d", minions, g.Cfg.Skills.DanEmCount)
	}
	boss.Pos = Vec{X: 9000, Y: 9000} // đại ca đi chỗ khác để không tự chém
	lvl := victim.Level
	for i := 0; i < 30; i++ {
		g.resolveSwordBody()
	}
	if victim.Level < lvl {
		t.Fatal("đàn em chém rớt cảnh giới nạn nhân — chúng không được chém thân")
	}
}

// Đàn em hết giờ thì biến mất, không ở lại cả trận.
func TestMinionsExpire(t *testing.T) {
	g := skillGame(t)
	boss := put(g, "dai_ca_dep", Vec{X: 3000, Y: 3000}, 20, 30)
	g.AddSkillCharge(boss)
	g.CastSkill(boss)
	g.Time += g.Cfg.Skills.DanEmSecs + 0.1
	g.updateSkillState(1.0 / 30)
	for _, p := range g.PlayerList() {
		if p.IsMinion {
			t.Fatal("đàn em hết giờ mà vẫn còn trên bản đồ")
		}
	}
}

// Giao Hỏa Tốc đi xuyên vật cản PHÁ ĐƯỢC nhưng không xuyên tường không phá được.
func TestGhostPassesBreakableOnly(t *testing.T) {
	cfg := config.MustLoadDefault()
	cfg.Room.MonsterCount, cfg.Items.CountPerRoom = 0, 0
	m := world.EmptyMap(cfg.Room.MapSize)
	m.Obstacles = []world.Obstacle{
		{ID: 0, Kind: "brick", MaxHP: 250, Shape: world.Shape{X: 3200, Y: 2900, W: 200, H: 200}},
		{ID: 1, Kind: "stone_wall", MaxHP: 0, Shape: world.Shape{X: 4200, Y: 2900, W: 200, H: 200}},
	}
	g := New(cfg, m, 1)
	p := put(g, "shipper", Vec{X: 3000, Y: 3000}, 10, 20)
	g.AddSkillCharge(p)
	g.CastSkill(p)

	inBrick := g.resolveBlockingFor(p, Vec{X: 3300, Y: 3000}, g.BodyR(p))
	if inBrick.Dist(Vec{X: 3300, Y: 3000}) > 1e-9 {
		t.Fatal("đang Giao Hỏa Tốc mà vẫn bị tường gạch (phá được) chặn")
	}
	inWall := g.resolveBlockingFor(p, Vec{X: 4300, Y: 3000}, g.BodyR(p))
	if inWall.Dist(Vec{X: 4300, Y: 3000}) < 1e-9 {
		t.Fatal("xuyên qua cả tường KHÔNG phá được — spec cấm")
	}
}

// Trúng Phun Trân Châu thì bị bịt mắt và server báo riêng cho người đó.
func TestTranChauBlinds(t *testing.T) {
	g := skillGame(t)
	a := put(g, "map_tra_sua", Vec{X: 3000, Y: 3000}, 10, 20)
	a.Facing = 0
	b := put(g, "cu_rua", Vec{X: 3000 + 400, Y: 3000}, 10, 20)
	g.DrainEvents()

	g.AddSkillCharge(a)
	g.CastSkill(a)

	if !b.HasStatus(StBlind, g.Time) {
		t.Fatal("trúng trân châu mà không bị bịt mắt")
	}
	found := false
	for _, e := range g.DrainEvents() {
		if e.Kind == "Blinded" && e.A == b.ID && e.To == b.ID {
			found = true
			if math.Abs(float64(e.F)-g.Cfg.Skills.TranChauSecs) > 1e-6 {
				t.Fatalf("báo bịt mắt %.2f giây, config nói %.2f", e.F, g.Cfg.Skills.TranChauSecs)
			}
		}
	}
	if !found {
		t.Fatal("không có event Blinded gửi riêng cho người bị trúng")
	}
}

// Vệt trái cây của Gánh Lướt làm người đi vào bị trượt theo quán tính.
func TestGanhLuotLeavesSlickTrail(t *testing.T) {
	g := skillGame(t)
	a := put(g, "co_ba_ganh", Vec{X: 3000, Y: 3000}, 10, 20)
	a.Facing = 0
	g.AddSkillCharge(a)
	g.CastSkill(a)
	if len(g.slicks) != g.Cfg.Skills.SlickCount {
		t.Fatalf("rơi %d vũng, muốn %d", len(g.slicks), g.Cfg.Skills.SlickCount)
	}
	b := put(g, "cu_rua", g.slicks[0].at, 10, 20)
	g.updateSkillState(1.0 / 30)
	if !b.HasStatus(StSlip, g.Time) {
		t.Fatal("đi vào vệt trái cây mà không trượt")
	}
}
