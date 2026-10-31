# Việc cần làm — spec v6.5 (kỹ năng theo phạm vi người dùng)

Đọc GAME_SPEC §6.3 (đã viết lại phần phạm vi + bảng kỹ năng). Config: thêm `skills.max_charges_tien`, `skills.range_unit`, `skills.ranges_u`, `skills.push_secs`.

1. Định nghĩa `U = player.pickup_r × scale_người_dùng` (1 hàm dùng chung server + client). Thay **mọi** số phạm vi cố định trong code kỹ năng bằng `ranges_u[...] × U`: Loa Phường nón/đẩy 6 U, Ru Ngủ 5 U, Đàn Em bám 3 U / tìm 8 U, Phun Trân Châu 8 U, Gánh Lướt 7 U (vệt rộng 1 U), Lưới Mì ném 4.5 U / bán kính 3 U, Rụt Cổ sóng 2 U / đẩy 3 U.
2. Khoảng đẩy lùi tính theo U của **người dùng**, không theo người bị đẩy. Đẩy trong 0.3s, trượt dọc vật cản, không xuyên.
3. VFX/chỉ báo phạm vi (nón, vòng tròn) vẽ theo cùng giá trị → nhìn thấy đúng vùng tác dụng.
4. Bí Kíp: Tiên giới giữ tối đa 3 lượt; rớt về Đại Thừa thì cắt còn 2. Nút kỹ năng hiện 3 chấm lượt ở Tiên giới.
5. Test theo checklist §17 (3 dòng mới) + grep đảm bảo không còn hằng số phạm vi trong code kỹ năng.
