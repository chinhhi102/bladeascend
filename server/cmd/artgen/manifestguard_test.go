package main

import (
	"os"
	"path/filepath"
	"testing"
)

// Manifest chỉ được THÊM dòng — sự cố 1/11: `prep -only crown_small` xoá 149 dòng im lặng.
func TestGuardManifestRefusesToDropLines(t *testing.T) {
	old := []string{"buffs/long.png", "ui/crown_small.png"}
	if err := guardManifest(old, []string{"ui/crown_small.png"}, false); err == nil {
		t.Fatal("manifest mất dòng mà không bị chặn")
	}
	if err := guardManifest(old, []string{"ui/crown_small.png"}, true); err != nil {
		t.Fatalf("-allow-drop vẫn bị chặn: %v", err)
	}
	if err := guardManifest(old, append(old, "ui/new.png"), false); err != nil {
		t.Fatalf("thêm dòng bị chặn: %v", err)
	}
}

// prep -only: manifest giữ nguyên mọi dòng cũ (kể cả dòng trỏ tới tệp không còn trên đĩa),
// chỉ thêm tệp của đúng job đó.
func TestScopedManifestOnlyAddsJobFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "ui"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"ui/crown_small.png", "ui/other.png"} {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := saveManifest(dir, []string{"buffs/long.png"}); err != nil {
		t.Fatal(err)
	}
	job := Job{Name: "plate_crown_small", Mode: modeSprite, Out: "ui/crown_small.png"}
	if err := updateManifest(dir, []Job{job}, prepOpts{scoped: true}); err != nil {
		t.Fatal(err)
	}
	got, _ := readManifest(dir)
	want := map[string]bool{"buffs/long.png": true, "ui/crown_small.png": true}
	if len(got) != len(want) {
		t.Fatalf("manifest = %v; muốn đúng dòng cũ + tệp của job (ui/other.png KHÔNG được thêm)", got)
	}
	for _, f := range got {
		if !want[f] {
			t.Fatalf("manifest có %q — không thuộc job", f)
		}
	}
	// Chạy toàn bộ (quét đĩa) sẽ làm mất "buffs/long.png" → phải dừng, không ghi.
	if err := updateManifest(dir, nil, prepOpts{}); err == nil {
		t.Fatal("quét toàn bộ làm mất dòng mà vẫn ghi")
	}
	after, _ := readManifest(dir)
	if len(after) != 2 {
		t.Fatalf("bị chặn mà manifest vẫn đổi: %v", after)
	}
}
