# Kết quả đo 10 tab Chrome headless

Chạy lại: `cd tools/browsertest && node run.mjs` — bài test **tự dựng server riêng ở
cổng 8093 với database tạm** rồi tắt khi xong, không đụng tới server đang có người chơi.
5 browser context riêng, mỗi context một người chơi; 5s đi lang thang rồi
5 tab lên Đạo Tổ đủ kiếm và cả 5 tab chụm một chỗ 10s.

> **FPS trong bảng này không dùng để kết luận được.** Chrome headless vẽ WebGL bằng CPU
> (SwiftShader, không đụng tới card đồ họa) và 10 tab tranh nhau CPU trên cùng một máy.
> Số đáng tin: **ping, KB/s, kích thước snapshot, tick server, số sprite, draw call**.
>
> Lượt B bóp băng thông **sau khi đã vào game**: ta đo lưu lượng lúc chơi, không đo lúc
> tải trang.

**Lượt A — mạng bình thường (LAN nội bộ)** — 58 mẫu trong giai đoạn đánh nhau

| Chỉ số | 5 tab Đạo Tổ (TB / p95 / tệ nhất) | 5 tab thường (TB / p95 / tệ nhất) |
|---|---|---|
| FPS *(TB / p5 / thấp nhất — xem ghi chú)* | 1.9 / 1.4 / 1.4 | *(chỉ đo ở tab 1)* |
| Cảnh giới (cấp) | 130.0 / 130.0 / 130.0 | — |
| Kiếm của chính tab | 1008.9 / 1050.0 / 1050.0 | — |
| Thời gian vẽ | 4.4 / 5.9 / 5.9 ms | *(chỉ đo ở tab 1)* |
| Ping | 609.6 / 1260.0 / 1637.0 ms | — |
| Băng thông xuống | 10.7 / 30.0 / 47.3 KB/s | — |
| Kích thước snapshot | 1.9 / 4.7 / 6.9 KB | — |
| Tick server | 139.1 / 368.8 / 405.6 ms | — |
| Số sprite | 4888.5 / 5115.0 / 5115.0 | *(chỉ đo ở tab 1)* |
| Draw call | 54.7 / 54.7 / 54.7 | *(chỉ đo ở tab 1)* |
| Entity trong AOI | 937.8 / 1227.0 / 1227.0 | — |
| Tổng kiếm thấy trên màn hình | 3605.7 / 5027.0 / 5027.0 | — |

Ảnh chụp lúc đông nhất: `docs/perf/browsertest-mang-binh-thuong.png` và
`docs/perf/browsertest-4g-yeu.png`.
