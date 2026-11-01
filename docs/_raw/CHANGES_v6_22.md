# Việc cần làm — GAME_SPEC v6.22 + ADMIN_SPEC v1.3 (bot + phòng 50)

Làm sau CHANGES_v6_21 (lạm phát EXP), trước 3b.

1. Bot sinh như người chơi mới: Luyện Khí 1, 1 kiếm, không buff/Bí Kíp, 5s bất tử. Bỏ `spawn_rank`, `max_spawn_realm`. Bot vẫn tu luyện trong trận.
2. Chế độ bot `auto | on | off`: toàn server + ghi đè từng phòng. Auto: > 40 người thật → ngừng sinh, < 35 → sinh lại. Tắt = không sinh mới + bot chết không hồi sinh. Nút "Rút bot ngay": bot không giao tranh trong 5s rời phòng (hiệu ứng khói), bot đang đánh rút sau.
3. Trang admin (Phòng & Khu): công tắc toàn server, chế độ từng phòng kèm lý do, nút rút bot, cột người thật/bot/giới hạn 50. Modal.confirm + audit.
4. `max_humans` 20 → 50. TRƯỚC khi deploy giới hạn 50: loadtest 50 người thật + 20 bot rải 5 khu, ghi tick p95 từng khu + KB/s vào PERF_LOG. Không đạt < 4 ms thì giữ 20 và báo tôi.
5. Test theo 3 dòng mới ở checklist §17. Deploy + báo.
