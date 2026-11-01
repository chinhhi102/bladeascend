package main

import (
	"compress/gzip"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"bladeascend/server/internal/admin"
	"bladeascend/server/internal/config"
	"bladeascend/server/internal/room"
	"bladeascend/server/internal/store"
	"bladeascend/server/internal/telemetry"
	"bladeascend/server/internal/world"
)

type server struct {
	cfg     *config.Config
	rawCfg  []byte
	mgr     *room.Manager
	store   *store.Store
	maps    []*world.Map
	debug   bool
	names   *room.NameFilter
	limiter *ipLimiter
	// telemetry là nil khi chạy với -telemetry=false; mọi chỗ dùng phải kiểm nil.
	telemetry *telemetry.Store
	// adminAuth không bao giờ nil, nhưng `Enabled()` là false khi chưa đặt
	// ADMIN_PASSWORD_HASH — lúc đó mọi đường dẫn /admin đều đóng (ADMIN_SPEC §1).
	adminAuth *admin.Auth
	// debugGate quyết định kết nối nào được dùng lệnh test (ADMIN_SPEC §1.1).
	debugGate *admin.DebugGate
	buildID   string
	startedAt time.Time
	// dbPath: đường dẫn file SQLite, để biết chỗ đặt thư mục ảnh góp ý (§8).
	dbPath string
	// mapsDir: thư mục bản đồ, để đọc `check.json` cho trang xem bố cục (§5.2).
	mapsDir string
	// env / envDefaults: môi trường đang chạy (APP_ENV) và mặc định phòng của nó (§9.5).
	env         string
	envDefaults config.RoomEnv
}

func (s *server) routes(assetsDir, clientDir string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/config", s.handleConfig)
	mux.HandleFunc("GET /api/leaderboard", s.handleLeaderboard)
	mux.HandleFunc("GET /api/rooms", s.handleRooms)
	mux.HandleFunc("GET /api/map/{id}", s.handleMap)
	mux.HandleFunc("GET /ws", s.handleWS)
	// Tóm tắt phiên lúc đóng tab: sendBeacon chỉ gửi được POST, và lúc đó WebSocket đã đứt.
	mux.HandleFunc("POST /telemetry/end", s.handleTelemetryEnd)
	mux.HandleFunc("GET /admin/perf", s.handleAdminPerf)
	s.adminRoutes(mux, adminDirOf(clientDir))
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", noListing(assetCache(http.FileServer(http.Dir(assetsDir))))))
	mux.Handle("GET /", spa(clientDir))
	return gzipMux(mux)
}

// hashedName nhận ra tên tệp do `tools/asset-pipeline` sinh ra: `<tên>.<8 chữ hex>.<đuôi>`.
//
// Chỉ những tệp này được cache một năm. Phân biệt bằng CHÍNH TÊN chứ không bằng thư mục, vì
// `manifest.json` nằm cùng chỗ với chúng mà lại là tệp duy nhất KHÔNG được cache lâu: nó là
// thứ trỏ tới tất cả những tệp kia, nên cache nó là người chơi mắc kẹt với bộ art cũ cho tới
// khi tự xoá cache — đúng kiểu lỗi không ai báo được vì "máy em vẫn thế".
var hashedName = regexp.MustCompile(`\.[0-9a-f]{8}\.[a-z0-9]+$`)

func assetCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hashedName.MatchString(r.URL.Path) {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			// manifest.json và mọi thứ chưa có hash: phải hỏi lại server mỗi lần.
			w.Header().Set("Cache-Control", "no-cache")
		}
		h.ServeHTTP(w, r)
	})
}

// gzipCompressible: chỉ nén thứ nén được. WebP/OGG/PNG đã nén rồi — gzip thêm một lượt là
// tốn CPU để tệp TO HƠN vài byte.
var gzipCompressible = map[string]bool{
	".json": true, ".js": true, ".mjs": true, ".css": true, ".html": true,
	".ttf": true, ".svg": true, ".txt": true, ".md": true, ".map": true,
}

type gzipWriter struct {
	http.ResponseWriter
	gz *gzip.Writer
}

func (g gzipWriter) Write(b []byte) (int, error) { return g.gz.Write(b) }

// WriteHeader xoá Content-Length lần cuối trước khi gửi đầu phản hồi.
//
// `http.ServeFile` đã tự bỏ Content-Length khi thấy có Content-Encoding, nhưng không có gì
// bảo đảm mọi handler sau này cũng vậy — và nếu một handler đặt lại nó thì trình duyệt cắt
// ngang tệp giữa đường. Chốt ở đây là chốt một lần cho mọi đường.
func (g gzipWriter) WriteHeader(code int) {
	g.Header().Del("Content-Length")
	g.ResponseWriter.WriteHeader(code)
}

