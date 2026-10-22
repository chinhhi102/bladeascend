# KIẾM XOAY TU TIÊN — Game Design & Technical Spec (v2)

> Tên tạm (working title). File này dành cho Claude Code: đặt tại `docs/GAME_SPEC.md`.
> Mọi con số nằm trong `config/game.json` (mục 16) — **không hardcode**.
> Art/animation prompts: `docs/ART_PROMPTS.md`. Hướng dẫn setup: `docs/SETUP_GUIDE.md`.

**Thay đổi v3.7:** thêm **Thiêu Đốt Tinh Huyết** — giữ phím để chạy nhanh ×1.8, đổi bằng EXP mỗi giây, không bao giờ làm rớt cảnh giới (§4.1c).

**Thay đổi v3.6:** vật cản cứng hơn — damage tính theo người (0.5s/lần) thay vì theo từng kiếm, HP tăng (§5.3, §8.2).

**Thay đổi v3.5:** clash_unit = 1; chỉ đổi chiều quay khi cùng cảnh giới; số kiếm có thể về 0 (không chết); nhịp Luyện Khí → Trúc Cơ 30–60s (§5.2, §6.1).

**Thay đổi v3.4:** thêm minimap (§12.1) và danh hiệu Top 1–3 phòng (§11.3); kiếm quay chậm lại còn 1 vòng/giây, Tốc Kiếm trả về ×1.6 (§4.1b).

**Thay đổi v3.3:** bảng xếp hạng dùng điểm Uy Danh thay EXP, thêm bảng Sát Thần (lần chém trúng) và Thần Tốc (lên Đại Thừa nhanh nhất); Ẩn Thân giảm còn 5s, tàng hình 80% (§6.2, §11).

**Thay đổi v3.2:** bán kính vòng kiếm cố định, không phụ thuộc số kiếm; chỉ vòng mới mới rộng hơn (§4.1a).

**Thay đổi v3.1:** vòng kiếm vẽ kiểu cánh quạt như quảng cáo — so le góc, lưỡi nghiêng 35°, mỗi vòng một màu, vòng ngoài lùi ra xa hơn (§4.1a).

**Thay đổi v3:** kiếm dài gấp đôi, quay 300ms/vòng, kiếm tối đa 60→400 xếp vòng lồng nhau (20 kiếm/vòng), bị chém rớt 1 đại cảnh giới, EXP rải khắp bản đồ, vật phẩm ×3, vật cản chỉ chặn ở chân đế + y-sort.

