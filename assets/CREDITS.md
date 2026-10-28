# Nguồn tài nguyên bên ngoài

Mọi thứ trong thư mục này đều dùng được cho mục đích thương mại. Thêm nguồn mới thì ghi vào
đây kèm giấy phép, đừng để lẫn lộn — lúc phát hành mà không biết ảnh nào ở đâu ra là rắc rối.

## Font

| Thứ | Nguồn | Giấy phép |
|---|---|---|
| `fonts/BeVietnamPro-700.ttf`, `-800.ttf` | [Be Vietnam Pro](https://fonts.google.com/specimen/Be+Vietnam+Pro) — Google Fonts | SIL Open Font License 1.1 |

Font tải về phục vụ tại chỗ thay vì nhúng từ Google: game chạy trong mạng LAN, máy người
chơi có thể không ra được Internet, mà tên nhân vật mất dấu thì còn tệ hơn là xấu.

## Hạt / VFX

Spec §3.6 đề xuất dùng bộ texture hạt CC0 cho tia lửa, khói, ánh sáng (ảnh AI sinh cho hạt
nhỏ thường kém ổn định). **Chưa tải về** — hiện mọi hạt đang vẽ bằng canvas trong code
(`render/ambient.ts`, `render/decals.ts`, `render/tienaura.ts`). Khi tải thì ghi vào đây:

| Bộ | Nguồn | Giấy phép |
|---|---|---|
| Kenney Particle Pack | https://kenney.nl/assets/particle-pack | CC0 |
| Brackeys VFX Bundle | https://brackeysgames.itch.io/brackeys-vfx-bundle | CC0 |

## Thư viện

| Thư viện | Giấy phép |
|---|---|
| [PixiJS](https://pixijs.com) | MIT |
| [pixi-filters](https://github.com/pixijs/filters) | MIT |
| [revolt-fx](https://github.com/revoltfx/revolt-fx) | MIT |
| [howler.js](https://howlerjs.com) | MIT |
| [@msgpack/msgpack](https://github.com/msgpack/msgpack-javascript) | ISC |

## Art do dự án tự sinh

Ảnh trong `characters/`, `monsters/`, `obstacles/`, `swords/`, `buffs/`, `beasts/`, `vfx/`,
`ui/`, `tiles/`, `decor/`, `zones/`, `items/` do công cụ `server/cmd/artgen` sinh bằng AI
theo prompt trong `docs/_raw/ART_PROMPTS.md`. Bản quyền thuộc dự án.
