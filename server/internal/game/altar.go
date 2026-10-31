package game

import (
	"math"

	"bladeascend/server/internal/geom"
	"bladeascend/server/internal/world"
)

// Đài Tiên Khí (spec §8.5): five fountains that spray Tiên khí orbs. Four small ones sit by
// the sect yards near the corners; the Thiên Tiên Đài in the centre drops one huge orb every
// 150s and announces it to the whole room 10s early. The orbs are the only source of Tiên khí
// (§3.4), so everyone — not just Tiên giới players — has a reason to fight over them.

// ItemTienKhi is the item kind of a Tiên khí orb; Value is how many points it carries.
const ItemTienKhi = "tien_khi"

// altarState tracks one altar's timers.
type altarState struct {
	a         world.Altar
	nextSpray float64 // đợt phun thường tiếp theo
	nextBurst float64 // đợt phun lớn tiếp theo
	warned    bool    // đã báo trước đợt lớn chưa
}

// resetAltars rebuilds the altar timers from the map (called when a game starts).
func (g *Game) resetAltars() {
	ac := g.Cfg.Altars
	g.altars = g.altars[:0]
	for _, a := range g.Map.Altars {
		st := altarState{a: a}
		if a.Kind == "center" {
			st.nextBurst = g.Time + ac.Center.EverySecs
		} else {
			st.nextSpray = g.Time + ac.Small.EverySecs
			st.nextBurst = g.Time + ac.Small.BurstEverySecs
		}
		g.altars = append(g.altars, st)
	}
}

// updateAltars runs the spray timers (spec §8.5).
func (g *Game) updateAltars() {
	ac := g.Cfg.Altars
	for i := range g.altars {
		st := &g.altars[i]
		kind := ac.Small
		if st.a.Kind == "center" {
			kind = ac.Center
		}
		// Báo trước đợt lớn: đài nhỏ rung 3s, Thiên Tiên Đài loan toàn phòng 10s.
		if !st.warned && st.nextBurst > 0 && g.Time >= st.nextBurst-kind.WarnSecs {
			st.warned = true
			g.emit(Event{Kind: "AltarWarn", A: uint32(st.a.ID), S: st.a.Kind,
				N: int(kind.WarnSecs), Global: st.a.Kind == "center",
				X: float32(st.a.X), Y: float32(st.a.Y)})
		}
		if st.a.Kind != "center" && g.Time >= st.nextSpray {
			st.nextSpray = g.Time + kind.EverySecs
			g.sprayTienKhi(st.a, g.rndInt(kind.Orbs), kind.OrbValue)
		}
		if st.nextBurst > 0 && g.Time >= st.nextBurst {
			st.warned = false
			if st.a.Kind == "center" {
				st.nextBurst = g.Time + kind.EverySecs
				g.sprayTienKhi(st.a, 1, kind.OrbValue)
			} else {
				st.nextBurst = g.Time + kind.BurstEverySecs
				g.sprayTienKhi(st.a, kind.BurstSmall, kind.OrbValue)
				g.sprayTienKhi(st.a, kind.BurstMedium, kind.MediumValue)
			}
			g.emit(Event{Kind: "AltarBurst", A: uint32(st.a.ID), S: st.a.Kind,
				Global: st.a.Kind == "center", X: float32(st.a.X), Y: float32(st.a.Y)})
		}
	}
}

// rndInt picks from a [min, max] pair in config; a single value means a fixed count.
func (g *Game) rndInt(r []int) int {
	switch len(r) {
	case 0:
		return 1
	case 1:
		return r[0]
	default:
		lo, hi := r[0], r[1]
		if hi <= lo {
			return lo
		}
		return lo + g.Rng.Intn(hi-lo+1)
	}
}

// sprayTienKhi throws n orbs out of the altar like a fountain: random angle, radius
// spray_r[0]..spray_r[1] (spec §8.5).
func (g *Game) sprayTienKhi(a world.Altar, n int, value float64) {
	ac := g.Cfg.Altars
	lo, hi := 300.0, 700.0
	if len(ac.SprayR) >= 2 {
		lo, hi = ac.SprayR[0], ac.SprayR[1]
	}
	for i := 0; i < n; i++ {
		ang := g.Rng.Float64() * 2 * math.Pi
		d := lo + g.Rng.Float64()*(hi-lo)
		g.addTienKhiOrb(Vec{X: a.X, Y: a.Y}.Add(geom.FromAngle(ang).Mul(d)), value)
	}
}

// tienOrbUnits là mệnh giá viên Tiên khí rơi ra, LỚN TRƯỚC (spec v6.4 §3.4).
//
// Thêm 500 và 100 vì Tiên khí đã ×20: Thái Ất 9 000 bị chém rơi 2 099 điểm. Với bộ cũ
// {10, 2, 1} thì đó là **210 viên** rải quanh một cái xác — vừa là một rừng vật phẩm che
// kín màn hình, vừa là 210 entity nhồi vào gói snapshot.
var tienOrbUnits = []float64{500, 100, 10, 2, 1}

// tienOrbMax là trần số viên mỗi lần rơi (spec v6.4 §3.4: "tối đa 40 viên").
const tienOrbMax = 40

