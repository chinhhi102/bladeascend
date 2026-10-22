# Việc cần làm — spec v3 (đã cập nhật docs/GAME_SPEC.md và docs/ART_PROMPTS.md)

Đọc lại các mục được nhắc tới trước khi code. Mọi số mới đã có trong mục 16 (config) — cập nhật config/game.json trước.

1. **Kiếm dài gấp đôi**: `length` 40 → 80, `width` 12. (GAME_SPEC §4.1a)
2. **Quay 300ms/vòng**: `rev_secs: 0.3` (ω ≈ 20.9 rad/s). Bắt buộc: va chạm theo góc quét (§4.1b) và vệt mờ chuyển động khi vẽ (chống hiệu ứng quay ngược). Tốc Kiếm giảm còn ×1.3.
3. **Kiếm tối đa** theo mảng `max_swords` [60,120,160,200,...,400]. **20 kiếm/vòng, vòng lồng nhau**: chuôi chung bán kính, vòng sau dài hơn 20% (tuyến tính theo vòng 1), vẽ vòng ngoài dưới cùng, vòng cuối thiếu kiếm vẫn thẳng hàng theo 20 khe. Camera lùi theo `reach`. (§3.1, §4.1a, §4.1)
4. **Bị chém rớt 1 đại cảnh giới, giữ tầng** (Trúc Cơ 2 → Luyện Khí 2). Luyện Khí bị chém = chết. Kiếm thừa rơi thành Bó Kiếm (≤10/bó). (§5.1)
5. **Bể kiếm** nhân `clash_unit` = 5 và thêm đẩy văng `clash_knockback`. (§5.2)
6. **EXP rải khắp bản đồ**: 1 viên / 140×140, mọc lại trong cùng ô 1000×1000 sau 5s, `pickup_r` 90. Chạy `cmd/sim` cho tới khi Luyện Khí → Trúc Cơ mất 60–90s. (§6.1)
7. **Vật phẩm**: 400/phòng, mọc lại sau 3s, item Kiếm +3, thêm Bó Kiếm +15, Rương Báu mỗi 60s nổ 8 item. (§6.2)
8. **Vật cản**: chặn thân bằng chân đế (đáy 30%), kiếm chém theo thân hình; y-sort khi vẽ, vật cản che người thì mờ 50%. (§8.2)
9. **Bản đồ dễ nhìn**: áp ColorMatrixFilter cho lớp nền + decor (giảm bão hòa ~50%, sáng ~10%, tương phản ~20%); lớp gameplay giữ nguyên + thêm bóng đổ và quầng sáng cho EXP/item/quái. Sinh lại tile nền theo prompt mới khi có thời gian. (ART_PROMPTS §9.0)
10. Cập nhật unit test theo checklist §17, đặc biệt: rớt cảnh giới giữ tầng, Luyện Khí bị chém là chết, vòng kiếm & góc khe, va chạm góc quét.
