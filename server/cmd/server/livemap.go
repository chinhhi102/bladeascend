package main

import (
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"

	"kiemxoay/server/internal/room"
)

// Xem bản đồ trực tiếp — ADMIN_SPEC §5.1.
//
// **Chỉ xem.** Không có thao tác nào lên trận ở đây (nút kick nằm ở §6, một đường khác).
//
// Ba ràng buộc của §5.1 quyết định toàn bộ thiết kế:
//
//  1. "tách khỏi luồng snapshot người chơi" — không chạm vào `diffPlayers` hay AOI. Đây là
//     một handler HTTP đọc trạng thái, người chơi không trả một byte nào cho nó.
//  2. "5 lần/giây" — client tự gọi lại; server không giữ kết nối dài. Một endpoint đọc là
//     thứ rẻ nhất có thể, và nó tự dừng khi admin đóng tab chứ không cần dọn gì.
//  3. "tối đa 2 admin xem cùng lúc" — đếm bằng `liveViewers`. Trần này không phải để tiết
//     kiệm băng thông mà để giữ lời hứa §11: admin xem KHÔNG được làm tick người chơi tăng
//     quá 0,3 ms. Mỗi người xem là 5 lần/giây chen vào goroutine phòng.

// maxLiveViewers theo §5.1.
const maxLiveViewers = 2

// liveViewers đếm người đang xem; mỗi request giữ một suất trong lúc nó chạy.
//
// Đếm REQUEST ĐANG CHẠY chứ không đếm "phiên đang mở": không có cách nào biết admin đã
// đóng tab hay chưa, nên mọi cách đếm theo phiên đều cần hết hạn, và hết hạn thì có lúc
// admin thứ nhất đã đóng tab mà admin thứ hai vẫn bị chặn.
var liveViewers atomic.Int32

// livePlayer là một chấm người trên bản đồ (§5.1).
type livePlayer struct {
	ID uint32 `json:"id"`
	// X/Y của người chơi KHÔNG lượng tử hoá theo ô như viên: chấm người di chuyển liên tục
	// và §5.1 cho bám theo một người, nên nhảy 16 đơn vị một bước là thấy giật. Làm tròn
	// tới số nguyên là đủ (1 đơn vị ≈ 1/13 điểm ảnh).
	X     int     `json:"x"`
	Y     int     `json:"y"`
	Name  string  `json:"name"`
	Char  string  `json:"char"`
	Bot   bool    `json:"bot"`
	Dead  bool    `json:"dead"`
	Level int     `json:"lv"`
	Scale float64 `json:"sc"`
	Reach int     `json:"reach"`
}

// liveQ: lượng tử hoá toạ độ cho bản đồ quản trị — §5.1 "lượng tử hoá như snapshot".
//
// Khu rộng 12 000 đơn vị vẽ vào canvas 900 px, tức 1 px ≈ 13 đơn vị. Gửi `9761.202420898862`
// cho một chấm rộng 1 px là gửi 13 chữ số để vẽ đúng một điểm ảnh. Chia cho 16 thì mất
// nhiều nhất nửa điểm ảnh — không ai nhìn ra — mà mỗi số còn 3 chữ số.
const liveQ = 16

// Viên Tiên khí và quái gửi dưới dạng MẢNG SỐ PHẲNG theo loại, không phải mảng đối tượng.
//
// Một khu có ~6 300 viên. Dạng `{"x":...,"y":...,"k":"orb"}` cho ra 390 KB mỗi lần gọi,
// nhân 5 lần/giây nhân 2 người xem là 3,9 MB/giây — cho một trang "chỉ xem". Dạng phẳng
// `{"orb":[610,696,...]}` còn ~45 KB. Và quan trọng hơn: dựng một slice int thay vì 6 300
// struct có chuỗi bên trong là thứ chạy được BÊN TRONG goroutine phòng mà không ăn vào
// ngân sách tick 4 ms (§13.6, và lời hứa §11 "không quá 0,3 ms").
type liveGroup map[string][]int

// liveView là gói gửi cho trang quản trị mỗi 200 ms.
type liveView struct {
	At   int64  `json:"at"`
	Room string `json:"room"`
	Zone string `json:"zone"`
	Map  string `json:"map"`
	Size float64 `json:"size"`
	// Broken: chỉ số của vật cản ĐANG BỊ CHÉM (§5.1 "hiện nét đứt"). Gửi chỉ số chứ không
	// gửi lại cả vật cản: bố cục tĩnh client đã có từ `/api/map/{id}`, và gửi lại 900 vật
	// cản 5 lần/giây là 4,5 MB/phút cho một thứ không bao giờ đổi.
	Broken  []int      `json:"broken"`
	Players []livePlayer `json:"players"`
	// Items/Monsters: {loại: [x1,y1,x2,y2,…]}, toạ độ đã chia cho `Q`.
	Items    liveGroup `json:"items"`
	Monsters liveGroup `json:"monsters"`
	Q        int       `json:"q"`
	// NItems/NMonsters: số thật, để trang quản trị nói đúng con số chứ không phải đếm lại
	// từ mảng đã gộp ô (hai viên cùng một ô thì chỉ còn một cặp toạ độ).
	NItems    int `json:"n_items"`
	NMonsters int `json:"n_monsters"`
}

