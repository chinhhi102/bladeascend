package main

import (
	"fmt"
	"math"

	"kiemxoay/server/internal/config"
)

// Thuật toán sinh cho 4 khu mới của thế giới 5 khu (spec §8.0, CHANGES_v5 mục 6).
//
// Khu 1 (Thiên Đình) giữ nguyên thuật toán §8.1 trong gen.go. Bốn khu dưới đây mỗi khu một
// kiểu bố cục riêng, nhưng đều phải tuân thủ cùng một bộ luật: tường biên không phá được,
// **mọi chỗ trống phải đi tới được nhau** (kiểm bằng flood fill), trống 300 quanh cổng và
// quanh đài, và không có khối hình học lộ liễu (§8.6).
//
// Mỗi khu có **4 cổng** thuộc mạng chung của cả thế giới và **2 Đài Tiên Khí nhỏ**; riêng
// Thiên Đình giữ 4 đài nhỏ + Thiên Tiên Đài ở giữa.

// zoneDef mô tả một khu: id file, tên hiển thị, nền, màu cổng.
type zoneDef struct {
	id     string
	name   string
	biome  string
	gateCo string
	build  func(*gen) error
}

// zoneDefs theo đúng thứ tự bảng §8.0.
var zoneDefs = []zoneDef{
	{"thien_dinh", "Thiên Đình", "grass", "gold", (*gen).thienDinh},
	{"linh_son_coc", "Linh Sơn Cốc", "dirt", "slate", (*gen).linhSonCoc},
	{"truc_lam", "Trúc Lâm Mê Cung", "grass", "green", (*gen).trucLam},
	{"pham_tran", "Phàm Trấn", "concrete", "red", (*gen).phamTran},
	{"bang_nguyet_ho", "Băng Nguyệt Hồ", "snow", "cyan", (*gen).bangNguyetHo},
}

// gateNet là tên mạng cổng chung của CẢ THẾ GIỚI (spec §8.4: 20 cổng, một mạng).
const gateNet = "world"

// thienDinh là khu 1 — giữ nguyên thuật toán §8.1.
func (g *gen) thienDinh() error {
	g.frame()
	g.roadsStep()
	// Cổng đặt TRƯỚC vật cản: b.gate() ghi vùng giữ trống 300 quanh cổng, nên vật cản sinh
	// sau sẽ tự tránh. Đặt sau thì không có cách nào bảo đảm luật "trống 300 quanh cổng".
	g.gatesStep()
	g.heavenDocks()
	g.quadrantContent()
	g.terrainZones()
	g.decorStep()
	return nil
}

// heavenDocks đặt 4 Bến Thiên Môn ở Thiên Đình (spec §8.4a).
//
// Mỗi khu nối cố định với một bến theo hướng: Linh Sơn Cốc → Bắc, Trúc Lâm → Đông,
// Phàm Trấn → Nam, Băng Nguyệt Hồ → Tây. Cố định chứ không ngẫu nhiên là có lý do: người
// chơi quen đường sẽ biết ai vừa từ khu nào lên, và đó là thông tin đáng giá ở khu trung
// tâm — khác hẳn cổng thường vốn cố tình giấu đích đến.
func (g *gen) heavenDocks() {
	c := g.size / 2
	const d = 1400
	// Thứ tự: khu 1 Bắc, khu 2 Đông, khu 3 Nam, khu 4 Tây.
	for i, p := range []Vec2{{c, c - d}, {c + d, c}, {c, c + d}, {c - d, c}} {
		g.b.heavenDock(p.X, p.Y, i+1)
	}
}

// fourGates đặt 4 cổng rải đều bốn phía của khu (spec §8.4: "khu khác: rải đều 4 phía").
func (g *gen) fourGates(color string) {
	s := g.size
	for _, p := range [][2]float64{{s * 0.5, s * 0.14}, {s * 0.86, s * 0.5}, {s * 0.5, s * 0.86}, {s * 0.14, s * 0.5}} {
		g.b.gate(p[0], p[1], gateNet, color)
	}
}

// twoAltars đặt 2 Đài Tiên Khí nhỏ (spec §8.0: mỗi khu ngoài Thiên Đình có 2 đài).
func (g *gen) twoAltars(a, bb Vec2) {
	g.b.altar("small", a.X, a.Y, 150, 400)
	g.b.altar("small", bb.X, bb.Y, 150, 400)
}

