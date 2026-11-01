import { blobToDataURL, compressImage, dataURLToBlob, ImageTooBigError } from './imagecompress';
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

/** MAX_IMAGES: tối đa 3 ảnh / góp ý, TÍNH CẢ ảnh chụp màn hình (§7.1). */
const MAX_IMAGES = 3;

interface Attachment {
  blob: Blob;
  url: string;
  isShot: boolean;
}

type Say = (text: string) => void;

/**
 * Attachments giữ danh sách ảnh đính kèm của form đang mở và vẽ ô xem trước.
 *
 * `pending` đếm ảnh đang nén dở: giới hạn 3 ảnh phải tính cả chúng, không thì người chơi
 * thả 5 ảnh một lúc là lọt cả 5 trước khi ảnh đầu kịp nén xong.
 */
class Attachments {
  readonly list: Attachment[] = [];
  pending = 0;
  onChange: () => void = () => {};

  constructor(private readonly row: HTMLElement) {}

  get room(): number {
    return MAX_IMAGES - this.list.length - this.pending;
  }

  hasShot(): boolean {
    return this.list.some((a) => a.isShot);
  }

  add(blob: Blob, isShot: boolean): void {
    const a = { blob, url: URL.createObjectURL(blob), isShot };
    // Ảnh màn hình luôn đứng đầu: nó là "cái người chơi đang thấy", ảnh tải lên là phụ.
    if (isShot) {
      this.list.unshift(a);
    } else {
      this.list.push(a);
    }
    this.render();
  }

  remove(a: Attachment): void {
    const i = this.list.indexOf(a);
    if (i >= 0) {
      this.list.splice(i, 1);
      URL.revokeObjectURL(a.url);
    }
    this.render();
  }

  removeShot(): void {
    const a = this.list.find((x) => x.isShot);
    if (a) {
      this.remove(a);
    }
  }

  dispose(): void {
    for (const a of this.list) {
      URL.revokeObjectURL(a.url);
    }
    this.list.length = 0;
  }

  render(): void {
    this.row.replaceChildren(...this.list.map((a) => this.thumb(a)));
    this.row.hidden = this.list.length === 0;
    this.onChange();
  }

  private thumb(a: Attachment): HTMLElement {
    const cell = document.createElement('div');
    cell.className = 'fb-thumb';
    const img = document.createElement('img');
    img.src = a.url;
    img.alt = '';
    cell.appendChild(img);
    if (a.isShot) {
      const tag = h('span', STR.feedbackShotBadge);
      tag.className = 'fb-thumb-tag';
      cell.appendChild(tag);
    }
    const x = document.createElement('button');
    x.type = 'button';
    x.className = 'fb-thumb-x';
    x.textContent = '×';
    x.title = STR.feedbackImageRemove;
    x.setAttribute('aria-label', STR.feedbackImageRemove);
    x.addEventListener('click', () => this.remove(a));
    cell.appendChild(x);
    return cell;
  }
}

let current: Attachments | null = null;

function close(): void {
  current?.dispose();
  current = null;
  host?.remove();
  host = null;
}

/** compressMessage dịch lỗi nén ảnh ra câu cho người chơi. */
function compressMessage(err: unknown): string {
  if (err instanceof ImageTooBigError) {
    return STR.feedbackImageTooBig;
  }
  return STR.feedbackImageUnsupported;
}

/** addFiles nén rồi đính kèm các tệp người chơi chọn / thả / dán. */
async function addFiles(att: Attachments, files: File[], say: Say): Promise<void> {
  if (files.length > att.room) {
    say(STR.feedbackImageMax);
  }
  const take = files.slice(0, Math.max(0, att.room));
  att.pending += take.length;
  att.onChange();
  await Promise.all(take.map(async (f) => {
    try {
      const blob = await compressImage(f);
      att.pending--;
      att.add(blob, false);
    } catch (err) {
      att.pending--;
      att.onChange();
      say(compressMessage(err));
    }
  }));
}

/** attachShot chụp màn hình (data URL do game đưa) rồi nén lại như mọi ảnh khác. */
async function attachShot(att: Attachments, say: Say): Promise<boolean> {
  if (att.room <= 0) {
    say(STR.feedbackImageMax);
    return false;
  }
  const url = getShot();
  if (!url) {
    say(STR.feedbackShotFailed);
    return false;
  }
  att.pending++;
  att.onChange();
  try {
    // Nén lại vì Safari chụp canvas ra PNG chứ không ra WebP.
    const blob = await compressImage(await dataURLToBlob(url));
    att.pending--;
    att.add(blob, true);
    return true;
  } catch (_) {
    att.pending--;
    att.onChange();
    say(STR.feedbackShotFailed);
    return false;
  }
}

