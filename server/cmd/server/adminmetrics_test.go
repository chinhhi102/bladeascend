package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"bladeascend/server/internal/metrics"

	_ "modernc.org/sqlite"
)

// API trang Hiệu năng (ADMIN_SPEC §12.3). Gọi thẳng handler, không qua phiên đăng nhập —
// `adminGuard` đã có test riêng; chỗ này kiểm HỢP ĐỒNG mà trang web dựa vào.

func metricsServer(t *testing.T) *server {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "m.db"))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	mdb, err := metrics.Open(db)
	if err != nil {
		t.Fatal(err)
	}
	return &server{metricsDB: mdb, buildID: "test", metrics: metrics.New(mdb, "test")}
}

// seedMinutes ghi `n` phút tick_ms với p95 cho sẵn.
func seedMinutes(t *testing.T, s *server, metric, scope string, start time.Time, n int, p95 float64) {
	t.Helper()
	var rows []metrics.Row
	for i := 0; i < n; i++ {
		var d metrics.Dist
		for k := 0; k < 100; k++ {
			d.Add(p95 * 0.5)
		}
		d.Add(p95)
		rows = append(rows, metrics.Row{TS: start.Add(time.Duration(i) * time.Minute).Unix(),
			Metric: metric, Scope: scope, Build: "test",
			Count: d.Count, Min: d.Min, Max: d.Max, Sum: d.Sum, P95: p95, Hist: d.Hist})
	}
	if err := s.metricsDB.WriteMinute(rows); err != nil {
		t.Fatal(err)
	}
}

func TestAPIHieuNangTraChuoiDiem(t *testing.T) {
	s := metricsServer(t)
	start := time.Now().Add(-10 * time.Minute).Truncate(time.Minute)
	seedMinutes(t, s, "tick_ms", "zone:0", start, 10, 2.5)
	if err := s.metricsDB.MarkDeploy(start.Add(3*time.Minute).Unix(), "b1", "deploy thử"); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	s.handleAdminMetrics(w, httptest.NewRequest(http.MethodGet, "/api/admin/metrics?metric=tick_ms&scope=zone:0&res=minute", nil))
	if w.Code != 200 {
		t.Fatalf("HTTP %d: %s", w.Code, w.Body.String())
	}
	var got struct {
		Points []struct {
			TS  int64   `json:"ts"`
			Avg float64 `json:"avg"`
			P95 float64 `json:"p95"`
		} `json:"points"`
		Deploys []struct{ TS int64 } `json:"deploys"`
		Budget  float64              `json:"budget"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Points) != 10 {
		t.Errorf("phải có 10 điểm, có %d", len(got.Points))
	}
	if got.Budget != 4 {
		t.Errorf("tick_ms phải kèm đường ngân sách 4 ms, có %v", got.Budget)
	}
	if len(got.Deploys) != 1 {
		t.Errorf("phải có 1 vạch deploy, có %d", len(got.Deploys))
	}
}

func TestAPIHieuNangXuatCSV(t *testing.T) {
	s := metricsServer(t)
	start := time.Now().Add(-5 * time.Minute).Truncate(time.Minute)
	seedMinutes(t, s, "tick_ms", "zone:0", start, 5, 3)
	w := httptest.NewRecorder()
	s.handleAdminMetricsCSV(w, httptest.NewRequest(http.MethodGet, "/api/admin/metrics.csv?metric=tick_ms&scope=zone:0&res=minute", nil))
	if w.Code != 200 {
		t.Fatalf("HTTP %d", w.Code)
	}
	lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
	if len(lines) != 6 { // 1 dòng tiêu đề + 5 phút
		t.Fatalf("CSV phải có 6 dòng, có %d:\n%s", len(lines), w.Body.String())
	}
	if !strings.HasPrefix(lines[0], "thoi_gian,ts,metric") {
		t.Errorf("thiếu dòng tiêu đề: %q", lines[0])
	}
}

// §12.3: vượt ngân sách LIÊN TỤC ≥ 5 phút mới cảnh báo; 4 phút thì im.
func TestCanhBaoChiKhiVuotLienTuc5Phut(t *testing.T) {
	s := metricsServer(t)
	now := time.Now().Truncate(time.Minute)
	seedMinutes(t, s, "tick_ms", "zone:0", now.Add(-4*time.Minute), 4, 9) // 4 phút vượt
	if a := s.metricAlerts(now); len(a) != 0 {
		t.Fatalf("4 phút vượt thì chưa được cảnh báo, có %d: %+v", len(a), a)
	}
	seedMinutes(t, s, "tick_ms", "zone:0", now.Add(-9*time.Minute), 5, 9) // thành 9 phút liên tục
	alerts := s.metricAlerts(now)
	if len(alerts) != 1 {
		t.Fatalf("9 phút vượt phải cảnh báo đúng 1 lần, có %d", len(alerts))
	}
	if alerts[0].Mins < 5 || alerts[0].Budget != 4 {
		t.Errorf("cảnh báo sai: %+v", alerts[0])
	}
	if !strings.Contains(alerts[0].Message, "tick_ms") {
		t.Errorf("câu cảnh báo phải nói rõ chỉ số: %q", alerts[0].Message)
	}
}

// Máy đang bận thì cảnh báo phải NÓI RA — nếu không, người đọc lại đi sửa game trong khi
// thủ phạm là cái máy (PERF_LOG B20, lượt #34).
func TestCanhBaoNoiRaKhiMayBan(t *testing.T) {
	s := metricsServer(t)
	now := time.Now().Truncate(time.Minute)
	seedMinutes(t, s, "tick_ms", "zone:0", now.Add(-6*time.Minute), 6, 12)
	seedMinutes(t, s, "load1", "machine", now.Add(-6*time.Minute), 6, 999) // load cao chót vót
	alerts := s.metricAlerts(now)
	if len(alerts) != 1 {
		t.Fatalf("phải có 1 cảnh báo, có %d", len(alerts))
	}
	if !strings.Contains(alerts[0].Message, "MÁY đang bận") {
		t.Errorf("load cao mà câu cảnh báo không nhắc tới máy: %q", alerts[0].Message)
	}
}

func TestAPIHieuNangTatThiBao404(t *testing.T) {
	s := &server{}
	w := httptest.NewRecorder()
	s.handleAdminMetrics(w, httptest.NewRequest(http.MethodGet, "/api/admin/metrics?metric=tick_ms", nil))
	if w.Code != http.StatusNotFound {
		t.Errorf("tắt thu gom thì phải 404, có %d", w.Code)
	}
}
