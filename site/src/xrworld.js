// xrworld.js — world mode: the level, at life scale, with you inside it.
//
// Everything specific to standing in a level rather than looking at one lives
// here, and it is deliberately a small amount: the placement is the session's
// own applyPlacement with different numbers (xrplacer.js), the darkness is fog
// (xrtorch.js), and getting about is an arc against an index (xrteleport.js /
// xrfloor.js). What is left is wiring, plus two decisions worth reading.
//
// WHAT COUNTS AS FLOOR. The index is built from the level's LAYER groups, not
// from everything in the scene. That excludes the sprite billboards — Ultima
// Underworld level 1 has 306 of them, all vertical quads, and a "side-facing
// triangle" grid built without the exclusion would treat every mushroom in the
// dungeon as a wall you cannot see past. It also gives `floor.source:
// "collision"` for free: 88 of the 370 shipped level documents carry a hidden
// role:"collision" layer, which is the game's own walkable surface and usually
// cleaner than the art.
//
// WHEN THE INDEX IS BUILT. A few frames AFTER the placement lands, not during
// it. Building costs single-digit to low-tens of milliseconds, which is nothing
// at load and a visible hitch at 90 Hz — and the frames right after entry are
// the ones where the viewer is looking around rather than teleporting.

import { THREE, ObjectLibrary, applyTransform, disposeScene } from './engine3d.js';
import { makeWorldPlacer } from './xrplacer.js';
import { buildFloor } from './xrfloor.js';
import { Torch } from './xrtorch.js';
import { Teleporter } from './xrteleport.js';

const BUILD_DELAY_FRAMES = 3;

// The horizontal radius of a subtree, in its own units.
function radiusOf(root) {
  const size = new THREE.Box3().setFromObject(root).getSize(new THREE.Vector3());
  return Math.max(size.x, size.z) / 2;
}

export class WorldMode {
  // { ar, stage, cfg, game, view, signal, onStatus }
  constructor(opts) {
    this.ar = opts.ar;
    this.stage = opts.stage;
    this.cfg = opts.cfg;
    this.game = opts.game;
    this.view = opts.view;
    this.onStatus = opts.onStatus || null;

    // Chrome rather than content: the shell's curtain and its content box both
    // skip anything flagged this way, so the teleport marker is neither hidden
    // while a level streams nor measured as part of it.
    this.group = new THREE.Group();
    this.group.name = 'xr-world';
    this.group.userData.xrChrome = true;
    this.stage.scene.add(this.group);

    this.ar.placer = makeWorldPlacer(this.cfg);
    // The far plane follows the FOG, not the content: a Need for Speed course is
    // 10 km of road and you can see eleven metres of it. This is both the
    // correct answer and the single biggest performance lever in the mode.
    this.ar.farFor = (k, size) => (this.cfg.fog
      ? this.cfg.fog.far * 1.5
      : (size ? Math.min(4000, Math.max(100, size.length() * k * 1.5)) : null));

    this.torch = new Torch(this.stage, this.cfg);
    this.torch.exempt(opts.ui);

    this.teleporter = new Teleporter({
      ar: this.ar,
      cfg: this.cfg,
      group: this.group,
      floor: () => this._floor,
      blockers: () => this._walls,
      onChange: () => this.onStatus?.(),
    });

    this._frames = 0;
    this._step = () => this._tick();
    this.stage.updaters.add(this._step);
  }

