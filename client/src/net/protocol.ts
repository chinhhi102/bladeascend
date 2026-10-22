// Wire types mirroring server/internal/netproto (MessagePack maps with short keys).

export interface PlayerState {
  id: number;
  x: number;
  y: number;
  f: number; // facing
  L: number; // level index
  n: number; // sword count
  ph: number; // ring phase at snapshot time
  d: number; // ring direction ±1
  w: number; // ω rad/s
  tl: number; // blade tilt (rad); ring sizes are derived from sc, n and the Trường Kiếm buff
  sc: number; // scale
  md: string; // mood = animation name
  bf: number; // buff bits
  fl: number; // flags
  ck?: number; // crack
  xp: number;
  nm?: string;
  ch?: string;
}

export interface MonsterState {
  id: number;
  x: number;
  y: number;
  k: string;
  hp: number;
  mx: number;
  r: number;
}

export interface ItemState {
  id: number;
  x: number;
  y: number;
  k: string;
  v?: number;
}

export interface GameEvent {
  k: string;
  a?: number;
  b?: number;
  x: number;
  y: number;
  n?: number;
  s?: string;
}

export interface Snapshot {
  t: 'snap';
  tk: number;
  tm: number;
  ack: number;
  p: PlayerState[] | null;
  m: MonsterState[] | null;
  ia?: ItemState[];
  ir?: number[];
  ev?: GameEvent[];
}

export interface Welcome {
  t: 'welcome';
  id: number;
  room: string;
  map: string;
  tick_hz: number;
  time: number;
  obs: { id: number; hp: number; b: boolean }[] | null;
  debug: boolean;
}

export interface LeaderboardMsg {
  t: 'lb';
  top: { id: number; nm: string; xp: number; L: number }[];
  rank: number;
  total: number;
}

export interface FeedMsg {
  t: 'feed';
  k: string;
  a: string;
  b?: string;
  r: number;
}

export interface DeadMsg {
  t: 'dead';
  killer: string;
  killer_char: string;
  peak_level: number;
  peak_exp: number;
  kills: number;
  survive: number;
  rank: number;
}

export interface PongMsg {
  t: 'pong';
  c: number;
  s: number;
}

export interface ErrorMsg {
  t: 'error';
  msg: string;
}

export type ServerMsg = Snapshot | Welcome | LeaderboardMsg | FeedMsg | DeadMsg | PongMsg | ErrorMsg;

// Player flags (netproto.Flag*).
export const FLAG = {
  invuln: 1,
  bush: 2,
  stealth: 4,
  transit: 8,
  dead: 16,
  bot: 32,
  revealed: 64,
} as const;

// Buff bit order = server game.BuffNames.
export const BUFFS = ['spin', 'haste', 'magnet', 'long', 'shield', 'exp2', 'stealth'] as const;

// Map JSON from GET /api/map/{id} (server/internal/world.Map).
export interface Shape {
  c?: boolean;
  x: number;
  y: number;
  w?: number;
  h?: number;
  r?: number;
}

export interface MapObstacle {
  id: number;
  k: string;
  s: Shape; // what you see; swords hit it
  f: Shape; // footprint at the bottom; blocks bodies (spec §8.2)
  hp: number;
}

export interface MapZone {
  id: number;
  k: string;
  s: Shape;
  wd?: number;
}

export interface MapGate {
  id: number;
  x: number;
  y: number;
  n: string;
  c: string;
}

export interface GameMap {
  id: string;
  name: string;
  size: number;
  biome: string;
  obstacles: MapObstacle[];
  zones: MapZone[];
  gates: MapGate[];
  decor: { sp: string; x: number; y: number }[];
}
