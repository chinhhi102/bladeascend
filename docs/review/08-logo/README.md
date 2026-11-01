# 08 — Logo Pattaya (02/11/2026)

Logo chủ dự án gửi (`docs/_raw/logo.zip`) đã gắn vào game:

| Chỗ | Logo | Ảnh duyệt |
|---|---|---|
| Trang chủ PC (tiếng Việt) | `logo_vi_1dong` | `home_vi_pc.png` |
| Trang chủ điện thoại (tiếng Việt) | `logo_vi_2dong` | `home_vi_iphone.png` |
| Trang chủ /en /pt /es /id /th | `logo_en` | `home_en_pc.png`, `home_en_iphone.png` |
| Màn tải (tiếng Việt, mọi màn hình) | `logo_vi_2dong` | `loading_vi_pc.png`, `loading_vi_iphone.png` |
| Màn tải trang ngoại ngữ | `logo_en` | `loading_en_pc.png` |
| Ảnh chia sẻ OG 1200×630 | `logo_vi_2dong` thay chữ tên game | `og_share_card_preview.jpg` — **CHỜ ANH DUYỆT** |

- `alt` của ảnh = tên game theo ngôn ngữ ("Đạo Hữu Xin Dừng Bước" / "BladeAscend"), nằm trong `<h1>`.
- Trang dùng WebP (logo 1 dòng 123 KB, 2 dòng 183 KB, EN 125 KB); trình duyệt chỉ tải ĐÚNG 1 bản.
  PNG gốc giữ trong `assets/src/brand/` (OG dựng từ PNG). Chặn vào trận 4.88 MB / 8 MB — ĐẠT.
- Ô chọn ngôn ngữ nay nằm hàng riêng phía trên logo ở mọi cỡ màn hình — hết đè lên tên game ở PC 1366.
- **Ảnh OG chưa phát hành** (anh dặn "chưa thay og:image cho tới khi tôi duyệt"). Duyệt xong:
  đặt `SHARE_CARD_APPROVED = True` trong `tools/make_brandart.py`, chạy `python3 tools/make_brandart.py`.
- Dựng lại logo: `python3 tools/make_logo.py` (khung `assets/src/brand/brand_logo_frame.png`,
  font `assets/src/fonts/Pattaya-Regular.ttf`). Font và khung không phát hành cho người chơi.

---

## Lưu trữ — ứng viên logo thư pháp 01/11 (không dùng)

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
