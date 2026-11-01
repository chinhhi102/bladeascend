# Tiến độ — BladeAscend (Đạo Hữu Xin Dừng Bước)

_Cập nhật: 2026-10-22_

## Cách chạy (máy dev)

> Server nghe `:9999` trên **mọi card mạng**, nên máy khác cùng Wi-Fi vào được bằng
> `http://<IP-máy-này>:9999`. Muốn đóng lại, chỉ cho máy này chơi:
> `go run ./cmd/server -addr 127.0.0.1:9999`.
```bash
cd server && go run ./cmd/server          # http://localhost:9999 (đã build client) — đổi cổng: -addr :8090
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
  - (Q8 đã xong: spec gốc v3.5 đặt `clash_unit` = 1 và sửa checklist §17 cho khớp.)

## Đã làm

- **Phiên 1/11/2026 (tối) — thư viện art, EXP_AUDIT, admin staging**
  - **Thư viện art** (luật 13, spec §13.7-3b). `assets/library/` + `INDEX.md`.
    - Cứu **205/205** ảnh từng bị xoá trong git, không tệp nào mất. Chuyển thêm 23 tệp Q31
      nhóm 2 (tổng 228).
    - Pipeline và mọi `check:*` không đọc thư mục này. Dist dựng lại **giống từng byte**
      (373 tệp, boot 2,77 MB).
    - `artgen` coi art trong thư viện là đã có. Trước đó nó định mua lại boba + 9 bộ regrow.
    - `check:dead` đề xuất chuyển vào thư viện, không đề xuất xoá.
  - **EXP_AUDIT (v6.21 bước 1)**: `docs/EXP_AUDIT.md`, `cmd/sim --audit`, ghi nguồn EXP
    (`Game.ExpTrace`).
    - Nguyên nhân thật: nhặt Bí Kíp khi đủ lượt **nhân đôi** tổng EXP (`items.go`), chiếm
      80–97 % EXP.
    - **Chờ chủ dự án duyệt Q33 rồi mới sửa** (gồm cả §6.0, bảng nhịp, `--farm` CI).
  - Bảng xếp hạng: điểm lấy từ cùng bản chụp đã giữ (13d3457). artgen prep `-only` chỉ đụng
    đúng tệp (b115cc2). Cả hai đã deploy staging + :9999.
  - Staging đọc `ADMIN_PASSWORD_HASH` từ `~/bladeascend-secrets/admin.env` (ngoài thư mục
    deploy). Chủ dự án tự đặt, hướng dẫn ở `docs/SETUP_ADMIN.md`.
  - **Còn lại theo thứ tự:** v6.21 bước 2–5 (sau Q33) → v6.22 (bot như người mới, bật/tắt bot,
    loadtest trước khi mở phòng 50) → 3b → 3c → 3d → 3e.

- **Thiên Môn (29/10/2026, spec §8.4a, CHANGES_v5 mục 16–18)** — lối đi **chắc chắn** tới
  Thiên Đình, ngược hẳn với 4 cổng thường vốn cố tình giấu đích đến.
  - Mỗi khu trừ Thiên Đình có **1 Thiên Môn ở tâm khu**, mapgen chừa quảng trường trống bán
    kính 500 quanh nó. Thiên Đình có **4 Bến Thiên Môn** cách tâm 1 400 ở đầu 4 đường chính.
  - **Một chiều, đích cố định**: mỗi khu nối cố định với một bến (Linh Sơn → Bắc, Trúc Lâm →
    Đông, Phàm Trấn → Nam, Băng Nguyệt → Tây). Cố định chứ không ngẫu nhiên là có lý do:
    người chơi quen đường sẽ biết ai vừa từ khu nào lên, và ở khu trung tâm đó là thông tin
    đáng giá.
  - Thiên Môn và bến **không nằm trong bể ngẫu nhiên** của cổng thường; **đứng trên bến
    không bị hút đi tiếp** — thiếu luật này thì người vừa lên Thiên Đình bị bắn ngược về khu
    khác ngay lập tức. Cả ba điều có test riêng.
  - Minimap: icon ⛩ vàng **luôn hiện**, cộng mũi tên ở góc chỉ hướng khi còn cách trên 1 200.
    Kill feed khi Top 3 bước qua. Bot chỉ dùng Thiên Môn khi đã Đại Thừa trở lên — dưới cảnh
    giới đó Thiên Môn chỉ ném chúng vào giữa ổ kiến lửa.

- **Kỹ năng nhân vật qua Bí Kíp (29/10/2026, spec §6.3, CHANGES_v5 mục 11–14)** — đủ 10 kỹ
  năng, vật phẩm Bí Kíp (`skill_scroll`, weight 10), khung trạng thái chung, nút bấm, lớp
  trân châu, bot biết dùng, và `cmd/sim --duel`.
  - **Khung trạng thái chung** (`internal/game/status.go`): chậm · trói · bịt mắt · trượt ·
    gồng · xuyên vật cản · mai rùa · lướt. Mười kỹ năng dùng chung một bảng thời hạn, nên
    thêm kỹ năng sau chỉ là thêm một hằng số. Trạng thái **không cộng dồn** — trúng Ru Ngủ
    hai lần vẫn chỉ 3 giây, nếu không hai người cùng kỹ năng khoá cứng một người vô hạn.
  - Phím **Space** / nút tròn trên điện thoại, hiện icon + số lượt, xám khi hết. Chết là mất
    hết lượt; rớt cảnh giới hay qua cổng thì không.
  - **Trân châu dính màn hình**: lớp Pixi riêng, nằm dưới HUD, dồn về giữa, trượt xuống để
    lại vệt. Bot trúng thì **tầm nhìn còn 40%** (§6.3) — không có luật này thì kỹ năng đó
    hoàn toàn vô dụng với bot.
  - **`cmd/sim --duel`** đấu chéo 10×10, in bảng tỉ lệ thắng + **phân tích khắc chế** (ai
    khắc ai) + **số lần thật sự bấm được kỹ năng**.

  **Ba lỗi bài đấu tìm ra, đều là lỗi thật:**
  1. **Đàn em mài kiếm chính đại ca** — không có phép kiểm "cùng phe", nên Gọi Đàn Em là kỹ
     năng tự sát. Đại Ca Dép từ **1,4% → 98,6%** tỉ lệ thắng sau khi sửa.
  2. **Thiếu hẳn luật "bot bị bịt mắt còn 40% tầm nhìn"** của §6.3; `tran_chau_bot_vision`
     có trong config mà không chỗ nào đọc.
  3. **Ba nhân vật chưa từng bấm kỹ năng lần nào** (0/144 lượt): luật bot đòi "đang bị dí",
     mà bot trong bài đấu thì không sợ ai. Nhìn bảng tỉ lệ thắng thì "kỹ năng yếu" và "bot
     không bao giờ bấm" giống hệt nhau — phải đếm số lần bấm mới phân biệt được.

  **Chưa cân bằng xong, và đó là chủ ý**: chủ dự án chốt 29/10 là cứ thả ra cho người chơi
  phản hồi rồi chỉnh sau. Hiện `be_ti` và `dai_ca_dep` chưa ai khắc được; `ba_tam_loa`,
  `co_ba_ganh`, `map_tra_sua`, `shipper` chưa khắc được ai. Chạy lại bảng bất cứ lúc nào:
  `cd server && go run ./cmd/sim --duel`.

- **Thế giới 5 khu xong phần còn lại (29/10/2026, CHANGES_v5 mục 4 goroutine, 7, 8, 9, 10)**
  — và lần đầu tiên **cả 5 ngân sách §13.6 đều ĐẠT**:

  | Chỉ số | Đo được | Ngân sách |
  |---|---|---|
  | Tick server (TB) | **1,2 ms** | < 4 ms ✅ |
  | Tick server (p95) | **1,69 ms** | < 4 ms ✅ |
  | Snapshot lớn nhất | **4,2 KB** | < 8 KB ✅ |
  | Băng thông / client | **18,7 KB/s** | < 40 KB/s ✅ |
  | Ping p95 | 0,00 ms | < 20 ms ✅ |

  Kịch bản là của spec v5: 20 người + bot rải 5 khu, 10 Đạo Tổ. `cmd/loadtest` giờ in thêm
  **tick từng khu** (0,24–0,86 ms) để khi một khu chậm hẳn thì thấy ngay là khu nào.

  **Mỗi khu một goroutine.** An toàn vì trong một tick không khu nào đọc dữ liệu khu khác.
  Để đạt được điều đó phải dọn **bốn chỗ dùng chung** mà `-race` chỉ ra:
  1. `rng` của phòng → mỗi khu một bộ sinh số riêng.
  2. Hàng chờ chuyển khu → hàng riêng mỗi khu, gom lại sau khi mọi khu xong.
  3. Mốc gửi lại tên người chơi (`namesAt/namesDue`) → chuyển vào Zone.
  4. Kill feed liên khu → gom vào `Zone.feeds`, gửi sau `wg.Wait()`; gửi thẳng sang client
     của khu khác là ghi vào kết nối người ta đúng lúc khu kia đang ghi.
  5. `handleDeaths` gọi xếp hạng toàn thế giới giữa lúc khu khác đang sửa dữ liệu → chụp
     hạng **một lần đầu mỗi tick**. Hạng trên bảng Tổng kết chậm một tick thì không ai nhận
     ra; đua dữ liệu thì có.

- **5 bản đồ cho 5 khu (29/10/2026, spec §8.0, CHANGES_v5 mục 6)** — `cd server && go run
  ./cmd/mapgen` sinh cả 5; xem riêng từng khu bằng `-zone X -seed N -png ra.png`.
  Ảnh duyệt: `docs/perf/zone0.png` … `zone4.png`.

  | # | Khu | Thuật toán | Vật cản |
  |---|---|---|---|
  | 0 | Thiên Đình | giữ nguyên §8.1 | 296 |
  | 1 | Linh Sơn Cốc | 5 dãy vách đá **so le** (dãy lẻ chạm mép trái, dãy chẵn mép phải) chia hẻm núi, mỗi dãy chừa 1–2 đèo rộng 200–300 có cầu treo; 2 cao nguyên quanh đài, mỗi cao nguyên 2 lối lên | 158 |
  | 2 | Trúc Lâm Mê Cung | mê cung **đào đường lùi dần** trên lưới 11×11 — bảo đảm mọi ô đi tới được nhau **theo cách nó dựng**, không phải rải tường rồi cầu trời; bụi rậm gấp 3; vài ao | 322 |
  | 3 | Phàm Trấn | lưới phố **cố tình lệch** (khoảng cách 1 500–2 100, mỗi phố xê dịch ±140) + một đại lộ chéo; mỗi ô phố nhồi 3–5 công trình, thỉnh thoảng bỏ trống một ô làm sân | 109 |
  | 4 | Băng Nguyệt Hồ | hồ băng đa giác ở giữa + 4 đảo tuyết, bờ tuyết phủ quanh (`snow_mult` 0,8) | 29 |

  Mỗi khu **4 cổng** thuộc mạng chung cả thế giới và **2 Đài Tiên Khí**; Thiên Đình giữ 4 đài
  nhỏ + Thiên Tiên Đài. Khu 0 trước đây đặt 8 cổng, nay đúng 4 theo §8.4.

  **Ba lỗi bộ kiểm bắt được trong lúc làm** (đều là lỗi thật, không phải chỉnh số cho qua):
  1. `ringWall` nhận `gapArc` là **độ dài cung tính bằng đơn vị**, tôi truyền 0.5 tưởng là
     radian → cao nguyên bị bịt kín, 3,3% ô trống bị nhốt.
  2. `addRoad` vẽ vùng đường bằng **hộp bao** của đoạn, nên một con phố chéo biến gần hết
     Phàm Trấn thành mặt đường. Đã sửa: đoạn chéo cắt thành nhiều khúc ngắn ôm sát đường.
  3. `house()` nhận **góc trên-trái** chứ không phải tâm; truyền tâm vào thì nhà tràn ra
     đường và bị từ chối — cả thị trấn chỉ còn 50 vật cản.

  Ba bản đồ cũ (`lang_tre`, `cho_dem`, `nui_da`) đã xoá, 5 khu thay thế.

- **Thế giới 5 khu — xong phần server (29/10/2026, spec v5 §8.0 §8.4, CHANGES_v5 mục 4+5)**:
  một phòng giờ là **5 khu**, mỗi khu một bản đồ 12 000² riêng, mô phỏng riêng, danh sách
  người riêng, AOI riêng (`internal/room/zone.go`). Số khu lấy từ `room.zones` trong config;
  để 1 là quay về thế giới một bản đồ như cũ.
  - Người mới vào **khu ít người thật nhất**; bot sinh và hồi sinh vào **khu ít nhân vật
    nhất**; trần người + bot vẫn tính cho cả phòng (20), không nhân năm.
  - **Bảng xếp hạng, danh hiệu, Uy Danh tính chung cả thế giới**; riêng **dấu trên minimap**
    (Đài Tiên Khí, rương, ngôi sao Top 1) chỉ vẽ của khu đang đứng — vẽ dấu khu khác lên thì
    người chơi đuổi theo một chấm không bao giờ tới được.
  - **Cổng nối cả thế giới**: 20 cổng chung một mạng, bước vào một cổng ra ở một cổng bất kỳ
    khác, không bao giờ ra lại cổng vừa vào. Khu không biết gì về khu khác — phòng làm người
    chỉ đường qua `GateRouter`/`MoveToZone`. Chuyển người làm ở **ranh giới tick**, lúc không
    khu nào đang chạy, nên không cần khoá.
  - Client nhận `welcome` mới giữa trận = vừa sang khu: tải lại bản đồ, **không** phát lại
    âm thanh hồi sinh hay hướng dẫn, hiện chữ lớn tên khu 1,5s và tên khu cạnh minimap.
  - Thử thật: đi vào cổng ở Thiên Đình → sang **Trúc Lâm Mê Cung**, giữ nguyên kiếm và Uy
    Danh. Test: 6 bài trong `internal/room/zone_test.go`, trong đó bài cổng bắn 2 000 lần
    kiểm đủ 5 khu đều tới được và ~80% lần sang khu khác đúng như spec.

  **Chưa làm trong mục 2**: mỗi khu 1 goroutine (hiện 5 khu chạy tuần tự trong một goroutine
  — đủ nhanh ở quy mô 20 người, chưa cần song song); 4 thuật toán mapgen mới (mục 6) nên 5
  khu hiện dùng lại 3 bản đồ cũ; băng trơn (mục 7); bảng 5 khu khi bấm M (mục 9); loadtest
  ghi tick từng khu (mục 10).
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

- **`teleport_out` của MỌI nhân vật cắt được 3/4 khung (29/10/2026)** — không phải lỗi ảnh:
  khung cuối là lúc nhân vật đã biến mất nên không còn pixel nào để bộ cắt tìm ra. Sinh lại
  bao nhiêu lần cũng vậy, chỉ tốn quota. Client vẫn chạy bình thường với 3 khung. Vì thế
  `artgen prep` giờ **chỉ đánh dấu sinh lại khi sheet thật sự hỏng**, dùng đúng ngưỡng của
  client (`render/assets.ts`: bỏ animation khi lệch quá 1,5× hoặc dưới 0,5×). Danh sách sinh
  lại nhờ vậy từ 15 xuống 4.
- **Danh sách sheet cắt sai nằm ở `assets/_raw/miscut.json`**, do chính `artgen prep` ghi ra
  chứ không gõ tay vào mã nguồn nữa. Ghi theo kiểu hợp nhất: chỉ gỡ tên nào vừa được cắt lại
  sạch sẽ — ảnh thô của job cũ có thể đã bị dọn đi nên prep bỏ qua chúng, mà coi "không cảnh
  báo" là "đã lành" thì sheet hỏng rơi khỏi ART_JOBS.md trong khi atlas vẫn giữ 22 khung sai.
- **2026-10-22 (phiên sau): 90/211 asset đủ** (artgen: 66 job xong, 130 còn thiếu). Mới có: khung thoại, rương, hit/death Ông Ba Vòng, death Đại Ca, hit Thằng Mập; run Cô Ba Gánh cắt lại được (không cần sinh lại).
  - `artgen run` giờ gửi **mỗi prompt đúng 1 lần**, tuần tự (`-c 1`), nghỉ 30s giữa prompt (`-gap`); hết quota thì chờ tới `reset_at` rồi gửi lại (lúc đó Gemini chưa vẽ gì).
  - `artgen prep`: sheet nhiều hàng mà cắt theo hàng ra sai (hàng dính nhau, ma bay tách khỏi xác) → tự cắt theo lưới cột × hàng, mỗi mảnh vào ô chứa tâm của nó.
  - **Lỗi đã biết (minigen):** sau khi sửa lỗi `metadata`, minigen dùng **chung 1 cuộc chat Gemini** cho mọi job (`chat_id` giống nhau) → Gemini lặp lại hình của prompt trước (5/11 ảnh sai: 4 ảnh vẽ rương, 1 ảnh chép sheet chết Ông Ba). Đã xóa ảnh sai. Cần minigen mở chat mới cho mỗi job trước khi sinh tiếp.
- **2026-10-22: theo ART_PROMPTS mới** — style "bựa", mô tả nhân vật + gag riêng từng anim, **1 animation / ảnh** cho nhân vật (lưới đúng cột × hàng), nền **magenta #FF00FF** (artgen tự nhận nền magenta hay trắng và xóa sạch viền hồng). Nhân vật cũ (style "cute") giữ tạm tới khi ảnh mới về. Tổng 174 ảnh game, đã có 20. Câu hỏi Q6 đã trả lời (xem `docs/QUESTIONS_DONE.md`).
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

- **Spec v5.7 — Tiên giới tính theo TỔNG Tiên khí (30/10/2026, §3.4)**. Trước đây cảnh giới
  Tiên nằm trong `Level` còn Tiên khí cộng dồn từng chặng — **hai nguồn sự thật**, và chúng
  lệch nhau đúng lúc bị chém. Giờ Tiên khí là nguồn duy nhất, cảnh giới suy ra từ nó:
  1–100 Chân Tiên · 101–300 Kim Tiên · 301–600 Thái Ất · 601–1000 Đại La · 1001–1500 Đạo Tổ
  (`tien_khi_ceilings`). Viên Tiên khí **đầu tiên** là vào Chân Tiên ngay.
  - **Vào Tiên giới cần Đại Thừa tầng 10 ĐẦY THANH** (694 960), không phải chỉ cần đứng ở
    tầng 10. Trước đây vừa chạm tầng 10 là Linh khí nhặt được cũng bị bỏ phí trong khi Tiên
    khí rất hiếm, nên người chơi kẹt cứng ở đó.
  - **Đạo Tổ nạp thanh 1 001 → 1 500**, chỉ khi đang đúng 1 500 thì Tiên khí dư mới đổi Uy
    Danh (10:1). Nhờ vậy Đạo Tổ **Thiêu Đốt được** — đó là câu trả lời cho Q26.
  - Thiêu Đốt ở Tiên giới = **8% độ rộng khoảng** của cảnh giới đang đứng: 8 · 16 · 24 · 32 ·
    40 mỗi giây. Theo độ rộng khoảng chứ không một con số chung, nên người ở đỉnh không được
    chạy rẻ hơn người mới vào.
- **Spec v5.8 — bị chém: thanh về MỨC SÀN cảnh giới mới (30/10/2026, §5.1)**. Trước giữ
  nguyên tầng và % tiến độ (Trúc Cơ tầng 2 → Luyện Khí tầng 2); giờ Trúc Cơ tầng 2 → **Luyện
  Khí tầng 1, 0%**, Kim Đan tầng 7 → Trúc Cơ tầng 1 0% (1 360). Chân Tiên bị chém → **sàn
  Đại Thừa** (346 800), Tiên khí 0. `TestRealmFloors` khoá cả 9 con số sàn vì `realm_mult`
  đổi một chữ số là cả bảng lệch mà không có gì báo.
- **Spec v5.9 — chém người Tiên giới không thấy Tiên khí rơi: ĐÃ TÌM RA NGUYÊN NHÂN
  (30/10/2026)**. Server **vẫn rơi đúng** — test tái hiện cho ra 3 viên tổng 24 điểm, khớp
  chính xác con số spec (30% × 80). Lỗi nằm ở chỗ **nhìn thấy**: viên văng trong bán kính
  `items.scatter_r` = 90, trong khi người Tiên giới có scale 5–6 nên thân rộng hơn 150 và
  cao hơn 400 pixel, **mà vật phẩm vẽ DƯỚI nhân vật** (`zIndex = y − 1000`). Viên rơi đúng,
  gửi đúng, nhưng nằm khuất hẳn dưới cái xác to đùng. Sửa theo spec: văng **200–500**
  (`tien_drop_scatter_min/max`), chia viên **10 → 2 → 1**, kèm event `tien_burst`. Phần rơi
  ra đất giờ **luôn** xảy ra kể cả người chém là phàm hay bot.
- **Bỏ `cmd/sim --duel` (30/10/2026, Q25)**. Chủ dự án chốt: kỹ năng có loại tăng sức mạnh,
  có loại khắc chế, nên **tỉ lệ thắng tay đôi không phản ánh giá trị** — cân bằng theo phản
  hồi người chơi và số liệu trận thật. Đã xoá hẳn công cụ thay vì giữ một thứ "có nhưng
  không được dùng".
- **Q20 đã chốt: giữ luật va kiếm đúng.** Mốc mới để so về sau: **84 lượt chết / 5 phút /
  40 bot**. Con số 608 lượt/10 phút ghi ở các mục cũ bên dưới là **đo trên luật va kiếm sai**
  (lỗi dấu `SegSegDist2`), không dùng để so nữa.

- **Minimap bấm mới mở, kill feed còn 1 tin (30/10/2026)**. Chủ dự án: *"khung chat điện
  thoại nhiều quá… mini map đang che mất phần trung tâm, làm nhỏ lại, hoặc cho phép bấm mới
  mở ra"*.
  - Trên **màn hẹp** minimap mặc định thu thành **nút 36px**, bấm mới mở. Ba nấc quay vòng:
    nút nhỏ → bản đồ góc (86px) → bản đồ to (252px) → nút nhỏ. Nấc giữa quan trọng: trước
    đây một cú chạm nhầm là nhảy thẳng sang bản đồ to phủ 70% màn hình, đúng cái "che mất
    phần trung tâm". Trên PC giữ nguyên hai nấc (góc ↔ to) — ở đó bản đồ góc không vướng gì
    mà bắt bấm thêm một lần chỉ làm chậm tay.
  - Kill feed **1 tin** (trước 2), chữ 10px, hộp mỏng. Tin loan toàn server cắt tối đa 2 dòng.
  - Mốc hẹp/rộng gom về `client/src/ui/layout.ts` để `hud.ts`, `minimap.ts` và CSS dùng chung
    một con số — ba nơi ba số thì sẽ có khổ màn hình rơi vào khe giữa chúng (bảng đã co lại
    nhưng minimap vẫn mở to).
  - Đo lại: HUD phủ **18–25%** màn hình (320×568 là 36%), không khối nào chồng nhau.
  - `uioverlap.mjs` sửa lại cho đúng: trước nó **tự nhét 2 tin** rồi đếm 2 — tức là kiểm
    chính cái mình vừa nhét. Giờ nhét dư 5 tin rồi bắn sự kiện `resize` để HUD tự cắt theo
    luật thật, nên con số "feed N tin" mới kiểm được luật.

- **HUD trên điện thoại chiếm gần hết màn hình — đã gọn lại (30/10/2026)**. Chủ dự án gửi ảnh
  chụp: *"chơi game mà không thấy gì"*. Đo lại bằng Chrome ở 4 khổ màn hình, worst case có
  đủ mọi loại thông báo cùng bật:
  - Bảng xếp hạng **Top 3** (trước 10 dòng), bỏ dòng danh hiệu — danh hiệu làm mỗi dòng cao
    gấp đôi mà cảnh giới+danh hiệu đã hiện trên đầu nhân vật rồi.
  - Kill feed **2 tin** (trước 5). Tin loan toàn server 1 dòng.
  - Minimap kẹp thêm theo **tỉ lệ màn hình** (≤ 24% cạnh ngắn): `size_mobile` là số px cố
    định, mà 120px trên máy rộng 360px là một phần ba bề ngang.
  - Khối trạng thái 77 → **54px**: bỏ nhãn "Linh khí" (màu thanh đã phân biệt), gộp
    "Kiếm x/y" và "Uy Danh n" về một dòng.
  - **Mốc "màn hẹp" xét CẢ HAI CHIỀU**, không chỉ bề ngang. Điện thoại xoay ngang là 800×360
    — rộng hơn mốc 720 nên bản đầu coi là "màn rộng" và vẫn vẽ 10 dòng + 5 tin: đo được
    **70% màn hình bị HUD phủ** và 6 cặp khối đè nhau. Chỗ chật ở đó là chiều cao.
  - Kill feed và tin toàn server vào **chung một cột flex** — ghim `bottom` riêng cho từng
    cái thì cái nào dài ra là đè cái kia (đo 212×21px trên máy 320). Cùng bài học với cột
    trái hôm 29/10.
  - Gợi ý xoay máy và dải báo Rương Báu dời khỏi chỗ đè minimap.
  Kết quả: **không khối nào chồng nhau** ở 360×800, 800×360, 320×568 và cửa sổ hẹp 420×900.
- **Hai lỗi CSS rò luật — đã sửa (30/10/2026)**. Cả hai cùng một kiểu: luật viết cho `.portrait`
  trần nên áp cho MỌI chỗ dùng class đó.
  1. `.portrait { width: 64px }` của màn chọn rò sang ô đại diện HUD (40px) → ảnh tràn hẳn
     ra ngoài khung tròn và đè lên chữ "Kiếm x/y". `.badge` không có `overflow:hidden` nên
     không có gì chặn.
  2. `.portrait { position: absolute }` của HUD rò ngược sang màn chọn → ảnh bị kéo ra khỏi
     luồng, **mười thẻ nhân vật trông như không có ảnh**.
  Giờ cả hai khoanh theo thẻ cha (`.badge .portrait`, `.char-card .portrait`), và `.badge`
  có `overflow:hidden` làm chốt chặn; huy hiệu tầng + vương miện chuyển ra `.badge-wrap` vì
  chúng cố tình thò ra ngoài khung.
- **Màn chọn nhân vật có ảnh + kỹ năng (30/10/2026)**. Mười nhân vật chỉ khác nhau ở kỹ năng,
  mà lưới cũ chỉ có tên nên người chơi chọn mù. Thêm khung xem trước: ảnh to, tên, tên kỹ
  năng và mô tả một dòng (`STR.skillDesc`, số lấy đúng spec v5.4/v5.5). Rê chuột để so, bấm
  để chọn. Ảnh lấy `portrait_neutral.png` thay vì `frames/idle_00.png` — khung hình toàn
  thân trên canvas 512 thu về ô 40px thì nhân vật bé bằng hạt gạo.

- **Spec v5.6 — Tiên giới đốt Tiên khí (30/10/2026, §4.1c)**. Đây là câu trả lời cho Q24:
  trước đó lên Tiên giới là mất hẳn khả năng chạy nhanh, vì nhánh duy nhất của Thiêu Đốt
  tiêu EXP mà người Tiên giới đứng đúng ở đáy cảnh giới. Giờ hai đường tu có hai loại nhiên
  liệu: người phàm đốt EXP, người Tiên giới đốt **8% Tiên khí của bước hiện tại mỗi giây**
  (`burn.tien_frac_per_sec`), không xuống dưới mốc khởi điểm của cảnh giới đang đứng.
  "Bước hiện tại" là `tien_khi_per_step[idx]` — bước đã ĐƯA họ tới cảnh giới đang đứng, chứ
  không phải bước đi tiếp; chính ví dụ của spec chốt cách đọc này ("Đạo Tổ: 8% × 500"), vì
  Đạo Tổ không còn bước đi tiếp nào. Đại Thừa tầng 10 chưa nhặt Tiên khí thì chưa đốt được.
  Đo trong game thật (Chân Tiên): Tiên khí **60 → 48 trong 1,5 giây = 8/giây** = 8% × 100,
  đúng con số spec; không rớt cảnh giới; nhả tay là tắt.
- **⚠ Đạo Tổ vẫn không Thiêu Đốt được — chờ chốt (Q26)**. Hai luật đều đúng theo câu chữ
  nhưng cộng lại thì chặn nhau: v5.6 nói không xuống dưới mốc khởi điểm của cảnh giới đang
  đứng, còn `overflowToScore` (Q22) ghim Tiên khí của Đạo Tổ **đúng** ở mốc 1 500 và đổi mọi
  phần dư sang Uy Danh. Nên Tiên khí của Đạo Tổ luôn bằng đáy → không còn gì để đốt, và
  người ở đỉnh lại là người duy nhất không chạy nhanh được. Đã khoá hiện trạng bằng
  `TestDaoToHasNoTienKhiHeadroom`; test ghi rõ là khoá hiện trạng, không khẳng định ý đồ.

- **Spec v5.4 + v5.5 — chỉnh 10 kỹ năng (29/10/2026, §6.3)**. Mọi con số vào `config/game.json`,
  có `TestSkillDurationsMatchSpec` khoá lại để lần sau spec đổi mà quên sửa config thì test đỏ.
  - **Bánh Bao Thần đổi hẳn luật**: trước hồi 30% số kiếm bể trong 10 giây, giờ hồi thẳng
    **20% số kiếm tối đa của cảnh giới**. Luật cũ có lỗ không chữa được bằng cách chỉnh số:
    ai chưa bị đánh thì ăn bánh bao chẳng được gì, nên kỹ năng vô dụng đúng lúc muốn dùng
    nhất. Sổ ghi kiếm bể (`lostSwords`) gỡ bỏ hẳn.
  - **Gọi Đàn Em**: 2 đàn em, 10s, mỗi con mang **50% số kiếm hiện tại của đại ca** (tối
    thiểu 1) xếp vòng như người chơi, nhỏ hơn 30%, không nhặt đồ, không qua cổng. Đại ca
    không bị trừ kiếm.
  - **Giao Hỏa Tốc**: ×3 trong 5s, xuyên **mọi** vật cản (trước chỉ xuyên thứ phá được),
    trừ biên khu. Hết giờ mà kẹt trong vật cản → trượt mượt ra chỗ trống trong 0,3s bằng
    chính `resolveBlocking` — không bật ra tức thời, vì nhảy một phát là client thấy đúng
    kiểu giật mà §8.2a vừa sửa xong.
  - **Phun Trân Châu**: 5s, và trong 5s đó **không nhặt được gì** — chặn ở đúng một cửa
    (`CanPickUp`, ngay đầu `resolvePickups`) chứ không rải vào từng loại vật phẩm, vì spec
    nói cả vật phẩm/Linh khí/Tiên khí/Bí Kíp đều đi xuyên qua. Người ngoài thấy icon 🚫 bay
    quanh nạn nhân: lớp trân châu chỉ nạn nhân nhìn thấy, không có icon thì từ ngoài nhìn
    vào không ai hiểu vì sao họ đứng cạnh viên Linh khí mà không nhặt.
  - **Gánh Lướt**: vệt trái cây giờ gây **choáng** (`StStun`) cùng lúc với trượt — không đi,
    không dùng kỹ năng, không Thiêu Đốt, kiếm vẫn quay. Tách khỏi `StSlip` vì trượt là đổi
    *cách* di chuyển còn choáng là *cấm* di chuyển.
  - Lưới Mì 5s · Gồng Cơ 5s · Rụt Cổ 5s · Loa Phường đẩy 500 · Ru Ngủ bán kính 500.
- **Bài đối chiếu va chạm phủ thêm Giao Hỏa Tốc (29/10/2026)**. Kỹ năng này đổi **cả hai**
  thứ quyết định vị trí — tốc độ ×3 và xuyên vật cản — nên nếu client không biết thì lệch
  mỗi tick. 1/4 số ca trong `cmd/collisionfixture` giờ đang xuyên. Đo: **0,0000 đơn vị lệch**
  trên 1 500 bước × 5 bản đồ; và khi cố tình tắt phần xuyên ở client thì bài kiểm báo **46 ca
  quá ngưỡng, nặng nhất 100,6 đơn vị** — nên đây không phải một dòng chữ xanh vô nghĩa.
- **⚠ Bài đấu từng cặp trước đây đo sai (sửa 29/10/2026)**. `--duel` cho mỗi bên **một vòng
  kiếm** (10 kiếm) trong khi trần Kim Đan là **100** — tức 1/10 số kiếm thật của người chơi
  cùng cảnh giới. Mọi kỹ năng tính theo TRẦN cảnh giới bị thổi phồng gấp mười: Bánh Bao Thần
  ra **100% thắng (54/54)** với một vòng, **75,9%** với đủ kiếm. Lý do cũ của việc dùng một
  vòng là "2 vòng trở lên thì 85% trận hoà", nhưng con số ấy đo TRƯỚC khi `clash_cooldown`
  về 0; đo lại thì đủ kiếm chỉ **9% hoà**, còn ít hơn một vòng (18%). Mặc định đã đổi sang
  **đủ trần cảnh giới**; `--duel-swords N` để quay lại kiểu cũ.

- **Data race thứ 6 — `moveToZone` đọc map của khu khác (sửa 29/10/2026)**. Lúc chia năm khu
  ra năm goroutine tôi tìm và sửa 5 race; con thứ 6 lọt vì nó nấp sau một hàm trông vô hại.
  `MoveToZone` (cổng dịch chuyển gọi, chạy TRONG goroutine của một khu) gọi `zoneOf(p.ID)`,
  mà `zoneOf` quét map `Players` của **cả 5 khu** — trong khi bốn khu kia đang tick và có thể
  đang xoá người. Ngoài đời đây không phải lỗi nhỏ: Go gặp đọc-ghi map song song thì
  `fatal error: concurrent map read and map write`, **sập cả server**. Sửa bằng cách gắn cứng
  hàm chuyển khu với khu cài nó (`moveFromZone(z)`) — khu nguồn vốn đã biết, không cần đi tìm.
  Đã ghi cảnh báo lên `zoneOf`. Kiểm: 30 lượt `-race` → 0 race (trước đó 3/20 lượt dính).
- **Phòng đủ 60 người mà vẫn có khu 8–9 (sửa 29/10/2026)**. Hai lỗi chồng nhau:
  1. `rebalanceZones` lặp 5 lượt nhưng việc chuyển khu chỉ xảy ra ở **ranh giới tick**, nên
     `z.Count()` không nhúc nhích trong lúc lặp. Cả năm lượt đều thấy tình trạng cũ → xếp
     năm lệnh cùng chiều → khu cho đi bị rút quá tay thành khu thiếu mới → lượt sau kéo
     ngược về, qua lại mãi. Giờ tự cộng sổ phần đã xếp hàng (`delta`) và quyết theo con số
     **dự kiến**, thêm điều kiện cho đi xong không được tụt dưới sàn, và `spareBot()` không
     bốc trúng con đã nằm trong hàng chờ.
  2. Vá sàn trước đây chạy theo **nhịp sinh bot 2 giây**, mà bot đi qua cổng bất cứ lúc nào
     → khu có thể nằm dưới sàn gần 2 giây. Giờ `rebalanceZones` chạy **mỗi tick** (rẻ: 5 khu,
     vài phép so), nên người vừa bước qua cổng không rơi vào khu gần như trống.
  `processRespawns` cũng dính cùng kiểu sai: bot hồi sinh rời khu đang **đúng sàn** thì chỗ
  thiếu chỉ chạy sang khu khác chứ không mất — đã chặn.
  `TestZoneFloorAndRoomCap` sửa lại cho đúng điều server hứa: sàn là *"tụt xuống thì được vá
  lại ngay"*, không phải *"không khoảnh khắc nào dưới sàn"* (bot có thể rời khu ở đúng tick
  cuối, lệnh bù chỉ chạy được ở tick sau). Kiểm: 40 lượt `-race` xanh; và vẫn đỏ khi cố tình
  tắt `rebalanceZones`, nên không phải nới lỏng cho qua.
- **Điều khiển gọn lại (29/10/2026)**. Thiêu Đốt Tinh Huyết thêm cách **nháy đúp rồi giữ**
  (300 ms, lệch dưới 40 px) bên cạnh Shift và nút 🔥 — hai cách cũ đều đòi tay thứ hai, còn
  nháy đúp giữ nguyên ngón/tay đang lái. Luật nhận nháy đúp (`tapDown`) viết một lần dùng
  chung cho chuột và cảm ứng, vì tách làm hai bản là kiểu lỗi chỉ hiện trên đúng một loại
  thiết bị. Thêm phím test **G** (+1 lượt kỹ năng) — không có nó thì muốn thử kỹ năng phải
  chạy quanh bản đồ chờ Bí Kíp rơi.
- **Kỹ năng giữ phím Space, KHÔNG dùng Ctrl (29/10/2026)**. Có thử Ctrl một nhịp rồi bỏ; lý
  do ghi lại ở đây để đừng ai đổi lại: phím đi là **WASD**, nên vừa giữ Ctrl vừa bấm W là
  **Ctrl+W — đóng tab**. Chrome xếp Ctrl+W vào nhóm phím tắt dành riêng cho trình duyệt,
  `preventDefault()` không chặn được. Cách duy nhất chặn thật là Keyboard Lock API, mà nó
  **chỉ chạy ở chế độ toàn màn hình** (ràng buộc cố ý của trình duyệt, để trang web bất kỳ
  không giam được người dùng trong tab) và chỉ có trên Chrome/Edge — tức là người chơi cửa
  sổ thường vẫn mất tab. Space không đụng phím tắt nào. Cùng lý do đó, tránh `Ctrl`, `Alt`
  và `F`-key cho bất kỳ phím game nào sau này.
- **HUD trái tự đè lên nhau — đã sửa (29/10/2026)**. Ba khối bên trái (chỉ số · hàng buff ·
  minimap) trước đây mỗi khối một `top:` cố định (12 / 84 / 136 px). Hễ khối chỉ số cao thêm
  một dòng — thanh Tiên khí ở Tiên giới, rồi nhãn tên khu mới thêm — là nó tràn xuống đè
  khối dưới. Đây là loại lỗi tự sinh lại: mỗi lần thêm một dòng vào HUD lại phải nhớ dời tay
  cả ba con số, và quên là chuyện chắc chắn xảy ra. Giờ gom vào `.hud-left` xếp dọc bằng
  flex nên khối sau tự đẩy xuống, thêm bao nhiêu dòng cũng không hỏng. Bảng hướng dẫn lúc
  mới vào (`.guide`) cũng đang nằm đè minimap → chuyển xuống **giữa-dưới**. Đo ở trường hợp
  xấu nhất (Đạo Tổ, có thanh Tiên khí, 1 050 kiếm, 5 buff): 12 → 119 → 171 → 205, không
  khối nào chồng nhau.
- **Nền bản đồ có vệt hồng — đã sửa (29/10/2026)**. 7 ảnh nền (`soft_grass` 17,3%,
  `bamboo_litter` 15,7%, `mossy_stone` 10,3%, …) vẫn còn magenta nguyên bản. `modeTile` cố ý
  **không** lọc magenta cho ảnh nền vì nền phải đặc — nhưng magenta *trong* ảnh nền nghĩa là
  bên sinh ảnh không vẽ kín khung, chứ không phải nền cần giữ. `patchMagenta` (`cmd/artgen/prep.go`)
  vá các pixel đó bằng màu trung bình vùng lành **và** ghi job vào `miscut.json` để lần sau
  vẽ lại. Quét lại: 0 asset còn magenta, chụp màn hình đo 0,000% magenta.
- **Giật khi chạm vật cản — đã sửa 3 nguyên nhân, còn sót ít (29/10/2026, spec v5 §8.2a)**.
  Thêm vào F3 dòng `hiệu chỉnh N/giây · lệch nhất X` (đếm số lần server kéo nhân vật về).
  Đo trước khi sửa: đi tới rìa bản đồ ra **20,2 lần/giây**, nhân vật ghim cứng ở y = 11 549.
  Ba chỗ client dự đoán khác server, tất cả đều bị bài đối chiếu 1 500 bước bắt được:

  | # | Lệch ở đâu | Hậu quả |
  |---|---|---|
  | 1 | Client bỏ qua `play_margin` (kẹp `[r, size−r]` thay vì `[margin+r, size−margin−r]`) | Cả vành ngoài rộng 420 đơn vị — ~13% bản đồ — bị kéo về mỗi tick |
  | 2 | Thứ tự đẩy ra khỏi vật cản khác nhau (lưới không gian vs mảng) | 4/1 000 ca lệch, nặng nhất **173,7 đơn vị** |
  | 3 | Client thiếu phép xét **đa giác** và thiếu luật vùng **nước** | Đi qua bùn/nước/gió hình méo là lệch tốc độ liên tục |

  Sau khi sửa: bài đối chiếu **0,0000 đơn vị lệch** trên 1 500 bước × 3 bản đồ; đo trong game
  **20,2 → 4,8 lần/giây** lúc cao nhất, nhân vật không còn ghim ở rìa.

  **Còn sót 1–5 lần/giây**: gần như chắc là đòn đẩy văng (`KnockVel`) — client không dự đoán
  được va chạm chiến đấu của server. Chưa xác minh; nếu muốn triệt thì phải gửi kèm vận tốc
  đẩy trong snapshot để client mô phỏng theo.

- **Thỉnh thoảng mất ngữ cảnh WebGL (28/10/2026)**: chạy 3 lượt kiểm tra bằng Chrome headless
  trên AMD RX 570 (RADV) thì **1 lượt** in `PixiJS Error: Could not initialize shader` kèm
  `WebGL context may be lost`. Khung hình không tụt (58 fps) nên có vẻ chỉ là trình điều
  khiển trong môi trường headless. **Nhưng client hiện KHÔNG bắt sự kiện `webglcontextlost`**
  — nếu chuyện này xảy ra với người chơi thật thì màn hình sẽ đen vĩnh viễn, phải tự tải lại
  trang. Chưa làm vì chưa thấy tái hiện trên máy thật; nếu có người báo "màn hình đen" thì
  đây là chỗ nhìn đầu tiên.
- **Tài khoản minigen `pt1` đang báo UNAUTHENTICATED (cookie hết hạn)** → cần đăng nhập lại cookie bên minigen rồi chạy tiếp `artgen run`.
- Gemini đôi khi vẽ sai số frame (5 thay vì 4, 6 thay vì 8). `artgen prep` giữ đúng số frame thực tế vẽ ra; anim vẫn chạy được.
- Ảnh của Bà Tám Loa lấy từ bản JPEG 1024px (người dùng tải từ Gemini) → hơi mờ, có ngôi sao watermark nhỏ; nên sinh lại khi dư quota (`artgen run -only char_ba_tam_loa_core -redo`).
- Cổng 8080 trên máy dev hiện đang bị chương trình khác chiếm → chạy server với `-addr :8090` (khi dùng `npm run dev` thì chạy `KX_SERVER=http://localhost:8090 npm run dev`).
- Cân bằng (sim 10 phút, 40 bot): lên Trúc Cơ trung bình ~7 phút, nhiều bot chết sớm ở Luyện Khí 1 — nếu thấy chậm, chỉnh `config/game.json` (ví dụ tăng EXP quái, giảm `exp_base`). Chưa đổi gì vì đây là quyết định cân bằng của bạn.
- Dự đoán chuyển động phía client không mô phỏng đẩy lùi (knockback) → khi bị chém thấy giật nhẹ 1 nhịp.
- **Spec v4.0/v4.1 — ĐÃ LÀM (22/10/2026)**: số cân bằng mới (`move_base` 195, `cell` 140), độ bền kiếm `durability_per_realm` 1.2 thay `clash_ratio_*`, UI không tràn màn hình, 3 lỗi bản đồ (lề 300 + tường biên + dãy tường nối liền). Test Tốc Kiếm không đụng `reach` — PASS ngay, logic vốn đã đúng.
- **Spec v4.0 — ĐÃ XONG CẢ 3 VIỆC LỚN (22/10/2026)**: lá chắn vòng kế ngoài cùng §5.2 (xem Q15 về mâu thuẫn công thức), Tiên khí + 5 Đài §3.4/§8.5, mapgen viết lại theo 9 bước §8.1 (`cmd/mapgen/gen.go`, có `--seed N --png out.png` xem trước).
- **Bản đồ sinh động (28/10/2026, spec v4.9 §8.6, CHANGES_v4 mục 43–47)**: vùng bùn/ao là **đa giác méo theo nhiễu** (`r(θ) = r0 × (1 + 0.25·nhiễu(θ))`, Chaikin ×2, 56 đỉnh) — `Shape` thêm `Poly`, **server xét đúng đa giác client vẽ**. 3 ao mỗi bản đồ có mặt nước gợn sóng. Decal rải Poisson-disk theo biome, bóng đổ + vụn dưới chân vật cản, vật cản xoay ±6°/lật theo id. Sự sống: mây trôi, lá rơi, đom đóm, sương, bướm, bóng chim — trần 150/60/0 hạt theo chất lượng. 80 job art mới vào `assets:report` (nền biome, decal, viền vùng, ambient, bảng tên, chân dung HUD).
  **Ba lần đo, ba lần sửa hiệu năng** (bản đầu làm 58 → 36 fps, draw call 237 → 611): (1) decal và hạt là `Graphics` riêng từng cái → nướng thành texture dùng chung; (2) ~2 000 decal rải khắp bản đồ bị duyệt hết mỗi khung → gom ô 2048 và tắt `renderable` ô ngoài khung nhìn; (3) bóng đổ vật cản vẫn là Graphics, 286 vật cản = 286 lượt vẽ → nướng nốt. Kết quả **draw call 611 → 118**, vẽ 0,9 ms.
