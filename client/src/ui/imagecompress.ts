/**
 * Thu nhỏ + nén ảnh góp ý ngay trong trình duyệt — ADMIN_SPEC §7.1.
 *
 * Cạnh dài ≤ 1 600 px, JPEG ≤ 300 KB. Vẽ lại qua canvas còn có lợi phụ: canvas không mang
 * EXIF theo, nên toạ độ GPS trong ảnh điện thoại không rời máy người chơi (server vẫn gỡ
 * lại lần nữa — không bao giờ tin client).
 *
 * Định dạng gửi đi luôn là JPEG: mọi trình duyệt mã hoá được (Safari không mã hoá WebP), và
 * server mã hoá lại được bằng thư viện chuẩn (Q39).
 */

/** MAX_SIDE: cạnh dài tối đa sau khi thu nhỏ (§7.1). */
export const MAX_SIDE = 1600;
/** MAX_BYTES: trần mỗi ảnh — trùng `FeedbackMaxImage` phía server. */
export const MAX_BYTES = 300 * 1024;

/** QUALITIES: hạ dần chất lượng cho tới khi lọt ngân sách. */
const QUALITIES = [0.85, 0.75, 0.65, 0.55, 0.45, 0.35];
/** SHRINK: lọt hết thang chất lượng mà vẫn nặng thì thu nhỏ thêm rồi thử lại. */
const SHRINK = 0.75;
const MAX_SHRINKS = 4;

/** ImageDecodeError: trình duyệt không đọc được tệp (thường là HEIC trên Chrome/Android). */
export class ImageDecodeError extends Error {}
/** ImageTooBigError: đã thu nhỏ hết mức mà vẫn quá 300 KB (rất hiếm). */
export class ImageTooBigError extends Error {}

interface Decoded {
  source: CanvasImageSource;
  width: number;
  height: number;
  release: () => void;
}

/** decode đọc ảnh, ưu tiên `createImageBitmap` (xoay theo EXIF orientation sẵn). */
async function decode(blob: Blob): Promise<Decoded> {
  if (typeof createImageBitmap === 'function') {
    try {
      const bmp = await createImageBitmap(blob, { imageOrientation: 'from-image' });
      return { source: bmp, width: bmp.width, height: bmp.height, release: () => bmp.close() };
    } catch (_) {
      // Safari cũ không nhận tuỳ chọn, hoặc không đọc được — thử đường <img>.
    }
  }
  return decodeWithImg(blob);
}

function decodeWithImg(blob: Blob): Promise<Decoded> {
  const url = URL.createObjectURL(blob);
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => resolve({ source: img, width: img.naturalWidth, height: img.naturalHeight,
      release: () => URL.revokeObjectURL(url) });
    img.onerror = () => {
      URL.revokeObjectURL(url);
      reject(new ImageDecodeError('decode'));
    };
    img.src = url;
  });
}

function toBlob(cv: HTMLCanvasElement, type: string, q: number): Promise<Blob | null> {
  return new Promise((resolve) => cv.toBlob(resolve, type, q));
}

/** draw vẽ ảnh lên canvas cỡ w×h; JPEG không có kênh alpha nên lót nền trắng trước. */
function draw(d: Decoded, w: number, h: number, opaque: boolean): HTMLCanvasElement {
  const cv = document.createElement('canvas');
  cv.width = w;
  cv.height = h;
  const g = cv.getContext('2d');
  if (!g) {
    throw new ImageDecodeError('canvas');
  }
  if (opaque) {
    g.fillStyle = '#fff';
    g.fillRect(0, 0, w, h);
  }
  g.imageSmoothingQuality = 'high';
  g.drawImage(d.source, 0, 0, w, h);
  return cv;
}

/** encodeUnder thử từng mức chất lượng; trả về blob đầu tiên lọt ngân sách, hoặc null. */
async function encodeUnder(cv: HTMLCanvasElement, type: string): Promise<Blob | null> {
  for (const q of QUALITIES) {
    const b = await toBlob(cv, type, q);
    if (!b || b.type !== type) {
      return null;
    }
    if (b.size <= MAX_BYTES) {
      return b;
    }
  }
  return null;
}

/**
 * compressImage trả về ảnh JPEG cạnh dài ≤ 1 600 px, ≤ 300 KB.
 *
 * Luôn JPEG (Q39, v6.34): server giải mã rồi mã hoá lại mọi ảnh bằng thư viện chuẩn Go, mà
 * Go chuẩn không đọc được WebP — gửi WebP là bị từ chối.
 */
export async function compressImage(blob: Blob): Promise<Blob> {
  const d = await decode(blob);
  try {
    if (!d.width || !d.height) {
      throw new ImageDecodeError('empty');
    }
    const type = 'image/jpeg';
    let scale = Math.min(1, MAX_SIDE / Math.max(d.width, d.height));
    for (let i = 0; i <= MAX_SHRINKS; i++, scale *= SHRINK) {
      const w = Math.max(1, Math.round(d.width * scale));
      const h = Math.max(1, Math.round(d.height * scale));
      const out = await encodeUnder(draw(d, w, h, type === 'image/jpeg'), type);
      if (out) {
        return out;
      }
    }
    throw new ImageTooBigError('too big');
  } finally {
    d.release();
  }
}

/** blobToDataURL: gói ảnh để gửi trong JSON (server nhận data URL base64). */
export function blobToDataURL(b: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const r = new FileReader();
    r.onload = () => resolve(String(r.result));
    r.onerror = () => reject(r.error ?? new Error('read'));
    r.readAsDataURL(b);
  });
}

/** dataURLToBlob: ảnh chụp màn hình đến dưới dạng data URL. */
export async function dataURLToBlob(url: string): Promise<Blob> {
  const r = await fetch(url);
  return r.blob();
}
