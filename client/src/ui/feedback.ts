import { STR } from './strings';

/**
 * Hòm thư góp ý phía người chơi — ADMIN_SPEC §7.1.
 *
 * Nút 💬 ở trang chủ và TRONG TRẬN. Trong trận nó không che HUD và **không tạm dừng game**:
 * server vẫn chạy, nhân vật vẫn đứng đó — nên form phải mở nhanh, đóng nhanh, và không
 * nuốt phím điều khiển.
 *
 * Tự đính kèm ngữ cảnh (§7.1) và **người chơi được THẤY mình gửi những gì** — một dòng liệt
 * kê rõ ràng. Gửi lén thông tin máy người ta là chuyện khác hẳn với gửi kèm có báo trước.
 */

/** ERR_KEEP: giữ 20 lỗi console gần nhất (§7.1, chỉ gửi khi chọn "Báo lỗi"). */
const ERR_KEEP = 20;

/** recent: vòng đệm lỗi. Ghi đè từ đầu khi đầy — lỗi mới quan trọng hơn lỗi cũ. */
const recent: string[] = [];

function push(line: string): void {
  recent.push(line.slice(0, 300));
  if (recent.length > ERR_KEEP) {
    recent.shift();
  }
}

/**
 * installErrorCapture bắt lỗi console để đính kèm vào báo lỗi.
 *
 * Bắt cả ba đường: `console.error` (lỗi do chính mình ghi), `window.onerror` (ngoại lệ
 * không bắt), `unhandledrejection` (Promise hỏng). Thiếu đường thứ ba là mất gần hết lỗi
 * của mã bất đồng bộ — mà phần lớn mã tải art và mạng đều là bất đồng bộ.
 */
export function installErrorCapture(): void {
  const orig = console.error.bind(console);
  console.error = (...args: unknown[]) => {
    push(args.map(String).join(' '));
    orig(...args);
  };
  window.addEventListener('error', (e) => {
    push(`${e.message} @ ${e.filename}:${e.lineno}`);
  });
  window.addEventListener('unhandledrejection', (e) => {
    push('promise: ' + String((e as PromiseRejectionEvent).reason));
  });
  // Lỗi xảy ra TRƯỚC khi module này nạp xong được `index.html` giữ hộ.
  const early = (window as { __kxEarlyErrors?: string[] }).__kxEarlyErrors;
  if (Array.isArray(early)) {
    for (const e of early) {
      push(String(e));
    }
  }
}

/** FeedbackContext: ngữ cảnh do người gọi cung cấp (khu, cảnh giới, ping, FPS…). */
export interface FeedbackContext {
  [k: string]: unknown;
}

type ContextFn = () => FeedbackContext;
type ShotFn = () => string | null;

let getContext: ContextFn = () => ({});
let getShot: ShotFn = () => null;

/** configure nối form với phần còn lại của game (ngữ cảnh + cách chụp màn hình). */
export function configure(ctx: ContextFn, shot: ShotFn): void {
  getContext = ctx;
  getShot = shot;
}

let host: HTMLElement | null = null;

/** mount gắn nút 💬 vào `parent`. Gọi một lần cho trang chủ, một lần cho HUD. */
export function mountButton(parent: HTMLElement, where: 'home' | 'match'): HTMLButtonElement {
  const b = document.createElement('button');
  b.className = 'fb-btn fb-' + where;
  b.type = 'button';
  b.textContent = '💬';
  b.title = STR.feedbackTitle;
  b.setAttribute('aria-label', STR.feedbackTitle);
  b.addEventListener('click', (e) => {
    e.preventDefault();
    e.stopPropagation();
    open();
  });
  parent.appendChild(b);
  return b;
}

function close(): void {
  host?.remove();
  host = null;
}

