package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"bladeascend/server/internal/admin"
	"bladeascend/server/internal/room"
)

// Cấu hình phòng — GAME_SPEC v6.28 §9.5, ADMIN_SPEC v1.4 §4: trần người thật/phòng, sàn nhân
// vật/khu, ngưỡng tự tắt bot. Áp dụng ngay, lưu SQLite (sống qua deploy), ghi `admin_audit`.

const (
	settingRoom    = "room_settings"
	settingBotMode = "bot_mode"
)

// restoreRoomSettings nạp lại giá trị admin đã đặt từ lần chạy trước. Dữ liệu hỏng thì bỏ
// qua (giữ mặc định môi trường) và ghi log — không chặn server khởi động vì một dòng cấu hình.
func restoreRoomSettings(mgr *room.Manager, st *admin.Store) {
	if raw, ok := st.Setting(settingRoom); ok {
		var rs room.RoomSettings
		if err := json.Unmarshal([]byte(raw), &rs); err != nil {
			log.Printf("cấu hình phòng đã lưu hỏng, dùng mặc định: %v", err)
		} else if err := mgr.SetSettings(rs); err != nil {
			log.Printf("cấu hình phòng đã lưu không hợp lệ, dùng mặc định: %v", err)
		}
	}
	if mode, ok := st.Setting(settingBotMode); ok {
		if err := mgr.SetBotMode(mode); err != nil {
			log.Printf("chế độ bot đã lưu không hợp lệ: %v", err)
		}
	}
	s := mgr.Settings()
	log.Printf("cấu hình phòng: trần %d, sàn %d/khu, tắt bot ở %d, bot %s",
		s.MaxHumans, s.ZoneMinChars, s.BotAutoOffAt, mgr.BotMode())
}

type roomSettingsView struct {
	Env        string            `json:"env"`
	Settings   room.RoomSettings `json:"settings"`
	Defaults   room.RoomSettings `json:"defaults"`
	Loadtested int               `json:"loadtested_humans"`
	// Warn: trần đang đặt cao hơn mức đã qua loadtest của máy này → trang admin tô vàng.
	Warn bool `json:"warn"`
}

func (s *server) roomSettingsView() roomSettingsView {
	cur := s.mgr.Settings()
	lt := s.envDefaults.LoadtestedHumans
	return roomSettingsView{Env: s.env, Settings: cur, Defaults: s.mgr.DefaultSettings(),
		Loadtested: lt, Warn: lt > 0 && cur.MaxHumans > lt}
}

func (s *server) handleAdminRoomSettings(w http.ResponseWriter, _ *http.Request) {
	adminJSON(w, http.StatusOK, s.roomSettingsView())
}

// handleAdminRoomSettingsSet: {"max_humans", "zone_min_chars", "bot_auto_off_at", "reason"}.
// Vượt mức loadtest vẫn cho lưu (spec: cảnh báo vàng, không chặn).
func (s *server) handleAdminRoomSettingsSet(w http.ResponseWriter, r *http.Request) {
	var b struct {
		room.RoomSettings
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&b); err != nil {
		adminJSON(w, http.StatusBadRequest, map[string]string{"error": "nội dung không hợp lệ"})
		return
	}
	before := s.mgr.Settings()
	if err := s.mgr.SetSettings(b.RoomSettings); err != nil {
		adminJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	js, _ := json.Marshal(b.RoomSettings)
	if err := s.adminAuth.Store().SetSetting(settingRoom, string(js)); err != nil {
		// Đã áp dụng trong bộ nhớ nhưng chưa lưu được: báo rõ để admin biết deploy sau sẽ mất.
		adminJSON(w, http.StatusInternalServerError, map[string]string{"error": "đã áp dụng nhưng chưa lưu được: " + err.Error()})
		return
	}
	s.adminAuth.Store().Audit("room_settings", "server", strings.TrimSpace(b.Reason), admin.ClientIP(r),
		map[string]any{"from": before, "to": b.RoomSettings})
	adminJSON(w, http.StatusOK, s.roomSettingsView())
}
