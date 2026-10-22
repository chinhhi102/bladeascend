# Câu hỏi đã xong (lưu trữ)

Những câu đã được trả lời — giữ lại để tra sau, không cần đọc nữa.
Câu **đang chờ** nằm ở `docs/QUESTIONS.md`.

| Câu | Nội dung | Kết luận |
|---|---|---|
| Q1–Q6 | Thưởng hạ sát, gió theo góc, viên EXP thuộc người chém, chết rơi kiếm, 1 anim/ảnh… | Trả lời trong `docs/_raw/ANSWERS.md`, đã áp dụng |
| Q7 | Nhịp Luyện Khí → Trúc Cơ | Chốt 30–60s (nay đã đổi lại, xem Q13) |
| Q8 | Checklist §17 ghi số kiếm mất theo bản cũ | Spec gốc đã sửa: `clash_unit` = 1, checklist khớp |
| Q9 | `min_count = 0` có làm chết không | Không. Spec §5.2: "0 kiếm **không chết**, chỉ bị tước vũ khí" |
| Q10 | Phạm vi nhặt vật phẩm | Giữ `pickup_r` = 90. Luật vốn đã đúng: kiếm không nhặt được đồ |
| Q11 | "Tính exp dư ra làm…" | Spec §3.2: Đại Thừa tầng 10 thì EXP thừa **bị bỏ**, xếp hạng dùng Uy Danh |
| Q12 | Tốc độ chạy 130 hay 260×sqrt(scale) | Gộp vào Q14 (nằm trong bảng 6 con số lệch) |

---

## Chi tiết

### Q7. Nhịp Luyện Khí → Trúc Cơ *(2026-10-22)*
Anh chốt **30–60s**. Lúc đó dùng `exp_orbs.cell = 160` → sim `--farm` đo 40–46s.
→ Sau đó anh đổi EXP gấp đôi + giảm mật độ 50% nên nhịp này không còn đúng. Xem **Q13**.

### Q8. Checklist §17 còn số kiếm mất theo bản cũ *(2026-10-22)*
§5.2 ghi mỗi lần va mất `clash_unit` kiếm, nhưng checklist §17 ghi số khác.
→ **Xong:** spec gốc v3.5 đặt `clash_unit = 1` và sửa checklist cho khớp ("mỗi bên mất 1 kiếm").

### Q9. `min_count = 0` không làm người chơi chết *(2026-10-22)*
Anh bảo "dù chỉ còn 1 kiếm cũng không chết → min kiếm = 0". Em chỉnh `min_count` = 0 nhưng
báo lại rằng số kiếm chưa bao giờ quyết định sống chết — `applyBodyHit` chỉ cho chết khi
nạn nhân đang ở Luyện Khí.
→ **Xong:** spec gốc §5.2 chốt luôn: *"0 kiếm **không chết**: chỉ bị tước vũ khí. Chém trúng
thân vẫn theo §5.1."* Đúng phương án đang code.

### Q10. Phạm vi nhặt vật phẩm *(2026-10-22)*
Anh bảo "phạm vi kiếm không được nhặt đồ, chỉ phạm vi nhân vật". Kiểm tra thì luật đã đúng
sẵn: `resolvePickups` chỉ dùng `pickup_r × scale`, kiếm chỉ đánh quái và vật cản.
Cảm giác "kiếm hút đồ" là do vòng nhặt rộng bằng ~77% vòng kiếm.
→ **Xong:** spec gốc giữ `pickup_r = 90`, không đổi gì.

### Q11. "Tính exp dư ra làm…" — câu bị cắt *(2026-10-22)*
Câu nhắn đứt giữa chừng nên em chưa dám đoán.
→ **Xong:** spec gốc §3.2 đã ghi rõ: *"Đạt Đại Thừa tầng 10: `total_exp` dừng ở trần, EXP
nhặt thêm bị bỏ (xếp hạng dùng Uy Danh §11, không dùng EXP)."* Đúng với code hiện tại.

### Q12. Tốc độ chạy: spec cứ lùi về 260 × sqrt(scale) *(2026-10-22)*
Anh nhắn "giảm 1/2, chỉ item mới tăng tốc" → code `move_base` = 130, bỏ `sqrt(scale)`.
Spec dán vào sau đó nhiều lần vẫn ghi 260 × sqrt(scale).
→ **Gộp vào Q14**: nó là 1 trong 6 con số mà `docs/_raw/` còn ghi giá trị cũ. Code vẫn giữ 130.
