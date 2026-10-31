// Package admin lo phần quản trị: đăng nhập, phiên, nhật ký thao tác, cấm người chơi,
// cờ nghi hack và hòm thư góp ý (ADMIN_SPEC v1.0).
//
// Tách thành package riêng chứ không nhét vào `cmd/server` vì ba lý do:
//   - nó có BẢNG DỮ LIỆU riêng và vòng đời riêng (phiên hết hạn, cấm hết hạn);
//   - nó là mặt tấn công duy nhất của server có mật khẩu, nên gom một chỗ thì soi được;
//   - `internal/room` không được phép phụ thuộc vào nó — admin NHÌN vào trận, không điều
//     khiển trận (ngoài kick, đi qua một cửa hẹp).
package admin

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Schema dựng mọi bảng của ADMIN_SPEC §8.
//
// Dùng CHUNG file SQLite với bảng xếp hạng (spec §8 nói rõ), nhưng tên bảng đều có tiền tố
// `admin_` hoặc là bảng mới hẳn — không đụng gì tới dữ liệu đã có.
const Schema = `
CREATE TABLE IF NOT EXISTS admin_sessions (
  id         TEXT PRIMARY KEY,   -- SHA-256 của token, KHÔNG lưu token gốc
  csrf       TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  expires_at INTEGER NOT NULL,
  ip         TEXT NOT NULL,
  user_agent TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_admin_sessions_exp ON admin_sessions(expires_at);

CREATE TABLE IF NOT EXISTS admin_audit (
  id      INTEGER PRIMARY KEY AUTOINCREMENT,
  at      INTEGER NOT NULL,
  action  TEXT NOT NULL,
  target  TEXT NOT NULL,
  reason  TEXT NOT NULL,
  ip      TEXT NOT NULL,
  details TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_admin_audit_at ON admin_audit(at DESC);

CREATE TABLE IF NOT EXISTS admin_login_attempts (
  id      INTEGER PRIMARY KEY AUTOINCREMENT,
  at      INTEGER NOT NULL,
  ip      TEXT NOT NULL,
  ok      INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_admin_attempts ON admin_login_attempts(ip, at DESC);
`

// Store là cửa duy nhất vào dữ liệu quản trị.
type Store struct {
	db *sql.DB
}

// New mở store trên một `*sql.DB` đã có và dựng bảng nếu chưa có.
func New(db *sql.DB) (*Store, error) {
	for _, q := range []string{Schema, FeedbackSchema} {
		if _, err := db.Exec(q); err != nil {
			return nil, fmt.Errorf("dựng bảng admin: %w", err)
		}
	}
	return &Store{db: db}, nil
}

// Session là một phiên đăng nhập đang mở.
type Session struct {
	// Token: chuỗi gốc đặt vào cookie. CHỈ có mặt lúc vừa tạo — đọc lại từ DB không bao giờ
	// lấy được, vì DB chỉ giữ băm của nó.
	Token   string
	CSRF    string
	Expires time.Time
}

// randomToken sinh `n` byte ngẫu nhiên dạng hex.
func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hashToken băm token phiên.
//
// Lưu BĂM chứ không lưu token: ai đọc được file SQLite (sao lưu, lỡ tay commit, máy bị
// chiếm) cũng không mạo danh được phiên đang mở. SHA-256 trần là đủ ở đây — token đã là 32
// byte ngẫu nhiên nên không có gì để dò, khác hẳn mật khẩu (thứ phải dùng bcrypt).
func hashToken(tok string) string {
	sum := sha256.Sum256([]byte(tok))
	return hex.EncodeToString(sum[:])
}

// CreateSession mở một phiên mới, hết hạn sau `ttl`.
func (s *Store) CreateSession(ttl time.Duration, ip, ua string) (*Session, error) {
	tok, err := randomToken(32) // spec §1: cookie phiên ngẫu nhiên 32 byte
	if err != nil {
		return nil, err
	}
	csrf, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	exp := now.Add(ttl)
	_, err = s.db.Exec(
		`INSERT INTO admin_sessions (id, csrf, created_at, expires_at, ip, user_agent) VALUES (?,?,?,?,?,?)`,
		hashToken(tok), csrf, now.Unix(), exp.Unix(), ip, ua)
	if err != nil {
		return nil, err
	}
	return &Session{Token: tok, CSRF: csrf, Expires: exp}, nil
}

