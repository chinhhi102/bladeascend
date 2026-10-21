# KIẾM XOAY TU TIÊN — Game Design & Technical Spec (v2)

> Tên tạm (working title). File này dành cho Claude Code: đặt tại `docs/GAME_SPEC.md`.
> Mọi con số nằm trong `config/game.json` (mục 16) — **không hardcode**.
> Art/animation prompts: `docs/ART_PROMPTS.md`. Hướng dẫn setup: `docs/SETUP_GUIDE.md`.

**Thay đổi so với v1:** đổi stack (Go server + TypeScript/PixiJS client trên web) · kiếm có thể bị bể khi va nhau (tỉ lệ theo chênh lệch cảnh giới) · scale 1 → 5 và tầm kiếm dài theo · chết ra màn hình chờ rồi match lại · bản đồ có địa hình, vật cản phá được (mọc lại sau 30s), cổng dịch chuyển · không login, nhập tên là chơi, có leaderboard · bộ vật phẩm buff có thời hạn.

---

## 1. Tổng quan

- Thể loại: multiplayer `.io` top-down 2D chạy trên **trình duyệt** (PC + điện thoại), vào là chơi, không cài đặt, không đăng nhập.
- Fantasy: dân phố tấu hài tu tiên bằng cách cho kiếm bay vòng quanh người.
- Core loop: **di chuyển → chém quái/vật cản → lượm EXP & vật phẩm → lên cảnh giới (to ra, kiếm dài và nhiều hơn) → chém người khác → bị chém thì rớt tầng, mất kiếm → tận dụng địa hình để lật kèo**.
- Điều khiển: WASD / phím mũi tên / giữ chuột trên PC; joystick ảo trên điện thoại. Không có nút tấn công.

## 2. Luồng màn hình

```
[Trang chủ] nhập tên + chọn nhân vật + nút "Vào chơi" + nút "Bảng xếp hạng"
      │
      ▼
[Đang tìm phòng...] → [Trong trận] ──chết──▶ [Màn hình chờ / Tổng kết]
                                                 ├─ "Chơi tiếp"  → matchmaking lại (có thể sang phòng khác)
                                                 ├─ "Bảng xếp hạng"
                                                 └─ "Đổi nhân vật" → Trang chủ
```

- Tên 2–16 ký tự, cho phép tiếng Việt có dấu, lọc từ tục, cho trùng tên.
- Tên & nhân vật đã chọn được nhớ trên máy (localStorage) cho lần sau.

## 3. Hệ thống tu luyện

### 3.1 Cảnh giới, kiếm tối đa, kích thước

`level_index L = r*10 + s` (r = cảnh giới 0..8, s = tầng 0..9) → 90 level.

```
max_swords(r) = 5 * (r + 1)
scale(L)      = 1 + 4 * L / 89          // 1.00 (Luyện Khí 1) → 5.00 (Đại Thừa 10), to ra ở MỖI tầng
```

| r | Cảnh giới | Kiếm tối đa | Scale tầng 1 → tầng 10 |
|---|---|---|---|
| 0 | Luyện Khí | 5 | 1.00 → 1.40 |
| 1 | Trúc Cơ | 10 | 1.45 → 1.85 |
| 2 | Kim Đan | 15 | 1.90 → 2.30 |
| 3 | Nguyên Anh | 20 | 2.35 → 2.75 |
| 4 | Hóa Thần | 25 | 2.80 → 3.20 |
| 5 | Luyện Hư | 30 | 3.25 → 3.65 |
| 6 | Hợp Thể | 35 | 3.70 → 4.10 |
| 7 | Độ Kiếp | 40 | 4.15 → 4.55 |
| 8 | Đại Thừa | 45 | 4.60 → 5.00 |

### 3.2 EXP

```
exp_needed(r, s) = round(100 * 1.2^r * (1 + 0.08 * s))
```

| Cảnh giới | EXP/tầng (1 → 10) | Tổng cả cảnh giới |
|---|---|---|
| Luyện Khí | 100 → 172 | 1 360 |
| Trúc Cơ | 120 → 206 | 1 632 |
| Kim Đan | 144 → 248 | 1 958 |
| Nguyên Anh | 173 → 297 | 2 350 |
| Hóa Thần | 207 → 357 | 2 820 |
| Luyện Hư | 249 → 428 | 3 384 |
| Hợp Thể | 299 → 514 | 4 061 |
| Độ Kiếp | 358 → 616 | 4 873 |
| Đại Thừa | 430 → 740 | 5 848 |

