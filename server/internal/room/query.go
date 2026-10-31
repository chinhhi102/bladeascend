package room

import "time"

// Đọc trạng thái phòng từ goroutine KHÁC (trang quản trị) — an toàn.
//
// Vì sao cần: `Room.Clients()` và `game.PlayerList()` duyệt map. Goroutine phòng ghi vào
// đúng những map đó mỗi khi có người vào hay ra. Duyệt map trong lúc map bị ghi KHÔNG phải
// là một cuộc đua lành tính — Go dừng hẳn chương trình với "concurrent map iteration and
// map write". Tức là: một người chơi bấm Vào trận đúng lúc trang quản trị tự làm mới (2
// giây một lần) là đủ để sập cả server.
//
// Cách sửa không phải là thêm mutex quanh từng map: goroutine phòng chạy 20 lần/giây và
// mọi lần khoá thêm đều rơi vào ngân sách tick 4 ms (§13.6). Thay vào đó, đẩy HÀM ĐỌC sang
// chạy trên chính goroutine phòng, giữa hai tick, rồi chờ kết quả. Goroutine phòng không
// phải khoá gì cả, và người đọc chờ nhiều nhất một tick.

// queryTimeout: chờ lâu hơn ngần này thì bỏ cuộc.
//
// Một tick là 50 ms ở 20 Hz. Để 2 giây là rộng rãi gấp nhiều lần, nhưng vẫn hữu hạn — nếu
// goroutine phòng kẹt thật thì trang quản trị phải báo "không đọc được", chứ không phải
// treo luôn cái tab, vì lúc đó người ta cần trang quản trị nhất.
const queryTimeout = 2 * time.Second

// Query chạy `fn` TRÊN goroutine phòng và chờ xong.
//
// Trả về false khi phòng đang quá tải hoặc đã đóng — người gọi bỏ qua phòng đó chứ không
// chờ mãi. Đừng gọi từ bên trong `fn` của một `Do`/`Query` khác: nó sẽ tự chờ chính mình.
func (r *Room) Query(fn func()) bool {
	done := make(chan struct{})
	select {
	case r.inbox <- func() { fn(); close(done) }:
	default:
		return false // hộp thư đầy: phòng đang quá tải, bỏ qua chứ không làm nó tệ thêm
	}
	select {
	case <-done:
		return true
	case <-time.After(queryTimeout):
		return false
	}
}

// ClientsSafe trả về bản sao danh sách client, đọc an toàn từ goroutine khác.
//
// `Clients()` KHÔNG an toàn và chỉ dùng được từ trong goroutine phòng; giữ nguyên nó vì
// các chỗ gọi bên trong phòng không nên phải trả giá cho một lần đổi goroutine.
func (r *Room) ClientsSafe() []*Client {
	var out []*Client
	if !r.Query(func() { out = r.Clients() }) {
		return nil
	}
	return out
}

// ClientList trả về client của khu này. Chỉ gọi từ trong goroutine phòng (tức là trong
// `Do`/`Query`) — xem đầu tệp.
func (z *Zone) ClientList() []*Client {
	out := make([]*Client, 0, len(z.clients))
	for _, c := range z.clients {
		out = append(out, c)
	}
	return out
}
