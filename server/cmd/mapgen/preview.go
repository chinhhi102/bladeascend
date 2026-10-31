package main

import (
	"image"
	"image/color"
	"image/png"
	"bladeascend/server/internal/config"
	"os"
)

// insidePoly: ray casting trên đa giác của một object Tiled (đỉnh lưu TƯƠNG ĐỐI so với X,Y).
func insidePoly(o object, x, y float64) bool {
	in := false
	n := len(o.Polygon)
	for i, j := 0, n-1; i < n; j, i = i, i+1 {
		xi, yi := o.X+o.Polygon[i].X, o.Y+o.Polygon[i].Y
		xj, yj := o.X+o.Polygon[j].X, o.Y+o.Polygon[j].Y
		if (yi > y) != (yj > y) && x < (xj-xi)*(y-yi)/(yj-yi)+xi {
			in = !in
		}
	}
	return in
}

// writePreview dumps a quick top-down PNG of the generated map so a seed can be eyeballed
// without starting the game (spec §8.1: `--seed N --png out.png`).
func writePreview(b *builder, path string) error {
	const px = 900
	k := px / b.size
	img := image.NewRGBA(image.Rect(0, 0, px, px))
	fill := func(x0, y0, x1, y1 float64, c color.RGBA) {
		for y := int(y0 * k); y <= int(y1*k); y++ {
			for x := int(x0 * k); x <= int(x1*k); x++ {
				if x >= 0 && y >= 0 && x < px && y < px {
					img.Set(x, y, c)
				}
			}
		}
	}
	fill(0, 0, b.size, b.size, color.RGBA{14, 16, 22, 255}) // vực ngoài bản đồ
	lo, hi := b.playable()
	fill(lo, lo, hi, hi, color.RGBA{58, 64, 52, 255}) // vùng chơi

	// Vùng ĐA GIÁC phải vẽ đúng hình, không được tô cả hộp bao: hồ băng hình méo mà tô
	// thành một ô vuông to thì ảnh duyệt nói dối về chỗ nào trơn chỗ nào không.
	fillPoly := func(o object, c color.RGBA) {
		x0, y0, x1, y1 := o.X, o.Y, o.X, o.Y
		for _, pt := range o.Polygon {
			x0, y0 = min(x0, o.X+pt.X), min(y0, o.Y+pt.Y)
			x1, y1 = max(x1, o.X+pt.X), max(y1, o.Y+pt.Y)
		}
		for py := int(y0 * k); py <= int(y1*k); py++ {
			for px2 := int(x0 * k); px2 <= int(x1*k); px2++ {
				if px2 < 0 || py < 0 || px2 >= px || py >= px {
					continue
				}
				if insidePoly(o, float64(px2)/k, float64(py)/k) {
					img.Set(px2, py, c)
				}
			}
		}
	}
	kinds := map[string]color.RGBA{
		"road":  {108, 104, 96, 255},
		"bush":  {46, 104, 58, 255},
		"mud":   {96, 76, 48, 255},
		"wind":  {60, 86, 110, 255},
		"water": {52, 84, 112, 255},
		"ice":   {176, 212, 232, 255},
		"snow":  {226, 232, 238, 255},
	}
	for _, o := range b.layers["zones"] {
		c, ok := kinds[o.Type]
		if !ok {
			continue
		}
		if len(o.Polygon) >= 3 {
			fillPoly(o, c)
			continue
		}
		fill(o.X, o.Y, o.X+o.Width, o.Y+o.Height, c)
	}
	for _, o := range b.layers["obstacles"] {
		c := color.RGBA{150, 150, 150, 255}
		if o.Type == config.BorderKind {
			c = color.RGBA{96, 108, 94, 255}
		}
		fill(o.X, o.Y, o.X+o.Width, o.Y+o.Height, c)
	}
	for _, o := range b.layers["gates"] {
		fill(o.X, o.Y, o.X+o.Width, o.Y+o.Height, color.RGBA{165, 91, 255, 255})
	}
	for _, o := range b.layers["altars"] {
		c := color.RGBA{159, 232, 255, 255}
		if o.Type == "center" {
			c = color.RGBA{255, 214, 10, 255}
		}
		fill(o.X, o.Y, o.X+o.Width, o.Y+o.Height, c)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
