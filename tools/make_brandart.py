#!/usr/bin/env python3
"""Dựng art thương hiệu bằng FONT — GAME_SPEC §20.

    python3 tools/make_brandart.py

Sinh ra 4 tệp trong `client/public/brand/`:

    logo_vi.png     1024×384   chữ tên game + khẩu hiệu, nền trong suốt
    icon_192.png     192×192   icon PWA
    icon_512.png     512×512   icon PWA (dùng luôn cho maskable)
    share_card.png  1200×630   ảnh chia sẻ OG

── Vì sao `client/public/` chứ không phải `assets/src/` ───────────────────────────────

Art của game đi qua `tools/asset-pipeline`: nén thành WebP và **gắn hash vào tên tệp**, rồi
code lấy địa chỉ thật bằng `assets.url(...)`. Bốn tệp này KHÔNG dùng được đường đó:

- Facebook, Zalo và Discord đọc `og:image` bằng con bọ riêng, không chạy JavaScript — nên
  không có cách nào gọi `assets.url()` cho chúng. Bản đầu của tôi ghi thẳng
  `/assets/ui/share_card.png`, và vì tên thật có hash nên đó là một đường dẫn 404 câm.
- Địa chỉ đổi mỗi lần build là một vấn đề RIÊNG: các mạng xã hội giữ ảnh chia sẻ trong
  cache rất lâu, và icon PWA đổi đường dẫn thì điện thoại tải lại icon mỗi bản build.
- WebP: Twitter không nhận ảnh chia sẻ WebP, và icon PWA khai `type: image/png` mà nội
  dung là WebP thì Android bỏ qua.

`client/public/` được Vite chép nguyên xi sang `dist/`, giữ đúng tên và đúng định dạng.

Vì sao một tệp Python chứ không phải một job `artgen`:

§20 ghi "logo chữ: tạm render bằng font". Chữ render bằng font thì phải ra **đúng từng
pixel như nhau mỗi lần chạy** — mô hình sinh ảnh không làm được việc đó, và nó cũng không
viết nổi tiếng Việt có dấu. Job `artgen` cho logo VẼ (có hình kiếm, có hoa văn) vẫn còn
trong danh mục với ưu tiên 3; tệp này lo phần chữ cho tới khi logo vẽ xong.

Lấy tên và khẩu hiệu từ `config/game.json` — cùng một nguồn với `<title>`, manifest và OG,
nên không có chỗ nào ghi tên game lần thứ hai.

Chạy lại sau khi đổi `config.brand`, rồi `node tools/asset-pipeline/index.mjs`.
"""
import json
import math
import pathlib

from PIL import Image, ImageDraw, ImageFont

ROOT = pathlib.Path(__file__).resolve().parent.parent
UI = ROOT / "client/public/brand"
FONT = ROOT / "assets/src/fonts/BeVietnamPro-800.ttf"

BRAND = json.loads((ROOT / "config/game.json").read_text(encoding="utf-8"))["brand"]

GOLD = (255, 214, 107, 255)
GOLD_DEEP = (107, 74, 18, 255)
PAPER = (203, 182, 138, 255)
DARK = (20, 17, 14, 255)
INK_DARK = (29, 26, 22, 255)


def font(px: int) -> ImageFont.FreeTypeFont:
    return ImageFont.truetype(str(FONT), px)


def fit(draw: ImageDraw.ImageDraw, text: str, max_w: int, start: int) -> ImageFont.FreeTypeFont:
    """Tìm cỡ chữ lớn nhất còn vừa `max_w`.

    Đo bằng `textbbox` chứ không nhân ước lượng theo số ký tự: tiếng Việt có dấu nên chiều
    cao và bề rộng không tỉ lệ với độ dài chuỗi, và "Đạo Hữu Xin Dừng Bước" dài hơn
    "BladeAscend" gần gấp đôi.
    """
    px = start
    while px > 8:
        f = font(px)
        if draw.textbbox((0, 0), text, font=f)[2] <= max_w:
            return f
        px -= 2
    return font(8)


