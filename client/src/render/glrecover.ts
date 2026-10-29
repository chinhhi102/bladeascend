import type { Profile } from '../ui/home';

/**
 * Khôi phục khi mất ngữ cảnh WebGL.
 *
 * Trình duyệt thu hồi ngữ cảnh WebGL khi trình điều khiển card màn hình bị reset, khi máy
 * ngủ dậy, hoặc khi hệ điều hành thấy một tab ngốn quá nhiều bộ nhớ GPU. Đo thật trên máy
 * dev (AMD RX 570, Chrome headless): cứ vài lượt chạy lại dính một lần.
 *
 * Trước đây client KHÔNG bắt sự kiện này: canvas đen vĩnh viễn, người chơi tưởng game treo
 * và chẳng có cách nào ngoài tự tải lại trang.
 *
 * Vì sao TẢI LẠI TRANG chứ không dựng lại renderer: mọi texture, atlas, bộ đệm hình học và
 * shader đều nằm trên GPU và mất sạch theo ngữ cảnh. Dựng lại từng thứ một là hàng trăm
 * dòng mã chỉ chạy đúng trong một tình huống hiếm — tức là loại mã không bao giờ được thử
 * và sẽ hỏng lặng lẽ. Tải lại trang thì chắc chắn đúng, và vì ta nhớ lại tên + nhân vật nên
 * người chơi tự vào lại phòng, chỉ thấy một nhịp gián đoạn.
 */

const KEY = 'kx.glrecover';
/**
 * Mốc thời gian lần mất ngữ cảnh gần nhất. Phải là khoá RIÊNG, vì `pendingRejoin` xoá KEY
 * ngay khi trang tải lại — dùng chung một khoá thì lần mất thứ hai lại tưởng là lần đầu và
 * trang cứ tải lại vòng tròn trên máy có card hỏng thật.
 */
const LAST_KEY = 'kx.glrecover.last';
/** Nhớ lâu hơn ngần này thì coi như lần chơi cũ, không tự vào lại nữa. */
const REMEMBER_SECS = 60;
/** Hai lần mất ngữ cảnh gần nhau hơn ngần này = máy đang hỏng thật, đừng tải lại vòng lặp. */
const LOOP_GUARD_SECS = 30;

interface Saved {
  name: string;
  char: string;
  at: number;
}

function read(): Saved | null {
  try {
    const raw = sessionStorage.getItem(KEY);
    return raw ? (JSON.parse(raw) as Saved) : null;
  } catch {
    return null;
  }
}

/**
 * pendingRejoin trả về hồ sơ cần vào lại sau khi tải lại trang vì mất đồ hoạ, rồi XOÁ dấu.
 * Xoá ngay cả khi đã quá hạn, để lần sau mở game không dính lại vết cũ.
 */
export function pendingRejoin(): Profile | null {
  const s = read();
  try {
    sessionStorage.removeItem(KEY);
  } catch {
    /* chế độ ẩn danh chặn sessionStorage thì thôi */
  }
  if (!s || Date.now() / 1000 - s.at > REMEMBER_SECS) {
    return null;
  }
  return { name: s.name, char: s.char };
}

/**
 * watchContextLoss bắt sự kiện mất ngữ cảnh trên canvas.
 *
 * `profile` là hàm chứ không phải giá trị: lúc gắn sự kiện thì người chơi còn ở trang chủ,
 * chưa chọn nhân vật.
 */
export function watchContextLoss(
  canvas: HTMLCanvasElement,
  profile: () => Profile | null,
  onLost: (canRecover: boolean) => void,
): void {
  canvas.addEventListener('webglcontextlost', (e) => {
    // Bắt buộc: không chặn sự kiện mặc định thì trình duyệt không bao giờ cấp lại ngữ cảnh,
    // và cũng không cho ta kịp làm gì.
    e.preventDefault();

    const now = Date.now() / 1000;
    const last = Number(sessionStorage.getItem(LAST_KEY) ?? 0);
    const looping = last > 0 && now - last < LOOP_GUARD_SECS;
    try {
      sessionStorage.setItem(LAST_KEY, String(now));
    } catch {
      /* không lưu được thì mất chốt chống lặp, vẫn hơn là không khôi phục gì */
    }
    onLost(!looping);
    if (looping) {
      return; // vừa tải lại xong lại mất tiếp: máy hỏng thật, đừng quay vòng
    }

    const p = profile();
    try {
      sessionStorage.setItem(KEY, JSON.stringify({ name: p?.name ?? '', char: p?.char ?? '', at: now }));
    } catch {
      /* không lưu được thì vẫn tải lại, chỉ mất phần tự vào lại phòng */
    }
    // Chờ một nhịp cho người chơi kịp đọc dòng thông báo rồi mới tải lại.
    window.setTimeout(() => location.reload(), 1200);
  });
}
