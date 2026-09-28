// Input times are emulated seconds. Preserve brief taps even when a slow worker
// receives both edges between slices. Physical/keyboard/touch states are
// already combined by the UI, so releasing one source cannot release another
// source.
export class InputQueue {
  constructor(hz) {
    this.hold = 1 / hz;
    this.buttons = 0;
    this.x = this.y = 0;
    this.pulses = new Map();
    this.keys = [];
    this.down = new Map();
    this.cursor = 0;
    this.sequence = 0;
  }
  enqueue(m, time) {
    const next = m.buttons >>> 0;
    for (let i = 0; i < 32; i++) {
      const bit = (2 ** i) >>> 0;
      if ((next & bit) && !(this.buttons & bit))
        this.pulses.set(bit, time + this.hold);
    }
    this.buttons = next;
    this.x = m.x || 0;
    this.y = m.y || 0;
    this.sequence = m.request;
    for (const [code, down] of m.keys || []) {
      let at = Math.max(time, this.cursor);
      if (down)
        this.down.set(code, at);
      else {
        at = Math.max(at, (this.down.get(code) ?? time) + .06);
        this.down.delete(code);
      }
      this.keys.push({at, code, down});
      this.cursor = at + (down ? 0 : .02);
    }
  }
  drain(time) {
    let buttons = this.buttons;
    for (const [bit, until] of this.pulses) {
      if (time < until)
        buttons |= bit;
      else
        this.pulses.delete(bit);
    }
    const keys = [];
    while (this.keys.length && this.keys[0].at <= time) {
      const k = this.keys.shift();
      keys.push([ k.code, k.down ]);
    }
    return {
      buttons : buttons >>> 0,
      x : this.x,
      y : this.y,
      keys,
      sequence : this.sequence
    };
  }
}
