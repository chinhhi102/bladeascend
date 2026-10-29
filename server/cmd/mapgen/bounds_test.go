package main

import (
	"fmt"
	"testing"

	"kiemxoay/server/internal/config"
)

// Spec §8.1: KHÔNG vật cản / vùng / decor nào được nằm ngoài bản đồ. Mọi object phải nằm
// trọn trong [mapMargin, size-mapMargin], tính cả kích thước của chính nó.
func TestNoObjectOutsideMapOverTenSeeds(t *testing.T) {
	const size = 12000
	gens := map[string]func(*builder){"lang_tre": langTre, "cho_dem": choDem, "nui_da": nuiDa}
	for seed := int64(1); seed <= 10; seed++ {
		for name, gen := range gens {
			t.Run(fmt.Sprintf("%s/seed%d", name, seed), func(t *testing.T) {
				b := newBuilder(size, seed)
				gen(b)
				if err := b.validate(); err != nil {
					t.Fatal(err)
				}
				// Kiểm tra lại bằng tay, không tin mỗi validate().
				n := 0
				for _, ln := range []string{"zones", "obstacles", "gates", "decor"} {
					for _, o := range b.layers[ln] {
						n++
						x1, y1 := o.X+o.Width, o.Y+o.Height
						if o.X < mapMargin || o.Y < mapMargin || x1 > size-mapMargin || y1 > size-mapMargin {
							t.Errorf("%s/%s (%.0f,%.0f)-(%.0f,%.0f) ra ngoài lề %g", ln, o.Type, o.X, o.Y, x1, y1, mapMargin)
						}
					}
				}
				if n == 0 {
					t.Fatal("bản đồ rỗng: test không kiểm được gì")
				}
			})
		}
	}
}

// Tường biên phải khép kín cả 4 cạnh, không có kẽ hở để lọt ra ngoài.
func TestBorderWallEnclosesMap(t *testing.T) {
	const size = 12000
	b := newBuilder(size, 1)
	b.border()
	lo, hi := mapMargin, float64(size)-mapMargin
	var top, bottom, left, right bool
	for _, o := range b.layers["obstacles"] {
		if o.Type != "stone_wall" {
			continue
		}
		switch {
		case o.Y == lo && o.Width >= hi-lo:
			top = true
		case o.Y+o.Height == hi && o.Width >= hi-lo:
			bottom = true
		case o.X == lo && o.Height >= hi-lo:
			left = true
		case o.X+o.Width == hi && o.Height >= hi-lo:
			right = true
		}
	}
	if !top || !bottom || !left || !right {
		t.Fatalf("tường biên hở: trên=%v dưới=%v trái=%v phải=%v", top, bottom, left, right)
	}
	if err := b.validate(); err != nil {
		t.Fatalf("chính tường biên lại nằm ngoài lề: %v", err)
	}
}

// Khu vực đặt ngẫu nhiên phải nằm hẳn trong tường biên, không đè lên tường.
func TestPlayableAreaIsInsideBorder(t *testing.T) {
	b := newBuilder(12000, 1)
	lo, hi := b.playable()
	if lo < mapMargin+borderThick || hi > b.size-mapMargin-borderThick {
		t.Fatalf("vùng chơi [%g, %g] chạm vào tường biên", lo, hi)
	}
}

// Spec §8.5: mỗi bản đồ có đúng 5 Đài Tiên Khí — 1 Thiên Tiên Đài ở giữa và 4 đài nhỏ.
func TestEveryMapHasFiveAltars(t *testing.T) {
	const size = 12000
	for name, gen := range map[string]func(*builder){"lang_tre": langTre, "cho_dem": choDem, "nui_da": nuiDa} {
		b := newBuilder(size, 7)
		gen(b)
		small, center := 0, 0
		for _, o := range b.layers["altars"] {
			switch o.Type {
			case "small":
				small++
			case "center":
				center++
			}
		}
		if center != 1 || small != 4 {
			t.Errorf("%s: %d đài trung tâm + %d đài nhỏ, muốn 1 + 4", name, center, small)
		}
	}
}

