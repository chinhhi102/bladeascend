package main

import (
	"time"

	"bladeascend/server/internal/safe"
	"bladeascend/server/internal/telemetry"
)

// Đưa panic đã bắt của server lên trang quản trị "Lỗi" (GAME_SPEC §13.8).
//
// Dùng CHUNG bảng `events` với lỗi client, loại `server_panic`. Một bảng riêng sẽ cho một
// trang riêng, và trang riêng là trang không ai mở: câu hỏi thật của người trực là "sáng
// nay có gì hỏng không", không phải "có gì hỏng ở phía server không".

// panicSession là `session_id` giả cho mọi panic của server.
//
// Cột đó NOT NULL và trang "Lỗi" đếm "số người dính" theo nó. Một mã cố định nghĩa là mọi
// panic server gộp thành "1 người dính" — đúng, vì chúng đều đến từ một tiến trình.
const panicSession = "server"

// panicFlood: trần số panic ghi xuống DB mỗi phút.
//
// Một panic trong vòng tick bắn 30 lần/giây cho mỗi khu. Không có trần thì 10 giây hỏng là
// 1 500 dòng giống hệt nhau, và DB telemetry — thứ đang phải dọn theo hạn — phình vì đúng
// cái sự cố mà ta muốn đọc gọn trong một dòng.
const panicFlood = 60

// kindFor tách panic GIẢ của lệnh test khỏi panic THẬT.
//
// Cả hai vẫn ghi xuống DB — `crashtest.mjs` nghiệm thu §13.8 bằng đúng sự có mặt của dòng
// giả. Nhưng chúng mang tên khác nhau, nên bảng "Lỗi" và cảnh báo chỉ đếm chuyện thật.
func kindFor(i safe.Info) string {
	if safe.IsFake(i) {
		return telemetry.KindServerPanicFake
	}
	return telemetry.KindServerPanic
}

func (s *server) wirePanicReports() {
	limit := safe.NewRate(panicFlood, time.Minute)
	safe.OnPanic(func(i safe.Info) {
		if s.telemetry == nil {
			return // chạy với -telemetry=false: `safe` đã in ra log, thế là đủ
		}
		if limit.Hit(time.Now()) {
			return
		}
		s.telemetry.AddEvent(&telemetry.Event{
			Session: panicSession,
			TS:      float64(time.Now().UnixMilli()),
			Kind:    kindFor(i),
			Data:    i.Text(),
		})
	})
}
