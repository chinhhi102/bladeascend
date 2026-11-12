# Tiến độ — BladeAscend (Đạo Hữu Xin Dừng Bước)

_Cập nhật: 2026-11-05 (bong bóng mặt §12.5 v6.38, Q44 đóng) · 2026-11-02 (v6.31–v6.36, mở công khai, logo, hiển thị kỹ năng; trang chủ + bộ đo + telemetry theo build). Nhật ký đầy đủ các phiên trước (22/10 → 01/11, 1 677 dòng):
`docs/history/PROGRESS_den_2026-11-01.md` — chuyển nguyên văn, không xoá._

## Cách chạy (máy dev)

```bash
cd server && go run ./cmd/server          # http://localhost:9999, mở cho cả mạng LAN
cd client && npm run dev                  # http://localhost:5173 (tự cập nhật khi sửa code)
cd client && npm run dev:admin            # trang quản trị React (cần server chạy)
cd server && go test -race ./... && go vet ./...
cd client && npm run build                # asset-pipeline + mọi check:* + game + admin + trang SEO
cd server && go run ./cmd/mapgen --check  # kiểm 5 bản đồ
tools/deploy.sh                           # deploy :9999 (rollback: tools/deploy.sh rollback)
(set -a; . docs/private/deploy; set +a; deploy/staging.sh)   # deploy staging
cd client && npm run perf:report          # báo cáo hiệu năng → docs/reports/<ngày>-hieu-nang/
node tools/browsertest/review.mjs <url> docs/review/<mục> <tên> [--full]   # chụp PC + iPhone
```

## Trạng thái theo mảng

| Mảng | Trạng thái | Ghi chú |
|---|---|---|
| Hiển thị kỹ năng v6.36 | ✅ chờ duyệt ảnh | đủ 5 lớp × 10 kỹ năng, event StatusStart/End, icon vector 64 px, HUD + mũi tên — `docs/review/skills/`, `docs/SKILL_VFX_AUDIT.md` |
| Luật chơi v6.30 | ✅ | chém trúng = chết (v6.26), chết = hết ván (v6.29), kỹ năng tự hồi (v6.25), Loa 360°, Trộm Kiếm |
| Bot | ✅ | sàn theo khu v6.35 (production Thiên Đình 30 · rìa 20; staging 15/10, admin chỉnh từng khu), tên đa quốc gia v6.34, lấp theo khu (v6.28), 5% đứng yên sau khi vào, dùng kỹ năng theo luật v6.9, không lộ cờ bot |
| Trang chủ | ✅ chờ duyệt ảnh | lưới 5 thẻ, nền kéo hết trang, ô chọn ngôn ngữ — `docs/review/06-…`, `07-…` |
| Logo Pattaya | ✅ | trang chủ PC/điện thoại/ngoại ngữ + màn tải — `docs/review/08-logo/` |
| Ảnh chia sẻ OG | ✅ ĐÃ DUYỆT, đang chạy | bản có logo · `SHARE_CARD_APPROVED = True` · `og:image` → `/brand/share_card.jpg` (213 KB) |
| Góp ý 3 ảnh | ✅ | client nén JPEG ≤ 300 KB, server gỡ EXIF + giải mã/mã hoá lại (Q39), Playwright 4 thiết bị |
| Admin React + antd | ✅ | đã chạy trên staging, ảnh 10 trang `docs/review/11-admin/` |
| Đa ngôn ngữ | ✅ bản nháp | vi + en/pt-BR/es/id/th (`_draft`, cần người bản ngữ duyệt — `client/src/i18n/NOTES.md`) |
| SEO | ✅ | `/`, `/en/`, `/pt/`, `/es/`, `/id/`, `/th/`, hreflang, `sitemap.xml` |
| Hiệu năng | ✅ server/mạng · 🔴 thiết bị thật | `docs/reports/2026-11-01-hieu-nang/`; FPS/GPU iPhone thật chưa đo |
| Đối chiếu spec | — | `docs/SPEC_AUDIT_v630.md` + mục 6b (rà lại 02/11): còn **2 mục** chưa làm — 1 anim giật, Web Worker/RenderTexture (lớp mặt đã ĐÓNG 05/11 bằng bong bóng mặt, Q44). 2 mục từng chấm là lỗi hoá ra chấm sai |
| Bộ đo browsertest | ✅ | 15 bài từng chết im lặng vì i18n (02/11) — nay dùng selector theo class, luật ở `tools/browsertest/README.md` |

## Lượt cuối 01/11/2026 — đã làm

Xem bảng 20 mục trong báo cáo cuối phiên (chat) và `git log --oneline 1476271..HEAD`. Tóm tắt:
B6–B9 trang chủ / logo / OG · C10–C12 góp ý ảnh, admin React, hướng dẫn mật khẩu staging ·
D13–D16 i18n, bản dịch nháp, chọn ngôn ngữ + font Thái, trang SEO · E17–E20 dọn V6_AUDIT
(bot, FlagBot, mệnh giá Tiên khí, khói vật cản, luật khống chế, Đạo Tổ trên minimap),
đối chiếu spec, `npm run perf:report`, PROGRESS/PERF_LOG.

## 02/11/2026 — đã làm

