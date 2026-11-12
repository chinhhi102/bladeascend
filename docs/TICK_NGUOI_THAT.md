# Tick server lúc NGƯỜI THẬT đang chơi — 30/10 → 12/11/2026

Nguồn: `server/data/telemetry.db`, bảng `server_samples` (mỗi 10 giây, mỗi khu một dòng:
tick trung bình, tick p95, số người, số bot, KB/s, %CPU, RSS). Ghép với bảng `samples` để
biết mốc nào CÓ người thật đang nối — **loại hết phiên `automated = 1`** (Playwright, bài đo
nhiều tab), đúng như báo cáo `docs/perf/real/`.

Lệnh dựng lại số liệu:

```bash
cd server && go run ./cmd/tickquery -db data/telemetry.db
```

Ngân sách: 30 tick/giây → **33,3 ms** là trần tuyệt đối. Ngưỡng cảnh báo admin
(`config/game.json` → `admin.tick_warn_ms`) là **3,5 ms**.

---

## 1. Kết luận ngắn

1. **Tick không phải vấn đề.** Khu đang có người thật: p95 trung vị **2,57 ms** / 33,3 ms
   ngân sách — dùng 7,7%. p99 của p95 là 19,77 ms; chỉ **2,24%** số mốc vượt nửa ngân sách.
2. **Nhưng số liệu này chưa chứng minh được gì về lúc đông.** Đông nhất từng ghi được là
   **3 NGƯỜI trong một khu**, và 87% số mốc chỉ có **1 người**. Tải tick hiện nay gần như
   hoàn toàn là **bot** (20 mỗi khu, 30 ở Thiên Đình). Trần phòng là 50 người — chưa có một
   mẫu người thật nào ở gần mức đó.
3. **Ngưỡng cảnh báo 3,5 ms đã mất tác dụng:** 37,9% số mốc có người vượt ngưỡng. Một cảnh
   báo nổ hơn một phần ba thời gian là cảnh báo không ai đọc nữa.
4. **Thiên Đình đắt gấp 2,8 lần khu thường** (p95 trung vị 4,80 ms so với 1,74–2,97). Đúng
   như B21 dự đoán khi đặt sàn bot 30 cho khu này, nhưng con số thật (4,80) **cao hơn** số
   đo giả lập S13 hồi 02/11 (3,22–3,75).
5. **Client tụt FPS KHÔNG phải lỗi server.** Lúc người thật tụt dưới 45 FPS, tick server cùng
   lúc là 2,59 ms (trung vị) — còn *thấp hơn* mức chung 3,29 ms. Nghẽn nằm ở máy người chơi
   (đúng hướng B15/B23), không ở server.
6. **Cảnh báo về chính phép đo:** server thật chạy trên máy dev dùng chung. Sáng nay, cùng
   build, cùng người chơi, tick p95 trung vị **2,13 ms lúc máy rảnh → 6,03 ms lúc máy bận**.
   Mọi so sánh tick giữa các build trong sổ này đều phải đọc kèm §5.

---

## 2. Số tổng quan

25 920 dòng mẫu server, 30/10 13:19 → 12/11 09:38. Phân loại theo ai đang nối:

| | Số dòng | tick TB | p95 TB | p95 trung vị | p95 max | vượt 3,5 ms | vượt 33,3 ms |
|---|---|---|---|---|---|---|---|
| **Chỉ người thật** | 15 230 | 1,51 | 3,47 | **2,29** | 97,72 | 30,4% | 27 |
| Người + máy đo | 390 | 1,79 | 5,02 | 3,71 | 23,56 | 51,3% | 0 |
| Chỉ máy đo | 1 470 | 1,17 | 3,31 | 2,13 | 43,76 | 30,3% | 1 |
| Không ai nối | 8 830 | 1,41 | 3,49 | 2,32 | 38,46 | 31,6% | 3 |

Hàng "không ai nối" gần bằng hàng "người thật" — vì tải là **bot**, không phải người.

Chỉ tính **khu đang có người** (humans > 0) lúc có người thật online — đây là con số đáng
dùng nhất:

```
n = 4 963 dòng · tick TB 1,93 ms · p95: TB 4,24 · trung vị 2,57 · p99 19,77 · max 97,72
vượt 3,5 ms: 37,9%  ·  vượt nửa ngân sách (16,65 ms): 2,24%  ·  vượt 33,3 ms: 7 dòng
CPU tiến trình: trung vị 18,5% · p95 44,6% · max 78,5%    RSS: trung vị 44 MB · max 56 MB
```

