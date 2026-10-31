import { Assets, Rectangle, Texture } from 'pixi.js';

import { CHARACTER_ANIMS } from './catalog';

// Bộ nạp asset theo NHÓM (GAME_SPEC §13.7 mục 3).
//
// Bản trước tải **mọi** tệp trong manifest rồi mới cho vào trang chủ: 547 tệp, 77,4 MB. Đó
// là nguyên nhân "chờ tải rất lâu" — không phải mạng chậm, mà là tải tất cả những thứ chưa
// cần. Giờ chỉ nhóm `boot` + nhân vật đã chọn + khu được xếp vào là chặn; còn lại tải nền
// trong lúc chơi.
//
// Đường dẫn trong mã nguồn vẫn là đường dẫn LOGIC (`ui/medal_gold.png`) y như cũ. Manifest
// dịch sang tên thật có hash (`ui/medal_gold.1a2b3c4d.webp`). Nhờ vậy đổi sang WebP + hash
// không phải sửa một chỗ gọi nào.

interface AtlasJson {
  cell: number;
  pivot_x: number;
  pivot_y: number;
  /** image: tên tệp ảnh (đã có hash) nằm cùng thư mục — pipeline ghi vào. */
  image?: string;
  frames: Record<string, { x: number; y: number; w: number; h: number }>;
  anims: Record<string, { frames: string[]; fps: number; loop: boolean }>;
}

interface ManifestFile {
  url: string;
  bytes: number;
  groups: string[];
}

interface Manifest {
  v: number;
  zones: string[];
  chars: string[];
  groups: Record<string, { bytes: number; files: string[] }>;
  files: Record<string, ManifestFile>;
}

export interface LoadProgress {
  /** byte đã tải xong (chỉ đếm tệp đã về đủ — đủ mịn vì tệp nặng nhất là 338 KB). */
  loaded: number;
  total: number;
}

/**
 * miscut nhận ra hoạt ảnh bị cắt sai số khung.
 *
 * Bộ cắt của artgen tìm khung bằng vùng pixel LIỀN NHAU, nên nhân vật nào có tay/chân/đồ vật
 * tách rời thân là bị đếm thành nhiều khung. Đã xảy ra thật: `dai_ca_dep` idle ra 22 khung
 * thay vì 4 và trong game thành một cái áo tím nhấp nháy. Thà bỏ hoạt ảnh hỏng và rơi về
 * hoạt ảnh thay thế còn hơn vẽ ra thứ đó.
 */
function miscut(name: string, got: number): boolean {
  const want = CHARACTER_ANIMS.find((a) => a.name === name)?.frames;
  return want !== undefined && (got > want * 1.5 || got * 2 < want);
}

export interface AnimSet {
  anims: Map<string, { textures: Texture[]; fps: number; loop: boolean }>;
  pivotX: number;
  pivotY: number;
}

class AssetStore {
  private man: Manifest | null = null;
  private readonly textures = new Map<string, Texture>();
  private readonly atlases = new Map<string, AnimSet>();
  /** loaded: đường dẫn logic đã tải xong. Dùng để không tải lại khi nhóm chồng nhau. */
  private readonly loaded = new Set<string>();
  private readonly inflight = new Map<string, Promise<void>>();

  /**
   * gen tăng mỗi lần có thêm asset về.
   *
   * Bộ vẽ so `gen` với lần nó kiểm trước để biết "có đáng thử đổi placeholder sang ảnh thật
   * không". Không có mốc này thì hoặc phải thử lại MỖI KHUNG (tốn vô ích, vì 99,9% khung
   * không có gì mới), hoặc không bao giờ thử lại và người chơi kẹt với placeholder — chính
   * là lỗi §13.7 cảnh báo.
   */
  gen = 0;

