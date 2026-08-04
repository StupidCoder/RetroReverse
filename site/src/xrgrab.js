// xrgrab.js — moving and resizing the content with your hands.
//
// One hand moves it, two hands scale and turn it. Both drive the SAME four
// numbers the automatic fit does (see ARSession.applyPlacement): a point of the
// content, a point of the room, a scale, and a bearing. There is no separate
// manipulation transform, and nothing here touches the rig — it hands over the
// four numbers and lets the one solve do the work.
//
// Two things make it feel right rather than merely correct.
//
// NOTHING JUMPS ON GRAB. Any (hold, at) pair satisfying k·hold − R·at = t
// describes the same placement, so on grab the pair is re-expressed around the
// point you actually grabbed — the content point under your ray, or the point
// between your hands — and the rig comes out bit-identical. The alternative,
// accumulating deltas from wherever the fit left things, drifts.
//
// IT IS SOLVED, NOT INTEGRATED. Every frame re-answers "where must the content
// be so that the bit I grabbed is still under my hand", from the current pose.
// Dropped frames and lost tracking cost nothing: the next good frame is right
// again, where an integrator would have banked the error.

import { THREE } from './engine3d.js';

const UP = new THREE.Vector3(0, 1, 0);
// A hand ray is worth about half a degree, and hands are never quite still, so
// a twist this small is noise rather than intent. Subtracted rather than
// thresholded, or rotation would jump by the dead zone the moment it engaged.
const YAW_DEADZONE = 0.07; // ~4°
// How far from the fitted size the viewer may go. Wide enough to put a level on
// a table or stand inside a room of it, tight enough that a fumbled two-hand
// pinch cannot flick a mansion out of existence.
const SCALE_MIN = 1 / 8;
const SCALE_MAX = 8;

export class Grabber {
  // { ar, pickables: () => Object3D[], onChange(active) }
  constructor(opts) {
    this.ar = opts.ar;
    this.pickables = opts.pickables || (() => []);
    this.onChange = opts.onChange || null;
    this.held = new Map();   // input source -> {}
    this.mode = null;        // null | 'move' | 'scale'
    this._ray = new THREE.Raycaster();
  }

  get active() { return !!this.mode; }

  // start is offered a pinch. It only takes it if the ray is actually on the
  // content: a pinch into thin air does nothing, which is the whole reason the
  // laser and its dot exist. (The menu has already had first refusal.)
  start(src, ray) {
    if (!this.ar?.placement || !ray) return false;
    if (this.held.size === 0 && !this._hitContent(ray)) return false;
    this.held.set(src, true);
    this._rebase();
    return true;
  }

  end(src) {
    if (!this.held.delete(src)) return;
    if (this.held.size === 0) {
      this.mode = null;
      this._base = null;
      // Remember where it ended up, so the next asset arrives here too.
      if (this.ar.placement) this.ar.setUserAnchor(this.ar.placement.at);
      this.onChange?.(false);
    } else {
      this._rebase(); // two hands became one: carry on from where it is
    }
  }

  cancel() {
    if (!this.held.size) return;
    this.held.clear();
    this.mode = null;
    this._base = null;
    this.onChange?.(false);
  }

  // update is given this frame's rays, keyed by input source, in REFERENCE
  // space. Sources that stopped being tracked simply do not appear, and the
  // grab holds its ground until they come back.
  update(rays) {
    if (!this.held.size || !this._base) return;
    const live = [...this.held.keys()].map((s) => rays.get(s)).filter(Boolean);
    if (!live.length) return;
    if (this.mode === 'scale' && live.length >= 2) this._scale(live[0], live[1]);
    else this._move(live[0]);
  }

  // ---- the two gestures ------------------------------------------------------

  // One hand: the content point you grabbed stays on your ray, at the distance
  // you grabbed it. Ray-attached rather than one-to-one with the hand, because
  // an arm reaches about 0.7 m and a room is bigger than that — this way a
  // small turn of the wrist walks a diorama across the floor, and drawing your
  // hand back brings it to you.
  //
  // Position only: the wrist's roll and pitch are not applied. Taking them is
  // what makes distance-grabbing feel like holding a fishing rod in a gale.
  _move(ray) {
    const b = this._base;
    const at = ray.origin.clone().addScaledVector(ray.dir, b.dist);
    this.ar.applyPlacement({ hold: b.hold, at, k: b.k, yaw: b.yaw });
  }

