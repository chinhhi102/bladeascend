#!/usr/bin/env bash
# Liệt kê và dừng tiến trình MỒ CÔI do bài đo / công cụ của dự án bỏ lại.
#
#   tools/cleanup.sh          # chỉ LIỆT KÊ, không dừng gì
#   tools/cleanup.sh --kill   # dừng thật
#
# ── Vì sao có tệp này ───────────────────────────────────────────────────────────────────
#
# Ngày 10/11 máy chạy load 31–46 trên 12 nhân và MỌI phép đo trình duyệt của lượt đó thành
# vô nghĩa: cùng một commit cho ra 0,66% rồi 18,8% khung nhảy cóc. Bài đo tự phá thứ nó
# đang đo. `browsertest/browser.mjs` chặn nguồn rò chính (trình duyệt không đóng khi bài đo
# ném lỗi); tệp này là lưới hứng cho những thứ lọt qua.
#
# ── KHÔNG đụng tới ──────────────────────────────────────────────────────────────────────
#
# Đây là máy dùng chung. Chỉ dừng thứ CHẮC CHẮN là của dự án và chắc chắn mồ côi:
#   - server đang phục vụ người chơi ở :9999  → không đụng
#   - Docker daemon, container, staging        → không đụng
#   - Chrome / VS Code / Claude của chủ máy    → không đụng
#   - tiến trình của phiên đang chạy lệnh này  → không đụng (tự loại theo cây cha-con)
set -uo pipefail
cd "$(dirname "$0")/.."

KILL=0
[ "${1:-}" = "--kill" ] && KILL=1

# Cổng của server THẬT — không bao giờ dừng.
KEEP_PORTS="9999"

found=0
report() { # <nhãn> <pid> <mô tả>
  printf '%-26s pid %-8s %s\n' "$1" "$2" "$3"
  found=$((found + 1))
}

mine=$$
is_mine() { # tiến trình nằm trong cây của chính lệnh này?
  local p=$1
  while [ "$p" -gt 1 ] 2>/dev/null; do
    [ "$p" = "$mine" ] && return 0
    p=$(ps -o ppid= -p "$p" 2>/dev/null | tr -d ' ')
    [ -z "$p" ] && return 1
  done
  return 1
}

kill_pid() { # <pid> <nhãn>
  if [ "$KILL" = "1" ]; then
    kill "$1" 2>/dev/null && echo "   → đã dừng $2" || echo "   → không dừng được $2 (có thể của người khác)"
  fi
}

echo "== tiến trình mồ côi của dự án =="

# 1. Trình duyệt Playwright (KHÔNG phải Chrome của chủ máy ở /opt/google).
while read -r pid etime cmd; do
  [ -z "${pid:-}" ] && continue
  is_mine "$pid" && continue
  report "trình duyệt bài đo" "$pid" "chạy ${etime}s"
  kill_pid "$pid" "chromium $pid"
done < <(ps -eo pid,etimes,args --no-headers 2>/dev/null | grep -F "ms-playwright" | grep -v grep | awk '{print $1, $2, $3}')

# 2. Server dev của bài đo: mọi cổng của dự án TRỪ cổng thật.
for port in 9998 9997 9996 9995; do
  case " $KEEP_PORTS " in *" $port "*) continue ;; esac
  for pid in $(ss -ltnp 2>/dev/null | grep ":$port " | grep -oP 'pid=\K[0-9]+' | sort -u); do
    is_mine "$pid" && continue
    report "server dev :$port" "$pid" "$(ps -o etimes= -p "$pid" 2>/dev/null | tr -d ' ')s"
    kill_pid "$pid" "server :$port"
  done
done

# 3. Bài giả lập / sinh art chạy quá lâu (bình thường vài phút là xong).
while read -r pid etime cmd; do
  [ -z "${pid:-}" ] && continue
  is_mine "$pid" && continue
  [ "${etime:-0}" -lt 1800 ] && continue
  report "công cụ chạy quá lâu" "$pid" "${etime}s · ${cmd:0:60}"
  kill_pid "$pid" "$cmd"
done < <(ps -eo pid,etimes,args --no-headers 2>/dev/null \
  | grep -E "cmd/sim|cmd/loadtest|artgen|browsertest/" | grep -v grep | awk '{print $1, $2, $3}')

if [ "$found" = "0" ]; then
  echo "   (không có)"
fi

load=$(cut -d' ' -f1 /proc/loadavg)
cores=$(nproc)
per=$(awk -v l="$load" -v c="$cores" 'BEGIN { printf "%.2f", l / c }')
echo ""
echo "tải máy: $load trên $cores nhân = $per/nhân"
awk -v p="$per" 'BEGIN { exit (p > 1.5) }' \
  && echo "→ máy rảnh, đo được" \
  || echo "→ MÁY BẬN: đừng kết luận từ phép đo lúc này (CLAUDE.md luật 16)"

[ "$KILL" = "1" ] || { [ "$found" = "0" ] || echo ""; [ "$found" = "0" ] || echo "(chạy lại với --kill để dừng thật)"; }
exit 0