  async init(): Promise<void> {
    try {
      const res = await fetch('/assets/manifest.json', { cache: 'no-cache' });
      if (!res.ok) {
        throw new Error(`HTTP ${res.status}`);
      }
      const m = (await res.json()) as Manifest | { files: string[] };
      if (Array.isArray((m as { files: string[] }).files)) {
        // manifest v1 (`files: string[]`): server đang phục vụ `assets/src` chứ không phải
        // `assets/dist`. Nói thẳng ra thay vì đọc rác rồi vẽ placeholder cả trận.
        throw new Error('manifest v1 — chưa chạy `node tools/asset-pipeline/index.mjs`');
      }
      this.man = m as Manifest;
    } catch (err) {
      console.warn('không đọc được manifest asset — game sẽ chạy bằng placeholder', err);
    }
  }

  /** ready: manifest đã đọc được chưa. Chưa thì mọi thứ là placeholder. */
  get ready(): boolean {
    return this.man !== null;
  }

  /** chars: danh sách nhân vật có art, theo manifest. */
  get chars(): string[] {
    return this.man?.chars ?? [];
  }

  /** has: manifest CÓ asset này (không nói đã tải xong chưa). */
  has(path: string): boolean {
    return this.man?.files[path] !== undefined;
  }

  /**
   * url trả về địa chỉ thật (có hash) để gán vào `<img src>`.
   *
   * Dùng cho DOM — thẻ `<img>` tự tải lấy, không cần đi qua bộ nạp Pixi, nên không phải chờ
   * nhóm nào tải xong. Trả về '' nếu manifest không có.
   */
  url(path: string): string {
    const f = this.man?.files[path];
    return f ? `/assets/${f.url}` : '';
  }

  /** files: hợp của các nhóm, bỏ trùng, bỏ phần đã tải. */
  pending(groups: string[]): string[] {
    const out = new Set<string>();
    for (const g of groups) {
      for (const rel of this.man?.groups[g]?.files ?? []) {
        if (!this.loaded.has(rel)) {
          out.add(rel);
        }
      }
    }
    return [...out];
  }

  /** bytes: tổng dung lượng của danh sách đường dẫn logic. */
  bytes(rels: string[]): number {
    let n = 0;
    for (const rel of rels) {
      n += this.man?.files[rel]?.bytes ?? 0;
    }
    return n;
  }

  /**
   * load tải các nhóm và báo tiến độ THEO BYTE (§13.7 mục 4).
   *
   * Byte chứ không phải số tệp: `ui/home_bg` nặng gấp 300 lần một icon buff, nên thanh tính
   * theo tệp chạy giật cục — vọt lên 90% rồi đứng im ở tệp cuối.
   */
  async load(groups: string[], onProgress?: (p: LoadProgress) => void): Promise<void> {
    const rels = this.pending(groups);
    const total = this.bytes(rels);
    let done = 0;
    onProgress?.({ loaded: 0, total });
    // Tải song song nhưng có TRẦN: 547 fetch cùng lúc làm trình duyệt tự xếp hàng và thanh
    // tiến độ đứng im mấy giây đầu. 12 là đủ bão hoà một đường 4G.
    const queue = [...rels];
    const worker = async (): Promise<void> => {
      for (;;) {
        const rel = queue.shift();
        if (rel === undefined) {
          return;
        }
        await this.one(rel);
        done += this.man?.files[rel]?.bytes ?? 0;
        onProgress?.({ loaded: done, total });
      }
    };
    await Promise.all(Array.from({ length: Math.min(12, Math.max(1, queue.length)) }, worker));
    this.gen++;
  }

  /** one tải một đường dẫn logic, gộp các lời gọi trùng nhau. */
  private one(rel: string): Promise<void> {
    const started = this.inflight.get(rel);
    if (started) {
      return started;
    }
    if (this.loaded.has(rel)) {
      return Promise.resolve();
    }
    const p = (async () => {
      try {
        if (rel.endsWith('atlas.json')) {
          await this.loadAtlas(rel);
        } else if (/\.(png|webp|jpg|jpeg)$/.test(rel)) {
          await this.loadTexture(rel);
        }
        // Font, âm thanh và tệp dữ liệu không đi qua đây: `loadFonts()` và `howler` tự tải
        // bằng URL lấy từ `url()`. Vẫn đánh dấu là xong để tiến độ theo byte chạy đủ 100%.
      } finally {
        this.loaded.add(rel);
        this.inflight.delete(rel);
      }
    })();
    this.inflight.set(rel, p);
    return p;
  }

