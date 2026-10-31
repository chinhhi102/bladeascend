package admin

import (
	"crypto/subtle"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// Cờ debug gắn với đăng nhập admin — ADMIN_SPEC v1.1 §1.1.
//
// Trước đây `?debug=true` trên URL là đủ để bật phím test (+EXP, lên cảnh giới, +kiếm,
// buff, thả bot…). Nghĩa là **bất kỳ ai cũng tự lên Đạo Tổ được** chỉ bằng cách sửa địa
// chỉ — và bảng xếp hạng toàn cục nhận cả những điểm đó.
//
// Giờ URL một mình KHÔNG BAO GIỜ đủ. Server mới là bên quyết định, qua ba cửa:
//   1. có phiên admin hợp lệ (cookie gửi kèm lúc mở WebSocket);
//   2. `APP_ENV=dev` VÀ kết nối từ localhost — để làm việc nhanh lúc phát triển;
//   3. `DEBUG_TOKEN` khớp, nhưng CHỈ khi (2) cũng đúng về môi trường — cho công cụ đo
//      (`loadtest`, `browsertest`, `netcode-bench`).
//
// Production không bao giờ nhận (2) và (3): `APP_ENV` không phải "dev" thì hai cửa đó đóng
// hẳn, kể cả khi ai đó đặt `DEBUG_TOKEN` trong môi trường.

// DebugGate quyết định một kết nối có được dùng lệnh debug hay không.
type DebugGate struct {
	auth *Auth
	// dev: APP_ENV=dev. Chỉ đọc một lần lúc khởi động — đổi biến môi trường giữa chừng
	// không được phép đổi quyền của các kết nối đang mở.
	dev   bool
	token string
}

// NewDebugGate đọc môi trường và dựng cửa.
func NewDebugGate(auth *Auth) *DebugGate {
	return &DebugGate{
		auth:  auth,
		dev:   strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "dev"),
		token: strings.TrimSpace(os.Getenv("DEBUG_TOKEN")),
	}
}

// DebugSource nói quyền debug đến TỪ ĐÂU — để ghi nhật ký và để gỡ lỗi khi ai đó hỏi
// "sao máy tôi không bật được debug".
type DebugSource string

const (
	DebugNone    DebugSource = ""
	DebugAdmin   DebugSource = "admin"      // có phiên đăng nhập
	DebugDev     DebugSource = "dev"        // APP_ENV=dev trên localhost
	DebugToken   DebugSource = "token"      // DEBUG_TOKEN, chỉ hợp lệ ở dev/localhost
	DebugRefused DebugSource = "bi_tu_choi" // chỉ dùng trong nhật ký
)

// Allow kiểm một request (lúc mở WebSocket hoặc lúc kiểm lại định kỳ).
func (g *DebugGate) Allow(r *http.Request) DebugSource {
	if g == nil {
		return DebugNone
	}
	if g.auth.SessionOf(r) != nil {
		return DebugAdmin
	}
	// Hai cửa còn lại CHỈ mở ở môi trường phát triển. `localEnv` xét cả `APP_ENV` lẫn địa
	// chỉ gọi tới: đặt APP_ENV=dev trên máy chủ thật mà vẫn mở cho cả Internet thì cửa này
	// còn tệ hơn cái nó thay thế.
	if !g.localEnv(r) {
		return DebugNone
	}
	if g.token != "" && r != nil {
		// So sánh theo thời gian hằng: token so bằng `==` rò rỉ độ dài khớp qua thời gian.
		got := r.URL.Query().Get("debug_token")
		if got == "" {
			got = r.Header.Get("X-Debug-Token")
		}
		if subtle.ConstantTimeCompare([]byte(got), []byte(g.token)) == 1 {
			return DebugToken
		}
	}
	if g.dev {
		return DebugDev
	}
	return DebugNone
}

// localEnv: `APP_ENV=dev` và kết nối đến từ máy cục bộ.
func (g *DebugGate) localEnv(r *http.Request) bool {
	if !g.dev || r == nil {
		return false
	}
	ip := net.ParseIP(ClientIP(r))
	return ip != nil && ip.IsLoopback()
}

// RecheckEvery: §1.1 "Phiên admin hết hạn giữa trận → tắt quyền debug ngay ở lần kiểm tiếp
// theo (kiểm lại mỗi 60s)".
const RecheckEvery = 60 * time.Second
