// Package world holds the static map: obstacles, zones, gates and the path grid (spec §8).
package world

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Obstacle: Shape is what the player sees (swords hit it); Foot is the footprint at the
// bottom that blocks bodies (spec §8.2). MaxHP == 0 means indestructible.
type Obstacle struct {
	ID    int    `json:"id"`
	Kind  string `json:"k"`
	Shape Shape  `json:"s"`
	Foot  Shape  `json:"f"`
	MaxHP int    `json:"hp"`
}

// Block is the shape that stops bodies: the footprint, or the whole shape when none is set.
func (o Obstacle) Block() Shape {
	if o.Foot.W > 0 {
		return o.Foot
	}
	return o.Shape
}

// LoadOptions are the config values the loader needs.
type LoadOptions struct {
	HP         map[string]int // obstacle type → HP (missing = indestructible)
	FootprintW float64        // default footprint width, fraction of the sprite
	FootprintH float64        // default footprint height, fraction of the sprite (at the bottom)
}

// Footprint is a rectangle w×h (fractions of the shape's bounds) sitting on the shape's bottom edge.
func Footprint(s Shape, w, h float64) Shape {
	lo, hi := s.Bounds()
	bw, bh := hi.X-lo.X, hi.Y-lo.Y
	fw, fh := bw*w, bh*h
	return Shape{X: lo.X + (bw-fw)/2, Y: hi.Y - fh, W: fw, H: fh}
}

// footprintOf reads the Tiled property footprint = "w,h" (fractions), else the defaults.
func footprintOf(o tmjObject, s Shape, opt LoadOptions) Shape {
	w, h := opt.FootprintW, opt.FootprintH
	if v := propString(o.Properties, "footprint", ""); v != "" {
		if _, err := fmt.Sscanf(strings.ReplaceAll(v, " ", ""), "%g,%g", &w, &h); err != nil {
			w, h = opt.FootprintW, opt.FootprintH
		}
	}
	if w <= 0 || h <= 0 {
		return Shape{}
	}
	return Footprint(s, w, h)
}

// Zone is a non-blocking terrain area: "bush", "mud" or "wind".
type Zone struct {
	ID      int     `json:"id"`
	Kind    string  `json:"k"`
	Shape   Shape   `json:"s"`
	WindDir float64 `json:"wd,omitempty"` // radians, wind zones only
}

// Gate is a teleport portal. Gates sharing a Network are linked.
type Gate struct {
	ID      int     `json:"id"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Network string  `json:"n"`
	Color   string  `json:"c"`
	// Heaven: đây là **Thiên Môn** (spec §8.4a) — cổng MỘT CHIỀU luôn ra ở Thiên Đình,
	// khác hẳn 4 cổng thường vốn bắn ngẫu nhiên khắp thế giới.
	Heaven bool `json:"hv,omitempty"`
	// Dock: đây là **Bến Thiên Môn** ở Thiên Đình — điểm RA của Thiên Môn khu `DockFor`.
	Dock    bool `json:"dk,omitempty"`
	DockFor int  `json:"df,omitempty"`
}

// HeavenGate trả về Thiên Môn của bản đồ này (nil nếu khu này không có).
func (m *Map) HeavenGate() *Gate {
	for i := range m.Gates {
		if m.Gates[i].Heaven {
			return &m.Gates[i]
		}
	}
	return nil
}

// DockFor trả về Bến Thiên Môn dành cho khu `zone` (nil nếu không có).
func (m *Map) DockFor(zone int) *Gate {
	for i := range m.Gates {
		if m.Gates[i].Dock && m.Gates[i].DockFor == zone {
			return &m.Gates[i]
		}
	}
	return nil
}

// Altar is a Đài Tiên Khí (spec §8.5): "small" quanh 4 sân môn phái, "center" ở giữa.
// Đài không phá được, kiếm đi xuyên; nó chỉ là điểm phun viên Tiên khí.
type Altar struct {
	ID   int     `json:"id"`
	Kind string  `json:"k"` // "small" | "center"
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	R    float64 `json:"r"`
}

// Decor is purely visual; the server only forwards it.
type Decor struct {
	Sprite string  `json:"sp"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
}

