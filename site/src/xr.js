// xr.js — immersive-AR ("diorama") mode for the 3D stages. The level is put
// into the room as a tabletop model floating in front of the viewer: a metre
// across by default, whatever the game's own units happen to be.
//
// Two transforms do the work, and WHICH one carries the scale matters:
//
//   scene.scale  — metres per world unit, so the level reads a metre across.
//   rig          — the camera's parent: position and yaw only, UNIT SCALE,
//                  and deliberately NOT a child of the scene.
//
// The scale cannot live on the rig. three's setProjectionFromUnion builds the
// frustum it culls against by measuring the IPD from cameraL/cameraR's
// matrixWorld — world units — while taking near and the fov extents from the
// projection matrices, which are metres. Under a rig scaled by 16 the IPD
// comes out 16× too large, zOffset with it, and `near2 = near + zOffset` puts
// the culling near plane at about half a metre: everything you lean toward
// vanishes whole. Scaling the scene instead leaves the camera in metres,
// which is the space three's XR path assumes.
//
// The rig is kept out of the scene graph for the same reason — as a child it
// would inherit scene.scale and hand the camera a non-unit scale again. Being
// parentless, its world matrix is its own, and _frame refreshes it each frame
// before three reads it.
//
// There is no console on the headset, so every failure has to be legible on
// the page itself: `status` reports what happened and survives session exit.

import { THREE } from './engine3d.js';

// Re-exported for the callers that already have three.js loaded; it lives in
// its own dependency-free module so a 2-D page can ask without importing any
// of this. See xrsupport.js.
import { arSupported } from './xrsupport.js';

export { arSupported };

// PerfMeter turns the frame clock and the renderer's own counters into one
// readout line, averaged over a window of frames. A headset has no devtools
// and no console, so the only honest way to know what a scene costs there is
// to make the scene say it: the numbers that matter are the ones measured on
// the part, not on a desktop profile of the same page.
//
// It reads Stage.perf (cumulative CPU milliseconds, split into the updaters
// and the render submission) and takes deltas, so it can be driven from a
// stage updater — which runs BEFORE the render, meaning both the renderer's
// counters and the CPU totals it sees lag by one frame. Over a 45-frame
// window that is noise, and an updater is the only place a view can hook
// without owning the loop.
export class PerfMeter {
  constructor(opts = {}) {
    this.every = opts.every || 45; // frames per readout (~0.5 s at 90 Hz)
    this.reset();
  }

  reset() {
    this._n = 0;
    this._t = 0;
    this._worst = 0;
    this._base = null;
  }

  // sample returns the readout line when a window closes, null otherwise.
  sample(dt, stage) {
    const p = stage.perf;
    if (!this._base) this._base = { upd: p.upd, draw: p.draw };
    this._n++;
    this._t += dt;
    this._worst = Math.max(this._worst, dt);
    if (this._n < this.every) return null;
    const n = this._n;
    const info = stage.renderer.info;
    const upd = (p.upd - this._base.upd) / n;
    const draw = (p.draw - this._base.draw) / n;
    // calls/triangles are per render() — in a session that is BOTH eyes,
    // since three draws the stereo pair inside one call. tex and prog are
    // totals resident on the GPU: what a draw call has to rebind between.
    const note = `${Math.round(n / this._t)} fps (worst ${Math.round(this._worst * 1000)} ms)`
      + ` · ${info.render.calls} calls · ${info.render.triangles} tris`
      // Resident geometries and textures are the leak instrument: a shell that
      // swaps content in a live session has no renderer teardown to fall back
      // on, so these must come back to a flat baseline after a switch.
      + ` · ${info.memory.geometries} geo · ${info.memory.textures} tex · ${info.programs?.length ?? 0} prog`
      + ` · cpu ${(upd + draw).toFixed(2)} ms (upd ${upd.toFixed(2)} · draw ${draw.toFixed(2)})`;
    this.reset();
    this._base = { upd: p.upd, draw: p.draw };
    return note;
  }
}

const UP = new THREE.Vector3(0, 1, 0);
const POSE_GIVE_UP = 120; // frames (~2 s) to wait for tracking before saying so
const SPAWN_KEY = 'rx.xr.spawn'; // the chosen floor spot, in bounded-floor coords