**Thay đổi v2 so với v1:** đổi stack (Go server + TypeScript/PixiJS client trên web) · kiếm có thể bị bể khi va nhau (tỉ lệ theo chênh lệch cảnh giới) · scale 1 → 5 và tầm kiếm dài theo · chết ra màn hình chờ rồi match lại · bản đồ có địa hình, vật cản phá được (mọc lại sau 30s), cổng dịch chuyển · không login, nhập tên là chơi, có leaderboard · bộ vật phẩm buff có thời hạn.

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
scale(L) = 1 + 4 * L / 89          // 1.00 (Luyện Khí 1) → 5.00 (Đại Thừa 10), to ra ở MỖI tầng
```

| r | Cảnh giới | Kiếm tối đa | Số vòng kiếm tối đa | Scale tầng 1 → tầng 10 |
|---|---|---|---|---|
| 0 | Luyện Khí | 60 | 3 | 1.00 → 1.40 |
| 1 | Trúc Cơ | 120 | 6 | 1.45 → 1.85 |
| 2 | Kim Đan | 160 | 8 | 1.90 → 2.30 |
| 3 | Nguyên Anh | 200 | 10 | 2.35 → 2.75 |
| 4 | Hóa Thần | 240 | 12 | 2.80 → 3.20 |
| 5 | Luyện Hư | 280 | 14 | 3.25 → 3.65 |
| 6 | Hợp Thể | 320 | 16 | 3.70 → 4.10 |
| 7 | Độ Kiếp | 360 | 18 | 4.15 → 4.55 |
| 8 | Đại Thừa | 400 | 20 | 4.60 → 5.00 |

Kiếm tối đa lấy từ mảng `max_swords` trong config (từ Nguyên Anh trở đi mặc định +40 mỗi cảnh giới).

### 3.2 EXP

```
exp_needed(r, s) = round(100 * 2.0^r * (1 + 0.08 * s))
```

**Mỗi cảnh giới tốn gấp đôi cảnh giới trước** (`realm_mult` = 2.0): Luyện Khí 1 360 → Trúc Cơ 2 720 → Kim Đan 5 440…

| Cảnh giới | EXP/tầng (1 → 10) | Tổng cả cảnh giới |
|---|---|---|
| Luyện Khí | 100 → 172 | 1 360 |
| Trúc Cơ | 200 → 344 | 2 720 |
| Kim Đan | 400 → 688 | 5 440 |
| Nguyên Anh | 800 → 1 376 | 10 880 |
| Hóa Thần | 1 600 → 2 752 | 21 760 |
| Luyện Hư | 3 200 → 5 504 | 43 520 |
| Hợp Thể | 6 400 → 11 008 | 87 040 |
| Độ Kiếp | 12 800 → 22 016 | 174 080 |
| Đại Thừa | 25 600 → 44 032 | 348 160 |

Tổng từ Luyện Khí 1 tới Đại Thừa 10: **694 960 EXP** (trước đây 28 288).

- Lưu `total_exp` cộng dồn, level suy ra từ bảng ngưỡng tích lũy tính sẵn.
- Đạt Đại Thừa tầng 10: `total_exp` dừng ở trần, EXP nhặt thêm bị bỏ (xếp hạng dùng Uy Danh §11, không dùng EXP).
- Lên tầng: glow nhỏ. **Đột phá cảnh giới**: anim `breakthrough`, sét, camera rung, đổi skin kiếm theo cảnh giới, nhân vật quanh đó hiện emote kinh ngạc.

## 4. Kiếm, vòng quay, tầm đánh

### 4.1a Vòng kiếm nhiều lớp (kiểu cánh quạt như quảng cáo)

Mục tiêu hình ảnh: nhiều vòng kiếm **so le, nghiêng, mỗi vòng một màu**, nhìn thành một dải dày xoáy như chong chóng. Tất cả các lớp đều lộ ra, không nấp sau nhau.

**Quy tắc cố định bán kính (BẮT BUỘC)**
- Bán kính và chiều dài của một vòng **chỉ phụ thuộc vào thứ tự vòng `k` và `scale`**, **KHÔNG phụ thuộc vào số kiếm `n`**. Thêm kiếm vào vòng đang có → chỉ lấp kín khe, vòng **không rộng ra**.
- Tầm chém chỉ tăng khi (a) mở **vòng mới** (kiếm thứ 21, 41, 61...), (b) nhân vật **lên tầng/cảnh giới** (scale tăng), hoặc (c) có buff Trường Kiếm.
- Xóa mọi công thức cũ kiểu `ring_radius = max(..., n * spacing / 2π)` (từ spec v1) khỏi code.
- Chuôi kiếm vòng 1 luôn sát người (`inner_r = (BODY_R + 6) * scale`), không có khoảng trống lớn giữa thân và kiếm.

Ví dụ ở cùng 1 tầng (scale không đổi): 1 kiếm, 10 kiếm, 20 kiếm → cùng 1 vòng, cùng bán kính. 21 kiếm → xuất hiện vòng 2 xa hơn. 40 kiếm → vẫn 2 vòng, bán kính không đổi.

**Chia vòng**
- Cứ **20 kiếm thành 1 vòng**. Vòng 1 (k = 0) là 20 kiếm đầu, vòng 2 là 20 kiếm tiếp...
- Khi n ≤ 20: n kiếm chia đều 360°. Khi n > 20: mỗi vòng có 20 "khe" cách nhau 18°; vòng chưa đủ 20 kiếm thì chiếm các khe rải đều (`slot = round(j * 20 / m)`).

**Hình học từng vòng** (`R` = tổng số vòng hiện có, `base = SWORD_LEN * scale * buff_long`, SWORD_LEN = 80)

```
inner_r           = (BODY_R + 6) * scale                     // bán kính chuôi vòng 1
hilt_r(k)         = inner_r + HILT_STEP * k * base           // HILT_STEP = 0.1  → vòng ngoài lùi ra xa hơn
len(k)            = base * (1 + (RING_GROWTH - HILT_STEP) * k) // RING_GROWTH = 0.2 → mũi vòng sau xa hơn vòng trước 20% chiều dài gốc
stagger(k)        = k * (18° / R)                            // so le: các vòng chia đều khoảng trống giữa 2 khe
tilt              = TILT_DEG * dir                           // TILT_DEG = 35°, lưỡi ngả về phía sau so với chiều quay
angle(i, k, t)    = phase + dir * ω * t + slot_angle(i) + stagger(k)
ω                 = 2π / rev_secs * buff_spin                // rev_secs = 1.0 → 1 vòng/giây; buff_spin = 1.6 khi có Tốc Kiếm
```

- Mỗi thanh kiếm đặt chuôi tại điểm `(hilt_r(k), angle)` và chĩa ra theo hướng **bán kính xoay thêm `tilt`**, dài `len(k)`.
- Đổi chiều quay → `tilt` lật theo, tween 0.1s (lưỡi luôn ngả về phía sau, như kéo lê theo vòng xoáy).
- Tầm chém thực tế: `reach = max_k |tip(k)|` (khoảng cách từ tâm tới mũi kiếm xa nhất, tính hình học vì kiếm nghiêng). Tham khảo tầng 10, full kiếm: Luyện Khí ≈ 200 · Trúc Cơ ≈ 340 · Kim Đan ≈ 500 · Đại Thừa ≈ 2 000 đơn vị.
- Config `ring_growth_mode: "linear" | "compound"` vẫn giữ để thử.

**Màu từng vòng**
- Vòng k tô màu `ring_tints[k % số_màu]` (mặc định cam → tím → xanh lá → đỏ → xanh lơ → vàng). Skin kiếm theo cảnh giới vẫn giữ hình dáng, chỉ nhuộm màu theo vòng.
- Vòng 1 dùng màu gốc của skin (tint trắng) để người chơi nhận ra cảnh giới.

**Thứ tự vẽ**: vòng ngoài cùng vẽ trước (dưới), vòng 1 vẽ sau cùng (trên). Mỗi vòng có lớp spin blur cùng màu (mục 4.1b).

**Hitbox (server dùng đúng hình học trên)**
- Mỗi kiếm: capsule từ chuôi tới mũi theo hướng đã nghiêng, độ dày `sword_width = 12 * scale`.
- Thân: hình tròn `BODY_R * scale` (BODY_R = 30).
- Số kiếm tối thiểu **0** (có thể bị tước sạch kiếm). Bắt đầu: 1 kiếm. Thêm/bớt kiếm hoặc đổi số vòng (stagger thay đổi): tween vị trí 0.2s.

### 4.1b Va chạm và hiển thị khi kiếm quay
- **Tốc độ gốc: 1 vòng / 1 giây** (`rev_secs: 1.0`). Có buff Tốc Kiếm: 1 vòng / 0.625 giây.
- **Va chạm theo góc quét**: ở 30 tick/s, mỗi tick kiếm quay ~12° (~19° khi có buff). Vẫn kiểm tra **cung đã quét** từ tick trước tới tick này (sweep), không chỉ vị trí tức thời, để kiếm không "xuyên" qua người. Mục tiêu nằm trong vành khăn `[inner_r, reach]` và trong cung quét của bất kỳ kiếm nào → trúng.
- **Hiển thị**: ở 60 fps mỗi frame quay ~6°, nhỏ hơn nhiều so với khoảng cách 18° giữa 2 kiếm → nhìn rõ từng thanh kiếm quay, không còn hiệu ứng quay ngược. Vệt mờ (spin blur) chỉ cần nhẹ, mảnh, cùng màu vòng; đậm hơn khi có buff Tốc Kiếm.
- Dùng `ParticleContainer` của PixiJS cho kiếm (có thể tới hàng nghìn sprite trên màn hình).

### 4.1 Tốc độ di chuyển & camera

```
move_speed (đơn vị thế giới/s) = 260 * sqrt(scale) * buff_haste(1.4 | 1.0) * zone_mult
camera: chiều cao vùng nhìn thấy (đơn vị thế giới) = max(900 * scale^0.85, 2.6 * reach)
```

→ Camera luôn lùi đủ xa để thấy trọn vòng kiếm của mình. Trên màn hình, nhân vật to lên dần nhưng vẫn nhìn được rộng hơn (bản đồ như "thu nhỏ" lại). Người nhỏ di chuyển trên màn hình nhanh hơn người lớn (≈260 px/s so với ≈150 px/s ở scale 5) → người nhỏ có cơ hội né và lật kèo.

### 4.1c Thiêu Đốt Tinh Huyết (chạy nhanh, đổi bằng tu vi)

Không có vật phẩm nào cho chạy nhanh theo ý muốn, nên người chơi có một cách **chủ động**: đốt chính tu vi của mình.

```
Giữ Shift (PC) / nút 🔥 (điện thoại) + đang di chuyển
  → tốc độ × burn.mult (1.8)
  → mỗi giây mất burn.exp_frac_per_sec (8%) × exp_needed(tầng hiện tại) EXP
