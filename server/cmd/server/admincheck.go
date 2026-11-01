package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Báo cáo `mapgen --check` và lỗi người chơi cho trang quản trị.
//
// Luồng "xem bản đồ trực tiếp" 5 lần/giây (`/api/admin/live`, ADMIN_SPEC v1.1 §5.1) đã BỎ
// theo ADMIN_SPEC v1.2 §5: trang Bản đồ giờ vẽ bố cục tĩnh bằng chính bộ vẽ của game, không
// chấm vị trí người chơi — không tốn một chu kỳ nào của goroutine phòng.

// handleAdminCheck trả báo cáo `mapgen --check` của bộ bản đồ ĐANG CHẠY — §5.2.
//
// Đọc từ đĩa mỗi lần gọi, không cache: tệp này thay đổi khi ai đó sinh lại bản đồ, và một
// báo cáo cũ nói về một bố cục không còn tồn tại thì tệ hơn hẳn việc không có báo cáo.
func (s *server) handleAdminCheck(w http.ResponseWriter, _ *http.Request) {
	b, err := os.ReadFile(filepath.Join(s.mapsDir, "check.json"))
	if err != nil {
		adminJSON(w, http.StatusNotFound, map[string]string{
			"error": "chưa có báo cáo. Chạy: cd server && go run ./cmd/mapgen --check",
		})
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(b)
}

// handleAdminErrors trả lỗi của NGƯỜI CHƠI THẬT, gom nhóm (ADMIN_SPEC §3).
//
// Đây là câu trả lời cho "người chơi iPhone báo luôn bị lỗi mà không biết lỗi gì": console
// trên máy họ thì mình không với tới, và bảo họ chụp màn hình console Safari iOS là chuyện
// không xảy ra.
func (s *server) handleAdminErrors(w http.ResponseWriter, r *http.Request) {
	if s.telemetry == nil {
		adminJSON(w, http.StatusOK, map[string]any{
			"off": true, "groups": []any{},
			"note": "server đang chạy với -telemetry=false nên không thu lỗi người chơi",
		})
		return
	}
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days <= 0 || days > 30 {
		days = 7
	}
	groups, err := s.telemetry.RecentErrors(time.Duration(days)*24*time.Hour, time.Now())
	if err != nil {
		adminJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	adminJSON(w, http.StatusOK, map[string]any{"days": days, "groups": groups})
}
