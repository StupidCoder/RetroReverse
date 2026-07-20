// The shared state every panel reads, and the one place it changes.
//
// Panels do not talk to each other. They put what they know into the store and
// subscribe to what they need, so adding a panel cannot break an existing one — which
// matters when the panel set is decided at runtime by what the target can do.

export class Store {
  constructor() {
    this.state = {
      platform: '',
      keyLegend: '',
      aspectNum: 0, // display aspect num:den for non-square pixels (0 = square)
      aspectDen: 0,
      toggles: [], // runtime on/off switches the target declares (debug.Toggler)
      title: '',
      caps: new Set(),

      frame: null, // the last frame capture (command list, size)
      prov: null, // Int32Array: which command drew each pixel
      selected: -1, // the current command
      pick: null, // the picked pixel {x, y}
      playing: false,
      cpuRunning: false, // the CPU is free-running (Continue) on a CPU-only target, streaming its scanout
      pc: null, // the CPU's program counter, as a hex string
      status: '',
    };
    this.subs = new Map(); // key -> [fn]
  }

  // can reports whether the current target backs a capability.
  can(cap) {
    return this.state.caps.has(cap);
  }

  get(key) {
    return this.state[key];
  }

  // set updates keys and notifies whoever subscribed to each one that changed.
  set(patch) {
    const changed = [];
    for (const [k, v] of Object.entries(patch)) {
      if (this.state[k] !== v) {
        this.state[k] = v;
        changed.push(k);
      }
    }
    for (const k of changed) {
      for (const fn of this.subs.get(k) || []) fn(this.state[k], this.state);
    }
  }

  // touch notifies subscribers of a key whose value was mutated in place.
  touch(key) {
    for (const fn of this.subs.get(key) || []) fn(this.state[key], this.state);
  }

  // on subscribes to a key. Scoped the same way as Conn.on, so a panel's subscriptions
  // die with the panel when the target changes and the dock is rebuilt.
  on(key, fn) {
    if (!this.subs.has(key)) this.subs.set(key, []);
    this.subs.get(key).push(fn);
    if (this.scope) this.scope.push({ key, fn });
  }

  beginScope() {
    this.scope = [];
  }

  endScope() {
    const taken = this.scope || [];
    this.scope = null;
    return () => {
      for (const { key, fn } of taken) {
        const arr = this.subs.get(key) || [];
        const i = arr.indexOf(fn);
        if (i >= 0) arr.splice(i, 1);
      }
    };
  }
}
