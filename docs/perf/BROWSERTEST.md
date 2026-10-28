# Kết quả đo 10 tab Chrome headless

Chạy lại: `cd tools/browsertest && node run.mjs` — bài test **tự dựng server riêng ở
cổng 8089 với database tạm** rồi tắt khi xong, không đụng tới server đang có người chơi.
10 browser context riêng, mỗi context một người chơi; 30s đi lang thang rồi
5 tab lên Đạo Tổ đủ kiếm và cả 10 tab chụm một chỗ 60s.

> **FPS trong bảng này không dùng để kết luận được.** Chrome headless vẽ WebGL bằng CPU
> (SwiftShader, không đụng tới card đồ họa) và 10 tab tranh nhau CPU trên cùng một máy.
> Số đáng tin: **ping, KB/s, kích thước snapshot, tick server, số sprite, draw call**.
>
> Lượt B bóp băng thông **sau khi đã vào game**: ta đo lưu lượng lúc chơi, không đo lúc
> tải trang.

**Lượt A — mạng bình thường (LAN nội bộ)** — 600 mẫu trong giai đoạn đánh nhau

| Chỉ số | 5 tab Đạo Tổ (TB / p95 / tệ nhất) | 5 tab thường (TB / p95 / tệ nhất) |
|---|---|---|
| FPS *(TB / p5 / thấp nhất — xem ghi chú)* | 15.1 / 9.1 / 4.9 | *(chỉ đo ở tab 1)* |
| Cảnh giới (cấp) | 130.0 / 130.0 / 130.0 | 1.1 / 5.0 / 5.0 |
| Kiếm của chính tab | 1036.5 / 1050.0 / 1050.0 | 2.1 / 10.0 / 10.0 |
| Thời gian vẽ | 2.1 / 4.2 / 5.2 ms | *(chỉ đo ở tab 1)* |
| Ping | 68.4 / 181.0 / 457.0 ms | 15.9 / 53.0 / 154.0 ms |
| Băng thông xuống | 7.8 / 33.4 / 43.1 KB/s | 2.5 / 9.1 / 20.0 KB/s |
| Kích thước snapshot | 1.6 / 2.7 / 4.5 KB | 0.6 / 2.3 / 2.9 KB |
| Tick server | 290.7 / 660.4 / 661.2 ms | 291.8 / 660.4 / 661.2 ms |
| Số sprite | 4918.4 / 5140.0 / 5328.0 | *(chỉ đo ở tab 1)* |
| Draw call | 255.8 / 280.4 / 283.3 | *(chỉ đo ở tab 1)* |
| Entity trong AOI | 929.9 / 1152.0 / 1190.0 | 297.9 / 384.0 / 397.0 |
| Tổng kiếm thấy trên màn hình | 3045.2 / 5255.0 / 5263.0 | 793.3 / 2107.0 / 3146.0 |

**Lượt B — giả lập 4G yếu (trễ 100 ms, 1.6 Mbps xuống, 750 Kbps lên)** — 610 mẫu trong giai đoạn đánh nhau

| Chỉ số | 5 tab Đạo Tổ (TB / p95 / tệ nhất) | 5 tab thường (TB / p95 / tệ nhất) |
|---|---|---|
| FPS *(TB / p5 / thấp nhất — xem ghi chú)* | 15.5 / 9.8 / 1.4 | *(chỉ đo ở tab 1)* |
| Cảnh giới (cấp) | 130.0 / 130.0 / 130.0 | 0.6 / 1.0 / 7.0 |
| Kiếm của chính tab | 997.6 / 1050.0 / 1050.0 | 1.1 / 4.0 / 4.0 |
| Thời gian vẽ | 1.8 / 4.8 / 8.9 ms | *(chỉ đo ở tab 1)* |
| Ping | 84.0 / 181.0 / 411.0 ms | 28.4 / 75.0 / 156.0 ms |
| Băng thông xuống | 7.2 / 41.0 / 98.4 KB/s | 2.0 / 11.8 / 34.0 KB/s |
| Kích thước snapshot | 1.9 / 3.2 / 7.1 KB | 0.6 / 1.2 / 2.9 KB |
| Tick server | 511.2 / 952.6 / 978.2 ms | 517.8 / 952.6 / 978.2 ms |
| Số sprite | 5235.0 / 5486.0 / 5634.0 | *(chỉ đo ở tab 1)* |
| Draw call | 560.1 / 524.5 / 1624.9 | *(chỉ đo ở tab 1)* |
| Entity trong AOI | 1134.4 / 1345.0 / 1382.0 | 316.8 / 402.0 / 427.0 |
| Tổng kiếm thấy trên màn hình | 2974.7 / 5304.0 / 5331.0 | 873.3 / 4152.0 / 4207.0 |

Ảnh chụp lúc đông nhất: `docs/perf/browsertest-mang-binh-thuong.png` và
`docs/perf/browsertest-4g-yeu.png`.
