package room

import (
	"testing"

	"bladeascend/server/internal/config"

	"bladeascend/server/internal/netproto"
)

// Vào lại phiên khi F5 — GAME_SPEC §7.1 (v6.53), checklist §17:
// "F5 giữa trận → trong 20 s vào lại đúng nhân vật (cảnh giới, kiếm, Uy Danh, khu);
//  bị chém trong lúc mất kết nối → chết, F5 xong thấy Tổng kết."

// decodeAll giải mã mọi gói một kết nối đã nhận, lọc theo loại.
func sentOfKind(t *testing.T, c *recConn, kind string) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, b := range c.sent {
		var m map[string]any
		if err := netproto.Decode(b, &m); err != nil {
			continue // snapshot nhị phân, không phải msgpack
		}
		if s, _ := m["t"].(string); s == kind {
			out = append(out, m)
		}
	}
	return out
}

// joinOne vào trận và trả về (client, mã phiên server vừa cấp).
func joinOne(t *testing.T, m *Manager, name string) (*Client, *recConn, string) {
	t.Helper()
	conn := &recConn{}
	c := NewClient(conn, false)
	m.Join(c, netproto.Join{Name: name, Char: "a_hon"})
	r := m.AllRooms()[0]
	settle(r, m.cfg)
	w := sentOfKind(t, conn, "welcome")
	if len(w) == 0 {
		t.Fatal("không nhận được welcome")
	}
	key, _ := w[len(w)-1]["sk"].(string)
	if len(key) != 64 {
		t.Fatalf("mã phiên phải là 32 byte hex (64 ký tự), được %d: %q", len(key), key)
	}
	return c, conn, key
}

// resumeManager: 1 khu, không bot/quái/vật phẩm — để phép đếm nhân vật nói đúng chuyện.
func resumeManager(t *testing.T) (*Manager, *config.Config) {
	t.Helper()
	cfg := config.MustLoadDefault()
	cfg.Room.MonsterCount = 0
	cfg.Items.CountPerRoom = 0
	cfg.Room.Zones = 1
	cfg.Room.ZoneMinChars = config.Uniform(0)
	m := NewManager(cfg, nil, &memRecorder{})
	m.Start = func(*Room) {}
	return m, cfg
}

