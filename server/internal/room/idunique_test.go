package room

import (
	"fmt"
	"testing"

	"bladeascend/server/internal/config"
	"bladeascend/server/internal/world"
)

// Id entity phải duy nhất trong CẢ PHÒNG, và giữ nguyên khi đi qua cổng.
//
// Lỗi gốc: `nextID` bắt đầu từ 0 ở MỌI khu, mà năm khu sinh cùng số quái và cùng số vật
// phẩm theo config — nên khu 0 và khu 1 cấp ra đúng những con số giống nhau. Đo trên phòng
// 5 khu, tick đầu tiên: 79 id trùng.
//
// Vì sao nó tệ hơn vẻ ngoài: `Room.rankAt` là bảng CẢ PHÒNG khoá theo id, nên hai người
// khác khu cùng id ghi đè hạng của nhau → danh hiệu phát sai. Bảng xếp hạng cũng gửi id, và
// client so `e.id === myId` để tô vàng dòng của mình → tô nhầm người.

func idRoom(t *testing.T) (*Manager, *config.Config) {
	t.Helper()
	cfg := config.MustLoadDefault()
	cfg.Room.ZoneMinChars = 0
	var maps []*world.Map
	for i := 0; i < cfg.Room.Zones; i++ {
		mp := world.EmptyMap(cfg.Room.MapSize)
		mp.ID = fmt.Sprintf("z%d", i)
		maps = append(maps, mp)
	}
	m := NewManager(cfg, maps, nil)
	m.Start = func(*Room) {}
	return m, cfg
}

// allIDs gom mọi id đang sống trong phòng, kèm nơi tìm thấy.
func allIDs(r *Room) (map[uint32]string, []string) {
	seen := map[uint32]string{}
	var dup []string
	add := func(id uint32, what, zone string) {
		if where, ok := seen[id]; ok {
			dup = append(dup, fmt.Sprintf("id %d: %s và %s/%s", id, where, zone, what))
			return
		}
		seen[id] = zone + "/" + what
	}
	for i := 0; i < r.ZoneCount(); i++ {
		z := r.Zone(i)
		for id := range z.G.Players {
			add(id, "người", z.Name)
		}
		for id := range z.G.Monsters {
			add(id, "quái", z.Name)
		}
		for id := range z.G.Items {
			add(id, "vật phẩm", z.Name)
		}
	}
	return seen, dup
}

func TestEntityIDsUniqueAcrossZonesAfter1000Transfers(t *testing.T) {
	m, cfg := idRoom(t)
	c := NewClient(&fakeConn{}, false)
	r := joinWith(m, c, "DiVongQuanh")
	dt := 1 / float64(cfg.Net.TickHz)
	r.Step(dt)

	wantID := c.PlayerID
	if wantID == 0 {
		t.Fatal("chưa có PlayerID sau tick đầu")
	}
	if _, dup := allIDs(r); len(dup) > 0 {
		t.Fatalf("đã trùng id ngay từ đầu: %v", dup[:min(3, len(dup))])
	}

	// 1 000 lượt qua cổng, vòng quanh 5 khu.
	for i := 0; i < 1000; i++ {
		from := r.ZoneOf(wantID)
		if from == nil {
			t.Fatalf("lượt %d: mất nhân vật khỏi phòng", i)
		}
		to := r.Zone((from.Index + 1) % r.ZoneCount())
		p := from.G.Players[wantID]
		if p == nil {
			t.Fatalf("lượt %d: không thấy nhân vật ở %s", i, from.Name)
		}
		r.queueTransfer(transfer{playerID: wantID, from: from, to: to, at: p.Pos, destGate: -1})
		r.Step(dt)

		// Id phải GIỮ NGUYÊN: client đang vẽ theo id này, đổi id là nhân vật biến mất rồi
		// hiện lại — đúng kiểu nhấp nháy mà Protocol v2 vừa dẹp xong.
		if c.PlayerID != wantID {
			t.Fatalf("lượt %d: id đổi từ %d thành %d khi qua cổng", i, wantID, c.PlayerID)
		}
		if got := r.ZoneOf(wantID); got == nil || got.Index != to.Index {
			t.Fatalf("lượt %d: chuyển sang %s không thành", i, to.Name)
		}
		// Kiểm trùng ở vài mốc: kiểm MỖI lượt thì 1 000 × ~7 000 entity = quá chậm, mà
		// trùng id là thứ xuất hiện rồi ở lại, không thoáng qua.
		if i%100 == 0 || i == 999 {
			if _, dup := allIDs(r); len(dup) > 0 {
				t.Fatalf("lượt %d: %d id trùng, ví dụ %v", i, len(dup), dup[:min(3, len(dup))])
			}
		}
	}

	seen, dup := allIDs(r)
	if len(dup) > 0 {
		t.Errorf("sau 1 000 lượt qua cổng: %d id trùng", len(dup))
	}
	t.Logf("1 000 lượt qua cổng · %d entity · 0 id trùng · id người chơi giữ nguyên %d", len(seen), wantID)
}

// Mỗi khu phải có khoảng id RIÊNG, không chồng lên nhau.
//
// Kiểm trực tiếp tính chất, không chỉ kiểm "hiện tại chưa trùng": một phòng vắng có thể
// chưa trùng chỉ vì chưa sinh đủ entity, và bài kiểm sẽ xanh cho tới ngày nó không nên xanh.
func TestZoneIDRangesDoNotOverlap(t *testing.T) {
	m, cfg := idRoom(t)
	c := NewClient(&fakeConn{}, false)
	r := joinWith(m, c, "P")
	r.Step(1 / float64(cfg.Net.TickHz))

	for i := 0; i < r.ZoneCount(); i++ {
		z := r.Zone(i)
		want := uint32(i) << 24
		for id := range z.G.Monsters {
			if id&0xFF000000 != want {
				t.Fatalf("khu %d (%s) cấp id %d (tiền tố %d), phải là tiền tố %d",
					i, z.Name, id, id>>24, i)
			}
		}
	}
	// Trần: khu đông nhất ~7 000 entity, khoảng mỗi khu 16 777 215.
	const perZone = 1 << 24
	most := 0
	for i := 0; i < r.ZoneCount(); i++ {
		if n := len(r.Zone(i).G.Monsters) + len(r.Zone(i).G.Items) + len(r.Zone(i).G.Players); n > most {
			most = n
		}
	}
	if most*100 > perZone {
		t.Errorf("khu đông nhất %d entity — quá gần trần %d của một khoảng id", most, perZone)
	}
	t.Logf("5 khoảng id riêng biệt · khu đông nhất %d entity / %d mỗi khoảng", most, perZone)
}
