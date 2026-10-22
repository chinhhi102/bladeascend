# Tiến độ — Kiếm Xoay Tu Tiên

_Cập nhật: 2026-10-22_

## Cách chạy (máy dev)
```bash
cd server && go run ./cmd/server          # http://localhost:8080 (đã build client) — đổi cổng: -addr :8090
cd client && npm run dev                  # http://localhost:5173 (tự cập nhật khi sửa code)
cd server && go test ./... && go vet ./...
cd client && npm run build
cd server && go run ./cmd/sim --bots 40 --minutes 10     # giả lập cân bằng
cd server && go run ./cmd/mapgen                         # tạo lại 3 bản đồ (ghi đè!)
cd client && npm run assets:report                       # asset nào còn thiếu (thêm --all để xem cả cái đã có)
cd server && go run ./cmd/artgen status                  # tình trạng theo từng ảnh cần sinh
```
Phím test khi chơi (server chạy mặc định `-debug=true`): **E** +100 EXP · **R** lên cảnh giới · **K** +10 kiếm · **1–7** buff · **X** tự chết · **H** bật/tắt hitbox kiếm của server.
Đo nhịp farm: `cd server && go run ./cmd/sim --bots 20 --minutes 3 --farm` (không ai bị chém; thêm `--orb-cell N`, `--monsters N` để thử).
Giả lập ping: thêm `?lag=150` vào địa chỉ (ví dụ `http://localhost:5173/?lag=150`). Giới hạn FPS: `?fps=30`.

## Ưu tiên hiện tại (2026-10-22)
**Gameplay trước, art sau.** Game chơi đầy đủ khi `assets/` trống (đã test: chạy server với thư mục asset rỗng → vào chơi, lên cảnh giới, buff, chết → Tổng kết → Chơi tiếp, không lỗi JS).
Placeholder khi thiếu ảnh: nhân vật = hình tròn màu + chữ viết tắt (A Hớn → `AH`) + tên anim; vật cản phá được = hình chữ nhật đổi màu theo HP (xanh → vàng → đỏ, đống đổ nát màu xám); tường không phá = xám; kiếm = thanh dài màu theo cảnh giới; quái = hình tròn + số HP; vật phẩm = bong bóng màu.
Danh mục asset theo đúng tên ART_PROMPTS: `client/src/render/catalog.ts` (dùng chung cho loader và `assets:report`).

