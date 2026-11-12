# Lỗi ĐỨNG HÌNH giữa trận — 12/11/2026 · **CHƯA SỬA**

> Một người chơi thật trên `:9999` đang chơi mượt 60 FPS thì **game đứng hình hẳn**, Chrome
> báo "Trang không phản hồi". Lỗi ném ra từ **hệ thống bộ lọc (filter) của PixiJS**. Đã
> khoanh được chỗ vỡ và chứng minh được hậu quả; **chưa dựng lại được** nên chưa biết nguyên
> nhân gốc. Tài liệu này để phiên sau bắt tiếp — đọc §5 trước nếu chỉ muốn sửa.

Bản đầy đủ (id phiên, ảnh người chơi gửi) ở `docs/private/loi-dung-hinh-12-11/` — ảnh có tên
người chơi thật nên không đưa vào đây (luật 15).

---

## 1. Chuyện gì đã xảy ra

Người chơi A, phiên `969839…`, Chrome 156 / Linux / GTX 1650, build
`20261111_2011_52adf2e`, khu **Trúc Lâm Mê Cung**, mức chất lượng **Cao**, ping LAN 1 ms.

| Giờ (12/11) | Việc |
|---|---|
| 09:30:21 | vào trận |
| 09:31 → 10:01 | 20 lần rời/quay lại tab · 6 lần đổi khu (lần cuối 09:53:52) |
| *trong phiên* | **cửa sổ bị đổi cỡ**: lúc vào `1920×993`, lúc sập `1717×820` |
| 10:16:45 | đột phá lên tầng 3 (`realm` 111 → 112) |
| 10:18:06 | **rớt** xuống tầng 1 (`realm` 110) — khớp HUD trong ảnh "Tiên khí 1/221" |
| 10:19:26 | mẫu đo ghi **2 lần hồi sinh trong 10 giây** (đang chết đi chết lại) |
| **10:19:36** | **mẫu đo cuối cùng** — xem bảng dưới |
| **10:19:39** | **lỗi JS** — từ đây không còn mẫu nào, cũng không có tín hiệu đóng tab |
| 10:20:08 | người chơi chụp màn hình: hộp thoại **"Trang không phản hồi"** |
| 10:35:26 | **người chơi còn lại vẫn chơi bình thường** → không phải lỗi server |

Mẫu đo cuối (10 giây trước khi chết) — **mọi thứ đều đẹp**, không có dấu hiệu quá tải:

| fps | khung p95 | khung max | long task | sprite | kiếm vẽ | người trong tầm | trong đó Tiên | texture |
|---|---|---|---|---|---|---|---|---|
| 59,9 | 18,5 ms | 30,8 ms | 0 | 5 217 | 870 | 8 | 6 | 274 MB |

## 2. Lỗi nằm ở đâu

```
Uncaught TypeError: Cannot set properties of undefined (setting 'uProjectionMatrix')
    at Oe.bind              (RenderTargetSystem-zDhw65Cj.js:182:28886)
    at Oe.push              (RenderTargetSystem-zDhw65Cj.js:182:29379)
    at I._setupFilterTextures (webworkerAll-BA5quiSG.js:56:5532)
    at I.push               (webworkerAll-BA5quiSG.js:56:2241)
    at Ce.execute           (RenderTargetSystem-zDhw65Cj.js:182:6068)
bối cảnh: {"phase":"match","texMB":274,"tex":480,"dpr":1,"canvas":"1717x820","chars":6,"fps":60,"q":"Cao"}
```

Giải mã bằng chính bundle đã deploy (`server/data/deploy/client/static/`): `Oe` là
`GlobalUniformSystem`, `I` là `FilterSystem`. Dòng ném lỗi là dòng cuối trong đoạn này
(`pixi.js/lib/rendering/renderers/shared/renderTarget/GlobalUniformSystem.mjs:57-60`):

