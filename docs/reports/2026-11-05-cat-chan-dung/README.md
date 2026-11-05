# 05/11 — Cắt chân dung lấy phần mặt (HUD §12.4 · bong bóng mặt §12.5 · bảng xếp hạng §11.5 · màn chọn)

## Lỗi

Ảnh chụp 05/11: ô đại diện góc trên trái và bong bóng mặt chỉ thấy **cái nón với hai con
mắt**, không thấy hết khuôn mặt.

Nguyên nhân: cả hai chỗ cắt ảnh bằng **một luật chung** `background: center top / 180%` —
lấy nửa trên ảnh, phóng 1,8 lần. Luật đó chỉ đúng nếu mặt nằm giữa và chiếm gần hết mép
trên. Bộ chân dung thật là ảnh **bán thân/toàn thân**, đầu nhỏ, lệch trái lệch phải tuỳ
ảnh, nên khung tròn hứng trúng trán với nón. Spec §12.4 ghi "cắt từ đầu tới vai" — luật cũ
không làm được việc đó.

## Sửa

Mỗi chân dung có **khung đầu riêng** `[trái, trên, cạnh]` (tỉ lệ cạnh ảnh), để trong
`client/src/render/headbox.ts`, sinh bằng:

```
node tools/portrait-headbox.mjs                       # ghi lại headbox.ts
node tools/portrait-headbox.mjs --sheet /tmp/box.png  # kèm tờ ảnh kiểm bằng mắt
```

Tool dò cổ bằng alpha (hàng hẹp nhất phía trên thân) rồi lấy phần trên cổ làm đầu. **Cằm
thì không tin vào cổ**: với mặt có râu, hàm bạnh hay cổ rụt, chỗ hẹp nhất là quai hàm nằm
TRÊN cằm — lấy đúng chỗ đó là cắt mất mồm. Đầu của bộ art này cao khoảng 1,35 lần bề ngang
nên đáy khung suy từ bề ngang, chỉ dùng cổ khi nó còn thấp hơn thế.

Còn lại 4 ảnh máy dò sai, đều là **ảnh toàn thân** (`co_ba_ganh/neutral`,
`ong_ba_vong/evil`, `shipper/evil`, `su_phu_mi/evil`): chỗ hẹp nhất rơi xuống eo hoặc khe
hai chân nên khung ôm cả người. Bốn ảnh đó đo bằng mắt, để trong bảng `OVERRIDE` của tool
và luôn thắng máy. 36/40 ảnh còn lại do máy dò.

Bốn nơi dùng khung này — tất cả gọi chung `headCrop`/`headBox`, không ai tự chép công thức:
- `client/src/ui/hud.ts` — ô đại diện HUD 64 px (§12.4).
- `client/src/render/emote.ts` — bong bóng mặt, sprite Pixi (§12.5).
- `client/src/ui/roomboard-parts.ts` — chân dung tròn bảng xếp hạng 28–40 px (§11.5).
- `client/src/ui/dom.ts` — ảnh nhân vật ở màn chọn 64 px + khung xem trước 84 px.

Thêm nhân vật mới thì chạy lại tool, xem tờ ảnh kiểm, ảnh nào lệch thì thêm vào `OVERRIDE`.

## Trước / sau

![trước sau](truoc-sau.png)

Cột trái: luật cũ. Cột phải: khung đầu riêng. Mỗi nhân vật 2 ảnh (`neutral`, `panic`).

## Không đổi

Màn **Tổng kết** (bóng ma `portrait_cry`) và **màn tải** (người chạy) vẫn dùng cả tấm ảnh:
ở đó cố ý hiện nguyên nhân vật chứ không phải cái mặt trong ô nhỏ.
