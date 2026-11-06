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

**Q47 mở**: spec v6.41 ghi ngưỡng Nam Châm ≤ 1,75×, nhưng chạy 6 lượt thì Trúc Cơ lên 1,83×
và Kim Đan 1,86× — đặt 1,75 là bài `--magnet-always` đỏ tuỳ seed (đã thử, trượt Kim Đan hai
lượt liên tiếp). Config giữ **2,0** như Q46 chốt; chờ chủ dự án chọn sửa spec hay hạ config.

## Lỗi / hạn chế đã biết

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

## Việc tiếp theo

> Chỉ để ở đây việc CHƯA xong. Mục đã chốt thì xoá ngay — một dòng cũ sót lại là lần sau có
> người (hoặc Claude) hỏi lại thứ anh đã trả lời rồi. Ngày 05/11 đã dính đúng vậy với ảnh OG.

1. Xem ảnh **bong bóng mặt** (`docs/review/emote/`, 3 PC + 3 iPhone) — Q44 đã đóng, code đã
   chạy, chỉ còn duyệt bằng mắt. Thấy chướng thì chỉnh `R` (cỡ) hoặc `LIFE` (thời gian hiện)
   trong `client/src/render/emote.ts`.
2. Điền email trong `SECURITY.md`.
3. Xem ảnh trang admin trên staging (`docs/review/11-admin/`) rồi tự đăng nhập thử.
4. Tìm người bản ngữ duyệt 5 bản dịch + danh sách tên bot (`client/src/i18n/NOTES.md`).
5. **Đo FPS/GPU trên điện thoại thật** — cũng là thứ mở khoá mục 4 của bảng hiệu năng
   (chất lượng mặc định, đang hoãn chờ số liệu người thật đã lọc).
