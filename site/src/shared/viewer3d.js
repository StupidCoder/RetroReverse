// Viewer3D — a generic, manifest-driven 3-D viewer the Studio drives like any other. It wraps a
// Stage3D, a map of renderer plugins and a manifest base URL. init() fetches the manifest and
// returns its models (the browse list); showItem() resets the stage, resolves the plugin for the
// item's kind and lets it build the scene. It exposes the same surface the Studio expects of a
// 3-D viewer: the canvas lives inside `el` (the screen filter captures it), `active` gates the RAF
// loop when hidden, and `three`/`camera`/`controls` let the keyboard camera dolly & pan.
import { Stage3D } from './stage3d.js';
import { resolveRenderer } from './renderers.js';
import { applyWireframe } from './wireframe.js';

export class Viewer3D {
  constructor(el, hud, { base, renderers } = {}) {
    this.el = el;
    this.hud = hud;
    this.base = base;
    this.renderers = renderers || {};
    this.stage = new Stage3D(el);
  }

  // ---- the surface the Studio + shared helpers read ----
  get canvas() { return this.stage.canvas; }
  // camera.js threeCam() looks for v.three.camera / v.camera+controls; expose both shapes.
  get three() {
    const s = this.stage;
    return { renderer: s.renderer, scene: s.scene, camera: s.camera, controls: s.controls };
  }
  get camera() { return this.stage.camera; }
  get controls() { return this.stage.controls; }
  // Studio layer toggles. The generic one is "wireframe" (every mesh in the built scene);
  // plugins add their own behaviour by exposing stage.setLayer if needed.
  setLayer(id, on) {
    if (id === 'wireframe') applyWireframe(this.stage.scene, on);
    else if (this.stage.setLayer) this.stage.setLayer(id, on);
  }
  // A fly-through plugin (e.g. Stunt Car's stunt-model) publishes its FlyCam as stage.fly; expose
  // it so the Studio's KeyboardCamera cedes the arrow keys to it (it checks viewer.fly.enabled).
  // Non-fly plugins (Elite) leave it null — showItem() resets stage.fly before each build.
  get fly() { return this.stage.fly ?? null; }
  get active() { return this.stage.active; }
  set active(v) { this.stage.active = v; }           // setMountActive() pauses hidden viewers here
  start() { this.stage.start(); }
  stop() { this.stage.stop(); }
  // The active plugin may publish a native pixel grid (Elite's low-res C64 raster) so the global
  // CRT filter can lock its scanlines/mask to it; null for full-res 3-D content.
  pixelGrid() { return this.stage.pixelGrid ? this.stage.pixelGrid() : null; }

  // Fetch the manifest and return the browse list: its models (Elite's ships), any bespoke views[]
  // items (Stunt Car's circuits), and any levels[] (Ultima Underworld's dungeon levels), each carrying
  // kind/file/data/…. Elite ships only models[] and Stunt Car only views[], so those are unchanged.
  async init() {
    this.manifest = await fetch(this.base + 'manifest.json').then((r) => r.json());
    const m = this.manifest;
    return (m.models || []).concat(m.views || [], m.levels || []);
  }

  // Tear down the previous plugin, clear the stage and reset every plugin hook, then resolve and
  // run the renderer for this item's kind.
  async showItem(item) {
    if (this.stage.disposePlugin) this.stage.disposePlugin();
    this.stage.clear();
    this.stage.render = null;
    this.stage.onFrame = null;
    this.stage.pixelGrid = null;
    this.stage.hud = null;
    this.stage.fly = null; // a fly plugin republishes this; non-fly plugins (Elite) leave it null
    this.stage.disposePlugin = null;

    const load = resolveRenderer(item.kind, this.renderers);
    const mod = await load();
    const plugin = mod.default || mod;
    this._current = await plugin.build({ item, base: this.base, stage: this.stage });

    if (this.hud) this.hud.textContent = this.stage.hud || item.name || '';
    return this._current;
  }
}
