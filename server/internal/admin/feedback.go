package admin

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Hòm thư góp ý — ADMIN_SPEC §7, §8.
//
// Đây là đường dữ liệu DUY NHẤT trong cả server đi từ người chơi vào trang quản trị, nên
// nó cũng là đường duy nhất kẻ xấu có thể dùng để tấn công chính chủ dự án. Ba lớp phòng:
//   - nội dung luôn là TEXT: trang admin dựng DOM bằng `textContent`, không `innerHTML`;
//   - ảnh kiểm bằng CHỮ KÝ BYTE, không tin đuôi file hay `Content-Type`;
//   - ảnh lưu NGOÀI thư mục phục vụ công khai, chỉ đọc được sau khi đăng nhập.

// FeedbackSchema — spec §8.
const FeedbackSchema = `
CREATE TABLE IF NOT EXISTS feedback (
  id        INTEGER PRIMARY KEY AUTOINCREMENT,
  at        INTEGER NOT NULL,
  guest_id  TEXT NOT NULL,
  name      TEXT NOT NULL,
  kind      TEXT NOT NULL,
  body      TEXT NOT NULL,
  context   TEXT NOT NULL,
  image     TEXT NOT NULL,
  status    TEXT NOT NULL DEFAULT 'moi',
  note      TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_feedback_at ON feedback(at DESC);
CREATE INDEX IF NOT EXISTS idx_feedback_status ON feedback(status);

CREATE TABLE IF NOT EXISTS bans (
  id        INTEGER PRIMARY KEY AUTOINCREMENT,
  guest_id  TEXT NOT NULL DEFAULT '',
  ip        TEXT NOT NULL DEFAULT '',
  reason    TEXT NOT NULL,
  from_at   INTEGER NOT NULL,
  until_at  INTEGER,            -- NULL = vĩnh viễn
  by_ip     TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_bans_guest ON bans(guest_id);
CREATE INDEX IF NOT EXISTS idx_bans_ip ON bans(ip);
`

// Giới hạn nội dung — §7.1.
const (
	FeedbackMinLen   = 5
	FeedbackMaxLen   = 1000
	FeedbackMaxImage = 300 * 1024 // 300 KB
	// FeedbackPerWindow / FeedbackWindow: 3 góp ý / 10 phút / guest_id (§7.1).
	FeedbackPerWindow = 3
	FeedbackWindow    = 10 * time.Minute
)

// FeedbackKinds — §7.1. Thứ tự này đi vào giao diện, nên thêm loại mới thì thêm vào CUỐI.
var FeedbackKinds = []string{"bug", "balance", "idea", "other"}

func validKind(k string) bool {
	for _, v := range FeedbackKinds {
		if v == k {
			return true
		}
	}
	return false
}

