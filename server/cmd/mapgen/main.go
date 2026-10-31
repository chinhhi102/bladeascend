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
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"kiemxoay/server/internal/config"
)

const tile = 100 // Tiled grid cell in world units

type prop struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type object struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
	Rotation float64 `json:"rotation"`
	Visible  bool    `json:"visible"`
	Ellipse  bool    `json:"ellipse,omitempty"`
	Point    bool    `json:"point,omitempty"`
	// Polygon: đỉnh TƯƠNG ĐỐI so với (X, Y) theo đúng định dạng Tiled (spec §8.6).
	Polygon    []point `json:"polygon,omitempty"`
	Properties []prop  `json:"properties,omitempty"`
}

type point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
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

// Spec §8.1: KHÔNG object nào được nằm ngoài bản đồ. Mọi vật cản / vùng / decor phải nằm
// trọn trong [mapMargin, size-mapMargin] — tính cả kích thước của chính nó.
const (
	mapMargin   float64 = 300 // lề bắt buộc quanh bản đồ
	borderThick float64 = 120 // bề dày tường biên không phá được
)

// lakeArea là một mặt hồ: hình tròn bị bẹp theo trục dọc, đúng hình `organicZone` sinh ra.
type lakeArea struct {
	cx, cy, r, squash float64
}

// contains: điểm này có nằm trên mặt hồ không (nới thêm `pad` ra ngoài mép).
func (l lakeArea) contains(x, y, pad float64) bool {
	dy := (y - l.cy) / math.Max(l.squash, 1e-6)
	dx := x - l.cx
	rr := l.r + pad
	return dx*dx+dy*dy <= rr*rr
}

// rect is a placed footprint used to keep obstacles apart.
type rect struct{ x0, y0, x1, y1 float64 }

// solidRect là một khối đã đặt, KÈM BẬC của nó.
//
// Phải nhớ bậc vì §8.1b tính hành lang theo `corridor_min[min(bậcA, bậcB)]`: khe giữa một
// tảng đá khổng lồ (bậc 9) và một bức tường gạch (bậc 2) chỉ cần 160, không cần 340 — ai
// chém được gạch thì đã đi xuyên qua rồi, khe chỉ phải lọt người KHÔNG chém được cả hai, mà
// người đó là cảnh giới 1.
//
// Dùng một con số hành lang chung cho mọi cặp là lý do bản đồ kẹt ở 2% mật độ: chỉ 2/70 ô
// lưới nhận được đảo, vì mỗi đảo bậc 9 đòi 340 trống quanh cả những viên gạch bậc 2.
type solidRect struct {
	r    rect
	tier int
}

func (a rect) overlaps(b rect, margin float64) bool {
	return a.x0-margin < b.x1 && b.x0-margin < a.x1 && a.y0-margin < b.y1 && b.y0-margin < a.y1
}

// builder accumulates objects per layer and rejects overlapping placements.
type builder struct {
	size   float64
	seed   int64
	rng    *rand.Rand
	nextID int
	layers map[string][]object
	solid  []solidRect
	keep   []rect // areas that must stay clear (spawn plaza)
	// cover: lưới 50 đánh dấu ô nào đã bị chân đế vật cản chiếm, cập nhật DẦN mỗi lần thêm
	// một khối. Có nó thì `footprintDensity` là một phép chia, không phải quét lại toàn bộ.
	//
	// Quét lại mỗi lần là 230² ô × 500 vật cản = 26 triệu phép so sánh, mà `islandsStep` gọi
	// nó tới 4 000 lần — tức 100 tỉ phép cho một khu.
	cover    []bool
	coverN   int
	coverLo  float64
	coverHit int
	// vis / visHit: cùng lưới nhưng đếm HÌNH VẼ thay vì chân đế. Mục tiêu mật độ của §8.1b
	// tính theo hình vẽ (spec v6.12), nên bộ sinh phải lặp theo đúng con số bộ kiểm đo.
	vis    []bool
	visHit int
	// air / airHit: ô CÁCH MỌI VẬT CẢN ít nhất `breathR`. Đếm ngược (đánh dấu ô GẦN vật
	// cản) nên `airHit` là số ô đã mất chỗ thở.
	//
	// Có nó thì bộ sinh biết DỪNG: khoảng thở là ràng buộc ngang hàng với mật độ, không
	// phải thứ đo sau. Thiếu nó thì vòng lặp cứ nhồi cho đủ mật độ rồi mới phát hiện đã
	// nuốt sạch chỗ trống — đo được mật độ 15% / khoảng thở 13%.
	air    []bool
	airHit int
	// footW/footH: tỉ lệ chân đế lấy từ `obstacles.footprint_w/h`. Giữ trong builder để
	// `addSolid` không phải nạp config mỗi lần.
	footW, footH float64
	// obs: bảng bậc + hành lang tối thiểu (§8.2).
	obs config.Obstacles
	// lakeKeep: mặt hồ — KHÔNG đặt vật cản lên (chủ dự án chốt 31/10: "mặt hồ gần như
	// trống, vật cản ngoài đảo ≤ 1%").
	//
	// Mặt hồ Băng Nguyệt Hồ là sân trượt băng để rượt nhau; đó là thứ làm khu ấy khác bốn
	// khu kia. Rải vật cản lên đó thì nó thành một khu như mọi khu, chỉ khác màu nền.
	lakeKeep []lakeArea
	// decorSkipped: vật trang trí bị chốt chặn trong `decor()` từ chối vì không có art.
	// Giữ lại để in ra; bỏ qua trong im lặng là đúng cái lỗi Q27-2.
	decorSkipped []string
}

