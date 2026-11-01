package store

import (
	"math"
	"testing"
	"time"
)

// Phân bố ngôn ngữ 7 ngày đếm theo NGƯỜI (guest_id), không theo mạng; bỏ mạng cũ và mạng
// chưa ghi ngôn ngữ (spec v6.34 §9.4).
func TestLangShares(t *testing.T) {
	s, err := Open(":memory:", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	recent, old := now.Add(-time.Hour).Unix(), now.Add(-8*24*time.Hour).Unix()
	for _, l := range []Life{
		{GuestID: "a", Lang: "vi", EndedAt: recent},
		{GuestID: "a", Lang: "vi", EndedAt: recent}, // cùng người chơi 2 mạng = 1 người
		{GuestID: "b", Lang: "vi", EndedAt: recent},
		{GuestID: "c", Lang: "pt", EndedAt: recent},
		{GuestID: "d", Lang: "th", EndedAt: old},  // ngoài 7 ngày
		{GuestID: "e", Lang: "", EndedAt: recent}, // client cũ
	} {
		if err := s.Record(l); err != nil {
			t.Fatal(err)
		}
	}
	got := s.LangShares(now)
	if len(got) != 2 || math.Abs(got["vi"]-2.0/3) > 1e-9 || math.Abs(got["pt"]-1.0/3) > 1e-9 {
		t.Fatalf("LangShares = %v, cần vi 2/3, pt 1/3", got)
	}
}
