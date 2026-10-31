package room

import (
	"strconv"
	"sync"
	"time"
)

// Cờ nghi hack — ADMIN_SPEC §6.1.
//
// **Chỉ gắn cờ. Không có một dòng nào ở đây ngắt kết nối của ai.** Đó không phải là thiếu
// sót để bổ sung sau: server đã là bên quyết định mọi luật chơi, nên gói tin bịa không
// nặn ra điểm được. Cờ dùng để NHÌN THẤY ai đang thử phá, rồi người thật quyết định —
// vì mọi điều kiện dưới đây đều có cách vô tội để xảy ra:
//
//   - mạng chập chờn làm gói tin vỡ → "Gói tin lỗi";
//   - hai anh em chung nhà cùng chơi → "Nhiều kết nối";
//   - máy yếu, nhấp liên tục vì tưởng treo → "Input bất thường";
//   - gặp đúng lúc trùm rơi đồ → "Tăng điểm bất thường".
//
// Tự kick theo những thứ này là đuổi nhầm người chơi thật, và người bị đuổi nhầm thì không
// quay lại.

// FlagKind là một loại cờ (§6.1).
type FlagKind string

const (
	FlagInput    FlagKind = "input"     // > 2× tần suất input cho phép trong 10 giây
	FlagBadPkt   FlagKind = "goi_loi"   // gói không giải mã được / sai schema > 5 lần/phút
	FlagRefused  FlagKind = "tu_choi"   // hành động bị từ chối > 10 lần/phút
	FlagMultiCon FlagKind = "nhieu_ket" // > 3 kết nối cùng IP hoặc cùng guest_id
	FlagScore    FlagKind = "diem_vot"  // điểm tăng nhanh hơn 5× trung vị cùng cảnh giới
	FlagFakePing FlagKind = "ping_gia"  // pong sai mẫu / thời gian âm
)

// FlagNames là tên tiếng Việt để hiện ở trang quản trị.
var FlagNames = map[FlagKind]string{
	FlagInput:    "Input bất thường",
	FlagBadPkt:   "Gói tin lỗi",
	FlagRefused:  "Hành động bị từ chối",
	FlagMultiCon: "Nhiều kết nối",
	FlagScore:    "Tăng điểm bất thường",
	FlagFakePing: "Ping giả",
}

// Ngưỡng — §6.1. Để ở đây chứ không ở `config/game.json` vì đây không phải số cân bằng
// game: đổi nó không làm trận đấu khác đi, chỉ làm trang quản trị ồn hơn hay im hơn.
const (
	flagInputWindow   = 10 * time.Second
	flagInputFactor   = 2 // lần so với MaxMsgsPerSec
	flagBadPktPerMin  = 5
	flagRefusedPerMin = 10
	flagMultiConn     = 3
	flagScoreFactor   = 5
	// flagTTL: cờ cũ hơn mức này không hiện nữa. Một người gói vỡ mười phút trước rồi mạng
	// ổn lại thì không còn là chuyện đáng nhìn.
	flagTTL = 15 * time.Minute
)

// Flag là một lần gắn cờ, kèm bằng chứng (§6.1 "mỗi cờ lưu bằng chứng").
type Flag struct {
	Kind  FlagKind `json:"kind"`
	Name  string   `json:"name"`
	At    int64    `json:"at"`
	Count int      `json:"count"`
	// Evidence: số liệu LÚC GẮN CỜ, bằng chữ đọc được. Lưu chữ chứ không lưu số thô vì ba
	// tháng sau không ai nhớ `{"n":37,"lim":12}` nghĩa là gì, còn "37 gói/phút, trần 12"
	// thì đọc là hiểu.
	Evidence string `json:"evidence"`
}

// flagCounter đếm sự kiện trong một cửa sổ trượt đơn giản.
type flagCounter struct {
	start time.Time
	n     int
}

// bump cộng một sự kiện; trả về số đếm trong cửa sổ hiện tại.
func (c *flagCounter) bump(now time.Time, window time.Duration) int {
	if now.Sub(c.start) > window {
		c.start, c.n = now, 0
	}
	c.n++
	return c.n
}

// Flags theo dõi cờ của MỘT kết nối.
//
// Có mutex riêng vì nó bị đụng từ hai goroutine: goroutine đọc WebSocket (gói lỗi, input,
// pong) và goroutine phòng (điểm, hành động bị từ chối).
type Flags struct {
	mu      sync.Mutex
	raised  map[FlagKind]*Flag
	input   flagCounter
	badPkt  flagCounter
	refused flagCounter
}

