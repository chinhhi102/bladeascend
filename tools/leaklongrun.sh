#!/usr/bin/env bash
# tools/leaklongrun.sh — bài đo DÀI (2 giờ) cho câu "bão hoà hay rò rỉ" — Q57, §13.9 v6.59.
#
#   tools/leaklongrun.sh                  # chạy ngay (cần máy rảnh, 0 người chơi)
#   tools/leaklongrun.sh --tonight        # đặt một mục cron MỘT LẦN cho 02:00 đêm nay
#   tools/leaklongrun.sh --cancel         # bỏ mục cron một lần đó
#
# ── Vì sao tách khỏi `leaknightly.sh` ───────────────────────────────────────────────────
#
# Bài hằng đêm là CỔNG GÁC: 60 phút, chạy mọi đêm, kết quả vào chỗ chặn deploy. Bài này là
# một phép đo MỘT LẦN để trả lời đúng một câu hỏi (Q57), dài gấp đôi, và không được phép
# ghi vào mốc kỷ lục hay sổ đêm — nên nó không truyền `STATE_DIR`.
#
# ── Vì sao 02:00 chứ không 03:00 ────────────────────────────────────────────────────────
#
# Chủ dự án duyệt "chạy trong khung 03:00, không chiếm máy giờ làm việc". Nhưng 03:00 là giờ
# của chính bài hằng đêm, và 04:30 là giờ job dọn dữ liệu (ADMIN §13). Bài này dài 2 giờ 10,
# nên 02:00 → 04:10 là khoảng duy nhất không đè lên cái nào. Nó đặt một KHOÁ để bài hằng đêm
# 03:00 tự bỏ qua đêm đó và ghi rõ lý do — một cổng gác bị bỏ qua phải nói ra, không được
# im lặng.
set -uo pipefail

cd "$(dirname "$0")/.."
REPO="$PWD"
PORT="${LEAK_PORT:-9992}"
SEED="${LEAK_SEED:-20261123}"
MINUTES="${MINUTES:-130}"   # 10 phút khởi động + 120 phút cửa sổ
OUTDIR="$REPO/server/data/leak-repeat"
LOCK="$REPO/server/data/leak-longrun.lock"
MAX_LOAD="${MAX_LOAD:-1.5}"
GAME_URL="${GAME_URL:-http://127.0.0.1:9999}"
CRON_TAG="# bladeascend-leaklongrun-once"

GOROOT="${GOROOT:-$HOME/.cache/kx-gomod/golang.org/toolchain@v0.0.1-go1.26.0.linux-amd64}"
if [ -x "$GOROOT/bin/go" ]; then
  export GOROOT GOMODCACHE="${GOMODCACHE:-$HOME/.cache/kx-gomod}"
  export PATH="$GOROOT/bin:$PATH"
fi
NVM_LATEST="$(ls -d "$HOME/.nvm/versions/node"/v* 2>/dev/null | sort -V | tail -1)"
if [ -n "$NVM_LATEST" ] && [ -x "$NVM_LATEST/bin/node" ]; then
  export PATH="$NVM_LATEST/bin:$PATH"
fi

case "${1:-}" in
  --tonight)
    mkdir -p "$OUTDIR"
    LINE="0 2 * * * cd $REPO && tools/leaklongrun.sh >> $OUTDIR/longrun-cron.log 2>&1 $CRON_TAG"
    ( crontab -l 2>/dev/null | grep -v "$CRON_TAG"; echo "$LINE" ) | crontab -
    echo "đã đặt mục cron MỘT LẦN cho 02:00:"
    echo "  $LINE"
    echo "Mục này tự xoá khi bài ĐO THẬT bắt đầu — máy bận lúc 02:00 thì lần hẹn còn nguyên cho đêm sau."
    exit 0
    ;;
  --cancel)
    crontab -l 2>/dev/null | grep -v "$CRON_TAG" | crontab -
    echo "đã bỏ mục cron một lần của bài đo dài"
    exit 0
    ;;
esac

mkdir -p "$OUTDIR"
STAMP="$(date +%Y%m%d_%H%M)"
BUILD="$(git rev-parse --short HEAD 2>/dev/null || echo '?')"
JSON="$OUTDIR/$STAMP-longrun.json"
LOG="$OUTDIR/$STAMP-longrun-client.log"

cleanup() {
  rm -f "$LOCK"
  kill "$(cat /tmp/kx-leaklongrun.pid 2>/dev/null)" 2>/dev/null || true
}
trap cleanup EXIT

# ── Hai cổng như bài hằng đêm: không có ai chơi, và máy rảnh (luật 16) ──────────────────
humans="$(curl -s --max-time 5 "$GAME_URL/api/rooms" 2>/dev/null \
  | grep -o '"humans":[0-9]*' | cut -d: -f2 | awk '{s+=$1} END {print s+0}')"
if [ "${humans:-0}" -gt 0 ]; then
  echo "bỏ qua: đang có ${humans} người chơi trên $GAME_URL"
  exit 0
fi
per="$(awk -v n="$(nproc)" '{printf "%.2f", $1/n}' /proc/loadavg)"
if awk -v p="$per" -v m="$MAX_LOAD" 'BEGIN{exit !(p>m)}'; then
  echo "bỏ qua: máy đang bận ($per/nhân > $MAX_LOAD) — luật 16"
  exit 0
fi

# Tự xoá mục cron — CHỈ khi đã qua hết cổng và sắp đo thật.
#
# Bản đầu tôi xoá ngay lúc khởi động, trước khi kiểm cổng: máy bận lúc 02:00 là mục cron bị
# tiêu mà phép đo không chạy, và sáng ra chỉ có một dòng "bỏ qua" trong log. Một lần hẹn
# phải sống tới khi nó đo được.
crontab -l 2>/dev/null | grep -v "$CRON_TAG" | crontab - 2>/dev/null || true

# Khoá để bài hằng đêm 03:00 biết mà nhường.
echo "$$ $(date -Is) longrun ${MINUTES}m" > "$LOCK"
echo "$(date +%H:%M) bắt đầu bài đo dài: $MINUTES phút · build $BUILD · hạt giống $SEED · tải $per/nhân"

( cd server && APP_ENV=dev go run ./cmd/server -addr "127.0.0.1:$PORT" -debug=true -seed "$SEED" \
    > "$OUTDIR/$STAMP-longrun-server.log" 2>&1 & echo $! > /tmp/kx-leaklongrun.pid )
for _ in $(seq 1 60); do
  curl -s -o /dev/null --max-time 2 "http://127.0.0.1:$PORT/api/config" && break
  sleep 2
done

# KHÔNG truyền STATE_DIR: bài một lần không được ghi vào mốc kỷ lục hay sổ đêm.
MINUTES="$MINUTES" BUILD="$BUILD" JSON_OUT="$JSON" \
  node tools/browsertest/leakcheck.mjs "http://127.0.0.1:$PORT" > "$LOG" 2>&1
rc=$?
echo "bài đo xong, mã thoát $rc · nhật ký $LOG"
tail -20 "$LOG"

echo ""
echo "═══ PHÂN ĐỊNH BÃO HOÀ (luật đã ghi trước khi chạy, §13.9 v6.59) ═══"
node tools/browsertest/leaksaturation.mjs "$JSON"
echo "mã phân định: $?  (0 = bão hoà · 1 = còn tăng · 3 = không kết luận)"
