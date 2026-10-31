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


ART = ROOT / "assets/src"

# Bốn nhân vật lên ảnh chia sẻ. Chọn cho KHÁC SILHOUETTE nhau — mập/gầy/cao/lùn — vì ở
# 400 px thì chi tiết mặt biến mất và chỉ còn dáng người đọc được.
CARD_CHARS = ["a_hon", "co_ba_ganh", "shipper", "cu_rua"]
HERO = "dai_ca_dep"  # người đứng giữa, có vòng kiếm


def split_name(name: str) -> list[str]:
    """Chia tên game làm hai cụm — cùng phép chia với `client/brand-split.ts`.

    Ngắt ở đây phải TRÙNG với trang chủ và màn hình tải, nếu không ảnh chia sẻ xuống dòng
    một kiểu còn trang web kiểu khác.
    """
    w = name.split()
    if len(w) < 4:
        return [name]
    cut = len(w) // 2
    return [" ".join(w[:cut]), " ".join(w[cut:])]


def sword_ring(size: int, radius: int, blades: int = 10, front: bool | None = None) -> Image.Image:
    """Vòng kiếm NHIỀU MÀU quanh một nhân vật.

    Dùng thẳng art kiếm của 9 cảnh giới (`assets/src/swords/realm_*.png`) chứ không tô màu
    giả: chín thanh đó vốn đã khác màu nhau theo cảnh giới, nên "vòng kiếm nhiều màu" trên
    ảnh chia sẻ đúng là thứ người chơi thấy trong game, không phải hiệu ứng riêng cho ảnh
    quảng cáo.

    Mũi kiếm hướng RA NGOÀI theo bán kính — cùng lý do đã ghi ở `icon()`: kiếm nằm tiếp
    tuyến thì ở cỡ nhỏ đọc thành một vòng hạt cườm.

    `front` chia vòng làm hai nửa để dán TRƯỚC và SAU người:

        None  — cả vòng
        False — nửa trên (sau lưng)
        True  — nửa dưới (trước mặt)

    Dán cả vòng ra sau người thì nửa dưới khuất hẳn sau thân, và cái còn nhìn thấy là vài
    thanh kiếm rải rác quanh đầu — trông như đồ rơi, không như kiếm đang bay quanh. Trong
    game kiếm bay cả trước lẫn sau người; ảnh chia sẻ phải nói đúng điều đó.
    """
    ring = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    files = sorted((ART / "swords").glob("realm_*.png"))
    if not files:
        return ring
    cx = cy = size / 2
    for i in range(blades):
        th = 360 * i / blades
        # 0° là bên phải, 90° là PHÍA DƯỚI (trục y của ảnh hướng xuống) — nửa dưới là nửa
        # nằm trước mặt người.
        if front is not None and (0 < th < 180) != front:
            continue
        src = Image.open(files[i % len(files)]).convert("RGBA")
        # Cắt sát viền TRƯỚC khi thu nhỏ. Art kiếm là canvas 256×256 mà lưỡi chỉ chiếm
        # khoảng nửa; thu thẳng cả canvas thì lưỡi thật chỉ còn một nửa cỡ mình tưởng —
        # đó là lý do vòng kiếm ở hai bản trước trông thưa và nhỏ dù số đo bán kính đúng.
        bb = src.getbbox()
        if bb:
            src = src.crop(bb)
        bl = int(size * 0.30)
        blade = src.resize((max(1, int(src.width * bl / src.height)), bl), Image.LANCZOS)
        # `expand=True` để lưỡi không bị cắt góc khi xoay; rồi dán theo tâm đã xoay.
        rot = blade.rotate(-th - 90, resample=Image.BICUBIC, expand=True)
        px = cx + radius * math.cos(math.radians(th)) - rot.width / 2
        py = cy + radius * math.sin(math.radians(th)) - rot.height / 2
        ring.alpha_composite(rot, (int(px), int(py)))
    return ring


def fit_char(cid: str, height: int) -> Image.Image | None:
    """Lấy ảnh đứng của một nhân vật, cắt sát viền rồi thu về đúng chiều cao."""
    p = ART / "characters" / cid / "portrait_neutral.png"
    if not p.exists():
        return None
    im = Image.open(p).convert("RGBA")
    bb = im.getbbox()  # bỏ viền trong suốt, nếu không mỗi nhân vật cao một kiểu
    if bb:
        im = im.crop(bb)
    w = max(1, int(im.width * height / im.height))
    return im.resize((w, height), Image.LANCZOS)


