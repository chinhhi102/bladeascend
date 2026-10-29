package room

import (
	"kiemxoay/server/internal/bot"
	"kiemxoay/server/internal/game"
)

// updateBots implements the bot fill rules of spec §9.3.
//
// Hai con số, hai mục đích khác nhau:
//   - `room.bot_fill_target` là **trần của cả phòng** (người + bot). Thế giới 5 khu là để
//     bớt cảnh vừa vào đã bị dí, nhưng tổng số nhân vật vẫn phải có trần, nếu không mỗi
//     phòng thành năm phòng và tick server tăng gấp năm.
//   - `room.zone_min_chars` là **sàn của từng khu**. Không có nó thì bot dồn vào vài khu
//     và vẫn có khu trống trơn — người chơi qua cổng sang đó thấy một thế giới chết.
func (r *Room) updateBots(dt float64) {
	r.processRespawns()
	r.botTimer += dt
	if r.botTimer < r.cfg.Bots.SpawnEverySecs {
		return
	}
	r.botTimer = 0
	// Người thật được ưu tiên trong trần phòng: họ vào bất cứ lúc nào, nên nếu tổng vượt
	// trần thì phải bớt bot NGAY, không đợi bot chết. Chờ chết thì một phòng đông người có
	// thể vượt trần suốt mấy phút.
	r.trimOverCap()
	// Bot đi lại giữa các khu (qua cổng, hồi sinh) nên một khu có thể tụt xuống dưới sàn
	// trong khi phòng đã chạm trần — lúc đó không sinh thêm được, phải ĐIỀU bot từ khu đông
	// sang. Không làm thế thì sàn mỗi khu chỉ đúng lúc mới mở phòng.
	r.rebalanceZones()
	room := r.Humans() + r.Bots()
	target := r.cfg.Room.BotFillTarget
	// Chỉ `bot_fill_target` mới chặn việc sinh bot. `max_humans` là trần NGƯỜI THẬT, hai
	// con số khác nhau: 20 người + 40 bot vẫn đúng ý "tối đa 20 người, tổng 60 nhân vật".
	// Bản cũ dừng sinh bot ngay khi đủ 20 người, nên phòng đông người lại thành phòng vắng.
	if room >= target {
		return
	}
	// Ưu tiên khu nào chưa đạt sàn; hết việc đó thì rải đều cho tới khi chạm trần phòng.
	budget := min(target-room, r.cfg.Bots.SpawnBatch)
	for i := 0; i < budget; i++ {
		z := r.zoneBelowFloor()
		if z == nil {
			z = r.thinnestByCount()
		}
		level, swords := r.randomBotLevel()
		r.spawnBotIn(z, level, swords)
	}
}

// trimOverCap bớt bot cho tới khi tổng nhân vật về đúng trần phòng. Bỏ bot ở khu ĐÔNG
// nhất trước, để việc cắt giảm không tạo ra khu trống.
func (r *Room) trimOverCap() {
	for r.Humans()+r.Bots() > r.cfg.Room.BotFillTarget {
		var fat *Zone
		for _, z := range r.zones {
			if len(z.brains) > 0 && (fat == nil || z.Count() > fat.Count()) {
				fat = z
			}
		}
		if fat == nil {
			return // không còn bot nào để bớt
		}
		for id := range fat.brains {
			fat.G.RemovePlayer(id)
			delete(fat.brains, id)
			break
		}
	}
}

// rebalanceZones điều một bot từ khu đông nhất sang khu đang dưới sàn.
func (r *Room) rebalanceZones() {
	for i := 0; i < len(r.zones); i++ {
		thin := r.zoneBelowFloor()
		if thin == nil {
			return
		}
		var fat *Zone
		for _, z := range r.zones {
			if z == thin || len(z.brains) == 0 {
				continue
			}
			if fat == nil || z.Count() > fat.Count() {
				fat = z
			}
		}
		// Chỉ điều khi khu kia thật sự dư: điều qua điều lại giữa hai khu cùng thiếu thì
		// chỉ làm bot nhấp nháy giữa hai bản đồ.
		if fat == nil || fat.Count() <= thin.Count()+1 {
			return
		}
		for id := range fat.brains {
			r.queueTransfer(transfer{playerID: id, from: fat, to: thin,
				at: game.Vec{X: thin.G.Map.Size / 2, Y: thin.G.Map.Size / 2}})
			break
		}
	}
}

// zoneBelowFloor trả về khu thiếu người nhất trong số các khu chưa đạt sàn, nil nếu đủ cả.
func (r *Room) zoneBelowFloor() *Zone {
	var worst *Zone
	for _, z := range r.zones {
		if z.Count() >= r.cfg.Room.ZoneMinChars {
			continue
		}
		if worst == nil || z.Count() < worst.Count() {
			worst = z
		}
	}
	return worst
}

