package admin

import (
	"database/sql"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

// Checklist ADMIN_SPEC §11 — mỗi mục một bài test.

func testStore(t *testing.T) *Store {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+t.TempDir()+"/a.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	s, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// testAuth dựng Auth với mật khẩu cho trước, KHÔNG qua biến môi trường.
func testAuth(t *testing.T, password string) *Auth {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(password), MinBcryptCost)
	if err != nil {
		t.Fatal(err)
	}
	return &Auth{store: testStore(t), hash: h}
}

// req dựng một request "đến từ localhost" — `secureEnough` cho localhost qua HTTP.
func req() *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/api/admin/login", nil)
	r.RemoteAddr = "127.0.0.1:5555"
	return r
}

// §11: "Sai mật khẩu 5 lần → khoá IP 15 phút."
func TestLoginLocksAfterFiveFailures(t *testing.T) {
	a := testAuth(t, "mat-khau-rat-dai-123")
	for i := 0; i < MaxFailures; i++ {
		if _, err := a.Login(req(), "sai"); err != ErrBadPassword {
			t.Fatalf("lần %d: lỗi %v, muốn ErrBadPassword", i+1, err)
		}
	}
	// Lần thứ 6 phải bị KHOÁ, và **đúng mật khẩu cũng không vào được** — nếu không thì
	// khoá chẳng chặn được ai: kẻ dò chỉ cần đoán trúng là xong.
	if _, err := a.Login(req(), "mat-khau-rat-dai-123"); err != ErrLocked {
		t.Fatalf("sau %d lần sai, đăng nhập đúng vẫn qua được (lỗi %v)", MaxFailures, err)
	}
}

// Đăng nhập ĐÚNG xoá sạch nợ: gõ nhầm vài lần rồi vào được thì không bị phạt về sau.
func TestSuccessClearsFailureCount(t *testing.T) {
	a := testAuth(t, "mat-khau-rat-dai-123")
	for i := 0; i < MaxFailures-1; i++ {
		_, _ = a.Login(req(), "sai")
	}
	if _, err := a.Login(req(), "mat-khau-rat-dai-123"); err != nil {
		t.Fatalf("đăng nhập đúng bị từ chối: %v", err)
	}
	// Sau lần đúng, bộ đếm phải về 0 — còn 4 lần sai nữa mới bị khoá.
	for i := 0; i < MaxFailures-1; i++ {
		if _, err := a.Login(req(), "sai"); err != ErrBadPassword {
			t.Fatalf("lần sai %d sau khi đăng nhập đúng: %v", i+1, err)
		}
	}
}

// Phiên: token hợp lệ tra ra được, token sai thì không, và DB không giữ token gốc.
func TestSessionLookupAndStorage(t *testing.T) {
	st := testStore(t)
	sess, err := st.CreateSession(time.Hour, "1.2.3.4", "ua")
	if err != nil {
		t.Fatal(err)
	}
	if got := st.LookupSession(sess.Token); got == nil || got.CSRF != sess.CSRF {
		t.Fatal("không tra lại được phiên vừa tạo")
	}
	if st.LookupSession("token-bia-dat") != nil {
		t.Error("token bịa mà vẫn tra ra phiên")
	}
	// DB chỉ được giữ BĂM: ai đọc được file sao lưu cũng không mạo danh phiên đang mở.
	var stored string
	if err := st.db.QueryRow(`SELECT id FROM admin_sessions`).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored == sess.Token {
		t.Error("token phiên bị lưu NGUYÊN VĂN trong DB")
	}
	st.DeleteSession(sess.Token)
	if st.LookupSession(sess.Token) != nil {
		t.Error("đăng xuất rồi mà phiên vẫn còn")
	}
}

// Phiên hết hạn thì không dùng được nữa.
func TestSessionExpires(t *testing.T) {
	st := testStore(t)
	sess, err := st.CreateSession(-time.Minute, "1.2.3.4", "ua") // đã hết hạn từ đầu
	if err != nil {
		t.Fatal(err)
	}
	if st.LookupSession(sess.Token) != nil {
		t.Error("phiên hết hạn vẫn tra ra được")
	}
}

