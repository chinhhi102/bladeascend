# Kết quả đo 7 tab Chrome headless — server https://kiemtien.staging.maveragroups.com

Chạy lại: `cd tools/browsertest && node run.mjs --url https://kiemtien.staging.maveragroups.com --tabs 7` — tab chạy
trên máy đo, nối qua Internet tới server thật; ping và KB/s là số qua mạng thật.
7 browser context riêng, mỗi context một người chơi; 30s đi lang thang rồi
5 tab lên Đạo Tổ đủ kiếm và cả 7 tab chụm một chỗ 60s.

> Bài này chạy **GPU thật** (`--use-gl=angle --use-angle=vulkan`), không dùng
> SwiftShader nữa. Nhưng **7 tab vẫn chia nhau MỘT card và một CPU**, nên cột FPS chỉ
> nói lên "7 người chơi trên một máy", không nói lên trải nghiệm của một người chơi thật.
>
> Server chạy trên máy riêng, không tranh CPU với trình duyệt: cột **tick server** là số thật
> của máy đó (cộng thêm bot lấp phòng).
>
> Số đáng tin nhất ở bảng này: **ping, KB/s, kích thước snapshot, số entity trong AOI**.
>
> Lượt B bóp băng thông **sau khi đã vào game**: ta đo lưu lượng lúc chơi, không đo lúc
> tải trang.

**Lượt A — mạng bình thường (Internet tới server thật)** — 441 mẫu trong giai đoạn đánh nhau

| Chỉ số | 5 tab Đạo Tổ (TB / p95 / tệ nhất) | 2 tab thường (TB / p95 / tệ nhất) |
|---|---|---|
| FPS *(TB / p5 / thấp nhất — xem ghi chú)* | 31.6 / 11.4 / 0.5 | *(chỉ đo ở tab 1)* |
| Cảnh giới (cấp) | 130.0 / 130.0 / 130.0 | 0.0 / 0.0 / 1.0 |
| Kiếm của chính tab | 1042.4 / 1050.0 / 1050.0 | 1.0 / 1.0 / 1.0 |
| Thời gian vẽ | 1.2 / 2.6 / 3.0 ms | *(chỉ đo ở tab 1)* |
| Ping | 242.1 / 367.0 / 740.0 ms | 195.5 / 202.0 / 231.0 ms |
| Băng thông xuống | 25.5 / 44.9 / 69.3 KB/s | 7.3 / 11.6 / 23.9 KB/s |
| Kích thước snapshot | 1.2 / 2.7 / 4.4 KB | 0.4 / 0.7 / 2.4 KB |
| Tick server | 2.2 / 3.6 / 6.4 ms | 1.9 / 2.7 / 3.0 ms |
| Số sprite | 3839.3 / 4197.0 / 4260.0 | *(chỉ đo ở tab 1)* |
| Draw call | 689.9 / 277.1 / 11996.2 | *(chỉ đo ở tab 1)* |
| Entity trong AOI | 935.4 / 1347.0 / 1471.0 | 282.8 / 389.0 / 396.0 |
| Tổng kiếm thấy trên màn hình | 1447.0 / 2615.0 / 3313.0 | 107.3 / 1046.0 / 1148.0 |

**Lượt B — giả lập 4G yếu (trễ 100 ms, 1.6 Mbps xuống, 750 Kbps lên)** — 432 mẫu trong giai đoạn đánh nhau

| Chỉ số | 5 tab Đạo Tổ (TB / p95 / tệ nhất) | 2 tab thường (TB / p95 / tệ nhất) |
|---|---|---|
| FPS *(TB / p5 / thấp nhất — xem ghi chú)* | 23.6 / 15.9 / 10.0 | *(chỉ đo ở tab 1)* |
| Cảnh giới (cấp) | 130.0 / 130.0 / 130.0 | 1.1 / 3.0 / 3.0 |
| Kiếm của chính tab | 1020.6 / 1050.0 / 1050.0 | 2.3 / 10.0 / 10.0 |
| Thời gian vẽ | 1.6 / 2.9 / 3.1 ms | *(chỉ đo ở tab 1)* |
| Ping | 275.6 / 399.0 / 750.0 ms | 208.6 / 225.0 / 253.0 ms |
| Băng thông xuống | 22.3 / 34.2 / 58.7 KB/s | 9.2 / 17.7 / 43.4 KB/s |
| Kích thước snapshot | 1.0 / 2.1 / 2.6 KB | 0.4 / 0.8 / 1.3 KB |
| Tick server | 3.2 / 5.4 / 6.9 ms | 2.7 / 4.6 / 6.8 ms |
| Số sprite | 4215.0 / 4438.0 / 4591.0 | *(chỉ đo ở tab 1)* |
| Draw call | 408.8 / 442.8 / 444.8 | *(chỉ đo ở tab 1)* |
| Entity trong AOI | 934.8 / 1174.0 / 1246.0 | 330.1 / 405.0 / 456.0 |
| Tổng kiếm thấy trên màn hình | 1471.2 / 2253.0 / 2857.0 | 274.0 / 1068.0 / 1104.0 |

Ảnh chụp lúc đông nhất: `docs/perf/browsertest-staging-mang-binh-thuong.png` và
`docs/perf/browsertest-staging-4g-yeu.png`.
