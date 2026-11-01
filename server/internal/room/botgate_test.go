package room

import (
	"testing"
)

// Bot phải biết dùng cổng (spec v5 mục 8). Không có luật này thì bot nào rơi vào khu vắng
// sẽ quanh quẩn ở đó tới hết trận và cả một khu của thế giới thành vùng chết.
func TestBotsEventuallyChangeZone(t *testing.T) {
	m, cfg := zoneManager(t)
	cfg.Room.ZoneMinChars = 2
	joinN(m, 1)
	r := m.AllRooms()[0]
	settle(r, cfg)

	dt := 1 / float64(cfg.Net.TickHz)
	// Ghi lại bot nào đang ở khu nào, rồi chạy một lúc xem có con nào đổi khu không.
	where := map[uint32]int{}
	for _, z := range r.Zones() {
		for id := range z.brains {
			where[id] = z.Index
		}
	}
	if len(where) == 0 {
		// Bot sinh dần theo bots.spawn_every_secs; chạy thêm cho tới khi có.
		for i := 0; i < int(float64(cfg.Net.TickHz)*cfg.Bots.SpawnEverySecs*4) && len(where) == 0; i++ {
			r.Step(dt)
			for _, z := range r.Zones() {
				for id := range z.brains {
					where[id] = z.Index
				}
			}
		}
	}
	if len(where) == 0 {
		t.Skip("phòng chưa sinh bot nào")
	}

	moved := 0
	for i := 0; i < int(float64(cfg.Net.TickHz)*180); i++ {
		r.Step(dt)
		for _, z := range r.Zones() {
			for id := range z.brains {
				if old, ok := where[id]; ok && old != z.Index {
					moved++
					where[id] = z.Index
				}
			}
		}
		if moved > 0 {
			break
		}
	}
	if moved == 0 {
		t.Fatal("chạy 3 phút mà không bot nào qua cổng sang khu khác")
	}
}

// Spec v6.28 §9.3 (Q34): bot lấp THEO KHU — mỗi khu đủ `zone_min_chars` nhân vật, không
// còn trần tổng của cả phòng. 2 người, sàn 10 → mỗi khu ≥ 10, cả phòng ≈ 50.
func TestZoneFloorPerZone(t *testing.T) {
	m, cfg := zoneManager(t)
	cfg.Room.ZoneMinChars = 10
	joinN(m, 2)
	r := m.AllRooms()[0]
	dt := 1 / float64(cfg.Net.TickHz)
	for i := 0; i < int(float64(cfg.Net.TickHz)*cfg.Bots.SpawnEverySecs*40); i++ {
		r.Step(dt)
	}
	floorAll := cfg.Room.ZoneMinChars * len(r.Zones())
	// Bot đi qua cổng có thể làm một khu dư vài con (vơi dần khi chết), nhưng không được
	// nổ số: trần mềm = sàn × số khu + một nhịp sinh mỗi khu.
	if total := r.Humans() + r.Bots(); total < floorAll || total > floorAll+cfg.Bots.SpawnBatch*len(r.Zones()) {
		t.Fatalf("phòng có %d nhân vật, muốn ≈ %d (sàn %d × %d khu)", total, floorAll, cfg.Room.ZoneMinChars, len(r.Zones()))
	}
	// Sàn mỗi khu là lời hứa "khu tụt xuống dưới sàn thì được vá lại NGAY", không phải
	// "không khoảnh khắc nào dưới sàn". Hai thứ đó khác nhau, và chỉ cái thứ nhất là đúng:
	// bot đi qua cổng bất cứ lúc nào, nên một khu có thể mất người ở đúng tick cuối cùng,
	// và lệnh điều bù chỉ chạy được ở ranh giới tick sau. Bản cũ của bài test đo đúng một
	// khoảnh khắc nên đỏ rải rác (khoảng 3/40 lượt) mà chẳng có gì hỏng thật.
	//
	// Nên ở đây cho hệ thống một khoảng ngắn để tự chỉnh, rồi mới kết luận. Nếu luật vá sàn
	// hỏng thật thì chạy bao lâu cũng không đạt, bài test vẫn bắt được.
	settle := int(float64(cfg.Net.TickHz) * 1.0)
	below := func() (*Zone, []int) {
		counts := make([]int, 0, len(r.Zones()))
		var bad *Zone
		for _, q := range r.Zones() {
			counts = append(counts, q.Count())
			if bad == nil && q.Count() < cfg.Room.ZoneMinChars {
				bad = q
			}
		}
		return bad, counts
	}
	for i := 0; i < settle; i++ {
		if bad, _ := below(); bad == nil {
			break
		}
		r.Step(dt)
	}
	if bad, counts := below(); bad != nil {
		t.Fatalf("sau %d tick vẫn có khu dưới sàn: %s chỉ %d nhân vật, sàn là %d (phân bố %v)",
			settle, bad.Name, bad.Count(), cfg.Room.ZoneMinChars, counts)
	}
}