## 3. Theo khu và theo mật độ

| Khu | Dòng | p95 TB | p95 trung vị | p95 max | vượt 3,5 ms |
|---|---|---|---|---|---|
| 0 Thiên Đình (sàn bot 30) | 1 935 | 6,58 | **4,80** | 97,72 | 61,3% |
| 1 Linh Sơn Cốc | 673 | 3,88 | 2,97 | 19,02 | 43,4% |
| 2 Trúc Lâm Mê Cung | 820 | 2,84 | 2,11 | 24,75 | 26,0% |
| 3 Phàm Trấn | 773 | 2,28 | 1,85 | 12,75 | 13,6% |
| 4 Băng Nguyệt Hồ | 762 | 2,11 | 1,74 | 10,53 | 11,3% |

Theo số nhân vật trong khu (người + bot) — tick tăng **nhanh hơn tuyến tính**:

| Nhân vật | Dòng | p95 trung vị | vượt 3,5 ms |
|---|---|---|---|
| 1–5 | 322 | 0,45 | 0,9% |
| 6–10 | 733 | 1,21 | 5,2% |
| 11–20 | 12 454 | 2,24 | 27,5% |
| 21–30 | 1 976 | **5,27** | 66,6% |
| 31+ | 70 | 4,54 | 64,3% |

Từ 20 lên 30 nhân vật (+50%) thì tick p95 **tăng 2,4 lần**. Nếu quan hệ này giữ nguyên, một
khu 50 nhân vật sẽ ở khoảng 12–20 ms — vẫn dưới 33,3 ms nhưng hết sạch biên an toàn. **Đây
là phép ngoại suy, không phải số đo**; muốn biết thật thì phải chạy `cmd/loadtest` (S12).

Số NGƯỜI thật trong một khu, toàn bộ 13 ngày:

| 1 người | 2 người | 3 người |
|---|---|---|
| 4 334 mốc | 597 mốc | 32 mốc |

## 4. Người chơi tụt FPS — có phải tại server không?

Lấy 5 191 mẫu client của người thật, ghép với tick server cùng khu ±10 giây:

| | Tick p95 server cùng lúc (trung vị) |
|---|---|
| Lúc client tụt dưới 45 FPS (1 329 mẫu, 25,6%) | **2,59 ms** |
| Mẫu client bất kỳ | 3,29 ms |

Tick lúc client giật còn *thấp hơn* lúc bình thường → **không có liên hệ**. Nguyên nhân giật
nằm ở máy người chơi: rò rỉ texture (B23, đã sửa) và chi phí cây cảnh PixiJS (B15, còn mở).

## 5. Vì sao KHÔNG được so tick giữa các build bằng số này

Cùng một ngày sáng nay, cùng build `20261111_2011_52adf2e`, cùng 1–2 người chơi, chỉ khác
mức bận của máy:

| Khoảng | Dòng | tick TB | p95 trung vị | p95 max | vượt 3,5 ms |
|---|---|---|---|---|---|
| 09:00 → 09:31 (máy rảnh) | 26 | 1,07 | **2,13** | 4,98 | 15% |
| 09:31 → 09:50 (máy đang biên dịch Go) | 160 | 2,88 | **6,03** | 20,26 | 68% |

Ba bằng chứng nữa cho thấy đó là CÁI MÁY chứ không phải luật game:

- **Log server**: 160 dòng `tick chậm` (> 25 ms) kể từ lần khởi động 11/11 20:13 — **tất cả**
  nằm trong 09:31–09:44 sáng nay, **0 dòng** trong 13 giờ trước đó. Giai đoạn 09:31–09:44 là
  đúng lúc phiên làm việc này chạy `go run` liên tục trên cùng máy.
- **Phần tốn thời gian nhảy lung tung**: cùng một phút có dòng "bot 37,4 ms (20 bot)", dòng
  "game 38,3 ms", dòng "gửi 28,6 ms". Một đoạn mã nặng thì luôn nặng ở cùng một chỗ; nhảy
  ngẫu nhiên giữa các giai đoạn là dấu hiệu **cả tiến trình bị dừng**.
