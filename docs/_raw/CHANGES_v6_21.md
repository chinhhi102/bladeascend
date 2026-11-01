# Việc cần làm — GAME_SPEC v6.21 (lạm phát EXP) — ƯU TIÊN TRƯỚC 3b

Lỗi người chơi báo: từ Luyện Hư lên Chân Tiên < 20s khi có Nam Châm.

1. **Xác minh nguyên nhân trước khi sửa**: liệt kê MỌI nguồn EXP trong code (viên môi trường, quái, chém người, chém bot, Đài/Tiên khí, buff, phím debug) và MỌI hệ số nhân (theo cảnh giới, scale, buff). Ghi bảng vào docs/EXP_AUDIT.md: nguồn → công thức đang chạy → EXP/giây đo được ở Luyện Khí / Luyện Hư / Đại Thừa (có và không Nam Châm). Tái hiện lỗi bằng `cmd/sim` trước khi đổi gì.
2. Áp luật §6.0: viên EXP và quái × `1.5^r`; bán kính hút `pickup_r × sqrt(scale)`; Nam Châm × 1.5; cướp EXP từ bot 25%; không nguồn nào khác nhân theo cảnh giới. Config `exp_economy` + `buffs.magnet.mult = 1.5`.
3. `cmd/sim --farm` và `--farm --magnet-always`: in bảng thời gian từng cảnh giới, so với bảng §6.0 (±25%). Đưa vào CI. Bài kiểm riêng: Luyện Hư → Chân Tiên farm thuần + Nam Châm ≥ 15 phút.
4. Mạng chơi đạt cảnh giới cao nhờ lỗi này trên bảng xếp hạng: liệt kê (như LEADERBOARD_AUDIT), KHÔNG tự ẩn — tôi duyệt.
5. Deploy, báo bảng trước/sau. Rồi mới quay lại 3b.
