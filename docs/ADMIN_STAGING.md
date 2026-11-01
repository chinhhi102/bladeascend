# Tự đưa mật khẩu quản trị lên STAGING

> Anh tự làm — Claude không đặt, không lưu, không đọc mật khẩu (CLAUDE.md luật 14).
> Tốn khoảng 3 phút. Chỉ cần làm lại khi muốn đổi mật khẩu.

Staging không đọc mật khẩu, chỉ đọc **hash bcrypt** (`ADMIN_PASSWORD_HASH`) từ tệp
`~/bladeascend-secrets/admin.env` **trên máy staging**. Tệp này nằm ngoài thư mục deploy
nên deploy không xoá và không lọt vào ảnh Docker. Chưa có tệp → `/admin` tắt, game vẫn chạy.

## Bước 1 — Tạo hash trên máy anh

Mở terminal ở thư mục dự án:

```bash
cd server && go run ./cmd/adminpass > /tmp/staging_admin.env && cd ..
```

Nó hỏi `Mật khẩu quản trị:` → gõ mật khẩu (≥ 12 ký tự, nên KHÁC mật khẩu `:9999`) → Enter.
Màn hình không in gì thêm là đúng (dòng hash nằm trong tệp).

## Bước 2 — Bọc hash trong nháy đơn

Docker compose coi `$2a$12$…` là biến và nuốt mất nếu không bọc (đã dính 1/11):

```bash
sed -i -E "s/^ADMIN_PASSWORD_HASH=(.*)$/ADMIN_PASSWORD_HASH='\1'/" /tmp/staging_admin.env
```

## Bước 3 — Đưa lên staging, xoá bản trên máy mình

```bash
set -a; . docs/private/deploy; set +a
ssh -i "$DEPLOY_KEY" "$DEPLOY_HOST" 'mkdir -p ~/bladeascend-secrets && chmod 700 ~/bladeascend-secrets'
scp -i "$DEPLOY_KEY" /tmp/staging_admin.env "$DEPLOY_HOST:bladeascend-secrets/admin.env"
ssh -i "$DEPLOY_KEY" "$DEPLOY_HOST" 'chmod 600 ~/bladeascend-secrets/admin.env'
rm /tmp/staging_admin.env
```

## Bước 4 — Khởi động lại container

```bash
ssh -i "$DEPLOY_KEY" "$DEPLOY_HOST" 'cd ~/bladeascend && docker compose up -d --force-recreate'
```

(Hoặc `deploy/staging.sh` — dòng "CẢNH BÁO … /admin TẮT" ở cuối phải biến mất.)

## Bước 5 — Kiểm tra

Vào <https://bladeascend.staging.maveragroups.com/admin> → đăng nhập bằng mật khẩu vừa gõ.

## Gặp lỗi

| Thấy gì | Làm gì |
|---|---|
| `/admin` báo "Trang quản trị chưa bật" | Tệp chưa tới nơi hoặc chưa làm bước 4. Chạy lại bước 3–4. |
| Log báo "không phải hash bcrypt hợp lệ" | Quên bước 2. Làm lại từ bước 1. |
| Quên mật khẩu | Làm lại từ bước 1 — ghi đè tệp cũ, không mất dữ liệu. |
| Đăng nhập xong bị đá ra | Bình thường sau khi đổi mật khẩu: mọi phiên cũ hết hạn. |

Bản `:9999` (máy LAN) làm khác — xem `docs/SETUP_ADMIN.md`.
