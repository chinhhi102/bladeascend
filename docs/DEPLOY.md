# Đưa game lên mạng (VPS + HTTPS bằng Caddy)

Mục tiêu: bạn bè mở `https://kiemxoay.ten-mien-cua-ban.com` là chơi được.
Tổng chi phí: VPS ~5 USD/tháng + tên miền ~10 USD/năm (hoặc subdomain miễn phí, xem bước 2).

## 0. Chơi thử trong mạng nhà (LAN) trước — không tốn tiền
1. Máy tính và điện thoại cùng Wi-Fi.
2. Trên máy tính chạy 2 cửa sổ terminal:
   - `cd server && go run ./cmd/server`
   - `cd client && npm run dev`
3. Cửa sổ `npm run dev` in ra dòng `Network: http://192.168.x.x:5173` → mở địa chỉ đó trên điện thoại.
   (Không thấy dòng Network? Windows có thể hỏi cho phép Node qua tường lửa → bấm Allow.)

## 1. Thuê VPS
- Chọn nhà cung cấp bất kỳ: Vultr, DigitalOcean, Hetzner, Linode, hoặc VPS Việt Nam (ping thấp hơn cho bạn bè VN).
- Cấu hình đủ dùng: **1 vCPU, 1 GB RAM, Ubuntu 24.04**. Chọn vị trí gần người chơi (Singapore / Việt Nam).
- Sau khi tạo xong, nhà cung cấp cho bạn **địa chỉ IP** (ví dụ `203.0.113.10`) và mật khẩu root.

## 2. Trỏ tên miền về VPS
- Mua tên miền (Namecheap, Cloudflare, Mắt Bão, PA Việt Nam…) → vào phần **DNS** → thêm bản ghi:
  - Loại `A`, Tên `kiemxoay` (hoặc `@`), Giá trị = IP của VPS.
- Không muốn mua tên miền: dùng dịch vụ miễn phí như duckdns.org (tạo `tenban.duckdns.org` trỏ về IP).
- Đợi 5–30 phút cho DNS cập nhật. Kiểm tra: `ping kiemxoay.ten-mien-cua-ban.com` ra đúng IP là được.

## 3. Build game trên máy của bạn
Trong thư mục dự án:
```bash
cd client && npm run build && cd ..
cd server && GOOS=linux GOARCH=amd64 go build -o kiemxoay ./cmd/server && cd ..
```
Windows PowerShell thay dòng thứ hai bằng:
```powershell
cd server; $env:GOOS="linux"; $env:GOARCH="amd64"; go build -o kiemxoay ./cmd/server; Remove-Item Env:GOOS, Env:GOARCH; cd ..
```

## 4. Chép lên VPS
```bash
ssh root@203.0.113.10 "mkdir -p /opt/kiemxoay/server/data /opt/kiemxoay/client"
scp server/kiemxoay root@203.0.113.10:/opt/kiemxoay/server/
scp -r config maps assets root@203.0.113.10:/opt/kiemxoay/
scp -r client/dist root@203.0.113.10:/opt/kiemxoay/client/
```
(Lần sau cập nhật game: build lại rồi chạy lại 4 lệnh này, sau đó `systemctl restart kiemxoay`.)

## 5. Cài Caddy (tự lấy chứng chỉ HTTPS miễn phí)
SSH vào VPS: `ssh root@203.0.113.10` rồi chạy:
```bash
apt update && apt install -y caddy
cat > /etc/caddy/Caddyfile <<'EOF'
kiemxoay.ten-mien-cua-ban.com {
    encode gzip
    reverse_proxy 127.0.0.1:8080
}
EOF
systemctl reload caddy
```
Caddy tự chuyển WebSocket (`/ws`) và tự gia hạn HTTPS. Nhớ thay tên miền cho đúng.

## 6. Chạy game như một dịch vụ (tự khởi động lại khi VPS reboot)
```bash
useradd --system --home /opt/kiemxoay kiemxoay || true
chown -R kiemxoay:kiemxoay /opt/kiemxoay
cat > /etc/systemd/system/kiemxoay.service <<'EOF'
[Unit]
Description=Kiem Xoay Tu Tien game server
After=network.target

[Service]
User=kiemxoay
WorkingDirectory=/opt/kiemxoay/server
ExecStart=/opt/kiemxoay/server/kiemxoay -addr 127.0.0.1:8080 -debug=false \
  -config /opt/kiemxoay/config/game.json -maps /opt/kiemxoay/maps \
  -assets /opt/kiemxoay/assets -client /opt/kiemxoay/client/dist \
  -db /opt/kiemxoay/server/data/leaderboard.db
Restart=always
RestartSec=2

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable --now kiemxoay
systemctl status kiemxoay --no-pager
```
`-debug=false` **tắt phím test** (E/R/K/X) để người chơi không gian lận.

## 7. Mở tường lửa
```bash
ufw allow OpenSSH && ufw allow 80 && ufw allow 443 && ufw --force enable
```
Không mở cổng 8080 — chỉ Caddy được nói chuyện với game.

## 8. Kiểm tra
- Mở `https://kiemxoay.ten-mien-cua-ban.com` → thấy trang chủ, ổ khóa HTTPS xanh.
- Xem log game: `journalctl -u kiemxoay -f` · log Caddy: `journalctl -u caddy -f`.
- Xem có bao nhiêu phòng: `https://kiemxoay.ten-mien-cua-ban.com/api/rooms`.

## 9. Sao lưu bảng xếp hạng
Bảng xếp hạng nằm trong 1 file: `/opt/kiemxoay/server/data/leaderboard.db`. Thỉnh thoảng tải về:
`scp root@203.0.113.10:/opt/kiemxoay/server/data/leaderboard.db ./backup-leaderboard.db`

## Lỗi thường gặp
| Triệu chứng | Cách sửa |
|---|---|
| Trình duyệt báo không an toàn / không vào được | DNS chưa trỏ đúng IP, hoặc chưa mở cổng 80/443. Xem `journalctl -u caddy` |
| Vào được trang nhưng kẹt "Đang kết nối máy chủ" | Game server chưa chạy: `systemctl status kiemxoay` |
| "Client chưa được build" | Chưa chép `client/dist` (bước 3–4) |
| Lag khi đông người | Nâng VPS lên 2 vCPU; chọn vị trí gần người chơi |
