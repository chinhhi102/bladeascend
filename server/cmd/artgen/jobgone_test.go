package main

import (
	"errors"
	"testing"
)

// `jobGone` quyết định artgen có XOÁ job id đã lưu hay không, và quyết định sai theo chiều
// nào cũng tốn thật:
//   - bỏ sót 404 → ảnh đó kẹt vĩnh viễn, mọi đợt chạy sau đều hỏng y hệt (đã xảy ra với
//     `plate_shine_streak`, hỏng hai đợt liền với cùng một job id);
//   - bắt nhầm lỗi hết giờ chờ → vứt một job vẫn đang vẽ và tốn thêm một lượt quota.
func TestJobGone(t *testing.T) {
	for _, c := range []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"404 đúng câu minigen trả về",
			errors.New(`GET /jobs/af8d0df0: HTTP 404 {"detail":"Job not found or expired"}`), true},
		{"404 không kèm câu chữ", errors.New("GET /jobs/x: HTTP 404"), true},
		{"chỉ có câu chữ, mã khác", errors.New("HTTP 410 job not found"), true},
		{"hết giờ chờ — job vẫn đang vẽ, PHẢI giữ id",
			errors.New("job af8d0df0 timed out"), false},
		{"minigen lỗi tạm 500", errors.New("GET /jobs/x: HTTP 500 server error"), false},
		{"hết quota", errors.New("job failed: quota exceeded"), false},
		{"mất mạng", errors.New("dial tcp: connection refused"), false},
	} {
		if got := jobGone(c.err); got != c.want {
			t.Errorf("%s: jobGone = %v, muốn %v", c.name, got, c.want)
		}
	}
}
