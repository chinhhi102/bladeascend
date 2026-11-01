# Rà soát CHANGES v6.x so với CODE — 01/11/2026

Đối chiếu từng mục trong `docs/_raw/CHANGES_v6*.md` với code (`server/`, `client/src`, `config/game.json`, `maps/`).
Bản đang chạy trên `:9999` là build `20261101_1607_83f9f4e` = HEAD, nên mọi thứ ghi XONG dưới đây **đã có trên server**.

## Tóm tắt

| Trạng thái | Số mục |
|---|---|
| XONG | 57 |
| MỘT PHẦN | 17 |
| CHƯA | 13 |
| **Tổng** | **87** |

Theo file: v6 (4/5/0) · v6.1 (4/5/0) · v6.2 (5/1/0) · v6.4 (11/0/0) · v6.5 (5/0/0) · v6.6 (5/0/0) · v6.7+v6.8 (13/0/0) · v6.9–v6.11 (6/3/0) · v6.16 (3/3/9) · v6.19 (1/0/4) — (XONG/MỘT PHẦN/CHƯA).

**Vì sao chủ dự án "không thấy gì":** phần lớn v6.0/v6.2/v6.4/v6.5/v6.6/v6.7 là thứ **không nhìn thấy trực tiếp** (giao thức mạng, telemetry, dung lượng tải, đổi số Tiên khí, kiểm bản đồ). Những thứ người chơi THẤY RÕ nhất trong v6.16 (tải ảnh góp ý, admin antd, đa ngôn ngữ, SEO) và v6.19 (bảng xếp hạng mới) thì **chưa làm** — chỉ phần đổi tên (v6.16 D) và danh hiệu theo phòng (v6.19 §1) là xong.

### Danh sách MỘT PHẦN
1. **v6 #2** — NETCODE §8: bước 1 (chống nhấp nháy) và 3 (Protocol v2 nhị phân, delta, keyframe) xong; **bước 2 Room Shared Cache, 4 encode-once, 5 ZoneReady + IndexedDB, 7 Service Worker: CHƯA**; bước 6 chỉ có cache nameplate, không có Web Worker, không có cache texture vòng kiếm dùng chung.
2. **v6 #3** — v2 **thay hẳn** v1 (commit `e8717f1`), không qua giai đoạn chạy song song sau cờ `proto=1`; bài §7 chưa qua hết (không có ZoneReady/SW để kiểm).
3. **v6 #4** — có `tools/browsertest/flicker.mjs`, `netcode-bench.mjs`, `run.mjs`; nhưng không có 1 lệnh chạy đủ 7 bài §7 (đổi khu cache, mất mạng 5s, vào lần 2 qua SW không thể kiểm vì chưa có tính năng).
4. **v6 #6** — `cmd/perfreport` chỉ chạy **S1**; S2/S3 phải chạy browsertest rồi điền tay (spec: S1–S3).
5. **v6 #8** — PERF_LOG chỉ có lượt cho các bước đã làm (1, 3); bước 2,4,5,6,7 chưa có.
6. **v6.1 #2** — 70% quay lại cùng tên 4–12s: có. **5% đứng yên 3–8s sau khi vào: CHƯA** (`idle_after_join_*` có trong config nhưng không code nào đọc). Bot mới sau khi "rời" vào theo nhịp `spawn_every_secs: 2`, không phải 5–20s.
7. **v6.1 #3** — 1 000 tên bot đều nằm trong bộ ký tự BitmapFont (đã kiểm: 0 ký tự thiếu), nhưng **không có font dự phòng** cho tên (người thật) có ký tự ngoài bộ — server chỉ lọc ký tự điều khiển.
8. **v6.1 #4** — **Còn lộ bot**: server gửi cờ `FlagBot` cho MỌI client trong snapshot (`server/internal/room/client.go:430`, `client/src/net/protocol.ts:243`) → ai mở DevTools/đọc gói là biết ai là bot.
9. **v6.1 #8** — artgen có job thân+mặt (`server/cmd/artgen/bodyface.go`), art mặt đã có cho 3 nhân vật (`assets/src/characters/{a_hon,dai_ca_dep,su_phu_mi}/faces/`), nhưng **client không ghép lớp mặt lên thân** (không có code dùng `faces/`) → người chơi không thấy.
10. **v6.1 #9** — GIF trước/sau có (`docs/anim_preview/`), `docs/JITTER_REPORT.md` có; PERF_LOG **không có mục "Chất lượng hoạt ảnh"**.
11. **v6.2 #6** — CSV có dòng S6 30/10 + 31/10 (2 dòng), nhưng bảng PERF_LOG §3 chỉ chép dòng 30/10.
12. **v6.9 #4** — Bot **chưa theo luật v6.9** (`server/internal/bot/bot.go:62-95`): Ru Ngủ/Hút Mì dùng nhánh mặc định "có địch trong 2.5× tầm là dùng" (comment còn ghi "Lưới Mì"); Sư Phụ Mì không kiểm nón/địch nhiều kiếm; Shipper dùng bất cứ khi có địch gần (spec: Đài sắp phun / Tiên khí trong 8 U / bị khống chế).
13. **v6.9 #5** — test xong; **số liệu tỉ lệ chọn 10 nhân vật trước/sau chưa gửi**.
14. **v6.9 #8** — test Cô Ba xong (`stall_test.go`); bot Cô Ba chưa theo luật (spec: đặt chặn trước con mồi yếu / sau lưng khi chạy; code: có địch gần là dùng theo hướng mặt).
15. **v6.16 #3** — server kiểm định dạng theo chữ ký byte + trần 300 KB (`server/internal/admin/feedback.go:57,134`), nhưng **chỉ 1 ảnh, không mã hoá lại → EXIF/GPS không bị xoá**, admin không có PreviewGroup.
16. **v6.16 #9** — kill feed đã gửi mã + tham số (`netproto.Feed`), danh hiệu gửi id; nhưng tên khu vẫn gửi chuỗi tiếng Việt (`ZoneName`), và client không có lớp dịch.
17. **v6.16 #12** — chỉ có trang chủ tiếng Việt với title/description/OG (`client/index.html:18-39`) + ảnh chụp `client/public/shots/`; **không có /en/ /pt/ /es/ /id/ /th/, hreflang, sitemap.xml**.

