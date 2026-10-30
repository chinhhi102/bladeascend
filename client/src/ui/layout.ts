/**
 * Mốc "màn hẹp" dùng chung cho toàn HUD.
 *
 * Phải có một chỗ duy nhất: `hud.ts` cắt số dòng, `minimap.ts` quyết định có thu nhỏ hay
 * không, và CSS có media query riêng. Ba nơi mà ba con số thì sẽ có khổ màn hình rơi vào
 * khe giữa chúng — ví dụ bảng xếp hạng đã co lại nhưng minimap vẫn mở to.
 *
 * Xét **cả hai chiều**, không chỉ bề ngang: điện thoại xoay ngang là 800×360, rộng hơn mốc
 * 720 nhưng chật theo chiều cao. Bản chỉ xét bề ngang từng để HUD phủ 70% màn hình ở đúng
 * khổ đó (đo 30/10).
 *
 * ⚠ Đổi hai số này thì phải đổi cả media query trong `style.css`
 * (`@media (max-width: 720px), (max-height: 480px)`) cho khớp.
 */
export const NARROW_W = 720;
export const SHORT_H = 480;

export function narrowScreen(): boolean {
  return window.innerWidth <= NARROW_W || window.innerHeight <= SHORT_H;
}
