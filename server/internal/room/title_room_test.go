package room

import (
	"fmt"
	"testing"

	"bladeascend/server/internal/config"
	"bladeascend/server/internal/netproto"
	"bladeascend/server/internal/world"
)

// fiveZoneRoom dựng một phòng 5 KHU THẬT.
//
// `testManager` truyền `maps = nil` nên phòng chỉ có một khu — và một bài kiểm "danh hiệu
// theo phòng hay theo khu" chạy trên phòng một khu thì không kiểm được gì. Bản đầu của tôi
// dùng nó rồi `t.Skip` khi hai người rơi cùng khu: một test bỏ qua là một test không tồn
// tại, chỉ khác ở chỗ nó trông như đã chạy.
func fiveZoneRoom(t *testing.T) (*Manager, *config.Config) {
	t.Helper()
	cfg := config.MustLoadDefault()
	cfg.Room.MonsterCount = 0
	cfg.Items.CountPerRoom = 0
	cfg.Room.ZoneMinChars = 0
	cfg.TitleHold = 0 // kiểm AI được danh hiệu, không kiểm chống nhấp nháy
	var maps []*world.Map
	for i := 0; i < cfg.Room.Zones; i++ {
		mp := world.EmptyMap(cfg.Room.MapSize)
		mp.ID = fmt.Sprintf("zone%d", i)
		maps = append(maps, mp)
	}
	m := NewManager(cfg, maps, nil)
	m.Start = func(*Room) {}
	return m, cfg
}

// Danh hiệu tính theo CẢ PHÒNG, không theo khu — GAME_SPEC v6.19 §11.3.
//
// Lỗi cũ không ném ra gì cả: `updateTitles` xếp hạng trong `g.Ranking()`, tức là trong KHU
// NÀY. Một phòng 5 khu vì thế phát ra **năm vương miện Top 1**, và hai người trong số đó
// gặp nhau qua cổng thì cả hai đều đeo. Chỉ nhìn mới thấy — nên phải có test.

// titleOf tìm danh hiệu của một người, dù họ đang ở khu nào.
func titleOf(r *Room, id uint32) int {
	for i := 0; i < r.ZoneCount(); i++ {
		if p, ok := r.Zone(i).G.Players[id]; ok {
			return p.Title
		}
	}
	return -1
}

// splitZones đảm bảo hai người ở HAI KHU KHÁC NHAU.
//
// Không dựa vào luật xếp khu: cả hai `Join` xảy ra TRƯỚC tick đầu tiên, nên mọi khu đều
// đang có 0 người và luật "khu ít người nhất" (§8.0) chọn cùng một khu cho cả hai. Đó là
// hành vi đúng — nên bài kiểm phải tự đặt người, chứ không phải đổi luật cho vừa bài kiểm.
func splitZones(t *testing.T, r *Room, ca, cb *Client) (*Zone, *Zone) {
	t.Helper()
	za := r.ZoneOf(ca.PlayerID)
	zb := r.ZoneOf(cb.PlayerID)
	if za == nil || zb == nil {
		t.Fatal("không tìm được khu của người chơi")
	}
	if za != zb {
		return za, zb
	}
	// Bê người thứ hai sang một khu khác, đúng đường mà cổng dịch chuyển đi.
	var other *Zone
	for i := 0; i < r.ZoneCount(); i++ {
		if r.Zone(i) != za {
			other = r.Zone(i)
			break
		}
	}
	if other == nil {
		t.Fatal("phòng chỉ có một khu — bài kiểm này cần 5 khu")
	}
	p := zb.G.Players[cb.PlayerID]
	r.queueTransfer(transfer{playerID: cb.PlayerID, from: zb, to: other, at: p.Pos, destGate: -1})
	r.Step(1.0 / 20)
	zb = r.ZoneOf(cb.PlayerID)
	if zb == nil || zb == za {
		t.Fatal("chuyển khu không thành")
	}
	return za, zb
}

