package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Qua tunnel thì MỌI người trên Internet đều có địa chỉ 127.0.0.1.
//
// Port forwarding của VS Code nối từ dịch vụ của Microsoft vào `127.0.0.1` của máy này, nên
// nhìn từ server mọi request đều là loopback. Mọi cửa khoá bằng "chỉ localhost" — `/metrics`,
// `/debug/pprof` — bỗng mở toang. `EXPOSED=true` là công tắc tắt niềm tin đó.
func TestExposedThiKhongTinLocalhostNua(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	r.RemoteAddr = "127.0.0.1:54321"

	t.Setenv("EXPOSED", "")
	if !isLocalRequest(r) {
		t.Fatal("bình thường thì 127.0.0.1 phải được coi là nội bộ")
	}

	t.Setenv("EXPOSED", "true")
	if isLocalRequest(r) {
		t.Fatal("EXPOSED=true mà vẫn tin 127.0.0.1 — qua tunnel là cả Internet đọc được /metrics")
	}
	// Địa chỉ LAN thật cũng không còn được tin: tunnel có thể đặt bất kỳ X-Forwarded-For nào.
	r.RemoteAddr = "192.168.1.50:1234"
	if isLocalRequest(r) {
		t.Fatal("EXPOSED=true thì địa chỉ LAN cũng không được tin")
	}
}

// Công tắc phải đọc được ngay, không cần khởi động lại — một công tắc an toàn mà phải nhớ
// khởi động lại mới có tác dụng là một công tắc người ta quên bật.
func TestExposedDocNgay(t *testing.T) {
	t.Setenv("EXPOSED", "TRUE")
	if !exposed() {
		t.Fatal("EXPOSED=TRUE (hoa) phải nhận")
	}
	t.Setenv("EXPOSED", " true ")
	if !exposed() {
		t.Fatal("EXPOSED có khoảng trắng phải nhận")
	}
	t.Setenv("EXPOSED", "1")
	if exposed() {
		t.Fatal(`chỉ "true" mới bật — "1" phải KHÔNG bật, để không ai tưởng đã bật`)
	}
}