- Lưu `total_exp` cộng dồn, level suy ra từ bảng ngưỡng tích lũy tính sẵn.
- Lên tầng: glow nhỏ. **Đột phá cảnh giới**: anim `breakthrough`, sét, camera rung, đổi skin kiếm theo cảnh giới, nhân vật quanh đó hiện emote kinh ngạc.

## 4. Kiếm, vòng quay, tầm đánh

Mọi kích thước của kiếm nhân với `scale` → nhân vật to thì vòng kiếm rộng hơn **và** từng thanh kiếm dài hơn.

```
ring_radius   = scale * max(70, n * 22 / (2π)) * buff_long_sword(1.3 | 1.0)
sword_length  = 40 * scale * buff_long_sword
sword_width   = 10 * scale
angle_i(t)    = phase + dir * ω * t + 2π * i / n        // dir = +1 | -1
ω             = 1.8 rad/s * buff_spin(1.6 | 1.0)
```

- Hitbox kiếm: **capsule** nằm theo hướng bán kính, từ `ring_radius - sword_length/2` đến `ring_radius + sword_length/2`, độ dày `sword_width`.
- Hitbox thân: hình tròn `BODY_R = 30 * scale`.
- Số kiếm tối thiểu luôn là **1**. Bắt đầu: 1 kiếm.
- Thêm/bớt kiếm: phân bổ lại góc bằng tween 0.2s.

### 4.1 Tốc độ di chuyển & camera

```
move_speed (đơn vị thế giới/s) = 260 * sqrt(scale) * buff_haste(1.4 | 1.0) * zone_mult
camera: chiều cao vùng nhìn thấy (đơn vị thế giới) = 900 * scale^0.85
```

→ Trên màn hình, nhân vật to lên dần nhưng vẫn nhìn được rộng hơn (bản đồ như "thu nhỏ" lại). Người nhỏ di chuyển trên màn hình nhanh hơn người lớn (≈260 px/s so với ≈150 px/s ở scale 5) → người nhỏ có cơ hội né và lật kèo.

## 5. Chiến đấu

Thứ tự xử lý mỗi tick server:

### 5.1 Kiếm A chạm thân B
(A ≠ B, B không bất tử, không có khiên, không đang dịch chuyển)
- Nếu B ở Luyện Khí tầng 1 (`L == 0`) → B **chết** (mục 7). Khi đó `loss = B.total_exp` hiện tại; A nhận 50% loss + **thưởng hạ sát `kill_bonus_exp` (30)**, rơi 30% loss thành viên EXP, cộng 1 kill cho A.
- Ngược lại: `loss = exp_needed(level ngay dưới level hiện tại của B)`; `B.total_exp = max(0, B.total_exp - loss)` → B rớt 1 tầng (có thể rớt cảnh giới).
- A nhận `50% * loss` EXP; rơi `30% * loss` dạng viên EXP quanh B.
- B bất tử **1s** (nhấp nháy), bị đẩy lùi khỏi A.
- Mỗi tick, mỗi nạn nhân tối đa 1 hit.
- **Rớt cảnh giới**: số kiếm > max mới → kiếm thừa rơi xuống đất thành item kiếm (ai lượm cũng được).
- Nếu B có buff Khiên: khiên vỡ, không mất EXP, vẫn có 1s bất tử.

### 5.2 Kiếm A chạm kiếm B — bể kiếm

Luôn: **cả hai vòng đổi chiều quay** (`dir *= -1`), VFX tia lửa + SFX "keng". Cooldown mỗi cặp (A, B) 0.3s.

Sau đó tính bể kiếm, với `d = |rA - rB|` (chênh lệch cảnh giới) và `k = d + 1` (tỉ lệ đổi kiếm):

| Trường hợp | Kết quả |
|---|---|
| Cùng cảnh giới (d = 0) | Mỗi bên mất 1 kiếm |
| Khác cảnh giới | Bên thấp mất `k` kiếm; bên cao nhận `crack += lost / k` (lost = số kiếm bên thấp thực sự mất). Khi `crack ≥ 1` → bên cao mất `floor(crack)` kiếm, `crack -= floor(crack)` |

Mọi trường hợp: số kiếm không bao giờ xuống dưới 1 (nếu đã còn 1 thì không mất thêm, chỉ đổi chiều).

Bảng tỉ lệ đổi kiếm (bên cao : bên thấp):

