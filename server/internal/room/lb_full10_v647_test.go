package room

import (
	"testing"

	"bladeascend/server/internal/config"
	"bladeascend/server/internal/netproto"
)

// GAME_SPEC §11.5 (v6.47) — "Server LUÔN gửi đủ top 10 + hạng của người nhận trong mỗi gói
// Leaderboard (không gửi phần thay đổi / danh sách tạm thiếu người khi đang giữ 3 s)."
//
// Đây là nửa SERVER của lỗi 07/11 "bảng tụt còn 3 dòng rồi bung lại 10". `heldBoard` rút
// người chết/rời khỏi thứ tự đang hiện NGAY (đúng — giữ tên người đã đi là sai), nhưng thứ
// tự MỚI thì phải chờ đứng yên 3 s. Khoảng giữa hai việc đó, `held` ngắn hơn 10 và gói đi ra
// đúng y số dòng ngắn ấy.

// churnRoom dựng phòng 5 khu với `n` người và nhịp giữ 3 s THẬT (khác `fiveZoneRoom`, nó đặt
// TitleHold = 0 để kiểm phần danh hiệu).
func churnRoom(t *testing.T, n int) (*Room, []*lbConn) {
	t.Helper()
	m, cfg := fiveZoneRoom(t)
	cfg.TitleHold = 3
	conns := make([]*lbConn, n)
	for i := range conns {
		conns[i] = &lbConn{}
		c := NewClient(conns[i], false)
		m.Join(c, netproto.Join{Name: "Bot", Char: "a_hon"})
	}
	r := m.AllRooms()[0]
	r.Step(1 / float64(cfg.Net.TickHz))
	// Uy Danh khác nhau để thứ tự xác định, không phụ thuộc thứ tự vào phòng.
	for i, p := range r.Ranking() {
		p.Score = float64(1000 - 10*i)
	}
	return r, conns
}

// rowsNow gửi một vòng bảng rồi trả số dòng trong gói của người đầu tiên.
func rowsNow(t *testing.T, r *Room, c *lbConn) int {
	t.Helper()
	r.sendLeaderboards()
	_, rows := decodeLb(t, c.last)
	return len(rows)
}

func TestLeaderboardAlwaysSendsFullTop10(t *testing.T) {
	cfg := config.MustLoadDefault()
	const n = 20
	r, conns := churnRoom(t, n)
	dt := 1 / float64(cfg.Net.TickHz)

	if got := rowsNow(t, r, conns[0]); got != leaderboardTop {
		t.Fatalf("phòng %d người, gói đầu chỉ có %d dòng; muốn %d", n, got, leaderboardTop)
	}

	// Giết lần lượt 7 người trên bảng — mỗi lần cách nhau nửa giây, tức luôn nằm trong cửa
	// sổ giữ 3 s của lần trước. Đây đúng là cảnh chủ dự án chụp được: một pha giao tranh ở
	// top làm bảng rụng dần.
	worst := leaderboardTop
	for killed := 0; killed < 7; killed++ {
		top := r.Ranking()
		for _, z := range r.zones {
			if p, ok := z.G.Players[top[0].ID]; ok {
				p.Dead, p.DiedAt = true, z.G.Time
			}
		}
		for i := 0; i < int(cfg.Net.TickHz)/2; i++ {
			r.Step(dt)
			if got := rowsNow(t, r, conns[len(conns)-1]); got < worst {
				worst = got
			}
		}
	}

	alive := 0
	for _, p := range r.Ranking() {
		_ = p
		alive++
	}
	want := min(leaderboardTop, alive)
	if worst < want {
		t.Errorf("giữa pha giao tranh gói chỉ còn %d dòng (còn %d người sống); spec nói luôn đủ %d",
			worst, alive, want)
	}
}

// Chết GIỮA hai mốc: thứ tự được giữ ở đầu tick, bảng được gửi muộn hơn trong tick. Ai chết
// trong khoảng đó rơi khỏi `Ranking()` nên không còn con trỏ `*Player` để mã hoá — bản đầu
// của v6.47 bỏ qua họ và gửi 9 dòng. Cổng deploy bắt được đúng ca này (11/11).
func TestLeaderboardFullEvenWhenSomeoneDiesMidTick(t *testing.T) {
	r, conns := churnRoom(t, 20)
	if got := rowsNow(t, r, conns[0]); got != leaderboardTop {
		t.Fatalf("gói đầu có %d dòng, muốn %d", got, leaderboardTop)
	}

	// Giết 3 người trên bảng mà KHÔNG bước nhịp: `heldBoard` chưa biết gì.
	for i, p := range r.Ranking() {
		if i >= 3 {
			break
		}
		for _, z := range r.zones {
			if q, ok := z.G.Players[p.ID]; ok {
				q.Dead, q.DiedAt = true, z.G.Time
			}
		}
	}
	if got := rowsNow(t, r, conns[len(conns)-1]); got != leaderboardTop {
		t.Errorf("3 người chết giữa tick → gói còn %d dòng, spec nói luôn đủ %d", got, leaderboardTop)
	}
}

// Phòng ít hơn 10 người: gói gửi đúng số người CÓ THẬT — "luôn đủ top 10" nói về việc không
// bỏ sót ai, không phải đệm dòng rỗng. Ô trống là việc của client (§11.5).
func TestLeaderboardSendsEveryoneWhenRoomSmall(t *testing.T) {
	r, conns := churnRoom(t, 4)
	if got := rowsNow(t, r, conns[0]); got != 4 {
		t.Fatalf("phòng 4 người mà gói có %d dòng; muốn 4", got)
	}
}
