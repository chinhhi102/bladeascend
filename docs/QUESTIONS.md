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

