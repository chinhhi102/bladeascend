# Việc cần làm — GAME_SPEC v6.19 (bảng xếp hạng phòng)

Đọc §11.3 (danh hiệu theo phòng) và §11.5 (thiết kế bảng).

1. **Danh hiệu theo cả phòng**: kiểm mọi chỗ tính hạng/danh hiệu — phải gộp người thật + bot của cả 5 khu, Top 1 phòng chỉ 1 người. Thêm test: 2 người ở 2 khu, người Uy Danh cao hơn là Top 1 duy nhất, người kia thấy đúng danh hiệu khi qua khu của Top 1.
2. **Gói Leaderboard 1 Hz**: mỗi dòng `{id, name, character, title_id, rank, zone_id, realm, tier, score, stealth}` + hạng thật của người nhận. Thêm vào schema Protocol v2 + golden.
3. **Giao diện** theo §11.5: thẻ Top 1 (to, viền vàng chuyển sắc, vệt sáng, vương miện vẽ, chân dung 40px, lấp lánh), Top 2/3 (thẻ nửa rộng, bạc/đồng, chân dung 28px), 4–10 dạng dòng gọn (danh hiệu dòng phụ), dòng của mình ghim cuối. Mỗi người: tên · danh hiệu · cảnh giới (viên thuốc màu §12.3) · khu (icon + tên ngắn, "???" khi Ẩn Thân) · Uy Danh.
4. Hoạt ảnh đổi hạng (trượt, ▲▼, Top 1 lật thẻ + chớp vàng). Điện thoại: thu gọn mặc định (3 chân dung + "Bạn #N"), chạm để mở.
5. Không tràn màn hình (≤ 300px PC, cắt "…"), i18n, chất lượng Thấp tắt hiệu ứng động. Chụp ảnh WebKit iPhone + PC 1366×768 gửi tôi duyệt.
