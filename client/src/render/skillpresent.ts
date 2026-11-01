import type { Container } from 'pixi.js';

import { sound, type SfxName } from '../audio/sound';
import { type GameConfig, skillRange } from '../config';
import { type GameEvent, STATUS, statusOn, type StatusEv } from '../net/protocol';
import { STR } from '../ui/strings';
import { EdgeArrows } from './edgearrow';
import type { FloatTextLayer } from './floattext';
import type { PlayerSprite } from './player';
import { SkillFx } from './skillfx';
import { SkillVfx } from './skillvfx';
import { type StatusIconKey, STATUS_ICON_KEYS } from './statusicons';
import { statusTracker } from './statustrack';
import type { Vfx } from './vfx';

/**
 * SkillPresenter: toàn bộ phần NHÌN THẤY của kỹ năng (spec v6.36 §6.5) — nhận event
 * `SkillCast`, `SkillEffect`, `StatusStart/StatusEnd`, `StallUp` và vẽ lớp ①②⑤ + rung,
 * mũi tên mép màn hình. Lớp ③④ trên từng người nằm ở `StatusView` (đọc `statusTracker`).
 *
 * Tách khỏi `GameScene` vì scene đã lo đủ thứ (camera, bản đồ, pool view…); mười kỹ năng ×
 * năm lớp nhét thêm vào đó là một hàm `switch` 300 dòng không ai đọc nổi.
 */

export interface Caster {
  x: number;
  y: number;
  f: number;
  sc: number;
}

/** SkillHost: những gì presenter cần mượn của scene. */
export interface SkillHost {
  myId(): number;
  casterAt(id: number): Caster | null;
  sprite(id: number): PlayerSprite | undefined;
  toScreen(x: number, y: number): { x: number; y: number };
  screen(): { w: number; h: number };
  readonly floatText: FloatTextLayer;
  readonly vfx: Vfx;
}

/** Tiếng lúc dùng kỹ năng (lớp ①), theo nhân vật — tệp do `tools/make_sfx.py` sinh. */
const CAST_SFX: Record<string, SfxName> = {
  a_hon: 'skill_banh_bao',
  ba_tam_loa: 'skill_loa',
  ong_ba_vong: 'skill_hat_ru',
  dai_ca_dep: 'skill_dan_em',
  map_tra_sua: 'skill_tran_chau',
  co_ba_ganh: 'skill_bay_hang',
  su_phu_mi: 'skill_hut_mi',
  be_ti: 'skill_gong_co',
  shipper: 'skill_gom_don',
  cu_rua: 'skill_rut_co',
};

const C = {
  loa: 0xffd23f,
  ru: 0xb48ef0,
  steal: 0x9b59b6,
  pearl: 0x6b4a2f,
  noodle: 0xf2c14e,
  stall: 0xe08f3c,
  shell: 0x7ad7b0,
  brace: 0xff8c2a,
  smoke: 0xeeeeee,
  dust: 0xc9a46a,
} as const;

/** Rung khung màn hình người bị trúng: nhẹ hơn hẳn rung khi bị chém thân. */
const HIT_SHAKE_PX = 8;
const HIT_SHAKE_SECS = 0.2;
/** Đàn em "chạy từ mép vào": vệt lướt dài ngần này đơn vị thế giới tới chỗ server đặt. */
const DASH_IN_LEN = 600;

export interface StatusChip {
  key: StatusIconKey;
  secs: number;
  frac: number;
}

export class SkillPresenter {
  private readonly fx: SkillFx;
  private readonly vfx = new SkillVfx();
  readonly arrows = new EdgeArrows();
  private readonly later: { at: number; run: () => void }[] = [];
  private readonly dashes = new Map<number, { x: number; y: number; at: number }>();
  /** lastWave gộp nhiều event cùng một lần cast (server gửi mỗi nạn nhân một event). */
  private lastWave = '';
  private pruneAt = 0;

  constructor(private readonly cfg: GameConfig, private readonly host: SkillHost) {
    this.fx = new SkillFx(cfg);
  }

  /** layers: hai lớp nằm trong camera (chỉ báo phạm vi dưới, hiệu ứng ngắn trên). */
  get layers(): Container[] {
    return [this.fx.layer, this.vfx.layer];
  }

