package store

import (
	"sort"
	"sync"
	"time"
)

// Mốc trung vị Uy Danh theo cảnh giới — ADMIN_SPEC §6.1 "Tăng điểm bất thường".
//
// Chủ dự án chốt 31/10 (QUESTIONS Q29): lấy trung vị TRƯỢT 24 GIỜ từ dữ liệu người chơi
// thật, rơi về bảng chốt sẵn khi chưa đủ 50 mạng chơi trong 24 giờ.
//
// Vì sao không lấy trung vị của người đang trong phòng: phòng ít người thì trung vị nhiễu
// nặng, và nếu cả phòng đều gian thì không ai bị gắn cờ — mốc so sánh trôi theo đúng thứ nó
// phải phát hiện.
//
// Vì sao không chốt cứng trong config: mỗi lần đổi cân bằng là mốc sai, và không ai nhớ
// chạy lại `cmd/sim` để cập nhật. Dữ liệu thật thì tự theo.

// medianMinLives: dưới ngần này mạng chơi trong 24 giờ thì số đo quá ít để tin.
//
// 50 là mức chủ dự án chốt. Trung vị của 5 mạng chơi thì một người chơi giỏi bất thường đủ
// kéo mốc lên gấp đôi, và cờ "tăng điểm bất thường" sẽ im lặng đúng lúc cần nói.
const medianMinLives = 50

// medianTTL: tính lại mỗi ngần này. Truy vấn quét 24 giờ dữ liệu nên không gọi mỗi tick.
const medianTTL = 5 * time.Minute

// fallbackMedian là bảng chốt sẵn, dùng khi chưa đủ dữ liệu thật.
//
// Lấy từ `cmd/sim --bots 40 --minutes 20` (bot chơi theo đúng luật đang chạy): Uy Danh mỗi
// phút ở từng cảnh giới. Đổi cân bằng thì chạy lại sim và sửa bảng này — nhưng chỉ cần khi
// server mới mở hoặc vắng người; bình thường dữ liệu thật đè lên.
var fallbackMedian = map[int]float64{
	0: 40, 1: 60, 2: 85, 3: 110, 4: 140, 5: 170, 6: 200,
	7: 240, 8: 290, 9: 340, 10: 390, 11: 440, 12: 490, 13: 540,
}

// medianCache nằm TRONG `Store`, không phải biến toàn gói.
//
// Bản đầu tôi để nó là biến toàn gói và nó sai theo hai đường: hai `Store` (ví dụ trong
// test) dùng chung một cache, nên kết quả phụ thuộc thứ tự chạy; và một bộ nhớ đệm toàn
// cục thì không có cách nào vứt đi khi đóng DB. Cache là trạng thái CỦA MỘT KHO DỮ LIỆU.
type medianCache struct {
	mu   sync.Mutex
	at   time.Time
	byR  map[int]float64
	real bool // số đo đến từ dữ liệu thật hay từ bảng chốt sẵn
	n    int
}

// ScoreMedianPerMin trả về Uy Danh mỗi phút TRUNG VỊ của cảnh giới `realm`, và `true` nếu
// con số đến từ dữ liệu người chơi thật.
//
// `realmOf` đổi `peak_level` sang chỉ số cảnh giới — truyền vào vì `store` không biết luật
// cảnh giới (nó nằm ở `internal/game`), và gói này không được phụ thuộc gói đó.
func (s *Store) ScoreMedianPerMin(realm int, realmOf func(level int) int, now time.Time) (float64, bool) {
	s.medians.mu.Lock()
	fresh := s.medians.byR != nil && now.Sub(s.medians.at) < medianTTL
	s.medians.mu.Unlock()
	if !fresh {
		s.refreshMedians(realmOf, now)
	}
	s.medians.mu.Lock()
	defer s.medians.mu.Unlock()
	if v, ok := s.medians.byR[realm]; ok && s.medians.real {
		return v, true
	}
	if v, ok := fallbackMedian[realm]; ok {
		return v, false
	}
	return 0, false
}

// refreshMedians tính lại bảng từ 24 giờ dữ liệu gần nhất.
func (s *Store) refreshMedians(realmOf func(level int) int, now time.Time) {
	since := now.Add(-24 * time.Hour).Unix()
	// Bỏ mạng chơi quá ngắn: 10 giây đầu luôn cho Uy Danh/phút cao vô lý vì mẫu số bé.
	// Và bỏ mạng ĐÃ BỊ ẨN — chúng chính là những mạng nghi gian, để lại thì mốc so sánh
	// được kéo lên bởi đúng thứ nó phải phát hiện.
	rows, err := s.db.Query(`SELECT peak_level, score, survive_secs FROM lives
		WHERE ended_at >= ? AND survive_secs >= 30 AND score > 0
		AND NOT EXISTS (SELECT 1 FROM hidden_lives h
			WHERE h.guest_id = lives.guest_id AND h.ended_at = lives.ended_at)`, since)
	if err != nil {
		return
	}
	defer rows.Close()

	perRealm := map[int][]float64{}
	n := 0
	for rows.Next() {
		var level int
		var score, secs float64
		if err := rows.Scan(&level, &score, &secs); err != nil {
			return
		}
		if secs <= 0 {
			continue
		}
		r := realmOf(level)
		perRealm[r] = append(perRealm[r], score/secs*60)
		n++
	}
	if rows.Err() != nil {
		return
	}

	out := map[int]float64{}
	for r, vals := range perRealm {
		// Mỗi cảnh giới cũng phải có đủ mẫu riêng, không chỉ tổng đủ: 200 mạng chơi mà
		// chỉ 2 mạng ở Đạo Tổ thì mốc của Đạo Tổ vẫn là trung vị của 2 con số.
		if len(vals) < 5 {
			continue
		}
		sort.Float64s(vals)
		out[r] = vals[len(vals)/2]
	}

	s.medians.mu.Lock()
	s.medians.at, s.medians.byR, s.medians.n = now, out, n
	s.medians.real = n >= medianMinLives
	s.medians.mu.Unlock()
}

// MedianSource mô tả mốc đang dùng, cho trang quản trị hiện ra.
func (s *Store) MedianSource() (real bool, lives int) {
	s.medians.mu.Lock()
	defer s.medians.mu.Unlock()
	return s.medians.real, s.medians.n
}
