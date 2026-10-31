package room

import (
	"testing"

	"kiemxoay/server/internal/netproto"
)

// recConn giữ lại từng gói đã gửi để test đọc đúng cái client nhận được. Không dùng
// `snapshotFor` trực tiếp: gọi nó song song với broadcast thật sẽ TỰ ĂN MẤT gói xoá, vì
// `knownItems` bị xoá dấu ngay ở lần dựng snapshot đầu tiên.
type recConn struct{ sent [][]byte }

func (r *recConn) Send(b []byte) bool {
	r.sent = append(r.sent, append([]byte(nil), b...))
	return true
}

// Nhặt xong thì viên phải BIẾN MẤT trên màn hình người chơi, tức snapshot kế tiếp phải có
// id của nó trong ItemsRemoved. Sau khi đổi `knownItems` từ map băm sang slice theo chỉ số
// (tối ưu mục 26) thì đây là chỗ dễ sai nhất, nên khoá lại bằng test.
func TestItemDisappearsAfterPickup(t *testing.T) {
	m, _, cfg := testManager(t)
	// Tắt bot: phòng giờ lấp tới 60 nhân vật, và một con bot đi ngang có thể nhặt mất viên
	// trước người chơi — test thành đỏ ngẫu nhiên vì lý do chẳng liên quan gì tới luật xoá
	// vật phẩm mà nó đang kiểm.
	cfg.Room.BotFillTarget = 0
	conn := &recConn{}
	c := NewClient(conn, false)
	m.Join(c, netproto.Join{Name: "P0", Char: "a_hon"})
	r := m.AllRooms()[0]
	dt := 1 / float64(cfg.Net.TickHz)
	r.Step(dt)

	p := r.Zone(0).G.Players[c.PlayerID]
	p.InvulnUntil = 1e9
	// Thả viên trong tầm nhìn nhưng NGOÀI tầm nhặt, để server kịp gửi nó cho client; nếu
	// thả ngay dưới chân thì nhặt xong trong tick đầu, client chưa từng thấy nó.
	at := p.Pos
	at.X += 400
	it := r.Zone(0).G.SpawnItemForTest("orb", 10, at)
	itemID := it.ID
	for tick := 0; tick < 20; tick++ {
		r.Step(dt)
	}
	p.Pos = at // bước lên đúng chỗ viên nằm

	for tick := 0; tick < 120; tick++ {
		r.Step(dt)
	}
	// Gói snapshot nay đi NHỊ PHÂN (Protocol v2), không còn msgpack — xem `snapwire.go`.
	var seenAdded, seenRemoved bool
	for _, s := range snapsOf(t, conn.sent) {
		for _, a := range s.ItemsAdded {
			if a.ID == itemID {
				seenAdded = true
			}
		}
		for _, d := range s.ItemsDel {
			if d.ID == itemID {
				seenRemoved = true
			}
		}
	}
	if !seenAdded {
		t.Fatal("server chưa bao giờ gửi viên này cho client")
	}
	if _, still := r.Zone(0).G.Items[itemID]; still {
		t.Fatal("viên vẫn còn trong game — người chơi chưa nhặt được, test sai kịch bản")
	}
	if !seenRemoved {
		t.Fatal("nhặt xong mà server KHÔNG báo client xoá viên — viên nằm lại trên màn hình")
	}
}

// Slot của vật phẩm được DÙNG LẠI. Khi một viên bị nhặt và một viên mới chiếm đúng slot đó
// ngay trong cùng một tick, server vẫn phải báo client xoá viên CŨ. Thiếu chuyện này thì
// người chơi nhặt xong mà viên vẫn nằm trên màn hình mãi mãi.
func TestRemovalReportedEvenWhenSlotReused(t *testing.T) {
	m, _, cfg := testManager(t)
	// Tắt bot: phòng giờ lấp tới 60 nhân vật, và một con bot đi ngang có thể nhặt mất viên
	// trước người chơi — test thành đỏ ngẫu nhiên vì lý do chẳng liên quan gì tới luật xoá
	// vật phẩm mà nó đang kiểm.
	cfg.Room.BotFillTarget = 0
	conn := &recConn{}
	c := NewClient(conn, false)
	m.Join(c, netproto.Join{Name: "P0", Char: "a_hon"})
	r := m.AllRooms()[0]
	dt := 1 / float64(cfg.Net.TickHz)
	r.Step(dt)

	p := r.Zone(0).G.Players[c.PlayerID]
	p.InvulnUntil = 1e9
	at := p.Pos
	at.X += 400
	oldID := r.Zone(0).G.SpawnItemForTest("orb", 10, at).ID
	for tick := 0; tick < 20; tick++ {
		r.Step(dt) // để server kịp gửi viên cũ cho client
	}

	// Nhặt viên cũ rồi thả ngay một viên mới: viên mới sẽ nhận đúng slot vừa được nhả.
	oldSlot := r.Zone(0).G.Items[oldID].Slot()
	p.Pos = at
	r.Step(dt)
	if _, still := r.Zone(0).G.Items[oldID]; still {
		t.Fatal("chưa nhặt được viên cũ, test sai kịch bản")
	}
	// Thả viên mới cho tới khi có một viên nhận đúng slot vừa nhả ra. Thả ở CHỖ KHÁC để nó
	// không bị nhặt mất, nhưng vẫn trong tầm nhìn.
	far := p.Pos
	far.X += 700
	reused := false
	for i := 0; i < 200 && !reused; i++ {
		reused = r.Zone(0).G.SpawnItemForTest("orb", 10, far).Slot() == oldSlot
	}
	if !reused {
		t.Fatalf("không có viên nào nhận lại slot %d — không tái hiện được kịch bản", oldSlot)
	}
	for tick := 0; tick < 10; tick++ {
		r.Step(dt)
	}

	var removed bool
	for _, s := range snapsOf(t, conn.sent) {
		for _, d := range s.ItemsDel {
			if d.ID == oldID {
				removed = true
			}
		}
	}
	if !removed {
		t.Fatal("slot bị viên mới chiếm nên server quên báo xoá viên cũ — viên cũ kẹt lại trên màn hình")
	}
}