  /** handle trả `true` nếu event là của kỹ năng (scene khỏi xét tiếp). */
  handle(e: GameEvent, now: number): boolean {
    switch (e.k) {
      case 'SkillCast':
        this.onCast(e, now);
        return true;
      case 'SkillEffect':
        this.onEffect(e, now);
        return true;
      case 'StatusStart':
        this.onStatusStart(e, now);
        return true;
      case 'StatusEnd':
        this.onStatusEnd(e, now);
        return true;
      case 'StallUp':
        this.vfx.stallFlash(e.x, e.y, e.f ?? 0, now);
        return true;
      case 'StallDown':
        this.vfx.puff(e.x, e.y, 120, C.dust, now);
        return true;
      default:
        return false;
    }
  }

  update(now: number): void {
    this.fx.update(now);
    this.vfx.update(now);
    this.arrows.update(now);
    for (let i = this.later.length - 1; i >= 0; i--) {
      if (now >= this.later[i].at) {
        this.later[i].run();
        this.later.splice(i, 1);
      }
    }
    this.resolveDashes(now);
    if (now >= this.pruneAt) {
      this.pruneAt = now + 1;
      statusTracker.prune(now);
    }
  }

  clear(): void {
    this.fx.clear();
    this.vfx.clear();
    this.arrows.clear();
    this.later.length = 0;
    this.dashes.clear();
    statusTracker.clear();
  }

  /** chips: icon + giây còn lại của CHÍNH MÌNH cho HUD (dưới ảnh đại diện). */
  chips(id: number, st: number, now: number): StatusChip[] {
    const out: StatusChip[] = [];
    for (const key of STATUS_ICON_KEYS) {
      const codes: StatusEv[] = key === 'stun' ? ['stun', 'slip'] : [key];
      const on = codes.some((c) => isBit(c) ? statusOn(st, c as (typeof STATUS)[number]) : statusTracker.active(id, c, now));
      if (!on) {
        continue;
      }
      let secs = 0;
      let frac = 1;
      for (const c of codes) {
        const l = statusTracker.left(id, c, now);
        if (l && l.secs > secs) {
          secs = l.secs;
          frac = l.frac;
        }
      }
      out.push({ key, secs, frac });
    }
    return out;
  }

  // ── lớp ① ──────────────────────────────────────────────────────────────────────────

  private onCast(e: GameEvent, now: number): void {
    const sp = this.host.sprite(e.a ?? 0);
    // Bong bóng thoại tấu hài khi dùng kỹ năng (spec §6.3).
    if (e.s && STR.skillLines[e.s]) {
      sp?.showBubble(STR.skillLines[e.s], now);
    }
    sp?.playSkill(now);
    const sfx = CAST_SFX[e.s ?? ''];
    if (sfx) {
      sound.play(sfx, e.a === this.host.myId() ? 1 : 0.5);
    }
  }

  // ── lớp ② + hiệu ứng ngắn ─────────────────────────────────────────────────────────

  private onEffect(e: GameEvent, now: number): void {
    const at = this.host.casterAt(e.a ?? 0);
    switch (e.s) {
      case 'trom_kiem_hit':
        this.onSteal(e, at, now);
        return;
      case 'ru_ngu_huy':
        this.host.floatText.show(STR.skillCancelled, e.x, e.y, now, 0xffb3b3);
        return;
      case 'dan_em_gone':
        this.host.vfx.play('death_smoke', e.x, e.y, 160, now);
        this.host.floatText.show(STR.minionBye, e.x, e.y - 60, now, 0xe0c3f5);
        return;
      default:
    }
    // Từ đây trở xuống cần người dùng trong khung nhìn: vùng vẽ theo U của CHÍNH họ —
    // không biết scale thì thà không vẽ còn hơn vẽ một vòng sai cỡ.
    if (at && at.sc > 0) {
      this.onCasterEffect(e, at, now);
    }
  }

