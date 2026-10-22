// Subset of config/game.json the client needs. Fetched from GET /api/config — never hardcode numbers.

export interface BuffCfg {
  mult: number;
  secs: number;
  reveal_r?: number;
  visible_alpha?: number; // Ẩn Thân: độ đậm khi bị nhìn từ xa (spec §6.2)
}

export interface GameConfig {
  cultivation: {
    realms: string[];
    tiers_per_realm: number;
    exp_base: number;
    realm_mult: number;
    tier_step: number;
    scale_min: number;
    scale_max: number;
  };
  sword: {
    tween_secs: number;
    max_swords: number[];
    swords_per_ring: number;
    length: number;
    width: number;
    hilt_gap: number;
    ring_growth: number;
    ring_growth_mode: string;
    hilt_step: number;
    tilt_deg: number;
    ring_tints: string[];
    rev_secs: number;
  };
  player: {
    body_r: number;
    move_base: number;
    camera_view_h: number;
    camera_exp: number;
    camera_reach_mult: number;
    name_min: number;
    name_max: number;
  };
  burn: { mult: number; exp_frac_per_sec: number; keep_realm: boolean };
  buffs: Record<string, BuffCfg>;
  zones: { mud_mult: number; wind_with: number; wind_against: number; wind_with_max_deg: number; wind_against_min_deg: number };
  gates: { radius: number };
  items: { item_r: number; chest_warn_secs: number };
  obstacles: { occlusion_alpha: number };
  net: { tick_hz: number; snapshot_hz: number; interp_ms: number };
  characters: string[];
  titles: string[];
  minimap: { size_pc: number; size_mobile: number; show_top1: boolean; show_chest: boolean; show_others: boolean };
  score: { hit_base: number; kill_bonus: number; upset_per_realm: number; breakthrough_per_realm: number; top1_per_sec: number };
}

export async function loadConfig(): Promise<GameConfig> {
  const res = await fetch('/api/config');
  if (!res.ok) {
    throw new Error(`config: HTTP ${res.status}`);
  }
  return (await res.json()) as GameConfig;
}

export function realmOf(cfg: GameConfig, level: number): number {
  return Math.floor(level / cfg.cultivation.tiers_per_realm);
}

export function tierOf(cfg: GameConfig, level: number): number {
  return (level % cfg.cultivation.tiers_per_realm) + 1;
}

/** maxSwords mirrors game.Progression.MaxSwords: max_swords[r], +40 per realm past the array. */
export function maxSwords(cfg: GameConfig, realm: number): number {
  const a = cfg.sword.max_swords;
  if (!a.length) {
    return 1;
  }
  return realm < a.length ? a[Math.max(realm, 0)] : a[a.length - 1] + 40 * (realm - a.length + 1);
}

/** cameraViewHeight backs off far enough to show the whole ring (spec §4.1). */
export function cameraViewHeight(cfg: GameConfig, scale: number, reach: number): number {
  const p = cfg.player;
  return Math.max(p.camera_view_h * Math.pow(scale, p.camera_exp), p.camera_reach_mult * reach);
}

/** levelThresholds[L] = total EXP needed to reach level L (mirrors game.Progression). */
export function levelThresholds(cfg: GameConfig): number[] {
  const c = cfg.cultivation;
  const levels = c.realms.length * c.tiers_per_realm;
  const out: number[] = [];
  let sum = 0;
  for (let L = 0; L <= levels; L++) {
    out.push(sum);
    const r = Math.floor(L / c.tiers_per_realm);
    const s = L % c.tiers_per_realm;
    sum += Math.round(c.exp_base * Math.pow(c.realm_mult, r) * (1 + c.tier_step * s));
  }
  return out;
}
