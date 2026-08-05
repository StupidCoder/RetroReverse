// xremulate.js — a headset made of a mouse and a keyboard.
//
// `?xrfake=1` already runs the MENU on a desktop, but it does it by not having
// a session at all: no rig, no reference space, no poses, so `applyPlacement`
// and everything that calls it are untested by it. The last XR round's first
// three on-device bug reports all lived in exactly that gap, and world mode adds
// a placement solve, a ballistic arc and a teleport commit to it.
//
// So this is the other thing: a real `ImmersiveSession`, with a real rig and the
// real solve, fed synthetic XRFrames. The session cannot tell the difference,
// which is the whole point — what runs at the desk is the code that will run on
// the part.
//
// The mapping, chosen so the two hands a headset gives you land on things a desk
// actually has:
//
//   mouse position   the HAND. The target ray goes from just below the eye
//                    through the cursor, so pointing at the floor with the mouse
//                    is pointing at the floor with your hand.
//   left click       a pinch — selectstart, select, selectend, in that order.
//   right-drag       look around (arrow keys do it too, as FlyCam already does).
//   WASD             walk, in metres, so a 2 m door can be walked up to.
//
// What it CANNOT tell you is anything about the compositor: real IPD, real
// framebuffer scale, real reprojection, real tracking loss. It closes the gap
// between "the menu draws" and "the maths is right", not the one between a
// desktop and a Quest.

import { THREE } from './engine3d.js';

const UP = new THREE.Vector3(0, 1, 0);
const HALF_PI = Math.PI / 2 - 0.001;

export class Emulator {
  // { eyeHeight (m), speed (m/s), lookSpeed (rad/s) }
  constructor(stage, opts = {}) {
    this.stage = stage;
    this.eyeHeight = opts.eyeHeight ?? 1.62;
    this.speed = opts.speed ?? 2.2;
    this.lookSpeed = opts.lookSpeed ?? 1.4;

    // Head state, in reference space — which for a local-floor session means
    // y = 0 is the floor, so the eye starts at eye height and stays there.
    this.pos = new THREE.Vector3(0, this.eyeHeight, 0);
    this.yaw = 0;
    this.pitch = 0;
    this.ndc = new THREE.Vector2(0, -0.25); // cursor; starts a little below centre

    this._keys = new Set();
    this._look = null;      // right-drag origin
    this._last = performance.now();

    // The reference space is an opaque token everywhere it is used — the code
    // only ever passes it back to getViewerPose/getPose — so an empty object
    // with a name is a complete implementation of it.
    this.refSpace = { emulated: 'local-floor' };

    // Same for the session: `active` reads it for truthiness, `inputSources` is
    // iterated, and the three select events are subscribed to. end() has to
    // actually end, because that is how the shell's exit button gets out.
    const listeners = new Map();
    this.session = {
      emulated: true,
      inputSources: [],
      addEventListener: (t, fn) => {
        if (!listeners.has(t)) listeners.set(t, new Set());
        listeners.get(t).add(fn);
      },
      removeEventListener: (t, fn) => listeners.get(t)?.delete(fn),
      end: async () => { this.onEnd?.(); },
    };
    this._emit = (type, src) => {
      for (const fn of listeners.get(type) || []) {
        try { fn({ type, inputSource: src }); } catch (e) { console.error(`emulated ${type}`, e); }
      }
    };

    // One hand. Its targetRaySpace is, again, only ever handed back to
    // getPose(), so the token can be anything.
    this.hand = { handedness: 'right', targetRayMode: 'tracked-pointer', targetRaySpace: { emulated: 'hand' } };
    this.session.inputSources.push(this.hand);

    this._hook();
  }

  // ---- input ------------------------------------------------------------------

  _hook() {
    const cv = this.stage.canvas;
    this._onMove = (e) => {
      const r = cv.getBoundingClientRect();
      this.ndc.set(((e.clientX - r.left) / r.width) * 2 - 1, -((e.clientY - r.top) / r.height) * 2 + 1);
      if (this._look) {
        this.yaw -= (e.clientX - this._look.x) * 0.005;
        this.pitch = clamp(this.pitch - (e.clientY - this._look.y) * 0.005, -HALF_PI, HALF_PI);
        this._look = { x: e.clientX, y: e.clientY };
      }
    };
    this._onDown = (e) => {
      if (e.button === 2) { this._look = { x: e.clientX, y: e.clientY }; cv.setPointerCapture?.(e.pointerId); return; }
      if (e.button !== 0) return;
      this._pinching = true;
      this._emit('selectstart', this.hand);
    };
    this._onUp = (e) => {
      if (e.button === 2) { this._look = null; return; }
      if (e.button !== 0 || !this._pinching) return;
      this._pinching = false;
      // The real order: select lands BEFORE selectend, and code that treats a
      // pinch as "commit on release" depends on it.
      this._emit('select', this.hand);
      this._emit('selectend', this.hand);
    };
    this._onKey = (e) => {
      // Let the browser keep its own chords; only bare keys are ours.
      if (e.metaKey || e.ctrlKey || e.altKey) return;
      const on = e.type === 'keydown';
      if (on) this._keys.add(e.code); else this._keys.delete(e.code);
      if (on && MOVE.has(e.code)) e.preventDefault();
    };
    this._onMenu = (e) => e.preventDefault();
    this._onBlur = () => this._keys.clear(); // else a key held through a tab switch sticks

    cv.addEventListener('pointermove', this._onMove);
    cv.addEventListener('pointerdown', this._onDown);
    window.addEventListener('pointerup', this._onUp);
    cv.addEventListener('contextmenu', this._onMenu);
    window.addEventListener('keydown', this._onKey);
    window.addEventListener('keyup', this._onKey);
    window.addEventListener('blur', this._onBlur);
  }

