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

**Lượt B — giả lập 4G yếu (trễ 100 ms, 1.6 Mbps xuống, 750 Kbps lên)** — 670 mẫu trong giai đoạn đánh nhau

| Chỉ số | 5 tab Đạo Tổ (TB / p95 / tệ nhất) | 5 tab thường (TB / p95 / tệ nhất) |
|---|---|---|
| FPS *(TB / p5 / thấp nhất — xem ghi chú)* | 20.6 / 8.8 / 6.8 | *(chỉ đo ở tab 1)* |
| Cảnh giới (cấp) | 130.0 / 130.0 / 130.0 | 2.4 / 10.0 / 10.0 |
| Kiếm của chính tab | 1049.8 / 1050.0 / 1050.0 | 3.0 / 10.0 / 10.0 |
| Thời gian vẽ | 0.6 / 0.8 / 0.9 ms | *(chỉ đo ở tab 1)* |
| Ping | 84.3 / 264.0 / 371.0 ms | 65.2 / 294.0 / 731.0 ms |
| Băng thông xuống | 7.1 / 11.9 / 73.6 KB/s | 3.5 / 6.1 / 10.1 KB/s |
| Kích thước snapshot | 0.4 / 0.8 / 3.8 KB | 0.2 / 0.3 / 1.1 KB |
| Tick server | 1.4 / 3.1 / 7.6 ms | 1.2 / 2.7 / 6.5 ms |
| Số sprite | 1862.8 / 1918.0 / 1929.0 | *(chỉ đo ở tab 1)* |
| Draw call | 222.1 / 227.0 / 229.3 | *(chỉ đo ở tab 1)* |
| Entity trong AOI | 580.2 / 917.0 / 1062.0 | 283.9 / 367.0 / 384.0 |
| Tổng kiếm thấy trên màn hình | 1090.9 / 1133.0 / 1377.0 | 402.2 / 1142.0 / 1169.0 |

Ảnh chụp lúc đông nhất: `docs/perf/browsertest-mang-binh-thuong.png` và
`docs/perf/browsertest-4g-yeu.png`.
