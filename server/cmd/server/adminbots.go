package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"bladeascend/server/internal/admin"
	"bladeascend/server/internal/room"
)

// Điều khiển bot — ADMIN_SPEC §4 "Điều khiển bot", GAME_SPEC §9.5.
//
// Mọi thao tác ghi `admin_audit` (kèm lý do nếu có). Chế độ toàn server sống trong bộ nhớ:
// khởi động lại thì về `bots.mode_default` của config — cố ý, để một lần "Tắt" quên bật lại
// không theo server qua các lần cập nhật.

type botRoomRow struct {
	ID string `json:"id"`
	room.BotStatus
	OK bool `json:"ok"` // false = phòng không trả lời kịp (quá tải), số liệu trống
}

func (s *server) handleAdminBots(w http.ResponseWriter, _ *http.Request) {
	rows := []botRoomRow{}
	for _, rm := range s.mgr.AllRooms() {
		st, ok := rm.BotStatusNow()
		rows = append(rows, botRoomRow{ID: rm.ID, BotStatus: st, OK: ok})
	}
	adminJSON(w, http.StatusOK, map[string]any{"mode": s.mgr.BotMode(), "rooms": rows})
}

// handleAdminBotsSet: {"room": "" | "r3", "mode": "auto|on|off|" , "reason": "..."}.
// room rỗng = đổi chế độ toàn server; mode rỗng chỉ hợp lệ cho phòng (= theo toàn server).
func (s *server) handleAdminBotsSet(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Room   string `json:"room"`
		Mode   string `json:"mode"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&b); err != nil {
		adminJSON(w, http.StatusBadRequest, map[string]string{"error": "nội dung không hợp lệ"})
		return
	}
	b.Reason = strings.TrimSpace(b.Reason)
	before, err := s.applyBotMode(b.Room, b.Mode)
	if err != nil {
		adminJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	target := b.Room
	if target == "" {
		target = "server"
	}
	s.adminAuth.Store().Audit("bot_mode", target, b.Reason, admin.ClientIP(r),
		map[string]any{"from": before, "to": b.Mode})
	adminJSON(w, http.StatusOK, map[string]int{"ok": 1})
}

// applyBotMode đổi chế độ, trả chế độ cũ (để ghi kiểm toán "từ → sang").
func (s *server) applyBotMode(roomID, mode string) (string, error) {
	if roomID == "" {
		before := s.mgr.BotMode()
		return before, s.mgr.SetBotMode(mode)
	}
	rm := s.mgr.RoomByID(roomID)
	if rm == nil {
		return "", errNoRoom
	}
	st, _ := rm.BotStatusNow()
	return st.Override, rm.SetBotOverride(mode)
}

func (s *server) handleAdminBotsWithdraw(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Room   string `json:"room"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&b); err != nil {
		adminJSON(w, http.StatusBadRequest, map[string]string{"error": "nội dung không hợp lệ"})
		return
	}
	rm := s.mgr.RoomByID(b.Room)
	if rm == nil {
		adminJSON(w, http.StatusNotFound, map[string]string{"error": errNoRoom.Error()})
		return
	}
	n := rm.WithdrawBots()
	s.adminAuth.Store().Audit("bot_withdraw", b.Room, strings.TrimSpace(b.Reason), admin.ClientIP(r),
		map[string]any{"marked": n})
	adminJSON(w, http.StatusOK, map[string]int{"ok": 1, "marked": n})
}

var errNoRoom = errors.New("không có phòng này (có thể vừa đóng)")
