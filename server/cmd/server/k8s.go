package main

// Phần chuẩn bị để chạy trong Kubernetes — DEPLOY_K8S_SPEC §1.
//
// Không đổi luật chơi, không đổi giao thức. Toàn bộ nội dung tệp này là những thứ một bộ
// điều phối (orchestrator) cần ở một tiến trình: biết nó còn sống không, sẵn sàng nhận
// người chưa, đang tốn bao nhiêu, và làm sao bảo nó tắt mà không mất dữ liệu.
//
// Chạy ngoài Kubernetes thì mọi thứ ở đây vẫn đúng: `/healthz` luôn trả 200, `/readyz` đọc
// đúng trạng thái thật, `/metrics` vẫn có số. Không có nhánh "chỉ chạy trên k8s" — một
// đường chạy chỉ sống trong production là một đường chưa ai thử.

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"bladeascend/server/internal/metrics"
)

// ── Cấu hình qua biến môi trường (§1.1) ─────────────────────────────────────────────────

// envOr trả giá trị biến môi trường nếu có, ngược lại trả `def`.
//
// Dùng làm MẶC ĐỊNH cho cờ dòng lệnh chứ không đè lên cờ: thứ tự ưu tiên phải là
// cờ > biến môi trường > mặc định. Ngược lại thì một biến môi trường sót lại trong shell sẽ
// âm thầm vô hiệu hoá cờ người ta vừa gõ — loại lỗi mất hàng giờ mới tìm ra.
func envOr(name, def string) string {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		return v
	}
	return def
}

// envDur đọc một khoảng thời gian; giá trị hỏng thì cảnh báo và dùng mặc định.
func envDur(name string, def time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(name))
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		log.Printf("⚠ %s=%q không đọc được (%v) — dùng %v", name, v, err, def)
		return def
	}
	return d
}

// ── GOMEMLIMIT (§1.7) ───────────────────────────────────────────────────────────────────

// applyMemLimit đặt GOMEMLIMIT bằng 70% hạn mức bộ nhớ của container.
//
// Vì sao cần: Go mặc định chỉ gom rác theo TỈ LỆ tăng heap, nó không biết gì về hạn mức
// cgroup. Chạm hạn mức thì nhân hệ điều hành giết tiến trình bằng OOM — không log, không
// panic, không một dòng nào trong trang "Lỗi". Đặt GOMEMLIMIT thì Go gom rác gắt dần khi
// tới gần, và tiến trình chậm lại thay vì biến mất.
//
// 70% vì heap không phải toàn bộ bộ nhớ tiến trình: stack goroutine, bộ đệm SQLite, mmap
// của runtime nằm ngoài. Chừa 30% cho phần đó.
//
// Đọc theo thứ tự: biến môi trường `GOMEMLIMIT` (ai đã đặt tay thì tôn trọng) → `MEM_LIMIT_MB`
// (ConfigMap/Downward API) → hạn mức cgroup v2 → không làm gì.
func applyMemLimit() {
	if os.Getenv("GOMEMLIMIT") != "" {
		return // Go runtime tự đọc biến này rồi
	}
	var limitBytes int64
	if v := strings.TrimSpace(os.Getenv("MEM_LIMIT_MB")); v != "" {
		if mb, err := strconv.ParseInt(v, 10, 64); err == nil && mb > 0 {
			limitBytes = mb * 1024 * 1024
		}
	}
	if limitBytes == 0 {
		limitBytes = cgroupMemLimit()
	}
	if limitBytes <= 0 {
		return
	}
	lim := limitBytes * 70 / 100
	debug.SetMemoryLimit(lim)
	log.Printf("GOMEMLIMIT = %d MB (70%% của %d MB)", lim/(1024*1024), limitBytes/(1024*1024))
}

