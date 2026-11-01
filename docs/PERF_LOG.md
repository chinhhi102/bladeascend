# PERF_LOG — Sổ số liệu hiệu năng qua từng lượt tối ưu

> Đặt tại `docs/PERF_LOG.md`. **Agent cập nhật sau MỖI lượt tối ưu** (và sau mọi thay đổi có thể ảnh hưởng hiệu năng: VFX, bản đồ, kỹ năng...). Dữ liệu thô dạng máy đọc được ghi song song vào `docs/perf_log.csv`.
> Quy tắc vàng: **chỉ so sánh các lần đo cùng kịch bản** (cột `Kịch bản`). Số từ kịch bản khác nhau không được đặt cạnh nhau để kết luận.

---

## 1. Kịch bản đo chuẩn (không được đổi; muốn đổi thì tạo kịch bản mới)

| Mã | Lệnh | Mô tả |
|---|---|---|
| **S1** | `go run ./cmd/loadtest --scenario S1 --seed 7` | 40 nhân vật, 10 Đạo Tổ đủ kiếm đứng chụm 1 khu, 120s. Đo server + mạng. |
| **S2** | `node tools/browsertest --scenario S2 --net normal` | 10 tab Chrome headless, rải 5 khu, 5 tab lên Đạo Tổ, 60s chụm + 60s tản. Đo mạng + ổn định hình. |
| **S3** | `node tools/browsertest --scenario S2 --net slow4g` | Như S2 nhưng giả lập 4G yếu (100 ms, 1.6 Mbps / 750 Kbps). |
| **S4** | Chơi tay 1 tab có giao diện, `?debug=true`, máy chuẩn (ghi cấu hình ở §4) | Đứng giữa trận đông nhất 60s, đọc F3. Đo FPS thật. |
| **S5** | `go run ./cmd/sim --bots 40 --minutes 5 --seed 7` | Nhịp chơi (không phải hiệu năng): lượt chết, thời gian lên cảnh giới. |
| **S6** | `go run ./cmd/perfreport --real --since 7d` | Số liệu **người chơi thật** trên server LAN `:9999` (TELEMETRY_SPEC). Ghi cả số phiên N; N < 5 thì tự đánh dấu "ít mẫu, tham khảo". |
| **S9** | `node tools/browsertest/iphone.mjs <url> <giây>` | **WebKit giả lập iPhone** (DPR 3, 390×844): vào trận, đứng chỗ đông, đọc bộ nhớ texture. Dùng WebKit chứ không Chromium vì người báo lỗi dùng Safari iPhone. **Chỉ con số bộ nhớ là chuyển sang máy thật**; FPS ở đây là tốc độ vẽ phần mềm của máy dựng, không phải của iPhone. |
| **S10** | `node tools/browsertest/pingsource.mjs <url> <giây> [--cpu=N] [--daoto] [--tabs=N]` | **Trễ đến từ đâu**: chạy đồng thời ping luồng chính và ping trong Web Worker, cùng đường truyền cùng server. Chênh lệch = phần trễ do luồng chính bận. `--cpu=N` bóp CPU giả lập máy yếu; `--daoto` lên Đạo Tổ 1050 kiếm. |
| **S11** | `node tools/browsertest/cputrace.mjs <url> <giây> [--cpu=N] [--daoto]` | **Hồ sơ CPU luồng chính** (CDP `Profiler`, 1000 mẫu/giây): thời gian TỰ THÂN từng hàm. Trả lời "mỗi khung tốn N ms thì N ms đó nằm ở hàm nào". |
| **WANDER** | `node tools/browsertest/respawn-wander.mjs [client-dir] 120 4` | **Respawn 2 giây như người chơi gặp**: server riêng 9993, 60 bot, 4 tab đi lang thang 120 s, đọc thẳng `world.respawnBy` (tách do Despawn / do hết hạn / quái). `client-dir` trỏ vào `server/data/deploy-archive/<build>/client` để so theo build. Dao động giữa các lượt ~±10 — so khác biệt cỡ chục lần, không so 2 với 4. |
| **QCOST** | `node tools/browsertest/qualitycost.mjs 20 --zone=2 --cpu=4 --runs=2` | **Mức chất lượng tốn ở đâu**: server riêng 9995, vào đúng một khu, CPU 4×, ép mức tay, hạ từng hiệu ứng bằng `?qf=`. FPS p1 với 2 lượt KHÔNG đủ mẫu; tin cột "cập nhật cảnh (ms)". |
| **S8** | `node tools/browsertest/livemap.mjs <cổng>` | **Giá của trang quản trị** (ADMIN_SPEC §11): đo tick TB cao nhất khi KHÔNG ai xem bản đồ, rồi khi 2 admin xem ở 5 lần/giây, lấy hiệu. Trần 0,3 ms. |
| **S7** | `cd client && npm run assets:size` + `node tools/browsertest/loadtime.mjs <url>` | **Dung lượng tải & thời gian vào trận** (GAME_SPEC §13.7). Hai nửa: dung lượng lấy từ manifest (chính xác tuyệt đối), thời gian đo trong Chrome có bóp mạng 4 Mbps / 10 Mbps / lần 2 có cache. |

Mỗi kịch bản chạy **3 lần**, ghi trung vị. Cùng máy, cùng commit build, đóng ứng dụng nặng khác.

**S6 là ngoại lệ và là kịch bản đáng tin nhất về trải nghiệm thật.** Nó không "chạy" lần nào
cả — số liệu do chính người chơi gửi về trong lúc chơi, trên máy thật và mạng thật. Dòng S6
được sinh **tự động 2 lần/ngày (12:00 và 23:30)** vào `docs/perf/real/`, luôn kèm build id.

Vì sao nó đáng tin hơn S1–S5: cả năm kịch bản kia chạy trên máy dev, và máy dev có GPU rời.
Ngày 30/10 đội test mất hẳn hình trên laptop GPU tích hợp vì một texture 1 GB mà **không một
kịch bản nào trong S1–S5 phát hiện được** — máy dev nuốt được 1 GB nên mọi con số đều đẹp.
S6 nhìn thấy đúng những máy đó.

Đầu mỗi phiên làm việc: chép các dòng S6 mới trong `docs/perf/real/` vào bảng §3.

## 2. Bộ chỉ số (cột cố định)

| Nhóm | Chỉ số | Đơn vị | Ngân sách | Nguồn |
|---|---|---|---|---|
| Server | Tick TB / p95 (khu nặng nhất) | ms | < 3 / < 5 | loadtest, F3 |
| Server | CPU process | % 1 lõi | – | loadtest |
| Mạng | Băng thông xuống TB / đỉnh mỗi client | KB/s | < 25 / < 40 | browsertest, F3 |
| Mạng | Gói lớn nhất | KB | < 6 | browsertest, F3 |
| Mạng | Ping p95 (S3) | ms | tăng < 20 so với phòng trống | browsertest |
| Client | FPS TB / 1% thấp nhất | fps | ≥ 60 / ≥ 45 | F3 (S4) |
| Client | Thời gian vẽ / frame | ms | < 4 | F3 |
| Client | Draw call / Sprite | số | < 200 / – | F3 |
| Client | Bộ nhớ texture | MB | < 300 | F3 |
| Hình ảnh | Khung có placeholder | số | **0** | F3 |
| Hình ảnh | Entity respawn trong 2s | số / phút | **0** | F3 |
| Hình ảnh | Đổi khu (cache / chưa cache) | ms | < 300 / < 1 500 | browsertest |
| Nhịp chơi | Lượt chết / bot / 5 phút | số | (theo dõi) | sim (S5) |
| Tải về | Gói chặn vào trận | MB | **< 8** | `assets:size` (S7) |
| Tải về | Toàn bộ game | MB | **< 80** | `assets:size` (S7) |
| Tải về | Vào trận lần đầu, 4G ~10 Mbps | s | **< 10** | `loadtime.mjs` (S7) |
| Tải về | Vào trận lần 2 (có cache) | s | (gần như 0 byte) | `loadtime.mjs` (S7) |

## 3. Bảng tổng hợp theo lượt

Mỗi lượt 1 dòng **cho mỗi kịch bản đã chạy**. Ô đạt ngân sách ghi ✅, vượt ghi ❌, không đo ghi `–`. Cột `Δ` so với **lượt trước cùng kịch bản**.

