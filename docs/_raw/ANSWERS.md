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

---

## v3.7 — Tiên giới (5 cảnh giới mới sau Đại Thừa)
Đọc spec §3.4 (mới). Tóm tắt việc cần làm:
1. Config: thêm `tien_realms`, `tien_kills_per_step`, `tien_bot_kill_value`, `tien_same_victim_cooldown_secs`, `tien_scale_step`, `tien_extra_swords`; mở rộng `max_swords` tới 600.
2. Server: thêm bộ đếm `tien_kills` (float vì bot = 0.5). Từ Đại Thừa tầng 10 chỉ lên bằng sát mạng (mỗi bước 10/20/30/40/50). 1 sát mạng = chém trúng thân thành công; cùng người trong 60s không tính. Tiên giới không chia tầng.
3. Bị chém ở Tiên giới: rớt 1 cảnh giới, `tien_kills` về mức khởi điểm của cảnh giới mới; Chân Tiên rớt về Đại Thừa tầng 10 (EXP giữ trần, `tien_kills` = 0).
4. Scale Tiên giới: 5.2 → 6.0 (+0.2 mỗi cảnh giới). Kiếm tối đa +40 mỗi cảnh giới.
5. Client: HUD thanh Sát mạng thay thanh EXP; hiệu ứng Thiên Kiếp + kill feed phi thăng; hào quang Đạo Tổ; Đạo Tổ hiện trên minimap cả phòng; bảng phòng xếp Tiên giới trên Đại Thừa.
6. Uy Danh, bể kiếm, sát thương vật cản dùng nguyên công thức với r tới 13.
7. Skin kiếm & VFX mới trong ART_PROMPTS §7, §10 (dùng placeholder trước).
8. Unit test theo checklist §17.

---

## v3.8 — Số kiếm theo vòng + Linh khí / Tiên khí
Đọc spec §3.1, §3.4, §4.1a, §8.2.
1. Vòng thứ k (1-index) chứa `10 × k` kiếm. Mỗi cảnh giới mở thêm 1 vòng: cảnh giới r có `r + 1` vòng, `max_swords = 5 × (r+1) × (r+2)` → [10, 30, 60, 100, 150, 210, 280, 360, 450, 550, 660, 780, 910, 1050]. Config: `ring_capacity_step: 10`, bỏ `swords_per_ring` và `tien_extra_swords`.
2. Lấp vòng trong trước; vòng đầy chia đều `36°/k`, vòng ngoài chưa đầy chia đều `360°/m`. So le: vòng k ≥ 2 lệch nửa khe của chính nó. Bán kính/chiều dài vòng vẫn chỉ phụ thuộc k và scale.
3. UI: dưới Tiên giới đổi mọi chữ "EXP" thành **Linh khí**; từ Chân Tiên trở lên thanh tiến độ là **Tiên khí** (đầy bằng sát mạng, màu vàng kim). Giữ nguyên thanh, chỉ đổi tên/màu.
4. Sửa unit test & checklist §17 theo số mới (Luyện Khí 10 kiếm 1 vòng, kiếm thứ 11 mở vòng 2...). Bảng thời gian phá vật cản §8.2 đã tính lại theo số vòng mới.

---

## v3.9 — Hào quang Tiên giới
Đọc spec §3.4 (mục "Hào quang sau lưng") và ART_PROMPTS §10.
1. Config: thêm `tien_aura_colors`.
2. Client: lớp hào quang vẽ sau lưng nhân vật (dưới sprite nhân vật, dưới vòng kiếm, trên mặt đất), tâm ngang ngực, đường kính ≈ 2.2 × chiều cao nhân vật, to theo scale, nhịp thở ±5% chu kỳ 2s. Màu theo cảnh giới Tiên; Đạo Tổ dùng sprite âm dương riêng.
3. Ẩn Thân/bụi: hào quang ẩn/mờ theo nhân vật. Rớt cảnh giới: đổi màu ngay, về Đại Thừa thì tắt, kèm hiệu ứng vỡ.
4. Chưa có art: placeholder là hình tròn phát sáng (radial gradient) theo màu config.

---