// splitTienOrbs chia `total` thành mệnh giá, lớn trước, **không quá `tienOrbMax` viên**.
//
// Vượt trần thì gộp phần nhỏ vào MỘT viên (spec: "nếu vượt thì gộp phần nhỏ vào viên lớn
// hơn"), nên tổng điểm rơi ra **luôn đúng bằng** `floor(total)` — đó là thứ test khoá lại.
// Gộp bằng cách giữ 39 viên lớn nhất rồi dồn tất cả phần còn lại vào viên thứ 40.
//
// Tách thành hàm thuần (không đụng `g`) để test được bằng bảng, không cần dựng cả trận.
func splitTienOrbs(total float64) []float64 {
	left := math.Floor(total)
	if left <= 0 {
		return nil
	}
	out := make([]float64, 0, tienOrbMax)
	for _, unit := range tienOrbUnits {
		for left >= unit {
			out = append(out, unit)
			left -= unit
		}
	}
	if len(out) <= tienOrbMax {
		return out
	}
	// `out` đã giảm dần (mệnh giá lớn trước), nên phần đuôi chính là phần nhỏ nhất.
	rest := 0.0
	for _, v := range out[tienOrbMax-1:] {
		rest += v
	}
	return append(out[:tienOrbMax-1:tienOrbMax-1], rest)
}

// scatterTienKhi drops robbed Tiên khí around a victim (spec §3.4), splitting the total into
// denominations so a single pickup cannot take it all back.
//
// Bán kính là chỗ đã hỏng và khó thấy: bản cũ dùng `items.scatter_r` = 90, trong khi người
// Tiên giới có scale 5–6 nên thân rộng hơn 150 và cao hơn 400 pixel — mà vật phẩm vẽ DƯỚI
// nhân vật (`zIndex = y − 1000`). Kết quả: viên rơi đúng, gửi đúng, nhưng nằm khuất hẳn
// dưới cái xác to đùng, nên chủ dự án chém xong "không thấy Tiên khí rơi ra".
//
// Test khoá lại phần SỐ LIỆU (tổng điểm rơi); phần nhìn thấy được thì bán kính lo.
func (g *Game) scatterTienKhi(at Vec, total float64) {
	c := g.Cfg.Cultivation
	lo := math.Max(1, c.TienDropScatterMin)
	hi := math.Max(lo, c.TienDropScatterMax)
	for _, unit := range splitTienOrbs(total) {
		ang := g.Rng.Float64() * 2 * math.Pi
		d := lo + g.Rng.Float64()*(hi-lo)
		g.addTienKhiOrb(at.Add(geom.FromAngle(ang).Mul(d)), unit)
	}
}

// addTienKhiOrb places one orb, clamped inside the map, with a lifetime (§8.5: 60s then it
// fades; Tiên khí never respawns somewhere else).
func (g *Game) addTienKhiOrb(at Vec, value float64) {
	if value <= 0 {
		return
	}
	at = g.clampToMap(at)
	it := g.addItem(ItemTienKhi, value, at)
	it.ExpireAt = g.Time + math.Max(1, g.Cfg.Altars.OrbLifeSec)
	g.noteExpiring(it)
}

// clampToMap keeps a point inside the playable area so orbs never land behind the border wall.
func (g *Game) clampToMap(p Vec) Vec {
	lo, hi := g.PlayBounds(g.Cfg.Items.ItemR + 20)
	return Vec{X: geom.Clamp(p.X, lo, hi), Y: geom.Clamp(p.Y, lo, hi)}
}

// expireItems removes Tiên khí orbs whose lifetime ran out (spec §8.5).
//
// Chỉ duyệt DANH SÁCH VIÊN CÓ HẠN, không duyệt cả bản đồ vật phẩm. Một phòng có hàng nghìn
// viên Linh khí vĩnh viễn và chỉ dăm chục viên Tiên khí có hạn, mà bản cũ quét hết mọi thứ
// 30 lần mỗi giây: pprof đo được **21% CPU server** nằm ở đúng vòng lặp này, phần lớn là
// chi phí duyệt map chứ không phải phép so sánh.
func (g *Game) expireItems() {
	if len(g.expiring) == 0 {
		return
	}
	keep := g.expiring[:0]
	for _, id := range g.expiring {
		it, ok := g.Items[id]
		if !ok {
			continue // đã bị nhặt hoặc xoá vì lý do khác
		}
		if g.Time >= it.ExpireAt {
			g.removeItem(it)
			g.emit(Event{Kind: "ItemExpired", A: it.ID, X: float32(it.Pos.X), Y: float32(it.Pos.Y)})
			continue
		}
		keep = append(keep, id)
	}
	g.expiring = keep
}

// noteExpiring ghi một viên có hạn vào danh sách cần theo dõi. Gọi ngay sau khi đặt
// `ExpireAt`; quên gọi thì viên đó tồn tại mãi.
func (g *Game) noteExpiring(it *Item) { g.expiring = append(g.expiring, it.ID) }

// AltarMarkers is what the minimap needs: where the altars are and which are about to burst
// (spec §8.5, §12.1).
type AltarMarker struct {
	X, Y    float64
	Kind    string
	Warning bool
}

func (g *Game) AltarMarkers() []AltarMarker {
	out := make([]AltarMarker, 0, len(g.altars))
	for _, st := range g.altars {
		out = append(out, AltarMarker{X: st.a.X, Y: st.a.Y, Kind: st.a.Kind, Warning: st.warned})
	}
	return out
}