- **Spec v4.6 VFX (28/10/2026)**: `vfxLayer` riêng gắn `AdvancedBloomFilter` một lần cho cả lớp; hào quang Tiên 4 lớp blend cộng; đột phá có `ShockwaveFilter` + khựng hình 80ms + chớp trắng + chữ cảnh giới bật ra. Cài `pixi-filters`, `revolt-fx`.
- **Spec v4.7 thần thú (28/10/2026)**: bay quanh chủ liên tục, động tác ngẫu nhiên mỗi 8–15s, Đạo Tổ 2 con lệch pha 180°, phản ứng lóe sáng khi nhặt Tiên khí / khiên đỡ đòn, bay sát + vệt đỏ khi Thiêu Đốt. Phím **V** đổi chất lượng Cao/Vừa/Thấp; máy yếu thì **tự hạ** (FPS < 45 suốt 3s hạ một bậc, > 55 suốt 10s nâng lại), nhưng người chơi đã tự chọn thì khoá.
- **Spec v4.8 bảng tên + v4.9 ảnh đại diện (28/10/2026)**: xem commit `c840446` và `eca03d2`.
- **Thần thú luôn lượn lờ (28/10/2026, spec v4.7 §3.5, CHANGES_v4 mục 34 + 36)**: bỏ kiểu "mỗi 20s bay 1 vòng" và bỏ luôn nút tắt. Sau đột phá, thần thú thu nhỏ 60% và **ở lại bay quanh chủ suốt thời gian còn ở Tiên giới**: quỹ đạo elip 1.3–1.8 × chiều cao nhân vật, 1 vòng/4s, ép dẹt 0.35, độ cao nhấp nhô chu kỳ 3s, tâm bám chủ lerp 0.1 nên lúc chạy bị kéo dài ra sau. Mỗi 8–15s làm một động tác 1.5–3s (xoắn ốc vút lên · số 8 · đổi chiều · lao ra rồi về · cuộn quanh đầu), vào/ra bằng bao hình sin nên không giật. Đạo Tổ có **2 con lệch pha 180°**, một trắng một đen. Đốt sau đi theo vị trí đầu các khung trước, lưu trong `Float64Array` cấp phát một lần (spec cấm tạo object mỗi frame). Chất lượng Cao/Vừa/Thấp = 16/10/6 đốt, vệt hạt 100/50/0%, chủ nhỏ hơn 40 px chỉ vẽ đầu + 4 đốt, vệt hạt chỉ bật cho 6 con gần nhất. Ẩn Thân/trong bụi thì mờ 20% theo chủ. **Phím V giờ đổi mức chất lượng** (§3.6) thay cho bật/tắt thần thú. Đo tại chỗ: Đạo Tổ song long, 9 người, 1 912 entity trong AOI → **60 fps, vẽ 1,0 ms, 245 draw call**.
- **Đo hiệu năng TRƯỚC tối ưu (28/10/2026, CHANGES_v4 mục 24)** — `cd server && go run ./cmd/loadtest -secs 15`, đúng kịch bản §13.6 (40 nhân vật, 10 người Đạo Tổ đủ 1 050 kiếm đứng chụm giữa bản đồ):

  | Chỉ số | Đo được | Ngân sách §13.6 | |
  |---|---|---|---|
  | Tick server (trung bình) | **11,45 ms** | < 4 ms | VƯỢT ~3× |
  | Tick server (p95) | **19,89 ms** | < 4 ms | VƯỢT ~5× |
  | Snapshot lớn nhất / client | **38,3 KB** | < 8 KB | VƯỢT ~5× |
  | Băng thông xuống / client | **43,1 KB/s** | < 40 KB/s | VƯỢT nhẹ |
  | Ping p95 | 0,00 ms | < 20 ms | ĐẠT |

  Ping 0 ms vì loadtest chạy trong cùng tiến trình, không qua mạng thật — con số này chỉ
  chứng minh đường trả Pong **không phải chờ vòng lặp game**, không nói gì về mạng thật.
  Snapshot 38,3 KB là đỉnh lúc một loạt vật phẩm cùng lọt vào tầm nhìn khổng lồ của người
  Đạo Tổ (`ItemsAdded` dồn vào một gói) — đúng chỗ mục 27 (delta + chỉ gửi sự kiện) nhắm tới.
