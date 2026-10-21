package game

import (
	"math"

	"kiemxoay/server/internal/geom"
)

const (
	tagMonster uint8 = iota + 1
	tagObstacle
	tagClash
)

// canFight is true when a player can deal and receive sword interactions this tick.
func (g *Game) canFight(p *Player) bool { return !p.Dead && !p.InTransit(g.Time) }

// ringReach returns the inner and outer radius covered by a ring's swords.
func ringReach(rg RingGeom) (float64, float64) {
	return rg.Radius - rg.Length/2 - rg.Width/2, rg.Radius + rg.Length/2 + rg.Width/2
}

// swordTouchesCircle returns true when any sword of ring owner touches circle (c, r).
func (g *Game) swordTouchesCircle(owner *Player, c Vec, r float64) (int, bool) {
	inner, outer := ringReach(owner.Geom)
	d := owner.Pos.Dist(c)
	if d+r < inner || d-r > outer {
		return 0, false
	}
	reach := (r + owner.Geom.Width/2) * (r + owner.Geom.Width/2)
	for i := 0; i < owner.Ring.Count; i++ {
		a, b := SwordSegment(owner.Pos, owner.Ring, owner.Geom, i)
		if geom.PointSegDist2(c, a, b) <= reach {
			return i, true
		}
	}
	return 0, false
}

// resolveSwordBody applies spec §5.1: at most one hit per victim per tick.
func (g *Game) resolveSwordBody() {
	players := g.PlayerList()
	for _, b := range players {
		if !g.canFight(b) || b.Invulnerable(g.Time) {
			continue
		}
		for _, a := range players {
			if a == b || !g.canFight(a) {
				continue
			}
			if _, hit := g.swordTouchesCircle(a, b.Pos, g.BodyR(b)); hit {
				g.applyBodyHit(a, b)
				break
			}
		}
	}
}

func (g *Game) applyBodyHit(a, b *Player) {
	pc := g.Cfg.Player
	a.Buffs[BuffStealth] = 0 // Ẩn Thân ends as soon as the owner's sword hits someone
	b.InvulnUntil = g.Time + pc.InvulnOnHit
	b.RevealedUntil = g.Time + g.Cfg.Zones.BushRevealOnHitSecs
	push := b.Pos.Sub(a.Pos).Norm()
	b.KnockVel = push.Mul(pc.Knockback / pc.KnockbackSecs)
	b.KnockUntil = g.Time + pc.KnockbackSecs
	ev := Event{Kind: "Hit", A: a.ID, B: b.ID, X: float32(b.Pos.X), Y: float32(b.Pos.Y)}
	if b.HasBuff(BuffShield, g.Time) {
		b.Buffs[BuffShield] = 0
		ev.S = "shield"
		g.emit(ev)
		g.emote(b, "shield_break")
		return
	}
	g.emit(ev)
	if b.Level == 0 {
		g.kill(a, b)
		return
	}
	g.setMood(b, MoodHit, g.Cfg.Mood.HitSecs)
	oldRealm := g.Prog.Realm(b.Level)
	loss := g.Prog.ExpNeeded(b.Level - 1)
	g.setTotalExp(b, b.TotalExp-loss)
	g.GainExp(a, loss*pc.AttackerShare, true)
	g.scatterOrbs(b.Pos, loss*pc.DropShare, 0)
	if g.Prog.Realm(b.Level) < oldRealm {
		g.emit(Event{Kind: "KillFeed", Global: true, S: "realm_drop", A: a.ID, B: b.ID, N: g.Prog.Realm(b.Level)})
	}
}

// kill handles a Luyện Khí 1 player being cut down (spec §7).
func (g *Game) kill(a, b *Player) {
	pc := g.Cfg.Player
	// Spec §5.1: at Luyện Khí 1 the "loss" is the victim's remaining EXP, plus a kill bonus.
	loss := b.TotalExp
	g.GainExp(a, loss*pc.AttackerShare+pc.KillBonusExp, true)
	g.scatterOrbs(b.Pos, loss*pc.DropShare, 0)
	g.dropDeathSwords(b)
	a.Kills++
	b.Dead, b.DiedAt, b.KilledBy = true, g.Time, a.ID
	b.KillerName, b.KillerChar = a.Name, a.Char
	b.Moving = false
	g.setMood(b, MoodDeath, pc.DeathAnimSecs)
	g.emit(Event{Kind: "Death", A: a.ID, B: b.ID, X: float32(b.Pos.X), Y: float32(b.Pos.Y)})
	feed := "kill"
	if b.lastTeleportIn > 0 && g.Time-b.lastTeleportIn < g.Cfg.Gates.FeedSecs {
		feed = "gate_kill"
	}
	g.emit(Event{Kind: "KillFeed", Global: true, S: feed, A: a.ID, B: b.ID})
}

// DebugKill kills a player without an attacker (debug key only).
func (g *Game) DebugKill(b *Player) {
	g.dropDeathSwords(b)
	b.Dead, b.DiedAt, b.KilledBy, b.Moving = true, g.Time, 0, false
	b.KillerName, b.KillerChar = "", ""
	g.setMood(b, MoodDeath, g.Cfg.Player.DeathAnimSecs)
	g.emit(Event{Kind: "Death", B: b.ID, X: float32(b.Pos.X), Y: float32(b.Pos.Y)})
}

