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

## Q48 — Nhảy cóc khi nhìn người chạy nhanh từ tab khác (CÒN MỞ — bước 1 xong, chưa đạt)

**Chốt của chủ dự án 07/11**: làm theo bước, đo lại sau mỗi bước, đạt (0 nhảy cóc, 0 đứng
hình) thì dừng. Không đổi `snapshot_hz`.

**Bước 1 — `interp_ms` 100 → 150: ĐÃ LÀM, có tác dụng rõ, nhưng CHƯA ĐẠT.**

| Bản | Khung nhảy cóc / 1 200 | Đứng hình |
|---|---|---|
| `interp_ms` 100 (nền), 2 lượt | 32 và 46 (**2,67% · 3,83%**) | 32 · 24 |
| **`interp_ms` 150, 4 lượt** | 3–21 (**0,25% – 2,23%**) | 4–59 |

Giảm 5–10 lần nhưng không về 0. Đang chạy với 150.

**Bước 2 — ĐÃ THỬ HAI BẢN, CẢ HAI TỆ HƠN, ĐÃ HOÀN TÁC.**

Viết đúng như anh mô tả: kho cạn thì ngoại suy theo vận tốc cuối tối đa 100 ms rồi đứng
chờ; gói mới tới mà lệch thì kéo mượt về trong 100 ms, lệch > 300 đv mới nhảy thẳng; áp cho
người khác + quái, không áp cho mình.

- **Bản 1**: 16,5% khung nhảy cóc (nền 2,7%), tốc vẽ trung vị **1 281 đv/s vượt trần 975**.
  Nguyên nhân: gieo khoản lệch mỗi lần có gói mới (20 lần/giây), khoản cũ chưa tiêu hết đã
  bị thay → entity đứng lệch vĩnh viễn và chạy đuổi.
- **Bản 2**: chỉ gieo khi không còn khoản nào đang tiêu và lệch > 2 đv → entity **đứng im
  hẳn** trên màn người xem.

Tôi hoàn tác cả hai thay vì đẩy netcode nửa vời lên `:9999`.

**Cần anh quyết hai việc:**

1. **Làm tiếp bước 2 thế nào?** Đề nghị: viết **bài đo đơn vị cho riêng `followSnaps`**
   trước (bơm một chuỗi gói giả có lỗ hổng, canh quỹ đạo ra), rồi mới ráp vào client. Hai
   lần vừa rồi tôi thử thẳng trên trình duyệt nên mỗi vòng mất ~10 phút và không nhìn thấy
   trạng thái bên trong. Hoặc anh muốn dừng ở bước 1 và chấp nhận 0,25–2,2%?
2. **Chưa đưa `gomdon-smooth.mjs` vào cổng trước deploy** như anh dặn, vì hai lẽ: ngưỡng
   "0 nhảy cóc" sẽ chặn **mọi** lần deploy kể từ hôm nay (hiện 0,25–2,2%), và bài này cần
   một server dev riêng + trình duyệt nên không chạy chung được với các `check:*` thuần
   Node. Chạy tay: `cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998
   -debug=true` rồi `node tools/browsertest/gomdon-smooth.mjs http://127.0.0.1:9998`.
   Anh muốn tôi đặt ngưỡng tạm theo mức hiện tại (vd ≤ 3%) để nó bắt được thoái lui, hay
   chờ bước 2 xong mới gắn?

