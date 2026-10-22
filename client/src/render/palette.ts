// Placeholder colours taken from the character palettes in docs/ART_PROMPTS.md §5.
export const CHAR_COLORS: Record<string, [number, number]> = {
  a_hon: [0xe63946, 0xf4a261],
  ba_tam_loa: [0x2a9d8f, 0xe9c46a],
  ong_ba_vong: [0x8d99ae, 0x6d597a],
  dai_ca_dep: [0x7209b7, 0xffd60a],
  map_tra_sua: [0xcdb4db, 0xa47148],
  co_ba_ganh: [0xbc6c25, 0x606c38],
  su_phu_mi: [0xfb8500, 0x8ecae6],
  be_ti: [0xe76f51, 0x2a9d8f],
  shipper: [0xffbe0b, 0x3a86ff],
  cu_rua: [0x52b788, 0x6f4e37],
};

// Sword skin colours per realm (ART_PROMPTS §7), used when the sword PNG is missing.
export const REALM_COLORS = [0xc8a165, 0xb0b7bf, 0xd0705a, 0xe0e6ee, 0xff7b29, 0x7fd4ff, 0xffe14d, 0x8a3fd1, 0x9ff0c8];

// Aura colours per buff (ART_PROMPTS §8); stealth has no aura.
export const BUFF_COLORS: Record<string, number> = {
  spin: 0x3ee6ff,
  haste: 0xff9a3c,
  magnet: 0xff4d4d,
  long: 0xb36bff,
  shield: 0xffd23f,
  exp2: 0x5dff7a,
};

export const GATE_COLORS: Record<string, number> = { purple: 0xa55bff, blue: 0x3a86ff, orange: 0xff8c1a };

// Top 1–3 room title colours: vàng kim / bạc / đồng (spec §11.3).
export const TITLE_COLORS = [0xffd700, 0xd8dde3, 0xcd7f32];

export const MONSTER_VARIANTS: Record<string, string[]> = {
  small: ['banh_bao_cau', 'dep_lao'],
  medium: ['noi_com', 'ga_hoang'],
  large: ['muoi'],
  boss: ['vua_ghe'],
};

export const MONSTER_COLORS: Record<string, number> = { small: 0xf7e7c6, medium: 0xffd166, large: 0x9a8c98, boss: 0xe63946 };

export function charColor(id: string): [number, number] {
  return CHAR_COLORS[id] ?? [0x999999, 0x666666];
}
