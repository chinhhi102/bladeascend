package room

import (
	"sort"

	"kiemxoay/server/internal/game"
	"kiemxoay/server/internal/netproto"
)

const (
	leaderboardTop = 10
	nameResendSecs = 2.0

	// Ngân sách gói tin (spec §13.6 + CHANGES_v4 mục 27). Một người Đạo Tổ có tầm nhìn
	// khổng lồ nên hàng trăm vật phẩm cùng lọt vào một lúc; không chặn thì gói đầu tiên
	// phình lên 38 KB — đo thật bằng cmd/loadtest.
	itemBudgetBytes = 2048 // ItemsAdded tối đa mỗi gói, ưu tiên cái gần người chơi
	itemWireBytes   = 34   // một ItemState sau khi mã hoá msgpack, đo thực tế
	maxEntities     = 150  // trần entity động mỗi snapshot, ưu tiên gần và to
	// Viên EXP nhỏ ngoài bán kính này không gửi khi camera đã lùi xa (spec mục 27).
	smallOrbRadius = 1500
	smallOrbValue  = 20 // "nhỏ" = dưới mức này; viên vừa/lớn và Tiên khí luôn gửi
	// Người/quái ở xa hơn nửa khung nhìn chỉ gửi 10 Hz thay vì 20 Hz.
	farRateDivisor = 2
)

// broadcast sends each client a snapshot filtered by its area of interest (spec §13.3).
func (r *Room) broadcast() {
	events := r.pending
	r.pending = nil
	r.namesDue = r.G.Time-r.namesAt >= nameResendSecs
	if r.namesDue {
		r.namesAt = r.G.Time
	}
	r.sendFeeds(events)
	for id, c := range r.clients {
		viewer, ok := r.G.Players[id]
		if !ok {
			continue
		}
		c.sendSnapshot(r.snapshotFor(c, viewer, events))
	}
}

func (r *Room) snapshotFor(c *Client, viewer *game.Player, events []game.Event) netproto.Snapshot {
	g := r.G
	s := netproto.Snapshot{T: "snap", Tick: r.tick, Time: g.Time, Ack: viewer.LastSeq, SrvMs: r.stepMs}
	seen := map[uint32]bool{}
	// Names go out the first time a client sees a player, and again every couple of seconds
	// so a client that missed the first packet still recovers.
	resendNames := r.namesDue
	// Người ở XA chỉ gửi 10 Hz thay vì 20 Hz (spec mục 27): client nội suy nên bỏ một nhịp
	// không ai thấy, mà nửa số byte của người ở rìa màn hình thì tiết kiệm thật.
	// Người MỚI thấy lần đầu luôn gửi ngay, nếu không họ "hiện ra" trễ nửa nhịp.
	half := g.ViewHalfExtentsX(viewer) / 2
	slow := r.tick%farRateDivisor != 0
	for _, p := range g.PlayerList() {
		if !g.CanSee(viewer, p) {
			continue
		}
		seen[p.ID] = true
		known := c.knownPlayers[p.ID]
		if known && slow && p.ID != viewer.ID && p.Pos.Dist(viewer.Pos) > half {
			continue
		}
		s.Players = append(s.Players, playerState(g, viewer, p, resendNames || !known))
		c.knownPlayers[p.ID] = true
	}
	for id := range c.knownPlayers {
		if _, alive := g.Players[id]; !alive {
			delete(c.knownPlayers, id)
		}
	}
	for _, m := range g.Monsters {
		if !g.InView(viewer, m.Pos) {
			continue
		}
		if slow && m.Pos.Dist(viewer.Pos) > half {
			continue // quái ở xa: 10 Hz là đủ, chúng đi lang thang chậm
		}
		if len(s.Players)+len(s.Monsters) >= maxEntities {
			break // trần entity động mỗi gói (spec mục 27)
		}
		ms := netproto.MonsterState{ID: m.ID, X: float32(m.Pos.X), Y: float32(m.Pos.Y), HP: m.HP}
		if !c.knownMonsters[m.ID] {
			// Lần đầu thấy: gửi kèm loại, máu tối đa và bán kính — ba thứ không bao giờ đổi.
			c.knownMonsters[m.ID] = true
			ms.Kind, ms.Max, ms.R = m.Kind, m.MaxHP, float32(m.R)
		}
		s.Monsters = append(s.Monsters, ms)
	}
	for id := range c.knownMonsters {
		if _, alive := g.Monsters[id]; !alive {
			delete(c.knownMonsters, id)
		}
	}
	r.diffItems(c, viewer, &s)
	s.Events = r.eventsFor(viewer, events, seen)
	return s
}

