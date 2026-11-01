package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"image"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"bladeascend/server/internal/admin"

	_ "modernc.org/sqlite"
)

// Góp ý nhiều ảnh qua HTTP — ADMIN_SPEC §7.1, §9. Kiểm hợp đồng mà trang admin dựa vào:
// danh sách có `images: N`, ảnh thứ i ở `/image?i=`, thiếu `i` = ảnh đầu.

func feedbackServer(t *testing.T) *server {
	t.Helper()
	t.Setenv("ADMIN_PASSWORD_HASH", "")
	dir := t.TempDir()
	db, err := sql.Open("sqlite", "file:"+filepath.Join(dir, "lb.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	st, err := admin.New(db)
	if err != nil {
		t.Fatal(err)
	}
	auth, err := admin.NewAuth(st)
	if err != nil {
		t.Fatal(err)
	}
	return &server{adminAuth: auth, dbPath: filepath.Join(dir, "lb.db"), buildID: "test"}
}

// jpegWithGPS: JPEG thật kèm một khối APP1 Exif có chuỗi "GPSSECRET".
func jpegWithGPS(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewGray(image.Rect(0, 0, 8, 8)), nil); err != nil {
		t.Fatal(err)
	}
	j := buf.Bytes()
	payload := []byte("Exif\x00\x00MM\x00\x2AGPSSECRET")
	seg := []byte{0xFF, 0xE1, 0, 0}
	binary.BigEndian.PutUint16(seg[2:], uint16(len(payload)+2))
	out := append([]byte{}, j[:2]...)
	out = append(out, seg...)
	out = append(out, payload...)
	return append(out, j[2:]...)
}

func postFeedback(t *testing.T, s *server, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	r := httptest.NewRequest(http.MethodPost, "/api/feedback", bytes.NewReader(b))
	w := httptest.NewRecorder()
	s.handleFeedback(w, r)
	return w
}

func getImage(s *server, id, query string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodGet, "/api/admin/feedback/"+id+"/image"+query, nil)
	r.SetPathValue("id", id)
	w := httptest.NewRecorder()
	s.handleAdminFeedbackImage(w, r)
	return w
}

func TestFeedbackHTTPMultipleImages(t *testing.T) {
	s := feedbackServer(t)
	j := jpegWithGPS(t)
	data := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(j)
	w := postFeedback(t, s, map[string]any{
		"guest_id": "g1", "name": "A", "kind": "bug", "body": "ba ảnh đây",
		"images": []string{data, data, base64.StdEncoding.EncodeToString(j)},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("POST: %d %s", w.Code, w.Body)
	}

	lw := httptest.NewRecorder()
	s.handleAdminFeedback(lw, httptest.NewRequest(http.MethodGet, "/api/admin/feedback", nil))
	var list []map[string]any
	if err := json.Unmarshal(lw.Body.Bytes(), &list); err != nil || len(list) != 1 {
		t.Fatalf("list: %v %s", err, lw.Body)
	}
	if list[0]["images"] != float64(3) {
		t.Fatalf("images = %v, muốn 3", list[0]["images"])
	}

	for _, q := range []string{"", "?i=0", "?i=1", "?i=2"} {
		iw := getImage(s, "1", q)
		if iw.Code != http.StatusOK || iw.Header().Get("Content-Type") != "image/jpeg" {
			t.Fatalf("ảnh %q: %d %s", q, iw.Code, iw.Header().Get("Content-Type"))
		}
		if strings.Contains(iw.Body.String(), "GPSSECRET") {
			t.Errorf("ảnh %q còn EXIF/GPS", q)
		}
	}
	for _, q := range []string{"?i=3", "?i=-1", "?i=x"} {
		if iw := getImage(s, "1", q); iw.Code != http.StatusNotFound {
			t.Errorf("ảnh %q: muốn 404, được %d", q, iw.Code)
		}
	}
}

// Client cũ (còn trong bộ nhớ đệm) gửi một ảnh ở trường `image` — vẫn phải nhận.
func TestFeedbackHTTPLegacySingleImage(t *testing.T) {
	s := feedbackServer(t)
	data := base64.StdEncoding.EncodeToString(jpegWithGPS(t))
	w := postFeedback(t, s, map[string]any{
		"guest_id": "g1", "kind": "bug", "body": "client cũ", "image": data,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("POST: %d %s", w.Code, w.Body)
	}
	if iw := getImage(s, "1", ""); iw.Code != http.StatusOK {
		t.Fatalf("ảnh: %d", iw.Code)
	}
}

func TestFeedbackHTTPRejects(t *testing.T) {
	s := feedbackServer(t)
	data := base64.StdEncoding.EncodeToString(jpegWithGPS(t))
	cases := map[string]map[string]any{
		"4 ảnh": {"guest_id": "g1", "kind": "bug", "body": "quá nhiều ảnh",
			"images": []string{data, data, data, data}},
		"image + 3 images": {"guest_id": "g2", "kind": "bug", "body": "quá nhiều ảnh",
			"image": data, "images": []string{data, data, data}},
		"HTML giả ảnh": {"guest_id": "g3", "kind": "bug", "body": "ảnh giả",
			"images": []string{base64.StdEncoding.EncodeToString([]byte("<html>hi</html>"))}},
		"base64 hỏng": {"guest_id": "g4", "kind": "bug", "body": "ảnh hỏng",
			"images": []string{"!!!"}},
	}
	for name, body := range cases {
		if w := postFeedback(t, s, body); w.Code != http.StatusBadRequest {
			t.Errorf("%s: muốn 400, được %d %s", name, w.Code, w.Body)
		}
	}
}
