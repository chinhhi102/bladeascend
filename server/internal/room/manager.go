package room

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"bladeascend/server/internal/config"
	"bladeascend/server/internal/netproto"
	"bladeascend/server/internal/safe"
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
	// resume: mã phiên §7.1 → phòng đang giữ nhân vật đó. Guarded by mu.
	//
	// Ở Manager chứ không ở Room vì lúc client nối lại, nó CHƯA thuộc phòng nào — chỉ có
	// một chuỗi mã trong tay, và phải tìm ra phòng từ chuỗi đó. Quét hết các phòng cũng
	// ra, nhưng quét phải vào goroutine của từng phòng, tức là chờ tới lượt tick của từng
	// cái; một bảng tra là O(1) và không đụng vào tick nào.
	resume map[string]*Room
}

func NewManager(cfg *config.Config, maps []*world.Map, rec Recorder) *Manager {
	if len(maps) == 0 {
		maps = []*world.Map{world.EmptyMap(cfg.Room.MapSize)}
	}
	return &Manager{cfg: cfg, maps: maps, recorder: rec, rng: rand.New(rand.NewSource(time.Now().UnixNano())),
		resume: map[string]*Room{},
		// Vòng chạy của phòng là một goroutine, nên nó phải có recover của RIÊNG nó (§13.8).
		// `Step` đã bọc từng phần bên trong; lớp này chỉ còn đỡ những gì nằm ngoài Step
		// (bản thân vòng lặp, ticker) — và nếu nó bắt được thì phòng được khởi động lại
		// chứ không biến mất im lặng khỏi server.
		Start: func(r *Room) { safe.Go("nen", func() string { return "vòng phòng " + r.ID }, r.Run) }}
}

// Join matches a client into a room: the fullest room below room_max_humans (giá trị admin
// đang đặt, spec v6.28 §9.2), ties → oldest.
func (m *Manager) Join(c *Client, j netproto.Join) *Room {
	// Vào LẠI phiên cũ đi trước ghép phòng mới (§7.1): có mã phiên còn hiệu lực thì người
	// này không phải người mới, họ đang quay lại giữa ván.
	if j.Resume != "" {
		m.mu.Lock()
		r := m.resume[j.Resume]
		m.mu.Unlock()
		if r != nil {
			// Thử vào lại NGAY TRONG goroutine của phòng, và nếu không được thì vào trận
			// mới cũng ngay tại đó.
			//
			// Không chờ kết quả rồi mới quyết: `Manager.Join` chạy trên goroutine ĐỌC
			// WEBSOCKET của chính người chơi đó, và chặn nó để đợi một phòng tick là mời
			// một cú treo — phòng đang kẹt thì người chơi ngồi nhìn màn hình tải.
			r.Do(func() {
				if !r.resumeHuman(c, j, j.Resume) {
					m.joinFresh(c, j)
				}
			})
			return r
		}
	}
	return m.joinFresh(c, j)
}

// rememberResume / holdResume / forgetResume là sổ tra mã phiên → phòng (§7.1).
//
// `rememberResume` ghi mã của một người ĐANG CHƠI: cần nó để tab thứ hai mở bằng cùng mã
// tìm được phòng mà chiếm nhân vật. `holdResume` ghi mã của người đang mất kết nối.
// Hai hàm làm cùng một việc nhưng tách tên để chỗ gọi đọc ra ý định.
func (m *Manager) rememberResume(key string, r *Room) {
	m.mu.Lock()
	m.resume[key] = r
	m.mu.Unlock()
}

func (m *Manager) holdResume(key string, r *Room, _ uint32) {
	m.rememberResume(key, r)
}

func (m *Manager) forgetResume(key string) {
	if key == "" {
		return
	}
	m.mu.Lock()
	delete(m.resume, key)
	m.mu.Unlock()
}

// ResumeCount là số mã phiên đang nhớ — bài test rò rỉ (§13.9) kiểm nó về 0.
func (m *Manager) ResumeCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.resume)
}

func (m *Manager) joinFresh(c *Client, j netproto.Join) *Room {
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
			// Dồn trái, XOÁ ô cuối, rồi mới hạ `len`.
			//
			// `append(m.rooms[:i], m.rooms[i+1:]...)` một mình là **rò rỉ bộ nhớ**: nó dồn
			// trái và hạ `len`, nhưng ô cuối của mảng nền vẫn giữ con trỏ. Mà một `*Room`
			// là cả một thế giới — 5 khu, ~28 000 viên EXP, 5 lưới A*, ≈ 13 MB — và bộ dọn
			// rác nhìn MẢNG NỀN chứ không nhìn `len`, nên phòng đã đóng không bao giờ được
			// thu hồi.
			//
			// Tìm ra ngày 12/11 bằng `cmd/leakcheck` (§13.9): sau khi MỌI phòng đã đóng,
			// heap vẫn còn nguyên một thế giới. Cùng họ với lỗi `itemGrid.remove`.
			last := len(m.rooms) - 1
			copy(m.rooms[i:], m.rooms[i+1:])
			m.rooms[last] = nil
			m.rooms = m.rooms[:last]
			break
		}
	}
	// Mã phiên §7.1 trỏ tới phòng này cũng phải quên: một mã còn sống là một con trỏ
	// `*Room` còn sống, và nó giữ lại đúng cái thế giới vừa đóng.
	m.forgetRoomLocked(r)
	return true
}

// forgetRoomLocked xoá mọi mã phiên trỏ tới một phòng. Gọi khi đang giữ `m.mu`.
func (m *Manager) forgetRoomLocked(r *Room) {
	for k, v := range m.resume {
		if v == r {
			delete(m.resume, k)
		}
	}
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