  dispose() {
    const cv = this.stage.canvas;
    cv.removeEventListener('pointermove', this._onMove);
    cv.removeEventListener('pointerdown', this._onDown);
    window.removeEventListener('pointerup', this._onUp);
    cv.removeEventListener('contextmenu', this._onMenu);
    window.removeEventListener('keydown', this._onKey);
    window.removeEventListener('keyup', this._onKey);
    window.removeEventListener('blur', this._onBlur);
  }

  // ---- the frame ----------------------------------------------------------------

  // frame integrates the input and returns something shaped like an XRFrame.
  //
  // It also writes the camera's LOCAL pose, because that is a thing the runtime
  // does and nobody else will: in a session three sets camera.position and
  // camera.quaternion from the view pose each frame, and the rig turns that into
  // a world pose. Skip it and the rig moves under a camera that never does.
  frame() {
    const now = performance.now();
    const dt = Math.min(0.1, (now - this._last) / 1000);
    this._last = now;
    this._step(dt);

    const quat = new THREE.Quaternion().setFromEuler(new THREE.Euler(this.pitch, this.yaw, 0, 'YXZ'));
    this.stage.camera.position.copy(this.pos);
    this.stage.camera.quaternion.copy(quat);

    const viewer = pose(this.pos, quat);
    const handPose = this._handPose(quat);
    return {
      emulated: true,
      getViewerPose: (ref) => (ref === this.refSpace ? viewer : null),
      getPose: (space, ref) => {
        if (ref !== this.refSpace) return null;
        if (space === this.hand.targetRaySpace) return handPose;
        if (space === this.refSpace) return pose(new THREE.Vector3(), new THREE.Quaternion());
        return null; // bounded-floor and friends: absent, as on a headset without them
      },
    };
  }

  _step(dt) {
    const k = this._keys;
    const fast = k.has('ShiftLeft') || k.has('ShiftRight') ? 3 : 1;
    if (k.has('ArrowLeft')) this.yaw += this.lookSpeed * dt;
    if (k.has('ArrowRight')) this.yaw -= this.lookSpeed * dt;
    if (k.has('ArrowUp')) this.pitch = Math.min(HALF_PI, this.pitch + this.lookSpeed * dt);
    if (k.has('ArrowDown')) this.pitch = Math.max(-HALF_PI, this.pitch - this.lookSpeed * dt);

    // Walking is yaw-only: looking at your feet must not drive you into them.
    const fwd = new THREE.Vector3(-Math.sin(this.yaw), 0, -Math.cos(this.yaw));
    const right = new THREE.Vector3(Math.cos(this.yaw), 0, -Math.sin(this.yaw));
    const v = new THREE.Vector3();
    if (k.has('KeyW')) v.add(fwd);
    if (k.has('KeyS')) v.sub(fwd);
    if (k.has('KeyD')) v.add(right);
    if (k.has('KeyA')) v.sub(right);
    if (v.lengthSq() > 0) this.pos.addScaledVector(v.normalize(), this.speed * fast * dt);
    // Crouch and reach, so a floor-anchored placement can be inspected from the
    // height it will actually be seen from.
    if (k.has('KeyQ')) this.pos.y = Math.max(0.2, this.pos.y - this.speed * dt);
    if (k.has('KeyE')) this.pos.y = Math.min(3, this.pos.y + this.speed * dt);
  }

  // The hand ray: out of the eye, through the cursor. Offset down and to the
  // right so the arc visibly leaves a hand rather than the bridge of your nose —
  // and so an arc drawn from it is not permanently hidden by the near plane.
  _handPose(headQuat) {
    const cam = this.stage.camera;
    const tan = Math.tan(((cam.fov || 50) * Math.PI) / 360);
    const dir = new THREE.Vector3(this.ndc.x * tan * (cam.aspect || 1), this.ndc.y * tan, -1)
      .normalize()
      .applyQuaternion(headQuat);
    const origin = this.pos.clone()
      .add(new THREE.Vector3(0.18, -0.22, 0).applyQuaternion(headQuat));
    // Matrix4.lookAt(eye, target) puts +Z along (eye - target), so passing the
    // origin as the EYE and a point along the ray as the target leaves -Z on the
    // ray — which is where WebXR puts a target ray.
    const m = new THREE.Matrix4().lookAt(origin, origin.clone().add(dir), UP);
    const q = new THREE.Quaternion().setFromRotationMatrix(m);
    return pose(origin, q);
  }
}

const MOVE = new Set(['KeyW', 'KeyA', 'KeyS', 'KeyD', 'KeyQ', 'KeyE',
  'ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight']);

const clamp = (v, lo, hi) => (v < lo ? lo : v > hi ? hi : v);

// An XRRigidTransform, to the extent anything here reads one: position and
// orientation for the viewer pose, matrix for an input source's target ray.
function pose(p, q) {
  const m = new THREE.Matrix4().compose(p, q, new THREE.Vector3(1, 1, 1));
  return {
    transform: {
      position: { x: p.x, y: p.y, z: p.z, w: 1 },
      orientation: { x: q.x, y: q.y, z: q.z, w: q.w },
      matrix: Float32Array.from(m.elements),
    },
  };
}
