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

## Trả lời ngày 31/10/2026

**Q28.** Chọn **B** — ảnh của thư "Xong"/"Bỏ qua" quá 90 ngày thì dọn, chữ giữ mãi. Chạy bằng job hằng ngày, ghi audit. *(Làm 31/10: `admin.PurgeOldImages` + `runHousekeeper`, 2 test.)*

**Q29.** Đồng ý **C**, rơi về B khi < 50 mạng / 24 giờ. *(Làm 31/10: `store.ScoreMedianPerMin`, 5 test — gồm cả test chống việc mạng ĐÃ ẨN kéo mốc lên.)*

**Q30.** Làm **A** — nhả atlas nhân vật không còn ai trong khu dùng, trễ 10 giây trước khi nhả. Hạ chất lượng phải trả lại texture thật. Mục tiêu ≤ 180 MB lúc đông người. *(Làm 31/10: `assets.keepChars` + `releaseIdleCharsNow`.)*

**Q33.** a. sửa Bí Kíp +20 ngay · b. phương án A (bảng nhịp là chuẩn) · c. 396 bản ảnh cũ không chuyển, ghi cách lấy bằng git. *(Làm 1/11: hệ số chốt orb 1,5 · √scale · Nam Châm 1,2 · ô viên 160 · bot 25 % — `docs/EXP_AUDIT.md` mục KẾT QUẢ; sim --farm vào CI.)*

Nội dung câu hỏi gốc:

## Q33 — Lạm phát EXP: nguyên nhân thật khác nghi phạm của §6.0 (v6.21)

**Bối cảnh.** CHANGES v6.21 bước 1. Chi tiết và số đo nằm trong `docs/EXP_AUDIT.md`.
Đo bằng `cd server && go run ./cmd/sim --audit`.

- **Nguyên nhân chính là lỗi code.** Ở `server/internal/game/items.go:220`, nhặt Bí Kíp khi đã đủ
  lượt cộng `TotalExp + 20` thay vì +20, tức nhân đôi tổng EXP mỗi lần nhặt. Bí Kíp chiếm 80–97 %
  EXP. Tái hiện được: Luyện Hư → Đại Thừa 10 trong ~70 s khi có Nam Châm.
- Viên EXP và quái trong code **không** nhân theo cảnh giới (spec nghi ×2^r).
- Bỏ lỗi Bí Kíp đi thì farm thuần hiện tại chậm hơn bảng §6.0. Áp nguyên văn §6.0
  (×1,5^r, √scale, Nam Châm ×1,5) thì Đại Thừa nhanh gấp đôi bảng, và Nam Châm vượt trần 1,5 lần.

**Câu hỏi.**
1. Sửa lỗi Bí Kíp về +20 như spec — đồng ý làm ngay?
2. Bảng nhịp §6.0 và các hệ số §6.0 không thể cùng đạt. Chọn:
   - **A (khuyên).** Giữ BẢNG làm chuẩn. Claude chỉnh `exp_economy` bằng `cmd/sim --farm` cho
     trúng ±25 %, rồi anh sửa chữ §6.0 theo hệ số chốt.
   - B. Giữ hệ số §6.0, nới bảng cho khớp số đo.
   - C. Chỉ sửa Bí Kíp + 25 % khi chém bot, giữ nguyên phần còn lại.

**Tạm dùng:** chưa sửa gì. Chờ anh trả lời.


**Q31.** Nhóm 1 xoá · nhóm 2 giữ ở `assets/src` nhưng loại khỏi `assets/dist` · nhóm 3 nối dây. *(Làm 31/10: 71 tệp xoá, `NEVER_SHIP` nới rộng, `icon_skill_scroll` nối vào nút kỹ năng. Dist 462 → 373 tệp, 9,89 → 7,49 MB. `check:dead` nay sạch.)*

Nội dung câu hỏi gốc:

## Q28 — Ảnh góp ý: giữ bao lâu, dọn thế nào?

