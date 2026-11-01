/**
 * Nhận dạng thiết bị và ngân sách bộ nhớ GPU (NETCODE_SPEC §5 L1).
 *
 * Vì sao cần riêng một tệp: iOS **không báo lỗi khi hết bộ nhớ GPU** — Safari giết hẳn tab.
 * Không có ngoại lệ để bắt, không có dòng log nào. Nên cách duy nhất là tự đặt trần rồi tự
 * hạ chất lượng TRƯỚC khi chạm trần.
 *
 * Đo ngày 31/10 trước khi sửa: 1,4 GB texture lúc vừa vào trận, 1,9 GB khi lên Đạo Tổ — gấp
 * 7 lần ngân sách của một iPhone. Khớp với báo cáo "trên iPhone luôn bị lỗi".
 */

/** isMobile: điện thoại / máy tính bảng. Dùng cả cảm ứng lẫn chuỗi UA vì iPad giả làm Mac. */
export function isMobile(): boolean {
  const touch = 'ontouchstart' in window || navigator.maxTouchPoints > 1;
  const ua = navigator.userAgent;
  return touch && (/Android|iPhone|iPad|iPod|Mobile/i.test(ua) || /Mac/.test(ua));
}

/**
 * isIOS: iPhone/iPad.
 *
 * iPadOS 13+ khai UA là "Macintosh", nên phải xét thêm `maxTouchPoints` — một chiếc Mac
 * thật không có màn cảm ứng. Thiếu vế đó là iPad rơi vào nhánh máy bàn và ăn trọn ngân sách
 * của máy bàn, đúng cái đang làm tab bị giết.
 */
export function isIOS(): boolean {
  const ua = navigator.userAgent;
  return /iPhone|iPad|iPod/.test(ua) || (/Mac/.test(ua) && navigator.maxTouchPoints > 1);
}

/**
 * gpuBudgetMB là trần bộ nhớ texture ước tính cho thiết bị này.
 *
 * iOS 256 MB: Safari trên iPhone thường bị hệ điều hành thu hồi tab quanh mức vài trăm MB
 * bộ nhớ đồ hoạ, và con số đó không công bố ở đâu cả. 256 là mức chủ dự án chốt 31/10 —
 * thấp hơn hẳn mọi ngưỡng quan sát được, để còn chỗ cho chính trình duyệt.
 */
export function gpuBudgetMB(): number {
  if (isIOS()) {
    return 256;
  }
  return isMobile() ? 384 : 1024;
}

/**
 * rendererResolution là `resolution` truyền cho PixiJS — số điểm ảnh thật trên mỗi điểm ảnh CSS.
 *
 * iPhone có `devicePixelRatio` 3: để nguyên thì mọi render texture (kể cả bộ đệm khung hình)
 * tốn **9 lần** so với DPR 1, và 2,25 lần so với mức 2. Chặn ở 2 là gần như không thấy khác
 * trên màn hình nhỏ; mức **Thấp chặn ở 1.5** để cứu những máy yếu nhất.
 */
export function rendererResolution(low: boolean): number {
  const dpr = window.devicePixelRatio || 1;
  return Math.min(dpr, low ? 1.5 : 2);
}

let gpuCache: string | null = null;

/**
 * gpuName là tên GPU (WEBGL_debug_renderer_info), rỗng nếu trình duyệt không cho biết.
 *
 * Hỏi trên một canvas RIÊNG rồi vứt ngay — giống `telemetry/session.ts` — vì phải biết TRƯỚC
 * khi Pixi dựng ngữ cảnh của game (mức chất lượng quyết định cả `resolution`).
 */
export function gpuName(): string {
  if (gpuCache !== null) {
    return gpuCache;
  }
  gpuCache = '';
  const cv = document.createElement('canvas');
  const gl = (cv.getContext('webgl2') ?? cv.getContext('webgl')) as WebGLRenderingContext | null;
  if (!gl) {
    return gpuCache;
  }
  const ext = gl.getExtension('WEBGL_debug_renderer_info');
  if (ext) {
    gpuCache = String(gl.getParameter(ext.UNMASKED_RENDERER_WEBGL) ?? '');
  }
  gl.getExtension('WEBGL_lose_context')?.loseContext();
  return gpuCache;
}

/**
 * isIntegratedGpu: GPU tích hợp / dựng hình bằng phần mềm.
 *
 * Cùng luật với `deviceClass` phía server (`internal/telemetry/report.go`), để nhóm "PC GPU
 * tích hợp" trong báo cáo và nhóm được mặc định mức Thấp là MỘT nhóm. Sửa một bên thì sửa
 * cả bên kia.
 */
export function isIntegratedGpu(gpu: string): boolean {
  const g = gpu.toLowerCase();
  return (
    g.includes('intel') || g.includes('uhd') || g.includes('iris') ||
    (g.includes('vega') && g.includes('graphics')) ||
    g.includes('swiftshader') || g.includes('llvmpipe')
  );
}
