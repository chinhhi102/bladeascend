/**
 * Chống sập phía client — GAME_SPEC §13.8.
 *
 * ── Vấn đề ───────────────────────────────────────────────────────────────────────────────
 *
 * Vòng vẽ là một hàm dài gọi mười mấy hệ thống nối đuôi nhau. Một `undefined` trong lớp vẽ
 * bảng tên ném lỗi ra khỏi cả `render()`, nên bản đồ, nhân vật, HUD và minimap của khung đó
 * cũng không được vẽ. Ticker của Pixi gọi lại ở khung sau, lại ném ở đúng chỗ đó — màn hình
 * đứng hình vĩnh viễn vì MỘT hệ thống hỏng. Người chơi thấy "game treo", không thấy "thiếu
 * bảng tên".
 *
 * ── Cách làm ─────────────────────────────────────────────────────────────────────────────
 *
 * Mỗi hệ thống là một **ranh giới**: `sysRun('vfx', …)`. Lỗi thì
 *
 *  1. tắt riêng hệ thống đó **2 giây** rồi tự thử lại — phần lớn lỗi vẽ là do một trạng thái
 *     nhất thời (texture chưa nạp xong, entity vừa bị xoá), hai giây sau là qua;
 *  2. lỗi **≥ 3 lần trong 10 giây** ở cùng hệ thống → tắt hẳn **tới hết mạng**. Thử lại mãi
 *     một thứ hỏng thật chỉ tốn CPU và làm ngập telemetry.
 *
 * Game vẫn chơi được, chỉ thiếu đúng hiệu ứng đó. Và vì mọi lần tắt đều gửi telemetry kèm
 * bối cảnh, ta BIẾT ai đang chơi thiếu thứ gì — khác hẳn `try {} catch {}` nuốt im lặng.
 *
 * Module này KHÔNG phụ thuộc Pixi hay DOM: nó là sổ trạng thái thuần, nên test được bằng
 * vitest mà không cần dựng renderer.
 */

/**
 * Khoá hệ thống. Danh sách bám theo §13.8 ("nội suy, vòng kiếm, bảng tên, thần thú, VFX,
 * emote, minimap, HUD, âm thanh") cộng những ranh giới có thật trong `scene.render`.
 *
 * ASCII, không dấu: `check:i18n` chặn chuỗi tiếng Việt ngoài `ui/strings.ts` (quy tắc 5), và
 * mấy khoá này còn đi thẳng vào telemetry nên càng không được dịch.
 */
export type SysKey =
  | 'interp' | 'players' | 'monsters' | 'items' | 'beasts'
  | 'ambient' | 'map' | 'vfx' | 'skills' | 'emote'
  | 'minimap' | 'hud' | 'audio' | 'telemetry';

/** SYS_KEYS: thứ tự hiện trên bảng F3. */
export const SYS_KEYS: SysKey[] = [
  'interp', 'players', 'monsters', 'items', 'beasts',
  'ambient', 'map', 'vfx', 'skills', 'emote',
  'minimap', 'hud', 'audio', 'telemetry',
];

/** Tắt tạm bao lâu sau một lỗi (§13.8: 2 s). */
export const SYS_COOLDOWN_MS = 2000;
/** Cửa sổ đếm lỗi lặp (§13.8: 3 lần / 10 s). */
export const SYS_WINDOW_MS = 10_000;
export const SYS_STRIKES = 3;

interface SysState {
  /** offUntil: mốc (ms) được chạy lại. 0 = đang chạy. */
  offUntil: number;
  /** dead: đã tắt hẳn tới hết mạng. */
  dead: boolean;
  /** hits: mốc các lần lỗi còn trong cửa sổ 10 s. */
  hits: number[];
  /** errors: tổng số lỗi của hệ thống này trong mạng hiện tại. */
  errors: number;
  /** lastMsg: thông báo lỗi gần nhất (cho bảng F3). */
  lastMsg: string;
}

const state = new Map<SysKey, SysState>();

function st(k: SysKey): SysState {
  let s = state.get(k);
  if (!s) {
    s = { offUntil: 0, dead: false, hits: [], errors: 0, lastMsg: '' };
    state.set(k, s);
  }
  return s;
}

/** Báo cáo một hệ thống vừa lỗi. `dead` = đã tắt hẳn tới hết mạng. */
export type SysErrorHook = (info: { key: SysKey; msg: string; err: unknown; dead: boolean; errors: number }) => void;

let onError: SysErrorHook | null = null;

/** setSysErrorHook cắm nơi nhận báo cáo (telemetry). Gọi một lần lúc khởi động. */
export function setSysErrorHook(fn: SysErrorHook | null): void {
  onError = fn;
}

/**
 * now() tách ra để test điều khiển được đồng hồ.
 *
 * `performance.now()` không tua được, mà luật của §13.8 toàn là luật về THỜI GIAN (2 giây,
 * 3 lần trong 10 giây). Test phải chứng minh được đúng mấy con số đó.
 */
let clock: () => number = () => performance.now();

