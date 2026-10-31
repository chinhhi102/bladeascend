package main

// Vật trang trí: chọn theo KHU, và chỉ đặt cái NÀO CÓ ART (Q27-2, GAME_SPEC §8.6 mục 3).
//
// Vì sao phải có file này. `mapview.buildDecor()` gặp `decor/<tên>.png` không tồn tại thì
// `continue` — im lặng tuyệt đối, không log, không placeholder. Mapgen thì rải tên theo
// bảng `biomes[].decor` gõ tay (`rock_small`, `pine`, `lantern`, `well`, `banner`,
// `bamboo`, `bush_small`), còn `assets/src/decor/` lại chứa bộ "tấu hài phố" của
// ART_PROMPTS §9 (`red_plastic_stool`, `street_food_cart`, …). **Không tên nào trùng**, nên
// 543 vật trang trí trong 5 bản đồ không vẽ ra thứ gì suốt nhiều tuần.
//
// Luật mới, theo đúng chủ dự án chốt 31/10:
//  1. Mỗi vật trang trí gắn **tag khu** — nó hợp chủ đề khu nào.
//  2. Mapgen **chỉ đặt cái có art thật**. Chưa có art thì không đặt, chứ không đặt rồi để
//     client bỏ qua trong im lặng.
//  3. Cái hợp chủ đề mà chưa có art thì có **job artgen**; art về là mapgen tự đặt, không
//     phải sửa thêm dòng nào.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"bladeascend/server/internal/config"
)

// decorZones gắn mỗi vật trang trí với các khu nó hợp chủ đề.
//
// Khoá là tên tệp trong `assets/src/decor/<tên>.png`. Bảng này là NGUỒN DUY NHẤT quyết định
// khu nào rải cái gì — đừng gõ danh sách decor trực tiếp vào `zones.go` nữa.
//
// `wall_border` KHÔNG có mặt ở đây: nó là texture tường biên lát liền mạch, không phải vật
// rải rác. Đặt nó như decor thì được một đoạn tường lơ lửng giữa đồng.
var decorZones = map[string][]string{
	// ——— Đã có art (ART_PROMPTS §9) ———
	// Bộ "tấu hài phố": đúng chất Phàm Trấn. Thiên Đình cũng nhận vài món vì khu trung tâm
	// có một góc phố (bảng §8.1 bước 3, góc "Phàm Trấn" và "Hắc Thị").
	"red_plastic_stool":   {"pham_tran", "thien_dinh"},
	"old_bicycle_leaning": {"pham_tran"},
	"pile_of_watermelons": {"pham_tran"},
	"street_food_cart":    {"pham_tran", "thien_dinh"},
	"banana_tree":         {"pham_tran"},
	"rusty_water_jar":     {"pham_tran", "truc_lam"},
	"broken_tv":           {"pham_tran"},
	"lazy_sleeping_dog":   {"pham_tran", "thien_dinh"},
	"bush_with_flowers":   {"thien_dinh", "truc_lam", "linh_son_coc", "pham_tran"},
	// Băng Nguyệt Hồ cũng nhận đèn đá: `rock_small` và `pine` của nó đều chưa có art, mà
	// `TestEveryZoneGetsSomeDecor` đòi mỗi khu phải đặt được ít nhất MỘT loại — "chỉ đặt cái
	// có art" có một cách thoả mãn rất tệ là không đặt gì cả. Đèn đá giữa hồ băng dưới ánh
	// trăng cũng đúng chất tên khu hơn là bụi hoa.
	"stone_lantern": {"thien_dinh", "truc_lam", "linh_son_coc", "bang_nguyet_ho"},

	// ——— Hợp chủ đề nhưng CHƯA CÓ ART → có job artgen, chưa có thì không đặt ———
	"rock_small": {"linh_son_coc", "bang_nguyet_ho", "thien_dinh"},
	"pine":       {"linh_son_coc", "bang_nguyet_ho", "thien_dinh"},
	"bamboo":     {"truc_lam", "thien_dinh"},
	"bush_small": {"truc_lam", "thien_dinh"},
	"lantern":    {"pham_tran", "thien_dinh"},
	"well":       {"pham_tran", "thien_dinh"},
	"banner":     {"pham_tran", "thien_dinh"},
}

