// Subset of config/game.json the client needs. Fetched from GET /api/config — never hardcode numbers.

export interface BuffCfg {
  mult: number;
  secs: number;
  reveal_r?: number;
}

export interface GameConfig {
  cultivation: {
    realms: string[];
    tiers_per_realm: number;
    swords_per_realm: number;
    scale_min: number;
    scale_max: number;
  };
  sword: { tween_secs: number };
  player: {
    body_r: number;
    move_base: number;
    camera_view_h: number;
    camera_exp: number;
    name_min: number;
    name_max: number;
  };
  buffs: Record<string, BuffCfg>;
  zones: { mud_mult: number; wind_with: number; wind_against: number };
  gates: { radius: number };
  net: { tick_hz: number; snapshot_hz: number; interp_ms: number };
  characters: string[];
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

export function maxSwords(cfg: GameConfig, realm: number): number {
  return cfg.cultivation.swords_per_realm * (realm + 1);
}

export function cameraViewHeight(cfg: GameConfig, scale: number): number {
  return cfg.player.camera_view_h * Math.pow(scale, cfg.player.camera_exp);
}
