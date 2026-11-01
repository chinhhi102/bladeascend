#!/usr/bin/env bash
# tools/deploy.sh — đưa bản mới lên server LAN :9999 một cách AN TOÀN.
#
#   tools/deploy.sh                      # kiểm tra, build, deploy, khởi động lại
#   tools/deploy.sh --no-restart
#   tools/deploy.sh rollback             # liệt kê các bản lùi được
#   tools/deploy.sh rollback <build-id>  # quay về đúng bản đó, < 1 phút
#
# ── Vì sao có file này ──────────────────────────────────────────────────────────────────
# Ngày 30/10 chủ dự án đang chơi thì game đứng. Nguyên nhân không phải lỗi game: server
# phục vụ THẲNG từ `client/dist` và `assets/`, mà tôi thì build lại hai thư mục đó liên tục
# trong lúc làm việc. Trình duyệt tải `index.html` trỏ tới `index-ABC.js`, vài giây sau tôi
# build lại và tệp đó biến mất — client treo giữa chừng.
#
# Cùng lúc đó còn một lỗi im lặng hơn: `npm run check:telemetry-off` build bản RELEASE (tắt
# telemetry) ngay đè lên `client/dist`. Server vẫn chạy, vẫn có người chơi, nhưng không thu
# được một mẫu nào — và không có gì báo.
#
# Cách sửa: server phục vụ từ `server/data/deploy/`, một thư mục CHỈ script này ghi vào, và
# ghi theo kiểu đổi tên nguyên khối. Build trong lúc làm việc không còn chạm tới người chơi.
set -euo pipefail

cd "$(dirname "$0")/.."
REPO="$PWD"
DEPLOY="$REPO/server/data/deploy"
STAGE="$DEPLOY.new"
# ARCHIVE giữ các bản đã deploy để lùi lại được.
#
# Chép bằng LIÊN KẾT CỨNG (`cp -al`): hai bản liền nhau khác nhau vài tệp, phần còn lại
# dùng chung inode. 5 bản gần nhất tốn thêm vài MB chứ không phải 50 MB. Đây cũng là lý do
# lùi được trong vài giây: không chép gì, chỉ đổi tên thư mục.
ARCHIVE="$REPO/server/data/deploy-archive"
KEEP_BUILDS=8

GOROOT="${GOROOT:-$HOME/.cache/kx-gomod/golang.org/toolchain@v0.0.1-go1.26.0.linux-amd64}"
if [ -x "$GOROOT/bin/go" ]; then
  export GOROOT GOMODCACHE="${GOMODCACHE:-$HOME/.cache/kx-gomod}"
  export PATH="$GOROOT/bin:$PATH"
fi

# ── Hàm dùng chung cho deploy và rollback ───────────────────────────────────────────────
#
# Tách ra thành hàm vì `rollback` phải chạy ĐÚNG những bước này. Viết lại một bản rút gọn
# cho rollback là cách chắc chắn nhất để một ngày nào đó lùi về bản cũ rồi phát hiện server
# không lên, đúng lúc đang cần lùi gấp.