def wordmark(w: int, h: int, bg: tuple | None, name: str, tagline: str, foot: str = "") -> Image.Image:
    img = Image.new("RGBA", (w, h), bg or (0, 0, 0, 0))
    d = ImageDraw.Draw(img)
    pad = int(w * 0.07)

    f_name = fit(d, name, w - 2 * pad, int(h * 0.30))
    bb = d.textbbox((0, 0), name, font=f_name)
    nx, ny = (w - bb[2]) // 2 - bb[0], int(h * 0.30) - bb[1]
    # Viền tối dưới chữ: logo phải đọc được cả trên nền sáng và nền tối, vì ảnh chia sẻ
    # hiện trên Facebook (nền trắng) và trong Discord (nền tối).
    d.text((nx, ny + max(2, h // 150)), name, font=f_name, fill=GOLD_DEEP)
    d.text((nx, ny), name, font=f_name, fill=GOLD)

    f_tag = fit(d, tagline, w - 2 * pad, int(h * 0.12))
    bb = d.textbbox((0, 0), tagline, font=f_tag)
    d.text(((w - bb[2]) // 2 - bb[0], ny + f_name.size + int(h * 0.06)), tagline, font=f_tag, fill=PAPER)

    if foot:
        f_foot = fit(d, foot, w - 2 * pad, int(h * 0.09))
        bb = d.textbbox((0, 0), foot, font=f_foot)
        d.text(((w - bb[2]) // 2 - bb[0], h - pad - (bb[3] - bb[1])), foot, font=f_foot, fill=GOLD)
    return img


def icon(size: int) -> Image.Image:
    """Icon: một vòng kiếm quanh tâm — hình chính của game, nhận ra được ở 192 px.

    Không đặt chữ vào icon: ở 192 px thì "BladeAscend" cao 9 px và thành một vệt xám. Vòng
    kiếm giữ được hình ở mọi cỡ, và nó đúng là thứ người chơi nhìn thấy suốt cả trận.

    Kiếm hướng MŨI RA NGOÀI theo bán kính. Bản đầu tôi vẽ kiếm nằm tiếp tuyến cho "đúng
    vật lý quỹ đạo" — ở cỡ icon chúng thành mười viên kim cương xếp vòng, không ai đọc ra
    là kiếm. Hình ở cỡ nhỏ phải đọc được trước, đúng sau.
    """
    ss = 4  # vẽ to gấp 4 rồi thu nhỏ — cạnh nghiêng mới mịn
    s = size * ss
    img = Image.new("RGBA", (s, s), DARK)
    d = ImageDraw.Draw(img)
    cx = cy = s / 2

    blades, r0, r1 = 10, s * 0.17, s * 0.44
    for i in range(blades):
        th = 2 * math.pi * i / blades - math.pi / 2
        ux, uy = math.cos(th), math.sin(th)   # hướng ra ngoài
        px, py = -uy, ux                      # vuông góc, để tạo bề ngang lưỡi
        half = s * 0.021                      # nửa bề ngang chuôi lưỡi
        tip = (cx + ux * r1, cy + uy * r1)
        # Lưỡi: chuôi rộng, vuốt nhọn dần ra mũi. Bốn điểm, không phải hình thoi cân.
        d.polygon(
            [
                tip,
                (cx + ux * (r0 + half) + px * half, cy + uy * (r0 + half) + py * half),
                (cx + ux * r0, cy + uy * r0),
                (cx + ux * (r0 + half) - px * half, cy + uy * (r0 + half) - py * half),
            ],
            fill=GOLD,
        )
        # Chắn tay: một gạch ngắn ngang chuôi, đủ để mắt đọc ra "kiếm" chứ không phải "gai".
        g = s * 0.038
        d.line(
            [(cx + ux * r0 * 1.02 + px * g, cy + uy * r0 * 1.02 + py * g),
             (cx + ux * r0 * 1.02 - px * g, cy + uy * r0 * 1.02 - py * g)],
            fill=GOLD_DEEP, width=max(1, int(s * 0.012)),
        )

    # Người ở tâm: một đĩa nhỏ, KHÔNG quầng sáng.
    # Bản đầu có quầng 14 lớp alpha chồng lên nhau; chúng cộng thành một vành trắng và cả
    # icon trông như quả trứng ốp la.
    r = s * 0.105
    d.ellipse([cx - r, cy - r, cx + r, cy + r], fill=PAPER)
    r2 = r * 0.62
    d.ellipse([cx - r2, cy - r2, cx + r2, cy + r2], fill=GOLD_DEEP)
    return img.resize((size, size), Image.LANCZOS)


def main() -> None:
    UI.mkdir(parents=True, exist_ok=True)
    name_vi, tag, name, domain = BRAND["name_vi"], BRAND["tagline_vi"], BRAND["name"], BRAND["domain"]
    out = {
        "logo_vi.png": wordmark(1024, 384, None, name_vi, tag),
        "icon_192.png": icon(192),
        "icon_512.png": icon(512),
        "share_card.png": wordmark(1200, 630, DARK, name_vi, tag, f"{name} · {domain}"),
    }
    for fn, img in out.items():
        p = UI / fn
        img.save(p)
        print(f"  {fn:<18} {img.width}×{img.height}  {p.stat().st_size // 1024} KB")
    print(f"\n{len(out)} tệp trong {UI.relative_to(ROOT)} — Vite chép nguyên xi sang dist/, không qua asset-pipeline.")


if __name__ == "__main__":
    main()
