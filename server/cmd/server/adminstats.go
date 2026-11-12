package main

import (
	"net/http"
	"runtime"
	"time"

	"bladeascend/server/internal/room"
)

// Dashboard live — ADMIN_SPEC §3 và §4.
//
// Mọi con số đọc được TỪ TRẠNG THÁI ĐANG CHẠY, không từ một bảng thống kê riêng: thống kê
// riêng là một nguồn sự thật thứ hai, và nó sai lệch dần đúng vào lúc cần tin nó nhất.
//
// Chi phí phải gần bằng không với người chơi — §11 đòi "admin xem không làm tăng tick ms
// người chơi quá 0,3 ms". Nên ở đây chỉ ĐỌC, không khoá gì lâu, không cấp phát theo số
// entity, và gọi 2 giây một lần chứ không mỗi tick.

// zoneStat là một dòng của bảng Phòng & Khu (§4).
type zoneStat struct {
	Index    int     `json:"index"`
	Name     string  `json:"name"`
	Map      string  `json:"map"`
	Humans   int     `json:"humans"`
	Bots     int     `json:"bots"`
	TickAvg  float64 `json:"tick_avg"`
	TickP95  float64 `json:"tick_p95"`
	TopLevel int     `json:"top_level"`
	TopRealm string  `json:"top_realm"`
	Monsters int     `json:"monsters"`
	Items    int     `json:"items"`
	// BrokenObstacles: vật cản đang bị chém, chờ mọc lại (§4).
	BrokenObstacles int `json:"broken_obstacles"`
	// HotSecs: tick p95 > admin.tick_warn_ms liên tục bao lâu (giây). Dashboard cảnh báo khi
	// ≥ admin.tick_warn_secs.
	HotSecs float64 `json:"hot_secs"`
	// Degraded / Panics: khu đã bắt được panic (GAME_SPEC §13.8). `degraded` bật khi
	// panic ≥ 3 lần/phút, và tự tắt sau 5 phút im.
	Degraded bool `json:"degraded,omitempty"`
	Panics   int  `json:"panics,omitempty"`
}

type roomStat struct {
	ID     string     `json:"id"`
	Humans int        `json:"humans"`
	Bots   int        `json:"bots"`
	Zones  []zoneStat `json:"zones"`
}

// adminStats là toàn bộ gói số liệu gửi cho dashboard.
type adminStats struct {
	At      int64  `json:"at"`
	Build   string `json:"build"`
	Uptime  int64  `json:"uptime_secs"`
	Players struct {
		Humans int `json:"humans"`
		Bots   int `json:"bots"`
		Rooms  int `json:"rooms"`
	} `json:"players"`
	Server struct {
		Goroutines int     `json:"goroutines"`
		HeapMB     float64 `json:"heap_mb"`
		TickAvgMax float64 `json:"tick_avg_max"`
		TickP95Max float64 `json:"tick_p95_max"`
		// Ngưỡng cảnh báo vàng (config `admin`) — dashboard so `hot_secs` từng khu với số này.
		TickWarnMs   float64 `json:"tick_warn_ms"`
		TickWarnSecs float64 `json:"tick_warn_secs"`
	} `json:"server"`
	// Realms: phân bố cảnh giới của NGƯỜI THẬT (§3 nhóm "Trò chơi"). Chỉ số cảnh giới →
	// số người. Bot không tính: dashboard là để nhìn người chơi thật.
	Realms map[int]int `json:"realms"`
	TienN  int         `json:"tien_count"`
	Rooms  []roomStat  `json:"rooms"`
	// FeedbackNew: số góp ý chưa đọc, hiện ở thanh bên và dashboard (§3, §7.2).
	FeedbackNew int `json:"feedback_new"`
	// Flagged: số người thật đang mang ít nhất một cờ nghi hack (§6.1). Chỉ là con số ở
	// thanh bên — không có gì chạy theo nó.
	Flagged int `json:"flagged"`
	// ErrorsToday: số LOẠI lỗi khác nhau người chơi gặp trong 24 giờ (§3).
	//
	// Đếm loại chứ không đếm lần: một lỗi trong vòng vẽ bắn 60 lần/giây, và con số "18 432"
	// ở thanh bên chỉ nói lên rằng có ai đó đang mở tab, không nói có mấy vấn đề.
	ErrorsToday int `json:"errors_today"`
	// DBWarning: câu cảnh báo khi tệp dữ liệu vượt 1 GB (ADMIN_SPEC §13). Rỗng = ổn.
	//
	// Nằm ở dashboard chứ không chỉ ở trang "Dọn dữ liệu": trang đó là nơi người ta vào
	// KHI ĐÃ BIẾT có vấn đề, còn đây là nơi duy nhất họ mở mỗi ngày.
	DBWarning string `json:"db_warning,omitempty"`
	// Degraded: khu đang "suy giảm" vì panic ≥ 3 lần/phút (GAME_SPEC §13.8).
	Degraded []string `json:"degraded,omitempty"`
	// Leak: kết quả bài kiểm rò rỉ hằng đêm (§13.9 v6.54). nil = chưa chạy lần nào.
	Leak *LeakNightly `json:"leak,omitempty"`
	// LeakWarning: câu cảnh báo dashboard khi kết quả đỏ chưa xem, hoặc cron đã ngừng chạy.
	LeakWarning string `json:"leak_warning,omitempty"`
}

