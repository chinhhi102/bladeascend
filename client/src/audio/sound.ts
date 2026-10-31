import { Howl, Howler } from 'howler';

import { assets } from '../render/assets';

// SFX names from docs/ART_PROMPTS.md §12. A sound plays only if assets/audio/sfx/<name>.{ogg,mp3,wav} exists.
const SFX = [
  'sword_clash', 'sword_hit_body', 'pickup_sword', 'pickup_buff', 'pickup_exp', 'buff_end', 'sword_break',
  'shield_break', 'chest_incoming', 'chest_open', 'obstacle_hit', 'obstacle_break', 'obstacle_regrow',
  'gate_in', 'gate_out', 'gate_warn', 'monster_hurt', 'monster_pop', 'levelup', 'breakthrough', 'realm_drop',
  'death', 'respawn',
  // Kỹ năng v6.9/v6.10. Thiếu tệp thì `sound.play` im lặng bỏ qua, nên thêm tên vào đây
  // trước khi có âm thanh là an toàn — và `assets:report` sẽ liệt kê chúng là còn thiếu.
  'skill_hat_ru', 'skill_sleep', 'skill_hut_mi', 'skill_gom_don', 'skill_bay_hang',
] as const;

export type SfxName = (typeof SFX)[number];

const EXTS = ['ogg', 'mp3', 'wav'];
const MIN_GAP_MS = 60; // the same sound at most every 60 ms

class Sound {
  private sfx = new Map<string, Howl>();
  private lastPlayed = new Map<string, number>();
  private bgm: Howl | null = null;
  private pickupChain = 0;
  private pickupAt = 0;
  muted = false;

  init(): void {
    // Địa chỉ lấy qua `assets.url`: tệp trong `assets/dist` có hash trong tên, nên ghép
    // tay `/assets/audio/sfx/<tên>.ogg` là 404 im lặng — âm thanh không phát và không báo gì.
    for (const name of SFX) {
      const src = EXTS.map((e) => assets.url(`audio/sfx/${name}.${e}`)).filter(Boolean);
      if (src.length) {
        this.sfx.set(name, new Howl({ src, volume: 0.5 }));
      }
    }
    const bgm = EXTS.map((e) => assets.url(`audio/bgm/main.${e}`)).filter(Boolean);
    if (bgm.length) {
      this.bgm = new Howl({ src: bgm, loop: true, volume: 0.25, html5: true });
    }
  }

  /** startMusic must run from a user gesture (mobile autoplay rules); Howler unlocks audio then. */
  startMusic(): void {
    if (this.bgm && !this.bgm.playing()) {
      this.bgm.play();
    }
  }

  play(name: SfxName, volume = 1): void {
    const h = this.sfx.get(name);
    if (!h || this.muted) {
      return;
    }
    const now = performance.now();
    if (now - (this.lastPlayed.get(name) ?? 0) < MIN_GAP_MS) {
      return;
    }
    this.lastPlayed.set(name, now);
    const id = h.play();
    h.volume(0.5 * volume, id);
    if (name === 'pickup_exp') {
      // Pitch rises while collecting quickly (ART_PROMPTS §12).
      this.pickupChain = now - this.pickupAt < 400 ? Math.min(this.pickupChain + 1, 12) : 0;
      this.pickupAt = now;
      h.rate(1 + this.pickupChain * 0.05, id);
    }
  }

  toggleMute(): boolean {
    this.muted = !this.muted;
    Howler.mute(this.muted);
    return this.muted;
  }
}

export const sound = new Sound();
