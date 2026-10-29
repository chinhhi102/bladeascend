// Command stress tăng dần số người chơi cho tới khi máy này không còn gánh nổi, để biết
// trước khi người thật vào đông thì phòng tụt ở mốc nào.
//
//	go run ./cmd/stress                      # 10 → 100 người, mỗi mốc 10 giây
//	go run ./cmd/stress -from 20 -to 200 -step 20 -secs 15
//
// Khác `cmd/sim` ở chỗ: sim chỉ chạy `game.Tick` thuần, còn stress chạy **nguyên đường thật**
// — ghép phòng, bot, va chạm, dựng snapshot theo tầm nhìn từng người, mã hoá msgpack, gửi đi.
// Phần đắt nhất của server nằm ở dựng + mã hoá snapshot chứ không nằm ở Tick, nên chỉ đo Tick
// là thấy mọi thứ đều nhẹ trong khi thực tế đã nghẽn.
package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"kiemxoay/server/internal/config"
	"kiemxoay/server/internal/netproto"
	"kiemxoay/server/internal/room"
)

// sink đóng vai một người chơi đang kết nối: chỉ đếm số byte server đẩy xuống.
type sink struct {
	bytes int64
	msgs  int64
}

func (s *sink) Send(b []byte) bool {
	s.bytes += int64(len(b))
	s.msgs++
	return true
}

type result struct {
	players   int
	rooms     int
	items     int
	realRatio float64 // thời gian CPU / thời gian game, 1.0 = vừa đủ thời gian thực
	stepAvgMs float64
	stepP99Ms float64
	kbPerSec  float64 // băng thông xuống mỗi người
}

func main() {
	from := flag.Int("from", 10, "số người chơi bắt đầu")
	to := flag.Int("to", 100, "số người chơi tối đa")
	step := flag.Int("step", 10, "mỗi mốc tăng bao nhiêu người")
	secs := flag.Float64("secs", 10, "mỗi mốc chạy bao nhiêu giây game")
	budget := flag.Float64("budget", 0.8, "ngưỡng coi là lag: dùng quá bao nhiêu phần thời gian thực")
	seed := flag.Int64("seed", 1, "random seed")
	flag.Parse()

	cfg := config.MustLoadDefault()
	fmt.Printf("Máy này gánh được bao nhiêu người? (tick %d/giây, snapshot %d/giây, ngưỡng lag %.0f%% thời gian thực)\n\n",
		cfg.Net.TickHz, cfg.Net.SnapshotHz, *budget*100)
	fmt.Printf("%7s %6s %7s %10s %10s %10s %11s %s\n",
		"người", "phòng", "vật phẩm", "CPU/thực", "bước TB", "bước p99", "KB/s mỗi", "kết luận")

	var last *result
	for n := *from; n <= *to; n += *step {
		r := measure(cfg, n, *secs, *seed)
		verdict := "mượt"
		if r.realRatio >= *budget {
			verdict = "LAG"
		}
		fmt.Printf("%7d %6d %8d %9.0f%% %9.2fms %9.2fms %10.1f  %s\n",
			r.players, r.rooms, r.items, r.realRatio*100, r.stepAvgMs, r.stepP99Ms, r.kbPerSec, verdict)
		if r.realRatio >= *budget {
			fmt.Printf("\nNgưỡng: gánh mượt tới khoảng %d người, %d người là bắt đầu lag.\n", lastPlayers(last), r.players)
			return
		}
		last = &r
	}
	fmt.Printf("\nChạy hết %d người vẫn mượt (cao nhất %.0f%% thời gian thực). Tăng -to để tìm trần thật.\n",
		*to, worst(last))
}

func lastPlayers(r *result) int {
	if r == nil {
		return 0
	}
	return r.players
}

func worst(r *result) float64 {
	if r == nil {
		return 0
	}
	return r.realRatio * 100
}

// measure chạy n người trong secs giây game và đo xem tốn bao nhiêu thời gian CPU.
func measure(cfg *config.Config, n int, secs float64, seed int64) result {
	c := *cfg // bản sao: không để mốc này ảnh hưởng mốc sau
	m := room.NewManager(&c, nil, nil)
	m.Start = func(*room.Room) {} // tự bước tay, không chạy goroutine riêng
	rng := rand.New(rand.NewSource(seed))
	sinks := make([]*sink, 0, n)
	for i := 0; i < n; i++ {
		s := &sink{}
		sinks = append(sinks, s)
		m.Join(room.NewClient(s, false), netproto.Join{Name: fmt.Sprintf("Nguoi%03d", i), Char: "a_hon"})
	}
	rooms := m.AllRooms()

	dt := 1 / float64(c.Net.TickHz)
	ticks := int(secs / dt)
	// Chạy 2 giây cho vật phẩm, EXP và bot mọc đầy rồi mới bấm giờ.
	for i := 0; i < int(2/dt); i++ {
		for _, r := range rooms {
			r.Step(dt)
		}
	}
	for _, s := range sinks {
		s.bytes, s.msgs = 0, 0
	}

	durs := make([]float64, 0, ticks)
	t0 := time.Now()
	for i := 0; i < ticks; i++ {
		// Người thật thì luôn bấm phím đi loanh quanh; không có input thì họ đứng yên và
		// server được lợi oan (ít va chạm, ít vật phẩm nhặt).
		if i%6 == 0 {
			feedInputs(rooms, rng)
		}
		s0 := time.Now()
		for _, r := range rooms {
			r.Step(dt)
		}
		durs = append(durs, float64(time.Since(s0).Microseconds())/1000)
	}
	wall := time.Since(t0).Seconds()

	sort.Float64s(durs)
	sum := 0.0
	for _, d := range durs {
		sum += d
	}
	var bytes int64
	items := 0
	for _, s := range sinks {
		bytes += s.bytes
	}
	for _, r := range rooms {
		items += len(r.Zone(0).G.Items)
	}
	return result{
		players:   n,
		rooms:     len(rooms),
		items:     items,
		realRatio: wall / secs,
		stepAvgMs: sum / float64(max(len(durs), 1)),
		stepP99Ms: durs[int(math.Min(float64(len(durs)-1), float64(len(durs))*0.99))],
		kbPerSec:  float64(bytes) / 1024 / secs / float64(max(n, 1)),
	}
}

// feedInputs cho mỗi người một hướng đi ngẫu nhiên, đúng cách client gửi lên.
func feedInputs(rooms []*room.Room, rng *rand.Rand) {
	for _, r := range rooms {
		for _, c := range r.Clients() {
			c.SetInput(netproto.Input{T: "in", Angle: uint8(rng.Intn(256)), Moving: true})
		}
	}
}
