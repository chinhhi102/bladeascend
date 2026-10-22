# Câu hỏi về thiết kế (cần anh xác nhận)

Mỗi câu: bối cảnh → các phương án → **tạm dùng** (đã code theo phương án này, đổi được dễ).

Q1–Q6 đã trả lời trong `docs/ANSWERS.md` và đã áp dụng.

## Q7. (đã trả lời 2026-10-22) Nhịp Luyện Khí → Trúc Cơ

Anh chốt: **30–60s là được** (spec §6.1 đang ghi 60–90s, nhờ sửa lại cho khớp). Đang dùng `exp_orbs.cell = 160` → sim `--farm` đo được 40–46s. Các cảnh giới sau giữ nguyên luật (đo: Kim Đan ≈ 1 phút, Nguyên Anh ≈ 1.2 phút).

## Q8. Checklist §17 còn số kiếm mất khi va chạm theo bản cũ

**Bối cảnh:** §5.2 mới ghi mỗi lần va mất `clash_unit` = 5 kiếm (bảng: Trúc Cơ −5 / Luyện Khí −10…), nhưng checklist §17 vẫn ghi "mỗi bên mất 1 kiếm", "Luyện Khí mất 2, Trúc Cơ mất 1", "Luyện Khí mất 4, Nguyên Anh mất 1".

**Tạm dùng:** theo §5.2 (nhân 5), unit test đã viết theo §5.2. Anh sửa giúp checklist §17 nếu đúng vậy.

## Q9. `min_count = 0` không làm người chơi chết (22/10/2026)

**Bối cảnh:** anh bảo "dù chỉ còn 1 kiếm người chơi cũng không thể chết → chỉnh min kiếm = 0". Em đã chỉnh `min_count` 1 → 0 (`config/game.json`, spec §5.2, test `TestRingCanBeStrippedToZero`). Nhưng đọc lại code thì **số kiếm chưa bao giờ quyết định sống chết**: `applyBodyHit` trong `server/internal/game/combat.go:70` tính `dies = Prog.Realm(b.Level) == 0` — tức **chỉ người đang ở Luyện Khí bị chém vào thân mới chết**, từ Trúc Cơ trở lên chỉ rớt 1 cảnh giới. Vậy 0 kiếm = bị tước vũ khí, không phải chết.

**Các phương án nếu anh muốn người chơi chết được:**
1. Giữ nguyên: chỉ Luyện Khí chết (luật spec §5.1 hiện tại).
2. Thêm luật "hết kiếm = chết": ai bị đánh rớt về 0 kiếm thì chết luôn, bất kể cảnh giới.
3. "Hết kiếm thì mới chém chết được": còn kiếm → rớt cảnh giới; 0 kiếm mà bị chạm thân → chết.

**Tạm dùng:** phương án 1 (chưa đổi luật chết), chỉ áp dụng đúng yêu cầu `min_count = 0`.

## Q10. Phạm vi nhặt vật phẩm (22/10/2026)

**Bối cảnh:** anh bảo "phạm vi của kiếm không thu thập được vật phẩm, chỉ phạm vi nhân vật mới lượm". Kiểm tra code thì **luật này đã đúng sẵn**: `resolvePickups` (`server/internal/game/items.go:83`) chỉ dùng `PickupRadius = pickup_r × scale`, còn kiếm (`resolveSwordWorld`) chỉ đánh quái và vật cản, không chạm vào item. Client không hề tự nhặt.

**Lý do nhìn như kiếm đang hút đồ:** `pickup_r` = 90 trong khi tầm với của vòng kiếm trong cùng ≈ 117 (cùng nhân `scale`) → vòng nhặt rộng bằng ~77% vòng kiếm, nên item biến mất gần như đúng lúc lưỡi kiếm lướt qua.

**Các phương án:**
1. Giữ 90 (không đổi gì, luật vốn đã đúng ý).
2. Giảm còn ~45 (1.5 × `body_r`) → vòng nhặt nằm hẳn trong vòng kiếm, phải đi đè lên item mới lượm.
3. Giảm còn ~60 (2 × `body_r`) — mức trung gian.

**Tạm dùng:** phương án 1, chờ anh chốt con số.

## Q11. "Tính exp dư ra làm ..." — câu bị cắt (22/10/2026)

**Bối cảnh:** anh nhắn "hãy tính exp dư ra làm" nhưng câu chưa hết nên em chưa dám đoán. Chỗ EXP hiện đang bị bỏ phí trong code: `GainExp` (`server/internal/game/exp.go:13`) kẹp `total_exp` tại `Prog.MaxExp()` → khi đã Đại Thừa tầng 10 thì **mọi EXP nhặt thêm đều bị vứt bỏ** (không vào bảng xếp hạng, không đổi được gì).