function asFile(it: File | DataTransferItem): File | null {
  if (it instanceof File) {
    return it;
  }
  return it.kind === 'file' ? it.getAsFile() : null;
}

function imageFiles(list: FileList | DataTransferItemList | null | undefined): File[] {
  const out: File[] = [];
  const items: ArrayLike<File | DataTransferItem> = list ?? [];
  for (const it of Array.from(items)) {
    const f = asFile(it);
    // `type` rỗng vẫn nhận để THỬ giải mã: một số máy Android báo ảnh HEIC không có MIME.
    if (f && (f.type === '' || f.type.startsWith('image/'))) {
      out.push(f);
    }
  }
  return out;
}

/** buildImagePicker dựng nút "Thêm ảnh" + hàng ô xem trước, gắn kéo-thả và dán vào form. */
function buildImagePicker(box: HTMLElement, say: Say): { el: HTMLElement; att: Attachments } {
  const thumbs = document.createElement('div');
  thumbs.className = 'fb-thumbs';
  const att = new Attachments(thumbs);

  // `accept="image/*"` + `multiple`: trên iPhone/Android mở thư viện ảnh hoặc máy ảnh.
  const input = document.createElement('input');
  input.type = 'file';
  input.accept = 'image/*';
  input.multiple = true;
  input.className = 'fb-file';
  input.hidden = true;
  input.addEventListener('change', () => {
    void addFiles(att, imageFiles(input.files), say);
    input.value = ''; // chọn lại đúng tệp vừa xoá vẫn phải bắn `change`
  });
  const add = document.createElement('button');
  add.type = 'button';
  add.className = 'btn fb-add';
  add.textContent = STR.feedbackAddImage;
  add.addEventListener('click', () => input.click());
  const hint = h('span', STR.feedbackImageHint);
  hint.className = 'fb-hint';
  const bar = document.createElement('div');
  bar.className = 'fb-addrow';
  bar.append(add, hint);

  const wrap = document.createElement('div');
  wrap.className = 'fb-images';
  wrap.append(bar, thumbs, input);
  thumbs.hidden = true;
  bindDropAndPaste(box, att, say);
  return { el: wrap, att };
}

function bindDropAndPaste(box: HTMLElement, att: Attachments, say: Say): void {
  box.addEventListener('dragover', (e) => {
    e.preventDefault();
    box.classList.add('fb-drop');
  });
  box.addEventListener('dragleave', (e) => {
    if (e.target === box) {
      box.classList.remove('fb-drop');
    }
  });
  box.addEventListener('drop', (e) => {
    e.preventDefault();
    box.classList.remove('fb-drop');
    void addFiles(att, imageFiles(e.dataTransfer?.files), say);
  });
  // Dán: chỉ chặn hành vi mặc định khi có ẢNH — dán chữ vào ô nội dung vẫn phải chạy.
  box.addEventListener('paste', (e) => {
    const files = imageFiles(e.clipboardData?.items);
    if (files.length > 0) {
      e.preventDefault();
      void addFiles(att, files, say);
    }
  });
}

/** sendFeedback gửi lá thư; trả về true nếu server nhận. Lỗi thì tự báo qua `say`. */
async function sendFeedback(kind: string, text: string, att: Attachments, say: Say): Promise<boolean> {
  try {
    const ctx: FeedbackContext = { ...getContext(), ua: navigator.userAgent,
      screen: `${window.innerWidth}x${window.innerHeight}`, dpr: window.devicePixelRatio };
    if (kind === 'bug') {
      ctx.console = recent.slice(-ERR_KEEP);
    }
    const images = await Promise.all(att.list.map((x) => blobToDataURL(x.blob)));
    const r = await fetch('/api/feedback', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        guest_id: ctx.guest_id ?? '', name: ctx.name ?? '',
        kind, body: text, context: ctx, images,
      }),
    });
    if (r.ok) {
      return true;
    }
    const j = (await r.json().catch(() => ({}))) as { error?: string };
    say(j.error || STR.feedbackFailed);
  } catch (_) {
    say(STR.feedbackFailed);
  }
  return false;
}

function buildKindSelect(): HTMLSelectElement {
  const kind = document.createElement('select');
  kind.className = 'fb-kind';
  for (const [value, label] of Object.entries(STR.feedbackKinds)) {
    const o = document.createElement('option');
    o.value = value;
    o.textContent = label;
    kind.appendChild(o);
  }
  return kind;
}

function buildBody(): HTMLTextAreaElement {
  const body = document.createElement('textarea');
  body.className = 'fb-body';
  body.rows = 5;
  body.maxLength = 1000;
  body.placeholder = STR.feedbackPlaceholder;
  return body;
}

