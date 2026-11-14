// Package safe là lớp chống sập của server (GAME_SPEC §13.8).
//
// Nguyên tắc của spec: bắt lỗi ở **ranh giới**, không bọc từng dòng. Ranh giới của server
// này có đúng bốn loại — một tick của khu, một gói tin của một người, một goroutine nền,
// một request HTTP. Gói này cho mỗi loại một hàm bọc, và cả bốn đều làm cùng ba việc:
//
//  1. `recover()` — tiến trình không chết.
//  2. Ghi stack + bối cảnh (khu nào, người nào, tick thứ mấy) ra log VÀ ra ngoài qua
//     `OnPanic`, để trang quản trị "Lỗi" thấy được. Nuốt lỗi im lặng là điều cấm.
//  3. Trả về cho người gọi biết CÓ panic, để họ tự quyết hồi phục (bỏ tick, ngắt người đó).
//
// Vì sao không dùng một `recover()` to ở `main`: recover chỉ cứu được goroutine đang chạy.
// Panic trong goroutine của một khu làm chết CẢ tiến trình dù main có recover hay không —
// đó chính là lý do phải có hàm bọc ở từng chỗ `go`.
package safe

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Info là một lần panic đã bắt được.
type Info struct {
	// Boundary: loại ranh giới, chuỗi CỐ ĐỊNH (`tick`, `goi-tin`, `nen`, `http`).
	// Cố định vì nó nằm trong chữ ký gom nhóm của trang "Lỗi" — nhét tên khu hay id người
	// vào đây là cùng một lỗi tách thành năm nhóm.
	Boundary string
	// Value: giá trị panic, một dòng.
	Value string
	// Detail: bối cảnh (khu, người, tick). KHÔNG vào chữ ký, chỉ để đọc.
	Detail string
	// Stack: vết gọi đã rút gọn (xem shortStack).
	Stack string
}

// Text là dạng chữ gửi lên trang quản trị: dòng đầu là chữ ký, phần sau là bối cảnh.
// FakePrefix: tiền tố của mọi panic GIẢ do lệnh test `dbg panic` bắn ra (checklist §17).
//
// Vì sao cần một hằng số thay vì so chuỗi rải rác: bảng lỗi 24 giờ ngày 13/11 có 18 dòng
// `server_panic` mà cả 18 đều là lỗi giả của `crashtest.mjs` — mỗi lượt deploy lại thêm vài
// dòng nữa. Lẫn vào thống kê thì người trực không phân biệt được "hôm nay server có sập
// không" với "hôm nay đã deploy mấy lượt".
const FakePrefix = "lỗi giả"

// IsFake: panic này do lệnh test bắn ra, không phải sự cố thật.
func IsFake(i Info) bool {
	return strings.HasPrefix(i.Value, FakePrefix)
}

func (i Info) Text() string {
	var b strings.Builder
	b.WriteString(i.Value)
	b.WriteString(" @ ")
	b.WriteString(i.Boundary)
	if i.Detail != "" {
		b.WriteString("\n")
		b.WriteString(i.Detail)
	}
	if i.Stack != "" {
		b.WriteString("\n")
		b.WriteString(i.Stack)
	}
	return b.String()
}

var (
	hookMu sync.RWMutex
	hook   func(Info)
	// count đếm tổng số panic đã bắt — trang /admin/perf và bài test đọc.
	count atomic.Int64
)

// OnPanic cắm nơi nhận mọi panic đã bắt (thường là telemetry → trang "Lỗi").
func OnPanic(fn func(Info)) {
	hookMu.Lock()
	hook = fn
	hookMu.Unlock()
}

// Count là tổng số panic đã bắt từ lúc khởi động.
func Count() int64 { return count.Load() }

func report(i Info) {
	count.Add(1)
	log.Printf("PANIC ĐÃ BẮT [%s] %s | %s\n%s", i.Boundary, i.Value, i.Detail, i.Stack)
	hookMu.RLock()
	fn := hook
	hookMu.RUnlock()
	if fn != nil {
		// Người nhận cũng có thể panic (DB đóng giữa chừng…). Panic trong bộ xử lý panic là
		// cách chắc chắn nhất để mất cả tiến trình ở đúng lúc tệ nhất.
		func() {
			defer func() { _ = recover() }()
			fn(i)
		}()
	}
}

// Do chạy fn và bắt panic. Trả về true nếu CÓ panic (người gọi tự quyết hồi phục).
//
// `detail` là hàm chứ không phải chuỗi: Do chạy 30 lần/giây cho mỗi khu, dựng chuỗi bối
// cảnh mỗi tick là tự tạo ra rác để đi đo rác.
func Do(boundary string, detail func() string, fn func()) (panicked bool) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		panicked = true
		d := ""
		if detail != nil {
			d = safeDetail(detail)
		}
		report(Info{Boundary: boundary, Value: oneLine(r), Detail: d, Stack: shortStack(3)})
	}()
	fn()
	return false
}

