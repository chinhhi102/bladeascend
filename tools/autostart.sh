#!/usr/bin/env bash
# TỰ CHẠY LẠI SERVER SAU KHI MÁY KHỞI ĐỘNG.
#
#   tools/autostart.sh --install     # cài mục @reboot vào crontab
#   tools/autostart.sh --status      # xem đã cài chưa
#   tools/autostart.sh --uninstall   # gỡ
#
# ── Vì sao có tệp này ───────────────────────────────────────────────────────────────────
#
# Ngày 17/11/2026 máy khởi động lại lúc 17:33. Server nhận SIGTERM của trình tắt máy và tắt
# êm đúng bài (§1.3: /readyz trả 503 → tiễn người chơi → checkpoint SQLite → xong sau 0,5 s).
# Rồi nó **nằm im 44 tiếng**.
#
# Không ai báo. Không cảnh báo nào kêu. Game chỉ đơn giản là không vào được, và chuyện đó
# lộ ra vì tình cờ có người đi tìm một lỗi khác. Phần tắt êm làm đúng tới từng bước — phần
# thiếu là không có gì dựng nó dậy.
#
# Dùng crontab `@reboot` chứ không systemd: máy này đã chạy `tools/leaknightly.sh` bằng
# crontab của người dùng, không cần sudo, và một cơ chế đã quen thì người sau còn đọc được.
set -euo pipefail

REPO="$(cd "$(dirname "$0")/.." && pwd)"
MARK="# bladeascend-autostart"
LOG="$REPO/server/data/autostart.log"
# `sleep 20`: chờ mạng và đĩa sẵn sàng. `@reboot` chạy rất sớm, và một server lên trước khi
# có mạng thì bind được cổng nhưng không ai tới được.
LINE="@reboot sleep 20 && cd $REPO && tools/deploy.sh start >> $LOG 2>&1 $MARK"

cur() { crontab -l 2>/dev/null || true; }

case "${1:---status}" in
  --install)
    if cur | grep -qF "$MARK"; then
      echo "đã cài sẵn:"
      cur | grep -F "$MARK"
      exit 0
    fi
    { cur; echo "$LINE"; } | crontab -
    echo "đã cài. Sau mỗi lần máy khởi động, server dựng lại bản trong server/data/deploy."
    echo "Nhật ký: $LOG"
    ;;
  --uninstall)
    cur | grep -vF "$MARK" | crontab -
    echo "đã gỡ mục @reboot."
    ;;
  --status)
    if cur | grep -qF "$MARK"; then
      echo "ĐÃ CÀI:"
      cur | grep -F "$MARK"
    else
      echo "CHƯA CÀI — máy khởi động lại là game offline tới khi có người để ý."
      echo "Cài bằng: tools/autostart.sh --install"
    fi
    ;;
  *)
    echo "dùng: tools/autostart.sh [--install|--uninstall|--status]" >&2
    exit 2
    ;;
esac
