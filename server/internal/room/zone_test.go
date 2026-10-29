package room

import (
	"fmt"
	"testing"

	"kiemxoay/server/internal/config"
	"kiemxoay/server/internal/game"
	"kiemxoay/server/internal/netproto"
	"kiemxoay/server/internal/world"
)

// zoneManager dựng một phòng 5 khu (spec §8.0). Khác `testManager` ở chỗ đó: các test luật
// chiến đấu chạy 1 khu cho đơn giản, còn ở đây chính thế giới nhiều khu mới là thứ cần thử.
func zoneManager(t *testing.T) (*Manager, *config.Config) {
	t.Helper()
	cfg := config.MustLoadDefault()
	cfg.Room.MonsterCount = 0
	cfg.Items.CountPerRoom = 0
	cfg.Room.BotFillTarget = 0 // bot tự sinh sẽ làm nhiễu phép đếm người mỗi khu
	// Bản đồ rỗng mặc định KHÔNG có cổng, mà cổng chính là thứ cần thử ở đây — nên tự đặt
	// 4 cổng mỗi khu đúng như spec §8.4.
	var maps []*world.Map
	for i := 0; i < cfg.Room.Zones; i++ {
		mp := world.EmptyMap(cfg.Room.MapSize)
		mp.ID = fmt.Sprintf("zone%d", i)
		half := cfg.Room.MapSize / 2
		for g := 0; g < 4; g++ {
			mp.Gates = append(mp.Gates, world.Gate{
				ID:      i*10 + g,
				Network: "world",
				X:       half + float64(1000*(1-2*(g%2))),
				Y:       half + float64(1000*(1-2*(g/2))),
			})
		}
		maps = append(maps, mp)
	}
	m := NewManager(cfg, maps, nil)
	m.Start = func(*Room) {}
	return m, cfg
}

// settle chạy vài tick để phòng nuốt hết hàng chờ — `Manager.Join` chỉ xếp việc vào inbox,
// nhân vật chưa thật sự có mặt cho tới khi phòng tick.
func settle(r *Room, cfg *config.Config) {
	dt := 1 / float64(cfg.Net.TickHz)
	for i := 0; i < 3; i++ {
		r.Step(dt)
	}
}

func TestRoomHasFiveZones(t *testing.T) {
	m, cfg := zoneManager(t)
	joinN(m, 1)
	r := m.AllRooms()[0]
	settle(r, cfg)
	if got := len(r.Zones()); got != cfg.Room.Zones {
		t.Fatalf("phòng có %d khu, muốn %d", got, cfg.Room.Zones)
	}
	for i, z := range r.Zones() {
		if z.Name == "" {
			t.Fatalf("khu %d chưa có tên", i)
		}
		if z.G.ZoneIndex != i {
			t.Fatalf("khu %d mang ZoneIndex %d", i, z.G.ZoneIndex)
		}
	}
}

// Người mới phải vào khu ÍT NGƯỜI THẬT NHẤT (spec §8.0) — đó là cả lý do có 5 khu: để
// người mới không vừa vào đã bị dí.
func TestNewPlayersSpreadAcrossZones(t *testing.T) {
	m, cfg := zoneManager(t)
	joinN(m, cfg.Room.Zones)
	r := m.AllRooms()[0]
	settle(r, cfg)
	for _, z := range r.Zones() {
		if z.Humans() != 1 {
			counts := make([]int, 0, len(r.Zones()))
			for _, q := range r.Zones() {
				counts = append(counts, q.Humans())
			}
			t.Fatalf("%d người vào %d khu mà chia thành %v, muốn mỗi khu 1", cfg.Room.Zones, cfg.Room.Zones, counts)
		}
	}
}

// Bảng xếp hạng tính CHUNG CẢ THẾ GIỚI (spec §8.0), không phải theo từng khu.
func TestRankingSpansZones(t *testing.T) {
	m, cfg := zoneManager(t)
	cs := joinN(m, 3)
	r := m.AllRooms()[0]
	settle(r, cfg)
	// Cho mỗi người một mức Uy Danh khác nhau, họ đang ở ba khu khác nhau.
	want := []float64{300, 200, 100}
	for i, c := range cs {
		z := r.ZoneOf(c.PlayerID)
		if z == nil {
			t.Fatalf("người %d không ở khu nào", i)
		}
		z.G.Players[c.PlayerID].Score = want[i]
	}
	rank := r.Ranking()
	if len(rank) < 3 {
		t.Fatalf("xếp hạng chỉ có %d người, muốn ít nhất 3", len(rank))
	}
	for i := 0; i+1 < 3; i++ {
		if rank[i].Score < rank[i+1].Score {
			t.Fatalf("xếp hạng không giảm dần: %v rồi %v", rank[i].Score, rank[i+1].Score)
		}
	}
	if rank[0].ID != cs[0].PlayerID {
		t.Fatalf("người Uy Danh cao nhất không đứng đầu bảng chung")
	}
}

