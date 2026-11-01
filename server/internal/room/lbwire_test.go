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
