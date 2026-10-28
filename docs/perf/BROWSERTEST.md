# Kết quả đo 10 tab Chrome headless

Chạy lại: `cd tools/browsertest && node run.mjs` — bài test **tự dựng server riêng ở
cổng 8099 với database tạm** rồi tắt khi xong, không đụng tới server đang có người chơi.
10 browser context riêng, mỗi context một người chơi; 30s đi lang thang rồi
5 tab lên Đạo Tổ đủ kiếm và cả 10 tab chụm một chỗ 60s.

> **FPS trong bảng này không dùng để kết luận được.** Chrome headless vẽ WebGL bằng CPU
> (SwiftShader, không đụng tới card đồ họa) và 10 tab tranh nhau CPU trên cùng một máy.
> Số đáng tin: **ping, KB/s, kích thước snapshot, tick server, số sprite, draw call**.
>
> Lượt B bóp băng thông **sau khi đã vào game**: ta đo lưu lượng lúc chơi, không đo lúc
> tải trang.

**Lượt A — mạng bình thường (LAN nội bộ)** — 582 mẫu trong giai đoạn đánh nhau

| Chỉ số | 5 tab Đạo Tổ (TB / p95 / tệ nhất) | 5 tab thường (TB / p95 / tệ nhất) |
|---|---|---|
| FPS *(TB / p5 / thấp nhất — không tin được, xem ghi chú)* | 10.0 / 10.0 / 10.0 | 8.0 / 0.0 / 0.0 |
| Cảnh giới (cấp) | 0.0 / 0.0 / 0.0 | 0.0 / 0.0 / 0.0 |
| Kiếm của chính tab | 0.6 / 1.0 / 1.0 | 0.8 / 1.0 / 1.0 |
| Thời gian vẽ | 1.9 / 5.3 / 5.3 ms | 2.1 / 3.4 / 3.4 ms |
| Ping | 627.1 / 1053.0 / 2124.0 ms | 607.6 / 1082.0 / 1154.0 ms |
| Băng thông xuống | 0.0 / 0.0 / 0.0 KB/s | 0.0 / 0.0 / 0.0 KB/s |
| Kích thước snapshot | 0.7 / 0.9 / 0.9 KB | 0.8 / 1.0 / 1.0 KB |
| Tick server | 2.4 / 3.1 / 3.1 ms | 1.5 / 1.8 / 1.8 ms |
| Số sprite | 4009.3 / 4099.0 / 4099.0 | 4080.5 / 4125.0 / 4125.0 |
| Draw call | 68.7 / 114.4 / 114.4 | 93.1 / 188.9 / 188.9 |
| Entity trong AOI | 330.0 / 404.0 / 404.0 | 325.4 / 410.0 / 410.0 |
| Tổng kiếm thấy trên màn hình | 26.2 / 61.0 / 61.0 | 69.7 / 131.0 / 131.0 |

**Lượt B — giả lập 4G yếu (trễ 100 ms, 1.6 Mbps xuống, 750 Kbps lên)** — 604 mẫu trong giai đoạn đánh nhau

| Chỉ số | 5 tab Đạo Tổ (TB / p95 / tệ nhất) | 5 tab thường (TB / p95 / tệ nhất) |
|---|---|---|
| FPS *(TB / p5 / thấp nhất — không tin được, xem ghi chú)* | 10.0 / 10.0 / 10.0 | 10.0 / 10.0 / 10.0 |
| Cảnh giới (cấp) | 0.0 / 0.0 / 0.0 | 0.7 / 2.0 / 10.0 |
| Kiếm của chính tab | 0.6 / 1.0 / 1.0 | 1.0 / 1.0 / 7.0 |
| Thời gian vẽ | 0.7 / 0.7 / 0.7 ms | 1.3 / 2.7 / 2.7 ms |
| Ping | 445.9 / 480.0 / 497.0 ms | 475.4 / 495.0 / 2349.0 ms |
| Băng thông xuống | 0.0 / 0.0 / 0.0 KB/s | 1.2 / 0.0 / 41.3 KB/s |
| Kích thước snapshot | 0.8 / 1.3 / 1.3 KB | 0.8 / 1.0 / 1.1 KB |
| Tick server | 1.4 / 1.5 / 1.5 ms | 1.8 / 3.4 / 3.4 ms |
| Số sprite | 3983.0 / 4037.0 / 4037.0 | 4072.5 / 4167.0 / 4207.0 |
| Draw call | 65.6 / 70.1 / 70.1 | 74.1 / 113.0 / 113.0 |
| Entity trong AOI | 325.0 / 393.0 / 393.0 | 298.8 / 374.0 / 493.0 |
| Tổng kiếm thấy trên màn hình | 21.2 / 29.0 / 29.0 | 68.0 / 131.0 / 170.0 |

Ảnh chụp lúc đông nhất: `docs/perf/browsertest-mang-binh-thuong.png` và
`docs/perf/browsertest-4g-yeu.png`.
