# Nguồn & giấy phép tài nguyên

| Thư mục / file | Nguồn | Giấy phép |
|---|---|---|
| `characters/`, `monsters/`, `swords/`, `items/`, `buffs/`, `tiles/`, `zones/`, `obstacles/`, `gates/`, `vfx/`, `ui/`, `decor/` | Sinh bằng minigen (Gemini) theo prompt trong `docs/ART_PROMPTS.md`, hậu kỳ bằng `server/cmd/artgen` | Ảnh do AI tạo cho dự án; kiểm tra điều khoản dịch vụ của nhà cung cấp trước khi phát hành thương mại |
| `audio/sfx/*.ogg` | Tự tổng hợp bằng `tools/make_sfx.py` (ffmpeg) | CC0 — tự làm |
| `audio/bgm/main.ogg` | Tự tổng hợp bằng `tools/make_bgm.py` | CC0 — tự làm |

Khi thay bằng âm thanh tải từ kenney.nl / opengameart / freesound / pixabay: giữ đúng tên file và ghi thêm 1 dòng vào bảng này (nguồn + license, CC-BY cần ghi tên tác giả).
