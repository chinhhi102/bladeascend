package room

import (
	"testing"

	"bladeascend/server/internal/game"
	"bladeascend/server/internal/netproto"
	"bladeascend/server/internal/wire"
)

// lbConn giữ lại gói bảng xếp hạng nhị phân cuối cùng nhận được.
type lbConn struct{ last []byte }

func (c *lbConn) Send(b []byte) bool {
	if len(b) > 0 && b[0] == wire.TypeLbHead {
		c.last = append([]byte(nil), b...)
	}
	return true
}

// decodeLb giải mã một gói [LbHead][LbRow]… như client làm.
func decodeLb(t *testing.T, b []byte) (wire.LbHead, []wire.LbRow) {
	t.Helper()
	if len(b) == 0 || b[0] != wire.TypeLbHead {
		t.Fatal("chưa nhận được gói bảng xếp hạng nhị phân nào")
	}
	head, n, err := wire.DecodeLbHead(b[1:])
	if err != nil {
		t.Fatal(err)
	}
	pos := 1 + n
	var rows []wire.LbRow
	for i := 0; i < int(head.Count); i++ {
		if b[pos] != wire.TypeLbRow {
			t.Fatalf("dòng %d mở đầu bằng byte %d, muốn TypeLbRow", i, b[pos])
		}
		row, m, err := wire.DecodeLbRow(b[pos+1:])
		if err != nil {
			t.Fatal(err)
		}
		rows = append(rows, row)
		pos += 1 + m
	}
	if pos != len(b) {
		t.Fatalf("gói dài %d byte nhưng chỉ đọc %d — thừa dòng hoặc đếm sai `count`", len(b), pos)
	}
	return head, rows
}

// Gói Leaderboard v6.19 (spec §11.5): khu, cảnh giới, tầng, hạng thật của người nhận — và
// người Ẩn Thân thì SERVER giấu khu.
func TestLeaderboardRowsCarryZoneRealmAndHideStealth(t *testing.T) {
	m, cfg := fiveZoneRoom(t)
	connA, connB := &lbConn{}, &lbConn{}
	ca, cb := NewClient(connA, false), NewClient(connB, false)
	m.Join(ca, netproto.Join{Name: "Manh", Char: "a_hon"})
	m.Join(cb, netproto.Join{Name: "An", Char: "cu_rua"})
	r := m.AllRooms()[0]
	r.Step(1 / float64(cfg.Net.TickHz))
	za, zb := splitZones(t, r, ca, cb)
	pa, pb := za.G.Players[ca.PlayerID], zb.G.Players[cb.PlayerID]
	pa.Score, pb.Score = 4931.4, 120
	pa.Level = 2*cfg.Cultivation.TiersPerRealm + 4 // Kim Đan tầng 5
	pb.Buffs[game.BuffStealth] = zb.G.Time + 60
	r.Step(1 / float64(cfg.Net.TickHz)) // thứ tự đã giữ nhận Uy Danh mới (TitleHold = 0)

	r.sendLeaderboards()

	head, rows := decodeLb(t, connB.last)
	if head.Rank != 2 || head.Total != 2 || head.MyScore != 120 || len(rows) != 2 {
		t.Fatalf("đầu gói của B: %+v, %d dòng; muốn hạng 2/2, Uy Danh 120, 2 dòng", head, len(rows))
	}
	top := rows[0]
	if top.ID != ca.PlayerID || top.Rank != 1 || top.Score != 4931 {
		t.Errorf("dòng 1 = %+v; muốn Manh hạng 1, Uy Danh 4931", top)
	}
	if top.Realm != 2 || top.Tier != 5 {
		t.Errorf("cảnh giới/tầng = %d/%d; muốn 2/5 (Kim Đan tầng 5)", top.Realm, top.Tier)
	}
	if int(top.ZoneID) != za.Index || top.Stealth != 0 {
		t.Errorf("khu của Manh = %d (ẩn %d); muốn %d", top.ZoneID, top.Stealth, za.Index)
	}
	if int(top.Character) != indexOf(cfg.Characters, "a_hon") {
		t.Errorf("nhân vật = %d; muốn chỉ số của a_hon", top.Character)
	}
	hidden := rows[1]
	if hidden.ZoneID != lbZoneHidden || hidden.Stealth != 1 {
		t.Errorf("người Ẩn Thân lộ khu: zone=%d stealth=%d — server phải gửi 255, không phải khu thật",
			hidden.ZoneID, hidden.Stealth)
	}
	headA, _ := decodeLb(t, connA.last)
	if headA.Rank != 1 || headA.MyScore != 4931 {
		t.Errorf("đầu gói của A = %+v; muốn hạng 1, Uy Danh 4931", headA)
	}
}

