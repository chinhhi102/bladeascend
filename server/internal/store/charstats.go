package store

// Số liệu theo nhân vật từ trận thật — spec v6.30 §6.3: cân bằng theo người chơi thật (≥ 2 ngày):
// tỉ lệ chọn, Uy Danh/phút, thời gian sống, cảnh giới cao nhất mỗi mạng, số lần dùng kỹ năng.

// CharStat: một dòng của bảng theo nhân vật.
type CharStat struct {
	Character    string  `json:"character"`
	Lives        int     `json:"lives"`
	PickShare    float64 `json:"pick_share"`     // phần trăm số mạng
	ScorePerMin  float64 `json:"score_per_min"`  // Uy Danh / phút sống
	AvgSurvive   float64 `json:"avg_survive"`    // giây
	AvgPeakRealm float64 `json:"avg_peak_realm"` // cảnh giới cao nhất TB mỗi mạng (0 = Luyện Khí)
	CastsPerMin  float64 `json:"casts_per_min"`  // lần dùng kỹ năng / phút sống
	Casts        int     `json:"casts"`          // tổng lần dùng kỹ năng
}

type charAcc struct {
	CharStat
	score, secs, realms float64
}

// CharacterStats gom các mạng kết thúc từ `since` (unix) trở đi, bỏ mạng đã ẩn và mạng
// dưới 30 giây (vào rồi thoát ngay, không nói gì về nhân vật). `realmOf` đổi cấp → cảnh giới:
// trung bình cảnh giới phải tính TỪNG MẠNG, cảnh giới của cấp trung bình là con số khác.
func (s *Store) CharacterStats(since int64, realmOf func(level int) int) ([]CharStat, error) {
	rows, err := s.db.Query(`SELECT character, score, survive_secs, skill_casts, peak_level
		FROM lives WHERE ended_at >= ? AND survive_secs >= 30
		AND NOT EXISTS (SELECT 1 FROM hidden_lives h WHERE h.guest_id = lives.guest_id AND h.ended_at = lives.ended_at)`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	acc := map[string]*charAcc{}
	var order []string
	total := 0
	for rows.Next() {
		var ch string
		var score, secs float64
		var casts, peak int
		if err := rows.Scan(&ch, &score, &secs, &casts, &peak); err != nil {
			return nil, err
		}
		a := acc[ch]
		if a == nil {
			a = &charAcc{CharStat: CharStat{Character: ch}}
			acc[ch] = a
			order = append(order, ch)
		}
		a.Lives++
		a.Casts += casts
		a.score += score
		a.secs += secs
		a.realms += float64(realmOf(peak))
		total++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return finishCharStats(acc, order, total), nil
}

func finishCharStats(acc map[string]*charAcc, order []string, total int) []CharStat {
	out := make([]CharStat, 0, len(order))
	for _, ch := range order {
		a := acc[ch]
		c := a.CharStat
		if a.secs > 0 {
			c.ScorePerMin = a.score / (a.secs / 60)
			c.CastsPerMin = float64(a.Casts) / (a.secs / 60)
		}
		c.AvgSurvive = a.secs / float64(c.Lives)
		c.AvgPeakRealm = a.realms / float64(c.Lives)
		c.PickShare = 100 * float64(c.Lives) / float64(max(total, 1))
		out = append(out, c)
	}
	// Nhiều mạng nhất lên đầu; bằng nhau thì theo tên cho ổn định.
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && (out[j].Lives > out[j-1].Lives ||
			out[j].Lives == out[j-1].Lives && out[j].Character < out[j-1].Character); j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}