- **Mục 27 — gói snapshot gọn lại (28/10/2026)**: cùng lệnh `go run ./cmd/loadtest`, cùng kịch bản:

  | Chỉ số | TRƯỚC | SAU | Ngân sách §13.6 | |
  |---|---|---|---|---|
  | Tick server (trung bình) | 11,45 ms | **14,04 ms** | < 4 ms | VƯỢT ~3,5× |
  | Tick server (p95) | 19,89 ms | **25,09 ms** | < 4 ms | VƯỢT ~6× |
  | Snapshot lớn nhất / client | 38,3 KB | **8,3 KB** (−78%) | < 8 KB | sát mép |
  | Băng thông xuống / client | 43,1 KB/s | **21,4 KB/s** (−50%) | < 40 KB/s | ĐẠT |
  | Ping p95 | 0,00 ms | 0,00 ms | < 20 ms | ĐẠT |

  Làm những gì: `ItemsAdded` cắt theo ngân sách 2 KB/gói ưu tiên vật phẩm gần người chơi ·
  bỏ viên EXP nhỏ ngoài 1 500 khi camera đã lùi xa · người/quái ngoài nửa tầm nhìn chỉ gửi
  10 Hz thay vì 20 Hz · trần 150 entity mỗi gói · client nghẽn thì **bỏ gói cũ** chứ không
  xếp hàng (`SendLossy`) · quái chỉ gửi loại / máu tối đa / bán kính **lần đầu** client thấy
  nó (46 → 27 byte/con) · người chơi bỏ `omega` (client tự tính từ config), `mood` thành chỉ
  số thay cho chuỗi, `exp` chỉ gửi cho chính chủ (102 → 88 byte/người).

  Tick server **xấu đi** chứ không tốt lên — việc xếp vật phẩm theo khoảng cách tốn thêm CPU.
  Đó là mục 26, xem pprof bên dưới.
- **pprof: 88% CPU server nằm trong `diffItems` (28/10/2026, mục 26)** — `go run ./cmd/loadtest -secs 30 -cpuprofile cpu.prof`, 10 hàm tốn nhất:

  | | Hàm | flat | cum |
  |---|---|---|---|
  | 1 | `math.archExp` | 19,2% | 19,2% |
  | 2 | `math.archHypot` | 8,1% | 8,1% |
  | 3 | `maps.(*Iter).Next` (duyệt map) | 7,0% | 8,6% |
  | 4 | `room.(*Room).diffItems` | 6,0% | **87,9%** |
  | 5 | `game.(*Game).ViewHalfExtents` | 6,0% | **44,0%** |
  | 6 | `math.archLog` | 4,9% | 4,9% |
  | 7 | `maps.ctrlGroup.matchH2` | 3,9% | 3,9% |
  | 8 | `runtime.mapaccess1_fast32` | 3,9% | 15,5% |
  | 9 | `math.pow` | 3,5% | 35,1% |
  | 10 | `math.ldexp` | 3,3% | 3,7% |
  |  | `game.(*Game).InViewRadius` | 3,0% | 48,1% |
  |  | `room.(*Room).wantItem` | 1,4% | 47,9% |

  Đọc bảng: gần như toàn bộ CPU đi vào vòng lặp lọc vật phẩm theo tầm nhìn, và bên trong nó
  `ViewHalfExtents` gọi `math.Pow(Scale, CameraExp)` — tức `exp(log())` — **cho từng vật phẩm,
  từng client, từng tick**, trong khi giá trị đó chỉ phụ thuộc người xem. `archHypot` là do
  xếp vật phẩm theo khoảng cách dùng `math.Hypot` thay vì so bình phương. Nhóm `maps.*`
  (≈15%) là tra `knownItems` và duyệt toàn bộ vật phẩm của bản đồ cho mỗi client.
- **Mục 26 — sửa xong, tick 14,04 → 2,55 ms (28/10/2026)**:

  | Chỉ số | TRƯỚC mục 27 | SAU mục 27 | SAU mục 26 | Ngân sách |
  |---|---|---|---|---|
  | Tick server (trung bình) | 11,45 ms | 14,04 ms | **2,55 ms** | < 4 ms ✅ |
  | Tick server (p95) | 19,89 ms | 25,09 ms | **5,00 ms** | < 4 ms ❌ |
  | Snapshot lớn nhất / client | 38,3 KB | 8,3 KB | **8,0 KB** | < 8 KB ✅ |
  | Băng thông xuống / client | 43,1 KB/s | 21,4 KB/s | **20,0 KB/s** | < 40 KB/s ✅ |
  | CPU server | 42% | 42% | **8%** | |

  Bốn việc, theo đúng thứ tự pprof chỉ ra:
  1. `ViewHalfExtents` gọi `math.Pow(Scale, CameraExp)` — tức `exp(log())` — cho **từng vật
     phẩm, từng client, từng tick**, trong khi nó chỉ phụ thuộc người xem. Giờ tính sẵn
     trong `refreshDerived` (`Player.ViewHW/ViewHH`) và lấy một lần trước vòng lặp.
  2. Xếp vật phẩm theo khoảng cách dùng `math.Hypot` → so **bình phương** khoảng cách.
  3. `diffItems` duyệt **toàn bộ** `g.Items` cho mỗi client → dùng `itemGrid` vốn đã có sẵn,
     chỉ quét những ô chạm khung nhìn. Ô lưới đổi từ `map[uint32]*Item` sang slice, xoá O(1)
     bằng hoán đổi với phần tử cuối (có `itemgrid_test.go` riêng vì kiểu xoá này dễ làm mất
     phần tử khác).
  4. `knownItems` của client đổi từ `map[uint32]bool` sang **slice theo chỉ số dày đặc**
     (`Item.Slot()`), lưu id để nhận ra slot đã được viên khác dùng lại. Hết hẳn băm trong
     vòng nóng.

  Rồi phần tối ưu đó **phơi ra một lỗi hình học thật** (xem mục dưới); sửa lỗi xong thì CPU
  tăng lại, nên phải tối ưu tiếp vòng so kiếm. Số cuối cùng, ba lần chạy liên tiếp:
  **tick 4,88 / 5,01 / 5,04 ms · p95 6,2–6,7 ms · snapshot 8,1–8,5 KB · 22 KB/s**.
- **LỖI: `SegSegDist2` sai dấu — kiếm va nhau ở vị trí ma (28/10/2026)**: công thức giao
  điểm hai đoạn thẳng viết `cross(d2, r.Mul(-1))` thay vì `cross(d2, r)`, tức **đảo dấu cả
  hai tham số** t và u. Hệ quả: hàm báo "hai đoạn cắt nhau" đúng vào lúc phần **kéo dài
  ngược** của chúng cắt nhau. Đo được một trường hợp thật: hai thanh kiếm cách nhau 207 đơn
  vị (xa hơn tổng chiều dài hai lưỡi) vẫn bị tính là va vào nhau. Bắt được nhờ bản tối ưu
  lọc đúng theo hình học lại "lệch" so với bản cũ. Đã sửa, có `internal/geom/segseg_test.go`
  khoá cả hai chiều.

  **Ảnh hưởng cân bằng — cần người chủ dự án xem lại**: `go run ./cmd/sim --bots 40
  --minutes 5` cho **84 lượt chết** (2,10 lượt/bot). Con số cũ ghi trong tài liệu là 608
  lượt trong 10 phút (≈ 304 trong 5 phút) — nhưng con số cũ đo khi va kiếm còn lỗi, tức là
  có rất nhiều cú va ảo. Nay va kiếm chỉ tính khi hai lưỡi thật sự chạm nhau nên bể kiếm
  chậm hơn nhiều. Đây là **đổi nhịp trận đấu**, không phải chỉnh số — nên ghi vào
  `docs/QUESTIONS.md` chờ quyết định, chưa đụng vào `config/game.json`.
- **Vòng so kiếm (`sparseRingsTouch`) — 83% CPU server (28/10/2026)**: pprof trên **server
  thật** (có bot, bản đồ thật, 10 tab trình duyệt thật) cho hình ảnh khác hẳn `cmd/loadtest`:
  ở đó 88% CPU nằm ở lọc vật phẩm, còn ở đây 83% nằm ở so kiếm. Ba việc đã làm:
  1. Kiếm của b tính sẵn một lần mỗi bước quét — bản cũ gọi `swordAt(b…)` đúng
     `a.Count × b.Count` lần thay vì `b.Count` lần.
  2. Bỏ sớm lưỡi nằm ở phía đối diện (cách tâm đối phương quá `bán kính ngoài + nửa lưỡi +
     bề rộng`), và so khoảng cách hai **tâm** lưỡi trước khi gọi `SegSegDist2`.
  3. `swordAtRot`: hướng lưỡi suy từ hướng chuôi bằng một phép quay thay vì cos/sin lần hai;
     `RingOf` tính một lần thay vì hai; trả luôn chiều dài lưỡi để khỏi đo lại bằng `Hypot`.

  **Hai hướng đã thử và BỎ** (đo ra tệ hơn, ghi lại để khỏi thử lại):
  - Nhớ vị trí kiếm theo từng người mỗi tick → **12,01 ms**. Số bước quét `n` phụ thuộc
    *cặp* người chơi nên bộ nhớ đệm bị huỷ liên tục.
  - Dựng sẵn bảng hướng kiếm trong `ComputeRingGeom` → **9,53 ms**; chuyển sang dựng lười
    trên Player → **9,61 ms**. Lý do: bể kiếm xảy ra mỗi tick nên bảng 1 050 phần tử phải
    dựng lại liên tục, tốn hơn phần tiết kiệm được.