| # | Ngày | Commit | Lượt / thay đổi | Kịch bản | Tick TB / p95 (ms) | KB/s TB / đỉnh | Gói max (KB) | FPS TB / 1% | Vẽ (ms) | Draw call | Placeholder | Respawn/phút | Δ chính | Nút thắt còn lại |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 0 | 2026-10-2x | ? | **Mốc trước tối ưu** | S1 | 11.45 / 19.89 ❌ | 43.1 / – ❌ | 38.3 ❌ | – | – | – | – | – | – | ItemsAdded dồn 1 gói khi camera Đạo Tổ lùi xa; server tick vượt 3–5× |
| 0b | 2026-10-2x | ? | Mốc trước tối ưu (game thật 9 người, 1 912 entity) | S4\* | 3.40 / – | 61.7 / – ❌ | 3.0 | 60 / – ✅ | 1.0 ✅ | 245 ❌ | ? | ? | – | Băng thông; draw call |
| 1 | 2026-10-2x | ? | v4.9 bản đồ sinh động — bản đầu | S4\* | – | – | – | 36 / – ❌ | – | 611 ❌ | – | – | FPS 58 → 36 | Decal/hạt là Graphics riêng; duyệt 2 000 decal/frame; bóng vật cản Graphics |
| 2 | 2026-10-2x | ? | v4.9 — bake decal, gom chunk 2048, bake bóng | S4\* | – | – | – | ~58 / – | 0.9 ✅ | 118 ✅ | – | – | Draw call 611 → 118 | – |
| 3 | 2026-10-2x | ? | Sửa lỗi `geom.SegSegDist2` (sai dấu) | S5 | – | – | – | – | – | – | – | – | Lượt chết 304 → 84 / 5 phút (luật đúng) | Số đo nhịp cũ không còn giá trị |
| **4** | 2026-10-30 | `f5720c0` | **Mốc chuẩn** — sau tối ưu mạng (mục 27) + server (mục 26) | S1 | 1.56 ✅ / 3.33 ✅ | 8.6 / – ✅ | 6.8 ❌ | – | – | – | – | – | Tick 11.45 → 1.56 · gói 38.3 → 6.8 KB | Gói còn 6.8 KB vs ngân sách mới 6 KB; cần Protocol v2 nhị phân |
| 4 | 2026-10-30 | `f5720c0` | Mốc chuẩn (client, mạng thường) | S2 | 0.7 / 1.6 ✅ | 19.5 / 45.5 ❌ | 2.5 ✅ | 7.6 / 3.6 ⚠ | 2.5 ✅ | 0 ⚠ | chưa đo | chưa đo | — | Băng thông đỉnh 45.5 KB/s ở tab Đạo Tổ |
| 4 | 2026-10-30 | `f5720c0` | Mốc chuẩn (client, 4G yếu) | S3 | 0.3 / 0.5 ✅ | 7.7 / 13.6 ✅ | 0.6 ✅ | – | – | – | chưa đo | chưa đo | — | **Cột Đạo Tổ rỗng** — xem ghi chú dưới |
| 4 | 2026-10-30 | `f5720c0` | Mốc chuẩn (nhịp chơi) | S5 | – | – | – | – | – | – | – | – | Lượt chết 2.10 → **4.70** / bot / 5 phút | Do luật v5.8 (bị chém về sàn cảnh giới) |
| 6 | 2026-10-30 | `ab0f706` | **Sửa texture 1 GB** — `filterArea` bám khung nhìn + atlas 128 + boulder 512 | S4\*\* | – | – | – | 58 / – | 0.5 | 112 | 0 | – | Texture GPU **1 076 → 102 MB**; art tĩnh 607 → 378 MB | Vẫn còn placeholder/respawn (xem dòng 7) |
| 7 | 2026-10-30 | `d34def5` | Mở telemetry trên `:9999` — **lượt chạy kiểm chứng, KHÔNG phải người thật** | S6\*\*\* | 0.35 / 1.54 ✅ | 7.9 / 8.2 ✅ | 0.5 ✅ | 59 / 28 ⚠ | – | 54 ✅ | 12 ❌ | 13.50 ❌ | — | Placeholder và respawn vẫn vượt ngân sách → hướng A bước 2b |
| 8 | 2026-10-30 | `9bca2c9` | **NETCODE bước 2b** — DescRequest, nhân vật dự phòng, mờ ra 150 ms, View Pool | FLICKER | – | – | – | 60 / – | – | 36 | **0** ✅ | **0.4** ⚠ | Placeholder 252→**0** qua 16 lượt | Respawn chưa về 0 hẳn (2/16 lượt có 1) — cần `Despawn` tường minh ở bước 3 |
| – | 2026-10-30 | `20261030_2123_200e5bf` | **Người thật** 30/10 11:30 → 23:30 (N=68 phiên, 57 người) | S6 | 0.70 / 5.30 ✅ | 16.4 / 26.7 ⚠ | 3.8 ✅ | 59 / 28 ⚠ | – | 73 ✅ | 50 ❌ | 4.10 ❌ | **dòng S6 TỰ ĐỘNG đầu tiên** | FPS 1% thấp 28 (ngân sách 45); placeholder 50 và respawn 4.10 — đo trên build TRƯỚC bước 2b+§13.7 |
| 9a | 2026-10-30 | `200e5bf` | **Mốc TRƯỚC §13.7** — bộ nạp tải cả manifest trước trang chủ | S7 | – | – | – | – | – | – | – | – | – | Gói chặn **77.4 MB** (vượt 9.7×); tổng 201.5 MB (vượt 2.5×) |
| **9b** | 2026-10-30 | `b55c14f` | **§13.7 A — asset-pipeline + nhóm tải + màn hình tải** | S7 | – | – | – | 58 / – | 0.6 | 36 | **0** ✅ | – | Gói chặn 77.4 → **7.02 MB** ✅ · tổng 201.5 → **11.41 MB** ✅ · vào trận 10 Mbps **8.3 s** ✅ · lần 2 **2.8 s** | Dư ngân sách gói chặn chỉ 0.98 MB; `zone_thien_dinh` 3.24 MB là phần nặng nhất |
| **10** | 2026-10-31 | (phiên này) | **Trang chủ hiện ngay + tải gói lúc gõ tên** (chủ dự án chốt 31/10) | S7 | – | – | – | 60 / – | 0.32 | 40 | **0** ✅ | – | **Chờ thực** (bấm Vào chơi → trong trận) 10 Mbps 6.0 → **2.2 s**, Fast 4G 12.4 → **6.0 s**, có cache → **0.3 s** · trang chủ hiện 4.0 → **2.4 s** | Fast 4G dao động 2.4–9.4 s: 6 giây gõ tên chỉ phủ được ~40% gói 7 MB ở 4 Mbps |
| 13 | 2026-10-31 | (phiên này) | **v6.6 Tiên khí ×10** (từ ×20) — trần [1 000…15 000], `tien_frac_per_sec` 0.008 | S5 | – | – | – | – | – | – | – | – | Lên Đạo Tổ ôm 2 đài **62 → 31 phút** (mục tiêu spec ≈ 30) | Thiêu Đốt vẫn đúng 8/16/24/32/40 mỗi giây |
| **12** | 2026-10-31 | (phiên này) | **Protocol v2 bước nhỏ 1** — Entity Store: client hành động theo `Despawn`, giữ mô tả qua ranh giới | NETBENCH | – | – | 6.9 (v1) | – | – | – | **0** ✅ | **0** ✅ | respawn **0/4 lượt** với **4 lần quay lại THẬT** · placeholder 0/4 744 khung | Gói vẫn 6.9 KB — cần đường nhị phân v2, chưa làm |
| **11** | 2026-10-31 | (phiên này) | **Bộ nhớ GPU 1 426 → 150 MB** — chia ô nền 1 024 + huỷ ô xa (iPhone) | S4\*\* | – | – | – | 44 / – | – | 266 | 0 ✅ | – | Texture **1 426 → 150 MB đỉnh** (−89%) trên giả lập iPhone 2 phút chỗ đông | Chạy bằng Chromium giả lập, **chưa phải WebKit** (thiếu `libavif16`) |
| 11b | 2026-10-31 | (phiên này) | Như trên, máy bàn 1280×720 | S4\*\* | – | – | – | 58 / – | 0.81 | 118 | 0 ✅ | – | Texture **1 416 → 98 MB** (−93%) | – |
| 10b | 2026-10-31 | `494d517` | Q27 — nền đúng biome + trang trí bản đồ **vẽ ra thật** (543 vật trước đây vô hình) | S4\*\* | – | – | – | 60 / – | 0.32 | 40 | 0 ✅ | – | Băng Nguyệt Hồ (+91 vật): draw call 36 → 40, vẽ 0.6 → 0.32 ms | – |
| 10c | 2026-10-31 | `494d517` | Như trên, đo trên `:9999` SAU KHI DEPLOY — **khu nhiều trang trí nhất** | S4\*\* | – | – | – | 58 / – | 0.9 | 34 | 0 ✅ | – | Phàm Trấn, 139 vật trang trí, nền `concrete` mới: 58 fps · 34 draw call · hiệu chỉnh 0/giây | – |
| **18** | 2026-10-31 | `fc429eb`→ | **A4 trang quản trị xem bản đồ trực tiếp (ADMIN_SPEC §5)** | S8 | 1.412 → 1.417 ✅ | – | 50.8 (gói xem, không phải gói người chơi) | – | – | – | – | – | Tick +0.005 ms khi 2 admin xem 5 lần/giây (trần §11: 0.3) · gói xem 398 → 50.8 KB | – |
| **19** | 2026-10-31 | `1040dd6`→ | **Bước 5 iPhone: lỗi từ xa lên trang quản trị + sửa bộ đếm FPS** | S9 | – | – | – | 9.3 (tham khảo, WebKit phần mềm) | – | – | – | – | texMB đỉnh **113,5** / trần iOS 256 · FPS từ 0 → đo được thật | iPhone THẬT đạt **216 MB (84% trần)** với 7 nhân vật — xem B13 |
| **20** | 2026-10-31 | `e8717f1` | **Protocol v2 — snapshot nhị phân, bỏ hẳn msgpack đường nóng** | S1 | 1.17 ✅ / 2.19 ✅ | 3.9 / – ✅ | **4.4** ✅ | – | – | – | – | 0 | gói 5.5 → **4.4 KB** · băng thông 9.3 → **3.9 KB/s** · tick 1.47 → 1.17 ms | `ir` (vật phẩm rời tầm) nay chiếm 62% gói lớn nhất — 558 id × 5 byte trong một đợt |
| **20** | 2026-10-31 | `6cc900a` | Q30 — nhả atlas nhân vật không ai dùng | S9 | – | – | – | 15.5 (tham khảo) | – | – | – | – | texMB đỉnh 113.5 → **95.8** (máy rảnh) · texture 218 → 149 | **Số này DAO ĐỘNG theo số nhân vật khác nhau trên màn hình.** Đo lại lúc máy đang chạy bài 10 tab: 140 MB / 230 texture. Xem B13c. |
| **20** | 2026-10-31 | `c5a8ab5` | **Protocol v2 — bài 10 tab Chrome, mạng thường** | S2 | 1.3 ✅ / 3.0 ✅ | 6.8 / **23.4** ✅ | **2.0** ✅ | 11.7 / 1.4 | 0.6 ✅ | 341 | 0 | – | Băng thông đỉnh tab Đạo Tổ **45.5 → 23.4 KB/s** (ngân sách 40) · snapshot đỉnh 2.0 KB | FPS 11.7 là 10 tab Chrome trên MỘT máy, không phải máy người chơi |
| **20** | 2026-10-31 | `c5a8ab5` | **Protocol v2 — bài 10 tab, 4G yếu** | S3 | 1.4 ✅ / 3.1 ✅ | 7.1 / 11.9 (p95) ✅ | **3.8** ✅ | 20.6 / 6.8 | 0.6 ✅ | 222 | 0 | – | snapshot tệ nhất 3.8 KB (<6) · ping p95 264 ms · tick không đổi so với mạng thường | **Mẫu tệ nhất 73,6 KB/s** — xem ghi chú ở §4: bóp mạng làm gói dồn lại rồi về một đợt, nên con số tức thời vọt lên; p95 11,9 mới là mức thật |
| **21** | 2026-10-31 | `4d2d52e` | **Tách nguồn trễ ping (anh yêu cầu)** | S10 | – | – | – | 58.6 → 24.3 (p10, CPU 4×) | 0.7–3.6 | – | – | – | ping luồng chính p95 **27/12/10/79 ms** (1 tab rảnh / Đạo Tổ / 10 tab / CPU 4×) · ping Worker **0–1 ms** ở CẢ BỐN | **Trễ là do LUỒNG CHÍNH client**, không phải mạng hay server |
| **21** | 2026-10-31 | `4d2d52e` | Hồ sơ CPU: 41 ms mỗi khung nằm ở đâu | S11 | – | – | – | 24.3 (p10, CPU 4×) | 3.6 | 341 | – | – | giải mã gói **0,00 ms** · Pixi vẽ 3,6 ms · còn lại 37,7 ms | **Toàn bộ nằm trong PixiJS**: duyệt cây cảnh + gom batch, không có nút thắt nào trong mã game |
| **22** | 2026-11-01 | `b402f80` | **Respawn 2 giây: AOI có trễ áp cả trạng thái gửi đi (người + quái)** | WANDER | – | – | – | – | – | – | – | **31 → 2** / 8 phút-tab ✅ | Respawn 3,88 → **0,25**/phút-tab; quái 28 → **0**; do hết hạn 26 → **0** | 2 lần còn lại do `Despawn` đúng luật (ra quá lề 900 rồi quay lại) |
| 22 | 2026-11-01 | `b402f80` | Như trên — giá băng thông (cùng máy, trước/sau) | S1 | 1.35 → 1.41 ✅ / 2.30 → 2.37 ✅ | 3.4 → **4.1** / – | 3.6 → **4.0** ✅ | – | – | – | – | – | +0,7 KB/s vì gửi tiếp người/quái ngay ngoài mép | Trong ngân sách; chấp nhận để đổi lấy respawn ≈ 0 |
| **23** | 2026-11-01 | `477e84f` | **Mức chất lượng tốn ở đâu — tắt từng thứ** (khu 2, CPU 4×, RTX 3060) | QCOST | – | – | – | 59.9 / 30.0 (Vừa) · 59.9 / 29.9 (Cao) | 1.65 (Vừa) · 2.20 (Cao) | 112 · 141 | – | – | Cao = Vừa **+0,55 ms (+33 %)**, KHÔNG gấp đôi; `ambient` chiếm ~0,45 ms | Chưa đo được máy yếu thật — xem §6 B17 |