// ---------------------------------------------------------------------------
// Khu 2 — Linh Sơn Cốc: vách đá chia hẻm núi, nối bằng đèo hẹp và cầu treo
// ---------------------------------------------------------------------------

// linhSonCoc dựng các dãy vách đá chạy ngang chia khu thành nhiều hẻm núi, mỗi dãy chừa
// vài **đèo hẹp** 200–300 để qua lại (spec §8.0). Người to khó lách qua đèo — đó là điểm
// đặc biệt của khu này.
//
// Vì sao vách đá chạy NGANG chứ không toả tròn: hẻm núi chỉ có nghĩa khi nó buộc người ta
// đi dọc theo một hướng rồi tranh nhau cái đèo. Vách toả tròn thì chỗ nào cũng vòng được.
func (g *gen) linhSonCoc() error {
	s := g.size
	lo, hi := g.b.playable()
	g.twoAltars(Vec2{s * 0.22, s * 0.24}, Vec2{s * 0.78, s * 0.76})
	// Thiên Môn trên một bình nguyên giữa các hẻm núi (spec §8.4a).
	g.b.heavenGate(s/2, s/2)
	g.fourGates("slate")

	// Vách đá xếp SO LE: dãy lẻ chạm mép trái và hụt mép phải, dãy chẵn ngược lại. Nhờ vậy
	// đường đi uốn thành hình rắn bò từ trên xuống dưới — vẫn ra hẻm núi, mà **không bao
	// giờ nhốt kín một vùng nào**. Bản đầu cho vách chạy suốt từ mép này sang mép kia và
	// chỉ dựa vào mấy cái đèo; bộ kiểm báo ngay 3,3% ô trống bị nhốt.
	ridges := 5
	for i := 1; i <= ridges; i++ {
		y := lo + (hi-lo)*float64(i)/float64(ridges+1) + g.b.rnd(-200, 200)
		x0, x1 := lo+120, hi-900
		if i%2 == 0 {
			x0, x1 = lo+900, hi-120
		}
		g.cliffRidge(x0, x1, y, 1+g.b.rng.Intn(2))
	}
	// Hai "đỉnh núi": cao nguyên nhỏ quanh đài, tường bao có đúng 2 lối lên.
	for _, c := range []Vec2{{s * 0.22, s * 0.24}, {s * 0.78, s * 0.76}} {
		g.plateau(c, 900, 2)
	}
	// Đá tảng rải rác trong lòng hẻm cho đỡ trống, nhưng thưa — khu này sống bằng địa hình.
	for _, p := range g.poisson(900, 500) {
		g.b.boulder(p.X, p.Y, g.b.rnd(90, 170), 160)
	}
	g.b.zone("bush", s*0.45, s*0.1, 900, 700, true)
	g.b.organicZone("mud", s*0.6, s*0.35, 700, 0.6)
	for _, d := range []string{"rock_small", "pine"} {
		g.b.decor(d, 70)
	}
	return nil
}

// cliffRidge dựng một dãy vách đá chạy từ x0 tới x1 ở độ cao y, chừa `passes` đèo hẹp.
//
// Đèo rộng 200–300 đúng như spec; chỗ đèo còn đặt một "cầu treo" (vùng đường) để người chơi
// nhìn là biết đi qua được, không phải dò dẫm.
func (g *gen) cliffRidge(x0, x1, y float64, passes int) {
	const thick = 220
	gaps := make([]float64, 0, passes)
	for i := 1; i <= passes; i++ {
		gaps = append(gaps, x0+(x1-x0)*float64(i)/float64(passes+1)+g.b.rnd(-400, 400))
	}
	cur := x0
	for _, gx := range gaps {
		w := g.b.rnd(200, 300)
		if gx-w/2 > cur {
			g.cliffRun(cur, y, gx-w/2-cur, thick)
		}
		// Cầu treo: nền đường ở ngay cái đèo, rộng đúng bằng đèo.
		g.b.zone("road", gx-w/2, y-thick/2-40, w, thick+80, false)
		cur = gx + w/2
	}
	if x1 > cur {
		g.cliffRun(cur, y, x1-cur, thick)
	}
}

