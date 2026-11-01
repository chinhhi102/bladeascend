# Việc cần làm — GAME_SPEC v6.26 (chém trúng = chết)

Làm cùng nhóm "sửa lỗi ảnh hưởng chơi" (trước giai đoạn 2 của lượt lớn).

1. Chém trúng thân (đủ điều kiện §5.2: nạn nhân ≤ 1 vòng, không khiên, không bất tử) → **chết ở mọi cảnh giới**. Bỏ luật rớt đại cảnh giới, bỏ 1s bất tử sau khi bị chém (trừ khi khiên đỡ).
2. Thưởng người chém: 30% `total_exp` nạn nhân, trần = `sàn(r_A + 2) − A.total_exp`; nạn nhân bot → 15%. Rơi tại xác 30% thành viên Linh khí (≤ 40 viên, mệnh giá lớn trước) + 50% kiếm thành Bó Kiếm. 40% còn lại mất.
3. Tiên giới: mất hết Tiên khí; người chém (hấp thụ được) nhận 30%, trần +1 cảnh giới Tiên; luôn rơi 30% thành viên Tiên khí.
4. Uy Danh: mọi lần chém trúng = chém chết (100 × (r+1) × lật kèo + 50). Bảng Sát Thần đếm số người chém chết.
5. Dọn code/anim/kill feed liên quan "rớt cảnh giới" do bị chém (giữ anim `realm_drop` trong code nhưng không còn gọi từ combat; art giữ nguyên).
6. Cập nhật cmd/sim (farm + duel) và bảng nhịp §6.0 nếu cần: chết là mất hết nên đo lại "thời gian sống trung bình" và "cảnh giới cao nhất trung bình mỗi mạng" — gửi tôi số trước/sau.
7. Test theo 3 dòng mới ở checklist §17; xoá các test cũ về rớt cảnh giới.
