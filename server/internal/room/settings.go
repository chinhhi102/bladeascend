package room

import (
	"fmt"

	"bladeascend/server/internal/config"
)

// Cấu hình phòng chỉnh ở admin — spec v6.28 §9.5, ADMIN_SPEC v1.4 §4.
//
// Mặc định lấy theo môi trường (`config.ApplyEnv` chép `room_env` vào cfg.Room/cfg.Bots);
// admin ghi đè lúc chạy, `cmd/server` lưu SQLite để sống qua deploy.

// RoomSettings: ba thông số admin chỉnh được.
type RoomSettings struct {
	MaxHumans int `json:"max_humans"` // trần người thật/phòng, bot không tính
	// ZoneMinChars: sàn nhân vật (người + bot) TỪNG KHU (spec v6.35), khoá = id bản đồ khu.
	ZoneMinChars config.ZoneFloors `json:"zone_min_chars"`
	BotAutoOffAt int               `json:"bot_auto_off_at"` // tự tắt sinh bot khi người thật ≥ số này
}

// Giới hạn hợp lệ — chặn gõ nhầm (vd 500 người/phòng), không phải số cân bằng.
const (
	maxSettingHumans = 200
	maxSettingFloor  = 60
)

// Validate: số âm hay quá lớn bị từ chối, để một lần gõ nhầm không làm phòng nổ bot.
func (s RoomSettings) Validate() error {
	switch {
	case s.MaxHumans < 1 || s.MaxHumans > maxSettingHumans:
		return fmt.Errorf("trần người thật phải trong 1–%d", maxSettingHumans)
	case len(s.ZoneMinChars) == 0:
		return fmt.Errorf("thiếu sàn nhân vật mỗi khu")
	case s.BotAutoOffAt < 0 || s.BotAutoOffAt > maxSettingHumans:
		return fmt.Errorf("ngưỡng tự tắt bot phải trong 0–%d (0 = không tự tắt)", maxSettingHumans)
	default:
		return validFloors(s.ZoneMinChars)
	}
}

func validFloors(f config.ZoneFloors) error {
	for zone, n := range f {
		if zone == "" || n < 0 || n > maxSettingFloor {
			return fmt.Errorf("sàn nhân vật khu %q phải trong 0–%d", zone, maxSettingFloor)
		}
	}
	return nil
}

// DefaultSettings: mặc định của môi trường đang chạy (đã ApplyEnv vào cfg).
func (m *Manager) DefaultSettings() RoomSettings {
	return RoomSettings{MaxHumans: m.cfg.Room.MaxHumans, ZoneMinChars: m.cfg.Room.ZoneMinChars.Clone(),
		BotAutoOffAt: m.cfg.Bots.AutoOffAt}
}

// Settings: giá trị admin đã đặt, chưa đặt thì mặc định môi trường.
func (m *Manager) Settings() RoomSettings {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.settingsLocked()
}

func (m *Manager) settingsLocked() RoomSettings {
	if m.settings != nil {
		return *m.settings
	}
	return m.DefaultSettings()
}

// SetSettings áp dụng ngay: phòng đọc lại ở đầu tick kế tiếp. Hạ trần không đuổi ai —
// phòng đang vượt trần chỉ ngừng nhận người mới (matchmaking so `reserved` với trần).
func (m *Manager) SetSettings(s RoomSettings) error {
	if err := s.Validate(); err != nil {
		return err
	}
	s.ZoneMinChars = s.ZoneMinChars.Clone() // người gọi giữ map gốc; phòng đọc bản này từ goroutine khác
	m.mu.Lock()
	m.settings = &s
	m.mu.Unlock()
	return nil
}

// latchSettings chốt cấu hình cho tick này ở pha ĐƠN LUỒNG (như `botsOn`): `onBotDeath`
// đọc `r.set` từ goroutine của khu.
func (r *Room) latchSettings() {
	if r.mgr != nil {
		r.set = r.mgr.Settings()
		return
	}
	r.set = RoomSettings{MaxHumans: r.cfg.Room.MaxHumans, ZoneMinChars: r.cfg.Room.ZoneMinChars.Clone(),
		BotAutoOffAt: r.cfg.Bots.AutoOffAt}
}

// floorOf: sàn nhân vật của khu z theo cấu hình đã chốt cho tick này.
func (r *Room) floorOf(z *Zone) int { return r.set.ZoneMinChars.For(z.G.Map.ID) }
