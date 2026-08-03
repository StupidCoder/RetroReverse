// xr.js — immersive-AR ("diorama") mode for the 3D stages. The level is put
// into the room as a tabletop model floating in front of the viewer: a metre
// across by default, whatever the game's own units happen to be.
//
// The session does NOT scale the scene. It parents the stage camera into a
// rig group that carries the position, yaw and uniform scale, so every world
// coordinate in the document — fog distances, fly speed, route points,
// cutscene camera keyframes — keeps meaning exactly what it meant before, and
// leaving the session is one removal. three composes the rig for us
// (cameraXR.matrixWorld = rig.matrixWorld · view.matrix, applied to the left
// and right eye cameras too), so a rig scaled by 16 also scales the stereo
// eye separation, which is what a world where 1 m = 16 units needs.
//
// There is no console on the headset, so every failure has to be legible on
// the page itself: `status` reports what happened and survives session exit.

import { THREE } from './engine3d.js';

// One probe per page load, awaited by whoever needs it. `navigator.xr` is
// undefined outside a secure context, which is the usual reason the button
// never appears.
export const arSupported = (async () => {
  try {
    if (!navigator.xr?.isSessionSupported) return false;
    return await navigator.xr.isSessionSupported('immersive-ar');
  } catch {
    return false;
  }
})();

const UP = new THREE.Vector3(0, 1, 0);
const POSE_GIVE_UP = 120; // frames (~2 s) to wait for tracking before saying so

export class ARSession {
  // { stage, contentBox: () => THREE.Box3|null, targetSize (m), distance (m),
  //   dropBelowEye (m), onChange(active), onStatus(text) }
  constructor(opts) {
    this.stage = opts.stage;
    this.contentBox = opts.contentBox;
    this.targetSize = opts.targetSize || 1.0;
    this.distance = opts.distance ?? 1.5;
    this.dropBelowEye = opts.dropBelowEye ?? 0.15;
    // The world-space horizontal direction the viewer should be looking ALONG
    // — i.e. the document's own establishing shot, so the diorama presents the
    // side the level was framed from. Defaults to looking toward +Z (a viewer
    // on the −Z side), which is where most of these documents put the camera.
    this.frontDir = opts.frontDir || new THREE.Vector3(0, 0, 1);
    this.onScene = opts.onScene || null;   // the view's own scene tweaks (sky, cutscenes)
    this.onChange = opts.onChange || null; // the shell's button/filter state, assigned later
    this.onStatus = opts.onStatus || null;
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
        optionalFeatures: ['dom-overlay'],
        domOverlay: { root: document.body },
      });
    } catch (e) {
      this.onScene?.(false);
      this.status(`AR unavailable: ${e.message || e.name}`);
      throw e;
    }

    this.session = session;
    this._snap = stage.cameraSnapshot();
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

    this.rig = new THREE.Group();
    this.rig.name = 'xr-rig';
    stage.scene.add(this.rig);
    this.rig.add(stage.camera);
    // In a session the near/far planes are read in METRES and pushed to the
    // compositor as depthNear/depthFar, whatever the world units are.
    stage.camera.near = 0.02;
    stage.camera.far = 100;
    stage.camera.updateProjectionMatrix();

    this._place = true;
    this._waited = 0;
    stage.onXRFrame = (frame) => this._frame(frame);
    // The controller trigger re-places the diorama in front of wherever you
    // are now looking (and re-measures the content, so a level whose rooms
    // streamed in after entry gets refitted).
    session.addEventListener('select', () => { this._place = true; this._waited = 0; });
    // Restore off three's OWN sessionend, not the session's 'end'. A listener
    // on the session fires in registration order, and ours would be added
    // before three's (setSession registers it) — so it would run while
    // isPresenting is still true and before the drawing buffer is restored,
    // and the aspect fix inside cameraRestore would be skipped, leaving the
    // desktop view stretched. three dispatches 'sessionend' last, after
    // clearing the flag and restoring the size.
    this._onEnd = () => this._end();
    stage.renderer.xr.addEventListener('sessionend', this._onEnd);

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

  // _frame runs once per XR frame, before the scene updaters. It does nothing
  // at all once the diorama is placed.
  _frame(frame) {
    if (!this._place || !frame) return;
    const ref = this.stage.renderer.xr.getReferenceSpace();
    const pose = ref && frame.getViewerPose(ref);
    if (!pose) {
      // Tracking can take a few frames to settle; only complain if it never does.
      if (++this._waited === POSE_GIVE_UP) this.status('AR: no head pose — tracking not available');
      return;
    }
    const box = this.contentBox?.();
    if (!box || box.isEmpty()) {
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

    // Where the diorama's centre should land, in reference space (metres).
    const at = new THREE.Vector3(p.x, p.y - this.dropBelowEye, p.z).addScaledVector(fwd, this.distance);

    const size = box.getSize(new THREE.Vector3());
    const centre = box.getCenter(new THREE.Vector3());
    const span = Math.max(size.x, size.z) || Math.max(size.y, 1);
    const scale = span / this.targetSize; // world units per metre

    // Yaw so that the viewer's forward maps onto the document's establishing
    // direction: we need R·fwd = frontDir, and for horizontal unit vectors
    // that is the difference of their bearings. (Rotating BY the head's own
    // bearing instead of by the difference turns the diorama the wrong way —
    // it is the inverse rotation that carries fwd onto a fixed axis.)
    const yaw = Math.atan2(this.frontDir.x, this.frontDir.z) - Math.atan2(fwd.x, fwd.z);

    const rig = this.rig;
    rig.position.set(0, 0, 0);
    rig.quaternion.setFromAxisAngle(UP, yaw);
    rig.scale.setScalar(scale);
    rig.updateMatrixWorld(true);
    // Solve the translation directly: with the rig at the origin its world
    // matrix is R·S, so t = centre − R·S·at puts the reference-space point
    // `at` exactly on the content's centre. Correct for any starting
    // orientation and wherever the guardian happened to put the floor origin.
    rig.position.copy(centre).sub(at.clone().applyMatrix4(rig.matrixWorld));
    rig.updateMatrixWorld(true);

    this._place = false;
    this._fit = `${fmt(size.x)}×${fmt(size.y)}×${fmt(size.z)} u · ${fmt(scale)} u/m · ${this.targetSize} m at ${this.distance} m`;
    this.status(`in AR · ${this._fit}`);
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
      this.rig.remove(stage.camera);
      stage.scene.remove(this.rig);
      this.rig = null;
    }
    stage.scene.background = this._bg;
    stage.scene.fog = this._fog;
    stage.inputEnabled = true;
    stage.cameraRestore(this._snap);
    this._snap = null;
    this.session = null;
    this.status(this._fit ? `left AR · fit was ${this._fit}` : 'left AR');
    this.onScene?.(false);
    this.onChange?.(false);
  }
}

function fmt(v) {
  return Math.abs(v) >= 100 ? v.toFixed(0) : v.toFixed(v < 10 ? 2 : 1);
}