- **Bài test 10 tab, TRƯỚC/SAU mục 26 (28/10/2026)** — cùng kịch bản, 5 tab Đạo Tổ 1 050
  kiếm chụm giữa bản đồ, ~5 300 kiếm trên màn hình, ~1 400 entity trong AOI:

  | Chỉ số (nhóm Đạo Tổ, TB) | TRƯỚC | SAU |
  |---|---|---|
  | Tick server | 269 ms | **8,2 ms** |
  | Ping | 101 ms | 339 ms* |
  | Kích thước snapshot | 1,2 KB | 1,7 KB |
  | FPS (tab 1) | 16,2 | 4,2 |
  | Thời gian vẽ (tab 1) | 2,2 ms | 16,1 ms |

  *Ping và FPS xấu đi vì nay **nút thắt chuyển sang client**: trước đây server nghẽn nên
  client nhận ít gói và vẽ ít; giờ server trả đủ 20 Hz, mười tab cùng vẽ 5 000 sprite trên
  một card RX 570 và chính việc vẽ trở thành chỗ tắc (draw call 431). Đây chính là điều
  kiện mà kế hoạch đặt ra cho **mục 25.1 (gộp vòng kiếm thành một RenderTexture)** và
  **mục 25.6 (giải mã msgpack trong Web Worker)** — hai mục này nay có bằng chứng để làm.
- **Bài test 10 tab dùng GPU thật (28/10/2026)**: máy có AMD RX 570; đo được
  `--use-gl=angle --use-angle=vulkan` → RADV, `--use-angle=gl` → radeonsi, còn
  `--use-gl=swiftshader` và `--use-gl=egl` đều rơi về SwiftShader (CPU). Đổi sang Vulkan
  nên cả 10 tab chạy khổ 960×540 với chất lượng mặc định. Server chạy riêng trên lõi CPU
  tách khỏi trình duyệt bằng `taskset`, vì không tách thì tick 2,55 ms bị thổi lên 269 ms
  do goroutine bị hệ điều hành gạt ra liên tục.
- **`cmd/loadtest` và `cmd/stress` (28/10/2026)**: `loadtest` đo đúng một kịch bản §13.6 và in bảng đối chiếu ngân sách; `stress` quét 10 → 320 người tìm trần của máy (đo được: **160 người vẫn mượt**, 200 người lag). Cả hai chạy nguyên đường thật (ghép phòng, bot, snapshot theo tầm nhìn, mã hoá msgpack), khác `cmd/sim` vốn chỉ chạy `game.Tick` thuần.
- **Overlay F3 (28/10/2026, §13.6)**: `?debug=true` → phím **F3** bật/tắt bảng đo: fps · ms/khung · thời gian vẽ · tick server (server gửi kèm trong snapshot) · ping · KB/s · cỡ snapshot · số sprite · số draw call (bọc hàm vẽ WebGL) · số entity trong AOI · số kiếm · mức giãn kiếm.
- **Spec v4.3–v4.5 (28/10/2026)**: bỏ hẳn **Trường Kiếm** — tầm chém chỉ tăng theo số vòng và cảnh giới, `ComputeRingGeom`/`ringGeom` bỏ tham số `longMult`. Thêm **Cương Kiếm** (`hard`): độ bền va kiếm × 1.5, sát thương quái/vật cản × 2, 10s, weight 7. Bể kiếm giờ so theo **độ bền D = 1.2^r × hệ số buff** chứ không so cảnh giới, nên Luyện Khí có Cương Kiếm cứng hơn Trúc Cơ tay không; đổi chiều quay vẫn chỉ xét cảnh giới. `shield_ring` → `body_hit_max_rings: 1`; bỏ `tien_khi_to_exp_base`.
- **Thần thú Tiên giới (28/10/2026, spec v4.2 §3.5)**: VFX chuỗi đốt ở client (`render/beast.ts`) — 1 đầu + 14 đốt + 1 đuôi, đốt sau đi theo vị trí đầu lúc trước như con rắn, quỹ đạo xoắn ốc từ chân lên, đốt ở nửa sau vòng xoắn vẽ dưới nhân vật nửa trước vẽ trên. 3 kiểu: đột phá 2.5s · thường trực mỗi 20s 1.5s (phím **V** tắt/bật) · gãy vụn khi rớt cảnh giới 0.8s. Trần 6 con cùng lúc, Ẩn Thân/trong bụi không phát. Placeholder tam giác+tròn; 15 job art đã vào `assets:report` nhóm "Thần thú Tiên giới" theo ART_PROMPTS §10.1.
- **Linh khí và Tiên khí tách hẳn (28/10/2026)**: từ Đại Thừa tầng 10 trở lên chỉ hấp thụ Tiên khí, dưới đó chỉ hấp thụ Linh khí — viên không dùng được đi xuyên qua và client làm mờ 30%. Lệch §3.4 (spec còn ghi đổi 50×2^r) → **Q18**. Phím test **R** giờ cấp thẳng Tiên khí khi đã qua Đại Thừa 10, nếu không sẽ tắc không thử được Tiên giới.
- **Bể kiếm nhanh gấp 9 lần (28/10/2026)**: tách `clash_cooldown` (mới, = 0 → tính mỗi tick) khỏi `flip_cooldown` (giữ 0.3s, chỉ giới hạn đổi chiều quay). Hai vòng ngoài 30 kiếm ngang cảnh giới mài sạch nhau trong đúng 1 giây. Lệch câu "cooldown 0.3s" của §5.2 → **Q19**.
- **Người chơi không chặn đường nhau (28/10/2026)**: lá chắn §5.2 chuyển từ luật DI CHUYỂN sang luật CHÉM — đi xuyên qua nhau thoải mái, nhưng còn ≥ 2 vòng thì kiếm địch không chạm tới thân. Giả lập 40 bot/10 phút: 148 → 608 lượt chết. Hệ quả: Bó Kiếm khi rớt cảnh giới gần như không còn xảy ra → **Q17**.
- **Không ai lọt ra ngoài tường biên (28/10/2026)**: `resolveBlocking` kẹp vào vùng chơi chứ không phải cả bản đồ (trước kẹp `[r, size-r]` tức cả dải lề ngoài tường). Quái vật cũng vậy. `internal/game/bounds_test.go` giữ luật này.
- **Vành mờ vòng kiếm (28/10/2026)**: chỉ vẽ cho vòng thật sự quét kín (≥ 30 kiếm, hoặc ≥ 19 khi có Tốc Kiếm) — vòng thưa không còn tô vành mờ làm phạm vi nhìn rộng hơn thực tế.
- **Cửa sau debug (28/10/2026)**: mở `?debug=true` thì có `window.kx = { scene, world }` để soi bằng console (ví dụ `kx.scene.beastCount()`).
- **⚠ Nhịp farm đang lệch spec (22/10/2026)**: `--farm` đo được Luyện Khí → Trúc Cơ **165s**, §6.1 vẫn ghi 30–60s. Do ba thay đổi cộng lại: tốc độ chạy một nửa + EXP gấp đôi mỗi cảnh giới + mật độ giảm 50%. Đại Thừa giờ cần ≈7.7 giờ farm → bảng Thần Tốc sẽ trống. Phương án: **Q13 trong `docs/QUESTIONS.md`** — chờ chốt, chưa đổi gì.
- **EXP mỗi cảnh giới gấp đôi cảnh giới trước (22/10/2026)**: `realm_mult` 1.2 → **2.0**. Tổng tới Đại Thừa 10: 28 288 → **694 960**.
- **Mật độ giảm 50% (22/10/2026)**: `exp_orbs.cell` 160 → **226** (≈160×√2, số viên giảm một nửa), `items.count_per_room` 400 → **200**.
- **Bot sinh ra tối đa Trúc Cơ (22/10/2026)**: `bots.realm_weights` [60,30,10] → **[60,40]**. Bot vẫn farm lên cao hơn sau khi sinh — đó là bình thường.
- **⚠ Spec gốc chuyển sang `docs/_raw/` (22/10/2026)**: chỉ chủ dự án sửa, Claude không ghi vào đó. `docs/GAME_SPEC.md` cũ đã xoá để khỏi có hai bản. Bản `_raw` hiện **thiếu 13 thứ đã code hôm nay** → danh sách đầy đủ ở **Q14 trong QUESTIONS.md**, cần dán lại vào `_raw`.
- **Tiên giới (22/10/2026, spec §3.4)**: 5 cảnh giới sau Đại Thừa (Chân Tiên → Đạo Tổ), không chia tầng, chỉ lên bằng **sát mạng** (mỗi cú chém trúng thân = 1, bot = 0.5, cùng nạn nhân 60s không tính, chỉ đếm từ Đại Thừa tầng 10). Bị chém rớt 1 cảnh giới và bộ đếm về mức khởi điểm. Level 90/100/110/120/130 nên `Realm(L) = L/10` vẫn đúng; `setTotalExp` không được đụng vào level của người Tiên giới. Hào quang sau lưng mỗi cảnh giới một màu, HUD đổi thanh Linh khí → Tiên khí vàng kim.
- **Kiếm tối đa theo công thức (22/10/2026, §3.1)**: `max_swords[r] = 5(r+1)(r+2)` → 10, 30, 60, 100, 150, 210, 280, 360, 450, 550, 660, 780, 910, 1050. Cảnh giới r có đúng r+1 vòng. `ring_step` đổi tên thành `ring_capacity_step` cho khớp spec gốc.
- **Bảng xếp hạng & danh hiệu (22/10/2026, spec v3.8)**: bảng phòng xếp theo **Uy Danh** (không còn theo cảnh giới), mỗi dòng ghi **danh hiệu** thay cảnh giới; cảnh giới+tầng chuyển lên **đầu nhân vật** (dưới danh hiệu, trên tên). Danh hiệu mở rộng Top 1→10, chữ nằm trong `config.titles` + `STR.titles` (phải sửa cả hai cho khớp).
- **Tin loan toàn server (22/10/2026, §11.4)**: hạ người đang Top 5 của phòng → kill feed gửi tới mọi phòng, hiện nền đỏ viền vàng. Ngưỡng ở `serverWideTopN` trong `room/broadcast.go`.
- **Bot sinh ngang hạng 11 (22/10/2026, §9.3)**: `bots.spawn_rank` = 11. Đủ người thì bot sinh bằng level người hạng 11; chưa đủ thì Luyện Khí. **Không còn trần cảnh giới** — `bots.realm_weights` đã bị xoá khỏi config vì hạng 11 tự nó là cái phanh.
- **artgen cắt sprite lệch (sửa 22/10/2026)**: vài pixel rác sót lại sau khi xoá nền (có ảnh chỉ 3 pixel ở góc) kéo khung nội dung ra cả ảnh → sprite bị thu nhỏ và dồn sang trái. `alphaBBoxDense` giờ bỏ các mảng rời dưới 1% mảng lớn nhất. 9 skin kiếm đã prep lại, căn giữa đúng.
- **artgen đếm sai ảnh đã có (sửa 22/10/2026)**: trước đây chỉ nhìn `assets/_raw/<job>.png`, mà thư mục đó là scratch nên bị dọn → `plan` báo "0 generated" và `run` sẽ vẽ lại toàn bộ 196 ảnh, đốt sạch quota. Giờ `alreadyHave` xét cả file đã prep trong `assets/`, đồng thời tôn trọng `redoReasons` (ảnh cần vẽ lại vì đổi luật vẫn tính là thiếu). Thực tế chỉ còn **22 ảnh** cần sinh: 9 skin kiếm + 3 nền (vẽ lại theo v3/v3.1) và 10 ref_sheet ưu tiên 9.
- **Thiêu Đốt Tinh Huyết (22/10/2026, spec v3.7 §4.1c)**: giữ Shift (PC) hoặc nút 🔥 (điện thoại) khi đang chạy → tốc độ ×1.8, mỗi giây mất 8% EXP của tầng hiện tại. Không rớt cảnh giới (dừng ở đáy cảnh giới), Luyện Khí đốt cạn về 0 vẫn sống. Có vòng lửa cho cả phòng thấy. Số để chỉnh: `burn.mult`, `burn.exp_frac_per_sec`, `burn.keep_realm`.
- **Va kiếm không đẩy văng nữa (22/10/2026)**: `clash_knockback` 150 → **0**. Hai người đứng mài nhau tới khi một bên hết kiếm. Đặt lại > 0 là bật lại kiểu cũ.
- **Tốc độ chạy (22/10/2026)**: `move_base` 260 → **130** và bỏ hệ số `sqrt(scale)`. Spec dán vào sau đó cứ lùi lại về `260 * sqrt(scale)`; đang giữ theo yêu cầu nhắn trực tiếp — xem Q12.
- **Tốc độ xoay kiếm (22/10/2026)**: `rev_secs` → **1.0** (1 vòng/giây theo spec v3.4), Tốc Kiếm trả về ×1.6.
- **Số kiếm tối thiểu (22/10/2026)**: `min_count` 1 → **0** (spec v3.5). Bị tước sạch kiếm thì không chạm được ai, hiện emote `no_swords`, bot chuyển sang chạy/farm. **Không chết** — luật chết vẫn chỉ áp cho Luyện Khí bị chém vào thân.
- **Spec v3.3–v3.6 đã code xong (22/10/2026)**: điểm Uy Danh + 3 bảng xếp hạng (Uy Danh/Sát Thần/Thần Tốc), danh hiệu Top 1–3 phòng, minimap, Ẩn Thân mờ 20%, clash_unit = 1 + chỉ đổi chiều khi cùng cảnh giới, damage vật cản theo người. Cổng dịch chuyển giờ ghép cặp 2 chiều cố định, `transit_secs` = 2s. Vật phẩm nhặt được bay hút vào người 0.22s.
- Schema SQLite thêm cột `score/hits/dai_thua_secs/top1_secs`, tự nâng cấp database cũ khi mở. API đổi `metric=exp|kills` → `board=score|hits|speed`.
- Môi trường dev: `~/go/pkg/mod` có một số thư mục thuộc quyền `root` (dấu vết của lần chạy `sudo go` trước) → `go build/test` báo "permission denied". Đang né bằng `GOMODCACHE=~/.cache/kx-gomod`. Sửa tận gốc: `sudo chown -R chinhdc:chinhdc ~/go/pkg/mod`. Ngoài ra `go` không còn trong PATH, đang dùng bản trong cache: `~/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.11.linux-amd64/bin/go`.

## Việc tiếp theo gợi ý
1. Đăng nhập lại minigen → chạy các đợt art ở trên.
2. Chơi thử nhiều tab / điện thoại qua LAN, ghi lại cảm giác cân bằng.
3. Deploy theo `docs/DEPLOY.md`.

## Staging (2026-10-22)
- Chạy tại https://kiemtien.staging.maveragroups.com (container `kiemtien-web` trên máy staging, Caddy chung chuyển tiếp vào).
- Deploy lại: `set -a; . docs/private/deploy; set +a; deploy/staging.sh` (build Go arm64 + client ở máy mình, rsync lên `~/kiemxoaytutien`, `docker compose up -d --build`, cài khối Caddy có kiểm tra trước khi reload).
- Bảng xếp hạng lưu ở volume `kiemtien_kiemtien-data`, không mất khi deploy lại. Server đang để `-debug=true` (người chơi vẫn phải mở `?debug=true`).

### Đo tải staging (2026-10-30)
- Browsertest qua Internet: `cd tools/browsertest && node run.mjs --url https://kiemtien.staging.maveragroups.com --tabs 7`
  → kết quả ghi `docs/perf/BROWSERTEST-staging.md` (không đè số đo trên máy). Mỗi lượt ghi file ngay khi xong.
- Chỉ chạy **7 tab**: staging giới hạn 8 kết nối mỗi IP, chừa 1 suất cho người chơi thật cùng mạng văn phòng
  (lần chạy 8 tab bị người thật chen vào → tab 8 ăn HTTP 429).
- Số đo: ping ~200 ms (tab thường) / ~250 ms (tab Đạo Tổ), xuống 7–25 KB/s mỗi người, tick server 2–3 ms.
  Container trên staging (ARM 4 nhân): CPU TB 28% / p95 40% / max 54% của 1 nhân, RAM tối đa 54 MB
  — cho 1 phòng 7 người thật + bot lấp đầy 60.
- Lỗi đã biết khi đo: nếu máy đo có container docker khởi động lại liên tục (vd. `mavera-sentry-worker` hỏng
  redis) thì Chrome báo `ERR_NETWORK_CHANGED` và tab rớt liên tục → dừng container đó trước khi đo.
  Chạy Chrome trong docker không giải quyết được vì không có GPU (vẽ bằng CPU, 800% CPU, tab 2 không mở nổi).
- Log staging có `tick chậm ~27 ms` do AI bot (25 ms) ở khu Trúc Lâm Mê Cung khi phòng không có người — cần xem.

## Đụng file giữa hai hướng (worktree song song, 30/10/2026)

Hai hướng chạy trên hai nhánh/worktree riêng (`kx-netcode` ← `netcode-2a`, `kx-anim` ←
`anim-v61`). Phân chia sở hữu để không giẫm chân:

| File | Chủ sở hữu | Ghi chú |
|---|---|---|
| `client/src/net/state.ts` | **A** | vòng đời entity, giữ 2 s |
| `client/src/render/scene.ts` — tạo/huỷ/pool/fade view | **A** | |
| `client/src/render/player.ts` — `setCharacter`, mờ vào | **A** | ⚠ **đụng**: hướng B sẽ sửa bộ điều khiển hoạt ảnh trong CÙNG file |
| `server/internal/room/bots.go`, `botnames.go` | **B** | tên bot §9.4 |
| `server/internal/config/config.go` — nạp `bot_names.txt` | **B** | |
| `tools/atlas-pack` | **B** | chưa đụng |

**`player.ts` là file duy nhất cả hai cần.** Hướng A đã merge trước (`1695b15`), hướng B
rebase sau — đúng thứ tự chủ dự án định. Khi hướng B làm bộ điều khiển hoạt ảnh, phần A
đã nằm sẵn trong `main` nên chỉ cần rebase, không phải gộp tay.

Thứ tự merge: **A (2a) ✅ → B1 tên bot ✅ → A (2b) → B hoạt ảnh** (sau khi chủ dự án duyệt
bảng điểm độ giật).

### Một cái bẫy khi chạy song song

Hai lượt `go test ./...` đỏ ở `TestRemovalReportedEvenWhenSlotReused` hoá ra **không phải
lỗi sản phẩm cũng không phải lỗi test**: tôi sửa `config/game.json` bằng `perl -i` trong
lúc test đang build, nên test đọc phải file viết dở (thiếu `pickup_r` → vòng nhặt 0 → đúng
triệu chứng "chưa nhặt được viên"). Đo lại dưới đúng điều kiện cũ: **0/4 lượt đỏ**.
→ Đừng sửa `config/` trong lúc test đang chạy.

---

## Phiên 30/10/2026 (chiều) — lỗi hình trên laptop, telemetry, điểm độ giật

### 1. Đội test mất hình trên laptop — đã sửa, đã deploy

Ảnh chụp của đội test lúc 12:18 cho thấy mặt đất và nhân vật mất hẳn texture, chỉ còn nét
viền trắng. Không phải lỗi mạng.

**Nguyên nhân**: `ground.filters` / `decor.filters` đặt bộ lọc màu lên container trải khắp
bản đồ 12 000×12 000 → PixiJS dựng render texture bằng cỡ container → **16 384² ×4 byte =
1 024 MB cho một tấm**. Laptop GPU tích hợp có `MAX_TEXTURE_SIZE` 8 192 hoặc thiếu VRAM nên
cấp phát hỏng. Máy dev có GPU rời nên nuốt được — vì thế không kịch bản S1–S5 nào bắt được.

