# Nguồn tài nguyên bên ngoài

Mọi thứ trong thư mục này đều dùng được cho mục đích thương mại. Thêm nguồn mới thì ghi vào
đây kèm giấy phép, đừng để lẫn lộn — lúc phát hành mà không biết ảnh nào ở đâu ra là rắc rối.

## Font

| Thứ | Nguồn | Giấy phép |
|---|---|---|
| `fonts/BeVietnamPro-700.ttf`, `-800.ttf` | [Be Vietnam Pro](https://fonts.google.com/specimen/Be+Vietnam+Pro) — Google Fonts | SIL Open Font License 1.1 |
| `fonts/NotoSansThai.ttf` (biến thiên độ đậm, tải 01/11/2026) | [Noto Sans Thai](https://fonts.google.com/noto/specimen/Noto+Sans+Thai) — github.com/google/fonts `ofl/notosansthai` | SIL Open Font License 1.1 |
| Pattaya — dùng trong thiết kế logo (tệp nguồn logo `docs/_raw/logo.zip`; chưa có tệp font trong `assets/src/fonts/`) | [Pattaya](https://fonts.google.com/specimen/Pattaya) — Google Fonts | SIL Open Font License 1.1 |

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

Giấy phép đọc từ `license` trong `package.json` của từng gói (`client/node_modules`,
`tools/asset-pipeline/node_modules`) và từ tệp `LICENSE` trong bộ nhớ đệm module Go
(`go env GOMODCACHE`), ngày 02/11/2026, đúng phiên bản đang khoá trong repo. Thêm hay nâng
thư viện thì cập nhật bảng này.

### Client (`client/package.json`) — đóng gói vào game

| Thư viện | Phiên bản | Giấy phép |
|---|---|---|
| [PixiJS](https://pixijs.com) (`pixi.js`) | 8.22.0 | MIT |
| [pixi-filters](https://github.com/pixijs/filters) | 6.1.5 | MIT |
| [revolt-fx](https://github.com/bma73/revolt-fx) | 1.3.6 | MIT |
| [howler.js](https://howlerjs.com) (`howler`) | 2.2.4 | MIT |
| [@msgpack/msgpack](https://github.com/msgpack/msgpack-javascript) | 3.1.3 | ISC |
| [React](https://react.dev) (`react`, `react-dom`) — trang quản trị | 18.3.1 | MIT |
| [Ant Design](https://ant.design) (`antd`) — trang quản trị | 5.29.3 | MIT |

### Client — chỉ dùng lúc build (không đóng gói)

| Thư viện | Phiên bản | Giấy phép |
|---|---|---|
| [Vite](https://vite.dev) | 7.3.7 | MIT |
| [TypeScript](https://www.typescriptlang.org) | 5.9.3 | Apache-2.0 |
| `@types/howler`, `@types/node`, `@types/react`, `@types/react-dom` (DefinitelyTyped) | 2.2.13 · 22.20.5 · 18.3.12 · 18.3.1 | MIT |

### Công cụ (`tools/`)

| Thư viện | Dùng ở | Phiên bản | Giấy phép |
|---|---|---|---|
| [sharp](https://sharp.pixelplumbing.com) (kèm libvips nhị phân, LGPL-3.0) | `tools/asset-pipeline` | 0.35.5 | Apache-2.0 |
| [Playwright](https://playwright.dev) | `tools/browsertest` (đo hiệu năng, chụp ảnh) | ^1.49.0 | Apache-2.0 — *cần xác minh* (chưa cài trên máy soạn) |

### Server (`server/go.mod`)

| Module | Phiên bản | Giấy phép |
|---|---|---|
| [github.com/coder/websocket](https://github.com/coder/websocket) | v1.8.15 | ISC |
| [github.com/vmihailenco/msgpack/v5](https://github.com/vmihailenco/msgpack) | v5.4.1 | BSD-2-Clause |
| [github.com/vmihailenco/tagparser/v2](https://github.com/vmihailenco/tagparser) | v2.0.0 | BSD-2-Clause |
| [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) | v1.61.0 | BSD-3-Clause |
| modernc.org/libc | v1.79.0 | BSD-3-Clause (kèm `LICENSE-3RD-PARTY.md` cho mã musl… bên trong) |
| modernc.org/mathutil | v1.7.1 | BSD-3-Clause |
| modernc.org/memory | v1.12.1 | BSD-3-Clause |
| [github.com/google/uuid](https://github.com/google/uuid) | v1.6.0 | BSD-3-Clause |
| [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) | v0.58.0 | BSD-3-Clause |
| [golang.org/x/sys](https://pkg.go.dev/golang.org/x/sys) | v0.49.0 | BSD-3-Clause |
| github.com/remyoudompheng/bigfft | v0.0.0-20230129092748 | BSD-3-Clause |
| [github.com/dustin/go-humanize](https://github.com/dustin/go-humanize) | v1.1.0 | MIT |
| [github.com/mattn/go-isatty](https://github.com/mattn/go-isatty) | v0.0.24 | MIT |
| [github.com/ncruces/go-strftime](https://github.com/ncruces/go-strftime) | v1.1.0 | MIT |

Không có thư viện nào giấy phép "copyleft" bắt buộc công khai mã của game (GPL/AGPL) — nên
chọn giấy phép nào cho mã nguồn dự án cũng không vướng (xem `docs/LICENSE_OPTIONS.md`).

## Art do dự án tự sinh

Ảnh trong `characters/`, `monsters/`, `obstacles/`, `swords/`, `buffs/`, `beasts/`, `vfx/`,
`ui/`, `tiles/`, `decor/`, `zones/`, `items/` do công cụ `server/cmd/artgen` sinh bằng AI
theo prompt trong `docs/_raw/ART_PROMPTS.md`. Bản quyền thuộc dự án.
