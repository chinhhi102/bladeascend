# Rà trạng thái toàn bộ việc đã giao — 31/10/2026

> Lập theo yêu cầu "Bước 7". Đối chiếu **code đang chạy** với từng mục trong các tệp
> `docs/_raw/CHANGES_*.md` và `NETCODE_SPEC`. Mỗi dòng có **bằng chứng** (tệp, khoá config,
> hoặc số đo) chứ không phải nhận định.
>
> Cách đọc trạng thái:
> **Xong** — làm đủ, có test hoặc số đo · **Một phần** — làm được phần chính, còn hở rõ ·
> **Chưa làm** — chưa có gì · **Làm khác spec** — có làm, nhưng theo cách khác, kèm lý do.

---

## 1. CHANGES_v4

### Mục 25–28 — Tối ưu (§13.4, §13.6)

| # | Việc | Trạng thái | Bằng chứng |
|---|---|---|---|
| 25 | Vòng kiếm vẽ 1 RenderTexture rồi chỉ xoay sprite | **Làm khác spec** | `render/swordring.ts` dùng `ParticleContainer` + `Particle` với `{position,rotation,vertex,color}` động. Cùng mục đích (gom về rất ít lệnh vẽ) nhưng **tốt hơn cho ca này**: RenderTexture phải vẽ lại mỗi khi số kiếm đổi — mà số kiếm đổi liên tục trong lúc đánh nhau. Draw call 611 → 118 (PERF_LOG #2). |
| 25 | BitmapText · ParticleContainer · cull · LOD | **Xong** | `render/fonts.ts`, `render/orblayer.ts`, `mapview.ts` (cull chunk), `scene.ts` (`detail`) |
| 25 | Giải mã msgpack trong Web Worker | **Chưa làm** | Không có `new Worker` nào trong `client/src`. Mục tiêu hiệu năng đã đạt bằng đường khác (tick 11.45 → 1.56 ms, PERF_LOG #4), nên chưa cần. |
| 25 | Tự hạ chất lượng + tuỳ chọn Cao/Vừa/Thấp | **Xong** | `render/quality.ts` — `autoQualityByMemory`, phím V |
| 26 | Server: va chạm theo vòng, không cấp phát mỗi tick | **Xong** | `game/combat.go` dùng `Geom.Reach` lọc thô rồi mới xét; PERF_LOG #4 |
| 27 | Mạng: lượng tử hoá, delta, bỏ snapshot cũ | **Xong** | `netproto` dùng `float32`; `room/broadcast.go` `knownPlayers`; `LossyConn.SendLossy` |
| 28 | Đạt bảng §13.6 với loadtest | **Xong** | `cmd/loadtest`; PERF_LOG #4: tick 1.56 ✅ / p95 3.33 ✅ / 8.6 KB/s ✅ · **gói 6.8 KB vs ngân sách 6 KB ❌** → đó là lý do có Protocol v2 |

### Mục 29–33 — VFX (§3.6)

| # | Việc | Trạng thái | Bằng chứng |
|---|---|---|---|
| 29 | `pixi-filters` + `revolt-fx`, vfxLayer + bloom | **Xong** | có trong `client/package.json`; `render/vfx.ts` |
| 30 | Hào quang Tiên giới 4 lớp, Đạo Tổ âm dương | **Xong** | `render/tienaura.ts` |
| 31 | Đột phá: Shockwave + hit-stop + chớp trắng | **Xong** | `render/breakthrough.ts` dùng `ShockwaveFilter` |
| 32 | Kenney / Brackeys CC0 + CREDITS.md | **Làm khác spec** | `assets/src/CREDITS.md` có, nhưng art hiện dùng là **tự sinh bằng `cmd/artgen`**, không tải gói CC0. Không cần nữa. |
| 33 | Gắn vào Cao/Vừa/Thấp + đo FPS | **Một phần** | Đã gắn (`quality.ts`). Ảnh trước/sau 5 người Tiên giới **chưa gửi**. |

### Mục 34–36 — Thần thú luôn lượn lờ (§3.5)

| # | Việc | Trạng thái | Bằng chứng |
|---|---|---|---|
| 34 | Ở lại bay quanh chủ, elip ~4s, động tác ngẫu nhiên | **Xong** | `render/beast.ts` (18,8 KB) |
| 35 | Phản ứng (lóe khi nhặt Tiên khí, vệt đỏ khi Thiêu Đốt), Đạo Tổ 2 con lệch pha | **Xong** | `render/beast.ts` |
| 36 | Số đốt theo chất lượng 16/10/6, xa chỉ vẽ đầu + 4 đốt | **Xong** | `render/beast.ts` + `quality.ts` |

### Mục 37–41 — Bảng tên (§12.3)

| # | Việc | Trạng thái | Bằng chứng |
|---|---|---|---|
| 37 | Nền, ruy băng theo hạng, huy hiệu cảnh giới, ▲●▼ | **Xong** | `render/nameplate.ts` (13,7 KB) |
| 38 | Cỡ cố định PC/mobile, Be Vietnam Pro | **Xong** | `render/fonts.ts` |
| 39 | Declutter (< 40 px chỉ tên, > 12 bảng ẩn ruy băng xa) | **Xong** | `render/nameplate.ts` |
| 40 | Cache bảng tên thành texture, BitmapFont đủ dấu | **Xong** | `render/nameplate.ts`, `render/fonts.ts` |
| 41 | Hoạt ảnh lên/mất hạng, số lăn, đột phá chớp | **Xong** | `render/nameplate.ts` |

### Mục 42–47 — HUD + bản đồ sinh động

| # | Việc | Trạng thái | Bằng chứng |
|---|---|---|---|
| 42 | Chân dung HUD, viền cảnh giới, 4 biểu cảm | **Xong** | `ui/hud.ts` (`portrait`) |
| 43 | Mặt đất trộn texture theo nhiễu, bake chunk | **Xong** | `render/mapview.ts`, `GROUND_CHUNK = 1024` (giảm từ 2048 cho iPhone) |
| 44 | mapgen xuất vùng đa giác, **server dùng chính đa giác đó** | **Xong** | `cmd/mapgen/preview.go`; `world.Zone` |
| 45 | Decal 30–60/chunk, bóng vật cản, ao gợn sóng | **Xong** | `render/decals.ts` (`PER_CHUNK` 55/30/12 theo chất lượng) |
| 46 | Ambient theo biome, giới hạn 150/60/0 | **Xong** | `render/ambient.ts` |
| 47 | Chưa có art thì vẽ thủ tục; thiếu gì vào assets:report | **Xong** | `check:assets` + `check:dead` (mới, bước 6) |

---

## 2. CHANGES_v5 (mục 1–27)

| # | Việc | Trạng thái | Bằng chứng |
|---|---|---|---|
| 1 | F3 "hiệu chỉnh vị trí/giây"; debug H vẽ footprint 2 phía | **Xong** | `fixPerSec` trong `scene.perfStats()`; `toggleHitbox()` |
| 2 | Dùng chung 1 thuật toán va chạm, lệch < 0.5 | **Xong** | `cmd/collisionfixture` + `npm run check:collision`. Đã bắt thật hai lần: lệch 4.03 rồi 12.98 đơn vị, cả hai đã sửa. |
| 3 | Trượt dọc, lặp ≤ 3 lần/tick, hiệu chỉnh mượt | **Xong** | `game/movement.go` `for iter := 0; iter < 3; iter++` (2 chỗ) |
| 4 | 5 khu, mỗi khu 1 goroutine, chuyển ở ranh giới tick | **Xong** | `room/room.go` `stepZone` chạy song song; `applyTransfers` sau khi mọi khu xong |
| 5 | Cổng đích ngẫu nhiên toàn mạng | **Xong** | `config` `"exit_pool": "all"` |
| 6 | mapgen 5 thuật toán khu + `--zone --seed --png` | **Xong** | `cmd/mapgen/` — đã sinh lại 3 lượt bố cục trong tháng 10 |
| 7 | Băng trơn quán tính, bờ tuyết ×0.8 | **Xong** | `config` `ice: {friction 0.92, accel 468}` |
| 8 | 20 người/phòng, spawn vào khu ít người nhất | **Xong** | `room/manager.go` |
| 9 | HUD tên khu, minimap theo khu, M liệt kê 5 khu | **Xong** | `ui/hud.ts`, `ui/minimap.ts` |
| 10 | loadtest 20 người rải 5 khu, ghi tick từng khu | **Xong** | `cmd/loadtest` |
| 11 | Bí Kíp `skill_scroll` weight 10, tối đa 2 lượt | **Xong** | `config.items.weights.skill_scroll = 10` |
| 12 | Khung kỹ năng chung + hệ trạng thái có thời hạn, 10 kỹ năng | **Xong** | `game/skills.go` |
| 13 | Thằng Mập — Phun Trân Châu bịt mắt | **Xong** | `render/boba.ts` |
| 14 | Bot nhặt Bí Kíp và dùng kỹ năng | **Xong** | `room/bots.go` |
| 15 | Art placeholder + vào assets:report | **Xong** | `check:assets` |
| 16–18 | Thiên Môn: quảng trường, một chiều, bến cố định, minimap | **Xong** | `world.HeavenGate`; `config.heaven_gate` |
| 19 | Đại Ca — Gọi Đàn Em (2 đàn em 10s, 50% kiếm) | **Xong** | `game/skills.go`; test `skills_v55_test.go` |
| 20–25 | Chỉnh thời lượng/phạm vi các kỹ năng | **Xong** | `config.skills.ranges_u`; test v5.5 |
| 26 | Shipper xuyên vật cản, hết giờ đẩy ra mượt | **Xong** | `gom_don_r/secs` trong config |
| 27 | Cập nhật test | **Xong** | 19 tệp test trong `internal/game` |

---

## 3. CHANGES_v6_4 — dung lượng tải + Tiên khí ×20

| # | Việc | Trạng thái | Bằng chứng |
|---|---|---|---|
| 1 | `npm run assets:size` | **Xong** | `client/scripts/assets-size.ts` |
| 2 | `tools/asset-pipeline`: WebP, atlas, hash, manifest v2 | **Xong** | `assets/dist/manifest.json` — 462 tệp, 9,89 MB |
| 3 | Server phục vụ `assets/dist`, immutable + nén | **Xong** | `cmd/server/routes` |
| 4 | Loader theo nhóm boot/zone/chars/fx/audio | **Xong** | `manifest.groups` |
| 5 | Màn hình tải HTML thuần, thanh theo byte | **Xong** | `ui/loading.ts` |
| 6 | `assets:budget` vào CI (boot ≤ 8 MB) | **Xong** | trong `npm run build`. Đã chặn thật: 9,16 MB → xoá art chết → 4,52 MB |
| 7 | Đo lại thời gian vào trận 4G | **Xong** | `tools/browsertest/loadtime.mjs`; PERF_LOG nút thắt B9 |
| 8–11 | Tiên khí ×20 | **Thay bằng v6.6** | xem mục 5 dưới |

## 4. CHANGES_v6_5 — kỹ năng theo phạm vi U

| # | Việc | Trạng thái | Bằng chứng |
|---|---|---|---|
| 1 | `U = pickup_r × scale`, mọi phạm vi theo `ranges_u` | **Xong** | `config.skills.range_unit: "pickup_r"`; 22 khoá trong `ranges_u` |
| 2 | Đẩy lùi theo U của **người dùng**, 0.3s, trượt dọc | **Xong** | `config.skills.push_secs: 0.3` |
| 3 | VFX/chỉ báo phạm vi vẽ theo cùng giá trị | **Xong** | `render/skillfx.ts` |
| 4 | Tiên giới 3 lượt, rớt Đại Thừa cắt còn 2 | **Xong** | `config.skills.max_charges_tien: 3` |
| 5 | Test + grep không còn hằng số phạm vi | **Xong** | test §17 |

**Ghi chú sửa lỗi:** ba kỹ năng từng nhân `scale` **hai lần** (`skillRange` đã gồm scale, rồi
`targetsNear`/`inCone` nhân lại). Nón của Bà Tám ở Đạo Tổ ra 19 440 đơn vị trên bản đồ
12 000 — tức là trúng cả khu. Đã sửa.

## 5. CHANGES_v6_6 — Tiên khí ×10 (kiểm config ĐANG CHẠY)

| # | Việc | Spec yêu cầu | Config đang chạy | Trạng thái |
|---|---|---|---|---|
| 1 | `tien_khi_ceilings` | `[1000, 3000, 6000, 10000, 15000]` | `[1000, 3000, 6000, 10000, 15000]` | **Xong ✓** |
| 2 | `burn_tien_frac_per_sec` | `0.008` | `burn.tien_frac_per_sec = 0.008` | **Xong ✓** (tên khoá đặt ở khối `burn` cho khớp `burn_exp_per_sec`; đã ghi lý do trong `config.go`) |
| 3 | Mệnh giá viên 500/100/10/2/1, tối đa 40 | — | giữ nguyên | **Xong** |
| 4 | Test theo checklist §17 | — | — | **Xong** |
| 5 | Ước tính thời gian lên Đạo Tổ ôm 2 đài | ≈ 30 phút | **31 phút** (từ 62) | **Xong** — PERF_LOG #13 |

## 6. CHANGES_v6_7 / v6.8 — vật cản theo bậc

| # | Việc | Trạng thái | Bằng chứng |
|---|---|---|---|
| 1 | Bỏ HP/damage/nứt vỡ; thêm `min_realm`, `corridor_min`… | **Xong** | config không còn `hp`/`dmg_cooldown_secs`/`dmg_per_extra_ring`; có đủ 4 khoá mới |
| 2 | Đủ bậc → xoá ngay; chưa đủ → tia lửa + "Cần {cảnh giới}" | **Xong** | `cut_hint_cooldown_secs: 3.0` |
| 3 | Mọc lại 30s, nhấp nháy 3s trước | **Xong** | `game/obstacles.go` |
| 4 | **Biên khu thành loại `border` riêng** | **Xong** | `TierBorder = 99`. Đây là lỗi thật: biên từng là `stone_wall` bậc 5 nên **Luyện Hư chém được biên và đi ra ngoài bản đồ**. |
| 5 | Bot: chém được = đi qua được (+0.3s) | **Xong** | `room/bots.go` |
| 6 | `mapgen --check` r = 0..13, vào CI | **Xong** | `cmd/mapgen/check.go`; chạy trong `tools/deploy.sh` |
| 7–9 | Sinh lại 5 khu theo §8.1b, gửi ảnh duyệt | **Xong** | 3 lượt bố cục + 1 lượt tinh chỉnh; khoảng thở 0% → 28–57%, không còn túi kín |
| 10 | `min_realm` trải 0–13 | **Xong** | 15 loại từ `fence: 0` tới `seal_stele: 13` |
| 11 | Cương Kiếm chém bậc ≤ cảnh giới + 1 | **Xong** | `hard_buff_extra_tier: 1` |
| 12 | 8 loại vật cản mới + `border` | **Xong** | đủ 8 loại trong `min_realm`; art dùng alias mượn tới khi artgen sinh xong |
| 13 | Mỗi khu ≥ 1 vật cản bậc ≥ 9 | **Xong** | `cmd/mapgen` |

## 7. CHANGES_v6_9 / v6.10 / v6.11 — làm lại kỹ năng

| # | Việc | Trạng thái | Bằng chứng |
|---|---|---|---|
| 1 | Ông Ba Võng — Ru Ngủ Trưa (niệm 0.5s, ngủ 2.5s, va kiếm một chiều) | **Xong** | `ru_ngu_r: 4.0`, `ru_ngu_secs: 2.5`, `ru_ngu_channel: 0.5` |
| 2 | Sư Phụ Mì — Hút Mì Đoạt Kiếm (nón 5 U, cướp 25%, trần 20%) | **Xong** | `hut_mi_cone: 5.0`, `hut_mi_steal: 0.25`, `hut_mi_cap_of_own_max: 0.2` |
| 3 | Shipper — Gom Đơn Hỏa Tốc (4s, hút 3 U, miễn khống chế) | **Xong** | `gom_don_r: 3.0`, `gom_don_secs: 4.0` |
| 4 | Bot dùng theo luật mới | **Xong** | `room/bots.go` |
| 5 | Test + **số liệu thật tỉ lệ chọn 10 nhân vật trước/sau** | **Một phần** | Test xong. **Số liệu telemetry chưa gửi** — spec ghi "sau 2–3 ngày", bản này deploy 31/10 nên chưa đủ ngày. |
| 6–8 | Cô Ba — Bày Hàng Chặn Lối (5 sạp, 6 U, 6s, vỏ chuối) | **Xong** | `bay_sap_dist/len/count/secs`, `vo_chuoi_w`; `ganh_dash`/`ganh_trail_w` đã bỏ |
| 9 | Sạp cao hơn 1 bậc, chủ đi xuyên | **Xong** | `bay_sap_tier_bonus: 1`, `bay_sap_owner_passes: true` |

---

## 8. Protocol v2 (NETCODE_SPEC)

Mục tiêu dừng mà anh đặt: **respawn = 0** và **gói < 6 KB**. Vế đầu **đạt**, vế sau **chưa**.

| Phần | Trạng thái | Bằng chứng |
|---|---|---|
| Schema chung + generator sinh cả hai phía | **Xong** | `cmd/protogen` sinh Go + TS từ `proto/schema.txt` (schema `64277c847bf551ea`, v2); `SCHEMA_HASH` chốt hai bên |
| Golden / đối chiếu dây | **Xong** | `cmd/protofixture` + `npm run check:wire` — **10 015 gói** (Spawn 1671 · Update 1670 · Desc 1669 · Despawn 1669 · SelfState 1668 · Input 1668) khớp cả hai chiều. Chạy lại hôm nay: OK. |
| Vòng đời Spawn / Desc / Update / **Despawn** | **Một phần** | **`Despawn` xong và tường minh**; `Spawn/Desc/Update` vẫn đi msgpack v1 |
| Delta theo `ack`, keyframe 2s | **Chưa làm** | — |
| Lượng tử hoá nhị phân | **Chưa làm** | v1 đang dùng `float32`, chưa phải lượng tử hoá theo §4.1 |
| **Cờ chọn v1/v2 để quay lui** | **Chưa làm — CỐ Ý** | Chưa có message v2 nào chạy trên dây để mà chọn. Làm cờ trước khi có thứ để bật là làm một công tắc không nối vào đâu. |

**Kết quả đo được (PERF_LOG + PROGRESS):**

- respawn **0 / 4 lượt quay lại thật**, placeholder **0 / 4 744 khung**, nhấp nháy do
  `Despawn` **= 0** → **vế thứ nhất của mục tiêu dừng: ĐẠT.**
- gói **6,9 KB**, ngân sách 6 KB → **vế thứ hai: CHƯA.**

**Mổ gói lớn nhất ra để biết nhắm vào đâu:**

| Phần | KB | % |
|---|---|---|
| `ev` (sự kiện) | 2.5 | 37% |
| `ia` (vật phẩm vào tầm) | 1.7 | 25% |
| còn lại | 2.7 | 38% |

Chỗ cần nhắm **không phải** state người chơi — `ev` + `ia` chiếm **62%** gói. Đó là việc
lớn, chưa làm trong lượt này.

---

## 9. Tóm tắt những gì CHƯA làm

**Mục nhỏ (đổi số/config) chưa làm:** không còn mục nào. Đã kiểm từng khoá config của
v6.4 → v6.11 ở trên; tất cả khớp spec.

**Mục lớn chưa làm — chờ anh xếp thứ tự** (cũng ghi ở `docs/QUESTIONS.md`):

| Việc | Vì sao chưa | Ở đâu |
|---|---|---|
| Protocol v2: chuyển `ev` + `ia` sang nhị phân | 62% gói nằm ở đây; cần làm trước rồi mới tới cờ v1/v2 | NETCODE_SPEC §4.1 |
| Giải phóng atlas nhân vật trên iOS | iPhone thật chạm **216/256 MB**, hạ chất lượng không trả lại bộ nhớ | Q30, nút thắt B13 |
| 34 tệp art không ai nạp | Trong đó `ui/icon_skill_scroll.png` nghi là **thiếu dây nối**, không phải art thừa | Q31 |
| Web Worker giải mã msgpack (v4 mục 25) | Mục tiêu hiệu năng đã đạt bằng đường khác | — |
| Ảnh FPS trước/sau 5 người Tiên giới (v4 mục 33) | Chưa chụp | — |
| Tỉ lệ chọn 10 nhân vật sau v6.9 (mục 5) | Spec ghi "sau 2–3 ngày"; bản này mới deploy 31/10 | — |