export class ARSession {
  // { stage, contentBox: () => THREE.Box3|null, targetSize (m), distance (m),
  //   dropBelowEye (m), onChange(active), onStatus(text) }
  constructor(opts) {
    this.stage = opts.stage;
    this.contentBox = opts.contentBox;
    this.targetSize = opts.targetSize || 1.0;
    this.distance = opts.distance ?? 1.5;
    this.dropBelowEye = opts.dropBelowEye ?? 0.15;
    // Where the content meets the room. 'eye' hangs its CENTRE at head height
    // (a diorama, a wall map); 'floor' stands its LOWEST point on the physical
    // floor. The latter is free: the session already requests a local-floor
    // reference space, in which y = 0 is the floor the headset measured, so
    // the object only has to be anchored by its underside instead of its
    // middle.
    this.anchor = opts.anchor === 'floor' ? 'floor' : 'eye';
    // Which extent the targetSize applies to: a level wants its FOOTPRINT a
    // metre across (height follows), a single object its longest axis, so a
    // tall thin model does not come out enormous.
    this.fitAxis = opts.fitAxis || 'horizontal'; // 'horizontal' | 'longest'
    // A wall-sized map does not fit an AXIS, it fits a BOX. Given the content
    // size in world units, return metres per world unit; overrides
    // fitAxis/targetSize entirely.
    this.fitScale = opts.fitScale || null; // (size: Vector3) => number
    // Content that can only be BUILT once the session is running — a map
    // snapshot read out of another renderer — reports through here. While it
    // is false the placement waits instead of fitting an empty scene.
    this.ready = opts.ready || null; // () => boolean
    // The world-space horizontal direction the viewer should be looking ALONG
    // — i.e. the shot the diorama should present. A level passes its
    // document's establishing camera; the object viewer passes a function, so
    // whichever side you had orbited to is the side you get. Defaults to
    // looking toward +Z (a viewer on the −Z side).
    this.frontDir = opts.frontDir || new THREE.Vector3(0, 0, 1);
    this.onScene = opts.onScene || null;   // the view's own scene tweaks (sky, cutscenes)
    this.onChange = opts.onChange || null; // the shell's button/filter state, assigned later
    this.onStatus = opts.onStatus || null;
    // Hooks for a shell that keeps ONE session across many contents.
    // onFrame gets every XR frame (the menu raycasts its pointer there);
    // onSelect gets first refusal on a pinch and returns true if it consumed
    // it, so pointing at a menu row does not also re-place the scene behind it;
    // onPlaced fires once each time a fit succeeds, which is when new content
    // is safe to reveal.
    this.onFrame = opts.onFrame || null;
    this.onSelect = opts.onSelect || null;
    this.onSelectStart = opts.onSelectStart || null;
    this.onSelectEnd = opts.onSelectEnd || null;
    this.onPlaced = opts.onPlaced || null;
    this.placement = null;  // { hold, at, k, yaw } — see applyPlacement
    this._userAt = null;    // where the viewer last dragged content to
    // Carried on the instance so the shell can gate its button without
    // importing this module (and pulling three.js into the landing page).
    this.supported = arSupported;
    this.session = null;
    this.rig = null;
    this._place = false;
    this._waited = 0;
    this._fit = '';
  }

  get active() { return !!this.session; }

  status(text) { this.onStatus?.(text); }

  // setContent points a LIVE session at different content. Everything the fit
  // depends on is re-read here and a placement is requested; nothing about the
  // session itself is touched, which is the whole point — the alternative is
  // ending it and asking for a new one, and WebXR wants a user gesture for
  // that which an asset load has long since spent.
  setContent(c = {}) {
    if (c.contentBox) this.contentBox = c.contentBox;
    if (c.frontDir) this.frontDir = c.frontDir;
    this.ready = c.ready || null;
    const fit = c.fit || {};
    this.targetSize = fit.targetSize ?? 1.0;
    this.distance = fit.distance ?? 1.5;
    this.dropBelowEye = fit.dropBelowEye ?? 0.15;
    this.anchor = fit.anchor === 'floor' ? 'floor' : 'eye';
    this.fitAxis = fit.fitAxis || 'horizontal';
    this.fitScale = fit.fitScale || null;
    this.rePlace();
  }

