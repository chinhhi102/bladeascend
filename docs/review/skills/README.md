# Ảnh duyệt hiển thị kỹ năng — spec v6.36 §6.5

Chụp 02/11/2026 bằng `node tools/browsertest/skillshots.mjs` (server dev cổng 9998).
Ba trình duyệt cùng phòng: người xem (đứng phía trên), người dùng kỹ năng, người bị trúng
(đứng bên phải người dùng 3,5 U). Tên trong ảnh là tên thử của đội làm game.

| Nhân vật | Người đứng xem | Màn hình người bị trúng / người dùng |
|---|---|---|
| A Hớn — Bánh Bao Thần | `a_hon_xem.jpg` | `a_hon_dung.jpg` (tự thân) |
| Bà Tám — Loa Phường | `ba_tam_loa_xem.jpg` | `ba_tam_loa_trung.jpg` — bị đẩy ra ngoài khung → mũi tên đỏ mép màn hình |
| Ông Ba — Ru Ngủ / Trộm Kiếm | `ong_ba_vong_xem.jpg` | `ong_ba_vong_trung.jpg` — icon ngủ + giây dưới ảnh đại diện |
| Đại Ca — Gọi Đàn Em | `dai_ca_dep_xem.jpg` | `dai_ca_dep_dung.jpg` (tự thân) |
| Thằng Mập — Phun Trân Châu | `map_tra_sua_xem.jpg` | `map_tra_sua_trung.jpg` — trân châu dính kính |
| Cô Ba — Bày Hàng | `co_ba_ganh_xem.jpg` | `co_ba_ganh_trung.jpg` — sạp có bạt cam + thanh đếm |
| Sư Phụ Mì — Hút Mì | `su_phu_mi_xem.jpg` | `su_phu_mi_trung.jpg` |
| Bé Tí — Gồng Cơ | `be_ti_xem.jpg` | `be_ti_dung.jpg` (tự thân) |
| Shipper — Gom Đơn | `shipper_xem.jpg` | `shipper_dung.jpg` (tự thân) |
| Cụ Rùa — Rụt Cổ | `cu_rua_xem.jpg` | `cu_rua_dung.jpg` (tự thân, lúc đang rụt) |

Ảnh chụp đúng một khoảnh khắc (0,15–0,9 s sau khi bấm), nên hiệu ứng ngắn (sóng, sợi mì,
kiếm bay) có ảnh bắt được lúc đầu, có ảnh lúc gần tan. Muốn xem đủ thì chơi thử trên staging.
