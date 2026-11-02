# Rà soát bảng xếp hạng — mạng chơi đáng ngờ

> **Bản công khai, đã ẩn danh** (CLAUDE.md quy tắc 15): tên người chơi thật thay bằng
> "Người chơi A, B…", `guest_id` chỉ giữ 6 ký tự đầu (hai `guest_id` LAN của B và C trùng
> 6 ký tự đầu `179264…` — phân biệt bằng nhãn A/B/C). Tên test của đội làm game (`DaoTo`,
> `Test01`, `iPhoneTest`…) giữ nguyên. Bản gốc đầy đủ chỉ nằm ở máy chủ dự án:
> `docs/private/LEADERBOARD_AUDIT.md` (không đưa lên git).

> Ghi theo yêu cầu ngày 31/10/2026. **Chưa ẩn, chưa xoá gì cả.** Tài liệu này là danh sách
> để anh xem rồi quyết. Nút "Ẩn khỏi bảng" đã có ở trang quản trị (`/admin/#board`) —
> nó chỉ ẩn, không bao giờ xoá, và bỏ ẩn được bất cứ lúc nào.

## 1. Trước hết: lỗi này là của server, không phải của người chơi

Cờ `-debug` trong `server/cmd/server/main.go` **mặc định là `true`**, và `tools/deploy.sh`
không bao giờ tắt nó. Ở mọi bản deploy **trước `20261031_1748_5321dd6`**, server cấp bảng
lệnh test cho **bất kỳ ai** mở game bằng `?debug=true` — kể cả người lạ trong mạng LAN.
Bảng đó có lệnh nâng cảnh giới tức thì.

Nên phần lớn những dòng dưới đây không phải "ai đó phá game". Nó là: cửa mở, có người đẩy
thử, cửa mở thật. Ba mạng chơi đạt **tầng 89 với 0 đòn đánh ra và 0 lần hạ ai** — không có
cách chơi nào cho ra con số đó; chỉ lệnh test mới cho. Mà lệnh test thì server đã tự đưa.

Cửa đã đóng từ `5321dd…` (ADMIN_SPEC §1.1): nay server tự quyết ai được debug — chỉ phiên
đăng nhập quản trị, hoặc dev + localhost + `DEBUG_TOKEN`. `?debug=true` từ client chỉ còn là
một lời **xin**, server không nghe. Đã kiểm trên chính bản đang chạy ở `:9999`:

```
$ node tools/browsertest/dbggate.mjs 9999
KHÔNG đăng nhập    server cấp debug = false · cảnh giới 0 → 0  ✓ lệnh bị từ chối
```

Và từ `5321dd…`, mạng chơi nào đã dùng lệnh test thì **không được ghi vào bảng xếp hạng nữa**
(`room.record()` kiểm `c.DebugUsed`). Tức là danh sách này sẽ không dài thêm.

## 2. Cách chọn ra danh sách

Mốc so sánh **không** lấy từ cảm giác. Lấy từ `cmd/sim` chạy trên đúng luật đang chạy:

```
$ go run ./cmd/sim --bots 40 --minutes 20
  Đại Thừa       2 ·  429.1s
  Chân Tiên      2 ·  478.7s
```

Chơi thật cần ~429 giây để lên Đại Thừa, ~479 giây để vào Tiên giới. Lấy **một nửa** mốc đó
làm ngưỡng — rộng tay, để người chơi giỏi gặp may không bị gắn oan:

| Điều kiện | Ngưỡng |
|---|---|
| Vào Tiên giới (tầng ≥ 90) | sống < 240 s |
| Lên Đại Thừa (tầng ≥ 80) | sống < 215 s |

Kết quả trên 729 mạng chơi (22/10 11:49 → 31/10 17:03): **43 mạng đáng ngờ**.

Cột `guest_id` chia hai nhóm, và đây là cách chia **đáng tin**, không phải đoán theo tên:

- `179324…` — dạng *mốc-thời-gian + chuỗi ngẫu nhiên*. Dạng này chỉ
  sinh ra khi `crypto.randomUUID` **không dùng được**, mà nó chỉ không dùng được khi trang
  mở qua HTTP thường — tức là **vào bằng địa chỉ LAN**. Người chơi thật.
- `3210c7…` — dạng UUID, nghĩa là trang mở từ `localhost`. Đó là **máy chủ**:
  phần lớn là tôi chạy kiểm thử.


