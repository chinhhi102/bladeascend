package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"kiemxoay/server/internal/admin"
	"kiemxoay/server/internal/netproto"
	"kiemxoay/server/internal/room"
)

// Trang quản trị — ADMIN_SPEC §1, §9.
//
// Luật then chốt của §1: **HTML và JS của trang admin chỉ phục vụ sau khi đăng nhập.** Đó
// không phải thừa thãi bên cạnh việc chặn API: mã client để công khai là tặng kẻ tấn công
// bản đồ đầy đủ của mọi endpoint, mọi tên trường, mọi luồng thao tác. Nên `adminAssets`
// nằm SAU cùng một cửa như API.

// adminRoutes gắn mọi đường dẫn quản trị.
func (s *server) adminRoutes(mux *http.ServeMux, adminDir string) {
	// Công khai: chỉ đúng trang đăng nhập và API đăng nhập.
	mux.HandleFunc("GET /admin/login", s.handleAdminLoginPage)
	mux.HandleFunc("POST /api/admin/login", s.handleAdminLogin)
	// Sau đăng nhập.
	//
	// Đăng xuất cũng nằm sau cửa, không phải ở khối công khai: nó XOÁ phiên, tức là một
	// request thay đổi dữ liệu, và §1 đòi mọi request như thế phải mang `X-Admin-Token`.
	// Để ngoài thì một trang bất kỳ POST sang đây là đá admin ra khỏi phiên đang làm việc.
	mux.HandleFunc("POST /api/admin/logout", s.adminGuard(s.handleAdminLogout))
	mux.HandleFunc("GET /api/admin/me", s.adminGuard(s.handleAdminMe))
	mux.HandleFunc("GET /api/admin/audit", s.adminGuard(s.handleAdminAudit))
	mux.HandleFunc("GET /api/admin/stats", s.adminGuard(s.handleAdminStats))
	mux.HandleFunc("GET /api/admin/feedback", s.adminGuard(s.handleAdminFeedback))
	mux.HandleFunc("GET /api/admin/feedback.csv", s.adminGuard(s.handleAdminFeedbackCSV))
	mux.HandleFunc("PATCH /api/admin/feedback/{id}", s.adminGuard(s.handleAdminFeedbackPatch))
	mux.HandleFunc("GET /api/admin/players", s.adminGuard(s.handleAdminPlayers))
	mux.HandleFunc("POST /api/admin/kick", s.adminGuard(s.handleAdminKick))
	mux.HandleFunc("GET /api/admin/bans", s.adminGuard(s.handleAdminBans))
	mux.HandleFunc("POST /api/admin/unban", s.adminGuard(s.handleAdminUnban))
	mux.HandleFunc("POST /api/admin/clearflag", s.adminGuard(s.handleAdminClearFlag))
	mux.HandleFunc("GET /api/admin/lives", s.adminGuard(s.handleAdminLives))
	mux.HandleFunc("GET /api/admin/hidden", s.adminGuard(s.handleAdminHidden))
	mux.HandleFunc("POST /api/admin/hide", s.adminGuard(s.handleAdminHide))
	mux.HandleFunc("GET /api/admin/feedback/{id}/image", s.adminGuard(s.handleAdminFeedbackImage))
	// CÔNG KHAI: người chơi gửi góp ý (§9). Có rate limit trong `AddFeedback`.
	mux.HandleFunc("POST /api/feedback", s.handleFeedback)
	mux.Handle("GET /admin/", s.adminPageGuard(adminDir))
}

// adminGuard chặn mọi handler API phía sau nó.
//
// Trả 401 kèm JSON chứ không chuyển hướng: đây là API, và một trang JS nhận 302 về HTML
// login sẽ cố parse HTML thành JSON rồi báo một lỗi chẳng liên quan gì.
func (s *server) adminGuard(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := s.adminAuth.SessionOf(r)
		if sess == nil {
			adminJSON(w, http.StatusUnauthorized, map[string]string{"error": "chưa đăng nhập"})
			return
		}
		// CSRF: mọi request THAY ĐỔI dữ liệu phải mang token lấy từ `/api/admin/me`
		// (spec §1). `SameSite=Strict` đã chặn gần hết, nhưng nó là thuộc tính của trình
		// duyệt — token là thứ server tự kiểm được.
		if r.Method != http.MethodGet && r.Header.Get("X-Admin-Token") != sess.CSRF {
			adminJSON(w, http.StatusForbidden, map[string]string{"error": "thiếu hoặc sai X-Admin-Token"})
			return
		}
		h(w, r)
	}
}