  // _fitSky makes a camera-attached layer behave like a horizon: infinitely far,
  // never in the depth buffer, never in the fog.
  //
  // THE CUE IS DISPARITY, NOT DISTANCE, and getting that wrong cost two rounds.
  // Need for Speed's horizon is a UNIT-RADIUS cylinder centred on the mid-eye,
  // so each eye sits 32 mm off its centre: 6.4% parallax, and it reads as a
  // painted wall you could touch. Scaling it to 0.45 of the far plane — 283 m —
  // sounds like plenty and is not. A 64 mm IPD at 283 m still subtends about 47
  // arcseconds of binocular disparity, and human stereoacuity runs to 20-30
  // arcseconds, so the horizon remains measurably nearer than infinity. Pushing
  // it further only fights the far plane; you would need a kilometre or two.
  //
  // So the dome is not pushed away, it is centred on the EYE BEING DRAWN rather
  // than on the head. Both eyes then receive the identical image, disparity is
  // exactly zero, and the sky is at infinity by construction at whatever radius
  // is convenient. three calls onBeforeRender once per sub-camera of an
  // ArrayCamera and honours a matrixWorld written inside it (verified: two
  // calls, L then R, and the per-eye transform reaches modelViewMatrix), which
  // is the whole mechanism.
  //
  // Scale then only has to satisfy two much weaker constraints: inside the far
  // plane, so it is drawn at all — SM64DS's vr01 is 59,000 units, a 645 km dome
  // that was being clipped away entirely — and beyond the terrain.
  _fitSky() {
    const sky = this.cfg.sky;
    const scene = this.stage.scene;
    for (const o of this._skyHooked || []) o.onBeforeRender = () => {};
    this._skyHooked = [];
    for (const root of scene.children) {
      const ly = root.userData?.layer;
      if (!ly) continue;
      const camAttached = ly.attach === 'camera' || ly.attach === 'cameraYaw';
      if (!camAttached && ly.role !== 'sky') continue;
      if (!sky.show) { root.visible = false; continue; }

      let k = sky.scale;
      if (k === 'auto') {
        // Unscaled first: this runs again on every re-place, and measuring a
        // root that already carries a previous fit would square it.
        root.scale.setScalar(1);
        root.updateMatrixWorld(true);
        // Measure the layer's own radius rather than assume one: Need for Speed
        // ships a unit cylinder, other games ship domes of tens of thousands of
        // units, and both want to end up in the same place.
        const radius = radiusOf(root);
        const farUnits = (this.stage.camera.far || 100) / this.cfg.metresPerUnit;
        k = radius > 1e-6 ? (farUnits * 0.75) / radius : 1;
      }
      root.scale.setScalar(k);
      const r = radiusOf(root) * this.cfg.metresPerUnit;
      this._skyNote = `sky ${r > 1000 ? `${(r / 1000).toFixed(1)} km` : `${r.toFixed(0)} m`} · per-eye`;

      // Both directions, and this is the correction that matters: an earlier
      // version only ever pushed a horizon OUT, on the reasoning that a dome
      // already far away had nothing to gain. It has everything to lose —
      // Super Mario 64 DS's vr01 is 59,000 units across, which at 10.3 m per
      // unit is a 607 km dome, i.e. entirely outside any far plane worth having
      // and therefore not drawn at all. A horizon belongs at one distance: just
      // inside the far plane, wherever it started.
      //
      // Pulling one IN is only safe with painter's-order treatment, so that is
      // asserted here rather than assumed of the document. Need for Speed's sky
      // layer declares renderOrder/depthTest itself; SM64DS's declares neither,
      // and a 160 m dome depth-tested against a 165 m course would be sliced
      // open by its own hills. A camera-attached horizon is by definition
      // infinitely far, so it must never take part in depth at all.
      root.renderOrder = -1000;
      const eye = new THREE.Vector3();
      // Re-centred per eye, immediately before that eye draws it. Assigned to
      // every mesh rather than to the group, because onBeforeRender is a
      // per-object hook and there is no guarantee the sky's meshes are drawn
      // consecutively; each one re-solves the group and is then correct.
      const perEye = (renderer, scene, cam) => {
        cam.getWorldPosition(eye);
        root.parent?.worldToLocal(eye);
        root.position.copy(eye);
        root.updateMatrixWorld(true);
      };
      root.traverse((o) => {
        o.renderOrder = -1000;
        if (o.isMesh) { o.onBeforeRender = perEye; this._skyHooked.push(o); }
        for (const m of o.material ? (Array.isArray(o.material) ? o.material : [o.material]) : []) {
          if (!m.depthTest && !m.depthWrite) continue;
          m.depthTest = false;
          m.depthWrite = false;
          m.needsUpdate = true;
        }
      });

      if (!sky.fog) {
        root.traverse((o) => {
          for (const m of o.material ? (Array.isArray(o.material) ? o.material : [o.material]) : []) {
            if (m.fog === false) continue;
            m.fog = false;
            m.needsUpdate = true;
          }
        });
      }
    }
  }

  // The interactor the shell talks to IS the teleporter; world mode has no
  // grabber, because dragging a room you are standing in would undo the one
  // thing life scale is for.
  get interactor() { return this.teleporter; }

  _tick() {
    if (this._floor || this._failed) return;
    // Not until the placement has landed: the index is built in content units,
    // and before a fit there is no telling whether the content is even loaded.
    if (!this.ar.placement) return;
    if (++this._frames < BUILD_DELAY_FRAMES) return;
    try {
      this._build();
    } catch (e) {
      this._failed = true;
      console.error('xr floor index', e);
      this.onStatus?.(`floor index failed: ${e.message || e}`);
    }
  }