// Feedback là một lá thư góp ý.
type Feedback struct {
	ID       int64  `json:"id"`
	At       int64  `json:"at"`
	GuestID  string `json:"guest_id"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Body     string `json:"body"`
	Context  string `json:"context"`
	HasImage bool   `json:"has_image"`
	Status   string `json:"status"`
	Note     string `json:"note"`
}

// FeedbackStatuses — §7.2.
var FeedbackStatuses = []string{"moi", "da_xem", "dang_lam", "xong", "bo_qua"}

func validStatus(s string) bool {
	for _, v := range FeedbackStatuses {
		if v == s {
			return true
		}
	}
	return false
}

var (
	ErrFeedbackTooShort = errors.New("nội dung phải dài ít nhất 5 ký tự")
	ErrFeedbackKind     = errors.New("loại góp ý không hợp lệ")
	ErrFeedbackRate     = errors.New("gửi quá nhanh, thử lại sau vài phút")
	ErrFeedbackBanned   = errors.New("tài khoản đang bị cấm")
	ErrNotAnImage       = errors.New("tệp đính kèm không phải ảnh")
)

// cleanText chuẩn hoá chuỗi do NGƯỜI CHƠI nhập.
//
// Bỏ ký tự điều khiển (trừ xuống dòng và tab) và cắt độ dài. KHÔNG escape HTML ở đây:
// escape lúc lưu là sai chỗ — dữ liệu trong DB nên là thứ người ta gõ, còn an toàn là việc
// của chỗ HIỂN THỊ (trang admin dùng `textContent`). Escape hai lần cho ra `&amp;lt;` và
// người đọc thấy rác thay vì nội dung thật.
func cleanText(s string, max int) string {
	s = strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, s)
	s = strings.TrimSpace(s)
	if len([]rune(s)) > max {
		s = string([]rune(s)[:max])
	}
	return s
}

// imageExt nhận dạng ảnh bằng CHỮ KÝ BYTE, trả về đuôi file hoặc "".
//
// Không tin đuôi file hay `Content-Type` do client gửi (§8 "Kiểm định dạng ảnh thật"): cả
// hai đều do kẻ gửi đặt. Một tệp tên `a.webp` chứa HTML mà server lưu rồi phục vụ lại là
// một lỗ XSS mở thẳng vào trang quản trị.
func imageExt(b []byte) string {
	switch {
	case len(b) >= 12 && bytes.Equal(b[0:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WEBP")):
		return ".webp"
	case len(b) >= 8 && bytes.Equal(b[0:8], []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}):
		return ".png"
	case len(b) >= 3 && b[0] == 0xFF && b[1] == 0xD8 && b[2] == 0xFF:
		return ".jpg"
	default:
		return ""
	}
}

// LooksLikeImage: byte này có đúng là ảnh không (kiểm lại lúc PHỤC VỤ, không chỉ lúc nhận).
func LooksLikeImage(b []byte) bool { return imageExt(b) != "" }

// AddFeedback lưu một lá thư. `img` có thể nil.
func (s *Store) AddFeedback(dir, guestID, name, kind, body string, ctx any, img []byte) (int64, error) {
	if !validKind(kind) {
		return 0, ErrFeedbackKind
	}
	body = cleanText(body, FeedbackMaxLen)
	if len([]rune(body)) < FeedbackMinLen {
		return 0, ErrFeedbackTooShort
	}
	name = cleanText(name, 32)
	guestID = cleanText(guestID, 64)
	if guestID == "" {
		return 0, errors.New("thiếu guest_id")
	}
	if n, _ := s.feedbackCount(guestID, FeedbackWindow); n >= FeedbackPerWindow {
		return 0, ErrFeedbackRate
	}
	ctxJSON := "{}"
	if ctx != nil {
		if b, err := json.Marshal(ctx); err == nil && len(b) < 16*1024 {
			ctxJSON = string(b)
		}
	}
	rel := ""
	if len(img) > 0 {
		ext := imageExt(img)
		if ext == "" {
			return 0, ErrNotAnImage
		}
		if len(img) > FeedbackMaxImage {
			return 0, fmt.Errorf("ảnh quá %d KB", FeedbackMaxImage/1024)
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return 0, err
		}
		// Tên tệp do SERVER đặt hoàn toàn, không lấy một ký tự nào từ client — tên do client
		// gửi là đường vào của `../../` và của mọi trò với đường dẫn.
		rel = fmt.Sprintf("fb_%d_%d%s", time.Now().UnixNano(), len(img), ext)
		if err := os.WriteFile(filepath.Join(dir, rel), img, 0o644); err != nil {
			return 0, err
		}
	}
	res, err := s.db.Exec(
		`INSERT INTO feedback (at, guest_id, name, kind, body, context, image) VALUES (?,?,?,?,?,?,?)`,
		time.Now().Unix(), guestID, name, kind, body, ctxJSON, rel)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) feedbackCount(guestID string, window time.Duration) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM feedback WHERE guest_id = ? AND at >= ?`,
		guestID, time.Now().Add(-window).Unix()).Scan(&n)
	return n, err
}

