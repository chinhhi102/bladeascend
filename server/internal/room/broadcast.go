package room

import (
	"kiemxoay/server/internal/game"
	"kiemxoay/server/internal/netproto"
)

const (
	leaderboardTop = 10
	nameResendSecs = 2.0
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
		c.send(r.snapshotFor(c, viewer, events))
	}
}

func (r *Room) snapshotFor(c *Client, viewer *game.Player, events []game.Event) netproto.Snapshot {
	g := r.G
	s := netproto.Snapshot{T: "snap", Tick: r.tick, Time: g.Time, Ack: viewer.LastSeq, SrvMs: r.stepMs}
	seen := map[uint32]bool{}
	// Names go out the first time a client sees a player, and again every couple of seconds
	// so a client that missed the first packet still recovers.
	resendNames := r.namesDue
	for _, p := range g.PlayerList() {
		if !g.CanSee(viewer, p) {
			continue
		}
		seen[p.ID] = true
		s.Players = append(s.Players, playerState(g, viewer, p, resendNames || !c.knownPlayers[p.ID]))
		c.knownPlayers[p.ID] = true
	}
	for id := range c.knownPlayers {
		if _, alive := g.Players[id]; !alive {
			delete(c.knownPlayers, id)
		}
	}
	for _, m := range g.Monsters {
		if g.InView(viewer, m.Pos) {
			s.Monsters = append(s.Monsters, netproto.MonsterState{ID: m.ID, X: float32(m.Pos.X), Y: float32(m.Pos.Y),
				Kind: m.Kind, HP: m.HP, Max: m.MaxHP, R: float32(m.R)})
		}
	}
	r.diffItems(c, viewer, &s)
	s.Events = r.eventsFor(viewer, events, seen)
	return s
}

// diffItems sends items entering the view and ids of items that left it or were picked up.
func (r *Room) diffItems(c *Client, viewer *game.Player, s *netproto.Snapshot) {
	g := r.G
	for _, it := range g.Items {
		in := g.InView(viewer, it.Pos)
		if in && !c.knownItems[it.ID] {
			c.knownItems[it.ID] = true
			s.ItemsAdded = append(s.ItemsAdded, netproto.ItemState{ID: it.ID, X: float32(it.Pos.X), Y: float32(it.Pos.Y),
				Kind: it.Kind, Value: float32(it.Value)})
		}
	}
	for id := range c.knownItems {
		it, ok := g.Items[id]
		if !ok || !g.InView(viewer, it.Pos) {
			delete(c.knownItems, id)
			s.ItemsRemoved = append(s.ItemsRemoved, id)
		}
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