  // rePlace asks for a fresh fit on the next frame that has a head pose. It is
  // also the way back from a drag that went wrong — hence forgetting where the
  // viewer last put things, so the spawn logic gets to choose again.
  rePlace() {
    this._userAt = null;
    this._place = true;
    this._waited = 0;
  }

  // applyPlacement is the whole of "where is the content" — and the only place
  // that writes the rig. Four numbers say it:
  //
  //   hold  a point of the CONTENT, in unscaled content units
  //   at    a point of the ROOM, in reference space
  //   k     metres per content unit
  //   yaw   the content's bearing
  //
  // and they mean "put hold at at, this big, facing this way". The rig maps
  // reference space to world space, so with the rig's rotation R the solve is
  // t = k·hold − R·at.
  //
  // Everything places through this: the automatic fit, and the hands. That is
  // what makes grabbing cheap — dragging is not a new transform stack, it is
  // the same four numbers with `at` coming from a hand instead of from a spawn
  // point. It also means a grab can re-express (hold, at) for the point you
  // actually grabbed WITHOUT anything moving, because any pair satisfying
  // k·hold − R·at = t describes the same placement.
  applyPlacement({ hold, at, k, yaw }) {
    const rig = this.rig;
    if (!rig) return;
    this.placement = { hold: hold.clone(), at: at.clone(), k, yaw };
    const scene = this.stage.scene;
    scene.scale.setScalar(k);
    scene.updateMatrixWorld(true);
    rig.position.set(0, 0, 0);
    rig.quaternion.setFromAxisAngle(UP, yaw);
    rig.scale.set(1, 1, 1);
    rig.updateMatrixWorld(true);
    rig.position.copy(hold).multiplyScalar(k).sub(at.clone().applyMatrix4(rig.matrixWorld));
    rig.updateMatrixWorld(true);

    // The metres ACHIEVED, not the metres requested — more use to both callers.
    // The head's height is the tell for whether local-floor is being measured
    // or guessed: it should read as your actual eye height. If it does and the
    // content still looks wrong, the placement is at fault; if it reads far off,
    // the runtime's floor is, and no amount of arithmetic here fixes that.
    const s = this._size;
    if (s) {
      this._fit = `${fmt(s.x)}×${fmt(s.y)}×${fmt(s.z)} u · `
        + `${fmt(s.x * k)}×${fmt(s.y * k)}×${fmt(s.z * k)} m · ${this.anchor}`
        + (this._kFit ? ` · ${(k / this._kFit).toFixed(2)}× fit` : '')
        + ` · eye ${fmt(this._eye ?? 0)} m`
        + (this._spawnNote ? ` · ${this._spawnNote}` : '');
    }
  }

  // setUserAnchor remembers where the viewer dragged the content to, so the
  // NEXT asset arrives in the same place — you arrange your room once, not once
  // per asset. Stored in bounded coords on the next frame that has one, which
  // is what makes it survive the session (see _store).
  setUserAnchor(at) {
    this._userAt = at.clone();
    this._storePending = true;
  }

  // reattach re-installs the per-frame hook on the stage. A shell that swaps
  // content resets the stage between assets, and anything it clears that
  // belongs to the SESSION has to come back — losing this one costs the
  // pointer, the placement solve and the menu's anchoring at once, with no
  // error anywhere. Cheap enough to call after every swap rather than trust
  // that the reset left it alone.
  reattach() {
    if (this.session) this.stage.onXRFrame = (frame) => this._frame(frame);
  }

