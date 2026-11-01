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

Dán **nguyên dòng đó** vào tệp `docs/private/admin.env` (tạo mới nếu chưa có):

```bash
cd server && go run ./cmd/adminpass > ../docs/private/admin.env
```

Rồi deploy lại:

```bash
tools/deploy.sh
```

Dòng cuối phải **không còn** cảnh báo "trang quản trị TẮT". Vào `http://<IP máy>:9999/admin`
và đăng nhập bằng mật khẩu vừa đặt.

## Vài điều cần biết

- `docs/private/` nằm trong `.gitignore`, nên mật khẩu không bao giờ lọt vào git. Deploy đọc
  `docs/private/admin.env` trước, rơi về `server/data/admin.env` nếu không có.
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

## Bật trang quản trị trên STAGING

Staging đọc hash từ tệp `~/bladeascend-secrets/admin.env` **trên máy staging**. Tệp này nằm
ngoài thư mục deploy, nên deploy không ghi đè và nó không lọt vào ảnh Docker. Anh tự làm,
Claude không đụng vào mật khẩu hay hash (CLAUDE.md luật 14).

1. Trên máy anh, tạo hash (nên dùng mật khẩu KHÁC với `:9999`):
   ```bash
   cd server && go run ./cmd/adminpass > /tmp/staging_admin.env
   ```
2. Đưa tệp lên staging rồi xoá bản trên máy mình:
   ```bash
   set -a; . docs/private/deploy; set +a
   ssh -i "$DEPLOY_KEY" "$DEPLOY_HOST" 'mkdir -p ~/bladeascend-secrets && chmod 700 ~/bladeascend-secrets'
   # Bọc hash trong nháy đơn: docker compose coi `$2a$12$…` trong env_file là biến và
   # nuốt mất → server báo "không phải hash bcrypt hợp lệ" (đã dính 1/11).
   sed -i -E "s/^ADMIN_PASSWORD_HASH=(.*)$/ADMIN_PASSWORD_HASH='\1'/" /tmp/staging_admin.env
   scp -i "$DEPLOY_KEY" /tmp/staging_admin.env "$DEPLOY_HOST:bladeascend-secrets/admin.env"
   ssh -i "$DEPLOY_KEY" "$DEPLOY_HOST" 'chmod 600 ~/bladeascend-secrets/admin.env'
   rm /tmp/staging_admin.env
   ```
3. Khởi động lại để container đọc tệp:
   ```bash
   ssh -i "$DEPLOY_KEY" "$DEPLOY_HOST" 'cd ~/bladeascend && docker compose up -d --force-recreate'
   ```
   Hoặc chạy lại `deploy/staging.sh`. Dòng "CẢNH BÁO … /admin TẮT" ở cuối phải biến mất.
4. Vào `https://bladeascend.staging.maveragroups.com/admin` và đăng nhập.

Đã làm 1/11 theo yêu cầu chủ dự án: mật khẩu staging nằm trong
`docs/private/staging-admin-password.txt`, còn hash trong `docs/private/staging-admin.env`.
Cả hai được gitignore và chmod 600. Đổi mật khẩu thì xoá hai tệp đó rồi làm lại từ bước 1.