func (r *Room) spawnBot() *game.Player {
	level, swords := r.randomBotLevel()
	return r.spawnBotAt(level, swords)
}

// spawnBotAt thả bot vào khu ÍT NHÂN VẬT NHẤT (spec v5 mục 8) để năm khu không bị lệch —
// một khu chật ních trong khi khu khác trống trơn thì thế giới rộng cũng vô nghĩa.
func (r *Room) spawnBotAt(level, swords int) *game.Player {
	return r.spawnBotIn(r.thinnestByCount(), level, swords)
}

func (r *Room) spawnBotIn(z *Zone, level, swords int) *game.Player {
	if z == nil {
		return nil
	}
	names := r.cfg.Bots.Names
	p := z.G.AddPlayer(game.SpawnOptions{
		Name:  names[r.rng.Intn(len(names))],
		Char:  r.cfg.Characters[r.rng.Intn(len(r.cfg.Characters))],
		IsBot: true, Level: level, Swords: swords,
	})
	z.brains[p.ID] = bot.New(p.ID, r.rng.Int63(), z.G)
	return p
}

// thinnestByCount: khu ít nhân vật nhất (người + bot), hoà thì bốc ngẫu nhiên.
func (r *Room) thinnestByCount() *Zone {
	best := []*Zone{}
	low := 1 << 30
	for _, z := range r.zones {
		switch n := z.Count(); {
		case n < low:
			low, best = n, []*Zone{z}
		case n == low:
			best = append(best, z)
		}
	}
	if len(best) == 0 {
		return nil
	}
	return best[r.rng.Intn(len(best))]
}

// randomBotLevel decides what a new bot spawns as (spec §9.3).
//
// Bots slot in at `spawn_rank` (11) — level with whoever sits just below the room's Top 10 —
// so they never show up strong enough to bully the leaders. A room without that many players
// yet has nobody to measure against, so its bots start at Luyện Khí instead. There is no
// realm ceiling: in a room full of Nguyên Anh, rank 11 is Nguyên Anh and so are new bots.
//
// Xếp hạng lấy của CẢ PHÒNG, không phải của một khu: bot vào khu vắng vẫn phải vừa sức với
// mặt bằng chung, nếu không thì khu vắng thành chỗ nuôi quân toàn bot yếu.
func (r *Room) randomBotLevel() (int, int) {
	prog := r.zones[0].G.Prog
	level := r.rng.Intn(prog.Tiers()) // mặc định: Luyện Khí, tầng ngẫu nhiên
	if n := r.cfg.Bots.SpawnRank; n >= 1 {
		if ranking := r.Ranking(); len(ranking) >= n {
			level = ranking[n-1].Level
		}
	}
	realm := prog.Realm(level)
	return level, 1 + r.rng.Intn(prog.MaxSwords(realm))
}

func (r *Room) onBotDeath(z *Zone, p *game.Player) {
	if r.Humans()+r.Bots()-1 >= r.cfg.Room.BotFillTarget {
		z.G.RemovePlayer(p.ID)
		delete(z.brains, p.ID)
		return
	}
	z.respawns = append(z.respawns, pendingRespawn{id: p.ID, at: z.G.Time + r.cfg.Bots.RespawnSecs})
}

// respawnQueued chỉ hỏi hàng chờ CỦA KHU đó — gọi từ goroutine của chính khu ấy.
func (z *Zone) respawnQueued(id uint32) bool {
	for _, q := range z.respawns {
		if q.id == id {
			return true
		}
	}
	return false
}

// processRespawns hồi sinh bot tới hạn. Bot hồi sinh **chuyển sang khu ít nhân vật nhất**
// (spec v5 mục 8), không ở lại chỗ cũ — nếu không thì khu đông sẽ tự nuôi lại chính nó.
func (r *Room) processRespawns() {
	for _, z := range r.zones {
		keep := z.respawns[:0]
		for _, q := range z.respawns {
			if z.G.Time < q.at {
				keep = append(keep, q)
				continue
			}
			p, ok := z.G.Players[q.id]
			if !ok {
				continue
			}
			level, swords := r.randomBotLevel()
			if to := r.thinnestByCount(); to != nil && to != z {
				z.G.RemovePlayer(p.ID)
				delete(z.brains, p.ID)
				r.spawnBotIn(to, level, swords)
				continue
			}
			z.G.Respawn(p, level, swords)
		}
		z.respawns = keep
	}
}
