// Package bot drives AI players (spec §10). Bots only use what a player at their position could see.
package bot

import (
	"math"
	"math/rand"

	"kiemxoay/server/internal/game"
	"kiemxoay/server/internal/geom"
)

type Vec = geom.Vec

// States of spec §10.
const (
	StateFarm   = "farm"
	StateHunt   = "hunt"
	StateFlee   = "flee"
	StateAmbush = "ambush"
	StateWander = "wander"
)

const (
	farmSearchR   = 1800.0
	pathMaxExpand = 2500
	repathSecs    = 1.5
	waypointReach = 60.0
)

// Brain is one bot's AI state and personality.
type Brain struct {
	ID         uint32
	State      string
	rng        *rand.Rand
	nextDecide float64
	reaction   float64
	aimNoise   float64
	aggression float64

	goal        Vec
	hasGoal     bool
	stopNear    float64 // stop moving within this distance of the goal
	path        []Vec
	pathFor     Vec
	pathAt      float64
	ambushUntil float64
}

func New(id uint32, seed int64, g *game.Game) *Brain {
	r := rand.New(rand.NewSource(seed))
	bc := g.Cfg.Bots
	return &Brain{
		ID: id, State: StateWander, rng: r,
		reaction:   bc.ReactionMin + r.Float64()*(bc.ReactionMax-bc.ReactionMin),
		aimNoise:   r.Float64() * 0.35,
		aggression: 0.3 + r.Float64()*0.7,
	}
}

// Update chooses an input for this tick and applies it to the game.
func (b *Brain) Update(g *game.Game) {
	me, ok := g.Players[b.ID]
	if !ok || me.Dead {
		return
	}
	if g.Time >= b.nextDecide {
		bc := g.Cfg.Bots
		b.decide(g, me)
		b.nextDecide = g.Time + bc.DecideMin + b.rng.Float64()*(bc.DecideMax-bc.DecideMin) + b.reaction*0.3
	}
	angle, moving := b.steer(g, me)
	g.SetInput(b.ID, 0, angle, moving)
}

func (b *Brain) setGoal(p Vec, stopNear float64) {
	b.goal, b.hasGoal, b.stopNear = p, true, stopNear
}

func (b *Brain) decide(g *game.Game, me *game.Player) {
	bc := g.Cfg.Bots
	myPow := g.Power(me)
	threat, prey := b.scanPlayers(g, me, myPow)
	switch {
	case threat != nil:
		b.State = StateFlee
		b.flee(g, me, threat)
	case prey != nil && b.rng.Float64() < 0.5+b.aggression/2:
		b.State = StateHunt
		// Aim so the target body sits on our sword ring.
		b.setGoal(prey.Pos, me.Geom.Reach*0.6)
	case b.State == StateAmbush && g.Time < b.ambushUntil:
		// keep waiting in the bush
	case b.shouldAmbush(g, me):
		b.State = StateAmbush
		b.ambushUntil = g.Time + bc.AmbushSecs
	default:
		b.farm(g, me)
	}
}

// scanPlayers returns the nearest visible threat and the nearest visible prey.
func (b *Brain) scanPlayers(g *game.Game, me *game.Player, myPow float64) (*game.Player, *game.Player) {
	bc := g.Cfg.Bots
	var threat, prey *game.Player
	bestT, bestP := math.Inf(1), math.Inf(1)
	for _, o := range g.Players {
		if o == me || !g.CanSee(me, o) {
			continue
		}
		d := o.Pos.Dist(me.Pos)
		pow := g.Power(o)
		if pow > myPow*bc.FleeRatio && d < bc.FleeR*me.Scale && d < bestT {
			threat, bestT = o, d
		}
		if pow < myPow*bc.HuntRatio && d < bc.HuntR*me.Scale && d < bestP && !o.Invulnerable(g.Time) {
			prey, bestP = o, d
		}
	}
	return threat, prey
}

func (b *Brain) flee(g *game.Game, me, threat *game.Player) {
	away := me.Pos.Sub(threat.Pos).Norm()
	// Prefer a bush or a gate roughly away from the threat (spec §10 Flee).
	if spot, ok := b.nearestCover(g, me, g.Cfg.Bots.AmbushR*2); ok && spot.Sub(me.Pos).Norm().Dot(away) > 0.2 {
		b.setGoal(spot, 0)
		return
	}
	b.setGoal(me.Pos.Add(away.Mul(800)), 0)
}

