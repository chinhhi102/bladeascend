package main

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bladeascend/server/internal/admin"
)

// Hòm thư góp ý — ADMIN_SPEC §7, §9.

// feedbackDir: ảnh góp ý lưu NGOÀI thư mục phục vụ công khai (§8).
//
// Nếu để trong `assets/` hay `client/dist/` thì bất kỳ ai đoán được tên tệp cũng xem được
// ảnh chụp màn hình của người khác — mà ảnh đó có thể chứa tên, vị trí, cả khung chat.
func feedbackDir(dbPath string) string {
	return filepath.Join(filepath.Dir(dbPath), "feedback")
}

// feedbackBody là gói người chơi gửi lên.
type feedbackBody struct {
	GuestID string `json:"guest_id"`
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Body    string `json:"body"`
	// Context: ngữ cảnh tự đính kèm (§7.1) — build, thiết bị, khu, cảnh giới, ping, FPS…
	// Nhận dạng tự do rồi lưu nguyên JSON: nó chỉ để ĐỌC, không có luật nào chạy trên nó,
	// nên buộc kiểu ở đây chỉ tạo ra việc phải sửa mỗi lần client muốn gửi thêm một trường.
	Context map[string]any `json:"context"`
	// Image: ảnh đơn kiểu cũ (client trước khi có "Thêm ảnh"), base64 (data URL hoặc trần).
	// Vẫn nhận vì người chơi có thể còn giữ bản client cũ trong bộ nhớ đệm.
	Image string `json:"image"`
	// Images: 0..3 ảnh (chụp màn hình + ảnh tải lên), cùng dạng base64 như Image (§7.1).
	Images []string `json:"images"`
}

// feedbackMaxBody: 3 ảnh × 300 KB nở 4/3 khi base64 ≈ 1,2 MB, cộng ngữ cảnh và chữ.
const feedbackMaxBody = 1600 << 10

// decodeFeedbackImages gom ảnh kiểu cũ và kiểu mới rồi giải base64.
func decodeFeedbackImages(b feedbackBody) ([][]byte, error) {
	all := b.Images
	if b.Image != "" {
		all = append([]string{b.Image}, all...)
	}
	if len(all) > admin.FeedbackMaxImages {
		return nil, admin.ErrTooManyImages
	}
	out := make([][]byte, 0, len(all))
	for _, raw := range all {
		if raw == "" {
			continue
		}
		if i := strings.Index(raw, ","); strings.HasPrefix(raw, "data:") && i > 0 {
			raw = raw[i+1:]
		}
		dec, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return nil, errors.New("ảnh không giải mã được")
		}
		out = append(out, dec)
	}
	return out, nil
}

// handleFeedback nhận góp ý từ người chơi — đường dẫn CÔNG KHAI duy nhất ghi vào DB admin.
func (s *server) handleFeedback(w http.ResponseWriter, r *http.Request) {
	// Trần kích thước đặt ở ĐÂY chứ không tin `Content-Length`: header đó do client gửi.
	var b feedbackBody
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, feedbackMaxBody)).Decode(&b); err != nil {
		adminJSON(w, http.StatusBadRequest, map[string]string{"error": "nội dung không hợp lệ"})
		return
	}
	ip := admin.ClientIP(r)
	st := s.adminAuth.Store()
	// Người bị cấm không gửi được (§7.1). Kiểm cả guest_id lẫn IP.
	if ban := st.BanOf(b.GuestID, ip); ban != nil {
		adminJSON(w, http.StatusForbidden, map[string]string{"error": admin.ErrFeedbackBanned.Error()})
		return
	}
	imgs, err := decodeFeedbackImages(b)
	if err != nil {
		adminJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	// Ghi IP (băm) vào ngữ cảnh để sau này còn lần được, nhưng KHÔNG trả lại cho client.
	if b.Context == nil {
		b.Context = map[string]any{}
	}
	b.Context["ip"] = ip
	b.Context["server_build"] = s.buildID

	id, err := st.AddFeedback(feedbackDir(s.dbPath), b.GuestID, b.Name, b.Kind, b.Body, b.Context, imgs)
	if err != nil {
		code := http.StatusBadRequest
		if err == admin.ErrFeedbackRate {
			code = http.StatusTooManyRequests
		}
		adminJSON(w, code, map[string]string{"error": err.Error()})
		return
	}
	adminJSON(w, http.StatusOK, map[string]any{"ok": 1, "id": id})
}