| Chênh lệch | Ví dụ | Tỉ lệ |
|---|---|---|
| 0 | Kim Đan vs Kim Đan | 1 : 1 |
| 1 | Trúc Cơ vs Luyện Khí | 1 : 2 |
| 2 | Kim Đan vs Luyện Khí | 1 : 3 |
| 3 | Nguyên Anh vs Luyện Khí | 1 : 4 |
| 8 | Đại Thừa vs Luyện Khí | 1 : 9 |

Ví dụ Luyện Khí (r=0) vs Nguyên Anh (r=3) → k = 4: mỗi lần va, Luyện Khí mất 4 kiếm, Nguyên Anh nhận 1 vết nứt → mất 1 kiếm. Nếu Luyện Khí chỉ còn 3 kiếm (mất được 2 vì tối thiểu giữ 1) → Nguyên Anh chỉ nhận 2/4 vết nứt, chưa mất kiếm.

`crack` lưu trên vòng kiếm, reset về 0 khi chết. Hiển thị vết nứt nhẹ trên vòng kiếm khi `crack > 0`.

Kiếm bị bể: VFX vỡ vụn + SFX, mảnh văng ra (không thành item). Chủ nhân hiện emote "Kiếm của taaa!".

### 5.3 Kiếm chạm quái / vật cản
- Quái: 1 damage mỗi lần chạm, cooldown mỗi cặp (kiếm, quái) 0.25s. Quái chết → **toàn bộ EXP rơi thành viên EXP** tại chỗ. Trong `orb_owner_secs` (1.5s) đầu, chỉ người chém hạ quái được nhặt; sau đó ai cũng nhặt được.
- Vật cản: `1 + r` damage (cảnh giới càng cao chém càng nhanh), cooldown mỗi cặp 0.25s. **Kiếm xuyên qua vật cản và không bị mất/bể** khi chém vật cản.
- Kiếm không tương tác với chính chủ.

## 6. Vật phẩm

| Item | Hiệu ứng | Thời hạn | Weight spawn |
|---|---|---|---|
| Kiếm | +1 kiếm (vĩnh viễn). Nếu đã max → +20 EXP | – | 40 |
| Tốc Kiếm | ω × 1.6 | 12s | 15 |
| Thần Hành | Tốc chạy × 1.4 | 10s | 15 |
| Nam Châm | Bán kính hút đồ × 3 | 15s | 8 |
| Trường Kiếm | Vòng kiếm & chiều dài kiếm × 1.3 | 12s | 7 |
| Kim Cang (khiên) | Chặn 1 lần bị chém vào thân | 8s hoặc tới khi vỡ | 7 |
| Song Bội Đan | EXP nhận × 2 | 20s | 5 |
| Ẩn Thân Phù | Người khác không thấy mình (trừ khi ở rất gần: 120 × scale của họ); hết hiệu lực ngay khi kiếm mình chém trúng người | 6s | 3 |
| Viên EXP nhỏ/vừa/lớn | +5 / +20 / +60 EXP | – | rơi từ quái |
| **Rương Báu** (hiếm) | Người đầu tiên **chạm vào** → rương nổ ra 5 item ngẫu nhiên văng xung quanh (ai lượm cũng được) | – | 1 rương mỗi 2 phút, báo trước toàn phòng 10s |

- Nhặt lại buff đang có → làm mới thời gian (không cộng dồn).
- HUD hiển thị icon buff + vòng đếm ngược. Mỗi buff có aura màu riêng trên nhân vật để người khác "đọc" được (trừ Ẩn Thân).
- Bán kính hút đồ mặc định: `60 * scale`.
- Item spawn cách người chơi ≥ 300 đơn vị, mật độ theo config.

## 7. Chết & màn hình chờ

- **Rơi kiếm khi chết** (người và bot): `floor(n * death_sword_drop)` kiếm (mặc định 50%) văng ra thành item kiếm quanh xác. Ví dụ chết với 5 kiếm → rơi 2.
- **Người thật**: anim `death` 1s → rời phòng → **Màn hình Tổng kết** hiển thị:
  - Bị ai chém (tên + nhân vật), câu tấu hài random ("Bạn đã được tiễn về quê nuôi cá").
  - Cảnh giới cao nhất đạt được, tổng EXP đỉnh, số người đã chém, thời gian sống, hạng trong phòng.
  - Nút: **Chơi tiếp** (matchmaking lại ngay, giữ tên + nhân vật), **Bảng xếp hạng**, **Đổi nhân vật**.
