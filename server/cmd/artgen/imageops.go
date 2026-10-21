package main

import (
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg" // minigen may return JPEG
	"image/png"
	"math"
	"os"
	"path/filepath"
)

func loadImage(path string) (*image.NRGBA, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	src, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	out := image.NewNRGBA(src.Bounds().Sub(src.Bounds().Min))
	draw.Draw(out, out.Bounds(), src, src.Bounds().Min, draw.Src)
	return out, nil
}

func savePNG(path string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	return enc.Encode(f, img)
}

// whiteness is 0 for saturated/dark pixels and 1 for pure white.
func whiteness(c color.NRGBA) float64 {
	lo := math.Min(float64(c.R), math.Min(float64(c.G), float64(c.B)))
	hi := math.Max(float64(c.R), math.Max(float64(c.G), float64(c.B)))
	if hi-lo > 30 {
		return 0
	}
	return math.Max(0, (lo-200)/55)
}

// removeBackground picks the keying method from the corner colour: flat magenta (current
// prompts, ART_PROMPTS §5.1) or white (images generated before the switch).
func removeBackground(img *image.NRGBA) {
	if isMagenta(img.NRGBAAt(2, 2)) && isMagenta(img.NRGBAAt(img.Bounds().Dx()-3, img.Bounds().Dy()-3)) {
		removeMagenta(img)
		return
	}
	removeWhite(img)
}

func magentaStrength(c color.NRGBA) float64 {
	return math.Min(float64(c.R), float64(c.B)) - float64(c.G)
}

func isMagenta(c color.NRGBA) bool {
	return c.R > 150 && c.B > 150 && magentaStrength(c) > 60
}

// removeMagenta floods magenta from the border, keys pure magenta holes anywhere, and
// un-mixes magenta out of the anti-aliased edge band so no pink fringe is left.
func removeMagenta(img *image.NRGBA) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	bg := floodFrom(img, w, h, isMagenta)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := img.NRGBAAt(x, y)
			if bg[y*w+x] || (magentaStrength(c) > 200 && c.R > 220 && c.B > 220) {
				bg[y*w+x] = true
				a := magentaAlpha(c)
				if a < 0.25 {
					a = 0 // background noise
				}
				img.SetNRGBA(x, y, unmixMagenta(c, a))
			}
		}
	}
	// Edge band: up to 2 px next to the background, partially covered by magenta.
	for pass := 0; pass < 2; pass++ {
		next := append([]bool(nil), bg...)
		for i := range bg {
			x, y := i%w, i/w
			c := img.NRGBAAt(x, y)
			if bg[i] || !touchesBG(bg, x, y, w, h) || magentaStrength(c) < 30 {
				continue
			}
			img.SetNRGBA(x, y, unmixMagenta(c, magentaAlpha(c)))
			next[i] = true
		}
		bg = next
	}
}

// magentaAlpha estimates subject coverage from how magenta a pixel is (foreground ≈ not magenta).
func magentaAlpha(c color.NRGBA) float64 {
	return geomClamp((255-magentaStrength(c))/255, 0, 1)
}

// unmixMagenta recovers the foreground colour of a pixel blended over #FF00FF with coverage a.
func unmixMagenta(c color.NRGBA, a float64) color.NRGBA {
	if a < 0.02 {
		return color.NRGBA{}
	}
	un := func(v, bgv float64) uint8 { return uint8(geomClamp((v-(1-a)*bgv)/a, 0, 255)) }
	return color.NRGBA{R: un(float64(c.R), 255), G: un(float64(c.G), 0), B: un(float64(c.B), 255), A: uint8(a * 255)}
}

func geomClamp(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }

// floodFrom marks pixels matching pred that are connected to the image border.
func floodFrom(img *image.NRGBA, w, h int, pred func(color.NRGBA) bool) []bool {
	bg := make([]bool, w*h)
	stack := make([]int, 0, w*4)
	push := func(x, y int) {
		i := y*w + x
		if bg[i] || !pred(img.NRGBAAt(x, y)) {
			return
		}
		bg[i] = true
		stack = append(stack, i)
	}
	for x := 0; x < w; x++ {
		push(x, 0)
		push(x, h-1)
	}
	for y := 0; y < h; y++ {
		push(0, y)
		push(w-1, y)
	}
	for len(stack) > 0 {
		i := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		x, y := i%w, i/w
		if x > 0 {
			push(x-1, y)
		}
		if x < w-1 {
			push(x+1, y)
		}
		if y > 0 {
			push(x, y-1)
		}
		if y < h-1 {
			push(x, y+1)
		}
	}
	return bg
}

// removeWhite flood-fills near-white pixels connected to the border and makes them
// transparent. The thick black outline of the style keeps white inside the sprite intact.
func removeWhite(img *image.NRGBA) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	bg := floodFrom(img, w, h, func(c color.NRGBA) bool { return whiteness(c) >= 0.55 })
	applyMask(img, bg, w, h)
}

// applyMask clears background pixels and softens the 1px fringe next to them.
func applyMask(img *image.NRGBA, bg []bool, w, h int) {
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			c := img.NRGBAAt(x, y)
			if bg[i] {
				img.SetNRGBA(x, y, color.NRGBA{})
				continue
			}
			if touchesBG(bg, x, y, w, h) {
				c.A = uint8(float64(c.A) * (1 - 0.8*whiteness(c)))
				img.SetNRGBA(x, y, c)
			}
		}
	}
}

func touchesBG(bg []bool, x, y, w, h int) bool {
	for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		nx, ny := x+d[0], y+d[1]
		if nx >= 0 && ny >= 0 && nx < w && ny < h && bg[ny*w+nx] {
			return true
		}
	}
	return false
}