```

- **Chỉ tốn khi thật sự chạy.** Đứng yên giữ phím thì không mất gì.
- **Cộng dồn với Thần Hành**: 130 × 1.4 × 1.8 ≈ 327 đơn vị/s — đủ để thoát khỏi một vòng kiếm đang đuổi.
- **Không bao giờ rớt cảnh giới vì đốt** (`burn.keep_realm` = true): đốt chỉ tụt được tới **đáy cảnh giới đang ở**, chạm đáy thì tự tắt và tốc độ về bình thường. Rớt cảnh giới vẫn chỉ đến từ việc bị chém (§5.1). Ở Luyện Khí thì đáy là 0 EXP — đốt cạn cũng **không chết**.
- Giá tiền tỉ lệ với tầng hiện tại nên công bằng ở mọi cảnh giới: Luyện Khí 1 mất 8 EXP/s, Đại Thừa 10 mất ≈3 500 EXP/s.
- **Hiển thị**: vòng lửa đỏ + lưỡi lửa liếm quanh chân (cả phòng đều thấy, cờ `FlagBurning`), nên đốt là tự báo cho kẻ săn mình biết. Màn Tổng kết hiện tổng "Tu vi đã thiêu đốt".
- **Bot** chỉ đốt khi đang ở trạng thái Flee — đúng lúc đáng đánh đổi.

## 5. Chiến đấu

Thứ tự xử lý mỗi tick server:

### 5.1 Kiếm A chạm thân B — rớt 1 đại cảnh giới
(A ≠ B, B không bất tử, không có khiên, không đang dịch chuyển)
- Nếu B đang ở **Luyện Khí** (bất kỳ tầng nào) → B **chết** (mục 7). `loss = B.total_exp`; A nhận 50% loss + `kill_bonus_exp` (30), rơi 30% loss thành viên EXP, +1 kill.
- Ngược lại B **rớt nguyên 1 cảnh giới, giữ nguyên tầng**: `L' = L - 10` (ví dụ Trúc Cơ tầng 2 → Luyện Khí tầng 2). Phần trăm tiến độ trong tầng được giữ: `total_exp' = cum(L') + progress% * exp_needed(L')`. `loss = total_exp - total_exp'`.
- A nhận `50% * loss` EXP; rơi `30% * loss` dạng viên EXP quanh B.
- B bất tử **1s** (nhấp nháy), bị đẩy lùi khỏi A.
- Mỗi tick, mỗi nạn nhân tối đa 1 hit.
- Số kiếm > max của cảnh giới mới → kiếm thừa rơi xuống đất dưới dạng **bó kiếm** (mỗi bó tối đa 10 kiếm) để không tạo hàng trăm vật phẩm.
- Nếu B có buff Khiên: khiên vỡ, không mất gì, vẫn có 1s bất tử.

