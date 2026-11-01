import { STR } from './ui/strings';

// Subset of config/game.json the client needs. Fetched from GET /api/config — never hardcode numbers.

export interface SpeedCfg {
  cap: number;
  cap_gom_don: number;
  floor: number;
  haste: number;
  burn: number;
  gom_don: number;
  wind_with: number;
  wind_against: number;
  mud: number;
  snow: number;
  slow_status: number;
}

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
    // Tiên giới (spec §3.4): 5 cảnh giới trên Đại Thừa, lên bằng sát mạng.
    tien_realms: string[];
    tien_kills_per_step: number[];
    tien_scale_step: number;
    tien_aura_colors: string[];
  };
  sword: {
    tween_secs: number;
    max_swords: number[];
    ring_capacity_step: number;
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
    /** pickup_r: bán kính hút Linh khí — cũng là đơn vị **U** của phạm vi kỹ năng (§6.3). */
    pickup_r: number;
    camera_view_h: number;
    camera_exp: number;
    camera_reach_mult: number;
    name_min: number;
    name_max: number;
  };
  /** Tốc độ CỘNG trên base rồi kẹp (spec v6.25 §4.1) — khớp `config.Speed` bên server. */
  speed: SpeedCfg;
  burn: { exp_frac_per_sec: number; keep_realm: boolean };
  buffs: Record<string, BuffCfg>;
  /** Mặt băng ở Băng Nguyệt Hồ (spec §8.0): quán tính khi trượt. */
  ice: { friction: number; accel: number };
  /**
   * Kỹ năng nhân vật (spec §6.3). Client chỉ cần những con số ảnh hưởng tới **dự đoán vị
   * trí**: dự đoán sai tốc độ là bị server kéo về mỗi tick (§8.2a). Phần còn lại của bảng
   * kỹ năng là việc của server.
   */
  skills: {
    ru_ngu_move: number;
    tran_chau_secs: number;
    /** Hồi chiêu theo nhân vật: [giây hồi 1 lượt, lượt tích tối đa] (spec v6.25 §6.3). */
    cooldowns: Record<string, [number, number]>;
    /** Tiên giới +N lượt tích tối đa. */
    tien_extra_charges: number;
    /** Góc nón của hai kỹ năng hình nón — chỉ báo phạm vi phải vẽ đúng góc server xét. */
    hut_mi_cone_deg: number;
    tran_chau_cone_deg: number;
    /**
     * ranges_u: hệ số phạm vi theo U (spec v6.5 §6.3). Client vẽ chỉ báo bằng **đúng bộ số
     * server dùng để xét trúng** — lấy từ đây chứ không gõ lại, nếu không thì vòng tròn
     * người chơi thấy và vùng thật sự ăn đòn là hai thứ khác nhau.
     */
    ranges_u: Record<string, number>;
  };
  zones: { wind_with_max_deg: number; wind_against_min_deg: number };
  gates: { radius: number };
  items: { item_r: number; chest_warn_secs: number };
  obstacles: { occlusion_alpha: number; hard_buff_extra_tier: number; cut_hint_cooldown_secs: number };
  net: {
    tick_hz: number;
    snapshot_hz: number;
    interp_ms: number;
    /** Lệch quá ngần này thì nhảy thẳng thay vì kéo dần (spec §8.2a). */
    correction_snap: number;
    /** Hằng số thời gian kéo nhân vật về vị trí server, tính bằng mili giây (spec §8.2a). */
    correction_tau_ms: number;
  };
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

/**
 * skillU: đơn vị phạm vi kỹ năng của một người (GAME_SPEC §6.3) — **khớp từng chữ với
 * `Game.SkillU` bên server**: `pickup_r × scale`.
 *
 * Client chỉ dùng nó để VẼ chỉ báo (nón, vòng tròn). Server mới là bên xét trúng. Nhưng hai
 * bên phải ra cùng một con số, nếu không người chơi ngắm theo một vòng tròn rồi ăn đòn ở
 * ngoài vòng đó — kiểu sai không ai chứng minh được bằng mắt, chỉ thấy "game ăn gian".
 */
export function skillU(cfg: GameConfig, scale: number): number {
  return cfg.player.pickup_r * scale;
}

/** skillRange: `ranges_u[key] × U`. Thiếu khoá thì 0 — không vẽ gì còn hơn vẽ sai vùng. */
export function skillRange(cfg: GameConfig, scale: number, key: string): number {
  return (cfg.skills.ranges_u?.[key] ?? 0) * skillU(cfg, scale);
}

export function realmOf(cfg: GameConfig, level: number): number {
  return Math.floor(level / cfg.cultivation.tiers_per_realm);
}

export function tierOf(cfg: GameConfig, level: number): number {
  return (level % cfg.cultivation.tiers_per_realm) + 1;
}

/** isTien: cảnh giới r thuộc Tiên giới (spec §3.4). */
export function isTien(cfg: GameConfig, realm: number): boolean {
  return realm >= cfg.cultivation.realms.length;
}

/**
 * banksTienKhi: từ Đại Thừa tầng 10 trở lên thì chỉ tu Tiên khí (spec §3.4). Dưới mức đó chỉ
 * tu Linh khí — hai đường tu tách hẳn, không đổi qua lại.
 */
export function banksTienKhi(cfg: GameConfig, level: number): boolean {
  const c = cfg.cultivation;
  return level >= c.realms.length * c.tiers_per_realm - 1;
}

/** absorbable: viên này người chơi hiện tại có hấp thụ được không. */
export function absorbable(kind: string, banks: boolean): boolean {
  if (kind === 'orb') {
    return !banks;
  }
  if (kind === 'tien_khi') {
    return banks;
  }
  return true;
}

/** tienIndex: vị trí 0-based trong tien_realms, -1 nếu còn phàm. */
export function tienIndex(cfg: GameConfig, realm: number): number {
  return isTien(cfg, realm) ? realm - cfg.cultivation.realms.length : -1;
}

/** realmName gộp cả 9 cảnh giới phàm và 5 cảnh giới Tiên. */
export function realmName(cfg: GameConfig, realm: number): string {
  // Bản dịch (spec §19 bảng thuật ngữ) đè tên tiếng Việt trong config; rỗng = dùng config.
  const tr = STR.realmNames[realm];
  if (tr) {
    return tr;
  }
  const c = cfg.cultivation;
  if (realm < c.realms.length) {
    return c.realms[realm] ?? '';
  }
  return c.tien_realms[realm - c.realms.length] ?? '';
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