**Sửa**: `filterArea` bám khung nhìn (bộ lọc chỉ đổi màu từng điểm ảnh, không lấy mẫu điểm
bên cạnh, nên cắt vùng cho kết quả y hệt). Kèm cắt art thừa cỡ sau khi ĐO cỡ vẽ thật.

| | Trước | Sau |
|---|---|---|
| Texture GPU lúc chơi | 1 076 MB | **102 MB** |
| Art tĩnh | 607 MB | 378 MB |

Giữ nguyên cổng/tường ở 512 px vì đo ra chúng vẽ 220–360 px — 512 là đúng cỡ. Nếu cắt theo
cảm giác "cái gì cũng to quá" thì đã làm rỗ hình ở mức thu phóng gần.

### 2. Hướng C — telemetry người chơi thật ✅ đã merge, đang thu trên `:9999`

Theo `TELEMETRY_SPEC` + `CHANGES_v6_2`. Đã xong: DB riêng (WAL, ghi theo lô ngoài vòng tick),
gói `tm` + `POST /telemetry/end`, server tự ghi tick/khu + CPU/RAM mỗi 10s, báo cáo 7 mục
§4.1 (`perfreport --real`), trang `/admin/perf` (LAN + `KX_ADMIN_PASS`), tổng hợp tự động
12:00 & 23:30 ra `docs/perf/real/`, S6 trong PERF_LOG + cột `n_sessions` trong CSV.

Client nằm sau cờ `TELEMETRY=on` + `import()` động; `npm run check:telemetry-off` kiểm bản
release không còn `sendBeacon` / `WEBGL_debug_renderer_info` / `/telemetry/end` / `longtask`.

**Lệnh deploy đang dùng** (cần cả hai, thiếu cờ là client không thu):
```
cd client && TELEMETRY=on BUILD_ID=$(date +%Y%m%d_%H%M)_$(git rev-parse --short HEAD) npm run build
cd ../server && go build -o kxserver ./cmd/server
KX_ADMIN_PASS=... ./kxserver -build "<build id>"
```

### 3. Hướng B2 — điểm độ giật ✅ xong, **ĐANG CHỜ CHỦ DỰ ÁN DUYỆT**

Bảng điểm: `docs/JITTER_REPORT.md`. Phím `J` (`?debug=true`) hiện tên anim · frame · pha ·
điểm giật trên đầu từng nhân vật. **Chưa đụng vào hoạt ảnh** — đúng CHANGES_v6_1 mục 5.

Ba câu hỏi đang chờ anh trả lời nằm ở cuối `JITTER_REPORT.md` (ngưỡng đỏ, có đồng ý tách hai
bảng không, và thứ tự C→B hay B→C).

## Lỗi đã biết còn mở

- **Nhấp nháy + placeholder CHƯA đạt ngân sách trên server thật.** PERF_LOG lượt #5 từng ghi
  "0 ✅" — sai, vì đo trên server test yên tĩnh. Trên `:9999`: respawn 3/0/1/1/0 và
  placeholder 252/0/24/4/0 qua 5 lượt `flicker.mjs`; telemetry xác nhận 12 khung placeholder
  và 13.5 respawn/phút. Nút thắt B3 **mở lại** → việc của hướng A bước 2b.
- Dòng S6 trong PERF_LOG (#7) là **lượt kiểm chứng của agent**, không phải người thật (N=1,
  tab headless). **Dòng S6 tự động đầu tiên đã có** (`docs/perf/real/2026-10-30_2330.md`,
  N=68 phiên / 57 người) và đã chép vào PERF_LOG §3 — xem dòng không số, ngay trên 9a.

## Việc tiếp theo

1. **Hướng A bước 2b** (đã có bằng chứng số thúc đẩy): View Pool, mờ ra 150 ms, `DescRequest`,
   cấm placeholder ở build thật, `check:assets`.
2. Hướng B hoạt ảnh (mục 6–8) — **chỉ sau khi anh duyệt bảng điểm**.
3. ~~Chép dòng S6 tự động đầu tiên vào PERF_LOG §3~~ — **xong 31/10**.

## Phiên 30/10 (chiều, tiếp) — item Bí Kíp, hoạt ảnh B + C

### Spec v6.3 anh vừa chốt
Ngưỡng giật 45 giữ tạm · tách 2 nhóm (lặp đọc TB, một lần đọc "lạc lõng") · thứ tự **B → C → A**.
Kèm yêu cầu mới: phím `[` / `]` xem lần lượt từng animation — **đã làm**.

### Bí Kíp hiện ra placeholder
`items/skill_scroll.png` có từ lâu; `itemTexture` đoán đường dẫn `buffs/<loại>.png` nên tra
trượt. Sai trong im lặng — không có ảnh thì vẽ placeholder, nhìn hệt art chưa xong.
Thêm `npm run check:items` để lần sau máy báo.

### B — chuẩn hoá frame ✅
`tools/atlas-pack` → `artgen normalize`. Kết quả: anim lặp **20.5 → 15.7 (−24%)**, một lần
**4.8 → 2.0 (−57%)**, đỏ 3+2 → **1+0**, không anim nào tệ đi.

Ba chỗ spec nguyên văn không chạy được trên art thật, đều có số:
1. B.2 áp thẳng ra co giãn 896% / dịch 4 116 px (vì `death` cố ý biến hình) → trần ±15%.
2. Neo chân là phép của nhân vật hình người; rùa/cua/gà tệ đi gấp đôi → **chấm cả hai bản,
   giữ bản tốt hơn** (25 anim tự bị bỏ qua).
3. "Đầu trôi > 4 px" làm tiêu chí sinh lại ra 102/149 → dùng tiêu chí của anh (còn đỏ sau B).

**Lỗ hổng đã vá**: bộ dò cắt-sai chỉ ĐẾM frame nên bỏ lọt `co_ba_ganh idle` (1 frame chứa cả
bốn nhân vật, 3 frame là mảnh vụn 5–12 px) — chính là animation tệ nhất toàn bộ art.
Thêm `warnFrameSizes`. Sinh lại xong giờ sạch.

### C — bộ điều khiển hoạt ảnh ✅
`render/animctl.ts`: crossfade 100 ms, giữ pha run↔chase↔flee, nhịp chân theo tốc độ thật,
lật có co ngang, nghiêng khi rẽ, nhún/squash/thở. Tốc độ lấy từ vị trí ĐÃ NỘI SUY.
**Chưa đo bằng số được** (đây là cảm giác) — cần anh chơi thử và nhận xét.

### Genart
Sinh lại 23 sheet hỏng qua 2 lượt. Miscut **15 → 5** (5 tile còn lại cứ sót nền magenta sau
2 lần sinh; đã vá tạm bằng màu trung bình nên trong game không thấy vệt hồng).
Thêm job `item_bundle` — chờ anh chốt (xem QUESTIONS.md).

### Còn mở
- **A (một thân nhiều mặt)** — cần art mới, làm sau cùng theo thứ tự anh chốt.
- **Hướng A netcode bước 2b** — View Pool, mờ ra, `DescRequest`, cấm placeholder build thật.
  Vẫn là nút thắt B3 đang mở, có số đo thúc đẩy.
- `dep_lao idle` 56.8 còn đỏ — nhiều khả năng là hoạt ảnh đảo mắt cố ý, cần anh nhìn bằng `J`.

## Phiên 30/10 (chiều, tiếp 2) — trả lời của anh về B+C, rồi netcode 2b

Năm việc anh giao trong `ANSWERS.md` "Trả lời báo cáo B + C" — **xong cả năm**.

1. **Push GitHub** ✅ — `main` đã lên tới `9bca2c9`, kèm `docs/anim_preview/`.
2. **Bó Kiếm** ✅ — prompt lấy nguyên văn ART_PROMPTS §8 (bó **5** kiếm, ruy băng đỏ, hào
   quang vàng), vẽ to hơn bong bóng kiếm đơn 20%, bó rơi từ xác in số kiếm lên bong bóng.
3. **Danh sách ngoại lệ** ✅ — `assets/jitter_exceptions.json`, hiện **rỗng** (chưa thêm
   `dep_lao` như anh dặn). Anim được duyệt không tính đỏ nhưng **vẫn giữ điểm**, có 3 test canh.
4. **CI** ✅ — `.github/workflows/ci.yml`: `go vet`, `go test -race`, `check:assets`,
   `check:items`, build, `check:telemetry-off`.
5. **Netcode 2b** ✅ (phần placeholder) — xem dưới.

### Netcode 2b — placeholder đạt ngân sách, respawn còn hở

Đo trên `:9999` bằng `flicker.mjs`, **16 lượt**:

| | trước 2a | sau 2a | sau 2b |
|---|---|---|---|
| Khung placeholder | 462/596 | 252 / 0 / 24 / 4 / 0 | **0 ở cả 16 lượt** ✅ |
| Respawn trong 2s | 38 | 3 / 0 / 1 / 1 / 0 | 2/16 lượt có đúng 1 |

**Gốc rễ của placeholder là `DescRequest`**: tên + nhân vật chỉ gửi LẦN ĐẦU, server nhớ bằng
`knownPlayers`. Client mất mô tả đó (vứt view, rớt gói) thì server **không bao giờ** gửi lại
→ client kẹt vĩnh viễn với một entity không biết vẽ gì. Giờ client hỏi, server quên đúng
những id được hỏi.

**Respawn chưa về 0 và sẽ không về 0 ở bước 2b** — client không phân biệt được "vắng mặt vì
AOI" với "đã rời đi thật". Spec nói rõ 2b chỉ là bản vá rẻ; cần `Despawn` tường minh từ
server ở **bước 3**.

### Còn mở
- **Netcode bước 3** — `Despawn` tường minh, để respawn về 0 hẳn.
- **A (một thân nhiều mặt)** — cần art mới, sau cùng theo thứ tự anh chốt.
- `dep_lao idle` 56.8 — chờ anh xem bằng `J` + `]` rồi quyết có thêm vào ngoại lệ không.
- 5 tile vẫn sót nền magenta sau 2 lượt sinh lại (đã vá tạm bằng màu trung bình).
- **Lưu ý khi commit**: dùng `git commit -F <file>` thay vì `-m "..."` — dấu backtick trong
  message bị shell diễn giải, commit `9bca2c9` mất vài từ vì lỗi này.

## Phiên 30/10 (chiều, tiếp 3) — NETCODE bước 3: nền tảng Protocol v2 (§4.9)

Spec bắt buộc làm §4.9 **trước** khi chạm vào message. Đã xong phần nền đó.

### Một nguồn schema + generator hai phía
`proto/schema.txt` là nguồn sự thật duy nhất. `cd server && go run ./cmd/protogen` sinh ra
**cả hai**: `server/internal/wire/wire_gen.go` và `client/src/net/wire.gen.ts`. Không ai viết
`DataView` bằng tay nữa.

Định dạng dòng tự viết thay vì YAML: dự án có luật không thêm thư viện khi chưa hỏi, spec chỉ
nói "vd schema.yaml" chứ không chỉ đích danh thư viện. Định dạng này còn dễ đọc hơn YAML.

6 message theo §4.3: `Spawn` `Desc` `Update` `Despawn` `SelfState` `Input`.
`Update` dùng bitmask — entity đứng yên thì gói chỉ còn **6 byte**.

### Bài đối chiếu hai chiều
```
cd server && go run ./cmd/protofixture -o /tmp/wire.json
cd client && npm run check:wire -- /tmp/wire.json
```
**10 015 gói** (15 golden nhắm vào biên + 10 000 fuzz seed cố định), kiểm **hai chiều**:
client giải mã byte của server rồi mã hoá ngược lại, so từng byte. Hiện **khớp hết**.

Đã kiểm chính bài kiểm: đổi 1 byte trong fixture → báo đúng gói, đúng trường, thoát mã 1.

### `schema_hash` chống đọc rác
Băm nội dung schema nhúng vào cả hai phía. Lệch là từ chối, thay vì đọc rác im lặng — đúng
lỗi đã dính một lần với trường `mood`.

### CI
Thêm 2 bước: sinh lại từ schema rồi `git diff --exit-code` (ai sửa schema mà quên chạy
generator thì CI đỏ), và chạy bài đối chiếu với fixture **sinh tươi** mỗi lần.

### Còn lại của bước 3
Phần nền xong, chưa động vào đường truyền thật:
- Vòng đời `Spawn/Desc/Update/Despawn` tường minh ở server (đây mới là thứ đóng nốt respawn).
- Delta theo ack + keyframe 2s.
- Lượng tử hoá §4.6 (vị trí mình giữ i32 fixed-point, người khác u16).
- Bật v2 sau cờ `proto=2`, v1 chạy song song.

## Phiên 30/10 (tối) — spec v6.4 phần A: dung lượng tải & màn hình tải (§13.7)

**Kết quả một dòng: gói chặn người chơi vào trận 77,4 MB → 7,02 MB (−91%), vào trận lần đầu
ở 4G 10 Mbps 8,3 giây. Cả hai ngân sách §13.7 ĐẠT.** Số đầy đủ: PERF_LOG lượt #9.

### Đo trước đã chỉ ra nguyên nhân, và nó không phải "ảnh quá nặng"

`npm run assets:size` (mới viết, §13.7 mục 1) đọc ra **77,4 MB / 547 tệp tải trước khi hiện
trang chủ** — vì `AssetStore.preload()` tải *mọi* tệp .png trong manifest, gồm cả bộ art của
cả 10 nhân vật và vật cản của cả 5 khu bản đồ, chỉ để hiện một ô nhập tên.

Con số đó quan trọng vì nó đổi hẳn thứ tự việc. Nén ảnh một mình thì không đủ: 77,4 MB giảm
84% vẫn là 12 MB chặn người chơi, vẫn vượt ngân sách 8 MB. **Việc lớn nhất là chia nhóm tải,
nén chỉ là việc thứ hai.** Nếu chỉ đi nén ảnh thì đã làm xong 84% công việc mà vẫn không
đạt, và sẽ kết luận là "ngân sách spec quá chặt".

### Làm gì

1. **`assets/src` (nguồn) / `assets/dist` (phát hành)** — §13.7 mục 2. artgen ghi vào `src`;
   server phục vụ `dist`. `dist` đã gitignore: nó là bản nén của thứ đã nằm trong repo.
2. **`tools/asset-pipeline`** (Node + `sharp`, spec đã duyệt): PNG → WebP q72–82, thu nhỏ về
   kích thước hiển thị, tên tệp kèm hash nội dung, `manifest.json` v2 có **byte + nhóm tải**
   từng tệp. Có bộ đệm theo vân tay (nguồn + tham số xử lý): lần đầu 23 s, chạy lại 0,2 s —
   cần thiết vì `npm run build` gọi nó mỗi lần.
3. **Nhóm tải** `home / boot / zone_<khu> / char_<nhân vật> / fx_high / audio / debug / legal`.
   Chỉ `home` chặn trang chủ (0,86 MB); `boot + char_mình + zone_mình` chặn vào trận.
   Nhóm khu **suy ra từ `maps/*.tmj`** (loại vật cản thật của từng khu) chứ không gõ tay —
   `type` của đối tượng Tiled đúng bằng tên thư mục `obstacles/<kind>/`, nên không lệch được.
4. **Server**: tệp có hash → `Cache-Control: immutable` 1 năm, `manifest.json` → `no-cache`,
   gzip cho JS/JSON/CSS/font (WebP/OGG bỏ qua — nén lại chỉ tốn CPU để tệp to hơn).
5. **Màn hình tải** nằm trong `index.html` với CSS nội tuyến nên hiện từ byte đầu tiên, trước
   cả khi bundle JS về. Thanh **theo byte** (`%` + `x,x / y,y MB`), dòng trạng thái theo đúng
   4 bước của spec, **nhân vật mình chọn chạy trên thanh** (chân dung trước, đổi sang khung
   `run` thật cắt từ atlas bằng `background-position` khi atlas về), 14 mẹo/câu bựa đổi mỗi
   4 s, nút "Thử lại" **tải tiếp** phần còn thiếu. Icon `⬇ %` góc dưới cho nhóm tải nền.
6. **`npm run assets:budget`** chặn build khi vượt (boot ≤ 8 MB, tổng ≤ 80 MB), in 20 tệp
   nặng nhất *của đúng gói bị vượt*. Đã vào `npm run build` và CI.

### Ba thứ chiếm phần lớn mức giảm — hai trong ba không phải nén ảnh

1. **Bỏ 5 loại tệp không mã nào đọc khỏi bản phát hành (−129 MB nguồn)**: `frames/`,
   `frames_norm/`, `ref_sheet.png`, và `body/` + `faces/` — art của **hướng A, chưa có mã nào
   dùng**. Cả 5 đều nằm trong `manifest.json` cũ và người chơi **tải thật**. `ref_sheet` còn
   bị bộ nạp lọc ra lúc tải nhưng vẫn tính vào manifest — tức là tải xong rồi bỏ.
2. **Chia nhóm tải (−64 MB khỏi gói chặn)**.
3. **PNG → WebP + thu nhỏ (72,2 → 11,4 MB, −84%)**.

### Ba chỗ spec áp nguyên văn không ra kết quả, đều có số đo

1. **"Frame nhân vật lưu ở 256 px"** — atlas hiện dùng ô **128 px**, tức đã nhỏ hơn mức spec
   yêu cầu. Thu nhỏ nữa là làm nhân vật mờ đi mà không đổi lấy được byte nào (WebP đã lấy
   hết). Nên atlas chỉ đổi định dạng, không thu nhỏ.
2. **Âm thanh 96/128 kb/s** — nguồn đã là 22–44 kHz mono ở 31–80 kb/s, tức **dưới** mức spec.
   Ép libvorbis lên 128 kb/s ở 22 050 Hz thì nó báo `encoder setup failed`, và nếu chạy được
   cũng chỉ làm tệp to hơn. Pipeline **đo bitrate trước rồi mới quyết định nén**.
3. **Trim viền trong suốt + ghép atlas item/icon** — **chưa làm, cố ý**. Trim một sprite lẻ
   là dịch điểm neo của nó, mà hơn 300 chỗ gọi `assets.texture()` đặt anchor theo quy ước
   riêng; với atlas thì phải ghép lại và thêm `trim`/`orig` cho từng khung. Phần thưởng ước
   tính 5–15% vì WebP đã nén vùng alpha rỗng gần như miễn phí, mà **cả hai ngân sách đã
   ĐẠT**. Theo quy tắc "đạt thì dừng". Cần tới khi gói chặn tiến sát 8 MB.

### Rủi ro mới sinh ra, và chỗ nó sẽ cắn

Tên tệp giờ có hash → **mọi chỗ ghép tay `/assets/...` là 404 trong im lặng**. Đã chuyển 8
chỗ sang `assets.url()`: `<img>` chân dung / huy chương / hồn ma, nền trang chủ, icon buff,
font, âm thanh, và `border-image` của màn Tổng kết (chỗ cuối trước đây nằm trong `style.css`
và vite vẫn cảnh báo "didn't resolve at build time" — nó chạy được chỉ vì server phục vụ
thẳng thư mục nguồn). Luật này đã ghi vào `CLAUDE.md`.

### Tự kiểm bằng trình duyệt thật

`go vet` · `go test ./...` 256 test · `npm run build` — pass. Chạy thật trong Chromium:
trang chủ **11 chân dung thật, 0 placeholder**; trong trận **0 placeholder**, 58 fps, vẽ
0,6 ms, 36 draw call, **0 request HTTP ≥ 400, 0 lỗi console**; `/assets/.../frames/...` trả
404 đúng như phải thế. Header kiểm riêng: WebP có hash → `immutable`, font → `immutable` +
gzip, `manifest.json` → `no-cache`.

## ~~Lỗi đã biết còn mở~~ ĐÃ SỬA (ghi 30/10 tối · đóng 1/11)

