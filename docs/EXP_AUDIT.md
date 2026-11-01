# EXP_AUDIT — lạm phát Linh khí (CHANGES v6.21, bước 1)

Ngày 1/11/2026. Báo cáo này gửi chủ dự án **trước khi sửa**, đúng như CHANGES_v6_21 yêu cầu.
Chưa đổi luật nào. Thay đổi duy nhất trong code là phần đo: mỗi lượng EXP giờ mang theo nguồn
của nó (`server/internal/game/expsource.go`, `Game.ExpTrace`). Có thêm lệnh đo `cmd/sim --audit`.

Đo lại: `cd server && go run ./cmd/sim --audit --seeds 3 --seed 1`

## Kết luận ngắn

1. **Nguyên nhân chính là một lỗi code, không phải hệ số cân bằng.** Nhặt **Bí Kíp khi đã đủ
   2 lượt** thì spec cho **+20 Linh khí** (§6.3). Code lại gọi
   `GainExp(p, p.TotalExp + 20)` (`items.go:220`). `GainExp` vốn đã CỘNG vào tổng, nên mỗi lần
   nhặt **tổng EXP nhân đôi**. Mà EXP cần tăng ×2 mỗi cảnh giới, nên mỗi cuốn Bí Kíp ≈ lên
   **một cảnh giới**. Bí Kíp có trọng số 10, thuộc loại hay gặp. Nam Châm hút rộng ×3 nên vơ
   được nhiều Bí Kíp hơn, vì vậy lỗi lộ rõ nhất khi có Nam Châm.
   - Tái hiện: bắt đầu Luyện Hư 1, chỉ farm, có Nam Châm → **Đại Thừa 10 sau 69–74 giây**
     (2 trên 3 seed). Bí Kíp chiếm **80–97 %** tổng EXP.
2. **Nghi phạm trong spec §6.0 không có trong code.** Viên EXP môi trường và viên rơi từ quái
   **không** nhân theo cảnh giới. Chúng cố định 5 / 20 / 60 và EXP quái 10 / 30 / 90 / 400.
   Bán kính hút đúng là nhân thẳng theo scale, và Nam Châm đúng là ×3.
3. **Bỏ lỗi Bí Kíp đi thì farm thuần hiện nay lại CHẬM hơn bảng mục tiêu §6.0**, không nhanh
   hơn. Áp nguyên văn luật §6.0 (viên × 1,5^r, hút × √scale, Nam Châm × 1,5) thì cảnh giới
   thấp và giữa trúng bảng. **Đại Thừa nhanh gấp đôi bảng**, và **Nam Châm nhanh hơn mức trần
   1,5 lần**. Xem phần "Câu hỏi cần anh chốt" ở cuối.

## Mọi nguồn EXP trong code

`r` = cảnh giới (0 = Luyện Khí … 8 = Đại Thừa). `scale` = 1 + 4·L/89 (L = cấp 0…89): Luyện Khí 1
= 1,0 · Luyện Hư 1 = 3,25 · Đại Thừa 1 = 4,6. EXP cần cho cả một cảnh giới = 1 360 × 2^r.

| Nguồn | Công thức trong code | Nhân theo cảnh giới? | Tệp |
|---|---|---|---|
| Viên EXP môi trường | 5 / 20 / 60 (tỉ lệ 80/18/2), mọc lại sau 5 s trong ô 1 000² | **Không** | `orbs.go` `addAmbientOrb` |
| Bán kính hút đồ | `pickup_r(90) × scale` × Nam Châm 3 → 90 / 292 / 414 (Nam Châm: 270 / 877 / 1 242) | Gián tiếp, qua scale (×1 → ×4,6) | `items.go:82` |
| Quái | Rơi viên tổng 10 / 30 / 90 / 400 (nhỏ/vừa/lớn/trùm), viên thuộc người chém 1,5 s | **Không** | `combat.go` `killMonster` |
| Chém trúng người (không chết) | Nạn nhân rớt 1 cảnh giới về sàn. Người chém nhận **50 %** phần mất, 30 % văng thành viên | Gián tiếp: phần mất ≈ cả 1 cảnh giới của nạn nhân | `combat.go:101` |
| Chém **bot** | Giống hệt người thật (50 %) | Như trên | `combat.go:101` |
| Hạ người/bot ở Luyện Khí | 50 % EXP nạn nhân + 30 | Không | `combat.go:128` |
| Nhặt kiếm khi đã đủ kiếm | +20 (`items.overcap_exp`) | Không | `exp.go:81` |
| **Nhặt Bí Kíp khi đủ lượt** | **+ (TotalExp + 20)** ← lỗi, spec ghi +20 | **Có: ×2 tổng EXP** | `items.go:220` |
| Song Bội Đan | × 2 mọi nguồn "boosted" (viên, quái, chém, kiếm dư), 20 s | Không | `exp.go` |
| Đài / Tiên khí | Chỉ cho **Tiên khí**, không cho Linh khí | — | `altar.go` |
| Phím debug `exp` / `realm` / `max` | Chỉ chạy khi server bật `-debug`. Staging và :9999 **không** bật | — | `room/debug.go` |

