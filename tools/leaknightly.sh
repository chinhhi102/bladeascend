#!/usr/bin/env bash
# tools/leaknightly.sh — bài kiểm rò rỉ HẰNG ĐÊM (GAME_SPEC §13.9, v6.57).
#
#   tools/leaknightly.sh              # chạy đủ: server 60 phút game + client 60 phút thật
#   tools/leaknightly.sh --force      # chạy kể cả khi máy bận / có người online
#   tools/leaknightly.sh --install    # cài mục cron 03:00 (in ra dòng cron, idempotent)
#   tools/leaknightly.sh --uninstall  # gỡ mục cron
#   tools/leaknightly.sh --status     # in kết quả lần chạy gần nhất
#
# ── Vì sao TÁCH khỏi cổng deploy (v6.54) ────────────────────────────────────────────────
#
# Bài client chạy 60 phút THẬT (v6.57; trước là 30). Để nó trong cổng deploy nghĩa là mỗi
# lần sửa một dòng chữ cũng phải chờ một tiếng — và cái giá đó không được trả bằng sự kiên
# nhẫn, nó được trả bằng `SKIP_LEAK=1`. Một cổng ai cũng bỏ qua là một cổng không tồn tại.
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
# v6.57: bài client chạy 60 phút (phút 0–10 khởi động, cửa sổ đo 10–60). Cửa sổ 10 phút của
# v6.54 quá ngắn để kết luận — sáu lượt cùng build cho dốc chênh nhau hơn 10 lần (B40).
CLIENT_MINUTES="${LEAK_CLIENT_MINUTES:-60}"
SERVER_MINUTES="${LEAK_SERVER_MINUTES:-60}"
# Hạt giống cố định: cùng thế giới mỗi đêm. Khác hạt là khác vị trí viên, khác quái, khác
# bot — tức khác số đối tượng đang sống, mà đó chính là con số bài này đo (§13.9 v6.57).
LEAK_SEED="${LEAK_SEED:-20261123}"
# STATE_DIR: nơi giữ mốc kỷ lục và sổ đêm. Chỉ bài HẰNG ĐÊM được ghi vào đây.
STATE_DIR="${STATE_DIR:-$REPO/server/data}"
MAX_LOAD="${MAX_LOAD:-1.5}"
GAME_URL="${GAME_URL:-http://127.0.0.1:9999}"

GOROOT="${GOROOT:-$HOME/.cache/kx-gomod/golang.org/toolchain@v0.0.1-go1.26.0.linux-amd64}"
if [ -x "$GOROOT/bin/go" ]; then
  export GOROOT GOMODCACHE="${GOMODCACHE:-$HOME/.cache/kx-gomod}"
  export PATH="$GOROOT/bin:$PATH"
fi

# Node: cron chạy với PATH trần, tức là `node` của hệ thống.
#
# Đêm 13/11 bài rò rỉ báo ĐỎ với đúng một dòng: "You are running Node.js 18.19.1. Playwright
# requires Node.js 20 or higher." Không có rò rỉ nào cả — cron lấy Node 18 của hệ thống
# trong khi máy này chạy Node 24 qua nvm. Một cổng gác báo đỏ vì lý do môi trường thì sớm
# muộn cũng bị bỏ qua cho xong, và khi nó đỏ THẬT thì không ai tin nữa.
#
# Tìm bản Node mới nhất trong nvm; không có thì dùng `node` trên PATH và để Playwright tự
# báo lỗi (có log rõ còn hơn chạy sai âm thầm).
NVM_LATEST="$(ls -d "$HOME/.nvm/versions/node"/v* 2>/dev/null | sort -V | tail -1)"
if [ -n "$NVM_LATEST" ] && [ -x "$NVM_LATEST/bin/node" ]; then
  export PATH="$NVM_LATEST/bin:$PATH"
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
# `verdict` là chữ của §13.9 v6.57: xanh · vang (cảnh báo, không chặn) · do (chặn deploy) ·
# chua_ket_luan. Cờ `ok` giữ nguyên nghĩa cũ cho `tools/deploy.sh`: false mới chặn, nên
# VÀNG vẫn ok = true — spec nói rõ vàng không chặn.
write_result() {
  local ok="$1" skipped="$2" note="$3" verdict="${4:-}"
  cat > "$OUT.tmp" <<JSON
{
  "at": $NOW_TS,
  "ok": $ok,
  "verdict": "$verdict",
  "skipped": "$skipped",
  "note": "$note",
  "client_minutes": $CLIENT_MINUTES,
  "server_minutes": $SERVER_MINUTES,
  "log": "leak-logs/$STAMP"
}
JSON
  mv "$OUT.tmp" "$OUT"
}

# ── Cổng 0: có bài đo DÀI đang chạy thì nhường ──────────────────────────────────────────
#
# `tools/leaklongrun.sh` (bài 2 giờ của Q57) đặt khoá này. Hai bài cùng chạy thì cả hai đều
# là số nhiễu — nhưng một cổng gác bị bỏ qua phải NÓI RA, không được im lặng, nên đêm đó
# vẫn ghi kết quả kèm lý do.
LONGRUN_LOCK="$REPO/server/data/leak-longrun.lock"
if [ -f "$LONGRUN_LOCK" ]; then
  echo "bỏ qua: đang có bài đo dài chạy ($(cat "$LONGRUN_LOCK" 2>/dev/null))"
  write_result true "nhường bài đo dài" "bỏ qua, chạy lại đêm sau" "chua_ket_luan"
  exit 0
