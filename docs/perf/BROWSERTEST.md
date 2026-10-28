# Kết quả đo 10 tab Chrome headless

Chạy lại: `cd tools/browsertest && node run.mjs` — bài test **tự dựng server riêng ở
cổng 8081 với database tạm** rồi tắt khi xong, không đụng tới server đang có người chơi.
10 browser context riêng, mỗi context một người chơi; 30s đi lang thang rồi
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

**Lượt A — mạng bình thường (LAN nội bộ)** — 599 mẫu trong giai đoạn đánh nhau

| Chỉ số | 5 tab Đạo Tổ (TB / p95 / tệ nhất) | 5 tab thường (TB / p95 / tệ nhất) |
|---|---|---|
| FPS *(TB / p5 / thấp nhất — xem ghi chú)* | 7.6 / 2.6 / 1.4 | *(chỉ đo ở tab 1)* |
| Cảnh giới (cấp) | 130.0 / 130.0 / 130.0 | 2.0 / 8.0 / 8.0 |
| Kiếm của chính tab | 975.4 / 1050.0 / 1050.0 | 1.5 / 7.0 / 10.0 |
| Thời gian vẽ | 5.3 / 11.1 / 12.7 ms | *(chỉ đo ở tab 1)* |
| Ping | 159.6 / 458.0 / 934.0 ms | 39.0 / 141.0 / 430.0 ms |
| Băng thông xuống | 9.6 / 39.8 / 91.0 KB/s | 3.7 / 14.2 / 61.1 KB/s |
| Kích thước snapshot | 1.4 / 3.1 / 5.8 KB | 0.6 / 2.5 / 3.5 KB |
| Tick server | 511.2 / 1371.9 / 1665.7 ms | 518.8 / 1371.9 / 1665.7 ms |
| Số sprite | 4929.5 / 5069.0 / 5213.0 | *(chỉ đo ở tab 1)* |
| Draw call | 281.4 / 317.4 / 317.4 | *(chỉ đo ở tab 1)* |
| Entity trong AOI | 766.5 / 1066.0 / 1150.0 | 233.9 / 361.0 / 372.0 |
| Tổng kiếm thấy trên màn hình | 2621.3 / 4225.0 / 4242.0 | 487.1 / 2122.0 / 4223.0 |

Ảnh chụp lúc đông nhất: `docs/perf/browsertest-mang-binh-thuong.png` và
`docs/perf/browsertest-4g-yeu.png`.
