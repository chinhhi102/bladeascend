// Định dạng số, thời gian và nhãn tiếng Việt dùng chung cho trang quản trị.

export const f1 = (n: number): string => (Math.round(n * 10) / 10).toFixed(1);

export const int = (n: number): string => Math.round(n).toLocaleString('vi-VN');

export function uptime(secs: number): string {
  const h = Math.floor(secs / 3600);
  const m = Math.floor((secs % 3600) / 60);
  return h > 0 ? `${h}h${String(m).padStart(2, '0')}` : `${m}m`;
}

/** duration: "45 giây" / "2 phút" / "1 giờ 3 phút". */
export function duration(secs: number): string {
  if (secs < 60) {
    return `${Math.round(secs)} giây`;
  }
  const m = Math.floor(secs / 60);
  if (m < 60) {
    return `${m} phút`;
  }
  return `${Math.floor(m / 60)} giờ ${m % 60} phút`;
}

export function fullTime(unixSecs: number): string {
  return new Date(unixSecs * 1000).toLocaleString('vi-VN');
}

/** relTime: "3 phút trước" / "sau 2 giờ" — ADMIN_SPEC §2.1 "thời gian hiển thị tương đối". */
export function relTime(unixSecs: number, nowMs = Date.now()): string {
  const d = Math.round(nowMs / 1000 - unixSecs);
  const abs = Math.abs(d);
  const steps: [number, number, string][] = [[60, 1, 'giây'], [3600, 60, 'phút'], [86400, 3600, 'giờ'], [86400 * 30, 86400, 'ngày']];
  for (const [lim, div, unit] of steps) {
    if (abs < lim) {
      const n = Math.max(1, Math.floor(abs / div));
      return d >= 0 ? `${n} ${unit} trước` : `sau ${n} ${unit}`;
    }
  }
  return new Date(unixSecs * 1000).toLocaleDateString('vi-VN');
}

/**
 * REALMS phải khớp `cultivation.realms` + `tien_realms` trong config. Lệch thì chỉ sai NHÃN
 * trên trang quản trị, không đụng luật chơi.
 */
const REALMS = ['Luyện Khí', 'Trúc Cơ', 'Kim Đan', 'Nguyên Anh', 'Hóa Thần', 'Luyện Hư',
  'Hợp Thể', 'Độ Kiếp', 'Đại Thừa', 'Chân Tiên', 'Kim Tiên', 'Thái Ất Ngọc Tiên',
  'Đại La Kim Tiên', 'Đạo Tổ'];

export const realmName = (r: number): string => REALMS[r] ?? `cảnh giới ${r}`;

export const FB_KINDS: Record<string, string> = { bug: 'Báo lỗi', balance: 'Cân bằng', idea: 'Ý tưởng', other: 'Khác' };

export const FB_STATUS: Record<string, string> = { moi: 'Mới', da_xem: 'Đã xem', dang_lam: 'Đang làm', xong: 'Xong', bo_qua: 'Bỏ qua' };

export const FB_STATUS_COLOR: Record<string, string> = { moi: 'blue', da_xem: 'default', dang_lam: 'gold', xong: 'green', bo_qua: 'default' };

export const ERR_LABEL: Record<string, string> = {
  error: 'Lỗi JavaScript',
  webgl_context_lost: 'Mất ngữ cảnh WebGL (màn hình trắng)',
  gpu_pressure: 'Chạm trần bộ nhớ đồ hoạ',
};

/** BAN_SPANS: thời hạn §6.2. 0 = chỉ mời ra, -1 = vĩnh viễn. */
export const BAN_SPANS: { secs: number; label: string }[] = [
  { secs: 0, label: 'Chỉ mời ra (vào lại được ngay)' },
  { secs: 3600, label: 'Cấm 1 giờ' },
  { secs: 86400, label: 'Cấm 24 giờ' },
  { secs: 604800, label: 'Cấm 7 ngày' },
  { secs: -1, label: 'Cấm vĩnh viễn' },
];

export const BOT_MODES: { value: string; label: string }[] = [
  { value: 'auto', label: 'Tự động' },
  { value: 'on', label: 'Bật' },
  { value: 'off', label: 'Tắt' },
];

export const BOT_LABEL: Record<string, string> = { auto: 'Tự động', on: 'Bật', off: 'Tắt', '': 'Theo toàn server' };

/** Mức màu theo ngưỡng (ADMIN_SPEC §3): xanh / vàng / đỏ. */
export type Level = 'ok' | 'warn' | 'bad';

export const LEVEL_COLOR: Record<Level, string> = { ok: '#52c41a', warn: '#faad14', bad: '#ff4d4f' };

/**
 * tickLevel — §3 "tô đỏ nếu > 4 ms". Config chưa có khối `admin.thresholds` nên vàng giữ mốc
 * 2,5 ms như bản trang cũ.
 */
export function tickLevel(ms: number): Level {
  if (ms > 4) {
    return 'bad';
  }
  return ms > 2.5 ? 'warn' : 'ok';
}

/** countLevel: thứ nên bằng 0 (cờ, lỗi): > 0 là vàng. */
export const countLevel = (n: number): Level => (n > 0 ? 'warn' : 'ok');

/** includesCI: tìm không phân biệt hoa thường. */
export function includesCI(hay: string | undefined, needle: string): boolean {
  return (hay ?? '').toLowerCase().includes(needle.toLowerCase());
}
