package admin

import (
	"bytes"
	"database/sql"
	"os"
	"testing"
	"time"
)

// Góp ý nhiều ảnh — ADMIN_SPEC §7.1: tối đa 3 ảnh × 300 KB, EXIF bị gỡ trước khi lưu.

func TestAddFeedbackMultipleImagesStrippedAndIndexed(t *testing.T) {
	s, dir := testStore(t), t.TempDir()
	withGPS := insertAfterSOI(testJPEG(t), jpegSeg(0xE1, exifGPSBlock()))
	imgs := [][]byte{testPNG(t), withGPS, testPNG(t)}
	id, err := s.AddFeedback(dir, "g1", "A", "bug", "ảnh nhiều tấm", nil, imgs)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := s.ListFeedback("", "", "", 10)
	if err != nil || len(rows) != 1 {
		t.Fatalf("list: %v %d", err, len(rows))
	}
	if rows[0].Images != 3 || !rows[0].HasImage {
		t.Fatalf("images = %d, has_image = %v", rows[0].Images, rows[0].HasImage)
	}
	for i := 0; i < 3; i++ {
		p, err := s.FeedbackImage(dir, id, i)
		if err != nil {
			t.Fatalf("ảnh %d: %v", i, err)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(b, []byte(gpsMarker)) || bytes.Contains(b, []byte("Exif")) {
			t.Errorf("ảnh %d lưu xuống đĩa còn EXIF/GPS", i)
		}
	}
	if _, err := s.FeedbackImage(dir, id, 3); err == nil {
		t.Error("i = 3 phải lỗi (chỉ có 3 ảnh)")
	}
	if _, err := s.FeedbackImage(dir, id, -1); err == nil {
		t.Error("i âm phải lỗi")
	}
}

func TestAddFeedbackRejectsBadImages(t *testing.T) {
	s, dir := testStore(t), t.TempDir()
	png := testPNG(t)
	big := append(append([]byte{}, png...), make([]byte, FeedbackMaxImage)...)
	cases := map[string][][]byte{
		"4 ảnh":          {png, png, png, png},
		"quá 300 KB":     {big},
		"không phải ảnh": {png, []byte("<html>")},
		"PNG hỏng":       {png[:20]},
	}
	for name, imgs := range cases {
		if _, err := s.AddFeedback(dir, "g-"+name, "A", "bug", "nội dung đủ dài", nil, imgs); err == nil {
			t.Errorf("%s: phải bị từ chối", name)
		}
	}
	// Từ chối thì không để lại tệp mồ côi nào.
	if ents, _ := os.ReadDir(dir); len(ents) != 0 {
		t.Errorf("còn %d tệp mồ côi sau khi từ chối", len(ents))
	}
}

func TestFeedbackNoImage(t *testing.T) {
	s, dir := testStore(t), t.TempDir()
	id, err := s.AddFeedback(dir, "g", "A", "idea", "không có ảnh", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	rows, _ := s.ListFeedback("", "", "", 10)
	if rows[0].Images != 0 || rows[0].HasImage {
		t.Error("thư không ảnh lại báo có ảnh")
	}
	if _, err := s.FeedbackImage(dir, id, 0); err == nil {
		t.Error("đòi ảnh của thư không ảnh phải lỗi")
	}
}

// Rate limit 3 góp ý / 10 phút / guest_id vẫn giữ.
func TestFeedbackRateLimitKept(t *testing.T) {
	s, dir := testStore(t), t.TempDir()
	for i := 0; i < FeedbackPerWindow; i++ {
		if _, err := s.AddFeedback(dir, "g", "A", "idea", "lần thứ mấy đó", nil, nil); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.AddFeedback(dir, "g", "A", "idea", "lần quá tay", nil, nil); err != ErrFeedbackRate {
		t.Fatalf("muốn ErrFeedbackRate, được %v", err)
	}
}

// DB tạo trước khi có nhiều ảnh: không có cột `images`. Mở bằng server mới thì dòng cũ vẫn
// đọc được ảnh (qua `image`), và không mất dòng nào.
func TestFeedbackMigrationKeepsOldRows(t *testing.T) {
	dbPath := t.TempDir() + "/old.db"
	db, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE feedback (
	  id INTEGER PRIMARY KEY AUTOINCREMENT, at INTEGER NOT NULL, guest_id TEXT NOT NULL,
	  name TEXT NOT NULL, kind TEXT NOT NULL, body TEXT NOT NULL, context TEXT NOT NULL,
	  image TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'moi', note TEXT NOT NULL DEFAULT '')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO feedback (at, guest_id, name, kind, body, context, image)
	  VALUES (?, 'g', 'A', 'bug', 'thư cũ', '{}', 'fb_1_10.webp'), (?, 'h', 'B', 'idea', 'thư cũ 2', '{}', '')`,
		time.Now().Unix(), time.Now().Unix()); err != nil {
		t.Fatal(err)
	}
	s, err := New(db)
	if err != nil {
		t.Fatal(err)
	}
	// Mở lần hai: migration không được chạy lại mà lỗi "duplicate column".
	if _, err := New(db); err != nil {
		t.Fatalf("mở lần hai: %v", err)
	}
	rows, err := s.ListFeedback("", "", "", 10)
	if err != nil || len(rows) != 2 {
		t.Fatalf("mất thư cũ: %v, còn %d", err, len(rows))
	}
	got := map[string]int{}
	for _, f := range rows {
		got[f.Body] = f.Images
	}
	if got["thư cũ"] != 1 || got["thư cũ 2"] != 0 {
		t.Errorf("số ảnh của thư cũ sai: %v", got)
	}
	var id int64
	_ = db.QueryRow(`SELECT id FROM feedback WHERE body = 'thư cũ'`).Scan(&id)
	p, err := s.FeedbackImage("/d", id, 0)
	if err != nil || p != "/d/fb_1_10.webp" {
		t.Errorf("ảnh thư cũ: %q %v", p, err)
	}
}

// Dọn ảnh cũ phải xoá MỌI ảnh của một thư nhiều ảnh, không chỉ ảnh đầu.
func TestPurgeRemovesAllImagesOfLetter(t *testing.T) {
	s, dir := testStore(t), t.TempDir()
	id, err := s.AddFeedback(dir, "g", "A", "bug", "ba ảnh cũ", nil,
		[][]byte{testPNG(t), testJPEG(t), testPNG(t)})
	if err != nil {
		t.Fatal(err)
	}
	old := time.Now().AddDate(0, 0, -FeedbackImageKeepDays-1).Unix()
	if _, err := s.db.Exec(`UPDATE feedback SET at = ?, status = 'xong' WHERE id = ?`, old, id); err != nil {
		t.Fatal(err)
	}
	n, err := s.PurgeOldImages(dir, time.Now())
	if err != nil || n != 3 {
		t.Fatalf("muốn xoá 3 ảnh, xoá %d (%v)", n, err)
	}
	if ents, _ := os.ReadDir(dir); len(ents) != 0 {
		t.Errorf("còn %d tệp ảnh sau khi dọn", len(ents))
	}
	rows, _ := s.ListFeedback("", "", "", 10)
	if len(rows) != 1 || rows[0].Images != 0 || rows[0].Body == "" {
		t.Errorf("sau khi dọn: %+v", rows)
	}
}
