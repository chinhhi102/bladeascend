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
	"sort"
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
	// Generators sometimes paint a duller pink/purple instead of #FF00FF: key out whatever
	// flat colour fills the border when it is neither white nor true magenta.
	key, flat := borderKey(img)
	switch {
	case flat && whiteness(key) == 0 && colorDist(key, color.NRGBA{255, 0, 255, 255}) > keyBG:
		removeKeyed(img, key)
	case magentaShare(img) > 0.15:
		removeMagenta(img)
	default:
		removeWhite(img)
	}
}

const (
	keyBG   = 60 // colour distance below this is background
	keyEdge = 130
	keySeed = 18 // pixels this close to the key are background even when walled off
)

// borderKey returns the median border colour when the border is one flat colour.
func borderKey(img *image.NRGBA) (color.NRGBA, bool) {
	b := img.Bounds()
	var px []color.NRGBA
	for x := b.Min.X; x < b.Max.X; x += 5 {
		px = append(px, img.NRGBAAt(x, b.Min.Y+2), img.NRGBAAt(x, b.Max.Y-3))
	}
	for y := b.Min.Y; y < b.Max.Y; y += 5 {
		px = append(px, img.NRGBAAt(b.Min.X+2, y), img.NRGBAAt(b.Max.X-3, y))
	}
	ch := func(f func(color.NRGBA) uint8) uint8 {
		v := make([]int, len(px))
		for i, c := range px {
			v[i] = int(f(c))
		}
		sort.Ints(v)
		return uint8(v[len(v)/2])
	}
	key := color.NRGBA{ch(func(c color.NRGBA) uint8 { return c.R }), ch(func(c color.NRGBA) uint8 { return c.G }), ch(func(c color.NRGBA) uint8 { return c.B }), 255}
	near := 0
	for _, c := range px {
		if colorDist(c, key) < keyBG {
			near++
		}
	}
	return key, float64(near) > 0.7*float64(len(px))
}

func colorDist(a, b color.NRGBA) float64 {
	dr, dg, db := float64(a.R)-float64(b.R), float64(a.G)-float64(b.G), float64(a.B)-float64(b.B)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

// removeKeyed is removeMagenta for an arbitrary flat background colour.
func removeKeyed(img *image.NRGBA, key color.NRGBA) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	isBG := func(c color.NRGBA) bool { return colorDist(c, key) < keyBG }
	bg := floodFromSeeds(img, w, h, isBG, func(c color.NRGBA) bool { return colorDist(c, key) < keySeed })
	applyMask(img, bg, w, h)
	for pass := 0; pass < 2; pass++ {
		next := append([]bool(nil), bg...)
		for i := range bg {
			x, y := i%w, i/w
			c := img.NRGBAAt(x, y)
			if bg[i] || !touchesBG(bg, x, y, w, h) {
				continue
			}
			if a := geomClamp((colorDist(c, key)-keyBG/2)/(keyEdge-keyBG/2), 0, 1); a < 1 {
				img.SetNRGBA(x, y, unmixKey(c, key, a))
				next[i] = true
			}
		}
		bg = next
	}
}

func unmixKey(c, key color.NRGBA, a float64) color.NRGBA {
	if a <= 0 {
		return color.NRGBA{}
	}
	f := func(v, k uint8) uint8 { return uint8(geomClamp((float64(v)-(1-a)*float64(k))/a, 0, 255)) }
	return color.NRGBA{f(c.R, key.R), f(c.G, key.G), f(c.B, key.B), uint8(a * float64(c.A))}
}

// magentaShare samples the image and returns the fraction of clearly magenta pixels; corners
// alone are unreliable because some sheets have a cell border drawn on the image edge.
func magentaShare(img *image.NRGBA) float64 {
	b := img.Bounds()
	n, m := 0, 0
	for y := b.Min.Y; y < b.Max.Y; y += 7 {
		for x := b.Min.X; x < b.Max.X; x += 7 {
			n++
			if isMagenta(img.NRGBAAt(x, y)) {
				m++
			}
		}
	}
	return float64(m) / float64(max(n, 1))
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
	// Seed from the border and from every strongly magenta pixel: grid borders can wall cells
	// off from the image edge.
	bg := floodFromSeeds(img, w, h, isMagenta, func(c color.NRGBA) bool {
		return magentaStrength(c) > 200 && c.R > 220 && c.B > 220
	})
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

// floodFromSeeds is floodFrom plus extra seed pixels anywhere in the image.
func floodFromSeeds(img *image.NRGBA, w, h int, pred, seed func(color.NRGBA) bool) []bool {
	bg := floodFrom(img, w, h, pred)
	stack := []int{}
	for i := range bg {
		if !bg[i] && seed(img.NRGBAAt(i%w, i/w)) {
			bg[i] = true
			stack = append(stack, i)
		}
	}
	for len(stack) > 0 {
		i := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		x, y := i%w, i/w
		for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := x+d[0], y+d[1]
			if nx < 0 || ny < 0 || nx >= w || ny >= h {
				continue
			}
			j := ny*w + nx
			if !bg[j] && pred(img.NRGBAAt(nx, ny)) {
				bg[j] = true
				stack = append(stack, j)
			}
		}
	}
	return bg
}

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