/** setSysClock chỉ dùng trong test. */
export function setSysClock(fn: (() => number) | null): void {
  clock = fn ?? ((): number => performance.now());
}

/**
 * sysRun chạy một hệ thống trong ranh giới bắt lỗi.
 *
 * Trả về true nếu hệ thống ĐÃ CHẠY (không bị tắt và không ném lỗi) — vài chỗ cần biết để
 * bỏ qua bước phụ thuộc phía sau.
 *
 * Giá phải trả khi không có lỗi: một lần `Map.get`, một phép so số và một khối try. V8 không
 * tính phí try/catch khi không ném — đo trên máy dev: không thấy khác biệt trong `vẽ ms`.
 */
export function sysRun(key: SysKey, fn: () => void): boolean {
  const s = st(key);
  if (s.dead) {
    return false;
  }
  const t = clock();
  if (s.offUntil > t) {
    return false;
  }
  s.offUntil = 0;
  try {
    const left = faults.get(key) ?? 0;
    if (left > 0) {
      faults.set(key, left - 1);
      throw markFake(new Error(`lỗi giả ở hệ thống "${key}" (kx.fault)`));
    }
    fn();
    return true;
  } catch (err) {
    noteSysError(key, err, t);
    return false;
  }
}

/**
 * faults: số lần ném lỗi giả còn lại của mỗi hệ thống (checklist §17).
 *
 * Chỉ mở ra qua `window.kx` ở chế độ `?debug=true`, nhưng mã thì luôn có trong bundle: một
 * đường thử chỉ tồn tại ở bản dev là một đường thử chưa bao giờ chạy trên bản thật.
 */
const faults = new Map<SysKey, number>();

/**
 * markFake đóng dấu "đây là LỖI GIẢ do bài kiểm bắn ra" lên một Error.
 *
 * Vì sao cần: ngày 13/11, bảng lỗi 24 giờ có 57 dòng `lỗi giả trong khâu vẽ (kx.faultDraw)`
 * nằm lẫn với 95 dòng lỗi THẬT. Mỗi lượt deploy chạy `crashtest.mjs` là thêm vài dòng giả
 * nữa. Lọc bằng cách so chuỗi "lỗi giả" thì hỏng ngay khi ai đó sửa câu chữ, nên dấu phải
 * đi theo chính đối tượng lỗi.
 */
export function markFake<T>(err: T): T {
  (err as { kxFake?: boolean }).kxFake = true;
  return err;
}

/** isFake: lỗi này có phải do bài kiểm bắn ra không. */
export function isFake(err: unknown): boolean {
  return Boolean((err as { kxFake?: boolean } | null)?.kxFake);
}

/** faultSystem hẹn hệ thống `key` ném lỗi ở `n` khung tiếp theo. */
export function faultSystem(key: SysKey, n = 1): void {
  faults.set(key, Math.max(0, n));
}

/** noteSysError ghi một lỗi của hệ thống và quyết định tắt tạm hay tắt hẳn. */
export function noteSysError(key: SysKey, err: unknown, t = clock()): void {
  const s = st(key);
  s.errors++;
  s.lastMsg = errText(err);
  s.hits = s.hits.filter((x) => t - x < SYS_WINDOW_MS);
  s.hits.push(t);
  if (s.hits.length >= SYS_STRIKES) {
    s.dead = true;
    s.offUntil = Number.POSITIVE_INFINITY;
  } else {
    s.offUntil = t + SYS_COOLDOWN_MS;
  }
  // Console vẫn phải có: người đang mở DevTools cần thấy vết gọi đầy đủ, mà telemetry chỉ
  // chở 5 dòng stack đã cắt đường dẫn.
  console.warn(`[chống sập] hệ thống "${key}" lỗi${s.dead ? ' — TẮT HẲN tới hết mạng' : ` — tắt ${SYS_COOLDOWN_MS / 1000}s`}`, err);
  onError?.({ key, msg: s.lastMsg, err, dead: s.dead, errors: s.errors });
}

/** sysOff: hệ thống này đang bị tắt (tạm hoặc hẳn) không. */
export function sysOff(key: SysKey): boolean {
  const s = state.get(key);
  return !!s && (s.dead || s.offUntil > clock());
}

/** sysDead: đã tắt hẳn tới hết mạng chưa. */
export function sysDead(key: SysKey): boolean {
  return !!state.get(key)?.dead;
}

/**
 * sysResetRun xoá sổ khi vào mạng mới.
 *
 * "Tới hết mạng" của §13.8 nghĩa là hết VÁN, không phải hết đời tab: chết rồi bấm "Chơi
 * tiếp" phải được thử lại từ đầu — rất nhiều lỗi vẽ chỉ gắn với một khu hoặc một đối thủ cụ
 * thể và sẽ không tái diễn ở ván sau.
 */
export function sysResetRun(): void {
  state.clear();
  resetDrawGuard();
}

