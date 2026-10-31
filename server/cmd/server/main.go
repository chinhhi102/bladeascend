// Command server runs the game: HTTP API, WebSocket rooms and the built client.
//
//	go run ./cmd/server              # http://localhost:9999
//	go run ./cmd/server -debug=false # production: disable debug keys
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"syscall"
	"time"

	"bladeascend/server/internal/admin"
	"bladeascend/server/internal/config"
	"bladeascend/server/internal/game"
	"bladeascend/server/internal/room"
	"bladeascend/server/internal/store"
	"bladeascend/server/internal/world"
)

func main() {
	// `:9999` nghe trên MỌI card mạng, nên máy khác cùng Wi-Fi vào được bằng
	// `http://<IP-máy-này>:9999`. Muốn đóng lại, chỉ cho máy này chơi:
	// `go run ./cmd/server -addr 127.0.0.1:9999`.
	addr := flag.String("addr", ":9999", "địa chỉ nghe; \":9999\" = mở ra mạng LAN, \"127.0.0.1:9999\" = chỉ máy này")
	cfgPath := flag.String("config", config.FindRepoFile("config/game.json"), "config file")
	mapsDir := flag.String("maps", config.FindRepoFile("maps"), "Tiled maps directory")
	// Phục vụ `assets/dist` — bản đã nén của `tools/asset-pipeline` (§13.7). Trỏ vào
	// `assets/src` là gửi nguyên ảnh gốc 1024–2048 px cho người chơi: đúng cái lỗi §13.7
	// sinh ra để sửa. Chưa chạy pipeline thì `routes` tự rơi về `assets/src` và cảnh báo.
	assetsDir := flag.String("assets", config.FindRepoFile("assets/dist"), "assets directory (bản đã nén của asset-pipeline)")
	clientDir := flag.String("client", config.FindRepoFile("client/dist"), "built client directory")
	dbPath := flag.String("db", "", "SQLite file (default: server/data/leaderboard.db)")
	// Từ ADMIN_SPEC §1.1, cờ này KHÔNG còn là thứ quyết định ai được debug — `admin.DebugGate`
	// quyết, và nó chỉ mở cho phiên đăng nhập quản trị (hoặc dev + localhost + DEBUG_TOKEN).
	// Cờ này chỉ còn là công tắc tổng để tắt hẳn. `?debug=true` từ client là một lời xin,
	// server không nghe. Mô tả cũ ("người chơi tự mở bằng ?debug=true") đã sai và nó từng
	// là nguyên nhân thật: xem docs/LEADERBOARD_AUDIT.md §1.
	debug := flag.Bool("debug", true, "công tắc tổng cho lệnh test; ai được dùng thì do phiên admin quyết (§1.1)")
	// Trần kết nối mỗi IP là để chống lạm dụng (§13.5). Chạy tools/browsertest thì 10 tab
	// cùng đến từ 127.0.0.1 nên phải nới, nếu không tab thứ 9 trở đi ăn HTTP 429.
	maxConns := flag.Int("max-conns-per-ip", 0, "ghi đè net.max_conns_per_ip (0 = theo config)")
	// Hồ sơ CPU cho công cụ đo: cmd/loadtest chạy trong ống nghiệm, còn đây là server thật
	// với bot, bản đồ thật và client thật — hai nơi cho ra hình ảnh CPU khác hẳn nhau.
	cpuProf := flag.String("cpuprofile", "", "ghi hồ sơ CPU ra file này cho tới khi tắt server")
	// Telemetry người chơi thật (TELEMETRY_SPEC). Bật mặc định vì `:9999` là server của đội
	// test; bản release chạy với `-telemetry=false` và client release không đóng gói module.
	telemetryOn := flag.Bool("telemetry", true, "thu số liệu hiệu năng của người chơi thật")
	telemetryDB := flag.String("telemetry-db", "", "SQLite telemetry (mặc định: server/data/telemetry.db)")
	buildID := flag.String("build", "", "mã build ghi kèm mọi số liệu (mặc định: thời điểm khởi động)")
	flag.Parse()

	if *cpuProf != "" {
		f, err := os.Create(*cpuProf)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}

	cfg, raw, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	maps, err := world.LoadDir(*mapsDir, world.LoadOptions{Tier: cfg.Obstacles.ObstacleTier, FootprintW: cfg.Obstacles.FootprintW, FootprintH: cfg.Obstacles.FootprintH})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("loaded %d maps", len(maps))
	st := openStore(*dbPath, cfg)
	defer st.Close()

	if *maxConns > 0 {
		cfg.Net.MaxConnsPerIP = *maxConns
	}
	mgr := room.NewManager(cfg, maps, st)
	mgr.AllowDebug = *debug // công tắc tổng; ai được dùng thì `admin.DebugGate` quyết (§1.1)
	// Mốc trung vị cho cờ "Tăng điểm bất thường" (ADMIN_SPEC §6.1, QUESTIONS Q29).
	//
	// Cắm bằng hàm vì gói `room` không được chạm SQLite giữa tick: `ScoreMedianPerMin` giữ
	// cache 5 phút, nên lời gọi từ vòng tick gần như luôn chỉ đọc một map trong bộ nhớ.
	prog := game.NewProgression(cfg.Cultivation, cfg.Sword.MaxSwords)
	mgr.ScoreMedian = func(realm int) float64 {
		v, _ := st.ScoreMedianPerMin(realm, prog.Realm, time.Now())
		return v
	}
	tel := startTelemetry(telemetryPath(*telemetryDB), buildIDOrNow(*buildID), *telemetryOn)
	// Trang quản trị (ADMIN_SPEC §1). Dùng CHUNG file SQLite với bảng xếp hạng, như spec §8
	// nói — tên bảng đều có tiền tố `admin_` nên không đụng dữ liệu đã có.
	dbFile := storePath(*dbPath)
	adminStore, err := admin.New(st.DB())
	if err != nil {
		log.Fatal(err)
	}
	adminAuth, err := admin.NewAuth(adminStore)
	if err != nil {
		log.Fatal(err) // hash sai định dạng hoặc cost quá thấp: dừng hẳn, không chạy nửa vời
	}
	adminStore.PurgeExpired()
	log.Print(adminEnabledNote(adminAuth))

	srv := &server{cfg: cfg, rawCfg: raw, mgr: mgr, store: st, maps: maps, debug: *debug,
		adminAuth: adminAuth, debugGate: admin.NewDebugGate(adminAuth),
		buildID: buildIDOrNow(*buildID), startedAt: time.Now(), dbPath: dbFile, mapsDir: *mapsDir,
		names: room.LoadNameFilter(cfg, filepath.Join(filepath.Dir(*cfgPath), "bad_words.txt")), limiter: newIPLimiter(cfg.Net.MaxConnsPerIP),
		telemetry: tel}
	stopSampler := make(chan struct{})
	// Dọn dẹp hằng ngày (Q28): xoá ảnh góp ý của thư đã khép quá 90 ngày, giữ chữ mãi.
	// Chạy kể cả khi tắt telemetry — nó dọn DB quản trị, không liên quan telemetry.
	go srv.runHousekeeper(stopSampler)
	if tel != nil {
		defer tel.Close()
		go srv.runServerSampler(stopSampler)
		go srv.runRollupScheduler(stopSampler)
	}
	// Chưa chạy pipeline thì `assets/dist` rỗng và game vào trận với placeholder trắng
	// trơn. Chuyện đó phải nói ngay ở dòng log đầu tiên: lần trước thiếu hai tệp font,
	// server vẫn trả HTTP 200 cho mọi thứ, và chủ dự án vào lại 8 lần mới báo.
	if _, err := os.Stat(filepath.Join(*assetsDir, "manifest.json")); err != nil {
		log.Printf("⚠ CHƯA CÓ ASSET: %s/manifest.json không tồn tại.\n"+
			"  Chạy:  node tools/asset-pipeline/index.mjs\n"+
			"  (hoặc `cd client && npm run build` — nó tự gọi pipeline trước.)", *assetsDir)
	}
	mux := srv.routes(*assetsDir, *clientDir)
	log.Printf("%s server listening on %s (debug=%v)", cfg.Brand.Name, *addr, *debug)
	httpSrv := &http.Server{Addr: *addr, Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	// Tắt êm khi nhận SIGINT/SIGTERM: cần thế thì hồ sơ CPU mới được ghi xuống đĩa. Bị
	// SIGKILL thì file hồ sơ rỗng.
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch
		if *cpuProf != "" {
			pprof.StopCPUProfile()
			log.Printf("đã ghi hồ sơ CPU vào %s", *cpuProf)
		}
		close(stopSampler)
		if tel != nil {
			// Đóng telemetry TRƯỚC khi tắt HTTP: Close ghi nốt hàng đợi xuống đĩa, nếu không
			// thì 5 giây số liệu cuối cùng — thường là 5 giây thú vị nhất — mất sạch.
			if err := tel.Close(); err != nil {
				log.Printf("telemetry: đóng: %v", err)
			}
		}
		_ = httpSrv.Close()
	}()
	if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

// telemetryPath mặc định đặt DB cạnh leaderboard, nhưng là FILE RIÊNG (§3): telemetry ghi
// liên tục và bị dọn theo hạn 30 ngày, không được đụng vào dữ liệu bảng xếp hạng.
func telemetryPath(path string) string {
	if path != "" {
		return path
	}
	repo := filepath.Dir(filepath.Dir(config.FindRepoFile("config/game.json")))
	dir := filepath.Join(repo, "server", "data")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatal(err)
	}
	return filepath.Join(dir, "telemetry.db")
}

// buildIDOrNow: mã build ghi kèm mọi dòng số liệu (§5), để so "trước/sau một build" được.
// Không truyền `-build` thì lấy thời điểm khởi động — vẫn phân biệt được hai lần deploy.
func buildIDOrNow(id string) string {
	if id != "" {
		return id
	}
	return time.Now().Format("2006-01-02_1504")
}

// storePath suy ra đường dẫn file SQLite thật từ cờ `-db` (rỗng = mặc định).
//
// Tách khỏi `openStore` vì hai nơi cần nó: mở DB, và biết chỗ đặt thư mục ảnh góp ý
// (`data/feedback/`, §8 — ngoài thư mục phục vụ công khai).
func storePath(path string) string {
	if path != "" {
		return path
	}
	repo := filepath.Dir(filepath.Dir(config.FindRepoFile("config/game.json")))
	dir := filepath.Join(repo, "server", "data")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatal(err)
	}
	return filepath.Join(dir, "leaderboard.db")
}

func openStore(path string, cfg *config.Config) *store.Store {
	path = storePath(path)
	st, err := store.Open(path, time.Duration(cfg.Net.LeaderboardCacheSc*float64(time.Second)))
	if err != nil {
		log.Fatal(err)
	}
	return st
}
