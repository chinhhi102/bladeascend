package room

import (
	"fmt"
	"testing"

	"kiemxoay/server/internal/config"
	"kiemxoay/server/internal/world"
)

// Id entity có trùng nhau giữa các khu của CÙNG một phòng không?
func TestEntityIDsUniqueAcrossZones(t *testing.T) {
	cfg := config.MustLoadDefault()
	cfg.Room.BotFillTarget = 0
	var maps []*world.Map
	for i := 0; i < cfg.Room.Zones; i++ {
		mp := world.EmptyMap(cfg.Room.MapSize)
		mp.ID = fmt.Sprintf("z%d", i)
		maps = append(maps, mp)
	}
	m := NewManager(cfg, maps, nil)
	m.Start = func(*Room) {}
	c := NewClient(&fakeConn{}, false)
	r := joinWith(m, c, "P")
	r.Step(1.0 / 20)

	seen := map[uint32]string{}
	dup := 0
	for i := 0; i < r.ZoneCount(); i++ {
		z := r.Zone(i)
		for id := range z.G.Players {
			if where, ok := seen[id]; ok {
				dup++
				if dup <= 3 {
					t.Logf("  id %d có ở CẢ %s lẫn %s", id, where, z.Name)
				}
			}
			seen[id] = z.Name
		}
		for id := range z.G.Monsters {
			if where, ok := seen[id]; ok {
				dup++
				if dup <= 3 {
					t.Logf("  id %d (quái) có ở CẢ %s lẫn %s", id, where, z.Name)
				}
			}
			seen[id] = z.Name
		}
	}
	t.Logf("tổng %d id, %d lần trùng", len(seen), dup)
	if dup > 0 {
		t.Errorf("%d id trùng nhau giữa các khu trong CÙNG một phòng", dup)
	}
}