> **Cả hai lỗi dưới đây đã sửa xong.** Giữ lại phần mô tả vì nó giải thích vì sao
> `BIOME_TILES` đánh khoá bằng tên biome chứ không bằng tên khu — người đọc sau mà không
> biết chuyện này rất dễ "sửa lại cho dễ đọc" và làm hỏng lần nữa.
>
> - Nền 5 khu: `BIOME_TILES` nay đánh khoá `grass`/`dirt`/`concrete`/`snow`, khớp đúng
>   `map.biome`. Xem `client/src/render/catalog.ts`.
> - Trang trí: mapgen chỉ chọn decor có trong manifest, mỗi decor gắn tag khu.
>
> Câu hỏi tương ứng đã trả lời — xem `docs/QUESTIONS_DONE.md`.

Hai chỗ **art có trong repo mà game không bao giờ vẽ**. Cả hai có từ trước phiên đó, tìm ra
trong lúc chia nhóm tải (phải trả lời "tệp này của khu nào" mới lộ ra).

1. **Nền 5 khu đều đang là cỏ.** `mapview.buildGround()` tra `BIOME_TILES[map.biome]`, nhưng
   `BIOME_TILES` đánh khoá bằng **tên khu** (`'Linh Sơn'`, `'Trúc Lâm'`, `'Phàm Trấn'`,
   `'Hắc Thị'`) còn `map.biome` là **tên biome** (`'grass'`, `'dirt'`, `'concrete'`,
   `'snow'`). Tra luôn trượt → rơi về `['grass']`. Hậu quả: Băng Nguyệt Hồ, Linh Sơn Cốc,
   Phàm Trấn đều trải cỏ; **14 trong 19 ảnh nền chưa từng hiện ra** (0,43 MB vẫn nằm trong
   nhóm `boot`). `tiles/snow.png`, `dirt.png`, `concrete.png` có sẵn và tên khớp đúng giá trị
   `biome` — nên có thể chỉ cần đổi khoá của `BIOME_TILES`.
2. **Không vật trang trí nào được vẽ.** `buildDecor()` tìm `decor/<d.sp>.png` với `d.sp` lấy
   nguyên từ `type` trong Tiled: `rock_small`, `pine`, `lantern`, `well`, `banner`, `bamboo`,
   `bush_small`. `assets/src/decor/` lại chứa `red_plastic_stool`, `old_bicycle_leaning`,
   `street_food_cart`, `banana_tree`… — **không tên nào trùng**, nên `continue` ở mọi vật, và
   299 vật trang trí trong 5 bản đồ không hiện gì cả. `npm run check:assets` không bắt được
   vì nó chỉ đối chiếu vật cản, không đối chiếu decor.

~~Hỏi anh: hai bộ tên decor này là đổi mapgen cho khớp art đã có, hay vẽ thêm art theo tên
mapgen đang dùng?~~ → **Đã trả lời và đã sửa** (Q27, xem `QUESTIONS_DONE.md`): mapgen chỉ
chọn decor có trong manifest.

## Phiên 31/10 — trả lời Q27 + trang chủ hiện ngay

### Q27-1 ✅ nền 5 khu — và chẩn đoán hôm qua của tôi SAI một nửa

Hôm qua tôi báo "cả 5 khu đều trải cỏ". Sai. Có **hai** bảng cùng tên `BIOME_TILES` và chúng
đã trôi khỏi nhau; chỉ bảng trong `mapview.ts` là bảng game thật sự dùng:

| Bảng | Khoá | Ai đọc | Hậu quả |
|---|---|---|---|
| `mapview.ts` | `grass` `market` `concrete` `dirt` `rock` | **bộ vẽ** | `market`/`rock` là khoá chết; **thiếu `snow`** → chỉ Băng Nguyệt Hồ trải cỏ. Và bộ dành cho hẻm núi nằm dưới khoá chết `rock`, nên Linh Sơn Cốc nhận `dirt_path` + `village_grass` — cỏ làng trong hẻm núi đá |
| `catalog.ts` | `'Linh Sơn'` `'Trúc Lâm'` `'Phàm Trấn'` `'Hắc Thị'` | `assets:report`, `check:assets` | đòi sinh những tile game không bao giờ tra tới |

Hai bảng cùng tên cho cùng một việc là gốc rễ; sửa khoá mà để nguyên hai bảng thì sẽ trôi
lại. Nên **gộp còn một bảng** ở `catalog.ts` (nơi Node đọc được, không import Pixi), đánh
khoá bằng giá trị `biome` thật trong `.tmj`, và `mapview` nhập từ đó.

Sửa luôn hai lỗi cùng họ trong chính hàm `buildGround()`:
- `BIOME_COLOR` (màu nền khi chưa có art) được tra bằng **tên tệp tile** thay vì biome →
  khớp hay trượt là chuyện tình cờ. Giờ tra thẳng `map.biome`, thêm `snow`.
- Mặt nạ trộn gieo theo `biome`, mà Thiên Đình và Trúc Lâm cùng `grass` → hai khu có mảng
  loang **giống nhau từng vệt**. Gieo theo `map.id`.

Nền từng khu sau khi sửa (`npm run check:assets` in ra mỗi lần chạy):

| Khu | biome | Nền |
|---|---|---|
| Thiên Đình | grass | grass, soft_grass, bamboo_litter |
| Linh Sơn Cốc | dirt | dirt, **rocky_ground, fine_gravel** |
| Trúc Lâm Mê Cung | grass | grass, soft_grass, bamboo_litter |
| Phàm Trấn | concrete | concrete, worn_cobble, **dusty_earth** |
| Băng Nguyệt Hồ | snow | **snow, ice** |

**Nhờ anh nhìn 3 khu in đậm** — Linh Sơn Cốc, Phàm Trấn, Băng Nguyệt Hồ đổi nền hẳn.

Còn mở: Thiên Đình và Trúc Lâm **cùng** `biome: "grass"` nên trải cùng bộ nền. Muốn Trúc Lâm
có nền tre riêng (`bamboo_floor` chưa dùng) thì đổi `zoneDefs` cho `truc_lam` sang một biome
riêng rồi sinh lại bản đồ — đổi dữ liệu bản đồ nên chờ anh.

### Q27-2 ✅ trang trí: tag khu + chỉ đặt cái có art

`cmd/mapgen/decor.go` mới là nguồn duy nhất:
- `decorZones`: **mỗi vật trang trí gắn tag khu** (`red_plastic_stool → pham_tran, thien_dinh`…).
- `haveDecorArt()` **đọc đĩa** `assets/src/decor/` chứ không gõ danh sách vào mã — art về là
  mapgen dùng ngay, không phải sửa thêm dòng nào.
- Chốt chặn trong `builder.decor()`: **không đặt vật không có art**, chặn ở nơi ĐẶT nên mọi
  đường gọi đều an toàn, kể cả ba hàm sinh bản đồ cũ gọi thẳng vào đó.
- Mapgen **in ra** loại nào chưa đặt được vì thiếu art, kèm đúng tên job artgen.

**Mật độ là TỔNG của khu, không phải mỗi loại.** Bản đầu tôi để "mỗi loại n lần" và Phàm Trấn
nhảy lên 447 vật chỉ vì nó hợp 9 loại — khi 7 ảnh còn thiếu về thì nó tự lên ~720. Mật độ bản
đồ không được phụ thuộc vào việc hôm nay có bao nhiêu tệp art.

Số trang trí sau khi sửa: 81 / 107 / 131 / 139 / 91 — **gần y như trước** (75/107/131/139/91),
chỉ khác là **giờ chúng vẽ ra thật**. Đối chiếu từng bản đồ: ngoài lớp `decor`, obstacles /
zones / gates / altars **giống hệt** bản cũ.

Art còn thiếu: 7 ảnh (`rock_small`, `pine`, `bamboo`, `bush_small`, `lantern`, `well`,
`banner`) → job artgen mới **`decor_nature`**. Chưa sinh được vì tài khoản minigen đang hết
hạn cookie. Khi nào anh đăng nhập lại: `cd server && go run ./cmd/artgen run -only decor_nature`
rồi `go run ./cmd/mapgen`.

**Bốn test mới** (`cmd/mapgen/decor_test.go`), và cái thứ hai đã bắt lỗi thật ngay:
1. `TestDecorPlacedAlwaysHasArt` — 5 khu × 10 seed, mọi vật đặt ra đều có ảnh.
2. `TestEveryZoneGetsSomeDecor` — mỗi khu phải đặt được ít nhất MỘT loại. **Test này đỏ ngay
   lần đầu**: Băng Nguyệt Hồ chỉ gắn `rock_small` + `pine`, cả hai chưa có art → khu đó trống
   trơn. "Chỉ đặt cái có art" có một cách thoả mãn rất tệ là không đặt gì cả, và chỉ test này
   phân biệt được. Đã gắn thêm `stone_lantern` cho khu đó.
3. `TestDecorZonesNamesAreKnown` — tên khu trong bảng phải tồn tại; gõ sai là vật đó im lặng
   không bao giờ được đặt ở đâu.
4. `check:assets` thêm hai phép kiểm đọc **từ bản đồ**: mỗi `.tmj` phải tra ra ≥ 1 ảnh nền có
   thật, và mọi decor bản đồ ĐẶT RA phải có ảnh.

### Trang nhập tên hiện ngay, tải gói trong lúc gõ tên ✅

Trang chủ không còn chờ asset nào. Ảnh chân dung và nền là `<img>` / `background-image` của
DOM nên **trình duyệt tự tải lấy** — chỉ cần `manifest.json` để biết địa chỉ, không phải đi
qua bộ nạp Pixi. Vậy trang chủ chỉ phụ thuộc hai lời gọi nhỏ: `/api/config` + `manifest.json`.

Mọi thứ còn lại (`home`, `boot`, font, atlas của nhân vật đã lưu) tải **nền** trong lúc người
chơi gõ tên và ngắm 10 thẻ kỹ năng — khoảng thời gian trước đây bỏ không.

Bấm "Vào chơi": `assets.pending()` rỗng thì **không hiện thanh tiến độ** và vào thẳng. Hiện
một thanh nhảy từ 0 lên 100% trong một khung hình chỉ làm người ta giật mình.

Hai chỗ phải cẩn thận:
- **Font vẫn chặn vào trận**, chỉ không chặn trang chủ. Bảng tên vẽ bằng `BitmapText`, thiếu
  font là nó ném lỗi giữa trận. Font bắt đầu tải từ lúc mở trang nên hầu như luôn xong sẵn.
- Vá ảnh trang chủ **tại chỗ**, không dựng lại màn hình: người chơi đang gõ tên, dựng lại là
  xoá mất chữ họ vừa gõ.

Số đo (trung vị 3 lượt, bóp mạng bằng CDP, 6 giây gõ tên): xem PERF_LOG lượt #10.

## Phiên 31/10 (tiếp) — phần B Tiên khí ×20, genart, hai lỗi trang chủ

### Hai lỗi trang chủ chủ dự án chụp màn hình — CẢ HAI LÀ LỖI TÔI VỪA GÂY RA

**1. Nút "Bảng xếp hạng" mất hẳn chữ.** Hôm qua tôi thêm `:root { color-scheme: dark; }`
vào `index.html` cho màn hình tải. Luật đó rò ra toàn trang: thẻ `<button>` lấy màu chữ từ
**bảng mặc định của trình duyệt** (`color: ButtonText`), không thừa kế từ cha — nên ở chế độ
tối nó hoá **trắng**. Mọi nút của game đặt `background: #fff` mà không đặt `color`, thành
chữ trắng trên nền trắng. `font: inherit` không che được `color`.

Sửa hai lớp: `color-scheme` thu về `#boot` (chỉ màn hình tải), **và** `.btn` / `.tab` /
`.char-card` đặt `color: var(--ink)` tường minh. Lớp thứ hai mới là lớp chống tái phát —
nút không được phép phụ thuộc vào `color-scheme` của trang.

**2. Phần trên panel bị cắt, không thấy ô nhập tên.** Nội dung trang chủ cao **630 px**, mà
`.panel { max-height: 100% }` ở cửa sổ 561 px chỉ còn 529 → phần thừa trôi vào thanh cuộn
**bên trong** panel, gần như không nhìn thấy. Cuộn được không có nghĩa là dùng được: thứ đầu
tiên phải làm là nhập tên, mà nó chính là thứ bị đẩy khỏi màn hình.

Thêm hai mức cửa sổ thấp (`max-height: 720px` và `560px`): bóp lề, thu ảnh nhân vật, ẩn phụ
đề / dòng hướng dẫn / mô tả kỹ năng — **không bao giờ ẩn ô nhập tên và nút "Vào chơi"**.
Đo lại: 650×561 · 650×480 · 1280×800 · 390×640 đều **vừa khít, không phải cuộn**. Còn
844×390 (điện thoại xoay ngang) vẫn thừa 78 px — ghi vào lỗi đã biết.

### Phần B — Tiên khí ×20 (CHANGES_v6_4 mục 8–11) ✅

- `tien_khi_ceilings` → **[2 000, 6 000, 12 000, 20 000, 30 000]**. Đáy: 1 / 2 001 / 6 001 /
  12 001 / 20 001. Toàn bộ `progression.go` vốn đã suy ra từ config nên không phải sửa dòng nào.
- `burn.tien_frac_per_sec` **0.08 → 0.004**. Đây là chỗ dễ sai nhất của lượt ×20: đổi trần mà
  quên đổi tỉ lệ thì một giây chạy nhanh tốn cả trăm viên. Đốt vẫn **8 · 16 · 24 · 32 · 40**
  mỗi giây, đúng lượng tuyệt đối cũ.
- **Mệnh giá viên 500 / 100 / 10 / 2 / 1, tối đa 40 viên** (`splitTienOrbs`). Với bộ cũ
  {10, 2, 1} thì 2 099 điểm rơi ra là **210 viên** quanh một cái xác — một rừng vật phẩm và
  210 entity nhồi vào gói snapshot. Vượt trần thì **gộp** phần nhỏ, nên tổng điểm luôn đúng.
- Sprite **viên 100 và 500** (`items/tien_orb_100/500.png`) + `tienKhiRadius` 5 mức. Viên
  1/2/10 vẫn vẽ bằng Graphics — trên màn hình chúng bé bằng hạt đậu.
- Thanh HUD có dấu phân cách nghìn: `Tiên khí 9.999/10.000`.
- Test: sửa 14 chỗ trong `burn_tien_test.go` + `altar_test.go` theo checklist §17 mới, thêm
  3 test cho `splitTienOrbs` (tổng đúng · ví dụ 2 099 ra đúng 18 viên · trần 40 chặn thật).

**⚠ Hệ quả cân bằng, nhờ anh quyết** — spec nói rõ "nguồn Tiên khí từ đài giữ nguyên", nên
tôi KHÔNG đụng vào, nhưng con số đáng nhìn: 2 Đài Tiên Khí của một khu phun **5,3 điểm/giây**
cộng lại. Một người ôm trọn cả hai đài, không ai tranh:

| Mốc | Thời gian |
|---|---|
| Chân Tiên → Kim Tiên (2 000) | ~6 phút |
| Lên tới Đạo Tổ (20 001) | **~62 phút** |
| Nạp đầy 30 000 để đổi Uy Danh | ~94 phút |

Trước ×20 thì tới Đạo Tổ chỉ ~3 phút. Nguồn thứ hai là 30% Tiên khí rơi ra khi chém người
Tiên giới — nhưng nguồn đó chỉ chảy *sau khi* đã có người lên cao, nên nó không rút ngắn
người đầu tiên. Nếu anh muốn Đạo Tổ nằm trong tầm một phiên chơi thì phải tăng đài
(`altars.small.orb_value` hoặc `burst_*`), và đó là con số cân bằng của anh.

### genart ✅ — và một lỗi đường dẫn tôi gây ra hôm qua

`artgen status` thoát lỗi `assets/docs/ASSETS_STATUS.md: no such file`. Nguyên nhân: nó lấy
thư mục `docs` bằng `filepath.Dir(assets)` — đúng khi art ở `<repo>/assets`, nhưng sau khi
art chuyển sang `<repo>/assets/src` (§13.7) thì ra `<repo>/assets/docs`. Ba chỗ, đã chuyển
sang `config.FindRepoFile("docs/...")`.

**Sinh xong 9 ảnh**: 7 vật trang trí (`rock_small` `pine` `bamboo` `bush_small` `lantern`
`well` `banner`) + 2 viên Tiên khí (100, 500).

**Lượt đầu hỏng theo kiểu tệ nhất và đáng ghi lại**: tôi gộp 7 vật vào MỘT sheet 4+3. Bên vẽ
vẽ đủ cả 7, nhưng bộ cắt tìm khung bằng **vùng pixel liền nhau** nên cây thông chạm bụi tre
bị đếm thành một khung — ra 5 khung cho 7 tên, và mọi tên sau chỗ dính bị **gán lệch**:
`bush_small.png` hoá ra là cái đèn lồng, `lantern.png` là cái giếng. Art sai nhãn tệ hơn hẳn
art thiếu: `check:assets` thấy đủ tệp nên báo xanh, và bản đồ rải đèn lồng ở chỗ lẽ ra là bụi
cây. Đã xoá 5 ảnh đó và chuyển sang **mỗi vật một job sprite riêng** — một vật một ảnh thì
không có gì để cắt sai.

Sinh lại bản đồ: giờ mỗi khu có bộ trang trí đúng chất của nó — Linh Sơn Cốc đá + thông,
Trúc Lâm tre + bụi, Phàm Trấn cả bộ phố + giếng + băng-rôn, Băng Nguyệt Hồ thông + đá + đèn
đá. Tổng số vật giữ nguyên (81/107/131/139/90) vì mật độ là **tổng của khu**.

## Phiên 31/10 (tiếp 2) — v6.5 kỹ năng theo U, và ba lỗi hình ảnh/di động

### CHANGES_v6_5 ✅ — phạm vi kỹ năng theo U

`U = player.pickup_r × scale_người_dùng` — đúng bán kính vòng hút Linh khí của họ (Luyện Khí
1 = 90 · Đại Thừa 10 = 450 · Đạo Tổ = 540). **Một hàm dùng chung hai phía**: `Game.SkillU`
(server) và `skillU()` (client, `config.ts`), cùng công thức, cùng bảng hệ số.

- Mọi hệ số vào `config/game.json → skills.ranges_u`; **mười khoá phạm vi tuyệt đối cũ bị
  xoá khỏi config và khỏi struct Go**. Xoá thay vì để lại là có chủ ý: trình biên dịch chỉ
  ra đủ 10 chỗ phải sửa, không sót chỗ nào.
- Đẩy lùi tính theo U của **người dùng**, trong `skills.push_secs` (0,3s) — tách hẳn khỏi
  `knock_secs` (0,25s) của đòn bể kiếm. Dùng chung một khoá thì chỉnh cái này lệch cái kia.
- Đàn Em có **tầm tìm 8 U** và **vòng bám 3 U** quanh đại ca. Bản trước đàn em lấy đối thủ
  gần nhất trong CẢ KHU, nên hai đứa bỏ đại ca chạy nửa bản đồ trong khi kỹ năng chỉ sống
  10 giây. Vòng bám mạnh hơn mục tiêu: ra ngoài 3 U thì quay về, dù đang đuổi ai.
- Vệt trơn của Cô Ba lưu bán kính **vào từng vũng** lúc tạo. Vũng sống 4 giây, trong 4 giây
  đó Cô Ba có thể lên cảnh giới, rớt cảnh giới hoặc rời phòng — tính lại theo người thì vũng
  tự phình ra hay biến mất mà không có lý do nhìn thấy được.
- **Bí Kíp: Tiên giới giữ 3 lượt**, phàm giới 2. Cắt lượt thừa đặt trong `refreshDerived` —
  chỗ duy nhất MỌI đường đổi cảnh giới đều đi qua (nhặt EXP, bị chém, Thiêu Đốt, phím test).
