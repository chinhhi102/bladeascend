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

// Spec §5.2: đi vào bị chặn như va tường — không lọt qua lá chắn, và không bị bật ngược.
func TestShieldBlocksWalkingIn(t *testing.T) {
	g := quietGame(t, nil)
	wall := g.AddPlayer(SpawnOptions{Name: "Wall", Level: 20, Swords: 60})
	mover := g.AddPlayer(SpawnOptions{Name: "Mover", Level: 20, Swords: 8})
	place(g, wall, Vec{X: 6000, Y: 6000}, 0)
	place(g, mover, Vec{X: 6000 + 2000, Y: 6000}, 0)
	mover.Moving, mover.InputAngle = true, math.Pi // đi thẳng vào
	need := g.minSeparation(mover, wall)
	if need <= 0 {
		t.Fatal("thiết lập sai: đối thủ 3 vòng phải có lá chắn")
	}
	for i := 0; i < 30*20; i++ {
		g.Tick(dt)
		wall.Pos = Vec{X: 6000, Y: 6000} // giữ nguyên để đo
	}
	d := mover.Pos.Dist(wall.Pos)
	if d < need-1 {
		t.Fatalf("lọt qua lá chắn: cách %.0f, tối thiểu %.0f", d, need)
	}
	if d > need+60 {
		t.Fatalf("bị bật ra quá xa: cách %.0f, tối thiểu %.0f (phải dừng sát lá chắn)", d, need)
	}
}

// Lá chắn chỉ chặn người ĐANG tự đi; người đứng yên không bị hất ra, nếu không kẻ mạnh
// sẽ tự đẩy nạn nhân ra ngoài tầm kiếm của chính mình.
func TestShieldDoesNotShoveStandingVictim(t *testing.T) {
	g := quietGame(t, nil)
	big := g.AddPlayer(SpawnOptions{Name: "Big", Level: 20, Swords: 60})
	small := g.AddPlayer(SpawnOptions{Name: "Small", Level: 20, Swords: 1})
	place(g, big, Vec{X: 6000, Y: 6000}, 0)
	at := Vec{X: 6000 + g.Geom(big).Inner, Y: 6000}
	place(g, small, at, 0)
	small.InvulnUntil = 1e9
	for i := 0; i < 30; i++ {
		g.Tick(dt)
	}
	if d := small.Pos.Dist(at); d > 1 {
		t.Fatalf("người đứng yên bị đẩy %.0f đơn vị", d)
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
