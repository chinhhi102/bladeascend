package game

import (
	"math"
	"testing"

	"bladeascend/server/internal/world"
)

// Spec v6.25 §4.1: tốc độ CỘNG trên base, kẹp [0.4, 2.0] (Gom Đơn: 3.0).

// speedGame: một khu có vùng gió (hướng 0 rad = sang phải) và vùng bùn tách nhau.
func speedGame(t *testing.T) (*Game, *Player) {
	t.Helper()
	sq := func(x, y float64) world.Shape { return world.Shape{X: x, Y: y, W: 400, H: 400} }
	m := world.EmptyMap(12000)
	m.Zones = append(m.Zones,
		world.Zone{ID: len(m.Zones), Kind: "wind", Shape: sq(1000, 1000), WindDir: 0},
		world.Zone{ID: len(m.Zones) + 1, Kind: "mud", Shape: sq(3000, 1000)},
		world.Zone{ID: len(m.Zones) + 2, Kind: "mud", Shape: sq(3000, 1000)}, // bùn chồng bùn: vẫn chỉ −0.4
	)
	g := quietGame(t, m)
	p := g.AddPlayer(SpawnOptions{Name: "P", Level: 15, Swords: 1})
	return g, p
}

func wantSpeed(t *testing.T, g *Game, got, mult float64, what string) {
	t.Helper()
	if want := g.Cfg.Player.MoveBase * mult; math.Abs(got-want) > 1e-6 {
		t.Fatalf("%s: tốc %.2f, muốn ×%.2f = %.2f", what, got, mult, want)
	}
}

func TestSpeedAdditiveCapped(t *testing.T) {
	g, p := speedGame(t)
	right := Vec{X: 1, Y: 0}
	wind, mud := Vec{X: 1200, Y: 1200}, Vec{X: 3200, Y: 1200}
	place(g, p, Vec{X: 6000, Y: 6000}, 0)
	g.SetInput(p.ID, 1, 0, true, true) // giữ Thiêu Đốt
	g.ApplyBuff(p, "haste")
	wantSpeed(t, g, g.MoveSpeedAt(p, wind, right), 2.0, "Thần Hành + Thiêu Đốt + theo gió (cũ ×3.78)")

	g.ApplyStatus(p, StGhost, 4)
	wantSpeed(t, g, g.MoveSpeedAt(p, wind, right), 3.0, "Gom Đơn + mọi buff")

	g2, q := speedGame(t)
	place(g2, q, Vec{X: 6000, Y: 6000}, 0)
	g2.SetInput(q.ID, 1, 0, true, true)
	wantSpeed(t, g2, g2.MoveSpeedAt(q, mud, right), 1.4, "Thiêu Đốt trong bùn")
	g2.SetInput(q.ID, 1, 0, true, false)
	wantSpeed(t, g2, g2.MoveSpeedAt(q, mud, right), 0.6, "chỉ bùn (hai vùng chồng nhau)")
	wantSpeed(t, g2, g2.MoveSpeedAt(q, wind, Vec{X: -1}), 0.8, "ngược gió")
}

// Sàn 0.4 chỉ áp cho làm chậm; khống chế đứng im vẫn là 0.
func TestSpeedFloorAndFrozen(t *testing.T) {
	g, p := speedGame(t)
	g.Cfg.Speed.Mud = -2
	wantSpeed(t, g, g.MoveSpeedAt(p, Vec{X: 3200, Y: 1200}, Vec{X: 1}), 0.4, "sàn")
	g.ApplyStatus(p, StStun, 1)
	wantSpeed(t, g, g.MoveSpeedAt(p, Vec{X: 3200, Y: 1200}, Vec{X: 1}), 0, "choáng")
}