  async enter() {
    if (this.session) return;
    const stage = this.stage;
    this.status('starting AR…');
    // Before the snapshot: the view's tweaks include stopping a running
    // cutscene, and a cutscene player restores its own camera pose on dispose
    // — which must land before we record the pose to come back to.
    this.onScene?.(true);
    let session;
    try {
      session = await navigator.xr.requestSession('immersive-ar', {
        // local-floor must be REQUIRED: renderer.xr.setSession awaits
        // session.requestReferenceSpace() internally, and a rejection there
        // means the animation loop never starts — a black session with no
        // error anywhere.
        requiredFeatures: ['local-floor'],
        // bounded-floor is what makes a spawn point meaningful: its origin is
        // fixed to the guardian, so a point expressed in it means the same
        // physical spot next session, and its boundsGeometry is the play area.
        // hand-tracking is not needed to POINT — a pinch already arrives as a
        // select from a tracked-pointer input source — but asking costs nothing
        // and is what would let a ray start at the fingertip rather than at the
        // runtime's own grip pose.
        optionalFeatures: ['dom-overlay', 'bounded-floor', 'hand-tracking'],
        domOverlay: { root: document.body },
      });
    } catch (e) {
      this.onScene?.(false);
      this.status(`AR unavailable: ${e.message || e.name}`);
      throw e;
    }

    this.session = session;
    this._snap = stage.cameraSnapshot();
    // Remember WHICH scene these belong to. A shell that swaps content replaces
    // stage.scene outright, and restoring the first level's fog onto the last
    // one's scene at exit would be a haze nobody asked for.
    this._scene = stage.scene;
    this._bg = stage.scene.background;
    this._fog = stage.scene.fog;
    // A Color background is harmless (three forces a transparent clear for an
    // alpha-blend environment), but a TEXTURE background draws a real box mesh
    // that would paint over passthrough, and fog hangs coloured haze in the
    // room. Both go for the duration.
    stage.scene.background = null;
    stage.scene.fog = null;
    stage.controls.enabled = false;
    stage.inputEnabled = false;

    this._sceneScale = stage.scene.scale.clone();
    this.rig = new THREE.Group();
    this.rig.name = 'xr-rig';
    this.rig.add(stage.camera); // parentless on purpose — see the note at the top
    // In a session the near/far planes are read in METRES and pushed to the
    // compositor as depthNear/depthFar, whatever the world units are.
    stage.camera.near = 0.02;
    stage.camera.far = 100;
    stage.camera.updateProjectionMatrix();

    this._place = true;
    this._waited = 0;
    this.reattach();
    // The controller trigger re-places the diorama in front of wherever you
    // are now looking (and re-measures the content, so a level whose rooms
    // streamed in after entry gets refitted).
    // The trigger re-places the content. If a controller is pointing at the
    // floor it goes THERE — which is the whole point in a small room, where
    // "a fixed distance ahead" usually lands you against a wall with no way to
    // walk around the far side.
    // A pinch is either a click on the menu or a grab of the content; it no
    // longer re-places anything. Pointing at the floor to place used to live
    // here, and dragging the content where you want it is strictly better —
    // "Recentre" on the panel is the way back.
    session.addEventListener('select', (e) => this.onSelect?.(e.inputSource));
    session.addEventListener('selectstart', (e) => this.onSelectStart?.(e.inputSource));
    session.addEventListener('selectend', (e) => this.onSelectEnd?.(e.inputSource));
    // Restore off three's OWN sessionend, not the session's 'end'. A listener
    // on the session fires in registration order, and ours would be added
    // before three's (setSession registers it) — so it would run while
    // isPresenting is still true and before the drawing buffer is restored,
    // and the aspect fix inside cameraRestore would be skipped, leaving the
    // desktop view stretched. three dispatches 'sessionend' last, after
    // clearing the flag and restoring the size.
    this._onEnd = () => this._end();
    stage.renderer.xr.addEventListener('sessionend', this._onEnd);

    // Asked for SEPARATELY rather than as the session's reference space: if
    // the runtime cannot provide it, setSession would reject and the session
    // would come up black. This way it is simply absent and the fallbacks
    // apply.
    this._bounded = await session.requestReferenceSpace('bounded-floor').catch(() => null);

    stage.renderer.xr.setReferenceSpaceType('local-floor');
    try {
      await stage.renderer.xr.setSession(session);
    } catch (e) {
      // three never took the session over, so its sessionend will not fire:
      // unwind by hand.
      session.end().catch(() => {});
      this._end();
      this.status(`AR session failed: ${e.message || e.name}`);
      throw e;
    }
    this.status('in AR — trigger re-places the scene');
    this.onChange?.(true);
  }

  exit() {
    this.session?.end().catch(() => {});
  }

