// Package game is the authoritative rules engine: entities, tick, collisions, EXP, swords, buffs.
package game

import (
	"math"
	"math/rand"
	"sort"

	"kiemxoay/server/internal/config"
	"kiemxoay/server/internal/world"
)

// Event is sent to clients (spec §13.3). Kind names match the spec's event list.
type Event struct {
	Kind   string  `msgpack:"k"`
	A      uint32  `msgpack:"a,omitempty"`
	B      uint32  `msgpack:"b,omitempty"`
	X      float32 `msgpack:"x"`
	Y      float32 `msgpack:"y"`
	N      int     `msgpack:"n,omitempty"`
	S      string  `msgpack:"s,omitempty"`
	Global bool    `msgpack:"-"` // broadcast to the whole room regardless of AOI
	To     uint32  `msgpack:"-"` // when set, only this player receives it
}

// Game owns one room's world state.
type Game struct {
	Cfg  *config.Config
	Prog *Progression
	Map  *world.Map
	Rng  *rand.Rand
	Time float64

	Players   map[uint32]*Player
	Monsters  map[uint32]*Monster
	Items     map[uint32]*Item
	Obstacles []ObstacleState
	Grid      *world.Grid
	Events    []Event

	nextID       uint32
	order        []uint32 // player ids in join order, for deterministic iteration
	cooldowns    map[uint64]float64
	obsIndex     *staticIndex
	gridDirty    bool
	chestNextAt  float64
	chestAt      *Vec
	poolCount    int
	itemRespawns []float64
	orbRespawns  []pendingOrb
	itemGrid     *itemGrid
	orbCount     int
	bossCount    int
}

func New(cfg *config.Config, m *world.Map, seed int64) *Game {
	g := &Game{
		Cfg: cfg, Prog: NewProgression(cfg.Cultivation, cfg.Sword.MaxSwords), Map: m,
		Rng:       rand.New(rand.NewSource(seed)),
		Players:   map[uint32]*Player{},
		Monsters:  map[uint32]*Monster{},
		Items:     map[uint32]*Item{},
		cooldowns: map[uint64]float64{},
	}
	g.Obstacles = make([]ObstacleState, len(m.Obstacles))
	for i, o := range m.Obstacles {
		g.Obstacles[i] = ObstacleState{HP: o.MaxHP}
	}
	g.obsIndex = newStaticIndex(m, 500)
	g.Grid = world.NewGrid(m, cfg.Bots.PathGrid)
	g.Grid.Rebuild(g.ObstacleSolid)
	g.chestNextAt = cfg.Items.ChestEverySecs
	g.itemGrid = newItemGrid(m.Size, 500)
	g.seedOrbs()
	return g
}

func (g *Game) newID() uint32 {
	g.nextID++
	return g.nextID
}

func (g *Game) emit(e Event) { g.Events = append(g.Events, e) }

// DrainEvents returns and clears this tick's events.
func (g *Game) DrainEvents() []Event {
	ev := g.Events
	g.Events = nil
	return ev
}

// PlayerList returns players in join order.
func (g *Game) PlayerList() []*Player {
	out := make([]*Player, 0, len(g.order))
	for _, id := range g.order {
		if p, ok := g.Players[id]; ok {
			out = append(out, p)
		}
	}
	return out
}

// SpawnOptions controls how a new player enters the world.
type SpawnOptions struct {
	Name, Char, GuestID string
	IsBot               bool
	Level               int
	Swords              int
}

func (g *Game) AddPlayer(o SpawnOptions) *Player {
	p := &Player{ID: g.newID(), Name: o.Name, Char: o.Char, GuestID: o.GuestID, IsBot: o.IsBot,
		EmoteReady: map[string]float64{}, Bush: -1}
	g.Players[p.ID] = p
	g.order = append(g.order, p.ID)
	g.resetPlayer(p, o.Level, o.Swords)
	return p
}

// resetPlayer (re)spawns a player at a safe spot (spec §7).
func (g *Game) resetPlayer(p *Player, level, swords int) {
	p.TotalExp = g.Prog.Threshold(level)
	p.Level = level
	p.PeakExp, p.PeakLevel, p.Kills = p.TotalExp, level, 0
	// A new life starts from zero Uy Danh (spec §11.1).
	p.Score, p.Hits, p.Top1Secs, p.DaiThuaSecs = 0, 0, 0, 0
	p.ScoreBy, p.HitCooldown, p.ScoredRealms = map[string]float64{}, map[string]float64{}, map[int]bool{}
	p.Title, p.pendingTitle, p.titleSince = 0, 0, 0
	p.Ring = Ring{Count: max(g.Cfg.Sword.StartCount, swords), Dir: 1, Angle: g.Rng.Float64() * 2 * math.Pi}
	p.Ring.Tilt = RestTilt(g.Cfg.Sword, p.Ring.Dir)
	p.Buffs = [BuffCount]float64{}
	p.Dead, p.KilledBy, p.Transit = false, 0, nil
	p.Pos = g.safeSpawn()
	p.SpawnedAt = g.Time
	p.InvulnUntil = g.Time + g.Cfg.Player.InvulnOnSpawn
	g.setMood(p, MoodRespawn, g.Cfg.Mood.RespawnSecs)
	g.refreshDerived(p)
}

