# Câu hỏi về thiết kế (cần anh xác nhận)

Mỗi câu: bối cảnh → các phương án → **tạm dùng** (đã code theo phương án này, đổi được dễ).

Q1–Q6 đã trả lời trong `docs/ANSWERS.md` và đã áp dụng.

## Q7. (đã trả lời 2026-10-22) Nhịp Luyện Khí → Trúc Cơ

Anh chốt: **30–60s là được** (spec §6.1 đang ghi 60–90s, nhờ sửa lại cho khớp). Đang dùng `exp_orbs.cell = 160` → sim `--farm` đo được 40–46s. Các cảnh giới sau giữ nguyên luật (đo: Kim Đan ≈ 1 phút, Nguyên Anh ≈ 1.2 phút).

## Q8. Checklist §17 còn số kiếm mất khi va chạm theo bản cũ

**Bối cảnh:** §5.2 mới ghi mỗi lần va mất `clash_unit` = 5 kiếm (bảng: Trúc Cơ −5 / Luyện Khí −10…), nhưng checklist §17 vẫn ghi "mỗi bên mất 1 kiếm", "Luyện Khí mất 2, Trúc Cơ mất 1", "Luyện Khí mất 4, Nguyên Anh mất 1".

**Tạm dùng:** theo §5.2 (nhân 5), unit test đã viết theo §5.2. Anh sửa giúp checklist §17 nếu đúng vậy.
