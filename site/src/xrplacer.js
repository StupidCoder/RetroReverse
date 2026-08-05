// xrplacer.js — where the content goes, once per placement request.
//
// A placer answers exactly four numbers, and applyPlacement in xr.js turns them
// into a rig:
//
//   hold  a point of the CONTENT, in unscaled content units
//   at    a point of the ROOM, in reference space
//   k     metres per content unit
//   yaw   the content's bearing
//
// meaning "put hold at at, this big, facing this way". The two modes the Studio
// offers are two answers to that one question:
//
//   dioramaPlace   the level as a tabletop model: hold is the content's centre
//                  (or its underside), at is a spot on your floor, k makes it a
//                  metre across.
//   worldPlace     you inside the level: hold is the spawn tile at FLOOR level,
//                  at is where you stand, k is the game's real metres per unit.
//
// Splitting them this way rather than by branching inside the session is what
// keeps the status line, the far plane, the pose-wait counter and the reveal
// curtain single-implementation — each of those has a consumer that must not
// care which mode is running, and every one of them was a bug the last time
// something in this file grew a second code path.
//
// A placer returns null to mean "not this frame, and I have already said why".

import { THREE } from './engine3d.js';

// ---- the diorama ---------------------------------------------------------------

export function dioramaPlace(s, { frame, ref, pose }) {
  // Measure unscaled: contentBox reads world matrices, and on a recenter the
  // scene already carries the previous fit's scale.
  const scene = s.stage.scene;
  scene.scale.set(1, 1, 1);
  scene.updateMatrixWorld(true);
  const box = s.contentBox?.();
  if (!box || box.isEmpty()) {
    scene.scale.copy(s._sceneScale);
    scene.updateMatrixWorld(true);
    s.status('AR: nothing to place (empty content box)');
    return null;
  }

  const p = pose.transform.position;
  const fwd = headForward(pose);

  const size = box.getSize(new THREE.Vector3());
  const centre = box.getCenter(new THREE.Vector3());
  // Which point of the content is being placed, and where it goes. For a floor
  // anchor both drop to ground level: the box's bottom-centre, onto y = 0 ahead
  // of the viewer. fwd is already flattened, so the target stays on the floor
  // plane.
  const hold = s.anchor === 'floor'
    ? new THREE.Vector3(centre.x, box.min.y, centre.z)
    : centre;
  const spawn = s._spawnPoint(frame, ref);
  // A spot the viewer dragged content to is taken WHOLE, height included: if you
  // lifted the last map to eye level, the next one belongs there too. Everything
  // else only contributes x/z, and the height comes from the anchor — so an
  // eye-level diorama hangs ABOVE the place a floor model would stand on.
  const at = s._userAt ? s._userAt.clone()
    : spawn
      ? new THREE.Vector3(spawn.x, s.anchor === 'floor' ? 0 : p.y - s.dropBelowEye, spawn.z)
      : s.anchor === 'floor'
        ? new THREE.Vector3(p.x, 0, p.z).addScaledVector(fwd, s.distance)
        : new THREE.Vector3(p.x, p.y - s.dropBelowEye, p.z).addScaledVector(fwd, s.distance);
  const span = s.fitAxis === 'longest'
    ? Math.max(size.x, size.y, size.z)
    : Math.max(size.x, size.z) || Math.max(size.y, 1);
  // metres per world unit
  const k = s.fitScale ? s.fitScale(size) : s.targetSize / (span || 1);

  const front = typeof s.frontDir === 'function' ? s.frontDir() : s.frontDir;
  return {
    hold, at, k, size,
    yaw: bearingDelta(front, fwd),
    note: s._userAt ? 'spawn: where you put it' : s._spawnNote,
  };
}

// ---- world mode ------------------------------------------------------------------

// worldPlace stands you in the level. Everything it needs comes from the preset,
// so it does not measure the content at all — which is the point: the whole
// reason a preset exists is that no amount of measuring a GLB tells you how big
// a door is meant to be.
//
// `cfg` is a loaded preset (xrpreset.js). `spawn` may be overridden per call, so
// a teleport and the initial entry go through this same function.
export function makeWorldPlacer(cfg, hooks = {}) {
  return function worldPlace(s, { frame, ref, pose }) {
    const spawn = hooks.spawn?.() || cfg.spawn;
    const hold = new THREE.Vector3().fromArray(spawn.pos);

    // Where you stand. The play area's centroid means "the middle of the room
    // you can actually walk in" — better than the reference origin, which is
    // wherever the headset happened to be told the floor was. Height is always
    // 0: local-floor defines y = 0 as the measured floor, so the level's floor
    // meeting your floor costs nothing.
    //
    // NOT _spawnPoint: that prefers a saved anchor from a previous DIORAMA, and
    // arriving in the Abyss at the spot where you last parked a tabletop model
    // is not a feature.
    const centre = s.playAreaCentre(frame, ref);
    const at = centre ? centre.setY(0) : new THREE.Vector3(0, 0, 0);

    // The direction the preset wants you facing, mapped onto the direction you
    // are ACTUALLY facing right now, so entering never means turning round
    // first. Same formula the diorama uses, same reason.
    const dir = new THREE.Vector3().fromArray(spawn.dir || [0, 0, -1]);
    dir.y = 0;
    if (dir.lengthSq() < 1e-9) dir.set(0, 0, -1);
    const yaw = spawn.absolute
      ? Math.atan2(dir.x, dir.z)
      : bearingDelta(dir.normalize(), headForward(pose));

    return {
      hold, at, k: cfg.metresPerUnit, yaw,
      // The size the status line reports is the room you can SEE, not the
      // content's bounding box: a ten-kilometre course's diagonal says nothing
      // about whether the doors are the right height.
      size: new THREE.Vector3(1, 1, 1).multiplyScalar(cfg.fog.far / cfg.metresPerUnit),
      note: centre ? 'spawn: play-area centre' : 'spawn: reference origin',
    };
  };
}

// ---- shared ----------------------------------------------------------------------

// The head's forward, flattened. Yaw only: content sits level however the
// viewer's head is tilted.
export function headForward(pose) {
  const o = pose.transform.orientation;
  const fwd = new THREE.Vector3(0, 0, -1)
    .applyQuaternion(new THREE.Quaternion(o.x, o.y, o.z, o.w));
  fwd.y = 0;
  if (fwd.lengthSq() < 1e-6) fwd.set(0, 0, -1); // straight up/down: fall back to −Z
  return fwd.normalize();
}

// The yaw that carries `fwd` onto `front`. We need R·fwd = front, and for
// horizontal unit vectors that is the difference of their bearings. (Rotating BY
// the head's own bearing instead of by the difference turns the content the
// wrong way — it is the inverse rotation that carries fwd onto a fixed axis.)
export function bearingDelta(front, fwd) {
  return Math.atan2(front.x, front.z) - Math.atan2(fwd.x, fwd.z);
}
