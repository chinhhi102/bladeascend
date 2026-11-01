package room

import "testing"

// Cảnh báo admin (CHANGES v6.36 mục 5): đếm thời gian p95 > ngưỡng LIÊN TỤC theo bước 10 s,
// một cửa sổ dưới ngưỡng là về 0.
func TestTickHotSecs(t *testing.T) {
	var ts tickStats
	window := func(ms float64) {
		for i := 0; i < tickWindow; i++ {
			ts.add(ms, 3.5)
		}
	}
	for i := 0; i < 30; i++ {
		window(4.0)
	}
	if got := ts.hot(); got != 300 {
		t.Fatalf("30 cửa sổ 10 s vượt ngưỡng = 300 s, ra %.0f", got)
	}
	window(1.0)
	if got := ts.hot(); got != 0 {
		t.Fatalf("một cửa sổ dưới ngưỡng phải đặt lại 0, ra %.0f", got)
	}
}