// cliffRun xẻ một đoạn vách dài thành nhiều khối lệch nhau để không ra một thanh chữ nhật
// dài thẳng tắp (spec §8.6: không để lộ khối hình học).
func (g *gen) cliffRun(x, y, length, thick float64) {
	for cur := x; cur < x+length; {
		w := math.Min(g.b.rnd(300, 700), x+length-cur)
		dy := g.b.rnd(-60, 60)
		g.b.wall(cur, y-thick/2+dy, w, thick)
		cur += w
	}
}

// plateau bao quanh một điểm bằng tường vòng có đúng `ways` lối lên.
func (g *gen) plateau(c Vec2, r float64, ways int) {
	angles := make([]float64, 0, ways)
	for i := 0; i < ways; i++ {
		angles = append(angles, float64(i)*2*math.Pi/float64(ways)+g.b.rnd(0, 1))
	}
	// gapArc là ĐỘ DÀI CUNG tính bằng đơn vị, không phải radian — 700 là một lối lên rộng
	// rãi. Lần đầu tôi truyền 0.5 tưởng là radian, cao nguyên bị bịt kín và bộ kiểm báo
	// ngay 3,3% ô trống bị nhốt.
	g.ringWall(c, r, 200, angles, 700)
}

// ---------------------------------------------------------------------------
// Khu 3 — Trúc Lâm Mê Cung: mê cung rặng tre, nhiều ngõ cụt, bụi gấp 3
// ---------------------------------------------------------------------------

// trucLam dựng mê cung bằng thuật toán **đào đường lùi dần** (recursive backtracking) trên
// lưới ô vuông: bắt đầu từ một ô, đập tường sang ô hàng xóm chưa thăm, kẹt thì lùi lại.
//
// Chọn thuật toán này vì nó bảo đảm **mọi ô đều đi tới được nhau** theo đúng cách nó dựng —
// không phải rải tường ngẫu nhiên rồi cầu trời flood fill không báo lỗi. Nó cũng sinh ra
// nhiều ngõ cụt, đúng thứ spec muốn cho khu phục kích.
func (g *gen) trucLam() error {
	s := g.size
	lo, hi := g.b.playable()
	const cells = 11
	cell := (hi - lo) / cells
	// Hai khoảng trống giữa mê cung để đặt đài: đục hẳn 2 ô liền nhau.
	a := Vec2{lo + cell*2.5, hi - cell*2.5}
	bb := Vec2{hi - cell*2.5, lo + cell*2.5}
	g.twoAltars(a, bb)
	// Thiên Môn ở khoảng trống trung tâm mê cung (spec §8.4a) — đường vào ngoằn ngoèo.
	g.b.heavenGate(s/2, s/2)
	g.fourGates("green")

	walls := carveMaze(cells, func(n int) int { return g.b.rng.Intn(n) })
	for _, w := range walls {
		x := lo + float64(w.cx)*cell
		y := lo + float64(w.cy)*cell
		if w.vertical {
			g.bambooRow(x, y, cell, true)
		} else {
			g.bambooRow(x, y, cell, false)
		}
	}
	// Bụi rậm GẤP 3 so với khu thường (spec §8.0) — đây là khu tầm nhìn ngắn.
	for i := 0; i < 36; i++ {
		g.b.zone("bush", g.b.rnd(lo+300, hi-900), g.b.rnd(lo+300, hi-900), g.b.rnd(400, 800), g.b.rnd(350, 700), true)
	}
	for i := 0; i < 4; i++ { // vài ao
		g.b.organicZone("water", g.b.rnd(lo+900, hi-900), g.b.rnd(lo+900, hi-900), g.b.rnd(400, 700), 0.7)
	}
	// Rặng tre không phá được (spec §8.0), nhưng rải thêm ít xe và sạp bỏ hoang trong các
	// ngõ cụt: có thứ đập được thì địa hình còn đổi đôi chút, và mê cung bớt đơn điệu.
	for i := 0; i < 30; i++ {
		x, y := g.b.rnd(lo+400, hi-700), g.b.rnd(lo+400, hi-700)
		kind := "cart"
		if g.b.rng.Intn(2) == 0 {
			kind = "stall"
		}
		g.b.obstacle(kind, x, y, g.b.rnd(180, 280), g.b.rnd(140, 220), 90)
	}
	for _, d := range []string{"bamboo", "bush_small"} {
		g.b.decor(d, 90)
	}
	return nil
}