v6.31–33 luật cộng `Stat()` + nội tại 10 nhân vật (`9121f0b`) · mở công khai repo: README,
CONTRIBUTING, SECURITY, CoC, CI cho fork, quét dữ liệu người chơi (`ebc3450`) · v6.34 tên bot
7 ngôn ngữ 60/25/15 (`77b2cdb`) · Q39 mã hoá lại ảnh góp ý (`732c744`) · logo Pattaya
(`47a62ba`) · giấy phép MIT + art CC BY-NC-SA + mục "Tên và logo" (`994e508`) · v6.35 sàn bot
theo khu + loadtest S13 (PERF_LOG #28) + Nam Châm `bonus: 0.2`.

**v6.36** — OG có logo đã duyệt + SECURITY chỉ qua GitHub private reporting + cảnh báo admin
tick p95 > 3,5 ms suốt 5 phút (`bbba31f`) · bảng rà 10 kỹ năng (`0ba5c75`) · đủ 5 lớp hiển thị
kỹ năng, event `StatusStart/StatusEnd`, icon trạng thái vector + job artgen `icon_status_*`,
HUD icon + giây, rung, mũi tên mép màn hình, sạp Cô Ba có bạt/đếm/vỏ chuối, 12 tiếng kỹ năng
tạm (`ca58995`) · ảnh duyệt + công cụ `tools/browsertest/skillshots.mjs` (`bb1cd96`).

Chụp lại ảnh kỹ năng: `cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true`
rồi `node tools/browsertest/skillshots.mjs http://127.0.0.1:9998` (phím debug mới `U` = kéo
người thật khác về cạnh mình, chỉ ở chế độ dev).

### 02/11 (chiều) — trang chủ, bộ đo, telemetry

**Trang chủ: nút "Vào chơi" biến mất ở 4/7 cỡ màn hình** (`c40329f`), gồm cả 1366×768 — laptop
phổ thông nhất. Cả panel là một khối `overflow-y: auto` nên thứ bị đẩy khuất là thứ nằm cuối,
tức hai cái nút; trang chủ từ 1/11 dày thêm ảnh logo, ô chọn ngôn ngữ, dòng hồi chiêu, bốn
thanh chỉ số. Nay lưới nhân vật + khung chi tiết vào vùng cuộn riêng (`.home-scroll`), phần
đầu và phần cuối luôn nằm ngoài vùng cuộn — thêm nội dung nữa cũng không đẩy mất nút. Kèm:
thẻ cho 3 dòng hồi chiêu và tên xuống 2 dòng trên màn hẹp (trước đó cắt thành "tích 2…",
"Thằng …"). **7/7 cỡ ĐẠT.**

**15 bài browsertest đã chết im lặng từ lúc có i18n** (`7c230f9`). Tất cả tìm phần tử bằng
`getByText('Vào chơi')`; Chromium headless báo locale Anh nên nút thành "Play" → TimeoutError
ngay ở cỡ màn hình đầu tiên. Trong quãng đó trang chủ vỡ như trên mà không bài nào kêu. Nay
dùng selector theo class, luật ghi ở `tools/browsertest/README.md`.

**Hai phép kiểm canh sai thứ, đã sửa.** `panel.scrollHeight - clientHeight` luôn bằng 0 sau
khi panel đổi sang `overflow: hidden` → sẽ luôn ĐẠT. "Mỗi thẻ đúng một dòng" thì ĐẠT trong
khi thẻ hiện "Thằng …" — một dòng và đọc được là hai chuyện khác nhau. Nay canh: nút "Vào
chơi" có trong vùng nhìn thấy không, và tên có bị cắt không.

**Telemetry tách theo build + loại phiên máy** (`64b2734`). Báo cáo gộp cả ngày mà 31/10
deploy 17 lần. Thêm mục 8 "Theo BUILD"; ngưỡng cảnh báo §4.3 chỉ tính build mới nhất. Và:
**64/67 phiên buổi tối 31/10 đến từ đúng một máy chạy Chromium của Playwright** — bài đo 10
tab. Cảnh báo "96% mẫu FPS 1% dưới 30" phần lớn là tiếng vọng của chính bài đo. Client nay
gửi `navigator.webdriver`, báo cáo loại phiên máy ra và khai số đã loại.

**Gói hỏng không còn thoát ra ngoài** (`64b2734`). Đường snapshot nhị phân có `try/catch` từ
đầu, đường msgpack thì không — nên một gói hỏng ném thẳng ra khỏi `ws.onmessage`
(`RangeError: Extra 18 of 121 byte(s)`, Safari, telemetry 31/10). Nay `decodeFrame()` trong
`net/frame.ts` không bao giờ ném; gói hỏng → bỏ qua + báo telemetry kèm loại gói, độ dài,
8 byte hex đầu. `npm run check:decode` vào chuỗi build: 8 gói golden + 12 000 gói fuzz + phép
kiểm trạng thái.

**Respawn: điều tra xong, không sửa gì** (`7c230f9`). Tách theo build (chỉ người thật): đỉnh
829 lần ở build `20261031_1050`, tụt hẳn sau đó — đúng lúc bản sửa `KEEP_GONE_SECS` lên.
`respawn-wander.mjs` trên bản hôm nay: 0. Theo luật "bench phải tái hiện được trước khi sửa"
nên dừng. Việc đã làm: ba con số nguyên nhân (Despawn · hết hạn 2s · quái) nay đi về telemetry
và in trong cảnh báo — trước đó chỉ có tổng, nên mất nửa ngày dựng lại chuỗi build mới biết
bản nào đã đóng nó.

**Quy tắc 12 — vì sao lượt này KHÔNG thêm dòng PERF_LOG.** Ba thay đổi hôm nay đều không nằm
trên đường nóng: CSS/DOM trang chủ (ngoài trận), selector của bài đo (không vào bản phát
hành), và ba số nguyên respawn đi kèm mẫu telemetry 10 giây một lần. `decodeFrame` — thứ duy
nhất nằm trên đường gói tin — lên từ 1/11 nên đã nằm trong số đo #29. Một chỗ đáng để mắt khi
có máy thật: vùng cuộn trang chủ dùng `mask-image`, trên điện thoại yếu nó có thể tốn; trang
chủ không nằm trong kịch bản S1–S13 nào nên chưa có số.

### 06/11 — máy quay mượt hơn (chủ dự án: "góc quay hơi giật, nhất là qua vật cản")

- **Nhịp khung đều** (`client/src/render/framepace.ts`): bỏ `ticker.maxFPS = 60` trên máy
  tính — Pixi làm tròn xuống nên bỏ khung trên màn 60 Hz, và vẽ lệch nhịp trên 120/144 Hz.
  Nay đo tần số quét, vẽ đúng mỗi n khung quét (60 Hz → 60, 120 → 60, 144 → 72). Điện thoại
  giữ trần 45 như cũ.
- **Nhân vật của mình**: mỗi khung chỉ nội suy thẳng giữa hai bước dự đoán (không chạy lại
  phép trượt va chạm với bước lẻ — nguồn rung sát vật cản); chia theo nhịp gửi THẬT.
- **Máy quay đuổi theo** nhân vật (`player.camera_follow_ms` = 70 trong `config/game.json`)
  thay vì dính cứng; nhảy xa (cổng, hồi sinh, đổi khu) thì đặt thẳng.
- Đo: `node tools/browsertest/wallhug.mjs http://127.0.0.1:9998` (phần `mayQuay`) — PERF_LOG #30.

### 05/11 — bong bóng mặt (§12.5 v6.38, đóng Q44)

Hướng A "một thân nhiều mặt" bỏ hẳn. Thay bằng **bong bóng mặt**: dùng lại 4 chân dung
biểu cảm §12.4 **sẵn có** làm emote nổi trên đầu 1,5 s khi tâm trạng đổi — `chase` → cười
nham hiểm, `flee` → hoảng, ngủ/mù → hoảng, hạ được người → cười. Viền trắng, đuôi chỉ xuống,
nảy khi hiện, cooldown 3 s/người, ẩn khi Ẩn Thân/trong bụi, tắt ở chất lượng Thấp khi > 15
người. **Không cần art mới.**

Đọc tâm trạng thẳng từ snapshot (`md`, `st`) chứ không nối thêm event; "hạ được người" đi
qua `KillFeed` có nạn nhân — `KillFeed` còn dùng cho đột phá và Thiên Môn nên phải lọc bằng
`e.b`, không thì ai đột phá cũng cười nham hiểm một cái.

Hai lỗi tự gây ra và tự bắt khi xem ảnh:
- **Trộn đơn vị**: vẽ bán kính `22 * zoom` trong khi quy ước của `Nameplate` là vẽ cỡ pixel
  cố định rồi `scale = 1/zoom`. Bong bóng teo mất ở mức phóng thường gặp — 24 ảnh chụp đầu
  tiên không dính cái nào dù log báo đã bật.
- **Đè lên bảng tên**: đặt lệch phải vẫn che mất tên người chơi. Nay đo mép trên của cụm
  bảng tên rồi đặt lên trên hẳn.

Art thử 54 tệp (`body/` + `faces/` của 3 nhân vật) chuyển vào `assets/library/q44_than_mat_roi/`
theo luật 13, INDEX ghi rõ lý do. `bodyface.go` giữ lại nhưng `abandoned = true` nên `artgen
plan` không còn liệt 6 job của hướng đã bỏ — prompt vẫn còn đó nếu sau này làm rig cắt mảnh 2D.

Công cụ mới `tools/browsertest/emoteshots.mjs`: bong bóng sống 1,5 s nên chụp theo nhịp cố
định là trò may rủi; nay `emote.ts` báo một dòng `[EMOTE]` ra console khi `?debug=true` và
bài chụp bấm máy ngay khi nghe thấy. Ảnh: `docs/review/emote/` (3 PC + 3 iPhone).

### 05/11 (chiều) — khung tròn chỉ thấy cái nón: khung đầu riêng cho từng chân dung

Chủ dự án chụp màn hình: ô đại diện HUD (§12.4) và bong bóng mặt (§12.5) chỉ thấy **cái nón
với hai con mắt**. Cả hai cắt ảnh bằng **một luật chung** `background: center top / 180%` —
luật đó chỉ đúng nếu mặt nằm giữa và chiếm gần hết mép trên, mà bộ chân dung thật là ảnh
bán thân/toàn thân, đầu nhỏ và lệch. Spec §12.4 ghi "cắt từ đầu tới vai", luật cũ không làm
được việc đó.

Nay mỗi chân dung có **khung đầu riêng** `[trái, trên, cạnh]` trong
`client/src/render/headbox.ts`, sinh bằng `node tools/portrait-headbox.mjs` (thêm `--sheet
<ảnh.png>` để có tờ ảnh kiểm). Tool dò **hai tầng**: alpha tìm CỘT của đầu (chỗ hẹp nhất
phía trên thân là cổ), rồi **màu da** tìm MẶT trong cột đó — gộp mọi mảng da, nới lên 42%
cho tóc/nón và xuống 10% cho râu/cằm. 34/40 ảnh máy dò đúng; 6 ảnh toàn thân đo bằng mắt
trong `OVERRIDE` (`ong_ba_vong/evil`, `shipper/evil`, cả 4 biểu cảm của `a_hon` — nhân vật
hở da gần hết người nên vùng da ôm trọn cả người).

Hai hướng sai phải bỏ mới ra được cái trên, cả hai đều đã lọt tới tay chủ dự án: **(a)** chỉ
dùng alpha rồi đoán cằm theo cổ — mặt có râu/hàm bạnh thì chỗ hẹp nhất là quai hàm NẰM TRÊN
cằm, khung cắt ngang mồm; **(b)** bù bằng "đầu cao 1,35 lần bề ngang" — ảnh nón to thì khung
phình ôm cả người, ba ảnh Cô Ba ra đúng `[0, 0, 1]`. Vùng da thì không phải đoán.

**Bài học về cách duyệt**: tờ ảnh kiểm 120 px hình VUÔNG nói dối hai lần. Khung tròn bo mất
bốn góc nên ô vuông trông ổn mà trong game vẫn cắt; và ở 120 px thì chính cái cằm bị cắt
cũng không nhìn ra. Giờ duyệt bằng hai bài chụp thật: trang DOM dùng chính `style.css` với
đủ 5 cỡ ô, và trang Pixi nạp THẲNG lớp `Emote` thật. Chi tiết trong báo cáo.

Bốn nơi dùng chung `headCrop`/`headBox`, không ai tự chép công thức: ô đại diện HUD (§12.4),
bong bóng mặt (§12.5), chân dung tròn bảng xếp hạng 28–40 px (§11.5), ảnh nhân vật màn chọn
64 px + khung xem trước 84 px. Hai chỗ sau trước đây thu **cả tấm ảnh** vào ô nên mặt bé bằng
hạt gạo. Màn Tổng kết và màn tải giữ nguyên cả tấm ảnh — ở đó cố ý hiện nguyên nhân vật.

Trước/sau + cách làm: `docs/reports/2026-11-05-cat-chan-dung/`.

### 06/11 — spec v6.40: bán kính nhặt tuyến tính · chết rơi kiếm theo sổ 10 giây

**1. Bán kính nhặt = `90 × scale × (1 + Σ thưởng %)`** (§4.0, §6.0). Trước là
`90 × √scale` với thưởng tính trên 90 cố định. Hai thay đổi đi kèm: base ĐÃ nhân cảnh giới
nên mọi khoản thưởng là **% của bán kính theo cảnh giới** (Nam Châm đáng 45 px ở Luyện Khí,
270 px ở Đạo Tổ), và trần đổi từ ×4 thành **×2**. Config gom vào khối mới
`pickup { base, scale_mode, cap_mult }` — `player.pickup_r`, `stats.pickup_cap`,
`exp_economy.pickup_scale_exp` bỏ hẳn, không để hai nguồn số. Nam Châm +0.2 → **+0.5**.
Con số nghiệm thu §17 khớp đúng: Luyện Khí 1 = 90 px, Đạo Tổ = 540, Shipper Đại Thừa 10 +
Gom Đơn + Nam Châm = 720, Thằng Mập Đạo Tổ + Nam Châm = 891.

**`orb_value_mult` 1.5 → 1.32** (§6.0 luật 2 yêu cầu chỉnh lại sau khi đổi bán kính). Hút xa
hơn ở cảnh giới cao làm farm nhanh gấp ~2,4 lần ở Đại Thừa; 1.32 kéo về đúng bảng nhịp —
`cmd/sim --farm` xanh cả 9 cảnh giới, hai lượt liên tiếp (cộng dồn 27,3 và 27,6 phút so với
mốc ~28).

**Q46 — đã chốt (06/11)**: `cmd/sim --farm --magnet-always` trượt luật "≤ 1,5 lần nhanh hơn
bảng" ở 4–6 cảnh giới thấp tuỳ seed (Luyện Khí 1,55× · Trúc Cơ 1,65–1,81× · Kim Đan
1,65–1,70× · Nguyên Anh 1,60–1,65× · Hóa Thần và Luyện Hư ~1,58×). Không chữa được bằng
`orb_value_mult`: hệ số đó vào theo `m^r` nên ở Luyện Khí `m^0 = 1`. Chủ dự án **chấp nhận
khoảng đo được** → trần nới **1,5 → 2,0**, và chuyển từ hằng số trong `cmd/sim/farm.go` ra
`exp_economy.magnet_max_speedup` (luật 2: số cân bằng nằm trong config). Chọn 2,0 chứ không
1,8 vì Trúc Cơ dao động 1,65–1,81× giữa các seed — 1,8 đỏ ngay lượt chạy thứ hai. Nay cả
hai bài xanh, exit 0; bài "Luyện Hư 1 → Chân Tiên ≥ 15 phút" đo 17,2–17,9 phút.

