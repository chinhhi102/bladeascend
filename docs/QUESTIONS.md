# Câu hỏi đang chờ anh trả lời

Mỗi câu: bối cảnh → phương án → **tạm dùng** (đã code theo phương án này, đổi lại dễ).
Câu đã xong ở `docs/QUESTIONS_DONE.md`.

---

*(Chưa có câu nào đang chờ. Q15–Q19 anh đã trả lời trong `docs/_raw/ANSWERS.md` ngày
2026-10-28, đã chuyển sang QUESTIONS_DONE.md và code đã làm theo.)*

Việc đang chờ anh **quyết**, không phải câu hỏi về luật chơi:

- **CHANGES_v4 mục 25–27 (tối ưu)**: chính anh ghi *"Gửi tôi bảng số đo trước khi tối ưu"*.
  Bảng đo đã có ở `docs/PROGRESS.md` mục "Đo hiệu năng trước tối ưu" — chờ anh xem rồi cho
  đi tiếp.
- **CHANGES_v4 mục 29–33 (v4.6 nâng cấp VFX)**: cần cài **hai thư viện mới** —
  `pixi-filters` và `revolt-fx` — mà CLAUDE.md ghi "không thêm thư viện khi chưa hỏi".
  Spec có ghi tên cụ thể nên coi như anh đã duyệt; em hỏi lại một câu cho chắc trước khi
  thêm vào `package.json`.

## Q20 — Sửa lỗi va kiếm làm trận đấu chậm hẳn lại, có cần chỉnh cân bằng không?

**Bối cảnh**: tối ưu CPU server (CHANGES_v4 mục 26) phát hiện `geom.SegSegDist2` viết sai
dấu, khiến hai thanh kiếm cách nhau cả trăm đơn vị vẫn bị tính là va vào nhau (chi tiết
trong `docs/PROGRESS.md`). Lỗi đã sửa, có test khoá lại.

**Vấn đề**: mọi con số cân bằng đo trước đây đều đo trên luật va kiếm sai.
`go run ./cmd/sim --bots 40 --minutes 5` giờ cho **84 lượt chết** (2,10 lượt/bot), trong khi
con số ghi trong PROGRESS là 608 lượt trong 10 phút (≈ 304 trong 5 phút). Tức là trận đấu
chậm lại khoảng **3,5 lần**.

**Các phương án**:
1. Giữ nguyên — nhịp chậm lại là nhịp ĐÚNG theo spec, trước đây nhanh là vì lỗi.
2. Tăng tốc bể kiếm để quay lại nhịp cũ (ví dụ giảm `sword.durability_base`, hoặc tăng số
   kiếm rụng mỗi lần va).
3. Tăng tầm chém / bề rộng lưỡi để hai vòng kiếm dễ chạm nhau hơn.

**Claude không tự chọn**: đây là đổi luật chơi, không phải sửa lỗi. Cần chủ dự án quyết.

## Q22 — `tien_overflow_score`: spec ghi 1.0, anh nhắn 10:1 (0.1)

`docs/_raw/CHANGES_v5.md` mục "Trả lời Q21" ghi giữ `tien_overflow_score` **1.0**, nhưng
ngay trước đó anh nhắn *"tỉ lệ tiên khí và uy danh là 10-1, tức 10 tiên khí chỉ được 1 uy
danh"* tức **0.1**.

Tôi đang để **0.1** theo lời nhắn, vì nó mới hơn và nói rõ tỉ lệ. Nếu spec mới đúng thì đổi
`config/game.json` → `cultivation.tien_overflow_score: 1.0` và sửa con số khoá trong
`server/internal/game/newrules_test.go`.

## Q23 — Spec còn ghi phím điều khiển cũ (chỉ cần anh dán lại vào `_raw`)

Anh nhắn ngày 29/10: *"thiêu đốt tinh huyết là shift, hoặc double-click long, sử dụng skill
= ctrl hoặc bấm vào button"*. Code đã làm đúng vậy, nhưng `docs/_raw/GAME_SPEC.md` vẫn ghi
bản cũ — tôi không ghi vào `_raw` nên nhờ anh sửa hai dòng:

- **§4.1c dòng 314**: `Giữ **Shift** (PC) hoặc giữ nút **🔥** (điện thoại)` →
  thêm *"hoặc **nháy đúp rồi giữ** chuột / chạm 2 lần rồi giữ"*.
- **§6.3 dòng 428**: `Phím **Space** (PC)` → `Phím **Ctrl** (PC)`.
- **§18.1 dòng 1116**: bảng hướng dẫn giờ liệt kê thêm dòng Kỹ năng.

**Một việc cần anh quyết**: phím đi là WASD, nên giữ Ctrl rồi bấm W là **Ctrl+W = đóng tab**.
Trình duyệt không cho trang chặn phím đó (chi tiết trong `docs/PROGRESS.md`). Tôi đã làm
phần giảm thiểu khoá bàn phím khi toàn màn hình, nhưng nó chỉ chạy ở chế độ toàn màn hình
và chỉ trên Chrome/Edge. Nếu anh thấy rủi ro mất tab là quá lớn thì đổi sang phím khác —
`Q`, `F` hoặc **chuột phải** đều không đụng phím tắt trình duyệt.