  // _frame runs once per XR frame, before the scene updaters and before the
  // render. Beyond placing the diorama it keeps the rig's world matrix fresh:
  // the rig has no parent, so scene.updateMatrixWorld() inside render will not
  // reach it, and three reads parent.matrixWorld when it composes the eye
  // cameras.
  _frame(frame) {
    this.rig?.updateMatrixWorld(true);
    // A drag finished: write where it ended up, in bounded coords so it means
    // the same physical spot next session. Deferred to a frame because that is
    // where the bounded-space pose lives.
    if (this._storePending && this._userAt && frame) {
      const ref = this.stage.renderer.xr.getReferenceSpace();
      if (ref) { this._store(frame, ref, this._userAt); this._storePending = false; }
    }
    // The menu aims its pointer here: it needs the frame and the reference
    // space, and this is the one callback that has both. Before the placement
    // work below, so a pointer keeps tracking even while content is streaming.
    if (frame && this.onFrame) {
      const ref = this.stage.renderer.xr.getReferenceSpace();
      if (ref) { try { this.onFrame(frame, ref); } catch (e) { console.error('xr onFrame', e); } }
    }
    if (!this._place || !frame) return;
    // Not "nothing to place" — not YET. Reset the pose counter too, or a slow
    // build would trip POSE_GIVE_UP and report a tracking failure that never
    // happened.
    if (this.ready && !this.ready()) { this._waited = 0; return; }
    const ref = this.stage.renderer.xr.getReferenceSpace();
    const pose = ref && frame.getViewerPose(ref);
    if (!pose) {
      // Tracking can take a few frames to settle; only complain if it never does.
      if (++this._waited === POSE_GIVE_UP) this.status('AR: no head pose — tracking not available');
      return;
    }
    // Measure unscaled: contentBox reads world matrices, and on a recenter the
    // scene already carries the previous fit's scale.
    const scene = this.stage.scene;
    scene.scale.set(1, 1, 1);
    scene.updateMatrixWorld(true);
    const box = this.contentBox?.();
    if (!box || box.isEmpty()) {
      scene.scale.copy(this._sceneScale);
      scene.updateMatrixWorld(true);
      this._place = false;
      this.status('AR: nothing to place (empty content box)');
      return;
    }

    const p = pose.transform.position, o = pose.transform.orientation;
    const q = new THREE.Quaternion(o.x, o.y, o.z, o.w);
    // Yaw only: the diorama sits level however the viewer's head is tilted.
    const fwd = new THREE.Vector3(0, 0, -1).applyQuaternion(q);
    fwd.y = 0;
    if (fwd.lengthSq() < 1e-6) fwd.set(0, 0, -1); // straight up/down: fall back to −Z
    fwd.normalize();

    const size = box.getSize(new THREE.Vector3());
    const centre = box.getCenter(new THREE.Vector3());
    // Which point of the content is being placed, and where it goes. For a
    // floor anchor both drop to ground level: the box's bottom-centre, onto
    // y = 0 ahead of the viewer. fwd is already flattened, so the target stays
    // on the floor plane.
    const hold = this.anchor === 'floor'
      ? new THREE.Vector3(centre.x, box.min.y, centre.z)
      : centre;
    const spawn = this._spawnPoint(frame, ref);
    // A spot the viewer dragged content to is taken WHOLE, height included: if
    // you lifted the last map to eye level, the next one belongs there too.
    // Everything else only contributes x/z, and the height comes from the
    // anchor — so an eye-level diorama hangs ABOVE the place a floor model
    // would stand on.
    const at = this._userAt ? this._userAt.clone()
      : spawn
        ? new THREE.Vector3(spawn.x, this.anchor === 'floor' ? 0 : p.y - this.dropBelowEye, spawn.z)
        : this.anchor === 'floor'
          ? new THREE.Vector3(p.x, 0, p.z).addScaledVector(fwd, this.distance)
          : new THREE.Vector3(p.x, p.y - this.dropBelowEye, p.z).addScaledVector(fwd, this.distance);
    if (this._userAt) this._spawnNote = 'spawn: where you put it';
    const span = this.fitAxis === 'longest'
      ? Math.max(size.x, size.y, size.z)
      : Math.max(size.x, size.z) || Math.max(size.y, 1);
    // metres per world unit
    const k = this.fitScale ? this.fitScale(size) : this.targetSize / (span || 1);

    // Yaw so that the viewer's forward maps onto the requested establishing
    // direction: we need R·fwd = frontDir, and for horizontal unit vectors
    // that is the difference of their bearings. (Rotating BY the head's own
    // bearing instead of by the difference turns the diorama the wrong way —
    // it is the inverse rotation that carries fwd onto a fixed axis.)
    const front = typeof this.frontDir === 'function' ? this.frontDir() : this.frontDir;
    const yaw = Math.atan2(front.x, front.z) - Math.atan2(fwd.x, fwd.z);

    this._size = size;
    this._eye = p.y;
    this._kFit = k;                 // what the grab's scale limits are measured against
    this.applyPlacement({ hold, at, k, yaw });

    this._place = false;
    this.status(`in AR · ${this._fit}`);
    // New content is built hidden and revealed here: until a fit has landed it
    // would be drawn at the PREVIOUS content's scale and rig offset, which for
    // one or two frames looks like the room lurching.
    this.onPlaced?.(this._fit);
  }