// ListFeedback đọc danh sách, mới nhất trước. `kind`/`status` rỗng = không lọc.
func (s *Store) ListFeedback(kind, status, q string, limit int) ([]Feedback, error) {
	sb := strings.Builder{}
	sb.WriteString(`SELECT id, at, guest_id, name, kind, body, context, image, status, note FROM feedback WHERE 1=1`)
	var args []any
	if kind != "" {
		sb.WriteString(` AND kind = ?`)
		args = append(args, kind)
	}
	if status != "" {
		sb.WriteString(` AND status = ?`)
		args = append(args, status)
	}
	if q != "" {
		// LIKE với tham số buộc — không bao giờ ghép chuỗi vào câu SQL.
		sb.WriteString(` AND body LIKE ?`)
		args = append(args, "%"+q+"%")
	}
	sb.WriteString(` ORDER BY at DESC LIMIT ?`)
	args = append(args, limit)

	rows, err := s.db.Query(sb.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Feedback
	for rows.Next() {
		var f Feedback
		var img string
		if err := rows.Scan(&f.ID, &f.At, &f.GuestID, &f.Name, &f.Kind, &f.Body,
			&f.Context, &img, &f.Status, &f.Note); err != nil {
			return nil, err
		}
		f.HasImage = img != ""
		out = append(out, f)
	}
	return out, rows.Err()
}

// UnreadFeedback đếm thư ở trạng thái "mới" (số hiện ở thanh bên, §7.2).
func (s *Store) UnreadFeedback() int {
	var n int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM feedback WHERE status = 'moi'`).Scan(&n)
	return n
}

// UpdateFeedback đổi trạng thái và/hoặc ghi chú.
func (s *Store) UpdateFeedback(id int64, status, note string) error {
	if status != "" && !validStatus(status) {
		return errors.New("trạng thái không hợp lệ")
	}
	if status != "" {
		if _, err := s.db.Exec(`UPDATE feedback SET status = ? WHERE id = ?`, status, id); err != nil {
			return err
		}
	}
	if note != "" {
		if _, err := s.db.Exec(`UPDATE feedback SET note = ? WHERE id = ?`,
			cleanText(note, 2000), id); err != nil {
			return err
		}
	}
	return nil
}

// FeedbackImage trả về đường dẫn tuyệt đối tới ảnh của một lá thư.
//
// Đọc tên tệp TỪ DB chứ không nhận từ URL, rồi `filepath.Base` một lần nữa: kể cả khi DB
// có dòng hỏng, cũng không ra khỏi được thư mục ảnh.
func (s *Store) FeedbackImage(dir string, id int64) (string, error) {
	var rel string
	if err := s.db.QueryRow(`SELECT image FROM feedback WHERE id = ?`, id).Scan(&rel); err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("không có lá thư này")
		}
		return "", err
	}
	if rel == "" {
		return "", errors.New("lá thư này không có ảnh")
	}
	return filepath.Join(dir, filepath.Base(rel)), nil
}

// ── Cấm người chơi (§6.2) ────────────────────────────────────────────────────────────────

// Ban là một lệnh cấm đang hiệu lực.
type Ban struct {
	ID      int64  `json:"id"`
	GuestID string `json:"guest_id"`
	IP      string `json:"ip"`
	Reason  string `json:"reason"`
	From    int64  `json:"from"`
	Until   *int64 `json:"until"` // nil = vĩnh viễn
}

// BanOf trả về lệnh cấm đang áp lên guest_id hoặc IP này, nil nếu không bị cấm.
func (s *Store) BanOf(guestID, ip string) *Ban {
	now := time.Now().Unix()
	row := s.db.QueryRow(
		`SELECT id, guest_id, ip, reason, from_at, until_at FROM bans
		 WHERE (guest_id <> '' AND guest_id = ?) OR (ip <> '' AND ip = ?)
		 ORDER BY (until_at IS NULL) DESC, until_at DESC LIMIT 1`, guestID, ip)
	var b Ban
	var until sql.NullInt64
	if err := row.Scan(&b.ID, &b.GuestID, &b.IP, &b.Reason, &b.From, &until); err != nil {
		return nil
	}
	if until.Valid {
		if until.Int64 <= now {
			return nil // đã hết hạn
		}
		v := until.Int64
		b.Until = &v
	}
	return &b
}

// ErrBanReason: cấm mà không nêu lý do thì người bị cấm không biết mình làm sai gì, và ba
// tháng sau chính admin cũng không nhớ.
var ErrBanReason = errors.New("phải nêu lý do cấm")

// Ban cấm theo guest_id VÀ IP (§6.2). `until` nil = vĩnh viễn.
//
// Cấm cả hai là chủ ý của spec: chỉ cấm guest_id thì xoá localStorage là vào lại được ngay;
// chỉ cấm IP thì đụng cả nhà người ta. Cấm cả hai thì gỡ cũng gỡ cả hai — xem `Unban`.
func (s *Store) Ban(guestID, ip, reason string, until *int64, byIP string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return ErrBanReason
	}
	if len(reason) > 300 {
		reason = reason[:300]
	}
	var u any
	if until != nil {
		u = *until
	}
	_, err := s.db.Exec(
		`INSERT INTO bans (guest_id, ip, reason, from_at, until_at, by_ip) VALUES (?,?,?,?,?,?)`,
		guestID, ip, reason, time.Now().Unix(), u, byIP)
	return err
}

// Unban gỡ MỌI lệnh cấm đang áp lên guest_id này, kể cả lệnh cấm theo IP đi kèm.
//
// Xoá theo guest_id một mình là chưa đủ: `Ban` ghi cả IP vào cùng dòng, và nếu chỉ xoá dòng
// khớp guest_id thì một dòng cũ khớp IP vẫn chặn người đó — admin bấm "gỡ cấm", thấy báo
// thành công, mà người kia vẫn không vào được.
func (s *Store) Unban(guestID string) error {
	_, err := s.db.Exec(
		`DELETE FROM bans WHERE guest_id = ? OR (ip <> '' AND ip IN (SELECT ip FROM bans WHERE guest_id = ?))`,
		guestID, guestID)
	return err
}

// ListBans liệt kê lệnh cấm CÒN HIỆU LỰC, mới nhất trước.
func (s *Store) ListBans() ([]Ban, error) {
	rs, err := s.db.Query(
		`SELECT id, guest_id, ip, reason, from_at, until_at FROM bans
		 WHERE until_at IS NULL OR until_at > ? ORDER BY from_at DESC LIMIT 500`, time.Now().Unix())
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	out := []Ban{}
	for rs.Next() {
		var b Ban
		var until sql.NullInt64
		if err := rs.Scan(&b.ID, &b.GuestID, &b.IP, &b.Reason, &b.From, &until); err != nil {
			return nil, err
		}
		if until.Valid {
			v := until.Int64
			b.Until = &v
		}
		// IP không bao giờ gửi nguyên về trình duyệt, kể cả cho admin: danh sách cấm là thứ
		// mở suốt ngày trong một tab, và nó không cần IP đầy đủ để dùng được.
		b.IP = maskIPStore(b.IP)
		out = append(out, b)
	}
	return out, rs.Err()
}

// maskIPStore che IP trước khi rời server. Giống `maskIP` bên cmd/server nhưng để ở đây để
// package này không bao giờ trả ra IP thô, dù ai gọi.
func maskIPStore(ip string) string {
	if ip == "" {
		return ""
	}
	if i := strings.LastIndex(ip, ":"); i > 0 && strings.Count(ip, ":") > 1 {
		p := strings.Split(ip, ":")
		return p[0] + ":" + p[1] + ":x:x"
	}
	p := strings.Split(ip, ".")
	if len(p) == 4 {
		return p[0] + "." + p[1] + ".x.x"
	}
	return "x.x.x.x"
}
