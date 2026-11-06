# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

Bối cảnh chung: lượt cuối 01/11/2026; cập nhật 02/11: Q37–Q42 đã xong (spec v6.35). Chi tiết từng dòng: `docs/SPEC_AUDIT_v630.md`.
Mỗi câu ghi **phương án đang dùng** (an toàn, không đổi luật) — anh chỉ cần trả lời nếu muốn khác.

## Q47 — Spec v6.41 ghi Nam Châm ≤ 1,75× nhưng chạy 6 lượt đo được tới 1,86×

**Bối cảnh**: Q46 chốt "chấp nhận khoảng đo được", spec v6.41 ghi con số **1,75×** vào §6.0
và §17. Con số đó lấy từ bảng tôi gửi hôm trước — bảng ấy chỉ có **2–3 lượt** mỗi cảnh giới.
Chạy lại **6 lượt** (seed khác nhau) thì biên rộng hơn hẳn:

| Cảnh giới | min | **max** | trung bình |
|---|---:|---:|---:|
| Luyện Khí | 1,44 | 1,59 | 1,53 |
| Trúc Cơ | 1,61 | **1,83** | 1,71 |
| Kim Đan | 1,68 | **1,86** | 1,77 |
| Nguyên Anh | 1,53 | 1,69 | 1,61 |
| Hóa Thần | 1,48 | 1,67 | 1,61 |
| Luyện Hư | 1,52 | 1,68 | 1,60 |
| Hợp Thể | 1,41 | 1,58 | 1,50 |
| Độ Kiếp | 1,30 | 1,44 | 1,38 |
| Đại Thừa | 1,27 | 1,38 | 1,32 |

Đặt ngưỡng 1,75 thì **Trúc Cơ và Kim Đan đỏ** (đã thử: hai lượt liên tiếp đều trượt Kim
Đan ở 1,80× và 1,88×). 1,85 vẫn đỏ Kim Đan. 1,9 xanh cả 6 lượt nhưng chỉ dư 2%. 2,0 xanh
với dư ~7%.

**Phương án đang dùng**: giữ `exp_economy.magnet_max_speedup = 2.0` như Q46 đã chốt hôm
trước — không hạ xuống 1,75, vì làm thế là cố tình để một bài nghiệm thu đỏ nửa số lượt
chạy. Luật chơi KHÔNG đổi: Nam Châm vẫn +0.5, `orb_value_mult` vẫn 1.32. Đây thuần tuý là
ngưỡng của bài đo.

**Cần anh chốt một trong hai**:
1. **Sửa spec 1,75 → 2,0** (§6.0 hai chỗ + khối config §17). Ngưỡng phải có dư cho dao động
   seed: riêng Trúc Cơ biên đã ±7% quanh trung bình.
2. **Giữ 1,75 trong spec** và chấp nhận bài `--magnet-always` đỏ tuỳ seed — nếu vậy nói
   tôi biết để hạ config về 1,75 cho khớp spec.

Nếu muốn ngưỡng chặt mà vẫn ổn định thì hướng thứ ba: ghim seed cố định cho bài đo (bỏ seed
theo giờ) rồi đặt 1,9 — nhưng khi đó bài chỉ còn canh đúng một seed.
