# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

Bối cảnh chung: lượt cuối 01/11/2026; cập nhật 02/11: Q37–Q42 đã xong (spec v6.35). Chi tiết từng dòng: `docs/SPEC_AUDIT_v630.md`.
Mỗi câu ghi **phương án đang dùng** (an toàn, không đổi luật) — anh chỉ cần trả lời nếu muốn khác.

## Q55 — §13.9 còn một dòng ghi "client 30 phút", v6.57 đã đổi thành 60

Không phải câu hỏi về luật chơi, chỉ là một dòng spec sót lại. Khối **"Bài kiểm rò rỉ
client"** của §13.9 (v6.57) nói rõ: 60 phút, phút 0–10 khởi động, cửa sổ đo 10–60. Nhưng
dòng **"Lịch chạy"** ngay dưới đó vẫn ghi *"bài rò rỉ client 30 phút + server 60 phút chạy
hằng đêm 03:00"*.

**Phương án đang dùng:** làm theo khối v6.57 — `tools/leaknightly.sh` chạy client **60
phút** (`LEAK_CLIENT_MINUTES` mặc định đổi 30 → 60), server giữ 60 phút game. Claude không
sửa `docs/_raw/`, nên nhờ anh sửa con số 30 ở dòng "Lịch chạy" khi tiện.

**Hệ quả cần biết:** đêm nào cũng chiếm thêm nửa tiếng máy (bài client xong lúc ~04:10, còn
job janitor 04:30 thì không đụng nhau). Nếu anh muốn giữ 30 phút để đỡ tốn máy thì nói —
nhưng cửa sổ 20 phút sẽ không đủ cho 5 khối × 10 phút của chính v6.57.
