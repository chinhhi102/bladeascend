package world

import "testing"

// §8.2: "Bot: tìm đường coi vật cản mình chém được là 'đi qua được, tốn thêm 0.3s'."
//
// Bài test dựng một bức tường cắt ngang bản đồ có đúng MỘT chỗ chém được. Người chém được
// chỗ đó phải đi xuyên qua; người không chém được phải đi vòng. Nếu bot dùng `FindPath`
// thường thì cả hai đều đi vòng — tức một Đạo Tổ lội nửa bản đồ để tránh cái rào tre mà họ
// chạm vào là tan.
func TestFindPathCuttingGoesThroughWhatItCanCut(t *testing.T) {
	m := EmptyMap(2000)
	// Tường kín từ y=0 tới y=2000, trừ khối ở giữa (id 1) là loại chém được.
	m.Obstacles = []Obstacle{
		{ID: 0, Kind: "stone_wall", Tier: 5, Shape: Shape{X: 900, Y: 0, W: 200, H: 900}},
		{ID: 1, Kind: "fence", Tier: 0, Shape: Shape{X: 900, Y: 900, W: 200, H: 200}},
		{ID: 2, Kind: "stone_wall", Tier: 5, Shape: Shape{X: 900, Y: 1100, W: 200, H: 900}},
	}
	g := NewGrid(m, 50)
	g.Rebuild(func(int) bool { return true })

	from, to := Vec2(400, 1000), Vec2(1600, 1000)
	none := func(int) bool { return false }
	only0 := func(id int) bool { return id == 1 }

	// Không chém được gì → không có đường nào (tường kín cả bản đồ).
	if p := g.FindPathCutting(from, to, 20000, 1, none); p != nil {
		t.Fatalf("tường kín mà vẫn tìm ra đường %d bước", len(p))
	}
	// Chém được khối giữa → đi thẳng qua.
	p := g.FindPathCutting(from, to, 20000, 1, only0)
	if p == nil {
		t.Fatal("chém được rào ở giữa mà không tìm ra đường")
	}
	crossed := false
	for _, w := range p {
		if w.X > 900 && w.X < 1100 {
			crossed = true
		}
	}
	if !crossed {
		t.Error("đường đi không xuyên qua chỗ chém được — bot vẫn đi vòng")
	}
}

// Phí 0.3 giây phải THẬT: đi xuyên vật cản chém được chỉ nên chọn khi nó ngắn hơn hẳn, chứ
// không phải "cứ chém là nhanh nhất". Đường vòng ngắn hơn thì bot phải đi vòng.
func TestFindPathCuttingPrefersTheOpenWayWhenPenaltyIsHigh(t *testing.T) {
	m := EmptyMap(2000)
	// Một khối rào duy nhất chắn đúng đường thẳng; quanh nó trống hoàn toàn.
	m.Obstacles = []Obstacle{{ID: 0, Kind: "fence", Tier: 0, Shape: Shape{X: 950, Y: 950, W: 100, H: 100}}}
	g := NewGrid(m, 50)
	g.Rebuild(func(int) bool { return true })

	from, to := Vec2(700, 1000), Vec2(1300, 1000)
	cut := func(int) bool { return true }
	// Phí lớn (50 ô): lách sang bên cạnh chỉ tốn ~1 ô nên bot phải chọn lách.
	for _, w := range g.FindPathCutting(from, to, 20000, 50, cut) {
		if g.Blocked(w) {
			t.Fatalf("phí chém 50 ô mà bot vẫn chém xuyên tại %+v", w)
		}
	}
	// Phí 0: đi xuyên là đường ngắn nhất, và lưới phải cho phép.
	hit := false
	for _, w := range g.FindPathCutting(from, to, 20000, 0, cut) {
		if g.Blocked(w) {
			hit = true
		}
	}
	if !hit {
		t.Error("phí 0 mà vẫn đi vòng — ô chém được không hề rẻ hơn")
	}
}