#### LAN — 6 mạng

| Thời điểm kết thúc | Tên | guest_id | Tầng | Uy Danh | Đòn | Hạ | Sống | Khu | Lý do nghi |
|---|---|---|---|---|---|---|---|---|---|
| 30/10 10:18:42 | Người chơi A | `179324…` | 100 | 2261 | 15 | 5 | 164 s | thien_dinh | Tiên giới (tầng 100) trong 164 s |
| 22/10 15:02:51 | Người chơi B | `179264…` | 89 | 737 | 0 | 0 | 17 s | nui_da | Đại Thừa (tầng 89) trong 17 s |
| 22/10 15:53:09 | Người chơi B (tên 2) | `179264…` | 89 | 739 | 0 | 0 | 20 s | lang_tre | Đại Thừa (tầng 89) trong 20 s |
| 22/10 14:27:44 | Người chơi B | `179264…` | 89 | 1377 | 21 | 6 | 69 s | nui_da | Đại Thừa (tầng 89) trong 69 s |
| 28/10 09:49:14 | Người chơi C (tên 2) | `179264…` | 89 | 720 | 0 | 0 | 149 s | lang_tre | Đại Thừa (tầng 89) trong 149 s |
| 22/10 13:02:36 | Người chơi C | `179264…` | 89 | 1520 | 10 | 3 | 190 s | cho_dem | Đại Thừa (tầng 89) trong 190 s |

#### LOCAL — 37 mạng

