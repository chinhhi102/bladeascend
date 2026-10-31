# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

## Q28 — Ảnh góp ý: giữ bao lâu, dọn thế nào?

**Bối cảnh.** A3 (ADMIN_SPEC §7) lưu ảnh chụp màn hình vào `server/data/feedback/`, mỗi ảnh
tối đa 300 KB. Spec không nói hạn giữ. Telemetry có hạn 30 ngày (TELEMETRY_SPEC §3) nhưng
góp ý thì khác — một báo lỗi cũ vẫn có giá trị khi lỗi đó tái xuất hiện.

**Phương án.**
- A. Giữ mãi. Đơn giản, nhưng 100 ảnh/ngày × 300 KB = 11 GB/năm.
- B. Dọn ảnh của thư đã ở trạng thái "Xong"/"Bỏ qua" quá 90 ngày; giữ chữ mãi.
- C. Dọn mọi ảnh quá 180 ngày.

**Tạm dùng: A (giữ mãi).** Không tự xoá dữ liệu người chơi là luật anh đặt cho lượt này.
Dung lượng chưa phải vấn đề ở quy mô hiện tại; khi nào chạm vài GB thì chọn B hay C.

## Q29 — Cờ "Tăng điểm bất thường" (§6.1) cần mốc trung vị từ đâu?

**Bối cảnh.** §6.1 định nghĩa cờ này là "Uy Danh / Linh khí tăng nhanh hơn 5× trung vị cùng
cảnh giới trong 2 phút". Cần một bảng trung vị theo cảnh giới để so.

**Phương án.**
- A. Tính trung vị TRỰC TIẾP từ người đang chơi trong phòng. Tự thích nghi, nhưng phòng ít
     người thì trung vị nhiễu, và nếu cả phòng đều hack thì không ai bị gắn cờ.
- B. Tính sẵn từ `cmd/sim` rồi chốt vào `config/game.json`. Ổn định, nhưng phải cập nhật
     mỗi lần đổi cân bằng.
- C. Lấy trung vị trượt 24 giờ từ bảng `lives` (dữ liệu người chơi thật đã lọc debug).

**Tạm dùng: C**, rơi về B khi chưa đủ dữ liệu (< 50 mạng chơi trong 24 giờ). C dùng dữ liệu
thật nên không lệch khi cân bằng đổi, mà vẫn không nhiễu như A.
