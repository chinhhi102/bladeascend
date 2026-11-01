# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

## Q36 — `cmd/sim --duel` trượt hàng rào 35–65 %, chỉnh hồi chiêu ±30 % không cứu được

**Bối cảnh:** làm CHANGES v6.26/v6.27 (sau giai đoạn 1). Sim đấu tay đôi Kim Đan, đủ trần
kiếm, 20 trận/cặp, 60 s/trận; thắng = chém chết đối thủ (sự kiện Death). Đã thử 2 bộ số:

| Nhân vật | Theo spec (hồi · lượt) | Thắng | Hồi chỉnh tới ±30 % | Thắng |
|---|---|---|---|---|
| Bé Tí | 35 s · 1 | **100 %** | 45 s | **99 %** |
| Sư Phụ Mì | 30 s · 2 | 69 % | 39 s | 74 % |
| Đại Ca Dép | 40 s · 1 | 69 % | 52 s | 73 % |
| A Hớn | 30 s · 2 | 68 % | 39 s | 72 % |
| Thằng Mập | 20 s · 2 | 52 % ✅ | 20 s | 43 % ✅ |
| Cụ Rùa | 30 s · 2 | 33 % | 21 s | 34 % |
| Shipper | 20 s · 2 | 26 % | 14 s | 28 % |
| Cô Ba | 22 s · 2 | 22 % | 15 s | 20 % |
| Bà Tám | 18 s · 2 | 17 % | 13 s | 13 % |
| Ông Ba Võng | 25 s · 2 | 17 % | 18 s | 19 % |

Hồi chiêu gần như không đổi kết quả: vào trận ai cũng có sẵn 1 lượt, và nay chém trúng là
chết nên **lần bấm đầu tiên quyết định trận**. Kỹ năng thắng va kiếm (Gồng Cơ ×2 độ bền,
cướp kiếm, hồi kiếm, đàn em) luôn thắng kỹ năng khống chế/thoát thân. Ngoài ra 35–38 % trận hoà.

**Phương án:**
- **A.** Đổi hiệu ứng, không đổi hồi: vd Gồng Cơ ×1.5 thay ×2 / 3 s; Hút Mì 15 %; A Hớn 12 %;
  đàn em 20 % kiếm. Anh duyệt số rồi em chạy lại sim.
- **B.** Nới hàng rào cho "kỹ năng gây khó chịu" (spec §6.3 đã cho phép yếu hơn khi đấu tay đôi):
  chỉ áp trần 65 % cho mọi người, bỏ sàn 35 %. Bé Tí vẫn phải sửa.
- **C.** Vào trận 0 lượt (phải chờ hồi lượt đầu) để hồi chiêu có tác dụng.

**Tạm thời:** giữ đúng bảng hồi chiêu của spec (không đổi số), sim báo trượt. Không tự đổi
hiệu ứng kỹ năng vì đó là luật trong spec.

**Ghi chú nhỏ cùng lượt:** §11.1 ví dụ "Trúc Cơ chém Nguyên Anh → 800" nay phải là 850 (mọi
lần chém trúng đều là chém chết, +50 — CHANGES v6.26 mục 4). Code theo 850; anh sửa ví dụ khi tiện.
Tương tự, checklist §17 dòng "Bị chém ở Luyện Khí… người chém nhận 50% EXP nạn nhân + 30" còn số
cũ; code theo v6.26 (30% + thưởng hạ, trần +2 cảnh giới).
