package store

import (
	"testing"
	"time"
)

func TestTopByBoardAndPeriod(t *testing.T) {
	s, err := Open(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.Local) // a Wednesday
	old := now.AddDate(0, 0, -10).Unix()
	lives := []Life{
		{GuestID: "g1", Name: "A", Score: 500, Hits: 1, EndedAt: now.Unix()},
		{GuestID: "g2", Name: "B", Score: 900, Hits: 0, EndedAt: now.AddDate(0, 0, -1).Unix()},
		{GuestID: "g3", Name: "C", Score: 9000, Hits: 7, EndedAt: old},
	}
	for _, l := range lives {
		if err := s.Record(l); err != nil {
			t.Fatal(err)
		}
	}
	check := func(period, board string, want ...string) {
		t.Helper()
		rows, err := s.Top(period, board, now)
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) != len(want) {
			t.Fatalf("%s/%s: %d rows, want %v", period, board, len(rows), want)
		}
		for i, w := range want {
			if rows[i].Name != w {
				t.Fatalf("%s/%s row %d = %s, want %s", period, board, i, rows[i].Name, w)
			}
		}
	}
	check("day", "score", "A")
	check("week", "score", "B", "A")
	check("all", "score", "C", "B", "A")
	check("all", "hits", "C", "A", "B")
	if _, err := s.Top("year", "score", now); err == nil {
		t.Fatal("bad period accepted")
	}
	if _, err := s.Top("all", "exp", now); err == nil {
		t.Fatal("bảng không tồn tại vẫn được chấp nhận")
	}
}

// Spec §11.2: bảng Thần Tốc chỉ gồm những mạng đã đạt Đại Thừa, xếp theo thời gian tăng dần.
func TestSpeedBoardOnlyReachedDaiThua(t *testing.T) {
	s, err := Open(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.Local)
	for _, l := range []Life{
		{GuestID: "g1", Name: "Nhanh", DaiThuaSecs: 300, EndedAt: now.Unix()},
		{GuestID: "g2", Name: "Chậm", DaiThuaSecs: 900, EndedAt: now.Unix()},
		{GuestID: "g3", Name: "ChưaĐạt", DaiThuaSecs: 0, Score: 99999, EndedAt: now.Unix()},
	} {
		if err := s.Record(l); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := s.Top("all", "speed", now)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("%d dòng, muốn 2 (bỏ mạng chưa lên Đại Thừa)", len(rows))
	}
	if rows[0].Name != "Nhanh" || rows[1].Name != "Chậm" {
		t.Fatalf("thứ tự sai: %s, %s", rows[0].Name, rows[1].Name)
	}
}

// Mở lại một database tạo từ bản cũ (chưa có cột score/hits/…) phải tự nâng cấp được.
func TestMigratesOldDatabase(t *testing.T) {
	path := t.TempDir() + "/old.db"
	s, err := Open(path, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Record(Life{GuestID: "g1", Name: "Cũ", Score: 10, EndedAt: 1}); err != nil {
		t.Fatal(err)
	}
	s.Close()
	again, err := Open(path, 0) // chạy lại migration trên database đã nâng cấp
	if err != nil {
		t.Fatalf("mở lại database đã có sẵn: %v", err)
	}
	defer again.Close()
	rows, err := again.Top("all", "score", time.Now())
	if err != nil || len(rows) != 1 || rows[0].Name != "Cũ" {
		t.Fatalf("đọc lại dữ liệu cũ hỏng: %v, %v", rows, err)
	}
}
