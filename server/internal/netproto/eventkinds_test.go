package netproto

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// Bảng `EventKinds` phải phủ MỌI loại sự kiện server thật sự bắn ra.
//
// Vì sao kiểm bằng cách đọc mã nguồn chứ không bằng một danh sách chép tay: một danh sách
// chép tay sẽ cũ đi đúng lúc không ai để ý — ai đó thêm `Event{Kind: "ThingMoi"}` trong
// `skills.go`, không ai nhớ sửa bảng, và từ đó mọi sự kiện loại đó đi ra dây mang chỉ số
// 255. Client không ném lỗi; nó chỉ im lặng không vẽ gì. Test này đọc thẳng `internal/game`
// và `internal/room`, nên nó đỏ ngay trong lần chạy test đầu tiên sau khi thêm loại mới.
func TestEventKindsCoversEveryKindTheServerEmits(t *testing.T) {
	re := regexp.MustCompile(`Event\{Kind:\s*"([A-Za-z_]+)"`)
	found := map[string]string{} // loại → tệp đầu tiên thấy nó

	for _, pkg := range []string{"game", "room"} {
		dir := filepath.Join("..", pkg)
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("đọc %s: %v", dir, err)
		}
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
				continue
			}
			b, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				t.Fatal(err)
			}
			for _, m := range re.FindAllStringSubmatch(string(b), -1) {
				if _, seen := found[m[1]]; !seen {
					found[m[1]] = pkg + "/" + e.Name()
				}
			}
		}
	}
	if len(found) == 0 {
		t.Fatal("không quét được loại sự kiện nào — biểu thức tìm kiếm hỏng, " +
			"và một test không tìm thấy gì thì luôn xanh")
	}

	var missing []string
	for k, where := range found {
		if EventKindIndex(k) == EventKindUnknown {
			missing = append(missing, k+" ("+where+")")
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("%d loại sự kiện server bắn ra mà KHÔNG có trong EventKinds — chúng sẽ đi ra\n"+
			"dây với chỉ số 255 và client im lặng không vẽ gì:\n  %s\n"+
			"Thêm vào CUỐI bảng trong eventkinds.go (chèn giữa là dịch chỉ số của mọi loại sau nó).",
			len(missing), strings.Join(missing, "\n  "))
	}

	// Chiều ngược lại chỉ là CẢNH BÁO: một loại trong bảng mà mã không còn bắn ra nữa thì
	// vô hại (tốn một ô trong bảng), và xoá nó đi lại dịch chỉ số của mọi loại sau — tệ hơn
	// hẳn. Nên chỉ ghi ra để biết.
	for _, k := range EventKinds {
		if _, ok := found[k]; !ok {
			t.Logf("lưu ý: %q có trong bảng nhưng không tệp nào bắn ra (giữ lại cho khỏi dịch chỉ số)", k)
		}
	}
}

// Chỉ số phải ỔN ĐỊNH: nó là một phần của giao thức trên dây.
func TestEventKindIndexIsStable(t *testing.T) {
	// Vài mốc rải đều bảng. Nếu ai đó chèn một loại vào giữa, ít nhất một dòng dưới đây đỏ.
	for _, c := range []struct {
		kind string
		want uint8
	}{
		{"AltarBurst", 0},
		{"Blinded", 2},
		{"Hit", 10},
		{"RingFlip", 21},
		{"TienKhiPicked", 31},
	} {
		if got := EventKindIndex(c.kind); got != c.want {
			t.Errorf("EventKindIndex(%q) = %d, muốn %d — có ai chèn loại mới vào GIỮA bảng?\n"+
				"Chỉ số là giao thức: chèn giữa làm client vẽ nhầm hiệu ứng mà không báo lỗi.",
				c.kind, got, c.want)
		}
	}
	if EventKindIndex("KhongCoThat") != EventKindUnknown {
		t.Error("loại lạ phải trả EventKindUnknown")
	}
	if len(EventKinds) > EventKindUnknown {
		t.Fatalf("bảng có %d loại, vượt mốc %d dành cho 'không biết'", len(EventKinds), EventKindUnknown)
	}
}
