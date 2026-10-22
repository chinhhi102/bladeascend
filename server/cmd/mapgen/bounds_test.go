package main

import (
	"fmt"
	"testing"
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
