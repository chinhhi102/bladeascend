# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

Bối cảnh chung: lượt cuối 01/11/2026; cập nhật 02/11: Q37–Q42 đã xong (spec v6.35). Chi tiết từng dòng: `docs/SPEC_AUDIT_v630.md`.
Mỗi câu ghi **phương án đang dùng** (an toàn, không đổi luật) — anh chỉ cần trả lời nếu muốn khác.

## Q48 — Nhìn người chạy Gom Đơn từ tab khác: đứng hình rồi nhảy cóc (CÓ TỪ TRƯỚC v6.45)

**Bối cảnh**: nghiệm thu v6.45 có dòng "nhìn Shipper đang Gom Đơn từ một tab khác — không
giật, không nhảy cóc". Tôi viết bài đo `tools/browsertest/gomdon-smooth.mjs`: một tab đứng
xem, một tab chạy Shipper; lấy mẫu vị trí VẼ RA mỗi khung, chuẩn hoá theo thời gian khung
rồi so với trần tốc `move_base × cap_gom_don`.

| Kịch bản | Trần tốc | Tốc vẽ lớn nhất | Khung nhảy cóc | Khung đứng hình | Dựng lại entity |
|---|---:|---:|---:|---:|---:|
| Đi bộ, KHÔNG Gom Đơn | 975 | 606 | **0** | 0 | 0 |
| Gom Đơn, cấu hình CŨ ×3 | 585 | 3 025 (5,2× trần) | **40** | 95 | 0 |
| Gom Đơn, cấu hình MỚI ×5 | 975 | 4 979 (5,1× trần) | **23** | 56 | 0 |

**v6.45 KHÔNG gây ra lỗi này** — cấu hình cũ ×3 còn nhiều khung hỏng hơn, và tỉ lệ vượt
trần gần như y nhau (5,2× so với 5,1×). Đây là lỗi có sẵn, chỉ là chưa ai đo.

**Không phải do va chạm phía client**: `moveStep` (bản port của `resolveBlocking`) chỉ chạy
cho CHÍNH MÌNH (`state.ts:617`), người khác thuần nội suy. Đối chiếu va chạm client↔server
vẫn **lệch 0,0000 đơn vị**.

**Nghi nhất là kho nội suy cạn**: `interp_ms` 100 với `snapshot_hz` 20 (một gói mỗi 50 ms)
chỉ dư đúng một gói. Số liệu khớp: khung đứng hình và khung nhảy cóc đi theo cặp
(56 ↔ 23, 95 ↔ 40) — giữ nguyên vị trí vì hết gói, rồi bù một nhát khi gói tới. Ở tốc đi bộ
một nhát bù là vài đơn vị nên không ai thấy; ở 975 đv/s nó là hàng chục đơn vị.

**Phương án đang dùng**: không đổi gì. Đây là chỉnh netcode (§NETCODE_SPEC), không nằm
trong v6.44/v6.45, và tôi không tự ý đụng vào nhịp mạng.

**Hướng xử lý nếu anh muốn làm** (cần đo lại sau mỗi hướng):
1. **Nâng `interp_ms` lên 150** — dư hai gói thay vì một. Đổi một số, rủi ro thấp nhất; cái
   giá là người khác hiện trễ thêm 50 ms.
2. **Nâng `snapshot_hz` lên 30** (bằng `tick_hz`) — mượt nhất, nhưng tốn băng thông; phải đo
   lại §B3b trong PERF_LOG trước khi chốt.
3. **Ngoại suy khi hết gói**: client tự đoán tiếp theo vận tốc cuối thay vì đứng im. Mượt mà
   không tốn mạng, nhưng đoán sai thì giật kiểu khác — việc lớn nhất trong ba hướng.
