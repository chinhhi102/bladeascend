#!/usr/bin/env bash
# tools/leakrepeat.sh — chạy bài rò rỉ client NHIỀU LƯỢT trên cùng một commit rồi so dốc.
#
#   tools/leakrepeat.sh                # 3 lượt × 60 phút (mặc định của §13.9 v6.57)
#   tools/leakrepeat.sh 3 60           # số lượt, số phút mỗi lượt
#   WAIT_IDLE=0 tools/leakrepeat.sh    # chạy ngay, kể cả khi máy bận (số sẽ là số nhiễu)
#
# ── Việc này để làm gì ──────────────────────────────────────────────────────────────────
#
# Spec v6.57 đòi KỊCH BẢN CỐ ĐỊNH, và đòi chứng minh là nó cố định thật: "chạy 3 lượt liên
# tiếp trên cùng một commit và báo 3 dốc. Nếu 3 dốc còn chênh > 3 lần thì kịch bản chưa cố
# định thật". Một lượt đo không nói được gì về chuyện đó — chỉ có nhiều lượt GIỐNG HỆT
# nhau mới nói được.
#
# Cũng chính là công cụ cho câu hỏi "heap có tăng theo SỐ LƯỢT ĐỔI KHU không" (B38): spec
# đòi ≥ 3 lượt mỗi bên, và đây là cách chạy ba lượt một bên mà không phải ngồi canh.
#
# ── Vì sao chờ máy rảnh ─────────────────────────────────────────────────────────────────
#
# CLAUDE.md luật 16. Máy bận không chỉ làm số xấu đi — nó làm client tụt FPS xuống dưới
# ngưỡng tự hạ chất lượng, mà hạ chất lượng ĐỔI chính những thứ đang đo (bớt filter, bớt
# texture). Thà chờ vài tiếng rồi có ba con số so được với nhau.
set -uo pipefail

cd "$(dirname "$0")/.."
REPO="$PWD"
RUNS="${1:-3}"
MINUTES="${2:-60}"
PORT="${LEAK_PORT:-9996}"
SEED="${LEAK_SEED:-20261123}"
OUTDIR="${OUTDIR:-$REPO/server/data/leak-repeat}"
# STATE_DIR: nơi ghi `leak-spread.json` (sàn nhiễu cho cổng mức nền, v6.59).
STATE_DIR="${STATE_DIR:-$REPO/server/data}"
MAX_LOAD="${MAX_LOAD:-1.5}"
WAIT_IDLE="${WAIT_IDLE:-1}"
# Chờ tối đa 8 tiếng rồi bỏ cuộc: im lặng chờ mãi thì sáng hôm sau không có số mà cũng
# không có lời giải thích nào.
WAIT_MAX_SECS="${WAIT_MAX_SECS:-28800}"

GOROOT="${GOROOT:-$HOME/.cache/kx-gomod/golang.org/toolchain@v0.0.1-go1.26.0.linux-amd64}"
if [ -x "$GOROOT/bin/go" ]; then
  export GOROOT GOMODCACHE="${GOMODCACHE:-$HOME/.cache/kx-gomod}"
  export PATH="$GOROOT/bin:$PATH"
fi
NVM_LATEST="$(ls -d "$HOME/.nvm/versions/node"/v* 2>/dev/null | sort -V | tail -1)"
if [ -n "$NVM_LATEST" ] && [ -x "$NVM_LATEST/bin/node" ]; then
  export PATH="$NVM_LATEST/bin:$PATH"
fi

mkdir -p "$OUTDIR"
BUILD="$(git rev-parse --short HEAD 2>/dev/null || echo '?')"
STAMP="$(date +%Y%m%d_%H%M)"

load_per_cpu() { awk -v n="$(nproc)" '{printf "%.2f", $1/n}' /proc/loadavg; }

