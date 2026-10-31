package room

import (
	"testing"

	"kiemxoay/server/internal/game"
)

// Trần sự kiện va kiếm: giữ cái GẦN NGƯỜI XEM nhất, bỏ phần còn lại (ngân sách gói §13.6).
//
// Va kiếm là sự kiện đông nhất của gói — đo trên bản đồ v6.12 một gói đỉnh có 42 `RingFlip`
// trong 60 sự kiện, chiếm 1,8 KB / 27% cả gói. Bỏ bớt được vì nó KHÔNG mang luật chơi:
// chiều quay mới nằm ở trường `d` của mỗi người trong mọi gói.
func TestCapRingFlipsKeepsNearestAndSpareOthers(t *testing.T) {
	viewer := &game.Player{ID: 1, Pos: game.Vec{X: 1000, Y: 1000}}
	var events []game.Event
	// 20 va kiếm, cái thứ i cách người xem 100×(i+1).
	for i := 0; i < 20; i++ {
		events = append(events, game.Event{Kind: "RingFlip",
			X: float32(1000 + 100*(i+1)), Y: 1000})
	}
	// Cùng với vài sự kiện KHÁC — chúng phải đi qua nguyên vẹn, không bị cắt lây.
	events = append(events, game.Event{Kind: "Death", X: 9000, Y: 9000})
	events = append(events, game.Event{Kind: "ItemPicked", X: 9000, Y: 9000})

	out := capRingFlips(events, viewer, 5)

	flips, others := 0, 0
	for _, e := range out {
		if e.Kind == "RingFlip" {
			flips++
			// Chỉ 5 cái gần nhất: xa nhất được giữ là cái thứ 5, cách 500.
			if d := float64(e.X) - 1000; d > 500.5 {
				t.Errorf("giữ lại va kiếm cách %.0f — phải giữ 5 cái gần nhất (≤ 500)", d)
			}
		} else {
			others++
		}
	}
	if flips != 5 {
		t.Errorf("giữ %d va kiếm, trần là 5", flips)
	}
	if others != 2 {
		t.Errorf("sự kiện khác còn %d, phải còn đủ 2 — trần chỉ được cắt RingFlip", others)
	}
}

// Dưới trần thì không đụng vào gì cả, và trần 0 nghĩa là không giới hạn.
func TestCapRingFlipsNoOpBelowLimit(t *testing.T) {
	viewer := &game.Player{ID: 1, Pos: game.Vec{X: 0, Y: 0}}
	events := []game.Event{
		{Kind: "RingFlip", X: 100}, {Kind: "RingFlip", X: 200}, {Kind: "Hit", X: 300},
	}
	if got := len(capRingFlips(events, viewer, 5)); got != 3 {
		t.Errorf("3 sự kiện dưới trần 5 mà còn %d", got)
	}
	if got := len(capRingFlips(events, viewer, 0)); got != 3 {
		t.Errorf("trần 0 = không giới hạn, mà còn %d", got)
	}
}