type Map struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Size      float64    `json:"size"`
	Biome     string     `json:"biome"`
	Obstacles []Obstacle `json:"obstacles"`
	Zones     []Zone     `json:"zones"`
	Gates     []Gate     `json:"gates"`
	Altars    []Altar    `json:"altars"`
	// PlayMargin: khoảng cách từ mép bản đồ tới vùng chơi được (ngoài đó là tường biên và
	// vực tối). Mọi thứ sinh ra phải nằm trong [PlayMargin, Size-PlayMargin] (spec §8.1).
	PlayMargin float64 `json:"play_margin"`
	Decor      []Decor `json:"decor"`
}

// EmptyMap is used by tests and when no .tmj file is available.
func EmptyMap(size float64) *Map {
	m := &Map{ID: "empty", Name: "Bãi Trống", Size: size, Biome: "grass"}
	m.normalize()
	return m
}

// normalize replaces nil slices so the JSON for the client always has arrays.
func (m *Map) normalize() {
	if m.Obstacles == nil {
		m.Obstacles = []Obstacle{}
	}
	if m.Zones == nil {
		m.Zones = []Zone{}
	}
	if m.Gates == nil {
		m.Gates = []Gate{}
	}
	if m.Altars == nil {
		m.Altars = []Altar{}
	}
	if m.Decor == nil {
		m.Decor = []Decor{}
	}
}

// GatesInNetwork returns the other gates linked to gate g.
func (m *Map) GatesInNetwork(g Gate) []Gate {
	var out []Gate
	for _, o := range m.Gates {
		if o.Network == g.Network && o.ID != g.ID {
			out = append(out, o)
		}
	}
	return out
}

