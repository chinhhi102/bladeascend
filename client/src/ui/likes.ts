import { guestId } from './home';
import { h } from './dom';
import { STR } from './strings';

/**
 * Lượt thích nhân vật — GAME_SPEC §6.6 (v6.48).
 *
 * Mỗi người một lượt thích cho một nhân vật; bấm nhân vật khác thì lượt chuyển sang, bấm
 * lại thì bỏ. Chỉ người đã chơi xong một mạng ≥ 60 giây mới thích được.
 *
 * Lượt thích KHÔNG ảnh hưởng luật chơi — nó chỉ nói cho chủ dự án biết nhân vật nào được
 * ưa. Nên mọi lỗi ở đây đều "hỏng êm": mất mạng, server chưa có bảng, chưa đủ điều kiện…
 * đều chỉ làm cái tim không hiện hoặc không bấm được, không bao giờ chặn ai vào trận.
 */

export interface LikeState {
  counts: Record<string, number>;
  top: string;
  mine: string;
  canLike: boolean;
}

const EMPTY: LikeState = { counts: {}, top: '', mine: '', canLike: false };

let state: LikeState = EMPTY;
const listeners = new Set<() => void>();

/** likeState: bảng đếm đang giữ (rỗng khi chưa tải xong). */
export function likeState(): LikeState {
  return state;
}

/** onLikes đăng ký vẽ lại khi bảng đếm đổi; trả hàm huỷ đăng ký. */
export function onLikes(fn: () => void): () => void {
  listeners.add(fn);
  return () => listeners.delete(fn);
}

function apply(raw: unknown): void {
  const d = raw as { counts?: Record<string, number>; top?: string; mine?: string; can_like?: boolean };
  state = { counts: d.counts ?? {}, top: d.top ?? '', mine: d.mine ?? '', canLike: !!d.can_like };
  for (const fn of listeners) {
    fn();
  }
}

/** loadLikes đọc bảng đếm. Hỏng thì im lặng giữ bảng cũ. */
export async function loadLikes(): Promise<void> {
  try {
    const r = await fetch(`/api/likes?guest=${encodeURIComponent(guestId())}`);
    if (r.ok) {
      apply(await r.json());
    }
  } catch {
    // mất mạng: cái tim đứng im, không có gì vỡ
  }
}

/**
 * toggleLike bật/tắt lượt thích cho `char`. Trả về câu cần nói với người chơi ('' = xong
 * êm) — gọi xong nhớ vẽ lại, `onLikes` đã tự bắn.
 */
export async function toggleLike(char: string): Promise<string> {
  const want = state.mine === char ? '' : char;
  try {
    const r = await fetch('/api/likes', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ guest: guestId(), character: want }),
    });
    if (r.ok) {
      apply(await r.json());
      return '';
    }
    const err = (await r.json().catch(() => ({}))) as { error?: string };
    return err.error === 'ip_limit' ? STR.likeIPLimit : STR.likeNeedPlay;
  } catch {
    return STR.likeFailed;
  }
}

/** likeCount: "1,2k" khi quá nghìn (§6.6), số thường khi ít. */
export function likeCount(n: number): string {
  if (n < 1000) {
    return String(n);
  }
  return (n / 1000).toLocaleString(undefined, { maximumFractionDigits: 1 }) + 'k';
}

/**
 * likeButton: nút ❤ dùng chung cho khung chi tiết nhân vật và màn Tổng kết.
 *
 * Tự vẽ lại theo `onLikes`, nên chỗ gọi không phải nhớ cập nhật. Hàm huỷ gắn vào chính phần
 * tử (`el.remove()` không gọi hàm huỷ, nhưng màn hình bị thay thì listener cũ chỉ còn gọi
 * một hàm vô hại — rẻ hơn việc bắt mọi chỗ gọi phải dọn).
 */
export function likeButton(char: () => string, onMsg: (msg: string) => void = () => {}): HTMLElement {
  const btn = h('button', 'like-btn');
  const heart = h('span', 'like-heart', '♥');
  const count = h('span', 'like-count');
  btn.append(heart, count);
  const draw = (): void => {
    const c = char();
    const s = likeState();
    const mine = s.mine === c && c !== '';
    btn.classList.toggle('on', mine);
    btn.classList.toggle('off', !s.canLike);
    btn.title = !s.canLike ? STR.likeNeedPlay : mine ? STR.likedTitle : STR.likeTitle;
    count.textContent = likeCount(s.counts[c] ?? 0);
  };
  btn.addEventListener('click', async (e) => {
    e.stopPropagation();
    onMsg(await toggleLike(char()));
  });
  onLikes(draw);
  draw();
  return btn;
}

/** likeBadge: nhãn "Được thích nhất" cho nhân vật đứng đầu; ẩn khi chưa ai thích. */
export function likeBadge(char: () => string): HTMLElement {
  const el = h('span', 'like-top', STR.likeTop);
  const draw = (): void => {
    const s = likeState();
    el.style.display = s.top !== '' && s.top === char() ? '' : 'none';
  };
  onLikes(draw);
  draw();
  return el;
}