// diffItems sends items entering the view and ids of items that left it or were picked up.
// wantItem quyết định một vật phẩm có đáng gửi cho người này không.
//
// Viên EXP NHỎ ngoài 1 500 đơn vị thì bỏ khi camera đã lùi xa (spec mục 27): người Đại
// Thừa/Tiên giới nhìn thấy cả vùng 4 000 đơn vị, trong đó có hàng nghìn viên nhỏ mà trên màn
// hình chúng chỉ là mấy chấm — gửi hết thì tốn băng thông cho thứ không ai nhặt tới.
// `d2` là BÌNH PHƯƠNG khoảng cách, không phải khoảng cách: so sánh cho ra cùng kết quả mà
// khỏi phải khai căn. pprof đo được `math.Hypot` ăn 8% CPU server chỉ vì chỗ này chạy cho
// từng vật phẩm, từng client, từng tick.
func (r *Room) wantItem(viewer *game.Player, hw, hh float64, it *game.Item, d2 float64) bool {
	if !game.InViewAt(viewer.Pos, hw, hh, it.Pos, 0) {
		return false
	}
	if it.Kind != "orb" || float64(it.Value) >= smallOrbValue {
		return true
	}
	return d2 <= smallOrbRadius*smallOrbRadius
}

// diffItems gửi vật phẩm mới lọt vào tầm nhìn và id của những cái vừa rời đi.
//
// ItemsAdded bị chặn theo NGÂN SÁCH BYTE và xếp theo khoảng cách: lúc một người Tiên giới
// mới vào vùng đông đồ, cả trăm vật phẩm cùng lọt vào tầm nhìn trong một tick. Gửi hết là
// một gói 38 KB (đo thật) — trên mạng 4G yếu thì gói đó một mình chiếm gần nửa giây. Phần
// còn lại tự sang các gói sau, không mất gì vì `knownItems` chỉ đánh dấu cái đã gửi.
// itemCand: một vật phẩm ứng viên kèm BÌNH PHƯƠNG khoảng cách tới người xem.
type itemCand struct {
	it *game.Item
	d2 float64
}

func (r *Room) diffItems(c *Client, viewer *game.Player, s *netproto.Snapshot) {
	g := r.G
	// Lấy nửa khung nhìn MỘT LẦN cho cả vòng lặp: nó chỉ phụ thuộc người xem. Trước đây mỗi
	// vật phẩm gọi lại `ViewHalfExtents`, và hàm đó có `math.Pow` — 44% CPU server đi vào
	// đúng chỗ này (pprof, CHANGES_v4 mục 26).
	hw, hh := g.ViewHalfExtents(viewer)
	vp := viewer.Pos
	// Chỉ duyệt các ô lưới chạm khung nhìn, không duyệt toàn bộ vật phẩm của bản đồ. Bản
	// đồ có hàng nghìn viên EXP; duyệt hết cho TỪNG người, 20 lần/giây, là vài triệu lượt
	// tra bảng băm mỗi giây — pprof đo được 35% CPU server nằm ở đó.
	if n := g.SlotCount(); len(c.knownItems) < n {
		c.knownItems = append(c.knownItems, make([]uint32, n-len(c.knownItems))...)
	}
	known := c.knownItems
	// Quét XOÁ TRƯỚC, thêm sau. Thứ tự này là bắt buộc, không phải tuỳ ý:
	//
	// Slot được DÙNG LẠI. Viên EXP vừa bị nhặt sẽ nhả slot ra, và một viên mới có thể chiếm
	// đúng slot đó ngay trong cùng một tick. Nếu thêm trước thì vòng thêm ghi đè
	// `known[slot]` bằng id của viên MỚI, rồi vòng quét thấy "slot này có chủ hợp lệ" và
	// không bao giờ báo client xoá viên CŨ — người chơi nhặt xong mà viên vẫn nằm đó mãi.
	// Quét trước thì viên cũ được báo xoá, slot trả về 0, rồi viên mới mới được thêm.
	for slot, id := range known {
		if id == 0 {
			continue
		}
		it := g.ItemBySlot(slot)
		ok := it != nil && it.ID == id
		if ok {
			dx, dy := it.Pos.X-vp.X, it.Pos.Y-vp.Y
			ok = r.wantItem(viewer, hw, hh, it, dx*dx+dy*dy)
		}
		if !ok {
			known[slot] = 0
			s.ItemsRemoved = append(s.ItemsRemoved, id)
		}
	}
	cands := r.itemBuf[:0]
	g.ItemsNear(vp, hw, hh, func(it *game.Item) {
		if known[it.Slot()] == it.ID {
			return
		}
		dx, dy := it.Pos.X-vp.X, it.Pos.Y-vp.Y
		d2 := dx*dx + dy*dy
		if r.wantItem(viewer, hw, hh, it, d2) {
			cands = append(cands, itemCand{it, d2})
		}
	})
	r.itemBuf = cands
	sort.Slice(cands, func(a, b int) bool { return cands[a].d2 < cands[b].d2 })
	budget := itemBudgetBytes
	for _, cd := range cands {
		if budget < itemWireBytes {
			break // hết ngân sách gói này; cái còn lại đi gói sau
		}
		budget -= itemWireBytes
		known[cd.it.Slot()] = cd.it.ID
		s.ItemsAdded = append(s.ItemsAdded, netproto.ItemState{ID: cd.it.ID, X: float32(cd.it.Pos.X), Y: float32(cd.it.Pos.Y),
			Kind: cd.it.Kind, Value: float32(cd.it.Value)})
	}
}