### Danh sách CHƯA
1. **v6.16 #1** — Nút "Thêm ảnh", kéo-thả, Ctrl+V, tối đa 3 ảnh (form chỉ có 1 ảnh chụp màn hình — `client/src/ui/feedback.ts:195`).
2. **v6.16 #2** — Thu nhỏ ≤1600px, nén WebP ≤300 KB, báo lỗi HEIC.
3. **v6.16 #4** — Kiểm Playwright WebKit iPhone + Chromium Android cho tải ảnh.
4. **v6.16 #5** — Admin React + antd v5 (admin vẫn là JS thuần `client/admin/admin.js`, không có antd).
5. **v6.16 #6** — Dashboard Card/Statistic, antd Table, Modal.confirm.
6. **v6.16 #7** — Bản đồ toàn cảnh bằng bộ vẽ game; vẫn còn luồng vị trí 5 Hz (`server/cmd/server/livemap.go:15-30`).
7. **v6.16 #8** — `client/src/i18n/vi.json` + CI bắt chuỗi cứng (không có thư mục `i18n/`; chuỗi vẫn ở `ui/strings.ts`).
8. **v6.16 #10** — Bản dịch en, pt-BR, es, id, th + `i18n/NOTES.md`.
9. **v6.16 #11** — Chọn ngôn ngữ (tự động/cờ/`?lang=`), Noto Sans Thai.
10. **v6.19 #2** — Gói Leaderboard dòng mới `{id,name,character,title_id,rank,zone_id,realm,tier,score,stealth}` trong schema v2 + golden (hiện `LeaderEntry` chỉ có id/nm/xp/L/sc/ti, msgpack — `server/internal/netproto/messages.go:275-282`; không có trong `proto/schema.txt`).
11. **v6.19 #3** — Giao diện §11.5 (thẻ Top 1 viền vàng/vương miện vẽ/chân dung, Top 2/3 bạc/đồng, khu + viên thuốc cảnh giới). Hiện là danh sách chữ đơn giản (`client/src/ui/hud.ts:478-497`).
12. **v6.19 #4** — Hoạt ảnh đổi hạng, thu gọn trên điện thoại.
13. **v6.19 #5** — Giới hạn ≤300px + cắt "…", tắt hiệu ứng ở chất lượng Thấp, ảnh duyệt WebKit/PC (chỉ có cắt dòng theo nửa màn hình — chưa phải yêu cầu §11.5).

