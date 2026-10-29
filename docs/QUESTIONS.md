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

## Q21 — Hai luật mới chủ dự án yêu cầu, GAME_SPEC chưa ghi

Chủ dự án yêu cầu trực tiếp ngày 29/10/2026, tôi đã làm trong code nhưng **không được ghi
vào `docs/_raw/`**, nên spec và code đang lệch nhau. Nhờ anh cập nhật spec:

**1. Cổng dịch chuyển bắn ngẫu nhiên** (lệch §8.4 — spec hiện ghi "cổng đi 2 chiều, ghép
cặp cố định"). Đã làm: bước vào một cổng thì ra ở **một cổng bất kỳ khác cùng mạng**, không
bao giờ ra lại chính cổng vừa vào. Bật/tắt bằng `gates.random_exit` (mặc định `true`); đặt
`false` là quay về đúng luật cũ.

Lưu ý tôi hiểu câu "cổng dịch chuyển sẽ dịch chuyển ngẫu nhiên" là **đích đến ngẫu nhiên**.
Nếu ý anh là **bản thân cái cổng tự di chuyển chỗ trên bản đồ** thì nói một tiếng, đó là
việc khác hẳn và tôi làm lại.

**2. Đạo Tổ: Tiên khí dư đổi thành Uy Danh** (spec chưa nói gì về chuyện này). Đã làm: tới
Đạo Tổ là hết cảnh giới để lên, nên mỗi điểm Tiên khí nhặt thêm đổi thành
`cultivation.tien_overflow_score` điểm Uy Danh — chủ dự án chốt tỉ lệ **10 Tiên khí = 1 Uy
Danh**, tức `0.1` và thanh Tiên khí **giữ đúng ở
mốc Đạo Tổ** thay vì phình mãi.

Vì sao giữ thanh ở mốc đỉnh: `robTienKhi` tính `loss = tiên_khí − mốc_khởi_điểm_mới`. Nếu
cho phình thì một người Đạo Tổ ngồi lâu thành cục mồi khổng lồ, ai chém được là giàu to —
không chặn thì luật cướp đoạt §3.4 méo hẳn. Nếu anh muốn giữ nguyên cục mồi đó thì báo.
