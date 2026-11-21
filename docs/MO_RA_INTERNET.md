# Mở game ra Internet cho bạn bè vào thử

Hiện `:9999` chỉ chạy trong mạng LAN. Tệp này nói cách mở nó ra Internet bằng **Forward a
Port** của VS Code (đúng cái nút trong ảnh chụp), và những gì phải làm TRƯỚC khi bấm.

> **Nói trước cho khỏi nhầm:** cách này để **bạn bè vào thử**, không phải để phát hành. Xem
> mục "Vì sao đây chưa phải bản phát hành" ở cuối.

## 1. Bật công tắc an toàn (BẮT BUỘC, làm trước)

Tunnel nối từ dịch vụ của Microsoft vào `127.0.0.1` của máy này. Nhìn từ server, **mọi
người trên Internet đều có địa chỉ 127.0.0.1** — nên mọi cửa đang khoá bằng "chỉ cho
localhost" bỗng mở toang:

| cửa | bình thường | qua tunnel, nếu không làm gì |
|---|---|---|
| `/debug/pprof` | chỉ máy này | **cả Internet** — hồ sơ heap chứa nguyên trạng bộ nhớ tiến trình: tên người chơi, nội dung gói tin |
| `/metrics` | chỉ mạng nội bộ | **cả Internet** |
| `/admin` | mật khẩu | mật khẩu (không đổi) |

Công tắc `EXPOSED=true` tắt niềm tin đó: không tin loopback nữa, `/metrics` đóng, `/debug/pprof`
không đăng ký kể cả khi có cờ `-pprof`.

```bash
cd ~/git/bladeascend
EXPOSED=true tools/deploy.sh start
```

Xem log xác nhận:

```bash
grep -a "EXPOSED" server/data/bladeascend.log | tail -2
```

Phải thấy: `EXPOSED=true: server coi như đang mở ra Internet…`

## 2. Đổi mật khẩu quản trị

`/admin` vẫn mở ra Internet và **chỉ có mật khẩu che**. Trang đó xem được dữ liệu người chơi
và bật được quyền lệnh test.

```bash
cd server && go run ./cmd/adminpass > ../docs/private/admin.env
```

Rồi khởi động lại (`EXPOSED=true tools/deploy.sh start`). Mật khẩu KHÔNG bao giờ ghi vào
repo — `docs/private/` đã nằm trong `.gitignore` (luật 14).

## 3. Bấm nút trong ảnh

1. Trong VS Code, bảng **Ports** → **Forward a Port**.
2. Gõ `9999`, Enter. VS Code tạo một địa chỉ dạng
   `https://<ngẫu-nhiên>-9999.<vùng>.devtunnels.ms`.
3. **Quan trọng:** mặc định cổng là **Private** — chỉ mình bác vào được. Chuột phải vào dòng
   cổng → **Port Visibility** → **Public**. Không làm bước này thì bạn bè bấm link ra trang
   đăng nhập GitHub.

WebSocket chạy được qua tunnel: client tự chọn `wss://` khi trang mở bằng `https://`
(`client/src/net/conn.ts`), không phải sửa gì.

## 4. Kiểm trước khi gửi link

```bash
# thay <URL> bằng địa chỉ devtunnels
curl -s -o /dev/null -w "%{http_code}\n" <URL>/healthz        # phải 200
curl -s -o /dev/null -w "%{http_code}\n" <URL>/metrics        # phải 403
curl -s -o /dev/null -w "%{http_code}\n" <URL>/debug/pprof/   # phải 404
```

Ba dòng này là cách duy nhất biết công tắc đã ăn. `200 / 403 / 404` mới gửi link đi.

Lần đầu mở link, Microsoft có thể chèn một trang cảnh báo "bạn sắp vào một tunnel" — bấm
tiếp là vào game. Không bỏ được trang đó từ phía mình.

## 5. Tắt khi không dùng

Chuột phải dòng cổng → **Stop Forwarding Port**. Tunnel sống theo VS Code: đóng VS Code hoặc
tắt máy là link chết.

Sau khi tắt, bỏ `EXPOSED` để pprof và `/metrics` dùng lại được:

```bash
tools/deploy.sh start
```

## Vì sao đây chưa phải bản phát hành

| | |
|---|---|
| Link đổi mỗi lần | Trừ khi tạo tunnel có tên; bạn bè phải xin link mới mỗi lần |
| Phải mở VS Code | Đóng VS Code là game offline |
| Qua máy chủ trung gian | Mọi gói đi vòng qua dịch vụ của Microsoft — ping cao hơn hẳn LAN, và netcode của game đo ping để nội suy |
| Giới hạn băng thông | Dev tunnel có hạn mức; một phòng đông có thể chạm |
| Máy này là máy làm việc | Máy đã khởi động lại một lần làm game offline 44 tiếng (21/11). Có `@reboot` rồi, nhưng máy bận thì game giật — nút thắt B27 |

Muốn phát hành thật thì đi theo `docs/_raw/DEPLOY_K8S_SPEC.md` (§1 đã xong; §3/§4 chưa làm),
hoặc đơn giản hơn là một VPS có HTTPS như milestone M7 trong spec.