### 5.2 Kiếm A chạm kiếm B — bể kiếm

Luôn: VFX tia lửa + SFX "keng". Cooldown mỗi cặp (A, B) 0.3s.

**Không đẩy văng nữa** (`clash_knockback` = 0): hai người **đứng nguyên mà mài nhau**. Vòng kiếm chỉ có tác dụng chặn — còn kiếm thì đối thủ không chém được vào thân — nên trận đánh là cuộc đua xem **ai hết kiếm trước**. Bên hết kiếm trước lãnh trọn §5.1. Đặt `clash_knockback` > 0 để bật lại kiểu đẩy văng cũ.

**Đổi chiều quay chỉ khi cùng cảnh giới**: `rA == rB` → cả hai vòng đổi chiều (`dir *= -1`). Khác cảnh giới → **không ai đổi chiều**, cả hai giữ nguyên chiều quay.

Sau đó tính bể kiếm, với `d = |rA - rB|`, `k = d + 1` và đơn vị mất kiếm `u = clash_unit` (mặc định **1**):

| Trường hợp | Kết quả |
|---|---|
| Cùng cảnh giới (d = 0) | Mỗi bên mất `u` kiếm, cả hai đổi chiều |
| Khác cảnh giới | Bên thấp mất `u * k` kiếm; bên cao nhận `crack += lost / k` (lost = số kiếm bên thấp thực sự mất). Khi `crack ≥ 1` → bên cao mất `floor(crack)` kiếm, `crack -= floor(crack)`. Không ai đổi chiều |

Số kiếm có thể về **0** (`min_count = 0`). Lý do: vòng kiếm dày chặn hết kiếm đối thủ trước khi chạm tới thân, nên phải tước được sạch kiếm thì mới chém được vào người.
- 0 kiếm **không chết**: chỉ bị tước vũ khí. Chém trúng thân vẫn theo §5.1 (Trúc Cơ trở lên rớt 1 cảnh giới, Luyện Khí chết).
- Người 0 kiếm hiện emote hoảng loạn, nhặt item Kiếm/Bó Kiếm để có lại vòng. Bot 0 kiếm luôn chuyển sang Flee.
- Bên 0 kiếm không còn kiếm để va, nên không còn tính bể kiếm với họ.

Bảng tỉ lệ đổi kiếm (bên cao : bên thấp), u = 1:

| Chênh lệch | Ví dụ | Mỗi lần va | Đổi chiều |
|---|---|---|---|
| 0 | Kim Đan vs Kim Đan | mỗi bên −1 | Có |
| 1 | Trúc Cơ vs Luyện Khí | Luyện Khí −2; Trúc Cơ −1 | Không |
| 3 | Nguyên Anh vs Luyện Khí | Luyện Khí −4; Nguyên Anh −1 | Không |
| 8 | Đại Thừa vs Luyện Khí | Luyện Khí −9; Đại Thừa −1 | Không |

`crack` lưu trên người, reset về 0 khi chết. Hiển thị vết nứt nhẹ trên vòng kiếm khi `crack > 0`.

Kiếm bị bể: VFX vỡ vụn + SFX, mảnh văng ra (không thành item). Chủ nhân hiện emote "Kiếm của taaa!".

### 5.3 Kiếm chạm quái / vật cản
- Quái: 1 damage mỗi lần chạm, cooldown mỗi cặp (kiếm, quái) 0.25s. Quái chết → **toàn bộ EXP rơi thành viên EXP** tại chỗ. Trong `orb_owner_secs` (1.5s) đầu, chỉ người chém hạ quái được nhặt; sau đó ai cũng nhặt được.
- Vật cản: tính damage **theo người, không theo từng thanh kiếm** (trước đây mỗi thanh kiếm gây damage riêng → 60–400 kiếm xay nát vật cản trong chớp mắt). Mỗi người chỉ gây damage lên 1 vật cản tối đa 1 lần / 0.5s, khi có ít nhất 1 kiếm của họ chạm vật cản:
  `damage = (r + 1) × (1 + 0.5 × (số_vòng − 1))` → mỗi vòng thêm +50% sức phá. Nhiều người cùng chém thì cộng dồn.
  **Kiếm xuyên qua vật cản và không bị mất/bể** khi chém vật cản.
- Kiếm không tương tác với chính chủ.

## 6. EXP rải khắp bản đồ & vật phẩm

### 6.1 Viên EXP môi trường (nguồn EXP chính)
- Bản đồ luôn phủ đều viên EXP: mật độ mặc định **1 viên / 226×226 đơn vị** (≈ 2 800 viên trên bản đồ 12 000²). `cell` = 226 ≈ 160 × √2, tức **một nửa mật độ** so với bản trước (số viên tỉ lệ nghịch với `cell²`).
- Tỉ lệ: nhỏ +5 (80%) · vừa +20 (18%) · lớn +60 (2%).
- Viên bị nhặt → 5s sau mọc lại ở vị trí ngẫu nhiên **trong cùng ô lưới 1 000×1 000** (khu vực bị vét sạch sẽ tự đầy lại, không dồn về một chỗ).
- Bán kính hút đồ mặc định: `90 * scale`.
- Nhịp độ mục tiêu (người chơi chỉ farm, không đánh nhau): **Luyện Khí → Trúc Cơ ≈ 30–60s** (hiện đo được 40–46s với `cell = 160`), tới Kim Đan ≈ 1 phút, tới Nguyên Anh ≈ 1.2 phút. Dùng `cmd/sim` đo và chỉnh mật độ cho tới khi đạt.
- Mạng: viên EXP đứng yên → chỉ gửi theo AOI khi vào tầm nhìn, sau đó chỉ gửi sự kiện nhặt/mọc.

