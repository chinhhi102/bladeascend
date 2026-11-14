package room

import (
	"crypto/rand"
	"encoding/hex"
	"log"

	"bladeascend/server/internal/game"
	"bladeascend/server/internal/netproto"
)

// Vào lại phiên khi F5 — GAME_SPEC §7.1 (v6.53).
//
// ── Vấn đề ───────────────────────────────────────────────────────────────────────────────
//
// Mất kết nối giữa trận (F5, mạng rớt, chuyển app, mất WebGL rồi tải lại) từng là **mất
// luôn ván**: `RemoveClient` xoá nhân vật khỏi thế giới ngay lập tức. Người chơi tu luyện
// 20 phút lên Nguyên Anh, lỡ tay F5, và quay lại thành Luyện Khí 1 kiếm.
//
// ── Cách làm ─────────────────────────────────────────────────────────────────────────────
//
// Nhân vật **Ở LẠI thế giới 20 giây** sau khi mất kết nối: đứng yên, vòng kiếm vẫn quay,
// **vẫn chém chết được**. Đây là điểm quan trọng nhất của spec và là lý do không chọn cách
// đơn giản hơn (đóng băng / bất tử / cất ra ngoài): nếu thoát mạng mà an toàn thì rút dây
// mạng thành một kỹ năng né chết.
//
// Client giữ **mã phiên 32 byte** trong `sessionStorage`. Nối lại kèm mã → server gắn lại
// đúng nhân vật đó (cảnh giới, kiếm, Uy Danh, lượt kỹ năng, khu, buff) và cấp **mã mới**;
// mã cũ hết hiệu lực ngay, nên một mã bị lộ chỉ dùng được một lần.
//
// Quá 20 giây hoặc nhân vật đã chết trong lúc mất kết nối → ván kết thúc như thoát thường;
// người chơi thấy màn Tổng kết (chết) hoặc trang chủ (hết hạn).

const (
	// resumeKeyBytes: 32 byte ngẫu nhiên như spec §7.1 ghi (64 ký tự hex trên dây).
	resumeKeyBytes = 32
	// HoldSecs: nhân vật ở lại thế giới bao lâu sau khi mất kết nối (§7.1: 20 s).
	HoldSecs = 20.0
	// deadKeepSecs: giữ màn Tổng kết của người chết-lúc-mất-kết-nối bao lâu.
	//
	// Dài hơn 20 s của `HoldSecs` vì nó phục vụ việc khác: người bị chém đúng lúc đang tải
	// lại trang phải thấy mình CHẾT, chứ không phải về trang chủ như hết hạn thường. Tải
	// lại trang trên điện thoại 4G mất tới nửa phút.
	deadKeepSecs = 120.0
)

// newResumeKey sinh mã phiên. Dùng crypto/rand: đoán được mã là chiếm được nhân vật người
// khác, nên đây là bí mật thật, không phải một id cho vui.
func newResumeKey() string {
	b := make([]byte, resumeKeyBytes)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand hỏng nghĩa là hệ điều hành hỏng. Không có đường lùi an toàn nào —
		// trả chuỗi rỗng để `resumeOf` không bao giờ khớp, tức là mất tính năng vào lại
		// chứ không phải cấp một mã đoán được.
		log.Printf("§7.1: không sinh được mã phiên: %v", err)
		return ""
	}
	return hex.EncodeToString(b)
}

// hold là một nhân vật đang chờ chủ nối lại (hoặc đã chết và chờ người ta quay lại để xem
// màn Tổng kết).
type hold struct {
	key      string
	playerID uint32
	// left: `z.G.Time` của khu lúc mất kết nối. Dùng đồng hồ của KHU chứ không dùng
	// `time.Now()` vì mọi thứ khác trong tick đều đo bằng đồng hồ đó.
	left float64
	// dead: màn Tổng kết đã dựng sẵn khi nhân vật chết trong lúc chủ đang mất kết nối.
	// nil = chưa chết.
	dead *netproto.Dead
	// deadAt: `z.G.Time` lúc chết, để biết khi nào quên màn Tổng kết đó đi.
	deadAt float64
	// debugUsed / lang: chép từ kết nối đã mất, vì `record` đọc chúng từ `z.clients` —
	// mà lúc hết hạn chờ thì kết nối không còn ở đó nữa (xem Room.record).
	debugUsed bool
	lang      string
}

// resumeKeyOf là mã phiên của một kết nối đang chơi (rỗng nếu chưa vào trận).
//
// Giữ trên `Client` chứ không trên `game.Player`: nó là bí mật của KẾT NỐI, không phải
// thuộc tính của nhân vật, và `game` không được biết gì về mạng.

// attachResume ghi mã phiên mới cho một kết nối vừa vào trận / vừa nối lại.
func (r *Room) attachResume(c *Client, key string) {
	c.ResumeKey = key
	if key == "" {
		return
	}
	r.mgr.rememberResume(key, r)
}

