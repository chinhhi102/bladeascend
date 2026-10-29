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

## Q25 — Bảng tỉ lệ thắng sau v5.4/v5.5 (spec mục 21 và 27 nhờ gửi bảng)

Đã chạy `go run ./cmd/sim --duel` sau khi code xong v5.4 + v5.5. **Đọc bảng này cần một
cảnh báo trước**: bài đo cũ cho mỗi bên một vòng kiếm (10) trong khi trần Kim Đan là 100,
nên nó thổi phồng gấp mười mọi kỹ năng tính theo trần — Bánh Bao Thần ra 100% thắng. Đã đổi
mặc định sang **đủ trần cảnh giới** (9% hoà, còn ít hơn kiểu cũ 18%). Bảng dưới là bản mới.

| Nhân vật | Tỉ lệ thắng | Khắc được | Bị khắc bởi |
|---|---|---|---|
| Đại Ca Dép | **87,0%** | 8 người | Cô Ba |
| Bé Tí | **85,2%** | 7 người | Cô Ba, Đại Ca |
| A Hớn | **75,9%** | 6 người | Bé Tí, Cô Ba, Đại Ca |
| Cô Ba Gánh | 53,7% | 5 người | A Hớn, Bà Tám, Bé Tí |
| Cụ Rùa | 42,6% | 5 người | 4 người |
| Shipper | 31,5% | Bà Tám | 7 người |
| Bà Tám Loa | 24,1% | Thằng Mập | 8 người |
| Ông Ba Võng | 22,2% | Sư Phụ Mì | 5 người |
| Thằng Mập | 20,4% | **không ai** | 8 người |
| Sư Phụ Mì | 13,0% | Bà Tám | 8 người |

**Điểm sáng**: 9/10 nhân vật đều vừa khắc được ai đó vừa bị ai đó khắc — đúng ý anh
*"không phải ai yếu ai mạnh mà nên là khắc chế lẫn nhau"*. Kể cả Đại Ca 87% vẫn bị Cô Ba
khắc, nên không ai là lựa chọn vô đối.

**Ba chỗ lệch**: Đại Ca 87% và Bé Tí 85% quá cao (spec nói > 60% là phải giảm); Thằng Mập
không khắc được ai. Tôi **chưa chỉnh** vì anh đã nhắn *"không nhất thiết phải cân bằng, cứ
để người chơi feedback"*, và vì mọi con số này đều là số spec vừa chốt. Nếu anh muốn chỉnh
thì ba nút gần nhất: `dan_em_sword_share` (0.5), `brace_secs` (5), `tran_chau_secs` (5).

## Q26 — Đạo Tổ vẫn không Thiêu Đốt được, dù v5.6 đã cho Tiên giới đốt Tiên khí

v5.6 §4.1c nói người Tiên giới đốt Tiên khí và **"không bao giờ xuống dưới mốc khởi điểm của
cảnh giới đang đứng"**. Luật đó chạy đúng cho Chân Tiên → Kim Tiên: họ đứng giữa hai mốc nên
luôn có phần dư để đốt.

**Đạo Tổ thì không**, vì một luật khác đã có từ trước: `overflowToScore` ghim Tiên khí đúng
ở mốc 1 500 và đổi mọi phần dư sang Uy Danh (10:1, Q22). Nên Tiên khí của Đạo Tổ **luôn
bằng đúng đáy** → không còn gì để đốt. Hai luật đều đúng theo câu chữ, nhưng cộng lại thì
người ở đỉnh lại là người duy nhất không chạy nhanh được.

Ví dụ trong chính spec — *"Đạo Tổ: 8% × 500 = 40/giây"* — có thể đọc hai cách, và tôi không
đoán: (a) chỉ minh hoạ cách tính 8% bằng con số bước lớn nhất, hay (b) khẳng định Đạo Tổ
phải đốt được thật.

**Các phương án**:
1. Giữ nguyên — Đạo Tổ đã đạt đỉnh, đổi Tiên khí dư sang Uy Danh có giá trị hơn chạy nhanh.
2. Cho Đạo Tổ giữ một phần dư làm "bình xăng": đổi sang Uy Danh phần vượt quá
   `1500 + tien_khi_per_step[4]` thay vì vượt quá 1500.
3. Riêng Đạo Tổ lấy đáy là mốc Kim Tiên (1 000) — đốt được nhưng vẫn không rớt cảnh giới vì
   cảnh giới tính theo `Level`, không theo Tiên khí.

**Tạm dùng**: phương án 1 (hiện trạng, không cần đổi mã). Đã khoá bằng
`TestDaoToHasNoTienKhiHeadroom` — test ghi rõ là khoá hiện trạng, không phải khẳng định ý đồ.