fi

# ── Cổng 1: không có ai đang chơi ───────────────────────────────────────────────────────
humans="$(curl -s --max-time 5 "$GAME_URL/api/rooms" 2>/dev/null \
  | grep -o '"humans":[0-9]*' | cut -d: -f2 | awk '{s+=$1} END {print s+0}')"
humans="${humans:-0}"
if [ "$FORCE" = "0" ] && [ "$humans" -gt 0 ]; then
  echo "bỏ qua: đang có $humans người chơi trên $GAME_URL"
  write_result true "có $humans người đang chơi" "bỏ qua, chạy lại đêm sau" "chua_ket_luan"
  exit 0
fi

# ── Cổng 2: máy rảnh (luật 16) ──────────────────────────────────────────────────────────
load1="$(cut -d' ' -f1 /proc/loadavg)"
ncpu="$(nproc)"
per="$(awk -v l="$load1" -v n="$ncpu" 'BEGIN{printf "%.2f", l/n}')"
echo "tải máy: $load1 trên $ncpu nhân = $per/nhân"
if [ "$FORCE" = "0" ] && awk -v p="$per" -v m="$MAX_LOAD" 'BEGIN{exit !(p>m)}'; then
  echo "bỏ qua: máy đang bận ($per/nhân > $MAX_LOAD) — luật 16"
  write_result true "máy bận $per/nhân" "bỏ qua, chạy lại đêm sau" "chua_ket_luan"
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
cli_verdict=""
if [ -d tools/browsertest/node_modules/playwright ]; then
  echo "== rò rỉ CLIENT ($CLIENT_MINUTES phút thật, hạt giống $LEAK_SEED) =="
  ( cd server && APP_ENV=dev go run ./cmd/server -addr "127.0.0.1:$PORT" -debug=true -seed "$LEAK_SEED" \
      > "$LOGDIR/$STAMP-devserver.log" 2>&1 & echo $! > /tmp/kx-leaknightly.pid )
  # Chờ server lên hẳn rồi mới mở trình duyệt: `go run` phải biên dịch trước, và trên máy
  # bận việc đó có khi mất gần một phút.
  for _ in $(seq 1 60); do
    curl -s -o /dev/null --max-time 2 "http://127.0.0.1:$PORT/api/config" && break
    sleep 2
  done
  # STATE_DIR: chỉ bài hằng đêm mới được ghi mốc kỷ lục và sổ đêm (§13.9 v6.57). BUILD đi
  # kèm để sau này đọc sổ còn biết mốc đó do bản nào đặt ra.
  MINUTES="$CLIENT_MINUTES" STATE_DIR="$STATE_DIR" BUILD="$(git rev-parse --short HEAD 2>/dev/null || echo '?')" \
    node tools/browsertest/leakcheck.mjs "http://127.0.0.1:$PORT" \
    > "$LOGDIR/$STAMP-client.log" 2>&1
  cli_rc=$?
  tail -24 "$LOGDIR/$STAMP-client.log"
  kill "$(cat /tmp/kx-leaknightly.pid 2>/dev/null)" 2>/dev/null || true
  pkill -x -f "exe/server" 2>/dev/null || true
  # Mã thoát của `leakcheck.mjs`: 0 = xanh hoặc VÀNG (vàng không chặn), 1 = ĐỎ,
  # 3 = KHÔNG KẾT LUẬN (bài quá ngắn). Trộn 3 vào "đỏ" là biến một lượt chạy hỏng thành
  # một cáo buộc rò rỉ — và cáo buộc sai thì lần sau không ai tin cổng này nữa.
  case "$cli_rc" in
    0) cli_verdict="$(grep -q '^VÀNG' "$LOGDIR/$STAMP-client.log" && echo vang || echo xanh)" ;;
    3) cli_verdict="chua_ket_luan"; notes="${notes:+$notes · }client KHÔNG KẾT LUẬN" ;;
    *) cli_verdict="do"; fail=1; notes="${notes:+$notes · }client rò rỉ" ;;
  esac
  [ "$cli_verdict" = "vang" ] && notes="${notes:+$notes · }client VÀNG (3 đêm dốc dương)"
else
  notes="${notes:+$notes · }bỏ bài client (chưa cài Playwright)"
fi

# Giữ 30 đêm nhật ký, đủ để so một tháng mà không phình đĩa.
ls -1t "$LOGDIR"/*-server.log 2>/dev/null | tail -n +31 | xargs -r rm -f
ls -1t "$LOGDIR"/*-client.log 2>/dev/null | tail -n +31 | xargs -r rm -f
ls -1t "$LOGDIR"/*-devserver.log 2>/dev/null | tail -n +31 | xargs -r rm -f

if [ "$fail" = "1" ]; then
  write_result false "" "${notes:-có bài đỏ}" "do"
  echo ""
  echo "ĐỎ: $notes — xem $LOGDIR/$STAMP-*.log"
  echo "Deploy kế tiếp sẽ bị chặn cho tới khi xác nhận (tools/deploy.sh in sẵn cách xác nhận)."
  exit 1
fi
write_result true "" "${notes:-đạt}" "${cli_verdict:-xanh}"
echo ""
if [ "$cli_verdict" = "vang" ]; then
  echo "VÀNG: dốc dương 3 đêm liên tiếp — mở việc điều tra, KHÔNG chặn deploy (§13.9 v6.57)."
else
  echo "ĐẠT: không bài nào rò rỉ."
fi
