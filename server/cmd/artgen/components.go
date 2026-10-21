package main

import (
	"image"
	"image/draw"
	"log"
	"sort"
)

// Frame cutting by connected pixel regions (ART_PROMPTS §5.1): every frame is cut from its
// own pixels, so a sheet that drifts off the grid — or an arm reaching into the next cell —
// still slices cleanly.

const (
	inkAlpha      = 40  // alpha above this counts as drawn
	smallBlobFrac = 0.2 // blobs smaller than 20% of a typical frame are fragments (sweat, sparks)
	noiseArea     = 30  // px; specks below this are dropped
)

type blob struct {
	ids  []int32
	rect image.Rectangle
	area int
}

// labelBand labels 8-connected drawn regions inside band. labels is indexed relative to band.
func labelBand(img *image.NRGBA, band image.Rectangle) ([]int32, []*blob) {
	w, h := band.Dx(), band.Dy()
	labels := make([]int32, w*h)
	var blobs []*blob
	stack := make([]int, 0, 1024)
	for start := range labels {
		if labels[start] != 0 || img.NRGBAAt(band.Min.X+start%w, band.Min.Y+start/w).A <= inkAlpha {
			continue
		}
		id := int32(len(blobs) + 1)
		b := &blob{ids: []int32{id}, rect: image.Rect(start%w, start/w, start%w+1, start/w+1)}
		labels[start] = id
		stack = append(stack[:0], start)
		for len(stack) > 0 {
			i := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			x, y := i%w, i/w
			b.area++
			b.rect = b.rect.Union(image.Rect(x, y, x+1, y+1))
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					nx, ny := x+dx, y+dy
					if nx < 0 || ny < 0 || nx >= w || ny >= h {
						continue
					}
					j := ny*w + nx
					if labels[j] == 0 && img.NRGBAAt(band.Min.X+nx, band.Min.Y+ny).A > inkAlpha {
						labels[j] = id
						stack = append(stack, j)
					}
				}
			}
		}
		blobs = append(blobs, b)
	}
	return labels, blobs
}

// groupBlobs attaches fragments to the nearest big blob and returns big groups sorted left→right.
func groupBlobs(blobs []*blob, want int) []*blob {
	areas := make([]int, 0, len(blobs))
	for _, b := range blobs {
		areas = append(areas, b.area)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(areas)))
	if len(areas) == 0 {
		return nil
	}
	top := areas[:min(max(want, 1), len(areas))]
	ref := top[len(top)/2]
	var big, small []*blob
	for _, b := range blobs {
		switch {
		case float64(b.area) >= smallBlobFrac*float64(ref):
			big = append(big, b)
		case b.area >= noiseArea:
			small = append(small, b)
		default:
		}
	}
	for _, s := range small {
		if near := nearestBlob(big, s.rect); near != nil {
			near.ids = append(near.ids, s.ids...)
			near.rect = near.rect.Union(s.rect)
		}
	}
	sort.Slice(big, func(i, j int) bool {
		return big[i].rect.Min.X+big[i].rect.Max.X < big[j].rect.Min.X+big[j].rect.Max.X
	})
	return big
}

func nearestBlob(blobs []*blob, r image.Rectangle) *blob {
	var best *blob
	bestD := 1 << 30
	for _, b := range blobs {
		dx := max(0, max(b.rect.Min.X-r.Max.X, r.Min.X-b.rect.Max.X))
		dy := max(0, max(b.rect.Min.Y-r.Max.Y, r.Min.Y-b.rect.Max.Y))
		if d := dx*dx + dy*dy; d < bestD {
			best, bestD = b, d
		}
	}
	return best
}

// isolate copies only the pixels of the given labels inside clip (band-relative) into a
// tight transparent image.
func isolate(img *image.NRGBA, band image.Rectangle, labels []int32, ids []int32, clip image.Rectangle) *image.NRGBA {
	keep := map[int32]bool{}
	for _, id := range ids {
		keep[id] = true
	}
	w := band.Dx()
	out := image.NewNRGBA(clip)
	for y := clip.Min.Y; y < clip.Max.Y; y++ {
		for x := clip.Min.X; x < clip.Max.X; x++ {
			if l := labels[y*w+x]; l != 0 && keep[l] {
				out.SetNRGBA(x, y, img.NRGBAAt(band.Min.X+x, band.Min.Y+y))
			}
		}
	}
	box := alphaBBox(out, out.Bounds())
	if box.Empty() {
		return nil
	}
	tight := image.NewNRGBA(image.Rect(0, 0, box.Dx(), box.Dy()))
	draw.Draw(tight, tight.Bounds(), out, box.Min, draw.Src)
	return tight
}

// rowFrames cuts one sheet row into frame images.
func rowFrames(img *image.NRGBA, band image.Rectangle, w rowWant) []*image.NRGBA {
	labels, blobs := labelBand(img, band)
	groups := groupBlobs(blobs, w.n)
	cols := make([][2]int, len(groups))
	for i, g := range groups {
		cols[i] = [2]int{g.rect.Min.X, g.rect.Max.X}
	}
	debugf("  want %d blobs %d groups %v single %.0f", w.n, len(blobs), cols, w.single)
	counts := make([]int, len(groups))
	for i := range counts {
		counts[i] = 1
	}
	if len(groups) != w.n && w.single > 0 {
		counts = framesPerSpan(cols, w)
	}
	if total := sum(counts); total < w.n-1 {
		log.Printf("  warning: row has %d frames, expected %d — frames may overlap; regenerate this sheet if it looks wrong", total, w.n)
	}
	var out []*image.NRGBA
	for i, g := range groups {
		for _, clip := range splitClip(img, band, g.rect, counts[i]) {
			if f := isolate(img, band, labels, g.ids, clip); f != nil {
				out = append(out, f)
			}
		}
	}
	return out
}

// splitClip divides a group of touching frames into k clips at the emptiest columns.
func splitClip(img *image.NRGBA, band, r image.Rectangle, k int) []image.Rectangle {
	if k <= 1 {
		return []image.Rectangle{r}
	}
	abs := r.Add(band.Min)
	var out []image.Rectangle
	for _, c := range inkCuts(img, abs, k) {
		out = append(out, image.Rect(r.Min.X+c[0], r.Min.Y, r.Min.X+c[1], r.Max.Y))
	}
	return out
}

// groupWidths measures frame widths across all bands to estimate a single frame width.
func groupWidths(img *image.NRGBA, bands []image.Rectangle, rows []int) [][][2]int {
	out := make([][][2]int, len(bands))
	for i, band := range bands {
		_, blobs := labelBand(img, band)
		for _, g := range groupBlobs(blobs, rows[i]) {
			out[i] = append(out[i], [2]int{g.rect.Min.X, g.rect.Max.X})
		}
	}
	return out
}

func sum(xs []int) int {
	t := 0
	for _, x := range xs {
		t += x
	}
	return t
}