  _build() {
    const meshes = this._contentMeshes();
    const t0 = performance.now();
    this._floor = buildFloor(meshes, {
      mode: 'up',
      maxSlope: this.cfg.floor.maxSlope,
      twoSided: this.cfg.floor.twoSided,
    });
    if (this.cfg.teleport.blockers) {
      this._walls = buildFloor(meshes, { mode: 'side', minSlope: 70 });
    }
    this._note = `floor ${this._floor.note} in ${Math.round(performance.now() - t0)} ms`;
    if (!this._floor.n) this._note = 'floor index: NOTHING walkable found — teleport will not work';
    this.onStatus?.();
  }

  // The layer groups that make up the ground, as plain arrays in the content's
  // own coordinate space. Content space rather than world space so the index
  // survives a change of scale — `hold` is in content units too, and the arc is
  // integrated there.
  _contentMeshes() {
    const scene = this.stage.scene;
    scene.updateMatrixWorld(true);
    const inv = new THREE.Matrix4().copy(scene.matrixWorld).invert();
    const want = this.cfg.floor.source;
    const out = [];
    for (const root of scene.children) {
      const ly = root.userData?.layer;
      if (!ly) continue;                       // placements, billboards, chrome
      if (ly.attach === 'camera' || ly.attach === 'cameraYaw') continue;
      if (want === 'collision' && ly.role !== 'collision') continue;
      if (want === 'visible' && (ly.role === 'sky' || ly.role === 'collision')) continue;
      if (want !== 'collision' && want !== 'visible' && ly.id !== want) continue;
      // A hidden collision layer is exactly what we want to read and exactly
      // what a visibility check would skip, so there is no visibility check.
      root.traverse((o) => {
        const g = o.isMesh && o.geometry?.attributes?.position;
        if (!g) return;
        out.push({
          positions: g.array,
          indices: o.geometry.index ? o.geometry.index.array : null,
          matrix: new THREE.Matrix4().multiplyMatrices(inv, o.matrixWorld).elements,
        });
      });
    }
    return out;
  }

  // ---- props ------------------------------------------------------------------

  // Extra objects the preset puts in the world — the Ferrari and the Diablo on
  // the start line. Instantiated through the SAME ObjectLibrary the level's own
  // placements use, so a prop is a placement in every respect except that a
  // person wrote it rather than the game.
  async loadProps(signal) {
    const props = this.cfg.props;
    if (!props.length) return;
    this.lib = new ObjectLibrary(this.game, signal);
    for (const p of props) {
      try {
        const inst = await this.lib.instance(p.object, { scene: p.variant || undefined });
        if (signal?.aborted) return;
        applyTransform(inst.node, {
          pos: p.pos,
          rot: [0, (p.rotY * Math.PI) / 180, 0],
          scale: p.scale,
        });
        this.stage.scene.add(inst.node);
        (this._props ||= []).push(inst);
        // playAnim, not actions[] — ObjectLibrary exposes the former and has no
        // such field as the latter, so `anim` in a preset had never once played.
        if (p.anim && inst.playAnim) inst.playAnim(p.anim);
        if (inst.update) {
          const u = (dt, cam, t) => inst.update(dt, cam, t);
          (this._propUpdaters ||= []).push(u);
          this.stage.updaters.add(u);
        }
      } catch (e) {
        // One bad prop id must not cost the whole session — say so and carry on.
        console.error(`xr prop ${p.object}`, e);
        this.onStatus?.(`prop ${p.object}: ${e.message || e}`);
      }
    }
    // Late arrivals need the torch patch too, if it is in use.
    for (const inst of this._props || []) this.torch.scan(inst.node);
  }

  // Re-run over whatever is in the scene now. Cheap, and the safety net for a
  // level that streams rooms in after the mount resolves.
  //
  // The sky is fitted from here rather than from the constructor because it is
  // sized against the FAR PLANE, and the far plane is set by farFor when a
  // placement lands — which has not happened yet when the mode is built.
  rescan() {
    this.torch.scan(this.stage.scene);
    this._fitSky();
  }

  get note() {
    return [this.torch.note, this._skyNote, this._note].filter(Boolean).join(' · ');
  }

  dispose() {
    this.stage.updaters.delete(this._step);
    for (const o of this._skyHooked || []) o.onBeforeRender = () => {};
    this._skyHooked = null;
    this.teleporter.dispose();
    this.torch.dispose();
    for (const u of this._propUpdaters || []) this.stage.updaters.delete(u);
    this._propUpdaters = null;
    for (const inst of this._props || []) { inst.node.removeFromParent(); disposeScene(inst.node); }
    this._props = null;
    this.lib?.dispose();
    this.group.removeFromParent();
    this._floor = this._walls = null;
    // The placer and the far-plane policy go back to the session's defaults; the
    // session outlives the mode when the browser switches to an asset that has
    // no preset.
    this.ar.farFor = null;
  }
}