  // _spawnPoint resolves where content should stand, in reference space, in
  // this order: a spot stored from a previous session, the centre of the play
  // area, or nothing (meaning "a fixed distance ahead").
  //
  // The spot pointed at with a trigger used to come first. Dragging content
  // where you want it replaced it — and now WRITES the stored spot, so the
  // arrangement still survives the session, just chosen by hand instead of by
  // aiming at the floor.
  _spawnPoint(frame, ref) {
    const b = this._bounded;
    if (!b) { this._spawnNote = 'spawn: ahead (no bounded-floor)'; return null; }
    const pose = frame.getPose(b, ref);
    if (!pose) return null;
    const m = new THREE.Matrix4().fromArray(pose.transform.matrix);

    const saved = readStored();
    if (saved) {
      this._spawnNote = 'spawn: saved';
      return new THREE.Vector3(saved.x, 0, saved.z).applyMatrix4(m).setY(0);
    }
    // The play area's centroid — which is what "somewhere I can walk all the
    // way around" means, and needs no interaction at all.
    const pts = b.boundsGeometry;
    if (pts?.length >= 3) {
      let x = 0, z = 0;
      for (const q of pts) { x += q.x; z += q.z; }
      this._spawnNote = `spawn: play-area centre (${pts.length}pt)`;
      return new THREE.Vector3(x / pts.length, 0, z / pts.length).applyMatrix4(m).setY(0);
    }
    this._spawnNote = 'spawn: ahead (no bounds)';
    return null;
  }

  // Stored in BOUNDED coordinates. local-floor's origin is re-established every
  // session, so a point saved in it would land somewhere new each time; the
  // bounded space is pinned to the guardian, so it does not.
  _store(frame, ref, pt) {
    const b = this._bounded;
    if (!b) return;
    const pose = frame.getPose(ref, b);
    if (!pose) return;
    const v = pt.clone().applyMatrix4(new THREE.Matrix4().fromArray(pose.transform.matrix));
    try { localStorage.setItem(SPAWN_KEY, JSON.stringify({ x: v.x, z: v.z })); } catch { /* private mode */ }
  }

  _end() {
    if (!this.session) return; // both exit() and the headset's own exit land here
    const stage = this.stage;
    if (this._onEnd) {
      stage.renderer.xr.removeEventListener('sessionend', this._onEnd);
      this._onEnd = null;
    }
    stage.onXRFrame = null;
    if (this.rig) {
      this.rig.remove(stage.camera); // the rig was never in the scene
      this.rig = null;
    }
    // Onto the scene these were taken FROM, which a content swap may have
    // replaced since. Restoring them onto whatever is current instead would
    // hand the last level the first one's fog.
    const scene = this._scene || stage.scene;
    if (this._sceneScale) {
      scene.scale.copy(this._sceneScale);
      scene.updateMatrixWorld(true);
      this._sceneScale = null;
    }
    scene.background = this._bg;
    scene.fog = this._fog;
    this._scene = null;
    stage.inputEnabled = true;
    stage.cameraRestore(this._snap);
    this._snap = null;
    this.session = null;
    this.status(this._fit ? `left AR · fit was ${this._fit}` : 'left AR');
    this.onScene?.(false);
    this.onChange?.(false);
  }
}

function readStored() {
  try {
    const v = JSON.parse(localStorage.getItem(SPAWN_KEY) || 'null');
    return v && Number.isFinite(v.x) && Number.isFinite(v.z) ? v : null;
  } catch { return null; }
}

function fmt(v) {
  return Math.abs(v) >= 100 ? v.toFixed(0) : v.toFixed(v < 10 ? 2 : 1);
}