**Bối cảnh.** A3 (ADMIN_SPEC §7) lưu ảnh chụp màn hình vào `server/data/feedback/`, mỗi ảnh
tối đa 300 KB. Spec không nói hạn giữ. Telemetry có hạn 30 ngày (TELEMETRY_SPEC §3) nhưng
góp ý thì khác — một báo lỗi cũ vẫn có giá trị khi lỗi đó tái xuất hiện.

**Phương án.**
- A. Giữ mãi. Đơn giản, nhưng 100 ảnh/ngày × 300 KB = 11 GB/năm.
- B. Dọn ảnh của thư đã ở trạng thái "Xong"/"Bỏ qua" quá 90 ngày; giữ chữ mãi.
- C. Dọn mọi ảnh quá 180 ngày.

**Tạm dùng: A (giữ mãi).** Không tự xoá dữ liệu người chơi là luật anh đặt cho lượt này.
Dung lượng chưa phải vấn đề ở quy mô hiện tại; khi nào chạm vài GB thì chọn B hay C.

## Q29 — Cờ "Tăng điểm bất thường" (§6.1) cần mốc trung vị từ đâu?

**Bối cảnh.** §6.1 định nghĩa cờ này là "Uy Danh / Linh khí tăng nhanh hơn 5× trung vị cùng
cảnh giới trong 2 phút". Cần một bảng trung vị theo cảnh giới để so.

**Phương án.**
- A. Tính trung vị TRỰC TIẾP từ người đang chơi trong phòng. Tự thích nghi, nhưng phòng ít
     người thì trung vị nhiễu, và nếu cả phòng đều hack thì không ai bị gắn cờ.
- B. Tính sẵn từ `cmd/sim` rồi chốt vào `config/game.json`. Ổn định, nhưng phải cập nhật
     mỗi lần đổi cân bằng.
- C. Lấy trung vị trượt 24 giờ từ bảng `lives` (dữ liệu người chơi thật đã lọc debug).

**Tạm dùng: C**, rơi về B khi chưa đủ dữ liệu (< 50 mạng chơi trong 24 giờ). C dùng dữ liệu
thật nên không lệch khi cân bằng đổi, mà vẫn không nhiễu như A.

## Q30 — iPhone chạm 84% trần bộ nhớ đồ hoạ: giải phóng atlas nhân vật thế nào?

**Bối cảnh.** Máy iPhone thật (Safari 18.7) gửi về `gpu_pressure` ở **216 MB / trần 256**,
trên bản build ĐÃ có đợt giảm bộ nhớ ngày 31/10. Tệ hơn: cơ chế tự hạ chất lượng chạy rồi
mà bộ nhớ **tăng** (214 MB ở mức "Vừa" → 216 MB ở mức "Thấp"), nghĩa là hạ chất lượng
không trả lại texture đã nạp. Sự kiện ghi `chars:7` — bảy atlas nhân vật đang nằm trong bộ
nhớ, và đó là phần lớn chi phí (GAME_SPEC §13.7 nạp `char_<nhân vật>` theo người gặp phải).

**Phương án.**
- A. **Giải phóng atlas của nhân vật không còn ai trong khu dùng.** Đúng gốc rễ, nhưng phải
     đếm tham chiếu cẩn thận; nhả nhầm thì người đó hiện placeholder giữa trận.
- B. **Hạ độ phân giải atlas khi chạm 80% trần** — nạp lại bản 512 thay vì 1024. Giải phóng
     thật, nhưng có một nhịp khựng lúc đổi.
- C. **Trần số atlas (vd 5), nhân vật thứ 6 trở đi vẽ placeholder.** Rẻ nhất, chắc chắn
     nhất về bộ nhớ, nhưng người chơi thấy hình sai — đổi một lỗi thành một lỗi khác.