wait_idle() {
  [ "$WAIT_IDLE" = "0" ] && return 0
  local waited=0
  while awk -v p="$(load_per_cpu)" -v m="$MAX_LOAD" 'BEGIN{exit !(p>m)}'; do
    if [ "$waited" -ge "$WAIT_MAX_SECS" ]; then
      echo "bỏ cuộc: máy vẫn bận ($(load_per_cpu)/nhân) sau $((waited / 3600)) giờ chờ — luật 16"
      return 1
    fi
    echo "$(date +%H:%M) máy bận $(load_per_cpu)/nhân (> $MAX_LOAD) — chờ 5 phút rồi xem lại"
    sleep 300
    waited=$((waited + 300))
  done
  echo "$(date +%H:%M) máy rảnh $(load_per_cpu)/nhân — bắt đầu"
  return 0
}

# RETRY: một lượt bị máy bận giữa chừng được chạy lại tối đa ngần này lần. Không chạy lại
# vô hạn: nếu máy bận cả đêm thì cái cần báo là "không đo được", không phải một vòng lặp im
# lặng tới sáng.
RETRY="${RETRY:-2}"

for i in $(seq 1 "$RUNS"); do
  attempt=0
  while :; do
    # Chờ máy rảnh TRƯỚC TỪNG LƯỢT, không chỉ lúc bắt đầu: ba lượt là hơn ba tiếng, và
    # việc của dự án khác không hỏi giờ.
    wait_idle || exit 2
    echo ""
    echo "═══ lượt $i/$RUNS (lần thử $((attempt + 1))) · $MINUTES phút · build $BUILD · hạt giống $SEED ═══"
    # Server MỚI mỗi lượt: cùng hạt giống nhưng server đã chạy 1 giờ thì bot, viên và quái
    # đã ở một trạng thái khác hẳn lúc mới dựng. Muốn ba lượt so được với nhau thì cả hai
    # phía — client và thế giới — phải bắt đầu từ cùng một chỗ.
    ( cd server && APP_ENV=dev go run ./cmd/server -addr "127.0.0.1:$PORT" -debug=true -seed "$SEED" \
        > "$OUTDIR/$STAMP-run$i-server.log" 2>&1 & echo $! > /tmp/kx-leakrepeat.pid )
    for _ in $(seq 1 60); do
      curl -s -o /dev/null --max-time 2 "http://127.0.0.1:$PORT/api/config" && break
      sleep 2
    done
    # KHÔNG truyền STATE_DIR: mốc kỷ lục và sổ đêm là trí nhớ của cổng gác hằng đêm, ba
    # lượt nghiệm thu không được phép ghi vào đó.
    MINUTES="$MINUTES" BUILD="$BUILD" JSON_OUT="$OUTDIR/$STAMP-run$i.json" \
      node tools/browsertest/leakcheck.mjs "http://127.0.0.1:$PORT" \
      > "$OUTDIR/$STAMP-run$i-client.log" 2>&1
    rc=$?
    echo "  mã thoát $rc · nhật ký $OUTDIR/$STAMP-run$i-client.log"
    tail -3 "$OUTDIR/$STAMP-run$i-client.log"
    kill "$(cat /tmp/kx-leakrepeat.pid 2>/dev/null)" 2>/dev/null || true
    sleep 5
    echo "  tải máy sau lượt này: $(load_per_cpu)/nhân"
    # Mã 3 = KHÔNG KẾT LUẬN (máy bận giữa chừng). Giữ một lượt như thế trong bộ ba là tự
    # bỏ một chân ghế: ba dốc chỉ so được với nhau khi cả ba đo trong cùng điều kiện.
    if [ "$rc" = "3" ] && [ "$attempt" -lt "$RETRY" ]; then
      attempt=$((attempt + 1))
      echo "  → lượt $i không kết luận được, chạy lại lần $((attempt + 1))/$((RETRY + 1))"
      continue
    fi
    break
  done
done

echo ""
echo "═══ $RUNS dốc, cùng commit $BUILD ═══"
# `--write-spread`: ghi lại ĐỘ CHÊNH mức nền của bộ lượt này vào `leak-spread.json`.
# Đó là sàn nhiễu của cổng mốc kỷ lục — §13.9 v6.59 bắt ngưỡng vượt mốc phải ≥ 2× con số
# này, nên nó phải được đo lại mỗi khi có bộ ba mới, không phải gõ tay một lần rồi quên.
node tools/browsertest/leakcompare.mjs "$OUTDIR/$STAMP"-run*.json --write-spread "$STATE_DIR"
