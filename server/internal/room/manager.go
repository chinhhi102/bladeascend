package room

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"bladeascend/server/internal/config"
	"bladeascend/server/internal/netproto"
	"bladeascend/server/internal/world"
)

// Manager does matchmaking across rooms (spec §9.2).
type Manager struct {
	mu       sync.Mutex
	cfg      *config.Config
	maps     []*world.Map
	rooms    []*Room
	recorder Recorder
	nextID   int
	rng      *rand.Rand
	// Start launches a room's goroutine; cmd/sim replaces it to step rooms manually.
	Start func(r *Room)
	// AllowDebug: server chạy với -debug, nên client xin ?debug=true thì được bật.
	AllowDebug bool
	// ScoreMedian: Uy Danh/phút trung vị theo cảnh giới, cho cờ "Tăng điểm bất thường"
	// (§6.1). `cmd/server` cắm vào; nil thì không gắn cờ loại này.
	ScoreMedian func(realm int) float64
	// LangShares: phân bố ngôn ngữ người chơi toàn server 7 ngày (mã → tỉ lệ), cho tên bot
	// §9.4. `cmd/server` cắm vào; nil = chưa có số liệu, dùng người thật đang online.
	LangShares func() map[string]float64
	// settings: cấu hình phòng admin đã đặt (settings.go), nil = mặc định môi trường. Guarded by mu.
	settings *RoomSettings
	// botMode: chế độ bot toàn server (botmode.go). Guarded by mu.
	botMode string
}

func NewManager(cfg *config.Config, maps []*world.Map, rec Recorder) *Manager {
	if len(maps) == 0 {
		maps = []*world.Map{world.EmptyMap(cfg.Room.MapSize)}
	}
	return &Manager{cfg: cfg, maps: maps, recorder: rec, rng: rand.New(rand.NewSource(time.Now().UnixNano())),
		Start: func(r *Room) { go r.Run() }}
}

// Join matches a client into a room: the fullest room below room_max_humans (giá trị admin
// đang đặt, spec v6.28 §9.2), ties → oldest.
func (m *Manager) Join(c *Client, j netproto.Join) *Room {
	m.mu.Lock()
	var best *Room
	limit := m.settingsLocked().MaxHumans
	for _, r := range m.rooms {
		if r.closed || r.reserved >= limit {
			continue
		}
		if best == nil || r.reserved > best.reserved {
			best = r // rooms are kept in creation order, so ties keep the oldest
		}
	}
	if best == nil {
		best = m.createLocked()
	}
	best.reserved++
	m.mu.Unlock()
	best.Do(func() { best.addHuman(c, j) })
	return best
}

func (m *Manager) createLocked() *Room {
	m.nextID++
	r := newRoom(fmt.Sprintf("r%d", m.nextID), m.cfg, m.zoneMaps(), m, m.recorder, m.rng.Int63())
	// Mốc trung vị dùng chung cho mọi phòng (§6.1) — cắm từ Manager để phòng mới sinh ra
	// giữa chừng cũng có, không phải nhớ gắn tay.
	r.ScoreMedian = m.ScoreMedian
	r.LangShares = m.LangShares
	m.rooms = append(m.rooms, r)
	m.Start(r)
	return r
}

// zoneMaps chọn bản đồ cho 5 khu của một phòng (spec §8.0).
//
// Mỗi khu phải là một bản đồ KHÁC NHAU; khi chưa đủ 5 bản đồ thì dùng lại bản đã có để
// phòng vẫn chạy — thà năm khu giống nhau còn hơn không vào được game. `cmd/mapgen` sẽ
// sinh đủ 5 kiểu khu (spec v5 mục 6).
func (m *Manager) zoneMaps() []*world.Map {
	n := m.cfg.Room.Zones
	if n < 1 {
		n = 1
	}
	byID := map[string]*world.Map{}
	for _, mp := range m.maps {
		byID[mp.ID] = mp
	}
	out := make([]*world.Map, 0, n)
	for i := 0; i < n; i++ {
		if mp, ok := byID[ZoneMapIDs[min(i, len(ZoneMapIDs)-1)]]; ok {
			out = append(out, mp)
			continue
		}
		// Chưa sinh đủ 5 bản đồ khu thì dùng lại bản có sẵn — thà năm khu giống nhau còn
		// hơn không vào được game. Sinh đủ bằng `cd server && go run ./cmd/mapgen`.
		out = append(out, m.maps[i%len(m.maps)])
	}
	return out
}

// ZoneInfo: id bản đồ + tên hiển thị của một khu — trang admin vẽ một ô sàn cho mỗi khu.
type ZoneInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ZoneList: các khu của một phòng, đúng thứ tự §8.0.
func (m *Manager) ZoneList() []ZoneInfo {
	maps := m.zoneMaps()
	out := make([]ZoneInfo, len(maps))
	for i, mp := range maps {
		out[i] = ZoneInfo{ID: mp.ID, Name: ZoneName(i)}
	}
	return out
}

// release frees a human slot when a player leaves or dies.
func (m *Manager) release(r *Room) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r.reserved > 0 {
		r.reserved--
	}
}

func (m *Manager) humansOf(r *Room) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return r.reserved
}

// tryClose removes an empty room; it fails if a join raced in.
func (m *Manager) tryClose(r *Room) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r.reserved > 0 {
		return false
	}
	r.closed = true
	for i, o := range m.rooms {
		if o == r {
			m.rooms = append(m.rooms[:i], m.rooms[i+1:]...)
			break
		}
	}
	return true
}

// BroadcastFeed sends one kill-feed line to every room except `from` (spec §11.4: hạ một
// người trong Top 5 thì cả server đều biết). Each room delivers it on its own goroutine.
func (m *Manager) BroadcastFeed(f netproto.Feed, from *Room) {
	m.mu.Lock()
	rooms := append([]*Room(nil), m.rooms...)
	m.mu.Unlock()
	for _, r := range rooms {
		if r == from {
			continue // phòng gốc đã nhận qua đường thường
		}
		room := r
		room.Do(func() {
			for _, c := range room.Clients() {
				c.send(f)
			}
		})
	}
}

// RoomInfo is a summary for the status endpoint.
type RoomInfo struct {
	ID     string `json:"id"`
	Map    string `json:"map"`
	Humans int    `json:"humans"`
}

func (m *Manager) Rooms() []RoomInfo {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]RoomInfo, 0, len(m.rooms))
	for _, r := range m.rooms {
		out = append(out, RoomInfo{ID: r.ID, Map: r.zones[0].G.Map.ID, Humans: r.reserved})
	}
	return out
}

// AllRooms returns the live rooms (for cmd/sim).
func (m *Manager) AllRooms() []*Room {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]*Room(nil), m.rooms...)
}