**Tạm dùng: chưa làm gì, chỉ ghi lại** (nút thắt B13 trong PERF_LOG). Đây là việc lớn và
anh dặn mục lớn chưa làm thì ghi vào bảng chứ không tự làm. Nghiêng về **A**, rơi về B nếu
đếm tham chiếu hoá ra không đáng tin.

**Việc đã làm trong lượt này:** sửa `perfStats().fps` (trước luôn bằng 0 với người chơi
thường, nên mọi bằng chứng gửi về đều thiếu đúng con số cần nhất — nút thắt B14, đã đóng).

## Q31 — 34 tệp art không nơi nào nạp: xoá, hay nối vào code?

**Bối cảnh.** `npm run check:dead` (mới, bước 6) soi 462 tệp trong `assets/dist` và thấy
**34 tệp / 0,58 MB** không có đường nào trong client xin tới. Dưới ngưỡng chặn build (1 MB)
nên nó chỉ cảnh báo. Tôi **không xoá gì** — vì ít nhất một mục trông giống *thiếu dây nối*
chứ không phải *art thừa*.

**Chia làm ba nhóm.**

1. **Gần chắc là rác, xoá được:**
   - `buffs/long.png` — buff Trường Kiếm đã bỏ từ spec v4.3.
   - `obstacles/{bamboo_wall,brick,fence}/states_04.png` — v6.7 rút còn một trạng thái
     (`OBSTACLE_STATES = ['intact']`), 04 là đồ cũ.
   - `vfx/boba_1..6.png` — `boba.ts` vẽ viên trân châu bằng canvas, không dùng ảnh.
   - `obstacles/stall/regrow/*` — art mọc lại, client không nạp.

2. **Art có sẵn nhưng chưa khu nào dùng** (giữ hay xoá là việc thiết kế bản đồ):
   `obstacles/{rope_bridge,village_roof,cliff_wall,lantern_string,ice_island_edge,bamboo_corner}.png`,
   `tiles/{cobblestone,mountain_rock,mossy_stone,damp_earth}.png`,
   `decals/{frozen_reeds,ice_skid,snow_pile}.png`,
   `characters/dan_em/atlas.png` (không nằm trong 10 nhân vật chơi được).

3. **Nghi THIẾU DÂY NỐI, không phải thừa art:**
   - `ui/icon_skill_scroll.png` — vật phẩm `skill_scroll` có thật trong `config/game.json`
     (trọng số 10) nhưng không chỗ nào trong client gọi icon này. Nếu đúng là quên nối thì
     xoá ảnh là xoá nhầm, và lỗi hiển thị vẫn còn.
   - `gates/{blue,orange,purple}/burst_06.png` — khung 06 của hiệu ứng cổng; danh mục chỉ
     khai tới 05. Có thể artgen vẽ dư một khung, cũng có thể danh mục khai thiếu.

**Tạm dùng: giữ nguyên tất cả, chỉ cảnh báo.** Xoá art là việc một chiều, và nhóm 3 cho
thấy "không ai nạp" đôi khi có nghĩa là *code thiếu*, không phải *art thừa*. Anh xem nhóm 3
trước: nếu `icon_skill_scroll` đúng là quên nối thì đó là một lỗi hiển thị đang sống.

## Q32 — Spec §13.6 mục 8 chưa ghi luật chất lượng anh chốt 1/11 (chỉ cần anh sửa chữ)

Bối cảnh: lượt 1/11, việc 4 (chất lượng mặc định). Tệp: `client/src/render/quality.ts`.

Code đã làm đúng lời anh chốt trong chat:
- Mặc định: PC → **Vừa**; điện thoại / GPU tích hợp → **Thấp**. Không ai bắt đầu ở Cao.
- Vừa → Cao chỉ khi ≥ 55 FPS suốt **20 giây**.

Spec §13.6 mục 8 hiện ghi: "FPS ổn lại **10s** → bật dần" và không nói mức mặc định. Tôi giữ
**Thấp → Vừa = 10 giây** theo spec, chỉ bậc lên Cao là 20 giây. Không sửa `docs/_raw/`.

