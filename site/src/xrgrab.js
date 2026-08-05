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
// How far from the fitted size the viewer may go. The fit puts a level's
// footprint at about a metre, so 32x is roughly life size — enough to stand in
// an Ultima Underworld corridor or on a Crazy Taxi street rather than look down
// at them. The lower bound only has to reach "on a table".
//
// A single two-handed pull cannot span this, and does not need to: each grab
// re-bases, so spreading, letting go and spreading again multiplies. Two or
// three ratchets is the whole range.
const SCALE_MIN = 1 / 8;
const SCALE_MAX = 32;

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
    // Never act on a baseline that has not been finished against a real ray.
    if (!this.held.size || !this._base || this._needRays) return;
    const live = [...this.held.keys()].map((s) => rays.get(s)).filter(Boolean);
    if (!live.length) return;
    if (this.mode === 'scale' && live.length >= 2) this._scale(live[0], live[1]);
    else this._move(live[0]);
  }

  // ---- the two gestures ------------------------------------------------------

  // One hand: the anchor keeps its position RELATIVE TO THE HAND — the offset
  // is stored in the hand's own frame and carried by the hand's rotation. When
  // you grabbed by pointing at something, that offset lies along the ray, so
  // this is ray-attach: a turn of the wrist walks a diorama across the floor,
  // and drawing your hand back brings it to you. That is the behaviour worth
  // having, because an arm reaches about 0.7 m and a room is bigger.
  //
  // It is stored as an offset rather than a distance so that it also works when
  // the anchor is NOT on the ray. Releasing one hand after a two-handed scale
  // leaves the anchor between where your hands were, which is off to the side
  // of whatever the remaining hand is pointing at; forcing it onto the ray
  // teleported the model by however far off-axis it happened to be.
  //
  // The hand's rotation moves the anchor, but is never applied to the content's
  // own orientation — that is what would make a distance grab feel like holding
  // a fishing rod in a gale.
  _move(ray) {
    const b = this._base;
    if (!b.off) return;
    const at = ray.origin.clone().add(b.off.clone().applyQuaternion(ray.quat));
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
    // MINUS. The yaw is the rig's, and the rig maps reference space to world
    // space — so the content's apparent rotation in the room is its inverse.
    // Working it through for a content point c about the pivot:
    //   q' − at = Rδ⁻¹ (q − at)
    // i.e. turning the rig by +δ turns what you see by −δ. Twisting your hands
    // one way and watching the model go the other is exactly this sign.
    this.ar.applyPlacement({ hold: b.hold, at, k, yaw: b.yaw - dYaw });
  }

  // ---- baselines -------------------------------------------------------------

  // _rebase re-expresses the placement around whatever is being held now,
  // without moving anything. Called on every change of hands, so going from one
  // to two and back again is seamless instead of a series of jumps.
  _rebase() {
    const p = this.ar.placement;
    if (!p) return;
    this.mode = this.held.size >= 2 ? 'scale' : 'move';
    // off/gap stay null until rebaseWith has seen a ray. _move and _scale both
    // refuse to act without them: acting on a zero offset would snap the
    // content onto the hand, which is the worst thing a grab can do.
    this._base = { hold: p.hold.clone(), at: p.at.clone(), k: p.k, yaw: p.yaw, off: null, gap: 0, bearing: 0 };
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
      // Anchor on the surface you pointed at, if you pointed at one. If not —
      // the hand left over from a two-handed scale, which is aimed at nothing
      // in particular — anchor on where the content already is. Either way the
      // offset is taken in the HAND'S frame, so nothing has to lie on the ray
      // and nothing moves.
      const hit = this._hitContent(ray);
      const at = hit ? hit.at : p.at.clone();
      b.at = at;
      b.hold = at.clone().applyMatrix4(M).divideScalar(p.k);
      b.off = at.clone().sub(ray.origin).applyQuaternion(ray.quat.clone().invert());
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
    return { at: hits[0].point.clone().applyMatrix4(inv) };
  }
}

function clamp(v, lo, hi) { return Math.max(lo, Math.min(hi, v)); }

// Shortest way round, so a twist across the ±π seam does not spin the model.
function wrap(a) {
  while (a > Math.PI) a -= 2 * Math.PI;
  while (a < -Math.PI) a += 2 * Math.PI;
  return a;
}