  private onCasterEffect(e: GameEvent, at: Caster, now: number): void {
    const r = this.cfg.player.body_r * at.sc;
    switch (e.s) {
      case 'banh_bao':
        this.onBanhBao(e, at, now);
        break;
      case 'loa':
        this.onLoa(e, at, now);
        break;
      case 'tran_chau': {
        const sk = this.cfg.skills;
        this.fx.cone(e.x, e.y, e.f ?? 0, at.sc, 'tran_chau_cone', sk.tran_chau_cone_deg, C.pearl, now);
        this.vfx.pearls(e.x, e.y, e.f ?? 0, skillRange(this.cfg, at.sc, 'tran_chau_cone'), sk.tran_chau_cone_deg, now);
        break;
      }
      case 'ru_ngu_hat':
        // Vòng BÁO TRƯỚC trong lúc hát: ai còn nửa giây để chạy ra khỏi 4 U thì thoát.
        this.fx.circle(at.x, at.y, at.sc, 'ru_ngu_r', C.ru, now);
        this.vfx.notes(at.x, at.y, skillRange(this.cfg, at.sc, 'ru_ngu_r'), C.ru, now);
        break;
      case 'ru_ngu':
        if (this.once(`ru:${e.a}`, now)) {
          this.fx.circle(at.x, at.y, at.sc, 'ru_ngu_r', C.ru, now);
          sound.play('skill_sleep', 0.5);
        }
        break;
      case 'trom_kiem':
        this.fx.circle(at.x, at.y, at.sc, 'ru_ngu_r', C.steal, now);
        break;
      case 'hut_mi':
      case 'hut_mi_miss':
        this.onHutMi(e, at, now);
        break;
      case 'rut_co_burst':
        this.fx.circle(e.x, e.y, at.sc, 'rut_co_wave', C.shell, now);
        this.vfx.wave(e.x, e.y, skillRange(this.cfg, at.sc, 'rut_co_wave'), C.shell, now);
        this.host.sprite(e.a ?? 0)?.showBubble(STR.skillShellOut, now);
        break;
      case 'gong_co':
        this.vfx.puff(at.x, at.y, r * 2, C.brace, now);
        break;
      case 'gom_don':
      case 'dan_em':
        this.vfx.puff(at.x, at.y + r * 0.75, r * 2, C.dust, now);
        break;
      case 'rut_co':
        this.vfx.puff(at.x, at.y, r * 2, C.shell, now);
        break;
      default:
        break;
    }
  }

  /** Bánh Bao: hơi nóng + kiếm bay từ cái bánh vào vòng + "+N kiếm"; ⑤ ợ một tiếng, khói tan. */
  private onBanhBao(e: GameEvent, at: Caster, now: number): void {
    const r = this.cfg.player.body_r * at.sc;
    const sp = this.host.sprite(e.a ?? 0);
    const bx = at.x + r * 0.7;
    const by = at.y - r * 1.4;
    this.vfx.bun(bx, by, r, now);
    const n = e.n ?? 0;
    if (n > 0) {
      const reach = sp?.ringReach || r * 3;
      this.vfx.swords({ x0: bx, y0: by, x1: at.x, y1: at.y - reach }, n, 0xc9a227, now + 0.2, 0.25);
      this.host.floatText.show(STR.swordsGained(n), at.x, at.y - r * 2, now, 0x9cff9c);
    }
    this.later.push({
      at: now + 0.95,
      run: () => {
        const p = this.host.casterAt(e.a ?? 0) ?? at;
        this.host.sprite(e.a ?? 0)?.showBubble(STR.skillBurp, now + 0.95);
        this.vfx.puff(p.x + r * 0.5, p.y - r * 1.5, r, C.smoke, now + 0.95);
      },
    });
  }

  /** Loa Phường: sóng âm lan 6 U (một lần mỗi cast) + vòng kiếm từng nạn nhân loé khi đổi chiều. */
  private onLoa(e: GameEvent, at: Caster, now: number): void {
    if (this.once(`loa:${e.a}`, now)) {
      this.fx.circle(at.x, at.y, at.sc, 'loa_radius', C.loa, now);
      this.vfx.wave(at.x, at.y, skillRange(this.cfg, at.sc, 'loa_radius'), C.loa, now);
    }
    const reach = this.host.sprite(e.b ?? 0)?.ringReach ?? 0;
    if (reach > 0) {
      this.vfx.wave(e.x, e.y, reach * 1.1, 0xffffff, now, 0.25);
    }
  }

  /** Hút Mì: nón 60° loé + sợi mì nối hai người, kiếm bay dọc sợi mì; hụt thì mì rơi. */
  private onHutMi(e: GameEvent, at: Caster, now: number): void {
    const miss = e.s === 'hut_mi_miss';
    const face = miss ? e.f ?? at.f : at.f;
    this.fx.cone(at.x, at.y, face, at.sc, 'hut_mi_cone', this.cfg.skills.hut_mi_cone_deg, C.noodle, now);
    if (miss) {
      const len = skillRange(this.cfg, at.sc, 'hut_mi_cone');
      this.vfx.noodle({ x0: at.x, y0: at.y, x1: at.x + Math.cos(face) * len, y1: at.y + Math.sin(face) * len }, now, true);
      return;
    }
    const r = this.cfg.player.body_r * at.sc;
    this.vfx.noodle({ x0: at.x, y0: at.y - r * 0.6, x1: e.x, y1: e.y - r * 0.6 }, now, false);
    const n = e.n ?? 0;
    this.vfx.swords({ x0: e.x, y0: e.y - r * 0.6, x1: at.x, y1: at.y - r * 0.6 }, n, C.noodle, now + 0.12);
    this.host.floatText.show(STR.swordsStolen(n), e.x, e.y, now, 0xffe14d);
    this.host.floatText.show(STR.swordsGained(n), at.x, at.y - r * 1.6, now + 0.3, 0x9cff9c);
  }

