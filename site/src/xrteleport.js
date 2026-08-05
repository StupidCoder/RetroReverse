// xrteleport.js — moving around a level you are standing inside.
//
// The arc is a thrown ball, not a laser, for the reason every VR title settles
// on one: a straight ray aimed at a floor ten metres off is a two-degree gesture
// and lands where it likes, while an arc lands where it falls. It is also the
// only shape that answers correctly in a dungeon with a ledge over a floor —
// see the header of xrfloor.js, which does the intersecting.
//
// Two things here are less obvious than they look.
//
//   THE WHOLE ARC IS IN CONTENT UNITS. The hand pose arrives in reference space
//   and the floor index is built in the content's own coordinates, so SOMETHING
//   has to convert. Converting the hand ray once — and dividing gravity and
//   speed by the scene scale with it — is one conversion a frame instead of
//   twenty, keeps the index valid when the scale is retuned, and lets the marker
//   hang in the content group with no counter-scale of its own.
//
//   THE COMMIT IS DEFERRED TO A FRAME. `selectend` arrives from a session event
//   listener, and computing where you land needs the head pose, which exists
//   only inside an XRFrame. So the release sets a flag and the next frame spends
//   it. xrgrab.js already had to solve this (its _needRays/rebaseWith split);
//   this is the same shape.

import { THREE } from './engine3d.js';
import { arcSamples, hitArc, blocked, floorUnder } from './xrfloor.js';

const UP = new THREE.Vector3(0, 1, 0);
const G = 9.81;

export class Teleporter {
  // { ar, cfg, group, floor: () => grid, blockers: () => grid|null,
  //   onTeleport(point), onStatus() }
  constructor(opts) {
    this.ar = opts.ar;
    this.cfg = opts.cfg;
    this.floor = opts.floor;
    this.blockers = opts.blockers || (() => null);
    this.onTeleport = opts.onTeleport || null;
    this.onChange = opts.onChange || null;

    this.aiming = null;   // the input source currently pinching
    this.hit = null;      // { point, normal } in content units
    this.reason = '';     // why the current aim will not commit
    this._commit = null;  // set on release, spent on the next frame
    this._turn = 0;       // snap-turn edge detector

    this.group = new THREE.Group();
    this.group.name = 'xr-teleport';
    this.group.visible = false;
    opts.group.add(this.group);
    this._build();
  }

  _build() {
    const n = this.cfg.teleport.steps;
    // A line, not a tube: it is drawn depth-free over everything (you must be
    // able to see where you are aiming through the wall you are aiming past),
    // and a tube's triangles would be rebuilt every frame for no gain.
    const geo = new THREE.BufferGeometry();
    geo.setAttribute('position', new THREE.BufferAttribute(new Float32Array((n + 1) * 3), 3));
    this.line = new THREE.Line(geo, new THREE.LineBasicMaterial({
      color: 0x8fd0ff, transparent: true, opacity: 0.85, depthTest: false, depthWrite: false, fog: false,
    }));
    this.line.renderOrder = 1005;
    this.line.frustumCulled = false;
    this.group.add(this.line);

    // The marker is a ring rather than a disc so you can see the floor you are
    // about to stand on through it, and it is built at unit radius and SCALED —
    // the group lives in content units and the radius is authored in metres.
    this.ring = new THREE.Mesh(
      new THREE.RingGeometry(0.78, 1, 32),
      new THREE.MeshBasicMaterial({
        color: 0x8fd0ff, transparent: true, opacity: 0.9, side: THREE.DoubleSide,
        depthTest: false, depthWrite: false, fog: false,
      }),
    );
    this.ring.renderOrder = 1006;
    this.ring.frustumCulled = false;
    this.ring.visible = false;
    this.group.add(this.ring);
  }

  // ---- the interactor contract the shell talks to -------------------------------

  selectStart(src) {
    this.aiming = src;
    this.group.visible = true;
    this.onChange?.();
  }

