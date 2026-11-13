package main

// Tắt êm khi nhận SIGTERM — DEPLOY_K8S_SPEC §1.3.

import (
	"log"
	"time"
)

// shutdownDeadline: hạn chót cho cả quá trình tắt.
//
// Spec cho `terminationGracePeriodSeconds: 60` và đòi thoát trong ≤ 45 s. Để 40 s: chừa
// một khoảng cho phần việc sau hàm này (đóng telemetry, đóng HTTP) vẫn nằm trong 45.
const shutdownDeadline = 40 * time.Second

// gracefulShutdown tiễn người chơi rồi khoá sổ dữ liệu.
//
// Thứ tự có chủ ý:
//  1. `/readyz` trả 503 NGAY — Ingress ngừng gửi người mới tới trong lúc ta còn đang dọn.
//     Làm bước này sau cùng thì người vừa vào sẽ bị tiễn ra ngay, tệ hơn là không vào được.
//  2. Tiễn người đang chơi: màn Tổng kết + điểm + câu "đang cập nhật" (`Room.ShutdownDrain`).
//  3. Chốt SQLite (checkpoint WAL). Bỏ bước này thì dữ liệu vẫn còn trong tệp `-wal`, và nó
//     chỉ được gộp vào lần mở sau — mà lần mở sau là một pod khác, trên cùng một PVC, nên
//     thường thì không sao. "Thường thì" không đủ cho dữ liệu người chơi.
func (s *server) gracefulShutdown() {
	start := time.Now()
	s.health.startDrain("đang tắt để cập nhật")
	log.Print("tắt êm: /readyz đã trả 503, ngừng nhận người mới")

	done := make(chan int, 1)
	go func() {
		n := 0
		for _, rm := range s.mgr.AllRooms() {
			n += rm.ShutdownDrain()
		}
		done <- n
	}()
	select {
	case n := <-done:
		log.Printf("tắt êm: đã tiễn %d người chơi", n)
	case <-time.After(shutdownDeadline / 2):
		// Phòng treo thì không chờ thêm: thà mất màn Tổng kết của vài người còn hơn bị
		// SIGKILL giữa lúc đang ghi SQLite.
		log.Print("⚠ tắt êm: hết hạn chờ tiễn người chơi, đi tiếp để kịp khoá sổ dữ liệu")
	}

	// Cho các gói vừa gửi kịp ra khỏi socket. Không có nhịp này thì câu "đang cập nhật"
	// nằm trong bộ đệm và chết cùng tiến trình.
	time.Sleep(500 * time.Millisecond)

	if err := s.store.Checkpoint(); err != nil {
		log.Printf("⚠ tắt êm: checkpoint SQLite lỗi: %v", err)
	} else {
		log.Print("tắt êm: đã checkpoint SQLite")
	}
	log.Printf("tắt êm: xong sau %.1f s", time.Since(start).Seconds())
}
