package room

import (
	"kiemxoay/server/internal/bot"
	"kiemxoay/server/internal/game"
)

// updateBots implements the bot fill rules of spec §9.3.
func (r *Room) updateBots(dt float64) {
	r.processRespawns()
	r.botTimer += dt
	if r.botTimer < r.cfg.Bots.SpawnEverySecs {
		return
	}
	r.botTimer = 0
	humans, bots := len(r.clients), len(r.brains)
	target := r.cfg.Room.BotFillTarget
	if humans >= r.cfg.Room.MaxHumans || humans+bots >= target {
		return
	}
	for i := 0; i < min(target-humans-bots, r.cfg.Bots.SpawnBatch); i++ {
		r.spawnBot()
	}
}

func (r *Room) spawnBot() {
	level, swords := r.randomBotLevel()
	names := r.cfg.Bots.Names
	p := r.G.AddPlayer(game.SpawnOptions{
		Name:  names[r.rng.Intn(len(names))],
		Char:  r.cfg.Characters[r.rng.Intn(len(r.cfg.Characters))],
		IsBot: true, Level: level, Swords: swords,
	})
	r.brains[p.ID] = bot.New(p.ID, r.rng.Int63(), r.G)
}

// randomBotLevel picks a realm by realm_weights, a random tier and [1, max] swords.
func (r *Room) randomBotLevel() (int, int) {
	total := 0
	for _, w := range r.cfg.Bots.RealmWeights {
		total += w
	}
	roll, realm := r.rng.Intn(max(1, total)), 0
	for i, w := range r.cfg.Bots.RealmWeights {
		if roll < w {
			realm = i
			break
		}
		roll -= w
	}
	level := realm*r.G.Prog.Tiers() + r.rng.Intn(r.G.Prog.Tiers())
	return level, 1 + r.rng.Intn(r.G.Prog.MaxSwords(realm))
}

func (r *Room) onBotDeath(p *game.Player) {
	if len(r.clients)+len(r.brains)-1 >= r.cfg.Room.BotFillTarget {
		r.G.RemovePlayer(p.ID)
		delete(r.brains, p.ID)
		return
	}
	r.respawns = append(r.respawns, pendingRespawn{id: p.ID, at: r.G.Time + r.cfg.Bots.RespawnSecs})
}

func (r *Room) respawnQueued(id uint32) bool {
	for _, q := range r.respawns {
		if q.id == id {
			return true
		}
	}
	return false
}

func (r *Room) processRespawns() {
	keep := r.respawns[:0]
	for _, q := range r.respawns {
		if r.G.Time < q.at {
			keep = append(keep, q)
			continue
		}
		if p, ok := r.G.Players[q.id]; ok {
			level, swords := r.randomBotLevel()
			r.G.Respawn(p, level, swords)
		}
	}
	r.respawns = keep
}