### Có làm nhưng người chơi KHÓ/KHÔNG thấy
- **Toàn bộ v6.0 (netcode), v6.2 (telemetry), v6.4 A (dung lượng), mapgen --check**: thay đổi hạ tầng, chỉ thấy gián tiếp (ít giật, tải nhanh hơn). Telemetry chỉ có trong build `TELEMETRY=on` (deploy.sh đã bật).
- **v6.4 B / v6.6 (Tiên khí)**: chỉ thấy khi đã vào Tiên giới (Chân Tiên trở lên). Viên 100/500 có art riêng.
- **v6.5 / v6.9 / v6.10 / v6.11 (kỹ năng)**: chỉ dùng được khi nhặt **Bí Kíp** (weight 10) và chơi đúng nhân vật đó.
- **Lớp mặt v6.1 #8**: art đã có cho 3 nhân vật nhưng client chưa dùng → không hiện.
- **Trang quản trị** cần `server/data/admin.env` (chạy `go run ./cmd/adminpass`), nếu không thì tắt.
- Ghi chú nhỏ: vật cản bị chém vẫn phát VFX `death_smoke` (`client/src/render/scene.ts:1110-1112`) — spec v6.7 ghi "xoá ngay, không anim, 1 SFX bụp". Mệnh giá viên Tiên khí 500/100/10/2/1 và trần 40 là **hằng số trong code** (`server/internal/game/altar.go:114,117`), không nằm trong `config/game.json` (trái quy tắc 2 của CLAUDE.md).

---

## CHANGES_v6.md — NETCODE / Protocol v2

| Mục | Trạng thái | Bằng chứng / còn thiếu |
|---|---|---|
| 1. Đọc spec, gửi phân tích + kế hoạch | XONG | `docs/NETCODE_PLAN.md` |
| 2. Bước 1–7 §8 | MỘT PHẦN | B1 ✔ `client/src/render/viewpool.ts`, Despawn `server/internal/room/aoi.go:16`, DescRequest, cấm placeholder build thật `client/src/render/assets.ts:261`. B3 ✔ `proto/schema.txt` (v2, 13 message), keyframe `server/internal/room/client.go:113-122`, `cmd/protogen`. B2 Room Shared Cache ✘ (không có RoomFull/RoomDelta/version). B4 encode-once ✘. B5 ZoneReady/IndexedDB ✘ (0 kết quả grep). B6 Web Worker ✘ (không có `new Worker`), cache texture vòng ✘ (dùng ParticleContainer). B7 Service Worker ✘. |
| 3. v2 song song v1 sau cờ | MỘT PHẦN | v1 đã bị thay hẳn (`e8717f1`), không có cờ `proto=1`; bắt tay `pv` + `SchemaHash` (`server/internal/room/room.go:450`) |
| 4. Bài §7 thành test 1 lệnh | MỘT PHẦN | `tools/browsertest/flicker.mjs`, `netcode-bench.mjs`; thiếu bài đổi khu/mất mạng/SW |
| 5. CI chặn build thiếu asset bắt buộc | XONG | `.github/workflows/ci.yml:73` `check:assets`; `tools/deploy.sh:210` |
| 6. PERF_LOG + `perfreport` S1–S3 | MỘT PHẦN | `docs/PERF_LOG.md`, `docs/perf_log.csv`, `server/cmd/perfreport/main.go:1-10` — chỉ S1 |
| 7. Mốc chuẩn #4 S1–S5 | XONG | `docs/PERF_LOG.md:70-73` |
| 8. Mỗi bước = 1 lượt PERF_LOG §5 | MỘT PHẦN | chỉ các bước đã làm |
| 9. CLAUDE.md quy tắc 12 | XONG | `CLAUDE.md` quy tắc 12 |

## CHANGES_v6_1.md — Tên bot + hoạt ảnh

