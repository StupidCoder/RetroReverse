// xrmap.js — the AR session shell for a 2-D level.
//
// The map itself is drawn by xrtiles.js (one draw call for every cell) and its
// cast by xrsprites.js (one for every placement), both built from the plain
// arrays tileatlas.js produces. This file owns only the session: the hidden
// stage, the placement, the toggles and the readout.
//
// It used to snapshot the level out of PixiJS and hang the pixels as textured
// chunks. That worked and was bandwidth-bound on a headset — 17.7 MB of baked
// texture for Sonic act01 against 77 KB of tile art for the same level, which
// on a part with limited memory bandwidth is the whole difference between 50
// and 90 fps. Sampling the source art instead deleted the chunk pump, the
// half-resolution budget, the per-repaint canvas readback, the phase
// precomputation that worked around it, and the still/live duality with it.

import { ARSession, arSupported } from './xr.js';

export class MapAR {
  // { el, params, onStatus, model: () => ({ tm, atlasImg, tickHz, clip,
  //   placements, advance(dt), begin(), end() }) }
  constructor(opts) {
    this.opts = opts;
    this.supported = arSupported;
    this.onChange = null;
    this.keying = this._param('xrkey') !== 'none';
    this._ar = null;
    this._built = false;
    this._note = '';
    this._running = false;
  }

  get active() { return !!this._ar?.active; }

  _param(k) {
    const p = this.opts.params;
    return p?.get?.(k) ?? p?.[k];
  }

  _num(k, d) {
    const n = parseFloat(this._param(k));
    return Number.isFinite(n) && n > 0 ? n : d;
  }

  status(t) {
    this._base = t;
    const bits = [t];
    if (this._note) bits.push(this._note);
    if (this._perfNote) bits.push(this._perfNote);
    this.opts.onStatus?.(bits.join(' · '));
  }

  // Rebuilding is now milliseconds rather than a re-snapshot, so the backdrop
  // toggle is instant — which matters, because it is wrong by default on the
  // top-down maps and right on the side-scrollers, and no rule in the data
  // separates them.
  setKeying(on) {
    if (on === this.keying) return;
    this.keying = on;
    this._dropContent();
  }

  async enter() {
    if (!this._ar) await this._build();
    if (!this._built) this._buildContent();
    await this._ar.enter();
  }

  exit() { this._ar?.exit(); }

  dispose() {
    this._ar?.exit();
    this._stop();
    this._dropContent();
    this._teardown?.();
    this._teardown = null;
    this._ar = null;
  }

  // _stop hands the flat view its clock back. The session drives the placement
  // animations itself (window rAF is not reliably serviced while presenting),
  // so the pixi ticker is stopped for its duration and every exit path must
  // come through here.
  _stop() {
    if (!this._running) return;
    this._running = false;
    try { this.opts.model().end?.(); } catch { /* the view may already be gone */ }
  }

  _dropContent() {
    this._built = false;
    for (const part of [this._tiles, this._sprites]) {
      if (!part) continue;
      this._ar?.stage.scene.remove(part.group);
      part.dispose();
    }
    this._tiles = null;
    this._sprites = null;
  }

  async _build() {
    const { el } = this.opts;
    const { THREE, Stage } = await import('./engine3d.js');
    this.THREE = THREE;
    ({ buildTileModel: this.buildTileModel } = await import('./tileatlas.js'));
    ({ TileMapMesh: this.TileMapMesh } = await import('./xrtiles.js'));
    ({ SpriteLayer: this.SpriteLayer } = await import('./xrsprites.js'));

    // An immersive session renders into its own framebuffer, so the canvas
    // backing this context never has to be on screen — and must not be, or it
    // would cover the 2-D view.
    const mount = document.createElement('div');
    mount.style.cssText = 'position:absolute;left:0;top:0;width:0;height:0;overflow:hidden;pointer-events:none';
    el.appendChild(mount);
    const stage = new Stage(mount, { fov: 50, near: 0.01, far: 200 });
    this.stage = stage;

    const maxW = this._num('xrw', this._num('xrsize', 4.0));
    const maxH = this._num('xrh', 2.5);

    this._ar = new ARSession({
      stage,
      contentBox: () => {
        const { w, h } = this._mapPx();
        return new THREE.Box3(new THREE.Vector3(-w / 2, -h / 2, 0), new THREE.Vector3(w / 2, h / 2, 0));
      },
      fitScale: (size) => Math.min(maxW / (size.x || 1), maxH / (size.y || 1)),
      ready: () => this._built,
      dropBelowEye: 0,                        // a picture hangs at eye level
      frontDir: new THREE.Vector3(0, 0, -1),  // the quads' faces are +Z
      onStatus: (t) => this.opts.onStatus?.(t),
    });
    this._ar.onChange = (on) => {
      if (!on) this._stop();
      this.onChange?.(on);
    };
    this._maxW = maxW;

    // Everything animates, always, and it costs a slot-table write plus at
    // most 429 quads — so there is no longer a still mode to choose.
    this._perf = { n: 0, t: 0, cpu: 0, worst: 0 };
    this._step = (dt) => {
      if (!stage.renderer.xr.isPresenting) return; // the hidden stage loops forever
      if (this._built) {
        const t0 = performance.now();
        const m = this.opts.model();
        const clamped = Math.min(dt, 0.05);
        m.advance(clamped);                       // the placements' own clock
        this._tiles.advance(clamped * (m.tickHz || 60)); // engine frames
        this._sprites?.sync();
        this._perf.cpu += performance.now() - t0;
      }
      const p = this._perf;
      p.n++; p.t += dt; p.worst = Math.max(p.worst, dt);
      if (p.n >= 45) {
        const info = stage.renderer.info.render;
        this._perfNote = `${Math.round(p.n / p.t)} fps (worst ${Math.round(p.worst * 1000)} ms) · `
          + `${info.calls} calls · ${info.triangles} tris · cpu ${(p.cpu / p.n).toFixed(2)} ms`;
        this.status(this._base || 'in AR');
        this._perf = { n: 0, t: 0, cpu: 0, worst: 0 };
      }
    };
    stage.updaters.add(this._step);

    this._teardown = () => { stage.dispose(); mount.remove(); };
  }

