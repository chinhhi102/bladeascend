// Ô chọn ngôn ngữ ở trang chủ (GAME_SPEC §19): "Tự động" + 6 ngôn ngữ đợt 1, kèm cờ.
//
// Dùng `<select>` thật chứ không tự vẽ danh sách: trên điện thoại nó mở bộ chọn của hệ
// điều hành (to, dễ bấm, đọc được bằng trình đọc màn hình), và không cần JS để đóng/mở.
// Đổi ngôn ngữ thì tải lại trang — chữ đã dựng bằng ngôn ngữ cũ không bị lẫn.
import { chooseLang, LANG, LANGS, savedLang } from '../i18n';
import { STR } from './strings';

export function langPicker(): HTMLElement {
  const sel = document.createElement('select');
  sel.className = 'lang-pick';
  sel.setAttribute('aria-label', STR.lang.label);
  const auto = document.createElement('option');
  auto.value = '';
  const cur = LANGS.find((l) => l.code === LANG);
  auto.textContent = `🌐 ${STR.lang.auto}${cur ? ` (${cur.flag})` : ''}`;
  sel.append(auto);
  for (const l of LANGS) {
    const o = document.createElement('option');
    o.value = l.code;
    o.textContent = `${l.flag} ${l.name}`;
    o.lang = l.code;
    sel.append(o);
  }
  // Đang để tự động thì ô hiện "Tự động"; đã chọn tay (hoặc ?lang=) thì hiện ngôn ngữ đó.
  const url = new URLSearchParams(location.search).get('lang');
  sel.value = url || savedLang() ? LANG : '';
  sel.addEventListener('change', () => chooseLang(sel.value));
  return sel;
}
