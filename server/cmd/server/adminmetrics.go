package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

// API cho trang "Hiệu năng" (ADMIN_SPEC §12.3). Chỉ đọc; mọi đường đều qua `adminGuard`.

// seriesPoint là một khoảng trên biểu đồ.
type seriesPoint struct {
	TS    int64   `json:"ts"`
	Count int64   `json:"count"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Avg   float64 `json:"avg"`
	P95   float64 `json:"p95"`
	Build string  `json:"build"`
}

// resTable đổi tên độ phân giải người dùng chọn thành tên bảng.
func resTable(res string) (string, time.Duration, bool) {
	switch res {
	case "minute", "":
		return "metrics_minute", time.Minute, true
	case "hour":
		return "metrics_hour", time.Hour, true
	case "day":
		return "metrics_day", 24 * time.Hour, true
	}
	return "", 0, false
}

// seriesQuery đọc tham số chung của hai điểm cuối (JSON và CSV).
func (s *server) seriesQuery(r *http.Request) (table string, from, to int64, metric, scope string, err error) {
	q := r.URL.Query()
	table, _, ok := resTable(q.Get("res"))
	if !ok {
		return "", 0, 0, "", "", fmt.Errorf("res phải là minute | hour | day")
	}
	now := time.Now()
	to = now.Unix()
	from = now.Add(-24 * time.Hour).Unix()
	if v := q.Get("from"); v != "" {
		if n, e := strconv.ParseInt(v, 10, 64); e == nil {
			from = n
		}
	}
	if v := q.Get("to"); v != "" {
		if n, e := strconv.ParseInt(v, 10, 64); e == nil {
			to = n
		}
	}
	metric, scope = q.Get("metric"), q.Get("scope")
	if metric == "" {
		return "", 0, 0, "", "", fmt.Errorf("thiếu tham số metric")
	}
	return table, from, to, metric, scope, nil
}

// handleAdminMetrics trả chuỗi điểm của MỘT chỉ số. Trang chồng hai chỉ số bằng cách gọi
// hai lần — rẻ hơn và đơn giản hơn một API nhận danh sách.
func (s *server) handleAdminMetrics(w http.ResponseWriter, r *http.Request) {
	if s.metricsDB == nil {
		http.Error(w, "bộ thu gom số liệu đang tắt", http.StatusNotFound)
		return
	}
	table, from, to, metric, scope, err := s.seriesQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rows, err := s.metricsDB.Query(table, from, to, metric, scope, 20000)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out := make([]seriesPoint, 0, len(rows))
	for _, x := range rows {
		out = append(out, seriesPoint{TS: x.TS, Count: x.Count, Min: x.Min, Max: x.Max, Avg: x.Avg(), P95: x.P95, Build: x.Build})
	}
	marks, err := s.metricsDB.DeployMarks(from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"points": out, "deploys": marks, "budget": budgetOf(metric)})
}

// budgetOf: đường ngưỡng ngân sách vẽ trên biểu đồ (§12.3). 0 = chỉ số này không có ngưỡng.
func budgetOf(metric string) float64 {
	switch metric {
	case "tick_ms":
		return 4 // ngân sách tick §13.6
	case "kbps_out":
		return 40 // ngân sách băng thông mỗi client §13.6
	}
	return 0
}

// handleAdminMetricsCSV xuất đúng những gì đang xem ra CSV (§12.3).
func (s *server) handleAdminMetricsCSV(w http.ResponseWriter, r *http.Request) {
	if s.metricsDB == nil {
		http.Error(w, "bộ thu gom số liệu đang tắt", http.StatusNotFound)
		return
	}
	table, from, to, metric, scope, err := s.seriesQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rows, err := s.metricsDB.Query(table, from, to, metric, scope, 20000)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s_%s.csv"`, metric, scope))
	cw := csv.NewWriter(w)
	defer cw.Flush()
	_ = cw.Write([]string{"thoi_gian", "ts", "metric", "scope", "build", "count", "min", "max", "avg", "p95"})
	for _, x := range rows {
		_ = cw.Write([]string{
			time.Unix(x.TS, 0).Format("2006-01-02 15:04:05"), strconv.FormatInt(x.TS, 10),
			x.Metric, x.Scope, x.Build, strconv.FormatInt(x.Count, 10),
			strconv.FormatFloat(x.Min, 'f', 3, 64), strconv.FormatFloat(x.Max, 'f', 3, 64),
			strconv.FormatFloat(x.Avg(), 'f', 3, 64), strconv.FormatFloat(x.P95, 'f', 3, 64),
		})
	}
}

// handleAdminMetricsCatalog liệt kê chỉ số/phạm vi đang có + số phút đã bỏ vì không ai chơi.
func (s *server) handleAdminMetricsCatalog(w http.ResponseWriter, r *http.Request) {
	if s.metricsDB == nil {
		http.Error(w, "bộ thu gom số liệu đang tắt", http.StatusNotFound)
		return
	}
	cat, err := s.metricsDB.Catalog(time.Now().Add(-14 * 24 * time.Hour).Unix())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dropped := 0
	if s.metrics != nil {
		dropped = s.metrics.Dropped()
	}
	writeJSON(w, map[string]any{"metrics": cat, "droppedMinutes": dropped, "zones": zoneNames(s)})
}

// zoneNames: tên khu cho ô chọn phạm vi, để không phải đọc "zone:3" mà đoán.
func zoneNames(s *server) map[string]string {
	out := map[string]string{"server": "toàn server", "machine": "máy chủ"}
	for _, rm := range s.mgr.AllRooms() {
		rm.Query(func() {
			for i := 0; i < rm.ZoneCount(); i++ {
				z := rm.Zone(i)
				out["zone:"+strconv.Itoa(z.Index)] = z.Name
			}
		})
		break // năm khu giống nhau ở mọi phòng
	}
	return out
}

