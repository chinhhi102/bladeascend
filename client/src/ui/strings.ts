// Mọi chữ người chơi thấy (CLAUDE.md luật 5). Từ GAME_SPEC §19 chữ nằm trong
// `src/i18n/{lang}.json`; file này chỉ DỰNG `STR` từ từ điển đã chọn — mẫu câu có `{tên}`
// thành hàm, còn lại giữ nguyên. Thêm chữ mới: thêm khoá vào `i18n/vi.json` (và để trống ở
// các ngôn ngữ khác — `npm run check:i18n` nhắc), rồi dùng `STR.khoá` như cũ.
import { splitName } from '../../brand-split';
import { DICT as d, LANG, fmt, num, pickFmt } from '../i18n';

/** Thương hiệu do Vite cắm vào lúc build từ `config/game.json` — GAME_SPEC §20. */
declare const __BRAND__: { name: string; domain: string; name_vi: string; tagline_vi: string };

export const BRAND = __BRAND__;

const IS_VI = LANG === 'vi';

type Rec = Record<string, string>;

export const STR = {
  ...d,
  /** Tên game: người Việt thấy "Đạo Hữu Xin Dừng Bước", ngôn ngữ khác thấy "BladeAscend" (§20). */
  title: IS_VI ? BRAND.name_vi : BRAND.name,
  /**
   * Tên game chia hai cụm để NGẮT DÒNG CÓ KIỂM SOÁT (chủ dự án chốt 1/11): hai `<span>`
   * không-ngắt nên không bao giờ có chữ lẻ một mình. Cắt theo số chữ từ `config.brand`.
   */
  titleParts: IS_VI ? splitName(BRAND.name_vi) : [BRAND.name],
  subtitle: d.tagline || BRAND.tagline_vi,
  feedbackKinds: d.feedbackKinds as Rec,
  skillNames: d.skillNames as Rec,
  skillDesc: d.skillDesc as Rec,
  skillLines: d.skillLines as Rec,
  characters: d.characters as Rec,
  charactersShort: d.charactersShort as Rec,
  charNick: d.charNick as Rec,
  buffs: d.buffs as Rec,
  scoreSources: d.scoreSources as Rec,
  emotes: d.emotes as Record<string, string[]>,
  /** Tên cảnh giới đã dịch; rỗng = dùng tên trong config (tiếng Việt). */
  realmNames: d.realmNames as string[],
  needRealm: (realm: string): string => fmt(d.needRealm, { realm }),
  swordsStolen: (n: number): string => fmt(d.swordsStolen, { n }),
  skillCooldown: (secs: number, n: number): string => fmt(d.skillCooldown, { secs, n }),
  skillWait: (secs: number): string => fmt(d.skillWait, { secs }),
  zoneEntered: (name: string): string => fmt(d.zoneEntered, { name }),
  loadChar: (name: string): string => fmt(d.loadChar, { name }),
  loadZone: (zone: string): string => fmt(d.loadZone, { zone }),
  loadMB: (done: number, total: number): string => fmt(d.loadMB, { done: done.toFixed(1), total: total.toFixed(1) }),
  loadBg: (pct: number): string => fmt(d.loadBg, { pct: Math.round(pct * 100) }),
  qualitySet: (name: string): string => fmt(d.qualitySet, { name }),
  qualityAuto: (name: string): string => fmt(d.qualityAuto, { name }),
  burnHint: (mult: number): string => fmt(d.burnHint, { mult }),
  skillBtnHint: (isTouch: boolean): string => (isTouch ? d.skillBtnHintTouch : d.skillBtnHintPc),
  realmTier: (realm: string, tier: number): string => fmt(d.realmTier, { realm, tier }),
  // Có dấu phân cách nghìn (v6.4): Tiên khí ×20, năm chữ số dính liền thì đọc nhầm hàng.
  tienKhi: (have: number, need: number): string => fmt(d.tienKhi, { have: num(Math.floor(have)), need: num(need) }),
  swords: (n: number, max: number): string => fmt(d.swords, { n, max }),
  board: {
    ...d.board,
    players: (n: number): string => fmt(d.board.players, { n }),
    /** Dòng ghim của mình khi ngoài top 10, và hàng thu gọn trên điện thoại: "Bạn #14". */
    youRank: (rank: number): string => (rank > 0 ? fmt(d.board.youRank, { rank }) : d.board.youNoRank),
  },
  yourRank: (rank: number, total: number): string => fmt(d.yourRank, { rank, total }),
  ping: (ms: number): string => fmt(d.ping, { ms }),
  chestIncoming: (secs: number): string => fmt(d.chestIncoming, { secs }),
  killedBy: (name: string): string => fmt(d.killedBy, { name }),
  secs: (s: number): string => {
    const m = Math.floor(s / 60);
    const r = Math.floor(s % 60);
    return m > 0 ? fmt(d.secsMin, { m, s: r }) : fmt(d.secsOnly, { s: r });
  },
  scoreHud: (n: number): string => fmt(d.scoreHud, { n: num(n) }),
  scorePop: (n: number): string => fmt(d.scorePop, { n: num(n) }),
  // Kill feed (spec §7) — server gửi MÃ + tham số, client dịch theo ngôn ngữ người xem (§19).
  feed: {
    heaven_gate: (a: string, title: string): string => fmt(d.feed.heaven_gate, { a, title }),
    kill: (a: string, b: string): string => pickFmt(d.feed.kill, { a, b }),
    realm_drop: (a: string, b: string, realm: string): string => pickFmt(d.feed.realm_drop, { a, b, realm }),
    breakthrough: (a: string, realm: string): string => pickFmt(d.feed.breakthrough, { a, realm }),
    serverTag: (rank: number): string => fmt(d.feed.serverTag, { rank }),
    tien_ascend: (a: string, realm: string): string => fmt(d.feed.tien_ascend, { a, realm }),
    title_gain: (a: string): string => fmt(d.feed.title_gain, { a }),
    title_lose: (a: string): string => fmt(d.feed.title_lose, { a }),
    gate_kill: (a: string, b: string): string => pickFmt(d.feed.gate_kill, { a, b }),
  },
  // Bảng điều khiển mờ vài giây đầu trận (spec §18.1).
  guideKeys: (isTouch: boolean): { key: string; what: string }[] => (isTouch ? d.guideKeysTouch : d.guideKeysPc),
};

export function pick<T>(list: T[]): T {
  return list[Math.floor(Math.random() * list.length)];
}
