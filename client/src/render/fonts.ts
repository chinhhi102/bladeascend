import { BitmapFont, Cache, type Texture } from 'pixi.js';

import { declareTexture } from './gpumem';

import { LANG } from '../i18n';
import { assets } from './assets';

/**
 * Font cho bảng tên (spec §12.3): Be Vietnam Pro, **đủ dấu tiếng Việt**.
 *
 * Font tải từ `assets/fonts/` của chính server chứ không lấy từ Google Fonts — game chạy
 * trong mạng LAN, máy người chơi có thể không ra được Internet, mà tên nhân vật mất dấu thì
 * còn tệ hơn là xấu.
 *
 * Chữ dựng bằng `BitmapText` chứ không phải `Text` (spec §12.3 phần hiệu năng): `Text` sinh
 * một texture mới mỗi lần đổi nội dung, mà bảng tên đổi liên tục (lên tầng, đổi hạng).
 * BitmapFont vẽ sẵn bộ ký tự một lần rồi ghép từng chữ cái.
 */

/**
 * MỘT font bitmap cho TẤT CẢ chữ trong thế giới — tên, cảnh giới, tiêu đề, số bay.
 *
 * Mỗi `BitmapFont` là một atlas riêng. Bộ ký tự tiếng Việt đầy đủ (~230 glyph) ở cỡ 36 cần
 * hai trang 512², tức 2,8 MB cho MỘT font; ba font là 8,5 MB trong khi trần của cả hạng mục
 * "Chữ/số bay" là 2 MB (§13.10). Gộp về một font sinh ở cỡ 26 thì vừa ĐÚNG một trang.
 *
 * Giá phải trả, nói thẳng: tên người chơi vẽ ở cỡ 36 nên bị phóng 1,38 lần từ atlas 26 —
 * mềm đi một chút. Chữ cảnh giới (26) và số bay (30) gần như không đổi. Đổi lại là bỏ được
 * hơn 6 MB, và `scenediff` là chỗ nói cuối cùng xem mắt có nhìn ra không.
 */
export const FONT_NAME = 'kx-text';
export const FONT_REALM = FONT_NAME;
/**
 * FONT_TITLE giờ TRỎ VÀO CHÍNH FONT_REALM — không còn atlas riêng.
 *
 * Mỗi `BitmapFont` là một atlas riêng, và với bộ ký tự tiếng Việt đầy đủ thì mỗi atlas ngốn
 * hai trang 512² = 2,8 MB. Ba font là 8,5 MB cho hạng mục "Chữ/số bay" có trần 2 MB (§13.10).
 *
 * Font tiêu đề chỉ khác font cảnh giới ở chỗ NGHIÊNG, và chữ nghiêng giả được bằng
 * `skew.x` ngay trên node (xem `nameplate.ts`) — một phép biến hình, không tốn điểm ảnh nào.
 */
export const FONT_TITLE = FONT_NAME;
/** TITLE_SKEW: độ nghiêng giả cho chữ tiêu đề, thay cho một atlas italic riêng. */
export const TITLE_SKEW = -0.22;

/** Bộ ký tự: ASCII in được + toàn bộ dấu tiếng Việt + vài ký hiệu bảng tên dùng tới. */
function charset(): string[] {
  const ascii = Array.from({ length: 95 }, (_, i) => String.fromCharCode(32 + i)).join('');
  const viet =
    'àáâãèéêìíòóôõùúýăđĩũơưạảấầẩẫậắằẳẵặẹẻẽếềểễệỉịọỏốồổỗộớờởỡợụủứừửữựỳỵỷỹ' + // i18n-ok: bộ ký tự font, không phải câu chữ
    'ÀÁÂÃÈÉÊÌÍÒÓÔÕÙÚÝĂĐĨŨƠƯẠẢẤẦẨẪẬẮẰẲẴẶẸẺẼẾỀỂỄỆỈỊỌỎỐỒỔỖỘỚỜỞỠỢỤỦỨỪỬỮỰỲỴỶỸ'; // i18n-ok: bộ ký tự font, không phải câu chữ
  const marks = '▲▼●★☯♚'; // chỉ báo nguy hiểm và vài ký hiệu hạng
  // Tiếng Thái (spec §19): sinh sẵn cả khối U+0E00–0E7F chỉ khi đang chọn tiếng Thái —
  // các ngôn ngữ khác không tốn texture cho 87 ký tự không bao giờ dùng.
  const thai = LANG === 'th' ? Array.from({ length: 0x80 }, (_, i) => String.fromCharCode(0x0e00 + i)).join('') : '';
  return [ascii + viet + marks + thai];
}

