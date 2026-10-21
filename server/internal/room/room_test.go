package room

import (
	"fmt"
	"testing"

	"kiemxoay/server/internal/config"
	"kiemxoay/server/internal/netproto"
	"kiemxoay/server/internal/store"
)

type fakeConn struct{ msgs int }

func (f *fakeConn) Send([]byte) bool { f.msgs++; return true }

type memRecorder struct{ lives []store.Life }

func (m *memRecorder) Record(l store.Life) error { m.lives = append(m.lives, l); return nil }

func testManager(t *testing.T) (*Manager, *memRecorder, *config.Config) {
	t.Helper()
	cfg := config.MustLoadDefault()
	cfg.Room.MonsterCount = 0
	cfg.Items.CountPerRoom = 0
	rec := &memRecorder{}
	m := NewManager(cfg, nil, rec)
	m.Start = func(*Room) {} // tests step rooms manually
	return m, rec, cfg
}

func joinN(m *Manager, n int) []*Client {
	var out []*Client
	for i := 0; i < n; i++ {
		c := NewClient(&fakeConn{}, false)
		m.Join(c, netproto.Join{Name: fmt.Sprintf("P%d", i), Char: "a_hon"})
		out = append(out, c)
	}
	return out
}

func stepSecs(r *Room, secs float64) {
	dt := 1 / float64(r.cfg.Net.TickHz)
	for t := 0.0; t < secs; t += dt {
		r.Step(dt)
	}
}

func TestBotFillToTwenty(t *testing.T) {
	m, _, _ := testManager(t)
	cs := joinN(m, 5)
	r := m.AllRooms()[0]
	r.Step(1.0 / 30)
	for _, c := range cs {
		r.G.Players[c.PlayerID].InvulnUntil = 1e9 // keep humans alive while bots fill
	}
	stepSecs(r, 30)
	if r.Humans() != 5 || r.Bots() != 15 {
		t.Fatalf("humans %d bots %d, want 5 and 15", r.Humans(), r.Bots())
	}
	joinN(m, 3)
	stepSecs(r, 5)
	if r.Bots() != 15 {
		t.Fatalf("joining humans must not remove bots: bots = %d", r.Bots())
	}
}

func TestTwentyFirstHumanGetsNewRoom(t *testing.T) {
	m, _, cfg := testManager(t)
	joinN(m, cfg.Room.MaxHumans)
	joinN(m, 1)
	rooms := m.AllRooms()
	if len(rooms) != 2 {
		t.Fatalf("rooms = %d, want 2", len(rooms))
	}
	if m.humansOf(rooms[0]) != 20 || m.humansOf(rooms[1]) != 1 {
		t.Fatalf("humans = %d / %d", m.humansOf(rooms[0]), m.humansOf(rooms[1]))
	}
}

func TestMatchmakingPrefersFullestRoom(t *testing.T) {
	m, _, _ := testManager(t)
	a := m.createLocked()
	b := m.createLocked()
	a.reserved, b.reserved = 3, 7
	c := NewClient(&fakeConn{}, false)
	if got := m.Join(c, netproto.Join{Name: "X"}); got != b {
		t.Fatalf("joined %s, want the fuller room %s", got.ID, b.ID)
	}
}

func TestDeathRecordsHumanOnlyAndFreesSlot(t *testing.T) {
	m, rec, _ := testManager(t)
	cs := joinN(m, 1)
	r := m.AllRooms()[0]
	stepSecs(r, 3) // let bots spawn
	p := r.G.Players[cs[0].PlayerID]
	p.Dead, p.DiedAt = true, r.G.Time
	stepSecs(r, 1.2)
	if len(rec.lives) != 1 || rec.lives[0].Name != "P0" {
		t.Fatalf("recorded %+v, want exactly the human", rec.lives)
	}
	if cs[0].Room() != nil || m.humansOf(r) != 0 {
		t.Fatal("dead human should leave the room and free the slot")
	}
	for _, b := range r.G.Players {
		b.Dead, b.DiedAt = true, r.G.Time
	}
	stepSecs(r, 1.2)
	if len(rec.lives) != 1 {
		t.Fatal("bots must never be recorded on the leaderboard")
	}
}

func TestEmptyRoomCloses(t *testing.T) {
	m, _, cfg := testManager(t)
	cs := joinN(m, 1)
	r := m.AllRooms()[0]
	stepSecs(r, 0.1)
	r.RemoveClient(cs[0])
	stepSecs(r, cfg.Room.EmptyCloseSecs+1)
	if !r.closed || len(m.AllRooms()) != 0 {
		t.Fatal("room should close after empty_close_secs with no humans")
	}
}

func TestNameFilter(t *testing.T) {
	cfg := config.MustLoadDefault()
	f := LoadNameFilter(cfg, config.FindRepoFile("config/bad_words.txt"))
	if f.Clean("A") != "" || f.Clean("Một cái tên rất là dài quá") != "" {
		t.Fatal("length limits not enforced")
	}
	if got := f.Clean("  Tèo   Kiếm  "); got != "Tèo Kiếm" {
		t.Fatalf("got %q", got)
	}
	if got := f.Clean("fuckboy"); got != "****boy" {
		t.Fatalf("profanity not masked: %q", got)
	}
}
