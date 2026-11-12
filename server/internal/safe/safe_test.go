package safe

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDoBatPanicVaBaoRaNgoai(t *testing.T) {
	var got Info
	OnPanic(func(i Info) { got = i })
	defer OnPanic(nil)

	ran := false
	if p := Do("tick", func() string { return "khu Thiên Đình" }, func() { panic("vỡ rồi") }); !p {
		t.Fatal("Do phải báo CÓ panic")
	}
	if p := Do("tick", nil, func() { ran = true }); p {
		t.Fatal("Do không được báo panic khi fn chạy bình thường")
	}
	if !ran {
		t.Fatal("fn không chạy")
	}
	if got.Value != "vỡ rồi" || got.Boundary != "tick" || got.Detail != "khu Thiên Đình" {
		t.Fatalf("thông tin panic sai: %+v", got)
	}
	if !strings.Contains(got.Stack, "safe_test.go") {
		t.Fatalf("vết gọi phải trỏ về chỗ panic, có: %q", got.Stack)
	}
	// Vết gọi không được chứa khung của chính gói `safe` — chúng che mất dòng đáng đọc.
	if strings.Contains(got.Stack, "safe.Do") {
		t.Fatalf("vết gọi còn khung của gói safe: %q", got.Stack)
	}
}

func TestTextDongDauLaChuKy(t *testing.T) {
	i := Info{Boundary: "goi-tin", Value: "index out of range", Detail: "người 7", Stack: "a\nb"}
	first := strings.SplitN(i.Text(), "\n", 2)[0]
	if first != "index out of range @ goi-tin" {
		t.Fatalf("dòng đầu sai: %q", first)
	}
}

func TestDetailPanicKhongLamChetLopBat(t *testing.T) {
	var got Info
	OnPanic(func(i Info) { got = i })
	defer OnPanic(nil)
	// Hàm dựng bối cảnh cũng chạy trong lúc đang xử lý panic (trạng thái có thể đã hỏng).
	// Nó panic thì lớp bắt phải nuốt, không được để panic thứ hai thoát ra.
	Do("tick", func() string { panic("bối cảnh cũng vỡ") }, func() { panic("lỗi gốc") })
	if got.Value != "lỗi gốc" || !strings.Contains(got.Detail, "không lấy được bối cảnh") {
		t.Fatalf("%+v", got)
	}
}

func TestHTTPTraVe500VaGhiLai(t *testing.T) {
	var got Info
	OnPanic(func(i Info) { got = i })
	defer OnPanic(nil)
	h := HTTP(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("handler vỡ") }))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/api/rooms", nil))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("mã trả về %d", w.Code)
	}
	if got.Boundary != "http" || got.Detail != "GET /api/rooms" {
		t.Fatalf("%+v", got)
	}
}

func TestHTTPKhongNuotErrAbortHandler(t *testing.T) {
	n := 0
	OnPanic(func(Info) { n++ })
	defer OnPanic(nil)
	h := HTTP(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic(http.ErrAbortHandler) }))
	func() {
		defer func() {
			if r := recover(); r != http.ErrAbortHandler {
				t.Fatalf("ErrAbortHandler phải được ném tiếp cho net/http, có %v", r)
			}
		}()
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/assets/x.webp", nil))
	}()
	if n != 0 {
		t.Fatal("client tự ngắt giữa chừng không phải lỗi, không được ghi vào trang Lỗi")
	}
}

func TestGoBatPanicTrongGoroutine(t *testing.T) {
	done := make(chan Info, 1)
	OnPanic(func(i Info) { done <- i })
	defer OnPanic(nil)
	Go("nen", nil, func() { panic("goroutine nền vỡ") })
	select {
	case i := <-done:
		if i.Value != "goroutine nền vỡ" {
			t.Fatalf("%+v", i)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("không nhận được báo cáo panic")
	}
}

func TestRateCuaSoTruot(t *testing.T) {
	r := NewRate(3, time.Minute)
	t0 := time.Unix(1_700_000_000, 0)
	if r.Hit(t0) || r.Hit(t0.Add(time.Second)) {
		t.Fatal("chưa đủ 3 lần mà đã báo")
	}
	if !r.Hit(t0.Add(2 * time.Second)) {
		t.Fatal("lần thứ 3 trong cửa sổ phải báo")
	}
	// Cửa sổ TRƯỢT, không phải bộ đếm reset theo phút: 2 lần cuối phút này + 2 lần đầu
	// phút sau vẫn phải chạm ngưỡng.
	r2 := NewRate(3, time.Minute)
	r2.Hit(t0.Add(50 * time.Second))
	r2.Hit(t0.Add(59 * time.Second))
	if !r2.Hit(t0.Add(61 * time.Second)) {
		t.Fatal("cửa sổ trượt phải thấy 3 lần trong 11 giây")
	}
	// Quá cửa sổ thì quên.
	if n := r2.Count(t0.Add(3 * time.Minute)); n != 0 {
		t.Fatalf("còn %d mốc sau khi cửa sổ trôi qua", n)
	}
}
