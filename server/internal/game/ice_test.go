package game

import (
	"math"
	"testing"

	"kiemxoay/server/internal/config"
	"kiemxoay/server/internal/world"
)

// iceGame dựng một thế giới có đúng một hồ băng to ở giữa.
func iceGame(t *testing.T) *Game {
	t.Helper()
	cfg := config.MustLoadDefault()
	cfg.Room.MonsterCount, cfg.Items.CountPerRoom = 0, 0
	m := world.EmptyMap(cfg.Room.MapSize)
	c := cfg.Room.MapSize / 2
	m.Zones = append(m.Zones, world.Zone{Kind: "ice", Shape: world.Shape{X: c - 3000, Y: c - 3000, W: 6000, H: 6000}})
	return New(cfg, m, 1, 0)
}

// Trên mặt băng, thả phím thì vẫn TRÔI TIẾP — đó là cả điểm của khu Băng Nguyệt Hồ
// (spec §8.0). Trên đất thì thả phím là dừng ngay.
func TestIceKeepsSlidingAfterInputStops(t *testing.T) {
	g := iceGame(t)
	c := g.Cfg.Room.MapSize / 2
	dt := 1 / float64(g.Cfg.Net.TickHz)

	p := &Player{Pos: Vec{X: c, Y: c}, Moving: true, InputAngle: 0}
	g.refreshDerived(p)
	for i := 0; i < 60; i++ { // hai giây tăng tốc
		g.movePlayerForTest(p, dt)
	}
	cruise := p.Vel.X
	if cruise < g.Cfg.Player.MoveBase*0.8 {
		t.Fatalf("chạy thẳng trên băng chỉ đạt %.0f, muốn gần %.0f", cruise, g.Cfg.Player.MoveBase)
	}

	// Thả phím: phải trôi thêm một quãng đáng kể rồi mới gần dừng.
	p.Moving = false
	before := p.Pos.X
	for i := 0; i < 30; i++ {
		g.movePlayerForTest(p, dt)
	}
	slid := p.Pos.X - before
	if slid < 40 {
		t.Fatalf("thả phím xong chỉ trôi %.1f đơn vị — không còn là mặt băng", slid)
	}
	if p.Vel.X >= cruise {
		t.Fatalf("thả phím mà vận tốc không giảm: %.1f → %.1f", cruise, p.Vel.X)
	}
}

// Ra khỏi băng là hết quán tính ngay: trên đất thả phím phải đứng im tại chỗ.
func TestLandHasNoInertia(t *testing.T) {
	g := iceGame(t)
	dt := 1 / float64(g.Cfg.Net.TickHz)
	p := &Player{Pos: Vec{X: 1200, Y: 1200}, Moving: true, InputAngle: 0, Vel: Vec{X: 500, Y: 0}}
	g.refreshDerived(p)
	g.movePlayerForTest(p, dt)
	if p.Vel.X != 0 || p.Vel.Y != 0 {
		t.Fatalf("trên đất mà còn vận tốc %v — quán tính phải chỉ có trên băng", p.Vel)
	}
	p.Moving = false
	at := p.Pos
	for i := 0; i < 10; i++ {
		g.movePlayerForTest(p, dt)
	}
	if p.Pos.Dist(at) > 1e-9 {
		t.Fatalf("trên đất thả phím mà vẫn trôi %.2f đơn vị", p.Pos.Dist(at))
	}
}

// Tốc độ tới hạn trên băng phải khớp công thức accel × dt / (1 − friction); nếu lệch thì
// chạy thẳng trên băng nhanh hoặc chậm hẳn so với trên đất và cân bằng hỏng.
func TestIceTerminalSpeedMatchesConfig(t *testing.T) {
	g := iceGame(t)
	c := g.Cfg.Room.MapSize / 2
	dt := 1 / float64(g.Cfg.Net.TickHz)
	p := &Player{Pos: Vec{X: c, Y: c}, Moving: true, InputAngle: 0}
	g.refreshDerived(p)
	for i := 0; i < 400; i++ {
		g.movePlayerForTest(p, dt)
		p.Pos = Vec{X: c, Y: c} // giữ tại chỗ để không trôi ra khỏi hồ
	}
	want := g.Cfg.Ice.Accel * dt / (1 - g.Cfg.Ice.Friction)
	if math.Abs(p.Vel.X-want) > want*0.02 {
		t.Fatalf("tốc độ tới hạn %.1f, công thức nói %.1f", p.Vel.X, want)
	}
	if math.Abs(want-g.Cfg.Player.MoveBase) > g.Cfg.Player.MoveBase*0.05 {
		t.Fatalf("tới hạn trên băng %.0f lệch quá xa move_base %.0f — chỉnh ice.accel", want, g.Cfg.Player.MoveBase)
	}
}