func newBuilder(size float64, seed int64) *builder {
	c := size / 2
	b := &builder{size: size, seed: seed, rng: rand.New(rand.NewSource(seed)), nextID: 1, layers: map[string][]object{},
		keep: []rect{{c - 600, c - 600, c + 600, c + 600}}}
	lo, hi := b.playable()
	b.coverLo = lo
	b.coverN = int((hi - lo) / coverCell)
	b.cover = make([]bool, b.coverN*b.coverN)
	b.vis = make([]bool, b.coverN*b.coverN)
	b.air = make([]bool, b.coverN*b.coverN)
	b.obs = config.MustLoadDefault().Obstacles
	b.footW, b.footH = b.obs.FootprintW, b.obs.FootprintH
	return b
}

// coverCell phải BẰNG lưới của `cmd/mapgen --check`: hai nơi đo cùng một con số, lệch lưới
// là mapgen tưởng đã đủ mật độ trong khi bộ kiểm báo còn trống.
const coverCell = 50.0

// addSolid ghi một khối vào danh sách chặn VÀ cập nhật lưới phủ.
//
// Mọi chỗ thêm vật cản phải đi qua đây. Thêm thẳng vào `b.solid` thì lưới phủ lệch, và
// `islandsStep` quay mãi vì mật độ không bao giờ tăng.
//
// Lưới phủ đếm **CHÂN ĐẾ**, không phải hình vẽ. Hai thứ chênh nhau hơn ba lần: chân đế mặc
// định rộng 90% và cao 30% hình vẽ, nằm ở đáy (`world.Footprint`). Đếm hình vẽ thì mapgen
// báo 13,4% trong khi `--check` đo 6,3% — hai con số của cùng một bản đồ, và mapgen dừng
// sớm vì tưởng đã đủ.
//
// Ô tính là bị phủ khi **tâm ô** nằm trong chân đế, đúng phép của `--check`. Lấy cả dải
// chỉ số x0..x1 thì một khối rộng 140 chiếm 3–4 ô (150–200 đơn vị), dôi ra 40%.
func (b *builder) addSolid(r rect, tier int) {
	b.solid = append(b.solid, solidRect{r, tier})
	b.stampCover(r)
}

// stampVisual đánh dấu ô mà HÌNH VẼ của `r` phủ lên (không co về chân đế), và đồng thời
// đánh dấu vành `breathR` quanh nó là "mất chỗ thở".
func (b *builder) stampVisual(r rect) {
	b.stampGrid(b.vis, &b.visHit, r)
	b.stampGrid(b.air, &b.airHit, rect{r.x0 - breathR, r.y0 - breathR, r.x1 + breathR, r.y1 + breathR})
}

// breathing: phần diện tích khu cách MỌI vật cản ít nhất `breathR` (§8.1b mục 6).
func (b *builder) breathing() float64 {
	if b.coverN == 0 {
		return 1
	}
	return 1 - float64(b.airHit)/float64(b.coverN*b.coverN)
}

// stampCover đánh dấu các ô lưới mà CHÂN ĐẾ của `r` phủ lên.
func (b *builder) stampCover(r rect) {
	b.stampGrid(b.cover, &b.coverHit, footprintRect(r, b.footW, b.footH))
	b.stampVisual(r)
}