- Khi vào lại: Luyện Khí tầng 1, 1 kiếm, không buff, bất tử 2s, spawn ở chỗ an toàn (không có ai trong 800 đơn vị).
- **Bot**: chết thì xử lý theo mục 9.3 (respawn tại chỗ hoặc bị xóa).
- Kill feed tấu hài trong trận: `"{A} vừa tiễn {B} về quê nuôi cá"`, `"{B} bị {A} chém rớt xuống {realm}"`, `"{A} đột phá {realm}, cả phòng run rẩy"`, `"{A} chui cổng ra liền bị {B} chém, đen thôi đỏ quên đi"`.

## 8. Bản đồ & địa hình

### 8.1 Tổng quan
- Mỗi phòng dùng 1 bản đồ, chọn ngẫu nhiên từ danh sách. Kích thước 12 000 × 12 000, tường biên không phá được.
- Bản đồ vẽ bằng **Tiled** (phần mềm miễn phí), xuất file `.tmj` (JSON). Server đọc object layers, client đọc tile + decor.
- 3 bản đồ ban đầu: **Làng Tre** (cân bằng), **Chợ Đêm** (nhiều hẻm, sạp hàng), **Núi Đá** (rộng, đá tảng lớn, mạng cổng dịch chuyển).

### 8.2 Vật cản (chặn thân, có độ cứng)

| Loại | HP | Ghi chú |
|---|---|---|
| Hàng rào tre | 30 | Mỏng, dài |
| Sạp hàng / xe đẩy | 60 | Vỡ ra: 30% rơi 1 item |
| Tường gạch | 150 | |
| Đá tảng | 400 | Rất lớn |
| Tường đá cổ / biên map | ∞ | Không phá được |

- Hình dạng: hình chữ nhật hoặc hình tròn (lấy từ Tiled). Thân nhân vật bị đẩy ra khi chạm. Quái cũng bị chặn.
- Kiếm đi xuyên vật cản và gây damage (mục 5.3).
- Hiển thị 3 trạng thái theo HP: nguyên vẹn (>66%), nứt (33–66%), sắp vỡ (<33%).
- **Vỡ** → thành đống đổ nát, đi qua được. Sau **30s** mọc lại:
  - 3s trước khi mọc: hiệu ứng nhấp nháy cảnh báo.
  - Nếu có nhân vật/quái đứng trong vùng lúc mọc → chờ tới khi trống rồi mới mọc (không kẹt người).
  - Mọc lại với HP đầy.

### 8.3 Vùng địa hình (không chặn)

| Vùng | Hiệu ứng | Dùng để lật kèo |
|---|---|---|
| Bụi rậm | Ai đứng trong bụi thì người ngoài bụi không thấy (trừ khi đứng gần 150 × scale). Bị chém trúng thì lộ 2s | Phục kích, trốn |
| Bùn lầy | Tốc chạy × 0.6 | Dụ kẻ đuổi vào bùn |
| Đường gió | Lệch so với hướng gió < 60° → × 1.5 (theo gió); > 120° → × 0.8 (ngược gió); còn lại × 1.0 (ngang gió) | Đường thoát nhanh |
| Hẻm hẹp | Các lối đi rộng 150 / 250 / 350 giữa vật cản không phá được | Người to (thân rộng) không chui lọt, người nhỏ trốn vào được |

### 8.4 Cổng dịch chuyển
- Cổng đi theo cặp (A ↔ B) hoặc mạng 3–4 cổng (vào 1 cổng → ra ngẫu nhiên 1 cổng khác trong mạng).
- Kích hoạt khi tâm nhân vật đi vào bán kính cổng (150). Anim `teleport_out` 0.4s → xuất hiện ở cổng đích với anim `teleport_in` 0.4s.
- Cổng đích **lóe sáng 0.5s trước** khi người xuất hiện (cảnh báo nhẹ cho người đứng gần — vẫn dễ bị bất ngờ).
- **Không bất tử khi vừa ra cổng.** Trong lúc đang dịch chuyển (0.8s) không bị chém và không chém được.
- Cooldown mỗi người: 8s mới dùng cổng tiếp được.
- Người to cỡ nào cũng dùng được.

## 9. Phòng, matchmaking, bot

### 9.1 Phòng
- Tối đa **20 người thật**/phòng; bot không tính → tối đa ~40 nhân vật.
- Phòng tự hủy khi 0 người thật trong 60s.

