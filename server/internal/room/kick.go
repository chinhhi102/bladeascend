package room

import (
	"bladeascend/server/internal/netproto"
)

// Kick — ADMIN_SPEC §6.2.
//
// Đây là chỗ DUY NHẤT trong server ngắt kết nối của một người chơi, và nó chỉ chạy khi một
// người bấm nút ở trang quản trị với lý do nhập bằng tay. Không có bộ đếm nào, không có
// goroutine nào quét cờ nghi hack rồi gọi vào đây — §6.1 nói thẳng "chỉ gắn cờ, không tự
// kick", nên hàm này không bao giờ được gọi từ bên trong `internal/room`.

// Kick mời mọi kết nối của `guestID` ra khỏi phòng; trả về số kết nối đã ngắt.
//
// Gửi lý do TRƯỚC khi đóng. Đóng trước rồi gửi là gửi vào chỗ không còn ai nghe — người
// chơi chỉ thấy "mất kết nối", đúng thứ §6.2 muốn tránh.
func (m *Manager) Kick(guestID, reason string) int {
	n := 0
	for _, r := range m.AllRooms() {
		var targets []*Client
		r.Query(func() {
			for _, c := range r.Clients() {
				if c.GuestID == guestID {
					targets = append(targets, c)
				}
			}
		})
		// Gửi và đóng NGOÀI `Query`: `Conn.Send` có thể chặn khi hàng đợi đầy, và chặn ở
		// trong đó là giữ luôn goroutine phòng — cả khu đứng hình vì một kết nối tắc.
		for _, c := range targets {
			if b, err := netproto.Encode(netproto.Kicked{T: netproto.TKicked, Reason: reason}); err == nil {
				c.Conn.Send(b)
			}
			if cl, ok := c.Conn.(interface{ Close() }); ok {
				cl.Close()
			}
			n++
		}
	}
	return n
}