func TestMatKetNoiThiNhanVatOLai20Giay(t *testing.T) {
	m, cfg := resumeManager(t)
	c, _, key := joinOne(t, m, "Teo")
	r := m.AllRooms()[0]
	z := r.Zone(0)
	id := c.PlayerID
	// Cho nhân vật một trạng thái đáng để giữ lại.
	r.Do(func() {
		p := z.G.Players[id]
		z.G.ForceState(p, 25, 20, p.Pos)
		p.Score = 1234
		p.InvulnUntil = 0 // bỏ bất tử lúc vào trận: bài này đo chuyện khác
	})
	// Qua hết 5 giây bất tử lúc vào trận (§7) — nếu không thì phép kiểm "không bất tử"
	// bên dưới chỉ đang đo cái bất tử của lúc spawn.
	stepSecs(r, cfg.Player.InvulnOnSpawn+0.5)
	want, wantLevel := z.G.Players[id].Ring.Count, z.G.Players[id].Level

	r.RemoveClient(c)
	settle(r, cfg)
	p, alive := z.G.Players[id]
	if !alive {
		t.Fatal("§7.1: nhân vật phải Ở LẠI thế giới sau khi mất kết nối")
	}
	if !p.Offline {
		t.Fatal("nhân vật phải được đánh dấu mất kết nối (không nhặt đồ được)")
	}
	if p.Invulnerable(z.G.Time) {
		t.Fatal("§7.1: KHÔNG bất tử — thoát mạng không được là cách né chết")
	}
	if z.G.CanPickUp(p) {
		t.Fatal("§7.1: đang mất kết nối thì không nhặt được gì")
	}
	if r.HeldCount() != 1 {
		t.Fatalf("muốn 1 chỗ giữ, có %d", r.HeldCount())
	}

	// Nối lại trong 20 giây → ĐÚNG nhân vật đó, với đủ cảnh giới / kiếm / Uy Danh.
	c2 := NewClient(&recConn{}, false)
	conn2 := c2.Conn.(*recConn)
	m.Join(c2, netproto.Join{Name: "Teo", Char: "a_hon", Resume: key})
	settle(r, cfg)
	if c2.PlayerID != id {
		t.Fatalf("vào lại phải gắn đúng nhân vật cũ: được %d, muốn %d", c2.PlayerID, id)
	}
	p = z.G.Players[id]
	if p.Level != wantLevel || p.Ring.Count != want || p.Score != 1234 {
		t.Fatalf("mất trạng thái khi vào lại: cảnh giới %d, %d kiếm, Uy Danh %.0f", p.Level, p.Ring.Count, p.Score)
	}
	if p.Offline {
		t.Fatal("vào lại rồi mà vẫn còn cờ mất kết nối")
	}
	if r.HeldCount() != 0 {
		t.Fatal("chỗ giữ phải được dọn sau khi vào lại")
	}
	w := sentOfKind(t, conn2, "welcome")
	if len(w) == 0 {
		t.Fatal("vào lại không có welcome")
	}
	last := w[len(w)-1]
	if resumed, _ := last["rsd"].(bool); !resumed {
		t.Fatal("welcome phải nói đây là VÀO LẠI, để client hiện \"Đang trở lại trận…\"")
	}
	// Mã MỚI, mã cũ hết hiệu lực ngay (§7.1).
	newKey, _ := last["sk"].(string)
	if newKey == "" || newKey == key {
		t.Fatalf("phải cấp mã phiên MỚI: cũ %q, mới %q", key, newKey)
	}
	c3 := NewClient(&recConn{}, false)
	m.Join(c3, netproto.Join{Name: "Teo", Char: "a_hon", Resume: key})
	settle(r, cfg)
	if c3.PlayerID == id {
		t.Fatal("mã cũ phải hết hiệu lực ngay sau khi cấp mã mới")
	}
}

func TestQua20GiayThiNhanVatRoiPhong(t *testing.T) {
	m, cfg := resumeManager(t)
	c, _, key := joinOne(t, m, "Teo")
	r := m.AllRooms()[0]
	z := r.Zone(0)
	id := c.PlayerID
	r.RemoveClient(c)
	settle(r, cfg)

	stepSecs(r, HoldSecs-1)
	if _, ok := z.G.Players[id]; !ok {
		t.Fatal("19 giây: nhân vật vẫn phải còn")
	}
	stepSecs(r, 2)
	if _, ok := z.G.Players[id]; ok {
		t.Fatal("quá 20 giây: nhân vật phải rời phòng như thoát thường")
	}
	if n := m.ResumeCount(); n != 0 {
		t.Fatalf("mã phiên hết hạn phải được quên (còn %d) — §13.9 cấm bảng tra lớn mãi", n)
	}
	// Mã cũ giờ chỉ là một chuỗi vô nghĩa: vào trận mới như người lạ.
	c2 := NewClient(&recConn{}, false)
	m.Join(c2, netproto.Join{Name: "Teo", Char: "a_hon", Resume: key})
	settle(r, cfg)
	if c2.PlayerID == id {
		t.Fatal("mã hết hạn không được gắn lại nhân vật cũ")
	}
	if p := z.G.Players[c2.PlayerID]; p == nil || p.Level != 0 || p.Ring.Count != cfg.Sword.StartCount {
		t.Fatal("hết hạn thì vào trận mới từ Luyện Khí 1, 1 kiếm (§7)")
	}
}

