package game

import "testing"

// Ô lưới giờ là slice và xoá bằng cách hoán đổi với phần tử cuối (tối ưu mục 26). Kiểu xoá
// đó rất dễ làm mất phần tử khác nếu `slot` không được cập nhật, nên phải có test riêng:
// thêm, xoá giữa chừng, xoá hai lần, và quét theo hình chữ nhật đều phải đúng.
func TestItemGridAddRemoveEach(t *testing.T) {
	ig := newItemGrid(1000, 100)
	mk := func(id uint32, x, y float64) *Item {
		it := &Item{ID: id, Pos: Vec{X: x, Y: y}}
		ig.add(it)
		return it
	}
	a, b, c := mk(1, 50, 50), mk(2, 60, 60), mk(3, 70, 70)
	d := mk(4, 950, 950) // ô khác hẳn

	got := func(lo, hi Vec) map[uint32]bool {
		out := map[uint32]bool{}
		ig.each(lo, hi, func(it *Item) { out[it.ID] = true })
		return out
	}

	all := got(Vec{}, Vec{X: 1000, Y: 1000})
	for _, id := range []uint32{1, 2, 3, 4} {
		if !all[id] {
			t.Fatalf("quét cả lưới thiếu viên %d: %v", id, all)
		}
	}

	// Xoá viên ở GIỮA ô: hai viên còn lại phải còn nguyên, không được mất theo.
	ig.remove(b)
	left := got(Vec{}, Vec{X: 100, Y: 100})
	if len(left) != 2 || !left[a.ID] || !left[c.ID] {
		t.Fatalf("xoá viên giữa làm hỏng ô: %v", left)
	}

	// Xoá hai lần không được đụng vào viên khác.
	ig.remove(b)
	if left = got(Vec{}, Vec{X: 100, Y: 100}); len(left) != 2 {
		t.Fatalf("xoá lần hai làm mất viên: %v", left)
	}

	// Xoá viên vừa bị hoán đổi lên chỗ của b (c đã nhảy vào slot của b).
	ig.remove(c)
	if left = got(Vec{}, Vec{X: 100, Y: 100}); len(left) != 1 || !left[a.ID] {
		t.Fatalf("xoá viên đã bị hoán đổi sai: %v", left)
	}

	// Hình chữ nhật chỉ chạm ô gần gốc thì không được trả về viên ở góc xa.
	if near := got(Vec{}, Vec{X: 100, Y: 100}); near[d.ID] {
		t.Fatalf("quét ô gần lại trả về viên ở xa")
	}
}
