// Every asset the game can use, named exactly as in docs/ART_PROMPTS.md §13.
// Shared by the renderer (labels for placeholders) and `npm run assets:report`.
// Keep this file free of browser/Pixi imports so Node can run the report directly.

export interface AnimSpec {
  name: string;
  frames: number;
  loop: boolean;
  short: string; // label drawn on placeholders
}

// ART_PROMPTS §4.4
export const CHARACTER_ANIMS: AnimSpec[] = [
  { name: 'idle', frames: 4, loop: true, short: 'idle' },
  { name: 'run', frames: 8, loop: true, short: 'run' },
  { name: 'chase', frames: 8, loop: true, short: 'chase' },
  { name: 'flee', frames: 8, loop: true, short: 'flee' },
  { name: 'hit', frames: 4, loop: false, short: 'hit' },
  { name: 'levelup', frames: 6, loop: false, short: 'lvl+' },
  { name: 'breakthrough', frames: 10, loop: false, short: 'đột phá' },
  { name: 'realm_drop', frames: 6, loop: false, short: 'rớt' },
  { name: 'death', frames: 8, loop: false, short: 'chết' },
  { name: 'respawn', frames: 6, loop: false, short: 'hồi' },
  { name: 'teleport_out', frames: 4, loop: false, short: 'vào cổng' },
  { name: 'teleport_in', frames: 4, loop: false, short: 'ra cổng' },
];

export const CHARACTER_IDS = ['a_hon', 'ba_tam_loa', 'ong_ba_vong', 'dai_ca_dep', 'map_tra_sua', 'co_ba_ganh', 'su_phu_mi', 'be_ti', 'shipper', 'cu_rua'];

// ART_PROMPTS §6
export const MONSTER_IDS = ['banh_bao_cau', 'dep_lao', 'noi_com', 'ga_hoang', 'muoi', 'vua_ghe'];
export const MONSTER_ANIMS: AnimSpec[] = [
  { name: 'idle', frames: 4, loop: true, short: 'idle' },
  { name: 'hurt', frames: 2, loop: false, short: 'đau' },
  { name: 'death', frames: 6, loop: false, short: 'nổ' },
];

export const BUFF_IDS = ['spin', 'haste', 'magnet', 'hard', 'shield', 'exp2', 'stealth'];

// ART_PROMPTS §10.1 — thần thú Tiên giới, ghép đốt bằng code (GAME_SPEC §3.5).
// Rồng dùng chung cho Kim Tiên / Ngọc long / Song long Đạo Tổ, chỉ đổi tint.
export const BEAST_PARTS: { id: string; parts: string[] }[] = [
  { id: 'dragon', parts: ['head', 'body', 'tail', 'leg'] },
  { id: 'crane', parts: ['head', 'body', 'tail'] },
  { id: 'phoenix', parts: ['head', 'body', 'tail', 'feather'] },
];
export const BEAST_MOTES = ['cloud', 'spark', 'glow', 'feather'];

/**
 * BIOME_TILES — texture nền theo biome (ART_PROMPTS §9.0b, GAME_SPEC §8.6).
 *
 * **Một bảng duy nhất, khoá là giá trị `biome` THẬT trong `maps/*.tmj`.** `mapview` nhập từ
 * đây; đừng khai lại ở chỗ khác.
 *
 * Trước đây có **hai** bảng cùng tên và chúng đã trôi khỏi nhau:
 * - `mapview.ts` giữ một bảng đánh khoá bằng biome, nhưng có `grass` `market` `concrete`
 *   `dirt` `rock` — mà bản đồ chỉ sinh ra `grass` `dirt` `concrete` `snow`. Nên `market` và
 *   `rock` là khoá chết, còn **`snow` thì thiếu**: Băng Nguyệt Hồ rơi về `['grass']` và trải
 *   cỏ giữa hồ băng. Hơn nữa bộ dành cho hẻm núi lại nằm dưới khoá `rock`, nên Linh Sơn Cốc
 *   (biome `dirt`) nhận `dirt_path` + `village_grass` — cỏ làng trong hẻm núi đá.
 * - `catalog.ts` giữ một bảng khác, đánh khoá bằng **tên khu** (`'Linh Sơn'`, `'Hắc Thị'`…)
 *   lấy từ cột `name` của `biomes[]` bên `cmd/mapgen/gen.go` — tức tên bốn GÓC của Thiên
 *   Đình, không phải thuộc tính `biome`. Bảng này chỉ `assets:report` / `check:assets` đọc,
 *   nên nó đòi sinh những tile mà game không bao giờ tra tới.
 *
 * Cả hai sai im lặng: không log, không placeholder, chỉ là nền sai. `npm run check:assets`
 * giờ canh đúng chỗ đó — mỗi `.tmj` phải tra ra ít nhất một ảnh nền CÓ THẬT.
 *
 * Lưu ý còn mở: Thiên Đình và Trúc Lâm **cùng** `biome: "grass"` nên trải cùng bộ nền. Muốn
 * Trúc Lâm có nền tre riêng (`bamboo_floor`) thì đổi `zoneDefs` bên mapgen cho `truc_lam`
 * thành một biome riêng rồi sinh lại bản đồ — đó là đổi dữ liệu bản đồ nên chờ chủ dự án.
 */