| Thời điểm kết thúc | Tên | guest_id | Tầng | Uy Danh | Đòn | Hạ | Sống | Khu | Lý do nghi |
|---|---|---|---|---|---|---|---|---|---|
| 29/10 11:59:49 | DaoTo | `26e673…` | 130 | 0 | 0 | 0 | 1 s | nui_da | Tiên giới (tầng 130) trong 1 s |
| 29/10 11:55:13 | DaoTo | `134d9e…` | 130 | 1 | 0 | 0 | 1 s | nui_da | Tiên giới (tầng 130) trong 1 s |
| 29/10 11:56:59 | DaoTo | `cfc9f8…` | 130 | 1 | 0 | 0 | 1 s | lang_tre | Tiên giới (tầng 130) trong 1 s |
| 29/10 11:58:37 | DaoTo | `f4f19f…` | 130 | 1 | 0 | 0 | 1 s | cho_dem | Tiên giới (tầng 130) trong 1 s |
| 29/10 11:59:41 | DaoTo | `352057…` | 130 | 1 | 0 | 0 | 1 s | nui_da | Tiên giới (tầng 130) trong 1 s |
| 29/10 12:02:29 | DaoTo | `4863cb…` | 130 | 1 | 0 | 0 | 1 s | nui_da | Tiên giới (tầng 130) trong 1 s |
| 29/10 11:57:15 | DaoTo | `37bb88…` | 130 | 577 | 1 | 1 | 2 s | lang_tre | Tiên giới (tầng 130) trong 2 s |
| 29/10 14:57:55 | HudTest | `eba65d…` | 130 | 0 | 0 | 0 | 3 s | truc_lam | Tiên giới (tầng 130) trong 3 s |
| 28/10 13:04:57 | DbgChk | `be5d4a…` | 130 | 78 | 1 | 1 | 4 s | nui_da | Tiên giới (tầng 130) trong 4 s |
| 29/10 12:00:18 | DaoTo | `ac0a71…` | 130 | 655 | 2 | 2 | 5 s | nui_da | Tiên giới (tầng 130) trong 5 s |
| 29/10 12:02:35 | DaoTo | `4863cb…` | 130 | 103 | 0 | 0 | 5 s | nui_da | Tiên giới (tầng 130) trong 5 s |
| 29/10 11:56:05 | DaoTo | `c7ab11…` | 130 | 260 | 4 | 3 | 11 s | nui_da | Tiên giới (tầng 130) trong 11 s |
| 28/10 13:33:18 | Test01 | `c73b64…` | 130 | 0 | 0 | 0 | 11 s | nui_da | Tiên giới (tầng 130) trong 11 s |
| 28/10 13:33:33 | Test01 | `b7ec86…` | 130 | 600 | 11 | 1 | 14 s | nui_da | Tiên giới (tầng 130) trong 14 s |
| 28/10 14:54:03 | Người chơi D | `565297…` | 130 | 1846 | 0 | 0 | 61 s | lang_tre | Tiên giới (tầng 130) trong 61 s |
| 28/10 12:52:54 | TestThThuTianThu | `565297…` | 130 | 2462 | 9 | 5 | 61 s | cho_dem | Tiên giới (tầng 130) trong 61 s |
| 28/10 10:42:55 | Người chơi E | `3210c7…` | 130 | 2333 | 7 | 6 | 63 s | nui_da | Tiên giới (tầng 130) trong 63 s |
| 28/10 13:55:33 | TestThThuTianThu | `565297…` | 130 | 320 | 5 | 3 | 70 s | cho_dem | Tiên giới (tầng 130) trong 70 s |
| 28/10 10:44:28 | Người chơi E | `3210c7…` | 130 | 1958 | 1 | 0 | 91 s | nui_da | Tiên giới (tầng 130) trong 91 s |
| 28/10 12:36:14 | TestThThuTianThu | `565297…` | 130 | 2347 | 12 | 8 | 102 s | nui_da | Tiên giới (tầng 130) trong 102 s |
| 31/10 17:03:54 | Người chơi G (tên 2) | `5a6442…` | 130 | 4162 | 42 | 19 | 117 s | thien_dinh | Tiên giới (tầng 130) trong 117 s |
| 28/10 14:28:44 | Người chơi D | `565297…` | 130 | 2248 | 9 | 3 | 119 s | nui_da | Tiên giới (tầng 130) trong 119 s |
| 28/10 10:46:41 | TestThThuTianThu | `565297…` | 130 | 2070 | 5 | 1 | 161 s | nui_da | Tiên giới (tầng 130) trong 161 s |
| 28/10 14:06:14 | TestThThuTianThu | `565297…` | 130 | 3395 | 32 | 6 | 167 s | cho_dem | Tiên giới (tầng 130) trong 167 s |
| 29/10 14:03:27 | Người chơi G | `5a6442…` | 130 | 2629 | 10 | 3 | 216 s | thien_dinh | Tiên giới (tầng 130) trong 216 s |
| 28/10 13:54:13 | Người chơi F | `22af96…` | 120 | 1684 | 2 | 1 | 28 s | cho_dem | Tiên giới (tầng 120) trong 28 s |
| 28/10 14:57:57 | Người chơi D | `565297…` | 120 | 1739 | 2 | 2 | 41 s | lang_tre | Tiên giới (tầng 120) trong 41 s |
| 31/10 11:05:38 | iPhoneTest | `91032a…` | 120 | 1597 | 0 | 0 | 45 s | pham_tran | Tiên giới (tầng 120) trong 45 s |
| 31/10 10:47:00 | iPhoneTest | `5b4b73…` | 120 | 4515 | 21 | 0 | 48 s | bang_nguyet_ho | Tiên giới (tầng 120) trong 48 s |
| 31/10 10:06:48 | iPhoneTest | `8e3696…` | 120 | 5362 | 66 | 7 | 54 s | bang_nguyet_ho | Tiên giới (tầng 120) trong 54 s |
| 28/10 10:43:56 | TestThanThu | `565297…` | 120 | 1587 | 0 | 0 | 59 s | nui_da | Tiên giới (tầng 120) trong 59 s |
| 28/10 13:43:32 | Người chơi F | `22af96…` | 110 | 1520 | 10 | 2 | 89 s | nui_da | Tiên giới (tầng 110) trong 89 s |
| 31/10 09:38:13 | V65Test | `ae4d18…` | 89 | 727 | 0 | 0 | 7 s | linh_son_coc | Đại Thừa (tầng 89) trong 7 s |
| 28/10 12:17:13 | TestThThuTianThu | `565297…` | 89 | 2070 | 8 | 0 | 40 s | nui_da | Đại Thừa (tầng 89) trong 40 s |
| 28/10 12:18:19 | TestThThuTianThu | `565297…` | 89 | 720 | 0 | 0 | 58 s | nui_da | Đại Thừa (tầng 89) trong 58 s |
| 28/10 11:31:24 | TestThThuTianThu | `565297…` | 80 | 1534 | 7 | 6 | 106 s | cho_dem | Đại Thừa (tầng 80) trong 106 s |
| 28/10 11:34:23 | TestThThuTianThu | `565297…` | 80 | 795 | 1 | 1 | 116 s | cho_dem | Đại Thừa (tầng 80) trong 116 s |

