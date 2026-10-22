# Câu hỏi về thiết kế (cần anh/chị xác nhận)

Mỗi câu: bối cảnh → các phương án → **tạm dùng** (đã code theo phương án này, đổi được dễ).

Q1–Q6 đã trả lời trong `docs/ANSWERS.md` và đã áp dụng.

## Q7. Nhịp lên cấp sau Trúc Cơ nhanh hơn mục tiêu §6.1 (spec v3)

**Bối cảnh:** spec v3, `config/game.json` → `exp_orbs`, đo bằng `cd server && go run ./cmd/sim --bots 20 --minutes 8 --farm --orb-cell <N>` (chỉ farm, không ai bị chém).

| `exp_orbs.cell` | Trúc Cơ | Kim Đan | Nguyên Anh |
|---|---|---|---|
| 140 (mặc định spec) | 32s | 47s | 57s |
| 300 (**đang dùng**) | 66–82s | 97–111s | 113–130s |
| Mục tiêu §6.1 | 60–90s | ≈ 3 phút | ≈ 6 phút |

Chỉnh mật độ viên EXP chỉ đạt được mốc Trúc Cơ. Sau đó người chơi to lên: bán kính hút đồ (`90 × scale`), tốc chạy và vòng kiếm (chém quái tức thì) cùng tăng, nên EXP/giây tăng nhanh hơn lượng EXP cần (`× 1.2` mỗi cảnh giới).

**Phương án:**
- **A.** Tăng `realm_mult` (ví dụ 1.2 → 1.6) để mỗi cảnh giới cần nhiều EXP hơn hẳn.
- **B.** Bán kính hút đồ tăng chậm hơn: `90 × sqrt(scale)` thay vì `90 × scale`.
- **C.** Viên EXP môi trường cho ít EXP hơn khi người nhặt ở cảnh giới cao (ví dụ × 1/(r+1)), để EXP cảnh giới cao đến từ quái/người chơi.
- **D.** Giữ nguyên, chấp nhận nhịp nhanh.

**Tạm dùng:** `cell = 300` (đạt mốc Trúc Cơ), các luật khác giữ đúng spec. Chọn A/B/C thì em chỉnh và chạy lại sim.

## Q8. Checklist §17 còn số kiếm mất khi va chạm theo bản cũ

**Bối cảnh:** §5.2 mới ghi mỗi lần va mất `clash_unit` = 5 kiếm (bảng: Trúc Cơ −5 / Luyện Khí −10…), nhưng checklist §17 vẫn ghi "mỗi bên mất 1 kiếm", "Luyện Khí mất 2, Trúc Cơ mất 1", "Luyện Khí mất 4, Nguyên Anh mất 1".

**Tạm dùng:** theo §5.2 (nhân 5), unit test đã viết theo §5.2. Anh/chị sửa giúp checklist §17 nếu đúng vậy.