  /**
   * bg: tiến độ các nhóm đang tải NỀN, cho icon % góc dưới trong trận (§13.7 mục 4 dòng cuối).
   */
  private bgLoaded = 0;
  private bgTotal = 0;

  bgProgress(): { pct: number; running: boolean } {
    if (this.bgTotal <= 0 || this.bgLoaded >= this.bgTotal) {
      return { pct: 1, running: false };
    }
    return { pct: this.bgLoaded / this.bgTotal, running: true };
  }

  /**
   * loadBackground tải nền, không chặn ai, không bao giờ reject.
   *
   * Thứ tự nhóm truyền vào LÀ thứ tự ưu tiên (§13.7: "ưu tiên khu kề cổng gần nhất").
   */
  loadBackground(groups: string[]): void {
    const rels = this.pending(groups);
    if (!rels.length) {
      return;
    }
    this.bgTotal += this.bytes(rels);
    void (async () => {
      for (const rel of rels) {
        try {
          await this.one(rel);
        } catch (err) {
          console.warn('asset nền lỗi', rel, err);
        }
        this.bgLoaded += this.man?.files[rel]?.bytes ?? 0;
        this.gen++;
      }
    })();
  }

  /**
   * charAtlas lấy atlas nhân vật, có NHÂN VẬT DỰ PHÒNG (NETCODE_SPEC §5 Hiển thị mục 1).
   *
   * Build thật CẤM placeholder. Thiếu atlas của một nhân vật thì vẽ bằng nhân vật dự phòng
   * — art thật, nhìn vẫn ra một người — chứ không vẽ vòng tròn màu kèm chữ viết tắt.
   *
   * Lý do cấm: vòng tròn placeholder là công cụ của người làm game, không phải thứ người
   * chơi nên thấy. Ngày 30/10 telemetry đo được 12 khung placeholder trên server thật, và
   * người chơi không có cách nào biết đó là "art chưa xong" hay "game hỏng".
   *
   * Với bộ nạp theo nhóm, nhánh dự phòng còn gánh thêm một việc: trong vài giây đầu trận,
   * atlas của 9 nhân vật kia CHƯA tải xong (nhóm `chars` tải nền). Vẽ bằng nhân vật mình đã
   * chọn là đúng tinh thần "placeholder đã có rồi đổi sang ảnh thật".
   */
  charAtlas(char: string): { set: AnimSet; fallback: boolean } | null {
    const want = this.atlas(`characters/${char}`);
    if (want) {
      return { set: want, fallback: false };
    }
    // Yêu cầu tải ngay nhân vật đang thiếu: nó vừa xuất hiện trong khung hình nên là nhân
    // vật ĐÁNG tải trước nhất, dù nhóm `chars` đang xếp hàng theo thứ tự khác.
    if (this.has(`characters/${char}/atlas.json`)) {
      this.loadBackground([`char_${char}`]);
    }
    for (const alt of this.fallbackChars) {
      const set = this.atlas(`characters/${alt}`);
      if (set) {
        if (!this.warnedChars.has(char)) {
          this.warnedChars.add(char);
          console.warn(`chưa có atlas nhân vật "${char}" — dùng tạm "${alt}" (NETCODE_SPEC §5.1)`);
        }
        return { set, fallback: true };
      }
    }
    return null;
  }

