package main

import (
	"os"
	"path/filepath"
	"testing"
)

// Art đã chuyển vào thư viện phải tính là ĐÃ CÓ — không thì `artgen run` mua lại nó.
func TestLibraryArtCountsAsHave(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	put := func(p string) {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("png"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	put(filepath.Join(root, "library", "vfx_boba", "vfx", "boba_1.png"))
	put(filepath.Join(root, "library", "obstacle_regrow", "obstacles", "stall", "regrow", "regrow_00.png"))
	libFiles = libraryIndex(libraryDir(src))
	libOnce.Do(func() {}) // chặn Do() đọc lại thư mục thật

	sprite := Job{Name: "vfx_boba_1", Mode: modeSprite, Out: "vfx/boba_1.png"}
	sheet := Job{Name: "obs_stall_regrow", Mode: modeSheet,
		Segs: []Seg{{Anim: "regrow", Frames: 6, Out: "obstacles/stall/regrow"}}}
	missing := Job{Name: "vfx_boba_2", Mode: modeSprite, Out: "vfx/boba_2.png"}
	if !haveOutputs(src, sprite) || !haveOutputs(src, sheet) {
		t.Fatal("art trong assets/library bị tính là thiếu → artgen run sẽ mua lại")
	}
	if haveOutputs(src, missing) {
		t.Fatal("ảnh không có ở đâu cả mà vẫn tính là đã có")
	}
}
