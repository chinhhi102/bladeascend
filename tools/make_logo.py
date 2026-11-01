#!/usr/bin/env python3
"""Dựng 3 logo chữ Pattaya trong khung vàng — chủ dự án gửi 02/11/2026 (logo.zip).

    python3 tools/make_logo.py            # ghi đè 3 PNG trong assets/src/brand/
    python3 tools/make_logo.py --out DIR  # ghi ra chỗ khác để so trước khi thay

Đầu vào (đều nằm trong repo, chạy được ở mọi máy):
    assets/src/brand/brand_logo_frame.png   khung vẽ trên nền hồng sen (#FF00FF) để tách nền
    assets/src/fonts/Pattaya-Regular.ttf    font SIL OFL 1.1 (xem Pattaya-OFL.txt, CREDITS.md)

Đầu ra (assets/src/brand/):
    logo_vi_1dong.png   "Đạo Hữu Xin Dừng Bước" một dòng — trang chủ PC
    logo_vi_2dong.png   hai dòng, khung kéo cao ×1.45 — điện thoại + màn tải
    logo_en.png         "BladeAscend" — /en /pt /es /id /th

Bản gốc của chủ dự án viết bằng numpy và đường dẫn /tmp; bản này chỉ dùng Pillow (máy dev
không có numpy) và đường dẫn trong repo. Cùng thuật toán: tách nền hồng theo độ "hồng"
`((r+b)/2 − g − 60) / 80`, khử viền hồng, rồi đặt chữ vàng chuyển màu + viền nâu + bóng.

Font Pattaya KHÔNG phát hành cho người chơi (asset-pipeline chặn): chữ đã vẽ sẵn vào PNG,
người chơi chỉ tải ảnh WebP của đúng logo mình cần.
"""
import argparse
import pathlib

from PIL import Image, ImageChops, ImageDraw, ImageFilter, ImageFont, ImageMath

ROOT = pathlib.Path(__file__).resolve().parent.parent
BRAND = ROOT / "assets/src/brand"
FRAME = BRAND / "brand_logo_frame.png"
FONT = ROOT / "assets/src/fonts/Pattaya-Regular.ttf"

# Toạ độ lòng khung đo trên bản khung rộng 1376 px (bản chủ dự án dùng). Khung gốc 2752 px
# nên thu về đúng 1376 trước để toạ độ — và cỡ logo ra (1288 px) — khớp bản gốc.
FRAME_W = 1376
INNER = (283, 1097, 296, 472)  # x0, x1, y0, y1

GOLD_TOP, GOLD_MID, GOLD_BOT = (255, 240, 170), (245, 190, 60), (170, 95, 20)
STROKE = (62, 28, 8, 255)
SHADOW = (30, 10, 0, 200)


def magenta_mask(rgb: Image.Image) -> Image.Image:
    """Độ hồng 0..1 (mode F): 1 = nền hồng sen, 0 = khung."""
    r, g, b = (c.convert("F") for c in rgb.split())
    return ImageMath.eval("min(max(((r+b)/2 - g - 60)/80, 0.0), 1.0)", r=r, g=g, b=b)


def frame_rgba() -> tuple[Image.Image, tuple[int, int]]:
    """Khung đã tách nền, cắt sát; kèm toạ độ góc cắt (để dời toạ độ lòng khung)."""
    src = Image.open(FRAME).convert("RGB")
    src = src.resize((FRAME_W, round(src.height * FRAME_W / src.width)), Image.LANCZOS)
    m = magenta_mask(src)
    r, g, b = (c.convert("F") for c in src.split())
    # Khử viền hồng: kéo phần hồng thừa ra khỏi kênh đỏ và lam ở mép khung.
    spill = ImageMath.eval("max((r+b)/2 - g, 0.0) * m", r=r, g=g, b=b, m=m)
    r2 = ImageMath.eval("r - s*0.5", r=r, s=spill).convert("L")
    b2 = ImageMath.eval("b - s*0.5", b=b, s=spill).convert("L")
    alpha = ImageMath.eval("(1.0 - m) * 255", m=m).convert("L")
    img = Image.merge("RGBA", (r2, g.convert("L"), b2, alpha))
    bbox = img.getbbox()
    return img.crop(bbox), (bbox[0], bbox[1])


