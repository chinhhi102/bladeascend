# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

Bối cảnh chung: lượt cuối 01/11/2026; cập nhật 02/11: Q37–Q42 đã xong (spec v6.35). Chi tiết từng dòng: `docs/SPEC_AUDIT_v630.md`.
Mỗi câu ghi **phương án đang dùng** (an toàn, không đổi luật) — anh chỉ cần trả lời nếu muốn khác.

> Hiện KHÔNG có câu nào đang chờ. Mọi câu đã trả lời nằm ở `docs/QUESTIONS_DONE.md`.

## JS heap đang gấp đôi mục tiêu của spec — có nâng thành điều kiện chặn không?

**Bối cảnh.** §13.10 ghi: *"JS heap mục tiêu ≤ 80 MB (đo riêng)"*. Từ trước tới nay **không
bài đo nào canh con số này** — cổng rò rỉ §13.9 chỉ hỏi "có TĂNG quá 5% không", nên một
build đứng yên ở 190 MB vẫn qua cổng.

**Số đo (21/11, lượt 52 phút, 100 lượt đổi khu, máy rảnh):**

| | |
|---|---|
| JS heap ở phút 20 | ~136 MB |
| JS heap ở phút 50 | ~185 MB |
| mục tiêu §13.10 | **≤ 80 MB** |

Tức là vượt mục tiêu **ngay từ lúc mới ổn định**, rồi còn tăng tiếp ~1 MB/phút (nút thắt
B38). Texture và GPU MB thì sạch — đây là chuyện của JS, không phải của GPU.

**Đã làm:** `tools/browsertest/leakcheck.mjs` nay in con số này mỗi lượt, kèm sổ tài nguyên
theo loại để biết loại nào phình.

**Câu hỏi.** §13.9 liệt kê đích danh những gì CHẶN deploy (cửa sổ ổn định tăng > 5%, GPU MB
vượt ngân sách thiết bị); 80 MB nằm ở §13.10 và được gọi là "mục tiêu". Tôi để nó ở mức
**cảnh báo**, không tự thêm điều kiện chặn (luật 4). Chủ dự án chốt giúp:

1. Giữ cảnh báo, sửa dần — hay
2. Nâng thành điều kiện chặn như GPU MB. Nếu chọn cách này thì cần chốt mốc: chặn ngay ở
   80 MB (hiện đang trượt, tức là chặn mọi lượt deploy cho tới khi sửa xong), hay đặt một
   mốc tạm (vd 140 MB) rồi siết dần?
