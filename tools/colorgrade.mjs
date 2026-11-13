// MA TRẬN CHỈNH MÀU NỀN — một nguồn sự thật cho cả client lẫn asset-pipeline.
//
// ART_PROMPTS §9.0: mặt đất phải dịu và ít tương phản để vật thể gameplay nổi lên. Phép đó
// vốn là một `ColorMatrixFilter` gắn trên container — tức một đích vẽ toàn màn hình mỗi
// khung. GAME_SPEC §13.10 bỏ nó đi, và phép màu chuyển sang hai chỗ:
//
//   - **Lớp nền**: tính thẳng trong shader trộn (`render/groundshader.ts`).
//   - **Decal và trang trí**: áp vào CHÍNH ẢNH lúc build (`tools/asset-pipeline`).
//
// Hai chỗ đó phải dùng ĐÚNG một ma trận. Nếu không, mặt đất và mấy cụm cỏ trên nó sẽ lệch
// màu nhau — loại lỗi không ai gọi tên được, chỉ thấy "nhìn sai sai".
//
// Tệp này viết bằng JS thuần vì asset-pipeline chạy trong Node và không có PixiJS. Phép
// toán chép từ `ColorMatrixFilter` của Pixi, và `client/scripts/colorgrade-check.ts` so hai
// bên với nhau trong mỗi lần build — chép tay mà không có bài canh thì sớm muộn cũng trôi.

/** Tham số chỉnh màu nền. Đổi ở ĐÂY, cả client lẫn pipeline theo cùng lúc. */
export const GROUND_GRADE = { saturate: -0.5, brightness: 0.9, contrast: -0.2 };

/**
 * nhân hai ma trận 4×5 theo đúng `ColorMatrixFilter._multiply` của Pixi.
 *
 * Thứ tự: `multiply(cũ, mới)` — Pixi gọi `_multiply(out, this.matrix, matrix)`, tức ma
 * trận ĐANG CÓ đứng trước, phép mới đứng sau. Tôi viết ngược lần đầu và `check:colorgrade`
 * bắt được ngay: offset ra 0,10 thay vì 0,09 (thiếu một lần nhân với brightness 0,9). Phép
 * nhân ma trận không giao hoán, và chỗ sai này chỉ hiện ra thành "màu hơi lệch".
 */
function multiply(a, b) {
  const out = new Array(20).fill(0);
  for (let row = 0; row < 4; row++) {
    const r = row * 5;
    for (let col = 0; col < 4; col++) {
      out[r + col] = a[r] * b[col] + a[r + 1] * b[col + 5] + a[r + 2] * b[col + 10] + a[r + 3] * b[col + 15];
    }
    out[r + 4] = a[r] * b[4] + a[r + 1] * b[9] + a[r + 2] * b[14] + a[r + 3] * b[19] + a[r + 4];
  }
  return out;
}

const IDENTITY = [1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 1, 0];

function saturateMatrix(amount) {
  const x = (amount * 2) / 3 + 1;
  const y = (x - 1) * -0.5;
  return [x, y, y, 0, 0, y, x, y, 0, 0, y, y, x, 0, 0, 0, 0, 0, 1, 0];
}

function brightnessMatrix(b) {
  return [b, 0, 0, 0, 0, 0, b, 0, 0, 0, 0, 0, b, 0, 0, 0, 0, 0, 1, 0];
}

function contrastMatrix(amount) {
  const v = (amount || 0) + 1;
  const o = -0.5 * (v - 1);
  return [v, 0, 0, 0, o, 0, v, 0, 0, o, 0, 0, v, 0, o, 0, 0, 0, 1, 0];
}

/**
 * groundMatrix trả ma trận 4×5 của phép chỉnh màu nền.
 *
 * Thứ tự áp PHẢI giống bản cũ: `saturate` rồi `brightness(multiply)` rồi
 * `contrast(multiply)`. Nhân ma trận không giao hoán — đổi thứ tự là đổi màu.
 */
export function groundMatrix() {
  let m = saturateMatrix(GROUND_GRADE.saturate);
  m = multiply(m, brightnessMatrix(GROUND_GRADE.brightness));
  m = multiply(m, contrastMatrix(GROUND_GRADE.contrast));
  return m;
}

/**
 * sharpGrade đổi ma trận 4×5 thành hai bước mà `sharp` làm được: `recomb` (3×3) rồi
 * `linear` (nhân 1, cộng offset theo kênh).
 *
 * `sharp.recomb` không có phần cộng, còn ma trận của ta có (phép `contrast` sinh ra nó), nên
 * phải tách. Offset của Pixi tính trên thang 0–1; `sharp` làm việc trên thang 0–255.
 */
export function sharpGrade() {
  const m = groundMatrix();
  return {
    recomb: [
      [m[0], m[1], m[2]],
      [m[5], m[6], m[7]],
      [m[10], m[11], m[12]],
    ],
    offset255: [m[4] * 255, m[9] * 255, m[14] * 255],
  };
}