restart_server() {
  local build="$1"
  # Tắt cả tên MỚI lẫn tên CŨ.
  #
  # Binary đổi tên `kxserver` → `bladeascend` ngày 1/11. Bản đang chạy lúc đổi vẫn mang
  # tên cũ, nên nếu chỉ tìm tên mới thì deploy không tắt được nó, rồi bản mới kẹt cổng
  # 9999 và cả lượt deploy hỏng — với một dòng lỗi "address already in use" mà người đọc
  # không nối được với việc đổi tên. Giữ dòng `kxserver` tới khi chắc không còn máy nào
  # chạy bản trước 1/11; xoá sớm thì chính lúc cần nó là lúc không có.
  pkill -f 'bladeascend' 2>/dev/null || true
  pkill -f 'kxserver' 2>/dev/null || true
  sleep 2
  # Mật khẩu quản trị đọc từ `docs/private/admin.env` (thư mục riêng của chủ dự án, đã
  # nằm trong .gitignore từ 22/10).
  #
  # Trước đây chỗ này truyền `KX_ADMIN_PASS`, một tên biến **server không bao giờ đọc**:
  # `admin.NewAuth` chỉ nhìn `ADMIN_PASSWORD_HASH`. Hậu quả là trang quản trị TẮT trên
  # `:9999` suốt từ lúc làm A1 — mọi thứ A1–A4 dựng lên đều không vào được, và dòng log
  # "trang quản trị: tắt" trôi qua mà không ai đọc.
  local adminenv="$REPO/docs/private/admin.env"
  [ -f "$adminenv" ] || adminenv="$REPO/server/data/admin.env"
  local adminhash=""
  if [ -f "$adminenv" ]; then
    # ĐỌC bằng `cut`, KHÔNG `source`.
    #
    # Hash bcrypt bắt đầu bằng `$2a$12$...`. Nạp tệp bằng `.` thì bash **khai triển** `$2`
    # thành tham số vị trí thứ hai — và với `set -u` nó còn dừng hẳn script:
    # "line 1: $2: unbound variable". Đã dính đúng vậy lúc 21:25.
    adminhash=$(grep -m1 '^ADMIN_PASSWORD_HASH=' "$adminenv" | cut -d= -f2-)
  fi
  ( cd "$REPO/server" && setsid env ADMIN_PASSWORD_HASH="$adminhash" ./bladeascend \
      -build "$build" \
      -client "$DEPLOY/client" \
      -assets "$DEPLOY/assets" \
      > data/bladeascend.log 2>&1 < /dev/null & ) >/dev/null 2>&1 </dev/null
  # Ba lần chuyển hướng trên subshell là CẦN, không thừa.
  #
  # Tiến trình server chạy nền kế thừa fd của subshell. Khi gọi deploy qua ống
  # (`tools/deploy.sh | grep ...`) thì fd đó LÀ cái ống, và server giữ đầu ghi của ống mở
  # suốt đời nó — `grep` ngồi chờ EOF không bao giờ tới, nên lệnh deploy "treo" dù đã xong
  # từ lâu. Chuyển hướng ở chính `./bladeascend` là chưa đủ: subshell vẫn cầm bản sao.
  sleep 4
  if ! grep -q "trang quản trị: BẬT" "$REPO/server/data/bladeascend.log" 2>/dev/null; then
    echo "   ⚠ trang quản trị TẮT — chưa có docs/private/admin.env." >&2
    echo "     Tạo bằng:  cd server && go run ./cmd/adminpass > ../docs/private/admin.env" >&2
  fi
  if ! ss -ltn 2>/dev/null | grep -q ':9999'; then
    echo "LỖI: server không lên. Xem server/data/bladeascend.log" >&2
    tail -5 "$REPO/server/data/bladeascend.log" >&2
    return 1
  fi
}

# self_check hỏi một câu duy nhất: NGƯỜI CHƠI THẬT có vào được không.
#
# "Server đang lắng nghe" không trả lời được câu đó. Lần deploy 30/10 server lên bình
# thường, trả HTTP 200 cho mọi trang, mà game vẫn đứng ở màn hình chờ vì thiếu đúng hai tệp
# font — và chủ dự án phải vào lại 8 lần mới báo.
self_check() {
  local fail=0
  # Địa chỉ font lấy TỪ MANIFEST, không gõ tay: tên tệp trong `assets/dist` mang hash nội
  # dung nên đổi mỗi lần art đổi. Gõ tay là bài tự kiểm sẽ báo đỏ đúng cái nó không nên
  # báo, rồi người ta sẽ bỏ qua nó — và lần thiếu font thật thì cũng bỏ qua luôn.
  local font_urls
  font_urls=$(python3 - "$DEPLOY/assets/manifest.json" <<'PY'
import json, sys
m = json.load(open(sys.argv[1]))
for rel in ('fonts/BeVietnamPro-700.ttf', 'fonts/BeVietnamPro-800.ttf'):
    f = m['files'].get(rel)
    if f:
        print('/assets/' + f['url'])
PY
)
  local u code
  for u in / /api/config /assets/manifest.json $font_urls; do
    code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 5 "http://127.0.0.1:9999$u" || echo 000)
    if [ "$code" != "200" ]; then
      echo "   LỖI $u → HTTP $code" >&2
      fail=1
    fi
  done
  # Tệp bundle mà index.html trỏ tới phải tồn tại — đây là chỗ hỏng khi build đè lên thư
  # mục đang phục vụ.
  local js
  for js in $(grep -o 'static/[A-Za-z0-9._-]*\.js' "$DEPLOY/client/index.html"); do
    code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 5 "http://127.0.0.1:9999/$js" || echo 000)
    [ "$code" = "200" ] || { echo "   LỖI /$js → HTTP $code" >&2; fail=1; }
  done
  return "$fail"
}