### 9.2 Matchmaking
```
on_join(player):
  candidates = rooms where humans < 20
  room = candidate có humans lớn nhất; hòa → phòng tạo sớm nhất
  if none: room = create_room(random_map)
```

### 9.3 Bot
```
every 2s per room:
  if humans < 20 and humans + bots < 20:
      spawn min(20 - humans - bots, 2) bots
on human join/leave:  KHÔNG xóa bot
on bot death:
  if humans + (bots - 1) >= 20: remove bot
  else: respawn bot sau 3s
```
- Bot spawn với level trọng số: 60% Luyện Khí, 30% Trúc Cơ, 10% Kim Đan; số kiếm ngẫu nhiên [1, max].
- Bot dùng tên + nhân vật ngẫu nhiên, **không vào leaderboard toàn cục**.

## 10. Bot AI

Quyết định lại mỗi 0.3–0.6s (random), tìm đường bằng A* trên lưới 100 đơn vị (cập nhật khi vật cản vỡ/mọc).

| State | Điều kiện | Hành vi |
|---|---|---|
| Farm | Mặc định | Tới quái, item, vật cản yếu gần nhất |
| Hunt | Mục tiêu trong 600 × scale có `power` < 0.8× mình | Đuổi, nhắm vòng kiếm vào thân; tránh lao kiếm vào kiếm đối thủ cùng/cao cảnh giới hơn |
| Flee | Có kẻ `power` > 1.25× trong 500 × scale | Chạy ngược hướng; ưu tiên hẻm hẹp (nếu mình nhỏ hơn), bụi rậm, cổng |
| Ambush | Mình nhỏ, gần bụi rậm/cổng | Nấp chờ kẻ yếu đi ngang |
| Wander | Không có gì | Đi ngẫu nhiên |

- `power = sword_count * (r + 1) * ring_radius`.
- Tham số tính cách random mỗi bot: `reaction_delay` 0.2–0.6s, `aim_noise`, `aggression`.
- Bot chỉ "thấy" những gì người chơi cùng vị trí thấy được (tôn trọng bụi rậm, ẩn thân).

## 11. Leaderboard (không cần login)

- Client tạo `guest_id` (UUID ngẫu nhiên) lưu ở localStorage, gửi kèm khi vào trận.
- **Trong phòng** (realtime, 1 lần/giây): top 10 theo `total_exp` hiện tại + hạng của mình.
- **Toàn cục**: tab **Hôm nay / Tuần này / Mọi thời đại**, 2 bảng:
  - **Tu Vi Đỉnh Cao**: EXP cao nhất đạt được trong 1 mạng.
  - **Sát Thần**: số người chém chết nhiều nhất trong 1 mạng.
- Lưu mỗi "mạng" vào SQLite: `guest_id, name, character, peak_exp, peak_level, kills, survive_secs, map, ended_at`.
- API: `GET /api/leaderboard?period=day|week|all&metric=exp|kills` → top 50, cache 10s.
- Điểm số do server tính → không gian lận được từ client. Bot không được ghi.

## 12. Trạng thái nhân vật, animation & tấu hài

Server tính `mood` và gửi xuống (mọi người thấy giống nhau). Giữ mood tối thiểu 0.5s để không nhấp nháy.

| Anim | Khi nào | Tấu hài |
|---|---|---|
| `idle` | Đứng yên | Ngáp, gãi đầu |
| `run` | Chạy, không có địch gần | La toáng "AAAA" |
| `chase` | Đang tiến lại kẻ yếu hơn trong tầm | Cười nham hiểm, xoa tay |
| `flee` | Đang chạy khỏi kẻ mạnh hơn | Giơ tay hoảng loạn, mồ hôi văng |
| `hit` | Bị chém (0.6s) | Mặt bẹp, sao quay quanh đầu |
| `levelup` | Lên tầng | Tạo dáng, lóe sáng |
| `breakthrough` | Lên cảnh giới | Sét đánh, tóc dựng |
| `realm_drop` | Rớt cảnh giới | Ngồi bệt khóc |
| `teleport_out` / `teleport_in` | Dịch chuyển | Bị hút xoáy / văng ra chóng mặt |
| `death` | Chết | Dẹp lép như bánh tráng, hồn bay lên |
| `respawn` | Vào trận | Rơi từ trời xuống, cắm mặt, đứng dậy phủi bụi |

