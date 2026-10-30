// Màn hình tải (GAME_SPEC §13.7 mục 4) — điều khiển cái khung HTML/CSS có sẵn trong
// `index.html`.
//
// Lớp này KHÔNG dựng DOM. Khung đã nằm trong HTML để hiện được từ byte đầu tiên; ở đây chỉ
// đổi chữ, đổi độ rộng thanh, và chạy nhân vật trên thanh. Dựng lại bằng JS là quay về
// đúng quãng im lặng mà §13.7 muốn xoá.

import { assets, type LoadProgress } from '../render/assets';
import { STR } from './strings';

/** TIP_MS: mẹo/câu bựa đổi mỗi 4 giây (§13.7 mục 4). */
const TIP_MS = 4000;
/** RUN_FPS: nhịp chân của nhân vật chạy trên thanh. */
const RUN_FPS = 12;

function el<T extends HTMLElement>(id: string): T | null {
  return document.getElementById(id) as T | null;
}

export class LoadingScreen {
  private readonly root = el('boot');
  private readonly stepEl = el('boot-step');
  private readonly fill = el('boot-fill');
  private readonly bytesEl = el('boot-bytes');
  private readonly tipEl = el('boot-tip');
  private readonly retryEl = el<HTMLButtonElement>('boot-retry');
  private readonly runner = el('boot-runner');
  private tipTimer = 0;
  private tipIdx = Math.floor(Math.random() * STR.loadTips.length);
  private runTimer = 0;
  /** runFrames: khung chạy cắt từ atlas, dùng `background-position`. Rỗng = chưa có atlas. */
  private runFrames: { x: number; y: number; w: number; h: number }[] = [];
  private runSheet = '';
  private runSheetSize = { w: 0, h: 0 };

  constructor() {
    this.cycleTip();
    this.tipTimer = window.setInterval(() => this.cycleTip(), TIP_MS);
  }

  private cycleTip(): void {
    if (!this.tipEl) {
      return;
    }
    this.tipIdx = (this.tipIdx + 1) % STR.loadTips.length;
    this.tipEl.style.opacity = '0';
    window.setTimeout(() => {
      if (this.tipEl) {
        this.tipEl.textContent = STR.loadTips[this.tipIdx] ?? '';
        this.tipEl.style.opacity = '1';
      }
    }, 220);
  }

  step(text: string): void {
    this.root?.classList.remove('failed');
    if (this.stepEl) {
      this.stepEl.textContent = text;
    }
  }

  /** bytes vẽ thanh theo BYTE thật (§13.7: không dùng % theo số tệp). */
  bytes(p: LoadProgress): void {
    const pct = p.total > 0 ? p.loaded / p.total : 1;
    if (this.fill) {
      this.fill.style.width = `${(pct * 100).toFixed(1)}%`;
    }
    if (this.runner) {
      this.runner.style.left = `${(pct * 100).toFixed(1)}%`;
    }
    if (this.bytesEl) {
      // Nhóm rỗng (lần vào thứ hai, trình duyệt đã cache hết) thì đừng hiện "0,0 / 0,0 MB"
      // — nhìn như lỗi. §13.7 nói lần sau "thanh chạy vọt lên gần như ngay lập tức".
      this.bytesEl.textContent = p.total > 0 ? `${Math.round(pct * 100)}%  ·  ${STR.loadMB(p.loaded / 1e6, p.total / 1e6)}` : STR.loadCached;
    }
  }

  /**
   * setRunner cho nhân vật người chơi chọn chạy trên thanh.
   *
   * Hai mức, theo đúng thứ tự asset về: chân dung (nhóm `home`, có sẵn lúc này) nhún lên
   * nhún xuống; rồi khi atlas của nhân vật về thì đổi sang **khung `run` thật**, cắt bằng
   * `background-position` — không cần PixiJS, nên dùng được ngay trên màn hình tải.
   */
  setRunner(char: string): void {
    if (!this.runner || !char) {
      return;
    }
    const portrait = assets.url(`characters/${char}/portrait_neutral.png`);
    if (portrait) {
      this.runner.style.backgroundImage = `url(${portrait})`;
    }
    void this.upgradeRunner(char);
  }

