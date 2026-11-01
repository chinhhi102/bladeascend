package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// GAME_SPEC §19: `/en/` trả trang tiếng Anh riêng; đường dẫn lạ vẫn rơi về trang gốc.
func TestSpaServesLanguageIndex(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("VI"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "en"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "en", "index.html"), []byte("EN"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := spa(dir)
	for path, want := range map[string]string{"/": "VI", "/en/": "EN", "/xyz": "VI", "/th/": "VI"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if got := strings.TrimSpace(rec.Body.String()); got != want {
			t.Errorf("GET %s = %q, muốn %q", path, got, want)
		}
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/en", nil))
	if rec.Code != 301 || rec.Header().Get("Location") != "/en/" {
		t.Errorf("GET /en = %d → %q, muốn 301 → /en/", rec.Code, rec.Header().Get("Location"))
	}
}