- Emote bubble khi: mất kiếm, khiên vỡ, nhặt Rương Báu, bị lộ khỏi bụi. Mỗi loại có cooldown.
- Client tự lật sprite trái/phải theo hướng chạy.

## 13. Kiến trúc kỹ thuật

### 13.1 Stack đã chọn

| Phần | Công nghệ | Lý do |
|---|---|---|
| Server | **Go** (bản ổn định mới nhất) | Hiệu năng cao, code dễ đọc, compile nhanh, ra 1 file chạy duy nhất, Claude viết Go rất ổn định |
| WebSocket | `github.com/coder/websocket` | Gọn, hiện đại |
| Mã hóa gói tin | MessagePack (`vmihailenco/msgpack/v5` ↔ `@msgpack/msgpack`) | Nhỏ hơn JSON nhiều, dễ debug hơn binary tự chế |
| Database | SQLite qua `modernc.org/sqlite` | Không cần cài DB server, không cần CGO |
| Client | **TypeScript + Vite + PixiJS** | Chạy ngay trên trình duyệt PC & điện thoại (đúng kiểu `.io`), PixiJS vẽ hàng nghìn sprite 2D mượt bằng WebGL |
| Âm thanh | `howler.js` | Xử lý được việc điện thoại chặn auto-play |
| Bản đồ | Tiled (`.tmj`) | Vẽ map bằng kéo thả, không cần code |

Nguyên tắc: **server quyết định mọi thứ** (authoritative). Client chỉ gửi hướng di chuyển và vẽ lại những gì server gửi.

### 13.2 Cấu trúc repo
```
kiem-xoay/
├── CLAUDE.md
├── docs/            GAME_SPEC.md · ART_PROMPTS.md · SETUP_GUIDE.md · PROGRESS.md
├── config/game.json                 # mọi con số cân bằng; server đọc, client lấy qua GET /api/config
├── maps/            lang_tre.tmj · cho_dem.tmj · nui_da.tmj
├── server/
│   ├── go.mod
│   ├── cmd/server/main.go           # HTTP + WebSocket + phục vụ file client đã build
│   ├── cmd/sim/main.go              # chạy giả lập không giao diện với N bot để cân bằng game
│   └── internal/
│       ├── game/      # luật chơi thuần: entity, tick, va chạm, exp, kiếm, buff (có unit test)
│       ├── world/     # bản đồ, vật cản, vùng, cổng, lưới tìm đường
│       ├── bot/       # AI
│       ├── room/      # phòng, matchmaking, bot fill
│       ├── netproto/  # định nghĩa message
│       └── store/     # SQLite, leaderboard
├── client/
│   ├── package.json · vite.config.ts · index.html
│   └── src/
│       ├── main.ts
│       ├── net/       # kết nối WebSocket, giải mã, nội suy
│       ├── render/    # PixiJS: nhân vật, kiếm, bản đồ, VFX
│       ├── ui/        # trang chủ, HUD, màn hình chết, leaderboard, joystick
│       └── audio/
└── assets/          # xem ART_PROMPTS.md
```

### 13.3 Mạng
- Tick server 30 Hz, mỗi phòng 1 goroutine. Snapshot gửi client 20 Hz.
- Client → Server: `Join{name, character, guest_id}`, `Input{seq, angle_u8, moving}` (30 Hz), `Ping`.
- Server → Client: `Welcome`, `Snapshot`, `Event`, `Leaderboard` (1 Hz).
- **Không gửi vị trí từng thanh kiếm**: gửi `phase, dir, n, ω, ring_radius, sword_length, phase_tick`; client tự tính góc. Đổi chiều → event `RingFlip`.
- Event: `Hit, SwordBreak, Death, LevelUp, Breakthrough, RealmDrop, RingFlip, ObstacleDamaged, ObstacleBroken, ObstacleRegrow, TeleportOut, TeleportIn, ItemPicked, BuffStart, BuffEnd, ChestIncoming, KillFeed`.
- AOI: chỉ gửi những gì nằm trong vùng nhìn của người đó (+ lề). Người đang ẩn thân / trong bụi **không được gửi** cho người không được thấy (chống hack nhìn xuyên).
- Client: nội suy người khác (trễ 100ms); tự dự đoán chuyển động của mình + hiệu chỉnh theo server.

