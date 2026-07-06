// FlyCam — free-flight camera controls for level/course models, layered on top of
// an OrbitControls instance (mouse-drag orbit/zoom keep working; this adds motion).
//
//   desktop: WASD moves (forward/strafe relative to the view), arrow keys look
//   touch:   two virtual analog sticks in the lower corners — left moves, right looks
//
// It works by nudging the camera and the OrbitControls target together: WASD/left
// stick translates both (so the orbit pivot travels with you), arrows/right stick
// rotates the target around the camera (a first-person look). Small objects and
// characters keep the plain auto-rotating orbit — viewers enable this only for
// levels, via setEnabled().
import * as THREE from 'three';

const COARSE = matchMedia('(pointer: coarse)').matches;
const UP = new THREE.Vector3(0, 1, 0);
const MOVE = 0.22;  // fraction of the model size covered per second
const ROT = 1.0;    // look speed, radians per second (~57°/s)
const PITCH_MARGIN = 0.12; // keep the view this far (radians) from straight up/down

// A short "how to fly" hint for the HUD, matching the input device.
export const flyHint = COARSE ? 'sticks: left to move, right to look' : 'WASD move · arrow keys look';

export class FlyCam {
  constructor(camera, controls, el) {
    this.camera = camera;
    this.controls = controls;
    this.el = el;
    this.enabled = false;
    this.scale = 10;      // world size of the current model, set by the viewer
    this.moveScale = 1;   // per-viewer movement-speed multiplier (look speed unaffected)
    this.keys = new Set();

    this.shift = false; // Shift doubles movement speed (not look speed)
    this._onKeyDown = (e) => {
      this.shift = e.shiftKey;
      if (!this.enabled || e.altKey || e.ctrlKey || e.metaKey) return;
      if (HANDLED_KEYS.has(e.code)) {
        this.keys.add(e.code);
        e.preventDefault(); // arrows scroll the page otherwise
      }
    };
    this._onKeyUp = (e) => { this.shift = e.shiftKey; this.keys.delete(e.code); };
    this._onBlur = () => { this.keys.clear(); this.shift = false; };
    window.addEventListener('keydown', this._onKeyDown);
    window.addEventListener('keyup', this._onKeyUp);
    window.addEventListener('blur', this._onBlur);

    this.sticks = null;
    if (COARSE) this._buildSticks();

    this._off = new THREE.Vector3();
    this._fwd = new THREE.Vector3();
    this._right = new THREE.Vector3();
  }

  // Viewers flip this per model: levels fly, objects/characters keep the orbit.
  setEnabled(on) {
    this.enabled = on;
    this.keys.clear();
    if (this.sticks) {
      for (const s of [this.sticks.l, this.sticks.r]) {
        s.base.style.display = on ? 'block' : 'none';
        this._resetStick(s);
      }
    }
  }

  setScale(size) { this.scale = size || 10; }

  // Per-viewer movement-speed multiplier (1 = default). Rotation is unaffected —
  // e.g. UW's levels are huge, so its viewer sets this low to slow translation
  // while keeping the look speed the same.
  setMoveScale(m) { this.moveScale = m || 1; }

