# Kết quả đo 10 tab Chrome headless

Chạy lại: `cd tools/browsertest && node run.mjs` — bài test **tự dựng server riêng ở
cổng 8079 với database tạm** rồi tắt khi xong, không đụng tới server đang có người chơi.
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

**Lượt A — mạng bình thường (LAN nội bộ)** — 672 mẫu trong giai đoạn đánh nhau

| Chỉ số | 5 tab Đạo Tổ (TB / p95 / tệ nhất) | 5 tab thường (TB / p95 / tệ nhất) |
|---|---|---|
| FPS *(TB / p5 / thấp nhất — xem ghi chú)* | 4.2 / 1.2 / 0.5 | *(chỉ đo ở tab 1)* |
| Cảnh giới (cấp) | 130.0 / 130.0 / 130.0 | 1.5 / 5.0 / 6.0 |
| Kiếm của chính tab | 684.2 / 1050.0 / 1050.0 | 1.1 / 7.0 / 10.0 |
| Thời gian vẽ | 16.1 / 28.0 / 30.4 ms | *(chỉ đo ở tab 1)* |
| Ping | 338.6 / 1245.0 / 3910.0 ms | 121.9 / 378.0 / 2612.0 ms |
| Băng thông xuống | 30.4 / 59.5 / 271.6 KB/s | 7.9 / 17.0 / 26.4 KB/s |
| Kích thước snapshot | 1.7 / 3.8 / 5.8 KB | 0.6 / 2.4 / 2.9 KB |
| Tick server | 8.2 / 15.1 / 59.4 ms | 9.3 / 19.1 / 87.3 ms |
| Số sprite | 3627.3 / 3947.0 / 3971.0 | *(chỉ đo ở tab 1)* |
| Draw call | 431.5 / 505.5 / 505.5 | *(chỉ đo ở tab 1)* |
| Entity trong AOI | 1101.8 / 1408.0 / 1459.0 | 318.7 / 391.0 / 406.0 |
| Tổng kiếm thấy trên màn hình | 2623.6 / 5275.0 / 5426.0 | 695.2 / 2234.0 / 5211.0 |

**Lượt B — giả lập 4G yếu (trễ 100 ms, 1.6 Mbps xuống, 750 Kbps lên)** — 658 mẫu trong giai đoạn đánh nhau

| Chỉ số | 5 tab Đạo Tổ (TB / p95 / tệ nhất) | 5 tab thường (TB / p95 / tệ nhất) |
|---|---|---|
| FPS *(TB / p5 / thấp nhất — xem ghi chú)* | 9.3 / 0.5 / 0.3 | *(chỉ đo ở tab 1)* |
| Cảnh giới (cấp) | 130.0 / 130.0 / 130.0 | 1.6 / 6.0 / 10.0 |
| Kiếm của chính tab | 870.7 / 1050.0 / 1050.0 | 1.3 / 4.0 / 10.0 |
| Thời gian vẽ | 5.0 / 8.3 / 10.5 ms | *(chỉ đo ở tab 1)* |
| Ping | 994.4 / 6861.0 / 7443.0 ms | 376.9 / 167.0 / 8970.0 ms |
| Băng thông xuống | 27.8 / 54.4 / 126.1 KB/s | 10.7 / 33.6 / 142.5 KB/s |
| Kích thước snapshot | 1.4 / 3.1 / 5.7 KB | 0.5 / 0.9 / 2.6 KB |
| Tick server | 8.2 / 16.3 / 156.3 ms | 5.2 / 10.8 / 29.5 ms |
| Số sprite | 4805.1 / 5016.0 / 5322.0 | *(chỉ đo ở tab 1)* |
| Draw call | 217.2 / 308.1 / 309.8 | *(chỉ đo ở tab 1)* |
| Entity trong AOI | 996.5 / 1316.0 / 1416.0 | 327.6 / 397.0 / 446.0 |
| Tổng kiếm thấy trên màn hình | 2462.3 / 4538.0 / 5215.0 | 436.2 / 1939.0 / 3103.0 |

Ảnh chụp lúc đông nhất: `docs/perf/browsertest-mang-binh-thuong.png` và
`docs/perf/browsertest-4g-yeu.png`.
