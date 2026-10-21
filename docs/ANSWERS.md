# Trả lời QUESTIONS.md (đã cập nhật vào docs/GAME_SPEC.md)

## Q1 — EXP khi chém chết người ở Luyện Khí tầng 1
**Chốt: (b) + thưởng hạ sát.** `loss = total_exp hiện tại của nạn nhân`; người chém nhận 50% loss + `kill_bonus_exp` (30, mới trong config), rơi 30% loss thành viên EXP, +1 kill. Spec mục 5.1.

## Q2 — Đường gió
**Chốt: (b).** Lệch < 60° → ×1.5; lệch > 120° → ×0.8; còn lại ×1.0. Thêm `wind_with_max_deg: 60`, `wind_against_min_deg: 120` vào config. Spec mục 8.3.

## Q3 — Rương Báu
**Giữ phương án tạm.** Người đầu tiên chạm vào → nổ 5 item văng xung quanh, ai lượm cũng được. Spec mục 6.

## Q4 — EXP từ quái
**Giữ phương án tạm, bổ sung:** viên EXP thuộc về người chém hạ quái trong 1.5s đầu (`orb_owner_secs`), sau đó ai cũng nhặt được — tránh bị cướp công. Spec mục 5.3.

## Q5 — Chết có rơi kiếm không
**Đổi: có rơi.** Chết (người và bot) → `floor(n * death_sword_drop)` kiếm (mặc định 0.5) văng ra thành item kiếm quanh xác. Người chém có phần thưởng, người khác có cơ hội "hôi của". Thêm `death_sword_drop: 0.5` vào config. Spec mục 7.

→ Việc cần làm: cập nhật config/game.json, sửa code Q1/Q2/Q4/Q5, thêm unit test tương ứng (xem checklist mục 17), xóa các câu đã trả lời khỏi QUESTIONS.md.