// Go chạy fn trong goroutine mới có recover (spec §13.8: "mọi goroutine có recover()").
func Go(boundary string, detail func() string, fn func()) {
	go func() {
		Do(boundary, detail, fn)
	}()
}

// HTTP là middleware recover cho mọi handler (spec §13.8: "HTTP handler có middleware recover").
func HTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// `http.ErrAbortHandler` là cách CHÍNH THỨC của net/http để bỏ một phản hồi giữa
		// chừng (client ngắt kết nối khi đang tải). Nó không phải lỗi, và ghi nó vào trang
		// "Lỗi" thì mỗi lần ai đó tắt tab giữa lúc tải art lại thành một báo động giả.
		defer func() {
			r2 := recover()
			if r2 == nil {
				return
			}
			if err, ok := r2.(error); ok && err == http.ErrAbortHandler {
				panic(r2) // trả về cho net/http xử lý như thường
			}
			report(Info{
				Boundary: "http",
				Value:    oneLine(r2),
				Detail:   r.Method + " " + r.URL.Path,
				Stack:    shortStack(3),
			})
			// Phản hồi 500 chỉ gửi được khi chưa ghi gì ra. Ghi rồi thì header đã đi, và
			// `WriteHeader` lần hai chỉ in một dòng cảnh báo vô nghĩa vào log.
			defer func() { _ = recover() }()
			http.Error(w, "lỗi máy chủ", http.StatusInternalServerError)
		}()
		next.ServeHTTP(w, r)
	})
}

func safeDetail(fn func() string) (s string) {
	defer func() {
		if r := recover(); r != nil {
			s = "(không lấy được bối cảnh: " + oneLine(r) + ")"
		}
	}()
	return fn()
}

func oneLine(v any) string {
	s := fmt.Sprint(v)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	s = strings.TrimSpace(s)
	if len(s) > 200 {
		s = s[:200]
	}
	if s == "" {
		s = "panic rỗng"
	}
	return s
}

// shortStack lấy 8 khung gọi đầu TÍNH TỪ chỗ panic, bỏ phần runtime và phần của gói này.
//
// Vết đầy đủ của Go dài hàng trăm dòng và phần trên cùng luôn là `runtime.gopanic` —
// không nói gì cả. Tám khung của mã mình thì đủ để biết lỗi ở đâu, và ngắn đủ để lọt vào
// một ô trên trang quản trị.
func shortStack(skip int) string {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(skip, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	var out []string
	for len(out) < 8 {
		f, more := frames.Next()
		if f.Function == "" && !more {
			break
		}
		short := shortFunc(f.Function)
		if !strings.HasPrefix(f.Function, "runtime.") && !isWrapper(short) {
			out = append(out, fmt.Sprintf("%s (%s:%d)", short, shortFile(f.File), f.Line))
		}
		if !more {
			break
		}
	}
	return strings.Join(out, "\n")
}

// isWrapper: khung gọi của chính lớp bọc này — bỏ khỏi vết gọi vì nó có ở MỌI vết và chỉ
// đẩy dòng đáng đọc xuống dưới.
//
// Lọc theo đúng tên hàm chứ không theo tiền tố gói: lọc cả gói thì test của chính gói này
// (cùng package) cũng bị cắt, và vết gọi còn mỗi `testing.tRunner`.
func isWrapper(short string) bool {
	for _, p := range []string{"safe.Do", "safe.Go", "safe.HTTP", "safe.report", "safe.shortStack", "safe.safeDetail"} {
		if strings.HasPrefix(short, p) {
			return true
		}
	}
	return false
}

func shortFunc(name string) string {
	if i := strings.LastIndexByte(name, '/'); i >= 0 {
		return name[i+1:]
	}
	return name
}

func shortFile(path string) string {
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}
	return path
}

// Rate đếm sự kiện trong một cửa sổ trượt — dùng cho luật "≥ 3 lần/phút thì khu suy giảm".
//
// Giữ mốc thời gian thay vì một bộ đếm reset theo phút: bộ đếm reset bỏ lọt đúng trường hợp
// đáng lo nhất (2 lần cuối phút này + 2 lần đầu phút sau = 4 lần trong 10 giây, mà không
// phút nào chạm 3).
type Rate struct {
	mu     sync.Mutex
	window time.Duration
	limit  int
	at     []time.Time
}

// NewRate tạo bộ đếm: quá `limit` lần trong `window` thì `Hit` trả true.
func NewRate(limit int, window time.Duration) *Rate {
	return &Rate{limit: limit, window: window}
}

// Hit ghi một lần và trả về true khi đã chạm ngưỡng trong cửa sổ.
func (r *Rate) Hit(now time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	cut := now.Add(-r.window)
	keep := r.at[:0]
	for _, t := range r.at {
		if t.After(cut) {
			keep = append(keep, t)
		}
	}
	r.at = append(keep, now)
	return len(r.at) >= r.limit
}

// Count là số lần trong cửa sổ hiện tại.
func (r *Rate) Count(now time.Time) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	cut := now.Add(-r.window)
	n := 0
	for _, t := range r.at {
		if t.After(cut) {
			n++
		}
	}
	return n
}
