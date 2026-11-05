// SINH TỰ ĐỘNG bằng `node tools/portrait-headbox.mjs` — đừng sửa tay.
// Sửa số thì sửa bảng OVERRIDE trong tool rồi chạy lại, nếu không lần chạy sau mất hết.

/**
 * Khung đầu của từng chân dung: `[trái, trên, cạnh]` theo TỈ LỆ cạnh ảnh (0–1), luôn vuông.
 *
 * Dùng cho hai chỗ cắt ảnh vào khung TRÒN: ô đại diện HUD (§12.4) và bong bóng mặt (§12.5).
 * Chân dung là ảnh bán thân/toàn thân nên cắt bằng một luật chung ("lấy nửa trên, phóng
 * 180%") là ra cái nón với hai con mắt — mỗi ảnh phải có khung riêng.
 */
export const HEAD_BOX: Record<string, readonly [number, number, number]> = {
  'a_hon/neutral': [0.148, 0.014, 0.68],
  'a_hon/evil': [0.2, 0.08, 0.64], // đo bằng mắt
  'a_hon/panic': [0.078, 0.006, 0.842],
  'a_hon/cry': [0.146, 0.014, 0.707],
  'ba_tam_loa/neutral': [0.459, 0.094, 0.391],
  'ba_tam_loa/evil': [0.246, 0, 0.381],
  'ba_tam_loa/panic': [0.125, 0.037, 0.461],
  'ba_tam_loa/cry': [0.225, 0, 0.377],
  'be_ti/neutral': [0.324, 0.076, 0.416],
  'be_ti/evil': [0.311, 0.055, 0.463],
  'be_ti/panic': [0.281, 0.08, 0.438],
  'be_ti/cry': [0.402, 0.094, 0.344],
  'co_ba_ganh/neutral': [0.3, 0.03, 0.42], // đo bằng mắt
  'co_ba_ganh/evil': [0.066, 0.051, 0.818],
  'co_ba_ganh/panic': [0.051, 0.035, 0.877],
  'co_ba_ganh/cry': [0.113, 0.094, 0.746],
  'cu_rua/neutral': [0.22, 0.07, 0.58], // đo bằng mắt
  'cu_rua/evil': [0.293, 0, 0.541],
  'cu_rua/panic': [0.22, 0.06, 0.62], // đo bằng mắt
  'cu_rua/cry': [0.25, 0.08, 0.58], // đo bằng mắt
  'dai_ca_dep/neutral': [0.328, 0, 0.418],
  'dai_ca_dep/evil': [0.367, 0.061, 0.404],
  'dai_ca_dep/panic': [0.244, 0, 0.545],
  'dai_ca_dep/cry': [0.219, 0, 0.648],
  'map_tra_sua/neutral': [0.27, 0, 0.463],
  'map_tra_sua/evil': [0.311, 0, 0.439],
  'map_tra_sua/panic': [0.182, 0, 0.545],
  'map_tra_sua/cry': [0.248, 0, 0.42],
  'ong_ba_vong/neutral': [0.348, 0, 0.512],
  'ong_ba_vong/evil': [0.33, 0.09, 0.36], // đo bằng mắt
  'ong_ba_vong/panic': [0.406, 0.045, 0.426],
  'ong_ba_vong/cry': [0.3, 0.06, 0.54], // đo bằng mắt
  'shipper/neutral': [0.314, 0, 0.568],
  'shipper/evil': [0.45, 0.17, 0.4], // đo bằng mắt
  'shipper/panic': [0.45, 0.18, 0.4], // đo bằng mắt
  'shipper/cry': [0.39, 0.14, 0.44], // đo bằng mắt
  'su_phu_mi/neutral': [0.31, 0.11, 0.5], // đo bằng mắt
  'su_phu_mi/evil': [0.32, 0.1, 0.38], // đo bằng mắt
  'su_phu_mi/panic': [0.229, 0, 0.713],
  'su_phu_mi/cry': [0.256, 0, 0.578],
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
