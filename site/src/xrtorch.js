// xrtorch.js — a torch, on geometry that cannot be lit.
//
// Not one shipped level GLB carries a NORMAL attribute, and every one of them
// declares KHR_materials_unlit, so they load as MeshBasicMaterial. A PointLight
// at the head would do literally nothing. That single fact decides the design.
//
// FOG IS THE TORCH. THREE.Fog blends every fragment toward a colour by distance,
// MeshBasicMaterial honours it by default, GLTFLoader's unlit materials keep it,
// and — the part that settles the argument — the billboard batch in
// billboards.js gets it for free. One assignment dims walls, sprites, props and
// sky together, with no shader edited and nothing to keep in sync. Flicker is
// then a number written into scene.fog every frame: no recompiles, no uniforms
// to plumb, and it doubles as the far-plane policy (xr.js farFor), which is the
// largest single performance lever in world mode — the ten kilometres of Need
// for Speed road you cannot see stop being drawn.
//
// The artifact this buys, honestly: THREE.Fog is per-fragment on view DEPTH
// (-mvPosition.z), not on radius, so turning your head changes a fixed wall's
// brightness slightly. It is also precisely the artifact every 1992 software
// renderer had. `torch.radial: true` swaps in a real radial falloff, and the
// reason it is opt-in rather than default is below.
//
// The radial patch is the only part of this that edits a shader, and it has to
// be careful, because billboards.js:220 ALREADY owns onBeforeCompile and
// customProgramCacheKey on its batch material — the sprite quads exist only
// because of its begin_vertex replacement. So the patch CHAINS (calls whatever
// hook was there first) and APPENDS to the cache key, and marks the material so
// patching twice cannot emit a duplicate uniform and fail to compile.

import { THREE } from './engine3d.js';

export class Torch {
  // { stage, cfg } — cfg is a normalised preset (xrpreset.js)
  constructor(stage, cfg) {
    this.stage = stage;
    this.cfg = cfg;
    this.on = true;
    this._patched = new WeakSet();
    this._materials = 0;
    this._t = 0;

    const f = cfg.fog;
    this._base = f ? { near: f.near, far: f.far } : null;
    if (f) {
      // Set BEFORE the curtain lifts. Assigning scene.fog mid-session forces a
      // program change on every material three tracks — 55 of them on an Alpine
      // course — and paying that while presenting is a visible stall.
      stage.scene.fog = new THREE.Fog(new THREE.Color(f.color), f.near, f.far);
    }
    if (cfg.background) stage.scene.background = new THREE.Color(cfg.background);

    // The menu is in this scene too, and it must not fog out: a browser you
    // cannot read at arm's length is a browser you cannot leave.
    this._unfog = [];

    this._step = (dt) => this._tick(dt);
    stage.updaters.add(this._step);
  }

  // exempt keeps a subtree out of the fog — the menu, the pointer, the teleport
  // marker. Called with the UI group; the materials are Basic and their `fog`
  // flag is a program define, so this costs one recompile each, once.
  exempt(root) {
    root.traverse((o) => {
      for (const m of mats(o)) {
        if (m.fog === false) continue;
        m.fog = false;
        m.needsUpdate = true;
        this._unfog.push(m);
      }
    });
  }

  // scan is idempotent and cheap enough to re-run: a level that streams rooms
  // adds materials for seconds after the mount resolves, and a one-shot walk at
  // placement time would light the first room and leave the rest bright.
  scan(root) {
    if (!this.cfg.torch.radial) return;
    root.traverse((o) => {
      for (const m of mats(o)) this._patch(m);
    });
  }

  setEnabled(on) {
    this.on = on;
    const f = this.stage.scene.fog;
    if (!f || !this._base) return;
    // "Off" is the fog pushed out of the way rather than removed: removing it
    // recompiles everything, and the whole point of the flicker path is that it
    // never does.
    f.near = on ? this._base.near : 1e6;
    f.far = on ? this._base.far : 1e7;
  }