export const BIOME_TILES: Record<string, string[]> = {
  // Thiên Đình + Trúc Lâm. Giữ nguyên bộ `mapview` đang dùng — đây là nền chủ dự án đã xem.
  grass: ['grass', 'soft_grass', 'bamboo_litter'],
  // Linh Sơn Cốc — hẻm núi đá. Đây là bộ trước nằm dưới khoá chết `rock`.
  dirt: ['dirt', 'rocky_ground', 'fine_gravel'],
  // Phàm Trấn — phố chợ. Bộ trước nằm dưới khoá chết `market`.
  concrete: ['concrete', 'worn_cobble', 'dusty_earth'],
  // Băng Nguyệt Hồ — thiếu hẳn trước đây.
  snow: ['snow', 'ice'],
};
export const DECAL_IDS = [
  'ground_cracks', 'grass_tuft', 'fallen_leaves', 'pebbles', 'puddle', 'tree_roots',
  'wildflowers', 'cart_tracks', 'roof_tiles', 'brick_rubble', 'bamboo_leaves', 'moss_patch',
  'footprints',
];
export const ZONE_EDGES = ['mud_edge', 'pond_shore', 'stone_path_edge', 'bush_edge'];
// Hạt dùng chung (spec §3.6): thả bộ CC0 Kenney/Brackeys vào assets/vfx/particles/.
export const PARTICLE_IDS = ['spark', 'smoke', 'light', 'star', 'dot'];
export const AMBIENT_IDS = ['falling_leaf', 'firefly', 'butterfly', 'chicken', 'lantern', 'bird', 'cloud_shadow'];

// ART_PROMPTS §11.1 / §11.2 — bảng tên và chân dung HUD (GAME_SPEC §12.3, §12.4).
export const PLATE_IDS = ['ribbon_top1', 'ribbon_top2', 'ribbon_top3', 'crown_small', 'realm_pill', 'shine_streak'];
export const PORTRAIT_MOODS = ['neutral', 'evil', 'panic', 'cry'];
export const OBSTACLE_IDS = ['fence', 'stall', 'cart', 'brick', 'boulder', 'stone_wall'];
export const OBSTACLE_STATES = ['intact', 'cracked', 'nearly_broken', 'rubble'];
export const GATE_COLORS = ['purple', 'blue', 'orange'];
export const GATE_ANIMS: AnimSpec[] = [
  { name: 'idle', frames: 6, loop: true, short: 'idle' },
  { name: 'warn', frames: 4, loop: false, short: 'warn' },
  { name: 'burst', frames: 6, loop: false, short: 'burst' },
];
export const VFX: AnimSpec[] = [
  { name: 'clash_spark', frames: 6, loop: false, short: 'keng' },
  { name: 'stun_stars', frames: 4, loop: true, short: 'sao' },
  { name: 'levelup_aura', frames: 6, loop: false, short: 'lvl' },
  { name: 'lightning', frames: 6, loop: false, short: 'sét' },
  { name: 'death_smoke', frames: 6, loop: false, short: 'khói' },
  { name: 'run_dust', frames: 4, loop: false, short: 'bụi' },
  { name: 'sword_shatter', frames: 6, loop: false, short: 'vỡ' },
  { name: 'shield_break', frames: 6, loop: false, short: 'khiên' },
];
export const DECOR_IDS = ['red_plastic_stool', 'old_bicycle_leaning', 'pile_of_watermelons', 'street_food_cart', 'banana_tree',
  'rusty_water_jar', 'broken_tv', 'lazy_sleeping_dog', 'bush_with_flowers', 'stone_lantern', 'wall_border'];
/**
 * DECOR_NATURE: bộ cây cỏ / đá / giếng mà `cmd/mapgen/decor.go` gắn tag cho Linh Sơn Cốc,
 * Trúc Lâm, Băng Nguyệt Hồ (job artgen `decor_nature`, Q27-2).
 *
 * Tách khỏi `DECOR_IDS` để `assets:report` liệt kê riêng — bộ trên đã có art, bộ này chưa.
 * Mapgen **không đặt** cái nào chưa có art, nên thiếu là bản đồ trống trải chứ không vẽ
 * placeholder; vì vậy `check:assets` chỉ cảnh báo, không chặn build.
 */