## 3. Đọc hai bảng trên

**Nhóm LAN (6 mạng, 3 guest_id).** Đây là nhóm duy nhất cần anh quyết.

- `179264…` (*Người chơi B* / *Người chơi B (tên 2)*) — ba mạng tầng 89. Hai
  trong đó **0 đòn, 0 hạ**: chắc chắn là lệnh test.
- `179264…` (*Người chơi C* / *Người chơi C (tên 2)*) — hai mạng tầng 89,
  một mạng **0 đòn, 0 hạ**.
- `179324…` (*Người chơi A*) — tầng 100 trong 164 s, có 15 đòn và
  5 lần hạ. Mạng này **ít rõ ràng nhất**: 164 s vẫn nhanh bất thường, nhưng người này có
  đánh thật. Tôi không chứng minh được nó là lệnh test.

**Nhóm LOCAL (37 mạng).** Gần như toàn bộ là rác kiểm thử của tôi: `DaoTo` (7 mạng tầng 130
trong 1–5 giây), `HudTest`, `DbgChk`, `Test01`, `V65Test`, `iPhoneTest`, `TestThThuTianThu`.
Những dòng này không có lý do gì tồn tại trong bảng xếp hạng.

Bốn guest_id trong nhóm LOCAL mang tên người thật (`Người chơi E`, `Người chơi F`, `Người chơi G`/`Người chơi G (tên 2)`) —
tức là **có người chơi ngay trên máy chủ**. Tôi không tách được đó là anh hay là tôi, nên
không đoán.

## 4. Đề nghị — nhưng anh quyết

Tôi **không tự ẩn** gì cả. Đề nghị của tôi:

| Nhóm | Đề nghị | Vì sao |
|---|---|---|
| 37 mạng LOCAL | Ẩn hết | Rác kiểm thử. Không ai mất gì. |
| 5 mạng tầng 89 (2 guest_id LAN) | Ẩn | `0 đòn / 0 hạ` không giải thích được bằng cách chơi. |
| *Người chơi A* tầng 100 | **Giữ lại** | Có đánh thật. Nhanh đáng ngờ, nhưng ngờ không phải bằng chứng — và cửa debug lúc đó server tự mở. |

**Không đề nghị cấm ai.** Cửa mở là lỗi của server; phạt người đi qua cửa đó thì không công
bằng. Nếu sau khi cửa đã đóng (từ `5321dd…`) mà vẫn có người đạt mốc bất thường thì lúc đó
mới là chuyện khác — và lúc đó bảng ở `/admin/#board` sẽ tự chỉ ra.

## 5. Dùng trang quản trị

`/admin/#board` → tích **"Chỉ mạng đáng ngờ"**. Mỗi dòng có nút *Ẩn khỏi bảng*; bấm thì
**bắt buộc** nhập lý do (gợi ý sẵn bằng chứng đã tính), lý do lưu vào `/admin/#audit`.
Dòng đã ẩn hiện mờ kèm nút *Bỏ ẩn*.

Ẩn chỉ ghi một dòng vào bảng `hidden_lives`; bảng `lives` không bị sửa. Không có đường nào
trong cả server xoá được một mạng chơi.

## 2. Lỗi Bí Kíp nhân đôi EXP (v6.21 / Q33) — rà ngày 1/11/2026

> **Chưa ẩn, chưa xoá gì.** Danh sách để anh xem rồi quyết, giống mục 1.

**Lỗi:** từ commit `0561a5…` (29/10 14:26) tới bản sửa (1/11), nhặt Bí Kíp khi đã đủ lượt thì
**nhân đôi** tổng EXP, thay vì +20 (docs/EXP_AUDIT.md). Người nhặt không làm gì sai. Họ chỉ
nhặt một vật phẩm, và server cho quá tay.

**Không có bằng chứng trực tiếp.** Bảng `lives` không ghi EXP theo nguồn. Từ bản sửa trở đi
`Game.ExpTrace` mới ghi được, nhưng chưa lưu vào DB. Nên chỉ lọc được theo **dấu hiệu**:

- Khi chưa có lỗi, farm thuần cần ≥ 25 phút để lên Đại Thừa (bảng §6.0: 21,9 phút tới hết
  Độ Kiếp; luật cũ còn chậm hơn).
- Lên Đại Thừa trong **dưới 10 phút** mà **dưới 30 đòn đánh** thì không còn nguồn EXP nào
  khác giải thích được. Đó là nhóm "Rất nghi".
