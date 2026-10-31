package main

import (
	"sort"
	"testing"

	"kiemxoay/server/internal/config"
)

// TestDecorPlacedAlwaysHasArt — 5 khu × 10 seed, MỌI vật trang trí đặt ra đều phải có ảnh.
//
// Đây là phép kiểm chủ dự án yêu cầu ở Q27-2, và nó canh một lỗi không để lại dấu vết:
// `mapview.buildDecor()` gặp `decor/<tên>.png` không có thì `continue` — không log, không
// placeholder, build xanh. 543 vật trang trí trong 5 bản đồ không vẽ gì suốt nhiều tuần.
//
// Chạy nhiều seed vì mapgen rải ngẫu nhiên: một seed có thể tình cờ không đặt được loại
// đang thiếu art, và test sẽ xanh vì may mắn.
func TestDecorPlacedAlwaysHasArt(t *testing.T) {
	cfg := config.MustLoadDefault()
	have := haveDecorArt()
	if len(have) == 0 {
		t.Skip("không đọc được assets/src/decor — bỏ qua (môi trường không có art)")
	}
	placed := map[string]int{}
	for zi, zd := range zoneDefs {
		for s := int64(0); s < 10; s++ {
			b, _, err := buildZoneWithRetries(cfg.Room.MapSize, 1000+int64(zi)*100+s, cfg.Mapgen, zi)
			if err != nil {
				t.Fatalf("khu %s seed %d: sinh thất bại: %v", zd.id, s, err)
			}
			for _, o := range b.layers["decor"] {
				placed[o.Type]++
				if !have[o.Type] {
					t.Errorf("khu %s seed %d đặt trang trí %q nhưng KHÔNG có assets/src/decor/%s.png",
						zd.id, s, o.Type, o.Type)
				}
			}
			// Chốt chặn trong `builder.decor` phải là thứ DUY NHẤT từ chối, và nó chỉ được
			// từ chối cái đã biết là thiếu art — từ chối một tên có art là lỗi khác.
			for _, skipped := range b.decorSkipped {
				if have[skipped] {
					t.Errorf("khu %s seed %d từ chối %q dù CÓ art", zd.id, s, skipped)
				}
			}
		}
	}
	if len(placed) == 0 {
		t.Fatal("không khu nào đặt được vật trang trí nào — bảng decorZones hoặc art đã lệch")
	}
	names := make([]string, 0, len(placed))
	for n := range placed {
		names = append(names, n)
	}
	sort.Strings(names)
	t.Logf("đặt được %d loại trên 50 bản đồ: %v", len(placed), names)
}

// TestEveryZoneGetsSomeDecor — mỗi khu phải đặt được ÍT NHẤT MỘT loại trang trí.
//
// Không có phép kiểm này thì "chỉ đặt cái có art" có một cách thoả mãn rất tệ: không đặt gì
// cả. Băng Nguyệt Hồ chỉ được gắn tag cho `rock_small` + `pine` (cả hai chưa có art) và
// `bush_with_flowers` — nếu lỡ bỏ cái cuối ra khỏi bảng thì khu đó trống trơn mà test trên
// vẫn xanh.
func TestEveryZoneGetsSomeDecor(t *testing.T) {
	have := haveDecorArt()
	if len(have) == 0 {
		t.Skip("không đọc được assets/src/decor")
	}
	for _, zd := range zoneDefs {
		use, missing := decorFor(zd.id, have)
		if len(use) == 0 {
			t.Errorf("khu %s không có loại trang trí nào vừa hợp chủ đề vừa có art (thiếu art: %v)", zd.id, missing)
		}
		t.Logf("%-16s dùng %d loại %v · chờ art %v", zd.id, len(use), use, missing)
	}
}

// TestDecorZonesNamesAreKnown — mọi tên trong `decorZones` phải trỏ tới một khu THẬT.
//
// Gõ sai tên khu (`"pham_tran "`, `"phamtran"`) thì vật trang trí đó im lặng không bao giờ
// được đặt ở đâu cả — đúng kiểu lỗi Q27-2, chỉ ở một tầng khác.
func TestDecorZonesNamesAreKnown(t *testing.T) {
	known := map[string]bool{}
	for _, zd := range zoneDefs {
		known[zd.id] = true
	}
	for name, zones := range decorZones {
		if len(zones) == 0 {
			t.Errorf("trang trí %q không gắn khu nào", name)
		}
		for _, z := range zones {
			if !known[z] {
				t.Errorf("trang trí %q gắn khu %q không tồn tại", name, z)
			}
		}
	}
	for _, zd := range zoneDefs {
		if _, ok := decorPerZone[zd.id]; !ok {
			t.Errorf("khu %s thiếu mật độ trong decorPerZone", zd.id)
		}
	}
}