// beginHold giữ nhân vật lại trong thế giới thay vì xoá ngay (§7.1).
//
// Trả về false khi không giữ được (không có mã phiên, nhân vật đã chết, hoặc phòng đang
// đóng) — lúc đó người gọi phải dọn như thoát thường.
func (r *Room) beginHold(z *Zone, c *Client) bool {
	if c.ResumeKey == "" || r.closed {
		return false
	}
	p, ok := z.G.Players[c.PlayerID]
	if !ok {
		return false
	}
	// Giữ chỗ KỂ CẢ khi nhân vật đang chết.
	//
	// Nhân vật chết nhưng hoạt ảnh chết chưa chạy xong (§7: 1 giây) thì `handleDeaths`
	// chưa dựng màn Tổng kết. Nếu ở đây từ chối giữ chỗ, kết nối bị dọn luôn và màn Tổng
	// kết đi vào hư vô — người chơi tải lại trang và thấy TRANG CHỦ, không hiểu ván vừa
	// rồi đi đâu. Giữ chỗ thì một giây sau `handleDeaths` thấy `z.holds[id]` và cất màn
	// Tổng kết vào đó (xem noteHeldDeath). Bắt được bằng `tools/browsertest/rejoin.mjs`:
	// bấm chết rồi F5 ngay trong 120 ms.
	// Đứng yên NGAY: input cuối cùng của người chơi có thể là "đang chạy", và một nhân vật
	// vô chủ chạy thẳng 20 giây vào góc bản đồ thì nối lại cũng chẳng còn gì để cứu.
	p.Vel = game.Vec{}
	p.Moving, p.Burning = false, false
	p.Offline = true
	h := &hold{key: c.ResumeKey, playerID: c.PlayerID, left: z.G.Time, debugUsed: c.DebugUsed, lang: c.Lang}
	z.holds[c.PlayerID] = h
	r.mgr.holdResume(c.ResumeKey, r, c.PlayerID)
	delete(z.clients, c.PlayerID)
	c.PlayerID = 0
	c.setRoom(nil)
	return true
}

// expireHolds dọn những nhân vật quá hạn chờ (§7.1: 20 giây) và những màn Tổng kết đã cũ.
//
// Chạy ở RANH GIỚI TICK (sau khi mọi khu xong) để được phép đụng vào nhiều khu cùng lúc.
func (r *Room) expireHolds() {
	for _, z := range r.zones {
		for id, h := range z.holds {
			switch {
			case h.dead != nil:
				// Đã chết: nhân vật không còn trong thế giới, chỉ còn màn Tổng kết chờ chủ
				// quay lại xem. Giữ lâu hơn 20 s (xem deadKeepSecs).
				if z.G.Time-h.deadAt >= deadKeepSecs {
					r.dropHold(z, id, h)
				}
			case z.G.Time-h.left >= HoldSecs:
				// Hết 20 giây: rời phòng như thoát thường — ghi lại mạng chơi rồi xoá.
				if p, ok := z.G.Players[id]; ok && !p.Dead {
					r.record(z, p)
				}
				z.G.RemovePlayer(id)
				r.dropHold(z, id, h)
			}
		}
	}
}

// dropHold xoá một chỗ giữ khỏi khu, khỏi sổ của Manager, và trả lại suất người của phòng.
func (r *Room) dropHold(z *Zone, id uint32, h *hold) {
	delete(z.holds, id)
	r.mgr.forgetResume(h.key)
	r.mgr.release(r)
}

// holdOf tìm chỗ giữ theo mã phiên trong cả phòng.
func (r *Room) holdOf(key string) (*Zone, *hold) {
	for _, z := range r.zones {
		for _, h := range z.holds {
			if h.key == key {
				return z, h
			}
		}
	}
	return nil, nil
}

// liveClientOf tìm kết nối ĐANG CHƠI mang mã phiên này (trường hợp mở tab thứ hai).
func (r *Room) liveClientOf(key string) (*Zone, *Client) {
	for _, z := range r.zones {
		for _, c := range z.clients {
			if c.ResumeKey == key {
				return z, c
			}
		}
	}
	return nil, nil
}