  // Two hands: the gap between them sets the size and their bearing sets the
  // facing, both pivoting on the point BETWEEN your hands — which is what makes
  // it feel like zooming rather than like the model fleeing sideways.
  _scale(r1, r2) {
    const b = this._base;
    const d = r1.origin.distanceTo(r2.origin);
    if (!(d > 1e-4) || !(b.gap > 1e-4)) return;

    const k = clamp(b.k * (d / b.gap), b.kMin, b.kMax);
    const bearing = Math.atan2(r2.origin.x - r1.origin.x, r2.origin.z - r1.origin.z);
    let dYaw = wrap(bearing - b.bearing);
    dYaw = Math.abs(dYaw) < YAW_DEADZONE ? 0 : dYaw - Math.sign(dYaw) * YAW_DEADZONE;

    const at = r1.origin.clone().add(r2.origin).multiplyScalar(0.5);
    this.ar.applyPlacement({ hold: b.hold, at, k, yaw: b.yaw + dYaw });
  }

  // ---- baselines -------------------------------------------------------------

  // _rebase re-expresses the placement around whatever is being held now,
  // without moving anything. Called on every change of hands, so going from one
  // to two and back again is seamless instead of a series of jumps.
  _rebase() {
    const p = this.ar.placement;
    if (!p) return;
    this.mode = this.held.size >= 2 ? 'scale' : 'move';
    this._base = { hold: p.hold.clone(), at: p.at.clone(), k: p.k, yaw: p.yaw, dist: 0, gap: 0, bearing: 0 };
    // Finished in rebaseWith: a pinch event arrives outside the frame loop, and
    // the ray poses only exist inside it.
    this._needRays = true;
  }

  // rebaseWith finishes what _rebase started, once the frame's rays are known.
  // Split because a pinch event arrives outside the frame loop, and the ray
  // poses only exist inside it.
  rebaseWith(rays) {
    if (!this._needRays || !this._base) return;
    const live = [...this.held.keys()].map((s) => rays.get(s)).filter(Boolean);
    if (!live.length) return;
    const p = this.ar.placement;
    const M = this.ar.rig.matrixWorld;
    const b = this._base;
    b.k = p.k; b.yaw = p.yaw;
    b.kMin = (this.ar._kFit || p.k) * SCALE_MIN;
    b.kMax = (this.ar._kFit || p.k) * SCALE_MAX;

    if (this.mode === 'scale' && live.length >= 2) {
      const [r1, r2] = live;
      b.gap = r1.origin.distanceTo(r2.origin);
      b.bearing = Math.atan2(r2.origin.x - r1.origin.x, r2.origin.z - r1.origin.z);
      // The room point is the midpoint between the hands; the content point is
      // whatever is there right now — so the pair still solves to this rig.
      const mid = r1.origin.clone().add(r2.origin).multiplyScalar(0.5);
      b.at = mid;
      b.hold = mid.clone().applyMatrix4(M).divideScalar(p.k);
    } else {
      const ray = live[0];
      const hit = this._hitContent(ray);
      // Grabbed distance along the ray. With a hit that is the surface you
      // pointed at; without one (a hand that joined mid-gesture) keep the
      // distance the placement already has, which changes nothing.
      const at = hit ? hit.at : p.at.clone();
      b.dist = hit ? hit.dist : at.distanceTo(ray.origin);
      b.at = at;
      b.hold = at.clone().applyMatrix4(M).divideScalar(p.k);
    }
    this._needRays = false;
    this.onChange?.(true);
  }

  // _hitContent casts a reference-space ray at the content and reports where it
  // landed, in reference space. The ray has to go through the rig to reach
  // world space, where the content actually lives.
  _hitContent(ray) {
    const objs = this.pickables();
    if (!objs.length || !this.ar.rig) return null;
    const M = this.ar.rig.matrixWorld;
    const origin = ray.origin.clone().applyMatrix4(M);
    const dir = ray.dir.clone().transformDirection(M).normalize();
    this._ray.set(origin, dir);
    this._ray.near = 0;
    this._ray.far = Infinity;
    const hits = this._ray.intersectObjects(objs, true);
    if (!hits.length) return null;
    const inv = new THREE.Matrix4().copy(M).invert();
    const at = hits[0].point.clone().applyMatrix4(inv);
    return { at, dist: at.distanceTo(ray.origin) };
  }
}

function clamp(v, lo, hi) { return Math.max(lo, Math.min(hi, v)); }

// Shortest way round, so a twist across the ±π seam does not spin the model.
function wrap(a) {
  while (a > Math.PI) a -= 2 * Math.PI;
  while (a < -Math.PI) a += 2 * Math.PI;
  return a;
}