  // ── Nhả atlas nhân vật không còn ai dùng (QUESTIONS Q30, chủ dự án chốt 31/10) ────────
  //
  // Vì sao cần: iPhone thật gửi về `gpu_pressure` ở **216 MB / trần 256** với `chars:7` —
  // bảy atlas nhân vật nằm trong bộ nhớ. Tự hạ chất lượng chạy rồi mà bộ nhớ TĂNG (214 ở
  // "Vừa" → 216 ở "Thấp"): hạ chất lượng không trả lại texture đã nạp.
  //
  // Cách làm: đếm xem mỗi nhân vật còn bao nhiêu người trong khu đang dùng. Về 0 thì CHỜ
  // 10 GIÂY rồi mới nhả. Chờ là bắt buộc — người chơi đi ra đi vào ranh giới tầm nhìn liên
  // tục, và nhả ngay là nạp lại ngay, cứ thế mỗi vài giây: tốn băng thông, giật khung hình,
  // mà bộ nhớ chẳng giảm.

  /** CHAR_RELEASE_DELAY_MS: chờ ngần này sau khi không còn ai dùng mới nhả. */
  private static readonly CHAR_RELEASE_DELAY_MS = 10_000;

  /** charIdleSince: nhân vật → thời điểm bắt đầu không còn ai dùng. */
  private readonly charIdleSince = new Map<string, number>();

  /**
   * keepChars báo danh sách nhân vật ĐANG CẦN, gọi mỗi khung từ vòng vẽ.
   *
   * Nhận cả danh sách thay vì đếm tăng/giảm từng con: đếm tham chiếu bằng tay thì chỉ cần
   * một đường thoát quên giảm là con số kẹt mãi trên 0 và không bao giờ nhả được — mà
   * đường thoát thì có nhiều (chết, qua cổng, despawn, đổi khu, mất kết nối). Danh sách
   * "ai đang trên màn hình" thì không có đường nào để quên.
   */
  keepChars(inUse: Iterable<string>, mine: string, now: number): void {
    const live = new Set(inUse);
    live.add(mine); // nhân vật của chính mình không bao giờ nhả
    for (const rel of this.atlases.keys()) {
      const m = /^characters\/(.+)$/.exec(rel);
      if (!m) {
        continue;
      }
      const char = m[1];
      if (live.has(char)) {
        this.charIdleSince.delete(char);
        continue;
      }
      const since = this.charIdleSince.get(char);
      if (since === undefined) {
        this.charIdleSince.set(char, now);
        continue;
      }
      if (now - since >= AssetStore.CHAR_RELEASE_DELAY_MS) {
        this.releaseChar(char);
      }
    }
  }

  /**
   * releaseIdleCharsNow nhả NGAY mọi atlas đang rảnh, không chờ 10 giây.
   *
   * Gọi khi chạm trần bộ nhớ (`gpuPressure`). Lúc đó Safari sắp thu hồi tab, và 10 giây
   * chờ cho êm là một sự xa xỉ — chờ để tránh nạp lại liên tục thì vô nghĩa nếu tab chết
   * trước khi kịp nạp lại lần nào.
   *
   * Trả về số atlas đã nhả, để chỗ gọi ghi vào bằng chứng gửi về server.
   */
  releaseIdleCharsNow(): number {
    let n = 0;
    for (const char of [...this.charIdleSince.keys()]) {
      this.releaseChar(char);
      n++;
    }
    return n;
  }

  /**
   * releaseChar nhả atlas của một nhân vật khỏi bộ nhớ đồ hoạ.
   *
   * Gọi `Texture.destroy(true)` — tham số `true` huỷ cả nguồn ảnh. Thiếu nó thì chỉ bỏ
   * texture mà ảnh gốc vẫn nằm trên GPU, và con số `texMB` không nhúc nhích: đúng cái đã
   * xảy ra với cơ chế tự hạ chất lượng.
   */
  private releaseChar(char: string): void {
    const rel = `characters/${char}`;
    const set = this.atlases.get(rel);
    if (!set) {
      return;
    }
    const seen = new Set<unknown>();
    for (const anim of set.anims.values()) {
      for (const tex of anim.textures) {
        const src = (tex as unknown as { source?: unknown }).source;
        if (src && !seen.has(src)) {
          seen.add(src);
        }
        tex.destroy(false); // khung hình chỉ là cửa sổ vào ảnh nguồn
      }
    }
    // Huỷ ảnh NGUỒN đúng một lần: mọi khung của một atlas dùng chung một ảnh.
    for (const src of seen) {
      (src as { destroy?: () => void }).destroy?.();
    }
    this.atlases.delete(rel);
    this.charIdleSince.delete(char);
    // Bỏ dấu "đã tải" của MỌI tệp thuộc nhân vật này, không chỉ tên nhóm: `loaded` chứa
    // đường dẫn logic từng tệp, nên quên xoá chúng là lần sau `loadBackground` tưởng đã có
    // và không tải lại — nhân vật đó vĩnh viễn hiện bằng người dự phòng.
    for (const key of [...this.loaded]) {
      if (key.startsWith(`characters/${char}/`)) {
        this.loaded.delete(key);
      }
    }
  }

