# Việc cần làm — spec v6.9 (làm lại 3 kỹ năng)

Đọc GAME_SPEC §6.3 (3 dòng mới trong bảng + mục "Nguyên tắc thiết kế kỹ năng"). Config `skills.ranges_u`: bỏ `luoi_mi_*`; thêm `ru_ngu_secs`, `ru_ngu_channel`, `hut_mi_cone`, `hut_mi_steal`, `hut_mi_cap_of_own_max`, `gom_don_r`, `gom_don_secs`; `ru_ngu_r` 5 → 4.

1. **Ông Ba Võng — Ru Ngủ Trưa**: niệm 0.5s (bị chém thân trong lúc niệm → huỷ, mất lượt). Địch trong 4 U: trạng thái `asleep` 2.5s = đứng im + vòng kiếm ngừng quay + **va kiếm một chiều** (chỉ người ngủ bị bể; người thức không mất kiếm, không tăng crack). Bị chém thân → tỉnh. Dùng chung hệ trạng thái có thời hạn; Shipper Gom Đơn miễn.
2. **Sư Phụ Mì — Hút Mì Đoạt Kiếm**: chọn 1 địch gần nhất trong nón 60° dài 5 U. Cướp `floor(25% kiếm hiện có của địch)` từ vòng ngoài cùng, giới hạn `20% kiếm tối đa của Sư Phụ`; nhập vào vòng Sư Phụ tới trần, phần dư rơi Bó Kiếm. Hụt → mất lượt. Kiếm bay dọc sợi mì 0.5s (client), server áp ngay.
3. **Shipper — Gom Đơn Hỏa Tốc**: 4s tốc ×3, xuyên mọi vật cản trừ biên, miễn khống chế (huỷ trạng thái chậm/trói/choáng/ngủ đang có khi kích hoạt + miễn mới), hút mọi viên/vật phẩm/Bó Kiếm trong 3 U mỗi tick (bỏ qua luật bịt mắt không nhặt đồ). Hết giờ kẹt trong vật cản → đẩy ra mượt.
4. Bot dùng theo luật mới (§6.3 dòng bot).
5. Test theo 3 dòng mới ở checklist §17. Sau đó cho tôi số liệu thật từ telemetry: tỉ lệ chọn 10 nhân vật trước/sau bản này (sau 2–3 ngày).

## v6.10 — Cô Ba đổi kỹ năng (không trùng Shipper)
6. **Cô Ba — Bày Hàng Chặn Lối** (thay Gánh Lướt): đặt 5 sạp hình vòng cung cách 3 U theo hướng nhắm, dài 6 U, tồn tại 6s. Sạp là vật cản tạm có `min_realm` = cảnh giới Cô Ba (dùng chung hệ vật cản §8.2, không mọc lại, hết giờ biến mất). Bỏ đoạn trùng người. Vỏ chuối rộng 1 U dọc chân sạp: dẫm → trượt + choáng 1s. Shipper Gom Đơn đi xuyên.
7. Config: bỏ `ganh_dash`, `ganh_trail_w`; thêm `bay_sap_dist`, `bay_sap_len`, `bay_sap_count`, `bay_sap_secs`, `vo_chuoi_w`. Hướng nhắm: chuột (PC), hướng joystick cuối (mobile).
8. Test theo dòng mới ở checklist §17; bot theo luật mới.

## v6.11 — Sạp Cô Ba cao hơn 1 bậc
9. Sạp tạm của Cô Ba: `min_realm = min(13, cảnh giới Cô Ba + bay_sap_tier_bonus)` (= +1). Cô Ba (chủ sạp) đi xuyên sạp của mình (`bay_sap_owner_passes`), người khác va chạm bình thường. Sửa test: Cô Ba Kim Đan → sạp bậc 3, người Kim Đan không chém được, Nguyên Anh chém được.