  selectEnd(src) {
    if (this.aiming !== src) return;
    this.aiming = null;
    this.group.visible = false;
    // Not "teleport now" — "teleport on the next frame that has a head pose".
    if (this.hit && !this.reason) this._commit = this.hit.point.slice();
    this.hit = null;
    this.onChange?.();
  }

  cancel() {
    this.aiming = null;
    this._commit = null;
    this.hit = null;
    this.group.visible = false;
  }

  frame(frame, ref, rays) {
    this._snapTurn(frame, ref);
    if (this._commit) {
      const to = this._commit;
      this._commit = null;
      this._teleport(frame, ref, to);
      return;
    }
    if (!this.aiming) return;
    const ray = rays?.get(this.aiming);
    if (!ray) { this.ring.visible = false; return; }
    this._aim(ray);
  }

  get note() {
    if (!this.aiming) return '';
    return this.reason ? ` · ${this.reason}` : ' · teleport ready';
  }

  dispose() {
    this.group.removeFromParent();
    this.line.geometry.dispose();
    this.line.material.dispose();
    this.ring.geometry.dispose();
    this.ring.material.dispose();
  }

  // ---- aiming -----------------------------------------------------------------

  _aim(ray) {
    const k = this.ar.stage.scene.scale.x || 1;
    const rig = this.ar.rig;
    const t = this.cfg.teleport;
    if (!rig) return;

    // Reference space -> world (through the rig) -> content (divide out the
    // scene scale, which is a pure uniform, so this is exact).
    const origin = ray.origin.clone().applyMatrix4(rig.matrixWorld).divideScalar(k);
    const dir = ray.dir.clone().transformDirection(rig.matrixWorld);

    const pts = arcSamples(
      [origin.x, origin.y, origin.z],
      [dir.x, dir.y, dir.z],
      t.speed / k, G / k, t.flightTime, t.steps,
      this._pts || (this._pts = new Float32Array((t.steps + 1) * 3)),
    );
    const hit = hitArc(this.floor(), pts);

    this.hit = hit ? { point: hit.point, normal: hit.normal } : null;
    this.reason = hit ? this._reject(hit, origin, k) : 'no floor there';
    this._draw(pts, hit, k);
  }

  // _reject is where a landing that the geometry allows and the player should
  // not get is turned down. Each of these is a real way to break a level.
  _reject(hit, origin, k) {
    const t = this.cfg.teleport;
    const d = Math.hypot(hit.point[0] - origin.x, hit.point[2] - origin.z) * k;
    if (d > t.maxRange) return `too far (${d.toFixed(0)} m)`;
    // Climbing. The arc happily lands on a rooftop or the top of a wall it
    // passed over; a step up you could not have climbed is not somewhere to be.
    const rise = (hit.point[1] - this._standY(origin)) * k;
    if (Number.isFinite(rise) && rise > t.maxRise) return `too high (${rise.toFixed(1)} m up)`;
    // Walls. The floor index knows nothing about them, so an arc thrown over a
    // partition lands cleanly in the next room.
    if (t.blockers) {
      const walls = this.blockers();
      if (walls && blocked(walls, [origin.x, origin.y, origin.z], hit.point)) return 'no line of sight';
    }
    return '';
  }

  // Where the thrower is standing, in content units — the floor under the hand,
  // not the hand. -Infinity when there is nothing underneath, which makes the
  // rise test pass: refusing to teleport because we could not work out where you
  // already are would be the worst of both answers.
  _standY(origin) {
    const g = this.floor();
    const y = g && floorUnder(g, origin.x, origin.y, origin.z);
    return y == null ? -Infinity : y;
  }