  /** fallbackChars: thứ tự thử nhân vật dự phòng. Đặt bởi `setFallbackChars` lúc có config. */
  private fallbackChars: string[] = [];
  private readonly warnedChars = new Set<string>();

  /** setFallbackChars nhận danh sách nhân vật từ config (thứ tự trong config là thứ tự thử). */
  setFallbackChars(chars: string[]): void {
    this.fallbackChars = chars;
  }

  private async loadTexture(rel: string): Promise<Texture | null> {
    const url = this.url(rel);
    if (!url) {
      return null;
    }
    try {
      const tex = (await Assets.load(url)) as Texture;
      this.textures.set(rel, tex);
      return tex;
    } catch (err) {
      console.warn('texture failed', rel, err);
      return null;
    }
  }

  private async loadAtlas(jsonRel: string): Promise<void> {
    const dir = jsonRel.slice(0, jsonRel.lastIndexOf('/'));
    const url = this.url(jsonRel);
    if (!url) {
      return;
    }
    try {
      const res = await fetch(url);
      const atlas = (await res.json()) as AtlasJson;
      // Tên ảnh lấy từ chính atlas (`image`, có hash) — không suy ra từ đường dẫn, vì hash
      // của ảnh không liên quan gì tới hash của JSON.
      const imgRel = `${dir}/atlas.png`;
      const imgUrl = atlas.image ? `/assets/${dir}/${atlas.image}` : this.url(imgRel);
      if (!imgUrl) {
        return;
      }
      const base = (await Assets.load(imgUrl)) as Texture;
      this.textures.set(imgRel, base);
      this.loaded.add(imgRel);
      const set: AnimSet = { anims: new Map(), pivotX: atlas.pivot_x, pivotY: atlas.pivot_y };
      for (const [name, a] of Object.entries(atlas.anims)) {
        if (miscut(name, a.frames.length)) {
          // Sheet bị cắt sai (xem artgen prep): bỏ hẳn hoạt ảnh này để nhân vật rơi về
          // hoạt ảnh thay thế, thay vì nhấp nháy mấy mảnh vụn của chính nó.
          console.warn(`atlas ${dir}: bỏ "${name}" vì cắt sai (${a.frames.length} khung)`);
          continue;
        }
        const textures = a.frames.map((f) => {
          const r = atlas.frames[f];
          return new Texture({ source: base.source, frame: new Rectangle(r.x, r.y, r.w, r.h) });
        });
        set.anims.set(name, { textures, fps: a.fps, loop: a.loop });
      }
      this.atlases.set(dir, set);
    } catch (err) {
      console.warn('atlas failed', jsonRel, err);
    }
  }

  texture(path: string): Texture | null {
    return this.textures.get(path) ?? null;
  }

  /** Frames like vfx/clash_spark/clash_spark_00.png … in order, or [] if missing. */
  frames(dir: string, anim: string): Texture[] {
    const out: Texture[] = [];
    for (let i = 0; i < 99; i++) {
      const t = this.textures.get(`${dir}/${anim}_${String(i).padStart(2, '0')}.png`);
      if (!t) {
        break;
      }
      out.push(t);
    }
    return out;
  }

  atlas(dir: string): AnimSet | null {
    return this.atlases.get(dir) ?? null;
  }
}

export const assets = new AssetStore();
