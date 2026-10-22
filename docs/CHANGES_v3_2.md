# Việc cần làm — spec v3.2 (sửa lỗi vòng kiếm rộng ra theo số kiếm)

Lỗi hiện tại: thêm kiếm thì vòng rộng ra, và có khoảng trống lớn giữa thân và kiếm. Sai so với docs/GAME_SPEC.md §4.1a ("Quy tắc cố định bán kính").

1. Tìm và xóa mọi chỗ bán kính vòng phụ thuộc số kiếm (công thức cũ `max(ring_min_r, n * spacing / 2π)` từ spec v1, các key config `ring_min_r`, `spacing` nếu còn) — cả server lẫn client.
2. Bán kính chuôi và chiều dài kiếm chỉ dùng `hilt_r(k)` và `len(k)` theo §4.1a (phụ thuộc k, scale, buff Trường Kiếm). Vòng 1 chuôi sát người: `inner_r = (BODY_R + 6) * scale`.
3. Thêm kiếm vào vòng chưa đầy → chỉ phân bổ lại góc (tween 0.2s), bán kính giữ nguyên.
4. Unit test: ở cùng scale, `hilt_r` và `len` của vòng 1 bằng nhau với n = 1, 10, 20; n = 21 → có vòng 2 với `hilt_r(1) > hilt_r(0)`; n = 40 → bán kính 2 vòng giống hệt n = 21.
5. Kiểm tra lại bằng phím debug vẽ hitbox.
