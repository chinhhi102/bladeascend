import { gpuBudgetMB, gpuName, isIntegratedGpu, isMobile } from './device.ts';

/**
 * Mức chất lượng đồ họa (spec §3.6): Cao / Vừa / Thấp. Người chơi đổi bằng phím V, lựa chọn
 * lưu trong localStorage.
 *
 * Thần thú giờ luôn bật (spec v4.7 §3.5) nên số đốt phải co theo máy, không thể để cố định.
 */

export type Quality = 'high' | 'medium' | 'low';

/**
 * KEY chỉ lưu lựa chọn TAY của người chơi (phím V).
 *
 * Khoá cũ `kx.quality` bị bỏ có chủ đích: nó lưu cả những lần MÁY tự hạ/nâng, nên một lần
 * tự nâng lên Cao trên máy khoẻ thì lần sau mở game ở máy đó vẫn bắt đầu bằng Cao — đúng cái
 * mặc định mà chủ dự án chốt 1/11 muốn bỏ. Đọc lại khoá cũ là kéo theo mọi lựa chọn máy đã
 * làm thay người.
 */
const KEY = 'kx.quality.manual';
const ORDER: Quality[] = ['high', 'medium', 'low'];

export const QUALITY_NAME: Record<Quality, string> = { high: 'Cao', medium: 'Vừa', low: 'Thấp' };

/** Số đốt thần thú theo chất lượng (spec §3.5: Cao 16 · Vừa 10 · Thấp 6). */
export const BEAST_SEGMENTS: Record<Quality, number> = { high: 16, medium: 10, low: 6 };

/** Tỉ lệ hạt vệt theo sau (Thấp = tắt hẳn). */
export const TRAIL_RATE: Record<Quality, number> = { high: 1, medium: 0.5, low: 0 };

/**
 * defaultQuality: mức bắt đầu khi người chơi chưa tự chọn (chủ dự án chốt 1/11).
 *
 * Điện thoại và GPU tích hợp → Thấp; PC còn lại → Vừa. **Không ai bắt đầu ở Cao**: telemetry
 * 31/10 đo Cao vẽ p95 64,9 ms so với Vừa 18,2 ms — mở game ở Cao là bắt người chơi chịu
 * giật trong 3 giây đầu cho tới khi `autoQuality` kịp hạ. Cao chỉ tới bằng `autoQuality`
 * (≥ 55 FPS suốt 20 giây) hoặc bằng phím V.
 */
export function defaultQuality(mobile: boolean, gpu: string): Quality {
  return mobile || isIntegratedGpu(gpu) ? 'low' : 'medium';
}

let manual = false;
let current: Quality = read();

function read(): Quality {
  let v: string | null = null;
  try {
    v = localStorage.getItem(KEY);
  } catch {
    // Chế độ riêng tư / chặn lưu trữ: coi như chưa chọn.
  }
  if (ORDER.includes(v as Quality)) {
    manual = true;
    return v as Quality;
  }
  return typeof document === 'undefined' ? 'medium' : defaultQuality(isMobile(), gpuName());
}

export function quality(): Quality {
  return current;
}

/**
 * QualityFeature: những thứ khác nhau giữa các mức chất lượng, gọi tên được từng cái.
 *
 * Tách ra để ĐO: "Cao tốn gần gấp đôi Vừa" (telemetry 31/10) là con số gộp, không nói được
 * phần nào nằm ở đâu. `?qf=bloom:medium` hạ đúng MỘT thứ xuống mức khác trong khi phần còn
 * lại giữ nguyên — `tools/browsertest/qualitycost.mjs` dùng nó để tắt từng thứ mà đo.
 */
export type QualityFeature = 'bloom' | 'ambient' | 'decals' | 'beast' | 'motes' | 'boba';

const FEATURES: readonly QualityFeature[] = ['bloom', 'ambient', 'decals', 'beast', 'motes', 'boba'];

/** overrides đọc `?qf=tên:mức,...` một lần lúc nạp. Không có tham số thì rỗng. */
const overrides: ReadonlyMap<QualityFeature, Quality> = readOverrides();

function readOverrides(): Map<QualityFeature, Quality> {
  const m = new Map<QualityFeature, Quality>();
  const raw = typeof location === 'undefined' ? null : new URLSearchParams(location.search).get('qf');
  for (const part of (raw ?? '').split(',')) {
    const [f, q] = part.split(':');
    if (FEATURES.includes(f as QualityFeature) && ORDER.includes(q as Quality)) {
      m.set(f as QualityFeature, q as Quality);
    }
  }
  return m;
}