\* Dòng có dấu `*`: đo trước khi chuẩn hóa kịch bản, chỉ để tham khảo, **không dùng làm mốc so sánh**. Lượt tối ưu kế tiếp phải chạy lại đủ S1–S4 để tạo **mốc chuẩn mới (#4)**.

\*\* Dòng #6 đo bằng **một tab headless** (`--use-gl=angle --use-angle=gl`), không phải S4 đúng
nghĩa (S4 cần tab có giao diện, người ngồi đọc F3 60 giây). Con số FPS ở đây chỉ để chứng minh
hình **không hỏng** sau khi sửa; số FPS thật của máy người chơi phải lấy từ S4 hoặc S6.

\*\*\* Dòng #7 là **lượt chạy kiểm chứng đường ống telemetry**, không phải số liệu người thật:
N = 1 phiên và phiên đó là tab headless của chính tôi. Ghi lại để đối chiếu khi dòng S6 **thật
sự tự động** đầu tiên xuất hiện. Hai ô ❌ (12 khung placeholder, 13.50 respawn/phút) **là số
thật** và là phần dư của NETCODE bước 2a — chúng chính là việc của bước 2b.

## 4. Máy đo chuẩn

| Vai trò | Cấu hình |
|---|---|
| Server / loadtest | Intel Core i5-12400F (6 nhân / 12 luồng), 32 GB RAM, Linux Mint 22.3 (nhân 7.0.0) |
| Client S2–S3 (headless) | Cùng máy trên · Chromium 156.0.8078.4 (Playwright) · 10 tab 960×540 |
| Client S4 (có giao diện) | Cùng máy trên · AMD Radeon RX 570 (RADV) · Chromium 156 · 1 tab 1280×720 |
| Điện thoại (khi có) | _(chưa đo)_ |

**Hai điều phải nhớ khi đọc số S2/S3**, nếu không sẽ kết luận sai:

1. **Chrome headless vẽ WebGL bằng CPU (SwiftShader) nếu không ép dùng GPU.** Mười tab
   SwiftShader tranh nhau 12 luồng thì chính việc vẽ trở thành nút thắt, và FPS trong S2/S3
   **không nói gì về máy người chơi**. Các kịch bản ở đây chạy với
   `--use-gl=angle --use-angle=gl` để dùng card thật; đo được `--use-angle=vulkan` → RADV,
   `--use-gl=swiftshader` và `--use-gl=egl` đều rơi về CPU.
2. **Server phải chạy tách lõi khỏi trình duyệt.** Không tách thì tick 2,55 ms bị thổi lên
   **269 ms** chỉ vì goroutine bị hệ điều hành gạt ra liên tục — số đo thành vô nghĩa.
   S1 chạy một mình nên không dính; S2/S3 dùng `taskset`.

FPS thật của máy người chơi chỉ lấy từ **S4** (một tab, có giao diện, GPU thật).

## 5. Chi tiết từng lượt

### Lượt #4 — Mốc chuẩn sau tối ưu mục 26+27 (2026-10-30, commit `f5720c0`)

Mục tiêu: tạo **mốc chuẩn** để mọi lượt NETCODE v6 so vào (CHANGES_v6 mục 7). Các dòng #0–#3
đo trước khi chuẩn hoá kịch bản nên không dùng để so nữa.

Thay đổi so với #0: tối ưu mạng (CHANGES_v4 mục 27) + server (mục 26) — xem commit `d132a60`.

Kết quả S1 (trung vị 3 lần, `go run ./cmd/perfreport`):

| Chỉ số | Trước (#0) | Sau (#4) | Δ | Ngân sách §2 |
|---|---|---|---|---|
| Tick TB | 11.45 ms | **1.56 ms** | −86% | < 3 ✅ |
| Tick p95 | 19.89 ms | **3.33 ms** | −83% | < 5 ✅ |
| Gói lớn nhất | 38.3 KB | **6.8 KB** | −82% | < 6 ❌ |
| Băng thông xuống | 43.1 KB/s | **8.6 KB/s** | −80% | < 25 ✅ |
| CPU process | – | 5% một lõi | – | – |

Nút thắt tìm được (có bằng chứng, xem commit `d132a60`):
1. `ItemPicked` — **173 sự kiện / 8.1 KB / 74% gói** (mổ gói bằng `cmd/loadtest`) — đã sửa.
2. `sparseRingsTouch` — **57% CPU** (pprof) — đã sửa bằng cách chỉ xét vòng ngoài cùng.
3. `expireItems` — **21% CPU** (pprof) — đã sửa bằng danh sách viên có hạn riêng.

**Chưa đo được, và vì sao** (không ghi bừa vào bảng):
- **S4 (FPS thật)**: cần một tab có giao diện, người ngồi xem 60 giây. Tôi không chạy được
  kịch bản này; các cột FPS phải do chủ dự án chạy tay rồi điền.
- **S3 cột Đạo Tổ rỗng**: dưới 4G yếu + 10 tab, cả 5 tab cần nâng lên Đạo Tổ đều mất ngữ
  cảnh WebGL (`PixiJS Error: Could not initialize shader`), `glrecover` tải lại trang, nên
  lúc nâng cấp thì `window.kx` chưa có → 0/5 tab lên được Đạo Tổ. **Số S3 hiện chỉ phản ánh
  tab thường.** Đây là lỗi của BÀI ĐO, không phải của game.
- **Draw call ở S2 ra 0.0**: bộ đếm draw call không hoạt động ở cấu hình headless đang dùng.
  Không được đọc "0 draw call" là tốt — nó là **số không đo được**.

Tác dụng phụ đã biết:
- **Nhịp chơi đổi**: lượt chết 2.10 → **4.70 / bot / 5 phút** (S5). Không phải hồi quy hiệu
  năng — là hệ quả của luật v5.8 (bị chém thì thanh về sàn cảnh giới, nên nhiều người rơi
  về Luyện Khí, nơi bị chém là chết). Ghi lại vì quy tắc §7.4 đòi ghi cả chỉ số xấu đi.

Việc tiếp theo đề xuất:
1. Sửa bài đo trước khi làm tiếp: S3 phải nâng được Đạo Tổ, và bộ đếm draw call phải chạy.
   Đo bằng một thước hỏng thì mọi kết luận sau đó đều không tin được.
2. NETCODE bước 2 (chống nhấp nháy) — xem `docs/NETCODE_PLAN.md`.
3. Gói 6.8 KB vs ngân sách mới 6 KB: cần Protocol v2 nhị phân (bước 3), msgpack còn tên
   trường nên vẫn tốn ~35% mỗi entity.

---

### Lượt #5 — NETCODE bước 2a: chống nhấp nháy & placeholder (2026-10-30, nhánh `netcode-2a`)

Mục tiêu: hai chỉ số §1 đang hỏng hẳn — entity respawn trong 2 s, và khung có placeholder.
Cả hai ngân sách đều là **0**.

Thay đổi (NETCODE_SPEC §4.3, §4.4, §5 Hiển thị 4–5):
- `state.ts`: giữ entity thêm **2 giây** sau khi vắng mặt, thay vì xoá ngay khi vắng một gói.
- `scene.ts` + `PlayerSprite.setCharacter`: đổi nhân vật **tại chỗ**, không huỷ rồi dựng lại.
- Mờ dần vào 120 ms khi nhân vật lọt vào khung hình.

Kết quả (`flicker.mjs`, **bot đi qua lại đúng mép khung nhìn 30 giây**, như bài kiểm tra §7):

| Chỉ số | Trước | Sau | Ngân sách §1 |
|---|---|---|---|
| Entity respawn trong 2 s | **38** | **0** | 0 ✅ |
| Khung có placeholder | **462 / 596 (78%)** | **0 / 595** | 0 ✅ |
| Entity rời đi (thật) | 44 | 3 | – |

44 → 3 là đúng: ba lần còn lại là entity rời đi **thật** và được xoá sau 2 giây chờ.

> ⚠ **Sửa lại ngày 30/10 — hai ô "0 ✅" ở trên chỉ đúng với SERVER TEST, không đúng với
> `:9999` thật.** Bảng trên đo trên một server test dựng riêng, ít người, bot do chính bài
> test thả ra. Chạy lại đúng `flicker.mjs` đó trên `:9999` (nhiều người, 5 khu, bot chuyển
> khu liên tục) ra **respawn 3 / 0 / 1 / 1 / 0 và placeholder 252 / 0 / 24 / 4 / 0 qua 5 lượt**
> — tức là **chưa đạt ngân sách**, và biến thiên lớn giữa các lượt.
>
> Telemetry trên `:9999` xác nhận lại bằng số liệu độc lập: **12 khung placeholder, 13.5
> respawn/phút** (dòng #7).
>
> Bài học, ghi ra để không lặp lại: tôi kết luận "đạt ngân sách" sau MỘT lượt đo trên một
> môi trường yên tĩnh hơn môi trường thật. Một lượt đo xanh trên server test không chứng minh
> được gì về server có người chơi.

Nút thắt **B3 mở lại** (xem §6): phần dễ đã xong, phần còn lại thuộc bước 2b.

Bị chặn / chưa làm được:
- **View Pool** (§5 L2) và **mờ dần RA 150 ms** chưa làm — thuộc bước 2b, cần chỗ giữ view
  sau khi despawn. Bản 2a đã đủ đưa hai chỉ số về 0, nên tách ra cho commit gọn.
- `DescRequest`, cấm placeholder ở build thật, `check:assets` — bước 2b.

Tác dụng phụ / rủi ro:
- Người **rời phòng thật** nay còn đứng im tối đa 2 giây trước khi biến mất. Đúng ý
  NETCODE §4.3 (`Despawn` sau 2 s ngoài tầm), nhưng là thay đổi nhìn thấy được.
- Bộ nhớ: tối đa thêm 2 giây entity trong bảng — không đáng kể ở quy mô 60 nhân vật.

Việc tiếp theo: bước 2b (View Pool, mờ ra, cấm placeholder build thật, `check:assets`).

---

### Lượt #6 — Texture 1 GB làm hỏng hình trên laptop (2026-10-30, commit `ab0f706`)

Khởi nguồn **không phải từ một kịch bản đo**, mà từ ảnh chụp của đội test: trên laptop + wifi,
mặt đất và nhân vật mất hẳn texture, chỉ còn nét viền trắng.

Đo trên máy thật (`renderer.texture.managedTextures`), không đoán:

| Chỉ số | Trước | Sau |
|---|---|---|
| Tổng texture GPU lúc đang chơi | **1 076 MB** | **102 MB** |
| Tấm lớn nhất | 16384×16384 = **1 024 MB** | 1969² = 14.8 MB |
| Art tĩnh (tổng mọi ảnh trong manifest) | 607 MB | 378 MB |
| FPS / draw call (1 tab headless) | 58 / 112 | 58 / 112 |

**Nguyên nhân chính** — `ground.filters` và `decor.filters` đặt bộ lọc màu lên container trải
khắp bản đồ 12 000×12 000. PixiJS dựng render texture bằng cỡ container, làm tròn lên lũy
thừa 2 là 16 384² ×4 byte = 1 GB cho **một** tấm. Laptop GPU tích hợp có `MAX_TEXTURE_SIZE`
8 192 hoặc thiếu VRAM nên cấp phát hỏng. Sửa bằng `filterArea` bám khung nhìn — bộ lọc chỉ
đổi màu từng điểm ảnh, không lấy mẫu điểm bên cạnh, nên cắt vùng cho kết quả y hệt.

**Art thừa cỡ** — đo cỡ vẽ thật rồi mới cắt:

| Thứ | Ảnh nguồn | Vẽ ra màn hình | Kết luận |
|---|---|---|---|
| Nhân vật | 256 px | 87 px (Luyện Khí) / 30 px (Đạo Tổ) | → 128 px, 220 → 55 MB |
| boulder | 1024 px | rộng TB 208 đơn vị bản đồ | → 512 px, 80 → 20 MB |
| items / buffs / decor | 256 px | 54–69 px | → 128 px |
| **cổng, tường, quầy hàng** | 512 px | 220–360 px | **giữ nguyên** — 512 là đúng cỡ |

Dòng cuối là chỗ đáng nhớ: nếu cắt theo cảm giác "cái gì cũng to quá" thì đã cắt nhầm cổng
và tường xuống 256 và làm rỗ hình ở mức thu phóng gần.

**Vì sao không kịch bản nào trong S1–S5 bắt được**: cả năm chạy trên máy có GPU rời, nuốt
được 1 GB. Đây chính là lý do S6 tồn tại.

### Lượt #7 — Mở telemetry trên `:9999` (2026-10-30, commit `d34def5`)

Không phải một lượt tối ưu — là lượt **bật dụng cụ đo**. Ghi vào sổ vì nó đổi cách mọi lượt
sau này được đo.

Chạy kiểm chứng đầu-cuối 45 giây trên `:9999`: 1 phiên, 4 mẫu, đủ 7 mục §4.1, nhận đúng
nhóm thiết bị, và mục 5 ghép được client ↔ server (tick p95 1.68 ms lúc client tụt khung).

Ngưỡng §4.3 tự báo ngay hai thứ: **12 khung placeholder** và **9 lần respawn trong 2 giây** —
cả hai là phần dư của NETCODE bước 2a mà bảng #5 trước đó ghi nhầm là "0 ✅".

Chi phí đo được: **không có mẫu nào vượt 1 KB/s**, và FPS trong lượt có telemetry không khác
lượt không có (58–59 cả hai) — đạt ràng buộc §1.1 và §1.2.

### Lượt #9 — Dung lượng tải: 77,4 → 7,0 MB (2026-10-30, GAME_SPEC §13.7 phần A)

Mục tiêu: người chơi phàn nàn "chờ tải rất lâu". Ngân sách §13.7 là **gói chặn vào trận
≤ 8 MB, toàn bộ ≤ 80 MB, vào trận lần đầu ở 4G ~10 Mbps ≤ 10 s**.

**Đo trước đã tìm ra nguyên nhân, và nó không phải "ảnh quá nặng".**
`AssetStore.preload()` tải **mọi tệp .png trong manifest** rồi mới hiện trang chủ — 547 tệp,
**77,4 MB**, trong đó có cả bộ art của 10 nhân vật và vật cản của cả 5 khu bản đồ, chỉ để
hiện một ô nhập tên. Nén ảnh thôi thì không đủ: 77,4 MB giảm 84% vẫn là 12 MB chặn người
chơi, vẫn vượt ngân sách.

Thay đổi:
1. `assets/src` (nguồn) / `assets/dist` (phát hành) — §13.7 mục 2.
2. `tools/asset-pipeline`: PNG → WebP (q72–82), thu nhỏ về kích thước hiển thị, tên tệp kèm
   hash nội dung, `manifest.json` v2 có byte + nhóm tải. Có bộ đệm (chạy lại: 0,2 s).
3. **Nhóm tải** `home / boot / zone_<khu> / char_<nhân vật> / fx_high / audio / debug`.
   Nhóm khu suy ra từ `maps/*.tmj` (loại vật cản thật của từng khu), không gõ tay.
4. Server: tệp có hash → `Cache-Control: immutable` 1 năm; `manifest.json` → `no-cache`;
   gzip cho JS/JSON/CSS/font.
5. Màn hình tải HTML/CSS nội tuyến trong `index.html`, thanh **theo byte**, nhân vật mình
   chọn chạy trên thanh, mẹo đổi mỗi 4 s, nút "Thử lại" tải tiếp.
6. `npm run assets:budget` chặn build khi vượt ngân sách; đã vào `npm run build` và CI.

Kết quả:

| Chỉ số | Trước (#9a) | Sau (#9b) | Δ | Ngân sách |
|---|---|---|---|---|
| Gói chặn vào trận | **77.4 MB** | **7.02 MB** | −91% | < 8 MB ✅ |
| Toàn bộ game | 201.5 MB | **11.41 MB** | −94% | < 80 MB ✅ |
| Số tệp chặn | 547 | 348 | −36% | – |
| Hiện trang chủ (4 Mbps) | (không hiện tới khi tải xong) | **4.0 s** | – | – |
| Vào trận, 4G ~10 Mbps | – | **8.3 s** | – | < 10 s ✅ |
| Vào trận, Fast 4G (4 Mbps) | – | 16.4 s | – | (ngoài ngân sách) |
| Vào trận lần 2 (cache) | – | **2.8 s**, 0.06 MB qua dây | – | gần như 0 ✅ |
| Khung có placeholder trong trận | – | **0** | – | 0 ✅ |
| FPS / vẽ / draw call | – | 58 / 0.6 ms / 36 | – | không xấu đi |

**Ba thứ chiếm phần lớn mức giảm, theo thứ tự — và hai trong ba không phải nén ảnh:**
1. **Bỏ 7 loại tệp không ai đọc khỏi bản phát hành (−129 MB nguồn)**: `frames/`,
   `frames_norm/` (ảnh từng khung, nguồn của atlas), `ref_sheet.png` (tờ tham chiếu cho bên
   vẽ), và `body/` + `faces/` — **art của hướng A chưa có mã nào dùng**. Năm nhóm này đều
   nằm trong `manifest.json` cũ và người chơi **tải thật**.
2. **Chia nhóm tải (−64 MB khỏi gói chặn)**: 10 nhân vật → chỉ nhân vật mình; 5 khu → chỉ
   khu được xếp vào; VFX Tiên giới và âm thanh → tải nền.
3. **PNG → WebP + thu nhỏ (72,2 → 11,4 MB, −84%)**.

Nút thắt tìm được (có số đo):
1. `AssetStore.preload()` — tải 547/547 tệp manifest trước trang chủ — `render/assets.ts`
   (bản cũ, dòng 97–116). Đây là toàn bộ nguyên nhân, không phải dung lượng ảnh.
2. `zone_thien_dinh` = **3,24 MB** (103 tệp), gần nửa gói chặn. Thiên Đình dùng cả 6 loại
   vật cản; 4 khu kia dùng 2–4 loại (0,52–2,27 MB). Đây là phần nặng nhất còn lại.
3. Khung `break`/`regrow` của vật cản = **2,6 MB** trong tổng 4,5 MB của `obstacles/`. Chúng
   chỉ chạy lúc vật cản bị phá. Chưa chuyển sang tải nền vì vẫn vừa ngân sách, và hoãn chúng
   nghĩa là lần phá vật cản đầu tiên rơi về placeholder.

Bị chặn / chưa làm được:
- **Trim viền trong suốt + ghi offset** (§13.7 mục 2, gạch đầu dòng 2) — **chưa làm**.
  Trim một sprite lẻ là dịch điểm neo của nó, mà hơn 300 chỗ gọi `assets.texture()` đặt
  anchor theo quy ước riêng; với atlas thì phải ghép lại và thêm `trim`/`orig` cho từng
  khung. Đo trước cho thấy phần thưởng nhỏ: WebP đã nén vùng alpha rỗng gần như miễn phí,
  ước tính còn 5–15%. Cả hai ngân sách đã ĐẠT mà không cần nó, nên theo quy tắc "đạt thì
  dừng" tôi để lại. Cần khi nào gói chặn tiến sát 8 MB.
- **Ghép atlas cho item/icon/decal** (cùng gạch đầu dòng) — chưa làm. Lý do giống trên: nhóm
  đó đã xuống 0,25 MB sau WebP, ghép atlas chỉ tiết kiệm số **request**, không tiết kiệm
  byte; và vẫn còn 1 MB dư ngân sách.
- **Âm thanh**: nguồn đã là 22–44 kHz mono, 31–80 kb/s — tức đã DƯỚI mức 96/128 kb/s spec
  yêu cầu. Pipeline đo trước rồi quyết định, và quyết định là giữ nguyên (nén lên sẽ làm
  tệp to hơn). libvorbis còn báo lỗi khi bị ép 128 kb/s ở 22 050 Hz.

Tác dụng phụ / rủi ro:
- Mọi đường dẫn asset giờ đi qua `assets.url()` vì tên tệp có hash. Chỗ nào còn ghép tay
  `/assets/...` sẽ **404 trong im lặng**. Đã sửa 8 chỗ (DOM `<img>`, nền CSS, font, âm
  thanh, `border-image` của màn Tổng kết — chỗ cuối này vite vẫn đang cảnh báo
  "didn't resolve at build time" từ trước).
- Chạy server mà chưa dựng `assets/dist` thì game vẽ placeholder toàn bộ. Server in cảnh
  báo kèm đúng lệnh cần chạy ngay ở dòng log đầu; `tools/deploy.sh` thì dừng hẳn.
- `assets/dist/` đã gitignore (sản phẩm dựng lại được). CI tự dựng trước khi build.

Việc tiếp theo đề xuất:
1. Fast 4G (4 Mbps) vào trận **16,4 s**. Dòng ngân sách §13.7 nói 10 Mbps nên đây không phải
   ô ❌, nhưng nếu muốn 4 Mbps cũng dưới 10 s thì chỗ cắt rẻ nhất là chuyển khung
   `break`/`regrow` (2,6 MB) sang tải nền.
2. `zone_thien_dinh` 3,24 MB — chia nhỏ theo loại vật cản hay giảm `maxPx` của `obstacles`
   từ 512 xuống 384 (đo trước rồi quyết).
3. Đo trên **điện thoại thật** — bảng §4 vẫn trống dòng điện thoại, và đó mới là máy có
   đường mạng mà §13.7 nhắm tới.

### Lượt #10 — "Thời gian chờ thực" 6,0 → 2,2 giây (2026-10-31)

Mục tiêu: chủ dự án chốt 31/10 — "trang nhập tên hiện ngay khi trang mở; tải gói boot ngầm
trong lúc người chơi gõ tên/chọn nhân vật; bấm Vào chơi mà chưa xong mới hiện thanh".

**Đổi cả chỉ số, không chỉ đổi code.** Lượt #9 đo "mở trang → trong trận", và con số đó
trộn hai thứ người chơi cảm nhận hoàn toàn khác nhau: thời gian họ *ngồi chờ* và thời gian
họ *đang chọn nhân vật*. Chỉ số đúng là **"chờ thực" = từ lúc bấm Vào chơi tới lúc ở trong
trận**. Mọi byte tải xong trước lúc bấm là byte miễn phí.

Thay đổi: trang chủ không chờ asset nào. Ảnh chân dung + nền là `<img>` /
`background-image` nên trình duyệt tự tải; trang chủ chỉ cần `/api/config` +
`manifest.json`. `home` / `boot` / font / atlas nhân vật đã lưu tải **nền** trong lúc gõ
tên. Bấm Vào chơi mà `assets.pending()` rỗng thì vào thẳng, không hiện thanh.

Kết quả (trung vị **3 lượt**, bóp mạng bằng CDP, `TYPE_MS=6000`):

| Chỉ số | Trước (#9b) | Sau (#10) | Δ | Ngân sách |
|---|---|---|---|---|
| **Chờ thực**, 4G 10 Mbps | 6.0 s | **2.2 s** | −63% | – |
| **Chờ thực**, Fast 4G 4 Mbps | 12.4 s | **6.0 s** | −52% | – |
| **Chờ thực**, lần 2 (cache) | 2.8 s | **0.3 s** | −89% | – |
| Trang chủ hiện, Fast 4G | 4.0 s | **2.4 s** | −40% | – |
| Trang chủ hiện, 10 Mbps | 2.3 s | **1.8 s** | −22% | – |
| Trang chủ hiện, có cache | 1.4 s | **1.0 s** | −29% | – |
| Mở trang → trong trận, 10 Mbps | 8.3 s | 10.0 s | +1.7 s | < 10 s (sát mép) |

Hai dòng cuối phải đọc kèm nhau, nếu không sẽ kết luận ngược: "mở trang → trong trận" **tăng**
1,7 giây chỉ vì phép đo mới có **6 giây ngồi gõ tên** ở giữa. Trừ đi thì còn 4,0 giây tải
thật, so với 8,3 giây trước đây. Giữ cột này trong bảng đúng là để không ai dùng một mình nó.

Con số "trước" của ba dòng "chờ thực" là **suy ra từ chính các lượt đo #9b**: ở bản đó không
có gì tải nền, nên chờ thực = tổng − thời gian trang chủ hiện (12,4 = 16,4 − 4,0 ·
6,0 = 8,3 − 2,3 · 2,8 − 1,4 ≈ 1,4 → lấy 2,8 là mức bảo thủ). Không phải đo lại, và ghi rõ ở
đây để đừng ai tưởng là số đo trực tiếp.

Nút thắt tìm được (có số đo):
1. **Fast 4G dao động 2,4 – 9,4 giây** qua 6 lượt. Ở 4 Mbps, gói chặn 7 MB cần ~14 giây, mà
   6 giây gõ tên chỉ phủ được ~40%. Nên ở mạng này người chơi vẫn chờ thật — con số phụ
   thuộc họ gõ tên nhanh hay chậm.
2. **Lần 2 có cache: 0,3 giây và chỉ 0,04 MB qua dây.** Tức khi byte đã nằm trong máy thì
   phần chờ còn lại gần như bằng 0 — không phải giải mã texture như tôi đoán ở lượt #9.
   Lượt #9 đo ra 2,8 giây cho trường hợp này; phần chênh là do bản đó vẫn chặn ở nhóm `home`.
3. Vật trang trí bản đồ **bắt đầu vẽ ra thật** (543 vật trước đây vô hình, xem Q27-2):
   draw call 36 → 40, thời gian vẽ 0,6 → 0,32 ms, FPS 60. Không phải nút thắt. Đo thêm trên
   `:9999` sau khi deploy, đúng **khu nhiều trang trí nhất** (Phàm Trấn, 139 vật, nền
   `concrete` mới): **58 fps, 0,9 ms vẽ, 34 draw call, 0 placeholder, hiệu chỉnh 0/giây**.
   Chúng dùng chung texture nên PixiJS gộp lô được; thêm vật trang trí gần như miễn phí.

Bị chặn / chưa làm được:
- `TYPE_MS = 6 giây` là **giả định duy nhất** của bài đo, không phải số đo. Đổi được bằng
  biến môi trường. Số thật phải lấy từ telemetry S6 (thời gian từ lúc trang chủ hiện tới lúc
  bấm Vào chơi) — hiện telemetry chưa ghi mốc đó.
- Chưa đo trên **điện thoại thật**; bảng §4 vẫn trống dòng điện thoại.

Tác dụng phụ / rủi ro:
- **Font vẫn chặn vào trận**, chỉ không chặn trang chủ: bảng tên vẽ bằng `BitmapText`, thiếu
  font là ném lỗi giữa trận. Font tải nền từ lúc mở trang nên hầu như luôn xong sẵn.
- Trang chủ vá ảnh **tại chỗ** khi manifest về, không dựng lại màn hình — dựng lại là xoá
  mất tên người chơi đang gõ.

Việc tiếp theo đề xuất:
1. Ghi mốc "trang chủ hiện → bấm Vào chơi" vào telemetry, để thay giả định 6 giây bằng số thật.
2. Muốn Fast 4G cũng dưới ~3 giây: chuyển khung `break`/`regrow` của vật cản (2,6 MB) sang
   nhóm tải nền.

### Lượt #12 — Protocol v2 bước nhỏ 1: Entity Store (2026-10-31)

Mục tiêu dừng chủ dự án chốt: **respawn = 0 với bài đo có bộ đếm "quay lại" > 0 thật**, và
gói < 6 KB. Lượt này đóng vế thứ nhất.

**Vế khó không phải code, mà là bài đo.** `netcode-bench` từ lượt #8 luôn in ra respawn = 0
kèm một cảnh báo ở cuối: *"không entity nào QUAY LẠI khung nhìn"*. Số 0 đó vô nghĩa — không
ai quay lại thì hiển nhiên không ai nhấp nháy. Ba lần sửa bài đo, ba nguyên nhân khác nhau,
tất cả đều là **đoán quãng đường bằng đồng hồ**:

1. Đi 4,5 giây mỗi chiều. Đo ra **798 đơn vị trong 11 giây** (≈72/giây, không phải 195 như
   `move_base` — vật cản và va chạm ăn phần lớn), trong khi bán kính RA là nửa khung nhìn
   830 + lề 900 = **1 730**. Chưa bao giờ ra khỏi tầm.
2. Cho tab "người đi" đi: nó ra được, nhưng ngoài kia **bị bot chém chết, hồi sinh với id
   mới** — với tab quan sát đó là người khác, không phải "quay lại".
3. Đo khoảng cách tới tab kia: phụ thuộc chỗ hai tab ngẫu nhiên sinh ra. Đo được một lượt
   chúng cách nhau **3 900 đơn vị** và không bao giờ nhìn thấy nhau. 4 lượt ra 0 lần quay
   lại, 2 lượt khác ra 2 — **cùng một bản build**.

Bản chạy được: người quan sát tự đi, và đi theo **khoảng cách tới chỗ mình đứng lúc đầu**,
không theo đồng hồ và không theo tab kia. Ra xa 2 400 → đứng 3 giây (qua hạn trễ 2 giây) →
về trong 700. Mọi entity trong khung nhìn lúc đầu đều bị kéo ra rồi kéo vào.

Thay đổi phía game: **Entity Store** (§4.3) — client giờ **hành động theo `Despawn`**, và
giữ lại mô tả (tên + nhân vật) trong 30 giây sau khi xoá.

Đây đúng là chỗ lượt trước làm hỏng và đã ghi lại: bật client xoá theo `ds` làm nhấp nháy
**tệ đi** (8 lần/15 lượt so với 2/13), vì server xoá đúng luật nhưng người đó quay lại đủ
nhanh để mắt thấy. Giữ mô tả thì ranh giới despawn→spawn không còn là "dựng lại từ đầu":
`ViewPool` trả lại sprite cũ, không ai thấy gì.

Kết quả (`node tools/browsertest/netcode-bench.mjs 4`):

| Chỉ số | Trước (#8) | Sau (#12) | Ngân sách |
|---|---|---|---|
| Lần entity **quay lại** khung nhìn | **0** (bài đo không chạm tới) | **4** | > 0 để số respawn có nghĩa |
| Respawn trong 2 giây | 2/16 lượt có 1 | **0 / 4 lượt** | **0** ✅ |
| Khung có placeholder | 0 | **0 / 4 744** | **0** ✅ |
| Nhấp nháy do `Despawn` | – | **0** | – |
| Gói snapshot lớn nhất | 6.8 KB | **6.9 KB** | < 6 KB ❌ |

Bị chặn / chưa làm được:
- **Gói vẫn 6,9 KB**, chưa dưới 6. Vế này cần **đường nóng nhị phân v2** (§4.1): schema và
  generator đã xong từ bước trước (`proto/schema.txt`, 10 015 gói đối chiếu), nhưng
  `Spawn/Desc/Update` vẫn đi bằng msgpack v1 — chỉ `Despawn` đã tường minh. Mổ gói cho thấy
  phần lớn nằm ở `ev` (2,5 KB, 37%) và `ia` (1,7 KB, 25%), tức **sự kiện và vật phẩm**,
  không phải state người chơi — nên bước tiếp theo nên nhắm vào đó trước.
- **Cờ `proto=1|2` và thoả thuận lúc handshake chưa làm.** Chưa cần: chưa có message v2 nào
  chạy trên dây để mà chọn. Làm cờ trước khi có thứ để bật là làm một công tắc không nối vào đâu.

Việc tiếp theo đề xuất:
1. Chuyển `ev` + `ia` sang đường nhị phân (hai phần chiếm 62% gói).
2. Rồi mới tới `Spawn/Desc/Update`, kèm cờ `proto`.

### Lượt #11 — Bộ nhớ GPU 1,4 GB → 150 MB (2026-10-31)

Mục tiêu: người chơi báo **"trên iPhone luôn bị lỗi"**. iOS không báo lỗi khi hết bộ nhớ
đồ hoạ — Safari **giết hẳn tab**, không ngoại lệ, không log.

**Đo trước, và phải viết dụng cụ đo trước đã.** Cột `tex_mb` trong telemetry **luôn bằng 0
từ ngày mở**: PixiJS không công bố tổng bộ nhớ texture, nên không ai từng nhìn thấy con số
này trên máy người chơi. Con số "1 076 → 102 MB" ở lượt #6 là đo tay một lần rồi thôi.

Viết `render/gpumem.ts` (bọc bốn hàm cấp phát/giải phóng texture của WebGL) rồi đo lại:

| | Máy bàn 1280×720 | Giả lập iPhone DPR 3 |
|---|---|---|
| Vừa vào trận | **1 416 MB** | **1 426 MB** |
| Đạo Tổ, đông người | **1 911 MB** | **1 958 MB** |

Gấp **7 lần** ngân sách của một iPhone.

**Một tấm texture duy nhất chiếm 1 365 MB: 16 384 × 16 384.** Nguồn gốc: mặt nạ sprite của
PixiJS cấp phát render texture **bằng kích thước vật bị che**, và lớp nền trộn (§8.6 mục 1)
trải khắp bản đồ 12 000 → làm tròn lên luỹ thừa 2 là 16 384². Lượt #6 từng sửa đúng lỗi
cùng loại cho `filterArea` và ghi lại rất rõ — nhưng mặt nạ đi một đường khác, và `filterArea`
không áp dụng cho mặt nạ.

Thay đổi:
1. **Chia ô nền 1 024 đơn vị** (CHANGES mục 2): lớp nền chính vẫn là một `TilingSprite` phủ
   cả bản đồ (không mặt nạ → không cấp phát gì), lớp trộn chia ô quanh khung nhìn. Ô ra
   ngoài tầm bị **huỷ hẳn** — `renderable = false` vẫn giữ texture trên GPU, đúng thứ đang cắt.
2. `resolution = min(DPR, 2)`, mức **Thấp dùng 1.5**. iPhone DPR 3 để nguyên là tốn gấp 9 lần.
3. **Điện thoại không tải sẵn 10 atlas nhân vật còn lại** — mỗi cái 1 280×1 152 ≈ 7,5 MB
   GPU, mười cái là 75 MB cho những nhân vật có thể không gặp lần nào. `charAtlas` đã tự
   yêu cầu đúng nhân vật vừa lọt vào khung hình.
4. **Ngân sách GPU theo thiết bị** (`render/device.ts`): iOS 256 MB · điện thoại khác 384 ·
   máy bàn 1 024. Chạm 80% hạ một bậc chất lượng, vượt trần xuống Thấp ngay — **bỏ qua cả
   lựa chọn tay của người chơi**, vì không ai chọn "cho tab chết".

Kết quả (`node tools/browsertest/iphone.mjs`, 2 phút đứng chỗ đông):

| Chỉ số | Trước | Sau | Δ | Ngân sách |
|---|---|---|---|---|
| Texture, giả lập iPhone, vào trận | 1 426 MB | **38 MB** | −97% | – |
| Texture, giả lập iPhone, đỉnh 2 phút | 1 958 MB | **150 MB** | −92% | < 256 MB ✅ |
| Texture, máy bàn, Đạo Tổ đông | 1 911 MB | **98 MB** | −95% | – |
| FPS giả lập iPhone | – | 44 | – | (trần 45 trên điện thoại) |
| Lỗi / tab bị giết trong 2 phút | – | **0** | – | 0 ✅ |

Đường cong bộ nhớ còn cho thấy việc huỷ ô nền hoạt động thật: nó **tụt** 149 → 138 MB ở
phút 1.4 khi người chơi đi khỏi vùng cũ, chứ không chỉ tăng đều.

Nút thắt tìm được (có số đo):
1. Mặt nạ nền map-sized — 1 365 MB trong MỘT tấm. Đã đóng.
2. Atlas nhân vật 1 280×1 152 = **7,5 MB mỗi nhân vật**, 11 nhân vật = 82 MB. Trên điện
   thoại giờ chỉ tải khi gặp; trên máy bàn vẫn tải sẵn cho mượt.
3. Số texture vẫn bò lên 58 → 260 trong 2 phút (ô nền + nhân vật + khu). MB thì đứng yên
   quanh 140–150, nên chưa phải nút thắt — nhưng là chỗ cần nhìn nếu con số lại trôi.

Bị chặn / chưa làm được:
- **Chưa chạy được WebKit thật** (`libavif16` cần quyền sudo). Lượt đo này là Chromium giả
  lập iPhone: con số bộ nhớ đúng, nhưng cách Safari thu hồi tab thì không mô phỏng được.
  Cài: `sudo apt-get install -y libavif16` rồi `npx playwright install webkit`.
- **Texture vòng kiếm ≤ 1 024 trên điện thoại** (NETCODE_SPEC §5 L1) — chưa làm được vì
  **cache texture vòng kiếm chưa tồn tại**: `swordring.ts` vẽ từng thanh bằng sprite, không
  bake ra `RenderTexture`. Đó là việc của bước Protocol v2 / L1, không phải một dòng sửa.
- **Atlas ≤ 2 048**: hiện atlas lớn nhất là 1 280×1 152, đã nằm dưới trần. Chưa thêm phép
  kiểm tự động chặn atlas tương lai vượt 2 048.

Tác dụng phụ / rủi ro:
- Ô nền dựng lại khi người chơi quay lại vùng cũ. Mặt nạ gieo theo **toạ độ ô**, nên cùng
  một ô luôn ra cùng mảng loang — nền không "đổi kiểu" sau lưng người chơi.
- Draw call tăng (118 → 266 ở chỗ đông) vì mỗi ô nền là một lượt vẽ. Vẫn dưới ngân sách
  200 ở máy bàn... **ở chỗ đông trên điện thoại thì 266 là vượt** — ghi vào nút thắt B10.

Việc tiếp theo đề xuất:
1. Chạy lại bằng WebKit thật sau khi cài `libavif16`.
2. Gộp ô nền trộn thành một lớp duy nhất (hoặc bake sẵn) để kéo draw call xuống.

### Lượt #22 — Respawn 2 giây: vùng quan tâm có trễ áp cả trạng thái (2026-11-01, commit `b402f80`)

Mục tiêu: respawn/phút về 0 (ngân sách §4.3). Người thật 1/11 trên build `20261101_1149_6e73318`
(SAU bản sửa decode) vẫn ghi **100 lần / 17 phút** (`perfreport --real`, mục 8 "Theo BUILD").

Đo trước rồi mới sửa. `netcode-bench` 3 lượt ra `rời=0 · quay lại=0` — bài đó không ép được
entity ra/vào khung (tab đi không lách vật cản tới được tab kia; `bot_fill_target = 0` khiến
`trimOverCap` xoá ngay bot thả bằng phím B). Viết bài mới `tools/browsertest/respawn-wander.mjs`
(kịch bản **WANDER**: server riêng cổng 9993, 60 bot, 4 tab đi lang thang 120 s, đọc thẳng
`world.respawnBy`), chạy được trên client cũ trong `deploy-archive/` để so theo build.

| Lần đo | Server | Respawn | do Despawn | do hết hạn 2 s | trong đó quái | Lỗi JS |
|---|---|---|---|---|---|---|
| Trước | `6e73318` | **31** | 5 | 26 | (chưa tách) | 0 |
| Sửa người chơi | + Perceivable/Tracks | 20 | 0 | 20 | – | 0 |
| (lần lặp, tách quái) | như trên | 30 | 2 | 28 | **28** | 0 |
| **Sau** | `b402f80` | **2** | 2 | **0** | **0** | 0 |

Nút thắt tìm được (có bằng chứng):
1. `room/broadcast.go` `snapshotFor` gửi trạng thái theo `CanSee` — có phép thử ĐÚNG MÉP khung
   nhìn — trong khi `aoi.go` theo dõi tới lề ra 900 + chờ 2 s. Lớp trễ chỉ quyết định lúc gửi
   `Despawn`, không quyết định lúc gửi trạng thái. Test `TestSnapshotKeepsSendingTrackedPlayerJustOutsideView`
   trượt trên code cũ, qua trên code mới.
2. Quái không có lớp trễ nào: 28/30 lần respawn của lần đo tách nguồn.

Giá (S1, cùng máy, trước/sau): băng thông 3,4 → 4,1 KB/s · gói lớn nhất 3,6 → 4,0 KB · tick
1,35 → 1,41 ms. Chấp nhận: vẫn trong ngân sách, đổi lấy respawn ≈ 0.

Việc tiếp theo: xem lại respawn trên telemetry người thật khi có ≥ 5 phiên ở build mới.

### Lượt #23 — Mức Cao tốn ở đâu (2026-11-01, commit `477e84f`)

Mục tiêu: chủ dự án hỏi "vì sao Cao tốn gần gấp đôi Vừa" (telemetry 31/10: vẽ p95 64,9 vs 18,2 ms).
Công cụ: `tools/browsertest/qualitycost.mjs` (kịch bản **QCOST**: server riêng 9995, vào đúng
khu 2 — khu tệ nhất 31/10 — CPU bóp 4×, ép mức bằng `kx.quality.manual`, hạ từng thứ bằng
`?qf=tên:medium`). 2 lượt/cấu hình × 20 s, trung vị. GPU: RTX 3060 (ANGLE/OpenGL).

| Cấu hình | FPS p50 | FPS p1 | Cập nhật cảnh (ms) | Draw call | Sprite |
|---|---|---|---|---|---|
| Vừa | 59.9 | 30.0 | 1.65 | 112 | 3 567 |
| Cao | 59.9 | 29.9 | **2.20** | 141 | 4 451 |
| Cao −bloom | 59.9 | 29.9 | 2.06 | 163 | 4 478 |
| Cao −ambient | 59.9 | 59.5 | **1.75** | 134 | 4 432 |
| Cao −decals | 59.9 | 29.9 | 2.12 | 152 | **3 428** |
| Cao −beast | 59.9 | 59.5 | 2.04 | 159 | 4 094 |
| Cao −motes | 59.9 | 29.9 | 2.30 | 247 | 4 296 |
| Thấp | 59.9 | 59.5 | 1.89 | 157 | 2 496 |
| ĐT (giả lập) Thấp | 59.9 | 59.5 | 2.59 | 219 | 2 702 |
| ĐT (giả lập) Vừa | 59.9 | 59.5 | 2.46 | 140 | 3 182 |

Kết luận có bằng chứng:
1. Trên GPU thật, kể cả CPU 4×, **mọi mức đều 60 FPS p50**. Cao hơn Vừa **0,55 ms cập nhật cảnh
   (+33 %)**, không phải gấp đôi.
2. `ambient` (150 hạt mây/lá ở Cao so với 60 ở Vừa) chiếm **~0,45 ms** trong 0,55 ms đó; `decals`
   chiếm ~1 000 sprite nhưng gần như không tốn ms.
3. FPS p1 dao động 30 ↔ 59,5 theo kiểu hai giá trị (một khung rơi trong 20 s) — 2 lượt là KHÔNG
   ĐỦ mẫu để kết luận gì về p1 giữa các cấu hình.
4. Con số "gấp đôi" của telemetry 31/10 lẫn phiên Playwright (Chrome/156, 64/75 phiên) chạy
   10 tab chung một máy — không phải máy người chơi.

Thay đổi đi kèm (chủ dự án chốt 1/11): mặc định PC = Vừa, điện thoại/GPU tích hợp = Thấp, lên
Cao chỉ khi ≥ 55 FPS suốt 20 s. Sửa luôn lỗi `autoQuality` nâng ngược ngay sau khi
`autoQualityByMemory` hạ (bộ đếm không đặt lại) — `npm run check:quality` bắt được.

Bị chặn: **chưa đo được điện thoại thật / GPU tích hợp thật** — giả lập trên RTX 3060 không nói gì
về mục tiêu "điện thoại p50 ≥ 40 / p1 ≥ 25". Cần telemetry người thật ở build mới.

### Mẫu — copy cho mỗi lượt

```
### Lượt #N — <tên ngắn> (YYYY-MM-DD, commit abc1234)

Mục tiêu: <chỉ số muốn giảm, vì sao>
Thay đổi: <liệt kê ngắn, link mục spec / CHANGES>

Kết quả (trung vị 3 lần):
| Chỉ số | Trước (#M) | Sau | Δ | Ngân sách |
|---|---|---|---|---|
| ... | | | | |

Nút thắt tìm được (top 3, có bằng chứng):
1. <hàm / hệ thống> — <số đo: % CPU từ pprof, ms, KB, draw call> — <file:dòng>
2. ...

Bị chặn / chưa làm được:
- <việc> — <lý do: thiếu asset, rủi ro, cần anh quyết...>

Tác dụng phụ / rủi ro: <hình ảnh, gameplay, test nào bị ảnh hưởng>
Việc tiếp theo đề xuất: <1–3 việc, xếp theo lợi ích / công sức>
```

## 6. Nút thắt đang mở (cập nhật liên tục)

| # | Nút thắt | Ảnh hưởng | Bằng chứng | Trạng thái | Lượt xử lý |
|---|---|---|---|---|---|
| B17 | **Chưa có số đo máy yếu THẬT cho mức chất lượng mặc định** | Không xác nhận được mục tiêu điện thoại p50 ≥ 40 / p1 ≥ 25, PC p1 ≥ 30 | Lượt #23: trên RTX 3060 + CPU 4× mọi mức đều 60 FPS; telemetry 31/10 lẫn phiên máy | Mở — chờ ≥ 5 phiên người thật ở build `≥ 20261101_16xx`, đọc `perfreport --real` mục 3 + 8 | – |
| B18 | **`TestRemovalReportedEvenWhenSlotReused` chập chờn** | CI đỏ ngẫu nhiên | Trượt 1/6 lần trên code CŨ (`367eb68`), không liên quan respawn | Mở | – |
| B1 | ItemsAdded dồn 1 gói khi camera lùi xa | Gói 38 KB, nghẽn mạng yếu | Lượt #0 | **Đóng** — gói 38.3 → 6.8 KB | #4 |
| B2 | Server tick vượt ngân sách khi 10 Đạo Tổ chụm | Lag server | Lượt #0 | **Đóng** — tick 11.45 → 1.56 ms | #4 |
| B3b | **Chỉ số respawn DAO ĐỘNG MẠNH giữa các lượt** | Phương pháp đo | Cùng một bản build, 8 lượt `flicker.mjs` ra 0,0,0,2,3,1,0,0. Và `netcode-bench` cũ **không bao giờ ép được entity quay lại khung nhìn**, nên số respawn = 0 của nó là con số hiển nhiên, không chứng minh gì. | **Đóng (#22)** — kịch bản WANDER (`respawn-wander.mjs`, 60 bot, 4 tab × 120 s) ép được entity ra/vào thật và tách nguồn: trước 31 (26 hết hạn, quái 28/30 lần tách), sau 2 (cả hai do Despawn đúng luật). Gốc: trạng thái gửi theo mép khung nhìn, không theo lớp trễ AOI | #22 |
| B3 | Nhấp nháy + placeholder khi entity vào khung | Hình ảnh | Trên `:9999` THẬT, 16 lượt `flicker.mjs` sau bước 2b: **placeholder 0/16 lượt** (trước 2b: 252/0/24/4/0), **respawn 2/16 lượt có đúng 1** (trước 2b ~1.0/30s, trước 2a 38/30s) | **Đóng hẳn (#20).** Sau Protocol v2, `netcode-bench` 3 lượt: **respawn 0/3 lượt với 5 lần quay lại THẬT**, placeholder **0 / 8 707 khung**, nhấp nháy do `Despawn` 0, do luật hết hạn 2 giây 0. | #5, #7, #8 |
| B12 | **Trang quản trị đọc trạng thái phòng từ goroutine HTTP** | **Sập server**, không phải chậm | `go test -race ./internal/room/` trên bản cũ: DATA RACE ngay lượt đầu. `Room.Clients()` và `collectStats()` duyệt `z.clients` / `g.Players` / `g.Obstacles` trong lúc goroutine phòng ghi vào — Go dừng hẳn chương trình với "concurrent map iteration and map write". Dashboard làm mới 2 giây/lần từ A1, nên chỉ cần một người bấm Vào trận đúng nhịp. | **Đóng (#18)** — mọi lần đọc đi qua `Room.Query`, chạy trên chính goroutine phòng giữa hai tick. Test giữ: 248 lần đọc song song với 120 lượt vào/ra, `-race` sạch. | #18 |
| B15 | **Mỗi khung tốn 41 ms trên máy yếu (CPU 4×), FPS p10 24,3 < 40** | Người chơi máy cũ thấy giật | Hồ sơ CPU (S11, 40s, Đạo Tổ 1050 kiếm): `(program)` 15,8% · `RenderTargetSystem.Ge` 6,4% · `_getGlobalBoundsRecursive` 4,3% · `collectRenderablesSimple` 4,2% · `packIndex`/`packAttributes`/`packQuadAttributes` ~6% · `set font` 1,7%. **Không có hàm nào của mã game trong top 22.** Toàn bộ là PixiJS duyệt cây cảnh và gom batch, tỉ lệ với số sprite (2 724 sprite, 341 draw call). | **Mở** — hướng sửa là giảm SỐ SPRITE/kích thước cây cảnh (vòng 1 050 kiếm là nguồn chính), không phải tối ưu mã game. Chưa làm: đó là một lượt tối ưu riêng. | – |
| B16 | ~~Giải mã gói trên luồng chính~~ — **không phải nút thắt** | – | Đo S11: giải mã 0,00 ms/khung. Protocol v2 làm nó gần như miễn phí (gói nhị phân 0,3 KB). | **Đóng** — Web Worker giải mã gói (CHANGES_v4 mục 25) **bỏ hẳn**: chuyển 0 ms sang luồng khác chỉ thêm một lớp sao chép bộ nhớ. Chủ dự án duyệt 31/10. | – |
| B13c | **Số đo S9 dao động mạnh theo tải máy và số nhân vật trên màn hình** | Phương pháp đo | Cùng một build, hai lượt cách nhau 20 phút: **95,8 MB / 149 texture** (máy rảnh) và **140,2 MB / 230 texture** (máy đang chạy bài 10 tab, FPS tụt còn 2,1). Nhả atlas cần vòng vẽ chạy và cần 10 giây rảnh; máy 2 fps thì gần như không có nhịp nào để nhả. | **Mở** — S9 chỉ so được khi máy ở cùng trạng thái. Nên ghi kèm "máy rảnh / máy bận" vào mỗi dòng, như quy tắc vàng §1 đòi với kịch bản. | – |
| B13b | **Chưa đo lại được trên iPhone THẬT sau khi sửa** | Không biết 216 MB đã xuống chưa | Giả lập WebKit cho 95.8 MB (từ 113.5), nhưng nó chỉ có 2–3 nhân vật trên màn hình còn máy thật ghi `chars:7`. Con số giả lập chứng minh CƠ CHẾ chạy, không chứng minh mục tiêu ≤ 180 MB lúc đông người. | **Mở** — cần một lượt chơi thật trên iPhone rồi đọc `/admin/#errors`. | – |
| B13 | **iPhone thật chạm 216 MB / trần 256 MB, và hạ chất lượng KHÔNG giải phóng texture** | Safari iOS thu hồi tab → người chơi thấy "luôn bị lỗi" | Sự kiện `gpu_pressure` từ máy thật, build `20261031_1050_e9267eb` (tức là **đã có** bản giảm bộ nhớ 76cc1e8): `texMB 214, q="Vừa"` rồi `texMB 216, q="Thấp"` cách nhau vài giây — tự hạ chất lượng xong bộ nhớ **tăng**, không giảm. 7 nhân vật đã nạp (`chars:7`); mỗi atlas nhân vật là phần lớn chi phí. Giả lập WebKit trên bản hiện tại: 113 MB với ít nhân vật hơn. | **Đóng (#20)** — `assets.keepChars` nhả atlas không ai dùng (trễ 10s; chạm trần thì nhả ngay), và `destroy(true)` huỷ cả ảnh nguồn nên `texMB` giảm thật. Giả lập WebKit: 113.5 → 95.8 MB, texture 218 → 149. Xem B13b về việc đo lại trên máy thật. | #20 |
| B14 | **`perfStats().fps` luôn bằng 0 với người chơi thường** | Bằng chứng gửi về server thiếu đúng con số cần nhất | Mọi sự kiện `gpu_pressure` từ máy thật đều ghi `"fps":0`, kể cả lúc máy đang tụt khung. Nguyên nhân: bộ đếm khung nằm trong `perfLine`, mà `perfLine` chỉ chạy khi mở overlay F3. | **Đóng (#19)** — dời `sampleFps` vào vòng vẽ. Đo lại trên WebKit: 0 → 9,3. | #19 |
| B11 | **Gói "xem bản đồ trực tiếp" 398 KB** | 2 admin xem = 3,9 MB/giây; tick +0,498 ms (vượt trần §11 là 0,3) | Bản đầu gửi 6 300 viên dạng `{"x":9761.202420898862,…}`. Đo bằng `livemap.mjs`. | **Đóng (#18)** — lượng tử hoá 16 đơn vị + mảng số phẳng theo loại + gộp ô trùng: **398 → 50,8 KB**, tick **+0,498 → +0,005 ms**. Số viên thật vẫn gửi riêng (`n_items`) nên không im lặng giấu mất gì. | #18 |
| B4 | Băng thông game thật 61.7 KB/s | Mạng yếu | Lượt #0b | **Đóng (#20)** — sau Protocol v2, S2 đo lại: đỉnh tab Đạo Tổ **23.4 KB/s** (trước 45.5; ngân sách 40), trung bình 6.8. Snapshot đỉnh 2.0 KB, trung bình 0.3. | #20 |
| B5 | Gói snapshot 6.8 KB vs ngân sách mới 6 KB | Mạng yếu | Lượt #4 | Mở — cần Protocol v2 nhị phân (NETCODE bước 3) | – |
| **B6** | **Bài đo S3 không nâng được Đạo Tổ**; draw call đọc ra 0 | Không đánh giá được kịch bản nặng trên mạng yếu | Lượt #4 | Mở — **phải sửa trước khi tin số S3** | – |
| B7 | Lần tải đầu 77,4 MB / 547 tệp trước khi hiện trang chủ | Người chơi chờ rất lâu rồi bỏ | Lượt #9a (`assets:size`) | **Đóng** — gói chặn 7,02 MB, vào trận 8,3 s ở 10 Mbps | #9 |
| **B8** | Dư ngân sách gói chặn chỉ còn **0,98 MB**; `zone_thien_dinh` chiếm 3,24 MB | Thêm một khu hoặc một bộ vật cản mới là vượt 8 MB | Lượt #9b | Mở — `assets:budget` sẽ chặn build khi vượt, nên không lọt ra người chơi được. Cắt tiếp: khung `break`/`regrow` 2,6 MB sang tải nền, hoặc `obstacles` `maxPx` 512 → 384 | – |
| **B9** | Fast 4G (4 Mbps) vào trận **16,4 s** | Người chơi mạng yếu | Lượt #9b | Mở — **ngoài** ngân sách §13.7 (dòng ngân sách ghi 10 Mbps), ghi lại để không ai tưởng đã xong | – |
| **B10** | **Bộ nhớ GPU 1,4–1,9 GB** vì mặt nạ nền map-sized 16 384² | iPhone Safari giết tab → "luôn bị lỗi" | Lượt #11 (`render/gpumem.ts`) | **Đóng** — 150 MB đỉnh, dưới ngân sách iOS 256 MB | #11 |
| **B11** | Draw call **266** ở chỗ đông trên điện thoại (ngân sách < 200) | Chi phí của việc chia ô nền | Lượt #11 | Mở — gộp ô nền trộn hoặc bake sẵn | – |
| **B12** | Chưa chạy được **WebKit thật**, chỉ Chromium giả lập iPhone | Không xác nhận được hành vi Safari | Lượt #11 | Mở — cần `sudo apt-get install -y libavif16` | – |

## 7. Quy tắc cho agent
1. **Đo trước, rồi mới sửa.** Mỗi lượt tối ưu bắt đầu bằng 1 dòng "Trước" (hoặc dùng dòng mốc gần nhất cùng kịch bản nếu chưa đổi code).
2. Sau mỗi lượt: thêm dòng vào §3, thêm khối chi tiết §5, cập nhật §6, ghi dòng CSV vào `docs/perf_log.csv`.
3. Nút thắt phải có **bằng chứng số** (pprof top 10, F3, kích thước gói) — không ghi phỏng đoán.
4. Lượt làm **xấu đi** chỉ số nào (dù vì lý do đúng, như thêm VFX) vẫn phải ghi rõ, kèm lý do chấp nhận.
5. Viết lệnh `npm run perf:report` (hoặc `go run ./cmd/perfreport`) chạy S1–S3 và in sẵn 1 dòng bảng §3 + 1 dòng CSV, để việc ghi sổ không phụ thuộc chép tay.

## 8. Định dạng CSV (`docs/perf_log.csv`)
```
run,date,commit,label,scenario,tick_avg_ms,tick_p95_ms,cpu_pct,kbps_avg,kbps_peak,pkt_max_kb,ping_p95_ms,fps_avg,fps_low1,draw_ms,draw_calls,sprites,tex_mb,placeholder_frames,respawn_per_min,zone_switch_cached_ms,zone_switch_cold_ms,deaths_per_bot_5min,blocking_mb,total_dl_mb,ttp_10mbps_s,ttp_4mbps_s,ttp_cached_s,notes,n_sessions
```

`n_sessions` chỉ có nghĩa với S6 (số phiên người thật gộp vào dòng đó); các kịch bản khác để TRỐNG, không điền 0 — "không áp dụng" và "0 phiên" là hai chuyện khác nhau.
