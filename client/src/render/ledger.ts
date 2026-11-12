/**
 * Sổ tài nguyên — GAME_SPEC §13.9 ("vệ sinh bộ nhớ: không gì được tăng mãi").
 *
 * ── Vì sao cần một cái sổ ────────────────────────────────────────────────────────────────
 *
 * Rò rỉ trên GPU không gây lỗi. Nó chỉ làm con số `texMB` bò lên trong nhiều phút rồi một
 * ngày nào đó Safari giết tab giữa trận — và lúc đó không còn gì để đọc. Chuyện đã xảy ra
 * đúng như vậy: 120 MB thành 1,4 GB qua 181 lượt đổi khu, không một thông báo nào.
 *
 * Cách duy nhất bắt được loại này là ĐẾM, liên tục, theo loại, và so con số ở phút thứ 5
 * với con số ở phút thứ 30. Sổ này là chỗ lấy những con số đó.
 *
 * ── Cách đếm ─────────────────────────────────────────────────────────────────────────────
 *
 * Duyệt cây cảnh và phân loại theo `renderPipeId` — một chuỗi Pixi tự gắn trong constructor
 * (`'sprite'`, `'graphics'`, `'text'`, `'bitmapText'`, `'tilingSprite'`, `'particleContainer'`…).
 *
 * Vì sao không dùng `constructor.name`: bản phát hành được minify, nên `Sprite` thành `w`.
 * Một bảng số liệu chỉ đúng ở bản dev là một bảng số liệu vô dụng — rò rỉ xảy ra trên máy
 * người chơi, với bản họ tải về.
 *
 * Những thứ KHÔNG nằm trong cây cảnh (view trong pool, chữ trong cache) tự khai báo qua
 * `ledgerSet`. Pool chính là chỗ dễ biến thành rò rỉ nhất — nó giữ tham chiếu CÓ CHỦ Ý,
 * nên bộ dọn rác không bao giờ đụng tới.
 */

/** Kiểu nút Pixi mà sổ theo dõi. Khớp `renderPipeId`, trừ `container` (nút không vẽ gì). */
const KINDS = ['sprite', 'graphics', 'text', 'bitmapText', 'tilingSprite', 'mesh', 'particleContainer'] as const;
export type ResKind = (typeof KINDS)[number];

export interface Census {
  /** nodes: tổng số nút trong cây cảnh (gồm cả Container rỗng). */
  nodes: number;
  /** filters: số filter đang gắn trên cây — mỗi cái là một lượt render-to-texture. */
  filters: number;
  /** byKind: số nút theo loại. Khoá luôn có mặt, kể cả khi bằng 0. */
  byKind: Record<string, number>;
  /** extra: những thứ ngoài cây cảnh, do các lớp tự khai (pool, cache). */
  extra: Record<string, number>;
}

/** Nút cây cảnh ở mức tối thiểu sổ cần biết. Không import kiểu của Pixi để test chạy trong Node. */
interface Node {
  children?: Node[];
  renderPipeId?: string;
  filters?: unknown;
}

/** extra: bảng do các lớp tự khai (`ledgerSet`). */
const extra: Record<string, number> = {};

/**
 * ledgerSet khai số lượng một loại tài nguyên KHÔNG nằm trong cây cảnh.
 *
 * Đặt (set) chứ không cộng/trừ: đếm tăng-giảm bằng tay thì chỉ cần một đường thoát quên
 * giảm là con số sai vĩnh viễn — cùng bài học với `assets.keepChars`.
 */
export function ledgerSet(key: string, n: number): void {
  extra[key] = n;
}

/**
 * census duyệt cây cảnh và đếm theo loại.
 *
 * Giá: O(số nút). Gọi 1 lần/giây (không phải mỗi khung) — bảng F3 vốn đã duyệt cả cây mỗi
 * khung để đếm sprite, nên thêm một lượt mỗi giây không đáng kể.
 */
export function census(root: Node): Census {
  const byKind: Record<string, number> = {};
  for (const k of KINDS) {
    byKind[k] = 0;
  }
  let nodes = 0;
  let filters = 0;
  const walk = (n: Node): void => {
    nodes++;
    const id = n.renderPipeId;
    if (id !== undefined && byKind[id] !== undefined) {
      byKind[id]++;
    }
    const f = n.filters;
    if (Array.isArray(f)) {
      filters += f.length;
    } else if (f) {
      filters += 1;
    }
    for (const c of n.children ?? []) {
      walk(c);
    }
  };
  walk(root);
  return { nodes, filters, byKind, extra: { ...extra } };
}

/** censusFlat gộp thành một bảng số phẳng cho `perfStats()` và telemetry. */
export function censusFlat(c: Census): Record<string, number> {
  const out: Record<string, number> = { nodes: c.nodes, filters: c.filters };
  for (const [k, v] of Object.entries(c.byKind)) {
    out[`n_${k}`] = v;
  }
  for (const [k, v] of Object.entries(c.extra)) {
    out[`n_${k}`] = v;
  }
  return out;
}

/** censusLine: một dòng cho bảng F3. Bỏ loại bằng 0 để dòng còn đọc được. */
export function censusLine(c: Census): string {
  const parts = [`nút ${c.nodes}`]; // i18n-ok: lớp số liệu debug
  for (const [k, v] of Object.entries(c.byKind)) {
    if (v > 0) {
      parts.push(`${k} ${v}`);
    }
  }
  for (const [k, v] of Object.entries(c.extra)) {
    parts.push(`${k} ${v}`);
  }
  if (c.filters > 0) {
    parts.push(`filter ${c.filters}`);
  }
  return `tài nguyên: ${parts.join(' · ')}`; // i18n-ok: lớp số liệu debug
}

/**
 * capArray cắt một mảng lịch sử về trần, bỏ phần CŨ NHẤT.
 *
 * Một hàm dùng chung cho mọi mảng lịch sử của §13.9 (kill feed, lỗi, mẫu telemetry, vết
 * netcode). Viết tay `while (a.length > N) a.shift()` ở năm chỗ thì sẽ có một chỗ viết
 * `>=`, một chỗ quên hẳn, và không chỗ nào nói trần là bao nhiêu khi đọc lại sau sáu tháng.
 */
export function capArray<T>(a: T[], max: number): void {
  if (a.length > max) {
    a.splice(0, a.length - max);
  }
}
