package game

import (
	"math"
	"testing"
)

// Spec §5.2: shield_r là mũi kiếm vòng KẾ ngoài cùng; còn 1 vòng thì không còn lá chắn.
func TestShieldRadiusIsSecondOutermostRing(t *testing.T) {
	g := quietGame(t, nil)
	one := g.AddPlayer(SpawnOptions{Name: "One", Level: 20, Swords: 8}) // 1 vòng
	if g.Geom(one).Rings != 1 {
		t.Fatalf("8 kiếm phải là 1 vòng, đang %d", g.Geom(one).Rings)
	}
	if s := g.ShieldR(one); s != 0 {
		t.Fatalf("còn 1 vòng mà vẫn có lá chắn %v — §5.1 sẽ không bao giờ chạm được thân", s)
	}
	three := g.AddPlayer(SpawnOptions{Name: "Three", Level: 20, Swords: 60}) // 3 vòng
	geo := g.Geom(three)
	if geo.Rings != 3 {
		t.Fatalf("60 kiếm phải là 3 vòng, đang %d", geo.Rings)
	}
	want := geo.TipDist(1, three.Ring.Tilt) + geo.Width/2 // vòng thứ 2 (kế ngoài cùng)
	if s := g.ShieldR(three); math.Abs(s-want) > 1e-9 {
		t.Fatalf("lá chắn = %v, muốn mũi kiếm vòng kế ngoài %v", s, want)
	}
	if s := g.ShieldR(three); s >= geo.Reach {
		t.Fatal("lá chắn phải nằm trong tầm với: vòng ngoài cùng là vùng mài kiếm")
	}
}

// Người chơi không chặn đường nhau: đi thẳng vào người khác thì vẫn tới được tận nơi,
// không khựng lại giữa khoảng không.
func TestPlayersDoNotBlockEachOther(t *testing.T) {
	g := quietGame(t, nil)
	wall := g.AddPlayer(SpawnOptions{Name: "Wall", Level: 20, Swords: 60})
	mover := g.AddPlayer(SpawnOptions{Name: "Mover", Level: 20, Swords: 8})
	place(g, wall, Vec{X: 6000, Y: 6000}, 0)
	place(g, mover, Vec{X: 6000 + 2000, Y: 6000}, 0)
	mover.Moving, mover.InputAngle = true, math.Pi // đi thẳng vào
	mover.InvulnUntil = 1e9                        // chỉ đo đường đi, không để chết giữa chừng
	closest := math.Inf(1)
	for i := 0; i < 30*20; i++ {
		g.Tick(dt)
		wall.Pos = Vec{X: 6000, Y: 6000} // giữ nguyên để đo
		closest = math.Min(closest, mover.Pos.Dist(wall.Pos))
	}
	// Tốc độ 195/s × 20s = 3900 > 2000 nên nếu không bị chặn thì phải chạm được tới thân.
	if closest > g.BodyR(wall)+g.BodyR(mover) {
		t.Fatalf("vẫn bị chặn: gần nhất chỉ tới %.0f đơn vị — phải đi được tới sát thân", closest)
	}
}

// Spec §5.2: còn lá chắn thì kiếm địch quét trùm lên thân cũng không ăn — phải mài rụng
// vòng ngoài trước. Đây là chỗ thay cho việc chặn đường đi ngày xưa.
func TestShieldedBodyTakesNoHit(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Level: 20, Swords: 1})
	b := g.AddPlayer(SpawnOptions{Name: "B", Level: 20, Swords: 60}) // 3 vòng → còn lá chắn
	place(g, a, Vec{X: 6000, Y: 6000}, 0)
	place(g, b, bladeMid(a), math.Pi/2) // kiếm của A nằm ngay trên thân B
	a.InvulnUntil = 1e9
	if !g.Shielded(b) {
		t.Fatal("thiết lập sai: 60 kiếm phải còn lá chắn")
	}
	lv := b.Level
	g.resolveSwordBody()
	if b.Dead || b.Level != lv {
		t.Fatalf("còn lá chắn mà vẫn ăn đòn: dead=%v level %d → %d", b.Dead, lv, b.Level)
	}
}

// Hết lá chắn (còn 1 vòng) thì §5.1 chém được thân ngay.
func TestUnshieldedBodyTakesHit(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Level: 20, Swords: 1})
	b := g.AddPlayer(SpawnOptions{Name: "B", Level: 20, Swords: 5}) // 1 vòng → không lá chắn
	place(g, a, Vec{X: 6000, Y: 6000}, 0)
	place(g, b, bladeMid(a), math.Pi/2)
	a.InvulnUntil = 1e9
	if g.Shielded(b) {
		t.Fatal("thiết lập sai: 5 kiếm phải là 1 vòng, không còn lá chắn")
	}
	g.resolveSwordBody()
	if !b.Dead {
		t.Fatal("không còn lá chắn mà vẫn không ăn cú chém thân")
	}
}

// Spec §5.2: kiếm bể trừ từ vòng ngoài cùng trước → vòng ngoài hết thì lá chắn lùi 1 lớp.
func TestShieldFallsBackWhenOuterRingBreaks(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "P", Level: 20, Swords: 31}) // 3 vòng (10+20+1)
	if g.Geom(p).Rings != 3 {
		t.Fatalf("31 kiếm phải là 3 vòng, đang %d", g.Geom(p).Rings)
	}
	before := g.ShieldR(p)
	g.LoseSwords(p, 1) // mất đúng cây duy nhất của vòng 3
	if g.Geom(p).Rings != 2 {
		t.Fatalf("mất kiếm vòng ngoài phải còn 2 vòng, đang %d", g.Geom(p).Rings)
	}
	if after := g.ShieldR(p); after >= before {
		t.Fatalf("lá chắn phải lùi vào: %v → %v", before, after)
	}
}
