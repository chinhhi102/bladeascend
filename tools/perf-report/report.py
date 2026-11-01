#!/usr/bin/env python3
"""Báo cáo hiệu năng tự sinh từ `docs/perf_log.csv` (CLAUDE.md luật 12).

    cd client && npm run perf:report        # hoặc: python3 tools/perf-report/report.py

Ghi ra `docs/reports/<ngày>-hieu-nang/`: một PNG cho mỗi chỉ số + README.md.

Luật của báo cáo, giống PERF_LOG: **chỉ so các lần đo CÙNG kịch bản**. Mỗi biểu đồ là một
chỉ số × một kịch bản (kịch bản có nhiều lần đo nhất cho chỉ số đó), trục ngang là số lượt
trong CSV theo thứ tự ghi. Đường xám đứt là ngân sách.

Chỉ dùng Pillow (đã có cho `tools/make_brandart.py`) — không thêm thư viện vẽ biểu đồ.
"""
from __future__ import annotations

import csv
import datetime as dt
import pathlib
from collections import Counter

from PIL import Image, ImageDraw, ImageFont

ROOT = pathlib.Path(__file__).resolve().parents[2]
CSV = ROOT / "docs/perf_log.csv"
FONT = ROOT / "assets/src/fonts/BeVietnamPro-700.ttf"

W, H = 900, 420
PAD_L, PAD_R, PAD_T, PAD_B = 84, 28, 64, 70
BG = (255, 255, 255)
INK = (32, 30, 28)
MUTED = (110, 106, 100)
GRID = (230, 228, 224)
SERIES = (37, 99, 175)  # một chuỗi số → một màu, không cần chú giải
BUDGET = (150, 146, 140)

# (cột CSV, tên hiển thị, đơn vị, ngân sách, "thấp hơn là tốt?")
METRICS = [
    ("tick_p95_ms", "Tick server p95", "ms", 4.0, True),
    ("pkt_max_kb", "Gói snapshot lớn nhất", "KB", 6.0, True),
    ("kbps_avg", "Băng thông xuống / người chơi (TB)", "KB/s", 40.0, True),
    ("fps_avg", "FPS trung bình", "FPS", 40.0, False),
    ("draw_calls", "Draw call / khung", "", None, True),
    ("tex_mb", "Bộ nhớ texture", "MB", 180.0, True),
    ("blocking_mb", "Tải trước khi vào trận", "MB", 8.0, True),
    ("ttp_10mbps_s", "Chờ vào trận (4G 10 Mbps)", "s", 10.0, True),
]


def font(px: int) -> ImageFont.FreeTypeFont:
    try:
        return ImageFont.truetype(str(FONT), px)
    except OSError:
        return ImageFont.load_default()


def num(v: str | None) -> float | None:
    try:
        return float((v or "").strip())
    except ValueError:
        return None


def series_for(rows: list[dict], col: str) -> tuple[str, list[tuple[str, float, str]]]:
    """Kịch bản có nhiều lần đo nhất cho cột này, và các điểm (lượt, giá trị, commit)."""
    have = [r for r in rows if num(r.get(col)) is not None and (r.get("scenario") or "").strip()]
    if not have:
        return "", []
    scen = Counter(r["scenario"] for r in have).most_common(1)[0][0]
    pts = [(r["run"], num(r[col]), (r.get("commit") or "")[:7], (r.get("label") or "").strip()) for r in have if r["scenario"] == scen]
    return scen, pts


def nice_max(v: float) -> float:
    for e in range(-2, 6):
        for m in (1, 1.2, 1.6, 2, 2.4, 3, 4, 5, 6, 8):  # chia 4 vạch ra số tròn
            c = m * 10**e
            if c >= v:
                return c
    return v


def chart(title: str, unit: str, budget: float | None, scen: str, pts: list, out: pathlib.Path) -> None:
    img = Image.new("RGB", (W, H), BG)
    d = ImageDraw.Draw(img)
    d.text((PAD_L, 18), f"{title} — kịch bản {scen}", font=font(20), fill=INK)
    vals = [p[1] for p in pts] + ([budget] if budget else [])
    top = nice_max(max(vals) * 1.1 or 1)
    pw, ph = W - PAD_L - PAD_R, H - PAD_T - PAD_B
    y = lambda v: PAD_T + ph - v / top * ph  # noqa: E731
    f12 = font(13)
    for i in range(5):  # lưới mờ
        v = top * i / 4
        d.line([(PAD_L, y(v)), (W - PAD_R, y(v))], fill=GRID, width=1)
        lab = f"{v:g}"
        d.text((PAD_L - 10 - d.textlength(lab, font=f12), y(v) - 8), lab, font=f12, fill=MUTED)
    if unit:
        d.text((12, PAD_T - 26), unit, font=f12, fill=MUTED)
    if budget:
        yb = y(budget)
        for x in range(PAD_L, W - PAD_R, 12):
            d.line([(x, yb), (min(x + 6, W - PAD_R), yb)], fill=BUDGET, width=2)
        # Chú thích đường ngân sách ở góc trên phải, ngoài vùng cột — đặt cạnh đường thì
        # luôn có một cột nào đó đè lên.
        lab = f"- - -  ngân sách {budget:g} {unit}"
        d.text((W - PAD_R - d.textlength(lab, font=f12), 24), lab, font=f12, fill=MUTED)
    n = len(pts)
    slot = pw / max(n, 1)
    bw = max(6, min(44, slot * 0.55))
    for i, (run, v, commit, _label) in enumerate(pts):
        cx = PAD_L + slot * (i + 0.5)
        base = int(PAD_T + ph)
        y0 = int(min(y(v), base - 2))  # số 0 vẫn có một vạch mỏng thay vì lỗi vẽ
        box = [int(cx - bw / 2), y0, int(cx + bw / 2), base]
        if base - y0 >= 10:
            d.rounded_rectangle(box, radius=4, fill=SERIES)
        else:
            d.rectangle(box, fill=SERIES)
        # Nhãn số chỉ ở lượt đầu và lượt cuối — đủ đọc "trước → sau", không rối.
        if i in (0, n - 1):
            lab = f"{v:g}"
            d.text((cx - d.textlength(lab, font=f12) / 2, y(v) - 18), lab, font=f12, fill=INK)
        lab = f"#{run}"
        d.text((cx - d.textlength(lab, font=f12) / 2, PAD_T + ph + 8), lab, font=f12, fill=MUTED)
        if commit:
            d.text((cx - d.textlength(commit, font=font(11)) / 2, PAD_T + ph + 26), commit, font=font(11), fill=MUTED)
    d.line([(PAD_L, PAD_T + ph), (W - PAD_R, PAD_T + ph)], fill=MUTED, width=1)
    img.save(out, optimize=True)


