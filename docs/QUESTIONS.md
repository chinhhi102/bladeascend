# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

## Q34 — `bot_fill_target`: spec ghi 20, anh chốt 60 hôm 29/10

**Bối cảnh:** lượt lớn, mục 1.2 (v6.22). Spec §9.3 và khối config §16 ghi
`bot_fill_target (20)`. Code đang để **60** theo quyết định của anh ngày 29/10 (commit
`0561a52`: trần phòng 60 nhân vật, tối đa 20 người thật, sàn 10 nhân vật mỗi khu).
Hai con số không đi được với nhau: 5 khu × sàn 10 = 50 nhân vật, mà trần phòng 20 thì một
người vào sẽ thấy mỗi khu chỉ 4 nhân vật.

**Phương án:**
- **A.** Giữ 60 (và sàn 10/khu). Anh sửa spec §9.3 + §16 thành 60.
- **B.** Theo spec: 20, đồng thời hạ `zone_min_chars` xuống 4. Thế giới 5 khu sẽ vắng hẳn.

**Tạm thời:** giữ **60** (A). Đây là quyết định đã chạy ổn từ 29/10, và B làm thế giới vắng.

## Q35 — Phòng 50 người chưa qua loadtest trên server thật → đang giữ 20

**Bối cảnh:** lượt lớn, mục 1.2. Spec §9.1 cho phép mở 50 người thật/phòng sau khi loadtest
50 người + bot đạt tick p95 mỗi khu < 4 ms. Kết quả (PERF_LOG lượt #24):

| Máy | p95 khu nặng nhất | Kết quả |
|---|---|---|
| Máy dev (i5-12400F) | 2.02 ms | đạt |
| Staging (ARM 4 nhân), 3 lần đo | 5.19 / 4.28 / 5.16 ms | **không đạt** |

Phần vượt chỉ nằm ở **một khu**: khu có 10 người Đạo Tổ đủ kiếm đứng chụm (cảnh nặng nhất).
Bốn khu kia chỉ 1,6–2,3 ms.

**Phương án:**
- **A.** Giữ 20 cho tới khi tối ưu được khu nặng (đưa vào việc tối ưu sau).
- **B.** Mở 50 trên máy chủ mạnh hơn (production khác staging?). Anh cho biết máy production.
- **C.** Mở 50 ngay, chấp nhận khi 10 Đạo Tổ chụm thì khu đó hơi trễ.

**Tạm thời:** **A**, giữ `room.max_humans = 20`, đúng như lệnh "không đạt thì giữ 20".


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
