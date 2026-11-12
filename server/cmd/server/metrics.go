package main

import (
	"bufio"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"bladeascend/server/internal/metrics"
	"bladeascend/server/internal/room"
)

// Vòng lấy mẫu 1 giây của bộ thu gom số liệu (ADMIN_SPEC §12.2).
//
// Chạy NGOÀI vòng tick. Chỗ duy nhất chạm goroutine phòng là `Room.Query` ở `sampleMetrics`,
// và nó chỉ vét vài bộ đếm rồi thoát — cùng kiểu với `sampleRooms` của telemetry.

const metricSampleEvery = time.Second

// runMetricsCollector lấy mẫu mỗi giây cho tới khi server tắt.
func (s *server) runMetricsCollector(stop <-chan struct{}) {
	t := time.NewTicker(metricSampleEvery)
	defer t.Stop()
	var cpu cpuSnap
	var mach machSnap
	last := map[int]int64{} // byte đã gửi của từng khu ở lần lấy mẫu trước
	var lastGC uint64
	for {
		select {
		case <-stop:
			s.metrics.Close(time.Now())
			return
		case <-t.C:
			now := time.Now()
			s.sampleMetrics(now, &cpu, &mach, last, &lastGC)
			s.metrics.Flush(now)
		}
	}
}

func (s *server) sampleMetrics(now time.Time, cpu *cpuSnap, mach *machSnap, last map[int]int64, lastGC *uint64) {
	c := s.metrics
	humans := 0
	for _, rm := range s.mgr.AllRooms() {
		var zs []room.ZoneMetrics
		if !rm.Query(func() { zs = rm.DrainMetrics() }) {
			continue // phòng quá tải: bỏ một nhịp, không treo bộ thu gom
		}
		for _, z := range zs {
			scope := "zone:" + strconv.Itoa(z.Zone)
			humans += z.Humans
			c.AddDist("tick_ms", scope, z.Tick)
			c.AddDist("sys_input_ms", scope, z.Input)
			c.AddDist("sys_bots_ms", scope, z.Bots)
			c.AddDist("sys_move_ms", scope, z.Move)
			c.AddDist("sys_skills_ms", scope, z.Skills)
			c.AddDist("sys_collide_ms", scope, z.Collide)
			c.AddDist("sys_world_ms", scope, z.World)
			c.AddDist("sys_deaths_ms", scope, z.Deaths)
			c.AddDist("sys_send_ms", scope, z.Send)
			c.Add("humans", scope, float64(z.Humans))
			c.Add("bots", scope, float64(z.BotCount))
			c.Add("monsters", scope, float64(z.Monsters))
			c.Add("items", scope, float64(z.Items))
			c.Add("obstacles_broken", scope, float64(z.BrokenObstacles))
			c.Add("swords", scope, float64(z.Swords))
			c.Add("skills_per_min", scope, float64(z.NSkills)*60/metricSampleEvery.Seconds())
			c.Add("deaths_per_min", scope, float64(z.NDeaths)*60/metricSampleEvery.Seconds())
			if prev, ok := last[z.Zone]; ok && z.SentBytes >= prev {
				c.Add("kbps_out", scope, float64(z.SentBytes-prev)/1024/metricSampleEvery.Seconds())
			}
			last[z.Zone] = z.SentBytes
		}
	}
	c.NoteHumans(humans)
	c.Add("humans", "server", float64(humans))

	// Tiến trình server.
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	c.Add("heap_mb", "server", float64(ms.HeapAlloc)/(1024*1024))
	c.Add("goroutines", "server", float64(runtime.NumGoroutine()))
	// GC: cộng dồn từ lúc chạy → lấy hiệu để ra mili-giây dừng TRONG giây vừa rồi.
	if *lastGC > 0 && ms.PauseTotalNs >= *lastGC {
		c.Add("gc_pause_ms", "server", float64(ms.PauseTotalNs-*lastGC)/1e6)
	}
	*lastGC = ms.PauseTotalNs
	pct, _ := cpuPercent(cpu)
	c.Add("cpu_pct", "server", pct)

	// Cái MÁY (§12.1) — để phân biệt "game chậm" với "máy đang bận việc khác". Đây chính là
	// câu hỏi đã làm hỏng mấy lượt đo trước (PERF_LOG B20, B13c, lượt #34).
	if l, ok := loadAvg1(); ok {
		c.Add("load1", "machine", l)
	}
	if p, ok := mach.pct(); ok {
		c.Add("cpu_pct", "machine", p)
	}
	if f, ok := memAvailableMB(); ok {
		c.Add("mem_free_mb", "machine", f)
	}
}

// loadAvg1 đọc load 1 phút của cả máy.
func loadAvg1() (float64, bool) {
	b, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, false
	}
	f := strings.Fields(string(b))
	if len(f) == 0 {
		return 0, false
	}
	v, err := strconv.ParseFloat(f[0], 64)
	return v, err == nil
}

// machSnap giữ mốc /proc/stat lần trước để tính %CPU toàn máy giữa hai lần gọi.
type machSnap struct{ total, idle float64 }

func (m *machSnap) pct() (float64, bool) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0, false
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	if !sc.Scan() {
		return 0, false
	}
	fields := strings.Fields(sc.Text())
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0, false
	}
	var total, idle float64
	for i, s := range fields[1:] {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, false
		}
		total += v
		if i == 3 || i == 4 { // idle + iowait
			idle += v
		}
	}
	prevTotal, prevIdle := m.total, m.idle
	m.total, m.idle = total, idle
	dt := total - prevTotal
	if prevTotal == 0 || dt <= 0 {
		return 0, false
	}
	return 100 * (1 - (idle-prevIdle)/dt), true
}

// memAvailableMB: RAM còn dùng được (MemAvailable, không phải MemFree — MemFree trên máy đã
// chạy lâu luôn nhỏ vì cache chiếm hết, và đọc nó ra là tưởng máy sắp hết RAM).
func memAvailableMB() (float64, bool) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, false
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "MemAvailable:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return 0, false
		}
		kb, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			return 0, false
		}
		return kb / 1024, true
	}
	return 0, false
}

// startMetrics mở file số liệu và ghi mốc deploy của bản đang chạy.
//
// File RIÊNG, không dùng chung với bảng xếp hạng hay telemetry: bảng này ghi mỗi phút và bị
// dọn theo hạn 14/180 ngày, trong khi `leaderboard.db` là dữ liệu người chơi phải giữ.
func startMetrics(path, build string, on bool) *metrics.DB {
	if !on {
		return nil
	}
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)")
	if err != nil {
		log.Printf("metrics: không mở được %s: %v — chạy tiếp, không thu số liệu", path, err)
		return nil
	}
	db.SetMaxOpenConns(1)
	mdb, err := metrics.Open(db)
	if err != nil {
		log.Printf("metrics: %v — chạy tiếp, không thu số liệu", err)
		return nil
	}
	if err := mdb.MarkDeploy(time.Now().Unix(), build, "server khởi động"); err != nil {
		log.Printf("metrics: ghi mốc deploy: %v", err)
	}
	log.Printf("metrics: thu số liệu hiệu năng vào %s (build %s)", path, build)
	return mdb
}

// metricsPath mặc định đặt cạnh telemetry, file riêng (§12.2).
func metricsPath(path string) string {
	if path != "" {
		return path
	}
	return filepath.Join(filepath.Dir(telemetryPath("")), "metrics.db")
}