  private async upgradeRunner(char: string): Promise<void> {
    const jsonUrl = assets.url(`characters/${char}/atlas.json`);
    if (!jsonUrl || !this.runner) {
      return;
    }
    try {
      const atlas = (await (await fetch(jsonUrl)).json()) as {
        image?: string;
        frames: Record<string, { x: number; y: number; w: number; h: number }>;
        anims: Record<string, { frames: string[] }>;
      };
      const names = atlas.anims.run?.frames ?? atlas.anims.idle?.frames ?? [];
      const frames = names.map((n) => atlas.frames[n]).filter(Boolean);
      if (!frames.length || !atlas.image) {
        return;
      }
      const sheet = `/assets/characters/${char}/${atlas.image}`;
      // Cần kích thước thật của sheet để đặt `background-size` theo tỉ lệ ô/sheet.
      const size = await new Promise<{ w: number; h: number }>((resolve, reject) => {
        const img = new Image();
        img.onload = () => resolve({ w: img.naturalWidth, h: img.naturalHeight });
        img.onerror = reject;
        img.src = sheet;
      });
      this.runFrames = frames;
      this.runSheet = sheet;
      this.runSheetSize = size;
      this.runner.classList.add('frames');
      let i = 0;
      this.drawRunFrame(0);
      window.clearInterval(this.runTimer);
      this.runTimer = window.setInterval(() => {
        i = (i + 1) % this.runFrames.length;
        this.drawRunFrame(i);
      }, 1000 / RUN_FPS);
    } catch {
      // Không đọc được atlas thì cứ để chân dung nhún — đây là phần trang trí, không
      // được phép làm hỏng việc tải.
    }
  }

  private drawRunFrame(i: number): void {
    const f = this.runFrames[i];
    if (!f || !this.runner) {
      return;
    }
    const box = 56; // khớp `#boot .runner` trong index.html
    const k = box / Math.max(f.w, f.h);
    this.runner.style.backgroundImage = `url(${this.runSheet})`;
    this.runner.style.backgroundSize = `${this.runSheetSize.w * k}px ${this.runSheetSize.h * k}px`;
    this.runner.style.backgroundPosition = `${-f.x * k}px ${-f.y * k}px`;
    this.runner.style.backgroundRepeat = 'no-repeat';
  }

  /** fail hiện nút "Thử lại"; `onRetry` chạy khi bấm. Tải TIẾP, không tải lại từ đầu. */
  fail(onRetry: () => void): void {
    this.root?.classList.add('failed');
    if (this.stepEl) {
      this.stepEl.textContent = STR.loadFailed;
    }
    if (this.retryEl) {
      this.retryEl.onclick = () => {
        this.root?.classList.remove('failed');
        onRetry();
      };
    }
  }

  /**
   * done ẩn màn hình tải (mờ ra 0,25 s theo CSS).
   *
   * ẨN chứ không tháo khỏi DOM: cùng một màn hình này còn phải dùng lại ở bước "Đang gọi
   * nhân vật…" và "Đang dựng {khu}…" sau khi người chơi bấm Vào chơi. Dựng lại khung HTML
   * bằng JS mỗi lần là tự tay bỏ đi lý do nó nằm trong `index.html`.
   */
  done(): void {
    window.clearInterval(this.tipTimer);
    window.clearInterval(this.runTimer);
    this.root?.classList.add('gone');
    if (this.root) {
      // `pointer-events: none` của `.gone` là đủ cho chuột, nhưng không đủ cho bộ đọc màn
      // hình và phím Tab — nút "Thử lại" vẫn focus được qua một màn hình vô hình.
      this.root.setAttribute('aria-hidden', 'true');
    }
  }

  /** show bật lại màn hình tải cho bước tiếp theo, mẹo và nhân vật chạy tiếp tục. */
  show(): void {
    this.root?.classList.remove('gone', 'failed');
    this.root?.removeAttribute('aria-hidden');
    window.clearInterval(this.tipTimer);
    this.tipTimer = window.setInterval(() => this.cycleTip(), TIP_MS);
    if (this.runFrames.length) {
      let i = 0;
      window.clearInterval(this.runTimer);
      this.runTimer = window.setInterval(() => {
        i = (i + 1) % this.runFrames.length;
        this.drawRunFrame(i);
      }, 1000 / RUN_FPS);
    }
  }
}

/**
 * bgLoadIndicator: icon % ở góc dưới cho các nhóm đang tải NỀN (§13.7 mục 4, dòng cuối).
 *
 * Tự ẩn khi xong. Đọc trực tiếp `assets.bgProgress()` mỗi 250 ms thay vì đăng ký lắng nghe
 * — một `setInterval` rẻ hơn và không có gì để rò rỉ.
 */
export function startBgLoadIndicator(): void {
  const box = el('bgload');
  if (!box) {
    return;
  }
  window.setInterval(() => {
    const { pct, running } = assets.bgProgress();
    box.style.display = running ? 'block' : 'none';
    if (running) {
      box.textContent = STR.loadBg(pct);
    }
  }, 250);
}