| Mục | Trạng thái | Bằng chứng / còn thiếu |
|---|---|---|
| 1. Tên từ `bot_names.txt`, không lặp, nhớ 10 phút, 15% biến thể | XONG | `server/internal/config/config.go:585`; `server/internal/room/botnames.go:14-73`; test `botnames_test.go` |
| 2. Chết: 70/30, 5% đứng yên, phân bố nhân vật | MỘT PHẦN | 70% cùng tên 4–12s `server/internal/room/bots.go:269-280`; phân bố `botnames.go:117`. **Thiếu 5% đứng yên** (`IdleAfterJoin*` chỉ khai báo `config.go:413-414`); vào lại sau 2s chứ không 5–20s |
| 3. Font dự phòng, 1 000 tên đủ dấu | MỘT PHẦN | Bộ ký tự `client/src/render/fonts.ts:22-29` phủ đủ 1 000 tên; **không có font dự phòng** |
| 4. Rà chỗ lộ bot | MỘT PHẦN | Hồi sinh/tên đã sửa; **`FlagBot` vẫn gửi xuống client** `server/internal/room/client.go:430` |
| 5. Phím J + điểm độ giật | XONG | `docs/JITTER_REPORT.md`; `client/src/render/jitter.ts` |
| 6. C — bộ điều khiển anim | XONG | `client/src/render/animctl.ts:13-45` (crossfade 0.1s, giữ pha, fps theo tốc độ, lật có trễ, nghiêng, squash, nhún, thở) |
| 7. B — chuẩn hoá frame | XONG | `tools/atlas-pack`, `server/cmd/artgen/normalize.go`, `jitter.go`, `preview.go` |
| 8. A — một thân nhiều mặt | MỘT PHẦN | artgen `bodyface.go`; art mặt 3 nhân vật; **client chưa ghép lớp mặt** |
| 9. Đo lại + PERF_LOG + GIF | MỘT PHẦN | GIF `docs/anim_preview/`; **PERF_LOG không có mục "Chất lượng hoạt ảnh"** |

## CHANGES_v6_2.md — Telemetry

| Mục | Trạng thái | Bằng chứng / còn thiếu |
|---|---|---|
| 1. Client telemetry sau cờ build | XONG | `client/src/telemetry/`; `client/vite.config.ts:46-49`; `check:telemetry-off` (số đo FPS ≤1 không kiểm lại) |
| 2. Server message + `/telemetry/end` + DB | XONG | `server/cmd/server/http.go:57`; `server/internal/telemetry/telemetry.go` |
| 3. `perfreport --real`, `/admin/perf`, tổng hợp 12:00/23:30 | XONG | `server/cmd/perfreport/main.go:89`; `http.go:58`; `server/internal/telemetry/schedule.go:16`; `docs/perf/real/` |
| 4. S6 + cột `n_sessions` | XONG | `docs/PERF_LOG.md:17`; header `docs/perf_log.csv` |
| 5. Deploy + thu thật | XONG | `docs/perf/real/2026-10-30_2330.md` (N=68) |
| 6. Ghi build id, chép S6 mỗi phiên | MỘT PHẦN | build id ✔ (`tools/deploy.sh`); bảng §3 thiếu 2 dòng S6 ngày 31/10 (CSV có) |

## CHANGES_v6_4.md — Dung lượng tải + Tiên khí ×20

| Mục | Trạng thái | Bằng chứng / còn thiếu |
|---|---|---|
| 1. `assets:size` | XONG | `client/scripts/assets-size.ts` |
| 2. `tools/asset-pipeline` | XONG | `tools/asset-pipeline/index.mjs`, `groups.mjs` |
| 3. Cache immutable + gzip | XONG | `server/cmd/server/http.go:62,76` |
| 4. Loader theo nhóm, chờ 2s qua cổng | XONG | `client/src/render/assets.ts:9,227`; `client/src/main.ts:14` |
| 5. Màn hình tải HTML theo byte | XONG | `client/src/ui/loading.ts` (mẹo, thử lại, nhân vật chạy); icon % nền `assets.ts:222` |
| 6. `assets:budget` vào CI | XONG | `client/scripts/assets-budget.ts:19-20`; trong `npm run build` |
| 7. Đo lại 4G | XONG | `tools/browsertest/loadtime.mjs` |
| 8. Trần Tiên khí | XONG | thay bằng v6.6 — `config/game.json:27-32` |
| 9. Thiêu Đốt Tiên giới | XONG | thay bằng v6.6 — `config/game.json:128` |
| 10. Viên 500/100/10/2/1, ≤40, sprite 100/500 | XONG | `server/internal/game/altar.go:114,117`; `client/src/render/catalog.ts:225-228`; `assets/src/items/tien_orb_{100,500}.png` (⚠ hằng số trong code, không trong config) |
| 11. Test §17 | XONG | theo số v6.6 — `burn_tien_test.go:197-242` |

