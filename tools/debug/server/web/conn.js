// The socket to the debugger.
//
// Text messages carry structure; binary messages carry pixels, framed by the 24-byte
// header protocol.go documents. The payload lands 4-byte aligned, so the typed-array
// views below are windows onto the received buffer, not copies of it.
//
// Replies are addressed to a request by seq. Events (a watch hit, a breakpoint stop, a
// free-running frame) carry no seq and arrive unasked — panels subscribe to those by
// message type.

export const KIND_IMAGE = 1;
export const KIND_PROV = 2;

export const STREAM_MAIN = 0;
export const STREAM_SURFACE = 1;

export class Conn {
  constructor() {
    this.seq = 0;
    this.handlers = new Map(); // type -> [fn]
    this.binaryHandlers = [];
    this.onOpen = () => {};
    this.onClose = () => {};

    this.ws = new WebSocket(`ws://${location.host}/ws`);
    this.ws.binaryType = 'arraybuffer';
    this.ws.onopen = () => this.onOpen();
    this.ws.onclose = () => this.onClose();
    this.ws.onmessage = (e) => {
      if (typeof e.data === 'string') {
        const m = JSON.parse(e.data);
        for (const fn of this.handlers.get(m.type) || []) fn(m);
      } else {
        const m = decode(e.data);
        for (const fn of this.binaryHandlers) fn(m);
      }
    };
  }

  // on subscribes to a message type. Several panels may listen to the same type.
  on(type, fn) {
    if (!this.handlers.has(type)) this.handlers.set(type, []);
    this.handlers.get(type).push(fn);
  }

  onBinary(fn) {
    this.binaryHandlers.push(fn);
  }

  // send queues an op and returns the sequence number its reply will echo, so a caller
  // can drop replies to requests it has already superseded.
  send(op, args = null) {
    const seq = (this.seq = (this.seq % 0xffff) + 1);
    this.ws.send(JSON.stringify({ op, seq, args }));
    return seq;
  }
}

function decode(buf) {
  const h = new DataView(buf, 0, 24);
  const magic = String.fromCharCode(h.getUint8(0), h.getUint8(1), h.getUint8(2), h.getUint8(3));
  if (magic !== 'RDB2') throw new Error(`bad binary message magic ${magic}`);
  const kind = h.getUint8(4);
  const seq = h.getUint16(6, true);
  const stream = h.getUint16(8, true);
  const w = h.getUint32(12, true);
  const h_ = h.getUint32(16, true);

  if (kind === KIND_IMAGE) {
    const pix = new Uint8ClampedArray(buf, 24, w * h_ * 4);
    return { kind, seq, stream, w, h: h_, image: new ImageData(pix, w, h_) };
  }
  if (kind === KIND_PROV) {
    return { kind, seq, stream, w, h: h_, prov: new Int32Array(buf, 24, w * h_) };
  }
  throw new Error(`unknown binary kind ${kind}`);
}
