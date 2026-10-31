package admin

import (
	"errors"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Đăng nhập và kiểm soát truy cập — ADMIN_SPEC §1.

const (
	// CookieName: tên cookie phiên.
	CookieName = "kx_admin"
	// SessionTTL: 12 giờ (spec §1).
	SessionTTL = 12 * time.Hour
	// LockWindow / MaxFailures: 5 lần sai trong 15 phút thì khoá IP 15 phút (spec §1).
	LockWindow  = 15 * time.Minute
	MaxFailures = 5
	// MinBcryptCost: spec đòi cost ≥ 12. Hash yếu hơn bị TỪ CHỐI ngay lúc khởi động chứ
	// không âm thầm chấp nhận — một hash cost 4 trông y hệt hash cost 12 trong biến môi
	// trường, và không ai phát hiện ra cho tới khi bị dò trúng.
	MinBcryptCost = 12
)

// Errors mà `Login` có thể trả về.
var (
	ErrLocked      = errors.New("IP đang bị khoá vì thử sai quá nhiều")
	ErrBadPassword = errors.New("sai mật khẩu")
	ErrNotEnabled  = errors.New("chưa đặt ADMIN_PASSWORD_HASH nên trang quản trị bị tắt")
	ErrInsecure    = errors.New("trang quản trị chỉ cho đăng nhập qua HTTPS (hoặc localhost)")
	ErrNotAllowed  = errors.New("IP không nằm trong ADMIN_IP_ALLOWLIST")
)

// Auth gói mọi quyết định "ai được vào".
type Auth struct {
	store *Store
	hash  []byte
	// allow: danh sách IP/CIDR được phép, nil = cho tất cả (spec §1 "tuỳ chọn").
	allow []*net.IPNet
}

// NewAuth đọc cấu hình từ biến môi trường.
//
// Mật khẩu KHÔNG nằm trong code hay config (spec §1) — chỉ `ADMIN_PASSWORD_HASH`. Không đặt
// biến đó thì trang quản trị TẮT HẲN, chứ không rơi về một mật khẩu mặc định: mật khẩu mặc
// định là thứ không ai đổi và ai cũng biết.
func NewAuth(store *Store) (*Auth, error) {
	h := strings.TrimSpace(os.Getenv("ADMIN_PASSWORD_HASH"))
	if h == "" {
		return &Auth{store: store}, nil // tắt, không phải lỗi
	}
	cost, err := bcrypt.Cost([]byte(h))
	if err != nil {
		return nil, errors.New("ADMIN_PASSWORD_HASH không phải hash bcrypt hợp lệ")
	}
	if cost < MinBcryptCost {
		return nil, errors.New("ADMIN_PASSWORD_HASH có cost quá thấp; spec đòi ≥ 12")
	}
	a := &Auth{store: store, hash: []byte(h)}
	for _, part := range strings.Split(os.Getenv("ADMIN_IP_ALLOWLIST"), ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if !strings.Contains(part, "/") {
			if ip := net.ParseIP(part); ip != nil {
				bits := 32
				if ip.To4() == nil {
					bits = 128
				}
				a.allow = append(a.allow, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
				continue
			}
		}
		if _, n, err := net.ParseCIDR(part); err == nil {
			a.allow = append(a.allow, n)
		}
	}
	return a, nil
}

// Enabled: trang quản trị có bật không.
func (a *Auth) Enabled() bool { return a != nil && len(a.hash) > 0 }

// ClientIP lấy IP thật của request.
//
// Đọc `X-Forwarded-For` vì server chạy sau Caddy; lấy phần tử ĐẦU TIÊN (client gốc). Nếu
// không có proxy thì `RemoteAddr`. Dùng cho khoá-theo-IP và cho nhật ký, nên sai ở đây
// nghĩa là khoá nhầm người hoặc không khoá được ai.
func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// ipAllowed: IP này có nằm trong danh sách cho phép không (rỗng = cho tất cả).
func (a *Auth) ipAllowed(ip string) bool {
	if len(a.allow) == 0 {
		return true
	}
	p := net.ParseIP(ip)
	if p == nil {
		return false
	}
	for _, n := range a.allow {
		if n.Contains(p) {
			return true
		}
	}
	return false
}

// secureEnough: kết nối có đủ an toàn để gửi mật khẩu không (spec §1).
//
// Cho localhost qua để còn phát triển được; mọi địa chỉ khác bắt buộc HTTPS. Nhận ra HTTPS
// qua `r.TLS` hoặc `X-Forwarded-Proto` (Caddy kết thúc TLS rồi mới chuyển tiếp).
func secureEnough(r *http.Request) bool {
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		return true
	}
	ip := net.ParseIP(ClientIP(r))
	return ip != nil && ip.IsLoopback()
}

// Login kiểm mật khẩu và mở phiên. Trả về phiên hoặc lỗi.
func (a *Auth) Login(r *http.Request, password string) (*Session, error) {
	if !a.Enabled() {
		return nil, ErrNotEnabled
	}
	ip := ClientIP(r)
	if !a.ipAllowed(ip) {
		a.store.Audit("login_blocked_ip", ip, "ngoài allowlist", ip, nil)
		return nil, ErrNotAllowed
	}
	if !secureEnough(r) {
		return nil, ErrInsecure
	}
	if a.store.RecentFailures(ip, LockWindow) >= MaxFailures {
		// Ghi cả lần bị khoá: không ghi thì nhật ký chỉ thấy 5 lần sai rồi im, và không
		// phân biệt được "bỏ cuộc" với "vẫn đang dò nhưng bị chặn".
		a.store.Audit("login_locked", ip, "quá số lần thử", ip, nil)
		return nil, ErrLocked
	}
	if err := bcrypt.CompareHashAndPassword(a.hash, []byte(password)); err != nil {
		a.store.RecordLogin(ip, false)
		a.store.Audit("login_fail", ip, "sai mật khẩu", ip, nil)
		return nil, ErrBadPassword
	}
	a.store.RecordLogin(ip, true)
	sess, err := a.store.CreateSession(SessionTTL, ip, r.UserAgent())
	if err != nil {
		return nil, err
	}
	a.store.Audit("login_ok", ip, "", ip, nil)
	return sess, nil
}

// SetCookie đặt cookie phiên theo đúng thuộc tính spec §1 đòi.
func SetCookie(w http.ResponseWriter, r *http.Request, sess *Session) {
	http.SetCookie(w, &http.Cookie{
		Name:  CookieName,
		Value: sess.Token,
		Path:  "/",
		// `Secure` chỉ bật khi thật sự chạy HTTPS: bật trên localhost HTTP thì trình duyệt
		// vứt cookie đi và không ai đăng nhập nổi lúc phát triển.
		Secure:   r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https"),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  sess.Expires,
	})
}

// ClearCookie xoá cookie phiên.
func ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: CookieName, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, SameSite: http.SameSiteStrictMode,
	})
}

// SessionOf lấy phiên hợp lệ của request, nil nếu chưa đăng nhập.
func (a *Auth) SessionOf(r *http.Request) *Session {
	if !a.Enabled() {
		return nil
	}
	c, err := r.Cookie(CookieName)
	if err != nil {
		return nil
	}
	if !a.ipAllowed(ClientIP(r)) {
		return nil
	}
	return a.store.LookupSession(c.Value)
}

// Logout xoá phiên hiện tại.
func (a *Auth) Logout(r *http.Request) {
	if c, err := r.Cookie(CookieName); err == nil {
		a.store.DeleteSession(c.Value)
		a.store.Audit("logout", ClientIP(r), "", ClientIP(r), nil)
	}
}

// Store trả về store bên dưới (cho handler ghi nhật ký).
func (a *Auth) Store() *Store { return a.store }