### 13.4 Va chạm & hiệu năng
- Lưới không gian (spatial hash) cho thân, quái, item, vật cản.
- Kiếm–thân: kiểm tra thân B có nằm trong vành khăn của vòng A không, rồi chỉ xét những kiếm có góc gần hướng tới B.
- Kiếm–kiếm: chỉ xét cặp vòng chồng nhau; tìm kiếm gần 2 giao điểm của 2 vòng tròn, kiểm tra capsule–capsule.
- Vật cản tĩnh index sẵn theo lưới.
- Mục tiêu: 1 phòng đầy (40 nhân vật × 45 kiếm, 120 quái, ~300 vật cản) < 3ms/tick.

### 13.5 An toàn cơ bản
- Kiểm tra & giới hạn mọi input từ client; rate limit tin nhắn và số kết nối mỗi IP.
- Lọc tên tục, giới hạn độ dài.

## 14. Quái (PvE)

Quái **không gây sát thương**, đi lang thang, chạy khi bị chém, hiện số HP trên người.

| Loại | HP | EXP | Drop item | Weight |
|---|---|---|---|---|
| Nhỏ | 3 | 10 | 10% | 60 |
| Vừa | 8 | 30 | 25% | 30 |
| Lớn | 20 | 90 | 60% | 9 |
| Boss mini | 80 | 400 | 100% (3 item) | 1, tối đa 1/phòng |

Duy trì ~150 quái/phòng (bản đồ lớn hơn v1).

## 15. Milestones

Mỗi milestone là 1 phiên làm việc với Claude Code (prompt mẫu trong `SETUP_GUIDE.md`).

| # | Nội dung | Xong khi |
|---|---|---|
| M0 | Tạo repo, server Go "hello", client Vite + PixiJS kết nối WebSocket | Mở trình duyệt thấy chấm tròn di chuyển theo phím, vị trí do server gửi |
| M1 | Vòng kiếm, quái, item kiếm, EXP, 9 cảnh giới, scale + camera. Hình placeholder | Chơi 1 mình, lên Trúc Cơ thấy to ra, kiếm dài ra, camera lùi ra |
| M2 | PvP: chém thân, rớt tầng, bể kiếm theo cảnh giới, chết → màn hình Tổng kết → chơi tiếp | Mở 2 tab đánh nhau được |
| M3 | Phòng, matchmaking, bot fill, bot AI cơ bản, trang chủ nhập tên, leaderboard SQLite | 1 tab = 1 phòng có 19 bot; trang leaderboard có dữ liệu |
| M4 | Bản đồ Tiled: vật cản phá được + mọc lại, bụi, bùn, gió, hẻm, cổng; bot biết dùng địa hình | Chơi được trên 3 bản đồ |
| M5 | Toàn bộ vật phẩm buff + Rương Báu; netcode mượt (nội suy, dự đoán, AOI, ẩn thân) | Giả lập ping 150ms vẫn mượt |
| M6 | Ghép art, animation, mood, VFX, SFX, UI hoàn chỉnh | Đủ anim mục 12 |
| M7 | Tối ưu điện thoại, joystick, deploy lên VPS có HTTPS | Bạn bè vào được bằng link |

Sau mỗi milestone: chạy `go run ./cmd/sim` để xem cân bằng (thời gian trung bình lên mỗi cảnh giới, số kiếm trung bình, số lần chết).

## 16. Config mặc định (`config/game.json`)

