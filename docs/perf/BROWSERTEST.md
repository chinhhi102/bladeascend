# Kết quả đo 10 tab Chrome headless

Chạy lại: `cd tools/browsertest && node run.mjs` — bài test **tự dựng server riêng ở
cổng 8099 với database tạm** rồi tắt khi xong, không đụng tới server đang có người chơi.
10 browser context riêng, mỗi context một người chơi; 60s đi lang thang rồi
5 tab lên Đạo Tổ đủ kiếm và cả 10 tab chụm một chỗ 60s.

> Bài này chạy **GPU thật** (`--use-gl=angle --use-angle=vulkan`), không dùng
> SwiftShader nữa. Nhưng **10 tab vẫn chia nhau MỘT card và một CPU**, nên cột FPS chỉ
> nói lên "10 người chơi trên một máy", không nói lên trải nghiệm của một người chơi thật.
>
> Server chạy riêng trên lõi 0-2, trình duyệt trên lõi 3-11.
> Con số **tick server chuẩn** lấy ở `cd server && go run ./cmd/loadtest`, nơi không có
> trình duyệt tranh CPU.
>
> Số đáng tin nhất ở bảng này: **ping, KB/s, kích thước snapshot, số entity trong AOI**.
>
> Lượt B bóp băng thông **sau khi đã vào game**: ta đo lưu lượng lúc chơi, không đo lúc
> tải trang.

**Lượt A — mạng bình thường (LAN nội bộ)** — 615 mẫu trong giai đoạn đánh nhau

| Chỉ số | 5 tab Đạo Tổ (TB / p95 / tệ nhất) | 5 tab thường (TB / p95 / tệ nhất) |
|---|---|---|
| FPS *(TB / p5 / thấp nhất — xem ghi chú)* | 11.7 / 5.8 / 1.4 | *(chỉ đo ở tab 1)* |
| Cảnh giới (cấp) | 130.0 / 130.0 / 130.0 | 4.3 / 16.0 / 17.0 |
| Kiếm của chính tab | 1026.7 / 1050.0 / 1050.0 | 2.6 / 10.0 / 10.0 |
| Thời gian vẽ | 0.6 / 0.9 / 1.0 ms | *(chỉ đo ở tab 1)* |
| Ping | 88.2 / 307.0 / 980.0 ms | 25.7 / 109.0 / 323.0 ms |
| Băng thông xuống | 6.8 / 14.3 / 23.4 KB/s | 3.6 / 5.5 / 9.2 KB/s |
| Kích thước snapshot | 0.3 / 0.8 / 2.0 KB | 0.1 / 0.2 / 1.1 KB |
| Tick server | 1.3 / 3.0 / 6.8 ms | 1.1 / 2.4 / 6.6 ms |
| Số sprite | 2723.7 / 2752.0 / 2759.0 | *(chỉ đo ở tab 1)* |
| Draw call | 340.6 / 355.9 / 361.2 | *(chỉ đo ở tab 1)* |
| Entity trong AOI | 497.2 / 908.0 / 1248.0 | 235.9 / 327.0 / 417.0 |
| Tổng kiếm thấy trên màn hình | 1344.5 / 2114.0 / 2212.0 | 45.1 / 86.0 / 1095.0 |

Ảnh chụp lúc đông nhất: `docs/perf/browsertest-mang-binh-thuong.png` và
`docs/perf/browsertest-4g-yeu.png`.