// LookupSession trả về phiên còn hạn ứng với token, hoặc nil.
func (s *Store) LookupSession(tok string) *Session {
	if tok == "" {
		return nil
	}
	var csrf string
	var exp int64
	err := s.db.QueryRow(`SELECT csrf, expires_at FROM admin_sessions WHERE id = ?`,
		hashToken(tok)).Scan(&csrf, &exp)
	if err != nil || time.Now().Unix() >= exp {
		return nil
	}
	return &Session{Token: tok, CSRF: csrf, Expires: time.Unix(exp, 0)}
}

// DeleteSession xoá một phiên (đăng xuất).
func (s *Store) DeleteSession(tok string) {
	if tok != "" {
		_, _ = s.db.Exec(`DELETE FROM admin_sessions WHERE id = ?`, hashToken(tok))
	}
}

// PurgeExpired dọn phiên và lần-thử-đăng-nhập đã quá hạn.
//
// Dọn lần-thử cũ là BẮT BUỘC, không phải dọn dẹp cho gọn: `RecentFailures` đếm trong cửa sổ
// 15 phút, mà bảng cứ phình mãi thì mỗi lần đăng nhập lại quét thêm một ít. Giữ 7 ngày để
// còn tra được lịch sử dò mật khẩu.
func (s *Store) PurgeExpired() {
	now := time.Now().Unix()
	_, _ = s.db.Exec(`DELETE FROM admin_sessions WHERE expires_at < ?`, now)
	_, _ = s.db.Exec(`DELETE FROM admin_login_attempts WHERE at < ?`, now-7*24*3600)
}

// RecordLogin ghi một lần thử đăng nhập.
func (s *Store) RecordLogin(ip string, ok bool) {
	v := 0
	if ok {
		v = 1
	}
	_, _ = s.db.Exec(`INSERT INTO admin_login_attempts (at, ip, ok) VALUES (?,?,?)`,
		time.Now().Unix(), ip, v)
}

// RecentFailures đếm số lần SAI của một IP trong `window` gần đây.
//
// Chỉ đếm từ lần ĐÚNG gần nhất trở đi: đăng nhập thành công xoá sạch nợ, nếu không thì gõ
// sai bốn lần rồi vào đúng xong vẫn bị khoá ở lần sai tiếp theo — người dùng thật bị phạt
// vì đã từng gõ nhầm.
func (s *Store) RecentFailures(ip string, window time.Duration) int {
	since := time.Now().Add(-window).Unix()
	var lastOK sql.NullInt64
	_ = s.db.QueryRow(`SELECT MAX(at) FROM admin_login_attempts WHERE ip = ? AND ok = 1 AND at >= ?`,
		ip, since).Scan(&lastOK)
	from := since
	if lastOK.Valid && lastOK.Int64 > from {
		from = lastOK.Int64
	}
	var n int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM admin_login_attempts WHERE ip = ? AND ok = 0 AND at > ?`,
		ip, from).Scan(&n)
	return n
}

// AuditEntry là một dòng nhật ký thao tác.
type AuditEntry struct {
	At      int64  `json:"at"`
	Action  string `json:"action"`
	Target  string `json:"target"`
	Reason  string `json:"reason"`
	IP      string `json:"ip"`
	Details string `json:"details"`
}

// Audit ghi một thao tác admin (spec §1: MỌI hành động đều phải ghi).
func (s *Store) Audit(action, target, reason, ip string, details any) {
	js := "{}"
	if details != nil {
		if b, err := json.Marshal(details); err == nil {
			js = string(b)
		}
	}
	_, _ = s.db.Exec(`INSERT INTO admin_audit (at, action, target, reason, ip, details) VALUES (?,?,?,?,?,?)`,
		time.Now().Unix(), action, target, reason, ip, js)
}

// RecentAudit đọc `limit` dòng nhật ký mới nhất.
func (s *Store) RecentAudit(limit int) ([]AuditEntry, error) {
	rows, err := s.db.Query(
		`SELECT at, action, target, reason, ip, details FROM admin_audit ORDER BY at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AuditEntry
	for rows.Next() {
		var e AuditEntry
		if err := rows.Scan(&e.At, &e.Action, &e.Target, &e.Reason, &e.IP, &e.Details); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
