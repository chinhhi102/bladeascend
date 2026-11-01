import type { DeadMsg } from '../net/protocol';
import type { Profile } from './home';

/**
 * Chết = hết ván (spec v6.29 §7): ván đã kết thúc thì TẢI LẠI TRANG vẫn phải hiện lại màn
 * Tổng kết, không tự vào trận. Nhớ gói `dead` trong sessionStorage — sống theo tab, mất khi
 * đóng tab — tới lúc người chơi tự bấm "Chơi tiếp" hoặc "Đổi nhân vật".
 */

const KEY = 'kx.runended';

export interface EndedRun {
  dead: DeadMsg;
  profile: Profile;
}

export function saveEndedRun(dead: DeadMsg, profile: Profile): void {
  try {
    sessionStorage.setItem(KEY, JSON.stringify({ dead, profile }));
  } catch {
    // Ẩn danh chặn sessionStorage: tải lại sẽ về trang chủ — vẫn không tự vào trận.
  }
}

export function endedRun(): EndedRun | null {
  try {
    const raw = sessionStorage.getItem(KEY);
    return raw ? (JSON.parse(raw) as EndedRun) : null;
  } catch {
    return null;
  }
}

export function clearEndedRun(): void {
  try {
    sessionStorage.removeItem(KEY);
  } catch {
    // không xoá được thì lần tải sau hiện lại Tổng kết — người chơi bấm "Chơi tiếp" là xong
  }
}
