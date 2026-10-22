package main

import (
	"image"
	"image/color"
	"testing"
)

// A red disc with an anti-aliased edge on magenta must come out with a transparent
// background and no pink fringe.
func TestRemoveMagenta(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 40, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 40; x++ {
			img.SetNRGBA(x, y, color.NRGBA{255, 0, 255, 255})
		}
	}
	for y := 10; y < 30; y++ {
		for x := 10; x < 30; x++ {
			img.SetNRGBA(x, y, color.NRGBA{200, 30, 20, 255})
		}
	}
	// 50% blend of red over magenta at the edge.
	img.SetNRGBA(9, 20, color.NRGBA{227, 15, 137, 255})
	// Purple suit colour from a palette must survive.
	img.SetNRGBA(20, 20, color.NRGBA{114, 9, 183, 255})
	removeBackground(img)
	if a := img.NRGBAAt(1, 1).A; a != 0 {
		t.Fatalf("background alpha %d", a)
	}
	if c := img.NRGBAAt(15, 15); c.A != 255 || c.R != 200 {
		t.Fatalf("subject changed: %+v", c)
	}
	if c := img.NRGBAAt(20, 20); c.A != 255 {
		t.Fatalf("purple suit keyed out: %+v", c)
	}
	edge := img.NRGBAAt(9, 20)
	if edge.A == 0 || edge.A == 255 || edge.B > 80 {
		t.Fatalf("edge should be semi-transparent red without magenta: %+v", edge)
	}
}

func TestRowLayoutSplitsSharedRows(t *testing.T) {
	j := Job{Rows: []int{6, 6, 8}, Segs: []Seg{{Frames: 6}, {Frames: 6}, {Frames: 4}, {Frames: 4}}}
	l := rowLayout(j)
	if len(l[2]) != 2 || l[2][0].seg != 2 || l[2][1].seg != 3 || l[0][0].count != 6 {
		t.Fatalf("layout %+v", l)
	}
}

// A 2x2 sheet where each bottom frame is a ghost floating above a body: the grid cut must keep
// ghost and body together and put the wisp under the top-right frame in the bottom row.
func TestGridFramesKeepsStackedParts(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 400, 400))
	fill := func(x0, y0, x1, y1 int) {
		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				img.SetNRGBA(x, y, color.NRGBA{A: 255})
			}
		}
	}
	fill(40, 20, 160, 180)   // top-left body
	fill(240, 140, 360, 180) // top-right body lying flat
	fill(60, 220, 140, 290)  // bottom-left ghost
	fill(40, 310, 160, 380)  // bottom-left body
	fill(260, 200, 270, 215) // wisp just below the flat body, belongs to bottom-right
	fill(240, 310, 360, 380) // bottom-right body
	rows := gridFrames(img, 2, 2)
	if len(rows) != 2 || len(rows[0]) != 2 || len(rows[1]) != 2 {
		t.Fatalf("want 2x2 frames, got %v rows", rows)
	}
	if h := rows[1][0].Bounds().Dy(); h < 150 {
		t.Errorf("ghost and body split: bottom-left frame height %d", h)
	}
	if h := rows[0][1].Bounds().Dy(); h > 45 {
		t.Errorf("wisp glued to the frame above: top-right frame height %d", h)
	}
}