// ── Phía admin (§7.2) ────────────────────────────────────────────────────────────────────

func (s *server) handleAdminFeedback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	rows, err := s.adminAuth.Store().ListFeedback(q.Get("kind"), q.Get("status"), q.Get("q"), 300)
	if err != nil {
		adminJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if rows == nil {
		rows = []admin.Feedback{}
	}
	adminJSON(w, http.StatusOK, rows)
}

func (s *server) handleAdminFeedbackPatch(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		adminJSON(w, http.StatusBadRequest, map[string]string{"error": "id không hợp lệ"})
		return
	}
	var b struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&b); err != nil {
		adminJSON(w, http.StatusBadRequest, map[string]string{"error": "nội dung không hợp lệ"})
		return
	}
	if err := s.adminAuth.Store().UpdateFeedback(id, b.Status, b.Note); err != nil {
		adminJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.adminAuth.Store().Audit("feedback_update", strconv.FormatInt(id, 10), b.Status,
		admin.ClientIP(r), map[string]any{"status": b.Status, "has_note": b.Note != ""})
	adminJSON(w, http.StatusOK, map[string]int{"ok": 1})
}

// handleAdminFeedbackImage phục vụ ảnh góp ý — CHỈ sau khi đăng nhập (§8).
func (s *server) handleAdminFeedbackImage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// `?i=` chọn ảnh thứ i (0-based); thiếu `i` = ảnh đầu, như trước khi có nhiều ảnh.
	idx := 0
	if v := r.URL.Query().Get("i"); v != "" {
		if idx, err = strconv.Atoi(v); err != nil {
			http.NotFound(w, r)
			return
		}
	}
	path, err := s.adminAuth.Store().FeedbackImage(feedbackDir(s.dbPath), id, idx)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// Kiểm chữ ký BYTE một lần nữa lúc phục vụ, không chỉ lúc nhận. Tệp trên đĩa có thể đã
	// bị thay bằng đường khác (sao lưu, đồng bộ, tay người); phục vụ đúng thứ mình đã xác
	// minh là rẻ hơn nhiều so với phục vụ HTML vào trang quản trị.
	var ct string
	switch filepath.Ext(path) {
	case ".webp":
		ct = "image/webp"
	case ".png":
		ct = "image/png"
	case ".jpg":
		ct = "image/jpeg"
	}
	if ct == "" || !admin.LooksLikeImage(b) {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "no-store")
	// Chặn trình duyệt tự đoán kiểu nội dung: đó là đường mà một tệp "ảnh" biến thành HTML.
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
	_, _ = w.Write(b)
}

// handleAdminFeedbackCSV xuất CSV (§7.2).
func (s *server) handleAdminFeedbackCSV(w http.ResponseWriter, r *http.Request) {
	rows, err := s.adminAuth.Store().ListFeedback("", "", "", 5000)
	if err != nil {
		adminJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="gop-y.csv"`)
	cw := csv.NewWriter(w)
	defer cw.Flush()
	_ = cw.Write([]string{"id", "thoi_gian", "loai", "trang_thai", "ten", "guest_id", "noi_dung", "ghi_chu", "so_anh"})
	for _, f := range rows {
		_ = cw.Write([]string{
			strconv.FormatInt(f.ID, 10),
			time.Unix(f.At, 0).Format(time.RFC3339),
			f.Kind, f.Status, f.Name, f.GuestID, f.Body, f.Note,
			strconv.Itoa(f.Images),
		})
	}
}
