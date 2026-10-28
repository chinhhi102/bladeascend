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
