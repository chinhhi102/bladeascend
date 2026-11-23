# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

Bối cảnh chung: lượt cuối 01/11/2026; cập nhật 02/11: Q37–Q42 đã xong (spec v6.35). Chi tiết từng dòng: `docs/SPEC_AUDIT_v630.md`.
Mỗi câu ghi **phương án đang dùng** (an toàn, không đổi luật) — anh chỉ cần trả lời nếu muốn khác.


## Q56 — mốc kỷ lục gác MỨC NỀN, mà mức nền chênh 31% giữa ba lượt giống hệt nhau

**Bối cảnh.** Ba lượt nghiệm thu 60 phút, cùng commit, cùng kịch bản cố định, máy rảnh
(`tools/leakrepeat.sh 3 60`, PERF_LOG #55):

| | dốc heap | mức nền (phút 10) |
|---|---:|---:|
| lượt 1 | +1,565 MB/phút | **88,0 MB** |
| lượt 2 | +1,317 | **112,2** |
| lượt 3 | +1,424 | **115,1** |

**Dốc đã tái lập được** (chênh 1,2 lần — phép nghiệm thu của v6.57 ĐẠT). **Mức nền thì
chưa**: lượt 2 và 3 khớp nhau ~3%, riêng lượt 1 thấp hơn 24% và thấp đều từ phút 4 trở đi.

**Vì sao nó đụng vào luật.** Mốc kỷ lục §13.9 lấy **mức nền thấp nhất từng đo** rồi chặn
build vượt > 10%. Với khoảng 31% này: nếu lượt 1 đặt mốc 88 thì lượt 2 và lượt 3 — cùng
build, cùng kịch bản, máy rảnh, không rò rỉ gì thêm — đều ĐỎ vì "vượt mốc 28%". Cổng sẽ đỏ
vì nhiễu, và một cổng đỏ vì nhiễu thì sớm muộn bị bỏ qua.

**Phương án đang dùng (không đổi luật):** giữ nguyên v6.58 — mốc vẫn gác, vẫn 10%.

**Đề xuất:** làm cho nó **đối xứng**. Spec đã bắt *hạ* mốc phải có 2 đêm xác nhận; đề nghị
*vượt* mốc cũng cần **2 đêm liên tiếp** mới ĐỎ (đêm đầu là thẻ vàng). Một dòng luật, bịt
đúng chỗ nhiễu, và không nới ngưỡng 10%.

Hai cách khác nếu anh thích hơn: (b) mốc so với **trung vị 3 đêm hợp lệ gần nhất** thay vì
giá trị thấp nhất từng đo; (c) nới ngưỡng 10% → 35%, nhưng cách này làm cổng gần như mất
tác dụng với một thoái lui thật 20–30%.