/** sysSummary: danh sách hệ thống đang hỏng, cho bảng F3 và telemetry. */
export function sysSummary(): { key: SysKey; dead: boolean; errors: number; msg: string }[] {
  const out: { key: SysKey; dead: boolean; errors: number; msg: string }[] = [];
  for (const k of SYS_KEYS) {
    const s = state.get(k);
    if (s && s.errors > 0) {
      out.push({ key: k, dead: s.dead, errors: s.errors, msg: s.lastMsg });
    }
  }
  return out;
}

/** sysLine: một dòng cho bảng F3; rỗng khi mọi thứ bình thường. */
export function sysLine(): string {
  const bad = sysSummary();
  if (bad.length === 0) {
    return '';
  }
  return `hệ thống hỏng: ${bad.map((b) => `${b.key}${b.dead ? ' (tắt hẳn)' : ` ×${b.errors}`}`).join(' · ')}`; // i18n-ok: lớp số liệu debug
}

/** sysOffList: tên các hệ thống đang tắt, để đính kèm mọi báo cáo lỗi (§13.8 "hệ thống đang chạy"). */
export function sysOffList(): string {
  return sysSummary().filter((b) => b.dead).map((b) => b.key).join(',');
}

/** guardLines: dòng F3 của lớp chống sập; mảng RỖNG khi chưa có sự cố nào. */
export function guardLines(rebuilds: number, watchdog: number): string[] {
  const out: string[] = [];
  const sys = sysLine();
  if (sys) {
    out.push(sys);
  }
  const d = drawGuardStats();
  if (d.errors > 0 || rebuilds > 0 || watchdog > 0) {
    out.push(`chống sập: lỗi vẽ ${d.errors} · dựng lại ${rebuilds} · watchdog ${watchdog}${filtersOff() ? ' · filter TẮT' : ''}`); // i18n-ok: lớp số liệu debug
  }
  return out;
}

/** errText rút gọn một lỗi thành một dòng đọc được. */
export function errText(err: unknown): string {
  if (err instanceof Error) {
    return `${err.name}: ${err.message}`;
  }
  return String(err).slice(0, 200);
}

/* ─────────────────────────────────────────────────────────────────────────────────────────
 * Lỗi trong khâu VẼ của Pixi — nhánh riêng, luật riêng (§13.8 đoạn 2).
 *
 * Khác hẳn lỗi hệ thống: ở đây không có "hệ thống" nào để tắt, vì lỗi xảy ra bên trong
 * `renderer.render()` khi nó duyệt cây cảnh. Luật của spec: bỏ khung đó + tắt bộ lọc đang
 * bật, và nếu lặp ≥ 3 lần/10 s thì dựng lại renderer + cây cảnh.
 * ───────────────────────────────────────────────────────────────────────────────────────── */

export type DrawAction = 'skip' | 'rebuild';

const draw = { hits: [] as number[], total: 0, rebuilds: 0, lastMsg: '' };

/** noteDrawError ghi một lỗi vẽ và nói nên làm gì: bỏ khung, hay dựng lại renderer. */
export function noteDrawError(err: unknown, t = clock()): DrawAction {
  draw.total++;
  draw.lastMsg = errText(err);
  draw.hits = draw.hits.filter((x) => t - x < SYS_WINDOW_MS);
  draw.hits.push(t);
  const rebuild = draw.hits.length >= SYS_STRIKES;
  if (rebuild) {
    // Dựng lại là một lần "chữa"; đếm lại từ đầu, nếu không thì mọi khung sau đó đều đòi
    // dựng lại và ta có một vòng dựng-lại vô tận.
    draw.hits.length = 0;
    draw.rebuilds++;
  }
  console.warn(`[chống sập] lỗi vẽ Pixi${rebuild ? ' — DỰNG LẠI renderer' : ' — bỏ khung + tắt bộ lọc'}`, err);
  return rebuild ? 'rebuild' : 'skip';
}

/** drawStatsGuard: số liệu lỗi vẽ cho bảng F3 / telemetry. */
export function drawGuardStats(): { errors: number; rebuilds: number; msg: string } {
  return { errors: draw.total, rebuilds: draw.rebuilds, msg: draw.lastMsg };
}

/**
 * Công tắc TẮT HẲN bộ lọc (bloom, sóng xung kích, color matrix) sau một lỗi vẽ.
 *
 * Spec §13.8: lỗi trong khâu vẽ của Pixi → bỏ khung đó **và tắt bộ lọc đang bật**. Lý do
 * thực tế: gần như mọi lỗi vẽ từng gặp (`uProjectionMatrix` là ví dụ trong chính spec) đến
 * từ đường render-to-texture của filter, không từ sprite thường. Tắt filter là cách chữa
 * rẻ nhất, và cái mất là hiệu ứng chứ không phải trận đấu.
 *
 * Một chiều: bật lại trong cùng một mạng thì lỗi quay lại ngay. Mạng mới thì `sysResetRun`
 * mở lại.
 */
let filtersKilled = false;

export function filtersOff(): boolean {
  return filtersKilled;
}

export function killFilters(): void {
  filtersKilled = true;
}

export function resetDrawGuard(): void {
  filtersKilled = false;
  draw.hits.length = 0;
  draw.total = 0;
  draw.rebuilds = 0;
  draw.lastMsg = '';
}
