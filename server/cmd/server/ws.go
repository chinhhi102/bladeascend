package main

import (
	"context"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/coder/websocket"

	"bladeascend/server/internal/admin"
	"bladeascend/server/internal/config"
	"bladeascend/server/internal/netproto"
	"bladeascend/server/internal/room"
	"bladeascend/server/internal/telemetry"
)

const (
	sendQueue    = 256
	maxMsgBytes  = 4096
	writeTimeout = 5 * time.Second
)

// wsConn adapts a WebSocket to room.Conn with a buffered writer goroutine.
type wsConn struct {
	out chan []byte
}

func (c *wsConn) Send(b []byte) bool {
	select {
	case c.out <- b:
		return true
	default:
		return false // slow client: drop rather than stall the room
	}
}

// SendLossy đẩy gói MỚI vào hàng đợi, bỏ gói cũ nhất nếu đầy (spec mục 27).
//
// Dùng cho snapshot: client mạng yếu chỉ cần bản mới nhất, một bản cũ nằm trong hàng chỉ làm
// nhân vật giật lùi rồi nhảy tới. Khác hẳn `welcome` hay `leaderboard` — mất là mất luôn,
// nên những gói đó vẫn đi qua Send thường.
func (c *wsConn) SendLossy(b []byte) bool {
	for i := 0; i < 2; i++ {
		select {
		case c.out <- b:
			return true
		default:
			select {
			case <-c.out: // vứt gói cũ nhất rồi thử lại
			default:
			}
		}
	}
	return false
}

func (s *server) handleWS(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if !s.limiter.acquire(ip) {
		http.Error(w, "too many connections", http.StatusTooManyRequests)
		return
	}
	defer s.limiter.release(ip)
	ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	ws.SetReadLimit(maxMsgBytes)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	conn := &wsConn{out: make(chan []byte, sendQueue)}
	go writeLoop(ctx, ws, conn.out, cancel)
	// Quyền debug chốt NGAY LÚC BẮT TAY, từ phiên admin trong cookie (ADMIN_SPEC §1.1).
	//
	// `?debug=true` trên URL không còn ý nghĩa gì ở đây — client chỉ được bật phím test khi
	// server nói có, và server nói có khi có phiên admin (hoặc APP_ENV=dev trên localhost).
	src := s.debugGate.Allow(r)
	client := room.NewClient(conn, src != admin.DebugNone)
	client.DebugSource = string(src)
	// Giữ IP thô và User-Agent TRÊN KẾT NỐI, không ghi ra đĩa: §6 cần hiện `113.161.x.x`
	// và cần so khi cấm theo IP, cả hai chỉ có nghĩa khi người đó đang online.
	client.IP, client.UA = ip, r.UserAgent()
	// Kiểm lại mỗi 60 giây: phiên admin có thể hết hạn GIỮA TRẬN. Giữ `r` để đọc lại cookie
	// — request gốc sống đúng bằng kết nối WebSocket nên dùng nó là an toàn.
	client.RecheckDebug = func() (bool, string) {
		s := s.debugGate.Allow(r)
		return s != admin.DebugNone, string(s)
	}
	// Băm IP ngay tại đây, không giữ IP thô ở bất cứ đâu khác (TELEMETRY_SPEC §1.3): nó chỉ
	// dùng để biết "hai phiên này cùng một mạng", không để truy ra ai.
	s.readLoop(ctx, ws, client, telemetry.Hash(ip))
	if rm := client.Room(); rm != nil {
		rm.RemoveClient(client)
	}
	_ = ws.Close(websocket.StatusNormalClosure, "")
}

func writeLoop(ctx context.Context, ws *websocket.Conn, out <-chan []byte, cancel func()) {
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		case b := <-out:
			wctx, done := context.WithTimeout(ctx, writeTimeout)
			err := ws.Write(wctx, websocket.MessageBinary, b)
			done()
			if err != nil {
				return
			}
		}
	}
}

// readLoop decodes client messages and enforces the per-connection rate limit.
func (s *server) readLoop(ctx context.Context, ws *websocket.Conn, c *room.Client, ipHash string) {
	windowStart, count := time.Now(), 0
	for {
		_, data, err := ws.Read(ctx)
		if err != nil {
			return
		}
		if time.Since(windowStart) > time.Second {
			windowStart, count = time.Now(), 0
		}
		if count++; count > s.cfg.Net.MaxMsgsPerSec {
			continue
		}
		var env netproto.Envelope
		if netproto.Decode(data, &env) != nil {
			// Gói vỡ là chuyện bình thường khi mạng tệ; chỉ thành cờ khi nó LẶP LẠI
			// (> 5 lần/phút, §6.1). Ngưỡng nằm trong `NoteBadPacket`.
			c.NoteBadPacket(time.Now())
			continue
		}
		s.dispatch(c, env.T, data, ipHash)
	}
}