/** open mở form góp ý. */
export function open(): void {
  if (host) {
    return;
  }
  host = document.createElement('div');
  host.className = 'fb-wrap';
  // Bấm ra ngoài thì đóng — nhưng CHỈ khi bấm đúng lớp nền, không phải bấm trong form.
  host.addEventListener('pointerdown', (e) => {
    if (e.target === host) {
      close();
    }
  });
  // Chặn phím lọt xuống game: đang gõ góp ý mà nhân vật chạy là chuyện vô lý.
  for (const ev of ['keydown', 'keyup', 'keypress'] as const) {
    host.addEventListener(ev, (e) => e.stopPropagation());
  }

  const box = document.createElement('form');
  box.className = 'fb-box';
  box.appendChild(h('h3', STR.feedbackTitle));

  const kind = document.createElement('select');
  kind.className = 'fb-kind';
  for (const [value, label] of Object.entries(STR.feedbackKinds)) {
    const o = document.createElement('option');
    o.value = value;
    o.textContent = label;
    kind.appendChild(o);
  }
  box.appendChild(kind);

  const body = document.createElement('textarea');
  body.className = 'fb-body';
  body.rows = 5;
  body.maxLength = 1000;
  body.placeholder = STR.feedbackPlaceholder;
  box.appendChild(body);

  const shotRow = document.createElement('label');
  shotRow.className = 'fb-row';
  const shot = document.createElement('input');
  shot.type = 'checkbox';
  shot.checked = true; // mặc định bật vì loại mặc định là "Báo lỗi" (§7.1)
  shotRow.append(shot, document.createTextNode(' ' + STR.feedbackAttachShot));
  box.appendChild(shotRow);

  // Dòng "sẽ gửi kèm": người chơi phải THẤY mình gửi gì đi.
  const note = h('p', STR.feedbackWillSend);
  note.className = 'fb-note';
  box.appendChild(note);
  const updateNote = () => {
    note.textContent = STR.feedbackWillSend + (kind.value === 'bug' ? STR.feedbackWillSendBug : '');
  };
  kind.addEventListener('change', () => {
    shot.checked = kind.value === 'bug';
    updateNote();
  });
  updateNote();

  const msg = h('div', '');
  msg.className = 'fb-msg';
  box.appendChild(msg);

  const row = document.createElement('div');
  row.className = 'fb-actions';
  const cancel = document.createElement('button');
  cancel.type = 'button';
  cancel.className = 'btn';
  cancel.textContent = STR.feedbackCancel;
  cancel.addEventListener('click', close);
  const send = document.createElement('button');
  send.type = 'submit';
  send.className = 'btn fb-send';
  send.textContent = STR.feedbackSend;
  row.append(cancel, send);
  box.appendChild(row);

  box.addEventListener('submit', async (e) => {
    e.preventDefault();
    const text = body.value.trim();
    if (text.length < 5) {
      msg.textContent = STR.feedbackTooShort;
      return;
    }
    send.disabled = true;
    msg.textContent = STR.feedbackSending;
    try {
      const ctx: FeedbackContext = { ...getContext(), ua: navigator.userAgent,
        screen: `${window.innerWidth}x${window.innerHeight}`, dpr: window.devicePixelRatio };
      if (kind.value === 'bug') {
        ctx.console = recent.slice(-ERR_KEEP);
      }
      const r = await fetch('/api/feedback', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          guest_id: ctx.guest_id ?? '', name: ctx.name ?? '',
          kind: kind.value, body: text, context: ctx,
          image: shot.checked ? (getShot() ?? '') : '',
        }),
      });
      if (r.ok) {
        msg.textContent = STR.feedbackThanks;
        window.setTimeout(close, 1400);
        return;
      }
      const j = (await r.json().catch(() => ({}))) as { error?: string };
      msg.textContent = j.error || STR.feedbackFailed;
    } catch (_) {
      msg.textContent = STR.feedbackFailed;
    }
    send.disabled = false;
  });

  host.appendChild(box);
  document.body.appendChild(host);
  body.focus();
}

function h(tag: string, text: string): HTMLElement {
  const n = document.createElement(tag);
  n.textContent = text;
  return n;
}