Anh chỉ cần thêm vào §13.6 mục 8 (đề xuất câu chữ):
> Mặc định: PC Vừa, điện thoại/GPU tích hợp Thấp. Thấp→Vừa sau 10s ổn ≥ 55 FPS; Vừa→Cao sau 20s.
> Chỉ lựa chọn tay của người chơi được lưu; máy tự đổi không lưu.

Nếu anh muốn Thấp → Vừa cũng 20 giây thì báo, đổi một hằng số.

**Trả lời (1/11):** GAME_SPEC v6.20 §13.6 mục 8 — giữ Thấp→Vừa 10 s, Vừa→Cao 20 s; thêm bước khẩn ở Thấp < 30 FPS, khoá tự nâng 5 phút sau khi nâng rồi bị hạ, phím V có nấc "Tự động". Đã làm, `npm run check:quality` kiểm cả ba.

## Q34 — `bot_fill_target`: spec ghi 20, anh chốt 60 hôm 29/10

**Bối cảnh:** lượt lớn, mục 1.2 (v6.22). Spec §9.3 và khối config §16 ghi
`bot_fill_target (20)`. Code đang để **60** theo quyết định của anh ngày 29/10 (commit
`0561a52`: trần phòng 60 nhân vật, tối đa 20 người thật, sàn 10 nhân vật mỗi khu).
Hai con số không đi được với nhau: 5 khu × sàn 10 = 50 nhân vật, mà trần phòng 20 thì một
người vào sẽ thấy mỗi khu chỉ 4 nhân vật.

**Phương án:**
- **A.** Giữ 60 (và sàn 10/khu). Anh sửa spec §9.3 + §16 thành 60.
- **B.** Theo spec: 20, đồng thời hạ `zone_min_chars` xuống 4. Thế giới 5 khu sẽ vắng hẳn.

**Tạm thời:** giữ **60** (A). Đây là quyết định đã chạy ổn từ 29/10, và B làm thế giới vắng.

## Q35 — Phòng 50 người chưa qua loadtest trên server thật → đang giữ 20

