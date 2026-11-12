package game

import "time"

// Đo thời gian TỪNG NHÓM VIỆC trong một tick (ADMIN_SPEC §12.1).
//
// Vì sao cần: con số tick tổng chỉ nói "chậm", không nói "chậm ở đâu". Ngày 12/11 Thiên Đình
// p95 17,3 ms mà loadtest 110 nhân vật hôm 02/11 chỉ 3,2–3,75 ms — không có bảng này thì chỉ
// còn cách đoán giữa "máy bận", "kỹ năng v6.48", "bot đã lên Đạo Tổ nên nhiều kiếm hơn".
//
// Giá phải trả: 5 lần đọc đồng hồ mỗi tick mỗi khu. `time.Now()` trên Linux amd64 là
// `clock_gettime` qua vDSO (~25 ns), tức ~0,000 1 ms mỗi tick — dưới trần 0,1 ms của §12 hai
// bậc độ lớn. Đo bằng bench chứ không tin lời: `BenchmarkTickProfileOverhead`.
type TickProfile struct {
	// Move: di chuyển người, quái, đàn em + quay vòng kiếm.
	Move float64
	// Skills: trạng thái kỹ năng và hồi chiêu (v6.48 §6.3).
	Skills float64
	// Collide: kiếm↔thân, kiếm↔kiếm, kiếm↔vật cản, gãy kiếm, nhặt đồ.
	Collide float64
	// World: sinh quái/vật phẩm, đài, hết hạn, tâm trạng, top 1, danh hiệu, lưới va chạm.
	World float64
}

// Add cộng dồn hai bảng — dùng khi gom nhiều tick lại.
func (p *TickProfile) Add(o TickProfile) {
	p.Move += o.Move
	p.Skills += o.Skills
	p.Collide += o.Collide
	p.World += o.World
}

// Profile là bảng thời gian của tick GẦN NHẤT, tính bằng mili-giây.
func (g *Game) Profile() TickProfile { return g.prof }

// Counters là các bộ đếm CỘNG DỒN từ lúc khu sinh ra: lượt dùng kỹ năng và lượt chết
// (§12.1 "lượt kỹ năng/phút, lượt chết/phút"). Cộng dồn chứ không phải "từ lần đọc trước"
// để nhiều người đọc cùng lúc không vét mất số của nhau — ai cần tốc độ thì tự trừ hai lần
// đọc liên tiếp.
func (g *Game) Counters() (casts, deaths int) { return g.nCasts, g.nDeaths }

// ms: số mili-giây kể từ `t`, và mốc mới cho đoạn kế tiếp.
func ms(t time.Time) (float64, time.Time) {
	now := time.Now()
	return float64(now.Sub(t).Nanoseconds()) / 1e6, now
}
