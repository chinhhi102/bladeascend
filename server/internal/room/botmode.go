package room

import (
	"fmt"
	"time"

	"bladeascend/server/internal/bot"
	"bladeascend/server/internal/game"
)

// Bật / tắt bot — spec v6.22 §9.5, ADMIN_SPEC §4.
//
// Ba chế độ, đặt cho TOÀN SERVER và ghi đè từng phòng. Bot bị tắt (tay hay tự động) thì
// không sinh mới, bot chết không hồi sinh: bot vơi dần tự nhiên, không biến mất giữa trận.

const (
	BotAuto   = "auto"
	BotOn     = "on"
	BotOff    = "off"
	BotFollow = "" // phòng: theo chế độ toàn server
)

// ValidBotMode: chế độ hợp lệ cho toàn server (`room` = false) hay cho phòng (`room` = true,
// thêm "" = theo toàn server).
func ValidBotMode(mode string, room bool) bool {
	switch mode {
	case BotAuto, BotOn, BotOff:
		return true
	case BotFollow:
		return room
	default:
		return false
	}
}

// BotMode / SetBotMode: chế độ toàn server.
func (m *Manager) BotMode() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.botMode == "" {
		if ValidBotMode(m.cfg.Bots.ModeDefault, false) {
			return m.cfg.Bots.ModeDefault
		}
		return BotAuto
	}
	return m.botMode
}

func (m *Manager) SetBotMode(mode string) error {
	if !ValidBotMode(mode, false) {
		return fmt.Errorf("chế độ bot không hợp lệ: %q", mode)
	}
	m.mu.Lock()
	m.botMode = mode
	m.mu.Unlock()
	return nil
}

// RoomByID: phòng đang mở có mã `id`, nil nếu không có.
func (m *Manager) RoomByID(id string) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range m.rooms {
		if r.ID == id && !r.closed {
			return r
		}
	}
	return nil
}

// SetBotOverride đặt chế độ riêng của phòng ("" = theo toàn server). Chạy trong goroutine
// của phòng, nên không cần khoá.
func (r *Room) SetBotOverride(mode string) error {
	if !ValidBotMode(mode, true) {
		return fmt.Errorf("chế độ bot không hợp lệ: %q", mode)
	}
	r.Do(func() { r.botOverride = mode })
	return nil
}

// WithdrawBots: "Rút bot ngay" — đánh dấu mọi bot đang có. Bot nào không giao tranh trong
// `withdraw_idle_secs` thì rời ngay (kèm khói), bot đang đánh rút khi thoát giao tranh.
// Trả số bot được đánh dấu.
func (r *Room) WithdrawBots() int {
	n, _ := ask(r, func() int {
		if r.withdraw == nil {
			r.withdraw = map[uint32]bool{}
		}
		n := 0
		for _, z := range r.zones {
			for id := range z.brains {
				r.withdraw[id] = true
				n++
			}
		}
		return n
	})
	return n
}

// ask chạy fn trong goroutine của phòng và chờ kết quả — tối đa 2 s, để một phòng kẹt hay
// đã đóng không treo luôn request của trang quản trị.
func ask[T any](r *Room, fn func() T) (T, bool) {
	done := make(chan T, 1)
	r.Do(func() { done <- fn() })
	select {
	case v := <-done:
		return v, true
	case <-time.After(2 * time.Second):
		var zero T
		return zero, false
	}
}

// botModeNow: chế độ đang áp dụng cho phòng (đã gộp ghi đè).
func (r *Room) botModeNow() string {
	if r.botOverride != BotFollow {
		return r.botOverride
	}
	if r.mgr != nil {
		return r.mgr.BotMode()
	}
	if r.cfg.Bots.ModeDefault != "" {
		return r.cfg.Bots.ModeDefault
	}
	return BotAuto
}

// updateBotSwitch chốt `botsOn` cho tick này — ở pha ĐƠN LUỒNG, trước khi các khu chạy song
// song (`onBotDeath` đọc nó từ goroutine của khu). Trễ 40/35 của chế độ tự động nằm ở đây.
func (r *Room) updateBotSwitch() {
	r.applyBotSwitch(r.Humans())
}

// applyBotSwitch: phần luật của `updateBotSwitch`, nhận số người để test được trễ 40/35 mà
// không phải dựng 41 kết nối thật.
func (r *Room) applyBotSwitch(humans int) {
	switch {
	case humans > r.cfg.Bots.AutoOffAboveHumans && r.cfg.Bots.AutoOffAboveHumans > 0:
		r.botAutoOff = true
	case humans < r.cfg.Bots.AutoOnBelowHumans:
		r.botAutoOff = false
	default:
	}
	switch r.botModeNow() {
	case BotOff:
		r.botsOn = false
	case BotOn:
		r.botsOn = true
	default:
		r.botsOn = !r.botAutoOff
	}
}

// withdrawIdle cho bot đã bị "Rút bot ngay" rời phòng khi chúng hết giao tranh.
//
// "Giao tranh" = va kiếm / chém / bị chém trong `withdraw_idle_secs`, HOẶC não bot đang
// săn hay đang chạy trốn (bị đuổi). Rút một con đang đánh dở thì đối thủ của nó thấy kẻ thù
// bốc hơi giữa nhát chém — đúng thứ spec bảo tránh.
func (r *Room) withdrawIdle() {
	if len(r.withdraw) == 0 {
		return
	}
	idle := r.cfg.Bots.WithdrawIdleSecs
	for _, z := range r.zones {
		for id, b := range z.brains {
			if !r.withdraw[id] {
				continue
			}
			p := z.G.Players[id]
			if p != nil && !p.Dead && inCombat(z.G, p, b, idle) {
				continue
			}
			if p != nil {
				z.G.EmitBotLeave(p)
				z.departed = append(z.departed, p.Name)
			}
			z.G.RemovePlayer(id)
			delete(z.brains, id)
			delete(r.withdraw, id)
		}
	}
	// Bot đã chết/rời bằng đường khác thì bỏ khỏi danh sách, không giữ rác.
	for id := range r.withdraw {
		if r.ZoneOf(id) == nil {
			delete(r.withdraw, id)
		}
	}
}

func inCombat(g *game.Game, p *game.Player, b *bot.Brain, idle float64) bool {
	if p.CombatAt > 0 && g.Time-p.CombatAt < idle {
		return true
	}
	return b.State == bot.StateHunt || b.State == bot.StateFlee
}

// BotStatus: tình trạng bot của phòng cho trang quản trị.
type BotStatus struct {
	Override  string `json:"override"`  // "" = theo toàn server
	Effective string `json:"effective"` // auto / on / off
	AutoOff   bool   `json:"auto_off"`  // chế độ tự động đang tắt vì đông người
	Spawning  bool   `json:"spawning"`  // tick này có được sinh bot không
	Humans    int    `json:"humans"`
	Bots      int    `json:"bots"`
	MaxHumans int    `json:"max_humans"`
	Withdraw  int    `json:"withdrawing"` // số bot đang chờ rút
}

// BotStatusNow đọc tình trạng bot — chạy trong goroutine của phòng để không đua dữ liệu.
func (r *Room) BotStatusNow() (BotStatus, bool) {
	return ask(r, func() BotStatus {
		return BotStatus{Override: r.botOverride, Effective: r.botModeNow(), AutoOff: r.botAutoOff,
			Spawning: r.botsOn, Humans: r.Humans(), Bots: r.Bots(), MaxHumans: r.cfg.Room.MaxHumans,
			Withdraw: len(r.withdraw)}
	})
}