// resolveSwordSword applies spec §5.2 to every pair of overlapping rings.
func (g *Game) resolveSwordSword() {
	players := g.PlayerList()
	for i, a := range players {
		if !g.canFight(a) {
			continue
		}
		for _, b := range players[i+1:] {
			if !g.canFight(b) {
				continue
			}
			if at, ok := g.ringsTouch(a, b); ok && g.cooldownReady(pairKey(tagClash, a.ID, b.ID, 0), g.Cfg.Sword.FlipCooldown) {
				g.clash(a, b, at)
			}
		}
	}
}

// ringsTouch finds a contact point between any sword of a and any sword of b.
func (g *Game) ringsTouch(a, b *Player) (Vec, bool) {
	_, outA := ringReach(a.Geom)
	_, outB := ringReach(b.Geom)
	d := a.Pos.Dist(b.Pos)
	if d > outA+outB {
		return Vec{}, false
	}
	reach := (a.Geom.Width + b.Geom.Width) / 2
	for i := 0; i < a.Ring.Count; i++ {
		p1, q1 := SwordSegment(a.Pos, a.Ring, a.Geom, i)
		if !nearRing(b, p1.Add(q1).Mul(0.5), a.Geom.Length/2+reach) {
			continue
		}
		for j := 0; j < b.Ring.Count; j++ {
			p2, q2 := SwordSegment(b.Pos, b.Ring, b.Geom, j)
			if geom.SegSegDist2(p1, q1, p2, q2) <= reach*reach {
				return p1.Add(q1).Add(p2).Add(q2).Mul(0.25), true
			}
		}
	}
	return Vec{}, false
}

// nearRing is a cheap test: can a point within slack of p touch any sword of owner?
func nearRing(owner *Player, p Vec, slack float64) bool {
	inner, outer := ringReach(owner.Geom)
	d := owner.Pos.Dist(p)
	return d+slack >= inner && d-slack <= outer
}

func (g *Game) clash(a, b *Player, at Vec) {
	a.Ring.Dir, b.Ring.Dir = -a.Ring.Dir, -b.Ring.Dir
	for _, p := range []*Player{a, b} {
		g.emit(Event{Kind: "RingFlip", A: p.ID, N: int(p.Ring.Dir), X: float32(at.X), Y: float32(at.Y)})
	}
	out := ResolveClash(g.Cfg.Sword,
		ClashSide{Realm: g.Prog.Realm(a.Level), Count: a.Ring.Count, Crack: a.Ring.Crack},
		ClashSide{Realm: g.Prog.Realm(b.Level), Count: b.Ring.Count, Crack: b.Ring.Crack})
	a.Ring.Crack, b.Ring.Crack = out.CrackA, out.CrackB
	g.breakSwords(a, out.LostA)
	g.breakSwords(b, out.LostB)
}

