package main

import (
	"net/http"
	"net/http/pprof"
)

// Hồ sơ CPU/heap của CHÍNH tiến trình đang phục vụ người chơi.
//
// Vì sao không dùng cờ `-cpuprofile` có sẵn: cờ đó bắt đầu ghi lúc khởi động và chỉ đóng
// file khi tắt server — muốn xem 30 giây lúc tick đang xấu thì phải tắt server đi, tức là
// đúng lúc cần đo nhất thì không đo được. Điểm cuối này lấy được 30 giây bất kỳ, không
// đụng gì tới người đang chơi (pprof lấy mẫu 100 Hz, tốn ~1–2% CPU).
//
// Khoá: CHỈ máy cục bộ (127.0.0.1) và chỉ khi chạy với `-pprof`. Không mở ra LAN, không
// token: ai gõ được lệnh trên máy này thì đã đọc được cả thư mục dữ liệu rồi, còn mở ra LAN
// thì một máy bất kỳ trong nhà cũng xin được hồ sơ bộ nhớ của server.
func registerPprof(mux *http.ServeMux) {
	gate := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			if ip != "127.0.0.1" && ip != "::1" {
				http.NotFound(w, r) // 404 chứ không 403: ngoài máy này thì không cần biết nó tồn tại
				return
			}
			h(w, r)
		}
	}
	mux.HandleFunc("GET /debug/pprof/", gate(pprof.Index))
	mux.HandleFunc("GET /debug/pprof/profile", gate(pprof.Profile))
	mux.HandleFunc("GET /debug/pprof/heap", gate(pprof.Handler("heap").ServeHTTP))
	mux.HandleFunc("GET /debug/pprof/goroutine", gate(pprof.Handler("goroutine").ServeHTTP))
	mux.HandleFunc("GET /debug/pprof/allocs", gate(pprof.Handler("allocs").ServeHTTP))
	mux.HandleFunc("GET /debug/pprof/symbol", gate(pprof.Symbol))
	mux.HandleFunc("GET /debug/pprof/trace", gate(pprof.Trace))
}