list_builds() {
  if [ ! -d "$ARCHIVE" ] || [ -z "$(ls -A "$ARCHIVE" 2>/dev/null)" ]; then
    echo 'Kho rỗng — chưa có bản nào lùi được. Chạy tools/deploy.sh một lần là có.'
    return 0
  fi
  local cur=""
  [ -f "$DEPLOY/BUILD_ID" ] && cur=$(cat "$DEPLOY/BUILD_ID")
  echo "Các bản lùi được (mới nhất trước):"
  local b
  for b in $(ls -1 "$ARCHIVE" | sort -r); do
    if [ "$b" = "$cur" ]; then
      echo "  $b   ← đang chạy"
    else
      echo "  $b"
    fi
  done
  echo
  echo "Lùi về:  tools/deploy.sh rollback <build-id>"
}

do_rollback() {
  local want="$1"
  local src="$ARCHIVE/$want"
  if [ ! -d "$src" ]; then
    echo "LỖI: không có bản \"$want\" trong kho." >&2
    echo >&2
    list_builds >&2
    exit 1
  fi
  echo "== LÙI VỀ $want =="
  # KHÔNG build, KHÔNG kiểm tra lại: bản này đã qua đủ cổng kiểm lúc deploy lần đầu, và
  # lúc cần lùi thì thứ quý nhất là thời gian. Lùi phải xong trong vài giây.
  local t0=$SECONDS
  rm -rf "$DEPLOY.rollback"
  cp -al "$src" "$DEPLOY.rollback"
  rm -rf "$DEPLOY.prev"
  [ -d "$DEPLOY" ] && mv "$DEPLOY" "$DEPLOY.prev"
  mv "$DEPLOY.rollback" "$DEPLOY"
  rm -rf "$DEPLOY.prev"
  echo "   đổi thư mục xong ($((SECONDS - t0))s)"
  restart_server "$want"
  echo "== tự kiểm bản vừa lùi =="
  if ! self_check; then
    echo "LỖI: bản lùi về KHÔNG chơi được. Server vẫn chạy nó — xem log." >&2
    exit 1
  fi
  echo "   server đang chạy build $want, mọi tệp cần thiết đều tải được ($((SECONDS - t0))s tổng)"
}

# ── Điều hướng lệnh ─────────────────────────────────────────────────────────────────────
if [ "${1:-}" = "rollback" ]; then
  if [ -n "${2:-}" ]; then
    do_rollback "$2"
  else
    list_builds
  fi
  exit 0
fi

BUILD_ID="$(date +%Y%m%d_%H%M)_$(git rev-parse --short HEAD)"
echo "== build id: $BUILD_ID =="

echo "== 1/5 kiểm tra =="
# Kiểm TRƯỚC khi build: deploy một bản thiếu art hoặc hỏng test thì người chơi lãnh đủ.
( cd server && go vet ./... && go test ./... >/dev/null )
# check:assets và check:dead là HAI CHIỀU của cùng một câu hỏi: code đòi tệp không có
# (người chơi thấy placeholder), và tệp có mà không ai đòi (người chơi tải về rồi vứt).
( cd client && npm run check:assets >/dev/null && npm run check:items >/dev/null && npm run check:dead >/dev/null )
echo "   go vet, go test, check:assets, check:items, check:dead — OK"

