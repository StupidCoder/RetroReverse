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
    this.touch={x:0,y:0,down:false};this.touchEvents=[];this.touchUntil=0;
  }
  enqueue(m, time) {
    if(m.touch){
      const t=m.touch;if(!Number.isFinite(t.x)||!Number.isFinite(t.y))throw Error('Invalid touch coordinate');
      const at=t.down?time:Math.max(time,this.touchUntil);
      if(t.down)this.touchUntil=time+this.hold;
      if(this.touchEvents.length>=4096)throw Error('Touch queue overflow');
      this.touchEvents.push({at,x:Math.trunc(t.x),y:Math.trunc(t.y),down:!!t.down});
    }
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
    while(this.touchEvents.length&&this.touchEvents[0].at<=time)this.touch=this.touchEvents.shift();
    return {
      touch:this.touch,
      buttons : buttons >>> 0,
      x : this.x,
      y : this.y,
      keys,
      sequence : this.sequence
    };
  }
}
