package room

import (
	"testing"

	"kiemxoay/server/internal/netproto"
	"kiemxoay/server/internal/wire"
)

// Bộ giải mã gói nhị phân DÙNG CHO TEST — Protocol v2.
//
// Viết bản giải mã riêng ở đây chứ không gọi lại hàm mã hoá là chủ ý: nếu test dùng chính
// mã của bên gửi để kiểm bên gửi thì nó chỉ chứng minh "hàm bằng chính nó". Bản này đi
// ngược từ byte lên, đúng như client phải làm.
//
// Đối chiếu với client thật thì đã có `cmd/protofixture` + `npm run check:wire` (10 031 gói,
// cả hai chiều). Chỗ này lo phần khác: KHUNG gói — thứ tự message, có lẫn byte thừa không,
// delta có đúng không.


// joinWith đưa một client vào phòng, tắt bot cho kết quả ổn định.
//
// Bot bật thì số entity trong tầm nhìn dao động theo từng lượt chạy, và mấy bài kiểm dưới
// đây đếm số gói — chúng sẽ đỏ ngẫu nhiên vì lý do chẳng liên quan gì tới giao thức.
func joinWith(m *Manager, c *Client, name string) *Room {
	m.Join(c, netproto.Join{Name: name, Char: "a_hon"})
	return m.AllRooms()[0]
}

// decodedSnap là nội dung một gói sau khi mở ra.
type decodedSnap struct {
	Head       wire.SnapHead
	Spawns     []wire.Spawn
	Descs      []wire.Desc
	Updates    []wire.Update
	Despawns   []wire.Despawn
	Monsters   []wire.Monster
	ItemsAdded []wire.ItemAdd
	ItemsDel   []wire.ItemDel
	Events     []wire.Event
	SelfState  []wire.SelfState
	SelfExtra  []wire.SelfExtra
}

// decodeSnap mở một gói nhị phân. Lỗi ở đây là lỗi THẬT: gói server gửi mà không đọc nổi
// nghĩa là client cũng không đọc nổi.
func decodeSnap(t *testing.T, b []byte) (decodedSnap, bool) {
	t.Helper()
	var out decodedSnap
	if len(b) == 0 || b[0] != wire.TypeSnapHead {
		return out, false // không phải snapshot (welcome, leaderboard… vẫn đi msgpack)
	}
	p := 1
	h, n, err := wire.DecodeSnapHead(b[p:])
	if err != nil {
		t.Fatalf("SnapHead hỏng: %v", err)
	}
	out.Head = h
	p += n

	for p < len(b) {
		typ := b[p]
		p++
		var n int
		var err error
		switch typ {
		case wire.TypeSpawn:
			var v wire.Spawn
			v, n, err = wire.DecodeSpawn(b[p:])
			out.Spawns = append(out.Spawns, v)
		case wire.TypeDesc:
			var v wire.Desc
			v, n, err = wire.DecodeDesc(b[p:])
			out.Descs = append(out.Descs, v)
		case wire.TypeUpdate:
			var v wire.Update
			v, n, err = wire.DecodeUpdate(b[p:])
			out.Updates = append(out.Updates, v)
		case wire.TypeDespawn:
			var v wire.Despawn
			v, n, err = wire.DecodeDespawn(b[p:])
			out.Despawns = append(out.Despawns, v)
		case wire.TypeMonster:
			var v wire.Monster
			v, n, err = wire.DecodeMonster(b[p:])
			out.Monsters = append(out.Monsters, v)
		case wire.TypeItemAdd:
			var v wire.ItemAdd
			v, n, err = wire.DecodeItemAdd(b[p:])
			out.ItemsAdded = append(out.ItemsAdded, v)
		case wire.TypeItemDel:
			var v wire.ItemDel
			v, n, err = wire.DecodeItemDel(b[p:])
			out.ItemsDel = append(out.ItemsDel, v)
		case wire.TypeEvent:
			var v wire.Event
			v, n, err = wire.DecodeEvent(b[p:])
			out.Events = append(out.Events, v)
		case wire.TypeSelfState:
			var v wire.SelfState
			v, n, err = wire.DecodeSelfState(b[p:])
			out.SelfState = append(out.SelfState, v)
		case wire.TypeSelfExtra:
			var v wire.SelfExtra
			v, n, err = wire.DecodeSelfExtra(b[p:])
			out.SelfExtra = append(out.SelfExtra, v)
		default:
			t.Fatalf("loại message lạ %d ở vị trí %d — khung gói hỏng, "+
				"client sẽ đọc rác từ đây trở đi", typ, p-1)
		}
		if err != nil {
			t.Fatalf("giải mã loại %d: %v", typ, err)
		}
		p += n
	}
	return out, true
}

// snapsOf mở mọi gói snapshot trong danh sách đã gửi.
func snapsOf(t *testing.T, sent [][]byte) []decodedSnap {
	t.Helper()
	var out []decodedSnap
	for _, raw := range sent {
		if s, ok := decodeSnap(t, raw); ok {
			out = append(out, s)
		}
	}
	return out
}

