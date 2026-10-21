package game

import (
	"math"
	"testing"

	"kiemxoay/server/internal/config"
	"kiemxoay/server/internal/world"
)

const dt = 1.0 / 30

// quietGame returns a game on an empty map with no random spawns.
func quietGame(t *testing.T, m *world.Map) *Game {
	t.Helper()
	cfg := config.MustLoadDefault()
	cfg.Room.MonsterCount = 0
	cfg.Items.CountPerRoom = 0
	cfg.Items.ChestEverySecs = 1e9
	if m == nil {
		m = world.EmptyMap(cfg.Room.MapSize)
	}
	return New(cfg, m, 1)
}

// place puts a player at pos with sword 0 pointing at angle, clearing spawn invulnerability.
func place(g *Game, p *Player, pos Vec, angle float64) {
	p.Pos, p.Ring.Angle, p.InvulnUntil = pos, angle, 0
	g.refreshDerived(p)
}

func hasEvent(evs []Event, kind string) bool {
	for _, e := range evs {
		if e.Kind == kind {
			return true
		}
	}
	return false
}

func TestBodyHitDropsOneTierAndRewards(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Level: 0, Swords: 1})
	b := g.AddPlayer(SpawnOptions{Name: "B", Level: 10, Swords: 1})
	place(g, a, Vec{X: 5000, Y: 5000}, 0)
	place(g, b, Vec{X: 5000 + a.Geom.Radius, Y: 5000}, math.Pi/2)
	before := b.TotalExp
	g.resolveSwordBody()
	loss := g.Prog.ExpNeeded(9)
	if b.Level != 9 || b.TotalExp != before-loss {
		t.Fatalf("B level %d exp %v, want level 9 exp %v", b.Level, b.TotalExp, before-loss)
	}
	if a.TotalExp != loss*0.5 {
		t.Fatalf("A exp = %v, want %v", a.TotalExp, loss*0.5)
	}
	if !b.Invulnerable(g.Time) {
		t.Fatal("B should be invulnerable after a hit")
	}
	if !hasEvent(g.DrainEvents(), "RealmDrop") {
		t.Fatal("Trúc Cơ 1 → Luyện Khí 10 must emit RealmDrop")
	}
	orbs := 0.0
	for _, it := range g.Items {
		if it.Kind == "orb" {
			orbs += it.Value
		}
	}
	if math.Abs(orbs-math.Round(loss*0.3)) > 1e-9 {
		t.Fatalf("dropped orb EXP = %v, want %v", orbs, math.Round(loss*0.3))
	}
}

func TestRealmDropSpillsExtraSwords(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Level: 0, Swords: 1})
	b := g.AddPlayer(SpawnOptions{Name: "B", Level: 10, Swords: 8})
	place(g, a, Vec{X: 5000, Y: 5000}, 0)
	place(g, b, Vec{X: 5000 + a.Geom.Radius, Y: 5000}, math.Pi/16)
	a.InvulnUntil = 1e9 // B's larger ring would otherwise hit A first
	g.resolveSwordBody()
	if b.Ring.Count != 5 {
		t.Fatalf("B swords = %d, want 5 (Luyện Khí max)", b.Ring.Count)
	}
	swords := 0
	for _, it := range g.Items {
		if it.Kind == "sword" {
			swords++
		}
	}
	if swords != 3 {
		t.Fatalf("sword items on ground = %d, want 3", swords)
	}
}

func TestLuyenKhi1Dies(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Level: 5, Swords: 1})
	b := g.AddPlayer(SpawnOptions{Name: "B", Level: 0, Swords: 1})
	place(g, a, Vec{X: 5000, Y: 5000}, 0)
	place(g, b, Vec{X: 5000 + a.Geom.Radius, Y: 5000}, math.Pi/2)
	g.resolveSwordBody()
	if !b.Dead || b.KilledBy != a.ID || a.Kills != 1 {
		t.Fatalf("B dead=%v killedBy=%d A kills=%d", b.Dead, b.KilledBy, a.Kills)
	}
	if !hasEvent(g.DrainEvents(), "Death") {
		t.Fatal("missing Death event")
	}
}

func TestShieldBlocksOneHit(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Swords: 1})
	b := g.AddPlayer(SpawnOptions{Name: "B", Level: 15, Swords: 1})
	place(g, a, Vec{X: 5000, Y: 5000}, 0)
	place(g, b, Vec{X: 5000 + a.Geom.Radius, Y: 5000}, math.Pi/2)
	g.ApplyBuff(b, "shield")
	g.resolveSwordBody()
	if b.Level != 15 || b.HasBuff(BuffShield, g.Time) || !b.Invulnerable(g.Time) {
		t.Fatalf("shield: level %d shield %v invuln %v", b.Level, b.HasBuff(BuffShield, g.Time), b.Invulnerable(g.Time))
	}
}