  /** Trộm Kiếm: mỗi lần trộm được — kiếm bay vệt tím từ nạn nhân về vòng Ông Ba. */
  private onSteal(e: GameEvent, at: Caster | null, now: number): void {
    const n = e.n ?? 0;
    this.host.floatText.show(STR.swordsStolen(n), e.x, e.y, now, 0xffe14d);
    if (!at) {
      return;
    }
    this.vfx.swords({ x0: e.x, y0: e.y, x1: at.x, y1: at.y }, n, C.steal, now, 0.2);
    this.host.floatText.show(STR.swordsGained(n), at.x, at.y - this.cfg.player.body_r * at.sc * 1.6, now + 0.3, 0xd7a8f0);
  }

  // ── trạng thái: rung, mũi tên, hiệu ứng tan ──────────────────────────────────────

  private onStatusStart(e: GameEvent, now: number): void {
    const id = e.a ?? 0;
    const src = e.b ?? 0;
    const name = statusTracker.start(id, e.n ?? 0, e.f ?? 0, src, now);
    if (name === 'minion') {
      this.dashes.set(id, { x: e.x, y: e.y, at: now });
      return;
    }
    const me = this.host.myId();
    if (id !== me || src === 0 || src === me) {
      return;
    }
    this.host.vfx.shake(HIT_SHAKE_PX, HIT_SHAKE_SECS, now);
    const sc = this.host.screen();
    const p = this.host.toScreen(e.x, e.y);
    this.arrows.show(p.x, p.y, sc.w, sc.h, now);
  }

  private onStatusEnd(e: GameEvent, now: number): void {
    const id = e.a ?? 0;
    const name = statusTracker.end(id, e.n ?? 0);
    const sc = this.host.casterAt(id)?.sc ?? 1;
    const r = this.cfg.player.body_r * sc;
    const loud = id === this.host.myId() ? 0.6 : 0.3;
    switch (name) {
      case 'blind':
        this.vfx.fall(e.x, e.y, this.host.sprite(id)?.ringReach || r * 3, 0x2b1a10, now);
        break;
      case 'brace':
        this.vfx.puff(e.x, e.y - r * 0.5, r * 1.6, 0xffffff, now); // xì hơi
        break;
      case 'ghost':
        this.host.vfx.play('death_smoke', e.x, e.y + r * 0.6, r * 2.4, now); // phanh gấp, khói bánh xe
        break;
      case 'steal':
        this.host.sprite(id)?.showBubble(STR.skillYawn, now);
        this.vfx.puff(e.x, e.y, r * 1.6, C.steal, now);
        break;
      case 'sleep':
        this.vfx.puff(e.x, e.y - r * 1.2, r, 0xbcd4ff, now);
        break;
      default:
        return;
    }
    sound.play('skill_end', loud);
  }

  /** Đàn em hiện ra: vệt lướt từ phía đại ca kéo dài ra ngoài, kết thúc đúng chỗ server đặt. */
  private resolveDashes(now: number): void {
    for (const [id, d] of this.dashes) {
      const m = this.host.casterAt(id);
      if (!m) {
        if (now - d.at > 1) {
          this.dashes.delete(id);
        }
        continue;
      }
      this.dashes.delete(id);
      const dx = m.x - d.x;
      const dy = m.y - d.y;
      const len = Math.hypot(dx, dy) || 1;
      this.vfx.streak({ x0: m.x + (dx / len) * DASH_IN_LEN, y0: m.y + (dy / len) * DASH_IN_LEN, x1: m.x, y1: m.y }, C.dust, now);
      this.vfx.puff(m.x, m.y + this.cfg.player.body_r * m.sc * 0.75, this.cfg.player.body_r * m.sc * 1.6, C.dust, now + 0.3);
    }
  }

  /** once: lần đầu trong cùng một nhịp (50 ms) của khoá này — gộp event từng nạn nhân. */
  private once(key: string, now: number): boolean {
    const k = `${key}:${Math.round(now * 20)}`;
    if (this.lastWave === k) {
      return false;
    }
    this.lastWave = k;
    return true;
  }
}

function isBit(c: StatusEv): boolean {
  return (STATUS as readonly string[]).includes(c);
}
