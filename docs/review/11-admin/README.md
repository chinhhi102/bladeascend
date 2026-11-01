# C11 — trang quản trị React + antd trên STAGING (PC 1366×768 + iPhone WebKit)

Chụp 01/11 23:35 sau khi deploy bản `04fc86a`, đăng nhập bằng mật khẩu anh đã đưa lên
(công cụ đọc từ `docs/private/`, không in ra). Chụp lại: `node tools/browsertest/adminshots.mjs
<url> docs/review/11-admin <tệp-mật-khẩu>` (`PAGES=map` để chụp riêng một trang).

| Trang | Ảnh |
|---|---|
| Dashboard (Card/Statistic, tick theo khu đỏ khi > 4 ms) | `dashboard_*` |
| Phòng & Khu (cấu hình phòng + điều khiển bot) | `rooms_*` |
| Góp ý (Table + Drawer, ảnh đính kèm) | `feedback_*` |
| Bản đồ toàn cảnh (bộ vẽ thật của game, Xuất PNG) | `map_*` |
| Người chơi · Lỗi người chơi · Danh sách cấm | `players_*` `errors_*` `bans_*` |
| Nhân vật (bảng mục A5) · Bảng xếp hạng · Nhật ký | `chars_*` `board_*` `audit_*` |

Lỗi JS trên trang: 0 (cả hai trình duyệt).
