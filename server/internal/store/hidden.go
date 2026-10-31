package store

import (
	"errors"
	"strings"
	"time"
)

// Ẩn mạng chơi khỏi bảng xếp hạng — ADMIN_SPEC §6.2.
//
// Cả tệp này chỉ làm được hai việc: thêm một dòng vào `hidden_lives`, và bỏ dòng đó đi.
// KHÔNG có hàm nào xoá được `lives`. Đó là chủ ý: một thao tác quản trị chạy lúc 2 giờ sáng
// trên một linh cảm thì phải sửa lại được lúc 9 giờ sáng khi đã tỉnh.

// ErrHideReason: ẩn mà không nêu lý do thì ba tháng sau không ai biết vì sao nó bị ẩn —
// kể cả người đã ẩn nó.
var ErrHideReason = errors.New("phải nêu lý do ẩn")

// HideReasonMax là trần độ dài lý do.
const HideReasonMax = 300

// Hidden là một mạng chơi đang bị ẩn, kèm số liệu của nó để trang admin hiện lại được.
type Hidden struct {
	GuestID string  `json:"guest_id"`
	EndedAt int64   `json:"ended_at"`
	Reason  string  `json:"reason"`
	At      int64   `json:"at"`
	ByIP    string  `json:"by_ip"`
	Name    string  `json:"name"`
	Score   float64 `json:"score"`
	Level   int     `json:"peak_level"`
	Survive float64 `json:"survive_secs"`
}

// Hide ẩn một mạng chơi. Gọi lại với cùng (guest_id, ended_at) thì ghi đè lý do.
func (s *Store) Hide(guestID string, endedAt int64, reason, byIP string, now time.Time) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return ErrHideReason
	}
	if len(reason) > HideReasonMax {
		reason = reason[:HideReasonMax]
	}
	_, err := s.db.Exec(`INSERT INTO hidden_lives (guest_id, ended_at, reason, at, by_ip)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(guest_id, ended_at) DO UPDATE SET reason=excluded.reason, at=excluded.at, by_ip=excluded.by_ip`,
		guestID, endedAt, reason, now.Unix(), byIP)
	if err == nil {
		s.dropCache()
	}
	return err
}

// Unhide bỏ ẩn.
func (s *Store) Unhide(guestID string, endedAt int64) error {
	_, err := s.db.Exec(`DELETE FROM hidden_lives WHERE guest_id = ? AND ended_at = ?`, guestID, endedAt)
	if err == nil {
		s.dropCache()
	}
	return err
}

// ListHidden trả về mọi mạng chơi đang bị ẩn, mới ẩn trước.
//
// LEFT JOIN ở đây thì đúng (ngược với `query`): nếu dòng `lives` đã mất vì lý do nào đó thì
// vẫn phải hiện dấu vết việc ẩn, chứ không im lặng bỏ qua.
func (s *Store) ListHidden() ([]Hidden, error) {
	rs, err := s.db.Query(`SELECT h.guest_id, h.ended_at, h.reason, h.at, h.by_ip,
			COALESCE(l.name,''), COALESCE(l.score,0), COALESCE(l.peak_level,0), COALESCE(l.survive_secs,0)
		FROM hidden_lives h
		LEFT JOIN lives l ON l.guest_id = h.guest_id AND l.ended_at = h.ended_at
		ORDER BY h.at DESC LIMIT 2000`)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	out := []Hidden{}
	for rs.Next() {
		var h Hidden
		if err := rs.Scan(&h.GuestID, &h.EndedAt, &h.Reason, &h.At, &h.ByIP,
			&h.Name, &h.Score, &h.Level, &h.Survive); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rs.Err()
}

// dropCache xoá cache bảng xếp hạng.
//
// Thiếu chỗ này thì ẩn xong bảng vẫn hiện tên đó tới hết `cacheTTL` — và người bấm nút sẽ
// bấm lại vài lần nữa vì tưởng nút hỏng.
func (s *Store) dropCache() {
	s.mu.Lock()
	s.cache = map[string]cached{}
	s.mu.Unlock()
}

// LifeRow là một dòng `lives` thô, cho trang admin rà soát (§6.2).
type LifeRow struct {
	GuestID string  `json:"guest_id"`
	Name    string  `json:"name"`
	Char    string  `json:"character"`
	Score   float64 `json:"score"`
	Hits    int     `json:"hits"`
	Level   int     `json:"peak_level"`
	Kills   int     `json:"kills"`
	Survive float64 `json:"survive_secs"`
	Map     string  `json:"map"`
	EndedAt int64   `json:"ended_at"`
	Hidden  bool    `json:"hidden"`
	Reason  string  `json:"hide_reason"`
}

// RecentLives trả về các mạng chơi gần nhất, KỂ CẢ đang bị ẩn (có cờ `hidden`).
//
// Trang admin cần thấy cả hai loại: muốn bỏ ẩn một mạng thì phải nhìn thấy nó trước.
func (s *Store) RecentLives(limit int) ([]LifeRow, error) {
	if limit <= 0 || limit > 1000 {
		limit = 300
	}
	rs, err := s.db.Query(`SELECT l.guest_id, l.name, l.character, l.score, l.hits, l.peak_level,
			l.kills, l.survive_secs, l.map, l.ended_at, h.reason IS NOT NULL, COALESCE(h.reason,'')
		FROM lives l
		LEFT JOIN hidden_lives h ON h.guest_id = l.guest_id AND h.ended_at = l.ended_at
		ORDER BY l.ended_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	out := []LifeRow{}
	for rs.Next() {
		var r LifeRow
		if err := rs.Scan(&r.GuestID, &r.Name, &r.Char, &r.Score, &r.Hits, &r.Level,
			&r.Kills, &r.Survive, &r.Map, &r.EndedAt, &r.Hidden, &r.Reason); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rs.Err()
}