// Cả 5 khu đều phải sinh được và tuân thủ cùng một bộ luật (spec §8.0, CHANGES_v5 mục 6):
// đúng 4 cổng, đủ đài, không vật cản ra ngoài bản đồ, mọi chỗ trống đi tới được nhau.
func TestEveryZoneGenerates(t *testing.T) {
	cfg := config.MustLoadDefault()
	for zi, zd := range zoneDefs {
		b, _, err := buildZoneWithRetries(cfg.Room.MapSize, int64(1000+zi), cfg.Mapgen, zi)
		if err != nil {
			t.Fatalf("khu %d (%s) không sinh được: %v", zi, zd.name, err)
		}
		// Đếm riêng ba loại: 4 cổng thường mỗi khu (§8.4), 1 Thiên Môn ở mỗi khu TRỪ Thiên
		// Đình, và 4 Bến Thiên Môn chỉ ở Thiên Đình (§8.4a).
		normal, heaven, dock := 0, 0, 0
		for _, o := range b.layers["gates"] {
			switch {
			case propOf(o, "heaven") == "1":
				heaven++
			case propOf(o, "dock") == "1":
				dock++
			default:
				normal++
			}
		}
		if normal != 4 {
			t.Errorf("khu %d (%s) có %d cổng thường, spec §8.4 nói mỗi khu 4", zi, zd.name, normal)
		}
		wantHeaven, wantDock := 1, 0
		if zi == 0 {
			wantHeaven, wantDock = 0, 4 // Thiên Đình không có Thiên Môn, nhưng có 4 bến
		}
		if heaven != wantHeaven || dock != wantDock {
			t.Errorf("khu %d (%s) có %d Thiên Môn và %d bến, muốn %d và %d",
				zi, zd.name, heaven, dock, wantHeaven, wantDock)
		}
		wantAltars := 2
		if zi == 0 {
			wantAltars = 5 // Thiên Đình: 4 đài nhỏ + Thiên Tiên Đài (§8.0)
		}
		if got := len(b.layers["altars"]); got != wantAltars {
			t.Errorf("khu %d (%s) có %d đài, muốn %d", zi, zd.name, got, wantAltars)
		}
		if len(b.layers["obstacles"]) < 20 {
			t.Errorf("khu %d (%s) chỉ có %d vật cản — trống trơn", zi, zd.name, len(b.layers["obstacles"]))
		}
	}
}

// Mê cung tre phải LIÊN THÔNG theo đúng cách nó được dựng: thuật toán đào đường lùi dần
// thăm mọi ô, nên không ô nào bị nhốt. Khoá lại vì đây là điểm cốt tử của khu đó.
func TestCarveMazeReachesEveryCell(t *testing.T) {
	const n = 9
	walls := carveMaze(n, func(k int) int { return k / 2 })
	blocked := map[[3]int]bool{}
	for _, w := range walls {
		blocked[[3]int{w.cx, w.cy, boolInt(w.vertical)}] = true
	}
	seen := map[[2]int]bool{{0, 0}: true}
	stack := [][2]int{{0, 0}}
	for len(stack) > 0 {
		c := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := c[0]+d[0], c[1]+d[1]
			if nx < 0 || ny < 0 || nx >= n || ny >= n || seen[[2]int{nx, ny}] {
				continue
			}
			wall := [3]int{max(nx, c[0]), c[1], 1}
			if d[0] == 0 {
				wall = [3]int{c[0], max(ny, c[1]), 0}
			}
			if blocked[wall] {
				continue
			}
			seen[[2]int{nx, ny}] = true
			stack = append(stack, [2]int{nx, ny})
		}
	}
	if len(seen) != n*n {
		t.Fatalf("mê cung chỉ đi tới được %d/%d ô — có ngõ bị nhốt", len(seen), n*n)
	}
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// propOf đọc một thuộc tính chuỗi của object trong bản đồ.
func propOf(o object, name string) string {
	for _, p := range o.Properties {
		if p.Name == name {
			if v, ok := p.Value.(string); ok {
				return v
			}
		}
	}
	return ""
}
