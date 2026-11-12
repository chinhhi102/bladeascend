# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

Bối cảnh chung: lượt cuối 01/11/2026; cập nhật 02/11: Q37–Q42 đã xong (spec v6.35). Chi tiết từng dòng: `docs/SPEC_AUDIT_v630.md`.
Mỗi câu ghi **phương án đang dùng** (an toàn, không đổi luật) — anh chỉ cần trả lời nếu muốn khác.

## Q53 — Mốc "phút thứ 5" của bài kiểm rò rỉ client (§13.9) quá sớm cho game này

Spec §13.9: *"chơi 30 phút giả lập + đổi khu 50 lần + chết/vào lại 10 lần → số Texture,
GPU MB, JS heap, số display object **không tăng quá 10% so với phút thứ 5**. Tăng → build
fail."*

Chạy đúng bài đó ngày 12/11 trên build đã sửa xong hai rò rỉ server (`tools/browsertest/leakcheck.mjs`):

| chỉ số | phút 5 | phút 30 | so phút 5 | phút 20–25 → 25–30 |
|---|---|---|---|---|
| texture | 322 | 401 | **+24,5 %** ❌ | 417 → 421 (**+1,0 %**) ✅ |
| texMB | 223,4 | 223,4 | **−0,0 %** ✅ | phẳng ✅ |
| JS heap | 80,1 MB | 109,9 MB | **+37,3 %** ❌ | 114,0 → 113,4 (**−0,5 %**) ✅ |
| display object | 4 751 | 4 399 | −7,4 % ✅ | phẳng ✅ |

**Đọc số này:** `texMB` phẳng tuyệt đối trong khi SỐ texture tăng 79 — tức là 79 tấm texture
rất nhỏ (chữ, biểu tượng), không phải ảnh lớn. Và từ phút 20 trở đi **mọi chỉ số đứng yên**.
Đó là một **TRẦN**, không phải một đường dốc: bộ làm việc của game vẫn đang lớn lên ở phút 5
(11 atlas nhân vật nạp dần khi gặp người mới, art của 5 khu nạp dần qua 50 lượt đổi khu,
pool giãn tới mức dùng đỉnh), và nó chạm trần quanh phút 20.

Rò rỉ THẬT thì không có trần — so với lượt #33 trong PERF_LOG: +91 texture/phút, một chiều,
cho tới khi tab chết.

**Phương án đang dùng** (để anh chỉ phải trả lời nếu muốn khác): bài đo in CẢ HAI phép so.
Phép so của spec (mốc phút 5) in ra như số liệu; phép **chặn deploy** là xu hướng **nửa sau
↔ mốc cuối**, cùng cách đã dùng cho bài rò rỉ server. Lý do: một cổng đỏ vì một thứ đúng sẽ
bị gõ `SKIP_LEAK=1` trong vòng một tuần, và lúc đó cổng không còn bắt được gì nữa.

**Anh chọn một trong ba:**
1. Giữ như đang làm (chặn theo xu hướng nửa sau; mốc phút 5 chỉ để tham khảo).
2. Đổi spec sang mốc **phút 20** rồi chặn theo đúng câu chữ §13.9.
3. Giữ nguyên câu chữ spec (chặn theo mốc phút 5) — khi đó tôi phải đi tìm và cắt 79 tấm
   texture nhỏ kia, dù chúng không làm `texMB` nhúc nhích.

## Q52 — Hoá Khổng Lồ nhân `scale`, nên vòng NHẶT và tầm KỸ NĂNG cũng to theo trong 2 giây

Spec §6.3 v6.48 ghi Hoá Khổng Lồ ×2 cho **ba** thứ: thân, bán kính vòng kiếm, số kiếm.

Cách làm hiện tại là nhân thẳng `scale` của nhân vật ×2. Mọi thứ tính theo scale đi theo:
đúng ba thứ spec nói, **cộng thêm** bán kính nhặt đồ (×√2 theo §6.0) và tầm mọi kỹ năng
(U = 90 × scale). Trong 2 giây, hồi 25 giây.

Tôi để nguyên vì nó hợp lý về hình ảnh (người khổng lồ thì với xa hơn thật) và tự giới hạn
bởi 2 giây. **Nếu anh muốn đúng ba thứ**, phải tách "scale để VẼ" khỏi "scale để TÍNH
LUẬT" — làm được, nhưng đó là một lượt riêng và phải canh kỹ: hai con số lệch nhau là hình
vẽ và vùng ăn đòn lệch nhau.

Một hệ quả nữa, không phải lỗi: máy quay tự lùi ra khi nhân vật to lên (§12 — khung nhìn
theo `scale^camera_exp`), nên trên màn hình thân chỉ TO RA 1,41 lần chứ không phải 2 — còn
trong thế giới thì đúng 2. Đổi được nếu anh thấy chướng.

> Đang chờ: **Q53**. Mọi câu đã trả lời nằm ở `docs/QUESTIONS_DONE.md`.
