package room

import (
	"testing"

	"bladeascend/server/internal/game"
	"bladeascend/server/internal/netproto"
)

// Lệnh test THIẾU tham số phải LỖI, không được im lặng hiểu thành một lệnh khác
// (GAME_SPEC §13.9 v6.58, mục 4).
//
// Vì sao có bài test này: `netproto.Debug.N` từng là `int`, nên gói thiếu `n` giải mã ra 0 —
// một chỉ số khu hợp lệ. Bài rò rỉ gọi `dbg('zone')` không tham số, server hiểu là "sang khu
// 0", và lệnh bỏ qua khi đã ở khu 0. Kết quả: SÁU lượt đo 30–52 phút tự nhận là "50 lượt đổi
// khu" trong khi gần như không đổi khu lần nào, và không một dòng lỗi nào. Đây là bài test
// canh đúng cái im lặng đó.
func TestDebugZoneRequiresN(t *testing.T) {
	m, cfg := zoneManager(t)
	cs := joinN(m, 1)
	r := m.AllRooms()[0]
	settle(r, cfg)
	c := cs[0]

	zoneOf := func() int {
		z := r.zoneOf(c.PlayerID)
		if z == nil {
			t.Fatal("người chơi không còn trong phòng nào")
		}
		return z.Index
	}
	start := zoneOf()

	// THIẾU n: không được đi đâu cả.
	r.Debug(c, netproto.Debug{T: netproto.TDebug, Cmd: "zone"})
	stepSecs(r, 0.5)
	if got := zoneOf(); got != start {
		t.Fatalf("`zone` thiếu tham số đã bê người từ khu %d sang khu %d — phải không làm gì", start, got)
	}

	// CÓ n: phải sang đúng khu đó.
	want := (start + 2) % cfg.Room.Zones
	n := want
	r.Debug(c, netproto.Debug{T: netproto.TDebug, Cmd: "zone", N: &n})
	stepSecs(r, 0.5)
	if got := zoneOf(); got != want {
		t.Fatalf("`zone` n=%d đưa tới khu %d, phải là khu %d", want, got, want)
	}

	// Khu KHÔNG TỒN TẠI: cũng không làm gì (mặc định cũ là "khu kế tiếp" — đúng kiểu đoán
	// mà mục 4 của v6.58 cấm).
	bad := cfg.Room.Zones + 7
	r.Debug(c, netproto.Debug{T: netproto.TDebug, Cmd: "zone", N: &bad})
	stepSecs(r, 0.5)
	if got := zoneOf(); got != want {
		t.Fatalf("`zone` với khu %d (không tồn tại) đã bê người sang khu %d — phải không làm gì", bad, got)
	}
}

// `invuln` và `nogate` thiếu n thì trước đây thành "0 giây", tức một lệnh rỗng im lặng:
// bài đo tưởng đã bật, mà thật ra nhân vật vẫn bị bot hạ và vẫn bị cổng hút đi.
func TestDebugInvulnAndNogateRequireN(t *testing.T) {
	m, cfg := zoneManager(t)
	cs := joinN(m, 1)
	r := m.AllRooms()[0]
	settle(r, cfg)
	c := cs[0]

	// Đọc thẳng, không qua `r.Do`: trong bài test phòng KHÔNG chạy goroutine riêng, hộp thư
	// chỉ được rút khi `stepSecs` tick — nên một closure gửi qua `r.Do` rồi đọc biến ngay
	// sau đó luôn thấy giá trị cũ, và bài test sẽ "đạt" cả khi lệnh chạy đúng.
	player := func() *game.Player {
		z := r.zoneOf(c.PlayerID)
		if z == nil {
			t.Fatal("người chơi không còn trong phòng nào")
		}
		return z.G.Players[c.PlayerID]
	}
	read := func() (float64, float64) {
		p := player()
		return p.InvulnUntil, p.GateReadyAt
	}
	// Dọn sạch mốc do lúc vào trận đặt, để phép so dưới đây chỉ nói về lệnh test.
	//
	// `r.Debug` xếp việc vào hộp thư phòng, nên phải cho phòng tick một nhịp rồi mới đọc —
	// đọc ngay thì bài test luôn thấy 0 và luôn "đạt" cả khi lệnh chạy đúng.
	p0 := player()
	p0.InvulnUntil, p0.GateReadyAt = 0, 0

	r.Debug(c, netproto.Debug{T: netproto.TDebug, Cmd: "invuln"})
	r.Debug(c, netproto.Debug{T: netproto.TDebug, Cmd: "nogate"})
	stepSecs(r, 0.2)
	if invuln, gate := read(); invuln != 0 || gate != 0 {
		t.Fatalf("lệnh thiếu n vẫn đặt mốc (invuln %.1f, nogate %.1f) — phải không làm gì", invuln, gate)
	}

	n := 120
	r.Debug(c, netproto.Debug{T: netproto.TDebug, Cmd: "invuln", N: &n})
	r.Debug(c, netproto.Debug{T: netproto.TDebug, Cmd: "nogate", N: &n})
	stepSecs(r, 0.2)
	invuln, gate := read()
	if invuln < 100 || gate < 100 {
		t.Fatalf("lệnh có n=120 chưa đặt mốc (invuln %.1f, nogate %.1f)", invuln, gate)
	}
}