echo "== 2/5 build client (TELEMETRY=on) =="
# TELEMETRY=on BẮT BUỘC cho server đội test. Thiếu cờ này thì server vẫn chạy ngon lành mà
# không thu được gì — và chỉ phát hiện ra khi đi tìm số liệu rồi thấy bảng trống.
( cd client && TELEMETRY=on BUILD_ID="$BUILD_ID" npm run build >/dev/null )
if ! grep -ql sendBeacon client/dist/static/*.js; then
  echo "LỖI: bản vừa build KHÔNG có telemetry. Dừng deploy." >&2
  exit 1
fi
echo "   có telemetry ✓"

echo "== 3/5 build server =="
( cd server && go build -o bladeascend ./cmd/server )

echo "== 4/5 dựng thư mục deploy =="
rm -rf "$STAGE"
mkdir -p "$STAGE/client" "$STAGE/assets"
cp -r client/dist/. "$STAGE/client/"
# Trang quản trị nằm NGOÀI `client/dist` (ADMIN_SPEC §1: mã admin không đi chung bundle
# công khai của game), nên phải chép riêng. Server tìm nó ở `<client>/admin`.
mkdir -p "$STAGE/client/admin"
cp -r client/admin/. "$STAGE/client/admin/"
# Chép NGUYÊN `assets/dist` — sản phẩm của `tools/asset-pipeline` (GAME_SPEC §13.7 mục 2).
# `npm run build` ở bước 2/5 đã dựng nó và đã qua cổng `assets:budget`.
#
# Trước đây chỗ này chép `assets/` rồi loại trừ ba thư mục ảnh nguồn, và đó là một danh sách
# loại trừ phải tự nhớ nới ra mỗi lần thêm thư mục nguồn mới. Giờ không cần: `assets/dist`
# chỉ chứa thứ phát hành, vì đó là định nghĩa của nó.
#
# Vẫn KHÔNG chép theo `manifest.json`. Lần đầu tôi chép theo manifest và làm server hỏng
# hẳn: manifest chỉ liệt kê .png/.json/âm thanh, KHÔNG có font `.ttf`. Thiếu font thì
# `loadFonts()` chặn ngay trước khi game khởi động — người chơi thấy màn hình đứng mãi ở
# "đang kết nối". Chủ dự án vào lại 8 lần trong 5 phút mới báo. (Manifest v2 đã có font,
# nhưng luật "chép cả thư mục" vẫn an toàn hơn và không mất gì.)
if [ ! -f "$REPO/assets/dist/manifest.json" ]; then
  echo "LỖI: chưa có assets/dist/manifest.json. Chạy: node tools/asset-pipeline/index.mjs" >&2
  exit 1
fi
rsync -a --delete --exclude '.pipeline-cache.json' "$REPO/assets/dist/" "$STAGE/assets/"
echo "   chép $(find "$STAGE/assets" -type f | wc -l) tệp asset"
echo "$BUILD_ID" > "$STAGE/BUILD_ID"
# Đổi tên NGUYÊN KHỐI: không có khoảnh khắc nào thư mục ở trạng thái nửa vời.
rm -rf "$DEPLOY.old"
[ -d "$DEPLOY" ] && mv "$DEPLOY" "$DEPLOY.old"
mv "$STAGE" "$DEPLOY"
rm -rf "$DEPLOY.old"

if [ "${1:-}" = "--no-restart" ]; then
  echo "== xong (không khởi động lại) =="
  exit 0
fi

echo "== 5/5 khởi động lại :9999 =="
restart_server "$BUILD_ID"

echo "== 6/6 tự kiểm bản vừa deploy =="
if ! self_check; then
  echo "LỖI: bản vừa deploy KHÔNG chơi được. Server vẫn chạy bản này — sửa rồi deploy lại." >&2
  echo "Lùi nhanh về bản trước:  tools/deploy.sh rollback" >&2
  exit 1
fi
echo "   server đang chạy build $BUILD_ID, mọi tệp cần thiết đều tải được"

# Cất vào kho SAU KHI đã tự kiểm xong: kho chỉ nên chứa bản đã chứng minh là chơi được.
# Cất một bản hỏng vào kho là đặt sẵn một cái bẫy cho lần lùi gấp sau này.
mkdir -p "$ARCHIVE"
rm -rf "$ARCHIVE/$BUILD_ID"
cp -al "$DEPLOY" "$ARCHIVE/$BUILD_ID"
# Giữ KEEP_BUILDS bản gần nhất.
ls -1 "$ARCHIVE" | sort -r | tail -n "+$((KEEP_BUILDS + 1))" | while read -r old; do
  rm -rf "${ARCHIVE:?}/$old"
done
echo "   đã cất vào kho ($(ls -1 "$ARCHIVE" | wc -l) bản lùi được — tools/deploy.sh rollback)"
