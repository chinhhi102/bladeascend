# D14–D16 — trang chủ theo ngôn ngữ (PC 1366×768 + iPhone WebKit)

Chụp 01/11 trên bản build mới nhất (server dev :9998).

| Ảnh | Xem gì |
|---|---|
| `home_en_*` | `/en/` — chữ English, tên nhân vật giữ tiếng Việt, biệt danh riêng |
| `home_en_*_full` | cả trang `/en/`: giới thiệu, 3 ảnh, cách chơi, FAQ, chân trang bằng English |
| `home_th_*` | `/th/` — Noto Sans Thai chỉ tải ở trang này |
| `home_pt-BR_*` | `/?lang=pt-BR` — thẻ "Recarga 30s · 2 cargas" (dài nhất) vẫn vừa thẻ |
| `home_vi_*` | trang gốc sau khi dời ô chọn ngôn ngữ trên điện thoại |

Lỗi tự tìm ra và đã sửa trong lượt này:
- iPhone: ô "อัตโนมัติ" đè lên tên game → trên màn ≤ 600 px ô chọn nằm thành một hàng riêng.
- `/th/`, `/en/`…: màn hình tải còn chữ Việt ~5 s trên WebKit trước khi JS nạp từ điển →
  `seo-pages` thay luôn tên game, khẩu hiệu, dòng "Đang tải" theo ngôn ngữ.

Mọi bản dịch là **bản nháp** (`"_draft": true`) — cần người bản ngữ duyệt: `client/src/i18n/NOTES.md`.
