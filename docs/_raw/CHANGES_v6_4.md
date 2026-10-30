# Việc cần làm — spec v6.4

Anh đã dán GAME_SPEC v6.4 vào docs/_raw/. Hai phần độc lập, làm phần A trước (người chơi đang chờ tải rất lâu).

## A. Dung lượng tải & màn hình tải (§13.7)
1. `npm run assets:size`: tổng dung lượng theo thư mục, top 30 file nặng nhất, số file. Ghi PERF_LOG mốc "trước".
2. `tools/asset-pipeline` chạy trước build: art gốc ở `assets/src/`, xuất `assets/dist/`. Thu nhỏ về kích thước hiển thị (frame nhân vật 256 px), trim + giữ pivot, ghép atlas (mỗi nhân vật 1–2 atlas 2048), WebP chất lượng ~80, âm thanh OGG/MP3 nén, tên file kèm hash, `manifest.json` có dung lượng byte + nhóm tải. Nếu cần thư viện xử lý ảnh (vd `sharp`) → đã duyệt.
3. Server phục vụ `assets/dist` với `Cache-Control: immutable` cho file có hash; gzip/brotli cho JS/JSON.
4. Loader tải theo nhóm `boot` / `zone_*` / `chars` / `fx_high` / `audio` (§13.7 bảng). Thiếu asset → placeholder rồi mờ dần sang ảnh thật. Qua cổng sang khu chưa tải: chờ tối đa 2s.
5. Màn hình tải HTML/CSS thuần hiện ngay: thanh tiến độ **theo byte** (%, MB), dòng trạng thái theo bước, nhân vật chạy trên thanh, mẹo/câu bựa; lỗi mạng → "Thử lại" tải tiếp. Trong trận: icon % tải nền.
6. `npm run assets:budget` vào CI: boot ≤ 8 MB, tổng ≤ 80 MB, vượt → fail.
7. Đo lại: dung lượng boot/tổng, thời gian vào trận lần đầu với throttling "Fast 4G" và lần 2 (cache). Ghi PERF_LOG "trước/sau".

## B. Tiên khí ×20 (§3.4, §4.1c)
8. `tien_khi_ceilings`: [2000, 6000, 12000, 20000, 30000]. Đáy: 1 / 2 001 / 6 001 / 12 001 / 20 001. Đạo Tổ đổi Uy Danh khi đầy 30 000.
9. Thiêu Đốt ở Tiên giới giữ lượng tuyệt đối cũ (8/16/24/32/40 mỗi giây) → `burn_tien_frac_per_sec: 0.004`.
10. Viên Tiên khí rơi khi bị chém: mệnh giá 500/100/10/2/1, tối đa 40 viên/lần. Thêm sprite viên 100 và 500 (to và sáng hơn) vào manifest + artgen.
11. Sửa test theo checklist §17 (Thái Ất 9 000 bị chém → 2 001, +3 499, rơi 2 099; Đạo Tổ 30 000).
