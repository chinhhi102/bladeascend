# Việc cần làm — GAME_SPEC v6.28 + ADMIN_SPEC v1.4 (Q34, Q35)

Q34: bot lấp THEO KHU, sàn `zone_min_chars = 10` mỗi khu (bỏ `bot_fill_target` tổng 20/60).
Q35: trần người thật là cấu hình — production 50, staging 20; không cần chờ loadtest staging đạt 50.

1. Thông số `room_max_humans`, `zone_min_chars`, `bot_auto_off_at`, `bot_mode` lưu SQLite, chỉnh ở admin, áp dụng ngay, sống qua deploy. Mặc định lấy theo APP_ENV từ `config.room_env` (production 50/10/40, staging 20/10/20).
2. Bot: mỗi 2s mỗi khu lấp đủ 10 nhân vật khi được phép; tự tắt khi người thật ≥ `bot_auto_off_at`, bật lại < ngưỡng − 5. Bot không tính vào trần.
3. Hạ trần không đuổi ai — phòng vượt trần chỉ ngừng nhận người mới.
4. Admin (React/antd khi làm 3c; trước đó thêm tạm vào trang hiện có): thẻ "Cấu hình phòng" + cảnh báo khi trần vượt mức đã qua loadtest của máy đang chạy.
5. Kiểm production :9999 = 50 người thật (đã đạt 2,02 ms trên i5-12400F); staging giữ 20. Chuyển Q34, Q35 sang QUESTIONS_DONE.
