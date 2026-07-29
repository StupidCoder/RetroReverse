// pancam.js — inertial scrolling for the 2-D views.
//
// The arrow keys used to jump a fixed number of pixels per key-repeat event,
// so a level scrolled in hops the size of a tile and at the mercy of the OS
// key-repeat rate. Here the keys drive a VELOCITY instead: a short spin-up
// while held, a longer glide after release, integrated once per rendered
// frame. A drag release feeds its own velocity into the same integrator, so
// throwing the map coasts to a stop as well.
//
// The host owns the camera: PanInput only produces screen-space deltas and
// calls back. Shared by the tilemap viewer (level2d) and the flat 3-D stages
// (level3d's pan2d mode) so both scroll with the same feel.

const DIRS = {
  ArrowLeft: [1, 0], ArrowRight: [-1, 0], ArrowUp: [0, 1], ArrowDown: [0, -1],
};
const ZOOMS = { Equal: 1, NumpadAdd: 1, Minus: -1, NumpadSubtract: -1 };

const SPEED = 1300;      // screen px/s at full tilt
const BOOST = 2.6;       // shift multiplier
const TAU_ON = 0.07;     // spin-up time constant (held)
const TAU_OFF = 0.26;    // glide time constant (released) — "a little inertia"
const STOP = 3;          // px/s below which the glide is dropped
const FLING_MAX = 7000;  // px/s cap on a thrown drag
const FLING_WINDOW = 90; // ms of pointer history a fling averages over
const FLING_STALE = 130; // ms since the last move beyond which a release is not a throw
const ZOOM_RATE = 1.7;   // e-folds/s while +/- is held
const ZTAU_ON = 0.08, ZTAU_OFF = 0.12;
const ZSTOP = 0.02;

// blend moves v towards target with a time constant, framerate-independently.
const blend = (v, target, tau, dt) => v + (target - v) * (1 - Math.exp(-dt / tau));

export class PanInput {
  // onPan(dx, dy): scroll the content by these screen pixels. May return the
  //   delta actually applied ({x, y}) so a camera that hit its bounds kills
  //   the glide instead of pressing against the edge.
  // onZoom(factor): zoom about the view centre.
  constructor({ onPan, onZoom, speed = SPEED }) {
    this.onPan = onPan;
    this.onZoom = onZoom;
    this.speed = speed;
    this.keys = new Set();
    this.shift = false;
    this.vx = 0; this.vy = 0; this.vz = 0;
    this._trail = [];   // recent drag samples, for the fling
    this._down = (e) => {
      this.shift = e.shiftKey;
      if (e.altKey || e.ctrlKey || e.metaKey) return;
      // never eat the arrows while the user is typing in the Studio's chrome
      const t = e.target;
      if (t instanceof HTMLElement && (t.isContentEditable || /^(INPUT|TEXTAREA|SELECT)$/.test(t.tagName))) return;
      if (!DIRS[e.code] && !ZOOMS[e.code]) return;
      this.keys.add(e.code);
      e.preventDefault();
    };
    this._up = (e) => { this.shift = e.shiftKey; this.keys.delete(e.code); };
    this._blur = () => this.keys.clear();
    window.addEventListener('keydown', this._down);
    window.addEventListener('keyup', this._up);
    window.addEventListener('blur', this._blur);
  }

  // step integrates one frame; the host calls it from its own ticker/RAF.
  step(dt) {
    if (!(dt > 0)) return;
    dt = Math.min(dt, 0.1); // a backgrounded tab must not launch the camera
    let ax = 0, ay = 0, az = 0;
    for (const code of this.keys) {
      const d = DIRS[code];
      if (d) { ax += d[0]; ay += d[1]; }
      const z = ZOOMS[code];
      if (z) az += z;
    }
    ax = Math.max(-1, Math.min(1, ax));
    ay = Math.max(-1, Math.min(1, ay));
    const norm = ax && ay ? Math.SQRT1_2 : 1;
    const v = this.speed * (this.shift ? BOOST : 1) * norm;
    this.vx = blend(this.vx, ax * v, ax ? TAU_ON : TAU_OFF, dt);
    this.vy = blend(this.vy, ay * v, ay ? TAU_ON : TAU_OFF, dt);
    if (Math.abs(this.vx) < STOP && !ax) this.vx = 0;
    if (Math.abs(this.vy) < STOP && !ay) this.vy = 0;
    if (this.vx || this.vy) {
      const dx = this.vx * dt, dy = this.vy * dt;
      const got = this.onPan(dx, dy);
      // Bounds: an axis that refused most of its step has hit the edge of the
      // map — drop its velocity so the release doesn't keep pushing.
      if (got) {
        if (Math.abs(got.x) < Math.abs(dx) * 0.5) this.vx = 0;
        if (Math.abs(got.y) < Math.abs(dy) * 0.5) this.vy = 0;
      }
    }
    if (this.onZoom) {
      this.vz = blend(this.vz, az * ZOOM_RATE, az ? ZTAU_ON : ZTAU_OFF, dt);
      if (Math.abs(this.vz) < ZSTOP && !az) this.vz = 0;
      if (this.vz) this.onZoom(Math.exp(this.vz * dt));
    }
  }

  stop() { this.vx = this.vy = this.vz = 0; this._trail.length = 0; }

  // ---- drag throw ---------------------------------------------------------
  // The host reports each drag delta; the release turns the last few
  // milliseconds of it into a glide.

  dragMove(dx, dy, now = performance.now()) {
    this._trail.push({ t: now, dx, dy });
    while (this._trail.length && now - this._trail[0].t > FLING_WINDOW) this._trail.shift();
  }

  dragEnd(now = performance.now()) {
    const tr = this._trail;
    this._trail = [];
    if (tr.length < 2 || now - tr[tr.length - 1].t > FLING_STALE) return;
    const span = (now - tr[0].t) / 1000;
    if (span <= 0) return;
    let dx = 0, dy = 0;
    for (const s of tr) { dx += s.dx; dy += s.dy; }
    this.fling(dx / span, dy / span);
  }

  fling(vx, vy) {
    const clamp = (v) => Math.max(-FLING_MAX, Math.min(FLING_MAX, v));
    this.vx = clamp(vx);
    this.vy = clamp(vy);
    if (Math.abs(this.vx) < STOP * 8) this.vx = 0; // a slow drag is a placement, not a throw
    if (Math.abs(this.vy) < STOP * 8) this.vy = 0;
  }

  dispose() {
    window.removeEventListener('keydown', this._down);
    window.removeEventListener('keyup', this._up);
    window.removeEventListener('blur', this._blur);
    this.stop();
  }
}
