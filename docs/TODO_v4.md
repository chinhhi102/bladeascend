# Spec v4.0 — phần chưa làm

Ba việc lớn còn lại sau phiên 22/10/2026. Phần đã làm ghi trong `docs/PROGRESS.md`.

## 1. Lá chắn vòng kế ngoài cùng (§5.2, CHANGES_v4 #4–5)

- `shield_r` = bán kính mũi kiếm của **vòng kế ngoài cùng** (còn 1 vòng → bằng bán kính thân).
- Server ràng buộc `dist(A,B) ≥ max(reachA + shield_rB, reachB + shield_rA)`, xử lý như va
  tường: **trượt dọc theo ràng buộc, không bật ra**.
- Kiếm bể trừ từ **vòng ngoài cùng** trước; vòng ngoài hết kiếm → `R` giảm 1 → shield lùi 1 lớp.
- Chỉ khi đối thủ còn **1 vòng** thì kiếm mới chạm được tới thân (§5.1).
- Bot AI phải hiểu: muốn vào chém thân thì phải mài hết vòng ngoài trước.

**Ảnh hưởng**: thay hẳn mô hình va chạm người-với-người hiện tại (giờ chỉ đẩy/không đẩy).
Cần đụng `resolveSwordSword`, `movePlayer` (ràng buộc khoảng cách), `bot.decide`.

## 2. Tiên khí + 5 Đài Tiên Khí (§3.4, §8.5, CHANGES_v4 #6–10)

Thay **toàn bộ** hệ sát mạng đang có (`Player.TienKills`, `addTienKill`, `tien_kills_per_step`):

- `tien_khi` (số nguyên), mốc **[100, 200, 300, 400, 500]**.
- **4 Đài nhỏ**: mỗi 5s phun 2–3 viên 2 điểm; mỗi 30s phun 15 viên 2 điểm + 2 viên 10 điểm,
  báo trước 3s. **Thiên Tiên Đài** trung tâm: mỗi 150s 1 viên **150 điểm**, báo toàn phòng
  trước 10s. Viên tồn tại 60s.
- Người **dưới** Đại Thừa 10 nhặt Tiên khí → đổi thành `50 × 2^r` Linh khí mỗi điểm.
- Bị chém ở Tiên giới: Tiên khí về mốc khởi điểm cảnh giới mới; **người chém nhận 50%**
  (đổi Linh khí nếu chưa đủ điều kiện), **30% rơi** thành viên Tiên khí quanh nạn nhân.
- Minimap hiện 5 đài, nhấp nháy khi sắp phun lớn.

**Lưu ý**: hệ sát mạng hiện tại (`server/internal/game/tien.go` + `tien_test.go`) phải xoá,
không phải sửa. Config `tien_kills_per_step`, `tien_bot_kill_value`,
`tien_same_victim_cooldown_secs` thay bằng mốc Tiên khí và thông số đài.

## 3. Viết lại mapgen theo 9 bước §8.1 (CHANGES_v4 #11)

Hiện mới **sửa 3 lỗi** (lề 300 + validate, tường biên, dãy tường nối liền) trên thuật toán cũ.
Chưa làm thuật toán mới:

1. Thiên Tiên Đài trung tâm · 2. 4 sân môn phái có Đài Tiên Khí · 3. đường chính + vành đai
· 4. đối xứng xoay 4 phía · 5. 4 biome theo Voronoi nhiễu · 6. Poisson-disk + cụm
· 7. cổng nối chéo · 8. kiểm tra liên thông · 9. kiểm tra hợp lệ (đã có).

Thêm `cmd/mapgen --seed N --png out.png` xuất ảnh xem trước. Art mới (đường đá, đài) dùng
placeholder trước.

**Phụ thuộc**: bước 1–2 cần biết Đài Tiên Khí là gì → nên làm **sau** việc 2.