// alphaBBoxDense is alphaBBox that ignores specks: it keeps only connected blobs holding at
// least 1% of the biggest blob's pixels. Background removal regularly leaves a handful of
// stray pixels in a corner, and with a plain bounding box those alone stretch the content box
// across the whole image — the sprite then comes out tiny and pushed to one side.
func alphaBBoxDense(img *image.NRGBA, r image.Rectangle) image.Rectangle {
	w, h := r.Dx(), r.Dy()
	if w == 0 || h == 0 {
		return image.Rectangle{}
	}
	opaque := func(x, y int) bool { return img.NRGBAAt(r.Min.X+x, r.Min.Y+y).A > 24 }
	seen := make([]bool, w*h)
	type blob struct {
		n   int
		box image.Rectangle
	}
	var blobs []blob
	var stack []int
	for i := range seen {
		if seen[i] || !opaque(i%w, i/w) {
			continue
		}
		seen[i] = true
		stack = append(stack[:0], i)
		b := blob{}
		for len(stack) > 0 {
			p := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			x, y := p%w, p/w
			b.n++
			px := image.Rect(x, y, x+1, y+1)
			if b.box.Empty() {
				b.box = px
			} else {
				b.box = b.box.Union(px)
			}
			for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				nx, ny := x+d[0], y+d[1]
				if nx < 0 || ny < 0 || nx >= w || ny >= h {
					continue
				}
				if q := ny*w + nx; !seen[q] && opaque(nx, ny) {
					seen[q] = true
					stack = append(stack, q)
				}
			}
		}
		blobs = append(blobs, b)
	}
	if len(blobs) == 0 {
		return image.Rectangle{}
	}
	biggest := 0
	for _, b := range blobs {
		biggest = max(biggest, b.n)
	}
	var out image.Rectangle
	for _, b := range blobs {
		if b.n*100 < biggest {
			continue // dưới 1% mảng lớn nhất → bụi, bỏ
		}
		if out.Empty() {
			out = b.box
		} else {
			out = out.Union(b.box)
		}
	}
	return out.Add(r.Min)
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

// eraseGridLines clears straight dark lines that cross almost the whole image (cell borders
// some tools draw despite the prompt) and returns the y positions of the horizontal ones, which
// are reliable row boundaries, plus the x positions of full-height vertical ones (cell borders).
func eraseGridLines(img *image.NRGBA) ([]int, []int) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	dark := func(x, y int) bool {
		c := img.NRGBAAt(x, y)
		return c.A > 128 && int(c.R)+int(c.G)+int(c.B) < 3*90
	}
	var cols, rows []int
	for x := 0; x < w; x++ {
		if countLine(h, func(i int) bool { return dark(x, i) }) > 0.8*float64(h) {
			cols = append(cols, x)
		}
	}
	for y := 0; y < h; y++ {
		if countLine(w, func(i int) bool { return dark(i, y) }) > 0.8*float64(w) {
			rows = append(rows, y)
		}
	}
	// Clear 2 px around each line as well: anti-aliased edges would leave a faint border.
	for _, x := range cols {
		for dx := -2; dx <= 2; dx++ {
			for y := 0; y < h; y++ {
				if x+dx >= 0 && x+dx < w {
					img.SetNRGBA(x+dx, y, color.NRGBA{})
				}
			}
		}
	}
	for _, y := range rows {
		for dy := -2; dy <= 2; dy++ {
			for x := 0; x < w; x++ {
				if y+dy >= 0 && y+dy < h {
					img.SetNRGBA(x, y+dy, color.NRGBA{})
				}
			}
		}
	}
	return rows, cols
}

func countLine(n int, dark func(int) bool) float64 {
	c := 0
	for i := 0; i < n; i++ {
		if dark(i) {
			c++
		}
	}
	return float64(c)
}