// stampGrid đánh dấu ô có TÂM nằm trong `f`, đúng phép của `cmd/mapgen --check`.
func (b *builder) stampGrid(grid []bool, hits *int, f rect) {
	y0 := int((f.y0 - b.coverLo) / coverCell)
	y1 := int((f.y1-b.coverLo)/coverCell) + 1
	x0 := int((f.x0 - b.coverLo) / coverCell)
	x1 := int((f.x1-b.coverLo)/coverCell) + 1
	for y := max(y0, 0); y <= min(y1, b.coverN-1); y++ {
		cy := b.coverLo + (float64(y)+0.5)*coverCell
		if cy < f.y0 || cy > f.y1 {
			continue
		}
		for x := max(x0, 0); x <= min(x1, b.coverN-1); x++ {
			cx := b.coverLo + (float64(x)+0.5)*coverCell
			if cx < f.x0 || cx > f.x1 {
				continue
			}
			if i := y*b.coverN + x; !grid[i] {
				grid[i] = true
				*hits++
			}
		}
	}
}

// footprintRect là bản `world.Footprint` cho `rect`: rộng `fw` phần, cao `fh` phần, nằm ở
// ĐÁY hình. Viết lại ở đây thay vì gọi `world` vì mapgen làm việc trên `rect` chứ không
// `world.Shape`; hai công thức phải khớp, và `TestMapgenDensityMatchesCheck` canh điều đó.
func footprintRect(r rect, fw, fh float64) rect {
	w, h := r.x1-r.x0, r.y1-r.y0
	nw, nh := w*fw, h*fh
	x := r.x0 + (w-nw)/2
	return rect{x, r.y1 - nh, x + nw, r.y1}
}

// footprintDensity: phần diện tích khu chơi được bị chân đế vật cản chiếm.
// removeSolid xoá một vật cản khỏi cả `solid` lẫn lớp `obstacles`, rồi dựng lại lưới phủ.
//
// Hai mảng đó SONG SONG: mọi đường thêm vật cản (`obstacle`, `obstacleForce`, `boulder`,
// `wall`) đều ghi vào cả hai, cùng thứ tự. Xoá lệch một bên là bản đồ có vật cản vô hình
// (chặn mà không vẽ) hoặc vật cản ma (vẽ mà không chặn) — cả hai đều im lặng.
func (b *builder) removeSolid(i int) {
	if i < 0 || i >= len(b.solid) {
		return
	}
	b.solid = append(b.solid[:i], b.solid[i+1:]...)
	obs := b.layers["obstacles"]
	if i < len(obs) {
		b.layers["obstacles"] = append(obs[:i], obs[i+1:]...)
	}
	b.rebuildCover()
}

// rebuildCover dựng lại lưới phủ từ đầu. Chỉ gọi sau khi XOÁ — lưới cộng dồn không trừ
// ngược được (một ô có thể do nhiều khối cùng phủ).
func (b *builder) rebuildCover() {
	for i := range b.cover {
		b.cover[i] = false
		b.vis[i] = false
		b.air[i] = false
	}
	b.coverHit, b.visHit, b.airHit = 0, 0, 0
	for _, s := range b.solid {
		b.stampCover(s.r)
	}
}

// visualDensity: phần diện tích khu bị HÌNH VẼ vật cản phủ — con số §8.1b mục 5 nói tới
// (18–28%). Khác `footprintDensity` (chân đế) đúng tỉ lệ chân đế / hình vẽ ≈ 27%.
func (b *builder) visualDensity() float64 {
	if b.coverN == 0 {
		return 0
	}
	return float64(b.visHit) / float64(b.coverN*b.coverN)
}

func (b *builder) footprintDensity() float64 {
	if b.coverN == 0 {
		return 0
	}
	return float64(b.coverHit) / float64(b.coverN*b.coverN)
}

// tierOf tra bậc của một loại vật cản từ config — không viết số vào mã.
func (b *builder) tierOf(kind string) int { return b.obs.ObstacleTier(kind) }

func (b *builder) add(layerName string, o object) {
	o.ID = b.nextID
	o.Visible = true
	b.nextID++
	b.layers[layerName] = append(b.layers[layerName], o)
}

// inBounds is the §8.1 rule on its own so every placement path can share it.
func (b *builder) inBounds(r rect) bool {
	return r.x0 >= mapMargin && r.y0 >= mapMargin && r.x1 <= b.size-mapMargin && r.y1 <= b.size-mapMargin
}

// playable is the area left for random placement: inside the border wall.
func (b *builder) playable() (lo, hi float64) {
	return mapMargin + borderThick, b.size - mapMargin - borderThick
}

