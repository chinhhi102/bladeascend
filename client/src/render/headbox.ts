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
  'a_hon/neutral': [0.22, 0.05, 0.52], // đo bằng mắt
  'a_hon/evil': [0.2, 0.07, 0.66], // đo bằng mắt
  'a_hon/panic': [0.16, 0.01, 0.66], // đo bằng mắt
  'a_hon/cry': [0.26, 0.05, 0.48], // đo bằng mắt
  'ba_tam_loa/neutral': [0.309, 0.201, 0.574],
  'ba_tam_loa/evil': [0.238, 0.086, 0.504],
  'ba_tam_loa/panic': [0.148, 0.162, 0.586],
  'ba_tam_loa/cry': [0.197, 0.041, 0.535],
  'be_ti/neutral': [0.24, 0.139, 0.59],
  'be_ti/evil': [0.244, 0.188, 0.619],
  'be_ti/panic': [0.205, 0.205, 0.57],
  'be_ti/cry': [0.256, 0.176, 0.566],
  'co_ba_ganh/neutral': [0.281, 0.113, 0.438],
  'co_ba_ganh/evil': [0.252, 0.225, 0.561],
  'co_ba_ganh/panic': [0.197, 0.178, 0.617],
  'co_ba_ganh/cry': [0.248, 0.232, 0.498],
  'cu_rua/neutral': [0.166, 0.08, 0.715],
  'cu_rua/evil': [0.121, 0.063, 0.875],
  'cu_rua/panic': [0.131, 0.064, 0.791],
  'cu_rua/cry': [0.197, 0.203, 0.568],
  'dai_ca_dep/neutral': [0.172, 0.1, 0.75],
  'dai_ca_dep/evil': [0.281, 0.166, 0.594],
  'dai_ca_dep/panic': [0.225, 0.039, 0.596],
  'dai_ca_dep/cry': [0.17, 0.123, 0.725],
  'map_tra_sua/neutral': [0.178, 0, 0.676],
  'map_tra_sua/evil': [0.266, 0.027, 0.572],
  'map_tra_sua/panic': [0.117, 0, 0.766],
  'map_tra_sua/cry': [0.182, 0, 0.607],
  'ong_ba_vong/neutral': [0.246, 0.078, 0.699],
  'ong_ba_vong/evil': [0.29, 0.06, 0.47], // đo bằng mắt
  'ong_ba_vong/panic': [0.344, 0.141, 0.582],
  'ong_ba_vong/cry': [0.189, 0.049, 0.766],
  'shipper/neutral': [0.402, 0.068, 0.561],
  'shipper/evil': [0.55, 0.24, 0.45], // đo bằng mắt
  'shipper/panic': [0.449, 0.23, 0.346],
  'shipper/cry': [0.344, 0.158, 0.555],
  'su_phu_mi/neutral': [0.137, 0.189, 0.52],
  'su_phu_mi/evil': [0.326, 0.139, 0.379],
  'su_phu_mi/panic': [0.293, 0.109, 0.588],
  'su_phu_mi/cry': [0.227, 0, 0.719],
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