## CHANGES_v6_5.md — Kỹ năng theo U

| Mục | Trạng thái | Bằng chứng / còn thiếu |
|---|---|---|
| 1. `U = pickup_r × scale`, mọi phạm vi theo `ranges_u` | XONG | `server/internal/game/skillrange.go:19-31`; `config/game.json:499-522` |
| 2. Đẩy theo U người dùng, 0.3s | XONG | `skillrange.go:48-54` |
| 3. VFX phạm vi cùng giá trị | XONG | `client/src/render/skillfx.ts:9-10` |
| 4. Tiên giới 3 lượt, rớt cắt còn 2, 3 chấm | XONG | `server/internal/game/skills.go:52-80`; `game.go:292`; `client/src/ui/hud.ts:450-459` |
| 5. Test + không còn hằng số phạm vi | XONG | `skillrange_test.go` (`TestNoHardcodedSkillRanges`) |

## CHANGES_v6_6.md — Tiên khí ×10

| Mục | Trạng thái | Bằng chứng / còn thiếu |
|---|---|---|
| 1. Trần [1000…15000] | XONG | `config/game.json:27-32` |
| 2. Burn 0.008 | XONG | `config/game.json:128` (khối `burn`) |
| 3. Mệnh giá giữ nguyên | XONG | `altar.go:114` |
| 4. Test §17 | XONG | `burn_tien_test.go:67,74,197-215,242` |
| 5. Ước tính lên Đạo Tổ ≈30 phút | XONG | PERF_LOG #13 (31 phút) |

## CHANGES_v6_7.md (+ v6.8) — Vật cản theo bậc + bản đồ

| Mục | Trạng thái | Bằng chứng / còn thiếu |
|---|---|---|
| 1. Bỏ HP; thêm khoá config | XONG | `config/game.json:256-293` |
| 2. Đủ bậc xoá ngay; thiếu → tia lửa + "Cần…" + viền tối | XONG | `server/internal/game/combat.go:513-562`; `client/src/render/scene.ts:1103-1108`; viền `mapview.ts:181,245` (⚠ vẫn có VFX khói khi phá `scene.ts:1110`) |
| 3. Mọc lại 30s, nhấp nháy 3s | XONG | `config/game.json:260-261`; `scene.ts:1114-1118` |
| 4. Loại `border` | XONG | `combat.go:514` `TierBorder`; 4 `border` mỗi `.tmj` |
| 5. Bot +0.3s | XONG | `server/internal/bot/bot.go:28-29` |
| 6. `mapgen --check` vào CI | XONG | `server/cmd/mapgen/check.go`; `.github/workflows/ci.yml:43`; `maps/check.json` bad=0, r=0..13 |
| 7. Sinh lại 5 khu | XONG | `maps/*.tmj`; mật độ theo mục tiêu mới v6.14 (8–14%) |
| 8. Hướng sửa từng khu §8.0 | XONG | các lượt bố cục (`docs/map_check/_duyet_*`); không kiểm lại từng số đo |
| 9. Gửi ảnh duyệt | XONG | `docs/map_check/` (41 ảnh) |
| 10. `min_realm` 0–13 | XONG | `config/game.json:262-278` |
| 11. Cương Kiếm +1 | XONG | `combat.go:518-521` |
| 12. 8 loại mới + art | XONG | `assets/src/obstacles/*` (16 thư mục, đủ art); alias dự phòng `catalog.ts:114` |
| 13. Mỗi khu ≥1 vật cản bậc ≥9 | XONG | đếm trong `.tmj`: mọi khu có `boulder_huge`/`ice_eternal`/`jade_pillar`/`seal_stele` |

## CHANGES_v6_9.md (+ v6.10, v6.11) — Làm lại kỹ năng

