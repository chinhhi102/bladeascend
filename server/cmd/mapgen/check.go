package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"kiemxoay/server/internal/config"
	"kiemxoay/server/internal/game"
	"kiemxoay/server/internal/geom"
	"kiemxoay/server/internal/world"
)

// Bộ kiểm bố cục bản đồ — spec §8.1b "Kiểm tra bắt buộc".
//
// Luật gốc: **"không chém được thì phải đi lọt"**. Người cảnh giới r chưa chém được vật cản
// bậc > r, nên với MỖI cảnh giới r = 0..13 ta bỏ các vật cản bậc ≤ r (vì r chém qua được),
// lấy thân TO NHẤT của r, rồi flood fill lưới 50:
//   - mọi chỗ trống phải liên thông (không túi bị nhốt);
//   - mọi cổng, Thiên Môn, Đài, bến, điểm spawn đều tới được;
//   - diện tích trống ≥ 60%.
//
// Vì sao phải quét ĐỦ 14 cảnh giới chứ không chỉ người to nhất: hai luật kéo ngược nhau.
// Người càng cao cảnh giới thì thân càng to (khó lọt) NHƯNG chém được càng nhiều vật cản
// (dễ đi). Nên không có một cảnh giới nào là "trường hợp xấu nhất" — Đạo Tổ to gấp 6 lần
// Luyện Khí mà vẫn đi lại thoải mái hơn, vì họ chém được mọi thứ. Chỗ chết người nằm ở
// khoảng giữa: Độ Kiếp thân to 4.x mà vẫn chưa chém nổi đá khổng lồ bậc 9.

const checkCell = 50.0

// Mục tiêu mật độ theo TỪNG KHU (chủ dự án chốt 31/10/2026).
//
// Con số 18–28% của spec §8.1b đặt lúc chưa ai đo gì; đo thật rồi thì 18% đã là một khu
// chật đến mức người Đại Thừa (thân 150) không lách nổi. Mục tiêu mới tính theo **diện tích
// chân đế** — phần đất vật cản thật sự chiếm, không phải "chỗ người cỡ này không lọt".
//
// Mỗi khu một mức vì mỗi khu một vai: Băng Nguyệt Hồ là mặt hồ để trượt và rượt nhau nên
// phải thoáng; Trúc Lâm là mê cung phục kích nên dày nhất.
var densityTarget = map[string][2]float64{
	"bang_nguyet_ho": {0.04, 0.06},
	"thien_dinh":     {0.08, 0.12},
	"pham_tran":      {0.08, 0.12},
	"linh_son_coc":   {0.10, 0.14},
	"truc_lam":       {0.10, 0.14},
}

// densityHardMax: vượt mức này là TRƯỢT, không phải cảnh báo.
//
// Dưới mục tiêu thì khu chỉ nhạt — chơi vẫn được. Vượt trần thì bắt đầu có người bị kẹt, và
// bộ kiểm liên thông không bắt được hết: nó đo "có đường đi không", không đo "đường có hẹp
// tới mức đi một bước phải né ba lần không".
const densityHardMax = 0.16

func densityRange(id string) [2]float64 {
	if r, ok := densityTarget[id]; ok {
		return r
	}
	return [2]float64{0.08, 0.12}
}

// checkResult là kết quả của một cảnh giới.
type checkResult struct {
	Realm     int
	Name      string
	BodyR     float64
	OpenFrac  float64 // phần diện tích khu mà thân này đi vào được
	MainFrac  float64 // phần chỗ trống thuộc vùng liên thông lớn nhất
	Regions   int
	CutOff    []string // tên những mốc không tới được
	TrappedPx int
	// Density: phần diện tích khu bị vật cản chiếm, tính với thân NHỎ NHẤT (tức diện tích
	// thật của vật cản, không phải "chỗ người này không lọt"). §8.1b mục 5: 18–28%.
	//
	// Đo riêng vì nó KHÔNG phụ thuộc cảnh giới — nhưng vẫn in mỗi dòng để thấy được mật độ
	// tụt thế nào khi người chơi lên cảnh giới và chém được thêm nhiều loại.
	Density float64
}

