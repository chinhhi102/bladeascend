package main

import (
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
)

// Thư viện art (spec §13.7-3b, luật 13 CLAUDE.md): art đã sinh mà không dùng nữa nằm ở
// `assets/library/<nhóm>/<đường dẫn như trong assets/src>`, không bao giờ bị xoá.
//
// artgen phải coi art trong thư viện là ĐÃ CÓ. Không thế thì `artgen run` thấy
// `vfx/boba_1.png` "thiếu" (vì đã rời assets/src) và mua lại đúng tấm ảnh đang nằm trong thư
// viện — đúng chuyện đã suýt xảy ra 1/11: `artgen status` liệt boba + 9 bộ regrow là "còn
// thiếu" sau khi chúng bị xoá ngày 31/10.

var (
	libOnce  sync.Once
	libFiles map[string]bool
)

// libraryDir: thư mục thư viện cạnh `assets/src`.
func libraryDir(assets string) string {
	return filepath.Join(filepath.Dir(filepath.Clean(assets)), "library")
}

// libraryIndex đọc thư viện MỘT lần: tập đường dẫn logic (bỏ tên nhóm ở đầu).
func libraryIndex(lib string) map[string]bool {
	out := map[string]bool{}
	_ = filepath.WalkDir(lib, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, rerr := filepath.Rel(lib, p)
		if rerr != nil {
			return nil
		}
		parts := strings.SplitN(filepath.ToSlash(rel), "/", 2)
		if len(parts) == 2 {
			out[parts[1]] = true
		}
		return nil
	})
	return out
}

// inLibrary: `rel` (đường dẫn logic, dấu /) có nằm trong thư viện không.
func inLibrary(assets, rel string) bool {
	libOnce.Do(func() { libFiles = libraryIndex(libraryDir(assets)) })
	return libFiles[filepath.ToSlash(rel)]
}

// haveArt: tệp art đã có — trong assets/src HOẶC trong thư viện.
func haveArt(assets, rel string) bool {
	return exists(filepath.Join(assets, filepath.FromSlash(rel))) || inLibrary(assets, rel)
}