func (s *server) dispatch(c *room.Client, t string, data []byte, ipHash string) {
	switch t {
	case netproto.TInput:
		var in netproto.Input
		if netproto.Decode(data, &in) == nil {
			c.SetInput(in)
			c.NoteInput(time.Now(), s.cfg.Net.MaxMsgsPerSec)
		} else {
			c.NoteBadPacket(time.Now())
		}
	case netproto.TJoin:
		s.handleJoin(c, data)
	case netproto.TPing:
		var p netproto.Ping
		if netproto.Decode(data, &p) == nil {
			c.NotePing(time.Now(), p.Client)
			if b, err := netproto.Encode(netproto.Pong{T: "pong", Client: p.Client, Server: float64(time.Now().UnixMilli())}); err == nil {
				c.Conn.Send(b)
			}
		}
	case netproto.TDescRequest:
		// NETCODE_SPEC §5 Hiển thị mục 2: client thiếu mô tả entity thì HỎI, thay vì vẽ
		// placeholder. Không cần ở trong phòng mới hỏi được — chỉ ghi nhận, snapshot kế
		// tiếp của chính client đó sẽ chở mô tả.
		var d netproto.DescRequest
		if netproto.Decode(data, &d) == nil && len(d.IDs) > 0 {
			c.ForgetDesc(d.IDs)
		}
	case netproto.TTelemetry:
		s.handleTelemetry(c, ipHash, data)
	case netproto.TDebug:
		var d netproto.Debug
		// Bỏ lệnh test trong im lặng từng làm hai lượt đo 10 tab ra bảng số sai mà không ai
		// biết vì sao. Giờ nói rõ bỏ vì lý do gì.
		// Từ chối MỌI lệnh debug từ kết nối không có quyền — kể cả khi client bị sửa để
		// gửi thẳng gói (§1.1). Lệnh bị từ chối tính vào cờ nghi hack §6.1.
		switch rm := c.Room(); {
		case !s.debug:
			log.Printf("dbg bỏ qua: server không chạy với -debug")
		case !c.DebugAllowed(time.Now()):
			c.NoteRefused(time.Now())
			s.auditRefusedDebug(c, data)
		case netproto.Decode(data, &d) != nil:
			log.Printf("dbg bỏ qua: gói hỏng (%d byte)", len(data))
		case rm == nil:
			log.Printf("dbg bỏ qua: client %d chưa ở trong phòng nào", c.PlayerID)
		default:
			// Mỗi lệnh debug ĐƯỢC DÙNG đều vào nhật ký (§1.1), và đánh dấu mạng này đã
			// "nhiễm debug" để không ghi vào bảng xếp hạng toàn cục.
			c.DebugUsed = true
			s.auditDebugCmd(c, d)
			rm.Debug(c, d)
		}
	default:
	}
}

func (s *server) handleJoin(c *room.Client, data []byte) {
	if c.Room() != nil {
		return // already playing
	}
	var j netproto.Join
	if netproto.Decode(data, &j) != nil {
		return
	}
	j.Name = s.names.Clean(j.Name, j.Lang)
	if j.Name == "" {
		sendErr(c, "bad_name")
		return
	}
	if !slices.Contains(s.cfg.Characters, j.Char) {
		j.Char = s.cfg.Characters[0]
	}
	if len(j.GuestID) > 64 {
		j.GuestID = j.GuestID[:64]
	}
	c.GuestID = j.GuestID
	c.Lang = config.NormLang(j.Lang)
	// Người bị cấm không vào được (§6.2) — và ĐƯỢC BIẾT vì sao, cùng thời điểm hết hạn.
	//
	// Chặn ở đây chứ không ở lúc bắt tay WebSocket: lúc bắt tay chưa có `guest_id` (nó nằm
	// trong gói `join`), nên chặn sớm hơn thì chỉ chặn được theo IP — và cấm IP một mình
	// thì đụng cả nhà người ta.
	if ban := s.adminAuth.Store().BanOf(c.GuestID, c.IP); ban != nil {
		until := int64(-1) // vĩnh viễn
		if ban.Until != nil {
			until = *ban.Until
		}
		if b, err := netproto.Encode(netproto.Kicked{
			T: netproto.TKicked, Reason: ban.Reason, Until: until,
		}); err == nil {
			c.Conn.Send(b)
		}
		return
	}
	rm := s.mgr.Join(c, j)
	log.Printf("join %q (%s) → room %s", j.Name, j.Char, rm.ID)
}

func sendErr(c *room.Client, msg string) {
	if b, err := netproto.Encode(netproto.Error{T: "error", Msg: msg}); err == nil {
		c.Conn.Send(b)
	}
}