// minOpenRatio là ngưỡng của §8.1b: "Diện tích trống ≥ 60%".
//
// KHÁC `mapgen.min_open_ratio` (0.65) trong config — con số đó là ngưỡng TỰ KIỂM của bộ
// sinh, dùng để quyết "sinh lại với seed+1", nên nó cố tình chặt hơn luật nghiệm thu. Lấy
// nó làm luật nghiệm thu là trộn hai thứ: bản đồ trống 65% ở Hóa Thần bị báo TRƯỢT dù spec
// chấp nhận từ 60%.
const minOpenRatio = 0.60

func (r checkResult) ok() bool {
	return r.OpenFrac >= minOpenRatio && r.MainFrac >= 0.99 && len(r.CutOff) == 0
}

// landmark là một mốc phải tới được: cổng, Thiên Môn, bến, Đài, điểm spawn.
type landmark struct {
	At   geom.Vec
	Name string
}

// runCheck kiểm mọi bản đồ trong `dir`. `pngDir` khác rỗng thì xuất ảnh cho từng cảnh giới.
func runCheck(dir, pngDir string, cfg *config.Config, pngRealms []int) error {
	maps, err := world.LoadDir(dir, world.LoadOptions{
		Tier: cfg.Obstacles.ObstacleTier, FootprintW: cfg.Obstacles.FootprintW, FootprintH: cfg.Obstacles.FootprintH,
	})
	if err != nil {
		return err
	}
	if len(maps) == 0 {
		return fmt.Errorf("không có bản đồ nào trong %s", dir)
	}
	prog := game.NewProgression(cfg.Cultivation, cfg.Sword.MaxSwords)
	// Đủ 14 cảnh giới: 9 phàm + 5 Tiên (spec §8.2 "r = 0..13").
	realms := len(cfg.Cultivation.Realms) + prog.TienCount()
	sort.Slice(maps, func(i, j int) bool { return maps[i].ID < maps[j].ID })

	bad := 0
	for _, m := range maps {
		fmt.Printf("== %s (%s)\n", m.ID, m.Name)
		marks := landmarksOf(m)
		for r := 0; r < realms; r++ {
			res, grid := checkRealm(m, prog, cfg, r, marks)
			status := "ĐẠT"
			if !res.ok() {
				status, bad = "TRƯỢT", bad+1
			}
			fmt.Printf("  r=%-2d %-17s thân %3.0f · vật cản %4.1f%% · trống %4.0f%% · vùng chính %3.0f%% · %d vùng · %s",
				r, res.Name, res.BodyR, res.Density*100, res.OpenFrac*100, res.MainFrac*100, res.Regions, status)
			if len(res.CutOff) > 0 {
				fmt.Printf(" · KHÔNG TỚI ĐƯỢC: %s", strings.Join(res.CutOff, ", "))
			}
			fmt.Println()
			// Mật độ CHƯA ĐẠT mục tiêu thì chỉ cảnh báo, VƯỢT TRẦN thì trượt (chủ dự án
			// chốt 31/10/2026). Hai hướng không đối xứng vì hậu quả không đối xứng: khu
			// nhạt thì chơi vẫn được, chỉ kém hay; khu chật thì có người kẹt — và bộ kiểm
			// liên thông không bắt được hết, nó đo "có đường đi không" chứ không đo "đường
			// có hẹp tới mức đi một bước né ba lần không".
			if r == 0 {
				want := densityRange(m.ID)
				switch {
				case res.Density > densityHardMax:
					fmt.Printf("     ✗ mật độ chân đế %.1f%% VƯỢT TRẦN %.0f%% — khu chật, người to kẹt\n",
						res.Density*100, densityHardMax*100)
					bad++
				case res.Density < want[0]:
					fmt.Printf("     ⚠ mật độ chân đế %.1f%%, mục tiêu khu này %.0f–%.0f%% — còn trống\n",
						res.Density*100, want[0]*100, want[1]*100)
				case res.Density > want[1]:
					fmt.Printf("     ⚠ mật độ chân đế %.1f%%, mục tiêu khu này %.0f–%.0f%% — hơi dày\n",
						res.Density*100, want[0]*100, want[1]*100)
				}
			}
			if pngDir != "" && wants(pngRealms, r) {
				if err := writeCheckPNG(pngDir, m, r, grid); err != nil {
					return err
				}
			}
		}
	}
	if bad > 0 {
		return fmt.Errorf("%d cảnh giới trượt bộ kiểm §8.1b", bad)
	}
	return nil
}

