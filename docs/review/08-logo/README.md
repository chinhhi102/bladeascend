# Mục 8 — Logo thư pháp (ỨNG VIÊN, chưa dùng)

Sinh bằng artgen/minigen ngày 01/11/2026, job `brand_logo_vi` + `brand_logo_frame`
(ảnh gốc: `assets/src/_raw/brand_logo_vi.png`, `brand_logo_frame.png` — không xoá).

- `ung_vien_logo_vi.jpg` — chữ thư pháp mực + vàng lá.
- `ung_vien_khung_logo.jpg` — khung vàng hai kiếm chéo (nền hồng sẽ được xoá khi `artgen prep`).

**Trang chủ vẫn dùng logo chữ (font) hiện tại** cho tới khi anh duyệt.

Soi kỹ dấu trước khi duyệt — Claude thấy nghi:
1. "Đạo": chấm nặng nằm tách xa dưới chữ, dễ đọc thành "Đao".
2. "Dừng": dấu trên ư trông giống dấu hỏi ("Dủng") hơn là dấu huyền.
3. Logo xếp 3 dòng — trang chủ hiện xếp 1 dòng (PC) / 2 dòng (điện thoại).

Nếu chưa ưng: bảo Claude chạy lại
`cd server && go run ./cmd/artgen run -only brand_logo_vi -redo -n 1` (mỗi lần 1 ảnh hạn mức).
Duyệt xong: sửa `tools/make_brandart.py` để ghép từ tệp nguồn → trang chủ, màn tải, ảnh chia sẻ đổi cùng lúc.
