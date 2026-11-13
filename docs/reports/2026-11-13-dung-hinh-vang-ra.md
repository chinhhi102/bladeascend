# Đứng hình + văng ra trên `:9999` — điều tra 13/11/2026

Nguồn: `server/data/telemetry.db` (kèm `-wal`), 24 giờ tới 18:05 ngày 13/11. Không ghi, không
xoá gì trên `:9999`. Tên người chơi và id đầy đủ **không** chép vào đây (quy tắc 15): phiên
gọi bằng 6 ký tự đầu của mã phiên.

## Kết luận ngắn

**Đứng hình và văng ra là MỘT nguyên nhân**, và nó là lỗi do chính tôi đưa vào lúc 14:45 cùng
ngày: shader trộn nền dựng mesh với geometry thiếu bốn thuộc tính riêng. Đã tái hiện, đã sửa,
đã có bài kiểm chặn tái phát.

**Còn một hiện tượng thứ hai chưa giải thích được**: 3 mẫu đứng hình thật (khung 8–29 giây)
trong một phiên, không đi kèm lỗi nào, không đổi khu. Đã thêm breadcrumb 30 việc gần nhất để
lần sau có manh mối; chưa sửa gì vì chưa biết sửa gì.

## 1. Lỗi JS theo chữ ký, 24 giờ

| chữ ký | số lần | phiên | build | lần đầu → cuối |
|---|---:|---:|---|---|
| `Error: shader and geometry incompatible, geometry missing the "aLocal" attribute` | **95** (22 của người thật) | 3 người thật + 27 máy | mọi build từ `f24ada9` (13/11 14:45) trở đi | 13/11 14:45 → 17:46 |
| `Error: lỗi giả trong khâu vẽ (kx.faultDraw)` | 57 | – | nhiều | – |

Dòng thứ hai là **lỗi giả do `crashtest.mjs` tự bắn ra** để nghiệm thu §13.8 — không phải lỗi.

Không có chữ ký nào khác trong 24 giờ.

### Thiết bị của người dính

| phiên | vào lúc | build | trình duyệt | GPU | nhân | webgl | max_tex |
|---|---|---|---|---|---:|---:|---:|
| `1056ed` | 17:17 | `b6c1c72` | Chrome 155 / Linux | AMD Radeon RX 570 (radeonsi) | 12 | 2 | 16384 |
| `dcefa5` | 17:21 | `b6c1c72` | Chrome 156 / Linux | NVIDIA GTX 1650 | 20 | 2 | 32768 |
| `577598` | 17:22 | `b6c1c72` | Chrome 149 / Linux | NVIDIA GTX 1650 | 20 | 2 | 32768 |

Cả ba đều là máy bàn, WebGL2, hỗ trợ đầy đủ texture nén. Giai đoạn: **100% trong trận**
(`phase: "match"`), không có lỗi nào ở trang chủ hay lúc tải.

## 2. Chuỗi sự kiện — phiên `dcefa5`, rõ nhất

```
17:21:00  draw_error        aLocal        ← vừa vào trận
17:25:50  zone_switch 2
17:25:50  draw_error ×2     aLocal
17:34:20  zone_switch 4
17:34:20  draw_error ×2     aLocal
17:35:30  zone_switch 0
17:35:30  draw_error ×3     aLocal
17:35:30  webgl_context_lost
17:35:30  renderer_rebuilt
17:51:05  webgl_context_lost (canRecover=true, texMB 44, 6 nhân vật)
```

**Mỗi lần đổi khu là 2–3 lỗi vẽ, 1:1, không trượt lần nào.** Lần thứ ba trong 10 giây thì lớp
chống sập §13.8 dựng lại renderer; đường lùi cuối của nó là `location.reload()` — nhìn từ ghế
người chơi đúng là *đứng hình mấy giây rồi văng ra khỏi trận*.

## 3. Nguyên nhân gốc

`client/src/render/groundshader.ts`: `GroundField` dựng mesh bằng một `MeshGeometry` trống,
chỉ có `aPosition`/`aUV`. Bốn thuộc tính của shader trộn — `aLocal`, `aOffset`, `aMaskUV`,
`aMaskSel` — chỉ được thêm vào trong `setChunks()`. Giữa hai thời điểm đó mà có một khung
được vẽ là Pixi ném lỗi.

Và lúc đổi khu thì **chắc chắn** có: `MapView` mới dựng xong, art ô nền của khu mới còn đang
tải, nên `updateGroundChunks` thoát sớm ở `if (!this.baseName)` và `setChunks` không bao giờ
được gọi. Đường thứ hai: `setChunks([])` cho khoá `''`, trùng giá trị khởi tạo của `gridKey`,
nên bị coi là "không có gì đổi" và thoát sớm.

Sửa (ba lớp, mỗi lớp tự nó đủ):

1. Geometry khởi tạo có **đủ bốn thuộc tính**, buffer rỗng. Pixi so chữ ký theo TÊN, không
   theo số đỉnh.
2. `gridKey` khởi tạo `null` thay vì `''`.
3. Mesh `visible = false` khi chưa có ô nào — một mesh rỗng không vẽ ra gì nhưng vẫn đi qua
   khâu kiểm shader/geometry.

## 4. Vì sao mọi cổng kiểm đều xanh

