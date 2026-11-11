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

// step chạy `secs` giây giờ game bằng nhịp tick thật — dùng cho kỹ năng có phần hồi/đếm
// dần. Gọi thẳng `updateSkillState` chứ không `Tick`: `Tick` còn chạy quái, vật phẩm và va
// kiếm, những thứ làm nhiễu đúng con số đang đo.
func step(g *Game, secs float64) {
	const dt = 1.0 / 30
	for t := 0.0; t < secs; t += dt {
		g.Time += dt
		g.updateSkillState(dt)
	}
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

// Loa Phường (v6.48): lấy hơi 0.8s rồi HÉT — đẩy lùi toả tròn + đổi chiều vòng kiếm địch.
func TestLoaPhuongPushesAllAround(t *testing.T) {
	g := skillGame(t)
	a := put(g, "ba_tam_loa", Vec{X: 3000, Y: 3000}, 10, 20)
	a.Facing = 0 // nhìn sang phải — v6.30: hướng nhìn không còn quan trọng
	front := put(g, "cu_rua", Vec{X: 3000 + 300, Y: 3000}, 10, 20)
	back := put(g, "cu_rua", Vec{X: 3000 - 300, Y: 3000}, 10, 20)
	far := put(g, "cu_rua", Vec{X: 3000, Y: 3000 + g.skillRange(a, "loa_radius") + 200}, 10, 20)
	dirBefore, farDir := back.Ring.Dir, far.Ring.Dir

	giveCharge(a)
	if !g.CastSkill(a) {
		t.Fatal("không dùng được Loa Phường")
	}
	if front.KnockUntil > g.Time {
		t.Fatal("chưa lấy hơi xong đã đẩy được người — mất luôn cách chống lại kỹ năng")
	}
	step(g, g.skillNum("loa_channel")+0.1)
	if front.KnockUntil <= g.Time || back.KnockUntil <= g.Time || back.Ring.Dir == dirBefore {
		t.Fatal("Loa toả tròn 360° (v6.30): người trước mặt VÀ sau lưng đều phải bị đẩy + đổi chiều")
	}
	if far.KnockUntil > g.Time || far.Ring.Dir != farDir {
		t.Fatal("người ngoài 6 U cũng trúng")
	}
}

// Checklist §17 v6.44: Ông Ba Võng trong 4s Trộm Kiếm va kiếm với người cùng cảnh giới →
// **cả hai mất kiếm như thường**; đối thủ mất bao nhiêu thì Ông Ba +đúng số đó. Ông lãi
// ròng khi mài nhau, không còn miễn mẻ (bỏ ở v6.44).
func TestTromKiemStealsClashedSwords(t *testing.T) {
	g := skillGame(t)
	a := put(g, "ong_ba_vong", Vec{X: 3000, Y: 3000}, 10, 10)
	b := put(g, "cu_rua", Vec{X: 3500, Y: 3000}, 10, 30)
	g.finishRuNgu(a) // hát xong → nằm võng Trộm Kiếm
	if !g.stealing(a) {
		t.Fatal("hát xong mà không vào Trộm Kiếm")
	}
	b.Status[StSleep] = 0
	b.Pos.X = a.Pos.X + (a.Geom.Reach+b.Geom.Reach)*0.8
	beforeA, beforeB := a.Ring.Count, b.Ring.Count
	for i := 0; i < 20; i++ {
		g.resolveSwordSword()
		g.Time += 1.0 / 30
	}
	lostB := beforeB - b.Ring.Count
	if lostB <= 0 {
		t.Fatal("hai vòng không chạm nhau — bài test đặt sai")
	}
	// Sổ "kiếm bể 10 s gần nhất" chỉ ghi kiếm BỂ, nên nó đúng bằng số Ông Ba tự mẻ —
	// phần cướp của đối thủ không vào sổ ai cả (§7).
	lostA := g.swordsLostRecently(a)
	if lostA <= 0 {
		t.Fatal("v6.44: kiếm Ông Ba phải mẻ như thường, không còn miễn mẻ")
	}
	if want := beforeA - lostA + lostB; a.Ring.Count != want {
		t.Fatalf("Ông Ba %d → %d kiếm (mẻ %d, cướp %d): muốn %d",
			beforeA, a.Ring.Count, lostA, lostB, want)
	}
	// Cùng cảnh giới và cùng nội tại độ bền thì hai bên mẻ bằng nhau, nên Ông Ba hoà vốn
	// (mẻ N, cướp N) chứ không tăng kiếm. Cái lợi nằm ở chỗ ĐỐI THỦ lỗ trắng N thanh: mài
	// qua mài lại thì ông không bao giờ lỗ, còn người kia cạn dần.
	if a.Ring.Count < beforeA {
		t.Fatalf("Ông Ba lỗ kiếm khi mài nhau: %d → %d", beforeA, a.Ring.Count)
	}
	if b.Ring.Count >= beforeB {
		t.Fatalf("đối thủ không mất kiếm: %d → %d", beforeB, b.Ring.Count)
	}
	g.Time = a.StealUntil + 0.01
	if g.stealing(a) {
		t.Fatal("hết 4 s mà vẫn Trộm Kiếm")
	}
}

// Trộm tới trần thì phần dư rơi Bó Kiếm, không bốc hơi, không đổi EXP.
func TestTromKiemOverflowDropsBundles(t *testing.T) {
	g := skillGame(t)
	a := put(g, "ong_ba_vong", Vec{X: 3000, Y: 3000}, 10, 10)
	maxSw := g.Prog.MaxSwords(g.Prog.Realm(a.Level))
	a.Ring.Count = maxSw - 1
	g.refreshDerived(a)
	exp := a.TotalExp
	g.giveSwords(a, 5)
	n := 0
	for _, it := range g.Items {
		if it.Kind == "bundle" {
			n += int(it.Value)
		}
	}
	if a.Ring.Count != maxSw || n != 4 || a.TotalExp != exp {
		t.Fatalf("vòng %d/%d, rơi %d (muốn %d/%d, rơi 4, EXP không đổi)", a.Ring.Count, maxSw, n, maxSw, maxSw)
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

// Hút Mì Đoạt Kiếm (v6.48 §6.3): húp 0.6s rồi cướp 25% kiếm của MỌI địch trong 5 U,
// không giới hạn số cướp.
func TestHutMiStealsQuarterFromEveryoneNear(t *testing.T) {
	g := skillGame(t)
	a := put(g, "su_phu_mi", Vec{X: 3000, Y: 3000}, 50, 5)
	r := g.skillRange(a, "hut_mi_r")
	b := put(g, "cu_rua", Vec{X: 3000 + r*0.5, Y: 3000}, 10, 40)
	c := put(g, "cu_rua", Vec{X: 3000 - r*0.5, Y: 3000}, 10, 40) // SAU LƯNG: v6.48 là bán kính
	far := put(g, "cu_rua", Vec{X: 3000, Y: 3000 + r*2}, 10, 40)
	before := a.Ring.Count

	giveCharge(a)
	g.CastSkill(a)
	if b.Ring.Count != 40 {
		t.Fatal("chưa húp xong đã cướp được kiếm")
	}
	step(g, g.skillNum("hut_mi_channel")+0.1)

	for _, o := range []*Player{b, c} {
		if got := 40 - o.Ring.Count; got != 10 {
			t.Errorf("địch mất %d kiếm, muốn 10 (25%% của 40)", got)
		}
	}
	if far.Ring.Count != 40 {
		t.Errorf("người ngoài 5 U cũng mất %d kiếm", 40-far.Ring.Count)
	}
	if got := a.Ring.Count - before; got != 20 {
		t.Errorf("Sư Phụ nhận %d kiếm, muốn 20 (hai người × 10, không còn trần)", got)
	}
}

// v6.48 BỎ trần "20% kiếm tối đa của mình": phần vượt trần vòng của Sư Phụ rơi xuống đất
// thành Bó Kiếm, không bốc hơi — người khác (kể cả nạn nhân) nhặt lại được.
func TestHutMiOverflowDropsAsBundles(t *testing.T) {
	g := skillGame(t)
	a := put(g, "su_phu_mi", Vec{X: 3000, Y: 3000}, 0, 1) // Luyện Khí: trần vòng nhỏ
	b := put(g, "cu_rua", Vec{X: 3000 + g.skillRange(a, "hut_mi_r")*0.5, Y: 3000}, 89, 200)
	ownMax := g.Prog.MaxSwords(g.Prog.Realm(a.Level))

	before := b.Ring.Count
	giveCharge(a)
	g.CastSkill(a)
	step(g, g.skillNum("hut_mi_channel")+0.1)

	stolen := before - b.Ring.Count
	if want := before / 4; stolen != want {
		t.Fatalf("cướp %d kiếm, muốn %d (25%% của %d — không còn trần)", stolen, want, before)
	}
	if a.Ring.Count != ownMax {
		t.Errorf("Sư Phụ có %d kiếm, phải đầy đúng trần vòng %d", a.Ring.Count, ownMax)
	}
	_, dropped := countItems(g, "bundle")
	if want := stolen - (ownMax - 1); int(dropped) != want {
		t.Errorf("rơi %.0f kiếm thành Bó Kiếm, muốn %d (phần vượt trần)", dropped, want)
	}
}

// Hụt thì VẪN mất lượt (§6.3 mục 3: "Hút Mì chỉ 1 mục tiêu, hụt vẫn mất lượt").
func TestHutMiMissStillSpendsCharge(t *testing.T) {
	g := skillGame(t)
	a := put(g, "su_phu_mi", Vec{X: 3000, Y: 3000}, 10, 20)
	// Đứng ngoài 5 U (v6.48 là BÁN KÍNH, không còn nón — đứng sau lưng vẫn dính).
	put(g, "cu_rua", Vec{X: 3000 - g.skillRange(a, "hut_mi_r")*2, Y: 3000}, 10, 40)

	// vào trận có sẵn đúng 1 lượt (v6.25)
	if !g.CastSkill(a) {
		t.Fatal("không dùng được kỹ năng")
	}
	step(g, g.skillNum("hut_mi_channel")+0.1)
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
	// Độ bền: Gồng Cơ +1.0, cộng (không nhân) với Cương Kiếm +0.5 và nội tại Bé Tí (spec §4.0).
	inn := g.Cfg.Innate["be_ti"].Durability
	if m := g.durBonus(p); math.Abs(m-(inn+1.0)) > 1e-9 {
		t.Fatalf("khoản bền khi gồng %+.2f, muốn %+.2f", m, inn+1.0)
	}
	g.ApplyBuff(p, "hard")
	if m := g.durBonus(p); math.Abs(m-(inn+1.5)) > 1e-9 {
		t.Fatalf("gồng + Cương Kiếm %+.2f, muốn %+.2f (cộng, không nhân dồn)", m, inn+1.5)
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

// Bánh Bao Thần (v6.48): hồi NGAY 10% kiếm tối đa, rồi 20% nữa rải đều trong 4 giây.
func TestBanhBaoHealsNowThenOverTime(t *testing.T) {
	g := skillGame(t)
	sc := g.Cfg.Skills
	p := put(g, "a_hon", Vec{X: 3000, Y: 3000}, 30, 10)
	maxSw := g.Prog.MaxSwords(g.Prog.Realm(p.Level))

	before := p.Ring.Count
	giveCharge(p)
	g.CastSkill(p)

	wantNow := int(float64(maxSw) * sc.BanhBaoNow)
	if got := p.Ring.Count - before; got != wantNow {
		t.Fatalf("hồi ngay %d kiếm, muốn %d (%.0f%% của trần %d)", got, wantNow, sc.BanhBaoNow*100, maxSw)
	}

	// Nửa chừng: đã nhả khoảng một nửa phần hồi dần, chưa nhả hết.
	wantLater := int(float64(maxSw) * sc.BanhBaoRegen)
	step(g, sc.BanhBaoRegenSecs/2)
	half := p.Ring.Count - before - wantNow
	if half <= 0 || half >= wantLater {
		t.Fatalf("giữa chừng hồi thêm %d, muốn trong khoảng (0, %d) — phải là hồi DẦN", half, wantLater)
	}

	step(g, sc.BanhBaoRegenSecs/2+0.2)
	if got := p.Ring.Count - before; got != wantNow+wantLater {
		t.Fatalf("hết 4 giây hồi tổng %d, muốn %d", got, wantNow+wantLater)
	}
}

// "Bị chém chết thì mất phần chưa hồi" (§6.3 v6.48).
func TestBanhBaoRegenLostOnDeath(t *testing.T) {
	g := skillGame(t)
	p := put(g, "a_hon", Vec{X: 3000, Y: 3000}, 30, 10)
	giveCharge(p)
	g.CastSkill(p)
	if p.regenLeft <= 0 {
		t.Fatal("ăn bánh xong mà không nợ thanh kiếm nào — không có gì để mất")
	}
	g.onDeathClear(p)
	if p.regenLeft != 0 {
		t.Fatalf("chết rồi còn nợ %d kiếm sẽ hồi", p.regenLeft)
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
	// Phần hồi dần cũng không được vượt trần.
	step(g, g.Cfg.Skills.BanhBaoRegenSecs+0.2)
	if p.Ring.Count != maxSw {
		t.Fatalf("hồi dần xong có %d kiếm, vượt trần %d", p.Ring.Count, maxSw)
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

// Bung Cơ (v6.48 §6.3): hết 4 giây gồng thì đẩy địch trong 2 U ra 3 U và mỗi người mất 10%
// số kiếm hiện có — kiếm rơi quanh HỌ thành Bó Kiếm, Bé Tí không nhận được thanh nào.
func TestBungCoPushesAndDropsSwords(t *testing.T) {
	g := skillGame(t)
	p := put(g, "be_ti", Vec{X: 3000, Y: 3000}, 4, 20)
	o := put(g, "cu_rua", Vec{X: 3000 + g.skillRange(p, "bung_co_wave")*0.5, Y: 3000}, 4, 20)
	beforeMe, beforeThem, beforeAt := p.Ring.Count, o.Ring.Count, o.Pos

	giveCharge(p)
	g.CastSkill(p)
	if !o.HasStatus(StBrace, g.Time) && o.Ring.Count != beforeThem {
		t.Fatal("vừa gồng xong đã mất kiếm — phần bung phải đợi hết 4 giây")
	}
	step(g, g.Cfg.Skills.BraceSecs+0.2)

	want := int(float64(beforeThem) * g.skillNum("bung_co_lose"))
	if got := beforeThem - o.Ring.Count; got != want {
		t.Errorf("địch mất %d kiếm, muốn %d (10%% của %d)", got, want, beforeThem)
	}
	if p.Ring.Count != beforeMe {
		t.Errorf("Bé Tí nhận %d kiếm — bung cơ KHÔNG cướp kiếm", p.Ring.Count-beforeMe)
	}
	if _, swords := countItems(g, "bundle"); int(swords) != want {
		t.Errorf("rơi %.0f kiếm thành Bó Kiếm, muốn %d", swords, want)
	}
	if o.KnockUntil <= g.Time && o.Pos.Dist(beforeAt) < 1 {
		t.Error("địch không bị đẩy")
	}
}

// Ở ngoài 2 U thì không dính gì.
func TestBungCoOnlyHitsNearby(t *testing.T) {
	g := skillGame(t)
	p := put(g, "be_ti", Vec{X: 3000, Y: 3000}, 4, 20)
	far := put(g, "cu_rua", Vec{X: 3000 + g.skillRange(p, "bung_co_wave")*2, Y: 3000}, 4, 20)
	before := far.Ring.Count
	giveCharge(p)
	g.CastSkill(p)
	step(g, g.Cfg.Skills.BraceSecs+0.2)
	if far.Ring.Count != before {
		t.Fatalf("người đứng ngoài tầm cũng mất %d kiếm", before-far.Ring.Count)
	}
}

// Loa Phường v6.48 nhịp 1: lấy hơi 0.8s thì HÚT mọi thứ nhặt được trong 6 U về mình —
// theo đúng luật hấp thụ thường (§6.3).
func TestLoaVacuumsItemsInRadius(t *testing.T) {
	g := skillGame(t)
	p := put(g, "ba_tam_loa", Vec{X: 3000, Y: 3000}, 10, 5)
	r := g.skillRange(p, "loa_radius")
	// Một bó kiếm trong tầm, một bó ngoài tầm.
	g.spawnItemNear("bundle", 4, Vec{X: 3000 + r*0.6, Y: 3000}, 0)
	g.spawnItemNear("bundle", 4, Vec{X: 3000 + r*2, Y: 3000}, 0)
	before := p.Ring.Count

	giveCharge(p)
	g.CastSkill(p)
	if n, _ := countItems(g, "bundle"); n != 2 {
		t.Fatal("chưa lấy hơi xong đã hút được đồ")
	}
	step(g, g.skillNum("loa_channel")+0.1)

	if n, _ := countItems(g, "bundle"); n != 1 {
		t.Errorf("còn %d bó kiếm trên đất, muốn 1 (bó trong 6 U phải bị hút)", n)
	}
	if p.Ring.Count <= before {
		t.Error("hút bó kiếm mà số kiếm không tăng")
	}
}

// "Đang bị mù thì không hút được" (§6.3) — dùng chung luật `CanPickUp` với việc nhặt thường.
func TestLoaVacuumBlockedWhenBlind(t *testing.T) {
	g := skillGame(t)
	p := put(g, "ba_tam_loa", Vec{X: 3000, Y: 3000}, 10, 5)
	g.spawnItemNear("bundle", 4, Vec{X: 3000 + g.skillRange(p, "loa_radius")*0.5, Y: 3000}, 0)
	g.ApplyStatus(p, StBlind, 5)

	giveCharge(p)
	g.CastSkill(p)
	step(g, g.skillNum("loa_channel")+0.1)

	if n, _ := countItems(g, "bundle"); n != 1 {
		t.Fatalf("đang mù mà vẫn hút được đồ (còn %d bó)", n)
	}
}

// Phun Trân Châu v6.48: nạn nhân bị MÙ và ĐẢO ĐIỀU KHIỂN 3 giây; Thằng Mập tự hoá khổng lồ
// ×2 trong 2 giây — thân, vòng kiếm và số kiếm đều gấp đôi.
func TestTranChauBlindsInvertsAndGrows(t *testing.T) {
	g := skillGame(t)
	sc := g.Cfg.Skills
	fat := put(g, "map_tra_sua", Vec{X: 3000, Y: 3000}, 10, 20)
	fat.Facing = 0
	vic := put(g, "a_hon", Vec{X: 3000 + g.skillRange(fat, "tran_chau_cone")*0.4, Y: 3000}, 10, 20)
	scale0, reach0, sw0 := fat.Scale, fat.Geom.Reach, fat.Ring.Count

	giveCharge(fat)
	g.CastSkill(fat)

	if !vic.HasStatus(StBlind, g.Time) || !vic.HasStatus(StInvert, g.Time) {
		t.Fatal("nạn nhân phải vừa mù vừa bị đảo điều khiển")
	}
	if fat.Scale <= scale0*1.9 || fat.Geom.Reach <= reach0*1.5 {
		t.Errorf("khổng lồ: scale %.2f→%.2f, tầm %.0f→%.0f — phải ×2", scale0, fat.Scale, reach0, fat.Geom.Reach)
	}
	if want := int(float64(sw0) * sc.GiantMult); fat.Ring.Count != want {
		t.Errorf("khổng lồ có %d kiếm, muốn %d", fat.Ring.Count, want)
	}
	if fat.TempSwords != fat.Ring.Count-sw0 {
		t.Errorf("kiếm tạm = %d, muốn %d", fat.TempSwords, fat.Ring.Count-sw0)
	}

	// Hết 2 giây: kiếm tạm tan, thân co lại.
	step(g, sc.GiantSecs+0.1)
	if fat.Ring.Count != sw0 || fat.TempSwords != 0 {
		t.Errorf("hết khổng lồ còn %d kiếm (tạm %d), muốn %d", fat.Ring.Count, fat.TempSwords, sw0)
	}
	if math.Abs(fat.Scale-scale0) > 1e-9 {
		t.Errorf("hết khổng lồ mà scale còn %.2f, muốn %.2f", fat.Scale, scale0)
	}
}

// Đảo điều khiển: bấm sang phải thì đi sang TRÁI (§6.3 v6.48).
func TestInvertFlipsInput(t *testing.T) {
	g := skillGame(t)
	p := put(g, "a_hon", Vec{X: 3000, Y: 3000}, 10, 5)
	g.SetInput(p.ID, 1, 0, true, false)
	g.movePlayer(p, 0.1)
	if p.Pos.X <= 3000 {
		t.Fatal("chưa bị đảo mà bấm phải lại không đi sang phải")
	}

	p.Pos = Vec{X: 3000, Y: 3000}
	g.ApplyStatus(p, StInvert, 3)
	g.SetInput(p.ID, 2, 0, true, false)
	g.movePlayer(p, 0.1)
	if p.Pos.X >= 3000 {
		t.Fatalf("đang bị đảo mà vẫn đi sang phải (x %.0f)", p.Pos.X)
	}
}

// Kiếm TẠM: bể trước kiếm thật, không rơi khi chết, không bị cướp.
func TestTempSwordsBreakFirstAndNeverTransfer(t *testing.T) {
	g := skillGame(t)
	fat := put(g, "map_tra_sua", Vec{X: 3000, Y: 3000}, 10, 20)
	giveCharge(fat)
	g.CastSkill(fat)
	temp := fat.TempSwords
	if temp <= 2 {
		t.Fatalf("chỉ có %d kiếm tạm — bài này cần nhiều hơn", temp)
	}

	// Bể 2 cây: trừ vào kiếm TẠM trước.
	g.loseSwordsHit(fat, 2, true)
	if fat.TempSwords != temp-2 {
		t.Errorf("kiếm tạm còn %d, muốn %d — kiếm tạm phải bể trước", fat.TempSwords, temp-2)
	}
	if g.swordsLostRecently(fat) != 0 {
		t.Error("kiếm tạm bể mà vào sổ rơi-khi-chết")
	}

	// Cướp: kẻ cướp chỉ nhận được phần kiếm THẬT.
	thief := put(g, "su_phu_mi", Vec{X: 3000 + 50, Y: 3000}, 10, 1)
	before := thief.Ring.Count
	got := g.stealLoseSwords(fat, fat.TempSwords)
	if got != 0 {
		t.Errorf("cướp được %d kiếm tạm — spec nói không cướp được", got)
	}
	g.giveSwords(thief, got)
	if thief.Ring.Count != before {
		t.Errorf("kẻ cướp có thêm %d kiếm từ kiếm tạm", thief.Ring.Count-before)
	}

	// Chết: kiếm tạm không rơi.
	real := fat.Ring.Count - fat.TempSwords
	g.dropDeathSwords(fat)
	if _, swords := countItems(g, "bundle"); int(swords) > real+2 {
		t.Errorf("xác rơi %.0f kiếm, chỉ được rơi tối đa %d (kiếm tạm không rơi)", swords, real+2)
	}
}

// Dòng §17 (v6.50/v6.51): "Cụ Rùa rụt mai giữa 3 địch: 2s không mất kiếm, không chết;
// địch trong 4 U bị kéo về 2s; kiếm địch bể rơi thành Bó Kiếm, số kiếm của cụ không tăng."
func TestRutCoTauntsAndDropsFoeSwords(t *testing.T) {
	g := skillGame(t)
	cu := put(g, "cu_rua", Vec{X: 3000, Y: 3000}, 10, 30)
	r := g.skillRange(cu, "rut_co_taunt_r")
	var foes []*Player
	for i, off := range []Vec{{X: r * 0.5}, {X: -r * 0.5}, {Y: r * 0.5}} {
		foes = append(foes, put(g, "a_hon", Vec{X: 3000, Y: 3000}.Add(off), 10, 20+i))
	}
	far := put(g, "a_hon", Vec{X: 3000 + r*3, Y: 3000}, 10, 20)
	swordsBefore := cu.Ring.Count

	giveCharge(cu)
	g.CastSkill(cu)

	for _, o := range foes {
		if !o.HasStatus(StTaunt, g.Time) {
			t.Fatalf("địch trong 4 U không bị khiêu khích")
		}
	}
	if far.HasStatus(StTaunt, g.Time) {
		t.Error("địch ngoài 4 U cũng bị khiêu khích")
	}
	// Bị kéo VỀ PHÍA cụ, không tự chọn hướng: bấm chạy ra xa vẫn tiến lại gần.
	o := foes[0]
	d0 := o.Pos.Dist(cu.Pos)
	g.SetInput(o.ID, 1, 0, true, false) // bấm sang phải = ra xa
	for i := 0; i < 10; i++ {
		g.movePlayer(o, 1.0/30)
	}
	if o.Pos.Dist(cu.Pos) >= d0 {
		t.Errorf("bị khiêu khích mà vẫn đi ra xa: %.0f → %.0f", d0, o.Pos.Dist(cu.Pos))
	}
	if g.CanCastSkill(o) {
		t.Error("đang bị khiêu khích mà vẫn dùng được kỹ năng")
	}

	// Kiếm địch bể khi va với cụ: rơi thành Bó Kiếm, cụ KHÔNG nhận.
	foeBefore := o.Ring.Count
	lostA, lostB := g.shellDrop(cu, o, 0, 3, o.Pos)
	if lostA != 0 || lostB != 0 {
		t.Errorf("còn %d/%d kiếm bể lại lần nữa — shellDrop phải nuốt hết", lostA, lostB)
	}
	if got := foeBefore - o.Ring.Count; got != 3 {
		t.Errorf("địch mất %d kiếm, muốn 3", got)
	}
	if cu.Ring.Count != swordsBefore {
		t.Errorf("cụ có thêm %d kiếm — spec nói cụ KHÔNG nhận (khác Ông Ba)", cu.Ring.Count-swordsBefore)
	}
	if _, swords := countItems(g, "bundle"); int(swords) != 3 {
		t.Errorf("rơi %.0f kiếm thành Bó Kiếm quanh chỗ va, muốn 3", swords)
	}
	if g.swordsLostRecently(o) != 0 {
		t.Error("kiếm đã rơi thành Bó Kiếm còn vào sổ rơi-khi-chết — nạn nhân chết sẽ rơi hai lần")
	}

	// Trong mai: không mất kiếm.
	if g.loseSwordsHit(cu, 5, true) != 0 {
		t.Error("trong mai mà vẫn bị bể kiếm")
	}
}
