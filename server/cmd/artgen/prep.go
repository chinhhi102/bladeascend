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
	feetY       = 460.0 / 512.0 // pivot: feet on y=460 of a 512 canvas (ART_PROMPTS §1)
	charHeight  = 0.72          // character ≈ 70% of canvas height
	centerFill  = 0.9
	minSheetGap = 6
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
		log.Printf("prep %s", j.Name)
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
		sg := Seg{W: j.W, H: j.H, Pivot: j.Pivot}
		return savePNG(out, placeOne(img, box, sg, fitScale(sg, box.Dx(), box.Dy())))
	case modeSheet:
		return prepSheet(img, j, assets)
	default:
		return fmt.Errorf("unknown mode %q", j.Mode)
	}
}

// prepSheet writes each segment's frames under assets/<seg.Out>/.
func prepSheet(img *image.NRGBA, j Job, assets string) error {
	removeBackground(img)
	cells := splitSheet(img, j.Rows)
	boxes := make([]image.Rectangle, 0, len(cells))
	for _, c := range cells {
		b := alphaBBox(img, c)
		if b.Empty() {
			b = c
		}
		boxes = append(boxes, b)
	}
	scales := sharedScales(j, boxes)
	i := 0
	for _, sg := range j.Segs {
		for f := 0; f < sg.Frames && i < len(boxes); f, i = f+1, i+1 {
			s := scales[groupKey(sg)]
			if sg.Each {
				s = fitScale(sg, boxes[i].Dx(), boxes[i].Dy())
			}
			name := fmt.Sprintf("%s_%02d.png", sg.Anim, f)
			if f < len(sg.Names) {
				name = sg.Names[f] + ".png"
			}
			if err := savePNG(filepath.Join(assets, sg.Out, name), placeOne(img, boxes[i], sg, s)); err != nil {
				return err
			}
		}
	}
	return nil
}

func groupKey(sg Seg) string { return fmt.Sprintf("%dx%d/%s", sg.W, sg.H, sg.Pivot) }

// sharedScales gives every non-"each" segment group one scale so a character keeps the same
// size across all its animations.
func sharedScales(j Job, boxes []image.Rectangle) map[string]float64 {
	maxW, maxH := map[string]int{}, map[string]int{}
	i := 0
	for _, sg := range j.Segs {
		k := groupKey(sg)
		for f := 0; f < sg.Frames && i < len(boxes); f, i = f+1, i+1 {
			maxW[k], maxH[k] = max(maxW[k], boxes[i].Dx()), max(maxH[k], boxes[i].Dy())
		}
	}
	out := map[string]float64{}
	for _, sg := range j.Segs {
		k := groupKey(sg)
		out[k] = fitScale(sg, max(1, maxW[k]), max(1, maxH[k]))
	}
	return out
}

func fitScale(sg Seg, w, h int) float64 {
	W, H := float64(sg.W), float64(sg.H)
	if sg.Pivot == "feet" {
		return math.Min(charHeight*H/float64(h), 0.95*W/float64(w))
	}
	return math.Min(centerFill*W/float64(w), centerFill*H/float64(h))
}

// placeOne scales box by s and draws it on a W×H canvas at the segment's pivot.
func placeOne(img *image.NRGBA, b image.Rectangle, sg Seg, s float64) *image.NRGBA {
	w, h := max(1, int(float64(b.Dx())*s)), max(1, int(float64(b.Dy())*s))
	scaled := resize(img, b, w, h)
	canvas := image.NewNRGBA(image.Rect(0, 0, sg.W, sg.H))
	x, y := (sg.W-w)/2, (sg.H-h)/2
	if sg.Pivot == "feet" {
		y = max(0, int(feetY*float64(sg.H))-h)
	}
	draw.Draw(canvas, image.Rect(x, y, x+w, y+h), scaled, image.Point{}, draw.Over)
	return canvas
}

// splitSheet finds frame cells using empty gaps; rows[i] is the frame count of row i.
func splitSheet(img *image.NRGBA, rows []int) []image.Rectangle {
	b := img.Bounds()
	bands := fitSpans(spans(img, b, false, minSheetGap), len(rows), b.Dy())
	var cells []image.Rectangle
	for ri, rs := range bands {
		band := image.Rect(b.Min.X, b.Min.Y+rs[0], b.Max.X, b.Min.Y+rs[1])
		cols := spans(img, band, true, minSheetGap)
		if len(cols) < rows[ri] {
			cols = inkCuts(img, band, rows[ri]) // frames touch: cut at the emptiest columns
		}
		for _, cs := range fitSpans(cols, rows[ri], b.Dx()) {
			cell := image.Rect(b.Min.X+cs[0], band.Min.Y, b.Min.X+cs[1], band.Max.Y)
			dropEdgeFragments(img, cell)
			cells = append(cells, cell)
		}
	}
	return cells
}

// inkCuts splits a band into n cells, moving each even boundary to the column with the least ink nearby.
func inkCuts(img *image.NRGBA, band image.Rectangle, n int) [][2]int {
	w := band.Dx()
	ink := make([]int, w)
	for x := 0; x < w; x++ {
		for y := band.Min.Y; y < band.Max.Y; y++ {
			if img.NRGBAAt(band.Min.X+x, y).A > 40 {
				ink[x]++
			}
		}
	}
	cuts := []int{0}
	win := w / n / 3
	for k := 1; k < n; k++ {
		best := k * w / n
		for x := max(0, best-win); x < min(w, k*w/n+win); x++ {
			if ink[x] < ink[best] {
				best = x
			}
		}
		cuts = append(cuts, best)
	}
	cuts = append(cuts, w)
	out := make([][2]int, n)
	for i := range out {
		out[i] = [2]int{cuts[i], cuts[i+1]}
	}
	return out
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
