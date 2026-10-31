package main

import (
	"net/http"
	"runtime"
	"time"

	"kiemxoay/server/internal/room"
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
	} `json:"server"`
	// Realms: phân bố cảnh giới của NGƯỜI THẬT (§3 nhóm "Trò chơi"). Chỉ số cảnh giới →
	// số người. Bot không tính: dashboard là để nhìn người chơi thật.
	Realms map[int]int `json:"realms"`
	TienN  int         `json:"tien_count"`
	Rooms  []roomStat  `json:"rooms"`
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

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	st.Server.Goroutines = runtime.NumGoroutine()
	st.Server.HeapMB = float64(ms.HeapAlloc) / (1024 * 1024)

	rooms := s.mgr.AllRooms()
	st.Players.Rooms = len(rooms)
	st.Rooms = make([]roomStat, 0, len(rooms))
	for _, r := range rooms {
		rs := roomStat{ID: r.ID, Humans: r.Humans(), Bots: r.Bots()}
		st.Players.Humans += rs.Humans
		st.Players.Bots += rs.Bots
		for i := 0; i < r.ZoneCount(); i++ {
			rs.Zones = append(rs.Zones, s.zoneStat(r, i, &st))
		}
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
		TickAvg: avg, TickP95: p95,
		Monsters: len(g.Monsters), Items: g.ItemCount(),
		TopLevel: -1,
	}
	for _, o := range g.Obstacles {
		if o.Broken {
			zs.BrokenObstacles++
		}
	}
	// Cảnh giới cao nhất và phân bố: chỉ đếm NGƯỜI THẬT.
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
