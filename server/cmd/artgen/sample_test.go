package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestPrepSample slices a real sheet: ARTGEN_SAMPLE=path/to/raw.png ARTGEN_OUT=dir go test -run Sample
func TestPrepSample(t *testing.T) {
	src := os.Getenv("ARTGEN_SAMPLE")
	if src == "" {
		t.Skip("set ARTGEN_SAMPLE to a raw sheet")
	}
	out := os.Getenv("ARTGEN_OUT")
	img, err := loadImage(src)
	if err != nil {
		t.Fatal(err)
	}
	j := Job{Rows: []int{4, 8}, Segs: []Seg{
		{Anim: "idle", Frames: 4, Out: filepath.Base(src), W: 512, H: 512, Pivot: "feet"},
		{Anim: "run", Frames: 8, Out: filepath.Base(src), W: 512, H: 512, Pivot: "feet"},
	}}
	if err := prepSheet(img, j, out); err != nil {
		t.Fatal(err)
	}
}
