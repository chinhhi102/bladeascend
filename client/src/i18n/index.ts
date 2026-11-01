// Đa ngôn ngữ (GAME_SPEC §19). Mọi chữ người chơi thấy nằm trong `i18n/{lang}.json`;
// `ui/strings.ts` dựng `STR` từ từ điển đã chọn ở đây.
//
// Chọn ngôn ngữ: `?lang=xx` trên URL > lựa chọn đã lưu > `navigator.language` > Tiếng Việt.
// Thiếu khoá ở ngôn ngữ nào thì rơi về English, rồi Tiếng Việt (spec §19).
//
// Nạp bằng top-level await NGAY KHI module này được import: mọi module dùng `STR` đều chờ
// xong rồi mới chạy, nên không chỗ nào kịp đọc nhầm chữ tiếng Việt. Chỉ tải file của ngôn
// ngữ đang chọn (+ English làm dự phòng) — không nhét cả 6 ngôn ngữ vào gói khởi động.
import vi from './vi.json';

export type Dict = typeof vi;

export interface LangInfo {
  code: string;
  name: string;
  flag: string;
  /** Thư mục trang SEO tĩnh của ngôn ngữ (spec §19 "SEO đa ngôn ngữ"). */
  path: string;
}

/** Đợt 1 (spec §19). Thứ tự = thứ tự nút cờ trên trang chủ. */
export const LANGS: LangInfo[] = [
  { code: 'vi', name: 'Tiếng Việt', flag: '🇻🇳', path: '/' },
  { code: 'en', name: 'English', flag: '🇬🇧', path: '/en/' },
  { code: 'pt-BR', name: 'Português', flag: '🇧🇷', path: '/pt/' },
  { code: 'es', name: 'Español', flag: '🇪🇸', path: '/es/' },
  { code: 'id', name: 'Bahasa Indonesia', flag: '🇮🇩', path: '/id/' },
  { code: 'th', name: 'ไทย', flag: '🇹🇭', path: '/th/' },
];

const STORE_KEY = 'kx_lang';

type Loader = () => Promise<{ default: Partial<Dict> }>;
// `import.meta.glob` để Vite tách mỗi ngôn ngữ thành một chunk riêng, tải khi cần.
const loaders = import.meta.glob<{ default: Partial<Dict> }>(['./*.json', '!./vi.json']) as Record<string, Loader>;

/** Đổi mã trình duyệt ("pt-PT", "en-US", "th") về một mã đợt 1, hoặc '' nếu không hỗ trợ. */
export function matchLang(tag: string | null | undefined): string {
  if (!tag) {
    return '';
  }
  const t = tag.toLowerCase();
  for (const l of LANGS) {
    if (t === l.code.toLowerCase()) {
      return l.code;
    }
  }
  const base = t.split('-')[0];
  if (base === 'pt') {
    return 'pt-BR'; // đợt 1 chỉ có bản Brasil; người Bồ đọc vẫn hiểu
  }
  return LANGS.find((l) => l.code.split('-')[0] === base)?.code ?? '';
}

function readStore(): string {
  try {
    return localStorage.getItem(STORE_KEY) ?? '';
  } catch {
    return '';
  }
}

/** Ngôn ngữ người chơi chọn tay (không phải tự động), '' nếu đang để tự động. */
export function savedLang(): string {
  return matchLang(readStore());
}

function detect(): string {
  const url = typeof location === 'undefined' ? null : new URLSearchParams(location.search).get('lang');
  const fromPath = typeof location === 'undefined' ? '' : LANGS.find((l) => l.path !== '/' && location.pathname.startsWith(l.path))?.code;
  const nav = typeof navigator === 'undefined' ? [] : (navigator.languages ?? [navigator.language]);
  for (const c of [url, readStore(), fromPath, ...nav]) {
    const m = matchLang(c);
    if (m) {
      return m;
    }
  }
  return 'vi';
}

/** Lưu lựa chọn rồi tải lại trang — đổi ngôn ngữ giữa chừng thì chữ đã dựng sẽ lẫn hai thứ tiếng. */
export function chooseLang(code: string): void {
  try {
    if (code) {
      localStorage.setItem(STORE_KEY, code);
    } else {
      localStorage.removeItem(STORE_KEY);
    }
  } catch {
    // Không lưu được (chế độ riêng tư): vẫn đổi qua URL.
  }
  const u = new URL(location.href);
  u.searchParams.delete('lang');
  if (code) {
    u.searchParams.set('lang', code);
  }
  location.href = u.toString();
}

function isObj(v: unknown): v is Record<string, unknown> {
  return typeof v === 'object' && v !== null && !Array.isArray(v);
}

/** Gộp sâu: `over` đè `base`; chuỗi rỗng và mảng rỗng ở `over` coi như chưa dịch. */
function merge<T>(base: T, over: unknown): T {
  if (!isObj(base) || !isObj(over)) {
    return base;
  }
  const out: Record<string, unknown> = { ...base };
  for (const [k, v] of Object.entries(over)) {
    const b = (base as Record<string, unknown>)[k];
    if (v === '' || v === null || (Array.isArray(v) && v.length === 0)) {
      continue;
    }
    out[k] = isObj(b) && isObj(v) ? merge(b, v) : v;
  }
  return out as T;
}

async function load(code: string): Promise<Partial<Dict>> {
  const f = loaders[`./${code}.json`];
  if (!f) {
    return {};
  }
  try {
    return (await f()).default;
  } catch {
    return {}; // mạng hỏng giữa chừng: vẫn chơi được bằng tiếng Việt
  }
}

export const LANG = detect();

async function build(): Promise<Dict> {
  if (LANG === 'vi') {
    return vi;
  }
  const [en, own] = await Promise.all([load('en'), LANG === 'en' ? Promise.resolve({}) : load(LANG)]);
  return merge(merge(vi, en), own);
}

export const DICT: Dict = await build();

/** Locale cho số và ngày (dấu phân cách nghìn đúng kiểu từng nước). */
export const LOCALE = (DICT._meta?.locale as string | undefined) ?? 'vi-VN';

if (typeof document !== 'undefined') {
  document.documentElement.lang = LANG;
}

/** Điền `{tên}` trong mẫu câu. Tham số thiếu thì để nguyên `{tên}` cho dễ thấy lỗi. */
export function fmt(tpl: string, p: Record<string, string | number> = {}): string {
  return tpl.replace(/\{(\w+)\}/g, (m, k: string) => (k in p ? String(p[k]) : m));
}

/** Bốc ngẫu nhiên một mẫu trong danh sách rồi điền tham số (kill feed nhiều cách nói). */
export function pickFmt(list: string[] | string, p: Record<string, string | number> = {}): string {
  const arr = Array.isArray(list) ? list : [list];
  return fmt(arr[Math.floor(Math.random() * arr.length)] ?? '', p);
}

/** Số có phân cách nghìn theo ngôn ngữ đang chọn. */
export function num(n: number): string {
  return Math.round(n).toLocaleString(LOCALE);
}