// bambooRow đặt một đoạn rặng tre dài `length`, xẻ thành nhiều khúc lệch nhau.
func (g *gen) bambooRow(x, y, length float64, vertical bool) {
	const thick = 130
	for cur := 0.0; cur < length; {
		seg := math.Min(g.b.rnd(220, 420), length-cur)
		jitter := g.b.rnd(-25, 25)
		if vertical {
			g.b.obstacle("bamboo_wall", x-thick/2+jitter, y+cur, thick, seg, 0)
		} else {
			g.b.obstacle("bamboo_wall", x+cur, y-thick/2+jitter, seg, thick, 0)
		}
		cur += seg
	}
}

// mazeWall là một đoạn tường của mê cung: góc trên-trái ô và hướng.
type mazeWall struct {
	cx, cy   int
	vertical bool
}

// carveMaze trả về danh sách tường CÒN LẠI sau khi đào đường lùi dần trên lưới n×n.
//
// Tách riêng khỏi `gen` và nhận hàm random ngoài vào để test được mà không cần dựng cả bộ
// sinh bản đồ.
func carveMaze(n int, intn func(int) int) []mazeWall {
	// right[x][y]: còn tường giữa ô (x,y) và (x+1,y)? down tương tự theo trục dọc.
	right := make([][]bool, n)
	down := make([][]bool, n)
	seen := make([][]bool, n)
	for i := 0; i < n; i++ {
		right[i], down[i], seen[i] = make([]bool, n), make([]bool, n), make([]bool, n)
		for j := 0; j < n; j++ {
			right[i][j], down[i][j] = true, true
		}
	}
	type cellPos struct{ x, y int }
	stack := []cellPos{{0, 0}}
	seen[0][0] = true
	for len(stack) > 0 {
		c := stack[len(stack)-1]
		var next []cellPos
		for _, d := range []cellPos{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := c.x+d.x, c.y+d.y
			if nx >= 0 && ny >= 0 && nx < n && ny < n && !seen[nx][ny] {
				next = append(next, cellPos{nx, ny})
			}
		}
		if len(next) == 0 {
			stack = stack[:len(stack)-1]
			continue
		}
		p := next[intn(len(next))]
		switch {
		case p.x > c.x:
			right[c.x][c.y] = false
		case p.x < c.x:
			right[p.x][p.y] = false
		case p.y > c.y:
			down[c.x][c.y] = false
		default:
			down[p.x][p.y] = false
		}
		seen[p.x][p.y] = true
		stack = append(stack, p)
	}
	var out []mazeWall
	for x := 0; x < n; x++ {
		for y := 0; y < n; y++ {
			if right[x][y] && x < n-1 {
				out = append(out, mazeWall{x + 1, y, true})
			}
			if down[x][y] && y < n-1 {
				out = append(out, mazeWall{x, y + 1, false})
			}
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Khu 4 — Phàm Trấn: lưới phố, nhà, chợ đêm
// ---------------------------------------------------------------------------

// phamTran dựng một thị trấn: lưới phố thẳng, các dãy nhà xen giữa, chợ đêm có hẻm hẹp
// 150/250/350, và 2 quảng trường nhỏ đặt đài.
//
// Khu này có **nhiều vật cản phá được nhất** (spec §8.0) nên địa hình đổi liên tục trong
// trận — tường gạch và sạp hàng đều đập được, khác hẳn vách đá hay rặng tre.
func (g *gen) phamTran() error {
	s := g.size
	lo, hi := g.b.playable()
	sqA := Vec2{s * 0.27, s * 0.27}
	sqB := Vec2{s * 0.73, s * 0.73}
	g.twoAltars(sqA, sqB)
	// Thiên Môn ở quảng trường chính giữa thị trấn (spec §8.4a).
	g.b.heavenGate(s/2, s/2)
	g.fourGates("red")

	// Lưới phố — nhưng **cố tình không đều**. Bản đầu tôi chia đúng 1 800 một ô và ra một
	// cái bàn cờ hoàn hảo: §8.6 cấm đúng kiểu đó ("không khối hình học lộ ra"), và thị trấn
	// thật thì phố nào cũng lệch. Giờ mỗi con phố xê dịch ±260 và khoảng cách giữa chúng
	// thay đổi 1 500–2 100, nên không hàng nào thẳng hàng với hàng nào.
	xs := g.streetLines(lo, hi)
	ys := g.streetLines(lo, hi)
	// Xê dịch cả con phố một lần, không xê dịch riêng từng đầu: lệch hai đầu thì phố hoá
	// chéo và hộp bao vùng đường phình ra gấp đôi bề rộng thật.
	for _, x := range xs {
		d := g.b.rnd(-140, 140)
		g.addRoad(Vec2{x + d, lo + 120}, Vec2{x + d, hi - 120}, g.b.rnd(260, 340))
	}
	for _, y := range ys {
		d := g.b.rnd(-140, 140)
		g.addRoad(Vec2{lo + 120, y + d}, Vec2{hi - 120, y + d}, g.b.rnd(260, 340))
	}
	// Một con đường chéo cắt qua thị trấn, phá hẳn cảm giác bàn cờ.
	g.addRoad(Vec2{lo + 400, lo + 1800}, Vec2{hi - 1800, hi - 400}, 340)

	// Lấp từng ô phố bằng VÀI công trình, không phải một cái ở giữa. Đây là khu "nhiều vật
	// cản phá được nhất" (spec §8.0) — bản đầu chỉ đặt một nhà mỗi ô và cả thị trấn ra đúng
	// 44 vật cản, trống hơn cả Thiên Đình.
	for i := 0; i+1 < len(xs); i++ {
		for j := 0; j+1 < len(ys); j++ {
			x0, x1 := xs[i]+260, xs[i+1]-260
			y0, y1 := ys[j]+260, ys[j+1]-260
			if x1-x0 < 500 || y1-y0 < 500 {
				continue
			}
			if g.b.rng.Intn(7) == 0 {
				continue // thỉnh thoảng bỏ trống một ô làm sân, cho thị trấn có chỗ thở
			}
			g.fillBlock(x0, y0, x1, y1, sqA, sqB)
		}
	}
	// Hẻm hẹp đúng ba bề rộng của spec.
	for i, w := range []float64{150, 250, 350} {
		g.b.alley(lo+float64(i+1)*2400, lo+1200, 1500, w, i%2 == 0)
		g.b.alley(lo+float64(i+1)*2100, hi-2700, 1500, w, i%2 == 1)
	}
	g.b.organicZone("mud", s*0.5, s*0.82, 650, 0.6)
	for _, d := range []string{"lantern", "well", "banner"} {
		g.b.decor(d, 60)
	}
	return nil
}

// fillBlock nhồi công trình vào một ô phố.
//
// `house()` nhận GÓC TRÊN-TRÁI chứ không phải tâm, và nhà rộng tới 900 — bản đầu tôi truyền
// một điểm ngẫu nhiên trong ô làm tâm nên nhà tràn ra đường, bị từ chối, và cả thị trấn chỉ
// còn 50 vật cản. Giờ chọn góc sao cho nhà NẰM TRỌN trong ô, và thử nhiều lần vì ô đã chật
// dần sau mỗi công trình.
func (g *gen) fillBlock(x0, y0, x1, y1 float64, sqA, sqB Vec2) {
	for attempt := 0; attempt < 14; attempt++ {
		switch g.b.rng.Intn(5) {
		case 0, 1:
			c := Vec2{g.b.rnd(x0+300, x1-300), g.b.rnd(y0+250, y1-250)}
			if g.placeable(c, sqA, sqB) {
				g.marketStall(c)
			}
		case 2:
			c := Vec2{g.b.rnd(x0, x1-600), g.b.rnd(y0, y1-120)}
			if g.placeable(c, sqA, sqB) {
				g.fenceRun(biomes[2], c)
			}
		default:
			w, h := 500.0, 420.0
			if x1-x0 < w+40 || y1-y0 < h+40 {
				continue
			}
			c := Vec2{g.b.rnd(x0, x1-w), g.b.rnd(y0, y1-h)}
			// Kiểm cả bốn góc nhà, không chỉ góc trên-trái: nhà đặt sát mép ô vẫn có thể
			// thò một góc ra lòng đường.
			if g.placeable(c, sqA, sqB) && g.placeable(Vec2{c.X + w, c.Y + h}, sqA, sqB) {
				g.house(biomes[2], c)
			}
		}
	}
}

// placeable: điểm này có đặt công trình được không (không trên đường, không sát quảng trường).
func (g *gen) placeable(c, sqA, sqB Vec2) bool {
	return !g.onRoad(c, 200) && c.dist(sqA) >= 1000 && c.dist(sqB) >= 1000
}

// streetLines trả về vị trí các con phố theo một trục: khoảng cách thay đổi 1 500–2 100 để
// không ra lưới đều tăm tắp, và luôn dừng trước mép để không còn dải vụn sát tường biên.
func (g *gen) streetLines(lo, hi float64) []float64 {
	out := []float64{}
	for x := lo + g.b.rnd(1200, 1700); x < hi-1200; x += g.b.rnd(1500, 2100) {
		out = append(out, x)
	}
	return out
}

// marketStall xếp một cụm sạp hàng thành lưới nhỏ, tạo hẻm giữa các sạp.
func (g *gen) marketStall(at Vec2) {
	rows, cols := 2+g.b.rng.Intn(2), 2+g.b.rng.Intn(3)
	const w, h, gap = 260.0, 180.0, 200.0
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			x := at.X + float64(c)*(w+gap) - float64(cols)*(w+gap)/2
			y := at.Y + float64(r)*(h+gap) - float64(rows)*(h+gap)/2
			if !g.onRoad(Vec2{x + w/2, y + h/2}, 260) {
				g.b.obstacle("stall", x, y, w, h, 60)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Khu 5 — Băng Nguyệt Hồ: hồ băng trơn, đảo băng, bờ tuyết
// ---------------------------------------------------------------------------

// bangNguyetHo dựng một hồ băng lớn ở giữa với vài đảo, bờ tuyết bao quanh, rất ít vật cản.
//
// Mặt băng là vùng `ice`: di chuyển có quán tính (spec §8.0, làm ở mục 7). Bờ tuyết là vùng
// `snow` làm chậm ×0.8. Ít vật cản là có chủ ý — cái khó của khu này là **điều khiển**, nếu
// nhồi thêm vật cản thì trượt vào tường suốt và chỉ gây ức chế.
func (g *gen) bangNguyetHo() error {
	s := g.size
	lo, hi := g.b.playable()
	c := Vec2{s / 2, s / 2}
	// Bờ tuyết phủ gần hết khu, hồ băng nằm đè lên giữa.
	g.b.zone("snow", lo+60, lo+60, hi-lo-120, hi-lo-120, false)
	g.b.organicZone("ice", c.X, c.Y, 3900, 0.9)

	islandA := Vec2{c.X - 1700, c.Y - 1100}
	islandB := Vec2{c.X + 1700, c.Y + 1100}
	g.twoAltars(islandA, islandB)
	// Thiên Môn trên đảo băng giữa hồ (spec §8.4a) — phải trượt băng mới tới.
	g.b.organicZone("snow", c.X, c.Y, 700, 0.9)
	g.b.heavenGate(c.X, c.Y)
	g.fourGates("cyan")

	// Đảo băng: vùng tuyết nhỏ giữa hồ (đứng được, không trượt) viền vài tảng băng.
	for _, is := range []Vec2{islandA, islandB, {c.X + 1500, c.Y - 1800}, {c.X - 1500, c.Y + 1800}} {
		g.b.organicZone("snow", is.X, is.Y, 820, 0.85)
		for i := 0; i < 5; i++ {
			a := float64(i) / 5 * 2 * math.Pi
			g.b.boulder(is.X+math.Cos(a)*700, is.Y+math.Sin(a)*700, g.b.rnd(70, 120), 200)
		}
	}
	// Vật cản ngoài hồ: thưa, chỉ để bờ tuyết không trống trơn.
	for _, p := range g.poisson(1400, 400) {
		if p.dist(c) < 4200 {
			continue // trong lòng hồ thì để trống
		}
		g.b.boulder(p.X, p.Y, g.b.rnd(80, 150), 220)
	}
	for _, d := range []string{"rock_small", "pine"} {
		g.b.decor(d, 50)
	}
	return nil
}

// ---------------------------------------------------------------------------

// generateZone chạy bộ sinh của một khu rồi kiểm tra theo cùng một bộ luật.
func generateZone(b *builder, cfg config.Mapgen, zone int) error {
	if zone < 0 || zone >= len(zoneDefs) {
		return fmt.Errorf("khu %d không có (chỉ có 0..%d)", zone, len(zoneDefs)-1)
	}
	g := &gen{b: b, cfg: cfg, size: b.size}
	b.border()
	if err := zoneDefs[zone].build(g); err != nil {
		return err
	}
	return g.check()
}
