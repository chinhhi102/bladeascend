package store

import (
	"math"
	"testing"
	"time"
)

func TestCharacterStats(t *testing.T) {
	s, err := Open(":memory:", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().Unix()
	for _, l := range []Life{
		{GuestID: "a", Name: "A", Character: "a_hon", Score: 600, SurviveSecs: 120, SkillCasts: 4, PeakLevel: 25, EndedAt: now},
		{GuestID: "b", Name: "B", Character: "a_hon", Score: 0, SurviveSecs: 60, SkillCasts: 2, EndedAt: now},
		{GuestID: "c", Name: "C", Character: "be_ti", Score: 100, SurviveSecs: 10, EndedAt: now}, // < 30s: bỏ
	} {
		if err := s.Record(l); err != nil {
			t.Fatal(err)
		}
	}
	got, err := s.CharacterStats(now-60, func(l int) int { return l / 10 })
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Character != "a_hon" || got[0].Lives != 2 {
		t.Fatalf("got %+v", got)
	}
	if math.Abs(got[0].ScorePerMin-200) > 1e-9 || math.Abs(got[0].CastsPerMin-2) > 1e-9 || got[0].AvgSurvive != 90 {
		t.Fatalf("số sai: %+v", got[0])
	}
	if got[0].AvgPeakRealm != 1 || got[0].Casts != 6 { // (2 + 0) / 2 mạng
		t.Fatalf("cảnh giới TB %v / lần dùng %d, muốn 1 / 6", got[0].AvgPeakRealm, got[0].Casts)
	}
}