// decorPerZone là **tổng** số lần THỬ đặt trang trí của cả khu, chia đều cho các loại dùng
// được. `builder.decor` chỉ đặt được khi chỗ đó trống, nên con số thật luôn nhỏ hơn.
//
// Chia theo khu vì mật độ phải khác nhau: Trúc Lâm là rừng tre dày, Băng Nguyệt Hồ là mặt
// băng trống — rải cùng mật độ thì một bên rậm rạp, một bên như bãi rác giữa hồ.
//
// **Tổng, không phải mỗi loại.** Bản đầu tôi để "mỗi loại n lần" và Phàm Trấn nhảy lên 447
// vật chỉ vì nó hợp 9 loại; khi 7 ảnh còn thiếu về thì nó thành 12 loại và tự lên ~720. Mật
// độ bản đồ không được phụ thuộc vào việc hôm nay có bao nhiêu tệp art.
var decorPerZone = map[string]int{
	"thien_dinh":     120,
	"linh_son_coc":   140,
	"truc_lam":       180,
	"pham_tran":      180,
	"bang_nguyet_ho": 100,
}

// decorDirs là những chỗ có thể chứa art trang trí, thử theo thứ tự.
//
// `assets/src` là nguồn (artgen ghi vào đây). `assets/dist` là bản phát hành của
// `tools/asset-pipeline`, tên tệp có hash nên không tra theo tên trần được — vì vậy mapgen
// luôn soi `assets/src`. Hai chỗ luôn cùng một bộ TÊN, chỉ khác định dạng và hash.
func decorDirs() []string {
	return []string{config.FindRepoFile(filepath.Join("assets", "src", "decor"))}
}

// haveDecorArt đọc đĩa để biết vật trang trí nào CÓ art thật.
//
// Đọc đĩa chứ không gõ danh sách vào mã: gõ tay thì lần sau artgen sinh xong một ảnh mới,
// bản đồ vẫn không đặt nó, và không ai biết tại sao. Đọc đĩa thì art về là mapgen dùng ngay.
func haveDecorArt() map[string]bool {
	have := map[string]bool{}
	for _, dir := range decorDirs() {
		ents, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range ents {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".png") {
				continue
			}
			have[strings.TrimSuffix(e.Name(), ".png")] = true
		}
	}
	return have
}

// decorFor trả về danh sách vật trang trí đặt được ở khu `zoneID`: hợp chủ đề VÀ có art.
// Trả thêm danh sách hợp chủ đề mà thiếu art, để mapgen in ra chứ không bỏ qua trong im lặng.
func decorFor(zoneID string, have map[string]bool) (use, missing []string) {
	for name, zones := range decorZones {
		fits := false
		for _, z := range zones {
			if z == zoneID {
				fits = true
				break
			}
		}
		if !fits {
			continue
		}
		if have[name] {
			use = append(use, name)
		} else {
			missing = append(missing, name)
		}
	}
	// Sắp xếp để bản đồ sinh ra KHÔNG phụ thuộc thứ tự lặp map của Go. Thiếu dòng này thì
	// cùng một seed cho ra hai bản đồ khác nhau giữa hai lần chạy — và `go test` sẽ đỏ lúc
	// này, xanh lúc khác mà không ai hiểu vì sao.
	sort.Strings(use)
	sort.Strings(missing)
	return use, missing
}

// decorStepFor rải vật trang trí của một khu. Gọi từ hàm `build` của từng khu trong zones.go.
func (g *gen) decorStepFor(zoneID string) {
	use, missing := decorFor(zoneID, haveDecorArt())
	total := decorPerZone[zoneID]
	if total == 0 {
		total = 120
	}
	if len(use) > 0 {
		n := total / len(use)
		if n < 1 {
			n = 1
		}
		for _, sprite := range use {
			g.b.decor(sprite, n)
		}
	}
	// Ghi vào builder, không giữ riêng trong `gen`: `main` chỉ nắm `*builder` sau khi sinh,
	// nên để trong `gen` là cảnh báo không bao giờ in ra — đúng kiểu im lặng Q27-2 muốn xoá.
	g.b.decorSkipped = append(g.b.decorSkipped, missing...)
}

// uniqueSorted bỏ trùng rồi sắp xếp — dùng cho danh sách tên in ra cho người đọc.
func uniqueSorted(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

// decorJobsFor trả về tên job artgen sinh ra những vật trang trí này.
//
// Phải khớp `cmd/artgen/catalog.go`: bộ phố nằm trong job `decor_sheet`, bộ cây cỏ / đá /
// giếng nằm trong `decor_nature`. Hai job sinh cả một tấm sheet nên một job lo nhiều tên.
func decorJobsFor(names []string) []string {
	nature := map[string]bool{}
	for _, n := range []string{"rock_small", "pine", "bamboo", "bush_small", "lantern", "well", "banner"} {
		nature[n] = true
	}
	jobs := map[string]bool{}
	for _, n := range names {
		if nature[n] {
			jobs["decor_nature"] = true
		} else {
			jobs["decor_sheet"] = true
		}
	}
	out := make([]string, 0, len(jobs))
	for j := range jobs {
		out = append(out, j)
	}
	sort.Strings(out)
	return out
}