// handleAdminLive trả ảnh chụp một khu tại thời điểm gọi.
func (s *server) handleAdminLive(w http.ResponseWriter, r *http.Request) {
	if n := liveViewers.Add(1); n > maxLiveViewers {
		liveViewers.Add(-1)
		adminJSON(w, http.StatusTooManyRequests,
			map[string]string{"error": "đã có 2 người đang xem bản đồ trực tiếp (§5.1)"})
		return
	}
	defer liveViewers.Add(-1)

	roomID := r.URL.Query().Get("room")
	zoneIdx, _ := strconv.Atoi(r.URL.Query().Get("zone"))

	var out *liveView
	for _, rm := range s.mgr.AllRooms() {
		if roomID != "" && rm.ID != roomID {
			continue
		}
		// Gom TRÊN goroutine phòng — xem `room/query.go`. Duyệt `g.Players` từ goroutine
		// HTTP làm Go dừng hẳn chương trình.
		rm.Query(func() {
			if zoneIdx < 0 || zoneIdx >= rm.ZoneCount() {
				return
			}
			out = liveViewOf(rm, zoneIdx)
		})
		break
	}
	if out == nil {
		adminJSON(w, http.StatusNotFound, map[string]string{"error": "không có phòng/khu này"})
		return
	}
	adminJSON(w, http.StatusOK, out)
}

// liveViewOf gom ảnh chụp. CHỈ gọi từ trong `Query`.
func liveViewOf(rm *room.Room, zoneIdx int) *liveView {
	z := rm.Zone(zoneIdx)
	g := z.G
	v := &liveView{
		At: time.Now().UnixMilli(), Room: rm.ID, Zone: z.Name, Map: g.Map.ID, Size: g.Map.Size,
		Broken: []int{}, Players: []livePlayer{}, Items: liveGroup{}, Monsters: liveGroup{},
		Q: liveQ,
	}
	for i, o := range g.Obstacles {
		if o.Broken {
			v.Broken = append(v.Broken, i)
		}
	}
	for _, p := range g.PlayerList() {
		v.Players = append(v.Players, livePlayer{
			ID: p.ID, X: int(p.Pos.X), Y: int(p.Pos.Y), Name: p.Name, Char: p.Char,
			Bot: p.IsBot, Dead: p.Dead, Level: p.Level, Scale: math.Round(p.Scale*100) / 100,
			Reach: int(p.Geom.Reach),
		})
	}
	// Gộp các viên rơi vào cùng một ô lượng tử: chúng vẽ chồng lên đúng một điểm ảnh, nên
	// gửi cả hai là gửi thừa. `NItems` vẫn giữ con số thật để không im lặng giấu mất gì.
	seen := map[int32]bool{}
	for _, it := range g.Items {
		v.NItems++
		qx, qy := int(it.Pos.X)/liveQ, int(it.Pos.Y)/liveQ
		key := int32(qx)<<12 | int32(qy&0xfff)
		if seen[key] {
			continue
		}
		seen[key] = true
		v.Items[it.Kind] = append(v.Items[it.Kind], qx, qy)
	}
	for _, m := range g.Monsters {
		v.NMonsters++
		v.Monsters[m.Kind] = append(v.Monsters[m.Kind], int(m.Pos.X)/liveQ, int(m.Pos.Y)/liveQ)
	}
	return v
}

// handleAdminCheck trả báo cáo `mapgen --check` của bộ bản đồ ĐANG CHẠY — §5.2.
//
// Đọc từ đĩa mỗi lần gọi, không cache: tệp này thay đổi khi ai đó sinh lại bản đồ, và một
// báo cáo cũ nói về một bố cục không còn tồn tại thì tệ hơn hẳn việc không có báo cáo.
func (s *server) handleAdminCheck(w http.ResponseWriter, _ *http.Request) {
	b, err := os.ReadFile(filepath.Join(s.mapsDir, "check.json"))
	if err != nil {
		adminJSON(w, http.StatusNotFound, map[string]string{
			"error": "chưa có báo cáo. Chạy: cd server && go run ./cmd/mapgen --check",
		})
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(b)
}
