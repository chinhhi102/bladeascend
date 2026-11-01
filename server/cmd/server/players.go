package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"bladeascend/server/internal/admin"
	"bladeascend/server/internal/room"
)

// Bảng người chơi, cờ nghi hack, kick và cấm — ADMIN_SPEC §6.
//
// **Không có một dòng nào trong tệp này chạy tự động.** Mọi thứ ở đây đều do một người bấm
// nút, với lý do nhập bằng tay, và để lại dấu trong `admin_audit`. §6.1 nói thẳng "chỉ gắn
// cờ, không tự kick" — nên không có goroutine nào quét cờ rồi hành động, và sẽ không có.

// playerRow là một dòng của bảng §6.
type playerRow struct {
	GuestID string `json:"guest_id"`
	Name    string `json:"name"`
	Char    string `json:"character"`
	// Device/IP: §6 đòi che bớt IP. Che ở SERVER, không ở trình duyệt — gửi IP đầy đủ xuống
	// rồi mới che bằng JavaScript thì nó vẫn nằm trong tab mở sẵn của bất cứ ai mượn máy.
	Device string `json:"device"`
	IPMask string `json:"ip_mask"`

	Room    string `json:"room"`
	Zone    string `json:"zone"`
	Realm   string `json:"realm"`
	Level   int    `json:"level"`
	Swords  int    `json:"swords"`
	Score   float64 `json:"score"`
	Hits    int    `json:"hits"`
	Alive   float64 `json:"alive_secs"`

	PingMs      float64 `json:"ping_ms"`
	FPS         float64 `json:"fps"`
	Corrections int     `json:"corrections"`

	Flags []room.Flag `json:"flags"`
	// DebugSource: kết nối này có quyền lệnh test không, và từ đâu. Trống = không có.
	DebugSource string `json:"debug_source"`
	DebugUsed   bool   `json:"debug_used"`
}

// maskIP che hai nhóm cuối của IPv4 (§6: `113.161.x.x`).
//
// IPv6 che từ nhóm thứ ba trở đi. Che theo ký tự thì một địa chỉ ngắn có thể lộ gần hết,
// nên cắt theo dấu phân cách.
func maskIP(ip string) string {
	if ip == "" {
		return ""
	}
	if strings.Contains(ip, ":") {
		p := strings.Split(ip, ":")
		if len(p) > 2 {
			return p[0] + ":" + p[1] + ":x:x"
		}
		return "x:x"
	}
	p := strings.Split(ip, ".")
	if len(p) == 4 {
		return p[0] + "." + p[1] + ".x.x"
	}
	return "x.x.x.x"
}

// deviceOf rút gọn User-Agent thành "Trình duyệt · Hệ điều hành".
//
// Rút gọn ở server để trang quản trị không phải chở một chuỗi 200 ký tự cho mỗi người, và
// để cái nhìn đầu tiên đọc được ngay — "Safari · iPhone" nói nhiều hơn "Mozilla/5.0 (iPhone;
// CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15…".
func deviceOf(ua string) string {
	if ua == "" {
		return "?"
	}
	var br, os string
	switch {
	// Thứ tự quan trọng: Edge và Chrome đều tự xưng là Safari, Chrome cũng tự xưng Edge
	// không có. Kiểm từ cái hiếm tới cái phổ biến.
	case strings.Contains(ua, "Edg/"):
		br = "Edge"
	case strings.Contains(ua, "OPR/"):
		br = "Opera"
	case strings.Contains(ua, "Firefox/"):
		br = "Firefox"
	case strings.Contains(ua, "Chrome/"):
		br = "Chrome"
	case strings.Contains(ua, "Safari/"):
		br = "Safari"
	default:
		br = "?"
	}
	switch {
	case strings.Contains(ua, "iPhone"):
		os = "iPhone"
	case strings.Contains(ua, "iPad"):
		os = "iPad"
	case strings.Contains(ua, "Android"):
		os = "Android"
	case strings.Contains(ua, "Windows"):
		os = "Windows"
	case strings.Contains(ua, "Mac OS X"):
		os = "Mac"
	case strings.Contains(ua, "Linux"):
		os = "Linux"
	default:
		os = "?"
	}
	return br + " · " + os
}

// handleAdminPlayers trả về mọi người thật đang online (§6).
func (s *server) handleAdminPlayers(w http.ResponseWriter, _ *http.Request) {
	now := time.Now()
	rows := []playerRow{}
	// Đếm kết nối theo IP và theo guest_id TRƯỚC, để gắn cờ "Nhiều kết nối" (§6.1).
	byIP, byGuest := map[string]int{}, map[string]int{}
	type entry struct {
		c  *room.Client
		rw playerRow
	}
	var all []entry

	// Gom TRÊN goroutine của từng phòng: duyệt `z.clients` từ goroutine HTTP là lỗi
	// "concurrent map iteration and map write", và Go dừng hẳn chương trình khi gặp nó.
	// Xem `room/query.go`.
	for _, r := range s.mgr.AllRooms() {
		r.Query(func() {
			for _, c := range r.Clients() {
				byIP[c.IP]++
				byGuest[c.GuestID]++
				all = append(all, entry{c: c, rw: s.playerRow(r, c, now)})
			}
		})
	}
	for _, e := range all {
		e.c.NoteManyConns(now, byIP[e.c.IP], "IP")
		e.c.NoteManyConns(now, byGuest[e.c.GuestID], "guest_id")
		e.rw.Flags = e.c.Flags.List(now)
		rows = append(rows, e.rw)
	}
	// Người có cờ lên đầu — đó là lý do mở bảng này.
	sort.SliceStable(rows, func(i, j int) bool { return len(rows[i].Flags) > len(rows[j].Flags) })
	adminJSON(w, http.StatusOK, rows)
}