// onLake: hình này có chạm mặt hồ không. Lùi vào 120 so với mép hồ để bờ vẫn đặt được —
// "dồn vật cản lên đảo và bờ".
func (b *builder) onLake(r rect) bool {
	for _, l := range b.lakeKeep {
		if l.contains((r.x0+r.x1)/2, (r.y0+r.y1)/2, -120) {
			return true
		}
	}
	return false
}

// nearSolid: có vật cản nào trong bán kính `r` quanh điểm này không.
//
// Dùng để BẮT nội dung rải rác của khu bám vào cụm thay vì trải đều (§8.1b mục 6). Nhà,
// rào, sạp, bãi đá — tất cả đều đi qua đây, nên chỉ cần một chốt là cả năm khu đổi nết.
//
// Không có chốt này thì dù cụm có gọn tới đâu, lớp rải rác vẫn phủ kín khoảng trống giữa
// các cụm: đo được khoảng thở 11–29% trong khi cụm một mình cho 45–60%.
func (b *builder) nearSolid(x, y, r float64) bool {
	rr := rect{x - r, y - r, x + r, y + r}
	for _, s := range b.solid {
		if s.tier >= config.TierBorder {
			continue // biên khu không tính là "có cụm ở đây"
		}
		if rr.overlaps(s.r, 0) {
			return true
		}
	}
	return false
}

func (b *builder) free(r rect, margin float64) bool {
	if !b.inBounds(r) || b.onLake(r) {
		return false
	}
	for _, k := range b.keep {
		if r.overlaps(k, 0) {
			return false
		}
	}
	for _, s := range b.solid {
		if r.overlaps(s.r, margin) {
			return false
		}
	}
	return true
}

