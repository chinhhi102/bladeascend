# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

Bối cảnh chung: lượt cuối 01/11/2026; cập nhật 02/11: Q37–Q42 đã xong (spec v6.35). Chi tiết từng dòng: `docs/SPEC_AUDIT_v630.md`.
Mỗi câu ghi **phương án đang dùng** (an toàn, không đổi luật) — anh chỉ cần trả lời nếu muốn khác.

## Q49 — Bảng xếp hạng 10 dòng vs "không quá nửa màn hình" (đang dùng: THU NHỎ cả bảng)

**Bối cảnh**: v6.47 §11.5 bắt bảng luôn có đúng 3 thẻ Top + 7 ô dòng + ô của mình, kích
thước cố định. Luật cũ (anh chốt 31/10) là bảng không bao giờ cao quá **nửa màn hình**.
Hai luật này không cùng đúng được nếu chỉ có một cỡ chữ: bảng đủ 10 dòng cao ~539 px, mà
nửa màn hình laptop 1366×768 chỉ có 384 px.

Bản cũ xử lý bằng `trimToScreen()` — ẩn dần dòng cuối cho vừa. Đó chính là lỗi anh thấy:
đo 07/11 trên phòng 21 người, bảng chỉ hiện **1–6 dòng** và số dòng đổi theo độ dài tên.

**Phương án đang dùng (an toàn, không đổi luật nào)**: giữ đủ 10 dòng, THU NHỎ cả bảng bằng
`transform: scale(k)` với `k = nửa màn hình / chiều cao tự nhiên`. Hệ số tính từ cỡ màn
hình chứ không từ dữ liệu nên trong lúc chơi nó không đổi.

- 1920×1080: k = 1,00 (bảng 539 px, vừa nguyên cỡ)
- 1280×720: k = 0,67 (bảng hiện cao 360 px)
- 960×540: k = 0,50 (bảng hiện cao 270 px — chữ khá nhỏ)

**Cần anh quyết**: ở màn hình thấp, anh muốn (a) thu nhỏ cả bảng như đang làm, (b) giữ cỡ
chữ và cho bảng cao hơn nửa màn hình, hay (c) rút xuống 5 dòng ở màn hình thấp (bảng vẫn
cố định, chỉ là cố định ở con số khác tuỳ cỡ màn hình)?

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