```json
{
  "cultivation": {
    "realms": ["Luyện Khí","Trúc Cơ","Kim Đan","Nguyên Anh","Hóa Thần","Luyện Hư","Hợp Thể","Độ Kiếp","Đại Thừa"],
    "tiers_per_realm": 10, "swords_per_realm": 5,
    "exp_base": 100, "realm_mult": 1.2, "tier_step": 0.08,
    "scale_min": 1.0, "scale_max": 5.0
  },
  "sword": {
    "start_count": 1, "min_count": 1,
    "ring_min_r": 70, "spacing": 22, "length": 40, "width": 10,
    "omega": 1.8, "flip_cooldown": 0.3,
    "clash_ratio_base": 1,
    "clash_ratio_per_realm": 1
  },
  "player": {
    "body_r": 30, "move_base": 260, "pickup_r": 60,
    "invuln_on_hit": 1.0, "invuln_on_spawn": 2.0,
    "attacker_share": 0.5, "drop_share": 0.3, "kill_bonus_exp": 30,
    "death_sword_drop": 0.5, "orb_owner_secs": 1.5,
    "camera_view_h": 900, "camera_exp": 0.85
  },
  "buffs": {
    "spin":   { "mult": 1.6, "secs": 12 },
    "haste":  { "mult": 1.4, "secs": 10 },
    "magnet": { "mult": 3.0, "secs": 15 },
    "long":   { "mult": 1.3, "secs": 12 },
    "shield": { "secs": 8 },
    "exp2":   { "mult": 2.0, "secs": 20 },
    "stealth":{ "secs": 6, "reveal_r": 120 }
  },
  "items": {
    "weights": { "sword": 40, "spin": 15, "haste": 15, "magnet": 8, "long": 7, "shield": 7, "exp2": 5, "stealth": 3 },
    "count_per_room": 120, "overcap_exp": 20, "chest_every_secs": 120, "chest_items": 5
  },
  "obstacles": {
    "regrow_secs": 30, "regrow_warn_secs": 3,
    "hp": { "fence": 30, "stall": 60, "brick": 150, "boulder": 400 },
    "stall_item_drop": 0.3
  },
  "zones": { "mud_mult": 0.6, "wind_with": 1.5, "wind_against": 0.8, "wind_with_max_deg": 60, "wind_against_min_deg": 120, "bush_reveal_r": 150, "bush_reveal_on_hit_secs": 2 },
  "gates": { "radius": 150, "transit_secs": 0.8, "exit_warn_secs": 0.5, "cooldown_secs": 8 },
  "room": {
    "max_humans": 20, "bot_fill_target": 20, "map_size": 12000,
    "monster_count": 150, "empty_close_secs": 60, "safe_spawn_r": 800
  },
  "net": { "tick_hz": 30, "snapshot_hz": 20, "interp_ms": 100 }
}
```

Tỉ lệ đổi kiếm `k = clash_ratio_base + clash_ratio_per_realm * chênh_lệch_cảnh_giới`. Mặc định 1 + 1×d (Trúc Cơ 1:2, Nguyên Anh 1:4). Muốn cảnh giới cao áp đảo hơn thì tăng `clash_ratio_per_realm` (ví dụ 2 → Nguyên Anh đổi 1:7).

## 17. Checklist test chấp nhận

- [ ] Luyện Khí tầng 1 nhỏ hơn rõ ràng so với Luyện Khí tầng 10; Đại Thừa tầng 10 scale = 5.
- [ ] Kiếm dài ra và vòng rộng ra theo scale.
- [ ] Luyện Khí không có quá 5 kiếm; lượm thêm → +20 EXP.
- [ ] 2 người cùng cảnh giới va kiếm → mỗi bên mất 1 kiếm, cả hai đổi chiều quay.
- [ ] Trúc Cơ va Luyện Khí → Luyện Khí mất 2, Trúc Cơ mất 1.
- [ ] Luyện Khí (≥5 kiếm) va Nguyên Anh → Luyện Khí mất 4, Nguyên Anh mất 1.
- [ ] Luyện Khí còn 3 kiếm va Nguyên Anh → Luyện Khí còn 1, Nguyên Anh chưa mất (nứt 2/4).
- [ ] Không ai bao giờ có dưới 1 kiếm.
- [ ] Bị chém ở Trúc Cơ tầng 1 → Luyện Khí tầng 10, kiếm thừa rơi ra đất.
- [ ] Bị chém ở Luyện Khí tầng 1 → màn hình Tổng kết, bấm "Chơi tiếp" vào lại được; người chém nhận 50% EXP nạn nhân + 30; xác rơi 50% số kiếm.
- [ ] Viên EXP từ quái chỉ người chém nhặt được trong 1.5s đầu.
- [ ] Đi chéo 90° so với gió → tốc độ không đổi.
- [ ] Vật cản bị chém vỡ → đi qua được → 30s sau mọc lại, không kẹt người.
- [ ] Kiếm chém vật cản không bị mất.
- [ ] Vào cổng → ra ở cổng khác, cổng đích lóe sáng trước, không có bất tử khi ra.
- [ ] Người scale lớn không chui lọt hẻm 150.
- [ ] Buff hết hạn đúng giờ, nhặt lại thì làm mới thời gian.
- [ ] Người ẩn thân không xuất hiện trong gói tin gửi cho người ở xa.
- [ ] Phòng 5 người → 15 bot; thêm người không xóa bot; người thứ 21 sang phòng mới.
- [ ] Leaderboard ghi đúng sau khi chết, bot không có trong bảng toàn cục.
