import { BitmapFont } from 'pixi.js';

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

export const FONT_NAME = 'kx-name';
export const FONT_REALM = 'kx-realm';
export const FONT_TITLE = 'kx-title';

/** Bộ ký tự: ASCII in được + toàn bộ dấu tiếng Việt + vài ký hiệu bảng tên dùng tới. */
function charset(): string[] {
  const ascii = Array.from({ length: 95 }, (_, i) => String.fromCharCode(32 + i)).join('');
  const viet =
    'àáâãèéêìíòóôõùúýăđĩũơưạảấầẩẫậắằẳẵặẹẻẽếềểễệỉịọỏốồổỗộớờởỡợụủứừửữựỳỵỷỹ' +
    'ÀÁÂÃÈÉÊÌÍÒÓÔÕÙÚÝĂĐĨŨƠƯẠẢẤẦẨẪẬẮẰẲẴẶẸẺẼẾỀỂỄỆỈỊỌỎỐỒỔỖỘỚỜỞỠỢỤỦỨỪỬỮỰỲỴỶỸ';
  const marks = '▲▼●★☯♚'; // chỉ báo nguy hiểm và vài ký hiệu hạng
  return [ascii + viet + marks];
}

let ready: Promise<void> | null = null;

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
    try {
      await Promise.all(faces.map(async (f) => document.fonts.add(await f.load())));
    } catch {
      // Không tải được font thì vẫn chạy tiếp: BitmapFont rơi về font hệ thống, chữ hơi
      // khác nhưng không mất dấu (system-ui trên mọi hệ đều có bộ tiếng Việt).
    }
    const family = 'Be Vietnam Pro, system-ui, sans-serif';
    const chars = charset();
    // Sinh ở cỡ lớn rồi thu nhỏ khi vẽ: chữ nét hơn trên màn hình độ phân giải cao.
    BitmapFont.install({ name: FONT_NAME, style: { fontFamily: family, fontSize: 36, fontWeight: '800', fill: 0xffffff }, chars });
    BitmapFont.install({ name: FONT_REALM, style: { fontFamily: family, fontSize: 26, fontWeight: '700', fill: 0xffffff }, chars });
    BitmapFont.install({ name: FONT_TITLE, style: { fontFamily: family, fontSize: 26, fontWeight: '700', fontStyle: 'italic', fill: 0xffffff }, chars });
  })();
  return ready;
}
