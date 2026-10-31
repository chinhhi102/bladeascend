# Câu hỏi đã xong (lưu trữ)

Những câu đã được trả lời — giữ lại để tra sau, không cần đọc nữa.
Câu **đang chờ** nằm ở `docs/QUESTIONS.md`.

> Mỗi câu đúng **một dòng**. Trước đây Q7–Q14 nằm cả trong bảng lẫn trong phần "Chi tiết"
> bên dưới, nên đọc file này là đọc hai lần cùng một nội dung. Cần chi tiết đầy đủ thì tra
> `docs/_raw/ANSWERS.md` (lời chủ dự án) và `docs/PROGRESS.md` (việc đã làm).

| Câu | Nội dung | Kết luận |
|---|---|---|
| Q1–Q6 | Thưởng hạ sát, gió theo góc, viên EXP thuộc người chém, chết rơi kiếm, 1 anim/ảnh | Trả lời trong `_raw/ANSWERS.md`, đã áp dụng |
| Q7 | Nhịp Luyện Khí → Trúc Cơ | Chốt 30–60s; sau đó đổi lại, xem Q13 |
| Q8 | Checklist §17 ghi số kiếm mất theo bản cũ | Spec sửa: `clash_unit` = 1, checklist khớp |
| Q9 | `min_count = 0` có làm chết không | Không. 0 kiếm chỉ bị tước vũ khí; chỉ Luyện Khí bị chém thân mới chết |
| Q10 | Phạm vi nhặt vật phẩm | Giữ `pickup_r` = 90. Luật vốn đã đúng: kiếm không nhặt được đồ |
| Q11 | "Tính exp dư ra làm…" (câu bị cắt) | §3.2: Đại Thừa 10 thì EXP thừa bị bỏ, xếp hạng dùng Uy Danh |
| Q12 | Tốc độ chạy 130 hay 260×√scale | Gộp vào Q14; code giữ 130, sau đó Q13 đổi lên 195 |
| Q13 | Nhịp farm quá chậm | Chấp nhận nhịp hiện tại — lên cấp chủ yếu nhờ cướp Linh khí. `move_base` → 195, `exp_orbs.cell` → 140 |
| Q14 | `_raw` thiếu 13 mục so với code | Spec v4.0 gộp đủ; `realm_mult` giữ 2.0 |
| Q15 | Lá chắn §5.2 làm không ai chém chết được ai | Hiểu nhầm lá chắn thành tường chắn đường. Chuyển thành **luật chém**: người chơi không chặn đường nhau, còn 1 vòng mới bị chém thân |
| Q16 | Trường Kiếm | Bỏ hẳn (v4.3); v4.4 thay bằng **Cương Kiếm** |
| Q17 | Bó Kiếm khi rớt cảnh giới | Phương án 1 — Bó Kiếm chủ yếu từ xác người chết và vật phẩm bản đồ |
| Q18 | Linh khí / Tiên khí | Tách hẳn, không quy đổi qua lại. Bỏ `tien_khi_to_exp_base` |
| Q19 | Tốc độ bể kiếm | Giữ `clash_cooldown: 0`, `flip_cooldown: 0.3` |
| Q20 | Sửa lỗi va kiếm làm trận chậm hẳn, có chỉnh cân bằng không | **Phương án 1** — giữ luật đúng. Mốc mới: 84 lượt chết / 5 phút / 40 bot; số cũ (608/10 phút) đo trên luật sai. Chơi thử rồi mới quyết; nếu chỉnh thì ưu tiên tăng `sword.width` |
| Q21 | Cổng ngẫu nhiên + Đạo Tổ đổi Tiên khí dư | Cổng là **đích ngẫu nhiên** (không phải cổng tự đổi chỗ), §8.4 viết lại thành mạng chung 20 cổng. Đạo Tổ đổi dư sang Uy Danh, ghi vào §3.4 |
| Q22 | `tien_overflow_score`: 1.0 hay 0.1 | **0.1** (10 Tiên khí = 1 Uy Danh) |
| Q23 | Spec còn ghi cách bật Thiêu Đốt cũ | Spec v5.6 §4.1c thêm "nháy đúp rồi giữ / chạm 2 lần rồi giữ"; §18.1 thêm dòng Kỹ năng |
| Q24 | Lên Tiên giới là mất Thiêu Đốt | **Phương án 2** — Tiên giới đốt **Tiên khí**, 8% mỗi giây, không dưới đáy cảnh giới (spec v5.6, sau đó v5.7 đổi mốc 8% sang độ rộng khoảng) |
| Q25 | Bảng tỉ lệ thắng 10×10 sau v5.4/v5.5 | **Bỏ hẳn cách cân bằng theo tỉ lệ thắng 1-đấu-1** — kỹ năng có loại tăng sức mạnh, có loại khắc chế, tay đôi không phản ánh giá trị. Đã xoá `cmd/sim --duel`; cân bằng theo phản hồi người chơi + số liệu trận thật (§6.3) |
| Q26 | Đạo Tổ vẫn không Thiêu Đốt được | Thay toàn bộ cách tính Tiên giới (spec v5.7): cảnh giới theo **tổng Tiên khí cộng dồn**, Đạo Tổ nạp 1 001 → 1 500 và chỉ đổi Uy Danh khi đang đúng 1 500 → Đạo Tổ Thiêu Đốt được |

