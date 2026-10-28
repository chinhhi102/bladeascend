# Việc cần làm — spec v4.0

**Trước tiên:** docs/GAME_SPEC.md v4.0 đã gộp toàn bộ 13 mục trong Q14 (những gì anh nhắn trực tiếp). Anh sẽ dán bản này vào `docs/_raw/GAME_SPEC.md`. Đối chiếu với code; chỗ nào spec v4.0 ghi khác code hiện tại thì làm theo spec, trừ danh hiệu Top 4–10 (giữ bộ chữ đang có trong code, chỉ chuyển vào config `titles`).

## Trả lời Q13, Q14
- Q13: chấp nhận nhịp hiện tại. Lên cấp chủ yếu nhờ cướp Linh khí, không cần đạt mốc farm. §6.1 đã sửa.
- Q14: `realm_mult` giữ **2.0**.

## Thay đổi số
1. `player.move_base` 130 → **195** (+50%), vẫn không nhân sqrt(scale). §4.1
2. `exp_orbs.cell` 226 → **140**. `items.count_per_room` giữ 200.
3. `invuln_on_spawn` 5.0, `clash_knockback` 0 (đã đúng trong code).

## Lá chắn vòng kế ngoài cùng (§5.2)
4. Không đẩy văng. Mỗi người có `shield_r` = bán kính mũi kiếm vòng kế ngoài cùng (còn 1 vòng thì = thân). Server ràng buộc `dist(A,B) ≥ max(reachA + shield_rB, reachB + shield_rA)` như va tường (trượt, không bật). Kiếm bể trừ từ vòng ngoài cùng; vòng ngoài hết → shield lùi 1 lớp.
5. Bot AI: hiểu rằng phải mài hết vòng ngoài mới tiến vào được.

## Tiên khí & Đài Tiên Khí (§3.4, §8.5) — bỏ sát mạng
6. Xóa toàn bộ logic `tien_kills`. Thay bằng `tien_khi` (số nguyên), mốc [100, 200, 300, 400, 500].
7. 4 Đài nhỏ (mỗi 5s phun 2–3 viên 2 điểm; mỗi 30s phun 15 viên 2 điểm + 2 viên 10 điểm, báo trước 3s) + Thiên Tiên Đài trung tâm (mỗi 150s 1 viên 150 điểm, báo toàn phòng trước 10s). Viên tồn tại 60s.
8. Người dưới Đại Thừa 10 nhặt Tiên khí → đổi `50 × 2^r` Linh khí mỗi điểm.
9. Bị chém ở Tiên giới: Tiên khí về mốc khởi điểm cảnh giới mới; người chém nhận 50% (đổi Linh khí nếu chưa đủ điều kiện), 30% rơi thành viên Tiên khí.
10. Minimap hiện 5 đài, nhấp nháy khi sắp phun lớn.

## Sinh bản đồ (§8.1)
11. Viết lại `mapgen` theo 9 bước §8.1: Thiên Tiên Đài, 4 sân môn phái có Đài Tiên Khí, đường chính + vành đai, đối xứng xoay 4 phía, 4 biome theo Voronoi nhiễu, Poisson-disk + cụm, cổng nối chéo, kiểm tra liên thông. Thêm `cmd/mapgen --seed N --png out.png` xuất ảnh xem trước. Art mới (đường đá, đài) dùng placeholder trước.

## Sửa lỗi
12. **Tốc Kiếm đang tăng tầm chém — sai.** `buff_spin` chỉ được nhân vào ω. Kiểm tra lại chỗ áp buff (có thể đang nhầm sang `buff_long`) và icon/aura của hai buff. Thêm unit test: có Tốc Kiếm thì `reach` không đổi, góc sau 1s = 360° × 1.6.

## Test
13. Theo checklist §17 (các dòng mới ở đầu mục).

## v4.1 — Độ bền kiếm (bổ sung)
14. Bỏ `clash_ratio_base`, `clash_ratio_per_realm`. Thêm `durability_per_realm: 1.2`, `durability_mode: "compound"`.
15. Mỗi lần va khác cảnh giới: bên thấp mất `clash_unit` (1) kiếm; bên cao `crack += lost / k` với `k = 1.2^d`. Cùng cảnh giới: mỗi bên −1, đổi chiều (như cũ).
16. Unit test: Trúc Cơ vs Luyện Khí sau 12 lần va → Luyện Khí −12, Trúc Cơ −10. Mode "linear": `k = 1 + 0.2 × d`.

## v4.2 — VFX thần thú Tiên giới
17. Đọc spec §3.5. Làm ở client: thần thú dạng chuỗi đốt (đầu + 12–20 đốt + đuôi) chạy theo quỹ đạo xoắn ốc từ chân lên; đốt ở nửa sau vòng xoắn vẽ dưới nhân vật, nửa trước vẽ trên.
18. Kích hoạt: event Breakthrough vào cảnh giới Tiên (bản to 2.5s), đồng hồ client mỗi 20s (bản nhỏ 1.5s, tắt được trong cài đặt), event RealmDrop (vỡ vụn 0.8s). Tối đa 6 con trên màn hình.
19. Placeholder trước (tam giác + hình tròn tô màu theo bảng §3.5); art theo ART_PROMPTS §10.1 thả vào sau.

## v4.3 — Bỏ vật phẩm Trường Kiếm
20. Xóa buff/item `long` (Trường Kiếm) khỏi: config (`buffs.long`, `items.weights.long`), server (spawn, áp buff, công thức chiều dài kiếm — bỏ hệ số `buff_long`), gói tin (bỏ field `buff_long`), client (icon HUD, aura, art loader), bot AI, phím debug buff, assets:report.
21. Item `long` còn trên bản đồ của phòng đang chạy: bỏ qua khi load (không crash).
22. Chạy lại toàn bộ test; xóa/sửa test liên quan Trường Kiếm.