func (g *Game) breakSwords(p *Player, n int) {
	if n <= 0 {
		return
	}
	lost := g.LoseSwords(p, n)
	if lost > 0 {
		g.emit(Event{Kind: "SwordBreak", A: p.ID, N: lost, X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
		g.emote(p, "sword_lost")
	}
}

// resolveSwordWorld applies spec §5.3: swords vs monsters and obstacles.
func (g *Game) resolveSwordWorld() {
	for _, a := range g.PlayerList() {
		if !g.canFight(a) {
			continue
		}
		g.hitMonsters(a)
		g.hitObstacles(a)
	}
}

func (g *Game) hitMonsters(a *Player) {
	_, outer := ringReach(a.Geom)
	for _, m := range g.Monsters {
		if m.Pos.Dist(a.Pos) > outer+m.R {
			continue
		}
		i, hit := g.swordTouchesCircle(a, m.Pos, m.R)
		if !hit || !g.cooldownReady(pairKey(tagMonster, a.ID, m.ID, i), g.Cfg.Monsters.HitCooldown) {
			continue
		}
		m.HP--
		m.FleeUntil, m.FleeFrom = g.Time+g.Cfg.Monsters.FleeSecs, a.Pos
		if m.HP <= 0 {
			g.killMonster(m, a)
			continue
		}
		g.emit(Event{Kind: "MonsterHurt", A: m.ID, N: m.HP, X: float32(m.Pos.X), Y: float32(m.Pos.Y)})
	}
}

// killMonster drops the monster's EXP as orbs reserved for the killer for orb_owner_secs (spec §5.3).
func (g *Game) killMonster(m *Monster, killer *Player) {
	delete(g.Monsters, m.ID)
	if m.Kind == "boss" {
		g.bossCount--
	}
	g.emit(Event{Kind: "MonsterDie", A: m.ID, S: m.Kind, X: float32(m.Pos.X), Y: float32(m.Pos.Y)})
	g.scatterOrbs(m.Pos, m.Exp, killer.ID)
	k := g.Cfg.Monsters.Kinds[m.Kind]
	if g.Rng.Float64() < k.Drop {
		for i := 0; i < k.DropItems; i++ {
			g.spawnItemNear(g.randomItemKind(), 0, m.Pos, m.R+40)
		}
	}
}

func (g *Game) hitObstacles(a *Player) {
	_, outer := ringReach(a.Geom)
	dmg := 1 + g.Prog.Realm(a.Level)
	g.QueryObstacles(a.Pos, outer, func(id int) {
		st := &g.Obstacles[id]
		o := g.Map.Obstacles[id]
		if o.MaxHP == 0 || st.Broken {
			return
		}
		for i := 0; i < a.Ring.Count; i++ {
			p, q := SwordSegment(a.Pos, a.Ring, a.Geom, i)
			if !o.Shape.SegmentHits(p, q, a.Geom.Width/2) {
				continue
			}
			if g.cooldownReady(pairKey(tagObstacle, a.ID, uint32(id), i), g.Cfg.Obstacles.HitCooldown) {
				g.DamageObstacle(id, dmg)
			}
		}
	})
}

// DamageObstacle reduces an obstacle's HP and breaks it at zero (spec §8.2).
func (g *Game) DamageObstacle(id, dmg int) {
	st := &g.Obstacles[id]
	o := g.Map.Obstacles[id]
	if o.MaxHP == 0 || st.Broken {
		return
	}
	st.HP = max(0, st.HP-dmg)
	c := o.Shape.Center()
	if st.HP > 0 {
		g.emit(Event{Kind: "ObstacleDamaged", A: uint32(id), N: st.HP, X: float32(c.X), Y: float32(c.Y)})
		return
	}
	st.Broken, st.BrokenAt, st.Warned = true, g.Time, false
	g.gridDirty = true
	g.emit(Event{Kind: "ObstacleBroken", A: uint32(id), X: float32(c.X), Y: float32(c.Y)})
	if (o.Kind == "stall" || o.Kind == "cart") && g.Rng.Float64() < g.Cfg.Obstacles.StallItemDrop {
		g.spawnItemNear(g.randomItemKind(), 0, c, 30)
	}
}

// updateObstacles regrows broken obstacles once their area is clear (spec §8.2).
func (g *Game) updateObstacles() {
	oc := g.Cfg.Obstacles
	for id := range g.Obstacles {
		st := &g.Obstacles[id]
		if !st.Broken {
			continue
		}
		c := g.Map.Obstacles[id].Shape.Center()
		if !st.Warned && g.Time >= st.BrokenAt+oc.RegrowSecs-oc.RegrowWarnSecs {
			st.Warned = true
			g.emit(Event{Kind: "ObstacleRegrow", A: uint32(id), S: "warn", X: float32(c.X), Y: float32(c.Y)})
		}
		if g.Time < st.BrokenAt+oc.RegrowSecs || g.obstacleOccupied(id) {
			continue
		}
		st.Broken, st.HP = false, g.Map.Obstacles[id].MaxHP
		g.gridDirty = true
		g.emit(Event{Kind: "ObstacleRegrow", A: uint32(id), X: float32(c.X), Y: float32(c.Y)})
	}
}

// obstacleOccupied reports whether any body or monster stands where obstacle id would regrow.
func (g *Game) obstacleOccupied(id int) bool {
	sh := g.Map.Obstacles[id].Shape
	for _, p := range g.Players {
		if !p.Dead && sh.OverlapsCircle(p.Pos, g.BodyR(p)) {
			return true
		}
	}
	for _, m := range g.Monsters {
		if sh.OverlapsCircle(m.Pos, m.R) {
			return true
		}
	}
	return false
}

// dropDeathSwords spills floor(n * death_sword_drop) swords around a corpse (spec §7).
func (g *Game) dropDeathSwords(b *Player) {
	n := int(math.Floor(float64(b.Ring.Count) * g.Cfg.Player.DeathSwordDrop))
	b.Ring.Count -= n
	for i := 0; i < n; i++ {
		g.spawnItemNear("sword", 0, b.Pos, b.Geom.Radius)
	}
}

// scatterOrbs drops total EXP as orbs around pos, largest sizes first. A non-zero owner
// reserves the orbs for that player for orb_owner_secs.
func (g *Game) scatterOrbs(pos Vec, total float64, owner uint32) {
	vals := g.Cfg.Items.OrbValues
	sizes := []float64{vals["large"], vals["medium"], vals["small"]}
	remaining := math.Round(total)
	for _, v := range sizes {
		for remaining >= v && v > 0 && g.orbCount < g.Cfg.Items.MaxOrbs {
			g.ownItem(g.spawnItemNear("orb", v, pos, g.Cfg.Items.ScatterR), owner)
			remaining -= v
		}
	}
	if remaining > 0 && g.orbCount < g.Cfg.Items.MaxOrbs {
		g.ownItem(g.spawnItemNear("orb", remaining, pos, g.Cfg.Items.ScatterR), owner)
	}
}

func (g *Game) ownItem(it *Item, owner uint32) {
	if owner != 0 {
		it.Owner, it.OwnerUntil = owner, g.Time+g.Cfg.Player.OrbOwnerSecs
	}
}