  _tick(dt) {
    const f = this.stage.scene.fog;
    const t = this.cfg.torch;
    if (!f || !this._base || !this.on || !t.flicker) return;
    this._t += dt;
    // Two incommensurable rates, so the flame never finds a loop: a fast
    // guttering and a slow breath. Amplitude is a FRACTION of the reach, so a
    // torch that lights ten metres wavers by a metre at flicker 0.1 and never
    // goes dark.
    const a = Math.sin(this._t * t.flickerHz * 2 * Math.PI);
    const b = Math.sin(this._t * t.flickerHz * 0.61 * 2 * Math.PI + 1.7);
    const k = 1 + t.flicker * (a * 0.6 + b * 0.4);
    f.far = this._base.far * k;
    f.near = this._base.near * k;
    if (this.uniforms) this.uniforms.rxTorchRange.value = f.far;
  }

  // ---- the optional radial falloff ---------------------------------------------

  _patch(m) {
    if (!m || this._patched.has(m)) return;
    this._patched.add(m);
    this._materials++;
    if (!this.uniforms) {
      this.uniforms = {
        rxTorchRange: { value: this.cfg.fog?.far ?? 20 },
        rxTorchInner: { value: (this.cfg.fog?.near ?? 1) },
      };
    }
    const uniforms = this.uniforms;
    const prev = m.onBeforeCompile;
    const prevKey = m.customProgramCacheKey;
    m.onBeforeCompile = function (shader, renderer) {
      // Whatever was here first goes first: the billboard batch builds its quads
      // in onBeforeCompile, and replacing that hook deletes every sprite in the
      // level.
      prev?.call(this, shader, renderer);
      Object.assign(shader.uniforms, uniforms);
      shader.vertexShader = shader.vertexShader
        .replace('#include <common>', '#include <common>\nvarying float vRxDist;')
        // AFTER project_vertex, where mvPosition exists. View space is metres —
        // the camera is never scaled, only the scene is — and the view origin IS
        // the eye, which is where the torch is. So the radius wanted here is
        // just the length of mvPosition, and no world-space uniform is needed.
        .replace('#include <project_vertex>', '#include <project_vertex>\nvRxDist = length(mvPosition.xyz);');
      shader.fragmentShader = shader.fragmentShader
        .replace('#include <common>', `#include <common>
varying float vRxDist;
uniform float rxTorchRange;
uniform float rxTorchInner;`)
        // Last thing before tone mapping, so it dims the finished colour rather
        // than one term of it.
        .replace('#include <tonemapping_fragment>', `gl_FragColor.rgb *= smoothstep(rxTorchRange, rxTorchInner, vRxDist);
#include <tonemapping_fragment>`);
    };
    m.customProgramCacheKey = function () {
      return `${prevKey ? prevKey.call(this) : ''}|rx-torch`;
    };
    // Without this the patch is ignored: three's needsProgramChange check does
    // not consult the cache key, and a material that has already compiled once
    // keeps its program.
    m.needsUpdate = true;
  }

  get note() {
    const f = this.stage.scene.fog;
    if (!f) return 'torch: off (no fog in preset)';
    return `torch: ${this.on ? `${f.far.toFixed(1)} m` : 'off'}`
      + (this.cfg.torch.radial ? ` · radial (${this._materials} mat)` : '');
  }

  dispose() {
    this.stage.updaters.delete(this._step);
    for (const m of this._unfog) { m.fog = true; m.needsUpdate = true; }
    this._unfog = [];
    // The fog itself is NOT restored here: the session snapshotted the scene's
    // original fog on entry and puts it back on exit, onto the scene it took it
    // from (which a content swap may since have replaced). Undoing it here as
    // well would be the second owner of one field.
  }
}

function mats(o) {
  if (!o.material) return [];
  return Array.isArray(o.material) ? o.material : [o.material];
}