// GateExit is the one gate that g leads to. Gates in a network are paired two by two in ID
// order (0↔1, 2↔3, …), so every gate is a fixed two-way door: whatever you came out of
// takes you straight back. A network with an odd gate left over links it to the first one.
// GateNetwork trả về mọi cổng KHÁC cùng mạng với g, xếp theo ID cho ổn định.
//
// Dùng cho luật "cổng dịch chuyển ngẫu nhiên": người chơi bước vào một cổng thì ra ở một
// cổng bất kỳ trong mạng, không phải cổng cố định như trước.
func (m *Map) GateNetwork(g Gate) []Gate {
	var out []Gate
	for _, o := range m.Gates {
		if o.Network == g.Network && o.ID != g.ID {
			out = append(out, o)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (m *Map) GateExit(g Gate) (Gate, bool) {
	var net []Gate
	for _, o := range m.Gates {
		if o.Network == g.Network {
			net = append(net, o)
		}
	}
	if len(net) < 2 {
		return Gate{}, false
	}
	sort.Slice(net, func(i, j int) bool { return net[i].ID < net[j].ID })
	for i, o := range net {
		if o.ID != g.ID {
			continue
		}
		if pair := i ^ 1; pair < len(net) {
			return net[pair], true
		}
		return net[0], true // lẻ cổng: nối về cổng đầu mạng
	}
	return Gate{}, false
}

// LoadDir loads every .tmj file in dir.
func LoadDir(dir string, opt LoadOptions) ([]*Map, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.tmj"))
	if err != nil {
		return nil, err
	}
	var maps []*Map
	for _, f := range files {
		m, err := LoadTMJ(f, opt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		maps = append(maps, m)
	}
	return maps, nil
}

// LoadTMJ parses a Tiled JSON map.
func LoadTMJ(path string, opt LoadOptions) (*Map, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var t tmjFile
	if err := json.Unmarshal(raw, &t); err != nil {
		return nil, err
	}
	id := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	m := &Map{ID: id, Size: float64(t.Width * t.TileWidth)}
	m.Name = propString(t.Properties, "name", id)
	m.PlayMargin = propFloat(t.Properties, "play_margin", 0)
	m.Biome = propString(t.Properties, "biome", "grass")
	for _, l := range t.Layers {
		m.addLayer(l, opt)
	}
	m.normalize()
	return m, nil
}

func (m *Map) addLayer(l tmjLayer, opt LoadOptions) {
	for _, o := range l.Objects {
		kind := o.kind()
		switch strings.ToLower(l.Name) {
		case "obstacles":
			sh := o.shape()
			m.Obstacles = append(m.Obstacles, Obstacle{ID: len(m.Obstacles), Kind: kind, Shape: sh,
				Foot: footprintOf(o, sh, opt), MaxHP: opt.HP[kind]})
		case "zones":
			m.Zones = append(m.Zones, Zone{ID: len(m.Zones), Kind: kind, Shape: o.shape(),
				WindDir: propFloat(o.Properties, "dir", 0) * math.Pi / 180})
		case "gates":
			c := o.shape().Center()
			m.Gates = append(m.Gates, Gate{ID: len(m.Gates), X: c.X, Y: c.Y,
				Network: propString(o.Properties, "network", "a"), Color: propString(o.Properties, "color", "purple"),
				Heaven:  propString(o.Properties, "heaven", "") == "1",
				Dock:    propString(o.Properties, "dock", "") == "1",
				DockFor: int(propFloat(o.Properties, "dock_for", -1))})
		case "altars":
			sh := o.shape()
			c := sh.Center()
			m.Altars = append(m.Altars, Altar{ID: len(m.Altars), Kind: kind, X: c.X, Y: c.Y,
				R: math.Max(sh.W, sh.H) / 2})
		case "decor":
			c := o.shape().Center()
			m.Decor = append(m.Decor, Decor{Sprite: kind, X: c.X, Y: c.Y})
		default:
		}
	}
}

type tmjProp struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type tmjObject struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	Class   string  `json:"class"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
	Ellipse bool    `json:"ellipse"`
	Point   bool    `json:"point"`
	// Polygon: đỉnh tương đối so với (X, Y) — vùng địa hình hữu cơ của spec §8.6.
	Polygon    []tmjPoint `json:"polygon"`
	Properties []tmjProp  `json:"properties"`
}

type tmjPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type tmjLayer struct {
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	Objects []tmjObject `json:"objects"`
}

type tmjFile struct {
	Width      int        `json:"width"`
	Height     int        `json:"height"`
	TileWidth  int        `json:"tilewidth"`
	TileHeight int        `json:"tileheight"`
	Properties []tmjProp  `json:"properties"`
	Layers     []tmjLayer `json:"layers"`
}

// kind reads the Tiled class (1.9+ "class", older "type"), falling back to the name.
func (o tmjObject) kind() string {
	for _, k := range []string{o.Class, o.Type, o.Name} {
		if k != "" {
			return strings.ToLower(k)
		}
	}
	return ""
}

func (o tmjObject) shape() Shape {
	if len(o.Polygon) >= 3 {
		// Đa giác hữu cơ (spec §8.6). Đỉnh trong Tiled là toạ độ TƯƠNG ĐỐI so với gốc của
		// object, đổi sang tuyệt đối để server và client cùng một hệ.
		pts := make([]float64, 0, len(o.Polygon)*2)
		for _, pt := range o.Polygon {
			pts = append(pts, o.X+pt.X, o.Y+pt.Y)
		}
		var sh Shape
		sh.SetPoly(pts)
		return sh
	}
	if o.Point {
		return Shape{Circle: true, X: o.X, Y: o.Y, R: 1}
	}
	if o.Ellipse {
		r := math.Min(o.Width, o.Height) / 2
		return Shape{Circle: true, X: o.X + o.Width/2, Y: o.Y + o.Height/2, R: r}
	}
	return Shape{X: o.X, Y: o.Y, W: o.Width, H: o.Height}
}

func propString(ps []tmjProp, name, def string) string {
	for _, p := range ps {
		if p.Name == name {
			if s, ok := p.Value.(string); ok {
				return s
			}
		}
	}
	return def
}

func propFloat(ps []tmjProp, name string, def float64) float64 {
	for _, p := range ps {
		if p.Name == name {
			if f, ok := p.Value.(float64); ok {
				return f
			}
		}
	}
	return def
}
