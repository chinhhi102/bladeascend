package telemetry

import "testing"

// Lỗi vẽ phải gom theo THÔNG BÁO BÊN TRONG, không phải theo nguyên văn JSON.
//
// Ngày 13/11, 95 lỗi vẽ thật của người chơi không hiện lấy một dòng trên trang "Lỗi" vì
// `draw_error` chưa nằm trong `ErrorKinds`. Thêm vào rồi thì phải gom đúng: `data` của nó
// chở `texMB`, `fps`, `zone`… đổi mỗi khung, nên gom theo nguyên văn là mỗi khung một nhóm.
func TestDrawErrorGomTheoThongBaoBenTrong(t *testing.T) {
	a := `{"err":"Error: shader and geometry incompatible, geometry missing the \"aLocal\" attribute","texMB":33,"fps":2,"zone":2}`
	b := `{"err":"Error: shader and geometry incompatible, geometry missing the \"aLocal\" attribute","texMB":46,"fps":60,"zone":0}`
	c := `{"err":"TypeError: x is not a function","texMB":33,"fps":2,"zone":2}`

	if ErrorSig("draw_error", a) != ErrorSig("draw_error", b) {
		t.Fatalf("cùng một lỗi, khác số liệu → phải CÙNG nhóm:\n  %s\n  %s",
			ErrorSig("draw_error", a), ErrorSig("draw_error", b))
	}
	if ErrorSig("draw_error", a) == ErrorSig("draw_error", c) {
		t.Fatal("hai lỗi khác hẳn nhau mà gom chung một nhóm")
	}
	// `data` cụt giữa chừng (trường `data` cắt ở 500 ký tự) vẫn phải ra một nhóm dùng được.
	if got := ErrorSig("draw_error", `{"err":"Error: abc`); got == "" {
		t.Fatal("JSON cụt → chữ ký rỗng")
	}
}

// Lỗi GIẢ của crashtest không được tính là lỗi (mục 4, 14/11).
func TestLoiGiaKhongNamTrongErrorKinds(t *testing.T) {
	for _, k := range ErrorKinds {
		if k == KindServerPanicFake || k == "draw_error_gia" || k == "client_sys_off_gia" {
			t.Fatalf("loại lỗi giả %q lọt vào ErrorKinds — bảng lỗi sẽ đếm cả lượt deploy", k)
		}
	}
	found := false
	for _, k := range ErrorKinds {
		if k == "draw_error" {
			found = true
		}
	}
	if !found {
		t.Fatal("draw_error KHÔNG nằm trong ErrorKinds — lỗi vẽ thật lại vô hình như ngày 13/11")
	}
}
