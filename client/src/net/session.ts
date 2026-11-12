import type { Profile } from '../ui/home';

/**
 * Mã phiên để vào lại đúng nhân vật sau khi mất kết nối — GAME_SPEC §7.1 (v6.53).
 *
 * ── Nó thay thế cái gì ───────────────────────────────────────────────────────────────────
 *
 * Trước đây chỉ có `pendingRejoin` trong `render/glrecover.ts`: nó nhớ **tên + nhân vật**
 * rồi sau khi tải lại trang thì vào một trận MỚI. Nghe thì giống, nhưng với người chơi là
 * hai chuyện khác hẳn — vào trận mới nghĩa là mất sạch cảnh giới, kiếm và Uy Danh của ván
 * đang chơi. Và nó chỉ chạy đúng một trường hợp (mất ngữ cảnh WebGL), không chạy khi F5 hay
 * khi mạng rớt.
 *
 * Bây giờ server cấp **mã phiên 32 byte** lúc vào trận; nối lại kèm mã thì được gắn lại
 * ĐÚNG nhân vật cũ (server giữ nó trong thế giới 20 giây).
 *
 * ── Vì sao sessionStorage ────────────────────────────────────────────────────────────────
 *
 * Sống qua F5, chết khi đóng tab — đúng tuổi thọ của một ván. `localStorage` thì sai: mở
 * lại game ngày hôm sau sẽ gửi một mã từ hôm qua. Và mã là **bí mật** (ai có nó thì chiếm
 * được nhân vật), nên càng không nên nằm lâu.
 *
 * Bản dự phòng trong bộ nhớ là yêu cầu thẳng của spec: chế độ ẩn danh / iOS khoá bên thứ ba
 * ném lỗi ngay ở `setItem`. Nó không sống qua F5, nhưng nó CỨU được trường hợp mất kết nối
 * mà không tải lại trang (mạng rớt rồi WebSocket tự nối lại) — đúng trường hợp hay gặp nhất.
 */

const KEY = 'kx.session';

/**
 * REMEMBER_SECS: client ngừng gửi mã sau ngần này giây.
 *
 * Dài hơn 20 giây của server (`HoldSecs`) một cách CÓ CHỦ Ý: quá 20 giây thì nhân vật đã
 * rời thế giới, nhưng nếu họ **bị chém chết** trong lúc mất kết nối thì server còn giữ màn
 * Tổng kết 120 giây để trả lại (`deadKeepSecs`). Client ngừng sớm = người chơi tải lại
 * trang trên 4G chậm sẽ thấy trang chủ và không bao giờ biết mình đã chết.
 *
 * Ai quyết định là server, không phải client: mã hết hiệu lực thì server cho vào trận mới.
 */
const REMEMBER_SECS = 120;

export interface MatchSession {
  key: string;
  name: string;
  char: string;
  /** at: mốc Unix (giây) lúc lưu, để bỏ mã quá cũ. */
  at: number;
}

/** mem: bản dự phòng khi trình duyệt chặn sessionStorage (§7.1). */
let mem: MatchSession | null = null;

/** saveSession ghi mã phiên server vừa cấp, kèm tên + nhân vật để gửi lại gói `join`. */
export function saveSession(key: string, p: Profile): void {
  if (!key) {
    return;
  }
  const s: MatchSession = { key, name: p.name, char: p.char, at: Date.now() / 1000 };
  mem = s;
  try {
    sessionStorage.setItem(KEY, JSON.stringify(s));
  } catch {
    // Ẩn danh chặn ghi: `mem` vẫn cứu được lần nối lại không qua tải trang.
  }
}

/** loadSession đọc phiên còn hạn; null nếu không có hoặc đã quá cũ. */
export function loadSession(): MatchSession | null {
  let s = mem;
  try {
    const raw = sessionStorage.getItem(KEY);
    if (raw) {
      s = JSON.parse(raw) as MatchSession;
    }
  } catch {
    // đọc không được thì dùng bản trong bộ nhớ
  }
  if (!s || !s.key || !s.name) {
    return null;
  }
  if (Date.now() / 1000 - s.at > REMEMBER_SECS) {
    clearSession();
    return null;
  }
  return s;
}

/** clearSession xoá mã — gọi khi ván kết thúc (chết, bấm thoát, bị tab khác chiếm). */
export function clearSession(): void {
  mem = null;
  try {
    sessionStorage.removeItem(KEY);
  } catch {
    // Không xoá được thì lần tải sau gửi một mã đã chết; server cho vào trận mới. Không sao.
  }
}
