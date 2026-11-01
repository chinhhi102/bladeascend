package store

import (
	"sync"
	"time"
)

// Phân bố ngôn ngữ người chơi toàn server 7 ngày — phần 25% của tỉ lệ tên bot (spec v6.34 §9.4).

const (
	langWindow = 7 * 24 * time.Hour
	langTTL    = 10 * time.Minute
)

type langCache struct {
	mu     sync.Mutex
	at     time.Time
	shares map[string]float64
}

// LangShares trả mã ngôn ngữ → tỉ lệ số NGƯỜI (guest_id khác nhau) chơi trong 7 ngày.
// Cache 10 phút: vòng tick gọi hàm này thì gần như luôn chỉ đọc map trong bộ nhớ.
// Chưa có dữ liệu (DB cũ, mạng chưa ghi `lang`) → map rỗng.
func (s *Store) LangShares(now time.Time) map[string]float64 {
	c := &s.langs
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.shares != nil && now.Sub(c.at) < langTTL {
		return c.shares
	}
	c.shares, c.at = s.queryLangShares(now.Add(-langWindow).Unix()), now
	return c.shares
}

func (s *Store) queryLangShares(since int64) map[string]float64 {
	out := map[string]float64{}
	rows, err := s.db.Query(`SELECT lang, COUNT(DISTINCT guest_id) FROM lives
		WHERE ended_at >= ? AND lang != '' GROUP BY lang`, since)
	if err != nil {
		return out
	}
	defer rows.Close()
	total := 0.0
	for rows.Next() {
		var lang string
		var n int
		if rows.Scan(&lang, &n) == nil {
			out[lang] = float64(n)
			total += float64(n)
		}
	}
	for k := range out {
		out[k] /= total
	}
	return out
}