export const DECOR_NATURE = ['rock_small', 'pine', 'bamboo', 'bush_small', 'lantern', 'well', 'banner'];
export const SFX = ['sword_clash', 'sword_hit_body', 'pickup_sword', 'pickup_buff', 'pickup_exp', 'buff_end', 'sword_break',
  'shield_break', 'chest_incoming', 'chest_open', 'obstacle_hit', 'obstacle_break', 'obstacle_regrow', 'gate_in', 'gate_out',
  'gate_warn', 'monster_hurt', 'monster_pop', 'levelup', 'breakthrough', 'realm_drop', 'death', 'respawn'];
export const REALMS = 9;

/** One entry of the report: a group of files that together make one usable asset. */
export interface CatalogEntry {
  group: string;
  label: string;
  files: string[]; // paths relative to assets/
  job: string; // `artgen` job (docs/ART_JOBS.md) that produces it
  // Files that must exist for the asset to be usable: the first frame of each animation.
  // Generators often draw a few more or fewer frames than asked, which is fine.
  required?: string[];
}

function frames(dir: string, a: AnimSpec): string[] {
  return Array.from({ length: a.frames }, (_, i) => `${dir}/${a.name}_${String(i).padStart(2, '0')}.png`);
}

function characterEntries(): CatalogEntry[] {
  return CHARACTER_IDS.flatMap((c) =>
    CHARACTER_ANIMS.map((a) => ({ group: 'Nhân vật', label: `${c} · ${a.name}`, files: frames(`characters/${c}/frames`, a), job: `char_${c}_${a.name}` })),
  );
}

function monsterEntries(): CatalogEntry[] {
  return MONSTER_IDS.map((m) => ({
    group: 'Quái',
    label: m,
    files: MONSTER_ANIMS.flatMap((a) => frames(`monsters/${m}/frames`, a)),
    job: `mon_${m}`,
  }));
}

function itemEntries(): CatalogEntry[] {
  const out: CatalogEntry[] = [];
  for (let r = 0; r < REALMS; r++) {
    out.push({ group: 'Kiếm', label: `realm_${r}`, files: [`swords/realm_${r}.png`], job: `sword_realm_${r}` });
  }
  out.push({ group: 'Vật phẩm', label: 'sword', files: ['items/sword.png'], job: 'item_sword' });
  // Viên Tiên khí mệnh giá lớn (spec v6.4 §3.4 mục 10). Chỉ 100 và 500 có ảnh riêng; 1/2/10
  // vẫn vẽ bằng Graphics vì trên màn hình chúng bé bằng hạt đậu.
  for (const v of [100, 500]) {
    out.push({ group: 'Vật phẩm', label: `Tiên khí ${v}`, files: [`items/tien_orb_${v}.png`], job: `item_tien_orb_${v}` });
  }
  for (const o of ['orb_small', 'orb_medium', 'orb_large']) {
    out.push({ group: 'Vật phẩm', label: o, files: [`items/${o}.png`], job: `item_${o}` });
  }
  for (const b of BUFF_IDS) {
    out.push({ group: 'Vật phẩm', label: b, files: [`buffs/${b}.png`], job: `buff_${b}` });
  }
  for (const b of BEAST_PARTS) {
    for (const part of b.parts) {
      out.push({ group: 'Thần thú Tiên giới', label: `${b.id} ${part}`, files: [`beasts/${b.id}_${part}.png`], job: `beast_${b.id}_${part}` });
    }
  }
  for (const [biome, tiles] of Object.entries(BIOME_TILES)) {
    for (const t of tiles) {
      out.push({ group: 'Nền bản đồ', label: `${biome} · ${t}`, files: [`tiles/${t}.png`], job: `tile_${t}` });
    }
  }
  for (const d of DECAL_IDS) {
    out.push({ group: 'Decal mặt đất', label: d, files: [`decals/${d}.png`], job: `decal_${d}` });
  }
  for (const e of ZONE_EDGES) {
    out.push({ group: 'Decal mặt đất', label: `viền ${e}`, files: [`decals/${e}.png`], job: `decal_${e}` });
  }
  for (const p of PARTICLE_IDS) {
    out.push({ group: 'Hạt VFX', label: p, files: [`vfx/particles/${p}.png`], job: `particle_${p}` });
  }
  for (const a of AMBIENT_IDS) {
    out.push({ group: 'Sự sống bản đồ', label: a, files: [`ambient/${a}.png`], job: `ambient_${a}` });
  }
  for (const p of PLATE_IDS) {
    out.push({ group: 'Bảng tên', label: p, files: [`ui/${p}.png`], job: `plate_${p}` });
  }
  for (const c of CHARACTER_IDS) {
    for (const m of PORTRAIT_MOODS) {
      out.push({ group: 'Chân dung HUD', label: `${c} · ${m}`, files: [`characters/${c}/portrait_${m}.png`], job: `portrait_${c}_${m}` });
    }
  }
  for (const m of BEAST_MOTES) {
    out.push({ group: 'Thần thú Tiên giới', label: `hạt ${m}`, files: [`beasts/mote_${m}.png`], job: `beast_mote_${m}` });
  }
  out.push({ group: 'Vật phẩm', label: 'chest', files: [...frames('items/chest', { name: 'idle', frames: 4, loop: true, short: '' }), ...frames('items/chest', { name: 'open', frames: 6, loop: false, short: '' })], job: 'chest' });
  return out;
}

