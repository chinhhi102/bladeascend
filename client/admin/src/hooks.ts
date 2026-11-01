import { createContext, useCallback, useContext, useEffect, useRef, useState } from 'react';

import { api } from './api';
import type { Stats } from './types';

export interface Loaded<T> {
  data: T | null;
  error: string;
  loading: boolean;
  reload: () => void;
}

/**
 * useFetch tải `path`, tải lại khi `path` đổi, và mỗi `everyMs` nếu có (0 = không tự làm mới).
 * Gọi lại trong lúc còn đang chờ thì bỏ qua nhịp đó — không chồng request.
 */
export function useFetch<T>(path: string | null, everyMs = 0): Loaded<T> {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const busy = useRef(false);
  const alive = useRef(true);

  const load = useCallback(async () => {
    if (!path || busy.current) {
      return;
    }
    busy.current = true;
    setLoading(true);
    try {
      const d = await api<T>(path);
      if (alive.current) {
        setData(d);
        setError('');
      }
    } catch (e) {
      if (alive.current) {
        setError(e instanceof Error ? e.message : String(e));
      }
    } finally {
      busy.current = false;
      if (alive.current) {
        setLoading(false);
      }
    }
  }, [path]);

  useEffect(() => {
    alive.current = true;
    void load();
    const t = everyMs > 0 ? window.setInterval(() => void load(), everyMs) : 0;
    return () => {
      alive.current = false;
      window.clearInterval(t);
    };
  }, [load, everyMs]);

  return { data, error, loading, reload: () => void load() };
}

export interface StatsState {
  now: Stats | null;
  /** ago: gói của ~1 phút trước, để vẽ mũi tên tăng/giảm (§2.1). */
  ago: Stats | null;
  error: string;
}

export const StatsContext = createContext<StatsState>({ now: null, ago: null, error: '' });

export const useStats = (): StatsState => useContext(StatsContext);

/** HISTORY: 31 nhịp × 2 s ≈ 1 phút. */
const HISTORY = 31;

/** useStatsPoll: §3 — dashboard cập nhật mỗi 2 giây. Một nguồn cho cả trang. */
export function useStatsPoll(): StatsState {
  const { data, error } = useFetch<Stats>('/api/admin/stats', 2000);
  const hist = useRef<Stats[]>([]);
  if (data && hist.current[hist.current.length - 1] !== data) {
    hist.current = [...hist.current, data].slice(-HISTORY);
  }
  return { now: data, ago: hist.current.length > 1 ? hist.current[0] : null, error };
}

export interface Route {
  page: string;
  args: string[];
}

function parseHash(): Route {
  const [page, ...args] = location.hash.replace(/^#/, '').split('/');
  return { page: page || 'dashboard', args: args.map(decodeURIComponent) };
}

/** useHashRoute: điều hướng bằng `#trang/tham/số` — giữ được liên kết cũ `#feedback`, `#board`… */
export function useHashRoute(): Route {
  const [r, setR] = useState(parseHash);
  useEffect(() => {
    const on = (): void => setR(parseHash());
    window.addEventListener('hashchange', on);
    return () => window.removeEventListener('hashchange', on);
  }, []);
  return r;
}

export function go(page: string, ...args: (string | number)[]): void {
  location.hash = [page, ...args.map((a) => encodeURIComponent(String(a)))].join('/');
}
