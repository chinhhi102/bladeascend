package room

import (
	"kiemxoay/server/internal/bot"
	"kiemxoay/server/internal/config"
	"kiemxoay/server/internal/game"
	"kiemxoay/server/internal/world"
)

// Zone là MỘT khu vực của phòng (spec §8.0): một thế giới 12 000 × 12 000 riêng, mô phỏng
// riêng, danh sách người chơi riêng, vùng nhìn (AOI) riêng.
//
// Năm khu **không nối nhau bằng đường đi** — tường biên mỗi khu không qua được — nên hai
// khu không bao giờ phải biết gì về nhau trong một tick. Nhờ vậy mỗi khu chạy được trên
// một goroutine mà không cần khoá: việc duy nhất đi qua ranh giới là **chuyển người giữa
// các khu**, và việc đó làm ở ranh giới tick, lúc không khu nào đang chạy.
//
// Những thứ tính chung CẢ THẾ GIỚI (bảng xếp hạng phòng, danh hiệu, Uy Danh) vẫn nằm ở
// Room, không nằm ở đây.
type Zone struct {
	// Index 0..4 theo bảng §8.0; Name là tên hiển thị cho HUD và kill feed.
	Index int
	Name  string
	G     *game.Game

	clients map[uint32]*Client
	brains  map[uint32]*bot.Brain
	pending []game.Event

	// itemBuf dùng lại giữa các lần dựng snapshot — spec §13.4 cấm cấp phát mỗi tick.
	itemBuf []itemCand
	// stepMs: thời gian xử lý một tick của riêng khu này, gửi kèm snapshot cho overlay F3.
	stepMs  float32
	tick    uint32
	snapAcc float64
}

// ZoneMapIDs: id bản đồ của từng khu, do `cmd/mapgen` sinh ra (spec §8.0).
var ZoneMapIDs = []string{"thien_dinh", "linh_son_coc", "truc_lam", "pham_tran", "bang_nguyet_ho"}

// ZoneNames là tên 5 khu theo đúng thứ tự bảng §8.0.
var ZoneNames = []string{"Thiên Đình", "Linh Sơn Cốc", "Trúc Lâm Mê Cung", "Phàm Trấn", "Băng Nguyệt Hồ"}

// ZoneName trả tên khu thứ i, an toàn với chỉ số ngoài khoảng.
func ZoneName(i int) string {
	if i < 0 || i >= len(ZoneNames) {
		return ""
	}
	return ZoneNames[i]
}

func newZone(index int, cfg *config.Config, m *world.Map, seed int64) *Zone {
	return &Zone{
		Index:   index,
		Name:    ZoneName(index),
		G:       game.New(cfg, m, seed),
		clients: map[uint32]*Client{},
		brains:  map[uint32]*bot.Brain{},
	}
}

// Game mở Game của khu cho công cụ đo và test.
func (z *Zone) Game() *game.Game { return z.G }

// Humans đếm người thật đang ở khu này — `clients` chỉ chứa người, bot nằm ở `brains`.
// Dùng để chọn khu ít người nhất khi vào trận (§8.0).
func (z *Zone) Humans() int { return len(z.clients) }

// Count là tổng nhân vật (người + bot) trong khu — dùng để rải bot cho đều.
func (z *Zone) Count() int { return len(z.clients) + len(z.brains) }
