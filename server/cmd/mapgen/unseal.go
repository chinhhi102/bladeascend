package main

import (
	"math"

	"bladeascend/server/internal/config"
	"bladeascend/server/internal/game"
)

// Đục lối cho những "túi" bị nhốt — §8.1b "Toàn bộ chỗ trống phải liên thông".
//
// Đảo vật cản đặt cạnh địa hình sẵn có (dải vách của Linh Sơn Cốc, rặng tre của Trúc Lâm)
// thỉnh thoảng khép kín một khoảng giữa chúng. Đo được: Linh Sơn Cốc có một cái sân kín
// **45 ô lưới liền nhau** ≈ 335 × 335 đơn vị — thừa chỗ cho mấy người đánh nhau, nhưng
// không ai vào được. Bộ kiểm cũ bỏ sót vì nó đo TỈ LỆ (46 ô / 82 000 = 0,06%) thay vì hỏi
// "có phòng kín nào không".
//
// Vì sao sửa SAU khi đặt chứ không phòng lúc đặt: một cái túi là tính chất của **cả cụm**
// vật cản quanh nó, không của riêng khối vừa đặt. Khối thứ tư có thể khép kín một khoảng mà
// ba khối trước đó hoàn toàn hợp lệ. Muốn phòng lúc đặt thì mỗi lần đặt phải flood fill lại
// cả bản đồ — hàng nghìn lần mỗi khu.
//
// Cách đục: tìm túi, rồi **xoá hẳn một khối vật cản** nằm trên vành túi. Xoá chứ không thu
// nhỏ vì thu nhỏ hay để lại một khe hẹp hơn hành lang tối thiểu — tức vẫn là túi với người
// to hơn một chút, và bộ kiểm lại đỏ ở cảnh giới khác.

// unsealCells: túi từ ngần này ô trở lên thì phải đục. Khớp `maxTrapCells` của bộ kiểm —
// hai chỗ phải cùng một con số, nếu không bộ sinh để lại đúng thứ bộ kiểm sẽ bắt.
const unsealCells = maxTrapCells

// unsealStep đục mọi túi bị nhốt ở MỌI cảnh giới r = 0..13.
//
// Phải quét đủ 14 cảnh giới vì tập vật cản khác nhau theo cảnh giới (bậc ≤ r thì chém được
// nên coi như không có) VÀ thân người to dần. Một khoảng thông với Luyện Khí có thể bị khép
// với Đại Thừa: khe 100 thì thân 42 lọt, thân 150 thì không.
func (g *gen) unsealStep(prog realmSizes) {
	for r := 0; r < prog.count; r++ {
		// Trần số lần đục cho mỗi cảnh giới: mỗi lần xoá một khối rồi tính lại. Trần để
		// không quay vô hạn nếu có túi nào không khối nào đục được (ví dụ chỉ bao bởi biên).
		for pass := 0; pass < 40; pass++ {
			if !g.unsealOnce(r, prog.bodyR[r]) {
				break
			}
		}
	}
}

// realmSizes là bán kính thân to nhất của từng cảnh giới, tính sẵn để `unsealStep` không
// phải biết gì về `game.Progression`.
type realmSizes struct {
	count int
	bodyR []float64
}

// unsealOnce tìm MỘT túi và đục nó. Trả về false khi không còn túi nào.
func (g *gen) unsealOnce(realm int, bodyR float64) bool {
	lo, hi := g.b.playable()
	n := int((hi - lo) / coverCell)
	if n <= 0 {
		return false
	}
	// Ô bị chặn với cảnh giới này: chỉ tính vật cản bậc > realm (bậc ≤ realm thì họ chém
	// được, nên với họ bức tường đó không tồn tại).
	blocked := make([]bool, n*n)
	for i := range g.b.solid {
		s := &g.b.solid[i]
		if s.tier <= realm {
			continue
		}
		f := footprintRect(s.r, g.b.footW, g.b.footH)
		stampCircleInflated(blocked, n, lo, f, bodyR)
	}
	label, sizes, main := components(blocked, n)
	if main < 0 {
		return false
	}
	// Túi lớn nhất khác vùng chính.
	worst, worstSize := -1, 0
	for id, sz := range sizes {
		if id != main && sz > worstSize {
			worst, worstSize = id, sz
		}
	}
	if worst < 0 || worstSize < unsealCells {
		return false
	}
	return g.carveOpening(label, n, lo, worst, main, realm)
}

