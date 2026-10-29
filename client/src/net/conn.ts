import { decode, encode } from '@msgpack/msgpack';

import type { ServerMsg } from './protocol';

type Handler = (msg: ServerMsg) => void;

const PING_EVERY_MS = 2000;
const RECONNECT_MS = 1500;

// readLagMs supports ?lag=150 to simulate 150 ms round-trip ping (spec M5).
function readLagMs(): number {
  const v = Number(new URLSearchParams(location.search).get('lag') ?? 0);
  return Number.isFinite(v) && v > 0 ? v : 0;
}

/** Conn wraps the game WebSocket: MessagePack framing, optional fake lag, ping measurement. */
export class Conn {
  private ws: WebSocket | null = null;
  private readonly halfLag = readLagMs() / 2;
  private pingTimer = 0;
  private onOpenCb: (() => void) | null = null;
  private onCloseCb: (() => void) | null = null;
  private shouldReconnect = true;
  rttMs = 0;
  // Đếm băng thông cho overlay đo (spec §13.6): bao nhiêu KB/s đang nhận và gói snapshot
  // gần nhất to bao nhiêu. Không có số này thì không phân biệt được nghẽn mạng với nghẽn vẽ.
  private bytesAcc = 0;
  private bytesAt = performance.now();
  kbPerSec = 0;
  lastSnapBytes = 0;

  constructor(private readonly onMessage: Handler) {}

  connect(onOpen: () => void, onClose: () => void): void {
    this.onOpenCb = onOpen;
    this.onCloseCb = onClose;
    this.open();
  }

  private open(): void {
    const proto = location.protocol === 'https:' ? 'wss' : 'ws';
    const ws = new WebSocket(`${proto}://${location.host}/ws`);
    ws.binaryType = 'arraybuffer';
    ws.onopen = () => {
      this.pingTimer = window.setInterval(() => this.send({ t: 'ping', c: performance.now() }), PING_EVERY_MS);
      this.onOpenCb?.();
    };
    ws.onmessage = (e: MessageEvent<ArrayBuffer>) => {
      const bytes = e.data.byteLength;
      this.bytesAcc += bytes;
      const now = performance.now();
      if (now - this.bytesAt >= 1000) {
        this.kbPerSec = this.bytesAcc / 1024 / ((now - this.bytesAt) / 1000);
        this.bytesAcc = 0;
        this.bytesAt = now;
      }
      const msg = decode(new Uint8Array(e.data)) as ServerMsg;
      if (msg.t === 'snap') {
        this.lastSnapBytes = bytes;
      }
      this.delay(() => this.receive(msg));
    };
    ws.onclose = () => {
      window.clearInterval(this.pingTimer);
      this.onCloseCb?.();
      if (this.shouldReconnect) {
        window.setTimeout(() => this.open(), RECONNECT_MS);
      }
    };
    this.ws = ws;
  }

  private receive(msg: ServerMsg): void {
    if (msg.t === 'pong') {
      this.rttMs = Math.round(performance.now() - msg.c);
    }
    this.onMessage(msg);
  }

  private delay(fn: () => void): void {
    if (this.halfLag > 0) {
      window.setTimeout(fn, this.halfLag);
    } else {
      fn();
    }
  }

  /** stop đóng kết nối và KHÔNG nối lại — dùng khi nối lại cũng vô ích (lệch giao thức). */
  stop(): void {
    this.shouldReconnect = false;
    window.clearInterval(this.pingTimer);
    this.ws?.close();
  }

  get isOpen(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  send(msg: Record<string, unknown>): void {
    const data = encode(msg);
    this.delay(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(data);
      }
    });
  }
}
