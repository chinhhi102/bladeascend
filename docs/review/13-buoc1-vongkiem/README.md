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