## Hai việc chờ quyết, nay đã có

- **Tối ưu (CHANGES_v4 mục 25–27)**: **đi tiếp**. Thứ tự chủ dự án chốt: mạng (27) → server
  (26, pprof) → client phần rẻ (25.1 vòng kiếm thành texture, BitmapText, cull). Đo lại
  `loadtest` + 10 tab Chrome headless (thường và 4G yếu), ghi số trước/sau.
- **Thư viện `pixi-filters` và `revolt-fx`**: **đã duyệt**. Quy tắc chung đã thêm vào
  `CLAUDE.md`: thư viện spec ghi đích danh coi như đã duyệt, thư viện khác vẫn hỏi.

---

## Bó Kiếm chưa có job art (30/10/2026) — ✅ ĐÃ TRẢ LỜI

**Bối cảnh**: đang làm item art sau khi anh báo "item cho skill còn thiếu". Rà hết 13 loại
vật phẩm thì thấy **Bó Kiếm là loại duy nhất không có ảnh** — nó đang vẽ bằng Graphics (ba
lưỡi chéo buộc dây đỏ). `ART_PROMPTS.md` không có job nào cho nó.

**Em đã làm**: thêm job `item_bundle` → `items/bundle.png` vào `cmd/artgen/catalog.go`, prompt
đồng bộ với `item_sword` (bong bóng xà phòng bóng, bên trong là ba thanh kiếm nhỏ phát sáng
bắt chéo buộc dải vải đỏ), cỡ 128 px như kiếm vì vẽ ra cùng kích thước.

**Cần anh chốt**: có đưa `item_bundle` vào `ART_PROMPTS.md` không? Nếu anh muốn Bó Kiếm giữ
kiểu vẽ tay như hiện tại thì nói em bỏ job đi — em chưa sinh ảnh, chỉ mới thêm job.

Hai loại còn lại cố ý vẽ tay và em KHÔNG đụng vào: `sword` (đã có ảnh, dùng nhánh riêng) và
`tien_khi` (viên vàng kim đổi kích thước + vành ngoài theo giá trị, ảnh tĩnh không tả được).

**Trả lời (ANSWERS 30/10):** ĐƯA VÀO. ART_PROMPTS §8 đã thêm prompt (bong bóng chứa bó **5**
kiếm buộc ruy băng đỏ, xoè nhẹ, hào quang vàng). Đã sinh ảnh, vẽ to hơn bong bóng kiếm thường
20%, bó rơi từ xác in số kiếm lên bong bóng. Xong trong commit `9bca2c9`.

---

## Q27 — Tên vật trang trí + nền 5 khu · ĐÃ TRẢ LỜI 31/10

**Anh chốt:**
1. Nền: đổi khoá `BIOME_TILES` sang **tên biome**, thêm test "mỗi khu phải tra ra đúng ảnh
   nền tồn tại trong manifest".
2. Trang trí: mapgen **chỉ chọn decor có trong manifest**; mỗi decor **gắn tag khu**
   (`red_plastic_stool → pham_tran`). Decor hợp chủ đề mà chưa có art (`rock_small`, `pine`,
   `lantern`, `well`…) → **thêm job artgen**, chưa có art thì **không đặt**. Test: 5 khu ×
   10 seed, mọi decor đặt ra đều có ảnh.

**Đã làm xong cả hai** — xem PROGRESS "Phiên 31/10". Vài điều đáng ghi lại:

- Chẩn đoán ban đầu của tôi sai một nửa: **có hai bảng `BIOME_TILES`** đã trôi khỏi nhau, và
  chỉ Băng Nguyệt Hồ trải cỏ (thiếu khoá `snow`), không phải cả 5 khu. Linh Sơn Cốc thì nhận
  bộ nền của làng vì bộ nền núi đá nằm dưới một khoá chết (`rock`). Đã **gộp còn một bảng**.
- Mật độ trang trí phải là **tổng của khu**, không phải mỗi loại: để "mỗi loại n lần" thì
  Phàm Trấn nhảy lên 447 vật chỉ vì nó hợp 9 loại, và sẽ tự lên ~720 khi 7 ảnh còn thiếu về.
- Test "mỗi khu phải có ít nhất MỘT loại trang trí" bắt được lỗi thật ngay lần đầu: Băng
  Nguyệt Hồ trống trơn vì cả hai loại được gắn cho nó đều chưa có art. "Chỉ đặt cái có art"
  có một cách thoả mãn rất tệ là không đặt gì cả.

**Còn chờ anh (không chặn gì):**
- 7 ảnh của job `decor_nature` chưa sinh được — tài khoản minigen hết hạn cookie. Đăng nhập
  lại rồi `go run ./cmd/artgen run -only decor_nature` và `go run ./cmd/mapgen`.
- Thiên Đình và Trúc Lâm cùng `biome: "grass"` nên trải cùng bộ nền. Muốn Trúc Lâm có nền tre
  riêng thì đổi `zoneDefs` cho `truc_lam` sang biome riêng + sinh lại bản đồ.
