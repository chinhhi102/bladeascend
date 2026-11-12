#!/usr/bin/env bash
# tools/leaknightly.sh — bài kiểm rò rỉ HẰNG ĐÊM (GAME_SPEC §13.9, v6.54).
#
#   tools/leaknightly.sh              # chạy đủ: server 60 phút game + client 30 phút thật
#   tools/leaknightly.sh --force      # chạy kể cả khi máy bận / có người online
#   tools/leaknightly.sh --install    # cài mục cron 03:00 (in ra dòng cron, idempotent)
#   tools/leaknightly.sh --uninstall  # gỡ mục cron
#   tools/leaknightly.sh --status     # in kết quả lần chạy gần nhất
#
# ── Vì sao TÁCH khỏi cổng deploy (v6.54) ────────────────────────────────────────────────
#
# Bài client chạy 30 phút THẬT. Để nó trong cổng deploy nghĩa là mỗi lần sửa một dòng chữ
# cũng phải chờ nửa tiếng — và cái giá đó không được trả bằng sự kiên nhẫn, nó được trả
# bằng `SKIP_LEAK=1`. Một cổng ai cũng bỏ qua là một cổng không tồn tại.
#
# Nên: chạy hằng đêm lúc 03:00, khi không có ai chơi và máy rảnh. Kết quả ghi thành một
# tệp JSON; trang quản trị đọc nó và hiện thẻ cảnh báo khi đỏ, còn `tools/deploy.sh` từ
# chối deploy tới khi có người xem lại và xác nhận. Tức là bài vẫn gác, chỉ không gác bằng
# cách bắt người ta chờ.
#
# ── Vì sao có cổng "máy rảnh" ───────────────────────────────────────────────────────────
#
# Luật 16 của CLAUDE.md. Bài này đo bộ nhớ chứ không đo thời gian nên ít nhạy với tải hơn
# bài FPS, nhưng máy quá tải vẫn làm tụt FPS trình duyệt xuống dưới ngưỡng tự hạ chất
# lượng — mà hạ chất lượng ĐỔI chính những con số đang đo (bớt filter, bớt texture).
set -uo pipefail

cd "$(dirname "$0")/.."
REPO="$PWD"
OUT="$REPO/server/data/leak-nightly.json"
LOGDIR="$REPO/server/data/leak-logs"
PORT="${LEAK_PORT:-9994}"
CLIENT_MINUTES="${LEAK_CLIENT_MINUTES:-30}"
SERVER_MINUTES="${LEAK_SERVER_MINUTES:-60}"
MAX_LOAD="${MAX_LOAD:-1.5}"
GAME_URL="${GAME_URL:-http://127.0.0.1:9999}"

GOROOT="${GOROOT:-$HOME/.cache/kx-gomod/golang.org/toolchain@v0.0.1-go1.26.0.linux-amd64}"
if [ -x "$GOROOT/bin/go" ]; then
  export GOROOT GOMODCACHE="${GOMODCACHE:-$HOME/.cache/kx-gomod}"
  export PATH="$GOROOT/bin:$PATH"
fi

CRON_TAG="# bladeascend-leaknightly"
CRON_LINE="0 3 * * * cd $REPO && tools/leaknightly.sh >> $LOGDIR/cron.log 2>&1 $CRON_TAG"

case "${1:-}" in
  --install)
    mkdir -p "$LOGDIR"
    # Giữ idempotent: xoá dòng cũ của chính mình rồi thêm lại, không nhân bản sau mỗi lần cài.
    ( crontab -l 2>/dev/null | grep -v "$CRON_TAG"; echo "$CRON_LINE" ) | crontab -
    echo "đã cài mục cron:"
    echo "  $CRON_LINE"
    exit 0
    ;;
  --uninstall)
    crontab -l 2>/dev/null | grep -v "$CRON_TAG" | crontab -
    echo "đã gỡ mục cron bladeascend-leaknightly"
    exit 0
    ;;
  --status)
    if [ -f "$OUT" ]; then cat "$OUT"; else echo '{"note":"chưa chạy lần nào"}'; fi
    exit 0
    ;;
esac

FORCE=0
[ "${1:-}" = "--force" ] && FORCE=1

mkdir -p "$LOGDIR"
NOW_TS="$(date +%s)"
STAMP="$(date +%Y%m%d_%H%M)"

# write_result ghi tệp JSON mà trang quản trị và deploy.sh cùng đọc.
#
# Ghi qua tệp tạm rồi `mv`: deploy.sh có thể đọc đúng lúc bài này đang ghi, và đọc một
# tệp JSON mới viết được một nửa thì nó sẽ coi là "không có kết quả" — im lặng mất cổng.
write_result() {
  local ok="$1" skipped="$2" note="$3"
  cat > "$OUT.tmp" <<JSON
{
  "at": $NOW_TS,
  "ok": $ok,
  "skipped": "$skipped",
  "note": "$note",
  "client_minutes": $CLIENT_MINUTES,
  "server_minutes": $SERVER_MINUTES,
  "log": "leak-logs/$STAMP"
}
JSON
  mv "$OUT.tmp" "$OUT"
}

