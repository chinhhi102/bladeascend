# Nguồn tài nguyên bên ngoài

Mọi thứ trong thư mục này đều dùng được cho mục đích thương mại. Thêm nguồn mới thì ghi vào
đây kèm giấy phép, đừng để lẫn lộn — lúc phát hành mà không biết ảnh nào ở đâu ra là rắc rối.

## Font

| Thứ | Nguồn | Giấy phép |
|---|---|---|
| `fonts/BeVietnamPro-700.ttf`, `-800.ttf` | [Be Vietnam Pro](https://fonts.google.com/specimen/Be+Vietnam+Pro) — Google Fonts | SIL Open Font License 1.1 |
| `fonts/NotoSansThai.ttf` (biến thiên độ đậm, tải 01/11/2026) | [Noto Sans Thai](https://fonts.google.com/noto/specimen/Noto+Sans+Thai) — github.com/google/fonts `ofl/notosansthai` | SIL Open Font License 1.1 |

Font tải về phục vụ tại chỗ thay vì nhúng từ Google: game chạy trong mạng LAN, máy người
chơi có thể không ra được Internet, mà tên nhân vật mất dấu thì còn tệ hơn là xấu.

## Hạt / VFX

| Bộ | Nguồn | Giấy phép |
|---|---|---|
| **Kenney Particle Pack 1.1** — Kenney Vleugels | https://kenney.nl/assets/particle-pack | **CC0 1.0** |
| Brackeys VFX Bundle | https://brackeysgames.itch.io/brackeys-vfx-bundle | CC0 — *chưa tải* |

File dùng từ Kenney Particle Pack (bản nền trong suốt), đặt trong `assets/vfx/particles/`:

| Tên trong dự án | File gốc | Dùng ở đâu |
|---|---|---|
| `spark.png` | `spark_04.png` | vệt lửa thần thú Kim Tiên |
| `smoke.png` | `smoke_01.png` | bóng mây, sương mù, vệt mây thần thú |
| `light.png` | `light_01.png` | quầng sáng hào quang Tiên, đom đóm, vệt sáng Ngọc long |
| `star.png` | `star_08.png` | lông vũ phượng |
| `dot.png` | `circle_05.png` | hạt tròn dùng chung |
| `magic.png`, `flare.png`, `trace.png` | cùng tên | để dành |

Giấy phép CC0 cho phép dùng thương mại, **không bắt buộc ghi công** — vẫn ghi ở đây vì đó là
phép lịch sự tối thiểu với người cho không.

Thả thêm hạt mới vào `assets/vfx/particles/<tên>.png` là code tự dùng, không phải sửa gì
(`render/particles.ts`).

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