**Spec còn ghi 1,5** ở §6.0 và §17 (`magnet_max_speedup` trong khối config §17 cũng ghi
1.5). Claude không sửa `docs/_raw/` — chủ dự án cập nhật hai chỗ đó khi tiện.

**2. Chết rơi kiếm theo sổ cửa sổ trượt** (§7). Trước: `floor(kiếm_còn × 0.5)` — mài sạch
kiếm người ta rồi giết thì xác gần như trống. Nay mỗi người có sổ **"kiếm BỂ trong 10 giây
gần nhất"**, chết thì rơi sổ đó **cộng số kiếm còn trên người**. Kiếm bị **cướp** (Hút Mì,
Trộm Kiếm) không vào sổ — số đó đã sang vòng người khác, ghi nữa là nhân đôi kiếm trong
phòng. Trần tổng = kiếm tối đa của cảnh giới nạn nhân; chia đều **tối đa 12 bó** (bó có thể
> 10 cây — Đạo Tổ 1 050 kiếm mà chia 10 cây/bó là 105 vật phẩm quanh một cái xác).

6 bài test mới trong `death_drop_v640_test.go` phủ đúng hai dòng §17: kịch bản Kim Đan
60 kiếm mài mất 50 trong 8 s, kiếm ngoài cửa sổ không tính, Hút Mì và Trộm Kiếm không tính,
trần cảnh giới, và 4 con số bán kính nhặt của spec.

### 06/11 (chiều) — quái chiếm bao nhiêu % EXP farm · `monster_value_mult` hoá ra không tồn tại

Câu hỏi: `exp_economy.monster_value_mult` vẫn 1.5 trong khi orb đã 1.32 — quái có bị lệch
hệ số không?

**Không có gì lệch, vì khoá đó không tồn tại.** `grep` cả `config/game.json`, toàn bộ Go và
client: không chỗ nào đọc `monster_value_mult`. EXP quái đi qua đúng `OrbValueMult`
(`combat.go:597` → `scatterOrbs(m.Exp × OrbValueMult(killer))`), tức đã là **1.32** từ lúc
đổi v6.40 — đúng §6.0 luật 4 "cùng hệ số, không riêng". Con số 1.5 chỉ còn trong **khối
config minh hoạ §17** của spec, là chữ cũ sót lại chưa bao giờ được cài.

Vẫn đo như yêu cầu, vì "không có khoá" chưa chứng minh được "quái không đáng kể".
`cmd/sim --audit` nay quét **đủ 9 cảnh giới** thay vì lấy mẫu 3 cái (0/5/8) — câu hỏi kiểu
này sẽ còn lặp lại mỗi lần đổi hệ số EXP.

| Cảnh giới | % EXP từ quái (không Nam Châm) | (có Nam Châm) |
|---|---:|---:|
| Luyện Khí | 5,6% | 3,7% |
| Trúc Cơ | 11,1% | 4,3% |
| Kim Đan | 6,5% | 11,5% |
| Nguyên Anh | 5,9% | 6,0% |
| Hóa Thần | 4,7% | 6,9% |
| Luyện Hư | 5,8% | 5,8% |
| Hợp Thể | 6,9% | 4,6% |
| Độ Kiếp | 6,9% | 4,6% |
| Đại Thừa | 4,7% | 3,1% |

Cao nhất **11,5%**, dưới ngưỡng 15% ở mọi cảnh giới → **không chỉnh gì**. Gần như toàn bộ
EXP farm đến từ viên môi trường (88–96%); quái chỉ là gia vị.

**Q47 — đã chốt (06/11)**: spec v6.41 ghi ngưỡng Nam Châm ≤ 1,75×, nhưng chạy 6 lượt thì
Trúc Cơ lên 1,83× và Kim Đan 1,86× — đặt 1,75 là bài `--magnet-always` đỏ tuỳ seed (đã thử,
trượt Kim Đan hai lượt liên tiếp). Chủ dự án **chấp nhận số đo**: spec v6.42 sửa ngưỡng
thành **2,0×** cho khớp config, và bỏ luôn khoá minh hoạ `monster_value_mult` chưa bao giờ
tồn tại. Không đổi luật chơi, không đổi config — chỉ spec đuổi theo thực tế.

**Xoá `docs/_raw/CLAUDE.md`** (06/11): bản cũ hơn bản gốc ở gốc repo — còn ghi cổng 8080,
spec ở `docs/GAME_SPEC.md`, thiếu luật 10 (hỏi khi spec chưa rõ), 11 (không chờ art) và 15
(không commit dữ liệu người chơi). Nó đã lọt vào ngữ cảnh một phiên chat và suýt được làm
theo. Nay chỉ còn `CLAUDE.md` ở gốc repo; dòng cảnh báo trong đó viết lại cho đúng.

### 06/11 — vá `docs/perf_log.csv`: 3 dòng output `grep` lọt vào tệp dữ liệu

Gộp stash số đo cũ thì lộ ra tệp đang hỏng. Dòng 62–64 không phải dữ liệu mà là output của
`grep` bị ghi đè vào: `1 matches in 1 files:`, một dòng trống, và một dòng bị cắt cụt
`61:0:30 → 05/11 12:00",S6,0.48,2.22,…` kết thúc bằng dấu `...`. Lọt vào từ commit
`362258a`.

Vá bằng cách đối chiếu **báo cáo gốc** (`docs/perf/real/*.md` §2 có sẵn dòng CSV đầy đủ),
không đoán: khoảng 04/11 23:30 → 05/11 12:00 đúng là `0.48 / 2.22` (N=1) như báo cáo ghi.
Dòng cùng khoảng trong stash (`0.92 / 5.93`) là bản sinh cũ hơn, bỏ. Dòng
05/11 12:00 → 05/11 23:30 (`1.20 / 7.80`, N=9) khớp đúng báo cáo `2026-11-05_2330.md` nên
thêm vào. Nay 11 dòng S6 đều phân tích được bằng `csv`, đúng thứ tự thời gian.

**Rút ra**: `docs/perf_log.csv` là tệp dữ liệu, đừng bao giờ ghi vào bằng `>>` sau một lệnh
có đường ống — một lần nhầm là hỏng im lặng, và chỉ lộ ra khi có người đọc kỹ ba tháng sau.

### 06/11 (tối) — xếp hạng theo CẢNH GIỚI trước · bỏ điểm giữ Top 1

**1. Thứ tự xếp hạng** (chủ dự án chốt): cảnh giới/tầng trước, hoà thì Uy Danh, hoà nữa thì
tổng EXP. Áp dụng **toàn bộ**, không chỉ mặt bảng: danh hiệu Top 1–10, vương miện, và hệ số
lật kèo khi chém người hạng cao đều đi theo thứ tự này. Trước đây Uy Danh đứng trước.

Phép so sánh gom về **một hàm duy nhất** `game.RankLess`. Trước đó nó bị chép tay ở **ba**
chỗ — `game.Ranking`, `room.Ranking`, và `game.RankOf` — và đúng như rủi ro của kiểu chép
này, `RankOf` đã lệch ngay khi đổi luật: `Ranking()[0]` chỉ vào một người còn
`RankOf(...) == 1` chỉ vào người khác. Bài test mới bắt được ngay.

