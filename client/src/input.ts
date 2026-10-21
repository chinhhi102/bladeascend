// Movement input: WASD / arrow keys / hold mouse on PC, floating virtual joystick on touch (spec §1).

const KEYS: Record<string, [number, number]> = {
  KeyW: [0, -1], ArrowUp: [0, -1],
  KeyS: [0, 1], ArrowDown: [0, 1],
  KeyA: [-1, 0], ArrowLeft: [-1, 0],
  KeyD: [1, 0], ArrowRight: [1, 0],
};

const JOY_RADIUS = 60; // px
const JOY_DEADZONE = 8; // px

export interface MoveInput {
  angle: number;
  moving: boolean;
}

export class Input {
  private readonly held = new Set<string>();
  private mouseDown = false;
  private mouse = { x: 0, y: 0 };
  private touchId: number | null = null;
  private touchStart = { x: 0, y: 0 };
  private touchNow = { x: 0, y: 0 };
  private lastAngle = 0;
  private readonly joyBase: HTMLDivElement;
  private readonly joyKnob: HTMLDivElement;
  enabled = false;

  constructor(private readonly surface: HTMLElement) {
    this.joyBase = document.createElement('div');
    this.joyBase.className = 'joy-base';
    this.joyKnob = document.createElement('div');
    this.joyKnob.className = 'joy-knob';
    this.joyBase.appendChild(this.joyKnob);
    document.body.appendChild(this.joyBase);
    this.bindKeyboard();
    this.bindMouse();
    this.bindTouch();
  }

  static isTouchDevice(): boolean {
    return 'ontouchstart' in window || navigator.maxTouchPoints > 0;
  }

  private bindKeyboard(): void {
    window.addEventListener('keydown', (e) => {
      if (KEYS[e.code] && this.enabled) {
        this.held.add(e.code);
        e.preventDefault();
      }
    });
    window.addEventListener('keyup', (e) => this.held.delete(e.code));
    window.addEventListener('blur', () => {
      this.held.clear();
      this.mouseDown = false;
    });
  }

  private bindMouse(): void {
    this.surface.addEventListener('mousedown', (e) => {
      this.mouseDown = true;
      this.mouse = { x: e.clientX, y: e.clientY };
    });
    window.addEventListener('mousemove', (e) => {
      this.mouse = { x: e.clientX, y: e.clientY };
    });
    window.addEventListener('mouseup', () => {
      this.mouseDown = false;
    });
  }

  private bindTouch(): void {
    const opts = { passive: false } as const;
    this.surface.addEventListener('touchstart', (e) => {
      const t = e.changedTouches[0];
      if (this.touchId === null && t) {
        this.touchId = t.identifier;
        this.touchStart = this.touchNow = { x: t.clientX, y: t.clientY };
        this.showJoystick(true);
      }
      e.preventDefault();
    }, opts);
    this.surface.addEventListener('touchmove', (e) => {
      for (const t of Array.from(e.changedTouches)) {
        if (t.identifier === this.touchId) {
          this.touchNow = { x: t.clientX, y: t.clientY };
        }
      }
      e.preventDefault();
    }, opts);
    const end = (e: TouchEvent): void => {
      for (const t of Array.from(e.changedTouches)) {
        if (t.identifier === this.touchId) {
          this.touchId = null;
          this.showJoystick(false);
        }
      }
    };
    this.surface.addEventListener('touchend', end);
    this.surface.addEventListener('touchcancel', end);
  }

  private showJoystick(on: boolean): void {
    this.joyBase.style.display = on ? 'block' : 'none';
    this.joyBase.style.left = `${this.touchStart.x - JOY_RADIUS}px`;
    this.joyBase.style.top = `${this.touchStart.y - JOY_RADIUS}px`;
  }

  /** read returns the current movement direction (angle in radians, y down). */
  read(): MoveInput {
    if (!this.enabled) {
      return { angle: this.lastAngle, moving: false };
    }
    if (this.touchId !== null) {
      return this.readTouch();
    }
    let dx = 0;
    let dy = 0;
    for (const k of this.held) {
      dx += KEYS[k][0];
      dy += KEYS[k][1];
    }
    if (dx === 0 && dy === 0 && this.mouseDown) {
      dx = this.mouse.x - window.innerWidth / 2;
      dy = this.mouse.y - window.innerHeight / 2;
    }
    if (dx === 0 && dy === 0) {
      return { angle: this.lastAngle, moving: false };
    }
    this.lastAngle = Math.atan2(dy, dx);
    return { angle: this.lastAngle, moving: true };
  }

  private readTouch(): MoveInput {
    const dx = this.touchNow.x - this.touchStart.x;
    const dy = this.touchNow.y - this.touchStart.y;
    const d = Math.hypot(dx, dy);
    const k = d > JOY_RADIUS ? JOY_RADIUS / d : 1;
    this.joyKnob.style.transform = `translate(${dx * k}px, ${dy * k}px)`;
    if (d < JOY_DEADZONE) {
      return { angle: this.lastAngle, moving: false };
    }
    this.lastAngle = Math.atan2(dy, dx);
    return { angle: this.lastAngle, moving: true };
  }
}
