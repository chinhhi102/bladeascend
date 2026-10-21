# Tiến độ — Kiếm Xoay Tu Tiên

_Cập nhật: 2026-10-21_

## Cách chạy (máy dev)
```bash
cd server && go run ./cmd/server          # http://localhost:8080 (đã build client) — đổi cổng: -addr :8090
cd client && npm run dev                  # http://localhost:5173 (tự cập nhật khi sửa code)
cd server && go test ./... && go vet ./...
cd client && npm run build
cd server && go run ./cmd/sim --bots 40 --minutes 10     # giả lập cân bằng
cd server && go run ./cmd/mapgen                         # tạo lại 3 bản đồ (ghi đè!)
cd server && go run ./cmd/artgen status                  # xem art nào còn thiếu
```
Phím test khi chơi (server chạy mặc định `-debug=true`): **E** +100 EXP · **R** lên cảnh giới · **K** +1 kiếm · **1–7** buff · **X** tự chết.
Giả lập ping: thêm `?lag=150` vào địa chỉ (ví dụ `http://localhost:5173/?lag=150`). Giới hạn FPS: `?fps=30`.

## Đã làm
| Milestone | Nội dung | Trạng thái |
|---|---|---|
| M0 | Repo theo mục 13.2, server Go + WebSocket + MessagePack, client Vite + PixiJS, `config/game.json` | ✅ |
| M1 | Vòng kiếm (capsule), quái 4 loại, item kiếm, EXP + 9 cảnh giới, scale 1→5, camera lùi theo scale, phím debug | ✅ |
| M2 | Chém thân → rớt tầng / chết, va kiếm đổi chiều + bể kiếm theo bảng 5.2 (có vết nứt), màn hình Tổng kết → Chơi tiếp | ✅ |
| M3 | Phòng + matchmaking (phòng đông nhất, hòa → cũ nhất), bot fill 9.3, bot AI (Farm/Hunt/Flee/Ambush/Wander, A*), trang chủ, leaderboard SQLite + `GET /api/leaderboard`, `cmd/sim` | ✅ |
| M4 | Đọc Tiled `.tmj`, 3 bản đồ tạo bằng code (`cmd/mapgen`), vật cản phá được + mọc lại 30s (cảnh báo 3s, không mọc đè người), bụi, bùn, gió (theo góc), hẻm 150/250/350, cổng cặp + mạng 4 cổng, bot né vào bụi/cổng. Hướng dẫn: `docs/MAP_GUIDE.md` | ✅ |
| M5 | 7 buff + làm mới thời gian, Rương Báu (báo trước 10s), nội suy 100ms, dự đoán + hiệu chỉnh vị trí mình, AOI, không gửi người ẩn thân/trong bụi, giả lập ping `?lag=` | ✅ |
| M6 | `artgen`: sinh ảnh minigen `/jobs` (gộp nhiều anim/1 ảnh), xóa nền, cắt sheet, căn pivot, ghép atlas + `manifest.json`; anim theo mood, VFX, 23 SFX + nhạc nền tự tổng hợp, UI đầy đủ. Danh sách art thiếu: `docs/ASSETS_STATUS.md` | ✅ (art còn thiếu, xem dưới) |
| M7 | Joystick ảo, giao diện điện thoại (dọc/ngang + gợi ý xoay), giới hạn FPS (PC 60 / ĐT 45), hướng dẫn VPS + Caddy HTTPS: `docs/DEPLOY.md` | ✅ (chưa deploy thật — cần bạn thuê VPS) |

Câu hỏi Q1–Q5 đã áp dụng theo `docs/ANSWERS.md` (thưởng hạ sát +30, gió theo góc 60°/120°, viên EXP thuộc người chém 1.5s, chết rơi 50% kiếm).

