package store

import (
	"testing"
	"time"
)

// Mốc trung vị Uy Danh — ADMIN_SPEC §6.1, QUESTIONS Q29.
//
// Chủ dự án chốt: trung vị trượt 24 giờ từ dữ liệu thật, rơi về bảng chốt sẵn khi chưa đủ
// 50 mạng chơi. Hai nhánh, và nhánh "chưa đủ dữ liệu" là nhánh hay bị quên nhất — nó chỉ
// chạy khi server mới mở hoặc vắng người, tức là đúng lúc không ai ngồi nhìn.

// realmOfTest: 10 tầng một cảnh giới, đủ để test không phụ thuộc luật cảnh giới thật.
func realmOfTest(level int) int { return level / 10 }

func seedScored(t *testing.T, s *Store, n int, realm int, perMin float64, endedAt int64) {
	t.Helper()
	for i := 0; i < n; i++ {
		// survive_secs 120 giây → score = perMin × 2.
		if err := s.Record(Life{
			GuestID: "g", Name: "P", Character: "a_hon",
			Score: perMin * 2, PeakLevel: realm*10 + 1, SurviveSecs: 120,
			Map: "m", EndedAt: endedAt,
		}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMedianFallsBackWhenTooFewLives(t *testing.T) {
	s := openTmp(t)
	now := time.Now()
	// 10 mạng chơi — dưới mốc 50.
	seedScored(t, s, 10, 3, 999, now.Add(-time.Hour).Unix())
	got, real := s.ScoreMedianPerMin(3, realmOfTest, now)
	if real {
		t.Error("10 mạng chơi mà đã tin là dữ liệu thật — trung vị của mẫu bé thì một người " +
			"chơi giỏi bất thường đủ kéo mốc lên gấp đôi")
	}
	if got != fallbackMedian[3] {
		t.Errorf("phải rơi về bảng chốt sẵn %.0f, được %.0f", fallbackMedian[3], got)
	}
}

func TestMedianUsesRealDataWhenEnough(t *testing.T) {
	s := openTmp(t)
	now := time.Now()
	at := now.Add(-time.Hour).Unix()
	// 60 mạng chơi ở cảnh giới 2, tất cả 100 Uy Danh/phút.
	seedScored(t, s, 60, 2, 100, at)
	got, real := s.ScoreMedianPerMin(2, realmOfTest, now)
	if !real {
		t.Fatal("60 mạng chơi mà vẫn rơi về bảng chốt sẵn")
	}
	if got < 99 || got > 101 {
		t.Errorf("trung vị phải ≈100, được %.1f", got)
	}
}

// Mạng chơi ĐÃ BỊ ẨN không được kéo mốc lên.
//
// Đây là chỗ dễ sai nhất của cả tính năng: những mạng bị ẩn CHÍNH LÀ những mạng nghi gian,
// nên để chúng trong phép tính nghĩa là mốc so sánh được nâng lên bởi đúng thứ nó phải phát
// hiện — và cờ sẽ im lặng dần theo thời gian.
func TestHiddenLivesDoNotRaiseTheMedian(t *testing.T) {
	s := openTmp(t)
	now := time.Now()
	at := now.Add(-time.Hour).Unix()
	seedScored(t, s, 60, 4, 100, at)
	clean, _ := s.ScoreMedianPerMin(4, realmOfTest, now)

	// Thêm 60 mạng "gian" gấp 50 lần, rồi ẩn hết.
	for i := 0; i < 60; i++ {
		if err := s.Record(Life{GuestID: "gian", Name: "X", Character: "a_hon",
			Score: 5000 * 2, PeakLevel: 41, SurviveSecs: 120, Map: "m",
			EndedAt: at + int64(i)}); err != nil {
			t.Fatal(err)
		}
		if err := s.Hide("gian", at+int64(i), "kiểm thử", "127.0.0.1", now); err != nil {
			t.Fatal(err)
		}
	}
	// Ép tính lại (cache 5 phút).
	s.refreshMedians(realmOfTest, now.Add(medianTTL+time.Second))
	dirty, _ := s.ScoreMedianPerMin(4, realmOfTest, now.Add(medianTTL+2*time.Second))
	if dirty > clean*1.5 {
		t.Errorf("mạng ĐÃ ẨN vẫn kéo mốc lên: %.0f → %.0f. Mốc so sánh đang được nâng bởi "+
			"chính những mạng nghi gian, và cờ sẽ im lặng dần.", clean, dirty)
	}
}

// Mỗi cảnh giới phải có đủ mẫu RIÊNG, không chỉ tổng đủ.
func TestMedianNeedsEnoughSamplesPerRealm(t *testing.T) {
	s := openTmp(t)
	now := time.Now()
	at := now.Add(-time.Hour).Unix()
	seedScored(t, s, 60, 1, 100, at) // cảnh giới 1: nhiều mẫu
	seedScored(t, s, 2, 9, 9999, at) // cảnh giới 9: đúng 2 mẫu, cao vô lý
	s.refreshMedians(realmOfTest, now)
	got, _ := s.ScoreMedianPerMin(9, realmOfTest, now)
	if got == 9999 {
		t.Error("2 mẫu mà đã lấy làm trung vị của cả một cảnh giới")
	}
	if got != fallbackMedian[9] {
		t.Errorf("thiếu mẫu riêng thì phải rơi về bảng chốt sẵn %.0f, được %.0f", fallbackMedian[9], got)
	}
}

// Mạng chơi cũ hơn 24 giờ không được tính.
func TestMedianIgnoresOldLives(t *testing.T) {
	s := openTmp(t)
	now := time.Now()
	seedScored(t, s, 60, 5, 100, now.Add(-48*time.Hour).Unix())
	s.refreshMedians(realmOfTest, now)
	if _, real := s.ScoreMedianPerMin(5, realmOfTest, now); real {
		t.Error("mạng chơi 48 giờ trước vẫn được tính vào trung vị 24 giờ")
	}
}
