package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Bảo vệ `assets/src/manifest.json` khỏi chính `artgen prep` (chủ dự án chốt 1/11).
//
// Sự cố: `prep -only crown_small` quét lại cả thư mục rồi ghi đè manifest. 149 dòng trỏ tới
// ảnh không còn trên đĩa (đã bị xoá trước đó) biến mất trong im lặng — không lỗi, không cảnh
// báo, chỉ một dòng "prep plate_crown_small". Nên giờ: manifest chỉ được THÊM dòng; mất dòng
// nào là dừng, không ghi, trừ khi người chạy nói rõ `-allow-drop`.

// readManifest đọc danh sách tệp hiện có trong manifest. Thiếu tệp → rỗng.
func readManifest(assets string) ([]string, error) {
	b, err := os.ReadFile(filepath.Join(assets, "manifest.json"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var m struct {
		Files []string `json:"files"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("manifest hỏng: %w", err)
	}
	return m.Files, nil
}

// saveManifest ghi manifest đúng định dạng cũ (sắp xếp, thụt 1 dấu cách).
func saveManifest(assets string, files []string) error {
	sort.Strings(files)
	b, _ := json.MarshalIndent(map[string]any{"files": files}, "", " ")
	return os.WriteFile(filepath.Join(assets, "manifest.json"), b, 0o644)
}

// guardManifest trả lỗi nếu `next` thiếu dòng nào của `old`.
func guardManifest(old, next []string, allowDrop bool) error {
	have := make(map[string]bool, len(next))
	for _, f := range next {
		have[f] = true
	}
	var lost []string
	for _, f := range old {
		if !have[f] {
			lost = append(lost, f)
		}
	}
	if len(lost) == 0 || allowDrop {
		return nil
	}
	show := lost
	if len(show) > 8 {
		show = show[:8]
	}
	return fmt.Errorf("manifest sẽ MẤT %d dòng (%s…) — DỪNG, không ghi. Cố ý thì chạy lại với -allow-drop",
		len(lost), strings.Join(show, ", "))
}

// updateManifest: chạy `-only` thì chỉ THÊM tệp của các job vừa prep; chạy toàn bộ thì quét
// lại đĩa. Cả hai đều qua guardManifest.
func updateManifest(assets string, done []Job, o prepOpts) error {
	old, err := readManifest(assets)
	if err != nil {
		return err
	}
	var next []string
	if o.scoped {
		next = mergeFiles(old, jobFiles(assets, done))
	} else if next, err = scanManifest(assets); err != nil {
		return err
	}
	if err := guardManifest(old, next, o.allowDrop); err != nil {
		return err
	}
	return saveManifest(assets, next)
}

// mergeFiles hợp hai danh sách, bỏ trùng.
func mergeFiles(a, b []string) []string {
	seen := make(map[string]bool, len(a)+len(b))
	out := make([]string, 0, len(a)+len(b))
	for _, f := range append(append([]string{}, a...), b...) {
		if !seen[f] {
			seen[f] = true
			out = append(out, f)
		}
	}
	return out
}

// jobFiles: tệp game của các job (đường dẫn tương đối, dấu /) mà đang có trên đĩa.
func jobFiles(assets string, jobs []Job) []string {
	var out []string
	add := func(rel string) {
		if exists(filepath.Join(assets, filepath.FromSlash(rel))) {
			out = append(out, rel)
		}
	}
	for _, j := range jobs {
		if j.Mode != modeSheet {
			add(j.Out)
			continue
		}
		for _, sg := range j.Segs {
			for _, f := range segOutputs(assets, sg) {
				add(f)
			}
		}
	}
	return out
}

// segOutputs: frame của nhân vật/quái nằm trong `frames/` và game chỉ dùng atlas; đoạn khác
// (vật cản, cổng…) là từng ảnh rời.
func segOutputs(assets string, sg Seg) []string {
	if filepath.Base(sg.Out) == "frames" || filepath.Base(sg.Out) == "frames_norm" {
		dir := filepath.ToSlash(filepath.Dir(sg.Out))
		return []string{dir + "/atlas.png", dir + "/atlas.json"}
	}
	if len(sg.Names) > 0 {
		out := make([]string, 0, len(sg.Names))
		for _, n := range sg.Names {
			out = append(out, sg.Out+"/"+n+".png")
		}
		return out
	}
	m, _ := filepath.Glob(filepath.Join(assets, filepath.FromSlash(sg.Out), sg.Anim+"_*.png"))
	out := make([]string, 0, len(m))
	for _, p := range m {
		rel, _ := filepath.Rel(assets, p)
		out = append(out, filepath.ToSlash(rel))
	}
	return out
}

// atlasDirsOf: thư mục `frames` (của nhân vật/quái, đường dẫn đầy đủ) mà các job vừa ghi vào.
func atlasDirsOf(assets string, jobs []Job) []string {
	seen := map[string]bool{}
	var out []string
	for _, j := range jobs {
		for _, sg := range j.Segs {
			if filepath.Base(sg.Out) == "frames" && !seen[sg.Out] {
				seen[sg.Out] = true
				out = append(out, filepath.Join(assets, filepath.FromSlash(sg.Out)))
			}
		}
	}
	return out
}
