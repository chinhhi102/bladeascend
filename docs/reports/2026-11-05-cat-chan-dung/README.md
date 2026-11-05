# 05/11 — Cắt chân dung vào khung tròn (ô đại diện HUD §12.4 · bong bóng mặt §12.5)

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

Tool dò cổ bằng alpha (hàng hẹp nhất phía trên thân) rồi lấy phần trên cổ làm đầu. Máy dò
sai ở ảnh toàn thân và mặt có râu → bảng `OVERRIDE` trong tool là số đo bằng mắt, luôn
thắng máy. Hiện 28/40 ảnh do máy dò, 12 ảnh đo tay.

Hai nơi dùng khung này:
- `client/src/ui/hud.ts` — đặt `background-size`/`background-position` theo khung đầu.
- `client/src/render/emote.ts` — đặt `scale`/vị trí sprite sao cho cạnh khung đầu vừa đúng
  đường kính vòng tròn.

Thêm nhân vật mới thì chạy lại tool, xem tờ ảnh kiểm, ảnh nào lệch thì thêm vào `OVERRIDE`.

## Trước / sau

![trước sau](truoc-sau.png)

Cột trái: luật cũ. Cột phải: khung đầu riêng. Mỗi nhân vật 2 ảnh (`neutral`, `panic`).

## Chưa làm

- Chân dung tròn ở **bảng xếp hạng** (`.rb-face`, 28–40 px) và màn chọn nhân vật vẫn hiện
  **cả ảnh** thu nhỏ (`object-fit: cover/contain`) — không bị cắt mất mặt, nhưng ở cỡ đó
  thì mặt bé tí. Dùng chung khung đầu này sẽ rõ hơn; chờ chủ dự án quyết.
