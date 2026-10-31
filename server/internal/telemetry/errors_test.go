package telemetry

import (
	"strings"
	"testing"
	"time"
)

// Gom nhóm lỗi — ADMIN_SPEC §3.
//
// Hai điều phải đúng, và cả hai đều đã sai trong đời thật nếu làm cẩu thả:
//  1. cùng một lỗi ở HAI BẢN BUILD phải là MỘT nhóm. Tên tệp có hash (`index-CzSoRTTv.js`)
//     và số dòng đổi mỗi lần build, nên giữ số trong chữ ký là không bao giờ thấy được
//     "lỗi này đã có từ tuần trước".
//  2. xếp theo SỐ NGƯỜI trước, không theo số lần. Một lỗi vòng vẽ bắn 3 000 lần cho đúng
//     một người đang nghịch sẽ che mất một lỗi chỉ xảy ra một lần nhưng với mười người.

func TestErrorSigIgnoresBuildHashAndLineNumbers(t *testing.T) {
	a := "Uncaught TypeError: Cannot read properties of null (reading 'texture')\n" +
		"    at a_.updateBody (index-CzSoRTTv.js:896:28935)"
	b := "Uncaught TypeError: Cannot read properties of null (reading 'texture')\n" +
		"    at a_.updateBody (index-CZiHOvTh.js:901:31002)"
	if errorSig("error", a) != errorSig("error", b) {
		t.Errorf("cùng một lỗi ở hai bản build phải gộp một nhóm:\n  %q\n  %q",
			errorSig("error", a), errorSig("error", b))
	}
	// Nhưng hai lỗi KHÁC NHAU thì không được gộp.
	c := "Uncaught TypeError: Cannot read properties of null (reading 'parent')"
	if errorSig("error", a) == errorSig("error", c) {
		t.Error("hai lỗi khác nhau bị gộp làm một")
	}
	// Chữ ký chỉ lấy DÒNG ĐẦU: vết gọi hàm khác nhau giữa hai lần vẫn là một lỗi.
	if strings.Contains(errorSig("error", a), "updateBody") {
		t.Errorf("chữ ký không được chứa vết gọi hàm: %q", errorSig("error", a))
	}
}

func TestRecentErrorsGroupsAndRanksByPeopleNotCount(t *testing.T) {
	s, err := Open(t.TempDir()+"/t.db", "test-build")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Now()
	add := func(sess, kind, data string, ago time.Duration) {
		s.AddEvent(&Event{Session: sess, Kind: kind, Data: data,
			TS: float64(now.Add(-ago).UnixMilli())})
	}
	for i := 0; i < 3; i++ {
		s.AddSession(&Session{ID: string(rune('a'+i)), Browser: "Safari/18",
			OS: "iPhone OS 18_7", Mobile: true, Build: "b1"}, "iphash")
	}
	// Lỗi A: một người, 300 lần (vòng vẽ).
	for i := 0; i < 300; i++ {
		add("a", "error", "Uncaught TypeError: x is null\n at f (index-AAA.js:1:2)", time.Minute)
	}
	// Lỗi B: ba người, mỗi người 1 lần — ở ba bản build khác nhau.
	add("a", "error", "Uncaught RangeError: quá tay\n at g (index-AAA.js:9:9)", 2*time.Minute)
	add("b", "error", "Uncaught RangeError: quá tay\n at g (index-BBB.js:11:40)", 3*time.Minute)
	add("c", "error", "Uncaught RangeError: quá tay\n at g (index-CCC.js:77:1)", 4*time.Minute)
	s.Sync()

	g, err := s.RecentErrors(time.Hour, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(g) != 2 {
		for _, x := range g {
			t.Logf("  nhóm: %q × %d (%d người)", x.Sig, x.Count, x.Sessions)
		}
		t.Fatalf("ba bản build của cùng một lỗi phải gộp: muốn 2 nhóm, được %d", len(g))
	}
	if g[0].Sessions != 3 {
		t.Errorf("nhóm đầu phải là lỗi dính NHIỀU NGƯỜI nhất (3), được %d người / %d lần",
			g[0].Sessions, g[0].Count)
	}
	if g[1].Count != 300 {
		t.Errorf("nhóm sau phải là lỗi 300 lần của một người, được %d lần", g[1].Count)
	}
	if g[0].Mobile != 3 {
		t.Errorf("đếm người dùng điện thoại sai: muốn 3, được %d", g[0].Mobile)
	}
	if len(g[0].Devices) != 1 || g[0].Devices[0] != "Safari / iPhone" {
		t.Errorf("tên thiết bị phải gọn và gộp được, được %v", g[0].Devices)
	}
}

// Sự kiện chở JSON số liệu gộp theo LOẠI, không theo nội dung.
//
// Đo trên dữ liệu thật ngày 31/10: 2 người chạm trần bộ nhớ hiện thành 4 dòng, chỉ vì
// "q":"Thấp" khác "q":"Vừa" và một sự kiện có data rỗng.
func TestNonErrorKindsGroupByKindAlone(t *testing.T) {
	s, err := Open(t.TempDir()+"/t.db", "b")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Now()
	for i, d := range []string{`{"mb":216,"q":"Thấp"}`, `{"mb":214,"q":"Vừa"}`, ""} {
		s.AddEvent(&Event{Session: string(rune('a' + i)), Kind: "gpu_pressure", Data: d,
			TS: float64(now.Add(-time.Minute).UnixMilli())})
	}
	s.Sync()
	g, err := s.RecentErrors(time.Hour, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(g) != 1 {
		t.Fatalf("chạm trần bộ nhớ là MỘT vấn đề: muốn 1 nhóm, được %d", len(g))
	}
	if g[0].Sessions != 3 || g[0].Count != 3 {
		t.Errorf("muốn 3 người / 3 lần, được %d / %d", g[0].Sessions, g[0].Count)
	}
}

// Lỗi ngoài khoảng thời gian hỏi thì không được lọt vào.
func TestRecentErrorsRespectsWindow(t *testing.T) {
	s, err := Open(t.TempDir()+"/t.db", "b")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Now()
	s.AddEvent(&Event{Session: "a", Kind: "error", Data: "cũ",
		TS: float64(now.Add(-48 * time.Hour).UnixMilli())})
	s.AddEvent(&Event{Session: "a", Kind: "error", Data: "mới",
		TS: float64(now.Add(-time.Minute).UnixMilli())})
	s.Sync()
	g, err := s.RecentErrors(24*time.Hour, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(g) != 1 || g[0].Sample != "mới" {
		t.Fatalf("muốn chỉ lỗi trong 24 giờ, được %d nhóm %+v", len(g), g)
	}
}