// §1: chưa đặt ADMIN_PASSWORD_HASH thì trang quản trị TẮT HẲN, không rơi về mật khẩu mặc
// định. Mật khẩu mặc định là thứ không ai đổi và ai cũng biết.
func TestDisabledWithoutHash(t *testing.T) {
	a := &Auth{store: testStore(t)}
	if a.Enabled() {
		t.Fatal("không có hash mà vẫn báo đang bật")
	}
	if _, err := a.Login(req(), "bat-ky-gi"); err != ErrNotEnabled {
		t.Errorf("lỗi %v, muốn ErrNotEnabled", err)
	}
	if a.SessionOf(httptest.NewRequest(http.MethodGet, "/admin/", nil)) != nil {
		t.Error("trang quản trị tắt mà vẫn trả về phiên")
	}
}

// §1: "Trên HTTP thường thì không cho đăng nhập (trừ localhost để phát triển)."
func TestRejectsPlainHTTPFromRemote(t *testing.T) {
	a := testAuth(t, "mat-khau-rat-dai-123")
	r := httptest.NewRequest(http.MethodPost, "/api/admin/login", nil)
	r.RemoteAddr = "203.0.113.9:4444" // không phải localhost, không TLS
	if _, err := a.Login(r, "mat-khau-rat-dai-123"); err != ErrInsecure {
		t.Fatalf("lỗi %v, muốn ErrInsecure — mật khẩu không được đi qua HTTP trần", err)
	}
	// Qua proxy TLS (Caddy) thì được.
	r2 := httptest.NewRequest(http.MethodPost, "/api/admin/login", nil)
	r2.RemoteAddr = "203.0.113.9:4444"
	r2.Header.Set("X-Forwarded-Proto", "https")
	if _, err := a.Login(r2, "mat-khau-rat-dai-123"); err != nil {
		t.Errorf("sau proxy HTTPS vẫn bị từ chối: %v", err)
	}
}

// §1 "Giới hạn IP (tuỳ chọn)".
func TestIPAllowlist(t *testing.T) {
	a := testAuth(t, "mat-khau-rat-dai-123")
	_, n, _ := net.ParseCIDR("10.0.0.0/8")
	a.allow = append(a.allow, n)

	r := httptest.NewRequest(http.MethodPost, "/api/admin/login", nil)
	r.RemoteAddr = "127.0.0.1:1" // localhost KHÔNG nằm trong 10.0.0.0/8
	if _, err := a.Login(r, "mat-khau-rat-dai-123"); err != ErrNotAllowed {
		t.Fatalf("IP ngoài allowlist vẫn vào được (lỗi %v)", err)
	}
	r2 := httptest.NewRequest(http.MethodPost, "/api/admin/login", nil)
	r2.RemoteAddr = "10.1.2.3:1"
	r2.Header.Set("X-Forwarded-Proto", "https")
	if _, err := a.Login(r2, "mat-khau-rat-dai-123"); err != nil {
		t.Errorf("IP trong allowlist bị chặn: %v", err)
	}
}

// §1: lấy IP thật qua X-Forwarded-For (server chạy sau Caddy). Sai ở đây nghĩa là khoá
// nhầm người, hoặc tệ hơn: khoá chung một IP proxy cho tất cả.
func TestClientIPPrefersForwardedFor(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.1:777"
	r.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")
	if got := ClientIP(r); got != "203.0.113.5" {
		t.Errorf("ClientIP = %q, muốn client gốc 203.0.113.5", got)
	}
	r2 := httptest.NewRequest(http.MethodGet, "/", nil)
	r2.RemoteAddr = "198.51.100.7:888"
	if got := ClientIP(r2); got != "198.51.100.7" {
		t.Errorf("không có XFF thì phải lấy RemoteAddr, được %q", got)
	}
}

// §1: MỌI hành động admin phải vào nhật ký.
func TestAuditRecordsLogins(t *testing.T) {
	a := testAuth(t, "mat-khau-rat-dai-123")
	_, _ = a.Login(req(), "sai")
	_, _ = a.Login(req(), "mat-khau-rat-dai-123")
	rows, err := a.store.RecentAudit(10)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, e := range rows {
		seen[e.Action] = true
	}
	for _, want := range []string{"login_fail", "login_ok"} {
		if !seen[want] {
			t.Errorf("nhật ký thiếu %q (có: %v)", want, seen)
		}
	}
}
