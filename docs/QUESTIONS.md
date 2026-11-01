# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

Bối cảnh chung: lượt cuối 01/11/2026. Chi tiết từng dòng: `docs/SPEC_AUDIT_v630.md`.
Mỗi câu ghi **phương án đang dùng** (an toàn, không đổi luật) — anh chỉ cần trả lời nếu muốn khác.

## Q37 — Số liệu config lệch spec (config đã chỉnh theo sim, spec chưa ghi lại)
| Khoá | Spec | Config đang chạy |
|---|---|---|
| `gates.transit_secs` (qua cổng) | 0.8 s | **2.0 s** |
| `minimap.size_mobile` | 120 px | **96 px** |
| `buffs.magnet.mult` | 1.5 | **1.2** |
| `exp_orbs.cell` | 140 | **160** |
| `monster_count` | ~150 / phòng | **150 / khu** (750 / phòng) |
**Đang dùng:** giữ config (đã chạy ổn, có sim). Anh chọn: (a) giữ config và sửa spec, hoặc
(b) đổi config về đúng spec — Claude đổi + chạy lại sim.

## Q38 — Spec còn sót luật "chết rớt cảnh giới" (đã bỏ từ v6.26)
Còn ở §1, §3.2, §5.2, §11.4, §12, §16 (`hit_drops_realm`, `tien_rob_share` 0.5), §17 (#55, #64, #80).
Đoạn giới thiệu §20 (chép nguyên văn ra trang chủ) còn ghi "mười bí kíp riêng" dù kỹ năng giờ tự
hồi, không còn nhặt Bí Kíp. **Đang dùng:** code + trang chủ theo v6.26/v6.29; đoạn §20 giữ
nguyên văn. Anh sửa spec khi tiện.

## Q39 — Góp ý ảnh: "mã hoá lại" hay "gỡ metadata"?
ADMIN_SPEC §7.1 ghi server "mã hoá lại ảnh". Go chuẩn không mã hoá được WebP nếu không thêm
thư viện, nên server đang **gỡ theo cấu trúc tệp** EXIF/GPS/XMP (JPEG, PNG, WebP) — đã test bằng
ảnh có GPS thật. **Đang dùng:** gỡ metadata. Muốn mã hoá lại thật thì cần duyệt thêm thư viện.

## Q40 — ADMIN_SPEC: "không thêm framework UI" vs §2.1 duyệt React + antd
Đã làm theo §2.1 (React 18 + antd 5, bundle riêng, không vào gói người chơi). Đề nghị xoá câu
"không thêm framework UI" cho khỏi mâu thuẫn. Kịch bản PERF_LOG S8 trỏ `livemap.mjs` (luồng 5 Hz
đã bỏ theo §5) — đề nghị bỏ S8.

## Q41 — Font dự phòng cho tên lạ (§19 "vẽ bằng Text thường")
PixiJS 8 tự vẽ thêm ký tự thiếu vào BitmapFont lúc gặp, rơi về font hệ thống (chữ Hán, emoji
hiện đúng). **Đang dùng:** cơ chế này thay cho Text riêng — cùng kết quả, ít texture hơn. Đánh
dấu XONG nếu anh đồng ý.

## Q42 — Lớp mặt nhân vật (§12.5 A) — chưa làm
Art mặt mới có 3/10 nhân vật; ghép lớp mặt ở client cần đủ art + đổi pipeline atlas. **Đang
dùng:** mỗi mood một bộ frame như cũ. Cần anh quyết có sinh tiếp art mặt cho 7 nhân vật còn lại.
