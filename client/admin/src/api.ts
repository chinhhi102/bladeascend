// Gọi API quản trị — ADMIN_SPEC §1, §9.
//
// Mọi request đổi dữ liệu (khác GET) mang `X-Admin-Token` = CSRF token lấy từ
// `GET /api/admin/me`. 401 → về trang đăng nhập (phiên hết hạn hoặc đã đăng xuất).

let csrf = '';

export function setCsrf(token: string): void {
  csrf = token;
}

export class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message);
  }
}

interface CallOptions {
  method?: 'GET' | 'POST' | 'PATCH' | 'DELETE';
  body?: unknown;
}

/** errorText rút câu lỗi server trả về (`{"error": "..."}`), không có thì "HTTP 500". */
async function errorText(r: Response): Promise<string> {
  try {
    const j = (await r.json()) as { error?: string };
    return j.error ?? `HTTP ${r.status}`;
  } catch {
    return `HTTP ${r.status}`;
  }
}

export async function api<T>(path: string, opts: CallOptions = {}): Promise<T> {
  const method = opts.method ?? 'GET';
  const headers: Record<string, string> = {};
  if (method !== 'GET') {
    headers['X-Admin-Token'] = csrf;
  }
  if (opts.body !== undefined) {
    headers['Content-Type'] = 'application/json';
  }
  const r = await fetch(path, {
    method,
    credentials: 'same-origin',
    headers,
    body: opts.body === undefined ? undefined : JSON.stringify(opts.body),
  });
  if (r.status === 401) {
    location.href = '/admin/login';
    throw new ApiError('chưa đăng nhập', 401);
  }
  if (!r.ok) {
    throw new ApiError(await errorText(r), r.status);
  }
  return (await r.json()) as T;
}

export function post<T>(path: string, body: unknown): Promise<T> {
  return api<T>(path, { method: 'POST', body });
}

/** errMsg: chữ hiển thị cho một lỗi bất kỳ bắt được. */
export function errMsg(e: unknown): string {
  return e instanceof Error ? e.message : String(e);
}
