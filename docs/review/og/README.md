# Mục 9 — Ảnh chia sẻ OG mới (CHỜ DUYỆT)

- `share_card.jpg` — 1200×630, ~182 KB (< 300 KB). Dựng bằng `python3 tools/make_brandart.py`.
- `xem_nhu_zalo_400px.jpg` — cùng ảnh thu còn 400 px như Zalo/Messenger hiển thị.

**og:image chưa đổi** — vẫn trỏ ảnh cũ `client/public/brand/share_card.jpg`.
Duyệt xong: mở `tools/make_brandart.py`, đổi `SHARE_CARD_APPROVED = False` → `True`,
chạy `python3 tools/make_brandart.py`, rồi deploy. (Hoặc nhắn Claude "duyệt OG".)