function mapEntries(): CatalogEntry[] {
  const out: CatalogEntry[] = [];
  for (const t of ['grass', 'concrete', 'dirt', 'mud']) {
    out.push({ group: 'Bản đồ', label: `tile ${t}`, files: [`tiles/${t}.png`], job: `tile_${t}` });
  }
  out.push({ group: 'Bản đồ', label: 'bush', files: ['zones/bush.png'], job: 'zone_bush' });
  out.push({ group: 'Bản đồ', label: 'wind', files: frames('zones', { name: 'wind', frames: 4, loop: true, short: '' }), job: 'zone_wind' });
  for (const o of OBSTACLE_IDS) {
    out.push({ group: 'Vật cản', label: `${o} states`, files: OBSTACLE_STATES.map((s) => `obstacles/${o}/${s}.png`), job: `obs_${o}_states` });
    // Đoạn tường/rào dựng dọc cần ảnh riêng — xoay ảnh ngang 90° trông sai bóng đổ (§8.2).
    out.push({ group: 'Vật cản', label: `${o} states (đoạn dọc)`, files: OBSTACLE_STATES.map((s) => `obstacles/${o}/${s}_vertical.png`), job: `obs_${o}_states_vertical` });
    if (o !== 'stone_wall') {
      for (const a of ['break', 'regrow']) {
        out.push({ group: 'Vật cản', label: `${o} ${a}`, files: frames(`obstacles/${o}/${a}`, { name: a, frames: 6, loop: false, short: '' }), job: `obs_${o}_${a}` });
      }
    }
  }
  for (const c of GATE_COLORS) {
    out.push({ group: 'Cổng', label: c, files: GATE_ANIMS.flatMap((a) => frames(`gates/${c}`, a)), job: `gate_${c}` });
  }
  out.push({ group: 'Trang trí', label: 'decor phố', files: DECOR_IDS.map((d) => `decor/${d}.png`), job: 'decor_sheet' });
  out.push({ group: 'Trang trí', label: 'decor cây cỏ / đá / giếng', files: DECOR_NATURE.map((d) => `decor/${d}.png`), job: 'decor_nature' });
  return out;
}

function fxUiAudioEntries(): CatalogEntry[] {
  const out: CatalogEntry[] = VFX.map((v) => ({ group: 'VFX', label: v.name, files: frames(`vfx/${v.name}`, v), job: `vfx_${v.name}` }));
  out.push({ group: 'VFX', label: 'crack', files: ['vfx/crack.png'], job: 'vfx_crack' });
  out.push({ group: 'UI', label: 'badges', files: Array.from({ length: REALMS }, (_, i) => `ui/badge_${i}.png`), job: 'ui_badges' });
  out.push({ group: 'UI', label: 'medals', files: ['gold', 'silver', 'bronze'].map((m) => `ui/medal_${m}.png`), job: 'ui_badges' });
  out.push({ group: 'UI', label: 'speech bubble', files: ['ui/speech_bubble.png'], job: 'ui_speech' });
  out.push({ group: 'UI', label: 'scroll', files: ['ui/scroll.png'], job: 'ui_scroll' });
  out.push({ group: 'UI', label: 'home background', files: ['ui/home_bg.png'], job: 'ui_home_bg' });
  for (const s of SFX) {
    out.push({ group: 'Âm thanh', label: s, files: [`audio/sfx/${s}.ogg`], job: '' });
  }
  out.push({ group: 'Âm thanh', label: 'bgm', files: ['audio/bgm/main.ogg'], job: '' });
  return out;
}

export function catalog(): CatalogEntry[] {
  return [...characterEntries(), ...monsterEntries(), ...itemEntries(), ...mapEntries(), ...fxUiAudioEntries()].map((e) => ({
    ...e,
    required: e.required ?? e.files.filter((f) => !/_\d\d\.png$/.test(f) || f.endsWith('_00.png')),
  }));
}

/** Placeholder label for a character: initials of its id, e.g. "a_hon" → "AH". */
export function initials(id: string): string {
  return id
    .split('_')
    .map((w) => w[0] ?? '')
    .join('')
    .toUpperCase()
    .slice(0, 3);
}

export function animShort(name: string): string {
  return CHARACTER_ANIMS.find((a) => a.name === name)?.short ?? name;
}