// Gói phải ĐỌC HẾT, không thừa byte nào.
//
// Thừa byte là dấu hiệu hai phía không cùng số trường, và client sẽ đọc message tiếp theo
// từ giữa chừng — không lỗi, chỉ rác. Vòng lặp trong `decodeSnap` tự bắt (nó sẽ gặp một
// "loại message lạ"), nên test này chỉ cần chạy một trận thật rồi mở hết gói ra.
func TestEverySnapshotDecodesCleanly(t *testing.T) {
	m, _, cfg := testManager(t)
	cfg.Room.BotFillTarget = 0
	conn := &recConn{}
	c := NewClient(conn, false)
	r := joinWith(m, c, "NguoiChoi")
	dt := 1 / float64(cfg.Net.TickHz)
	for i := 0; i < 120; i++ {
		r.Step(dt)
	}
	snaps := snapsOf(t, conn.sent)
	if len(snaps) == 0 {
		t.Fatal("không gói snapshot nào được gửi — test không chạm tới thứ nó định kiểm")
	}
	// Ít nhất một gói phải là keyframe (mốc 2 giây, 120 tick ở 20 Hz = 6 giây).
	var keyframes int
	for _, s := range snaps {
		if s.Head.Flags&1 != 0 {
			keyframes++
		}
	}
	if keyframes < 2 {
		t.Errorf("6 giây mà chỉ %d keyframe — mốc %gs không chạy; mất một gói là client kẹt "+
			"với trạng thái sai cho tới khi entity đổi hết mọi trường",
			keyframes, keyframeEverySecs)
	}
	t.Logf("%d gói, %d keyframe, gói đầu %d byte", len(snaps), keyframes, len(conn.sent[0]))
}

// Entity ĐỨNG YÊN không được tốn byte nào.
//
// Đây là lời hứa trung tâm của delta (§4.3). Nếu nó hỏng thì mọi thứ vẫn CHẠY ĐÚNG — chỉ
// tốn băng thông — nên không có bài kiểm nào khác bắt được.
func TestStandingStillCostsNothing(t *testing.T) {
	m, _, cfg := testManager(t)
	cfg.Room.BotFillTarget = 0
	conn := &recConn{}
	c := NewClient(conn, false)
	r := joinWith(m, c, "DungYen")
	dt := 1 / float64(cfg.Net.TickHz)
	// Một tick TRƯỚC khi lấy người chơi: `Join` chỉ xếp hàng, nhân vật được dựng ở tick kế
	// tiếp. Lấy sớm hơn thì `g.Players[id]` là nil.
	r.Step(dt)
	p := r.Zone(0).G.Players[c.PlayerID]
	p.InvulnUntil = 1e9

	// Chạy cho qua keyframe đầu rồi giữ nguyên mọi thứ.
	for i := 0; i < 5; i++ {
		r.Step(dt)
	}
	// Đóng băng: không input, không buff, vòng kiếm ngừng quay.
	p.Ring.Dir = 0
	before := len(conn.sent)
	for i := 0; i < 10; i++ {
		p.Ring.Angle = 0 // giữ pha đứng yên: bình thường nó quay nên luôn có bit RingPhase
		r.Step(dt)
	}
	var updatesForMe int
	for _, s := range snapsOf(t, conn.sent[before:]) {
		if s.Head.Flags&1 != 0 {
			continue // keyframe chở lại toàn bộ, đúng theo thiết kế
		}
		for _, u := range s.Updates {
			if u.ID == c.PlayerID && u.Changed != 0 {
				updatesForMe++
			}
		}
	}
	if updatesForMe > 2 {
		t.Errorf("đứng yên 10 tick mà vẫn gửi %d Update — delta không ăn; "+
			"xem `diffPlayer` có so ở dạng ĐÃ LƯỢNG TỬ không", updatesForMe)
	}
}

// Lượng tử hoá không được làm sai lệch quá mức §4.6 hứa.
func TestQuantisationStaysWithinBudget(t *testing.T) {
	for _, in := range []float64{0, 0.4, 123.7, 5999.5, 11999.9} {
		got := float64(netproto.QuantPos(in))
		if d := got - in; d > 1 || d < -1 {
			t.Errorf("vị trí %.2f → %.0f, lệch %.2f (trần 1 đơn vị)", in, got, d)
		}
	}
	// Góc: 1/256 vòng = 1.4°, mắt không thấy trên một nhân vật.
	for _, rad := range []float64{0, 1, 3.14159, 6.2831} {
		back := float64(quantAngle(rad)) / 256 * 2 * 3.14159265358979
		if d := back - rad; d > 0.025 || d < -0.025 {
			t.Errorf("góc %.4f rad → %.4f, lệch %.4f (trần 0.025 = 1.4°)", rad, back, d)
		}
	}
	// Pha vòng kiếm cần 16 bit: client NỘI SUY giữa hai snapshot và 1/256 vòng ở 20 Hz là
	// thấy giật. Kiểm nó chặt hơn góc nhìn 100 lần.
	for _, rad := range []float64{0, 1, 3.14159, 6.2831} {
		back := float64(quantAngle16(rad)) / 65536 * 2 * 3.14159265358979
		if d := back - rad; d > 0.0002 || d < -0.0002 {
			t.Errorf("pha vòng kiếm %.5f rad → %.5f, lệch %.5f (trần 0.0002)", rad, back, d)
		}
	}
}