// carveOpening xoá một khối vật cản ngăn giữa túi `trap` và vùng chính.
//
// Chọn khối nào: khối có ô lưới vừa chạm TÚI vừa chạm VÙNG CHÍNH (tức đúng cái nút chai),
// và trong số đó lấy khối NHỎ NHẤT — xoá cái nhỏ thì bản đồ đổi ít nhất.
func (g *gen) carveOpening(label []int, n int, lo float64, trap, main, realm int) bool {
	best, bestArea := -1, math.Inf(1)
	for i := range g.b.solid {
		s := &g.b.solid[i]
		if s.tier <= realm || s.tier >= config.TierBorder {
			continue // vô hình với cảnh giới này, hoặc là biên khu — không bao giờ xoá biên
		}
		touchTrap, touchMain := false, false
		x0, y0, x1, y1 := cellSpan(s.r, n, lo)
		for y := y0; y <= y1 && !(touchTrap && touchMain); y++ {
			for x := x0; x <= x1; x++ {
				// Vành 2 ô quanh khối: ô của chính khối thì bị chặn nên không có nhãn.
				for _, d := range [4][2]int{{2, 0}, {-2, 0}, {0, 2}, {0, -2}} {
					nx, ny := x+d[0], y+d[1]
					if nx < 0 || ny < 0 || nx >= n || ny >= n {
						continue
					}
					switch label[ny*n+nx] {
					case trap:
						touchTrap = true
					case main:
						touchMain = true
					}
				}
			}
		}
		if !touchTrap || !touchMain {
			continue
		}
		if a := (s.r.x1 - s.r.x0) * (s.r.y1 - s.r.y0); a < bestArea {
			best, bestArea = i, a
		}
	}
	if best < 0 {
		return false
	}
	g.b.removeSolid(best)
	return true
}

// stampCircleInflated đánh dấu mọi ô mà tâm ô nằm trong `f` nở ra `r` — tức chỗ một hình
// tròn bán kính `r` không đứng được. Cùng phép với `blockedAt` của bộ kiểm.
func stampCircleInflated(blocked []bool, n int, lo float64, f rect, r float64) {
	x0 := int((f.x0 - r - lo) / coverCell)
	x1 := int((f.x1+r-lo)/coverCell) + 1
	y0 := int((f.y0 - r - lo) / coverCell)
	y1 := int((f.y1+r-lo)/coverCell) + 1
	for y := max(y0, 0); y <= min(y1, n-1); y++ {
		cy := lo + (float64(y)+0.5)*coverCell
		for x := max(x0, 0); x <= min(x1, n-1); x++ {
			cx := lo + (float64(x)+0.5)*coverCell
			dx := math.Max(math.Max(f.x0-cx, 0), cx-f.x1)
			dy := math.Max(math.Max(f.y0-cy, 0), cy-f.y1)
			if dx*dx+dy*dy < r*r {
				blocked[y*n+x] = true
			}
		}
	}
}

func cellSpan(r rect, n int, lo float64) (x0, y0, x1, y1 int) {
	x0 = max(int((r.x0-lo)/coverCell)-1, 0)
	y0 = max(int((r.y0-lo)/coverCell)-1, 0)
	x1 = min(int((r.x1-lo)/coverCell)+1, n-1)
	y1 = min(int((r.y1-lo)/coverCell)+1, n-1)
	return
}

// components gắn nhãn vùng trống liên thông (4 hướng) và trả về vùng lớn nhất.
func components(blocked []bool, n int) (label []int, sizes []int, main int) {
	label = make([]int, n*n)
	for i := range label {
		label[i] = -1
	}
	main = -1
	for s := 0; s < n*n; s++ {
		if blocked[s] || label[s] >= 0 {
			continue
		}
		id, size := len(sizes), 0
		stack := []int{s}
		label[s] = id
		for len(stack) > 0 {
			i := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			size++
			x, y := i%n, i/n
			for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				nx, ny := x+d[0], y+d[1]
				if nx < 0 || ny < 0 || nx >= n || ny >= n {
					continue
				}
				if j := ny*n + nx; !blocked[j] && label[j] < 0 {
					label[j] = id
					stack = append(stack, j)
				}
			}
		}
		sizes = append(sizes, size)
		if main < 0 || size > sizes[main] {
			main = id
		}
	}
	return label, sizes, main
}

// realmSizesOf tính sẵn bán kính thân TO NHẤT của từng cảnh giới (tầng 10 với người phàm —
// spec §8.1b), đúng bảng mà `cmd/mapgen --check` dùng.
//
// Tính một lần rồi truyền vào thay vì để `unsealStep` tự hỏi `game.Progression`: nó gọi
// trong vòng lặp 14 cảnh giới × 40 lượt đục, và `NewProgression` dựng lại cả bảng ngưỡng
// EXP mỗi lần.
func realmSizesOf(cfg *config.Config) realmSizes {
	prog := game.NewProgression(cfg.Cultivation, cfg.Sword.MaxSwords)
	n := len(cfg.Cultivation.Realms) + prog.TienCount()
	out := realmSizes{count: n, bodyR: make([]float64, n)}
	for r := 0; r < n; r++ {
		out.bodyR[r] = cfg.Player.BodyR * prog.Scale(r*prog.Tiers()+prog.Tiers()-1)
	}
	return out
}
