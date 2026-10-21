/** h creates an element with a class and optional text/children. */
export function h<K extends keyof HTMLElementTagNameMap>(
  tag: K,
  cls = '',
  ...children: (Node | string)[]
): HTMLElementTagNameMap[K] {
  const el = document.createElement(tag);
  if (cls) {
    el.className = cls;
  }
  for (const c of children) {
    el.append(c);
  }
  return el;
}

export function button(label: string, cls: string, onClick: () => void): HTMLButtonElement {
  const b = h('button', cls, label);
  b.addEventListener('click', onClick);
  return b;
}

/** Screen shows exactly one overlay at a time. */
export class Screens {
  private current: HTMLElement | null = null;

  constructor(private readonly root: HTMLElement) {}

  show(el: HTMLElement | null): void {
    this.current?.remove();
    this.current = el;
    if (el) {
      this.root.appendChild(el);
    }
  }
}

/** portrait shows the character's first idle frame, or a coloured circle placeholder. */
export function portrait(char: string, has: (p: string) => boolean, color: string): HTMLElement {
  if (has(`characters/${char}/atlas.png`)) {
    const img = h('img', 'portrait');
    img.src = `/assets/characters/${char}/frames/idle_00.png`;
    img.alt = '';
    return img;
  }
  const d = h('div', 'portrait placeholder');
  d.style.background = color;
  return d;
}
