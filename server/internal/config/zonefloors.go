package config

import (
	"encoding/json"
	"fmt"
	"sort"
)

// ZoneFloors: sàn nhân vật (người + bot) THEO KHU — spec v6.35 §9.5.
//
//	{"thien_dinh": 30, "default": 20}
//
// Khoá là id bản đồ của khu (`maps/<id>.tmj`), "default" cho mọi khu không ghi riêng.
// Đọc được cả dạng cũ là MỘT con số (mọi khu như nhau): cấu hình admin đã lưu trong SQLite
// trước v6.35 vẫn nạp lại đúng sau deploy.
type ZoneFloors map[string]int

// ZoneFloorDefault: khoá áp cho mọi khu không ghi riêng.
const ZoneFloorDefault = "default"

// Uniform: mọi khu cùng một sàn n.
func Uniform(n int) ZoneFloors { return ZoneFloors{ZoneFloorDefault: n} }

// For: sàn của khu có bản đồ `zoneID`.
func (f ZoneFloors) For(zoneID string) int {
	if n, ok := f[zoneID]; ok {
		return n
	}
	return f[ZoneFloorDefault]
}

// UnmarshalJSON nhận số (dạng cũ) hoặc bảng theo khu.
func (f *ZoneFloors) UnmarshalJSON(b []byte) error {
	var n int
	if err := json.Unmarshal(b, &n); err == nil {
		*f = Uniform(n)
		return nil
	}
	m := map[string]int{}
	if err := json.Unmarshal(b, &m); err != nil {
		return fmt.Errorf("zone_min_chars: cần số hoặc {\"<khu>\": số, \"default\": số}: %w", err)
	}
	*f = m
	return nil
}

// Clone: bản sao (map dùng chung giữa các goroutine phải sao ra trước khi sửa).
func (f ZoneFloors) Clone() ZoneFloors {
	out := make(ZoneFloors, len(f))
	for k, v := range f {
		out[k] = v
	}
	return out
}

// String: "thien_dinh 30 · default 20" — cho log và nhật ký admin.
func (f ZoneFloors) String() string {
	keys := make([]string, 0, len(f))
	for k := range f {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	s := ""
	for i, k := range keys {
		if i > 0 {
			s += " · "
		}
		s += fmt.Sprintf("%s %d", k, f[k])
	}
	return s
}
