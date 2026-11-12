package game

import (
	"testing"

	"bladeascend/server/internal/config"
	"bladeascend/server/internal/world"
)

// Vệ sinh bộ nhớ §13.9 — lưới vật phẩm không được giữ con trỏ của viên đã xoá.
//
// Đây là bài gác cho một rò rỉ THẬT tìm ra ngày 12/11: `ig.cells[i] = c[:last]` hạ `len`
// nhưng mảng nền vẫn giữ con trỏ ở ô cuối, nên mọi viên EXP đã nhặt đều không bao giờ được
// thu hồi — heap server tăng đều 7 MB/phút.
//
// Bài này KHÔNG đo heap (số đo heap phụ thuộc GC và tải máy, xem luật 16). Nó đếm thẳng số
// con trỏ còn sót ngoài `len` — một con số xác định, lặp lại được, và nói đúng cái cần nói.
func TestLuoiVatPhamKhongGiuConTroDaXoa(t *testing.T) {
	cfg := config.MustLoadDefault()
	cfg.Room.MonsterCount = 0
	cfg.Items.CountPerRoom = 0
	g := New(cfg, world.EmptyMap(cfg.Room.MapSize), 1, 0)

	// Thả rồi xoá nhiều lượt ở CÙNG một chỗ: ô nào cũng có lúc chỉ còn một viên, và đó
	// đúng là trường hợp làm lộ lỗi (hoán đổi với chính mình rồi cắt).
	at := Vec{X: cfg.Room.MapSize / 2, Y: cfg.Room.MapSize / 2}
	for i := 0; i < 200; i++ {
		it := g.addItem("orb", 5, at)
		g.removeItem(it)
	}
	if a := g.GridAudit(); a.Stale != 0 {
		t.Fatalf("còn %d con trỏ sót trong mảng nền của lưới (phải 0)", a.Stale)
	}

	// Thêm nhiều viên vào một ô rồi xoá theo thứ tự ngẫu nhiên: lần này phép hoán đổi
	// thật sự chạy, và ô cuối vẫn phải được xoá.
	//
	// Đếm theo CHÊNH LỆCH, không theo số tuyệt đối: bản đồ đã có sẵn hàng nghìn viên EXP
	// môi trường do `seedOrbs` rải (§6.1).
	base := g.GridAudit().Len
	var items []*Item
	for i := 0; i < 50; i++ {
		items = append(items, g.addItem("orb", 5, at))
	}
	for _, i := range []int{49, 0, 25, 13, 7, 48, 1} {
		g.removeItem(items[i])
		items[i] = nil
	}
	a := g.GridAudit()
	if a.Stale != 0 {
		t.Fatalf("còn %d con trỏ sót sau khi xoá xen kẽ", a.Stale)
	}
	live := 0
	for _, it := range items {
		if it != nil {
			live++
		}
	}
	if a.Len-base != live {
		t.Fatalf("lưới giữ thêm %d viên, vừa thả và còn lại %d", a.Len-base, live)
	}
}

// Slot phải được trả lại, nếu không `knownItems` của mỗi client phình theo thời gian.
func TestSlotVatPhamDuocDungLai(t *testing.T) {
	cfg := config.MustLoadDefault()
	cfg.Room.MonsterCount = 0
	cfg.Items.CountPerRoom = 0
	g := New(cfg, world.EmptyMap(cfg.Room.MapSize), 1, 0)
	before, _, _ := g.SlotAudit()
	at := Vec{X: cfg.Room.MapSize / 2, Y: cfg.Room.MapSize / 2}
	for i := 0; i < 500; i++ {
		g.removeItem(g.addItem("orb", 5, at))
	}
	after, _, _ := g.SlotAudit()
	if after > before+1 {
		t.Fatalf("bảng slot phình từ %d lên %d sau 500 lượt thả/xoá", before, after)
	}
}
