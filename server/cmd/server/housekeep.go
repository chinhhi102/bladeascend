package main

import (
	"log"

	"kiemxoay/server/internal/admin"
	"time"
)

// Việc dọn dẹp chạy hằng ngày — ADMIN_SPEC §9, QUESTIONS Q28.
//
// Một goroutine, chạy mỗi 24 giờ, và **ghi lại mọi thứ nó xoá vào nhật ký quản trị**. Một
// việc tự động âm thầm xoá dữ liệu là thứ không ai phát hiện ra cho tới khi cần đúng dữ
// liệu đó — nên nó phải để lại dấu như bất kỳ thao tác tay nào.

// housekeepEvery: nhịp chạy. Mỗi ngày là đủ cho một hạn tính bằng tháng.
const housekeepEvery = 24 * time.Hour

// adminKeepDays nói lại hạn giữ ảnh, để log và nhật ký không chép tay con số.
func adminKeepDays() int { return admin.FeedbackImageKeepDays }

func (s *server) runHousekeeper(stop <-chan struct{}) {
	// Chạy một lần sau 1 phút rồi mới vào nhịp ngày: server khởi động lại nhiều lần trong
	// lúc làm việc, và chờ nguyên 24 giờ nghĩa là việc dọn thực tế không bao giờ chạy.
	t := time.NewTimer(time.Minute)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			s.housekeepOnce(time.Now())
			t.Reset(housekeepEvery)
		}
	}
}

// housekeepOnce làm một lượt dọn. Tách khỏi vòng lặp để test gọi thẳng được.
func (s *server) housekeepOnce(now time.Time) {
	st := s.adminAuth.Store()
	// Ảnh góp ý của thư đã khép lại quá 90 ngày (Q28 phương án B: dọn ẢNH, giữ CHỮ mãi).
	n, err := st.PurgeOldImages(feedbackDir(s.dbPath), now)
	if err != nil {
		log.Printf("dọn ảnh góp ý: %v", err)
		return
	}
	if n == 0 {
		return // không có gì để dọn: đừng làm bẩn nhật ký bằng những dòng "đã xoá 0"
	}
	log.Printf("dọn ảnh góp ý: xoá %d ảnh của thư đã khép quá %d ngày", n, adminKeepDays())
	st.Audit("purge_feedback_images", "", "dọn tự động hằng ngày", "system",
		map[string]any{"deleted": n, "keep_days": adminKeepDays()})
}