function buildActions(): { row: HTMLElement; send: HTMLButtonElement } {
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
  return { row, send };
}

function buildHost(): HTMLElement {
  const wrap = document.createElement('div');
  wrap.className = 'fb-wrap';
  // Bấm ra ngoài thì đóng — nhưng CHỈ khi bấm đúng lớp nền, không phải bấm trong form.
  wrap.addEventListener('pointerdown', (e) => {
    if (e.target === wrap) {
      close();
    }
  });
  // Chặn phím lọt xuống game: đang gõ góp ý mà nhân vật chạy là chuyện vô lý.
  for (const ev of ['keydown', 'keyup', 'keypress'] as const) {
    wrap.addEventListener(ev, (e) => e.stopPropagation());
  }
  return wrap;
}

/** open mở form góp ý. */
export function open(): void {
  if (host) {
    return;
  }
  host = buildHost();
  const box = document.createElement('form');
  box.className = 'fb-box';
  const kind = buildKindSelect();
  const body = buildBody();
  const msg = h('div', '');
  msg.className = 'fb-msg';
  const say: Say = (t) => {
    msg.textContent = t;
  };

  const shotRow = document.createElement('label');
  shotRow.className = 'fb-row';
  const shot = document.createElement('input');
  shot.type = 'checkbox';
  shot.className = 'fb-shot';
  shotRow.append(shot, document.createTextNode(' ' + STR.feedbackAttachShot));

  const picker = buildImagePicker(box, say);
  const att = picker.att;
  current = att;

  // Dòng "sẽ gửi kèm": người chơi phải THẤY mình gửi gì đi.
  const note = h('p', STR.feedbackWillSend);
  note.className = 'fb-note';
  const { row, send } = buildActions();
  box.append(h('h3', STR.feedbackTitle), kind, body, shotRow, picker.el, note, msg, row);

  wireForm({ box, kind, body, shot, note, send, att, say });
  host.appendChild(box);
  document.body.appendChild(host);
  body.focus();
}

interface FormParts {
  box: HTMLFormElement;
  kind: HTMLSelectElement;
  body: HTMLTextAreaElement;
  shot: HTMLInputElement;
  note: HTMLElement;
  send: HTMLButtonElement;
  att: Attachments;
  say: Say;
}

/** wireForm nối các phần của form với nhau: ô tick ảnh màn hình, loại góp ý, nút gửi. */
function wireForm(p: FormParts): void {
  const { box, kind, body, shot, note, send, att, say } = p;
  // Ô tick và ô xem trước "Màn hình" là MỘT trạng thái: bỏ tick = xoá ô, xoá ô = bỏ tick.
  const syncShot = async (want: boolean, quiet = false): Promise<void> => {
    if (want && !att.hasShot()) {
      shot.checked = await attachShot(att, quiet ? () => {} : say);
    } else if (!want) {
      att.removeShot();
    }
  };
  shot.addEventListener('change', () => void syncShot(shot.checked));
  const updateNote = (): void => {
    note.textContent = STR.feedbackWillSend + (kind.value === 'bug' ? STR.feedbackWillSendBug : '');
  };
  kind.addEventListener('change', () => {
    shot.checked = kind.value === 'bug';
    void syncShot(shot.checked);
    updateNote();
  });
  updateNote();

  let busyShown = false;
  att.onChange = () => {
    if (att.pending === 0) {
      shot.checked = att.hasShot();
    }
    send.disabled = att.pending > 0;
    if (att.pending > 0 && !busyShown) {
      say(STR.feedbackImageBusy);
      busyShown = true;
    } else if (att.pending === 0 && busyShown) {
      busyShown = false;
      if (box.querySelector('.fb-msg')?.textContent === STR.feedbackImageBusy) {
        say('');
      }
    }
  };

  box.addEventListener('submit', async (e) => {
    e.preventDefault();
    const text = body.value.trim();
    if (text.length < 5) {
      say(STR.feedbackTooShort);
      return;
    }
    send.disabled = true;
    say(STR.feedbackSending);
    if (await sendFeedback(kind.value, text, att, say)) {
      say(STR.feedbackThanks);
      window.setTimeout(close, 1400);
      return;
    }
    send.disabled = false;
  });

  // Mặc định là "Báo lỗi" nên mặc định kèm ảnh chụp màn hình (§7.1). Chụp NGAY lúc mở:
  // khung hình người chơi đang thấy khi gặp lỗi là thứ đáng gửi nhất.
  shot.checked = true;
  void syncShot(true, true); // tự kèm: không chụp được (trang chủ chưa có canvas) thì im lặng bỏ tick
}

function h(tag: string, text: string): HTMLElement {
  const n = document.createElement(tag);
  n.textContent = text;
  return n;
}
