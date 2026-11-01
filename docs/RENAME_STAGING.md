# Đổi tên hạ tầng staging sang `bladeascend` — lệnh từng bước

Chủ dự án chốt 1/11/2026. Phía repo **đã sửa xong và đã commit**; tệp này là phần phải chạy
trên MÁY CHỦ staging, vì đổi ở repo mà không đổi trên máy đó thì `deploy/staging.sh` sẽ
rsync vào một thư mục rỗng mới và để nguyên bản cũ đang chạy.

Mỗi lệnh có dòng **Mong đợi** — so từng chữ. Lệch thì **dừng lại**, đừng làm bước sau.

## Những cái tên phải đổi

| Thứ | Tên cũ | Tên mới |
|---|---|---|
| Thư mục trên máy chủ | `~/kiemxoaytutien` | `~/bladeascend` |
| Dự án compose | `kiemtien` | `bladeascend` |
| Container | `kiemtien-web` | `bladeascend-web` |
| **Volume dữ liệu** | **`kiemtien_kiemtien-data`** | `bladeascend-data` |
| Tệp cấu hình Caddy | `kiemtien-staging.caddy` | `bladeascend-staging.caddy` |
| Tên miền | `kiemtien.staging.maveragroups.com` | `bladeascend.staging.maveragroups.com` |

> **Volume là chỗ nguy hiểm duy nhất.** Docker KHÔNG đổi tên volume được; phải chép sang
> volume mới.
>
> **Bản trước của tài liệu này ghi sai tên volume** (`kiemtien-data`). Tên thật — đã kiểm
> trên máy chủ ngày 1/11 bằng `docker inspect kiemtien-web` — là `kiemtien_kiemtien-data`,
> vì compose tự thêm tiền tố tên dự án. Chạy lệnh chép với tên sai thì Docker **tạo một
> volume rỗng mang tên đó**, chép rỗng sang, và vẫn in `CHEP-XONG`. Đó là lý do các bước
> dưới đây đếm số dòng trước và sau, chứ không tin vào chữ "xong".
>
> `deploy/docker-compose.yml` nay đặt tên volume cố định `bladeascend-data` và khai
> `external: true`: nếu volume mới chưa có, `docker compose up` **báo lỗi và dừng** thay vì
> lặng lẽ tạo volume rỗng.
>
> Tên miền mới nằm trong wildcard `*.staging.maveragroups.com` nên **không phải sửa DNS**.

## Chuẩn bị

Mở terminal trên máy mình, nạp biến môi trường rồi vào máy chủ:

```bash
cd ~/git/bladeascend
set -a; . docs/private/deploy; set +a
ssh -i "$DEPLOY_KEY" "$DEPLOY_HOST"
```

Từ đây tới hết bước 6 là **chạy trên máy chủ staging**.

Dán hàm đếm này vào terminal máy chủ **một lần** (dùng lại ở bước 0, 3, 7). Nó đếm số dòng
bốn bảng trong một volume hoặc một tệp `.db`, trên **bản chép** của tệp để không bao giờ
khoá DB đang chạy:

```bash
dem() {  # dem <tên-volume>   hoặc   dem <đường-dẫn-tệp.db>
  if [ -f "$1" ]; then M="-v $(dirname "$1"):/d:ro"; F="/d/$(basename "$1")"
  else M="-v $1:/d:ro"; F=/d/leaderboard.db; fi
  docker run --rm $M alpine sh -c "apk add -q sqlite >/dev/null && cp $F /tmp/x.db &&
    echo \"kiem-tra: \$(sqlite3 /tmp/x.db 'PRAGMA integrity_check')\" &&
    for t in lives feedback bans admin_audit; do
      n=\$(sqlite3 /tmp/x.db \"SELECT count(*) FROM \$t\" 2>/dev/null) || n='(chưa có bảng)'
      echo \"\$t \$n\"; done"
}
```

Mong đợi: không in gì (chỉ định nghĩa hàm).

## Bước 0 — Sao lưu ra NGOÀI Docker

```bash
docker inspect kiemtien-web --format '{{range .Mounts}}{{.Name}} -> {{.Destination}}{{end}}'
```

Mong đợi: `kiemtien_kiemtien-data -> /data`. Ra tên khác thì **dùng đúng tên đó** thay cho
`kiemtien_kiemtien-data` ở mọi lệnh bên dưới.

