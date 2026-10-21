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

// prepSheet writes each segment's frames under assets/<seg.Out>/. A row owned by a single
// animation keeps however many frames the model actually drew (Gemini sometimes draws 5 or 6
// instead of 4 or 8); rows shared by several animations are cut to the exact count.
func prepSheet(img *image.NRGBA, j Job, assets string) error {
	removeBackground(img)
	layout := rowLayout(j)
	rows := splitSheet(img, j.Rows, layout)
	type frame struct {
		seg int
		box image.Rectangle
	}
	var frames []frame
	for r, cells := range rows {
		ci := 0
		for _, part := range layout[r] {
			n := part.count
			if len(layout[r]) == 1 {
				n = len(cells)
			}
			for k := 0; k < n && ci < len(cells); k, ci = k+1, ci+1 {
				b := alphaBBox(img, cells[ci])
				if b.Empty() {
					continue
				}
				frames = append(frames, frame{part.seg, b})
			}
		}
	}
	boxes := make([][]image.Rectangle, len(j.Segs))
	for _, f := range frames {
		boxes[f.seg] = append(boxes[f.seg], f.box)
	}
	scales := sharedScales(j, boxes)
	for si, sg := range j.Segs {
		clearStaleFrames(filepath.Join(assets, sg.Out), sg)
		for f, box := range boxes[si] {
			s := scales[groupKey(sg)]
			if sg.Each {
				s = fitScale(sg, box.Dx(), box.Dy())
			}
			name := fmt.Sprintf("%s_%02d.png", sg.Anim, f)
			if f < len(sg.Names) {
				name = sg.Names[f] + ".png"
			}
			if err := savePNG(filepath.Join(assets, sg.Out, name), placeOne(img, box, sg, s)); err != nil {
				return err
			}
		}
	}
	return nil
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

// sharedScales gives every non-"each" segment group one scale so a character keeps the same
// size across all its animations.
func sharedScales(j Job, boxes [][]image.Rectangle) map[string]float64 {
	maxW, maxH := map[string]int{}, map[string]int{}
	for si, sg := range j.Segs {
		k := groupKey(sg)
		for _, b := range boxes[si] {
			maxW[k], maxH[k] = max(maxW[k], b.Dx()), max(maxH[k], b.Dy())
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

// splitSheet finds frame cells row by row using empty gaps.
func splitSheet(img *image.NRGBA, rows []int, layout [][]rowPart) [][]image.Rectangle {
	b := img.Bounds()
	raw := spans(img, b, false, minSheetGap)
	bands := fitSpans(raw, len(rows), b.Dy())
	debugf("row spans %v → %v", raw, bands)
	bandRects := make([]image.Rectangle, len(bands))
	colSpans := make([][][2]int, len(bands))
	for ri, rs := range bands {
		bandRects[ri] = image.Rect(b.Min.X, b.Min.Y+rs[0], b.Max.X, b.Min.Y+rs[1])
		colSpans[ri] = mergeFragments(spans(img, bandRects[ri], true, minSheetGap))
	}
	single := singleFrameWidth(colSpans)
	out := make([][]image.Rectangle, len(bands))
	for ri, band := range bandRects {
		for _, cs := range rowCells(img, band, colSpans[ri], rowWant{rows[ri], len(layout[ri]) > 1, single}) {
			cell := image.Rect(b.Min.X+cs[0], band.Min.Y, b.Min.X+cs[1], band.Max.Y)
			dropEdgeFragments(img, cell)
			out[ri] = append(out[ri], cell)
		}
	}
	return out
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
	exact  bool    // several animations share the row: must cut exactly n
	single float64 // estimated width of one frame
}

// rowCells decides how to cut one row into frames. Rows owned by one animation keep the
// number of frames the model actually drew.
func rowCells(img *image.NRGBA, band image.Rectangle, cols [][2]int, w rowWant) [][2]int {
	debugf("  want %d cols %v single %.0f", w.n, cols, w.single)
	if len(cols) == w.n {
		return cols
	}
	if w.exact {
		if len(cols) > w.n {
			return fitSpans(cols, w.n, band.Dx())
		}
		return inkCuts(img, band, w.n)
	}
	if w.single <= 0 {
		return cols
	}
	counts := framesPerSpan(cols, w)
	var out [][2]int
	for i, c := range cols {
		if counts[i] == 1 {
			out = append(out, c)
			continue
		}
		sub := image.Rect(band.Min.X+c[0], band.Min.Y, band.Min.X+c[1], band.Max.Y)
		for _, p := range inkCuts(img, sub, counts[i]) {
			out = append(out, [2]int{c[0] + p[0], c[0] + p[1]})
		}
	}
	return out
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

// mergeFragments folds spans narrower than 35% of the median (sparks, sweat drops) into a neighbour.
func mergeFragments(sp [][2]int) [][2]int {
	for len(sp) > 1 {
		med := medianWidth(sp)
		idx := -1
		for i, s := range sp {
			if float64(s[1]-s[0]) < 0.35*med {
				idx = i
				break
			}
		}
		if idx < 0 {
			return sp
		}
		j := idx - 1
		if j < 0 || (idx+1 < len(sp) && sp[idx+1][0]-sp[idx][1] < sp[idx][0]-sp[j][1]) {
			j = idx + 1
		}
		lo, hi := min(idx, j), max(idx, j)
		sp[lo] = [2]int{sp[lo][0], sp[hi][1]}
		sp = append(sp[:hi], sp[hi+1:]...)
	}
	return sp
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

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
