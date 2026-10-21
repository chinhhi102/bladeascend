# Câu hỏi về thiết kế (cần anh/chị xác nhận)

Mỗi câu: bối cảnh → các phương án → **tạm dùng** (đã code theo phương án này, đổi được dễ).

## Q1 — Chém chết người ở Luyện Khí tầng 1 thì người chém nhận bao nhiêu EXP?
- Milestone: M2 · File: `server/internal/game/combat.go` (`kill`)
- Spec 5.1 chỉ định nghĩa `loss = exp_needed(level ngay dưới)` — không có "level ngay dưới" của tầng 1.
- Phương án: (a) không nhận gì; (b) coi `loss = total_exp hiện tại của nạn nhân` (luôn < 100); (c) một số cố định trong config.
- **Tạm dùng (b)**: người chém nhận 50%, rơi 30% thành viên EXP — giống công thức thường.

## Q2 — Đường gió: thế nào là "theo hướng gió" / "ngược gió"?
- Milestone: M4 · File: `server/internal/game/movement.go` (`ZoneMult`)
- Phương án: (a) góc giữa hướng chạy và hướng gió < 90° là theo, còn lại là ngược; (b) có vùng "ngang gió" ×1.0 khi lệch gần 90°.
- **Tạm dùng (a)**.

## Q3 — Rương Báu mở thế nào?
- Milestone: M5 · File: `server/internal/game/items.go`
- Spec ghi "nổ ra 5 item" và emote khi "nhặt Rương Báu", không nói là chạm vào hay phải chém.
- **Tạm dùng**: chạm vào (như nhặt đồ) → rương nổ ra 5 item ngẫu nhiên xung quanh.

## Q4 — Quái chết: EXP vào thẳng người chém hay rơi thành viên EXP?
- Milestone: M1 · File: `server/internal/game/combat.go` (`killMonster`)
- Core loop ghi "chém quái → lượm EXP", bảng mục 6 có "Viên EXP … rơi từ quái".
- **Tạm dùng**: toàn bộ EXP của quái rơi thành viên EXP (ai lượm cũng được).

## Q5 — Chết có làm rơi kiếm không?
- Milestone: M2 · Spec chỉ nói kiếm thừa rơi khi rớt cảnh giới.
- **Tạm dùng**: chết không rơi kiếm (vào lại với 1 kiếm theo mục 7).