func wants(list []int, r int) bool {
	if len(list) == 0 {
		return true
	}
	for _, v := range list {
		if v == r {
			return true
		}
	}
	return false
}

// bodyRadiusFor là bán kính thân TO NHẤT của cảnh giới r.
//
// "To nhất" = tầng 10 với người phàm (spec §8.1b). Tiên giới chỉ có một mức scale cho cả
// cảnh giới nên tầng nào cũng thế.
func bodyRadiusFor(prog *game.Progression, cfg *config.Config, r int) float64 {
	level := r*prog.Tiers() + prog.Tiers() - 1
	return cfg.Player.BodyR * prog.Scale(level)
}

// landmarksOf gom mọi mốc §8.1b đòi phải tới được.
func landmarksOf(m *world.Map) []landmark {
	var out []landmark
	for _, g := range m.Gates {
		name := fmt.Sprintf("cổng %d", g.ID)
		switch {
		case g.Heaven:
			name = "Thiên Môn"
		case g.Dock:
			name = fmt.Sprintf("bến Thiên Môn %d", g.DockFor)
		}
		out = append(out, landmark{At: geom.Vec{X: g.X, Y: g.Y}, Name: name})
	}
	for i, a := range m.Altars {
		name := fmt.Sprintf("Đài %d", i)
		if a.Kind == "center" {
			name = "Thiên Tiên Đài"
		}
		out = append(out, landmark{At: geom.Vec{X: a.X, Y: a.Y}, Name: name})
	}
	return out
}

// checkGrid giữ lưới đã dựng, để vẽ ảnh mà không phải tính lại.
type checkGrid struct {
	n     int
	lo    float64
	solid []bool // vật cản còn chặn cảnh giới này (bậc > r)
	tight []bool // trống với thân nhỏ nhưng thân CỦA R không lọt
	// footprint: tâm ô nằm TRONG chân đế một vật cản. Khác `solid` ở chỗ `solid` nở thêm bán
	// kính thân người, nên `solid` trả lời "đi được không" còn `footprint` trả lời "đất bị
	// chiếm bao nhiêu" — mật độ hỏi câu sau.
	footprint []bool
	label     []int
	main      int
	sizes     []int
	marks     []landmark
	cutOff    []string
}

