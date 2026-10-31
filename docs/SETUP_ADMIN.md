# Bật trang quản trị trên `:9999`

> Việc này **anh phải tự làm một lần** — mật khẩu không được nằm trong code hay trong git.
> Chưa làm thì trang `/admin` tắt hoàn toàn: A1–A4 (dashboard, bản đồ live, hòm thư góp ý,
> người chơi, bảng xếp hạng) đều không vào được.

## Làm gì

```bash
cd server && go run ./cmd/adminpass
```

Nó hỏi mật khẩu, rồi in ra một dòng như:

```
ADMIN_PASSWORD_HASH=$2a$12$....
```

Dán **nguyên dòng đó** vào tệp `server/data/admin.env` (tạo mới nếu chưa có):

```bash
cd server && go run ./cmd/adminpass > data/admin.env
```

Rồi deploy lại:

```bash
tools/deploy.sh
```

Dòng cuối phải **không còn** cảnh báo "trang quản trị TẮT". Vào `http://<IP máy>:9999/admin`
và đăng nhập bằng mật khẩu vừa đặt.

## Vài điều cần biết

- `server/data/` nằm trong `.gitignore`, nên mật khẩu không bao giờ lọt vào git.
- Trong tệp chỉ có **hash bcrypt**, không có mật khẩu. Ai đọc được tệp cũng không đăng nhập
  được — nhưng cũng đừng để lộ, vì hash dò được bằng máy mạnh.
- Quên mật khẩu thì chạy lại `adminpass`, ghi đè tệp, deploy lại. Không mất dữ liệu gì.
- Đổi mật khẩu sẽ **đá mọi phiên đang mở** ra, vì phiên gắn với hash cũ.

## Vì sao trước đây nó tắt mà không ai biết

`tools/deploy.sh` truyền biến `KX_ADMIN_PASS` — một tên server **không bao giờ đọc**;
`admin.NewAuth` chỉ nhìn `ADMIN_PASSWORD_HASH`. Server vẫn chạy bình thường, vẫn trả HTTP
200 cho mọi trang game, chỉ in một dòng log "trang quản trị: tắt" rồi thôi.

Từ 31/10 `deploy.sh` đọc `server/data/admin.env` và **in cảnh báo ở bước cuối** nếu trang
vẫn tắt — để lần sau không ai phải đọc log mới biết.
