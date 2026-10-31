# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

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