// adminPageGuard phục vụ HTML/JS của trang quản trị, CHỈ khi đã đăng nhập.
func (s *server) adminPageGuard(dir string) http.Handler {
	fs := http.StripPrefix("/admin/", http.FileServer(http.Dir(dir)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.adminAuth.SessionOf(r) == nil {
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}
		// Trang quản trị KHÔNG được cache: nó chứa dữ liệu người chơi thật, và một bản lưu
		// trong cache trình duyệt vẫn đọc được sau khi đăng xuất.
		w.Header().Set("Cache-Control", "no-store")
		if r.URL.Path == "/admin/" || r.URL.Path == "/admin" {
			http.ServeFile(w, r, filepath.Join(dir, "index.html"))
			return
		}
		fs.ServeHTTP(w, r)
	})
}

func adminJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// handleAdminLoginPage phục vụ trang đăng nhập — trang quản trị DUY NHẤT công khai.
func (s *server) handleAdminLoginPage(w http.ResponseWriter, r *http.Request) {
	if !s.adminAuth.Enabled() {
		http.Error(w, "Trang quản trị chưa bật: chưa đặt ADMIN_PASSWORD_HASH.", http.StatusNotFound)
		return
	}
	if s.adminAuth.SessionOf(r) != nil {
		http.Redirect(w, r, "/admin/", http.StatusFound)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(adminLoginHTML)
}

func (s *server) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&body); err != nil {
		adminJSON(w, http.StatusBadRequest, map[string]string{"error": "nội dung không hợp lệ"})
		return
	}
	sess, err := s.adminAuth.Login(r, body.Password)
	if err != nil {
		// Mã lỗi phân biệt "bị khoá" với "sai mật khẩu" để người dùng thật biết phải chờ;
		// nhưng KHÔNG nói "tài khoản không tồn tại" hay gì khác giúp thu hẹp phạm vi dò.
		code := http.StatusUnauthorized
		if err == admin.ErrLocked {
			code = http.StatusTooManyRequests
		}
		adminJSON(w, code, map[string]string{"error": err.Error()})
		return
	}
	admin.SetCookie(w, r, sess)
	adminJSON(w, http.StatusOK, map[string]string{"ok": "1"})
}

func (s *server) handleAdminLogout(w http.ResponseWriter, r *http.Request) {
	s.adminAuth.Logout(r)
	admin.ClearCookie(w)
	adminJSON(w, http.StatusOK, map[string]string{"ok": "1"})
}

// handleAdminMe trả CSRF token cho phiên hiện tại (spec §9).
func (s *server) handleAdminMe(w http.ResponseWriter, r *http.Request) {
	sess := s.adminAuth.SessionOf(r)
	adminJSON(w, http.StatusOK, map[string]any{
		"csrf":    sess.CSRF,
		"expires": sess.Expires.Unix(),
		"build":   s.buildID,
	})
}

