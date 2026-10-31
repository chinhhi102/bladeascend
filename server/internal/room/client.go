package room

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	"kiemxoay/server/internal/game"
	"kiemxoay/server/internal/netproto"
)

// Conn is the transport for one browser tab; implemented by the WebSocket handler.
type Conn interface {
	// Send queues a message without blocking; it returns false when the queue is full.
	Send(b []byte) bool
}

// LossyConn là Conn biết bỏ gói CŨ khi hàng đợi đầy (spec mục 27).
//
// Mạng yếu thì hàng đợi đầy, và một snapshot cũ nằm trong hàng là thứ vô giá trị: client chỉ
// cần bản MỚI NHẤT. Giữ cái cũ rồi vứt cái mới là cách chắc chắn nhất để người chơi thấy
// nhân vật giật lùi.
type LossyConn interface {
	SendLossy(b []byte) bool
}

// Client is a connected human. The room goroutine owns everything except the input fields.
type Client struct {
	Conn    Conn
	GuestID string
	// Debug: kết nối này được dùng lệnh test không (ADMIN_SPEC §1.1).
	//
	// KHÔNG lấy từ `?debug=true` trên URL. Server tự quyết định lúc bắt tay (phiên admin,
	// hoặc APP_ENV=dev trên localhost, hoặc DEBUG_TOKEN ở dev) rồi mới báo cho client. URL
	// một mình không bao giờ đủ — nếu không thì ai cũng tự lên Đạo Tổ bằng cách sửa địa chỉ.
	Debug bool
	// DebugSource ghi quyền debug đến TỪ ĐÂU, để vào nhật ký `admin_audit`.
	DebugSource string
	// RecheckDebug kiểm lại quyền (phiên admin có thể hết hạn GIỮA TRẬN). nil = không kiểm.
	RecheckDebug func() (bool, string)
	debugNextAt  time.Time
	// DebugUsed: kết nối này đã dùng ÍT NHẤT MỘT lệnh debug.
	//
	// Một khi bật thì không bao giờ tắt trong phiên đó: §1.1 nói mạng chơi có dùng lệnh
	// debug thì không được ghi vào bảng xếp hạng toàn cục, và "đã dùng rồi" là chuyện đã
	// rồi — tắt debug đi không làm sạch lại số điểm đã nặn ra.
	DebugUsed bool
	// RefusedCmds đếm lệnh bị từ chối, cho cờ nghi hack "Hành động bị từ chối" (§6.1).
	RefusedCmds int

	// Flags: cờ nghi hack của kết nối này (§6.1). Không bao giờ nil — xem `NewClient`.
	// CHỈ gắn cờ; không có đường nào từ đây dẫn tới việc ngắt kết nối (xem flags.go).
	Flags *Flags
	// IP là địa chỉ THÔ, chỉ giữ trong bộ nhớ của kết nối đang mở và chỉ để hai việc §6 cần:
	// hiện che bớt cho admin (`113.161.x.x`), và so khi cấm theo IP. Không ghi ra đĩa, không
	// vào telemetry — chỗ đó dùng bản băm (TELEMETRY_SPEC §1.3).
	IP string
	// UA là User-Agent, để cột "Thiết bị / trình duyệt" (§6).
	UA string
	// JoinedAt là lúc kết nối vào phòng, cho cột "thời gian sống".
	JoinedAt time.Time

	// pingMu giữ ba số dưới đây: chúng được ghi từ goroutine đọc WebSocket và đọc từ
	// goroutine gom số liệu cho trang quản trị.
	pingMu sync.Mutex
	rttMs  float64
	fps    float64
	// corrections: số lần server kéo vị trí client về / phút (§6).
	corrections int
	corrStart   time.Time
	// lastPingMs: mốc đồng hồ client gửi ở lần ping trước (xem NotePing).
	lastPingMs float64

	PlayerID uint32
	room     *Room

	knownPlayers map[uint32]bool
	// knownItems[slot] = id của viên đã gửi ở slot đó (0 = chưa gửi gì). Dùng SLICE theo chỉ
	// số dày đặc chứ không dùng map băm: pprof đo được việc tra `knownItems[id]` trong vòng
	// lọc tầm nhìn — chạy cho từng viên, từng người, 20 lần/giây — ăn ~25% CPU server. Lưu
	// id chứ không lưu bool vì slot được DÙNG LẠI: viên mới nhận slot của viên cũ thì id
	// khác nhau, so id là biết ngay client chưa thấy viên mới này.
	knownItems []uint32
	// knownSlots: danh sách GỌN các slot đang khác 0, để vòng quét xoá khỏi phải duyệt cả
	// mảng `knownItems` dài hàng nghìn phần tử (xem `diffItems`).
	knownSlots []int32
	// knownMonsters: máu LẦN CUỐI đã gửi cho client này, theo id quái. Có mặt trong map = đã
	// gửi phần tĩnh (loại, máu tối đa, bán kính). Dùng để chỉ gửi máu khi nó đổi (spec mục 27).
	knownMonsters map[uint32]int

	// sentBytes: tổng byte đã gửi cho client này (telemetry §3). Atomic vì goroutine ghi
	// telemetry đọc nó trong lúc goroutine phòng đang cộng vào.
	sentBytes atomic.Int64
	sentMark  atomic.Int64 // mốc của lần lấy mẫu trước (xem sentSince)

	// descAsk: entity mà client xin gửi lại mô tả (xem ForgetDesc). Khoá bằng `inMu` vì nó
	// được ghi từ goroutine đọc WebSocket và đọc từ goroutine phòng.
	descAsk map[uint32]bool

	// aoi theo dõi client này đã biết những entity nào, để gửi `Despawn` tường minh
	// (NETCODE_SPEC §4.3). Chỉ goroutine phòng đụng tới.
	aoi *aoiTracker

	evMu     sync.Mutex
	evWindow time.Time
	evCount  int

	inMu   sync.Mutex
	seq    uint32
	angle  float64
	moving bool
	burn   bool
	// skill: nút kỹ năng vừa được bấm (sự kiện một lần, xem SetInput).
	skill bool
	dirty bool
}