/** NAME_PX: cỡ SINH atlas (không phải cỡ vẽ). Xem ghi chú ở `FONT_NAME`. */
const NAME_PX = 28;

let ready: Promise<void> | null = null;

/** installed: ba font đã sinh, giữ lại để đo được (`fontAtlasInfo`). */
let installed: [string, unknown][] = [];

/** pagesOf: các trang atlas của một BitmapFont, bất kể Pixi gọi chúng là `pages` hay gì khác. */
function pagesOf(f: unknown): Texture[] {
  const pages = (f as { pages?: { texture: Texture }[] } | null)?.pages ?? [];
  return pages.map((p) => p.texture).filter(Boolean);
}

/**
 * fontAtlasInfo: mỗi font chiếm mấy trang, mỗi trang cỡ bao nhiêu điểm ảnh THẬT.
 *
 * Có nó vì hạng mục "Chữ/số bay" của sổ VRAM (§13.10) vượt trần 2 MB hơn bốn lần mà không
 * nói được vượt ở đâu: `CanvasPool` của Pixi chỉ để lại cái nhãn "canvas 512×512".
 */
/** px: cỡ ĐIỂM ẢNH THẬT của một texture (Pixi báo `width` theo cỡ logic). */
function px(t: Texture): { w: number; h: number } {
  const s2 = t.source as { pixelWidth?: number; pixelHeight?: number; width: number; height: number; resolution?: number };
  return {
    w: s2.pixelWidth ?? Math.round(s2.width * (s2.resolution ?? 1)),
    h: s2.pixelHeight ?? Math.round(s2.height * (s2.resolution ?? 1)),
  };
}

export function fontAtlasInfo(): { name: string; pages: string[]; mb: number }[] {
  return installed.map(([name, f]) => {
    const ps = pagesOf(f);
    return {
      name,
      // Điểm ảnh THẬT, không phải cỡ logic: `source.width` đã chia cho `resolution`, nên
      // tính theo nó thì một trang 512² báo thành 1 707² và con số MB sai gấp 11 lần.
      pages: ps.map((t) => `${px(t).w}x${px(t).h}`),
      mb: +(ps.reduce((n, t) => n + px(t).w * px(t).h * 4, 0) / 1048576).toFixed(2),
    };
  });
}