```bash
mkdir -p ~/backup
B=~/backup/bladeascend-$(date +%Y%m%d).db
docker run --rm -v kiemtien_kiemtien-data:/d:ro -v ~/backup:/b alpine \
  cp /d/leaderboard.db /b/$(basename $B)
ls -l $B
```

Mong đợi: một dòng `-rw-r--r-- ... bladeascend-2026MMDD.db`, cỡ **khác 0** (ngày 1/11 là
61 440 byte).

Kiểm tệp sao lưu **mở được** và so với bản đang chạy:

```bash
dem $B
dem kiemtien_kiemtien-data
```

Mong đợi: hai khối **giống hệt nhau**, dòng đầu `kiem-tra: ok`. Bản staging cũ chỉ có bảng
`lives` (ngày 1/11: `lives 167`); ba bảng kia in `(chưa có bảng)` — đúng, vì bản cũ ra đời
trước trang quản trị. **Ghi lại số `lives`** — gọi là **N**.

Rồi chép bản sao lưu về máy mình cho chắc (chạy ở **terminal máy mình**, terminal thứ hai):

```bash
cd ~/git/bladeascend && set -a; . docs/private/deploy; set +a
mkdir -p ~/backup
scp -i "$DEPLOY_KEY" "$DEPLOY_HOST:backup/bladeascend-$(date +%Y%m%d).db" ~/backup/
ls -l ~/backup/bladeascend-*.db
```

Mong đợi: cùng cỡ byte với dòng `ls -l` trên máy chủ.

## Bước 1 — Xem hiện trạng (chưa đổi gì)

```bash
docker ps --format '{{.Names}} {{.Status}}' | grep -E 'kiemtien|bladeascend'
docker volume ls --format '{{.Name}}' | grep -E 'kiemtien|bladeascend'
ls -d ~/kiemxoaytutien ~/bladeascend
```

Mong đợi:
```
kiemtien-web Up ...
kiemtien_kiemtien-data
/home/ubuntu/kiemxoaytutien
ls: cannot access '/home/ubuntu/bladeascend': No such file or directory
```
Nếu đã thấy `bladeascend-web` hoặc volume `bladeascend-data` thì có ai đó deploy trước rồi —
**dừng lại, báo Claude**, đừng chép đè.

## Bước 2 — Dừng bản cũ

```bash
cp ~/kiemxoaytutien/docker-compose.yml ~/backup/docker-compose.kiemtien.yml
cd ~/kiemxoaytutien && docker compose down && cd ~
docker ps --format '{{.Names}}' | grep kiemtien || echo DA-DUNG
```

Mong đợi: dòng cuối `DA-DUNG`. Dừng chứ **chưa xoá volume** — `down` không đụng volume nếu
không có cờ `-v`. **Đừng bao giờ thêm `-v`.**

## Bước 3 — Chép dữ liệu sang volume mới

Đếm lại bản cũ **sau khi đã dừng** (có thể có người chơi giữa bước 0 và bước 2):

```bash
dem kiemtien_kiemtien-data
```

Mong đợi: `kiem-tra: ok`, `lives` ≥ N. Số này là **số chuẩn** để so ở dưới.

```bash
docker volume create bladeascend-data
docker run --rm -v bladeascend-data:/to alpine sh -c 'ls -A /to | wc -l'
```

Mong đợi: dòng `bladeascend-data`, rồi `0` (volume mới rỗng). Khác `0` → **dừng lại**.

```bash
docker run --rm -v kiemtien_kiemtien-data:/from:ro -v bladeascend-data:/to alpine \
  sh -c 'cd /from && cp -a . /to/ && ls -l /to'
dem bladeascend-data
```

Mong đợi: `ls` thấy `leaderboard.db` cùng cỡ byte với bước 0; khối `dem` **giống hệt**
khối đếm bản cũ ngay trên. Lệch một dòng nào → **dừng lại**, không làm tiếp.

## Bước 4 — Đổi tên thư mục

```bash
mv ~/kiemxoaytutien ~/bladeascend && ls -d ~/bladeascend
```

Mong đợi: `/home/ubuntu/bladeascend`.

## Bước 5 — Gỡ tệp Caddy cũ

```bash
sudo cp /opt/astery/caddy/conf.d/kiemtien-staging.caddy ~/backup/
sudo rm -f /opt/astery/caddy/conf.d/kiemtien-staging.caddy
docker exec docker-caddy-1 caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile \
  && docker exec docker-caddy-1 caddy reload --config /etc/caddy/Caddyfile --adapter caddyfile \
  && echo CADDY-OK
```