Không có hệ số nào khác nhân EXP theo cảnh giới.

Riêng **chém bot**: bot sinh ra tới Hợp Thể (`bots.max_spawn_realm = 6`). Chém trúng một bot
Hợp Thể thì nó mất trọn Luyện Hư = 43 520 EXP, người chém nhận **21 760**. Con số đó bằng nửa
cảnh giới Luyện Hư mỗi nhát. Đây không phải lỗi bị báo, nhưng là nguồn thứ hai đáng kể. v6.21
(25 % khi nạn nhân là bot) và v6.22 (bot sinh ở Luyện Khí 1) đều nhắm vào nó.

## Đo 1 — EXP/s theo nguồn, GIỮ NGUYÊN cảnh giới (code hiện tại)

Farm thuần: không ai chém được ai, 19 bot Luyện Khí farm cùng khu. Người được đo là một bot
farm (cùng não bot), mang đủ kiếm tối đa của cảnh giới đó. Mỗi ô là 60 giây × 3 seed. Lên cấp
thì bị đặt lại về đúng cấp đó, để con số là của CẢNH GIỚI đó.

| Cảnh giới | Nam Châm | **Tổng EXP/s** | viên môi trường | viên quái | kiếm dư | **Bí Kíp đủ lượt** | lần nhặt Bí Kíp đủ lượt | EXP cần / cảnh giới | **giây / cảnh giới** | giây / cảnh giới nếu bỏ Bí Kíp |
|---|---|---|---|---|---|---|---|---|---|---|
| Luyện Khí | không | 21,8 | 20,8 | 0,2 | 0,9 | 0 | 0 | 1 360 | 62 | 62 |
| Luyện Khí | có | 52,7 | 49,4 | 2,5 | 0,8 | 0 | 0 | 1 360 | 26 | 26 |
| Luyện Hư | không | 351,5 | 105,2 | 5,2 | 2,1 | **239,0** | 1 | 43 520 | 124 | 387 |
| Luyện Hư | có | 1 275,9 | 289,7 | 8,4 | 4,6 | **973,3** | 4 | 43 520 | **34** | 144 |
| Đại Thừa | không | 13 730,5 | 361,1 | 11,8 | 6,3 | **13 351,3** | 7 | 348 160 | **25** | 918 |
| Đại Thừa | có | 4 274,2 | 483,8 | 12,6 | 4,0 | **3 773,8** | 2 | 348 160 | 81 | 696 |

Ở Luyện Khí, Bí Kíp không gây hại vì tổng EXP còn nhỏ. Từ Luyện Hư trở lên, nó chiếm phần lớn
EXP. Chém người/bot không xuất hiện ở đây vì đây là farm thuần.

## Đo 2 — tái hiện lỗi 1/11 (code hiện tại)

Bắt đầu Luyện Hư 1, chỉ farm, không đặt lại cấp, tối đa 30 phút.

| Nam Châm | Seed | Hợp Thể | Độ Kiếp | Đại Thừa | Đại Thừa 10 đầy | Chân Tiên | Phần EXP từ Bí Kíp |
|---|---|---|---|---|---|---|---|
| có | 0 | 7 s | 52 s | 58 s | **74 s** | 121 s | 94 % (4 lần) |
| có | 1 | 27 s | 61 s | 67 s | **69 s** | 240 s | 97 % (4 lần) |
| có | 2 | 8 s | 146 s | 417 s | 502 s | 502 s | 80 % (4 lần) |
| không | 0 | 19 s | — | — | — | — | 72 % (1 lần) |
| không | 1 | (nhảy cóc) | 505 s | 505 s | — | — | 91 % (3 lần) |
| không | 2 | 235 s | 397 s | 401 s | 515 s | 515 s | 89 % (4 lần) |