/** qualityFor là mức chất lượng của MỘT thứ — bằng `quality()` trừ khi URL ghi đè. */
export function qualityFor(f: QualityFeature): Quality {
  return overrides.get(f) ?? current;
}

/** setQuality đổi mức trong phiên này, KHÔNG lưu — máy tự đổi thì không được nhớ thay người. */
export function setQuality(q: Quality): void {
  current = q;
  // Đổi mức là bắt đầu đếm lại. Thiếu hai dòng này thì `autoQualityByMemory` vừa hạ vì
  // thiếu bộ nhớ, khung sau `autoQuality` thấy "đã ổn từ lâu" và nâng ngược lên ngay —
  // `check:quality` bắt được đúng chuyện đó.
  lowSince = 0;
  okSince = 0;
}

/** cycleQuality đổi sang mức tiếp theo và trả về mức mới. Người chơi tự chọn thì khoá lại và lưu. */
export function cycleQuality(): Quality {
  manual = true;
  setQuality(ORDER[(ORDER.indexOf(current) + 1) % ORDER.length]);
  try {
    localStorage.setItem(KEY, current);
  } catch {
    // Không lưu được thì lựa chọn chỉ sống trong phiên này — vẫn đúng ý người chơi lúc này.
  }
  return current;
}

/**
 * Tự hạ chất lượng (spec §13.6 mục 8): FPS dưới 45 suốt 3 giây thì hạ một bậc; ổn lại trên
 * 55 suốt 10 giây thì nâng lại một bậc — RIÊNG bậc lên Cao phải ổn ≥ 55 suốt 20 giây
 * (chủ dự án chốt 1/11), vì Cao tốn gần gấp đôi Vừa và nâng sớm là tụt lại ngay.
 *
 * Không đụng tới lựa chọn của người chơi: ai đã tự bấm phím V thì giữ nguyên mức họ chọn —
 * máy tự ý đổi ngược ý người dùng là thứ gây ức chế nhất.
 */
const UP_SECS = 10;
const UP_TO_HIGH_SECS = 20;
let lowSince = 0;
let okSince = 0;

/**
 * gpuPressure: bộ nhớ texture đã vượt ngân sách thiết bị chưa (NETCODE_SPEC §5 L1).
 *
 * Trả về mức 0 (ổn) · 1 (sát trần, hạ một bậc) · 2 (vượt trần, xuống Thấp ngay).
 */
export function gpuPressure(texMB: number): 0 | 1 | 2 {
  const budget = gpuBudgetMB();
  if (texMB >= budget) {
    return 2;
  }
  return texMB >= budget * 0.8 ? 1 : 0;
}

/**
 * autoQualityByMemory hạ chất lượng khi bộ nhớ GPU chạm ngân sách — **kể cả khi FPS vẫn đẹp**.
 *
 * Đây là điểm khác hẳn `autoQuality`: trên iOS, hết bộ nhớ đồ hoạ KHÔNG làm tụt khung hình
 * trước. Máy chạy 60 fps tới đúng khoảnh khắc Safari giết tab. Nên nhìn FPS là nhìn sai chỗ;
 * phải nhìn chính con số bộ nhớ.
 *
 * Và hạ chất lượng ở đây **bỏ qua cả lựa chọn tay của người chơi** (`manual`): họ chọn mức
 * nào là quyền của họ, nhưng không ai chọn "cho tab chết".
 */
export function autoQualityByMemory(texMB: number): boolean {
  const level = gpuPressure(texMB);
  if (level === 0) {
    return false;
  }
  const want: Quality = level === 2 ? 'low' : ORDER[Math.min(ORDER.indexOf(current) + 1, ORDER.length - 1)];
  if (want === current) {
    return false;
  }
  setQuality(want);
  return true;
}

export function autoQuality(fps: number, now: number): void {
  if (manual || fps <= 0) {
    return;
  }
  if (fps < 45) {
    okSince = 0;
    lowSince = lowSince || now;
    if (now - lowSince >= 3 && current !== 'low') {
      setQuality(ORDER[ORDER.indexOf(current) + 1]);
      lowSince = now;
    }
    return;
  }
  lowSince = 0;
  if (fps >= 55) {
    okSince = okSince || now;
    const need = current === 'medium' ? UP_TO_HIGH_SECS : UP_SECS;
    if (now - okSince >= need && current !== 'high') {
      setQuality(ORDER[ORDER.indexOf(current) - 1]);
      okSince = now;
    }
  } else {
    okSince = 0;
  }
}
