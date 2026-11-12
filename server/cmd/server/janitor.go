package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"bladeascend/server/internal/admin"
	"bladeascend/server/internal/safe"
)

// Bộ dọn dữ liệu — ADMIN_SPEC §13 (v1.6).
//
// Phần lịch chạy và phần HTTP nằm ở đây; luật giữ dữ liệu nằm trong `internal/admin`
// (janitor.go). Chia vậy vì luật giữ là thứ phải đọc được một mình, không lẫn với chuyện
// đồng hồ và route.
//
// ── Vì sao 04:30 ─────────────────────────────────────────────────────────────────────────
//
// Spec chốt giờ đó, và nó đúng với dữ liệu thật: khung 4–5 giờ sáng là lúc duy nhất trong
// ngày server LAN không có ai. VACUUM khoá cả tệp và chép lại toàn bộ — chạy lúc có người
// là vài giây đứng hình giữa trận.

// janitorHour / janitorMinute: giờ chạy hằng ngày (§13: 04:30).
const (
	janitorHour   = 4
	janitorMinute = 30
)

// nextJanitorRun là mốc 04:30 kế tiếp tính từ `now`.
//
// Tính bằng đồng hồ TƯỜNG, không bằng "cứ 24 giờ một lần": server khởi động lại nhiều lần
// trong ngày, và nhịp 24 giờ tính từ lúc khởi động sẽ trôi dần tới đúng giờ cao điểm.
func nextJanitorRun(now time.Time) time.Time {
	at := time.Date(now.Year(), now.Month(), now.Day(), janitorHour, janitorMinute, 0, 0, now.Location())
	if !at.After(now) {
		at = at.AddDate(0, 0, 1)
	}
	return at
}

func (s *server) runJanitor(stop <-chan struct{}) {
	if s.janitor == nil {
		return
	}
	for {
		wait := time.Until(nextJanitorRun(time.Now()))
		t := time.NewTimer(wait)
		select {
		case <-stop:
			t.Stop()
			return
		case <-t.C:
			s.janitorNightly(time.Now())
		}
	}
}

// janitorNightly là lượt chạy 04:30.
//
// Chưa được "bật" (xem `Janitor.Armed`) thì nó CHỈ chạy bản xem trước và ghi danh sách vào
// nhật ký — không xoá một dòng nào. Chủ dự án bấm "Dọn ngay" một lần ở trang quản trị là
// bật, và từ đêm sau job này xoá thật.
func (s *server) janitorNightly(now time.Time) {
	if s.janitor.Armed() {
		s.janitorOnce(now)
		return
	}
	plan := s.janitor.Plan(now)
	log.Printf("dọn dữ liệu (CHƯA BẬT — chỉ chạy thử): sẽ xoá %d dòng. Bấm \"Dọn ngay\" ở trang quản trị để bật.", plan.Rows)
	s.adminAuth.Store().Audit("janitor_dryrun", "", "chạy thử tự động 04:30 (chưa được bật xoá thật)", "system",
		map[string]any{"rows": plan.Rows, "jobs": plan.Jobs})
}

// janitorOnce chạy một lượt dọn và ghi vào nhật ký quản trị.
//
// Ghi nhật ký là bắt buộc, không phải cho đẹp: một việc tự động âm thầm xoá dữ liệu là thứ
// không ai phát hiện ra cho tới khi cần đúng dữ liệu đó.
func (s *server) janitorOnce(now time.Time) admin.JanitorResult {
	res := s.janitor.Run(now)
	log.Printf("dọn dữ liệu: xoá %d dòng · VACUUM %v · giải phóng %.1f MB%s",
		res.Rows, res.Vacuumed, float64(res.FreedBytes)/(1<<20), vacuumSuffix(res))
	s.adminAuth.Store().Audit("janitor_run", "", "dọn tự động 04:30", "system", map[string]any{
		"rows": res.Rows, "vacuumed": res.Vacuumed, "freed_bytes": res.FreedBytes,
		"jobs": res.Jobs,
	})
	return res
}

func vacuumSuffix(res admin.JanitorResult) string {
	if res.VacuumNote == "" {
		return ""
	}
	return " · " + res.VacuumNote
}

// janitorView là payload của trang "Dọn dữ liệu" (§13).
type janitorView struct {
	Tables []admin.TableSize   `json:"tables"`
	Files  []admin.FileSize    `json:"files"`
	Plan   admin.JanitorResult `json:"plan"`
	// Last: lượt chạy thật gần nhất, đọc từ `admin_audit`.
	Last    *admin.AuditEntry `json:"last"`
	Warning string            `json:"warning"`
	NextAt  int64             `json:"next_at"`
	// Armed: job 04:30 đã được phép xoá thật chưa. false = mỗi đêm chỉ chạy thử.
	Armed bool `json:"armed"`
}

