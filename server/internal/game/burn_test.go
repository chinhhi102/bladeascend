package game

import (
	"math"
	"testing"
)

// Spec §4.1c: giữ Thiêu Đốt Tinh Huyết → chạy nhanh hơn ×mult, đổi lại mất EXP mỗi giây.
func TestBurnSpeedsUpAndCostsExp(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "P", Level: 15, Swords: 1})
	// Đặt ở giữa tầng: nếu đứng đúng ngưỡng thì chỉ đốt một chút là tụt tầng ngay và
	// giá mỗi giây đổi theo, không đo được con số sạch.
	g.GainExp(p, 0.5*g.Prog.ExpNeeded(15), false)
	place(g, p, Vec{X: 6000, Y: 6000}, 0)
	base := g.MoveSpeed(p)

	g.SetInput(p.ID, 1, 0, true, true)
	if !g.Burning(p) {
		t.Fatal("giữ phím mà không đốt")
	}
	if s := g.MoveSpeed(p); math.Abs(s-base*g.Cfg.Burn.Mult) > 1e-9 {
		t.Fatalf("tốc độ khi đốt = %v, muốn %v", s, base*g.Cfg.Burn.Mult)
	}

	before := p.TotalExp
	want := g.Cfg.Burn.ExpFracPerSec * g.Prog.ExpNeeded(p.Level) // mất trong 1 giây
	for i := 0; i < int(1/dt); i++ {
		g.Tick(dt)
	}
	spent := before - p.TotalExp
	if math.Abs(spent-want) > want*0.02 {
		t.Fatalf("1 giây đốt mất %v EXP, muốn ≈%v", spent, want)
	}
	if math.Abs(p.BurnedExp-spent) > 1e-6 {
		t.Fatalf("BurnedExp = %v, muốn %v", p.BurnedExp, spent)
	}
}

// Đứng yên thì không đốt: chỉ trả giá khi thật sự chạy nhanh.
func TestBurnOnlyWhileMoving(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "P", Level: 15, Swords: 1})
	place(g, p, Vec{X: 6000, Y: 6000}, 0)
	g.SetInput(p.ID, 1, 0, false, true)
	if g.Burning(p) {
		t.Fatal("đứng yên mà vẫn đốt EXP")
	}
	before := p.TotalExp
	for i := 0; i < 30; i++ {
		g.Tick(dt)
	}
	if p.TotalExp != before {
		t.Fatalf("đứng yên vẫn mất EXP: %v → %v", before, p.TotalExp)
	}
}

// Spec §4.1c: đốt có thể tụt tầng nhưng KHÔNG bao giờ rớt cảnh giới.
func TestBurnStopsAtRealmFloor(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "P", Level: 12, Swords: 1}) // Trúc Cơ tầng 3
	place(g, p, Vec{X: 6000, Y: 6000}, 0)
	floor := g.Prog.Threshold(10) // đáy Trúc Cơ
	g.SetInput(p.ID, 1, 0, true, true)
	for i := 0; i < 30*120 && p.TotalExp > floor; i++ {
		g.Tick(dt)
	}
	if g.Prog.Realm(p.Level) != 1 {
		t.Fatalf("đốt làm rớt xuống cảnh giới %d, phải giữ nguyên Trúc Cơ", g.Prog.Realm(p.Level))
	}
	if p.Level != 10 {
		t.Fatalf("đốt dừng ở level %d, muốn 10 (Trúc Cơ tầng 1)", p.Level)
	}
	if math.Abs(p.TotalExp-floor) > 1e-6 {
		t.Fatalf("total_exp = %v, muốn dừng đúng đáy cảnh giới %v", p.TotalExp, floor)
	}
	// Chạm đáy rồi thì hết đốt được, tốc độ trở lại bình thường.
	if g.Burning(p) {
		t.Fatal("hết EXP trong cảnh giới mà vẫn đốt được")
	}
	if s := g.MoveSpeed(p); math.Abs(s-g.Cfg.Player.MoveBase) > 1e-9 {
		t.Fatalf("tốc độ sau khi cạn = %v, muốn về %v", s, g.Cfg.Player.MoveBase)
	}
}

// Luyện Khí đốt tới 0 EXP thì dừng, không âm và không chết.
func TestBurnAtLowestRealmStopsAtZero(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "P", Level: 5, Swords: 1})
	place(g, p, Vec{X: 6000, Y: 6000}, 0)
	g.SetInput(p.ID, 1, 0, true, true)
	for i := 0; i < 30*120 && p.TotalExp > 0; i++ {
		g.Tick(dt)
	}
	if p.TotalExp != 0 || p.Level != 0 {
		t.Fatalf("total_exp = %v, level = %d; muốn 0 và 0", p.TotalExp, p.Level)
	}
	if p.Dead {
		t.Fatal("đốt cạn EXP không được làm chết người chơi")
	}
}

// Thiêu Đốt cộng dồn với buff Thần Hành.
func TestBurnStacksWithHaste(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "P", Level: 15, Swords: 1})
	place(g, p, Vec{X: 6000, Y: 6000}, 0)
	g.SetInput(p.ID, 1, 0, true, true)
	g.ApplyBuff(p, "haste")
	want := g.Cfg.Player.MoveBase * g.Cfg.Buffs["haste"].Mult * g.Cfg.Burn.Mult
	if s := g.MoveSpeed(p); math.Abs(s-want) > 1e-9 {
		t.Fatalf("tốc độ = %v, muốn %v", s, want)
	}
}