Test: 50 test Go (EXP, scale, bảng bể kiếm, rớt cảnh giới + rơi kiếm thừa, khiên, buff hết hạn/làm mới, Song Bội Đan, vật cản vỡ/mọc lại, hẻm hẹp, cổng không bất tử, ẩn thân/bụi không lộ trong gói tin, bot fill 5 người → 15 bot, người thứ 21 sang phòng mới, leaderboard không ghi bot, lọc tên tục…).
Hiệu năng: 40 bot, bản đồ thật → ~0.16 ms/tick (mục tiêu < 3 ms).

## Art (minigen)
- **2026-10-22: theo ART_PROMPTS mới** — style "bựa", mô tả nhân vật + gag riêng từng anim, **1 animation / ảnh** cho nhân vật (lưới đúng cột × hàng), nền **magenta #FF00FF** (artgen tự nhận nền magenta hay trắng và xóa sạch viền hồng). Nhân vật cũ (style "cute") giữ tạm tới khi ảnh mới về. Tổng 174 ảnh game, đã có 20. Câu hỏi Q6 trong `docs/QUESTIONS.md`.
- Tự động: `bash tools/art_watch.sh` — chờ tài khoản minigen sẵn sàng, mỗi đợt 30 ảnh, nghỉ 2.5 giờ giữa các đợt.
- Đã có: 9 skin kiếm, bong bóng kiếm, 3 viên EXP, 7 icon buff, 3 nhân vật (A Hớn, Bà Tám Loa, Ông Ba Võng) anim idle + run.
- Còn thiếu ~91 ảnh → game đang dùng placeholder (hình tròn + chữ / hình khối). Xem `docs/ASSETS_STATUS.md`.
- Quota minigen ~60–100 ảnh / 5 giờ → sinh theo đợt ưu tiên:
  ```bash
  cd server
  go run ./cmd/artgen run -p 1 -n 30 && go run ./cmd/artgen prep   # đợt 1: 7 nhân vật còn lại, quái, nền, vật cản, cổng
  go run ./cmd/artgen run -p 2 -n 30 && go run ./cmd/artgen prep   # đợt 2: anim bị chém/chết, VFX, huy hiệu, rương
  go run ./cmd/artgen run -p 4 -n 40 && go run ./cmd/artgen prep   # đợt 3: còn lại
  ```
- Nếu Gemini trả lời bằng chữ nhưng ảnh vẫn có trong lịch sử Gemini: tải ảnh về, đặt tên `assets/_raw/<tên job>.png` (tên job xem `artgen plan`), chạy `artgen prep`.

## Lỗi / hạn chế đã biết
- **Tài khoản minigen `pt1` đang báo UNAUTHENTICATED (cookie hết hạn)** → cần đăng nhập lại cookie bên minigen rồi chạy tiếp `artgen run`.
- Gemini đôi khi vẽ sai số frame (5 thay vì 4, 6 thay vì 8). `artgen prep` giữ đúng số frame thực tế vẽ ra; anim vẫn chạy được.
- Ảnh của Bà Tám Loa lấy từ bản JPEG 1024px (người dùng tải từ Gemini) → hơi mờ, có ngôi sao watermark nhỏ; nên sinh lại khi dư quota (`artgen run -only char_ba_tam_loa_core -redo`).
- Cổng 8080 trên máy dev hiện đang bị chương trình khác chiếm → chạy server với `-addr :8090` (khi dùng `npm run dev` thì chạy `KX_SERVER=http://localhost:8090 npm run dev`).
- Cân bằng (sim 10 phút, 40 bot): lên Trúc Cơ trung bình ~7 phút, nhiều bot chết sớm ở Luyện Khí 1 — nếu thấy chậm, chỉnh `config/game.json` (ví dụ tăng EXP quái, giảm `exp_base`). Chưa đổi gì vì đây là quyết định cân bằng của bạn.
- Dự đoán chuyển động phía client không mô phỏng đẩy lùi (knockback) → khi bị chém thấy giật nhẹ 1 nhịp.

## Việc tiếp theo gợi ý
1. Đăng nhập lại minigen → chạy các đợt art ở trên.
2. Chơi thử nhiều tab / điện thoại qua LAN, ghi lại cảm giác cân bằng.
3. Deploy theo `docs/DEPLOY.md`.
