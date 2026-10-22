# Câu hỏi đang chờ anh trả lời

Mỗi câu: bối cảnh → phương án → **tạm dùng** (đã code theo phương án này, đổi lại dễ).
Câu đã xong ở `docs/QUESTIONS_DONE.md`.

---

## Q15. Công thức lá chắn §5.2 làm không ai chém chết được ai

**Bối cảnh:** §5.2 ghi ràng buộc đối xứng:

> `dist(A, B) ≥ max(reachA + shield_rB, reachB + shield_rA)`

Code đúng y công thức này thì **không ai hạ được ai**. Lý do: người có từ 2 vòng trở lên có
`shield_r` (mũi kiếm vòng kế ngoài) **lớn hơn chính tầm với của mình trừ đi 1 lớp**, nên vế
`reachB + shield_rA` đẩy nạn nhân ra xa hơn tầm kiếm của A.

Ví dụ đo thật (Kim Đan, 20 kiếm = 2 vòng): `reachA` ≈ 245, `shield_rA` ≈ 222. Nạn nhân 1 vòng
(`reach` ≈ 117, không còn lá chắn) bị giữ ở `117 + 222 = 339` — xa hơn 245 nên A vĩnh viễn
không chạm tới thân B. Trận đấu không bao giờ kết thúc được bằng §5.1.

**Tạm dùng — hai điều chỉnh, đã ghi chú trong `shield.go`:**
1. **Chỉ áp theo hướng người đang đi**: `dist ≥ reach(người đi) + shield(người kia)`. Giữ đúng
   ý "chỉ người đang đi bị chặn như va tường", và vẫn bịt kín các vòng trong.
2. **Còn 1 vòng thì `shield_r` = 0** (spec ghi `BODY_R × scale`). Với `BODY_R × scale`, mũi
   kiếm chỉ chạm **tiếp tuyến** mặt ngoài thân nên §5.1 bắn được hay không tuỳ sai số số học.
   Để 0 thì "còn 1 vòng mới chạm được tới thân" xảy ra dứt khoát.

Nếu anh muốn giữ nguyên công thức đối xứng, cần nói rõ **làm sao kết thúc được trận** — ví dụ
lá chắn chỉ chặn thân chứ không chặn kiếm, hoặc `shield_r` nhỏ hơn `reach` một lớp.