def gold_gradient(w: int, h: int) -> Image.Image:
    col = Image.new("RGB", (1, h))
    for y in range(h):
        t = y / max(1, h - 1)
        a, b, k = (GOLD_TOP, GOLD_MID, t / 0.5) if t < 0.5 else (GOLD_MID, GOLD_BOT, (t - 0.5) / 0.5)
        col.putpixel((0, y), tuple(int(a[i] + (b[i] - a[i]) * k) for i in range(3)))
    return col.resize((w, h))


def fit_font(lines: list[str], box_w: int, box_h: int, size_max: int = 140):
    for size in range(size_max, 20, -2):
        ft = ImageFont.truetype(str(FONT), size)
        dims = [ft.getbbox(ln) for ln in lines]
        ws = [d[2] - d[0] for d in dims]
        hs = [d[3] - d[1] for d in dims]
        gap = int(size * 0.12)
        if max(ws) <= box_w and sum(hs) + gap * (len(lines) - 1) <= box_h:
            break
    return ft, size, dims, ws, hs, gap


def text_layer(lines: list[str], box_w: int, box_h: int) -> Image.Image:
    ft, size, dims, ws, hs, gap = fit_font(lines, box_w, box_h)
    pad = 40
    w, h = box_w + pad * 2, box_h + pad * 2
    mask = Image.new("L", (w, h), 0)
    d = ImageDraw.Draw(mask)
    y = (h - (sum(hs) + gap * (len(lines) - 1))) // 2
    for ln, dm, lw, lh in zip(lines, dims, ws, hs):
        d.text(((w - lw) // 2 - dm[0], y - dm[1]), ln, font=ft, fill=255)
        y += lh + gap
    stroke = mask.filter(ImageFilter.MaxFilter(int(size * 0.13) | 1))
    shadow = stroke.filter(ImageFilter.GaussianBlur(size * 0.06))
    out = Image.new("RGBA", (w, h), (0, 0, 0, 0))
    out.paste(Image.new("RGBA", (w, h), SHADOW), (int(size * 0.05), int(size * 0.07)), shadow)
    out.paste(Image.new("RGBA", (w, h), STROKE), (0, 0), stroke)
    out.paste(gold_gradient(w, h).convert("RGBA"), (0, 0), mask)
    # Vệt sáng mép trên: mặt nạ co lại, mờ dần theo (1 − y/h)³, nửa độ đậm.
    fade = Image.new("L", (1, h))
    for yy in range(h):
        fade.putpixel((0, yy), int(255 * 0.5 * (1 - yy / max(1, h - 1)) ** 3))
    hl = ImageChops.multiply(mask.filter(ImageFilter.MinFilter(5)), fade.resize((w, h)))
    out.paste(Image.new("RGBA", (w, h), (255, 255, 230, 255)), (0, 0), hl)
    return out


def compose(lines: list[str], stretch: float = 1.0) -> Image.Image:
    fr, (ox, oy) = frame_rgba()
    if stretch != 1.0:
        fr = fr.resize((fr.width, int(fr.height * stretch)), Image.LANCZOS)
    x0, x1 = INNER[0] - ox, INNER[1] - ox
    y0, y1 = int((INNER[2] - oy) * stretch), int((INNER[3] - oy) * stretch)
    tl = text_layer(lines, int((x1 - x0) * 0.9), int((y1 - y0) * 0.86))
    canvas = Image.new("RGBA", fr.size, (0, 0, 0, 0))
    canvas.alpha_composite(fr)
    cx, cy = (x0 + x1) // 2, (y0 + y1) // 2
    canvas.alpha_composite(tl, (cx - tl.width // 2, cy - tl.height // 2))
    return canvas


LOGOS = {
    "logo_vi_2dong": (["Đạo Hữu", "Xin Dừng Bước"], 1.45),
    "logo_vi_1dong": (["Đạo Hữu Xin Dừng Bước"], 1.0),
    "logo_en": (["BladeAscend"], 1.0),
}


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--out", type=pathlib.Path, default=BRAND)
    out = ap.parse_args().out
    out.mkdir(parents=True, exist_ok=True)
    for name, (lines, stretch) in LOGOS.items():
        img = compose(lines, stretch)
        p = out / f"{name}.png"
        img.save(p, optimize=True)
        print(f"  {p}  {img.width}×{img.height}  {p.stat().st_size // 1024} KB")


if __name__ == "__main__":
    main()
