// The socket to the debugger. Text messages carry structure; binary messages carry
// pixels, framed by the 16-byte header protocol.go documents. The payload lands
// 4-byte aligned, so the typed-array views below are windows onto the received
// buffer, not copies of it.

export const KIND_IMAGE = 1;
export const KIND_PROV = 2;

export class Conn {
  constructor() {
    this.seq = 0;
    this.onJSON = () => {};
    this.onBinary = () => {};
    this.onOpen = () => {};
    this.onClose = () => {};

    const url = `ws://${location.host}/ws`;
    this.ws = new WebSocket(url);
    this.ws.binaryType = 'arraybuffer';
    this.ws.onopen = () => this.onOpen();
    this.ws.onclose = () => this.onClose();
    this.ws.onmessage = (e) => {
      if (typeof e.data === 'string') this.onJSON(JSON.parse(e.data));
      else this.onBinary(decode(e.data));
    };
  }

  // send queues an op and returns the sequence number the reply will echo, so a
  // caller can drop replies to requests it has already superseded.
  send(op, params = {}) {
    const seq = (this.seq = (this.seq + 1) & 0xffff);
    this.ws.send(JSON.stringify({ op, seq, ...params }));
    return seq;
  }
}

function decode(buf) {
  const h = new DataView(buf, 0, 16);
  const magic = String.fromCharCode(h.getUint8(0), h.getUint8(1), h.getUint8(2), h.getUint8(3));
  if (magic !== 'RDB1') throw new Error(`bad binary message magic ${magic}`);
  const kind = h.getUint8(4);
  const seq = h.getUint16(6, true);
  const w = h.getUint32(8, true);
  const h_ = h.getUint32(12, true);

  if (kind === KIND_IMAGE) {
    const pix = new Uint8ClampedArray(buf, 16, w * h_ * 4);
    return { kind, seq, w, h: h_, image: new ImageData(pix, w, h_) };
  }
  if (kind === KIND_PROV) {
    return { kind, seq, w, h: h_, prov: new Int32Array(buf, 16, w * h_) };
  }
  throw new Error(`unknown binary kind ${kind}`);
}
