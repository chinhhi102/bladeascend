# Việc cần làm — GAME_SPEC v6.27 (khống chế không khoá kiếm, Thằng Mập)

Làm cùng CHANGES_v6_26.

1. Luật chung: mọi hiệu ứng khống chế (choáng, ngủ, trói, trượt) chỉ khoá di chuyển / kỹ năng / Thiêu Đốt / tầm nhìn. Vòng kiếm của người bị khống chế luôn quay + va kiếm hai chiều + chém thân bình thường. Rà mọi trạng thái trong hệ "trạng thái có thời hạn" và xoá mọi chỗ dừng quay / va kiếm một chiều.
2. Ông Ba Võng Ru Ngủ: bỏ "kiếm ngủ không đánh trả" và "ngừng quay"; giữ đứng im 2.5s + khoá kỹ năng/Thiêu Đốt. Hồi 25s, tích 2.
3. Thằng Mập bịt mắt 5s: thêm cờ `blind_no_body_hit` — kiếm người bị mù không gây chết khi chạm thân (va kiếm vẫn tính). Vòng kiếm người mù có lớp trân châu mờ để người khác nhận ra. Hồi 20s, tích 2. Bot bị mù cũng vậy.
4. Test theo 2 dòng mới ở checklist §17; xoá test cũ "chỉ người ngủ bị bể kiếm". Chạy lại cmd/sim --duel (hàng rào 35–65%).