def verdict(v: float, budget: float | None, lower: bool) -> str:
    if budget is None:
        return "—"
    ok = v <= budget if lower else v >= budget
    return "✅ đạt" if ok else "🔴 chưa đạt"


def main() -> None:
    rows = list(csv.DictReader(CSV.open(encoding="utf-8")))
    day = dt.date.today().isoformat()
    out = ROOT / "docs/reports" / f"{day}-hieu-nang"
    out.mkdir(parents=True, exist_ok=True)
    lines = [
        f"# Báo cáo hiệu năng — sinh tự động {day}",
        "",
        f"> Nguồn: `docs/perf_log.csv` ({len(rows)} lượt đo). Sinh bằng `npm run perf:report`.",
        "> Mỗi biểu đồ chỉ so các lượt **cùng kịch bản** (PERF_LOG §1). Đường xám đứt = ngân sách.",
        "",
        "| Chỉ số | Kịch bản | Lượt đầu | Lượt mới nhất | Ngân sách | Trạng thái | Lượt mới nhất là |",
        "|---|---|---|---|---|---|---|",
    ]
    charts = []
    for i, (col, title, unit, budget, lower) in enumerate(METRICS, 1):
        scen, pts = series_for(rows, col)
        if not pts:
            continue
        fn = f"{i:02d}_{col}.png"
        chart(title, unit, budget, scen, pts, out / fn)
        first, last = pts[0], pts[-1]
        b = f"{'<' if lower else '≥'} {budget:g} {unit}" if budget else "—"
        lines.append(f"| {title} | {scen} | {first[1]:g} {unit} (#{first[0]}) | {last[1]:g} {unit} (#{last[0]}) | {b} | {verdict(last[1], budget, lower)} | {last[3][:60]} |")
        charts.append((title, scen, fn, len(pts)))
    lines += ["", "## Biểu đồ", ""]
    for title, scen, fn, n in charts:
        lines += [f"### {title} ({scen}, {n} lượt)", "", f"![{title}]({fn})", ""]
    lines += [
        "## Đối chiếu",
        "",
        "Báo cáo làm tay ngày 01/11 (`docs/_raw/README.md` + 6 PNG) dùng số từ nhiều kịch bản",
        "khác nhau trong cùng một cột. Bảng trên chỉ so cùng kịch bản nên con số \"trước\" có thể",
        "khác bảng tay — đó là chủ ý, không phải lỗi.",
        "",
        "Đối chiếu lượt 01/11 (đo lại trong lượt cuối, #26–#27):",
        "",
        "| Chỉ số | Báo cáo tay 01/11 | Đo lại #26 | Ghi chú |",
        "|---|---|---|---|",
        "| Tick p95 | 3,0–3,3 ms | 4,01 ms (#26, máy bận) → **1,38 ms (#27, máy rảnh)** | #26 đo lúc máy chạy 2 agent song song; #27 cùng lệnh lúc rảnh — đạt |",
        "| Gói snapshot lớn nhất | 2,0–4,4 KB | 2,4 KB | khớp |",
        "| Băng thông / người | 4–7 KB/s | 5,7 KB/s | khớp |",
        "| Tải trước khi vào trận | 2,77 MB | 3,65 MB | bảng tay chỉ tính nhóm `boot`; #26 cộng cả `home` 0,89 MB |",
        "| Chờ vào trận 4G 10 Mbps | 2,2 s | 0,6 s bấm→trận · 8,8 s mở trang→trận | bảng tay là \"trang chủ hiện\" (#26 cũng 2,2 s) |",
        "| FPS / GPU iPhone thật | chưa đo | chưa đo | CẦN MÁY THẬT |",
        "",
    ]
    (out / "README.md").write_text("\n".join(lines), encoding="utf-8")
    print(f"{len(charts)} biểu đồ → {out.relative_to(ROOT)}/")


if __name__ == "__main__":
    main()