## Spec v3 / v3.1 / v3.2 (2026-10-22) — ✅ xong
- **Vòng kiếm**: 20 kiếm/vòng, vòng lồng nhau so le `k·18°/R`, lưỡi nghiêng 35° ngả về sau và lật trong 0.1s khi đổi chiều, bán kính chỉ theo vòng + scale (không theo số kiếm), vòng sau lùi xa hơn; kiếm tối đa 60→400 theo `max_swords`; quay 1 vòng/300ms; va chạm theo **góc quét** giữa 2 tick (vòng dày coi cả vành khăn là vùng chém). Camera lùi theo `2.6 × reach`.
- **Vẽ (client)**: `ParticleContainer`, mỗi vòng một màu `ring_tints` (vòng 1 giữ màu skin cảnh giới), vòng ngoài vẽ dưới, vệt mờ (vành khăn mờ + 2 bóng kiếm kéo sau). Thân kiếm vòng 2+ là placeholder trắng/xám cho tới khi sinh lại skin. Phím **H** vẽ capsule hitbox server chồng lên để so.
- **Chiến đấu**: bị chém rớt nguyên 1 cảnh giới, giữ tầng và % tiến độ; Luyện Khí bị chém là chết; kiếm thừa / kiếm rơi khi chết thành **Bó Kiếm** (≤10/bó); bể kiếm ×5 (`clash_unit`) + hai người bị đẩy văng 150 trong 0.2s.
- **EXP & vật phẩm**: viên EXP rải khắp bản đồ (`exp_orbs`), nhặt xong 5s mọc lại trong cùng ô 1000×1000; 400 vật phẩm/phòng mọc lại sau 3s; item Kiếm +3, Bó Kiếm +15; Rương mỗi 60s nổ 8 món; hút đồ 90×scale. Lưới không gian cho item (nhặt + AOI nhanh).
- **Nhịp**: anh chốt Luyện Khí → Trúc Cơ 30–60s → `exp_orbs.cell = 160` (sim đo 40–56s).
- **Vật cản**: chặn thân bằng **chân đế** (90% rộng × 30% cao ở đáy, Tiled đặt tay được bằng thuộc tính `footprint = "w,h"`), kiếm chém theo thân hình; y-sort với nhân vật; vật cản che mình thì mờ 50%. Nền + decor qua `ColorMatrixFilter` (giảm bão hòa/sáng/tương phản); viên EXP/item/quái có bóng đổ + quầng sáng.
- Test mới: rớt cảnh giới giữ tầng, Luyện Khí tầng bất kỳ chết, bó kiếm, bảng bể kiếm ×5, đẩy văng, góc so le R=1/2/3, vòng thiếu kiếm rải đều khe, reach tăng theo vòng, bán kính không đổi theo số kiếm (v3.2), tilt đổi dấu, va chạm góc quét, 1 vòng/300ms, chân đế, viên EXP mọc lại đúng vùng, vật phẩm mọc lại sau 3s. Tổng 64 test Go.
- **Còn lại / lỗi đã biết**:
  - Hàng rào/tường trong 3 bản đồ cũ được vẽ cao hơn hình trong Tiled (kiểu 3/4), nên vùng kiếm chém (theo hình Tiled) hơi thấp hơn phần nhìn thấy. Cần chỉnh bản đồ (`cmd/mapgen`) cho hình Tiled = hình vẽ.
  - Art cần sinh lại: 9 skin kiếm (lưỡi trắng/xám) và 3 nền đất (nhạt) — đã đánh dấu "Sinh lại" trong `docs/ART_JOBS.md`. Bó Kiếm chưa có tên art trong ART_PROMPTS → đang vẽ placeholder.
  - Q8 (`docs/QUESTIONS.md`): checklist §17 còn số kiếm mất theo bản cũ.

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

Bổ sung gameplay 2026-10-22: đột phá cảnh giới → người xung quanh hiện emote kinh ngạc (§3.2); bị chém lộ khỏi bụi → emote "Lộ rồi!" (§12); vòng kiếm có vết nứt khi `crack > 0` (§5.2).

Câu hỏi Q1–Q5 đã áp dụng theo `docs/ANSWERS.md` (thưởng hạ sát +30, gió theo góc 60°/120°, viên EXP thuộc người chém 1.5s, chết rơi 50% kiếm).

Test: 50 test Go (EXP, scale, bảng bể kiếm, rớt cảnh giới + rơi kiếm thừa, khiên, buff hết hạn/làm mới, Song Bội Đan, vật cản vỡ/mọc lại, hẻm hẹp, cổng không bất tử, ẩn thân/bụi không lộ trong gói tin, bot fill 5 người → 15 bot, người thứ 21 sang phòng mới, leaderboard không ghi bot, lọc tên tục…).
Hiệu năng: 40 bot, bản đồ thật → ~0.16 ms/tick (mục tiêu < 3 ms).

## Art (minigen)
- **2026-10-22 (phiên sau): 90/211 asset đủ** (artgen: 66 job xong, 130 còn thiếu). Mới có: khung thoại, rương, hit/death Ông Ba Vòng, death Đại Ca, hit Thằng Mập; run Cô Ba Gánh cắt lại được (không cần sinh lại).
  - `artgen run` giờ gửi **mỗi prompt đúng 1 lần**, tuần tự (`-c 1`), nghỉ 30s giữa prompt (`-gap`); hết quota thì chờ tới `reset_at` rồi gửi lại (lúc đó Gemini chưa vẽ gì).
  - `artgen prep`: sheet nhiều hàng mà cắt theo hàng ra sai (hàng dính nhau, ma bay tách khỏi xác) → tự cắt theo lưới cột × hàng, mỗi mảnh vào ô chứa tâm của nó.
  - **Lỗi đã biết (minigen):** sau khi sửa lỗi `metadata`, minigen dùng **chung 1 cuộc chat Gemini** cho mọi job (`chat_id` giống nhau) → Gemini lặp lại hình của prompt trước (5/11 ảnh sai: 4 ảnh vẽ rương, 1 ảnh chép sheet chết Ông Ba). Đã xóa ảnh sai. Cần minigen mở chat mới cho mỗi job trước khi sinh tiếp.
