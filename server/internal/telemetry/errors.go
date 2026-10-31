package telemetry

import (
	"strings"
	"time"
)

// Lỗi của người chơi thật, gom lại cho trang quản trị — ADMIN_SPEC §3, TELEMETRY_SPEC §4.
//
// Vì sao cần: một người chơi iPhone báo "luôn bị lỗi" và không có cách nào biết lỗi gì.
// Console của máy họ thì mình không với tới, và bảo họ chụp màn hình console trên Safari
// iOS là điều không xảy ra. Nên client gửi lỗi về, và chỗ này đọc lại.

// ErrorKinds là các loại sự kiện được coi là "lỗi" ở trang quản trị.
//
// `gpu_pressure` và `webgl_context_lost` không phải ngoại lệ JavaScript, nhưng với người
// chơi chúng LÀ lỗi — màn hình trắng hoặc tab bị hệ điều hành thu hồi. Gom chung một chỗ
// vì người xem đang hỏi "có ai gặp chuyện gì không", không hỏi "có exception nào không".
var ErrorKinds = []string{"error", "webgl_context_lost", "gpu_pressure"}

// ErrorGroup là một nhóm lỗi GIỐNG NHAU.
//
// Gom nhóm chứ không liệt kê từng dòng: một lỗi trong vòng vẽ bắn ra 60 lần mỗi giây, và
// một danh sách 3 000 dòng giống hệt nhau thì không đọc được. Thứ cần biết là "có mấy loại
// lỗi, mỗi loại bao nhiêu người dính".
type ErrorGroup struct {
	Kind string `json:"kind"`
	// Sig: chữ ký gộp nhóm — dòng đầu của thông báo, đã bỏ số (xem errorSig).
	Sig      string `json:"sig"`
	Sample   string `json:"sample"`
	Count    int    `json:"count"`
	Sessions int    `json:"sessions"`
	FirstAt  int64  `json:"first_at"`
	LastAt   int64  `json:"last_at"`
	// Devices: danh sách "Trình duyệt / Hệ điều hành" đã gặp, gọn.
	Devices []string `json:"devices"`
	Builds  []string `json:"builds"`
	Mobile  int      `json:"mobile"`
}

// errorSig dựng chữ ký để gộp nhóm.
//
// Bỏ SỐ khỏi chữ ký: tên tệp có hash đổi mỗi lần build (`index-CzSoRTTv.js:896:28935`), nên
// giữ số lại thì cùng một lỗi ở hai bản build thành hai nhóm, và không bao giờ thấy được
// "lỗi này đã có từ tuần trước".
func errorSig(kind, msg string) string {
	// Sự kiện KHÔNG phải ngoại lệ JavaScript (`gpu_pressure`, `webgl_context_lost`) chở
	// JSON SỐ LIỆU, không phải thông báo. Gộp cả loại làm MỘT nhóm.
	//
	// Nếu gộp theo nội dung thì `"q":"Thấp"` và `"q":"Vừa"` thành hai nhóm riêng, và một
	// sự kiện có `data` rỗng thành nhóm thứ ba — ba dòng cho đúng một vấn đề. Đo trên dữ
	// liệu thật ngày 31/10: 2 người chạm trần bộ nhớ hiện thành 2 nhóm + 2 nhóm.
	if kind != "error" {
		return kind
	}
	line := msg
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}
	line = strings.TrimSpace(line)
	if len(line) > 200 {
		line = line[:200]
	}
	var b strings.Builder
	lastDigit := false
	for _, r := range line {
		if r >= '0' && r <= '9' {
			if !lastDigit {
				b.WriteByte('#')
			}
			lastDigit = true
			continue
		}
		lastDigit = false
		b.WriteRune(r)
	}
	sig := b.String()
	if sig == "" {
		sig = kind
	}
	return sig
}

// RecentErrors gom lỗi trong `since` gần đây.
func (s *Store) RecentErrors(since time.Duration, now time.Time) ([]ErrorGroup, error) {
	from := now.Add(-since).UnixMilli()
	rs, err := s.db.Query(`SELECT e.ts, e.kind, e.data, e.session_id,
			COALESCE(x.browser,''), COALESCE(x.os,''), COALESCE(x.mobile,0), COALESCE(x.build,'')
		FROM events e LEFT JOIN sessions x ON x.session_id = e.session_id
		WHERE e.kind IN ('error','webgl_context_lost','gpu_pressure') AND e.ts >= ?
		ORDER BY e.ts DESC LIMIT 5000`, from)
	if err != nil {
		return nil, err
	}
	defer rs.Close()

	type acc struct {
		g        *ErrorGroup
		sessions map[string]bool
		devices  map[string]bool
		builds   map[string]bool
	}
	groups := map[string]*acc{}
	var order []string
	for rs.Next() {
		var ts int64
		var kind, data, sess, browser, os, build string
		var mobile int
		if err := rs.Scan(&ts, &kind, &data, &sess, &browser, &os, &mobile, &build); err != nil {
			return nil, err
		}
		sig := kind + "|" + errorSig(kind, data)
		a := groups[sig]
		if a == nil {
			a = &acc{
				g: &ErrorGroup{Kind: kind, Sig: errorSig(kind, data), Sample: data,
					FirstAt: ts / 1000, LastAt: ts / 1000},
				sessions: map[string]bool{}, devices: map[string]bool{}, builds: map[string]bool{},
			}
			groups[sig] = a
			order = append(order, sig)
		}
		a.g.Count++
		if ts/1000 < a.g.FirstAt {
			a.g.FirstAt = ts / 1000
		}
		if ts/1000 > a.g.LastAt {
			a.g.LastAt = ts / 1000
		}
		if !a.sessions[sess] {
			a.sessions[sess] = true
			a.g.Sessions++
			if mobile != 0 {
				a.g.Mobile++
			}
		}
		if d := shortDevice(browser, os); d != "" {
			a.devices[d] = true
		}
		if build != "" {
			a.builds[build] = true
		}
	}
	out := make([]ErrorGroup, 0, len(order))
	for _, sig := range order {
		a := groups[sig]
		a.g.Devices = keysOf(a.devices)
		a.g.Builds = keysOf(a.builds)
		out = append(out, *a.g)
	}
	// Nhiều NGƯỜI dính lên trước, rồi mới tới nhiều lần.
	//
	// Xếp theo số lần là sai thứ tự ưu tiên: một lỗi bắn 3 000 lần cho đúng một người đang
	// nghịch sẽ che mất một lỗi chỉ xảy ra một lần nhưng với mười người khác nhau.
	for i := 1; i < len(out); i++ {
		for j := i; j > 0; j-- {
			if out[j].Sessions > out[j-1].Sessions ||
				(out[j].Sessions == out[j-1].Sessions && out[j].Count > out[j-1].Count) {
				out[j], out[j-1] = out[j-1], out[j]
				continue
			}
			break
		}
	}
	return out, rs.Err()
}

func shortDevice(browser, os string) string {
	// `browser` từ UA hay là "Version/18.7.5" trên Safari iOS — cắt số bản để gộp nhóm.
	if i := strings.IndexByte(browser, '/'); i > 0 {
		browser = browser[:i]
	}
	if browser == "Version" {
		browser = "Safari"
	}
	if i := strings.IndexByte(os, ' '); i > 0 && strings.HasPrefix(os, "iPhone") {
		os = os[:i]
	}
	switch {
	case browser == "" && os == "":
		return ""
	case browser == "":
		return os
	case os == "":
		return browser
	}
	return browser + " / " + os
}

func keysOf(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}