### 6.2 Vật phẩm

| Item | Hiệu ứng | Thời hạn | Weight spawn |
|---|---|---|---|
| Kiếm | +3 kiếm. Nếu đã max → +20 EXP | – | 40 |
| Bó Kiếm | +15 kiếm (cũng là dạng rơi ra khi rớt cảnh giới/chết, khi đó chứa đúng số kiếm rơi) | – | 6 |
| Tốc Kiếm | ω × 1.6 | 12s | 12 |
| Thần Hành | Tốc chạy × 1.4 | 10s | 12 |
| Nam Châm | Bán kính hút đồ × 3 | 15s | 8 |
| Trường Kiếm | Chiều dài mọi vòng kiếm × 1.3 | 12s | 7 |
| Kim Cang (khiên) | Chặn 1 lần bị chém vào thân | 8s hoặc tới khi vỡ | 7 |
| Song Bội Đan | EXP nhận × 2 | 20s | 5 |
| Ẩn Thân Phù | Tàng hình **80%**: người khác thấy mình và vòng kiếm mờ (độ đậm 20%), không thấy tên, không hiện trên minimap. Ở rất gần (120 × scale của người nhìn) thì hiện rõ 100%. Hết hiệu lực ngay khi kiếm mình chém trúng người | **5s** (tối đa, nhặt lại chỉ làm mới về 5s) | 3 |
| **Rương Báu** (hiếm) | Người đầu tiên **chạm vào** → nổ ra 8 item ngẫu nhiên văng xung quanh | – | 1 rương mỗi 60s, báo trước toàn phòng 10s |

- Số vật phẩm duy trì: **200/phòng** (giảm 50% so với 400), item bị nhặt → 3s sau mọc lại chỗ khác.
- Nhặt lại buff đang có → làm mới thời gian (không cộng dồn).
- **Hiệu ứng nhặt**: vật phẩm và viên EXP **bay hút về phía người nhặt** (0.22s, nhỏ dần và mờ dần) rồi mới biến mất, không biến mất đột ngột. Chỉ là hiệu ứng phía client; server vẫn xoá item ngay khi nhặt.
- HUD hiển thị icon buff + vòng đếm ngược. Mỗi buff có aura màu riêng (trừ Ẩn Thân).
- Item spawn cách người chơi ≥ 300 đơn vị.

## 7. Chết & màn hình chờ

- **Rơi kiếm khi chết** (người và bot): `floor(n * death_sword_drop)` kiếm (mặc định 50%) văng ra quanh xác dưới dạng Bó Kiếm (mỗi bó ≤ 10). Ví dụ chết với 60 kiếm → 3 bó, tổng 30 kiếm.
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
| Sạp hàng / xe đẩy | 80 | Vỡ ra: 30% rơi 1 item |
| Tường gạch | 250 | |
| Đá tảng | 800 | Rất lớn |
| Tường đá cổ / biên map | ∞ | Không phá được |

Thời gian 1 người (đủ kiếm tối đa của cảnh giới) phá vỡ, theo công thức §5.3:

| Cảnh giới (số vòng) | Sức phá/giây | Rào tre | Sạp hàng | Tường gạch | Đá tảng |
|---|---|---|---|---|---|
| Luyện Khí (3) | 4 | 7.5s | 20s | 63s | 200s |
| Trúc Cơ (6) | 14 | 2s | 6s | 18s | 57s |
| Kim Đan (8) | 27 | 1s | 3s | 9s | 30s |
| Nguyên Anh (10) | 44 | 0.7s | 1.8s | 5.7s | 18s |
| Hóa Thần (12) | 65 | 0.5s | 1.2s | 3.8s | 12s |
| Đại Thừa (20) | 189 | tức thì | 0.4s | 1.3s | 4.2s |

→ Người nhỏ gần như không phá nổi tường gạch, đá tảng → dùng làm chỗ nấp thật sự. Người khổng lồ cảnh giới cao thì càn quét được — đúng cảm giác "đại năng".

- **Hai vùng va chạm khác nhau cho mỗi vật cản:**
  - **Chân đế (footprint)** — chặn thân nhân vật và quái. Chỉ là phần sát mặt đất: mặc định hình elip/chữ nhật rộng 90% sprite, cao **30% sprite, nằm ở đáy**. Trong Tiled đặt được thủ công bằng thuộc tính `footprint` (nếu không có thì tính tự động như trên).
  - **Thân hình (visual bounds)** — vùng kiếm chém trúng để gây damage (đúng với cái người chơi nhìn thấy).
