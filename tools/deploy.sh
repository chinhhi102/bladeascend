#!/usr/bin/env bash
# tools/deploy.sh — đưa bản mới lên server LAN :9999 một cách AN TOÀN.
#
#   tools/deploy.sh            # kiểm tra, build, deploy, khởi động lại
#   tools/deploy.sh --no-restart
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

GOROOT="${GOROOT:-$HOME/.cache/kx-gomod/golang.org/toolchain@v0.0.1-go1.26.0.linux-amd64}"
if [ -x "$GOROOT/bin/go" ]; then
  export GOROOT GOMODCACHE="${GOMODCACHE:-$HOME/.cache/kx-gomod}"
  export PATH="$GOROOT/bin:$PATH"
fi

BUILD_ID="$(date +%Y%m%d_%H%M)_$(git rev-parse --short HEAD)"
echo "== build id: $BUILD_ID =="

echo "== 1/5 kiểm tra =="
# Kiểm TRƯỚC khi build: deploy một bản thiếu art hoặc hỏng test thì người chơi lãnh đủ.
( cd server && go vet ./... && go test ./... >/dev/null )
( cd client && npm run check:assets >/dev/null && npm run check:items >/dev/null )
echo "   go vet, go test, check:assets, check:items — OK"

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
( cd server && go build -o kxserver ./cmd/server )

echo "== 4/5 dựng thư mục deploy =="
rm -rf "$STAGE"
mkdir -p "$STAGE/client" "$STAGE/assets"
cp -r client/dist/. "$STAGE/client/"
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
pkill -f 'kxserver' 2>/dev/null || true
sleep 2
cd server
# `-client` và `-assets` trỏ vào thư mục deploy, KHÔNG phải thư mục làm việc.
setsid env KX_ADMIN_PASS="${KX_ADMIN_PASS:-kiemxoay}" ./kxserver \
  -build "$BUILD_ID" \
  -client "$DEPLOY/client" \
  -assets "$DEPLOY/assets" \
  > data/kxserver.log 2>&1 < /dev/null &
sleep 4
if ! ss -ltn 2>/dev/null | grep -q ':9999'; then
  echo "LỖI: server không lên. Xem server/data/kxserver.log" >&2
  tail -5 data/kxserver.log >&2
  exit 1
fi

echo "== 6/6 tự kiểm bản vừa deploy =="
# "Server đang lắng nghe" KHÔNG có nghĩa là chơi được. Lần deploy trước server lên bình
# thường, trả HTTP 200 cho mọi trang, mà game vẫn đứng ở màn hình chờ vì thiếu đúng hai tệp
# font — và chủ dự án phải vào lại 8 lần mới báo. Nên bước cuối phải hỏi: một người chơi
# thật có vào được không.
fail=0
# Địa chỉ font lấy TỪ MANIFEST, không gõ tay: tên tệp trong `assets/dist` mang hash nội
# dung nên đổi mỗi lần art đổi. Gõ tay là bài tự kiểm sẽ báo đỏ đúng cái nó không nên báo,
# rồi người ta sẽ bỏ qua nó — và lần thiếu font thật thì cũng bỏ qua luôn.
FONT_URLS=$(python3 - "$DEPLOY/assets/manifest.json" <<'PY'
import json, sys
m = json.load(open(sys.argv[1]))
for rel in ('fonts/BeVietnamPro-700.ttf', 'fonts/BeVietnamPro-800.ttf'):
    f = m['files'].get(rel)
    if f:
        print('/assets/' + f['url'])
PY
)
for u in / /api/config /assets/manifest.json $FONT_URLS; do
  code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 5 "http://127.0.0.1:9999$u" || echo 000)
  if [ "$code" != "200" ]; then
    echo "   LỖI $u → HTTP $code" >&2
    fail=1
  fi
done
# Tệp bundle mà index.html trỏ tới phải tồn tại — đây là chỗ hỏng khi build đè lên thư mục
# đang phục vụ.
for js in $(grep -o 'static/[A-Za-z0-9._-]*\.js' "$DEPLOY/client/index.html"); do
  code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 5 "http://127.0.0.1:9999/$js" || echo 000)
  [ "$code" = "200" ] || { echo "   LỖI /$js → HTTP $code" >&2; fail=1; }
done
if [ "$fail" != "0" ]; then
  echo "LỖI: bản vừa deploy KHÔNG chơi được. Server vẫn chạy bản này — sửa rồi deploy lại." >&2
  exit 1
fi
echo "   server đang chạy build $BUILD_ID, mọi tệp cần thiết đều tải được"