- **44 dòng có tick trung bình LỚN HƠN tick p95** — chỉ xảy ra khi vài tick dài hàng trăm ms
  kéo trung bình lên. Và 15 dòng lặp y hệt dòng trước cùng khu: vòng game đứng im giữa hai
  lần lấy mẫu.

Lúc viết báo cáo này máy đang: load 9,6–13 trên 12 nhân · VS Code 92% CPU · 3 tiến trình
`claude` · **2 GB swap dùng hết**. Server game là một tiến trình bình thường trên cái máy đó.

**Và đây là chỗ quy tắc 16 chưa đủ:** load lúc đó là ~13 trên 12 nhân = **1,08/nhân**, tức
là **ĐẠT** ngưỡng "máy rảnh < 1,5" mà `tools/cleanup.sh` in ra — nhưng tick vẫn xấu đi gần
3 lần. Ngưỡng 1,5 hợp cho bài đo THÔNG LƯỢNG (FPS trung bình, băng thông); tick là số đo
**ĐỘ TRỄ**, chỉ cần goroutine phòng bị trượt lượt một nhịp là p95 đã nhảy. Với tick nên đòi
chặt hơn: load/nhân **< 0,5**, không có tiến trình nào chiếm trên ~50% một nhân, và không
dịch Go / không chạy bài đo trình duyệt trong suốt khoảng đo.

**Khu đối chứng** (0 người, 19–21 bot — khối lượng việc gần như không đổi giữa các build)
cho thấy nhiễu này dịch chuyển theo ngày, không theo code:

| Ngày | p95 trung vị khu đối chứng |
|---|---|
| 02/11 | 2,23 |
| 05/11 | 2,32 |
| 06/11 | 2,60 |
| 07/11 | 2,51 |
| 10/11 | 2,12 |
| **11/11** | **3,14** |
| 12/11 | 2,53 |

Chia khu đông cho khu đối chứng CÙNG BUILD thì khử bớt được nhiễu máy. Phần còn lại mới là
"khu đông đắt hơn mấy lần":

| Build | p95 khu 0 (21–30 n.vật) | p95 khu đối chứng | Tỉ lệ |
|---|---|---|---|
| `20261105_1544_3647348` | 2,95 | 2,23 | 1,32× |
| `20261106_1624_d7da481` | 6,92 | 2,76 | 2,51× |
| `20261106_1650_a959ba8` | 6,89 | 2,61 | 2,64× |
| `20261107_1114_b0fe1db` | 4,14 | 2,17 | 1,91× |
| `20261110_1257_d2f1331` | 3,10 | 1,50 | 2,07× |
| `20261111_1655_ec38aee` | 12,13 | 4,13 | 2,93× |
| `20261111_2011_52adf2e` (đang chạy) | 6,01 | 2,53 | 2,37× |

Tỉ lệ đi từ ~1,3–1,9× (trước 06/11) lên ~2,4–2,9× (từ 06/11). 06/11 là ngày lên bộ kỹ năng
v6.48 §6.3. **Đây là dấu hiệu, chưa phải kết luận** — mẫu nhỏ, máy nhiễu, và thành phần
trong khu (cảnh giới của bot, số kiếm) cũng đổi theo thời gian. Muốn kết luận thì phải chạy
S12/S13 trên máy rảnh ở hai commit.

## 6. Việc nên làm

1. **Đo lại tick bằng `cmd/loadtest` trên máy rảnh** (`tools/cleanup.sh` phải in
   load/nhân < 1,5), ở commit `d2f1331` và commit hiện tại, cùng seed — để trả lời dứt điểm
   câu "từ 06/11 khu đông có đắt lên thật không".
2. **Xem lại `admin.tick_warn_ms = 3,5`.** Nổ 37,9% thời gian với tải hiện tại. Hoặc nâng
   ngưỡng, hoặc tính theo khu (Thiên Đình có sàn bot 30 nên vốn dĩ đắt hơn).
3. **Đừng kết luận gì về 50 người từ dữ liệu người thật** — mẫu thật chưa bao giờ quá 3
   người/khu. Giới hạn 50 vẫn chỉ được chứng minh bằng bot (lượt #24).
4. **Ghi "máy rảnh / máy bận" vào mọi dòng tick** như §6 B13c đã đòi cho S9. Số tick của
   server chạy trên máy dev mà không có cột này thì không so được với nhau.