// DebugAllowed trả về quyền debug HIỆN TẠI, kiểm lại định kỳ.
//
// Kiểm lại là bắt buộc chứ không phải cẩn thận thừa: §1.1 nói "phiên admin hết hạn giữa
// trận → tắt quyền debug ngay ở lần kiểm tiếp theo". Không kiểm lại thì một phiên 12 giờ
// đã hết hạn vẫn giữ quyền cho tới khi người đó đóng tab — có thể là nhiều ngày.
func (c *Client) DebugAllowed(now time.Time) bool {
	if c.RecheckDebug != nil && now.After(c.debugNextAt) {
		c.debugNextAt = now.Add(debugRecheckEvery)
		c.Debug, c.DebugSource = c.RecheckDebug()
	}
	return c.Debug
}

// debugRecheckEvery khớp `admin.RecheckEvery` (§1.1: 60 giây). Để hằng số riêng ở đây để
// `internal/room` không phải phụ thuộc vào package `admin`.
const debugRecheckEvery = 60 * time.Second

func NewClient(c Conn, debug bool) *Client {
	return &Client{Conn: c, Debug: debug, Flags: newFlags(), JoinedAt: time.Now()}
}

// NotePing ghi RTT đo được, và gắn cờ "Ping giả" nếu mốc thời gian vô lý (§6.1).
//
// `clientMs` là đồng hồ của MÁY NGƯỜI CHƠI nên nó lệch so với server là bình thường — múi
// giờ, đồng hồ chạy sai, máy vừa ngủ dậy. Nên không so hai đồng hồ với nhau; chỉ đo khoảng
// cách giữa hai lần chính client đó gửi, và chỉ coi là giả khi nó đi LÙI.
func (c *Client) NotePing(now time.Time, clientMs float64) {
	c.pingMu.Lock()
	last := c.lastPingMs
	c.lastPingMs = clientMs
	if last > 0 {
		if d := clientMs - last; d < 0 {
			c.pingMu.Unlock()
			c.NoteFakePing(now, d)
			return
		}
	}
	c.pingMu.Unlock()
}

// NoteRTT ghi độ trễ vòng do client tự báo qua telemetry.
func (c *Client) NoteRTT(rtt, fps float64) {
	c.pingMu.Lock()
	defer c.pingMu.Unlock()
	if rtt > 0 {
		c.rttMs = rtt
	}
	if fps > 0 {
		c.fps = fps
	}
}

// NoteCorrection đếm một lần server kéo vị trí client về (§6 cột "hiệu chỉnh / phút").
func (c *Client) NoteCorrection(now time.Time) {
	c.pingMu.Lock()
	defer c.pingMu.Unlock()
	if now.Sub(c.corrStart) > time.Minute {
		c.corrStart, c.corrections = now, 0
	}
	c.corrections++
}

// Net trả về RTT, FPS và số hiệu chỉnh/phút cho trang quản trị.
func (c *Client) Net() (rttMs, fps float64, corrections int) {
	c.pingMu.Lock()
	defer c.pingMu.Unlock()
	return c.rttMs, c.fps, c.corrections
}

