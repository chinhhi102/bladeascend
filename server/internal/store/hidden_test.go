package store

import (
	"path/filepath"
	"testing"
	"time"
)

// Ẩn mạng chơi — ADMIN_SPEC §6.2.
//
// Điều cần giữ không phải "nút ẩn chạy được" mà là ba điều khó thấy hơn:
// bảng xếp hạng phải NGAY LẬP TỨC bớt tên đó (có cache ở giữa), dòng `lives` phải CÒN
// NGUYÊN, và bỏ ẩn phải trả lại đúng như cũ.

func openTmp(t *testing.T) *Store {
	t.Helper()
	// cacheTTL dài để test chứng minh được dropCache có tác dụng: nếu Hide quên xoá cache
	// thì Top() trả lại bản cũ và test đỏ. cacheTTL = 0 sẽ che mất đúng lỗi đó.
	s, err := Open(filepath.Join(t.TempDir(), "t.db"), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func seedLife(t *testing.T, s *Store, guest, name string, score float64, endedAt int64) {
	t.Helper()
	if err := s.Record(Life{
		GuestID: guest, Name: name, Character: "a_hon", Score: score,
		PeakExp: score, PeakLevel: 30, Kills: 3, SurviveSecs: 120, Map: "linh_son_coc",
		EndedAt: endedAt,
	}); err != nil {
		t.Fatal(err)
	}
}

func names(rows []Life) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.Name
	}
	return out
}

func TestHideRemovesFromBoardAndUnhideRestores(t *testing.T) {
	s := openTmp(t)
	now := time.Now()
	at := now.Add(-time.Minute).Unix()
	seedLife(t, s, "g-sach", "NgườiNgayThẳng", 100, at)
	seedLife(t, s, "g-gian", "KẻNghiGian", 9999, at)

	// Trước khi ẩn: kẻ nghi gian đứng đầu.
	top, err := s.Top("day", "score", now)
	if err != nil {
		t.Fatal(err)
	}
	if len(top) != 2 || top[0].Name != "KẻNghiGian" {
		t.Fatalf("trước khi ẩn muốn [KẻNghiGian NgườiNgayThẳng], được %v", names(top))
	}

	if err := s.Hide("g-gian", at, "Đạo Tổ trong 63 giây", "127.0.0.1", now); err != nil {
		t.Fatal(err)
	}

	// Sau khi ẩn: phải biến khỏi bảng NGAY, dù cacheTTL là một giờ.
	top, err = s.Top("day", "score", now)
	if err != nil {
		t.Fatal(err)
	}
	if len(top) != 1 || top[0].Name != "NgườiNgayThẳng" {
		t.Fatalf("sau khi ẩn muốn [NgườiNgayThẳng], được %v (cache chưa xoá?)", names(top))
	}

	// Dòng `lives` phải CÒN. Đây là lời hứa trung tâm của cả tính năng.
	rows, err := s.RecentLives(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("ẩn KHÔNG được xoá dữ liệu: muốn 2 dòng lives, còn %d", len(rows))
	}
	var found bool
	for _, r := range rows {
		if r.GuestID == "g-gian" {
			found = true
			if !r.Hidden {
				t.Error("dòng bị ẩn phải mang cờ hidden để admin bỏ ẩn được")
			}
			if r.Reason != "Đạo Tổ trong 63 giây" {
				t.Errorf("lý do ẩn mất: %q", r.Reason)
			}
		}
	}
	if !found {
		t.Error("mạng chơi bị ẩn đã biến khỏi lives — đó là xoá, không phải ẩn")
	}

	// Bỏ ẩn thì trả lại đúng như cũ.
	if err := s.Unhide("g-gian", at); err != nil {
		t.Fatal(err)
	}
	top, _ = s.Top("day", "score", now)
	if len(top) != 2 || top[0].Name != "KẻNghiGian" {
		t.Fatalf("bỏ ẩn phải trả lại nguyên trạng, được %v", names(top))
	}
}

func TestHideNeedsReason(t *testing.T) {
	s := openTmp(t)
	now := time.Now()
	at := now.Unix()
	seedLife(t, s, "g", "A", 10, at)
	for _, r := range []string{"", "   ", "\n\t "} {
		if err := s.Hide("g", at, r, "1.2.3.4", now); err != ErrHideReason {
			t.Errorf("ẩn với lý do %q phải bị từ chối, được %v", r, err)
		}
	}
	// Và nó thật sự không ẩn gì.
	if top, _ := s.Top("day", "score", now); len(top) != 1 {
		t.Error("từ chối rồi mà vẫn ẩn")
	}
}

// Ẩn hai lần không được làm bảng mất người: NOT EXISTS thay vì LEFT JOIN trong `query` là để
// chống đúng chuyện này, và chuyện đó chỉ lộ ra khi có nhiều hơn một dòng hidden khớp.
func TestHideTwiceKeepsBoardSizeAndUpdatesReason(t *testing.T) {
	s := openTmp(t)
	now := time.Now()
	at := now.Unix()
	seedLife(t, s, "g1", "A", 10, at)
	seedLife(t, s, "g2", "B", 20, at)

	if err := s.Hide("g1", at, "lý do đầu", "1.1.1.1", now); err != nil {
		t.Fatal(err)
	}
	if err := s.Hide("g1", at, "lý do sau", "2.2.2.2", now); err != nil {
		t.Fatalf("ẩn lại phải ghi đè lý do, không lỗi khoá trùng: %v", err)
	}
	hid, err := s.ListHidden()
	if err != nil {
		t.Fatal(err)
	}
	if len(hid) != 1 {
		t.Fatalf("ẩn hai lần cùng một mạng vẫn chỉ là một dòng, được %d", len(hid))
	}
	if hid[0].Reason != "lý do sau" || hid[0].ByIP != "2.2.2.2" {
		t.Errorf("muốn lý do mới nhất, được %q bởi %q", hid[0].Reason, hid[0].ByIP)
	}
	if hid[0].Name != "A" {
		t.Errorf("danh sách ẩn phải kèm tên để đọc được, được %q", hid[0].Name)
	}
	if top, _ := s.Top("day", "score", now); len(top) != 1 || top[0].Name != "B" {
		t.Errorf("muốn [B], được %v", names(top))
	}
}

// Lý do quá dài bị cắt, không bị từ chối: admin dán cả đoạn log vào là chuyện bình thường,
// và mất bản ghi kiểm toán vì lý do dài quá thì tệ hơn là lưu 300 chữ đầu.
func TestHideReasonTruncated(t *testing.T) {
	s := openTmp(t)
	now := time.Now()
	at := now.Unix()
	seedLife(t, s, "g", "A", 10, at)
	long := ""
	for i := 0; i < 400; i++ {
		long += "x"
	}
	if err := s.Hide("g", at, long, "1.1.1.1", now); err != nil {
		t.Fatal(err)
	}
	hid, _ := s.ListHidden()
	if len(hid) != 1 || len(hid[0].Reason) != HideReasonMax {
		t.Fatalf("muốn lý do cắt còn %d, được %d dòng / dài %d", HideReasonMax, len(hid), len(hid[0].Reason))
	}
}