- Đánh nhiều thì giao tranh (cướp 50 % EXP, chém bot cấp cao) giải thích được một phần. Đó là
  nhóm "Có thể".
- `Test01`–`Test05` ở tầng 130 không có mốc Đại Thừa: đó là `cmd/loadtest` đặt thẳng lên Đạo Tổ.

### Staging — 30 mạng tới Luyện Hư trở lên / 198 mạng kết thúc từ 29/10 14:26

**Rất nghi — lên Đại Thừa < 10 phút mà < 30 đòn đánh (farm thuần không thể: không lỗi thì ≥ 25 phút)** — 3 mạng

| Kết thúc | Tên | guest_id | Cảnh giới đỉnh (tầng) | Lên Đại Thừa sau | Đòn | Hạ | Sống | Uy Danh | Khu |
|---|---|---|---|---|---|---|---|---|---|
| 31/10 13:03 | Người chơi H | `a0eb00…` | Chân Tiên (90) | 156 s | 2 | 0 | 193 s | 1312 | thien_dinh |
| 30/10 21:46 | Người chơi H (tên 2) | `a0eb00…` | Đại La (120) | 169 s | 0 | 0 | 487 s | 1661 | bang_nguyet_ho |
| 30/10 20:55 | Người chơi H (tên 2) | `a0eb00…` | Thái Ất (110) | 341 s | 19 | 1 | 646 s | 4244 | linh_son_coc |

**Có thể — lên Đại Thừa < 10 phút nhưng đánh nhiều (giao tranh giải thích được một phần)** — 4 mạng

| Kết thúc | Tên | guest_id | Cảnh giới đỉnh (tầng) | Lên Đại Thừa sau | Đòn | Hạ | Sống | Uy Danh | Khu |
|---|---|---|---|---|---|---|---|---|---|
| 30/10 23:59 | Người chơi I | `1edb55…` | Kim Tiên (100) | 81 s | 83 | 4 | 822 s | 4810 | bang_nguyet_ho |
| 31/10 00:27 | Người chơi I | `1edb55…` | Đạo Tổ (130) | 104 s | 151 | 16 | 1455 s | 13904 | bang_nguyet_ho |
| 01/11 19:14 | Người chơi J | `ee92ec…` | Chân Tiên (90) | 123 s | 63 | 11 | 360 s | 4446 | truc_lam |
| 01/11 18:48 | Người chơi K | `89c2c4…` | Chân Tiên (90) | 175 s | 61 | 7 | 436 s | 5792 | truc_lam |

**Tới Luyện Hư+ nhưng chưa lên Đại Thừa hoặc lâu hơn 10 phút — không đủ dấu hiệu** — 3 mạng

| Kết thúc | Tên | guest_id | Cảnh giới đỉnh (tầng) | Lên Đại Thừa sau | Đòn | Hạ | Sống | Uy Danh | Khu |
|---|---|---|---|---|---|---|---|---|---|
| 30/10 23:45 | Người chơi I | `1edb55…` | Hợp Thể (67) | — | 0 | 0 | 131 s | 420 | linh_son_coc |
| 31/10 13:10 | Người chơi H | `a0eb00…` | Luyện Hư (56) | — | 2 | 0 | 271 s | 567 | linh_son_coc |
| 30/10 00:36 | Người chơi H (tên 3) | `a0eb00…` | Luyện Hư (52) | — | 0 | 0 | 105 s | 309 | bang_nguyet_ho |

**Công cụ đo (cmd/loadtest, lệnh `max` lên thẳng Đạo Tổ) — không phải lỗi Bí Kíp** — 20 mạng

