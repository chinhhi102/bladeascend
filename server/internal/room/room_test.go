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
	// v6.28 §9.3: bot lấp khu (1 khu trong test) cho đủ sàn `zone_min_chars`.
	wantBots := cfg.Room.ZoneMinChars - 5
	if r.Humans() != 5 || r.Bots() != wantBots {
		t.Fatalf("có %d người %d bot, muốn 5 và %d (sàn khu %d)", r.Humans(), r.Bots(), wantBots, cfg.Room.ZoneMinChars)
	}
	joinN(m, 3)
	r.Step(1.0 / 30)
	for _, p := range r.Zone(0).G.Players {
		p.InvulnUntil = 1e9 // v3 swords are lethal to Luyện Khí: keep everyone alive for this check
	}
	stepSecs(r, 5)
	// Người vào KHÔNG xoá bot (v6.28 §9.3): bot dư chỉ vơi dần khi chết.
	if r.Bots() != wantBots {
		t.Fatalf("thêm người mà bot còn %d, muốn giữ nguyên %d", r.Bots(), wantBots)
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

// Spec v6.28 §9.1–9.2: trần người thật/phòng là giá trị admin đang đặt (mặc định production
// 50). Hạ trần không đuổi ai — phòng đang vượt chỉ ngừng nhận người mới.
func TestRoomCapFollowsAdminSetting(t *testing.T) {
	m, _, cfg := testManager(t)
	if m.Settings().MaxHumans != 50 {
		t.Fatalf("mặc định production = %d, muốn 50", m.Settings().MaxHumans)
	}
	joinN(m, cfg.Room.MaxHumans)
	joinN(m, 1)
	rooms := m.AllRooms()
	if len(rooms) != 2 || m.humansOf(rooms[0]) != 50 || m.humansOf(rooms[1]) != 1 {
		t.Fatalf("%d phòng, người %d / %d — muốn 2 phòng 50 / 1", len(rooms), m.humansOf(rooms[0]), m.humansOf(rooms[1]))
	}
	s := m.Settings()
	s.MaxHumans = 20
	if err := m.SetSettings(s); err != nil {
		t.Fatal(err)
	}
	joinN(m, 1)
	if m.humansOf(rooms[0]) != 50 || m.humansOf(rooms[1]) != 2 {
		t.Fatalf("hạ trần 20: phòng 1 còn %d (không được đuổi ai), phòng 2 có %d (muốn 2)",
			m.humansOf(rooms[0]), m.humansOf(rooms[1]))
	}
	if err := m.SetSettings(RoomSettings{MaxHumans: 0, ZoneMinChars: 10, BotAutoOffAt: 20}); err == nil {
		t.Fatal("trần 0 người phải bị từ chối")
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

// Vùng quan tâm có trễ phải áp vào cả TRẠNG THÁI gửi đi, không chỉ `Despawn` (NETCODE_SPEC
// §4.4). Đo 1/11: server cắt người ngay ở mép khung nhìn trong khi bộ theo dõi vẫn coi họ là
// "đang giữ"; keyframe 2 giây sau client xoá họ, rồi họ bước vào lại = respawn (26/31 lần).
func TestSnapshotKeepsSendingTrackedPlayerJustOutsideView(t *testing.T) {
	m, _, _ := testManager(t)
	cs := joinN(m, 2)
	r := m.AllRooms()[0]
	r.Step(1.0 / 30)
	z := r.Zone(0)
	viewer, other := z.G.Players[cs[0].PlayerID], z.G.Players[cs[1].PlayerID]
	viewer.Pos, other.Pos = z.G.Vec(3000, 3000), z.G.Vec(3200, 3000)
	has := func(s netproto.Snapshot) bool {
		for _, p := range s.Players {
			if p.ID == other.ID {
				return true
			}
		}
		return false
	}
	// Tick chẵn để không rơi vào nhịp 10 Hz bỏ người ở xa.
	z.tick = 0
	if !has(r.snapshotFor(z, cs[0], viewer, nil)) {
		t.Fatal("thiết lập sai: người đứng cạnh phải có trong snapshot")
	}
	hw, _ := z.G.ViewHalfExtents(viewer)
	// Ra khỏi khung nhìn (cả vòng kiếm) nhưng còn trong lề ra 900.
	other.Pos = z.G.Vec(3000+hw+z.G.Geom(other).Reach+300, 3000)
	if z.G.CanSee(viewer, other) {
		t.Fatal("thiết lập sai: người thử phải ra ngoài khung nhìn")
	}
	if !has(r.snapshotFor(z, cs[0], viewer, nil)) {
		t.Fatal("ra mép nhưng bộ theo dõi còn giữ mà KHÔNG gửi nữa: đây là nguồn respawn 2 giây")
	}
	// Trốn vào bụi thì dù còn trong hạn chờ cũng KHÔNG được gửi vị trí.
	other.Bush, other.RevealedUntil = 0, 0
	if has(r.snapshotFor(z, cs[0], viewer, nil)) {
		t.Fatal("người trốn trong bụi vẫn bị gửi vị trí xuống")
	}
}

// Quái cũng phải có trễ vào/ra (NETCODE_SPEC §4.4). Đo 1/11: 28/30 lần "respawn trong 2
// giây" là quái lang thang qua lại mép màn hình.
func TestSnapshotMonsterHysteresis(t *testing.T) {
	m, _, _ := testManager(t)
	cs := joinN(m, 1)
	r := m.AllRooms()[0]
	r.Step(1.0 / 30)
	z := r.Zone(0)
	viewer := z.G.Players[cs[0].PlayerID]
	viewer.Pos = z.G.Vec(3000, 3000)
	for id := range z.G.Monsters {
		delete(z.G.Monsters, id) // chỉ giữ đúng con quái của bài thử
	}
	q := &game.Monster{ID: 999999, Kind: "slime", R: 20, HP: 5, MaxHP: 5, Pos: z.G.Vec(3100, 3000)}
	z.G.Monsters[q.ID] = q
	find := func() *netproto.MonsterState {
		z.tick = 0 // tick chẵn: không rơi vào nhịp 10 Hz bỏ quái ở xa
		s := r.snapshotFor(z, cs[0], viewer, nil)
		for i := range s.Monsters {
			if s.Monsters[i].ID == q.ID {
				return &s.Monsters[i]
			}
		}
		return nil
	}
	if got := find(); got == nil || got.Kind == "" {
		t.Fatal("thiết lập sai: quái trong khung nhìn phải được gửi kèm loại")
	}
	hw, _ := z.G.ViewHalfExtents(viewer)
	q.Pos = z.G.Vec(3000+hw+400, 3000) // ngoài khung, trong lề ra 900
	if find() == nil {
		t.Fatal("quái vừa ra mép mà thôi gửi ngay: client xoá sau 2 giây rồi hiện lại")
	}
	q.Pos = z.G.Vec(3000+hw+2000, 3000) // ra hẳn ngoài lề ra
	if find() != nil {
		t.Fatal("quái ra ngoài lề ra vẫn còn gửi: phí băng thông")
	}
	q.Pos = z.G.Vec(3100, 3000) // quay lại: client đã xoá nó, phải gửi lại đủ loại
	if got := find(); got == nil || got.Kind == "" {
		t.Fatal("quái quay lại sau khi đã quên mà không kèm loại: client không biết vẽ con gì")
	}
}

// Spec v6.29 §7 / checklist §17: người thật chết → rời phòng, chờ 2 phút không tự vào lại.
// Muốn chơi tiếp phải gửi lại `join` (nút "Chơi tiếp") và bắt đầu từ Luyện Khí 1.
func TestDeadHumanNeverAutoRejoins(t *testing.T) {
	m, _, _ := testManager(t)
	cs := joinN(m, 1)
	r := m.AllRooms()[0]
	stepSecs(r, 1)
	id := cs[0].PlayerID
	p := r.Zone(0).G.Players[id]
	p.Dead, p.DiedAt = true, r.Zone(0).G.Time
	stepSecs(r, 120)
	if cs[0].Room() != nil {
		t.Fatal("chết rồi mà vẫn bị gắn lại vào phòng")
	}
	for _, z := range r.Zones() {
		if _, ok := z.G.Players[id]; ok {
			t.Fatal("nhân vật đã chết vẫn nằm trong khu sau 2 phút (hồi sinh tại chỗ?)")
		}
	}
	got := m.Join(cs[0], netproto.Join{Name: "P0", Char: "a_hon"})
	stepSecs(got, 0.1)
	np := got.ZoneOf(cs[0].PlayerID).G.Players[cs[0].PlayerID]
	if np == nil || np.Level != 0 || np.Dead {
		t.Fatalf("Chơi tiếp: %+v, muốn sống ở Luyện Khí 1", np)
	}
}