## Q16 — Trường Kiếm
**Bỏ hẳn** (phương án 2). Spec v4.3 đã xóa khỏi §6.2 và §4.1a, v4.4 thay bằng **Cương Kiếm** — anh dán spec mới nhất vào `docs/_raw/`. Làm theo CHANGES_v4.md mục 20–23. Ảnh `assets/buffs/long.png` để đó, không dùng.

## Q17 — Bó Kiếm khi rớt cảnh giới
**Chấp nhận phương án 1.** Spec §5.2 đã ghi rõ: Bó Kiếm chủ yếu từ xác người chết và vật phẩm bản đồ. Luật chém thân theo số vòng (Q15) đã được ghi vào spec thay cho lá chắn chặn đường (config `body_hit_max_rings: 1`, bỏ `shield_ring`).

## Q18 — Linh khí / Tiên khí tách hẳn
**Giữ đúng như em đã làm.** Spec §3.4 đã sửa: không quy đổi qua lại, người phàm chém người Tiên giới chỉ nhận Uy Danh (và ngược lại). Bỏ `tien_khi_to_exp_base` khỏi config.

## Q19 — Tốc độ bể kiếm
**Giữ như em đã làm**: `clash_cooldown: 0` (bể mỗi tick), `flip_cooldown: 0.3`. Spec §5.2 đã sửa.

---

## Trả lời báo cáo đo (mục 24)

Nhận xét: client trong game thật đang ổn (60 fps, vẽ 1 ms). Chỗ vượt ngân sách là **server tick** và **băng thông** (61,7 KB/s trong game thật, đỉnh snapshot 38 KB). Trên 4G yếu, gói to và đều đặn dễ gây nghẽn hàng đợi → ping hiển thị tăng. Vì vậy ưu tiên mạng và server trước client.

### 1. Tối ưu — đi tiếp, theo thứ tự này
1. **Mục 27 (mạng) trước**: delta snapshot, lượng tử hóa, item/EXP chỉ gửi sự kiện, **chia nhỏ ItemsAdded** (tối đa ~2 KB mỗi gói, phần còn lại gửi dần ở các tick sau, ưu tiên gần người chơi), không gửi viên nhỏ ngoài 1 500 khi camera lùi xa, 20/10 Hz theo khoảng cách, bỏ snapshot cũ khi client nghẽn.
2. **Mục 26 (server)**: chạy `pprof` trên loadtest trước, gửi tôi top 10 hàm tốn CPU, rồi mới sửa đúng chỗ (va chạm theo vòng, chỉ vòng ngoài cùng, không cấp phát mỗi tick).
3. **Mục 25 (client)**: chỉ làm phần rẻ mà lợi lớn — vẽ mỗi vòng kiếm thành 1 texture, BitmapText, cull. Web Worker và tự hạ chất lượng để sau, chỉ làm nếu đo thấy cần.
4. Đo lại bằng loadtest + F3 sau mỗi bước, ghi bảng trước/sau vào PROGRESS.md.
5. **Thêm 1 phép đo mạng thật**: chạy client với Chrome DevTools → Network throttling "Fast 4G" và "Slow 4G", ghi ping p95 và KB/s. Ghi luôn server đang đặt ở đâu (nước/khu vực) trong PROGRESS.md.

### 2. Thư viện
**Duyệt** `pixi-filters` và `revolt-fx`. Thêm vào CLAUDE.md quy tắc: *"Thư viện được spec ghi đích danh coi như đã duyệt; thư viện khác vẫn phải hỏi."*

### 3. Thứ tự các buổi tiếp theo
1. Tối ưu (mục 25–27 như trên).
2. Bảng tên — mục 37–41 (nhỏ, thấy ngay).
3. Ảnh đại diện + bản đồ sinh động — mục 42–47 (làm sau tối ưu để đo được chi phí bake chunk).
4. VFX Tiên giới — mục 29–33, rồi thần thú luôn lượn lờ — mục 34–36 (nếu chưa làm).

Mỗi buổi xong: test + build pass, cập nhật PROGRESS.md, báo số FPS/KB/s nếu có ảnh hưởng hiệu năng.
