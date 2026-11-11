# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

Bối cảnh chung: lượt cuối 01/11/2026; cập nhật 02/11: Q37–Q42 đã xong (spec v6.35). Chi tiết từng dòng: `docs/SPEC_AUDIT_v630.md`.
Mỗi câu ghi **phương án đang dùng** (an toàn, không đổi luật) — anh chỉ cần trả lời nếu muốn khác.

## Q52 — Hoá Khổng Lồ nhân `scale`, nên vòng NHẶT và tầm KỸ NĂNG cũng to theo trong 2 giây

Spec §6.3 v6.48 ghi Hoá Khổng Lồ ×2 cho **ba** thứ: thân, bán kính vòng kiếm, số kiếm.

Cách làm hiện tại là nhân thẳng `scale` của nhân vật ×2. Mọi thứ tính theo scale đi theo:
đúng ba thứ spec nói, **cộng thêm** bán kính nhặt đồ (×√2 theo §6.0) và tầm mọi kỹ năng
(U = 90 × scale). Trong 2 giây, hồi 25 giây.

Tôi để nguyên vì nó hợp lý về hình ảnh (người khổng lồ thì với xa hơn thật) và tự giới hạn
bởi 2 giây. **Nếu anh muốn đúng ba thứ**, phải tách "scale để VẼ" khỏi "scale để TÍNH
LUẬT" — làm được, nhưng đó là một lượt riêng và phải canh kỹ: hai con số lệch nhau là hình
vẽ và vùng ăn đòn lệch nhau.

Một hệ quả nữa, không phải lỗi: máy quay tự lùi ra khi nhân vật to lên (§12 — khung nhìn
theo `scale^camera_exp`), nên trên màn hình thân chỉ TO RA 1,41 lần chứ không phải 2 — còn
trong thế giới thì đúng 2. Đổi được nếu anh thấy chướng.

> Hiện KHÔNG có câu nào đang chờ. Mọi câu đã trả lời nằm ở `docs/QUESTIONS_DONE.md`.
