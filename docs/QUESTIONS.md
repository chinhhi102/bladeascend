# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

## Q33 — Lạm phát EXP: nguyên nhân thật khác nghi phạm của §6.0 (v6.21)

**Bối cảnh.** CHANGES v6.21 bước 1. Chi tiết và số đo nằm trong `docs/EXP_AUDIT.md`.
Đo bằng `cd server && go run ./cmd/sim --audit`.

- **Nguyên nhân chính là lỗi code.** Ở `server/internal/game/items.go:220`, nhặt Bí Kíp khi đã đủ
  lượt cộng `TotalExp + 20` thay vì +20, tức nhân đôi tổng EXP mỗi lần nhặt. Bí Kíp chiếm 80–97 %
  EXP. Tái hiện được: Luyện Hư → Đại Thừa 10 trong ~70 s khi có Nam Châm.
- Viên EXP và quái trong code **không** nhân theo cảnh giới (spec nghi ×2^r).
- Bỏ lỗi Bí Kíp đi thì farm thuần hiện tại chậm hơn bảng §6.0. Áp nguyên văn §6.0
  (×1,5^r, √scale, Nam Châm ×1,5) thì Đại Thừa nhanh gấp đôi bảng, và Nam Châm vượt trần 1,5 lần.

**Câu hỏi.**
1. Sửa lỗi Bí Kíp về +20 như spec — đồng ý làm ngay?
2. Bảng nhịp §6.0 và các hệ số §6.0 không thể cùng đạt. Chọn:
   - **A (khuyên).** Giữ BẢNG làm chuẩn. Claude chỉnh `exp_economy` bằng `cmd/sim --farm` cho
     trúng ±25 %, rồi anh sửa chữ §6.0 theo hệ số chốt.
   - B. Giữ hệ số §6.0, nới bảng cho khớp số đo.
   - C. Chỉ sửa Bí Kíp + 25 % khi chém bot, giữ nguyên phần còn lại.

**Tạm dùng:** chưa sửa gì. Chờ anh trả lời.

