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

**Lượt A — mạng bình thường (LAN nội bộ)** — 610 mẫu trong giai đoạn đánh nhau

| Chỉ số | 5 tab Đạo Tổ (TB / p95 / tệ nhất) | 5 tab thường (TB / p95 / tệ nhất) |
|---|---|---|
| FPS *(TB / p5 / thấp nhất — không tin được, xem ghi chú)* | 10.0 / 10.0 / 10.0 | 10.0 / 10.0 / 10.0 |
| Cảnh giới (cấp) | 0.0 / 0.0 / 0.0 | 0.4 / 2.0 / 2.0 |
| Kiếm của chính tab | 0.2 / 1.0 / 1.0 | 0.4 / 1.0 / 1.0 |
| Thời gian vẽ | 0.8 / 1.1 / 1.1 ms | 1.7 / 3.1 / 3.1 ms |
| Ping | 414.7 / 479.0 / 1899.0 ms | 403.4 / 448.0 / 1375.0 ms |
| Băng thông xuống | 0.0 / 0.0 / 0.0 KB/s | 0.0 / 0.0 / 0.0 KB/s |
| Kích thước snapshot | 0.7 / 0.9 / 0.9 KB | 0.6 / 0.9 / 0.9 KB |
| Tick server | 2.0 / 2.9 / 2.9 ms | 1.3 / 1.7 / 1.7 ms |
| Số sprite | 4289.5 / 4325.0 / 4325.0 | 4261.6 / 4369.0 / 4369.0 |
| Draw call | 70.2 / 78.5 / 78.5 | 41.3 / 72.7 / 72.7 |
| Entity trong AOI | 254.9 / 303.0 / 303.0 | 252.6 / 315.0 / 315.0 |
| Tổng kiếm thấy trên màn hình | 13.6 / 19.0 / 19.0 | 24.2 / 35.0 / 35.0 |

**Lượt B — giả lập 4G yếu (trễ 100 ms, 1.6 Mbps xuống, 750 Kbps lên)** — 619 mẫu trong giai đoạn đánh nhau

| Chỉ số | 5 tab Đạo Tổ (TB / p95 / tệ nhất) | 5 tab thường (TB / p95 / tệ nhất) |
|---|---|---|
| FPS *(TB / p5 / thấp nhất — không tin được, xem ghi chú)* | 10.0 / 10.0 / 10.0 | 8.0 / 0.0 / 0.0 |
| Cảnh giới (cấp) | 0.0 / 0.0 / 0.0 | 0.8 / 4.0 / 4.0 |
| Kiếm của chính tab | 0.6 / 1.0 / 1.0 | 0.6 / 1.0 / 1.0 |
| Thời gian vẽ | 1.0 / 2.0 / 2.0 ms | 3.2 / 6.4 / 6.4 ms |
| Ping | 711.8 / 1234.0 / 2188.0 ms | 684.1 / 1180.0 / 1826.0 ms |
| Băng thông xuống | 0.0 / 0.0 / 0.0 KB/s | 0.0 / 0.0 / 0.0 KB/s |
| Kích thước snapshot | 0.5 / 0.7 / 0.7 KB | 0.5 / 0.7 / 0.7 KB |
| Tick server | 1.8 / 2.2 / 2.2 ms | 2.4 / 4.1 / 4.1 ms |
| Số sprite | 4279.2 / 4432.0 / 4432.0 | 4279.4 / 4340.0 / 4340.0 |
| Draw call | 94.9 / 203.2 / 203.2 | 44.7 / 100.4 / 100.4 |
| Entity trong AOI | 188.1 / 273.0 / 273.0 | 225.6 / 310.0 / 310.0 |
| Tổng kiếm thấy trên màn hình | 20.3 / 52.0 / 52.0 | 49.4 / 111.0 / 111.0 |

Ảnh chụp lúc đông nhất: `docs/perf/browsertest-mang-binh-thuong.png` và
`docs/perf/browsertest-4g-yeu.png`.
