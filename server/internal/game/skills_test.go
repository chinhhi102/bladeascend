package game

import (
	"math"
	"testing"

	"bladeascend/server/internal/config"
	"bladeascend/server/internal/world"
)

func skillGame(t *testing.T) *Game {
	t.Helper()
	cfg := config.MustLoadDefault()
	cfg.Room.MonsterCount, cfg.Items.CountPerRoom = 0, 0
	return New(cfg, world.EmptyMap(cfg.Room.MapSize), 1, 0)
}

// put đặt một nhân vật ở đúng chỗ, bất tử tắt, để test điều khiển hoàn toàn.
// giveCharge thêm một lượt kỹ năng cho test (v6.25 bỏ Bí Kíp; lượt tự hồi theo thời gian,
// test không muốn chờ hồi chiêu thật).
func giveCharge(p *Player) bool {
	p.SkillCharges++
	return true
}

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

// Chết là mất hết lượt (spec §6.3).
func TestDeathClearsCharges(t *testing.T) {
	g := skillGame(t)
	p := put(g, "be_ti", Vec{X: 3000, Y: 3000}, 10, 20)
	giveCharge(p)
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

	giveCharge(a)
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

// Ru Ngủ Trưa (spec v6.27 §6.3): hát 0.5s rồi mọi địch trong 4 U ngủ gục 2.5s — đứng im,
// nhưng vòng kiếm VẪN QUAY (khống chế không khoá kiếm).
func TestRuNguSleepsAfterChannel(t *testing.T) {
	g := skillGame(t)
	a := put(g, "ong_ba_vong", Vec{X: 3000, Y: 3000}, 10, 20)
	b := put(g, "cu_rua", Vec{X: 3100, Y: 3000}, 10, 20)
	fastMove, fastSpin := g.MoveSpeed(b), g.Omega(b)

	giveCharge(a)
	g.CastSkill(a)

	// Chưa hát xong thì CHƯA ai ngủ — đây là nửa giây để địch chạy ra khỏi vòng.
	if b.HasStatus(StSleep, g.Time) {
		t.Fatal("vừa bấm đã ngủ — giai đoạn hát 0.5s không chạy")
	}
	if g.MoveSpeed(a) != 0 {
		t.Error("đang hát ru mà Ông Ba Võng vẫn đi được — spec nói đứng yên 0.5s")
	}
	g.Time += g.skillNum("ru_ngu_channel") + 1e-9
	g.updateSkillState(1.0 / 30)

	if !b.HasStatus(StSleep, g.Time) {
		t.Fatal("hát xong mà địch không ngủ")
	}
	if g.MoveSpeed(b) != 0 {
		t.Errorf("đang ngủ mà vẫn đi được %.1f", g.MoveSpeed(b))
	}
	if g.Omega(b) != fastSpin {
		t.Errorf("đang ngủ vòng kiếm quay %.2f rad/s, muốn %.2f — v6.27: kiếm vẫn quay", g.Omega(b), fastSpin)
	}
	g.Time += g.skillNum("ru_ngu_secs") + 0.01
	if g.MoveSpeed(b) != fastMove || g.Omega(b) != fastSpin {
		t.Fatal("hết giờ rồi mà vẫn còn ngủ")
	}
}

// Checklist §17: "Ông Ba Võng bị chém trong 0.5s hát → mất chiêu."
//
// Mất chiêu nghĩa là KHÔNG ai ngủ và lượt kỹ năng cũng không được hoàn — nếu hoàn thì việc
// bị chém chỉ làm chậm, và kỹ năng không còn cách nào chống lại.
func TestRuNguCancelledByBodyHit(t *testing.T) {
	g := skillGame(t)
	a := put(g, "ong_ba_vong", Vec{X: 3000, Y: 3000}, 10, 20)
	b := put(g, "cu_rua", Vec{X: 3100, Y: 3000}, 10, 20)
	// vào trận có sẵn đúng 1 lượt (v6.25)
	g.CastSkill(a)
	if a.SkillCharges != 0 {
		t.Fatalf("lượt còn %d sau khi dùng", a.SkillCharges)
	}

	g.applyBodyHit(b, a) // bị chém giữa câu hát
	g.Time += g.skillNum("ru_ngu_channel") + 1e-9
	g.updateSkillState(1.0 / 30)

	if b.HasStatus(StSleep, g.Time) {
		t.Fatal("bị chém giữa câu hát mà chiêu vẫn phát")
	}
	if a.SkillCharges != 0 {
		t.Error("mất chiêu mà được hoàn lượt — vậy thì không có rủi ro nào")
	}
}

// Hút Mì Đoạt Kiếm (spec v6.9 §6.3) — checklist §17: "địch 100 kiếm → Sư Phụ cướp 25
// (không vượt 20% tối đa của mình, dư rơi Bó Kiếm); không ai trong nón → hụt, mất lượt."
func TestHutMiStealsQuarterOfTargetSwords(t *testing.T) {
	g := skillGame(t)
	// Sư Phụ cấp 50 để trần 20%-kiếm-tối-đa-của-mình KHÔNG phải là thứ chặn ở đây — bài
	// này đo đúng phép "25% của địch". Trần có bài riêng ngay dưới.
	a := put(g, "su_phu_mi", Vec{X: 3000, Y: 3000}, 50, 5)
	a.Facing = 0
	b := put(g, "cu_rua", Vec{X: 3000 + g.skillRange(a, "hut_mi_cone")*0.5, Y: 3000}, 10, 40)
	if lim := int(float64(g.Prog.MaxSwords(g.Prog.Realm(a.Level))) * g.skillNum("hut_mi_cap_of_own_max")); lim < 10 {
		t.Fatalf("trần của Sư Phụ chỉ %d — bài này cần trần ≥ 10 mới đo được phép 25%%", lim)
	}

	giveCharge(a)
	g.CastSkill(a)

	if got := 40 - b.Ring.Count; got != 10 {
		t.Errorf("địch mất %d kiếm, muốn 10 (25%% của 40)", got)
	}
	if got := a.Ring.Count - 5; got != 10 {
		t.Errorf("Sư Phụ nhận %d kiếm, muốn 10", got)
	}
}

// Trần "20% kiếm tối đa của MÌNH": không thì một Sư Phụ cấp thấp hút một Đạo Tổ là nhảy
// vọt lên trăm kiếm mà không tu luyện gì.
func TestHutMiCappedByOwnMax(t *testing.T) {
	g := skillGame(t)
	a := put(g, "su_phu_mi", Vec{X: 3000, Y: 3000}, 0, 1) // Luyện Khí: kiếm tối đa nhỏ
	a.Facing = 0
	b := put(g, "cu_rua", Vec{X: 3000 + g.skillRange(a, "hut_mi_cone")*0.5, Y: 3000}, 89, 200)
	ownMax := g.Prog.MaxSwords(g.Prog.Realm(a.Level))
	cap := int(float64(ownMax) * g.skillNum("hut_mi_cap_of_own_max"))

	before := b.Ring.Count
	giveCharge(a)
	g.CastSkill(a)

	if got := before - b.Ring.Count; got != cap {
		t.Errorf("cướp %d kiếm, trần phải là %d (20%% của %d)", got, cap, ownMax)
	}
}

// Hụt thì VẪN mất lượt (§6.3 mục 3: "Hút Mì chỉ 1 mục tiêu, hụt vẫn mất lượt").
func TestHutMiMissStillSpendsCharge(t *testing.T) {
	g := skillGame(t)
	a := put(g, "su_phu_mi", Vec{X: 3000, Y: 3000}, 10, 20)
	a.Facing = 0
	// Đứng sau lưng: ngoài nón.
	put(g, "cu_rua", Vec{X: 3000 - 200, Y: 3000}, 10, 40)

	// vào trận có sẵn đúng 1 lượt (v6.25)
	if !g.CastSkill(a) {
		t.Fatal("không dùng được kỹ năng")
	}
	if a.SkillCharges != 0 {
		t.Error("hụt mà được giữ lượt")
	}
	if a.Ring.Count != 20 {
		t.Errorf("hụt mà vẫn được %d kiếm", a.Ring.Count-20)
	}
}

// Gồng Cơ miễn khống chế và gỡ luôn thứ đang dính (spec §6.3).
func TestGongCoBlocksControl(t *testing.T) {
	g := skillGame(t)
	p := put(g, "be_ti", Vec{X: 3000, Y: 3000}, 10, 20)
	g.ApplyStatus(p, StRoot, 5)
	giveCharge(p)
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
	giveCharge(p)
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

// Bánh Bao Thần hồi 20% số kiếm TỐI ĐA của cảnh giới, không phụ thuộc vừa bị đánh hay chưa
// (spec v5.5 §6.3).
func TestBanhBaoRestoresShareOfMax(t *testing.T) {
	g := skillGame(t)
	p := put(g, "a_hon", Vec{X: 3000, Y: 3000}, 30, 10)
	maxSw := g.Prog.MaxSwords(g.Prog.Realm(p.Level))

	before := p.Ring.Count
	giveCharge(p)
	g.CastSkill(p)

	want := int(float64(maxSw) * g.Cfg.Skills.BanhBaoMaxShare)
	if got := p.Ring.Count - before; got != want {
		t.Fatalf("hồi %d kiếm, muốn %d (20%% của trần %d)", got, want, maxSw)
	}
}

// ... và không bao giờ vượt trần của cảnh giới: gần đầy kiếm thì chỉ hồi phần còn thiếu.
func TestBanhBaoNeverExceedsMax(t *testing.T) {
	g := skillGame(t)
	p := put(g, "a_hon", Vec{X: 3000, Y: 3000}, 30, 10)
	maxSw := g.Prog.MaxSwords(g.Prog.Realm(p.Level))
	g.AddSwords(p, maxSw-p.Ring.Count-1) // còn thiếu đúng 1 kiếm

	giveCharge(p)
	g.CastSkill(p)

	if p.Ring.Count != maxSw {
		t.Fatalf("có %d kiếm, trần là %d — phải lấp đúng tới trần", p.Ring.Count, maxSw)
	}
}

// Đàn em của Đại Ca không chém thân được (spec §6.3) — chúng chỉ mài kiếm.
func TestMinionsCannotBodyHit(t *testing.T) {
	g := skillGame(t)
	boss := put(g, "dai_ca_dep", Vec{X: 3000, Y: 3000}, 20, 30)
	victim := put(g, "cu_rua", Vec{X: 3000 + 60, Y: 3000}, 0, 1)
	giveCharge(boss)
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
	giveCharge(boss)
	g.CastSkill(boss)
	g.Time += g.Cfg.Skills.DanEmSecs + 0.1
	g.updateSkillState(1.0 / 30)
	for _, p := range g.PlayerList() {
		if p.IsMinion {
			t.Fatal("đàn em hết giờ mà vẫn còn trên bản đồ")
		}
	}
}

// Giao Hỏa Tốc đi xuyên MỌI vật cản (spec v5.5 §6.3), nhưng vẫn không ra khỏi biên khu.
func TestGhostPassesEverythingButBorder(t *testing.T) {
	cfg := config.MustLoadDefault()
	cfg.Room.MonsterCount, cfg.Items.CountPerRoom = 0, 0
	m := world.EmptyMap(cfg.Room.MapSize)
	m.Obstacles = []world.Obstacle{
		{ID: 0, Kind: "brick", Tier: 2, Shape: world.Shape{X: 3200, Y: 2900, W: 200, H: 200}},
		{ID: 1, Kind: "jade_pillar", Tier: 11, Shape: world.Shape{X: 4200, Y: 2900, W: 200, H: 200}},
		{ID: 2, Kind: config.BorderKind, Tier: config.TierBorder, Shape: world.Shape{X: 5200, Y: 2900, W: 200, H: 200}},
	}
	g := New(cfg, m, 1, 0)
	p := put(g, "shipper", Vec{X: 3000, Y: 3000}, 10, 20)
	giveCharge(p)
	g.CastSkill(p)

	for _, c := range []struct {
		name string
		at   Vec
	}{
		{"tường gạch (phá được)", Vec{X: 3300, Y: 3000}},
		{"cột ngọc (bậc 11)", Vec{X: 4300, Y: 3000}},
	} {
		if got := g.resolveBlockingFor(p, c.at, g.BodyR(p)); got.Dist(c.at) > 1e-9 {
			t.Errorf("đang Giao Hỏa Tốc mà vẫn bị %s chặn", c.name)
		}
	}

	// Biên khu thì không: spec nói rõ "trừ biên khu vực". Hai thứ phải chặn được:
	// vật cản loại `border` giữa hai khu, và rìa bản đồ.
	wall := Vec{X: 5300, Y: 3000}
	if got := g.resolveBlockingFor(p, wall, g.BodyR(p)); got.Dist(wall) < 1e-9 {
		t.Error("xuyên qua biên khu — spec cấm")
	}
	out := Vec{X: -500, Y: 3000}
	if got := g.resolveBlockingFor(p, out, g.BodyR(p)); got.Dist(out) < 1e-9 {
		t.Error("xuyên ra ngoài bản đồ — không có đất, không gì kéo vào lại")
	}
}

// Hết Giao Hỏa Tốc giữa lòng vật cản thì được đẩy MƯỢT ra chỗ trống, không bật ra tức thời.
func TestGhostUnsticksSmoothly(t *testing.T) {
	cfg := config.MustLoadDefault()
	cfg.Room.MonsterCount, cfg.Items.CountPerRoom = 0, 0
	m := world.EmptyMap(cfg.Room.MapSize)
	m.Obstacles = []world.Obstacle{
		{ID: 0, Kind: "stone_wall", Tier: 5, Shape: world.Shape{X: 3200, Y: 2900, W: 400, H: 400}},
	}
	g := New(cfg, m, 1, 0)
	p := put(g, "shipper", Vec{X: 3000, Y: 3000}, 10, 20)
	giveCharge(p)
	g.CastSkill(p)

	// Đi vào giữa vách đá rồi để kỹ năng hết hạn.
	stuck := Vec{X: 3400, Y: 3100}
	p.Pos = stuck
	dt := 1.0 / 30
	g.Time += g.skillNum("gom_don_secs") + 0.01
	g.updateSkillState(dt)

	if p.unstickUntil <= g.Time {
		t.Fatal("hết Giao Hỏa Tốc trong lòng vách đá mà không khởi động việc đẩy ra")
	}
	if p.Pos.Dist(stuck) > 1e-9 {
		t.Fatalf("bật ra ngay %0.1f đơn vị — phải trượt dần, nếu không client thấy giật",
			p.Pos.Dist(stuck))
	}
	// Chạy hết 0,3 giây: phải ra tới chỗ trống thật.
	for i := 0; i < int(g.Cfg.Skills.GhostUnstickSecs/dt)+2; i++ {
		g.Time += dt
		g.updateSkillState(dt)
	}
	if free := g.resolveBlocking(p.Pos, g.BodyR(p)); free.Dist(p.Pos) > 1e-6 {
		t.Errorf("đẩy xong vẫn còn kẹt trong vật cản (lệch %.1f)", free.Dist(p.Pos))
	}
}

// Trúng Phun Trân Châu thì bị bịt mắt và server báo riêng cho người đó.
func TestTranChauBlinds(t *testing.T) {
	g := skillGame(t)
	a := put(g, "map_tra_sua", Vec{X: 3000, Y: 3000}, 10, 20)
	a.Facing = 0
	b := put(g, "cu_rua", Vec{X: 3000 + 400, Y: 3000}, 10, 20)
	g.DrainEvents()

	giveCharge(a)
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
	giveCharge(a)
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
