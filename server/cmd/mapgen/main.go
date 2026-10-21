// Command mapgen writes the three starter maps (spec §8.1) as Tiled .tmj files into maps/.
// They can be opened and edited in Tiled afterwards (see docs/MAP_GUIDE.md).
//
//	go run ./cmd/mapgen
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"

	"kiemxoay/server/internal/config"
)

const tile = 100 // Tiled grid cell in world units

type prop struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type object struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Width      float64 `json:"width"`
	Height     float64 `json:"height"`
	Rotation   float64 `json:"rotation"`
	Visible    bool    `json:"visible"`
	Ellipse    bool    `json:"ellipse,omitempty"`
	Point      bool    `json:"point,omitempty"`
	Properties []prop  `json:"properties,omitempty"`
}

type layer struct {
	ID        int      `json:"id"`
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	DrawOrder string   `json:"draworder"`
	Opacity   float64  `json:"opacity"`
	Visible   bool     `json:"visible"`
	X         int      `json:"x"`
	Y         int      `json:"y"`
	Objects   []object `json:"objects"`
}

type tmj struct {
	Type         string  `json:"type"`
	Version      string  `json:"version"`
	TiledVersion string  `json:"tiledversion"`
	Orientation  string  `json:"orientation"`
	RenderOrder  string  `json:"renderorder"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	TileWidth    int     `json:"tilewidth"`
	TileHeight   int     `json:"tileheight"`
	Infinite     bool    `json:"infinite"`
	NextLayerID  int     `json:"nextlayerid"`
	NextObjectID int     `json:"nextobjectid"`
	Properties   []prop  `json:"properties"`
	Layers       []layer `json:"layers"`
	Tilesets     []any   `json:"tilesets"`
}

// rect is a placed footprint used to keep obstacles apart.
type rect struct{ x0, y0, x1, y1 float64 }

func (a rect) overlaps(b rect, margin float64) bool {
	return a.x0-margin < b.x1 && b.x0-margin < a.x1 && a.y0-margin < b.y1 && b.y0-margin < a.y1
}

// builder accumulates objects per layer and rejects overlapping placements.
type builder struct {
	size   float64
	rng    *rand.Rand
	nextID int
	layers map[string][]object
	solid  []rect
	keep   []rect // areas that must stay clear (spawn plaza)
}

func newBuilder(size float64, seed int64) *builder {
	c := size / 2
	return &builder{size: size, rng: rand.New(rand.NewSource(seed)), nextID: 1, layers: map[string][]object{},
		keep: []rect{{c - 600, c - 600, c + 600, c + 600}}}
}

func (b *builder) add(layerName string, o object) {
	o.ID = b.nextID
	o.Visible = true
	b.nextID++
	b.layers[layerName] = append(b.layers[layerName], o)
}

func (b *builder) free(r rect, margin float64) bool {
	if r.x0 < 150 || r.y0 < 150 || r.x1 > b.size-150 || r.y1 > b.size-150 {
		return false
	}
	for _, k := range b.keep {
		if r.overlaps(k, 0) {
			return false
		}
	}
	for _, s := range b.solid {
		if r.overlaps(s, margin) {
			return false
		}
	}
	return true
}

// obstacle places a rectangle obstacle if the spot is free (margin keeps paths open).
func (b *builder) obstacle(kind string, x, y, w, h, margin float64) bool {
	r := rect{x, y, x + w, y + h}
	if !b.free(r, margin) {
		return false
	}
	b.solid = append(b.solid, r)
	b.add("obstacles", object{Type: kind, X: x, Y: y, Width: w, Height: h})
	return true
}

func (b *builder) boulder(cx, cy, r, margin float64) bool {
	rr := rect{cx - r, cy - r, cx + r, cy + r}
	if !b.free(rr, margin) {
		return false
	}
	b.solid = append(b.solid, rr)
	b.add("obstacles", object{Type: "boulder", X: cx - r, Y: cy - r, Width: 2 * r, Height: 2 * r, Ellipse: true})
	return true
}

// wall forces an indestructible wall segment (used for alleys; ignores spacing).
func (b *builder) wall(x, y, w, h float64) {
	b.solid = append(b.solid, rect{x, y, x + w, y + h})
	b.add("obstacles", object{Type: "stone_wall", X: x, Y: y, Width: w, Height: h})
}

func (b *builder) zone(kind string, x, y, w, h float64, ellipse bool, props ...prop) {
	b.add("zones", object{Type: kind, X: x, Y: y, Width: w, Height: h, Ellipse: ellipse, Properties: props})
}

func (b *builder) gate(x, y float64, network, color string) {
	b.keep = append(b.keep, rect{x - 300, y - 300, x + 300, y + 300})
	b.add("gates", object{Type: "gate", X: x - 150, Y: y - 150, Width: 300, Height: 300, Ellipse: true,
		Properties: []prop{{"network", "string", network}, {"color", "string", color}}})
}

func (b *builder) decor(sprite string, n int) {
	for i := 0; i < n; i++ {
		x, y := b.rnd(300, b.size-300), b.rnd(300, b.size-300)
		if b.free(rect{x - 60, y - 60, x + 60, y + 60}, 0) {
			b.add("decor", object{Type: sprite, X: x, Y: y, Point: true})
		}
	}
}

func (b *builder) rnd(lo, hi float64) float64 { return lo + b.rng.Float64()*(hi-lo) }

// scatter tries n times to place something via fn at random positions.
func (b *builder) scatter(n int, fn func(x, y float64) bool) {
	placed, tries := 0, 0
	for placed < n && tries < n*30 {
		tries++
		if fn(b.rnd(200, b.size-200), b.rnd(200, b.size-200)) {
			placed++
		}
	}
}

// alley builds a corridor of indestructible walls with an opening of the given width
// (150 / 250 / 350 — spec §8.3 hẻm hẹp). Big players cannot fit through narrow ones.
func (b *builder) alley(x, y, length float64, gap float64, vertical bool) {
	thick := 120.0
	if vertical {
		b.wall(x, y, thick, length)
		b.wall(x+thick+gap, y, thick, length)
		return
	}
	b.wall(x, y, length, thick)
	b.wall(x, y+thick+gap, length, thick)
}

// windLane is a long strip; dir in degrees (0 = east, 90 = south).
func (b *builder) windLane(x, y, w, h, dirDeg float64) {
	b.zone("wind", x, y, w, h, false, prop{"dir", "float", dirDeg})
}

func (b *builder) write(path, name, biome string) error {
	n := int(b.size / tile)
	order := []string{"zones", "obstacles", "gates", "decor"}
	var layers []layer
	for i, ln := range order {
		layers = append(layers, layer{ID: i + 1, Name: ln, Type: "objectgroup", DrawOrder: "topdown", Opacity: 1, Visible: true,
			Objects: append([]object{}, b.layers[ln]...)})
	}
	m := tmj{Type: "map", Version: "1.10", TiledVersion: "1.10.2", Orientation: "orthogonal", RenderOrder: "right-down",
		Width: n, Height: n, TileWidth: tile, TileHeight: tile, NextLayerID: len(order) + 1, NextObjectID: b.nextID,
		Properties: []prop{{"name", "string", name}, {"biome", "string", biome}}, Layers: layers, Tilesets: []any{}}
	data, err := json.MarshalIndent(m, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func main() {
	out := flag.String("out", config.FindRepoFile("maps"), "output directory")
	flag.Parse()
	cfg := config.MustLoadDefault()
	size := cfg.Room.MapSize
	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatal(err)
	}
	gens := []struct {
		id, name, biome string
		fn              func(*builder)
	}{
		{"lang_tre", "Làng Tre", "grass", langTre},
		{"cho_dem", "Chợ Đêm", "concrete", choDem},
		{"nui_da", "Núi Đá", "dirt", nuiDa},
	}
	for i, gdef := range gens {
		b := newBuilder(size, int64(1000+i))
		gdef.fn(b)
		path := filepath.Join(*out, gdef.id+".tmj")
		if err := b.write(path, gdef.name, gdef.biome); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: %d obstacles, %d zones, %d gates → %s\n", gdef.name, len(b.layers["obstacles"]), len(b.layers["zones"]), len(b.layers["gates"]), path)
	}
}

// langTre: balanced village — fences, houses, ponds of mud, bushes, two gate pairs.
func langTre(b *builder) {
	s := b.size
	b.gate(1500, 1500, "a", "purple")
	b.gate(s-1500, s-1500, "a", "purple")
	b.gate(s-1500, 1500, "b", "blue")
	b.gate(1500, s-1500, "b", "blue")
	b.windLane(800, s/2-200, s-1600, 400, 0)
	b.alley(s/2-2500, 2200, 1400, 150, true)
	b.alley(s/2+2200, s-3600, 1400, 250, true)
	b.alley(2400, s/2+1800, 1400, 350, false)
	b.scatter(40, func(x, y float64) bool { // houses: brick walls
		w, h := b.rnd(300, 600), b.rnd(150, 300)
		return b.obstacle("brick", x, y, w, h, 220)
	})
	b.scatter(70, func(x, y float64) bool { // bamboo fences, long and thin
		if b.rng.Intn(2) == 0 {
			return b.obstacle("fence", x, y, b.rnd(400, 900), 50, 200)
		}
		return b.obstacle("fence", x, y, 50, b.rnd(400, 900), 200)
	})
	b.scatter(25, func(x, y float64) bool { return b.obstacle("stall", x, y, 220, 180, 200) })
	b.scatter(15, func(x, y float64) bool { return b.boulder(x, y, b.rnd(180, 300), 250) })
	b.scatter(26, func(x, y float64) bool {
		r := b.rnd(250, 450)
		b.zone("bush", x-r, y-r, 2*r, 2*r, true)
		return true
	})
	b.scatter(8, func(x, y float64) bool {
		b.zone("mud", x, y, b.rnd(500, 900), b.rnd(400, 700), true)
		return true
	})
	for _, d := range []string{"banana_tree", "pile_of_watermelons", "red_plastic_stool", "lazy_sleeping_dog", "rusty_water_jar", "bush_with_flowers"} {
		b.decor(d, 14)
	}
}

// choDem: night market — rows of stalls and carts, many alleys of all widths.
func choDem(b *builder) {
	s := b.size
	b.gate(1200, s/2, "a", "orange")
	b.gate(s-1200, s/2, "a", "orange")
	b.gate(s/2, 1200, "b", "purple")
	b.gate(s/2, s-1200, "b", "purple")
	gaps := []float64{150, 250, 350}
	for i := 0; i < 9; i++ {
		x, y := b.rnd(1500, s-3000), b.rnd(1500, s-3000)
		b.alley(x, y, b.rnd(1000, 1800), gaps[i%3], i%2 == 0)
	}
	// Market rows: stalls/carts on a loose grid with lanes between them.
	for gy := 900.0; gy < s-900; gy += 700 {
		for gx := 900.0; gx < s-900; gx += 520 {
			if b.rng.Float64() < 0.45 {
				continue
			}
			kind := "stall"
			if b.rng.Intn(3) == 0 {
				kind = "cart"
			}
			b.obstacle(kind, gx+b.rnd(-60, 60), gy+b.rnd(-60, 60), 240, 180, 150)
		}
	}
	b.scatter(25, func(x, y float64) bool { return b.obstacle("brick", x, y, b.rnd(400, 800), 120, 200) })
	b.scatter(12, func(x, y float64) bool {
		r := b.rnd(220, 380)
		b.zone("bush", x-r, y-r, 2*r, 2*r, true)
		return true
	})
	b.scatter(6, func(x, y float64) bool {
		b.zone("mud", x, y, b.rnd(400, 700), b.rnd(300, 600), true)
		return true
	})
	b.windLane(s/2-200, 600, 400, s-1200, 90)
	b.windLane(600, s/4, s-1200, 350, 180)
	for _, d := range []string{"red_plastic_stool", "street_food_cart", "old_bicycle_leaning", "broken_tv", "stone_lantern"} {
		b.decor(d, 18)
	}
}

// nuiDa: wide rocky land with huge boulders, ridges and a 4-gate network.
func nuiDa(b *builder) {
	s := b.size
	net := [][2]float64{{1800, 1800}, {s - 1800, 1800}, {s - 1800, s - 1800}, {1800, s - 1800}}
	for _, p := range net {
		b.gate(p[0], p[1], "ring", "orange")
	}
	b.gate(s/2, 2600, "c", "blue")
	b.gate(s/2, s-2600, "c", "blue")
	b.scatter(45, func(x, y float64) bool { return b.boulder(x, y, b.rnd(250, 600), 300) })
	b.scatter(14, func(x, y float64) bool { // ridges
		if b.rng.Intn(2) == 0 {
			b.wall(x, y, b.rnd(900, 1600), 140)
		} else {
			b.wall(x, y, 140, b.rnd(900, 1600))
		}
		return true
	})
	b.alley(s/2-1500, s/2-200, 1000, 150, false)
	b.alley(s/2+700, s/2+600, 1000, 250, true)
	b.scatter(20, func(x, y float64) bool { return b.obstacle("fence", x, y, b.rnd(400, 800), 50, 220) })
	b.scatter(14, func(x, y float64) bool {
		r := b.rnd(250, 420)
		b.zone("bush", x-r, y-r, 2*r, 2*r, true)
		return true
	})
	b.scatter(10, func(x, y float64) bool {
		b.zone("mud", x, y, b.rnd(600, 1200), b.rnd(500, 900), true)
		return true
	})
	b.windLane(800, 800, s-1600, 380, 0)
	b.windLane(800, s-1200, s-1600, 380, 180)
	b.windLane(s-1200, 1300, 380, s-2600, 90)
	for _, d := range []string{"stone_lantern", "rusty_water_jar", "banana_tree"} {
		b.decor(d, 16)
	}
}
