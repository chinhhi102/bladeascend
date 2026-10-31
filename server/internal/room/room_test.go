package room

import (
	"fmt"
	"testing"

	"bladeascend/server/internal/config"
	"bladeascend/server/internal/game"
	"bladeascend/server/internal/netproto"
	"bladeascend/server/internal/store"
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
	// Các test dưới đây kiểm luật chiến đấu và bot, không kiểm thế giới nhiều khu: để 5 khu
	// thì hai người tham gia có thể rơi vào hai khu khác nhau và không bao giờ gặp nhau.
	// Thế giới 5 khu có test riêng trong zone_test.go.
	cfg.Room.Zones = 1
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

func TestBotFillToTarget(t *testing.T) {
	m, _, cfg := testManager(t)
	cs := joinN(m, 5)
	r := m.AllRooms()[0]
	r.Step(1.0 / 30)
	for _, c := range cs {
		r.Zone(0).G.Players[c.PlayerID].InvulnUntil = 1e9 // keep humans alive while bots fill
	}
	stepSecs(r, 30)
	// Trần cả phòng là `room.bot_fill_target`; bot lấp cho đủ phần còn lại.
	wantBots := cfg.Room.BotFillTarget - 5
	if r.Humans() != 5 || r.Bots() != wantBots {
		t.Fatalf("có %d người %d bot, muốn 5 và %d (trần phòng %d)", r.Humans(), r.Bots(), wantBots, cfg.Room.BotFillTarget)
	}
	joinN(m, 3)
	r.Step(1.0 / 30)
	for _, p := range r.Zone(0).G.Players {
		p.InvulnUntil = 1e9 // v3 swords are lethal to Luyện Khí: keep everyone alive for this check
	}
	stepSecs(r, 5)
	// Người thật được ƯU TIÊN trong trần phòng: thêm người thì bot phải nhường chỗ dần
	// (bot chết thì không hồi sinh nữa), nhưng tổng số nhân vật không bao giờ vượt trần.
	if total := r.Humans() + r.Bots(); total > cfg.Room.BotFillTarget {
		t.Fatalf("phòng có %d nhân vật, vượt trần %d", total, cfg.Room.BotFillTarget)
	}
	if r.Humans() != 8 {
		t.Fatalf("có %d người thật, muốn 8", r.Humans())
	}
}

// Spec §9.3: phòng chưa đủ `spawn_rank` người thì bot sinh ra ở Luyện Khí.
func TestBotsSpawnAtLuyenKhiWhenRoomIsSmall(t *testing.T) {
	m, _, _ := testManager(t)
	joinN(m, 1)
	r := m.AllRooms()[0]
	for i := 0; i < 200; i++ {
		level, swords := r.randomBotLevel()
		if realm := r.Zone(0).G.Prog.Realm(level); realm != 0 {
			t.Fatalf("phòng ít người mà bot sinh ở cảnh giới %d (level %d), muốn Luyện Khí", realm, level)
		}
		if swords < 1 || swords > r.Zone(0).G.Prog.MaxSwords(0) {
			t.Fatalf("bot sinh ra với %d kiếm, ngoài khoảng [1, %d]", swords, r.Zone(0).G.Prog.MaxSwords(0))
		}
	}
}

// Spec §9.3: đủ người thì bot sinh đúng ngang với người đang đứng hạng spawn_rank (11),
// không còn trần cảnh giới nào cả.
func TestBotsSpawnAtRank11NoRealmCap(t *testing.T) {
	m, _, cfg := testManager(t)
	joinN(m, 1)
	r := m.AllRooms()[0]

	// Dựng đủ người, hạng cao dần theo Uy Danh; hạng 11 rơi vào Nguyên Anh tầng 5 (level 34).
	for i := 0; i < cfg.Bots.SpawnRank+3; i++ {
		level := 34
		if i < cfg.Bots.SpawnRank-1 {
			level = 60 // Top 10 toàn người cảnh giới rất cao
		}
		p := r.Zone(0).G.AddPlayer(game.SpawnOptions{Name: "P", Level: level, Swords: 1, IsBot: true})
		p.Score = float64(100 * (cfg.Bots.SpawnRank + 3 - i))
	}
	want := r.Zone(0).G.Ranking()[cfg.Bots.SpawnRank-1].Level
	if r.Zone(0).G.Prog.Realm(want) <= 1 {
		t.Fatalf("thiết lập sai: hạng 11 phải ở cảnh giới cao hơn Trúc Cơ, đang là level %d", want)
	}
	for i := 0; i < 50; i++ {
		level, swords := r.randomBotLevel()
		if level != want {
			t.Fatalf("bot sinh ở level %d, muốn %d (ngang hạng %d, không bị chặn trần)",
				level, want, cfg.Bots.SpawnRank)
		}
		if maxS := r.Zone(0).G.Prog.MaxSwords(r.Zone(0).G.Prog.Realm(level)); swords < 1 || swords > maxS {
			t.Fatalf("bot sinh với %d kiếm, ngoài [1, %d]", swords, maxS)
		}
	}
}

// Spec §11.4: hạ một người trong Top 5 của phòng → dòng kill feed loan ra MỌI phòng.
func TestTopFiveKillAnnouncedServerWide(t *testing.T) {
	m, _, _ := testManager(t)
	joinN(m, 1)
	r := m.AllRooms()[0]
	r.Step(1.0 / 30)

	killer := r.Zone(0).G.AddPlayer(game.SpawnOptions{Name: "SátThủ", Level: 30, Swords: 20})
	victim := r.Zone(0).G.AddPlayer(game.SpawnOptions{Name: "ĐạiCa", Level: 0, Swords: 1})
	victim.Score = 9999 // đang đứng hạng 1 trong phòng
	collected := map[string][]netproto.Feed{}
	collect := func(room *Room) {
		for _, c := range room.Clients() {
			c.Conn = &feedSpy{room: room.ID, out: collected}
		}
	}
	collect(r)

	// Phòng thứ hai để kiểm tra tin có lan sang không.
	other := m.createLocked()
	c2 := NewClient(&fakeConn{}, false)
	other.addHuman(c2, netproto.Join{Name: "NgườiPhòngKhác", Char: "a_hon"})
	collect(other)

	// Đặt nạn nhân nằm trong vòng kiếm của sát thủ rồi chạy tới khi chết thật.
	killer.Pos = r.Zone(0).G.Vec(6000, 6000)
	victim.Pos = r.Zone(0).G.Vec(6000+r.Zone(0).G.Geom(killer).Inner, 6000)
	killer.InvulnUntil, victim.InvulnUntil = 0, 0
	for i := 0; i < 120 && !victim.Dead; i++ {
		r.Step(1.0 / 30)
		victim.Pos = r.Zone(0).G.Vec(6000+r.Zone(0).G.Geom(killer).Inner, 6000)
	}
	if !victim.Dead {
		t.Fatal("không chém chết được nạn nhân Luyện Khí")
	}
	other.drainInbox() // tin lan sang chạy trên goroutine phòng kia

	// Bot trong phòng cũng sinh feed (đột phá, rớt cảnh giới…) nên phải lọc đúng dòng.
	findKill := func(room string) (netproto.Feed, bool) {
		for _, f := range collected[room] {
			if f.B == "ĐạiCa" && (f.Kind == "kill" || f.Kind == "gate_kill") {
				return f, true
			}
		}
		return netproto.Feed{}, false
	}
	f, ok := findKill(r.ID)
	if !ok {
		t.Fatalf("phòng gốc không nhận được dòng kill feed: %+v", collected[r.ID])
	}
	if f.Top != 1 || f.Room != r.ID {
		t.Fatalf("dòng feed: Top = %d, Room = %q; muốn 1 và %q", f.Top, f.Room, r.ID)
	}
	got, ok := findKill(other.ID)
	if !ok {
		t.Fatalf("phòng khác không nhận được tin loan toàn server: %+v", collected[other.ID])
	}
	if got.Top != 1 {
		t.Fatalf("tin loan sang phòng khác sai: %+v", got)
	}
	// Phòng khác KHÔNG được nhận feed thường của phòng này.
	for _, f := range collected[other.ID] {
		if f.Top == 0 {
			t.Fatalf("feed thường bị lọt sang phòng khác: %+v", f)
		}
	}
}

// feedSpy ghi lại các gói Feed gửi tới một client.
type feedSpy struct {
	room string
	out  map[string][]netproto.Feed
}

func (s *feedSpy) Send(b []byte) bool {
	var f netproto.Feed
	if err := netproto.Decode(b, &f); err == nil && f.T == "feed" {
		s.out[s.room] = append(s.out[s.room], f)
	}
	return true
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
	p := r.Zone(0).G.Players[cs[0].PlayerID]
	p.Dead, p.DiedAt = true, r.Zone(0).G.Time
	stepSecs(r, 1.2)
	if len(rec.lives) != 1 || rec.lives[0].Name != "P0" {
		t.Fatalf("recorded %+v, want exactly the human", rec.lives)
	}
	if cs[0].Room() != nil || m.humansOf(r) != 0 {
		t.Fatal("dead human should leave the room and free the slot")
	}
	for _, b := range r.Zone(0).G.Players {
		b.Dead, b.DiedAt = true, r.Zone(0).G.Time
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
	viewer, hider := r.Zone(0).G.Players[cs[0].PlayerID], r.Zone(0).G.Players[cs[1].PlayerID]
	viewer.Pos, hider.Pos = r.Zone(0).G.Vec(3000, 3000), r.Zone(0).G.Vec(3400, 3000)
	hider.Title = 1
	r.Zone(0).G.ApplyBuff(hider, "stealth")
	snap := r.snapshotFor(r.Zone(0), cs[0], viewer, r.Zone(0).G.DrainEvents())
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
