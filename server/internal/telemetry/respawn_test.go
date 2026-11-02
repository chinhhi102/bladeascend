package telemetry

import (
	"strings"
	"testing"
	"time"
)

// Nhấp nháy entity phải đi về kèm NGUYÊN NHÂN, không chỉ tổng.
//
// Chuyện đã xảy ra: báo cáo 31/10 ghi "1.357 lần entity respawn trong 2s (ngân sách 0)".
// Đi tìm lại ngày 2/11 thì dữ liệu người thật chỉ có TỔNG, nên không trả lời được câu duy
// nhất đáng hỏi — do `Despawn` của server, do luật hết hạn 2 giây ở client, hay do quái.
// Ba nguyên nhân cần ba cách chữa khác hẳn nhau, và phải dựng lại cả chuỗi build theo thời
// gian mới lần ra được là bản sửa nào đã đóng nó.
//
// Bảng chia nguyên nhân lúc đó ĐÃ tồn tại trong client (`world.respawnBy`) nhưng chỉ
// `tools/browsertest/respawn-wander.mjs` đọc — tức chỉ thấy được trên máy dev, đúng chỗ lỗi
// gần như không bao giờ xảy ra.
func TestNhapNhayGhiKemNguyenNhan(t *testing.T) {
	st := open(t)
	now := float64(time.Now().UnixMilli())
	st.AddSession(&Session{ID: "s1", Build: "b1", GPU: "NVIDIA RTX"}, "ip")
	st.AddSample(&Sample{
		Session: "s1", TS: now, FPSAvg: 59, FPSP1: 50,
		Respawn2s: 7, RespawnDs: 4, RespawnHethan: 2, RespawnQuai: 1,
	})
	flush(t, st)

	d, err := Load(st.DB(), Opts{From: 0, To: time.Now().Unix() + 60, Label: "test"})
	if err != nil {
		t.Fatalf("đọc: %v", err)
	}
	if len(d.Samples) != 1 {
		t.Fatalf("mong 1 mẫu, nhận %d", len(d.Samples))
	}
	got := d.Samples[0]
	if got.Respawn2s != 7 || got.RespawnDs != 4 || got.RespawnHethan != 2 || got.RespawnQuai != 1 {
		t.Errorf("bảng chia nguyên nhân không đi hết đường: tổng=%d ds=%d hethan=%d quai=%d, mong 7/4/2/1",
			got.Respawn2s, got.RespawnDs, got.RespawnHethan, got.RespawnQuai)
	}

	// Cảnh báo phải NÓI RA nguyên nhân. Một cảnh báo chỉ có tổng buộc người đọc đi dựng lại
	// cả chuỗi build mới biết chữa ở đâu — đúng việc đã mất nửa ngày 2/11.
	md := d.Markdown()
	if !strings.Contains(md, "7 lần entity respawn") {
		t.Fatalf("thiếu cảnh báo respawn:\n%s", md)
	}
	for _, phai := range []string{"Despawn 4", "hết hạn 2s 2", "quái 1"} {
		if !strings.Contains(md, phai) {
			t.Errorf("cảnh báo thiếu %q — có tổng mà không có nguyên nhân thì không chữa được:\n%s",
				phai, md)
		}
	}
}

// DB cũ KHÔNG có ba cột mới. `CREATE TABLE IF NOT EXISTS` không đụng bảng đã tồn tại, nên
// thiếu `ALTER TABLE` là mọi lần ghi mẫu sẽ lỗi — và telemetry nuốt lỗi ghi (luật 2 của gói
// này: đầy thì vứt, không chặn), nên nó sẽ hỏng HOÀN TOÀN IM LẶNG.
func TestMoLaiDBCuVanGhiDuocMauMoi(t *testing.T) {
	dir := t.TempDir() + "/cu.db"

	st1, err := Open(dir, "b1")
	if err != nil {
		t.Fatalf("mở lần 1: %v", err)
	}
	// Giả lập DB "cũ": bỏ hẳn ba cột vừa thêm.
	for _, c := range []string{"respawn_ds", "respawn_hethan", "respawn_quai"} {
		if _, err := st1.DB().Exec("ALTER TABLE samples DROP COLUMN " + c); err != nil {
			t.Fatalf("bỏ cột %s: %v", c, err)
		}
	}
	_ = st1.Close()

	// Mở lại: `Open` phải tự thêm cột về.
	st2, err := Open(dir, "b1")
	if err != nil {
		t.Fatalf("mở lại DB cũ: %v", err)
	}
	t.Cleanup(func() { _ = st2.Close() })
	st2.AddSession(&Session{ID: "s1", Build: "b1"}, "ip")
	st2.AddSample(&Sample{Session: "s1", TS: float64(time.Now().UnixMilli()), Respawn2s: 3, RespawnDs: 3})
	st2.Sync()

	d, err := Load(st2.DB(), Opts{From: 0, To: time.Now().Unix() + 60, Label: "test"})
	if err != nil {
		t.Fatalf("đọc: %v", err)
	}
	if len(d.Samples) != 1 {
		t.Fatalf("mẫu không ghi được vào DB cũ sau khi nâng cấp schema: nhận %d mẫu", len(d.Samples))
	}
	if d.Samples[0].RespawnDs != 3 {
		t.Errorf("cột mới thêm bằng ALTER TABLE không nhận giá trị: ds=%d, mong 3", d.Samples[0].RespawnDs)
	}
}
