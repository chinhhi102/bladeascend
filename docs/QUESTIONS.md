# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

Bối cảnh chung: lượt cuối 01/11/2026; cập nhật 02/11: Q37–Q42 đã xong (spec v6.35). Chi tiết từng dòng: `docs/SPEC_AUDIT_v630.md`.
Mỗi câu ghi **phương án đang dùng** (an toàn, không đổi luật) — anh chỉ cần trả lời nếu muốn khác.

## Q45 — spec v6.38 bỏ hướng A nhưng hai chỗ trong spec vẫn đòi nó

Tìm ra 05/11 khi rà lại spec so với code. Không chặn gì — chỉ là lần rà sau sẽ chấm sai nếu
để nguyên.

| Chỗ | Đang ghi | Vấn đề |
|---|---|---|
| Dòng 1284 (§12.5 A) | "Biểu cảm … là **lớp mặt riêng** chồng lên đầu … Đổi mood = **chỉ đổi mặt**" | Mô tả đầy đủ hướng đã bỏ, nằm ngay dưới ghi chú v6.38 nói đã bỏ |
| Dòng 1748 (§17) | "Đổi run → chase → flee: … **chỉ đổi mặt + tư thế**" | Vế "chỉ đổi mặt" nay KHÔNG đạt được theo thiết kế |

**Phương án đang dùng:** để nguyên, không sửa gì trong code. Bong bóng mặt đã làm đúng v6.38.

Hỏi: anh muốn sửa hai dòng đó thế nào?
- Dòng 1284: giữ làm lịch sử (ghi rõ "ĐÃ BỎ — xem Q44") hay xoá hẳn?
- Dòng 1748: bỏ vế "chỉ đổi mặt", giữ hai vế còn lại (không reset frame · không lật qua lại ·
  tốc độ bước chân khớp) — cả ba vế đó đều đang ĐẠT trong `animctl.ts`. Có thể thêm một dòng
  §17 mới cho bong bóng mặt nếu anh muốn nó nằm trong checklist nghiệm thu.
