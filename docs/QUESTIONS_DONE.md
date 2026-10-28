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
