import { Assets, Rectangle, Texture } from 'pixi.js';

// Asset paths follow docs/ART_PROMPTS.md §13. Anything missing falls back to placeholders,
// so dropping a PNG into assets/ and re-running `artgen prep` is enough to use real art.

interface AtlasJson {
  cell: number;
  pivot_x: number;
  pivot_y: number;
  frames: Record<string, { x: number; y: number; w: number; h: number }>;
  anims: Record<string, { frames: string[]; fps: number; loop: boolean }>;
}

export interface AnimSet {
  anims: Map<string, { textures: Texture[]; fps: number; loop: boolean }>;
  pivotX: number;
  pivotY: number;
}

class AssetStore {
  private files = new Set<string>();
  private textures = new Map<string, Texture>();
  private atlases = new Map<string, AnimSet>();

  async init(): Promise<void> {
    try {
      const res = await fetch('/assets/manifest.json', { cache: 'no-cache' });
      if (res.ok) {
        const m = (await res.json()) as { files: string[] };
        this.files = new Set(m.files);
      }
    } catch (err) {
      console.warn('no asset manifest, using placeholders', err);
    }
  }

  has(path: string): boolean {
    return this.files.has(path);
  }

  /** Loads every listed single image plus atlases. Missing files are simply skipped. */
  async preload(onProgress?: (p: number) => void): Promise<void> {
    const singles = [...this.files].filter((f) => f.endsWith('.png') && !f.endsWith('atlas.png') && !f.includes('ref_sheet'));
    const atlases = [...this.files].filter((f) => f.endsWith('atlas.json'));
    const total = singles.length + atlases.length;
    let done = 0;
    const tick = (): void => {
      done++;
      onProgress?.(total ? done / total : 1);
    };
    await Promise.all([
      ...singles.map(async (f) => {
        await this.loadTexture(f);
        tick();
      }),
      ...atlases.map(async (f) => {
        await this.loadAtlas(f);
        tick();
      }),
    ]);
  }

  private async loadTexture(path: string): Promise<Texture | null> {
    try {
      const tex = (await Assets.load(`/assets/${path}`)) as Texture;
      this.textures.set(path, tex);
      return tex;
    } catch (err) {
      console.warn('texture failed', path, err);
      return null;
    }
  }

  private async loadAtlas(jsonPath: string): Promise<void> {
    const dir = jsonPath.slice(0, jsonPath.lastIndexOf('/'));
    try {
      const res = await fetch(`/assets/${jsonPath}`);
      const atlas = (await res.json()) as AtlasJson;
      const base = await this.loadTexture(`${dir}/atlas.png`);
      if (!base) {
        return;
      }
      const set: AnimSet = { anims: new Map(), pivotX: atlas.pivot_x, pivotY: atlas.pivot_y };
      for (const [name, a] of Object.entries(atlas.anims)) {
        const textures = a.frames.map((f) => {
          const r = atlas.frames[f];
          return new Texture({ source: base.source, frame: new Rectangle(r.x, r.y, r.w, r.h) });
        });
        set.anims.set(name, { textures, fps: a.fps, loop: a.loop });
      }
      this.atlases.set(dir, set);
    } catch (err) {
      console.warn('atlas failed', jsonPath, err);
    }
  }

  texture(path: string): Texture | null {
    return this.textures.get(path) ?? null;
  }

  /** Frames like vfx/clash_spark/clash_spark_00.png … in order, or [] if missing. */
  frames(dir: string, anim: string): Texture[] {
    const out: Texture[] = [];
    for (let i = 0; i < 99; i++) {
      const t = this.textures.get(`${dir}/${anim}_${String(i).padStart(2, '0')}.png`);
      if (!t) {
        break;
      }
      out.push(t);
    }
    return out;
  }

  atlas(dir: string): AnimSet | null {
    return this.atlases.get(dir) ?? null;
  }
}

export const assets = new AssetStore();