  _draw(pts, hit, k) {
    const n = hit ? hit.seg + 1 : this.cfg.teleport.steps;
    const pos = this.line.geometry.attributes.position;
    const arr = pos.array;
    for (let i = 0; i <= this.cfg.teleport.steps; i++) {
      const j = Math.min(i, n) * 3;
      // Past the hit the polyline collapses onto the landing point, so the arc
      // visibly STOPS at the floor instead of carrying on through it.
      const src = hit && i > n ? [hit.point[0], hit.point[1], hit.point[2]] : [pts[j], pts[j + 1], pts[j + 2]];
      arr[i * 3] = src[0]; arr[i * 3 + 1] = src[1]; arr[i * 3 + 2] = src[2];
    }
    pos.needsUpdate = true;
    this.line.geometry.computeBoundingSphere();

    const ok = hit && !this.reason;
    this.line.material.color.set(ok ? 0x8fd0ff : 0xff8a6b);
    this.line.material.opacity = ok ? 0.85 : 0.4;
    this.ring.visible = !!ok;
    if (!ok) return;

    const r = this.cfg.teleport.markerRadius / k;
    this.ring.scale.setScalar(r);
    // Lifted off the surface along its own normal: a ring coplanar with the
    // floor z-fights it even with depthTest off, because the two are the same
    // pixels.
    this.ring.position.set(
      hit.point[0] + hit.normal[0] * r * 0.05,
      hit.point[1] + hit.normal[1] * r * 0.05,
      hit.point[2] + hit.normal[2] * r * 0.05,
    );
    this.ring.quaternion.setFromUnitVectors(
      new THREE.Vector3(0, 0, 1),
      new THREE.Vector3(hit.normal[0], hit.normal[1], hit.normal[2]),
    );
  }

  // ---- committing ----------------------------------------------------------------

  // The teleport itself is one applyPlacement — the SAME call the initial spawn
  // and the diorama fit make. `hold` becomes the spot you aimed at and `at`
  // becomes where you are standing right now, so the floor you picked arrives
  // under your feet at your own floor's height. Scale and bearing are untouched:
  // a teleport moves you, it does not re-frame the world.
  _teleport(frame, ref, to) {
    const pose = frame.getViewerPose(ref);
    const pl = this.ar.placement;
    if (!pose || !pl) return;
    const p = pose.transform.position;
    this.ar.applyPlacement({
      hold: new THREE.Vector3(to[0], to[1], to[2]),
      at: new THREE.Vector3(p.x, 0, p.z),
      k: pl.k,
      yaw: pl.yaw,
    });
    this.onTeleport?.(to);
    this.onChange?.();
  }

  // A flick of the thumbstick turns the world instead of you: a guardian is not
  // a swivel chair, and teleporting into a corridor facing a wall is otherwise
  // a dead end. Re-expressed around the HEAD so nothing translates — the same
  // trick xrgrab.js uses to re-anchor a grab without anything jumping.
  _snapTurn(frame, ref) {
    const deg = this.cfg.teleport.snapTurn;
    const pl = this.ar.placement;
    if (!deg || !pl) return;
    let axis = 0;
    for (const src of this.ar.session?.inputSources || []) {
      const a = src.gamepad?.axes;
      if (a && a.length >= 4 && Math.abs(a[2]) > Math.abs(axis)) axis = a[2];
    }
    if (Math.abs(axis) < 0.7) { this._turn = 0; return; }
    if (this._turn) return;                 // held: one turn per push
    this._turn = Math.sign(axis);
    const pose = frame.getViewerPose(ref);
    if (!pose) return;
    const p = pose.transform.position;
    const yaw = pl.yaw - this._turn * (deg * Math.PI) / 180;
    // Rotating about the head means the head must be BOTH the room point and
    // (through the current placement) the content point it maps to, or the world
    // swings you around its own origin.
    const rig = this.ar.rig;
    const at = new THREE.Vector3(p.x, 0, p.z);
    const hold = at.clone().applyMatrix4(rig.matrixWorld).divideScalar(pl.k);
    this.ar.applyPlacement({ hold: new THREE.Vector3(hold.x, pl.hold.y, hold.z), at, k: pl.k, yaw });
    this.onChange?.();
  }
}
