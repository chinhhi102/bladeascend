# Kết quả đo 10 tab Chrome headless

Chạy lại: `cd tools/browsertest && node run.mjs` — bài test **tự dựng server riêng ở
cổng 8099 với database tạm** rồi tắt khi xong, không đụng tới server đang có người chơi.
3 browser context riêng, mỗi context một người chơi; 5s đi lang thang rồi
2 tab lên Đạo Tổ đủ kiếm và cả 3 tab chụm một chỗ 10s.

> **FPS trong bảng này không dùng để kết luận được.** Chrome headless vẽ WebGL bằng CPU
> (SwiftShader, không đụng tới card đồ họa) và 10 tab tranh nhau CPU trên cùng một máy.
> Số đáng tin: **ping, KB/s, kích thước snapshot, tick server, số sprite, draw call**.
>
> Lượt B bóp băng thông **sau khi đã vào game**: ta đo lưu lượng lúc chơi, không đo lúc
> tải trang.

**Lượt A — mạng bình thường (LAN nội bộ)** — 34 mẫu trong giai đoạn đánh nhau

| Chỉ số | 5 tab Đạo Tổ (TB / p95 / tệ nhất) | 5 tab thường (TB / p95 / tệ nhất) |
|---|---|---|
| FPS *(TB / p5 / thấp nhất — không tin được, xem ghi chú)* | 10.5 / 10.0 / 10.0 | 10.3 / 10.0 / 10.0 |
| Cảnh giới (cấp) | 118.8 / 130.0 / 130.0 | 1.0 / 1.0 / 1.0 |
| Kiếm của chính tab | 958.8 / 1050.0 / 1050.0 | 1.0 / 1.0 / 1.0 |
| Thời gian vẽ | 2.1 / 3.2 / 3.6 ms | 0.6 / 0.6 / 0.6 ms |
| Ping | 328.8 / 408.0 / 408.0 ms | 461.3 / 652.0 / 652.0 ms |
| Băng thông xuống | 37.8 / 83.6 / 133.2 KB/s | 11.8 / 25.3 / 25.3 KB/s |
| Kích thước snapshot | 2.9 / 4.0 / 4.1 KB | 1.0 / 1.1 / 1.1 KB |
| Tick server | 75.5 / 259.0 / 360.4 ms | 65.1 / 259.0 / 259.0 ms |
| Số sprite | 2273.9 / 2521.0 / 2628.0 | 1742.0 / 1770.0 / 1770.0 |
| Draw call | 78.7 / 111.8 / 111.8 | 51.4 / 53.0 / 53.0 |
| Entity trong AOI | 2065.1 / 2749.0 / 2758.0 | 366.7 / 376.0 / 376.0 |
| Tổng kiếm thấy trên màn hình | 1940.1 / 2131.0 / 2135.0 | 12.8 / 19.0 / 19.0 |

**Lượt B — giả lập 4G yếu (trễ 100 ms, 1.6 Mbps xuống, 750 Kbps lên)** — 33 mẫu trong giai đoạn đánh nhau

| Chỉ số | 5 tab Đạo Tổ (TB / p95 / tệ nhất) | 5 tab thường (TB / p95 / tệ nhất) |
|---|---|---|
| FPS *(TB / p5 / thấp nhất — không tin được, xem ghi chú)* | 11.1 / 10.0 / 10.0 | 10.4 / 10.0 / 10.0 |
| Cảnh giới (cấp) | 59.1 / 130.0 / 130.0 | 0.7 / 1.0 / 1.0 |
| Kiếm của chính tab | 477.8 / 1050.0 / 1050.0 | 1.0 / 1.0 / 1.0 |
| Thời gian vẽ | 1.3 / 2.6 / 2.9 ms | 0.4 / 0.5 / 0.5 ms |
| Ping | 327.3 / 495.0 / 1050.0 ms | 458.9 / 540.0 / 540.0 ms |
| Băng thông xuống | 35.1 / 88.4 / 91.9 KB/s | 6.4 / 9.8 / 9.8 KB/s |
| Kích thước snapshot | 2.2 / 3.8 / 3.9 KB | 0.3 / 0.5 / 0.5 KB |
| Tick server | 2.0 / 2.8 / 3.6 ms | 2.5 / 3.4 / 3.4 ms |
| Số sprite | 2048.6 / 2450.0 / 2467.0 | 1700.5 / 1749.0 / 1749.0 |
| Draw call | 77.6 / 142.2 / 142.2 | 53.4 / 86.1 / 86.1 |
| Entity trong AOI | 1479.9 / 2923.0 / 2924.0 | 357.5 / 442.0 / 442.0 |
| Tổng kiếm thấy trên màn hình | 508.5 / 1110.0 / 1118.0 | 1.1 / 2.0 / 2.0 |

Ảnh chụp lúc đông nhất: `docs/perf/browsertest-mang-binh-thuong.png` và
`docs/perf/browsertest-4g-yeu.png`.
