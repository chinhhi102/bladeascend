package main

import (
	"fmt"
	"image"
	"image/draw"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
)

const (
	feetY        = 460.0 / 512.0 // pivot: feet on y=460 of a 512 canvas (ART_PROMPTS §1)
	charHeight   = 0.72          // character ≈ 70% of canvas height
	centerFill   = 0.9
	minSheetGap  = 6
)

func prepAll(jobs []Job, assets string) {
	for _, j := range jobs {
		raw := rawPath(assets, j)
		if _, err := os.Stat(raw); err != nil {
			continue
		}
		if err := prepJob(j, raw, assets); err != nil {
			log.Printf("prep %s: %v", j.Name, err)
			continue
		}
		log.Printf("prep %s → %s", j.Name, j.Out)
	}
	if err := buildAtlases(assets); err != nil {
		log.Printf("atlas: %v", err)
	}
	if err := writeManifest(assets); err != nil {
		log.Printf("manifest: %v", err)
	}
}

func prepJob(j Job, raw, assets string) error {
	img, err := loadImage(raw)
	if err != nil {
		return err
	}
	out := filepath.Join(assets, j.Out)
	switch j.Mode {
	case modeRaw:
		b := img.Bounds()
		return savePNG(out, resize(img, b, 1024, 1024*b.Dy()/b.Dx()))
	case modeTile:
		return savePNG(out, resize(img, centerSquare(img.Bounds()), j.W, j.H))
	case modeWide:
		return savePNG(out, resize(img, coverRect(img.Bounds(), j.W, j.H), j.W, j.H))
	case modeSprite:
		removeBackground(img)
		box := alphaBBox(img, img.Bounds())
		if box.Empty() {
			return fmt.Errorf("empty image after background removal")
		}
		frames := place(img, []image.Rectangle{box}, j)
		return savePNG(out, frames[0])
	case modeSheet:
		return prepSheet(img, j, out)
	default:
		return fmt.Errorf("unknown mode %q", j.Mode)
	}
}

func prepSheet(img *image.NRGBA, j Job, outDir string) error {
	removeBackground(img)
	cells := splitSheet(img, j)
	boxes := make([]image.Rectangle, 0, len(cells))
	for _, c := range cells {
		b := alphaBBox(img, c)
		if b.Empty() {
			b = c
		}
		boxes = append(boxes, b)
	}
	for i, f := range place(img, boxes, j) {
		name := fmt.Sprintf("%s_%02d.png", j.Anim, i)
		if i < len(j.Names) {
			name = j.Names[i] + ".png"
		}
		if err := savePNG(filepath.Join(outDir, name), f); err != nil {
			return err
		}
	}
	return nil
}

// splitSheet finds frame cells using empty gaps; falls back to an even grid.
func splitSheet(img *image.NRGBA, j Job) []image.Rectangle {
	b := img.Bounds()
	rows := fitSpans(spans(img, b, false, minSheetGap), j.Rows, b.Dy())
	var cells []image.Rectangle
	for ri, rs := range rows {
		band := image.Rect(b.Min.X, b.Min.Y+rs[0], b.Max.X, b.Min.Y+rs[1])
		want := j.Cols
		if ri == len(rows)-1 {
			want = j.Frames - j.Cols*(j.Rows-1)
		}
		for _, cs := range fitSpans(spans(img, band, true, minSheetGap), want, b.Dx()) {
			cells = append(cells, image.Rect(b.Min.X+cs[0], band.Min.Y, b.Min.X+cs[1], band.Max.Y))
		}
	}
	return cells
}

// fitSpans merges the closest spans until exactly want remain; too few → even split.
func fitSpans(sp [][2]int, want, total int) [][2]int {
	for len(sp) > want {
		best, bestGap := 0, math.MaxInt
		for i := 0; i+1 < len(sp); i++ {
			// Prefer merging small fragments (sweat drops, sparks) into neighbours.
			gap := sp[i+1][0] - sp[i][1] + min(sp[i][1]-sp[i][0], sp[i+1][1]-sp[i+1][0])/4
			if gap < bestGap {
				best, bestGap = i, gap
			}
		}
		sp[best] = [2]int{sp[best][0], sp[best+1][1]}
		sp = append(sp[:best+1], sp[best+2:]...)
	}
	if len(sp) == want {
		return sp
	}
	even := make([][2]int, want)
	for i := range even {
		even[i] = [2]int{i * total / want, (i + 1) * total / want}
	}
	return even
}

// place scales every box by one shared factor and draws it at the job's pivot.
func place(img *image.NRGBA, boxes []image.Rectangle, j Job) []*image.NRGBA {
	maxW, maxH := 1, 1
	for _, b := range boxes {
		maxW, maxH = max(maxW, b.Dx()), max(maxH, b.Dy())
	}
	W, H := float64(j.W), float64(j.H)
	s := math.Min(centerFill*W/float64(maxW), centerFill*H/float64(maxH))
	if j.Pivot == "feet" {
		s = math.Min(charHeight*H/float64(maxH), 0.95*W/float64(maxW))
	}
	out := make([]*image.NRGBA, 0, len(boxes))
	for _, b := range boxes {
		w, h := max(1, int(float64(b.Dx())*s)), max(1, int(float64(b.Dy())*s))
		scaled := resize(img, b, w, h)
		canvas := image.NewNRGBA(image.Rect(0, 0, j.W, j.H))
		x := (j.W - w) / 2
		y := (j.H - h) / 2
		if j.Pivot == "feet" {
			y = max(0, int(feetY*H)-h)
		}
		draw.Draw(canvas, image.Rect(x, y, x+w, y+h), scaled, image.Point{}, draw.Over)
		out = append(out, canvas)
	}
	return out
}

func centerSquare(b image.Rectangle) image.Rectangle {
	s := min(b.Dx(), b.Dy())
	x, y := b.Min.X+(b.Dx()-s)/2, b.Min.Y+(b.Dy()-s)/2
	return image.Rect(x, y, x+s, y+s)
}

// coverRect crops b to the aspect ratio w:h around its centre.
func coverRect(b image.Rectangle, w, h int) image.Rectangle {
	cw, ch := b.Dx(), b.Dx()*h/w
	if ch > b.Dy() {
		ch, cw = b.Dy(), b.Dy()*w/h
	}
	x, y := b.Min.X+(b.Dx()-cw)/2, b.Min.Y+(b.Dy()-ch)/2
	return image.Rect(x, y, x+cw, y+ch)
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
