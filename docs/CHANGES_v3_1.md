# Việc cần làm — spec v3.1 (vòng kiếm kiểu quảng cáo)

Đọc docs/GAME_SPEC.md §4.1a mới. Cập nhật config/game.json: thêm `hilt_step`, `tilt_deg`, `ring_tints`.

1. Server (`internal/game`): tính vị trí kiếm theo hình học mới — chuôi vòng k tại `hilt_r(k)`, chiều dài `len(k)`, so le `stagger(k) = k * 18°/R`, nghiêng `tilt = 35° * dir`. Hitbox capsule theo hướng đã nghiêng. `reach` tính hình học từ mũi kiếm xa nhất (dùng cho camera và AI).
2. Client: vẽ đúng hình học đó (dùng chung công thức, lấy số từ /api/config). Mỗi vòng nhuộm màu `ring_tints[k % len]`, vòng 1 giữ màu gốc. Vẽ vòng ngoài trước, vòng 1 sau cùng. Spin blur cùng màu với vòng.
3. Đổi chiều quay → tilt lật theo, tween 0.1s.
4. Thêm phím debug vẽ capsule hitbox của server chồng lên kiếm để kiểm tra khớp.
5. Unit test: góc so le đúng khi R = 1, 2, 3; vòng chưa đủ 20 kiếm vẫn rải đều khe; reach tăng khi thêm vòng; tilt đổi dấu khi đổi chiều.
6. Skin kiếm hiện có màu đậm thì nhuộm sẽ xỉn — tạm thời dùng placeholder thân kiếm màu trắng/xám cho tới khi sinh lại skin theo ART_PROMPTS §7.