- **2026-10-22: theo ART_PROMPTS mới** — style "bựa", mô tả nhân vật + gag riêng từng anim, **1 animation / ảnh** cho nhân vật (lưới đúng cột × hàng), nền **magenta #FF00FF** (artgen tự nhận nền magenta hay trắng và xóa sạch viền hồng). Nhân vật cũ (style "cute") giữ tạm tới khi ảnh mới về. Tổng 174 ảnh game, đã có 20. Câu hỏi Q6 trong `docs/QUESTIONS.md`.
- Q6 (2026-10-22): nhân vật 1 anim/ảnh; quái/cổng/rương gộp các anim của cùng 1 đối tượng, mỗi anim 1 hàng; VFX mỗi hiệu ứng 1 ảnh; vật cản `break`/`regrow` tách riêng; ảnh tĩnh gộp thoải mái. Bộ cắt frame giờ tìm từng hình theo vùng pixel liền nhau (sheet lệch lưới vẫn dùng được). Tổng 186 ảnh game.
- **Danh sách prompt để sinh bằng công cụ khác: `docs/ART_JOBS.md`** (tạo lại: `go run ./cmd/artgen prompts`). Lưu ảnh thành `assets/_raw/<tên job>.png` rồi chạy `go run ./cmd/artgen prep`.
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
- **Tốc độ chạy (22/10/2026)**: `move_base` 260 → **130** và bỏ hệ số `sqrt(scale)` → tốc độ không còn tăng theo tu vi; chỉ buff Thần Hành (nhặt vật phẩm) ×1.4 và địa hình mới đổi tốc độ. Hệ quả cần chơi thử: camera vẫn lùi ra khi lên cảnh giới, nên người tu vi cao nhìn trên màn hình đi chậm hơn nhiều (≈130 px/s ở scale 1 so với ≈33 px/s ở scale 5). Nếu thấy người cấp cao "ì" quá thì giảm độ lùi camera (`camera_exp`) hoặc tăng `move_base`.
- **Tốc độ xoay kiếm (22/10/2026)**: `rev_secs` 0.3 → **0.6** (1 vòng / 600ms, chậm một nửa). Mỗi tick kiếm quét 20° thay vì 40°; một vòng 20 kiếm vẫn quét kín 400° nên vành khăn còn kín, nhưng vòng dưới 18 kiếm giờ có kẽ hở lách qua được.
- **Số kiếm tối thiểu (22/10/2026)**: `min_count` 1 → **0**. Người chơi có thể bị tước sạch kiếm, khi đó không chạm được ai và phải đi nhặt kiếm. Lưu ý: điều này **không** làm người chơi chết — luật chết vẫn là `hit_drops_realm`, chỉ Luyện Khí bị chém mới chết (xem Q9 trong QUESTIONS.md).
- Môi trường dev: `~/go/pkg/mod` có một số thư mục thuộc quyền `root` (dấu vết của lần chạy `sudo go` trước) → `go build/test` báo "permission denied". Đang né bằng `GOMODCACHE=~/.cache/kx-gomod`. Sửa tận gốc: `sudo chown -R chinhdc:chinhdc ~/go/pkg/mod`. Ngoài ra `go` không còn trong PATH, đang dùng bản trong cache: `~/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.11.linux-amd64/bin/go`.

## Việc tiếp theo gợi ý
1. Đăng nhập lại minigen → chạy các đợt art ở trên.
2. Chơi thử nhiều tab / điện thoại qua LAN, ghi lại cảm giác cân bằng.
3. Deploy theo `docs/DEPLOY.md`.
