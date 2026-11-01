package store

// Số liệu theo nhân vật từ trận thật — spec v6.25 §6.3: "tỉ lệ chọn, Uy Danh/phút, thời gian
// sống, số lần dùng kỹ năng theo nhân vật (≥ 2 ngày)". Nhân vật có Uy Danh/phút lệch > 30%
// so với trung vị → chỉnh hồi chiêu ±20%.

// CharStat: một dòng của bảng theo nhân vật.
type CharStat struct {
	Character   string  `json:"character"`
	Lives       int     `json:"lives"`
	PickShare   float64 `json:"pick_share"`    // phần trăm số mạng
	ScorePerMin float64 `json:"score_per_min"` // Uy Danh / phút sống
	AvgSurvive  float64 `json:"avg_survive"`   // giây
	CastsPerMin float64 `json:"casts_per_min"` // lần dùng kỹ năng / phút sống
}

// CharacterStats gom các mạng kết thúc từ `since` (unix) trở đi, bỏ mạng đã ẩn và mạng
// dưới 30 giây (vào rồi thoát ngay, không nói gì về nhân vật).
func (s *Store) CharacterStats(since int64) ([]CharStat, error) {
	rows, err := s.db.Query(`SELECT character, COUNT(*), SUM(score), SUM(survive_secs), SUM(skill_casts)
		FROM lives WHERE ended_at >= ? AND survive_secs >= 30
		AND NOT EXISTS (SELECT 1 FROM hidden_lives h WHERE h.guest_id = lives.guest_id AND h.ended_at = lives.ended_at)
		GROUP BY character ORDER BY COUNT(*) DESC`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CharStat
	total := 0
	for rows.Next() {
		var c CharStat
		var score, secs float64
		var casts int
		if err := rows.Scan(&c.Character, &c.Lives, &score, &secs, &casts); err != nil {
			return nil, err
		}
		if secs > 0 {
			c.ScorePerMin = score / (secs / 60)
			c.CastsPerMin = float64(casts) / (secs / 60)
			c.AvgSurvive = secs / float64(c.Lives)
		}
		total += c.Lives
		out = append(out, c)
	}
	for i := range out {
		out[i].PickShare = 100 * float64(out[i].Lives) / float64(max(total, 1))
	}
	return out, rows.Err()
}