// SetInput is called from the connection goroutine; the room applies it on its next tick.
// resetKnown xoá trí nhớ "đã gửi gì cho client này" — gọi khi vào phòng và khi sang khu
// khác, vì sang khu là đổi hẳn bản đồ, mọi id cũ không còn nghĩa gì.
func (c *Client) resetKnown() {
	c.knownPlayers, c.knownMonsters = map[uint32]bool{}, map[uint32]int{}
	// Đổi phòng/khu là quên sạch: entity của khu cũ không còn liên quan, và giữ lại thì
	// gói đầu tiên ở khu mới sẽ chở một đống `Despawn` vô nghĩa.
	if c.aoi == nil {
		c.aoi = newAOITracker()
	} else {
		c.aoi.Forget()
	}
	c.knownItems = c.knownItems[:0]
	c.knownSlots = c.knownSlots[:0]
}

func (c *Client) SetInput(in netproto.Input) {
	c.inMu.Lock()
	defer c.inMu.Unlock()
	c.seq = in.Seq
	c.angle = float64(in.Angle) / 256 * 2 * math.Pi
	c.moving, c.burn = in.Moving, in.Burn
	// Nút kỹ năng là sự kiện MỘT LẦN, không phải trạng thái giữ phím: gộp bằng OR để một
	// lần bấm không bị mất khi hai gói input rơi vào cùng một tick.
	c.skill = c.skill || in.Skill
	c.dirty = true
}

func (c *Client) takeInput() (seq uint32, angle float64, moving, burn, skill, ok bool) {
	c.inMu.Lock()
	defer c.inMu.Unlock()
	d, sk := c.dirty, c.skill
	c.dirty, c.skill = false, false
	return c.seq, c.angle, c.moving, c.burn, sk, d
}

// Room returns the room the client is currently in (nil between matches).
func (c *Client) Room() *Room {
	c.inMu.Lock()
	defer c.inMu.Unlock()
	return c.room
}

func (c *Client) setRoom(r *Room) {
	c.inMu.Lock()
	defer c.inMu.Unlock()
	c.room = r
}

func (c *Client) send(v any) {
	b, err := netproto.Encode(v)
	if err == nil {
		c.countOut(len(b))
		c.Conn.Send(b)
	}
}

// countOut cộng dồn số byte đã gửi cho client này (TELEMETRY_SPEC §3: "KB/s gửi mỗi client").
//
// Đếm ở ĐÂY chứ không ở tầng WebSocket vì đây là chỗ duy nhất biết gói thuộc về client nào
// mà vẫn nằm trong `room`; tầng dưới chỉ thấy một đống byte không tên.
func (c *Client) countOut(n int) { c.sentBytes.Add(int64(n)) }

// SentBytes là tổng byte đã gửi cho client này từ lúc nối. Gọi được từ goroutine bất kỳ.
func (c *Client) SentBytes() int64 { return c.sentBytes.Load() }

// sentSince trả về số byte đã gửi KỂ TỪ lần gọi trước, rồi đặt lại mốc.
func (c *Client) sentSince() int64 {
	cur := c.sentBytes.Load()
	return cur - c.sentMark.Swap(cur)
}

// ForgetDesc quên rằng đã gửi mô tả của những entity này, để snapshot sau gửi lại
// (NETCODE_SPEC §5 Hiển thị mục 2).
//
// Chỉ quên entity client HỎI, không quên sạch bảng: quên sạch thì gói kế tiếp phải chở tên
// và nhân vật của toàn bộ người trong tầm nhìn — đúng thứ `knownPlayers` sinh ra để tránh.
//
// Có trần: một client hỏng (hoặc cố tình) có thể bắn `dreq` liên tục để ép server gửi lại
// mô tả mọi người mỗi gói. Trần làm việc đó thành vô ích.
func (c *Client) ForgetDesc(ids []uint32) {
	const maxPerRequest = 32
	c.inMu.Lock()
	defer c.inMu.Unlock()
	if c.descAsk == nil {
		c.descAsk = map[uint32]bool{}
	}
	for i, id := range ids {
		if i >= maxPerRequest {
			break
		}
		c.descAsk[id] = true
	}
}

// takeDescAsks lấy danh sách entity cần gửi lại mô tả, rồi xoá.
// Gọi TRONG goroutine phòng (lúc dựng snapshot).
func (c *Client) takeDescAsks() map[uint32]bool {
	c.inMu.Lock()
	defer c.inMu.Unlock()
	if len(c.descAsk) == 0 {
		return nil
	}
	out := c.descAsk
	c.descAsk = nil
	return out
}

