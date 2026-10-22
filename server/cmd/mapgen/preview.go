package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

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

	kinds := map[string]color.RGBA{
		"road":   {108, 104, 96, 255},
		"bush":   {46, 104, 58, 255},
		"mud":    {96, 76, 48, 255},
		"wind":   {60, 86, 110, 255},
	}
	for _, o := range b.layers["zones"] {
		if c, ok := kinds[o.Type]; ok {
			fill(o.X, o.Y, o.X+o.Width, o.Y+o.Height, c)
		}
	}
	for _, o := range b.layers["obstacles"] {
		c := color.RGBA{150, 150, 150, 255}
		if o.Type == "stone_wall" {
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
