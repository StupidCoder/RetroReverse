// The shared state every panel reads, and the one place it changes.
//
// Panels do not talk to each other. They put what they know into the store and
// subscribe to what they need, so adding a panel cannot break an existing one — which
// matters when the panel set is decided at runtime by what the target can do.

export class Store {
  constructor() {
    this.state = {
      platform: '',
      title: '',
      caps: new Set(),

      frame: null, // the last frame capture (command list, size)
      prov: null, // Int32Array: which command drew each pixel
      selected: -1, // the current command
      pick: null, // the picked pixel {x, y}
      playing: false,
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

  on(key, fn) {
    if (!this.subs.has(key)) this.subs.set(key, []);
    this.subs.get(key).push(fn);
  }
}