func TestOneHitPerVictimPerTick(t *testing.T) {
	g := quietGame(t, nil)
	b := g.AddPlayer(SpawnOptions{Name: "B", Level: 20, Swords: 1})
	place(g, b, Vec{X: 5000, Y: 5000}, math.Pi/2)
	for _, x := range []float64{-1, 1} {
		a := g.AddPlayer(SpawnOptions{Name: "A", Swords: 1})
		ang := math.Pi // at the right of B, point left
		if x > 0 {
			ang = 0
		}
		place(g, a, Vec{X: 5000 - x*a.Geom.Radius, Y: 5000}, ang)
	}
	g.resolveSwordBody()
	if b.Level != 19 {
		t.Fatalf("B level = %d, want 19 (only one hit per tick)", b.Level)
	}
}

func TestClashFlipsBothRings(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Swords: 3})
	b := g.AddPlayer(SpawnOptions{Name: "B", Swords: 3})
	place(g, a, Vec{X: 5000, Y: 5000}, 0)
	// Ring B centred so its sword 0 (pointing left) overlaps A's sword 0 (pointing right).
	place(g, b, Vec{X: 5000 + 2*a.Geom.Radius, Y: 5000}, math.Pi)
	g.resolveSwordSword()
	if a.Ring.Dir != -1 || b.Ring.Dir != -1 {
		t.Fatalf("dirs = %v, %v; want both flipped", a.Ring.Dir, b.Ring.Dir)
	}
	if a.Ring.Count != 2 || b.Ring.Count != 2 {
		t.Fatalf("same realm should lose 1 each: %d, %d", a.Ring.Count, b.Ring.Count)
	}
	g.resolveSwordSword() // cooldown 0.3s prevents an immediate second clash
	if a.Ring.Dir != -1 {
		t.Fatal("clash cooldown not respected")
	}
}

func TestSwordPickupOvercapGivesExp(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "P", Swords: 5})
	g.AddSword(p)
	if p.Ring.Count != 5 || p.TotalExp != 20 {
		t.Fatalf("count %d exp %v, want 5 swords and +20 EXP", p.Ring.Count, p.TotalExp)
	}
	q := g.AddPlayer(SpawnOptions{Name: "Q", Swords: 1})
	g.AddSword(q)
	if q.Ring.Count != 2 {
		t.Fatalf("count %d, want 2", q.Ring.Count)
	}
}

func TestBuffExpiresAndRefreshes(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "P", Swords: 1})
	g.ApplyBuff(p, "haste")
	for g.Time < 5 {
		g.Tick(dt)
	}
	g.ApplyBuff(p, "haste") // refresh, not stack
	if exp := p.Buffs[BuffHaste]; math.Abs(exp-(g.Time+10)) > 1e-9 {
		t.Fatalf("haste expiry %v, want %v", exp, g.Time+10)
	}
	for g.Time < 14.9 {
		g.Tick(dt)
	}
	if !p.HasBuff(BuffHaste, g.Time) {
		t.Fatal("haste ended early")
	}
	for g.Time < 15.2 {
		g.Tick(dt)
	}
	if p.HasBuff(BuffHaste, g.Time) {
		t.Fatal("haste did not expire")
	}
}

func TestExp2DoublesGains(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "P", Swords: 1})
	g.ApplyBuff(p, "exp2")
	g.GainExp(p, 30, true)
	if p.TotalExp != 60 {
		t.Fatalf("exp = %v, want 60", p.TotalExp)
	}
}

func TestLongBuffExtendsRing(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "P", Swords: 1})
	r0 := p.Geom.Radius
	g.ApplyBuff(p, "long")
	if math.Abs(p.Geom.Radius-r0*1.3) > 1e-9 {
		t.Fatalf("radius %v, want %v", p.Geom.Radius, r0*1.3)
	}
}

func obstacleMap(kind string, hp int) *world.Map {
	m := world.EmptyMap(12000)
	m.Obstacles = []world.Obstacle{{ID: 0, Kind: kind, MaxHP: hp, Shape: world.Shape{X: 6000, Y: 6000, W: 200, H: 40}}}
	return m
}

func TestObstacleBreaksAndRegrowsWhenClear(t *testing.T) {
	g := quietGame(t, obstacleMap("fence", 30))
	g.DamageObstacle(0, 30)
	if !g.Obstacles[0].Broken || g.ObstacleSolid(0) {
		t.Fatal("fence should be broken and passable")
	}
	p := g.AddPlayer(SpawnOptions{Name: "P", Swords: 1})
	place(g, p, Vec{X: 6100, Y: 6020}, 0) // standing on the rubble
	for g.Time < 31 {
		g.Tick(dt)
		p.Pos = Vec{X: 6100, Y: 6020}
	}
	if !g.Obstacles[0].Broken {
		t.Fatal("obstacle regrew on top of a player")
	}
	p.Pos = Vec{X: 3000, Y: 3000}
	g.Tick(dt)
	if g.Obstacles[0].Broken || g.Obstacles[0].HP != 30 {
		t.Fatalf("obstacle should regrow with full HP: %+v", g.Obstacles[0])
	}
}