func (s *server) playerRow(r *room.Room, c *room.Client, now time.Time) playerRow {
	rtt, fps, corr := c.Net()
	row := playerRow{
		GuestID: c.GuestID, Room: r.ID,
		Device: deviceOf(c.UA), IPMask: maskIP(c.IP),
		PingMs: rtt, FPS: fps, Corrections: corr,
		Alive:       now.Sub(c.JoinedAt).Seconds(),
		DebugSource: c.DebugSource, DebugUsed: c.DebugUsed,
		Flags:       []room.Flag{},
	}
	if z := r.ZoneOf(c.PlayerID); z != nil {
		row.Zone = z.Name
		if p := z.G.Players[c.PlayerID]; p != nil {
			row.Name, row.Char, row.Level = p.Name, p.Char, p.Level
			row.Swords, row.Score, row.Hits = p.Ring.Count, p.Score, p.Hits
			row.Realm = z.G.Prog.RealmName(z.G.Prog.Realm(p.Level))
		}
	}
	return row
}

// ── Kick và cấm (§6.2) ───────────────────────────────────────────────────────────────────

// kickBody là yêu cầu kick / cấm. `Reason` BẮT BUỘC — xem `ErrNeedReason`.
type kickBody struct {
	GuestID string `json:"guest_id"`
	Reason  string `json:"reason"`
	// BanSecs: 0 = chỉ kick, không cấm. -1 = cấm vĩnh viễn.
	BanSecs int64 `json:"ban_secs"`
}

func (s *server) handleAdminKick(w http.ResponseWriter, r *http.Request) {
	var b kickBody
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&b); err != nil {
		adminJSON(w, http.StatusBadRequest, map[string]string{"error": "nội dung không hợp lệ"})
		return
	}
	b.Reason = strings.TrimSpace(b.Reason)
	if b.GuestID == "" {
		adminJSON(w, http.StatusBadRequest, map[string]string{"error": "thiếu guest_id"})
		return
	}
	// Lý do bắt buộc cho CẢ kick lẫn cấm. Người bị mời ra sẽ ĐỌC được lý do này — một màn
	// hình "Bạn đã bị mời ra khỏi phòng" không kèm lý do chỉ tạo ra một người giận mà không
	// hiểu vì sao, và người đó sẽ vào lại ngay.
	if b.Reason == "" {
		adminJSON(w, http.StatusBadRequest, map[string]string{"error": "phải nêu lý do"})
		return
	}
	ip := admin.ClientIP(r)
	st := s.adminAuth.Store()

	// Tìm IP của người đó TRƯỚC khi kick: kick xong là kết nối mất, và cùng với nó là IP.
	var targetIP string
	for _, rm := range s.mgr.AllRooms() {
		rm.Query(func() {
			for _, c := range rm.Clients() {
				if c.GuestID == b.GuestID {
					targetIP = c.IP
				}
			}
		})
	}
	act := "kick"
	if b.BanSecs != 0 {
		act = "kick_ban"
		var until *int64
		if b.BanSecs > 0 {
			t := time.Now().Add(time.Duration(b.BanSecs) * time.Second).Unix()
			until = &t
		}
		if err := st.Ban(b.GuestID, targetIP, b.Reason, until, ip); err != nil {
			adminJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}
	n := s.mgr.Kick(b.GuestID, b.Reason)
	st.Audit(act, b.GuestID, b.Reason, ip,
		map[string]any{"ban_secs": b.BanSecs, "connections": n, "ip": maskIP(targetIP)})
	adminJSON(w, http.StatusOK, map[string]any{"ok": 1, "kicked": n})
}

func (s *server) handleAdminBans(w http.ResponseWriter, _ *http.Request) {
	rows, err := s.adminAuth.Store().ListBans()
	if err != nil {
		adminJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	adminJSON(w, http.StatusOK, rows)
}

func (s *server) handleAdminUnban(w http.ResponseWriter, r *http.Request) {
	var b struct {
		GuestID string `json:"guest_id"`
		Reason  string `json:"reason"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&b); err != nil || b.GuestID == "" {
		adminJSON(w, http.StatusBadRequest, map[string]string{"error": "thiếu guest_id"})
		return
	}
	if err := s.adminAuth.Store().Unban(b.GuestID); err != nil {
		adminJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.adminAuth.Store().Audit("unban", b.GuestID, b.Reason, admin.ClientIP(r), nil)
	adminJSON(w, http.StatusOK, map[string]int{"ok": 1})
}

// handleAdminClearFlag bỏ một cờ nghi hack (§6.1 "admin có thể bỏ cờ — đánh dấu nhầm").
func (s *server) handleAdminClearFlag(w http.ResponseWriter, r *http.Request) {
	var b struct {
		GuestID string `json:"guest_id"`
		Kind    string `json:"kind"`
		// Reason: lý do admin gõ ở hộp xác nhận (có thể trống → ghi loại cờ như trước).
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&b); err != nil || b.GuestID == "" {
		adminJSON(w, http.StatusBadRequest, map[string]string{"error": "thiếu guest_id"})
		return
	}
	n := 0
	for _, rm := range s.mgr.AllRooms() {
		rm.Query(func() {
			for _, c := range rm.Clients() {
				if c.GuestID == b.GuestID {
					c.Flags.Clear(room.FlagKind(b.Kind))
					n++
				}
			}
		})
	}
	reason := sanitizeReason(b.Reason)
	if reason == "" {
		reason = b.Kind
	}
	s.adminAuth.Store().Audit("clear_flag", b.GuestID, reason, admin.ClientIP(r), map[string]any{"kind": b.Kind})
	adminJSON(w, http.StatusOK, map[string]int{"ok": 1, "cleared": n})
}
