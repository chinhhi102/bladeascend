# Việc cần làm — spec v6.6 (Tiên khí ×10)

1. `cultivation.tien_khi_ceilings`: [2000, 6000, 12000, 20000, 30000] → **[1000, 3000, 6000, 10000, 15000]**. Đáy: 1 / 1 001 / 3 001 / 6 001 / 10 001. Đạo Tổ đổi Uy Danh khi đầy 15 000.
2. `player.burn_tien_frac_per_sec`: 0.004 → **0.008** (Thiêu Đốt Tiên giới vẫn đúng 8/16/24/32/40 mỗi giây).
3. Giữ mệnh giá viên rơi 500/100/10/2/1, tối đa 40 viên.
4. Sửa test theo checklist §17: Thái Ất 4 500 bị chém → Kim Tiên 1 001, người chém +1 749, rơi 1 049; Chân Tiên 800 bị chém → rơi 240; Đạo Tổ trần 15 000.
5. Chạy lại ước tính thời gian lên Đạo Tổ khi ôm 2 đài (mục tiêu ≈ 30 phút), ghi PERF_LOG/PROGRESS.
