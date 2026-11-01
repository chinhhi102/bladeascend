# Đổi tên hạ tầng staging sang `bladeascend` — lệnh từng bước

Chủ dự án chốt 1/11/2026. Phía repo **đã sửa xong và đã commit**; tệp này là phần phải chạy
trên MÁY CHỦ staging, vì đổi ở repo mà không đổi trên máy đó thì `deploy/staging.sh` sẽ
rsync vào một thư mục rỗng mới và để nguyên bản cũ đang chạy.

## Những cái tên phải đổi

Lần đổi tên 31/10 chỉ nhắm tới thư mục `~/kiemxoaytutien`. Thật ra còn **năm** cái tên nữa
nằm trong cùng một tệp đã được `brand-check` miễn trừ, nên không ai nhìn thấy chúng:

| Thứ | Tên cũ | Tên mới |
|---|---|---|
| Thư mục trên máy chủ | `~/kiemxoaytutien` | `~/bladeascend` |
| Dự án compose | `kiemtien` | `bladeascend` |
| Container | `kiemtien-web` | `bladeascend-web` |
| **Volume dữ liệu** | `kiemtien-data` | `bladeascend-data` |
| Tệp cấu hình Caddy | `kiemtien-staging.caddy` | `bladeascend-staging.caddy` |
| Tên miền | `kiemtien.staging.maveragroups.com` | `bladeascend.staging.maveragroups.com` |

> **Volume là chỗ nguy hiểm duy nhất.** Nó chứa `/data` — bảng xếp hạng và telemetry của
> staging. Docker KHÔNG đổi tên volume được; phải chép sang volume mới. Bỏ bước 3 là mất
> sạch dữ liệu đó mà không có cảnh báo nào.
>
> Tên miền mới nằm trong wildcard `*.staging.maveragroups.com` nên **không phải sửa DNS**.

## Chạy từng bước

Mở terminal trên máy mình, nạp biến môi trường rồi vào máy chủ:

```bash
cd ~/git/bladeascend
set -a; . docs/private/deploy; set +a
ssh -i "$DEPLOY_KEY" "$DEPLOY_HOST"
```

Từ đây là **chạy trên máy chủ staging**.

### 1. Xem hiện trạng (chưa đổi gì)

```bash
docker ps --filter name=kiemtien
docker volume ls | grep kiemtien
ls -d ~/kiemxoaytutien
```

Mong đợi: thấy container `kiemtien-web` đang chạy, volume `kiemtien-data`, thư mục tồn tại.

### 2. Dừng bản cũ

```bash
cd ~/kiemxoaytutien && docker compose down
```

Dừng chứ **chưa xoá volume** — `down` không đụng volume nếu không có cờ `-v`. Đừng bao giờ
thêm `-v` ở bước này.

### 3. Chép dữ liệu sang volume mới

```bash
docker volume create bladeascend-data
docker run --rm \
  -v kiemtien-data:/from \
  -v bladeascend-data:/to \
  alpine sh -c 'cd /from && cp -a . /to/ && echo CHEP-XONG && ls -l /to'
```

Phải thấy `CHEP-XONG` và danh sách tệp (ít nhất là tệp `.db`). Không thấy thì **dừng lại**,
đừng làm tiếp.

### 4. Đổi tên thư mục

```bash
mv ~/kiemxoaytutien ~/bladeascend
ls -d ~/bladeascend
```

### 5. Gỡ tệp Caddy cũ

```bash
sudo rm -f /opt/astery/caddy/conf.d/kiemtien-staging.caddy
docker exec docker-caddy-1 caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile \
  && docker exec docker-caddy-1 caddy reload --config /etc/caddy/Caddyfile --adapter caddyfile
```

Làm bước này **trước** khi deploy bản mới: để lại tệp cũ thì Caddy giữ một site block trỏ
tới container `kiemtien-web` đã chết, và nó sẽ báo lỗi mỗi lần validate.

### 6. Thoát, rồi deploy từ máy mình

```bash
exit
cd ~/git/bladeascend
set -a; . docs/private/deploy; set +a
deploy/staging.sh
```

Dòng cuối phải in: `deployed: https://bladeascend.staging.maveragroups.com`

### 7. Kiểm

```bash
curl -sI https://bladeascend.staging.maveragroups.com | head -1
curl -s  https://bladeascend.staging.maveragroups.com/api/rooms | head -c 200
```

Mong đợi `HTTP/2 200` và một mẩu JSON danh sách phòng. Rồi mở trang bằng trình duyệt, vào
chơi thử một ván, xem bảng xếp hạng có còn dữ liệu cũ không (nếu có — bước 3 đã đúng).

### 8. Dọn volume cũ — CHỈ KHI bước 7 đã xong xuôi

```bash
docker volume rm kiemtien-data
```

Đừng chạy bước này cùng ngày nếu chưa chắc. Volume cũ không tốn gì đáng kể, và nó là đường
lùi duy nhất nếu bước 3 chép thiếu.

## Nếu hỏng giữa chừng

Mọi bước đều lùi được, vì bản cũ chỉ bị *dừng* chứ không bị xoá:

```bash
mv ~/bladeascend ~/kiemxoaytutien
cd ~/kiemxoaytutien && git -C ~/git/bladeascend show HEAD~1:deploy/docker-compose.yml > docker-compose.yml
docker compose up -d
```

rồi cài lại tệp Caddy cũ từ repo (`git show HEAD~1:deploy/kiemtien-staging.caddy`).

## VPS thì sao?

**Không có VPS thật.** `docs/DEPLOY.md` là bản hướng dẫn mẫu — IP trong đó (`203.0.113.10`)
thuộc dải dành riêng cho tài liệu, tên miền là `ten-mien-cua-ban.com`. Tệp đó đã đổi sang
`bladeascend` ở phần chữ, nhưng **không có máy nào cần chạy lệnh**. Khi nào dựng VPS thật
thì cứ làm theo tài liệu đó, tên đã đúng sẵn.