func newFlags() *Flags { return &Flags{raised: map[FlagKind]*Flag{}} }

// raise gắn hoặc cập nhật một cờ.
func (f *Flags) raise(k FlagKind, now time.Time, evidence string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if cur, ok := f.raised[k]; ok {
		cur.Count++
		cur.At = now.Unix()
		cur.Evidence = evidence
		return
	}
	f.raised[k] = &Flag{Kind: k, Name: FlagNames[k], At: now.Unix(), Count: 1, Evidence: evidence}
}

// List trả về các cờ còn hiệu lực, bản sao (người gọi ở goroutine khác).
func (f *Flags) List(now time.Time) []Flag {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []Flag{}
	for k, fl := range f.raised {
		if now.Sub(time.Unix(fl.At, 0)) > flagTTL {
			delete(f.raised, k)
			continue
		}
		out = append(out, *fl)
	}
	return out
}

// Clear bỏ một cờ (admin đánh dấu nhầm — §6.1). Kind rỗng = bỏ hết.
func (f *Flags) Clear(k FlagKind) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if k == "" {
		f.raised = map[FlagKind]*Flag{}
		return
	}
	delete(f.raised, k)
}

// ── Các đường báo sự kiện ────────────────────────────────────────────────────────────────

// NoteInput báo một gói input. `limit` là MaxMsgsPerSec của config.
func (c *Client) NoteInput(now time.Time, limit int) {
	f := c.Flags
	f.mu.Lock()
	n := f.input.bump(now, flagInputWindow)
	f.mu.Unlock()
	max := limit * flagInputFactor * int(flagInputWindow/time.Second)
	if limit > 0 && n > max {
		f.raise(FlagInput, now, fmtEvidence(n, max, "gói input trong 10 giây, trần"))
	}
}

// NoteBadPacket báo một gói không giải mã được hoặc sai schema.
func (c *Client) NoteBadPacket(now time.Time) {
	f := c.Flags
	f.mu.Lock()
	n := f.badPkt.bump(now, time.Minute)
	f.mu.Unlock()
	if n > flagBadPktPerMin {
		f.raise(FlagBadPkt, now, fmtEvidence(n, flagBadPktPerMin, "gói lỗi trong 1 phút, trần"))
	}
}

// NoteRefused báo một hành động bị server từ chối.
func (c *Client) NoteRefused(now time.Time) {
	c.RefusedCmds++
	f := c.Flags
	f.mu.Lock()
	n := f.refused.bump(now, time.Minute)
	f.mu.Unlock()
	if n > flagRefusedPerMin {
		f.raise(FlagRefused, now, fmtEvidence(n, flagRefusedPerMin, "hành động bị từ chối trong 1 phút, trần"))
	}
}

// NoteFakePing báo một pong sai mẫu.
//
// Gọi khi mốc thời gian client gửi lên nằm ở TƯƠNG LAI, hoặc xa quá khứ tới mức vô lý.
// Ping giả là cách kinh điển để xin server bù trừ độ trễ nhiều hơn thực tế.
func (c *Client) NoteFakePing(now time.Time, skewMs float64) {
	c.Flags.raise(FlagFakePing, now, fmtMs(skewMs))
}

// NoteManyConns báo có quá nhiều kết nối cùng IP hoặc cùng guest_id.
func (c *Client) NoteManyConns(now time.Time, n int, what string) {
	if n > flagMultiConn {
		c.Flags.raise(FlagMultiCon, now, fmtEvidence(n, flagMultiConn, "kết nối cùng "+what+", trần"))
	}
}

// NoteScoreSpike báo điểm tăng nhanh bất thường so với trung vị cùng cảnh giới.
func (c *Client) NoteScoreSpike(now time.Time, rate, median float64) {
	if median > 0 && rate > median*flagScoreFactor {
		c.Flags.raise(FlagScore, now,
			fmtRate(rate, median, rate/median))
	}
}

// fmtEvidence dựng câu bằng chứng dạng "37 gói lỗi trong 1 phút, trần 5".
func fmtEvidence(n, limit int, what string) string {
	return itoa(n) + " " + what + " " + itoa(limit)
}

func fmtMs(ms float64) string {
	return "mốc thời gian lệch " + itoa(int(ms)) + " ms so với đồng hồ server"
}

func fmtRate(rate, median, times float64) string {
	return "Uy Danh +" + itoa(int(rate)) + "/phút, trung vị cùng cảnh giới " +
		itoa(int(median)) + "/phút (" + itoa(int(times)) + "×)"
}

func itoa(n int) string { return strconv.Itoa(n) }
