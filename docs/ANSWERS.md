# Trả lời QUESTIONS.md (đã cập nhật vào docs/GAME_SPEC.md)

## Q1 — EXP khi chém chết người ở Luyện Khí tầng 1
**Chốt: (b) + thưởng hạ sát.** `loss = total_exp hiện tại của nạn nhân`; người chém nhận 50% loss + `kill_bonus_exp` (30, mới trong config), rơi 30% loss thành viên EXP, +1 kill. Spec mục 5.1.

## Q2 — Đường gió
**Chốt: (b).** Lệch < 60° → ×1.5; lệch > 120° → ×0.8; còn lại ×1.0. Thêm `wind_with_max_deg: 60`, `wind_against_min_deg: 120` vào config. Spec mục 8.3.

## Q3 — Rương Báu
**Giữ phương án tạm.** Người đầu tiên chạm vào → nổ 5 item văng xung quanh, ai lượm cũng được. Spec mục 6.

## Q4 — EXP từ quái
**Giữ phương án tạm, bổ sung:** viên EXP thuộc về người chém hạ quái trong 1.5s đầu (`orb_owner_secs`), sau đó ai cũng nhặt được — tránh bị cướp công. Spec mục 5.3.

## Q5 — Chết có rơi kiếm không
**Đổi: có rơi.** Chết (người và bot) → `floor(n * death_sword_drop)` kiếm (mặc định 0.5) văng ra thành item kiếm quanh xác. Người chém có phần thưởng, người khác có cơ hội "hôi của". Thêm `death_sword_drop: 0.5` vào config. Spec mục 7.

→ Việc cần làm: cập nhật config/game.json, sửa code Q1/Q2/Q4/Q5, thêm unit test tương ứng (xem checklist mục 17), xóa các câu đã trả lời khỏi QUESTIONS.md.

## Q6 — "1 animation/sheet" có áp dụng cho quái, cổng, VFX?
**Đồng ý (a), kèm quy tắc gộp** (đã ghi vào ART_PROMPTS §5.1):
- Nhân vật: 1 anim/ảnh (120 ảnh).
- Ảnh tĩnh (decor, huy hiệu, icon, skin kiếm): gộp thoải mái.
- Vật cản: 4 trạng thái của cùng 1 vật cản chung 1 ảnh; `break`/`regrow` tách riêng.
- Quái, cổng, Rương Báu, VFX nhiều frame: gộp các anim của **cùng 1 đối tượng**, **mỗi anim 1 hàng**, ghi rõ số frame từng hàng trong prompt. Không trộn 2 đối tượng có animation trong 1 ảnh.
- Ảnh gộp bị lệch → chỉ sinh lại phần lỗi dưới dạng tách riêng.

**Việc cần làm thêm trong `cmd/artgen`**: bộ cắt frame tìm từng hình theo vùng pixel liền nhau (connected components) thay vì chia lưới cứng, sau đó căn pivot (nhân vật: chân về y=460, không co giãn từng frame; vật thể/VFX: căn giữa). Sheet hơi lệch lưới vẫn dùng được → tiết kiệm quota sinh lại.

---

## Q7 — Nhịp Luyện Khí → Trúc Cơ
**Chốt 30–60s**, giữ `exp_orbs.cell = 160`. Spec §6.1, config và checklist §17 đã sửa cho khớp.

## Q8 — Số kiếm mất khi va chạm
**Đổi:** `clash_unit = 1` (bỏ nhân 5). Checklist §17 giữ đúng các con số 1:1, 1:2, 1:4.
**Luật mới:** chỉ đổi chiều quay khi **cùng cảnh giới**. Khác cảnh giới → không ai đổi chiều (config `flip_only_same_realm: true`). Đẩy văng `clash_knockback` vẫn áp dụng mọi trường hợp. Spec §5.2 đã viết lại.

## Q9 — `min_count = 0`
**Xác nhận phương án 1:** số kiếm về 0 chỉ là bị tước vũ khí, không chết. Lý do cần về 0: vòng kiếm dày chặn hết kiếm đối thủ, phải tước sạch kiếm mới chém được vào thân. Chém thân vẫn theo §5.1 (Trúc Cơ trở lên rớt 1 cảnh giới, Luyện Khí chết). Bot 0 kiếm → Flee. Sửa test checklist: "Luyện Khí còn 3 kiếm va Nguyên Anh → còn 0, Nguyên Anh nứt 3/4".

## Q10 — Phạm vi nhặt vật phẩm
**Chưa có câu trả lời — giữ nguyên 90**, chờ anh chốt.

## Q11 — EXP dư
Xếp hạng giờ dùng Uy Danh (spec §11) nên EXP sau trần không còn tác dụng → **giữ nguyên code hiện tại** (kẹp tại trần, bỏ phần dư). Đã ghi rõ vào spec §3.2.

→ Việc cần làm: cập nhật config (`clash_unit: 1`, `flip_only_same_realm: true`, `min_count: 0`, `cell: 160`), sửa logic đổi chiều trong va chạm kiếm, sửa unit test theo checklist §17, chuyển Q7–Q9, Q11 sang mục đã trả lời trong QUESTIONS.md.

---

## v3.6 — Vật cản cứng hơn
Nguyên nhân vật cản yếu: damage tính theo **từng thanh kiếm**, nên 60–400 kiếm cùng chém thì vật cản tan ngay.
1. Đổi sang tính theo **người**: mỗi người gây damage lên 1 vật cản tối đa 1 lần / 0.5s, `damage = (r + 1) × (1 + 0.5 × (số_vòng − 1))`. Nhiều người cùng chém thì cộng dồn. (spec §5.3)
2. HP mới: rào tre 30, sạp hàng 80, tường gạch 250, đá tảng 800 (spec §8.2, có bảng thời gian phá theo cảnh giới).
3. Config: `obstacles.hp`, thêm `dmg_cooldown_secs: 0.5`, `dmg_per_extra_ring: 0.5`.
4. Unit test theo checklist §17 (Luyện Khí phá rào ≈ 7.5s, Đại Thừa phá đá ≈ 4s).
