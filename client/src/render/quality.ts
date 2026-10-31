import { gpuBudgetMB } from './device';

/**
 * Mức chất lượng đồ họa (spec §3.6): Cao / Vừa / Thấp. Người chơi đổi bằng phím V, lựa chọn
 * lưu trong localStorage.
 *
 * Thần thú giờ luôn bật (spec v4.7 §3.5) nên số đốt phải co theo máy, không thể để cố định.
 */

export type Quality = 'high' | 'medium' | 'low';

const KEY = 'kx.quality';
const ORDER: Quality[] = ['high', 'medium', 'low'];

export const QUALITY_NAME: Record<Quality, string> = { high: 'Cao', medium: 'Vừa', low: 'Thấp' };

/** Số đốt thần thú theo chất lượng (spec §3.5: Cao 16 · Vừa 10 · Thấp 6). */
export const BEAST_SEGMENTS: Record<Quality, number> = { high: 16, medium: 10, low: 6 };

/** Tỉ lệ hạt vệt theo sau (Thấp = tắt hẳn). */
export const TRAIL_RATE: Record<Quality, number> = { high: 1, medium: 0.5, low: 0 };

let current: Quality = read();

function read(): Quality {
  const v = localStorage.getItem(KEY);
  return ORDER.includes(v as Quality) ? (v as Quality) : 'high';
}

export function quality(): Quality {
  return current;
}

export function setQuality(q: Quality): void {
  current = q;
  localStorage.setItem(KEY, q);
}

/** cycleQuality đổi sang mức tiếp theo và trả về mức mới. Người chơi tự chọn thì khoá lại. */
export function cycleQuality(): Quality {
  manual = true;
  setQuality(ORDER[(ORDER.indexOf(current) + 1) % ORDER.length]);
  return current;
}

/**
 * Tự hạ chất lượng (spec §13.6 mục 8): FPS dưới 45 suốt 3 giây thì hạ một bậc; ổn lại trên
 * 55 suốt 10 giây thì nâng lại một bậc.
 *
 * Không đụng tới lựa chọn của người chơi: ai đã tự bấm phím V thì giữ nguyên mức họ chọn —
 * máy tự ý đổi ngược ý người dùng là thứ gây ức chế nhất.
 */
let manual = false;
let lowSince = 0;
let okSince = 0;

/**
 * gpuPressure: bộ nhớ texture đã vượt ngân sách thiết bị chưa (NETCODE_SPEC §5 L1).
 *
 * Trả về mức 0 (ổn) · 1 (sát trần, hạ một bậc) · 2 (vượt trần, xuống Thấp ngay).
 */
export function gpuPressure(texMB: number): 0 | 1 | 2 {
  const budget = gpuBudgetMB();
  if (texMB >= budget) {
    return 2;
  }
  return texMB >= budget * 0.8 ? 1 : 0;
}

/**
 * autoQualityByMemory hạ chất lượng khi bộ nhớ GPU chạm ngân sách — **kể cả khi FPS vẫn đẹp**.
 *
 * Đây là điểm khác hẳn `autoQuality`: trên iOS, hết bộ nhớ đồ hoạ KHÔNG làm tụt khung hình
 * trước. Máy chạy 60 fps tới đúng khoảnh khắc Safari giết tab. Nên nhìn FPS là nhìn sai chỗ;
 * phải nhìn chính con số bộ nhớ.
 *
 * Và hạ chất lượng ở đây **bỏ qua cả lựa chọn tay của người chơi** (`manual`): họ chọn mức
 * nào là quyền của họ, nhưng không ai chọn "cho tab chết".
 */
export function autoQualityByMemory(texMB: number): boolean {
  const level = gpuPressure(texMB);
  if (level === 0) {
    return false;
  }
  const want: Quality = level === 2 ? 'low' : ORDER[Math.min(ORDER.indexOf(current) + 1, ORDER.length - 1)];
  if (want === current) {
    return false;
  }
  setQuality(want);
  return true;
}

export function autoQuality(fps: number, now: number): void {
  if (manual || fps <= 0) {
    return;
  }
  if (fps < 45) {
    okSince = 0;
    lowSince = lowSince || now;
    if (now - lowSince >= 3 && current !== 'low') {
      setQuality(ORDER[ORDER.indexOf(current) + 1]);
      lowSince = now;
    }
    return;
  }
  lowSince = 0;
  if (fps > 55) {
    okSince = okSince || now;
    if (now - okSince >= 10 && current !== 'high') {
      setQuality(ORDER[ORDER.indexOf(current) - 1]);
      okSince = now;
    }
  } else {
    okSince = 0;
  }
}
