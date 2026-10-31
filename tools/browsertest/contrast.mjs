// Đo độ tương phản chữ/nền THẬT trên trang đang chạy (WCAG 2.1).
//
//   import { dimText } from './contrast.mjs';
//   const bad = await dimText(page, '#intro');
//
// Trả về mảng mô tả những phần tử có chữ mờ hơn ngưỡng. Mảng rỗng = đạt.
//
// ── Vì sao có tệp này ──────────────────────────────────────────────────────────────────
//
// Hai lần trong một tuần tôi làm ra chữ không đọc được, và cả hai lần đều cùng một cách:
// chép một màu từ chỗ khác sang chỗ có bảng màu NGƯỢC LẠI.
//
//   · Form góp ý: `color: var(--ink, #e8eaed)` chép từ trang quản trị, ở đó `--ink` sáng
//     màu. Trong game `--ink` là `#1d1a16`. Chữ đen trên nền đen, 1,04:1.
//   · Đoạn giới thiệu §20: `color: var(--ink)` trên nền trang `#2b2b2b`. 1,1:1.
//
// Cả hai lần tôi đều đã NHÌN ảnh chụp và không thấy — chữ đen trên nền đen trông hệt như
// một khoảng trống, và mắt đọc khoảng trống là "chỗ này chưa có gì". Con số thì không bỏ
// sót. Nên phép kiểm này phải chạy ở mọi khung mới, không phải chỉ ở khung đã từng hỏng.
export async function dimText(page, selector, min = 3) {
  return page.evaluate(
    ({ selector, min }) => {
      const lum = (c) => {
        const [r, g, b] = c.map((v) => {
          const s = v / 255;
          return s <= 0.03928 ? s / 12.92 : ((s + 0.055) / 1.055) ** 2.4;
        });
        return 0.2126 * r + 0.7152 * g + 0.0722 * b;
      };
      const parse = (s) => (s.match(/[\d.]+/g) ?? []).slice(0, 3).map(Number);

      // Nền THẬT = nền của tổ tiên gần nhất không trong suốt.
      //
      // Không lấy `backgroundColor` của chính phần tử: gần như mọi thẻ chữ đều trong suốt,
      // và `rgba(0,0,0,0)` quy ra đen — nên mọi chữ sáng sẽ "đạt" và mọi chữ tối sẽ "hỏng",
      // bất kể trang thật trông thế nào.
      const bgOf = (el) => {
        for (let n = el; n; n = n.parentElement) {
          const bg = getComputedStyle(n).backgroundColor;
          const p = parse(bg);
          if (p.length >= 3 && !/rgba\(.*,\s*0\s*\)/.test(bg)) return p;
        }
        return [255, 255, 255];
      };

      const root = document.querySelector(selector);
      if (!root) return [`KHÔNG TÌM THẤY ${selector}`];
      const bad = [];
      for (const el of [root, ...root.querySelectorAll('*')]) {
        // Chỉ xét phần tử có chữ TRỰC TIẾP: bỏ qua thì một thẻ bọc sẽ bị tính theo chữ của
        // con nó và báo trùng.
        const own = [...el.childNodes].some((n) => n.nodeType === 3 && n.textContent.trim());
        if (!own) continue;
        const cs = getComputedStyle(el);
        if (cs.visibility === 'hidden' || cs.display === 'none' || Number(cs.opacity) < 0.1) continue;
        const r = el.getBoundingClientRect();
        if (r.width < 1 || r.height < 1) continue;
        const fg = parse(cs.color);
        if (fg.length < 3) continue;
        const a = lum(fg) + 0.05;
        const b = lum(bgOf(el)) + 0.05;
        const ratio = a > b ? a / b : b / a;
        if (ratio < min) {
          bad.push(`${el.tagName.toLowerCase()}.${el.className || '-'} = ${ratio.toFixed(2)}:1`);
        }
      }
      return bad;
    },
    { selector, min },
  );
}