**Các phương án em đoán:**
1. EXP dư vẫn cộng vào điểm bảng xếp hạng dù đã kịch cảnh giới.
2. EXP dư tự đổi thành kiếm (giống `overcap_exp` nhưng ngược chiều).
3. EXP dư khi lên tầng được giữ lại (thật ra đã đúng sẵn: level suy ra từ `total_exp` nên không mất gì).

**Tạm dùng:** chưa làm gì, chờ anh nói rõ.

## Q12. Tốc độ chạy: spec dán vào cứ lùi về 260 × sqrt(scale) (22/10/2026) — **CẦN ANH CHỐT**

**Bối cảnh:** sáng nay anh nhắn trực tiếp: *"tốc độ chạy hơi nhanh → chỉnh lại 1/2, và chỉ lượm được item mới tăng tốc chứ không tăng theo tu vi"*. Em đã làm: `move_base` 260 → 130 và bỏ hệ số `sqrt(scale)`.

Sau đó anh dán spec mới **hai lần** (v3.4 rồi v3.5/v3.6), lần nào §4.1 cũng quay về `260 * sqrt(scale)` kèm nguyên câu cũ "≈260 px/s so với ≈150 px/s ở scale 5". Em đoán đây là do file spec được sinh lại từ bản gốc chứ không phải anh đổi ý (vì những chỗ anh sửa tay như min_count = 0, clash_unit = 1, cell = 160 thì đều được giữ nguyên qua các lần dán). **Tạm dùng:** giữ 130, không có `sqrt(scale)`, và sửa lại §4.1 cho khớp.

**Hệ quả đo được (`cmd/sim --farm --bots 20 --seed 7`):** tốc độ chậm đi một nửa nên nhặt EXP cũng chậm theo → Luyện Khí lên Trúc Cơ mất **78.5s**, trong khi §6.1 yêu cầu **30–60s**. Hiện game đang KHÔNG đạt mốc này.

| `move_base` | `exp_orbs.cell` | Luyện Khí → Trúc Cơ |
|---|---|---|
| 130 (đang dùng) | 160 (đang dùng) | 78.5s ❌ |
| 130 | 120 | 53.3s ✅ |
| 260 × sqrt(scale) | 160 | ~40–46s ✅ (số cũ trong spec) |

**Ba phương án, cái nào cũng sửa 1 dòng:**
1. Giữ tốc độ 130, **giảm `exp_orbs.cell` 160 → 120** để nhịp về lại 53s. (Em nghiêng về cách này vì bám đúng yêu cầu anh nhắn.)
2. Giữ tốc độ 130, chấp nhận nhịp 78s và sửa §6.1 thành 60–90s.
3. Trả tốc độ về `260 * sqrt(scale)` đúng như spec đang ghi, giữ cell 160.

Em chưa đổi gì thêm, chờ anh chọn.

## Q13. Nhịp độ sau khi EXP gấp đôi + mật độ giảm 50% (22/10/2026)

**Bối cảnh:** anh yêu cầu 3 thay đổi đều làm chậm tiến độ: tốc độ chạy còn một nửa (Q12), mỗi cảnh giới tốn gấp đôi cảnh giới trước (`realm_mult` 1.2 → 2.0), mật độ EXP và vật phẩm giảm 50% (`cell` 160 → 226, `count_per_room` 400 → 200). Em đã làm đủ cả ba.

**Đo được (`sim --farm --bots 20 --seed 7`):**

| Mốc | Trước hôm nay | Bây giờ | §6.1 đang ghi |
|---|---|---|---|
| Luyện Khí → Trúc Cơ | 40–46s | **165s** | 30–60s ❌ |
| → Kim Đan | ~60s | **321s** | ≈1 phút ❌ |
| → Nguyên Anh | ~72s | **417s** | ≈1.2 phút ❌ |
| Tổng EXP tới Đại Thừa 10 | 28 288 | **694 960** | – |

**Hai điểm cần anh quyết:**

1. **§6.1 đang mâu thuẫn với config.** Spec vẫn ghi mục tiêu "Luyện Khí → Trúc Cơ 30–60s" nhưng thực tế 165s. Em không tự sửa mục tiêu trong spec (quy tắc 4). Nếu 165s là ý anh thì em sửa §6.1 thành "≈2.5–3 phút"; nếu không thì giảm `cell` lại.

2. **Đại Thừa gần như không tới được.** Tốc độ farm đo được ≈25 EXP/s → 694 960 EXP ≈ **7.7 giờ farm liên tục**. Hệ quả: bảng **Thần Tốc** (thời gian lên Đại Thừa) sẽ mãi trống, và phần cuối cây tu luyện không ai thấy. Nếu muốn giữ EXP gấp đôi mà vẫn tới được Đại Thừa trong một ván .io (~10–20 phút), cần tăng nguồn EXP ở cảnh giới cao — ví dụ viên EXP to hơn theo cảnh giới, hoặc EXP từ quái/chém người nhân theo cảnh giới.

Em chưa đổi gì thêm, chờ anh chốt.
