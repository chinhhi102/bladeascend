/**
 * Bộ đếm LÔ VẼ và số nút cây cảnh — dùng chung cho `Scene`, cảnh thử và bài đo.
 *
 * Tách ra khỏi `scene.ts` ngày 13/11: cảnh thử (`testbed.ts`) cần đúng bộ đếm này, mà nạp
 * cả `scene.ts` chỉ để lấy một biến đếm là kéo theo toàn bộ `GameScene` vào gói tải chậm.
 * Mục 5 của lượt tối ưu (giảm lô vẽ) cũng đọc từ đây.
 */

/**
 * drawCalls đếm số lệnh vẽ WebGL mỗi khung. Pixi không công bố con số này, mà nó lại là chỉ
 * số quan trọng nhất để biết nghẽn ở GPU hay ở CPU, nên bọc thẳng hàm vẽ của ngữ cảnh WebGL.
 * Chỉ bật ở chế độ debug (spec §13.6).
 */
export const drawCalls = { perFrame: 0, acc: 0, frames: 0, on: false, ready: false };

/**
 * countDrawCalls bọc hàm vẽ — gọi một lần. Trước 1/11 chỉ gắn khi URL có `?debug=true`, còn
 * phím F3 thì mở được overlay ở MỌI trang: overlay đọc một bộ đếm chưa bao giờ được gắn và
 * ghi "draw call 0" giữa lúc đang vẽ 5 000 sprite (chủ dự án chụp được). Giờ mở overlay là
 * gắn (`perfLine`), và chưa đo xong cửa sổ đầu thì ghi "—" chứ không ghi số 0 giả.
 */
export function countDrawCalls(gl: WebGL2RenderingContext): void {
  if (drawCalls.on) {
    return;
  }
  drawCalls.on = true;
  for (const name of ['drawElements', 'drawArrays', 'drawElementsInstanced', 'drawArraysInstanced'] as const) {
    const fn = gl[name] as (...a: unknown[]) => void;
    (gl as unknown as Record<string, unknown>)[name] = (...a: unknown[]): void => {
      drawCalls.acc += 1;
      fn.apply(gl, a);
    };
  }
}

/** countNodes đếm số đối tượng hiển thị đang nằm trong cây vẽ. */
export function countNodes(c: { children?: unknown[] }): number {
  let n = 1;
  for (const k of (c.children ?? []) as { children?: unknown[] }[]) {
    n += countNodes(k);
  }
  return n;
}

/**
 * ── Đếm lô vẽ THEO TỪNG LỚP, và vì sao lô bị ngắt ───────────────────────────────────────
 *
 * Tổng số lô vẽ nói "đang tốn", không nói "tốn ở đâu". Mục 5 của lượt tối ưu cần chỉ đích
 * danh lớp nào ngốn lô và vì cớ gì, nếu không thì mọi cách chữa đều là đoán.
 *
 * Cách đo số lô của MỘT lớp: tắt mọi lớp khác, vẽ một khung, đọc bộ đếm. Thô nhưng đúng và
 * không phải móc vào ruột bộ gom lô của Pixi — mà ruột đó đổi theo từng bản Pixi, còn bài đo
 * thì phải sống lâu hơn một lần nâng cấp.
 *
 * Phần "vì sao ngắt" không đo được trực tiếp, nên thay bằng những thứ GÂY ra ngắt lô, đếm
 * bằng cách duyệt cây: đổi texture (đếm số nguồn texture khác nhau), đổi blend mode, vật
 * không gom lô được (Graphics, Text, Mesh), và filter/mask (mỗi cái là một lượt vẽ ra texture
 * riêng). Một lớp 40 lô với 38 texture khác nhau và một lớp 40 lô với 1 texture nhưng 20
 * Graphics là hai bệnh khác hẳn nhau.
 */
export interface LayerStat {
  name: string;
  calls: number;
  nodes: number;
  /** Số vật theo loại ống vẽ (`renderPipeId` — nhãn bền, không đổi khi minify). */
  kinds: Record<string, number>;
  /** Số NGUỒN texture khác nhau: mỗi lần đổi nguồn giữa hai vật liền nhau là một lần ngắt lô. */
  textures: number;
  /** Tên 40 nguồn ảnh hay gặp nhất trong lớp — để biết gom atlas nào thì bớt được lô vẽ. */
  texList?: { name: string; n: number }[];
  /** Số blend mode khác nhau. Đổi blend là ngắt lô, không có cách nào gom. */
  blends: number;
  filters: number;
  masks: number;
}

interface Walkable {
  visible?: boolean;
  renderPipeId?: string;
  blendMode?: string;
  filters?: unknown;
  mask?: unknown;
  texture?: { source?: { uid?: number; label?: string; width?: number; height?: number } };
  children?: Walkable[];
}

function tally(c: Walkable, st: LayerStat, tex: Set<number>, bl: Set<string>,
  names?: Map<number, { name: string; n: number }>): void {
  st.nodes++;
  const id = c.renderPipeId;
  if (id) {
    st.kinds[id] = (st.kinds[id] ?? 0) + 1;
  }
  if (c.blendMode) {
    bl.add(c.blendMode);
  }
  const src = c.texture?.source;
  const uid = src?.uid;
  if (uid !== undefined) {
    tex.add(uid);
    if (names) {
      // Tên nguồn + CỠ: hai thứ đủ để biết nên gom atlas nào. Một lớp 92 texture nghĩa là
      // ít nhất 92 lần đổi texture trong một khung, và mỗi lần đổi là một lô vẽ mới.
      const e = names.get(uid);
      if (e) {
        e.n++;
      } else {
        const lbl = src?.label ?? '?';
        names.set(uid, { name: `${lbl.split('/').slice(-2).join('/')} ${src?.width}x${src?.height}`, n: 1 });
      }
    }
  }
  if (Array.isArray(c.filters) ? c.filters.length > 0 : Boolean(c.filters)) {
    st.filters++;
  }
  if (c.mask) {
    st.masks++;
  }
  for (const k of c.children ?? []) {
    tally(k, st, tex, bl, names);
  }
}

/**
 * measureLayers trả bảng lô vẽ theo lớp. `render` phải vẽ ĐÚNG một khung.
 *
 * Trả lại nguyên trạng `visible` của mọi lớp khi xong — kể cả khi `render` ném lỗi, vì bỏ
 * một lớp ở trạng thái ẩn là làm hỏng chính cái cảnh đang đo.
 */
export function measureLayers(render: () => void, layers: { name: string; c: Walkable }[]): LayerStat[] {
  const was = layers.map((l) => l.c.visible !== false);
  const out: LayerStat[] = [];
  try {
    for (let i = 0; i < layers.length; i++) {
      for (let j = 0; j < layers.length; j++) {
        layers[j].c.visible = i === j;
      }
      drawCalls.acc = 0;
      render();
      const st: LayerStat = {
        name: layers[i].name, calls: drawCalls.acc, nodes: 0, kinds: {},
        textures: 0, blends: 0, filters: 0, masks: 0,
      };
      const tex = new Set<number>();
      const bl = new Set<string>();
      const names = new Map<number, { name: string; n: number }>();
      tally(layers[i].c, st, tex, bl, names);
      st.textures = tex.size;
      st.texList = [...names.values()].sort((a, b) => b.n - a.n).slice(0, 40);
      st.blends = bl.size;
      out.push(st);
    }
  } finally {
    for (let j = 0; j < layers.length; j++) {
      layers[j].c.visible = was[j];
    }
  }
  return out;
}