  // Advance one frame; call before controls.update(). Returns whether any fly
  // input is active (viewers can use it to cancel animations if needed).
  update(dt) {
    if (!this.enabled) return false;
    const k = this.keys;
    const l = this.sticks ? this.sticks.l : ZERO;
    const r = this.sticks ? this.sticks.r : ZERO;

    // input axes, each -1..1: move forward/strafe, look yaw/pitch
    const mf = (k.has('KeyW') ? 1 : 0) - (k.has('KeyS') ? 1 : 0) - l.y;
    const ms = (k.has('KeyD') ? 1 : 0) - (k.has('KeyA') ? 1 : 0) + l.x;
    const ry = (k.has('ArrowLeft') ? 1 : 0) - (k.has('ArrowRight') ? 1 : 0) - r.x;
    const rp = (k.has('ArrowUp') ? 1 : 0) - (k.has('ArrowDown') ? 1 : 0) - r.y;
    if (!mf && !ms && !ry && !rp) return false;

    const pos = this.camera.position;
    const target = this.controls.target;
    const off = this._off.copy(target).sub(pos);

    // look: rotate the view offset around the camera
    if (ry) off.applyAxisAngle(UP, ry * ROT * dt);
    if (rp) {
      this._right.copy(off).cross(UP).normalize(); // camera right (forward × up)
      const rotated = off.clone().applyAxisAngle(this._right, rp * ROT * dt);
      const a = rotated.angleTo(UP);
      if (a > PITCH_MARGIN && a < Math.PI - PITCH_MARGIN) off.copy(rotated);
    }
    target.copy(pos).add(off);

    // move: translate camera and orbit pivot together, view-relative
    if (mf || ms) {
      const v = this.scale * MOVE * this.moveScale * dt * (this.shift ? 2 : 1);
      this._fwd.copy(off).normalize();
      this._right.copy(this._fwd).cross(UP).normalize();
      pos.addScaledVector(this._fwd, mf * v).addScaledVector(this._right, ms * v);
      target.addScaledVector(this._fwd, mf * v).addScaledVector(this._right, ms * v);
    }
    return true;
  }

  dispose() {
    window.removeEventListener('keydown', this._onKeyDown);
    window.removeEventListener('keyup', this._onKeyUp);
    window.removeEventListener('blur', this._onBlur);
    if (this.sticks) {
      this.sticks.l.base.remove();
      this.sticks.r.base.remove();
      this.sticks = null;
    }
  }

  // ---- virtual analog sticks (touch devices) ----

  _buildSticks() {
    if (getComputedStyle(this.el).position === 'static') this.el.style.position = 'relative';
    const mk = (side) => {
      const base = document.createElement('div');
      base.style.cssText =
        `position:absolute;bottom:84px;${side}:26px;width:112px;height:112px;border-radius:50%;` + // above the corner fab + HUD
        'background:rgba(110,130,160,.16);border:1px solid rgba(150,175,210,.35);' +
        'display:none;z-index:6;touch-action:none;-webkit-tap-highlight-color:transparent;' +
        'user-select:none;-webkit-user-select:none;';
      const knob = document.createElement('div');
      knob.style.cssText =
        'position:absolute;left:50%;top:50%;width:46px;height:46px;border-radius:50%;' +
        'background:rgba(170,195,230,.45);border:1px solid rgba(200,220,245,.5);' +
        'transform:translate(-50%,-50%);pointer-events:none;';
      base.appendChild(knob);
      this.el.appendChild(base);
      const s = { base, knob, id: null, x: 0, y: 0 };
      base.addEventListener('pointerdown', (e) => {
        s.id = e.pointerId;
        base.setPointerCapture(e.pointerId);
        this._trackStick(s, e);
      });
      base.addEventListener('pointermove', (e) => { if (e.pointerId === s.id) this._trackStick(s, e); });
      const end = (e) => { if (e.pointerId === s.id) this._resetStick(s); };
      base.addEventListener('pointerup', end);
      base.addEventListener('pointercancel', end);
      return s;
    };
    this.sticks = { l: mk('left'), r: mk('right') };
  }

  _trackStick(s, e) {
    const rect = s.base.getBoundingClientRect();
    const R = rect.width / 2;
    let x = (e.clientX - rect.left - R) / R;
    let y = (e.clientY - rect.top - R) / R;
    const len = Math.hypot(x, y);
    if (len > 1) { x /= len; y /= len; }
    s.x = x; s.y = y;
    s.knob.style.left = `${50 + x * 32}%`;
    s.knob.style.top = `${50 + y * 32}%`;
  }

  _resetStick(s) {
    s.id = null; s.x = 0; s.y = 0;
    s.knob.style.left = '50%';
    s.knob.style.top = '50%';
  }
}

const ZERO = { x: 0, y: 0 };
const HANDLED_KEYS = new Set(['KeyW', 'KeyA', 'KeyS', 'KeyD', 'ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown']);
