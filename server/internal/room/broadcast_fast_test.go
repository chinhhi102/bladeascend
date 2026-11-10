package room

import (
	"testing"

	"bladeascend/server/internal/game"
)

// Q48 bước 1: người chạy nhanh hơn 2× tốc gốc LUÔN gửi đủ nhịp, kệ khoảng cách.
//
// Đo 10/11 trước khi có luật này: Shipper Gom Đơn (975 đv/s) băng qua ranh giới "nửa khung
// nhìn" mấy lần mỗi giây nên nhịp gửi nhảy 20↔10 Hz — khoảng cách gói lớn nhất 167 ms, kho
// nội suy 150 ms cạn, người xem thấy đứng hình rồi nhảy cóc.
func TestFastMoverAlwaysFullRate(t *testing.T) {
	m, _, _ := testManager(t)
	joinN(m, 1)
	r := m.AllRooms()[0]
	r.Step(1.0 / 30)
	g := r.Zone(0).G

	ship := g.AddPlayer(game.SpawnOptions{Name: "Shipper", Char: "shipper", Level: 10, Swords: 1})
	if fastMover(g, ship) {
		t.Fatal("đi bộ mà đã tính là chạy nhanh")
	}
	base := g.Cfg.Player.MoveBase
	g.ApplyStatus(ship, game.StGhost, 4) // Gom Đơn
	if got := g.MoveSpeed(ship); got <= base*2 {
		t.Fatalf("Gom Đơn cho tốc %.0f, phải hơn 2× base %.0f — bài kiểm đặt sai", got, base*2)
	}
	if !fastMover(g, ship) {
		t.Fatalf("Gom Đơn %.0f đv/s mà không tính là chạy nhanh", g.MoveSpeed(ship))
	}
}