# ── Cổng 1: không có ai đang chơi ───────────────────────────────────────────────────────
humans="$(curl -s --max-time 5 "$GAME_URL/api/rooms" 2>/dev/null \
  | grep -o '"humans":[0-9]*' | cut -d: -f2 | awk '{s+=$1} END {print s+0}')"
humans="${humans:-0}"
if [ "$FORCE" = "0" ] && [ "$humans" -gt 0 ]; then
  echo "bỏ qua: đang có $humans người chơi trên $GAME_URL"
  write_result true "có $humans người đang chơi" "bỏ qua, chạy lại đêm sau"
  exit 0
fi

# ── Cổng 2: máy rảnh (luật 16) ──────────────────────────────────────────────────────────
load1="$(cut -d' ' -f1 /proc/loadavg)"
ncpu="$(nproc)"
per="$(awk -v l="$load1" -v n="$ncpu" 'BEGIN{printf "%.2f", l/n}')"
echo "tải máy: $load1 trên $ncpu nhân = $per/nhân"
if [ "$FORCE" = "0" ] && awk -v p="$per" -v m="$MAX_LOAD" 'BEGIN{exit !(p>m)}'; then
  echo "bỏ qua: máy đang bận ($per/nhân > $MAX_LOAD) — luật 16"
  write_result true "máy bận $per/nhân" "bỏ qua, chạy lại đêm sau"
  exit 0
fi

fail=0
notes=""

# ── Bài 1: rò rỉ SERVER (đồng hồ giả, nhanh, không phụ thuộc tải máy) ───────────────────
echo "== rò rỉ SERVER ($SERVER_MINUTES phút game) =="
( cd server && go run ./cmd/leakcheck -minutes "$SERVER_MINUTES" ) > "$LOGDIR/$STAMP-server.log" 2>&1
srv_rc=$?
tail -6 "$LOGDIR/$STAMP-server.log"
if [ "$srv_rc" != "0" ]; then
  fail=1
  notes="server rò rỉ"
fi

# ── Bài 2: rò rỉ CLIENT (30 phút thật, cần Playwright) ──────────────────────────────────
if [ -d tools/browsertest/node_modules/playwright ]; then
  echo "== rò rỉ CLIENT ($CLIENT_MINUTES phút thật) =="
  ( cd server && APP_ENV=dev go run ./cmd/server -addr "127.0.0.1:$PORT" -debug=true \
      > "$LOGDIR/$STAMP-devserver.log" 2>&1 & echo $! > /tmp/kx-leaknightly.pid )
  # Chờ server lên hẳn rồi mới mở trình duyệt: `go run` phải biên dịch trước, và trên máy
  # bận việc đó có khi mất gần một phút.
  for _ in $(seq 1 60); do
    curl -s -o /dev/null --max-time 2 "http://127.0.0.1:$PORT/api/config" && break
    sleep 2
  done
  MINUTES="$CLIENT_MINUTES" node tools/browsertest/leakcheck.mjs "http://127.0.0.1:$PORT" \
    > "$LOGDIR/$STAMP-client.log" 2>&1
  cli_rc=$?
  tail -14 "$LOGDIR/$STAMP-client.log"
  kill "$(cat /tmp/kx-leaknightly.pid 2>/dev/null)" 2>/dev/null || true
  pkill -x -f "exe/server" 2>/dev/null || true
  if [ "$cli_rc" != "0" ]; then
    fail=1
    notes="${notes:+$notes · }client rò rỉ"
  fi
else
  notes="${notes:+$notes · }bỏ bài client (chưa cài Playwright)"
fi

# Giữ 30 đêm nhật ký, đủ để so một tháng mà không phình đĩa.
ls -1t "$LOGDIR"/*-server.log 2>/dev/null | tail -n +31 | xargs -r rm -f
ls -1t "$LOGDIR"/*-client.log 2>/dev/null | tail -n +31 | xargs -r rm -f
ls -1t "$LOGDIR"/*-devserver.log 2>/dev/null | tail -n +31 | xargs -r rm -f

if [ "$fail" = "1" ]; then
  write_result false "" "${notes:-có bài đỏ}"
  echo ""
  echo "ĐỎ: $notes — xem $LOGDIR/$STAMP-*.log"
  echo "Deploy kế tiếp sẽ bị chặn cho tới khi xác nhận (tools/deploy.sh in sẵn cách xác nhận)."
  exit 1
fi
write_result true "" "${notes:-đạt}"
echo ""
echo "ĐẠT: không bài nào rò rỉ."