- **Chỉ báo phạm vi** (`render/skillfx.ts`): nón cho Loa Phường / Phun Trân Châu, vòng tròn
  cho Ru Ngủ / Lưới Mì / sóng Rụt Cổ, vẽ bằng **đúng bộ số server xét trúng**. Trước đây
  client không xử lý `SkillEffect` một dòng nào — không có chỉ báo nào cả. Giờ phạm vi co
  giãn theo cảnh giới (540 ở Luyện Khí → 3 240 ở Đạo Tổ) nên không vẽ ra thì người chơi
  không có cách nào biết mình đứng trong hay ngoài tầm.
- Nút kỹ năng hiện **dãy chấm lượt**: 3 chấm ở Tiên giới, 2 ở phàm giới.

**Sáu test mới** (`skillrange_test.go`), trong đó hai cái canh loại lỗi không nhìn thấy được:
- `TestPushDistanceUsesCasterU` — lấy nhầm `o.Scale` (người bị đẩy) thay vì `p.Scale` vẫn
  chạy, vẫn đẩy, chỉ là người to bị đẩy xa hơn. Trong game không ai phân biệt nổi.
- `TestNoHardcodedSkillRanges` — đọc chính `skills.go` / `skillstate.go` / `skillrange.go`,
  cấm mọi số ≥ 3 chữ số ngoài chú thích. Một con số 500 sót lại không làm test nào khác đỏ;
  nó chỉ làm kỹ năng của Đạo Tổ ngắn bằng của Luyện Khí.

### Ba lỗi hình ảnh — cả ba đều do tôi gây ra trong hai ngày qua

**1. Nút "Bảng xếp hạng" mất chữ** — `:root { color-scheme: dark }` tôi thêm cho màn hình
tải làm thẻ `<button>` lấy màu chữ TRẮNG từ bảng mặc định chế độ tối, mà mọi nút đều nền
trắng. Thu `color-scheme` về `#boot`, **và** đặt `color: var(--ink)` tường minh cho `.btn` /
`.tab` / `.char-card` — lớp sau mới là lớp chống tái phát.

**2. Trang chủ cắt mất ô nhập tên trên cửa sổ thấp** — nội dung 630 px trong panel 529 px,
phần thừa trôi vào thanh cuộn bên trong panel. Thêm hai mức `max-height` 720/560 px. Đo lại
650×561 · 650×480 · 1280×800 · 390×640: vừa khít. Còn 844×390 (điện thoại ngang) thừa 78 px.

**3. Khung giấy cuộn màn Tổng kết vỡ thành bốn mẩu ở bốn góc** — và đây là lỗi dây chuyền
đáng ghi nhất. `border-image-slice: 205 199 213 160` là **pixel của ảnh gốc 1024×768**. Hôm
qua `tools/asset-pipeline` thu `ui/scroll.png` về 256×192 theo luật chung của nhóm `ui` — và
160 + 199 = 359 đã rộng hơn cả tấm ảnh. Trình duyệt bỏ phép cắt, mỗi góc lĩnh trọn tấm cuộn
giấy tí hon.

Sửa hai lớp: **cắt theo phần trăm** (26.7% / 19.4% / 27.7% / 15.6%) nên đổi cỡ ảnh bao nhiêu
lần nữa cũng không hỏng lại, **và** pipeline giữ `ui/scroll.png` ở 1024 cho nét. Bài học rộng
hơn: *thu nhỏ ảnh làm hỏng mọi thứ đo bằng pixel của ảnh đó* — 9-slice, atlas, pivot. Chỗ nào
còn con số pixel trỏ vào một tấm ảnh thì chỗ đó là một quả mìn chờ pipeline giẫm phải.

### "Trên điện thoại lúc nào cũng chơi lại từ đầu" — tìm được cơ chế, chưa xác nhận trên máy thật

Bằng chứng có thật: nhật ký server 31/10 cho thấy một người **vào lại phòng 8 lần trong 40
giây** (`join "Thái đâu" (shipper)` lúc 09:23:32 · :38 · :43 · :47 · :52 · :55 · 09:24:02 ·
:05). Đồng thời bảng `sessions` của telemetry **không có lấy một phiên điện thoại nào** —
76 phiên đều là máy bàn Linux.

Cơ chế khớp với cả hai điều trên: `glrecover` hẹn `location.reload()` **1,2 giây sau khi
mất ngữ cảnh WebGL**. Trên điện thoại, chuyển sang app khác / khoá màn hình / kéo thông báo
là đủ để trình duyệt thu hồi ngữ cảnh. Lệnh tải lại chạy **trong lúc người chơi đang ở app
khác**; lúc quay lại thì trang đã tải lại và `pendingRejoin` tự nhảy vào một phòng mới. Nhìn
từ phía người chơi: "bật game lên là nó chơi lại từ đầu". Chốt chống lặp 30 giây lại nằm
trong `sessionStorage` — chế độ ẩn danh chặn ghi là mất chốt, và vòng lặp chạy vô tận.

Đã sửa ba chỗ:
1. **Không bao giờ tải lại khi tab đang ẩn.** Chờ tới lúc người chơi quay lại rồi mới quyết.
2. Bắt `webglcontextrestored` — trình duyệt cấp lại ngữ cảnh (chuyện thường trên điện thoại)
   thì **huỷ lệnh tải lại**, về trang chủ, không tự nhảy vào phòng mới.
3. Chốt chống lặp có bản dự phòng trong bộ nhớ, không phụ thuộc `sessionStorage`.

Và để lần sau có bằng chứng thay vì suy luận: **telemetry bắt đầu thu từ lúc vào trang chủ**,
không đợi vào trận nữa. Trước đây nó chỉ bật trong `onWelcome`, nên máy nào không vào nổi
trận thì không để lại dấu vết nào — thiết bị hỏng nặng nhất lại là thiết bị im lặng nhất.

**Chưa khẳng định đã hết**: tôi không có máy thật để thử, và bản giả lập Pixel 7 trong
Playwright chạy tốt (44,5 fps, 0 lỗi). Nhờ anh thử lại trên đúng cái điện thoại đó rồi báo:
máy gì, trình duyệt gì, và nó chơi lại từ đầu **trước hay sau** khi vào được trận. Lần này
telemetry sẽ ghi lại phiên đó kể cả khi game không vào nổi.

## Phiên 31/10 (tiếp 3) — iPhone: bộ nhớ GPU 1,4 GB → 150 MB, và ghi lỗi từ xa

### Tìm ra nguyên nhân "trên iPhone luôn bị lỗi" — và phải viết dụng cụ đo trước

iOS **không báo lỗi khi hết bộ nhớ đồ hoạ**: Safari giết hẳn tab, không ngoại lệ để bắt,
không dòng log nào. Nên không có cách nào ngoài tự đo.

Mà cột `tex_mb` trong telemetry **luôn bằng 0 từ ngày mở** — PixiJS không công bố tổng bộ
nhớ texture, nên không ai từng nhìn thấy con số này trên máy người chơi. Viết
`render/gpumem.ts` (bọc bốn hàm cấp phát/giải phóng texture của WebGL) rồi đo:

| | Máy bàn | Giả lập iPhone DPR 3 |
|---|---|---|
| Vừa vào trận | **1 416 MB** | **1 426 MB** |
| Đạo Tổ, đông người | **1 911 MB** | **1 958 MB** |

Gấp **7 lần** ngân sách của một iPhone. Và **một tấm texture duy nhất chiếm 1 365 MB:
16 384 × 16 384.**

Nguồn: mặt nạ sprite của PixiJS cấp phát render texture **bằng kích thước vật bị che**, mà
lớp nền trộn (§8.6 mục 1) trải khắp bản đồ 12 000 → làm tròn lên luỹ thừa 2 là 16 384².
Lượt #6 từng sửa đúng lỗi cùng loại cho `filterArea` và ghi lại rất rõ — nhưng **mặt nạ đi
một đường khác**, và `filterArea` không áp dụng cho mặt nạ. Cùng một cái bẫy, cửa khác.

### Sửa (CHANGES mục 2)

1. **Ô nền 1 024** — lớp nền chính vẫn một `TilingSprite` phủ cả bản đồ (không mặt nạ →
   không cấp phát gì); lớp trộn chia ô quanh khung nhìn, ô xa **huỷ hẳn** (ẩn thì texture
   vẫn nằm trên GPU, đúng thứ đang cắt). Mặt nạ gieo theo **toạ độ ô** nên nền không đổi
   kiểu sau lưng người chơi khi họ quay lại.
2. `resolution = min(DPR, 2)`, mức **Thấp 1.5**.
3. **Điện thoại không tải sẵn 10 atlas nhân vật còn lại** (7,5 MB GPU mỗi cái).
4. **Ngân sách GPU theo thiết bị** (`render/device.ts`): iOS 256 · điện thoại 384 · máy bàn
   1 024 MB. Chạm 80% hạ một bậc, vượt trần xuống Thấp ngay — **bỏ qua cả lựa chọn tay của
   người chơi**, vì không ai chọn "cho tab chết". Kiểm mỗi giây, không mỗi khung: con số này
   đổi khi NẠP ẢNH, không đổi khi vẽ.

Kết quả (2 phút đứng chỗ đông, giả lập iPhone): **38 MB vào trận · 150 MB đỉnh · 0 lỗi ·
44 fps**. Máy bàn: 1 416 → **98 MB**. Đường cong còn cho thấy việc huỷ ô hoạt động thật —
bộ nhớ **tụt** 149 → 138 MB ở phút 1.4 khi người chơi đi khỏi vùng cũ.

### Ghi lỗi từ xa (CHANGES mục 1)

- `window.onerror` / `unhandledrejection` / `webgl_context_lost` / `restored` /
  `gpu_pressure` đều gửi về server **kèm bối cảnh**: giai đoạn (`home`/`loading`/`match`),
  `texMB`, số texture, DPR, cỡ canvas, số người trên màn hình, FPS, mức chất lượng.
  Chỉ mỗi dòng thông báo thì vô dụng khi máy báo lỗi ở xa: cùng một `TypeError: null` có thể
  do thiếu art, do hết bộ nhớ GPU, hay do mất ngữ cảnh WebGL — kèm bối cảnh thì phân biệt
  được ngay từ một dòng.
- **Bộ đệm lỗi SỚM trong `index.html`**: một thẻ `<script>` chạy trước bundle gắn
  `window.onerror` và đẩy vào `window.__kxEarlyErrors`; telemetry nạp xong thì vét gửi đi.
  Không có nó thì **mọi lỗi trước khi bundle chạy xong đều biến mất** — đúng loại lỗi làm
  game không vào nổi trận, tức loại đang phải tìm.
- **`?debug=true` hiện lỗi ngay trên màn hình** (dải đỏ dưới đáy). iPhone không mở được
  console, nên đây là cách duy nhất người cầm máy đọc được lỗi rồi chụp màn hình gửi về.
- Cột `tex_mb` / `tex_uploads` trong telemetry **giờ có số thật**, không còn 0.

### Chưa làm được, nói rõ

- **Chưa chạy được WebKit thật**: cần `sudo apt-get install -y libavif16` rồi
  `npx playwright install webkit`. Lượt đo trên là **Chromium giả lập iPhone** — con số bộ
  nhớ đúng, nhưng cách Safari thu hồi tab thì không mô phỏng được. `tools/browsertest/iphone.mjs`
  đã viết sẵn để tự dùng WebKit khi có, và **nói rõ** khi phải rơi về Chromium.
- **Texture vòng kiếm ≤ 1 024 trên điện thoại** (CHANGES mục 2, NETCODE_SPEC §5 L1) — chưa
  làm được vì **cache texture vòng kiếm chưa tồn tại**: `swordring.ts` vẽ từng thanh bằng
  sprite, không bake ra `RenderTexture`. Đó là việc của L1 trong Protocol v2, không phải một
  dòng sửa. Ghi lại để đừng ai tưởng đã xong.
- **Atlas ≤ 2 048**: atlas lớn nhất hiện là 1 280×1 152, đã dưới trần. Chưa thêm phép kiểm
  tự động chặn atlas tương lai vượt 2 048.
- **Draw call lên 266** ở chỗ đông (ngân sách < 200) vì mỗi ô nền là một lượt vẽ — nút thắt
  B11 mới, đổi bộ nhớ lấy lệnh vẽ. Trên điện thoại đây là cái giá phải theo dõi.

## Phiên 31/10 (tiếp 4) — Protocol v2 bước nhỏ 1: Entity Store

Mục tiêu dừng anh chốt: **respawn = 0 với bài đo có bộ đếm "quay lại" > 0 thật**, gói < 6 KB.
Lượt này đóng vế thứ nhất. Số đầy đủ: PERF_LOG lượt #12.

### Vế khó không phải code — là bài đo

`netcode-bench` từ bước 2b luôn in respawn = 0 kèm một cảnh báo ở cuối: *"không entity nào
QUAY LẠI khung nhìn"*. Số 0 đó **vô nghĩa**: không ai quay lại thì hiển nhiên không ai nhấp
nháy. Ba lần sửa, ba nguyên nhân khác nhau, cùng một gốc là **đoán quãng đường bằng đồng hồ**:

1. Đi 4,5 giây mỗi chiều. Đo ra **798 đơn vị trong 11 giây** (≈72/giây, không phải 195 như
   `move_base` — vật cản và va chạm ăn phần lớn), trong khi bán kính RA = nửa khung nhìn 830
   + lề 900 = **1 730**. Chưa bao giờ ra khỏi tầm.
2. Cho tab "người đi" đi: ra được, nhưng ngoài kia **bị bot chém chết, hồi sinh với id mới**
   — với tab quan sát đó là người khác, không phải "quay lại".
3. Đo khoảng cách tới tab kia: phụ thuộc chỗ hai tab ngẫu nhiên sinh ra. Một lượt chúng cách
   nhau **3 900 đơn vị** và không bao giờ nhìn thấy nhau. 4 lượt ra 0 lần quay lại, 2 lượt
   khác ra 2 — **cùng một bản build**. Đúng loại dao động mà bài đo này sinh ra để xoá.

Bản chạy được: người quan sát **tự đi**, và đi theo **khoảng cách tới chỗ mình đứng lúc đầu**
— không theo đồng hồ, không theo tab kia. Ra xa 2 400 → đứng 3 giây (qua hạn trễ 2 giây) →
về trong 700.

### Entity Store (§4.3)

Client giờ **hành động theo `Despawn`**, và giữ lại mô tả (tên + nhân vật) 30 giây sau khi xoá.

Đây đúng là chỗ lượt trước làm hỏng và đã ghi lại trong `state.ts`: bật client xoá theo `ds`
làm nhấp nháy **tệ đi** (8 lần/15 lượt so với 2/13), vì server xoá đúng luật nhưng người đó
quay lại đủ nhanh để mắt thấy. Giữ mô tả thì ranh giới despawn→spawn không còn là "dựng lại
từ đầu": `ViewPool` trả lại sprite cũ, không ai thấy gì.

**Kết quả: 4 lần quay lại thật / 4 lượt · respawn 0 · placeholder 0 trên 4 744 khung ·
nhấp nháy do `Despawn` = 0.** Cảnh báo cuối bài đo đã hết.

### Chưa làm — vế thứ hai của mục tiêu dừng

**Gói vẫn 6,9 KB**, chưa dưới 6. Cần đường nóng **nhị phân** v2 (§4.1): schema + generator đã
xong từ bước trước (10 015 gói đối chiếu), nhưng `Spawn/Desc/Update` vẫn đi bằng msgpack v1 —
mới chỉ `Despawn` là tường minh.

Mổ gói lớn nhất ra thì thấy chỗ cần nhắm **không phải state người chơi**:

| Phần | KB | % |
|---|---|---|
| `ev` (sự kiện) | 2.5 | 37% |
| `ia` (vật phẩm vào tầm) | 1.7 | 25% |
| còn lại | 2.7 | 38% |

Nên bước sau nên chuyển `ev` + `ia` sang nhị phân trước — hai phần đó chiếm 62% gói.

**Cờ `proto=1|2` và thoả thuận lúc handshake cũng chưa làm, và cố ý chưa làm**: chưa có
message v2 nào chạy trên dây để mà chọn. Làm cờ trước khi có thứ để bật là làm một công tắc
không nối vào đâu.

## Phiên 31/10 (tiếp 5) — v6.6: Tiên khí ×10 thay vì ×20

Anh dán `CHANGES_v6_6.md` ngay sau khi tôi báo con số 62 phút. Đây đúng là câu trả lời cho
cảnh báo cân bằng đó, nên làm luôn.

- Trần **[1 000, 3 000, 6 000, 10 000, 15 000]** (×10 so với v5.9, trước đó v6.4 để ×20).
  Đáy: 1 / 1 001 / 3 001 / 6 001 / 10 001.
- `burn.tien_frac_per_sec` **0.004 → 0.008**. Thiêu Đốt vẫn **đúng 8 · 16 · 24 · 32 · 40**
  mỗi giây — tỉ lệ đi ngược lại đúng bằng mức trần đổi, cả hai lần.
- Mệnh giá viên 500/100/10/2/1 và trần 40 viên giữ nguyên.
- Sửa test theo checklist §17 mới. Cả bốn con số khớp đúng từ code, không chỉnh tay:
  Thái Ất 4 500 bị chém → Kim Tiên **1 001**, người chém **+1 749**, rơi **1 049**;
  Chân Tiên 800 → rơi **240**; Đạo Tổ trần **15 000**.

**Nhịp lên Đạo Tổ, ôm trọn 2 Đài Tiên Khí của một khu (5,33 điểm/giây):**

| Mốc | v6.4 (×20) | v6.6 (×10) |
|---|---|---|
| Chân Tiên → Kim Tiên | 6 phút | **3 phút** |
| Lên tới Đạo Tổ | 62 phút | **31 phút** |
| Nạp đầy để đổi Uy Danh | 94 phút | **47 phút** |

Mục tiêu anh ghi là ≈ 30 phút — **31 phút**, khớp.

Lưu ý một chỗ tôi làm chưa sạch: `CHANGES_v6_6.md` anh dán lúc 09:48 đã bị cuốn vào một
commit trước đó qua `git add -A` mà tôi **chưa đọc**. Nội dung không bị sửa, chỉ là nó nằm
trong lịch sử trước khi được làm. Từ giờ tôi soát `docs/_raw/` trước mỗi lần commit.

## Phiên 31/10 (tiếp 6) — ADMIN_SPEC A2/A3/A4 + CI art chết + rà toàn bộ

Bảy bước theo yêu cầu "làm hết một lượt, không dừng giữa chừng để hỏi". Mỗi bước một
commit riêng, chạy đủ cổng kiểm, deploy lên `:9999`.

| Bước | Việc | Build deploy |
|---|---|---|
| 1 | A3 hòm thư góp ý (§7) | `20261031_1811_53be62c` |
| 2 | Rà bảng xếp hạng + nút ẩn (§6.2) | `20261031_1827_0efc19c` |
| — | Sửa form góp ý chữ đen trên nền đen (anh báo bằng ảnh) | `20261031_1830_3517dc5` |
| 3 | A2 người chơi, cờ nghi hack, kick/cấm (§6) | `20261031_1843_fc429eb` |
| 4 | A4 xem bản đồ trực tiếp + bố cục theo cảnh giới (§5) | `20261031_1900_1040dd6` |
| 5 | Lỗi người chơi lên trang quản trị + sửa bộ đếm FPS | `20261031_1915_9973dec` |
| 6 | CI chặn art không ai nạp (`check:dead`) | `20261031_1919_b335083` |
| 7 | Rà trạng thái toàn bộ → `docs/STATUS_AUDIT.md` | — |

### Ba lỗi tìm ra trong lúc làm, không phải việc được giao