// freeFor là `free` có xét BẬC: khe với mỗi khối đã đặt lấy theo `corridor_min[min(bậc)]`
// (§8.1b), thay vì một con số chung.
func (b *builder) freeFor(r rect, tier int, corridor func(int) float64) bool {
	if !b.inBounds(r) || b.onLake(r) {
		return false
	}
	for _, k := range b.keep {
		if r.overlaps(k, 0) {
			return false
		}
	}
	for _, s := range b.solid {
		if r.overlaps(s.r, corridor(min(tier, s.tier))) {
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
	b.addSolid(r, b.tierOf(kind))
	b.add("obstacles", object{Type: kind, X: x, Y: y, Width: w, Height: h})
	return true
}

// obstacleForce đặt một khối KHÔNG kiểm chỗ trống.
//
// Chỉ dùng cho các khối BÊN TRONG một đảo, sau khi `freeFor` đã duyệt cả hộp bao của đảo
// (xem `placeIsland`). Trong một đảo các khối phải chồng mép lên nhau mới thành một cụm
// đặc; đi qua `obstacle` thì `free(r, 0)` từ chối ngay vì chồng mép LÀ chồng, và kết quả là
// mỗi đảo chỉ còn đúng cái lõi — đo được: 81 đảo đặt xuống mà chỉ thêm 81 vật cản.
func (b *builder) obstacleForce(kind string, x, y, w, h float64) {
	r := rect{x, y, x + w, y + h}
	if !b.inBounds(r) || b.onLake(r) {
		return
	}
	b.addSolid(r, b.tierOf(kind))
	b.add("obstacles", object{Type: kind, X: x, Y: y, Width: w, Height: h})
}

func (b *builder) boulder(cx, cy, r, margin float64) bool {
	rr := rect{cx - r, cy - r, cx + r, cy + r}
	if !b.free(rr, margin) {
		return false
	}
	b.addSolid(rr, b.tierOf("boulder"))
	b.add("obstacles", object{Type: "boulder", X: cx - r, Y: cy - r, Width: 2 * r, Height: 2 * r, Ellipse: true})
	return true
}

// wall forces a wall segment through, ignoring the usual spacing rules (it still may not
// leave the map). Dùng cho biên khu, hẻm, vách núi.
//
// `kind` quyết định BẬC chém được (§8.2). Trước v6.7 mọi tường ở đây là `stone_wall` và
// được coi là "không phá được" vì bản cũ gán `MaxHP: 0` cho loại này. Giờ `stone_wall` có
// bậc 5, nên để nguyên là từ Luyện Hư trở lên **chém thủng cả biên khu và đi ra ngoài bản
// đồ**. Mọi chỗ gọi phải nói rõ loại nào.
func (b *builder) wall(kind string, x, y, w, h float64) bool {
	r := rect{x, y, x + w, y + h}
	if !b.inBounds(r) {
		return false
	}
	// Tường cũng phải tránh vùng giữ trống (quanh cổng, đài): hẻm sinh sau cổng, không
	// kiểm thì sẽ mọc đè lên cổng và phá luật "trống 300 quanh cổng" của §8.1 bước 8.
	for _, k := range b.keep {
		if r.overlaps(k, 0) {
			return false
		}
	}
	b.addSolid(r, b.tierOf(kind))
	b.add("obstacles", object{Type: kind, X: x, Y: y, Width: w, Height: h})
	return true
}

// border draws the unbreakable wall ring around the map (spec §8.1). The four runs overlap at
// the corners so the ring never has a gap to slip through.
func (b *builder) border() {
	lo, hi := mapMargin, b.size-mapMargin
	span := hi - lo
	b.wall(config.BorderKind, lo, lo, span, borderThick)             // trên
	b.wall(config.BorderKind, lo, hi-borderThick, span, borderThick) // dưới
	b.wall(config.BorderKind, lo, lo, borderThick, span)             // trái
	b.wall(config.BorderKind, hi-borderThick, lo, borderThick, span) // phải
}

func (b *builder) zone(kind string, x, y, w, h float64, ellipse bool, props ...prop) bool {
	if !b.inBounds(rect{x, y, x + w, y + h}) {
		return false
	}
	b.add("zones", object{Type: kind, X: x, Y: y, Width: w, Height: h, Ellipse: ellipse, Properties: props})
	return true
}

func (b *builder) gate(x, y float64, network, color string) bool {
	if !b.inBounds(rect{x - 150, y - 150, x + 150, y + 150}) {
		return false
	}
	b.keep = append(b.keep, rect{x - 300, y - 300, x + 300, y + 300})
	b.add("gates", object{Type: "gate", X: x - 150, Y: y - 150, Width: 300, Height: 300, Ellipse: true,
		Properties: []prop{{"network", "string", network}, {"color", "string", color}}})
	return true
}

// heavenGate đặt **Thiên Môn** (spec §8.4a): to gấp đôi cổng thường, chừa trống bán kính
// 500 quanh nó — spec đòi một quảng trường ở tâm khu, có ít nhất 2 đường dẫn vào.
func (b *builder) heavenGate(x, y float64) bool {
	const r = 200.0
	if !b.inBounds(rect{x - r, y - r, x + r, y + r}) {
		return false
	}
	b.keep = append(b.keep, rect{x - 500, y - 500, x + 500, y + 500})
	b.add("gates", object{Type: "gate", X: x - r, Y: y - r, Width: 2 * r, Height: 2 * r, Ellipse: true,
		Properties: []prop{{"network", "string", gateNet}, {"color", "string", "gold"}, {"heaven", "string", "1"}}})
	return true
}

// heavenDock đặt **Bến Thiên Môn** ở Thiên Đình — điểm ra cố định của Thiên Môn khu `zone`.
// Spec §8.4a: cách tâm ~1 400, trống bán kính 400.
func (b *builder) heavenDock(x, y float64, zone int) bool {
	const r = 150.0
	if !b.inBounds(rect{x - r, y - r, x + r, y + r}) {
		return false
	}
	b.keep = append(b.keep, rect{x - 400, y - 400, x + 400, y + 400})
	b.add("gates", object{Type: "gate", X: x - r, Y: y - r, Width: 2 * r, Height: 2 * r, Ellipse: true,
		Properties: []prop{{"network", "string", gateNet}, {"color", "string", "gold"},
			{"dock", "string", "1"}, {"dock_for", "float", float64(zone)}}})
	return true
}

// altar places a Đài Tiên Khí and keeps a clear_r radius around it free of obstacles
// (spec §8.5). Kind is "small" or "center".
func (b *builder) altar(kind string, cx, cy, r, clear float64) bool {
	if !b.inBounds(rect{cx - r, cy - r, cx + r, cy + r}) {
		return false
	}
	b.keep = append(b.keep, rect{cx - clear, cy - clear, cx + clear, cy + clear})
	b.add("altars", object{Type: kind, X: cx - r, Y: cy - r, Width: 2 * r, Height: 2 * r, Ellipse: true})
	return true
}

// decor rải `n` lần thử một vật trang trí. Chỉ đặt được ở chỗ trống.
//
// Chốt chặn cuối: **không đặt vật trang trí không có art** (Q27-2). `decorStepFor` đã lọc
// rồi, nhưng ba hàm sinh bản đồ cũ (`langTre`, `choDem`, `nuiDa`) gọi thẳng vào đây, và
// chính kiểu "gọi thẳng, gõ tên tay" là cách 543 vật vô hình lọt vào bản đồ. Chặn ở nơi
// ĐẶT thì mọi đường gọi đều an toàn, kể cả đường viết sau này.
func (b *builder) decor(sprite string, n int) {
	if !haveDecorArt()[sprite] {
		b.decorSkipped = append(b.decorSkipped, sprite)
		return
	}
	lo, hi := b.playable()
	for i := 0; i < n; i++ {
		x, y := b.rnd(lo+60, hi-60), b.rnd(lo+60, hi-60)
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
		lo, hi := b.playable()
		if fn(b.rnd(lo, hi), b.rnd(lo, hi)) {
			placed++
		}
	}
}

// alley builds a corridor between two brick walls with an opening of the given width
// (spec §8.3 hẻm hẹp). Người to không lọt qua hẻm hẹp.
//
// Tường hẻm là **gạch (bậc 2)**, không phải bất khả phá: §8.1b buộc khe giữa hai vật cản
// bậc A, B phải ≥ `corridor_min[min(A,B)]`, và với bậc 2 con số đó là 160. Bậc cao hơn đòi
// hành lang rộng hơn, tức "hẻm hẹp" sẽ không còn hẹp. Kim Đan trở lên chém được tường hẻm
// — đúng ý §8.2: người to không chui được thì chém, đó là lựa chọn, không phải bế tắc.
func (b *builder) alley(x, y, length float64, gap float64, vertical bool) {
	thick := 120.0
	if vertical {
		b.wall("brick", x, y, thick, length)
		b.wall("brick", x+thick+gap, y, thick, length)
		return
	}
	b.wall("brick", x, y, length, thick)
	b.wall("brick", x, y+thick+gap, length, thick)
}

// altars places the five Đài Tiên Khí (spec §8.5): four by the sect yards near the corners
// and the Thiên Tiên Đài in the middle. Must run BEFORE obstacles so clear_r stays empty.
func (b *builder) placeAltars(clear float64) {
	c := b.size / 2
	b.altar("center", c, c, 220, clear)
	d := b.size * 0.25
	for _, p := range [][2]float64{{d, d}, {b.size - d, d}, {d, b.size - d}, {b.size - d, b.size - d}} {
		b.altar("small", p[0], p[1], 150, clear)
	}
}

// validate is the §8.1 check that must pass before a map is written: nothing may stick out
// of [mapMargin, size-mapMargin]. It reports every offender so a bad generator is obvious.
func (b *builder) validate() error {
	var bad []string
	for _, ln := range []string{"zones", "obstacles", "gates", "decor"} {
		for _, o := range b.layers[ln] {
			r := rect{o.X, o.Y, o.X + o.Width, o.Y + o.Height}
			if o.Point {
				r = rect{o.X, o.Y, o.X, o.Y}
			}
			if !b.inBounds(r) {
				bad = append(bad, fmt.Sprintf("%s/%s tại (%.0f,%.0f)-(%.0f,%.0f)", ln, o.Type, r.x0, r.y0, r.x1, r.y1))
			}
		}
	}
	if len(bad) > 0 {
		return fmt.Errorf("%d object nằm ngoài [%g, %g]: %s", len(bad), mapMargin, b.size-mapMargin, strings.Join(bad, "; "))
	}
	return nil
}

// windLane is a long strip; dir in degrees (0 = east, 90 = south).
func (b *builder) windLane(x, y, w, h, dirDeg float64) bool {
	return b.zone("wind", x, y, w, h, false, prop{"dir", "float", dirDeg})
}

// zonePoly đặt một vùng địa hình hình HỮU CƠ (spec §8.6). Đỉnh đưa vào là toạ độ tuyệt đối;
// Tiled lưu tương đối nên trừ đi gốc trước khi ghi.
func (b *builder) zonePoly(kind string, pts []float64, props ...prop) bool {
	if len(pts) < 6 {
		return false
	}
	bb := polyBounds(pts)
	if !b.inBounds(bb) {
		return false
	}
	rel := make([]point, 0, len(pts)/2)
	for i := 0; i < len(pts); i += 2 {
		rel = append(rel, point{X: pts[i] - bb.x0, Y: pts[i+1] - bb.y0})
	}
	b.add("zones", object{Type: kind, X: bb.x0, Y: bb.y0, Width: bb.x1 - bb.x0, Height: bb.y1 - bb.y0,
		Polygon: rel, Properties: props})
	return true
}

// organicZone là lối tắt hay dùng: vùng méo quanh một tâm.
func (b *builder) organicZone(kind string, cx, cy, r float64, squash float64, props ...prop) bool {
	// Hồ băng / hồ nước ĐỦ LỚN thì ghi lại làm vùng cấm đặt vật cản. Ghi ở đây chứ không
	// để khu tự nhớ: `organicZone` là cửa duy nhất sinh ra mặt hồ, nên không sót được.
	if (kind == "ice" || kind == "water") && r > 1500 {
		b.lakeKeep = append(b.lakeKeep, lakeArea{cx, cy, r, squash})
	}
	return b.zonePoly(kind, organicPoly(cx, cy, r, 14, squash, b.seed+int64(len(b.layers["zones"]))), props...)
}

func (b *builder) write(path, name, biome string) error {
	if err := b.validate(); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	n := int(b.size / tile)
	order := []string{"zones", "obstacles", "gates", "altars", "decor"}
	var layers []layer
	for i, ln := range order {
		layers = append(layers, layer{ID: i + 1, Name: ln, Type: "objectgroup", DrawOrder: "topdown", Opacity: 1, Visible: true,
			Objects: append([]object{}, b.layers[ln]...)})
	}
	m := tmj{Type: "map", Version: "1.10", TiledVersion: "1.10.2", Orientation: "orthogonal", RenderOrder: "right-down",
		Width: n, Height: n, TileWidth: tile, TileHeight: tile, NextLayerID: len(order) + 1, NextObjectID: b.nextID,
		Properties: []prop{{"name", "string", name}, {"biome", "string", biome}, {"play_margin", "float", mapMargin + borderThick}}, Layers: layers, Tilesets: []any{}}
	data, err := json.MarshalIndent(m, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func main() {
	out := flag.String("out", config.FindRepoFile("maps"), "output directory")
	seed := flag.Int64("seed", 0, "seed; 0 = seed mặc định theo từng khu")
	zone := flag.Int("zone", -1, "chỉ sinh khu này (0..4); -1 = sinh cả 5 khu")
	png := flag.String("png", "", "xuất ảnh xem trước ra file này (chỉ dùng với --zone)")
	// --check KHÔNG sinh bản đồ: nó kiểm bản đồ ĐANG CÓ trong `maps/` (spec §8.1b). Kiểm
	// cái đã ghi ra đĩa chứ không cái vừa dựng trong bộ nhớ, vì đó mới là cái người chơi
	// nhận — và vì CI phải bắt được cả trường hợp ai đó sửa .tmj bằng Tiled.
	check := flag.Bool("check", false, "kiểm bố cục 5 khu × 14 cảnh giới theo §8.1b, không sinh gì")
	checkPNG := flag.String("check-png", "", "xuất ảnh bộ kiểm vào thư mục này (vd docs/map_check)")
	checkRealmsArg := flag.String("check-realms", "", "chỉ xuất ảnh cho các cảnh giới này (vd \"0,4,8\"); rỗng = đủ 14")
	flag.Parse()
	cfg := config.MustLoadDefault()
	if *check {
		dir := *checkPNG
		if dir == "" && *checkRealmsArg != "" {
			dir = config.FindRepoFile("docs/map_check")
		}
		if err := runCheck(*out, dir, cfg, parseRealms(*checkRealmsArg)); err != nil {
			log.Fatal(err)
		}
		return
	}
	size := cfg.Room.MapSize
	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatal(err)
	}

	want := make([]int, 0, len(zoneDefs))
	if *zone >= 0 {
		if *zone >= len(zoneDefs) {
			log.Fatalf("khu %d không có (chỉ có 0..%d)", *zone, len(zoneDefs)-1)
		}
		want = append(want, *zone)
	} else {
		for i := range zoneDefs {
			want = append(want, i)
		}
	}
	if *png != "" && len(want) != 1 {
		log.Fatal("--png phải đi kèm --zone (mỗi lần xem trước một khu)")
	}

	for _, zi := range want {
		zd := zoneDefs[zi]
		base := int64(1000 + zi)
		if *seed != 0 {
			base = *seed
		}
		b, used, err := buildZoneWithRetries(size, base, cfg.Mapgen, zi)
		if err != nil {
			log.Fatalf("khu %d (%s): %v", zi, zd.name, err)
		}
		path := filepath.Join(*out, zd.id+".tmj")
		if err := b.write(path, zd.name, zd.biome); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("khu %d %s (seed %d): %d vật cản, %d vùng, %d cổng, %d đài, %d trang trí → %s\n",
			zi, zd.name, used, len(b.layers["obstacles"]), len(b.layers["zones"]),
			len(b.layers["gates"]), len(b.layers["altars"]), len(b.layers["decor"]), path)
		// Vật trang trí hợp chủ đề khu này mà chưa có art: NÓI RA. Trước đây mapgen cứ đặt
		// rồi client bỏ qua trong im lặng, và 543 vật vô hình nằm trong bản đồ nhiều tuần.
		if miss := uniqueSorted(b.decorSkipped); len(miss) > 0 {
			fmt.Printf("   ⚠ chưa đặt %d loại trang trí vì thiếu art: %s\n", len(miss), strings.Join(miss, ", "))
			// In ĐÚNG tên job artgen. Ghép `"decor_" + tên` cho ra `decor_bamboo` — một job
			// không tồn tại, và một lệnh gợi ý sai thì tệ hơn là không gợi ý gì.
			fmt.Printf("     sinh art: cd server && go run ./cmd/artgen run -only %s\n", strings.Join(decorJobsFor(miss), " -only "))
		}
		if *png != "" {
			if err := writePreview(b, *png); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("ảnh xem trước → %s\n", *png)
		}
	}
}

// buildZoneWithRetries sinh một khu, thử lại với seed+1 khi kiểm tra không đạt (§8.1 bước 8).
func buildZoneWithRetries(size float64, seed int64, mg config.Mapgen, zone int) (*builder, int64, error) {
	tries := mg.MaxRetries
	if tries <= 0 {
		tries = 10
	}
	var last error
	for i := 0; i < tries; i++ {
		b := newBuilder(size, seed+int64(i))
		if err := generateZone(b, mg, zone); err != nil {
			last = err
			continue
		}
		return b, seed + int64(i), nil
	}
	return nil, 0, fmt.Errorf("sinh thất bại sau %d lần (lỗi cuối: %w)", tries, last)
}

// buildWithRetries runs the §8.1 generator, retrying with seed+1 when validation fails
// (bước 8: "sai thì sinh lại với seed+1, tối đa 10 lần").
func buildWithRetries(size float64, seed int64, mg config.Mapgen) (*builder, int64, error) {
	tries := mg.MaxRetries
	if tries <= 0 {
		tries = 10
	}
	var last error
	for i := 0; i < tries; i++ {
		b := newBuilder(size, seed+int64(i))
		if err := generate(b, mg); err != nil {
			last = err
			continue
		}
		return b, seed + int64(i), nil
	}
	return nil, 0, fmt.Errorf("sinh bản đồ thất bại sau %d lần (lỗi cuối: %w)", tries, last)
}

// langTre: balanced village — fences, houses, ponds of mud, bushes, two gate pairs.
func langTre(b *builder) {
	s := b.size
	b.border()         // tường biên không phá được (§8.1)
	b.placeAltars(300) // 5 Đài Tiên Khí (§8.5), đặt trước vật cản để giữ trống quanh đài
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
	b.border()         // tường biên không phá được (§8.1)
	b.placeAltars(300) // 5 Đài Tiên Khí (§8.5), đặt trước vật cản để giữ trống quanh đài
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
	b.border()         // tường biên không phá được (§8.1)
	b.placeAltars(300) // 5 Đài Tiên Khí (§8.5), đặt trước vật cản để giữ trống quanh đài
	net := [][2]float64{{1800, 1800}, {s - 1800, 1800}, {s - 1800, s - 1800}, {1800, s - 1800}}
	for _, p := range net {
		b.gate(p[0], p[1], "ring", "orange")
	}
	b.gate(s/2, 2600, "c", "blue")
	b.gate(s/2, s-2600, "c", "blue")
	b.scatter(45, func(x, y float64) bool { return b.boulder(x, y, b.rnd(250, 600), 300) })
	b.scatter(14, func(x, y float64) bool { // ridges
		if b.rng.Intn(2) == 0 {
			b.wall("cliff", x, y, b.rnd(900, 1600), 140)
		} else {
			b.wall("cliff", x, y, 140, b.rnd(900, 1600))
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

// parseRealms đọc "0,4,8" thành []int{0,4,8}. Rỗng → nil, nghĩa là đủ 14 cảnh giới.
func parseRealms(s string) []int {
	var out []int
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, err := strconv.Atoi(part)
		if err != nil {
			log.Fatalf("--check-realms: %q không phải số", part)
		}
		out = append(out, n)
	}
	return out
}