func (s *server) handleAdminAudit(w http.ResponseWriter, r *http.Request) {
	rows, err := s.adminAuth.Store().RecentAudit(200)
	if err != nil {
		adminJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if rows == nil {
		rows = []admin.AuditEntry{}
	}
	adminJSON(w, http.StatusOK, rows)
}

// adminDirOf tìm thư mục chứa trang quản trị đã build.
func adminDirOf(clientDir string) string {
	d := filepath.Join(clientDir, "admin")
	if _, err := os.Stat(d); err == nil {
		return d
	}
	return d // vẫn trả về: handler sẽ 404 từng tệp, không làm sập server
}

// adminLoginHTML là trang đăng nhập, nhúng thẳng vào server.
//
// Nhúng chứ không lấy từ `client/dist`: trang này phải phục vụ được NGAY cả khi thư mục
// client chưa build xong, và nó là thứ duy nhất của khu quản trị được phép công khai — giữ
// nó ở đây thì không bao giờ có chuyện một tệp admin khác vô tình nằm cạnh và lộ theo.
var adminLoginHTML = []byte(`<!doctype html>
<html lang="vi"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Quản trị — Kiếm Xoay Tu Tiên</title>
<style>
 :root{color-scheme:dark}
 body{margin:0;min-height:100vh;display:grid;place-items:center;background:#15171c;
      color:#e8eaed;font:16px/1.5 system-ui,sans-serif}
 form{background:#1e2129;padding:28px;border-radius:12px;width:min(92vw,340px);
      box-shadow:0 8px 32px #0008}
 h1{margin:0 0 4px;font-size:19px}
 p{margin:0 0 18px;color:#9aa0a6;font-size:13px}
 input{width:100%;box-sizing:border-box;padding:11px 12px;border-radius:8px;
       border:1px solid #3a3f49;background:#14161b;color:#e8eaed;font-size:15px}
 button{width:100%;margin-top:12px;padding:11px;border:0;border-radius:8px;
        background:#4c7dff;color:#fff;font-size:15px;font-weight:600;cursor:pointer}
 button:disabled{opacity:.6;cursor:default}
 .err{margin-top:12px;color:#ff8f8f;font-size:13px;min-height:1.5em}
</style></head><body>
<form id="f">
 <h1>Trang quản trị</h1>
 <p>Chỉ chủ dự án.</p>
 <input id="pw" type="password" autocomplete="current-password" placeholder="Mật khẩu" autofocus>
 <button id="go" type="submit">Đăng nhập</button>
 <div class="err" id="err"></div>
</form>
<script>
const f=document.getElementById('f'),pw=document.getElementById('pw'),
      err=document.getElementById('err'),go=document.getElementById('go');
f.addEventListener('submit',async e=>{
  e.preventDefault(); err.textContent=''; go.disabled=true;
  try{
    const r=await fetch('/api/admin/login',{method:'POST',
      headers:{'Content-Type':'application/json'},
      body:JSON.stringify({password:pw.value})});
    if(r.ok){location.href='/admin/';return;}
    const j=await r.json().catch(()=>({}));
    // textContent, không innerHTML: thông báo lỗi đi từ server về, và nguyên tắc §1 là
    // KHÔNG BAO GIỜ dựng HTML từ dữ liệu động ở khu quản trị.
    err.textContent=j.error||('Lỗi '+r.status);
  }catch(_){err.textContent='Không gọi được server.';}
  go.disabled=false; pw.select();
});
</script></body></html>
`)

// adminEnabledNote in ra lúc khởi động để chủ dự án biết trang quản trị bật hay tắt.
func adminEnabledNote(a *admin.Auth) string {
	if a.Enabled() {
		return "trang quản trị: BẬT (/admin)"
	}
	return "trang quản trị: tắt (đặt ADMIN_PASSWORD_HASH để bật — `go run ./cmd/adminpass`)"
}

// sanitizeReason cắt lý do thao tác về độ dài hợp lý và bỏ ký tự điều khiển.
//
// Lý do do admin gõ nên không phải dữ liệu thù địch, nhưng nó đi thẳng vào nhật ký rồi hiển
// thị lại — giữ nó là text thuần ngay từ cửa vào thì chỗ hiển thị không phải lo.
func sanitizeReason(s string) string {
	s = strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\t' {
			return -1
		}
		return r
	}, s)
	if len(s) > 500 {
		s = s[:500]
	}
	return strings.TrimSpace(s)
}

// auditDebugCmd ghi một lệnh debug ĐƯỢC DÙNG vào nhật ký (ADMIN_SPEC §1.1).
//
// Ghi đủ ngữ cảnh để sau này đọc lại biết chuyện gì đã xảy ra: lệnh gì, ai, phòng/khu nào,
// và quyền debug đến từ đâu. Thiếu nguồn quyền thì không phân biệt được "chủ dự án test"
// với "ai đó mượn được phiên admin".
func (s *server) auditDebugCmd(c *room.Client, d netproto.Debug) {
	zone := ""
	if rm := c.Room(); rm != nil {
		zone = rm.ID
	}
	s.adminAuth.Store().Audit("debug_cmd", c.GuestID, d.Cmd, "", map[string]any{
		"cmd": d.Cmd, "n": d.N, "player": c.PlayerID, "room": zone, "source": c.DebugSource,
	})
}

// auditRefusedDebug ghi một lệnh debug BỊ TỪ CHỐI.
//
// Đây là tín hiệu nghi hack chứ không phải lỗi vận hành: client bình thường không bao giờ
// gửi gói debug khi server chưa cấp quyền — nó chỉ gửi sau khi nhận `debug: true` trong
// Welcome. Gói debug không có quyền nghĩa là ai đó đang gửi thủ công.
func (s *server) auditRefusedDebug(c *room.Client, data []byte) {
	var d netproto.Debug
	cmd := "(gói hỏng)"
	if netproto.Decode(data, &d) == nil {
		cmd = d.Cmd
	}
	log.Printf("dbg TỪ CHỐI: client %d không có quyền debug (lệnh %q, lần thứ %d)",
		c.PlayerID, cmd, c.RefusedCmds)
	s.adminAuth.Store().Audit("debug_refused", c.GuestID, cmd, "không có quyền debug",
		map[string]any{"cmd": cmd, "player": c.PlayerID, "refused_total": c.RefusedCmds})
}