- **Sắp xếp lớp vẽ theo chân (y-sort)**: nhân vật đứng phía sau (y chân nhỏ hơn đáy vật cản) được vẽ khuất sau vật cản; phần vật cản che lên nhân vật mình thì mờ đi 50% để vẫn thấy mình.
- **Vật cản dạng dãy** (hàng rào, tường): art phải liền mạch, nối tiếp nhau không hở (tile ngang). Chỗ nào art có khe hở thì footprint cũng phải hở theo — người chơi thấy đi qua được thì phải đi qua được.
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
- **Cổng luôn đi 2 chiều (A ↔ B)**: mỗi cổng có đúng một cổng đối diện cố định, ra ở cổng nào thì bước vào chính cổng đó là quay về chỗ cũ. Mạng nhiều cổng được ghép cặp theo thứ tự id (0↔1, 2↔3…), **không bắn ngẫu nhiên**.
- Kích hoạt khi tâm nhân vật đi vào bán kính cổng (150). Anim `teleport_out` 1s → xuất hiện ở cổng đích với anim `teleport_in` 1s.
- Cổng đích **lóe sáng 0.5s trước** khi người xuất hiện (cảnh báo nhẹ cho người đứng gần — vẫn dễ bị bất ngờ).
- **Không bất tử khi vừa ra cổng.** Trong lúc đang dịch chuyển (**2s**, `transit_secs`) không bị chém và không chém được.
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

- `power = sword_count * (r + 1) * reach`.
- Tham số tính cách random mỗi bot: `reaction_delay` 0.2–0.6s, `aim_noise`, `aggression`.
- Bot chỉ "thấy" những gì người chơi cùng vị trí thấy được (tôn trọng bụi rậm). Người đang Ẩn Thân: bot chỉ phát hiện trong 50% tầm nhìn bình thường.

## 11. Leaderboard (không cần login)

EXP **không** dùng để xếp hạng toàn cục vì có trần (Đại Thừa tầng 10) → ai chơi lâu cũng bằng nhau. Thay bằng điểm **Uy Danh** thưởng cho việc đánh nhau, lật kèo và giữ ngôi.

### 11.1 Điểm Uy Danh (tính trong 1 mạng, server tính)

| Hành động | Điểm |
|---|---|
| **Chém trúng người** (làm họ rớt cảnh giới hoặc chết) | `100 × (cảnh giới nạn nhân + 1) × lật_kèo` |
| Thưởng thêm khi chém chết | +50 |
| **Lật kèo**: nạn nhân cao cảnh giới hơn mình | `lật_kèo = 1 + 0.5 × (r_nạn_nhân − r_mình)`, tối thiểu 1 |
| **Đột phá** cảnh giới mới (lần đầu trong mạng đó) | `20 × r_mới` |
| **Giữ Top 1 phòng** | +1 điểm mỗi giây |

- Ví dụ: Trúc Cơ (r=1) chém trúng Nguyên Anh (r=3) → `100 × 4 × (1 + 0.5×2) = 800`. Đại Thừa chém Luyện Khí → `100 × 1 × 1 + 50 = 150`.
- Nạn nhân là **bot** → nhận 50% điểm.
- **Chống cày điểm**: chém trúng cùng 1 người (cùng `guest_id`) lần nữa trong vòng 60s → 0 điểm.
- HUD hiển thị Uy Danh hiện tại; mỗi lần cộng điểm hiện số bay lên (+800 LẬT KÈO!).
- Màn hình Tổng kết hiển thị bảng chi tiết điểm theo từng nguồn.

### 11.2 Các bảng xếp hạng

- **Trong phòng** (realtime, 1 lần/giây): top 10 theo **cảnh giới + tầng hiện tại**, hòa thì so `total_exp` (để biết ai đang mạnh mà né/săn), kèm Uy Danh. Top 1–3 có danh hiệu (§11.3).
- **Toàn cục**, tab **Hôm nay / Tuần này / Mọi thời đại**, 3 bảng:
  - **Uy Danh** (bảng chính): điểm Uy Danh cao nhất trong 1 mạng.
  - **Sát Thần**: số lần chém trúng người nhiều nhất trong 1 mạng.
  - **Thần Tốc**: thời gian ngắn nhất từ lúc vào trận tới Đại Thừa (chỉ ghi những mạng đã đạt Đại Thừa; xếp tăng dần).
- Lưu mỗi "mạng" vào SQLite: `guest_id, name, character, score, hits, kills, peak_level, secs_to_dai_thua (null nếu chưa đạt), top1_secs, survive_secs, map, ended_at`.
- API: `GET /api/leaderboard?period=day|week|all&board=score|hits|speed` → top 50, cache 10s.
- Điểm do server tính → client không gian lận được. Bot không được ghi.

### 11.3 Danh hiệu Top 3 phòng

| Hạng | Danh hiệu | Màu | Hiển thị trên đầu |
|---|---|---|---|
| Top 1 | **Bần tăng chưa ngán ai bao giờ** | Vàng kim, chữ có viền sáng | Vương miện + danh hiệu |
| Top 2 | **Ngon thì nhào vô** | Bạc | Danh hiệu |
| Top 3 | **Chỗ này nhiều máu nè** | Đồng | Danh hiệu |

- Danh hiệu hiện **phía trên tên** nhân vật trong trận (ai cũng thấy), và cạnh tên trong bảng xếp hạng phòng.
- Cập nhật theo bảng phòng (1 lần/giây), có **chống nhấp nháy**: phải giữ hạng mới liên tục 3s mới đổi danh hiệu.
- Ai vừa giành Top 1 → kill feed: `"{A} đã lên ngôi 'Bần tăng chưa ngán ai bao giờ'!"` + SFX trống. Ai vừa mất Top 1 → `"{B} bị hạ bệ, bần tăng giờ đã ngán"`.
- Người Ẩn Thân / trong bụi: danh hiệu ẩn theo (không làm lộ vị trí).
- Danh hiệu nằm trong config (`titles`), sửa chữ không cần đụng code.

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

