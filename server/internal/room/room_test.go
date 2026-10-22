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
	r.Step(1.0 / 30)
	for _, p := range r.G.Players {
		p.InvulnUntil = 1e9 // v3 swords are lethal to Luyện Khí: keep everyone alive for this check
	}
	stepSecs(r, 5)
	if r.Bots() != 15 {
		t.Fatalf("joining humans must not remove bots: bots = %d", r.Bots())
	}
}

// Bot sinh ra không được quá Trúc Cơ. Kiểm tra mức SINH RA (randomBotLevel), không phải
// cảnh giới sau đó — bot farm lên cao hơn là chuyện bình thường.
func TestBotsSpawnNoHigherThanTrucCo(t *testing.T) {
	m, _, cfg := testManager(t)
	joinN(m, 1)
	r := m.AllRooms()[0]
	if n := len(cfg.Bots.RealmWeights); n != 2 {
		t.Fatalf("realm_weights có %d cảnh giới, muốn 2 (Luyện Khí + Trúc Cơ)", n)
	}
	seenTrucCo := false
	for i := 0; i < 500; i++ {
		level, swords := r.randomBotLevel()
		realm := r.G.Prog.Realm(level)
		if realm > 1 {
			t.Fatalf("bot sinh ra ở cảnh giới %d (level %d), không được quá Trúc Cơ", realm, level)
		}
		if realm == 1 {
			seenTrucCo = true
		}
		if swords < 1 || swords > r.G.Prog.MaxSwords(realm) {
			t.Fatalf("bot sinh ra với %d kiếm, ngoài khoảng [1, %d]", swords, r.G.Prog.MaxSwords(realm))
		}
	}
	if !seenTrucCo {
		t.Fatal("500 lần không có bot Trúc Cơ nào: realm_weights không được dùng")
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

// Spec §17: a stealthed player never appears in packets for far-away viewers.
// Spec §6.2 / §13.3 (v3.3): người Ẩn Thân VẪN được gửi xuống kèm cờ stealth để client vẽ mờ,
// nhưng không kèm tên và danh hiệu nên vẫn không lộ là ai.
func TestSnapshotSendsStealthFlaggedWithoutName(t *testing.T) {
	m, _, _ := testManager(t)
	cs := joinN(m, 2)
	r := m.AllRooms()[0]
	r.Step(1.0 / 30)
	viewer, hider := r.G.Players[cs[0].PlayerID], r.G.Players[cs[1].PlayerID]
	viewer.Pos, hider.Pos = r.G.Vec(3000, 3000), r.G.Vec(3400, 3000)
	hider.Title = 1
	r.G.ApplyBuff(hider, "stealth")
	snap := r.snapshotFor(cs[0], viewer, r.G.DrainEvents())
	var got *netproto.PlayerState
	for i, p := range snap.Players {
		if p.ID == hider.ID {
			got = &snap.Players[i]
		}
	}
	if got == nil {
		t.Fatal("người Ẩn Thân phải vẫn có trong snapshot (vẽ mờ 20%)")
	}
	if got.Flags&netproto.FlagStealth == 0 {
		t.Fatal("thiếu cờ stealth nên client không biết phải vẽ mờ")
	}
	if got.Name != "" || got.Title != 0 {
		t.Fatalf("Ẩn Thân mà vẫn lộ tên/danh hiệu: %q, title %d", got.Name, got.Title)
	}
}
