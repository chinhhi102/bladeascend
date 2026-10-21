# Câu hỏi về thiết kế (cần anh/chị xác nhận)

Mỗi câu: bối cảnh → các phương án → **tạm dùng** (đã code theo phương án này, đổi được dễ).

## Q6 — "Mỗi sheet chỉ chứa 1 animation" áp dụng cho cả quái / cổng / VFX?
- Milestone: M6 · File: `server/cmd/artgen/catalog.go`
- ART_PROMPTS §5.1 (mới) ghi "Mỗi sheet chỉ chứa **1 animation**" trong phần sửa lỗi nhân vật. Tách hết thì tốn thêm khoảng 25 ảnh quota.
- Phương án: (a) chỉ tách sheet **nhân vật** (lỗi lưới lệch thấy ở nhân vật); (b) tách tất cả (quái 3 ảnh/con, cổng 3 ảnh/màu, VFX 1 ảnh/hiệu ứng).
- **Tạm dùng (a)**: nhân vật 1 anim/ảnh (120 ảnh); quái, cổng, VFX, huy hiệu, đồ trang trí vẫn gộp nhiều thứ vào 1 ảnh để tiết kiệm quota.
