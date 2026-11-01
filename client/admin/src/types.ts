// Kiểu dữ liệu các API quản trị trả về. Tên trường khớp thẻ `json:"..."` bên server
// (`server/cmd/server/*.go`, `internal/admin`, `internal/store`, `internal/room`).

export interface ZoneStat {
  index: number;
  name: string;
  map: string;
  humans: number;
  bots: number;
  tick_avg: number;
  tick_p95: number;
  top_level: number;
  top_realm: string;
  monsters: number;
  items: number;
  broken_obstacles: number;
}

export interface RoomStat {
  id: string;
  humans: number;
  bots: number;
  zones: ZoneStat[] | null;
}

export interface Stats {
  at: number;
  build: string;
  uptime_secs: number;
  players: { humans: number; bots: number; rooms: number };
  server: { goroutines: number; heap_mb: number; tick_avg_max: number; tick_p95_max: number };
  realms: Record<string, number>;
  tien_count: number;
  rooms: RoomStat[] | null;
  feedback_new: number;
  flagged: number;
  errors_today: number;
}

export interface Me {
  csrf: string;
  expires: number;
  build: string;
}

export interface Flag {
  kind: string;
  name: string;
  at: number;
  count: number;
  evidence: string;
}

export interface Player {
  guest_id: string;
  name: string;
  character: string;
  device: string;
  ip_mask: string;
  room: string;
  zone: string;
  realm: string;
  level: number;
  swords: number;
  score: number;
  hits: number;
  alive_secs: number;
  ping_ms: number;
  fps: number;
  corrections: number;
  flags: Flag[] | null;
  debug_source: string;
  debug_used: boolean;
}

export interface Feedback {
  id: number;
  at: number;
  guest_id: string;
  name: string;
  kind: string;
  body: string;
  context: string;
  has_image: boolean;
  /** images: số ảnh 0–3 (góp ý nhiều ảnh). Bản server cũ không có trường này. */
  images?: number;
  status: string;
  note: string;
}

export interface Ban {
  id: number;
  guest_id: string;
  ip: string;
  reason: string;
  from: number;
  until: number | null;
}

export interface AuditEntry {
  at: number;
  action: string;
  target: string;
  reason: string;
  ip: string;
  details: string;
}

export interface ErrorGroup {
  kind: string;
  sig: string;
  sample: string;
  count: number;
  sessions: number;
  first_at: number;
  last_at: number;
  devices: string[] | null;
  builds: string[] | null;
  mobile: number;
}

export interface ErrorsResp {
  off?: boolean;
  note?: string;
  days?: number;
  groups: ErrorGroup[] | null;
}

export interface Life {
  guest_id: string;
  name: string;
  character: string;
  score: number;
  hits: number;
  peak_level: number;
  kills: number;
  survive_secs: number;
  map: string;
  ended_at: number;
  hidden: boolean;
  hide_reason: string;
}

export interface CharRow {
  character: string;
  lives: number;
  pick_share: number;
  score_per_min: number;
  avg_survive: number;
  avg_peak_realm: number;
  casts_per_min: number;
  casts: number;
}

export interface BotRoom {
  id: string;
  ok: boolean;
  override: string;
  effective: string;
  auto_off: boolean;
  spawning: boolean;
  humans: number;
  bots: number;
  max_humans: number;
  withdrawing: number;
}

export interface BotsResp {
  mode: string;
  rooms: BotRoom[] | null;
}

export interface RoomSettings {
  max_humans: number;
  zone_min_chars: number;
  bot_auto_off_at: number;
}

export interface RoomSettingsView {
  env: string;
  settings: RoomSettings;
  defaults: RoomSettings;
  loadtested_humans: number;
  warn: boolean;
}

/** CheckRow: một dòng báo cáo `mapgen --check` cho một cảnh giới (tên trường kiểu Go). */
export interface CheckRow {
  Realm: number;
  Name: string;
  BodyR: number;
  Density: number;
  Breathing: number;
  OpenFrac: number;
  BiggestTrap: number;
  CutOff: string[] | null;
}

export interface CheckReport {
  maps?: Record<string, CheckRow[]>;
  targets?: Record<string, [number, number]>;
}
