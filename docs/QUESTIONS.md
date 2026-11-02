# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

Bối cảnh chung: lượt cuối 01/11/2026; cập nhật 02/11: Q37–Q42 đã xong (spec v6.35). Chi tiết từng dòng: `docs/SPEC_AUDIT_v630.md`.
Mỗi câu ghi **phương án đang dùng** (an toàn, không đổi luật) — anh chỉ cần trả lời nếu muốn khác.

## Q44 — §12.5 A "một thân nhiều mặt": art thử KHÔNG dùng được, chọn hướng đi tiếp

Ngày 30/10 anh chốt: **sinh art cho 3 nhân vật để thử, CHƯA code phần A** — để biết lớp mặt
có khớp vào thân không rồi mới quyết (ghi trong `server/cmd/artgen/bodyface.go`). Art thử đã
có cho `a_hon`, `su_phu_mi`, `dai_ca_dep`. Ngày 02/11 tôi làm nốt bước đánh giá đó.

**Kết quả: không khớp.** Ba chỗ hỏng, nhìn bằng mắt là thấy
(`docs/review/q44_than_vs_mat.png` — vùng đầu của khung chạy đặt cạnh sprite mặt; cả bộ art
thử ở `docs/review/q44_bo_art_thu.png`):

1. **Thân đã có sẵn mặt đầy đủ** — mắt, lông mày, cả giọt mồ hôi. Câu `faceNeutral` trong
   prompt ("face area left plain and neutral") KHÔNG ăn. Chồng lớp mặt lên sẽ thành **hai
   cái mặt đè nhau** — đúng điều ghi chú trong code đã cảnh báo.
2. **Khác hẳn nhau về hình**: thân đội mũ bánh bao trơn, sprite mặt đội mũ có vành đỏ; nét
   vẽ và cách tô cũng khác. Hai bộ trông như hai nhân vật.
3. **Không chung neo, không chung tỉ lệ**: thân là góc NGHIÊNG, vùng đầu 217×120 px trong
   canvas 512; sprite mặt là góc CHÍNH DIỆN, 191×230 px và **có cả cổ với vai**. Spec đòi
   "cùng kích thước, cùng vị trí neo" — hiện không có cái nào.

Còn một chỗ thiếu ở đường ống: **`prep` chỉ neo CHÂN** (`pivot: feet`), không có neo đầu.
Muốn dán mặt thì phải biết hộp đầu ở đâu trong từng khung thân — dữ liệu đó chưa tồn tại.

**Tôi KHÔNG sinh art cho 7 nhân vật còn lại.** Lý do: theo đúng ý anh lúc chốt chia giai
đoạn — sinh tiếp bây giờ là tốn 14 lượt quota cho art chắc chắn không ghép được. Bảy nhân
vật kia hiện **chưa có job nào** (`bodyFaceChars` chỉ liệt 3 tên), nên không có gì chạy nhầm.

### Ba hướng, anh chọn

| | Hướng | Việc phải làm | Rủi ro |
|---|---|---|---|
| **A1** | Sửa prompt rồi thử lại trên 1 nhân vật | Ép thân vẽ đầu TRỐNG (giống cách đã phải ép chặt cho prompt mặt), vẽ mặt chỉ-đầu đúng góc nghiêng, thêm bước dò hộp đầu vào `prep` | Mô hình vẫn hay tự vẽ mặt vào thân; có thể tốn vài lượt thử nữa |
| **A2** | Bỏ hướng A, giữ nguyên mỗi mood một bộ frame | Không làm gì. Xoá 2 job + 3 bộ art thử khỏi danh mục | Mất mục tiêu "đổi mood không đổi dáng người" của §12.5 A |
| **A3** | Nhảy thẳng sang rig cắt mảnh 2D | Chính spec đã ghi là nâng cấp sau nếu A–C chưa đủ: cắt đầu/thân/tay/chân từ MỘT ảnh chuẩn, animation bằng code | Việc lớn nhất trong ba, nhưng giải luôn cả vấn đề frame lệch nhau |

**Phương án đang dùng:** A2 trên thực tế — client không đọc `faces/`, game chạy như cũ, không
có gì hỏng. Ba bộ art thử nằm im trong `assets/src`, không tệp nào nạp.

Tôi nghiêng về **A1 thử đúng một nhân vật nữa** rồi mới quyết, vì nó rẻ nhất và trả lời dứt
điểm được câu "cách sinh này có cứu được không". Nhưng đây là chuyện hình ảnh của game nên
anh quyết.
