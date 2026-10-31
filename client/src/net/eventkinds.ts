/**
 * Bảng loại sự kiện và loại quái — Protocol v2 (NETCODE_SPEC §4.1).
 *
 * ⚠ **PHẢI KHỚP TỪNG DÒNG** với `server/internal/netproto/eventkinds.go`. Server gửi chỉ
 * số; lệch một dòng là client vẽ nhầm hiệu ứng mà **không có lỗi nào được ném ra** — kiểu
 * hỏng khó tìm nhất.
 *
 * Có test giữ hai bên khớp nhau: `npm run check:wire` so bảng này với bảng bên Go.
 */
export const EVENT_KINDS = [
  'AltarBurst',
  'AltarWarn',
  'Blinded',
  'Breakthrough',
  'BuffEnd',
  'BuffStart',
  'ChestIncoming',
  'Death',
  'Emote',
  'GateWarn',
  'Hit',
  'ItemExpired',
  'ItemPicked',
  'KillFeed',
  'LevelUp',
  'MonsterDie',
  'MonsterHurt',
  'ObstacleBroken',
  'ObstacleRegrow',
  'ObstacleTooHard',
  'RealmDrop',
  'RingFlip',
  'Score',
  'SkillCast',
  'SkillCharge',
  'SkillEffect',
  'StallDown',
  'StallUp',
  'SwordBreak',
  'TeleportIn',
  'TeleportOut',
  'TienKhiPicked',
] as const;

/** Khớp `netproto.MonsterKinds` và `config.monsters.kinds`. */
export const MONSTER_KINDS = ['small', 'medium', 'large', 'boss'] as const;
