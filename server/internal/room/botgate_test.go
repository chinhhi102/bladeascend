package room

import (
	"testing"

	"kiemxoay/server/internal/game"
)

// Bot phải biết dùng cổng (spec v5 mục 8). Không có luật này thì bot nào rơi vào khu vắng
// sẽ quanh quẩn ở đó tới hết trận và cả một khu của thế giới thành vùng chết.
func TestBotsEventuallyChangeZone(t *testing.T) {
	m, cfg := zoneManager(t)
	cfg.Room.BotFillTarget = 10
	joinN(m, 1)
	r := m.AllRooms()[0]
	settle(r, cfg)

	dt := 1 / float64(cfg.Net.TickHz)
	// Ghi lại bot nào đang ở khu nào, rồi chạy một lúc xem có con nào đổi khu không.
	where := map[uint32]int{}
	for _, z := range r.Zones() {
		for id := range z.brains {
			where[id] = z.Index
		}
	}
	if len(where) == 0 {
		// Bot sinh dần theo bots.spawn_every_secs; chạy thêm cho tới khi có.
		for i := 0; i < int(float64(cfg.Net.TickHz)*cfg.Bots.SpawnEverySecs*4) && len(where) == 0; i++ {
			r.Step(dt)
			for _, z := range r.Zones() {
				for id := range z.brains {
					where[id] = z.Index
				}
			}
		}
	}
	if len(where) == 0 {
		t.Skip("phòng chưa sinh bot nào")
	}

	moved := 0
	for i := 0; i < int(float64(cfg.Net.TickHz)*180); i++ {
		r.Step(dt)
		for _, z := range r.Zones() {
			for id := range z.brains {
				if old, ok := where[id]; ok && old != z.Index {
					moved++
					where[id] = z.Index
				}
			}
		}
		if moved > 0 {
			break
		}
	}
	if moved == 0 {
		t.Fatal("chạy 3 phút mà không bot nào qua cổng sang khu khác")
	}
}

// Trần người + bot vẫn tính cho CẢ PHÒNG (spec v5 mục 8) — thế giới 5 khu là để bớt bị dí,
// không phải để nhân năm số nhân vật.
func TestBotFillCountsWholeRoomNotPerZone(t *testing.T) {
	m, cfg := zoneManager(t)
	cfg.Room.BotFillTarget = 20
	joinN(m, 2)
	r := m.AllRooms()[0]
	dt := 1 / float64(cfg.Net.TickHz)
	for i := 0; i < int(float64(cfg.Net.TickHz)*cfg.Bots.SpawnEverySecs*30); i++ {
		r.Step(dt)
	}
	if total := r.Humans() + r.Bots(); total > cfg.Room.BotFillTarget {
		t.Fatalf("phòng có %d nhân vật, trần là %d — đang đếm theo khu chứ không theo phòng",
			total, cfg.Room.BotFillTarget)
	}
	var _ = game.Vec{}
}