func TestSwordDamagesObstacleAndSurvives(t *testing.T) {
	g := quietGame(t, obstacleMap("brick", 150))
	p := g.AddPlayer(SpawnOptions{Name: "P", Level: 20, Swords: 1})
	place(g, p, Vec{X: 6100, Y: 6020 + 45 + 80}, -math.Pi/2)
	g.hitObstacles(p)
	if g.Obstacles[0].HP != 150-3 {
		t.Fatalf("brick HP = %d, want 147 (1 + realm 2)", g.Obstacles[0].HP)
	}
	if p.Ring.Count != 1 {
		t.Fatal("sword lost on obstacle")
	}
}

func TestBodyBlockedByObstacle(t *testing.T) {
	g := quietGame(t, obstacleMap("stone_wall", 0))
	p := g.AddPlayer(SpawnOptions{Name: "P", Swords: 1})
	place(g, p, Vec{X: 6100, Y: 5900}, 0)
	p.Moving, p.InputAngle = true, math.Pi/2
	for i := 0; i < 60; i++ {
		g.Tick(dt)
	}
	if p.Pos.Y > 6000-g.BodyR(p)+1 {
		t.Fatalf("player walked into the wall: y=%v", p.Pos.Y)
	}
}

func TestGateTeleportsWithoutInvulnerability(t *testing.T) {
	m := world.EmptyMap(12000)
	m.Gates = []world.Gate{{ID: 0, X: 2000, Y: 2000, Network: "a"}, {ID: 1, X: 9000, Y: 9000, Network: "a"}}
	g := quietGame(t, m)
	p := g.AddPlayer(SpawnOptions{Name: "P", Swords: 1})
	place(g, p, Vec{X: 2000, Y: 2000}, 0)
	g.Tick(dt)
	if !p.InTransit(g.Time) {
		t.Fatal("expected transit")
	}
	for g.Time < 1.0 {
		g.Tick(dt)
	}
	if p.Pos.Dist(Vec{X: 9000, Y: 9000}) > 400 {
		t.Fatalf("player at %+v, want near exit gate", p.Pos)
	}
	if p.Invulnerable(g.Time) || p.InTransit(g.Time) {
		t.Fatal("no invulnerability after leaving a gate")
	}
}

func TestMoveSpeedFormula(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "P", Level: 89, Swords: 1})
	if s := g.MoveSpeed(p); math.Abs(s-260*math.Sqrt(5)) > 1e-9 {
		t.Fatalf("speed = %v", s)
	}
}

func TestKillRewardAndDeathSwordDrop(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Level: 5, Swords: 1})
	b := g.AddPlayer(SpawnOptions{Name: "B", Level: 0, Swords: 5})
	place(g, a, Vec{X: 5000, Y: 5000}, 0)
	place(g, b, Vec{X: 5000 + a.Geom.Radius, Y: 5000}, math.Pi/2+math.Pi/5)
	a.InvulnUntil = 1e9
	b.TotalExp = 80
	aBefore := a.TotalExp
	g.resolveSwordBody()
	if !b.Dead {
		t.Fatal("B should die at Luyện Khí 1")
	}
	if got, want := a.TotalExp-aBefore, 80*0.5+30; math.Abs(got-want) > 1e-9 {
		t.Fatalf("killer gained %v, want %v (50%% + kill bonus)", got, want)
	}
	swords, orbs := 0, 0.0
	for _, it := range g.Items {
		switch it.Kind {
		case "sword":
			swords++
		case "orb":
			orbs += it.Value
		}
	}
	if swords != 2 {
		t.Fatalf("corpse dropped %d swords, want floor(5*0.5)=2", swords)
	}
	if orbs != 24 {
		t.Fatalf("dropped orb EXP %v, want 24 (30%% of 80)", orbs)
	}
}

func TestWindAngles(t *testing.T) {
	g := quietGame(t, nil)
	cases := []struct {
		deg  float64
		want float64
	}{{0, 1.5}, {45, 1.5}, {90, 1}, {119, 1}, {130, 0.8}, {180, 0.8}}
	for _, c := range cases {
		dir := Vec{X: math.Cos(c.deg * math.Pi / 180), Y: math.Sin(c.deg * math.Pi / 180)}
		if got := g.windMult(dir, 0); got != c.want {
			t.Errorf("%v° off the wind → %v, want %v", c.deg, got, c.want)
		}
	}
}

func TestMonsterOrbsReservedForKiller(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Swords: 1})
	b := g.AddPlayer(SpawnOptions{Name: "B", Swords: 1})
	place(g, a, Vec{X: 2000, Y: 2000}, 0)
	place(g, b, Vec{X: 7000, Y: 7000}, 0)
	m := &Monster{ID: g.newID(), Kind: "small", Pos: Vec{X: 7000, Y: 7000}, Exp: 10}
	g.killMonster(m, a)
	g.resolvePickups()
	if b.TotalExp != 0 {
		t.Fatalf("B stole reserved orbs: exp %v", b.TotalExp)
	}
	g.Time += g.Cfg.Player.OrbOwnerSecs + 0.01
	g.resolvePickups()
	if b.TotalExp != 10 {
		t.Fatalf("after the reservation B should collect 10 EXP, got %v", b.TotalExp)
	}
}