// handleAdminJanitor: GET — dung lượng + bản xem trước (dry-run) + lần chạy gần nhất.
//
// Dry-run chạy NGAY trong request GET, không phải sau một nút bấm: trang này tồn tại để
// trả lời "sắp xoá những gì", và bắt bấm thêm một nút để biết điều đó chỉ làm người ta
// bấm "chạy thật" mà chưa đọc.
func (s *server) handleAdminJanitor(w http.ResponseWriter, _ *http.Request) {
	if s.janitor == nil {
		writeJSON(w, map[string]any{"off": true, "note": "Bộ dọn chưa bật (server chạy không có telemetry)."})
		return
	}
	now := time.Now()
	tabs, files := s.janitor.Sizes()
	v := janitorView{Tables: tabs, Files: files, Plan: s.janitor.Plan(now),
		Warning: s.janitor.DBWarning(), NextAt: nextJanitorRun(now).Unix(), Armed: s.janitor.Armed()}
	if list, err := s.adminAuth.Store().RecentAudit(200); err == nil {
		for i := range list {
			if list[i].Action == "janitor_run" {
				v.Last = &list[i]
				break
			}
		}
	}
	writeJSON(w, v)
}

// handleAdminJanitorRun: POST — chạy THẬT. Luôn ghi nhật ký kèm IP người bấm.
func (s *server) handleAdminJanitorRun(w http.ResponseWriter, r *http.Request) {
	if s.janitor == nil {
		http.Error(w, "bộ dọn chưa bật", http.StatusServiceUnavailable)
		return
	}
	res := s.janitor.Run(time.Now())
	// Bấm nút này là một người đã đọc danh sách và đồng ý — từ giờ job 04:30 được xoá thật.
	s.janitor.Arm()
	s.adminAuth.Store().Audit("janitor_run", "", "chạy tay từ trang quản trị", clientIP(r), map[string]any{
		"rows": res.Rows, "vacuumed": res.Vacuumed, "freed_bytes": res.FreedBytes, "jobs": res.Jobs,
	})
	log.Printf("dọn dữ liệu (tay): xoá %d dòng · giải phóng %.1f MB", res.Rows, float64(res.FreedBytes)/(1<<20))
	writeJSON(w, res)
}

// startJanitor dựng bộ dọn và chạy lịch. Gọi từ `main` sau khi mọi DB đã mở.
func (s *server) startJanitor(stop <-chan struct{}) {
	dbs := admin.JanitorDBs{
		Main:        s.store.DB(),
		MainPath:    s.dbPath,
		FeedbackDir: feedbackDir(s.dbPath),
	}
	if s.telemetry != nil {
		dbs.Telemetry = s.telemetry.DB()
		dbs.TelemetryPath = telemetryPath("")
	}
	if s.metricsDB != nil {
		dbs.Metrics = s.metricsDB.SQL()
		dbs.MetricsPath = metricsPath("")
	}
	// Số người thật đang chơi = tổng suất phòng đang giữ. Đọc qua `Rooms()` vì nó đã khoá
	// đúng và trả bản chụp — gọi thẳng vào từng phòng là đọc song song với goroutine của nó.
	s.janitor = admin.NewJanitor(s.adminAuth.Store(), dbs, func() int {
		n := 0
		for _, r := range s.mgr.Rooms() {
			n += r.Humans
		}
		return n
	})
	safe.Go("nen", nil, func() { s.runJanitor(stop) })
}

// printJanitorPlan in bản xem trước ra màn hình rồi thôi — không xoá gì.
//
// Dùng cho lần đầu bật bộ dọn trên một DB thật: `go run ./cmd/server -janitor-dryrun`.
func (s *server) printJanitorPlan() {
	now := time.Now()
	plan := s.janitor.Plan(now)
	tabs, files := s.janitor.Sizes()
	fmt.Println("\n=== BỘ DỌN DỮ LIỆU — CHẠY THỬ (không xoá gì) ===")
	fmt.Printf("thời điểm: %s · đã bật xoá thật: %v\n\n", now.Format("2006-01-02 15:04"), s.janitor.Armed())
	fmt.Println("-- dung lượng theo tệp --")
	for _, f := range files {
		fmt.Printf("  %-18s %10s%s\n", f.Name, humanBytes(f.Bytes), map[bool]string{true: "  ⚠ > 1 GB"}[f.Warn])
	}
	fmt.Println("\n-- dung lượng theo bảng (ước lượng) --")
	for _, t := range tabs {
		if t.Rows == 0 {
			continue
		}
		fmt.Printf("  %-24s %9d dòng  %10s\n", t.Name, t.Rows, humanBytes(t.Bytes))
	}
	fmt.Println("\n-- SẼ XOÁ --")
	for _, j := range plan.Jobs {
		mark := " "
		if j.Rows > 0 {
			mark = "*"
		}
		fmt.Printf(" %s %-42s giữ %-22s %8d dòng%s\n", mark, j.What, j.Keep, j.Rows, errSuffix(j.Err))
	}
	fmt.Printf("\nTỔNG: %d dòng. %s\n", plan.Rows, plan.VacuumNote)
}

func errSuffix(e string) string {
	if e == "" {
		return ""
	}
	return "  LỖI: " + e
}

func humanBytes(b int64) string {
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.2f GB", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	default:
		return fmt.Sprintf("%.0f KB", float64(b)/1024)
	}
}
