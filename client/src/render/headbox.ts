// SINH TỰ ĐỘNG bằng `node tools/portrait-headbox.mjs` — đừng sửa tay.
// Sửa số thì sửa bảng OVERRIDE trong tool rồi chạy lại, nếu không lần chạy sau mất hết.

/**
 * Khung đầu của từng chân dung: `[trái, trên, cạnh]` theo TỈ LỆ cạnh ảnh (0–1), luôn vuông.
 *
 * Dùng cho mọi chỗ hiện MẶT nhân vật trong ô nhỏ: ô đại diện HUD (§12.4), bong bóng mặt
 * (§12.5), chân dung tròn ở bảng xếp hạng (§11.5), ảnh nhân vật ở màn chọn. Chân dung là
 * ảnh bán thân/toàn thân nên cắt bằng một luật chung ("lấy nửa trên, phóng 180%") là ra
 * cái nón với hai con mắt, còn thu cả tấm ảnh vào ô 28–40 px thì mặt bé bằng hạt gạo —
 * mỗi ảnh phải có khung riêng.
 */
export const HEAD_BOX: Record<string, readonly [number, number, number]> = {
  'a_hon/neutral': [0.121, 0.012, 0.734],
  'a_hon/evil': [0.135, 0.012, 0.721],
  'a_hon/panic': [0, 0, 1],
  'a_hon/cry': [0.092, 0.006, 0.818],
  'ba_tam_loa/neutral': [0.391, 0.182, 0.529],
  'ba_tam_loa/evil': [0.18, 0.068, 0.514],
  'ba_tam_loa/panic': [0.045, 0.141, 0.623],
  'ba_tam_loa/cry': [0.16, 0.023, 0.508],
  'be_ti/neutral': [0.252, 0.172, 0.561],
  'be_ti/evil': [0.229, 0.166, 0.627],
  'be_ti/panic': [0.205, 0.186, 0.59],
  'be_ti/cry': [0.344, 0.162, 0.463],
  'co_ba_ganh/neutral': [0.3, 0.03, 0.42], // đo bằng mắt
  'co_ba_ganh/evil': [0, 0, 1],
  'co_ba_ganh/panic': [0, 0, 1],
  'co_ba_ganh/cry': [0, 0, 1],
  'cu_rua/neutral': [0.25, 0.137, 0.635],
  'cu_rua/evil': [0.199, 0.041, 0.729],
  'cu_rua/panic': [0.197, 0.043, 0.68],
  'cu_rua/cry': [0.23, 0.201, 0.559],
  'dai_ca_dep/neutral': [0.256, 0.082, 0.564],
  'dai_ca_dep/evil': [0.295, 0.148, 0.547],
  'dai_ca_dep/panic': [0.148, 0.012, 0.734],
  'dai_ca_dep/cry': [0.105, 0.09, 0.875],
  'map_tra_sua/neutral': [0.188, 0.018, 0.627],
  'map_tra_sua/evil': [0.232, 0.02, 0.594],
  'map_tra_sua/panic': [0.086, 0.012, 0.734],
  'map_tra_sua/cry': [0.176, 0.021, 0.566],
  'ong_ba_vong/neutral': [0.258, 0.08, 0.691],
  'ong_ba_vong/evil': [0.33, 0.09, 0.36], // đo bằng mắt
  'ong_ba_vong/panic': [0.332, 0.139, 0.576],
  'ong_ba_vong/cry': [0.209, 0.025, 0.691],
  'shipper/neutral': [0.215, 0.01, 0.768],
  'shipper/evil': [0.45, 0.17, 0.4], // đo bằng mắt
  'shipper/panic': [0.314, 0.18, 0.547],
  'shipper/cry': [0.244, 0.137, 0.602],
  'su_phu_mi/neutral': [0.123, 0.184, 0.691],
  'su_phu_mi/evil': [0.31, 0.09, 0.42], // đo bằng mắt
  'su_phu_mi/panic': [0.037, 0.037, 0.963],
  'su_phu_mi/cry': [0.154, 0.082, 0.779],
};

/**
 * Khung mặc định cho ảnh chưa có trong bảng (nhân vật mới, art vừa thả vào). Lấy nửa trên
 * và hơi hẹp hai bên — không chuẩn bằng số đo riêng nhưng còn thấy được mặt.
 */
export const DEFAULT_HEAD_BOX: readonly [number, number, number] = [0.2, 0.02, 0.6];

/** headBox tra khung đầu theo nhân vật + biểu cảm. */
export function headBox(char: string, mood: string): readonly [number, number, number] {
  return HEAD_BOX[`${char}/${mood}`] ?? HEAD_BOX[`${char}/neutral`] ?? DEFAULT_HEAD_BOX;
}

/**
 * headCrop gắn ảnh chân dung vào một thẻ HTML sao cho khung đầu vừa đúng cả ô.
 *
 * Mọi chỗ hiện mặt trong ô nhỏ đều gọi hàm này, không ai tự chép công thức: ô phải VUÔNG
 * (khung đầu vuông) và thẻ phải có `overflow: hidden` nếu ô bo tròn.
 *
 * Công thức phần trăm của `background-position`: 0% là mép trái ảnh trùng mép trái ô,
 * 100% là mép phải trùng mép phải — nên toạ độ khung đầu phải chia cho phần ảnh THỪA ra
 * ngoài ô (1 - cạnh), không phải nhân thẳng với 100.
 */
export function headCrop(el: HTMLElement, url: string, char: string, mood = 'neutral'): void {
  const [bx, by, bs] = headBox(char, mood);
  el.style.backgroundRepeat = 'no-repeat';
  el.style.backgroundSize = `${(100 / bs).toFixed(2)}%`;
  el.style.backgroundPosition = bs >= 1 ? 'center'
    : `${((bx / (1 - bs)) * 100).toFixed(2)}% ${((by / (1 - bs)) * 100).toFixed(2)}%`;
  el.style.backgroundImage = `url(${url})`;
}
