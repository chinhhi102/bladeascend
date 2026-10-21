package game

import (
	"math"
	"sort"

	"kiemxoay/server/internal/geom"
)

const spawnsPerTick = 20

// randomItemKind picks a spawnable item by the weights of spec §6.
func (g *Game) randomItemKind() string {
	return weightedPick(g.Rng.Intn, g.Cfg.Items.Weights)
}

// weightedPick chooses a key from weights; keys are sorted so a seed gives a stable result.
func weightedPick(intn func(int) int, weights map[string]int) string {
	keys := make([]string, 0, len(weights))
	total := 0
	for k, w := range weights {
		keys = append(keys, k)
		total += w
	}
	sort.Strings(keys)
	if total <= 0 {
		return keys[0]
	}
	roll := intn(total)
	for _, k := range keys {
		roll -= weights[k]
		if roll < 0 {
			return k
		}
	}
	return keys[len(keys)-1]
}

// spawnItemNear drops an item at a random spot within spread of pos.
func (g *Game) spawnItemNear(kind string, value float64, pos Vec, spread float64) *Item {
	off := geom.FromAngle(g.Rng.Float64() * 2 * math.Pi).Mul(spread * (0.4 + 0.6*g.Rng.Float64()))
	at := g.resolveBlocking(pos.Add(off), g.Cfg.Items.ItemR)
	return g.addItem(kind, value, at)
}

func (g *Game) addItem(kind string, value float64, at Vec) *Item {
	it := &Item{ID: g.newID(), Kind: kind, Pos: at, Value: value}
	g.Items[it.ID] = it
	switch kind {
	case "orb":
		g.orbCount++
	case "chest":
	default:
		g.itemCount++
	}
	return it
}

func (g *Game) removeItem(it *Item) {
	delete(g.Items, it.ID)
	switch it.Kind {
	case "orb":
		g.orbCount--
	case "chest":
	default:
		g.itemCount--
	}
}

// PickupRadius is 60 * scale, × Nam Châm when active (spec §6).
func (g *Game) PickupRadius(p *Player) float64 {
	r := g.Cfg.Player.PickupR * p.Scale
	if p.HasBuff(BuffMagnet, g.Time) {
		r *= g.Cfg.Buffs["magnet"].Mult
	}
	return r
}

func (g *Game) resolvePickups() {
	for _, p := range g.PlayerList() {
		if p.Dead || p.InTransit(g.Time) {
			continue
		}
		r := g.PickupRadius(p) + g.Cfg.Items.ItemR
		r2 := r * r
		for _, it := range g.Items {
			if it.Pos.Dist2(p.Pos) <= r2 {
				g.pickup(p, it)
			}
		}
	}
}

func (g *Game) pickup(p *Player, it *Item) {
	g.removeItem(it)
	g.emit(Event{Kind: "ItemPicked", A: p.ID, B: it.ID, S: it.Kind, X: float32(it.Pos.X), Y: float32(it.Pos.Y)})
	switch it.Kind {
	case "orb":
		g.GainExp(p, it.Value, true)
	case "sword":
		g.AddSword(p)
	case "chest":
		for i := 0; i < g.Cfg.Items.ChestItems; i++ {
			g.spawnItemNear(g.randomItemKind(), 0, it.Pos, g.Cfg.Items.ScatterR*2)
		}
		g.emote(p, "chest")
	default:
		g.ApplyBuff(p, it.Kind)
	}
}

// ApplyBuff starts or refreshes a timed buff; picking it again never stacks (spec §6).
func (g *Game) ApplyBuff(p *Player, name string) {
	b, ok := BuffByName(name)
	if !ok {
		return
	}
	p.Buffs[b] = g.Time + g.Cfg.Buffs[name].Secs
	g.refreshDerived(p)
	g.emit(Event{Kind: "BuffStart", A: p.ID, S: name, N: int(g.Cfg.Buffs[name].Secs * 1000), X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}

// updateBuffs expires buffs and emits BuffEnd.
func (g *Game) updateBuffs(p *Player) {
	changed := false
	for b := 0; b < BuffCount; b++ {
		if p.Buffs[b] != 0 && p.Buffs[b] <= g.Time {
			p.Buffs[b] = 0
			changed = true
			g.emit(Event{Kind: "BuffEnd", A: p.ID, S: BuffNames[b], X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
		}
	}
	if changed {
		g.refreshDerived(p)
	}
}

func (g *Game) maintainSpawns() {
	ic := g.Cfg.Items
	for i := 0; i < spawnsPerTick && g.itemCount < ic.CountPerRoom; i++ {
		g.addItem(g.randomItemKind(), 0, g.randomFreePoint(ic.ItemR, ic.SpawnMinDist, 10))
	}
	for i := 0; i < spawnsPerTick && len(g.Monsters) < g.Cfg.Room.MonsterCount; i++ {
		g.spawnMonster()
	}
	g.updateChest()
}

func (g *Game) spawnMonster() {
	weights := map[string]int{}
	for name, k := range g.Cfg.Monsters.Kinds {
		if k.Max > 0 && name == "boss" && g.bossCount >= k.Max {
			continue
		}
		weights[name] = k.Weight
	}
	kind := weightedPick(g.Rng.Intn, weights)
	k := g.Cfg.Monsters.Kinds[kind]
	m := &Monster{ID: g.newID(), Kind: kind, R: k.Radius, HP: k.HP, MaxHP: k.HP, Exp: k.Exp,
		Pos: g.randomFreePoint(k.Radius, g.Cfg.Items.SpawnMinDist, 10), Heading: g.Rng.Float64() * 2 * math.Pi}
	if kind == "boss" {
		g.bossCount++
	}
	g.Monsters[m.ID] = m
}

// updateChest announces a Rương Báu 10s early, then spawns it (spec §6).
func (g *Game) updateChest() {
	ic := g.Cfg.Items
	if g.chestAt == nil && g.Time >= g.chestNextAt-ic.ChestWarnSecs {
		at := g.randomFreePoint(ic.ItemR*2, ic.SpawnMinDist, 20)
		g.chestAt = &at
		g.emit(Event{Kind: "ChestIncoming", Global: true, N: int(ic.ChestWarnSecs), X: float32(at.X), Y: float32(at.Y)})
	}
	if g.chestAt != nil && g.Time >= g.chestNextAt {
		g.addItem("chest", 0, *g.chestAt)
		g.chestAt = nil
		g.chestNextAt += ic.ChestEverySecs
	}
}

func (g *Game) updateTimers() {
	for _, p := range g.PlayerList() {
		if p.Dead {
			continue
		}
		g.updateBuffs(p)
		g.updateTransit(p)
	}
	g.updateObstacles()
}