func indexOf(xs []string, s string) int {
	for i, x := range xs {
		if x == s {
			return i
		}
	}
	return -1
}

// Thẻ hạng và danh hiệu dùng CHUNG nhịp giữ 3 s (chủ dự án duyệt 1/11: "thẻ hạng 2 đang mang
// danh hiệu hạng 3"). Hai người đổi hạng qua lại mỗi giây: ở MỌI lần gửi bảng, dòng hạng k
// phải đeo đúng danh hiệu k — không một khoảnh khắc nào lệch.
func TestBoardOrderAndTitlesNeverDisagree(t *testing.T) {
	m, cfg := fiveZoneRoom(t)
	cfg.TitleHold = 3
	conn := &lbConn{}
	ca, cb := NewClient(conn, false), NewClient(&lbConn{}, false)
	m.Join(ca, netproto.Join{Name: "Manh", Char: "a_hon"})
	m.Join(cb, netproto.Join{Name: "Yeu", Char: "a_hon"})
	r := m.AllRooms()[0]
	dt := 1 / float64(cfg.Net.TickHz)
	r.Step(dt)
	za, zb := splitZones(t, r, ca, cb)
	pa, pb := za.G.Players[ca.PlayerID], zb.G.Players[cb.PlayerID]
	pa.InvulnUntil, pb.InvulnUntil = 1e9, 1e9
	ticksPerSec := cfg.Net.TickHz
	for sec := 0; sec < 20; sec++ {
		// Đổi hạng mỗi giây trong 10 s đầu (dồn dập hơn nhịp giữ 3 s), rồi đứng yên.
		if sec < 10 && sec%2 == 0 {
			pa.Score, pb.Score = 200, 100
		} else if sec < 10 {
			pa.Score, pb.Score = 100, 200
		}
		for i := 0; i < ticksPerSec; i++ {
			r.Step(dt)
			r.sendLeaderboards()
			_, rows := decodeLb(t, conn.last)
			for _, row := range rows {
				if int(row.TitleID) != int(row.Rank) {
					t.Fatalf("giây %d: dòng hạng %d (%s) đeo danh hiệu %d — thẻ và danh hiệu lệch nhau",
						sec, row.Rank, row.Name, row.TitleID)
				}
			}
		}
	}
	// Đứng yên từ giây 10 (B 200 > A 100): sau nhịp giữ, B phải là hạng 1.
	_, rows := decodeLb(t, conn.last)
	if len(rows) < 2 || rows[0].ID != cb.PlayerID {
		t.Fatalf("sau khi đứng yên, hạng 1 phải là Yeu (200); bảng = %+v", rows)
	}
}

// heldBoard không bao giờ kẹt: thứ tự đổi sang kiểu MỚI liên tục, nhanh hơn nhịp giữ 3 s,
// thì sau 2×hold vẫn phải nhận. (Đổi qua lại rồi quay về đúng thứ tự đang hiện thì bảng giữ
// nguyên — đó chính là chống nhấp nháy, xem ca thứ hai.)
func TestHeldBoardNeverFreezes(t *testing.T) {
	var b heldBoard
	present := map[uint32]bool{1: true, 2: true, 3: true}
	b.step([]uint32{1, 2, 3}, present, 1, 3)
	churn := [][]uint32{{2, 1, 3}, {2, 3, 1}, {3, 2, 1}, {3, 1, 2}}
	adopted := -1.0
	for i, now := 0, 2.0; now < 20; i, now = i+1, now+1 {
		b.step(churn[i%len(churn)], present, now, 3)
		if adopted < 0 && b.held[0] != 1 {
			adopted = now
		}
	}
	if adopted < 0 || adopted > 2+2*3 {
		t.Fatalf("thứ tự đổi mỗi giây: bảng phải nhận thứ tự mới trong 2×hold = 6 s, nhận lúc %.0f", adopted)
	}

	// Đổi qua lại A↔B mỗi giây, quay về thứ tự đang hiện: bảng đứng yên, không nhấp nháy.
	var c heldBoard
	c.step([]uint32{1, 2}, present, 1, 3)
	for now := 2.0; now < 20; now++ {
		live := []uint32{1, 2}
		if int(now)%2 == 0 {
			live = []uint32{2, 1}
		}
		c.step(live, present, now, 3)
		if c.held[0] != 1 {
			t.Fatalf("giây %.0f: đổi qua lại liên tục mà bảng vẫn đổi theo — đó là nhấp nháy", now)
		}
	}

	// Người rời phòng rút khỏi bảng NGAY, không chờ giữ.
	b.step([]uint32{2}, map[uint32]bool{2: true}, 30, 3)
	if b.rank(1) != 0 || b.rank(2) != 1 {
		t.Fatalf("người đã rời vẫn nằm trên bảng: %v", b.held)
	}
}
