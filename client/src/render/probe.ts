/**
 * Đồng hồ bấm giờ theo TỪNG HỆ THỐNG trong vòng vẽ — cho bảng F3 (tối ưu máy yếu, bước 0).
 *
 * ── Vì sao cần ───────────────────────────────────────────────────────────────────────────
 *
 * F3 đang nói "vẽ 12,4 ms" nhưng không nói 12,4 ms đó đi đâu. `cputrace.mjs` cho hồ sơ theo
 * HÀM, đọc được nhưng phải dựng Chrome và đọc cây hàm — không dùng được lúc đang chơi trên
 * máy thật, mà máy thật mới là chỗ cần biết. Bảng này trả lời thẳng: nội suy bao nhiêu,
 * vòng kiếm bao nhiêu, bảng tên bao nhiêu…
 *
 * ── Giá phải trả khi TẮT ─────────────────────────────────────────────────────────────────
 *
 * Một phép so `if (!on)` rồi `return` cho mỗi mốc. Không `performance.now()`, không Map,
 * không cấp phát. Đo trên máy này: không thấy khác biệt trong `vẽ ms`.
 *
 * Bật tự động khi mở F3 (`Hud.perfVisible`), nên người chơi thường không bao giờ trả giá.
 */

/**
 * Tên mốc để ASCII, nhãn tiếng Việt để riêng ở `LABELS`.
 *
 * Không phải chuyện thẩm mỹ: `check:i18n` chặn chuỗi tiếng Việt nằm ngoài `ui/strings.ts`
 * (quy tắc 5), và nó chặn đúng — một ngày nào đó có người tưởng mấy chuỗi này là chữ cho
 * người chơi rồi đem đi dịch. Khoá ASCII thì không có chuyện đó.
 */
export type ProbeKey =
  | 'interp' | 'players' | 'monsters' | 'items' | 'beasts'
  | 'map' | 'ambient' | 'vfx' | 'skills' | 'other'
  // Mốc CON bên trong `players` — cộng lại xấp xỉ bằng `players`. Biết "người 3,5 ms" là
  // chưa đủ để sửa: phải biết 3,5 ms đó là vòng kiếm, bảng tên hay thân nhân vật.
  | 'p.body' | 'p.plate' | 'p.ring' | 'p.status';

const KEYS: ProbeKey[] = ['interp', 'players', 'monsters', 'items', 'beasts',
  'map', 'ambient', 'vfx', 'skills', 'other',
  'p.body', 'p.plate', 'p.ring', 'p.status'];

/** Nhãn hiện trên bảng F3. */
const LABELS: Record<ProbeKey, string> = {
  interp: 'nội suy', players: 'người', monsters: 'quái', items: 'vật phẩm', beasts: 'thần thú', // i18n-ok: lớp số liệu debug
  map: 'bản đồ', ambient: 'ambient', vfx: 'vfx', skills: 'kỹ năng', other: 'khác', // i18n-ok: lớp số liệu debug
  'p.body': 'ng.thân', 'p.plate': 'ng.bảngtên', 'p.ring': 'ng.vòngkiếm', 'p.status': 'ng.trạngthái', // i18n-ok: lớp số liệu debug
};

let on = false;
let last = 0;
/** Trung bình trượt mỗi mốc (ms). Mảng song song với `KEYS` — không Map, không cấp phát. */
const ema = new Float64Array(KEYS.length);
const idx = new Map<ProbeKey, number>(KEYS.map((k, i) => [k, i]));

/** Heap dùng (MB) — Chrome mới có; dùng để nhìn răng cưa GC. */
let heapNow = 0;
let heapMin = 0;
let heapMax = 0;
let heapAt = 0;

export function probeEnabled(v: boolean): void {
  if (on === v) {
    return;
  }
  on = v;
  ema.fill(0);
  heapMin = 0;
  heapMax = 0;
}

/** Gọi ở ĐẦU vòng vẽ. */
export function probeFrame(): void {
  if (!on) {
    return;
  }
  last = performance.now();
}