// AllowEvent là trần sự kiện telemetry của một kết nối (TELEMETRY_SPEC §2.3).
//
// Cần ở server chứ không chỉ ở client: sự kiện hay bắn nhất là `error`, và một tab dính lỗi
// trong vòng vẽ bắn nó 60 lần/giây. Trần ở client không cứu được khi chính client là thứ
// đang hỏng.
func (c *Client) AllowEvent() bool {
	now := time.Now()
	c.evMu.Lock()
	defer c.evMu.Unlock()
	if now.Sub(c.evWindow) >= time.Minute {
		c.evWindow, c.evCount = now, 0
	}
	c.evCount++
	return c.evCount <= eventsPerMin
}

// eventsPerMin: trần sự kiện telemetry mỗi phút mỗi kết nối.
const eventsPerMin = 40

// ZoneKbpsOut là KB/s TRUNG BÌNH MỘT NGƯỜI mà server gửi ra ở khu này.
//
// Trung bình một người chứ không phải tổng: tổng chỉ nói khu đông hay vắng, còn con số cần
// theo dõi là "một người tốn bao nhiêu băng thông", vì đó là thứ so được với ngân sách §1
// và không đổi khi số người đổi.
//
// Chỉ gọi TRONG goroutine phòng (qua `Room.Do`): nó duyệt map client của khu.
func (r *Room) ZoneKbpsOut(z *Zone, secs float64) float64 {
	if secs <= 0 || len(z.clients) == 0 {
		return 0
	}
	var total int64
	for _, c := range z.clients {
		total += c.sentSince()
	}
	return float64(total) / 1024 / secs / float64(len(z.clients))
}

// sendSnapshot gửi gói ảnh trạng thái: được phép bỏ bản cũ khi nghẽn, khác với `welcome` hay
// `leaderboard` vốn mất là mất luôn.
func (c *Client) sendSnapshot(v any) {
	b, err := netproto.Encode(v)
	if err != nil {
		return
	}
	c.countOut(len(b))
	if lossy, ok := c.Conn.(LossyConn); ok {
		lossy.SendLossy(b)
		return
	}
	c.Conn.Send(b)
}

// playerState converts a game player to its wire form as seen by one viewer. Ẩn Thân is
// per-viewer: far away the target is flagged stealth (drawn faded, no name, no title),
// up close they look normal (spec §6.2).
func playerState(g *game.Game, viewer, p *game.Player, withName bool) netproto.PlayerState {
	faded := g.StealthFaded(viewer, p)
	s := netproto.PlayerState{
		ID: p.ID, X: float32(p.Pos.X), Y: float32(p.Pos.Y), Facing: float32(p.Facing),
		Level: p.Level, Count: p.Ring.Count, Phase: float32(p.Ring.Angle), Dir: int8(p.Ring.Dir),
		Tilt: float32(p.Ring.Tilt), Scale: float32(p.Scale), Mood: game.MoodIndex(p.Mood),
		Crack: float32(p.Ring.Crack), Flags: playerFlags(g, viewer, p),
	}
	// Omega bỏ hẳn khỏi gói tin: client tự tính được từ rev_secs và buff Tốc Kiếm mà nó đã
	// có sẵn (spec mục 27). Gửi lại một con số suy ra được là lãng phí 7 byte mỗi người mỗi
	// gói, 20 lần mỗi giây.
	if !faded {
		s.Title = p.Title
	}
	if p == viewer {
		s.Exp = float32(p.TotalExp)
		s.Vx, s.Vy = float32(p.Vel.X), float32(p.Vel.Y)
		s.Charges = p.SkillCharges
		have, need := g.TienProgress(p)
		s.TienHave, s.TienNeed = float32(have), float32(need)
	}
	s.St = p.StatusBits(g.Time)
	for b := 0; b < game.BuffCount; b++ {
		if p.HasBuff(b, g.Time) {
			s.Buffs |= 1 << b
		}
	}
	if withName && !faded {
		s.Name, s.Char = p.Name, p.Char
	}
	return s
}

func playerFlags(g *game.Game, viewer, p *game.Player) uint8 {
	var f uint8
	set := func(cond bool, bit uint8) {
		if cond {
			f |= bit
		}
	}
	set(p.Invulnerable(g.Time), netproto.FlagInvuln)
	set(p.Bush >= 0, netproto.FlagBush)
	// Your own stealth shows whenever the buff is up (so you can tell); other people's only
	// once you are far enough that they fade (spec §6.2).
	set(p == viewer && p.HasBuff(game.BuffStealth, g.Time) || g.StealthFaded(viewer, p), netproto.FlagStealth)
	set(p.InTransit(g.Time), netproto.FlagTransit)
	set(p.Dead, netproto.FlagDead)
	set(p.IsBot, netproto.FlagBot)
	set(p.RevealedUntil > g.Time, netproto.FlagRevealed)
	set(g.Burning(p), netproto.FlagBurning)
	return f
}