### 12.1 Minimap

- Góc **trên bên trái** (dưới thanh cảnh giới), hình vuông bo góc, nền mờ 60%. Kích thước: 180px trên PC, 120px trên điện thoại. Chạm/nhấn phím `M` → phóng to giữa màn hình (bấm lại để thu nhỏ).
- **Hiển thị:**
  - Địa hình tĩnh vẽ sẵn 1 lần từ bản đồ Tiled: tường không phá được, đá tảng, bụi rậm, bùn, đường gió (màu nhạt, đơn giản).
  - Cổng dịch chuyển: chấm tròn theo màu cặp/mạng cổng.
  - **Mình**: mũi tên chỉ hướng di chuyển, màu nổi bật.
  - Khung chữ nhật mảnh thể hiện vùng camera đang nhìn.
  - **Top 1 phòng**: icon vương miện (cả phòng đều thấy → khuyến khích đi săn).
  - **Rương Báu**: icon rương nhấp nháy khi đang có, kể cả lúc báo trước 10s.
- **Không hiển thị** người chơi/bot khác (trừ Top 1), quái, item, viên EXP → giữ yếu tố bất ngờ và phục kích. Top 1 đang Ẩn Thân hoặc trong bụi thì ẩn khỏi minimap.
- Vật cản phá được không cần vẽ (đổi trạng thái liên tục).
- Dữ liệu: vị trí mình lấy ở client; vị trí Top 1 và Rương Báu gửi kèm gói `Leaderboard` (1 lần/giây).

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
- **Không gửi vị trí từng thanh kiếm**: gửi `phase, dir, n, ω, scale, buff_long, phase_tick` (client tự tính số vòng, chiều dài từng vòng); client tự tính góc. Đổi chiều → event `RingFlip`.
- Event: `Hit, SwordBreak, Death, LevelUp, Breakthrough, RealmDrop, RingFlip, ObstacleDamaged, ObstacleBroken, ObstacleRegrow, TeleportOut, TeleportIn, ItemPicked, BuffStart, BuffEnd, ChestIncoming, KillFeed`.
- AOI: chỉ gửi những gì nằm trong vùng nhìn của người đó (+ lề). Người **trong bụi** không được gửi cho người không được thấy (chống hack nhìn xuyên). Người **Ẩn Thân** vẫn được gửi (vì vẫn hiện mờ 20%), kèm cờ `stealth`.
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
    "tiers_per_realm": 10,
    "exp_base": 100, "realm_mult": 2.0, "tier_step": 0.08,
    "scale_min": 1.0, "scale_max": 5.0
  },
  "sword": {
    "start_count": 1, "min_count": 0,
    "max_swords": [60, 120, 160, 200, 240, 280, 320, 360, 400],
    "swords_per_ring": 20, "length": 80, "width": 12, "hilt_gap": 6,
    "ring_growth": 0.2, "ring_growth_mode": "linear", "hilt_step": 0.1, "tilt_deg": 35,
    "ring_tints": ["#FFFFFF", "#C77DFF", "#7CFC6A", "#FF6B6B", "#4CC9F0", "#FFD60A"],
    "rev_secs": 1.0, "flip_cooldown": 0.3,
    "clash_unit": 1, "flip_only_same_realm": true, "clash_knockback": 0, "drop_bundle_max": 10,
    "clash_ratio_base": 1,
    "clash_ratio_per_realm": 1
  },
  "player": {
    "body_r": 30, "move_base": 260, "pickup_r": 90,
    "invuln_on_hit": 1.0, "invuln_on_spawn": 2.0,
    "attacker_share": 0.5, "drop_share": 0.3, "kill_bonus_exp": 30,
    "death_sword_drop": 0.5, "orb_owner_secs": 1.5,
    "camera_view_h": 900, "camera_exp": 0.85, "camera_reach_mult": 2.6,
    "hit_drops_realm": true
  },
  "burn": { "mult": 1.8, "exp_frac_per_sec": 0.08, "keep_realm": true },
  "buffs": {
    "spin":   { "mult": 1.6, "secs": 12 },
    "haste":  { "mult": 1.4, "secs": 10 },
    "magnet": { "mult": 3.0, "secs": 15 },
    "long":   { "mult": 1.3, "secs": 12 },
    "shield": { "secs": 8 },
    "exp2":   { "mult": 2.0, "secs": 20 },
    "stealth":{ "secs": 5, "max_secs": 5, "visible_alpha": 0.2, "reveal_r": 120, "bot_detect_mult": 0.5 }
  },
  "items": {
    "weights": { "sword": 40, "bundle": 6, "spin": 12, "haste": 12, "magnet": 8, "long": 7, "shield": 7, "exp2": 5, "stealth": 3 },
    "sword_item_amount": 3, "bundle_amount": 15,
    "count_per_room": 400, "respawn_secs": 3, "overcap_exp": 20, "chest_every_secs": 60, "chest_items": 8
  },
  "exp_orbs": {
    "cell": 160, "values": [5, 20, 60], "weights": [80, 18, 2],
    "respawn_secs": 5, "respawn_region": 1000
  },
  "obstacles": {
    "footprint_w": 0.9, "footprint_h": 0.3, "occlusion_alpha": 0.5,
    "regrow_secs": 30, "regrow_warn_secs": 3,
    "hp": { "fence": 30, "stall": 80, "brick": 250, "boulder": 800 },
    "dmg_cooldown_secs": 0.5, "dmg_per_extra_ring": 0.5,
    "stall_item_drop": 0.3
  },
  "zones": { "mud_mult": 0.6, "wind_with": 1.5, "wind_against": 0.8, "wind_with_max_deg": 60, "wind_against_min_deg": 120, "bush_reveal_r": 150, "bush_reveal_on_hit_secs": 2 },
  "gates": { "radius": 150, "transit_secs": 2.0, "exit_warn_secs": 0.5, "cooldown_secs": 8 },
  "titles": ["Bần tăng chưa ngán ai bao giờ", "Ngon thì nhào vô", "Chỗ này nhiều máu nè"],
  "title_hold_secs": 3,
  "minimap": { "size_pc": 180, "size_mobile": 120, "show_top1": true, "show_chest": true, "show_others": false },
  "score": {
    "hit_base": 100, "kill_bonus": 50, "upset_per_realm": 0.5,
    "breakthrough_per_realm": 20, "top1_per_sec": 1,
    "bot_victim_mult": 0.5, "same_victim_cooldown_secs": 60
  },
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
- [ ] Cùng 1 tầng: 1 → 20 kiếm, bán kính vòng không đổi; kiếm thứ 21 mới mở vòng 2 rộng hơn.
- [ ] Luyện Khí tối đa 60 kiếm = 3 vòng so le, mỗi vòng một màu, cả 3 lớp đều nhìn thấy; lưỡi nghiêng về phía sau và lật khi đổi chiều.
- [ ] Mũi kiếm vòng 3 xa tâm hơn vòng 1; hitbox server khớp với hình vẽ (bật chế độ debug vẽ capsule để so).
- [ ] Kiếm quay đúng 1 vòng / 1 giây (0.625 giây khi có Tốc Kiếm), nhìn rõ từng thanh, không bị hiệu ứng quay ngược.
- [ ] Nhân vật đi sát mép trên tảng đá được (chỉ bị chặn ở chân đế), đứng sau đá thì bị che.
- [ ] Farm không đánh nhau: Luyện Khí → Trúc Cơ trong 30–60s (đo bằng cmd/sim).
- [ ] 2 người cùng cảnh giới va kiếm → mỗi bên mất 1 kiếm, cả hai đổi chiều quay.
- [ ] Trúc Cơ va Luyện Khí → Luyện Khí mất 2, Trúc Cơ mất 1, **không ai đổi chiều**.
- [ ] Luyện Khí (≥4 kiếm) va Nguyên Anh → Luyện Khí mất 4, Nguyên Anh mất 1, không ai đổi chiều.
- [ ] Luyện Khí còn 3 kiếm va Nguyên Anh → Luyện Khí còn 0, Nguyên Anh chưa mất (nứt 3/4).
- [ ] Người 0 kiếm vẫn sống; bị chém thân thì rớt cảnh giới (Luyện Khí thì chết).
- [ ] Bị chém ở Trúc Cơ tầng 2 → Luyện Khí tầng 2; kiếm vượt 60 rơi ra thành bó.
- [ ] Bị chém ở Luyện Khí (tầng bất kỳ) → màn hình Tổng kết, bấm "Chơi tiếp" vào lại được; người chém nhận 50% EXP nạn nhân + 30; xác rơi 50% số kiếm.
- [ ] Viên EXP từ quái chỉ người chém nhặt được trong 1.5s đầu.
- [ ] Đi chéo 90° so với gió → tốc độ không đổi.
- [ ] Giữ Shift khi đang chạy → nhanh ×1.8, EXP tụt dần, có vòng lửa; đứng yên giữ phím thì không mất EXP.
- [ ] Đốt tới đáy cảnh giới thì tự tắt, không rớt cảnh giới; Luyện Khí đốt cạn về 0 EXP vẫn sống.
- [ ] Vật cản bị chém vỡ → đi qua được → 30s sau mọc lại, không kẹt người.
- [ ] Kiếm chém vật cản không bị mất.
- [ ] Luyện Khí 60 kiếm phá rào tre ≈ 7.5s, không phá nổi đá tảng trong 1 phút; Đại Thừa phá đá tảng ≈ 4s.
- [ ] Vào cổng → ra ở cổng khác, cổng đích lóe sáng trước, không có bất tử khi ra.
- [ ] Người scale lớn không chui lọt hẻm 150.
- [ ] Buff hết hạn đúng giờ, nhặt lại thì làm mới thời gian.
- [ ] Ẩn Thân: người khác thấy mờ 20%, không tên; tối đa 5s; chém trúng người là hiện ngay.
- [ ] Trúc Cơ chém trúng Nguyên Anh → +800 Uy Danh; chém lại cùng người trong 60s → +0.
- [ ] Bảng Thần Tốc chỉ có những mạng đã đạt Đại Thừa.
- [ ] Top 1/2/3 phòng hiện đúng danh hiệu trên đầu và trong bảng; đổi hạng phải giữ 3s mới đổi danh hiệu.
- [ ] Minimap thấy mình, khung camera, cổng, Top 1, Rương Báu; không thấy người khác; Top 1 vào bụi thì biến khỏi minimap.
- [ ] Phòng 5 người → 15 bot; thêm người không xóa bot; người thứ 21 sang phòng mới.
- [ ] Leaderboard ghi đúng 3 bảng sau khi chết, bot không có trong bảng toàn cục.
