# Việc cần làm — spec v3.3 (bảng xếp hạng Uy Danh + nerf Ẩn Thân)

Đọc docs/GAME_SPEC.md §11 (viết lại toàn bộ) và dòng Ẩn Thân Phù ở §6.2. Cập nhật config/game.json: thêm khối `score`, sửa `buffs.stealth`.

**Bảng xếp hạng**
1. Server tính điểm Uy Danh theo bảng §11.1 (chém trúng người, lật kèo, chém chết, đột phá lần đầu, giữ Top 1 phòng mỗi giây). Bot làm nạn nhân → 50%. Chém lại cùng guest_id trong 60s → 0 điểm.
2. Thêm event `ScoreGain {amount, reason}` để client hiện số bay lên ("+800 LẬT KÈO!"). HUD hiển thị Uy Danh hiện tại.
3. Bảng trong phòng: xếp theo cảnh giới + tầng, hiện kèm Uy Danh; vương miện trên đầu người Top 1.
4. SQLite: migration thêm cột `score, hits, secs_to_dai_thua, top1_secs` (bỏ xếp hạng theo peak_exp). API `board=score|hits|speed`. Bảng speed chỉ lấy dòng có `secs_to_dai_thua`, xếp tăng dần.
5. Trang leaderboard: 3 tab Uy Danh / Sát Thần / Thần Tốc × 3 khoảng thời gian. Màn hình Tổng kết hiện chi tiết điểm theo nguồn.
6. Dữ liệu cũ xếp theo EXP: xóa hoặc giữ nhưng không hiển thị.

**Ẩn Thân Phù**
7. Thời hạn 5s, nhặt lại chỉ làm mới về 5s (không cộng dồn quá 5s).
8. Không lọc khỏi gói tin nữa; gửi kèm cờ `stealth`. Client vẽ người đó + vòng kiếm với alpha 0.2, ẩn tên, ẩn minimap; trong `reveal_r × scale người nhìn` thì alpha 1.0. Chính chủ thấy mình mờ 40%.
9. Bot: tầm phát hiện người Ẩn Thân × 0.5.
10. Unit test cho công thức điểm (các ví dụ trong §11.1), chống cày 60s, trần 5s của Ẩn Thân.
