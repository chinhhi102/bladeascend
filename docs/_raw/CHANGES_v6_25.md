# Việc cần làm — GAME_SPEC v6.25 (tốc độ cộng + kỹ năng hồi chiêu)

Làm ngay sau GIAI ĐOẠN 1 của lượt lớn (cùng nhóm "sửa lỗi ảnh hưởng chơi").

## A. Tốc độ cộng trên base (§4.1)
1. `move_speed = 195 × clamp(1 + Σ thưởng + Σ phạt, 0.4, trần)`; thưởng/phạt + trần trong `config.speed`. Bỏ mọi chỗ nhân dồn (buff × Thiêu Đốt × vùng × kỹ năng). Cùng nguồn không cộng 2 lần.
2. Client dự đoán dùng đúng công thức (chạy check:collision lại). Test: Thần Hành + Thiêu Đốt + theo gió = ×2.0; Shipper Gom Đơn + mọi buff ≤ ×3.0; Thiêu Đốt trong bùn = ×1.4.

## B. Kỹ năng hồi chiêu + tích lượt (§6.3)
3. Bỏ vật phẩm Bí Kíp (spawn, nhặt, quy đổi dư); art Bí Kíp chuyển vào assets/library (luật 13).
4. Mỗi nhân vật: `cooldowns[id] = [giây hồi 1 lượt, lượt tối đa]`; vào trận/hồi sinh 1 lượt; Tiên giới +1 lượt tối đa (rớt về Đại Thừa cắt bớt); khoá 3s giữa 2 lần; chết mất lượt.
5. Giảm: Đại Ca đàn em 30% kiếm, 8s; Bé Tí Gồng Cơ 4s.
6. UI nút kỹ năng: vòng hồi chiêu + chấm lượt + giây còn lại khi 0 lượt. Thẻ chọn nhân vật ở trang chủ hiện "Hồi Xs · tích N lượt".
7. U kỹ năng = 90 × scale (KHÔNG theo bán kính hút mới √scale).
8. Cân bằng: `cmd/sim --duel` mọi nhân vật trong 35–65%; chỉnh hồi chiêu ±30% để đạt, gửi tôi bảng hồi chiêu chốt + tỉ lệ thắng 10×10. Bot dùng kỹ năng khi có lượt.
9. Telemetry: ghi số lần dùng kỹ năng, Uy Danh/phút, thời gian sống theo nhân vật (để chỉnh sau 2 ngày).
