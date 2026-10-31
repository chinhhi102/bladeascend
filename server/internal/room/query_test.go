package room

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"kiemxoay/server/internal/netproto"
)

// Trang quản trị đọc song song với người vào/ra — ADMIN_SPEC §6, xem query.go.
//
// Chạy test này dưới `-race` (và cả không có -race). Nó chống lại một lỗi làm SẬP SERVER,
// không phải một lỗi hiển thị sai: `Room.Clients()` duyệt `z.clients`, goroutine phòng ghi
// vào đúng map đó mỗi khi có người vào hay ra, và Go dừng hẳn chương trình khi gặp
// "concurrent map iteration and map write". Một người bấm Vào trận đúng nhịp làm mới 2
// giây của trang quản trị là đủ.
//
// Đã thử bằng cách đổi `ClientsSafe()` thành `Clients()` trong test này: -race báo ngay.
func TestAdminReadsWhilePlayersComeAndGo(t *testing.T) {
	m, _, _ := testManager(t)

	// Giả lập goroutine phòng: nó là thứ sở hữu `z.clients`, và là thứ chạy các hàm gửi
	// vào `inbox`. Không có nó thì `Query` chờ mãi — đúng như ngoài đời khi phòng treo.
	var r *Room
	joinOne := func(i int) *Client {
		c := NewClient(&fakeConn{}, false)
		rm := m.Join(c, netproto.Join{Name: fmt.Sprintf("P%d", i), Char: "a_hon"})
		if r == nil {
			r = rm
		}
		return c
	}
	first := joinOne(0)
	_ = first

	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				r.Step(1.0 / 20)
				time.Sleep(time.Millisecond)
			}
		}
	}()

	// Người vào và ra liên tục.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i < 120; i++ {
			c := joinOne(i)
			time.Sleep(time.Millisecond)
			r.RemoveClient(c)
		}
	}()

	// Trang quản trị đọc liên tục, đúng kiểu `handleAdminPlayers` và `collectStats` làm.
	reads, withClients := 0, 0
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cs := r.ClientsSafe(); cs != nil {
			reads++
			if len(cs) > 0 {
				withClients++
			}
			// Đụng vào từng client như bảng §6 làm: đọc cờ và số mạng.
			for _, c := range cs {
				c.Flags.List(time.Now())
				c.Net()
			}
		}
		for i := 0; i < r.ZoneCount(); i++ {
			z := r.Zone(i)
			r.Query(func() { _ = len(z.ClientList()) })
		}
	}
	close(stop)
	wg.Wait()

	if reads == 0 {
		t.Fatal("không đọc được lần nào — Query hỏng hoặc goroutine phòng không chạy")
	}
	// Nếu mọi lần đọc đều rơi vào lúc phòng rỗng thì test chẳng chứng minh được gì: cuộc
	// đua chỉ xảy ra khi có client để duyệt.
	if withClients == 0 {
		t.Fatalf("%d lần đọc nhưng lần nào cũng rỗng — test không chạm tới chỗ cần kiểm", reads)
	}
	t.Logf("%d lần đọc, %d lần có client, không có cuộc đua nào", reads, withClients)
}

// Query phải BỎ CUỘC chứ không treo khi goroutine phòng không chạy.
//
// Trang quản trị là thứ người ta mở ĐÚNG LÚC server có vấn đề; nếu nó treo theo thì nó vô
// dụng đúng lúc cần nhất.
func TestQueryGivesUpWhenRoomStalls(t *testing.T) {
	m, _, _ := testManager(t)
	c := NewClient(&fakeConn{}, false)
	r := m.Join(c, netproto.Join{Name: "P", Char: "a_hon"})

	// Không ai chạy `Step`, nên không ai rút `inbox`.
	start := time.Now()
	ok := r.Query(func() { t.Error("không được chạy: không có goroutine phòng") })
	took := time.Since(start)

	if ok {
		t.Error("Query báo thành công trong khi phòng không chạy")
	}
	if took < queryTimeout {
		t.Errorf("bỏ cuộc quá sớm (%v) — chờ ít hơn một tick là bỏ nhầm phòng đang bận", took)
	}
	if took > queryTimeout*2 {
		t.Errorf("treo quá lâu (%v): trang quản trị sẽ đứng hình", took)
	}
	if cs := r.ClientsSafe(); cs != nil {
		t.Error("ClientsSafe phải trả nil khi không đọc được, không trả danh sách rỗng giả")
	}
}