Không bài đo nào **đổi khu**. `scenediff`, `drawbreak`, `scenebench`, `vramgate` đều đứng yên
một chỗ trong một khu. `crashtest.mjs` thì chỉ bắn **lỗi giả** để thử lớp bắt lỗi — nó chưa
bao giờ hỏi "lúc chơi bình thường có lỗi THẬT nào không".

Lớp chống sập §13.8 **có chạy** trên build này và làm đúng việc của nó (bắt lỗi, tắt bộ lọc,
dựng lại renderer). Vấn đề không nằm ở nó mà ở chỗ nó phải ra tay liên tục.

Ba cổng mới:

| bài | canh gì |
|---|---|
| `tools/browsertest/drawguard.mjs` | đổi đủ 5 khu, **0 lỗi vẽ**. Tái hiện được lỗi cũ: 9 `draw_error` + 3 `renderer_rebuilt` |
| `tools/browsertest/nocompress.mjs` | máy KHÔNG hỗ trợ nén khối: vào trận được, 0 placeholder, 0 lỗi vẽ, không tải tệp `.ktx2` nào |
| (sửa) `tools/deploy.sh` | log server **nối thêm** thay vì ghi đè |

## 5. Các nghi phạm khác, và vì sao loại

| nghi phạm | kết luận | bằng chứng |
|---|---|---|
| KTX2 trên máy không hỗ trợ nén | **loại** | cả ba máy đều WebGL2 + max_tex 16384/32768, AMD/NVIDIA. Thêm nữa: `nocompress.mjs` chạy đường lùi WebP — đạt cả bốn điều |
| Shader nền trên GPU khác | **đúng một nửa** — là shader nền, nhưng không phụ thuộc GPU: tái hiện được trên máy dev |
| Premultiply alpha | **loại** | lên `:9999` lúc 17:49 (`2b1c592`), sau toàn bộ sự cố 17:18–17:46 |
| Hết bộ nhớ GPU | **loại** | `texMB` lúc lỗi: 21–59 MB, xa mọi ngưỡng |
| Server panic / rate limit / §7.1 hết 20 s | **loại** | 18 `server_panic` trong 24 h đều do `crashtest` bắn lỗi giả; cả ba phiên **kết thúc bình thường** (476 s, 1822 s, 2474 s), không phiên nào bị server đóng |

## 6. Hiện tượng thứ hai — CHƯA giải thích được

Lọc chặt (`frame_p95 ≥ 200 ms`, bỏ 30 giây đầu phiên, bỏ mẫu cách `tab_hidden`/`tab_visible`
dưới 25 giây) thì 24 giờ chỉ còn **3 mẫu**, tất cả trong phiên `dcefa5`:

| giờ | fps | khung p95 | khung max | texMB | khu | cảnh giới | kiếm | có lỗi vẽ gần đó |
|---|---:|---:|---:|---:|---:|---:|---:|---|
| 17:22:35 | 0,1 | 29 139 ms | 29 139 ms | 38 | 0 | 0 | 1 | không |
| 17:42:45 | 0,3 | 7 942 ms | 7 942 ms | 45 | 0 | 101 | 656 | không |
| 17:43:05 | 3,6 | 1 618 ms | 7 927 ms | 51 | 0 | 101 | 487 | không |

Không lỗi, không đổi khu, không đổi cỡ cửa sổ, VRAM bình thường. Chưa đủ dữ kiện.

**Cảnh báo về phép đo:** `frame_max` một mình thì **vô dụng** — nó đếm cả quãng tab bị ẩn.
Có một mẫu `frame_max = 56 844 164 ms` (15,8 giờ): một tab để qua đêm. Mọi kết luận về "đứng
hình" phải dùng `frame_p95` và phải loại mẫu quanh `tab_hidden`/`tab_visible`.

Đã thêm **breadcrumb 30 việc gần nhất** gửi kèm mọi lỗi và mọi sự kiện WebGL (trường `bc`
trong `errorContext`): đổi khu, dùng kỹ năng, đổi cỡ cửa sổ, nạp atlas/texture hỏng, mất ngữ
cảnh WebGL, lỗi vẽ. Chỉ ghi VIỆC, không ghi gì người chơi gõ.

## 7. Hai chỗ chính tôi làm hỏng việc điều tra

1. **Log server bị xoá sạch.** `tools/deploy.sh` mở log bằng `>`, nên ba lượt deploy trong
   ngày đã xoá đúng quãng 17:18–17:46 cần soi. Cả cuộc điều tra phải dựa hết vào
   `telemetry.db`. Đã đổi sang `>>` kèm xoay vòng ở mốc 20 MB, và sửa chỗ kiểm "server đã
   lên" cho nó chỉ soi 40 dòng cuối (soi cả tệp thì trúng dòng của lượt trước, luôn báo xanh).
2. **Bản sao telemetry đầu tiên thiếu tệp `-wal`.** SQLite ở chế độ WAL giữ phần ghi mới nhất
   ở tệp bên cạnh, nên bản sao thiếu nó mất khoảng 15 phút cuối. Tôi suýt kết luận sai: hai
   phiên hiện ra "dur = 0, không kết thúc" (tức *văng ra*), trong khi dữ liệu đủ cho thấy cả
   hai **kết thúc bình thường** sau 30 và 41 phút. Chép `.db` thì phải chép cả `-wal`/`-shm`.
