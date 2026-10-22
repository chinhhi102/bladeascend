// Command server runs the game: HTTP API, WebSocket rooms and the built client.
//
//	go run ./cmd/server              # http://localhost:8080
//	go run ./cmd/server -debug=false # production: disable debug keys
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"kiemxoay/server/internal/config"
	"kiemxoay/server/internal/room"
	"kiemxoay/server/internal/store"
	"kiemxoay/server/internal/world"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	cfgPath := flag.String("config", config.FindRepoFile("config/game.json"), "config file")
	mapsDir := flag.String("maps", config.FindRepoFile("maps"), "Tiled maps directory")
	assetsDir := flag.String("assets", config.FindRepoFile("assets"), "assets directory")
	clientDir := flag.String("client", config.FindRepoFile("client/dist"), "built client directory")
	dbPath := flag.String("db", "", "SQLite file (default: server/data/leaderboard.db)")
	debug := flag.Bool("debug", true, "cho phép phím test; người chơi vẫn phải tự mở bằng ?debug=true")
	flag.Parse()

	cfg, raw, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	maps, err := world.LoadDir(*mapsDir, world.LoadOptions{HP: cfg.Obstacles.HP, FootprintW: cfg.Obstacles.FootprintW, FootprintH: cfg.Obstacles.FootprintH})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("loaded %d maps", len(maps))
	st := openStore(*dbPath, cfg)
	defer st.Close()

	mgr := room.NewManager(cfg, maps, st)
	mgr.AllowDebug = *debug // người chơi vẫn phải tự mở bằng ?debug=true mới thấy bảng
	srv := &server{cfg: cfg, rawCfg: raw, mgr: mgr, store: st, maps: maps, debug: *debug,
		names: room.LoadNameFilter(cfg, filepath.Join(filepath.Dir(*cfgPath), "bad_words.txt")), limiter: newIPLimiter(cfg.Net.MaxConnsPerIP)}
	mux := srv.routes(*assetsDir, *clientDir)
	log.Printf("Kiếm Xoay server listening on %s (debug=%v)", *addr, *debug)
	httpSrv := &http.Server{Addr: *addr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	log.Fatal(httpSrv.ListenAndServe())
}

func openStore(path string, cfg *config.Config) *store.Store {
	if path == "" {
		repo := filepath.Dir(filepath.Dir(config.FindRepoFile("config/game.json")))
		dir := filepath.Join(repo, "server", "data")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatal(err)
		}
		path = filepath.Join(dir, "leaderboard.db")
	}
	st, err := store.Open(path, time.Duration(cfg.Net.LeaderboardCacheSc*float64(time.Second)))
	if err != nil {
		log.Fatal(err)
	}
	return st
}