/** loadFonts nạp webfont rồi sinh ba BitmapFont. Gọi một lần lúc khởi động. */
export function loadFonts(): Promise<void> {
  if (ready) {
    return ready;
  }
  ready = (async () => {
    // Tên tệp trong `assets/dist` có hash → phải hỏi manifest, không ghép tay.
    const faces = (
      [
        ['fonts/BeVietnamPro-700.ttf', '700'],
        ['fonts/BeVietnamPro-800.ttf', '800'],
      ] as const
    )
      .map(([rel, weight]) => ({ url: assets.url(rel), weight }))
      .filter((f) => f.url)
      .map((f) => new FontFace('Be Vietnam Pro', `url(${f.url})`, { weight: f.weight }));
    // Noto Sans Thai CHỈ tải khi chọn tiếng Thái ("chỉ tải font của ngôn ngữ đang chọn").
    const thaiUrl = LANG === 'th' ? assets.url('fonts/NotoSansThai.ttf') : '';
    if (thaiUrl) {
      faces.push(new FontFace('Noto Sans Thai', `url(${thaiUrl})`, { weight: '100 900' }));
      document.documentElement.classList.add('lang-th');
    }
    try {
      await Promise.all(faces.map(async (f) => document.fonts.add(await f.load())));
    } catch {
      // Không tải được font thì vẫn chạy tiếp: BitmapFont rơi về font hệ thống, chữ hơi
      // khác nhưng không mất dấu (system-ui trên mọi hệ đều có bộ tiếng Việt).
    }
    // Ký tự ngoài bộ (tên người chơi chữ Hán, emoji…): Pixi 8 DynamicBitmapFont tự vẽ thêm
    // glyph lúc gặp (`ensureCharacters`), rơi về font hệ thống theo danh sách dưới — nên
    // không mất chữ (spec §19 "vẽ bằng Text thường" đạt cùng kết quả, có cache sẵn).
    const family = LANG === 'th' ? 'Be Vietnam Pro, Noto Sans Thai, system-ui, sans-serif' : 'Be Vietnam Pro, system-ui, sans-serif';
    const chars = charset();
    // Sinh ở cỡ lớn rồi thu nhỏ khi vẽ: chữ nét hơn trên màn hình độ phân giải cao.
    // Cỡ sinh HẠ XUỐNG để mỗi atlas vừa đúng MỘT trang 512².
    //
    // Bộ ký tự tiếng Việt đầy đủ (~230 glyph) ở cỡ 36 chiếm khoảng 1,3 trang, nên Pixi mở
    // trang thứ hai và một font tốn 2,8 MB thay vì 1,4. Chữ vẫn vẽ ở cỡ cũ (36/30/26) —
    // `BitmapText` co giãn glyph — nên thứ đổi là độ nét, không phải bố cục.
    // `resolution: 1` — mặc định của Pixi là 2, và đó là thủ phạm chính.
    //
    // DynamicBitmapFont vẽ mỗi glyph ở gấp đôi cỡ rồi thu lại, tức DIỆN TÍCH gấp bốn. Với
    // ~230 glyph tiếng Việt thì một font đầy hai trang 512² thay vì một. Mà chữ trong game
    // nằm trong THẾ GIỚI, rồi còn bị camera thu nhỏ (zoom thường < 1) — một cái tên cỡ 36
    // chỉ cao khoảng 16–20 điểm ảnh trên màn hình. Raster 2× là vẽ chi tiết cho thứ không
    // ai thấy, và trả bằng 2/3 ngân sách của cả hạng mục.
    BitmapFont.install({
      // `padding: 2` thay vì mặc định 4. Lề 4 điểm ảnh quanh MỖI glyph, với ~230 glyph ở cỡ
      // 28, làm bộ ký tự tràn sang trang thứ hai một cách sít sao — và trang thứ hai tốn
      // nguyên 1 MB cho vài chục glyph cuối. Lề 2 vẫn đủ chặn rỉ màu giữa các glyph khi lọc
      // tuyến tính (glyph cách nhau 4 điểm ảnh, mip cấp 1 chỉ với sang 1 điểm).
      name: FONT_NAME, chars, resolution: 1, padding: 2,
      style: { fontFamily: family, fontSize: NAME_PX, fontWeight: '800', fill: 0xffffff },
    });
    // `BitmapFont.install` KHÔNG trả về font — nó đăng ký vào `Cache` dưới khoá
    // `<tên>-bitmap`. Lấy lại từ đó để khai hạng mục cho từng trang atlas: thiếu bước này,
    // mỗi trang chỉ còn cái nhãn "canvas 512×512" và hạng mục "Chữ/số bay" không nói được
    // nó vượt trần ở đâu.
    installed = [FONT_NAME].map((n) => [n, Cache.get(`${n}-bitmap`)]);
    for (const [name, f] of installed) {
      for (const pg of pagesOf(f)) {
        declareTexture(pg, 'dynamic', `font:${name}`);
      }
    }
  })();
  return ready;
}
