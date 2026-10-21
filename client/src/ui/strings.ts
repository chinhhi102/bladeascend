// Every piece of text shown to players lives here (CLAUDE.md rule 5).

export const STR = {
  title: 'Kiếm Xoay Tu Tiên',
  subtitle: 'Dân phố tấu hài tu tiên bằng kiếm bay',
  namePlaceholder: 'Nhập tên (2–16 ký tự)',
  pickCharacter: 'Chọn nhân vật',
  play: 'Vào chơi',
  leaderboard: 'Bảng xếp hạng',
  findingRoom: 'Đang tìm phòng...',
  connecting: 'Đang kết nối máy chủ...',
  disconnected: 'Mất kết nối với máy chủ. Đang thử lại...',
  badName: 'Tên phải dài 2–16 ký tự.',
  controlsPc: 'Di chuyển: WASD / phím mũi tên / giữ chuột. Không có nút đánh — kiếm tự xoay!',
  controlsMobile: 'Kéo ngón tay để di chuyển. Kiếm tự xoay!',
  debugHelp: 'Phím test: E +100 EXP · R lên cảnh giới · K +1 kiếm · 1–7 buff · X tự chết',

  // HUD
  realmTier: (realm: string, tier: number) => `${realm} tầng ${tier}`,
  swords: (n: number, max: number) => `Kiếm ${n}/${max}`,
  roomTop: 'Top phòng',
  yourRank: (rank: number, total: number) => `Hạng của bạn: ${rank}/${total}`,
  ping: (ms: number) => `Ping ${ms}ms`,
  chestIncoming: (secs: number) => `Rương Báu sắp xuất hiện trong ${secs}s!`,
  maxLevel: 'Đỉnh cao tu vi',

  // Death / summary screen (spec §7)
  summaryTitle: 'Tổng kết',
  killedBy: (name: string) => `Bị ${name} chém`,
  killedByUnknown: 'Bị chém bởi một cao thủ vô danh',
  peakRealm: 'Cảnh giới cao nhất',
  peakExp: 'Tổng EXP đỉnh',
  kills: 'Số người đã chém',
  survive: 'Thời gian sống',
  roomRank: 'Hạng trong phòng',
  playAgain: 'Chơi tiếp',
  changeCharacter: 'Đổi nhân vật',
  back: 'Quay lại',
  deathJokes: [
    'Bạn đã được tiễn về quê nuôi cá.',
    'Kiếm gãy, người bay, mộng tu tiên tạm hoãn.',
    'Tu tiên không thành, về bán bánh bao tiếp.',
    'Ăn một nhát kiếm, ngộ ra chân lý: chạy nhanh hơn.',
    'Đạo tâm vỡ vụn như bánh tráng.',
    'Thôi thì kiếp sau làm quái cho đỡ khổ.',
  ],
  secs: (s: number) => {
    const m = Math.floor(s / 60);
    const r = Math.floor(s % 60);
    return m > 0 ? `${m} phút ${r} giây` : `${r} giây`;
  },

  // Global leaderboard (spec §11)
  lbTitle: 'Bảng xếp hạng',
  lbToday: 'Hôm nay',
  lbWeek: 'Tuần này',
  lbAll: 'Mọi thời đại',
  lbExp: 'Tu Vi Đỉnh Cao',
  lbKills: 'Sát Thần',
  lbEmpty: 'Chưa có ai ghi danh. Vào chém đi!',
  lbName: 'Tên',
  lbRealm: 'Cảnh giới',
  lbLoading: 'Đang tải...',
  lbError: 'Không tải được bảng xếp hạng.',

  // Kill feed (spec §7)
  feed: {
    kill: (a: string, b: string) => `${a} vừa tiễn ${b} về quê nuôi cá`,
    realm_drop: (a: string, b: string, realm: string) => `${b} bị ${a} chém rớt xuống ${realm}`,
    breakthrough: (a: string, realm: string) => `${a} đột phá ${realm}, cả phòng run rẩy`,
    gate_kill: (a: string, b: string) => `${b} chui cổng ra liền bị ${a} chém, đen thôi đỏ quên đi`,
  },

  // Emote bubbles (ART_PROMPTS §10)
  emotes: {
    sword_lost: ['Kiếm của taaa!'],
    shield_break: ['Khiên ơi!', 'Vỡ rồi!'],
    chest: ['Trúng mánh!', 'Giàu rồi!'],
    revealed: ['Lộ rồi!', 'Ối!'],
    run: ['AAAAA!', 'Tránh ra!'],
    chase: ['Hehehe...', 'Đứng lại đó!'],
    flee: ['Cứu tôi với!', 'Mẹ ơi!'],
    hit: ['Ối giời!', 'Đau quá!'],
    breakthrough: ['Ta đã ngộ rồi!!!'],
    teleport_in: ['Đây là đâu???'],
    bush: ['Suỵt...'],
    shocked: ['Hả???', 'Trời đất ơi!', 'Ghê vậy!'],
  } as Record<string, string[]>,

  characters: {
    a_hon: 'A Hớn Bánh Bao',
    ba_tam_loa: 'Bà Tám Loa',
    ong_ba_vong: 'Ông Ba Võng',
    dai_ca_dep: 'Đại Ca Dép Tổ Ong',
    map_tra_sua: 'Thằng Mập Trà Sữa',
    co_ba_ganh: 'Cô Ba Gánh Hàng',
    su_phu_mi: 'Sư Phụ Mì Gói',
    be_ti: 'Bé Tí Cơ Bắp',
    shipper: 'Anh Shipper Tia Chớp',
    cu_rua: 'Cụ Rùa Thiền',
  } as Record<string, string>,

  buffs: {
    spin: 'Tốc Kiếm',
    haste: 'Thần Hành',
    magnet: 'Nam Châm',
    long: 'Trường Kiếm',
    shield: 'Kim Cang',
    exp2: 'Song Bội Đan',
    stealth: 'Ẩn Thân Phù',
  } as Record<string, string>,

  rotateHint: 'Xoay ngang điện thoại để nhìn rộng hơn',
};

export function pick<T>(list: T[]): T {
  return list[Math.floor(Math.random() * list.length)];
}