**Bối cảnh:** lượt lớn, mục 1.2. Spec §9.1 cho phép mở 50 người thật/phòng sau khi loadtest
50 người + bot đạt tick p95 mỗi khu < 4 ms. Kết quả (PERF_LOG lượt #24):

| Máy | p95 khu nặng nhất | Kết quả |
|---|---|---|
| Máy dev (i5-12400F) | 2.02 ms | đạt |
| Staging (ARM 4 nhân), 3 lần đo | 5.19 / 4.28 / 5.16 ms | **không đạt** |

Phần vượt chỉ nằm ở **một khu**: khu có 10 người Đạo Tổ đủ kiếm đứng chụm (cảnh nặng nhất).
Bốn khu kia chỉ 1,6–2,3 ms.

**Phương án:**
- **A.** Giữ 20 cho tới khi tối ưu được khu nặng (đưa vào việc tối ưu sau).
- **B.** Mở 50 trên máy chủ mạnh hơn (production khác staging?). Anh cho biết máy production.
- **C.** Mở 50 ngay, chấp nhận khi 10 Đạo Tổ chụm thì khu đó hơi trễ.

**Tạm thời:** **A**, giữ `room.max_humans = 20`, đúng như lệnh "không đạt thì giữ 20".

**Đã trả lời (GAME_SPEC v6.28, CHANGES_v6_28, 1/11/2026):** bỏ `bot_fill_target`, bot lấp THEO
KHU với sàn `zone_min_chars = 10`; trần người thật là cấu hình theo môi trường — production 50,
staging 20 — chỉnh được ở admin (lưu SQLite, sống qua deploy). Tự tắt bot khi người thật ≥
`bot_auto_off_at` (production 40, staging 20), bật lại khi < ngưỡng − 5. Đã làm ở commit v6.28.

## Q36 — `cmd/sim --duel` trượt hàng rào 35–65 %, chỉnh hồi chiêu ±30 % không cứu được

**Bối cảnh:** làm CHANGES v6.26/v6.27 (sau giai đoạn 1). Sim đấu tay đôi Kim Đan, đủ trần
kiếm, 20 trận/cặp, 60 s/trận; thắng = chém chết đối thủ (sự kiện Death). Đã thử 2 bộ số:

| Nhân vật | Theo spec (hồi · lượt) | Thắng | Hồi chỉnh tới ±30 % | Thắng |
|---|---|---|---|---|
| Bé Tí | 35 s · 1 | **100 %** | 45 s | **99 %** |
| Sư Phụ Mì | 30 s · 2 | 69 % | 39 s | 74 % |
| Đại Ca Dép | 40 s · 1 | 69 % | 52 s | 73 % |
| A Hớn | 30 s · 2 | 68 % | 39 s | 72 % |
| Thằng Mập | 20 s · 2 | 52 % ✅ | 20 s | 43 % ✅ |
| Cụ Rùa | 30 s · 2 | 33 % | 21 s | 34 % |
| Shipper | 20 s · 2 | 26 % | 14 s | 28 % |
| Cô Ba | 22 s · 2 | 22 % | 15 s | 20 % |
| Bà Tám | 18 s · 2 | 17 % | 13 s | 13 % |
| Ông Ba Võng | 25 s · 2 | 17 % | 18 s | 19 % |

Hồi chiêu gần như không đổi kết quả: vào trận ai cũng có sẵn 1 lượt, và nay chém trúng là
chết nên **lần bấm đầu tiên quyết định trận**. Kỹ năng thắng va kiếm (Gồng Cơ ×2 độ bền,
cướp kiếm, hồi kiếm, đàn em) luôn thắng kỹ năng khống chế/thoát thân. Ngoài ra 35–38 % trận hoà.

**Phương án:**
- **A.** Đổi hiệu ứng, không đổi hồi: vd Gồng Cơ ×1.5 thay ×2 / 3 s; Hút Mì 15 %; A Hớn 12 %;
  đàn em 20 % kiếm. Anh duyệt số rồi em chạy lại sim.
- **B.** Nới hàng rào cho "kỹ năng gây khó chịu" (spec §6.3 đã cho phép yếu hơn khi đấu tay đôi):
  chỉ áp trần 65 % cho mọi người, bỏ sàn 35 %. Bé Tí vẫn phải sửa.
- **C.** Vào trận 0 lượt (phải chờ hồi lượt đầu) để hồi chiêu có tác dụng.

**Tạm thời:** giữ đúng bảng hồi chiêu của spec (không đổi số), sim báo trượt. Không tự đổi
hiệu ứng kỹ năng vì đó là luật trong spec.

**Ghi chú nhỏ cùng lượt:** §11.1 ví dụ "Trúc Cơ chém Nguyên Anh → 800" nay phải là 850 (mọi
lần chém trúng đều là chém chết, +50 — CHANGES v6.26 mục 4). Code theo 850; anh sửa ví dụ khi tiện.
Tương tự, checklist §17 dòng "Bị chém ở Luyện Khí… người chém nhận 50% EXP nạn nhân + 30" còn số
cũ; code theo v6.26 (30% + thưởng hạ, trần +2 cảnh giới).

**Đã trả lời (GAME_SPEC v6.30, 1/11/2026):** bỏ hàng rào tỉ lệ thắng tay đôi — `sim --duel` chỉ
còn số tham khảo; cân bằng theo số liệu người chơi thật (bảng "Nhân vật" trong admin). Bà Tám
toả tròn 360°, Ông Ba Võng thêm Trộm Kiếm 4 s, ví dụ Uy Danh sửa thành 850. Đã làm ở A1–A5.