  _mapPx() {
    if (this._tiles) return this._tiles.model.mapPx;
    const { tm } = this.opts.model();
    const cell = tm.tileSize * (tm.blocks ? tm.blocks.size : 1);
    return { w: tm.width * cell, h: tm.height * cell };
  }

  // Synchronous and measured in milliseconds, so it happens inside the click
  // that opens the session — no pump, no progress gate.
  _buildContent() {
    const THREE = this.THREE;
    const t0 = performance.now();
    const m = this.opts.model();

    const img = m.atlasImg;
    const cv = document.createElement('canvas');
    cv.width = img.width; cv.height = img.height;
    const c2 = cv.getContext('2d', { willReadFrequently: true });
    c2.imageSmoothingEnabled = false;
    c2.drawImage(img, 0, 0);

    const model = this.buildTileModel({
      tm: m.tm,
      atlas: c2.getImageData(0, 0, img.width, img.height),
      key: this.keying ? (this._forcedKey() || 'auto') : 'none',
      maxLayers: this.stage.renderer.getContext().getParameter(0x88FF), // MAX_ARRAY_TEXTURE_LAYERS
    });

    const aniso = Math.max(1, Math.round(this._num('xraniso', 4)));
    const blend = this._param('xrblend') === '1';
    this._tiles = new this.TileMapMesh(THREE, model, { renderer: this.stage.renderer, aniso, blend });
    this.stage.scene.add(this._tiles.group);

    this._sprites = new this.SpriteLayer(THREE, {
      placements: m.placements,
      clip: m.clip,
      W: model.mapPx.w, H: model.mapPx.h,
      renderer: this.stage.renderer,
    });
    this.stage.scene.add(this._sprites.group);

    const s = model.stats, sc = this._sprites.counts;
    const ms = Math.round(performance.now() - t0);
    this._note = `${s.bricks} bricks · ${s.slots} slots/${s.layers}L`
      + (model.P > 1 ? `/P${model.P}` : '')
      + ` · ${sc.sprites} sprites/${sc.sheets} sheets`
      + ` · ${Math.round((s.tilesKB + s.indexKB + s.farKB) / 1024 * 10) / 10} MB · ${ms} ms`
      + (model.key ? ` · key #${hex(model.key)} ${Math.round(model.keyStats.cellPct)}% cells` : ' · unkeyed');
    // A map that keys away to nothing is never what was wanted; say so rather
    // than present an empty wall.
    if (model.key && model.keyStats.cellPct > 95) this._note += ' — KEY REMOVES NEARLY EVERYTHING';

    this._ar.distance = this._num('xrdist', Math.max(1.2, 0.62 * this._maxW));
    m.begin?.();
    this._running = true;
    this._built = true;
  }

  _forcedKey() {
    const v = this._param('xrkey');
    const mm = /^#?([0-9a-f]{6})$/i.exec(v || '');
    if (!mm) return null;
    const n = parseInt(mm[1], 16);
    return { r: (n >> 16) & 255, g: (n >> 8) & 255, b: n & 255 };
  }
}

function hex({ r, g, b }) {
  return ((1 << 24) | (r << 16) | (g << 8) | b).toString(16).slice(1);
}