// nearestCover finds the closest bush centre or gate within r.
func (b *Brain) nearestCover(g *game.Game, me *game.Player, r float64) (Vec, bool) {
	best, bestD := Vec{}, r
	found := false
	for _, z := range g.Map.Zones {
		if z.Kind != "bush" {
			continue
		}
		if c := z.Shape.Center(); c.Dist(me.Pos) < bestD {
			best, bestD, found = c, c.Dist(me.Pos), true
		}
	}
	for _, gt := range g.Map.Gates {
		c := Vec{X: gt.X, Y: gt.Y}
		if c.Dist(me.Pos) < bestD && g.Time >= me.GateReadyAt {
			best, bestD, found = c, c.Dist(me.Pos), true
		}
	}
	return best, found
}

func (b *Brain) shouldAmbush(g *game.Game, me *game.Player) bool {
	if g.Prog.Realm(me.Level) > 0 || b.rng.Float64() > 0.15*(1-b.aggression+0.3) {
		return false
	}
	spot, ok := b.nearestCover(g, me, g.Cfg.Bots.AmbushR)
	if ok {
		b.setGoal(spot, 20)
	}
	return ok
}

// farm heads to the nearest monster, item or weak obstacle (spec §10 Farm).
func (b *Brain) farm(g *game.Game, me *game.Player) {
	target, stop, ok := b.nearestFarmTarget(g, me)
	if !ok {
		b.State = StateWander
		if !b.hasGoal || me.Pos.Dist(b.goal) < 200 {
			ang := b.rng.Float64() * 2 * math.Pi
			b.setGoal(me.Pos.Add(geom.FromAngle(ang).Mul(1500)), 0)
		}
		return
	}
	b.State = StateFarm
	b.setGoal(target, stop)
}

func (b *Brain) nearestFarmTarget(g *game.Game, me *game.Player) (Vec, float64, bool) {
	best, stop, bestD := Vec{}, 0.0, farmSearchR*math.Sqrt(me.Scale)
	found := false
	for _, it := range g.Items {
		if d := it.Pos.Dist(me.Pos); d < bestD*0.8 && g.InView(me, it.Pos) {
			best, stop, bestD, found = it.Pos, 0, d/0.8, true
		}
	}
	for _, m := range g.Monsters {
		if d := m.Pos.Dist(me.Pos); d < bestD && g.InView(me, m.Pos) {
			best, stop, bestD, found = m.Pos, me.Geom.Reach*0.8, d, true
		}
	}
	for id, o := range g.Map.Obstacles {
		st := g.Obstacles[id]
		if o.MaxHP == 0 || st.Broken || o.MaxHP > 60*(1+g.Prog.Realm(me.Level)) {
			continue
		}
		c := o.Shape.Center()
		if d := c.Dist(me.Pos); d < bestD*1.2 {
			best, stop, bestD, found = c, me.Geom.Reach*0.9, d/1.2, true
		}
	}
	return best, stop, found
}

// steer turns the current goal into a movement angle, using A* when the way is blocked.
func (b *Brain) steer(g *game.Game, me *game.Player) (float64, bool) {
	if !b.hasGoal || b.State == StateAmbush && me.Pos.Dist(b.goal) < 40 {
		return 0, false
	}
	d := me.Pos.Dist(b.goal)
	if d <= b.stopNear {
		// Orbit slightly so the ring sweeps over the target.
		return b.goal.Sub(me.Pos).Angle() + math.Pi/2, true
	}
	next := b.nextWaypoint(g, me)
	ang := next.Sub(me.Pos).Angle() + (b.rng.Float64()-0.5)*b.aimNoise
	return ang, true
}

func (b *Brain) nextWaypoint(g *game.Game, me *game.Player) Vec {
	if b.clearLine(g, me.Pos, b.goal) {
		b.path = nil
		return b.goal
	}
	if b.path == nil || b.pathFor.Dist(b.goal) > 200 || g.Time-b.pathAt > repathSecs {
		b.path = g.Grid.FindPath(me.Pos, b.goal, pathMaxExpand)
		b.pathFor, b.pathAt = b.goal, g.Time
	}
	for len(b.path) > 0 && me.Pos.Dist(b.path[0]) < waypointReach*math.Sqrt(me.Scale) {
		b.path = b.path[1:]
	}
	if len(b.path) == 0 {
		return b.goal
	}
	return b.path[0]
}

// clearLine samples the path grid between two points.
func (b *Brain) clearLine(g *game.Game, from, to Vec) bool {
	d := from.Dist(to)
	steps := int(d / (g.Grid.Cell / 2))
	for i := 1; i <= steps; i++ {
		if g.Grid.Blocked(from.Add(to.Sub(from).Mul(float64(i) / float64(steps)))) {
			return false
		}
	}
	return true
}
