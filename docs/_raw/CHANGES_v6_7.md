# Việc cần làm — spec v6.7 (vật cản theo bậc cảnh giới + sắp xếp bản đồ)

Đọc GAME_SPEC §8.0 (hướng sửa từng khu), §8.1b (luật sắp xếp + kiểm tra), §8.2 (bậc vật cản), §5.3. Ảnh `map_check_scale5.png` là kết quả kiểm 5 bản đồ hiện tại với người Đại Thừa (scale 5); `map_check_reference.py` là script Python tôi dùng để kiểm — port sang Go trong `cmd/mapgen --check`.

## A. Luật vật cản (server + client)
1. Bỏ HP, damage, trạng thái nứt/vỡ, đống đổ nát. Config: xoá `obstacles.hp`, `dmg_cooldown_secs`, `dmg_per_extra_ring`; thêm `min_realm`, `corridor_min`, `hard_buff_extra_tier`, `cut_hint_cooldown_secs`.
2. Kiếm của người có `realm ≥ min_realm` (Cương Kiếm: +1 bậc; Tiên giới: mọi bậc) chạm vật cản → xoá ngay, không anim, 1 SFX "bụp". Chưa đủ → không tác dụng; tia lửa + chữ "Cần {cảnh giới}" (≤ 1 lần/3s/vật cản); vật cản cao hơn bậc của mình có viền tối khi trong tầm kiếm.
3. Mọc lại 30s, nhấp nháy 3s trước, chờ trống/đẩy mượt như cũ. Sạp/xe vẫn 30% rơi item.
4. **Biên khu**: tách thành loại riêng `border` (không có min_realm, không bao giờ chém được). Hiện biên đang là 4 `stone_wall` 11 400 dài — đổi type, đừng để chúng nhận bậc 4.
5. Bot: vật cản mình chém được = đi qua được (+0.3s chi phí).

## B. Sắp xếp bản đồ
6. `cmd/mapgen --check`: với r = 0..13, thân to nhất của r, bỏ vật cản bậc ≤ r, flood fill lưới 50 → liên thông, mọi cổng/Thiên Môn/Đài/bến/spawn tới được, trống ≥ 60%; xuất `docs/map_check/{khu}_r{r}.png`. Đưa vào CI.
7. Sinh lại 5 khu theo §8.1b: 3 cấp đường (chính ≥ 600 không vật cản, phụ ≥ 300), vật cản thành đảo ≤ 1 200², không tường liền > 800, đảo phân tầng (viền bậc thấp, lõi bậc cao), trống ≥ 600 + ≥ 3 lối ≥ 400 quanh cổng/Đài/Thiên Môn/bến, mật độ 18–28%, đối xứng; khe giữa vật cản ≥ `corridor_min[min(bậc)]`.
8. Áp hướng sửa riêng từng khu trong §8.0 (Thiên Đình mở sân môn phái; Linh Sơn ≥ 3 đèo/dải vách ≥ 300; Trúc Lâm hành lang ≥ 220 + tre 3 vòng bậc 1/2/3; Phàm Trấn nới ngõ ≥ 160; Băng Nguyệt Hồ thêm 4–6 đảo băng có chỗ nấp).
9. Gửi tôi ảnh `--check` của 5 khu ở r = 0, 4, 8 (trước/sau) + ảnh xem trước `--png` từng khu để duyệt bố cục trước khi deploy.