```js
const uniformGroup = this._uniformsPool.pop() || this._createUniforms();
this._activeUniforms.push(uniformGroup);
const uniforms = uniformGroup.uniforms;
uniforms.uProjectionMatrix = globalUniformData.projectionMatrix;   // ← ném ở đây
```

Nghĩa là: **thứ Pixi lấy ra từ bể dùng lại của chính nó không phải một `UniformGroup`**
(không có trường `uniforms`). Đọc mã nguồn Pixi 8.22 thì chỉ có `reset()` đẩy đồ vào bể đó,
và nó chỉ đẩy lại đúng những gì đã lấy ra — tức **trạng thái nội bộ của Pixi đã hỏng từ
trước khung hình này**, lỗi ném ra chỉ là chỗ phát hiện ra, không phải chỗ gây ra.

### Vì sao hỏng một lần là chết hẳn

`pixi.js/lib/ticker/Ticker.mjs:123-130` đặt lịch khung kế tiếp **sau khi** `update()` trả về:

```js
this._tick = (time) => {
  this._requestId = null;
  if (this.started) {
    this.update(time);                       // ném lỗi ở đây →
    if (this.started && this._requestId === null && this._head.next) {
      this._requestId = requestAnimationFrame(this._tick);   // … dòng này không bao giờ chạy
    }
  }
};
```

`Ticker.update()` không có `try/catch`. Một ngoại lệ là vòng vẽ **không bao giờ quay lại**.
Khớp với số liệu: `telemetry.frame()` chạy theo từng khung (`client/src/telemetry/index.ts:147`)
nên telemetry tắt đúng giây đó; không có beacon `pagehide` nên tab **không bị đóng**, nó nằm đó
chết cứng.

### Chỗ CHƯA giải thích được

"Trang không phản hồi" nghĩa là luồng chính bị **chặn**, chứ không chỉ là ngừng vẽ. Vòng vẽ
chết thì trang vẫn phải bấm được. Đã soi hết các vòng `while` trong `client/src` — không cái
nào lặp vô hạn được. Nghi (chưa chứng minh): ngữ cảnh WebGL/driver kẹt ở trạng thái dở dang
mà bộ lọc để lại giữa chừng (`push` xong, `pop` không bao giờ chạy, texture mượn của
`TexturePool` không bao giờ trả).

## 3. Đã loại trừ

- **Không phải rò rỉ texture (B23)**: lúc sập 274 MB và đứng yên cả phiên — bản vá 11/11 đang
  chạy đúng (so với 1 423 MB hôm 11/11 trước khi sửa).
- **Không phải server**: người chơi còn lại không sao. 941 dòng `tick chậm` trong
  `server/data/bladeascend.log` giờ 09–10 là **nhiễu do máy bận** — lúc kiểm `tools/cleanup.sh`
  báo **4,80 tải/nhân**, vượt xa ngưỡng 1,5 của luật 16. Không rút kết luận gì từ đống đó.
- **Không phải lỗi `null.texture` hồi 30/10**: khác chữ ký hoàn toàn.
- **Không phải texture lọc khổng lồ** (kiểu 1 GB đã gặp ở `mapview.ts`): cả ba filter của game
  đều để `clipToViewport` mặc định = bật, nên Pixi cắt vùng lọc theo khung nhìn
  (`FilterSystem._calculateFilterBounds`).
- **Không phải sóng xung kích lúc đột phá**: filter đó chỉ gắn lên camera khi **chính mình**
  phi thăng (`scene.ts:1107-1111`); lần đổi cảnh giới gần nhất là **rớt**, 90 giây trước.

## 4. Đã thử dựng lại — chưa ra

Cả hai bài đều chạy sạch trên đúng pixi 8.22.0 + pixi-filters 6.1.5 của dự án:

1. Bộ lọc lồng nhau + `renderer.resize` qua 6 cỡ + đổi `resolution` (đúng kiểu
   [pixijs#12251](https://github.com/pixijs/pixijs/issues/12251)) → không lỗi.
2. `tools/browsertest/filtercrash.mjs` — dựng đúng cây cảnh của game (camera chứa
   ground/decor **dùng chung một** `ColorMatrixFilter`, lớp vfx có `AdvancedBloomFilter`,
   thỉnh thoảng gắn `ShockwaveFilter` lên camera, `filterArea` quét cả ra ngoài màn hình,
   resize và đổi resolution xen giữa) × 2 000 khung → không lỗi.

Việc đầu tiên của phiên sau nên là **nới bài 2 cho giống thật hơn**: chạy client thật trên
server debug, bắn lệnh `{t:'dbg', cmd:'realm', n:1}` (xem `tools/browsertest/dbggate.mjs`) để
ép đột phá liên tục, nhiều người Tiên giới trong tầm nhìn, vừa chơi vừa kéo đổi cỡ cửa sổ.

## 5. Đề nghị sửa (theo thứ tự)

1. **Chốt chặn + đồng hồ canh** — `client/src/main.ts:155`. Bọc `try/catch` quanh thân ticker,
   gửi lỗi về rồi **chạy lại vòng vẽ**; thêm `setInterval` canh "quá 2 giây không có khung
   nào" → gửi sự kiện `render_loop_chet`. Việc này đúng dù nguyên nhân gốc là gì: nó biến
   "đứng hình phải F5" thành "khựng một nhịp", và lần tái diễn sau sẽ có bằng chứng cả về
   đoạn bị treo.
2. **Bỏ hai cách dùng filter bất thường** (rẻ, loại hai nghi can):
   - `client/src/render/mapview.ts:336-337` — `ground` và `decor` đang dùng **chung một
     instance** `ColorMatrixFilter`; tách thành hai instance riêng.
   - `client/src/render/breakthrough.ts:62` — gắn filter lên `camera` mà camera **không có**
     `filterArea`; đặt `filterArea` bằng khung màn hình trong lúc sóng xung kích chạy.
3. **Nâng PixiJS 8.22.0 → 8.23.0**: có [#12259](https://github.com/pixijs/pixijs/pull/12259)
   (filter báo texture đã huỷ sau khi resize) và
   [#12369](https://github.com/pixijs/pixijs/pull/12369) (bể texture khớp cỡ màn hình). Nâng
   xong phải đo lại hiệu năng theo luật 12.
4. **Thêm bối cảnh vào telemetry lỗi** (`client/src/telemetry/index.ts:310` `errorContext`):
   số lần resize và cỡ canvas lúc vào, giây từ lần đổi khu gần nhất, danh sách filter đang
   bật. Thiếu mấy số này nên lần này phải suy ngược từ bundle.

## 6. Cách tra lại số liệu

Dữ liệu nằm trong `server/data/telemetry.db` (bảng `events`, `samples`, `sessions`). Trang
admin đã có mục lỗi (`server/internal/telemetry/errors.go` — `RecentErrors`, gộp theo chữ ký).
Muốn xem thô thì viết công cụ tạm **ngoài repo** rồi xoá — đừng để trong `server/cmd/`, ngày
12/11 một tệp tạm như thế đã bị commit của phiên khác gom vào (đã gỡ ở `a605aba`).

```sql
-- mọi lỗi client 14 ngày gần đây, kèm máy và build
SELECT e.ts, e.kind, e.data, s.browser, s.os, s.gpu, s.build
FROM events e LEFT JOIN sessions s ON s.session_id = e.session_id
WHERE e.kind IN ('error','webgl_context_lost','gpu_pressure')
ORDER BY e.ts DESC LIMIT 50;
```

Lưu ý khi đọc: `events.ts` tính bằng **mili giây**, còn `sessions.started_at` và
`feedback.at` tính bằng **giây** — đọc nhầm đơn vị là ra năm 1970.