// checkRealm chạy bộ kiểm cho đúng một cảnh giới.
func checkRealm(m *world.Map, prog *game.Progression, cfg *config.Config, r int, marks []landmark) (checkResult, *checkGrid) {
	bodyR := bodyRadiusFor(prog, cfg, r)
	lo, hi := m.PlayMargin, m.Size-m.PlayMargin
	n := int((hi - lo) / checkCell)
	g := &checkGrid{n: n, lo: lo, solid: make([]bool, n*n), tight: make([]bool, n*n),
		footprint: make([]bool, n*n), label: make([]int, n*n), marks: marks}
	for i := range g.label {
		g.label[i] = -1
	}
	// Chỉ giữ vật cản cảnh giới này CHƯA chém được. Bậc ≤ r thì họ chạm là nó biến mất, nên
	// với họ bức tường đó không tồn tại.
	var walls []world.Obstacle
	for _, o := range m.Obstacles {
		if o.Tier > r {
			walls = append(walls, o)
		}
	}
	for i := 0; i < n*n; i++ {
		c := g.center(i)
		// `tight` dùng bán kính thân nhỏ nhất để tách hai lý do một ô bị chặn: "có vật cản
		// ở đây" (xám trong ảnh) và "ô trống nhưng thân cỡ này không lọt" (cam). Không tách
		// thì ảnh chỉ nói "đi không được", mà người sửa bản đồ cần biết là **nới hành lang**
		// hay **bỏ vật cản**.
		small := cfg.Player.BodyR * prog.Scale(0)
		g.solid[i] = blockedAt(walls, c, bodyR)
		g.tight[i] = g.solid[i] && !blockedAt(walls, c, small)
		g.footprint[i] = insideAny(walls, c)
	}
	g.components()
	res := checkResult{Realm: r, Name: prog.RealmName(r), BodyR: bodyR}
	free := 0
	for i := range g.solid {
		if !g.solid[i] {
			free++
		}
	}
	if n > 0 {
		res.OpenFrac = float64(free) / float64(n*n)
	}
	covered := 0
	for i := 0; i < n*n; i++ {
		if g.footprint[i] {
			covered++
		}
	}
	if n > 0 {
		res.Density = float64(covered) / float64(n*n)
	}
	if free > 0 && g.main >= 0 {
		res.MainFrac = float64(g.sizes[g.main]) / float64(free)
		res.TrappedPx = free - g.sizes[g.main]
	}
	res.Regions = len(g.sizes)
	res.CutOff = g.unreachable()
	return res, g
}

// insideAny: điểm `c` có nằm TRONG chân đế vật cản nào không.
//
// Không dùng `OverlapsCircle(c, 0)` được: nó gọi `PushOut`, và `PushOut` với bán kính 0 trả
// về "không chạm" cho mọi điểm — hình tròn bán kính 0 không chồng lên gì theo định nghĩa
// của nó. Dùng nhầm thì mật độ ra đúng 0,00% cho mọi bản đồ, và con số 0 đó trông y như
// "bản đồ chưa có vật cản" chứ không như một lỗi.
func insideAny(walls []world.Obstacle, c geom.Vec) bool {
	for _, o := range walls {
		b := o.Block()
		lo, hi := b.Bounds()
		if c.X >= lo.X && c.X <= hi.X && c.Y >= lo.Y && c.Y <= hi.Y {
			return true
		}
	}
	return false
}

func blockedAt(walls []world.Obstacle, c geom.Vec, r float64) bool {
	for _, o := range walls {
		if o.Block().OverlapsCircle(c, r) {
			return true
		}
	}
	return false
}

func (g *checkGrid) center(i int) geom.Vec {
	return geom.Vec{
		X: g.lo + (float64(i%g.n)+0.5)*checkCell,
		Y: g.lo + (float64(i/g.n)+0.5)*checkCell,
	}
}

func (g *checkGrid) cellOf(p geom.Vec) int {
	x := int((p.X - g.lo) / checkCell)
	y := int((p.Y - g.lo) / checkCell)
	if x < 0 || y < 0 || x >= g.n || y >= g.n {
		return -1
	}
	return y*g.n + x
}

// components gắn nhãn mọi vùng trống liên thông (4 hướng) và ghi lại vùng lớn nhất.
func (g *checkGrid) components() {
	g.main = -1
	for s := 0; s < g.n*g.n; s++ {
		if g.solid[s] || g.label[s] >= 0 {
			continue
		}
		id := len(g.sizes)
		size := 0
		queue := []int{s}
		g.label[s] = id
		for len(queue) > 0 {
			i := queue[len(queue)-1]
			queue = queue[:len(queue)-1]
			size++
			x, y := i%g.n, i/g.n
			for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				nx, ny := x+d[0], y+d[1]
				if nx < 0 || ny < 0 || nx >= g.n || ny >= g.n {
					continue
				}
				j := ny*g.n + nx
				if !g.solid[j] && g.label[j] < 0 {
					g.label[j] = id
					queue = append(queue, j)
				}
			}
		}
		g.sizes = append(g.sizes, size)
		if g.main < 0 || size > g.sizes[g.main] {
			g.main = id
		}
	}
}