/** Gọi ngay SAU mỗi hệ thống — tính từ mốc trước tới giờ. */
export function probeMark(key: ProbeKey): void {
  if (!on) {
    return;
  }
  const now = performance.now();
  const i = idx.get(key) ?? KEYS.length - 1;
  ema[i] = ema[i] * 0.9 + (now - last) * 0.1;
  last = now;
}

/**
 * Mốc CON: cộng dồn trong một khung rồi mới vào trung bình trượt ở `probeSubFlush`.
 *
 * Khác `probeMark` vì nó chạy MỘT LẦN MỖI NGƯỜI, không phải một lần mỗi khung. Dùng
 * `probeMark` ở đây thì trung bình trượt chỉ nhớ người cuối cùng, không phải cả màn hình.
 */
const subAcc = new Float64Array(KEYS.length);
const SUB_KEYS: ProbeKey[] = ['p.body', 'p.plate', 'p.ring', 'p.status'];
let subLast = 0;

export function probeSub(): void {
  if (on) {
    subLast = performance.now();
  }
}

export function probeSubMark(key: ProbeKey): void {
  if (!on) {
    return;
  }
  const now = performance.now();
  subAcc[idx.get(key) ?? KEYS.length - 1] += now - subLast;
  subLast = now;
}

/** Gọi sau khi đã duyệt hết người trong khung: dồn tổng vào trung bình trượt. */
export function probeSubFlush(): void {
  if (!on) {
    return;
  }
  // CHỈ các mốc con. Quét cả mảng thì mấy mốc chính (đặt bằng `probeMark`) bị nhân 0.9 với
  // `subAcc = 0` mỗi khung và tắt dần về 0 — bảng sẽ nói "mọi hệ thống đều 0 ms".
  for (const k of SUB_KEYS) {
    const i = idx.get(k) ?? 0;
    ema[i] = ema[i] * 0.9 + subAcc[i] * 0.1;
    subAcc[i] = 0;
  }
}

/** Gọi một lần mỗi giây: chụp heap để thấy răng cưa GC. */
export function probeHeap(nowSec: number): void {
  if (!on || nowSec - heapAt < 1) {
    return;
  }
  heapAt = nowSec;
  const m = (performance as unknown as { memory?: { usedJSHeapSize: number } }).memory;
  if (!m) {
    return;
  }
  heapNow = m.usedJSHeapSize / 1048576;
  heapMin = heapMin === 0 ? heapNow : Math.min(heapMin, heapNow);
  heapMax = Math.max(heapMax, heapNow);
}

/** Hai dòng cho bảng F3; rỗng khi tắt. */
export function probeLines(): string[] {
  if (!on) {
    return [];
  }
  const parts: string[] = [];
  for (let i = 0; i < KEYS.length; i++) {
    if (ema[i] >= 0.05) {
      parts.push(`${LABELS[KEYS[i]]} ${ema[i].toFixed(2)}`);
    }
  }
  const out = [`ms/hệ thống: ${parts.join(' · ')}`]; // i18n-ok: lớp số liệu debug
  if (heapNow > 0) {
    // Biên độ răng cưa = cấp phát trong vòng khung. Phẳng = không rác, đúng mục tiêu bước 2.
    out.push(`heap ${heapNow.toFixed(0)} MB (thấp ${heapMin.toFixed(0)} · cao ${heapMax.toFixed(0)} · răng cưa ${(heapMax - heapMin).toFixed(0)})`); // i18n-ok: lớp số liệu debug
  }
  return out;
}

/** Số liệu thô cho bài đo tự động (`window.kx.scene.probe()`). */
export function probeStats(): Record<string, number> {
  const out: Record<string, number> = {};
  for (let i = 0; i < KEYS.length; i++) {
    out[KEYS[i]] = +ema[i].toFixed(3);
  }
  out.heapMB = +heapNow.toFixed(1);
  out.heapRangeMB = +(heapMax - heapMin).toFixed(1);
  return out;
}