// eventsFor keeps events the viewer may know about: global ones, their own, and those in view.
// Events about hidden players are dropped so stealth cannot be detected from events.
func (r *Room) eventsFor(viewer *game.Player, events []game.Event, seen map[uint32]bool) []game.Event {
	var out []game.Event
	for _, e := range events {
		if e.Kind == "KillFeed" || (e.To != 0 && e.To != viewer.ID) {
			continue
		}
		mine := e.A == viewer.ID || e.B == viewer.ID
		if !(e.Global || mine || r.eventVisible(viewer, e, seen)) {
			continue
		}
		out = append(out, e)
	}
	return out
}

// playerEvents are attributed to player A and hidden when A is not visible to the viewer.
var playerEvents = map[string]bool{"LevelUp": true, "Breakthrough": true, "RealmDrop": true,
	"RingFlip": true, "SwordBreak": true, "BuffStart": true, "BuffEnd": true, "Emote": true, "TeleportIn": true}

func (r *Room) eventVisible(viewer *game.Player, e game.Event, seen map[uint32]bool) bool {
	if !r.G.InView(viewer, game.Vec{X: float64(e.X), Y: float64(e.Y)}) {
		return false
	}
	if playerEvents[e.Kind] {
		return seen[e.A]
	}
	if e.Kind == "Hit" {
		return seen[e.A] || seen[e.B]
	}
	return true
}

// serverWideTopN: hạ được người đứng trong Top N của phòng thì cả server nghe thấy (§11.4).
const serverWideTopN = 5

// sendFeeds turns KillFeed events into named feed lines for everyone in the room, and
// relays the big ones to every other room.
func (r *Room) sendFeeds(events []game.Event) {
	for _, e := range events {
		if e.Kind != "KillFeed" {
			continue
		}
		f := netproto.Feed{T: "feed", Kind: e.S, A: r.nameOf(e.A), B: r.nameOf(e.B), Realm: e.N}
		// Với dòng chém chết, e.N là hạng của nạn nhân chứ không phải cảnh giới.
		kill := e.S == "kill" || e.S == "gate_kill"
		if kill {
			f.Realm = 0
			if e.N >= 1 && e.N <= serverWideTopN {
				f.Top, f.Room = e.N, r.ID
			}
		}
		for _, c := range r.clients {
			c.send(f)
		}
		if f.Top > 0 && r.mgr != nil {
			r.mgr.BroadcastFeed(f, r)
		}
	}
}

func (r *Room) nameOf(id uint32) string {
	if p, ok := r.G.Players[id]; ok {
		return p.Name
	}
	return ""
}

// sendLeaderboards sends the in-room top 10, each client's own rank and Uy Danh, and the
// minimap markers (spec §11.2, §12.1).
func (r *Room) sendLeaderboards() {
	ranking := r.G.Ranking()
	top := make([]netproto.LeaderEntry, 0, leaderboardTop)
	rankOf := map[uint32]int{}
	for i, p := range ranking {
		rankOf[p.ID] = i + 1
		if i < leaderboardTop {
			top = append(top, netproto.LeaderEntry{ID: p.ID, Name: p.Name, Exp: float32(p.TotalExp),
				Level: p.Level, Score: float32(p.Score), Title: p.Title})
		}
	}
	base := netproto.Leaderboard{T: "lb", Top: top, Total: len(ranking)}
	// Spec §12.1: the crown shows for the whole room, but a leader in a bush or under
	// Ẩn Thân drops off the minimap so it never gives their position away.
	if len(ranking) > 0 && r.cfg.Minimap.ShowTop1 {
		if lead := ranking[0]; !r.G.HiddenInBush(lead) && !lead.HasBuff(game.BuffStealth, r.G.Time) {
			base.Top1, base.Top1X, base.Top1Y = true, float32(lead.Pos.X), float32(lead.Pos.Y)
		}
	}
	for _, a := range r.G.AltarMarkers() {
		base.Altars = append(base.Altars, netproto.AltarMark{X: float32(a.X), Y: float32(a.Y), Kind: a.Kind, Warn: a.Warning})
	}
	if r.cfg.Minimap.ShowChest {
		if pos, warning, ok := r.G.ChestMarker(); ok {
			base.HasBau, base.ChestX, base.ChestY, base.ChestIn = true, float32(pos.X), float32(pos.Y), warning
		}
	}
	for id, c := range r.clients {
		msg := base
		msg.Rank = rankOf[id]
		if p, ok := r.G.Players[id]; ok {
			msg.MyScore = float32(p.Score)
		}
		c.send(msg)
	}
}
