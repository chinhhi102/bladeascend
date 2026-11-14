# Bước 1 tối ưu vòng kiếm — ảnh trước/sau (13/11/2026)

Hai ảnh này là **cảnh thử cố định** `?scene=testbed` (`client/src/render/testbed.ts`), chụp
bằng `tools/browsertest/scenediff.mjs`. Không mạng, không bot, không thời gian, camera ghim,
khung 1280×720 ở `resolution` 1 — nên hai lần chụp khác nhau chỉ có thể khác nhau vì CODE.

| | |
|---|---|
| `truoc.png` | `swordring.ts` ở `26b13cc` (bể hạt khai `dynamic`) |
| `sau.png` | `swordring.ts` ở `ca3c42f` (bể hạt khai TĨNH + bỏ 3 lượt `clear()` mỗi khung) |

## Hình: không đổi

```
lệch trung bình : 0.001/255
lệch lớn nhất   : 50/255
pixel lệch > 16 : 19 trên 921 600 (0,002%)
```

Con số này **thấp hơn cả nhiễu của chính bài đo**: chụp hai lần trên CÙNG một bản build cho
0,002/255 và 32 pixel. Nói cách khác, bước 1 không đổi một nét nào.

Để so sánh, đối chứng âm (tắt sạch vòng kiếm) cho **36/255 và 66% pixel** — bài đo thừa sức
bắt một thay đổi thật.

## Tốc độ: nhanh hơn 50–60%

Cùng cảnh, cùng **23 lô vẽ**, cùng **475 nút** cây cảnh (`tools/browsertest/scenebench.mjs`):

| CPU | Pixi vẽ ms trước | sau | FPS trước → sau |
|---|---:|---:|---|
| 1× | 2,52 | **1,27** | 60 → 60 |
| 4× | 8,92 | **3,30** | 60 → 60 |
| 6× | 14,29 | **5,70** | **54 → 60** |

## Vì sao phải có cảnh thử mới nói được hai điều trên

Trước khi có nó, mọi phép so đều vô nghĩa vì cảnh thật đổi theo bot:

- `weakperf` trên cảnh thật, cùng build, cùng máy rảnh: **185 lô vẽ / 47 FPS** rồi
  **447 lô vẽ / 23 FPS**.
- `ringcost` trên cảnh thật, cùng build, ba lượt: giá vòng kiếm **19,79 · 0,55 · −0,40 ms**.

Chi tiết ba cách so ảnh đã thất bại trước đó nằm ở đầu `client/src/render/testbed.ts`.

---

## Đổi ảnh mốc — 13/11/2026 (mặt nạ nền 64²/ô)

`testbed-ref.png` được chụp LẠI sau hai thay đổi của §13.10:

| | |
|---|---|
| vật cản `maxPx` 512 → 384 | `tools/asset-pipeline/groups.mjs`, hạng mục ngốn VRAM thứ hai |
| mặt nạ trộn nền 128² → 64² mỗi ô | `client/src/render/groundshader.ts`, atlas mặt nạ 4 MB → 1 MB |

Ảnh mốc CŨ giữ lại ở `testbed-ref-truoc-nen64.png` để đối chiếu.

```
lệch trung bình : 1.822/255   (ngưỡng 0.3)
lệch lớn nhất   : 109/255
pixel lệch > 16 : 10 946 trên 921 600 (1,188%)
```

**Vì sao chấp nhận lệch này mà không coi là hỏng hình.** Bản đồ lệch 16×9 cho thấy chênh
lệch rải ĐỀU trên nền (2–5/255 mỗi ô), không dồn vào vật cản hay nhân vật — tức thủ phạm là
mặt nạ nền, không phải ảnh bị nén hỏng. Mặt nạ trộn là nhiễu thủ tục sinh theo toạ độ texel:
hạ nửa độ phân giải thì hoa văn được *lấy mẫu lại*, mép mềm hơn vài phần trăm. Tần số sóng
đã nhân đôi để mảng loang giữ NGUYÊN cỡ trong toạ độ thế giới — thiếu bước bù đó thì hoa văn
to gấp đôi và cả bản đồ đổi mặt. Đã xem ảnh tận mắt: cỏ, đá, vệt đất, bốn vòng kiếm và lớp
hào quang Tiên đều y nguyên.

Đổi được gì: hạng mục **Nền 7,33 → 3,00 MB**, và tổng cảnh đông **51,16 → 46,32 MB**, tức
lần đầu lọt trần 50 MB của §13.10.

Đây là lần đổi ảnh mốc có chủ ý; luật vẫn là **không đổi ảnh mốc để bài đo xanh**, chỉ đổi
khi thay đổi hình là cố ý, đã soi mắt thường, và lý do ghi ngay tại đây.

---

## Đổi ảnh mốc lần 2 — 13/11/2026 (vật cản hiện ra)

Ảnh mốc cũ ở `testbed-ref-truoc-vatcan.png`. Ba thay đổi:

1. **Vật cản trong cảnh thử giờ mới VẼ RA.** Chúng khai từ đầu nhưng không bao giờ hiện, và
   lý do đáng ghi: hình tròn trong `GameMap` phải có cờ `c: true`. Thiếu nó thì `obstacleArt`
   coi là hình chữ nhật, đi đường xếp gạch và vẽ ra đúng con số không — chỉ còn vệt bóng dưới
   chân, không một lời cảnh báo. Ảnh mốc cũ vì vậy chưa gác phần vẽ vật cản.
2. `bamboo_thin` đổi thành `brick`: chỉ 10 trong 16 loại vật cản nằm trong nhóm `boot`, mà
   cảnh thử chỉ tải `home` + `boot` + một nhân vật. Để một loại thuộc `zone_*` trong cảnh là
   ảnh mốc có hình TẠM — và bài so ảnh sẽ canh hình tạm thay vì canh art.
3. KTX2 nhân sẵn alpha (xem PERF_LOG lượt 52) — đổi chút ít ở viền mọi ảnh nén.

Khối tròn xám xanh bên trái là **art thật của `boulder`** (tảng đá nhìn từ trên), không phải
hình tạm; tôi đã nhầm nó một lượt nên ghi lại đây.

Cảnh thử nay có thêm chế độ **ĐÔNG**: `?scene=testbed&crowd=1` — 12 nhân vật và 64 vật cản
đủ 10 loại, xếp lưới cố định. Dùng để đếm lô vẽ (`drawbreak.mjs --testbed --crowd`), KHÔNG
dùng cho `scenediff`.

---

## Đổi ảnh mốc lần 3 — 14/11/2026 (một font bitmap cho cả game)

Ba font bitmap (`kx-name` 36/800, `kx-realm` 26/700, `kx-title` 26/700 nghiêng) gộp thành
MỘT (`kx-text` 28/800), và nhãn số kiếm của Bó Kiếm đổi từ `Text` sang `BitmapText`.

```
lệch trung bình : 0.155/255   (ngưỡng 0.3 — ĐẠT)
pixel lệch > 16 : 3 256 (0,353%, ngưỡng 0,1%)
```

Bản đồ lệch 16×9 cho thấy chênh lệch **chỉ nằm ở bốn ô bảng tên**, mọi ô khác bằng 0. Đã cắt
vùng bảng tên phóng 2× để so tận mắt: nét chữ, độ đậm và vị trí y như cũ. Hạng mục "Chữ/số
bay" đi từ **8,78 xuống 1,70 MB** (trần 2).
