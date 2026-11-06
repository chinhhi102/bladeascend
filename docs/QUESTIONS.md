# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

Bối cảnh chung: lượt cuối 01/11/2026; cập nhật 02/11: Q37–Q42 đã xong (spec v6.35). Chi tiết từng dòng: `docs/SPEC_AUDIT_v630.md`.
Mỗi câu ghi **phương án đang dùng** (an toàn, không đổi luật) — anh chỉ cần trả lời nếu muốn khác.

## Q46 — Nam Châm +50% làm 4 cảnh giới thấp trượt luật "≤ 1,5 lần nhanh hơn" (§6.0, §17)

**Bối cảnh**: triển khai spec v6.40. Bán kính nhặt đổi sang `90 × scale × (1 + Σ%)`, Nam Châm
+0.2 → +0.5. Đã chỉnh `orb_value_mult` 1.5 → **1.32** theo đúng yêu cầu của §6.0 luật 2:
`cmd/sim --farm` trúng bảng nhịp cả 9 cảnh giới, hai lượt liên tiếp (cộng dồn 27,3 và 27,6
phút so với mốc ~28).

**Chỗ không thoả được**: `cmd/sim --farm --magnet-always` canh "có Nam Châm không được
nhanh hơn bảng quá 1,5 lần". Kết quả:

| Cảnh giới | Có Nam Châm (phút) | Sàn ≥ | Nhanh hơn bảng |
|---|---|---|---|
| Luyện Khí | 0,84 | 0,87 | 1,55× ❌ |
| Trúc Cơ | 0,86 | 1,00 | 1,74× ❌ |
| Kim Đan | 1,09 | 1,20 | 1,65× ❌ |
| Nguyên Anh | 1,31 | 1,40 | 1,60× ❌ |
| Luyện Hư | 2,10 | 2,20 | 1,57× ❌ |
| 4 cảnh giới còn lại | | | ✅ |

Bài "Luyện Hư 1 → Chân Tiên có Nam Châm suốt ≥ 15 phút" vẫn ✅ (19,4 phút).

**Vì sao không chữa được bằng `orb_value_mult`**: hệ số đó vào theo luỹ thừa cảnh giới
(`m^r`), nên ở Luyện Khí `m^0 = 1` — chỉnh bao nhiêu cũng không đổi cảnh giới 1. Mà chính 4
cảnh giới thấp mới là chỗ trượt. Trước v6.40 bài này ✅ thoải mái (Luyện Khí 1,15 so với sàn
0,87) vì Nam Châm chỉ +0.2.

**Phương án đang dùng**: giữ nguyên spec — Nam Châm +0.5, `orb_value_mult` 1.32, và để bài
`--magnet-always` đỏ ở 4 cảnh giới thấp. Không tự hạ Nam Châm xuống dưới con số spec ghi.

**Ba hướng chọn** (cần anh chốt):
1. **Hạ Nam Châm** còn +0.3 → +0.35: bài sẽ xanh lại, nhưng trái con số +50% spec vừa chốt.
2. **Nới luật ≤1,5×** cho 4 cảnh giới đầu (vd ≤1,8× dưới Hóa Thần): Nam Châm ở cảnh giới
   thấp vốn là phần thưởng lớn nhất, có thể đó là ý đồ.
3. **Nâng mốc bảng nhịp** của 4 cảnh giới thấp (Luyện Khí 1,3 → 1,6…): đổi bảng nhịp, mà
   §6.0 ghi rõ "không đổi bảng nhịp".