def share_card(w: int, h: int, name_vi: str, tag: str, foot: str) -> Image.Image:
    """Ảnh chia sẻ OG 1200×630 — chủ dự án chốt 1/11.

    Bản cũ chỉ là chữ vàng trên nền nâu trơn: đúng tên game nhưng không nói game là gì, và
    trên bảng tin Facebook nó trông y hệt một cái thẻ lỗi.

    Ràng buộc chi phối mọi cỡ chữ ở đây: **phải đọc được khi Zalo/Messenger thu còn
    400 px** — tức mọi thứ bị chia ba. Chữ nhỏ nhất trên ảnh là 34 px, còn 11 px sau khi
    thu; dưới mức đó là một vệt xám. Vì vậy ảnh này cố tình ít chữ.
    """
    # ── Nền: chính cảnh chợ quê của trang chủ ──────────────────────────────────────────
    bg_path = ART / "ui/home_bg.png"
    if bg_path.exists():
        bg = Image.open(bg_path).convert("RGBA")
        # Cắt theo đúng tỉ lệ 1200:630 rồi thu — không kéo giãn, vì kéo giãn thì mái chùa
        # và nón lá méo thấy rõ.
        tgt = w / h
        bw, bh = bg.size
        if bw / bh > tgt:
            nw = int(bh * tgt)
            bg = bg.crop(((bw - nw) // 2, 0, (bw + nw) // 2, bh))
        else:
            nh = int(bw / tgt)
            bg = bg.crop((0, (bh - nh) // 2, bw, (bh + nh) // 2))
        img = bg.resize((w, h), Image.LANCZOS)
    else:
        img = Image.new("RGBA", (w, h), DARK)

    # Màn tối chuyển dần: đậm ở trên (chỗ đặt tên game) và dưới (chỗ chữ nhỏ), nhạt ở giữa
    # để cảnh chợ còn nhìn ra. Không phủ đều một lớp: phủ đều thì chữ đọc được nhưng nền
    # thành một mảng bùn, mất luôn lý do dùng cảnh chợ.
    scrim = Image.new("L", (w, 1))
    for x in range(w):
        t = x / (w - 1)
        # Đậm hẳn ở nửa TRÁI (nền của khối chữ), nhạt dần sang phải để cảnh chợ còn nhìn
        # ra sau vòng kiếm. Phủ đều một lớp thì chữ đọc được nhưng nền thành một mảng bùn,
        # mất luôn lý do dùng cảnh chợ.
        scrim.putpixel((x, 0), int(255 * (0.82 - 0.56 * min(1.0, max(0.0, (t - 0.30) / 0.45)))))
    dark = Image.new("RGBA", (w, h), (12, 9, 6, 255))
    dark.putalpha(scrim.resize((w, h)))
    img.alpha_composite(dark)

    # ── Bố cục TRÁI / PHẢI ─────────────────────────────────────────────────────────────
    #
    # Ba bản trước xếp chồng theo chiều dọc: tên game trên, nhân vật dưới. Ảnh 1200×630
    # rất NGANG, nên cách đó để tên game ăn hết 340 px đầu và phần còn lại không đủ chỗ
    # cho một vòng tròn — vòng kiếm luôn bị cắt hoặc khuất sau đầu nhân vật, và cái đọc
    # ra là mấy thanh kiếm rơi vãi. Nới bán kính không cứu được: hết chỗ theo chiều cao.
    #
    # Chia dọc thì mỗi nửa làm đúng một việc: trái là CHỮ (game tên gì, chơi ra sao),
    # phải là HÌNH (game trông thế nào). Vòng kiếm có trọn nửa phải để tròn.
    # `cx_art` và `base_y` chọn sao cho CẢ VÒNG nằm trong khung: bản trước để 0,76/0,88 và
    # vòng thò ra khỏi mép phải lẫn mép dưới, mất ba lưỡi. Vòng bị cắt thì nó thôi là vòng.
    cx_art = int(w * 0.73)
    base_y = int(h * 0.80)
    hero = fit_char(HERO, int(h * 0.40))
    if hero:
        # Bán kính tính theo CHIỀU CAO NGƯỜI, và rộng hơn hẳn bóng người: lưỡi kiếm nằm
        # ngoài thân ở mọi hướng thay vì chồng lên đầu và vai.
        r = int(hero.height * 0.72)
        ring_size = int(r * 2 + hero.height * 0.36)
        hy = base_y - hero.height // 2
        rx, ry = cx_art - ring_size // 2, hy - ring_size // 2
        img.alpha_composite(sword_ring(ring_size, r, front=False), (rx, ry))
        img.alpha_composite(hero, (cx_art - hero.width // 2, base_y - hero.height))
        img.alpha_composite(sword_ring(ring_size, r, front=True), (rx, ry))

    # Ba người còn lại đứng THẤP DƯỚI NỬA TRÁI: có mặt để nói "mười nhân vật dở hơi",
    # nhưng nhỏ và thấp nên không tranh chỗ với khối chữ hay với vòng kiếm.
    for im, sx in zip([fit_char(cid, int(h * 0.20)) for cid in CARD_CHARS[:3]], [0.08, 0.22, 0.36]):
        if im:
            img.alpha_composite(im, (int(w * sx) - im.width // 2, int(h * 0.99) - im.height))

    # ── Chữ: cả khối CĂN TRÁI trong nửa trái ───────────────────────────────────────────
    d = ImageDraw.Draw(img)
    x0 = int(w * 0.055)
    col = int(w * 0.52)  # bề ngang dành cho chữ; phần còn lại là của vòng kiếm
    lines = split_name(name_vi)
    # Cỡ chữ lấy theo dòng DÀI NHẤT, để hai dòng cùng cỡ. Đo từng dòng rồi lấy min, chứ
    # tính theo số ký tự thì "Xin Dừng Bước" và "Đạo Hữu" ra hai cỡ lệch hẳn.
    f_name = min((fit(d, ln, col, int(h * 0.165)) for ln in lines), key=lambda f: f.size)
    y = int(h * 0.17)
    for ln in lines:
        bb = d.textbbox((0, 0), ln, font=f_name)
        # Viền tối dày quanh chữ: ảnh chia sẻ hiện trên nền trắng (Facebook) lẫn nền tối
        # (Discord), và ngay dưới chữ là cảnh chợ nhiều màu.
        for dx, dy in ((0, 4), (4, 0), (-4, 0), (0, -4), (3, 3), (-3, 3)):
            d.text((x0 - bb[0] + dx, y - bb[1] + dy), ln, font=f_name, fill=(38, 22, 6, 255))
        d.text((x0 - bb[0], y - bb[1]), ln, font=f_name, fill=GOLD)
        y += int(f_name.size * 1.04)

    # Khẩu hiệu và dòng chân đặt trên DẢI NỀN RIÊNG, không chỉ viền chữ.
    #
    # Viền chữ đủ cho tên game vì tên game to và dày. Hai dòng này nhỏ hơn ba lần và nằm
    # đè lên cảnh chợ nhiều màu — một bản trước chúng chìm hẳn, ở 400 px không đọc ra chữ
    # nào. Dải tối sau lưng là cách rẻ nhất để tương phản không phụ thuộc vào chỗ chữ rơi
    # trúng trên nền.
    def banner(text: str, max_h: float, top: float, color: tuple) -> int:
        f = fit(d, text, col, int(h * max_h))
        bb = d.textbbox((0, 0), text, font=f)
        tw, th = bb[2] - bb[0], bb[3] - bb[1]
        px, py = int(h * 0.026), int(h * 0.022)
        bx, by = x0 - px, int(top)
        strip = Image.new("RGBA", (tw + 2 * px, th + 2 * py), (0, 0, 0, 0))
        ImageDraw.Draw(strip).rounded_rectangle(
            [0, 0, strip.width - 1, strip.height - 1], radius=int(h * 0.026), fill=(14, 10, 7, 220))
        img.alpha_composite(strip, (max(0, bx), by))
        ImageDraw.Draw(img).text((max(0, bx) + px - bb[0], by + py - bb[1]), text, font=f, fill=color)
        return by + strip.height

    y = banner(tag, 0.060, y + h * 0.02, PAPER)
    # Dòng chân: lời hứa bán hàng duy nhất trên ảnh. Giữ to hơn mức "chữ nhỏ" thông thường
    # vì đây chính là câu quyết định người lạ có bấm hay không.
    banner(foot, 0.055, y + h * 0.028, GOLD)
    return img.convert("RGB")


def main() -> None:
    UI.mkdir(parents=True, exist_ok=True)
    name_vi, tag, name, domain = BRAND["name_vi"], BRAND["tagline_vi"], BRAND["name"], BRAND["domain"]
    out = {
        "logo_vi.png": wordmark(1024, 384, None, name_vi, tag),
        "icon_192.png": icon(192),
        "icon_512.png": icon(512),
        "share_card.png": share_card(1200, 630, name_vi, tag, "Chơi miễn phí trên trình duyệt"),
    }
    for fn, img in out.items():
        p = UI / fn
        img.save(p)
        print(f"  {fn:<18} {img.width}×{img.height}  {p.stat().st_size // 1024} KB")
    print(f"\n{len(out)} tệp trong {UI.relative_to(ROOT)} — Vite chép nguyên xi sang dist/, không qua asset-pipeline.")


if __name__ == "__main__":
    main()