func TestBiChemChetTrongLucMatKetNoiThiThayTongKet(t *testing.T) {
	m, cfg := resumeManager(t)
	c, _, key := joinOne(t, m, "Teo")
	r := m.AllRooms()[0]
	z := r.Zone(0)
	id := c.PlayerID
	r.RemoveClient(c)
	settle(r, cfg)

	// Chém chết nhân vật vô chủ — §7.1 nói rõ "vẫn có thể bị chém chết".
	r.Do(func() { z.G.DebugKill(z.G.Players[id]) })
	stepSecs(r, cfg.Player.DeathAnimSecs+0.2)
	if _, ok := z.G.Players[id]; ok {
		t.Fatal("chết rồi thì nhân vật phải rời thế giới")
	}

	// Tải lại trang xong: thấy màn TỔNG KẾT, không phải trang chủ, không vào trận mới.
	c2 := NewClient(&recConn{}, false)
	conn2 := c2.Conn.(*recConn)
	m.Join(c2, netproto.Join{Name: "Teo", Char: "a_hon", Resume: key})
	settle(r, cfg)
	if len(sentOfKind(t, conn2, "dead")) == 0 {
		t.Fatal("§7.1: bị chém lúc mất kết nối → vào lại phải thấy màn Tổng kết")
	}
	if len(sentOfKind(t, conn2, "welcome")) != 0 {
		t.Fatal("chết rồi thì KHÔNG được tự vào trận mới (§7 v6.29)")
	}
	if c2.PlayerID != 0 {
		t.Fatal("không được gắn vào nhân vật nào")
	}
}

func TestTabThuHaiChiemNhanVat(t *testing.T) {
	m, cfg := resumeManager(t)
	c1, conn1, key := joinOne(t, m, "Teo")
	r := m.AllRooms()[0]
	id := c1.PlayerID

	c2 := NewClient(&recConn{}, false)
	m.Join(c2, netproto.Join{Name: "Teo", Char: "a_hon", Resume: key})
	settle(r, cfg)

	if c2.PlayerID != id {
		t.Fatalf("tab mới phải nhận đúng nhân vật: được %d, muốn %d", c2.PlayerID, id)
	}
	if c1.PlayerID != 0 || c1.Room() != nil {
		t.Fatal("tab cũ phải bị tách khỏi nhân vật")
	}
	kicks := sentOfKind(t, conn1, "kick")
	if len(kicks) == 0 {
		t.Fatal("tab cũ phải được BÁO vì sao mất nhân vật")
	}
	if reason, _ := kicks[len(kicks)-1]["reason"].(string); reason != netproto.KickOtherTab {
		t.Fatalf("lý do phải là mã \"đang chơi ở tab khác\", được %q", reason)
	}
	if r.HeldCount() != 0 {
		t.Fatal("không được để lại chỗ giữ mồ côi")
	}
}

// Dùng lệnh test rồi rút mạng KHÔNG được rửa sạch vết: điểm nặn ra vẫn không lên bảng
// xếp hạng toàn cục (ADMIN_SPEC §1.1).
func TestVetLenhTestTheoQuaLanNoiLai(t *testing.T) {
	m, cfg := resumeManager(t)
	rec := &memRecorder{}
	m.recorder = rec
	c, _, key := joinOne(t, m, "Teo")
	r := m.AllRooms()[0]
	for _, z := range r.zones {
		z.G.Players[c.PlayerID].GuestID = "g1"
	}
	c.DebugUsed = true
	r.RemoveClient(c)
	settle(r, cfg)
	stepSecs(r, HoldSecs+1)
	if len(rec.lives) != 0 {
		t.Fatal("mạng có dùng lệnh test không được ghi, kể cả khi hết hạn chờ nối lại")
	}
	_ = key
}

// Nhân vật vô chủ đứng YÊN: input cuối cùng có thể là "đang chạy".
func TestNhanVatVoChuDungYen(t *testing.T) {
	m, cfg := resumeManager(t)
	c, _, _ := joinOne(t, m, "Teo")
	r := m.AllRooms()[0]
	z := r.Zone(0)
	id := c.PlayerID
	c.SetInput(netproto.Input{Moving: true, Angle: 0})
	settle(r, cfg)
	r.RemoveClient(c)
	settle(r, cfg)
	at := z.G.Players[id].Pos
	stepSecs(r, 3)
	if d := z.G.Players[id].Pos.Dist2(at); d > 1 {
		t.Fatalf("nhân vật vô chủ đi mất %.1f đơn vị", d)
	}

}