// cgroupMemLimit đọc hạn mức bộ nhớ cgroup v2; 0 nếu không có hoặc không giới hạn.
func cgroupMemLimit() int64 {
	b, err := os.ReadFile("/sys/fs/cgroup/memory.max")
	if err != nil {
		return 0
	}
	s := strings.TrimSpace(string(b))
	if s == "max" {
		return 0
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// ── Log JSON ra stdout (§1.4) ───────────────────────────────────────────────────────────

// jsonLogWriter bọc `log` để mỗi dòng thành một object JSON.
//
// Giữ nguyên mọi lời gọi `log.Printf` đang có trong repo — đổi hết sang một thư viện log có
// cấu trúc là một lượt sửa hàng trăm chỗ, và mỗi chỗ là một cơ hội đổi nhầm chữ. Ở đây chỉ
// đổi CÁCH IN: nội dung dòng vào trường `msg`, kèm thời gian và build.
//
// Bật bằng `LOG_FORMAT=json`. Mặc định vẫn là dòng chữ thường, vì người chạy `go run` trên
// máy mình đang đọc log bằng mắt chứ không bằng Loki.
type jsonLogWriter struct {
	build string
	mu    sync.Mutex
	out   *os.File
}

func (w *jsonLogWriter) Write(p []byte) (int, error) {
	msg := strings.TrimRight(string(p), "\n")
	rec := map[string]any{
		"ts":    time.Now().UTC().Format(time.RFC3339Nano),
		"level": levelOf(msg),
		"msg":   msg,
		"build": w.build,
	}
	b, err := json.Marshal(rec)
	if err != nil {
		return len(p), nil // không bao giờ để việc ghi log làm hỏng việc đang chạy
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	_, _ = w.out.Write(append(b, '\n'))
	return len(p), nil
}

// levelOf đoán mức log từ nội dung dòng.
//
// Đoán chứ không đòi: mọi `log.Printf` trong repo viết bằng tiếng Việt tự nhiên, không ai
// gắn mức. Quy ước ở đây chỉ dựa vào dấu hiệu đã dùng sẵn trong repo (⚠ và "LỖI").
func levelOf(msg string) string {
	switch {
	case strings.Contains(msg, "LỖI") || strings.Contains(msg, "panic"):
		return "error"
	case strings.Contains(msg, "⚠") || strings.Contains(msg, "cảnh báo"):
		return "warn"
	default:
		return "info"
	}
}

// setupLogging đặt định dạng log. Gọi một lần, sớm nhất có thể.
func setupLogging(build string) {
	if !strings.EqualFold(envOr("LOG_FORMAT", ""), "json") {
		return
	}
	log.SetFlags(0) // thời gian đã nằm trong trường `ts`
	log.SetOutput(&jsonLogWriter{build: build, out: os.Stdout})
}

// ── /healthz và /readyz (§1.2) ──────────────────────────────────────────────────────────

// health giữ trạng thái vòng đời của tiến trình.
type health struct {
	mu       sync.RWMutex
	draining bool
	reason   string
}

func (h *health) startDrain(reason string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.draining = true
	h.reason = reason
}

func (h *health) isDraining() (bool, string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.draining, h.reason
}

// handleHealthz — CÒN SỐNG không. Luôn 200 khi tiến trình còn trả lời được.
//
// Cố tình KHÔNG kiểm DB hay phòng: liveness mà kiểm phụ thuộc bên ngoài thì một lần SQLite
// bận sẽ làm Kubernetes GIẾT pod, và giết đúng lúc nó đang chật vật là cách chắc chắn nhất
// để biến một sự cố nhỏ thành một vòng khởi động lại không dứt. Thứ đó thuộc về `/readyz`.
func (s *server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok\n"))
}

// handleReadyz — SẴN SÀNG nhận người chơi chưa.
//
// 503 khi: đang tắt êm (§1.3), hoặc DB không mở được, hoặc không có phòng nào tick được.
// Trả 503 là bảo Ingress đừng gửi người mới tới; tiến trình vẫn sống và vẫn phục vụ người
// đang chơi.
func (s *server) handleReadyz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if drain, reason := s.health.isDraining(); drain {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"ready": false, "reason": reason})
		return
	}
	if err := s.store.DB().Ping(); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"ready": false, "reason": "db: " + err.Error()})
		return
	}
	// Phòng phải TICK được, không chỉ tồn tại: `Query` chạy trên chính goroutine phòng giữa
	// hai tick, nên nó trả false đúng khi phòng treo — và phòng treo là lúc người chơi thấy
	// game đứng hình, dù tiến trình vẫn trả lời HTTP ngon lành.
	rooms := s.mgr.AllRooms()
	if len(rooms) > 0 {
		alive := false
		for _, rm := range rooms {
			if rm.Query(func() {}) {
				alive = true
				break
			}
		}
		if !alive {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]any{"ready": false, "reason": "không phòng nào tick được"})
			return
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"ready": true, "build": s.buildID})
}

// ── /metrics định dạng Prometheus (§1.5) ────────────────────────────────────────────────

// promStore giữ mẫu MỚI NHẤT của mỗi (metric, scope), do bộ thu gom §12 soi gương sang.
//
// Chỉ giữ mẫu mới nhất chứ không gom: Prometheus tự lo phần lịch sử, và nó hỏi mỗi 15–30 s
// trong khi bộ thu gom lấy mẫu mỗi giây. Gom thêm một lần nữa ở đây là làm lại việc
// Prometheus sinh ra để làm.
type promStore struct {
	mu   sync.Mutex
	vals map[string]map[string]promVal
}

type promVal struct {
	avg, p95, max float64
	count         int64
	at            time.Time
}

