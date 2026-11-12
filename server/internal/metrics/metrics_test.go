package metrics

import (
	"database/sql"
	"math"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func openTest(t *testing.T) *DB {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "m.db"))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	m, err := Open(db)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// §12.4 mục 1: 10 phút có người → 10 dòng; 5 phút không ai → KHÔNG dòng nào.
func TestChiGhiPhutCoNguoiThat(t *testing.T) {
	db := openTest(t)
	c := New(db, "testbuild")
	base := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)
	for i := 0; i < 10; i++ { // 10 phút có 1 người
		for s := 0; s < 60; s++ {
			now := base.Add(time.Duration(i)*time.Minute + time.Duration(s)*time.Second)
			c.Flush(now)
			c.Add("tick_ms", "zone:0", 1+float64(s%3))
			c.NoteHumans(1)
		}
	}
	for i := 10; i < 15; i++ { // 5 phút không ai chơi (bot vẫn chạy nên vẫn có mẫu tick)
		for s := 0; s < 60; s++ {
			now := base.Add(time.Duration(i)*time.Minute + time.Duration(s)*time.Second)
			c.Flush(now)
			c.Add("tick_ms", "zone:0", 9)
			c.NoteHumans(0)
		}
	}
	c.Close(base.Add(15 * time.Minute))

	rows, err := db.Query("metrics_minute", base.Unix(), base.Add(time.Hour).Unix(), "tick_ms", "zone:0", 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 10 {
		t.Fatalf("phải có đúng 10 dòng phút (chỉ phút có người), có %d", len(rows))
	}
	for _, r := range rows {
		if r.Max > 3 {
			t.Errorf("dòng %d lẫn số của phút không có người (max %.1f)", r.TS, r.Max)
		}
		if r.Build != "testbuild" {
			t.Errorf("dòng phải mang build đang chạy, có %q", r.Build)
		}
	}
	if c.Dropped() != 5 {
		t.Errorf("phải đếm 5 phút đã bỏ, có %d", c.Dropped())
	}
}

// §12.4 mục 2: dòng giờ = gộp các phút — min của min, max của max, avg theo trọng số count.
func TestCuonGioKhopVoiCacPhut(t *testing.T) {
	db := openTest(t)
	base := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)
	// Ba phút với số mẫu KHÁC NHAU, để bắt lỗi "avg của giờ = trung bình các avg phút".
	mins := []struct {
		n        int64
		min, max float64
		sum      float64
	}{
		{n: 10, min: 1, max: 5, sum: 30},
		{n: 100, min: 2, max: 9, sum: 400},
		{n: 50, min: 0.5, max: 4, sum: 100},
	}
	var rows []Row
	for i, m := range mins {
		d := Dist{Count: m.n, Min: m.min, Max: m.max, Sum: m.sum, Hist: Hist{}}
		d.Hist[bucketOf(m.max)] = uint32(m.n)
		rows = append(rows, Row{TS: base.Add(time.Duration(i) * time.Minute).Unix(), Metric: "tick_ms",
			Scope: "zone:0", Build: "b", Count: d.Count, Min: d.Min, Max: d.Max, Sum: d.Sum, P95: d.P95(), Hist: d.Hist})
	}
	if err := db.WriteMinute(rows); err != nil {
		t.Fatal(err)
	}
	if err := db.Rollup(base.Add(90 * time.Minute)); err != nil {
		t.Fatal(err)
	}
	hrs, err := db.Query("metrics_hour", base.Add(-time.Hour).Unix(), base.Add(2*time.Hour).Unix(), "tick_ms", "zone:0", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hrs) != 1 {
		t.Fatalf("phải có đúng 1 dòng giờ, có %d", len(hrs))
	}
	h := hrs[0]
	if h.TS != base.Truncate(time.Hour).Unix() {
		t.Errorf("mốc giờ sai: %d", h.TS)
	}
	if h.Count != 160 {
		t.Errorf("count phải là 160, có %d", h.Count)
	}
	if h.Min != 0.5 {
		t.Errorf("min của giờ phải là min nhỏ nhất (0,5), có %v", h.Min)
	}
	if h.Max != 9 {
		t.Errorf("max của giờ phải là max lớn nhất (9), có %v", h.Max)
	}
	// avg theo TRỌNG SỐ: 530/160 = 3,3125. Trung bình cộng ba avg phút là 3,27 — khác.
	if got := h.Avg(); math.Abs(got-3.3125) > 1e-9 {
		t.Errorf("avg phải theo trọng số count: muốn 3,3125, có %v", got)
	}
}