| Mục | Trạng thái | Bằng chứng / còn thiếu |
|---|---|---|
| 1. Ru Ngủ Trưa | XONG | `server/internal/game/skills.go:192-214`; huỷ khi bị chém `combat.go:66-67`; va kiếm một chiều `combat.go:395`; test `skills_test.go:116-200` |
| 2. Hút Mì Đoạt Kiếm | XONG | `skills.go:250-317`; client `scene.ts:984-997`; test `skills_test.go:202-262` |
| 3. Gom Đơn Hỏa Tốc | XONG | `skills.go:339-352`; hút đồ bỏ qua bịt mắt `items.go:89-113`; test `skills_test.go:385-460` |
| 4. Bot theo luật mới | MỘT PHẦN | `bot.go:62-95` vẫn luật cũ chung chung |
| 5. Test + số liệu tỉ lệ chọn | MỘT PHẦN | test ✔; số liệu telemetry chưa gửi |
| 6. Bày Hàng Chặn Lối | XONG | `server/internal/game/stall.go:46-195`; `client/src/render/stallview.ts` |
| 7. Config sạp, hướng nhắm | XONG | `config/game.json:508-514` |
| 8. Test + bot | MỘT PHẦN | test `stall_test.go` ✔; bot Cô Ba chưa theo luật |
| 9. Sạp +1 bậc, chủ đi xuyên | XONG | `stall.go:54-58,149`; test `stall_test.go:32,86` |

## CHANGES_v6_16.md — Góp ý ảnh, admin antd, i18n, tên game

| Mục | Trạng thái | Bằng chứng / còn thiếu |
|---|---|---|
| 1. Thêm ảnh / kéo-thả / dán, tối đa 3 | CHƯA | `client/src/ui/feedback.ts:195` chỉ 1 ảnh chụp |
| 2. Thu nhỏ + WebP ≤300 KB, HEIC | CHƯA | — |
| 3. Server kiểm nội dung, mã hoá lại xoá EXIF | MỘT PHẦN | chữ ký byte + 300 KB `server/internal/admin/feedback.go:57,134`; không mã hoá lại, 1 ảnh |
| 4. Kiểm Playwright WebKit/Android | CHƯA | — |
| 5. Admin React + antd | CHƯA | `client/admin/admin.js` JS thuần |
| 6. Dashboard/Table/Modal antd | CHƯA | — |
| 7. Bản đồ toàn cảnh, bỏ luồng 5 Hz | CHƯA | `server/cmd/server/livemap.go:15-30` vẫn 5 Hz |
| 8. `i18n/vi.json` + CI | CHƯA | không có `client/src/i18n/` |
| 9. Server gửi mã sự kiện | MỘT PHẦN | `netproto.Feed` có Kind + tham số; `ZoneName` vẫn là chuỗi |
| 10. Bản dịch 5 ngôn ngữ | CHƯA | — |
| 11. Chọn ngôn ngữ + font Thai | CHƯA | — |
| 12. SEO đa ngôn ngữ | MỘT PHẦN | chỉ trang vi `client/index.html:18-39`; không hreflang/sitemap |
| 13. Brand config + hiển thị | XONG | `config/game.json:2-7`; `client/public/brand/`; artgen logo `server/cmd/artgen/catalog.go:785` (khẩu hiệu ngôn ngữ khác chờ i18n) |
| 14. Bỏ "Kiếm Xoay Tu Tiên" + CI | XONG | `client/scripts/brand-check.ts` trong `npm run build`; grep 0 kết quả |
| 15. Đổi tên project | XONG | `server/go.mod` module `bladeascend`; `client/package.json` `bladeascend-client`; repo đã đổi (`cd777f8`) |

## CHANGES_v6_19.md — Bảng xếp hạng phòng

| Mục | Trạng thái | Bằng chứng / còn thiếu |
|---|---|---|
| 1. Danh hiệu theo cả phòng + test | XONG | `4d2d52e`; `server/internal/room/title_room_test.go:88,131` |
| 2. Gói Leaderboard mới trong schema v2 | XONG 1/11 | `proto/schema.txt` LbHead/LbRow + golden/fuzz `cmd/protofixture`; test `room/lbwire_test.go` |
| 3. Giao diện thẻ Top 1/2/3 | XONG 1/11 (chờ duyệt ảnh) | `client/src/ui/roomboard.ts` |
| 4. Hoạt ảnh đổi hạng, thu gọn mobile | XONG 1/11 | trượt FLIP 0.25s, ▲▼ 2s, Top 1 rung + lật + chớp vàng; điện thoại thu gọn |
| 5. ≤300px, i18n, chất lượng Thấp, ảnh duyệt | MỘT PHẦN | ≤300px + cắt "…" + Thấp tắt hiệu ứng: xong; i18n chờ v6.16 C |