| Kết thúc | Tên | guest_id | Cảnh giới đỉnh (tầng) | Lên Đại Thừa sau | Đòn | Hạ | Sống | Uy Danh | Khu |
|---|---|---|---|---|---|---|---|---|---|
| 30/10 00:31 | Test02 | `842008…` | Đạo Tổ (130) | — | 38 | 9 | 248 s | 2050 | truc_lam |
| 30/10 00:31 | Test05 | `4ab8eb…` | Đạo Tổ (130) | — | 11 | 2 | 260 s | 688 | pham_tran |
| 30/10 00:31 | Test03 | `dbbaac…` | Đạo Tổ (130) | — | 1 | 0 | 260 s | 400 | pham_tran |
| 30/10 00:31 | Test04 | `b3c53b…` | Đạo Tổ (130) | — | 20 | 6 | 208 s | 915 | linh_son_coc |
| 30/10 00:31 | Test01 | `eebbeb…` | Đạo Tổ (130) | — | 42 | 11 | 260 s | 3688 | thien_dinh |
| 30/10 01:08 | Test01 | `d5354f…` | Đạo Tổ (130) | — | 19 | 1 | 63 s | 1250 | pham_tran |
| 30/10 01:08 | Test02 | `5ae30d…` | Đạo Tổ (130) | — | 28 | 3 | 63 s | 1667 | pham_tran |
| 30/10 01:08 | Test03 | `566a83…` | Đạo Tổ (130) | — | 38 | 5 | 63 s | 1650 | thien_dinh |
| 30/10 01:08 | Test04 | `709d89…` | Đạo Tổ (130) | — | 14 | 2 | 63 s | 723 | truc_lam |
| 30/10 01:08 | Test05 | `3a83c3…` | Đạo Tổ (130) | — | 32 | 2 | 236 s | 1100 | linh_son_coc |
| 30/10 01:25 | Test03 | `233cfc…` | Đạo Tổ (130) | — | 28 | 4 | 64 s | 1015 | pham_tran |
| 30/10 01:25 | Test05 | `4a2baa…` | Đạo Tổ (130) | — | 11 | 3 | 64 s | 828 | bang_nguyet_ho |
| 30/10 01:25 | Test04 | `68b9cc…` | Đạo Tổ (130) | — | 0 | 0 | 64 s | 18 | bang_nguyet_ho |
| 30/10 01:25 | Test01 | `75ffda…` | Đạo Tổ (130) | — | 25 | 3 | 64 s | 810 | truc_lam |
| 30/10 01:25 | Test02 | `a911c7…` | Đạo Tổ (130) | — | 7 | 0 | 64 s | 450 | thien_dinh |
| 30/10 01:31 | Test03 | `419004…` | Đạo Tổ (130) | — | 0 | 0 | 64 s | 0 | linh_son_coc |
| 30/10 01:31 | Test05 | `aa6569…` | Đạo Tổ (130) | — | 1 | 1 | 64 s | 75 | pham_tran |
| 30/10 01:31 | Test01 | `a792ef…` | Đạo Tổ (130) | — | 28 | 2 | 64 s | 1400 | truc_lam |
| 30/10 01:31 | Test02 | `3c3b2b…` | Đạo Tổ (130) | — | 13 | 1 | 64 s | 1156 | thien_dinh |
| 30/10 01:31 | Test04 | `b6a95f…` | Đạo Tổ (130) | — | 21 | 2 | 64 s | 1457 | bang_nguyet_ho |

### :9999 (LAN) — 4 mạng tới Luyện Hư trở lên / 11 mạng kết thúc từ 29/10 14:26

**Rất nghi — lên Đại Thừa < 10 phút mà < 30 đòn đánh (farm thuần không thể: không lỗi thì ≥ 25 phút)** — 1 mạng

| Kết thúc | Tên | guest_id | Cảnh giới đỉnh (tầng) | Lên Đại Thừa sau | Đòn | Hạ | Sống | Uy Danh | Khu |
|---|---|---|---|---|---|---|---|---|---|
| 01/11 17:31 | Người chơi L | `a50a9a…` | Chân Tiên (90) | 164 s | 15 | 4 | 177 s | 1748 | truc_lam |

**Có thể — lên Đại Thừa < 10 phút nhưng đánh nhiều (giao tranh giải thích được một phần)** — 3 mạng

| Kết thúc | Tên | guest_id | Cảnh giới đỉnh (tầng) | Lên Đại Thừa sau | Đòn | Hạ | Sống | Uy Danh | Khu |
|---|---|---|---|---|---|---|---|---|---|
| 30/10 00:07 | Người chơi M | `39c6d7…` | Thái Ất (110) | 58 s | 91 | 16 | 686 s | 7699 | pham_tran |
| 01/11 15:59 | Người chơi L | `a50a9a…` | Kim Tiên (100) | 77 s | 131 | 17 | 1014 s | 11603 | bang_nguyet_ho |
| 01/11 19:09 | Người chơi L | `a50a9a…` | Đại La (120) | 185 s | 389 | 27 | 1745 s | 41320 | linh_son_coc |

