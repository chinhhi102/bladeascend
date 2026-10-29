# Câu hỏi đã xong (lưu trữ)

Những câu đã được trả lời — giữ lại để tra sau, không cần đọc nữa.
Câu **đang chờ** nằm ở `docs/QUESTIONS.md`.

| Câu | Nội dung | Kết luận |
|---|---|---|
| Q1–Q6 | Thưởng hạ sát, gió theo góc, viên EXP thuộc người chém, chết rơi kiếm, 1 anim/ảnh… | Trả lời trong `docs/_raw/ANSWERS.md`, đã áp dụng |
| Q7 | Nhịp Luyện Khí → Trúc Cơ | Chốt 30–60s (nay đã đổi lại, xem Q13) |
| Q8 | Checklist §17 ghi số kiếm mất theo bản cũ | Spec gốc đã sửa: `clash_unit` = 1, checklist khớp |
| Q9 | `min_count = 0` có làm chết không | Không. Spec §5.2: "0 kiếm **không chết**, chỉ bị tước vũ khí" |
| Q10 | Phạm vi nhặt vật phẩm | Giữ `pickup_r` = 90. Luật vốn đã đúng: kiếm không nhặt được đồ |
| Q11 | "Tính exp dư ra làm…" | Spec §3.2: Đại Thừa tầng 10 thì EXP thừa **bị bỏ**, xếp hạng dùng Uy Danh |
| Q12 | Tốc độ chạy 130 hay 260×sqrt(scale) | Gộp vào Q14 |
| Q13 | Nhịp farm quá chậm | Chấp nhận nhịp hiện tại; lên cấp chủ yếu nhờ cướp Linh khí. `move_base` → 195, `cell` → 140 |
| Q14 | `_raw` thiếu 13 mục so với code | Spec v4.0 đã gộp đủ 13 mục; `realm_mult` giữ **2.0** |

---

## Chi tiết

### Q7. Nhịp Luyện Khí → Trúc Cơ *(2026-10-22)*
Anh chốt **30–60s**. Lúc đó dùng `exp_orbs.cell = 160` → sim `--farm` đo 40–46s.
→ Sau đó anh đổi EXP gấp đôi + giảm mật độ 50% nên nhịp này không còn đúng. Xem **Q13**.

### Q8. Checklist §17 còn số kiếm mất theo bản cũ *(2026-10-22)*
§5.2 ghi mỗi lần va mất `clash_unit` kiếm, nhưng checklist §17 ghi số khác.
→ **Xong:** spec gốc v3.5 đặt `clash_unit = 1` và sửa checklist cho khớp ("mỗi bên mất 1 kiếm").

### Q9. `min_count = 0` không làm người chơi chết *(2026-10-22)*
Anh bảo "dù chỉ còn 1 kiếm cũng không chết → min kiếm = 0". Em chỉnh `min_count` = 0 nhưng
báo lại rằng số kiếm chưa bao giờ quyết định sống chết — `applyBodyHit` chỉ cho chết khi
nạn nhân đang ở Luyện Khí.
→ **Xong:** spec gốc §5.2 chốt luôn: *"0 kiếm **không chết**: chỉ bị tước vũ khí. Chém trúng
thân vẫn theo §5.1."* Đúng phương án đang code.

### Q10. Phạm vi nhặt vật phẩm *(2026-10-22)*
Anh bảo "phạm vi kiếm không được nhặt đồ, chỉ phạm vi nhân vật". Kiểm tra thì luật đã đúng
sẵn: `resolvePickups` chỉ dùng `pickup_r × scale`, kiếm chỉ đánh quái và vật cản.
Cảm giác "kiếm hút đồ" là do vòng nhặt rộng bằng ~77% vòng kiếm.
→ **Xong:** spec gốc giữ `pickup_r = 90`, không đổi gì.

### Q11. "Tính exp dư ra làm…" — câu bị cắt *(2026-10-22)*
Câu nhắn đứt giữa chừng nên em chưa dám đoán.
→ **Xong:** spec gốc §3.2 đã ghi rõ: *"Đạt Đại Thừa tầng 10: `total_exp` dừng ở trần, EXP
nhặt thêm bị bỏ (xếp hạng dùng Uy Danh §11, không dùng EXP)."* Đúng với code hiện tại.

### Q12. Tốc độ chạy: spec cứ lùi về 260 × sqrt(scale) *(2026-10-22)*
Anh nhắn "giảm 1/2, chỉ item mới tăng tốc" → code `move_base` = 130, bỏ `sqrt(scale)`.
Spec dán vào sau đó nhiều lần vẫn ghi 260 × sqrt(scale).
→ **Gộp vào Q14**: nó là 1 trong 6 con số mà `docs/_raw/` còn ghi giá trị cũ. Code vẫn giữ 130.

### Q13. Nhịp độ quá chậm sau 3 thay đổi *(trả lời 2026-10-22, CHANGES_v4)*
Anh chốt: **chấp nhận nhịp hiện tại** — lên cấp chủ yếu nhờ cướp Linh khí của người khác chứ
không cần đạt mốc farm, §6.1 đã sửa theo. Kèm hai điều chỉnh: `move_base` 130 → **195** (+50%)
và `exp_orbs.cell` 226 → **140**; `items.count_per_room` giữ 200.

