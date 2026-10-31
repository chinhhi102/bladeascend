package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"kiemxoay/server/internal/admin"
	"kiemxoay/server/internal/store"
)

// Ẩn mạng chơi khỏi bảng xếp hạng — ADMIN_SPEC §6.2.
//
// Toàn bộ đường này chỉ ẩn và bỏ ẩn. Không có handler nào xoá `lives`, và sẽ không có:
// bảng xếp hạng là thành tích người ta bỏ hàng giờ ra đạt được, kể cả khi mình nghi nó gian.

func (s *server) handleAdminLives(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	rows, err := s.store.RecentLives(limit)
	if err != nil {
		adminJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	adminJSON(w, http.StatusOK, rows)
}

func (s *server) handleAdminHidden(w http.ResponseWriter, _ *http.Request) {
	rows, err := s.store.ListHidden()
	if err != nil {
		adminJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	adminJSON(w, http.StatusOK, rows)
}

// handleAdminHide ẩn hoặc bỏ ẩn một mạng chơi.
//
// Một handler cho cả hai chiều, vì chúng phải ghi cùng một kiểu bản ghi kiểm toán — tách ra
// là mời gọi chuyện một chiều có ghi kiểm toán còn chiều kia quên.
func (s *server) handleAdminHide(w http.ResponseWriter, r *http.Request) {
	var b struct {
		GuestID string `json:"guest_id"`
		EndedAt int64  `json:"ended_at"`
		Reason  string `json:"reason"`
		Unhide  bool   `json:"unhide"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&b); err != nil {
		adminJSON(w, http.StatusBadRequest, map[string]string{"error": "nội dung không hợp lệ"})
		return
	}
	if b.GuestID == "" || b.EndedAt == 0 {
		adminJSON(w, http.StatusBadRequest, map[string]string{"error": "thiếu guest_id hoặc ended_at"})
		return
	}
	ip := admin.ClientIP(r)
	act, err := "leaderboard_hide", error(nil)
	if b.Unhide {
		act, err = "leaderboard_unhide", s.store.Unhide(b.GuestID, b.EndedAt)
	} else {
		err = s.store.Hide(b.GuestID, b.EndedAt, b.Reason, ip, time.Now())
	}
	if err != nil {
		code := http.StatusInternalServerError
		if err == store.ErrHideReason {
			code = http.StatusBadRequest
		}
		adminJSON(w, code, map[string]string{"error": err.Error()})
		return
	}
	// Ghi kiểm toán SAU khi đổi xong, và ghi cả lý do: §9 đòi mọi thao tác quản trị để lại
	// dấu, và một dấu không có lý do thì chỉ chứng minh được "có người đã làm gì đó".
	s.adminAuth.Store().Audit(act, b.GuestID, strconv.FormatInt(b.EndedAt, 10), ip,
		map[string]any{"reason": b.Reason})
	adminJSON(w, http.StatusOK, map[string]int{"ok": 1})
}
