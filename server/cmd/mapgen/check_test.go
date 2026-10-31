package main

import (
	"testing"

	"kiemxoay/server/internal/config"
	"kiemxoay/server/internal/game"
	"kiemxoay/server/internal/geom"
	"kiemxoay/server/internal/world"
)

// Bộ kiểm §8.1b phải BẮT ĐƯỢC lỗi, không chỉ chạy xanh.
//
// Một bộ kiểm không bao giờ đỏ thì vô dụng, và kiểu hỏng đó im lặng: nó vẫn in "ĐẠT" cho
// mọi bản đồ, CI vẫn xanh, và cái túi kín vẫn nằm đó. Nên trước khi tin kết quả trên bản đồ
// thật, phải dựng một bản đồ CỐ TÌNH hỏng và xem nó có kêu không.

// boxedMap: một cái hộp kín bậc 5 quanh một cái Đài — người dưới Luyện Hư không vào được.
func boxedMap(tier int) *world.Map {
	m := world.EmptyMap(12000)
	m.PlayMargin = 420
	const x, y, w = 5000, 5000, 2000
	const t = 300 // bề dày tường, dày hơn thân to nhất để không ai lách qua
	m.Obstacles = []world.Obstacle{
		{ID: 0, Kind: "stone_wall", Tier: tier, Shape: world.Shape{X: x, Y: y, W: w, H: t}},
		{ID: 1, Kind: "stone_wall", Tier: tier, Shape: world.Shape{X: x, Y: y + w - t, W: w, H: t}},
		{ID: 2, Kind: "stone_wall", Tier: tier, Shape: world.Shape{X: x, Y: y, W: t, H: w}},
		{ID: 3, Kind: "stone_wall", Tier: tier, Shape: world.Shape{X: x + w - t, Y: y, W: t, H: w}},
	}
	m.Altars = []world.Altar{{Kind: "small", X: x + w/2, Y: y + w/2, R: 150}}
	return m
}

func TestCheckCatchesTrappedLandmark(t *testing.T) {
	cfg := config.MustLoadDefault()
	prog := game.NewProgression(cfg.Cultivation, cfg.Sword.MaxSwords)
	m := boxedMap(5) // tường bậc 5 = Luyện Hư mới chém được
	marks := landmarksOf(m)
	if len(marks) != 1 {
		t.Fatalf("dựng được %d mốc, muốn 1 cái Đài", len(marks))
	}

	// Nguyên Anh (3) chưa chém được bậc 5 → cái Đài bị nhốt, phải TRƯỢT.
	low, _ := checkRealm(m, prog, cfg, 3, marks)
	if len(low.CutOff) == 0 {
		t.Error("Đài bị nhốt trong hộp bậc 5 mà bộ kiểm không kêu")
	}
	if low.MainFrac >= 0.99 {
		t.Errorf("vùng chính %.0f%% — chỗ trống trong hộp phải tính là bị nhốt", low.MainFrac*100)
	}
	if low.ok() {
		t.Error("bộ kiểm báo ĐẠT cho một bản đồ có Đài không tới được")
	}

	// Luyện Hư (5) chém được bậc 5 → hộp biến mất với họ, phải ĐẠT.
	high, _ := checkRealm(m, prog, cfg, 5, marks)
	if len(high.CutOff) != 0 {
		t.Errorf("Luyện Hư chém được tường bậc 5 mà vẫn báo không tới được: %v", high.CutOff)
	}
	if !high.ok() {
		t.Error("bộ kiểm báo TRƯỢT cho cảnh giới chém thủng được cái hộp")
	}
}

// Luật "bỏ vật cản bậc ≤ r" phải thật sự chạy: cùng một bản đồ, cảnh giới cao hơn thấy ít
// vật cản hơn. Nếu không thì bộ kiểm đang đo bản đồ của người Luyện Khí cho cả 14 cảnh giới.
func TestCheckDropsCuttableObstacles(t *testing.T) {
	cfg := config.MustLoadDefault()
	prog := game.NewProgression(cfg.Cultivation, cfg.Sword.MaxSwords)
	m := boxedMap(3) // bậc 3 = Nguyên Anh
	marks := landmarksOf(m)

	below, _ := checkRealm(m, prog, cfg, 2, marks) // Kim Đan: hộp còn đó
	above, _ := checkRealm(m, prog, cfg, 3, marks) // Nguyên Anh: hộp biến mất
	if below.Density <= above.Density {
		t.Errorf("mật độ ở Kim Đan %.2f%% không cao hơn Nguyên Anh %.2f%% — luật 'bỏ bậc ≤ r' không chạy",
			below.Density*100, above.Density*100)
	}
	if above.Density != 0 {
		t.Errorf("Nguyên Anh chém được mọi vật cản của bản đồ này mà mật độ vẫn %.2f%%", above.Density*100)
	}
}

// Diện tích trống dưới ngưỡng phải TRƯỢT (spec §8.1b: "≥ 60%").
func TestCheckFailsOnTooLittleOpenSpace(t *testing.T) {
	cfg := config.MustLoadDefault()
	prog := game.NewProgression(cfg.Cultivation, cfg.Sword.MaxSwords)
	m := world.EmptyMap(12000)
	m.PlayMargin = 420
	// Một khối đặc phủ gần hết khu chơi được.
	lo := m.PlayMargin
	span := m.Size - 2*lo
	m.Obstacles = []world.Obstacle{
		{ID: 0, Kind: "jade_pillar", Tier: 11, Shape: world.Shape{X: lo, Y: lo, W: span, H: span * 0.8}},
	}
	res, _ := checkRealm(m, prog, cfg, 0, nil)
	if res.ok() {
		t.Errorf("trống chỉ %.0f%% mà vẫn ĐẠT (ngưỡng %.0f%%)", res.OpenFrac*100, minOpenRatio*100)
	}
}

// Mốc ở rìa ô bị chặn vẫn phải tính là TỚI ĐƯỢC nếu cạnh bên vào được.
//
// Tâm một cái cổng rộng 200 có thể rơi đúng vào ô mà thân người to không lọt, trong khi
// cổng vẫn vào được từ bên cạnh. Không nới ra vài ô thì bộ kiểm báo đỏ hàng loạt ở cảnh
// giới cao, và một bộ kiểm kêu sai liên tục thì người ta tắt nó đi.
func TestCheckTolerantAtLandmarkEdge(t *testing.T) {
	cfg := config.MustLoadDefault()
	prog := game.NewProgression(cfg.Cultivation, cfg.Sword.MaxSwords)
	m := world.EmptyMap(12000)
	m.PlayMargin = 420
	// Một cái cột nhỏ đặt ĐÚNG trên tâm cổng; quanh nó trống hoàn toàn.
	m.Obstacles = []world.Obstacle{
		{ID: 0, Kind: "jade_pillar", Tier: 11, Shape: world.Shape{X: 5950, Y: 5950, W: 100, H: 100}},
	}
	marks := []landmark{{At: geom.Vec{X: 6000, Y: 6000}, Name: "cổng 0"}}
	res, _ := checkRealm(m, prog, cfg, 0, marks)
	if len(res.CutOff) != 0 {
		t.Errorf("cổng có lối vào từ mọi phía mà vẫn báo không tới được: %v", res.CutOff)
	}
}
