# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

Bối cảnh chung: lượt cuối 01/11/2026; cập nhật 02/11: Q37–Q42 đã xong (spec v6.35). Chi tiết từng dòng: `docs/SPEC_AUDIT_v630.md`.
Mỗi câu ghi **phương án đang dùng** (an toàn, không đổi luật) — anh chỉ cần trả lời nếu muốn khác.

## Q50 — Hút Mì: §6.3 nói BÁN KÍNH 5 U, §6.5 và bảng tham khảo vẫn ghi NÓN (đang dùng: bán kính)

Spec v6.48 §6.3 (bảng kỹ năng) viết: *"mọi địch trong bán kính 5 U bị sợi mì nối vào"* —
tức ăn 360°. Nhưng hai chỗ khác trong cùng file vẫn là hình nón:

- bảng "đã nhân U": `Hút Mì: nón (5 U)`
- §6.5 lớp ②: `Nón 60° × 5 U loé`

**Đang dùng: BÁN KÍNH** — bảng kỹ năng §6.3 là chỗ ghi luật, hai chỗ kia là minh hoạ. Nếu
anh muốn giữ nón thì nói, tôi đổi lại trong vài phút (khoá `hut_mi_r` → `hut_mi_cone` +
`hut_mi_cone_deg`, và bot quay lại canh hướng).

Tương tự, nhỏ hơn: **Trân Châu** — bảng §6.3 v6.48 ghi *mù 3s*, còn khối "Trân châu dính
màn hình" bên dưới vẫn ghi 5s ở cả ba chỗ (hiệu ứng rơi/tan, không nhặt được, bot tầm nhìn).
Đang dùng **3 s cho LUẬT** (bảng §6.3), phần vẽ trên màn hình vẫn tan trong 5 s như cũ.

## Q51 — Cô Ba còn rải vỏ chuối nữa không? (đang GIỮ)

Bảng kỹ năng §6.3 v6.48 tả Cô Ba là "sạp bất hoại + đẩy + buộc vào sạp" — **không nhắc
vỏ chuối**. Nhưng §6.5 (bảng hiển thị) vẫn ghi: *"Sạp có tấm bạt cam + thanh đếm 6s nhỏ;
**vỏ chuối vàng dọc chân sạp**; người trượt: vệt trượt + sao choáng 1s"*.

**Đang GIỮ vỏ chuối** (trượt + choáng 1 s như cũ), vì bỏ một cơ chế mà spec chưa bảo bỏ
thì rủi ro hơn giữ. Nhưng cộng lại thì Cô Ba giờ khá nặng: chặn đường bất hoại + đẩy văng
+ buộc 3 s + vỏ chuối (trượt + choáng 1 s).

**Cần anh quyết**: (a) giữ nguyên cả bốn; (b) bỏ vỏ chuối (và sửa §6.5 cho khớp); (c) giữ
vỏ chuối nhưng bỏ phần choáng, chỉ còn trượt.