Đối chiếu: §6.0 muốn Luyện Hư → hết Đại Thừa mất khoảng 19 phút (≈ 1 150 s) khi farm thuần.
Hiện tại có Nam Châm chỉ mất **70 s**. Bảng còn có những bước nhảy hai cảnh giới chỉ trong một
tick, ví dụ Độ Kiếp 505 s = Đại Thừa 505 s. Một cú nhân đôi như thế không thể đến từ việc nhặt
viên. Bước từ Đại Thừa 10 lên Chân Tiên cũng đến từ Bí Kíp, vì đủ lượt ở đó cho +2 Tiên khí,
đủ để vào Tiên giới.

Báo cáo nói "Luyện Hư → Chân Tiên dưới 20 giây". Phép đo trên ra 69–74 s tới Đại Thừa 10.
Người chơi thật lên nhanh hơn nữa vì họ chủ động vơ Bí Kíp, đánh thêm bot, và có Song Bội Đan
(×2 không áp vào Bí Kíp, nhưng áp vào mọi viên).

## Thử nghiệm: áp nguyên văn luật §6.0 (bản vá tạm, ĐÃ GỠ, không commit)

Bản vá thử gồm: Bí Kíp đủ lượt = +20 · viên môi trường và viên quái × 1,5^r (r của người nhặt)
· hút = 90 × √scale · Nam Châm × 1,5. Cách đo giống Đo 1.

| Cảnh giới | Nam Châm | EXP/s | giây / cảnh giới | Bảng §6.0 | Lệch |
|---|---|---|---|---|---|
| Luyện Khí | không | 22,2 | 61 | ~78 | −22 % ✅ (trong ±25 %) |
| Luyện Khí | có | 29,9 | 45 | ≥ 52 (78 / 1,5) | ✅ |
| Luyện Hư | không | 183,5 | 237 | ~198 | +20 % ✅ |
| Luyện Hư | có | 424,9 | **102** | ≥ 132 | ❌ Nam Châm nhanh ×2,3 |
| Đại Thừa | không | 1 637,2 | **213** | ~396 | ❌ −46 % (nhanh quá) |
| Đại Thừa | có | 2 285,8 | **152** | ≥ 264 | ❌ |

Thử thêm hệ số 1,3 (cùng bản vá): Luyện Khí 58 s · Luyện Hư 244 s · Đại Thừa 849 s (Nam Châm:
57 / 159 / 500). Đại Thừa thành chậm gấp đôi bảng. Kết quả giữa các seed còn dao động mạnh vì
mỗi ô chỉ đo 60 s. Muốn chỉnh cho trúng thì phải dùng phép farm trọn đường (`--farm`, bước 3
của v6.21), không dùng phép đo ngắn này.

## Câu hỏi cần anh chốt (đã ghi vào docs/QUESTIONS.md, Q33)

1. **Sửa lỗi Bí Kíp (+20 đúng như spec) — tôi làm ngay khi anh gật.** Đây là lỗi code, không
   phải đổi luật.
2. **Luật §6.0 viết theo nghi phạm sai.** Áp nguyên văn thì Đại Thừa nhanh gấp đôi bảng, và
   Nam Châm vượt trần 1,5 lần. Bảng nhịp và bài kiểm CI (±25 %, Luyện Hư → Chân Tiên ≥ 15
   phút) không thể cùng đạt với `orb_value_mult = 1,5` và Nam Châm × 1,5. Anh chọn:
   - **A. Giữ BẢNG NHỊP làm chuẩn** (khuyên dùng). Tôi chỉnh `exp_economy` (hệ số viên, Nam
     Châm, có thể cả hút √scale) bằng `cmd/sim --farm` cho tới khi trúng bảng ±25 %. Anh chỉ
     cần sửa chữ trong §6.0 cho khớp con số chốt.
   - B. Giữ đúng các hệ số §6.0 (1,5 / √scale / × 1,5), nới bảng cho khớp kết quả đo.
   - C. Chỉ sửa lỗi Bí Kíp + 25 % khi chém bot, giữ nguyên phần còn lại như hiện tại. Farm thuần
     khi đó chậm hơn bảng (Luyện Hư 387 s, Đại Thừa 918 s).
3. Danh sách các mạng đã lên cảnh giới cao nhờ lỗi này (bước 4): sẽ làm sau khi sửa, giống
   LEADERBOARD_AUDIT, và không ẩn mạng nào khi chưa có anh duyệt. Bảng `lives` không lưu EXP
   theo nguồn, nên chỉ lọc được theo dấu hiệu, ví dụ thời gian từ Luyện Hư lên Đại Thừa dưới
   X phút.
