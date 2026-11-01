package telemetry

import (
	"strings"
	"testing"
	"time"
)

// Hai luật chủ dự án chốt 1/11, cả hai sinh ra từ cùng một sự cố:
//
// Báo cáo 31/10 kết luận "96% mẫu có FPS 1% dưới 30" và suýt nữa khởi động một đợt tối ưu
// lớn. Hai thứ làm con số đó vô nghĩa:
//
//  1. Nó gộp MỌI BUILD trong ngày. Hôm đó deploy 17 lần; các build chỉ có người thật cho
//     FPS p50 58–59, còn con số tổng bị kéo xuống bởi mấy build cuối.
//  2. Nó gộp cả PHIÊN DO MÁY ĐIỀU KHIỂN. 64 trong 67 phiên buổi tối là bài đo 10 tab chạy
//     trên đúng một máy — mười tab chia nhau một GPU thì FPS thấp là đương nhiên.
//
// Hai test dưới đây canh đúng hai luật đó.

func TestBaoCaoLoaiPhienDoMayDieuKhien(t *testing.T) {
	st := open(t)
	now := float64(time.Now().UnixMilli())

	// Một người thật chơi mượt, và hai "người" là tab của Playwright chạy ì ạch.
	st.AddSession(&Session{ID: "that", Browser: "Chrome", GPU: "NVIDIA RTX", Build: "b1"}, "ip1")
	st.AddSample(&Sample{Session: "that", TS: now, FPSAvg: 59, FPSP1: 48})

	for _, id := range []string{"bot1", "bot2"} {
		st.AddSession(&Session{ID: id, Browser: "Chrome", GPU: "ANGLE AMD", Build: "b1", Automated: true}, "ip2")
		st.AddSample(&Sample{Session: id, TS: now, FPSAvg: 12, FPSP1: 2})
	}
	flush(t, st)

	d, err := Load(st.DB(), Opts{From: 0, To: time.Now().Unix() + 60, Label: "test"})
	if err != nil {
		t.Fatalf("đọc: %v", err)
	}

	if len(d.Samples) != 1 {
		t.Fatalf("mẫu của máy phải bị loại: còn %d mẫu, mong 1", len(d.Samples))
	}
	if d.Samples[0].Session != "that" {
		t.Errorf("mẫu còn lại phải là của người thật, nhận %q", d.Samples[0].Session)
	}
	if d.AutoSessions != 2 || d.AutoSamples != 2 {
		t.Errorf("phải ĐẾM số bị loại để còn khai ra: %d phiên / %d mẫu, mong 2/2",
			d.AutoSessions, d.AutoSamples)
	}

	// Loại thì được, loại IM LẶNG thì không: báo cáo phải nói ra.
	md := d.Markdown()
	if !strings.Contains(md, "Đã LOẠI 2 phiên") {
		t.Errorf("báo cáo phải khai số phiên bị loại; không thấy trong:\n%s", md[:min(600, len(md))])
	}

	// Và cảnh báo phải tính trên người thật: FPS 59 thì không được kêu gì.
	if strings.Contains(md, "mẫu có FPS 1% dưới 30") {
		t.Errorf("cảnh báo FPS bị kéo xuống bởi phiên của máy:\n%s", md)
	}
}

// Soi MỘT phiên cụ thể thì giữ nguyên tất cả — kể cả phiên của bài đo. Lúc đó người ta
// đang gỡ lỗi chính phiên đó, lọc đi là lấy mất thứ họ vừa gọi tên.
func TestSoiMotPhienThiKhongLoc(t *testing.T) {
	st := open(t)
	now := float64(time.Now().UnixMilli())
	st.AddSession(&Session{ID: "bot1", Build: "b1", Automated: true}, "ip")
	st.AddSample(&Sample{Session: "bot1", TS: now, FPSAvg: 12, FPSP1: 2})
	flush(t, st)

	d, err := Load(st.DB(), Opts{Session: "bot1", Label: "một phiên"})
	if err != nil {
		t.Fatalf("đọc: %v", err)
	}
	if len(d.Samples) != 1 {
		t.Fatalf("soi một phiên thì không được lọc: còn %d mẫu, mong 1", len(d.Samples))
	}
}

func TestBangTheoBuildVaCanhBaoChiTinhBuildMoiNhat(t *testing.T) {
	st := open(t)
	base := time.Now().Add(-time.Hour)

	// Build cũ: chạy tệ. Build mới: đã sửa, chạy tốt. Nếu gộp hai build thì trung vị rơi
	// vào giữa và lần sửa biến mất khỏi báo cáo.
	st.AddSession(&Session{ID: "cu", Build: "20261031_1000_aaaaaaa", GPU: "NVIDIA RTX"}, "ip1")
	st.AddSession(&Session{ID: "moi", Build: "20261031_2000_bbbbbbb", GPU: "NVIDIA RTX"}, "ip2")
	for i := 0; i < 10; i++ {
		st.AddSample(&Sample{Session: "cu", TS: float64(base.UnixMilli()) + float64(i*1000), FPSAvg: 20, FPSP1: 5})
		st.AddSample(&Sample{Session: "moi", TS: float64(base.Add(30 * time.Minute).UnixMilli()), FPSAvg: 59, FPSP1: 50})
	}
	flush(t, st)

	d, err := Load(st.DB(), Opts{From: 0, To: time.Now().Unix() + 60, Label: "test"})
	if err != nil {
		t.Fatalf("đọc: %v", err)
	}
	if got := d.latestBuild(); got != "20261031_2000_bbbbbbb" {
		t.Fatalf("build mới nhất sai: %q", got)
	}
	if n := len(d.samplesOf(d.latestBuild())); n != 10 {
		t.Errorf("lọc theo build sai: %d mẫu, mong 10", n)
	}

	md := d.Markdown()
	if !strings.Contains(md, "## 8. Theo BUILD") {
		t.Fatalf("thiếu bảng theo build:\n%s", md)
	}
	if !strings.Contains(md, "← mới nhất") {
		t.Errorf("bảng phải chỉ rõ dòng nào là build mới nhất:\n%s", md)
	}
	// Build mới chạy 59 fps — không được có cảnh báo nào, dù build cũ rất tệ.
	if strings.Contains(md, "mẫu có FPS 1% dưới 30") {
		t.Errorf("cảnh báo phải tính trên build MỚI NHẤT, không gộp build cũ:\n%s", md)
	}
}
