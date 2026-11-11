import type { GameConfig } from '../config';
import { STR } from './strings';

/**
 * Mô tả kỹ năng ở bảng chọn nhân vật — **chữ từ i18n, SỐ từ config** (GAME_SPEC §6.6 v6.48).
 *
 * Vì sao phải tách: ngày 11/11 spec đổi gần như toàn bộ 10 kỹ năng, và mười dòng mô tả
 * trong `vi.json` vẫn ngồi đó kể về bản cũ — "hồi ngay 20% số kiếm", "bịt mắt 5 giây",
 * "chạy nhanh gấp 3". Người chơi đọc bảng chọn rồi vào trận thấy khác hẳn. Chuỗi mô tả
 * nay chỉ có chỗ trống `{tên}`; số điền từ `config/game.json`, nên đổi con số trong config
 * là bảng chọn tự đúng theo — ở cả sáu ngôn ngữ cùng lúc.
 *
 * `scripts/skilldesc-check.ts` chặn build nếu một chuỗi mô tả còn chữ số, hoặc dùng chỗ
 * trống không có ở đây.
 */

/** pct: 0.25 → "25" (số phần trăm, không kèm dấu %). */
function pct(v: number | undefined): string {
  return String(Math.round((v ?? 0) * 100));
}

/** num: bỏ số 0 thừa — 0.8 → "0,8"; 4 → "4". Dấu phẩy thập phân theo ngôn ngữ người xem. */
function num(v: number | undefined): string {
  return (v ?? 0).toLocaleString(undefined, { maximumFractionDigits: 2 });
}

/** u: hệ số phạm vi trong `ranges_u` (số giây / tỉ lệ / số lượng cũng nằm chung khối đó). */
function u(cfg: GameConfig, key: string): number {
  return cfg.skills.ranges_u?.[key] ?? 0;
}

/**
 * VARS: chỗ trống của từng nhân vật → giá trị lấy từ config.
 *
 * Mỗi nhân vật một bộ riêng thay vì một bộ chung: cùng chữ "secs" ở A Hớn là thời gian hồi
 * dần, ở Bé Tí là thời gian gồng. Gom chung thì tên nào cũng phải dài dòng để khỏi lẫn, và
 * người dịch sẽ không đoán nổi `{a_hon_regen_secs}` nghĩa là gì.
 */
const VARS: Record<string, (cfg: GameConfig) => Record<string, string>> = {
  a_hon: (c) => ({
    now: pct(c.skills.banh_bao_now), later: pct(c.skills.banh_bao_regen),
    secs: num(c.skills.banh_bao_regen_secs),
  }),
  ba_tam_loa: (c) => ({ channel: num(u(c, 'loa_channel')) }),
  ong_ba_vong: (c) => ({
    channel: num(u(c, 'ru_ngu_channel')), secs: num(u(c, 'ru_ngu_steal_secs')),
    sleep: num(u(c, 'ru_ngu_secs')),
  }),
  dai_ca_dep: (c) => ({
    n: num(c.skills.dan_em_count), secs: num(c.skills.dan_em_secs),
    share: pct(c.skills.dan_em_sword_share),
  }),
  map_tra_sua: (c) => ({
    secs: num(c.skills.tran_chau_secs), giant: num(c.skills.giant_secs),
    mult: num(c.skills.giant_mult),
  }),
  co_ba_ganh: (c) => ({
    n: num(u(c, 'bay_sap_count')), secs: num(u(c, 'bay_sap_secs')),
    tether: num(u(c, 'bay_sap_tether_secs')), slip: num(c.skills.slip_secs),
  }),
  su_phu_mi: (c) => ({ channel: num(u(c, 'hut_mi_channel')), pct: pct(u(c, 'hut_mi_steal')) }),
  be_ti: (c) => ({ secs: num(c.skills.brace_secs), pct: pct(u(c, 'bung_co_lose')) }),
  shipper: (c) => ({
    secs: num(u(c, 'gom_don_secs')), mult: num(c.speed?.cap_gom_don),
    pickup: pct(c.skills.gom_don_pickup),
  }),
  cu_rua: (c) => ({
    secs: num(c.skills.shell_secs), taunt: num(u(c, 'rut_co_taunt_secs')),
  }),
};

/** skillDescVars: bộ chỗ trống của một nhân vật (bài kiểm CI đọc luôn hàm này). */
export function skillDescVars(cfg: GameConfig, char: string): Record<string, string> {
  return VARS[char]?.(cfg) ?? {};
}

/**
 * skillDesc: mô tả đã điền số. Chỗ trống không có giá trị thì GIỮ NGUYÊN `{tên}` — để nó
 * đập vào mắt ngay trên bảng chọn, thay vì biến mất im lặng và người đọc tưởng câu văn vốn
 * thiếu vế đó.
 */
export function skillDesc(cfg: GameConfig, char: string): string {
  const vars = skillDescVars(cfg, char);
  return (STR.skillDesc[char] ?? '').replace(/\{(\w+)\}/g, (m, k: string) => vars[k] ?? m);
}
