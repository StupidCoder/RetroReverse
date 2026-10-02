// Input times are emulated seconds. Preserve brief taps even when a slow worker
// receives both edges between slices. Physical/keyboard/touch states are
// already combined by the UI, so releasing one source cannot release another
// source.
export class InputQueue {
  constructor(hz) {
    this.mousePending={x:0,y:0};this.mouseFrame=-1;
    this.mouseMask=0;this.mouseButtons=0;this.mouseButtonEvents=[];this.mouseButtonCursor=0;
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
  mouseForFrame(frame) {
    if(frame===this.mouseFrame)return {x:0,y:0};
    this.mouseFrame=frame;const result={};
    for(const axis of ['x','y']){result[axis]=Math.max(-127,Math.min(127,this.mousePending[axis]));this.mousePending[axis]-=result[axis];}
    return result;
  }
  enqueue(m, time) {
    if(m.mouse)for(const axis of ['x','y']){
      if(!Number.isFinite(m.mouse[axis]))throw Error('Invalid mouse movement');
      this.mousePending[axis]=Math.max(-4096,Math.min(4096,this.mousePending[axis]+Math.trunc(m.mouse[axis])));
    }
    if(m.touch){
      const t=m.touch;if(!Number.isFinite(t.x)||!Number.isFinite(t.y))throw Error('Invalid touch coordinate');
      const at=t.down?time:Math.max(time,this.touchUntil);
      if(t.down)this.touchUntil=time+this.hold;
      if(this.touchEvents.length>=4096)throw Error('Touch queue overflow');
      this.touchEvents.push({at,x:Math.trunc(t.x),y:Math.trunc(t.y),down:!!t.down});
    }
    const next = m.buttons >>> 0;
    if((next^this.buttons)&this.mouseMask){
      const at=Math.max(time,this.mouseButtonCursor);
      if(this.mouseButtonEvents.length>=4096)throw Error('Mouse button queue overflow');
      this.mouseButtonEvents.push({at,buttons:next&this.mouseMask});
      this.mouseButtonCursor=at+this.hold;
    }
    for (let i = 0; i < 32; i++) {
      const bit = (2 ** i) >>> 0;
      if (!(bit&this.mouseMask) && (next & bit) && !(this.buttons & bit))
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
    while(this.mouseButtonEvents.length&&this.mouseButtonEvents[0].at<=time)this.mouseButtons=this.mouseButtonEvents.shift().buttons;
    let buttons = (this.buttons&~this.mouseMask)|this.mouseButtons;
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