**2. Bỏ `top1_per_sec`** ("ai cũng giống nhau"). Người dẫn đầu không còn được +1 Uy Danh mỗi
giây. Đó là vòng lặp tự nuôi: lên đầu rồi thì càng đứng lâu càng khó bị đuổi kịp mà không
phải đánh thêm trận nào. Nay Uy Danh chỉ đến từ việc làm — chém trúng, hạ người, đột phá.
`top1_secs` vẫn **đếm** (hiện ở màn Tổng kết, lưu trong sổ), chỉ không quy ra điểm nữa.

Bỏ khỏi: `config/game.json`, `config.Score.Top1PerSec`, hằng `ScoreTop1`, kiểu config của
client. Màn Tổng kết không cần sửa — nó duyệt `score_by` chung nên khoá biến mất là dòng tự
biến mất.

**Spec cần anh sửa** (Claude không ghi vào `docs/_raw/`): §11.1 bảng nguồn điểm còn dòng
"Giữ Top 1 phòng | +1 điểm mỗi giây" (dòng 1096); §11.3 dòng 1106 và 1124 còn ghi xếp hạng
và danh hiệu theo Uy Danh; khối config §17 dòng 1611 còn `"top1_per_sec": 1`.

### 06/11 (đêm) — rà §11 theo spec v6.43: bắt được lỗi bảng tự mâu thuẫn

Đối chiếu từng dòng §11.1–§11.5 với code. Khớp hết, trừ **một lỗi do chính thay đổi xếp
hạng hôm nay sinh ra**:

`heldBoard` giữ **thứ tự** bảng 3 giây để chống nhấp nháy, nhưng `lbRows` đọc **cảnh giới
tức thời** từ `p.Level`. Trước v6.43 không sao vì cảnh giới chỉ là chữ trang trí trên dòng.
Nay nó là **khoá sắp xếp chính**, nên ai vừa đột phá sẽ hiện cảnh giới mới ngay trong lúc
vẫn nằm ở hạng cũ — bảng đọc như sai suốt 3–6 giây (hạng 5 cảnh giới cao hơn hạng 4).

