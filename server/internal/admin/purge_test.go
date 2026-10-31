package admin

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Dọn ảnh góp ý — QUESTIONS Q28, chủ dự án chốt phương án B.
//
// Luật: ảnh của thư ĐÃ KHÉP ("xong"/"bo_qua") quá 90 ngày thì xoá; **chữ giữ mãi**, và thư
// chưa khép thì không đụng tới dù cũ cỡ nào.
//
// Ba nhánh, và nhánh dễ sai nhất là nhánh thứ ba: một bộ dọn "theo tuổi" viết vội sẽ xoá cả
// ảnh của một báo lỗi đang điều tra dở.

func purgeStore(t *testing.T) (*Store, string) {
	t.Helper()
	dir := t.TempDir()
	s := testStore(t)
	return s, dir
}

// addWithImage thêm một lá thư có ảnh, với thời điểm và trạng thái cho trước.
func addWithImage(t *testing.T, s *Store, dir, name, status string, at time.Time) string {
	t.Helper()
	// PNG tối thiểu: `AddFeedback` kiểm CHỮ KÝ BYTE, không kiểm đuôi tệp.
	png := append([]byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}, make([]byte, 64)...)
	id, err := s.AddFeedback(dir, "g-"+name, name, "bug", "nội dung thử nghiệm", nil, png)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`UPDATE feedback SET at = ?, status = ? WHERE id = ?`,
		at.Unix(), status, id); err != nil {
		t.Fatal(err)
	}
	var file string
	if err := s.db.QueryRow(`SELECT image FROM feedback WHERE id = ?`, id).Scan(&file); err != nil {
		t.Fatal(err)
	}
	if file == "" {
		t.Fatal("thư vừa thêm không có ảnh — bài kiểm không chạm tới thứ nó định kiểm")
	}
	return file
}

func TestPurgeOldImagesKeepsTextAndOpenLetters(t *testing.T) {
	s, dir := purgeStore(t)
	now := time.Now()
	old := now.AddDate(0, 0, -FeedbackImageKeepDays-10)
	recent := now.AddDate(0, 0, -5)

	oldDone := addWithImage(t, s, dir, "cu-xong", "xong", old)
	oldSkip := addWithImage(t, s, dir, "cu-bo-qua", "bo_qua", old)
	oldOpen := addWithImage(t, s, dir, "cu-dang-lam", "dang_lam", old)
	newDone := addWithImage(t, s, dir, "moi-xong", "xong", recent)

	// Rate limit 3 lá/10 phút tính theo guest_id, nên bốn `guest_id` khác nhau ở trên là cố ý.
	n, err := s.PurgeOldImages(dir, now)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("muốn xoá đúng 2 ảnh (thư đã khép + quá hạn), xoá %d", n)
	}
	gone := func(f string) bool {
		_, err := os.Stat(filepath.Join(dir, filepath.Base(f)))
		return os.IsNotExist(err)
	}
	if !gone(oldDone) || !gone(oldSkip) {
		t.Error("ảnh của thư ĐÃ KHÉP quá hạn vẫn còn")
	}
	if gone(oldOpen) {
		t.Error("xoá mất ảnh của thư ĐANG LÀM — đó là bằng chứng của một việc dở dang")
	}
	if gone(newDone) {
		t.Error("xoá mất ảnh của thư mới khép 5 ngày")
	}

	// CHỮ phải còn nguyên — cả bốn lá.
	rows, err := s.ListFeedback("", "", "", 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 4 {
		t.Fatalf("dọn ảnh mà mất thư: còn %d/4 lá", len(rows))
	}
	for _, f := range rows {
		if f.Body == "" {
			t.Errorf("thư %d mất nội dung — chỉ được xoá ẢNH", f.ID)
		}
	}
	// Và cờ `has_image` phải tắt đúng hai lá đã dọn.
	withImg := 0
	for _, f := range rows {
		if f.HasImage {
			withImg++
		}
	}
	if withImg != 2 {
		t.Errorf("muốn còn 2 lá có ảnh, được %d", withImg)
	}
}

// Chạy hai lần liên tiếp không được lỗi và không xoá thêm gì.
func TestPurgeIsIdempotent(t *testing.T) {
	s, dir := purgeStore(t)
	now := time.Now()
	addWithImage(t, s, dir, "cu", "xong", now.AddDate(0, 0, -200))
	if n, err := s.PurgeOldImages(dir, now); err != nil || n != 1 {
		t.Fatalf("lượt đầu: xoá %d, lỗi %v", n, err)
	}
	if n, err := s.PurgeOldImages(dir, now); err != nil || n != 0 {
		t.Fatalf("lượt hai phải không xoá gì: xoá %d, lỗi %v", n, err)
	}
}
