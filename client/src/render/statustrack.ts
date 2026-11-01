import { STATUS_EV, type StatusEv } from '../net/protocol';

/**
 * StatusTracker giữ đồng hồ trạng thái của từng người theo event server (spec v6.36 §6.5):
 * `StatusStart {a, b, n, f}` / `StatusEnd {a, n}`.
 *
 * Bit `st` trong snapshot chỉ nói "đang bị"; số giây và người gây ra chỉ có ở event — vòng
 * đếm ngược, HUD và mũi tên đọc từ đây. Client KHÔNG tự tính thời hạn: hết event thì hết.
 */

export interface StatusTimer {
  until: number;
  total: number;
  /** src: id người gây; 0 = tự thân. */
  src: number;
}

/**
 * MIN_SHOW: mã sống quá ngắn để kịp nhìn thì vẫn HIỆN đủ ngần này giây. Bị Loa đẩy chỉ kéo
 * dài 0,3 s (`push_secs`) — icon 📢 loé 0,3 s thì không ai kịp thấy mình vừa bị gì. Chỉ đổi
 * thời gian HIỆN, không đổi luật: StatusEnd của mã này bị bỏ qua, đồng hồ tự hết.
 */
const MIN_SHOW: Partial<Record<StatusEv, number>> = { push: 1 };

/** Quá hạn ngần này giây mà chưa có StatusEnd (event lạc khi người đó ra khỏi tầm nhìn) thì tự dọn. */
const STALE_SECS = 1;

export class StatusTracker {
  private readonly byId = new Map<number, Map<StatusEv, StatusTimer>>();

  /** code: chỉ số trong `STATUS_EV`; mã lạ (server mới hơn client) thì bỏ qua. */
  start(id: number, code: number, secs: number, src: number, now: number): StatusEv | null {
    const name = STATUS_EV[code];
    if (!name || secs <= 0) {
      return null;
    }
    let m = this.byId.get(id);
    if (!m) {
      m = new Map();
      this.byId.set(id, m);
    }
    const show = Math.max(secs, MIN_SHOW[name] ?? 0);
    m.set(name, { until: now + show, total: show, src });
    return name;
  }

  end(id: number, code: number): StatusEv | null {
    const name = STATUS_EV[code];
    const m = this.byId.get(id);
    if (name && MIN_SHOW[name] !== undefined) {
      return name; // tự hết theo MIN_SHOW
    }
    if (!name || !m?.delete(name)) {
      return name ?? null;
    }
    if (m.size === 0) {
      this.byId.delete(id);
    }
    return name;
  }

  /** has: người này còn đồng hồ nào không (đường tắt cho người không bị gì — đa số). */
  has(id: number): boolean {
    return this.byId.has(id);
  }

  get(id: number, code: StatusEv): StatusTimer | undefined {
    return this.byId.get(id)?.get(code);
  }

  /** left: giây còn lại và tỉ lệ còn lại (0..1); null nếu không có đồng hồ. */
  left(id: number, code: StatusEv, now: number): { secs: number; frac: number } | null {
    const t = this.get(id, code);
    if (!t) {
      return null;
    }
    const secs = Math.max(0, t.until - now);
    return { secs, frac: t.total > 0 ? secs / t.total : 0 };
  }

  /** active: còn đồng hồ và chưa hết giờ (cho các mã không có bit trên dây: push, steal, minion). */
  active(id: number, code: StatusEv, now: number): boolean {
    const t = this.get(id, code);
    return t !== undefined && t.until > now;
  }

  prune(now: number): void {
    for (const [id, m] of this.byId) {
      for (const [k, t] of m) {
        if (t.until + STALE_SECS < now) {
          m.delete(k);
        }
      }
      if (m.size === 0) {
        this.byId.delete(id);
      }
    }
  }

  clear(): void {
    this.byId.clear();
  }
}

/** statusTracker: một bảng cho cả client — scene ghi theo event, PlayerSprite/HUD đọc. */
export const statusTracker = new StatusTracker();