func newPromStore() *promStore {
	return &promStore{vals: map[string]map[string]promVal{}}
}

func (p *promStore) note(metric, scope string, d metrics.Dist) {
	p.mu.Lock()
	defer p.mu.Unlock()
	m := p.vals[metric]
	if m == nil {
		m = map[string]promVal{}
		p.vals[metric] = m
	}
	m[scope] = promVal{avg: d.Avg(), p95: d.P95(), max: d.Max, count: d.Count, at: time.Now()}
}

// STALE_AFTER: mẫu cũ hơn ngần này thì thôi xuất.
//
// Một khu vừa đóng mà vẫn xuất số cũ mãi thì bảng Grafana sẽ vẽ một đường thẳng tắp, và
// đường thẳng tắp đọc ra là "ổn định" chứ không phải "không còn ai đo".
const promStaleAfter = 90 * time.Second

func (p *promStore) write(w *strings.Builder) {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	names := make([]string, 0, len(p.vals))
	for n := range p.vals {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		scopes := p.vals[name]
		keys := make([]string, 0, len(scopes))
		for k := range scopes {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		wrote := false
		for _, scope := range keys {
			v := scopes[scope]
			if now.Sub(v.at) > promStaleAfter {
				continue
			}
			if !wrote {
				fmt.Fprintf(w, "# TYPE bladeascend_%s gauge\n", name)
				wrote = true
			}
			lbl := promLabels(scope)
			fmt.Fprintf(w, "bladeascend_%s{%s} %g\n", name, lbl, v.avg)
			// Phân phối (tick…) mới có p95/max đáng xuất; mẫu đơn thì p95 = max = chính nó.
			if v.count > 1 {
				fmt.Fprintf(w, "bladeascend_%s_p95{%s} %g\n", name, lbl, v.p95)
				fmt.Fprintf(w, "bladeascend_%s_max{%s} %g\n", name, lbl, v.max)
			}
		}
	}
}

// promLabels đổi `scope` của bộ thu gom thành nhãn Prometheus.
//
// Bộ thu gom §12 dùng chuỗi `"zone:3"` hoặc `"server"`. Giữ nguyên chuỗi đó làm nhãn thì
// Grafana không lọc theo khu được, nên tách ra.
func promLabels(scope string) string {
	if rest, ok := strings.CutPrefix(scope, "zone:"); ok {
		return `scope="zone",zone="` + rest + `"`
	}
	return `scope="` + scope + `"`
}

// handleMetrics xuất số liệu cho Prometheus.
//
// Cổng NỘI BỘ, không công khai (NetworkPolicy chỉ cho Prometheus vào — §3). Ở đây vẫn chặn
// thêm một lớp: không mở ra ngoài khi `METRICS_PUBLIC` chưa bật, vì một ngày nào đó sẽ có
// người cắm Service kiểu LoadBalancer vào cổng này.
func (s *server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if !strings.EqualFold(envOr("METRICS_PUBLIC", ""), "true") && !isLocalRequest(r) {
		http.Error(w, "chỉ cho mạng nội bộ", http.StatusForbidden)
		return
	}
	var b strings.Builder
	// Vài số không đi qua bộ thu gom §12 vì chúng là của chính tiến trình, không của trận.
	fmt.Fprintf(&b, "# TYPE bladeascend_up gauge\nbladeascend_up{build=%q} 1\n", s.buildID)
	fmt.Fprintf(&b, "# TYPE bladeascend_uptime_seconds gauge\nbladeascend_uptime_seconds %g\n",
		time.Since(s.startedAt).Seconds())
	fmt.Fprintf(&b, "# TYPE bladeascend_goroutines gauge\nbladeascend_goroutines %d\n", runtime.NumGoroutine())
	drain, _ := s.health.isDraining()
	fmt.Fprintf(&b, "# TYPE bladeascend_draining gauge\nbladeascend_draining %d\n", boolToInt(drain))
	s.prom.write(&b)
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	_, _ = w.Write([]byte(b.String()))
}

// isLocalRequest: yêu cầu đến từ loopback hoặc mạng riêng (RFC1918 / CGNAT / link-local).
//
// Trong Kubernetes, Prometheus nằm ở một pod khác nên IP nguồn KHÔNG phải 127.0.0.1 — khoá
// theo loopback như `/debug/pprof` làm thì `/metrics` không bao giờ cạo được. Lớp chặn thật
// là `NetworkPolicy` (§3); chỗ này chỉ là lớp thứ hai cho ngày ai đó cắm nhầm một Service
// kiểu LoadBalancer vào cổng metrics.
func isLocalRequest(r *http.Request) bool {
	ip := net.ParseIP(clientIP(r))
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