Mong đợi: dòng cuối `CADDY-OK` (trước đó có vài dòng log của Caddy, bình thường).
Làm bước này **trước** khi deploy: để lại tệp cũ thì Caddy giữ một site block trỏ tới
container `kiemtien-web` đã dừng.

## Bước 6 — Thoát, rồi deploy từ máy mình

```bash
exit
```

Trên **máy mình**:

```bash
cd ~/git/bladeascend
set -a; . docs/private/deploy; set +a
deploy/staging.sh
```

Mong đợi: dòng cuối `deployed: https://bladeascend.staging.maveragroups.com`.
Nếu thấy `external volume "bladeascend-data" not found` → bước 3 chưa chạy; quay lại bước 3.

## Bước 7 — Kiểm

Trên **máy mình**:

```bash
curl -sI https://bladeascend.staging.maveragroups.com | head -1
curl -s  https://bladeascend.staging.maveragroups.com/api/rooms | head -c 120; echo
```

Mong đợi: `HTTP/2 200`, rồi một mẩu JSON bắt đầu bằng `[` hoặc `{`.

Vào lại máy chủ (lệnh `ssh` ở phần Chuẩn bị, dán lại hàm `dem`) và đếm trên volume **đang
chạy**:

```bash
docker ps --format '{{.Names}} {{.Status}}' | grep bladeascend
docker inspect bladeascend-web --format '{{range .Mounts}}{{.Name}}{{end}}'
dem bladeascend-data
```

Mong đợi:
```
bladeascend-web Up ...
bladeascend-data
kiem-tra: ok
lives <số chuẩn bước 3, hoặc lớn hơn nếu đã có người chơi>
feedback 0
bans 0
admin_audit 0
```
Bản mới tự tạo ba bảng kia lúc khởi động, nên giờ chúng in `0` thay vì `(chưa có bảng)`.
`lives` **nhỏ hơn** số chuẩn → dữ liệu không lên → làm phần "Nếu hỏng giữa chừng".

Cuối cùng mở trang bằng trình duyệt, chơi thử một ván, mở bảng xếp hạng: phải thấy tên cũ.

## Bước 8 — Xoá volume cũ: **sau 7 ngày chạy ổn**, không làm hôm nay

Đây là bước riêng, làm vào một ngày khác (sớm nhất 7 ngày sau bước 7). Volume cũ chỉ vài
chục KB, và là đường lùi duy nhất nếu bước 3 chép thiếu.

```bash
dem kiemtien_kiemtien-data
dem bladeascend-data
```

Mong đợi: `lives` của volume mới ≥ volume cũ. Đúng thì mới chạy:

```bash
docker volume rm kiemtien_kiemtien-data
```

Mong đợi: in lại đúng tên `kiemtien_kiemtien-data`. Giữ tệp `~/backup/bladeascend-*.db`.

## Nếu hỏng giữa chừng

Mọi bước tới bước 7 đều lùi được, vì bản cũ chỉ bị *dừng*, volume cũ còn nguyên:

```bash
cd ~/bladeascend && docker compose down; cd ~
mv ~/bladeascend ~/kiemxoaytutien
cp ~/backup/docker-compose.kiemtien.yml ~/kiemxoaytutien/docker-compose.yml
sudo cp ~/backup/kiemtien-staging.caddy /opt/astery/caddy/conf.d/
sudo rm -f /opt/astery/caddy/conf.d/bladeascend-staging.caddy
docker exec docker-caddy-1 caddy reload --config /etc/caddy/Caddyfile --adapter caddyfile
cd ~/kiemxoaytutien && docker compose up -d
```

Mong đợi: `docker ps` thấy lại `kiemtien-web Up`, và `dem kiemtien_kiemtien-data` in đúng số
chuẩn bước 3. (Thư mục `~/kiemxoaytutien` lúc này chứa bản build mới do `rsync --delete`
chép đè, nên `up` sẽ build lại bằng mã mới nhưng tên cũ — dữ liệu vẫn là volume cũ.)

## VPS thì sao?

**Không có VPS thật.** `docs/DEPLOY.md` là bản hướng dẫn mẫu — IP trong đó (`203.0.113.10`)
thuộc dải dành riêng cho tài liệu, tên miền là `ten-mien-cua-ban.com`. Không có máy nào cần
chạy lệnh.