### Q14. `docs/_raw/GAME_SPEC.md` thiếu 13 mục *(trả lời 2026-10-22, CHANGES_v4)*
Anh đã gộp đủ 13 mục vào spec v4.0 và dán lại vào `_raw`. `realm_mult` giữ **2.0**
(mỗi cảnh giới gấp đôi cảnh giới trước). Danh hiệu Top 4–10 giữ bộ chữ đang có trong code.

### Q15. Công thức lá chắn §5.2 làm không ai chém chết được ai *(xử lý 2026-10-28)*
Không cần anh trả lời nữa: nguyên nhân là **hiểu lá chắn thành tường chắn đường đi**. Anh báo
"đi tới gần người khác là không điều khiển nhân vật tiến lên được" → chuyển lá chắn thành
**luật chém** thay vì luật di chuyển:
- Người chơi **không chặn đường nhau** nữa, hai vòng kiếm chồng lên nhau mà mài.
- `resolveSwordBody` bỏ qua nạn nhân còn `ShieldR > 0` (tức còn ≥ 2 vòng). Mài rụng hết vòng
  ngoài, còn đúng 1 vòng thì §5.1 mới chém được thân — đúng câu "chỉ khi người đó còn 1 vòng
  thì kiếm đối thủ mới chạm được tới thân".

Công thức đối xứng `max(reachA + shield_B, reachB + shield_A)` trong spec không còn dùng tới,
vì nó chỉ có nghĩa khi lá chắn là tường chắn đường.

### Q16–Q19 *(anh trả lời 2026-10-28 trong `docs/_raw/ANSWERS.md`)*
- **Q16 Trường Kiếm**: bỏ hẳn. Spec v4.3 xoá khỏi §4.1a/§6.2, v4.4 thay bằng **Cương Kiếm**.
  Ảnh `assets/buffs/long.png` giữ lại nhưng không dùng.
- **Q17 Bó Kiếm khi rớt cảnh giới**: chấp nhận phương án 1 — Bó Kiếm chủ yếu từ xác người
  chết và vật phẩm bản đồ.
- **Q18 Linh khí / Tiên khí**: giữ như đã làm, không quy đổi qua lại. Bỏ
  `tien_khi_to_exp_base`. Người phàm chém người Tiên giới chỉ nhận Uy Danh và ngược lại.
- **Q19 Tốc độ bể kiếm**: giữ `clash_cooldown: 0`, `flip_cooldown: 0.3`.

## Q21 — Hai luật mới chủ dự án yêu cầu, GAME_SPEC chưa ghi

Chủ dự án yêu cầu trực tiếp ngày 29/10/2026, tôi đã làm trong code nhưng **không được ghi
vào `docs/_raw/`**, nên spec và code đang lệch nhau. Nhờ anh cập nhật spec:

**1. Cổng dịch chuyển bắn ngẫu nhiên** (lệch §8.4 — spec hiện ghi "cổng đi 2 chiều, ghép
cặp cố định"). Đã làm: bước vào một cổng thì ra ở **một cổng bất kỳ khác cùng mạng**, không
bao giờ ra lại chính cổng vừa vào. Bật/tắt bằng `gates.random_exit` (mặc định `true`); đặt
`false` là quay về đúng luật cũ.

Lưu ý tôi hiểu câu "cổng dịch chuyển sẽ dịch chuyển ngẫu nhiên" là **đích đến ngẫu nhiên**.
Nếu ý anh là **bản thân cái cổng tự di chuyển chỗ trên bản đồ** thì nói một tiếng, đó là
việc khác hẳn và tôi làm lại.

**2. Đạo Tổ: Tiên khí dư đổi thành Uy Danh** (spec chưa nói gì về chuyện này). Đã làm: tới
Đạo Tổ là hết cảnh giới để lên, nên mỗi điểm Tiên khí nhặt thêm đổi thành
`cultivation.tien_overflow_score` điểm Uy Danh — chủ dự án chốt tỉ lệ **10 Tiên khí = 1 Uy
Danh**, tức `0.1` và thanh Tiên khí **giữ đúng ở
mốc Đạo Tổ** thay vì phình mãi.

Vì sao giữ thanh ở mốc đỉnh: `robTienKhi` tính `loss = tiên_khí − mốc_khởi_điểm_mới`. Nếu
cho phình thì một người Đạo Tổ ngồi lâu thành cục mồi khổng lồ, ai chém được là giàu to —
không chặn thì luật cướp đoạt §3.4 méo hẳn. Nếu anh muốn giữ nguyên cục mồi đó thì báo.

**Đã trả lời (CHANGES_v5.md, 29/10/2026)**: cổng đúng là **đích ngẫu nhiên**, không phải cổng
tự đổi chỗ — spec §8.4 đã viết lại thành mạng chung 20 cổng của cả 5 khu. Đạo Tổ đổi Tiên khí
dư sang Uy Danh thì giữ nguyên cách đã làm, ghi vào §3.4.

**Còn một điểm lệch**: CHANGES_v5 ghi `tien_overflow_score` là **1.0**, nhưng sau đó anh nhắn
trực tiếp "10 tiên khí chỉ được 1 uy danh" tức **0.1**. Tôi đang để **0.1** theo lời nhắn (mới
hơn và nói rõ tỉ lệ). Xem Q22.
