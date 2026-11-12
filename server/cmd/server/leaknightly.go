package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Kết quả bài kiểm rò rỉ HẰNG ĐÊM — GAME_SPEC §13.9 (v6.54).
//
// `tools/leaknightly.sh` chạy lúc 03:00 và ghi một tệp JSON; server chỉ ĐỌC nó. Vì sao
// không cho server tự chạy bài đo: bài client cần Chrome headless và Playwright, tức là
// cần cả một môi trường Node — nhét thứ đó vào tiến trình đang phục vụ người chơi là đổi
// bản chất của tiến trình đó. Một tệp JSON là ranh giới đủ rộng và đủ rõ.
//
// Việc duy nhất server GHI vào tệp là cờ `ack`: "đã có người xem lại". Đó phải là hành
// động của một người đang đăng nhập trang quản trị, không phải một biến môi trường gõ cho
// qua — nếu không thì cổng chặn deploy chỉ là một chướng ngại cần gõ thêm một chữ.

// LeakNightly là nội dung `server/data/leak-nightly.json`.
type LeakNightly struct {
	At int64 `json:"at"`
	OK bool  `json:"ok"`
	// Skipped: lý do bỏ qua đêm đó (có người chơi, máy bận). Rỗng = đã chạy thật.
	Skipped string `json:"skipped"`
	Note    string `json:"note"`
	Log     string `json:"log"`
	// Ack: đã có người xem lại kết quả đỏ. Chỉ trang quản trị đặt được.
	Ack           bool   `json:"ack"`
	AckAt         int64  `json:"ack_at,omitempty"`
	ClientMinutes int    `json:"client_minutes"`
	ServerMinutes int    `json:"server_minutes"`
	Path          string `json:"-"`
}

// Stale: kết quả cũ hơn 48 giờ. Bài chạy hằng đêm, nên quá hai đêm không có kết quả nghĩa
// là cron hỏng — và một cổng im lặng ngừng chạy còn tệ hơn một cổng đỏ.
func (l *LeakNightly) Stale(now time.Time) bool {
	return l.At > 0 && now.Unix()-l.At > 48*3600
}

// Warning là câu cảnh báo cho dashboard; rỗng = không có gì phải nói.
func (l *LeakNightly) Warning(now time.Time) string {
	switch {
	case l == nil:
		return ""
	case !l.OK && !l.Ack:
		return "Bài kiểm rò rỉ bộ nhớ đêm qua ĐỎ: " + l.Note + " — deploy kế tiếp bị chặn tới khi xem lại."
	case l.Stale(now):
		return "Bài kiểm rò rỉ chưa chạy quá 48 giờ — kiểm lại mục cron (tools/leaknightly.sh --install)."
	}
	return ""
}

func leakNightlyPath(dbPath string) string {
	return filepath.Join(filepath.Dir(dbPath), "leak-nightly.json")
}

// readLeakNightly đọc tệp kết quả. Trả nil khi chưa chạy lần nào hoặc tệp hỏng.
//
// Tệp hỏng cũng trả nil chứ không báo lỗi: nó được một script shell ghi, và một lần ghi
// dở dang không được phép làm trang quản trị không mở được.
func readLeakNightly(dbPath string) *LeakNightly {
	p := leakNightlyPath(dbPath)
	b, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	var l LeakNightly
	if json.Unmarshal(b, &l) != nil {
		return nil
	}
	l.Path = p
	return &l
}

// handleAdminLeakAck đánh dấu "đã xem lại" để deploy kế tiếp đi tiếp được.
func (s *server) handleAdminLeakAck(w http.ResponseWriter, r *http.Request) {
	l := readLeakNightly(s.dbPath)
	if l == nil {
		http.Error(w, "chưa có kết quả nào", http.StatusNotFound)
		return
	}
	l.Ack, l.AckAt = true, time.Now().Unix()
	b, err := json.MarshalIndent(l, "", "  ")
	if err == nil {
		// Ghi qua tệp tạm rồi đổi tên: `tools/deploy.sh` có thể đọc đúng lúc này, và đọc
		// một tệp JSON mới viết được một nửa thì nó coi như không có cổng.
		tmp := l.Path + ".tmp"
		if err = os.WriteFile(tmp, b, 0o644); err == nil {
			err = os.Rename(tmp, l.Path)
		}
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.adminAuth.Store().Audit("leak_ack", "", "xác nhận đã xem kết quả rò rỉ đỏ", clientIP(r),
		map[string]any{"at": l.At, "note": l.Note})
	log.Printf("rò rỉ hằng đêm: đã xác nhận xem lại (%s)", l.Note)
	writeJSON(w, l)
}