// handleAdminMetricsSummary: thẻ tóm tắt đầu trang — 24 giờ qua, GIỜ TỆ NHẤT của từng chỉ
// số chính, kèm số người thật lúc đó (§12.3).
func (s *server) handleAdminMetricsSummary(w http.ResponseWriter, r *http.Request) {
	if s.metricsDB == nil {
		http.Error(w, "bộ thu gom số liệu đang tắt", http.StatusNotFound)
		return
	}
	now := time.Now()
	from := now.Add(-24 * time.Hour).Unix()
	type card struct {
		Metric string  `json:"metric"`
		Scope  string  `json:"scope"`
		TS     int64   `json:"ts"`
		P95    float64 `json:"p95"`
		Avg    float64 `json:"avg"`
		Humans float64 `json:"humans"`
		Load1  float64 `json:"load1"`
	}
	// Số người và load của từng giờ, để gắn vào giờ tệ nhất.
	humans := map[int64]float64{}
	load := map[int64]float64{}
	if rows, err := s.metricsDB.Query("metrics_hour", from, now.Unix(), "humans", "server", 2000); err == nil {
		for _, x := range rows {
			humans[x.TS] = x.Max
		}
	}
	if rows, err := s.metricsDB.Query("metrics_hour", from, now.Unix(), "load1", "machine", 2000); err == nil {
		for _, x := range rows {
			load[x.TS] = x.Avg()
		}
	}
	var out []card
	for _, m := range []string{"tick_ms", "sys_collide_ms", "sys_bots_ms", "sys_send_ms", "kbps_out", "gc_pause_ms", "heap_mb"} {
		rows, err := s.metricsDB.Query("metrics_hour", from, now.Unix(), m, "", 5000)
		if err != nil || len(rows) == 0 {
			continue
		}
		worst := rows[0]
		for _, x := range rows {
			if x.P95 > worst.P95 {
				worst = x
			}
		}
		out = append(out, card{Metric: m, Scope: worst.Scope, TS: worst.TS, P95: worst.P95,
			Avg: worst.Avg(), Humans: humans[worst.TS], Load1: load[worst.TS]})
	}
	writeJSON(w, map[string]any{"cards": out, "alerts": s.metricAlerts(now), "from": from, "to": now.Unix()})
}

// metricAlert là một cảnh báo cho dashboard (§12.3 dòng cuối).
type metricAlert struct {
	Metric  string  `json:"metric"`
	Scope   string  `json:"scope"`
	Mins    int     `json:"mins"`
	P95     float64 `json:"p95"`
	Budget  float64 `json:"budget"`
	Load1   float64 `json:"load1"`
	Message string  `json:"message"`
}

// alertMinutes: phải vượt ngưỡng liên tục ngần này phút mới báo (§12.3).
const alertMinutes = 5

// metricAlerts tìm những chỉ số đã vượt ngân sách LIÊN TỤC ≥ 5 phút gần đây.
//
// Kèm load máy cùng lúc, và nếu máy đang bận thì nói thẳng ra trong câu gợi ý: trên máy dev
// dùng chung, "tick 12 ms" lúc load 25 là câu chuyện về cái máy chứ không phải về game —
// đúng cái bẫy đã làm hỏng mấy lượt đo (PERF_LOG B20, B13c, lượt #34).
func (s *server) metricAlerts(now time.Time) []metricAlert {
	from := now.Add(-20 * time.Minute).Unix()
	load := map[int64]float64{}
	if rows, err := s.metricsDB.Query("metrics_minute", from, now.Unix(), "load1", "machine", 200); err == nil {
		for _, x := range rows {
			load[x.TS] = x.Avg()
		}
	}
	var out []metricAlert
	for _, m := range []string{"tick_ms", "kbps_out"} {
		budget := budgetOf(m)
		rows, err := s.metricsDB.Query("metrics_minute", from, now.Unix(), m, "", 2000)
		if err != nil {
			continue
		}
		// Gom theo phạm vi rồi đếm chuỗi phút vượt ngưỡng ở CUỐI dãy.
		byScope := map[string][]Row2{}
		for _, x := range rows {
			byScope[x.Scope] = append(byScope[x.Scope], Row2{TS: x.TS, P95: x.P95})
		}
		for scope, xs := range byScope {
			streak, p95, loadSum := 0, 0.0, 0.0
			for i := len(xs) - 1; i >= 0; i-- {
				if xs[i].P95 <= budget {
					break
				}
				streak++
				p95 = math.Max(p95, xs[i].P95)
				loadSum += load[xs[i].TS]
			}
			if streak < alertMinutes {
				continue
			}
			avgLoad := loadSum / float64(streak)
			msg := fmt.Sprintf("%s ở %s vượt ngân sách %.1f suốt %d phút (p95 cao nhất %.1f).", m, scope, budget, streak, p95)
			if avgLoad >= float64(runtime.NumCPU())*1.0 {
				msg += fmt.Sprintf(" Load máy %.1f trên %d nhân — nhiều khả năng do MÁY đang bận, không phải game.", avgLoad, runtime.NumCPU())
			}
			out = append(out, metricAlert{Metric: m, Scope: scope, Mins: streak, P95: p95, Budget: budget, Load1: avgLoad, Message: msg})
		}
	}
	return out
}

// Row2 là cặp (mốc, p95) rút gọn — chỉ dùng trong `metricAlerts`.
type Row2 struct {
	TS  int64
	P95 float64
}