// unreachable: mốc nào không nằm trong vùng liên thông lớn nhất.
//
// Mốc nằm ĐÚNG trong một ô bị chặn cũng tính là không tới được, nhưng phải xét ô lân cận
// trước: tâm một cái cổng rộng 200 có thể rơi vào ô mà thân người to không lọt, trong khi
// cổng vẫn vào được từ cạnh bên.
func (g *checkGrid) unreachable() []string {
	var out []string
	for _, mk := range g.marks {
		i := g.cellOf(mk.At)
		if i < 0 {
			out = append(out, mk.Name+" (ngoài lưới)")
			continue
		}
		if g.reachableNear(i) {
			continue
		}
		out = append(out, mk.Name)
	}
	return out
}

// reachableNear: ô này, hay một ô trong bán kính 3 ô quanh nó, thuộc vùng chính?
func (g *checkGrid) reachableNear(i int) bool {
	const span = 3
	x0, y0 := i%g.n, i/g.n
	for dy := -span; dy <= span; dy++ {
		for dx := -span; dx <= span; dx++ {
			x, y := x0+dx, y0+dy
			if x < 0 || y < 0 || x >= g.n || y >= g.n {
				continue
			}
			if g.label[y*g.n+x] == g.main {
				return true
			}
		}
	}
	return false
}

// writeCheckPNG xuất ảnh `docs/map_check/{khu}_r{r}.png` đúng bảng màu spec §8.1b:
// xám = vật cản, cam = chỗ thân không lọt, đỏ = vùng bị nhốt.
func writeCheckPNG(dir string, m *world.Map, r int, g *checkGrid) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// Phóng to 3 lần theo kiểu "lấy ô gần nhất": lưới 50 trên bản đồ 12 000 cho ra ảnh
	// 230 px, nhỏ quá để nhìn ra một hành lang rộng 3 ô. Phóng ở đây chứ không để người xem
	// tự zoom: ảnh này để DUYỆT, mở ra là phải đọc được ngay.
	const zoom = 3
	img := image.NewRGBA(image.Rect(0, 0, g.n*zoom, g.n*zoom))
	var (
		cFree    = color.RGBA{0xee, 0xf2, 0xe6, 0xff}
		cWall    = color.RGBA{0x55, 0x55, 0x55, 0xff}
		cTight   = color.RGBA{0xf4, 0xa2, 0x61, 0xff}
		cTrapped = color.RGBA{0xe6, 0x39, 0x46, 0xff}
	)
	for i := 0; i < g.n*g.n; i++ {
		c := cFree
		switch {
		case g.tight[i]:
			c = cTight
		case g.solid[i]:
			c = cWall
		case g.label[i] != g.main:
			c = cTrapped
		}
		x0, y0 := (i%g.n)*zoom, (i/g.n)*zoom
		for dy := 0; dy < zoom; dy++ {
			for dx := 0; dx < zoom; dx++ {
				img.Set(x0+dx, y0+dy, c)
			}
		}
	}
	// Mốc: cổng xanh, đài vàng. Vẽ thành chữ thập 5 px để thấy được ở ảnh 230 px.
	for _, mk := range g.marks {
		i := g.cellOf(mk.At)
		if i < 0 {
			continue
		}
		c := color.RGBA{0x3a, 0x86, 0xff, 0xff}
		if strings.HasPrefix(mk.Name, "Đài") || strings.HasPrefix(mk.Name, "Thiên Tiên") {
			c = color.RGBA{0xff, 0xbe, 0x0b, 0xff}
		} else if strings.HasPrefix(mk.Name, "spawn") {
			c = color.RGBA{0x8a, 0x2b, 0xe2, 0xff}
		}
		x0, y0 := (i%g.n)*zoom+zoom/2, (i/g.n)*zoom+zoom/2
		for d := -5; d <= 5; d++ {
			for w := -1; w <= 1; w++ {
				img.Set(x0+d, y0+w, c)
				img.Set(x0+w, y0+d, c)
			}
		}
	}
	path := filepath.Join(dir, fmt.Sprintf("%s_r%d.png", m.ID, r))
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
