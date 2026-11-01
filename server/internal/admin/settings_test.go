package admin

import "testing"

// Cấu hình admin (GAME_SPEC v6.28 §9.5) phải sống qua deploy: ghi rồi đọc lại, ghi đè được.
func TestSettingRoundTrip(t *testing.T) {
	s := testStore(t)
	if _, ok := s.Setting("room_settings"); ok {
		t.Fatal("chưa đặt mà đã có giá trị")
	}
	for _, v := range []string{`{"max_humans":20}`, `{"max_humans":30}`} {
		if err := s.SetSetting("room_settings", v); err != nil {
			t.Fatal(err)
		}
		if got, ok := s.Setting("room_settings"); !ok || got != v {
			t.Fatalf("đọc lại %q, muốn %q", got, v)
		}
	}
}