Đây đúng là lỗi chủ dự án chụp được hôm 1/11 với cột điểm ("dòng 8 = 75 đứng trên dòng
9 = 95"), chỉ đổi cột. Lần đó chữa bằng cách chụp điểm cùng nhịp với thứ tự; nay chụp thêm
**cấp**. `heldBoard.snapScores` → `snap(score, level)`, `lbRows` lấy cảnh giới từ bản chụp.
Test `TestHeldBoardSnapshotsLevelWithOrder` canh đúng ca đó.

Mấy mục còn lại đã đối chiếu và **đúng**, ghi lại để lần sau khỏi dò lại:

| §11 | Kiểm |
|---|---|
| 11.1 công thức điểm | `hit_base 100 × (r_nạn_nhân+1) × lật_kèo + kill_bonus 50`; ví dụ 850 và 150 của spec khớp |
| 11.1 bot, chống cày | `bot_victim_mult 0.5`; chặn cùng `guest_id` trong 60 s |
| 11.1 bỏ Top 1/giây | hằng `ScoreTop1` và khoá config đã xoá; `top1_secs` vẫn đếm |
| 11.2 bảng phòng | top 10, 1 lần/giây, chung cả 5 khu, một hàm `RankLess` |
| 11.2 bảng toàn cục | **vẫn xếp theo Uy Danh** — `ORDER BY score/hits/dai_thua_secs`, không dính thay đổi xếp hạng phòng |
| 11.2 lưu trữ | đủ cột; `GET /api/leaderboard?period&board` top 50, cache 10 s; bot không ghi |
| 11.3 danh hiệu | 10 danh hiệu, giữ 3 s, kill feed `title_gain`/`title_lose` khi lên/mất Top 1 |
| 11.3 Ẩn Thân | `player.ts` ẩn cả bảng tên (gồm danh hiệu) khi tàng hình hoặc trong bụi |

**Lệch tên, không phải lệch luật**: §11.2 ghi cột `secs_to_dai_thua`, CSDL đặt tên
`dai_thua_secs`. Chỉ là chữ trong spec, không ảnh hưởng gì.

### 07/11 — spec v6.44 (bỏ miễn mẻ của Trộm Kiếm) + v6.45 (Gom Đơn ×5)

**v6.44 — Trộm Kiếm**: bỏ "kiếm Ông Ba không mẻ". Va kiếm nay tính bình thường hai chiều,
chỉ giữ việc kiếm địch bể bay về vòng ông. Trong `stealOnClash` thứ tự có chủ ý: **trừ kiếm
của ông trước rồi mới cộng số cướp được** — cộng trước thì ông đang đầy vòng bị `giveSwords`
cắt phần dư rơi xuống đất, xong mới mẻ, thành ra vừa rơi đồ vừa mất kiếm.

Kiểm lại bằng test: cùng cảnh giới + cùng nội tại độ bền thì hai bên mẻ bằng nhau, nên ông
**hoà vốn** (mẻ N, cướp N) chứ không tăng kiếm — cái lợi nằm ở chỗ đối thủ lỗ trắng N thanh.
Dòng checklist §17 viết "ông lãi ròng" nên hiểu theo nghĩa tương đối đó.

**v6.45 — Gom Đơn ×5**: `speed.gom_don` 2.0 → 4.0, `cap_gom_don` 3.0 → 5.0 (975 đv/s).
Sửa `speed_test.go` (×3 → ×5) và `skills_v55_test.go` (trần 3 → 5).

**Nghiệm thu — hai mục ĐẠT, một mục TRƯỢT nhưng không phải lỗi mới:**

| Mục | Kết quả |
|---|---|
| Va chạm client↔server | ✅ 1 000 bước, **lệch lớn nhất 0,0000**, quá ngưỡng 0 |
| Entity nhấp nháy / AOI | ✅ 0 lần biến mất, 0 lần dựng lại trong 20 s |
| Không giật, không nhảy cóc | ❌ 23/1 201 khung vẽ vượt trần tốc — **xem Q48** |

Bài đo mới `tools/browsertest/gomdon-smooth.mjs`: hai tab, lấy mẫu vị trí VẼ RA mỗi khung,
chuẩn hoá theo thời gian khung rồi so với trần tốc. Chạy ba kịch bản mới đọc được kết quả:
đi bộ **0** khung hỏng · Gom Đơn cấu hình **cũ ×3: 40** khung · cấu hình **mới ×5: 23** khung.
Tức **v6.45 không gây ra lỗi này**, nó có sẵn và chưa ai đo. Nghi kho nội suy cạn
(`interp_ms` 100 với `snapshot_hz` 20 chỉ dư đúng một gói); khung đứng hình và khung nhảy
cóc đi theo cặp. Ba hướng xử lý ghi trong Q48, chưa đụng vào vì đó là chỉnh netcode.

**Hai lần đo đầu tôi làm sai phương pháp**, ghi lại để khỏi lặp: lần một so bước đi với
trung vị chung — nhưng Gom Đơn chỉ bật 4 giây trong 20 giây nên trung vị là tốc ĐI BỘ, mọi
khung chạy nhanh đều bị chấm là nhảy cóc (ra 260/1 200). Lần hai vẫn không chuẩn hoá theo
thời gian khung — một khung kẹt 100 ms thì đi 97 đơn vị là ĐÚNG. Chỉ sau khi có **mốc so
sánh không bật kỹ năng** (0 khung hỏng, trung vị 185 đv/s đúng tốc đi bộ) mới tin được số.

### 07/11 (chiều) — Q48 bước 1 xong, bước 2 hoàn tác · lịch dọn Docker hàng tuần

**Lịch dọn Docker**: `crontab` của người dùng, 04:00 Chủ nhật hàng tuần —
`docker system prune -af --filter "until=168h"`, log ở `~/.local/state/docker-prune.log`.
Chỉ xoá image không container nào dùng và quá 7 ngày; **không đụng volume**. Runner CI tự
host bơm ~45 GB image không dùng mỗi hai ngày, đó là thứ làm đầy ổ hai lần trong tuần.

**Q48 bước 1 — `interp_ms` 100 → 150**: số khung nhảy cóc giảm 5–10 lần (2,67–3,83% → 
0,25–2,23%) nhưng **chưa về 0**, nên Q48 vẫn mở. Chi tiết số đo ở PERF_LOG lượt #31.

**Q48 bước 2 — hoàn tác.** Viết đúng mô tả (ngoại suy tối đa 100 ms, kéo mượt 100 ms, nhảy
thẳng khi lệch > 300 đv, áp cho người khác + quái) nhưng hai bản đều tệ hơn nền: bản 1 ra
16,5% khung nhảy cóc vì gieo khoản lệch mỗi gói (20 lần/giây) nên entity đứng lệch vĩnh
viễn và chạy đuổi; bản 2 sửa chỗ gieo thì entity đứng im hẳn. Hoàn tác thay vì đẩy netcode
nửa vời lên server.

**Hai cái bẫy của quy trình, ghi lại để khỏi lặp:**
- Đổi code client thì **phải `npm run build`** trước khi đo — server dev chỉ phục vụ
  `client/dist`. Hai lượt đo đầu của bước 2 chạy trên bundle cũ, ra số y hệt nền, suýt kết
  luận "bước 2 không ảnh hưởng gì".
- Bài đo phải giữ mục tiêu **trong khung nhìn**: để Shipper chạy 2 giây liền ở 975 đv/s là
  đi 1 950 đơn vị, 84% số khung không thấy ai. Nay mỗi lượt chạy 0,7 giây và nạp lại lượt
  kỹ năng bằng phím `G` để gần như cả cửa sổ đo là chạy nhanh.

### 10/11 — Q48 bước 0 (đo) + bước 1 (xong) · bài đơn vị nội suy · cổng trước deploy

**Bước 0 nối được nhân quả**: thêm ba con số chẩn đoán vào `gomdon-smooth.mjs` — kho nội
suy còn dư, đồng hồ nhảy, khoảng cách gói của chính entity đang xem. **80% khung đứng hình
trùng đúng lúc kho cạn**; khoảng cách gói trung vị 66 ms (không phải 50), có lúc 167 ms.

**Bước 1 xong**: người chạy nhanh hơn 2× tốc gốc luôn gửi đủ nhịp, kệ khoảng cách
(`fastMover` trong `broadcast.go`, dùng `MoveSpeed` thật nên kỹ năng mới sau này tự vào
luật). Gói lớn nhất 167 → **67 ms**, gói cách > 75 ms: 16 → **0**, nhảy cóc còn
**0,66–0,75%**. Đứng hình vẫn còn 24–34 khung → chưa đạt, Q48 còn bước 2–5.

**Bài đơn vị trước, sửa sau** — `npm run check:interp`: bơm gói giả (dao động ±30 ms, mất
5%, đồng hồ trôi 2%, đổi tốc ×1 ↔ ×5) rồi canh quỹ đạo vẽ. Hiện đỏ đúng chỗ cần đỏ (3 khung
đứng hình). `followSnaps` tách ra `client/src/net/interp.ts`. Đây là thứ còn thiếu ở hai lần
thất bại hôm 07/11 — sửa thẳng trong `state.ts` rồi thử bằng trình duyệt, mỗi vòng 10 phút
và không thấy trạng thái bên trong.

**Cổng trước deploy**: `gomdon-smooth.mjs` vào `tools/deploy.sh` bước **2b**, ngưỡng tạm
≤ 3%. Đặt sau bước dựng client — lần đầu tôi đặt trước, tức là kiểm bản cũ, vô nghĩa.
`SKIP_SMOOTH=1` để bỏ qua trên máy không có Playwright.

### 10/11 (chiều) — Q48 bước 2 viết xong, CHƯA ráp vào client · bài đo tự biết lúc máy quá tải

**`Clock` (bước 2) nằm ở `client/src/net/interp.ts`, đã chứng minh trong bài đơn vị, nhưng
`state.ts` CHƯA dùng.** A/B trong `check:interp` — cùng chuỗi gói, chỉ đổi đồng hồ:

| Đồng hồ | Khung nhảy cóc | Tốc vẽ lớn nhất |
|---|---:|---:|
| Cũ (lấy ước tính tốt nhất, nhảy thẳng) | **15** | 2 194 đv/s (2,25× trần) |
| Mới (chỉnh tốc độ ≤ 8%, không nhảy) | **0** | 1 053 (1,08× trần) |

Chạy lại bằng `OLD_CLOCK=1 npm run check:interp` để so.

**Vì sao chưa ráp**: ráp vào rồi đo trên trình duyệt thì chập chờn — 2/5 lượt tốt, 3/5 lượt
kho nội suy âm 1,7–16 giây. Tìm được một nguyên nhân thật và đã sửa (đồng hồ khởi tạo từ
ước tính của ĐÚNG MỘT gói đầu tiên, lệch 1,07 s, rồi bò về đích 8%/giây mất 13 giây → thêm
cửa sổ "bắt sóng" 2 giây đầu bám thẳng đích). Nhưng vẫn còn chập chờn, nên trả `state.ts`
về bản đang chạy.

**Rồi phát hiện ra lý do thật sự của chập chờn: MÁY QUÁ TẢI.** `load average 39–47 trên máy
12 nhân` (có tiến trình khác ngoài việc của tôi). Bằng chứng không cãi được: đo lại **chính
commit đã deploy** trên máy đó ra 0,92% rồi 13% rồi 18,8% khung nhảy cóc — cùng một mã.

Nên **mọi số đo trình duyệt của lượt này từ bước 2 trở đi không dùng được**, kể cả kết luận
"bước 2 tệ hơn". Cái dùng được là bài đơn vị, vì nó chạy trong tiến trình, không phụ thuộc
nhịp khung.

Hai chốt thêm vào để lần sau không sập lại:
- `gomdon-smooth.mjs` **tự từ chối kết luận** khi `loadavg/số nhân > 1.5`, thoát mã 2. Đo
  tải bằng `loadavg` chứ không bằng thời gian khung của tab xem: lúc load 39 thì khung vẫn
  ~16,7 ms (headless vẽ nhẹ) nên nhìn vào khung không thấy gì.
- Cổng trong `tools/deploy.sh` **lấy đúng mã thoát của bài đo**. Trước đó tôi viết
  `node ... | tail -2` trong `if`, tức là đọc mã thoát của `tail` — cổng luôn xanh dù bài đo
  đỏ. Mã 1 chặn deploy, mã 2 (máy quá tải) chỉ cảnh báo rồi đi tiếp.

### 10/11 (tối) — tối ưu máy yếu, bước 0: dựng xong bộ đo, CHƯA có số sạch

Anh dặn "báo bảng này TRƯỚC khi sửa", nên lượt này chỉ dựng bộ đo, không sửa gì về hiệu năng.

**`client/src/render/probe.ts`** — bấm giờ TỪNG HỆ THỐNG trong vòng vẽ: nội suy · người ·
quái · vật phẩm · thần thú · bản đồ · ambient · vfx · kỹ năng. Hiện thêm hai dòng trong
bảng F3, kèm heap dùng và **biên độ răng cưa** (phẳng = không cấp phát trong vòng khung).
Bật theo trạng thái F3, tắt thì mỗi mốc chỉ là một phép so boolean — người chơi thường
không trả giá.

Khoá mốc để **ASCII**, nhãn tiếng Việt để riêng: `check:i18n` chặn chuỗi tiếng Việt ngoài
`ui/strings.ts` và nó chặn đúng — mấy chuỗi này trông y như chữ cho người chơi.

**`tools/browsertest/weakperf.mjs`** — in thẳng bảng anh yêu cầu: FPS + ms từng hệ thống ở
CPU 1× / 4× / 6×, heap và răng cưa, cộng phép thử nghẽn CPU hay GPU (hạ độ phân giải còn
0,5 rồi so FPS). Dựng sẵn phòng đông: thả 3 bot cạnh mình rồi kéo cảnh giới lên Đạo Tổ.
`scene.setResolution()` và `scene.fpsNow()` thêm vào chỉ để bài đo gọi.

**Chưa có số sạch**: máy chạy `load average 31–46 trên 12 nhân` suốt lượt (có `vitest` của
việc khác ăn 542% CPU). Bài đo tự từ chối kết luận ở mức tải đó. Chạy thử bằng
`IGNORE_LOAD=1` thì bảng ra đúng khuôn, nhưng con số là nhiễu máy — ví dụ hạ độ phân giải
còn 0,5 lại làm FPS **giảm** 30%, điều chỉ xảy ra khi máy đã nghẽn vì thứ khác.

Chạy lại khi máy rảnh:

```
cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9998 -debug=true &
node tools/browsertest/weakperf.mjs http://127.0.0.1:9998
```

### 11/11 — spec v6.46: Tiên giới 10 tầng · cờ `can_absorb_tien` · bảng xếp hạng ổn định

**1. Lỗi hiển thị Tiên khí ở Đại Thừa 10** — client đoán bằng CẤP (`level >= 89`) nên vừa
chạm Đại Thừa tầng 10 là viên Linh khí đã mờ đi, trong khi người chơi vẫn đang cần nhặt
chúng để nạp nốt thanh. Server vốn đã đúng (`BanksTienKhi` = `total_exp >= 694 960`); nay
gửi thành cờ `FlagCanTien` và client đọc cờ đó. Thiêu Đốt kéo xuống dưới mốc thì cờ tắt,
mọi thứ đổi ngược lại.

Cờ cũ `Flags` là `uint8` và 8 bit đã đầy ở `FlagBurning`. Nới thành `uint16` — **không đổi
gói tin**, vì giao thức nhị phân vốn đã chở `flags u16` (`proto/schema.txt`), chỉ struct
msgpack còn hẹp.

**2. Tiên giới chia 10 tầng mỗi cảnh giới** (§3.4). Tầng s nặng `1 + 0.08·s` như phàm giới;
scale `5.0 + 0.02 × (10k + s + 1)` → Chân Tiên 1 = 5,02 … Đạo Tổ 10 = 6,00. Vòng kiếm và
kiếm tối đa vẫn theo CẢNH GIỚI.

Mã hoá tầng Tiên vào chính `Level` (`TienLevel(idx, tier)`), dùng chung công thức
`Realm = L/10`, `Tier = L%10` với phàm giới — nhờ vậy bảng xếp hạng, danh hiệu, scale,
HUD tự hiểu tầng Tiên mà không phải sửa chỗ nào. Thanh Tiên khí nay là thanh của TẦNG
(Đạo Tổ cả cảnh giới rộng 5 000 điểm, nhích một viên không thấy gì động đậy). Lên tầng chỉ
đồng bộ `Level`; chỉ lên CẢNH GIỚI mới báo Thiên Kiếp. Thiêu Đốt rớt tầng nhưng sàn vẫn là
mốc khởi điểm cảnh giới nên không rớt cảnh giới.

Mốc tầng làm tròn **gần nhất**, không làm tròn lên: tròn lên lệch 1 ở Chân Tiên tầng 3
(238,2 → 239 trong khi spec ghi 154–238).

**3. Bảng xếp hạng chống nhảy** (§11.2): bỏ phá hoà bằng tổng EXP. Cùng tầng + cùng Uy Danh
→ `RankLess` trả `false` cả hai chiều, và mọi chỗ gọi đều `sort.SliceStable` trên
`PlayerList()` (thứ tự vào phòng) nên thứ tự giữ nguyên. Trước đây EXP đổi mỗi lần nhặt một
viên: 40 bot Luyện Khí cùng tầng, cùng 0 Uy Danh thì bảng đảo chỗ suốt ván. Luật "giữ 3 s
mới đổi" và "chết/rời rút ngay" đã có sẵn trong `heldBoard` từ 01/11.

4 bài test mới theo đúng 3 dòng checklist §17 (`tien_tiers_v646_test.go`). `go vet` sạch ·
`go test -race` 439 pass · `mapgen --check` 14/14 ĐẠT · build pass.

## Lỗi / hạn chế đã biết

- **ĐỨNG HÌNH giữa trận (MỞ, nặng nhất đang có)** — 12/11 một người chơi thật trên `:9999`
  đang 60 FPS thì game chết cứng, Chrome báo "Trang không phản hồi", phải tải lại trang. Lỗi
  từ hệ thống bộ lọc của PixiJS 8.22; một ngoại lệ trong ticker là vòng vẽ không bao giờ
  quay lại. Chưa dựng lại được → chưa biết nguyên nhân gốc. Đầy đủ:
  `docs/LOI_DUNG_HINH.md`; bài dò: `tools/browsertest/filtercrash.mjs`.
- Thiên Đình với sàn 30: p95 3,2–3,75 ms (S13), biên ~7% — PERF_LOG B21. 1/5 lần đo trượt ở
  Linh Sơn (4,74 ms, máy bận), chạy lại đạt.
- Lịch sử git còn tên/UUID người chơi + IP từ trước khi mở công khai (commit 0efc19c, 6cc900a,
  cd777f8, 510ccd5, 87d7cd6). **Đã làm sạch 02/11** (chủ dự án duyệt): filter-repo xoá mọi bản
  `docs/LEADERBOARD_AUDIT.md` cũ + 4 ảnh `docs/review/11-admin/{board,audit}_*.png`, thay chữ trong
  `server/cmd/hidelives/main.go`, thêm lại bản ẩn danh; force push → **mọi hash commit đã đổi**
  (bb1cd96 → 654b2fe, 5740e47 → 39805da…). Máy khác có bản clone cũ phải clone lại.
  `refs/pull/1..7` trên GitHub không chứa 5 commit có dữ liệu (đã kiểm). Còn sót: ai biết đúng
  hash cũ vẫn mở được trên GitHub tới khi GitHub dọn — muốn xoá ngay phải nhờ GitHub Support.
  Bản sao lưu trước khi làm sạch: `~/backup/bladeascend-mirror-2026-11-02` (khớp 5740e47).
  Khôi phục nếu cần: `cd ~/backup/bladeascend-mirror-2026-11-02 && git push --mirror --force <url>`.
- Kỹ năng: đàn em "chạy từ mép vào" và khung sạp "nhấp nháy trước khi bật" chỉ là lớp vẽ —
  server vẫn đặt đàn em/sạp ngay lập tức (không đổi luật). Tiếng kỹ năng là tiếng tạm tổng hợp.
- Icon trạng thái đang là vector; khi chạy job artgen `icon_status_*` thì ảnh thật tự thay.

- Tick p95 S1: #26 = 4,01 ms lúc máy bận → **#27 máy rảnh = 1,38 ms** (đạt, B20 đóng). Telemetry
  `:9999` vẫn thấy Thiên Đình p95 tới 6,6 ms vào giờ máy build song song — theo dõi trên dashboard admin.
- Trang `/en/`, `/th/`…: trình duyệt máy tự chọn ngôn ngữ theo `navigator.languages` — smoke test
  Playwright (tiếng Anh) phải thêm `?lang=vi`.
- Thẻ hồi chiêu tiếng Bồ "Recarga 30s · 2 cargas" dài nhất — vừa thẻ trên ảnh, cần xem máy thật.
- FPS điện thoại / bộ nhớ GPU iPhone thật **chưa đo** (B13b, B17) — cần máy thật.
- Telemetry staging ghi vào ổ container, mất mỗi lần deploy (cần `-telemetry-db /data/...`, hỏi trước).
- ~~Lớp mặt nhân vật §12.5 A~~ **ĐÓNG (Q44, spec v6.38)** — bỏ hướng A, thay bằng **bong bóng
  mặt**: dùng lại 4 chân dung §12.4 sẵn có làm emote nổi trên đầu 1,5 s khi tâm trạng đổi
  (`client/src/render/emote.ts`). Không cần art mới. 54 tệp art thử chuyển vào
  `assets/library/q44_than_mat_roi/` theo luật 13; `bodyface.go` giữ lại nhưng thôi phát job.
- 1 anim lặp còn vượt ngưỡng giật 45 (`docs/JITTER_REPORT.md`).
- Netcode bước 2/4/5/7 (Room Shared Cache, encode-once, IndexedDB, Service Worker) chưa làm — số đo hiện
  đạt ngân sách nên không phải nút thắt.
- Admin: dashboard còn thiếu CPU/KB/s/ping/FPS theo thiết bị; ô "xem bản đồ theo seed" cần API mới.
- `npm run check:telemetry-off` báo trượt từ trước (chuỗi `WEBGL_debug_renderer_info`), không nằm trong build.

## Hiệu năng — hai quyết định của chủ dự án 02/11

Sau khi tách telemetry theo build và loại phiên máy, danh sách ưu tiên hiệu năng rút lại:

| Mục | Quyết định |
|---|---|
| 0. Tách telemetry theo build | ✅ xong |
| 1. `RangeError` giải mã | ✅ xong — thiếu `try/catch` ở đường msgpack |
| 2. Respawn | ✅ điều tra xong, không tái hiện được → không sửa; đã thêm bảng nguyên nhân |
| 3. `addressModeU of null` | **BỎ** — chỉ xuất hiện trong phiên Playwright, không người thật nào dính. Ép WebGL cho mọi người là sửa một cái bóng |
| 4. Chất lượng mặc định / FPS | **HOÃN** — đợi vài ngày người thật chơi trên telemetry đã lọc rồi mới quyết |
| 5. Deploy + bảng trước/sau | chờ mục 4 |

Mục 4 hoãn vì nền số liệu cũ không dùng được: nó trộn bài đo 10 tab vào. Từ 02/11 telemetry
tự loại phiên máy, nên chỉ cần ít ngày là có nền sạch. Mục tiêu vẫn giữ: **điện thoại p50 ≥ 40,
p1 ≥ 25; PC p1 ≥ 30**. Hai phiên iPhone thật 31/10 cho FPS 11,5 — mẫu mỏng nhưng là tín hiệu
thật, nên đừng coi việc hoãn là coi nhẹ.

## Lượt 11/11 — spec v6.46 → v6.51 + Q48, Q49 (đã deploy `:9999`, đã push)

| Mục | Kết quả |
|---|---|
| **v6.46** `96cc3b8` | Cờ `can_absorb_tien` (hết lỗi Đại Thừa 10 chưa đầy thanh đã hiện Tiên khí); Tiên giới 10 tầng/cảnh giới; bảng xếp hạng bỏ phá hoà bằng EXP |
| **v6.47** `d9d3473` | Bảng xếp hạng KÍCH THƯỚC CỐ ĐỊNH. Lỗi ở cả hai phía: server gửi danh sách thiếu khi đang giữ 3 s (đo được 5 dòng trong phòng 13 người sống), client dựng lại DOM theo dữ liệu + `trimToScreen()` ẩn dòng theo độ dài tên (phòng 105 người chỉ hiện 5 dòng). Bài mới `rbstable.mjs` ghi 20 mẫu/giây, đã cắm vào cổng deploy |
| **v6.49** `e3bd837` | Q49: số dòng theo chiều cao màn hình (≥900 → 10 · 700–899 → 7 · <700 → 5), thu nhỏ sàn 0,85, bấm tiêu đề mở lớp phủ đủ 10 dòng |
| **v6.48 §6.3** `2022b27` `611b2cd` `fb8302f` `9c17c95` `9eef83f` `b8c6aaf` | Viết lại 9 kỹ năng theo chuẩn Ông Ba Võng 25 s × 2 (Shipper giữ nguyên): A Hớn hồi 10% + 20%/4 s · Bà Tám lấy hơi hút đồ rồi hét · Đại Ca đàn em 15% + nhặt đồ hộ · Thằng Mập mù + đảo điều khiển + khổng lồ ×2 · Cô Ba sạp bất hoại + đẩy + buộc vào sạp · Sư Phụ hút 25% MỌI địch 5 U · Bé Tí bung cơ · Cụ Rùa (v6.50/51) mai 2 s + khiêu khích |
| **v6.48 §6.6** `d24497e` `4d3ff63` | Mô tả kỹ năng sinh từ config (+ cổng CI `check:skilldesc` chặn số viết cứng ở cả 6 ngôn ngữ); lượt thích nhân vật (SQLite, `GET/POST /api/likes`, ❤ trên thẻ, nút ở Tổng kết, cột ♥ ở trang admin) |
| **Q48** `ec20f26` | Đồng hồ vẽ chỉnh tốc + trễ nội suy tự điều chỉnh. Hermite KHÔNG cần: "3,5% khung nhảy cóc" hoá ra là lỗi PHÉP ĐO (bài lấy mẫu lệch nhịp vòng vẽ). Đo lại bằng vết trong vòng vẽ: **0 nhảy cóc, 0 đứng hình**, 3 lượt máy rảnh. Cổng deploy hạ xuống 0/0 |

**Tối ưu máy yếu (B22) — XONG** `c2d5093`. Vòng kiếm thành container CHỈ XOAY (đặt kiếm
một lần theo toạ độ cục bộ, mỗi khung chỉ xoay container) và lớp trạng thái chỉ vẽ lại khi
đổi (không đổi thì tối đa 15 nhịp/giây). Đo bằng BỘ ĐẾM CÔNG VIỆC vì máy dev dùng chung:
đặt lại thanh kiếm 933–999 → **0** lượt mỗi khung; lớp trạng thái vẽ lại/lượt gọi 1,00 →
**0,27**. Chi tiết + bảng thời gian ở `PERF_LOG` lượt #28.

**Spec v6.52** `e1b1c3e`: Hút Mì chốt toả tròn (Q50), Trân Châu 3 s ở cả luật lẫn hiệu ứng
màn hình, Cô Ba bỏ vỏ chuối (Q51) — xoá luôn hệ thống "vệt trơn" vì Cô Ba là nguồn duy nhất.

**Không còn câu nào chờ anh** — Q48, Q49, Q50, Q51 đều đã chuyển sang `QUESTIONS_DONE.md`.

## Lượt 11/11 (chiều) — rà lỗi NGƯỜI CHƠI gửi về và sửa

> Đã deploy `:9999` (build `20261111_2011_52adf2e`) và đã push — commit `52adf2e`.

Đọc thẳng `server/data/telemetry.db` (bảng `events`, chỉ phiên `automated = 0`) thay vì đoán.

**Ba ngoại lệ JavaScript đều là chuyện cũ, đã sửa xong từ trước:** `null.texture` ở
`updateBody` (11 phiên, build 30/10) · `null.addressModeU` (3 phiên, build 31/10) ·
`RangeError` khi giải mã (1 phiên, build 31/10). Không build nào từ 01/11 trở đi còn ngoại lệ.

**Thứ còn sống trên build hôm nay là bộ nhớ GPU.** `gpu_pressure` và `webgl_context_lost`
xuất hiện trên chính `ec38aee`, và một phiên 55 phút ghi **1 431 MB / 13 736 texture** (trần
iOS là 256 MB — Safari không báo lỗi, nó giết tab). Ba phiên dài khác: 812–1 117 MB.

Nguyên nhân là **rò rỉ texture**, một gốc hai chỗ — *huỷ node không huỷ ảnh trên GPU*:

1. Mỗi ô nền 1 024 có một mặt nạ loang 128×128 riêng. Ô đi khỏi khung nhìn thì node bị huỷ
   còn ảnh ở lại. Đi lại trong bản đồ là bỏ quên ~90 tấm mỗi phút.
2. Đổi khu dựng `MapView` mới, cái cũ chỉ bị gỡ khỏi cảnh rồi bỏ mặc — telemetry đếm được
   **181 lượt đổi khu** một phiên.

Sửa: ô nền giữ kèm danh sách texture của riêng nó và huỷ đúng chừng ấy (không huỷ chung được
vì `TilingSprite` dùng chung ảnh tile); thêm `MapView.destroy()` gọi khi đổi khu.

Đo bằng kịch bản MỚI `TEXLEAK` (`tools/browsertest/texleak.mjs`, PERF_LOG §1) — **mọi kịch
bản cũ đều chỉ chạy 20–120 giây nên không cái nào thấy được**:

| | trước | sau |
|---|---|---|
| 5 phút | 118 → 625 texture, một chiều | – |
| 14 phút | – | 191 → 224, nhấp nhô 168–276, không xu hướng |
| nhịp | +91 texture/phút · +17,9 MB/phút | +2 texture/phút · −0,6 MB/phút |

Kèm theo: `?textrace=1` gom texture **theo nơi cấp phát** (`window.__texsites()`). Không có
nó thì chỉ biết "có rò rỉ" chứ không biết ở đâu — vết ngăn xếp lúc tải ảnh lên chỉ thấy ruột
PixiJS. Mặc định tắt, không tốn gì cho người chơi. Chi tiết ở PERF_LOG lượt #33 (nút thắt B23).

## Lượt 12/11 (sáng) — đọc số liệu TICK của người chơi thật

Không sửa gì, chỉ đọc: `server/cmd/tickquery` (mới) bóc bảng `server_samples` trong
`telemetry.db` — 25 920 dòng (mỗi khu × 10 s), 30/10 → 12/11, loại hết phiên do máy điều
khiển. Báo cáo đầy đủ ở **`docs/TICK_NGUOI_THAT.md`**, tóm tắt ở `PERF_LOG` lượt #34.

- **Tick rất thoáng:** khu đang có người thật, p95 trung vị **2,57 ms / 33,3 ms** ngân sách.
- **Nhưng chưa nói được gì về lúc đông:** đông nhất từng ghi là **3 người trong một khu**,
  87% số mốc chỉ có 1 người. Tải tick hiện nay là bot, không phải người.
- Thiên Đình đắt gấp ~2,8 lần khu thường (p95 4,80 so với 1,74–2,97) — sàn bot 30.
- Client giật **không phải tại server**: lúc người thật tụt < 45 FPS, tick cùng lúc 2,59 ms,
  còn thấp hơn mức chung 3,29.
- **Phép đo bị chính máy dev làm nhiễu:** cùng build, cùng người chơi, sáng nay 09:00–09:31
  máy rảnh p95 2,13 ms → 09:31–09:50 máy đang biên dịch Go **6,03 ms**. 160 dòng `tick chậm`
  trong log server đều nằm gọn trong 13 phút đó. Hai nút thắt mới: **B24** (khu đông có thể
  đã đắt lên từ 06/11 — phải đo lại trên máy rảnh mới dám kết luận) và **B25** (ngưỡng cảnh
  báo 3,5 ms nổ 37,9% thời gian).

## Lượt 12/11 (trưa) — điều tra tick Thiên Đình + bộ thu gom số liệu (ADMIN_SPEC §12)

> Đã deploy `:9999` (build `20261112_1052_286d086`).

**Tick 17,3 ms KHÔNG phải do game.** pprof 30 giây trên chính server đang chạy: không ai
chơi → 20 ms mẫu CPU (0,067%); 1 client + 110 bot → **1,88 giây, tức 6,26% của MỘT nhân**.
Cùng lúc đó cái máy ở **load 47 / 12 nhân, CPU 89,6%, swap 2 GB cạn** (một `vitest` của dự
án khác chiếm 425% CPU suốt buổi). Một tiến trình không tiêu CPU mà tick 17 ms thì nó đang
**chờ được xếp lịch**, không phải tính lâu. Nút thắt mới **B27**.

**Kỹ năng v6.48 vô can.** Bảng "tick đi đâu" mới (input · bot AI · di chuyển · kỹ năng · va
chạm · thế giới · chết · gửi gói, đo MỌI tick): cột kỹ năng **0,00–0,01 ms**. Trong kịch bản
nặng S13 (110 nhân vật, 10 500 kiếm): va chạm 0,98 ms (54%), gửi gói 0,57 (31%), bot AI 0,15.
B24 đóng.

**Sửa chỗ nóng nhất — bot quét cả khu mỗi tick.** `nearestFarmTarget` duyệt toàn bộ viên
(**~4 900 viên/khu**), toàn bộ quái và toàn bộ vật cản, cho mỗi bot mỗi tick ≈ 4,4 triệu
phép đo khoảng cách mỗi giây mỗi khu. Nay chỉ quét trong tầm bằng lưới ô có sẵn (kết quả
chọn mục tiêu không đổi). Kèm: bỏ `math.Hypot` ở `geom.Vec.Dist`/`Len` và A*.
**Thời gian CPU cho cùng một trận S13: 25,8 → 19,4 giây (−24,7%)** (đo bằng CPU chứ không
đồng hồ vì máy đang load 53). Đối chiếu va chạm client↔server sau khi đổi phép căn: lệch
**0,0000 đơn vị**. Chỗ tốn nhất còn lại là đường gửi gói → **B26**.

**ADMIN_SPEC §12 xong:** bộ thu gom phút/giờ/ngày (`metrics_minute/hour/day` +
`deploy_marks` trong `server/data/metrics.db`), lấy mẫu 1 giây, **chỉ ghi phút có người
thật**, p95 gộp được nhờ histogram; kèm load máy / CPU máy / RAM trống / GC / heap /
goroutine / kiếm trong khu. Chi phí đo được **0,0004 ms mỗi tick, 0 cấp phát** (trần §12 là
0,1 ms). Trang **Hiệu năng** trong admin: chọn chỉ số/phạm vi/phút-giờ-ngày/khoảng, dải
min–max + TB + p95 + đường ngân sách + vạch deploy, chồng được 2 chỉ số, bảng + xuất CSV,
thẻ "giờ tệ nhất 24 h", cảnh báo vượt ngưỡng ≥ 5 phút (tự nói "có thể do máy bận").
**Còn cần anh duyệt bằng mắt** (checklist §12.4 mục 3): tôi không đăng nhập được trang
admin nên chưa tự xem được biểu đồ thật.

## Lượt 12/11 (chiều) — lỗi ĐỨNG HÌNH của người chơi · **điều tra xong, CHƯA SỬA**

Không sửa gì, chỉ truy: một người chơi thật trên `:9999` đang 60 FPS thì **game đứng hình
hẳn**, Chrome báo "Trang không phản hồi". Báo cáo đầy đủ ở **`docs/LOI_DUNG_HINH.md`**.

- Lỗi ném ra từ **hệ thống bộ lọc của PixiJS** (`FilterSystem._setupFilterTextures` →
  `globalUniforms.push`): thứ Pixi lấy ra từ bể dùng lại của chính nó không phải một
  `UniformGroup` → trạng thái nội bộ Pixi đã hỏng từ trước khung hình đó.
- **Một lỗi là chết hẳn**: `Ticker.update()` của Pixi 8.22 không có `try/catch`, và khung kế
  tiếp chỉ được đặt lịch SAU khi `update()` trả về. Telemetry (chạy theo từng khung) tắt đúng
  giây đó; không có beacon đóng tab → tab nằm đó chết cứng.
- Ngay trước lúc chết **mọi số đều đẹp**: 59,9 FPS, khung p95 18,5 ms, 0 long task, texture
  274 MB (bản vá rò rỉ 11/11 đang chạy đúng). Người chơi còn lại không sao → không phải server.
- Đã loại: rò rỉ texture, server, lỗi `null.texture` 30/10, texture lọc khổng lồ, sóng xung
  kích lúc đột phá. **Chưa dựng lại được** — hai bài thử (`tools/browsertest/filtercrash.mjs`)
  chạy sạch trên đúng pixi 8.22.0.
- 941 dòng `tick chậm` trong log giờ 09–10 là **nhiễu do máy bận** (4,80 tải/nhân lúc kiểm,
  luật 16) — không rút kết luận gì từ đó.

## Lượt 12/11 (chiều–tối) — spec v6.53: chống sập, vào lại phiên, vệ sinh bộ nhớ, janitor

Làm theo đúng thứ tự chủ dự án giao. Mỗi mục một commit.

### 1. §13.8 chống sập — ĐÃ SỬA lỗi đứng hình của lượt trước

Đây chính là việc số 1 của danh sách "Việc tiếp theo" lượt trước (`docs/LOI_DUNG_HINH.md`).

- **Client** (`client/src/render/guard.ts` mới): mỗi hệ thống vẽ là một ranh giới. Lỗi →
  tắt riêng nó 2 giây; lỗi ≥ 3 lần/10 giây → tắt hẳn tới hết mạng. 14 ranh giới.
- Lỗi trong khâu VẼ của Pixi → bỏ khung + tắt mọi bộ lọc; lặp 3 lần/10 giây → **dựng lại
  renderer + cây cảnh** (không tải lại trang); dựng lại hỏng → tải lại + vào lại phiên §7.1.
  Phải **giành lại quyền gọi `render()`** từ `TickerPlugin` thì mới đặt được `try/catch`
  quanh nó — để nguyên thì lỗi ném ra từ trong ticker của Pixi, đúng như lỗi 12/11.
- Watchdog: 3 giây không có khung (tab đang hiện) → khởi động lại ticker; vẫn chết → dựng lại.
- Mỗi gói tin (`conn.receive`) và mỗi nút bấm (`ui/dom.button`) có try/catch riêng.
- **Server** (`server/internal/safe` mới): `recover()` ở bốn ranh giới — tick khu, gói tin
  một người, goroutine nền, request HTTP. Panic 1 tick → bỏ tick của KHU đó, 4 khu kia chạy
  tiếp. Panic ≥ 3 lần/phút ở một khu → cờ "suy giảm" lên dashboard. Panic khi xử lý gói →
  ngắt đúng người đó.
- Panic server lên trang "Lỗi người chơi" (loại `server_panic`), gom theo chữ ký.
- **Lỗi giả để nghiệm thu** (§17): `kx.fault('vfx', 3)`, `kx.faultDraw(3)` ở client;
  lệnh test `panic` n=0/1/2 ở server.

### 2. PixiJS 8.22 → 8.23

Hai bản vá đi thẳng vào phần trên: **#12259** (filter cảnh báo texture đã huỷ sau resize) và
**#12369** (texture lấy từ pool đúng resolution). Gói tăng 0,3 kB gzip.

### 3. §7.1 vào lại phiên khi F5 (`server/internal/room/resume.go`, `client/src/net/session.ts`)

Mã phiên 32 byte; mất kết nối thì nhân vật **ở lại thế giới 20 giây** — đứng yên, vòng kiếm
vẫn quay, **vẫn chém chết được**, không nhặt đồ. Nối lại kèm mã → đúng nhân vật cũ, cấp mã
mới. Tab thứ 2 chiếm nhân vật, tab cũ thấy "Bạn đang chơi ở tab khác". Chết trong lúc mất
kết nối → quay lại thấy màn Tổng kết. **Thay hẳn `pendingRejoin`**: mọi đường tải lại trang
giờ đi chung một cơ chế. Giao thức lên v7.

### 4. §13.9 vệ sinh bộ nhớ — tìm ra HAI rò rỉ THẬT

Bài kiểm mới (`server/cmd/leakcheck`) chạy 60 phút game bằng đồng hồ giả và bắt được ngay:

- **`itemGrid.remove` bỏ quên con trỏ ở ô cuối** (`c[:last]` hạ `len` nhưng mảng nền vẫn
  giữ con trỏ). Mọi viên EXP đã nhặt KHÔNG BAO GIỜ được thu hồi → heap server **+7 MB/phút,
  420 MB sau một giờ**. Đủ giết tiến trình trên VPS 1 GB.
- **`Manager.tryClose` bỏ quên con trỏ phòng** — cùng họ lỗi, nhưng một `*Room` là cả một
  thế giới ≈ 13 MB. Sau khi mọi phòng đóng, heap vẫn còn nguyên một thế giới.

Sau khi sửa, 60 phút / 2 703 lượt vào-ra: heap mỗi phòng **+2,9 %**, mọi thứ về mức nền khi
hết người. Thêm: sổ tài nguyên client (F3 + telemetry), trần pool chữ nổi, trần vết netcode
(40 000 → 1 200 số), trần hàng sự kiện telemetry, quét mục mồ côi mỗi 60 giây ở server.

### 5. ADMIN §13 janitor

Job 04:30, bảng giữ dữ liệu đúng spec, **gộp trước khi xoá** (chữ ký lỗi và thống kê mạng
chơi theo ngày giữ mãi), VACUUM tuần và chỉ khi 0 người online, trang "Dọn dữ liệu" có bản
xem trước chạy sẵn, cảnh báo dashboard khi dữ liệu > 1 GB.

### Ghi chú

- `TestBotFillToTarget` đỏ ngẫu nhiên ~1/25 lần từ lâu (hạt giống phòng lấy từ `time.Now()`);
  đã sửa bài test, 120/120 lần chạy đều xanh.
- Cổng deploy có thêm hai bài rò rỉ. Bài client chạy **30 phút thật** theo spec — hạ bằng
  `LEAK_MINUTES=8 tools/deploy.sh`, bỏ bằng `SKIP_LEAK=1`.

## Việc tiếp theo

> Chỉ để ở đây việc CHƯA xong. Mục đã chốt thì xoá ngay — một dòng cũ sót lại là lần sau có
> người (hoặc Claude) hỏi lại thứ anh đã trả lời rồi. Ngày 05/11 đã dính đúng vậy với ảnh OG.

1. **Xem lại cổng rò rỉ của deploy**: bài client chạy 30 phút THẬT mỗi lần deploy (spec
   §13.9 ghi "trước deploy, tự động"). Nếu thấy quá dài thì chốt một con số ngắn hơn và tôi
   sửa mặc định — đừng để nó thành thói quen gõ `SKIP_LEAK=1`.
2. Xem ảnh **bong bóng mặt** (`docs/review/emote/`, 3 PC + 3 iPhone) — Q44 đã đóng, code đã
   chạy, chỉ còn duyệt bằng mắt. Thấy chướng thì chỉnh `R` (cỡ) hoặc `LIFE` (thời gian hiện)
   trong `client/src/render/emote.ts`.
3. Điền email trong `SECURITY.md`.
4. Xem ảnh trang admin trên staging (`docs/review/11-admin/`) rồi tự đăng nhập thử.
5. Tìm người bản ngữ duyệt 5 bản dịch + danh sách tên bot (`client/src/i18n/NOTES.md`).
6. **Xem trang Hiệu năng mới** (`/admin/#perf`): kiểm checklist ADMIN_SPEC §12.4 mục 3 —
   chồng biểu đồ tick Thiên Đình với load máy, thấy vạch deploy, bấm Xuất CSV.
7. **Đo FPS/GPU trên điện thoại thật** — cũng là thứ mở khoá mục 4 của bảng hiệu năng
   (chất lượng mặc định, đang hoãn chờ số liệu người thật đã lọc).