// Cuộn hai lần không được nhân đôi dữ liệu (server tắt/bật giữa giờ).
func TestCuonHaiLanKhongNhanDoi(t *testing.T) {
	db := openTest(t)
	base := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)
	d := Dist{}
	for i := 0; i < 10; i++ {
		d.Add(float64(i))
	}
	row := Row{TS: base.Unix(), Metric: "x", Scope: "s", Count: d.Count, Min: d.Min, Max: d.Max, Sum: d.Sum, P95: d.P95(), Hist: d.Hist}
	if err := db.WriteMinute([]Row{row}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if err := db.Rollup(base.Add(2 * time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	hrs, _ := db.Query("metrics_hour", 0, base.Add(3*time.Hour).Unix(), "x", "s", 10)
	if len(hrs) != 1 || hrs[0].Count != 10 {
		t.Fatalf("cuộn 3 lần phải vẫn là 1 dòng, count 10 — có %d dòng, count %v", len(hrs), hrs[0].Count)
	}
}

func TestP95VaHistogram(t *testing.T) {
	var d Dist
	for i := 1; i <= 100; i++ {
		d.Add(float64(i))
	}
	// Histogram log sai số 4%: p95 thật là 95.
	if got := d.P95(); math.Abs(got-95) > 95*0.05 {
		t.Errorf("p95 phải quanh 95 (sai số ≤ 5%%), có %v", got)
	}
	if d.Min != 1 || d.Max != 100 {
		t.Errorf("min/max phải CHÍNH XÁC: %v %v", d.Min, d.Max)
	}
	// Mã hoá rồi giải mã phải giữ nguyên.
	back := DecodeHist(EncodeHist(d.Hist))
	if len(back) != len(d.Hist) {
		t.Fatalf("số ô sau khi giải mã khác: %d vs %d", len(back), len(d.Hist))
	}
	for k, v := range d.Hist {
		if back[k] != v {
			t.Errorf("ô %d: %d ≠ %d", k, back[k], v)
		}
	}
	// Một mẫu duy nhất: p95 không được vượt ra ngoài [min,max].
	var one Dist
	one.Add(3.0)
	if got := one.P95(); got != 3.0 {
		t.Errorf("một mẫu 3,0 thì p95 phải là 3,0, có %v", got)
	}
}

// Counter (bản dùng trong vòng tick) phải cho cùng kết quả như Dist.
func TestCounterKhopVoiDist(t *testing.T) {
	var c Counter
	var d Dist
	for i := 1; i <= 500; i++ {
		v := float64(i%37) / 3
		c.Add(v)
		d.Add(v)
	}
	got := c.Drain()
	if got.Count != d.Count || got.Min != d.Min || got.Max != d.Max || math.Abs(got.Sum-d.Sum) > 1e-9 {
		t.Errorf("Counter khác Dist: %+v vs %+v", got, d)
	}
	if got.P95() != d.P95() {
		t.Errorf("p95 khác: %v vs %v", got.P95(), d.P95())
	}
	if after := c.Drain(); after.Count != 0 {
		t.Errorf("Drain phải đặt lại bộ đếm, còn %d mẫu", after.Count)
	}
}

// §12.4 mục 4: thu gom không được làm tick tăng quá 0,1 ms.
//
// Phần thêm vào mỗi tick là: 5 lần đọc đồng hồ trong `game.Tick`/`stepZone` và 9 lần
// `Counter.Add`. Bench đúng chừng đó, rồi in ra số mili-giây mỗi tick.
func BenchmarkThuGomMoiTick(b *testing.B) {
	var cs [9]Counter
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		t0 := time.Now()
		for j := range cs {
			t1 := time.Now()
			cs[j].Add(float64(t1.Sub(t0).Nanoseconds()) / 1e6)
			if j%2 == 0 { // ~5 lần đọc đồng hồ mỗi tick
				t0 = t1
			}
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/1e6, "ms/tick")
}