func TestTop1IsRoomWideNotPerZone(t *testing.T) {
	m, cfg := fiveZoneRoom(t)
	ca, cb := NewClient(&fakeConn{}, false), NewClient(&fakeConn{}, false)
	m.Join(ca, netproto.Join{Name: "Manh", Char: "a_hon"})
	m.Join(cb, netproto.Join{Name: "Yeu", Char: "a_hon"})
	r := m.AllRooms()[0]
	dt := 1 / float64(cfg.Net.TickHz)
	r.Step(dt)

	// Ép hai người vào HAI KHU KHÁC NHAU — đó là cả điểm của bài kiểm.
	za, zb := splitZones(t, r, ca, cb)

	pa, pb := za.G.Players[ca.PlayerID], zb.G.Players[cb.PlayerID]
	if pa == nil || pb == nil {
		t.Fatalf("không thấy nhân vật sau khi tách khu (za=%s zb=%s)", za.Name, zb.Name)
	}
	pa.InvulnUntil, pb.InvulnUntil = 1e9, 1e9
	pa.Score = 5000 // người mạnh
	pb.Score = 100  // người yếu, nhưng là người DUY NHẤT trong khu của mình

	for i := 0; i < 10; i++ {
		r.Step(dt)
	}

	ta, tb := titleOf(r, ca.PlayerID), titleOf(r, cb.PlayerID)
	if ta != 1 {
		t.Errorf("người Uy Danh cao nhất phòng (5000) phải là Top 1, được danh hiệu %d", ta)
	}
	if tb == 1 {
		t.Errorf("người Uy Danh 100 đang đeo vương miện Top 1 — danh hiệu đang tính THEO KHU.\n" +
			"Một phòng 5 khu sẽ có 5 người cùng là Top 1, và khi họ gặp nhau qua cổng thì cả\n" +
			"hai cùng đeo vương miện.")
	}
	if tb != 2 {
		t.Errorf("người thứ hai của PHÒNG phải là Top 2, được %d", tb)
	}
}

// Người ở khu khác vẫn thấy ĐÚNG danh hiệu của Top 1 khi gặp (v6.19 §1).
//
// Danh hiệu nằm trên `game.Player`, và snapshot chở nó đi — nên khi Top 1 sang khu của
// người kia, giá trị đi theo. Kiểm để chắc việc cắm `RoomRank` không làm danh hiệu bị
// tính lại theo khu mới.
func TestTitleFollowsPlayerAcrossZones(t *testing.T) {
	m, cfg := fiveZoneRoom(t)
	ca, cb := NewClient(&fakeConn{}, false), NewClient(&fakeConn{}, false)
	m.Join(ca, netproto.Join{Name: "Manh", Char: "a_hon"})
	m.Join(cb, netproto.Join{Name: "Yeu", Char: "a_hon"})
	r := m.AllRooms()[0]
	dt := 1 / float64(cfg.Net.TickHz)
	r.Step(dt)

	za, zb := splitZones(t, r, ca, cb)
	pa, pb := za.G.Players[ca.PlayerID], zb.G.Players[cb.PlayerID]
	if pa == nil || pb == nil {
		t.Fatalf("không tìm thấy nhân vật sau khi tách khu (pa=%v pb=%v)", pa != nil, pb != nil)
	}
	pa.InvulnUntil, pb.InvulnUntil = 1e9, 1e9
	pa.Score, pb.Score = 5000, 100
	for i := 0; i < 10; i++ {
		r.Step(dt)
	}
	if got := titleOf(r, ca.PlayerID); got != 1 {
		t.Fatalf("chuẩn bị sai: người mạnh chưa là Top 1 (danh hiệu %d)", got)
	}

	// Chuyển người mạnh sang khu của người yếu rồi tick tiếp.
	// `queueTransfer` + một tick: `applyTransfers` chạy sau khi mọi khu tick xong, đúng
	// đường mà cổng dịch chuyển đi.
	r.queueTransfer(transfer{
		playerID: ca.PlayerID, from: za, to: zb, at: pb.Pos, destGate: -1,
	})
	for i := 0; i < 10; i++ {
		r.Step(dt)
	}
	if got := titleOf(r, ca.PlayerID); got != 1 {
		t.Errorf("Top 1 sang khu khác thì MẤT danh hiệu (còn %d) — hạng đang tính lại theo khu", got)
	}
	if got := titleOf(r, cb.PlayerID); got == 1 {
		t.Errorf("người yếu thành Top 1 sau khi Top 1 sang khu mình (danh hiệu %d)", got)
	}
}
