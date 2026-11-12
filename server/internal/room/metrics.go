package room

import (
	"bladeascend/server/internal/metrics"
)

// Số liệu hiệu năng của MỘT khu, gom trong vòng tick (ADMIN_SPEC §12.1).
//
// Khác với `tickStats` (cửa sổ trượt 10 giây cho telemetry và trang admin), chỗ này là
// SỔ TÍCH LUỸ: ai vét thì lấy hết và đặt lại về 0. Hai sổ riêng vì hai người đọc có nhịp
// khác nhau (telemetry 10 s, bộ thu gom 1 s) và nếu dùng chung thì người này vét mất số
// của người kia.

// zoneCounters: mỗi hệ thống một bộ đếm. Tên trường đi thẳng ra tên chỉ số ở §12.1.
type zoneCounters struct {
	tick    metrics.Counter // tổng một tick
	input   metrics.Counter // nhận và áp input người chơi
	bots    metrics.Counter // bot AI
	move    metrics.Counter // di chuyển + quay vòng kiếm
	skills  metrics.Counter // trạng thái kỹ năng + hồi chiêu (v6.48)
	collide metrics.Counter // kiếm↔thân, kiếm↔kiếm, kiếm↔vật cản, nhặt đồ
	world   metrics.Counter // sinh quái/đồ, đài, hết hạn, danh hiệu, lưới
	deaths  metrics.Counter // xử lý chết + hồi sinh
	send    metrics.Counter // AOI + đóng gói + gửi snapshot
	// lastCasts/lastDeaths: giá trị bộ đếm cộng dồn của `Game` ở lần vét trước, để ra số
	// LƯỢT trong khoảng (§12.1 "lượt kỹ năng/phút, lượt chết/phút").
	lastCasts, lastDeaths int
}

// ZoneMetrics là một lần vét sổ của một khu.
type ZoneMetrics struct {
	Zone int
	Name string
	// Thời gian, tính bằng mili-giây.
	Tick, Input, Bots, Move, Skills, Collide, World, Deaths, Send metrics.Dist
	// Lượt kể từ lần vét trước.
	NDeaths, NSkills int
	// Ảnh chụp tại thời điểm vét.
	Humans, BotCount, Monsters, Items, BrokenObstacles, Swords int
	// SentBytes: tổng byte đã gửi cho khu này từ lúc server chạy — CỘNG DỒN.
	//
	// Không dùng `ZoneKbpsOut` ở đây: hàm đó VÉT bộ đếm (`sentSince` swap), nên bộ thu gom
	// 1 giây sẽ ăn mất số của bộ lấy mẫu telemetry 10 giây và cột KB/s trong `docs/perf/real/`
	// sẽ tụt xuống gần 0 mà không ai hiểu vì sao.
	SentBytes int64
}

// DrainMetrics vét sổ của mọi khu trong phòng. PHẢI gọi qua `Room.Query` — nó đọc dữ liệu
// mà goroutine phòng đang ghi mỗi tick (xem query.go).
func (r *Room) DrainMetrics() []ZoneMetrics {
	out := make([]ZoneMetrics, 0, len(r.zones))
	for _, z := range r.zones {
		m := ZoneMetrics{
			Zone: z.Index, Name: z.Name,
			Tick: z.m.tick.Drain(), Input: z.m.input.Drain(), Bots: z.m.bots.Drain(),
			Move: z.m.move.Drain(), Skills: z.m.skills.Drain(), Collide: z.m.collide.Drain(),
			World: z.m.world.Drain(), Deaths: z.m.deaths.Drain(), Send: z.m.send.Drain(),
			Humans: z.Humans(), BotCount: z.Count() - z.Humans(),
			Monsters: len(z.G.Monsters), Items: z.G.ItemCount(),
			SentBytes: z.sentTotal(),
		}
		casts, deaths := z.G.Counters()
		m.NSkills, m.NDeaths = casts-z.m.lastCasts, deaths-z.m.lastDeaths
		z.m.lastCasts, z.m.lastDeaths = casts, deaths
		for _, o := range z.G.Obstacles {
			if o.Broken {
				m.BrokenObstacles++
			}
		}
		// Số KIẾM đang bay trong khu: nghi can số một của chi phí va chạm (kiếm↔kiếm là
		// phép so từng cặp). Một khu 30 bot Phàm Nhân và một khu 30 bot Đạo Tổ có cùng số
		// nhân vật nhưng số kiếm chênh nhau cả chục lần — không có cột này thì hai cái đó
		// trông y hệt nhau trên biểu đồ.
		for _, p := range z.G.PlayerList() {
			if !p.Dead {
				m.Swords += p.Ring.Count
			}
		}
		out = append(out, m)
	}
	return out
}

// sentTotal cộng byte đã gửi của mọi client trong khu, KHÔNG đặt lại bộ đếm nào.
func (z *Zone) sentTotal() int64 {
	var n int64
	for _, c := range z.clients {
		n += c.sentBytes.Load()
	}
	return n
}