// Cổng dịch chuyển nối CẢ THẾ GIỚI (spec §8.4): bước vào cổng có thể ra ở khu khác, và mọi
// trạng thái — kiếm, Uy Danh, Tiên khí — phải giữ nguyên qua chuyến đi.
func TestGateCarriesPlayerToAnotherZone(t *testing.T) {
	m, cfg := zoneManager(t)
	c := joinN(m, 1)[0]
	r := m.AllRooms()[0]
	settle(r, cfg)
	from := r.ZoneOf(c.PlayerID)
	if from == nil {
		t.Fatal("người chơi không ở khu nào")
	}
	if len(from.G.Map.Gates) == 0 {
		t.Skip("bản đồ không có cổng")
	}
	p := from.G.Players[c.PlayerID]
	p.Score, p.Ring.Count = 123, 7
	from.G.Respawn(p, 25, 7)
	p.Score = 123
	p.InvulnUntil = 1e9

	// Đứng vào cổng rồi chạy tới khi hết thời gian dịch chuyển.
	gate := from.G.Map.Gates[0]
	p.Pos = game.Vec{X: gate.X, Y: gate.Y}
	p.GateReadyAt = 0
	dt := 1 / float64(cfg.Net.TickHz)
	var to *Zone
	for i := 0; i < int(float64(cfg.Net.TickHz)*(cfg.Gates.TransitSecs+1)); i++ {
		r.Step(dt)
		if z := r.ZoneOf(c.PlayerID); z != nil && z != from {
			to = z
			break
		}
	}
	if to == nil {
		t.Skip("lần bốc ngẫu nhiên này ra cổng cùng khu — xem TestGateEventuallyReachesEveryZone")
	}
	moved := to.G.Players[c.PlayerID]
	if moved == nil {
		t.Fatal("sang khu mới mà không thấy nhân vật đâu")
	}
	// Uy Danh chỉ được TĂNG (điểm sống sót cộng dồn trong lúc đo), tuyệt đối không được
	// mất khi sang khu — chuyển khu phải giữ nguyên trạng thái (spec §8.0).
	if moved.Score < 123 || moved.Ring.Count != 7 {
		t.Fatalf("qua cổng xong mất trạng thái: Uy Danh %.0f (phải ≥ 123), kiếm %d (muốn 7)", moved.Score, moved.Ring.Count)
	}
	if _, still := from.G.Players[c.PlayerID]; still {
		t.Fatal("nhân vật còn sót lại ở khu cũ — sẽ bị đếm hai lần")
	}
	if _, ok := to.clients[c.PlayerID]; !ok {
		t.Fatal("kết nối chưa được chuyển sang khu mới, client sẽ không nhận snapshot nào nữa")
	}
}

// Với 20 cổng chung một mạng, đích phải rải ra CẢ NĂM KHU chứ không quanh quẩn một chỗ.
func TestGateEventuallyReachesEveryZone(t *testing.T) {
	m, cfg := zoneManager(t)
	joinN(m, 1)
	r := m.AllRooms()[0]
	settle(r, cfg)
	if len(r.Zone(0).G.Map.Gates) == 0 {
		t.Skip("bản đồ không có cổng")
	}
	seen := map[int]int{}
	gate := r.Zone(0).G.Map.Gates[0]
	for i := 0; i < 2000; i++ {
		d, ok := r.routeGate(0, gate)
		if !ok {
			t.Fatal("không chọn được cổng ra")
		}
		if d.Zone == 0 && d.GateID == gate.ID {
			t.Fatal("ra lại đúng cổng vừa bước vào — người chơi kẹt tại chỗ")
		}
		seen[d.Zone]++
	}
	if len(seen) != len(r.Zones()) {
		t.Fatalf("cổng chỉ dẫn tới %d/%d khu: %v", len(seen), len(r.Zones()), seen)
	}
	// ~80% số lần phải sang khu khác (spec §8.4 với 20 cổng chia đều 5 khu).
	if other := 2000 - seen[0]; other < 1400 {
		t.Fatalf("chỉ %d/2000 lần sang khu khác, spec nói khoảng 80%%", other)
	}
}

// Welcome phải mang tên khu, nếu không HUD không biết hiện chữ gì (spec §8.0).
func TestWelcomeCarriesZone(t *testing.T) {
	m, cfg := zoneManager(t)
	conn := &recConn{}
	c := NewClient(conn, false)
	m.Join(c, netproto.Join{Name: "P0", Char: "a_hon"})
	r := m.AllRooms()[0]
	settle(r, cfg)
	var w netproto.Welcome
	found := false
	for _, raw := range conn.sent {
		if netproto.Decode(raw, &w) == nil && w.T == "welcome" {
			found = true
		}
	}
	if !found {
		t.Fatal("không nhận được gói welcome")
	}
	if w.ZoneName == "" {
		t.Fatal("welcome thiếu tên khu")
	}
	if z := r.ZoneOf(c.PlayerID); z == nil || w.Zone != z.Index {
		t.Fatalf("welcome báo khu %d, người chơi lại ở khu khác", w.Zone)
	}
}
