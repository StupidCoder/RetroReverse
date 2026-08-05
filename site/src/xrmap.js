// xrmap.js — a 2-D level as three.js content: the whole tilemap in one draw
// call (xrtiles.js) and its whole cast in another (xrsprites.js), both built
// from the plain arrays tileatlas.js produces.
//
// It used to own an AR session too — a hidden stage, the placement, the exit
// paths. The XR shell owns the session now, and this owns only the content,
// which is exactly what lets a wall map and a 3-D diorama swap places without
// the session noticing.
//
// It also used to snapshot the level out of PixiJS and hang the pixels as
// textured chunks. That worked and was bandwidth-bound on a headset — 17.7 MB
// of baked texture for Sonic act01 against 77 KB of tile art for the same
// level, which on a part with limited memory bandwidth is the whole difference
// between 50 and 90 fps. Sampling the source art instead deleted the chunk
// pump, the half-resolution budget, the per-repaint canvas readback, the phase
// precomputation that worked around it, and the still/live duality with it.

export class TileContent {
  // { model: () => ({ tm, atlasImg, tickHz, clip, placements, advance(dt),
  //   begin(), end() }), params }
  constructor(opts) {
    this.opts = opts;
    this.keying = this._param('xrkey') !== 'none';
    this._parts = null;
    this._stage = null;
    this._running = false;
    this.note = '';

    const maxW = this._num('xrw', this._num('xrsize', 4.0));
    const maxH = this._num('xrh', 2.5);
    this._maxW = maxW;
    // A wall-sized map fits a BOX, not an axis; and it hangs at eye level
    // rather than dropping below it, because it is a picture.
    this.fit = {
      fitScale: (size) => Math.min(maxW / (size.x || 1), maxH / (size.y || 1)),
      dropBelowEye: 0,
      distance: this._num('xrdist', Math.max(1.2, 0.62 * maxW)),
    };
    // The quads' faces are +Z, so the viewer belongs on the +Z side looking
    // back along −Z. A plain object: this module must not pull in three.js,
    // and the session only ever reads .x and .z off it.
    this.frontDir = { x: 0, y: 0, z: -1 };
  }

  _param(k) {
    const p = this.opts.params;
    return p?.get?.(k) ?? p?.[k];
  }

  _num(k, d) {
    const n = parseFloat(this._param(k));
    return Number.isFinite(n) && n > 0 ? n : d;
  }

  // Rebuilding is milliseconds rather than a re-snapshot, so the backdrop
  // toggle is instant — which matters, because it is wrong by default on the
  // top-down maps and right on the side-scrollers, and no rule in the data
  // separates them.
  async setKeying(on) {
    if (on === this.keying) return;
    this.keying = on;
    if (!this._stage) return;
    const stage = this._stage;
    this._drop();
    await this.build(stage);
  }

  async build(stage) {
    this._stage = stage;
    const { THREE } = await import('./engine3d.js');
    const { buildTileModel } = await import('./tileatlas.js');
    const { TileMapMesh } = await import('./xrtiles.js');
    const { SpriteLayer } = await import('./xrsprites.js');
    this.THREE = THREE;

    const t0 = performance.now();
    const m = this.opts.model();

    const img = m.atlasImg;
    const cv = document.createElement('canvas');
    cv.width = img.width; cv.height = img.height;
    const c2 = cv.getContext('2d', { willReadFrequently: true });
    c2.imageSmoothingEnabled = false;
    c2.drawImage(img, 0, 0);

    const model = buildTileModel({
      tm: m.tm,
      atlas: c2.getImageData(0, 0, img.width, img.height),
      key: this.keying ? (this._forcedKey() || 'auto') : 'none',
      maxLayers: stage.renderer.getContext().getParameter(0x88FF), // MAX_ARRAY_TEXTURE_LAYERS
    });

    const aniso = Math.max(1, Math.round(this._num('xraniso', 4)));
    const blend = this._param('xrblend') === '1';
    const tiles = new TileMapMesh(THREE, model, { renderer: stage.renderer, aniso, blend });
    const sprites = new SpriteLayer(THREE, {
      placements: m.placements,
      clip: m.clip,
      W: model.mapPx.w, H: model.mapPx.h,
      renderer: stage.renderer,
    });

    const group = new THREE.Group();
    group.name = 'xr-tilemap';
    group.add(tiles.group, sprites.group);
    stage.scene.add(group);
    this._parts = { group, tiles, sprites, model, tickHz: m.tickHz || 60 };

    const s = model.stats, sc = sprites.counts;
    const ms = Math.round(performance.now() - t0);
    this.note = `${s.bricks} bricks · ${s.slots} slots/${s.layers}L`
      + (model.P > 1 ? `/P${model.P}` : '')
      + ` · ${sc.sprites} sprites/${sc.sheets} sheets`
      + ` · ${Math.round((s.tilesKB + s.indexKB + s.farKB) / 1024 * 10) / 10} MB · ${ms} ms`
      + (model.key ? ` · key #${hex(model.key)} ${Math.round(model.keyStats.cellPct)}% cells` : ' · unkeyed');
    // A map that keys away to nothing is never what was wanted; say so rather
    // than present an empty wall.
    if (model.key && model.keyStats.cellPct > 95) this.note += ' — KEY REMOVES NEARLY EVERYTHING';

    if (!this._running) { m.begin?.(); this._running = true; }
    return this;
  }

  // The session drives the clock: window rAF is not reliably serviced while
  // presenting, so the pixi ticker is stopped for the session's duration and
  // this advances the placements instead.
  advance(dt) {
    if (!this._parts) return;
    const clamped = Math.min(dt, 0.05);
    this.opts.model().advance(clamped);            // the placements' own clock
    this._parts.tiles.advance(clamped * this._parts.tickHz); // engine frames
    this._parts.sprites?.sync();
  }

  contentBox() {
    const THREE = this.THREE;
    if (!THREE || !this._parts) return null;
    const { w, h } = this._parts.model.mapPx;
    return new THREE.Box3(new THREE.Vector3(-w / 2, -h / 2, 0), new THREE.Vector3(w / 2, h / 2, 0));
  }

  _drop() {
    if (!this._parts) return;
    const { group, tiles, sprites } = this._parts;
    group.removeFromParent();
    tiles.dispose();
    sprites?.dispose();
    this._parts = null;
  }

  // Every exit path comes through here: the flat view gets its clock back.
  dispose() {
    this._drop();
    if (this._running) {
      this._running = false;
      try { this.opts.model().end?.(); } catch { /* the view may already be gone */ }
    }
    this._stage = null;
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
