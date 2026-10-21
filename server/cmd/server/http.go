package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"kiemxoay/server/internal/config"
	"kiemxoay/server/internal/room"
	"kiemxoay/server/internal/store"
	"kiemxoay/server/internal/world"
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
}

func (s *server) routes(assetsDir, clientDir string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/config", s.handleConfig)
	mux.HandleFunc("GET /api/leaderboard", s.handleLeaderboard)
	mux.HandleFunc("GET /api/rooms", s.handleRooms)
	mux.HandleFunc("GET /api/map/{id}", s.handleMap)
	mux.HandleFunc("GET /ws", s.handleWS)
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", noListing(http.FileServer(http.Dir(assetsDir)))))
	mux.Handle("GET /", spa(clientDir))
	return mux
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

// handleLeaderboard: GET /api/leaderboard?period=day|week|all&metric=exp|kills (spec §11).
func (s *server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	period, metric := r.URL.Query().Get("period"), r.URL.Query().Get("metric")
	if period == "" {
		period = "all"
	}
	if metric == "" {
		metric = "exp"
	}
	rows, err := s.store.Top(period, metric, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]any{"period": period, "metric": metric, "rows": rows})
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
func spa(dir string) http.Handler {
	fs := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := filepath.Join(dir, filepath.Clean("/"+r.URL.Path))
		if st, err := os.Stat(p); err != nil || st.IsDir() {
			if _, err := os.Stat(filepath.Join(dir, "index.html")); err != nil {
				http.Error(w, "Client chưa được build. Chạy: cd client && npm run build (hoặc dùng npm run dev)", http.StatusNotFound)
				return
			}
			http.ServeFile(w, r, filepath.Join(dir, "index.html"))
			return
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

// clientIP honours X-Forwarded-For only from a local reverse proxy (Caddy).
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" && (host == "127.0.0.1" || host == "::1") {
		return strings.TrimSpace(strings.Split(fwd, ",")[0])
	}
	return host
}
