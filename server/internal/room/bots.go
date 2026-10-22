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

// randomBotLevel decides what a new bot spawns as (spec §9.3).
//
// Bots slot in at `spawn_rank` (11) — level with whoever sits just below the room's Top 10 —
// so they never show up strong enough to bully the leaders. A room without that many players
// yet has nobody to measure against, so its bots start at Luyện Khí instead. There is no
// realm ceiling: in a room full of Nguyên Anh, rank 11 is Nguyên Anh and so are new bots.
func (r *Room) randomBotLevel() (int, int) {
	level := r.rng.Intn(r.G.Prog.Tiers()) // mặc định: Luyện Khí, tầng ngẫu nhiên
	if n := r.cfg.Bots.SpawnRank; n >= 1 {
		if ranking := r.G.Ranking(); len(ranking) >= n {
			level = ranking[n-1].Level
		}
	}
	realm := r.G.Prog.Realm(level)
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
