package geom

import (
	"math"
	"testing"
)

// SegSegDist2 từng sai dấu tham số giao điểm: nó báo hai đoạn cắt nhau khi phần KÉO DÀI
// NGƯỢC của chúng cắt nhau. Trong game, hai thanh kiếm cách nhau hai trăm đơn vị vẫn bị
// tính là va vào nhau. Test này khoá lại cả hai chiều.
func TestSegSegDist2(t *testing.T) {
	cases := []struct {
		name                   string
		p1, q1, p2, q2         Vec
		want                   float64
	}{
		{"cắt nhau hình chữ thập", Vec{-1, 0}, Vec{1, 0}, Vec{0, -1}, Vec{0, 1}, 0},
		{"chạm đúng đầu mút", Vec{0, 0}, Vec{1, 0}, Vec{1, 0}, Vec{2, 0}, 0},
		{"song song cách 3", Vec{0, 0}, Vec{10, 0}, Vec{0, 3}, Vec{10, 3}, 9},
		// Hai đoạn này KHÔNG cắt nhau; chỉ khi kéo dài ngược cả hai thì mới gặp nhau ở
		// gốc toạ độ. Bản cũ trả về 0 ở đây.
		{"chỉ cắt khi kéo dài ngược", Vec{1, 0}, Vec{5, 0}, Vec{0, 1}, Vec{0, 5}, 2},
		{"rời nhau hẳn", Vec{0, 0}, Vec{1, 0}, Vec{10, 10}, Vec{11, 10}, 81 + 100},
	}
	for _, c := range cases {
		if got := SegSegDist2(c.p1, c.q1, c.p2, c.q2); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("%s: SegSegDist2 = %.4f, muốn %.4f", c.name, got, c.want)
		}
	}
}