// Respawn puts a dead player back at Luyện Khí 1 with 1 sword.
func (g *Game) Respawn(p *Player, level, swords int) { g.resetPlayer(p, level, swords) }

func (g *Game) RemovePlayer(id uint32) {
	delete(g.Players, id)
	for i, o := range g.order {
		if o == id {
			g.order = append(g.order[:i], g.order[i+1:]...)
			break
		}
	}
}

// SetInput applies a client input (already validated by the caller).
func (g *Game) SetInput(id uint32, seq uint32, angle float64, moving bool) {
	p, ok := g.Players[id]
	if !ok {
		return
	}
	p.InputAngle, p.Moving = angle, moving
	if seq > p.LastSeq {
		p.LastSeq = seq
	}
}

// refreshDerived recomputes scale and ring geometry from level, sword count and buffs.
func (g *Game) refreshDerived(p *Player) {
	p.Scale = g.Prog.Scale(p.Level)
	long := 1.0
	if p.HasBuff(BuffLong, g.Time) {
		long = g.Cfg.Buffs["long"].Mult
	}
	p.Geom = ComputeRingGeom(g.Cfg.Sword, g.Cfg.Player.BodyR, p.Scale, p.Ring.Count, long)
}

func (g *Game) BodyR(p *Player) float64 { return g.Cfg.Player.BodyR * p.Scale }

// Omega is the spin rate: one revolution per rev_secs, × Tốc Kiếm.
func (g *Game) Omega(p *Player) float64 {
	w := 2 * math.Pi / math.Max(g.Cfg.Sword.RevSecs, 1e-3)
	if p.HasBuff(BuffSpin, g.Time) {
		w *= g.Cfg.Buffs["spin"].Mult
	}
	return w
}

// Tick advances the simulation by dt seconds in the order of spec §5.
func (g *Game) Tick(dt float64) {
	g.Time += dt
	g.updateTimers()
	g.movePlayers(dt)
	g.moveMonsters(dt)
	for _, p := range g.PlayerList() {
		if !p.Dead {
			g.spinRing(p, dt)
		}
	}
	g.resolveSwordBody()
	g.resolveSwordSword()
	g.resolveSwordWorld()
	g.resolvePickups()
	g.maintainSpawns()
	g.updateMoods()
	g.updateTop1(dt)
	g.updateTitles()
	if g.gridDirty {
		g.Grid.Rebuild(g.ObstacleSolid)
		g.gridDirty = false
	}
	g.pruneCooldowns()
}

// spinRing advances the ring and records the arc it swept this tick (spec §4.1b).
func (g *Game) spinRing(p *Player, dt float64) {
	r := &p.Ring
	step := r.Dir * g.Omega(p) * dt
	r.SweepLo, r.Sweep = r.Angle, math.Abs(step)
	if step < 0 {
		r.SweepLo = r.Angle + step
	}
	r.Angle = math.Mod(r.Angle+step, 2*math.Pi)
	stepTilt(g.Cfg.Sword, r, dt)
}

// pairKey packs two ids and a tag into one cooldown key.
func pairKey(tag uint8, a, b uint32, sub int) uint64 {
	return uint64(tag)<<60 | uint64(sub&0xff)<<52 | uint64(a&0x3ffffff)<<26 | uint64(b&0x3ffffff)
}

// onCooldown checks a cooldown without claiming it; pair it with startCooldown when the
// action actually happens (cooldownReady does both at once).
func (g *Game) onCooldown(key uint64) bool {
	until, ok := g.cooldowns[key]
	return ok && until > g.Time
}

func (g *Game) startCooldown(key uint64, secs float64) { g.cooldowns[key] = g.Time + secs }

func (g *Game) cooldownReady(key uint64, secs float64) bool {
	if until, ok := g.cooldowns[key]; ok && until > g.Time {
		return false
	}
	g.cooldowns[key] = g.Time + secs
	return true
}

func (g *Game) pruneCooldowns() {
	if int(g.Time*10)%30 != 0 {
		return
	}
	for k, v := range g.cooldowns {
		if v <= g.Time {
			delete(g.cooldowns, k)
		}
	}
}

// Ranking returns living players sorted by total_exp descending.
func (g *Game) Ranking() []*Player {
	var out []*Player
	for _, p := range g.PlayerList() {
		if !p.Dead {
			out = append(out, p)
		}
	}
	// Spec §11.2: the room board ranks by current realm+tier so players can tell at a glance
	// who is dangerous; total_exp only breaks ties within a tier.
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Level != out[j].Level {
			return out[i].Level > out[j].Level
		}
		return out[i].TotalExp > out[j].TotalExp
	})
	return out
}

// Vec is a convenience constructor for other packages' tests.
func (g *Game) Vec(x, y float64) Vec { return Vec{X: x, Y: y} }