func (s *server) handleAdminStats(w http.ResponseWriter, _ *http.Request) {
	adminJSON(w, http.StatusOK, s.collectStats())
}

// collectStats gom số liệu của mọi phòng.
func (s *server) collectStats() adminStats {
	var st adminStats
	st.At = time.Now().Unix()
	st.Build = s.buildID
	st.Uptime = int64(time.Since(s.startedAt).Seconds())
	st.Realms = map[int]int{}
	st.FeedbackNew = s.adminAuth.Store().UnreadFeedback()
	if s.telemetry != nil {
		if g, err := s.telemetry.RecentErrors(24*time.Hour, time.Now()); err == nil {
			st.ErrorsToday = len(g)
		}
	}

	if s.janitor != nil {
		st.DBWarning = s.janitor.DBWarning()
	}
	st.Leak = readLeakNightly(s.dbPath)
	st.LeakWarning = st.Leak.Warning(time.Now())

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	st.Server.Goroutines = runtime.NumGoroutine()
	st.Server.TickWarnMs, st.Server.TickWarnSecs = s.cfg.Admin.TickWarnMs, s.cfg.Admin.TickWarnSecs
	st.Server.HeapMB = float64(ms.HeapAlloc) / (1024 * 1024)

	rooms := s.mgr.AllRooms()
	st.Players.Rooms = len(rooms)
	st.Rooms = make([]roomStat, 0, len(rooms))
	for _, r := range rooms {
		// Gom TRÊN goroutine phòng: `zoneStat` duyệt `g.Players` và `g.Obstacles`, mà
		// goroutine phòng ghi vào đúng hai thứ đó mỗi tick. Duyệt map trong lúc map bị ghi
		// làm Go dừng hẳn chương trình ("concurrent map iteration and map write") — tức là
		// một người bấm Vào trận đúng nhịp làm mới 2 giây của trang này là đủ sập server.
		// Xem `room/query.go`.
		var rs roomStat
		if !r.Query(func() {
			rs = roomStat{ID: r.ID, Humans: r.Humans(), Bots: r.Bots()}
			for i := 0; i < r.ZoneCount(); i++ {
				rs.Zones = append(rs.Zones, s.zoneStat(r, i, &st))
			}
		}) {
			continue // phòng quá tải: bỏ qua một nhịp, không treo cả trang
		}
		st.Players.Humans += rs.Humans
		st.Players.Bots += rs.Bots
		st.Rooms = append(st.Rooms, rs)
	}
	return st
}

func (s *server) zoneStat(r *room.Room, i int, st *adminStats) zoneStat {
	z := r.Zone(i)
	avg, p95 := z.TickStats()
	if avg > st.Server.TickAvgMax {
		st.Server.TickAvgMax = avg
	}
	if p95 > st.Server.TickP95Max {
		st.Server.TickP95Max = p95
	}
	g := z.G
	zs := zoneStat{
		Index: z.Index, Name: z.Name, Map: g.Map.ID,
		Humans: z.Humans(), Bots: z.Count() - z.Humans(),
		TickAvg: avg, TickP95: p95, HotSecs: z.TickHotSecs(),
		Monsters: len(g.Monsters), Items: g.ItemCount(),
		TopLevel: -1,
	}
	if h := z.Health(time.Now()); h.Panics > 0 {
		zs.Degraded, zs.Panics = h.Degraded, h.Panics
		if h.Degraded {
			st.Degraded = append(st.Degraded, z.Name)
		}
	}
	for _, o := range g.Obstacles {
		if o.Broken {
			zs.BrokenObstacles++
		}
	}
	// Cảnh giới cao nhất và phân bố: chỉ đếm NGƯỜI THẬT.
	for _, c := range z.ClientList() {
		if len(c.Flags.List(time.Now())) > 0 {
			st.Flagged++
		}
	}
	for _, p := range g.PlayerList() {
		if p.IsBot || p.Dead {
			continue
		}
		realm := g.Prog.Realm(p.Level)
		st.Realms[realm]++
		if g.Prog.IsTien(realm) {
			st.TienN++
		}
		if p.Level > zs.TopLevel {
			zs.TopLevel = p.Level
			zs.TopRealm = g.Prog.RealmName(realm)
		}
	}
	return zs
}
