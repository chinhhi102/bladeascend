# Việc cần làm — spec v6.56

Spec đã cập nhật: `docs/GAME_SPEC.md` §8.4a, §6.3, §16, §17. Dán file mới vào `docs/_raw/` rồi làm 2 việc dưới.

## 1. Mọi cổng đều hai chiều, đích ngẫu nhiên (§8.4a)

Bỏ hoàn toàn luật "Thiên Môn một chiều, đích cố định".

1. Thiên Môn (1 cái ở tâm mỗi khu, trừ Thiên Đình) trở thành **cổng thứ 5 của khu**, cùng một luật với 4 cổng thường: đích ngẫu nhiên trong mạng cổng toàn thế giới (`exit_pool: "all"`), không ra lại chính cổng vừa vào.
2. 4 **Bến Thiên Môn** ở Thiên Đình trở thành **cổng hai chiều bình thường**: vào được, ra được, đích ngẫu nhiên. Không còn khu nào "thuộc" bến nào.
3. Config: xóa `gates.heaven_gate.map`, thêm `one_way: false` và `exit_pool: "all"`. Vị trí/kích thước giữ nguyên (`radius` 200, `landing_r_from_center` 1400, `landing_clear_r` 400, `zone_center_plaza_r` 500).
4. Giữ nguyên phần nhận diện: Thiên Môn to gấp đôi + cột sáng vàng, icon minimap riêng + mũi tên chỉ hướng, kill feed khi Top 3 qua Thiên Môn. Nó vẫn là mốc dễ nhìn, chỉ không còn là "đường chắc chắn tới Thiên Đình".
5. Bot: bỏ logic "đi Thiên Môn để lên Thiên Đình" (vì không còn bảo đảm tới đó) — bot dùng cổng gần nhất như cổng thường.
6. Test: 100 lượt qua Thiên Môn và qua 4 cổng Thiên Đình → đích rải đủ cả 5 khu, không lần nào ra lại cổng vừa vào, không lần nào kẹt trong vật cản. Xóa/sửa test cũ assert "Thiên Môn luôn ra Thiên Đình".

**Lưu ý thiết kế**: sau thay đổi này không còn lối đi bảo đảm tới Thiên Đình nữa — ai muốn lên đó phải đi cổng tới khi gặp. Nếu tỉ lệ người lên Thiên Đình tụt mạnh thì báo số liệu, đừng tự thêm cơ chế bù.

## 2. Sư Phụ Mì Gói: 2 lượt tích (§6.3)

7. `skills.cooldowns.su_phu_mi`: `[35, 1]` → `[35, 2]` (hồi 35s, tích tối đa 2 lượt). Chỉ đổi số lượt, hiệu ứng kỹ năng giữ nguyên (trói 5s, cướp kiếm).
8. Bảng chọn nhân vật + tooltip HUD phải hiện đúng "35s · 2 lượt".
9. Cập nhật test cân bằng kỹ năng (mốc so sánh giờ là "Sư Phụ 35s × 2").

Xong: build + test pass, ghi `PROGRESS.md`.
