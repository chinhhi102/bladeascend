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

// rowPart is one segment's share of a sheet row.
type rowPart struct {
	seg   int
	count int
}

// rowLayout maps each row to the segments drawn in it, following the expected frame order.
func rowLayout(j Job) [][]rowPart {
	out := make([][]rowPart, len(j.Rows))
	si, left := 0, 0
	if len(j.Segs) > 0 {
		left = j.Segs[0].Frames
	}
	for r, n := range j.Rows {
		for n > 0 && si < len(j.Segs) {
			take := min(n, left)
			out[r] = append(out[r], rowPart{si, take})
			n, left = n-take, left-take
			if left == 0 && si+1 < len(j.Segs) {
				si++
				left = j.Segs[si].Frames
			} else if left == 0 {
				si++
			}
		}
	}
	return out
}

// prepSheet writes each segment's frames under assets/<seg.Out>/. Rows are found by empty
// horizontal gaps; frames inside a row by connected pixel regions (components.go). A row owned
// by a single animation keeps however many frames the model actually drew.
func prepSheet(img *image.NRGBA, j Job, assets string) error {
	removeBackground(img)
	layout := rowLayout(j)
	frames := make([][]*image.NRGBA, len(j.Segs))
	for r, row := range sheetRows(img, j.Rows) {
		fi := 0
		for _, part := range layout[r] {
			n := part.count
			if len(layout[r]) == 1 {
				n = len(row)
			}
			for k := 0; k < n && fi < len(row); k, fi = k+1, fi+1 {
				frames[part.seg] = append(frames[part.seg], row[fi])
			}
		}
	}
	scales := sharedScales(j, frames)
	for si, sg := range j.Segs {
		clearStaleFrames(filepath.Join(assets, sg.Out), sg)
		for f, fr := range frames[si] {
			b := fr.Bounds()
			s := scales[groupKey(sg)]
			if sg.Each {
				s = fitScale(sg, b.Dx(), b.Dy())
			}
			name := fmt.Sprintf("%s_%02d.png", sg.Anim, f)
			if f < len(sg.Names) {
				name = sg.Names[f] + ".png"
			}
			if err := savePNG(filepath.Join(assets, sg.Out, name), placeOne(fr, b, sg, s)); err != nil {
				return err
			}
		}
	}
	return nil
}

// sheetRows splits the sheet into row bands and each band into isolated frame images.
func sheetRows(img *image.NRGBA, rows []int) [][]*image.NRGBA {
	b := img.Bounds()
	raw := spans(img, b, false, minSheetGap)
	bands := fitSpans(raw, len(rows), b.Dy())
	debugf("row spans %v → %v", raw, bands)
	rects := make([]image.Rectangle, len(bands))
	for i, rs := range bands {
		rects[i] = image.Rect(b.Min.X, b.Min.Y+rs[0], b.Max.X, b.Min.Y+rs[1])
	}
	single := singleFrameWidth(groupWidths(img, rects, rows))
	out := make([][]*image.NRGBA, len(rects))
	for i, band := range rects {
		out[i] = rowFrames(img, band, rowWant{n: rows[i], single: single})
	}
	return out
}

// clearStaleFrames removes old {anim}_NN.png files so a re-generated animation with fewer
// frames does not keep leftovers from the previous image.
func clearStaleFrames(dir string, sg Seg) {
	if len(sg.Names) > 0 {
		return
	}
	old, _ := filepath.Glob(filepath.Join(dir, sg.Anim+"_[0-9][0-9].png"))
	for _, f := range old {
		_ = os.Remove(f)
	}
}

func groupKey(sg Seg) string { return fmt.Sprintf("%dx%d/%s", sg.W, sg.H, sg.Pivot) }

// sharedScales gives every non-"each" segment group one scale so a character keeps its
// original proportions across all frames and animations (no per-frame resizing).
func sharedScales(j Job, frames [][]*image.NRGBA) map[string]float64 {
	maxW, maxH := map[string]int{}, map[string]int{}
	for si, sg := range j.Segs {
		k := groupKey(sg)
		for _, f := range frames[si] {
			maxW[k], maxH[k] = max(maxW[k], f.Bounds().Dx()), max(maxH[k], f.Bounds().Dy())
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
		k := sg.HeightK
		if k == 0 {
			k = 1
		}
		return math.Min(charHeight*k*H/float64(h), 0.95*W/float64(w))
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

func debugf(format string, args ...any) {
	if os.Getenv("ARTGEN_DEBUG") != "" {
		log.Printf(format, args...)
	}
}

// singleFrameWidth estimates one frame's width from spans that are not two frames touching.
func singleFrameWidth(rows [][][2]int) float64 {
	var widths []int
	for _, sp := range rows {
		if !hasWideSpan(sp) {
			for _, s := range sp {
				widths = append(widths, s[1]-s[0])
			}
		}
	}
	if len(widths) == 0 {
		for _, sp := range rows {
			for _, s := range sp {
				widths = append(widths, s[1]-s[0])
			}
		}
	}
	if len(widths) == 0 {
		return 0
	}
	sort.Ints(widths)
	return float64(widths[len(widths)/2])
}

type rowWant struct {
	n      int
	single float64 // estimated width of one frame
}

// framesPerSpan estimates how many touching frames each span holds. When the estimate is
// within one of the requested count, the requested count wins (spread by span width).
func framesPerSpan(cols [][2]int, w rowWant) []int {
	counts := make([]int, len(cols))
	est, total := 0, 0
	for i, c := range cols {
		// Spans up to ~1.5 frames wide are one frame with a wide pose (arms out).
		counts[i] = max(1, int(math.Round(float64(c[1]-c[0])/w.single-0.2)))
		est += counts[i]
		total += c[1] - c[0]
	}
	if est == w.n || est < w.n-1 || est > w.n+1 || w.n < len(cols) {
		return counts
	}
	// Largest-remainder split of w.n frames proportional to span width, at least 1 each.
	rem := make([]float64, len(cols))
	given := 0
	for i, c := range cols {
		exact := float64(w.n) * float64(c[1]-c[0]) / float64(total)
		counts[i] = max(1, int(exact))
		rem[i] = exact - float64(counts[i])
		given += counts[i]
	}
	for given < w.n {
		best := 0
		for i := range rem {
			if rem[i] > rem[best] {
				best = i
			}
		}
		counts[best]++
		rem[best] = -1
		given++
	}
	return counts
}

func medianWidth(sp [][2]int) float64 {
	w := make([]int, len(sp))
	for i, s := range sp {
		w[i] = s[1] - s[0]
	}
	sort.Ints(w)
	return float64(w[len(w)/2])
}

// hasWideSpan reports a span much wider than the others: two frames touching.
func hasWideSpan(sp [][2]int) bool {
	if len(sp) == 0 {
		return true
	}
	med := medianWidth(sp)
	for _, s := range sp {
		if float64(s[1]-s[0]) > 1.6*med {
			return true
		}
	}
	return len(sp) == 1
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