// gzipMux nén gzip cho JS/JSON/CSS/font (GAME_SPEC §13.7 mục 2).
//
// Nén tại thời điểm gửi chứ không nén sẵn thành `.gz` vì nó phải phủ CẢ `client/dist`
// (vite sinh ra) lẫn `assets/dist` (pipeline sinh ra) — một chỗ làm là đủ cho hai nguồn.
// Phần nặng nhất của game là WebP và nó không đi qua nhánh nén này.
func gzipMux(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") || !gzipCompressible[strings.ToLower(path.Ext(r.URL.Path))] {
			h.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding")
		// Content-Length của bản CHƯA nén sẽ sai sau khi nén; http.ServeFile đặt nó, nên
		// phải xoá. Để lại thì trình duyệt cắt ngang tệp và báo lỗi giải mã.
		w.Header().Del("Content-Length")
		gz := gzip.NewWriter(w)
		defer func() { _ = gz.Close() }()
		h.ServeHTTP(gzipWriter{ResponseWriter: w, gz: gz}, r)
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}

func (s *server) handleConfig(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write(s.rawCfg)
}

// handleLeaderboard: GET /api/leaderboard?period=day|week|all&board=score|hits|speed (spec §11.2).
func (s *server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	period, board := r.URL.Query().Get("period"), r.URL.Query().Get("board")
	if period == "" {
		period = "all"
	}
	if board == "" {
		board = "score" // Uy Danh là bảng chính
	}
	rows, err := s.store.Top(period, board, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]any{"period": period, "board": board, "rows": rows})
}

func (s *server) handleRooms(w http.ResponseWriter, _ *http.Request) { writeJSON(w, s.mgr.Rooms()) }

// handleMap serves the parsed map so the client does not need its own Tiled parser.
func (s *server) handleMap(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	for _, m := range s.maps {
		if m.ID == id {
			w.Header().Set("Cache-Control", "public, max-age=60")
			writeJSON(w, m)
			return
		}
	}
	if id == "empty" {
		writeJSON(w, world.EmptyMap(s.cfg.Room.MapSize))
		return
	}
	http.NotFound(w, r)
}

func noListing(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") || strings.HasPrefix(r.URL.Path, "_raw") {
			http.NotFound(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// spa serves the built client and falls back to index.html.
//
// index.html is never cached: its <script> points at a content-hashed bundle, so a plain F5
// after a new build always picks the new game up. The hashed files under /static/ carry a
// new name whenever they change, so they are safe to cache for a long time.
func spa(dir string) http.Handler {
	fs := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := filepath.Join(dir, filepath.Clean("/"+r.URL.Path))
		if st, err := os.Stat(p); err != nil || st.IsDir() {
			if _, err := os.Stat(filepath.Join(dir, "index.html")); err != nil {
				http.Error(w, "Client chưa được build. Chạy: cd client && npm run build (hoặc dùng npm run dev)", http.StatusNotFound)
				return
			}
			w.Header().Set("Cache-Control", "no-store, must-revalidate")
			http.ServeFile(w, r, filepath.Join(dir, "index.html"))
			return
		}
		if strings.HasPrefix(r.URL.Path, "/static/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		// Go không biết đuôi `.webmanifest`, nên `http.FileServer` đoán ra `text/plain` và
		// Chrome bỏ qua manifest — không báo lỗi, chỉ lặng lẽ không cho "Thêm vào màn hình
		// chính". `mime.AddExtensionType` ở `init()` là cách đúng nhưng nó sửa biến toàn
		// cục của tiến trình; một dòng ở đây chỉ ảnh hưởng đường phục vụ tệp tĩnh.
		if strings.HasSuffix(r.URL.Path, ".webmanifest") {
			w.Header().Set("Content-Type", "application/manifest+json; charset=utf-8")
		}
		fs.ServeHTTP(w, r)
	})
}

// ipLimiter caps simultaneous WebSocket connections per IP (spec §13.5).
type ipLimiter struct {
	mu    sync.Mutex
	max   int
	count map[string]int
}

func newIPLimiter(max int) *ipLimiter { return &ipLimiter{max: max, count: map[string]int{}} }

func (l *ipLimiter) acquire(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.count[ip] >= l.max {
		return false
	}
	l.count[ip]++
	return true
}

func (l *ipLimiter) release(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.count[ip]--; l.count[ip] <= 0 {
		delete(l.count, ip)
	}
}

// clientIP honours X-Forwarded-For only from a reverse proxy on this host or its docker
// network (Caddy); the game port is never published publicly.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	isProxy := ip != nil && (ip.IsLoopback() || ip.IsPrivate())
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" && isProxy {
		return strings.TrimSpace(strings.Split(fwd, ",")[0])
	}
	return host
}