// resumeHuman gắn một kết nối mới vào nhân vật cũ (§7.1). Trả false nếu không gắn được —
// người gọi vào trận mới như người lạ.
func (r *Room) resumeHuman(c *Client, j netproto.Join, key string) bool {
	// Tab thứ hai với CÙNG mã phiên: tab mới nhận nhân vật, tab cũ bị đẩy ra với lời nhắn
	// riêng (§7.1 dòng cuối). Không làm thế thì hai tab cùng gửi input cho một nhân vật và
	// nó giật qua giật lại — một lỗi không ai đoán ra khi chỉ nhìn màn hình.
	if z, old := r.liveClientOf(key); old != nil {
		old.send(netproto.Kicked{T: netproto.TKicked, Reason: netproto.KickOtherTab, Until: 0})
		h := &hold{key: key, playerID: old.PlayerID, left: z.G.Time, debugUsed: old.DebugUsed, lang: old.Lang}
		z.holds[old.PlayerID] = h
		delete(z.clients, old.PlayerID)
		old.PlayerID = 0
		old.setRoom(nil)
		// Suất người: tab cũ nhả, tab mới giữ — tổng không đổi, nên không gọi release.
	}
	z, h := r.holdOf(key)
	if z == nil {
		return false
	}
	// Chết trong lúc mất kết nối → KHÔNG vào lại trận; gửi thẳng màn Tổng kết (§7.1).
	if h.dead != nil {
		c.send(*h.dead)
		delete(z.holds, h.playerID)
		r.mgr.forgetResume(h.key)
		r.mgr.release(r)
		return true
	}
	p, ok := z.G.Players[h.playerID]
	// Chết nhưng `handleDeaths` CHƯA kịp cất màn Tổng kết.
	//
	// `handleDeaths` chỉ xử lý sau khi hết hoạt ảnh chết (`death_anim_secs`, đang là 1 giây),
	// nên có một khe: bị chém lúc 0,0 giây, nối lại lúc 0,5 giây thì `h.dead` còn `nil` và
	// nhánh dưới bỏ luôn cái hold — người chơi rơi vào một TRẬN MỚI, mất cả ván mà không
	// hiểu vì sao. Đúng điều §7.1 cấm. Ngày 14/11 bài `rejoin.mjs` trượt một lượt deploy ở
	// đúng khe này, và chỉ phân biệt được với "chém hụt" nhờ dòng log của `killid`.
	//
	// Ở đây làm đúng việc `handleDeaths` sẽ làm: ghi sổ ván, gửi màn Tổng kết, bỏ hold.
	if ok && p.Dead {
		d := r.deadMsg(z, p)
		r.record(z, p)
		r.noteHeldDeath(z, p, d) // cũng gỡ nhân vật khỏi thế giới
		c.send(d)
		delete(z.holds, h.playerID)
		r.mgr.forgetResume(h.key)
		r.mgr.release(r)
		return true
	}
	if !ok {
		r.dropHold(z, h.playerID, h)
		return false
	}
	delete(z.holds, h.playerID)
	p.Offline = false
	// Tên và nhân vật giữ nguyên của ván đang chơi: cho phép đổi lúc nối lại là một đường
	// đổi tên giữa trận mà không ai kiểm.
	_ = j
	c.Debug = c.Debug && (r.mgr == nil || r.mgr.AllowDebug)
	// Vết "đã dùng lệnh test" theo NHÂN VẬT qua lần nối lại: cùng một ván, cùng số điểm
	// đã nặn ra. Nối lại mà sạch vết là một đường rửa điểm bằng F5.
	c.DebugUsed = c.DebugUsed || h.debugUsed
	if c.Lang == "" {
		c.Lang = h.lang
	}
	c.PlayerID = p.ID
	c.resetKnown()
	z.clients[p.ID] = c
	c.setRoom(r)
	// Mã MỚI, mã cũ hết hiệu lực ngay (§7.1).
	r.mgr.forgetResume(key)
	r.attachResume(c, newResumeKey())
	w := r.welcomeFor(z, p, c)
	w.Resumed = true
	w.Session = c.ResumeKey
	c.send(w)
	log.Printf("§7.1: %q vào lại đúng nhân vật (phòng %s, khu %s, cảnh giới %d, %d kiếm)",
		p.Name, r.ID, z.Name, p.Level, p.Ring.Count)
	return true
}

// noteHeldDeath cất màn Tổng kết của một nhân vật chết TRONG LÚC chủ đang mất kết nối.
//
// Spec §7.1: "bị chém trong lúc mất kết nối → chết, F5 xong thấy Tổng kết". Không có chỗ
// này thì người đó quay lại và thấy trang chủ, không hiểu vì sao mất cả ván.
func (r *Room) noteHeldDeath(z *Zone, p *game.Player, d netproto.Dead) bool {
	h, ok := z.holds[p.ID]
	if !ok || h.dead != nil {
		return false
	}
	cp := d
	h.dead = &cp
	h.deadAt = z.G.Time
	z.G.RemovePlayer(p.ID)
	return true
}

// HeldCount là số nhân vật đang chờ chủ nối lại — trang quản trị và `checkEmpty` cần nó.
func (r *Room) HeldCount() int {
	n := 0
	for _, z := range r.zones {
		n += len(z.holds)
	}
	return n
}

// HoldAgeSecs là tuổi của một chỗ giữ, cho test đọc.
func (z *Zone) holdAge(id uint32) float64 {
	h, ok := z.holds[id]
	if !ok {
		return -1
	}
	return z.G.Time - h.left
}
