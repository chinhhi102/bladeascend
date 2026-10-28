package game

import (
	"math"
	"testing"

	"kiemxoay/server/internal/world"
)

// playMap is an empty map that declares a border margin, like the ones mapgen writes.
func playMap(size, margin float64) *world.Map {
	m := world.EmptyMap(size)
	m.PlayMargin = margin
	return m
}

// Spec §8.1: không người chơi nào được đứng ngoài tường biên, dù bị hất ra bằng cách nào.
func TestPlayerCannotLeavePlayArea(t *testing.T) {
	const size, margin = 12000, 420
	g := quietGame(t, playMap(size, margin))
	p := g.AddPlayer(SpawnOptions{Name: "P", Level: 20, Swords: 30})
	r := g.BodyR(p)
	lo, hi := g.PlayBounds(r)

	// Đi thẳng vào từng góc trong 20 giây.
	for _, ang := range []float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2, math.Pi / 4, 5 * math.Pi / 4} {
		place(g, p, Vec{X: size / 2, Y: size / 2}, 0)
		p.Moving, p.InputAngle = true, ang
		for i := 0; i < 30*20; i++ {
			g.Tick(dt)
			if p.Pos.X < lo-0.5 || p.Pos.X > hi+0.5 || p.Pos.Y < lo-0.5 || p.Pos.Y > hi+0.5 {
				t.Fatalf("đi hướng %.2f rad ra ngoài: (%.0f,%.0f), vùng chơi [%.0f,%.0f]", ang, p.Pos.X, p.Pos.Y, lo, hi)
			}
		}
	}

	// Bị đặt thẳng ra ngoài (knockback, dịch chuyển lỗi) thì tick sau phải bị kéo vào.
	for _, out := range []Vec{{X: 10, Y: 10}, {X: size - 10, Y: size / 2}, {X: size / 2, Y: -500}} {
		p.Moving = false
		place(g, p, out, 0)
		g.Tick(dt)
		if p.Pos.X < lo-0.5 || p.Pos.X > hi+0.5 || p.Pos.Y < lo-0.5 || p.Pos.Y > hi+0.5 {
			t.Fatalf("đặt ở (%.0f,%.0f) không được kéo vào: còn (%.0f,%.0f)", out.X, out.Y, p.Pos.X, p.Pos.Y)
		}
	}
}

// Knockback mạnh ở sát biên cũng không được hất người chơi ra ngoài tường.
func TestKnockbackCannotPushOutOfPlayArea(t *testing.T) {
	const size, margin = 12000, 420
	g := quietGame(t, playMap(size, margin))
	p := g.AddPlayer(SpawnOptions{Name: "P", Level: 20, Swords: 30})
	r := g.BodyR(p)
	lo, hi := g.PlayBounds(r)
	place(g, p, Vec{X: lo + 5, Y: lo + 5}, 0)
	p.KnockVel = Vec{X: -4000, Y: -4000}
	p.KnockUntil = g.Time + 1
	for i := 0; i < 30; i++ {
		g.Tick(dt)
		if p.Pos.X < lo-0.5 || p.Pos.Y < lo-0.5 {
			t.Fatalf("knockback hất ra ngoài: (%.0f,%.0f), tối thiểu %.0f", p.Pos.X, p.Pos.Y, lo)
		}
	}
	_ = hi
}

// Quái vật cũng không được đi ra ngoài tường biên.
func TestMonstersStayInPlayArea(t *testing.T) {
	const size, margin = 12000, 420
	cfgGame := quietGame(t, playMap(size, margin))
	cfgGame.Cfg.Room.MonsterCount = 20
	for i := 0; i < 20; i++ {
		cfgGame.spawnMonster()
	}
	if len(cfgGame.Monsters) == 0 {
		t.Fatal("không sinh được quái nào: test không kiểm được gì")
	}
	for i := 0; i < 30*30; i++ {
		cfgGame.Tick(dt)
	}
	for _, m := range cfgGame.Monsters {
		lo, hi := cfgGame.PlayBounds(m.R)
		if m.Pos.X < lo-0.5 || m.Pos.X > hi+0.5 || m.Pos.Y < lo-0.5 || m.Pos.Y > hi+0.5 {
			t.Fatalf("quái %s ra ngoài tường: (%.0f,%.0f), vùng [%.0f,%.0f]", m.Kind, m.Pos.X, m.Pos.Y, lo, hi)
		}
	}
}
