package store

import (
	"testing"
	"time"
)

func TestTopByMetricAndPeriod(t *testing.T) {
	s, err := Open(":memory:", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.Local) // a Wednesday
	old := now.AddDate(0, 0, -10).Unix()
	lives := []Life{
		{GuestID: "g1", Name: "A", PeakExp: 500, Kills: 1, EndedAt: now.Unix()},
		{GuestID: "g2", Name: "B", PeakExp: 900, Kills: 0, EndedAt: now.AddDate(0, 0, -1).Unix()},
		{GuestID: "g3", Name: "C", PeakExp: 9000, Kills: 7, EndedAt: old},
	}
	for _, l := range lives {
		if err := s.Record(l); err != nil {
			t.Fatal(err)
		}
	}
	check := func(period, metric string, want ...string) {
		t.Helper()
		rows, err := s.Top(period, metric, now)
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) != len(want) {
			t.Fatalf("%s/%s: %d rows, want %v", period, metric, len(rows), want)
		}
		for i, w := range want {
			if rows[i].Name != w {
				t.Fatalf("%s/%s row %d = %s, want %s", period, metric, i, rows[i].Name, w)
			}
		}
	}
	check("day", "exp", "A")
	check("week", "exp", "B", "A")
	check("all", "exp", "C", "B", "A")
	check("all", "kills", "C", "A", "B")
	if _, err := s.Top("year", "exp", now); err == nil {
		t.Fatal("bad period accepted")
	}
}
