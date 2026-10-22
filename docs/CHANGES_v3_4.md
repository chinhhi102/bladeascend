# Việc cần làm — spec v3.4 (minimap + danh hiệu Top 3)

Đọc docs/GAME_SPEC.md §11.3 và §12.1. Cập nhật config/game.json: thêm `titles`, `title_hold_secs`, `minimap`.

**Danh hiệu**
1. Server: từ bảng phòng, tính Top 1–3 có chống nhấp nháy (giữ hạng 3s mới đổi). Gửi danh hiệu theo entity trong snapshot (hoặc trong gói Leaderboard).
2. Client: danh hiệu phía trên tên (vàng/bạc/đồng), vương miện cho Top 1, danh hiệu cạnh tên trong bảng phòng. Ẩn khi người đó Ẩn Thân hoặc trong bụi.
3. Kill feed + SFX khi lên ngôi / bị hạ bệ Top 1.

**Minimap**
4. Vẽ sẵn 1 texture minimap từ bản đồ Tiled lúc vào phòng (tường cố định, đá tảng, bụi, bùn, gió, cổng).
5. Mỗi frame vẽ thêm: mũi tên của mình, khung camera. Mỗi giây cập nhật từ gói Leaderboard: vị trí Top 1 (nếu không Ẩn Thân/trong bụi), Rương Báu.
6. Góc trên trái, 180px PC / 120px điện thoại; phím M hoặc chạm để phóng to/thu nhỏ.
7. Không gửi vị trí người khác cho minimap (chống hack).
8. Unit test: chống nhấp nháy danh hiệu; Top 1 trong bụi không có trong dữ liệu minimap.

**Tốc độ quay kiếm**
9. `rev_secs`: 0.3 → **1.0** (1 vòng/giây). Buff Tốc Kiếm `spin.mult`: 1.3 → **1.6** (0.625 giây/vòng).
10. Giữ va chạm theo góc quét (§4.1b). Giảm độ đậm vệt mờ spin blur cho tốc độ gốc (chỉ vệt mảnh), đậm hơn khi có buff.
11. Sửa test liên quan tới tốc độ quay (góc sau 1 giây = 360°).