**1. Trang quản trị có thể làm SẬP SERVER.** `Room.Clients()` và `collectStats()` duyệt
`z.clients` / `g.Players` / `g.Obstacles` từ goroutine HTTP trong lúc goroutine phòng ghi
vào. Go dừng hẳn chương trình khi gặp "concurrent map iteration and map write" — một người
bấm Vào trận đúng nhịp làm mới 2 giây của dashboard là đủ. Dashboard đã chạy như thế từ A1.
Sửa bằng `Room.Query` (đẩy hàm đọc sang chạy trên goroutine phòng, giữa hai tick). Test
giữ: 248 lần đọc song song với 120 lượt vào/ra; đổi ngược lại thì `-race` báo ngay.

**2. `perfStats().fps` luôn bằng 0 với người chơi thường.** Bộ đếm khung nằm trong
`perfLine`, mà `perfLine` chỉ chạy khi mở overlay F3. Mà `perfStats()` chính là thứ đính
kèm vào sự kiện `gpu_pressure` gửi về server — nên mọi báo cáo "máy tôi chậm" từ máy thật
đều ghi `"fps":0`: đúng con số cần nhất thì trống.

**3. Form góp ý chữ đen trên nền đen.** Chép `color: var(--ink, …)` từ bảng màu trang quản
trị sang, chỗ đó `--ink` là màu sáng còn trong game nó gần đen. Tôi đã xem ảnh chụp
WebKit/iPhone lúc làm A3 và vẫn không thấy — ảnh thì mắt bỏ sót, con số thì không. Nay
`feedback.mjs` tính độ tương phản WCAG từng phần tử, dưới 3:1 là trượt.

### Phát hiện từ chính hệ thống ghi lỗi từ xa

iPhone thật (Safari 18.7) gửi về `gpu_pressure` ở **216 MB / trần 256**, trên bản build đã
có đợt giảm bộ nhớ. Tự hạ chất lượng chạy rồi mà bộ nhớ **tăng** (214 ở "Vừa" → 216 ở
"Thấp") — hạ chất lượng không trả lại texture đã nạp. `chars:7`: bảy atlas nhân vật trong
bộ nhớ. Nút thắt **B13 để mở**, câu hỏi **Q30**.

### Còn lại

Xem `docs/STATUS_AUDIT.md` mục 9. Gọn lại: **mục nhỏ (đổi số/config) không còn mục nào
chưa làm** — đã kiểm từng khoá config của v6.4 → v6.11. Mục lớn còn 3: `ev`+`ia` sang nhị
phân (62% gói), giải phóng atlas nhân vật trên iOS, và 34 tệp art không ai nạp (Q31).

## Phiên 31/10 (tiếp 7) — Protocol v2 thay hẳn v1, trả lời Q28–Q31, dọn bảng xếp hạng

| Bước | Việc | Commit |
|---|---|---|
| 0 | Tag git + `tools/deploy.sh rollback` (lùi thật: **6 giây**) · `go test -race ./...` toàn bộ: 20 gói, 0 race | `15dafff`, `4e129b3` |
| 1 | Protocol v2 — snapshot nhị phân, bỏ msgpack đường nóng | `e8717f1` |
| 2+3 | Q28–Q31 + ẩn 42 mạng đã duyệt + sửa trang quản trị đang tắt | `6cc900a` |

### Số đo Protocol v2

| | Trước (msgpack v1) | Sau (nhị phân v2) |
|---|---|---|
| Gói snapshot lớn nhất | 5,5 KB | **4,4 KB** |
| Băng thông / người | 9,3 KB/s | **3,9 KB/s** |
| Tick TB / p95 | 1,47 / 2,95 ms | **1,17 / 2,19 ms** |
| `ia` mỗi mục | 29 byte | 14 byte |
| `ev` mỗi mục | 43 byte | 17 byte |
| `m` mỗi mục | 21 byte | 10 byte |
| state người chơi | 89 byte | 14 byte (Update delta) |

Mục tiêu dừng anh đặt (< 6 KB) **đạt**. Nút thắt mới: `ir` (id vật phẩm rời tầm
nhìn) nay chiếm 62% gói lớn nhất — 558 id × 5 byte trong một đợt khi người Đạo Tổ
chạy nhanh. Chưa tối ưu thêm, đúng như anh dặn "đạt thì dừng".

### Ba lỗi tìm ra trong lúc làm

**1. Trang quản trị TẮT trên `:9999` suốt từ A1.** `deploy.sh` truyền
`KX_ADMIN_PASS`, một tên biến server không bao giờ đọc — `admin.NewAuth` chỉ nhìn
`ADMIN_PASSWORD_HASH`. Mọi thứ A1–A4 dựng lên đều không vào được, và dòng log
"trang quản trị: tắt" trôi qua mà tôi không đọc. Nay deploy đọc
`server/data/admin.env` và cảnh báo rõ nếu vẫn tắt.

**2. Bộ giải mã trả về thiếu entity.** Bản đầu chỉ trả entity mà gói hiện tại
"chạm tới". Hai đường sai, cả hai im lặng: người ĐỨNG YÊN hoàn toàn sẽ biến mất
khỏi màn hình sau vài giây (`state.expire`), và bộ nội suy thiếu mẫu nên nhân vật
trôi. Delta tiết kiệm ở đường truyền, không phải đường vẽ.

**3. `deploy.sh` treo khi gọi qua ống.** Server chạy nền kế thừa fd của subshell;
khi gọi `deploy.sh | grep` thì fd đó là cái ống và server giữ nó mở suốt đời.

### Còn lại

- **Anh cần chạy một lần**: `cd server && go run ./cmd/adminpass` → dán vào
  `server/data/admin.env`. Không có bước này thì trang quản trị vẫn tắt.
- Đo lại bộ nhớ trên **iPhone thật** sau Q30 (nút thắt B13b): giả lập cho 95,8 MB
  (từ 113,5) nhưng nó chỉ có 2–3 nhân vật, còn máy thật ghi `chars:7`.
- `ir` 62% gói — việc tối ưu tiếp theo nếu cần.

## Phiên 31/10 (tiếp 8) — chốt việc đổi tên: remote, push, và máy dev mất Go

Phiên này không sửa luật chơi. Việc là đóng nốt checklist đổi tên BladeAscend
(phần ngoài repo) và dựng lại máy dev để chạy được bộ kiểm.

| Việc | Trạng thái |
|---|---|
| Repo GitHub → `chinhhi102/bladeascend` | ✅ chủ dự án đổi |
| Thư mục `~/git/kiemxoaytutien` → `bladeascend` | ✅ chủ dự án đổi |
| `git remote set-url origin` → tên mới | ✅ |
| Push 57 commit tồn (local đi trước `origin/main` 57, sau 0) | ✅ tới `cd777f8` |

### Remote vẫn trỏ tên cũ — redirect là thứ đi mượn

Sau khi đổi tên trên GitHub, `origin` vẫn là `.../kiemxoaytutien` và `git fetch`
vẫn chạy, nhờ redirect GitHub giữ hộ. Nhưng redirect tắt ngay khi có ai tạo repo
mới trùng tên cũ — lúc đó `git push` đẩy vào repo lạ mà không báo gì. Đã đổi
`set-url`. Kiểm trước khi push: `0` commit phía sau → fast-forward thuần.

Đáng ghi: trong lúc đó repo trên GitHub **tên** đã là bladeascend nhưng **code**
bên trong vẫn là Kiếm Xoay, vì 57 commit chưa push. Đổi tên ở vỏ không đổi ruột.

### Máy dev không còn Go — và cache module bị root chiếm 1/4

`go test` / `go vet` không chạy được: máy không có Go (không phải sandbox che —
đã thử login shell, không có `/usr/local/go`, snap, hay `~/go/bin`). Lạ, vì các
phiên trước có sửa Go. Đã cài bản chính thức **go1.27.3** vào `~/.local/go`
(đối chiếu sha256), PATH thêm vào `~/.bashrc` kèm ghi chú cách gỡ.

Cài xong lộ ra một chỗ hỏng **đã nằm sẵn từ trước**: `36.797 / 147.810` mục trong
`~/go/pkg/mod` thuộc `root:root` — có lần Go bị chạy qua `sudo`. Nên Go báo
`permission denied` khi ghi cache, dù `touch` vào cùng thư mục thì được (rtk lược
cột quyền trong `ls -l`, phải dùng `stat -c '%A %U:%G'` mới thấy chủ sở hữu).

Chạy test ngay bằng cách trỏ `GOMODCACHE` sang `~/.cache/go-mod` (tải mới 411 MB).
Sau đó chủ dự án chạy `sudo chown -R chinhdc:chinhdc ~/go` rồi xoá cache tạm →
kiểm lại bằng cache mặc định: sạch. **Cố ý không** ghi `GOMODCACHE` vào `.bashrc`:
làm vậy là để máy mang hai cache song song vĩnh viễn, lỗi sẽ quay lại sau nhiều
tháng mà không ai nhớ tại sao.

### Bộ kiểm (quy tắc 7) — đủ cả hai nửa

| Bài | Kết quả |
|---|---|
| `go vet ./...` | No issues (cache mặc định, sau chown) |
| `go test ./...` | **336 pass / 23 gói** |
| `npm run build` | pass — ngân sách đạt, 0 art chết, `check:brand` sạch |

`check:brand` còn miễn trừ đúng 2 dòng ở `deploy/staging.sh`, đúng như đã thiết kế.

### Còn lại — chờ chủ dự án quyết

Ba việc, hai trong đó cùng một cái bẫy: **đổi tài liệu/script mà không đổi hạ
tầng thì người làm theo sẽ gặp thư mục rỗng hoặc `service not found`.**

- `deploy/staging.sh` dòng 18–19 vẫn rsync vào `~/kiemxoaytutien` trên máy staging.
- `docs/DEPLOY.md`: 20 dòng còn `/opt/kiemxoay` và tên service `kiemxoay`. Khác
  `PROGRESS.md` (kể chuyện cũ, miễn đúng) — đây là hướng dẫn đang dùng.
- Mua `bladeascend.io`. Code đã trỏ sẵn (`config.brand.domain` → `%BRAND_DOMAIN%`
  trong `client/index.html`), mua xong ảnh share Facebook/Zalo chạy, không sửa gì.

**Câu cần chốt**: đổi tên thư mục/service trên VPS + staging, hay giữ tên cũ ở
tầng hạ tầng? Giữ nguyên cũng được (người chơi không thấy), chỉ là `brand-check`
mang hai món nợ miễn trừ dài hạn.

## Phiên 1/11 — trang chủ + ảnh chia sẻ (5 việc chủ dự án chốt)

Commit `54b1bd9`. **Chưa deploy** — chủ dự án yêu cầu xem ảnh trước.

| # | Việc | Kết quả |
|---|---|---|
| 1 | Tên game không lẻ chữ · job logo p1 | ✅ |
| 2 | Thẻ nhân vật tên ngắn, 1 dòng | ✅ |
| 3 | 1366×768 không chạm mép, hint PC hiện | ✅ + 2 cỡ khác cũng đang hỏng |
| 4 | Phần dưới trang: ảnh, cách chơi, FAQ, chân trang | ✅ |
| 5 | Ảnh chia sẻ OG 1200×630 | ✅ chờ duyệt |

### Số đo trước / sau (7 cỡ màn hình)

| Cỡ | Cuộn ngầm trong panel | |
|---|---|---|
| | trước | sau |
| 360×800 | 26 px | **0** |
| 800×360 | 108 px | **0** |
| 320×568 | 135 px | **0** |
| 420×900 | 0 | 0 |
| 960×540 | 0 | 0 |
| **1366×768** | **73 px** | **0** |
| 1920×1080 | 0 | 0 |

Tên game bỏ lẻ chữ "Bước" ở **5/7 cỡ** → nay 0/7. Trang chủ thừa **404 px nền đen**
dưới chân trang → nay 0 px. Trang có **2 thẻ `h1`** cùng nội dung → nay 1.

### Ba điều học được

**1. `display: block` không có nghĩa là NHÌN THẤY.** Panel trang chủ có
`overflow-y: auto`; nội dung không vừa thì phần thừa không tràn ra ngoài mà trôi vào
một thanh cuộn gần như vô hình. `getComputedStyle` vẫn báo dòng hướng dẫn là
`block` trong khi người ngồi trước máy không thấy dòng nào. Phép đo đúng là
`scrollHeight - clientHeight` — đã đưa vào `uioverlap.mjs`.

**2. Khoảng trống giữa hai media query cũng là một vùng chưa ai thử.** Hai tầng
`max-height` cũ chặn ở 720 và 560 px. 768 và 800 — hai chiều cao phổ thông nhất —
rơi đúng vào giữa, không tầng nào áp, nên giữ nguyên cỡ đầy đủ. Bài kiểm cũ chạy
6 cỡ nhưng không cỡ nào nằm trong khoảng đó.

**3. Ảnh chụp từ game thật vẫn sai nếu cái "thật" đó là một nhân vật không làm gì.**
Bản đầu của `shots.mjs` chỉ chờ rồi chụp: nhân vật đứng nguyên chỗ hồi sinh, sau 84
giây vẫn "Luyện Khí tầng 2 · Kiếm 3/10 · hạng 59/59", ba ảnh gần như trùng nhau và
không ảnh nào có VÒNG KIẾM — thiếu đúng hình ảnh mà cả cái tên game nói về. Nay tool
tự kiểm ba thứ: phím dựng cảnh có tới nơi không (đếm kiếm trước/sau), có chụp trúng
màn tổng kết không, và đáy ảnh có phải vùng ngoài rìa bản đồ không.

### Ghi chú kỹ thuật

- `client/brand-split.ts` — một hàm chia tên game, dùng chung cho `vite.config.ts`
  (màn hình tải, HTML tĩnh) và `strings.ts` (bảng trang chủ, DOM). Hai runtime khác
  nhau nhưng phải ngắt ở cùng một chỗ, nếu không tên "nhảy" lúc màn hình tải mờ đi.
- `?debug=true` trên URL **không còn cấp quyền gì** (ADMIN_SPEC §1.1). Công cụ đo
  cần `APP_ENV=dev` + gọi từ localhost. Mất hai lượt chạy mới nhận ra vì lệnh bị
  chặn thì im lặng, không có dòng lỗi nào.
- `brand-check` nay quét cả `client/public/` — `privacy.html` viết thẳng tên game
  (Vite chép nguyên xi tệp trong `public/`, không thay token).
- Job `brand_logo_vi` ghi ra **art nguồn**, không đè bản phát hành. Bản đang dùng
  vẫn là bản render bằng font cho tới khi chủ dự án duyệt **từng dấu** — dễ sai nhất
  là "Hữu" (ngã) và "Bước" (sắc trên ơ).

### Còn lại

- **Chờ chủ dự án duyệt ảnh chia sẻ** rồi mới deploy. Chạy lại ảnh chụp:
  `cd server && APP_ENV=dev go run ./cmd/server -debug=true` rồi
  `node tools/browsertest/shots.mjs`.
- Chân trang ghi "English, Português… đang làm" — đúng sự thật, vì `client/src/i18n/`
  chưa có. Dựng ô chọn ngôn ngữ ở đây lúc này là một cái nút bấm không xảy ra gì.
- `share_card.png` nặng 1 MB (PNG, vì Twitter không nhận WebP). Không nằm trên
  đường vào game — chỉ con bọ mạng xã hội tải — nên chưa tối ưu.

## Phiên 1/11 (tiếp) — đổi tên staging, đơ game addressModeU, respawn, chất lượng mặc định, cổng

| # | Việc | Kết quả |
|---|---|---|
| 1 | `docs/RENAME_STAGING.md`: sao lưu, đếm dòng trước/sau, "Mong đợi" từng lệnh | ✅ — và **sửa tên volume SAI** (`kiemtien_kiemtien-data`, không phải `kiemtien-data`) |
| 1b | Chạy RENAME trên staging (anh duyệt) | ✅ `lives 167` khớp 4 lần đếm; xoá volume cũ **sau 8/11** |
| 2 | Respawn | ✅ **31 → 2** (WANDER), người thật trước đó 100/17 phút — PERF_LOG #22 |
| 3 | WebGL + `addressModeU` | ✅ ép `preference: 'webgl'`; lỗi thật nằm ở nhả atlas — **3 đường**, xem dưới |
| 4 | Chất lượng mặc định | ✅ PC Vừa · ĐT/GPU tích hợp Thấp · lên Cao cần 20 s; Cao chỉ +33 % chứ không gấp đôi — PERF_LOG #23 |
| 5 | Deploy | ✅ `:9999` + staging `https://bladeascend.staging.maveragroups.com` |
| – | Anh báo: game đơ (ảnh 15:58) | ✅ chính là `addressModeU`, đường thứ 3 |
| – | Anh báo: hình chữ nhật trên khối băng (ảnh 16:00) | ✅ viền "quá bậc" vẽ theo hình Tiled → nay ôm theo ảnh, chỉ hiện trong tầm kiếm (đúng spec §8.2) |
| – | Anh yêu cầu: cổng nổi trên vòng kiếm | ✅ lớp `gateLayer` ngay trên vòng kiếm |
| – | Anh hỏi: v6.x chưa xong? | Rà 87 mục: 57 xong · 17 một phần · 13 chưa — `docs/V6_AUDIT.md` |

### Đơ game `addressModeU` — ba đường, cùng một gốc

Gốc: `releaseChar` (nhả atlas nhân vật không ai dùng, Q30) huỷ ảnh nguồn trong khi vẫn còn chỗ giữ nó.
1. Sprite vẽ bằng **nhân vật dự phòng** — `keepChars` chỉ được khai tên muốn, không khai atlas đang vẽ.
2. **View Pool** — view cất > 10 s, atlas bị nhả; lấy ra cho cùng nhân vật thì `setCharacter` chỉ so tên.
3. **Bộ đệm `Assets`** — huỷ tay thì bộ đệm vẫn giữ xác; nạp lại trả về đúng ảnh đã huỷ. Đây là đường
   trong ảnh anh chụp (console: `[Cache] already has key` + `destroyed instead of unloaded`).

Pixi 8.22 vốn đã ưu tiên WebGL; lỗi nằm trong `GlTextureSystem.updateStyle`, không phải WebGPU.

### Lỗi hạ tầng tìm được trong lúc làm
- `tools/deploy.sh`: `pkill -f bladeascend` khớp cả đường dẫn repo → **deploy tự giết chính nó** (exit 144),
  và sẽ giết luôn `npm run dev`/editor. Nay `pkill -x`.
- `deploy/staging.sh` chép `assets/` thay vì `assets/dist/` → staging chạy **không ảnh nào**. Đã sửa.
- `deploy/docker-compose.yml`: volume nay tên cố định + `external` — thiếu thì dừng, không tạo rỗng.
- `netcode-bench.mjs` gọi binary tên cũ `kxserver`; và không bao giờ ép được entity ra/vào (xem PERF_LOG #22).

### Còn lại / lỗi đã biết
- **Telemetry staging không bền**: ghi vào `/app/server/data/telemetry.db` trong container, KHÔNG nằm trên
  volume → mất mỗi lần deploy. Cần thêm `-telemetry-db /data/telemetry.db` vào Dockerfile (chưa làm: đổi
  hạ tầng, hỏi anh trước).
- Chưa đo được **điện thoại / GPU tích hợp thật** cho mục tiêu p50 ≥ 40 / p1 ≥ 25 (B17).
- Viền "quá bậc" mới **chưa chụp ảnh kiểm** ở cảnh thật (đã kiểm cổng bằng ảnh, không lỗi JS).
- Test chập chờn `TestRemovalReportedEvenWhenSlotReused` (B18) — có từ trước, không do lượt này.
- Q32: spec §13.6 cần anh ghi luật chất lượng mới.
- v6.x chưa xong (xem `docs/V6_AUDIT.md`): lớn nhất là v6.16 (ảnh góp ý, admin antd, đa ngôn ngữ) và
  v6.19 mục 2–5 (giao diện bảng xếp hạng mới).
