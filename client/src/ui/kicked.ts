import { button, h } from './dom';
import { STR } from './strings';

/**
 * Màn hình "bị mời ra khỏi phòng" / "đang bị cấm" — ADMIN_SPEC §6.2.
 *
 * Luôn nói **vì sao**, và với lệnh cấm thì nói cả **tới bao giờ**. Một người bị mời ra mà
 * không biết lý do sẽ vào lại và làm đúng việc vừa rồi — §6.2 nói rõ kick thì "vào lại được
 * ngay", nên màn hình không kèm lý do chẳng ngăn được gì, chỉ tạo ra một người giận.
 *
 * Lý do là chữ do admin tự gõ. `h` ghép chuỗi bằng `append`, tạo ra text node — không bao
 * giờ qua innerHTML.
 */
export function kickedScreen(reason: string, until: number | undefined, onBack: () => void): HTMLElement {
  // `until` thiếu hoặc 0 = chỉ kick. -1 = cấm vĩnh viễn. Còn lại = mốc hết hạn.
  const banned = until !== undefined && until !== 0;
  const root = h('div', 'screen center');
  const panel = h('div', 'panel');

  panel.append(h('h2', 'kick-title', banned ? STR.bannedTitle : STR.kickedTitle));
  if (reason) {
    panel.append(h('p', 'kick-reason', STR.kickedReason + reason));
  }
  if (!banned) {
    panel.append(h('p', 'hint', STR.kickedRejoin));
  } else if (until === -1) {
    panel.append(h('p', 'hint', STR.bannedForever));
  } else {
    panel.append(h('p', 'hint', STR.bannedUntil + new Date(until * 1000).toLocaleString('vi-VN')));
  }

  const row = h('div', 'row');
  row.append(button(STR.kickedBack, 'btn primary', onBack));
  panel.append(row);

  root.append(panel);
  return root;
}
