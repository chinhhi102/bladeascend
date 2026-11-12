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
  #
  # `-x` (đúng TÊN tiến trình), KHÔNG `-f` (cả dòng lệnh). Từ khi repo đổi tên thành
  # `~/git/bladeascend`, chữ "bladeascend" nằm trong đường dẫn của MỌI lệnh chạy từ repo:
  # `pkill -f` giết luôn chính shell gọi deploy (exit 144, không một dòng log), `npm run dev`,
  # editor... Tên tiến trình của server là đúng tên binary `./bladeascend`.
  pkill -x 'bladeascend' 2>/dev/null || true
  pkill -x 'kxserver' 2>/dev/null || true
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
  # `-pprof`: mở /debug/pprof CHO 127.0.0.1 (cmd/server/pprof.go). Đây là server LAN của
  # đội test; muốn lấy hồ sơ CPU 30 giây lúc tick đang xấu thì phải có sẵn, vì tắt server đi
  # để bật hồ sơ là mất đúng cái cảnh cần đo.
  ( cd "$REPO/server" && setsid env ADMIN_PASSWORD_HASH="$adminhash" ./bladeascend \
      -build "$build" \
      -pprof \
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
# Q48: nhìn người chạy nhanh từ tab khác có nhảy cóc không. Cần server DEV riêng (window.kx
# chỉ có khi server cấp quyền debug) nên không chạy chung với các `check:*` thuần Node.
# Ngưỡng: 0 nhảy cóc, 0 đứng hình (Q48 xong 11/11 — đồng hồ chỉnh tốc + trễ nội suy tự
# điều chỉnh). Ngưỡng tạm ≤ 3% cũ đã bỏ cùng với cách đo cũ.
# Đặt SKIP_SMOOTH=1 để bỏ qua (máy không có Playwright).
if [ "${SKIP_SMOOTH:-0}" != "1" ] && [ -d tools/browsertest/node_modules/playwright ]; then
  echo "== 2b/5 nhìn Shipper Gom Đơn từ tab khác (bản VỪA dựng) =="
  ( cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9997 -debug=true >/tmp/kx-smooth.log 2>&1 & echo $! > /tmp/kx-smooth.pid )
  sleep 22
  # Lấy mã thoát của BÀI ĐO, không phải của `tail`: `cmd | tail` trả mã của `tail` nên
  # cổng này lúc đầu luôn xanh dù bài đo đỏ.
  set +e
  node tools/browsertest/gomdon-smooth.mjs http://127.0.0.1:9997 >/tmp/kx-smooth.out 2>&1
  smooth_rc=$?
  set -e
  tail -3 /tmp/kx-smooth.out

  # §11.5/§17 (v6.47): bảng xếp hạng không được co giãn. Chạy trên CÙNG server dev đó —
  # dựng thêm một server nữa chỉ để đo một bảng DOM là tốn 22 giây vô ích.
  echo "== 2c/5 bảng xếp hạng đứng yên (§11.5) =="
  set +e
  SECS=45 node tools/browsertest/rbstable.mjs http://127.0.0.1:9997 >/tmp/kx-rbstable.out 2>&1
  rb_rc=$?
  set -e
  tail -6 /tmp/kx-rbstable.out

  # §17 dòng 1 + dòng 3: lớp chống sập và vào lại phiên. Hai bài này CHỈ chạy đúng vào
  # ngày có lỗi thật — tức là ngày không ai ngồi xem — nên chúng phải nằm ở cổng.
  echo "== 2c2/5 chống sập bằng lỗi giả (§13.8) =="
  set +e
  node tools/browsertest/crashtest.mjs http://127.0.0.1:9997 >/tmp/kx-crash.out 2>&1
  crash_rc=$?
  set -e
  tail -4 /tmp/kx-crash.out

  echo "== 2c3/5 vào lại phiên khi F5 (§7.1) =="
  set +e
  node tools/browsertest/rejoin.mjs http://127.0.0.1:9997 >/tmp/kx-rejoin.out 2>&1
  rejoin_rc=$?
  set -e
  tail -4 /tmp/kx-rejoin.out

  kill "$(cat /tmp/kx-smooth.pid)" 2>/dev/null || true
  pkill -f "127.0.0.1:9997" 2>/dev/null || true
  if [ "$crash_rc" != "0" ]; then
    echo "LỖI: lớp chống sập §13.8 không đạt. Dừng deploy. Xem /tmp/kx-crash.out" >&2
    exit 1
  fi
  if [ "$rejoin_rc" != "0" ]; then
    echo "LỖI: vào lại phiên §7.1 không đạt. Dừng deploy. Xem /tmp/kx-rejoin.out" >&2
    exit 1
  fi
  if [ "$rb_rc" = "1" ]; then
    echo "LỖI: bảng xếp hạng co giãn (§11.5). Dừng deploy." >&2
    exit 1
  elif [ "$rb_rc" != "0" ]; then
    echo "   (bỏ qua: không đo được — xem /tmp/kx-rbstable.out)"
  fi
  # 0 = đạt · 1 = vượt ngưỡng, chặn deploy · 2 = máy quá tải, số không dùng được → chỉ cảnh báo
  if [ "$smooth_rc" = "1" ]; then
    echo "LỖI: còn nhảy cóc / đứng hình khi nhìn người chạy nhanh. Dừng deploy." >&2
    exit 1
  elif [ "$smooth_rc" = "2" ]; then
    echo "   (bỏ qua: máy quá tải, số đo không dùng được)"
  fi
fi

# ── Cổng rò rỉ bộ nhớ (GAME_SPEC §13.9) ────────────────────────────────────────────────
#
# Hai bài, hai phía, cả hai đều CHẶN deploy khi đỏ:
#
#   - server: `cmd/leakcheck` chạy 60 phút GAME bằng đồng hồ giả, xong trong vài chục giây,
#     và không phụ thuộc tải máy (nó đo heap của chính tiến trình, không đo FPS).
#   - client: `tools/browsertest/leakcheck.mjs` chạy THẬT 30 phút trên Chrome headless —
#     spec chốt con số đó vì rò rỉ GPU chỉ lộ ra sau nhiều phút (181 lượt đổi khu → 1,4 GB).
#
# 30 phút cho mỗi lần deploy là một cái giá thật. Hạ bằng `LEAK_MINUTES=8 tools/deploy.sh`
# khi đang sửa gấp, hoặc bỏ hẳn bằng `SKIP_LEAK=1` — nhưng đừng để nó thành thói quen: bài
# này là thứ duy nhất bắt được rò rỉ trước khi người chơi gặp.
#
# Lưu ý: hạ xuống dưới ~24 phút thì bài client **không còn gác được gì** — nó in số rồi nói
# thẳng "KHÔNG KẾT LUẬN", vì bộ làm việc của game chỉ chạm trần quanh phút 20 và một cửa sổ
# nằm trước mốc đó sẽ đỏ với mọi build, kể cả build sạch (xem Q53).
if [ "${SKIP_LEAK:-0}" != "1" ]; then
  echo "== 2d/5 rò rỉ bộ nhớ SERVER (§13.9, 60 phút game, đồng hồ giả) =="
  set +e
  ( cd server && go run ./cmd/leakcheck -minutes "${LEAK_SERVER_MINUTES:-60}" ) >/tmp/kx-leak-server.out 2>&1
  leak_srv=$?
  set -e
  tail -6 /tmp/kx-leak-server.out
  if [ "$leak_srv" != "0" ]; then
    echo "LỖI: server rò rỉ bộ nhớ (§13.9). Dừng deploy. Xem /tmp/kx-leak-server.out" >&2
    exit 1
  fi

  if [ -d tools/browsertest/node_modules/playwright ]; then
    echo "== 2e/5 rò rỉ bộ nhớ CLIENT (§13.9, ${LEAK_MINUTES:-30} phút thật) =="
    ( cd server && APP_ENV=dev go run ./cmd/server -addr 127.0.0.1:9996 -debug=true >/tmp/kx-leak.log 2>&1 & echo $! > /tmp/kx-leak.pid )
    sleep 22
    set +e
    MINUTES="${LEAK_MINUTES:-30}" node tools/browsertest/leakcheck.mjs http://127.0.0.1:9996 >/tmp/kx-leak-client.out 2>&1
    leak_cli=$?
    set -e
    tail -8 /tmp/kx-leak-client.out
    kill "$(cat /tmp/kx-leak.pid)" 2>/dev/null || true
    pkill -f "127.0.0.1:9996" 2>/dev/null || true
    if [ "$leak_cli" != "0" ]; then
      echo "LỖI: client rò rỉ bộ nhớ (§13.9). Dừng deploy. Xem /tmp/kx-leak-client.out" >&2
      exit 1
    fi
  fi
fi


echo "== 3/5 build server =="
( cd server && go build -o bladeascend ./cmd/server )

echo "== 4/5 dựng thư mục deploy =="
rm -rf "$STAGE"
mkdir -p "$STAGE/client" "$STAGE/assets"
cp -r client/dist/. "$STAGE/client/"
# Trang quản trị (React + antd, ADMIN_SPEC §2.1) là bundle RIÊNG, `npm run build` dựng vào
# `client/dist/admin` nên đã nằm trong lệnh chép ở trên. Server tìm nó ở `<client>/admin` và
# chỉ phục vụ sau khi đăng nhập; `index.html` của game không trỏ tới nó.
if [ ! -f "$STAGE/client/admin/index.html" ]; then
  echo "LỖI: bản build thiếu trang quản trị (client/dist/admin). Dừng deploy." >&2
  exit 1
fi
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