// alphaBBox returns the bounding box of pixels with alpha above a threshold.
func alphaBBox(img *image.NRGBA, r image.Rectangle) image.Rectangle {
	minX, minY, maxX, maxY := r.Max.X, r.Max.Y, r.Min.X-1, r.Min.Y-1
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if img.NRGBAAt(x, y).A > 24 {
				minX, minY = min(minX, x), min(minY, y)
				maxX, maxY = max(maxX, x), max(maxY, y)
			}
		}
	}
	if maxX < minX {
		return image.Rectangle{}
	}
	return image.Rect(minX, minY, maxX+1, maxY+1)
}

// resize does area-average (box) downscaling and bilinear upscaling; good enough for sprites.
func resize(src *image.NRGBA, r image.Rectangle, w, h int) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	sx, sy := float64(r.Dx())/float64(w), float64(r.Dy())/float64(h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.SetNRGBA(x, y, sampleBox(src, r, float64(x)*sx, float64(y)*sy, math.Max(sx, 1), math.Max(sy, 1)))
		}
	}
	return dst
}

// sampleBox averages a box of source pixels with premultiplied alpha.
func sampleBox(src *image.NRGBA, r image.Rectangle, fx, fy, bw, bh float64) color.NRGBA {
	x0, y0 := r.Min.X+int(fx), r.Min.Y+int(fy)
	x1, y1 := min(r.Max.X, r.Min.X+int(math.Ceil(fx+bw))), min(r.Max.Y, r.Min.Y+int(math.Ceil(fy+bh)))
	var sr, sg, sb, sa, n float64
	for y := y0; y < max(y1, y0+1); y++ {
		for x := x0; x < max(x1, x0+1); x++ {
			c := src.NRGBAAt(min(x, r.Max.X-1), min(y, r.Max.Y-1))
			a := float64(c.A) / 255
			sr, sg, sb, sa = sr+float64(c.R)*a, sg+float64(c.G)*a, sb+float64(c.B)*a, sa+a
			n++
		}
	}
	if sa < 1e-6 {
		return color.NRGBA{}
	}
	return color.NRGBA{R: uint8(sr / sa), G: uint8(sg / sa), B: uint8(sb / sa), A: uint8(255 * sa / n)}
}

// spans finds runs of non-empty lines along one axis of r (used to split sprite sheets).
func spans(img *image.NRGBA, r image.Rectangle, vertical bool, minGap int) [][2]int {
	n := r.Dy()
	if vertical {
		n = r.Dx()
	}
	filled := make([]bool, n)
	for i := 0; i < n; i++ {
		filled[i] = lineHasInk(img, r, vertical, i)
	}
	var out [][2]int
	start, gap := -1, 0
	for i := 0; i < n; i++ {
		switch {
		case filled[i] && start < 0:
			start, gap = i, 0
		case filled[i]:
			gap = 0
		case start >= 0:
			gap++
			if gap >= minGap {
				out = append(out, [2]int{start, i - gap + 1})
				start = -1
			}
		default:
		}
	}
	if start >= 0 {
		out = append(out, [2]int{start, n})
	}
	return out
}

func lineHasInk(img *image.NRGBA, r image.Rectangle, vertical bool, i int) bool {
	count := 0
	if vertical {
		for y := r.Min.Y; y < r.Max.Y; y++ {
			if img.NRGBAAt(r.Min.X+i, y).A > 40 {
				count++
			}
		}
	} else {
		for x := r.Min.X; x < r.Max.X; x++ {
			if img.NRGBAAt(x, r.Min.Y+i).A > 40 {
				count++
			}
		}
	}
	return count > 2
}

// dropEdgeFragments erases small blobs touching the left/right edge of a cell: pieces of
// the neighbouring frame cut off when frames touch.
func dropEdgeFragments(img *image.NRGBA, cell image.Rectangle) {
	w, h := cell.Dx(), cell.Dy()
	label := make([]int, w*h)
	var sizes []int
	var touches []bool
	for i := range label {
		x, y := i%w, i/w
		if label[i] != 0 || img.NRGBAAt(cell.Min.X+x, cell.Min.Y+y).A <= 40 {
			continue
		}
		id := len(sizes) + 1
		n, edge := floodLabel(img, cell, label, i, id)
		sizes, touches = append(sizes, n), append(touches, edge)
	}
	biggest := 0
	for _, n := range sizes {
		biggest = max(biggest, n)
	}
	for i, l := range label {
		if l != 0 && touches[l-1] && sizes[l-1]*10 < biggest*3 {
			img.SetNRGBA(cell.Min.X+i%w, cell.Min.Y+i/w, color.NRGBA{})
		}
	}
}

// floodLabel labels one 8-connected blob; returns its size and whether it touches a side edge.
func floodLabel(img *image.NRGBA, cell image.Rectangle, label []int, start, id int) (int, bool) {
	w, h := cell.Dx(), cell.Dy()
	stack := []int{start}
	label[start] = id
	n, edge := 0, false
	for len(stack) > 0 {
		i := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		n++
		x, y := i%w, i/w
		if x == 0 || x == w-1 {
			edge = true
		}
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				nx, ny := x+dx, y+dy
				if nx < 0 || ny < 0 || nx >= w || ny >= h {
					continue
				}
				j := ny*w + nx
				if label[j] == 0 && img.NRGBAAt(cell.Min.X+nx, cell.Min.Y+ny).A > 40 {
					label[j] = id
					stack = append(stack, j)
				}
			}
		}
	}
	return n, edge
}
